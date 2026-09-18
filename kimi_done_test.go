package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// Kimi fires a Stop for a sub-agent's own turn as well as for the main one,
// with the same payload, so a Stop that arrives while an instance is open is
// dropped as the sub-agent's (session 40). With BACKGROUND agents the turn's
// own Stop arrives while they are still open and was dropped for good: no
// finish cue, and the turn settled by the thirty-minute silence (Kimi's
// assessment, 2026-09-18). What tells the two apart is what follows. A
// sub-agent's Stop is claimed by its SubagentStop within the same second
// (the fixture: Stop, the Agent call's PostToolUse, SubagentStop); a turn's
// Stop with agents in the background is followed by nothing for minutes. So
// a dropped Done that no SubagentStop claims within two seconds is the
// turn's, and rings when the last open instance closes.
func TestABackgroundAgentsTurnRingsWhenTheLastOneIsIn(t *testing.T) {
	ev := func(k event.Kind, agent, text string) event.Event {
		e := event.Event{Kind: k, Src: "kimi", Text: text}
		if k == event.SubStart || k == event.SubEnd {
			e.Op, e.Agent = event.OpSub, agent
		}
		return e
	}
	now := time.Date(2026, 9, 18, 2, 0, 0, 0, time.UTC)
	r := reduce.New("background")
	step := func(e event.Event, dt time.Duration) {
		now = now.Add(dt)
		r.Apply(e, now)
	}
	step(ev(event.Prompt, "", "start two explores in the background"), 0)
	step(ev(event.SubStart, "explore", ""), time.Second)
	step(ev(event.SubStart, "explore", ""), 0)
	// The turn's own Stop, with both agents still running.
	step(ev(event.Done, "", "Both are running; I will report when they finish."), 2*time.Second)
	if st := r.State(now.Add(time.Second)); st.Pose == companion.Done {
		t.Fatal("a Stop with agents still open rang the finish")
	}
	// The first agent's own Stop, then its SubagentStop, as Kimi sends them.
	step(ev(event.Done, "", "found it"), 40*time.Second)
	step(ev(event.SubEnd, "explore", ""), 300*time.Millisecond)
	if st := r.State(now); st.Pose == companion.Done {
		t.Fatal("the finish rang with one agent still open")
	}
	// The second: the last one in, and the turn's Done rings.
	step(ev(event.Done, "", "found the other"), 30*time.Second)
	step(ev(event.SubEnd, "explore", ""), 300*time.Millisecond)
	if st := r.State(now); st.Pose != companion.Done {
		t.Fatalf("the last agent is in and the turn's Done did not ring: pose %v", st.Pose)
	}

	// FOREGROUND agents, the fixture's shape: each sub-agent's Stop is
	// followed within the second by the Agent call's end and its
	// SubagentStop, and the turn goes on. Nothing rings until the turn's
	// own Stop.
	now = time.Date(2026, 9, 18, 3, 0, 0, 0, time.UTC)
	r = reduce.New("foreground")
	step(ev(event.Prompt, "", "read the two files"), 0)
	step(event.Event{Kind: event.ToolStart, Src: "kimi", ID: "a1", Op: event.OpSub, Tool: "Agent"}, time.Second)
	step(ev(event.SubStart, "explore", ""), 0)
	step(ev(event.Done, "", "the first file says"), 5*time.Second)
	step(event.Event{Kind: event.ToolEnd, Src: "kimi", ID: "a1", Op: event.OpSub, Tool: "Agent"}, 100*time.Millisecond)
	step(ev(event.SubEnd, "explore", ""), 100*time.Millisecond)
	step(event.Event{Kind: event.ToolStart, Src: "kimi", ID: "t9", Op: event.OpRead, Tool: "Read"}, 500*time.Millisecond)
	if st := r.State(now); st.Pose == companion.Done {
		t.Fatal("a sub-agent's Stop rang the finish while the turn went on")
	}
	step(ev(event.Done, "", "both read"), 3*time.Second)
	if st := r.State(now); st.Pose != companion.Done {
		t.Fatalf("the turn's own Stop did not ring: pose %v", st.Pose)
	}

	// A NEW PROMPT moots a held Done: the old turn's finish never rings into
	// the new one.
	now = time.Date(2026, 9, 18, 4, 0, 0, 0, time.UTC)
	r = reduce.New("moot")
	step(ev(event.Prompt, "", "start one in the background"), 0)
	step(ev(event.SubStart, "explore", ""), time.Second)
	step(ev(event.Done, "", "running"), 2*time.Second)
	step(ev(event.Prompt, "", "and now something else"), 10*time.Second)
	step(ev(event.Done, "", "its own stop"), 20*time.Second)
	step(ev(event.SubEnd, "explore", ""), 300*time.Millisecond)
	if st := r.State(now); st.Pose == companion.Done {
		t.Fatal("the old turn's held Done rang into the new turn")
	}
}

// A StopFailure flagged as an interrupt is the user's Esc, not a failure:
// the owl must not worry over it (PostToolUseFailure already read the flag;
// StopFailure did not).
func TestAStopFailureThatIsAnInterruptStaysSilent(t *testing.T) {
	evs := hookTranslate([]string{"StopFailure", "kimi"}, []byte(`{"hook_event_name":"StopFailure","session_id":"s","is_interrupt":true,"error":"interrupted"}`))
	if len(evs) != 1 || evs[0].Kind != event.Interrupt {
		t.Fatalf("an interrupted StopFailure became %+v, want one Interrupt", evs)
	}
	evs = hookTranslate([]string{"StopFailure", "kimi"}, []byte(`{"hook_event_name":"StopFailure","session_id":"s","is_interrupt":false,"error":"provider error"}`))
	if len(evs) != 1 || evs[0].Kind != event.Error {
		t.Fatalf("a real StopFailure became %+v, want one Error", evs)
	}
}
