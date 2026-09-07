package scape

import (
	"fmt"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// His report of 2026-09-06, from a live 143x62 window: one yellow pixel on the
// sun's face. It is an ambient dust speck. stars() plots into the FAR layer and
// moon() paints only BACKGROUNDS, so a glyph under the disc survives and is
// composited against the body instead of the sky -- at his geometry, dead in
// the centre column one row above the middle.
//
// todoStars() has carried a guard against exactly this since it was written
// ("Never on the moon: it carries context remaining, and a star inside the
// disc would be read as part of it"); the ambient field never had one.
//
// The disc's painted set is found the way the crescent test finds it: against a
// MoonEdge="none" control frame, which paints everything BUT the disc. The
// sweep is deliberately wider than his window -- a guard fitted to one
// screenshot passes on that screenshot and still leaks a row taller, because
// moon() paints a cell when EITHER of its two half-rows is inside the disc.
func TestNoStarIsDrawnOnTheDisc(t *testing.T) {
	ambient := map[rune]bool{'.': true, '·': true, '+': true, '*': true}

	// Geometries either side of his 143x62 (scape 27), plus the design target
	// and the extremes the scape is built for. Several seeds, because the star
	// field is hashed: one seed proves nothing.
	widths := []int{80, 120, 143, 200}
	heights := []int{12, 22, 26, 27, 28, 30}
	seeds := []int64{5, 7, 11, 23, 42}
	// Context sweeps the disc's altitude and its phase; time of day sweeps
	// night (where the field is brightest) through midday.
	contexts := []float64{0.02, 0.35, 0.6, 0.9}
	tods := []float64{0.05, 0.5, 0.865}

	bad := 0
	for _, w := range widths {
		for _, h := range heights {
			for _, seed := range seeds {
				for _, ctx := range contexts {
					for _, tod := range tods {
						render := func(edge string) *canvas.Canvas {
							c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
							sh := NewShore(seed, false)
							sh.MoonX = 0.28
							sh.MoonEdge = edge
							// A full checklist too: the '*' field has its own
							// guard and it must be held to the same shape, or
							// the two drift apart again.
							sh.Update(c, 2, Activity{ContextUsed: ctx, TimeOfDay: tod, TodoDone: 9, TodoTotal: 12})
							return c
						}
						frame, control := render(""), render("none")
						for y := 0; y < h; y++ {
							for x := 0; x < w; x++ {
								ch, fg, bg := frame.ResolveAt(x, y, term.Profile256)
								ch0, fg0, bg0 := control.ResolveAt(x, y, term.Profile256)
								if ch == ch0 && fg == fg0 && bg == bg0 {
									continue // the disc does not paint here
								}
								if !ambient[ch] {
									continue
								}
								bad++
								if bad <= 8 {
									t.Errorf("%dx%d seed %d ctx %.2f tod %.3f: %q on the disc at (%d,%d) -- sky there is %q",
										w, h, seed, ctx, tod, string(ch), x, y, string(ch0))
								}
							}
						}
					}
				}
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d cells across the sweep carry a star glyph on the disc", bad)
	}
}

// The guard must follow moon()'s own half-row sampling. Testing whole rows
// instead (hypot(dx/2, dy) < rr) misses the ring of cells the disc reaches with
// only one of its two half-rows -- none at his exact 143x27, which is the trap:
// a fix measured against his screenshot alone would look complete.
func TestTheDiscGuardFollowsTheHalfRowSampling(t *testing.T) {
	const w = 143
	for _, h := range []int{22, 26, 27, 28, 30} {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(5, false)
		sh.MoonX = 0.28
		sh.Update(c, 2, Activity{ContextUsed: 0.02, TimeOfDay: 0.865})
		control := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sc := NewShore(5, false)
		sc.MoonX = 0.28
		sc.MoonEdge = "none"
		sc.Update(control, 2, Activity{ContextUsed: 0.02, TimeOfDay: 0.865})

		var painted, covered, missed int
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
				ch0, fg0, bg0 := control.ResolveAt(x, y, term.Profile256)
				if ch == ch0 && fg == fg0 && bg == bg0 {
					continue
				}
				painted++
				if sh.DiscCovers(x, y) {
					covered++
				} else {
					missed++
				}
			}
		}
		if painted == 0 {
			t.Fatalf("%dx%d: the control found no disc at all", w, h)
		}
		if missed != 0 {
			t.Errorf("%dx%d: the disc paints %d cells DiscCovers misses (of %d painted)", w, h, missed, painted)
		}
		t.Log(fmt.Sprintf("%dx%d: %d painted, %d covered", w, h, painted, covered))
	}
}
