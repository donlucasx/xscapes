package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE THIRD SCAPE: a mountain in the woods, with the owl.
//
// His ask, 2026-09-15: "sketching out some options for the 3rd scape, the
// mountain scape in the woods w the owl." The design question is the one the
// shore never had to answer, because the sea answered it: with no water, what
// carries the work? Three candidates below, one answer each, all drawn on the
// same ridge with the same owl so only the answer differs.
//
// Every painter closes on itself after LoopSecs: wind, sparks and water take
// their phase from t modulo the loop, so the page can cycle any of them.
var Forest = []Scene{
	{"Wind in the pines",
		[6]string{"the moon on the snowcap, daylight on the rock", "the real sky over the ridge", "the wind: how many crowns sway and how far, and the needles it carries", "rock, snow, the treeline, the forest floor", "the floor, in the needles", "the owl on a stump, right"},
		"The work is the wind. A light wind stirs the tallest crown; a gale takes all four and fills the air with needles. Count and coverage, no rate: a still frame says how hard from how many trees lean and how far. It is the one candidate whose motion is neither water nor fire, which is what makes it a third scape and not a re-skin of the first two.",
		paintPines},
	{"The fire in the clearing",
		[6]string{"the fire, and the sky", "the real sky over the ridge", "the flame's height and its sparks", "the clearing, stones, logs", "the ground by the fire", "the owl on a log, right"},
		"The work is the fire. The strongest read of the three at a glance, and the classic picture. The cost: the hearth already carries the work as a fire, and here the fire is the light slot too, so a busy agent brightens the world it is in.",
		paintClearing},
	{"The creek",
		[6]string{"the sky", "the real sky over the ridge", "whitewater: how much of the creek is foam", "the water, rocks, the banks", "the near bank", "the owl on a boulder, right"},
		"The work is the water again, verbatim from the shore: foam coverage rises with the agent. It is the safe option and it reads well, and it is also the reason the other two exist, because the aquarium already answered 'water again'.",
		paintCreek},
}

// LoopSecs is the loop every forest painter closes on.
const LoopSecs = 4.0

const (
	ridgeBase = 9  // the row the mountains stand on
	floorTop  = 18 // the first row of the forest floor
	bandTop   = 21 // the writing
)

// mountainHeight is the ridge's profile: two peaks, the taller one with a
// snowcap, both clear of the moon at x 12..14.
func mountainHeight(x int) int {
	a := 7 - math.Abs(float64(x-30))*7/16
	b := 5 - math.Abs(float64(x-58))*5/12
	return int(math.Round(math.Max(0, math.Max(a, b))))
}

// paintRidge is everything the three candidates share: the sky and its
// furniture, the ridge, the far treeline, the deep woods and the floor.
func paintRidge(c *canvas.Canvas, p scape.Palette, l float64, seed int64) {
	vramp(c, 0, 0, c.W-1, ridgeBase-1, p.SkyTop, p.SkyHorizon)
	far, mid := c.Far(), c.Mid()
	for y := 0; y < ridgeBase; y++ {
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
	rock, rim, snow := greyBetween(2, 9, l), greyBetween(4, 12, l), greyBetween(14, 22, l)
	for x := 0; x < c.W; x++ {
		h := mountainHeight(x)
		if h <= 0 {
			continue
		}
		top := ridgeBase - h
		for y := top; y < ridgeBase; y++ {
			c.SetBG(x, y, rock)
		}
		c.SetBG(x, top, rim)
		if h >= 6 {
			for y := top; y <= top+1 && y < ridgeBase; y++ {
				c.SetBG(x, y, snow)
			}
		}
	}
	// The far treeline on the ridge's foot, and the deep woods behind
	// everything nearer.
	fill(c, 0, ridgeBase, c.W-1, floorTop-1, greyBetween(1, 5, l))
	// Half-transparent, so it sits BEHIND the near pines: the cube has no
	// darker green than (0,95,0), so depth has to come from alpha.
	green := cube(0, 95, 0)
	for x := 0; x < c.W; x++ {
		if scape.HashF(x, ridgeBase, seed+40) < 0.55 {
			plot(mid, x, ridgeBase, '^', green, 0.5)
		}
		if scape.HashF(x, ridgeBase+1, seed+40) < 0.85 {
			plot(mid, x, ridgeBase+1, '^', green, 0.5)
		}
	}
	fill(c, 0, floorTop, c.W-1, bandTop-1, greyBetween(3, 8, l))
	for x := 0; x < c.W; x++ {
		if scape.HashF(x, floorTop, seed+42) < 0.25 {
			plot(mid, x, floorTop, '"', green, 0.7)
		}
	}
}

// drawPine draws one near pine with its tip at base-(h-1) and its trunk on
// base. sway is the crown's lean in cells at the tip, tapering to nothing at
// the lowest bough.
func drawPine(near *canvas.Layer, cx, h, base int, sway float64) {
	green, trunk := cube(0, 95, 0), cube(95, 95, 0)
	for r := 0; r < h-1; r++ {
		y := base - (h - 1) + r
		half := r / 2
		dx := 0
		if sway != 0 && h > 2 {
			dx = int(math.Round(sway * float64(h-2-r) / float64(h-2)))
		}
		for x := cx - half; x <= cx+half; x++ {
			plot(near, x+dx, y, '^', green, 1)
		}
	}
	plot(near, cx, base, '|', trunk, 1)
}

// owlAt draws the owl, facing left, where the cat would sit.
func owlAt(c *canvas.Canvas, x, y int) {
	for i := range Animals {
		if Animals[i].Name == "Owl" {
			a := &Animals[i]
			drawAnimal(c.Near(), a.body, 24, 28, a.Name, companion.Coats[a.Coat],
				a.eyes, a.eyeRow, a.eyeGlyph, a.nose, a.noseRow, x, y, true)
			return
		}
	}
}

// loopPhase is t folded into the loop, so every rate below closes on itself.
func loopPhase(t float64) float64 { return math.Mod(t, LoopSecs) }

func paintPines(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	paintRidge(c, p, l, seed)
	mid, near := c.Mid(), c.Near()
	t = loopPhase(t)
	// Four near pines. How MANY sway is the count; how FAR is the coverage.
	// The tallest move first -- a light wind stirs the crowns that stand
	// highest, and a gale takes the lot. Two cycles of wind per loop.
	type pine struct {
		cx, h int
		phase float64
	}
	pines := []pine{{9, 8, 0.0}, {23, 10, 1.9}, {38, 7, 3.1}, {52, 9, 4.4}}
	byHeight := []int{1, 3, 0, 2}
	moving := int(math.Ceil(level*float64(len(pines)) - 1e-9))
	sway := map[int]float64{}
	for k := 0; k < moving && k < len(byHeight); k++ {
		i := byHeight[k]
		sway[i] = (0.6 + 1.8*level) * math.Sin(2*math.Pi*t/(LoopSecs/2)+pines[i].phase)
	}
	for i, pn := range pines {
		drawPine(near, pn.cx, pn.h, floorTop-1, sway[i])
	}
	// Needles in the air, blown right at twenty columns a second through a
	// pattern the frame's width, so the loop closes.
	d := 0.01 + 0.09*level
	shift := int(math.Round(t * 20))
	needle := cube(95, 135, 0)
	for y := 3; y < floorTop; y++ {
		for x := 0; x < c.W; x++ {
			h := scape.HashF(((x-shift)%c.W+c.W)%c.W, y, seed+50)
			if h < d {
				r := '\''
				if h < d*0.4 {
					r = ','
				} else if h < d*0.7 {
					r = '-'
				}
				plot(mid, x, y, r, needle, 0.8)
			}
		}
	}
	// The owl on a stump, right.
	fill(c, 66, floorTop, 73, floorTop+1, cube(135, 95, 95))
	owlAt(c, 64, floorTop-7)
	writeBand(c, bandTop, greyBetween(4, 8, l))
}

func paintClearing(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	paintRidge(c, p, l, seed)
	mid, near := c.Mid(), c.Near()
	t = loopPhase(t)
	for _, pn := range [][2]int{{6, 9}, {15, 7}, {77, 8}} {
		drawPine(near, pn[0], pn[1], floorTop-1, 0)
	}
	// The fire: its height is the work.
	cx, base := 40, floorTop-1
	H := 2 + int(math.Round(level*5))
	flick := int(math.Round(t*6)) % 24
	// Firelight on the ground and the trunks, further the higher it burns,
	// and mostly by night: by day the sun has the light slot.
	R := 5.0 + 9.0*level
	warm := cube(135, 95, 0)
	for y := ridgeBase + 2; y < bandTop; y++ {
		for x := 0; x < c.W; x++ {
			d := math.Hypot(float64(x-cx)/2, float64(y-base))
			if d < R {
				a := (1 - d/R) * (0.55*(1-l) + 0.08)
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), warm, a))
			}
		}
	}
	for r := 0; r < H; r++ {
		y := base - r
		w := (H - r) / 2
		for x := cx - w; x <= cx+w; x++ {
			h := scape.HashF(x, y*31+flick, seed+61)
			if h < 0.22 {
				continue
			}
			g := '^'
			if h < 0.5 {
				g = '*'
			} else if h < 0.6 {
				g = ')'
			}
			col := cube(255, 95, 0)
			if r > H/2 {
				col = cube(255, 175, 0)
			}
			if r >= H-1 {
				col = cube(255, 215, 135)
			}
			plot(near, x, y, g, col, 1)
		}
	}
	// Sparks rise one row a frame through a 24-row pattern, so they close.
	rise := int(math.Round(t * 6))
	spread := 2.0 + 6.0*level
	for y := 3; y <= base-H-1; y++ {
		for x := cx - 9; x <= cx+9; x++ {
			if math.Abs(float64(x-cx)) > spread {
				continue
			}
			if scape.HashF(x, ((y+rise)%24+24)%24, seed+62) < 0.005+0.035*level {
				plot(mid, x, y, '.', cube(255, 215, 135), 0.9)
			}
		}
	}
	// Stones and logs.
	for _, dx := range []int{-5, -4, 4, 5} {
		plot(near, cx+dx, floorTop, 'o', grey(9), 1)
	}
	for x := cx - 2; x <= cx+2; x++ {
		plot(near, x, floorTop, '=', cube(95, 95, 0), 1)
	}
	// The owl on a log, right.
	fill(c, 63, floorTop, 74, floorTop, cube(135, 95, 95))
	owlAt(c, 64, floorTop-7)
	writeBand(c, bandTop, greyBetween(4, 8, l))
}

// creekTop is the water's near edge at a column: it meanders a row.
func creekTop(x int) int { return 13 + int(math.Round(math.Sin(float64(x)/7.0))) }

func paintCreek(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	paintRidge(c, p, l, seed)
	mid, near := c.Mid(), c.Near()
	t = loopPhase(t)
	for _, pn := range [][2]int{{6, 9}, {16, 7}} {
		drawPine(near, pn[0], pn[1], floorTop-1, 0)
	}
	water := cube(0, 95, 135)
	if l > 0.5 {
		water = cube(0, 135, 175)
	}
	green := cube(0, 95, 0)
	for x := 0; x < c.W; x++ {
		top := creekTop(x)
		for y := top; y < floorTop; y++ {
			c.SetBG(x, y, water)
		}
		if scape.HashF(x, top-1, seed+71) < 0.5 {
			plot(mid, x, top-1, '"', green, 0.8)
		}
	}
	rocks := [][2]int{{12, 15}, {31, 14}, {47, 16}, {56, 15}}
	for _, r := range rocks {
		plot(near, r[0], r[1], 'O', grey(10), 1)
		plot(near, r[0]+1, r[1], 'o', grey(9), 1)
	}
	// Whitewater: how much of the creek is foam is the work. It flows right
	// at twenty columns a second through a pattern the frame's width.
	d := 0.04 + 0.26*level
	shift := int(math.Round(t * 20))
	foam := cube(175, 215, 255)
	for y := 12; y < floorTop; y++ {
		for x := 0; x < c.W; x++ {
			if y < creekTop(x) || (y >= floorTop-2 && x >= 63 && x <= 75) {
				continue // above the water, or on the owl's boulder
			}
			h := scape.HashF(((x-shift)%c.W+c.W)%c.W, y, seed+70)
			for _, r := range rocks {
				if y == r[1] && x > r[0]+1 && x <= r[0]+4 {
					h *= 0.3 // rocks always throw a little foam downstream
				}
			}
			if h < d {
				g := '~'
				if h < d*0.3 {
					g = '-'
				} else if h < d*0.6 {
					g = '^'
				}
				plot(mid, x, y, g, foam, 0.75)
			}
		}
	}
	// The owl on a boulder in the water, right.
	fill(c, 63, floorTop-2, 75, floorTop-1, grey(9))
	owlAt(c, 64, floorTop-9)
	writeBand(c, bandTop, greyBetween(4, 8, l))
}
