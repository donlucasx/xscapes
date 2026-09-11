package reduce

import (
	"fmt"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
)

// HIS REPORT, 2026-09-11: "when the main agent is working, alone, it should
// have its eyes open." He had screenshotted a workflow at "13/14 agents done,
// 1h 0m 37s" with the companion asleep on the sand.
//
// The mechanism is TurnSilence. It force-closes a turn no `done` ever ended,
// and it measures quiet by MAIN-and-subagent event traffic alone -- so a turn
// whose work has moved into subagents goes quiet on paper while the machine is
// flat out. pose() then has no turn and no flight, returns Resting, and the
// crab draws a closed eye.
//
// Counted over his own run log, 124 files and 76,496 events folded through this
// reducer second by second: 3.53 hours of Resting inside stretches the log
// itself proves were work, against 85.28 hours of work in total.

// sub is one subagent's start or end.
func sub(kind event.Kind, id string) event.Event {
	return event.Event{Kind: kind, Agent: id, AgentType: "general-purpose"}
}

// TestTheCompanionStaysAwakeWhileItsKittensAreOnTheBeach is the guarantee the
// whole fix exists for, and it is also the one that makes the frame coherent:
// the litter and the companion read the same event. Drawing kittens on the sand
// beside a companion with its eyes shut says two contradictory things about the
// same instant.
func TestTheCompanionStaysAwakeWhileItsKittensAreOnTheBeach(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	for i := 0; i < 3; i++ {
		r.Apply(sub(event.SubStart, fmt.Sprintf("a%d", i)), at(float64(i+1)))
	}

	// Then nothing at all. The main thread is waiting on the fan-out, so it
	// fires no tool events of its own; this is the silence the timeout sees.
	for _, sec := range []float64{
		TurnSilence.Seconds() + 60,
		TurnSilence.Seconds() + 300,
		SubStale.Seconds() - 60,
	} {
		st := r.State(at(sec))
		if st.Kittens == 0 {
			t.Fatalf("at %.0fs the litter emptied on its own; this test is no longer testing anything", sec)
		}
		if st.Pose == companion.Resting {
			t.Errorf("at %.0fs the companion is asleep with %d kittens on the beach", sec, st.Kittens)
		}
	}
}

// TestHisFanOutKeepsTheEyesOpen is his screenshot, replayed: fourteen
// subagents, thirteen of them finishing one at a time across an hour, one still
// running at the end. Nothing here touches the main thread after the fan-out
// starts, which is exactly why the hour used to read as silence.
//
// ⚠ THE LIMIT THIS TEST DELIBERATELY STOPS AT, so nobody reads it as more than
// it is: the fix is bounded by SubStale. A subagent still running thirty
// minutes after its sub_start has its kitten swept, and the eyes go with it.
// Counted over his log, 26 of 639 finished subagents (4.1%) run longer than
// that and the longest ran 57.3 min, so the case is real, not exotic. Raising
// SubStale to an hour covers every subagent he has ever run and measures at
// 0.23 h -> 0.17 h of false sleep for no false work at all -- but it moves the
// LITTER, which is a different channel and his to rule on. It is not done here.
func TestHisFanOutKeepsTheEyesOpen(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	for i := 0; i < 14; i++ {
		r.Apply(sub(event.SubStart, fmt.Sprintf("a%02d", i)), at(float64(i)*0.5))
	}
	// Thirteen finish, one every four and a bit minutes, out to 54 minutes.
	// The fourteenth never does.
	ends := make([]float64, 13)
	for i := range ends {
		ends[i] = 120 + float64(i)*260
	}
	lastEnd := ends[len(ends)-1]

	var asleep []float64
	next := 0
	for sec := 1.0; sec <= lastEnd+TurnSilence.Seconds(); sec++ {
		for next < len(ends) && ends[next] <= sec {
			r.Apply(sub(event.SubEnd, fmt.Sprintf("a%02d", next)), at(ends[next]))
			next++
		}
		if r.State(at(sec)).Pose == companion.Resting {
			asleep = append(asleep, sec)
		}
	}
	if next != len(ends) {
		t.Fatalf("only %d of %d subagent ends were applied; the loop is not covering the hour", next, len(ends))
	}
	if len(asleep) > 0 {
		t.Errorf("the companion slept for %d s of his hour-long fan-out, first at %.0fs",
			len(asleep), asleep[0])
	}
}

// TestTheTurnStillClosesWhenNothingIsRunning is the other half of the trade.
// TurnSilence exists because `done` is the ONLY event that ends a turn, so a
// killed pane or a crashed agent would otherwise leave the companion working
// forever. Keeping a turn alive for its subagents must not weaken that.
func TestTheTurnStillClosesWhenNothingIsRunning(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(event.Event{Kind: event.ToolStart, ID: "t1"}, at(1))
	r.Apply(event.Event{Kind: event.ToolEnd, ID: "t1"}, at(2))

	late := at(TurnSilence.Seconds() + 60)
	st := r.State(late)
	if st.Act.Working {
		t.Error("a turn with no tools and no subagents must still time out")
	}
	if st.Pose != companion.Resting {
		t.Errorf("pose = %v, want Resting once a dead turn has timed out", st.Pose)
	}
}

// TestAStrandedSubagentDoesNotHoldTheCompanionAwakeForever bounds the fix.
// A subagent whose sub_end was lost would otherwise be a new way to pin the
// scene at "working" for the rest of the day -- the exact failure TurnSilence
// was written to prevent, re-entering by the back door. SubStale is what closes
// it: the kitten goes, and the companion goes with it.
//
// Measured over his log, this bound is generous enough to cost nothing real:
// subagent lifetimes run p50 7.4 min, p90 20.8 min, max 57.3 min, so SubStale
// at 30 min covers 96% of them.
func TestAStrandedSubagentDoesNotHoldTheCompanionAwakeForever(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(sub(event.SubStart, "ghost"), at(1))

	// Still awake while the kitten is on the beach.
	if got := r.State(at(SubStale.Seconds() - 60)).Pose; got == companion.Resting {
		t.Errorf("pose = %v at %.0fs, want the companion awake while the kitten is there", got, SubStale.Seconds()-60)
	}
	// Gone once the kitten is swept.
	late := at(SubStale.Seconds() + TurnSilence.Seconds() + 60)
	st := r.State(late)
	if st.Kittens != 0 {
		t.Fatalf("kittens = %d, expected the stale sweep to have taken it", st.Kittens)
	}
	if st.Pose != companion.Resting {
		t.Errorf("pose = %v, want Resting: a lost sub_end must not pin the companion forever", st.Pose)
	}
	if st.Act.Working {
		t.Error("a lost sub_end pinned the sea at working forever")
	}

	// And the same bound stated in WALL TIME rather than in terms of the
	// constants. An assertion written as SubStale+TurnSilence passes no matter
	// how large those get, which is exactly the kind of test that certifies a
	// runaway. Ninety minutes is the outside edge of anything defensible here,
	// and it leaves room for SubStale to go to an hour if he rules that way.
	//
	// It needs its OWN reducer: the clock here only goes forward (decay returns
	// early on a negative dt), so asking the one above about minute ninety
	// after it has already been advanced to minute thirty-five would read a
	// frozen state and pass on anything.
	const hardBound = 90 * 60
	r2 := New("s")
	r2.Apply(event.Event{Kind: event.Prompt}, at(0))
	r2.Apply(sub(event.SubStart, "ghost"), at(1))
	st2 := r2.State(at(hardBound))
	if st2.Pose != companion.Resting {
		t.Errorf("pose = %v ninety minutes after the last event; a stranded subagent must not outlast that", st2.Pose)
	}
	if st2.Act.Working {
		t.Error("still working ninety minutes after the last event")
	}
}

// TestASleepingCompanionNeverHasKittens states the invariant as the renderer
// sees it: over a long fan-out with a stranded subagent, there must be no
// instant where the scene draws a litter and a closed eye together.
func TestASleepingCompanionNeverHasKittens(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(sub(event.SubStart, "a"), at(1))
	r.Apply(sub(event.SubStart, "b"), at(2))
	r.Apply(sub(event.SubEnd, "b"), at(900))

	var bad int
	var sawKittens, sawRest bool
	for sec := 1.0; sec <= SubStale.Seconds()+TurnSilence.Seconds()+600; sec += 5 {
		st := r.State(at(sec))
		if st.Kittens > 0 {
			sawKittens = true
		}
		if st.Pose == companion.Resting {
			sawRest = true
		}
		if st.Kittens > 0 && st.Pose == companion.Resting {
			bad++
		}
	}
	if !sawKittens || !sawRest {
		t.Fatalf("sample never reached both states (kittens %v, rest %v); it proves nothing", sawKittens, sawRest)
	}
	if bad > 0 {
		t.Errorf("%d sampled instants drew kittens beside a sleeping companion", bad)
	}
}

// TestLiveSubagentsHoldTheSeaAsWellAsTheEyes pins the answer to "does this move
// the water too". It does, and deliberately: the encoding rule says the water
// is the WORK, and a running subagent is work. Holding the eyes open while
// letting the sea go flat would have the two channels disagree about the same
// fact.
func TestLiveSubagentsHoldTheSeaAsWellAsTheEyes(t *testing.T) {
	r := New("s")
	r.Apply(event.Event{Kind: event.Prompt}, at(0))
	r.Apply(sub(event.SubStart, "a"), at(1))

	// Long past TurnSilence, long past any heat: only the floor is left.
	st := r.State(at(TurnSilence.Seconds() + 600))
	if st.Kittens == 0 {
		t.Fatalf("the subagent was swept before the sample point; this test is measuring nothing")
	}
	if !st.Act.Working {
		t.Error("Act.Working went false while a subagent was running")
	}
	// Stated without reference to the constant, so setting TurnFloor to zero
	// cannot make this pass by moving the goalposts.
	if st.Act.Level <= 0 {
		t.Errorf("the sea went dead flat at %.3f while a subagent was running", st.Act.Level)
	}
	if st.Act.Level < TurnFloor-1e-9 {
		t.Errorf("sea fell to %.3f while a subagent ran, below the %.2f turn floor", st.Act.Level, TurnFloor)
	}
	// And it is only the FLOOR, not a swell: nothing is being reported, so the
	// water must be as low as an open turn is allowed to go.
	if st.Act.Level > TurnFloor+1e-6 {
		t.Errorf("sea at %.3f, above the floor, on no events at all", st.Act.Level)
	}
}

// SubStale MUST stay comfortably longer than TurnSilence, or the exemption that
// keeps the companion awake during a fan-out is dead code.
//
// It was, briefly: both were 30 minutes on 2026-09-11, so the litter was swept
// at the same moment the turn would have closed and the rule could never fire.
// Two tests caught it by refusing to measure nothing rather than by passing
// vacuously, which is the only reason it was noticed at all.
func TestTheLitterOutlivesTheTurnsSilence(t *testing.T) {
	if SubStale <= TurnSilence {
		t.Fatalf("SubStale %v is not longer than TurnSilence %v -- a turn with live "+
			"subagents can never outlast the silence, so the exemption is dead code",
			SubStale, TurnSilence)
	}
	// And by enough to be useful: a fan-out has to survive the bar with room to
	// spare, not by a second.
	if margin := SubStale - TurnSilence; margin < 15*time.Minute {
		t.Errorf("only %v between TurnSilence and SubStale; a fan-out that quiet "+
			"is swept almost as soon as it is exempt", margin)
	}
}
