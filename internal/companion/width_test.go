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
