package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// The published page must not change with the terminal it was built from.
// The HTML writers save and restore term.NoSplitCells around their own
// output, but the shore decides its disc's caps when it PAINTS (FlatCaps in
// shore.go), before any writer runs, so a page built from a Claude Code shell
// inside Terminal.app came out with a flat-capped sun in every shore clip
// (2026-09-17, caught by comparing a rebuild against the committed page: 139
// frames differed in colour, all on the shore). sitePage pins both switches
// for the whole render.
func TestThePageDoesNotChangeWithTheTerminal(t *testing.T) {
	build := func(noSplit, lower bool) string {
		wasN, wasL := term.NoSplitCells, term.LowerHalf
		term.NoSplitCells, term.LowerHalf = noSplit, lower
		defer func() { term.NoSplitCells, term.LowerHalf = wasN, wasL }()
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "anim"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "template.html"), []byte("{{fxjs}}\n{{fxcss}}\n{{cover}}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		page, err := sitePage(7, dir)
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	plain := build(false, false)
	apple := build(true, true) // what main() sets under TERM_PROGRAM=Apple_Terminal
	if plain != apple {
		t.Fatalf("the page changes with the terminal switches: %d vs %d bytes", len(plain), len(apple))
	}
}
