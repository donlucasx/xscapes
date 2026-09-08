package host

import (
	"os"
	"testing"
)

// ⚠ The suite must not inherit XSCAPES_TRACE from the shell that runs it.
//
// Found 2026-09-08 the only way it could be: `go test ./...` was run from
// inside a session started with `XSCAPES_TRACE=1 xscapes claude`, and every
// test that built a Host opened a REAL trace under ~/.config/xscapes/traces --
// sixteen files in one run, in the user's own config directory, while three
// replay tests failed with `open 1: no such file or directory` because they
// read the same variable expecting a PATH.
//
// This is term.NoSplitCells's trap wearing different clothes: a package that
// reads the ambient environment at run time cannot be tested from an
// environment that has an opinion. Clear it once, here, and hand the shell's
// value to the replay tests through traceEnv instead.
var traceEnv string

func TestMain(m *testing.M) {
	traceEnv = os.Getenv("XSCAPES_TRACE")
	if traceEnv == "" {
		traceEnv = os.Getenv("ASCIISCAPES_TRACE") // envx.Lookup falls back to it
	}
	os.Unsetenv("XSCAPES_TRACE")
	os.Unsetenv("ASCIISCAPES_TRACE")
	os.Exit(m.Run())
}

// tracePath is the trace a replay test should read, or "" to skip. The
// shorthand forms name no file -- they tell a LIVE host to pick its own path --
// so a replay has nothing to open and must skip rather than fail.
func tracePath(t *testing.T) string {
	t.Helper()
	switch traceEnv {
	case "", "1", "true", "yes", "TRUE", "YES":
		return ""
	}
	return traceEnv
}
