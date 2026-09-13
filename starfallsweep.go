package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donlucasx/xscapes/internal/scape"
)

// The sweep exists because this project has been burned twice by a measurement
// taken at one seed, one geometry or one hour. Everything it reports is read
// off a COMPOSED frame -- after drawScene -- because the opaque ask balloon and
// the readout are drawn into the same near layer the stars live in, and a
// sweep run on an un-composed frame is the repo's own "harness never enters
// production mode" failure.

type sweepRow struct {
	Label                 string
	Arrivals, NoStar      int
	Streak, InPlace       int
	CountErr, Eaten       int
	OffCanvas, OnDisc     int
	MinTrack, MaxTrack    int
	Stops                 map[string]int
	Exhausted, FullAlpha  int
	WorstContrast, MinSep float64
}

type sweepCfg struct {
	Label      string
	W, H       int
	Seed       int64
	Tod, Ctx   float64
	Mirror     bool
	Bubble     string
	Ask        bool
	Stars      int
	Track      trackCfg
	Dur, Curve float64
}

// sweepRun folds one configuration. Home cells are recovered by advancing the
// count ONE AT A TIME and taking the cell that appears, so the study follows
// whatever the product's placement does today rather than replicating it.
func sweepRun(cfg sweepCfg) sweepRow {
	row := sweepRow{Label: cfg.Label, Stops: map[string]int{},
		MinTrack: 1 << 30, WorstContrast: 1e18, MinSep: 1e18}

	base := starScene{W: cfg.W, H: cfg.H, Seed: cfg.Seed, Tod: cfg.Tod, Ctx: cfg.Ctx,
		Total: 32, Mirror: cfg.Mirror, Bubble: cfg.Bubble, Ask: cfg.Ask, T: 4}
	forb := starfallForbidden(base)
	_, sh, _, _, _ := starfallStage(base)

	prev := map[starPt]bool{}
	{
		s := base
		s.Done = 0
		c, _ := starfallRender(s)
		for _, p := range starCellsIn(c) {
			prev[p] = true
		}
	}
	for done := 1; done <= cfg.Stars; done++ {
		s := base
		s.Done = done
		s.Mag = scape.StarMagnitudeAt(done-1, cfg.Seed)
		c, _ := starfallRender(s)
		cur := starCellsIn(c)
		var fresh []starPt
		seen := map[starPt]bool{}
		for _, p := range cur {
			seen[p] = true
			if !prev[p] {
				fresh = append(fresh, p)
			}
		}
		var lit []starPt
		for p := range prev {
			lit = append(lit, p)
		}
		prev = seen
		if len(fresh) != 1 {
			row.NoStar++
			continue
		}
		row.Arrivals++
		home := fresh[0]
		track, why := starfallTrack(s, home, lit, forb, sh, cfg.Track)
		row.Stops[why]++
		if len(track) < 2 {
			row.InPlace++
		} else {
			row.Streak++
		}
		if len(track) < row.MinTrack {
			row.MinTrack = len(track)
		}
		if len(track) > row.MaxTrack {
			row.MaxTrack = len(track)
		}

		// The baseline is the SAME frame without the mark, not `done`. Small
		// geometries already draw fewer stars than the count -- 40x12 draws 21
		// of 32 -- for reasons that predate this feature, and charging that to
		// the fall was the first sweep's error: it reported 396 count errors,
		// none of which the fall caused.
		settled := s
		settled.Done, settled.Marks = done-1, nil
		sc2, _ := starfallRender(settled)
		want := len(starCellsIn(sc2)) + 1

		// Three phases per arrival. The count can only go wrong two ways --
		// the head is eaten by something drawn later, or it lands on a cell a
		// star already holds -- and both show at the head cell.
		for _, p := range []float64{0, 0.5, 1} {
			f := s
			f.Done = done - 1
			f.Marks = starfallMarks(track, p, cfg.Curve, cfg.Track.tail())
			fc, reps := starfallRender(f)
			ps := starCellsIn(fc)
			if len(ps) != want {
				row.CountErr++
			}
			if d := starMinSep(ps); d >= 0 && d < row.MinSep {
				row.MinSep = d
			}
			for _, r := range reps {
				if r.Eaten {
					row.Eaten++
				}
				if r.Exhausted {
					row.Exhausted++
				}
				if r.FullAlpha {
					row.FullAlpha++
				}
				if r.Contrast < row.WorstContrast {
					row.WorstContrast = r.Contrast
				}
				if r.X < 0 || r.X >= cfg.W || r.Y < 0 || r.Y >= cfg.H {
					row.OffCanvas++
				}
				if sh.DiscCovers(r.X, r.Y) {
					row.OnDisc++
				}
			}
		}
	}
	if row.MinTrack == 1<<30 {
		row.MinTrack = 0
	}
	if row.WorstContrast > 1e17 {
		row.WorstContrast = 0
	}
	if row.MinSep > 1e17 {
		row.MinSep = 0
	}
	return row
}

// tail is empty for the sweep: the proposal ships without one, and sweeping a
// tail that is not proposed would measure a thing nobody is being asked about.
func (c trackCfg) tail() []rune { return nil }

// starfallPopulation is the population, stated rather than implied: a full
// cross of every geometry against every hour, then one axis at a time around
// the base scene. It is NOT a full cross of all five axes -- that is 432
// configurations and about seven minutes -- and saying so is the point.
func starfallPopulation(seed int64) []sweepCfg {
	tc := trackCfg{N: 2, Rmax: 6, DyMax: 3, TrackMin: 4, Sep: 2}
	geos := []struct {
		w, h int
	}{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {125, 28}, {143, 27}, {153, 51}, {125, 62}}
	hours := []float64{0, 6. / 24, 12. / 24, 14. / 24, 18. / 24, 21. / 24}

	var out []sweepCfg
	mk := func(label string, w, h int, sd int64, tod, ctx float64, mirror bool, bub string, ask bool) {
		out = append(out, sweepCfg{Label: label, W: w, H: h, Seed: sd, Tod: tod, Ctx: ctx,
			Mirror: mirror, Bubble: bub, Ask: ask, Stars: 32, Track: tc, Dur: 0.50, Curve: 1})
	}
	// Every geometry against every hour.
	for _, g := range geos {
		for i, hr := range hours {
			// The balloon alternates so both states are covered at every size.
			bub, ask := "", false
			if i%2 == 1 {
				bub, ask = "allow Bash?", true
			}
			mk(fmt.Sprintf("%dx%d @ %02.0f:00", g.w, g.h, hr*24), g.w, g.h, seed, hr, 0.56, true, bub, ask)
		}
	}
	// One axis at a time around the base.
	for _, sd := range []int64{1, 7, 42} {
		mk(fmt.Sprintf("seed %d", sd), 125, 28, sd, 18./24, 0.56, true, "", false)
	}
	for _, ctx := range []float64{0.05, 0.50, 0.95} {
		mk(fmt.Sprintf("context %.2f", ctx), 125, 28, seed, 18./24, ctx, true, "", false)
	}
	for _, m := range []bool{true, false} {
		mk(fmt.Sprintf("mirror %v", m), 125, 28, seed, 18./24, 0.56, m, "", false)
	}
	mk("ask balloon up", 125, 28, seed, 18./24, 0.56, true, "allow Bash?", true)
	mk("done balloon up", 125, 28, seed, 18./24, 0.56, true, "tests passed", false)
	return out
}

func starfallSweep(seed int64) []string {
	cfgs := starfallPopulation(seed)
	var rows []sweepRow
	for _, c := range cfgs {
		rows = append(rows, sweepRun(c))
	}
	var out []string
	out = append(out, "THE STAR ARRIVING -- sweep, read off COMPOSED frames (after drawScene)")
	out = append(out, fmt.Sprintf("%d configurations, 32 slots each, 3 phases per arrival.", len(cfgs)))
	out = append(out, "")
	out = append(out, fmt.Sprintf("%-22s %5s %5s %6s %6s %6s %6s %6s %6s %7s",
		"config", "arriv", "strk", "inplc", "cnterr", "eaten", "ondisc", "fullA", "exhstd", "worstC"))
	tot := sweepRow{Stops: map[string]int{}, MinTrack: 1 << 30, WorstContrast: 1e18, MinSep: 1e18}
	for _, r := range rows {
		flag := ""
		if r.CountErr > 0 || r.Eaten > 0 || r.OffCanvas > 0 || r.OnDisc > 0 {
			flag = "   <-- LOOK"
		} else if r.Exhausted > 0 {
			flag = "   <-- ink exhausted"
		}
		out = append(out, fmt.Sprintf("%-22s %5d %5d %6d %6d %6d %6d %6d %6d %7.1f%s",
			r.Label, r.Arrivals, r.Streak, r.InPlace, r.CountErr, r.Eaten,
			r.OnDisc, r.FullAlpha, r.Exhausted, r.WorstContrast, flag))
		tot.Arrivals += r.Arrivals
		tot.NoStar += r.NoStar
		tot.Streak += r.Streak
		tot.InPlace += r.InPlace
		tot.CountErr += r.CountErr
		tot.Eaten += r.Eaten
		tot.OffCanvas += r.OffCanvas
		tot.OnDisc += r.OnDisc
		tot.Exhausted += r.Exhausted
		tot.FullAlpha += r.FullAlpha
		if r.WorstContrast < tot.WorstContrast && r.Arrivals > 0 {
			tot.WorstContrast = r.WorstContrast
		}
		if r.MinSep < tot.MinSep && r.Arrivals > 0 {
			tot.MinSep = r.MinSep
		}
		for k, v := range r.Stops {
			tot.Stops[k] += v
		}
	}
	out = append(out, "")
	out = append(out, fmt.Sprintf("TOTAL  %d arrivals (%d slots gave no single fresh star), %d streak / %d in place (%.1f%% streak)",
		tot.Arrivals, tot.NoStar, tot.Streak, tot.InPlace, 100*float64(tot.Streak)/float64(max(1, tot.Arrivals))))
	out = append(out, fmt.Sprintf("       count errors %d   marks eaten %d   off-canvas %d   on the disc %d",
		tot.CountErr, tot.Eaten, tot.OffCanvas, tot.OnDisc))
	out = append(out, fmt.Sprintf("       ink: %d path cells needed full alpha, %d EXHAUSTED the search; worst contrast %.1f",
		tot.FullAlpha, tot.Exhausted, tot.WorstContrast))
	out = append(out, fmt.Sprintf("       closest pair of stars on any swept frame: %.2f screen units (the bar is 2.00)", tot.MinSep))
	out = append(out, "")
	out = append(out, "why the walk stopped:")
	var keys []string
	for k := range tot.Stops {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return tot.Stops[keys[i]] > tot.Stops[keys[j]] })
	for _, k := range keys {
		out = append(out, fmt.Sprintf("   %-26s %5d", k, tot.Stops[k]))
	}
	return out
}

// starfallSweepHTML is the same sweep, small enough to sit at the top of the
// page: every geometry against every hour. The wider population goes to
// stderr, because a page nobody scrolls is not evidence.
func starfallSweepHTML(seed int64) string {
	tc := trackCfg{N: 2, Rmax: 6, DyMax: 3, TrackMin: 4, Sep: 2}
	geos := []struct{ w, h int }{{143, 27}, {125, 28}, {80, 24}, {40, 12}}
	hours := []float64{0, 12. / 24, 14. / 24, 18. / 24}
	var b strings.Builder
	b.WriteString(`<table><tr><th>geometry</th><th>hour</th><th>arrivals</th><th>streak</th>` +
		`<th>in place</th><th>count errors</th><th>eaten</th><th>on disc</th>` +
		`<th>worst contrast</th><th>closest pair</th></tr>`)
	for _, g := range geos {
		for i, hr := range hours {
			bub, ask := "", false
			if i%2 == 1 {
				bub, ask = "allow Bash?", true
			}
			r := sweepRun(sweepCfg{W: g.w, H: g.h, Seed: seed, Tod: hr, Ctx: 0.56,
				Mirror: true, Bubble: bub, Ask: ask, Stars: 32, Track: tc, Dur: 0.5, Curve: 1})
			cls := func(n int) string {
				if n > 0 {
					return ` class="bad"`
				}
				return ""
			}
			fmt.Fprintf(&b, `<tr><td>%d&times;%d</td><td>%02.0f:00%s</td><td>%d</td><td>%d</td>`+
				`<td>%d</td><td%s>%d</td><td%s>%d</td><td%s>%d</td><td>%.1f</td><td>%.2f</td></tr>`,
				g.w, g.h, hr*24, map[bool]string{true: " + ask", false: ""}[ask],
				r.Arrivals, r.Streak, r.InPlace,
				cls(r.CountErr), r.CountErr, cls(r.Eaten), r.Eaten, cls(r.OnDisc), r.OnDisc,
				r.WorstContrast, r.MinSep)
		}
	}
	b.WriteString(`</table>`)
	return b.String()
}
