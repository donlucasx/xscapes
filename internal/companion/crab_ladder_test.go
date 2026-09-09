package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// Hero's litter shrinks as it grows, on the cat's ladder and the cat's
// thresholds.
//
// It was the last "partial" on the crab's API table and the cost was measured
// rather than assumed. Folded through his own 341 hours of recordings, the
// litter peaks at 38 subagents with a p90 of 14; at ONE size, 80 columns holds
// eight sitters, so the crab dropped subagents 12.6% of the time the litter
// existed, as many as 18 at once, while the cat shrank them and kept drawing.
func TestTheCrabsLitterShrinksAsItGrows(t *testing.T) {
	// Every rung has to keep its eyes: the sockets are holes in the bitmap
	// with a character plotted on top, so a rung whose body fills them loses
	// the face and the litter becomes a row of blocks.
	for tier := range crabTiers {
		const w, h = 40, 6
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat := New("crab")
		cat.kitTier = tier
		// Enough sitters to hold this tier under TierFor's hysteresis.
		n := []int{1, 7, 11}[tier]
		cat.drawCrabKittens(c.Near(), 34, 0, n, w, 0, 0, 0.5, 7)
		eyes := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == 'o' {
					eyes++
				}
			}
		}
		if eyes < 2 {
			t.Errorf("tier %d: %d eye cells on screen, want at least one pair -- the bitmap is filling its own sockets", tier, eyes)
		}
	}

	// And the ladder buys real room. At 80 columns the companion stands at 64
	// and paces 5 inward, so the litter starts at 58.
	const px = 58
	fits := func(tier, w int) int {
		cw := crabWidth(tier)
		n := 0
		for px-1-cw-n*(cw+1) >= 0 {
			n++
		}
		return n
	}
	big, small := fits(0, 80), fits(2, 80)
	if small <= big {
		t.Errorf("the smallest rung fits %d sitters and the largest %d: the ladder buys nothing", small, big)
	}
	t.Logf("at 80 columns the litter holds %d sitters at full size and %d at the smallest rung", big, small)
}

func crabWidth(tier int) int { return ParseBitmap(crabTiers[tier].rows).W / 2 }
