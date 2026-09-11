// Command seeds asks one question their suite cannot: does the constellation
// hold at a seed nobody chose?
//
// TestNoTwoStarsCollide sweeps five seeds -- 5, 7, 11, 23, 42 -- and passes.
// But the seed is not a test fixture: CLAUDE.md says it is "a per-repo seed
// (keyed by git root) so the same repo always gets the same landscape shape",
// so the seed a user gets is effectively arbitrary. A guarantee that holds at
// five seeds and fails at the sixth is not a guarantee, it is a sample.
//
// The bar is the suite's own: 2.0 SCREEN units, a row counting double.
package main

import (
	"fmt"
	"math"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

const bar = 2.0

func lit(c *canvas.Canvas) [][2]int {
	var at [][2]int
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				at = append(at, [2]int{x, y})
			}
		}
	}
	return at
}

func warm(w, h int, seed int64, done int) *canvas.Canvas {
	sh := scape.NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 12; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Working: true, Level: 0.5,
			TimeOfDay: 20.0 / 24, ContextUsed: 0.3, TodoDone: done, TodoTotal: 32})
	}
	return c
}

func minGap(at [][2]int) (float64, [2]int, [2]int) {
	best := math.MaxFloat64
	var p, q [2]int
	for i := range at {
		for j := i + 1; j < len(at); j++ {
			dx := float64(at[i][0] - at[j][0])
			dy := float64(at[i][1]-at[j][1]) * 2.0
			if d := math.Hypot(dx, dy); d < best {
				best, p, q = d, at[i], at[j]
			}
		}
	}
	return best, p, q
}

func main() {
	fmt.Println("SEEDS -- the collision guarantee at seeds nobody picked.")
	fmt.Println("Bar is the suite's own: 2.0 screen units. 200 seeds per geometry,")
	fmt.Printf("counts 1..32, one shore warmed 12 frames each (proved warm-independent).\n\n")

	type g struct{ w, h int }
	geos := []g{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {125, 28}, {143, 27}, {153, 51}, {125, 62}}
	suite := map[int64]bool{5: true, 7: true, 11: true, 23: true, 42: true}

	fmt.Println("  geometry   seeds tested  seeds with a collision   worst gap   worst case")
	for _, gg := range geos {
		bad := map[int64]bool{}
		worst, worstAt := math.MaxFloat64, ""
		var badList []int64
		for seed := int64(1); seed <= 200; seed++ {
			for done := 1; done <= 32; done++ {
				at := lit(warm(gg.w, gg.h, seed, done))
				if len(at) < 2 {
					continue
				}
				d, p, q := minGap(at)
				if d < worst {
					worst = d
					worstAt = fmt.Sprintf("seed %d done %d: %v %v", seed, done, p, q)
				}
				if d < bar && !bad[seed] {
					bad[seed] = true
					badList = append(badList, seed)
				}
			}
		}
		sort.Slice(badList, func(i, j int) bool { return badList[i] < badList[j] })
		show := badList
		if len(show) > 12 {
			show = show[:12]
		}
		inSuite := 0
		for _, s := range badList {
			if suite[s] {
				inSuite++
			}
		}
		fmt.Printf("  %3dx%-3d   %12d  %22d   %9.2f   %s\n",
			gg.w, gg.h, 200, len(badList), worst, worstAt)
		if len(badList) > 0 {
			fmt.Printf("            failing seeds (first few): %v%s\n", show,
				map[bool]string{true: "", false: " ..."}[len(badList) <= 12])
			fmt.Printf("            of those, in the suite's own five: %d\n", inSuite)
		}
	}
}
