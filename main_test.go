package main

import (
	"os"
	"testing"
)

// TestMain keeps the suite out of the user's real state.
//
// The adapter tests feed genuine hook payloads through translate(), and a
// SessionStart payload records its session as the current one -- which, with
// no override, means writing "s" into ~/.config/asciiscapes/run/current. That
// file is what a scape started by hand follows, so running the tests left the
// next `asciiscapes -live` chasing a session that never existed. Found on
// 2026-09-01 by reading the file while testing the launcher, not by any test:
// a suite that writes outside its temp dir cannot notice that it did.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "asciiscapes-test")
	if err != nil {
		panic(err)
	}
	os.Setenv("ASCIISCAPES_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
