package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// His standalone windows of 2026-09-18, 132x41 and 71x39, showed two things
// the 24-row scape under an agent never does: the pines standing rows above
// the lake, and a blank in the ranges behind the owl that moved with it on a
// resize. Both were the 80x24 numbers surviving where every other element
// follows the layout.

// tallState is a working vista with the context at 45% and no litter.
func tallState(tod float64) reduce.State {
	return reduce.State{Act: scape.Activity{Working: true, Level: 0.6, ContextUsed: 0.45, TimeOfDay: tod, TodoDone: 3}, Pose: companion.Working}
}

// TestTheTreelineStaysWhereTheOwlIsBelowIt: the treeline dips behind the
// owl's head only where the head is in the trees. Where the owl's top row
// is at or below the near range's base, the range in the owl's span is the
// same as it would be with the owl elsewhere.
func TestTheTreelineStaysWhereTheOwlIsBelowIt(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "vista")
	for _, g := range [][2]int{{132, 41}, {71, 39}, {124, 30}, {125, 28}, {80, 24}} {
		W, H := g[0], g[1]
		with := renderVista(t, W, H, tallState(22.0/24), 3.0)
		owlX, owlY, _, _ := with.vista.Layout()
		lakeTop, _, _ := with.vista.Rows()
		// The same frame with the owl's span moved to the far left, so the
		// right-hand ranges carry no dip at all.
		f := newFrames(W, H, 7, false, true, 0, 0)
		f.vista.OwlX = 2
		f.vista.Update(f.c, 3.0, tallState(22.0/24).Act)
		differ := 0
		for y := max(0, lakeTop-4); y < lakeTop; y++ {
			for x := owlX - 1; x < owlX+13 && x < W; x++ {
				_, _, a := with.c.ResolveAt(x, y, term.Profile256)
				_, _, b := f.c.ResolveAt(x, y, term.Profile256)
				if a != b {
					differ++
				}
			}
		}
		if 2*owlY >= 2*lakeTop {
			if differ != 0 {
				t.Errorf("%dx%d: the owl's top row %d is below the near range's base %d, yet %d cells of the range behind it differ from the range with the owl elsewhere", W, H, owlY, lakeTop, differ)
			}
		} else if differ == 0 {
			t.Errorf("%dx%d: the owl's head is in the trees (top row %d, range base %d) and nothing dipped", W, H, owlY, lakeTop)
		}
	}
}

// TestThePinesStandOnTheMeadow: the pines' feet are on the meadow's last row
// at every height, and their tips scale with the scene, so at 41 rows the
// tree is not standing on the sky.
func TestThePinesStandOnTheMeadow(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "vista")
	for _, g := range [][2]int{{132, 41}, {71, 39}, {125, 28}, {80, 24}, {60, 20}} {
		W, H := g[0], g[1]
		f := renderVista(t, W, H, tallState(0.5), 3.0)
		_, _, bandTop := f.vista.Rows()
		bg := func(x, y int) term.RGB { _, _, b := f.c.ResolveAt(x, y, term.Profile256); return b }
		// The first pine's trunk is sub-column 9, cell 4; the meadow beside
		// it, well clear of both pines, at cell 40 (or the fire's far side).
		foot := bandTop - 1
		if bg(4, foot) == bg(W-30, foot) {
			t.Errorf("%dx%d: no pine on the meadow's last row %d: cell 4 is the meadow's own colour", W, H, foot)
		}
	}
	// At 132x41 the pines used to stand at rows 6..20 whatever the height;
	// scaled, the first tip is near row 12 and row 7 is sky.
	f := renderVista(t, 132, 41, tallState(0.5), 3.0)
	bg := func(x, y int) term.RGB { _, _, b := f.c.ResolveAt(x, y, term.Profile256); return b }
	if bg(4, 7) != bg(40, 7) {
		t.Errorf("132x41: a pine at row 7, above where the scaled tip should be")
	}
	if bg(4, 16) == bg(40, 16) {
		t.Errorf("132x41: no pine at row 16, where the scaled tree should be")
	}
}

// TestLivePinsTheLevel: -live -level holds one level instead of cycling the
// demo, which is the instrument for the A/B on his glass.
func TestLivePinsTheLevel(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "vista")
	f := newFrames(80, 24, 7, false, true, 0, 0.5)
	now := time.Now()
	a, b := f.state(now, 1.0), f.state(now, 9.0)
	if a.Act.Level == b.Act.Level && a.Pose == b.Pose {
		t.Fatalf("the demo cycle did not cycle between t=1 and t=9: level %v pose %v both times", a.Act.Level, a.Pose)
	}
	f.level = 1
	a, b = f.state(now, 1.0), f.state(now, 9.0)
	if a.Act.Level != 1 || b.Act.Level != 1 || !a.Act.Working || a.Pose != companion.Working {
		t.Errorf("pinned at 1: got level %v/%v working %v pose %v", a.Act.Level, b.Act.Level, a.Act.Working, a.Pose)
	}
	f.level = 0
	a = f.state(now, 1.0)
	if a.Act.Level != 0 || a.Act.Working || a.Pose != companion.Resting {
		t.Errorf("pinned at 0: got level %v working %v pose %v", a.Act.Level, a.Act.Working, a.Pose)
	}
}
