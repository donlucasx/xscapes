package main

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// eyeShine is the cat's, unchanged: an eye is an eye across companions, and the
// shine is the one part of the face the whole vocabulary shares.
var eyeShine = term.RGB{R: 168, G: 236, B: 176}

type crab struct {
	key, name, note string
	body            []string
	young           []string
	eyes            [2]int
	eyeRow          int
}

var crabs = []crab{
	{"broad", "Broad", "Both claws out at shell height, three leg pairs, stalk nubs. The most crab-shaped, and the widest thing that has ever sat in this box.",
		broadBody, crabYoung, [2]int{4, 7}, 1},
	{"lowrider", "Lowrider", "Flatter and wider, claws forward and low, eyes straight on the shell. The top row is empty so it sits ON the sand rather than standing on it.",
		lowriderBody, crabYoung, [2]int{3, 9}, 1},
	{"stalker", "Stalker", "Tall eyestalks over a compact shell. Two of seven rows go to the eyes, which is the cat's arrangement: the face carries the state and the body barely moves.",
		stalkerBody, crabYoung, [2]int{3, 7}, 0},
	{"fiddler", "Fiddler", "One oversized claw. The asymmetry is the identity, and the big claw is a gesture channel the cat does not have -- it can raise on its own for 'needs you' without the body changing.",
		fiddlerBody, fiddlerYoung, [2]int{4, 7}, 1},
}

// The coat candidates. Every one is an exact entry of the 256 cube, because a
// glyph colour is saturated 2.6x and snapped to the cube before it is painted
// and only off-cube choices move -- but that is measured on the page, not
// assumed here.
type coat struct {
	key, name string
	col       term.RGB
	note      string
}

var coats = []coat{
	{"salmon", "salmon", term.RGB{R: 255, G: 135, B: 135}, "the brief"},
	{"coral", "coral", term.RGB{R: 255, G: 135, B: 95}, "warmer, half a step to orange"},
	{"blush", "blush", term.RGB{R: 255, G: 175, B: 175}, "paler, closest to the cream that ships"},
	{"peach", "peach", term.RGB{R: 255, G: 175, B: 135}, "between salmon and sand"},
	{"rose", "rose", term.RGB{R: 215, G: 135, B: 135}, "deeper, muted"},
	{"shell", "shell pink", term.RGB{R: 215, G: 175, B: 175}, "dusty, nearly a neutral"},
}

// drawCrab draws one body through the companion pipeline at cell (x, y).
// Copied from notes/scapestudy's drawAnimal rather than imported -- that is a
// package main too -- and kept deliberately identical so the crabs are judged
// on the same terms as the owl and the frog were.
func drawCrab(near *canvas.Layer, rows []string, w, h int, coatCol term.RGB, eyes [2]int, eyeRow, x, y int, mirror bool) {
	bm := companion.ParseBitmap(rows)
	if mirror {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	companion.PlotRim(near, q, x, y)
	(&companion.Sprite{Rows: q, Body: coatCol}).Draw(near, x, y)
	cw := w / 2
	at := func(cell int) int {
		if mirror {
			return x + cw - 1 - cell
		}
		return x + cell
	}
	for _, e := range eyes {
		if e >= 0 {
			near.Plot(at(e), y+eyeRow, 'o', eyeShine, 1)
		}
	}
}

func shoreFrame(w, h int, tod, t float64, seed int64, draw func(c *canvas.Canvas, sh *scape.Shore)) *canvas.Canvas {
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.5, TimeOfDay: tod, ContextUsed: 0.3, TodoDone: 2, TodoTotal: 5}
	for k := 0; k < 12; k++ {
		sh.Update(c, t+float64(k)/20, act)
	}
	draw(c, sh)
	return c
}

// crabFrame lays the crab out exactly where live.go lays the cat out: the
// companion's own margin off the right edge, two rows off the bottom, the
// litter growing leftward.
func crabFrame(cr *crab, col term.RGB, w, h int, tod, t float64, seed int64, withYoung bool) (*canvas.Canvas, int, int) {
	var px, py int
	c := shoreFrame(w, h, tod, t, seed, func(c *canvas.Canvas, sh *scape.Shore) {
		near := c.Near()
		px, py = c.W-12-(2+c.W/32), c.H-2-7
		drawCrab(near, cr.body, 24, 28, col, cr.eyes, cr.eyeRow, px, py, true)
		if withYoung {
			ky := py + 7 - 4
			for i := 0; i < 2; i++ {
				drawCrab(near, cr.young, 12, 16, col, [2]int{-1, -1}, 0, px-1-6-i*7, ky, i == 0)
			}
		}
	})
	return c, px, py
}

// catFrame is the reference: the shipped cat, same place, same hour.
func catFrame(w, h int, tod, t float64, seed int64) (*canvas.Canvas, int, int) {
	var px, py int
	c := shoreFrame(w, h, tod, t, seed, func(c *canvas.Canvas, sh *scape.Shore) {
		cat := companion.NewCat()
		cat.FaceLeft(true)
		px, py = c.W-12-(2+c.W/32), c.H-2-7
		cat.Draw(c.Near(), px, py, t, companion.Working)
	})
	return c, px, py
}
