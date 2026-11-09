package host

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// The mechanism behind his Ghostty narrow-drag drift (2026-11-08 19:40,
// trace 20261108-194024, replayed through libghostty-vt): a scape frame
// composed for the old width lands after the terminal narrowed; with autowrap
// on, its last row wraps and scrolls the whole screen, agent's text included.
// The host's writes go out with autowrap off, so the row is clipped instead.
func TestAFrameForAWiderWindowDoesNotScrollTheScreen(t *testing.T) {
	const old, narrow, rows, band = 60, 40, 12, 7
	frame := func() string {
		var b strings.Builder
		b.WriteString(BeginPaint())
		for r := band + 1; r <= rows; r++ {
			fmt.Fprintf(&b, "\x1b[%d;1H%s", r, strings.Repeat("~", old))
		}
		b.WriteString(EndPaint(band))
		return b.String()
	}
	for _, guarded := range []bool{false, true} {
		sc := newScreen(old, rows)
		sc.feed(Open(true, band, rows))
		for r := 1; r <= band; r++ {
			sc.feed(fmt.Sprintf("\x1b[%d;1HLINE %02d", r, r))
		}
		sc.resizeGhostty(narrow, rows) // the window narrowed; the frame was already composed
		pr, pw, _ := os.Pipe()
		h := &Host{out: pw}
		if guarded {
			h.write(frame())
		} else {
			h.writeAgent(frame())
		}
		pw.Close()
		got, _ := io.ReadAll(pr)
		sc.feed(string(got))
		top := strings.TrimSpace(sc.rowAt(0))
		if guarded && top != "LINE 01" {
			t.Errorf("guarded write: band row 1 = %q, want LINE 01 (the screen scrolled)", top)
		}
		if !guarded && top == "LINE 01" {
			t.Errorf("control: the unguarded frame did not scroll the screen, so this test proves nothing")
		}
	}
}

// The host puts back the agent's own autowrap state, which Hermes turns off
// around every redraw: forcing it on in the middle of one would wrap the
// agent's lines.
func TestTheHostRestoresTheAgentsAutowrap(t *testing.T) {
	var states []bool
	f := &Filter{OnAutowrap: func(off bool) { states = append(states, off) }}
	f.Filter([]byte("\x1b[?7lchrome\x1b[?25l\x1b[?2004;7h\x1b[?7"))
	f.Filter([]byte("l"))
	if fmt.Sprint(states) != "[true false true]" {
		t.Errorf("autowrap seen as %v, want [true false true] (off, on in a list, off across a split read)", states)
	}

	for _, off := range []bool{false, true} {
		pr, pw, _ := os.Pipe()
		h := &Host{out: pw}
		h.agentNoWrap.Store(off)
		h.write("frame")
		pw.Close()
		got, _ := io.ReadAll(pr)
		want := "\x1b[?7lframe\x1b[?7h"
		if off {
			want = "\x1b[?7lframe"
		}
		if string(got) != want {
			t.Errorf("agent autowrap off=%v: host wrote %q, want %q", off, got, want)
		}
	}
}
