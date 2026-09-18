package host

import "testing"

// TestReallocBandLeavesTheAgentWhereItWas: re-allocating the band on a width
// change must hand the terminal back exactly as it found it, the band pinned
// with origin mode on and the agent's cursor on its own row. It used to end
// with the region reset, origin mode off and the cursor on the scape's
// first row (his trace of 2026-09-18 at a size mark: ESC[?6l ESC[32;55r
// ESC[32;1H ESC[24M ESC[r, then the next paint's ESC7 saving row 32). In
// that gap an agent's scroll moved the whole screen, and the paint's restore
// under origin mode put the agent's cursor at the band's top-left, where its
// next output landed (Kimi's assessment, 2026-09-18, and the s38 audit).
func TestReallocBandLeavesTheAgentWhereItWas(t *testing.T) {
	const W, H, band = 60, 24, 14
	s := newScreen(W, H)
	s.feed(EnterBand(band))
	s.feed("\x1b[5;8H")      // the agent's cursor: row 5, column 8, inside the band
	s.feed("\x1b[38;5;208m") // and a colour of its own
	s.feed(reallocBand(band+1, H))
	if s.y != 4 || s.x != 7 {
		t.Errorf("the agent's cursor moved to row %d column %d (0-based), want row 4 column 7", s.y, s.x)
	}
	if !s.origin {
		t.Error("origin mode is off after the re-allocation")
	}
	if s.top != 0 || s.bot != band-1 {
		t.Errorf("the scroll region is %d..%d (0-based), want 0..%d: the band is not pinned", s.top, s.bot, band-1)
	}
	if s.curFG != s.sFG || s.curFG == (s.blank().fg) {
		t.Errorf("the agent's colour was not restored")
	}
}
