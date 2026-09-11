// notes/s28-pacehold -- WHAT DOES "STILL" HAVE TO MEAN, measured on his own
// event log rather than argued from the comment.
//
//	go run ./notes/s28-pacehold
//
// THE QUESTION. pace.go's stillFor() names three poses that hold their ground
// -- NeedsYou, Done, Worried -- and implements "hold" as `return 0, false`.
// 0 is HOME, the column nearest the frame edge, so the companion does not hold
// anything: it teleports outward by up to paceSpan(w) cells at the frame the
// pose arrives. notes/s28-pace measures that jump. This asks the next question:
// what should it return instead?
//
// There are two candidates and only one number separates them.
//
//	SETTLED   dx = paceAt(Steps, stepDur, span) -- the offset of the last
//	          COMPLETED step, with the mid-stride interpolation dropped so the
//	          legs stop. Stateless, so it needs no change to the reducer and no
//	          new argument at any of drawScene's twenty call sites.
//	LATCHED   dx = whatever the offset was at the instant the hold began, held
//	          until the hold ends. Needs state that pace() does not have.
//
// The two differ ONLY when the step count advances while the pose is held,
// because Steps is the only input SETTLED reads. So the whole choice reduces to
// a count over his real sessions: how many MAIN-THREAD tool events land inside
// a held window, and how far do they carry the companion?
//
// WHY IT IS NOT OBVIOUS EITHER WAY. NeedsYou and Done should freeze Steps by
// construction -- the agent is blocked on the user, so it is running no tools
// -- but Worried persists across an error while the agent keeps working, and
// the comment in stillFor says worried "should look planted". If Worried
// carries dozens of main-thread tool events, SETTLED slides the companion a
// cell per event with its legs held still, which is worse than the defect.
//
// THE DATA is every .jsonl spool in his run directory, folded through a real
// reduce.Reducer at the events' own timestamps -- the same fold `xscapes tune`
// does, at event resolution rather than one second, because a step that lands
// and is answered inside the same second is still a step the eye sees.
//
// pace_shipped.go here is a SYMLINK to ../../pace.go, so paceSpan and paceAt
// are the shipped function bodies compiled from the same bytes on disk, not a
// transcription that can go stale.
package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// hisWidths are the window widths he has actually been measured at, from the
// session banners: 80 and 111 through 153. 40 is the brief's floor.
var hisWidths = []int{40, 80, 111, 124, 143, 153}

// window is one unbroken stretch of a held pose.
type window struct {
	pose     companion.State
	dur      time.Duration
	steps    int // main-thread tool events that landed INSIDE it
	enterDx  float64
	enterAge float64 // StepAge at the frame the hold arrived
	maxDrift float64 // worst |dx - enterDx| under SETTLED, in cells, at w=124
}

func settled(steps int, span int) float64 {
	dx, _ := paceAt(steps, stepDur, span)
	return dx
}

func held(p companion.State) bool {
	return p == companion.NeedsYou || p == companion.Done || p == companion.Worried
}

func main() {
	dir, err := event.RunDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "s28-pacehold:", err)
		os.Exit(1)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	sort.Strings(files)

	span124 := paceSpan(124)

	var wins []window
	sessions, kept, totalEvents := 0, 0, 0
	for _, f := range files {
		sessions++
		evs := load(f)
		if len(evs) < 20 {
			continue
		}
		kept++
		totalEvents += len(evs)

		r := reduce.New("pacehold")
		var cur *window
		prevPose := companion.Resting
		prevSteps := 0
		for _, e := range evs {
			at := time.UnixMilli(e.TS)
			r.Apply(e, at)
			st := r.State(at)

			if held(st.Pose) && !held(prevPose) {
				cur = &window{pose: st.Pose, enterDx: settled(st.Steps, span124), enterAge: st.StepAge}
				cur.dur = 0
				cur.steps = 0
				startAt := at
				// closure-free: remember the start on the struct via dur later
				_ = startAt
				winStart = at
			}
			if cur != nil {
				// ORDER MATTERS, and getting it wrong inflated this number
				// once already. A ToolStart CLEARS needsInput -- reduce.go:251,
				// "a tool starting means a permission prompt, if there was one,
				// was answered" -- so the step that ends an ask is the same
				// event that ends the window. Counting it before closing the
				// window credited the ask with a step that lands after it is
				// over, and turned 0 of 24 into 17 of 24. Close first.
				//
				// A pose can also change identity inside a held run (NeedsYou
				// -> Worried). That is still one unbroken hold: the companion
				// is not allowed to move on that transition either.
				if !held(st.Pose) {
					cur.dur = at.Sub(winStart)
					wins = append(wins, *cur)
					cur = nil
				} else if st.Steps != prevSteps {
					cur.steps += st.Steps - prevSteps
					if d := math.Abs(settled(st.Steps, span124) - cur.enterDx); d > cur.maxDrift {
						cur.maxDrift = d
					}
				}
			}
			prevPose, prevSteps = st.Pose, st.Steps
		}
		if cur != nil {
			cur.dur = time.UnixMilli(evs[len(evs)-1].TS).Sub(winStart)
			wins = append(wins, *cur)
		}
	}

	fmt.Println("==============================================================================")
	fmt.Println("WHAT \"STILL\" HAS TO MEAN -- held windows in his real event log")
	fmt.Println("==============================================================================")
	fmt.Printf("  %d spool files in %s\n", sessions, dir)
	fmt.Printf("  %d folded (>= 20 events), %d events, %d held windows\n\n", kept, totalEvents, len(wins))
	if len(wins) == 0 {
		fmt.Println("  EMPTY PROBE. Nothing below means anything. Stop here.")
		return
	}

	fmt.Println("0. THE PROBE IS NON-EMPTY")
	byPose := map[companion.State][]window{}
	for _, w := range wins {
		byPose[w.pose] = append(byPose[w.pose], w)
	}
	fmt.Printf("   %-12s %8s %10s %10s %10s %10s\n", "entered as", "windows", "median s", "p90 s", "with steps", "max steps")
	for _, p := range []companion.State{companion.NeedsYou, companion.Done, companion.Worried} {
		ws := byPose[p]
		if len(ws) == 0 {
			fmt.Printf("   %-12s %8d\n", p.String(), 0)
			continue
		}
		durs := make([]float64, len(ws))
		withSteps, maxSteps := 0, 0
		for i, w := range ws {
			durs[i] = w.dur.Seconds()
			if w.steps > 0 {
				withSteps++
			}
			if w.steps > maxSteps {
				maxSteps = w.steps
			}
		}
		sort.Float64s(durs)
		fmt.Printf("   %-12s %8d %10.1f %10.1f %10s %10d\n", p.String(), len(ws),
			pct(durs, 0.50), pct(durs, 0.90),
			fmt.Sprintf("%d (%.0f%%)", withSteps, 100*float64(withSteps)/float64(len(ws))), maxSteps)
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("1. DOES SETTLED SLIDE THE COMPANION DURING A HOLD?")
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("   SETTLED reads Steps, so it moves exactly when a main-thread tool")
	fmt.Println("   event lands inside a held window. Drift is in CELLS at w=124,")
	fmt.Println("   span", span124, "-- the width his banners quote most.")
	moved, worst := 0, 0.0
	var drifts []float64
	for _, w := range wins {
		if w.maxDrift > 0 {
			moved++
			drifts = append(drifts, w.maxDrift)
		}
		if w.maxDrift > worst {
			worst = w.maxDrift
		}
	}
	sort.Float64s(drifts)
	fmt.Printf("   %d of %d held windows drift at all (%.1f%%)\n", moved, len(wins), 100*float64(moved)/float64(len(wins)))
	if len(drifts) > 0 {
		fmt.Printf("   of those: median %.1f cells, p90 %.1f, worst %.1f (the strip is %d wide)\n",
			pct(drifts, 0.5), pct(drifts, 0.9), worst, span124)
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("1b. IS A STRIDE IN FLIGHT WHEN THE HOLD ARRIVES?")
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("   SETTLED finishes the step already in flight rather than freezing")
	fmt.Println("   mid-cell, so it can move the companion by at most ONE cell on the")
	fmt.Println("   frame the hold arrives -- and only if StepAge < stepDur =", stepDur, "s.")
	for _, p := range []companion.State{companion.NeedsYou, companion.Done, companion.Worried} {
		ws := byPose[p]
		inFlight := 0
		for _, w := range ws {
			if w.enterAge < stepDur {
				inFlight++
			}
		}
		if len(ws) == 0 {
			continue
		}
		fmt.Printf("   %-22s %3d of %3d entered mid-stride (%.1f%%)\n",
			p.String(), inFlight, len(ws), 100*float64(inFlight)/float64(len(ws)))
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("2. WHAT EACH RULE COSTS AT THE MOMENT OF THE ASK, EVERY WIDTH")
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("   The jump at the transition is the defect. It depends only on where")
	fmt.Println("   the companion stood, so it is exact arithmetic over every reachable")
	fmt.Println("   step count, not a sample.")
	fmt.Printf("   %-6s %-6s %-22s %-22s\n", "width", "span", "TODAY (return 0)", "SETTLED / LATCHED")
	for _, w := range hisWidths {
		sp := paceSpan(w)
		var today, fixed float64
		n := 2 * sp
		for k := 0; k <= n; k++ {
			before := settled(k, sp)
			today += math.Abs(0 - before)
			fixed += math.Abs(settled(k, sp) - before)
		}
		fmt.Printf("   %-6d %-6d mean %.2f cells, max %-4d mean %.2f cells, max %d\n",
			w, sp, today/float64(n+1), sp, fixed/float64(n+1), 0)
	}
}

// winStart is the start of the hold currently being accumulated. A package
// variable rather than a field because the window struct is copied into the
// slice and the start is only ever needed while one is open.
var winStart time.Time

func pct(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	i := int(p * float64(len(sorted)-1))
	return sorted[i]
}

func load(path string) []event.Event {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var evs []event.Event
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		e, err := event.Decode(line)
		if err != nil || e.TS == 0 {
			continue
		}
		evs = append(evs, e)
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].TS < evs[j].TS })
	return evs
}
