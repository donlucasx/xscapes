package host

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"testing"
)

// TestReplayTraceBlankRows walks a window of a real trace through the model
// and, after every one of the host's own paints, counts the scape rows that
// were drawn at the previous paint and are empty now -- a row that was wiped
// by something other than the scape. Paints inside the first 200 KB after a
// size change are left out: a resize blanks and repaints on its own. The
// blank top rows of his 18:57 screenshot (2026-09-17) were never traced;
// this says whether they happened in a traced run, and where.
//
//	XSCAPES_TRACE=~/.config/xscapes/traces/<t>.bin TRACE_TO=260000000 \
//	  go test ./internal/host -run TestReplayTraceBlankRows -v
func TestReplayTraceBlankRows(t *testing.T) {
	path := traceEnv
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace FILE to replay it")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	if len(marks) == 0 {
		t.Fatal("no marks: the trace's .log is missing")
	}
	from, to := 0, int(st.Size())
	if v, err := strconv.Atoi(os.Getenv("TRACE_FROM")); err == nil {
		from = v
	}
	if v, err := strconv.Atoi(os.Getenv("TRACE_TO")); err == nil && v < to {
		to = v
	}
	cur := marks[0]
	next := 1
	for next < len(marks) && marks[next].off <= from {
		cur = marks[next]
		next++
	}
	sc := newScreen(cur.cols, cur.rows)
	agent := cur.agent
	endPaint := regexp.MustCompile(`\x1b\[1;\d+r\x1b\[\?6h\x1b8`)
	afterResize := func(off int) bool {
		for _, m := range marks {
			if off >= m.off && off-m.off < 200_000 {
				return true
			}
		}
		return false
	}
	var prevDrawn []bool
	paints, wipedPaints, wipedRows, maxWiped, maxAt, firstAt := 0, 0, 0, 0, 0, -1
	var worst []string
	check := func(off int) {
		paints++
		drawn := make([]bool, sc.h)
		wiped := 0
		for y := agent; y < sc.h; y++ {
			drawn[y] = sc.rowAt(y) != ""
			if y < len(prevDrawn) && prevDrawn[y] && !drawn[y] && !afterResize(off) {
				wiped++
			}
		}
		if wiped > 0 {
			wipedPaints++
			wipedRows += wiped
			if firstAt < 0 {
				firstAt = off
			}
			if wiped > maxWiped {
				maxWiped, maxAt = wiped, off
				worst = worst[:0]
				for y := agent; y < sc.h; y++ {
					if y < len(prevDrawn) && prevDrawn[y] && !drawn[y] {
						worst = append(worst, strconv.Itoa(y+1))
					}
				}
			}
		}
		prevDrawn = drawn
	}
	const chunk = 1 << 20
	buf := make([]byte, chunk)
	pos := from
	var carry []byte
	for pos < to {
		segEnd := to
		if next < len(marks) && marks[next].off < segEnd {
			segEnd = marks[next].off
		}
		// The segment, chunk by chunk; a CSI cut by a chunk edge is carried.
		for pos < segEnd {
			n, err := f.ReadAt(buf[:min(chunk, segEnd-pos)], int64(pos))
			if n == 0 {
				if err != nil {
					t.Fatal(err)
				}
				break
			}
			b := append(carry, buf[:n]...)
			base := pos - len(carry)
			pos += n
			cut := len(b)
			if pos < segEnd {
				if i := bytes.LastIndexByte(b, esc); i >= 0 && len(b)-i <= 64 {
					if _, _, complete := csiEnd(b, i); !complete {
						cut = i
					}
				}
			}
			carry = append([]byte(nil), b[cut:]...)
			b = b[:cut]
			at := 0
			for _, m := range endPaint.FindAllIndex(b, -1) {
				sc.feed(string(b[at:m[1]]))
				at = m[1]
				check(base + m[1])
			}
			sc.feed(string(b[at:]))
		}
		if len(carry) > 0 && pos >= segEnd {
			sc.feed(string(carry))
			carry = nil
		}
		if next < len(marks) && marks[next].off == pos {
			sc.resizeAlt(marks[next].cols, marks[next].rows)
			agent = marks[next].agent
			prevDrawn = nil
			next++
		}
	}
	t.Logf("window %d..%d of %d bytes: %d paints", from, to, st.Size(), paints)
	t.Logf("paints after which a drawn scape row was empty (outside resizes): %d, rows %d in all", wipedPaints, wipedRows)
	if wipedPaints > 0 {
		t.Logf("first at offset %d; worst %d rows at offset %d: rows %v", firstAt, maxWiped, maxAt, worst)
	}
}
