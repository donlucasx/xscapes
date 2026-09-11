// Command hist answers two questions the main check left open.
//
//  1. "Spread them more vertically" is his ask, and rows-occupied is a weak
//     way to read it: eight rows can still be one row plus a scatter. This
//     prints the ROW HISTOGRAM, and the share of the constellation sitting on
//     its single busiest row.
//
//  2. THE ONE THE TESTS MAY NOT ASK. The layout is cached under a key that
//     includes the moon's column and radius, and the moon SINKS as context
//     fills. If anything in that key moves during a session, the whole
//     constellation is relaid and every star moves at once -- which is the
//     locked guarantee failing in the only way a user would ever see it.
//     Swept here with the checklist held still and the context run 0 -> 1.
package main

import (
	"fmt"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

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

func warm(w, h int, seed int64, done int, tod, ctx float64) *canvas.Canvas {
	sh := scape.NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 300; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Working: true, Level: 0.5,
			TimeOfDay: tod, ContextUsed: ctx, TodoDone: done, TodoTotal: 32})
	}
	return c
}

func main() {
	fmt.Println("==== A. ROW HISTOGRAM -- is it spread, or a top row plus a scatter? ====")
	fmt.Println("  busiest row's share of the whole constellation; lower is more spread.")
	fmt.Println()
	type g struct{ w, h int }
	for _, gg := range []g{{40, 12}, {80, 24}, {125, 28}, {125, 62}} {
		for _, done := range []int{12, 26, 32} {
			for _, seed := range []int64{5, 7, 11} {
				c := warm(gg.w, gg.h, seed, done, 20.0/24, 0.30)
				at := lit(c)
				rows := map[int]int{}
				for _, p := range at {
					rows[p[1]]++
				}
				ks := make([]int, 0, len(rows))
				for k := range rows {
					ks = append(ks, k)
				}
				sort.Ints(ks)
				busiest, br := 0, -1
				var parts string
				for _, k := range ks {
					if rows[k] > busiest {
						busiest, br = rows[k], k
					}
					parts += fmt.Sprintf(" r%d:%d", k, rows[k])
				}
				share := 0.0
				if len(at) > 0 {
					share = float64(busiest) / float64(len(at)) * 100
				}
				flag := ""
				if share >= 33 {
					flag = "  <-- a THIRD or more on one row"
				}
				fmt.Printf("  %3dx%-3d seed %2d  %2d stars  busiest row %d holds %d (%.0f%%)%s\n     %s\n",
					gg.w, gg.h, seed, len(at), br, busiest, share, flag, parts)
			}
		}
		fmt.Println()
	}

	fmt.Println("==== B. DOES THE SKY HOLD STILL AS THE CONTEXT FILLS? ====")
	fmt.Println("  Checklist held at 26; context run 0.00 -> 1.00 in twenty-one steps,")
	fmt.Println("  which is what a real session does. The lit set must not change at all.")
	fmt.Println("  (The layout cache is keyed on the moon's column and radius, and the")
	fmt.Println("   moon moves with the context -- so this is where it would break.)")
	fmt.Println()
	for _, gg := range []g{{40, 12}, {80, 24}, {125, 28}, {125, 62}, {153, 51}} {
		for _, seed := range []int64{7, 23} {
			var base map[[2]int]bool
			changed, firstAt := 0, -1.0
			counts := map[int]bool{}
			for i := 0; i <= 20; i++ {
				ctx := float64(i) / 20
				at := lit(warm(gg.w, gg.h, seed, 26, 20.0/24, ctx))
				counts[len(at)] = true
				now := map[[2]int]bool{}
				for _, p := range at {
					now[p] = true
				}
				if base == nil {
					base = now
					continue
				}
				diff := 0
				for p := range base {
					if !now[p] {
						diff++
					}
				}
				for p := range now {
					if !base[p] {
						diff++
					}
				}
				if diff > 0 {
					changed += diff
					if firstAt < 0 {
						firstAt = ctx
					}
				}
			}
			cs := make([]int, 0, len(counts))
			for k := range counts {
				cs = append(cs, k)
			}
			sort.Ints(cs)
			verdict := "holds still"
			if changed > 0 {
				verdict = fmt.Sprintf("MOVED -- %d cell changes, first at context %.2f", changed, firstAt)
			}
			fmt.Printf("  %3dx%-3d seed %2d  star counts seen across the sweep %v  -> %s\n",
				gg.w, gg.h, seed, cs, verdict)
		}
	}
}
