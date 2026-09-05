package canvas

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/term"
)

// A split cell on a ramp row: the half that is plain sky (the painter passed
// the cell's own background, as the disc painter does for the half outside
// the disc) must take the RAMP's tone for that quarter, the same tone its
// neighbours show. His 2026-09-05 report, Terminal.app 133x54 by day: "a thin
// line underneath the sun" -- the three cells under the disc's bottom tip
// resolved their sky half by rounding the cell's own colour (5fafd7) while
// the row around them took the path's tone (87afd7).
func TestATipCellsSkyHalfTakesTheRampsTone(t *testing.T) {
	was := term.Ramps
	term.Ramps = true
	defer func() { term.Ramps = was }()
	ramp := term.NewRamp(term.RGB{R: 0, G: 95, B: 175}, term.RGB{R: 175, G: 215, B: 255})
	moon := term.RGB{R: 215, G: 175, B: 135}
	// A span where rounding the cell's own colour and walking the path
	// disagree, so the test can fail: the same disagreement the report was.
	var t0, t1 float64
	found := false
	for a := 0.0; a < 0.9 && !found; a += 0.02 {
		c := New(1, 1, AlphaFar, AlphaMid, AlphaNear)
		c.SetBGRamp(0, 0, ramp, a, a+0.1)
		if term.Profile256.Quantise(c.BGAt(0, 0), false) != ramp.Tone(a+0.075) {
			t0, t1, found = a, a+0.1, true
		}
	}
	if !found {
		t.Fatal("no span separates rounding from the path; the test cannot fail")
	}
	c := New(5, 1, AlphaFar, AlphaMid, AlphaNear)
	for x := 0; x < 5; x++ {
		c.SetBGRamp(x, 0, ramp, t0, t1)
	}
	// The disc's bottom tip in the middle cell: moon above, sky below.
	c.SetBGHalves(2, 0, moon, c.BGAt(2, 0))
	_, _, want := c.ResolveAt(1, 0, term.Profile256)
	ch, _, got := c.ResolveAt(2, 0, term.Profile256)
	if ch != '▀' && ch != '▄' {
		t.Fatalf("tip cell drew %q, want a split cell", ch)
	}
	if got != want {
		t.Errorf("tip cell's sky half is %v, its neighbour's lower quarter is %v", got, want)
	}
	// And the moon half is still the moon.
	_, fg, _ := c.ResolveAt(2, 0, term.Profile256)
	if fg != term.Profile256.Quantise(moon, false) {
		t.Errorf("tip cell's moon half is %v, want %v", fg, term.Profile256.Quantise(moon, false))
	}
}
