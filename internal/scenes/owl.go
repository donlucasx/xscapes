package scenes

import (
	"math"
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

// OwlPick is which alternative the vista draws. OwlPlace is where, his ask
// of 2026-09-15 for alternatives: 0 on a mound in the meadow, 1 on a branch
// from the right edge above the treeline, 2 on a lower branch over the lake,
// 3 on a stump in the meadow, 4 on the end post of a fence from the right
// edge, the owlets on its rail. OwletCount is how many of the litter are drawn.
// HIS PICKS, 2026-09-15: the barn owl, on the mound, with the egg owlets.
var (
	OwlPick    = 3
	OwlPlace   = 0
	OwletCount = 2
)

// The owl's colours, all cube entries that survive both paths.
var (
	OwlCoat  = term.RGB{R: 175, G: 135, B: 95}  // a warm brown
	owlLight = term.RGB{R: 255, G: 215, B: 175} // the face and the belly
	owlWhite = term.RGB{R: 255, G: 255, B: 255}
	owlDark  = term.RGB{R: 38, G: 38, B: 38}
	owlBeak  = term.RGB{R: 215, G: 135, B: 0}
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
	// The beak is FILLED, his note: two cells of beak colour with the V on
	// them, not two strokes over the face.
	near.PlotOn(x+5, y+alt.beakRow, '\\', owlDark, owlBeak, 1)
	near.PlotOn(x+6, y+alt.beakRow, '/', owlDark, owlBeak, 1)
}

// THE LITTER. A subagent is an owlet, the way it is a kitten on the shore
// and a crablet on the beach; the count is the channel. Three looks, his
// ask of 2026-09-15 for alternatives:
//
//	0  the study's owlet bitmap, dot eyes -- the control
//	1  a chick in the parent's own style: tufts, a pale face, big eyes, a beak
//	2  a puff: a ball with eyes and nothing else
type OwletStyle struct {
	Name  string
	art   []string // 6 columns by 4 rows, the marks of OwlAlt.art
	light []string
	eyes  bool // white eyes with a dot, or the bitmap's own dots
	beak  bool
}

var OwletStyles = []OwletStyle{
	{Name: "Study owlet"},
	{Name: "Chick, the parent's style",
		art:   []string{".^..^.", "######", "######", ".^..^."},
		light: []string{"......", ".oooo.", ".oooo.", "......"},
		eyes:  true, beak: true},
	{Name: "Puff",
		art:  []string{".v##v.", "######", "######", ".^..^."},
		eyes: true},
	{Name: "Egg, pale belly",
		art:   []string{"..##..", ".####.", "######", ".^..^."},
		light: []string{"......", "......", ".oooo.", "......"},
		eyes:  true},
	{Name: "Wings up",
		art:   []string{"^.##.^", "######", "######", ".^..^."},
		light: []string{"......", ".oooo.", ".oooo.", "......"},
		eyes:  true, beak: true},
}

// OwletPick is the look the vista draws: his pick is the egg.
var OwletPick = 3

// expandSmall turns 6x4 cell art into the 12x16 bitmap the litter is.
func expandSmall(art []string) []string {
	px := make([][]byte, 16)
	for y := range px {
		px[y] = []byte(strings.Repeat(".", 12))
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
			}
		}
	}
	out := make([]string, 16)
	for y := range px {
		out[y] = string(px[y])
	}
	return out
}

// DrawOwlet draws one of the litter, 6x4 cells, in the picked look.
func DrawOwlet(c *canvas.Canvas, coat term.RGB, x, y int, faceLeft bool) {
	st := &OwletStyles[0]
	if OwletPick >= 0 && OwletPick < len(OwletStyles) {
		st = &OwletStyles[OwletPick]
	}
	var bm *companion.Bitmap
	if st.art == nil {
		bm = mustBitmap(owlet, 12, 16, "owlet")
		if faceLeft {
			bm = bm.Mirrored()
		}
	} else {
		bm = mustBitmap(expandSmall(st.art), 12, 16, st.Name)
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
				g := coat
				if st.light != nil && st.light[dy][dx] == 'o' {
					g = owlLight
				}
				near.PlotOn(x+dx, y+dy, ' ', g, g, 1)
			default:
				near.PlotOn(x+dx, y+dy, r, coat, c.BGAt(x+dx, y+dy), 1)
			}
		}
	}
	near.PlotOn(x+1, y+1, '•', owlDark, owlWhite, 1)
	near.PlotOn(x+4, y+1, '•', owlDark, owlWhite, 1)
	if st.beak {
		near.PlotOn(x+2, y+2, 'v', owlDark, owlBeak, 1)
	}
}

// The branch: from the right edge, its top edge flat where the birds sit, its
// thickness tapering to the tip. Sub-rows. It sits ABOVE the near ridge's
// treeline, against the lit ranges, or it is a dark limb on dark trees and
// vanishes -- which is what the first strip showed.
const (
	branchTip  = 88 // sub-column the branch reaches, x 44: room for three owlets
	branchTop  = 18 // the perch, row 9's upper half
	branchOwlY = 2  // the owl's box rows 2..8, feet on row 8
)

// paintBranch draws the limb with its perch at sub-row top.
func paintBranch(c *canvas.Canvas, col term.RGB, seed int64, top0 int) {
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
		top[u] = top0
		bot[u] = top0 + int(thick+0.5) - 1
		if u < branchTip+4 {
			top[u] = top0 + 1 // the tip thins from below
		}
	}
	for x := branchTip / 2; x < c.W; x++ {
		for y := top0 / 2; y <= (top0+4)/2; y++ {
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
	for _, tx := range []int{47, 53, 58, 62, 70, 76} {
		if scape.HashF(tx, 1, seed+97) < 0.8 {
			plot(mid, tx, top0/2+2, '\'', col, 1)
		}
	}
}

// THE OWL'S FIVE STATES (s35, 2026-09-16), drawn to the same vocabulary the
// crab and the cat use: the FACE is the channel. An owl is still, so nothing
// here breathes or wags; what moves is the eyes, which are two 2x2 blocks on a
// pale heart -- the biggest feature on the animal and the one a glance lands
// on. Each state is a different set of eye cells, so no two states share a
// screenshot:
//
//	resting    both lids down: a dark line across the top of each eye
//	working    open, the pupil out, one blink every seven seconds (the
//	           crab's pincer shuts on the same clock: idle motion, no meaning)
//	needs-you  WIDE: each eye a cell wider, the pupil centred, the ask
//	done       content: the eyes shut upward, '^ ^'
//	worried    amber slits: the lower row only, in the beak's colour
//
// ⚠ Not built: the crab's walk up to the screen. The owl asks with its eyes
// and the balloon; a bigger owl is a design round he has not had.
func DrawOwlPose(c *canvas.Canvas, x, y int, t float64, st companion.State) {
	pick := OwlPick
	if pick < 0 || pick >= len(OwlAlts) {
		pick = 0
	}
	alt := &OwlAlts[pick]
	DrawOwl(c, alt, OwlCoat, x, y)
	near := c.Near()
	eyes := [2]int{alt.eyeL, alt.eyeR}
	lidsDown := func() {
		for _, ex := range eyes {
			for r := 0; r < alt.eyeH; r++ {
				for k := 0; k < alt.eyeW; k++ {
					g := ' '
					if r == 0 {
						g = '-'
					}
					near.PlotOn(x+ex+k, y+alt.eyeRow+r, g, owlDark, owlLight, 1)
				}
			}
		}
	}
	switch st {
	case companion.Resting:
		lidsDown()
	case companion.Working:
		if math.Mod(t, 7) < 0.25 {
			lidsDown()
		}
	case companion.NeedsYou:
		for i, ex := range eyes {
			x0 := ex - 1 // the left eye grows outward (left), the right eye outward (right)
			if i == 1 {
				x0 = ex
			}
			for r := 0; r < alt.eyeH; r++ {
				for k := 0; k < alt.eyeW+1; k++ {
					near.PlotOn(x+x0+k, y+alt.eyeRow+r, ' ', owlWhite, owlWhite, 1)
				}
				near.PlotOn(x+x0+1, y+alt.eyeRow+r, ' ', owlDark, owlDark, 1)
			}
			near.PlotOn(x+x0+1, y+alt.eyeRow, '˙', owlWhite, owlDark, 1)
		}
	case companion.Done:
		for _, ex := range eyes {
			for r := 0; r < alt.eyeH; r++ {
				for k := 0; k < alt.eyeW; k++ {
					g := ' '
					if r == 0 {
						g = '^'
					}
					near.PlotOn(x+ex+k, y+alt.eyeRow+r, g, owlDark, owlLight, 1)
				}
			}
		}
	case companion.Worried:
		for _, ex := range eyes {
			for k := 0; k < alt.eyeW; k++ {
				near.PlotOn(x+ex+k, y+alt.eyeRow, ' ', owlLight, owlLight, 1)
				near.PlotOn(x+ex+k, y+alt.eyeRow+1, ' ', owlBeak, owlBeak, 1)
			}
			near.PlotOn(x+ex+alt.pupil, y+alt.eyeRow+1, ' ', owlDark, owlDark, 1)
		}
	}
}

// OwlHeadCol is the column of the owl's face centre inside its 12-wide box,
// where the balloon's pointer aims.
const OwlHeadCol = 5

// OwletW is the litter's pitch: an owlet is six cells wide and they sit a
// cell apart.
const OwletW = 7

// DrawOwlets draws n owlets on the grass to the left of the owl at x, the
// first nearest the owl, and stops at minX so the litter never stands in the
// fire. Returns how many fit; the count is the channel, so a frame too narrow
// for the litter shows fewer rather than piling them.
func DrawOwlets(c *canvas.Canvas, x, grassY, n, minX int) int {
	drawn := 0
	for k := 1; k <= n; k++ {
		ox := x - OwletW*k
		if ox < minX || ox < 0 {
			break
		}
		DrawOwlet(c, OwlCoat, ox, grassY, k%2 == 1)
		drawn++
	}
	return drawn
}
