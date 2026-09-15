package scenes

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE OWL, DRAWN FOR CUTE.
//
// His notes, 2026-09-15: "owl does not have to be anatomically perfect -
// gotta be likable and cute. create some alt approaches" and "owl could perch
// on a branch that appears from the right edge of the frame. try alt
// placements too, keep in mind the sub agents." The study's owl (animals.go)
// is a rectangle with two 'O' glyphs; on the vista it read as a box.
//
// The silhouettes below are generated from curves at the companion's own
// 24x28, so they go through the same halving as every animal. Three rules
// learned on the first strip, all kept here:
//
//   - FULL CELLS ARE GROUND, NOT GLYPHS. A block glyph's ink stops short of
//     the cell in every terminal measured, and the scene showed through as a
//     rule between every row of the body.
//   - EVERY CELL BRINGS ITS OWN GROUND (PlotOn), so a split background under
//     the sprite cannot win over it. That is what lost the study owl its eye
//     row and its feet on the vista.
//   - THE COAT IS A CUBE ENTRY. Taupe quantised to grey as ground and was
//     boosted to orange as a glyph, so the body and its edges disagreed.
//     Buff, 215,175,135, survives both paths unchanged, like the crab's
//     salmon; and it is neither the cat's cream nor the crab's salmon.
type OwlAlt struct {
	Name, Note string
	rows       []string
	eyeRow     int    // cell row of the eyes
	eyeL, eyeR int    // the left cell of each two-cell eye
	beakRow    int    // cell row of the beak
	wings      bool   // a folded-wing stroke down each side
	wingRows   [2]int // first and last cell row of the stroke
}

// OwlAlts are the candidates, lettered in this order on the strip.
var OwlAlts = []OwlAlt{
	{Name: "Round", Note: "One circle, no neck, tufts on top, feet under. Everything about it is soft.",
		rows: owlRound(), eyeRow: 2, eyeL: 3, eyeR: 7, beakRow: 3},
	{Name: "Egg", Note: "Taller and narrower, flat on its perch, the face in the upper third so it looks up.",
		rows: owlEgg(), eyeRow: 2, eyeL: 3, eyeR: 7, beakRow: 3},
	{Name: "Perched", Note: "A round head on a smaller body with folded wings as strokes, so it reads as a bird and not a ball.",
		rows: owlPerched(), eyeRow: 1, eyeL: 3, eyeR: 7, beakRow: 2, wings: true, wingRows: [2]int{4, 5}},
}

// OwlPick is which alternative the vista draws; OwlPlace is where: 0 on the
// mound in the meadow, 1 on a branch from the right edge.
var (
	OwlPick  = 0
	OwlPlace = 1
)

// OwlCoat is the owl's colour, a cube entry (see above).
var OwlCoat = term.RGB{R: 215, G: 175, B: 135}

var (
	owlWhite = term.RGB{R: 255, G: 255, B: 255}
	owlDark  = term.RGB{R: 38, G: 38, B: 38}
	owlBeak  = term.RGB{R: 135, G: 95, B: 0}
)

// shape builds a 24x28 bitmap from a predicate over pixels.
func shape(on func(x, y int) bool) []string {
	rows := make([]string, 28)
	for y := 0; y < 28; y++ {
		b := make([]byte, 24)
		for x := 0; x < 24; x++ {
			b[x] = '.'
			if on(x, y) {
				b[x] = '#'
			}
		}
		rows[y] = string(b)
	}
	return rows
}

func inEllipse(x, y int, cx, cy, rx, ry float64) bool {
	dx, dy := (float64(x)+0.5-cx)/rx, (float64(y)+0.5-cy)/ry
	return dx*dx+dy*dy <= 1
}

// Tufts and feet are drawn in ROW PAIRS, because the halving ORs rows 2k and
// 2k+1 and a feature on one row of a pair is either doubled or lost.
func box(x, y, x0, x1, y0, y1 int) bool { return x >= x0 && x <= x1 && y >= y0 && y <= y1 }

func owlRound() []string {
	return shape(func(x, y int) bool {
		return inEllipse(x, y, 12, 14.5, 11.5, 11.5) ||
			box(x, y, 5, 5, 0, 1) || box(x, y, 18, 18, 0, 1) ||
			box(x, y, 4, 7, 2, 3) || box(x, y, 16, 19, 2, 3) ||
			box(x, y, 7, 9, 26, 27) || box(x, y, 14, 16, 26, 27)
	})
}

func owlEgg() []string {
	return shape(func(x, y int) bool {
		return (inEllipse(x, y, 12, 15, 9.5, 13) && y >= 2 && y <= 25) ||
			box(x, y, 6, 8, 0, 1) || box(x, y, 15, 17, 0, 1) ||
			box(x, y, 4, 19, 24, 25) ||
			box(x, y, 7, 9, 26, 27) || box(x, y, 14, 16, 26, 27)
	})
}

func owlPerched() []string {
	return shape(func(x, y int) bool {
		return inEllipse(x, y, 12, 8, 8.5, 7.5) ||
			(inEllipse(x, y, 12, 19, 8.5, 8.5) && y <= 25) ||
			box(x, y, 5, 7, 0, 1) || box(x, y, 16, 18, 0, 1) ||
			box(x, y, 7, 9, 26, 27) || box(x, y, 14, 16, 26, 27)
	})
}

// plotBody draws quadrant rows with every cell as its own ground: full cells
// as a flat ground, edge cells as the glyph over what is behind.
func plotBody(c *canvas.Canvas, q []string, coat term.RGB, x, y int) {
	near := c.Near()
	for dy, row := range q {
		for dx, r := range []rune(row) {
			switch r {
			case ' ':
				continue
			case '█':
				near.PlotOn(x+dx, y+dy, ' ', coat, coat, 1)
			default:
				near.PlotOn(x+dx, y+dy, r, coat, c.BGAt(x+dx, y+dy), 1)
			}
		}
	}
}

// DrawOwl draws one alternative facing left with its top-left cell at x, y.
func DrawOwl(c *canvas.Canvas, alt *OwlAlt, coat term.RGB, x, y int) {
	near := c.Near()
	q := mustBitmap(alt.rows, 24, 28, alt.Name).Mirrored().ToQuadrant()
	companion.PlotRim(near, q, x, y)
	plotBody(c, q, coat, x, y)
	// Two cells an eye: a round pupil in the outer cell, white beside it, so
	// both eyes look the way the owl faces.
	for _, ex := range []int{alt.eyeL, alt.eyeR} {
		near.PlotOn(x+ex, y+alt.eyeRow, '●', owlDark, owlWhite, 1)
		near.PlotOn(x+ex+1, y+alt.eyeRow, ' ', owlWhite, owlWhite, 1)
	}
	near.PlotOn(x+6, y+alt.beakRow, 'v', owlBeak, coat, 1)
	if alt.wings {
		wing := term.RGB{R: 135, G: 95, B: 95}
		for r := alt.wingRows[0]; r <= alt.wingRows[1]; r++ {
			near.PlotOn(x+2, y+r, '(', wing, coat, 1)
			near.PlotOn(x+9, y+r, ')', wing, coat, 1)
		}
	}
}

// DrawOwlet draws one of the litter, 6x4 cells, facing left.
func DrawOwlet(c *canvas.Canvas, coat term.RGB, x, y int, faceLeft bool) {
	bm := mustBitmap(owlet, 12, 16, "owlet")
	if faceLeft {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	companion.PlotRim(c.Near(), q, x, y)
	plotBody(c, q, coat, x, y)
	c.Near().PlotOn(x+1, y+1, '•', owlDark, owlWhite, 1)
	c.Near().PlotOn(x+4, y+1, '•', owlDark, owlWhite, 1)
}

// The branch: from the right edge, its top edge flat where the birds sit, its
// thickness tapering to the tip. Sub-rows. It sits ABOVE the near ridge's
// treeline, against the lit ranges, or it is a dark limb on dark trees and
// vanishes -- which is what the first strip showed.
const (
	branchTip  = 100 // sub-column the branch reaches, x 50
	branchTop  = 18  // the perch, row 9's upper half
	branchOwlY = 2   // the owl's box rows 2..8, feet on row 8
)

func paintBranch(c *canvas.Canvas, col term.RGB, seed int64) {
	W2 := c.W * 2
	top := make([]int, W2)
	bot := make([]int, W2)
	for u := 0; u < W2; u++ {
		top[u], bot[u] = 999, -1
		if u < branchTip {
			continue
		}
		f := float64(u-branchTip) / float64(W2-1-branchTip) // 0 at the tip, 1 at the edge
		thick := 1 + 2.4*f
		top[u] = branchTop
		bot[u] = branchTop + int(thick+0.5) - 1
		if u < branchTip+4 {
			top[u] = branchTop + 1 // the tip thins from below
		}
	}
	for x := branchTip / 2; x < c.W; x++ {
		for y := branchTop / 2; y <= (branchTop+4)/2; y++ {
			var mask uint8
			for _, q := range []struct {
				u, v int
				bit  uint8
			}{{2 * x, 2 * y, 8}, {2*x + 1, 2 * y, 4}, {2 * x, 2*y + 1, 2}, {2*x + 1, 2*y + 1, 1}} {
				if q.v >= top[q.u] && q.v <= bot[q.u] {
					mask |= q.bit
				}
			}
			if mask == 0 {
				continue
			}
			if mask == 0b1111 {
				c.SetBG(x, y, col)
			} else {
				c.SetBGQuad(x, y, col, c.BGAt(x, y), mask)
			}
		}
	}
	// A few needles hanging from it.
	mid := c.Mid()
	for _, tx := range []int{53, 58, 62, 70, 76} {
		if scape.HashF(tx, 1, seed+97) < 0.8 {
			plot(mid, tx, branchTop/2+2, '\'', col, 1)
		}
	}
}
