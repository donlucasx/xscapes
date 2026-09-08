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

// No rule on the disc runs more than one cell wide.
//
// A split cell on Terminal.app always leaks about one device pixel of its
// background at the cell's bottom edge; that is the terminal, not us, and no
// glyph avoids it (the full block's ink runs 4.6..29.4 of a 30px row, so the
// last row is the background whatever is drawn). What turns that pixel into
// something he can see is the RUN -- neighbouring columns putting their sliver
// on the same device row, which is what the disc's flat top and bottom caps
// did: five cells, seventy pixels, straight across the sun. Measured in his
// 8.27.52 PM crop of 2026-09-07: 17px of sky, 12px of rim, then one pixel of
// (100,113,134), which is 0.43*rim + 0.57*sky.
//
// A one-cell rule is the silhouette's own texture at a shoulder and reads as
// an edge. Anything longer is a line. Shore.FlatCaps collapses exactly the
// cells that cannot buy roundness -- those whose neighbour's edge falls in the
// same cell on the same side -- and leaves the curve alone.
func TestNoRuleRunsAcrossTheDiscsCap(t *testing.T) {
	prev, prevHalf := term.NoSplitCells, term.LowerHalf
	term.NoSplitCells, term.LowerHalf = true, true
	defer func() { term.NoSplitCells, term.LowerHalf = prev, prevHalf }()

	// The sliver is 0.43 ink over 0.57 background, measured off his screen.
	sliver := func(fg, bg term.RGB) term.RGB {
		m := func(a, b uint8) uint8 { return uint8(0.43*float64(a) + 0.57*float64(b) + 0.5) }
		return term.RGB{R: m(fg.R, bg.R), G: m(fg.G, bg.G), B: m(fg.B, bg.B)}
	}
	far := func(a, b term.RGB) float64 {
		dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
		return math.Sqrt(dr*dr + dg*dg + db*db)
	}

	for _, w := range []int{60, 80, 100, 128, 143, 152} {
		for _, h := range []int{16, 20, 24, 27, 34, 40, 47} {
			for i := 0; i < 48; i++ {
				c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				sh := NewShore(7, false)
				sh.MoonX = 0.28
				sh.Update(c, 3, Activity{Working: true, Level: 0.5, TimeOfDay: float64(i) / 48, ContextUsed: 0})
				if !sh.moonPainted || sh.moonRR <= 0 {
					continue
				}
				// A cell rules when its sliver differs from the ink directly
				// above it AND from the cell below. Where the cell below
				// carries the background's own colour the sliver lands in its
				// own colour and nothing shows.
				rules := func(x, y int) bool {
					ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
					if ch != '▄' {
						return false
					}
					sl := sliver(fg, bg)
					_, _, below := c.ResolveAt(x, y+1, term.Profile256)
					return far(sl, fg) >= 6 && far(sl, below) >= 6
				}
				for y := 1; y < h-1; y++ {
					run := 0
					for x := 0; x < w; x++ {
						if !rules(x, y) {
							run = 0
							continue
						}
						run++
						if run > 1 {
							t.Fatalf("%dx%d tod=%.2f: %d cells of rule side by side ending at (%d,%d) -- "+
								"that is a line across the disc, not an edge", w, h, float64(i)/48, run, x, y)
						}
					}
				}
			}
		}
	}
}
