package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The sky and the sea are painted as ONE path through the 256 palette, not as
// a true-colour ramp rounded row by row.
//
// Rounding each row on its own is what produced the edge he asked about
// ("some of the 256 translations on the terminal are abrupt"). Measured at his
// geometry before this test existed: at 05:00 the sky's twelve rows went
// blue, grey, teal, grey, grey, grey, olive, grey, rose -- the true ramp passes
// close to neutral in the middle and the nearest palette entry flips between
// the grey ramp and whichever cube colour is a hair closer. Five crossings
// between grey and colour, and a 30-point CIE76 step at each one. No choice of
// endpoints fixes that: the flips happen BETWEEN them.
//
// A path cannot do it. Once a tone is left it is not returned to, and a ramp
// between two colours enters the greys at most once and leaves them at most
// once. Read from the RENDERED frame, at half-row resolution, at the size his
// window gives the scape: what the terminal shows, not what a function called
// here would return.

// tonesDown reads column x of the rendered 256 frame at half-row resolution.
func tonesDown(c *canvas.Canvas, x int) []term.RGB {
	var tones []term.RGB
	for y := 0; y < c.H; y++ {
		ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
		up, dn := bg, bg
		if ch == '▀' {
			up = fg
		}
		tones = append(tones, up, dn)
	}
	return tones
}

// skyAndSea renders one hour at his geometry and says where the sky ends and
// the open sea ends, in rows.
func skyAndSea(tod float64) (c *canvas.Canvas, hy, seaEnd int) {
	c = canvas.New(124, 27, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	sh.MoonX = 0.28
	for k := 0; k < 14; k++ {
		sh.Update(c, 3.0+float64(k)/20, Activity{Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3})
	}
	drop := 0.0
	for y := 1; y < c.H*3/4; y++ {
		if d := lumaOf(c.BGAt(2, y-1)) - lumaOf(c.BGAt(2, y)); d > drop {
			drop, hy = d, y
		}
	}
	// The last two rows above the sand are the waterline cell and its
	// neighbour, which blend toward wet sand and are not the depth ramp.
	seaEnd = sh.SandTop() - 2
	if seaEnd <= hy+1 {
		seaEnd = hy + 1
	}
	return c, hy, seaEnd
}

// isGrey is by what the eye sees, not by index: the cube's own diagonal
// (#5f5f5f, #878787, #afafaf ...) is grey on the screen, and a walk stepping
// from the ramp's #8a8a8a onto the cube's #878787 has not crossed anything.
func isGrey(c term.RGB) bool {
	mx, mn := c.R, c.R
	for _, v := range []uint8{c.G, c.B} {
		if v > mx {
			mx = v
		}
		if v < mn {
			mn = v
		}
	}
	return mx-mn <= 10
}

func TestTheSkyAndSeaAreEachOnePathThroughThePalette(t *testing.T) {
	for i := 0; i < 48; i++ {
		tod := float64(i) / 48
		c, hy, seaEnd := skyAndSea(tod)
		tones := tonesDown(c, 2)
		regions := []struct {
			name string
			run  []term.RGB
		}{
			{"sky", tones[:2*hy]},
			{"sea", tones[2*(hy+1) : 2*seaEnd]},
		}
		for _, rg := range regions {
			seen := map[term.RGB]bool{}
			var prev term.RGB
			flips, hard, worst := 0, 0, 0.0
			for k, tone := range rg.run {
				if k > 0 && tone != prev {
					if seen[tone] {
						t.Errorf("%05.2f %s: tone #%02x%02x%02x comes BACK at half-row %d after another tone replaced it -- that is a rounding wobble, not a ramp",
							tod*24, rg.name, tone.R, tone.G, tone.B, k)
					}
					if isGrey(tone) != isGrey(prev) {
						flips++
					}
					if d := term.DeltaE(prev, tone); d >= 25 {
						hard++
						if d > worst {
							worst = d
						}
					}
				}
				seen[tone] = true
				prev = tone
			}
			if flips > 2 {
				t.Errorf("%05.2f %s: crosses between the grey ramp and the cube %d times -- a ramp enters the greys once and leaves once at most",
					tod*24, rg.name, flips)
			}
			// The floor, measured at his geometry with this palette, 2026-09-04:
			// no sky or sea carries more than TWO steps of 25 or over, and the
			// largest is 33 -- the first green step out of the daylight zenith
			// with blue pinned by the horizon (08:00, 14:30-15:30), which no
			// path through the cube gets under. Rounding row by row had up to
			// six per sky. If this trips, the palette moved or the walk did.
			if hard > 2 {
				t.Errorf("%05.2f %s: %d steps of 25 or over (worst %.0f) -- the floor is two", tod*24, rg.name, hard, worst)
			}
			if worst > 34 {
				t.Errorf("%05.2f %s: a step of %.0f -- the measured floor is 33", tod*24, rg.name, worst)
			}
		}
	}
}
