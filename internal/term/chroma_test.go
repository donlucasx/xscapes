package term

import "testing"

// The boost exists to get glyphs off the greyscale ramp on a 256-colour
// terminal. It must not do that by recolouring things that are meant to be
// white -- the companion is the obvious one, and a peach cat is worse than a
// grey sea.
func TestBoostLeavesNeutralsAlone(t *testing.T) {
	neutral := []struct {
		n string
		c RGB
	}{
		{"cat fur", RGB{236, 228, 210}},
		{"moon", RGB{240, 236, 214}},
		{"foam", RGB{214, 232, 238}},
	}
	for _, e := range neutral {
		for _, k := range []float64{1.5, 2.2, 3.0, 6.0} {
			if got := e.c.Saturate(k); got != e.c {
				t.Errorf("%s moved at %.1fx: rgb(%d,%d,%d) -> rgb(%d,%d,%d)",
					e.n, k, e.c.R, e.c.G, e.c.B, got.R, got.G, got.B)
			}
		}
	}
}

// And it must actually work on the things that are meant to be colours.
func TestBoostRescuesColoursFromTheGreyRamp(t *testing.T) {
	for _, e := range []struct {
		n string
		c RGB
	}{
		{"sea glyph", RGB{120, 170, 200}},
		{"deep sea glyph", RGB{70, 110, 150}},
	} {
		if e.c.Index256() < 232 {
			continue // already fine unboosted
		}
		if got := e.c.Saturate(GlyphBoost).Index256(); got >= 232 {
			t.Errorf("%s still lands on the grey ramp after the boost: index %d", e.n, got)
		}
	}
}

// Brightness must broadly survive the boost, or the scene changes exposure
// rather than gaining colour.
//
// The tolerance is 20 rather than 0 because a colour already near the edge of
// the gamut cannot gain chroma without clipping: the worried amber pushes red
// past 255, the clamp holds it, and the result lands 15.5 luma darker. That is
// a bounded, one-sided cost on the most saturated colours in the palette, and
// worth it -- but it should stay bounded, which is what this asserts.
func TestBoostPreservesBrightness(t *testing.T) {
	luma := func(c RGB) float64 {
		return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	}
	for _, c := range []RGB{{120, 170, 200}, {244, 176, 96}, {70, 110, 150}} {
		before, after := luma(c), luma(c.Saturate(GlyphBoost))
		if d := after - before; d > 20 || d < -20 {
			t.Errorf("rgb(%d,%d,%d) changed brightness by %.1f", c.R, c.G, c.B, d)
		}
	}
}
