package reduce

import (
	"testing"
	"time"

	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
)

// base is an arbitrary fixed instant. Everything here is relative to it, so
// the tests are clock-free and run at whatever speed the machine manages.
var base = time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)

func at(sec float64) time.Time {
	return base.Add(time.Duration(sec * float64(time.Second)))
}

func TestIdleSeaIsFlat(t *testing.T) {
	r := New("s")
	if got := r.State(at(0)).Act.Level; got != 0 {
		t.Errorf("a session that has done nothing should be flat, got %.3f", got)
	}
	if got := r.State(at(0)).Pose; got != companion.Resting {
		t.Errorf("pose = %v, want Resting", got)
	}
}

// A single long shell command is the hardest case in the design: two events,
// ninety seconds apart, and the water must stay up for all of it. Encoding
// activity in event RATE would read this as idle.
func TestOneLongToolHoldsTheSeaUp(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1", Op: event.OpShell, Tool: "Bash"}, at(1))

	for _, sec := range []float64{5, 30, 60, 90} {
		st := r.State(at(sec))
		if st.Act.Level < FlightFloor-1e-9 {
			t.Errorf("at %.0fs into a running tool the sea fell to %.3f, below the %.2f floor",
				sec, st.Act.Level, FlightFloor)
		}
		if st.Pose != companion.Working {
			t.Errorf("at %.0fs the companion is %v, want Working", sec, st.Pose)
		}
	}

	r.Apply(event.Event{Kind: event.ToolEnd, ID: "t1", Op: event.OpShell, Tool: "Bash", MS: 90000}, at(91))
	r.Apply(event.Event{Kind: event.Done}, at(92))
	// Once the turn is closed the sea must actually settle, or "busy" stops
	// meaning anything.
	if got := r.State(at(92 + 5*TauFall.Seconds())).Act.Level; got > 0.05 {
		t.Errorf("sea did not settle after the turn closed: %.3f", got)
	}
}

func TestSeaRisesFastAndFallsSlow(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	for i := 0; i < 6; i++ {
		id := string(rune('a' + i))
		r.Apply(event.Event{Kind: event.ToolStart, ID: id}, at(float64(i)*0.2))
		r.Apply(event.Event{Kind: event.ToolEnd, ID: id}, at(float64(i)*0.2+0.1))
	}
	busy := r.State(at(1.5)).Act.Level
	if busy < 0.8 {
		t.Errorf("six quick tool calls should read as busy, got %.3f", busy)
	}

	r.Apply(event.Event{Kind: event.Done}, at(2))

	// One TauFall later the sea should be well down but not flat: a two
	// second gap between tools must not empty the ocean.
	half := r.State(at(2 + TauFall.Seconds())).Act.Level
	if half > busy*0.75 || half < 0.05 {
		t.Errorf("after one TauFall level is %.3f, expected a clear fall from %.3f but not to nothing", half, busy)
	}
}

// The precedence order is the whole point of having one: "working" and
// "needs you" are true simultaneously all the time.
func TestPosePrecedence(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t"}, at(1))
	if got := r.State(at(2)).Pose; got != companion.Working {
		t.Fatalf("pose = %v, want Working", got)
	}

	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(3))
	if got := r.State(at(3)).Pose; got != companion.NeedsYou {
		t.Errorf("needs_input must outrank working, got %v", got)
	}
	if got := r.State(at(3)).Bubble; got != "allow Bash?" {
		t.Errorf("bubble = %q", got)
	}

	r.Apply(event.Event{Kind: event.Error, Tool: "Bash", Detail: "exit 1"}, at(4))
	if got := r.State(at(4)).Pose; got != companion.Worried {
		t.Errorf("an error must outrank needs_input, got %v", got)
	}

	// Worried persists -- that is the locked behaviour. Only the user coming
	// back clears it, because a hook can tell us a command failed and can
	// never tell us the code is fixed.
	if got := r.State(at(4 + 10*TauFall.Seconds())).Pose; got != companion.Worried {
		t.Errorf("worried must persist, got %v", got)
	}
	r.Apply(event.Event{Kind: event.Prompt}, at(600))
	if got := r.State(at(600)).Pose; got == companion.Worried {
		t.Error("a new prompt should clear the worry")
	}
}

// An interrupt is the user pressing escape, not a failure. Treating it as one
// would put the cat in the worried pose every time Lucas changes his mind.
func TestToolStartClearsNeedsYou(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(1))
	if r.State(at(1)).Pose != companion.NeedsYou {
		t.Fatal("expected NeedsYou")
	}
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t"}, at(2))
	if got := r.State(at(2)).Pose; got != companion.Working {
		t.Errorf("a tool starting means the permission was answered; pose = %v", got)
	}
}

func TestSubagentsCountAndSurviveLostEvents(t *testing.T) {
	r := New("s")
	for _, id := range []string{"a", "b", "c"} {
		r.Apply(event.Event{Kind: event.SubStart, Agent: id}, at(0))
	}
	if got := r.State(at(1)).Kittens; got != 3 {
		t.Errorf("kittens = %d, want 3", got)
	}
	// A duplicate start must not double-count: the set is the point.
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a"}, at(2))
	if got := r.State(at(2)).Kittens; got != 3 {
		t.Errorf("duplicate start changed the count to %d", got)
	}
	r.Apply(event.Event{Kind: event.SubEnd, Agent: "b"}, at(3))
	if got := r.State(at(3)).Kittens; got != 2 {
		t.Errorf("kittens = %d, want 2", got)
	}
	// A stop event that never arrives must not strand a kitten forever.
	if got := r.State(at(3 + SubStale.Seconds() + 1)).Kittens; got != 0 {
		t.Errorf("stale subagents should age out, still showing %d", got)
	}
}

func TestTurnWithoutStopEventuallyCloses(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	if !r.State(at(1)).Act.Working {
		t.Fatal("expected working")
	}
	if r.State(at(TurnSilence.Seconds() + 1)).Act.Working {
		t.Error("a turn nothing closed must time out, or the cat works forever")
	}
}

func TestSandTail(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpRead, Tool: "Read",
		Target: "internal/auth/handler.go", Detail: "142 lines"}, at(0))
	r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpEdit, Tool: "Edit",
		Target: "internal/auth/handler.go", Detail: "+18 -2"}, at(1))

	lines := r.State(at(2)).Tail
	if len(lines) != 2 {
		t.Fatalf("tail has %d lines, want 2", len(lines))
	}
	if lines[0].Text != "read   internal/auth/handler.go  142 lines" {
		t.Errorf("first line = %q", lines[0].Text)
	}
	if lines[1].Age >= lines[0].Age {
		t.Error("newest line should be the least aged")
	}

	// The tail must not grow without bound, and must recede.
	for i := 0; i < 20; i++ {
		r.Apply(event.Event{Kind: event.ToolEnd, Op: event.OpRead, Target: "x.go"}, at(float64(3+i)))
	}
	if got := len(r.State(at(30)).Tail); got != TailLen {
		t.Errorf("tail length = %d, want %d", got, TailLen)
	}
	if got := len(r.State(at(30 + TailTTL.Seconds() + 1)).Tail); got != 0 {
		t.Errorf("the tide should take every line eventually, %d left", got)
	}
}

func TestContextOnlySetByContextEvents(t *testing.T) {
	r := New("s")
	if got := r.State(at(0)).Act.ContextUsed; got != 0 {
		t.Errorf("a fresh session must read as an empty context, got %.3f", got)
	}
	f := 0.42
	r.Apply(event.Event{Kind: event.Context, Frac: &f}, at(1))
	if got := r.State(at(1)).Act.ContextUsed; got != 0.42 {
		t.Errorf("context = %.3f, want 0.42", got)
	}
	// A tool event carries no reading and must not disturb it.
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t"}, at(2))
	if got := r.State(at(2)).Act.ContextUsed; got != 0.42 {
		t.Errorf("context moved to %.3f on an unrelated event", got)
	}
}

// --- regressions from the 2026-08-31 review ---

// The worst one. After any failed command the companion goes Worried, and the
// bubble used to be gated on the pose -- so a permission prompt the agent is
// actually blocked on went silent, and the session could sit waiting for input
// with nothing on screen saying so.
func TestWorriedDoesNotSwallowThePermissionPrompt(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.Error, Tool: "Bash", Detail: "exit 1"}, at(1))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(2))

	st := r.State(at(2))
	if st.Pose != companion.Worried {
		t.Errorf("pose = %v, want Worried (a break still outranks a question)", st.Pose)
	}
	if st.Bubble != "allow Bash?" {
		t.Errorf("bubble = %q; the agent is BLOCKED and the scene must say so even while worried", st.Bubble)
	}
}

func TestDoneKnockShowsThroughAWorry(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.Error, Tool: "Bash"}, at(1))
	r.Apply(event.Event{Kind: event.Done, Text: "tests still failing"}, at(2))
	if got := r.State(at(3)).Bubble; got != "tests still failing" {
		t.Errorf("bubble = %q, want the finish message", got)
	}
}

// A knock with no words is honest; a knock still showing the last permission
// request is a lie about what is being asked.
func TestDoneWithNoMessageDoesNotReuseTheOldBubble(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(0))
	r.Apply(event.Event{Kind: event.Done}, at(1))
	if got := r.State(at(2)).Bubble; got != "" {
		t.Errorf("bubble = %q, want empty", got)
	}
}

// Kill the agent while a Bash is running: no tool_end, no Stop, no session_end.
// The scene must eventually settle instead of insisting forever that the agent
// is hard at work.
func TestAKilledAgentEventuallySettles(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1", Op: event.OpShell}, at(1))

	// Still working a few minutes in -- a long command is normal.
	if st := r.State(at(120)); !st.Act.Working {
		t.Error("a two-minute command should still read as working")
	}
	// But not forever.
	late := at(FlightStale.Seconds() + TurnSilence.Seconds() + 60)
	st := r.State(late)
	if st.Act.Working {
		t.Error("a tool that never ended pinned the scene at working forever")
	}
	if st.Act.Level > 0.05 {
		t.Errorf("sea never settled: level %.3f", st.Act.Level)
	}
	if st.Pose != companion.Resting {
		t.Errorf("pose = %v, want Resting", st.Pose)
	}
}

// Auto-compaction re-announces the session mid-turn. That must not wipe the
// scene while the agent is still working.
func TestCompactDoesNotResetTheScene(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	for i := 0; i < 5; i++ {
		id := string(rune('a' + i))
		r.Apply(event.Event{Kind: event.ToolStart, ID: id}, at(float64(i)*0.2))
		r.Apply(event.Event{Kind: event.ToolEnd, ID: id, Op: event.OpRead, Target: "x.go"}, at(float64(i)*0.2+0.1))
	}
	r.Apply(event.Event{Kind: event.SubStart, Agent: "a1"}, at(2))
	before := r.State(at(2))

	r.Apply(event.Event{Kind: event.SessionStart, Text: "compact"}, at(3))
	after := r.State(at(3))

	if after.Kittens != before.Kittens {
		t.Errorf("compaction dropped the litter: %d -> %d", before.Kittens, after.Kittens)
	}
	if len(after.Tail) == 0 {
		t.Error("compaction wiped the sand")
	}
	if !after.Act.Working {
		t.Error("compaction stopped the turn while the agent was still working")
	}

	// A real new session still resets.
	r.Apply(event.Event{Kind: event.SessionStart, Text: "startup"}, at(4))
	if st := r.State(at(4)); st.Kittens != 0 || len(st.Tail) != 0 || st.Act.Working {
		t.Error("a genuinely new session must reset the scene")
	}
}

// TurnSilence has to measure silence, not turn length.
func TestALongBusyTurnDoesNotTimeOut(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	// Seven minutes of steady work, well past TurnSilence.
	for s := 10.0; s < 420; s += 10 {
		id := string(rune('a' + int(s)%26))
		r.Apply(event.Event{Kind: event.ToolStart, ID: id}, at(s))
		r.Apply(event.Event{Kind: event.ToolEnd, ID: id}, at(s+1))
	}
	// Then the thinking gap the TurnFloor constant exists for.
	if st := r.State(at(437)); !st.Act.Working {
		t.Error("a healthy seven-minute turn timed out mid-work")
	}
	// Genuine silence still closes it.
	if st := r.State(at(420 + TurnSilence.Seconds() + 30)); st.Act.Working {
		t.Error("a turn that really went quiet should close")
	}
}

// The brief locks done and needs_input as DISTINCT cues, and this was the
// oldest open gap against it: both used to raise the same NeedsYou pose and
// the same bubble.
func TestDoneAndNeedsInputAreDistinct(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(1))
	st := r.State(at(1))
	if st.Pose != companion.NeedsYou || !st.BubbleAsk {
		t.Errorf("needs_input: pose=%v ask=%v, want NeedsYou with an ask balloon", st.Pose, st.BubbleAsk)
	}

	r.Apply(event.Event{Kind: event.Done, Text: "tests passed"}, at(2))
	st = r.State(at(2))
	if st.Pose != companion.Done || st.BubbleAsk {
		t.Errorf("done: pose=%v ask=%v, want Done with the soft knock", st.Pose, st.BubbleAsk)
	}
	if st.Bubble != "tests passed" {
		t.Errorf("bubble = %q", st.Bubble)
	}

	// The knock is bounded; an unanswered ask is not.
	if got := r.State(at(2 + DoneHold.Seconds() + 1)).Pose; got != companion.Resting {
		t.Errorf("after DoneHold pose = %v, want Resting", got)
	}
}

// A nag can arrive after the knock: the agent finished, the user stayed away.
// The open ask must outrank the stale knock on both channels.
func TestAskAfterDoneOutranksTheKnock(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Done, Text: "all wired"}, at(0))
	r.Apply(event.Event{Kind: event.NeedsInput, Text: "waiting for your input"}, at(1))
	st := r.State(at(1))
	if st.Pose != companion.NeedsYou || !st.BubbleAsk || st.Bubble != "waiting for your input" {
		t.Errorf("pose=%v ask=%v bubble=%q; an open ask must outrank a stale knock",
			st.Pose, st.BubbleAsk, st.Bubble)
	}
}

// A scape attached partway through a session never sees the prompt.
func TestToolTrafficAloneOpensATurn(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1"}, at(0))
	if st := r.State(at(1)); !st.Act.Working {
		t.Error("tool traffic with no prompt must still read as working")
	}
}
