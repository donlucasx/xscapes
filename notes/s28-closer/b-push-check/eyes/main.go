// Part 2 of the blind check: overlay their eye sprite exactly where they say it
// goes, and test whether their 16x11 is a UNIFORM scale-up of the shipped crab
// (a push-in) or a vertical stretch (a different animal).
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

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
	".##.", ".##.", "####", "####", "####", "####", ".##.", ".##.",
}

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

// overlay stamps the eye sprite's cells over a rendered quadrant grid, exactly
// as Sprite.Draw would (non-space runes replace what is under them).
func overlay(q []string, eye []string, cols []int, row int) []string {
	out := append([]string(nil), q...)
	for _, c := range cols {
		for dy, er := range eye {
			r := []rune(out[row+dy])
			for dx, ch := range []rune(er) {
				if ch != ' ' {
					r[c+dx] = ch
				}
			}
			out[row+dy] = string(r)
		}
	}
	return out
}

// scale is a nearest-neighbour resample of a bitmap to (w,h). This is what "the
// same animal, nearer" means: one uniform magnification, no redrawing.
func scale(rows []string, w, h int) []string {
	sh, sw := len(rows), len([]rune(rows[0]))
	src := make([][]rune, sh)
	for i, r := range rows {
		src[i] = []rune(r)
	}
	out := make([]string, h)
	for y := 0; y < h; y++ {
		var sb strings.Builder
		sy := y * sh / h
		for x := 0; x < w; x++ {
			sx := x * sw / w
			sb.WriteRune(src[sy][sx])
		}
		out[y] = sb.String()
	}
	return out
}

func printQ(label string, q []string) {
	w := 0
	for _, r := range q {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	fmt.Printf("%s  (%d x %d cells)\n", label, w, len(q))
	fmt.Println("    +" + strings.Repeat("-", w) + "+")
	for i, r := range q {
		fmt.Printf(" %2d |%s|\n", i, r)
	}
	fmt.Println("    +" + strings.Repeat("-", w) + "+")
}

func pair(la, lb string, a, b []string) {
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
	fmt.Printf("   %-*s      %s\n", wa+2, la, lb)
	fmt.Printf("  +%s+      +%s+\n", strings.Repeat("-", wa), strings.Repeat("-", wb))
	for i := 0; i < n; i++ {
		ra, rb := strings.Repeat(" ", wa), strings.Repeat(" ", wb)
		if i < len(a) {
			ra = a[i] + strings.Repeat(" ", wa-len([]rune(a[i])))
		}
		if i < len(b) {
			rb = b[i] + strings.Repeat(" ", wb-len([]rune(b[i])))
		}
		fmt.Printf("%2d|%s|   %2d |%s|\n", i, ra, i, rb)
	}
	fmt.Printf("  +%s+      +%s+\n", strings.Repeat("-", wa), strings.Repeat("-", wb))
}

func main() {
	theirs := append(append([]string{}, theirUpper...), theirLower...)
	tq := companion.ParseBitmap(theirs).ToQuadrant()
	ship := append(append([]string{}, crabAsk...), crabLower...)
	sq := companion.ParseBitmap(ship).ToQuadrant()

	fmt.Println("== A. THEIR SPRITE WITH THEIR EYES OVERLAID (cells {5,9}, rows 2-3) ==")
	printQ("theirs + eyes", overlay(tq, companion.ParseBitmap(theirEye).ToQuadrant(), []int{5, 9}, 2))

	fmt.Println("== A2. MIRRORED, with the SAME eye cells (they claim mirror-exact) ==")
	tqm := companion.ParseBitmap(theirs).Mirrored().ToQuadrant()
	printQ("theirs mirrored + eyes", overlay(tqm, companion.ParseBitmap(theirEye).ToQuadrant(), []int{5, 9}, 2))

	// Is {5,9} actually mirror-exact for a 16-cell body? mirror maps cell c to
	// w-1-c for a 1-cell feature; a 2-cell eye at left-cell c maps to left-cell
	// w-2-c. 16-2-5 = 9, 16-2-9 = 5. Check the stalk pixels directly.
	fmt.Println("== A3. is the eye anchor really mirror-exact? checked in PIXELS ==")
	b := companion.ParseBitmap(theirs)
	m := b.Mirrored()
	same := true
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if bAt(b, x, y) != bAt(m, x, y) {
				same = false
			}
		}
	}
	fmt.Printf("  whole bitmap symmetric about the vertical axis? %v\n", same)
	fmt.Println("  stalk columns in the authored frame:", inkCols(theirUpper, 10))
	fmt.Println("  stalk columns in the mirrored frame :", inkCols(mirrorRows(theirUpper), 10))
	fmt.Println()

	fmt.Println("== B. IS IT A PUSH-IN? theirs vs a UNIFORM scale-up of the shipped crab ==")
	fmt.Printf("shipped source: %d x %d px.  theirs: %d x %d px.\n",
		len([]rune(ship[0])), len(ship), len([]rune(theirs[0])), len(theirs))
	fmt.Printf("width  %d -> %d = %.3fx\n", 24, 32, 32.0/24)
	fmt.Printf("height %d -> %d = %.3fx\n", 28, 44, 44.0/28)
	sw, tw := 24.0, 32.0
	uh := int(28*(tw/sw) + 0.5)
	fmt.Printf("a UNIFORM 1.333x push-in of the shipped 24x28 would be 32 x %d px = 16 x %d cells\n",
		uh, (uh+3)/4)
	up := scale(ship, 32, uh)
	uq := companion.ParseBitmap(up).ToQuadrant()
	pair("UNIFORM 1.33x of shipped", "THEIRS 16x11", uq, tq)
	fmt.Printf("\ntheirs is %d source rows taller than a uniform push-in (%d vs %d) = %.0f%% vertical stretch\n",
		44-37, 44, 37, 100*(44.0/37-1))

	fmt.Println()
	fmt.Println("== C. THE SHELL ALONE, aspect ratio ==")
	fmt.Printf("shipped crabLower: %d x %d px, aspect %.2f:1\n", 24, 16, 24.0/16)
	fmt.Printf("their lower      : %d x %d px, aspect %.2f:1\n", 32, 24, 32.0/24)
	fmt.Println("shipped shell:")
	printQ("crabLower", companion.ParseBitmap(crabLower).ToQuadrant())
	fmt.Println("their shell:")
	printQ("their lower", companion.ParseBitmap(theirLower).ToQuadrant())

	fmt.Println("== D. widest ink row and where the mass sits ==")
	rowInk("shipped", sq)
	fmt.Println()
	rowInk("theirs ", tq)
}

func bAt(b *companion.Bitmap, x, y int) bool {
	return b.On[y*b.W+x]
}

func mirrorRows(rows []string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		rr := []rune(r)
		for a, z := 0, len(rr)-1; a < z; a, z = a+1, z-1 {
			rr[a], rr[z] = rr[z], rr[a]
		}
		out[i] = string(rr)
	}
	return out
}

func inkCols(rows []string, row int) []int {
	var out []int
	for x, c := range []rune(rows[row]) {
		if c == '#' {
			out = append(out, x)
		}
	}
	return out
}

func rowInk(label string, q []string) {
	fmt.Printf("%s ink per cell row:\n", label)
	for i, r := range q {
		n, first, last := 0, -1, -1
		for x, c := range []rune(r) {
			if c != ' ' {
				n++
				if first < 0 {
					first = x
				}
				last = x
			}
		}
		span := 0
		if first >= 0 {
			span = last - first + 1
		}
		fmt.Printf("   row %2d: %2d cells inked, span %2d  %s\n", i, n, span, r)
	}
}
