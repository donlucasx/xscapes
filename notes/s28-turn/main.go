// Command turn asks whether HIS 2026-09-10 idea is reachable in this medium:
// the Claude app's mascot "moves from one side to another and uses colors to
// give the illusion of a 3d perspective when it turns."
//
// Two things have to be true and neither was known:
//  1. the sprite path must be able to paint MORE THAN ONE colour on one body
//  2. the 256 cube must actually HOLD a darker and a lighter salmon, or the
//     shading collapses to one tone and the turn reads as a flicker
//
// The coat is CrabCoat, cube index 210, locked 2026-09-07 precisely because it
// is an exact cube entry. This prints the cube-exact neighbours around it, and
// what the glyph path's 2.6x saturate does to each -- a shade that survives
// quantisation on paper and is eaten by GlyphBoost at runtime is no shade.
//
//	go run ./notes/s28-turn
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	coat := companion.CrabCoat
	fmt.Printf("CrabCoat  rgb(%d,%d,%d)  index %d  cube-exact: %v\n\n",
		coat.R, coat.G, coat.B, coat.Index256(), exact(coat))

	// The cube's six levels per channel. A tone is cube-exact only if every
	// channel is one of these, which is what makes it survive quantisation.
	lv := []uint8{0, 95, 135, 175, 215, 255}
	fmt.Println("candidate shades, cube-exact only, sorted dark -> light:")
	fmt.Printf("%-16s %6s %8s %9s   %s\n", "rgb", "index", "luma", "vs coat", "after the 2.6x glyph boost")
	type cand struct {
		c    term.RGB
		luma float64
	}
	var out []cand
	for _, r := range lv {
		for _, g := range lv {
			for _, b := range lv {
				c := term.RGB{R: r, G: g, B: b}
				// Same hue family as the coat: red dominant, green == blue,
				// which is what keeps a shaded crab a crab and not a rust
				// stain. The coat itself is 255,135,135.
				if !(r > g && g == b) {
					continue
				}
				out = append(out, cand{c, luma(c)})
			}
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].luma < out[i].luma {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	for _, o := range out {
		c := o.c
		boosted := c.Saturate(2.6)
		mark := ""
		if c == coat {
			mark = "  <- THE COAT"
		}
		fmt.Printf("rgb(%3d,%3d,%3d) %6d %8.1f %9.1f   -> rgb(%3d,%3d,%3d) idx %d %s%s\n",
			c.R, c.G, c.B, c.Index256(), o.luma, o.luma-luma(coat),
			boosted.R, boosted.G, boosted.B, boosted.Index256(),
			stable(c, boosted), mark)
	}
	fmt.Println("\n'boost KEEPS it' means the 2.6x saturate lands on the same cube index, so the")
	fmt.Println("shade you author is the shade that reaches the screen. The coat was chosen for")
	fmt.Println("exactly that property and any shade paired with it needs it too.")
}

func luma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

func exact(c term.RGB) bool {
	return term.FromIndex256(c.Index256()) == c
}

func stable(c, boosted term.RGB) string {
	if boosted.Index256() == c.Index256() {
		return "boost KEEPS it"
	}
	return "boost MOVES it"
}
