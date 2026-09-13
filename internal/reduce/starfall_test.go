package reduce

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// The reducer flies a star when the count goes UP, on the wall clock, and at no
// other time.
//
// ⚠ THE PHASE IS WALL-CLOCK AGE AND NEVER AN INTEGRATED dt. Shore.Update clamps
// a frame gap over a second, so a fall that accumulated dt would freeze in the
// air whenever the render stalled -- and a laptop that slept mid-fall would wake
// with the star still on its way down. Driving it from the age of a stamp means
// a stalled or suspended session wakes with the star LANDED, which is the only
// honest state: the work finished while nobody was looking.
func TestTheFallFiresOnlyWhenTheStarCountRises(t *testing.T) {
	base := time.Date(2026, 9, 12, 21, 0, 0, 0, time.UTC)
	r := New("test")
	at := func(d time.Duration) (bool, float64) {
		st := r.State(base.Add(d))
		return st.Act.Arriving, st.Act.ArrivalPhase
	}

	// ⚠ A session with no events is the case a zero clock gets WRONG:
	// now.Sub(a zero Time) is fifty-seven years, not "nothing is falling".
	if arriving, _ := at(0); arriving {
		t.Fatal("a session with no events is flying a star")
	}

	// A closed turn is one star. stars() counts closed turns when the agent
	// keeps no checklist, which is how the channel is actually bound today --
	// TodoWrite has fired zero times in the whole recorded history.
	r.Apply(event.Event{Kind: event.Prompt}, base)
	r.Apply(event.Event{Kind: event.Done}, base.Add(time.Second))

	for _, tc := range []struct {
		at     time.Duration
		want   bool
		lo, hi float64
		why    string
	}{
		{time.Second, true, 0, 0.01, "the launch"},
		{1250 * time.Millisecond, true, 0.49, 0.51, "half way down"},
		{1490 * time.Millisecond, true, 0.97, 0.99, "the last airborne frame"},
		{1500 * time.Millisecond, false, 0, 0, "landed: the settled sky owns it now"},
		{9 * time.Second, false, 0, 0, "and it stays landed"},
		{48 * time.Hour, false, 0, 0, "a session left open for two days"},
	} {
		arriving, phase := at(tc.at)
		if arriving != tc.want {
			t.Fatalf("at %v (%s): arriving=%v, want %v", tc.at, tc.why, arriving, tc.want)
		}
		if arriving && (phase < tc.lo || phase > tc.hi) {
			t.Fatalf("at %v (%s): phase %.3f, want %.2f..%.2f", tc.at, tc.why, phase, tc.lo, tc.hi)
		}
	}

	// ⚠ AND THE PHASE NEVER REACHES 1. The landing frame is the SETTLED frame,
	// so the star is drawn at home exactly once rather than by two branches in
	// consecutive frames.
	for d := time.Second; d < 1500*time.Millisecond; d += 5 * time.Millisecond {
		if arriving, phase := at(d); arriving && phase >= 1 {
			t.Fatalf("at %v the phase is %.4f -- a phase of 1 draws the landing twice", d, phase)
		}
	}

	// A SECOND star fires a second fall, from its own stamp.
	r.Apply(event.Event{Kind: event.Prompt}, base.Add(10*time.Second))
	r.Apply(event.Event{Kind: event.Done}, base.Add(11*time.Second))
	if arriving, phase := at(11250 * time.Millisecond); !arriving || phase < 0.49 || phase > 0.51 {
		t.Fatalf("the second star: arriving=%v phase=%.3f, want a fall at half way", arriving, phase)
	}

	// And an event that does not RAISE the count starts nothing. A tool call is
	// the commonest event there is -- 60,000 of them in the recorded history --
	// so a clock stamped by any event at all would keep the sky one star short
	// almost permanently.
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1", Tool: "Bash"}, base.Add(30*time.Second))
	if arriving, _ := at(30100 * time.Millisecond); arriving {
		t.Fatal("a tool call started a fall -- only a rising star count may")
	}
}

// State is called several times a frame in three of the study pages, and more
// than once a frame is normal. It may not restart the fall.
//
// That is the whole reason the rising edge lives in Apply and not in State: a
// clock stamped where frames are built would be re-stamped by every extra call
// and the star would never land.
func TestReadingTheStateDoesNotRestartTheFall(t *testing.T) {
	base := time.Date(2026, 9, 12, 21, 0, 0, 0, time.UTC)
	r := New("test")
	r.Apply(event.Event{Kind: event.Prompt}, base)
	r.Apply(event.Event{Kind: event.Done}, base.Add(time.Second))

	for i := 0; i < 50; i++ {
		r.State(base.Add(1200 * time.Millisecond))
	}
	st := r.State(base.Add(1499 * time.Millisecond))
	if !st.Act.Arriving || st.Act.ArrivalPhase < 0.99 {
		t.Fatalf("after fifty reads the phase is %.3f (arriving=%v), want the fall still on its last frame",
			st.Act.ArrivalPhase, st.Act.Arriving)
	}
	if st := r.State(base.Add(1600 * time.Millisecond)); st.Act.Arriving {
		t.Fatal("the fall outlived its half second because reading the state restarted it")
	}
}

// A count that DROPS stamps nothing, and the star that was already in the sky
// does not fly again.
//
// A todo list is replaced wholesale -- "3 of 7 done" after "9 of 12 done" -- and
// the count falls. Treating any CHANGE as an arrival would fly a star every time
// the agent started a new checklist, for work it had not done.
func TestAFallingCountFliesNothing(t *testing.T) {
	base := time.Date(2026, 9, 12, 21, 0, 0, 0, time.UTC)
	r := New("test")
	r.Apply(event.Event{Kind: event.Todo, N: 9, Of: 12}, base)
	if arriving, _ := func() (bool, float64) {
		st := r.State(base.Add(100 * time.Millisecond))
		return st.Act.Arriving, st.Act.ArrivalPhase
	}(); !arriving {
		t.Fatal("a checklist arriving 9 of 12 done did not fly its newest star")
	}
	// A new, shorter list: fewer done than before.
	r.Apply(event.Event{Kind: event.Todo, N: 3, Of: 7}, base.Add(10*time.Second))
	st := r.State(base.Add(10100 * time.Millisecond))
	if st.Act.Arriving {
		t.Fatal("the count dropped and a star flew anyway")
	}
	if st.Act.TodoDone != 3 {
		t.Fatalf("the sky shows %d stars, want 3", st.Act.TodoDone)
	}
	// And the next real arrival on the new list still flies.
	r.Apply(event.Event{Kind: event.Todo, N: 4, Of: 7}, base.Add(20*time.Second))
	if st := r.State(base.Add(20100 * time.Millisecond)); !st.Act.Arriving {
		t.Fatal("the first arrival after a shorter list did not fly")
	}
}

// A new conversation does not leave a star in mid-air.
//
// ⚠ reset() deliberately does NOT clear the remembered count, and that is the
// subtle half: it does not clear todoDone or turnsDone either, so the
// constellation survives a `clear`. Zeroing the memory would make the very next
// event look like a rise of everything already in the sky.
func TestANewConversationLandsWhateverWasFalling(t *testing.T) {
	base := time.Date(2026, 9, 12, 21, 0, 0, 0, time.UTC)
	r := New("test")
	r.Apply(event.Event{Kind: event.Prompt}, base)
	r.Apply(event.Event{Kind: event.Done}, base.Add(time.Second))
	if st := r.State(base.Add(1100 * time.Millisecond)); !st.Act.Arriving {
		t.Fatal("no fall to interrupt")
	}
	before := r.State(base.Add(1100 * time.Millisecond)).Act.TodoDone

	r.Apply(event.Event{Kind: event.SessionStart, Text: "clear"}, base.Add(1200*time.Millisecond))
	st := r.State(base.Add(1300 * time.Millisecond))
	if st.Act.Arriving {
		t.Fatal("a fall survived a new conversation")
	}
	if st.Act.TodoDone != before {
		t.Fatalf("the sky went from %d stars to %d across a reset; the count is not reset's business",
			before, st.Act.TodoDone)
	}
	// ⚠ And the next event must not read as a rise of the whole sky.
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1", Tool: "Bash"}, base.Add(2*time.Second))
	if st := r.State(base.Add(2100 * time.Millisecond)); st.Act.Arriving {
		t.Fatal("the first event after a reset flew a star that arrived long ago")
	}
}
