package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

// scene draws a candidate into a real settled Shore at the real ask level, with
// the waterline read off the shore rather than asserted, plus the real ask
// balloon at the row drawScene puts it on (top - len(rows)) and the real
// litter baseline (crab_kittens.go: ky = py + 7 - ch).
func drawInto(w, h int, level float64, q []string, xoff int, label string) {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	sandTop := sh.SandTop()

	const catW = 12
	right := 2 + w/32
	span := w / 16
	if span > 6 {
		span = 6
	}
	if span < 1 {
		span = 1
	}
	catX := w - catW - right
	sandTo := catX - 1 - span
	top := h - 2 - 7 // the shipped anchor, chh from Size() which stays 12x7

	grid := make([][]rune, h)
	for y := range grid {
		grid[y] = []rune(strings.Repeat(" ", w))
		mark := '~'
		if y >= sandTop {
			mark = '.'
		}
		for x := 0; x < w; x++ {
			grid[y][x] = mark
		}
	}
	put := func(rows []string, x, y int) {
		for dy, r := range rows {
			for dx, ch := range []rune(r) {
				if ch == ' ' {
					continue
				}
				gy, gx := y+dy, x+dx
				if gy >= 0 && gy < h && gx >= 0 && gx < w {
					grid[gy][gx] = ch
				}
			}
		}
	}
	// the litter, at the real baseline: two crablets of the 5-cell tier
	crab := companion.ParseBitmap(companion.CrabletSmall).ToQuadrant()
	ch := len(crab)
	cw := len([]rune(crab[0]))
	ky := top + 7 - ch
	for n := 0; n < 2; n++ {
		put(crab, catX-span-1-cw-n*(cw+1), ky)
	}
	// the sand tail
	tail := []string{"read  internal/auth/handler.go   142 lines",
		"edit  internal/auth/limiter.go   +18 -2"}
	for i, ln := range tail {
		if sandTop+i < h && 2+len(ln) <= sandTo {
			put([]string{ln}, 2, sandTop+i)
		}
	}
	// the companion, then the balloon at drawScene's row
	put(q, catX+xoff, top)
	bal := companion.MirrorTail(companion.Bubble("allow Bash?"))
	head := catX + 6 // lay.CatX + HeadCol()
	bx := head - companion.TailCol(bal)
	if bw := len([]rune(bal[0])); bx+bw > w {
		bx = w - bw
	}
	put(bal, bx, top-len(bal))

	fmt.Printf("--- %s  %dx%d level %.2f   waterline row %d, %d dry rows, companion top %d, feet %d\n",
		label, w, h, level, sandTop, h-sandTop, top, top+len(q)-1)
	for y := 0; y < h; y++ {
		tag := "    "
		if y == sandTop {
			tag = "W-> "
		}
		fmt.Printf("%s%2d |%s|\n", tag, y, string(grid[y]))
	}
	fmt.Println()
}

func scenes() {
	fmt.Println("\n=== 17. DRAWN INTO A REAL SHORE at the median ask level (0.59), with the")
	fmt.Println("    waterline READ OFF the shore, the real ask balloon at drawScene's row,")
	fmt.Println("    two crablets at the real litter baseline, and the real sand tail.")
	fmt.Println("    '~' is sea, '.' is dry sand, 'W->' marks the waterline row.")
	fmt.Println()
	shipQ := renderM(join(crabAskRows(), crabLowerRows()))
	shipQ = stamp(stamp(shipQ, []string{"O"}, 7, 2), []string{"O"}, 4, 2)
	pickQ := renderM(join(pickUpperRows(), pickLower))
	eq := render(pickEye)
	pickQ = stamp(stamp(pickQ, eq, 5, 1), eq, 9, 1)
	aQ := renderM(join(aUpper, aLower))
	ae := render(aEye)
	aQ = stamp(stamp(aQ, ae, 4, 2), ae, 8, 2)

	for _, g := range []struct{ w, h int }{{80, 24}, {124, 22}} {
		drawInto(g.w, g.h, 0.59, shipQ, 0, "SHIPPED 12x7 at catX")
		drawInto(g.w, g.h, 0.59, aQ, -1, "A 14x9 at catX-1")
		drawInto(g.w, g.h, 0.59, pickQ, -2, "PICK 16x9 at catX-2")
	}
}
