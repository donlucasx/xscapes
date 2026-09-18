package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// testdata/kimi/session.jsonl is Kimi Code CLI 0.39.1's own hook payloads,
// captured raw on 2026-09-17 and stitched into one session: a todo list, a
// shell command, two parallel "explore" subagents, a failing command, a
// permission prompt, an Esc, and /exit. This runs them through the hook
// translator and the reducer as a scape would, and checks what the scene
// would have said. Every assertion here was a defect before that day.
func TestARealKimiSessionThroughThePipeline(t *testing.T) {
	f, err := os.Open("testdata/kimi/session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := reduce.New("session_fixture")
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	var (
		prompts []string
		errors  []string
		asks    []string
		dones   int
		peak    int
		atStop  []companion.State
		todo    [2]int
		readTgt string
	)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		var p hookPayload
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			t.Fatalf("a real payload did not parse: %v\n%s", err, sc.Bytes())
		}
		if p.ToolUseID == "" {
			p.ToolUseID = p.ToolCallID
		}
		if p.AgentID == "" && p.AgentName != "" {
			p.AgentID = p.AgentName
		}
		p.Src = "kimi"
		// One second per payload, and a minute across each subagent's end so
		// the litter's dwell does not hide the count.
		now = now.Add(time.Second)
		if p.Event == "SubagentStop" {
			now = now.Add(time.Minute)
		}
		for _, e := range translate(p) {
			e.Src = "kimi"
			switch e.Kind {
			case event.Prompt:
				prompts = append(prompts, e.Text)
			case event.Error:
				errors = append(errors, e.Detail)
			case event.NeedsInput:
				asks = append(asks, e.Text)
			case event.Done:
				dones++
			case event.Todo:
				todo = [2]int{e.N, e.Of}
			case event.ToolEnd:
				if e.Tool == "Read" {
					readTgt = e.Target
				}
			}
			r.Apply(e, now)
		}
		st := r.State(now.Add(200 * time.Millisecond))
		if st.Kittens > peak {
			peak = st.Kittens
		}
		switch p.Event {
		case "Stop":
			atStop = append(atStop, st.Pose)
		case "PermissionRequest":
			if st.Bubble != "allow Bash?" || st.Pose != companion.NeedsYou {
				t.Fatalf("at the permission prompt: bubble %q pose %v", st.Bubble, st.Pose)
			}
		case "PermissionResult":
			// Approved: the ask is over on Kimi's word, before the command
			// has run, not at the next tool start or the turn's end.
			if st.Bubble != "" || st.Pose == companion.NeedsYou {
				t.Fatalf("after the approval: bubble %q pose %v", st.Bubble, st.Pose)
			}
		}
	}

	// The prompt's words reach the sand (an array of parts, not a string).
	if len(prompts) != 4 || !strings.HasPrefix(prompts[0], "Do exactly this, in order") {
		t.Fatalf("prompts = %q", prompts)
	}
	// The failure's reason reaches the sand (an object, not a string).
	if len(errors) != 1 || !strings.Contains(errors[0], "No such file") {
		t.Fatalf("errors = %q", errors)
	}
	// The permission prompt is the ask, and only it.
	if len(asks) != 1 || asks[0] != "allow Bash?" {
		t.Fatalf("asks = %q", asks)
	}
	// Five Stops reach the reducer: the two subagents' (dropped while they
	// are open, so the companion keeps working), the turn's (done), the one
	// after the failing command (WORRIED, not done: a failure you have not
	// seen is not cancelled by the turn ending, his ruling of 2026-09-15),
	// and the one after the approved command (done).
	if dones != 5 {
		t.Fatalf("the translator produced %d done events, want 5", dones)
	}
	want := []companion.State{companion.Working, companion.Working, companion.Done, companion.Worried, companion.Done}
	if len(atStop) != 5 || atStop[0] == companion.Done || atStop[1] == companion.Done ||
		atStop[2] != want[2] || atStop[3] != want[3] || atStop[4] != want[4] {
		t.Fatalf("poses at the five Stops: %v, want [not done, not done, done, worried, done]", atStop)
	}
	// Two explores are two kittens.
	if peak != 2 {
		t.Fatalf("peak litter %d, want 2", peak)
	}
	// Kimi's TodoList lights the stars: two of two done by the end.
	if todo != [2]int{2, 2} {
		t.Fatalf("todo = %v, want 2 of 2", todo)
	}
	// Kimi's Read names its file with `path`.
	if readTgt != "README.txt" {
		t.Fatalf("Read target %q", readTgt)
	}
	// After Esc and /exit: settled, silent.
	st := r.State(now.Add(time.Second))
	if st.Act.Working || st.Pose == companion.Done || st.Bubble != "" {
		t.Fatalf("after the interrupt and the exit: working %v pose %v bubble %q", st.Act.Working, st.Pose, st.Bubble)
	}
}

// The shapes one at a time, against Claude Code's, so the two never drift.
func TestKimiShapesBesideClaudes(t *testing.T) {
	kimi := func(t *testing.T, js string) hookPayload {
		p := payload(t, js)
		p.Src = "kimi"
		return p
	}
	// A prompt as parts and as a string.
	if got := translate(kimi(t, `{"hook_event_name":"UserPromptSubmit","session_id":"s","prompt":[{"type":"text","text":"fix the build"},{"type":"image"}]}`)); len(got) != 1 || got[0].Text != "fix the build" {
		t.Fatalf("parts prompt: %+v", got)
	}
	if got := translate(payload(t, `{"hook_event_name":"UserPromptSubmit","session_id":"s","prompt":"fix the build"}`)); len(got) != 1 || got[0].Text != "fix the build" {
		t.Fatalf("string prompt: %+v", got)
	}
	// An error as an object and as a string.
	if got := translate(kimi(t, `{"hook_event_name":"PostToolUseFailure","session_id":"s","tool_name":"Bash","tool_input":{"command":"cat x"},"error":{"code":"internal","message":"cat: x: No such file\nexit 1"}}`)); len(got) != 1 || got[0].Kind != event.Error || got[0].Detail != "cat: x: No such file" {
		t.Fatalf("object error: %+v", got)
	}
	if got := translate(payload(t, `{"hook_event_name":"PostToolUseFailure","session_id":"s","tool_name":"Bash","error":"boom"}`)); len(got) != 1 || got[0].Detail != "boom" {
		t.Fatalf("string error: %+v", got)
	}
	// TodoList counts "done"; TodoWrite still counts "completed".
	if got := translate(kimi(t, `{"hook_event_name":"PostToolUse","session_id":"s","tool_name":"TodoList","tool_input":{"todos":[{"status":"done","title":"a"},{"status":"in_progress","title":"b"},{"status":"pending","title":"c"}]}}`)); len(got) != 2 || got[1].Kind != event.Todo || got[1].N != 1 || got[1].Of != 3 {
		t.Fatalf("TodoList: %+v", got)
	}
	if got := translate(payload(t, `{"hook_event_name":"PostToolUse","session_id":"s","tool_name":"TodoWrite","tool_input":{"todos":[{"status":"completed"},{"status":"pending"}]}}`)); len(got) != 2 || got[1].N != 1 || got[1].Of != 2 {
		t.Fatalf("TodoWrite: %+v", got)
	}
	// AskUserQuestion is the ask on Kimi, a plain tool on Claude Code.
	ask := `{"hook_event_name":"PreToolUse","session_id":"s","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"Red or blue?","header":"Colour"}]}}`
	if got := translate(kimi(t, ask)); len(got) != 2 || got[1].Kind != event.NeedsInput || got[1].Text != "Red or blue?" {
		t.Fatalf("kimi AskUserQuestion: %+v", got)
	}
	if got := translate(payload(t, ask)); len(got) != 1 || got[0].Kind != event.ToolStart {
		t.Fatalf("claude AskUserQuestion: %+v", got)
	}
	// Esc.
	if got := translate(kimi(t, `{"hook_event_name":"Interrupt","session_id":"s","reason":"cancelled","turn_id":1}`)); len(got) != 1 || got[0].Kind != event.Interrupt || got[0].Text != "cancelled" {
		t.Fatalf("Interrupt: %+v", got)
	}
	// A background-task notice is not an ask.
	if got := translate(kimi(t, `{"hook_event_name":"Notification","session_id":"s","notification_type":"task.completed","task_id":"t1"}`)); len(got) != 0 {
		t.Fatalf("task.completed became %+v", got)
	}
}

// Kimi says when an approval was answered; Claude Code never does, so there
// the next tool starting is what clears the ask. On Kimi the answer itself
// clears it, before the command has run.
func TestAnApprovalAnsweredClearsTheAsk(t *testing.T) {
	p := payload(t, `{"hook_event_name":"PermissionResult","session_id":"s","tool_name":"Bash","tool_input":{"command":"ls"},"decision":"approved"}`)
	p.Src = "kimi"
	got := translate(p)
	if len(got) != 1 || got[0].Kind != event.Answered || got[0].Text != "approved" {
		t.Fatalf("PermissionResult: %+v", got)
	}
}
