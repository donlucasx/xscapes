package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// sessionSweep is the honest simulation: a real session raises BOTH the star
// count (turns closed and subagents finished, monotonic) and the context (also
// monotonic -- 11,015 samples in his log step up 914 times and down 6).
//
// The disc's altitude is context: my = hy*(0.22 + 0.62*ctx). The constellation
// occupies rows 1..hy*3/5. So the disc STARTS inside the constellation's band
// and sinks out of it through the first half of the window, sweeping a column
// of sky about ten cells wide as it goes -- and todoStars skips any cell the
// disc reaches.
//
// This counts, per star slot, how many times it goes lit -> unlit.
func sessionSweep() {
	fmt.Println("\n=== a real session: stars accumulate, context fills, disc sinks through the band ===")
	fmt.Printf("%-9s %6s %6s %8s %8s %10s  %s\n",
		"geom", "hy", "band", "discTop", "discBot", "blinkouts", "which slots, and over what context")
	for _, g := range []geom{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}, {143, 62}} {
		sh := scape.NewShore(7, false)
		tm := 0.0
		step := func(done int, ctx float64) *canvas.Canvas {
			c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: ctx, TimeOfDay: 11.0 / 24,
				TodoDone: done, TodoTotal: 32,
			})
			return c
		}
		for k := 0; k < 60; k++ {
			step(0, 0)
		}
		const steps = 200
		lit := map[[2]int]bool{}
		blink := map[[2]int]int{}
		spanLo := map[[2]int]float64{}
		spanHi := map[[2]int]float64{}
		for s := 0; s < steps; s++ {
			ctx := float64(s) / float64(steps-1) // 0 -> 1 across the window
			done := 1 + s*24/steps               // 1 -> 24 stars earned
			c := step(done, ctx)
			on := map[[2]int]bool{}
			for y := 0; y < g.h; y++ {
				for x := 0; x < g.w; x++ {
					if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
						on[[2]int{x, y}] = true
					}
				}
			}
			for cell := range lit {
				if !on[cell] {
					blink[cell]++
					if _, ok := spanLo[cell]; !ok {
						spanLo[cell] = ctx
					}
					spanHi[cell] = ctx
					delete(lit, cell) // count one transition, not one per frame
				}
			}
			for cell := range on {
				lit[cell] = true
			}
		}
		hy := int(float64(g.h) * 0.42)
		band := hy * 3 / 5
		mx, _ := sh.MoonPos()
		_ = mx
		detail := ""
		for cell, n := range blink {
			detail += fmt.Sprintf("col%d/row%d x%d (ctx %.2f-%.2f) ", cell[0], cell[1], n, spanLo[cell], spanHi[cell])
		}
		if detail == "" {
			detail = "none"
		}
		fmt.Printf("%-9s %6d %6s %8.1f %8.1f %10d  %s\n",
			fmt.Sprintf("%dx%d", g.w, g.h), hy, fmt.Sprintf("1-%d", band),
			float64(hy)*0.22, float64(hy)*0.84, len(blink), detail)
	}
}
