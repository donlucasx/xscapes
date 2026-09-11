package main

// askdur folds his real recordings and answers the question all four drafts
// assumed away: how LONG is an ask open? If it is open for many seconds, an
// approach animation is not a delay on the notification -- the balloon rings at
// t=0 and the crab walks up underneath it while he is still reading.
//
// needsInput is set by NeedsInput and cleared by Prompt, ToolStart or Done
// (internal/reduce/reduce.go). So the open window is NeedsInput -> next of those.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/donlucasx/xscapes/internal/event"
)

func askDurations() {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, ".config", "xscapes", "run", "*.jsonl"))
	sort.Strings(files)

	var open []float64
	asks, sessions, unclosed := 0, 0, 0
	var perSession []int
	for _, f := range files {
		evs := readEvents(f)
		if len(evs) < 20 {
			continue
		}
		sessions++
		n := 0
		pending := int64(-1)
		for _, e := range evs {
			switch e.Kind {
			case event.NeedsInput:
				if pending < 0 {
					pending = e.TS
				}
				asks++
				n++
			case event.Prompt, event.ToolStart, event.Done:
				if pending >= 0 {
					open = append(open, float64(e.TS-pending)/1000)
					pending = -1
				}
			}
		}
		if pending >= 0 {
			unclosed++
		}
		perSession = append(perSession, n)
	}
	sort.Float64s(open)
	fmt.Printf("HIS REAL LOG: %d sessions, %d needs_input events, %d closed windows, %d left open at session end\n",
		sessions, asks, len(open), unclosed)
	if len(open) == 0 {
		return
	}
	p := func(q float64) float64 { return open[int(q*float64(len(open)-1))] }
	fmt.Printf("  how long an ask stays open (s):  p10 %.1f  p25 %.1f  p50 %.1f  p75 %.1f  p90 %.1f  max %.0f\n",
		p(.10), p(.25), p(.50), p(.75), p(.90), open[len(open)-1])
	over := func(s float64) int {
		n := 0
		for _, v := range open {
			if v >= s {
				n++
			}
		}
		return n
	}
	fmt.Printf("  windows at least 1s: %d of %d (%.0f%%)   at least 2s: %d (%.0f%%)   at least 5s: %d (%.0f%%)\n",
		over(1), len(open), 100*float64(over(1))/float64(len(open)),
		over(2), 100*float64(over(2))/float64(len(open)),
		over(5), 100*float64(over(5))/float64(len(open)))
	sort.Ints(perSession)
	nz := 0
	for _, v := range perSession {
		if v > 0 {
			nz++
		}
	}
	fmt.Printf("  asks per session: %d of %d sessions have any at all; median of those that do: %d\n",
		nz, len(perSession), medianNonZero(perSession))
}

func medianNonZero(v []int) int {
	var nz []int
	for _, x := range v {
		if x > 0 {
			nz = append(nz, x)
		}
	}
	if len(nz) == 0 {
		return 0
	}
	return nz[len(nz)/2]
}

func readEvents(path string) []event.Event {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []event.Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var e event.Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			out = append(out, e)
		}
	}
	return out
}
