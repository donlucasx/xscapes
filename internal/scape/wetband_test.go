package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// The wet strip at the water's edge is the beach's own staircase, one tone
// down, and it never leaves the cube.
//
// He photographed the alternative on 2026-09-09: flat grey blocks along the
// shore. The strip was the interpolated palette value, and the cube has no warm
// colour below luma 96 -- its channel levels are 0, 95, 135, 175, 215, 255, so
// under 95 in green and blue there is only zero, while the grey ramp steps by
// ten the whole way down. Palette.WetSand runs luma 52 to 108, so it sat under
// that floor for most of the day and the nearest colour the terminal owns is a
// grey. Worse than a constant: warm at 09:00, grey at 10:00, warm 11:00 to
// 14:00, grey from 15:00 -- the strip changed hue back and forth through a
// working morning. See notes/wetsand.
func TestTheWetStripIsOneRungUnderTheBeach(t *testing.T) {
	seen := map[term.RGB]bool{}
	warmHours := 0
	for h := 0; h < 48; h++ {
		p := PaletteAt(float64(h) / 48)
		dry, wet := writeBandColor(p), wetBandColor(p)

		// Cube-exact, both of them, so quantisation cannot move either.
		for _, c := range []term.RGB{dry, wet} {
			if got := term.FromIndex256(c.Index256Keeping()); got != c {
				t.Fatalf("half-hour %d: %v is not a cube entry -- it draws as %v", h, c, got)
			}
		}
		// The strip has to read as WET: darker than the sand it sits on.
		if bandLuma(wet) >= bandLuma(dry) {
			t.Errorf("half-hour %d: the wet strip %v (luma %.0f) is not darker than the beach %v (luma %.0f)",
				h, wet, bandLuma(wet), dry, bandLuma(dry))
		}
		// And warm wherever the beach still has warmth to give. The beach's
		// last warm tone is the cube's own floor; under it the strip goes
		// neutral, which is his ruling and is consistent because by then the
		// whole scene is dark.
		if dry != sandNight {
			if wet.R == wet.G && wet.G == wet.B {
				t.Errorf("half-hour %d: the beach is %v but the strip is neutral %v", h, dry, wet)
			}
			warmHours++
		}
		seen[wet] = true
	}
	if warmHours == 0 {
		t.Fatal("no half-hour has a warm beach, so this test proves nothing")
	}
	// Three tones, like the beach. A fourth would mean the staircase had come
	// unstuck from the one it is supposed to follow.
	if len(seen) > 3 {
		t.Errorf("the strip takes %d distinct tones over the day, want at most 3: %v", len(seen), seen)
	}
	t.Logf("%d of 48 half-hours have a warm beach, and the strip is warm at every one of them; %d tones in all",
		warmHours, len(seen))
}
