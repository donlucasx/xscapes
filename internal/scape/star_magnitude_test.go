package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// magRead is the constellation read off a rendered night sky: how many stars
// drew, how many distinct tones they drew in, and the faintest star's luma
// over the brightest ambient speck's, which is the count's margin.
type magRead struct {
	stars, tones int
	margin       float64
}

func readMagnitude(w, h int, seed int64, hour float64) magRead {
	sh := NewShore(seed, false)
	hy := int(float64(h) * 0.42)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 200; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, Activity{
			Level: 0.5, Working: true, ContextUsed: 0.26, TimeOfDay: hour / 24,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	luma := func(c term.RGB) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }
	r := magRead{}
	lo, dust := 1e9, 0.0
	seen := map[int]bool{}
	for y := 0; y < hy; y++ {
		for x := 0; x < w; x++ {
			ru, fg, bg := c.ResolveAt(x, y, term.Profile256)
			switch ru {
			case '*':
				r.stars++
				seen[fg.Index256()] = true
				lo = min(lo, luma(fg))
			case '.', '·', '+':
				if fg != bg {
					dust = max(dust, luma(fg))
				}
			}
		}
	}
	r.tones = len(seen)
	r.margin = lo - dust
	return r
}

// TestStarMagnitudeIsWideAndStillCounts holds his 2026-09-16 ruling ("widen
// the range for star magnitude") against the one thing the channel may not
// spend. At night, across three geometries and three seeds: every finished
// todo draws, the constellation carries at least five distinct tones, and the
// faintest star stays a full cube step above the brightest dust speck. The
// numbers are the sweep's (notes/s35-magsweep), not a guess.
func TestStarMagnitudeIsWideAndStillCounts(t *testing.T) {
	check := func(t *testing.T, wantTones int, wantMargin float64) (worstTones int, worstMargin float64) {
		worstTones, worstMargin = 1<<30, 1e9
		for _, g := range [][2]int{{125, 28}, {80, 24}, {143, 27}} {
			for _, seed := range []int64{1, 7, 42} {
				for _, hour := range []float64{0, 22} {
					r := readMagnitude(g[0], g[1], seed, hour)
					if r.stars != 19 {
						t.Errorf("%dx%d seed %d %.0fh: %d stars drawn, want 19", g[0], g[1], seed, hour, r.stars)
					}
					worstTones = min(worstTones, r.tones)
					worstMargin = min(worstMargin, r.margin)
				}
			}
		}
		return
	}
	tones, margin := check(t, 5, 10)
	t.Logf("shipped floor %.2f: worst tone count %d, worst star-over-dust margin %.1f luma", starDimmest, tones, margin)
	if tones < 5 {
		t.Errorf("worst night sky carries %d tones, want at least 5", tones)
	}
	if margin < 10 {
		t.Errorf("faintest star is only %.1f luma above the brightest dust speck, want a cube step (10)", margin)
	}

	// The test has to be able to fail in BOTH directions, or it holds nothing.
	was := SetStarDimmest(0.70)
	if tn, _ := check(t, 5, 10); tn >= 5 {
		t.Errorf("at the old floor 0.70 the worst sky still carries %d tones; the tone bar is not doing anything", tn)
	}
	SetStarDimmest(0.40)
	if _, mg := check(t, 5, 10); mg >= 10 {
		t.Errorf("at floor 0.40 the faintest star is still %.1f above the dust; the margin bar is not doing anything", mg)
	}
	SetStarDimmest(was)
}
