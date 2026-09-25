package scenes

import (
	"math"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// FOUR VIEWS OUT OF THE SAME WINDOW. His note on round 4 (2026-09-24): "all
// three are very similar- improving in the right direction, but give me
// more diverse options, visually speaking". The skylines of rounds 2 to 4
// differed in the far layer behind the same block; these differ in what the
// window looks at.
//
//	1  ROOFTOPS   a high floor looking down: bands of roofs receding, tar,
//	              parapets, chimneys, tanks, skylights; the skyline small on
//	              the horizon; the nearest roof is the stage
//	2  THE CANYON a low floor: facades rising out of the top of the frame,
//	              no horizon, fire escapes zigzagging down them, the street's
//	              glow at the bottom; our landing is the stage
//	3  THE PARK   treetops under a big sky, lamp posts lit by night, the
//	              skyline a line behind; a terrace with a railing and pots is
//	              the stage
//	4  THE HARBOUR the city across still water with its lights laid on it, a
//	              boat; a dock is the stage. The water is STILL: the rain on
//	              the glass stays the work, and the harbour only reflects.
//
// RainView 0 is the block and skyline of rounds 2 to 4 (rain_study.go).
var RainView = 0

func paintView(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, nkit int, t, level float64) {
	switch RainView {
	case 1:
		paintRooftops(c, p, l, tn, seed, nkit)
	case 2:
		paintCanyon(c, p, l, tn, seed, nkit)
	case 3:
		paintPark(c, p, l, tn, seed, nkit, level)
	case 4:
		paintHarbour(c, p, l, tn, seed, nkit)
	}
}

var warmLight = cube(255, 215, 135)

// paintRooftops: a field of roofs at mixed depths, painted far to near:
// each building has its own top, width, roof depth and tone, so the
// parapets step instead of running in strips; the furniture is varied and
// placed off centre, and only some roofs carry a tank ("water tanks seemed
// copy pasted like a pattern", his note on the first version). The near
// roof is the stage.
func paintRooftops(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, nkit int) {
	haze, far2 := hazeAt(p, l), hazeFar(p, l)
	lf := litFrac(l)
	for x := 2; x <= 77; x++ {
		h := 3 + int(scape.HashF(x/2, 71, seed)*3)
		if x >= 10 && x <= 16 {
			h = 5
		}
		fill(c, x, h, x, 5, far2)
	}
	box(c, 30, 2, 33, 5, haze)
	box(c, 44, 3, 46, 5, haze)
	box(c, 55, 1, 58, 5, haze)
	plot(c.Mid(), 56, 0, '|', haze, 1)
	type roof struct {
		x0, x1, top, h int
		d              float64
		tone, kind     int
	}
	var roofs []roof
	for ring := 0; ring < 3; ring++ {
		x := -int(scape.HashF(ring, 60, seed) * 4)
		for k := 0; x < c.W; k++ {
			d := (float64(ring) + scape.HashF(k, 61+ring, seed)) / 3
			w := 4 + int(d*9) + int(scape.HashF(k, 62+ring, seed)*4)
			top := 6 + int(math.Round(d*6)) + int(scape.HashF(k, 63+ring, seed)*2)
			h := 1 + int(d*2) + int(scape.HashF(k, 64+ring, seed)*2)
			roofs = append(roofs, roof{x, x + w - 1, top, h, d,
				int(scape.HashF(k, 66+ring, seed)*3) - 1, int(scape.HashF(k, 67+ring, seed) * 8)})
			x += w + int(scape.HashF(k, 65+ring, seed)*2.2)
		}
	}
	sort.SliceStable(roofs, func(a, b int) bool { return roofs[a].top < roofs[b].top })
	tanks, antennas := 0, 0
	for _, r := range roofs {
		hz := 0.5 * (1 - r.d)
		tar := term.Lerp(clampGrey(greyK(tn.tar)+r.tone), p.SkyHorizon, hz)
		wall := term.Lerp(clampGrey(greyK(tn.tar)+3+r.tone), p.SkyHorizon, hz)
		x0, x1 := max(0, r.x0), min(c.W-1, r.x1)
		fill(c, x0, r.top, x1, r.top+r.h-1, tar)
		fill(c, x0, r.top, x1, r.top, lighten(tar, 0.2))
		fill(c, x1, r.top, x1, r.top+r.h-1, darken(tar, 0.2))
		// The facade under the roof, down to where the near roof begins.
		fill(c, x0, r.top+r.h, x1, 12, wall)
		fill(c, x1, r.top+r.h, x1, 12, darken(wall, 0.2))
		for wx := x0 + 1 + int(scape.HashF(r.x0, 68, seed)*2); wx < x1-1; wx += 3 {
			for y := r.top + r.h + 1; y <= 12; y += 2 {
				col := darken(wall, 0.28)
				if scape.HashF(wx, y, seed+9) < 0.22*lf {
					col = cube(175, 135, 95)
				}
				c.SetBG(wx, y, col)
			}
		}
		w := x1 - x0 + 1
		sil := darken(tar, 0.2)
		off := 1 + int(scape.HashF(r.x0, 69, seed)*float64(max(1, w-5)))
		switch r.kind {
		case 1:
			if tanks < 3 && w >= 6 {
				tw := 3
				if r.d > 0.55 && w >= 8 {
					tw = 4
				}
				waterTank(c, x0+min(off, w-tw-1), r.top, tw, 1+int(r.d*1.5), 1, sil)
				tanks++
			}
		case 2:
			bw := 2 + int(scape.HashF(r.x0, 70, seed)*3)
			bh := 1 + int(r.d*1.2)
			fill(c, x0+min(off, max(1, w-bw-1)), r.top-bh, x0+min(off, max(1, w-bw-1))+bw-1, r.top-1, sil)
		case 3:
			brick := term.Lerp(cube(135, 95, 95), tar, 0.4)
			fill(c, x0+off, r.top-1, x0+off, r.top, brick)
			if w >= 6 {
				fill(c, x0+min(off+2, w-2), r.top-1, x0+min(off+2, w-2), r.top, brick)
			}
		case 4:
			if r.h >= 2 {
				sk := darken(tar, 0.15)
				if lf > 0.5 {
					sk = cube(175, 135, 95)
				}
				fill(c, x0+off, r.top+1, x0+min(off+1+int(r.d*1.5), w-2), r.top+1, sk)
			}
		case 5:
			if r.h >= 2 {
				fill(c, x0+off, r.top+1, x0+off, r.top+1, lighten(tar, 0.3))
				if w >= 7 {
					fill(c, x0+off+2, r.top+1, x0+off+2, r.top+1, lighten(tar, 0.3))
				}
			}
		case 6:
			if antennas < 1 && r.d < 0.5 {
				plot(c.Mid(), x0+off, r.top-1, '|', sil, 1)
				plot(c.Mid(), x0+off, r.top-2, '|', sil, 1)
				antennas++
			}
		}
	}
	paintNearRoof(c, p, tn, seed, 0, c.W-1, 13, rainPaneBot, 31)
	xs := []int{24, 34, 46, 52, 28, 40, 22, 56}
	for i := 0; i < nkit && i < len(xs); i++ {
		drawKittenAt(c.Near(), xs[i], rainPaneBot-3, i%2 == 1)
	}
}

// paintCanyon: facades filling the pane, out of the top of the frame, big
// windows with sills, fire escapes zigzagging down two of them, the street
// lamps' glow at the bottom; our own landing with the kittens.
func paintCanyon(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, nkit int) {
	lf := litFrac(l)
	blocks := walkups(0, c.W-1, 0, 0, 14, 22, seed+300)
	if len(blocks) > 1 {
		blocks[1].day = cube(135, 95, 95)
	}
	if len(blocks) > 3 {
		blocks[3].day = cube(175, 135, 135)
	}
	for bi, b := range blocks {
		wall := wallOf(b, l)
		fill(c, b.x0, 0, b.x1, rainPaneBot, wall)
		fill(c, b.x1, 0, b.x1, rainPaneBot, darken(wall, 0.2))
		// A pilaster every other bay: the facade's own rhythm.
		for x := b.x0 + 6; x < b.x1-1; x += 6 {
			fill(c, x, 0, x, rainPaneBot, darken(wall, 0.07))
		}
		glass := darken(wall, 0.3)
		for x := b.x0 + 2; x+2 <= b.x1-2; x += 6 {
			for y := 1; y < rainPaneBot-1; y += 3 {
				col := glass
				if y <= 8 && scape.HashF(x, y, seed+9) < 0.22*lf {
					col = cube(175, 135, 95)
					if scape.HashF(x, y, seed+10) < 0.25 {
						col = cube(215, 175, 135)
					}
				}
				fill(c, x, y, x+2, y, col)
				fill(c, x, y+1, x+2, y+1, lighten(wall, 0.08)) // the sill
			}
		}
		if bi%2 == 1 && b.x1-b.x0 >= 16 {
			// The fire escape: landings every three rows, alternating
			// sides of a bay, a rail on each, the stair between them.
			iron := darken(wall, 0.42)
			rail := lighten(iron, 0.35)
			ex := b.x0 + 3
			for k, y := 0, 2; y < rainPaneBot-1; k, y = k+1, y+3 {
				lx := ex + (k%2)*6
				fill(c, lx, y, lx+4, y, iron)
				for x := lx; x <= lx+4; x++ {
					plot(c.Mid(), x, y-1, '=', rail, 1)
				}
				if y+3 < rainPaneBot-1 {
					nx := ex + ((k+1)%2)*6
					if nx > lx {
						line(c.Mid(), lx+4, y+1, nx+1, y+2, iron, 1)
					} else {
						line(c.Mid(), lx, y+1, nx+3, y+2, iron, 1)
					}
				}
			}
		}
	}
	// The street's glow, rising from below the sill by night.
	for y := rainPaneBot - 2; y <= rainPaneBot; y++ {
		k := 0.18 * lf * float64(y-rainPaneBot+3) / 3
		for x := 0; x < c.W; x++ {
			c.SetBG(x, y, term.Lerp(c.BGAt(x, y), warmLight, k))
		}
	}
	paintFireEscape(c, tn, 2, 38, 11, 16)
	xs := []int{10, 20, 30, 15, 25, 5, 33}
	for i := 0; i < nkit && i < len(xs); i++ {
		drawKittenAt(c.Near(), xs[i], 12, i%2 == 1)
	}
}

// paintPark: the view across a great park from a high window, composed the
// classic way: big dark trees framing both edges, a lake with a shoreline
// in the middle taking the sky, an arched bridge across its narrow end, a
// textured far canopy under the far side's towers (two twin-towered
// apartment houses and two singles, in haze), a path with its lamps along
// the near shore. Realism at this size is light: every crown is two offset
// ellipses, the lit side up and left, the shade down and right, the leaf
// texture split the same way; the shore trees show a trunk; the towers and
// the bridge lie on the lake. By night the foliage goes to charcoal (every
// green through night(); the first version left the back row blended toward
// the horizon, which is why it stayed green), the lamps make their pools,
// the far windows show. The terrace with its rail is the stage.
func paintPark(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, nkit int, level float64) {
	haze, far2 := hazeAt(p, l), hazeFar(p, l)
	lf := litFrac(l)
	nd := 1 - l
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	night := func(col rgb) rgb { return term.Lerp(col, grey(2), nd*0.95) }
	// The far side.
	for x := 2; x <= 77; x++ {
		h := 5 + int(scape.HashF(x/3, 81, seed)*2)
		if x >= 10 && x <= 16 {
			h = 7
		}
		fill(c, x, h, x, 8, far2)
	}
	pins := func(x0, y0, x1, y1 int) {
		for y := y0 + 1; y <= y1; y += 2 {
			for x := x0 + 1; x < x1; x += 2 {
				if scape.HashF(x, y, seed+12) < 0.18*lf {
					plot(c.Far(), x, y, '.', warmLight, 0.45)
				}
			}
		}
	}
	twin := func(x0 int) {
		box(c, x0, 6, x0+8, 8, haze)
		fill(c, x0, 6, x0+7, 6, lighten(haze, 0.12))
		box(c, x0+1, 3, x0+2, 5, haze)
		box(c, x0+6, 3, x0+7, 5, haze)
		for _, tx := range []int{x0 + 1, x0 + 2, x0 + 6, x0 + 7} {
			c.SetBGQuad(tx, 2, haze, ground(tx, 2), 0b0011)
		}
		plot(c.Mid(), x0+1, 1, '|', haze, 0.8)
		plot(c.Mid(), x0+7, 1, '|', haze, 0.8)
		pins(x0, 6, x0+8, 8)
		pins(x0, 3, x0+3, 5)
		pins(x0+5, 3, x0+8, 5)
	}
	twin(22)
	twin(52)
	box(c, 38, 4, 41, 8, haze)
	fill(c, 38, 4, 40, 4, lighten(haze, 0.12))
	pins(38, 4, 41, 8)
	box(c, 67, 5, 70, 8, haze)
	// A crown: the body ellipse, a lit ellipse offset up and left, a shade
	// row under; the lit side's leaves brighter, the shade side's darker.
	crown := func(cx, cy, rx, ry int, col rgb) {
		lit, shade := lighten(col, 0.14), darken(col, 0.2)
		ell := func(cx, cy, rx, ry int, t rgb, cap bool) {
			for dy := -ry; dy <= ry; dy++ {
				y := cy + dy
				hw := int(math.Round(float64(rx) * math.Sqrt(math.Max(0, 1-float64(dy*dy)/float64(ry*ry)))))
				if dy == -ry && cap {
					for x := cx - max(1, rx/2); x <= cx+max(1, rx/2); x++ {
						c.SetBGQuad(x, y, t, ground(x, y), 0b0011)
					}
					continue
				}
				sw := swayAt(cx, dy, ry)
				fill(c, cx-hw+sw, y, cx+hw+sw, y, t)
			}
		}
		ell(cx, cy, rx, ry, col, true)
		if ry > 1 {
			ell(cx-max(1, rx/4), cy-1, max(1, rx-2), ry-1, lit, false)
			fill(c, cx-rx/2, cy+ry, cx+rx/2, cy+ry, shade)
		}
	}
	// The far canopy: a field of small crowns in the haze, not a band.
	farGreen := night(term.Lerp(cube(0, 135, 0), p.SkyHorizon, 0.22))
	fill(c, 0, 9, c.W-1, 10, farGreen)
	farGreens := []rgb{farGreen, night(term.Lerp(cube(95, 135, 0), p.SkyHorizon, 0.25)), night(term.Lerp(cube(0, 95, 0), p.SkyHorizon, 0.2))}
	for k := 0; k < 40; k++ {
		cx := int(scape.HashF(k, 92, seed) * 80)
		cy := 9 + int(scape.HashF(k, 93, seed)*2)
		crown(cx, cy, 2+int(scape.HashF(k, 95, seed)*2), 1, farGreens[int(scape.HashF(k, 94, seed)*3)])
	}
	// The lawn under the whole middle ground, so the gaps between crowns
	// are grass and not sky (the "green splotch": every crown one green
	// on the sky's colour, with nothing between them).
	lawn := night(term.Lerp(cube(0, 95, 0), grey(6), 0.45))
	fill(c, 0, 11, c.W-1, 16, lawn)
	for y := 11; y <= 16; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed+61) < 0.08 {
				plot(c.Far(), x, y, '\'', lighten(lawn, 0.2), 0.5)
			}
		}
	}
	// THE PLAZA (his ruling: a fountain in a paved plaza, no lake, "add
	// some more details to the plaza"): paving in two tones, a bandstand at
	// its far side lit by night, benches at its edges, two lamps around the
	// fountain, and the fountain itself, which is the work.
	// Smaller, and in the LEFT pane so the fountain is not under the
	// mullion (his note of 2026-09-24): paving in two tones around the
	// fountain, two benches, the lamps flanking it; the bandstand in the
	// right pane; more park around them.
	if RainParkStyle != 0 {
		parkStyled(c, p, l, tn, seed, level)
		parkTerrace(c, l, tn, nkit)
		return
	}
	// Plain stone, and a lighter ring of paving round the pool (the two-tone
	// checker read as a transparency grid, his note); the bandstand in the
	// right pane; two benches.
	pave := night(grey(10))
	plazaRows := [][2]int{{22, 34}, {18, 38}, {18, 38}, {22, 34}}
	fx := 28
	for k, r := range plazaRows {
		y := 10 + k
		for x := r[0]; x <= r[1]; x++ {
			d := math.Hypot(float64(x-fx)/2, float64(y-12))
			t := pave
			if d < 3.6 {
				t = lighten(pave, 0.14)
			}
			c.SetBG(x, y, t)
		}
	}
	c.SetBGQuad(21, 10, pave, ground(21, 10), 0b0011)
	c.SetBGQuad(35, 10, pave, ground(35, 10), 0b0011)
	c.SetBGQuad(21, 13, pave, ground(21, 13), 0b1100)
	c.SetBGQuad(35, 13, pave, ground(35, 13), 0b1100)
	cream := night(cube(215, 215, 175))
	for x := 53; x <= 59; x++ {
		c.SetBGQuad(x, 8, cream, ground(x, 8), 0b0011)
	}
	fill(c, 52, 9, 60, 9, cream)
	fill(c, 53, 10, 59, 10, lighten(pave, 0.15))
	for _, x := range []int{53, 56, 59} {
		plot(c.Mid(), x, 10, '|', darken(cream, 0.4), 1)
	}
	if lf > 0.3 {
		for x := 54; x <= 58; x++ {
			c.SetBG(x, 10, term.Lerp(ground(x, 10), warmLight, 0.35*lf))
		}
	}
	bench := night(cube(135, 95, 95))
	for _, b := range [][2]int{{19, 11}, {36, 11}} {
		plot(c.Mid(), b[0], b[1], '=', bench, 1)
		plot(c.Mid(), b[0]+1, b[1], '=', bench, 1)
	}
	// THE FOUNTAIN, the centre and the work: a rimmed pool of the sea's
	// colour, the central jet up to six rows with two side jets, the crowns
	// breaking into arcs that fall either side, splash where they land,
	// ripples at the pool's edge while it runs; lit from below by night so
	// the water glows.
	rim := night(grey(15))
	pool := p.SeaNear
	fill(c, fx-4, 12, fx+4, 12, pool)
	fill(c, fx-3, 13, fx+3, 13, pool)
	c.SetBGQuad(fx-5, 12, rim, ground(fx-5, 12), 0b0101)
	c.SetBGQuad(fx+5, 12, rim, ground(fx+5, 12), 0b1010)
	c.SetBGQuad(fx-4, 13, rim, ground(fx-4, 13), 0b1100)
	c.SetBGQuad(fx+4, 13, rim, ground(fx+4, 13), 0b1100)
	for x := fx - 4; x <= fx+4; x++ {
		c.SetBGQuad(x, 11, rim, ground(x, 11), 0b0011)
	}
	jet, fall := grey(23), cube(175, 215, 255)
	if lf > 0.3 {
		jet, fall = warmLight, cube(255, 215, 135)
		for dx := -4; dx <= 4; dx++ {
			k := 0.45 * lf * (1 - float64(dx*dx)/20)
			for _, y := range []int{12, 13} {
				c.SetBG(fx+dx, y, term.Lerp(ground(fx+dx, y), warmLight, k))
			}
		}
	}
	h := 1 + int(math.Round(level*5))
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
	if h >= 3 {
		column(fx-2, h-2, 0.85)
		column(fx+2, h-2, 0.85)
	}
	// The arcs: from the crown outward and down, a parenthesis a row.
	ph := 0.0
	if rainAnim {
		ph = rainT * 4
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
		for _, dx := range []int{-4, 4} {
			if scape.HashF(dx, int(ph/2), seed+80) < 0.5*level {
				plot(c.Far(), fx+dx, 12, 'o', lighten(pool, 0.3), 0.6)
			}
		}
	}
	midG := night(cube(0, 135, 0))
	nearG := night(cube(0, 95, 0))
	trunk := night(cube(95, 95, 95))
	// KINDS of tree (his note: "diverse vegetation instead of a green
	// splotch"): round crowns in four greens, conifers, a willow, a copper
	// tree, shrubs along the path, every mid tree with a trunk and lawn
	// between them.
	greens := []rgb{midG, night(cube(95, 135, 0)), night(cube(0, 95, 0)), night(cube(95, 175, 95))}
	conifer := func(cx, top, h int, col rgb) {
		shade := darken(col, 0.25)
		for k := 0; k < h; k++ {
			y := top + k
			hw := k / 2
			fill(c, cx-hw, y, cx+hw, y, col)
			if hw > 0 {
				c.SetBG(cx+hw, y, shade)
			}
			if k%2 == 0 {
				c.SetBGQuad(cx-hw-1, y, col, ground(cx-hw-1, y), 0b0001)
				c.SetBGQuad(cx+hw+1, y, shade, ground(cx+hw+1, y), 0b0010)
			}
		}
		c.SetBG(cx, top+h, trunk)
	}
	willow := night(cube(95, 175, 95))
	copper := night(term.Lerp(cube(135, 95, 0), grey(8), 0.25))
	// Big crowns, few (the earlier scale he liked): round trees either
	// side, tall conifers right of the plaza, the willow left of it, the
	// copper tree by the bandstand, three shrubs.
	for k, tr := range [][3]int{{6, 12, 6}, {45, 12, 5}, {64, 12, 7}, {76, 12, 5}} {
		crown(tr[0], tr[1], tr[2], 3, greens[k%len(greens)])
		c.SetBG(tr[0], tr[1]+3, trunk)
	}
	conifer(51, 7, 7, night(cube(0, 95, 0)))
	conifer(55, 9, 5, night(cube(0, 135, 0)))
	crown(12, 11, 5, 3, willow)
	c.SetBG(12, 14, trunk)
	for _, g := range [][2]int{{9, 12}, {15, 12}, {10, 13}, {14, 13}, {12, 14}} {
		plot(c.Far(), g[0], g[1], ',', darken(willow, 0.3), 0.7)
	}
	crown(70, 10, 4, 2, copper)
	c.SetBG(70, 12, trunk)
	for _, sx := range []int{40, 60, 74} {
		crown(sx, 14, 2, 1, night(cube(95, 135, 0)))
	}
	// The lawn between the plaza and the terrace, under the rail: without it
	// the sky's dusk colour showed through as an orange band (the first
	// clip's close look). Then the path across it, and the lamps.
	path := lighten(night(cube(135, 95, 95)), 0.25)
	for x := 6; x <= 72; x++ {
		if ground(x, 14) != trunk {
			c.SetBG(x, 14, term.Lerp(ground(x, 14), path, 0.55))
		}
	}
	lamp := func(lx, ly int) {
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
						c.SetBG(x, y, term.Lerp(ground(x, y), warmLight, k))
					}
				}
			}
		}
	}
	lamp(16, 10)
	lamp(40, 11)
	lamp(66, 12)
	// The near trees, big and dark, framing the view at both edges, and a
	// smaller one left of the lake.
	crown(1, 13, 7, 3, nearG)
	crown(76, 13, 8, 3, nearG)
	for y := 8; y <= 16; y++ {
		for x := 0; x < c.W; x++ {
			bg := ground(x, y)
			isLit := bg == lighten(midG, 0.14) || bg == lighten(nearG, 0.14) || bg == lighten(greens[1], 0.14)
			isBody := bg == midG || bg == nearG || bg == greens[1]
			if isLit && scape.HashF(x, y, seed+88) < 0.16 {
				plot(c.Far(), x, y, '"', lighten(bg, 0.3), 0.6)
			} else if isBody && scape.HashF(x, y, seed+88) < 0.08 {
				plot(c.Far(), x, y, '.', darken(bg, 0.3), 0.6)
			}
		}
	}
	parkTerrace(c, l, tn, nkit)
}

// parkTerrace is the terrace: rail, balusters, floor, and the kittens.
func parkTerrace(c *canvas.Canvas, l float64, tn rainTones, nkit int) {
	floor := greyBetween(3, 8, l)
	for x := 0; x < c.W; x++ {
		plot(c.Mid(), x, 15, '=', tn.ironLit, 1)
		if x%2 == 0 {
			plot(c.Mid(), x, 16, '|', tn.ironDark, 1)
		}
	}
	fill(c, 0, 17, c.W-1, rainPaneBot, floor)
	fill(c, 0, 17, c.W-1, 17, lighten(floor, 0.15))
	if !drawKits(c) {
		xs := []int{24, 36, 48, 30, 42, 54, 22}
		for i := 0; i < nkit && i < len(xs); i++ {
			drawKittenAt(c.Near(), xs[i], rainPaneBot-3, i%2 == 1)
		}
	}
}

// paintHarbour: a working harbour across the water. The far quay under the
// skyline carries gantry cranes and stacks of containers, and a lighthouse
// stands at its end; the water is six rows deep so ships at three depths
// read at three sizes: a container ship and sailboats far, a ferry with
// its windows lit by night in the middle, a tug and a boat moored at our
// dock near. Chop on the water, wakes behind what moves, gulls by day, a
// buoy. The dock in front has its bollards, crates and a lamp, and is the
// stage. The water is STILL by design: the rain on the glass is the work.
func paintHarbour(c *canvas.Canvas, p scape.Palette, l float64, tn rainTones, seed int64, nkit int) {
	haze, far2 := hazeAt(p, l), hazeFar(p, l)
	lf := litFrac(l)
	far, mid, near := c.Far(), c.Mid(), c.Near()
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	// The far horizon: hills behind the city, a second row of towers in the
	// lighter haze, two tall stacks with their plume, a mast with a beacon,
	// then the near skyline.
	hills := term.Lerp(far2, p.SkyHorizon, 0.55)
	for x := 0; x < c.W; x++ {
		u := float64(x) / 80
		h := 4.6 - 1.4*math.Sin(u*2.6+0.9) - 0.7*math.Sin(u*7.3+1.7)
		fill(c, x, int(math.Round(h)), x, 8, hills)
	}
	for _, t := range [][3]int{{3, 4, 6}, {17, 3, 20}, {36, 4, 40}, {56, 3, 58}, {72, 4, 75}} {
		box(c, t[0], t[1], t[2], 8, far2)
	}
	plot(mid, 30, 2, '|', far2, 1)
	plot(mid, 30, 3, '|', far2, 1)
	fill(c, 30, 4, 30, 8, far2)
	fill(c, 33, 5, 33, 8, far2)
	for k, x := 0, 31; x <= 38; k, x = k+1, x+2 {
		plot(far, x, 1+k/3, ',', lighten(far2, 0.2), 0.45)
	}
	for x := 2; x <= 77; x++ {
		h := 5 + int(scape.HashF(x/3, 91, seed)*3)
		if x >= 10 && x <= 16 {
			h = 7
		}
		fill(c, x, h, x, 8, far2)
	}
	box(c, 22, 3, 25, 8, haze)
	box(c, 44, 2, 49, 8, haze)
	plot(mid, 46, 1, '|', haze, 1)
	plot(mid, 46, 0, '.', cube(175, 0, 0), 1)
	box(c, 63, 4, 66, 8, haze)
	for x := 3; x <= 76; x += 2 {
		if scape.HashF(x, 92, seed) < 0.35*lf {
			plot(far, x, 7, '.', warmLight, 0.6)
		}
	}
	// A tanker far out on the horizon line, tiny.
	fill(c, 52, 8, 59, 8, darken(far2, 0.25))
	fill(c, 58, 7, 58, 7, darken(far2, 0.25))
	// The quay: gantry cranes (a post, a boom, a tie, a hook) and the
	// containers under them; the lighthouse at the right end.
	iron := darken(haze, 0.45)
	crane := func(x, top int) {
		fill(c, x, top, x, 8, iron)
		fill(c, x-1, top, x+6, top, iron)
		line(mid, x+6, top+1, x+1, top+3, iron, 0.9)
		plot(mid, x+4, top+1, '|', iron, 0.9)
		plot(mid, x+4, top+2, '|', iron, 0.9)
	}
	crane(6, 4)
	crane(14, 5)
	crane(68, 4)
	// The containers' colours, well under the haze: a working port's
	// colour is there and does not shout (his note on the first version).
	boxes := []rgb{cube(95, 0, 0), cube(0, 95, 135), cube(135, 95, 0), cube(0, 95, 95)}
	for k, x := 0, 3; x <= 22; k, x = k+1, x+2 {
		if scape.HashF(x, 98, seed) < 0.75 {
			fill(c, x, 8, x+1, 8, term.Lerp(boxes[k%4], haze, 0.6))
		}
	}
	for k, x := 0, 66; x <= 76; k, x = k+1, x+2 {
		if scape.HashF(x, 98, seed) < 0.75 {
			fill(c, x, 8, x+1, 8, term.Lerp(boxes[(k+1)%4], haze, 0.6))
		}
	}
	fill(c, 78, 5, 78, 8, grey(20))
	fill(c, 78, 4, 78, 4, cube(175, 0, 0))
	// The water, still, six rows: the horizon's colour darkening toward the
	// dock; chop; the quay's lights laid on it by night.
	// The shore's own sea, tuned per hour on the cube (the first version
	// took the horizon's colour darkened, which is the skyline's colour, so
	// the water blended with what stood behind it).
	top := p.SeaFar
	vramp(c, 0, 9, c.W-1, 14, p.SeaFar, p.SeaNear)
	for y := 9; y <= 14; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed+99) < 0.07 {
				plot(far, x, y, '~', lighten(top, 0.18), 0.5)
			}
		}
	}
	for x := 3; x <= 76; x += 2 {
		if lf > 0.3 && scape.HashF(x, 92, seed) < 0.35*lf {
			for y := 9; y <= 11; y++ {
				if scape.HashF(x, y, seed+94) < 0.7 {
					plot(far, x, y, ':', warmLight, 0.5*lf)
				}
			}
		}
	}
	hull := greyBetween(1, 3, l)
	white := greyBetween(13, 19, l)
	wake := func(x0, x1, y int) {
		for x := x0; x <= x1; x++ {
			if scape.HashF(x, y, seed+95) < 0.7 {
				plot(far, x, y, '.', lighten(top, 0.3), 0.6)
			}
		}
	}
	// Motion, in the clip only: the ferry crosses once a loop, the tug works
	// its way left, the sailboats drift, the moored boat's mast rocks, the
	// gulls wheel, the lighthouse flashes every four seconds. All of it at a
	// constant pace whatever the agent does: the world's.
	drift := func(amp, w, ph float64) int {
		if !rainAnim {
			return 0
		}
		return int(math.Round(amp * math.Sin(rainT*w+ph)))
	}
	// Far: a container ship, long and low, its boxes and its bridge; two
	// sailboats.
	fill(c, 26, 10, 43, 10, hull)
	for k, x := 0, 28; x <= 37; k, x = k+1, x+2 {
		fill(c, x, 9, x+1, 9, term.Lerp(boxes[k%4], haze, 0.5))
	}
	fill(c, 40, 8, 42, 9, white)
	plot(mid, 41, 7, '|', hull, 1)
	wake(19, 25, 10)
	sail := func(x, y int) {
		fill(c, x, y, x+2, y, hull)
		plot(mid, x+1, y-2, '|', white, 1)
		c.SetBGQuad(x+1, y-1, white, ground(x+1, y-1), 0b1010)
		c.SetBGQuad(x+1, y-2, white, ground(x+1, y-2), 0b0010)
	}
	sail(6+drift(2, 0.4, 0), 11)
	sail(56+drift(2, 0.5, 2), 10)
	// Middle: the ferry, its windows lit by night, a buoy.
	fx := 48
	if rainAnim {
		fx = -14 + int(rainT/RainClipSecs*96)
	}
	fill(c, fx, 12, fx+13, 12, hull)
	fill(c, fx+1, 11, fx+12, 11, white)
	for x := fx + 2; x <= fx+11; x += 2 {
		col := darken(white, 0.35)
		if lf > 0.3 {
			col = warmLight
		}
		if x >= 0 && x < c.W {
			c.SetBG(x, 11, col)
		}
	}
	plot(mid, fx+7, 10, '|', hull, 1)
	plot(near, fx+13, 11, '.', cube(0, 135, 0), 0.8)
	wake(fx-7, fx-1, 12)
	plot(near, 40, 13, 'o', cube(175, 95, 95), 1)
	// Near: a tug, and the boat moored at our dock.
	tx := 64
	if rainAnim {
		tx = 74 - int(rainT/RainClipSecs*36)
	}
	fill(c, tx, 14, tx+9, 14, hull)
	fill(c, tx+2, 13, tx+5, 13, white)
	fill(c, tx+6, 13, tx+6, 13, hull)
	plot(near, tx+6, 12, '.', cube(175, 0, 0), 0.8)
	wake(tx+10, tx+15, 14)
	fill(c, 22, 14, 33, 14, hull)
	fill(c, 26, 13, 29, 13, white)
	mast := '|'
	if rainAnim && math.Sin(rainT*1.6) > 0.35 {
		mast = '/'
	}
	plot(mid, 24, 13, '|', hull, 1)
	plot(mid, 24, 12, mast, hull, 1)
	// Gulls by day, wheeling.
	if l > 0.4 {
		for k, g := range [][2]int{{30, 5}, {33, 4}, {58, 6}, {61, 5}} {
			plot(far, g[0]+drift(3, 0.7, float64(k)), g[1]+drift(1, 1.1, float64(k)*2), 'v', grey(18), 0.7)
		}
	}
	// The lighthouse flashes every four seconds by night.
	if lf > 0.3 && (!rainAnim || math.Mod(rainT, 4) < 0.5) {
		plot(near, 78, 4, '*', warmLight, 1)
		for dx := -3; dx <= 0; dx++ {
			for dy := -1; dy <= 1; dy++ {
				x, y := 78+dx, 4+dy
				c.SetBG(x, y, term.Lerp(ground(x, y), warmLight, 0.25*float64(4+dx)/4))
			}
		}
	}
	// The dock: its edge, the planks, bollards, crates, a lamp.
	deck := greyBetween(3, 7, l)
	fill(c, 0, 15, c.W-1, 15, lighten(deck, 0.25))
	fill(c, 0, 16, c.W-1, rainPaneBot, deck)
	for y := 16; y <= rainPaneBot; y++ {
		for x := (y * 3) % 8; x < c.W; x += 8 {
			plot(far, x, y, '|', darken(deck, 0.3), 0.6)
		}
	}
	for _, bx := range []int{8, 36} {
		fill(c, bx, 15, bx, 16, darken(deck, 0.5))
	}
	crate := term.Lerp(cube(135, 95, 95), deck, 0.3)
	fill(c, 44, 17, 47, 18, crate)
	fill(c, 45, 16, 46, 16, crate)
	for _, y := range []int{16, 17, 18} {
		plot(mid, 44, y, '|', darken(crate, 0.4), 0.8)
		plot(mid, 47, y, '|', darken(crate, 0.4), 0.8)
	}
	plot(near, 22, 17, 'o', darken(deck, 0.4), 1)
	plot(near, 23, 17, 'o', darken(deck, 0.4), 1)
	for y := 15; y <= 17; y++ {
		plot(mid, 54, y, '|', grey(9), 1)
	}
	if lf > 0.3 {
		plot(near, 54, 14, '*', warmLight, 0.9)
		for dy := -1; dy <= 3; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x, y := 54+dx, 14+dy
				k := 0.28 * lf * (1 - float64(dx*dx+dy*dy*2)/20)
				if k > 0 && x >= 0 && x < c.W {
					c.SetBG(x, y, term.Lerp(ground(x, y), warmLight, k))
				}
			}
		}
	}
	if !drawKits(c) {
		xs := []int{26, 36, 50, 30, 40, 58, 22}
		for i := 0; i < nkit && i < len(xs); i++ {
			drawKittenAt(near, xs[i], rainPaneBot-3, i%2 == 1)
		}
	}
}
