package main

import (
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// ---------------------------------------------------------------------------
// THE COMPOSITION FACES BOTH WAYS, AND THE CLAMPS HAVE TO KNOW WHICH.
//
// ⭐ THE FIRST VERSION OF THE NEAR POSE'S ROOM-MAKING ERASED THE ENTIRE
// ACTIVITY TAIL IN THE LEFT-ANCHORED LAYOUT, and it did it with the feature
// switched OFF.
//
// The shipped composition is MIRRORED (locked 2026-08-31): the companion is
// against the right edge and the sand and the litter grow leftward away from
// it, so "keep clear of the companion" is a MINIMUM on the far edge. Under
// `-mirror=false` every one of those terms reverses -- CatX is 5, PaceSpan is
// never set so it is zero, SandFrom sits just right of the companion and SandTo
// is the far frame edge, and DrawKittens starts the sitters at px + Size().w + 1
// instead of px - 1. A minimum there does not clear the companion, it walks
// into it: min(SandTo, catX-1) drove SandTo BELOW SandFrom and drawSand painted
// nothing at all.
//
// Measured, pristine HEAD against the working tree over 7,560 frames with the
// near pose off: all 3,780 unmirrored frames differed, 2,088 of them with the
// companion standing at home where even the pace fix changes nothing. After the
// fix, 0 unexplained.
//
// ⚠ WHY NO EXISTING TEST CAUGHT IT: near_test.go's renderNear hardcodes
// FaceLeft(true), because that is the shipped layout and it is where the whole
// feature was designed. Every guarantee in this build was stated against one
// facing. These run both.
// ---------------------------------------------------------------------------

// renderFacing is renderNear with the facing as an argument, and a tail long
// enough to run the sand all the way out to its budget at every width here --
// a short tail would clear the companion by accident and prove nothing.
func renderFacing(t *testing.T, name string, w, h, rung int, mirror bool, st reduce.State) (*canvas.Canvas, [3]int, int) {
	t.Helper()
	old := companion.Near
	companion.Near = rung
	defer func() { companion.Near = old }()

	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(7, false)
	cat := companion.New(name)
	cat.FaceLeft(mirror)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, mirror)
	sh.MoonX = lay.MoonX
	for i := 0; rung > 0 && i < 60; i++ {
		cat.Approach(0.05, true)
	}
	bdx, bw, bh := cat.DrawnBox(w)
	sh.Update(c, 2.0, st.Act)
	top := c.H - 2 - chh
	drawScene(c, sh, cat, lay, st, 2.0, 7, top)
	return c, [3]int{bdx, bw, bh}, top
}

func longTailState(pose companion.State, steps int, kittens int) reduce.State {
	st := askState(pose, steps, 0.5, kittens)
	long := "read  internal/companion/crab_near.go  612 lines and this line keeps going well " +
		"past the margin so the sand fills its whole budget at every width under test"
	st.Tail = []reduce.Line{
		{Text: long},
		{Text: long, Age: 0.3},
		{Text: long, Age: 0.6},
	}
	return st
}

// The sand is still written in BOTH compositions, at every rung.
//
// The assertion is deliberately crude -- "the tail put cells on the frame" --
// because the failure it exists to catch was not subtle: the sand's budget went
// negative and the whole channel went dark. A crude direct reading of the thing
// you care about beats a clever proxy; that is the lesson of s13.
func TestTheSandIsWrittenInBothCompositions(t *testing.T) {
	for _, name := range []string{"cat", "crab"} {
		for _, mirror := range []bool{true, false} {
			for _, rung := range []int{0, 1, 2} {
				for _, w := range hisWidths {
					h := hisHeights[w]
					for _, steps := range []int{0, 3, 6} {
						st := longTailState(companion.NeedsYou, steps, 4)
						withTail, _, _ := renderFacing(t, name, w, h, rung, mirror, st)
						bare := st
						bare.Tail = nil
						noTail, _, _ := renderFacing(t, name, w, h, rung, mirror, bare)

						a, b := inked(withTail), inked(noTail)
						cells := 0
						for p, r := range a {
							if b[p] != r {
								cells++
							}
						}
						if cells == 0 {
							t.Errorf("%s mirror=%v rung=%d w=%d steps=%d: the activity tail put ZERO cells "+
								"on the frame -- the sand's budget collapsed and the channel went dark",
								name, mirror, rung, w, steps)
						}
					}
				}
			}
		}
	}
}

// And nothing the sand writes lands under the companion, in either composition.
//
// drawSand runs LAST and Plot REPLACES a cell, so a tail that reached the
// companion would punch its own letters straight through the body rather than
// losing to it. This is the guarantee the clamp exists for; the test above is
// the guarantee that the clamp did not buy it by deleting the sand.
func TestTheSandNeverWritesUnderTheCompanionInEitherComposition(t *testing.T) {
	for _, name := range []string{"cat", "crab"} {
		for _, mirror := range []bool{true, false} {
			for _, rung := range []int{0, 1, 2} {
				for _, w := range hisWidths {
					h := hisHeights[w]
					for _, steps := range []int{0, 3, 6} {
						st := longTailState(companion.NeedsYou, steps, 4)
						withTail, box, top := renderFacing(t, name, w, h, rung, mirror, st)
						bare := st
						bare.Tail = nil
						noTail, _, _ := renderFacing(t, name, w, h, rung, mirror, bare)

						a, b := inked(withTail), inked(noTail)
						// The companion's drawn box, in absolute columns. catX
						// is recovered from the box rather than recomputed, so
						// this cannot drift from drawScene's own arithmetic.
						lay := compose(w, 12, mirror)
						dx, _ := pace(st, w)
						catX := lay.CatX + int(math.Round(dx))
						lo := catX + box[0]
						hi := lo + box[1] - 1
						for p, r := range a {
							if b[p] == r {
								continue // not a sand cell
							}
							x, y := p[0], p[1]
							if y >= top && y < top+box[2] && x >= lo && x <= hi {
								t.Errorf("%s mirror=%v rung=%d w=%d steps=%d: the sand wrote %q at (%d,%d), "+
									"inside the companion's drawn box columns %d..%d",
									name, mirror, rung, w, steps, r, x, y, lo, hi)
								break
							}
						}
					}
				}
			}
		}
	}
}
