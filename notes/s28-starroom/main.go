// Command starroom measures the sky the constellation is allowed to use, and
// what it is actually using, before anything is changed.
//
// HIS NOTE, 2026-09-10: "can we have the stars appear randomly across the sky,
// instead of from left to right? Also can we spread them more vertically?
// currently sitting in a narrow band pretty high up".
//
//	go run ./notes/s28-starroom
package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	fmt.Println("THE BAND: top=1, bot=hy*3/4 since 2026-09-10 (was 3/5), where hy is the horizon row.")
	fmt.Printf("%-10s %5s %8s %10s %12s %s\n", "geom", "hy", "band", "rows used", "sky rows", "share of the sky")
	for _, g := range []struct{ w, h int }{{80, 24}, {111, 25}, {125, 28}, {143, 27}, {153, 51}, {125, 62}} {
		hy := int(float64(g.h) * 0.42)
		bot := hy * 3 / 4
		if bot <= 1 {
			bot = 2
		}
		fmt.Printf("%-10s %5d %8s %10d %12d %.0f%%\n",
			fmt.Sprintf("%dx%d", g.w, g.h), hy, fmt.Sprintf("1..%d", bot), bot, hy,
			100*float64(bot)/float64(hy))
	}

	fmt.Println("\nHOW EVEN IS THE SPACING? A golden-ratio sequence is LOW-DISCREPANCY --")
	fmt.Println("it is deliberately as even as possible, which is why it reads as a ruler")
	fmt.Println("rather than as a sky. Gap between neighbouring stars, sorted:")
	const w, h = 125, 28
	for _, done := range []int{12, 19, 26} {
		cols := starCols(w, h, done)
		var gaps []int
		for i := 1; i < len(cols); i++ {
			gaps = append(gaps, cols[i]-cols[i-1])
		}
		mean, sd := stats(gaps)
		fmt.Printf("  %2d stars: gaps %v\n            mean %.1f  sd %.2f  (sd/mean %.2f -- a real random sky is ~1.00)\n",
			done, gaps, mean, sd, sd/mean)
	}

	fmt.Println("\nWHAT A POISSON SKY WOULD GIVE, for comparison -- same count, hashed")
	fmt.Println("positions with no spreading term at all:")
	for _, done := range []int{12, 19, 26} {
		var cols []int
		for i := 0; i < done; i++ {
			cols = append(cols, 3+int(scape.HashF(i, 91, 7)*float64(w-6)))
		}
		sortInts(cols)
		var gaps []int
		for i := 1; i < len(cols); i++ {
			gaps = append(gaps, cols[i]-cols[i-1])
		}
		mean, sd := stats(gaps)
		dup := 0
		for i := 1; i < len(cols); i++ {
			if cols[i]-cols[i-1] < 2 {
				dup++
			}
		}
		fmt.Printf("  %2d stars: mean %.1f  sd %.2f  (sd/mean %.2f)  pairs closer than 2 cells: %d\n",
			done, mean, sd, sd/mean, dup)
	}
	fmt.Println("\n  ⚠ That is the trade: a pure hash looks like a sky and COLLIDES. The")
	fmt.Println("  count is the channel, and two stars in adjacent cells read as one.")
}

func starCols(w, h, done int) []int {
	sh := scape.NewShore(7, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 120; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: 20.0 / 24,
			TodoDone: done, TodoTotal: 32,
		})
	}
	var cols []int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				cols = append(cols, x)
			}
		}
	}
	sortInts(cols)
	return cols
}

func stats(v []int) (mean, sd float64) {
	if len(v) == 0 {
		return 0, 0
	}
	for _, x := range v {
		mean += float64(x)
	}
	mean /= float64(len(v))
	for _, x := range v {
		sd += (float64(x) - mean) * (float64(x) - mean)
	}
	return mean, math.Sqrt(sd / float64(len(v)))
}

func sortInts(v []int) {
	for i := range v {
		for j := i + 1; j < len(v); j++ {
			if v[j] < v[i] {
				v[i], v[j] = v[j], v[i]
			}
		}
	}
}
