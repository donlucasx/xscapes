package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The stats line carries the agent's erases beside the bus's counts.
func TestTheStatsLineCarriesTheErases(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	appendEventLogStats(p, 1, 0, 3, time.Unix(1_789_700_000, 0))
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("%v in %q", err, b)
	}
	if m["erased"] != 3.0 || m["dropped"] != 1.0 {
		t.Fatalf("stats line %s", b)
	}
}

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
