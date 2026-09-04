package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A cell painted as two halves and then repainted as part of a ramp, or as
// one flat colour, is the ramp cell or the flat cell. The half-row record does
// not outlive the paint that set it -- see scape.TestTheMoonLeavesNoStaleTipsWhenItMoves
// for what it looked like when it did.
func TestHalvesDoNotOutliveARepaint(t *testing.T) {
	sky, moon := term.RGB{R: 0, G: 95, B: 175}, term.RGB{R: 215, G: 175, B: 135}
	c := New(4, 2, AlphaFar, AlphaMid, AlphaNear)
	c.SetBGHalves(1, 0, sky, moon)
	if ch, _, _ := c.ResolveAt(1, 0, term.Profile256); ch != '▀' {
		t.Fatalf("a half cell should resolve as U+2580, got %q", ch)
	}
	r := term.NewRamp(sky, term.RGB{R: 175, G: 215, B: 255})
	c.SetBGRamp(1, 0, r, 0, 0.1)
	if ch, fg, bg := c.ResolveAt(1, 0, term.Profile256); ch == '▀' && fg != bg && (fg == moon || bg == moon) {
		t.Errorf("after a ramp repaint the cell still shows the old halves: %q %v/%v", ch, fg, bg)
	}
	if got := c.BGAt(1, 0); got != r.True(0.05) {
		t.Errorf("after a ramp repaint BGAt is %v, want the ramp's true colour %v", got, r.True(0.05))
	}
	c.SetBGHalves(2, 0, sky, moon)
	c.SetBG(2, 0, sky)
	if ch, _, bg := c.ResolveAt(2, 0, term.Profile256); ch == '▀' || bg == term.FromIndex256(moon.Index256Keeping()) {
		t.Errorf("after a flat repaint the cell still shows the old halves: %q bg %v", ch, bg)
	}
}
