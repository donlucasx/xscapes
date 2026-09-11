package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// The harness. Everything here goes through companion.ParseBitmap and
// ToQuadrant, so what is printed is exactly what the scape would paint -- the
// same halve-then-pack that the shipped drawCrab uses, plus the eye glyphs
// plotted on top the way drawCrab plots them.

// pose is one authored frame: an upper half, and where the eye glyph goes.
type pose struct {
	name  string
	upper []string
	lower []string // nil means the shipped crabLower
	eyes  [2]int   // eye CELLS in the authored, unmirrored box
	row   int      // eye CELL row
	glyph rune
	note  string
}

// render composes upper+lower, packs to quadrants, and plots the eyes on top,
// exactly as drawCrab does (mirror included).
func render(p pose, mirror bool, lift int) []string {
	low := p.lower
	if low == nil {
		low = crabLower
	}
	rows := append(append([]string{}, p.upper...), low...)
	src := companion.ParseBitmap(rows)

	f := src.Blank()
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if srcAt(src, sx, sy+lift) {
				f.Set(sx, sy)
			}
		}
	}
	if mirror {
		f = f.Mirrored()
	}
	q := f.ToQuadrant()

	w := f.W / 2
	out := make([]string, len(q))
	copy(out, q)
	for _, e := range p.eyes {
		ex := e
		if mirror {
			ex = w - 1 - e
		}
		if p.row < 0 || p.row >= len(out) {
			continue
		}
		r := []rune(out[p.row])
		for len(r) <= ex {
			r = append(r, ' ')
		}
		r[ex] = p.glyph
		out[p.row] = string(r)
	}
	return out
}

// srcAt: Bitmap.at is unexported, so read the same way through ToQuadrant's
// public surface -- a one-pixel probe bitmap would be silly, so re-derive it.
func srcAt(b *companion.Bitmap, x, y int) bool {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return false
	}
	return b.On[y*b.W+x]
}

// inkCells counts painted cells, which is the only honest measure of "how big
// does it look" when the box cannot change.
func inkCells(rows []string) int {
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

// quads counts painted QUARTER cells, which is the honest measure of apparent
// size in this medium. Counting whole cells scores a half block the same as a
// full one, and half the drafts here trade one for the other -- an open pincer
// paints fewer cells and more air without the animal getting any smaller.
//
// The count is taken off the halved bitmap, because that is what the quadrant
// glyphs actually paint: 24 wide by 14 tall, so 336 is a solid 12x7 box.
func quads(p pose) int {
	low := p.lower
	if low == nil {
		low = crabLower
	}
	rows := append(append([]string{}, p.upper...), low...)
	b := companion.ParseBitmap(rows)
	n := 0
	for y := 0; y < (b.H+1)/2; y++ {
		for x := 0; x < b.W; x++ {
			if srcAt(b, x, y*2) || srcAt(b, x, y*2+1) {
				n++
			}
		}
	}
	return n
}

// topInk is the first cell row that carries any ink.
func topInk(rows []string) int {
	for i, r := range rows {
		if strings.TrimSpace(r) != "" {
			return i
		}
	}
	return len(rows)
}

// strip prints poses side by side with a gutter, so the progression reads in
// one look instead of one frame at a time.
func strip(title string, ps []pose, mirror bool, lifts []int) {
	fmt.Printf("\n%s%s\n", title, map[bool]string{true: "   (MIRRORED -- as shipped, companion on the right)", false: "   (authored)"}[mirror])
	var cols [][]string
	for i, p := range ps {
		lift := 0
		if lifts != nil {
			lift = lifts[i]
		}
		cols = append(cols, render(p, mirror, lift))
	}
	head := ""
	for i, p := range ps {
		head += fmt.Sprintf("%-14s", fmt.Sprintf("%d %s", i, p.name))
	}
	fmt.Println(head)
	for r := 0; r < 7; r++ {
		line := ""
		for _, c := range cols {
			cell := ""
			if r < len(c) {
				cell = c[r]
			}
			line += fmt.Sprintf("%-14s", cell)
		}
		fmt.Println(line)
	}
	stat := ""
	for _, c := range cols {
		stat += fmt.Sprintf("%-14s", fmt.Sprintf("%d cells", inkCells(c)))
	}
	fmt.Println(stat)
	stat = ""
	for i, c := range cols {
		stat += fmt.Sprintf("%-14s", fmt.Sprintf("eye r%d c%d,%d", ps[i].row, ps[i].eyes[0], ps[i].eyes[1]))
		_ = c
	}
	fmt.Println(stat)
}

// pair prints two poses stacked with a caption, for the A/B against the
// shipped ask.
func pair(a, b pose, mirror bool) {
	ra, rb := render(a, mirror, 0), render(b, mirror, 0)
	fmt.Printf("\n%-16s %-16s\n", a.name, b.name)
	for r := 0; r < 7; r++ {
		x, y := "", ""
		if r < len(ra) {
			x = ra[r]
		}
		if r < len(rb) {
			y = rb[r]
		}
		fmt.Printf("%-16s %-16s\n", x, y)
	}
	fmt.Printf("%-16s %-16s\n", fmt.Sprintf("%d cells, top r%d", inkCells(ra), topInk(ra)),
		fmt.Sprintf("%d cells, top r%d", inkCells(rb), topInk(rb)))
}
