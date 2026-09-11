// Command s28-guarantees re-runs the tide's three LOCKED guarantees, plus the
// two clamps tide.go relies on, across a sweep of geometries.
//
// All three guarantees caught the first tide build on 2026-09-10 and all three
// were measured at ONE geometry -- 120x26 for the churn and the tone count,
// 80x24 for the teleport. The scape is designed for 80x24, must "look fine at
// 40x12", and he runs it anywhere from ~107 to ~153 columns.
//
// THE METRICS ARE NOT REDEFINED HERE. Each is lifted verbatim from the test
// that owns it, and the calibration block at the top reproduces that test's own
// number at that test's own geometry before anything else is believed:
//
//	churn     TestTheBackdropHoldsStillBetweenFrames  internal/scape/beach_calm_test.go:97
//	          % of OPEN-SEA background cells that change between consecutive
//	          frames, 24 frames 0.1s apart, level 0.7, tod 0.3868. Ceiling 8%.
//	          Open sea is the test's own openSea(): rows int(H*0.42)+1 ..
//	          min(int(lastEdge))-2.
//	tones     TestNoRowIsAConfettiOfNearIdenticalTones  beach_calm_test.go:47
//	          max distinct BGAt colours on any one row of a frame. Ceiling 40.
//	teleport  TestActivityChangeDoesNotTeleportTheSea  shore_test.go:25
//	          mean |delta| of the waterline per column across a 0.30 -> 0.35
//	          activity step, against the mean over the last 40 steady frames.
//	          Ceiling: 8x the steady baseline.
//	band      TestTheWaterNeverReachesTheWritingBand  beach_calm_test.go:161
//	          no lastEdge value past writeTop-1.
//	horizon   tide.go's hyFloor, which has no test. hyFloor(sy) = sy-TideRange-1
//	          is meant to stop "a long quiet stretch pulling the shoreline up
//	          into the sky" -- but it is a function of sy alone and never sees
//	          the horizon row hy at all. Checked here directly: min(lastEdge)
//	          against hy, and the rendered open-sea row count.
//
// WHY reflect+unsafe: lastEdge, writeTop, tideAt and tideRow are unexported and
// the guarantees are defined on them. Copying the waterline maths into this
// file would measure a COPY of the product, which is how you get a green
// instrument over a red product. This reads the real fields the real tests read.
// Nothing here writes to product state except scape.Tide, which is an exported
// switch and is restored.
//
// The rendered frame is read too, and separately: the sea/sand boundary found
// by walking each column down from the horizon to the first WARM background
// cell (the sea is blue, the beach is warm; searange's technique). The
// waterline lives in the BACKGROUND and a glyph-only reader cannot see it.
//
// ONE shore per measurement, WARMED UP. A fresh scape.NewShore per frame has no
// history -- Shore.Update clamps any gap over a second to one nominal step --
// so two frames from two new Shores can be identical. That trap voided
// notes/searange once and notes/drift once.
//
//	go run ./notes/s28-guarantees            sweeps BOTH tide settings
//	XSCAPES_TIDE=0 go run ./notes/s28-guarantees -env   only what the env says
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
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// ---------- reading the fields the guarantees are defined on ----------

func field(sh *scape.Shore, name string) reflect.Value {
	v := reflect.ValueOf(sh).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("no field " + name)
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func edgeOf(sh *scape.Shore) []float64 { return field(sh, "lastEdge").Interface().([]float64) }
func writeTopOf(sh *scape.Shore) int   { return int(field(sh, "writeTop").Int()) }
func tideRowOf(sh *scape.Shore) int    { return int(field(sh, "tideRow").Int()) }
func tideAtOf(sh *scape.Shore) float64 { return field(sh, "tideAt").Float() }

// syOf recovers Update's sy: tideRow = sy - round(tideAt).
func syOf(sh *scape.Shore) int { return tideRowOf(sh) + int(math.Round(tideAtOf(sh))) }

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

// ---------- the metrics, lifted from the tests that own them ----------

// rowTones is beach_calm_test.go's rowTones: distinct BGAt colours per row.
func rowTones(c *canvas.Canvas) (worst, at int) {
	for y := 0; y < c.H; y++ {
		seen := map[term.RGB]bool{}
		for x := 0; x < c.W; x++ {
			seen[c.BGAt(x, y)] = true
		}
		if len(seen) > worst {
			worst, at = len(seen), y
		}
	}
	return
}

// rowTonesRendered is the same count on what the terminal is actually sent:
// the 256-cube-quantised background out of ResolveAt.
func rowTonesRendered(c *canvas.Canvas) (worst, at int) {
	for y := 0; y < c.H; y++ {
		seen := map[term.RGB]bool{}
		for x := 0; x < c.W; x++ {
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			seen[bg] = true
		}
		if len(seen) > worst {
			worst, at = len(seen), y
		}
	}
	return
}

// openSea is beach_calm_test.go's openSea, verbatim.
func openSea(edge []float64, h int) (from, to int) {
	lo := h
	for _, e := range edge {
		if int(e) < lo {
			lo = int(e)
		}
	}
	return int(float64(h)*0.42) + 1, lo - 2
}

func meanAbsDiff(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += math.Abs(a[i] - b[i])
	}
	return sum / float64(n)
}

// renderedWaterline walks each column down from the horizon to the first WARM
// background cell. Sea is blue, beach is warm, so the boundary is exact.
// Returns h when a column never turns warm (no beach visible at all).
func renderedWaterline(c *canvas.Canvas, hy int) []int {
	out := make([]int, c.W)
	for x := 0; x < c.W; x++ {
		out[x] = c.H
		for y := hy + 1; y < c.H; y++ {
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			if bg.R > bg.B {
				out[x] = y
				break
			}
		}
	}
	return out
}

// ---------- one geometry ----------

type row struct {
	w, h  int
	level float64
	tod   float64

	// Three churn readings, all the same metric, three different histories.
	// COLD is the guarantee exactly as written -- a fresh shore, 24 frames
	// from t=0.1. WARM is the same thing after 400 frames at a constant
	// level, i.e. the tide settled. STEP warms at level 0 and then steps to
	// this row's level, which is the tide actually MOVING -- the only state
	// in which the tide can churn the backdrop at all, and the one the
	// guarantee was written to catch.
	churnCold, churnWarm, churnStep float64
	cellsCold, cellsWarm, cellsStep int
	bandCold, bandWarm, bandStep    [2]int // from..to, the test's own band
	tonesCold, tonesColdRow         int    // the test's own single frame
	tonesBG, tonesBGRow, tonesRen   int    // worst of 24 warmed frames
	minEdge, maxEdge                float64
	hy, sy, writeTop                int
	hyFloor                         float64
	renSeaRows, renNoBeach          int
	degenerate                      bool
}

const (
	churnFrames = 24
	warmFrames  = 400
	frameDT     = 0.1
)

// churnRun measures the guarantee's metric over churnFrames frames of the shore
// it is handed, and returns the percentage, the cells compared and the band.
func churnRun(sh *scape.Shore, w, h int, act scape.Activity, tm *float64) (pct float64, cells int, band [2]int, frames []*canvas.Canvas) {
	for i := 0; i < churnFrames; i++ {
		cc := newCanvas(w, h)
		*tm += frameDT
		sh.Update(cc, *tm, act)
		frames = append(frames, cc)
	}
	from, to := openSea(edgeOf(sh), h)
	band = [2]int{from, to}
	changed := 0
	for i := 1; i < len(frames); i++ {
		for y := from; y <= to; y++ {
			if y < 0 || y >= h {
				continue
			}
			for x := 0; x < w; x++ {
				cells++
				if frames[i].BGAt(x, y) != frames[i-1].BGAt(x, y) {
					changed++
				}
			}
		}
	}
	if cells > 0 {
		pct = 100 * float64(changed) / float64(cells)
	}
	return
}

func measure(w, h int, level, tod float64) row {
	r := row{w: w, h: h, level: level, tod: tod}
	act := scape.Activity{Level: level, TimeOfDay: tod}

	// --- COLD: the guarantee verbatim, fresh shore, no warm-up ---
	shC := scape.NewShore(7, false)
	tmC := 0.0
	pc, cc, bc, frC := churnRun(shC, w, h, act, &tmC)
	if len(edgeOf(shC)) == 0 {
		r.degenerate = true
		return r
	}
	r.churnCold, r.cellsCold, r.bandCold = pc, cc, bc
	r.tonesCold, r.tonesColdRow = rowTones(frC[0])

	// --- WARM: one shore, 400 frames, then the same 24 ---
	sh := scape.NewShore(7, false)
	c := newCanvas(w, h)
	tm := 0.0
	for i := 0; i < warmFrames; i++ {
		tm += frameDT
		sh.Update(c, tm, act)
	}
	r.hy = int(float64(h) * 0.42)
	r.sy = syOf(sh)
	r.writeTop = writeTopOf(sh)
	r.hyFloor = float64(r.sy) - scape.TideRange - 1
	pw, cw, bw, frW := churnRun(sh, w, h, act, &tm)
	r.churnWarm, r.cellsWarm, r.bandWarm = pw, cw, bw
	for _, f := range frW {
		if n, at := rowTones(f); n > r.tonesBG {
			r.tonesBG, r.tonesBGRow = n, at
		}
		if n, _ := rowTonesRendered(f); n > r.tonesRen {
			r.tonesRen = n
		}
	}

	// --- STEP: settled at rest, then the activity jumps to this level and
	// the tide runs. Skipped when level is already 0.
	if level > 0 {
		shS := scape.NewShore(7, false)
		cs := newCanvas(w, h)
		tms := 0.0
		rest := scape.Activity{Level: 0, TimeOfDay: tod}
		for i := 0; i < warmFrames; i++ {
			tms += frameDT
			shS.Update(cs, tms, rest)
		}
		ps, cls, bs, frS := churnRun(shS, w, h, act, &tms)
		r.churnStep, r.cellsStep, r.bandStep = ps, cls, bs
		for _, f := range frS {
			if n, at := rowTones(f); n > r.tonesBG {
				r.tonesBG, r.tonesBGRow = n, at
			}
			if n, _ := rowTonesRendered(f); n > r.tonesRen {
				r.tonesRen = n
			}
		}
	}

	// --- edge extremes over 80 further phases, warmed ---
	r.minEdge, r.maxEdge = math.Inf(1), math.Inf(-1)
	for i := 0; i < 80; i++ {
		tm += frameDT
		sh.Update(c, tm, act)
		for _, e := range edgeOf(sh) {
			if e < r.minEdge {
				r.minEdge = e
			}
			if e > r.maxEdge {
				r.maxEdge = e
			}
		}
	}

	// --- rendered corroboration: the sea/sand boundary in the real frame ---
	wl := renderedWaterline(c, r.hy)
	lo := h
	for _, v := range wl {
		if v == c.H {
			r.renNoBeach++
		}
		if v < lo {
			lo = v
		}
	}
	r.renSeaRows = lo - (r.hy + 1)
	return r
}

// teleport is TestActivityChangeDoesNotTeleportTheSea at an arbitrary geometry.
func teleport(w, h int, quiet, busy float64) (baseline, jump float64, ok bool) {
	const fps, age = 20.0, 600.0
	sh := scape.NewShore(7, false)
	c := newCanvas(w, h)
	var prev []float64
	var steady float64
	n := 0
	for i := 0; i < int(age*fps); i++ {
		sh.Update(c, float64(i)/fps, scape.Activity{Working: true, Level: quiet})
		cur := append([]float64(nil), edgeOf(sh)...)
		if prev != nil && i > int(age*fps)-40 {
			steady += meanAbsDiff(prev, cur)
			n++
		}
		prev = cur
	}
	if n == 0 || len(prev) == 0 {
		return 0, 0, false
	}
	baseline = steady / float64(n)
	sh.Update(c, age, scape.Activity{Working: true, Level: busy})
	jump = meanAbsDiff(prev, edgeOf(sh))
	return baseline, jump, true
}

// ---------- his real sessions ----------

// levelSteps replays every event in ~/.config/xscapes/run/*.jsonl through the
// real reducer and reports the distribution of the per-frame change in
// Activity.Level -- because the teleport guarantee is defined on a 0.05 step
// and nobody has checked what size step his sessions actually produce.
func levelSteps() {
	dir, err := event.RunDir()
	if err != nil {
		fmt.Println("  (no run dir:", err, ")")
		return
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	sort.Strings(files)
	var steps []float64
	var activeSecs, movingSecs float64
	events, sessions, withEvents := 0, 0, 0
	maxStep, maxWhere := 0.0, ""
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		sessions++
		var evs []event.Event
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Bytes()
			if len(line) == 0 {
				continue
			}
			var e event.Event
			if json.Unmarshal(line, &e) != nil {
				continue
			}
			evs = append(evs, e)
		}
		fh.Close()
		if len(evs) == 0 {
			continue
		}
		withEvents++
		events += len(evs)
		r := reduce.New(evs[0].Session)
		last := -1.0
		lastT := time.Time{}
		var moveUntil time.Time
		for _, e := range evs {
			at := time.UnixMilli(e.TS)
			r.Apply(e, at)
			lv := r.State(at).Act.Level
			if last >= 0 {
				d := math.Abs(lv - last)
				steps = append(steps, d)
				if d > maxStep {
					maxStep, maxWhere = d, filepath.Base(f)
				}
				// Wall-clock accounting. A gap over a minute is him away from
				// the screen, not a scene anybody is looking at.
				gap := at.Sub(lastT).Seconds()
				if gap >= 0 && gap < 60 {
					activeSecs += gap
					// how much of this interval was inside a tide transition
					lo := lastT
					if moveUntil.After(lo) {
						hi := moveUntil
						if hi.After(at) {
							hi = at
						}
						movingSecs += hi.Sub(lo).Seconds()
					}
				}
				if d > 0.02 {
					moveUntil = at.Add(time.Duration(scape.TideEase * float64(time.Second)))
				}
			}
			last, lastT = lv, at
		}
	}
	fmt.Printf("  %d jsonl files, %d with at least one event, %d events, %d level transitions\n",
		len(files), withEvents, events, len(steps))
	if len(steps) == 0 {
		fmt.Println("  EMPTY SAMPLE -- no conclusion")
		return
	}
	sort.Float64s(steps)
	q := func(p float64) float64 { return steps[int(p*float64(len(steps)-1))] }
	over := func(t float64) int {
		n := 0
		for _, v := range steps {
			if v > t {
				n++
			}
		}
		return n
	}
	fmt.Printf("  |dLevel| between consecutive events: p50 %.4f  p90 %.4f  p99 %.4f  max %.4f (%s)\n",
		q(.5), q(.9), q(.99), maxStep, maxWhere)
	fmt.Printf("  transitions bigger than the guarantee's own 0.05 step: %d (%.2f%%);  bigger than 0.20: %d (%.2f%%)\n",
		over(0.05), 100*float64(over(0.05))/float64(len(steps)),
		over(0.20), 100*float64(over(0.20))/float64(len(steps)))
	fmt.Printf("  ACTIVE seconds (event gaps under 60s) %.0f; of those, %.0f (%.1f%%) fall within TideEase=%.0fs\n"+
		"  of a level change over 0.02 -- i.e. the tide is MOVING for that share of the time he is watching.\n",
		activeSecs, movingSecs, 100*movingSecs/math.Max(1, activeSecs), scape.TideEase)
}

// ---------- main ----------

var widths = []int{40, 60, 80, 111, 128, 153}
var heights = []int{12, 16, 24, 27, 40, 51}

func main() {
	useEnv := flag.Bool("env", false, "measure only the tide setting the environment gives")
	flag.Parse()

	fmt.Println("== s28: the tide's locked guarantees, swept ==")
	fmt.Printf("scape.Tide at startup = %v   TideRange=%.1f TideEase=%.1f  DefaultWriteRows=%d\n\n",
		scape.Tide, scape.TideRange, scape.TideEase, scape.DefaultWriteRows)

	fmt.Println("-- 0. HIS REAL SESSIONS: how big is a real activity step? --")
	levelSteps()
	fmt.Println()

	settings := []bool{true, false}
	if *useEnv {
		settings = []bool{scape.Tide}
	}
	orig := scape.Tide
	defer func() { scape.Tide = orig }()

	type viol struct{ text string }
	var viols []viol

	for _, tide := range settings {
		scape.Tide = tide
		label := "TIDE=1 (default)"
		if !tide {
			label = "TIDE=0 (old fixed waterline)"
		}
		fmt.Printf("================ %s ================\n", label)

		// --- calibration: reproduce the owning tests' own numbers ---
		fmt.Println("-- 1. CALIBRATION against the tests' own geometries --")
		cal := measure(120, 26, 0.7, 0.3868)
		fmt.Printf("  churn  120x26 lv0.70 tod0.3868 COLD (= the test): %.2f%% over %d cells, band rows %d..%d (ceiling 8%%)\n",
			cal.churnCold, cal.cellsCold, cal.bandCold[0], cal.bandCold[1])
		fmt.Printf("         same geometry WARM: %.2f%% / %d cells   STEP 0->0.7: %.2f%% / %d cells\n",
			cal.churnWarm, cal.cellsWarm, cal.churnStep, cal.cellsStep)
		fmt.Printf("  tones  120x26 COLD frame 0 (= the test): worst row %d = %d BG tones (ceiling 40)\n",
			cal.tonesColdRow, cal.tonesCold)
		fmt.Printf("         worst row over 48 warmed/stepped frames: %d tones on row %d; rendered/256 worst %d\n",
			cal.tonesBG, cal.tonesBGRow, cal.tonesRen)
		b, j, ok := teleport(80, 24, 0.30, 0.35)
		if ok {
			fmt.Printf("  telep  80x24 0.30->0.35: baseline %.4f rows, jump %.4f rows = %.1fx (ceiling 8x)\n", b, j, j/b)
		}
		fmt.Println()

		// --- the sweep ---
		fmt.Println("-- 2. SWEEP: churn / tones / clamps --")
		fmt.Println("   chC/chW/chS = churn COLD (the test verbatim) / WARM (tide settled) / STEP (tide moving).")
		fmt.Println("   nC/nS       = cells compared in the COLD and STEP runs. 0 means the band was EMPTY -- a")
		fmt.Println("                 0.00%% with a 0 beside it is a pass on NOTHING, not a pass.")
		fmt.Println("   toneC       = the test's own frame-0 count; toneW = worst of 48 warmed/stepped frames.")
		fmt.Println("   renR        = rows of open sea in the RENDERED frame (horizon to the shallowest waterline).")
		fmt.Printf("%-8s %-5s %-5s %3s %3s %4s %4s %5s %6s %5s %6s %5s %6s %5s %5s %6s %6s %s\n",
			"geom", "lv", "tod", "hy", "sy", "wTop", "renR", "chC%", "nC", "chW%", "nW", "chS%", "nS", "toneC", "toneW", "minE", "maxE", "flags")
		for _, h := range heights {
			for _, w := range widths {
				for _, lv := range []float64{0.0, 0.7, 1.0} {
					for _, tod := range []float64{0.3868, 0.75} {
						r := measure(w, h, lv, tod)
						if r.degenerate {
							fmt.Printf("%-8s %-5.2f %-5.2f  DEGENERATE: Update returned early (W<8 or H<6); lastEdge empty\n",
								fmt.Sprintf("%dx%d", w, h), lv, tod)
							continue
						}
						flags := ""
						where := fmt.Sprintf("%s %dx%d lv%.2f tod%.2f", label, w, h, lv, tod)
						if r.cellsCold == 0 {
							flags += " EMPTY-BAND"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: the churn guarantee's own open-sea band is EMPTY (rows %d..%d, %d cells). "+
									"The test's own metric reports 0.00%% and PASSES ON NOTHING here.",
								where, r.bandCold[0], r.bandCold[1], r.cellsCold)})
						}
						for _, k := range []struct {
							name  string
							pct   float64
							cells int
							band  [2]int
						}{
							{"COLD", r.churnCold, r.cellsCold, r.bandCold},
							{"WARM", r.churnWarm, r.cellsWarm, r.bandWarm},
							{"STEP", r.churnStep, r.cellsStep, r.bandStep},
						} {
							if k.cells > 0 && k.pct > 8 {
								flags += " CHURN>8/" + k.name
								viols = append(viols, viol{fmt.Sprintf(
									"%s: churn %s = %.2f%% of %d open-sea cells (rows %d..%d), ceiling 8%%",
									where, k.name, k.pct, k.cells, k.band[0], k.band[1])})
							}
						}
						if r.tonesCold > 40 {
							flags += " TONEC>40"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: the test's own frame carries %d distinct BG tones on row %d, ceiling 40",
								where, r.tonesCold, r.tonesColdRow)})
						} else if r.tonesBG > 40 {
							flags += " TONEW>40"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: row %d carries %d distinct BG tones at some phase (the test's own frame 0 shows only %d), ceiling 40",
								where, r.tonesBGRow, r.tonesBG, r.tonesCold)})
						}
						if r.minEdge < float64(r.hy) {
							flags += " EDGE<HORIZON"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: waterline reaches row %.2f, ABOVE the horizon at row %d (hyFloor only stops it at %.2f); "+
									"%d rendered rows of open sea", where, r.minEdge, r.hy, r.hyFloor, r.renSeaRows)})
						}
						if r.writeTop < h && r.maxEdge > float64(r.writeTop)-1 {
							flags += " EDGE>WRITETOP"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: waterline reaches %.2f, past writeTop-1 = %d", where, r.maxEdge, r.writeTop-1)})
						}
						if r.renSeaRows <= 0 {
							flags += " NO-SEA-RENDERED"
							viols = append(viols, viol{fmt.Sprintf(
								"%s: the RENDERED frame has %d rows of open sea between the horizon and the shallowest "+
									"waterline -- the sea is gone from the picture", where, r.renSeaRows)})
						}
						fmt.Printf("%-8s %-5.2f %-5.2f %3d %3d %4d %4d %5.2f %6d %5.2f %6d %5.2f %6d %5d %5d %6.2f %6.2f%s\n",
							fmt.Sprintf("%dx%d", w, h), lv, tod, r.hy, r.sy, r.writeTop, r.renSeaRows,
							r.churnCold, r.cellsCold, r.churnWarm, r.cellsWarm, r.churnStep, r.cellsStep,
							r.tonesCold, r.tonesBG, r.minEdge, r.maxEdge, flags)
					}
				}
			}
		}
		fmt.Println()

		// --- teleport across the sweep ---
		fmt.Println("-- 3. SWEEP: teleport (his own 0.30 -> 0.35 step, plus two bigger ones) --")
		fmt.Printf("%-8s %-12s %10s %10s %8s %s\n", "geom", "step", "baseline", "jump", "ratio", "verdict")
		for _, h := range heights {
			for _, w := range widths {
				for _, st := range [][2]float64{{0.30, 0.35}, {0.00, 1.00}, {0.90, 0.30}} {
					b, j, ok := teleport(w, h, st[0], st[1])
					if !ok {
						fmt.Printf("%-8s %-12s  DEGENERATE (no waterline)\n",
							fmt.Sprintf("%dx%d", w, h), fmt.Sprintf("%.2f->%.2f", st[0], st[1]))
						continue
					}
					ratio := math.Inf(1)
					if b > 0 {
						ratio = j / b
					}
					verdict := "ok"
					if ratio > 8 {
						verdict = "OVER 8x"
						if st == [2]float64{0.30, 0.35} {
							viols = append(viols, viol{fmt.Sprintf(
								"%s %dx%d: the LOCKED 0.30->0.35 step moved the waterline %.4f rows against a %.4f baseline = %.1fx, ceiling 8x",
								label, w, h, j, b, ratio)})
						}
					}
					if st != [2]float64{0.30, 0.35} && verdict == "OVER 8x" {
						verdict = "over 8x (NOT the locked step)"
					}
					fmt.Printf("%-8s %-12s %10.5f %10.5f %8.1f %s\n",
						fmt.Sprintf("%dx%d", w, h), fmt.Sprintf("%.2f->%.2f", st[0], st[1]),
						b, j, ratio, verdict)
				}
			}
		}
		fmt.Println()
	}

	// ---- 4. ISOLATION: which term drives the water above the horizon? ----
	scape.Tide = true
	origRange := scape.TideRange
	fmt.Println("================ 4. ISOLATION: vary ONE term ================")
	fmt.Println("TideRange is the only tide term that is exported and therefore the only one that can be")
	fmt.Println("varied without editing product source. TideRange=0 leaves the wash and the two x-sines")
	fmt.Println("intact and removes the withdrawal entirely, so it separates the two.")
	fmt.Printf("%-9s %-5s %-6s %4s %8s %8s %6s %6s\n", "geom", "lv", "TidRng", "hy", "minEdge", "maxEdge", "renR", "coldCells")
	for _, g := range [][2]int{{40, 12}, {80, 12}, {111, 12}, {80, 24}, {111, 27}, {153, 51}} {
		for _, lv := range []float64{0.0, 0.7} {
			for _, tr := range []float64{5, 4, 3, 2, 1, 0} {
				scape.TideRange = tr
				r := measure(g[0], g[1], lv, 0.3868)
				mark := ""
				if r.minEdge < float64(r.hy) {
					mark = "  <- above the horizon"
				}
				if r.renSeaRows <= 0 {
					mark += "  <- no sea rendered"
				}
				fmt.Printf("%-9s %-5.2f %-6.1f %4d %8.2f %8.2f %6d %9d%s\n",
					fmt.Sprintf("%dx%d", g[0], g[1]), lv, tr, r.hy, r.minEdge, r.maxEdge, r.renSeaRows, r.cellsCold, mark)
			}
			fmt.Println()
		}
	}
	scape.TideRange = origRange
	scape.Tide = orig

	// ---- 5. WHICH FRAMES CHURN, and what is different about them ----
	//
	// The churn readings above are an average over 23 frame-pairs. An average
	// cannot tell a sea that shimmers everywhere from one that is perfectly
	// still except for four frames that repaint whole. This prints the churn of
	// EVERY pair beside the shore's own tideRow, which is the integer the open
	// sea's depth ramp is anchored to (shore.go:547, `depth = max(1, tideRow -
	// hy)` under Tide).
	scape.Tide = true
	fmt.Println("================ 5. PER-FRAME: where the churn actually is ================")
	for _, g := range [][2]int{{120, 26}, {80, 24}, {153, 51}} {
		w, h := g[0], g[1]
		sh := scape.NewShore(7, false)
		cs := newCanvas(w, h)
		tm := 0.0
		for i := 0; i < warmFrames; i++ {
			tm += frameDT
			sh.Update(cs, tm, scape.Activity{Level: 0, TimeOfDay: 0.3868})
		}
		fmt.Printf("-- %dx%d, settled at rest (tideRow %d), then Level steps 0 -> 0.7 --\n", w, h, tideRowOf(sh))
		act := scape.Activity{Level: 0.7, TimeOfDay: 0.3868}
		var fr []*canvas.Canvas
		var rows []int
		var ats []float64
		for i := 0; i < 40; i++ {
			cc := newCanvas(w, h)
			tm += frameDT
			sh.Update(cc, tm, act)
			fr = append(fr, cc)
			rows = append(rows, tideRowOf(sh))
			ats = append(ats, tideAtOf(sh))
		}
		// The band is taken from the FINAL frame, which is what the guarantee
		// itself does (openSea(sh, fr[0]) reads sh.lastEdge, i.e. the last
		// Update). Taking it from the resting frame gives an empty band and a
		// column of 0.00%s that mean nothing.
		from, to := openSea(edgeOf(sh), h)
		nz, big := 0, 0
		for i := 1; i < len(fr); i++ {
			ch, tot := 0, 0
			for y := from; y <= to; y++ {
				if y < 0 || y >= h {
					continue
				}
				for x := 0; x < w; x++ {
					tot++
					if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
						ch++
					}
				}
			}
			pct := 0.0
			if tot > 0 {
				pct = 100 * float64(ch) / float64(tot)
			}
			// Split the churn: the BOTTOM row of the band is the one the
			// advancing waterline sweeps through, the rows above it are pure
			// depth ramp and can only change if the ramp itself is re-anchored.
			topCh, topTot := 0, 0
			for y := from; y < to; y++ {
				if y < 0 || y >= h {
					continue
				}
				for x := 0; x < w; x++ {
					topTot++
					if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
						topCh++
					}
				}
			}
			topPct := 0.0
			if topTot > 0 {
				topPct = 100 * float64(topCh) / float64(topTot)
			}
			// HOW BIG is the change, not just how many cells. A cell counts as
			// "changed" if one channel moved by one. A whole-sea repaint that
			// shifts every tone by 1 is invisible; one that shifts it by 20 is
			// a flash. Measured on the 256-cube-quantised background, which is
			// what the terminal is actually sent.
			meanD, maxD := 0.0, 0.0
			nD := 0
			for y := from; y <= to; y++ {
				if y < 0 || y >= h {
					continue
				}
				for x := 0; x < w; x++ {
					_, _, a := fr[i-1].ResolveAt(x, y, term.Profile256)
					_, _, b := fr[i].ResolveAt(x, y, term.Profile256)
					d := math.Abs(0.299*(float64(a.R)-float64(b.R))) +
						math.Abs(0.587*(float64(a.G)-float64(b.G))) +
						math.Abs(0.114*(float64(a.B)-float64(b.B)))
					meanD += d
					if d > maxD {
						maxD = d
					}
					nD++
				}
			}
			if nD > 0 {
				meanD /= float64(nD)
			}
			note := ""
			if rows[i] != rows[i-1] {
				note = fmt.Sprintf("  <- tideRow %d -> %d", rows[i-1], rows[i])
			}
			if pct > 0 {
				nz++
			}
			if pct > 50 {
				big++
			}
			if i <= 34 {
				fmt.Printf("   pair %2d: %6.2f%% of %d cells (above the last row: %6.2f%%)  dLuma mean %5.2f max %5.1f  tideAt %5.2f  tideRow %d%s\n",
					i, pct, tot, topPct, meanD, maxD, ats[i], rows[i], note)
			}
		}
		fmt.Printf("   -> band rows %d..%d; %d of %d pairs moved at all; %d moved more than half the open sea.\n\n",
			from, to, nz, len(fr)-1, big)
	}
	scape.Tide = orig

	// ---- 6. ISOLATE THE STEP CHURN by varying TideEase ----
	//
	// If the over-ceiling churn is the row-quantised tideRow crossing integer
	// boundaries, then slowing the tide down -- which puts FEWER crossings in
	// the same 24-frame window and changes nothing else -- must drop the churn
	// in proportion. TideEase is exported, so this varies exactly one term.
	scape.Tide = true
	origEase := scape.TideEase
	fmt.Println("================ 6. ISOLATION: vary TideEase (the only other exported tide term) ================")
	fmt.Printf("%-9s %-9s %10s %10s\n", "geom", "TideEase", "STEP churn", "cells")
	for _, g := range [][2]int{{120, 26}, {80, 24}, {153, 51}} {
		for _, ease := range []float64{3, 10, 30, 100, 300} {
			scape.TideEase = ease
			r := measure(g[0], g[1], 0.7, 0.3868)
			fmt.Printf("%-9s %-9.0f %9.2f%% %10d\n",
				fmt.Sprintf("%dx%d", g[0], g[1]), ease, r.churnStep, r.cellsStep)
		}
		fmt.Println()
	}
	scape.TideEase = origEase
	scape.Tide = orig

	fmt.Println("================ VIOLATIONS ================")
	if len(viols) == 0 {
		fmt.Println("none")
	}
	for i, v := range viols {
		fmt.Printf("%3d. %s\n", i+1, v.text)
	}
	fmt.Printf("\n%d violations over %d geometries x 3 levels x 2 times of day x %d tide settings\n",
		len(viols), len(widths)*len(heights), len(settings))
}
