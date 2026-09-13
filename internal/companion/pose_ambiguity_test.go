package companion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// A POSE HAS TO SURVIVE A SCREENSHOT, and that is a different claim from "the
// art is different".
//
// Three instruments were tried before this one and two of them lied:
//
//   - Diffing the BITMAPS proves a thing equals itself. The cat's shipped ask
//     WAS CatBody, so the honest answer was always zero.
//   - Diffing two frames at the same clock t measures the poses' breath and
//     wag being out of PHASE -- working breathes at 2.2 s, asking at 1.6 s --
//     and reports a rate difference as though a still frame could see it. It
//     said "37 cells differ" about two poses that were the same bitmap.
//
// The question a screenshot actually asks is whether the picture is AMBIGUOUS:
// is this frame one the animal also draws while simply working? So the measure
// is a SET OVERLAP over every distinct picture each pose can produce.
//
// Eyes are excluded on purpose. Two cells of colour out of eighty-four is not
// what a glance reads, and relying on them is exactly the defect these poses
// were drawn to fix.
func poseFrames(c *Cat, st State, samples int, eyes [2]int, eyeRow int) map[string]bool {
	out := map[string]bool{}
	cw, ch := c.Size()
	for i := 0; i < samples; i++ {
		cv := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Draw(cv.Near(), 2, 1, float64(i)/20, st)
		var b strings.Builder
		for y := 0; y < cv.H; y++ {
			for x := 0; x < cv.W; x++ {
				isEye := false
				if y == 1+eyeRow {
					for _, e := range eyes {
						if x == 2+cw-1-e {
							isEye = true
						}
					}
				}
				if isEye {
					b.WriteString("EYE;")
					continue
				}
				r, fg, _ := cv.ResolveAt(x, y, term.Profile256)
				fmt.Fprintf(&b, "%c%d;", r, fg.Index256())
			}
		}
		out[b.String()] = true
	}
	return out
}

// Mutation-tested, and the results are worth writing down:
//
//   - Reverting the cat's ask and finish to the working body reds this with
//     "88 of 91" and "2 of 2" -- the exact figures the design round measured.
//   - The crab's finish is DOUBLY encoded: reverting the claws alone still
//     passes, and reverting the lower body alone still passes. Only reverting
//     BOTH reproduces "2 of 2". So the claw twitch is not what makes the finish
//     readable -- it is a liveness cue, as intended -- and the settled lower
//     body is doing real work on its own.
func TestAskAndDoneSurviveAStillFrame(t *testing.T) {
	const samples = 1200
	for _, arm := range []struct {
		name    string
		make    func() *Cat
		eyes    [2]int
		eyeRow  int
		checked []State
	}{
		{"cat", NewCat, catEyeCells, 2, []State{NeedsYou, Done}},
		{"crab", NewCrab, crabEyeCells, crabEyeRow, []State{NeedsYou, Done}},
	} {
		c := arm.make()
		c.FaceLeft(true)
		work := poseFrames(c, Working, samples, arm.eyes, arm.eyeRow)
		if len(work) < 2 {
			t.Fatalf("%s: only %d distinct working pictures -- the instrument is not sampling motion",
				arm.name, len(work))
		}
		for _, st := range arm.checked {
			got := poseFrames(c, st, samples, arm.eyes, arm.eyeRow)
			shared := 0
			for k := range got {
				if work[k] {
					shared++
				}
			}
			if shared != 0 {
				t.Errorf("%s/%s: %d of %d of its BODY pictures are also working pictures -- "+
					"a screenshot cannot tell them apart, which is the defect the pose exists "+
					"to fix", arm.name, st, shared, len(got))
			}
		}
		// The positive control: Worried has had its own art all along and must
		// come out unambiguous too. If it does not, the instrument is broken
		// rather than the art.
		wor := poseFrames(c, Worried, samples, arm.eyes, arm.eyeRow)
		shared := 0
		for k := range wor {
			if work[k] {
				shared++
			}
		}
		if shared != 0 {
			t.Fatalf("%s: the CONTROL failed -- worried shares %d of %d pictures with working, "+
				"so this test is measuring the instrument and not the art", arm.name, shared, len(wor))
		}
	}
}
