// Command s28-stars-fun-check is an INDEPENDENT check of the constellation as
// it stands in the working tree. It shares no code with either agent's study
// or with internal/scape's own tests: every number below is read back off a
// RENDERED frame through canvas.ResolveAt on the 256 profile.
//
// Two traps this is built to avoid, both of which have voided instruments here
// before:
//
//   - A fresh NewShore per frame has NO history (Update clamps a >1 s gap to a
//     single nominal step), so an instrument built on one measures the first
//     frame of a session forever. Every frame below is one shore warmed with
//     300 Updates.
//   - A clean result from an empty window looks exactly like a pass. Every
//     section prints the star COUNT it actually found, so an empty frame is
//     visible as a zero rather than hiding as a pass.
package main

import (
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

const warmFrames = 300

// warm turns one shore for warmFrames updates and returns the last canvas.
func warm(w, h int, seed int64, done int, tod, ctx float64) (*scape.Shore, *canvas.Canvas) {
	sh := scape.NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < warmFrames; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Working: true, Level: 0.5,
			TimeOfDay: tod, ContextUsed: ctx, TodoDone: done, TodoTotal: 32})
	}
	return sh, c
}

// lit scans a rendered frame for the constellation. '*' belongs to the
// checklist alone -- the ambient field is '.', U+00B7 and '+' -- so a scan for
// '*' is the count a reader would make.
func lit(c *canvas.Canvas) [][2]int {
	var at [][2]int
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				at = append(at, [2]int{x, y})
			}
		}
	}
	sort.Slice(at, func(i, j int) bool {
		if at[i][0] != at[j][0] {
			return at[i][0] < at[j][0]
		}
		return at[i][1] < at[j][1]
	})
	return at
}

func key(p [2]int) [2]int { return p }

func main() {
	fmt.Println("STARS -- an independent check of the constellation in the WORKING TREE.")
	fmt.Printf("Every frame is ONE shore warmed with %d Updates; every glyph is read back\n", warmFrames)
	fmt.Println("with canvas.ResolveAt(x, y, term.Profile256). Nothing is taken from source.")
	fmt.Println()

	checkGuarantee()
	checkSpacing()
	checkGeometries()
	showSky()

	if failed {
		fmt.Println("\n==== RESULT: AT LEAST ONE CHECK FAILED ====")
		os.Exit(0) // a study, not a gate -- read the sections
	}
	fmt.Println("\n==== RESULT: every check above passed ====")
}

var failed bool

func fail(format string, a ...any) {
	failed = true
	fmt.Printf("  FAIL "+format+"\n", a...)
}

// ---------------------------------------------------------------------------
// 1. THE LOCKED GUARANTEE
//
// CLAUDE.md: "Position is fixed by index and seed so a star lights where it
// always was." Checked the way a reader would check it and NOT the way the
// code is organised: light the stars one at a time from 1 to 32 and require
// that the set of lit cells only ever GROWS. If any earlier star moved, a cell
// that was lit at n is dark at n+1, and the set is not a superset.
//
// Run two ways, because they can differ and only one of them is his:
//
//	FRESH   -- a new warmed shore for every count (what a test usually does)
//	LIVE    -- ONE shore whose checklist grows under it, which is the real
//	           session: the layout is cached, and a cache is where a
//	           "depends only on the index" guarantee goes to die.
// ---------------------------------------------------------------------------

func checkGuarantee() {
	fmt.Println("==== 1. THE LOCKED GUARANTEE: a star lights where it always was ====")
	fmt.Println("  Lighting 1..32 one at a time; the lit set must only GROW.")
	fmt.Println()

	type g struct{ w, h int }
	geos := []g{{40, 12}, {80, 24}, {125, 28}, {143, 27}, {153, 51}, {125, 62}}
	seeds := []int64{5, 7, 11, 23, 42}

	for _, gg := range geos {
		for _, seed := range seeds {
			// ---- FRESH: a new warmed shore per count.
			prev := map[[2]int]bool{}
			prevN := 0
			moves := 0
			var firstMove string
			for done := 1; done <= 32; done++ {
				_, c := warm(gg.w, gg.h, seed, done, 20.0/24, 0.30)
				at := lit(c)
				now := map[[2]int]bool{}
				for _, p := range at {
					now[key(p)] = true
				}
				for p := range prev {
					if !now[p] {
						moves++
						if firstMove == "" {
							firstMove = fmt.Sprintf("at %d->%d the cell [%d %d] went dark", prevN, done, p[0], p[1])
						}
					}
				}
				prev, prevN = now, done
			}
			if moves > 0 {
				fail("%dx%d seed %d FRESH: %d star-cells vanished as later stars lit -- %s",
					gg.w, gg.h, seed, moves, firstMove)
			}

			// ---- LIVE: one shore, the checklist growing under it.
			sh := scape.NewShore(seed, false)
			tm := 0.0
			step := func(done int) [][2]int {
				var c *canvas.Canvas
				for k := 0; k < warmFrames; k++ {
					c = canvas.New(gg.w, gg.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					tm += 0.08
					sh.Update(c, tm, scape.Activity{Working: true, Level: 0.5,
						TimeOfDay: 20.0 / 24, ContextUsed: 0.30, TodoDone: done, TodoTotal: 32})
				}
				return lit(c)
			}
			prev = map[[2]int]bool{}
			lmoves := 0
			var lFirst string
			for done := 1; done <= 32; done++ {
				now := map[[2]int]bool{}
				for _, p := range step(done) {
					now[key(p)] = true
				}
				for p := range prev {
					if !now[p] {
						lmoves++
						if lFirst == "" {
							lFirst = fmt.Sprintf("lighting the %dth star put out the cell [%d %d]", done, p[0], p[1])
						}
					}
				}
				prev = now
			}
			if lmoves > 0 {
				fail("%dx%d seed %d LIVE: %d star-cells vanished -- %s", gg.w, gg.h, seed, lmoves, lFirst)
			}
		}
	}
	fmt.Printf("  %d geometries x %d seeds x 32 counts, both ways = %d frames.\n",
		len(geos), len(seeds), len(geos)*len(seeds)*32*2)
	if !failed {
		fmt.Println("  ⇒ No earlier star ever moved. The locked guarantee HOLDS.")
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 2. SPACING -- is it less regular, and did that cost collisions?
//
// The baseline handed to me: today's golden ratio measures gap sd/mean 0.31 at
// twelve stars, 0.31 at nineteen, 0.43 at twenty-six, with ZERO pairs closer
// than two cells. A Poisson sky is about 1.00. Both halves have to improve or
// the change is not worth having: more scatter AND no collisions.
//
// Two measures, because the brief's number is one-dimensional and the sky is
// not:
//
//	COLUMN GAPS  -- sort the columns, take consecutive gaps. This is the
//	                measure the 0.31/0.43 baseline was taken with, so it is
//	                the only one that compares.
//	NEAREST NBR  -- distance to the closest other star in SCREEN units (a row
//	                counts double, a cell being about twice as tall as wide).
//	                This is what decides whether two marks read as one.
// ---------------------------------------------------------------------------

func sdOverMean(v []float64) (float64, float64, float64) {
	if len(v) == 0 {
		return 0, 0, 0
	}
	var s float64
	for _, x := range v {
		s += x
	}
	mean := s / float64(len(v))
	var q float64
	for _, x := range v {
		q += (x - mean) * (x - mean)
	}
	sd := math.Sqrt(q / float64(len(v)))
	if mean == 0 {
		return mean, sd, 0
	}
	return mean, sd, sd / mean
}

func checkSpacing() {
	fmt.Println("==== 2. SPACING -- less regular, and at what cost? ====")
	fmt.Println("  Baseline handed to me (HEAD's golden ratio, 125x28): sd/mean 0.31 / 0.31 / 0.43")
	fmt.Println("  at 12 / 19 / 26 stars, and ZERO pairs closer than 2 cells.")
	fmt.Println()
	fmt.Println("  seed  stars  drawn   col-gap sd/mean   nearest-nbr (screen units)   pairs <2 cells")

	for _, seed := range []int64{5, 7, 11, 23, 42} {
		for _, done := range []int{12, 19, 26} {
			_, c := warm(125, 28, seed, done, 20.0/24, 0.30)
			at := lit(c)

			cols := make([]float64, 0, len(at))
			for _, p := range at {
				cols = append(cols, float64(p[0]))
			}
			sort.Float64s(cols)
			gaps := make([]float64, 0, len(cols))
			for i := 1; i < len(cols); i++ {
				gaps = append(gaps, cols[i]-cols[i-1])
			}
			_, _, ratio := sdOverMean(gaps)

			// nearest neighbour, and collisions in plain cell distance
			minNN := math.MaxFloat64
			close2 := 0
			for i := range at {
				nn := math.MaxFloat64
				for j := range at {
					if i == j {
						continue
					}
					dx := float64(at[i][0] - at[j][0])
					dy := float64(at[i][1]-at[j][1]) * 2.0
					d := math.Hypot(dx, dy)
					if d < nn {
						nn = d
					}
					cd := math.Hypot(float64(at[i][0]-at[j][0]), float64(at[i][1]-at[j][1]))
					if i < j && cd < 2.0 {
						close2++
					}
				}
				if nn < minNN {
					minNN = nn
				}
			}
			if len(at) < 2 {
				minNN = 0
			}
			flagged := ""
			if close2 > 0 {
				flagged = "  <-- COLLISION"
				failed = true
			}
			fmt.Printf("  %4d  %5d  %5d   %13.2f   %20.2f (min)   %8d%s\n",
				seed, done, len(at), ratio, minNN, close2, flagged)
		}
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 3. THE GEOMETRIES -- the design floor and his tall window.
//
// "Does anything fall off, overlap the disc, or vanish?" -- counted, not
// asserted. Expected is min(done, 32); anything less has been eaten.
// ---------------------------------------------------------------------------

func checkGeometries() {
	fmt.Println("==== 3. GEOMETRIES -- does anything vanish, or land on the disc? ====")
	fmt.Println("  Expected = the number earned. 'eaten' = earned minus drawn.")
	fmt.Println()
	fmt.Println("  geometry   hour   earned  drawn  eaten  on disc  rows used   band     closest pair (cells)")

	type g struct{ w, h int }
	for _, gg := range []g{{40, 12}, {80, 24}, {125, 28}, {125, 62}, {153, 51}} {
		for _, hour := range []float64{2, 8, 13.5, 20} {
			for _, done := range []int{12, 32} {
				sh, c := warm(gg.w, gg.h, int64(gg.h%7+5), done, hour/24, 0.30)
				at := lit(c)
				onDisc := 0
				for _, p := range at {
					if sh.DiscCovers(p[0], p[1]) {
						onDisc++
					}
				}
				lo, hi := 1<<30, -1
				for _, p := range at {
					if p[1] < lo {
						lo = p[1]
					}
					if p[1] > hi {
						hi = p[1]
					}
				}
				rows := 0
				seen := map[int]bool{}
				for _, p := range at {
					if !seen[p[1]] {
						seen[p[1]] = true
						rows++
					}
				}
				closest := math.MaxFloat64
				for i := range at {
					for j := i + 1; j < len(at); j++ {
						d := math.Hypot(float64(at[i][0]-at[j][0]), float64(at[i][1]-at[j][1]))
						if d < closest {
							closest = d
						}
					}
				}
				if len(at) < 2 {
					closest = 0
				}
				eaten := done - len(at)
				mark := ""
				if eaten > 0 {
					mark += "  <-- VANISHED"
				}
				if onDisc > 0 {
					mark += "  <-- ON THE DISC"
				}
				if closest < 2.0 && len(at) >= 2 {
					mark += "  <-- MERGES"
				}
				if mark != "" {
					failed = true
				}
				hy := int(float64(gg.h) * 0.42)
				fmt.Printf("  %3dx%-3d  %5.1f  %7d %6d %6d %8d %10d   %d..%-2d of %2d sky  %6.2f%s\n",
					gg.w, gg.h, hour, done, len(at), eaten, onDisc, rows, lo, hi, hy, closest, mark)
			}
		}
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 4. THE PICTURE. Does it read as a scattered sky or still as a band?
// The half-block sky ramp is blanked so the constellation is legible; every
// '*', '.', U+00B7 and '+' below is the real rendered glyph.
// ---------------------------------------------------------------------------

func showSky() {
	fmt.Println("==== 4. THE PICTURE ====")
	type g struct {
		w, h int
		name string
	}
	for _, gg := range []g{{125, 28, "his 125x28"}, {40, 12, "the design floor 40x12"}, {125, 62, "his tall 125x62"}} {
		for _, done := range []int{12, 32} {
			sh, c := warm(gg.w, gg.h, 7, done, 20.0/24, 0.30)
			at := lit(c)
			hy := int(float64(gg.h) * 0.42)
			fmt.Printf("\n  %s, %d earned, %d drawn -- sky rows 0..%d\n", gg.name, done, len(at), hy)
			fmt.Printf("   +%s+\n", repeat("-", gg.w))
			for y := 0; y <= hy && y < gg.h; y++ {
				row := make([]rune, 0, gg.w)
				for x := 0; x < gg.w; x++ {
					r, _, _ := c.ResolveAt(x, y, term.Profile256)
					switch r {
					case '*', '.', '·', '+':
						row = append(row, r)
					default:
						if sh.DiscCovers(x, y) {
							row = append(row, 'O')
						} else {
							row = append(row, ' ')
						}
					}
				}
				fmt.Printf("%2d |%s|\n", y, string(row))
			}
			fmt.Printf("   +%s+\n", repeat("-", gg.w))
			rows := map[int]bool{}
			for _, p := range at {
				rows[p[1]] = true
			}
			fmt.Printf("   rows carrying a star: %d of %d sky rows\n", len(rows), hy+1)
		}
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, s[0])
	}
	return string(out)
}
