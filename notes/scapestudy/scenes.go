package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// Every scene colour below is a cube entry, the product's rule: cube() and
// grey() refuse anything else, so a tone that would round to something
// unexpected on Terminal.app cannot get in by accident.
type rgb = term.RGB

var cubeLevels = map[int]bool{0: true, 95: true, 135: true, 175: true, 215: true, 255: true}

func cube(r, g, b int) rgb {
	if !cubeLevels[r] || !cubeLevels[g] || !cubeLevels[b] {
		panic(fmt.Sprintf("not a cube colour: %d %d %d", r, g, b))
	}
	return rgb{R: uint8(r), G: uint8(g), B: uint8(b)}
}

// grey k is xterm index 232+k: 8, 18, ... 238.
func grey(k int) rgb {
	if k < 0 || k > 23 {
		panic(fmt.Sprintf("grey %d out of range", k))
	}
	v := uint8(8 + 10*k)
	return rgb{R: v, G: v, B: v}
}

func luma(c rgb) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }

// lit is how much daylight the world has: 0 at night, 1 at noon. Read off the
// palette's own star visibility so every scene keys off the same clock the
// shore does.
func lit(p scape.Palette) float64 { return 1 - math.Max(0, math.Min(1, p.StarVis)) }

// greyBetween picks a grey ramp entry between two levels by daylight.
func greyBetween(night, day int, l float64) rgb {
	return grey(int(math.Round(float64(night) + (float64(day)-float64(night))*l)))
}

func fill(c *canvas.Canvas, x0, y0, x1, y1 int, col rgb) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if x >= 0 && y >= 0 && x < c.W && y < c.H {
				c.SetBG(x, y, col)
			}
		}
	}
}

// vramp paints one ramp top to bottom through the palette, as the shore
// paints its sky: each cell is told the span of the ramp it covers.
func vramp(c *canvas.Canvas, x0, y0, x1, y1 int, top, bot rgb) {
	r := term.NewRamp(top, bot)
	n := math.Max(1, float64(y1-y0+1))
	for y := y0; y <= y1; y++ {
		fy := float64(y - y0)
		for x := x0; x <= x1; x++ {
			if x >= 0 && y >= 0 && x < c.W && y < c.H {
				c.SetBGRamp(x, y, r, fy/n, (fy+1)/n)
			}
		}
	}
}

func plot(l *canvas.Layer, x, y int, r rune, col rgb, a float64) {
	if x >= 0 && y >= 0 && x < l.W && y < l.H {
		l.Plot(x, y, r, col, a)
	}
}

func text(l *canvas.Layer, x, y int, s string, col rgb, a float64) {
	for i, r := range []rune(s) {
		plot(l, x+i, y, r, col, a)
	}
}

// inkFor is the writing's colour on a surface: light ink on a dark page,
// dark ink on a light one, both greys the cube holds.
func inkFor(surface rgb) rgb {
	if luma(surface) > 128 {
		return grey(3)
	}
	return grey(23)
}

// sample is the demo turn's sand, newest last, the way the shore writes it.
var sample = []string{
	"search  rate.Limiter  3 files",
	"edit  internal/auth/handler.go  +18 -2",
	"shell  go  exit 1",
}

// writeBand is the accumulator slot: the last three lines of work, written on
// whatever surface the scene offers, newest brightest.
func writeBand(c *canvas.Canvas, y0 int, surface rgb) {
	fill(c, 0, y0, c.W-1, c.H-1, surface)
	ink := inkFor(surface)
	for i, s := range sample {
		a := 0.55 + 0.45*float64(i)/float64(len(sample)-1)
		text(c.Near(), 2, y0+i, s, ink, a)
	}
}

// skyWindow paints the real sky into a rectangle: the shore's own palette for
// the hour, its stars at their visibility, and a small moon for context. This
// is the "sky is the world" slot for every indoor scene.
func skyWindow(c *canvas.Canvas, x0, y0, x1, y1 int, p scape.Palette, t float64, seed int64, moonX int) {
	vramp(c, x0, y0, x1, y1, p.SkyTop, p.SkyHorizon)
	far := c.Far()
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if scape.HashF(x, y, seed) < 0.06*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.7*p.StarVis)
			}
		}
	}
	if p.MoonVis > 0.05 && moonX >= x0 && moonX+2 <= x1 && y0+2 <= y1 {
		for y := y0 + 1; y <= y0+2; y++ {
			for x := moonX; x <= moonX+2; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
	}
}

// catAt draws the shipped companion, working, at a cell.
func catAt(c *canvas.Canvas, x, y int, faceLeft bool, t float64) {
	cat := companion.NewCat()
	cat.FaceLeft(faceLeft)
	cat.Draw(c.Near(), x, y, t, companion.Working)
}

// A scene is one painter. level is the work, 0..1, carried by the motion slot.
type scene struct {
	name  string
	slots [6]string // light, sky, motion (the work), surface, accumulator, companion
	note  string
	paint func(c *canvas.Canvas, tod, t, level float64, seed int64)
}

var scenes = []scene{
	{"Hearth and cabin",
		[6]string{"the fire, and daylight through one window", "the window: the real sky, stars, the moon", "the fire's height and its sparks", "stone, floorboards, a rug", "the floorboards in front of the rug", "the cat on the rug, facing the fire"},
		"Indoors, so the clock has to come in through the window. The fire is the only warm colour the cube gives freely, which is why this scene reads so easily on 256.",
		paintHearth},
	{"Rainy window",
		[6]string{"the city's lights, the sky behind the rain", "the whole pane: the real sky over a skyline", "rain density on the glass", "glass, mullions, a painted sill", "the sill", "the cat on the inside sill"},
		"The one scene the brief already planned. The companion gap closes with the cat on the sill; the frog is the alternate. Rain is a modifier the research found on every big upload, so it may belong to other scenes too.",
		paintRain},
	{"Café table",
		[6]string{"pendant bulbs by night, the street window by day", "a window onto the street", "steam off the cup", "wood, a saucer, the wall", "the near edge of the table", "the cat sitting on the table"},
		"The warmest indoor palette the cube holds without orange: tan and dusky wood. The steam is a thin motion slot; a busier street or a second cup would carry the work better.",
		paintCafe},
	{"Aquarium",
		[6]string{"the tank's lamp by night, the room by day", "the room above the tank", "fish count and bubbles", "water, gravel, plants, the stand", "the stand", "the cat in front of the glass"},
		"Water is the work again, kept blue by the same cube constraint as the sea. The room strip is thin; the clock reads mostly as the lamp's colour.",
		paintAquarium},
	{"Snow cabin",
		[6]string{"the lit window, the moon on the snow", "the sky above the pines", "snowfall density", "snow, logs, planks", "the snow in front of the porch", "the cat on the porch"},
		"The outdoor one, so the shore's own sky and moon carry over unchanged. Snow on 256 is the grey ramp, which is honest; the window is the one warm cell.",
		paintSnow},
}

func paintHearth(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	vramp(c, 0, 0, c.W-1, 20, greyBetween(2, 6, l), greyBetween(4, 9, l))
	// The window, top left: frame, then the world through it.
	fill(c, 3, 0, 22, 9, grey(8))
	skyWindow(c, 4, 1, 21, 8, p, t, seed, 8)
	fill(c, 12, 1, 13, 8, grey(8))
	fill(c, 4, 4, 21, 4, grey(8))
	// The hearth: surround, mantel, firebox, floor.
	fill(c, 44, 6, 71, 20, grey(5))
	fill(c, 44, 6, 71, 6, grey(9))
	fill(c, 48, 9, 67, 19, grey(0))
	fill(c, 44, 20, 71, 20, grey(7))
	// The glow inside: dark red above the logs, amber at them.
	fill(c, 50, 14, 65, 15, cube(95, 0, 0))
	fill(c, 50, 16, 65, 17, cube(135, 95, 0))
	text(c.Near(), 52, 18, "============", cube(135, 95, 0), 1)
	// Flames: height by the work, on the near layer, three tones.
	near := c.Near()
	for x := 50; x <= 65; x++ {
		h := 1 + int((2+4*level)*scape.HashF(x, 0, seed+1))
		for k := 0; k < h; k++ {
			y := 17 - k
			if y < 9 {
				break
			}
			col, r := cube(255, 135, 0), '^'
			if k >= h-1 {
				col, r = cube(255, 215, 0), '\''
			} else if k > 0 {
				col, r = cube(255, 175, 0), '*'
			}
			if scape.HashF(x, k, seed+2) > 0.82 {
				r = '"'
			}
			plot(near, x, y, r, col, 1)
		}
	}
	for x := 50; x <= 65; x++ {
		if scape.HashF(x, 19, seed+3) < 0.4 {
			plot(near, x, 19, '.', cube(215, 95, 0), 0.9)
		}
		if scape.HashF(x, 7, seed+4) < 0.08*(0.5+level) {
			plot(near, x, 10+int(scape.HashF(x, 8, seed+5)*3), '\'', cube(255, 175, 0), 0.8)
		}
	}
	// The rug and the cat, facing the fire.
	fill(c, 26, 18, 43, 20, cube(135, 95, 95))
	catAt(c, 29, 13, false, t)
	writeBand(c, 21, greyBetween(3, 6, l))
}

func paintRain(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	vramp(c, 0, 0, c.W-1, 11, p.SkyTop, p.SkyHorizon)
	far, mid := c.Far(), c.Mid()
	for y := 0; y <= 11; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed) < 0.05*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.6*p.StarVis)
			}
		}
	}
	if p.MoonVis > 0.05 {
		for y := 2; y <= 3; y++ {
			for x := 12; x <= 14; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
	}
	// The skyline: buildings from the horizon down, their windows lit by night.
	fill(c, 0, 12, c.W-1, 19, greyBetween(2, 5, l))
	for x := 0; x < c.W; x++ {
		h := 3 + int(scape.HashF(x/3, 1, seed)*6)
		for y := 12 - h; y < 12; y++ {
			c.SetBG(x, y, greyBetween(2, 5, l))
		}
		for y := 12 - h; y <= 18; y++ {
			if scape.HashF(x, y, seed+9) < 0.22 {
				col, a := cube(255, 215, 135), 0.9*(1-l)+0.15
				if l > 0.6 {
					col = grey(10)
				}
				plot(far, x, y, '.', col, a)
			}
		}
	}
	// Rain, on the glass: density is the work.
	d := 0.04 + 0.16*level
	for y := 0; y <= 19; y++ {
		for x := 0; x < c.W; x++ {
			h := scape.HashF(x, y+int(t*7), seed+20)
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
	// Mullions and the frame.
	fill(c, 0, 0, 1, 19, grey(5))
	fill(c, 78, 0, 79, 19, grey(5))
	fill(c, 39, 0, 40, 19, grey(5))
	fill(c, 0, 7, 79, 7, grey(5))
	// The sill, and the cat on it, inside.
	fill(c, 0, 20, c.W-1, 20, greyBetween(7, 11, l))
	catAt(c, 64, 13, true, t)
	writeBand(c, 21, greyBetween(4, 8, l))
}

func paintCafe(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	// The wall: cream by day, tan at dusk, dark by night.
	wall := cube(95, 95, 95)
	if l > 0.66 {
		wall = cube(215, 215, 175)
	} else if l > 0.33 {
		wall = cube(175, 135, 95)
	}
	fill(c, 0, 0, c.W-1, 13, wall)
	// The street window.
	fill(c, 5, 0, 41, 10, grey(8))
	skyWindow(c, 6, 1, 40, 9, p, t, seed, 10)
	fill(c, 23, 1, 24, 9, grey(8))
	fill(c, 6, 5, 40, 5, grey(8))
	if l < 0.4 {
		for _, x := range []int{9, 19, 29, 37} {
			plot(c.Far(), x, 8, '*', cube(255, 215, 135), 0.9)
		}
	}
	// Pendant bulbs over the counter side: lit by night, off by day.
	near := c.Near()
	for x := 48; x <= 76; x += 7 {
		fill(c, x, 0, x, 2, grey(6))
		col, a := cube(255, 215, 135), 1.0
		if l > 0.5 {
			col, a = grey(12), 0.8
		}
		plot(near, x, 3, 'o', col, a)
	}
	// The table, its far edge, the cup, the saucer. Tan by day, dusky wood
	// once the wall has gone tan, so the two never share a tone.
	table := cube(175, 135, 95)
	if l <= 0.66 {
		table = cube(135, 95, 95)
	}
	fill(c, 0, 14, c.W-1, 20, table)
	fill(c, 0, 14, c.W-1, 14, grey(9))
	fill(c, 19, 20, 28, 20, grey(21))
	fill(c, 21, 16, 26, 19, grey(23))
	fill(c, 22, 16, 25, 16, cube(95, 95, 0))
	fill(c, 27, 17, 27, 18, grey(23))
	// Steam is the work: wisps rise from the cup.
	n := 1 + int(level*5)
	for k := 0; k < n; k++ {
		x := 22 + k%4
		for y := 15; y >= 15-(3+k%3)-int(level*3); y-- {
			if y < 6 {
				break
			}
			dx := int(math.Round(math.Sin(float64(y)*0.9 + float64(k) + t)))
			r := '('
			if (y+k)%2 == 0 {
				r = ')'
			}
			plot(c.Mid(), x+dx, y, r, grey(16), 0.55)
		}
	}
	catAt(c, 64, 13, true, t)
	band := cube(135, 95, 95)
	if l <= 0.66 {
		band = cube(95, 95, 95)
	}
	writeBand(c, 21, band)
}

func paintAquarium(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	// The room above the tank, then the lid with its lamp.
	fill(c, 0, 0, c.W-1, 2, greyBetween(2, 20, l))
	fill(c, 0, 3, c.W-1, 3, grey(6))
	lamp := cube(255, 215, 135)
	if l > 0.5 {
		lamp = grey(14)
	}
	fill(c, 30, 3, 49, 3, lamp)
	// The water: lamp-lit at night, daylit by day. Both ends cube blues.
	top, bot := cube(0, 95, 135), cube(0, 95, 95)
	if l > 0.66 {
		top, bot = cube(95, 175, 215), cube(0, 135, 175)
	} else if l > 0.33 {
		top, bot = cube(0, 135, 175), cube(0, 95, 135)
	}
	vramp(c, 0, 4, c.W-1, 17, top, bot)
	fill(c, 0, 0, 1, 20, grey(8))
	fill(c, 78, 0, 79, 20, grey(8))
	// Gravel and its pebbles.
	fill(c, 2, 18, 77, 20, grey(6))
	far := c.Far()
	for x := 2; x <= 77; x++ {
		for y := 18; y <= 20; y++ {
			h := scape.HashF(x, y, seed+30)
			if h < 0.25 {
				plot(far, x, y, 'o', grey(12), 0.8)
			} else if h < 0.45 {
				plot(far, x, y, '.', cube(135, 95, 95), 0.8)
			}
		}
	}
	// Plants, rooted in the gravel.
	for _, px := range []int{8, 14, 40, 47} {
		h := 4 + int(scape.HashF(px, 0, seed+31)*6)
		for k := 0; k < h; k++ {
			y := 17 - k
			plot(c.Mid(), px, y, '|', cube(0, 135, 0), 1)
			if k%2 == 1 {
				plot(c.Mid(), px-1, y, '(', cube(95, 175, 95), 1)
				plot(c.Mid(), px+1, y, ')', cube(95, 175, 95), 1)
			}
		}
	}
	// Fish are the work: how many are out.
	near := c.Near()
	n := 2 + int(level*6)
	fishCols := []rgb{cube(255, 215, 135), cube(215, 175, 135), cube(255, 175, 175)}
	for k := 0; k < n; k++ {
		x := 6 + int(scape.HashF(k, 1, seed+32)*54)
		y := 5 + int(scape.HashF(k, 2, seed+33)*11)
		s := "<><"
		if scape.HashF(k, 3, seed+34) > 0.5 {
			s = "><>"
		}
		text(near, x, y, s, fishCols[k%3], 1)
	}
	// Bubbles from the airstone.
	for k := 0; k < 2+int(level*6); k++ {
		y := 17 - k*2 - int(t)%2
		if y < 4 {
			break
		}
		r := 'o'
		if k%3 == 2 {
			r = '.'
		}
		plot(c.Mid(), 30+(k%3)-1, y, r, cube(175, 215, 255), 0.7)
	}
	catAt(c, 64, 14, true, t)
	writeBand(c, 21, greyBetween(3, 6, l))
}

func paintSnow(c *canvas.Canvas, tod, t, level float64, seed int64) {
	p := scape.PaletteAt(tod)
	l := lit(p)
	vramp(c, 0, 0, c.W-1, 11, p.SkyTop, p.SkyHorizon)
	far, mid, near := c.Far(), c.Mid(), c.Near()
	for y := 0; y <= 10; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed) < 0.045*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.7*p.StarVis)
			}
		}
	}
	if p.MoonVis > 0.05 {
		for y := 2; y <= 3; y++ {
			for x := 60; x <= 62; x++ {
				c.SetBG(x, y, term.Lerp(c.BGAt(x, y), p.Moon, p.MoonVis))
			}
		}
	}
	// Snow ground, and the far tree line at the horizon.
	vramp(c, 0, 12, c.W-1, 20, greyBetween(9, 21, l), greyBetween(7, 19, l))
	for x := 0; x < c.W; x++ {
		if scape.HashF(x, 11, seed+40) < 0.5 {
			plot(mid, x, 11, '^', cube(0, 95, 0), 1)
		}
	}
	// Two near pines, right.
	for _, cx := range []int{69, 76} {
		for r := 0; r <= 8; r++ {
			y := 6 + r
			half := r / 2
			for x := cx - half; x <= cx+half; x++ {
				plot(near, x, y, '^', cube(0, 95, 0), 1)
			}
			if r%3 == 0 {
				plot(near, cx-half, y, '*', grey(23), 0.9)
			}
		}
		plot(near, cx, 15, '|', cube(95, 95, 0), 1)
	}
	// The cabin: roof under snow, log walls, a door, one lit window, a chimney.
	for r := 0; r <= 4; r++ {
		fill(c, 30-r*4-2, 5+r, 30+r*4+2, 5+r, grey(7))
		fill(c, 30-r*4-2, 5+r, 30-r*4, 5+r, grey(23))
	}
	fill(c, 12, 10, 48, 17, cube(135, 95, 95))
	for y := 10; y <= 17; y += 2 {
		fill(c, 12, y, 48, y, cube(95, 95, 95))
	}
	fill(c, 16, 12, 19, 17, grey(3))
	win := cube(255, 215, 135)
	if l > 0.5 {
		win = cube(215, 215, 175)
	}
	fill(c, 36, 11, 43, 14, win)
	fill(c, 39, 11, 40, 14, grey(4))
	fill(c, 36, 13, 43, 13, grey(4))
	fill(c, 44, 5, 46, 8, grey(6))
	for k := 0; k < 4; k++ {
		plot(mid, 45+k%2, 4-k, '~', grey(12), 0.5)
	}
	fill(c, 10, 18, 50, 18, grey(6))
	// Snowfall is the work.
	d := 0.02 + 0.10*level
	for y := 0; y <= 19; y++ {
		for x := 0; x < c.W; x++ {
			h := scape.HashF(x, y+int(t*3), seed+41)
			if h < d {
				r := '*'
				if h < d*0.5 {
					r = '.'
				}
				plot(mid, x, y, r, grey(23), 0.85)
			}
		}
	}
	catAt(c, 22, 11, true, t)
	writeBand(c, 21, greyBetween(8, 19, l))
}
