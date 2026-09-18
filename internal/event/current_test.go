package event

import "testing"

// The first session on a fresh machine never bound (2026-09-17): SetCurrent
// wrote run/current before anything had created run/, the error was
// discarded, and the scape kept polling a pointer that was never there. The
// spool created the directory a moment later, so the second session worked,
// and no machine that had run xscapes once could show it.
func TestTheFirstSessionOnAFreshMachineIsAnnounced(t *testing.T) {
	t.Setenv("XSCAPES_HOME", t.TempDir())
	if got := Current(); got != "" {
		t.Fatalf("a fresh home names a session: %q", got)
	}
	if err := SetCurrent("session_first"); err != nil {
		t.Fatalf("SetCurrent on a fresh home: %v", err)
	}
	if got := Current(); got != "session_first" {
		t.Fatalf("Current after the first SetCurrent = %q, want session_first", got)
	}
}
