package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// starCols is which columns carry a lit checklist star.
func starCols(t *testing.T, w, h, done int) []int {
	t.Helper()
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	sh.Update(c, 3, Activity{Level: 0.5, Working: true, ContextUsed: 0.3,
		TimeOfDay: 13.0 / 24, TodoDone: done, TodoTotal: 32})
	var xs []int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				xs = append(xs, x)
			}
		}
	}
	return xs
}

// The first few stars are spread across the sky, not piled in one corner.
//
// This is why the channel went unnoticed the first time it ever lit. Positions
// were laid out in index ORDER over a fixed 32 places, so the sky filled
// strictly left to right: at 153 columns one star landed on column 6, two on 6
// and 9, and sixteen had still not passed halfway. He saw his first lit
// constellation as two specks in the far-left corner and reported it as not
// shining at all.
func TestTheFirstStarsAreSpreadAcrossTheSky(t *testing.T) {
	const w, h = 153, 27
	for _, done := range []int{2, 3, 5} {
		xs := starCols(t, w, h, done)
		if len(xs) < 2 {
			t.Fatalf("%d stars asked for, %d drawn", done, len(xs))
		}
		lo, hi := xs[0], xs[0]
		for _, x := range xs {
			if x < lo {
				lo = x
			}
			if x > hi {
				hi = x
			}
		}
		// Half the sky is a low bar and the old layout failed it at every one
		// of these counts: 3 columns, 3 columns, 16 columns.
		if span := hi - lo; span < w/2 {
			t.Errorf("%d stars span columns %d..%d, only %d of %d wide -- they are piled up",
				done, lo, hi, span, w)
		}
	}
}

// And a star lights where it always was: the column for a given index cannot
// move as later ones arrive, or the constellation reshuffles itself every time
// the agent finishes something.
func TestAStarLightsWhereItAlwaysWas(t *testing.T) {
	const w, h = 153, 27
	first := starCols(t, w, h, 1)
	if len(first) != 1 {
		t.Fatalf("one star asked for, %d drawn", len(first))
	}
	for _, done := range []int{2, 5, 12, 32} {
		xs := starCols(t, w, h, done)
		found := false
		for _, x := range xs {
			if x == first[0] {
				found = true
			}
		}
		if !found {
			t.Errorf("with %d stars lit, the first one has moved off column %d: %v", done, first[0], xs)
		}
	}
}
