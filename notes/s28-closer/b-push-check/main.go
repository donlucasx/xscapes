// Blind check of the b-push candidate. Their rows are pasted VERBATIM below;
// nothing here is edited to make them work. Everything printed is what
// companion.ParseBitmap + Bitmap.ToQuadrant actually produce.
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// ---------------------------------------------------------------- THEIR ROWS

var theirUpper = []string{
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................########",
	"........................########",
	"###..###...##......##...########",
	"###..###...##......##...########",
	"###..###...##......##.....####..",
	"###..###...##......##.....####..",
	"###..###...##......##.....####..",
	"###..###...##......##.....####..",
	"########...##......##.....####..",
	"########...##......##.....####..",
	"########...##......##.....####..",
	"########...##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
}

var theirLower = []string{
	"################################",
	"################################",
	".##############################.",
	".##############################.",
	"..############################..",
	"..############################..",
	"...##########################...",
	"...##########################...",
	"..###...################...###..",
	"..###...################...###..",
	".###.....##############.....###.",
	".###.....##############.....###.",
	"###.......############.......###",
	"###.......############.......###",
	"###........##########........###",
	"###........##########........###",
	".##.........########.........##.",
	".##.........########.........##.",
	".##..........######..........##.",
	".##..........######..........##.",
	".#............####............#.",
	".#............####............#.",
	".#............................#.",
	".#............................#.",
}

var theirEye = []string{
	".##.",
	".##.",
	"####",
	"####",
	"####",
	"####",
	".##.",
	".##.",
}

// ------------------------------------------------- THE SHIPPED CRAB, COPIED
// Copied verbatim out of internal/companion/crab.go (unexported there) so the
// two can be put side by side at the same scale. Nothing in the repo is edited.

var crabAsk = []string{
	"..................##..##",
	"..................##..##",
	"..................######",
	"..................######",
	"##..##............######",
	"##..##.............####.",
	"######..............##..",
	".####...............##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

var crabLower = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	"..##..############..##..",
	"..##..############..##..",
	".##....##########....##.",
	".##....##########....##.",
	"##......########......##",
	"##......########......##",
	"##.......######.......##",
	".#.......######.......#.",
	".#........####........#.",
	".#....................#.",
	".#....................#.",
}

// ------------------------------------------------------------------- HELPERS

func lengths(name string, rows []string) (int, bool) {
	w := len([]rune(rows[0]))
	ok := true
	counts := map[int][]int{}
	for i, r := range rows {
		n := len([]rune(r))
		counts[n] = append(counts[n], i)
		if n != w {
			ok = false
		}
	}
	fmt.Printf("%-14s %d rows; row widths:", name, len(rows))
	for n, idx := range counts {
		fmt.Printf("  %d px x%d", n, len(idx))
		if !ok {
			fmt.Printf(" (rows %v)", idx)
		}
	}
	if ok {
		fmt.Printf("   -> RAGGED? no")
	} else {
		fmt.Printf("   -> RAGGED? YES, ParseBitmap will PANIC")
	}
	fmt.Println()
	return w, ok
}

func inkCells(q []string) (on, total int) {
	for _, r := range q {
		for _, c := range r {
			total++
			if c != ' ' {
				on++
			}
		}
	}
	return
}

func printQ(label string, q []string) {
	w, _ := 0, 0
	for _, r := range q {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	on, tot := inkCells(q)
	fmt.Printf("%s  -> %d cells wide x %d cells tall, ink %d of %d cells\n",
		label, w, len(q), on, tot)
	fmt.Println("    +" + strings.Repeat("-", w) + "+")
	for i, r := range q {
		fmt.Printf(" %2d |%s|\n", i, r)
	}
	fmt.Println("    +" + strings.Repeat("-", w) + "+")
}

// sideBySide prints two quadrant renders in one block at the same scale.
func sideBySide(la, lb string, a, b []string) {
	wa, wb := 0, 0
	for _, r := range a {
		if n := len([]rune(r)); n > wa {
			wa = n
		}
	}
	for _, r := range b {
		if n := len([]rune(r)); n > wb {
			wb = n
		}
	}
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	fmt.Printf("  %-*s     %s\n", wa+4, la, lb)
	fmt.Printf("  +%s+     +%s+\n", strings.Repeat("-", wa), strings.Repeat("-", wb))
	for i := 0; i < n; i++ {
		ra := strings.Repeat(" ", wa)
		if i < len(a) {
			ra = a[i] + strings.Repeat(" ", wa-len([]rune(a[i])))
		}
		rb := strings.Repeat(" ", wb)
		if i < len(b) {
			rb = b[i] + strings.Repeat(" ", wb-len([]rune(b[i])))
		}
		fmt.Printf("%2d|%s| %2d |%s|\n", i, ra, i, rb)
	}
	fmt.Printf("  +%s+     +%s+\n", strings.Repeat("-", wa), strings.Repeat("-", wb))
}

func main() {
	fmt.Println("== 1. ROW LENGTHS, before anything is parsed ==")
	uw, uok := lengths("their upper", theirUpper)
	lw, lok := lengths("their lower", theirLower)
	_, eok := lengths("their eye", theirEye)
	fmt.Printf("upper %d px wide, lower %d px wide -- same? %v\n", uw, lw, uw == lw)
	if !uok || !lok || !eok {
		fmt.Println("STOP: a ragged sprite parses to something other than the art they drew.")
		return
	}
	joined := append(append([]string{}, theirUpper...), theirLower...)
	fmt.Printf("joined: %d px wide x %d px tall  => predicted %d cells wide x %d cells tall\n",
		uw, len(joined), uw/2, (len(joined)+3)/4)
	fmt.Println()

	fmt.Println("== 2. THEIR SPRITE, RENDERED (ParseBitmap -> ToQuadrant) ==")
	src := companion.ParseBitmap(joined)
	q := src.ToQuadrant()
	printQ("their 16x11 candidate, AUTHORED FACING", q)
	fmt.Println()

	fmt.Println("== upper half alone (they claim 5 cell rows) ==")
	printQ("upper only", companion.ParseBitmap(theirUpper).ToQuadrant())
	fmt.Println("== lower half alone (they claim 6 cell rows) ==")
	printQ("lower only", companion.ParseBitmap(theirLower).ToQuadrant())
	fmt.Println()

	fmt.Println("== 3. MIRRORED, which is the SHIPPED composition ==")
	printQ("their 16x11, MIRRORED", src.Mirrored().ToQuadrant())
	fmt.Println()

	fmt.Println("== their eye bitmap ==")
	printQ("eye 4x8", companion.ParseBitmap(theirEye).ToQuadrant())
	fmt.Println()

	fmt.Println("== 4. THE SHIPPED CRAB AT ASK, same pipeline, same scale ==")
	ship := companion.ParseBitmap(append(append([]string{}, crabAsk...), crabLower...))
	shipQ := ship.ToQuadrant()
	printQ("shipped crabAsk + crabLower", shipQ)
	fmt.Println()

	fmt.Println("== SIDE BY SIDE, both AUTHORED FACING, same scale ==")
	sideBySide("SHIPPED 12x7", "THEIRS 16x11", shipQ, q)
	fmt.Println()
	fmt.Println("== SIDE BY SIDE, both MIRRORED (what actually ships) ==")
	sideBySide("SHIPPED mirrored", "THEIRS mirrored",
		ship.Mirrored().ToQuadrant(), src.Mirrored().ToQuadrant())
	fmt.Println()

	fmt.Println("== 5. ROOM ==")
	shipOn, shipTot := inkCells(shipQ)
	on, tot := inkCells(q)
	fmt.Printf("shipped: %d x %d cells, ink %d of %d (%.1f%%)\n",
		len([]rune(shipQ[0])), len(shipQ), shipOn, shipTot,
		100*float64(shipOn)/float64(shipTot))
	fmt.Printf("theirs : %d x %d cells, ink %d of %d (%.1f%%)\n",
		len([]rune(q[0])), len(q), on, tot, 100*float64(on)/float64(tot))
	fmt.Printf("linear growth: width %.2fx, height %.2fx\n",
		float64(len([]rune(q[0])))/float64(len([]rune(shipQ[0]))),
		float64(len(q))/float64(len(shipQ)))
	fmt.Printf("frame share of an 80x24 = 1920-cell frame: shipped %.1f%%, theirs %.1f%%\n",
		100*float64(shipOn)/1920, 100*float64(on)/1920)

	// Two rows sit under the feet (live.go: top = H-2-chh). So a sprite that is
	// N cells tall occupies rows [SandTop..] only if beach >= N - 2 ... check.
	fmt.Println()
	fmt.Println("beach rows at the median ask (brief's numbers), vs rows the sprite")
	fmt.Println("needs ABOVE its two spare rows if it stays at the shipped anchor:")
	beaches := []struct {
		geom string
		rows int
	}{{"80x24", 7}, {"111x25", 8}, {"124x22", 7}, {"143x27", 8}, {"153x51", 12}}
	for _, b := range beaches {
		fmt.Printf("  %-8s beach %2d rows | shipped needs %d above feet-2 = %d: %s | theirs needs %d: %s\n",
			b.geom, b.rows,
			len(shipQ), len(shipQ)-2, verdict(b.rows, len(shipQ)-2),
			len(q)-2, verdict(b.rows, len(q)-2))
	}
}

func verdict(beach, need int) string {
	if beach >= need {
		return fmt.Sprintf("fits (%d spare)", beach-need)
	}
	return fmt.Sprintf("OVER by %d rows -> into the sea", need-beach)
}
