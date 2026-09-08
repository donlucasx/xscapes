package host

import (
	"strings"
	"testing"
)

// The right-edge strip: after a WIDTH change Terminal.app holds every row's old
// cells past the new width, erase cannot reach them, and they show through in
// its side inset. DL is the one thing that does reach them -- it allocates a
// genuinely new row at the visible width.
//
// Replayed through the screen model with retainWidth on, which is the measured
// Terminal.app rule.
func TestDeletingTheBandDropsTheRetainedTail(t *testing.T) {
	s := newScreen(120, 24)
	s.retainWidth = true
	// Paint every row full width, then narrow.
	for r := 1; r <= 24; r++ {
		s.feed("\x1b[" + itoa(r) + ";1H" + strings.Repeat("X", 120))
	}
	s.resizeAlt(74, 24)
	wide := 0
	for _, row := range s.cells {
		if len(row) > 74 {
			wide++
		}
	}
	if wide == 0 {
		t.Fatal("no row kept a tail after narrowing -- the model is not reproducing the retain rule")
	}

	// The band is rows 15..24 here; reallocate it.
	s.feed(reallocBand(15, 24))
	for r := 15; r <= 24; r++ {
		if got := len(s.cells[r-1]); got > 74 {
			t.Errorf("band row %d still holds %d cells, want at most 74 -- the tail survived DL", r, got)
		}
	}
	// And the agent's rows above the band are untouched, which is the point of
	// deleting the band rather than the screen: Claude Code cannot redraw a
	// transcript that has scrolled.
	kept := 0
	for r := 1; r <= 14; r++ {
		if len(s.cells[r-1]) > 74 {
			kept++
		}
	}
	if kept == 0 {
		t.Error("the agent's rows lost their tails too -- reallocBand reached outside the band")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
