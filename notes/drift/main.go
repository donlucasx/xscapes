// Command drift answers one question: which way does the sea actually travel?
//
// His report on the tide build, 2026-09-09: "it still moves to the sides."
//
// Cross-correlates one frame against a later one, shifted in x and in y, and
// prints how well each shift matches. A pattern that travels leaves a clear
// peak away from zero on the axis it travels along; one that does not leaves a
// flat profile.
//
// ⚠ TWO TRAPS, both of which voided the first version of this.
//
// A FRESH Shore per frame measures nothing. Shore.Update clamps any gap over a
// second to one nominal step, so two frames built from two new Shores are
// IDENTICAL and every shift scores the same -- which reads exactly like "no
// travel". One shore, advanced in frame-sized steps.
//
// And `go test` CACHES, so the same command with a different environment
// variable can hand back the previous run's output verbatim. This is a main
// package for that reason.
//
// ⚠ The Y axis is CONFOUNDED under Tide and no direction should be read off
// it: the wash moves the waterline, the waterline sets `depth`, and `depth`
// sets the threshold that decides which cells carry a mark at all -- so the
// whole band of marks rides up and down with the wash regardless of which way
// the wave itself is going. Flipping the wave's own sign moved the y peak from
// -1 to -2, which is noise. X is the axis his report is about and X is clean.
//
//	go run ./notes/drift              (the tide, now the default)
//	XSCAPES_TIDE=0 go run ./notes/drift   (the old fixed waterline)
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

var lastT float64

func main() {
	const w, h = 120, 26
	lastT = 0
	lastT = 0
	// ONE shore, advanced in real frame-sized steps. A fresh one clamps any
	// gap over a second to a single nominal step, so two frames built from two
	// new Shores are IDENTICAL and every shift scores the same -- which is what
	// the first version of this measured, and it reported no travel at all.
	sh := scape.NewShore(7, false)
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := scape.Activity{Level: 0.6, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	step := func(n int) {
		for i := 0; i < n; i++ {
			c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh.Update(c, 0.08*float64(i)+lastT, act)
		}
	}
	_ = step
	frame := func(nsteps int) [][]bool {
		for i := 0; i < nsteps; i++ {
			c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			lastT += 0.08
			sh.Update(c, lastT, act)
		}
		g := make([][]bool, h)
		for y := range g {
			g[y] = make([]bool, w)
			for x := 0; x < w; x++ {
				r, _, _ := c.ResolveAt(x, y, term.Profile256)
				g[y][x] = r == '≈' || r == '~' || r == '≋'
			}
		}
		return g
	}
	a, b := frame(40), frame(18)
	score := func(dx, dy int) int {
		n := 0
		for y := 2; y < h-2; y++ {
			for x := 2; x < w-2; x++ {
				yy, xx := y+dy, x+dx
				if yy < 0 || yy >= h || xx < 0 || xx >= w {
					continue
				}
				if a[y][x] && b[yy][xx] {
					n++
				}
			}
		}
		return n
	}
	bestX, bestXs := 0, -1
	for dx := -4; dx <= 4; dx++ {
		if s := score(dx, 0); s > bestXs {
			bestXs, bestX = s, dx
		}
	}
	bestY, bestYs := 0, -1
	for dy := -4; dy <= 4; dy++ {
		if s := score(0, dy); s > bestYs {
			bestYs, bestY = s, dy
		}
	}
	fmt.Printf("TIDE=%v   best x-shift %+d   best y-shift %+d\n", scape.Tide, bestX, bestY)
	fmt.Print("  x: ")
	for dx := -4; dx <= 4; dx++ {
		fmt.Printf("%+d:%d  ", dx, score(dx, 0))
	}
	fmt.Print("\n  y: ")
	for dy := -4; dy <= 4; dy++ {
		fmt.Printf("%+d:%d  ", dy, score(0, dy))
	}
	fmt.Println()
	fmt.Println("  (+y = toward the shore, +x = to the right, 0 = no travel on that axis)")
}
