package host

import (
	"fmt"
	"strings"
	"testing"
)

// An agent that clears its display clears the whole terminal: ED is not
// bounded by the scroll region. Kimi Code CLI sends ESC[2J ESC[H ESC[3J on
// every step of a window drag (his trace of 2026-09-18: 21 size steps, 21
// clears, none during work); Claude Code never sends one. That was the
// difference between the scape blinking in and out under one agent and
// holding still under the other: each clear wiped the scape's rows, and the
// damage tracker, which repaints only rows whose content changed, left the
// still ones blank until its periodic full refresh. The filter confines an
// agent's erase to the band, and drops its clear of the scrollback, which is
// where the host mirrors the agent's own rows.
func TestAnAgentsEraseStaysInItsBand(t *testing.T) {
	const W, H, band = 40, 12, 8
	pin := func() (*screen, *Filter) {
		s := newScreen(W, H)
		s.feed(EnterBand(band))
		for y := 1; y <= band; y++ {
			s.feed(fmt.Sprintf("\x1b[%d;1Hagent row %d", y, y))
		}
		// The host paints the rows under the band the way the frame loop
		// does: region off, absolute rows, then the band pinned again.
		var b strings.Builder
		b.WriteString(BeginPaint())
		for y := band + 1; y <= H; y++ {
			fmt.Fprintf(&b, "\x1b[%d;1Hscape row %d", y, y)
		}
		b.WriteString(EndPaint(band))
		s.feed(b.String())
		f := &Filter{}
		f.Band.Store(band)
		return s, f
	}
	scapeKept := func(s *screen, what string) {
		t.Helper()
		for y := band; y < H; y++ {
			if got, want := s.rowAt(y), fmt.Sprintf("scape row %d", y+1); got != want {
				t.Errorf("%s: scape row %d reads %q, want %q", what, y+1, got, want)
			}
		}
	}

	// ESC[2J: every band row erased, every scape row kept.
	s, f := pin()
	s.feed("\x1b[4;3H")
	s.feed(string(f.Filter([]byte("\x1b[2J\x1b[H"))))
	for y := 0; y < band; y++ {
		if got := s.rowAt(y); got != "" {
			t.Errorf("ESC[2J: band row %d still reads %q", y+1, got)
		}
	}
	scapeKept(s, "ESC[2J")

	// ESC[J from row 4, column 3: the rest of row 4 and the band rows under
	// it erased; rows 1..3 and the scape kept.
	s, f = pin()
	s.feed("\x1b[4;3H")
	s.feed(string(f.Filter([]byte("\x1b[J"))))
	for y := 0; y < band; y++ {
		want := fmt.Sprintf("agent row %d", y+1)
		if y == 3 {
			want = "ag"
		}
		if y > 3 {
			want = ""
		}
		if got := s.rowAt(y); got != want {
			t.Errorf("ESC[J: band row %d reads %q, want %q", y+1, got, want)
		}
	}
	scapeKept(s, "ESC[J")

	// ESC[3J is dropped; ESC[1J is the band's own business and passes.
	f = &Filter{}
	f.Band.Store(band)
	if out := f.Filter([]byte("\x1b[3J")); len(out) != 0 {
		t.Errorf("ESC[3J forwarded as %q", out)
	}
	if out := f.Filter([]byte("\x1b[1J")); string(out) != "\x1b[1J" {
		t.Errorf("ESC[1J changed to %q", out)
	}

	// A clear cut across two reads is still confined.
	s, f = pin()
	s.feed("\x1b[4;3H")
	out := append(f.Filter([]byte("\x1b[")), f.Filter([]byte("2J"))...)
	s.feed(string(out))
	scapeKept(s, "ESC[2J split across reads")
}
