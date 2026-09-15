package scenes

import (
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE OWL, DRAWN FOR CUTE.
//
// His notes, 2026-09-15: "owl does not have to be anatomically perfect -
// gotta be likable and cute. create some alt approaches"; "owl could perch on
// a branch that appears from the right edge of the frame. try alt placements
// too, keep in mind the sub agents"; and on the first round, "not loving the
// owl, needs to be simpler in shape. more adorable." That round was one blob
// three ways. This one is four silhouettes, authored at CELL level -- 12x7,
// the box every companion gets -- because at this size the shape is decided
// by cells and the pixels only round the corners.
//
// What makes a small character adorable, and each design has all of it: the
// head is most of the body; the eyes take up the face, with a highlight in
// the pupil; a pale face disc or belly in a second colour; a tiny beak.
//
// Three rules from the strips, kept:
//
//   - FULL CELLS ARE GROUND, NOT GLYPHS. A block glyph's ink stops short of
//     the cell and the scene shows through as a rule between the rows.
//   - EVERY CELL BRINGS ITS OWN GROUND (PlotOn), so a split background under
//     the sprite cannot win over it.
//   - EVERY COLOUR IS A CUBE ENTRY that survives both the ground path and the
//     boosted glyph path unchanged (see the probe in the commit message).
type OwlAlt struct {
	Name, Note string
	// art is 12 columns by 7 rows. '#' a whole cell; 'v' its lower half;
	// '^' its upper half; '<' its right half; '>' its left half; '.' none.
	art []string
	// light marks whole cells painted in the pale colour: a face disc, a belly.
	light []string
	// The eyes: top-left cell of each, their size in cells, and where the
	// pupil column sits inside (0 = outer, 1 = middle).
	eyeRow, eyeL, eyeR, eyeW, eyeH, pupil int
	beakRow                               int
}

// OwlAlts are the candidates, lettered in this order on the strip.
var OwlAlts = []OwlAlt{
	{Name: "Bell", Note: "A dome with a flat bottom, two tufts, a pale face disc around huge eyes. The emoji owl.",
		art: []string{
			"..^......^..",
			".v########v.",
			"############",
			"############",
			"############",
			"############",
			"...^....^...",
		},
		light: []string{
			"............",
			"....oooo....",
			"..oooooooo..",
			"..oooooooo..",
			"...oooooo...",
			"............",
			"............",
		},
		eyeRow: 2, eyeL: 2, eyeR: 7, eyeW: 3, eyeH: 2, pupil: 1, beakRow: 4},
	{Name: "Puff", Note: "A ball. No tufts, no neck, eyes that are most of the face, a belly.",
		art: []string{
			"...v####v...",
			".v########v.",
			"############",
			"############",
			"############",
			".^########^.",
			"...^....^...",
		},
		light: []string{
			"............",
			"............",
			"............",
			"............",
			"....oooo....",
			"...oooooo...",
			"............",
		},
		eyeRow: 1, eyeL: 2, eyeR: 7, eyeW: 3, eyeH: 2, pupil: 1, beakRow: 3},
	{Name: "Loaf", Note: "Wide and low, sitting, tufts at the corners, eyes far apart, a pale belly. The hunched owl.",
		art: []string{
			"............",
			".^........^.",
			"v##########v",
			"############",
			"############",
			"############",
			".^########^.",
		},
		light: []string{
			"............",
			"............",
			"............",
			"............",
			"...oooooo...",
			"...oooooo...",
			"....oooo....",
		},
		eyeRow: 2, eyeL: 1, eyeR: 8, eyeW: 3, eyeH: 2, pupil: 1, beakRow: 4},
	{Name: "Barn", Note: "Tall and narrow, no tufts, the heart-shaped face of a barn owl in pale, close-set eyes.",
		art: []string{
			"...v####v...",
			"..########..",
			"..########..",
			"..########..",
			"..########..",
			"..########..",
			"...^....^...",
		},
		light: []string{
			"............",
			"..ooo..ooo..",
			"..oooooooo..",
			"..oooooooo..",
			"...oooooo...",
			"....oooo....",
			"............",
		},
		eyeRow: 2, eyeL: 3, eyeR: 7, eyeW: 2, eyeH: 2, pupil: 0, beakRow: 4},
}

// OwlPick is which alternative the vista draws; OwlPlace is where: 0 on the
// mound in the meadow, 1 on a branch from the right edge.
var (
	OwlPick  = 0
	OwlPlace = 1
)

// The owl's colours, all cube entries that survive both paths.
var (
	OwlCoat  = term.RGB{R: 175, G: 135, B: 95}  // a warm brown
	owlLight = term.RGB{R: 255, G: 215, B: 175} // the face and the belly
	owlWhite = term.RGB{R: 255, G: 255, B: 255}
	owlDark  = term.RGB{R: 38, G: 38, B: 38}
	owlBeak  = term.RGB{R: 135, G: 95, B: 0}
)

// expand turns 12x7 cell art into the 24x28 bitmap every companion is.
func expand(art []string) []string {
	px := make([][]byte, 28)
	for y := range px {
		px[y] = []byte(strings.Repeat(".", 24))
	}
	for cy, row := range art {
		for cx, m := range row {
			x0, y0 := 2*cx, 4*cy
			set := func(dx0, dx1, dy0, dy1 int) {
				for y := y0 + dy0; y <= y0+dy1; y++ {
					for x := x0 + dx0; x <= x0+dx1; x++ {
						px[y][x] = '#'
					}
				}
			}
			switch m {
			case '#':
				set(0, 1, 0, 3)
			case 'v':
				set(0, 1, 2, 3)
			case '^':
				set(0, 1, 0, 1)
			case '<':
				set(1, 1, 0, 3)
			case '>':
				set(0, 0, 0, 3)
			}
		}
	}
	out := make([]string, 28)
	for y := range px {
		out[y] = string(px[y])
	}
	return out
}

// DrawOwl draws one alternative with its top-left cell at x, y.
func DrawOwl(c *canvas.Canvas, alt *OwlAlt, coat term.RGB, x, y int) {
	near := c.Near()
	q := mustBitmap(expand(alt.art), 24, 28, alt.Name).ToQuadrant()
	companion.PlotRim(near, q, x, y)
	for dy, row := range q {
		for dx, r := range []rune(row) {
			switch r {
			case ' ':
				continue
			case '█':
				g := coat
				if alt.light != nil && alt.light[dy][dx] == 'o' {
					g = owlLight
				}
				near.PlotOn(x+dx, y+dy, ' ', g, g, 1)
			default:
				near.PlotOn(x+dx, y+dy, r, coat, c.BGAt(x+dx, y+dy), 1)
			}
		}
	}
	// The eyes: white, with a pupil column and a highlight in its top cell.
	for _, ex := range []int{alt.eyeL, alt.eyeR} {
		for r := 0; r < alt.eyeH; r++ {
			for k := 0; k < alt.eyeW; k++ {
				near.PlotOn(x+ex+k, y+alt.eyeRow+r, ' ', owlWhite, owlWhite, 1)
			}
		}
		pc := ex + alt.pupil
		for r := 0; r < alt.eyeH; r++ {
			near.PlotOn(x+pc, y+alt.eyeRow+r, ' ', owlDark, owlDark, 1)
		}
		near.PlotOn(x+pc, y+alt.eyeRow, '˙', owlWhite, owlDark, 1)
	}
	near.PlotOn(x+5, y+alt.beakRow, '\\', owlBeak, c.BGAt(x+5, y+alt.beakRow), 1)
	near.PlotOn(x+6, y+alt.beakRow, '/', owlBeak, c.BGAt(x+6, y+alt.beakRow), 1)
}

// DrawOwlet draws one of the litter, 6x4 cells, in the same manner.
func DrawOwlet(c *canvas.Canvas, coat term.RGB, x, y int, faceLeft bool) {
	bm := mustBitmap(owlet, 12, 16, "owlet")
	if faceLeft {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	near := c.Near()
	companion.PlotRim(near, q, x, y)
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
	near.PlotOn(x+1, y+1, '•', owlDark, owlWhite, 1)
	near.PlotOn(x+4, y+1, '•', owlDark, owlWhite, 1)
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
