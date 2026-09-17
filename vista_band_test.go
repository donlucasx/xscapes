package main

import (
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestTheVistaTailReadsOnTheBand: his note of 2026-09-16 on the hero
// variants, measured as drawn. At four hours and two geometries, (1) the
// band's first row is the meadow's last colour, no seam (within 8 luma,
// truecolor); (2) by day the band's last row is darker than its first, and
// on the cube it is still green rather than a grey; (3) every glyph of a
// four-line tail, oldest to newest, sits at least 40 luma from its own
// ground, in truecolor and on the cube.
func TestTheVistaTailReadsOnTheBand(t *testing.T) {
	tail := []reduce.Line{
		{Text: "read  internal/auth/handler.go  142 lines", Age: 1},
		{Text: "search  rate.Limiter  3 files", Age: 0.66},
		{Text: "edit  internal/auth/handler.go  +18 -2", Age: 0.33},
		{Text: "shell  go test ./internal/auth  exit 1", Age: 0, Bad: true},
	}
	lum := func(c term.RGB) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }
	for _, g := range [][2]int{{124, 30}, {80, 24}} {
		for _, tod := range []float64{0.0, 0.25, 0.5, 0.75} {
			c := canvas.New(g[0], g[1], canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			lay := compose(c.W, 12, true)
			v := scenes.NewVista(7, false)
			v.OwlX = lay.CatX
			act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.3, TimeOfDay: tod}
			v.Update(c, 3.0, act)
			st := reduce.State{Act: act, Tail: tail}
			drawVistaWith(c, v, lay, st, 3.0, nil)
			_, _, _, bandTop := v.Layout()
			// (1) continuity, at a column clear of the pines (left), the fire
			// (three eighths across) and the owl (right).
			x := c.W/2 + 8
			_, _, meadowLast := c.ResolveAt(x, bandTop-1, term.ProfileTrueColor)
			_, _, bandFirst := c.ResolveAt(x, bandTop, term.ProfileTrueColor)
			if d := math.Abs(lum(meadowLast) - lum(bandFirst)); d > 8 {
				t.Errorf("%dx%d tod %.2f: the band opens %.1f luma from the meadow's last row (%v vs %v)", g[0], g[1], tod, d, meadowLast, bandFirst)
			}
			_, _, bandLast := c.ResolveAt(x, c.H-1, term.ProfileTrueColor)
			if tod == 0.5 {
				// (2) deeper toward the bottom, and green on the cube.
				if lum(bandLast) >= lum(bandFirst) {
					t.Errorf("%dx%d noon: the band does not deepen (%v top, %v bottom)", g[0], g[1], bandFirst, bandLast)
				}
				for y := bandTop; y < c.H; y++ {
					_, _, q := c.ResolveAt(x, y, term.Profile256)
					if int(q.G) < int(q.R)+15 {
						t.Errorf("%dx%d noon row %d: the band is %v on the cube, a grey", g[0], g[1], y, q)
					}
				}
			}
			// (3) every glyph of the tail against its own ground.
			for _, prof := range []term.Profile{term.ProfileTrueColor, term.Profile256} {
				worst, wx, wy := 1e9, 0, 0
				n := 0
				for y := bandTop; y < c.H; y++ {
					for x := 0; x < c.W; x++ {
						ch, fg, bg := c.ResolveAt(x, y, prof)
						if ch == ' ' {
							continue
						}
						n++
						if d := math.Abs(lum(fg) - lum(bg)); d < worst {
							worst, wx, wy = d, x, y
						}
					}
				}
				if n < 60 { // three rows of band at 80x24 hold three of the four lines
					t.Fatalf("%dx%d tod %.2f: only %d glyphs in the band; the tail was not written", g[0], g[1], tod, n)
				}
				if worst < 40 {
					ch, fg, bg := c.ResolveAt(wx, wy, prof)
					t.Errorf("%dx%d tod %.2f %v: glyph %q at %d,%d is %.1f luma from its ground (%v on %v), floor 40", g[0], g[1], tod, prof, ch, wx, wy, worst, fg, bg)
				}
			}
		}
	}
}
