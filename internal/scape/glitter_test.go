package scape

import (
	"testing"

	"github.com/donlucasx/asciiscapes/internal/canvas"
)

// The moonlight on the water must sit under the moon. It used to be pinned to a
// hardcoded 0.72 of the width, so when the composition mirrored and the moon
// moved to 0.28 the shine stayed on the far side -- a reflection with no source.
func TestMoonlightFollowsTheMoon(t *testing.T) {
	for _, frac := range []float64{0.28, 0.5, 0.72} {
		c := canvas.New(100, 30, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(7, false)
		sh.MoonX = frac
		sh.Update(c, 3, Activity{Working: true, Level: 0.4})
		mx, _ := sh.MoonPos()

		// Measure the moonlight glyph specifically -- the sea draws on the
		// same layer, and averaging the whole layer just finds the middle of
		// the screen. (The first version of this test did exactly that and
		// reported 45 for every moon position, which is the giveaway.)
		sum, n := 0, 0
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				if cell := c.Mid().Cells[y*c.W+x]; cell.Set && cell.R == '•' {
					sum += x
					n++
				}
			}
		}
		if n == 0 {
			t.Fatalf("moon at %.2f: no moonlight drawn at all", frac)
		}
		mean := sum / n
		if d := mean - mx; d > 6 || d < -6 {
			t.Errorf("moon at %.2f sits at column %d but its moonlight centres on %d (%d columns away)",
				frac, mx, mean, d)
		}
	}
}

// A dark moon lays no path across the water.
func TestMoonlightDimsWithTheMoon(t *testing.T) {
	bright := glitterMass(t, 0.0)
	dark := glitterMass(t, 0.98)
	if dark >= bright {
		t.Errorf("a nearly-unlit moon shines as hard as a full one: %.2f vs %.2f", dark, bright)
	}
}

func glitterMass(t *testing.T, ctxUsed float64) float64 {
	t.Helper()
	c := canvas.New(100, 30, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	sh.Update(c, 3, Activity{Working: true, Level: 0.4, ContextUsed: ctxUsed})
	var sum float64
	for _, cell := range c.Mid().Cells {
		if cell.Set && cell.R == '•' {
			sum += cell.A
		}
	}
	return sum
}
