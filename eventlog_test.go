package main

import (
	"path/filepath"
	"testing"
)

// The log switches pick a path for =1 the way the trace does; a path is
// taken as given; unset is off.
func TestLogSwitchesPickAPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XSCAPES_HOME", home)
	if got, want := logPath("1", "hooks.jsonl"), filepath.Join(home, "hooks.jsonl"); got != want {
		t.Errorf("=1: %q, want %q", got, want)
	}
	if got, want := logPath("yes", "events.jsonl"), filepath.Join(home, "events.jsonl"); got != want {
		t.Errorf("=yes: %q, want %q", got, want)
	}
	if got := logPath("/tmp/mine.jsonl", "hooks.jsonl"); got != "/tmp/mine.jsonl" {
		t.Errorf("a path: %q", got)
	}
	if got := logPath("", "hooks.jsonl"); got != "" {
		t.Errorf("unset: %q", got)
	}
}
