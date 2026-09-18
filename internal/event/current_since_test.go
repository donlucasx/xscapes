package event

import (
	"testing"
	"time"
)

// A launcher must bind to the session that starts AFTER it, and a resumed
// session writes the same id the pointer already held: waiting for a
// different id never bound it (F1, session 38). CurrentSince answers by the
// pointer's write time, not its text.
func TestAResumedSessionIsCurrentSinceTheLaunch(t *testing.T) {
	t.Setenv("XSCAPES_HOME", t.TempDir())
	if err := SetCurrent("abc-123"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	launched := time.Now()
	if got := CurrentSince(launched); got != "" {
		t.Fatalf("the pointer written before the launch counts as current: %q", got)
	}
	time.Sleep(20 * time.Millisecond)
	if err := SetCurrent("abc-123"); err != nil { // the same id again: a resume
		t.Fatal(err)
	}
	if got := CurrentSince(launched); got != "abc-123" {
		t.Fatalf("the pointer written after the launch, same id, is not current: %q", got)
	}
	if got := CurrentSince(time.Now().Add(time.Hour)); got != "" {
		t.Fatalf("a pointer older than the moment asked about counts as current: %q", got)
	}
}
