// Command starflicker answers HIS report of 2026-09-10, made live inside the
// installed build: "I can see some of the constellation stars appearing and
// disappearing - once they appear, they should not disappear IMO."
//
// The count itself cannot be what he saw. stars() is turns + tasks/8 and both
// counters only ever increment; the one path that could drop it -- a todo list
// arriving and taking over from the fallback -- has fired ZERO times in all 122
// spools (grep '"kind":"todo"'), so it is not in play.
//
// That leaves the RENDER. todoStars skips any cell the disc reaches, and the
// disc's altitude is context:
//
//	discGeom: my = hy * (0.22 + 0.62*(1-lit)),  lit = 1 - ContextUsed
//
// so it starts at 0.22*hy -- INSIDE the constellation's band, which is rows
// 1..0.6*hy -- and sinks to 0.84*hy as the window fills. Context is not a
// stand-in either: 11,015 context events across 29 spools.
//
// This renders real frames and counts the '*' cells (todoStars owns that glyph;
// the ambient field is '.', '.', 0xb7, '+'), sweeping context the way a session
// does, and reports every star that goes OUT after it has been ON.
//
//	go run ./notes/s28-starflicker
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

type geom struct{ w, h int }

func main() {
	fmt.Printf("TIDE=%v\n\n", scape.Tide)
	geoms := []geom{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}}
	for _, g := range geoms {
		for _, done := range []int{8, 16, 32} {
			run(g, done, 11.0/24) // his clock this morning
		}
	}
	ambientSweep()
	resizeSweep()
	sessionSweep()
	timeSweep()
	fmt.Println("\nnight, in case the disc behaves differently as the moon:")
	run(geom{143, 27}, 16, 23.0/24)
}

func run(g geom, done int, tod float64) {
	// ONE shore, warmed up: a fresh one per frame has no history and Update
	// clamps a large gap to one nominal step. That trap voided two instruments
	// in session 27.
	sh := scape.NewShore(7, false)
	tm := 0.0
	warm := func(ctx float64) *canvas.Canvas {
		c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: ctx, TimeOfDay: tod,
			TodoDone: done, TodoTotal: 32,
		})
		return c
	}
	for k := 0; k < 60; k++ {
		warm(0)
	}

	// Which cells hold a lit constellation star, at each context step.
	const steps = 101
	seen := map[[2]int]bool{}   // has been ON at some point
	out := map[[2]int]float64{} // went OFF after being ON, at this context
	var counts []int
	for s := 0; s < steps; s++ {
		ctx := float64(s) / float64(steps-1)
		c := warm(ctx)
		on := map[[2]int]bool{}
		n := 0
		for y := 0; y < g.h; y++ {
			for x := 0; x < g.w; x++ {
				r, _, _ := c.ResolveAt(x, y, term.Profile256)
				if r == '*' {
					on[[2]int{x, y}] = true
					n++
				}
			}
		}
		counts = append(counts, n)
		for cell := range seen {
			if !on[cell] {
				if _, already := out[cell]; !already {
					out[cell] = ctx
				}
			}
		}
		for cell := range on {
			seen[cell] = true
		}
	}
	min, max := counts[0], counts[0]
	for _, n := range counts {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	fmt.Printf("%3dx%-3d done=%-3d tod=%.2f  lit '*' cells: min %d  max %d  cells that went OUT after being ON: %d\n",
		g.w, g.h, done, tod, min, max, len(out))
	if len(out) > 0 {
		fmt.Printf("            counts across the context sweep 0.00->1.00: %v\n", counts)
		for cell, ctx := range out {
			fmt.Printf("            star at col %d row %d goes out at context %.2f\n", cell[0], cell[1], ctx)
		}
	}
}
