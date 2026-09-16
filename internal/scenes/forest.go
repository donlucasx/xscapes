package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE THIRD SCAPE: a mountain vista, with the owl.
//
// His ask, 2026-09-15: the mountain scape in the woods with the owl. The
// first sketch answered the design question -- with no sea, what carries the
// work -- with three flat ridges and green triangles for trees, and his
// verdict was that the wind and the fire were right and the trees and the
// picture were not: "the scape reads pretty basic". His references (five
// photographs and one painting of ranges at dusk) share four things that
// sketch lacked, and this painter is built from them:
//
//   - DEPTH BY HAZE. Three ranges, each nearer one darker and each farther one
//     closer to the sky's own colour. The painting is the purest statement:
//     black foreground, dark blue, mid blue, pale blue, sky.
//   - A JAGGED SKYLINE, drawn at quarter-cell resolution through SetBGQuad,
//     so a summit is a shape and not a staircase. Snow sits in the gullies
//     below the peaks, not as a cap on one.
//   - LIGHT ON THE PEAKS. At dawn and dusk the upper faces take the horizon's
//     warm colour while the shadow stays cool; at noon the rock is plain.
//   - A VALLEY IN FRONT. A lake taking the sky's colour, a meadow of scrub,
//     and the foreground as silhouette: two black pines at the left edge,
//     framing the view rather than being it.
//
// The work has two candidate channels, both his, drawn as two scenes on the
// one vista so only the channel differs:
//
//   - THE WIND. The fire is fixed and is the night's light; the wind is the
//     work, seen in the smoke's lean, the flames' lean, the fraction of the
//     scrub that is bent flat, and what is in the air. Coverage and position.
//   - THE FIRE. Still air; the fire's height, its sparks and the thickness of
//     its smoke are the work. Position and count, and the fire is the light
//     as well, which is the cost.
//
// Every painter closes on itself after LoopSecs.
var Forest = []Scene{
	{"The vista: the wind is the work",
		[6]string{"the moon and the fire by night, the sun and the peaks' glow by day", "the real sky over three ranges", "the wind: how far the smoke and the flames lean, how much of the scrub is flat, what is in the air", "rock, snow, the lake, the meadow", "the meadow", "the owl on a boulder, right"},
		"The campfire is fixed and lights the meadow at night; the work is the wind through everything else. Still air is a straight column of smoke and upright grass; a gale streams the smoke sideways, flattens the flames and the scrub, and fills the air with what it has picked up. Coverage and position, never rate, so a screenshot says how hard.",
		func(c *canvas.Canvas, tod, t, level float64, seed int64) { paintVista(c, tod, t, level, seed, false) }},
	{"The vista: the fire is the work",
		[6]string{"the fire, and the sky", "the real sky over three ranges", "the fire's height, its sparks, the thickness of its smoke", "rock, snow, the lake, the meadow", "the meadow", "the owl on a boulder, right"},
		"Still air. The campfire is the work: an ember at rest, a blaze at full stretch, more sparks and thicker smoke as it climbs. The most legible read of the two, and it costs one thing: the fire is also the light, so a busy agent brightens the meadow it sits in.",
		func(c *canvas.Canvas, tod, t, level float64, seed int64) { paintVista(c, tod, t, level, seed, true) }},
}

// LoopSecs is the loop every forest painter closes on.
const LoopSecs = 4.0

// The owl's cell, top-left of its 12x7 box: rows 12..18, on the mound. The
// study's values; the live scape places it where the composition puts the
// companion (vista.go).
const owlX, owlY = 64, 12

// vistaLayout is where the vista's parts sit for one frame size. The study
// and the page clips use vista80, the 80x24 the scene was drawn at; the live
// scape derives one from the window (vistaLayoutFor). Every row that was a
// constant is here, so a wider or taller frame moves the parts and never
// the drawing.
type vistaLayout struct {
	W, H int
	// Sub-rows (twice the row) the three ranges stand on.
	farBase, midBase, nearBase int
	// Rows: the lake's first row, the meadow's first row, the writing's.
	lakeTop, meadowTop, bandTop int
	// The owl's box, top-left; the fire's column; the moon's column.
	owlX, owlY, fireX, moonX int
	// Where the treeline dips under the owl's head is 2*owlY (whole cells
	// behind the eyes); it is a sub-row here so a taller frame can lower it.
}

// vista80 is the layout the vista was drawn at. The numbers ARE the old
// constants: farBase 22 is row 11, the massif on row 13, the near ridge on
// row 14, the lake rows 14..15, the meadow 16..20, the writing from 21.
func vista80() vistaLayout {
	return vistaLayout{W: 80, H: 24,
		farBase: 22, midBase: 26, nearBase: 28,
		lakeTop: 14, meadowTop: 16, bandTop: 21,
		owlX: owlX, owlY: owlY, fireX: 30, moonX: 12}
}

// ridged is a skyline: layered sines folded at zero so every crossing is a
// summit, phases from the seed, sharpened so the valleys stay low. 0..1.
func ridged(u float64, seed int64, f0 float64) float64 {
	v, amp, freq, norm := 0.0, 1.0, f0, 0.0
	for o := int64(0); o < 4; o++ {
		ph := scape.HashF(int(o), 7, seed) * 2 * math.Pi
		v += amp * (1 - math.Abs(math.Sin(u*freq+ph)))
		norm += amp
		amp *= 0.55
		freq *= 2.13
	}
	return math.Pow(v/norm, 1.25)
}

// skyline is one range's top edge in sub-rows at sub-column u: base minus a
// height, with an envelope so the massif has a middle and the flanks fall.
func skyline(u int, base int, maxH float64, seed int64, f0 float64, centre, width float64) int {
	fu := float64(u)
	env := 0.35 + 0.65*math.Exp(-math.Pow((fu-centre)/width, 2))
	h := maxH * env * ridged(fu, seed, f0)
	h += (scape.HashF(u, 3, seed) - 0.5) * 0.8 // roughness
	return base - int(math.Round(h))
}

// paintRange fills every cell at or below the skyline with the range's
// colour, and the cells the skyline crosses as quarters, so the edge is
// drawn at half a cell in both directions. tint, when given, colours each
// sub-row by its depth below the local top -- snow and the peaks' glow.
func paintRange(c *canvas.Canvas, top []int, col term.RGB, until int, tint func(u, v, depth int) (term.RGB, bool)) {
	for x := 0; x < c.W; x++ {
		uL, uR := 2*x, 2*x+1
		tL, tR := top[uL], top[uR]
		first := min(tL, tR) / 2
		for y := max(0, first); y < until && y < c.H; y++ {
			var mask uint8
			shape := col
			// The tint of the cell is the tint of its shallowest quarter,
			// so a snow patch keeps its edge.
			tinted := false
			for _, q := range []struct {
				u, v int
				bit  uint8
			}{{uL, 2 * y, 8}, {uR, 2 * y, 4}, {uL, 2*y + 1, 2}, {uR, 2*y + 1, 1}} {
				t := top[q.u]
				if q.v >= t {
					mask |= q.bit
					if tint != nil && !tinted {
						if tc, ok := tint(q.u, q.v, q.v-t); ok {
							shape, tinted = tc, true
						}
					}
				}
			}
			if mask == 0 {
				continue
			}
			if mask == 0b1111 {
				c.SetBG(x, y, shape)
				continue
			}
			c.SetBGQuad(x, y, shape, c.BGAt(x, y), mask)
		}
	}
}

// pineSilhouette draws a conifer as a dark shape at quarter-cell resolution:
// a taper from tip to base in tiers, each bough wider at its bottom edge, in
// the near ridge's own colour. It widens all the way down -- the first
// version capped the width and the tree became a tower below the cap, with
// a wall one cell inside the frame that read as cut ("any reason why the
// trees cut abruptly on the left edge?", 2026-09-15). The frame's edge is
// what cuts a tree, never a wall of its own.
func pineSilhouette(c *canvas.Canvas, cu, tip, base int, col term.RGB, seed int64) {
	width := func(v int) float64 {
		if v < tip {
			return -1
		}
		d := float64(v - tip)
		tier := math.Mod(d, 5) / 5 // 0 at a bough's top, 1 at its skirt
		w := d / 2.6 * (0.7 + 0.3*tier)
		w += (scape.HashF(v, 5, seed) - 0.5) * 1.2
		return w
	}
	for y := tip / 2; y <= base/2 && y < c.H; y++ {
		for x := max(0, (cu-14)/2); x <= (cu+14)/2 && x < c.W; x++ {
			var mask uint8
			for _, q := range []struct {
				u, v int
				bit  uint8
			}{{2 * x, 2 * y, 8}, {2*x + 1, 2 * y, 4}, {2 * x, 2*y + 1, 2}, {2*x + 1, 2*y + 1, 1}} {
				if q.v > base {
					continue
				}
				if w := width(q.v); w >= 0 && math.Abs(float64(q.u-cu)) <= w {
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
}

// owlAt draws the picked owl (owl.go) facing left, with its litter: on the
// branch the owlets sit along the limb beside it; on the mound they sit on
// the meadow to its left, where the crablets sit on the sand.
func owlAt(c *canvas.Canvas, x, y, grassY int) {
	pick := OwlPick
	if pick < 0 || pick >= len(OwlAlts) {
		pick = 0
	}
	DrawOwl(c, &OwlAlts[pick], OwlCoat, x, y)
	for k := 1; k <= OwletCount; k++ {
		switch OwlPlace {
		case 1, 2:
			DrawOwlet(c, OwlCoat, x-6*k, y+3, k%2 == 1) // along the limb
		case 4:
			DrawOwlet(c, OwlCoat, x-2-6*k, 14, k%2 == 1) // standing on the top rail
		default:
			DrawOwlet(c, OwlCoat, x-7*k, grassY, k%2 == 1) // on the grass
		}
	}
}

// loopPhase is t folded into the loop, so every rate below closes on itself.
func loopPhase(t float64) float64 { return math.Mod(t, LoopSecs) }

// glowAt is how much the peaks are lit by a low sun: full at dusk and dawn,
// nothing at night or noon.
func glowAt(tod float64) float64 {
	g := math.Max(0, 1-math.Abs(tod-0.78)/0.09) + math.Max(0, 1-math.Abs(tod-0.24)/0.09)
	return math.Min(1, g)
}

func paintVista(c *canvas.Canvas, tod, t, level float64, seed int64, fireIsWork bool) {
	paintVistaL(c, vista80(), tod, t, level, seed, fireIsWork, nil)
}

// vistaLive is what the live scape adds to the study painter: the parts that
// come from the agent rather than from the clock and the work level. Nil is
// the study: the moon high, no checklist, the owl and the sample writing
// drawn by the painter itself.
type vistaLive struct {
	// ContextUsed sets the moon's altitude, as it does on the shore.
	ContextUsed float64
	// TodoDone lights that many stars in the sky.
	TodoDone int
	// The live scene draws its own owl (with a pose) and its own writing
	// (the real activity tail), so the painter leaves both slots bare.
	SkipOwl, SkipBand bool
}

// moonRow is where the moon sits for a context: row 1 fresh, sinking toward
// the far range as the window fills, never into it.
func (lay vistaLayout) moonRow(ctxUsed float64) int {
	lo, hi := 1, lay.farBase/2-4
	if hi <= lo {
		return lo
	}
	return lo + int(math.Round(ctxUsed*float64(hi-lo)))
}

func paintVistaL(c *canvas.Canvas, lay vistaLayout, tod, t, level float64, seed int64, fireIsWork bool, live *vistaLive) {
	farBase, midBase, nearBase := lay.farBase, lay.midBase, lay.nearBase
	lakeTop, meadowTop, bandTop := lay.lakeTop, lay.meadowTop, lay.bandTop
	owlX, owlY := lay.owlX, lay.owlY
	p := scape.PaletteAt(tod)
	l := lit(p)
	glow := glowAt(tod)
	warm := term.RGB{R: 255, G: 135, B: 95}
	t = loopPhase(t)
	frame := int(math.Round(t * 6))
	far, mid, near := c.Far(), c.Mid(), c.Near()
	// The ranges' skylines were placed in sub-columns at 80 wide; a wider
	// frame stretches them rather than repeating them.
	sx := float64(c.W) / 80

	// The sky, its stars, the moon or the sun.
	vramp(c, 0, 0, c.W-1, lakeTop-1, p.SkyTop, p.SkyHorizon)
	for y := 0; y < lakeTop-4; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed) < 0.05*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.6*p.StarVis)
			}
		}
	}
	moonY := 1
	if live != nil {
		moonY = lay.moonRow(live.ContextUsed)
	}
	if p.MoonVis > 0.05 {
		for y := moonY; y <= moonY+1; y++ {
			for x := lay.moonX; x <= lay.moonX+2; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
	}
	if live != nil && live.TodoDone > 0 {
		vistaStars(c, lay, p, seed, live.TodoDone)
	}

	// Three ranges. Rock is grey at night and warm stone by day; each farther
	// range is pulled toward the sky, which is what distance does.
	rock := term.Lerp(grey(3), term.RGB{R: 150, G: 138, B: 118}, l)
	farCol := term.Lerp(rock, p.SkyHorizon, 0.58)
	midCol := term.Lerp(rock, p.SkyHorizon, 0.22)
	nearCol := term.Lerp(grey(1), term.RGB{R: 34, G: 62, B: 36}, l)
	W2 := c.W * 2
	farTop := make([]int, W2)
	midTop := make([]int, W2)
	nearTop := make([]int, W2)
	for u := 0; u < W2; u++ {
		farTop[u] = skyline(u, farBase, 11, seed+11, 0.061, 40*sx, 90*sx)
		midTop[u] = skyline(u, midBase, 25, seed+23, 0.047, 84*sx, 56*sx)
		nearTop[u] = skyline(u, nearBase, 6, seed+31, 0.19, 60*sx, 160*sx)
	}
	paintRange(c, farTop, term.Lerp(farCol, warm, 0.22*glow), c.H, nil)
	snow := term.Lerp(term.Lerp(grey(15), grey(23), l), warm, 0.45*glow)
	midLit := term.Lerp(midCol, warm, 0.6*glow)
	paintRange(c, midTop, midCol, c.H, func(u, v, depth int) (term.RGB, bool) {
		// Snow in the upper faces and down the gullies, patchy, only where
		// the massif stands high enough to hold it.
		high := midBase - midTop[u]
		if high >= 7 {
			if depth <= 2 && scape.HashF(u/2, v, seed+41) < 0.75 {
				return snow, true
			}
			if depth <= 7 && scape.HashF(u/3, 9, seed+42) < 0.18 {
				return snow, true
			}
		}
		// The upper faces take the low sun.
		if depth <= 6 && glow > 0 {
			return midLit, true
		}
		return term.RGB{}, false
	})
	// The near ridge, with its treeline: a second skyline on top, at a
	// frequency of about one tree every two cells, so the tips are tips.
	for u := 0; u < W2; u++ {
		nearTop[u] -= int(math.Round(3.0 * ridged(float64(u), seed+53, 1.7)))
		// Whole cells behind the owl's head: the treeline dips to a flat
		// top under it, or its tufts are lost to the trees' quarter-cells.
		if OwlPlace == 0 && u >= 2*owlX-2 && u < 2*(owlX+12)+2 {
			nearTop[u] = 2 * (owlY)
		}
	}
	paintRange(c, nearTop, nearCol, c.H, nil)

	// The lake reflects the whole sky, darker, with the moon's light on it
	// by night.
	lake := term.Lerp(term.Lerp(p.SkyHorizon, p.SkyTop, 0.5), grey(0), 0.3)
	fill(c, 0, lakeTop, c.W-1, meadowTop-1, lake)
	if p.MoonVis > 0.05 && p.StarVis > 0.3 {
		for x := lay.moonX - 1; x <= lay.moonX+3; x++ {
			c.SetBG(x, lakeTop, term.Lerp(lake, p.Moon, 0.3*p.MoonVis*p.StarVis))
		}
	}
	// The meadow: olive by day, dark by night, darker toward us, and it
	// goes dim before the sky does.
	lm := math.Pow(l, 1.6)
	mTop := term.Lerp(grey(2), cube(95, 135, 0), lm)
	mBot := term.Lerp(grey(1), cube(95, 95, 0), lm)
	vramp(c, 0, meadowTop, c.W-1, bandTop-1, mTop, mBot)
	// The near shore wanders, and rises into a headland under the owl so
	// the lake is a body of water in the valley and not a stripe across it.
	shore := make([]int, W2)
	for u := 0; u < W2; u++ {
		fu := float64(u)
		sh := float64(2*lakeTop+2) + 1.6*(1+math.Sin(fu*0.06+1.1)) + (scape.HashF(u, 4, seed+63)-0.5)*0.9
		if head := float64(2*owlX - 16); fu > head {
			sh -= (fu - head) / 6 // the headland, rising under the owl
		}
		// ⚠ Whole cells under the owl. A plain glyph (Plot, not PlotOn) over
		// a cell whose background is QUARTERS loses to the quarters, by the
		// canvas's own rule that a star must not break the disc. The owl's
		// eyes are plain glyphs, and the eye row lost them when the shore's
		// edge ran under it. Seen in a screenshot, 2026-09-15.
		if u >= 2*owlX-4 {
			sh = float64(2 * lakeTop)
		}
		if fu < 8 {
			sh -= (8 - fu) / 3
		}
		shore[u] = int(math.Round(math.Max(float64(2*lakeTop), sh)))
	}
	paintRange(c, shore, mTop, meadowTop, nil)
	// Two pines at the left edge, in silhouette.
	pineCol := term.Lerp(nearCol, grey(0), 0.35)
	pineSilhouette(c, 9, 13, 41, pineCol, seed+61)
	pineSilhouette(c, 23, 19, 41, pineCol, seed+62)

	// The fire, on the meadow.
	cx, base := lay.fireX, meadowTop+3
	H, lean := 3, 0.0
	if fireIsWork {
		H = 2 + int(math.Round(level*4))
	} else {
		lean = level * 2.6
	}
	R := 7.0
	if fireIsWork {
		R = 4.0 + 9.0*level
	}
	// Firelight on the meadow by night.
	for y := lakeTop; y < bandTop; y++ {
		for x := 0; x < c.W; x++ {
			d := math.Hypot(float64(x-cx)/2, float64(y-base))
			if d < R {
				a := (1 - d/R) * (0.5*math.Pow(1-l, 2) + 0.04)
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), cube(135, 95, 0), a))
			}
		}
	}
	for x := cx - 2; x <= cx+2; x++ {
		plot(near, x, base+1, '=', cube(95, 95, 0), 1)
	}
	tipRow := base - (H - 1)
	for r := 0; r < H; r++ {
		y := base - r
		w := (H - r) / 2
		dx := 0
		if H > 1 {
			dx = int(math.Round(lean * float64(r) / float64(H-1)))
		}
		for x := cx - w; x <= cx+w; x++ {
			h := scape.HashF(x, y*31+frame%24, seed+71)
			if h < 0.2 {
				continue
			}
			g := '^'
			if h < 0.5 {
				g = '*'
			} else if h < 0.62 {
				g = ')'
			}
			col := cube(255, 95, 0)
			if r > H/2 {
				col = cube(255, 175, 0)
			}
			if r >= H-1 {
				col = cube(255, 215, 135)
			}
			plot(near, x+dx, y, g, col, 1)
		}
	}
	// Smoke: a column that leans with the wind, or thickens with the fire.
	smokeLean, smokeDensity := 0.25, 0.55
	if fireIsWork {
		smokeDensity = 0.2 + 0.6*level
	} else {
		smokeLean = 0.2 + 1.3*level
	}
	smokeCol := term.Lerp(grey(15), grey(19), l)
	for r := 1; r <= 10; r++ {
		y := tipRow - r
		if y < 1 {
			break
		}
		x := cx + int(math.Round(lean)) + int(math.Round(smokeLean*float64(r)))
		for k := -1; k <= 1; k++ {
			h := scape.HashF(x+k, (r+frame)%12, seed+81)
			if h < smokeDensity*(1-0.06*float64(r)) {
				g := '.'
				if h < smokeDensity*0.4 {
					g = ':'
				} else if h < smokeDensity*0.7 {
					g = '~'
				}
				plot(mid, x+k, y, g, smokeCol, 0.85)
			}
		}
	}
	if fireIsWork {
		// Sparks rise one row a frame through a 24-row pattern, so they close.
		spread := 1.0 + 5.0*level
		for y := 3; y < tipRow; y++ {
			for x := cx - 6; x <= cx+6; x++ {
				if math.Abs(float64(x-cx)) > spread {
					continue
				}
				if scape.HashF(x, ((y+frame)%24+24)%24, seed+62) < 0.004+0.04*level {
					plot(mid, x, y, '.', cube(255, 215, 135), 0.9)
				}
			}
		}
	}

	// The scrub. Upright is a tuft; bent is a stroke. In the wind the
	// FRACTION bent is the work, with a flutter at the margin so the meadow
	// moves without the fraction changing.
	scrub := term.Lerp(grey(7), cube(135, 175, 0), lm)
	for y := meadowTop; y < bandTop; y++ {
		for x := 0; x < c.W; x++ {
			if OwlPlace == 0 && x >= owlX-2 && y >= bandTop-2 {
				continue // the owl's boulder (its flat top keeps its scrub)
			}
			h := scape.HashF(x, y, seed+90)
			if h > 0.28 {
				continue
			}
			g := '"'
			if !fireIsWork {
				ph := scape.HashF(x, y, seed+91) * 2 * math.Pi
				bent := scape.HashF(x, y, seed+92)+0.18*math.Sin(2*math.Pi*t/2+ph) < level*0.95
				if bent {
					g = ','
					if scape.HashF(x, y, seed+93) < 0.5 {
						g = '_'
					}
				}
			}
			plot(mid, x, y, g, scrub, 0.85)
		}
	}
	// What the wind carries, blown right at twenty columns a second through
	// a pattern the frame's width, so the loop closes.
	if !fireIsWork {
		d := 0.004 + 0.05*level
		shift := int(math.Round(t * 20))
		leaf := term.Lerp(grey(10), cube(135, 175, 0), l)
		for y := 4; y < bandTop; y++ {
			for x := 0; x < c.W; x++ {
				h := scape.HashF(((x-shift)%c.W+c.W)%c.W, y, seed+50)
				if h < d {
					g := '\''
					if h < d*0.4 {
						g = ','
					} else if h < d*0.7 {
						g = '-'
					}
					plot(mid, x, y, g, leaf, 0.7)
				}
			}
		}
	}

	// The owl, right: on a branch from the right edge, or on a low mound
	// drawn as quarters with a flat top under its feet.
	switch OwlPlace {
	case 1:
		paintBranch(c, pineCol, seed, branchTop)
		owlAt(c, owlX, branchOwlY, meadowTop+1)
	case 2:
		// A lower limb, its perch on row 13, so it crosses the lake and
		// reads against the water.
		paintBranch(c, pineCol, seed, 26)
		owlAt(c, owlX, 6, meadowTop+1)
	case 3:
		// A stump in the meadow, the owl's own width: the pale cut face on
		// top with the owl standing on it, so the face shows around its
		// feet, and bark below in a wood lighter than the grass. His note
		// on the post: "does not read great" -- it was two cells of the
		// treeline's colour on a meadow nearly as dark.
		bark := greyBetween(6, 11, l)
		fill(c, owlX, bandTop-2, owlX+11, bandTop-1, bark)
		for _, bx := range []int{owlX + 2, owlX + 6, owlX + 9} {
			plot(near, bx, bandTop-2, '|', greyBetween(4, 8, l), 1)
			plot(near, bx+1, bandTop-1, '|', greyBetween(4, 8, l), 1)
		}
		fill(c, owlX, bandTop-3, owlX+11, bandTop-3, term.RGB{R: 215, G: 175, B: 135})
		owlAt(c, owlX, owlY, meadowTop+1)
	case 4:
		// A fence from the right edge in weathered grey: two posts and two
		// rails, the owl on the end post, the owlets standing on the top
		// rail.
		fence := greyBetween(9, 13, l)
		for _, px := range []int{owlX + 4, owlX + 13} {
			fill(c, px, meadowTop, px+1, bandTop-1, fence)
		}
		fill(c, owlX-14, meadowTop+1, c.W-1, meadowTop+1, fence)
		fill(c, owlX-14, meadowTop+3, c.W-1, meadowTop+3, fence)
		owlAt(c, owlX, 10, meadowTop+1)
	default:
		mound := make([]int, W2)
		for u := 0; u < W2; u++ {
			mound[u] = 999
			if d := float64(u-(2*owlX+10)) / 16; d > -1 && d < 1 {
				mound[u] = 2*bandTop - int(math.Round(5*math.Sqrt(1-d*d)))
				if math.Abs(d) < 0.78 {
					mound[u] = 2 * (bandTop - 3) // the flat top, the feet on it
				}
			}
		}
		paintRange(c, mound, greyBetween(5, 9, l), bandTop, nil)
		if live == nil || !live.SkipOwl {
			owlAt(c, owlX, owlY, meadowTop+1)
		}
	}
	if live == nil || !live.SkipBand {
		writeBand(c, bandTop, term.Lerp(grey(2), cube(95, 95, 0), l))
	} else {
		fill(c, 0, bandTop, c.W-1, c.H-1, lay.bandColor(l))
	}
}

// bandColor is the writing's ground: the meadow's dark end.
func (lay vistaLayout) bandColor(l float64) term.RGB {
	return term.Lerp(grey(2), cube(95, 95, 0), l)
}

// vistaStars is the checklist in the vista's sky: one star per finished
// todo, at a place fixed by its index and the seed so a star lights where it
// always was, across the sky on a golden-ratio sequence (what the shore
// learned in s27: index order piles the first few into one corner). The ink
// is lifted toward white until it clears the sky by the shore's own bar, so
// the count survives the day.
func vistaStars(c *canvas.Canvas, lay vistaLayout, p scape.Palette, seed int64, n int) {
	rows := lay.lakeTop - 6
	if rows < 2 || c.W < 8 {
		return
	}
	near := c.Near()
	const phi = 0.6180339887
	for i := 0; i < n && i < 64; i++ {
		fx := math.Mod(float64(i)*phi+scape.HashF(i, 7, seed), 1)
		x := 2 + int(fx*float64(c.W-4))
		y := 1 + int(scape.HashF(i, 9, seed)*float64(rows-1))
		// Keep off the moon's column.
		if x >= lay.moonX-1 && x <= lay.moonX+3 {
			x = (x + 5) % (c.W - 2)
		}
		ink := vistaStarInk(c, x, y, p.Star)
		near.Plot(x, y, '*', ink, 1)
	}
}

// vistaStarInk lifts a star's ink toward white in eight steps until it reads
// 55 luma over its ground, the shore's todoStarContrast; the last rung wins
// if none does.
func vistaStarInk(c *canvas.Canvas, x, y int, star term.RGB) term.RGB {
	ground := c.BGAt(x, y)
	white := term.RGB{R: 255, G: 255, B: 255}
	ink := star
	for step := 0; step <= 8; step++ {
		ink = term.Lerp(star, white, float64(step)/8)
		seen := term.Profile256.Quantise(ink, true)
		if luma(seen)-luma(term.Profile256.Quantise(ground, false)) >= 55 {
			return ink
		}
	}
	return ink
}
