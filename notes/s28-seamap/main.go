// Command s28-seamap asks the sea's mapping question at FOUR geometries instead
// of one, and asks it in a form that can be refuted: for every channel, at every
// level, it reports the reading AND the channel's own frame-to-frame noise. A
// plateau claim with no noise floor beside it is unfalsifiable -- "it stops
// moving" is worthless unless you know how much it moves when nothing changes.
//
//	go run ./notes/s28-seamap                 # the tide, which is the default
//	XSCAPES_TIDE=0 go run ./notes/s28-seamap  # the fixed waterline it replaced
//	go run ./notes/s28-seamap -probe          # colour dump, to justify "warm"
//	go run ./notes/s28-seamap -iso            # isolates the writing-band guard
//	go run ./notes/s28-seamap -menu           # costs the fix options
//
// WHAT IS MEASURED, and all of it off the RESOLVED frame (canvas.ResolveAt at
// term.Profile256), never off the source:
//
//   - whitecaps  the U+224B cells in the sea. The break cue.
//   - wavecells  every wave-ramp glyph in the sea. Coverage, which is the
//     channel the encoding rule says carries activity.
//   - bands      how many crest LINES a viewer could count: local maxima in the
//     per-row coverage profile. The swell count is 2+int(level*5.4) in source;
//     this is what survives into the picture.
//   - swing      the waterline's spatial range, max column minus min column.
//   - crests     local minima along the waterline.
//   - tiderow    the waterline's mean row, averaged over WHOLE wash cycles.
//   - wash       the waterline's TEMPORAL swing: the sd of the mean row across
//     frames. Under the tide this is a channel in its own right
//     (wash = (0.7+level*0.9)*scale) and notes/searange never had a
//     column for it, because before the tide it did not exist.
//
// The waterline is read from the BACKGROUND -- first row down the column whose
// background is warm (R > B) -- because that is where the channel lives. A
// glyph-only reader cannot see it at all, and that mistake already downgraded
// one verdict in this repo to a lead.
//
// TWO TRAPS THIS CODE IS BUILT AROUND:
//
//  1. ONE shore per level, warmed up. A fresh scape.NewShore per frame has no
//     history and Shore.Update clamps any gap over a second to one nominal
//     step, so two frames off two new Shores can be identical.
//  2. WHOLE WASH CYCLES. The wash is sin(tt*0.5) and tt advances at
//     dt*(0.55+level*1.45), so a wash cycle is 572 frames at level 0 and 157 at
//     level 1. notes/searange samples 80 frames at every level, which is 14% of
//     a cycle at rest -- the mean waterline it prints for a quiet sea is
//     measured over a fraction of one wave and can be off by most of the wash's
//     own amplitude. Every level here samples at least two full cycles.
package main

import (
	"flag"
	"fmt"
	"math"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

const dt = 0.08 // simulated seconds per frame, same step notes/searange uses

type geom struct {
	name string
	w, h int
}

var geoms = []geom{
	{"40x12   floor", 40, 12},
	{"80x24  target", 80, 24},
	{"111x25   his", 111, 25},
	{"153x51  wide", 153, 51},
}

var levels = []float64{
	0.00, 0.05, 0.10, 0.15, 0.20, 0.25, 0.30, 0.35, 0.40, 0.45, 0.50,
	0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90, 0.95, 1.00,
}

// waveGlyph is the sea's ramp plus its whitecap. Glitter is U+2022 and foam is
// U+2218/U+00B0, so neither is counted here; the sand's U+00B7 is excluded by
// the region, which stops two rows short of the waterline.
func waveGlyph(r rune) bool {
	switch r {
	case '·', '-', '~', '≈', '≋':
		return true
	}
	return false
}

// reading is one channel's sample set at one level.
type reading struct{ v []float64 }

func (r *reading) add(x float64) { r.v = append(r.v, x) }
func (r reading) mean() float64 {
	if len(r.v) == 0 {
		return math.NaN()
	}
	s := 0.0
	for _, x := range r.v {
		s += x
	}
	return s / float64(len(r.v))
}
func (r reading) sd() float64 {
	if len(r.v) < 2 {
		return math.NaN()
	}
	m, s := r.mean(), 0.0
	for _, x := range r.v {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(r.v)-1))
}

type row struct {
	lv                                        float64
	whitecap, wave, bands, swing, crests, top reading
	washSD                                    float64 // sd of the mean waterline across frames
	frames                                    int
	seaCells                                  int // total wave-glyph cells seen, the non-empty proof
	noWarm                                    int // columns with no warm row at all -- a void reading
	aboveHorizon                              int // columns clipped at the horizon: no sea left
	skyWarm                                   int // columns with a warm SKY cell -- the sun and moon, excluded
	seaRows                                   float64
}

func main() {
	probe := flag.Bool("probe", false, "dump a column of resolved backgrounds and exit")
	iso := flag.Bool("iso", false, "isolate the writing-band clamp by varying the room it leaves")
	menu := flag.Bool("menu", false, "cost the fix options against a model validated on the rendered frame")
	flag.Parse()
	if *menu {
		costMenu()
		return
	}
	if *probe {
		probeColumn()
		return
	}
	if *iso {
		isolate()
		return
	}

	fmt.Printf("XSCAPES TIDE = %v   (TideRange %.1f rows, TideEase %.1fs)\n\n", scape.Tide, scape.TideRange, scape.TideEase)

	for _, g := range geoms {
		rows := make([]row, 0, len(levels))
		for _, lv := range levels {
			rows = append(rows, sample(g, lv))
		}
		printGeom(g, rows)
	}
}

// ---------------------------------------------------------------------------
// ISOLATION: is the top end of the tide squashed by the writing band's guard?
//
// Shore.Update rescales the swell so it never reaches the writing band. The
// reference it measures from is the TIDE's row, so the room the guard allows is
// (writeTop-1) - tideRow, and at full activity the tide has come all the way in
// and that room is `beach - WriteRows - 1` rows. At 80x24 and 111x25 that is
// TWO rows against a wash-plus-coast deviation of about 2.6, so the guard fires
// -- and only at the busy end, which is exactly where the plateau is.
//
// This does not argue it. SandRows is an exported field and it moves `beach`
// and nothing else about the tide, so raising it varies the room ON ITS OWN and
// leaves the tide's own arithmetic untouched. If the top end opens up when the
// room does, the guard is the cause. If it does not, this hypothesis is dead.
func isolate() {
	fmt.Printf("XSCAPES TIDE = %v -- ISOLATING THE WRITING-BAND GUARD\n", scape.Tide)
	fmt.Println("varying ONLY Shore.SandRows, which moves the guard's room and nothing else in the tide")
	fmt.Println()
	for _, g := range []geom{{"80x24", 80, 24}, {"111x25", 111, 25}, {"153x51", 153, 51}} {
		wr := scape.DefaultWriteRows
		if m := g.h / 6; wr > m {
			wr = m
		}
		beach := g.h / 5
		if min := wr + 3; beach < min {
			beach = min
		}
		scale := math.Max(0.45, math.Min(1.35, math.Min(float64(g.w)/80.0, float64(g.h)/24.0)))
		fmt.Printf("--- %s: shipped beach=%d rows, band=%d, so the guard's room at full activity is %d rows;\n",
			g.name, beach, wr, beach-wr-1)
		fmt.Printf("    the wash plus the coast asks for %.2f rows there. Guard fires: %v\n",
			2.55*scale, float64(beach-wr-1) < 2.55*scale)
		fmt.Printf("    %-16s %8s %8s %8s %8s %8s %8s %8s %8s %8s\n",
			"SandRows", "0.60", "0.70", "0.80", "0.90", "1.00", "d(.7-1)", "wash.7", "wash1.0", "d(wash)")
		for _, sr := range []int{0, beach + 3, beach + 6} {
			label := fmt.Sprintf("%d (shipped)", beach)
			if sr > 0 {
				label = fmt.Sprintf("%d", sr)
			}
			var tops, washes []float64
			for _, lv := range []float64{0.60, 0.70, 0.80, 0.90, 1.00} {
				r := sampleWith(g, lv, func(s *scape.Shore) { s.SandRows = sr })
				tops = append(tops, r.top.mean())
				washes = append(washes, r.washSD)
			}
			fmt.Printf("    %-16s %8.2f %8.2f %8.2f %8.2f %8.2f %8.2f %8.2f %8.2f %8.2f\n",
				label, tops[0], tops[1], tops[2], tops[3], tops[4], tops[4]-tops[1],
				washes[1], washes[4], washes[4]-washes[1])
		}
		fmt.Println()
	}
}

// sample runs one level at one geometry and returns every channel's samples.
func sample(g geom, lv float64) row { return sampleWith(g, lv, nil) }

func sampleWith(g geom, lv float64, tweak func(*scape.Shore)) row {
	sh := scape.NewShore(7, false)
	if tweak != nil {
		tweak(sh)
	}
	act := scape.Activity{Level: lv, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	hy := int(float64(g.h) * 0.42)

	// Two whole wash cycles at this level, never fewer than 220 frames.
	cyc := 2 * (2 * math.Pi / 0.5) / (dt * (0.55 + lv*1.45))
	n := int(math.Ceil(cyc))
	if n < 220 {
		n = 220
	}

	tm := 0.0
	for k := 0; k < 500; k++ { // warm-up: ~40s, over ten TideEase constants
		tm += dt
		sh.Update(canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm, act)
	}

	out := row{lv: lv, frames: n}
	var meanRows []float64
	for k := 0; k < n; k++ {
		c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += dt
		sh.Update(c, tm, act)

		// The waterline, off the background: the first row down the column
		// whose background is warm, because the sea is blue and the beach is
		// warm and that boundary needs no colour matching.
		//
		// ⚠ THE SCAN STARTS AT hy+1 AND THAT IS NOT A DETAIL. The first
		// version of this instrument scanned from row 0 to catch a tide that
		// had withdrawn past the horizon, and caught the SUN and the MOON
		// instead -- both are painted warm into the sky background, so every
		// column under a disc reported a waterline up in the sky. It inflated
		// the swing at 111x25 from ~2 rows to 12.4 and put 34,000 phantom
		// "above the horizon" columns on the record. Counted below by
		// skyWarm, so the correction is visible rather than assumed.
		prof := make([]float64, g.w)
		for x := 0; x < g.w; x++ {
			for y := 0; y <= hy; y++ {
				if _, _, bg := c.ResolveAt(x, y, term.Profile256); bg.R > bg.B {
					out.skyWarm++
					break
				}
			}
			prof[x] = float64(g.h)
			for y := hy + 1; y < g.h; y++ {
				_, _, bg := c.ResolveAt(x, y, term.Profile256)
				if bg.R > bg.B {
					prof[x] = float64(y)
					break
				}
			}
			if prof[x] == float64(g.h) {
				out.noWarm++
			}
			// The scan floor. A column reading hy+1 has no sea left in it at
			// all: the water has withdrawn to the horizon or past it, and the
			// reading is CLIPPED, not measured.
			if prof[x] <= float64(hy)+1 {
				out.aboveHorizon++
			}
		}

		// Glyph channels, over the sea only: below the horizon and two rows
		// clear of the waterline, which is where foam lives.
		wc, wv := 0, 0
		cov := make([]float64, g.h)
		den := make([]float64, g.h)
		for x := 0; x < g.w; x++ {
			for y := hy + 1; float64(y) <= prof[x]-2; y++ {
				den[y]++
				r, _, _ := c.ResolveAt(x, y, term.Profile256)
				if !waveGlyph(r) {
					continue
				}
				wv++
				cov[y]++
				if r == '≋' {
					wc++
				}
			}
		}
		out.seaCells += wv
		out.whitecap.add(float64(wc))
		out.wave.add(float64(wv))

		// Bands: local maxima in the per-row coverage profile, over rows where
		// at least a third of the columns are sea. This is the countable
		// thing -- how many crest LINES are in the water.
		frac := make([]float64, g.h)
		for y := range cov {
			if den[y] >= float64(g.w)/3 {
				frac[y] = cov[y] / den[y]
			}
		}
		b := 0
		for y := 1; y < g.h-1; y++ {
			if frac[y] > 0.15 && frac[y] > frac[y-1] && frac[y] >= frac[y+1] {
				b++
			}
		}
		out.bands.add(float64(b))

		lo, hi, sum := math.Inf(1), math.Inf(-1), 0.0
		for _, v := range prof {
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
			sum += v
		}
		out.swing.add(hi - lo)
		mr := sum / float64(g.w)
		out.top.add(mr)
		meanRows = append(meanRows, mr)

		pk := 0
		for x := 2; x < g.w-2; x++ {
			if prof[x] < prof[x-2] && prof[x] <= prof[x+2] {
				pk++
			}
		}
		out.crests.add(float64(pk))
	}
	// The wash is the waterline's motion IN TIME, which is a different channel
	// from its shape in space and is the one the tide added.
	out.washSD = reading{meanRows}.sd()
	out.seaRows = out.top.mean() - float64(hy)
	return out
}

func printGeom(g geom, rows []row) {
	hy := int(float64(g.h) * 0.42)
	scale := math.Max(0.45, math.Min(1.35, math.Min(float64(g.w)/80.0, float64(g.h)/24.0)))
	fmt.Printf("================ %s   horizon row %d, scale %.2f ================\n", g.name, hy, scale)
	fmt.Printf("  sample: %d levels, %d-%d frames each (>=2 whole wash cycles), %d wave cells seen in total\n",
		len(rows), rows[len(rows)-1].frames, rows[0].frames, totalCells(rows))
	void := 0
	for _, r := range rows {
		void += r.noWarm
	}
	fmt.Printf("  columns with NO warm cell (a void reading): %d of %d scanned\n", void, scannedCols(rows, g.w))
	sw := 0
	for _, r := range rows {
		sw += r.skyWarm
	}
	fmt.Printf("  columns with a warm SKY cell (the sun and moon, why the scan starts below the horizon): %d of %d\n",
		sw, scannedCols(rows, g.w))
	ah := 0
	for _, r := range rows {
		ah += r.aboveHorizon
	}
	if ah > 0 {
		fmt.Printf("  !! columns CLIPPED at the horizon -- no sea left in them: %d of %d\n", ah, scannedCols(rows, g.w))
		fmt.Print("     by level:")
		for _, r := range rows {
			if r.aboveHorizon > 0 {
				fmt.Printf("  %.2f=%.0f%%", r.lv, 100*float64(r.aboveHorizon)/float64(r.frames*g.w))
			}
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Printf("  %-5s %13s %13s %11s %11s %11s %13s %8s %7s\n",
		"level", "whitecaps", "wavecells", "bands", "swing", "crests", "waterline", "wash", "searows")
	for _, r := range rows {
		fmt.Printf("  %-5.2f %6.1f+-%-5.2f %6.1f+-%-5.1f %5.2f+-%-4.2f %5.2f+-%-4.2f %5.2f+-%-4.2f %6.2f+-%-5.2f %8.2f %7.2f\n",
			r.lv,
			r.whitecap.mean(), r.whitecap.sd(),
			r.wave.mean(), r.wave.sd(),
			r.bands.mean(), r.bands.sd(),
			r.swing.mean(), r.swing.sd(),
			r.crests.mean(), r.crests.sd(),
			r.top.mean(), r.top.sd(),
			r.washSD, r.seaRows)
	}
	fmt.Println()
	verdict(g, rows)
	fmt.Println()
}

func totalCells(rows []row) int {
	n := 0
	for _, r := range rows {
		n += r.seaCells
	}
	return n
}
func scannedCols(rows []row, w int) int {
	n := 0
	for _, r := range rows {
		n += r.frames * w
	}
	return n
}

type chanDef struct {
	name string
	get  func(row) reading
}

// verdict turns each channel's table into the two numbers that decide whether
// it is alive above 0.70: how many noise-widths of signal are left up there,
// and where the reading last moved by more than its own noise.
func verdict(g geom, rows []row) {
	defs := []chanDef{
		{"whitecaps", func(r row) reading { return r.whitecap }},
		{"wavecells", func(r row) reading { return r.wave }},
		{"bands", func(r row) reading { return r.bands }},
		{"swing", func(r row) reading { return r.swing }},
		{"crests", func(r row) reading { return r.crests }},
		{"waterline", func(r row) reading { return r.top }},
	}
	// "steps" is the honest headline: how many readings a GLANCE can tell apart
	// across the whole range, given the channel's own frame-to-frame noise. A
	// screenshot is one frame, so the per-frame sd is the right noise for it.
	// "rise" is the same interval in Weber terms, because a viewer judges 400
	// against 280 cells by ratio, not by difference.
	fmt.Printf("  %-10s %7s %8s %8s %8s %7s %6s %6s %6s  %s\n",
		"channel", "noise", "0.00", "0.70", "1.00", "0.70-1", "steps", "d'.7-1", "rise%", "last level that moved > noise")
	for _, d := range defs {
		var pooled, m0, m70, m100 float64
		var sds []float64
		for _, r := range rows {
			if s := d.get(r).sd(); !math.IsNaN(s) {
				sds = append(sds, s)
			}
		}
		sort.Float64s(sds)
		if len(sds) > 0 {
			pooled = sds[len(sds)/2] // median per-frame sd across the level grid
		}
		m0 = d.get(rows[0]).mean()
		m70 = d.get(rows[14]).mean() // level 0.70
		m100 = d.get(rows[20]).mean()
		full := m100 - m0
		upper := m100 - m70
		var pct float64
		if full != 0 {
			pct = 100 * upper / full
		}
		var dprime float64
		s70, s100 := d.get(rows[14]).sd(), d.get(rows[20]).sd()
		if den := math.Sqrt((s70*s70 + s100*s100) / 2); den > 0 {
			dprime = math.Abs(upper) / den
		}
		// Last level whose 0.10 step exceeded the channel's own noise.
		last := -1.0
		for i := 0; i+2 < len(rows); i++ {
			if math.Abs(d.get(rows[i+2]).mean()-d.get(rows[i]).mean()) > pooled {
				last = rows[i].lv
			}
		}
		ls := "never -- dead at every level"
		if last >= 0 {
			ls = fmt.Sprintf("%.2f", last)
			if last >= 0.70 {
				ls += "  ALIVE above 0.70"
			} else {
				ls += "  PLATEAU from here"
			}
		}
		steps := 0.0
		if pooled > 0 {
			steps = math.Abs(full) / pooled
		}
		rise := math.NaN()
		if m70 != 0 {
			rise = 100 * (m100/m70 - 1)
		}
		fmt.Printf("  %-10s %7.2f %8.2f %8.2f %8.2f %6.0f%% %6.1f %6.2f %6.0f  %s\n",
			d.name, pooled, m0, m70, m100, pct, steps, dprime, rise, ls)
	}
	// The wash has one number per level, not a distribution, so it gets its own
	// line: how much the waterline moves in TIME, which is what the tide added.
	fmt.Printf("  %-10s %7s %8.2f %8.2f %8.2f %6.0f%% %6s %6s %6.0f  %s\n", "wash(sd)", "n/a",
		rows[0].washSD, rows[14].washSD, rows[20].washSD,
		pctOf(rows[20].washSD-rows[14].washSD, rows[20].washSD-rows[0].washSD), "n/a", "n/a",
		100*(rows[20].washSD/rows[14].washSD-1),
		"the waterline's swing IN TIME, which only the tide has")
	_ = g
}

func pctOf(a, b float64) float64 {
	if b == 0 {
		return math.NaN()
	}
	return 100 * a / b
}

// ---------------------------------------------------------------------------
// COSTING THE MENU.
//
// Every option below changes arithmetic inside Shore.Update, and no exported
// knob reaches it -- so rather than edit a product file that two other sessions
// are working in, the arithmetic is COPIED here verbatim and varied. That is
// only worth anything if the copy reproduces what the renderer actually draws,
// so `-menu` prints the SHIPPED model beside the SHIPPED measurement first. Read
// that line before reading any projection: if it does not match the rendered
// table, nothing under it is worth quoting.
type layout struct {
	hy, wr, beach, sy, writeTop int
	scale                       float64
}

func layoutOf(g geom, sandRows, bandFloor int) layout {
	l := layout{hy: int(float64(g.h) * 0.42)}
	l.wr = scape.DefaultWriteRows
	if m := g.h / 6; l.wr > m {
		l.wr = m
	}
	if l.wr < 2 {
		l.wr = 0
	}
	l.writeTop = g.h
	if l.wr > 0 && g.h > l.wr+4 {
		l.writeTop = g.h - l.wr
	}
	l.beach = g.h / 5
	if min := l.wr + bandFloor; l.wr > 0 && l.beach < min {
		l.beach = min
	}
	if l.beach < 4 {
		l.beach = 4
	}
	if sandRows > 0 {
		l.beach = sandRows
	}
	l.scale = math.Max(0.45, math.Min(1.35, math.Min(float64(g.w)/80.0, float64(g.h)/24.0)))
	l.sy = g.h - l.beach
	if l.sy <= l.hy+1 {
		l.sy = l.hy + 2
	}
	if l.sy > l.writeTop-1 {
		l.sy = l.writeTop - 1
	}
	return l
}

type opt struct {
	name      string
	bandFloor int  // the `wr + N` beach floor in Update
	washFree  bool // the guard measures the coast only, letting the tide's own wash through
	fitTide   bool // TideRange shrinks to the sea's actual depth so the water cannot reach the horizon
	tideRange float64
	// reserve keeps the tide this many rows short of sy at FULL activity, and
	// takes those rows out of the tide's own range rather than out of the
	// resting position -- so the quiet end of the picture is unchanged and the
	// guard gets the room it needs exactly where it was firing.
	reserve float64
}

// modelTide replays the tide arithmetic for one level and returns the mean
// waterline and the wash's sd, both over whole wash cycles.
func modelTide(g geom, lv float64, o opt) (top, washSD, seaRows, hitBand float64) {
	tr := scape.TideRange
	if o.tideRange > 0 {
		tr = o.tideRange
	}
	bf := 3
	if o.bandFloor > 0 {
		bf = o.bandFloor
	}
	l := layoutOf(g, 0, bf)
	rng := tr * l.scale
	if o.fitTide {
		// Never withdraw so far that the sea is gone: keep two rows of water
		// below the horizon at the quietest.
		if d := float64(l.sy-l.hy) - 2; rng > d {
			rng = math.Max(0, d)
		}
	}
	at := (1 - lv) * rng
	if o.reserve > 0 && rng > o.reserve {
		at = o.reserve + (1-lv)*(rng-o.reserve)
	}
	tideRow := float64(l.sy) - math.Round(at)
	base := float64(l.sy) - at
	if fl := float64(l.sy) - tr - 1; base < fl {
		base = fl
	}
	amp := (0.7 + lv*0.9) * l.scale

	coast := make([]float64, g.w)
	for x := 0; x < g.w; x++ {
		fx := float64(x)
		coast[x] = 0.60*l.scale*math.Sin(fx*0.11) + 0.35*l.scale*math.Sin(fx*0.047)
	}

	const nph = 400
	means := make([]float64, 0, nph)
	hits, cells := 0, 0
	for k := 0; k < nph; k++ {
		wash := amp * math.Sin(2*math.Pi*float64(k)/nph)
		sum := 0.0
		e := make([]float64, g.w)
		for x := 0; x < g.w; x++ {
			e[x] = base + wash + coast[x]
		}
		if l.writeTop < g.h {
			ref := tideRow
			room := math.Max(0.5, float64(l.writeTop-1)-ref)
			dev := 0.0
			for x := 0; x < g.w; x++ {
				d := math.Abs(e[x] - ref)
				if o.washFree {
					// The guard exists to keep the SWELL out of the writing.
					// The tide's own wash is already bounded by TideRange, so
					// measure the deviation without it and let the sheet move.
					d = math.Abs(base + coast[x] - ref)
				}
				if d > dev {
					dev = d
				}
			}
			if dev > room {
				k := room / dev
				for x := 0; x < g.w; x++ {
					e[x] = ref + (e[x]-ref)*k
				}
			}
		}
		for x := 0; x < g.w; x++ {
			cells++
			// The cost line for any option that loosens the guard: a column
			// that reaches the writing band is a column whose waterline lands
			// on the SAME row as its neighbours, which is the ruled shoreline
			// the guard was written to prevent.
			if e[x] > float64(l.writeTop)-1 {
				hits++
				e[x] = float64(l.writeTop) - 1
			}
			sum += e[x]
		}
		means = append(means, sum/float64(g.w))
	}
	return reading{means}.mean(), reading{means}.sd(), reading{means}.mean() - float64(l.hy),
		100 * float64(hits) / float64(cells)
}

func costMenu() {
	if !scape.Tide {
		fmt.Println("run this with the tide ON -- it costs tide options")
		return
	}
	opts := []opt{
		{name: "SHIPPED"},
		{name: "A beach floor wr+4", bandFloor: 4},
		{name: "B guard skips the wash", washFree: true},
		{name: "C both A and B", bandFloor: 4, washFree: true},
		{name: "D fit tide to sea depth", fitTide: true},
		{name: "E C + D + range 6.5", bandFloor: 4, washFree: true, fitTide: true, tideRange: 6.5},
		{name: "F reserve 1 row at full", reserve: 1},
		{name: "G F + D  (recommended)", reserve: 1, fitTide: true},
		{name: "H G + range 6.0", reserve: 1, fitTide: true, tideRange: 6.0},
	}
	for _, g := range geoms {
		fmt.Printf("================ %s ================\n", g.name)
		fmt.Printf("  %-24s %8s %8s %8s %7s %7s %7s %7s %7s %7s\n",
			"option", "top@0.0", "top@0.7", "top@1.0", "range", "wash@0", "wash@.7", "wash@1", "top30%", "band%")
		for _, o := range opts {
			t0, w0, sr0, _ := modelTide(g, 0.00, o)
			t7, w7, _, _ := modelTide(g, 0.70, o)
			t1, w1, _, hb := modelTide(g, 1.00, o)
			note := ""
			if sr0 < 1.5 {
				note = fmt.Sprintf("   !! %.1f rows of sea at rest", sr0)
			}
			// top30 is the share of the WASH channel's whole range that the
			// 0.70-1.00 band carries. 30% is what a linear channel gives.
			top30 := 100 * (w1 - w7) / (w1 - w0)
			fmt.Printf("  %-24s %8.2f %8.2f %8.2f %7.2f %7.2f %7.2f %7.2f %6.0f%% %6.2f%%%s\n",
				o.name, t0, t7, t1, t1-t0, w0, w7, w1, top30, hb, note)
		}
		fmt.Println()
	}
}

// probeColumn justifies the "warm means sand" test by printing what the
// backgrounds actually are down one column, rather than asserting it.
func probeColumn() {
	const w, h = 111, 25
	sh := scape.NewShore(7, false)
	act := scape.Activity{Level: 0.5, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	tm := 0.0
	for k := 0; k < 500; k++ {
		tm += dt
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm, act)
	}
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	tm += dt
	sh.Update(c, tm, act)
	hh := h
	fmt.Printf("column 40 at 111x25, level 0.50, 13:00. horizon row %d\n", int(float64(hh)*0.42))
	for y := 0; y < h; y++ {
		r, _, bg := c.ResolveAt(40, y, term.Profile256)
		warm := ""
		if bg.R > bg.B {
			warm = "  <- WARM"
		}
		fmt.Printf("  y=%2d  bg=(%3d,%3d,%3d)  glyph=%q%s\n", y, bg.R, bg.G, bg.B, r, warm)
	}
}
