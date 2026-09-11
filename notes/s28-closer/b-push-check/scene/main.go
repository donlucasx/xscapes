// Part 4: stand both crabs on the real foot line at the real median-ask
// waterline and see what is in the sea. '~' is sea, ':' is dry sand.
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

var theirUpper = []string{
	"........................###..###", "........................###..###",
	"........................###..###", "........................###..###",
	"........................###..###", "........................###..###",
	"........................########", "........................########",
	"###..###...##......##...########", "###..###...##......##...########",
	"###..###...##......##.....####..", "###..###...##......##.....####..",
	"###..###...##......##.....####..", "###..###...##......##.....####..",
	"########...##......##.....####..", "########...##......##.....####..",
	"########...##......##.....####..", "########...##......##.....####..",
	"..####.....##......##.....####..", "..####.....##......##.....####..",
}

var theirLower = []string{
	"################################", "################################",
	".##############################.", ".##############################.",
	"..############################..", "..############################..",
	"...##########################...", "...##########################...",
	"..###...################...###..", "..###...################...###..",
	".###.....##############.....###.", ".###.....##############.....###.",
	"###.......############.......###", "###.......############.......###",
	"###........##########........###", "###........##########........###",
	".##.........########.........##.", ".##.........########.........##.",
	".##..........######..........##.", ".##..........######..........##.",
	".#............####............#.", ".#............####............#.",
	".#............................#.", ".#............................#.",
}

var crabAsk = []string{
	"..................##..##", "..................##..##",
	"..................######", "..................######",
	"##..##............######", "##..##.............####.",
	"######..............##..", ".####...............##..",
	"..##....##....##....##..", "..##....##....##....##..",
	"..####################..", ".######################.",
}

var crabLower = []string{
	"########################", "########################",
	".######################.", "..####################..",
	"...##################...", "..##..############..##..",
	"..##..############..##..", ".##....##########....##.",
	".##....##########....##.", "##......########......##",
	"##......########......##", "##.......######.......##",
	".#.......######.......#.", ".#........####........#.",
	".#....................#.", ".#....................#.",
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

// place draws a W x H frame with the waterline at sandTop and the sprite's TOP
// at `top`, and reports how many of its rows are over the sea.
func place(label string, W, H, sandTop int, q []string, top, x int) {
	fmt.Printf("--- %s : frame %dx%d, SandTop row %d, sprite %dx%d at top row %d, feet row %d\n",
		label, W, H, sandTop, len([]rune(q[0])), len(q), top, top+len(q)-1)
	wet := 0
	for y := 0; y < H; y++ {
		fill := ':'
		if y < sandTop {
			fill = '~'
		}
		row := []rune(strings.Repeat(string(fill), W))
		if y >= top && y < top+len(q) {
			for dx, r := range []rune(q[y-top]) {
				if r != ' ' && x+dx < W {
					row[x+dx] = r
				}
			}
			if y < sandTop {
				wet++
			}
		}
		mark := "  "
		if y == sandTop {
			mark = "<-"
		}
		fmt.Printf("%2d %s|%s|\n", y, mark, string(row))
	}
	fmt.Printf("   ==> %d of the sprite's %d rows are OVER THE SEA\n\n", wet, len(q))
}

func main() {
	const W, H = 80, 24
	sandTop := sandTopAt(W, H, 0.59) // the median activity level at the ask
	ship := companion.ParseBitmap(append(append([]string{}, crabAsk...), crabLower...)).
		Mirrored().ToQuadrant()
	theirs := companion.ParseBitmap(append(append([]string{}, theirUpper...), theirLower...)).
		Mirrored().ToQuadrant()

	fmt.Printf("80x24 at the median ask level 0.59: SandTop = %d, dry rows = %d\n\n", sandTop, H-sandTop)

	// The shipped anchor: top = H-2-chh. Companion sits at the right margin.
	place("SHIPPED 12x7, shipped anchor", W, H, sandTop, ship, H-2-len(ship), W-17)
	place("THEIRS 16x11, shipped anchor", W, H, sandTop, theirs, H-2-len(theirs), W-21)
	// Their own proposal: spend both spare rows going down (under = 0).
	place("THEIRS 16x11, under=0 (feet on the last row)", W, H, sandTop, theirs, H-len(theirs), W-21)
}
