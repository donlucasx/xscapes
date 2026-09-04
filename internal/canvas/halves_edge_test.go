package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A flat cell between two half cells stays flat, and a glyph over a half cell
// yields to the halves.
//
// Both from the moon's edge, 2026-09-04. A rim cell of rgb(154,126,97) sat
// between the tip cells above and below it, whose recorded background is the
// mean of their two halves; the smoothing pulled the rim a quarter toward
// each mean and quantised the two results apart -- grey 128 over an olive, in
// the middle of a tan rim. And a far star on the tip cell above the disc's
// centre replaced the tip with one flat cell of the mean, a dark dot on the
// moon's crown.
func TestHalfCellsDoNotBleedIntoTheirNeighbours(t *testing.T) {
	sky := term.RGB{R: 48, G: 128, B: 197}
	rim := term.RGB{R: 154, G: 126, B: 97}
	c := New(3, 3, AlphaFar, AlphaMid, AlphaNear)
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c.SetBG(x, y, sky)
		}
	}
	c.SetBGHalves(1, 0, sky, rim)
	c.SetBG(1, 1, rim)
	c.SetBGHalves(1, 2, rim, sky)
	ch, fg, bg := c.ResolveAt(1, 1, term.Profile256)
	want := term.FromIndex256(rim.Index256Keeping())
	if ch == '▀' || bg != want {
		t.Errorf("the flat rim cell between two half cells resolves as %q %v/%v, want a flat %v", ch, fg, bg, want)
	}
	// A far-layer star on the top tip.
	c.Far().Plot(1, 0, '·', term.RGB{R: 220, G: 228, B: 244}, 0.3)
	ch, fg, bg = c.ResolveAt(1, 0, term.Profile256)
	if ch != '▀' {
		t.Errorf("a glyph over a half cell replaced the halves: %q %v/%v", ch, fg, bg)
	}
}
