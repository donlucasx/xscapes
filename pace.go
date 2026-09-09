package main

import (
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// The companion PACES, and every step is one main-thread tool event.
//
// His correction, and it is the right one: "they could be pacing a bit. Should
// be tied up to an actual agent action so its not random." A random drift is
// decoration, and decoration next to a scene where everything else means
// something is worse than stillness: it teaches the eye to ignore the companion.
//
// It also obeys the encoding rule, which a wander could not. The rule forbids
// encoding in RATE -- activity was first bound to wave speed and idle became
// indistinguishable from flat out. One step per event is a COUNT, and where the
// companion ends up is a POSITION, and both survive a screenshot. Nothing about
// the pace itself varies: a step is always one cell and always takes stepDur.
// What varies is how many steps there are, because that is how many things the
// agent did.
//
// And his earlier ask falls out for free rather than needing a rule of its own:
// an agent working ALONE makes all its tool calls on the main thread, so the
// companion steps constantly; one orchestrating six subagents makes far fewer
// itself, because the work has moved to the litter, so it settles down.
//
// Designed and measured in notes/charstudy/crab/pace.go; this is that study
// shipped, not a re-derivation.

// stepDur is how long one cell takes. CONSTANT, and that is the point.
const stepDur = 0.28

// paceSpan is how far the companion paces, in cells INWARD from home.
//
// His note on the mockup: "the companion should move towards the center of the
// screen instead of the far edge (right) where theres barely any margin." The
// companion's own margin is 2 + w/32, four or five columns at his window, so
// pacing outward is pacing into a wall and the animal spends half its time
// pressed against the frame. Inward there is room, but it is not free: the sand
// is written out to the companion's column and the litter grows leftward from
// it, so compose() reserves this strip from both.
func paceSpan(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

// paceAt is the offset from home after n steps, mid-step interpolation
// included, and whether a foot is currently down. The offset is NEGATIVE:
// inward, toward the centre of the screen.
//
// The path is a triangle -- in to the far end of the strip, then back out to
// home. Pacing, not a walk-off: it never leaves, and over a long session it
// averages out at home rather than drifting to one side.
func paceAt(steps int, sinceStep float64, span int) (dx float64, moving bool) {
	tri := func(n int) float64 {
		if span <= 0 {
			return 0
		}
		p := n % (2 * span)
		if p < 0 {
			p += 2 * span
		}
		if p > span {
			p = 2*span - p
		}
		return float64(p)
	}
	to := -tri(steps)
	if sinceStep >= stepDur || steps == 0 {
		return to, false
	}
	from := -tri(steps - 1)
	return from + (to-from)*(sinceStep/stepDur), true
}

// stillFor says a pose holds its ground. A raised claw that is also pacing is
// noise; "done" is held STILL on purpose so it cannot be misread as another
// ask; and worried persists, so it should look planted.
func stillFor(st reduce.State) bool {
	switch st.Pose {
	case companion.NeedsYou, companion.Done, companion.Worried:
		return true
	}
	return false
}

// pace is the companion's offset and stride for this frame.
func pace(st reduce.State, w int) (dx float64, moving bool) {
	if stillFor(st) {
		return 0, false
	}
	return paceAt(st.Steps, st.StepAge, paceSpan(w))
}
