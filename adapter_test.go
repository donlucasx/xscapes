package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
	"github.com/donlucasx/asciiscapes/internal/reduce"
)

// The payload shapes below are the real ones, taken from the Zod schema
// embedded in the Claude Code 2.1.251 binary and written down in
// notes/claude-hooks-verified.md. They are not invented, and they are not
// copied from documentation.

func shortHome(t *testing.T) {
	t.Helper()
	d, err := os.MkdirTemp("", "as")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	t.Setenv("ASCIISCAPES_HOME", d)
}

// TestTheIdleNagIsInvisible is the single most important test in the project.
//
// 41.6% of all Notification events in 13,033 measured on this machine are the
// idle nag, 2,220 of them at exactly sixty seconds after Stop. Wiring
// Notification straight to the companion makes the cat knock falsely about
// once a minute, forever, which would train the user to ignore the one signal
// the whole thing exists to deliver.
func TestTheIdleNagIsInvisible(t *testing.T) {
	if got := translate(payload(t, `{
		"hook_event_name": "Notification",
		"session_id": "s",
		"notification_type": "idle_prompt",
		"message": "Claude is waiting for your input"
	}`)); len(got) != 0 {
		t.Errorf("the idle nag produced %d events; it must produce none: %+v", len(got), got)
	}
}

func TestNotificationTypesAreAnAllowList(t *testing.T) {
	// Genuine asks. These are the ones a person actually has to answer.
	for _, nt := range []string{
		"permission_prompt", "worker_permission_prompt", "agent_needs_input",
		"elicitation_dialog", "elicitation_url_dialog",
	} {
		got := translate(payload(t, `{"hook_event_name":"Notification","session_id":"s","notification_type":"`+nt+`","message":"m"}`))
		if len(got) != 1 || got[0].Kind != event.NeedsInput {
			t.Errorf("%s produced %+v, want one needs_input", nt, got)
		}
	}
	// Everything else is ignored, INCLUDING types this build has never seen.
	// The list fails closed on purpose: a deny list would start knocking the
	// day Claude Code adds a new chatty notification type.
	for _, nt := range []string{
		"idle_prompt", "auth_success", "push_notification", "computer_use_enter",
		"quota_auto_resume_fired", "some_type_invented_next_year",
	} {
		if got := translate(payload(t, `{"hook_event_name":"Notification","session_id":"s","notification_type":"`+nt+`","message":"m"}`)); len(got) != 0 {
			t.Errorf("%s produced %+v, want nothing", nt, got)
		}
	}
}

// Subagent identity was believed to arrive for weeks and never confirmed.
// It does, as agent_type and agent_id, on dedicated hook events.
func TestSubagentIdentityArrives(t *testing.T) {
	got := translate(payload(t, `{
		"hook_event_name": "SubagentStart",
		"session_id": "s", "agent_id": "ag_01", "agent_type": "code-reviewer"
	}`))
	if len(got) != 1 || got[0].Kind != event.SubStart {
		t.Fatalf("got %+v", got)
	}
	if got[0].AgentType != "code-reviewer" {
		t.Errorf("agent type = %q, want code-reviewer", got[0].AgentType)
	}
}

// A user pressing escape is not a broken build.
func TestInterruptIsNotAFailure(t *testing.T) {
	fail := translate(payload(t, `{"hook_event_name":"PostToolUseFailure","session_id":"s","tool_name":"Bash",
		"tool_use_id":"t","error":"exit status 1","tool_input":{"command":"go test ./..."}}`))
	if len(fail) != 1 || fail[0].Kind != event.Error {
		t.Errorf("a real failure should be an error event, got %+v", fail)
	}
	stop := translate(payload(t, `{"hook_event_name":"PostToolUseFailure","session_id":"s","tool_name":"Bash",
		"tool_use_id":"t","error":"interrupted","is_interrupt":true,"tool_input":{"command":"sleep 90"}}`))
	if len(stop) != 1 || stop[0].Kind == event.Error {
		t.Errorf("an interrupt must not be an error, got %+v", stop)
	}
}

// A shell command is the one tool input that routinely holds a secret, and
// this string is written to a file and painted on a screen.
func TestShellCommandsAreReducedToTheProgramName(t *testing.T) {
	got := translate(payload(t, `{"hook_event_name":"PreToolUse","session_id":"s","tool_name":"Bash","tool_use_id":"t",
		"tool_input":{"command":"psql \"postgresql://user:hunter2@db.example.com/prod\" -c 'select 1'"}}`))
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	if got[0].Target != "psql" {
		t.Errorf("target = %q, want just the program name", got[0].Target)
	}
	if line := string(event.Encode(got[0])); contains(line, "hunter2") || contains(line, "postgresql://") {
		t.Errorf("the encoded event leaked the command's arguments: %s", line)
	}
}

// The whole pipeline: hook JSON on one side, a scene on the other.
func TestEndToEndFromRealHookPayloads(t *testing.T) {
	shortHome(t)
	const sess = "e2e-0001"

	bus, err := event.Listen(sess)
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()
	red := reduce.New(sess)

	base := time.Now()
	at := func(s int) time.Time { return base.Add(time.Duration(s) * time.Second) }

	step := func(now time.Time, payloads ...string) reduce.State {
		for _, raw := range payloads {
			p := payload(t, raw)
			for _, e := range translate(p) {
				e.Session, e.Agent, e.Src = p.Session, p.AgentID, "claude"
				if _, err := event.Emit(e); err != nil {
					t.Fatal(err)
				}
			}
		}
		// Drain until the bus goes quiet.
		for {
			select {
			case e := <-bus.C:
				red.Apply(e, now)
			case <-time.After(50 * time.Millisecond):
				return red.State(now)
			}
		}
	}

	st := step(at(0), `{"hook_event_name":"UserPromptSubmit","session_id":"e2e-0001","prompt":"add rate limiting"}`)
	if !st.Act.Working || st.Act.Level < reduce.TurnFloor {
		t.Errorf("thinking should hold the sea up: working=%v level=%.2f", st.Act.Working, st.Act.Level)
	}

	st = step(at(2),
		`{"hook_event_name":"PreToolUse","session_id":"e2e-0001","tool_name":"Read","tool_use_id":"t1","tool_input":{"file_path":"/x/auth/handler.go"}}`,
		`{"hook_event_name":"PostToolUse","session_id":"e2e-0001","tool_name":"Read","tool_use_id":"t1","duration_ms":412,"tool_input":{"file_path":"/x/auth/handler.go"}}`,
		`{"hook_event_name":"PreToolUse","session_id":"e2e-0001","tool_name":"Edit","tool_use_id":"t2","tool_input":{"file_path":"/x/auth/handler.go"}}`,
		`{"hook_event_name":"PostToolUse","session_id":"e2e-0001","tool_name":"Edit","tool_use_id":"t2","duration_ms":90,"tool_input":{"file_path":"/x/auth/handler.go"}}`)
	if len(st.Tail) != 2 {
		t.Errorf("the sand should hold two lines, got %d: %+v", len(st.Tail), st.Tail)
	}
	if st.Act.Level < 0.5 {
		t.Errorf("four tool events should read as busy, level = %.2f", st.Act.Level)
	}

	st = step(at(3),
		`{"hook_event_name":"SubagentStart","session_id":"e2e-0001","agent_id":"a1","agent_type":"code-reviewer"}`,
		`{"hook_event_name":"SubagentStart","session_id":"e2e-0001","agent_id":"a2","agent_type":"general-purpose"}`)
	if st.Kittens != 2 {
		t.Errorf("kittens = %d, want 2", st.Kittens)
	}

	// The nag, in the middle of a real session, must change nothing.
	before := st.Pose
	st = step(at(4), `{"hook_event_name":"Notification","session_id":"e2e-0001","notification_type":"idle_prompt","message":"waiting"}`)
	if st.Pose != before {
		t.Errorf("the idle nag moved the pose from %v to %v", before, st.Pose)
	}

	st = step(at(5), `{"hook_event_name":"Notification","session_id":"e2e-0001","notification_type":"permission_prompt","message":"Claude needs your permission to use Bash"}`)
	if st.Pose != companion.NeedsYou {
		t.Errorf("pose = %v, want NeedsYou", st.Pose)
	}
	if st.Bubble == "" {
		t.Error("a permission prompt should carry its message into the bubble")
	}

	st = step(at(6), `{"hook_event_name":"PostToolUseFailure","session_id":"e2e-0001","tool_name":"Bash","tool_use_id":"t3","error":"exit status 1","tool_input":{"command":"go test ./..."}}`)
	if st.Pose != companion.Worried {
		t.Errorf("pose = %v, want Worried", st.Pose)
	}

	// Worried persists until the user comes back, and a new prompt clears it.
	st = step(at(400), `{"hook_event_name":"UserPromptSubmit","session_id":"e2e-0001","prompt":"fix it"}`)
	if st.Pose == companion.Worried {
		t.Error("a new prompt should clear the worry")
	}
}

func payload(t *testing.T, raw string) hookPayload {
	t.Helper()
	var p hookPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("bad test payload: %v", err)
	}
	return p
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// --- regressions from the 2026-08-31 review ---

// The single most common way a secret appears literally in a shell line is the
// leading assignment form. Taking token zero published it verbatim.
func TestShellRedactionSurvivesLeadingAssignments(t *testing.T) {
	cases := []struct{ cmd, want string }{
		{"PGPASSWORD=hunter2 psql -h db -c 'select 1'", "psql"},
		{"VERCEL_TOKEN=abc123 vercel deploy --prod", "vercel"},
		{"sudo PGPASSWORD=hunter2 /usr/bin/psql -c x", "psql"},
		{"env AWS_SECRET_ACCESS_KEY=wJalr aws s3 ls", "aws"},
		{"GITHUB_TOKEN=ghp_x nohup gh pr list", "gh"},
		{"API_KEY=sk-live-9911", "env"},
		{"go test ./...", "go"},
		{"/opt/homebrew/bin/rg --json pattern", "rg"},
	}
	for _, c := range cases {
		got := program(c.cmd)
		if got != c.want {
			t.Errorf("program(%q) = %q, want %q", c.cmd, got, c.want)
		}
		// Whatever it returns must not carry any part of a secret.
		for _, leak := range []string{"hunter2", "abc123", "wJalr", "ghp_", "sk-live"} {
			if contains(got, leak) {
				t.Errorf("program(%q) = %q leaks %q", c.cmd, got, leak)
			}
		}
	}
}

// A URL's query string is where access tokens live.
func TestWebTargetsDropCredentials(t *testing.T) {
	cases := []struct{ raw, want string }{
		{"https://api.example.com/v1/me?access_token=ghp_16C7e42F292c", "api.example.com/v1"},
		{"https://user:hunter2@intranet.example.com/secret", "intranet.example.com/secret"},
		{"https://example.com/a/b/c#frag", "example.com/a"},
	}
	for _, c := range cases {
		if got := safeURL(c.raw); got != c.want {
			t.Errorf("safeURL(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
	got := translate(payload(t, `{"hook_event_name":"PreToolUse","session_id":"s","tool_name":"WebFetch",
		"tool_use_id":"t","tool_input":{"url":"https://api.example.com/v1/me?access_token=ghp_16C7e42F292c"}}`))
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	if line := string(event.Encode(got[0])); contains(line, "ghp_") || contains(line, "access_token") {
		t.Errorf("the encoded event leaked the token: %s", line)
	}
}

// A PostToolUse carrying a huge tool_response used to be discarded whole --
// taking tool_use_id with it, so the matching tool_start never closed and the
// sea stayed up forever.
func TestOversizePayloadStillClosesItsTool(t *testing.T) {
	huge := strings.Repeat("x", (1<<20)+4096)
	raw := `{"hook_event_name":"PostToolUse","session_id":"sess-9","tool_name":"Bash",` +
		`"tool_use_id":"tu-42","duration_ms":8123,"tool_response":"` + huge + `"}`

	var p hookPayload
	if json.Unmarshal([]byte(raw[:1<<20]), &p) == nil {
		t.Fatal("expected the truncated payload not to parse")
	}
	salvage([]byte(raw[:1<<20]), &p)

	if p.Event != "PostToolUse" || p.ToolUseID != "tu-42" {
		t.Fatalf("salvage lost the identity: %+v", p)
	}
	if p.Session != "sess-9" || p.DurationM != 8123 {
		t.Errorf("salvage lost fields: session=%q ms=%d", p.Session, p.DurationM)
	}
	got := translate(p)
	if len(got) != 1 || got[0].Kind != event.ToolEnd || got[0].ID != "tu-42" {
		t.Errorf("an oversize PostToolUse must still close its tool, got %+v", got)
	}
}

// Auto-compaction re-announces the session; the source has to reach the reducer.
func TestSessionStartCarriesItsSource(t *testing.T) {
	for _, src := range []string{"startup", "compact", "resume", "fork", "clear"} {
		got := translate(payload(t, `{"hook_event_name":"SessionStart","session_id":"s","source":"`+src+`"}`))
		if len(got) != 1 || got[0].Text != src {
			t.Errorf("source %q did not reach the event: %+v", src, got)
		}
	}
}
