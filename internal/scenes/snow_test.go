package scenes

import (
	"testing"
)

// TestThePickedSnowIsByName holds his pick of 2026-09-16 ("i like S1") by
// its name rather than its index, the lesson of the owl's working pick,
// which was off by one for a day.
func TestThePickedSnowIsByName(t *testing.T) {
	if SnowPick < 0 || SnowPick >= len(SnowStyles) {
		t.Fatalf("SnowPick %d is not a style", SnowPick)
	}
	if got, want := SnowStyles[SnowPick].Name, "light cover, subtle"; got != want {
		t.Errorf("SnowPick %d is %q, his pick is %q", SnowPick, got, want)
	}
}

// TestTheSnowStaysApart walks the day every half hour and holds, as the
// terminal draws them (after the cube), the seams his note of 2026-09-16
// named: "the different layers of the mountain range merge at some times of
// the day", "the contrast should be more subtle between the snow and the
// mountains". Measured before the fix: the far and mid ranges met at 23:00
// (both grey 237) and the far range met the lit face at 06:00 and 19:00
// (both entry 173); today's snow ran 33 to 120 over the face. Held: the far
// range 15 over both the mid body and the lit face, the snow 20 over the far
// range, and the snow 35 to 75 over the face.
func TestTheSnowStaysApart(t *testing.T) {
	for hh := 0.0; hh < 24; hh += 0.5 {
		rt := VistaRangeTones(SnowPick, hh/24)
		far, mid, face, snow := qluma(rt.Far), qluma(rt.Mid), qluma(rt.MidLit), qluma(rt.Snow)
		if far-mid < 15 || far-face < 15 {
			t.Errorf("%05.1fh: the far range (%.0f) is within 15 of the mid range (%.0f) or its lit face (%.0f)", hh, far, mid, face)
		}
		if snow-far < 20 {
			t.Errorf("%05.1fh: the snow (%.0f) is within 20 of the far range (%.0f)", hh, snow, far)
		}
		if d := snow - face; d < 35 || d > 75 {
			t.Errorf("%05.1fh: the snow is %.0f over the lit face; 35 to 75 is subtle", hh, d)
		}
	}
}
