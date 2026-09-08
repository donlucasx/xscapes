package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// `xscapes companion cat` reaches a scape that is ALREADY RUNNING.
//
// It did not, and both the CLI and the README promised it did. He switched to
// the cat on 2026-09-08 and the crab stayed on the beach: companionPref() was
// read exactly once, inside newFrames, which runs at startup, and even resize()
// recomputed the layout from the companion width it already held.
//
// The check is the layout as well as the name. Today both animals report the
// same Size() -- 12 by 7, because Size() returns the BOX and not the ink, and
// the cat's ink uses only nine of its twelve columns -- so the layout does not
// in fact move, and this test first asserted that it did and failed on the
// truth. It is kept as an invariant rather than a difference: whatever the
// companion measures, ccw and the layout agree with it, so a future companion
// with a different box cannot be drawn into the old one's hole.
func TestARunningScapePicksUpTheCompanion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XSCAPES_HOME", dir)
	pref := filepath.Join(dir, "companion")

	write := func(name string) {
		if err := os.WriteFile(pref, []byte(name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("crab")
	f := newFrames(100, 24, 7, false, true, 0.2, 0.5)
	if got := f.cat.Name(); got != "crab" {
		t.Fatalf("newFrames read the preference as %q, want crab", got)
	}
	f.frame(time.Now())

	// The switch, exactly as the subcommand makes it.
	write("cat")

	// Inside companionPoll nothing should change: a file read every frame is
	// twelve syscalls a second for a value that changes about never.
	f.frame(time.Now())
	if f.cat.Name() != "crab" {
		t.Fatalf("the poll interval is not being honoured: it switched before %v elapsed", companionPoll)
	}

	// After it, the scape is the cat, and the layout moved with it.
	f.frame(time.Now().Add(companionPoll + time.Millisecond))
	if got := f.cat.Name(); got != "cat" {
		t.Fatalf("a running scape did not pick up the switch: still %q", got)
	}
	w, h := f.cat.Size()
	if w != f.ccw || h != f.chh {
		t.Errorf("the companion measures %dx%d but the scene is laid out for %dx%d -- "+
			"the new animal is being drawn into the old one's hole", w, h, f.ccw, f.chh)
	}
	if f.lay.MoonX != compose(f.c.W, f.ccw, f.mirror).MoonX {
		t.Errorf("the layout was not recomputed for the new companion")
	}
}
