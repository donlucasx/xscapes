package scenes

import (
	"fmt"
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestTheVistaClearsBetweenFrames: sixty frames on ONE vista leave the sky
// with no more glyphs than a single frame paints, and a flying owlet's old
// cells are gone the frame after. Before the fix the sky's count climbed
// with every frame (the wind's debris stayed where each frame put it) and
// the flight drew a trail.
func TestTheVistaClearsBetweenFrames(t *testing.T) {
	act := scape.Activity{Working: true, Level: 0.65, ContextUsed: 0.4, TimeOfDay: 0.424}
	skyGlyphs := func(c *canvas.Canvas, rows int) int {
		n := 0
		for y := 0; y < rows; y++ {
			for x := 0; x < c.W; x++ {
				if r, _, _ := c.ResolveAt(x, y, term.Profile256); r != ' ' && r != '▀' && r != '▄' && r != '▌' && r != '▐' && r != '▖' && r != '▗' && r != '▘' && r != '▝' && r != '▙' && r != '▛' && r != '▜' && r != '▟' {
					n++
				}
			}
		}
		return n
	}
	v := NewVista(7, false)
	c := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	v.Update(c, 3.0, act)
	one := skyGlyphs(c, 10)
	for i := 1; i <= 60; i++ {
		v.Update(c, 3.0+float64(i)/12, act)
	}
	many := skyGlyphs(c, 10)
	if many > one+one/2 {
		t.Errorf("after 60 frames the sky holds %d glyphs against %d in one frame: the layers are not cleared", many, one)
	}
	// A flight leaves no trail: frame B on the same vista equals frame B on a fresh one.
	var f OwletFlock
	owlX, owlY, grassY, _ := v.Layout()
	v.Update(c, 10.0, act)
	DrawFlock(c, &f, 1, nil, owlX, owlY, grassY, v.FireX(), 52, 10.0, nil, 0)
	f.Sync(2, 10.0) // a second arrives at t = 10
	v.Update(c, 10.4, act)
	DrawFlock(c, &f, 2, nil, owlX, owlY, grassY, v.FireX(), 52, 10.4, nil, 0)
	v.Update(c, 11.0, act)
	DrawFlock(c, &f, 2, nil, owlX, owlY, grassY, v.FireX(), 52, 11.0, nil, 0)
	fresh := NewVista(7, false)
	d := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	fresh.Update(d, 11.0, act)
	g := OwletFlock{arrived: append([]float64(nil), f.arrived...)}
	DrawFlock(d, &g, 2, nil, owlX, owlY, grassY, fresh.FireX(), 52, 11.0, nil, 0)
	diff := 0
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			r1, f1, b1 := c.ResolveAt(x, y, term.Profile256)
			r2, f2, b2 := d.ResolveAt(x, y, term.Profile256)
			if r1 != r2 || f1 != f2 || b1 != b2 {
				diff++
			}
		}
	}
	if diff != 0 {
		t.Errorf("a flight left %d stale cells behind after one frame", diff)
	}
}

// TestTheArcIsAGauge: the live vista's body (the arc, style 5) is two flat
// colours with a level between them: at 0% used the centre column is the
// full colour top to bottom; at 50% its upper half is the empty colour and
// its lower half the full one; at 95% it is the empty colour but for the
// bottom. By day and by night, read off the drawn cells.
func TestTheArcIsAGauge(t *testing.T) {
	luma := func(c term.RGB) float64 { return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B) }
	column := func(ctx, tod float64) (top, bottom float64) {
		v := NewVista(7, false)
		c := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		v.Update(c, 3.0, scape.Activity{ContextUsed: ctx, TimeOfDay: tod})
		cx, cy, r := v.lay.moonCenter(v.MoonStyle, ctx)
		// Cells well inside the disc, a row in from either edge, so no
		// quarter-cell rim is read.
		_, _, t1 := c.ResolveAt(int(cx), int(cy-r+1.2), term.Profile256)
		_, _, b1 := c.ResolveAt(int(cx), int(cy+r-1.2), term.Profile256)
		return luma(t1), luma(b1)
	}
	// And no sky inside the disc at any level: every cell of the centre
	// column strictly inside the disc resolves to the two colours only. The
	// row where the level crossed a cell came out as sky (the quad kept the
	// sky ramp as its ground).
	for _, tod := range []float64{0.424, 0.917} {
		// Up to 80%: past that the body is setting behind the near ridge,
		// which rightly covers it.
		for ctx := 0.05; ctx <= 0.8; ctx += 0.05 {
			v := NewVista(7, false)
			c := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			v.Update(c, 3.0, scape.Activity{ContextUsed: ctx, TimeOfDay: tod})
			cx, cy, r := v.lay.moonCenter(v.MoonStyle, ctx)
			full, empty := term.RGB{R: 255, G: 255, B: 135}, term.RGB{R: 175, G: 135, B: 95}
			if tod > 0.9 {
				full, empty = term.RGB{R: 238, G: 238, B: 238}, term.RGB{R: 118, G: 118, B: 118}
			}
			q := func(col term.RGB) term.RGB { return term.Profile256.Quantise(col, false) }
			for y := int(cy - r); y <= int(cy+r)+1; y++ {
				if y < 0 || y >= c.H {
					continue
				}
				// Only cells whose four quarters all lie inside the disc.
				inside := true
				for _, d := range [][2]float64{{-0.25, -0.25}, {0.25, -0.25}, {-0.25, 0.25}, {0.25, 0.25}} {
					px, py := float64(int(cx))+0.5+d[0]-cx, float64(y)+0.5+d[1]-cy
					if (px/2)*(px/2)+py*py >= r*r {
						inside = false
					}
				}
				if !inside {
					continue
				}
				_, fg, bg := c.ResolveAt(int(cx), y, term.Profile256)
				for _, col := range []term.RGB{fg, bg} {
					if col != q(full) && col != q(empty) {
						t.Fatalf("tod %.3f used %.2f: cell (%d,%d) of the gauge resolves to %v, neither full %v nor empty %v", tod, ctx, int(cx), y, col, q(full), q(empty))
					}
				}
			}
		}
		top0, bot0 := column(0, tod)
		top5, bot5 := column(0.5, tod)
		top9, bot9 := column(0.8, tod)
		if math.Abs(top0-bot0) > 4 {
			t.Errorf("tod %.3f fresh: top %.1f and bottom %.1f differ; the disc should be one colour, full", tod, top0, bot0)
		}
		if top5 >= bot5-20 {
			t.Errorf("tod %.3f half: top %.1f is not darker than bottom %.1f; the level should be at the middle", tod, top5, bot5)
		}
		if top9 >= top0-20 || bot9 >= bot0-20 {
			t.Errorf("tod %.3f spent: top %.1f/%.1f bottom %.1f/%.1f; the disc should be the empty colour", tod, top9, top0, bot9, bot0)
		}
	}
}

// TestTheTreelineIsNeverRaisedBehindTheOwl: the near treeline may dip to a
// flat top behind the owl's head (whole cells under the ears), but it is
// never RAISED to it. His screenshot at a 120x30 window, 2026-09-16: a
// black square behind the owl, gone as the window grew. The painter set the
// treeline's top to the owl's head row across the box; in the 80x24 study
// that is a dip in a taller treeline, and live at a short height, where the
// head sits above the trees, the same line built a tower of treeline up to
// it. Measured on the rendered frame at 10:10: the treeline's top row in
// each column of the box, against the same column with the owl placed far
// away, so the natural silhouette is read off the painter itself.
func TestTheTreelineIsNeverRaisedBehindTheOwl(t *testing.T) {
	const tod = 0.424
	act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.45, TimeOfDay: tod}
	luma := func(c term.RGB) int { return (299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000 }
	// The treeline at 10:10 is grey 48 (luma 48); the sky's top is luma 71
	// and everything else above the band is brighter.
	const treeline = 60
	top := func(c *canvas.Canvas, x, until int) int {
		for y := 0; y < until && y < c.H; y++ {
			if _, _, bg := c.ResolveAt(x, y, term.Profile256); luma(bg) < treeline {
				return y
			}
		}
		return until
	}
	worst, worstAt := 0, ""
	for _, g := range [][2]int{{120, 11}, {120, 13}, {120, 14}, {120, 16}, {133, 20}, {131, 24}, {125, 28}, {121, 33}, {60, 20}, {80, 24}} {
		w, h := g[0], g[1]
		a := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		v := NewVista(7, false)
		v.OwlX = w - 16
		v.Update(a, 3.0, act)
		owlX, owlY, _, bandTop := v.Layout()
		b := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		far := NewVista(7, false)
		far.OwlX = 4
		far.Update(b, 3.0, act)
		until := min(bandTop, owlY+7)
		for x := owlX - 1; x <= owlX+12 && x < w; x++ {
			natural, got := top(b, x, until), top(a, x, until)
			if got < natural {
				if natural-got > worst {
					worst, worstAt = natural-got, fmt.Sprintf("%dx%d column %d: treeline top row %d, natural %d", w, h, x, got, natural)
				}
			}
		}
	}
	if worst > 0 {
		t.Fatalf("the treeline is raised behind the owl by up to %d rows: %s", worst, worstAt)
	}
}

// TestTheFireIsTheWorkToo: live, the work is the wind AND the fire, his
// word of 2026-09-16 ("how hard its working is represented by the wind +
// fire (both)"): at full stretch the fire has more flame cells and more
// sparks than at rest. And the light stays the night's: the ground of every
// cell above the band is the same at rest and at full stretch, so a busy
// agent does not brighten the meadow (the cost that lost the fire-only
// variant on 09-14, kept out on purpose).
func TestTheFireIsTheWorkToo(t *testing.T) {
	const W, H = 125, 28
	const night = 22.0 / 24
	frame := func(level float64) (*canvas.Canvas, *Vista) {
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		v := NewVista(7, false)
		v.OwlX = W - 16
		v.Update(c, 3.0, scape.Activity{Working: true, Level: level, ContextUsed: 0.45, TimeOfDay: night})
		return c, v
	}
	rest, v := frame(0)
	full, _ := frame(1)
	fireX := v.FireX()
	_, _, _, bandTop := v.Layout()
	flameInk := map[term.RGB]bool{cube(255, 95, 0): true, cube(255, 175, 0): true, cube(255, 215, 135): true}
	flames := func(c *canvas.Canvas) int {
		n := 0
		near := c.Near()
		for y := 0; y < bandTop; y++ {
			for x := fireX - 8; x <= fireX+8; x++ {
				if cell := near.Cells[y*W+x]; cell.Set && cell.R != ' ' && flameInk[cell.FG] {
					n++
				}
			}
		}
		return n
	}
	sparks := func(c *canvas.Canvas) int {
		n := 0
		mid := c.Layers[1]
		for y := 0; y < bandTop; y++ {
			for x := fireX - 8; x <= fireX+8; x++ {
				if cell := mid.Cells[y*W+x]; cell.Set && cell.R == '.' && cell.FG == cube(255, 215, 135) {
					n++
				}
			}
		}
		return n
	}
	if flames(full) <= flames(rest) {
		t.Errorf("flame cells at full stretch %d, at rest %d; the fire must climb with the work", flames(full), flames(rest))
	}
	if sparks(full) <= sparks(rest) {
		t.Errorf("sparks at full stretch %d, at rest %d; the fire must throw more at full stretch", sparks(full), sparks(rest))
	}
	for y := 0; y < bandTop; y++ {
		for x := 0; x < W; x++ {
			if rest.BGAt(x, y) != full.BGAt(x, y) {
				t.Fatalf("the ground at (%d,%d) is %v at rest and %v at full stretch: the light must be the night's, not the work's", x, y, rest.BGAt(x, y), full.BGAt(x, y))
			}
		}
	}
	t.Logf("flames %d → %d, sparks %d → %d, the ground unchanged", flames(rest), flames(full), sparks(rest), sparks(full))
}

