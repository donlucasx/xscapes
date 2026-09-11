// Command askroom answers the question HIS idea turns on, 2026-09-10: "can we
// animate the main Agent/companion to come CLOSER to the user before it
// prompts it?"
//
// A bigger companion needs beach to stand on, and the beach is not a fixed
// size any more -- since the tide shipped, the water comes IN as the agent
// works and withdraws as it goes quiet. So the room available for a bigger
// crab is not a constant; it is a function of the activity level AT THE MOMENT
// THE ASK FIRES.
//
// This folds his real recordings through the real reducer and reports the level
// at every needs_input, and what that level leaves as beach.
//
//	go run ./notes/s28-askroom
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

func main() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "xscapes", "run")
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	sort.Strings(files)

	var levels []float64
	asks, sessions := 0, 0
	for _, f := range files {
		evs := read(f)
		if len(evs) < 20 {
			continue
		}
		sessions++
		r := reduce.New("x")
		for _, e := range evs {
			now := time.UnixMilli(e.TS)
			r.Apply(e, now)
			if e.Kind == event.NeedsInput {
				asks++
				levels = append(levels, r.State(now).Act.Level)
			}
		}
	}
	if asks == 0 {
		fmt.Println("no needs_input events found -- nothing to report, and a clean")
		fmt.Println("result from an empty sample looks exactly like a pass.")
		return
	}
	sort.Float64s(levels)
	p := func(q float64) float64 { return levels[int(q*float64(len(levels)-1))] }
	fmt.Printf("sessions folded: %d   needs_input events: %d\n\n", sessions, asks)
	fmt.Printf("activity level when the ask fires:  p10 %.2f  p25 %.2f  p50 %.2f  p75 %.2f  p90 %.2f  max %.2f\n\n",
		p(0.10), p(0.25), p(0.50), p(0.75), p(0.90), p(1.0))

	fmt.Println("what that leaves to stand on (rows from the waterline to the bottom of the frame):")
	fmt.Printf("%-9s %8s %8s %8s %8s %8s\n", "geom", "p10", "p50", "p90", "busy=1", "today")
	for _, g := range []struct{ w, h int }{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}} {
		fmt.Printf("%-9s %8d %8d %8d %8d %8d\n",
			fmt.Sprintf("%dx%d", g.w, g.h),
			beach(g.w, g.h, p(0.10)), beach(g.w, g.h, p(0.50)),
			beach(g.w, g.h, p(0.90)), beach(g.w, g.h, 1.0), 7)
	}
	fmt.Println("\n'today' is the companion's own height. A number below it means the")
	fmt.Println("companion is already standing in the sea at that level.")
}

func beach(w, h int, level float64) int {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	return (h - 1) - sh.SandTop() + 1
}

func read(path string) []event.Event {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []event.Event
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1<<20), 1<<20)
	for s.Scan() {
		if e, err := event.Decode(s.Bytes()); err == nil {
			out = append(out, e)
		}
	}
	return out
}
