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

// The owl's cell, top-left of its 12x7 box: rows 12..18, on the mound.
const owlX, owlY = 64, 12

// The vista's rows, at 80x24. Sub-rows are twice these.
const (
	farBase   = 22 // sub-row the far range stands on (row 11)
	midBase   = 26 // the massif (row 13)
	nearBase  = 28 // the dark near ridge with its treeline (row 14)
	lakeTop   = 14 // rows 14..15
	meadowTop = 16 // rows 16..20
	bandTop   = 21 // the writing
)

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
// a notched taper from tip to base, in the near ridge's own colour.
func pineSilhouette(c *canvas.Canvas, cu, tip, base int, col term.RGB, seed int64) {
	width := func(v int) float64 {
		if v < tip {
			return -1
		}
		w := float64(v-tip) / 2.4
		w += (scape.HashF(v, 5, seed) - 0.5) * 1.6
		return math.Min(w, 7)
	}
	for y := tip / 2; y <= base/2 && y < c.H; y++ {
		for x := max(0, (cu-8)/2); x <= (cu+8)/2 && x < c.W; x++ {
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
func owlAt(c *canvas.Canvas, x, y int) {
	pick := OwlPick
	if pick < 0 || pick >= len(OwlAlts) {
		pick = 0
	}
	DrawOwl(c, &OwlAlts[pick], OwlCoat, x, y)
	if OwlPlace == 1 {
		DrawOwlet(c, OwlCoat, x-12, y+3, true)
		DrawOwlet(c, OwlCoat, x-6, y+3, false)
	} else {
		DrawOwlet(c, OwlCoat, x-14, meadowTop+1, true)
		DrawOwlet(c, OwlCoat, x-7, meadowTop+1, false)
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
	p := scape.PaletteAt(tod)
	l := lit(p)
	glow := glowAt(tod)
	warm := term.RGB{R: 255, G: 135, B: 95}
	t = loopPhase(t)
	frame := int(math.Round(t * 6))
	far, mid, near := c.Far(), c.Mid(), c.Near()

	// The sky, its stars, the moon or the sun.
	vramp(c, 0, 0, c.W-1, lakeTop-1, p.SkyTop, p.SkyHorizon)
	for y := 0; y < 10; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed) < 0.05*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.6*p.StarVis)
			}
		}
	}
	if p.MoonVis > 0.05 {
		for y := 1; y <= 2; y++ {
			for x := 12; x <= 14; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
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
		farTop[u] = skyline(u, farBase, 11, seed+11, 0.061, 40, 90)
		midTop[u] = skyline(u, midBase, 25, seed+23, 0.047, 84, 56)
		nearTop[u] = skyline(u, nearBase, 6, seed+31, 0.19, 60, 160)
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
		for x := 11; x <= 15; x++ {
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
		sh := 2*lakeTop + 2 + 1.6*(1+math.Sin(fu*0.06+1.1)) + (scape.HashF(u, 4, seed+63)-0.5)*0.9
		if fu > 112 {
			sh -= (fu - 112) / 6 // the headland
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
	cx, base := 30, meadowTop+3
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
			if OwlPlace == 0 && x >= 62 && y >= meadowTop+3 {
				continue // the owl's boulder
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
	if OwlPlace == 1 {
		paintBranch(c, pineCol, seed)
		owlAt(c, owlX, branchOwlY)
	} else {
		mound := make([]int, W2)
		for u := 0; u < W2; u++ {
			mound[u] = 999
			if d := float64(u-138) / 16; d > -1 && d < 1 {
				mound[u] = 2*bandTop - int(math.Round(5*math.Sqrt(1-d*d)))
				if math.Abs(d) < 0.78 {
					mound[u] = 2 * (bandTop - 3) // row 18 whole, the feet on it
				}
			}
		}
		paintRange(c, mound, greyBetween(5, 9, l), bandTop, nil)
		owlAt(c, owlX, owlY)
	}
	writeBand(c, bandTop, term.Lerp(grey(2), cube(95, 95, 0), l))
}
