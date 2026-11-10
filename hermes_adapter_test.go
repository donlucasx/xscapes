package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// Payload shapes from Hermes Agent's own source at 856e7580 (2026-11-08):
// tools/delegate_tool.py invokes subagent_start with child_session_id,
// child_subagent_id, child_role and child_goal; tools/delegate_tool_results.py
// invokes subagent_stop with child_session_id, child_role, child_status and
// duration_ms; pre_approval_request carries command and description
// (website/docs/user-guide/features/hooks.md). Shell hooks deliver all of them
// under "extra".

// TestHermesSubagentStartIsAnArrival: Hermes v0.14 had no start event, so a
// subagent only ever appeared when it finished. Newer Hermes announces the
// start, keyed by the same child_session_id its stop carries.
func TestHermesSubagentStartIsAnArrival(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"subagent_start","session_id":"p",
		"extra":{"child_session_id":"c1","child_subagent_id":"sa1","child_role":"leaf","child_goal":"read the tests"}}`))
	if len(got) != 1 || got[0].Kind != event.SubStart || got[0].Agent != "c1" {
		t.Fatalf("subagent_start = %+v, want one SubStart for agent c1", got)
	}
}

// TestHermesSubagentStopStillWorksAlone: an older Hermes sends only the stop,
// so the stop keeps announcing an arrival and a departure together.
func TestHermesSubagentStopStillWorksAlone(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"subagent_stop","session_id":"p",
		"extra":{"child_session_id":"c1","child_role":"leaf","child_status":"completed","duration_ms":4200}}`))
	if len(got) != 2 || got[0].Kind != event.SubStart || got[1].Kind != event.SubEnd || got[1].Agent != "c1" {
		t.Fatalf("subagent_stop = %+v, want SubStart+SubEnd for c1", got)
	}
}

// TestASecondStartDoesNotRestartTheDwell: with both events, the stop's
// arrival lands on a subagent already in the litter. It must not reset its
// start time, or every Hermes subagent would linger a full KittenDwell after
// it finished instead of leaving at max(end, start+dwell).
func TestASecondStartDoesNotRestartTheDwell(t *testing.T) {
	t0 := time.Unix(1_800_000_000, 0)
	r := reduce.New("s")
	r.Apply(event.Event{Kind: event.SubStart, Op: event.OpSub, Agent: "c1"}, t0)
	end := t0.Add(5 * time.Minute)
	r.Apply(event.Event{Kind: event.SubStart, Op: event.OpSub, Agent: "c1"}, end)
	r.Apply(event.Event{Kind: event.SubEnd, Op: event.OpSub, Agent: "c1"}, end)
	if k := r.State(end.Add(reduce.KittenDwell / 2)).Kittens; k != 0 {
		t.Fatalf("a subagent that ran five minutes is still in the litter %v after its end (%d)", reduce.KittenDwell/2, k)
	}
}

// TestHermesApprovalNamesTheCommand: Hermes's approval event has no
// tool_name, so the balloon read "allow ?". It names the program instead.
func TestHermesApprovalNamesTheCommand(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"pre_approval_request","session_id":"p",
		"extra":{"command":"rm -rf build","description":"recursive delete","pattern_key":"rm_rf","surface":"cli"}}`))
	if len(got) != 1 || got[0].Kind != event.NeedsInput || got[0].Text != "allow rm?" {
		t.Fatalf("pre_approval_request = %+v, want NeedsInput \"allow rm?\"", got)
	}
}

// TestHermesApprovalAnsweredTakesTheBalloonDown: the answer is the end of the
// ask, as PermissionResult is for Kimi. Without it the balloon waited for the
// next tool start.
func TestHermesApprovalAnsweredTakesTheBalloonDown(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"post_approval_response","session_id":"p",
		"extra":{"command":"rm -rf build","choice":"once","surface":"cli"}}`))
	if len(got) != 1 || got[0].Kind != event.Answered {
		t.Fatalf("post_approval_response = %+v, want one Answered", got)
	}
}

// TestHermesInstallsTheNewEvents: the installer registers the events the
// translator reads, or the arrival and the answer never fire.
func TestHermesInstallsTheNewEvents(t *testing.T) {
	for _, want := range []string{"subagent_start", "post_approval_response"} {
		found := false
		for _, e := range hermesHookEvents {
			found = found || e == want
		}
		if !found {
			t.Errorf("hermesHookEvents lacks %s", want)
		}
	}
}

// TestHermesTurnEndIsNotTheSessionEnd: Hermes fires on_session_end "at the
// very end of every run_conversation() call" (hooks.md), i.e. every turn. As
// a SessionEnd it reset the reducer the instant the turn's done landed, so
// the done knock, the litter and the stars vanished after every turn
// (measured 2026-11-08: done and session_end in the same millisecond).
// on_session_finalize is the session closing.
func TestHermesTurnEndIsNotTheSessionEnd(t *testing.T) {
	if got := translate(payload(t, `{"hook_event_name":"on_session_end","session_id":"p","extra":{}}`)); len(got) != 0 {
		t.Fatalf("on_session_end = %+v, want nothing (it ends a turn, not the session)", got)
	}
	got := translate(payload(t, `{"hook_event_name":"on_session_finalize","session_id":"p","extra":{}}`))
	if len(got) != 1 || got[0].Kind != event.SessionEnd {
		t.Fatalf("on_session_finalize = %+v, want SessionEnd", got)
	}
}

// TestHermesQuestionsRaiseTheAsk: Hermes blocks on the person three ways
// (tools/human_input_hooks.py: kind sudo, clarify or approval). A question
// or a password raises the ask; an approval is left to pre_approval_request
// so it never rings twice; the end of the wait takes the ask down.
func TestHermesQuestionsRaiseTheAsk(t *testing.T) {
	for kind, want := range map[string]string{"clarify": "a question for you", "sudo": "sudo password?"} {
		got := translate(payload(t, `{"hook_event_name":"on_human_input_request","session_id":"p",
			"extra":{"kind":"`+kind+`","request_id":"r1","prompt":"[redacted]","platform":"cli"}}`))
		if len(got) != 1 || got[0].Kind != event.NeedsInput || got[0].Text != want {
			t.Errorf("%s request = %+v, want NeedsInput %q", kind, got, want)
		}
		got = translate(payload(t, `{"hook_event_name":"on_human_input_resolved","session_id":"p",
			"extra":{"kind":"`+kind+`","request_id":"r1","outcome":"answered"}}`))
		if len(got) != 1 || got[0].Kind != event.Answered {
			t.Errorf("%s resolved = %+v, want one Answered", kind, got)
		}
	}
	for _, ev := range []string{"on_human_input_request", "on_human_input_resolved"} {
		if got := translate(payload(t, `{"hook_event_name":"`+ev+`","session_id":"p","extra":{"kind":"approval","request_id":"r2"}}`)); len(got) != 0 {
			t.Errorf("%s for an approval = %+v, want nothing (pre_approval_request owns it)", ev, got)
		}
	}
}

// TestHermesModelCallsFeedTheMoonAndTheCount: post_api_request carries the
// call's usage and the context window (agent/turn_response_intake.py); the
// plugin adds the session's running total. The moon takes prompt/window, the
// count takes the total, and both reach the scene's state.
func TestHermesModelCallsFeedTheMoonAndTheCount(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"post_api_request","session_id":"p",
		"extra":{"usage":{"prompt_tokens":250000,"total_tokens":251200},"context_length":1000000,"xscapes_total":4200000}}`))
	if len(got) != 1 || got[0].Kind != event.Context || got[0].Frac == nil || *got[0].Frac != 0.25 ||
		got[0].Window != 1000000 || got[0].Tokens != 4200000 {
		t.Fatalf("post_api_request = %+v, want Context frac 0.25 window 1M tokens 4.2M", got)
	}
	t0 := time.Unix(1_800_000_000, 0)
	r := reduce.New("p")
	r.Apply(got[0], t0)
	if st := r.State(t0); !st.ContextKnown || st.Act.ContextUsed != 0.25 || st.Act.Tokens != 4200000 {
		t.Fatalf("state after the call: known %v, used %v, tokens %d", st.ContextKnown, st.Act.ContextUsed, st.Act.Tokens)
	}
	if got := translate(payload(t, `{"hook_event_name":"post_api_request","session_id":"p","extra":{"usage":null}}`)); len(got) != 0 {
		t.Fatalf("a call without usage = %+v, want nothing", got)
	}
}

// TestHermesToolsPairByCallID: a Hermes tool's start and end carry its
// tool_call_id, so the scene knows a tool is still running (the heavier rain,
// and no thought puff over a running tool).
func TestHermesToolsPairByCallID(t *testing.T) {
	start := translate(payload(t, `{"hook_event_name":"pre_tool_call","tool_name":"terminal","tool_input":{"command":"ls"},"session_id":"p","extra":{"tool_call_id":"call_7"}}`))
	end := translate(payload(t, `{"hook_event_name":"post_tool_call","tool_name":"terminal","tool_input":{"command":"ls"},"session_id":"p","extra":{"tool_call_id":"call_7","duration_ms":40}}`))
	if len(start) != 1 || start[0].ID != "call_7" || len(end) != 1 || end[0].ID != "call_7" {
		t.Fatalf("start %+v, end %+v: want both with ID call_7", start, end)
	}
}
