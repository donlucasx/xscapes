package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// These tests exist because the defect they catch was invisible to every test
// in the repo and to every HTML preview, which render true RGB. It only shows
// up on the terminal Lucas actually uses.

func lumaOf(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// daylight is the window the scape is looked at in: 07:00 to 19:00.
func daylight(t float64) bool { return t >= 7.0/24 && t <= 19.0/24 }

// The sky and the sea must still be COLOURED on a 256-colour terminal.
//
// Measured on the palette this replaces: across 48 half-hours SeaFar landed on
// the greyscale ramp at 40 of them and SkyTop at 36, so for most of a working
// day the two biggest regions on screen had no colour at all on Terminal.app --
// while every HTML preview in the repo showed them blue, because those render
// true RGB. The picture Lucas saw and the picture the tests saw were different
// pictures.
//
// The rule is not "never grey". A BRIGHT grey is haze and a real sky has one;
// the night is monochrome on purpose and the cube gives no alternative. What
// must not happen is a DARK grey standing in for a colour during the hours the
// scape is looked at.
func TestSkyAndSeaKeepTheirColourThroughTheWorkingDay(t *testing.T) {
	type field struct {
		name string
		get  func(Palette) term.RGB
	}
	fields := []field{
		{"SkyTop", func(p Palette) term.RGB { return p.SkyTop }},
		{"SkyHorizon", func(p Palette) term.RGB { return p.SkyHorizon }},
		{"SeaFar", func(p Palette) term.RGB { return p.SeaFar }},
		{"SeaNear", func(p Palette) term.RGB { return p.SeaNear }},
	}
	for _, f := range fields {
		lost := 0
		var worst float64 = 999
		worstAt := 0.0
		for i := 0; i < 48; i++ {
			tod := float64(i) / 48
			if !daylight(tod) {
				continue
			}
			c := f.get(PaletteAt(tod))
			idx := c.Index256()
			shown := term.FromIndex256(idx)
			if idx >= 232 && lumaOf(shown) < 140 {
				lost++
				if lumaOf(shown) < worst {
					worst, worstAt = lumaOf(shown), tod
				}
				t.Logf("%s at %05.2f: asked for (%d,%d,%d), terminal shows grey %d",
					f.name, tod*24, c.R, c.G, c.B, shown.R)
			}
		}
		if lost > 0 {
			t.Errorf("%s falls to a DARK grey at %d of the 25 daylight half-hours (worst: grey %.0f at %05.2f) -- the colour is gone on the target terminal",
				f.name, lost, worst, worstAt*24)
		}
	}
}

// A gradient has to survive quantisation as a gradient.
//
// The sea's depth ramp used to collapse: at noon three different blues all
// landed on rgb(0,95,135) and eleven of the sea's fifteen rows were the same
// cell. Endpoints one cube step apart cannot do better than two bands, because
// the midpoint rounds back onto an end -- so the pair has to be at least two
// steps apart in some channel, and this is what checks it.
//
// It reads the rendered frame rather than recomputing the ramp, so a change to
// how the sea is painted is caught too, not just a change to the two colours.
func TestTheSeaShowsItsDepthOn256(t *testing.T) {
	for _, tod := range []float64{0.375, 0.5, 0.625} {
		c := canvas.New(120, 40, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(7, false)
		sh.MoonX = 0.28
		for i := 0; i < 12; i++ {
			sh.Update(c, 2+float64(i)/20, Activity{
				Working: true, Level: 0.6, TimeOfDay: tod, ContextUsed: 0.3})
		}
		hy := 0
		drop := 0.0
		var lum []float64
		var idx []int
		for y := 0; y < c.H; y++ {
			bg := c.BGAt(2, y)
			idx = append(idx, bg.Index256())
			lum = append(lum, lumaOf(term.FromIndex256(bg.Index256())))
		}
		for y := 1; y < c.H*3/4; y++ {
			if d := lum[y-1] - lum[y]; d > drop {
				drop, hy = d, y
			}
		}
		band := c.H
		for y := c.H - 1; y > hy; y-- {
			if idx[y] != idx[c.H-1] {
				band = y + 1
				break
			}
		}
		seen := map[int]bool{}
		run, worst := 0, 0
		prev := -1
		for y := hy; y < band; y++ {
			seen[idx[y]] = true
			if idx[y] == prev {
				run++
			} else {
				run = 1
			}
			if run > worst {
				worst = run
			}
			prev = idx[y]
		}
		rows := band - hy
		t.Logf("tod %.3f: sea is %d rows, %d distinct colours, longest flat run %d",
			tod, rows, len(seen), worst)
		if len(seen) < 4 {
			t.Errorf("tod %.3f: the sea shows only %d distinct colours over %d rows -- the depth ramp has collapsed",
				tod, len(seen), rows)
		}
		if worst > rows/2 {
			t.Errorf("tod %.3f: %d of the sea's %d rows are one colour -- that is a band, not a ramp",
				tod, worst, rows)
		}
	}
}
