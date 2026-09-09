package host

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestReplayTraceRetainDiff is the instrument session 25 asked for: "wire it
// and replay".
//
// screen.retainWidth models Terminal.app's measured width rule -- a row keeps
// every cell it has ever had, at the widest the window has been -- and until
// today it was true only in tests, while reallocBand ran with it ON. The model
// that decides what gets mirrored into SCROLLBACK is the one that had it off,
// so the question is what turning it on actually changes there. reallocBand
// does not settle it either way: it deletes rows agentRows+1..rows, which is
// the SCAPE's band, so the AGENT's rows -- the ones this model captures -- keep
// their retained tails in the real terminal.
//
// Replays ONE window of a real trace twice, once each way, and diffs the rows
// the mirror would write.
//
//	XSCAPES_TRACE=~/.config/xscapes/traces/20260908-102733.bin \
//	TRACE_BYTES=130000000 go test ./internal/host -run RetainDiff -v
//
// ⚠ Read the counts before believing the verdict. A window with no complete
// mirror block in it reports "identical" for the same reason an empty one
// does -- the trap that voided the first strikethrough comparison on
// 2026-09-09. This prints kept-row counts and the resizes it applied, and
// FAILS if the window carried no width change, because a width rule cannot be
// exercised without one.
func TestReplayTraceRetainDiff(t *testing.T) {
	path := tracePath(t)
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace FILE to replay it")
	}
	limit := int64(1 << 62)
	if s := os.Getenv("TRACE_BYTES"); s != "" {
		if _, err := fmt.Sscanf(s, "%d", &limit); err != nil {
			t.Fatalf("TRACE_BYTES=%q: want a byte count", s)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	widths := 0
	for i := 1; i < len(marks) && int64(marks[i].off) <= int64(len(b)); i++ {
		if marks[i].cols != marks[i-1].cols {
			widths++
		}
	}
	t.Logf("replaying %d bytes of %s, %d marks in range, %d of them WIDTH changes",
		len(b), path, len(marks), widths)
	if widths == 0 {
		t.Fatalf("this window carries no width change, so it cannot exercise the width rule at all -- "+
			"a clean result here would mean nothing. Marks in the whole trace: %d", len(marks))
	}

	off := replayKept(b, marks, false)
	on := replayKept(b, marks, true)
	t.Logf("kept rows: retainWidth=off %d, retainWidth=on %d", len(off), len(on))

	diff, shown := 0, 0
	for i := 0; i < len(off) && i < len(on); i++ {
		if off[i] == on[i] {
			continue
		}
		diff++
		if shown < 8 {
			shown++
			t.Logf("row %d differs:\n  off |%s|\n  on  |%s|", i, off[i], on[i])
		}
	}
	t.Logf("DIFF: %d of %d rows differ (%d rows only one side kept)",
		diff, min(len(off), len(on)), abs(len(off)-len(on)))
}

// replayKept feeds a trace through the model the way Host.History does and
// returns every row the mirror would have written, trimmed of trailing blanks.
func replayKept(b []byte, marks []traceMark, retain bool) []string {
	w, h := 120, 59
	if len(marks) > 0 {
		w, h = marks[0].cols, marks[0].rows
	}
	sc := newScreen(w, h)
	sc.capture = true
	sc.retainWidth = retain
	var kept []string
	drain := func() {
		for _, row := range sc.takeScrolled() {
			var sb strings.Builder
			for _, c := range row {
				sb.WriteRune(c.r)
			}
			kept = append(kept, strings.TrimRight(sb.String(), " "))
		}
	}
	fed := 0
	const chunk = 2048
	next := 1
	for fed < len(b) {
		end := min(fed+chunk, len(b))
		if next < len(marks) && marks[next].off <= end {
			end = marks[next].off
		}
		sc.feed(string(b[fed:end]))
		drain()
		fed = end
		if next < len(marks) && marks[next].off == fed {
			sc.resizeAlt(marks[next].cols, marks[next].rows)
			drain()
			next++
		}
	}
	return kept
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
