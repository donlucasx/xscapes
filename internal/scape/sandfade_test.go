package scape

import (
	"testing"

	"github.com/donlucasx/asciiscapes/internal/term"
)

// The activity tail fades toward the beach as the tide takes it. The target has
// to be the beach's ACTUAL colour this frame, not a constant: it used to be
// pinned to midnight's sand, so by mid-evening the oldest line sat 2.7 luma
// from the sand it was written on and simply could not be read.
func TestSandFadeTracksTheHour(t *testing.T) {
	luma := func(c term.RGB) float64 {
		return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	}

	for _, tod := range []float64{0.0, 0.25, 0.5, 0.75, 0.888} {
		sh := NewShore(7, false)
		sh.pal = PaletteAt(tod)
		beach := sh.SandColor()

		// Mirror what drawSand does: pick the ink direction from the beach.
		ink := term.RGB{R: 244, G: 236, B: 220}
		if luma(beach) > 140 {
			ink = term.RGB{R: 34, G: 26, B: 20}
		}
		prev := 1e9
		for _, age := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
			got := luma(term.Lerp(ink, beach, 0.10+0.62*age))
			contrast := got - luma(beach)
			if contrast < 0 {
				contrast = -contrast
			}
			// Every line must stay readable against the sand it sits on...
			if contrast < 18 {
				t.Errorf("tod %.3f age %.2f: only %.1f luma from the beach -- unreadable", tod, age, contrast)
			}
			// ...and older must never be more visible than newer.
			if contrast > prev {
				t.Errorf("tod %.3f age %.2f: contrast %.1f exceeds the younger line's %.1f", tod, age, contrast, prev)
			}
			prev = contrast
		}
	}
}
