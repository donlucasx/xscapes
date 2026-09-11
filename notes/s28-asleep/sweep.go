package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// sweep asks what the two timeouts actually buy and cost, over his own log.
//
// They are a single trade and it has never been measured. FlightStale drops a
// tool whose end never arrived; TurnSilence closes a turn nothing ever ended.
// Both exist to stop a lost event leaving the scene working forever, and both
// are what put the companion to sleep while a real tool was still running.
//
//	ASLEEP     Resting between a prompt and its done -- the eye is shut while
//	           the agent has not handed back. This is his report.
//	FALSE WORK Working after the last event of the session, which is the thing
//	           the timeouts are protecting against: a scene that never settles.
func sweep() {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, ".config", "xscapes", "run", "*.jsonl"))
	sort.Strings(files)
	var logs [][]event.Event
	for _, f := range files {
		if evs := read(f); len(evs) >= 20 {
			logs = append(logs, evs)
		}
	}
	fmt.Printf("\n%d sessions\n\n", len(logs))
	fmt.Printf("%-12s %-12s %10s %10s %12s\n", "FlightStale", "TurnSilence", "asleep h", "worst min", "false work h")
	oldF, oldT := reduce.FlightStale, reduce.TurnSilence
	defer func() { reduce.FlightStale, reduce.TurnSilence = oldF, oldT }()
	for _, fs := range []time.Duration{20 * time.Minute, 45 * time.Minute, 90 * time.Minute, 3 * time.Hour} {
		for _, ts := range []time.Duration{5 * time.Minute, 30 * time.Minute} {
			reduce.FlightStale, reduce.TurnSilence = fs, ts
			var slept, mid, worst, false_ time.Duration
			for _, evs := range logs {
				s, m, w := sleptMidWork(evs)
				slept += s
				mid += m
				if w > worst {
					worst = w
				}
				false_ += falseWork(evs)
			}
			fmt.Printf("%-12v %-12v %10.2f %10.1f %12.2f\n", fs, ts, slept.Hours(), worst.Minutes(), false_.Hours())
		}
	}
	fmt.Println("\nasleep = Resting between a prompt and its done (his report)")
	fmt.Println("false work = still Working an hour past the session's last event (what the timeouts protect)")
}

// falseWork is the cost side: how long the companion keeps working after the
// session has actually stopped producing events.
func falseWork(evs []event.Event) time.Duration {
	r := reduce.New("x")
	i := 0
	end := time.UnixMilli(evs[len(evs)-1].TS)
	for i < len(evs) {
		r.Apply(evs[i], time.UnixMilli(evs[i].TS))
		i++
	}
	var out time.Duration
	for now := end; now.Before(end.Add(time.Hour)); now = now.Add(time.Second) {
		if r.State(now).Pose == companion.Working {
			out += time.Second
		}
	}
	return out
}
