package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
)

func quad(rows []string) []string { return companion.ParseBitmap(rows).ToQuadrant() }

// frame prints a finished cell grid with a border, so a stray column or a lost
// row is visible rather than inferred.
func frame(name string, rows []string) {
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	fmt.Printf("%s  --  %d cells wide x %d tall\n", name, w, len(rows))
	fmt.Printf("   +%s+\n", rep("-", w))
	for i, r := range rows {
		fmt.Printf("%2d |%-*s|\n", i, w, r)
	}
	fmt.Printf("   +%s+\n\n", rep("-", w))
}

func showEye(name string, rows []string) {
	b := companion.ParseBitmap(rows)
	q := b.ToQuadrant()
	fmt.Printf("%s -> %d cells x %d\n", name, len([]rune(q[0])), len(q))
	for _, r := range q {
		fmt.Printf("      [%s]\n", r)
	}
	fmt.Println()
}

// overlay draws an eye bitmap's cells on top of a body's cells at the two eye
// positions, which is exactly what a second Sprite pass does: a blank cell in
// the eye sprite is transparent and lets the body show.
func overlay(body, eye []string, cells [2]int, row int) []string {
	out := make([][]rune, len(body))
	for i, r := range body {
		out[i] = []rune(r)
	}
	for _, ex := range cells {
		for dy, er := range eye {
			for dx, c := range []rune(er) {
				if c == ' ' {
					continue
				}
				y, x := row+dy, ex+dx
				if y < 0 || y >= len(out) || x < 0 || x >= len(out[y]) {
					continue
				}
				out[y][x] = 'E'
			}
		}
	}
	res := make([]string, len(out))
	for i, r := range out {
		res[i] = string(r)
	}
	return res
}

func overlayGlyph(body []string, cells [2]int, row int, g rune) []string {
	out := make([]string, len(body))
	copy(out, body)
	for _, ex := range cells {
		r := []rune(out[row])
		if ex >= 0 && ex < len(r) {
			r[ex] = g
			out[row] = string(r)
		}
	}
	return out
}

// mirrorCells maps eye cell positions through a horizontal flip. n is the eye's
// own width in cells: the flip moves its LEFT edge to w-1-(x+n-1).
func mirrorCells(c [2]int, w, n int) [2]int {
	return [2]int{w - c[1] - n, w - c[0] - n}
}

func sideBySide(names []string, blocks [][]string) {
	h := 0
	for _, b := range blocks {
		if len(b) > h {
			h = len(b)
		}
	}
	ws := make([]int, len(blocks))
	for i, b := range blocks {
		for _, r := range b {
			if n := len([]rune(r)); n > ws[i] {
				ws[i] = n
			}
		}
	}
	var head strings.Builder
	for i, n := range names {
		head.WriteString(fmt.Sprintf("  %-*s", ws[i]+2, n))
	}
	fmt.Println(head.String())
	// Feet aligned: every block sits on the same bottom row, which is how the
	// live scene draws them (top = H-2-height).
	for y := 0; y < h; y++ {
		var b strings.Builder
		for i, blk := range blocks {
			off := h - len(blk)
			cell := ""
			if y >= off {
				cell = blk[y-off]
			}
			b.WriteString(fmt.Sprintf("  %-*s", ws[i]+2, cell))
		}
		fmt.Println(strings.TrimRight(b.String(), " "))
	}
	fmt.Println()
}

// approach prints the push-in as the live loop would run it: three rungs, each
// with the half-cell lift the breathing already provides, feet on the same row.
func approach() {
	type rung struct {
		name       string
		up, low    []string
		eye        []string
		cells      [2]int
		row, cellW int
	}
	rungs := []rung{
		{"12x7 working (rest position)", shipWork, shipLower, nil, [2]int{4, 7}, 2, 12},
		{"14x9  step 1", midWork, midLower, midEye, midEyeCells, midEyeRow, 14},
		{"16x11 step 2, working", bigWork, bigLower, bigEyeSolid, bigEyeCells, bigEyeRow, 16},
		{"16x11 step 3, THE ASK", bigAsk, bigLower, bigEyeSolid, bigEyeCells, bigEyeRow, 16},
	}
	var blocks [][]string
	var names []string
	for _, r := range rungs {
		q := quad(join(r.up, r.low))
		var withEye []string
		if r.eye == nil {
			withEye = overlayGlyph(mirrored(join(r.up, r.low)), mirrorCells(r.cells, r.cellW, 1), r.row, 'o')
		} else {
			n := len([]rune(r.eye[0])) / 2
			withEye = overlay(mirrored(join(r.up, r.low)), quad(r.eye), mirrorCells(r.cells, r.cellW, n), r.row)
		}
		_ = q
		blocks = append(blocks, withEye)
		names = append(names, r.name)
	}
	sideBySide(names, blocks)

	fmt.Println("the half-cell lift, on the big pose -- the only in-between the medium has:")
	base := join(bigAsk, bigLower)
	for _, lift := range []int{0, 2} {
		fmt.Printf("  lift %d source rows:\n", lift)
		for _, r := range quad(shiftUp(base, lift)) {
			fmt.Printf("      %s\n", r)
		}
	}
	fmt.Println()
}

// shiftUp is drawCrab's breathing: read the source `lift` rows lower, so the
// body rises by half a cell when lift is 2.
func shiftUp(rows []string, lift int) []string {
	if lift == 0 {
		return rows
	}
	blank := rep(".", len([]rune(rows[0])))
	out := append([]string{}, rows[lift:]...)
	for i := 0; i < lift; i++ {
		out = append(out, blank)
	}
	return out
}

// shipCheck proves the hand-copied shipped rows in today.go are the real ones,
// by drawing the REAL companion and comparing the cells. A comparison against a
// stale copy measures nothing.
func shipCheck() {
	c := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cat := companion.New(companion.NameCrab)
	// t chosen so the breathing lift is 0 AND the blink is not firing. t=0 gives
	// sin(0)=0 (no lift) but math.Mod(0, 5.3) is 0, which is inside the 0.16
	// blink window -- the first run of this check reported two differing cells
	// and both were a blinking eye, not a copying error.
	cat.Draw(c.Near(), 0, 0, 1.0, companion.NeedsYou)
	got := strings.Split(c.RenderPlain(), "\n")

	want := overlayGlyph(quad(join(shipAsk, shipLower)), [2]int{4, 7}, 2, 'O')
	bad := 0
	for y := 0; y < len(want); y++ {
		for x, r := range []rune(want[y]) {
			if x < len([]rune(got[y])) && []rune(got[y])[x] != r {
				bad++
			}
		}
	}
	fmt.Printf("shipCheck: the copied shipped rows match the REAL crab at %d/%d cells",
		12*7-bad, 12*7)
	if bad == 0 {
		fmt.Println("  -- exact")
	} else {
		fmt.Printf("  -- %d DIFFER, the comparison below is against a stale copy\n", bad)
		for y := 0; y < 7; y++ {
			fmt.Printf("   want |%s|   got |%s|\n", want[y], string([]rune(got[y])[:12]))
		}
	}
	fmt.Println()
}
