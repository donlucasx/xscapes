// Blind check of candidate C ("lean", no scale change). Their rows, pasted
// verbatim from the handoff, rendered through the real ParseBitmap/ToQuadrant.
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// ---- THEIR ROWS, verbatim from the handoff JSON ----

var theirUpper = []string{
	"##..##............##..##",
	"##..##............##..##",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##...#........#...##..",
	"..##...#........#...##..",
	"..####################..",
	".######################.",
}

var theirLower = []string{
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

// ---- SHIPPED ROWS, copied verbatim out of internal/companion/crab.go ----

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

var shippedWork = []string{
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

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

var shippedDone = []string{
	"##..##............##..##",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

func join(a, b []string) []string {
	return append(append([]string{}, a...), b...)
}

// box is a solid w x h source rectangle, for ceiling arithmetic.
func box(w, h int) []string {
	rows := make([]string, h)
	for i := range rows {
		rows[i] = strings.Repeat("#", w)
	}
	return rows
}

// diff names every cell where two renders differ.
func diff(na string, a []string, nb string, b []string) {
	n := 0
	var where []string
	for y := 0; y < len(a) && y < len(b); y++ {
		ra, rb := []rune(a[y]), []rune(b[y])
		for x := 0; x < len(ra) && x < len(rb); x++ {
			if ra[x] != rb[x] {
				n++
				where = append(where, fmt.Sprintf("r%dc%d %q/%q", y, x, ra[x], rb[x]))
			}
		}
	}
	fmt.Printf("  %s vs %s: %d of 84 cells differ  %s\n", na, nb, n, strings.Join(where, "  "))
}

// widths reports every distinct row length, in order, so a ragged sprite is
// visible rather than a panic with no detail.
func widths(name string, rows []string) bool {
	ok := true
	w0 := len([]rune(rows[0]))
	for i, r := range rows {
		n := len([]rune(r))
		if n != w0 {
			fmt.Printf("  RAGGED %s row %d: %d runes, expected %d  %q\n", name, i, n, w0, r)
			ok = false
		}
	}
	fmt.Printf("  %-16s %d rows x %d cols, equal-length: %v\n", name, len(rows), w0, ok)
	return ok
}

// quarter counts painted quarter-cells: on-pixels of the VERTICALLY HALVED
// bitmap, which is the unit ToQuadrant packs 2x2 of into a cell. This is the
// measure candidate C reported, recomputed independently here.
func quarter(rows []string) (on, total int) {
	b := companion.ParseBitmap(rows)
	hh := (b.H + 1) / 2
	for y := 0; y < hh; y++ {
		for x := 0; x < b.W; x++ {
			a := y*2 < b.H && []rune(rows[y*2])[x] == '#'
			c := y*2+1 < b.H && []rune(rows[y*2+1])[x] == '#'
			if a || c {
				on++
			}
			total++
		}
	}
	return on, total
}

// inkCells counts cells that contain any ink at all -- the silhouette's
// footprint, which is what a glance actually sees.
func inkCells(q []string) int {
	n := 0
	for _, r := range q {
		for _, c := range r {
			if c != ' ' {
				n++
			}
		}
	}
	return n
}

// bbox is the tightest cell box holding ink.
func bbox(q []string) (x0, y0, x1, y1 int) {
	x0, y0 = 1<<30, 1<<30
	x1, y1 = -1, -1
	for y, r := range q {
		for x, c := range []rune(r) {
			if c != ' ' {
				if x < x0 {
					x0 = x
				}
				if x > x1 {
					x1 = x
				}
				if y < y0 {
					y0 = y
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	return
}

// withEyes overlays single-cell eye glyphs at the given cells/row, the way
// drawCrab plots them on top of the body.
func withEyes(q []string, cells [2]int, row int, glyph rune) []string {
	out := append([]string(nil), q...)
	if row < 0 || row >= len(out) {
		return out
	}
	r := []rune(out[row])
	for _, c := range cells {
		if c >= 0 && c < len(r) {
			r[c] = glyph
		}
	}
	out[row] = string(r)
	return out
}

func show(title string, rows []string, mirror bool, eyes *[2]int, eyeRow int, glyph rune) []string {
	b := companion.ParseBitmap(rows)
	if mirror {
		b = b.Mirrored()
	}
	q := b.ToQuadrant()
	drawn := q
	if eyes != nil {
		e := *eyes
		if mirror {
			w := len([]rune(q[0]))
			e = [2]int{w - 1 - e[0], w - 1 - e[1]}
		}
		drawn = withEyes(q, e, eyeRow, glyph)
	}
	fmt.Printf("\n%s  (%d cells wide x %d tall)\n", title, len([]rune(q[0])), len(q))
	for i, r := range drawn {
		fmt.Printf("  %2d |%s|\n", i, r)
	}
	return q
}

// sideBySide prints two renders on the same lines at the same scale.
func sideBySide(la, lb string, a, b []string) {
	fmt.Printf("\n  %-16s   %-16s\n", la, lb)
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		ra, rb := "", ""
		if i < len(a) {
			ra = a[i]
		}
		if i < len(b) {
			rb = b[i]
		}
		fmt.Printf("  |%-14s|   |%-14s|\n", ra, rb)
	}
}

func main() {
	fmt.Println("== 1. ROW LENGTHS ==")
	okU := widths("theirUpper", theirUpper)
	okL := widths("theirLower", theirLower)
	full := join(theirUpper, theirLower)
	okF := widths("their FULL", full)
	fmt.Printf("  all equal-length: %v\n", okU && okL && okF)

	fmt.Println("\n  lower half vs shipped crabLower:",
		strings.Join(theirLower, "|") == strings.Join(shippedLower, "|"))

	fmt.Println("\n== 2. THEIR SPRITE, RENDERED ==")
	theirQ := show("candidate C, bare bitmap", full, false, nil, 0, 0)
	show("candidate C + their eyes (cells 3,8 row 1, 'O')", full, false, &[2]int{3, 8}, 1, 'O')

	fmt.Println("\n== 3. MIRRORED (the shipped composition) ==")
	theirQM := show("candidate C, MIRRORED", full, true, nil, 0, 0)
	show("candidate C MIRRORED + eyes", full, true, &[2]int{3, 8}, 1, 'O')

	fmt.Println("\n== 4. THE SHIPPED CRAB, SAME SCALE ==")
	askQ := show("shipped crabAsk", join(shippedAsk, shippedLower), false, nil, 0, 0)
	show("shipped crabAsk + eyes (cells 4,7 row 2, 'O')", join(shippedAsk, shippedLower), false, &[2]int{4, 7}, 2, 'O')
	workQ := show("shipped crabWork", join(shippedWork, shippedLower), false, nil, 0, 0)
	show("shipped crabAsk MIRRORED", join(shippedAsk, shippedLower), true, nil, 0, 0)

	sideBySide("candidate C", "shipped crabAsk", theirQ, askQ)
	sideBySide("candidate C mir", "shipped work", theirQM, workQ)

	fmt.Println("\n== 5. MEASUREMENTS ==")
	type row struct {
		name string
		rows []string
	}
	for _, r := range []row{
		{"candidate C", full},
		{"shipped crabAsk", join(shippedAsk, shippedLower)},
		{"shipped crabWork", join(shippedWork, shippedLower)},
	} {
		on, tot := quarter(r.rows)
		b := companion.ParseBitmap(r.rows)
		q := b.ToQuadrant()
		x0, y0, x1, y1 := bbox(q)
		fmt.Printf("  %-18s quarter-cells %3d/%3d (%.1f%%)   ink cells %2d/%2d   bbox cols %d..%d rows %d..%d\n",
			r.name, on, tot, 100*float64(on)/float64(tot),
			inkCells(q), len(q)*len([]rune(q[0])), x0, x1, y0, y1)
	}

	fmt.Println("\n== 6. HeadCol arithmetic they claim ==")
	for _, e := range [][2]int{{4, 7}, {3, 8}} {
		fmt.Printf("  eyes %v -> HeadCol (a+b+1)/2 = %d ; mirrored in a 12-wide box -> {%d,%d}\n",
			e, (e[0]+e[1]+1)/2, 12-1-e[0], 12-1-e[1])
	}

	fmt.Println("\n== 6b. CELL-EXACT DIFF vs the two shipped poses ==")
	diff("candidate C", theirQ, "shipped crabAsk", askQ)
	diff("candidate C", theirQ, "shipped crabWork", workQ)
	doneQ := companion.ParseBitmap(join(shippedDone, shippedLower)).ToQuadrant()
	fmt.Println("\nshipped crabDone, for reference:")
	for i, r := range doneQ {
		fmt.Printf("  %2d |%s|\n", i, r)
	}
	diff("candidate C", theirQ, "shipped crabDone", doneQ)

	fmt.Println("\n== 6c. THE CEILING: how much apparent size is even reachable ==")
	solidUpper := make([]string, 12)
	for i := range solidUpper {
		solidUpper[i] = strings.Repeat("#", 24)
	}
	on, tot := quarter(join(solidUpper, shippedLower))
	fmt.Printf("  solid upper + shipped lower : %3d/%d quarter-cells (upper half maxed out)\n", on, tot)
	onL, _ := quarter(shippedLower)
	fmt.Printf("  shipped lower ALONE          : %3d  (of 192 quarter-cells in the lower 16 rows)\n", onL)
	fmt.Printf("  so the upper 12 rows can carry at most 144, and candidate C uses %d\n", 182-onL)
	solidAll := make([]string, 28)
	for i := range solidAll {
		solidAll[i] = strings.Repeat("#", 24)
	}
	on16, tot16 := quarter(box(32, 36))
	fmt.Printf("  a solid 12x7 box             : 336\n")
	fmt.Printf("  a solid 16x9 box (the room)  : %3d/%d\n", on16, tot16)

	fmt.Println("\n== 6d. THE THREE-BEAT APPROACH they describe, all mirrored, eyes on ==")
	// "the eye walks r2 c4,7 -> r1 c4,7 -> r1 c3,8"
	beat := func(rows []string, cells [2]int, row int, glyph rune) []string {
		b := companion.ParseBitmap(rows).Mirrored()
		q := b.ToQuadrant()
		w := len([]rune(q[0]))
		return withEyes(q, [2]int{w - 1 - cells[0], w - 1 - cells[1]}, row, glyph)
	}
	b1 := beat(join(shippedWork, shippedLower), [2]int{4, 7}, 2, 'o')
	b2 := beat(full, [2]int{4, 7}, 1, 'o')
	b3 := beat(full, [2]int{3, 8}, 1, 'O')
	fmt.Println("   beat 1 (work)     beat 2 (C, eyes 4/7)  beat 3 (C, eyes 3/8, ask)")
	for i := 0; i < 7; i++ {
		fmt.Printf("   |%s|      |%s|         |%s|\n", b1[i], b2[i], b3[i])
	}

	fmt.Println("\n== 6e. SYMMETRY AND POSE COLLISIONS ==")
	sym := func(name string, rows []string) {
		a := companion.ParseBitmap(rows).ToQuadrant()
		b := companion.ParseBitmap(rows).Mirrored().ToQuadrant()
		fmt.Printf("  %-18s mirror-identical: %v\n", name, strings.Join(a, "|") == strings.Join(b, "|"))
	}
	sym("candidate C", full)
	sym("shipped crabAsk", join(shippedAsk, shippedLower))
	sym("shipped crabWork", join(shippedWork, shippedLower))
	fmt.Println()
	diff("shipped crabWork", workQ, "shipped crabDone", doneQ)
	diff("shipped crabWork", workQ, "shipped crabAsk", askQ)

	fmt.Println("\n  the ceiling percentages they quoted, recomputed against working=188:")
	for _, v := range []struct {
		n string
		q int
	}{{"candidate C", 182}, {"shipped ask", 180}, {"solid upper", 266}, {"solid 12x7", 336}, {"solid 16x9", 576}} {
		fmt.Printf("    %-14s %3d  %+.1f%% vs working(188)   %+.1f%% vs solid 12x7 box(336)\n",
			v.n, v.q, 100*(float64(v.q)/188-1), 100*(float64(v.q)/336-1))
	}

	fmt.Println("\n== 7. THE 1px STALK, per-cell ==")
	// Their claim: the stalk at source cols 7 and 16 is one pixel wide, so it
	// half-fills its cell and shows as a quadrant corner. Cell column = col/2.
	for _, c := range []int{7, 16} {
		fmt.Printf("  source col %d lands in cell col %d (pairs with col %d)\n", c, c/2, c^1)
	}
}
