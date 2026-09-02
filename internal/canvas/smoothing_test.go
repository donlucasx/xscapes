package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// ramp fills a canvas with a vertical gradient between two colours.
func ramp(h int, a, b term.RGB) *Canvas {
	c := New(8, h, AlphaFar, AlphaMid, AlphaNear)
	for y := 0; y < h; y++ {
		col := term.Lerp(a, b, float64(y)/float64(h-1))
		for x := 0; x < c.W; x++ {
			c.SetBG(x, y, col)
		}
	}
	return c
}

// rampError measures how far the drawn frame is from the gradient it was asked
// for, sampled at HALF-ROW resolution -- which is the resolution the terminal
// actually gets, because a split cell paints its top half in the foreground and
// its bottom half in the background.
//
// This is the right question, and the first version of this test asked the
// wrong one. It counted distinct tones and found 7 with the split and 7
// without, which is correct and proves nothing: splitting a cell introduces no
// NEW colours, it places the edge between two existing ones twice as precisely.
// The gain is accuracy, not palette.
func rampError(c *Canvas, x int, a, b term.RGB) float64 {
	total := 0.0
	for y := 0; y < c.H; y++ {
		ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
		up, dn := bg, bg
		if ch == '\u2580' {
			up = fg
		}
		// the gradient's true colour at the centre of each half-row
		for i, got := range []term.RGB{up, dn} {
			t := (float64(y) + 0.25 + 0.5*float64(i)) / float64(c.H-1)
			want := term.Lerp(a, b, t)
			dr := float64(got.R) - float64(want.R)
			dg := float64(got.G) - float64(want.G)
			db := float64(got.B) - float64(want.B)
			total += 0.30*dr*dr + 0.59*dg*dg + 0.11*db*db
		}
	}
	return total / float64(c.H*2)
}

// Splitting cells must make the frame CLOSER to the gradient it was asked for.
// If it does not, every U+2580 on screen is being paid for and bought nothing.
func TestSplittingCellsTracksTheGradientMoreClosely(t *testing.T) {
	// A noon sky: #005fd7 at the zenith to #afd7ff at the horizon.
	a := term.RGB{R: 0, G: 95, B: 215}
	b := term.RGB{R: 175, G: 215, B: 255}
	c := ramp(17, a, b)

	saved := term.Shading
	defer func() { term.Shading = saved }()

	term.Shading = false
	off := rampError(c, 3, a, b)
	term.Shading = true
	on := rampError(c, 3, a, b)

	t.Logf("17-row sky, mean squared error against the true ramp: %.0f without the split, %.0f with (%.0f%% lower)",
		off, on, 100*(off-on)/off)
	if on >= off {
		t.Errorf("the split tracks the gradient no better: %.0f with it, %.0f without", on, off)
	}
}

// The glyph owns its cell. A background that has to show through a wave or a
// letter cannot also be two colours, and splitting one would eat the glyph.
func TestSplittingNeverTouchesACellWithAGlyphInIt(t *testing.T) {
	c := ramp(17, term.RGB{R: 0, G: 95, B: 215}, term.RGB{R: 175, G: 215, B: 255})
	saved := term.Shading
	defer func() { term.Shading = saved }()
	term.Shading = true

	for y := 0; y < c.H; y++ {
		c.Near().Plot(3, y, 'x', term.RGB{R: 255, G: 255, B: 255}, 1)
	}
	for y := 0; y < c.H; y++ {
		if ch, _, _ := c.ResolveAt(3, y, term.Profile256); ch != 'x' {
			t.Fatalf("row %d: the glyph was replaced by %q", y, ch)
		}
	}
	// And the column beside it, with no glyph, must still be splitting.
	split := 0
	for y := 0; y < c.H; y++ {
		if ch, _, _ := c.ResolveAt(4, y, term.Profile256); ch == '▀' {
			split++
		}
	}
	if split == 0 {
		t.Error("no cell split in a plain gradient column -- the control is dead")
	}
}

// Truecolor must be untouched: the split is a workaround for a palette that
// cannot hold the ramp, and a terminal that can hold it should get the real
// colours and a plain cell.
func TestTruecolorIsNeverSplit(t *testing.T) {
	c := ramp(17, term.RGB{R: 0, G: 95, B: 215}, term.RGB{R: 175, G: 215, B: 255})
	saved := term.Shading
	defer func() { term.Shading = saved }()
	term.Shading = true
	for y := 0; y < c.H; y++ {
		ch, _, bg := c.ResolveAt(3, y, term.ProfileTrueColor)
		if ch != ' ' {
			t.Fatalf("row %d: truecolor got glyph %q", y, ch)
		}
		if want := c.BGAt(3, y); bg != want {
			t.Fatalf("row %d: truecolor background changed, %v want %v", y, bg, want)
		}
	}
}
