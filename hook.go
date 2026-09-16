package main

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// hookPayload is the subset of Claude Code's hook input we read. The full
// schema is recorded in notes/claude-hooks-verified.md, extracted from the
// binary itself rather than from documentation.
//
// Everything is optional on purpose. This struct is fed by a program that
// ships new hook events regularly; a missing field must produce a quieter
// scene, never an error and never a crash inside the agent's turn.
type hookPayload struct {
	Event   string `json:"hook_event_name"`
	Session string `json:"session_id"`
	CWD     string `json:"cwd"`
	// Transcript is the session's JSONL, which every hook names; the spend
	// counter is summed off it (internal/spend).
	Transcript string `json:"transcript_path"`
	AgentID    string `json:"agent_id"`
	AgentType  string `json:"agent_type"`

	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolUseID string          `json:"tool_use_id"`
	DurationM int64           `json:"duration_ms"`

	// Kimi Code CLI (0.39) sends the same event names as Claude Code with
	// two keys spelled differently: the call id and the subagent's name.
	// Read off the binary's embedded source, not observed (s35).
	ToolCallID string `json:"tool_call_id"`
	AgentName  string `json:"agent_name"`

	// Hermes Agent (0.14) sends snake_case events of its own, and carries
	// everything past the tool fields under "extra" (agent/shell_hooks.py).
	Extra json.RawMessage `json:"extra"`

	NotificationType string `json:"notification_type"`
	Message          string `json:"message"`

	Error       string `json:"error"`
	IsInterrupt bool   `json:"is_interrupt"`

	LastAssistant string `json:"last_assistant_message"`
	Prompt        string `json:"prompt"`
	Source        string `json:"source"`
	Reason        string `json:"reason"`
	Trigger       string `json:"trigger"`
}

// needsYouTypes are the notification types that mean a human is actually being
// asked for something.
//
// This is an ALLOW list, and the direction matters. A deny list ("everything
// except idle_prompt") fails open: the day Claude Code adds a new chatty
// notification type, the cat starts knocking for it. An allow list fails
// closed -- a new type is silently ignored until we decide it deserves the
// bubble, which is the harmless direction to be wrong in.
var needsYouTypes = map[string]bool{
	"permission_prompt":        true,
	"worker_permission_prompt": true,
	"agent_needs_input":        true,
	"elicitation_dialog":       true,
	"elicitation_url_dialog":   true,
}

// maxPayload bounds what we read from a hook. A PostToolUse carrying the full
// output of a build can be far larger than anything we want to parse.
const maxPayload = 1 << 20

// salvage pulls the flat scalar fields out of a payload that would not parse.
// Deliberately dumb: it only recognises `"key": "value"` and `"key": number`
// at any depth, which is enough because every field here is unique by name and
// appears before the bulky ones.
func salvage(b []byte, p *hookPayload) {
	str := func(key string) string {
		m := regexp.MustCompile(`"` + key + `"\s*:\s*"((?:[^"\\]|\\.)*)"`).FindSubmatch(b)
		if m == nil {
			return ""
		}
		var out string
		if json.Unmarshal(append(append([]byte{'"'}, m[1]...), '"'), &out) != nil {
			return ""
		}
		return out
	}
	num := func(key string) int64 {
		m := regexp.MustCompile(`"` + key + `"\s*:\s*(-?[0-9]+)`).FindSubmatch(b)
		if m == nil {
			return 0
		}
		var out int64
		if json.Unmarshal(m[1], &out) != nil {
			return 0
		}
		return out
	}
	if p.Event == "" {
		p.Event = str("hook_event_name")
	}
	if p.Session == "" {
		p.Session = str("session_id")
	}
	if p.Transcript == "" {
		p.Transcript = str("transcript_path")
	}
	if p.ToolName == "" {
		p.ToolName = str("tool_name")
	}
	if p.ToolUseID == "" {
		p.ToolUseID = str("tool_use_id")
	}
	if p.AgentID == "" {
		p.AgentID = str("agent_id")
	}
	if p.NotificationType == "" {
		p.NotificationType = str("notification_type")
	}
	if p.DurationM == 0 {
		p.DurationM = num("duration_ms")
	}
}

// runHook is the adapter. It reads one hook payload on stdin, emits zero or
// more protocol events, and exits 0.
//
// Three rules govern everything here, all of them about not damaging the thing
// we are decorating:
//
//   - Never write to stdout. On UserPromptSubmit and SessionStart a hook's
//     stdout is fed to the model as context, so a stray print would inject
//     text into the agent's conversation.
//   - Never exit non-zero. A non-zero hook shows stderr to the user, and on
//     some events blocks the turn.
//   - Never take long. This runs on every tool call.
func runHook(args []string) {
	// Whatever happens below -- a wedged filesystem, a socket that will not
	// answer, a panic -- this process is gone in 250ms and the agent never
	// notices. It is the only guarantee here that does not depend on the
	// rest of the code being correct.
	watchdog := time.AfterFunc(250*time.Millisecond, func() { os.Exit(0) })
	defer watchdog.Stop()
	defer func() { recover() }()

	var p hookPayload
	if b, err := io.ReadAll(io.LimitReader(os.Stdin, maxPayload)); err == nil {
		if json.Unmarshal(b, &p) != nil {
			// The payload did not parse -- almost always because a big
			// tool_response pushed it past the cap and the read cut it
			// mid-object. Dropping it whole is the expensive failure: the
			// tool_use_id goes with it, so the matching tool_start is never
			// closed and the sea stays up forever. The scalar fields we need
			// sit near the front, ahead of the payload that made it big.
			salvage(b, &p)
		}
	}
	if p.Event == "" && len(args) > 0 {
		// The installed command names the event too, so a payload we could
		// not parse still produces the right kind of event.
		p.Event = args[0]
	}
	// `xscapes hook <Event> [agent]`: the installer writes which agent's
	// hooks these are, because Kimi's event names are Claude Code's.
	src := "claude"
	if len(args) > 1 && args[1] != "" {
		src = args[1]
	}
	if p.Session == "" {
		p.Session = event.SessionFromEnv()
	}
	// Kimi spells two keys differently; fold them in so translate reads one
	// shape.
	if p.ToolUseID == "" {
		p.ToolUseID = p.ToolCallID
	}
	if p.AgentID == "" && p.AgentName != "" {
		p.AgentID = p.AgentName
	}
	if p.AgentType == "" {
		p.AgentType = p.AgentName
	}

	for _, e := range translate(p) {
		e.Session = p.Session
		e.Src = src
		e.Transcript = p.Transcript
		if e.Agent == "" {
			e.Agent = p.AgentID
		}
		if e.AgentType == "" {
			e.AgentType = p.AgentType
		}
		_, _ = event.Emit(e)
	}
}

// hermesExtra is what Hermes puts under "extra": every keyword argument the
// hook was fired with, minus the tool fields. Names from
// website/docs/user-guide/features/hooks.md in the installed checkout (0.14).
type hermesExtra struct {
	UserMessage       string `json:"user_message"`
	AssistantResponse string `json:"assistant_response"`
	DurationMS        int64  `json:"duration_ms"`
	ChildRole         string `json:"child_role"`
	ChildStatus       string `json:"child_status"`
	ChildSessionID    string `json:"child_session_id"`
	Command           string `json:"command"`
	Reason            string `json:"reason"`
}

func (p hookPayload) hermes() hermesExtra {
	var x hermesExtra
	if len(p.Extra) > 0 {
		_ = json.Unmarshal(p.Extra, &x)
	}
	return x
}

// translate maps one hook payload to protocol events.
func translate(p hookPayload) []event.Event {
	switch p.Event {

	case "SessionStart":
		_ = event.SetCurrent(p.Session)
		// Source rides along so the reducer can tell a new conversation from
		// an auto-compaction re-announcing the same one mid-turn.
		return []event.Event{{Kind: event.SessionStart, Text: p.Source}}

	case "SessionEnd":
		return []event.Event{{Kind: event.SessionEnd, Text: p.Reason}}

	case "UserPromptSubmit":
		return []event.Event{{Kind: event.Prompt, Text: p.Prompt}}

	case "PreToolUse":
		op, target := classify(p.ToolName, p.ToolInput)
		return []event.Event{{
			Kind: event.ToolStart, Op: op, Tool: p.ToolName,
			Target: target, ID: p.ToolUseID,
		}}

	case "PostToolUse":
		op, target := classify(p.ToolName, p.ToolInput)
		e := event.Event{
			Kind: event.ToolEnd, Op: op, Tool: p.ToolName,
			Target: target, ID: p.ToolUseID, MS: p.DurationM,
		}
		// A checklist is its own channel, not a flavour of tool use, so it gets
		// its own event rather than counts bolted onto this one. The protocol
		// has carried `todo` with n/of since the beginning and nothing has ever
		// emitted it: TodoWrite was classified as an op and stopped there, so
		// the reducer's Todo case could only ever be reached by hand from
		// `xscapes emit`. Two halves of one feature that were never joined.
		if done, total, ok := todoCounts(p.ToolName, p.ToolInput); ok {
			return []event.Event{e, {Kind: event.Todo, N: done, Of: total}}
		}
		return []event.Event{e}

	case "PostToolUseFailure":
		op, target := classify(p.ToolName, p.ToolInput)
		// An interrupt is the user pressing escape. Treating that as a
		// failure would put the cat in the worried pose every time Lucas
		// changes his mind, which is both wrong and the fastest way to
		// teach someone to ignore the signal.
		kind := event.Error
		if p.IsInterrupt {
			kind = event.ToolEnd
		}
		return []event.Event{{
			Kind: kind, Op: op, Tool: p.ToolName, Target: target,
			ID: p.ToolUseID, MS: p.DurationM, Detail: firstLine(p.Error),
		}}

	case "PermissionRequest":
		op, target := classify(p.ToolName, p.ToolInput)
		return []event.Event{{
			Kind: event.NeedsInput, Op: op, Tool: p.ToolName, Target: target,
			Text: "allow " + p.ToolName + "?",
		}}

	case "Notification":
		// The whole 60-second-nag problem, solved by reading a field.
		// idle_prompt IS the nag: 2,220 of them at exactly sixty seconds
		// after Stop in the measured log. It is not in the allow list, so
		// it lands here and is dropped.
		switch {
		case needsYouTypes[p.NotificationType]:
			return []event.Event{{Kind: event.NeedsInput, Text: firstLine(p.Message)}}
		case p.NotificationType == "agent_completed":
			return []event.Event{{Kind: event.Done, Text: firstLine(p.Message)}}
		}
		return nil

	case "Stop":
		return []event.Event{{Kind: event.Done, Text: firstLine(p.LastAssistant)}}

	case "SubagentStart":
		return []event.Event{{
			Kind: event.SubStart, Op: event.OpSub,
			Agent: p.AgentID, AgentType: p.AgentType,
		}}

	case "SubagentStop":
		return []event.Event{{
			Kind: event.SubEnd, Op: event.OpSub,
			Agent: p.AgentID, AgentType: p.AgentType,
		}}

	case "PreCompact":
		return []event.Event{{Kind: event.Compact, Text: p.Trigger}}

	// Kimi Code CLI. Its event list is Claude Code's plus a few; the ones
	// that mean something here:
	case "StopFailure":
		// The turn ended in a failure of the agent itself (a provider error,
		// a model that would not answer): the companion worries, and the
		// sand says why. Interrupt is the user pressing escape and stays
		// silent, as PostToolUseFailure's is_interrupt does.
		return []event.Event{{Kind: event.Error, Detail: firstLine(p.Error)}}

	// Hermes Agent. pre_llm_call and post_llm_call fire ONCE PER TURN
	// (hooks.md: "before the tool-calling loop" / "after the tool-calling
	// loop completes"), which is what makes them the prompt and the done.
	case "on_session_start":
		_ = event.SetCurrent(p.Session)
		return []event.Event{{Kind: event.SessionStart, Text: "startup"}}
	case "on_session_reset":
		_ = event.SetCurrent(p.Session)
		return []event.Event{{Kind: event.SessionStart, Text: "clear"}}
	case "on_session_end", "on_session_finalize":
		return []event.Event{{Kind: event.SessionEnd}}
	case "pre_llm_call":
		return []event.Event{{Kind: event.Prompt, Text: firstLine(p.hermes().UserMessage)}}
	case "post_llm_call":
		return []event.Event{{Kind: event.Done, Text: firstLine(p.hermes().AssistantResponse)}}
	case "pre_tool_call":
		op, target := classify(p.ToolName, p.ToolInput)
		return []event.Event{{Kind: event.ToolStart, Op: op, Tool: p.ToolName, Target: target}}
	case "post_tool_call":
		op, target := classify(p.ToolName, p.ToolInput)
		return []event.Event{{Kind: event.ToolEnd, Op: op, Tool: p.ToolName, Target: target, MS: p.hermes().DurationMS}}
	case "pre_approval_request":
		op, target := classify(p.ToolName, p.ToolInput)
		return []event.Event{{Kind: event.NeedsInput, Op: op, Tool: p.ToolName, Target: target, Text: "allow " + p.ToolName + "?"}}
	case "subagent_stop":
		// Hermes has no subagent START event, only this. A litter that
		// never gets a start never shows, so the stop is announced as a
		// start and an end together: the reducer's dwell (KittenDwell)
		// keeps the arrival on screen for a minute, which is the same
		// treatment a subagent that lived seven seconds gets from Claude.
		x := p.hermes()
		id := x.ChildSessionID
		if id == "" {
			id = x.ChildRole
		}
		return []event.Event{
			{Kind: event.SubStart, Op: event.OpSub, Agent: id, AgentType: x.ChildRole},
			{Kind: event.SubEnd, Op: event.OpSub, Agent: id, AgentType: x.ChildRole},
		}
	}
	return nil
}

// todoCounts reads a TodoWrite call and reports how much of the list is done.
//
// ⚠ The payload shape here is INFERRED, not measured. notes/claude-hooks-verified.md
// covers every hook this program uses and says nothing about TodoWrite's
// tool_input, because in 13,682 recorded tool events and every transcript since
// the hook was installed, TodoWrite has been called exactly zero times. So this
// is written to fail quiet rather than fail wrong: an unrecognised shape
// reports ok=false and the sky shows nothing, which is the same as today.
//
// It counts "completed" and treats everything else as outstanding, so a status
// value nobody anticipated lands on "not done" rather than on "done".
func todoCounts(tool string, in json.RawMessage) (done, total int, ok bool) {
	if tool != "TodoWrite" || len(in) == 0 {
		return 0, 0, false
	}
	var p struct {
		Todos []struct {
			Status string `json:"status"`
		} `json:"todos"`
	}
	if err := json.Unmarshal(in, &p); err != nil || len(p.Todos) == 0 {
		return 0, 0, false
	}
	for _, t := range p.Todos {
		if strings.EqualFold(t.Status, "completed") {
			done++
		}
	}
	return done, len(p.Todos), true
}

// toolInput is the handful of tool_input keys that name a subject. Claude Code
// tools are not uniform about it, so this is a union rather than a per-tool
// table -- a new tool with a file_path gets a sensible sand line for free.
type toolInput struct {
	FilePath    string `json:"file_path"`
	Path        string `json:"path"`
	NotebookP   string `json:"notebook_path"`
	Pattern     string `json:"pattern"`
	Command     string `json:"command"`
	URL         string `json:"url"`
	Query       string `json:"query"`
	Description string `json:"description"`
	SubagentTyp string `json:"subagent_type"`
	Prompt      string `json:"prompt"`
}

// classify puts a tool into one of the coarse ops and finds its subject.
//
// This is not the per-tool taxonomy the brief killed: nothing downstream keys
// a *visual* off the op. It exists so the sand can say "read" rather than
// "Read", and so the reducer can tell a shell command from a file read when
// deciding how long the water stays up.
func classify(tool string, raw json.RawMessage) (event.Op, string) {
	var in toolInput
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &in)
	}

	subject := func() string {
		switch {
		case in.FilePath != "":
			return shorten(in.FilePath)
		case in.NotebookP != "":
			return shorten(in.NotebookP)
		case in.Path != "":
			return shorten(in.Path)
		case in.Pattern != "":
			return in.Pattern
		case in.URL != "":
			return safeURL(in.URL)
		case in.Query != "":
			return in.Query
		}
		return ""
	}

	switch tool {
	case "Read", "NotebookRead":
		return event.OpRead, subject()
	case "Write":
		return event.OpWrite, subject()
	case "Edit", "NotebookEdit", "MultiEdit":
		return event.OpEdit, subject()
	case "Grep", "Glob", "LS":
		return event.OpSearch, subject()
	case "Bash", "BashOutput", "KillShell":
		// The command, not its arguments. A shell line is the one tool input
		// that routinely contains a secret, and this string is written to a
		// file and painted on a screen.
		return event.OpShell, program(in.Command)
	case "WebFetch":
		return event.OpWeb, safeURL(in.URL)
	case "WebSearch":
		// A search query is the user's words, not a credential, but it is also
		// not something to write across the beach in full.
		return event.OpWeb, trimTo(in.Query, 40)
	case "Agent", "Task":
		s := in.SubagentTyp
		if s == "" {
			s = firstLine(in.Description)
		}
		return event.OpSub, s
	case "TodoWrite":
		return event.OpTodo, ""
	}
	if strings.HasPrefix(tool, "mcp__") {
		return event.OpMCP, strings.TrimPrefix(tool, "mcp__")
	}
	// Hermes Agent's tools (tools/*.py in the installed checkout). The
	// input key names are inferred from the tool names, unverified against a
	// live payload; a miss costs the subject, never the verb.
	switch tool {
	case "read_file":
		return event.OpRead, subject()
	case "write_file":
		return event.OpWrite, subject()
	case "patch":
		return event.OpEdit, subject()
	case "search_files", "session_search":
		return event.OpSearch, subject()
	case "terminal", "process":
		return event.OpShell, program(in.Command)
	case "web_search", "x_search":
		return event.OpWeb, trimTo(in.Query, 40)
	case "web_extract":
		return event.OpWeb, safeURL(in.URL)
	case "delegate_task", "mixture_of_agents":
		return event.OpSub, firstLine(in.Description)
	case "todo":
		return event.OpTodo, ""
	}
	return event.OpOther, subject()
}

// shorten makes a path readable in a beach's worth of columns: relative to the
// working directory when it is under it, and ~ for home.
func shorten(p string) string {
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, p); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

// program extracts the command name from a shell line and nothing else.
// `psql "postgresql://user:password@host"` must reach the sand as `psql`.
//
// The first token is NOT the program often enough to matter: the single most
// common way a secret appears literally in a shell line is the leading
// assignment form, `PGPASSWORD=hunter2 psql ...`, and taking token zero
// published it verbatim. Leading assignments, and the wrappers that usually
// carry them, are skipped rather than truncated -- half a secret is a secret.
func program(cmd string) string {
	for _, tok := range strings.Fields(cmd) {
		if isAssignment(tok) {
			continue
		}
		switch tok {
		case "sudo", "env", "command", "nohup", "time", "exec", "doas":
			continue
		}
		if i := strings.LastIndex(tok, "/"); i >= 0 && i+1 < len(tok) {
			tok = tok[i+1:]
		}
		if len(tok) > 24 {
			tok = tok[:24]
		}
		return tok
	}
	// Nothing but assignments and wrappers. Naming the shape is safe; naming
	// the variable is not, because the name is often the giveaway.
	if strings.TrimSpace(cmd) != "" {
		return "env"
	}
	return ""
}

// isAssignment reports whether tok is a leading NAME=VALUE.
func isAssignment(tok string) bool {
	i := strings.IndexByte(tok, '=')
	if i <= 0 {
		return false
	}
	for j, r := range tok[:i] {
		ok := r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (j > 0 && r >= '0' && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}

// safeURL keeps enough of a URL to say where the agent went and drops
// everything that carries a credential: the query string (access tokens live
// there), the fragment, and any userinfo. It is also the more legible line.
func safeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		// Not parseable as a URL: keep the scheme-ish head only.
		if i := strings.IndexAny(raw, "?#"); i >= 0 {
			raw = raw[:i]
		}
		return trimTo(raw, 60)
	}
	out := u.Host
	if p := strings.TrimPrefix(u.Path, "/"); p != "" {
		if i := strings.IndexByte(p, '/'); i >= 0 {
			p = p[:i]
		}
		out += "/" + p
	}
	return trimTo(out, 60)
}

func trimTo(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if len([]rune(s)) > 120 {
		s = string([]rune(s)[:119]) + "…"
	}
	return s
}
