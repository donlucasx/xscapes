// Part 3: does an 11-cell-tall companion fit anywhere, at any tide level?
// Measured against the real Shore, not read off their table.
//
// Arithmetic, stated once so it can be checked:
//
//	SandTop is "the first row that is dry sand in every column", so dry rows
//	are SandTop..H-1 inclusive = H - SandTop rows.
//	The companion is drawn at top = H-2-chh and its feet land on H-3, leaving
//	the two rows H-2 and H-1 spare.
//	Fully dry  <=>  H-2-chh >= SandTop  <=>  H - SandTop >= chh + 2.
//
// notes/s28-closer and notes/s28-askroom print (H-1-SandTop), one FEWER than
// the dry-row count, so their tables and mine differ by exactly one. Both are
// printed below so nobody has to guess which convention a number is in.
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

func sandTopAt(w, h int, level float64) int {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	return sh.SandTop()
}

func main() {
	geoms := []struct {
		name string
		w, h int
	}{{"80x24", 80, 24}, {"111x25", 111, 25}, {"124x22", 124, 22},
		{"143x27", 143, 27}, {"153x51", 153, 51}}
	levels := []float64{0.00, 0.10, 0.20, 0.30, 0.40, 0.50, 0.54, 0.59, 0.65, 0.83, 1.00}

	fmt.Println("DRY SAND ROWS (H - SandTop) at each tide level:")
	fmt.Printf("%-7s", "level")
	for _, g := range geoms {
		fmt.Printf("%9s", g.name)
	}
	fmt.Println()
	for _, lv := range levels {
		fmt.Printf("%-7.2f", lv)
		for _, g := range geoms {
			fmt.Printf("%9d", g.h-sandTopAt(g.w, g.h, lv))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("A companion chh cells tall is FULLY DRY iff dry rows >= chh+2.")
	fmt.Println("  shipped chh=7  needs  9 dry rows")
	fmt.Println("  theirs  chh=11 needs 13 dry rows")
	fmt.Println()

	for _, chh := range []int{7, 9, 10, 11} {
		fmt.Printf("== chh = %2d (needs %2d dry rows) ==\n", chh, chh+2)
		for _, g := range geoms {
			best := g.h - sandTopAt(g.w, g.h, 0.0) // water fully out = the ceiling
			atAsk := g.h - sandTopAt(g.w, g.h, 0.59)
			fmt.Printf("  %-8s ceiling(level 0.00) %2d dry | at median ask 0.59 %2d dry | ask:%s  best-case:%s\n",
				g.name, best, atAsk, ok(atAsk, chh+2), ok(best, chh+2))
		}
		fmt.Println()
	}
}

func ok(have, need int) string {
	if have >= need {
		return fmt.Sprintf(" FITS (+%d)", have-need)
	}
	return fmt.Sprintf(" WET by %d", need-have)
}
