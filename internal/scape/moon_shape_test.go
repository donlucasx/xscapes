package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The moon is a disc at every height, never a rectangle.
//
// His report, 2026-09-04, from a 124x52 window: "the moon looks worse than
// before". Measured (notes/moonprobe): at the 22 or 23 rows that window gives
// the scape, the disc's radius is 1.92 rows -- just under two -- so the rows
// that would have been its narrow tips fall outside it and the three that
// remain are all seven cells wide. A rectangle. At 27 rows the radius is two
// and the tips come back. The gradient work had not touched it; the shape was
// a function of the window height alone, and it was wrong at his.
//
// Read from the rendered frame: a cell counts as moon if it is brighter than
// the sky beside it, and a split cell (U+2580) counts by the half that is.
func TestTheMoonIsRoundAtEveryHeight(t *testing.T) {
	for h := 18; h <= 30; h++ {
		c := canvas.New(124, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(7, false)
		sh.MoonX = 0.28
		sh.Update(c, 3.0, Activity{Working: true, Level: 0.55, TimeOfDay: 0.0245, ContextUsed: 0.06})
		mx, my := sh.MoonPos()
		bright := func(col term.RGB, y int) bool {
			ref := lumaOf(c.BGAt(2, y))
			return lumaOf(col)-ref > 40
		}
		// Moon coverage per row, in half-cells, over a box wider than any disc;
		// and per row how many cells are FULL moon and how many are halves.
		var rows, fulls, halves []int
		var ys []int
		for y := my - 5; y <= my+5; y++ {
			if y < 0 || y >= c.H {
				continue
			}
			n, full, half := 0, 0, 0
			for x := mx - 10; x <= mx+10; x++ {
				if x < 0 || x >= c.W {
					continue
				}
				ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
				switch {
				case ch == '▀':
					if bright(fg, y) != bright(bg, y) {
						half++
					}
					if bright(fg, y) {
						n++
					}
					if bright(bg, y) {
						n++
					}
				case bright(bg, y):
					n += 2
					full++
				}
			}
			if n > 0 {
				rows = append(rows, n)
				fulls = append(fulls, full)
				halves = append(halves, half)
				ys = append(ys, y)
			}
		}
		if len(rows) < 3 {
			t.Errorf("h=%d: the moon covers only %d rows (%v) -- too small to be a disc", h, len(rows), rows)
			continue
		}
		widest := 0
		for _, n := range rows {
			if n > widest {
				widest = n
			}
		}
		top, bottom := rows[0], rows[len(rows)-1]
		t.Logf("h=%d: moon rows %v (half-cells per row, rows %d..%d)", h, rows, ys[0], ys[len(ys)-1])
		if top >= widest || bottom >= widest {
			t.Errorf("h=%d: the moon's top row is %d half-cells wide and its bottom %d against a widest of %d -- that is a rectangle, not a disc",
				h, top, bottom, widest)
		}
		// No pip: a tip row made of half cells must not carry a lone full
		// cell in its middle (the second report, 133x27: "the moon not
		// looking so good"). A flat top of several full cells with half-cell
		// shoulders is a disc; one full cell between halves is a spike.
		for _, k := range []int{0, len(rows) - 1} {
			if fulls[k] == 1 && halves[k] > 0 {
				t.Errorf("h=%d: row %d of the moon is one full cell between %d half cells -- a pip, not a curve", h, ys[k], halves[k])
			}
		}
	}
}
