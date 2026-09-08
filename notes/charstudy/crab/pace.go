package main

import (
	"fmt"
	"math"
)

// Hero, mid-stride: the legs swap phase and the body drops a pixel. Only the
// LOWER half changes, so a step costs one extra bitmap and nothing else.
var heroLowerStep = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	".##...############...##.",
	".##...############...##.",
	"##.....##########.....##",
	"##.....##########.....##",
	".#......########......#.",
	".#......########......#.",
	"##.......######.......##",
	"##.......######.......##",
	".#........####........#.",
	".#....................#.",
	"..#..................#..",
}

// The companion PACES, and every step is one main-thread tool event.
//
// His correction, and it is the right one: "they could be pacing a bit. Should
// be tied up to an actual agent action so its not random." A random drift is
// decoration, and decoration next to a scene where everything else means
// something is worse than stillness -- it teaches the eye to ignore the
// companion.
//
// It also settles the encoding rule, which a wander could not. The rule forbids
// encoding in RATE: activity was first bound to wave speed and idle became
// indistinguishable from flat out. One step per event is a COUNT, and where the
// crab ends up is a POSITION, and both survive a screenshot. Nothing about the
// pace itself varies -- a step is always one cell and always takes the same
// time. What varies is how many steps there are, because that is how many
// things the agent did.
//
// And his earlier ask falls out for free rather than needing a rule of its own:
// an agent working ALONE makes all its tool calls on the main thread, so the
// crab steps constantly. An agent orchestrating six subagents makes far fewer
// itself -- the work has moved to the crablets -- so it settles down. More
// active when it is on its own, without anything special-cased to say so.
//
// Subagent events do not move it. Those belong to the crablets.

// paceSpan is how far it paces, in cells INWARD from home -- toward the centre
// of the screen, not toward the frame edge.
//
// His note, looking at the mockup: "the companion should move towards the
// center of the screen instead of the far edge (right) where theres barely any
// margin." He is right, and the first version had it backwards. The companion's
// own margin is 2 + w/32 -- four or five columns at his window -- so pacing
// outward is pacing into a wall, and the animal spends half its time pressed
// against the edge of the frame.
//
// Inward there is real room, but it is not free: the activity tail is written
// on the bottom rows out to the companion's column and draws LAST, and the
// litter grows leftward from the same column. So the strip is RESERVED from
// both -- the sand's right bound and the litter's origin both pull in by
// paceSpan. That costs the tail a handful of columns and nothing else; it
// survives to 30 columns and this is six at his window.
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

const stepDur = 0.28 // seconds to cross one cell -- CONSTANT, never varies

// paceAt returns the offset from home after n steps, mid-step interpolation
// included, plus whether a foot is currently down. The offset is NEGATIVE:
// inward, toward the centre of the screen.
//
// The path is a triangle: in to the far end of the strip, then back out to
// home. Pacing, not a walk-off -- it never leaves, and over a long session it
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
	f := sinceStep / stepDur
	return from + (to-from)*f, true
}

// stepsBy counts the main-thread tool events that have landed by time t, and
// how long ago the last one was.
func stepsBy(t float64, events []float64) (n int, since float64) {
	since = 1e9
	for _, e := range events {
		if e <= t {
			n++
			since = t - e
		}
	}
	return n, since
}

// stillFor says whether a state holds its ground. A raised claw that is also
// pacing is noise; "done" is explicitly held STILL so it cannot be misread as
// another ask; and worried persists, so it should look planted.
func stillFor(pose string) bool {
	return pose == "needs you" || pose == "done" || pose == "worried"
}

// toolStream builds a plausible main-thread event stream at a given rate, with
// the gaps a real agent has rather than a metronome.
func toolStream(dur, perMin float64, seed int64) []float64 {
	if perMin <= 0 {
		return nil
	}
	var out []float64
	t := 0.0
	i := 0
	for t < dur {
		gap := 60 / perMin
		// Jitter the gap so it reads as work rather than as a clock.
		j := 0.45 + 1.35*math.Mod(float64((i*2654435761+int(seed)*40503)&0xffff)/65535.0, 1)
		t += gap * j
		if t < dur {
			out = append(out, t)
		}
		i++
	}
	return out
}

// paceProbe prints the offset sequence for each case, so "does it actually
// move" is answered by reading the numbers rather than by squinting at frames.
func paceProbe(seed int64, frames int, fps float64) {
	for _, c := range []struct {
		n      string
		perMin float64
	}{{"alone", 42}, {"two subagents", 18}, {"six subagents", 6}, {"idle", 0}} {
		ev := toolStream(float64(frames*8)/fps, c.perMin, seed)
		span := paceSpan(88)
		out := ""
		distinct := map[int]bool{}
		for f := 0; f < frames*8; f++ {
			t := float64(f) / fps
			n, since := stepsBy(t, ev)
			dx, _ := paceAt(n, since, span)
			out += string(rune('0' + int(-dx+0.5))) // shown as distance inward from home
			distinct[int(-dx+0.5)] = true
		}
		fmt.Printf("%-14s %2d events over %.1fs, span %d -> %s  (%d distinct columns)\n",
			c.n, len(ev), float64(frames*8)/fps, span, out, len(distinct))
	}
}
