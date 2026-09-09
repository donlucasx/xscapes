package host

import (
	"strings"
	"testing"
)

// A multi-byte rune split across two writes must survive whole.
//
// ⭐ This is the strikethrough he reported through seven sessions, and it was
// ours. The agent's output comes off a pty read, which splits wherever it
// likes, and Claude Code's interface is full of multi-byte glyphs -- its
// full-width rule is U+2500, three bytes at a time. Split one and []rune turns
// the orphaned bytes into U+FFFD: a single cell becomes THREE, every cell after
// it shifts two columns, and the overflow wraps onto the row below.
//
// The terminal never sees any of it. It gets the exact bytes and draws them
// correctly, which is why a bare agent is clean and why the live screen is
// clean. Only the MODEL was wrong -- and Host.mirror writes the model's rows
// into the terminal's scrollback, which is the one place he ever saw it.
func TestASplitRuneSurvivesTheWriteBoundary(t *testing.T) {
	const whole = "──── done ────"
	want := func() string {
		s := newScreen(40, 4)
		s.feed("\x1b[?47h\x1b[1;1H" + whole)
		return s.rowAt(0)
	}()
	if want != whole {
		t.Fatalf("unsplit baseline is already wrong: %q", want)
	}
	b := []byte(whole)
	for cut := 1; cut < len(b); cut++ {
		s := newScreen(40, 4)
		s.feed("\x1b[?47h\x1b[1;1H")
		s.feed(string(b[:cut]))
		s.feed(string(b[cut:]))
		if got := s.rowAt(0); got != want {
			t.Errorf("split after %d of %d bytes -> %q, want %q", cut, len(b), got, want)
		}
	}
}

// Three-way splits, and a rune split with an escape sequence either side of it,
// because pending has to carry a rune and an escape without losing either.
func TestSplitRunesAndEscapesTogether(t *testing.T) {
	in := "\x1b[1;1H─\x1b[38;5;210m─\x1b[0m─X"
	full := func() string {
		s := newScreen(20, 3)
		s.feed("\x1b[?47h" + in)
		return s.rowAt(0)
	}()
	if !strings.HasPrefix(full, "───X") {
		t.Fatalf("baseline %q, want to start ───X", full)
	}
	b := []byte(in)
	for i := 1; i < len(b)-1; i++ {
		for j := i + 1; j < len(b); j++ {
			s := newScreen(20, 3)
			s.feed("\x1b[?47h")
			s.feed(string(b[:i]))
			s.feed(string(b[i:j]))
			s.feed(string(b[j:]))
			if got := s.rowAt(0); got != full {
				t.Fatalf("split at %d,%d -> %q, want %q", i, j, got, full)
			}
		}
	}
}

// A lone trailing ESC still waits for its sequence: the rune fix must not have
// broken what pending already did.
func TestATrailingEscapeStillWaits(t *testing.T) {
	s := newScreen(20, 3)
	s.feed("\x1b[?47h\x1b[1;1HAB")
	s.feed("\x1b")
	s.feed("[1;1HCD")
	if got := s.rowAt(0); got != "CD" {
		t.Errorf("row %q, want CD -- a split escape was not held", got)
	}
}
