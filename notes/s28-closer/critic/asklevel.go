package main

// asklevel: the level DURING an open ask, not only at the instant it fires.
//
// Every draft measured the beach at the level at the moment needs_input
// arrives (p50 0.59) and designed against that. But the ask stays open for a
// median 130 seconds, heat decays with TauFall = 12 s, and nothing adds heat
// while the agent is blocked -- so the sea withdraws while the crab is still
// standing there. This folds his real recordings and reports the level at
// t+0, +3, +10, +30, +60 s into an ask that was still open at that moment.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

func askLevels() {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, ".config", "xscapes", "run", "*.jsonl"))
	sort.Strings(files)

	deltas := []float64{0, 1, 3, 10, 30, 60, 120}
	samples := make([][]float64, len(deltas))

	for _, f := range files {
		evs := readEvents(f)
		if len(evs) < 20 {
			continue
		}
		// Index of the event that closes an ask opened at i.
		r := reduce.New("x")
		for i, e := range evs {
			now := time.UnixMilli(e.TS)
			r.Apply(e, now)
			if e.Kind != event.NeedsInput {
				continue
			}
			// how long until this ask closes
			closeAt := int64(-1)
			for j := i + 1; j < len(evs); j++ {
				k := evs[j].Kind
				if k == event.Prompt || k == event.ToolStart || k == event.Done {
					closeAt = evs[j].TS
					break
				}
			}
			if closeAt < 0 {
				continue
			}
			open := float64(closeAt-e.TS) / 1000
			// Read the level forward on a COPY of the reducer's clock: State
			// only decays, so calling it at a later instant on the same
			// reducer is exactly what the live loop does when no event lands.
			snap := reduce.New("x")
			for j := 0; j <= i; j++ {
				snap.Apply(evs[j], time.UnixMilli(evs[j].TS))
			}
			for di, d := range deltas {
				if d > open {
					continue
				}
				st := snap.State(now.Add(time.Duration(d * float64(time.Second))))
				samples[di] = append(samples[di], st.Act.Level)
			}
		}
	}
	fmt.Printf("%8s %7s %8s %8s %8s   %s\n", "into ask", "n", "p10", "p50", "p90", "beach rows at p50 (80x24 / 111x25 / 124x22 / 143x27 / 153x51)")
	for di, d := range deltas {
		v := samples[di]
		if len(v) == 0 {
			continue
		}
		sort.Float64s(v)
		p := func(q float64) float64 { return v[int(q*float64(len(v)-1))] }
		med := p(.50)
		fmt.Printf("%7.0fs %7d %8.2f %8.2f %8.2f   ", d, len(v), p(.10), med, p(.90))
		for _, g := range []struct{ w, h int }{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}} {
			fmt.Printf("%3d ", g.h-sandTopAt(g.w, g.h, med))
		}
		fmt.Println()
	}
}
