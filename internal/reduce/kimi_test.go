package reduce

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
)

// Kimi Code CLI's hooks, measured 2026-09-17: a subagent is named by its
// PROFILE, its tool calls and its Stop carry no agent at all, and Esc fires
// Interrupt in place of Stop. These are the reducer's rules for that, keyed
// on Src == "kimi" and nothing else.

func kimiAt(sec float64) time.Time {
	return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC).Add(time.Duration(sec * float64(time.Second)))
}

func TestKimiTwoSubagentsOfOneProfileAreTwoKittens(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt, Src: "kimi"}, kimiAt(0))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "explore", Src: "kimi"}, kimiAt(1))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "explore", Src: "kimi"}, kimiAt(1.1))
	if k := r.State(kimiAt(2)).Kittens; k != 2 {
		t.Fatalf("two parallel explores: %d kittens, want 2", k)
	}
	// The first to end is the oldest; the litter shrinks by one, after the dwell.
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "explore", Src: "kimi"}, kimiAt(70))
	if k := r.State(kimiAt(71)).Kittens; k != 1 {
		t.Fatalf("after one SubagentStop: %d kittens, want 1", k)
	}
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "explore", Src: "kimi"}, kimiAt(72))
	if k := r.State(kimiAt(73)).Kittens; k != 0 {
		t.Fatalf("after both: %d kittens, want 0", k)
	}
	// A Claude Code event with the same name takes none of this path.
	c := New("c")
	c.Apply(event.Event{Kind: event.SubStart, Agent: "a1", Src: "claude"}, kimiAt(0))
	c.Apply(event.Event{Kind: event.SubStart, Agent: "a1", Src: "claude"}, kimiAt(1))
	if k := c.State(kimiAt(2)).Kittens; k != 1 {
		t.Fatalf("claude: one id twice is %d kittens, want 1", k)
	}
}

func TestKimiAStopWithASubagentOpenIsNotTheTurnsEnd(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt, Src: "kimi"}, kimiAt(0))
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpSub, Tool: "Agent", ID: "t1", Src: "kimi"}, kimiAt(1))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "explore", Src: "kimi"}, kimiAt(1.1))
	// The subagent's own Stop, ~30 ms before its SubagentStop, as measured.
	r.Apply(event.Event{Kind: event.Done, Src: "kimi"}, kimiAt(10))
	if st := r.State(kimiAt(10.5)); st.Pose == companion.Done || !st.Act.Working {
		t.Fatalf("the subagent's Stop ended the turn: pose %v working %v", st.Pose, st.Act.Working)
	}
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "explore", Src: "kimi"}, kimiAt(10.03))
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpSub, Tool: "Agent", ID: "t1", Src: "kimi"}, kimiAt(10.05))
	// The turn's own Stop, with nothing open.
	r.Apply(event.Event{Kind: event.Done, Src: "kimi"}, kimiAt(12))
	if st := r.State(kimiAt(12.5)); st.Pose != companion.Done {
		t.Fatalf("the turn's Stop did not ring done: pose %v", st.Pose)
	}
}

func TestKimiToolsDuringAForegroundAgentCallAreTheSubagents(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt, Src: "kimi"}, kimiAt(0))
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpShell, Tool: "Bash", ID: "b1", Src: "kimi"}, kimiAt(1))
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpShell, Tool: "Bash", ID: "b1", Src: "kimi"}, kimiAt(1.5))
	steps := r.State(kimiAt(2)).Steps
	if steps != 1 {
		t.Fatalf("a main-thread tool paced %d steps, want 1", steps)
	}
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpSub, Tool: "Agent", ID: "a1", Src: "kimi"}, kimiAt(3))
	r.Apply(event.Event{Kind: event.SubStart, Agent: "explore", Src: "kimi"}, kimiAt(3.1))
	// The subagent reads a file and fails a command; neither carries an agent.
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpRead, Tool: "Read", ID: "r1", Src: "kimi"}, kimiAt(4))
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpRead, Tool: "Read", ID: "r1", Src: "kimi"}, kimiAt(4.2))
	r.Apply(event.Event{Kind: event.Error, Op: event.OpShell, Tool: "Bash", ID: "b2", Src: "kimi"}, kimiAt(5))
	st := r.State(kimiAt(6))
	if st.Steps != steps+1 {
		// The Agent call itself is a main-thread step; the subagent's read is not.
		t.Fatalf("subagent tools paced the companion: steps %d, want %d", st.Steps, steps+1)
	}
	if st.Pose == companion.Worried {
		t.Fatalf("a subagent's failure worried the companion")
	}
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "explore", Src: "kimi"}, kimiAt(7))
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpSub, Tool: "Agent", ID: "a1", Src: "kimi"}, kimiAt(7.1))
	// With the call closed, the main thread's own failure worries as ever.
	r.Apply(event.Event{Kind: event.Error, Op: event.OpShell, Tool: "Bash", ID: "b3", Src: "kimi"}, kimiAt(8))
	if st := r.State(kimiAt(9)); st.Pose != companion.Worried {
		t.Fatalf("a main-thread failure after the call did not worry: %v", st.Pose)
	}
}

func TestAnInterruptEndsTheTurnQuietly(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt, Src: "kimi"}, kimiAt(0))
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpShell, Tool: "Bash", ID: "b1", Src: "kimi"}, kimiAt(1))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?", Src: "kimi"}, kimiAt(1.1))
	r.Apply(event.Event{Kind: event.Interrupt, Text: "cancelled", Src: "kimi"}, kimiAt(5))
	st := r.State(kimiAt(5.5))
	if st.Act.Working {
		t.Fatalf("the turn is still open after an interrupt")
	}
	if st.Pose == companion.Done || st.Bubble != "" {
		t.Fatalf("an interrupt rang: pose %v bubble %q", st.Pose, st.Bubble)
	}
	// And no knock later either: doneAt was never set.
	if st := r.State(kimiAt(6)); st.Pose == companion.Done {
		t.Fatalf("a done pose after an interrupt")
	}
}

func TestAnAnswerClearsTheAskAndNothingElse(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt, Src: "kimi"}, kimiAt(0))
	r.Apply(event.Event{Kind: event.ToolStart, Op: event.OpShell, Tool: "Bash", ID: "b1", Src: "kimi"}, kimiAt(1))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?", Src: "kimi"}, kimiAt(1.1))
	before := r.State(kimiAt(2))
	if before.Pose != companion.NeedsYou || before.Bubble != "allow Bash?" {
		t.Fatalf("the ask did not raise: pose %v bubble %q", before.Pose, before.Bubble)
	}
	r.Apply(event.Event{Kind: event.Answered, Text: "approved", Src: "kimi"}, kimiAt(5))
	st := r.State(kimiAt(5.5))
	if st.Pose == companion.NeedsYou || st.Bubble != "" {
		t.Fatalf("the answer left the ask up: pose %v bubble %q", st.Pose, st.Bubble)
	}
	// The command is still running: the turn stays open, the pace did not
	// step, and no done rang.
	if !st.Act.Working || st.Steps != before.Steps || st.Pose == companion.Done {
		t.Fatalf("an answer did more than clear the ask: working %v steps %d then %d pose %v", st.Act.Working, before.Steps, st.Steps, st.Pose)
	}
	// An answer with no ask up changes nothing: the done words stay.
	r.Apply(event.Event{Kind: event.Done, Text: "all set"}, kimiAt(10))
	r.Apply(event.Event{Kind: event.Answered, Text: "approved"}, kimiAt(11))
	if st := r.State(kimiAt(12)); st.Bubble != "all set" {
		t.Fatalf("an idle answer cleared the done words: %q", st.Bubble)
	}
}
