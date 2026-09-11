package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// ambientSweep isolates the OTHER star field -- the ambient one, which twinkles
// by design. The twinkle is gated:
//
//	twinkle := (0.55 + 0.45*sin(...)) * StarVis
//	if twinkle <= 0.02 { continue }
//
// so the glyph is not dimmed below 0.02, it is NOT DRAWN. Above StarVis 0.2 the
// minimum of that range stays over the gate and the field only breathes; below
// it, each speck crosses the gate every cycle and blinks out completely.
//
// The question that matters is whether the blink is VISIBLE after the 256-cube
// quantisation at that alpha, so this counts only cells whose resolved
// foreground actually differs from their own background.
func ambientSweep() {
	fmt.Println("\n=== the ambient field, sky rows only, visible cells only ===")
	fmt.Printf("%-5s %8s %10s %10s %12s %10s\n",
		"hour", "StarVis", "sky cells", "visible", "blinks/frame", "gate-hits")
	w, h := 143, 27
	hy := int(float64(h) * 0.42)
	for _, hr := range []float64{0, 5, 6, 7, 8, 9, 10, 11, 12, 13, 17, 18, 19, 20} {
		tod := hr / 24
		sh := scape.NewShore(7, false)
		tm := 0.0
		step := func() *canvas.Canvas {
			c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: 0.35, TimeOfDay: tod,
				TodoDone: 16, TodoTotal: 32,
			})
			return c
		}
		for k := 0; k < 60; k++ {
			step()
		}
		var vis, all, blinks float64
		prev := map[[2]int]bool{}
		const frames = 250
		for f := 0; f < frames; f++ {
			c := step()
			cur := map[[2]int]bool{}
			for y := 0; y < hy; y++ {
				for x := 0; x < w; x++ {
					r, fg, bg := c.ResolveAt(x, y, term.Profile256)
					// ONLY the ambient field's own glyphs. The sky is painted
					// as split cells, so a naive "not a space" filter counts
					// several hundred half-block gradient cells an hour and
					// reports the same number at noon, when StarVis is 0 and
					// there are no ambient stars at all. That reading was void.
					if r != '.' && r != '\u00b7' && r != '+' {
						continue
					}
					all++
					if fg != bg {
						vis++
						cur[[2]int{x, y}] = true
					}
				}
			}
			if f > 0 {
				for cell := range prev {
					if !cur[cell] {
						blinks++
					}
				}
			}
			prev = cur
		}
		fmt.Printf("%-5.0f %8.3f %10.1f %10.1f %12.2f %10s\n",
			hr, scape.PaletteAt(tod).StarVis, all/frames, vis/frames, blinks/frames,
			gate(scape.PaletteAt(tod).StarVis))
	}
	fmt.Println("\ngate-hits: YES means the twinkle's own minimum (0.10*StarVis) falls under the")
	fmt.Println("0.02 cut-off, so every speck switches OFF once a cycle instead of dimming.")
}

func gate(sv float64) string {
	if sv > 0 && 0.10*sv <= 0.02 {
		return "YES"
	}
	return "no"
}

// resizeSweep is the one input the session sweep holds still and his terminal
// does not. Constellation positions are x = frac*(W-6)+3 and
// y = 1 + hash*(hy*3/5 - 1), so BOTH depend on the geometry: a width change
// slides every star sideways and a height change re-rolls its row.
func resizeSweep() {
	fmt.Println("\n=== what a resize does to the constellation ===")
	fmt.Printf("%-16s %10s %10s %10s\n", "change", "stars", "moved", "share")
	type g struct{ w, h int }
	pairs := [][2]g{
		{{143, 27}, {142, 27}}, // one column narrower
		{{143, 27}, {119, 27}}, // his own 107->119 grow, reversed
		{{143, 27}, {143, 26}}, // one row shorter
		{{143, 27}, {143, 21}}, // the band loses rows to the agent
		{{124, 22}, {124, 23}},
	}
	for _, p := range pairs {
		a, b := starCells(p[0].w, p[0].h), starCells(p[1].w, p[1].h)
		moved := 0
		for cell := range a {
			if !b[cell] {
				moved++
			}
		}
		fmt.Printf("%-16s %10d %10d %9.0f%%\n",
			fmt.Sprintf("%dx%d->%dx%d", p[0].w, p[0].h, p[1].w, p[1].h),
			len(a), moved, 100*float64(moved)/float64(max1(len(a))))
	}
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func starCells(w, h int) map[[2]int]bool {
	sh := scape.NewShore(7, false)
	tm := 0.0
	var c *canvas.Canvas
	for k := 0; k < 60; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.35, TimeOfDay: 11.0 / 24,
			TodoDone: 16, TodoTotal: 32,
		})
	}
	out := map[[2]int]bool{}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				out[[2]int{x, y}] = true
			}
		}
	}
	return out
}
