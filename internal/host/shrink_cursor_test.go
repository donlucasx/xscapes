package host

import (
	"fmt"
	"strings"
	"testing"
)

// The split input box, stated directly.
//
// Measured 2026-09-03 with notes/shrinkprobe, eleven geometries: on the
// alternate screen a shrink pulls the CONTENT up by the rows lost and leaves
// the CURSOR on its absolute row. Claude Code keeps its input box on the last
// rows of its viewport and draws it purely by relative moves from the cursor,
// so after a shrink it draws one row too low per row lost -- and once the band
// has shrunk past the cursor, the host's own restore lands on row 1 and the
// box is painted at the TOP of the band over the transcript. Production did
// that on every shrink, by the tick size for a one- or two-row drag tick.
//
// What the host must do instead: put the text's bottom back on the band's
// bottom (the scape's share of the shrink is blank rows scrolled in at the
// top) and move the cursor up by the band's own shrink, all while the region
// is still the full screen, so the restore never lands outside it.
func TestShrinkKeepsTheInputBoxUnderTheCursor(t *testing.T) {
	const w = 120
	for _, tc := range []struct {
		name  string
		start int
		steps []int // heights after each tick
	}{
		{"one row", 30, []int{29}},
		{"two rows", 30, []int{28}},
		{"three rows", 30, []int{27}},
		{"six rows", 30, []int{24}},
		{"a drag, three ticks of two", 30, []int{28, 26, 24}},
		{"one row the band absorbs", 32, []int{31}},
		{"shrink then grow", 30, []int{24, 30}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := newScreen(w, tc.start)
			agent, _ := Band(tc.start)
			sc.feed(Open(true, agent, tc.start))
			drawAgentBox(sc, agent)
			if got := sc.y + 1; got != agent-1 {
				t.Fatalf("setup: cursor on row %d, want %d (the input row)", got, agent-1)
			}

			rows, band := tc.start, agent
			for _, h := range tc.steps {
				sc.resizeAlt(w, h) // the terminal
				next, _ := Band(h)
				sc.feed(resizeSequence(true, AppleTerminalRules, rows, h, band, next)) // the host's tick
				rows, band = h, next
			}
			// Claude redraws its box from wherever the cursor is, relatively.
			sc.feed("\r\x1b[1A\x1b[2K+==== new box ====+\r\n\x1b[2K| > hi again      |\r\n" +
				"\x1b[2K+=================+\r\x1b[1A\x1b[13C")

			// The old box must be gone entirely -- overwritten, not sitting a
			// few rows above -- and the new one must sit where the old one's
			// content now is. After a pure shrink that is the band's last rows;
			// after a grow that follows, the box stays put and the band grows
			// under it (measured: 16 -> 13 -> 13).
			wantTop := band - 2
			if last := tc.steps[len(tc.steps)-1]; last > tc.steps[0] || (len(tc.steps) > 1 && tc.steps[len(tc.steps)-1] > tc.steps[len(tc.steps)-2]) {
				prevBand, _ := Band(tc.steps[len(tc.steps)-2])
				wantTop = prevBand - 2
			}
			var old, boxTop []int
			for y := 0; y < sc.h; y++ {
				r := sc.rowAt(y)
				if strings.Contains(r, "old box") || strings.Contains(r, "| > hi   ") {
					old = append(old, y+1)
				}
				if strings.Contains(r, "new box") {
					boxTop = append(boxTop, y+1)
				}
			}
			if len(old) > 0 {
				t.Errorf("rows %v still hold the OLD box: the redraw landed somewhere else, which is the split box", old)
			}
			if len(boxTop) != 1 || boxTop[0] != wantTop {
				t.Errorf("new box top at rows %v, want row %d (band 1..%d)", boxTop, wantTop, band)
			}
			if got := sc.y + 1; got != wantTop+1 {
				t.Errorf("cursor on row %d after the redraw, want %d", got, wantTop+1)
			}
			if y := sc.y; y < sc.top || y > sc.bot {
				t.Errorf("cursor row %d is outside the band 1..%d", y+1, sc.bot+1)
			}
		})
	}
}

// drawAgentBox puts on the model what Claude Code leaves on screen in an
// established session: transcript rows filling the band and a three-row input
// box on its last rows, with the cursor after the prompt in the middle row.
// Origin mode is on, so row 1 is the band's row 1.
func drawAgentBox(sc *screen, band int) {
	for r := 1; r <= band-3; r++ {
		sc.feed(fmt.Sprintf("\x1b[%d;1HTRANSCRIPT ROW %02d", r, r))
	}
	sc.feed(fmt.Sprintf("\x1b[%d;1H+---- old box ----+", band-2))
	sc.feed(fmt.Sprintf("\x1b[%d;1H| > hi            |", band-1))
	sc.feed(fmt.Sprintf("\x1b[%d;1H+-----------------+", band))
	sc.feed(fmt.Sprintf("\x1b[%d;7H", band-1))
}

// The model must reproduce the measured restore: a saved row outside the new
// region lands on row 1. Without this the test above cannot see the defect,
// which is how it stayed invisible through session 13.
func TestTheModelHomesARestoreIntoARegionThatLostTheRow(t *testing.T) {
	sc := newScreen(40, 30)
	sc.feed("\x1b[1;17r\x1b[?6h\x1b[16;5H\x1b7")     // save on row 16 of a 17-row band
	sc.feed("\x1b[?6l\x1b[r\x1b[1;14r\x1b[?6h\x1b8") // restore into a 14-row band
	if got := sc.y + 1; got != 1 {
		t.Errorf("restore landed on row %d, measured row 1", got)
	}
	sc.feed("\x1b[?6l\x1b[r\x1b[16;5H\x1b7\x1b[1;17r\x1b[?6h\x1b8") // and a row the band holds
	if got := sc.y + 1; got != 16 {
		t.Errorf("restore into a band that holds the row landed on %d, want 16", got)
	}
}

// And the model must leave the cursor alone on an alternate-screen resize,
// clamping only when the screen shrinks past it (measured 16 -> 16 on a
// six-row shrink, 16 -> 10 on a shrink to ten rows).
func TestTheModelLeavesTheCursorOnAnAltResize(t *testing.T) {
	sc := newScreen(40, 30)
	sc.feed("\x1b[16;5H")
	sc.resizeAlt(40, 24)
	if got := sc.y + 1; got != 16 {
		t.Errorf("shrink moved the cursor to row %d; the terminal leaves it on 16", got)
	}
	sc.resizeAlt(40, 30)
	if got := sc.y + 1; got != 16 {
		t.Errorf("grow moved the cursor to row %d; the terminal leaves it on 16", got)
	}
	sc.resizeAlt(40, 10)
	if got := sc.y + 1; got != 10 {
		t.Errorf("a shrink past the cursor left it on row %d; the terminal clamps to 10", got)
	}
}
