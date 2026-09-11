// Command s28-drift-verify is an ADVERSARIAL re-check of notes/s28-drift.
//
// It does not rebuild that instrument. It attacks the three places where its
// headline could be an artifact rather than a reading:
//
//  1. IS 6.44 cells/s A RATE? The run tracker was calibrated against a KNOWN
//     shift and against a REPEATED FRAME (which is 0.000 by construction), but
//     never against a field that CHANGES WITHOUT TRANSLATING. A nearest-run
//     matcher always returns a nonzero |dx| for a decorrelating field, because
//     it matches a mark to whichever neighbour is nearest. Two tests settle it:
//     the pedestal (pair frames far enough apart to be independent) and the
//     gap sweep (a true rate is linear through the origin; a matching artifact
//     saturates at gap 1).
//
//  2. IS 3.00 : 1 A RATIO? It divides CELLS by ROWS. A terminal cell is about
//     twice as tall as it is wide, so cells and rows are not the same distance
//     and the quotient is not a speed ratio. Menlo's own metrics are read off
//     the font file and applied.
//
//  3. IS THE ~3-CELL PULL REALLY REFUTED? notes/s28-drift replays notes/drift
//     200 times on ONE shore advanced continuously, so its 200 trials run from
//     wave age 0 to about 1300 -- and its own section 9b shows the scene's
//     x-structure changes with age. notes/drift measures at wave age 6.6. This
//     replays the same procedure 200 times with every trial YOUNG (a fresh
//     shore, warm-up 0..199 frames, wave age under 30), which is the condition
//     notes/drift actually ran in.
//
//  4. THE CHANNEL THEY DID NOT READ. Both instruments count only the three sea
//     glyphs. The waterline is a BACKGROUND cell (shore.go's own comment says
//     the wave motion "lives in the glyphs and in the waterline cell below,
//     which keeps its own per-column edge"). This reads the background.
//
//  5. THINGS THEY DID NOT VARY: term.NoSplitCells (production on Terminal.app
//     is TRUE; both instruments ran FALSE), time of day, context used, seed.
//
// Every section prints its own sample size, and any section whose sample is
// empty says so instead of reporting a clean number.
//
//	go run ./notes/s28-drift-verify
//	XSCAPES_TIDE=0 go run ./notes/s28-drift-verify
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

var (
	W = 120
	H = 26
)

const (
	seed   = 7
	stepDT = 0.08
)

var act = scape.Activity{Level: 0.6, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}

// The cell aspect ratio, read off Menlo's own metrics with fontTools:
// advance 1233/2048 em = 0.6021, hhea line 2384/2048 em = 1.1641, so a cell is
// 1.933 times as tall as it is wide. Terminal.app adds its own leading on top
// (the repo measured 5.2px unfilled where the font predicts 3.7), so the true
// figure is a little LARGER than this, which makes it a conservative divisor.
const cellAspect = 1.933

func hyRow(h int) int { return int(math.Floor(float64(h) * 0.42)) }

func driftSet(r rune) bool { return r == '≈' || r == '~' || r == '≋' }

type mask [][]bool

// ---------------------------------------------------------------- rendering

// shoreRun renders n frames from ONE shore warmed by `warm` frames. A fresh
// Shore per frame has no history and Update clamps the gap, so two fresh
// Shores render identical frames -- that trap has voided instruments here.
func shoreRun(warm, n int, keep func(rune) bool) []mask {
	sh := scape.NewShore(seed, false)
	t := 0.0
	adv := func() *canvas.Canvas {
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		sh.Update(c, t, act)
		return c
	}
	for i := 0; i < warm; i++ {
		adv()
	}
	out := make([]mask, n)
	for i := range out {
		c := adv()
		m := make(mask, H)
		for y := range m {
			m[y] = make([]bool, W)
			for x := 0; x < W; x++ {
				r, _, _ := c.ResolveAt(x, y, term.Profile256)
				m[y][x] = keep(r)
			}
		}
		out[i] = m
	}
	return out
}

type cell struct {
	r      rune
	fg, bg term.RGB
}

// bgRun renders n frames and marks, in each row, every cell whose FULL
// appearance (glyph + foreground + background) differs from that row's modal
// appearance. Under Tide the sea's background ramp is constant along a row by
// construction, so the marks this produces are the waterline's own per-column
// edge and the wet band -- the background channel neither drift instrument
// read. Returns the masks and the total number of marked cells, so an empty
// sample cannot be mistaken for a clean one.
func bgRun(warm, n int) ([]mask, float64) {
	sh := scape.NewShore(seed, false)
	t := 0.0
	out := make([]mask, n)
	total := 0.0
	for i := 0; i < warm+n; i++ {
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		sh.Update(c, t, act)
		if i < warm {
			continue
		}
		m := make(mask, H)
		row := make([]cell, W)
		for y := 0; y < H; y++ {
			m[y] = make([]bool, W)
			counts := map[cell]int{}
			for x := 0; x < W; x++ {
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				row[x] = cell{r, fg, bg}
				counts[row[x]]++
			}
			var modal cell
			best := -1
			for k, v := range counts {
				if v > best {
					best, modal = v, k
				}
			}
			for x := 0; x < W; x++ {
				if row[x] != modal {
					m[y][x] = true
					total++
				}
			}
		}
		out[i-warm] = m
	}
	return out, total
}

// ------------------------------------------------------------- run tracking

type run struct{ x0, x1 int }

func (r run) centre() float64 { return (float64(r.x0) + float64(r.x1)) / 2 }
func (r run) length() int     { return r.x1 - r.x0 + 1 }

func runsIn(row []bool, x0, x1 int) []run {
	var out []run
	i := x0
	for i < x1 {
		if !row[i] {
			i++
			continue
		}
		j := i
		for j+1 < x1 && row[j+1] {
			j++
		}
		out = append(out, run{i, j})
		i = j + 1
	}
	return out
}

type res struct{ n, abs, signed, runs, runLen, matched, offered float64 }

func (r res) meanAbs() float64 {
	if r.n == 0 {
		return 0
	}
	return r.abs / r.n
}
func (r res) meanSigned() float64 {
	if r.n == 0 {
		return 0
	}
	return r.signed / r.n
}
func (r res) rate() float64 {
	if r.offered == 0 {
		return 0
	}
	return r.matched / r.offered
}
func (r res) meanRunLen() float64 {
	if r.runs == 0 {
		return 0
	}
	return r.runLen / r.runs
}

// trackRows is notes/s28-drift's matcher, reimplemented so this file does not
// depend on it: every run in frame f is matched to the nearest run in f+gap
// whose length is within a factor of two, and the signed displacement of the
// centres is recorded. `pair` chooses which frames are compared, so the same
// matcher can be pointed at adjacent frames or at independent ones.
func trackRows(ms []mask, y0, y1, gap, maxShift int) res {
	var r res
	for f := 0; f+gap < len(ms); f++ {
		a, b := ms[f], ms[f+gap]
		for y := y0; y < y1; y++ {
			ra := runsIn(a[y], 0, W)
			rb := runsIn(b[y], 0, W)
			for _, q := range ra {
				r.runs++
				r.runLen += float64(q.length())
			}
			if len(ra) == 0 || len(rb) == 0 {
				continue
			}
			for _, q := range ra {
				r.offered++
				best, bd := -1, math.Inf(1)
				for j, p := range rb {
					d := math.Abs(p.centre() - q.centre())
					if d > float64(maxShift) {
						continue
					}
					if p.length() > 2*q.length() || q.length() > 2*p.length() {
						continue
					}
					if d < bd {
						bd, best = d, j
					}
				}
				if best < 0 {
					continue
				}
				r.matched++
				d := rb[best].centre() - q.centre()
				r.n++
				r.abs += math.Abs(d)
				r.signed += d
			}
		}
	}
	return r
}

// trackRowsPaired is the same matcher on an explicit list of frame pairs, used
// for the decorrelation pedestal, where the two frames are far apart in time.
func trackRowsPaired(ms []mask, pairs [][2]int, y0, y1, maxShift int) res {
	var r res
	for _, pr := range pairs {
		a, b := ms[pr[0]], ms[pr[1]]
		for y := y0; y < y1; y++ {
			ra := runsIn(a[y], 0, W)
			rb := runsIn(b[y], 0, W)
			for _, q := range ra {
				r.runs++
				r.runLen += float64(q.length())
			}
			if len(ra) == 0 || len(rb) == 0 {
				continue
			}
			for _, q := range ra {
				r.offered++
				best, bd := -1, math.Inf(1)
				for j, p := range rb {
					d := math.Abs(p.centre() - q.centre())
					if d > float64(maxShift) {
						continue
					}
					if p.length() > 2*q.length() || q.length() > 2*p.length() {
						continue
					}
					if d < bd {
						bd, best = d, j
					}
				}
				if best < 0 {
					continue
				}
				r.matched++
				d := rb[best].centre() - q.centre()
				r.n++
				r.abs += math.Abs(d)
				r.signed += d
			}
		}
	}
	return r
}

func trackCols(ms []mask, y0, y1, gap, maxShift int) res {
	var r res
	col := make([]bool, H)
	col2 := make([]bool, H)
	for f := 0; f+gap < len(ms); f++ {
		a, b := ms[f], ms[f+gap]
		for x := 0; x < W; x++ {
			for y := 0; y < H; y++ {
				col[y], col2[y] = false, false
			}
			for y := y0; y < y1; y++ {
				col[y] = a[y][x]
				col2[y] = b[y][x]
			}
			ra := runsIn(col, y0, y1)
			rb := runsIn(col2, y0, y1)
			for _, q := range ra {
				r.runs++
				r.runLen += float64(q.length())
			}
			if len(ra) == 0 || len(rb) == 0 {
				continue
			}
			for _, q := range ra {
				r.offered++
				best, bd := -1, math.Inf(1)
				for j, p := range rb {
					d := math.Abs(p.centre() - q.centre())
					if d > float64(maxShift) {
						continue
					}
					if p.length() > 2*q.length() || q.length() > 2*p.length() {
						continue
					}
					if d < bd {
						bd, best = d, j
					}
				}
				if best < 0 {
					continue
				}
				r.matched++
				d := rb[best].centre() - q.centre()
				r.n++
				r.abs += math.Abs(d)
				r.signed += d
			}
		}
	}
	return r
}

func count(ms []mask, y0, y1 int) float64 {
	n := 0.0
	for _, m := range ms {
		for y := y0; y < y1; y++ {
			for x := 0; x < W; x++ {
				if m[y][x] {
					n++
				}
			}
		}
	}
	return n
}

// -------------------------------------------- notes/drift, replayed YOUNG

// youngReplay runs notes/drift's exact procedure `trials` times, each on a
// FRESH shore warmed by a different number of frames so the phase differs but
// the wave age stays where notes/drift's own measurement sits.
func youngReplay(trials int) (peaks []int, profile []float64, maxAge float64) {
	profile = make([]float64, 9)
	for tr := 0; tr < trials; tr++ {
		sh := scape.NewShore(seed, false)
		t := 0.0
		adv := func(n int) *canvas.Canvas {
			var c *canvas.Canvas
			for i := 0; i < n; i++ {
				c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				t += stepDT
				sh.Update(c, t, act)
			}
			return c
		}
		grab := func(c *canvas.Canvas) [][]bool {
			g := make([][]bool, H)
			for y := range g {
				g[y] = make([]bool, W)
				for x := 0; x < W; x++ {
					r, _, _ := c.ResolveAt(x, y, term.Profile256)
					g[y][x] = driftSet(r)
				}
			}
			return g
		}
		adv(tr) // 0..trials-1 frames of warm-up: a different phase, same youth
		a := grab(adv(40))
		b := grab(adv(18))
		age := float64(tr+58) * stepDT * (0.55 + act.Level*1.45)
		if age > maxAge {
			maxAge = age
		}
		score := func(dx int) int {
			n := 0
			for y := 2; y < H-2; y++ {
				for x := 2; x < W-2; x++ {
					xx := x + dx
					if xx < 0 || xx >= W {
						continue
					}
					if a[y][x] && b[y][xx] {
						n++
					}
				}
			}
			return n
		}
		best, bs := 0, -1
		for dx := -4; dx <= 4; dx++ {
			s := score(dx)
			profile[dx+4] += float64(s)
			if s > bs {
				bs, best = s, dx
			}
		}
		peaks = append(peaks, best)
	}
	for i := range profile {
		profile[i] /= float64(trials)
	}
	return
}

func histLine(peaks []int) (string, float64) {
	h := map[int]int{}
	sum := 0.0
	for _, p := range peaks {
		h[p]++
		sum += float64(p)
	}
	s := ""
	for dx := -4; dx <= 4; dx++ {
		s += fmt.Sprintf("%+d:%d ", dx, h[dx])
	}
	return s, sum / float64(len(peaks))
}

func profLine(p []float64) (string, float64, float64) {
	s := ""
	mx, mn := 0.0, math.Inf(1)
	for dx := -4; dx <= 4; dx++ {
		v := p[dx+4]
		s += fmt.Sprintf("%+d:%.1f ", dx, v)
		if v > mx {
			mx = v
		}
		if v < mn {
			mn = v
		}
	}
	if mx == 0 {
		return s, 0, 0
	}
	return s, (mx - mn) / mx * 100, (mx - p[4]) / p[4] * 100
}

// ------------------------------------------------------------------- report

func main() {
	out := os.Stdout
	hy := hyRow(H)
	y0, y1 := hy+1, H-6
	fmt.Fprintf(out, "ADVERSARIAL RE-CHECK of notes/s28-drift\n")
	fmt.Fprintf(out, "scape.Tide=%v  term.NoSplitCells=%v  term.Shading=%v  %dx%d  level %.2f  step %.2fs  rows %d..%d\n\n",
		scape.Tide, term.NoSplitCells, term.Shading, W, H, act.Level, stepDT, y0, y1-1)

	// ------------------------------------------------------------------ A
	fmt.Fprintln(out, "== A. IS 6.44 cells/s A RATE, OR THE MATCHER'S PEDESTAL? ==")
	fmt.Fprintln(out, "   A true translation is LINEAR in the gap: cells/s is the same at every gap.")
	fmt.Fprintln(out, "   A nearest-run matcher on a decorrelating field saturates immediately.")
	ms := shoreRun(400, 500, driftSet)
	g := count(ms, y0, y1)
	if g == 0 {
		fmt.Fprintln(out, "   EMPTY SAMPLE -- nothing drawn in the band. RESULT VOID.")
		return
	}
	fmt.Fprintf(out, "   sample: %.0f sea glyphs over %d frames (non-empty by that count)\n", g, len(ms))
	fmt.Fprintf(out, "   %-6s %-9s %-10s %-10s %-9s %s\n", "gap", "seconds", "mean|dx|", "cells/s", "signed", "match%")
	for _, gp := range []int{1, 2, 3, 4, 6, 8, 12, 16, 24} {
		r := trackRows(ms, y0, y1, gp, 6)
		fmt.Fprintf(out, "   %-6d %-9.2f %-10.3f %-10.2f %-+9.3f %.2f\n",
			gp, float64(gp)*stepDT, r.meanAbs(), r.meanAbs()/(float64(gp)*stepDT), r.meanSigned(), r.rate())
	}
	fmt.Fprintln(out, "   ...and with the search window widened with the gap, so a real slide is not clipped:")
	fmt.Fprintf(out, "   %-6s %-9s %-10s %-10s %-9s %s\n", "gap", "maxShift", "mean|dx|", "cells/s", "signed", "match%")
	for _, gp := range []int{1, 2, 3, 4, 6, 8, 12, 16, 24} {
		msh := 4 + 2*gp
		if msh > 30 {
			msh = 30
		}
		r := trackRows(ms, y0, y1, gp, msh)
		fmt.Fprintf(out, "   %-6d %-9d %-10.3f %-10.2f %-+9.3f %.2f\n",
			gp, msh, r.meanAbs(), r.meanAbs()/(float64(gp)*stepDT), r.meanSigned(), r.rate())
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "   THE PEDESTAL: the same matcher on frames far enough apart to be independent.")
	fmt.Fprintln(out, "   These frames have no relationship at all, so whatever it reports here is")
	fmt.Fprintln(out, "   matching noise and nothing else. notes/s28-drift's only null was a REPEATED")
	fmt.Fprintln(out, "   frame, which is 0.000 by construction and tests nothing.")
	for _, sep := range []int{97, 199, 307, 401} {
		var pairs [][2]int
		for i := 0; i+sep < len(ms); i++ {
			pairs = append(pairs, [2]int{i, i + sep})
		}
		if len(pairs) == 0 {
			fmt.Fprintf(out, "   separation %-4d NO PAIRS -- VOID\n", sep)
			continue
		}
		r := trackRowsPaired(ms, pairs, y0, y1, 6)
		fmt.Fprintf(out, "   separation %-4d (%.1fs apart, %d pairs)  mean|dx| %.3f  signed %+.3f  match%% %.2f\n",
			sep, float64(sep)*stepDT, len(pairs), r.meanAbs(), r.meanSigned(), r.rate())
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ B
	fmt.Fprintln(out, "== B. IS 3.00 : 1 A RATIO? CELLS ARE NOT ROWS ==")
	{
		fx := trackRows(ms, y0, y1, 3, 6)
		fy := trackCols(ms, y0, y1, 3, 6)
		fmt.Fprintf(out, "   horizontal: mean|dx| %.3f cells, mean run %.1f cells long, match%% %.2f\n",
			fx.meanAbs(), fx.meanRunLen(), fx.rate())
		fmt.Fprintf(out, "   vertical:   mean|dy| %.3f rows,  mean run %.1f rows long,  match%% %.2f  (band is %d rows tall)\n",
			fy.meanAbs(), fy.meanRunLen(), fy.rate(), y1-y0)
		fmt.Fprintf(out, "   their ratio, cells divided by rows:            %.2f : 1\n",
			fx.meanAbs()/math.Max(1e-9, fy.meanAbs()))
		fmt.Fprintf(out, "   the same ratio in SCREEN DISTANCE (cell is %.3f x taller than wide): %.2f : 1\n",
			cellAspect, fx.meanAbs()/math.Max(1e-9, fy.meanAbs()*cellAspect))
		fmt.Fprintln(out, "   ^ Menlo, from the font file: advance 0.6021 em, hhea line 1.1641 em.")
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ C
	fmt.Fprintln(out, "== C. notes/drift REPLAYED 200 TIMES, EVERY TRIAL YOUNG ==")
	fmt.Fprintln(out, "   notes/s28-drift's replay advances ONE shore through all 200 trials, so its")
	fmt.Fprintln(out, "   trials span wave age 0 to ~1300 while notes/drift measures at 6.6 -- and its")
	fmt.Fprintln(out, "   own section 9b says the x-structure changes with age. Here every trial is a")
	fmt.Fprintln(out, "   fresh shore with a different phase and the SAME youth.")
	{
		peaks, prof, maxAge := youngReplay(200)
		h, mean := histLine(peaks)
		p, spread, lift := profLine(prof)
		fmt.Fprintf(out, "   wave age spans 6.6 .. %.1f (notes/drift measures at 6.6)\n", maxAge)
		fmt.Fprintf(out, "   peak dx histogram: %s\n   mean peak dx %+.2f over 200 trials\n", h, mean)
		fmt.Fprintf(out, "   mean profile:      %s\n   spread (max-min)/max %.1f%%, lift over dx=0 %.1f%%\n",
			p, spread, lift)
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ D
	fmt.Fprintln(out, "== D. THE BACKGROUND CHANNEL NEITHER INSTRUMENT READ ==")
	fmt.Fprintln(out, "   shore.go's own comment: the wave motion 'lives in the glyphs and in the")
	fmt.Fprintln(out, "   waterline cell below, which keeps its own per-column edge'. Both drift")
	fmt.Fprintln(out, "   instruments count only three GLYPHS. This marks every cell whose glyph +fg")
	fmt.Fprintln(out, "   +bg differs from its row's modal appearance, which is the waterline's own")
	fmt.Fprintln(out, "   per-column edge and the wet band, and runs the same tracker over it.")
	{
		bms, tot := bgRun(400, 200)
		if tot == 0 {
			fmt.Fprintln(out, "   EMPTY SAMPLE -- no background variation found at all. RESULT VOID.")
		} else {
			by0, by1 := hy+1, H
			n := count(bms, by0, by1)
			fmt.Fprintf(out, "   sample: %.0f marked background cells over %d frames, rows %d..%d (non-empty by that count)\n",
				n, len(bms), by0, by1-1)
			for _, gp := range []int{1, 3, 6, 12} {
				r := trackRows(bms, by0, by1, gp, 6)
				fmt.Fprintf(out, "   gap %-3d (%.2fs)  mean|dx| %.3f = %.2f cells/s   signed %+.3f   match%% %.2f  mean run %.1f\n",
					gp, float64(gp)*stepDT, r.meanAbs(), r.meanAbs()/(float64(gp)*stepDT),
					r.meanSigned(), r.rate(), r.meanRunLen())
			}
			var pairs [][2]int
			for i := 0; i+151 < len(bms); i++ {
				pairs = append(pairs, [2]int{i, i + 151})
			}
			if len(pairs) > 0 {
				r := trackRowsPaired(bms, pairs, by0, by1, 6)
				fmt.Fprintf(out, "   pedestal (151 frames apart, %d pairs)  mean|dx| %.3f   signed %+.3f\n",
					len(pairs), r.meanAbs(), r.meanSigned())
			}
		}
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ E
	fmt.Fprintln(out, "== E. THINGS THEY DID NOT VARY ==")
	fmt.Fprintf(out, "   %-42s %-9s %-9s %-9s %s\n", "variant", "glyphs", "mean|dx|", "cells/s", "signed")
	report := func(name string) {
		mm := shoreRun(400, 300, driftSet)
		n := count(mm, y0, y1)
		if n == 0 {
			fmt.Fprintf(out, "   %-42s EMPTY SAMPLE -- VOID\n", name)
			return
		}
		r := trackRows(mm, y0, y1, 3, 6)
		fmt.Fprintf(out, "   %-42s %-9.0f %-9.3f %-9.2f %+.3f\n",
			name, n, r.meanAbs(), r.meanAbs()/(3*stepDT), r.meanSigned())
	}
	report("baseline (their settings)")

	// term.NoSplitCells: production on Terminal.app is TRUE and both drift
	// instruments ran FALSE. Set it, read, and RESTORE -- an instrument that
	// leaves a global flipped reports what it last set.
	wasSplit := term.NoSplitCells
	term.NoSplitCells = true
	report("term.NoSplitCells=true (his Terminal.app)")
	term.NoSplitCells = wasSplit
	fmt.Fprintf(out, "   (restored term.NoSplitCells=%v)\n", term.NoSplitCells)

	savedTOD := act.TimeOfDay
	for _, tod := range []float64{1.0 / 24, 6.0 / 24, 21.0 / 24} {
		act.TimeOfDay = tod
		report(fmt.Sprintf("time of day %04.1fh (they only ran 13.0h)", tod*24))
	}
	act.TimeOfDay = savedTOD

	savedCtx := act.ContextUsed
	for _, cu := range []float64{0.0, 0.95} {
		act.ContextUsed = cu
		report(fmt.Sprintf("context used %.2f (they only ran 0.30)", cu))
	}
	act.ContextUsed = savedCtx

	{
		mm := shoreRun(400, 300, driftSet)
		n := count(mm, y0, y1)
		sh2 := scape.NewShore(42, false)
		t := 0.0
		var m2 []mask
		for i := 0; i < 700; i++ {
			c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			t += stepDT
			sh2.Update(c, t, act)
			if i < 400 {
				continue
			}
			mk := make(mask, H)
			for y := range mk {
				mk[y] = make([]bool, W)
				for x := 0; x < W; x++ {
					r, _, _ := c.ResolveAt(x, y, term.Profile256)
					mk[y][x] = driftSet(r)
				}
			}
			m2 = append(m2, mk)
		}
		n2 := count(m2, y0, y1)
		diff := 0
		for i := range m2 {
			for y := y0; y < y1; y++ {
				for x := 0; x < W; x++ {
					if m2[i][y][x] != mm[i][y][x] {
						diff++
					}
				}
			}
		}
		r2 := trackRows(m2, y0, y1, 3, 6)
		fmt.Fprintf(out, "   %-42s %-9.0f %-9.3f %-9.2f %+.3f\n",
			"seed 42 instead of 7", n2, r2.meanAbs(), r2.meanAbs()/(3*stepDT), r2.meanSigned())
		fmt.Fprintf(out, "   (seed 7 vs seed 42: %.0f vs %.0f glyphs, %d differing cells -- the sea's\n", n, n2, diff)
		fmt.Fprintln(out, "    glyph set has no hash in it, so the seed cannot move this measurement.)")
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ G
	fmt.Fprintln(out, "== G. THE CONTROL SECTION A NEEDED: THE PEDESTAL AT THE SAME WINDOW ==")
	fmt.Fprintln(out, "   Widening the search window with the gap also widens the room for a WRONG")
	fmt.Fprintln(out, "   match, so section A's flat cells/s could be the window growing rather than")
	fmt.Fprintln(out, "   the pattern travelling. This runs the identical matcher, at the identical")
	fmt.Fprintln(out, "   window, on frames 199 apart -- which have no relationship at all.")
	fmt.Fprintf(out, "   %-6s %-9s %-11s %-11s %-9s %s\n", "gap", "maxShift", "real|dx|", "pedestal|dx|", "ratio", "match% real/ped")
	{
		var pairs [][2]int
		for i := 0; i+199 < len(ms); i++ {
			pairs = append(pairs, [2]int{i, i + 199})
		}
		for _, gp := range []int{1, 2, 3, 4, 6, 8, 12} {
			msh := 4 + 2*gp
			if msh > 30 {
				msh = 30
			}
			r := trackRows(ms, y0, y1, gp, msh)
			p := trackRowsPaired(ms, pairs, y0, y1, msh)
			fmt.Fprintf(out, "   %-6d %-9d %-11.3f %-11.3f %-9.2f %.2f / %.2f\n",
				gp, msh, r.meanAbs(), p.meanAbs(), r.meanAbs()/math.Max(1e-9, p.meanAbs()),
				r.rate(), p.rate())
		}
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ H
	fmt.Fprintln(out, "== H. THE INWARD SPEED AT THE SAME UNCLIPPED GAP ==")
	fmt.Fprintf(out, "   %-6s %-11s %-11s %-11s %-11s %s\n", "gap", "|dx| cells/s", "|dy| rows/s", "cells:rows", "SCREEN ratio", "match x/y")
	for _, gp := range []int{1, 2, 3} {
		fx := trackRows(ms, y0, y1, gp, 6)
		fy := trackCols(ms, y0, y1, gp, 6)
		dt := float64(gp) * stepDT
		fmt.Fprintf(out, "   %-6d %-11.2f %-11.2f %-11.2f %-11.2f %.2f / %.2f\n",
			gp, fx.meanAbs()/dt, fy.meanAbs()/dt,
			fx.meanAbs()/math.Max(1e-9, fy.meanAbs()),
			fx.meanAbs()/math.Max(1e-9, fy.meanAbs()*cellAspect), fx.rate(), fy.rate())
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ I
	fmt.Fprintln(out, "== I. IS THE YOUNG-SCAPE LEFTWARD LEAN SIGNIFICANT? A SIGN TEST ==")
	fmt.Fprintln(out, "   For each of 200 young trials, score(-4)+score(-3) minus score(+3)+score(+4).")
	fmt.Fprintln(out, "   Under 'no drift' that is positive in half the trials. Reported with the")
	fmt.Fprintln(out, "   same test on 200 trials of a REPEATED frame, where the answer must be 0.")
	{
		lefts, n := 0, 200
		var sum float64
		for tr := 0; tr < n; tr++ {
			sh := scape.NewShore(seed, false)
			t := 0.0
			adv := func(k int) *canvas.Canvas {
				var c *canvas.Canvas
				for i := 0; i < k; i++ {
					c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					t += stepDT
					sh.Update(c, t, act)
				}
				return c
			}
			grab := func(c *canvas.Canvas) [][]bool {
				g := make([][]bool, H)
				for y := range g {
					g[y] = make([]bool, W)
					for x := 0; x < W; x++ {
						r, _, _ := c.ResolveAt(x, y, term.Profile256)
						g[y][x] = driftSet(r)
					}
				}
				return g
			}
			adv(tr)
			a := grab(adv(40))
			b := grab(adv(18))
			score := func(dx int) float64 {
				k := 0.0
				for y := 2; y < H-2; y++ {
					for x := 2; x < W-2; x++ {
						xx := x + dx
						if xx < 0 || xx >= W {
							continue
						}
						if a[y][x] && b[y][xx] {
							k++
						}
					}
				}
				return k
			}
			s := score(-4) + score(-3) - score(3) - score(4)
			sum += s
			if s > 0 {
				lefts++
			}
		}
		fmt.Fprintf(out, "   leftward in %d of %d trials (chance is 100); mean left-minus-right %+.1f cells matched\n",
			lefts, n, sum/float64(n))
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ K
	fmt.Fprintln(out, "== K. IS IT WAVE AGE THAT DECIDES? THE SAME SIGN TEST, BY AGE ==")
	fmt.Fprintln(out, "   notes/s28-drift's replay runs ONE shore through all 200 trials, so its")
	fmt.Fprintln(out, "   trials age from 0 to ~1300 while notes/drift measures at 6.6. This runs")
	fmt.Fprintln(out, "   400 trials the same continuous way and buckets the sign statistic by the")
	fmt.Fprintln(out, "   wave age each trial was taken at. If the lean is there when young and gone")
	fmt.Fprintln(out, "   when old, their averaged profile is answering a different question.")
	{
		type bucket struct {
			lo, hi     float64
			n, lefts   int
			sum, prof0 float64
			prof       [9]float64
		}
		buckets := []*bucket{
			{lo: 0, hi: 30}, {lo: 30, hi: 100}, {lo: 100, hi: 300},
			{lo: 300, hi: 1000}, {lo: 1000, hi: 1e9},
		}
		sh := scape.NewShore(seed, false)
		t := 0.0
		adv := func(k int) *canvas.Canvas {
			var c *canvas.Canvas
			for i := 0; i < k; i++ {
				c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				t += stepDT
				sh.Update(c, t, act)
			}
			return c
		}
		grab := func(c *canvas.Canvas) [][]bool {
			g := make([][]bool, H)
			for y := range g {
				g[y] = make([]bool, W)
				for x := 0; x < W; x++ {
					r, _, _ := c.ResolveAt(x, y, term.Profile256)
					g[y][x] = driftSet(r)
				}
			}
			return g
		}
		for tr := 0; tr < 400; tr++ {
			a := grab(adv(40))
			b := grab(adv(18))
			age := t * (0.55 + act.Level*1.45)
			score := func(dx int) float64 {
				k := 0.0
				for y := 2; y < H-2; y++ {
					for x := 2; x < W-2; x++ {
						xx := x + dx
						if xx < 0 || xx >= W {
							continue
						}
						if a[y][x] && b[y][xx] {
							k++
						}
					}
				}
				return k
			}
			var p [9]float64
			for dx := -4; dx <= 4; dx++ {
				p[dx+4] = score(dx)
			}
			s := p[0] + p[1] - p[7] - p[8]
			for _, bk := range buckets {
				if age >= bk.lo && age < bk.hi {
					bk.n++
					bk.sum += s
					if s > 0 {
						bk.lefts++
					}
					for i := 0; i < 9; i++ {
						bk.prof[i] += p[i]
					}
					bk.prof0 += p[4]
				}
			}
		}
		fmt.Fprintf(out, "   %-16s %-7s %-14s %-16s %s\n", "wave age", "trials", "leftward", "mean L-minus-R", "profile max at dx")
		for _, bk := range buckets {
			if bk.n == 0 {
				fmt.Fprintf(out, "   %-16s NO TRIALS -- nothing to report\n",
					fmt.Sprintf("%.0f..%.0f", bk.lo, bk.hi))
				continue
			}
			best, bv := -4, -1.0
			for i := 0; i < 9; i++ {
				if bk.prof[i] > bv {
					bv, best = bk.prof[i], i-4
				}
			}
			hi := fmt.Sprintf("%.0f", bk.hi)
			if bk.hi > 1e8 {
				hi = "inf"
			}
			fmt.Fprintf(out, "   %-16s %-7d %-14s %-16.1f %+d\n",
				fmt.Sprintf("%.0f..%s", bk.lo, hi), bk.n,
				fmt.Sprintf("%d of %d", bk.lefts, bk.n), bk.sum/float64(bk.n), best)
		}
		fmt.Fprintln(out, "   (chance is half the trials leftward and a mean of 0.0)")
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ J
	fmt.Fprintln(out, "== J. THE BACKGROUND CHANNEL AT HIS TERMINAL'S FLAGS ==")
	{
		wasN := term.NoSplitCells
		term.NoSplitCells = true
		bms, tot := bgRun(400, 150)
		term.NoSplitCells = wasN
		if tot == 0 {
			fmt.Fprintln(out, "   EMPTY SAMPLE -- VOID")
		} else {
			by0, by1 := hy+1, H
			fmt.Fprintf(out, "   NoSplitCells=true: %.0f marked background cells over %d frames\n", count(bms, by0, by1), len(bms))
			for _, gp := range []int{1, 3, 6} {
				r := trackRows(bms, by0, by1, gp, 6)
				fmt.Fprintf(out, "   gap %-3d  mean|dx| %.3f = %.2f cells/s   signed %+.3f   match%% %.2f\n",
					gp, r.meanAbs(), r.meanAbs()/(float64(gp)*stepDT), r.meanSigned(), r.rate())
			}
		}
		fmt.Fprintf(out, "   (restored term.NoSplitCells=%v)\n", term.NoSplitCells)
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ L
	fmt.Fprintln(out, "== L. SIGNED DRIFT AGAINST WAVE AGE, ON THE TRACKER ==")
	fmt.Fprintln(out, "   Section K says the correlation profile leans LEFT while the scape is young")
	fmt.Fprintln(out, "   and stops leaning once it is old. This is the independent check: the run")
	fmt.Fprintln(out, "   tracker's SIGNED mean at the same ages. Two metrics that agree settle it;")
	fmt.Fprintln(out, "   two that disagree mean neither should be quoted as a direction.")
	fmt.Fprintf(out, "   %-9s %-10s %-9s %-14s %-14s %s\n", "warm-up", "wave age", "glyphs", "signed gap1", "signed gap3", "|dx| gap1")
	for _, wf := range []int{0, 25, 100, 400, 1600, 6400} {
		mm := shoreRun(wf, 200, driftSet)
		n := count(mm, y0, y1)
		if n == 0 {
			fmt.Fprintf(out, "   %-9d EMPTY SAMPLE -- VOID\n", wf)
			continue
		}
		r1 := trackRows(mm, y0, y1, 1, 6)
		r3 := trackRows(mm, y0, y1, 3, 6)
		fmt.Fprintf(out, "   %-9d %-10.0f %-9.0f %-+14.4f %-+14.4f %.3f\n",
			wf, float64(wf)*stepDT*(0.55+act.Level*1.45), n, r1.meanSigned(), r3.meanSigned(), r1.meanAbs())
	}
	fmt.Fprintln(out)

	// ------------------------------------------------------------------ F
	fmt.Fprintln(out, "== F. WHAT A COLUMN OF THE SEA ACTUALLY DOES OVER ONE SECOND ==")
	fmt.Fprintln(out, "   The direct observable, printed beside the verdict: one row, gap 3 apart,")
	fmt.Fprintln(out, "   so a 1.55-cell mean shift should be visible as marks moving one or two")
	fmt.Fprintln(out, "   cells sideways between consecutive lines.")
	{
		row := (y0 + y1) / 2
		for i := 0; i < 15; i += 3 {
			b := make([]byte, 0, 80)
			for x := 24; x < 100; x++ {
				if ms[i][row][x] {
					b = append(b, '#')
				} else {
					b = append(b, '.')
				}
			}
			fmt.Fprintf(out, "   t=%4.2fs row %d  %s\n", float64(i)*stepDT, row, string(b))
		}
	}
}
