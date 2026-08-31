package companion

import "testing"

// Every shipped sprite must be narrow-safe. This is the gate that stops
// kaomoji glyphs sneaking in and shearing the grid on someone else's terminal.
func TestCandidatesAreNarrowSafe(t *testing.T) {
	for i := range Candidates {
		s := &Candidates[i]
		if bad := s.UnsafeRunes(); len(bad) > 0 {
			t.Errorf("%s (%s): non-narrow runes %q", s.Name, s.Register, string(bad))
		}
	}
}

// The guard has to be able to fail, or it proves nothing.
func TestGuardRejectsWideAndAmbiguous(t *testing.T) {
	for _, r := range []rune{'人', 'つ', '・', 'ω', '⌒', '●', '∧', '·', '█', '─'} {
		if isNarrow(r) {
			t.Errorf("isNarrow(%q) = true, want false", r)
		}
	}
	for _, r := range []rune{'(', ')', '\\', '~', '^', '_', '‿', '⡀'} {
		if !isNarrow(r) {
			t.Errorf("isNarrow(%q) = false, want true", r)
		}
	}
}

func TestBubbleRowsAlign(t *testing.T) {
	rows := Bubble("tests passed")
	w := len([]rune(rows[0]))
	for i, r := range rows[:3] {
		if got := len([]rune(r)); got != w {
			t.Errorf("bubble row %d width %d, want %d (%q)", i, got, w, r)
		}
	}
}

func TestBitmapRowsUniform(t *testing.T) {
	for name, rows := range Bitmaps {
		w := len([]rune(rows[0]))
		for i, r := range rows {
			if got := len([]rune(r)); got != w {
				t.Errorf("%s row %d is %d wide, want %d: %q", name, i, got, w, r)
			}
		}
	}
}

// The generated renderings must also be safe to put in a terminal.
func TestGeneratedRenderingsAreSafe(t *testing.T) {
	b := ParseBitmap(PixelCat)
	for _, tc := range []struct {
		name string
		rows []string
	}{
		{"braille", b.ToBraille()},
		{"quadrant", b.ToQuadrant()},
		{"chars", b.ToChars()},
	} {
		for _, row := range tc.rows {
			for _, r := range row {
				if !isNarrow(r) && !isBlockSafe(r) {
					t.Errorf("%s: rune %q (U+%04X) is neither narrow nor a documented block", tc.name, r, r)
				}
			}
		}
	}
}

// Which renderings lean on ambiguous-width blocks. Not a failure — a report,
// so the exposure is never invisible.
func TestReportsAmbiguousReliance(t *testing.T) {
	b := ParseBitmap(PixelCat)
	for _, tc := range []struct {
		name string
		rows []string
	}{{"braille", b.ToBraille()}, {"quadrant", b.ToQuadrant()}, {"chars", b.ToChars()}} {
		seen := map[rune]bool{}
		for _, row := range tc.rows {
			for _, r := range row {
				if isBlockSafe(r) {
					seen[r] = true
				}
			}
		}
		var list []rune
		for r := range seen {
			list = append(list, r)
		}
		t.Logf("%-9s ambiguous-width runes used: %q", tc.name, string(list))
	}
}

// Eyes are plotted at a fixed cell row while the body shifts by two source
// pixels to breathe. If a socket straddles a cell boundary under that shift,
// the eyes end up outside the head.
func TestEyeSocketsSurviveBreathing(t *testing.T) {
	for name, er := range eyeRows {
		for _, lift := range []int{0, 2} {
			top, bot := (er[0]+lift)/4, (er[1]+lift)/4
			if top != bot {
				t.Errorf("%s: eye socket rows %d-%d straddle cell rows %d and %d at lift %d",
					name, er[0]+lift, er[1]+lift, top, bot, lift)
			}
		}
	}
	// The check has to be able to fail, or it proves nothing. Rows 5-6 are what
	// the kittens actually had: fine at rest, straddling once they breathe.
	if (5+2)/4 == (6+2)/4 {
		t.Error("guard cannot detect the straddle it was written for")
	}
}

// The declared eye cell columns must actually land on socket gaps. Deriving
// them from sprite width was right for two scales and wrong for the third.
func TestEyeColumnsLandOnSockets(t *testing.T) {
	for _, tr := range kitTiers {
		b := ParseBitmap(tr.rows)
		er := eyeRows[tr.name]
		for _, cell := range tr.eyes {
			hit := false
			for sx := cell * 2; sx < cell*2+2 && sx < b.W; sx++ {
				for sy := er[0]; sy <= er[1]; sy++ {
					if !b.at(sx, sy) {
						hit = true
					}
				}
			}
			if !hit {
				t.Errorf("%s: eye cell column %d has no socket gap at rows %d-%d",
					tr.name, cell, er[0], er[1])
			}
		}
	}
}

// One size for the whole litter, with a gap between growing and shrinking so a
// count hovering on a boundary does not flip every kitten each frame.
func TestTierHysteresis(t *testing.T) {
	tier := 0
	for _, step := range []struct{ n, want int }{
		{5, 0}, {6, 1}, {5, 1}, {4, 0}, // 5 does not grow them back; 4 does
		{10, 2}, {9, 2}, {8, 1}, // same gap at the second boundary
	} {
		tier = TierFor(step.n, tier)
		if tier != step.want {
			t.Errorf("n=%d: tier %d, want %d", step.n, tier, step.want)
		}
	}
}
