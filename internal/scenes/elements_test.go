package scenes

import (
	"fmt"
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestTheElementsStayApart walks a day of the live vista at 125x28 every
// half hour, element by element, and holds two things (his ask of
// 2026-09-16: every element "consistent and legible throughout"): no split
// glyph inside the mound or the band (the canvas's implied split drew a
// line across the mound at some hours), and by day (07:00 to 17:00) every
// seam at least its floor apart in mean luma. The floors are a little
// under what was measured after the fixes: lake/meadow ran 31-53 (it had
// been 0.3 at 08:00), meadow/mound 41-61 (11), near ridge/lake 50-66
// (18), sky/far range 43-77. The far/mid seam is by hue and snow more
// than luma and is not held here.
func TestTheElementsStayApart(t *testing.T) {
	luma := func(c term.RGB) float64 { return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B) }
	const W, H = 125, 28
	type reg struct {
		name string
		in   func(x, y int) bool
	}
	worst := map[string]struct {
		gap float64
		hr  float64
	}{}
	splitsTotal := 0
	for hh := 0.0; hh < 24; hh += 0.5 {
		tod := hh / 24
		v := NewVista(7, false)
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		v.Update(c, 3.0, scape.Activity{Working: true, Level: 0.35, ContextUsed: 0.3, TimeOfDay: tod, TodoDone: 2})
		lay := v.lay
		sx := float64(W) / 80
		W2 := W * 2
		far, mid, near := make([]int, W2), make([]int, W2), make([]int, W2)
		for u := 0; u < W2; u++ {
			far[u] = skyline(u, lay.farBase, 11, 7+11, 0.061, 40*sx, 90*sx)
			mid[u] = skyline(u, lay.midBase, 25, 7+23, 0.047, 84*sx, 56*sx)
			near[u] = skyline(u, lay.nearBase, 6, 7+31, 0.19, 60*sx, 160*sx) - int(math.Round(3.0*ridged(float64(u), 7+53, 1.7)))
		}
		row := func(x int, sk []int) int { return (sk[2*x] + 1) / 2 } // first whole row of the range at column x
		regs := []reg{
			{"sky", func(x, y int) bool { return y < row(x, far)-1 && y < lay.lakeTop-1 }},
			{"far range", func(x, y int) bool { return y >= row(x, far)+1 && y < row(x, mid)-1 && y < lay.lakeTop }},
			{"mid range", func(x, y int) bool { return y >= row(x, mid)+1 && y < row(x, near)-1 && y < lay.lakeTop }},
			{"near ridge", func(x, y int) bool { return y >= row(x, near)+1 && y < lay.lakeTop && x > 30 && x < lay.owlX-4 }},
			{"lake", func(x, y int) bool { return y >= lay.lakeTop && y < lay.meadowTop && x > 20 && x < lay.owlX-10 }},
			{"meadow", func(x, y int) bool {
				return y >= lay.meadowTop+1 && y < lay.bandTop && (x < lay.fireX-12 || x > lay.fireX+12) && x > 30 && x < lay.owlX-20
			}},
			{"mound", func(x, y int) bool {
				return x >= lay.owlX+2 && x <= lay.owlX+9 && y >= lay.bandTop-2 && y < lay.bandTop
			}},
			{"band", func(x, y int) bool { return y >= lay.bandTop }},
		}
		means := map[string]float64{}
		for _, r := range regs {
			n, sum, splits := 0, 0.0, 0
			for y := 0; y < H; y++ {
				for x := 0; x < W; x++ {
					if !r.in(x, y) {
						continue
					}
					ch, _, bg := c.ResolveAt(x, y, term.Profile256)
					if r.name == "mid range" && luma(bg) > 185 {
						continue // snow and glow, not the rock
					}
					if ch == '▀' || ch == '▄' {
						if r.name == "mound" || r.name == "band" || r.name == "lake" {
							splits++
						}
					}
					n++
					sum += luma(bg)
				}
			}
			if n > 0 {
				means[r.name] = sum / float64(n)
			}
			splitsTotal += splits
			if splits > 0 && r.name != "lake" {
				t.Errorf("%05.2fh: %d split glyphs inside the %s", hh, splits, r.name)
			}
		}
		pairs := [][2]string{{"sky", "far range"}, {"far range", "mid range"}, {"mid range", "near ridge"}, {"near ridge", "lake"}, {"lake", "meadow"}, {"meadow", "band"}, {"meadow", "mound"}, {"sky", "mid range"}}
		if hh >= 7 && hh <= 17 {
			for _, f := range []struct {
				a, b  string
				floor float64
			}{{"lake", "meadow", 25}, {"meadow", "mound", 35}, {"near ridge", "lake", 40}, {"sky", "far range", 35}} {
				if gap := math.Abs(means[f.a] - means[f.b]); gap < f.floor {
					t.Errorf("%05.2fh: %s and %s are %.1f luma apart, floor %.0f", hh, f.a, f.b, gap, f.floor)
				}
			}
		}
		if false {
			fmt.Printf("%05.2fh far/mid %5.1f  lake/meadow %5.1f  meadow/mound %5.1f  near/lake %5.1f  sky/far %5.1f  (means sky %.0f far %.0f mid %.0f near %.0f lake %.0f meadow %.0f mound %.0f band %.0f)\n", hh,
				math.Abs(means["far range"]-means["mid range"]), math.Abs(means["lake"]-means["meadow"]), math.Abs(means["meadow"]-means["mound"]), math.Abs(means["near ridge"]-means["lake"]), math.Abs(means["sky"]-means["far range"]),
				means["sky"], means["far range"], means["mid range"], means["near ridge"], means["lake"], means["meadow"], means["mound"], means["band"])
		}
		for _, p := range pairs {
			gap := math.Abs(means[p[0]] - means[p[1]])
			k := p[0] + " / " + p[1]
			if w, ok := worst[k]; !ok || gap < w.gap {
				worst[k] = struct {
					gap float64
					hr  float64
				}{gap, hh}
			}
		}
	}
	for _, k := range []string{"sky / far range", "far range / mid range", "near ridge / lake", "lake / meadow", "meadow / mound"} {
		w := worst[k]
		t.Logf("%-24s worst gap %5.1f luma at %05.2fh (splits counted in the lake are the shore's edge: %d over the day)", k, w.gap, w.hr, splitsTotal)
	}
}
