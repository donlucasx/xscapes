// Does the constellation depend on how long the shore has been running?
// If it does not, a wide seed sweep can warm for far fewer frames -- and if it
// does, every quick instrument in this study is void.
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func at(w, h int, seed int64, done, frames int) [][2]int {
	sh := scape.NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < frames; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Working: true, Level: 0.5,
			TimeOfDay: 20.0 / 24, ContextUsed: 0.3, TodoDone: done, TodoTotal: 32})
	}
	var out [][2]int
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				out = append(out, [2]int{x, y})
			}
		}
	}
	return out
}

func same(a, b [][2]int) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[[2]int]bool{}
	for _, p := range a {
		m[p] = true
	}
	for _, p := range b {
		if !m[p] {
			return false
		}
	}
	return true
}

func main() {
	type g struct{ w, h int }
	bad := 0
	for _, gg := range []g{{40, 12}, {80, 24}, {125, 28}, {125, 62}, {153, 51}} {
		for _, seed := range []int64{3, 7, 10, 19} {
			for _, done := range []int{1, 12, 26, 32} {
				ref := at(gg.w, gg.h, seed, done, 240)
				for _, f := range []int{8, 20, 60} {
					if !same(ref, at(gg.w, gg.h, seed, done, f)) {
						bad++
						fmt.Printf("  DIFFERS %dx%d seed %d done %d: %d frames vs 240\n", gg.w, gg.h, seed, done, f)
					}
				}
			}
		}
	}
	if bad == 0 {
		fmt.Println("The constellation is IDENTICAL at 8, 20, 60 and 240 warm frames,")
		fmt.Println("across 5 geometries x 4 seeds x 4 counts = 80 configurations.")
		fmt.Println("=> the layout does not depend on scene history, so a wide sweep may warm briefly.")
	} else {
		fmt.Printf("%d configurations depend on warm length -- quick instruments are VOID.\n", bad)
	}
}
