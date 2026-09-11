// Command asleep measures how long the companion is RESTING while the agent is
// working, by folding his real recordings through the REAL reducer and sampling
// the pose the way the render loop does.
//
// ⚠ IT EXISTS BECAUSE MY FIRST ANSWER WAS A PROXY. I counted gaps longer than
// TurnSilence and called the excess "asleep", which assumed a closed turn means
// a sleeping companion. It does not: pose() is
//
//	case r.turnOpn || len(r.flight) > 0: return Working
//
// so an in-flight tool keeps the eyes open for FlightStale (20 min) after the
// turn has closed. The proxy over-counted, and the repo has a memory about
// exactly this -- measure the observable, print it beside the verdict.
//
//	go run ./notes/s28-asleep
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

func main() {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, ".config", "xscapes", "run", "*.jsonl"))
	sort.Strings(files)

	var restingWorking, working, resting, other time.Duration
	sessions, samples := 0, 0
	worst := time.Duration(0)
	var midWork, worstSleep time.Duration
	for _, f := range files {
		evs := read(f)
		if len(evs) < 20 {
			continue
		}
		sessions++
		r := reduce.New("x")
		i := 0
		start := time.UnixMilli(evs[0].TS)
		end := time.UnixMilli(evs[len(evs)-1].TS)
		run := time.Duration(0)
		// One sample a second, which is what the live loop's slowest frame does.
		for now := start; !now.After(end); now = now.Add(time.Second) {
			for i < len(evs) && !time.UnixMilli(evs[i].TS).After(now) {
				r.Apply(evs[i], time.UnixMilli(evs[i].TS))
				i++
			}
			st := r.State(now)
			samples++
			switch st.Pose {
			case companion.Working:
				working += time.Second
				run = 0
			case companion.Resting:
				resting += time.Second
				// "Asleep while working" is the honest question: is the AGENT
				// still in a turn the user has not answered? The only evidence
				// that survives in the log is that more work arrives later --
				// so this is counted below, on the second pass.
				run += time.Second
				if run > worst {
					worst = run
				}
			default:
				other += time.Second
			}
		}
		// Second pass: a Resting stretch that is FOLLOWED by more work in the
		// same session, with no prompt in between, was work the whole time.
		sl, mw, wr := sleptMidWork(evs)
		restingWorking += sl
		midWork += mw
		if wr > worstSleep {
			worstSleep = wr
		}
	}
	fmt.Printf("sessions folded: %d   samples: %d (%.1f hours of wall clock)\n\n",
		sessions, samples, float64(samples)/3600)
	fmt.Printf("  Working  %7.2f h\n", working.Hours())
	fmt.Printf("  Resting  %7.2f h   <- of which %.2f h had work still arriving\n",
		resting.Hours(), restingWorking.Hours())
	fmt.Printf("  other    %7.2f h   (NeedsYou / Done / Worried)\n", other.Hours())
	fmt.Printf("\n  longest single Resting stretch: %.1f min\n", worst.Minutes())
	fmt.Printf("\nBETWEEN A PROMPT AND ITS DONE -- the agent had not handed back:\n")
	fmt.Printf("  that window totals        %7.2f h\n", midWork.Hours())
	fmt.Printf("  ⭐ ASLEEP inside it       %7.2f h  (%.1f%%)\n",
		restingWorking.Hours(), 100*restingWorking.Hours()/midWork.Hours())
	fmt.Printf("  longest single sleep      %7.1f min\n", worstSleep.Minutes())
	sweep()
	fmt.Printf("\nTurnSilence %v, FlightStale %v, SubStale %v\n",
		reduce.TurnSilence, reduce.FlightStale, reduce.SubStale)
}

// sleptMidWork is the honest measure and the second one I had to write.
//
// ⚠ THE FIRST VERSION SAMPLED AT EVENT TIMES, and the whole defect happens
// BETWEEN events: a single tool runs for 81 minutes, nothing fires in between,
// and an instrument that only looks when an event lands never sees the sleep at
// all. It reported 0.58 h and the true figure is larger. Same family of mistake
// as the proxy it was written to replace -- look where the thing happens.
//
// So: sample every second, and count a second as a defect when the pose is
// Resting while the agent has NOT handed back -- between a prompt and the done
// that answers it, which is exactly the window in which a closed eye is a lie.
func sleptMidWork(evs []event.Event) (slept, midWork time.Duration, worst time.Duration) {
	r := reduce.New("x")
	i := 0
	start := time.UnixMilli(evs[0].TS)
	end := time.UnixMilli(evs[len(evs)-1].TS)
	open := false
	run := time.Duration(0)
	for now := start; !now.After(end); now = now.Add(time.Second) {
		for i < len(evs) && !time.UnixMilli(evs[i].TS).After(now) {
			e := evs[i]
			switch e.Kind {
			case event.Prompt:
				open = true
			case event.Done, event.SessionEnd, event.NeedsInput:
				open = false
			}
			r.Apply(e, time.UnixMilli(e.TS))
			i++
		}
		if !open {
			run = 0
			continue
		}
		midWork += time.Second
		if r.State(now).Pose == companion.Resting {
			slept += time.Second
			run += time.Second
			if run > worst {
				worst = run
			}
		} else {
			run = 0
		}
	}
	return slept, midWork, worst
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
		if e, err := event.Decode(s.Bytes()); err == nil && e.TS > 0 {
			out = append(out, e)
		}
	}
	return out
}
