package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestAStarNeverGoesOut is HIS report of 2026-09-10, made live in the installed
// build: "I can see some of the constellation stars appearing and dissapearing-
// once they appear, they should not dissapear IMO."
//
// It renders the COMPOSED frame -- drawScene, not just Shore.Update -- because
// the scape is not the only thing that writes into the sky: drawReadout puts
// the context number under the disc from 40% used, and the disc sinks as the
// window fills, so the number sweeps down through the constellation's band.
// An instrument that stops at Shore.Update cannot see that and reports clean.
func TestAStarNeverGoesOut(t *testing.T) {
	for _, name := range []string{"cat", "crab"} {
		for _, g := range []struct{ w, h int }{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}} {
			f := newFrames(g.w, g.h, 7, false, true, 0, 11.0/24)
			f.cat = companion.New(name)
			f.cat.FaceLeft(true)
			f.ccw, f.chh = f.cat.Size()
			f.lay = compose(g.w, f.ccw, true)
			f.sh.MoonX = f.lay.MoonX
			f.profile = term.Profile256

			lit := map[[2]int]bool{}
			blink := map[[2]int]int{}
			at := map[[2]int]float64{}
			const steps = 200
			tm := 0.0
			for s := 0; s < steps; s++ {
				ctx := float64(s) / float64(steps-1)
				done := 1 + s*24/steps
				st := reduce.State{
					Act: scape.Activity{
						Level: 0.5, Working: true, ContextUsed: ctx,
						TimeOfDay: 11.0 / 24, TodoDone: done, TodoTotal: 32,
					},
					Pose: companion.Working,
					Tail: nil,
				}
				tm += 0.08
				f.c = canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				f.sh.Update(f.c, tm, st.Act)
				top := f.c.H - 2 - f.chh
				drawScene(f.c, f.sh, f.cat, f.lay, st, tm, f.seed, top)

				on := map[[2]int]bool{}
				for y := 0; y < g.h; y++ {
					for x := 0; x < g.w; x++ {
						if r, _, _ := f.c.ResolveAt(x, y, term.Profile256); r == '*' {
							on[[2]int{x, y}] = true
						}
					}
				}
				for cell := range lit {
					if !on[cell] {
						blink[cell]++
						if _, ok := at[cell]; !ok {
							at[cell] = ctx
						}
						delete(lit, cell)
					}
				}
				for cell := range on {
					lit[cell] = true
				}
			}
			if len(blink) > 0 {
				for cell, n := range blink {
					t.Logf("%s %dx%d: star col %d row %d went out %dx, first at context %.2f",
						name, g.w, g.h, cell[0], cell[1], n, at[cell])
				}
				t.Errorf("%s %dx%d: %d star slots go out after lighting", name, g.w, g.h, len(blink))
			} else {
				t.Logf("%s %dx%d: steady", name, g.w, g.h)
			}
		}
	}
	_ = time.Now
}
