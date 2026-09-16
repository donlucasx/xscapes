package scenes

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestTheMotionRoundStartsFromToday: with no motion picked, the round's
// painter draws exactly what the live vista draws today, every state, every
// cell, including the working blink; and the owlets likewise.
func TestTheMotionRoundStartsFromToday(t *testing.T) {
	states := []companion.State{companion.Resting, companion.Working, companion.NeedsYou, companion.Done, companion.Worried}
	for _, st := range states {
		for _, tt := range []float64{0, 0.1, 3.0, 7.0, 7.1, 7.3, 14.05} {
			a := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			b := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			for _, c := range []*canvas.Canvas{a, b} {
				for y := 0; y < c.H; y++ {
					for x := 0; x < c.W; x++ {
						c.SetBG(x, y, term.RGB{R: 40, G: 90, B: 40})
					}
				}
			}
			DrawOwlPose(a, 3, 2, tt, st)
			DrawOwlMoving(b, 3, 2, tt, st, nil, 0)
			for y := 0; y < a.H; y++ {
				for x := 0; x < a.W; x++ {
					r1, f1, b1 := a.ResolveAt(x, y, term.Profile256)
					r2, f2, b2 := b.ResolveAt(x, y, term.Profile256)
					if r1 != r2 || f1 != f2 || b1 != b2 {
						t.Fatalf("state %v t=%.2f: cell (%d,%d) differs: today %q %v/%v, round %q %v/%v", st, tt, x, y, r1, f1, b1, r2, f2, b2)
					}
				}
			}
		}
	}
	for _, n := range []int{1, 3} {
		a := canvas.New(40, 8, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		b := canvas.New(40, 8, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		DrawOwlets(a, 36, 2, n, 0)
		DrawOwletsMoving(b, 36, 2, n, 0, 5.0, nil, 0)
		for y := 0; y < a.H; y++ {
			for x := 0; x < a.W; x++ {
				r1, f1, b1 := a.ResolveAt(x, y, term.Profile256)
				r2, f2, b2 := b.ResolveAt(x, y, term.Profile256)
				if r1 != r2 || f1 != f2 || b1 != b2 {
					t.Fatalf("%d owlets: cell (%d,%d) differs", n, x, y)
				}
			}
		}
	}
	// And every candidate draws without panicking at every quarter second of its period.
	for st, ms := range OwlMotions {
		for i := range ms {
			for u := 0.0; u < ms[i].Period+1; u += 0.25 {
				c := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				DrawOwlMoving(c, 3, 2, u, st, &ms[i], 0)
			}
		}
	}
	for i := range OwletMotions {
		for u := 0.0; u < 8; u += 0.25 {
			c := canvas.New(40, 8, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			DrawOwletsMoving(c, 36, 2, 5, 0, u, &OwletMotions[i], 0)
		}
	}
}

// TestTheSmallWingsAreTwo: a small-winged owl paints cells in BOTH of the
// body's margins. The first cut put the right wing inside the body, where
// the overlay paints nothing, and he saw one wing.
func TestTheSmallWingsAreTwo(t *testing.T) {
	for _, kind := range []int{3, 4} {
		a := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		b := canvas.New(20, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		drawOwlWith(a, 3, 2, companion.Working, OwlMod{})
		drawOwlWith(b, 3, 2, companion.Working, OwlMod{Wings: kind})
		left, right := 0, 0
		for y := 0; y < a.H; y++ {
			for x := 0; x < a.W; x++ {
				r1, f1, b1 := a.ResolveAt(x, y, term.Profile256)
				r2, f2, b2 := b.ResolveAt(x, y, term.Profile256)
				if r1 == r2 && f1 == f2 && b1 == b2 {
					continue
				}
				if x <= 3+1 {
					left++
				}
				if x >= 3+10 {
					right++
				}
			}
		}
		if left == 0 || right == 0 {
			t.Errorf("wings kind %d: %d cells changed on the left, %d on the right; both wings must paint", kind, left, right)
		}
	}
}

// TestTheFlockComesAndGoes: a newcomer is in the air for OwletFlyIn seconds
// and then sits; an exit in progress is drawn in the air; the sitting count
// never exceeds the reducer's.
func TestTheFlockComesAndGoes(t *testing.T) {
	for place := range OwletPlaces {
		was := OwletPlace
		OwletPlace = place
		var f OwletFlock
		c := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		// Two owlets long present, then a third arrives at t = 10.
		if n := DrawFlock(c, &f, 2, nil, 108, 15, 20, 46, 52, 5.0, nil, 0); n != 2 {
			t.Fatalf("place %d: %d sitting, want 2", place, n)
		}
		if n := DrawFlock(c, &f, 3, nil, 108, 15, 20, 46, 52, 10.3, nil, 0); n != 2 {
			t.Errorf("place %d: the newcomer should be in the air at t=10.3, %d sitting", place, n)
		}
		if n := DrawFlock(c, &f, 3, nil, 108, 15, 20, 46, 52, 14.0, nil, 0); n != 3 {
			t.Errorf("place %d: the newcomer should have landed by t=14, %d sitting", place, n)
		}
		// The newcomer in the air never paints a cell of the owl's face
		// (drawn before the owl, its flight starts behind it): a frame with
		// the owl drawn after the flock equals one with the owl alone at the
		// owl's own cells.
		a := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		b := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		g := OwletFlock{arrived: append([]float64(nil), f.arrived...)}
		g.arrived[2] = 20.0
		for _, tt := range []float64{20.05, 20.2, 20.4} {
			DrawFlock(a, &g, 3, nil, 108, 15, 20, 46, 52, tt, nil, 0)
			DrawOwlPose(a, 108, 15, tt, companion.Working)
			DrawOwlPose(b, 108, 15, tt, companion.Working)
			// The owl's ink and rim: columns 1..10 of its box. Columns 0 and
			// 11 are the margins the barn owl leaves open, and an owlet on
			// its way out shows there for a frame: that is the emergence.
			for y := 15; y < 22; y++ {
				for x := 109; x < 119; x++ {
					r1, f1, b1 := a.ResolveAt(x, y, term.Profile256)
					r2, f2, b2 := b.ResolveAt(x, y, term.Profile256)
					if r1 != r2 || f1 != f2 || b1 != b2 {
						t.Fatalf("place %d t=%.2f: a flying owlet shows through the owl at (%d,%d)", place, tt, x, y)
					}
				}
			}
		}
		// One leaves: the count drops to 2 and an exit is under way.
		if n := DrawFlock(c, &f, 2, []float64{0.5}, 108, 15, 20, 46, 52, 30.0, nil, 0); n != 2 {
			t.Errorf("place %d: %d sitting during an exit, want 2", place, n)
		}
		if len(f.arrived) != 2 {
			t.Errorf("place %d: the flock remembers %d arrivals, want 2", place, len(f.arrived))
		}
		OwletPlace = was
	}
}

// TestAFlightClearsTheLitter: an arriving owlet's cells never meet a sitting
// owlet's cells at any point of its flight, for every placement and every
// litter size that fits, at 125x28 and 80x24. The first path was a low arc
// and crossed the nest at the litter's own height.
func TestAFlightClearsTheLitter(t *testing.T) {
	for _, g := range []struct{ w, owlX, owlY, grassY, fireX, minX int }{{125, 108, 15, 20, 46, 52}, {80, 64, 12, 17, 30, 36}} {
		for place := range OwletPlaces {
			for n := 2; n <= 6; n++ {
				k := n - 1 // the newcomer; 0..n-2 sit
				for u := 0.0; u <= 1.0; u += 0.01 {
					x0, y0, x1, y1, ok := FlightBox(place, k, u, g.owlX, g.owlY, g.grassY, g.fireX, g.w, g.minX)
					if !ok {
						break
					}
					for j := 0; j < k; j++ {
						sx, _, sok := OwletSpot(place, j, g.owlX, g.fireX, g.w, g.minX)
						if !sok {
							continue
						}
						if x0 < sx+6 && x1 > sx && y0 < g.grassY+4 && y1 > g.grassY {
							t.Fatalf("%dx: place %d, %d owlets: the newcomer at u=%.2f (%d,%d) crosses owlet %d sitting at (%d,%d)", g.w, place, n, u, x0, y0, j, sx, g.grassY)
						}
					}
				}
			}
		}
	}
}

// TestThePickedOwlNeverGrows: in every state, at every instant of its
// locked motion, the owl paints nothing outside its own 12-column box and
// the one-cell rim around it. His ruling of 2026-09-16, "small wings for
// both": done's flap uses the working owl's small wings, so no picked motion
// needs the 14-column art (whose ink reaches a column past the box and whose
// rim a column past that). Cells outside are compared against the same owl
// drawn with only the motion's hop, so a hop cannot pass for a wing.
func TestThePickedOwlNeverGrows(t *testing.T) {
	states := []companion.State{companion.Resting, companion.Working, companion.NeedsYou, companion.Done, companion.Worried}
	const ox, oy = 4, 3
	ground := term.RGB{R: 40, G: 90, B: 40}
	for _, st := range states {
		m := PickedOwlMotion(st)
		if m == nil || m.At == nil {
			continue
		}
		for u := 0.0; u < m.Period; u += 0.05 {
			mod := m.At(u)
			a := canvas.New(24, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			b := canvas.New(24, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			for _, c := range []*canvas.Canvas{a, b} {
				for y := 0; y < c.H; y++ {
					for x := 0; x < c.W; x++ {
						c.SetBG(x, y, ground)
					}
				}
			}
			drawOwlWith(a, ox, oy, st, OwlMod{Hop: mod.Hop})
			drawOwlWith(b, ox, oy, st, mod)
			for y := 0; y < a.H; y++ {
				for x := 0; x < a.W; x++ {
					if x >= ox-1 && x < ox+12+1 {
						continue
					}
					r1, f1, b1 := a.ResolveAt(x, y, term.Profile256)
					r2, f2, b2 := b.ResolveAt(x, y, term.Profile256)
					if r1 != r2 || f1 != f2 || b1 != b2 {
						t.Fatalf("state %v at %.2f s (wings kind %d): cell (%d,%d) painted outside the owl's 12-column box and its rim, columns %d..%d", st, u, mod.Wings, x, y, ox-1, ox+12)
					}
				}
			}
		}
	}
}

// TestThePicksAreByName: the locked owl and litter are the candidates he
// named, not whichever index happens to sit where. Found by a parallel
// session's audit of session 36: the working pick was 8, and index 8 in
// source order is "long hop" (round 2), while the motion he asked for and
// the record describes, "looks, blink, flutter", is index 7. Nothing
// asserted the picks by name, so the live vista drew the wrong motion under
// the right caption.
func TestThePicksAreByName(t *testing.T) {
	want := map[companion.State]string{
		companion.Resting:  "peek",
		companion.Working:  "looks, blink, flutter",
		companion.NeedsYou: "double blink, wave",
		companion.Done:     "flap and bounce",
		companion.Worried:  "squint, brows",
	}
	for st, name := range want {
		m := PickedOwlMotion(st)
		if m == nil {
			t.Errorf("state %v: no motion picked; want %q", st, name)
			continue
		}
		if m.Name != name {
			t.Errorf("state %v: picked index %d is %q; the ruling is %q", st, OwlMotionPick[st], m.Name, name)
		}
	}
	if m := PickedOwletMotion(); m == nil || m.Name != "the litter's life" {
		got := "<nil>"
		if m != nil {
			got = m.Name
		}
		t.Errorf("owlets: picked index %d is %q; the ruling is %q", OwletMotionPick, got, "the litter's life")
	}
}
