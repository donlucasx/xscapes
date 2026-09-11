package scape

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// THROWAWAY PROBE. Deleted before the work lands; it exists so the before and
// the after are measured on the RENDERED frame rather than argued about.

const probeW, probeH = 125, 28

// warm runs a single shore forward so the frame it hands back has history --
// Update clamps a >1 s gap to one nominal step, so a fresh NewShore per frame
// has none.
func probeWarm(seed int64, tod float64, frames int) (*Shore, *canvas.Canvas, float64) {
	sh := NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < frames; k++ {
		c = canvas.New(probeW, probeH, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, Activity{
			Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: tod,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	return sh, c, tm
}

func probeHY() int { h := probeH; return int(float64(h) * 0.42) }

// ambientCensus separates PLOTTED from LEGIBLE on the rendered frame.
func ambientCensus(c *canvas.Canvas) (plotted, legible int, glyphs map[rune]int, deltas []float64) {
	hy := probeHY()
	glyphs = map[rune]int{}
	for y := 0; y < hy; y++ {
		for x := 0; x < probeW; x++ {
			cell := c.Far().Cells[y*probeW+x]
			if !cell.Set {
				continue
			}
			plotted++
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			if fg != bg && r == cell.R {
				legible++
				glyphs[r]++
				deltas = append(deltas, math.Abs(bandLuma(fg)-bandLuma(bg)))
			}
		}
	}
	return
}

var probeHours = []float64{0, 2, 4, 6, 8, 10, 12, 14, 16, 17, 18, 19, 20, 20.4, 20.9, 22}

func TestZZAmbientTable(t *testing.T) {
	fmt.Printf("%-7s %8s %9s %9s %7s %7s %8s %8s\n",
		"hour", "StarVis", "plotted", "legible", "lost", "glyphs", "medD", "minD")
	for _, hr := range probeHours {
		_, c, _ := probeWarm(7, hr/24, 120)
		pl, lg, gl, ds := ambientCensus(c)
		lost := 0
		if pl > 0 {
			lost = 100 * (pl - lg) / pl
		}
		sort.Float64s(ds)
		med, mn := 0.0, 0.0
		if len(ds) > 0 {
			med, mn = ds[len(ds)/2], ds[0]
		}
		fmt.Printf("%-7.1f %8.3f %9d %9d %6d%% %7d %8.1f %8.1f\n",
			hr, PaletteAt(hr/24).StarVis, pl, lg, lost, len(gl), med, mn)
	}
}

// TestZZBlink counts specks that are drawn in one frame and GONE in the next --
// his "stars appearing and dissapearing" of 2026-09-10.
func TestZZBlink(t *testing.T) {
	fmt.Printf("%-7s %8s %12s %12s\n", "hour", "StarVis", "blinkOut/fr", "onScreen")
	for _, hr := range probeHours {
		sh := NewShore(7, false)
		hy := probeHY()
		tm := 0.0
		prev := map[int]bool{}
		outs, frames, on := 0, 0, 0
		for k := 0; k < 240; k++ {
			c := canvas.New(probeW, probeH, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 1.0 / 12 // inside.go paces at 12 fps
			sh.Update(c, tm, Activity{
				Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: hr / 24,
				TodoDone: 19, TodoTotal: 32,
			})
			cur := map[int]bool{}
			for y := 0; y < hy; y++ {
				for x := 0; x < probeW; x++ {
					if c.Far().Cells[y*probeW+x].Set {
						cur[y*probeW+x] = true
					}
				}
			}
			if k > 60 {
				for i := range prev {
					if !cur[i] {
						outs++
					}
				}
				frames++
				on += len(cur)
			}
			prev = cur
		}
		avgOn := 0.0
		if frames > 0 {
			avgOn = float64(on) / float64(frames)
		}
		fmt.Printf("%-7.1f %8.3f %12.3f %12.1f\n",
			hr, PaletteAt(hr/24).StarVis, float64(outs)/math.Max(1, float64(frames)), avgOn)
	}
}

// TestZZHeadroom asks what contrast a far-layer speck can REACH at all, per sky
// row per hour: the far layer is alpha 0.30, so white ink is worth 30% of the
// distance from the sky to white and no more.
func TestZZHeadroom(t *testing.T) {
	hy := probeHY()
	white := term.RGB{R: 255, G: 255, B: 255}
	fmt.Printf("%-7s", "hour")
	for y := 0; y < hy; y++ {
		fmt.Printf("%5d", y)
	}
	fmt.Println("   <- sky row, max |dLuma| with white ink at far alpha")
	for _, hr := range probeHours {
		_, c, _ := probeWarm(7, hr/24, 120)
		fmt.Printf("%-7.1f", hr)
		for y := 0; y < hy; y++ {
			best := 0.0
			for x := 0; x < probeW; x += 7 {
				nominal := c.BGAt(x, y)
				_, _, ground := c.ResolveAt(x, y, term.Profile256)
				seen := term.Profile256.Quantise(nominal.Blend(white, canvas.AlphaFar), true)
				if d := math.Abs(bandLuma(seen) - bandLuma(ground)); d > best {
					best = d
				}
			}
			fmt.Printf("%5.0f", best)
		}
		fmt.Println()
	}
}

// TestZZGolden dumps every '*' cell so the constellation can be compared
// byte-for-byte before and after.
func TestZZGolden(t *testing.T) {
	for _, hr := range []float64{0, 6, 12, 16, 18, 20.4, 22} {
		_, c, _ := probeWarm(7, hr/24, 120)
		var out []string
		for y := 0; y < probeH; y++ {
			for x := 0; x < probeW; x++ {
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				if r == '*' {
					out = append(out, fmt.Sprintf("%d:%d:%d:%d", x, y,
						fg.Index256Keeping(), bg.Index256Keeping()))
				}
			}
		}
		fmt.Printf("GOLDEN %.1f %v\n", hr, out)
	}
}
