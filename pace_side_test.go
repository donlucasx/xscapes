package main

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// THE PACE MOVES THE COMPANION INWARD, on whichever side it stands.
//
// paceAt returns a NEGATIVE offset -- inward, toward the centre -- and that is
// inward only for a companion anchored on the RIGHT. The unmirrored layout
// anchors it at column 5 on the LEFT, so the same offset walked it toward the
// frame edge instead of away from it.
//
// ⚠ WHAT I COULD NOT REPRODUCE, stated so nobody trusts it again: the note this
// was carded from said the companion reached COLUMN -1 at 124 columns. It does
// not. Measured across 80, 124 and 153 columns and every step count, the worst
// unmirrored position before the fix was catX = 1 -- pressed against the edge,
// never off it. The defect is the DIRECTION, not a crash off the frame, and a
// test written to the carded symptom would have been a test that cannot fail.
//
// ⚠ AND IT GOES THROUGH compose() AND pace(). An earlier version applied the
// sign flip itself and asserted its own arithmetic, which passed with the fix
// deleted from live.go.
func TestThePaceMovesInwardOnBothSides(t *testing.T) {
	cw, _ := companion.New(companion.DefaultName).Size()
	for _, w := range []int{40, 60, 80, 107, 124, 143, 153} {
		for _, mirror := range []bool{true, false} {
			lay := compose(w, cw, mirror)
			centre := float64(w) / 2
			home := float64(lay.CatX) + float64(cw)/2
			moved := 0
			for steps := 1; steps <= 24; steps++ {
				st := reduce.State{Steps: steps, StepAge: 0.1, Pose: companion.Working}
				dx := paceOffset(st, lay, w)
				if dx == 0 {
					continue
				}
				moved++
				at := home + dx
				// Inward means closer to the middle of the frame than home.
				if abs(at-centre) > abs(home-centre)+0.001 {
					t.Errorf("w=%d mirror=%v steps=%d: the companion moved OUTWARD -- home is "+
						"%.1f from the centre, it walked to %.1f",
						w, mirror, steps, abs(home-centre), abs(at-centre))
				}
				if at-float64(cw)/2 < 0 {
					t.Errorf("w=%d mirror=%v steps=%d: drawn at column %.0f", w, mirror, steps,
						at-float64(cw)/2)
				}
			}
			if moved == 0 {
				t.Errorf("w=%d mirror=%v: the companion never moved, so nothing was tested",
					w, mirror)
			}
			// The sand has to leave room for wherever it can walk to.
			if !mirror && lay.CatX+cw+lay.PaceSpan > lay.SandFrom {
				t.Errorf("w=%d: the companion can reach column %d but the sand starts at %d",
					w, lay.CatX+cw+lay.PaceSpan-1, lay.SandFrom)
			}
		}
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
