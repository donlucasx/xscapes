package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// Under term.LowerHalf the same split cell is drawn the other way up: U+2584
// with the LOWER colour in the foreground and the upper in the background --
// the two halves keep their colours, only the glyph and the order change.
// Terminal.app's U+2580 leaves a hairline above; its U+2584 does not.
func TestASplitCellFlipsUnderLowerHalf(t *testing.T) {
	was := term.LowerHalf
	defer func() { term.LowerHalf = was }()
	sky := term.RGB{R: 0, G: 95, B: 175}
	moon := term.RGB{R: 215, G: 175, B: 135}
	for _, lower := range []bool{false, true} {
		term.LowerHalf = lower
		c := New(3, 3, AlphaFar, AlphaMid, AlphaNear)
		c.SetBGHalves(1, 1, sky, moon)
		ch, fg, bg := c.ResolveAt(1, 1, term.Profile256)
		wantCh, wantFG, wantBG := '▀', sky, moon
		if lower {
			wantCh, wantFG, wantBG = '▄', moon, sky
		}
		if ch != wantCh || fg != wantFG || bg != wantBG {
			t.Errorf("LowerHalf=%v: got %q fg %v bg %v, want %q fg %v bg %v", lower, ch, fg, bg, wantCh, wantFG, wantBG)
		}
	}
}
