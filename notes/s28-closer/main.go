// Command closer measures the room a bigger companion would need, before any
// art is drawn. HIS idea, 2026-09-10: "can we animate the main Agent/companion
// to come CLOSER to the user before it prompts it?"
//
// The companion is drawn at top = H-2-chh, so the only free rows are the two
// under its feet. Everything else it could grow into is either the writing
// band or, at high activity, the sea.
//
//	go run ./notes/s28-closer
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

func main() {
	fmt.Printf("companion box today: 12 cols x 7 rows (24x28 source px; a cell is 2px wide, 4px tall)\n\n")
	fmt.Printf("%-9s %5s %6s %11s %10s %10s %8s  %s\n",
		"geom", "top", "feet", "rows under", "sand@rest", "sand@busy", "maxrows", "verdict")
	for _, g := range []struct{ w, h int }{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {143, 27}, {143, 62}, {153, 51}} {
		const chh = 7
		top := g.h - 2 - chh
		feet := top + chh - 1
		under := g.h - 1 - feet
		rest := sandTopAt(g.w, g.h, 0.0)
		busy := sandTopAt(g.w, g.h, 1.0)
		maxrows := (g.h - 1) - busy
		verdict := "room for a step up"
		if maxrows <= chh {
			verdict = "NO ROOM - head is in the water when busy"
		} else if maxrows < chh+2 {
			verdict = "one row only"
		}
		fmt.Printf("%-9s %5d %6d %11d %10d %10d %8d  %s\n",
			fmt.Sprintf("%dx%d", g.w, g.h), top, feet, under, rest, busy, maxrows, verdict)
	}
}

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
