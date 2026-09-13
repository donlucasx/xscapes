package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// starfallCadence answers the one question that can reject the feature outright:
// how often would he actually see this?
//
// It is derived from HIS OWN recordings, folded through the REAL reducer at the
// events' own timestamps, rather than quoted from a note. The number in
// reduce.go's comment ("p50 5.4 min") predates the session 27 rebind and is
// stale; re-deriving it is cheaper than arguing about it.
//
// An arrival is a frame where the reducer's TodoDone goes UP. That is the
// definition the product uses, so it automatically includes the closed turns
// and the finished subagents that actually light stars, and automatically
// excludes the todo events that have never once fired.
type cadence struct {
	Spools, Events int
	Arrivals       int
	Gaps           []float64 // seconds between consecutive arrivals, within a spool
	Worst10s       int       // most arrivals inside any 10 second window
	Worst60s       int
	Worst1h        int
	Under          map[float64]int // gaps shorter than a given number of seconds
}

func starfallCadenceOf() (cadence, error) {
	c := cadence{Under: map[float64]int{}}
	dir := filepath.Join(os.Getenv("HOME"), ".config", "xscapes", "run")
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil || len(files) == 0 {
		return c, fmt.Errorf("no spools under %s", dir)
	}
	var all []time.Time
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		c.Spools++
		red := reduce.New("cadence")
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<22)
		last := -1
		var prev time.Time
		for sc.Scan() {
			b := sc.Bytes()
			if len(b) == 0 {
				continue
			}
			var e event.Event
			if json.Unmarshal(b, &e) != nil {
				continue
			}
			c.Events++
			at := time.UnixMilli(e.TS)
			if e.TS == 0 {
				continue
			}
			red.Apply(e, at)
			done := red.State(at).Act.TodoDone
			if last >= 0 && done > last {
				// One arrival per step, because a jump of two is two stars.
				for k := 0; k < done-last; k++ {
					c.Arrivals++
					all = append(all, at)
					if !prev.IsZero() {
						c.Gaps = append(c.Gaps, at.Sub(prev).Seconds())
					}
					prev = at
				}
			}
			if done < last {
				prev = time.Time{} // a new turn's constellation; not a gap
			}
			last = done
		}
		fh.Close()
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Before(all[j]) })
	c.Worst10s = worstWindow(all, 10*time.Second)
	c.Worst60s = worstWindow(all, time.Minute)
	c.Worst1h = worstWindow(all, time.Hour)
	for _, g := range c.Gaps {
		for _, bar := range []float64{0.5, 0.8, 10, 30, 60} {
			if g < bar {
				c.Under[bar]++
			}
		}
	}
	sort.Float64s(c.Gaps)
	return c, nil
}

func worstWindow(ts []time.Time, w time.Duration) int {
	best, j := 0, 0
	for i := range ts {
		for ts[i].Sub(ts[j]) > w {
			j++
		}
		if n := i - j + 1; n > best {
			best = n
		}
	}
	return best
}

func (c cadence) pct(p float64) float64 {
	if len(c.Gaps) == 0 {
		return 0
	}
	i := int(p * float64(len(c.Gaps)-1))
	return c.Gaps[i]
}

// starfallCadenceHTML is the section. It states the duty cycle at the proposed
// duration and the worst burst his log actually contains, because "occasional"
// is his own word for what he asked for and it is a claim that can be checked.
func starfallCadenceHTML(dur float64) string {
	c, err := starfallCadenceOf()
	if err != nil {
		return fmt.Sprintf(`<p class="nt">No recordings to measure: %v. `+
			`The cadence is therefore UNMEASURED on this machine and nothing below claims it.</p>`, err)
	}
	if c.Arrivals == 0 {
		return `<p class="nt">The spools hold no arrivals at all: the reducer's TodoDone never ` +
			`rose in any recorded session. That would make this feature dead code, and it is ` +
			`exactly the trap that left the constellation dark until session 27.</p>`
	}
	var b strings.Builder
	duty := 100 * dur / math.Max(c.pct(0.5), 1e-9)
	fmt.Fprintf(&b, `<p class="nt">Folded through the real reducer at the events' own timestamps, `+
		`over <b>%d spools</b> and <b>%s events</b>. An arrival is a frame where TodoDone goes up, `+
		`which is the product's own definition &mdash; so this counts the closed turns and finished `+
		`subagents that really light stars, and excludes the todo events that have never fired.</p>`,
		c.Spools, commaOf(c.Events))
	fmt.Fprintf(&b, `<table><tr><th>arrivals</th><th>gap p10</th><th>gap p50</th><th>gap p90</th>`+
		`<th>under 10s</th><th>under 60s</th><th>worst 10s</th><th>worst minute</th><th>worst hour</th></tr>`+
		`<tr><td>%d</td><td>%s</td><td><b>%s</b></td><td>%s</td><td>%d</td><td>%d</td>`+
		`<td>%d</td><td>%d</td><td>%d</td></tr></table>`,
		c.Arrivals, durOf(c.pct(0.1)), durOf(c.pct(0.5)), durOf(c.pct(0.9)),
		c.Under[10], c.Under[60], c.Worst10s, c.Worst60s, c.Worst1h)
	fmt.Fprintf(&b, `<p class="nt">At the median gap a <b>%.2f s</b> fall is on screen `+
		`<b>%.3f%% of frames</b>. That is an event, not a texture.<br>`+
		`But the median is not the whole story: <b>%d gaps are under ten seconds</b> and the worst `+
		`burst in his log is <b>%d arrivals inside ten seconds</b>. His own word for what he asked `+
		`for was <i>"the occational shooting star"</i>. Coalescing (at most one fall in flight) `+
		`prevents overlap; only a minimum gap makes "occasional" true.</p>`,
		dur, duty, c.Under[10], c.Worst10s)
	return b.String()
}

func durOf(s float64) string {
	switch {
	case s < 90:
		return fmt.Sprintf("%.0f s", s)
	case s < 5400:
		return fmt.Sprintf("%.1f min", s/60)
	default:
		return fmt.Sprintf("%.1f h", s/3600)
	}
}

func commaOf(n int) string {
	s := fmt.Sprint(n)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}
