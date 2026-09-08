package host

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A trace that was ASKED FOR and is not happening has to say so.
//
// On 2026-09-07 he started a session to trace the scrollback defect, the scape
// came up, the defect reproduced on screen, and nothing was captured: the open
// had failed and openTrace returned in silence. The cost was a whole
// occurrence of a defect that has only ever been seen four times.
func TestAFailedTraceIsNotSilent(t *testing.T) {
	t.Setenv("XSCAPES_TRACE", "/nope/cannot/write.bin")
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w
	h := &Host{}
	h.openTrace()
	w.Close()
	os.Stderr = old
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	if h.trace != nil {
		t.Fatal("a trace was opened on an unwritable path")
	}
	if !strings.Contains(b.String(), "TRACING IS OFF") {
		t.Fatalf("nothing said about the failure; stderr was %q", b.String())
	}
}

// A directory that does not exist yet must not quietly turn tracing off.
func TestATraceMakesItsOwnDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "trace.bin")
	t.Setenv("XSCAPES_TRACE", path)
	h := &Host{}
	h.openTrace()
	if h.trace == nil {
		t.Fatal("no trace opened")
	}
	h.trace.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the trace file is not there: %v", err)
	}
}

// XSCAPES_TRACE=1 picks the path, so there is none to get wrong.
func TestTheShorthandPicksItsOwnPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XSCAPES_TRACE", "1")
	h := &Host{}
	h.openTrace()
	if h.trace == nil {
		t.Fatal("XSCAPES_TRACE=1 opened nothing")
	}
	h.trace.Close()
	dir := filepath.Join(home, ".config", "xscapes", "traces")
	ents, err := os.ReadDir(dir)
	if err != nil || len(ents) == 0 {
		t.Fatalf("nothing written under %s (err %v)", dir, err)
	}
	if !strings.HasSuffix(ents[0].Name(), ".bin") {
		t.Errorf("wrote %q, want a .bin", ents[0].Name())
	}
}
