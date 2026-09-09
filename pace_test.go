package main

import (
	"math"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// The companion paces INWARD, one step per MAIN-THREAD tool event.
//
// His two notes are the spec. "Should be tied up to an actual agent action so
// its not random" -- so the step count comes from events, never from a clock.
// And "the companion should move towards the center of the screen instead of
// the far edge (right) where theres barely any margin" -- so the offset is
// never positive, because outward is a wall.
func TestThePaceGoesInwardAndIsDrivenByEvents(t *testing.T) {
	for _, w := range []int{40, 80, 124, 200} {
		span := paceSpan(w)
		if span < 1 {
			t.Fatalf("w=%d: span %d leaves nowhere to pace", w, span)
		}
		// Every step of a long walk, at every phase of the stride.
		for n := 0; n < 4*span+3; n++ {
			for _, since := range []float64{0, stepDur / 3, stepDur * 0.99, stepDur, 5} {
				dx, _ := paceAt(n, since, span)
				if dx > 0 {
					t.Fatalf("w=%d n=%d since=%.2f: dx=%.2f is OUTWARD, into the margin", w, n, since, dx)
				}
				if dx < -float64(span) {
					t.Fatalf("w=%d n=%d: dx=%.2f goes past the reserved strip (%d)", w, n, dx, span)
				}
			}
		}
		// A triangle comes home: at 2*span steps it is back where it started.
		if dx, _ := paceAt(2*span, stepDur, span); dx != 0 {
			t.Errorf("w=%d: after a full there-and-back the companion is at %.2f, not home", w, dx)
		}
	}
}

// The sand and the litter keep out of the strip the companion walks.
func TestNothingIsDrawnInThePacingStrip(t *testing.T) {
	for _, w := range []int{40, 80, 124, 200} {
		lay := compose(w, 12, true)
		reach := lay.CatX - lay.PaceSpan
		if lay.SandTo >= reach {
			t.Errorf("w=%d: the sand runs to %d but the companion reaches %d", w, lay.SandTo, reach)
		}
	}
}

// A pose that holds its ground does not pace. A raised claw that is also
// walking is noise, and "done" is held still so it cannot read as another ask.
func TestTheHeldPosesDoNotPace(t *testing.T) {
	for _, st := range []reduce.State{
		{Pose: companion.NeedsYou, Steps: 9, StepAge: 0.1},
		{Pose: companion.Done, Steps: 9, StepAge: 0.1},
		{Pose: companion.Worried, Steps: 9, StepAge: 0.1},
	} {
		if dx, moving := pace(st, 124); dx != 0 || moving {
			t.Errorf("pose %v paced to %.2f (moving=%v); it should hold its ground", st.Pose, dx, moving)
		}
	}
	// And a working companion with the same numbers does move, so the test
	// above cannot pass by the pace being broken everywhere.
	if dx, _ := pace(reduce.State{Pose: companion.Working, Steps: 9, StepAge: 0.1}, 124); dx == 0 {
		t.Error("a working companion did not pace at all")
	}
}

// SUBAGENT tool events do not move the parent. Those belong to the litter, and
// one event spending two channels is exactly what the encoding rule forbids.
func TestOnlyMainThreadToolEventsStepTheCompanion(t *testing.T) {
	now := time.Now()
	r := reduce.New("test")
	main := event.Event{Kind: event.ToolStart, ID: "a", Tool: "Edit"}
	sub := event.Event{Kind: event.ToolStart, ID: "b", Tool: "Edit", Agent: "sub-1"}

	r.Apply(main, now)
	r.Apply(sub, now)
	r.Apply(sub, now)
	r.Apply(sub, now)
	got := r.State(now).Steps
	if got != 1 {
		t.Fatalf("one main-thread event and three subagent events gave %d steps, want 1 -- "+
			"a subagent moving the parent spends the same event on two channels", got)
	}

	// And the age is measured from the last MAIN-thread one.
	st := r.State(now.Add(2 * time.Second))
	if math.Abs(st.StepAge-2) > 0.01 {
		t.Errorf("StepAge %.2f, want ~2s since the main-thread event", st.StepAge)
	}
}
