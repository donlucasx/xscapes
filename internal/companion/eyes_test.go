package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// With the eyes filled, the two eye cells sit on fur at every hour; without,
// they show the scene. His 12:49 screengrab: each eye cell was the sea's
// colour, two blue holes in the head by day.
func TestFilledEyesSitOnFur(t *testing.T) {
	sea := term.RGB{R: 103, G: 153, B: 228}
	for _, fill := range []string{"", EyeFillCoat, EyeFillSocket} {
		c := canvas.New(40, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for i := range c.BG {
			c.BG[i] = sea
		}
		cat := NewCat()
		cat.FaceLeft(true)
		cat.SetEyeFill(fill)
		cw, ch := cat.Size()
		x, y := 40-cw-4, 14-2-ch
		cat.Draw(c.Near(), x, y, 1.0, Working)
		eyes := 0
		for yy := 0; yy < c.H; yy++ {
			for xx := 0; xx < c.W; xx++ {
				cell := c.Near().Cells[yy*c.W+xx]
				if !cell.Set || cell.R != 'o' || cell.FG != eyeCol {
					continue
				}
				eyes++
				_, _, bg := c.ResolveAt(xx, yy, term.Profile256)
				onSea := bg == term.Profile256.Quantise(sea, false)
				switch {
				case fill == "" && !onSea:
					t.Errorf("fill %q: eye at %d,%d is on %v, want the sea", fill, xx, yy, bg)
				case fill != "" && onSea:
					t.Errorf("fill %q: eye at %d,%d still shows the sea", fill, xx, yy)
				case fill == EyeFillCoat && bg != term.Profile256.Quantise(cat.coat, false):
					t.Errorf("fill coat: eye at %d,%d is on %v, want the coat", xx, yy, bg)
				}
			}
		}
		if eyes != 2 {
			t.Errorf("fill %q: found %d eyes, want 2", fill, eyes)
		}
	}
}
