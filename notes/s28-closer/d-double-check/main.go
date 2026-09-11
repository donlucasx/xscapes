// Command dcheck is a BLIND check of the d-double candidate ("the exact 2x").
// It does not import that agent's files. Their rows are pasted verbatim from
// what they returned, parsed with the real companion.ParseBitmap, and printed
// with the real ToQuadrant -- so what prints here is what the medium does with
// their art, not what they say it does.
//
//	go run ./notes/s28-closer/d-double-check
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// ---- THEIR ROWS, verbatim ----

var theirUpper = []string{
	"....................................####....####",
	"....................................####....####",
	"....................................####....####",
	"....................................####....####",
	"....................................############",
	"....................................############",
	"....................................############",
	"....................................############",
	"####....####........................############",
	"####....####........................############",
	"####....####..........................########..",
	"####....####..........................########..",
	"############............................####....",
	"############............................####....",
	"..########..............................####....",
	"..########..............................####....",
	"....####........####........####........####....",
	"....####........####........####........####....",
	"....####........####........####........####....",
	"....####........####........####........####....",
	"....########################################....",
	"....########################################....",
	"..############################################..",
	"..############################################..",
}

var theirLower = []string{
	"################################################",
	"################################################",
	"################################################",
	"################################################",
	"..############################################..",
	"..############################################..",
	"....########################################....",
	"....########################################....",
	"......####################################......",
	"......####################################......",
	"....####....########################....####....",
	"....####....########################....####....",
	"....####....########################....####....",
	"....####....########################....####....",
	"..####........####################........####..",
	"..####........####################........####..",
	"..####........####################........####..",
	"..####........####################........####..",
	"####............################............####",
	"####............################............####",
	"####............################............####",
	"####............################............####",
	"####..............############..............####",
	"####..............############..............####",
	"..##..............############..............##..",
	"..##..............############..............##..",
	"..##................########................##..",
	"..##................########................##..",
	"..##........................................##..",
	"..##........................................##..",
	"..##........................................##..",
	"..##........................................##..",
}

var theirEye = []string{
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
}

// ---- THE SHIPPED CRAB, copied from internal/companion/crab.go (crabAsk +
// crabLower) so the comparison is against the real thing at the real size. ----

var shippedAsk = []string{
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

var shippedLower = []string{
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

func lengths(name string, rows []string) (int, bool) {
	counts := map[int]int{}
	for _, r := range rows {
		counts[len([]rune(r))]++
	}
	ok := len(counts) == 1
	fmt.Printf("%-14s %2d rows, widths:", name, len(rows))
	for w, n := range counts {
		fmt.Printf(" %d(x%d)", w, n)
	}
	if ok {
		fmt.Printf("  -> UNIFORM\n")
	} else {
		fmt.Printf("  -> RAGGED (ParseBitmap panics)\n")
	}
	for i, r := range rows {
		if len([]rune(r)) != len([]rune(rows[0])) {
			fmt.Printf("    row %d is %d, expected %d: %q\n", i, len([]rune(r)), len([]rune(rows[0])), r)
		}
	}
	return len([]rune(rows[0])), ok
}

func show(title string, q []string) {
	w := 0
	for _, r := range q {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	fmt.Printf("\n%s  (%d cells wide x %d rows)\n", title, w, len(q))
	fmt.Println("   +" + strings.Repeat("-", w) + "+")
	for i, r := range q {
		fmt.Printf("%2d |%s|\n", i, r)
	}
	fmt.Println("   +" + strings.Repeat("-", w) + "+")
}

// sideBySide prints two renders on the same lines, bottom-aligned (they stand
// on the same sand), so "is it nearer" is a comparison and not a memory.
func sideBySide(aT string, a []string, bT string, b []string) {
	aw, bw := 0, 0
	for _, r := range a {
		if n := len([]rune(r)); n > aw {
			aw = n
		}
	}
	for _, r := range b {
		if n := len([]rune(r)); n > bw {
			bw = n
		}
	}
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	fmt.Printf("\n%-*s   %s\n", aw+2, aT, bT)
	for i := 0; i < n; i++ {
		ar, br := "", ""
		if k := i - (n - len(a)); k >= 0 {
			ar = a[k]
		}
		if k := i - (n - len(b)); k >= 0 {
			br = b[k]
		}
		fmt.Printf("|%-*s|   |%-*s|\n", aw, ar, bw, br)
	}
}

// halve reproduces ToQuadrant's first step so eye placement can be checked in
// SUBPIXELS, which is the unit an eye pass actually destroys.
func halve(b *companion.Bitmap) [][]bool {
	h := (b.H + 1) / 2
	out := make([][]bool, h)
	for y := 0; y < h; y++ {
		out[y] = make([]bool, b.W)
		for x := 0; x < b.W; x++ {
			a := y*2 < b.H && b.On[(y*2)*b.W+x]
			c := y*2+1 < b.H && b.On[(y*2+1)*b.W+x]
			out[y][x] = a || c
		}
	}
	return out
}

func main() {
	fmt.Println("=== 1. ROW LENGTHS ===")
	uw, uok := lengths("theirUpper", theirUpper)
	lw, lok := lengths("theirLower", theirLower)
	ew, _ := lengths("theirEye", theirEye)
	fmt.Printf("upper width %d, lower width %d, match: %v\n", uw, lw, uw == lw)
	if !uok || !lok || uw != lw {
		fmt.Println("STOP: cannot compose; ParseBitmap would panic on the join.")
		return
	}

	rows := append(append([]string{}, theirUpper...), theirLower...)
	src := companion.ParseBitmap(rows)
	fmt.Printf("\nsource: %d px wide x %d px tall (upper %d + lower %d)\n",
		src.W, src.H, len(theirUpper), len(theirLower))
	fmt.Printf("expected cells: %d wide x %d tall (W/2 x H/4)\n", src.W/2, src.H/4)

	q := src.ToQuadrant()
	show("=== 2. THEIRS, as ToQuadrant renders it ===", q)
	fmt.Printf("CLAIMED 24 cells x 14 rows. MEASURED %d x %d.\n", len([]rune(q[0])), len(q))

	qm := src.Mirrored().ToQuadrant()
	show("=== 3. THEIRS, MIRRORED (the shipped composition) ===", qm)

	// The shipped crab, same medium, for the size comparison.
	sh := companion.ParseBitmap(append(append([]string{}, shippedAsk...), shippedLower...))
	sq := sh.ToQuadrant()
	sqm := sh.Mirrored().ToQuadrant()
	show("=== 4. SHIPPED crabAsk + crabLower ===", sq)
	sideBySide("SHIPPED (12x7)", sqm, "THEIRS (mirrored)", qm)

	// --- eye pass ---
	fmt.Println("\n=== 5. THE EYE PASS ===")
	eb := companion.ParseBitmap(theirEye)
	eq := eb.ToQuadrant()
	fmt.Printf("eye bitmap %dx%d px -> %d cells wide x %d rows\n", ew, len(theirEye), len([]rune(eq[0])), len(eq))
	for _, r := range eq {
		fmt.Printf("   [%s]\n", r)
	}
	// Their claim: cell cols 8-9 and 14-15, cell rows 3-4, destroys ZERO body
	// subpixels. Count what the body actually has in those cells.
	hb := halve(src)
	check := func(cx0, cy0 int) {
		on, total := 0, 0
		for cy := cy0; cy < cy0+2; cy++ {
			for cx := cx0; cx < cx0+2; cx++ {
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						y, x := cy*2+dy, cx*2+dx
						total++
						if y < len(hb) && x < src.W && hb[y][x] {
							on++
						}
					}
				}
			}
		}
		fmt.Printf("   body subpixels under eye box at cell cols %d-%d rows %d-%d: %d of %d ON\n",
			cx0, cx0+1, cy0, cy0+1, on, total)
	}
	check(8, 3)
	check(14, 3)
	fmt.Println("   (a SOLID 2x2 eye replaces the cell outright: 'destroyed' = body ON where eye is OFF;")
	fmt.Println("    for this all-solid eyeAlert that is 0 by construction. What matters is whether the")
	fmt.Println("    body is ON there at all -- if it is 0 of 16 the eye floats with nothing behind it.)")

	// Where the shipped eye sits, for the same reading.
	shb := halve(sh)
	fmt.Println("\n   shipped eye, single cell at cols {4,7} row 2:")
	for _, cx := range []int{4, 7} {
		on := 0
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				y, x := 2*2+dy, cx*2+dx
				if y < len(shb) && shb[y][x] {
					on++
				}
			}
		}
		fmt.Printf("     col %d: %d of 4 subpixels ON under it\n", cx, on)
	}

	// --- overlay: draw the eye box into the render so placement is visible ---
	fmt.Println("\n=== 6. THEIRS with the eye box overlaid ('E') ===")
	over := make([][]rune, len(q))
	for i, r := range q {
		over[i] = []rune(r)
	}
	for _, cx0 := range []int{8, 14} {
		for cy := 3; cy < 5; cy++ {
			for cx := cx0; cx < cx0+2; cx++ {
				if cy < len(over) && cx < len(over[cy]) {
					over[cy][cx] = 'E'
				}
			}
		}
	}
	for i, r := range over {
		fmt.Printf("%2d |%s|\n", i, string(r))
	}

	// --- ink weight, the honest "is it bigger" number ---
	fmt.Println("\n=== 7. INK ===")
	countCells := func(rows []string) int {
		n := 0
		for _, r := range rows {
			for _, c := range r {
				if c != ' ' {
					n++
				}
			}
		}
		return n
	}
	fmt.Printf("shipped: %d cells wide x %d rows = %d box cells, %d inked\n",
		len([]rune(sq[0])), len(sq), len([]rune(sq[0]))*len(sq), countCells(sq))
	fmt.Printf("theirs : %d cells wide x %d rows = %d box cells, %d inked\n",
		len([]rune(q[0])), len(q), len([]rune(q[0]))*len(q), countCells(q))

	partTwo(q, sq)
	partThree(q, sq)
	partFour()
}
