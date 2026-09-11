package scape

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// ---------------------------------------------------------------------------
// The instrument. Every number below is read off a RENDERED frame through
// ResolveAt on the 256 profile, never off the source.
//
// A fresh NewShore per frame has no history -- Update clamps a gap of more than
// a second to one nominal step -- so an instrument built on one measures the
// first frame of a session forever. That trap voided two instruments in session
// 27. warmShore turns the same shore over a couple of hundred frames first.
// ---------------------------------------------------------------------------

type geom struct{ w, h int }

// hisGeometries are the window sizes on the record: the design target, the two
// he has run this week, the one the s28 instruments used, and the extremes the
// scape is built for.
var hisGeometries = []geom{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {125, 28}, {143, 27}, {153, 51}, {125, 62}}

func warmShore(w, h int, seed int64, act Activity) (*Shore, *canvas.Canvas) {
	sh := NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 240; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, act)
	}
	return sh, c
}

// litStars finds the constellation in a rendered frame. The '*' is its own
// glyph -- the ambient field is '.', '·' and '+' and shore.go keeps it
// that way on purpose -- so a scan for '*' is the count the eye would make.
func litStars(c *canvas.Canvas) [][2]int {
	var at [][2]int
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				at = append(at, [2]int{x, y})
			}
		}
	}
	return at
}

// starFrame is the cheap frame the position sweeps use: TWO updates rather than
// warmShore's two hundred and forty.
//
// That is a deliberate exception to the warm-up rule and it is held by a test
// of its own, TestTheConstellationDoesNotDependOnTheTide. The rule exists
// because the SEA has history -- Update clamps a gap of more than a second to
// one nominal step, so a shore built fresh per frame renders the first frame of
// a session forever. The constellation has none: its layout is a function of
// the geometry, its ink of the painted sky, and the only thing with history
// that could reach it is the foam, which TestNoStarIsEverInTheWater keeps two
// rows below the band. The sweeps below run 1280 frames; at the full warm-up
// they cost four minutes.
func starFrame(w, h int, seed int64, done int, tod, ctx float64) (*Shore, *canvas.Canvas) {
	return starFrameAt(w, h, seed, done, tod, ctx, 0)
}

// starFrameAt is starFrame at a given point in the wave phase, which is what
// the ambient field's twinkle rides on. It matters to the constellation only
// because a star drawn at less than full alpha composites over whatever the far
// layer put in its cell -- see TestEveryStarReadsAtEveryHour.
func starFrameAt(w, h int, seed int64, done int, tod, ctx float64, phase int) (*Shore, *canvas.Canvas) {
	sh := NewShore(seed, false)
	var c *canvas.Canvas
	for k := 0; k < 2+phase; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, 1+float64(k)*0.08, Activity{Working: true, Level: 0.5,
			TimeOfDay: tod, ContextUsed: ctx, TodoDone: done, TodoTotal: 32})
	}
	return sh, c
}

// The exception starFrame takes, proven rather than asserted: the same frame
// from a warmed shore and from a cold one carries the same stars, in the same
// cells, in the same tones.
func TestTheConstellationDoesNotDependOnTheTide(t *testing.T) {
	n, shimmer := 0, 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 42} {
			for _, tod := range []float64{0.02, 0.3125, 0.5, 0.5938, 0.83} {
				for _, ctx := range []float64{0.05, 0.5, 0.95} {
					_, cold := starFrame(g.w, g.h, seed, 32, tod, ctx)
					_, warm := warmShore(g.w, g.h, seed, Activity{Working: true, Level: 0.5,
						TimeOfDay: tod, ContextUsed: ctx, TodoDone: 32, TodoTotal: 32})
					a, b := litStars(cold), litStars(warm)
					if len(a) == 0 {
						t.Fatalf("%dx%d seed %d: nothing lit", g.w, g.h, seed)
					}
					if len(a) != len(b) {
						t.Fatalf("%dx%d seed %d tod %.2f ctx %.2f: %d stars cold, %d warm", g.w, g.h, seed, tod, ctx, len(a), len(b))
					}
					for i := range a {
						n++
						if a[i] != b[i] {
							t.Fatalf("%dx%d seed %d tod %.2f ctx %.2f: a star is at %v cold and %v warm", g.w, g.h, seed, tod, ctx, a[i], b[i])
						}
						idx := a[i][1]*cold.W + a[i][0]
						if cold.Far().Cells[idx].Set || warm.Far().Cells[idx].Set {
							// An ambient speck is in this cell and it twinkles
							// on the wave phase, which the two frames are at
							// different points of. The tone is allowed to
							// differ; the star is not.
							shimmer++
							continue
						}
						_, fa, ba := cold.ResolveAt(a[i][0], a[i][1], term.Profile256)
						_, fb, bb := warm.ResolveAt(b[i][0], b[i][1], term.Profile256)
						if fa != fb || ba != bb {
							t.Fatalf("%dx%d seed %d: the star at %v is %v on %v cold and %v on %v warm", g.w, g.h, seed, a[i], fa, ba, fb, bb)
						}
					}
				}
			}
		}
	}
	t.Logf("%d stars compared cold against warm: every one in the same cell, and every tone identical bar the %d sitting on a twinkling ambient speck", n, shimmer)
}

// screenGap is the distance between two cells in SCREEN units. A terminal cell
// is about twice as tall as it is wide, so two stars one row apart are as close
// on glass as two stars two columns apart -- and a pair that reads as one mark
// is a miscount in a channel whose whole job is a count.
func screenGap(a, b [2]int) float64 {
	dx := float64(a[0] - b[0])
	dy := float64(a[1]-b[1]) * starCellAspect
	return math.Hypot(dx, dy)
}

func minScreenGap(at [][2]int) (float64, [2]int, [2]int) {
	best := math.MaxFloat64
	var p, q [2]int
	for i := range at {
		for j := i + 1; j < len(at); j++ {
			if d := screenGap(at[i], at[j]); d < best {
				best, p, q = d, at[i], at[j]
			}
		}
	}
	return best, p, q
}

// columnGapSpread is the measurement session 28 started from and the one his
// complaint is about: sort the stars by column and look at the gaps between
// neighbours. sd/mean near 0 is a ruler; a truly random sky measures about 1.
func columnGapSpread(at [][2]int) (mean, sd float64) {
	cols := make([]int, len(at))
	for i, p := range at {
		cols[i] = p[0]
	}
	sort.Ints(cols)
	if len(cols) < 3 {
		return 0, 0
	}
	gaps := make([]float64, 0, len(cols)-1)
	for i := 1; i < len(cols); i++ {
		gaps = append(gaps, float64(cols[i]-cols[i-1]))
	}
	for _, g := range gaps {
		mean += g
	}
	mean /= float64(len(gaps))
	for _, g := range gaps {
		sd += (g - mean) * (g - mean)
	}
	return mean, math.Sqrt(sd / float64(len(gaps)))
}

// bandFor reproduces Update's own arithmetic for one geometry, so a test can
// ask what band a window gets without rendering a frame.
func bandFor(w, h int) (hy, top, bot int) {
	hy = int(float64(h) * 0.42)
	wr := DefaultWriteRows
	if m := h / 6; wr > m {
		wr = m
	}
	if wr < 2 {
		wr = 0
	}
	writeTop := h
	if wr > 0 && h > wr+4 {
		writeTop = h - wr
	}
	beach := h / 5
	if min := wr + 3; wr > 0 && beach < min {
		beach = min
	}
	if beach < 4 {
		beach = 4
	}
	scale := math.Min(float64(w)/80.0, float64(h)/24.0)
	scale = math.Max(0.45, math.Min(1.35, scale))
	sy := h - beach
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > writeTop-1 {
		sy = writeTop - 1
	}
	top, bot = starBand(hy, foamCeiling(sy, scale))
	return hy, top, bot
}

// seaFloorFor is the highest row the FOAM can reach at a geometry, by Update's
// own arithmetic.
func seaFloorFor(w, h int) int {
	hy := int(float64(h) * 0.42)
	wr := DefaultWriteRows
	if m := h / 6; wr > m {
		wr = m
	}
	if wr < 2 {
		wr = 0
	}
	writeTop := h
	if wr > 0 && h > wr+4 {
		writeTop = h - wr
	}
	beach := h / 5
	if min := wr + 3; wr > 0 && beach < min {
		beach = min
	}
	if beach < 4 {
		beach = 4
	}
	scale := math.Max(0.45, math.Min(1.35, math.Min(float64(w)/80.0, float64(h)/24.0)))
	sy := h - beach
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > writeTop-1 {
		sy = writeTop - 1
	}
	return foamCeiling(sy, scale)
}

func lumaOfStar(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// ---------------------------------------------------------------------------
// The guarantees.
// ---------------------------------------------------------------------------

// THE COUNT IS THE CHANNEL, so no two stars may sit close enough to read as
// one. Blue noise is the whole reason the layout is best-candidate rather than
// a hash: a hash gives a sky that collides (five pairs closer than two cells at
// nineteen stars, measured) and a low-discrepancy sequence gives a ruler.
//
// The bar is in SCREEN units, not cells. Two cells apart on a row and one row
// apart are the same distance on glass, and only one of them is two "cells".
func TestNoTwoStarsCollide(t *testing.T) {
	const bar = 2.0 // screen units: further apart than two columns, or one row
	worst := math.MaxFloat64
	var worstAt string
	frames := 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 11, 23, 42} {
			for done := 1; done <= 32; done++ {
				_, c := starFrame(g.w, g.h, seed, done, 20.0/24, 0.3)
				at := litStars(c)
				frames++
				if len(at) == 0 && done > 0 {
					t.Fatalf("%dx%d seed %d: %d done and NOTHING lit -- the instrument is reading an empty sky",
						g.w, g.h, seed, done)
				}
				d, p, q := minScreenGap(at)
				if len(at) < 2 {
					continue
				}
				if d < worst {
					worst = d
					worstAt = fmt.Sprintf("%dx%d seed %d done %d: %v and %v", g.w, g.h, seed, done, p, q)
				}
				if d < bar {
					t.Errorf("%dx%d seed %d done %d: two stars %.2f screen units apart at %v and %v -- they read as one",
						g.w, g.h, seed, done, d, p, q)
				}
			}
		}
	}
	t.Logf("%d frames: the closest pair anywhere is %.2f screen units -- %s", frames, worst, worstAt)
}

// ⭐ THE LOCKED ONE. CLAUDE.md: "Position is fixed by index and seed so a star
// lights where it always was." Best-candidate placement is chosen precisely
// because slot i depends only on slots below it; light them one at a time and
// every earlier star has to be exactly where it was.
func TestAStarNeverMovesAsLaterOnesLight(t *testing.T) {
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 42} {
			var prev [][2]int
			for done := 1; done <= 32; done++ {
				_, c := starFrame(g.w, g.h, seed, done, 20.0/24, 0.3)
				at := litStars(c)
				sort.Slice(at, func(i, j int) bool {
					if at[i][1] != at[j][1] {
						return at[i][1] < at[j][1]
					}
					return at[i][0] < at[j][0]
				})
				// Every star that was lit before must still be lit, in the same
				// cell. The new one is the only difference allowed.
				have := map[[2]int]bool{}
				for _, p := range at {
					have[p] = true
				}
				for _, p := range prev {
					if !have[p] {
						t.Fatalf("%dx%d seed %d: lighting star %d moved or put out the star at %v",
							g.w, g.h, seed, done, p)
					}
				}
				prev = at
			}
			if len(prev) == 0 {
				t.Fatalf("%dx%d seed %d: the sweep ended with an empty sky", g.w, g.h, seed)
			}
		}
	}
}

// The same guarantee against the clock and the context, which move the disc and
// repaint the sky under the stars: a star holds its cell through a whole
// session, not just through one frame.
func TestTheConstellationHoldsStillThroughASession(t *testing.T) {
	for _, g := range []geom{{125, 28}, {153, 51}, {80, 24}} {
		sh := NewShore(7, false)
		var want [][2]int
		tm := 0.0
		moved := 0
		for step := 0; step <= 100; step++ {
			ctx := float64(step) / 100
			tod := math.Mod(9.0/24+float64(step)*0.004, 1)
			var c *canvas.Canvas
			for k := 0; k < 8; k++ {
				c = canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				tm += 0.08
				sh.Update(c, tm, Activity{Working: true, Level: 0.5, TimeOfDay: tod,
					ContextUsed: ctx, TodoDone: 16, TodoTotal: 32})
			}
			at := litStars(c)
			if step == 0 {
				want = at
				if len(want) != 16 {
					t.Fatalf("%dx%d: %d stars lit at the start, want 16", g.w, g.h, len(want))
				}
				continue
			}
			if len(at) != len(want) {
				t.Errorf("%dx%d at context %.2f, %.2f of the day: %d stars lit, want %d",
					g.w, g.h, ctx, tod, len(at), len(want))
				moved++
				continue
			}
			for i := range at {
				if at[i] != want[i] {
					t.Errorf("%dx%d at context %.2f: a star moved from %v to %v", g.w, g.h, ctx, want[i], at[i])
					moved++
					break
				}
			}
		}
		t.Logf("%dx%d: 101 steps of context and clock, %d frames where the constellation was not identical", g.w, g.h, moved)
	}
}

// It still never lands on the disc, at every context level -- which is where
// the disc's altitude comes from. The band now reaches four fifths of the sky
// and the disc's whole descent (0.22 to 0.84 of hy) is inside it, so this
// matters more than it did.
func TestTheConstellationKeepsOffTheDisc(t *testing.T) {
	checked := 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 42} {
			for step := 0; step <= 20; step++ {
				ctx := float64(step) / 20
				sh, c := starFrame(g.w, g.h, seed, 32, 20.0/24, ctx)
				at := litStars(c)
				if len(at) == 0 {
					t.Fatalf("%dx%d seed %d ctx %.2f: nothing lit", g.w, g.h, seed, ctx)
				}
				for _, p := range at {
					checked++
					if sh.DiscCovers(p[0], p[1]) {
						t.Errorf("%dx%d seed %d ctx %.2f: a star at %v is on the disc", g.w, g.h, seed, ctx, p)
					}
				}
			}
		}
	}
	t.Logf("%d lit stars checked against the disc's own footprint", checked)
}

// ⭐ AND IT NO LONGER BLINKS OUT. His report, 2026-09-10: "I can see some of the
// constellation stars appearing and disappearing - once they appear, they
// should not disappear IMO." The cause was the disc sinking through the band
// and todoStars skipping any cell it covered. starColumns keeps the layout out
// of the moon's columns entirely, so the count is now flat across a session.
func TestNoStarGoesOutAsTheDiscSinks(t *testing.T) {
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 42} {
			sh := NewShore(seed, false)
			tm := 0.0
			seen := map[[2]int]bool{}
			out := 0
			for step := 0; step <= 100; step++ {
				var c *canvas.Canvas
				for k := 0; k < 4; k++ {
					c = canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					tm += 0.08
					sh.Update(c, tm, Activity{Working: true, Level: 0.5, TimeOfDay: 11.0 / 24,
						ContextUsed: float64(step) / 100, TodoDone: 32, TodoTotal: 32})
				}
				at := litStars(c)
				now := map[[2]int]bool{}
				for _, p := range at {
					now[p] = true
					seen[p] = true
				}
				for p := range seen {
					if !now[p] {
						out++
						if out <= 3 {
							t.Errorf("%dx%d seed %d: the star at %v went out at context %.2f",
								g.w, g.h, seed, p, float64(step)/100)
						}
						delete(seen, p)
					}
				}
			}
		}
	}
}

// ⭐ HIS FIRST ASK. "can we have the stars appear randomly across the sky,
// instead of from left to right?"
//
// The measurement that says whether it reads as a sky or as a ruler is the
// spread of the gaps between neighbouring columns, sd/mean. A perfectly even
// row measures 0; a Poisson sky -- the thing a sky actually is -- measures
// about 1, and collides.
//
// The layout it replaces is IN this test, so the comparison is made against the
// thing itself rather than against a number copied out of a note: goldenCols
// reproduces the golden-ratio placement that shipped on 2026-09-10. The bar is
// 0.50, chosen from what that layout can actually produce -- its worst case
// over these counts, seeds and geometries is printed in the log and is the
// reason the bar sits where it does.
func goldenLayout(i, w, hy int, seed int64) (int, int) {
	top, bot := 1, hy*3/5
	if bot <= top {
		bot = top + 1
	}
	frac := math.Mod((float64(i)+0.5)*0.6180339887, 1)
	x := int(frac*float64(w-6)) + 3
	x += int(HashF(i, 11, seed+31)*3) - 1
	y := top + int(HashF(i, 23, seed+37)*float64(bot-top))
	return x, y
}

func TestTheSkyIsScatteredNotRuled(t *testing.T) {
	const bar = 0.50
	oldWorst, oldBest := math.MaxFloat64, 0.0
	for _, g := range []geom{{125, 28}, {153, 51}, {143, 27}} {
		for _, done := range []int{12, 19, 26, 32} {
			worst, worstSeed := math.MaxFloat64, int64(0)
			for _, seed := range []int64{5, 7, 11, 23, 42} {
				_, c := starFrame(g.w, g.h, seed, done, 20.0/24, 0.3)
				at := litStars(c)
				if len(at) != done {
					t.Fatalf("%dx%d seed %d: %d stars lit, want %d", g.w, g.h, seed, len(at), done)
				}
				mean, sd := columnGapSpread(at)
				if mean <= 0 {
					t.Fatalf("%dx%d seed %d: no gaps to measure", g.w, g.h, seed)
				}
				if sd/mean < worst {
					worst, worstSeed = sd/mean, seed
				}
				// The layout this replaces, at the same count and seed.
				var old [][2]int
				hy := int(float64(g.h) * 0.42)
				for i := 0; i < done; i++ {
					x, y := goldenLayout(i, g.w, hy, seed)
					old = append(old, [2]int{x, y})
				}
				om, osd := columnGapSpread(old)
				if om > 0 {
					if osd/om < oldWorst {
						oldWorst = osd / om
					}
					if osd/om > oldBest {
						oldBest = osd / om
					}
				}
			}
			t.Logf("%dx%d %2d stars: column-gap sd/mean is %.2f at its most even (seed %d)", g.w, g.h, done, worst, worstSeed)
			if worst < bar {
				t.Errorf("%dx%d %d stars: sd/mean %.2f at seed %d -- that is a ruler, not a sky (bar %.2f)",
					g.w, g.h, done, worst, worstSeed, bar)
			}
		}
	}
	t.Logf("the golden-ratio layout this replaces measures %.2f to %.2f over the same sweep -- the bar of %.2f is above all of it",
		oldWorst, oldBest, bar)
	if oldBest >= bar {
		t.Errorf("the bar of %.2f is inside the old layout's range (it reached %.2f) -- it no longer separates them", bar, oldBest)
	}
}

// ⭐ HIS SECOND ASK. "Also can we spread them more vertically? currently
// sitting in a narrow band pretty high up."
//
// It was rows 1..hy*3/5 and every star was inside the top 55-60% of the sky.
// Three things are held here: the band the layout is ALLOWED, the fact that the
// stars actually reach into the lower half of it rather than piling at the top,
// and the one place the answer is "no" -- a scape so short that the tide brings
// the water up into the sky, where the band is bounded by the water instead and
// is SHALLOWER than it was. That case is 40x12, it is named, and the water
// guarantee below is what it is traded for.
func TestTheConstellationUsesMoreThanTheTopHalf(t *testing.T) {
	for _, g := range hisGeometries {
		hy, top, bot := bandFor(g.w, g.h)
		was := hy * 3 / 5
		switch {
		case bot > was:
			t.Logf("%dx%d: sky rows 0..%d, band %d..%d -- was 1..%d", g.w, g.h, hy-1, top, bot, was)
		default:
			// Only acceptable when the water is what stopped it, and then the
			// band is down to its floor of three rows. It still has to be clear
			// of the foam.
			sea := seaFloorFor(g.w, g.h)
			if bot >= sea {
				t.Errorf("%dx%d: the band reaches row %d and the foam reaches row %d", g.w, g.h, bot, sea)
			}
			if bot-top > 2 {
				t.Errorf("%dx%d: the band is rows %d..%d of %d sky rows -- no deeper than the one he complained about, and the water (row %d) is not why",
					g.w, g.h, top, bot, hy, sea)
			}
			t.Logf("%dx%d: sky rows 0..%d, band %d..%d -- bounded by the water at row %d, not by the sky",
				g.w, g.h, hy-1, top, bot, sea)
		}
		if bot >= hy {
			t.Errorf("%dx%d: the band reaches row %d of %d sky rows -- into the horizon", g.w, g.h, bot, hy)
		}
		if bot-top < 3 {
			continue // too few rows for "half" to mean anything
		}
		half := (top + bot) / 2
		for _, seed := range []int64{5, 7, 11, 23, 42} {
			_, c := starFrame(g.w, g.h, seed, 32, 20.0/24, 0.3)
			at := litStars(c)
			if len(at) == 0 {
				t.Fatalf("%dx%d seed %d: nothing lit", g.w, g.h, seed)
			}
			lower, deepest := 0, 0
			for _, p := range at {
				if p[1] > half {
					lower++
				}
				if p[1] > deepest {
					deepest = p[1]
				}
			}
			if lower*100/len(at) < 25 {
				t.Errorf("%dx%d seed %d: only %d of %d stars are below row %d -- still piled at the top",
					g.w, g.h, seed, lower, len(at), half)
			}
			if deepest < bot-1 {
				t.Errorf("%dx%d seed %d: the deepest star is row %d, and the band goes to %d",
					g.w, g.h, seed, deepest, bot)
			}
		}
	}
}

// ⚠ AND NEVER IN THE WATER. The deeper band reaches toward the horizon, and at
// the smallest geometries the tide reaches back: at 40x12 the water withdraws
// to row 4 of a five-row sky. foam() plots into the same near layer AFTER the
// constellation and wins the cell, so a star down there is simply deleted --
// which is a star that went out, the defect this session is closing.
//
// The old band had the same hole (it reached row 3 at 40x12 and the foam
// reaches row 3), so this is a guarantee the channel never had rather than a
// regression being fenced off.
func TestNoStarIsEverInTheWater(t *testing.T) {
	// The short end is swept as well as his own sizes, and it is not padding:
	// at heights 8, 9 and 10 the tide reaches row 1 or 2 of a three or
	// four-row sky, and the band's own minimum width will happily push a star
	// into it if it is not stopped. Measured before the floor was fixed: 2168
	// star-frames in the foam at 40x9.
	geoms := append([]geom{{30, 8}, {40, 9}, {40, 10}, {60, 11}, {80, 14}, {80, 16}}, hisGeometries...)
	for _, g := range geoms {
		for _, lvl := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
			sh := NewShore(7, false)
			tm := 0.0
			worst := 99
			for k := 0; k < 400; k++ {
				c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				tm += 0.08
				sh.Update(c, tm, Activity{Working: lvl > 0, Level: lvl, TimeOfDay: 11.0 / 24,
					ContextUsed: 0.5, TodoDone: 32, TodoTotal: 32})
				if k < 200 {
					continue // let the tide settle where this level puts it
				}
				top := 999
				for _, e := range sh.lastEdge {
					if r := int(e+0.5) - 1; r < top {
						top = r // the highest row foam() can speckle
					}
				}
				at := litStars(c)
				if len(at) == 0 && int(float64(g.h)*0.42) >= 3 {
					t.Fatalf("%dx%d level %.2f: nothing lit", g.w, g.h, lvl)
				}
				for _, p := range at {
					if d := top - p[1]; d < worst {
						worst = d
					}
					if p[1] >= top {
						t.Fatalf("%dx%d level %.2f: a star at %v is in the foam (highest foam row %d)",
							g.w, g.h, lvl, p, top)
					}
				}
			}
			if lvl == 0 {
				t.Logf("%dx%d: at rest the deepest star clears the highest foam row by %d rows", g.w, g.h, worst)
			}
		}
	}
}

// ⚠ AND THE READOUT NEVER COVERS ONE. This reproduces drawReadout's own
// placement from live.go -- the context number under the disc, or beside it
// once the disc rides too low for that -- and asserts the constellation is
// never under it.
//
// It is here because the deeper band CREATED this collision: over this exact
// sweep the golden-ratio layout it replaces covered a star at 0 of 6072 readout
// cells and the first deep layout covered 50, losing a star for a stretch of
// the context range in four of twenty-four window-and-seed pairs. drawReadout
// plots into the same near layer after the scape, so the star is deleted, and
// unlike the moon the number does not move on.
//
// ⚠ THE ARITHMETIC IS COPIED, and that is the point: readoutGround in shore.go
// mirrors it, and this test is what holds the two ends together. If drawReadout
// moves and this is not updated, this goes red rather than the channel going
// quietly wrong.
func TestTheReadoutNeverCoversAStar(t *testing.T) {
	cells, hits := 0, 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 11, 23, 42} {
			for step := 40; step <= 100; step++ {
				used := float64(step) / 100
				sh, c := starFrame(g.w, g.h, seed, 32, 20.0/24, used)
				at := map[[2]int]bool{}
				for _, p := range litStars(c) {
					at[p] = true
				}
				if len(at) == 0 {
					t.Fatalf("%dx%d seed %d used %.2f: nothing lit", g.w, g.h, seed, used)
				}
				// --- drawReadout, live.go ---
				mx, my := sh.MoonPos()
				rx, ry := sh.MoonExtent()
				txt := fmt.Sprintf("%.0f%%", (1-used)*100)
				if used >= 0.85 {
					txt += " left"
				}
				hy := int(float64(g.h)*0.42) + 1
				x, y := mx-len(txt)/2, my+ry+1
				if y > hy-1 {
					y = my
					x = mx + rx + 2
					if x+len(txt) > g.w-1 {
						x = mx - rx - 1 - len(txt)
					}
				}
				if x < 1 {
					x = 1
				}
				if x+len(txt) > g.w-1 {
					x = g.w - 1 - len(txt)
				}
				// --- end ---
				for i := 0; i < len(txt); i++ {
					cells++
					if at[[2]int{x + i, y}] {
						hits++
						if hits <= 5 {
							t.Errorf("%dx%d seed %d used %.2f: the readout %q covers the star at (%d,%d)",
								g.w, g.h, seed, used, txt, x+i, y)
						}
					}
				}
			}
		}
	}
	t.Logf("%d readout cells over the sweep, %d of them on a star", cells, hits)
}

// The constellation and the ambient star field share the sky but not a glyph:
// stars() draws '.', '·' and '+' into the FAR layer and the checklist owns '*'
// in the NEAR one. Where they land on the same cell the near layer wins, so the
// cell reads '*' and the count is unharmed -- but the ambient speck under it is
// gone, which is worth knowing and worth holding: if the ambient field ever
// took the cell, a finished todo would read as dust.
func TestTheConstellationWinsTheCellFromTheAmbientField(t *testing.T) {
	stars, shared := 0, 0
	for _, g := range hisGeometries {
		for _, seed := range []int64{5, 7, 42} {
			for i := 0; i < 8; i++ {
				tod := float64(i)/8 + 0.01
				_, c := starFrame(g.w, g.h, seed, 32, tod, 0.3)
				at := litStars(c)
				if len(at) == 0 {
					t.Fatalf("%dx%d seed %d tod %.2f: nothing lit", g.w, g.h, seed, tod)
				}
				for _, p := range at {
					stars++
					if c.Far().Cells[p[1]*c.W+p[0]].Set {
						shared++
					}
					if r, _, _ := c.ResolveAt(p[0], p[1], term.Profile256); r != '*' {
						t.Errorf("%dx%d seed %d tod %.2f: the cell at %v draws %q, not a star", g.w, g.h, seed, tod, p, string(r))
					}
				}
			}
		}
	}
	t.Logf("%d lit stars, %d of them over an ambient speck -- every one of them still draws '*'", stars, shared)
}

// The deeper band is only usable because the ink adapts to the ground it is
// painted on. This is the same 40-luma bar TestTheChecklistIsLegibleAtEveryHour
// holds, swept at 32 hours instead of 8 and at every geometry -- and the finer
// sweep is not decoration: the worst hour of the day is 17:45, which an
// eight-hour sweep steps straight over.
//
// The contrast is read INSIDE the star's own cell: ResolveAt hands back the
// glyph's colour and the ground it is drawn on, and those two are what the eye
// compares. Measuring against a neighbouring cell -- which is what the older
// test did, at column 1 of the same row -- can be out by a whole ramp step,
// and on a small scape by thirty luma.
func TestEveryStarReadsAtEveryHour(t *testing.T) {
	worst, worstAt := math.MaxFloat64, ""
	n := 0
	for _, g := range hisGeometries {
		for i := 0; i < 32; i++ {
			tod := float64(i) / 32
			// The phase is swept because a star is drawn at its magnitude's
			// alpha, not at full, so the ambient speck in its cell composites
			// through it -- and that speck twinkles. Measured, it moves the
			// quantised tone by up to twenty luma in EITHER direction: the
			// saturate-then-quantise path a glyph takes is not monotone in
			// luma, so "the ambient is brighter, so it can only help" is wrong.
			for phase := 0; phase < 4; phase++ {
				_, c := starFrameAt(g.w, g.h, 7, 32, tod, 0.3, phase*3)
				at := litStars(c)
				if len(at) == 0 {
					t.Fatalf("%dx%d tod %.4f: nothing lit", g.w, g.h, tod)
				}
				for _, p := range at {
					ch, fg, bg := c.ResolveAt(p[0], p[1], term.Profile256)
					if ch != '*' {
						t.Fatalf("%dx%d tod %.4f: %v is not a star", g.w, g.h, tod, p)
					}
					d := lumaOfStar(fg) - lumaOfStar(bg)
					n++
					if d < worst {
						worst, worstAt = d, fmt.Sprintf("%dx%d tod %.4f phase %d at %v", g.w, g.h, tod, phase, p)
					}
					if d < 40 {
						t.Errorf("%dx%d tod %.4f phase %d: the star at %v reads %+.1f luma above the ground it is on", g.w, g.h, tod, phase, p, d)
					}
				}
			}
		}
	}
	t.Logf("%d lit stars measured on the rendered frame: the dimmest reads %+.1f above its own ground -- %s", n, worst, worstAt)
}

// A star's magnitude is a fact about its slot, not about the frame: the same
// index is the same brightness every time, and the variety never takes one
// below the floor (that is what TestEveryStarReadsAtEveryHour holds). Without
// the variety a constellation is a row of identical asterisks, which is what a
// dashboard looks like.
func TestMagnitudesVaryAndHold(t *testing.T) {
	const w, h = 125, 28
	tones := map[[2]int]term.RGB{}
	distinct := map[term.RGB]int{}
	for pass := 0; pass < 3; pass++ {
		_, c := starFrame(w, h, 7, 32, 2.0/24, 0.3) // night: the palette tone, not the day lift
		at := litStars(c)
		if len(at) != 32 {
			t.Fatalf("pass %d: %d stars lit, want 32", pass, len(at))
		}
		for _, p := range at {
			_, fg, _ := c.ResolveAt(p[0], p[1], term.Profile256)
			if pass == 0 {
				tones[p] = fg
				distinct[fg]++
				continue
			}
			if tones[p] != fg {
				t.Errorf("the star at %v changed tone between frames: %v then %v", p, tones[p], fg)
			}
		}
	}
	if len(distinct) < 3 {
		t.Errorf("the whole constellation is %d tone(s) -- that is a row of identical asterisks, not a sky", len(distinct))
	}
	t.Logf("32 stars carry %d distinct tones at night, every one stable across frames", len(distinct))
}

// ---------------------------------------------------------------------------
// The renders and the numbers, for a human. Not a guarantee; it prints.
// ---------------------------------------------------------------------------

func TestReportTheSky(t *testing.T) {
	w, h := 125, 28
	hy, top, bot := bandFor(w, h)
	oldTop, oldBot := 1, hy*3/5
	for _, done := range []int{12, 19, 26} {
		sh, c := starFrame(w, h, 7, done, 20.0/24, 0.56)
		at := litStars(c)
		mean, sd := columnGapSpread(at)
		d, _, _ := minScreenGap(at)

		draw := func(pts [][2]int, disc bool) string {
			grid := make([][]rune, hy)
			for y := range grid {
				grid[y] = make([]rune, w)
				for x := range grid[y] {
					grid[y][x] = ' '
				}
			}
			if disc {
				for y := 0; y < hy; y++ {
					for x := 0; x < w; x++ {
						if sh.DiscCovers(x, y) {
							grid[y][x] = 'o'
						}
					}
				}
			}
			for _, p := range pts {
				if p[1] >= 0 && p[1] < hy && p[0] >= 0 && p[0] < w {
					grid[p[1]][p[0]] = '*'
				}
			}
			out := "+" + dashes(w) + "+\n"
			for y := 0; y < hy; y++ {
				out += "|" + string(grid[y]) + "|\n"
			}
			return out + "+" + dashes(w) + "+\n"
		}

		var old [][2]int
		for i := 0; i < done; i++ {
			x, y := goldenLayout(i, w, hy, 7)
			old = append(old, [2]int{x, y})
		}
		om, osd := columnGapSpread(old)
		od, _, _ := minScreenGap(old)

		out := fmt.Sprintf("\n=== %d STARS, 125x28, sky rows 0..%d ===\n\n", done, hy-1)
		out += fmt.Sprintf("BEFORE -- golden ratio, band %d..%d, column-gap sd/mean %.2f, closest pair %.1f screen units\n",
			oldTop, oldBot, osd/om, od)
		out += draw(old, false)
		out += fmt.Sprintf("\nAFTER -- dart-thrown, band %d..%d, column-gap sd/mean %.2f, closest pair %.1f screen units\n",
			top, bot, sd/mean, d)
		out += draw(at, true)
		out += "(o is the moon's footprint; the layout keeps off its columns entirely)\n"
		t.Log(out)
	}
}

func dashes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}
