package companion

import (
	"strings"
	"testing"
)

// The fill must be INVISIBLE at the shipped size, for every cat pose. That is
// the property that makes it safe to apply blindly rather than per-pose.
func TestNearFillHiddenIsInvisibleAt1x(t *testing.T) {
	for _, c := range []struct {
		name string
		rows []string
	}{
		{"CatBody", CatBody}, {"CatAsk", CatAsk},
		{"CatDone", CatDone}, {"CatWorried", CatWorried},
	} {
		a := ParseBitmap(c.rows).ToQuadrant()
		b := ParseBitmap(nearFillHidden(c.rows)).ToQuadrant()
		if strings.Join(a, "\n") != strings.Join(b, "\n") {
			t.Errorf("%s: the fill changed the 12x7 render", c.name)
		}
		// A positive control: it must actually fill something, or the check
		// above passes for the wrong reason.
		same := true
		for i := range c.rows {
			if nearFillHidden(c.rows)[i] != c.rows[i] {
				same = false
			}
		}
		if same {
			t.Errorf("%s: the fill changed nothing, so the invisibility check proves nothing "+
				"-- every cat body carries the muzzle at pair 10-11", c.name)
		}
	}
}

// And at 2x it must remove the mouth while KEEPING the tapers.
func TestNearFillHiddenKeepsTheTapers(t *testing.T) {
	raw := ParseBitmap(nearDouble(CatBody)).ToQuadrant()
	fixed := ParseBitmap(nearDouble(nearFillHidden(CatBody))).ToQuadrant()
	if !strings.Contains(raw[5], "▀▀▀") {
		t.Fatalf("the control is wrong: the unfixed doubled row 5 should carry the mouth, got %q", raw[5])
	}
	if strings.Contains(fixed[5], "▀▀▀") {
		t.Errorf("row 5 still carries the mouth: %q", fixed[5])
	}
	// EXACTLY ONE cell row may change, and it is the jaw. Anything else means
	// the fill is eating detail the bigger sprite exists to show -- the notch
	// between the ears went first to a more general version of this rule, and
	// half-row doubling flattens nine rows.
	for y := range raw {
		if y == 5 {
			continue
		}
		if y < len(fixed) && raw[y] != fixed[y] {
			t.Errorf("row %d changed and should not have: %q -> %q", y, raw[y], fixed[y])
		}
	}
	if raw[1] != fixed[1] || !strings.Contains(fixed[1], "▄▄▄▄▄") {
		t.Errorf("the notch between the ears was flattened: %q -> %q", raw[1], fixed[1])
	}
}

// The helper must reproduce the hand-authored body exactly, or there are two
// sources of truth for the same picture and they will drift.
func TestTheHelperReproducesTheHandFilledBody(t *testing.T) {
	got := nearFillHidden(CatBody)
	for i := range CatBodyNear2 {
		if got[i] != CatBodyNear2[i] {
			t.Fatalf("row %d: helper %q, hand-authored %q", i, got[i], CatBodyNear2[i])
		}
	}
}
