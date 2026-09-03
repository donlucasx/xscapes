package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// runTune replays real recordings through the reducer OFFLINE and reports what
// the sea actually did, so the constants can be set against a day of work
// instead of against a guess.
//
// `replay` is the wrong tool for this and always was: it plays a spool into a
// live scape at wall-clock speed, so checking one setting means sitting and
// watching an afternoon go by, and comparing two means doing it twice from
// memory. This runs the same events through the same reducer with no renderer
// attached, samples the level every second, and prints the distribution.
//
// The questions it is built to answer, in the order they matter:
//
//   - Is the sea SATURATED? A level pinned near 1 says nothing. The swells
//     carry how hard the agent is working; if they are always at maximum the
//     channel is dead and the scape is decoration.
//   - Is it DEAD while work is happening? The opposite failure, and the one a
//     glance would read as "the thing is broken".
//   - How long after the work stops does it take to settle? That is TauFall's
//     real value rather than its nominal one, and it is what decides whether
//     "it went quiet" is visible at a glance.
func runTune(args []string) {
	fs := flag.NewFlagSet("tune", flag.ExitOnError)
	dir := fs.String("dir", "", "directory of .jsonl spools (default: the live run dir)")
	sweep := fs.Bool("sweep", false, "sweep TauFall and Impulse and print a table")
	minEvents := fs.Int("min", 20, "ignore sessions with fewer events than this")
	fs.Parse(args)

	files, err := spoolFiles(*dir, fs.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "xscapes:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "xscapes tune: no spool files found")
		os.Exit(1)
	}

	if *sweep {
		sweepConstants(files, *minEvents)
		return
	}
	report(files, *minEvents)
}

func spoolFiles(dir string, args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	if dir == "" {
		d, err := event.RunDir()
		if err != nil {
			return nil, err
		}
		dir = d
	}
	return filepath.Glob(filepath.Join(dir, "*.jsonl"))
}

// maxGapLabel names the population every number in the report belongs to.
//
// A gap longer than this is the machine asleep or the user at lunch, not a
// quiet sea. Sampling through it would drown every real second in idle ones and
// report that the scape is calm, which would be true of the day and false of
// the thing being measured.
const maxGapLabel = "3m"

// run is one session folded through the reducer, sampled once a second.
type run struct {
	name    string
	events  int
	samples []lvlSample
	span    time.Duration
	settle  []time.Duration // from the last tool event of a turn to level < 0.1
}

type lvlSample struct {
	lvl     float64
	working bool
}

// fold replays one file. Sampling is at one second, which is the rate a glance
// happens at -- sampling per frame would weight a busy second the same as an
// idle one and say the sea is calm because most seconds are.
func fold(path string, minEvents int) (run, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return run{}, false
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
	if len(evs) < minEvents {
		return run{}, false
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].TS < evs[j].TS })

	r := reduce.New("tune")
	out := run{name: filepath.Base(path), events: len(evs)}
	at := func(ms int64) time.Time { return time.UnixMilli(ms) }
	start, end := at(evs[0].TS), at(evs[len(evs)-1].TS)
	out.span = end.Sub(start)

	const maxGap = 3 * time.Minute

	i := 0
	var lastBusy time.Time
	settling := false
	for t := start; !t.After(end); t = t.Add(time.Second) {
		for i < len(evs) && !at(evs[i].TS).After(t) {
			e := evs[i]
			r.Apply(e, at(e.TS))
			if e.Kind == event.ToolStart || e.Kind == event.ToolEnd {
				lastBusy, settling = at(e.TS), true
			}
			i++
		}
		st := r.State(t)
		out.samples = append(out.samples, lvlSample{st.Act.Level, st.Act.Working})
		if settling && st.Act.Level < 0.1 {
			out.settle = append(out.settle, t.Sub(lastBusy))
			settling = false
		}
		// Skip a long idle stretch rather than sampling every second of it.
		if i < len(evs) {
			if gap := at(evs[i].TS).Sub(t); gap > maxGap {
				t = at(evs[i].TS).Add(-time.Second)
			}
		}
	}
	return out, len(out.samples) > 0
}

func report(files []string, minEvents int) {
	var all []lvlSample
	var settle []time.Duration
	runs, evs := 0, 0
	var span time.Duration
	for _, f := range files {
		r, ok := fold(f, minEvents)
		if !ok {
			continue
		}
		runs++
		evs += r.events
		span += r.span
		all = append(all, r.samples...)
		settle = append(settle, r.settle...)
	}
	if len(all) == 0 {
		fmt.Println("no sessions large enough to say anything about")
		return
	}

	fmt.Printf("THE SEA, FOLDED FROM REAL RECORDINGS\n")
	fmt.Printf("  %d sessions, %d events, %s of session time, %d samples at 1s\n\n",
		runs, evs, span.Round(time.Minute), len(all))

	lvls := make([]float64, 0, len(all))
	var busy []float64
	for _, s := range all {
		lvls = append(lvls, s.lvl)
		if s.working {
			busy = append(busy, s.lvl)
		}
	}
	fmt.Println("  level, every second of every session:")
	printDist(lvls)
	fmt.Println("\n  level while a turn is open or a tool is running:")
	printDist(busy)

	sat, dead := 0, 0
	for _, v := range busy {
		if v > 0.95 {
			sat++
		}
		if v < 0.05 {
			dead++
		}
	}
	if len(busy) > 0 {
		fmt.Printf("\n  SATURATED (>0.95) while working: %5.1f%%   -- the swells say nothing here\n",
			100*float64(sat)/float64(len(busy)))
		fmt.Printf("  DEAD      (<0.05) while working: %5.1f%%   -- reads as broken\n",
			100*float64(dead)/float64(len(busy)))
	}
	// Name the population beside the number. "Session time" here EXCLUDES gaps
	// longer than three minutes, which are skipped rather than sampled -- so
	// this is the share of ACTIVE time, not of the day, and reading it as the
	// latter would make the scape look far busier than it is.
	fmt.Printf("  working at all:                  %5.1f%% of active time (gaps over %s skipped)\n",
		100*float64(len(busy))/float64(len(all)), maxGapLabel)

	if len(settle) > 0 {
		sort.Slice(settle, func(i, j int) bool { return settle[i] < settle[j] })
		fmt.Printf("\n  time from the last tool event to level < 0.1, %d occurrences:\n", len(settle))
		fmt.Printf("    median %s   90th %s   slowest %s   (TauFall is %s)\n",
			settle[len(settle)/2].Round(time.Second),
			settle[len(settle)*9/10].Round(time.Second),
			settle[len(settle)-1].Round(time.Second),
			reduce.TauFall)
	}
}

func printDist(v []float64) {
	if len(v) == 0 {
		fmt.Println("    (nothing)")
		return
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	q := func(f float64) float64 {
		i := int(f * float64(len(s)-1))
		return s[i]
	}
	fmt.Printf("    min %.2f  p10 %.2f  p25 %.2f  median %.2f  p75 %.2f  p90 %.2f  max %.2f\n",
		s[0], q(0.10), q(0.25), q(0.50), q(0.75), q(0.90), s[len(s)-1])
	// A histogram, because a distribution that is bimodal and one that is flat
	// have the same quartiles and are completely different seas.
	const bins = 10
	counts := make([]int, bins)
	for _, x := range s {
		b := int(x * bins)
		if b >= bins {
			b = bins - 1
		}
		counts[b]++
	}
	for b := 0; b < bins; b++ {
		bar := counts[b] * 40 / len(s)
		fmt.Printf("    %.1f-%.1f |%-40s| %5.1f%%\n", float64(b)/bins, float64(b+1)/bins,
			barOf(bar), 100*float64(counts[b])/float64(len(s)))
	}
}

func barOf(n int) string {
	b := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		b = append(b, '#')
	}
	return string(b)
}

// sweepConstants re-folds the whole recording for each candidate setting.
//
// The constants are package-level, so this mutates them and restores them --
// acceptable in a one-shot command, and it is the only way to ask "what would
// the sea have looked like" without a second implementation of the reducer to
// disagree with the first.
func sweepConstants(files []string, minEvents int) {
	tau, imp := reduce.TauFall, reduce.Impulse
	defer func() { reduce.TauFall, reduce.Impulse = tau, imp }()

	fmt.Println("SWEEP -- each row re-folds every recording with that setting")
	fmt.Printf("%-10s %-9s | %-8s %-8s %-8s | %-9s %-9s\n",
		"TauFall", "Impulse", "median", "p90", "max", "saturated", "dead")
	for _, tv := range []time.Duration{6 * time.Second, 12 * time.Second, 20 * time.Second, 40 * time.Second} {
		for _, iv := range []float64{0.15, 0.30, 0.50} {
			reduce.TauFall, reduce.Impulse = tv, iv
			var busy []float64
			for _, f := range files {
				r, ok := fold(f, minEvents)
				if !ok {
					continue
				}
				for _, s := range r.samples {
					if s.working {
						busy = append(busy, s.lvl)
					}
				}
			}
			if len(busy) == 0 {
				continue
			}
			sort.Float64s(busy)
			sat, dead := 0, 0
			for _, v := range busy {
				if v > 0.95 {
					sat++
				}
				if v < 0.05 {
					dead++
				}
			}
			fmt.Printf("%-10s %-9.2f | %-8.2f %-8.2f %-8.2f | %8.1f%% %8.1f%%\n",
				tv, iv, busy[len(busy)/2], busy[len(busy)*9/10], busy[len(busy)-1],
				100*float64(sat)/float64(len(busy)), 100*float64(dead)/float64(len(busy)))
		}
	}
}
