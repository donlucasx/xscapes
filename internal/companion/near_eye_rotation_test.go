package companion

import (
	"math"
	"testing"
)

// All four faces must be reachable, and roughly evenly. A hash that clumps
// would leave one of them never drawn, which is the same class of defect as the
// constellation channel that had never once lit.
func TestEveryRotatingEyeIsReachable(t *testing.T) {
	counts := make([]int, len(nearEyeRotation))
	const trials = 4000
	c := NewCrab()
	for i := 0; i < trials; i++ {
		c.askN = i
		counts[c.eyePick()]++
	}
	for i, n := range counts {
		if n == 0 {
			t.Errorf("%s is never drawn", nearEyeNames[i])
		}
		share := 100 * float64(n) / trials
		if share < 20 || share > 30 {
			t.Errorf("%s gets %.1f%% of asks, want roughly 25%%", nearEyeNames[i], share)
		}
	}
}

// THE PROPERTY THAT DECIDES WHETHER THIS READS AS CHARACTER OR AS A GLITCH:
// the eye must not change during an ask. It is a function of the ask ordinal,
// so a thousand frames of the same ask must all draw the same face.
func TestTheEyeIsConstantWithinOneAsk(t *testing.T) {
	withNear(t, 2)
	c := NewCrab()
	c.Approach(0.1, true) // the rising edge: ask #1
	want := nearEyeRotation[c.eyePick()]

	// ⚠ EXCLUDE THE BLINK, and t = 0 IS A BLINK. nearEye swaps in the lid
	// whenever math.Mod(t, 5.3) < 0.16, which is true at exactly t = 0 -- the
	// degenerate frame this package's own test notes warn about. The first
	// version of this test took its baseline there and then reported the
	// rotation as unstable at frame 4, when what it had captured was the lid.
	blinks, checked := 0, 0
	for i := 0; i < 1000; i++ {
		c.Approach(0.05, true) // still the same ask
		tt := float64(i) / 20
		got, _ := c.nearEye(NeedsYou, tt)
		if math.Mod(tt, 5.3) < 0.16 {
			blinks++
			continue
		}
		checked++
		for r := range got {
			if got[r] != want[r] {
				t.Fatalf("frame %d (t=%.2f) of the SAME ask drew a different eye -- the face "+
					"changes while he is reading the question", i, tt)
			}
		}
	}
	if checked == 0 {
		t.Fatal("every frame was a blink, so nothing was checked")
	}
	if blinks == 0 {
		t.Error("no blink in 50 s -- the exclusion above is not being exercised, so this " +
			"test would not notice the blink being wired into the rotation by mistake")
	}
}

// And it must change BETWEEN asks, or it is not a rotation. The counter only
// moves on the rising edge, so an ask has to end before the next one counts.
func TestTheEyeChangesBetweenAsks(t *testing.T) {
	withNear(t, 2)
	c := NewCrab()
	seen := map[int]bool{}
	for i := 0; i < 40; i++ {
		c.Approach(0.1, true) // rising edge
		seen[c.eyePick()] = true
		c.Approach(0.1, false) // answered; the ask ends
	}
	if len(seen) != len(nearEyeRotation) {
		t.Errorf("over 40 asks only %d of %d faces appeared", len(seen), len(nearEyeRotation))
	}
	// The rising edge is the whole mechanism: holding the ask open must NOT
	// advance the counter.
	c2 := NewCrab()
	c2.Approach(0.1, true)
	n := c2.askN
	for i := 0; i < 50; i++ {
		c2.Approach(0.1, true)
	}
	if c2.askN != n {
		t.Errorf("holding one ask open advanced the counter from %d to %d -- the eye would "+
			"change mid-question", n, c2.askN)
	}
}

// Every eye in the rotation has to be legal art: 4 columns by 8 rows, and
// authored in PAIRS, because ToQuadrant ORs rows 2k and 2k+1 and a row written
// once is deleted before it is drawn.
func TestTheRotatingEyesAreLegalArt(t *testing.T) {
	for i, art := range nearEyeRotation {
		if len(art) != 8 {
			t.Errorf("%s: %d rows, want 8", nearEyeNames[i], len(art))
			continue
		}
		for r, row := range art {
			if len(row) != 4 {
				t.Errorf("%s row %d: %d columns, want 4", nearEyeNames[i], r, len(row))
			}
		}
		for r := 0; r+1 < 8; r += 2 {
			if art[r] != art[r+1] {
				t.Errorf("%s rows %d and %d differ (%q, %q) -- a subpixel row written once is "+
					"OR'd away before it reaches the screen", nearEyeNames[i], r, r+1, art[r], art[r+1])
			}
		}
		// It must differ from the working bead, or the ask stops being a shape.
		same := true
		for r := range art {
			if art[r] != nearEyeOpen[r] {
				same = false
			}
		}
		if same {
			t.Errorf("%s is the working bead -- the ask would read only by colour", nearEyeNames[i])
		}
	}
}
