package host

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestReplayTraceKept is an instrument: replay a real trace through the model
// with capture on, the way Host.History does, and write every row the model
// KEEPS as it leaves the band -- the rows the mirror would write -- to
// KEPT_OUT, with two counts: rows carrying a 24-letter run with no space
// (two rows interleaved character by character look like that) and rows
// that repeat. Skipped unless XSCAPES_TRACE names a trace.
//
//	XSCAPES_TRACE=/tmp/scroll.bin KEPT_OUT=/tmp/kept.txt go test ./internal/host -run TestReplayTraceKept -v
//
// First run, 2026-09-04, on the three session-13 traces in /tmp (short,
// pre-mirror): 28, 72 and 43 rows kept, none interleaved, the only repeats
// the startup banner Claude Code redraws itself. The corruption the peer
// session photographed in a long mirror-era session (rows interleaved, blocks
// duplicated, the input box four times) is NOT in these traces, so it is
// either a long-session divergence of the model or the mirror's own bytes
// being fed back through it; a trace of a session that shows it decides.
func TestReplayTraceKept(t *testing.T) {
	path := os.Getenv("XSCAPES_TRACE")
	if path == "" {
		t.Skip("set XSCAPES_TRACE")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	w, h := 120, 59
	if len(marks) > 0 {
		w, h = marks[0].cols, marks[0].rows
	}
	sc := newScreen(w, h)
	sc.capture = true
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
	out := os.Getenv("KEPT_OUT")
	if out != "" {
		os.WriteFile(out, []byte(strings.Join(kept, "\n")), 0o600)
	}
	// Signatures. A token of 24+ letters with no space is what two rows
	// interleaved look like ("TurnSilencenisonotoconditionedoon").
	long, dup := 0, 0
	seen := map[string]int{}
	for _, l := range kept {
		for _, tok := range strings.Fields(l) {
			if len(tok) >= 24 && strings.Trim(tok, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") == "" {
				long++
				break
			}
		}
		if strings.TrimSpace(l) != "" {
			seen[l]++
			if seen[l] == 2 {
				dup++
			}
		}
	}
	fmt.Printf("%s: %d rows kept; %d with a 24+ letter run; %d distinct rows that repeat\n", path, len(kept), long, dup)
}
