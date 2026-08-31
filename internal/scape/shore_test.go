package scape

import (
	"math"
	"testing"

	"github.com/donlucasx/asciiscapes/internal/canvas"
)

// TestActivityChangeDoesNotTeleportTheSea is the regression test for the bug
// that only appears once activity is driven by real events.
//
// The scene ran for months on hardcoded levels, where the old wave clock --
// tt = t * (0.55 + Level*1.45) -- was indistinguishable from a correct one.
// It is wrong the moment Level moves: scaling the whole elapsed time means a
// change in activity shifts the wave field by (delta speed) * (session age),
// so the sea jumps, and jumps harder the longer the session has been running.
//
// The instrument is the waterline, not the rendered glyphs. Glyphs were the
// obvious choice and they are useless here: the sea legitimately repaints ~155
// of 1920 cells every frame at 20fps, and that noise floor is high enough to
// hide the bug -- a first version of this test passed against the broken code.
// The waterline is a sum of sines, so it moves smoothly and a phase jump
// stands out by two orders of magnitude.
func TestActivityChangeDoesNotTeleportTheSea(t *testing.T) {
	const (
		w, h  = 80, 24
		fps   = 20.0
		age   = 600.0 // ten minutes in, where the old bug was worst
		quiet = 0.30
		busy  = 0.35 // a SMALL step: one extra tool call, not a fan-out
	)

	s := NewShore(7, false)
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)

	// Run forward at a real frame rate so phase accumulates the way it does
	// live. Two far-apart timestamps would exercise the large-gap clamp
	// instead of the thing under test.
	var prev []float64
	var steady float64
	n := 0
	for i := 0; i < int(age*fps); i++ {
		s.Update(c, float64(i)/fps, Activity{Working: true, Level: quiet})
		cur := append([]float64(nil), s.lastEdge...)
		if prev != nil && i > int(age*fps)-40 {
			steady += meanAbsDiff(prev, cur)
			n++
		}
		prev = cur
	}
	if n == 0 || prev == nil {
		t.Fatal("no baseline frames measured")
	}
	baseline := steady / float64(n)

	// One more frame at the same clock rate, with activity nudged up.
	s.Update(c, age, Activity{Working: true, Level: busy})
	jump := meanAbsDiff(prev, s.lastEdge)

	t.Logf("waterline moves %.4f rows in a normal frame; %.4f rows across the activity step (%.0fx)",
		baseline, jump, jump/baseline)

	// The step legitimately raises the sea a little -- reach scales with
	// Level -- so it may move more than a quiet frame. What it must not do is
	// land the waves somewhere unrelated.
	if limit := baseline * 8; jump > limit {
		t.Errorf("activity step moved the waterline %.4f rows, over the %.4f limit (%.0fx a normal frame): "+
			"the wave clock is scaling elapsed time instead of integrating it", jump, limit, limit/baseline)
	}
}

// TestWavesKeepMovingAtConstantActivity guards the other direction: freezing
// the sea would pass the test above trivially.
func TestWavesKeepMovingAtConstantActivity(t *testing.T) {
	s := NewShore(7, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	s.Update(c, 0, Activity{Working: true, Level: 0.6})
	first := append([]float64(nil), s.lastEdge...)
	for i := 1; i <= 10; i++ {
		s.Update(c, float64(i)/20, Activity{Working: true, Level: 0.6})
	}
	if d := meanAbsDiff(first, s.lastEdge); d < 0.01 {
		t.Errorf("waterline barely moved over half a second at level 0.6: %.5f rows", d)
	}
}

func meanAbsDiff(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += math.Abs(a[i] - b[i])
	}
	return sum / float64(n)
}
