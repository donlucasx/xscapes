package reduce

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// The constellation counts closed turns, because the thing it was spec'd to
// count has never happened.
//
// Stars were "todos completed". Across his whole recorded history -- 60,000+
// tool calls -- TodoWrite has been called ZERO times, so the channel had never
// lit once. Closed turns are the same idea carried by an event that fires:
// measured over his sessions, p50 12, p75 16, p90 24, max 32.
func TestTheConstellationCountsClosedTurns(t *testing.T) {
	now := time.Now()
	r := New("t")
	if got := r.State(now).Act.TodoDone; got != 0 {
		t.Fatalf("a fresh session starts with %d stars, want 0", got)
	}
	for i := 1; i <= 5; i++ {
		r.Apply(event.Event{Kind: event.Done}, now)
		if got := r.State(now).Act.TodoDone; got != i {
			t.Errorf("after %d closed turns the sky shows %d stars", i, got)
		}
	}
	// The layout is a fixed number of PLACES, so a star lights where it always
	// was rather than sliding as the next one arrives.
	if got := r.State(now).Act.TodoTotal; got != StarsCap {
		t.Errorf("the sky lays out %d places, want the fixed %d", got, StarsCap)
	}
	// It fills a long session without pegging, and then holds.
	for i := 0; i < StarsCap*2; i++ {
		r.Apply(event.Event{Kind: event.Done}, now)
	}
	if got := r.State(now).Act.TodoDone; got != StarsCap {
		t.Errorf("past the cap the sky shows %d stars, want it held at %d", got, StarsCap)
	}

	// An agent that DOES keep a checklist still wins: it is telling us
	// something more specific than "a turn ended".
	r2 := New("t")
	r2.Apply(event.Event{Kind: event.Done}, now)
	r2.Apply(event.Event{Kind: event.Done}, now)
	r2.Apply(event.Event{Kind: event.Todo, N: 1, Of: 7}, now)
	st := r2.State(now)
	if st.Act.TodoDone != 1 || st.Act.TodoTotal != 7 {
		t.Errorf("with a real checklist the sky shows %d of %d, want 1 of 7 -- the turns should not override it",
			st.Act.TodoDone, st.Act.TodoTotal)
	}
}
