package scape

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The ambient dust field's guarantees.
//
// ⭐ HIS REPORT, 2026-09-11: "during the daytime I dont see any other characters
// tho like I saw before during nightime giving depth to the constellation." And
// on 2026-09-10, of the same field: "stars appearing and dissapearing", and
// "any ideas around different characters? ... I see more characters varying in
// sizes and shapes appearing."
//
// All three are one defect and one shortage. The defect is that the field used
// to encode the clock in ALPHA, so at a low StarVis its specks quantised onto
// the same cube index as the sky and could not be seen -- 53% of them lost at
// 16:00 -- while a hard gate at alpha 0.02 deleted rather than dimmed, so below
// StarVis 0.2 every speck switched off once a cycle. The shortage is that four
// glyph slots held three distinct shapes.
//
// Every number in these tests is read off the RENDERED cell through
// canvas.ResolveAt(x, y, term.Profile256). Nothing here asks the source what it
// meant to draw.

// ambientGeoms are the window sizes these guarantees are held at: his own two,
// the design target, and the floor the brief says must still look fine.
var ambientGeoms = [][2]int{{125, 28}, {143, 27}, {80, 24}, {40, 12}}

// ambientFrame warms ONE shore and hands back its last canvas. A fresh
// scape.NewShore per frame has no history -- Update clamps a >1 s gap to one
// nominal step -- so a field sampled off a cold shore is not the field he sees.
func ambientFrame(t *testing.T, w, h int, seed int64, hour float64, frames int) *canvas.Canvas {
	t.Helper()
	sh := NewShore(seed, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < frames; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 1.0 / 12 // inside.go paces at 12 fps
		sh.Update(c, tm, Activity{
			Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: hour / 24,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	return c
}

// skyRows is how many rows of a frame this scape gives the sky. It mirrors
// Update's own split so a test never reads the sea's rows as sky.
func skyRows(h int) int { return int(float64(h) * 0.42) }

// ambientCells is every cell the ambient field put a glyph in, over the sky.
func ambientCells(c *canvas.Canvas) [][2]int {
	var out [][2]int
	for y := 0; y < skyRows(c.H); y++ {
		for x := 0; x < c.W; x++ {
			if c.Far().Cells[y*c.W+x].Set {
				out = append(out, [2]int{x, y})
			}
		}
	}
	return out
}

// AT NOON THE AMBIENT FIELD IS EMPTY.
//
// HIS RULING, 2026-09-11: "make sure we can still see some during daytime, even
// if fainter than at nighttime."
//
// ⚠ THIS TEST USED TO ASSERT THE OPPOSITE -- TestAtNoonTheAmbientFieldIsEmpty,
// written from a brief that predates the ruling by about an hour. The intent it
// encoded was real and is recorded in CLAUDE.md ("the moon and the constellation
// are washed out at midday by design"): the sky is the WORLD, and a real midday
// sky has no stars in it. He has ruled the other way, for the thing that matters
// more in a scene somebody looks at -- an empty sky reads as a scape that has
// stopped working, which is the note he gave twice ("this felt too bare").
//
// So the guarantee flips. Noon is the FLOOR of the clock's range, not zero, and
// what has to hold is that the floor is small, present, and legible.
func TestAtNoonTheSkyIsSparseButNotEmpty(t *testing.T) {
	if v := PaletteAt(0.5).StarVis; v != 0 {
		t.Fatalf("noon StarVis is %v, not 0 -- this test is measuring the wrong thing", v)
	}
	checked := 0
	for _, g := range ambientGeoms {
		for _, seed := range []int64{1, 7, 42, 1009} {
			noon := len(ambientCells(ambientFrame(t, g[0], g[1], seed, 12.0, 60)))
			night := len(ambientCells(ambientFrame(t, g[0], g[1], seed, 0.0, 60)))
			if noon == 0 {
				t.Errorf("%dx%d seed %d: the midday sky is EMPTY -- his ruling is that some stay",
					g[0], g[1], seed)
			}
			// "fainter than at nighttime" -- the day must still be the thin end
			// of the range, or the clock has stopped meaning anything.
			if night > 0 && noon >= night {
				t.Errorf("%dx%d seed %d: %d specks at noon against %d at midnight -- "+
					"midday is supposed to be the thin end", g[0], g[1], seed, noon, night)
			}
			checked++
		}
	}
	if checked != len(ambientGeoms)*4 {
		t.Fatalf("only %d frames checked", checked)
	}
}

// NO SPECK IS EVER DRAWN IN A COLOUR INDISTINGUISHABLE FROM THE CELL BEHIND IT.
//
// ⭐ THIS IS THE WHOLE DEFECT. Measured at 125x28 on the shipped field: 39
// specks plotted at 16:00 and 18 of them resolved to a foreground the cube put
// on the same index as their own background, or within a luma step or two of
// it. His 20:22 screenshot and his 20:55 screenshot PLOT THE SAME 39 SPECKS --
// 2 of them cleared 24 luma in the first and 15 in the second, and that is the
// entire difference between a sky he called bare and a sky he called right.
//
// So the field is no longer allowed to draw what cannot be seen. speck() walks
// the opacity and the ink upward until the rendered cell clears ambientContrast
// against its own ground, and where nothing reaches, it plots nothing at all.
func TestNoSpeckIsInvisibleAgainstItsOwnGround(t *testing.T) {
	hours := []float64{0, 2, 4, 6, 8, 10, 11, 13, 14, 16, 17, 18, 19, 20, 20.4, 20.9, 22, 23}
	total, worst, worstAt := 0, math.MaxFloat64, ""
	for _, g := range ambientGeoms {
		for _, hr := range hours {
			c := ambientFrame(t, g[0], g[1], 7, hr, 60)
			for _, p := range ambientCells(c) {
				x, y := p[0], p[1]
				want := c.Far().Cells[y*c.W+x].R
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				if r != want {
					t.Errorf("%dx%d %04.1f (%d,%d): plotted %q, renders %q",
						g[0], g[1], hr, x, y, want, r)
					continue
				}
				if fg == bg {
					t.Errorf("%dx%d %04.1f (%d,%d): speck %q is the same cube index as its ground",
						g[0], g[1], hr, x, y, r)
				}
				d := math.Abs(bandLuma(fg) - bandLuma(bg))
				if d < ambientContrast {
					t.Errorf("%dx%d %04.1f (%d,%d): speck %q reads %.1f luma over its ground, bar is %.0f",
						g[0], g[1], hr, x, y, r, d, ambientContrast)
				}
				if d < worst {
					worst, worstAt = d, fmt.Sprintf("%dx%d %04.1f (%d,%d)", g[0], g[1], hr, x, y)
				}
				total++
			}
		}
	}
	// State the sample size. A clean result from an empty sample looks exactly
	// like a pass, and that mistake has voided three measurements here.
	if total < 1500 {
		t.Fatalf("only %d specks examined across %d geometries x %d hours -- too thin to mean anything",
			total, len(ambientGeoms), len(hours))
	}
	t.Logf("%d specks examined, faintest %.1f luma over its ground at %s", total, worst, worstAt)
}

// NO SPECK BLINKS OUT BETWEEN CONSECUTIVE FRAMES.
//
// ⚠ The old field's gate was `if twinkle <= 0.02 { continue }`, and a continue
// is not a dimmer -- it is a delete. twinkle was (0.55+0.45*sin) * StarVis, so
// below StarVis 0.2 the bottom of every speck's cycle crossed the gate and the
// speck vanished: 0.469 blink-outs a frame at 08:00, 0.480 at 10:00, 0.486 at
// 14:00, which at 12 fps is between five and six disappearances a second. That
// is his "stars appearing and dissapearing" of 2026-09-10.
//
// It is gone because eligibility is now a fact about the CELL and the hour --
// a hash against a threshold -- and the shimmer can only LIFT a speck above the
// opacity that was measured, never below it.
func TestNoSpeckBlinksOutBetweenFrames(t *testing.T) {
	// The three hours the old gate fired at, plus one either side of the 0.2
	// StarVis line it hinged on, plus the two he screenshotted.
	hours := []float64{6, 8, 9, 10, 11, 13, 14, 15, 16, 20.4, 20.9}
	frames, compared := 0, 0
	for _, g := range [][2]int{{125, 28}, {143, 27}} {
		for _, hr := range hours {
			sh := NewShore(7, false)
			tm := 0.0
			var prev map[[2]int]bool
			for k := 0; k < 180; k++ {
				c := canvas.New(g[0], g[1], canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				tm += 1.0 / 12
				sh.Update(c, tm, Activity{Level: 0.5, Working: true, ContextUsed: 0.56,
					TimeOfDay: hr / 24, TodoDone: 19, TodoTotal: 32})
				cur := map[[2]int]bool{}
				for _, p := range ambientCells(c) {
					cur[p] = true
				}
				if k > 30 { // let the shore warm before comparing
					for p := range prev {
						if !cur[p] {
							t.Errorf("%dx%d %04.1f frame %d: speck at (%d,%d) was drawn and is gone",
								g[0], g[1], hr, k, p[0], p[1])
						}
					}
					compared += len(prev)
					frames++
				}
				prev = cur
			}
		}
	}
	if frames < 3000 || compared < 50000 {
		t.Fatalf("only %d frame pairs / %d speck-observations compared -- too thin", frames, compared)
	}
	t.Logf("%d frame pairs, %d speck-observations, zero blink-outs", frames, compared)
}

// THE FIELD'S DENSITY FALLS MONOTONICALLY FROM MIDNIGHT TO NOON.
//
// This is the channel itself, and it is the house rule applied: encode in
// coverage, count or position -- never in rate. The clock used to be an alpha,
// which is a rate wearing a different coat; it is a COUNT now. Because the
// threshold moves and the hash under it does not, a daylight field is a strict
// SUBSET of the midnight one: the specks thin out and none of them moves, which
// is what keeps the sky a place rather than a screensaver.
func TestTheAmbientFieldThinsFromMidnightToNoon(t *testing.T) {
	for _, g := range ambientGeoms {
		var counts []int
		for k := 0; k <= 48; k++ { // every half hour, both ways round the clock
			c := ambientFrame(t, g[0], g[1], 7, float64(k)/2, 40)
			counts = append(counts, len(ambientCells(c)))
		}
		for k := 1; k <= 24; k++ {
			if counts[k] > counts[k-1] {
				t.Errorf("%dx%d: %04.1f has %d specks, %04.1f has %d -- the morning gets BUSIER",
					g[0], g[1], float64(k)/2, counts[k], float64(k-1)/2, counts[k-1])
			}
		}
		for k := 25; k <= 48; k++ {
			if counts[k] < counts[k-1] {
				t.Errorf("%dx%d: %04.1f has %d specks, %04.1f has %d -- the evening gets EMPTIER",
					g[0], g[1], float64(k)/2, counts[k], float64(k-1)/2, counts[k-1])
			}
		}
		// HIS RULING, 2026-09-11: "make sure we can still see some during
		// daytime, even if fainter than at nighttime." Noon is the FLOOR of
		// the range, not zero -- see ambientDayFloor. What has to hold is that
		// it really is the thinnest hour and that something is still there.
		if counts[24] == 0 {
			t.Errorf("%dx%d: noon is EMPTY, and his ruling is that some stay", g[0], g[1])
		}
		for k, n := range counts {
			if n < counts[24] {
				t.Errorf("%dx%d: %04.1f has %d specks, fewer than noon's %d -- "+
					"noon is supposed to be the thinnest hour of the day",
					g[0], g[1], float64(k)/2, n, counts[24])
			}
		}
		// Non-empty at both ends, or the monotonicity above is the trivial
		// kind: a field that is zero all day is monotone too.
		if counts[0] == 0 {
			t.Fatalf("%dx%d: midnight is empty -- nothing was measured", g[0], g[1])
		}
		if counts[0] <= counts[20] {
			t.Errorf("%dx%d: midnight %d is not richer than 10:00 %d",
				g[0], g[1], counts[0], counts[20])
		}
		// And the daylight field is a SUBSET of the night's, not a reshuffle.
		night := map[[2]int]bool{}
		for _, p := range ambientCells(ambientFrame(t, g[0], g[1], 7, 0, 40)) {
			night[p] = true
		}
		day := ambientCells(ambientFrame(t, g[0], g[1], 7, 16, 40))
		for _, p := range day {
			if !night[p] {
				t.Errorf("%dx%d: the 16:00 speck at (%d,%d) is not in the midnight field -- "+
					"the specks are MOVING, not thinning", g[0], g[1], p[0], p[1])
			}
		}
	}
}

// THE CONSTELLATION IS UNTOUCHED.
//
// The checklist owns '*' alone, it is drawn in the NEAR layer by todoStars, and
// nothing in this work may move it. The table below is every '*' cell -- column,
// row, foreground index, background index -- captured from HEAD a31a8e7 BEFORE
// the ambient field was changed, at his own 125x28 with 19 of 32 todos done.
//
// It is a golden and not a property because the coupling it guards is real but
// invisible: compositeBG blends every layer in order, so an ambient speck under
// a star whose own alpha is below 1 would tint the star's ink. Measured, the two
// fields share no cell at any of these hours -- but "they happen not to collide
// today" is exactly the kind of fact that stops being true when a density
// changes, which is what this change does.
func TestTheConstellationIsUntouchedByTheAmbientField(t *testing.T) {
	const w, h = 125, 28
	golden := []struct {
		hour  float64
		cells string
	}{
		{0.0, "22:1:110:232 31:1:247:232 47:1:110:232 105:2:153:232 120:2:111:232 11:3:110:233 21:3:153:233 68:3:147:233 20:5:147:234 24:5:147:234 54:5:153:234 78:5:111:234 4:6:153:234 106:6:153:234 70:7:147:235 110:7:153:235 10:8:153:235 51:8:110:235 114:8:153:235"},
		{6.0, "22:1:111:24 31:1:110:24 47:1:111:24 105:2:153:60 120:2:153:60 11:3:153:243 21:3:153:243 68:3:153:243 20:5:252:102 24:5:252:102 54:5:153:102 78:5:251:102 4:6:189:102 106:6:252:102 70:7:252:245 110:7:188:245 10:8:188:138 51:8:252:138 114:8:189:138"},
		{12.0, "22:1:153:25 31:1:117:25 47:1:153:25 105:2:189:25 120:2:153:25 11:3:153:32 21:3:153:32 68:3:153:32 20:5:254:75 24:5:254:75 54:5:254:75 78:5:254:75 4:6:254:75 106:6:254:75 70:7:254:110 110:7:254:110 10:8:254:110 51:8:159:110 114:8:254:110"},
		{16.0, "22:1:147:25 31:1:111:25 47:1:153:25 105:2:153:24 120:2:153:24 11:3:153:67 21:3:153:67 68:3:153:67 20:5:153:246 24:5:153:246 54:5:189:246 78:5:153:246 4:6:189:247 106:6:188:247 70:7:254:248 110:7:254:248 10:8:231:181 51:8:231:181 114:8:231:181"},
		{18.0, "22:1:111:25 31:1:111:25 47:1:147:25 105:2:189:24 120:2:153:24 11:3:147:67 21:3:153:67 68:3:153:67 20:5:252:102 24:5:252:102 54:5:189:102 78:5:251:102 4:6:189:138 106:6:252:138 70:7:188:174 110:7:189:174 10:8:188:174 51:8:188:174 114:8:189:174"},
		{20.4, "22:1:111:24 31:1:110:24 47:1:111:24 105:2:153:24 120:2:147:24 11:3:111:60 21:3:153:60 68:3:153:60 20:5:251:96 24:5:251:96 54:5:189:96 78:5:250:96 4:6:189:96 106:6:252:96 70:7:251:95 110:7:188:95 10:8:252:95 51:8:250:95 114:8:189:95"},
		{22.0, "22:1:110:234 31:1:110:234 47:1:111:234 105:2:153:235 120:2:147:235 11:3:110:235 21:3:153:235 68:3:153:235 20:5:147:237 24:5:147:237 54:5:153:237 78:5:147:237 4:6:153:237 106:6:153:237 70:7:250:238 110:7:153:238 10:8:251:239 51:8:146:239 114:8:153:239"},
	}
	for _, want := range golden {
		c := ambientFrame(t, w, h, 7, want.hour, 120)
		var got []string
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				if r == '*' {
					got = append(got, fmt.Sprintf("%d:%d:%d:%d", x, y,
						fg.Index256Keeping(), bg.Index256Keeping()))
				}
			}
		}
		if len(got) != 19 {
			t.Fatalf("%04.1f: %d stars on screen, the golden has 19 -- the sample is wrong, "+
				"not the colours", want.hour, len(got))
		}
		if s := strings.Join(got, " "); s != want.cells {
			t.Errorf("%04.1f: the constellation moved.\n  was %s\n  now %s", want.hour, want.cells, s)
		}
	}

	// The golden above is inert at 125x28 because the two fields happen to share
	// no cell there. At 143x27 they shared one and at 80x24 they shared two, and
	// under a lit star the dust is not harmless: compositeBG BLENDS the layers,
	// so an ambient speck tints the ink of a star whose own alpha is its
	// magnitude. Raising the dust's opacity moved five of those stars by a cube
	// step before stars() was taught to keep off them. This is the guarantee
	// that survives a geometry the golden does not cover.
	stars, checked := 0, 0
	for _, g := range ambientGeoms {
		for _, hr := range []float64{0, 6, 16, 18, 20.4, 22} {
			c := ambientFrame(t, g[0], g[1], 7, hr, 120)
			for y := 0; y < g[1]; y++ {
				for x := 0; x < g[0]; x++ {
					i := y*g[0] + x
					if !c.Near().Cells[i].Set || c.Near().Cells[i].R != '*' {
						continue
					}
					stars++
					if c.Far().Cells[i].Set {
						t.Errorf("%dx%d %04.1f: the star at (%d,%d) has dust under it (%q) -- "+
							"it tints the ink the checklist chose",
							g[0], g[1], hr, x, y, c.Far().Cells[i].R)
					}
				}
			}
			checked++
		}
	}
	if stars < 400 {
		t.Fatalf("only %d lit stars seen over %d frames -- too thin to mean anything",
			stars, checked)
	}
	t.Logf("%d lit stars over %d frames, none of them standing on dust", stars, checked)
}

// EVERY AMBIENT GLYPH SURVIVES THE 256 PATH, AND NONE OF THEM IS A STAR.
//
// HIS ASK, 2026-09-11: "any ideas around different characters? ... I see more
// characters varying in sizes and shapes appearing." The field had four slots
// holding three distinct shapes; it has thirteen holding nine.
//
// ⚠ The vocabulary is checked here rather than assumed from its codepoints.
// This repo has been burned by a glyph that measured wider than it looked --
// braille is 13.5% wider than ASCII in every macOS font though Unicode calls it
// Narrow -- so each of these was read out of Menlo Regular's own tables first
// (all thirteen present, all thirteen the same advance as 'M'), and each is
// rendered here through the real 256 path to prove the rune comes back out.
func TestTheAmbientVocabularyIsWideAndHoldsNoStar(t *testing.T) {
	seen := map[rune]bool{}
	for _, g := range ambientGlyphs {
		if g == '*' {
			t.Fatalf("'*' is the checklist's glyph and only the checklist's -- "+
				"a channel that shares a glyph with the scenery is not a channel (%q)", g)
		}
		seen[g] = true
	}
	if len(seen) < 8 {
		t.Errorf("the field holds %d distinct shapes; the sky he called bare had 3", len(seen))
	}
	// Each one, plotted into the far layer over a real sky, comes back out of
	// ResolveAt as itself.
	c := ambientFrame(t, 125, 28, 7, 0, 60)
	for i, g := range ambientGlyphs {
		x, y := 3+i, 2
		c.Far().Plot(x, y, g, term.RGB{R: 255, G: 255, B: 255}, 3.4)
		if r, _, _ := c.ResolveAt(x, y, term.Profile256); r != g {
			t.Errorf("%q does not survive the 256 path: it renders as %q", g, r)
		}
	}
	// And the sky he actually looks at carries most of them at once, rather
	// than the vocabulary being wide only in the source.
	for _, hr := range []float64{16, 20.4, 0} {
		f := ambientFrame(t, 125, 28, 7, hr, 120)
		shapes := map[rune]bool{}
		for _, p := range ambientCells(f) {
			shapes[f.Far().Cells[p[1]*f.W+p[0]].R] = true
		}
		if len(shapes) < 6 {
			t.Errorf("%04.1f: only %d distinct shapes on screen", hr, len(shapes))
		}
	}
}
