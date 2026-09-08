// rulecount counts the one-pixel rules Terminal.app draws through the scape,
// and measures how LONG each one runs.
//
// The geometry is measured, not assumed. In his 8.27.52 PM crop of 2026-09-07
// the disc's centre column reads: 17 px of sky (50,95,131), 12 px of rim
// (165,137,136), ONE px of (100,113,134), then the body. That single pixel is
// 0.43*rim + 0.57*sky to within two counts on every channel -- U+2584's ink
// stops at 29.4 of the 30 px row and the last device pixel row is an alpha
// blend back to the cell's BACKGROUND. Every split cell has one. No glyph
// avoids it: the whole block's ink runs 4.6..29.4, so the bottom 0.6 px is the
// background whatever is drawn there.
//
// A single pixel is not what he sees, though -- he sees a LINE. A rule is
// perceptible when many neighbouring columns put their sliver on the same
// device row, which is what a flat edge does. So the number that matters is
// the RUN: how many cells wide the rule is, and how far the sliver's colour
// sits from the colour above and below it.
//
//	go run ./notes/rulecount               # the shipped renderer, one frame
//	go run ./notes/rulecount -sweep        # every width x rows x hour
package main

import (
	"flag"
	"fmt"
	"math"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// inkFrac is how much of the last device pixel row U+2584's ink covers.
// Measured above: the sliver is 0.43 ink over 0.57 background.
const inkFrac = 0.43

type rule struct {
	y        int
	x0, x1   int // inclusive
	sliver   term.RGB
	above    term.RGB
	below    term.RGB
	strength float64
}

func dist(a, b term.RGB) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func blend(ink, bg term.RGB, f float64) term.RGB {
	m := func(a, b uint8) uint8 { return uint8(f*float64(a) + (1-f)*float64(b) + 0.5) }
	return term.RGB{R: m(ink.R, bg.R), G: m(ink.G, bg.G), B: m(ink.B, bg.B)}
}

// topColour is the colour a cell shows in its FIRST device pixel row.
// A split cell shows its background there; so does a flat one.
func topColour(c *canvas.Canvas, x, y int) term.RGB {
	if y < 0 || y >= c.H {
		return term.RGB{}
	}
	_, _, bg := c.ResolveAt(x, y, term.Profile256)
	return bg
}

// sliverAt reports the anomaly a cell leaves in its last device pixel row,
// and how far that anomaly sits from what touches it above and below.
// ok is false when the cell draws no half block, or when the anomaly is
// invisible because everything around it is the same colour.
func sliverAt(c *canvas.Canvas, x, y int) (sl, above, below term.RGB, strength float64, ok bool) {
	ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
	switch ch {
	case '▄':
		// bg 0..17, ink 17..29.4, then the blend back to bg.
		sl, above = blend(fg, bg, inkFrac), fg
	case '▀':
		// ink 4.6..17, bg 17..30 -- the bottom of the cell is already bg, so
		// there is no anomaly at this edge. Its cost is 4.6 px at the TOP.
		return sl, above, below, 0, false
	default:
		return sl, above, below, 0, false
	}
	below = topColour(c, x, y+1)
	// The sliver only reads as a rule where it differs from BOTH sides. Where
	// the cell below carries the background's colour the sliver lands in its
	// own colour and nothing shows.
	d := math.Min(dist(sl, above), dist(sl, below))
	if d < 6 {
		return sl, above, below, 0, false
	}
	return sl, above, below, d, true
}

// rules finds every horizontal run of sliver cells in a frame.
func rules(c *canvas.Canvas) []rule {
	var out []rule
	for y := 0; y < c.H; y++ {
		x := 0
		for x < c.W {
			sl, ab, be, st, ok := sliverAt(c, x, y)
			if !ok {
				x++
				continue
			}
			r := rule{y: y, x0: x, x1: x, sliver: sl, above: ab, below: be, strength: st}
			for x+1 < c.W {
				sl2, _, _, st2, ok2 := sliverAt(c, x+1, y)
				if !ok2 || dist(sl2, sl) > 12 {
					break
				}
				x++
				r.x1 = x
				if st2 < r.strength {
					r.strength = st2
				}
			}
			out = append(out, r)
			x++
		}
	}
	sort.Slice(out, func(i, j int) bool {
		wi := float64(out[i].x1-out[i].x0+1) * out[i].strength
		wj := float64(out[j].x1-out[j].x0+1) * out[j].strength
		return wi > wj
	})
	return out
}

func frame(cols, rows int, seed int64, tod, used float64) (*canvas.Canvas, int, int) {
	c := canvas.New(cols, rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	act := scape.Activity{TimeOfDay: tod, ContextUsed: used}
	for i := 0; i < 30; i++ {
		sh.Update(c, 0.5+float64(i)/20, act)
	}
	sh.Update(c, 2, act)
	mx, my := sh.MoonPos()
	return c, mx, my
}

func main() {
	cols := flag.Int("cols", 128, "columns")
	rows := flag.Int("rows", 27, "scape rows")
	seed := flag.Int64("seed", 7, "scene seed")
	tod := flag.Float64("tod", 0.80, "time of day")
	used := flag.Float64("used", 0.30, "context used")
	sweep := flag.Bool("sweep", false, "every width x rows x half-hour")
	flag.Parse()

	term.LowerHalf, term.NoSplitCells = true, true // Terminal.app

	if !*sweep {
		c, mx, my := frame(*cols, *rows, *seed, *tod, *used)
		rs := rules(c)
		fmt.Printf("%dx%d tod=%.2f used=%.2f  disc at (%d,%d)\n", *cols, *rows, *tod, *used, mx, my)
		fmt.Printf("%d rules\n", len(rs))
		for i, r := range rs {
			if i >= 12 {
				fmt.Printf("  ... and %d more\n", len(rs)-i)
				break
			}
			tag := "        "
			if r.y >= my-4 && r.y <= my+4 && r.x1 >= mx-8 && r.x0 <= mx+8 {
				tag = "ON DISC "
			}
			fmt.Printf("  %srow %2d  cols %3d-%-3d (%2d wide)  sliver %v between %v and %v  d=%.0f\n",
				tag, r.y, r.x0, r.x1, r.x1-r.x0+1, r.sliver, r.above, r.below, r.strength)
		}
		return
	}

	type key struct{ w, h int }
	worst := map[key]int{}
	totalCells, totalRuns, longest := 0, 0, 0
	for _, w := range []int{60, 80, 100, 128, 143, 152} {
		for _, h := range []int{16, 20, 24, 27, 34, 40, 47} {
			for i := 0; i < 48; i++ {
				c, _, _ := frame(w, h, *seed, float64(i)/48, *used)
				for _, r := range rules(c) {
					n := r.x1 - r.x0 + 1
					totalRuns++
					totalCells += n
					if n > longest {
						longest = n
					}
					if n > worst[key{w, h}] {
						worst[key{w, h}] = n
					}
				}
			}
		}
	}
	fmt.Printf("%d rule runs, %d cells of rule, longest run %d cells\n", totalRuns, totalCells, longest)
	ks := make([]key, 0, len(worst))
	for k := range worst {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return worst[ks[i]] > worst[ks[j]] })
	for i, k := range ks {
		if i >= 8 {
			break
		}
		fmt.Printf("  %3dx%-2d  longest run %d cells\n", k.w, k.h, worst[k])
	}
}
