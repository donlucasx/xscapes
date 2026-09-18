package main

import (
	"os"
	"testing"
)

// TestMain keeps the suite out of the user's real state.
//
// The adapter tests feed genuine hook payloads through translate(), and a
// SessionStart payload records its session as the current one -- which, with
// no override, means writing "s" into ~/.config/xscapes/run/current. That
// file is what a scape started by hand follows, so running the tests left the
// next `xscapes -live` chasing a session that never existed. Found on
// 2026-09-01 by reading the file while testing the launcher, not by any test:
// a suite that writes outside its temp dir cannot notice that it did.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "xscapes-test")
	if err != nil {
		panic(err)
	}
	os.Setenv("XSCAPES_HOME", dir)
	// The log switches too: a test that follows a live bus would otherwise
	// append to the real logs of whoever runs the suite inside a scape.
	os.Unsetenv("XSCAPES_HOOKLOG")
	os.Unsetenv("XSCAPES_EVENTLOG")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
