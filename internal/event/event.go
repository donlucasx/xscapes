// Package event is the waiting-layer protocol: one JSON object per line,
// written by an adapter, read by a scape.
//
// The scape never learns what a "hook" is and the adapter never learns what a
// wave is. Everything they share is in this file, which is why it is the piece
// worth being strict about -- a second adapter (Kimi, Hermes, a plain process
// watcher) should need nothing from the renderer but this.
package event

// Kind is what happened. It is an open string on the wire and a closed set of
// constants here, so an adapter from the future can send something we have not
// heard of and the engine ignores it rather than crashing or, worse, counting
// it as activity.
type Kind string

const (
	SessionStart Kind = "session_start"
	SessionEnd   Kind = "session_end"
	Prompt       Kind = "prompt"
	ToolStart    Kind = "tool_start"
	ToolEnd      Kind = "tool_end"
	Error        Kind = "error"
	TestPass     Kind = "test_pass"
	TestFail     Kind = "test_fail"
	Compact      Kind = "compact"
	NeedsInput   Kind = "needs_input"
	Done         Kind = "done"
	SubStart     Kind = "sub_start"
	SubEnd       Kind = "sub_end"
	Todo         Kind = "todo"
	Context      Kind = "context"
	// Interrupt is the user cutting the turn short (Esc). The turn is over
	// and nothing was finished: the scene settles and no cue rings, because
	// the person who pressed the key is the one who would have heard it.
	// Kimi Code CLI fires it IN PLACE OF Stop (measured 2026-09-17); an
	// adapter that did not send it would leave the sea up until TurnSilence.
	Interrupt Kind = "interrupt"
	// Answered is a pending ask resolved by the user: an approval given or
	// refused. Kimi Code CLI reports the moment (PermissionResult, measured
	// 2026-09-17); Claude Code never does, so there the next tool starting
	// is what clears the ask. It clears the ask and touches nothing else.
	Answered Kind = "answered"
)

// Kinds is every kind the reducer understands, in protocol order.
//
// It exists because `xscapes emit <word>` used to cast its first argument
// straight to a Kind with no validation: an unknown word went to the socket,
// was dropped by the reducer, and the command printed "sent ... to session".
// THREE JUNK `ask` EVENTS ARE IN HIS REAL LOG because of it -- `ask` is a
// reasonable guess, since `xscapes notify ask` accepts exactly that word.
var Kinds = []Kind{
	SessionStart, SessionEnd, Prompt, ToolStart, ToolEnd, Error,
	TestPass, TestFail, Compact, NeedsInput, Done, SubStart, SubEnd,
	Todo, Context, Interrupt, Answered,
}

// Known says the reducer has a meaning for this kind.
func Known(k Kind) bool {
	for _, v := range Kinds {
		if v == k {
			return true
		}
	}
	return false
}

// KindNames is Kinds as plain strings, for an error message that tells the
// caller what it could have said instead.
func KindNames() []string {
	out := make([]string, len(Kinds))
	for i, k := range Kinds {
		out[i] = string(k)
	}
	return out
}

// Op is the coarse class of a tool. The brief killed the per-tool weather
// taxonomy, and this is deliberately not a revival of it: nothing in the scene
// keys a *visual* off Op. It exists so the sand can say "read" instead of
// "Read", and so a 90-second shell command can be told from a 3ms file read
// when deciding how hard the sea is working.
type Op string

const (
	OpRead   Op = "read"
	OpWrite  Op = "write"
	OpEdit   Op = "edit"
	OpSearch Op = "search"
	OpShell  Op = "shell"
	OpWeb    Op = "web"
	OpSub    Op = "subagent"
	OpTodo   Op = "todo"
	OpMCP    Op = "mcp"
	OpOther  Op = "other"
)

// Event is one thing that happened. Everything past Kind is optional, and a
// consumer must treat a missing field as "not reported" rather than as zero --
// the difference matters for Frac, where 0 is a legitimate reading.
type Event struct {
	// V is the protocol major. It moves only when an existing key changes
	// meaning; adding keys and kinds is not a break. Absent reads as 1, so
	// `echo '{"kind":"done"}' | xscapes emit -` works by hand.
	V int `json:"v,omitempty"`

	// TS is Unix milliseconds at the emitter. The engine timestamps arrival
	// itself and orders by that, so a wrong clock on the emitter cannot make
	// the sea run backwards; TS is for logs and replay.
	TS      int64  `json:"ts,omitempty"`
	Session string `json:"session,omitempty"`
	Kind    Kind   `json:"kind"`

	// Agent is the subagent this event came from, empty for the main thread.
	// Claude Code puts agent_id on every payload that fires inside a subagent,
	// so a tool call can be attributed to the kitten that made it.
	Agent string `json:"agent,omitempty"`
	// AgentType labels a subagent ("general-purpose", "code-reviewer").
	AgentType string `json:"agent_type,omitempty"`

	Op Op `json:"op,omitempty"`
	// Tool is the adapter's own name for it ("Bash", "Edit") -- kept verbatim
	// because the sand shows it and a translated name would be a lie about
	// what the agent actually ran.
	Tool string `json:"tool,omitempty"`
	// Target is the file or subject, already shortened relative to the repo.
	Target string `json:"target,omitempty"`
	// Detail is the short result ("142 lines", "+18 -2", "ok 0.412s").
	Detail string `json:"detail,omitempty"`

	// ID correlates a tool_start with its tool_end, or a sub_start with its
	// sub_end. Without it a lost end event leaks a busy tool forever.
	ID string `json:"id,omitempty"`

	// MS is how long a finished tool took.
	MS int64 `json:"ms,omitempty"`

	// Frac is a 0..1 reading whose meaning depends on Kind: context used for
	// Context, completion for Todo. A pointer because 0 is meaningful.
	Frac *float64 `json:"frac,omitempty"`

	// N and Of carry counts (todos done / total).
	N  int `json:"n,omitempty"`
	Of int `json:"of,omitempty"`

	// Tokens is the session's running spend (unused by the Claude adapter,
	// whose spend is summed off the transcript; an adapter that knows its
	// total may send it); Window is the model's context window in tokens.
	// Both ride the Context event; zero means unknown.
	Tokens int64 `json:"tokens,omitempty"`
	Window int   `json:"window,omitempty"`

	// Text is human-facing: the prompt, the error, the message that goes in
	// the bubble. Truncated hard by the wire encoder.
	Text string `json:"text,omitempty"`

	// Transcript is the path of the agent's own transcript, when the adapter
	// knows it: the spend counter is summed off it (internal/spend).
	Transcript string `json:"transcript,omitempty"`

	// Src names the adapter ("claude", "manual"). Cheap, and it is how you
	// tell a hand-fired test event from a real one when a log looks wrong.
	Src string `json:"src,omitempty"`
}

// Known reports whether Kind is one this build understands. Unknown kinds are
// carried, logged and ignored -- never dropped silently at the transport,
// because the reason to look at the log is usually that something new arrived.
func (e Event) Known() bool {
	switch e.Kind {
	case SessionStart, SessionEnd, Prompt, ToolStart, ToolEnd, Error,
		TestPass, TestFail, Compact, NeedsInput, Done, SubStart, SubEnd,
		Todo, Context, Interrupt, Answered:
		return true
	}
	return false
}

// Busy reports whether this event is the agent doing work, which is the only
// question the sea asks.
func (e Event) Busy() bool {
	return e.Kind == ToolStart || e.Kind == ToolEnd || e.Kind == Prompt
}
