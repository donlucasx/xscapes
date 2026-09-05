package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A glyph plotted with its OWN background keeps it, over a flat cell, over a
// ramp cell and over a split cell, on both profiles. The companion's eyes are
// the reason: they are characters plotted in gaps of the body bitmap, so their
// cell showed the SCENE behind the head -- two holes to the sea by day.
func TestAGlyphWithItsOwnBackgroundKeepsIt(t *testing.T) {
	sea := term.RGB{R: 20, G: 90, B: 160}
	fur := term.RGB{R: 236, G: 228, B: 210}
	for _, p := range []term.Profile{term.ProfileTrueColor, term.Profile256} {
		c := New(6, 3, AlphaFar, AlphaMid, AlphaNear)
		for i := range c.BG {
			c.BG[i] = sea
		}
		ramp := term.NewRamp(term.RGB{R: 0, G: 60, B: 140}, term.RGB{R: 60, G: 140, B: 220})
		c.SetBGRamp(1, 1, ramp, 0.3, 0.6)
		c.SetBGHalves(2, 1, term.RGB{R: 0, G: 60, B: 140}, term.RGB{R: 60, G: 140, B: 220})
		c.Near().PlotOn(0, 1, 'o', term.RGB{R: 168, G: 236, B: 176}, fur, 1)
		c.Near().PlotOn(1, 1, 'o', term.RGB{R: 168, G: 236, B: 176}, fur, 1)
		c.Near().PlotOn(2, 1, 'o', term.RGB{R: 168, G: 236, B: 176}, fur, 1)
		c.Near().Plot(3, 1, 'o', term.RGB{R: 168, G: 236, B: 176}, 1)
		want := p.Quantise(fur, false)
		for x := 0; x < 3; x++ {
			r, _, bg := c.ResolveAt(x, 1, p)
			if r != 'o' || bg != want {
				t.Errorf("%v cell %d: glyph %q on %v, want 'o' on %v", p, x, r, bg, want)
			}
		}
		// The control: a plain Plot still shows the scene behind it.
		if _, _, bg := c.ResolveAt(3, 1, p); bg != p.Quantise(sea, false) {
			t.Errorf("%v control cell: bg %v, want the sea %v", p, bg, p.Quantise(sea, false))
		}
	}
}
