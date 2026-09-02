package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The sea and the beach are backdrop: the water says how much work, the sand
// carries the last few actions. Neither is meant to shimmer.
//
// Both ramps used to be measured from each column's own instantaneous
// waterline, so every column got its own gradient and the anchor moved with the
// swell. Measured at 120x26: a single row carried up to 87 distinct colours, and
// on a 256-colour terminal those near-identical browns quantise apart into
// olive, dusty red and grey -- which is why the beach read as blotches rather
// than sand, and would not sit still.
//
// The waterline band itself still moves. That is the tide, and it is the one
// place motion belongs.

func scapeFrames(n int) []*canvas.Canvas {
	sh := NewShore(7, false)
	var out []*canvas.Canvas
	for i := 0; i < n; i++ {
		c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, float64(i)*0.1, Activity{Level: 0.7, TimeOfDay: 0.3868})
		out = append(out, c)
	}
	return out
}

// rowTones counts the distinct background colours in each row.
func rowTones(c *canvas.Canvas) []int {
	out := make([]int, c.H)
	for y := 0; y < c.H; y++ {
		seen := map[term.RGB]bool{}
		for x := 0; x < c.W; x++ {
			seen[c.BGAt(x, y)] = true
		}
		out[y] = len(seen)
	}
	return out
}

func TestNoRowIsAConfettiOfNearIdenticalTones(t *testing.T) {
	c := scapeFrames(1)[0]
	tones := rowTones(c)
	worst, at := 0, 0
	for y, n := range tones {
		if n > worst {
			worst, at = n, y
		}
	}
	t.Logf("worst row is %d with %d distinct background tones", at, worst)
	// 87 before this was fixed. The waterline band legitimately varies per
	// column, so this is not one; it is a guard against the ramps going back
	// to being per-column.
	if worst > 40 {
		t.Errorf("row %d carries %d distinct tones, want at most 40", at, worst)
	}
}

func TestTheSeaBandsAreFlatAcrossARow(t *testing.T) {
	c := scapeFrames(1)[0]
	tones := rowTones(c)
	// The open sea sits between the horizon and the waterline band. Those rows
	// are pure depth and should be one colour each.
	busy := 0
	for y := 11; y <= 17; y++ {
		if tones[y] > 2 {
			busy++
			t.Logf("sea row %d has %d tones", y, tones[y])
		}
	}
	if busy > 0 {
		t.Errorf("%d open-sea rows are not flat; depth is a property of the row, not of the column", busy)
	}
}

func TestTheBackdropHoldsStillBetweenFrames(t *testing.T) {
	fr := scapeFrames(24)
	changed, total := 0, 0
	for i := 1; i < len(fr); i++ {
		for y := 11; y <= 17; y++ { // open sea, above the tide
			for x := 0; x < fr[i].W; x++ {
				total++
				if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
					changed++
				}
			}
		}
	}
	pct := 100 * float64(changed) / float64(total)
	t.Logf("open-sea cells changing between frames: %.2f%%", pct)
	// 58.60% before the ramps were anchored, 3.73% after -- sixteen times less
	// motion in the backdrop. What is left is the mean waterline rising and
	// falling, which is the sea's whole job: it says how hard the agent is
	// working. That must keep moving.
	if pct > 8 {
		t.Errorf("the open sea is restless again: %.2f%% of its cells change every frame, want under 8%%", pct)
	}
}
