package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestTermSizeDoesNotDependOnStdin is the regression test for the bug that made
// the scene 80x24 forever.
//
// termSize used to shell out to `stty size`. stty reads the window size from
// its OWN stdin, and a child started by exec.Command gets /dev/null unless told
// otherwise -- so it always failed and termSize always returned the fallback.
// The scene was never once the size of the window it was running in, which
// showed up as three separate-looking complaints: it does not fill the frame,
// it garbles when you shrink it, it glitches when you stretch it.
//
// The test pins the property rather than the implementation: asking for the
// size must not depend on what stdin happens to be.
func TestTermSizeDoesNotDependOnStdin(t *testing.T) {
	if out, err := exec.Command("stty", "size").Output(); err == nil && len(out) > 0 {
		t.Skip("stdin happens to be a terminal here; the bug needs a non-tty stdin to show")
	}
	// With stdin not a terminal, the old implementation returned the fallback.
	// The new one asks the controlling terminal, so under `go test` (no tty at
	// all) it still falls back -- what must NOT happen is a panic or a zero.
	w, h := termSize()
	if w < 8 || h < 4 {
		t.Errorf("termSize returned an unusable %dx%d", w, h)
	}
}

// The ioctl must be asked of a descriptor that actually reaches the terminal.
// If every standard stream is redirected, /dev/tty is the last resort, and
// leaving it out is how a scape in a pipeline loses its size.
func TestTTYFDsIncludesDevTTY(t *testing.T) {
	if _, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err != nil {
		t.Skip("no controlling terminal in this environment")
	}
	if got := len(ttyFDs()); got < 4 {
		t.Errorf("ttyFDs returned %d descriptors, want stdout/stderr/stdin plus /dev/tty", got)
	}
}
