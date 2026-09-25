package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE PARK, THREE WAYS. His note on the blob crowns (2026-09-25): "its a lot
// of detail and currently does not read great. Try a couple different
// approaches for the park scene and how you are generating the trees and
// the fountain within". Three ways of making trees and a fountain, each
// borrowed from something the product already does well:
//
//	1  SILHOUETTE  the vista's way: a treeline of bumps in one tone a depth,
//	               quarter-cell caps, two black crowns framing the view, a
//	               single white plume over a plain pool on the lawn. Few
//	               tones, big shapes, nothing to squint at.
//	2  SPRITES     the animals' way: authored tree bitmaps through the
//	               quadrant pipeline (round, conifer, poplar) with their
//	               trunks, a stone fountain sprite with glyph water. Crisp
//	               and illustrated.
//	3  GLYPHS      the ASCII way: foliage as masses of & % @ in two greens
//	               on the lawn, conifers of ^, a pool of ~, a glyph jet.
//
// RainParkStyle 0 is the blob-crown park of rounds 6 to 9, kept to compare.
// HIS PICK, 2026-09-25: "go w sprites".
var RainParkStyle = 2

func parkNight(col rgb, nd float64) rgb { return term.Lerp(col, grey(2), nd*0.95) }

// parkStyled draws the park's middle ground in the picked style: the
// vegetation, the fountain, the lamps and the framing trees. The far side,
// the lawn and the terrace are the park's own.
func parkStyled(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, level float64) {
	switch RainParkStyle {
	case 1:
		parkSilhouette(c, p, l, seed, level)
	case 2:
		parkSprites(c, p, l, seed, level)
	case 3:
		parkGlyphs(c, p, l, seed, level)
	}
}

// parkLamp is a lamp post with its pool of light by night.
func parkLamp(c *canvas.Canvas, lx, ly int, lf float64) {
	for y := ly + 1; y <= ly+2; y++ {
		plot(c.Mid(), lx, y, '|', grey(9), 1)
	}
	if lf > 0.3 {
		plot(c.Near(), lx, ly, '*', warmLight, 0.9)
		for dy := -1; dy <= 3; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x, y := lx+dx, ly+dy
				k := 0.32 * lf * (1 - float64(dx*dx+dy*dy*2)/20)
				if k > 0 && x >= 0 && x < c.W {
					c.SetBG(x, y, term.Lerp(c.BGAt(x, y), warmLight, k))
				}
			}
		}
	}
}

// parkPool is the fountain's water: a pool of the sea's colour on rows
// 12..13 round fx, its rim in stone, lit from below by night.
func parkPool(c *canvas.Canvas, p scape.Palette, nd, lf float64, fx int) {
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	rim := parkNight(grey(15), nd)
	fill(c, fx-4, 12, fx+4, 12, p.SeaNear)
	fill(c, fx-3, 13, fx+3, 13, p.SeaNear)
	c.SetBGQuad(fx-5, 12, rim, ground(fx-5, 12), 0b0101)
	c.SetBGQuad(fx+5, 12, rim, ground(fx+5, 12), 0b1010)
	c.SetBGQuad(fx-4, 13, rim, ground(fx-4, 13), 0b1100)
	c.SetBGQuad(fx+4, 13, rim, ground(fx+4, 13), 0b1100)
	for x := fx - 4; x <= fx+4; x++ {
		c.SetBGQuad(x, 11, rim, ground(x, 11), 0b0011)
	}
	if lf > 0.3 {
		for dx := -4; dx <= 4; dx++ {
			k := 0.45 * lf * (1 - float64(dx*dx)/20)
			for _, y := range []int{12, 13} {
				c.SetBG(fx+dx, y, term.Lerp(ground(fx+dx, y), warmLight, k))
			}
		}
	}
}

// parkPlume is the fountain's water in the air, the work: one plume from
// row 11 up to six rows, its crown breaking, drops falling either side.
func parkPlume(c *canvas.Canvas, lf, level float64, fx int, seed int64, sideJets bool) {
	jet, fall := grey(23), cube(175, 215, 255)
	if lf > 0.3 {
		jet, fall = warmLight, cube(255, 215, 135)
	}
	h := 1 + int(math.Round(level*5))
	ph := 0.0
	if rainAnim {
		ph = rainT * 4
	}
	column := func(x, h int, a float64) {
		for k := 0; k < h; k++ {
			y := 11 - k
			if y < 0 {
				break
			}
			r := '|'
			if k == h-1 {
				r = '\''
			}
			plot(c.Near(), x, y, r, jet, a)
		}
	}
	column(fx, h, 1)
	if sideJets && h >= 3 {
		column(fx-2, h-2, 0.85)
		column(fx+2, h-2, 0.85)
	}
	spread := int(math.Round(level * 3))
	for d := 1; d <= spread; d++ {
		y := 11 - h + 1 + d
		if y > 11 {
			break
		}
		plot(c.Near(), fx-d-1, y, '(', fall, 0.9)
		plot(c.Near(), fx+d+1, y, ')', fall, 0.9)
		if scape.HashF(d, int(ph), seed+77) < 0.6 {
			plot(c.Near(), fx-d, y, '.', fall, 0.7)
			plot(c.Near(), fx+d, y, '.', fall, 0.7)
		}
	}
	if level > 0.15 {
		for _, dx := range []int{-2, 0, 2} {
			if scape.HashF(dx, int(ph), seed+79) < 0.7 {
				plot(c.Near(), fx+dx, 12, '~', jet, 0.8)
			}
		}
	}
}

// parkSilhouette: the vista's way.
func parkSilhouette(c *canvas.Canvas, p scape.Palette, l float64, seed int64, level float64) {
	nd, lf := 1-l, litFrac(l)
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	// A treeline of bumps behind the lawn, one tone, its caps a shade
	// lighter, each bump swaying with the wind at its own phase.
	mid := parkNight(cube(0, 95, 0), nd)
	cap := lighten(mid, 0.14)
	for k := 0; k < 15; k++ {
		cx := 1 + k*6 + int(scape.HashF(k, 101, seed)*3)
		r := 3 + int(scape.HashF(k, 102, seed)*3)
		for y := 13 - r; y <= 13; y++ {
			dy := float64(13 - y)
			hw := int(math.Round(float64(r) * math.Sqrt(math.Max(0, 1-dy*dy/float64(r*r)))))
			sw := swayAt(cx, y-13, r)
			if y == 13-r {
				for x := cx - max(1, r/2) + sw; x <= cx+max(1, r/2)+sw; x++ {
					c.SetBGQuad(x, y, cap, ground(x, y), 0b0011)
				}
				continue
			}
			t := mid
			if y == 13-r+1 {
				t = cap
			}
			fill(c, cx-hw+sw, y, cx+hw+sw, y, t)
		}
	}
	// The pool in a clearing, and one plume.
	parkPool(c, p, nd, lf, 28)
	parkPlume(c, lf, level, 28, seed, false)
	parkLamp(c, 14, 12, lf)
	parkLamp(c, 44, 12, lf)
	// Two crowns framing the view, black-green, one tone.
	near := darken(parkNight(cube(0, 95, 0), nd), 0.3)
	for _, cx := range []int{1, 76} {
		for dy := -3; dy <= 3; dy++ {
			y := 13 + dy
			hw := int(math.Round(8 * math.Sqrt(math.Max(0, 1-float64(dy*dy)/9))))
			sw := swayAt(cx, dy, 3)
			if dy == -3 {
				for x := cx - 4 + sw; x <= cx+4+sw; x++ {
					c.SetBGQuad(x, y, near, ground(x, y), 0b0011)
				}
				continue
			}
			fill(c, cx-hw+sw, y, cx+hw+sw, y, near)
		}
	}
}

// The tree sprites: crowns and trunks as bitmaps, two source rows a cell
// row, the way the animals are drawn.
var (
	treeRound = []string{
		"...########...",
		".############.",
		"##############",
		"##############",
		"##############",
		"##############",
		".############.",
		"..##########..",
		"...########...",
		".....####.....",
		".......#......",
		".......#......",
	}
	treeConifer = []string{
		"....##....",
		"....##....",
		"...####...",
		"...####...",
		"..######..",
		"..######..",
		".########.",
		".########.",
		"..######..",
		"..######..",
		"##########",
		"##########",
		".########.",
		".########.",
		"##########",
		"##########",
	}
	treePoplar = []string{
		"...##...",
		"..####..",
		".######.",
		".######.",
		"########",
		"########",
		"########",
		"########",
		".######.",
		".######.",
		".######.",
		"..####..",
		"...##...",
		"...##...",
		"...##...",
		"...##...",
	}
	treeOak = []string{
		"....########....",
		"..############..",
		".##############.",
		"################",
		"################",
		"################",
		"################",
		".##############.",
		".##############.",
		"..############..",
		"....##########..",
		"......######....",
		".......###......",
		".......###......",
		".......###......",
		".......###......",
	}
	treeWillow = []string{
		"....######....",
		"..##########..",
		".############.",
		"##############",
		"##############",
		"##.########.##",
		"##.##.##.##.##",
		"##.##.##.##.##",
		".#.##.##.##.#.",
		".#..#.##.#..#.",
		"....#.##.#....",
		"......##......",
	}
	treeBush = []string{
		"..######..",
		".########.",
		"##########",
		"##########",
		".########.",
		"..######..",
		"....##....",
		"....##....",
	}
	fountainStone = []string{
		".....##.....",
		"..########..",
		".##########.",
		"....####....",
		"....####....",
		"############",
		"############",
		"############",
	}
)

// drawTree draws a crown bitmap at (x, y) in col with a trunk under it.
func drawTree(c *canvas.Canvas, rows []string, x, y int, col, trunk rgb, sway int) {
	bm := companion.ParseBitmap(rows)
	q := bm.ToQuadrant()
	near := c.Near()
	companion.PlotRim(near, q, x+sway, y)
	(&companion.Sprite{Rows: q, Body: col}).Draw(near, x+sway, y)
	w := bm.W / 2
	h := bm.H / 4
	fill(c, x+w/2, y+h, x+w/2, y+h, trunk)
}

// parkSprites: the animals' way.
func parkSprites(c *canvas.Canvas, p scape.Palette, l float64, seed int64, level float64) {
	nd, lf := 1-l, litFrac(l)
	green := parkNight(cube(0, 135, 0), nd)
	dark := parkNight(cube(0, 95, 0), nd)
	olive := parkNight(cube(95, 135, 0), nd)
	trunk := parkNight(cube(95, 95, 95), nd)
	sway := func(cx int) int { return swayAt(cx, -1, 3) }
	light := parkNight(cube(95, 175, 95), nd)
	// Far and small first, near and big last; no two neighbours the same
	// bitmap or the same green.
	drawTree(c, treePoplar, 44, 8, olive, trunk, sway(44))
	drawTree(c, treePoplar, 60, 8, green, trunk, sway(60))
	drawTree(c, treeConifer, 50, 9, dark, trunk, sway(50))
	drawTree(c, treeOak, 4, 9, green, trunk, sway(4))
	drawTree(c, treeBush, 36, 12, olive, trunk, sway(36))
	drawTree(c, treeOak, 64, 9, olive, trunk, sway(64))
	drawTree(c, treeConifer, 73, 10, dark, trunk, sway(73))
	drawTree(c, treeWillow, 13, 11, light, trunk, sway(13))
	drawTree(c, treeRound, 55, 11, green, trunk, sway(55))
	drawTree(c, treeBush, 70, 13, dark, trunk, sway(70))
	// The fountain: stone tiers, then its water.
	stone := parkNight(grey(15), nd)
	fill(c, 24, 13, 32, 13, p.SeaNear)
	bm := companion.ParseBitmap(fountainStone)
	q := bm.ToQuadrant()
	(&companion.Sprite{Rows: q, Body: stone}).Draw(c.Near(), 25, 12)
	if lf > 0.3 {
		for x := 24; x <= 32; x++ {
			c.SetBG(x, 13, term.Lerp(c.BGAt(x, 13), warmLight, 0.4*lf))
		}
	}
	parkPlume(c, lf, level, 28, seed, true)
	parkLamp(c, 20, 11, lf)
	parkLamp(c, 38, 11, lf)
	// The framing trees, big oaks at the edges, dark.
	drawTree(c, treeOak, -4, 11, dark, trunk, sway(0))
	drawTree(c, treeOak, 74, 11, dark, trunk, sway(76))
}

// parkGlyphs: the ASCII way.
func parkGlyphs(c *canvas.Canvas, p scape.Palette, l float64, seed int64, level float64) {
	nd, lf := 1-l, litFrac(l)
	far := c.Far()
	green := parkNight(cube(0, 135, 0), nd)
	dark := parkNight(cube(0, 95, 0), nd)
	light := parkNight(cube(95, 175, 95), nd)
	trunk := parkNight(cube(135, 95, 95), nd)
	leaf := []rune("&%@&%")
	mass := func(cx, cy, rx, ry int, a, b rgb) {
		sw := swayAt(cx, -1, ry)
		for dy := -ry; dy <= ry; dy++ {
			hw := int(math.Round(float64(rx) * math.Sqrt(math.Max(0, 1-float64(dy*dy)/float64(ry*ry)))))
			for x := cx - hw; x <= cx+hw; x++ {
				h := scape.HashF(x, cy+dy, seed+110)
				col := a
				if dy > 0 || h > 0.7 {
					col = b
				}
				plot(far, x+sw, cy+dy, leaf[int(h*5)], col, 0.95)
			}
		}
		for y := cy + ry + 1; y <= 14; y++ {
			plot(far, cx, y, '|', trunk, 1)
		}
	}
	cone := func(cx, top, h int, col rgb) {
		for k := 0; k < h; k++ {
			for x := cx - k/2; x <= cx+k/2; x++ {
				plot(far, x, top+k, '^', col, 0.95)
			}
		}
		plot(far, cx, top+h, '|', trunk, 1)
	}
	mass(6, 11, 5, 2, light, green)
	mass(14, 12, 4, 2, green, dark)
	cone(46, 8, 5, dark)
	cone(51, 9, 4, dark)
	mass(60, 11, 5, 2, green, dark)
	mass(70, 11, 4, 2, light, green)
	// The pool as a ring of tildes, the jet as glyphs.
	blue := p.SeaNear
	for x := 24; x <= 32; x++ {
		plot(far, x, 12, '~', lighten(blue, 0.3), 1)
		plot(far, x, 13, '~', blue, 1)
	}
	parkPlume(c, lf, level, 28, seed, false)
	parkLamp(c, 20, 11, lf)
	parkLamp(c, 38, 11, lf)
	// The framing masses, dark.
	mass(0, 13, 8, 3, dark, darken(dark, 0.3))
	mass(78, 13, 8, 3, dark, darken(dark, 0.3))
}
