package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func dist(a, b term.RGB) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return dr*dr + dg*dg + db*db
}

// discProbe finds every cell still drawn as a half block and reports how far
// apart its two halves are. A half block leaves a one-pixel rule at the cell's
// bottom edge whichever way up it is drawn -- but the rule is painted in the
// cell's BACKGROUND, which is the other half's colour. So when the two halves
// are nearly the same colour the split buys nothing and the rule is the only
// thing it produces.
func discProbe(seed int64) {
	term.LowerHalf = term.DetectSplit("Apple_Terminal")
	term.NoSplitCells = term.DetectNoSplit("Apple_Terminal")
	w := 126
	type bucket struct{ n, edge, interior int }
	buckets := map[string]*bucket{}
	var samples []string
	totalSplit, interior := 0, 0
	for _, h := range []int{20, 24, 28, 30, 34, 40, 47} {
		for i := 0; i < 48; i++ {
			tod := float64(i) / 48
			c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh := scape.NewShore(seed, false)
			sh.MoonX = 0.28
			act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3}
			for k := 0; k < 12; k++ {
				sh.Update(c, 3+float64(k)/40, act)
			}
			for y := 0; y < sh.SandTop(); y++ {
				for x := 0; x < w; x++ {
					r, fg, bg := c.ResolveAt(x, y, term.Profile256)
					if r != '▄' && r != '▀' {
						continue
					}
					if !sh.DiscCovers(x, y) {
						continue // only the sun and the moon
					}
					// An INTERIOR cell is one whose neighbours above and below
					// are also disc; an edge cell is not. A rule across the
					// face comes from the interior ones.
					if sh.DiscCovers(x, y-1) && sh.DiscCovers(x, y+1) {
						interior++
					}
					totalSplit++
					d := dist(fg, bg)
					key := "far  (>2000)"
					switch {
					case d == 0:
						key = "IDENTICAL"
					case d < 300:
						key = "near (<300)"
					case d < 2000:
						key = "mid  (<2000)"
					}
					if buckets[key] == nil {
						buckets[key] = &bucket{}
					}
					buckets[key].n++
					if len(samples) < 8 && d > 0 && d < 2000 {
						samples = append(samples, fmt.Sprintf(
							"  h=%d tod=%.2f (%2d,%2d)  up %3d,%3d,%3d  down %3d,%3d,%3d  dist %.0f",
							h, tod, x, y, bg.R, bg.G, bg.B, fg.R, fg.G, fg.B, d))
					}
				}
			}
		}
	}
	fmt.Printf("Half-block cells ON THE DISC across 7 heights x 48 half-hours: %d\n", totalSplit)
	fmt.Printf("  of those, INTERIOR (a rule across the face): %d\n\n", interior)
	for _, k := range []string{"IDENTICAL", "near (<300)", "mid  (<2000)", "far  (>2000)"} {
		if b := buckets[k]; b != nil {
			fmt.Printf("  %-14s %6d  (%.1f%%)\n", k, b.n, float64(b.n)/float64(totalSplit)*100)
		}
	}
	fmt.Println("\nsamples of the low-contrast splits -- the ones that buy nothing:")
	for _, s := range samples {
		fmt.Println(s)
	}
}
