package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE RAINY WINDOW, DEVELOPED. His ask of 2026-09-23, before the scape is
// built: "develop the scene's art further a bit. Add more detail to the
// skyline. Could there be a closer rooftop or a different type of place
// where sub agents could roam". Three rounds of notes the same afternoon:
// the windows "not working", the scene "busy", the litter over lit
// windows; "inspire yourself on NYC skylines"; the tanks "stand out (not in
// a good way), they should blend in"; the fire escape "needs more details,
// on its way tho"; "a more polished skyline, so try a few more
// alternatives".
//
// So: New York, in three depths (each farther thing closer to the sky's
// colour, the vista's rule), the far skyline AUTHORED rather than rolled,
// three ways (RainSkyline):
//
//	0  MIDTOWN: setback towers, one Empire-like with a lit crown, one with
//	   a tapered crown, two slabs; the block low under them.
//	1  THE ROW: a low, dense far silhouette with one modest tower; the
//	   block tall and near, brownstone walk-ups with fire escapes on two of
//	   the facades.
//	2  THE BRIDGE: a suspension tower's twin arches with the cables
//	   sagging away to both sides and their suspenders; a few towers
//	   behind; the block low.
//
// The block in all three: walk-ups ten to fourteen cells wide, brick, rosy
// brick or stone by day and greys by night, a cornice on the roofline and a
// shade column on the right edge for volume, windows TWO cells wide four
// apart, one a storey, dark glass by day; by night a third of the upper
// storeys lit, in apartment pairs, and NOTHING lit at or below the stage's
// top row, so the litter never fronts a lit window. Tanks in their own
// building's tone a step darker, so they belong to the roof they stand on.
//
// The stage (RainStage, and the three Scenes): the sill indoors; a lower
// roof across the alley; a fire escape landing outside the glass. Then the
// glass (the rain, the work), the frame, the sill and the cat. Study
// painters: nothing here is wired to the product.
var RainStudy = []Scene{
	{"Rain A: the block, the family on the sill",
		[6]string{"the sky and the city's lights", "the real sky over the block", "rain on the glass", "glass, a wooden frame, the sill", "the desk under the sill", "the cat on the sill, the kittens beside it"},
		"No new stage: the kittens sit indoors on the sill with the cat. Everything outside is the world.",
		func(c *canvas.Canvas, tod, t, level float64, seed int64) {
			paintRainStudy(c, tod, t, level, seed, rainStageSill, 3)
		}},
	{"Rain B: the roof below",
		[6]string{"the sky and the city's lights", "the real sky over the block", "rain on the glass", "glass, the frame, the sill; tar", "the desk under the sill", "the cat on the sill; the kittens on the roof across the alley"},
		"The window looks down on a lower roof: a tank, a bulkhead with its door lit, a skylight, vents, puddles. The kittens roam the roof.",
		func(c *canvas.Canvas, tod, t, level float64, seed int64) {
			paintRainStudy(c, tod, t, level, seed, rainStageRoof, 3)
		}},
	{"Rain C: the fire escape",
		[6]string{"the sky and the city's lights", "the real sky over the block", "rain on the glass", "glass, the frame, the sill; iron", "the desk under the sill", "the cat on the sill; the kittens on the landing outside"},
		"A fire escape landing right outside the glass, at the cat's own scale: stringers, the rail behind the kittens, the grating under them, the stair down at the end.",
		func(c *canvas.Canvas, tod, t, level float64, seed int64) {
			paintRainStudy(c, tod, t, level, seed, rainStageEscape, 3)
		}},
}

type rainStage int

const (
	rainStageSill rainStage = iota
	rainStageRoof
	rainStageEscape
)

// RainKittens overrides the kitten count the study draws when non-negative.
var RainKittens = -1

// RainSkyline picks the far skyline: 0 midtown, 1 the row, 2 the bridge,
// 3 anycity (no landmark: hills in haze, glass slabs, a mast, a crane).
var RainSkyline = 0

// RainWindows is what the block's lit windows say: 0 the clock (a share of
// the upper storeys lit by night, the world's); 1 the session's SPEND, with
// RainSpend 0..1 the share of the city that has come on (a level that only
// grows, the pond idea of s33, drawn here for his look).
var (
	RainWindows = 0
	RainSpend   = 0.0
)

// RainTodoLit is the number of finished todos lit as floors on the
// landmark, from the bottom up, when a skyline has one (midtown's Empire);
// zero draws none. A candidate for the accumulator slot, for his look.
var RainTodoLit = 0

// The window at 80x24: the pane is rows 0..19 inside a frame two cells
// wide, a centre mullion at 39..40 and the sashes' meeting rail on row 6.
// The sill is row 20, the writing from 21.
const (
	rainPaneBot = 19
	rainSill    = 20
	rainBand    = 21
	rainCatX    = 64
	rainCatY    = 13
	rainRail    = 6
)

type building struct {
	x0, x1, top int
	day         rgb // brick, rosy brick or stone by day
	tone        int // night grey off the block's
	roof        int // 0 none, 1 water tank, 2 bulkhead, 3 chimney
	escape      bool
}

// The block's daytime walls: brick, rosy brick and stone, the cube's own.
// No tan: its dusk blend quantised to mustard (135,95,0), the warm-dark trap
// (under 95 in blue the cube holds only zero). Brick's blend goes to the
// grey ramp and rosy brick's to brick, seen on the cube.
var dayWalls = []rgb{cube(135, 95, 95), grey(9), cube(175, 135, 135), grey(8), cube(135, 95, 95), grey(10)}

// walkups lays contiguous buildings across x0..x1 with tops in
// minTop..maxTop, neighbours never the same height.
func walkups(x0, x1, minTop, maxTop, minW, maxW int, seed int64) []building {
	var bs []building
	x, k := x0, 0
	prevTop := -1
	for x <= x1 {
		w := minW + int(scape.HashF(k, 41, seed)*float64(maxW-minW+1))
		if x1+1-x-w < minW {
			w = x1 + 1 - x
		}
		top := minTop + int(scape.HashF(k, 42, seed)*float64(maxTop-minTop+1))
		if top == prevTop {
			top = minTop + (top-minTop+1+int(scape.HashF(k, 43, seed)*float64(maxTop-minTop)))%(maxTop-minTop+1)
		}
		b := building{x0: x, x1: x + w - 1, top: top,
			day:  dayWalls[int(scape.HashF(k, 44, seed)*float64(len(dayWalls)))],
			tone: int(scape.HashF(k, 47, seed)*3) - 1,
			roof: int(scape.HashF(k, 46, seed) * 5),
		}
		if b.roof > 3 {
			b.roof = 0
		}
		bs = append(bs, b)
		prevTop = top
		x += w
		k++
	}
	return bs
}

func clampGrey(k int) rgb {
	if k < 0 {
		k = 0
	}
	if k > 23 {
		k = 23
	}
	return grey(k)
}

// rainTones are the city's greys for the hour. Night first, day second, as
// greyBetween reads them.
type rainTones struct {
	tar, ironDark, ironLit, frame, sill, desk rgb
	l                                         float64
}

func rainTonesAt(l float64) rainTones {
	return rainTones{
		tar:      greyBetween(2, 6, l),
		ironDark: greyBetween(1, 3, l),
		ironLit:  greyBetween(9, 6, l),
		frame:    greyBetween(11, 19, l),
		sill:     greyBetween(7, 11, l),
		desk:     greyBetween(4, 8, l),
		l:        l,
	}
}

// wallOf is a building's wall for the hour: its own grey by night, its
// brick or stone by day, blended between.
func wallOf(b building, l float64) rgb {
	return term.Lerp(clampGrey(3+b.tone), b.day, l)
}

// litFrac is how much of the city is lit: nothing at noon, most by night.
func litFrac(l float64) float64 { return math.Max(0, math.Min(1, (0.6-l)/0.5)) }

// darken pulls a colour toward black; lighten toward white.
func darken(c rgb, k float64) rgb  { return term.Lerp(c, rgb{}, k) }
func lighten(c rgb, k float64) rgb { return term.Lerp(c, grey(23), k) }

// hazeAt is the far layer's colour: a grey lifted toward the sky's own
// colour, lighter than the block by night because a city's night sky glows.
// hazeFar is the layer behind that, lifted further.
func hazeAt(p scape.Palette, l float64) rgb {
	if l < 0.3 {
		return term.Lerp(grey(7), p.SkyHorizon, 0.35)
	}
	return term.Lerp(grey(5), p.SkyHorizon, 0.35)
}

func hazeFar(p scape.Palette, l float64) rgb { return term.Lerp(hazeAt(p, l), p.SkyHorizon, 0.35) }

// farBot is the last row the far layer paints: the block or the stage
// covers everything below it.
const farBot = 14

// box paints a far-layer box with piers (every third column a shade
// darker, the window bays of a tower) and a shade column on its right edge.
func box(c *canvas.Canvas, x0, y0, x1, y1 int, col rgb) {
	fill(c, x0, y0, x1, y1, col)
	for x := x0 + 2; x < x1; x += 3 {
		fill(c, x, y0, x, y1, darken(col, 0.07))
	}
	if x1 > x0 {
		fill(c, x1, y0, x1, y1, darken(col, 0.18))
	}
}

// line plots a straight run of glyphs from (x0,y0) to (x1,y1), the glyph
// chosen by the slope, on a layer at an alpha. For cables, stays, jibs.
func line(l *canvas.Layer, x0, y0, x1, y1 int, col rgb, a float64) {
	dx, dy := x1-x0, y1-y0
	n := max(abs(dx), abs(dy))
	if n == 0 {
		plot(l, x0, y0, '-', col, a)
		return
	}
	r := '-'
	if abs(dy)*3 > abs(dx)*2 {
		r = '|'
		if abs(dx)*3 > abs(dy) {
			if (dx > 0) == (dy > 0) {
				r = '\\'
			} else {
				r = '/'
			}
		}
	} else if abs(dy)*4 > abs(dx) {
		if (dx > 0) == (dy > 0) {
			r = '\\'
		} else {
			r = '/'
		}
	}
	for k := 0; k <= n; k++ {
		x := x0 + int(math.Round(float64(dx*k)/float64(n)))
		y := y0 + int(math.Round(float64(dy*k)/float64(n)))
		plot(l, x, y, r, col, a)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// empire is the Empire's cousin: a nine-wide shaft from row top, three
// setbacks, a mast, the crown lit by night, and its floors lit from the
// bottom for finished todos when RainTodoLit says so.
func empire(c *canvas.Canvas, x, top int, haze, warm rgb, lf float64) {
	box(c, x, top, x+8, farBot, haze)
	box(c, x+1, top-2, x+7, top-1, haze)
	box(c, x+2, top-4, x+6, top-3, haze)
	box(c, x+3, top-5, x+5, top-5, haze)
	plot(c.Mid(), x+4, top-6, '|', haze, 1)
	if !rainAnim || math.Mod(rainT, 1) < 0.5 {
		plot(c.Far(), x+4, top-7, '.', cube(215, 0, 0), 1)
	}
	if lf > 0.3 {
		for xx := x + 3; xx <= x+5; xx++ {
			plot(c.Far(), xx, top-4, '\'', warm, 0.9*lf)
			plot(c.Far(), xx, top-3, '.', warm, 0.9*lf)
		}
	}
	for k := 0; k < RainTodoLit && k < 7; k++ {
		y := 12 - k
		x0, x1 := x+1, x+7
		if y < top {
			x0, x1 = x+2, x+6
		}
		if y < top-2 {
			x0, x1 = x+3, x+5
		}
		for xx := x0; xx <= x1; xx += 2 {
			plot(c.Far(), xx, y, '.', warm, 0.85)
		}
	}
	for y := top + 1; y <= 11; y += 2 {
		for xx := x + 1; xx < x+8; xx += 2 {
			if RainTodoLit == 0 && scape.HashF(xx, y, 12) < 0.16*lf {
				plot(c.Far(), xx, y, '.', warm, 0.45)
			}
		}
	}
}

// paintFar draws the far skyline, authored: four silhouettes.
func paintFar(c *canvas.Canvas, p scape.Palette, l float64, seed int64, sky int) {
	haze, far2 := hazeAt(p, l), hazeFar(p, l)
	lf := litFrac(l)
	far, mid := c.Far(), c.Mid()
	warm := cube(255, 215, 135)
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	pins := func(x0, y0, x1, y1 int) {
		for y := y0 + 1; y <= y1; y += 2 {
			for x := x0 + 1; x < x1; x += 2 {
				if scape.HashF(x, y, seed+12) < 0.14*lf {
					plot(far, x, y, '.', warm, 0.4)
				}
			}
		}
	}
	switch sky {
	case 0: // MIDTOWN
		// The layer behind: slabs in the lighter haze.
		box(c, 6, 8, 10, farBot, far2)
		box(c, 27, 6, 29, farBot, far2)
		box(c, 38, 4, 44, farBot, far2)
		box(c, 58, 7, 61, farBot, far2)
		box(c, 68, 5, 70, farBot, far2)
		// The tapered crown, Chrysler's cousin.
		box(c, 19, 5, 25, farBot, haze)
		fill(c, 20, 4, 24, 4, haze)
		c.SetBGQuad(20, 3, haze, ground(20, 3), 0b0001)
		fill(c, 21, 3, 23, 3, haze)
		c.SetBGQuad(24, 3, haze, ground(24, 3), 0b0010)
		c.SetBGQuad(21, 2, haze, ground(21, 2), 0b0001)
		fill(c, 22, 2, 22, 2, haze)
		c.SetBGQuad(23, 2, haze, ground(23, 2), 0b0010)
		c.SetBGQuad(22, 1, haze, ground(22, 1), 0b0011)
		plot(mid, 22, 0, '|', haze, 1)
		pins(19, 5, 25, 11)
		box(c, 31, 6, 35, farBot, haze)
		pins(31, 6, 35, 11)
		empire(c, 48, 5, haze, warm, lf)
		box(c, 62, 3, 66, farBot, haze)
		fill(c, 63, 2, 64, 2, haze)
		pins(62, 3, 66, 11)
		box(c, 72, 8, 76, farBot, haze)
	case 1: // THE ROW: a low dense silhouette, one landmark.
		for x := 2; x <= 77; x++ {
			seg := x / 3
			h := 5 + int(scape.HashF(seg, 51, seed)*4)
			if x >= 10 && x <= 16 {
				h = 8
			}
			fill(c, x, h, x, farBot, haze)
			if x%3 == 2 {
				fill(c, x, h, x, farBot, darken(haze, 0.14))
			}
		}
		empire(c, 56, 5, haze, warm, lf)
	case 2: // THE BRIDGE
		// Towers behind, in the lighter haze, and two nearer.
		box(c, 6, 7, 10, farBot, far2)
		box(c, 20, 6, 25, farBot, far2)
		box(c, 30, 4, 34, farBot, haze)
		pins(30, 4, 34, 11)
		box(c, 66, 6, 71, farBot, far2)
		box(c, 74, 8, 77, farBot, haze)
		// The tower's stone, darker and warmer than the haze; twin pointed
		// arches over a centre pier, the crossbeam above them.
		stone := term.Lerp(darken(haze, 0.15), cube(135, 95, 95), 0.15)
		fill(c, 46, 1, 47, farBot, stone)
		fill(c, 53, 1, 54, farBot, stone)
		fill(c, 48, 1, 52, 2, stone)
		fill(c, 50, 3, 50, farBot, stone)
		c.SetBGQuad(48, 3, stone, ground(48, 3), 0b1000)
		c.SetBGQuad(49, 3, stone, ground(49, 3), 0b1100)
		c.SetBGQuad(51, 3, stone, ground(51, 3), 0b1100)
		c.SetBGQuad(52, 3, stone, ground(52, 3), 0b0100)
		// The deck, and its lamps by night.
		fill(c, 4, 8, 77, 8, stone)
		if lf > 0.3 {
			for x := 6; x <= 76; x += 5 {
				plot(far, x, 7, '.', warm, 0.8*lf)
			}
		}
		// The main cables, sagging from the tower top to the anchorages,
		// with the necklace of lights; the fan of stays under them.
		cable := func(dir int) {
			n := 26.0
			for k := 0; k <= 26; k++ {
				u := float64(k) / n
				x := 50 + dir*(4+k)
				y := int(math.Round(1 + 6*u*u))
				r := '-'
				if int(math.Round(1+6*math.Pow((float64(k)+1)/n, 2))) != y {
					if dir > 0 {
						r = '\\'
					} else {
						r = '/'
					}
				}
				plot(mid, x, y, r, haze, 0.9)
				if lf > 0.3 && k%3 == 0 {
					plot(far, x, y, '.', warm, 0.8*lf)
				}
				if k%5 == 0 && y < 7 {
					for yy := y + 1; yy < 8; yy++ {
						plot(far, x, yy, '|', haze, 0.4)
					}
				}
			}
			for _, d := range []int{9, 15, 21} {
				line(far, 50+dir*4, 2, 50+dir*(4+d), 7, haze, 0.5)
			}
		}
		cable(1)
		cable(-1)
	case 3: // ANYCITY: hills, glass slabs, a mast, a crane. Nobody's skyline.
		for x := 2; x <= 77; x++ {
			u := float64(x) / 80
			h := 7.5 - 2.2*math.Sin(u*3.1+0.4) - 1.1*math.Sin(u*9.7+2.0)
			fill(c, x, int(math.Round(h)), x, farBot, far2)
		}
		box(c, 8, 8, 11, farBot, haze)
		box(c, 22, 4, 27, farBot, haze)
		pins(22, 4, 27, 11)
		box(c, 33, 7, 36, farBot, haze)
		box(c, 45, 3, 52, farBot, haze)
		pins(45, 3, 52, 11)
		plot(mid, 48, 2, '|', haze, 1)
		plot(mid, 48, 1, '|', haze, 1)
		plot(mid, 48, 0, '.', cube(215, 0, 0), 1)
		box(c, 60, 6, 64, farBot, haze)
		pins(60, 6, 64, 11)
		box(c, 70, 9, 74, farBot, haze)
		// The crane: a mast, its jib, the counter-jib, the hook line.
		for y := 3; y <= 9; y++ {
			plot(mid, 57, y, '|', haze, 1)
		}
		line(mid, 55, 3, 66, 3, haze, 1)
		line(mid, 57, 2, 63, 3, haze, 0.7)
		plot(mid, 64, 4, '|', haze, 0.8)
		plot(mid, 64, 5, '|', haze, 0.8)
	}
	// Aerial perspective: whatever stands in the far layer is lifted
	// toward the horizon's colour the lower it goes, the haze thicker
	// toward the ground.
	for y := 7; y <= farBot; y++ {
		// 0.35 at the bottom turned the whole band orange at dusk (the
		// fourth render); 0.25 keeps the glow a cast.
		k := 0.25 * float64(y-7) / float64(farBot-7)
		for x := 0; x < c.W; x++ {
			c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.SkyHorizon, k))
		}
	}
}

// paintBlock draws the walk-ups: wall, cornice, shade column, windows two
// wide, the roof furniture in the building's own tone.
func paintBlock(c *canvas.Canvas, blocks []building, base func(x int) int, l float64, seed int64, litMax int) {
	lf := litFrac(l)
	tanks := 0
	smoked := false
	if rainAnim {
		// A plane crossing the top of the sky with its lights, once a clip.
		px := -6 + int(rainT*6.2)
		if px >= -3 && px < c.W+3 {
			for dx := 0; dx <= 2; dx++ {
				plot(c.Far(), px+dx, 1, '-', grey(14), 0.8)
			}
			if math.Mod(rainT*2.5, 1) < 0.5 {
				plot(c.Far(), px-1, 1, '.', cube(215, 0, 0), 1)
			} else {
				plot(c.Far(), px+3, 1, '.', cube(0, 175, 0), 1)
			}
		}
	}
	for bi, b := range blocks {
		wall := wallOf(b, l)
		// Per column: the lower roof's edge can fall inside a building,
		// and a building filled to one base left a black column there
		// (the third render, 2026-09-23).
		for x := b.x0; x <= b.x1; x++ {
			fill(c, x, b.top, x, base(x), wall)
		}
		bot := base(b.x0)
		fill(c, b.x0, b.top, b.x1, b.top, lighten(wall, 0.18))
		fill(c, b.x1, b.top+1, b.x1, base(b.x1), darken(wall, 0.18))
		// Windows: two cells wide, four apart, one a storey.
		glass := darken(wall, 0.28)
		for x := b.x0 + 2; x+1 <= b.x1-2; x += 4 {
			for y := b.top + 2; y < bot; y += 2 {
				col := glass
				apt := 100*bi + (x-b.x0)/8
				pLit := 0.28 * lf
				if RainWindows == 1 {
					pLit = (0.04 + 0.6*RainSpend) * math.Max(lf, 0.5)
				}
				on := scape.HashF(apt, y, seed+9) < pLit
				if rainAnim && scape.HashF(apt, y, seed+13) < 0.08 {
					// One window in twelve switches now and then.
					if int(math.Floor((rainT+scape.HashF(x, y, seed+14)*5)/5))%2 == 1 {
						on = !on
					}
				}
				if y <= litMax && on {
					// Dim amber through wet glass; a few paler, one in
					// twenty cool. (255,215,135) stood out too much.
					k := scape.HashF(x, y, seed+10)
					col = cube(175, 135, 95)
					if k < 0.05 {
						col = cube(135, 175, 215)
					} else if k < 0.3 {
						col = cube(215, 175, 135)
					}
				}
				fill(c, x, y, x+1, y, col)
				// The sill under it, catching the light.
				if y+1 < bot {
					fill(c, x, y+1, x+1, y+1, lighten(wall, 0.1))
				}
			}
		}
		// Dentils under the cornice: a row of small teeth every other cell.
		for x := b.x0 + 1; x < b.x1; x += 2 {
			plot(c.Mid(), x, b.top+1, '\'', lighten(wall, 0.25), 0.8)
		}
		// A fire escape on the facade: landings every two storeys and the
		// stair between them, a shade darker than the wall.
		if b.escape {
			iron := darken(wall, 0.35)
			ex := b.x0 + 5
			for y := b.top + 3; y < bot; y += 4 {
				fill(c, ex, y, ex+3, y, iron)
				plot(c.Mid(), ex+1, y+1, '/', iron, 1)
				plot(c.Mid(), ex+2, y+2, '/', iron, 1)
			}
		}
		// The roof furniture in the building's own tone, a step darker.
		sil := darken(wall, 0.12)
		mx := (b.x0 + b.x1) / 2
		switch b.roof {
		case 1:
			if tanks < 1 && b.x1-b.x0 >= 9 {
				waterTank(c, mx-1, b.top-1, 4, 2, 1, sil)
				tanks++
			}
		case 2:
			fill(c, b.x0+2, b.top-1, b.x0+4, b.top-1, sil)
		case 3:
			fill(c, b.x1-2, b.top-2, b.x1-2, b.top-1, sil)
			if rainAnim && !smoked {
				// Smoke drifting off the chimney, up and to the right.
				for k := 0; k < 5; k++ {
					u := math.Mod(rainT*0.9+float64(k)*1.3, 6)
					plot(c.Far(), b.x1-2+int(u*1.2), b.top-3-int(u*0.6), ',', grey(12), 0.5*(1-u/6))
				}
				smoked = true
			}
		}
	}
}

// paintWindows is kept for the pieces of the first rounds that still call
// it; the block's windows are paintBlock's now.

// waterTank draws a New York roof tank: a conical cap, a tank tankH rows
// tall and w wide, on legs legH rows tall, its feet on row foot, in col.
func waterTank(c *canvas.Canvas, x, foot, w, tankH, legH int, col rgb) {
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	tankBot := foot - legH
	tankTop := tankBot - tankH + 1
	fill(c, x, tankTop, x+w-1, tankBot, col)
	c.SetBGQuad(x, tankTop-1, col, ground(x, tankTop-1), 0b0001)
	if w > 2 {
		fill(c, x+1, tankTop-1, x+w-2, tankTop-1, col)
	}
	c.SetBGQuad(x+w-1, tankTop-1, col, ground(x+w-1, tankTop-1), 0b0010)
	mx := x + w/2
	if w%2 == 0 {
		c.SetBGQuad(mx-1, tankTop-2, col, ground(mx-1, tankTop-2), 0b0001)
		c.SetBGQuad(mx, tankTop-2, col, ground(mx, tankTop-2), 0b0010)
	} else {
		c.SetBGQuad(mx, tankTop-2, col, ground(mx, tankTop-2), 0b0011)
	}
	for y := tankBot + 1; y <= foot; y++ {
		c.SetBGQuad(x, y, col, ground(x, y), 0b0101)
		c.SetBGQuad(x+w-1, y, col, ground(x+w-1, y), 0b1010)
	}
}

func paintRainStudy(c *canvas.Canvas, tod, t, level float64, seed int64, stage rainStage, nkit int) {
	if RainKittens >= 0 {
		nkit = RainKittens
	}
	p := scape.PaletteAt(tod)
	l := lit(p)
	tn := rainTonesAt(l)
	far := c.Far()
	if rainAnim {
		p = overcast(p)
	}

	// 1. The sky, its stars, the moon: the study's block, or in the clip the
	// shore's disc sinking as the context fills.
	vramp(c, 0, 0, c.W-1, 12, p.SkyTop, p.SkyHorizon)
	fill(c, 0, 13, c.W-1, rainPaneBot, p.SkyHorizon)
	for y := 0; y <= 12; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed) < 0.05*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.6*p.StarVis)
			}
		}
	}
	if rainAnim {
		drawDisc(c, p, rainCtx)
	} else if p.MoonVis > 0.05 {
		for y := 2; y <= 3; y++ {
			for x := 12; x <= 14; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
	}

	// 2. The far skyline; or one of the four other views, which draw
	// everything outside the glass themselves (rain_views.go).
	if RainView != 0 {
		paintView(c, p, l, tn, seed, nkit, t, level)
		paintGlassAndInside(c, tn, t, level, seed, stage, nkit)
		return
	}
	paintFar(c, p, l, seed, RainSkyline)

	// 3. The block. Its facades run to the base row, which the roof stage
	// raises where the lower roof stands in front; its tops depend on the
	// skyline (low under midtown and the bridge, tall for the row).
	base := func(x int) int { return rainPaneBot }
	if stage == rainStageRoof {
		base = func(x int) int {
			if x <= 57 {
				return 13
			}
			return rainPaneBot
		}
	}
	minTop, maxTop := 9, 11
	if RainSkyline == 1 {
		minTop, maxTop = 7, 9
	}
	blocks := walkups(0, c.W-1, minTop, maxTop, 10, 14, seed+200)
	for i := range blocks {
		blocks[i].escape = i%2 == 1 && blocks[i].x1-blocks[i].x0 >= 10
	}
	if stage == rainStageRoof {
		// The lower roof's tank stands against the sky: the walk-up behind
		// it is a low one.
		for i := range blocks {
			if blocks[i].x0 <= 30 && blocks[i].x1 >= 18 && blocks[i].top < 11 {
				blocks[i].top = 11
				blocks[i].roof = 0
				blocks[i].escape = false
			}
		}
	}
	litMax := 12
	if stage == rainStageEscape {
		// The landing is seen against the sky: the buildings across the
		// left pane are low ones, so the rail, the balusters and the
		// kittens stand on haze, not on a facade.
		litMax = 11
		for i := range blocks {
			if blocks[i].x0 <= 40 {
				blocks[i].top = 15
				blocks[i].roof = 0
				blocks[i].escape = false
			}
		}
	}
	paintBlock(c, blocks, base, l, seed, litMax)

	// 4. The stage.
	switch stage {
	case rainStageRoof:
		paintNearRoof(c, p, tn, seed, 0, 57, 13, rainPaneBot, 22)
		if !drawKits(c) {
			xs := []int{24, 34, 44, 28, 38, 50, 22, 54}
			for i := 0; i < nkit && i < len(xs); i++ {
				drawKittenAt(c.Near(), xs[i], rainPaneBot-3, i%2 == 1)
			}
		}
	case rainStageEscape:
		paintFireEscape(c, tn, 2, 38, 11, 16)
		xs := []int{10, 20, 30, 15, 25, 5, 33}
		for i := 0; i < nkit && i < len(xs); i++ {
			drawKittenAt(c.Near(), xs[i], 12, i%2 == 1)
		}
	}

	paintGlassAndInside(c, tn, t, level, seed, stage, nkit)
}

// paintGlassAndInside is the glass (the rain, the work), the frame, the
// sill, the cat, the family on the sill if that is the stage, and the desk.
func paintGlassAndInside(c *canvas.Canvas, tn rainTones, t, level float64, seed int64, stage rainStage, nkit int) {
	// 5. The glass: the rain, the work. The study's streaks in the stills;
	// in the clip, drops running down the glass (paintDrops).
	mid := c.Mid()
	if rainAnim {
		paintDrops(c, t, seed, RainView)
	} else {
		d := 0.04 + 0.16*level
		shift := int(math.Round(t * 6))
		for y := 0; y <= rainPaneBot; y++ {
			for x := 0; x < c.W; x++ {
				h := scape.HashF(x, ((y-shift)%RainPeriod+RainPeriod)%RainPeriod, seed+20)
				if h < d {
					r := '|'
					if h < d*0.35 {
						r = '\''
					}
					plot(mid, x, y, r, cube(135, 175, 215), 0.6)
				} else if h > 0.985 {
					plot(mid, x, y, ',', cube(175, 215, 255), 0.5)
				}
			}
		}
	}

	// 6. The frame: two posts, the centre mullion, the meeting rail. Wood,
	// so nothing outside shows on it: the glyph layers are cleared there.
	// Painted wood, pale, so it stands in front of whatever the view is
	// ("the window frame on the foreground gets lost", his note on the
	// harbour): a lit face and, on the side toward the glass, a shadowed
	// rabbet two steps darker.
	frameCell := func(x, y int, col rgb) {
		setSolid(c, x, y, col)
		c.Far().Erase(x, y)
		c.Mid().Erase(x, y)
		c.Near().Erase(x, y)
	}
	rabbet := clampGrey(greyK(tn.frame) - 3)
	for y := 0; y <= rainPaneBot; y++ {
		frameCell(0, y, tn.frame)
		frameCell(1, y, rabbet)
		frameCell(39, y, rabbet)
		frameCell(40, y, tn.frame)
		frameCell(78, y, rabbet)
		frameCell(79, y, tn.frame)
	}
	for x := 0; x < c.W; x++ {
		frameCell(x, rainRail, tn.frame)
	}
	drawKitsNow(c)

	// 7. Inside: the sill, the cat on it, the family beside it if that is
	// the stage, and the desk with the writing.
	fill(c, 0, rainSill, c.W-1, rainSill, tn.sill)
	if stage == rainStageSill && RainView == 0 {
		x := rainCatX - 2
		for i := 0; i < nkit; i++ {
			x -= 7
			if x < 2 {
				break
			}
			drawKittenAt(c.Near(), x, rainSill-4, i%2 == 0)
		}
	}
	if RainPot && RainView == 4 {
		drawSillPlant(c, 12, tn.l)
	}
	if rainAnim {
		drawClipCat(c, t)
		writeTail(c, rainBand, tn.desk, rainArcN.tail)
		return
	}
	catAt(c, rainCatX, rainCatY, true, t)
	writeBand(c, rainBand, tn.desk)
}

// RainPot puts the potted plant on the sill (his idea, 2026-09-24: "I like
// the idea of a plant pot on the window sill, explore it further").
var RainPot = true

// drawSillPlant is a trailing plant in a terracotta pot on the sill, inside
// the room like the cat: a dome of leaves in two greens over the pot, with
// vines hanging down its sides to the sill; the pot has a rim, a tapered
// body and a saucer. Drawn opaque over the glass, as the cat is. The room
// is lit, so the greens hold by night, a shade darker.
func drawSillPlant(c *canvas.Canvas, x int, l float64) {
	leaf := cube(0, 135, 0)
	leafLit := cube(95, 175, 95)
	if l < 0.3 {
		leaf, leafLit = cube(0, 95, 0), cube(0, 135, 0)
	}
	pot, rim, saucer := cube(175, 95, 95), cube(215, 135, 135), cube(135, 95, 95)
	near := c.Near()
	ground := func(x, y int) rgb { return c.BGAt(x, y) }
	clear := func(x0, y0, x1, y1 int) {
		for y := y0; y <= y1; y++ {
			for xx := x0; xx <= x1; xx++ {
				c.Far().Erase(xx, y)
				c.Mid().Erase(xx, y)
			}
		}
	}
	// The dome: rows 14..16, eight cells wide, the top row soft, the lit
	// side up and left.
	clear(x, 13, x+7, 19)
	for xx := x + 2; xx <= x+5; xx++ {
		c.SetBGQuad(xx, 13, leaf, ground(xx, 13), 0b0011)
	}
	fill(c, x, 14, x+7, 14, leaf)
	fill(c, x+1, 14, x+3, 14, leafLit)
	fill(c, x-1, 15, x+8, 15, leaf)
	fill(c, x, 15, x+2, 15, leafLit)
	fill(c, x, 16, x+7, 16, leaf)
	c.SetBGQuad(x-1, 16, leaf, ground(x-1, 16), 0b0100)
	c.SetBGQuad(x+8, 16, leaf, ground(x+8, 16), 0b1000)
	// The leaves themselves, over the dome: rounded leaves as paired
	// brackets on the lit side, their undersides as dots on the shade side,
	// a few gaps into the dark of the bush, and three blossoms. Then the
	// tips out of the dome and the vines down the pot's sides.
	deep := darken(leaf, 0.45)
	for _, g := range [][2]int{{x + 1, 14}, {x + 4, 15}, {x + 6, 14}, {x + 2, 16}, {x + 5, 16}} {
		c.SetBG(g[0], g[1], deep)
	}
	for _, g := range [][3]int{{x, 14, '('}, {x + 2, 14, ')'}, {x + 3, 15, '('}, {x + 1, 15, ')'}, {x + 5, 14, '('}, {x + 7, 15, ')'}, {x + 3, 16, '('}, {x + 6, 16, ')'}} {
		plot(near, g[0], g[1], rune(g[2]), leafLit, 1)
	}
	for _, g := range [][2]int{{x + 4, 14}, {x + 7, 16}, {x + 1, 16}, {x + 6, 15}} {
		plot(near, g[0], g[1], '.', deep, 1)
	}
	blossom := cube(215, 135, 135)
	plot(near, x+2, 13, '*', blossom, 1)
	plot(near, x+6, 13, '*', blossom, 1)
	plot(near, x+4, 15, '*', blossom, 1)
	plot(near, x-1, 14, '(', leafLit, 1)
	plot(near, x+8, 14, ')', leaf, 1)
	plot(near, x+3, 12, '\'', leafLit, 1)
	plot(near, x+5, 12, '\'', leaf, 1)
	plot(near, x-1, 17, '(', leaf, 1)
	plot(near, x-2, 18, '(', leaf, 1)
	plot(near, x-2, 19, '*', leafLit, 1)
	plot(near, x+8, 17, ')', leaf, 1)
	plot(near, x+8, 18, ')', leaf, 1)
	plot(near, x+9, 19, '*', leafLit, 1)
	// The pot: rim, body tapering, the saucer on the sill.
	fill(c, x, 17, x+7, 17, rim)
	fill(c, x+1, 18, x+6, 18, pot)
	c.SetBGQuad(x, 18, pot, ground(x, 18), 0b0101)
	c.SetBGQuad(x+7, 18, pot, ground(x+7, 18), 0b1010)
	fill(c, x+1, 19, x+6, 19, saucer)
	fill(c, x+2, 19, x+5, 19, pot)
}

// paintNearRoof is the lower building's roof the window looks down on,
// from x0..x1, its far edge (where the block behind it rises) at top, down
// to bot. Everything on it is in the roof's own greys.
func paintNearRoof(c *canvas.Canvas, p scape.Palette, tn rainTones, seed int64, x0, x1, top, bot, tankX int) {
	k := greyK(tn.tar)
	fill(c, x0, top, x1, top, clampGrey(k+2))
	fill(c, x0, top+1, x1, bot, tn.tar)
	for y := top + 2; y <= bot; y++ {
		for x := x0; x <= x1; x++ {
			if scape.HashF(x/3, y, seed+60) < 0.12 && scape.HashF(x, y, seed+61) < 0.8 {
				c.SetBG(x, y, term.Lerp(tn.tar, p.SkyHorizon, 0.3))
			} else if scape.HashF(x, y, seed+62) < 0.05 {
				plot(c.Far(), x, y, '.', clampGrey(k+3), 0.5)
			}
		}
	}
	// The tank: wood, a step lighter than the tar, standing against the sky
	// where the walk-up behind is low. Its feet on the roof.
	wood := clampGrey(k + 2)
	waterTank(c, tankX, top+1, 5, 3, 2, wood)
	// The bulkhead at the right end, with its door lit by night: the
	// kittens' way on and off the roof.
	fill(c, 46, top-2, 51, top, wood)
	fill(c, 51, top-2, 51, top, clampGrey(k))
	door := clampGrey(k)
	if litFrac(tn.l) > 0.4 {
		door = cube(255, 215, 135)
	}
	fill(c, 48, top-1, 49, top, door)
	// A skylight, lit from inside by night, glass by day; two vents.
	sk := term.Lerp(tn.tar, p.SkyHorizon, 0.45)
	if litFrac(tn.l) > 0.5 {
		sk = cube(175, 135, 95)
	}
	fill(c, 7, top+3, 11, top+4, sk)
	fill(c, 6, top+3, 6, top+4, wood)
	fill(c, 12, top+3, 12, top+4, wood)
	fill(c, 36, top+1, 36, top+1, wood)
	fill(c, 39, top+1, 39, top+1, wood)
}

// greyK is the ramp index of a grey the painters made, for stepping it.
func greyK(g rgb) int { return (int(g.R) - 8) / 10 }

// paintFireEscape is a landing outside the glass: two stringers rising out
// of the top of the pane, a top rail that catches the light, balusters,
// the grating with its lit lip and its slots, braces under it, and the
// stair down from the left end with its handrail.
func paintFireEscape(c *canvas.Canvas, tn rainTones, x0, x1, railRow, platRow int) {
	mid, near := c.Mid(), c.Near()
	for y := 0; y < platRow; y++ {
		plot(mid, x0+1, y, '|', tn.ironLit, 1)
		plot(mid, x1, y, '|', tn.ironLit, 1)
	}
	for x := x0; x <= x1; x++ {
		plot(mid, x, railRow, '=', tn.ironLit, 1)
		if x%2 == 1 {
			for y := railRow + 1; y < platRow; y++ {
				plot(mid, x, y, '|', tn.ironDark, 1)
			}
		} else {
			plot(mid, x, railRow+2, '-', tn.ironDark, 1)
		}
	}
	for x := x0 + 1; x <= x1; x += 9 {
		for y := railRow + 1; y < platRow; y++ {
			plot(mid, x, y, '|', tn.ironLit, 1)
		}
	}
	fill(c, x0, platRow, x1, platRow, tn.ironDark)
	for x := x0; x <= x1; x++ {
		r := '='
		if x%2 == 1 {
			r = '-'
		}
		plot(near, x, platRow, r, tn.ironLit, 1)
	}
	for x := x0 + 13; x <= x1; x += 12 {
		plot(near, x, platRow+1, '/', tn.ironDark, 1)
		plot(near, x-1, platRow+2, '/', tn.ironDark, 1)
		plot(near, x+1, platRow+1, '\\', tn.ironDark, 1)
		plot(near, x+2, platRow+2, '\\', tn.ironDark, 1)
	}
	for k := 0; k < 3; k++ {
		y := platRow + 1 + k
		x := x0 + 8 - 2*k
		plot(near, x, y, '/', tn.ironLit, 1)
		plot(near, x-1, y, '/', tn.ironLit, 1)
		plot(near, x+2, y, '/', tn.ironDark, 1)
	}
}

// The kittens' coat and eye, the companion package's own (kittens.go).
var (
	kitFur = term.RGB{R: 236, G: 228, B: 210}
	kitEye = term.RGB{R: 150, G: 214, B: 158}
)

// drawKittenAt draws one sitting kitten with its top-left at (x, y): 6x4
// cells, the way DrawKittens draws a sitter, without the litter's layout.
func drawKittenAt(near *canvas.Layer, x, y int, mirror bool) {
	bm := companion.ParseBitmap(companion.KittenSit)
	if mirror {
		bm = bm.Mirrored()
	}
	rows := bm.ToQuadrant()
	companion.PlotRim(near, rows, x, y)
	(&companion.Sprite{Rows: rows, Body: kitFur, Alpha: 1}).Draw(near, x, y)
	e1, e2 := x+1, x+3
	if mirror {
		e1, e2 = x+6-1-1, x+6-1-3
	}
	near.Plot(e1, y+1, 'o', kitEye, 1)
	near.Plot(e2, y+1, 'o', kitEye, 1)
}

// drawKittenAnim is drawKittenAt with the shore litter's life: the breath
// lifts the body two source rows (half a cell), the blink shuts the eyes.
func drawKittenAnim(near *canvas.Layer, x, y int, mirror, lift, blink bool) {
	bm := companion.ParseBitmap(companion.KittenSit)
	if lift {
		f := bm.Blank()
		for sy := 0; sy < f.H; sy++ {
			for sx := 0; sx < f.W; sx++ {
				if sy+2 < bm.H && bitmapAt(bm, sx, sy+2) {
					f.Set(sx, sy)
				}
			}
		}
		bm = f
	}
	if mirror {
		bm = bm.Mirrored()
	}
	rows := bm.ToQuadrant()
	companion.PlotRim(near, rows, x, y)
	(&companion.Sprite{Rows: rows, Body: kitFur, Alpha: 1}).Draw(near, x, y)
	e1, e2 := x+1, x+3
	if mirror {
		e1, e2 = x+6-1-1, x+6-1-3
	}
	glyph := 'o'
	if blink {
		glyph = '-'
	}
	near.Plot(e1, y+1, glyph, kitEye, 1)
	near.Plot(e2, y+1, glyph, kitEye, 1)
}

// bitmapAt reads one source pixel of a bitmap through its quadrant rows,
// since the bitmap's cells are not exported: the row of quadrants is
// rebuilt per call, which is fine for a 12x16 sprite in a study.
func bitmapAt(b *companion.Bitmap, x, y int) bool {
	rows := companion.KittenSit
	if y < 0 || y >= len(rows) || x < 0 || x >= len(rows[y]) {
		return false
	}
	return rows[y][x] == '#'
}
