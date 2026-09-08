package main

import (
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// A crab does not swim. It walks in until the water closes over the shell and
// the eyestalks carry on above the surface, which is both what a real one does
// and the thing that reads at five cells by two: the kitten's swim sprite is a
// head and shoulders, and the crab's is a shell crest with two periscopes.
//
// Same box as KittenSwim -- 10 x 8 px, five cells by two -- so a crablet takes
// exactly the room in a lane that a kitten takes, and the shipped lane logic
// would need no new numbers.
var crabletSwim = []string{
	"..#....#..",
	"..#....#..",
	"..#....#..",
	"..######..",
	".########.",
	".########.",
	"..######..",
	"...####...",
}

// crabletEyes is where the sockets land, in CELL columns, for each crablet
// sprite. Written down rather than derived from the sprite width -- the same
// trap kittens.go records hitting, where a formula was right for two scales and
// wrong for the third.
var crabletEyes = map[string][2]int{
	"sand": {2, 3},
	"swim": {1, 3},
}

var crabletEyeCol = term.RGB{R: 150, G: 214, B: 158} // dimmer than the parent's, as the kittens' is

// swims mirrors the shipped rule exactly: keyed on index so a subagent cannot
// climb in and out of the sea between frames, and roughly one in three.
func swims(i int, seed int64) bool { return i > 0 && companion.HashF(i, 11, seed) > 0.64 }

type pendEye struct {
	x0, x1, y int
	glyph     rune
}

// drawLitter draws the crablets that stayed on the sand, WITH eyes. Bodies
// first and faces after, so a neighbour's seam cannot take an eye -- the defect
// kittens.go was fixed for.
func drawLitter(near *canvas.Layer, sh *shape, col term.RGB, px, py int, idx []int, t float64, seed int64) {
	ky := py + 7 - 4
	var pend []pendEye
	for n, i := range idx {
		bm := companion.ParseBitmap(sh.crablet)
		if n%2 == 0 {
			bm = bm.Mirrored()
		}
		q := bm.ToQuadrant()
		x := px - 1 - 6 - n*7
		companion.PlotRim(near, q, x, ky)
		(&companion.Sprite{Rows: q, Body: col}).Draw(near, x, ky)
		e := crabletEyes["sand"]
		glyph := 'o'
		// Its own blink period, not just its own phase, so a litter never falls
		// into step -- the thing that makes a row of them look alive.
		if math.Mod(t+companion.HashF(i, 12, seed)*6.283, 4.2+companion.HashF(i, 16, seed)*3) < 0.17 {
			glyph = '-'
		}
		pend = append(pend, pendEye{x + e[0], x + e[1], ky, glyph})
	}
	for _, p := range pend {
		near.Plot(p.x0, p.y, p.glyph, crabletEyeCol, 1)
		near.Plot(p.x1, p.y, p.glyph, crabletEyeCol, 1)
	}
}

// drawSwimmers puts the rest in the water. The bob happens INSIDE the sprite
// box by shifting the source rows down, exactly as the kittens do it: moving
// the whole sprite would put it in the lane above, and the crest clipping off
// the bottom is what reads as the shell going under.
func drawSwimmers(near *canvas.Layer, col term.RGB, idx []int, w, seaTop, seaBot int, t float64, seed int64) int {
	if len(idx) == 0 {
		return 0
	}
	base := companion.ParseBitmap(crabletSwim)
	kw, kh := base.W/2, base.H/4
	_ = base
	top := seaTop + 1
	lanes := (seaBot - top) / kh
	if lanes < 1 {
		return 0
	}
	var pend []pendEye
	drawn := 0
	used := map[int]bool{}
	for _, i := range idx {
		lane := int(companion.HashF(i, 13, seed) * float64(lanes))
		y := top + lane*kh
		// Swimmers never share columns: two overlapping sprites interleave into
		// one broken shape, which is a defect he reported on the kittens.
		x := 2 + int(companion.HashF(i, 14, seed)*float64(w-kw-4))
		for k := 0; k < 24 && used[x/kw]; k++ {
			x = 2 + (x+kw+3)%(w-kw-4)
		}
		used[x/kw] = true

		phase := companion.HashF(i, 12, seed) * 6.283
		sink := 0
		if math.Sin(t*0.9+float64(x)*0.16+phase) > 0.35 {
			sink = 2
		}
		// The sink is applied to the SOURCE ROWS before parsing rather than to
		// the parsed bitmap: reading a pixel back out needs an unexported
		// accessor, and shifting the strings needs nothing.
		src := crabletSwim
		if sink > 0 {
			shifted := make([]string, 0, len(src))
			for k := 0; k < sink; k++ {
				shifted = append(shifted, strings.Repeat(".", len(src[0])))
			}
			shifted = append(shifted, src[:len(src)-sink]...)
			src = shifted
		}
		rows := companion.ParseBitmap(src).ToQuadrant()
		companion.PlotRim(near, rows, x, y)
		(&companion.Sprite{Rows: rows, Body: col, Alpha: 1}).Draw(near, x, y)
		// The waterline either side of the shell.
		near.Plot(x-1, y+kh-1, '~', col, 0.5)
		near.Plot(x+kw, y+kh-1, '~', col, 0.5)
		e := crabletEyes["swim"]
		glyph := 'o'
		if math.Mod(t+phase, 4.2+companion.HashF(i, 16, seed)*3) < 0.17 {
			glyph = '-'
		}
		// The stalks stay up even as the shell sinks, so the eyes do not move
		// with the bob: that is the whole idea of the pose.
		pend = append(pend, pendEye{x + e[0], x + e[1], y, glyph})
		drawn++
	}
	for _, p := range pend {
		near.Plot(p.x0, p.y, p.glyph, crabletEyeCol, 1)
		near.Plot(p.x1, p.y, p.glyph, crabletEyeCol, 1)
	}
	return drawn
}

// split divides n subagents into those on the sand and those in the water, by
// the shipped rule.
func split(n int, seed int64) (sand, water []int) {
	for i := 0; i < n; i++ {
		if swims(i, seed) {
			water = append(water, i)
		} else {
			sand = append(sand, i)
		}
	}
	return
}

// drawExits draws one crablet per FINISHED subagent, on its way out.
//
// The cat's exit swims along the surface toward the far edge and fades. A crab
// cannot do that and should not: it walks sideways into the surf and goes
// under, so the leaving is carried by the SHELL SINKING as well as by travel.
// The stalks are the last thing to go, which is the same pose the swim state is
// built on, read in the one direction that means "gone".
//
// Progress runs 0 to 1 and the reducer owns it; the litter count itself dropped
// the moment the end event arrived. Position and depth carry the leaving, so a
// still frame reads it -- the encoding rule again.
func drawExits(near *canvas.Layer, col term.RGB, exits []float64, px, w, seaTop, seaBot int, t float64, seed int64) int {
	if len(exits) == 0 {
		return 0
	}
	base := companion.ParseBitmap(crabletSwim)
	kw, kh := base.W/2, base.H/4
	top := seaTop + 1
	lanes := (seaBot - top) / kh
	if lanes < 1 {
		return 0
	}
	var pend []pendEye
	drawn := 0
	used := map[int]bool{}
	for i, p := range exits {
		if p >= 1 {
			continue // gone
		}
		lane := int(companion.HashF(i, 17, seed) * float64(lanes))
		for k := 0; k < lanes && used[lane]; k++ {
			lane = (lane + 1) % lanes
		}
		used[lane] = true
		y := top + lane*kh

		// From the water beside the litter out toward the far edge -- leftward,
		// because the companion sits on the right and the far edge is the one
		// the agent's own text occupies.
		startX := px - 10
		endX := -kw
		x := startX + int(float64(endX-startX)*p)
		if x+kw < 0 {
			continue
		}

		// The shell goes under as it goes out. Eight source rows is the whole
		// sprite, so by p=1 there is nothing above the line.
		sink := int(p * 8)
		if sink > 8 {
			sink = 8
		}
		src := crabletSwim
		if sink > 0 {
			if sink >= len(src) {
				continue
			}
			shifted := make([]string, 0, len(src))
			for k := 0; k < sink; k++ {
				shifted = append(shifted, strings.Repeat(".", len(src[0])))
			}
			shifted = append(shifted, src[:len(src)-sink]...)
			src = shifted
		}
		// Receding over the second half, as the cat's exits do.
		alpha := 1.0
		if p > 0.5 {
			alpha = 1.0 - (p-0.5)*1.6
		}
		if alpha < 0.15 {
			alpha = 0.15
		}
		rows := companion.ParseBitmap(src).ToQuadrant()
		companion.PlotRim(near, rows, x, y)
		(&companion.Sprite{Rows: rows, Body: col, Alpha: alpha}).Draw(near, x, y)
		near.Plot(x-1, y+kh-1, '~', col, 0.5*alpha)
		near.Plot(x+kw, y+kh-1, '~', col, 0.5*alpha)

		// The eyes go last, and only while the stalks are still above the line.
		if sink < 6 {
			e := crabletEyes["swim"]
			pend = append(pend, pendEye{x + e[0], x + e[1], y, 'o'})
		}
		drawn++
	}
	for _, p := range pend {
		near.Plot(p.x0, p.y, p.glyph, crabletEyeCol, 1)
		near.Plot(p.x1, p.y, p.glyph, crabletEyeCol, 1)
	}
	return drawn
}
