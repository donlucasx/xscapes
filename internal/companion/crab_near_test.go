package companion

import (
	"fmt"
	"hash/fnv"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/envx"
	"github.com/donlucasx/xscapes/internal/term"
)

// withNear arms a rung for one test and puts the flag back. The package var is
// read from the environment once at init, exactly as scape.Tide is, so a test
// sets the var rather than the environment.
func withNear(t *testing.T, rung int) {
	t.Helper()
	old := Near
	Near = rung
	t.Cleanup(func() { Near = old })
}

// nearCrab is a crab standing at the near pose, ready to draw.
func nearCrab(mirror bool) *Cat {
	c := NewCrab()
	c.FaceLeft(mirror)
	c.approach = 1
	return c
}

// his window widths, from the 30 real recordings the litter ladder was folded
// through. The companion's column is derived from the width, so a defect that
// only bites at one geometry is a defect that only bites at one of his windows.
var hisWidths = []int{80, 111, 124, 143, 153}

// hisStates is every pose the companion has. All five are drawn at every rung,
// because the RETREAT runs through the ladder wearing whatever pose the agent
// moved on to.
var hisStates = []State{Resting, Working, NeedsYou, Done, Worried}

// ---------------------------------------------------------------------------

// Nothing moves with the near pose unset.
//
// This is the promise the whole feature is gated on: with XSCAPES_NEAR unset,
// every frame is byte-identical to the one that shipped before it existed. The
// hashes below were taken from a pristine `git archive HEAD` export of 31264bc,
// before a line of crab_near.go was written: 256 frames per state -- both
// animals, both facings, stepping and not, four canvas geometries and sixteen
// times, 2560 frames in all -- covering the rune, the colour, the alpha, the
// per-cell background and the set flag of every cell.
//
// ⚠ It is deliberately a hash of the PAINTED CELLS and not a claim about the
// source. A refactor that "obviously cannot change the picture" is exactly the
// kind of thing this project has been burned by; crabBreathPeriod came out of
// drawCrab while this feature was built, and this is what says it came out
// clean.
func TestNothingMovesWithTheNearPoseUnset(t *testing.T) {
	withNear(t, 0)
	golden := []struct{ name, sum string }{
		{"crab/resting/256 frames", "75e6f283f7b133e5"},
		{"crab/working/256 frames", "0913d492b66b12e5"},
		{"crab/needs you/256 frames", "7cc91b37f3727ba5"},
		{"crab/all done/256 frames", "87b4f7aef6d72b85"},
		{"crab/something is broken/256 frames", "178daff7432fda85"},
		{"cat/resting/256 frames", "80f5d044664fe09d"},
		{"cat/working/256 frames", "45ea37f9272f30c5"},
		{"cat/needs you/256 frames", "b7d10105a2f0b4c1"},
		{"cat/all done/256 frames", "f2034baf926d86b5"},
		{"cat/something is broken/256 frames", "5844edee4034b395"},
	}
	got := shippedFrameHashes(t)
	if len(got) != len(golden) {
		t.Fatalf("the matrix changed shape: %d groups against %d goldens", len(got), len(golden))
	}
	for i, g := range golden {
		if got[i].name != g.name {
			t.Fatalf("group %d is %q, golden is %q -- the matrix moved under the golden", i, got[i].name, g.name)
		}
		if got[i].sum != g.sum {
			t.Errorf("%s renders %s, before this feature it rendered %s -- the shipped path is NOT byte-identical",
				g.name, got[i].sum, g.sum)
		}
	}
}

// And an armed flag on its own is not enough either: until the companion has
// actually walked a rung, an armed XSCAPES_NEAR still draws today's frame. That
// matters because the flag is set for a whole session and the ask is a median
// 130 seconds of it.
func TestAnArmedFlagAloneStillDrawsTheShippedCrab(t *testing.T) {
	for _, rung := range []int{1, 2} {
		withNear(t, rung)
		if dx, w, h := NewCrab().DrawnBox(200); dx != 0 || w != 12 || h != 7 {
			t.Errorf("XSCAPES_NEAR=%d, approach 0: box is (%d, %d, %d), want (0, 12, 7)", rung, dx, w, h)
		}
		a := frameOf(t, NewCrab(), NeedsYou, 2.5)
		withNear(t, 0)
		b := frameOf(t, NewCrab(), NeedsYou, 2.5)
		if a != b {
			t.Errorf("XSCAPES_NEAR=%d changes the frame before the walk has started", rung)
		}
	}
}

// OFF unless the environment asks. Both halves: the flag's own parse, and the
// package variable that init() left behind in this very process.
func TestTheNearPoseIsOffUnlessTheEnvironmentAsksForIt(t *testing.T) {
	for _, c := range []struct {
		v    string
		want int
	}{{"", 0}, {"0", 0}, {"off", 0}, {"no", 0}, {"true", 0}, {"3", 0}, {"", 0}, {"1", 1}, {"2", 2}} {
		if got := nearFromEnv(c.v); got != c.want {
			t.Errorf("XSCAPES_NEAR=%q arms rung %d, want %d", c.v, got, c.want)
		}
	}
	if v := envx.Lookup("NEAR"); v != "" {
		t.Skipf("XSCAPES_NEAR is %q in this shell, so the default cannot be read here", v)
	}
	if Near != 0 {
		t.Fatalf("XSCAPES_NEAR is unset and the ladder is armed at rung %d", Near)
	}
}

// Size() keeps returning the 12x7 box at every rung. It is the LAYOUT contract
// -- compose() derives the companion's margin, the litter's room, SandTo and
// the moon's column from it -- and a near pose that changed it would move the
// sand's right bound and the balloon with it.
func TestSizeStillReturnsTheShippedBoxAtEveryRung(t *testing.T) {
	for _, rung := range []int{0, 1, 2} {
		withNear(t, rung)
		c := nearCrab(true)
		if w, h := c.Size(); w != 12 || h != 7 {
			t.Errorf("rung %d: Size() is %dx%d, want 12x7", rung, w, h)
		}
	}
}

// The head column is the same at every rung, at every one of his widths.
//
// Asserted on the PAINTED FRAME, not on the arithmetic: the whole feature rests
// on an even-width sprite drawn at catX-(W-12)/2 keeping the shipped head
// column, and "I intended to centre it" is not evidence. The eye cells are
// found by their colour and their mean column compared, which is where the
// balloon's pointer and HeadCol() both aim.
// THE BALLOON'S POINTER FOLLOWS THE HEAD, at every rung and in both facings.
//
// ⚠ THIS REPLACES TestTheHeadColumnIsTheSameAtEveryRung, and the guarantee it
// asserted is GONE ON PURPOSE. That test said the head column never moves,
// which was true while each rung was CENTRED on the shipped head -- a cheap
// trick that kept bubbleX, HeadCol and compose() untouched.
//
// His ruling of 2026-09-10 ended it: "Id keep the companion away from the far
// right edge of the screen, needs some negative space when it comes closer."
// Centring spends half of every new column to the RIGHT, into the margin, so
// the gap to the frame edge ran 9 / 7 / 3 as the animal approached -- less air
// the closer it got. The sprite now hangs entirely into the scene, which buys
// the margin back and MOVES the head.
//
// So the invariant is no longer "the head does not move". It is this: wherever
// the head goes, the pointer goes with it. DrawnHeadCol is what live.go anchors
// the balloon to, and it must land between the painted eyes -- read off the
// rendered frame, not off the arithmetic that produced it.
func TestTheBalloonsPointerFollowsTheHeadAtEveryRung(t *testing.T) {
	for _, w := range hisWidths {
		for _, mirror := range []bool{true, false} {
			x := w - 12 - (2 + w/32)
			if !mirror {
				x = 5
			}
			for _, rung := range []int{0, 1, 2} {
				withNear(t, rung)
				c := canvas.New(w, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				crab := nearCrab(mirror)
				crab.Draw(c.Near(), x, 24-2-7, 2.5, NeedsYou)
				lo, hi, n := 1<<30, -1, 0
				for i, cell := range c.Near().Cells {
					if cell.Set && cell.FG == EyeAlert {
						col := i % w
						n++
						if col < lo {
							lo = col
						}
						if col > hi {
							hi = col
						}
					}
				}
				if n == 0 {
					t.Fatalf("w=%d mirror=%v rung=%d: no alert eye on the frame at all", w, mirror, rung)
				}
				got := x + crab.DrawnHeadCol(w)
				if got < lo || got > hi {
					t.Errorf("w=%d mirror=%v rung=%d: the pointer would go to column %d, "+
						"outside the eyes at %d..%d -- the balloon speaks from bare sand",
						w, mirror, rung, got, lo, hi)
				}
			}
		}
	}
}

// GROWING NEVER COSTS THE COMPANION ITS AIR ON THE FRAME EDGE.
//
// His ruling, 2026-09-10, on the first live look: "Id keep the companion away
// from the far right edge of the screen, needs some negative space when it
// comes closer." It is the same note he made at 124 columns in session 15.
//
// Measured before the fix, the gap between the rightmost ink and the frame edge
// ran BACKWARDS -- 9 / 7 / 3 at his window, and 3 / 1 / 0 at forty columns,
// where the biggest rung touched the edge outright. The bar is that it never
// shrinks; nearDX's quarter-of-the-growth term is what makes it open instead.
func TestComingCloserNeverCostsTheCompanionItsMargin(t *testing.T) {
	for _, w := range hisWidths {
		x := w - 12 - (2 + w/32)
		prev := -1
		for _, rung := range []int{0, 1, 2} {
			withNear(t, rung)
			c := canvas.New(w, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			crab := nearCrab(true)
			crab.Draw(c.Near(), x, 24-2-7, 2.5, NeedsYou)
			hi := -1
			for i, cell := range c.Near().Cells {
				if cell.Set && i%w > hi {
					hi = i % w
				}
			}
			if hi < 0 {
				t.Fatalf("w=%d rung=%d: nothing drawn at all", w, rung)
			}
			gap := w - 1 - hi
			if prev >= 0 && gap < prev {
				t.Errorf("w=%d rung %d leaves %d columns of air, down from %d at the rung below -- "+
					"it is being pushed INTO the edge as it comes closer", w, rung, gap, prev)
			}
			t.Logf("w=%-4d rung %d: %d columns of air", w, rung, gap)
			prev = gap
		}
	}
}

// The sprite's TOP ROW never descends into the sea, at any rung.
//
// This is the real ceiling, and it is why the near poses grow DOWNWARD from a
// fixed top row rather than standing on the shipped crab's feet. The beach is
// 7 to 8 rows at his widths and the tide buys none of them back, so a 14-row
// sprite standing on its feet would put its head five rows into the water. He
// allowed the opposite instead -- "OK if a part of its body is cut on the
// bottom of the screen" -- so the top row holds still and the legs run off the
// frame.
//
// Asserted as: no ink above y, at any rung, in any state, at any breath phase,
// facing either way. (plotRim's cleared ring is not ink -- it is the one-cell
// seam every sprite in this package already gets, and the shipped crab draws it
// at y-1 too.)
func TestTheNearPoseNeverReachesAboveItsTopRow(t *testing.T) {
	const w, h, x, y = 40, 24, 20, 15
	for _, rung := range []int{0, 1, 2} {
		withNear(t, rung)
		for _, st := range hisStates {
			for _, mirror := range []bool{true, false} {
				for _, tt := range []float64{0, 0.4, 1.1, 2.5, 3.7, 5.29} {
					c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					crab := nearCrab(mirror)
					crab.Draw(c.Near(), x, y, tt, st)
					dx, bw, bh := crab.DrawnBox(w)
					for i, cell := range c.Near().Cells {
						if !cell.Set || !isCrabInk(cell) {
							continue
						}
						cx, cy := i%w, i/w
						if cy < y {
							t.Fatalf("rung %d %v t=%.2f mirror=%v: ink at row %d, above the top row %d -- that is the sea",
								rung, st, tt, mirror, cy, y)
						}
						if cy >= y+bh || cx < x+dx || cx >= x+dx+bw {
							t.Fatalf("rung %d %v t=%.2f mirror=%v: ink at (%d,%d) outside the box DrawnBox promises (dx=%d %dx%d at %d,%d)",
								rung, st, tt, mirror, cx, cy, dx, bw, bh, x, y)
						}
					}
				}
			}
		}
	}
}

// Every state at a rung draws the same footprint. DrawnBox is measured off one
// pose at startup, so a state whose art was a row taller would hand the layout
// a box the sprite then overflows.
func TestEveryStateAtARungDrawsTheSameBox(t *testing.T) {
	for _, rung := range []int{1, 2} {
		withNear(t, rung)
		for _, stepping := range []bool{true, false} {
			for _, st := range hisStates {
				c := nearCrab(true)
				c.SetStepping(stepping)
				p := c.nearPoseFor(rung, st, 1.3)
				q := ParseBitmap(append(append([]string{}, p.upper...), p.lower...)).ToQuadrant()
				_, bw, bh := c.DrawnBox(200)
				if got, want := len([]rune(q[0])), bw; got != want {
					t.Errorf("rung %d %v step=%v: %d cells wide, DrawnBox says %d", rung, st, stepping, got, want)
				}
				if got, want := len(q), bh; got != want {
					t.Errorf("rung %d %v step=%v: %d rows tall, DrawnBox says %d", rung, st, stepping, got, want)
				}
			}
		}
	}
}

// The alert eye's hole renders as COAT, not as sea.
//
// The ring is what says wide open -- a solid block is a blob -- but canvas.Plot
// REPLACES a cell, so the hole in the middle of the ring is the cell's own
// background and shows whatever is behind the crab. That is his "two blue
// holes" report of 2026-09-04 at four times the area, and the fix is that every
// cell of the eye pass goes through PlotOn with a ground, spaces included.
//
// Rendered over a deliberate sea, and read back through the same ResolveAt and
// the same 256 profile the live frame is painted with, because the claim is
// about what reaches the screen.
func TestTheAlertEyesHoleIsTheCoatAndNotTheSea(t *testing.T) {
	sea := term.RGB{R: 40, G: 90, B: 140}
	wantBG := term.Profile256.Quantise(CrabCoat, false)
	seaBG := term.Profile256.Quantise(sea, false)
	if wantBG == seaBG {
		t.Fatalf("the probe is void: the coat and the sea quantise to the same tone %v", wantBG)
	}
	for _, rung := range []int{1, 2} {
		withNear(t, rung)
		for _, mirror := range []bool{true, false} {
			const w, h, x, y = 40, 20, 14, 4
			c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			for i := range c.BG {
				c.BG[i] = sea
			}
			crab := nearCrab(mirror)
			// t=2.5 is a frame with the eye open; t=0 is a blink.
			crab.Draw(c.Near(), x, y, 2.5, NeedsYou)
			n := 0
			for i, cell := range c.Near().Cells {
				if !cell.Set || cell.FG != EyeAlert {
					continue
				}
				n++
				cx, cy := i%w, i/w
				if _, _, bg := c.ResolveAt(cx, cy, term.Profile256); bg != wantBG {
					t.Errorf("rung %d mirror=%v: the alert eye cell at (%d,%d) renders on %v, want the coat %v (the sea is %v)",
						rung, mirror, cx, cy, bg, wantBG, seaBG)
				}
			}
			if n != 8 {
				t.Errorf("rung %d mirror=%v: %d alert eye cells, want 8 (two eyes, 2 cells by 2)", rung, mirror, n)
			}
		}
	}
}

// The alert eye is a RING, and it is a different SHAPE from the working eye.
//
// At 1x the ask eye is 'O' where working is 'o' -- the SAME CELL -- so the alert
// has only ever read by colour and by the raised claw, never by size. At 2 cells
// by 2 it can finally be a shape, and that is a channel the feature gets for
// free. The ring specifically: a solid block is a blob, and the HOLE is what
// says wide open.
func TestTheAlertEyeIsARingAndNotABlob(t *testing.T) {
	for _, rung := range []int{1, 2} {
		withNear(t, rung)
		eyeRunes := func(st State, want term.RGB) []rune {
			const w, h, x, y = 40, 20, 14, 4
			c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			crab := nearCrab(true)
			crab.Draw(c.Near(), x, y, 2.5, st) // 2.5 is not a blink frame
			var out []rune
			for _, cell := range c.Near().Cells {
				if cell.Set && cell.FG == want {
					out = append(out, cell.R)
				}
			}
			return out
		}
		alert := eyeRunes(NeedsYou, EyeAlert)
		open := eyeRunes(Working, eyeCol)
		if len(alert) != 8 || len(open) != 8 {
			t.Fatalf("rung %d: %d alert cells and %d working cells, want 8 each", rung, len(alert), len(open))
		}
		solid := 0
		for _, r := range alert {
			if r == '█' {
				solid++
			}
		}
		if solid == len(alert) {
			t.Errorf("rung %d: the alert eye is %q -- eight solid cells is a blob, not a ring", rung, string(alert))
		}
		if string(alert) == string(open) {
			t.Errorf("rung %d: the alert eye and the working eye draw the same glyphs %q, so ask still reads by colour alone",
				rung, string(alert))
		}
		t.Logf("rung %d: alert %q, working %q", rung, string(alert), string(open))
	}
}

// The approach reaches the near pose and comes back, and it is monotonic.
//
// The retreat is not a detail: the instant he answers is the one frame he is
// certainly looking at, and Worried outranks NeedsYou (reduce.go:606), so a
// main-thread error during an ask takes the pose away while the animal is still
// standing at the near size. It has to walk back, not cut.
func TestTheApproachWalksAllTheWayThereAndAllTheWayBack(t *testing.T) {
	for _, rung := range []int{1, 2} {
		withNear(t, rung)
		c := nearCrab(true)
		c.approach = 0
		const dt = 0.05
		last, seen := 0.0, []int{}
		for i := 0; i < 200 && c.approach < 1; i++ {
			c.Approach(dt, true)
			if c.approach < last {
				t.Fatalf("rung %d: the walk went backwards, %.3f after %.3f", rung, c.approach, last)
			}
			last = c.approach
			if r := c.nearRung(200); len(seen) == 0 || seen[len(seen)-1] != r {
				seen = append(seen, r)
			}
		}
		if c.approach != 1 || c.nearRung(200) != rung {
			t.Fatalf("rung %d: the walk stalled at %.3f, drawing rung %d", rung, c.approach, c.nearRung(200))
		}
		// Every rung is stood on, in order: that is what makes it read as
		// steps rather than as a zoom.
		if len(seen) != rung+1 {
			t.Errorf("rung %d: the walk passed through rungs %v, want one stop on each of 0..%d", rung, seen, rung)
		}
		for i, r := range seen {
			if r != i {
				t.Errorf("rung %d: the walk went through %v, out of order", rung, seen)
				break
			}
		}
		// And the whole walk takes one nearStep per rung -- pace.go's stepDur,
		// which is the beat the companion already walks on.
		if want := float64(rung) * nearStep; !within(float64(len(seen)-1)*nearStep, want, 1e-9) {
			t.Errorf("rung %d: %d steps, want %d", rung, len(seen)-1, rung)
		}

		back := 0
		for i := 0; i < 200 && c.approach > 0; i++ {
			prev := c.approach
			c.Approach(dt, false)
			if c.approach > prev {
				t.Fatalf("rung %d: the retreat went forwards, %.3f after %.3f", rung, c.approach, prev)
			}
			back++
		}
		if c.approach != 0 || c.nearRung(200) != 0 {
			t.Fatalf("rung %d: the retreat stalled at %.3f, drawing rung %d", rung, c.approach, c.nearRung(200))
		}
		if dx, bw, bh := c.DrawnBox(200); dx != 0 || bw != 12 || bh != 7 {
			t.Errorf("rung %d: home is the box (%d, %d, %d), want the shipped (0, 12, 7)", rung, dx, bw, bh)
		}
		t.Logf("rung %d: %.2f s out, %.2f s back", rung, float64(len(seen)-1)*nearStep, float64(back)*dt)
	}
}

// A stalled frame must not teleport the walk: at most one rung per frame,
// however long the gap. A resize, a compaction or a closed lid delivers a dt
// of whatever it likes, and the shore hit this same class of bug once already
// by scaling elapsed time instead of advancing phase.
func TestOneSlowFrameCannotSkipARung(t *testing.T) {
	withNear(t, 2)
	c := NewCrab() // at home, where a walk starts
	c.Approach(600, true)
	if got := c.nearRung(200); got != 1 {
		t.Errorf("a 600 s frame put the crab on rung %d, want 1 -- one rung per frame", got)
	}
}

// The cat never approaches. There is no cat art at this size, and the root
// package calls Approach and DrawnBox on whichever animal is configured.
func TestTheCatIsUntouchedByTheNearPose(t *testing.T) {
	for _, rung := range []int{0, 1, 2} {
		withNear(t, rung)
		c := NewCat()
		for i := 0; i < 40; i++ {
			c.Approach(0.1, true)
		}
		if c.approach != 0 || c.nearRung(200) != 0 {
			t.Errorf("rung %d: the cat walked to %.2f", rung, c.approach)
		}
		w, h := c.Size()
		if dx, bw, bh := c.DrawnBox(200); dx != 0 || bw != w || bh != h {
			t.Errorf("rung %d: the cat's box is (%d, %d, %d), want (0, %d, %d)", rung, dx, bw, bh, w, h)
		}
	}
}

// The picture itself, for the record and for the next person to look at rather
// than re-derive. Also an assertion: each rung has to put strictly more ink on
// the screen than the one below it, which is the one thing "closer" has to mean.
func TestEachRungIsBiggerThanTheOneBelowIt(t *testing.T) {
	// 64 wide, not 44: nearFits refuses a rung wider than half the picture, so
	// at 44 columns rung 2 is capped back to rung 1 and the two rungs render
	// identically -- the test would be measuring the CAP, not the ladder. This
	// is a picture of the art, so it gets a frame every rung can stand in.
	const w, h, x, y = 64, 22, 20, 6
	prev := 0
	for _, rung := range []int{0, 1, 2} {
		withNear(t, rung)
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		crab := nearCrab(true)
		crab.Draw(c.Near(), x, y, 2.5, NeedsYou)
		ink := 0
		for _, cell := range c.Near().Cells {
			if cell.Set && isCrabInk(cell) {
				ink++
			}
		}
		dx, bw, bh := crab.DrawnBox(w)
		t.Logf("rung %d: %d cells of ink, box dx=%d %dx%d", rung, ink, dx, bw, bh)
		for row := y; row < h; row++ {
			line := ""
			for col := 0; col < w; col++ {
				cell := c.Near().Cells[row*w+col]
				switch {
				case !cell.Set:
					line += "."
				case cell.FG == EyeAlert:
					line += "O"
				case cell.R == ' ':
					line += " "
				default:
					line += string(cell.R)
				}
			}
			t.Logf("  |%s|", line)
		}
		if ink <= prev {
			t.Errorf("rung %d puts %d cells of ink on the beach, no more than the %d below it", rung, ink, prev)
		}
		prev = ink
	}
}

// ---------------------------------------------------------------------------

func isCrabInk(cell canvas.Cell) bool {
	return cell.FG == CrabCoat || cell.FG == EyeAlert || cell.FG == eyeCol || cell.FG == eyeWorried
}

func within(a, b, eps float64) bool { return a-b < eps && b-a < eps }

func frameOf(t *testing.T, c *Cat, st State, tt float64) string {
	t.Helper()
	const w, h = 40, 20
	cv := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	c.FaceLeft(true)
	c.Draw(cv.Near(), 24, 10, tt, st)
	s := ""
	for _, cell := range cv.Near().Cells {
		s += fmt.Sprintf("%q|%v|%v|%v;", cell.R, cell.Set, cell.FG, cell.BG)
	}
	return s
}

// shippedFrameHashes folds 256 frames per state into one hash each: both
// animals, both facings, stepping and not, four canvas geometries and sixteen
// times. Grouped by state rather than hashed whole so a failure names the pose
// that moved.
func shippedFrameHashes(t *testing.T) []struct{ name, sum string } {
	t.Helper()
	var out []struct{ name, sum string }
	sizes := []struct{ w, h, x, y int }{
		{30, 14, 8, 4},
		{80, 20, 64, 11},
		{124, 24, 110, 15},
		{40, 9, 26, 1}, // short: the companion hangs off the bottom
	}
	// Fifteen times, not six, and the density was earned: at six samples a
	// change of the working breath from 2.2 s to 2.3 s did not move a single
	// frame, because the lift is a BOOLEAN (sin > 0.35) and both periods land
	// on the same side of it at all six. A golden that cannot fail is the exact
	// trap this repo has been burned by, so the grid now straddles the lift's
	// own edges (0.13 and 0.99 sit either side of it), the blink's 0.16 s window
	// on the 5.3 s cycle (0.165 is inside it by five thousandths) and the
	// resting claw's 7.2 s alternation. Each of those three was then broken by
	// hand and the golden confirmed RED.
	times := []float64{0, 0.13, 0.165, 0.31, 0.7, 0.99, 1.3, 1.7, 2.5, 3.05, 3.7, 4.4, 5.29, 5.31, 6.1, 7.3}
	for _, animal := range []string{NameCrab, NameCat} {
		for _, st := range hisStates {
			h := fnv.New64a()
			n := 0
			for _, mirror := range []bool{true, false} {
				for _, stepping := range []bool{true, false} {
					for _, sz := range sizes {
						for _, tt := range times {
							c := canvas.New(sz.w, sz.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
							a := New(animal)
							a.FaceLeft(mirror)
							a.SetStepping(stepping)
							a.Draw(c.Near(), sz.x, sz.y, tt, st)
							for _, cell := range c.Near().Cells {
								fmt.Fprintf(h, "%q|%v|%v|%.3f|%v|%v;", cell.R, cell.Set, cell.FG, cell.A, cell.BG, cell.HasBG)
							}
							n++
						}
					}
				}
			}
			out = append(out, struct{ name, sum string }{
				fmt.Sprintf("%s/%v/%d frames", animal, st, n),
				fmt.Sprintf("%016x", h.Sum64()),
			})
		}
	}
	return out
}
