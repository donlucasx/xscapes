package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// The moon carries context remaining, which is a fact about the AGENT. Its
// visibility used to be set by the palette, which is a fact about the WORLD --
// so the context readout faded out with the daylight and was invisible for the
// whole working day. Measured against its own sky the moon sat +198 luma at
// midnight, +14 at half past eleven and +10 at noon: "I dont see the sun/moon,
// nor the shine on the water".
//
// The rule the scene is built on is that the water is the work and the sky is
// the world, and that nothing crosses. An agent channel may not be switched off
// by the clock.
func TestTheMoonIsLegibleAtEveryHour(t *testing.T) {
	for _, tod := range []float64{0, 0.125, 0.25, 0.386, 0.474, 0.5, 0.625, 0.75, 0.875} {
		sh := NewShore(7, false)
		sh.MoonX = 0.28
		c := canvas.New(120, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, 2.0, Activity{Level: 0.5, TimeOfDay: tod, ContextUsed: 0.3})

		hy := int(float64(c.H) * 0.42)
		best := 0.0
		for y := 0; y < hy; y++ {
			ref := bandLuma(c.BGAt(2, y)) // the plain sky at this row
			for x := 0; x < c.W; x++ {
				if d := bandLuma(c.BGAt(x, y)) - ref; d > best {
					best = d
				}
			}
		}
		t.Logf("tod %.3f: moon reads %+.1f luma above its sky", tod, best)
		if best < 40 {
			t.Errorf("tod %.3f: the moon is only %+.1f luma above its sky -- the context readout is invisible at that hour",
				tod, best)
		}
	}
}

// One body in the sky, not two.
//
// His call: "what if we use the moon as a sun. Same object, same variable. only
// thing we gotta change is the color (like the rest of the scape does)". So the
// disc keeps carrying context by phase and altitude at every hour, and only its
// colour moves with the clock -- warm through the day, cool at night -- exactly
// as the sky, the sea and the sand already do. A second body would have meant a
// second encoding for the same variable.
func TestTheSkyBodyIsWarmByDayAndCoolAtNight(t *testing.T) {
	warm := func(tod float64) float64 {
		m := PaletteAt(tod).Moon
		return float64(m.R) - float64(m.B) // positive is warm, negative is cool
	}
	for _, tod := range []float64{0.30, 0.42, 0.50, 0.58, 0.70} {
		if w := warm(tod); w < 60 {
			t.Errorf("tod %.2f: the sun is only %.0f warmer than it is cool -- it will not read as a sun", tod, w)
		}
	}
	for _, tod := range []float64{0.0, 0.95} {
		if w := warm(tod); w > 0 {
			t.Errorf("tod %.2f: the moon is %.0f warm -- at night it should be cool, not golden", tod, w)
		}
	}
}
