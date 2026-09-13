package scape

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// The flight path, swept. star_steady_test.go is blind to the fall -- it counts
// SETTLED stars and the arriving one is drawn from a different branch -- so
// every guarantee below is one nothing else in the repo would catch.
//
// ⚠ ONE SEED, ONE GEOMETRY OR ONE HOUR IS NOT A MEASUREMENT. That is the trap
// session 28 paid for twice: a constellation pronounced steady "three ways" was
// steady at seed 7 only, and over seeds 1/7/42 the same layout had 891 of 1818
// frames with the wrong star count. The sweep below is the population, and the
// disc's position is on it because readoutGround and the moon's corridor both
// move with the context.
type arrCase struct {
	w, h  int
	seed  int64
	ctx   float64
	total int
}

func arrCases() []arrCase {
	var out []arrCase
	for _, g := range [][2]int{{125, 28}, {143, 27}, {107, 51}, {80, 24}, {60, 20}, {40, 12}} {
		for _, seed := range []int64{1, 7, 42} {
			for _, ctx := range []float64{0, 0.45, 0.93} {
				for _, total := range []int{5, 19, 32} {
					out = append(out, arrCase{g[0], g[1], seed, ctx, total})
				}
			}
		}
	}
	return out
}

// arrStage warms a shore the way the product does and hands back everything
// the track needs. ONE Shore, updated repeatedly: a fresh Shore has no wave
// clock and no tide, and that trap has voided two instruments here already.
func arrStage(c arrCase, tod float64) (*Shore, int, int, int, [][2]int) {
	cv := canvas.New(c.w, c.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(c.seed, false)
	act := Activity{Level: 0.5, Working: true, ContextUsed: c.ctx, TimeOfDay: tod,
		TodoDone: c.total, TodoTotal: c.total}
	for i := 0; i < 30; i++ {
		sh.Update(cv, 2.5+float64(i)/20, act)
	}
	sh.Update(cv, 4, act)

	hy := int(float64(c.h) * 0.42)
	scale := math.Max(0.45, math.Min(1.35, math.Min(float64(c.w)/80.0, float64(c.h)/24.0)))
	sy := c.h - 6
	if sy <= hy+1 {
		sy = hy + 2
	}
	top, bot := starBand(hy, foamCeiling(sy, hy, scale))
	if bot < top {
		return sh, hy, top, bot, nil
	}
	return sh, hy, top, bot, sh.starPlaces(c.w, hy, top, bot, c.total)
}

// Every cell of the fall is a cell the sky would have given a star, and the
// last one is the star's own place.
//
// The landing is the assertion that matters most and it is the one that is true
// BY CONSTRUCTION rather than by arithmetic: arrivalTrack walks outward from
// home and flies the result in reverse, so a guard can only ever trim the
// launch. The drafted version searched for a launch point and walked in, and
// put between 8.7% and 33.6% of its launches off the canvas depending on the
// geometry -- the same bug class, unreachable here.
func TestTheFallLandsOnTheStarsOwnPlace(t *testing.T) {
	flew, total := 0, 0
	byGeom := map[string][2]int{}
	for _, cs := range arrCases() {
		for _, tod := range []float64{0, 0.25, 0.5, 0.583, 0.75, 0.9} {
			sh, hy, top, bot, pts := arrStage(cs, tod)
			if pts == nil {
				continue
			}
			for idx := 1; idx < len(pts); idx++ {
				// A fresh key every slot: the cache is keyed on (geometry,
				// slot) and the sweep walks slots, so nothing is reused
				// across cases by accident.
				track := sh.arrivalTrack(cs.w, hy, top, bot, idx, pts)
				total++
				g := fmt.Sprintf("%dx%d", cs.w, cs.h)
				n := byGeom[g]
				n[1]++
				if len(track) == 0 {
					byGeom[g] = n
					continue // arrives in place; counted below
				}
				n[0]++
				byGeom[g] = n
				flew++
				home := pts[idx]
				if got := track[len(track)-1]; got != home {
					t.Fatalf("%dx%d seed %d ctx %.2f tod %.2f slot %d: landed on %v, want its own place %v",
						cs.w, cs.h, cs.seed, cs.ctx, tod, idx, got, home)
				}
				if len(track) < arrivalMin || len(track) > arrivalCols+1 {
					t.Fatalf("%dx%d slot %d: track of %d cells, want %d..%d",
						cs.w, cs.h, idx, len(track), arrivalMin, arrivalCols+1)
				}
				blocked := sh.readoutGround(cs.w, hy, top, bot)
				lit := pts[:idx]
				lastX, lastY := -1, -1
				for i, p := range track {
					x, y := p[0], p[1]
					switch {
					case x < 0 || x >= cs.w || y < 0 || y >= cs.h:
						t.Fatalf("%dx%d slot %d cell %d: %v is off the canvas", cs.w, cs.h, idx, i, p)
					case sh.DiscCovers(x, y):
						t.Fatalf("%dx%d slot %d cell %d: %v is on the disc -- the moon carries context and a star inside it reads as part of it",
							cs.w, cs.h, idx, i, p)
					case blocked[[2]int{x, y}]:
						t.Fatalf("%dx%d slot %d cell %d: %v is where the readout lands, so the star would be deleted mid-fall",
							cs.w, cs.h, idx, i, p)
					}
					if i > 0 {
						// One column a frame, and the drop never reverses:
						// anything else is not a fall.
						if dx := x - lastX; dx != 1 && dx != -1 {
							t.Fatalf("%dx%d slot %d cell %d: column moved by %d", cs.w, cs.h, idx, i, dx)
						}
						if dy := y - lastY; dy != 0 && dy != 1 {
							t.Fatalf("%dx%d slot %d cell %d: row moved by %d, want 0 or +1 (it falls)",
								cs.w, cs.h, idx, i, dy)
						}
					}
					lastX, lastY = x, y
					if i == len(track)-1 {
						continue // home IS a star's place
					}
					for _, q := range lit {
						if q[0] == x && q[1] == y {
							t.Fatalf("%dx%d slot %d cell %d: %v is a lit star -- two stars in one cell is a count error",
								cs.w, cs.h, idx, i, p)
						}
						d := math.Hypot(float64(x-q[0]), 2*float64(y-q[1]))
						if d < arrivalSep {
							t.Fatalf("%dx%d slot %d cell %d: %v passes %v at %.2f screen units, bar is %.2f -- they would merge into one mark",
								cs.w, cs.h, idx, i, p, q, d, arrivalSep)
						}
					}
				}
				// Away from the moon, or flipped because that side had no room.
				// Either way the whole track is on ONE side of home.
				dir := track[0][0] - home[0]
				if dir == 0 {
					t.Fatalf("%dx%d slot %d: launch shares home's column", cs.w, cs.h, idx)
				}
			}
		}
	}
	// The rate is REPORTED, not invented. A fall that almost never happens is
	// a feature that does not exist, and the bar is what the study measured
	// over its own 1,805 swept arrivals: 93.5% flew.
	rate := float64(flew) / float64(total)
	t.Logf("%d of %d arrivals fly (%.1f%%); the rest land in place", flew, total, rate*100)
	// Per geometry, because the in-place cases are not spread evenly: a small
	// sky runs out of room and a wide one almost never does, and which is which
	// is the number worth keeping.
	geoms := make([]string, 0, len(byGeom))
	for g := range byGeom {
		geoms = append(geoms, g)
	}
	sort.Strings(geoms)
	for _, g := range geoms {
		n := byGeom[g]
		t.Logf("  %-8s %5d of %5d fly (%.1f%%)", g, n[0], n[1], 100*float64(n[0])/float64(n[1]))
	}
	if rate < 0.80 {
		t.Fatalf("only %.1f%% of arrivals fly, want at least 80%% -- the fall is not reaching the sky", rate*100)
	}

	// ⚠ AND THE AVERAGE IS NOT THE BAR. An average that passes is not a ceiling
	// that holds -- the lesson two of session 29's rejected fixes taught -- and
	// this average is dragged down by one geometry on purpose.
	//
	// 40x12 flies 38.0% and that is NOT a defect: the sky there is five rows,
	// the band is three, and 32 places in it leave no six-column corridor clear
	// of the other stars. A refused track degrades to exactly the behaviour that
	// shipped before the fall existed -- the star appears in its place -- so the
	// small end loses the animation and nothing else. It is also the size where
	// the balloon's ground eats half the sky, and "no fall below this size" is a
	// decision he has not been asked for.
	//
	// The geometries he actually runs are held to a real floor, so a regression
	// at his own window cannot hide inside the mean.
	for _, g := range []string{"125x28", "143x27", "107x51", "80x24", "60x20"} {
		n := byGeom[g]
		if n[1] == 0 {
			t.Fatalf("%s is not on the sweep any more", g)
		}
		if r := float64(n[0]) / float64(n[1]); r < 0.95 {
			t.Fatalf("%s: only %.1f%% of arrivals fly, want at least 95%%", g, r*100)
		}
	}
}

// The head never skips a cell, at both frame rates that ship.
//
// A skipped cell is a teleport: the star is in two places over two frames with
// nothing between them, which reads as two stars and not as one falling. This
// is the criterion that chose LINEAR over the drafted ease-out -- an ease-out
// spends its first frames crossing several columns at once.
//
// ⚠ THE FLOOR IS THE HOSTED FRAME RATE, 12. `xscapes claude` runs at 12 fps and
// the standalone scape at 20, and half a second at 12 gives six frames for a
// six-column path, which is exactly enough. Below that it cannot hold: at 8 fps
// the fall gets four frames and the head jumps two columns at a time. That is a
// stated limit of `-fps` under 12, not a defect in the path -- the duration is
// his ruling and 0.50 s is what he picked.
func TestTheFallNeverSkipsACell(t *testing.T) {
	for _, fps := range []float64{12, 20, 30, 60} {
		frames := int(math.Round(0.5 * fps))
		for n := arrivalMin; n <= arrivalCols+1; n++ {
			prev, seq := 0, []int{0}
			for k := 1; k < frames; k++ {
				// The phase the reducer hands over: age/StarFall, strictly
				// below 1, so the landing frame is the settled frame.
				got := arrivalHead(n, float64(k)/float64(frames))
				if got < prev {
					t.Fatalf("fps %.0f track %d frame %d: head went backwards, %d then %d", fps, n, k, prev, got)
				}
				if got-prev > 1 {
					t.Fatalf("fps %.0f track %d frame %d: head skipped %d cells (%d then %d) -- a teleport at this rate",
						fps, n, k, got-prev, prev, got)
				}
				prev, seq = got, append(seq, got)
			}
			// And the last airborne frame is one step from home, so the
			// settled frame that follows it completes the fall rather than
			// finishing it early or jumping.
			if n-1-prev > 1 {
				t.Fatalf("fps %.0f track %d: last airborne head at %d of %d, so landing jumps %d cells (%v)",
					fps, n, prev, n-1, n-1-prev, seq)
			}
		}
	}
}

// A phase of 0 is the launch and nothing else, and a zero Activity flies
// nothing at all.
//
// ⚠ THIS IS THE TRAP THE SEPARATE BOOL EXISTS FOR. Had the phase been a bare
// float with "0 means launching now", every pinned frame in the repo -- the
// study pages, the site clips, the mockups, all of which build an Activity by
// hand -- would fire a fall on its first frame. ambientFrame alone has 19
// pinned cells.
func TestAZeroActivityFliesNothing(t *testing.T) {
	var zero Activity
	if zero.Arriving {
		t.Fatal("the zero Activity is arriving, so every hand-built frame in the repo would fly a star")
	}
	if arrivalHead(7, 0) != 0 {
		t.Fatal("phase 0 is not the launch cell")
	}
	if got := arrivalHead(7, 1); got != 6 {
		t.Fatalf("phase 1 lands on cell %d, want home at 6", got)
	}
	// Out-of-range phases are clamped rather than indexing off the track.
	if got := arrivalHead(7, 9); got != 6 {
		t.Fatalf("phase 9 gave cell %d, want the landing clamped to 6", got)
	}
	if got := arrivalHead(7, -3); got != 0 {
		t.Fatalf("phase -3 gave cell %d, want the launch clamped to 0", got)
	}
	if got := arrivalHead(1, 0.5); got != 0 {
		t.Fatalf("a one-cell track gave cell %d, want 0", got)
	}
}

// The path is fixed for the whole fall.
//
// It is CACHED for this and not to save the work: readoutGround moves with the
// disc, the disc sinks as the context fills, and a path rebuilt from scratch
// every frame could shift under the star in mid-air -- which would look like
// the one thing the no-skip test forbids.
func TestTheFallsPathDoesNotMoveUnderIt(t *testing.T) {
	cs := arrCase{125, 28, 7, 0.30, 32}
	sh, hy, top, bot, pts := arrStage(cs, 0.9)
	if pts == nil {
		t.Skip("no band at this geometry")
	}
	first := sh.arrivalTrack(cs.w, hy, top, bot, 11, pts)
	if len(first) < arrivalMin {
		t.Skip("slot 11 arrives in place here")
	}
	// The context climbing is what moves the readout and the disc. The track
	// for a slot already in flight must not care.
	cv := canvas.New(cs.w, cs.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for _, ctx := range []float64{0.35, 0.55, 0.80, 0.99} {
		sh.Update(cv, 4.2, Activity{Level: 0.5, Working: true, ContextUsed: ctx,
			TimeOfDay: 0.9, TodoDone: 32, TodoTotal: 32})
		again := sh.arrivalTrack(cs.w, hy, top, bot, 11, pts)
		if len(again) != len(first) {
			t.Fatalf("ctx %.2f: track is %d cells, was %d -- it moved mid-fall", ctx, len(again), len(first))
		}
		for i := range first {
			if again[i] != first[i] {
				t.Fatalf("ctx %.2f: cell %d moved from %v to %v mid-fall", ctx, i, first[i], again[i])
			}
		}
	}
}
