package scape

import (
	"fmt"
	"os"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The right population is glyphs that are MEANT to carry colour. Foam, fur,
// stars and the moon are near-neutral by design and belong in the greys;
// counting them as failures measures the wrong thing, and the first version of
// this test did exactly that and reported 59.3%.
func TestBoostClearsTheRampForColouredGlyphs(t *testing.T) {
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	for i := 0; i < 20; i++ {
		sh.Update(c, float64(i)/20, Activity{Working: true, Level: 0.7})
	}
	chroma := func(c term.RGB) int {
		mx, mn := int(c.R), int(c.R)
		for _, v := range []int{int(c.G), int(c.B)} {
			if v > mx {
				mx = v
			}
			if v < mn {
				mn = v
			}
		}
		return mx - mn
	}
	out := "\n  boost   of glyphs meant to be coloured, how many land on one\n"
	var atDefault float64
	for _, k := range []float64{1.0, 1.8, 2.6, 3.4} {
		n, cube := 0, 0
		for _, l := range c.Layers {
			for _, cell := range l.Cells {
				if !cell.Set || chroma(cell.FG) < 30 {
					continue // meant to be neutral
				}
				n++
				if cell.FG.Saturate(k).Index256() < 232 {
					cube++
				}
			}
		}
		pct := 100 * float64(cube) / float64(n)
		mark := ""
		if k == term.GlyphBoost {
			mark, atDefault = "   <- default", pct
		}
		out += fmt.Sprintf("   %.1fx    %5.1f%%  (%d glyphs)%s\n", k, pct, n, mark)
	}
	os.Stderr.WriteString(out)
	if atDefault < 95 {
		t.Errorf("only %.1f%% of coloured glyphs clear the grey ramp at the default", atDefault)
	}
}
