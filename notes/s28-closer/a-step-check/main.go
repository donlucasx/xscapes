// Blind check of the "minimal step" candidate (notes/s28-closer/a-step).
//
// Nothing here is theirs by import: their rows are PASTED as literals below,
// exactly as returned, and run through the shipped companion.ParseBitmap /
// ToQuadrant so the render is the medium's answer and not mine.
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// ---- THEIR ROWS, pasted verbatim ------------------------------------------

var theirUpper = []string{
	"....................###..###",
	"....................###..###",
	"....................###..###",
	"....................########",
	"###..###............########",
	"###..###.............######.",
	"###..###..............####..",
	"########..............####..",
	"########..............####..",
	".######..##......##...####..",
	"..####...##......##...####..",
	"..####...##......##...####..",
	"..####..####....####..####..",
	"..####..####....####..####..",
	"..########################..",
	".##########################.",
}

var theirLower = []string{
	"############################",
	"############################",
	".##########################.",
	"..########################..",
	"...######################...",
	"....####################....",
	"..##...##############...##..",
	"..##...##############...##..",
	".##.....############.....##.",
	".##.....############.....##.",
	"##.......##########.......##",
	"##.......##########.......##",
	"##........########........##",
	"##........########........##",
	".#.........######.........#.",
	".#.........######.........#.",
	".#..........####..........#.",
	".#..........####..........#.",
	".#........................#.",
	".#........................#.",
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

var theirPupil = []string{
	"....",
	"....",
	".##.",
	".##.",
	"....",
	"....",
	"....",
	"....",
}

// ---- TODAY'S SHIPPED CRAB, copied from internal/companion/crab.go ----------
// Copied rather than imported because crabAsk / crabLower are unexported. Any
// drift between this copy and the source would show up as a different render,
// so the comparison is checked below against companion.NewCrab()'s own Size().

var todayAsk = []string{
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

var todayLower = []string{
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

// ---- helpers ---------------------------------------------------------------

// checkRagged reports every row whose rune length differs from row 0. Doing it
// BEFORE ParseBitmap matters: ParseBitmap panics on ragged art, which tells you
// that something is wrong but not what.
func checkRagged(name string, rows []string) bool {
	want := len([]rune(rows[0]))
	ok := true
	for i, r := range rows {
		if n := len([]rune(r)); n != want {
			fmt.Printf("  RAGGED %s row %d: %d runes, want %d\n", name, i, n, want)
			ok = false
		}
	}
	fmt.Printf("  %-12s %d rows x %d cols  -> %d cells wide x %d cells tall  (equal-length: %v)\n",
		name, len(rows), want, want/2, len(rows)/4, ok)
	return ok
}

func ruler(w int) {
	var tens, ones strings.Builder
	for i := 0; i < w; i++ {
		if i%10 == 0 && i > 0 {
			tens.WriteByte(byte('0' + (i/10)%10))
		} else {
			tens.WriteByte(' ')
		}
		ones.WriteByte(byte('0' + i%10))
	}
	fmt.Println("     " + tens.String())
	fmt.Println("     " + ones.String())
}

func show(title string, q []string) {
	w := 0
	for _, r := range q {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	fmt.Printf("\n%s   (%d cells wide x %d cells tall)\n", title, w, len(q))
	ruler(w)
	for i, r := range q {
		fmt.Printf("  %2d |%s|\n", i, r)
	}
}

// sideBySide prints two renders on the same lines, bottom-aligned, because the
// question "is it nearer" is a question about the same feet-line.
func sideBySide(lt, rt string, l, r []string, gap int) {
	lw := 0
	for _, s := range l {
		if n := len([]rune(s)); n > lw {
			lw = n
		}
	}
	h := len(l)
	if len(r) > h {
		h = len(r)
	}
	fmt.Printf("\n%s   %s%s\n", lt, strings.Repeat(" ", max(0, lw+gap-len(lt))), rt)
	for i := 0; i < h; i++ {
		// bottom-align: both sprites stand on the same last line
		li, ri := i-(h-len(l)), i-(h-len(r))
		ls, rs := strings.Repeat(" ", lw), ""
		if li >= 0 {
			ls = l[li] + strings.Repeat(" ", lw-len([]rune(l[li])))
		}
		if ri >= 0 {
			rs = r[ri]
		}
		fmt.Printf("  %s%s%s\n", ls, strings.Repeat(" ", gap), rs)
	}
}

// sideBySideTop is sideBySide with the TOPS aligned instead of the feet, which
// is the geometry their answer proposes: top = H-2-7 whatever the sprite's
// height, so the head holds still and the body grows downward.
func sideBySideTop(lt, rt string, l, r []string, gap int) {
	lw := 0
	for _, s := range l {
		if n := len([]rune(s)); n > lw {
			lw = n
		}
	}
	h := len(l)
	if len(r) > h {
		h = len(r)
	}
	fmt.Printf("\n%s   %s%s\n", lt, strings.Repeat(" ", max(0, lw+gap-len(lt))), rt)
	for i := 0; i < h; i++ {
		ls, rs := strings.Repeat(" ", lw), ""
		if i < len(l) {
			ls = l[i] + strings.Repeat(" ", lw-len([]rune(l[i])))
		}
		if i < len(r) {
			rs = r[i]
		}
		mark := "   "
		if i >= len(l) {
			mark = "<< "
		}
		fmt.Printf("  %s%s%s%s\n", mark, ls, strings.Repeat(" ", gap), rs)
	}
	fmt.Println("     (rows marked << are the two free rows under today's feet)")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// overlay stamps an eye render onto a body render at a cell position, so the
// face can be judged where it will actually sit rather than in isolation.
func overlay(body, eye []string, cx, cy int) []string {
	out := make([][]rune, len(body))
	for i, r := range body {
		out[i] = []rune(r)
	}
	for dy, row := range eye {
		for dx, r := range []rune(row) {
			if r == ' ' {
				continue
			}
			y, x := cy+dy, cx+dx
			if y < 0 || y >= len(out) || x < 0 || x >= len(out[y]) {
				continue
			}
			out[y][x] = r
		}
	}
	res := make([]string, len(out))
	for i, r := range out {
		res[i] = string(r)
	}
	return res
}

func main() {
	fmt.Println("=== 1. ROW LENGTHS AND DECLARED SIZE ===")
	okU := checkRagged("upperRows", theirUpper)
	okL := checkRagged("lowerRows", theirLower)
	okE := checkRagged("eyeRows", theirEye)
	okP := checkRagged("nearPupil", theirPupil)
	if !(okU && okL && okE && okP) {
		fmt.Println("  ** at least one sprite is ragged; ParseBitmap would panic **")
	}
	if len([]rune(theirUpper[0])) != len([]rune(theirLower[0])) {
		fmt.Println("  ** upper and lower are different widths; they cannot be joined **")
	}
	body := append(append([]string{}, theirUpper...), theirLower...)
	fmt.Printf("  joined       %d rows x %d cols -> %d x %d cells (they claim 14 x 9)\n",
		len(body), len([]rune(body[0])), len([]rune(body[0]))/2, len(body)/4)

	todayRows := append(append([]string{}, todayAsk...), todayLower...)
	fmt.Printf("  today        %d rows x %d cols -> %d x %d cells\n",
		len(todayRows), len([]rune(todayRows[0])), len([]rune(todayRows[0]))/2, len(todayRows)/4)
	cw, ch := companion.NewCrab().Size()
	fmt.Printf("  companion.NewCrab().Size() = %d x %d cells (checks my copy of the shipped rows)\n", cw, ch)

	fmt.Println("\n=== 2. THEIR SPRITE, RENDERED BY ME ===")
	theirs := companion.ParseBitmap(body).ToQuadrant()
	show("THEIRS, ask pose, unmirrored (authored facing)", theirs)

	fmt.Println("\n=== 3. MIRRORED, which is the SHIPPED composition ===")
	mir := companion.ParseBitmap(body).Mirrored().ToQuadrant()
	show("THEIRS, mirrored", mir)

	fmt.Println("\n=== 4. TODAY'S SHIPPED crabAsk + crabLower, same renderer ===")
	today := companion.ParseBitmap(todayRows).ToQuadrant()
	show("TODAY (crabAsk), unmirrored", today)
	todayMir := companion.ParseBitmap(todayRows).Mirrored().ToQuadrant()
	show("TODAY (crabAsk), mirrored", todayMir)

	fmt.Println("\n=== 5. SIDE BY SIDE, BOTTOM-ALIGNED (both standing on the same line) ===")
	sideBySide("TODAY 12x7", "THEIRS 14x9", today, theirs, 6)
	sideBySide("TODAY mirrored", "THEIRS mirrored", todayMir, mir, 6)

	fmt.Println("\n=== 6. THE EYE, and where it lands ===")
	eye := companion.ParseBitmap(theirEye).ToQuadrant()
	show("eyeRows alone", eye)
	pup := companion.ParseBitmap(theirPupil).ToQuadrant()
	show("nearPupil alone", pup)

	// Their claim: cells {4,5} and {8,9}, cell rows 2-3.
	withEyes := overlay(theirs, eye, 4, 2)
	withEyes = overlay(withEyes, eye, 8, 2)
	show("THEIRS with the eye bitmap stamped at {4,5} and {8,9}, rows 2-3", withEyes)

	withPup := overlay(withEyes, pup, 4, 2)
	withPup = overlay(withPup, pup, 8, 2)
	show("...and the pupil pass on top (pupil shown as its own glyphs)", withPup)

	// Mirror check on the eye coordinates.
	fmt.Println("\n  mirror arithmetic, sprite width 14 cells:")
	for _, e := range []int{4, 5, 8, 9} {
		fmt.Printf("    cell %2d  ->  %2d after the flip\n", e, 14-1-e)
	}

	// And the same stamp on the MIRRORED body, at the mirrored cells.
	mirEyes := overlay(mir, eye, 14-1-5, 2)
	mirEyes = overlay(mirEyes, eye, 14-1-9, 2)
	show("MIRRORED body with the eye stamped at the mirrored cells", mirEyes)

	fmt.Println("\n=== 7. TODAY'S EYES for comparison: single glyph 'O' at cells {4,7}, row 2 ===")
	tw := len([]rune(today[0]))
	te := make([]string, len(today))
	copy(te, today)
	for _, c := range []int{4, 7} {
		r := []rune(te[2])
		if c < len(r) {
			r[c] = 'O'
		}
		te[2] = string(r)
	}
	_ = tw
	show("TODAY with its 'O' eyes marked", te)

	fmt.Println("\n=== 8. INK WEIGHT (how much of the box is actually filled) ===")
	for _, c := range []struct {
		name string
		rows []string
	}{{"today", todayRows}, {"theirs", body}} {
		b := companion.ParseBitmap(c.rows)
		on := 0
		for _, v := range b.On {
			if v {
				on++
			}
		}
		q := b.ToQuadrant()
		cells, filled := 0, 0
		for _, row := range q {
			for _, r := range row {
				cells++
				if r != ' ' {
					filled++
				}
			}
		}
		fmt.Printf("  %-7s source px %4d/%4d = %4.1f%%   cells with ink %3d/%3d = %4.1f%%\n",
			c.name, on, b.W*b.H, 100*float64(on)/float64(b.W*b.H),
			filled, cells, 100*float64(filled)/float64(cells))
	}

	fmt.Println("\n=== 5b. TOP-ALIGNED, which is the composition they actually PROPOSE ===")
	fmt.Println("   (fixed top: the head stays put and the body grows DOWN into the two free rows)")
	sideBySideTop("TODAY 12x7", "THEIRS 14x9", todayMir, mir, 6)

	fmt.Println("\n=== 9. THE ROOM: where the box lands under each anchoring rule ===")
	// Two rules. live.go SHIPS top = H-2-chh (box bottom pinned at H-3, so the
	// head RISES as the sprite grows). Their measurement uses top = H-2-7, a
	// fixed top, so the sprite grows DOWN into the two free rows.
	type geom struct {
		name string
		h    int
		sand int // rows of beach at the median ask, given
	}
	for _, g := range []geom{
		{"80x24", 24, 7}, {"111x25", 25, 8}, {"124x22", 22, 7},
		{"143x27", 27, 8}, {"153x51", 51, 12}, {"40x12", 12, 0},
	} {
		fmt.Printf("  %-8s H=%2d  beach=%2d rows\n", g.name, g.h, g.sand)
		for _, chh := range []int{7, 8, 9} {
			shipped := g.h - 2 - chh // live.go's rule
			fixedTop := g.h - 2 - 7  // their rule
			fmt.Printf("      %d cells tall: live.go top=%2d (rows %2d..%2d)   fixed-top top=%2d (rows %2d..%2d)\n",
				chh, shipped, shipped, shipped+chh-1, fixedTop, fixedTop, fixedTop+chh-1)
		}
	}

	fmt.Println("\n=== 10. THE LITTER'S GROUND LINE (crab_kittens.go:135, ky := py + 7 - ch) ===")
	fmt.Println("  The 7 is the parent's height, HARDCODED, and the baseline is measured from")
	fmt.Println("  the parent's TOP. Under their fixed-top rule py does not move, so:")
	for _, ch := range []int{4, 3} { // crablet tiers are 4 and 3 cells tall
		py := 0
		ky := py + 7 - ch
		fmt.Printf("    crablet %d cells tall: feet at row %d   |  parent 7 tall: feet row %d  ->  aligned\n",
			ch, ky+ch-1, py+7-1)
		fmt.Printf("    %-24s                       parent 9 tall: feet row %d  ->  crablets float %d rows above the parent's feet\n",
			"", py+9-1, (py+9-1)-(ky+ch-1))
	}

	fmt.Println("\n=== 11. WHAT THE EXTRA TWO COLUMNS COST THE SAND ===")
	for _, w := range []int{40, 80, 111, 124, 143, 153} {
		right := 2 + w/32
		span := w / 16
		if span > 6 {
			span = 6
		}
		if span < 1 {
			span = 1
		}
		to12 := (w - 12 - right) - 1 - span
		to14 := (w - 14 - right) - 1 - span
		fmt.Printf("  %3d cols: sand band %3d..%3d (%2d cols) at 12 wide  ->  %3d..%3d (%2d cols) at 14 wide   %+.1f%%\n",
			w, 2, to12, to12-2+1, 2, to14, to14-2+1,
			100*float64(to14-to12)/float64(to12-2+1))
	}
}
