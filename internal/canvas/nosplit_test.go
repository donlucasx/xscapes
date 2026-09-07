package canvas

import (
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// His ruling of 2026-09-07, after his own line-spacing test came back "no, it
// didnt": the ramp's split cells are NOT DRAWN on Terminal.app, because the
// half block leaves a 1 device pixel rule at the cell's bottom edge whichever
// way up it is drawn and nothing in the terminal can close it.
//
// He ruled the disc and the shoreline KEEP their split cells, so this asserts
// the ramp only. A test that demanded zero split cells everywhere would be
// asserting something he explicitly did not ask for.
//
// Both LowerHalf states are swept. Every test in this tree runs at LowerHalf's
// zero value, which is U+2580 -- the glyph that does NOT ship on his machine --
// so a guard written for one orientation is blind to the one he sees.
func TestTheRampIsOneToneACellWhenSplitCellsAreOff(t *testing.T) {
	for _, lower := range []bool{false, true} {
		wasL, wasN, wasS := term.LowerHalf, term.NoSplitCells, term.Shading
		term.LowerHalf, term.Shading = lower, true

		term.NoSplitCells = false
		split := rampSplitCells(t)
		term.NoSplitCells = true
		collapsed := rampSplitCells(t)

		term.LowerHalf, term.NoSplitCells, term.Shading = wasL, wasN, wasS

		if split == 0 {
			t.Fatalf("LowerHalf=%v: the ramp drew NO split cells with the flag off — "+
				"this test cannot detect the collapse and proves nothing", lower)
		}
		if collapsed != 0 {
			t.Errorf("LowerHalf=%v: %d ramp cells are still two colours with NoSplitCells on", lower, collapsed)
		}
		t.Logf("LowerHalf=%v: %d split ramp cells -> %d", lower, split, collapsed)
	}
}

// rampSplitCells paints a ramp down a canvas and counts the cells that resolve
// to two colours. The positive control is the caller's: with the flag off this
// must be non-zero, or the count is measuring nothing.
func rampSplitCells(t *testing.T) int {
	t.Helper()
	const w, h = 40, 12
	c := New(w, h, AlphaFar, AlphaMid, AlphaNear)
	r := term.NewRamp(term.RGB{R: 20, G: 40, B: 90}, term.RGB{R: 210, G: 170, B: 130})
	for y := 0; y < h; y++ {
		t0 := float64(y) / float64(h)
		t1 := float64(y+1) / float64(h)
		for x := 0; x < w; x++ {
			c.SetBGRamp(x, y, r, t0, t1)
		}
	}
	n := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
			if ch != ' ' && fg != bg {
				n++
			}
		}
	}
	return n
}

// An HTML page is not a terminal: it has no half-block hairline to avoid, and
// the studies exist to SHOW the split. main() sets NoSplitCells from
// TERM_PROGRAM before any subcommand runs, so without the guard in
// htmlFragmentWith the published page would change with the machine it was
// built on -- collapsed from Terminal.app, split from Ghostty or CI.
func TestHTMLIsTheSameWhicheverTerminalBuiltIt(t *testing.T) {
	const w, h = 20, 6
	build := func(noSplit bool) string {
		was := term.NoSplitCells
		term.NoSplitCells = noSplit
		defer func() { term.NoSplitCells = was }()
		c := New(w, h, AlphaFar, AlphaMid, AlphaNear)
		r := term.NewRamp(term.RGB{R: 20, G: 40, B: 90}, term.RGB{R: 210, G: 170, B: 130})
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c.SetBGRamp(x, y, r, float64(y)/float64(h), float64(y+1)/float64(h))
			}
		}
		return c.HTMLFragmentAs(11, term.Profile256)
	}
	on, off := build(true), build(false)
	if on != off {
		t.Errorf("the HTML changed with the terminal flag: %d bytes vs %d", len(on), len(off))
	}
	// Positive control: the fragment must actually contain a split cell, or
	// "they are equal" is the trivial equality of two collapsed pages.
	if !strings.Contains(off, "▀") && !strings.Contains(off, "▄") {
		t.Fatal("no half block in the HTML at all — this test cannot detect the leak")
	}
}
