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
