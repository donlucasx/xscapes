package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A quad cell is drawn so that no ink touches the cell's top edge whenever
// the top two quarters agree (Terminal.app's block glyphs start five pixels
// down); the mask picks the block, and both colours are backgrounds. On a
// ramp cell the ground takes the ramp's tone, like the cells beside it.
func TestAQuadCellKeepsItsTopAsBackgroundWhereItCan(t *testing.T) {
	moon := term.RGB{R: 215, G: 175, B: 135}
	sky := term.RGB{R: 0, G: 95, B: 175}
	for _, tc := range []struct {
		mask   uint8
		wantCh rune
		fgMoon bool // the moon is the FOREGROUND (ink), else the background
	}{
		{0b0011, '▄', true},  // moon in the lower half: ink below, sky on top
		{0b0001, '▗', true},  // lower-right only
		{0b1100, '▄', false}, // moon across the top: sky inked BELOW, moon as ground
		{0b1110, '▗', false}, // all but lower-right: ink the sky there
		{0b1000, '▘', true},  // top quarters differ: ink must touch the top
		{0b1010, '▌', true},
		{0b0110, '▞', true},
	} {
		c := New(1, 1, AlphaFar, AlphaMid, AlphaNear)
		c.SetBG(0, 0, sky)
		c.SetBGQuad(0, 0, moon, sky, tc.mask)
		ch, fg, bg := c.ResolveAt(0, 0, term.ProfileTrueColor)
		wantFG, wantBG := moon, sky
		if !tc.fgMoon {
			wantFG, wantBG = sky, moon
		}
		if ch != tc.wantCh || fg != wantFG || bg != wantBG {
			t.Errorf("mask %04b: got %q fg %v bg %v, want %q fg %v bg %v", tc.mask, ch, fg, bg, tc.wantCh, wantFG, wantBG)
		}
	}
	// Full and empty masks are flat.
	c := New(2, 1, AlphaFar, AlphaMid, AlphaNear)
	c.SetBG(0, 0, sky)
	c.SetBGQuad(0, 0, moon, sky, 0b1111)
	if ch, _, bg := c.ResolveAt(0, 0, term.ProfileTrueColor); ch != ' ' || bg != moon {
		t.Errorf("full mask: got %q on %v, want a flat moon cell", ch, bg)
	}
	c.SetBG(1, 0, sky)
	c.SetBGQuad(1, 0, moon, sky, 0)
	if ch, _, bg := c.ResolveAt(1, 0, term.ProfileTrueColor); ch != ' ' || bg != sky {
		t.Errorf("empty mask: got %q on %v, want the sky untouched", ch, bg)
	}
	// On a ramp row, the ground is the ramp's tone, the same as the neighbour.
	was := term.Ramps
	term.Ramps = true
	defer func() { term.Ramps = was }()
	ramp := term.NewRamp(term.RGB{R: 0, G: 95, B: 175}, term.RGB{R: 175, G: 215, B: 255})
	c = New(3, 1, AlphaFar, AlphaMid, AlphaNear)
	for x := 0; x < 3; x++ {
		c.SetBGRamp(x, 0, ramp, 0.42, 0.48)
	}
	c.SetBGQuad(1, 0, moon, c.BGAt(1, 0), 0b0011)
	_, _, want := c.ResolveAt(0, 0, term.Profile256)
	if ch, _, bg := c.ResolveAt(1, 0, term.Profile256); ch != '▄' || bg != want {
		t.Errorf("on a ramp: got %q on %v, want '▄' on the neighbour's tone %v", ch, bg, want)
	}
}
