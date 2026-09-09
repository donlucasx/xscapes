package host

import (
	"fmt"
	"strings"
	"testing"
)

// The host's OWN repositioning scroll must not be mirrored into scrollback.
//
// His 2026-09-08 traced session, thirteen resizes in one burst: the same
// paragraph landed in the terminal's scrollback three to six times over. The
// duplicates are in our own mirror writes, byte for byte, at different rows on
// different resizes.
//
// The loop: a resize makes Rebind emit ESC[NS to undo Terminal.app's grow-push;
// h.write feeds that to the screen model like any other byte; the model's
// scrollUp() keeps each row leaving the top, "gone for good" -- and then the
// agent repaints the whole page, so the same text is back on screen ready to be
// kept again on the next resize.
//
// scrollUp's premise is right for a scroll the AGENT causes and wrong for one
// the HOST emits: there the row is not leaving, it is being moved.
func TestAResizeIsNotMirrored(t *testing.T) {
	setup := func() *screen {
		s := newScreen(40, 12)
		s.alt, s.capture = true, true
		// Twelve rows of the agent's transcript.
		var b strings.Builder
		for i := 1; i <= 12; i++ {
			b.WriteString(fmt.Sprintf("\x1b[%d;1Hagent row %d", i, i))
		}
		s.feed(b.String())
		s.takeScrolled() // drain whatever the fill itself produced
		return s
	}

	// Control: when the AGENT scrolls, the row leaving the top IS gone and must
	// be kept. Without this the test could pass by never keeping anything.
	t.Run("agent scroll is mirrored", func(t *testing.T) {
		s := setup()
		s.feed("\x1b[12;1H\n") // a newline on the last row scrolls the region
		if got := s.takeScrolled(); len(got) == 0 {
			t.Fatal("an agent scroll kept no rows -- the mirror would lose the transcript")
		}
	})

	t.Run("a shrink is not mirrored", func(t *testing.T) {
		s := setup()
		s.resizeAlt(40, 8)
		got := s.takeScrolled()
		if len(got) > 0 {
			var rows []string
			for _, r := range got {
				rows = append(rows, strings.TrimSpace(rowANSIPlain(r)))
			}
			t.Fatalf("a shrink kept %d rows for the mirror (%q) -- the agent repaints after a "+
				"resize, so every shrink would mirror rows that are about to come straight back", len(got), rows)
		}
	})
}

// rowANSIPlain is the row's text without colour, for readable failures.
func rowANSIPlain(row []cell) string {
	var b strings.Builder
	for _, c := range row {
		b.WriteRune(c.r)
	}
	return b.String()
}
