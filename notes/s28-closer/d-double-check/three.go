package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

// sandTopAt settles a real Shore at a level and reads its waterline, the same
// way notes/s28-closer does. Copied, not imported: notes dirs are main packages.
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

// standIn draws the sprite where live.go puts it (top = H-2-chh) with the
// waterline marked, so "N rows under" is a picture and not a number.
func standIn(label string, q []string, w, h int, level float64) {
	chh := len(q)
	cw := 0
	for _, r := range q {
		if n := len([]rune(r)); n > cw {
			cw = n
		}
	}
	top := h - 2 - chh
	sandTop := sandTopAt(w, h, level)
	beach := h - sandTop
	need := chh + 2
	wet := need - beach
	if wet < 0 {
		wet = 0
	}
	fmt.Printf("\n%s  %dx%d  level %.2f   sandTop=%d  beach=%d rows  sprite %dx%d needs %d  -> %d ROWS WET\n",
		label, w, h, level, sandTop, beach, cw, chh, need, wet)

	// left gutter marks sea/sand; the sprite is drawn at its real column too.
	x0 := w - 2 - cw
	if x0 < 0 {
		x0 = 0
	}
	for y := 0; y < h; y++ {
		mark := "~~"
		if y >= sandTop {
			mark = ".."
		}
		line := strings.Repeat(" ", w)
		if y >= top && y-top < len(q) {
			row := []rune(q[y-top])
			b := []rune(line)
			for i, r := range row {
				if x0+i < len(b) {
					b[x0+i] = r
				}
			}
			line = string(b)
		}
		flag := " "
		if y >= top && y-top < len(q) && y < sandTop {
			flag = "<" // this row of the companion is in the sea
		}
		fmt.Printf("%2d %s|%s|%s\n", y, mark, strings.TrimRight(line, " "), flag)
	}
}

func partThree(theirQ, shippedQ []string) {
	fmt.Println("\n=== 10. THE ROOM, at the level his asks actually fire (p50 = 0.59) ===")
	fmt.Printf("%-9s %7s %7s %9s %9s\n", "geom", "beach", "need14", "wet(2x)", "wet(shipped)")
	for _, g := range []struct{ w, h int }{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}, {143, 62}} {
		beach := g.h - sandTopAt(g.w, g.h, 0.59)
		wet2 := 16 - beach
		if wet2 < 0 {
			wet2 = 0
		}
		wetS := 9 - beach
		if wetS < 0 {
			wetS = 0
		}
		fmt.Printf("%-9s %7d %7d %9d %9d\n", fmt.Sprintf("%dx%d", g.w, g.h), beach, 16, wet2, wetS)
	}

	fmt.Println("\n=== 11. DRAWN IN PLACE ===")
	standIn("THEIRS 2x", theirQ, 124, 22, 0.59)
	standIn("SHIPPED", shippedQ, 124, 22, 0.59)
	standIn("THEIRS 2x", theirQ, 153, 51, 0.59)
}
