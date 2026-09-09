package host

import (
	"strings"
	"testing"
)

// The mirror's model does NOT take Terminal.app's width rule, and that is a
// decision, not the oversight it looks like.
//
// screen.retainWidth models what the terminal really does -- a row keeps every
// cell it has ever had, at the widest the window has been (notes/width-audit.md)
// -- and Rules.RetainsWidth is honoured by reallocBand. Session 25 found that
// the mirror's model never connects the two and called it "a wiring gap and it
// is in the mirror path -- the strongest lead yet".
//
// Wired and replayed (TestReplayTraceRetainDiff), it is the opposite of a fix.
// Two independent windows of his own traces, thirteen and twelve width changes:
// 31 of 975 mirrored rows and 17 of 1472 come out MERGED, a fragment of the
// pre-resize row stitched onto the end of the current one --
//
//	don't scroll, don't resize, don't exit.files while you
//
// -- which is the exact signature of the scrollback corruption that took seven
// sessions to close. Faithful to the terminal and wrong for the mirror, for two
// reasons that both hold: the retained cells are stale content the agent has
// already redrawn somewhere else, and a mirrored line longer than the window
// WRAPS in the main buffer, so it corrupts the row under it as well.
//
// reallocBand does not settle this either way, and it is worth saying so
// because it looks like it should: it deletes rows agentRows+1..rows, which is
// the SCAPE's band. The AGENT's rows keep their tails in the real terminal.
func TestTheMirrorDropsTheCellsTheTerminalRetains(t *testing.T) {
	const wide, narrow, rows, band = 40, 18, 4, 3
	// The same feed both ways: fill the band at the wide size, narrow, then
	// push the rows out of the top so they are kept for the mirror.
	play := func(retain bool) []string {
		sc := newScreen(wide, rows)
		sc.capture = true
		sc.retainWidth = retain
		sc.feed(Open(true, band, rows))
		sc.feed(strings.Repeat("W", wide) + "\r\n")
		sc.feed(strings.Repeat("W", wide) + "\r\n")
		sc.resizeAlt(narrow, rows)
		for i := 0; i < 4; i++ {
			sc.feed("fresh\r\n")
		}
		var kept []string
		for _, row := range sc.takeScrolled() {
			var sb strings.Builder
			for _, c := range row {
				sb.WriteRune(c.r)
			}
			kept = append(kept, strings.TrimRight(sb.String(), " "))
		}
		return kept
	}

	off := play(false)
	if len(off) == 0 {
		t.Fatal("nothing was kept, so this proves nothing -- the feed never scrolled a row out of the band")
	}
	for i, row := range off {
		if len(row) > narrow {
			t.Errorf("kept row %d is %d cells wide in an %d-column window: %q\n"+
				"a mirrored row wider than the window wraps in the main buffer and takes the row under it",
				i, len(row), narrow, row)
		}
	}

	// And the other way, so the reason this flag stays off is on the record
	// rather than in a comment: the same feed keeps the stale tail.
	on := play(true)
	tails := 0
	for _, row := range on {
		if len(row) > narrow {
			tails++
		}
	}
	if tails == 0 {
		t.Fatalf("retainWidth=true kept no over-wide row, so this test no longer demonstrates what it claims:\n%q", on)
	}
	t.Logf("retainWidth=true keeps %d of %d rows carrying cells past the window: %q", tails, len(on), on)
}
