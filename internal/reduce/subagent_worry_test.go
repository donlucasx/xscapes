package reduce

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
)

// His ruling of 2026-09-07: "main-thread errors only".
//
// The companion was Worried 44.6% of active time, measured over 330 hours of
// his own recordings -- nearly as much as it was Working, and a cue that is up
// half the session cannot mean "something is broken". 590 of those 675 errors
// (87%) fired inside a SUBAGENT: a fan-out grepping for something that was not
// there, a probe that exited 1. Those are the subagent's business and it
// almost always carried on -- 99.4% of all errors are followed by a successful
// tool call, median 1.7 seconds later.
//
// Event.Agent is set only when the hook fires inside a subagent
// (notes/claude-hooks-verified.md), so the main thread is exactly Agent == "".
// Folded back through his real spools this takes worried to 16.4% and working
// to 77.7%, with needs-you and done unchanged.
func TestOnlyMainThreadErrorsWorryTheCompanion(t *testing.T) {
	errAt := func(r *Reducer, sec float64, agent string) {
		r.Apply(event.Event{
			Kind: event.Error, Op: event.OpShell, Tool: "Bash", Target: "grep",
			Detail: "Exit code 1", Agent: agent,
		}, at(sec))
	}

	t.Run("a subagent's error does not worry", func(t *testing.T) {
		r := New("s")
		r.Apply(event.Event{Kind: event.Prompt}, at(0))
		errAt(r, 1, "agent-7")
		if got := r.State(at(2)).Pose; got == companion.Worried {
			t.Errorf("a subagent error raised Worried; the pose should still be %v, got %v",
				companion.Working, got)
		}
	})

	t.Run("a main-thread error still worries", func(t *testing.T) {
		r := New("s")
		r.Apply(event.Event{Kind: event.Prompt}, at(0))
		errAt(r, 1, "")
		if got := r.State(at(2)).Pose; got != companion.Worried {
			t.Errorf("a main-thread error must still worry; got %v", got)
		}
	})

	// The positive control that makes the first case mean something: a
	// subagent error must still be recorded in every OTHER channel. It is a
	// real event -- it goes on the sand, it drives the sea -- it just does not
	// change what the companion's face says.
	t.Run("a subagent's error is still recorded", func(t *testing.T) {
		r := New("s")
		r.Apply(event.Event{Kind: event.Prompt}, at(0))
		before := r.State(at(1)).Act.Level
		errAt(r, 1, "agent-7")
		st := r.State(at(2))
		if st.Act.Level <= before {
			t.Errorf("the sea did not move for a subagent error: %v -> %v", before, st.Act.Level)
		}
		var found bool
		for _, l := range st.Tail {
			if l.Bad {
				found = true
			}
		}
		if !found {
			t.Error("a subagent error left no failed line on the sand")
		}
	})
}
