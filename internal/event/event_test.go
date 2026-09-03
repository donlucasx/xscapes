package event

import (
	"os"
	"strings"
	"testing"
	"time"
)

// tmpHome gives the test a short home.
//
// Not t.TempDir(): it embeds the test's NAME in the path, and sockaddr_un
// allows 103 bytes. Two tests here failed on exactly that the first time they
// ran, which is a fair warning about how little headroom the limit leaves.
func tmpHome(t *testing.T) {
	t.Helper()
	d, err := os.MkdirTemp("", "as")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	t.Setenv("XSCAPES_HOME", d)
}

func TestRoundTripOverSocket(t *testing.T) {
	tmpHome(t)
	bus, err := Listen("sess-round-trip")
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()

	sent := Event{
		Kind: ToolEnd, Session: "sess-round-trip", Op: OpEdit, Tool: "Edit",
		Target: "internal/scape/shore.go", Detail: "+18 -2", ID: "tu1", MS: 412,
	}
	viaSock, err := Emit(sent)
	if err != nil {
		t.Fatal(err)
	}
	if !viaSock {
		t.Fatal("a live engine was listening, but the emitter used the file fallback")
	}

	select {
	case got := <-bus.C:
		if got.Kind != sent.Kind || got.Target != sent.Target || got.MS != sent.MS {
			t.Errorf("round trip changed the event: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("nothing arrived on the bus")
	}
}

// The fallback is the whole reason a hook can run when no scape does. It must
// take the file path, silently, and still report success to the caller.
func TestFallsBackToSpoolWithNoEngine(t *testing.T) {
	tmpHome(t)
	viaSock, err := Emit(Event{Kind: Done, Session: "sess-no-engine", Text: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if viaSock {
		t.Fatal("claimed to reach a socket with no engine running")
	}
	p, _ := SpoolPath("sess-no-engine")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"kind":"done"`) {
		t.Errorf("spool does not hold the event: %s", b)
	}
}

// A second scape on the same session must be refused rather than silently
// stealing the socket from the first.
func TestSecondListenerIsRefused(t *testing.T) {
	tmpHome(t)
	a, err := Listen("sess-busy")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err := Listen("sess-busy"); err != ErrBusy {
		t.Errorf("second listener got %v, want ErrBusy", err)
	}
}

// A socket left by a killed engine must not block the next one forever.
func TestStaleSocketIsReclaimed(t *testing.T) {
	tmpHome(t)
	a, err := Listen("sess-stale")
	if err != nil {
		t.Fatal(err)
	}
	// Close the connection but leave the inode, which is exactly what a
	// SIGKILLed engine leaves behind.
	a.conn.Close()
	close(a.stop)

	p, _ := SockPath("sess-stale")
	if _, err := os.Stat(p); err != nil {
		t.Fatal("expected a stale socket file to still exist")
	}
	b, err := Listen("sess-stale")
	if err != nil {
		t.Fatalf("could not reclaim a stale socket: %v", err)
	}
	b.Close()
}

// Everything in an event is derived from something the agent touched, and all
// of it gets painted into a terminal that also emits its own escape sequences.
// A filename carrying a CSI sequence must not be able to move the cursor.
func TestControlCharactersAreStripped(t *testing.T) {
	evil := "src/\x1b[2Jgone\x07.go"
	line := Encode(Event{Kind: ToolEnd, Target: evil})
	if strings.ContainsAny(string(line), "\x1b\x07") {
		t.Errorf("encoded line still carries control characters: %q", line)
	}
	got, err := Decode(line)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(got.Target, "\x1b\x07") {
		t.Errorf("decoded target still carries control characters: %q", got.Target)
	}
	if !strings.Contains(got.Target, "gone.go") {
		t.Errorf("stripping ate the actual filename: %q", got.Target)
	}
}

// An oversized event must be shortened, never dropped and never sent at a size
// the datagram cannot carry.
func TestOversizeEventIsTruncatedNotDropped(t *testing.T) {
	line := Encode(Event{
		Kind: Prompt,
		Text: strings.Repeat("a very long pasted prompt ", 500),
		Tool: "Bash",
	})
	if len(line) > MaxLine {
		t.Errorf("encoded line is %d bytes, over the %d cap", len(line), MaxLine)
	}
	got, err := Decode(line)
	if err != nil {
		t.Fatalf("truncation produced invalid JSON: %v", err)
	}
	if got.Kind != Prompt {
		t.Errorf("truncation lost the kind: %+v", got)
	}
}

// An adapter from the future can send something this build has never heard of.
// It must survive and be visible, not crash and not count as activity.
func TestUnknownKindSurvives(t *testing.T) {
	got, err := Decode([]byte(`{"v":2,"kind":"aurora","session":"s","intensity":3}`))
	if err != nil {
		t.Fatalf("an unknown kind must decode, got %v", err)
	}
	if got.Known() {
		t.Error("aurora should not be a known kind")
	}
	if got.Busy() {
		t.Error("an unknown kind must not count as the agent working")
	}
}

// A hand-written event with no version reads as v1, so the protocol is usable
// from a shell without knowing anything about it.
func TestMissingVersionReadsAsOne(t *testing.T) {
	got, err := Decode([]byte(`{"kind":"done"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.V != 1 {
		t.Errorf("v = %d, want 1", got.V)
	}
}

// A session id is the one field an adapter passes through from elsewhere, so
// it must never be able to become a path.
func TestTagCannotEscapeTheRunDirectory(t *testing.T) {
	for _, in := range []string{"../../.ssh/id_rsa", "/etc/passwd", "..", "", "a/b"} {
		got := Tag(in)
		if strings.ContainsAny(got, "/.\\") {
			t.Errorf("Tag(%q) = %q, which contains a path character", in, got)
		}
		if got == "" {
			t.Errorf("Tag(%q) produced an empty name", in)
		}
	}
}
