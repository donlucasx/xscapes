// Command s28-seamap-verify is an ADVERSARIAL re-check of notes/s28-seamap.
//
// It exists to attack three things that instrument could not settle on its own:
//
//  1. THE CAUSE. s28-seamap isolates the writing-band guard by varying
//     Shore.SandRows. That moves the guard's room AND the sea's depth AND
//     hyFloor together -- three terms, one knob. Shore.WriteRows is the knob
//     that turns the guard OFF: `wr < 2` sets wr=0, which sets writeTop=c.H,
//     which makes `if writeTop < c.H` false and the whole rescale never runs.
//     Pinning SandRows to the shipped beach at the same time holds sy, hy and
//     scale byte-identical, so GUARD ON vs GUARD OFF is a one-term variation.
//     153x51 is the negative control: the guard is claimed not to fire there,
//     so turning it off must change nothing.
//
//  2. TIME OF DAY. Every reading in s28-seamap is taken at 13:00. The whole
//     waterline channel is read with `bg.R > bg.B`, which is a claim about the
//     midday palette. If the sea's mapping or the reader's discriminator moves
//     with the clock, the headline is a midday result quoted as a general one.
//
//  3. THE SEED. Every reading is scape.NewShore(7). One draw of the sea.
//
// Modes:
//
//	go run ./notes/s28-seamap-verify            # 1: guard on vs guard off
//	go run ./notes/s28-seamap-verify -tod       # 2: six hours, 111x25
//	go run ./notes/s28-seamap-verify -seed      # 3: six seeds, 111x25
//	go run ./notes/s28-seamap-verify -height    # the guard's room vs window height
//	go run ./notes/s28-seamap-verify -probe40   # what "warm" is at 40x12, level 0
//
// XSCAPES_TIDE=0 works on all of them.
package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

const dt = 0.08

func waveGlyph(r rune) bool {
	switch r {
	case '·', '-', '~', '≈', '≋':
		return true
	}
	return false
}

type cfg struct {
	w, h      int
	lv        float64
	seed      int64
	tod       float64 // hours/24
	writeRows int     // -1 = leave the shipped default
	sandRows  int
}

type res struct {
	topMean, topSD float64 // waterline mean row; sd is ACROSS FRAMES of the frame mean
	swing, crests  float64
	wave, whitecap float64
	frames         int
	seaCells       int // non-empty proof
	clipped        int // columns whose reading hit the scan floor hy+1
	scanned        int
	bandContact    int // columns whose waterline reading reached writeTop-1
	bandInside     int // columns whose waterline landed INSIDE the writing band (the ruled-shoreline cost)
	runMax         int // longest run of columns sharing one waterline row -- a ruled shoreline IS a long run
	maxEdge        float64
}

func run(c cfg) res {
	sh := scape.NewShore(c.seed, false)
	if c.writeRows >= 0 {
		sh.WriteRows = c.writeRows
	}
	if c.sandRows > 0 {
		sh.SandRows = c.sandRows
	}
	act := scape.Activity{Level: c.lv, Working: true, ContextUsed: 0.3, TimeOfDay: c.tod}
	hy := int(float64(c.h) * 0.42)

	// Same rule as s28-seamap: at least two whole wash cycles, never under 220.
	n := int(math.Ceil(2 * (2 * math.Pi / 0.5) / (dt * (0.55 + c.lv*1.45))))
	if n < 220 {
		n = 220
	}

	tm := 0.0
	for k := 0; k < 500; k++ { // warm-up. A fresh Shore per frame has no history.
		tm += dt
		sh.Update(canvas.New(c.w, c.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm, act)
	}

	// The writing band's top row, recomputed the way Update does, so the
	// band-contact cost can be read off the RENDERED waterline instead of a model.
	wr := scape.DefaultWriteRows
	if m := c.h / 6; wr > m {
		wr = m
	}
	writeTop := c.h
	if wr >= 2 && c.h > wr+4 {
		writeTop = c.h - wr
	}

	out := res{frames: n}
	var means []float64
	for k := 0; k < n; k++ {
		cv := canvas.New(c.w, c.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += dt
		sh.Update(cv, tm, act)

		prof := make([]float64, c.w)
		for x := 0; x < c.w; x++ {
			prof[x] = float64(c.h)
			for y := hy + 1; y < c.h; y++ {
				if _, _, bg := cv.ResolveAt(x, y, term.Profile256); bg.R > bg.B {
					prof[x] = float64(y)
					break
				}
			}
			out.scanned++
			if prof[x] <= float64(hy)+1 {
				out.clipped++
			}
			if prof[x] >= float64(writeTop)-1 {
				out.bandContact++
			}
			if prof[x] >= float64(writeTop) {
				out.bandInside++
			}
			if prof[x] > out.maxEdge {
				out.maxEdge = prof[x]
			}
		}

		wv, wc := 0, 0
		for x := 0; x < c.w; x++ {
			for y := hy + 1; float64(y) <= prof[x]-2; y++ {
				r, _, _ := cv.ResolveAt(x, y, term.Profile256)
				if !waveGlyph(r) {
					continue
				}
				wv++
				if r == '≋' {
					wc++
				}
			}
		}
		out.seaCells += wv
		out.wave += float64(wv)
		out.whitecap += float64(wc)

		// The ruled shoreline the guard exists to prevent is literally a long
		// run of columns on one row. Measure the run, not a proxy for it.
		runLen := 1
		for x := 1; x < c.w; x++ {
			if prof[x] == prof[x-1] {
				runLen++
			} else {
				runLen = 1
			}
			if runLen > out.runMax {
				out.runMax = runLen
			}
		}

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
		out.swing += hi - lo
		means = append(means, sum/float64(c.w))
		pk := 0
		for x := 2; x < c.w-2; x++ {
			if prof[x] < prof[x-2] && prof[x] <= prof[x+2] {
				pk++
			}
		}
		out.crests += float64(pk)
	}
	out.wave /= float64(n)
	out.whitecap /= float64(n)
	out.swing /= float64(n)
	out.crests /= float64(n)
	m, s := 0.0, 0.0
	for _, v := range means {
		m += v
	}
	m /= float64(len(means))
	for _, v := range means {
		s += (v - m) * (v - m)
	}
	out.topMean, out.topSD = m, math.Sqrt(s/float64(len(means)-1))
	return out
}

func main() {
	tod := flag.Bool("tod", false, "vary the time of day")
	seed := flag.Bool("seed", false, "vary the shore seed")
	height := flag.Bool("height", false, "vary the window height")
	swingF := flag.Bool("swing", false, "does the guard also squash the coast's spatial swing?")
	probe40 := flag.Bool("probe40", false, "dump a 40x12 column at level 0")
	flag.Parse()
	fmt.Printf("XSCAPES TIDE = %v  (TideRange %.1f, TideEase %.1f)\n\n", scape.Tide, scape.TideRange, scape.TideEase)
	switch {
	case *tod:
		varyTOD()
	case *seed:
		varySeed()
	case *swingF:
		varySwing()
	case *height:
		varyHeight()
	case *probe40:
		probe()
	default:
		guardIsolation()
	}
}

// ---------------------------------------------------------------------------
// 1. THE CAUSE, isolated on ONE term.
func guardIsolation() {
	fmt.Println("GUARD ON vs GUARD OFF at IDENTICAL geometry.")
	fmt.Println("SandRows is pinned to the shipped beach in every row, so sy, hy and scale never move.")
	fmt.Println("WriteRows 4 -> 0 makes `wr<2 => wr=0 => writeTop=c.H`, so `if writeTop < c.H` is false")
	fmt.Println("and the rescale in shore.go:330-357 never executes. That is the only difference.")
	fmt.Println()
	type g struct {
		name string
		w, h int
	}
	for _, gm := range []g{{"80x24  target", 80, 24}, {"111x25   his", 111, 25}, {"153x51 CONTROL", 153, 51}} {
		wr := scape.DefaultWriteRows
		if m := gm.h / 6; wr > m {
			wr = m
		}
		beach := gm.h / 5
		if min := wr + 3; beach < min {
			beach = min
		}
		if beach < 4 {
			beach = 4
		}
		room := beach - wr - 1
		fmt.Printf("=== %s : beach %d, band %d, guard room at full activity %d rows ===\n", gm.name, beach, wr, room)
		fmt.Printf("  %-26s %7s %7s %7s %7s %7s %8s %8s %8s\n",
			"config", "0.60", "0.70", "0.80", "0.90", "1.00", "wash@.7", "wash@1", "d(wash)")
		type cf struct {
			label string
			wrows int
			srows int
		}
		for _, c := range []cf{
			{"SHIPPED (untouched)", -1, 0},
			{"CONTROL SandRows pinned", 4, beach},
			{"GUARD OFF WriteRows=0", 0, beach},
		} {
			var tops, washes []float64
			var contact, inside, scanned, runmax int
			for _, lv := range []float64{0.60, 0.70, 0.80, 0.90, 1.00} {
				r := run(cfg{w: gm.w, h: gm.h, lv: lv, seed: 7, tod: 13.0 / 24, writeRows: c.wrows, sandRows: c.srows})
				tops = append(tops, r.topMean)
				washes = append(washes, r.topSD)
				if lv == 1.00 {
					contact, inside, scanned, runmax = r.bandContact, r.bandInside, r.scanned, r.runMax
				}
			}
			fmt.Printf("  %-26s %7.2f %7.2f %7.2f %7.2f %7.2f %8.2f %8.2f %8.2f | @1.0 reach band-top %5.2f%%  INSIDE band %5.2f%%  longest one-row run %d of %d cols\n",
				c.label, tops[0], tops[1], tops[2], tops[3], tops[4],
				washes[1], washes[4], washes[4]-washes[1],
				100*float64(contact)/float64(scanned),
				100*float64(inside)/float64(scanned), runmax, gm.w)
		}
		fmt.Println()
	}
}

// ---------------------------------------------------------------------------
// 2. THE CLOCK. Every s28-seamap number is 13:00.
func varyTOD() {
	fmt.Println("111x25, seed 7, SIX HOURS. Everything in notes/s28-seamap is 13:00.")
	fmt.Println("`clipped` counts columns that hit the scan floor and `noWarm` a reader that failed;")
	fmt.Println("either one non-zero means the R>B discriminator has stopped meaning `sand` at that hour.")
	fmt.Println()
	fmt.Printf("  %-7s %8s %8s %8s %8s %8s %9s %9s %10s\n",
		"hour", "top@0.0", "top@0.70", "top@1.00", "range", "0.70-1", "wash@1.0", "wave@0.0", "wave@1.00")
	for _, hr := range []float64{0, 4, 8, 13, 18, 21} {
		r0 := run(cfg{w: 111, h: 25, lv: 0.00, seed: 7, tod: hr / 24, writeRows: -1})
		r7 := run(cfg{w: 111, h: 25, lv: 0.70, seed: 7, tod: hr / 24, writeRows: -1})
		r1 := run(cfg{w: 111, h: 25, lv: 1.00, seed: 7, tod: hr / 24, writeRows: -1})
		rng := r1.topMean - r0.topMean
		pct := 0.0
		if rng != 0 {
			pct = 100 * (r1.topMean - r7.topMean) / rng
		}
		clip := r0.clipped + r7.clipped + r1.clipped
		fmt.Printf("  %02.0f:00   %8.2f %8.2f %8.2f %8.2f %7.0f%% %9.2f %9.1f %10.1f   clipped %d\n",
			hr, r0.topMean, r7.topMean, r1.topMean, rng, pct, r1.topSD, r0.wave, r1.wave, clip)
	}
}

// ---------------------------------------------------------------------------
// 3. THE SEED.
func varySeed() {
	fmt.Println("111x25, 13:00, SIX SEEDS. notes/s28-seamap only ever draws seed 7.")
	fmt.Println()
	fmt.Printf("  %-6s %8s %8s %8s %8s %8s %9s %9s\n",
		"seed", "top@0.0", "top@0.70", "top@1.00", "range", "0.70-1", "wash@1.0", "wave@1.00")
	for _, sd := range []int64{1, 3, 7, 42, 1234, 99999} {
		r0 := run(cfg{w: 111, h: 25, lv: 0.00, seed: sd, tod: 13.0 / 24, writeRows: -1})
		r7 := run(cfg{w: 111, h: 25, lv: 0.70, seed: sd, tod: 13.0 / 24, writeRows: -1})
		r1 := run(cfg{w: 111, h: 25, lv: 1.00, seed: sd, tod: 13.0 / 24, writeRows: -1})
		rng := r1.topMean - r0.topMean
		pct := 0.0
		if rng != 0 {
			pct = 100 * (r1.topMean - r7.topMean) / rng
		}
		fmt.Printf("  %-6d %8.2f %8.2f %8.2f %8.2f %7.0f%% %9.2f %9.1f\n",
			sd, r0.topMean, r7.topMean, r1.topMean, rng, pct, r1.topSD, r1.wave)
	}
}

// ---------------------------------------------------------------------------
// The guard's room is `beach - WriteRows - 1` and beach is c.H/5 floored at
// WriteRows+3. So the room is a step function of HEIGHT. If the guard is the
// cause, the wash's top-end response has to follow that step and nothing else.
func varyHeight() {
	fmt.Println("WIDTH FIXED AT 111, HEIGHT SWEPT. If the guard is the cause, the wash's 0.70->1.00")
	fmt.Println("response must track `room = beach - band - 1` and turn on where the room does.")
	fmt.Println()
	fmt.Printf("  %-8s %6s %5s %5s %6s %8s %8s %9s %9s\n",
		"height", "beach", "band", "room", "scale", "wash@.7", "wash@1.0", "d(wash)", "top range")
	for _, h := range []int{24, 25, 30, 35, 40, 45, 51} {
		wr := scape.DefaultWriteRows
		if m := h / 6; wr > m {
			wr = m
		}
		beach := h / 5
		if min := wr + 3; beach < min {
			beach = min
		}
		if beach < 4 {
			beach = 4
		}
		scale := math.Max(0.45, math.Min(1.35, math.Min(111.0/80.0, float64(h)/24.0)))
		r0 := run(cfg{w: 111, h: h, lv: 0.00, seed: 7, tod: 13.0 / 24, writeRows: -1})
		r7 := run(cfg{w: 111, h: h, lv: 0.70, seed: 7, tod: 13.0 / 24, writeRows: -1})
		r1 := run(cfg{w: 111, h: h, lv: 1.00, seed: 7, tod: 13.0 / 24, writeRows: -1})
		fmt.Printf("  %-8d %6d %5d %5d %6.2f %8.2f %8.2f %9.2f %9.2f\n",
			h, beach, wr, beach-wr-1, scale, r7.topSD, r1.topSD, r1.topSD-r7.topSD, r1.topMean-r0.topMean)
	}
}

// ---------------------------------------------------------------------------
// s28-seamap says the swing/crest channels died because tideEdge has no level
// term in x, and says plainly it did not vary that. It is true by construction
// -- the x terms are 0.60*scale*sin(fx*0.11)+0.35*scale*sin(fx*0.047), no level
// anywhere -- but then the swing should be FLAT with level, and the rendered
// table has it FALLING (1.93 -> 1.61 at 111x25). Something else is eating it.
// The guard rescales the WHOLE edge toward ref by k=room/dev, which compresses
// the coast as well as the wash, so turning the guard off separates the two.
func varySwing() {
	fmt.Println("Does the guard squash the COAST as well as the wash? Guard on vs off, identical geometry.")
	fmt.Println()
	fmt.Printf("  %-14s %-24s %8s %8s %8s %8s\n", "geometry", "config", "swing@.7", "swing@1", "crest@.7", "crest@1")
	for _, gm := range []struct {
		name string
		w, h int
	}{{"80x24", 80, 24}, {"111x25", 111, 25}, {"153x51 ctrl", 153, 51}} {
		wr := scape.DefaultWriteRows
		if m := gm.h / 6; wr > m {
			wr = m
		}
		beach := gm.h / 5
		if min := wr + 3; beach < min {
			beach = min
		}
		for _, c := range []struct {
			label string
			wrows int
		}{{"SHIPPED (guard on)", 4}, {"GUARD OFF", 0}} {
			r7 := run(cfg{w: gm.w, h: gm.h, lv: 0.70, seed: 7, tod: 13.0 / 24, writeRows: c.wrows, sandRows: beach})
			r1 := run(cfg{w: gm.w, h: gm.h, lv: 1.00, seed: 7, tod: 13.0 / 24, writeRows: c.wrows, sandRows: beach})
			fmt.Printf("  %-14s %-24s %8.2f %8.2f %8.2f %8.2f\n",
				gm.name, c.label, r7.swing, r1.swing, r7.crests, r1.crests)
		}
	}
}

// ---------------------------------------------------------------------------
// "The sea vanishes at 40x12" rests entirely on the R>B reader agreeing that
// row hy+1 is SAND. Print what is actually there instead of assuming it.
func probe() {
	w, h := 40, 12
	for _, lv := range []float64{0.00, 0.20, 1.00} {
		sh := scape.NewShore(7, false)
		act := scape.Activity{Level: lv, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
		tm := 0.0
		for k := 0; k < 500; k++ {
			tm += dt
			sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm, act)
		}
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += dt
		sh.Update(c, tm, act)
		fmt.Printf("=== 40x12, level %.2f, 13:00, horizon row %d, column 20 ===\n", lv, int(float64(h)*0.42))
		for y := 0; y < h; y++ {
			r, _, bg := c.ResolveAt(20, y, term.Profile256)
			tag := "  sky/sea"
			if bg.R > bg.B {
				tag = "  <- WARM (read as sand)"
			}
			fmt.Printf("  y=%2d  bg=(%3d,%3d,%3d)  glyph=%q%s\n", y, bg.R, bg.G, bg.B, r, tag)
		}
		fmt.Println()
	}
}
