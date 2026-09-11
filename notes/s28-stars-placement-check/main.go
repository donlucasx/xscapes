// s28-stars-placement-check is an INDEPENDENT audit of the constellation's
// placement. It shares no helper with internal/scape's own tests on purpose:
// every star it counts is read back off a RENDERED frame through
// canvas.ResolveAt on Profile256, which is the picture the terminal draws.
//
// It runs identically against HEAD (the golden-ratio layout) and against the
// working tree (the dart-thrown one), so every number below is a comparison
// made by one instrument rather than two reports.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

type geom struct{ w, h int }

// warm builds ONE shore and drives it for `n` updates before anything is
// sampled. A fresh Shore per frame has no history and Update clamps a >1s gap
// to a single nominal step, so a cold frame's sea has not moved -- the trap the
// brief names. Everything here samples the LAST frame of a warmed run.
func warm(g geom, seed int64, tod, ctx, level float64, done, total, n int) (*canvas.Canvas, *scape.Shore) {
	c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28 // production, mirrored composition (compose() in live.go)
	act := scape.Activity{Working: true, Level: level, TimeOfDay: tod,
		ContextUsed: ctx, TodoDone: done, TodoTotal: total}
	t := 0.0
	for i := 0; i < n; i++ {
		t += 0.1
		sh.Update(c, t, act)
	}
	return c, sh
}

// starsOf reads the constellation off the rendered frame. It scans EVERY row,
// not just the sky, so a star drawn into the sea is counted rather than missed.
// '*' belongs to the constellation alone in non-ASCII mode -- the ambient field
// is '.', '·' and '+'.
func starsOf(c *canvas.Canvas) [][2]int {
	var out [][2]int
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if ch, _, _ := c.ResolveAt(x, y, term.Profile256); ch == '*' {
				out = append(out, [2]int{x, y})
			}
		}
	}
	return out
}

func key(p [2]int) int { return p[1]*10000 + p[0] }

// colGapSpread is sd/mean over the gaps between successive star COLUMNS, the
// same statistic the brief quotes: 0.31-0.43 today, ~1.00 for a Poisson sky.
func colGapSpread(pts [][2]int) float64 {
	if len(pts) < 3 {
		return math.NaN()
	}
	xs := make([]float64, len(pts))
	for i, p := range pts {
		xs[i] = float64(p[0])
	}
	sort.Float64s(xs)
	gaps := make([]float64, 0, len(xs)-1)
	for i := 1; i < len(xs); i++ {
		gaps = append(gaps, xs[i]-xs[i-1])
	}
	mean := 0.0
	for _, g := range gaps {
		mean += g
	}
	mean /= float64(len(gaps))
	if mean == 0 {
		return math.NaN()
	}
	v := 0.0
	for _, g := range gaps {
		v += (g - mean) * (g - mean)
	}
	return math.Sqrt(v/float64(len(gaps))) / mean
}

// closest returns, over all pairs: the smallest raw-cell Euclidean distance,
// the smallest SCREEN-unit distance (a row is worth two columns), and how many
// pairs are touching -- |dx|<=1 and |dy|<=1, which is the pair a human reads as
// one star and therefore miscounts.
func closest(pts [][2]int) (raw, screen float64, touching int) {
	raw, screen = math.MaxFloat64, math.MaxFloat64
	for i := 0; i < len(pts); i++ {
		for j := i + 1; j < len(pts); j++ {
			dx := float64(pts[i][0] - pts[j][0])
			dy := float64(pts[i][1] - pts[j][1])
			if d := math.Hypot(dx, dy); d < raw {
				raw = d
			}
			if d := math.Hypot(dx, dy*2); d < screen {
				screen = d
			}
			if math.Abs(dx) <= 1 && math.Abs(dy) <= 1 {
				touching++
			}
		}
	}
	return
}

// readoutCells reproduces drawReadout (live.go:331-361) from the outside. It
// is a MIRROR of code this audit does not own, written from that source, so it
// answers "does the number land on a star" without trusting the layout's own
// copy of the same arithmetic.
func readoutCells(c *canvas.Canvas, sh *scape.Shore, used float64) [][2]int {
	if used < 0.40 {
		return nil
	}
	mx, my := sh.MoonPos()
	rx, ry := sh.MoonExtent()
	txt := fmt.Sprintf("%.0f%%", (1-used)*100)
	if used >= 0.85 {
		txt += " left"
	}
	hy := int(float64(c.H)*0.42) + 1
	x, y := mx-len(txt)/2, my+ry+1
	if y > hy-1 {
		y = my
		x = mx + rx + 2
		if x+len(txt) > c.W-1 {
			x = mx - rx - 1 - len(txt)
		}
	}
	if x < 1 {
		x = 1
	}
	if x+len(txt) > c.W-1 {
		x = c.W - 1 - len(txt)
	}
	out := make([][2]int, 0, len(txt))
	for i := range txt {
		out = append(out, [2]int{x + i, y})
	}
	return out
}

func picture(c *canvas.Canvas, sh *scape.Shore, rows int) string {
	var b strings.Builder
	b.WriteString("        +" + strings.Repeat("-", c.W) + "+\n")
	for y := 0; y < rows && y < c.H; y++ {
		b.WriteString("        |")
		for x := 0; x < c.W; x++ {
			ch, _, _ := c.ResolveAt(x, y, term.Profile256)
			switch {
			case ch == '*':
				b.WriteByte('*')
			case sh.DiscCovers(x, y):
				b.WriteByte('o')
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteString("|\n")
	}
	b.WriteString("        +" + strings.Repeat("-", c.W) + "+\n")
	return b.String()
}

func hy(c *canvas.Canvas) int { return int(float64(c.H) * 0.42) }

func main() {
	which := flag.String("run", "all", "all|locked|spread|picture|edges")
	flag.Parse()
	run := func(name string) bool { return *which == "all" || *which == name }

	if run("locked") {
		lockedGuarantee()
	}
	if run("spread") {
		spread()
	}
	if run("picture") {
		pictures()
	}
	if run("edges") {
		edges()
	}
	if run("hold") {
		hold()
	}
	if run("cost") {
		cost()
	}
}

// cost is the check the cache invites. starPlaces caches on a key that
// includes the moon's radius, and the fallback it guards (roomiestCell) is
// exhaustive -- every cell of the band against every star placed. If the key
// ever misses per frame, that fallback runs per frame, and 40x12 is where it
// runs at all. Time the WHOLE scape through a moving session, not the layout
// alone, because a per-frame recompute is only worth knowing about as a share
// of the frame.
func cost() {
	fmt.Println("== COST: whole-scape Update, a moving session (context and clock advancing) ==")
	for _, g := range []geom{{125, 28}, {153, 51}, {125, 62}, {40, 12}} {
		c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		t := 0.0
		for i := 0; i < 120; i++ { // warm
			t += 0.1
			sh.Update(c, t, scape.Activity{Working: true, Level: 0.5, TodoDone: 32, TodoTotal: 32})
		}
		const n = 600
		start := nowMs()
		for i := 0; i < n; i++ {
			t += 0.1
			sh.Update(c, t, scape.Activity{Working: true, Level: 0.5,
				TimeOfDay:   math.Mod(float64(i)*0.0017, 1),
				ContextUsed: float64(i) / float64(n), TodoDone: 32, TodoTotal: 32})
		}
		fmt.Printf("   %dx%d: %.3f ms a frame over %d frames\n", g.w, g.h, (nowMs()-start)/float64(n), n)
	}
	fmt.Println()
}

// hold is the risk the new cache key raises. starPlaces caches on a starKey
// that includes the moon's column and RADIUS -- so if either moves during a
// session, the darts are re-thrown against a different set of cells and every
// star lands somewhere new. That would be his complaint exactly, arriving
// through the fix for it. Sweep the three things that change inside one
// session -- context, the clock, the activity -- on ONE shore, and compare
// every frame's constellation with the first.
func hold() {
	fmt.Println("== HELD STILL: does anything that moves in a session move the stars? ==")
	fmt.Println("   one warmed shore per geometry, context 0.00-1.00, the clock round, activity 0-1")
	for _, g := range []geom{{125, 28}, {153, 51}, {125, 62}, {40, 12}, {124, 22}, {80, 24}} {
		for _, seed := range []int64{1, 7} {
			c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh := scape.NewShore(seed, false)
			sh.MoonX = 0.28
			var first []string
			frames, differ := 0, 0
			firstDiff := ""
			t := 0.0
			for step := 0; step <= 100; step++ {
				ctx := float64(step) / 100
				act := scape.Activity{Working: true, Level: float64(step%11) / 10,
					TimeOfDay: math.Mod(float64(step)*0.037, 1), ContextUsed: ctx,
					TodoDone: 32, TodoTotal: 32}
				// Warm inside the sweep: this is ONE shore carried through a
				// whole session, not a fresh one per sample.
				for i := 0; i < 24; i++ {
					t += 0.1
					sh.Update(c, t, act)
				}
				pts := starsOf(c)
				cur := make([]string, 0, len(pts))
				for _, p := range pts {
					cur = append(cur, fmt.Sprintf("%d,%d", p[0], p[1]))
				}
				sort.Strings(cur)
				frames++
				if first == nil {
					first = cur
					continue
				}
				if strings.Join(cur, " ") != strings.Join(first, " ") {
					differ++
					if firstDiff == "" {
						firstDiff = fmt.Sprintf("ctx %.2f: %d stars, was %d", ctx, len(cur), len(first))
					}
				}
			}
			fmt.Printf("   %dx%d seed %d: %d frames, %d with a different constellation %s\n",
				g.w, g.h, seed, frames, differ, firstDiff)
		}
	}
	fmt.Println()
}

// ---------------------------------------------------------------- the locked one

// "Position is fixed by index and seed so a star lights where it always was."
// Light them one at a time from 1 to 32 and check that the set only ever GROWS:
// every star present at k is in the same cell at k+1, and exactly one is added.
func lockedGuarantee() {
	fmt.Println("== LOCKED GUARANTEE: a star lights where it always was ==")
	fmt.Println("   done 1..32 one at a time, read off rendered frames; a violation is any")
	fmt.Println("   earlier star that MOVED or went OUT, or a frame whose count != done.")
	geoms := []geom{{125, 28}, {153, 51}, {80, 24}, {143, 27}, {124, 22}, {125, 62}, {40, 12}, {111, 25}}
	seeds := []int64{1, 7, 42}
	moved, missing, frames := 0, 0, 0
	worst := ""
	for _, g := range geoms {
		for _, seed := range seeds {
			prev := map[int]bool{}
			for done := 1; done <= 32; done++ {
				c, _ := warm(g, seed, 0.0, 0.3, 0.5, done, 32, 240)
				pts := starsOf(c)
				frames++
				cur := map[int]bool{}
				for _, p := range pts {
					cur[key(p)] = true
				}
				for k := range prev {
					if !cur[k] {
						moved++
						if worst == "" {
							worst = fmt.Sprintf("%dx%d seed %d: a star present at done=%d was gone at done=%d", g.w, g.h, seed, done-1, done)
						}
					}
				}
				if len(pts) != done {
					missing++
					if worst == "" {
						worst = fmt.Sprintf("%dx%d seed %d: done=%d drew %d stars", g.w, g.h, seed, done, len(pts))
					}
				}
				prev = cur
			}
		}
	}
	fmt.Printf("   %d frames (%d geometries x %d seeds x 32 counts), all warmed 240 updates\n", frames, len(geoms), len(seeds))
	fmt.Printf("   stars that moved or went out: %d\n", moved)
	fmt.Printf("   frames whose count != done:   %d\n", missing)
	if worst != "" {
		fmt.Printf("   first violation: %s\n", worst)
	}
	fmt.Println()

	// The same guarantee with the LIST ITSELF growing -- done == total, which
	// is what an agent ticking off a list it is still adding to looks like.
	// reduce holds TodoTotal at StarsCap today, so this is the stronger form
	// rather than the live one, and it is the one the layout's own design
	// claims: a slot depends only on the slots below it.
	fmt.Println("   -- and with the LIST growing too (done == total, 1..32) --")
	moved, missing, frames = 0, 0, 0
	worst = ""
	for _, g := range geoms {
		for _, seed := range seeds {
			prev := map[int]bool{}
			for n := 1; n <= 32; n++ {
				c, _ := warm(g, seed, 0.0, 0.3, 0.5, n, n, 240)
				pts := starsOf(c)
				frames++
				cur := map[int]bool{}
				for _, p := range pts {
					cur[key(p)] = true
				}
				for k := range prev {
					if !cur[k] {
						moved++
						if worst == "" {
							worst = fmt.Sprintf("%dx%d seed %d: a star at %d of %d was gone at %d of %d", g.w, g.h, seed, n-1, n-1, n, n)
						}
					}
				}
				if len(pts) != n {
					missing++
					if worst == "" {
						worst = fmt.Sprintf("%dx%d seed %d: %d of %d drew %d stars", g.w, g.h, seed, n, n, len(pts))
					}
				}
				prev = cur
			}
		}
	}
	fmt.Printf("   %d frames | stars that moved or went out: %d | count != done: %d\n", frames, moved, missing)
	if worst != "" {
		fmt.Printf("   first violation: %s\n", worst)
	}
	fmt.Println()
}

// ---------------------------------------------------------------- the spread

func spread() {
	fmt.Println("== SPREAD: is it less regular, and did it introduce collisions? ==")
	fmt.Println("   sd/mean over gaps between star COLUMNS. ~0.31-0.43 = the golden-ratio ruler,")
	fmt.Println("   ~1.00 = a Poisson sky. touching = pairs with |dx|<=1 and |dy|<=1 (reads as one star).")
	geoms := []geom{{125, 28}, {153, 51}, {80, 24}, {143, 27}, {125, 62}, {40, 12}}
	seeds := []int64{1, 3, 7, 11, 42}
	counts := []int{12, 19, 26, 32}
	for _, g := range geoms {
		fmt.Printf("   %dx%d\n", g.w, g.h)
		for _, n := range counts {
			var sds []float64
			minRaw, minScreen, touch := math.MaxFloat64, math.MaxFloat64, 0
			lit := 0
			for _, seed := range seeds {
				c, _ := warm(g, seed, 0.0, 0.3, 0.5, n, 32, 240)
				pts := starsOf(c)
				lit += len(pts)
				sds = append(sds, colGapSpread(pts))
				r, s, tc := closest(pts)
				if r < minRaw {
					minRaw = r
				}
				if s < minScreen {
					minScreen = s
				}
				touch += tc
			}
			lo, hi := math.MaxFloat64, -1.0
			for _, v := range sds {
				if v < lo {
					lo = v
				}
				if v > hi {
					hi = v
				}
			}
			fmt.Printf("     %2d stars: sd/mean %.2f-%.2f | closest raw %.2f cells, %.2f screen | touching pairs %d | lit %d/%d\n",
				n, lo, hi, minRaw, minScreen, touch, lit, n*len(seeds))
		}
	}
	fmt.Println()
}

// ---------------------------------------------------------------- the picture

func pictures() {
	fmt.Println("== THE SKY, AS RENDERED (o = the disc) ==")
	for _, tc := range []struct {
		g    geom
		seed int64
		n    int
	}{
		{geom{125, 28}, 7, 12}, {geom{125, 28}, 7, 26}, {geom{125, 28}, 3, 32},
		{geom{125, 28}, 42, 32}, {geom{153, 51}, 7, 32},
		{geom{40, 12}, 7, 12}, {geom{40, 12}, 7, 32}, {geom{125, 62}, 7, 32},
	} {
		c, sh := warm(tc.g, tc.seed, 0.0, 0.3, 0.5, tc.n, 32, 240)
		pts := starsOf(c)
		rows := hy(c) + 1
		fmt.Printf("   %dx%d seed %d, %d stars, sky rows 0..%d, sd/mean %.2f\n",
			tc.g.w, tc.g.h, tc.seed, tc.n, hy(c)-1, colGapSpread(pts))
		lo, hiRow := 99, -1
		for _, p := range pts {
			if p[1] < lo {
				lo = p[1]
			}
			if p[1] > hiRow {
				hiRow = p[1]
			}
		}
		fmt.Printf("   rows used: %d..%d of %d sky rows\n", lo, hiRow, hy(c))
		fmt.Print(picture(c, sh, rows))
		fmt.Println()
	}
}

// ---------------------------------------------------------------- the edges

// The design floor (40x12) and his tall window (125x62): does anything fall
// off, land on the disc, sink into the water, or get eaten by the readout, at
// any hour, any context, any activity?
func edges() {
	fmt.Println("== EDGES: 40x12 (design floor) and 125x62 (his tall window) ==")
	geoms := []geom{{40, 12}, {125, 62}, {40, 13}, {30, 8}, {80, 24}, {125, 28}}
	seeds := []int64{1, 7, 42}
	for _, g := range geoms {
		frames, lost, onDisc, belowHorizon, onReadout := 0, 0, 0, 0, 0
		firstLoss := ""
		bandLo, bandHi := 99, -1
		for _, seed := range seeds {
			for _, tod := range []float64{0.0, 0.25, 0.5, 0.6, 0.72, 0.875} {
				for _, ctxStep := range []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100} {
					ctx := float64(ctxStep) / 100
					for _, level := range []float64{0.0, 0.5, 1.0} {
						c, sh := warm(g, seed, tod, ctx, level, 32, 32, 240)
						pts := starsOf(c)
						frames++
						if len(pts) != 32 {
							lost++
							if firstLoss == "" {
								firstLoss = fmt.Sprintf("%dx%d seed %d tod %.2f ctx %.2f lvl %.1f: %d of 32 stars on screen",
									g.w, g.h, seed, tod, ctx, level, len(pts))
							}
						}
						rc := readoutCells(c, sh, ctx)
						rset := map[int]bool{}
						for _, p := range rc {
							rset[key(p)] = true
						}
						for _, p := range pts {
							if sh.DiscCovers(p[0], p[1]) {
								onDisc++
							}
							if p[1] >= hy(c) {
								belowHorizon++
							}
							if rset[key(p)] {
								onReadout++
							}
							if p[1] < bandLo {
								bandLo = p[1]
							}
							if p[1] > bandHi {
								bandHi = p[1]
							}
						}
					}
				}
			}
		}
		fmt.Printf("   %dx%d: sky is %d rows | band used %d..%d | %d frames, all with done=32\n",
			g.w, g.h, hy(canvas.New(g.w, g.h, 1, 1, 1)), bandLo, bandHi, frames)
		fmt.Printf("     frames not showing all 32: %d | stars on the disc: %d | below the horizon: %d | under the readout: %d\n",
			lost, onDisc, belowHorizon, onReadout)
		if firstLoss != "" {
			fmt.Printf("     first: %s\n", firstLoss)
		}
	}
	fmt.Println()
	if os.Getenv("XSCAPES_TIDE") != "" {
		fmt.Printf("   (XSCAPES_TIDE was set in the environment for this run)\n")
	}
}

func nowMs() float64 { return float64(time.Now().UnixNano()) / 1e6 }
