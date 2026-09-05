package reduce

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/event"
)

// A subagent that finishes after its dwell leaves the litter at once (the
// count is live) and swims off over KittenExit: the exit list carries its
// progress, then empties.
func TestAFinishedSubagentSwimsOffBeforeItIsGone(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(0))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "b"}, at(0))
	e := KittenDwell.Seconds() + 3 // past the dwell: the end is the exit
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "b"}, at(e))
	st := r.State(at(e))
	if st.Kittens != 1 {
		t.Errorf("kittens = %d right after the end, want 1", st.Kittens)
	}
	if len(st.KittenExits) != 1 || st.KittenExits[0] > 0.05 {
		t.Errorf("exits right after the end = %v, want one at ~0", st.KittenExits)
	}
	half := e + KittenExit.Seconds()/2
	if st := r.State(at(half)); len(st.KittenExits) != 1 || st.KittenExits[0] < 0.4 || st.KittenExits[0] > 0.6 {
		t.Errorf("exits halfway = %v, want one at ~0.5", st.KittenExits)
	}
	if st := r.State(at(e + KittenExit.Seconds() + 0.1)); len(st.KittenExits) != 0 || st.Kittens != 1 {
		t.Errorf("after the exit: exits %v kittens %d, want none and 1", st.KittenExits, st.Kittens)
	}
	// A subagent that reports again while leaving is back in the litter.
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "a"}, at(e+20))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(e+21))
	if st := r.State(at(e + 21)); st.Kittens != 1 || len(st.KittenExits) != 0 {
		t.Errorf("revived: kittens %d exits %v, want 1 and none", st.Kittens, st.KittenExits)
	}
}

// A subagent that finishes in seconds was invisible: its kitten came and went
// before anyone looked (his report, twice: 2026-09-04 and 2026-09-05, agents
// of 6.9 s and 15.7 s). So a kitten stays in the litter at least KittenDwell
// after its start, and only then swims off. The end event is remembered, the
// exit is deferred.
func TestAShortSubagentsKittenStaysForTheDwell(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(0))
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "a"}, at(5))
	d := KittenDwell.Seconds()
	for _, sec := range []float64{5, 30, d - 1} {
		if st := r.State(at(sec)); st.Kittens != 1 || len(st.KittenExits) != 0 {
			t.Errorf("at %.0fs: kittens %d exits %v, want 1 still sitting and no exit", sec, st.Kittens, st.KittenExits)
		}
	}
	if st := r.State(at(d + 0.1)); st.Kittens != 0 || len(st.KittenExits) != 1 || st.KittenExits[0] > 0.05 {
		t.Errorf("at the dwell: kittens %d exits %v, want 0 and one exit at ~0", st.Kittens, st.KittenExits)
	}
	if st := r.State(at(d + KittenExit.Seconds() + 0.1)); st.Kittens != 0 || len(st.KittenExits) != 0 {
		t.Errorf("after the swim-off: kittens %d exits %v, want none", st.Kittens, st.KittenExits)
	}
}

// A subagent that reports again during its dwell is simply running: the
// deferred exit must be forgotten, or the sweep would drop a live kitten at
// the old departure time.
func TestASubagentThatRestartsDuringItsDwellIsRunningAgain(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(0))
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "a"}, at(5))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(10))
	for _, sec := range []float64{10, KittenDwell.Seconds() + 1, KittenDwell.Seconds() + 20} {
		if st := r.State(at(sec)); st.Kittens != 1 || len(st.KittenExits) != 0 {
			t.Errorf("at %.0fs: kittens %d exits %v, want 1 running and no exit", sec, st.Kittens, st.KittenExits)
		}
	}
}
