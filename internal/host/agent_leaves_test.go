package host

import (
	"io"
	"os"
	"strings"
	"testing"
)

// Hermes ends its session with ESC[?1049l: it switches the terminal back to
// the main screen itself, then prints its "Resume this session with:" lines
// there and exits. The host kept painting the scape for 8 and 13 more frames
// onto that main screen (both of his traces of 2026-11-08, Kimi N4), and its
// band's scroll region stayed pinned, so the goodbye scrolled inside it.
func TestAnAgentLeavingTheAlternateScreenIsLetGo(t *testing.T) {
	left := false
	f := &Filter{OwnsAlt: true, OnLeave: func() { left = true }}
	f.Band.Store(20)
	out := string(f.Filter([]byte("bye\x1b[?1006l\x1b[?2004l\x1b[?1049l\x1b[<uResume this session")))
	if !left {
		t.Fatal("OnLeave not called for ESC[?1049l")
	}
	i := strings.Index(out, "\x1b[?1049l")
	if i < 0 {
		t.Fatalf("the agent's ESC[?1049l was not forwarded: %q", out)
	}
	if !strings.HasPrefix(out[i+len("\x1b[?1049l"):], "\x1b[?6l\x1b7\x1b[r\x1b8") {
		t.Errorf("no region release right after the switch: %q", out[i:])
	}
	if !strings.HasSuffix(out, "Resume this session") {
		t.Errorf("the goodbye did not pass: %q", out)
	}

	// Not a leave: other private modes, a split read, or a host on the main
	// screen (an agent there may run its own alternate screen).
	for _, tc := range []struct {
		name string
		own  bool
		in   []string
		want bool
	}{
		{"bracketed paste off", true, []string{"\x1b[?2004l"}, false},
		{"1049 in a list", true, []string{"\x1b[?1004;1049l"}, true},
		{"split across reads", true, []string{"\x1b[?10", "49l"}, true},
		{"1047", true, []string{"\x1b[?1047l"}, true},
		{"entering is not leaving", true, []string{"\x1b[?1049h"}, false},
		{"host on the main screen", false, []string{"\x1b[?1049l"}, false},
		{"not private", true, []string{"\x1b[1049l"}, false},
	} {
		got := false
		f := &Filter{OwnsAlt: tc.own, OnLeave: func() { got = true }}
		for _, part := range tc.in {
			f.Filter([]byte(part))
		}
		if got != tc.want {
			t.Errorf("%s: left=%v, want %v", tc.name, got, tc.want)
		}
	}
}

// After the agent left, nothing of the host's own reaches the terminal --
// no scape frame, no mirror batch, no Close -- while the agent's bytes do.
func TestTheHostWritesNothingOfItsOwnAfterTheAgentLeft(t *testing.T) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	h := &Host{out: pw}
	h.write("before")
	h.agentLeft.Store(true)
	h.write(BeginPaint() + "SCAPE" + EndPaint(8))
	h.writeAgent("goodbye")
	pw.Close()
	got, _ := io.ReadAll(pr)
	if string(got) != "beforegoodbye" {
		t.Errorf("wrote %q, want only the host's write from before and the agent's goodbye", got)
	}
}
