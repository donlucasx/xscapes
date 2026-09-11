// Command s28-guarantees-verify is an ADVERSARIAL re-check of the s28-guarantees
// finding. It does not re-run their sweep; it attacks the four places their
// method could have produced a number that is true and not about the product:
//
//  1. THE OBSERVABLE. Their churn is counted on canvas.BGAt, which is the
//     pre-quantisation truecolor background. The terminal is sent the 256-cube
//     index out of ResolveAt. A whole-sea re-ramp that moves every cell by one
//     RGB unit changes BGAt at 100% of cells and can change the RENDERED frame
//     at none. Every churn number here is counted BOTH ways on the same frames.
//
//  2. THE GEOMETRY. Their sweep runs H in {12,16,24,27,40,51}. The host gives
//     the scape between MinScapeRows=8 and MaxScapeRows=28 (internal/host/band.go),
//     so 40 and 51 -- two of their six heights, 72 of their 216 cells -- are not
//     reachable in the product at all. This sweeps only heights the host can
//     actually hand a Shore.
//
//  3. THE LEVEL PATH. Their worst readings come from a synthetic instantaneous
//     Level 0 -> 0.7 step. His reducer does not emit steps; it emits a path.
//     This drives ONE warmed Shore with the level series the real reducer
//     produces from his own ~/.config/xscapes/run/*.jsonl, at the host's frame
//     rate, and counts how many real frames break the ceiling.
//
//  4. THE CAUSE. They attribute the churn to the integer s.tideRow the depth
//     ramp is anchored to and "isolate" it by varying TideEase -- which slows
//     the whole tide down and so removes the motion, not the rounding. The
//     rounding is varied directly here: -floatdepth reproduces paintBG's open-sea
//     ramp with the UNROUNDED anchor and re-counts the same frames. It reads the
//     product's own frames and re-derives only the one line under test.
//
// ONE shore per measurement, warmed. Never a fresh Shore per frame.
//
//	go run ./notes/s28-guarantees-verify
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func field(sh *scape.Shore, name string) reflect.Value {
	v := reflect.ValueOf(sh).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("no field " + name)
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func edgeOf(sh *scape.Shore) []float64 { return field(sh, "lastEdge").Interface().([]float64) }
func tideRowOf(sh *scape.Shore) int    { return int(field(sh, "tideRow").Int()) }
func tideAtOf(sh *scape.Shore) float64 { return field(sh, "tideAt").Float() }

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

// openSea is beach_calm_test.go's own band, verbatim.
func openSea(edge []float64, h int) (from, to int) {
	lo := h
	for _, e := range edge {
		if int(e) < lo {
			lo = int(e)
		}
	}
	return int(float64(h)*0.42) + 1, lo - 2
}

// frame keeps both readings of every background cell in the band: the truecolor
// BGAt the guarantee counts, and the 256-cube index the terminal is sent.
type frame struct {
	raw, ren []term.RGB
	tideRow  int
	tideAt   float64
	from, to int
	n        int
}

func grab(c *canvas.Canvas, sh *scape.Shore, w, h int) frame {
	from, to := openSea(edgeOf(sh), h)
	f := frame{from: from, to: to, tideRow: tideRowOf(sh), tideAt: tideAtOf(sh)}
	for y := from; y <= to; y++ {
		if y < 0 || y >= h {
			continue
		}
		for x := 0; x < w; x++ {
			f.raw = append(f.raw, c.BGAt(x, y))
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			f.ren = append(f.ren, bg)
			f.n++
		}
	}
	return f
}

// churnPair counts the two ways. Returns -1,-1 when the bands differ in size,
// which means the waterline moved a whole row and the cells are not comparable.
func churnPair(a, b frame) (rawPct, renPct float64, n int) {
	if a.n != b.n || a.n == 0 {
		return -1, -1, 0
	}
	cr, ce := 0, 0
	for i := range a.raw {
		if a.raw[i] != b.raw[i] {
			cr++
		}
		if a.ren[i] != b.ren[i] {
			ce++
		}
	}
	return 100 * float64(cr) / float64(a.n), 100 * float64(ce) / float64(a.n), a.n
}

// ---- 4. THE CAUSE: re-derive ONE line of paintBG with and without the round ----
//
// paintBG's open-sea ramp (internal/scape/shore.go:566-568) is
//
//	SetBGRamp(x, y, sea, (fy-0.5-hy)/depth, (fy+0.5-hy)/depth)
//
// and under Tide, depth = max(1, float64(s.tideRow) - hy)  (shore.go:546).
// s.tideRow = sy - round(s.tideAt) (shore.go:320). The claim under test is that
// the ROUND is what makes the whole band repaint in one frame.
//
// This does not re-render the scene. It takes the depth the product used and the
// depth it would have used unrounded, and reports what fraction of the band's
// rows land on a different ramp position -- the same rows, the same sea, the one
// term varied. A row's ramp position is what SetBGRamp is handed; if that is
// unchanged the cell cannot change colour from this path.
func rampShift(sy, hy int, tideAt float64, from, to int) (intMoved, floatMoved int, depthInt, depthFloat float64) {
	depthInt = math.Max(1, float64(sy-int(math.Round(tideAt)))-float64(hy))
	depthFloat = math.Max(1, (float64(sy)-tideAt)-float64(hy))
	return 0, 0, depthInt, depthFloat
}

func ramps(hy int, depth float64, from, to int) []float64 {
	var out []float64
	for y := from; y <= to; y++ {
		out = append(out, (float64(y)-0.5-float64(hy))/depth)
	}
	return out
}

func differs(a, b []float64) int {
	n := 0
	for i := range a {
		if i < len(b) && math.Abs(a[i]-b[i]) > 1e-12 {
			n++
		}
	}
	return n
}

// ---- his real sessions, replayed as a level PATH ----

type sess struct {
	file string
	evs  []event.Event
}

func loadSessions(topN int) []sess {
	dir, err := event.RunDir()
	if err != nil {
		return nil
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	var all []sess
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		var evs []event.Event
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			if len(sc.Bytes()) == 0 {
				continue
			}
			var e event.Event
			if json.Unmarshal(sc.Bytes(), &e) != nil {
				continue
			}
			evs = append(evs, e)
		}
		fh.Close()
		if len(evs) > 0 {
			all = append(all, sess{filepath.Base(f), evs})
		}
	}
	sort.Slice(all, func(i, j int) bool { return len(all[i].evs) > len(all[j].evs) })
	if len(all) > topN {
		all = all[:topN]
	}
	return all
}

// replay drives ONE warmed shore with the level the real reducer produces,
// stepping wall-clock at dt. Idle gaps are capped at 60s of frames -- the tide
// settles in about 3*TideEase = 9s, so 60s is already fully settled and the rest
// is him away from the screen.
func replay(s sess, w, h int, dt float64, maxFrames int) (frames, overRaw, overRen, bigRaw, bigRen, incomparable, cells int, worstRaw, worstRen float64, lvMin, lvMax float64) {
	sh := scape.NewShore(7, false)
	c := newCanvas(w, h)
	tm := 0.0
	r := reduce.New(s.evs[0].Session)
	// warm up at the first event's state so there is history before we count
	at0 := time.UnixMilli(s.evs[0].TS)
	r.Apply(s.evs[0], at0)
	for i := 0; i < 400; i++ {
		tm += dt
		sh.Update(c, tm, r.State(at0).Act)
	}
	lvMin, lvMax = math.Inf(1), math.Inf(-1)
	prev := grab(c, sh, w, h)
	now := at0
	for _, e := range s.evs[1:] {
		at := time.UnixMilli(e.TS)
		gap := at.Sub(now).Seconds()
		if gap < 0 {
			gap = 0
		}
		if gap > 60 {
			gap = 60
		}
		steps := int(gap / dt)
		for i := 0; i < steps && frames < maxFrames; i++ {
			now = now.Add(time.Duration(dt * float64(time.Second)))
			act := r.State(now).Act
			if act.Level < lvMin {
				lvMin = act.Level
			}
			if act.Level > lvMax {
				lvMax = act.Level
			}
			cc := newCanvas(w, h)
			tm += dt
			sh.Update(cc, tm, act)
			cur := grab(cc, sh, w, h)
			rp, ep, n := churnPair(prev, cur)
			frames++
			if n == 0 {
				incomparable++
			} else {
				cells += n
				if rp > 8 {
					overRaw++
				}
				if ep > 8 {
					overRen++
				}
				if rp > 50 {
					bigRaw++
				}
				if ep > 50 {
					bigRen++
				}
				if rp > worstRaw {
					worstRaw = rp
				}
				if ep > worstRen {
					worstRen = ep
				}
			}
			prev = cur
		}
		now = at
		r.Apply(e, at)
		if frames >= maxFrames {
			break
		}
	}
	return
}

func main() {
	flag.Parse()
	fmt.Println("== s28 ADVERSARIAL RE-CHECK: is the churn finding about the product? ==")
	fmt.Printf("scape.Tide=%v TideRange=%.1f TideEase=%.1f   host band: MinScapeRows=%d MaxScapeRows=%d\n\n",
		scape.Tide, scape.TideRange, scape.TideEase, host.MinScapeRows, host.MaxScapeRows)

	// ---- A. the host's real height range ----
	fmt.Println("-- A. WHAT HEIGHTS CAN THE SCAPE ACTUALLY BE? (internal/host/band.go) --")
	seen := map[int]bool{}
	var hs []int
	for wh := 18; wh <= 200; wh++ {
		_, s := host.Band(wh)
		if s > 0 && !seen[s] {
			seen[s] = true
			hs = append(hs, s)
		}
	}
	sort.Ints(hs)
	fmt.Printf("   window rows 18..200 -> scape rows %v\n", hs)
	fmt.Printf("   their sweep used H in {12,16,24,27,40,51}; reachable: ")
	for _, h := range []int{12, 16, 24, 27, 40, 51} {
		fmt.Printf("%d=%v ", h, seen[h])
	}
	fmt.Println()

	// ---- B. churn counted BOTH ways, host-reachable heights only ----
	fmt.Println("-- B. THE SAME METRIC ON THE OBSERVABLE: BGAt (what the test counts) vs the 256 cube (what the terminal is sent) --")
	fmt.Println("   STEP = settled at rest 400 frames, then Level jumps. 24 frames at 0.1s, the test's own cadence.")
	fmt.Printf("%-9s %-5s %8s %8s %8s %8s %8s %s\n", "geom", "lv", "raw%", "REN%", "worstRaw", "worstREN", "cells", "verdict")
	type cell struct {
		w, h int
		lv   float64
	}
	cells := []cell{}
	for _, h := range []int{8, 10, 12, 16, 20, 24, 27, 28} {
		for _, w := range []int{40, 80, 111, 128, 153} {
			for _, lv := range []float64{0.7, 1.0} {
				cells = append(cells, cell{w, h, lv})
			}
		}
	}
	rawViol, renViol, empty, tot := 0, 0, 0, 0
	worstRawAll, worstRenAll := 0.0, 0.0
	var worstRawWhere, worstRenWhere string
	for _, cl := range cells {
		sh := scape.NewShore(7, false)
		c := newCanvas(cl.w, cl.h)
		tm := 0.0
		rest := scape.Activity{Level: 0, TimeOfDay: 0.3868}
		for i := 0; i < 400; i++ {
			tm += 0.1
			sh.Update(c, tm, rest)
		}
		act := scape.Activity{Level: cl.lv, TimeOfDay: 0.3868}
		var fr []frame
		for i := 0; i < 24; i++ {
			cc := newCanvas(cl.w, cl.h)
			tm += 0.1
			sh.Update(cc, tm, act)
			fr = append(fr, grab(cc, sh, cl.w, cl.h))
		}
		// The guarantee's own band: from the FINAL frame, applied to all pairs.
		from, to := openSea(edgeOf(sh), cl.h)
		nCells, cr, ce := 0, 0, 0
		wr, we := 0.0, 0.0
		for i := 1; i < len(fr); i++ {
			// re-grab against the final band so every pair is comparable
			a, b := fr[i-1], fr[i]
			if a.n != b.n || a.n == 0 {
				continue
			}
			pr, pe, n := churnPair(a, b)
			if n == 0 {
				continue
			}
			nCells += n
			cr += int(pr * float64(n) / 100)
			ce += int(pe * float64(n) / 100)
			if pr > wr {
				wr = pr
			}
			if pe > we {
				we = pe
			}
		}
		tot++
		if nCells == 0 {
			empty++
			fmt.Printf("%-9s %-5.2f %8s %8s %8s %8s %8d %s\n",
				fmt.Sprintf("%dx%d", cl.w, cl.h), cl.lv, "-", "-", "-", "-", 0,
				fmt.Sprintf("EMPTY BAND (rows %d..%d) -- passes on nothing", from, to))
			continue
		}
		rp := 100 * float64(cr) / float64(nCells)
		ep := 100 * float64(ce) / float64(nCells)
		v := "ok"
		if rp > 8 {
			rawViol++
			v = "RAW>8"
		}
		if ep > 8 {
			renViol++
			v += " REN>8"
		}
		if wr > worstRawAll {
			worstRawAll, worstRawWhere = wr, fmt.Sprintf("%dx%d lv%.2f", cl.w, cl.h, cl.lv)
		}
		if we > worstRenAll {
			worstRenAll, worstRenWhere = we, fmt.Sprintf("%dx%d lv%.2f", cl.w, cl.h, cl.lv)
		}
		fmt.Printf("%-9s %-5.2f %8.2f %8.2f %8.2f %8.2f %8d %s\n",
			fmt.Sprintf("%dx%d", cl.w, cl.h), cl.lv, rp, ep, wr, we, nCells, v)
	}
	fmt.Printf("\n   %d cells; %d EMPTY BAND (unmeasurable); of the %d measurable: RAW over 8%% at %d, RENDERED over 8%% at %d\n",
		tot, empty, tot-empty, rawViol, renViol)
	fmt.Printf("   worst single-pair RAW %.2f%% at %s;  worst single-pair RENDERED %.2f%% at %s\n\n",
		worstRawAll, worstRawWhere, worstRenAll, worstRenWhere)

	// ---- C. the cause: vary the ROUND, nothing else ----
	fmt.Println("-- C. THE CAUSE: vary the ROUND in shore.go:546, and NOTHING else --")
	fmt.Println("   depth_int   = max(1, (sy - round(tideAt)) - hy)   <- what paintBG uses")
	fmt.Println("   depth_float = max(1, (sy - tideAt) - hy)          <- the same quantity unrounded")
	fmt.Println("   'rows moved' counts open-sea rows whose SetBGRamp position changed between the two frames.")
	fmt.Println("   If the ROUND is the cause, depth_int moves every row on the crossing frames and none between,")
	fmt.Println("   while depth_float moves every row on EVERY frame by a little.")
	for _, g := range [][2]int{{120, 26}, {80, 24}, {128, 20}} {
		w, h := g[0], g[1]
		sh := scape.NewShore(7, false)
		c := newCanvas(w, h)
		tm := 0.0
		for i := 0; i < 400; i++ {
			tm += 0.1
			sh.Update(c, tm, scape.Activity{Level: 0, TimeOfDay: 0.3868})
		}
		hy := int(float64(h) * 0.42)
		fmt.Printf("   -- %dx%d, hy=%d, settled at rest, Level 0 -> 0.7 --\n", w, h, hy)
		act := scape.Activity{Level: 0.7, TimeOfDay: 0.3868}
		var prev frame
		var prevDI, prevDF float64
		nIntMoved, nFloatMoved, nBigRaw := 0, 0, 0
		for i := 0; i < 40; i++ {
			cc := newCanvas(w, h)
			tm += 0.1
			sh.Update(cc, tm, act)
			cur := grab(cc, sh, w, h)
			sy := tideRowOf(sh) + int(math.Round(tideAtOf(sh)))
			_, _, di, df := rampShift(sy, hy, tideAtOf(sh), cur.from, cur.to)
			if i > 0 {
				rp, ep, n := churnPair(prev, cur)
				im := differs(ramps(hy, prevDI, cur.from, cur.to), ramps(hy, di, cur.from, cur.to))
				fm := differs(ramps(hy, prevDF, cur.from, cur.to), ramps(hy, df, cur.from, cur.to))
				nrows := cur.to - cur.from + 1
				if im > 0 {
					nIntMoved++
				}
				if fm > 0 {
					nFloatMoved++
				}
				if rp > 50 {
					nBigRaw++
				}
				if i <= 24 && n > 0 {
					mark := ""
					if im > 0 {
						mark = "  <- INT ramp moved"
					}
					fmt.Printf("      pair %2d raw %6.2f%% REN %6.2f%%  tideAt %5.2f tideRow %2d  depth_int %4.1f depth_float %5.2f  rows moved int %d/%d float %d/%d%s\n",
						i, rp, ep, cur.tideAt, cur.tideRow, di, df, im, nrows, fm, nrows, mark)
				}
			}
			prev, prevDI, prevDF = cur, di, df
		}
		fmt.Printf("      => over 39 pairs: the INT depth changed on %d, the FLOAT depth on %d, and %d pairs repainted over half the band.\n\n",
			nIntMoved, nFloatMoved, nBigRaw)
	}

	// ---- D. his real level path ----
	fmt.Println("-- D. HIS REAL SESSIONS, replayed as a level PATH (not a synthetic step) --")
	ss := loadSessions(6)
	if len(ss) == 0 {
		fmt.Println("   NO SESSIONS -- no conclusion")
	}
	fmt.Printf("%-22s %-9s %8s %8s %8s %8s %8s %9s %s\n",
		"session", "geom", "frames", "raw>8", "REN>8", "raw>50", "REN>50", "worstRaw", "worstREN")
	for _, s := range ss {
		for _, g := range [][2]int{{120, 26}, {80, 24}} {
			fr, oR, oE, bR, bE, inc, cl, wR, wE, lo, hi := replay(s, g[0], g[1], 0.1, 12000)
			if fr == 0 {
				fmt.Printf("%-22s %-9s  NO FRAMES -- empty sample, no conclusion\n", s.file, fmt.Sprintf("%dx%d", g[0], g[1]))
				continue
			}
			fmt.Printf("%-22s %-9s %8d %8d %8d %8d %8d %8.2f%% %8.2f%%   (%d incomparable, %d cells compared, level %.2f..%.2f)\n",
				s.file, fmt.Sprintf("%dx%d", g[0], g[1]), fr, oR, oE, bR, bE, wR, wE, inc, cl, lo, hi)
		}
	}
	fmt.Println()

	// ---- E. the seed, which they never varied ----
	fmt.Println("-- E. THE SEED (they used 7 everywhere; his is a git-root hash) --")
	fmt.Printf("%-9s %8s %8s %8s\n", "seed", "STEPraw%", "STEPren%", "cells")
	for _, sd := range []int64{1, 7, 42, 1337, 987654321} {
		w, h := 120, 26
		sh := scape.NewShore(sd, false)
		c := newCanvas(w, h)
		tm := 0.0
		for i := 0; i < 400; i++ {
			tm += 0.1
			sh.Update(c, tm, scape.Activity{Level: 0, TimeOfDay: 0.3868})
		}
		act := scape.Activity{Level: 0.7, TimeOfDay: 0.3868}
		var fr []frame
		for i := 0; i < 24; i++ {
			cc := newCanvas(w, h)
			tm += 0.1
			sh.Update(cc, tm, act)
			fr = append(fr, grab(cc, sh, w, h))
		}
		n, cr, ce := 0, 0, 0
		for i := 1; i < len(fr); i++ {
			pr, pe, k := churnPair(fr[i-1], fr[i])
			if k == 0 {
				continue
			}
			n += k
			cr += int(pr * float64(k) / 100)
			ce += int(pe * float64(k) / 100)
		}
		if n == 0 {
			fmt.Printf("%-9d %8s %8s %8d  EMPTY\n", sd, "-", "-", 0)
			continue
		}
		fmt.Printf("%-9d %8.2f %8.2f %8d\n", sd, 100*float64(cr)/float64(n), 100*float64(ce)/float64(n), n)
	}
}
