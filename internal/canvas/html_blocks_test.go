package canvas

import (
	"strconv"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// TestAPageNeverAsksAFontForABlock: every half and quarter block the renderer
// can draw comes out of the HTML writers as a space on a background pattern,
// never as the glyph, in the classed path and the inline one. His phone drew
// the moon's U+2584 edge from a fallback font and the rows broke around it
// (2026-09-16); a background has no font to fall back to.
func TestAPageNeverAsksAFontForABlock(t *testing.T) {
	blocks := "▀▄█▌▐▖▗▘▙▚▛▜▝▞▟"
	ink, ground := term.RGB{R: 250, G: 130, B: 130}, term.RGB{R: 20, G: 40, B: 90}
	c := New(len([]rune(blocks))+2, 2, AlphaFar, AlphaMid, AlphaNear)
	c.Clear()
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			c.SetBG(x, y, ground)
		}
	}
	for i, r := range []rune(blocks) {
		c.Near().Plot(i+1, 0, r, ink, 1)
	}
	c.Near().Plot(1, 1, 'x', ink, 1) // a real glyph beside them, for contrast
	pal := &HTMLPalette{}
	classed := c.HTMLFragmentClassed(12, term.ProfileTrueColor, pal)
	inline := c.HTMLFragment(12)
	for _, out := range []string{classed, inline, pal.CSS()} {
		for _, r := range out {
			if r >= 0x2580 && r <= 0x259F {
				t.Fatalf("a block glyph %q reached the page:\n%s", r, out)
			}
		}
	}
	if !strings.Contains(inline, "x") || !strings.Contains(classed, "x") {
		t.Fatal("the real glyph beside the blocks was lost")
	}
	// The half blocks are two-stop gradients, the quadrants four layers, the
	// full block plain ink; and the classes carry them.
	css := pal.CSS()
	if n := strings.Count(css, "linear-gradient(") + strings.Count(css, "conic-gradient("); n < 14 {
		t.Errorf("%d gradients in the stylesheet, want the 14 partial blocks' worth at least:\n%s", n, css)
	}
	if !strings.Contains(css, "background:#fa8282}") {
		t.Errorf("the full block is not a plain ink ground:\n%s", css)
	}
	// One span per partial block, holding one space each: a run of them
	// must not share a background sized to the run.
	blocks14 := 0
	for i, k := range pal.keys {
		if k.mask == 0 || k.mask == maskFull {
			continue
		}
		blocks14++
		if n := strings.Count(classed, `class="p`+strconv.Itoa(i)+`"> </span>`); n != 1 {
			t.Errorf("block class p%d (mask %d) appears %d times as a single-space span, want 1", i, k.mask, n)
		}
	}
	if blocks14 != 14 {
		t.Errorf("%d partial-block classes, want 14", blocks14)
	}
}

// TestTheHalfBlockGradientFollowsTheHalf: the pattern is oriented by the
// mask, so an upper half and a lower half do not paint the same picture.
func TestTheHalfBlockGradientFollowsTheHalf(t *testing.T) {
	var up, down strings.Builder
	a, b := term.RGB{R: 1, G: 2, B: 3}, term.RGB{R: 200, G: 201, B: 202}
	writeBlockCSS(&up, a, b, maskTop, nil)
	writeBlockCSS(&down, a, b, maskBottom, nil)
	if up.String() != "background:linear-gradient(#010203 50%,#c8c9ca 50%)" {
		t.Errorf("upper half: %s", up.String())
	}
	if down.String() != "background:linear-gradient(#c8c9ca 50%,#010203 50%)" {
		t.Errorf("lower half: %s", down.String())
	}
	var q strings.Builder
	writeBlockCSS(&q, a, b, maskTopLeft|maskBottomRight, &b)
	want := "background:conic-gradient(transparent 0 25%,#010203 0 50%,transparent 0 75%,#010203 0)"
	if q.String() != want {
		t.Errorf("quadrants on a transparent ground:\n got %s\nwant %s", q.String(), want)
	}
}
