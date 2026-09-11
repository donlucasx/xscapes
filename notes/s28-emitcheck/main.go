// Command emitcheck answers his 2026-09-10 question -- "I did it, but didnt
// notice anything, what was supposed to happen?" -- by running the events he
// actually sent through the real reducer and printing the pose that results.
//
// The commands he was given were WRONG and this is where that shows. `emit`
// builds event.Kind straight from the word on the command line, so `emit ask`
// makes Kind("ask"), and the reducer's switch has no case for it: the event is
// delivered, accepted, and does nothing. Worse, `emit` still prints "sent ask
// to session s28test", because the SOCKET write succeeded -- the transport
// worked and the meaning was dropped, which is exactly the shape of failure
// that looks like success.
//
//	go run ./notes/s28-emitcheck
package main

import (
	"fmt"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

func main() {
	fmt.Printf("%-34s %-12s %-10s %s\n", "what was sent", "pose", "balloon", "verdict")
	fmt.Println("---------------------------------------------------------------------------------")

	// Exactly what he ran, in order.
	show("emit ask -text \"Allow Bash...\"", []event.Event{
		{Kind: event.Kind("ask"), Text: "Allow Bash(rm -rf)?"},
	}, 0)
	show("emit done", []event.Event{
		{Kind: event.Kind("ask"), Text: "Allow Bash(rm -rf)?"},
		{Kind: event.Kind("done")},
	}, 0)

	fmt.Println()
	// What he should have run.
	show("emit needs_input -text \"...\"", []event.Event{
		{Kind: event.NeedsInput, Text: "Allow Bash(rm -rf)?"},
	}, 0)
	show("emit done -text \"...\"", []event.Event{
		{Kind: event.NeedsInput, Text: "Allow Bash(rm -rf)?"},
		{Kind: event.Done, Text: "all set"},
	}, 0)
	show("  ...the same, 6 s later", []event.Event{
		{Kind: event.NeedsInput, Text: "Allow Bash(rm -rf)?"},
		{Kind: event.Done, Text: "all set"},
	}, 6*time.Second)

	fmt.Println()
	fmt.Printf("DoneHold is %v -- that is how long the finish pose holds before it goes back.\n", reduce.DoneHold)
	fmt.Println("A scape with no events at all sits in Resting, and Resting draws the eyes as '-'.")
}

func show(label string, evs []event.Event, wait time.Duration) {
	r := reduce.New("s28test")
	now := time.Now()
	for _, e := range evs {
		r.Apply(e, now)
	}
	st := r.State(now.Add(wait))
	balloon := st.Bubble
	if balloon == "" {
		balloon = "(none)"
	}
	verdict := "the scene changes"
	if st.Pose == companion.Resting && st.Bubble == "" {
		verdict = "NOTHING HAPPENS -- unknown kind, dropped"
	}
	fmt.Printf("%-34s %-12s %-10s %s\n", label, poseName(st.Pose), balloon, verdict)
}

func poseName(s companion.State) string {
	switch s {
	case companion.Resting:
		return "Resting"
	case companion.Working:
		return "Working"
	case companion.NeedsYou:
		return "NeedsYou"
	case companion.Worried:
		return "Worried"
	case companion.Done:
		return "Done"
	}
	return "?"
}
