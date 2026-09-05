package reduce

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/event"
)

// A subagent that finishes leaves the litter at once (the count is live) and
// swims off over KittenExit: the exit list carries its progress, then empties.
func TestAFinishedSubagentSwimsOffBeforeItIsGone(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(0))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "b"}, at(0))
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "b"}, at(3))
	st := r.State(at(3))
	if st.Kittens != 1 {
		t.Errorf("kittens = %d right after the end, want 1", st.Kittens)
	}
	if len(st.KittenExits) != 1 || st.KittenExits[0] > 0.05 {
		t.Errorf("exits right after the end = %v, want one at ~0", st.KittenExits)
	}
	half := 3 + KittenExit.Seconds()/2
	if st := r.State(at(half)); len(st.KittenExits) != 1 || st.KittenExits[0] < 0.4 || st.KittenExits[0] > 0.6 {
		t.Errorf("exits halfway = %v, want one at ~0.5", st.KittenExits)
	}
	if st := r.State(at(3 + KittenExit.Seconds() + 0.1)); len(st.KittenExits) != 0 || st.Kittens != 1 {
		t.Errorf("after the exit: exits %v kittens %d, want none and 1", st.KittenExits, st.Kittens)
	}
	// A subagent that reports again while leaving is back in the litter.
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "a"}, at(20))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(21))
	if st := r.State(at(21)); st.Kittens != 1 || len(st.KittenExits) != 0 {
		t.Errorf("revived: kittens %d exits %v, want 1 and none", st.Kittens, st.KittenExits)
	}
}
