package main

import (
	"fmt"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The arriving star, measured on the COMPOSED frame.
//
// internal/scape/arrival_test.go proves the flight path is a path the sky would
// give a star. It cannot prove the star is on his screen, and that is a
// different question with a different answer: drawScene paints the context
// readout and an OPAQUE ask balloon into the same near layer the stars live in,
// AFTER the scape has drawn. A star under either is simply deleted.
//
// That is not hypothetical here. It is the defect his 2026-09-10 report bought a
// fix for -- the readout was landing on 50 stars -- and the balloon was caught
// eating a companion's head in the same session. The fall puts a star through
// SIX MORE CELLS than the layout certifies, so it has to be read back off the
// rendered frame or it is not measured at all.
//
// Every assertion below is read with ResolveAt after drawScene. Nothing here
// trusts the layer.
type fallCase struct {
	label    string
	w, h     int
	seed     int64
	tod, ctx float64
	mirror   bool
	bubble   string
	ask      bool
	total    int
}

// fallCases varies every axis, and not as one cross product: the study's own
// population is every geometry against every hour, then ONE AXIS AT A TIME for
// seed, context, mirror and balloon. A full cross product of this many full
// composed frames would take minutes and would buy nothing the margins do not.
func fallCases() []fallCase {
	const total = 19
	var out []fallCase
	for _, g := range [][2]int{{125, 28}, {143, 27}, {107, 51}, {80, 24}, {60, 20}, {40, 12}} {
		// ⚠ 0.583 IS THE BINDING HOUR, 14:00. It is the only hour on the study's
		// whole sweep where the ink search exhausts on a path cell: the sky is
		// at its brightest and white stops winning. Leave it on.
		for _, tod := range []float64{0, 0.583, 0.78} {
			out = append(out, fallCase{
				label: fmt.Sprintf("%dx%d tod %.2f", g[0], g[1], tod),
				w:     g[0], h: g[1], seed: 7, tod: tod, ctx: 0.45,
				mirror: true, bubble: "allow Bash?", ask: true, total: total,
			})
		}
	}
	base := fallCase{label: "", w: 125, h: 28, seed: 7, tod: 0, ctx: 0.45,
		mirror: true, bubble: "allow Bash?", ask: true, total: total}
	add := func(label string, f func(*fallCase)) {
		c := base
		f(&c)
		c.label = label
		out = append(out, c)
	}
	add("seed 1", func(c *fallCase) { c.seed = 1 })
	add("seed 42", func(c *fallCase) { c.seed = 42 })
	// Both ends of the context range: the readout is absent below 40% used and
	// present for the rest of the session, and the disc sinks the whole way.
	add("ctx 0.00", func(c *fallCase) { c.ctx = 0 })
	add("ctx 0.93", func(c *fallCase) { c.ctx = 0.93 })
	add("unmirrored", func(c *fallCase) { c.mirror = false })
	add("no balloon", func(c *fallCase) { c.bubble, c.ask = "", false })
	add("done knock", func(c *fallCase) { c.bubble, c.ask = "tests passed", false })
	add("full list", func(c *fallCase) { c.total = 32 })
	return out
}

// fallFrame composes one whole frame the way the live loop does and returns the
// constellation as it actually rendered.
//
// ONE warmed Shore, thirty steps: a fresh Shore has no wave clock and no tide,
// and the tide sets the star band's floor through foamCeiling. Every frame of a
// case is warmed identically, which is what makes two of them comparable.
func fallFrame(c fallCase, done int, arriving bool, phase float64) (map[[2]int]bool, *scape.Shore) {
	cv := canvas.New(c.w, c.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(c.seed, false)
	cat := companion.New(companion.DefaultName)
	cat.FaceLeft(c.mirror)
	ccw, chh := cat.Size()
	lay := compose(c.w, ccw, c.mirror)
	sh.MoonX = lay.MoonX

	st := reduce.State{
		Act: scape.Activity{
			Level: 0.5, Working: true, ContextUsed: c.ctx, TimeOfDay: c.tod,
			TodoDone: done, TodoTotal: c.total,
			Arriving: arriving, ArrivalPhase: phase,
		},
		Pose: companion.Working,
	}
	if c.bubble != "" {
		st.Bubble, st.BubbleAsk = c.bubble, c.ask
		if c.ask {
			st.Pose = companion.NeedsYou
		} else {
			st.Pose = companion.Done
		}
	}
	for i := 0; i < 30; i++ {
		sh.Update(cv, 2.5+float64(i)/20, st.Act)
	}
	sh.Update(cv, 4, st.Act)
	drawScene(cv, sh, cat, lay, st, 4, c.seed, cv.H-2-chh)

	// '*' belongs to the constellation alone -- "a channel that shares a glyph
	// with the scenery is not a channel" -- so every one in the sky is one unit
	// of finished work. The sand is excluded by the row bound, since a tool name
	// on the beach may well contain one.
	hy := int(float64(c.h)*0.42) + 1
	if hy > cv.H {
		hy = cv.H
	}
	out := map[[2]int]bool{}
	for y := 0; y < hy; y++ {
		for x := 0; x < cv.W; x++ {
			if r, _, _ := cv.ResolveAt(x, y, term.Profile256); r == '*' {
				out[[2]int{x, y}] = true
			}
		}
	}
	return out, sh
}

// The sky is never wrong about the count while a star is falling, and the star
// is on screen for every frame of the fall.
//
// THE COUNT IS THE WHOLE CHANNEL. A constellation says "this session has got
// through n", and the fall is the one thing in the scene that takes a star out
// of its place -- so if anything on the way in is eaten by the balloon, or lands
// on a star already lit, the sky reads n-1 or n-2 for half a second. The bar is
// exact: settled stars plus one, never off by one in either direction.
func TestTheFallingStarIsOnScreenTheWholeWay(t *testing.T) {
	// ⚠ The phases are the ones the reducer can actually hand over. 0.99 is the
	// last frame of a fall and it is where the head reaches home; the settled
	// frame after it draws the same cell, which is what makes the landing one
	// picture and not two.
	phases := []float64{0, 0.5, 0.99}
	arrivals, skipped := 0, 0
	for _, c := range fallCases() {
		for done := 1; done <= c.total; done += 3 {
			settled, _ := fallFrame(c, done-1, false, 0)
			lit, _ := fallFrame(c, done, false, 0)

			// The home cell, taken by DIFFING the product's own two frames
			// rather than by asking the layout. It follows whatever the
			// placement does today and cannot drift from it.
			var home [2]int
			fresh := 0
			for p := range lit {
				if !settled[p] {
					home, fresh = p, fresh+1
				}
			}
			if fresh != 1 {
				// Small geometries draw fewer stars than the count for reasons
				// that predate this feature -- 40x12 draws 21 of 32 -- and
				// charging that to the fall was the first sweep's own error.
				skipped++
				continue
			}
			arrivals++

			for _, ph := range phases {
				fall, sh := fallFrame(c, done, true, ph)
				if len(fall) != len(settled)+1 {
					t.Fatalf("%s done %d phase %.2f: %d stars on screen, want %d -- the count is wrong mid-fall",
						c.label, done, ph, len(fall), len(settled)+1)
				}
				var head [2]int
				extra := 0
				for p := range fall {
					if !settled[p] {
						head, extra = p, extra+1
					}
				}
				// Every settled star still there, and exactly one cell added.
				// Both halves matter: a head that landed ON a settled star
				// would keep the total right while deleting a star.
				if extra != 1 {
					t.Fatalf("%s done %d phase %.2f: %d cells differ from the settled sky, want exactly the head",
						c.label, done, ph, extra)
				}
				for p := range settled {
					if !fall[p] {
						t.Fatalf("%s done %d phase %.2f: settled star %v is gone -- the fall displaced it",
							c.label, done, ph, p)
					}
				}
				if sh.DiscCovers(head[0], head[1]) {
					t.Fatalf("%s done %d phase %.2f: the head %v is on the disc, which carries context",
						c.label, done, ph, head)
				}
				if ph == 0.99 && head != home {
					t.Fatalf("%s done %d: the fall ended on %v, not on the star's own place %v",
						c.label, done, head, home)
				}
				if ph == 0 && head == home && c.w > 60 {
					// Not fatal on its own -- a refused track legitimately
					// arrives in place -- but at his widths it should be rare,
					// and internal/scape holds the rate to 95%.
					t.Logf("%s done %d: launched from home (no track)", c.label, done)
				}
			}
		}
	}
	t.Logf("%d arrivals composed and read back; %d skipped (the settled sky drew no new star)", arrivals, skipped)
	if arrivals < 100 {
		t.Fatalf("only %d arrivals measured -- the sweep is not measuring the feature", arrivals)
	}
}
