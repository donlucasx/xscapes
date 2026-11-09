package scenes

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

// By day the disc paints no unlit face: nothing it touches gets darker than
// the sky it sits on (the shore's ruling after his "the Sun seems to break
// sometimes").
func TestTheHarboursDayDiscHasNoDarkFace(t *testing.T) {
	for _, tc := range []struct {
		tod      float64
		day      bool
		wantDark bool
	}{{0.42, true, false}, {0.5, true, false}, {0.6, true, false}} {
		p := overcast(scape.PaletteAt(tc.tod))
		c := canvas.New(40, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		fill(c, 0, 0, 39, 11, p.SkyHorizon)
		drawDiscAt(c, p, 0.6, 15, 5, tc.day)
		dark := false
		for y := 0; y < 12; y++ {
			for x := 0; x < 40; x++ {
				if luma(c.BGAt(x, y)) < luma(p.SkyHorizon)-8 {
					dark = true
				}
			}
		}
		if dark != tc.wantDark {
			t.Errorf("tod %.2f: a face darker than the sky = %v, want %v", tc.tod, dark, tc.wantDark)
		}
	}
}
