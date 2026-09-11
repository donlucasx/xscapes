package notify

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMain keeps the whole package out of his real ~/.config/xscapes.
//
// New() now WRITES -- it materialises the embedded cues -- and every test in
// this package calls it. This is the same trap the s24 session hit with
// XSCAPES_TRACE, where a test run inside a traced session opened sixteen real
// trace files: a test that inherits a live environment variable quietly
// operates on the user's real state. XSCAPES_HOME is the one knob that moves
// all of it, so it is set once, here, before any test runs.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "xscapes-notify-test-")
	if err != nil {
		panic(err)
	}
	os.Setenv("XSCAPES_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// The cue is bytes compiled into the binary, so nothing checks it at build
// time: a truncated copy, a file saved in the wrong format, or an empty
// placeholder would all compile and would all fail as silence at the one
// moment the user is not looking at the screen.
func TestTheEmbeddedCuesAreRealWavs(t *testing.T) {
	for _, c := range []struct {
		name string
		b    []byte
	}{{"ask", askWAV}, {"done", doneWAV}} {
		if len(c.b) == 0 {
			t.Fatalf("%s cue is empty", c.name)
		}
		// 44 bytes is a canonical WAV header with no audio under it.
		if len(c.b) <= 44 {
			t.Errorf("%s cue is %d bytes, which is a header and no sound", c.name, len(c.b))
		}
		if len(c.b) > maxCueBytes {
			t.Errorf("%s cue is %d bytes, over the %d cap: at 16-bit mono 44.1 kHz that is more than 1.5 s, so something other than a short cue got embedded",
				c.name, len(c.b), maxCueBytes)
		}
		if got := string(c.b[0:4]); got != "RIFF" {
			t.Errorf("%s cue starts %q, want RIFF", c.name, got)
		}
		if got := string(c.b[8:12]); got != "WAVE" {
			t.Errorf("%s cue is a RIFF of type %q, want WAVE", c.name, got)
		}
		// The declared RIFF length has to match the bytes actually present, or
		// the file was truncated somewhere between make.py and go:embed and a
		// player will read off the end of it.
		want := uint32(len(c.b) - 8)
		if got := binary.LittleEndian.Uint32(c.b[4:8]); got != want {
			t.Errorf("%s cue declares %d bytes of RIFF payload, has %d", c.name, got, want)
		}
	}
}

// Ask rises and done falls on purpose. It is the same rule Bubble and
// DoneBubble follow on screen -- "shape carries the done/needs_input
// distinction wherever colour cannot" -- and a user listening from another
// pane has nothing BUT the shape.
func TestAskAndDoneAreDifferentSounds(t *testing.T) {
	if bytes.Equal(askWAV, doneWAV) {
		t.Fatal("the two embedded cues are the same bytes")
	}
	t.Setenv("XSCAPES_HOME", t.TempDir())
	p := New()
	if p.Silent() || p.bell {
		t.Skipf("no file player on this machine (%s)", p.Describe())
	}
	a, d := p.args[Ask], p.args[Done]
	if len(a) == 0 || len(d) == 0 {
		t.Fatalf("a sounding player has no arguments: ask=%v done=%v", a, d)
	}
	if a[len(a)-1] == d[len(d)-1] {
		t.Errorf("ask and done both play %q", a[len(a)-1])
	}
}

// The shipped default is the droplet, not the borrowed system chime. This is
// the whole point of the change, so it gets its own assertion rather than
// being implied by the ladder's order.
func TestTheDropletIsTheDefaultWhenHomeIsWritable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XSCAPES_HOME", dir)
	p := New()
	if playerBin() == "" {
		t.Skip("no afplay/paplay on this machine; the droplet rung is unreachable by design")
	}
	if p.source != ourCue {
		t.Fatalf("player is %q, want the embedded droplet", p.Describe())
	}
	got, err := os.ReadFile(filepath.Join(dir, "sounds", "ask.wav"))
	if err != nil {
		t.Fatalf("ask cue not on disk: %v", err)
	}
	if !bytes.Equal(got, askWAV) {
		t.Error("the materialised ask cue is not the embedded bytes")
	}
}

// Writing the cue on every start would mean writing it on every scape in every
// project, several times a day, for bytes that never change. The file is
// state, not output.
func TestTheCueIsWrittenOnceAndNotRewritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XSCAPES_HOME", dir)
	if p := New(); p.source != ourCue {
		t.Skipf("droplet rung unreachable here (%s)", p.Describe())
	}
	path := filepath.Join(dir, "sounds", "ask.wav")

	// Stamped back in time rather than compared at whatever resolution the
	// filesystem happens to keep: two New() calls a millisecond apart can
	// share a modification time on a one-second-resolution filesystem, and
	// then the test passes for a player that rewrites the file every start.
	old := time.Now().Add(-72 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	New()

	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st.ModTime().Equal(old) {
		t.Errorf("the cue was rewritten on the second start: mtime %v, want the stamped %v", st.ModTime(), old)
	}
}

// The reason the freshness check reads the bytes instead of trusting the size:
// the likely next edit to a cue is a new timbre at the same duration, which
// leaves a stale file at exactly the right length. A size-only check would
// keep playing the old sound forever with nothing to notice it.
func TestARetunedCueOfTheSameLengthIsRewritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XSCAPES_HOME", dir)
	if p := New(); p.source != ourCue {
		t.Skipf("droplet rung unreachable here (%s)", p.Describe())
	}
	path := filepath.Join(dir, "sounds", "ask.wav")

	stale := make([]byte, len(askWAV))
	copy(stale, askWAV)
	// A different sound, identical file length: flip the audio, not the header.
	for i := 64; i < len(stale); i++ {
		stale[i] ^= 0xFF
	}
	if err := os.WriteFile(path, stale, 0o600); err != nil {
		t.Fatal(err)
	}

	New()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, askWAV) {
		t.Error("a stale cue of the right length survived; the check is trusting the size")
	}
}

// A home it cannot write to must cost the user the SIGNATURE of the sound, not
// the sound. Going silent here would be the worst outcome of the three: the
// nudge is the 30% of the rubric, and a missed one looks exactly like an agent
// that is still working.
func TestAMaterialisationFailureFallsBackToSystemSounds(t *testing.T) {
	// A regular file where the directory has to go: MkdirAll cannot turn a
	// file into a directory, so materialisation fails the way a read-only home
	// or a full disk would, without needing either.
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XSCAPES_HOME", blocked)

	p := New()
	if p.Silent() {
		t.Fatal("an unwritable home silenced the notification entirely")
	}
	if p.source == ourCue {
		t.Fatal("claims to play the embedded droplet with nowhere to write it")
	}
	_, _, haveSystem := systemCue()
	switch {
	case playerBin() != "" && haveSystem:
		if p.source != systemCueName() {
			t.Errorf("player is %q, want the %s rung", p.Describe(), systemCueName())
		}
	default:
		if !p.bell {
			t.Errorf("player is %q, want the terminal bell floor", p.Describe())
		}
	}
}

// Describe is the ONLY way a user can tell which rung they are on. All three
// sounding rungs run the same command, so a description that names afplay and
// stops leaves someone hearing a twenty-year-old system chime with no way to
// find out that the shipped cue never got written.
func TestDescribeNamesTheRungItIsOn(t *testing.T) {
	if playerBin() == "" {
		t.Skip("no afplay/paplay on this machine; only the bell rung is reachable")
	}

	t.Run("droplet", func(t *testing.T) {
		t.Setenv("XSCAPES_HOME", t.TempDir())
		if got := New().Describe(); !strings.Contains(got, ourCue) {
			t.Errorf("Describe = %q, want it to name %q", got, ourCue)
		}
	})

	t.Run("system", func(t *testing.T) {
		if _, _, ok := systemCue(); !ok {
			t.Skip("no system sounds on this machine")
		}
		blocked := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("XSCAPES_HOME", blocked)
		got := New().Describe()
		if !strings.Contains(got, systemCueName()) {
			t.Errorf("Describe = %q, want it to name %q", got, systemCueName())
		}
		if strings.Contains(got, ourCue) {
			t.Errorf("Describe = %q, but the droplet is not what will play", got)
		}
	})
}

// Muting is the one instruction that outranks every rung, including the one
// that writes a file. A silenced scape must not touch the disk either.
func TestSilentBeatsEveryRung(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XSCAPES_HOME", dir)
	t.Setenv(SilentEnv, "1")

	p := New()
	if !p.Silent() {
		t.Errorf("player is %q with %s set, want silent", p.Describe(), SilentEnv)
	}
	if got := p.Describe(); got != "silent" {
		t.Errorf("Describe = %q, want silent", got)
	}
	p.Play(Ask) // must not panic, must not sound
	p.Play(Done)
	if _, err := os.Stat(filepath.Join(dir, "sounds")); !os.IsNotExist(err) {
		t.Errorf("a silenced player materialised the cues anyway (stat err %v)", err)
	}
}
