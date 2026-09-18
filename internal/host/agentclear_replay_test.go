package host

import (
	"os"
	"strconv"
	"testing"
)

// TestReplayTraceAgentClears replays a window of a real trace through the
// model twice -- the bytes as they went to the terminal, and the same bytes
// with every erase-display confined the way Filter now confines them -- and
// counts, right after each erase, how many of the scape's rows are blank.
// His trace of 2026-09-18 (Kimi Code CLI, a window drag): every one of the
// agent's clears blanked every scape row; confined, none do.
//
//	XSCAPES_TRACE=~/.config/xscapes/traces/<t>.bin TRACE_FROM=49500000 TRACE_TO=51400000 \
//	  go test ./internal/host -run TestReplayTraceAgentClears -v
func TestReplayTraceAgentClears(t *testing.T) {
	path := traceEnv
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace FILE to replay it")
	}
	all, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	if len(marks) == 0 {
		t.Fatal("no marks: the trace's .log is missing")
	}
	from, to := 0, len(all)
	if v, err := strconv.Atoi(os.Getenv("TRACE_FROM")); err == nil {
		from = v
	}
	if v, err := strconv.Atoi(os.Getenv("TRACE_TO")); err == nil && v < to {
		to = v
	}
	// The size in force at `from` is the last mark at or before it.
	cur := marks[0]
	next := 1
	for next < len(marks) && marks[next].off <= from {
		cur = marks[next]
		next++
	}
	type result struct{ clears, blankRows, scapeRows int }
	run := func(confine bool) result {
		sc := newScreen(cur.cols, cur.rows)
		agent := cur.agent
		f := &Filter{}
		f.Band.Store(int32(agent))
		var r result
		blankScape := func() int {
			n := 0
			for y := agent; y < sc.h; y++ {
				if sc.rowAt(y) == "" {
					n++
				}
			}
			return n
		}
		nx := next
		i := from
		for i < to {
			// Up to the next mark, or the end of the window.
			end := to
			if nx < len(marks) && marks[nx].off < end {
				end = marks[nx].off
			}
			b := all[i:end]
			// Feed byte by byte at the sequence level: each erase is
			// rewritten (or not) and the scape counted right after it.
			for k := 0; k < len(b); {
				if b[k] != esc {
					j := k
					for j < len(b) && b[j] != esc {
						j++
					}
					sc.feed(string(b[k:j]))
					k = j
					continue
				}
				e, final, complete := csiEnd(b, k)
				if !complete {
					sc.feed(string(b[k:]))
					break
				}
				seq := b[k : e+1]
				if final == 'J' {
					r.clears++
					if confine {
						seq = f.confineErase(seq)
					}
					sc.feed(string(seq))
					r.blankRows += blankScape()
					r.scapeRows += sc.h - agent
				} else {
					sc.feed(string(seq))
				}
				k = e + 1
			}
			i = end
			if nx < len(marks) && marks[nx].off == i {
				sc.resizeAlt(marks[nx].cols, marks[nx].rows)
				agent = marks[nx].agent
				f.Band.Store(int32(agent))
				nx++
			}
		}
		return r
	}
	raw, confined := run(false), run(true)
	t.Logf("window %d..%d of %d bytes, %d erase-display sequences", from, to, len(all), raw.clears)
	t.Logf("raw:      blank scape rows right after an erase: %d of %d", raw.blankRows, raw.scapeRows)
	t.Logf("confined: blank scape rows right after an erase: %d of %d", confined.blankRows, confined.scapeRows)
	if raw.clears == 0 {
		t.Skip("no erase in this window: nothing to compare")
	}
	if confined.blankRows != 0 {
		t.Errorf("confined erases still blank %d scape rows", confined.blankRows)
	}
}
