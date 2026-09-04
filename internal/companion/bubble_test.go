package companion

import (
	"strings"
	"testing"
)

// Every row of a balloon must be the same width, or the box shears.
func TestBalloonsDoNotShear(t *testing.T) {
	for _, text := range []string{"", "ok", "allow Bash?", "Rate limiting is in. 100 req/min per IP."} {
		for name, rows := range map[string][]string{
			"ask":           Bubble(text),
			"done":          DoneBubble(text),
			"ask mirrored":  MirrorTail(Bubble(text)),
			"done mirrored": MirrorTail(DoneBubble(text)),
		} {
			w := len([]rune(rows[0]))
			for i, r := range rows {
				if got := len([]rune(r)); got != w {
					t.Errorf("%s balloon for %q: row %d is %d runes, want %d", name, text, i, got, w)
				}
			}
		}
	}
}

// The two balloons are the two moments the scene asks for the user's eyes,
// and the brief locks them as distinct cues. Shape has to carry that on its
// own: colour is gone in a monochrome capture and in -plain.
func TestAskAndDoneBalloonsDiffer(t *testing.T) {
	a, d := Bubble("tests passed"), DoneBubble("tests passed")
	for i := range a {
		if a[i] == d[i] {
			t.Errorf("row %d renders identically in both balloons: %q", i, a[i])
		}
	}
}

// Mirrored, the pointer sits under the right shoulder, where the creature is.
func TestMirroredTailPointsRight(t *testing.T) {
	rows := MirrorTail(Bubble("allow Bash?"))
	last := []rune(rows[len(rows)-1])
	if last[len(last)-4] != 'v' {
		t.Errorf("mirrored pointer not under the right shoulder: %q", string(last))
	}
}

// A paragraph-long done message is cut to a glance; a short ask is untouched.
func TestBubbleTextIsCappedToAGlance(t *testing.T) {
	long := strings.Repeat("Two agents are running in parallel ", 4)
	rows := DoneBubble(long)
	if w := len([]rune(rows[1])); w > MaxBubbleText+4 {
		t.Errorf("balloon is %d wide for a long message; cap is %d plus the walls", w, MaxBubbleText)
	}
	if !strings.Contains(rows[1], "…") {
		t.Errorf("a cut message must say so: %q", rows[1])
	}
	if got := Bubble("allow Bash?")[1]; got != "| allow Bash? |" {
		t.Errorf("a short ask changed: %q", got)
	}
}
