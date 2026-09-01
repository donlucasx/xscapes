package host

import "testing"

// The host owns the scroll region. Claude Code emits ESC[r exactly once, at
// startup, wrapped in ESC7/ESC8 -- measured, see notes/claude-terminal-emissions.md.
// That one sequence resets the region to the whole screen, which would let the
// agent scroll over the scape. It has to be taken out of the stream on its way
// to the terminal.

func TestPlainBytesPassThroughUnchanged(t *testing.T) {
	var f Filter
	got := string(f.Filter([]byte("hello world")))
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestScrollRegionResetIsRemoved(t *testing.T) {
	var f Filter
	got := string(f.Filter([]byte("a\x1b[rb")))
	if got != "ab" {
		t.Errorf("got %q, want %q", got, "ab")
	}
}

func TestParameterisedScrollRegionIsRemoved(t *testing.T) {
	var f Filter
	got := string(f.Filter([]byte("a\x1b[1;10rb")))
	if got != "ab" {
		t.Errorf("got %q, want %q", got, "ab")
	}
}

func TestOtherEscapeSequencesSurvive(t *testing.T) {
	in := "\x1b[2K\x1b[1B\x1b[38;5;174m\x1b[?25l\x1b7\x1b8"
	var f Filter
	got := string(f.Filter([]byte(in)))
	if got != in {
		t.Errorf("got %q, want %q", got, in)
	}
}

// A read from the PTY can end anywhere, including the middle of an escape
// sequence. A filter that only looked at whole buffers would pass ESC[ through
// and then swallow a lone "r" out of the middle of a word.
func TestScrollRegionResetSplitAcrossReads(t *testing.T) {
	for cut := 1; cut < 3; cut++ {
		seq := "x\x1b[ry"
		var f Filter
		out := string(f.Filter([]byte(seq[:1+cut])))
		out += string(f.Filter([]byte(seq[1+cut:])))
		if out != "xy" {
			t.Errorf("cut after %d: got %q, want %q", cut, out, "xy")
		}
	}
}

func TestSplitSequenceThatIsNotDECSTBMSurvives(t *testing.T) {
	seq := "\x1b[2K"
	for cut := 1; cut < len(seq); cut++ {
		var f Filter
		out := string(f.Filter([]byte(seq[:cut])))
		out += string(f.Filter([]byte(seq[cut:])))
		if out != seq {
			t.Errorf("cut at %d: got %q, want %q", cut, out, seq)
		}
	}
}

// The bytes Claude actually sent, lifted from the capture. Everything but the
// three bytes of ESC[r has to reach the terminal, or its cursor save/restore
// pairs up wrong.
func TestClaudeStartupChunkLosesOnlyTheScrollRegionReset(t *testing.T) {
	in := "\x1b[39m\r\n\x1b7\x1b[r\x1b8\x1b[?25h\x1b[?25l\x1b[?2004h\x1b[?1004h"
	want := "\x1b[39m\r\n\x1b7\x1b8\x1b[?25h\x1b[?25l\x1b[?2004h\x1b[?1004h"
	var f Filter
	if got := string(f.Filter([]byte(in))); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Held bytes must not be lost when the stream ends mid-sequence.
func TestPendingBytesAreFlushedAtEOF(t *testing.T) {
	var f Filter
	out := string(f.Filter([]byte("done\x1b[")))
	out += string(f.Flush())
	if out != "done\x1b[" {
		t.Errorf("got %q, want %q", out, "done\x1b[")
	}
}

// A sequence that never terminates must not hold the agent's output hostage.
// Real CSI parameter runs are a handful of bytes; anything longer is garbage
// or a desynchronised stream, and the agent's text matters more than the
// filter's tidiness.
func TestUnterminatedSequenceIsReleasedRatherThanHeldForever(t *testing.T) {
	long := "\x1b["
	for i := 0; i < 300; i++ {
		long += "1"
	}
	var f Filter
	got := string(f.Filter([]byte(long)))
	if len(got) == 0 {
		t.Fatalf("filter held all %d bytes back", len(long))
	}
	if len(f.Flush())+len(got) != len(long) {
		t.Errorf("bytes lost: got %d out + %d pending, want %d", len(got), len(f.Flush()), len(long))
	}
}
