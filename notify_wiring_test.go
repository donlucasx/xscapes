package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/notify"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// The two halves of the nudge have to fit: the reducer decides what the bubble
// says and which knock it is, the knocker decides when that earns a sound.
// Testing them apart proves neither -- a BubbleAsk the reducer never set would
// pass both suites and ring the wrong sound in the one moment that matters.
//
// This runs a believable turn at the frame rate the live loop uses and records
// every sound it would have made.
func TestATurnRingsAskThenDone(t *testing.T) {
	base := time.Now()
	at := func(sec float64) time.Time {
		return base.Add(time.Duration(sec * float64(time.Second)))
	}

	red := reduce.New("s")
	var k notify.Knocker
	var rung []notify.Kind

	// Frames tick at 20fps, the live default. Events land between frames.
	feed := []struct {
		at float64
		ev event.Event
	}{
		{0, event.Event{Kind: event.Prompt, Text: "add rate limiting"}},
		{1, event.Event{Kind: event.ToolStart, ID: "t1", Op: event.OpRead}},
		{2, event.Event{Kind: event.ToolEnd, ID: "t1", Op: event.OpRead}},
		// The agent asks for permission, and Claude Code nags every 60s.
		{4, event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}},
		{64, event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}},
		{124, event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}},
		// He answers; work resumes; the turn finishes.
		{130, event.Event{Kind: event.ToolStart, ID: "t2", Op: event.OpShell}},
		{131, event.Event{Kind: event.ToolEnd, ID: "t2", Op: event.OpShell}},
		{132, event.Event{Kind: event.Done, Text: "Rate limiting is in."}},
	}

	next := 0
	for f := 0.0; f < 140; f += 0.05 {
		for next < len(feed) && feed[next].at <= f {
			red.Apply(feed[next].ev, at(f))
			next++
		}
		st := red.State(at(f))
		if kind, ring := k.Knock(st.Bubble, st.BubbleAsk, st.Pose == companion.Worried); ring {
			rung = append(rung, kind)
		}
	}

	want := []notify.Kind{notify.Ask, notify.Done}
	if len(rung) != len(want) {
		t.Fatalf("rang %v over one turn, want exactly %v -- two nags and 2800 frames must add nothing", rung, want)
	}
	for i := range want {
		if rung[i] != want[i] {
			t.Errorf("knock %d was %v, want %v", i, rung[i], want[i])
		}
	}
}

// A permission prompt arriving while the build is broken is the case that has
// already been fixed once for the bubble: the companion goes Worried, which
// outranks NeedsYou in the pose, so anything keyed off the POSE goes silent on
// the one event the agent is actually blocked on.
func TestABlockedAgentStillRingsWhileWorried(t *testing.T) {
	base := time.Now()
	at := func(sec float64) time.Time {
		return base.Add(time.Duration(sec * float64(time.Second)))
	}

	red := reduce.New("s")
	var k notify.Knocker
	red.Apply(event.Event{Kind: event.Prompt}, at(0))
	k.Knock(red.State(at(0)).Bubble, red.State(at(0)).BubbleAsk, false)

	red.Apply(event.Event{Kind: event.Error, Tool: "Bash", Detail: "exit 1"}, at(1))
	st := red.State(at(1))
	if _, ring := k.Knock(st.Bubble, st.BubbleAsk, st.Pose == companion.Worried); ring {
		t.Error("a failing command is not a knock: nothing is waiting on the user yet")
	}

	red.Apply(event.Event{Kind: event.NeedsInput, Text: "allow Bash?"}, at(2))
	st = red.State(at(2))
	kind, ring := k.Knock(st.Bubble, st.BubbleAsk, st.Pose == companion.Worried)
	if !ring || kind != notify.Ask {
		t.Errorf("blocked while worried gave (%v, %v), want (ask, true)", kind, ring)
	}
}

// A turn that ends with a main-thread failure still standing rings the
// worried done: his ruling of 2026-09-15, and 58 of the 381 finishes in his
// own log (15%) end that way. A failure inside a subagent does not, because
// the companion does not worry for the litter (his ruling of 2026-09-07) and
// the knock follows the pose's worry, not the sea's.
func TestATurnThatEndsBrokenRingsTheWorriedDone(t *testing.T) {
	run := func(errAgent string) []notify.Kind {
		base := time.Now()
		at := func(sec float64) time.Time {
			return base.Add(time.Duration(sec * float64(time.Second)))
		}
		red := reduce.New("s")
		var k notify.Knocker
		var rung []notify.Kind
		feed := []struct {
			at float64
			ev event.Event
		}{
			{0, event.Event{Kind: event.Prompt, Text: "run the suite"}},
			{1, event.Event{Kind: event.ToolStart, ID: "t1", Op: event.OpShell, Agent: errAgent}},
			{2, event.Event{Kind: event.Error, ID: "t1", Op: event.OpShell, Tool: "Bash", Detail: "Exit code 1", Agent: errAgent}},
			{3, event.Event{Kind: event.Done, Text: "Two tests fail; see above."}},
		}
		next := 0
		for f := 0.0; f < 6; f += 0.05 {
			for next < len(feed) && feed[next].at <= f {
				red.Apply(feed[next].ev, at(f))
				next++
			}
			st := red.State(at(f))
			if kind, ring := k.Knock(st.Bubble, st.BubbleAsk, st.Pose == companion.Worried); ring {
				rung = append(rung, kind)
			}
		}
		return rung
	}
	if got := run(""); len(got) != 1 || got[0] != notify.Worried {
		t.Errorf("a main-thread failure then a finish rang %v, want exactly [worried]", got)
	}
	if got := run("sub-1"); len(got) != 1 || got[0] != notify.Done {
		t.Errorf("a subagent failure then a finish rang %v, want exactly [done]", got)
	}
}
