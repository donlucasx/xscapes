package scape

import (
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// No cell in the middle of the disc is drawn as a half block.
//
// A half block on Terminal.app leaves a one-pixel rule at the cell's bottom
// edge: its ink runs to 29.4 of a 30px row and the last device pixel row falls
// back to the cell's background, which is the other half's colour. At the disc's
// EDGE that is the silhouette and it is what keeps the disc round. In the
// INTERIOR it is a rule straight across the face, and he reported it three
// times -- measured off his own screen 2026-09-07 at the disc's centre column:
// twelve pixels of one tone, one pixel of a blend, then sixty more of the tone.
//
// A cell whose neighbours above and below are also disc is an interior cell, so
// if it is a half block the rule has come back.
func TestNoRuleRunsAcrossTheDiscsFace(t *testing.T) {
	prev := term.NoSplitCells
	prevHalf := term.LowerHalf
	term.NoSplitCells, term.LowerHalf = true, true
	defer func() { term.NoSplitCells, term.LowerHalf = prev, prevHalf }()

	for _, w := range []int{60, 80, 100, 126, 152} {
		for _, h := range []int{16, 20, 24, 28, 34, 40, 47} {
			// FULL PHASE, and the reason is exactness rather than convenience.
			// By day the sun wanes as a crescent and moon() SKIPS a half-row
			// falling inside the terminator, so a cell can be geometrically
			// inside the disc with only ONE half painted -- a true silhouette
			// edge, which must keep its split. At full phase the terminator is
			// clear of the face entirely, so "geometrically inside" and
			// "painted" are the same thing and the check below is exact at
			// every hour.
			for i := 0; i < 24; i++ {
				tod := float64(i) / 24
				c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				sh := NewShore(7, false)
				sh.MoonX = 0.28
				sh.Update(c, 3, Activity{Working: true, Level: 0.5, TimeOfDay: tod, ContextUsed: 0})
				if !sh.moonPainted || sh.moonRR <= 0 {
					continue
				}
				// BOTH of a cell's own half-rows inside the disc is the real
				// invariant -- the same test moon() paints by. DiscCovers is
				// true when EITHER half is inside, so using it here would call
				// the disc's own edge ring "interior" and demand it stop
				// splitting, which is exactly what keeps the disc round.
				bothHalvesIn := func(x, y int) bool {
					fx := float64(x-sh.moonX) / 2.0
					for _, off := range [2]float64{-0.25, 0.25} {
						if math.Hypot(fx, float64(y-sh.moonY)+off) >= sh.moonRR {
							return false
						}
					}
					return true
				}
				for y := 1; y < h-1; y++ {
					for x := 0; x < w; x++ {
						if !bothHalvesIn(x, y) {
							continue
						}
						r, _, _ := c.ResolveAt(x, y, term.Profile256)
						if r == '▄' || r == '▀' {
							t.Fatalf("%dx%d tod=%.2f: cell (%d,%d) has BOTH halves inside the disc and is drawn as %q -- "+
								"that is a one-pixel rule across the face", w, h, tod, x, y, string(r))
						}
					}
				}
			}
		}
	}
}
