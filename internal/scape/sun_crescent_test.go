package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// By DAY the disc is the sun, and a sun has no unlit face. His report of
// 2026-09-06, from long live sessions: "the Sun seems to break sometimes, and
// fix itself - mostly has to do when the context depletes, the shadow
// underneath". It was the terminator: lit = 1 - ContextUsed, so the shadow
// disc slides across as the context fills and snaps clear at a compaction,
// and by day that unlit face is a slate mass hanging off the sun -- measured
// in his 134x71 frame at 49%, the lit body ends at x677 and the slate runs on
// to x719. His ruling: the sun WANES AS A CRESCENT.
//
// At night the moon keeps its shaded face; that is what makes the disc read
// as a whole body rather than a bitten one, and it is not what he reported.
func TestTheSunWanesAsACrescentButTheMoonKeepsItsShadedFace(t *testing.T) {
	const w, h = 130, 22
	const used = 0.6 // enough shadow that the far limb is well inside it

	render := func(tod float64, edge string) (*canvas.Canvas, *Shore) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(7, false)
		sh.MoonX = 0.28
		sh.MoonEdge = edge
		sh.Update(c, 2, Activity{ContextUsed: used, TimeOfDay: tod})
		return c, sh
	}
	// The far limb: the disc's rightmost column at a full disc, which the
	// terminator has swallowed by 60%.
	farLimb := func(tod float64) (x, y int) {
		c, sh := render(tod, "")
		control, _ := render(tod, "none")
		mx, my := sh.MoonPos()
		rx, _ := sh.MoonExtent()
		for dx := rx + 1; dx > 0; dx-- {
			if mx+dx >= w {
				continue
			}
			ch, fg, bg := c.ResolveAt(mx+dx, my, term.Profile256)
			ch0, fg0, bg0 := control.ResolveAt(mx+dx, my, term.Profile256)
			if ch != ch0 || fg != fg0 || bg != bg0 {
				return mx + dx, my
			}
		}
		t.Fatalf("tod %.2f: no disc found on the moon's row", tod)
		return 0, 0
	}
	painted := func(tod string, todv float64, x, y int) bool {
		c, _ := render(todv, "")
		control, _ := render(todv, "none")
		ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
		ch0, fg0, bg0 := control.ResolveAt(x, y, term.Profile256)
		t.Logf("%s cell (%d,%d): disc %q %v/%v  sky %q %v/%v", tod, x, y, string(ch), fg, bg, string(ch0), fg0, bg0)
		return ch != ch0 || fg != fg0 || bg != bg0
	}

	// Midday. The far limb sits inside the terminator, so nothing should be
	// painted there: the sun is a crescent open to the shadow, not a disc
	// with a slate half.
	const midday = 0.5
	x, y := farLimb(0.05) // the extent, taken at a phase that shows it
	if _, sh := render(midday, ""); sh.night() {
		t.Fatal("tod 0.5 is not day; the test is pointing at the wrong hour")
	}
	if painted("midday", midday, x, y) {
		t.Errorf("midday: the sun's far limb at (%d,%d) is painted at %.0f%% context; "+
			"the unlit face is still being drawn on the sun", x, y, used*100)
	}

	// Midnight. The moon keeps its shaded face, so the same cell IS painted.
	const midnight = 0.0
	if _, sh := render(midnight, ""); !sh.night() {
		t.Fatal("tod 0.0 is not night; the test is pointing at the wrong hour")
	}
	if !painted("midnight", midnight, x, y) {
		t.Errorf("midnight: the moon's far limb at (%d,%d) is NOT painted at %.0f%% context; "+
			"the shaded face has gone from the moon as well", x, y, used*100)
	}
}
