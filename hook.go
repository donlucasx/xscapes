package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donlucasx/asciiscapes/internal/event"
)

// hookPayload is the subset of Claude Code's hook input we read. The full
// schema is recorded in notes/claude-hooks-verified.md, extracted from the
// binary itself rather than from documentation.
//
// Everything is optional on purpose. This struct is fed by a program that
// ships new hook events regularly; a missing field must produce a quieter
// scene, never an error and never a crash inside the agent's turn.
type hookPayload struct {
	Event     string `json:"hook_event_name"`
	Session   string `json:"session_id"`
	CWD       string `json:"cwd"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`

	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
	ToolUseID string          `json:"tool_use_id"`
	DurationM int64           `json:"duration_ms"`

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
	if b, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20)); err == nil {
		_ = json.Unmarshal(b, &p)
	}
	if p.Event == "" && len(args) > 0 {
		// The installed command names the event too, so a payload we could
		// not parse still produces the right kind of event.
		p.Event = args[0]
	}
	if p.Session == "" {
		p.Session = event.SessionFromEnv()
	}

	for _, e := range translate(p) {
		e.Session = p.Session
		e.Src = "claude"
		e.Agent = p.AgentID
		if e.AgentType == "" {
			e.AgentType = p.AgentType
		}
		_, _ = event.Emit(e)
	}
}

// translate maps one hook payload to protocol events.
func translate(p hookPayload) []event.Event {
	switch p.Event {

	case "SessionStart":
		_ = event.SetCurrent(p.Session)
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
	}
	return nil
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
			return in.URL
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
		return event.OpShell, firstWord(in.Command)
	case "WebFetch", "WebSearch":
		return event.OpWeb, subject()
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

// firstWord keeps the program name from a shell command and drops everything
// after it. `psql "postgresql://user:password@host"` must reach the sand as
// `psql`, not as a credential.
func firstWord(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}
	for _, sep := range []string{" ", "\t", "\n"} {
		if i := strings.Index(cmd, sep); i > 0 {
			cmd = cmd[:i]
		}
	}
	if i := strings.LastIndex(cmd, "/"); i >= 0 && i+1 < len(cmd) {
		cmd = cmd[i+1:]
	}
	if len(cmd) > 24 {
		cmd = cmd[:24]
	}
	return cmd
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
