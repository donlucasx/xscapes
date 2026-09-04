package term

import (
	"fmt"
	"testing"
)

// The ends of a ramp are the ends the quantiser would have chosen, so a
// palette that was tuned so its keyframes land on cube entries keeps them.
func TestRampEndsAreTheQuantisedEnds(t *testing.T) {
	a, b := RGB{0, 95, 135}, RGB{255, 175, 135} // dawn: skyDawn to firePeach
	r := NewRamp(a, b)
	if got, want := r.Tone(0), FromIndex256(a.Index256Keeping()); got != want {
		t.Errorf("start: got %v want %v", got, want)
	}
	if got, want := r.Tone(0.999), FromIndex256(b.Index256Keeping()); got != want {
		t.Errorf("end: got %v want %v", got, want)
	}
	if got := r.True(0.5); got != Lerp(a, b, 0.5) {
		t.Errorf("True(0.5) = %v, want the plain blend %v", got, Lerp(a, b, 0.5))
	}
}

// Every tone on the path is a real palette entry, no channel leaves the range
// the two ends span, and a tone never repeats.
func TestRampStaysInTheBoxAndNeverRepeats(t *testing.T) {
	pairs := [][2]RGB{
		{{0, 95, 135}, {255, 175, 135}}, // dawn sky
		{{0, 95, 175}, {175, 215, 255}}, // noon sky
		{{0, 95, 175}, {255, 175, 135}}, // late afternoon sky
		{{0, 95, 95}, {135, 175, 255}},  // noon sea
		{{0, 95, 95}, {95, 135, 175}},   // dusk sea
		{{6, 8, 22}, {40, 46, 78}},      // midnight sky
		{{2, 58, 111}, {165, 98, 88}},   // the 20:30 sky, both ends off the lattice
	}
	for _, pr := range pairs {
		r := NewRamp(pr[0], pr[1])
		qa, qb := FromIndex256(pr[0].Index256Keeping()), FromIndex256(pr[1].Index256Keeping())
		t.Logf("%v -> %v: %s", pr[0], pr[1], describe(r.Path))
		seen := map[RGB]bool{}
		for i, c := range r.Path {
			if FromIndex256(c.Index256()) != c {
				t.Errorf("%v->%v: tone %d is %v, not a palette entry", pr[0], pr[1], i, c)
			}
			if seen[c] {
				t.Errorf("%v->%v: tone %v repeats", pr[0], pr[1], c)
			}
			seen[c] = true
			if c.Index256() >= 232 {
				continue // greys are bounded by brightness, not by channel
			}
			for k, v := range [3]uint8{c.R, c.G, c.B} {
				lo, hi := [3]uint8{qa.R, qa.G, qa.B}[k], [3]uint8{qb.R, qb.G, qb.B}[k]
				if lo > hi {
					lo, hi = hi, lo
				}
				if v < lo || v > hi {
					t.Errorf("%v->%v: tone %v leaves the box %v..%v in channel %d", pr[0], pr[1], c, qa, qb, k)
				}
			}
		}
	}
}

// Between two greys the path is the grey ramp, one step at a time, in order.
func TestRampBetweenGreysWalksTheGreyRamp(t *testing.T) {
	r := NewRamp(RGB{6, 8, 22}, RGB{40, 46, 78})
	if len(r.Path) < 3 {
		t.Fatalf("a night sky should carry several greys, got %v", r.Path)
	}
	prev := -1
	for _, c := range r.Path {
		i := c.Index256()
		if i < 232 {
			t.Errorf("night tone %v is not on the grey ramp", c)
		}
		if prev != -1 && i != prev+1 {
			t.Errorf("night ramp jumps from %d to %d", prev, i)
		}
		prev = i
	}
}

// The reason the path exists, on the ramp that measured worst: the sky at
// five in the morning, deep blue to peach through a neutral middle. Rounding
// row by row took six steps of 20 or more, flipping between the grey ramp and
// the cube through a teal and an olive. The path takes its one unavoidable
// green step through the greys instead and lands under the hop, and every
// step is one tone.
//
// 26 is the measured floor with this palette (the hop, see NewRamp), not a
// wish: a green step out of the dawn's slate blue is 32 by any other route.
func TestRampStepsAreSmallerThanRounding(t *testing.T) {
	// The sky at 05:00: five sixths of the way from midnight to the dawn
	// keyframe, so neither end is a palette entry and the middle passes
	// through neutral. This is the ramp the audit measured worst.
	a, b := RGB{1, 81, 116}, RGB{219, 153, 126}
	r := NewRamp(a, b)
	pathMax, pathHard := 0.0, 0
	for i := 1; i < len(r.Path); i++ {
		d := DeltaE(r.Path[i-1], r.Path[i])
		if d > pathMax {
			pathMax = d
		}
		if d >= 20 {
			pathHard++
		}
	}
	roundHard := 0
	var prev RGB
	var rounded []RGB
	for i := 0; i < 24; i++ {
		q := FromIndex256(Lerp(a, b, float64(i)/23).Index256Keeping())
		if i > 0 && q != prev && DeltaE(prev, q) >= 20 {
			roundHard++
		}
		if i == 0 || q != prev {
			rounded = append(rounded, q)
		}
		prev = q
	}
	t.Logf("dawn sky, the path:  %s", describe(r.Path))
	t.Logf("dawn sky, rounded:   %s", describe(rounded))
	t.Logf("dawn sky: largest step on the path %.1f; hard edges (20+) path %d, rounding %d", pathMax, pathHard, roundHard)
	if pathMax > 26 {
		t.Errorf("the path's largest step is %.1f, over the measured floor of 26", pathMax)
	}
	if pathHard >= roundHard {
		t.Errorf("the path has %d hard edges, rounding had %d -- the path is not doing its job", pathHard, roundHard)
	}
}

// Same ends, same path, from the cache; and equal shares of t.
func TestRampTonesShareTheRowsEqually(t *testing.T) {
	r := NewRamp(RGB{0, 95, 175}, RGB{175, 215, 255})
	k := len(r.Path)
	for i, c := range r.Path {
		lo, hi := (float64(i)+0.01)/float64(k), (float64(i)+0.99)/float64(k)
		if r.Tone(lo) != c || r.Tone(hi) != c {
			t.Errorf("tone %d should hold from %.3f to %.3f", i, lo, hi)
		}
	}
}

// describe writes a path as tones and the step into each, for the log.
func describe(path []RGB) string {
	out := ""
	for i, c := range path {
		if i > 0 {
			out += fmt.Sprintf(" -%.0f-> ", DeltaE(path[i-1], c))
		}
		out += fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return out
}
