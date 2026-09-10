// Command searange asks one question of the RENDERED sea: how much of the
// activity level's range does the picture actually use?
//
// `xscapes tune` reports SATURATED as level > 0.95, which is a claim about the
// NUMBER. This measures the picture. Four channels, each averaged over eighty
// phases so a reading is amplitude rather than wherever the waves happen to be:
//
//   - foam cells and wave-glyph coverage, counted off the resolved glyphs
//
//   - the waterline's own swing and crest count, found by walking each column
//     down to the first WARM cell -- the sea is blue and the beach is warm, so
//     the boundary is exact and needs no colour matching. That channel is
//     painted into the BACKGROUND, which a glyph-reading instrument cannot see
//     at all; leaving it out is what made the first version of this answer a
//     lead rather than a verdict.
//
//     go run ./notes/searange
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	const w, h = 111, 25
	hy := h*42/100 + 2
	fmt.Printf("TIDE=%v\n", scape.Tide)
	fmt.Printf("%-6s %8s %10s %8s %8s %8s\n", "level", "foam", "wavecells", "swing", "crests", "meanrow")
	for _, lv := range []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 1.0} {
		var foam, wave, swing, crests, mean float64
		const n = 80
		// ONE shore, WARMED UP. A fresh one per frame has no history, so
		// anything that integrates -- the wave clock, and under Tide the
		// tide's own easing -- reads as if the session had just started. That
		// trap voided this instrument once and the drift instrument once, both
		// on 2026-09-09.
		sh := scape.NewShore(7, false)
		act := scape.Activity{Level: lv, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
		tm := 0.0
		for k := 0; k < 400; k++ {
			tm += 0.08
			sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm, act)
		}
		for k := 0; k < n; k++ {
			c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, act)
			prof := make([]int, w)
			for x := 0; x < w; x++ {
				prof[x] = h
				for y := hy; y < h; y++ {
					r, _, bg := c.ResolveAt(x, y, term.Profile256)
					switch r {
					case '\u224b':
						foam++
						wave++
					case '~', '-':
						wave++
					}
					if prof[x] == h && bg.R > bg.B {
						prof[x] = y
					}
				}
			}
			lo, hi := h, 0
			for _, v := range prof {
				if v < lo {
					lo = v
				}
				if v > hi {
					hi = v
				}
			}
			swing += float64(hi - lo)
			sum := 0
			for _, v := range prof {
				sum += v
			}
			mean += float64(sum) / float64(w)
			pk := 0
			for x := 2; x < w-2; x++ {
				if prof[x] < prof[x-2] && prof[x] <= prof[x+2] {
					pk++
				}
			}
			crests += float64(pk)
		}
		fmt.Printf("%-6.2f %8.0f %10.0f %8.2f %8.2f %8.2f\n", lv, foam/n, wave/n, swing/n, crests/n, mean/n)
	}
	fmt.Println()
	fmt.Println("His real level while working, folded from 341h of recordings:")
	fmt.Println("  0.3-0.4 22.9%  0.4-0.5 18.0%  0.5-0.6 16.9%  0.6-0.7 14.0%")
	fmt.Println("  0.7-0.8  9.0%  0.8-0.9  6.4%  0.9-1.0 12.8%   -- above 0.70 is 28.2%")
}
