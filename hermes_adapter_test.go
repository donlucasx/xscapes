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
