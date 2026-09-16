package watch

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

func kinds(ev []event.Event) []event.Kind {
	var k []event.Kind
	for _, e := range ev {
		k = append(k, e.Kind)
	}
	return k
}

func equal(a, b []event.Kind) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var t0 = time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)

// A silent program announces the session once and then says nothing.
func TestSilenceIsOnlyASessionStart(t *testing.T) {
	s := New(Options{})
	if got := kinds(s.Step(t0)); !equal(got, []event.Kind{event.SessionStart}) {
		t.Fatalf("first step: %v, want [session_start]", got)
	}
	for i := 1; i <= 40; i++ {
		if got := s.Step(t0.Add(time.Duration(i) * 250 * time.Millisecond)); len(got) != 0 {
			t.Fatalf("step %d in silence emitted %v", i, kinds(got))
		}
	}
}

// Enter is a prompt; a keystroke without Enter is not.
func TestEnterIsAPrompt(t *testing.T) {
	s := New(Options{})
	s.Step(t0)
	s.Input([]byte("hel"))
	if got := s.Step(t0.Add(250 * time.Millisecond)); len(got) != 0 {
		t.Fatalf("typing without Enter emitted %v", kinds(got))
	}
	s.Input([]byte("lo\r"))
	if got := kinds(s.Step(t0.Add(500 * time.Millisecond))); !equal(got, []event.Kind{event.Prompt}) {
		t.Fatalf("Enter emitted %v, want [prompt]", got)
	}
}

// Three seconds of output is three tool events, one a second; four seconds
// of silence after it is one sand line and one done, and then nothing.
func TestABurstThenQuietIsWorkThenDone(t *testing.T) {
	s := New(Options{})
	s.Step(t0)
	var got []event.Kind
	now := t0
	for i := 0; i < 12; i++ { // 3 s at 4 steps a second
		now = now.Add(250 * time.Millisecond)
		s.Output([]byte("building...\n"))
		got = append(got, kinds(s.Step(now))...)
	}
	if !equal(got, []event.Kind{event.ToolStart, event.ToolStart, event.ToolStart}) {
		t.Fatalf("3 s of output emitted %v, want three tool_start", got)
	}
	got = nil
	var last []event.Event
	for i := 0; i < 20; i++ { // 5 s of silence
		now = now.Add(250 * time.Millisecond)
		ev := s.Step(now)
		got = append(got, kinds(ev)...)
		if len(ev) > 0 {
			last = ev
		}
	}
	if !equal(got, []event.Kind{event.ToolEnd, event.Done}) {
		t.Fatalf("quiet after the burst emitted %v, want [tool_end done]", got)
	}
	if last[0].Target != "building..." || last[0].MS < 2500 {
		t.Errorf("the sand line is %q over %d ms; want the last printed line over ~3000 ms", last[0].Target, last[0].MS)
	}
	// Silence stays silent.
	for i := 0; i < 40; i++ {
		now = now.Add(250 * time.Millisecond)
		if ev := s.Step(now); len(ev) != 0 {
			t.Fatalf("emitted %v long after the burst ended", kinds(ev))
		}
	}
}

// An echo -- a keystroke repeated, a menu redrawn -- is not work, so it does
// not end in a done.
func TestAnEchoIsNotWork(t *testing.T) {
	s := New(Options{})
	s.Step(t0)
	s.Output([]byte("h"))
	got := kinds(s.Step(t0.Add(250 * time.Millisecond)))
	for i := 2; i < 40; i++ {
		got = append(got, kinds(s.Step(t0.Add(time.Duration(i)*250*time.Millisecond)))...)
	}
	// The one pulse of output still counts as a heartbeat of work for the
	// sea; what must NOT happen is a done.
	for _, k := range got {
		if k == event.Done || k == event.ToolEnd {
			t.Fatalf("an echo produced %v", got)
		}
	}
}

// A program that never stops writing never finishes.
func TestContinuousOutputNeverFinishes(t *testing.T) {
	s := New(Options{})
	s.Step(t0)
	now := t0
	for i := 0; i < 240; i++ { // a minute
		now = now.Add(250 * time.Millisecond)
		s.Output([]byte("."))
		for _, e := range s.Step(now) {
			if e.Kind == event.Done {
				t.Fatalf("done rang at step %d while output was still flowing", i)
			}
		}
	}
}

// The sand line is the last thing printed, without the escape sequences and
// without the spinner frames a carriage return overwrote.
func TestLastLineStripsEscapesAndSpinners(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\x1b[32mok\x1b[0m  internal/notify\n\x1b[1mPASS\x1b[0m\n", "PASS"},
		{"⠋ thinking\r⠙ thinking\r⠹ done in 3.2s\r\n", "⠹ done in 3.2s"},
		{"\x1b]0;title\x07\x1b[2J\x1b[H$ ", "$"},
		{"", ""},
		{"a line that is far longer than the forty-eight rune cap this keeps for the sand\n", "a line that is far longer than the forty-eight r"},
	}
	for _, c := range cases {
		if got := LastLine([]byte(c.in)); got != c.want {
			t.Errorf("LastLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
