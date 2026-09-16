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

// TestTheSpendCounterSitsTopRightInEveryScape: with a spend known, both
// composed scapes carry the number in the top-right corner, legible on
// its own ground at every hour; with none known, the corner is untouched.
func TestTheSpendCounterSitsTopRightInEveryScape(t *testing.T) {
	luma := func(c term.RGB) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }
	read := func(c *canvas.Canvas, x, y, n int) (string, float64) {
		s, worst := "", 999.0
		for i := 0; i < n; i++ {
			r, fg, bg := c.ResolveAt(x+i, y, term.Profile256)
			s += string(r)
			if gap := luma(fg) - luma(bg); gap < worst {
				worst = gap
			}
		}
		return s, worst
	}
	for _, scapeName := range []string{"shore", "vista"} {
		t.Setenv("XSCAPES_SCAPE", scapeName)
		for _, g := range [][2]int{{80, 24}, {125, 28}, {60, 20}, {40, 12}} {
			for _, hr := range []float64{0, 0.25, 0.5, 0.78} {
				st := reduce.State{Act: scape.Activity{Working: true, Level: 0.4, ContextUsed: 0.3, TimeOfDay: hr, TodoDone: 5, Tokens: 1_234_567}, Pose: companion.Working}
				f := newFrames(g[0], g[1], 7, false, true, 0, 0)
				f.compose(st, 3.0)
				txt, x, y, ok := scape.TokensLabel(g[0], st.Act.Tokens)
				if !ok {
					t.Fatalf("%s %dx%d: no label", scapeName, g[0], g[1])
				}
				got, gap := read(f.c, x, y, len(txt))
				if f.vista != nil {
					// Under the vista's floor the owl's box reaches the
					// counter's row and the corner is the owl's.
					if _, owlY, _, _ := f.vista.Layout(); owlY <= scape.TokensRow {
						if got == txt {
							t.Errorf("vista %dx%d: the owl's box reaches row %d, yet the counter is drawn over it", g[0], g[1], scape.TokensRow)
						}
						continue
					}
				}
				if got != txt {
					t.Errorf("%s %dx%d %.2f: the corner reads %q, want %q", scapeName, g[0], g[1], hr, got, txt)
				}
				if gap < 100 {
					t.Errorf("%s %dx%d %.2f: the counter's ink is only %.0f luma over its ground", scapeName, g[0], g[1], hr, gap)
				}
				// No spend: the corner is the sky's.
				bare := newFrames(g[0], g[1], 7, false, true, 0, 0)
				st.Act.Tokens = 0
				bare.compose(st, 3.0)
				if s, _ := read(bare.c, x, y, len(txt)); s == txt {
					t.Errorf("%s %dx%d: no spend known, yet the corner reads %q", scapeName, g[0], g[1], s)
				}
			}
		}
	}
	// A star already on a counter cell keeps it: the label yields the cell.
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	txt, x, y, _ := scape.TokensLabel(80, 1_234_567)
	c.Near().Plot(x+3, y, '*', term.RGB{R: 255, G: 255, B: 255}, 1)
	drawTokens(c, 1_234_567)
	got, _ := read(c, x, y, len(txt))
	want := txt[:3] + "*" + txt[4:]
	if got != want {
		t.Errorf("with a star on the counter the corner reads %q, want %q", got, want)
	}
	_ = fmt.Sprint
}
