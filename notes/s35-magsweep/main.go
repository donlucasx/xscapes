// Command magsweep answers his 2026-09-16 ruling on the s28 leftover -- "widen
// the range for star magnitude" -- by rendering the product at several alpha
// floors and reading the constellation off the FRAME, at more than one seed
// and geometry, because a measurement at one seed is not a measurement.
//
// What "wider" may cost is the COUNT, the one thing this channel may not
// spend: a star dimmer than the ambient dust stops reading as a star. So
// beside the tone count and the luma spread, every floor reports the
// faintest star against the brightest dust speck in the same sky.
//
//	go run ./notes/s35-magsweep
package main

import (
	"fmt"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

type geom struct{ w, h int }

var geoms = []geom{{125, 28}, {80, 24}, {143, 27}}
var seeds = []int64{1, 7, 42}
var hours = []float64{0, 20, 22, 6}
var floors = []float64{0.70, 0.60, 0.50, 0.40, 0.30}

func luma(c term.RGB) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }

type read struct {
	stars, tones                        int
	lo, hi, minContrast, dustHi, margin float64
}

func probe(g geom, seed int64, hr float64) read {
	sh := scape.NewShore(seed, false)
	hy := int(float64(g.h) * 0.42)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 200; k++ {
		c = canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.26, TimeOfDay: hr / 24,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	r := read{lo: 1e9, hi: -1e9, minContrast: 1e9, dustHi: -1e9}
	seen := map[int]bool{}
	for y := 0; y < hy; y++ {
		for x := 0; x < g.w; x++ {
			ru, fg, bg := c.ResolveAt(x, y, term.Profile256)
			switch ru {
			case '*':
				r.stars++
				seen[fg.Index256()] = true
				l := luma(fg)
				r.lo, r.hi = min(r.lo, l), max(r.hi, l)
				r.minContrast = min(r.minContrast, l-luma(bg))
			case '.', '·', '+':
				if fg != bg {
					r.dustHi = max(r.dustHi, luma(fg))
				}
			}
		}
	}
	r.tones = len(seen)
	if r.dustHi < 0 {
		r.dustHi = 0
	}
	r.margin = r.lo - r.dustHi
	return r
}

func main() {
	fmt.Println("floor  hour   stars(min/max)  tones(min..max)  luma lo..hi (worst)  spread(min)  contrast(min)  faintest-star minus brightest-dust (min)")
	for _, fl := range floors {
		scape.SetStarDimmest(fl)
		for _, hr := range hours {
			var stMin, stMax, tnMin, tnMax = 1 << 30, 0, 1 << 30, 0
			lo, hi, spread, contrast, margin := 1e9, -1e9, 1e9, 1e9, 1e9
			for _, g := range geoms {
				for _, sd := range seeds {
					r := probe(g, sd, hr)
					stMin, stMax = min(stMin, r.stars), max(stMax, r.stars)
					tnMin, tnMax = min(tnMin, r.tones), max(tnMax, r.tones)
					lo, hi = min(lo, r.lo), max(hi, r.hi)
					spread = min(spread, r.hi-r.lo)
					contrast = min(contrast, r.minContrast)
					margin = min(margin, r.margin)
				}
			}
			fmt.Printf("%.2f   %4.0fh   %3d/%-3d          %2d..%-2d          %5.1f..%5.1f       %6.1f       %6.1f          %6.1f\n",
				fl, hr, stMin, stMax, tnMin, tnMax, lo, hi, spread, contrast, margin)
		}
		fmt.Println()
	}
	_ = sort.Ints
}
