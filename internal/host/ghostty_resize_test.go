package host

import (
	"fmt"
	"strings"
	"testing"
)

// resizeGhostty is Ghostty's alternate screen on a resize, from its source
// (see Rules): a shrink scrolls rows off the top and the cursor follows its
// row; a grow with the cursor above the last row keeps the content at the top
// and appends blank rows, cursor unchanged.
func (s *screen) resizeGhostty(w, h int) {
	if h < s.h {
		s.resizeScrolling(w, h)
		return
	}
	s.resize(w, h)
}

// The split input box, on Ghostty. Same shape as the Terminal.app test, with
// the terminal's own rules in the model: DECRC restores the row verbatim, and
// the resize moves content and cursor the way PageList.zig says. The scape is
// repainted in full after every tick, as the host does (dmg.reset), so a sky
// row left inside the band is the host's doing and not a stale frame.
//
// Red on 2026-09-04 under Terminal.app's sequences: a four-row grow left the
// old box's rows on screen, lost transcript rows 1-4 and dragged a sky row
// into the band; a six-row shrink put the box six rows above the band's
// bottom. Green with the rules picked by RulesFor("ghostty").
func TestGhosttyResizeKeepsTheInputBoxUnderTheCursor(t *testing.T) {
	const w = 120
	const scapeBG = 25
	paintScape := func(sc *screen, agent, rows int) {
		var b strings.Builder
		b.WriteString(BeginPaint())
		for r := agent + 1; r <= rows; r++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[48;5;%dm%s\x1b[0m", r, scapeBG, strings.Repeat(" ", w))
		}
		b.WriteString(EndPaint(agent))
		sc.feed(b.String())
	}
	for _, tc := range []struct {
		name  string
		start int
		steps []int
	}{
		{"grow one row", 56, []int{57}},
		{"grow four rows", 56, []int{60}},
		{"a drag up, three ticks of two", 56, []int{58, 60, 62}},
		{"shrink one row", 60, []int{59}},
		{"shrink six rows", 60, []int{54}},
		{"a drag down, three ticks of two", 60, []int{58, 56, 54}},
		{"shrink then grow back", 60, []int{54, 60}},
		{"grow then shrink back", 56, []int{60, 56}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := newScreen(w, tc.start)
			sc.restoreAbsolute = true
			agent, _ := Band(tc.start)
			sc.feed(Open(true, agent, tc.start))
			drawAgentBox(sc, agent)
			paintScape(sc, agent, tc.start)
			// A grow leaves the text where it is and the band grows under it;
			// a shrink moves the text up by the band's own shrink, so the box
			// keeps its distance from the band's bottom. Same rule as the
			// Terminal.app sequences produce (the probe of 2026-09-04 put
			// both at row 27 after a grow of four and a shrink of four).
			rows, band, boxTop := tc.start, agent, agent-2
			for _, h := range tc.steps {
				sc.resizeGhostty(w, h)
				next, _ := Band(h)
				sc.feed(resizeSequence(true, RulesFor("ghostty"), rows, h, band, next))
				if h < rows {
					boxTop -= band - next
				}
				rows, band = h, next
				paintScape(sc, band, rows)
			}
			sc.feed("\r\x1b[1A\x1b[2K+==== new box ====+\r\n\x1b[2K| > hi again      |\r\n" +
				"\x1b[2K+=================+\r\x1b[1A\x1b[13C")

			var old, box, scapeInBand []int
			for y := 0; y < sc.h; y++ {
				r := sc.rowAt(y)
				if strings.Contains(r, "old box") || strings.Contains(r, "| > hi   ") {
					old = append(old, y+1)
				}
				if strings.Contains(r, "new box") {
					box = append(box, y+1)
				}
				if bg, uniform := sc.bgRunAt(y); y < band && uniform && bg == scapeBG {
					scapeInBand = append(scapeInBand, y+1)
				}
			}
			if len(old) > 0 {
				t.Errorf("rows %v still hold the OLD box", old)
			}
			if len(box) != 1 || box[0] != boxTop {
				t.Errorf("new box top at rows %v, want row %d (band 1..%d of %d)", box, boxTop, band, rows)
			}
			if got := sc.y + 1; got != boxTop+1 {
				t.Errorf("cursor on row %d after the redraw, want %d", got, boxTop+1)
			}
			if len(scapeInBand) > 0 {
				t.Errorf("band rows %v are painted in the scape's colour", scapeInBand)
			}
			// A grow destroys nothing: every transcript row is still there.
			if last := tc.steps[len(tc.steps)-1]; len(tc.steps) == 1 && last > tc.start {
				for r := 1; r <= agent-3; r++ {
					if !strings.Contains(sc.rowAt(r-1), fmt.Sprintf("TRANSCRIPT ROW %02d", r)) {
						t.Errorf("transcript row %d is not on row %d after a grow", r, r)
					}
				}
			}
		})
	}
}
