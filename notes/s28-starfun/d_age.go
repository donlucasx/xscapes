package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/term"
)

// candidateD -- AGE. The oldest stars cooler or dimmer than the newest, so the
// sky carries recency for free.
func candidateD() {
	fmt.Println("\n\n==== 5. CANDIDATE D -- AGE ====")

	fmt.Println("\n  D1 -- AGE BY BRIGHTNESS. Dead on arrival, and the number says why:")
	fmt.Println("  todoStarFloor is 0.85 and it is a FLOOR, not a setting -- it exists so the")
	fmt.Println("  channel is legible at NOON, where StarVis is 0. Alpha may run 0.85..1.00.")
	p := placeLive(12)[11]
	var idx []int
	fmt.Printf("  %-8s %-18s %-8s %s\n", "alpha", "rendered fg", "index", "luma")
	for _, a := range []float64{0.85, 0.90, 0.95, 1.00} {
		s := warm(0, 0, CTX)
		s.plotStar(p, s.pal.Star, a)
		_, fg, _ := s.c.ResolveAt(p.x, p.y, term.Profile256)
		idx = append(idx, fg.Index256())
		fmt.Printf("  %-8.2f %-18s %-8d %.1f\n", a, rgbs(fg), fg.Index256(), luma(fg))
	}
	fmt.Printf("  ⇒ %d distinct tones in the whole legal range. An age gradient over 32 stars\n", distinct(idx))
	fmt.Printf("    needs 32. REJECT D1 -- and going below the floor is not a trade, it hides\n")
	fmt.Println("    the oldest stars at midday, which is a COUNT ERROR by another door.")

	fmt.Println("\n  D2 -- AGE BY HUE (cool oldest -> warm newest, alpha held at the floor)")
	fmt.Println("  Brightness is out, so try the other axis. ⚠ The glyph path saturates chroma")
	fmt.Println("  2.6x (term.GlyphBoost) before quantising, so a gentle hue walk is amplified")
	fmt.Println("  into whatever the cube happens to hold. Measured on the rendered frame:")
	cool := term.RGB{R: 185, G: 205, B: 255}
	warmc := term.RGB{R: 255, G: 225, B: 180}
	for _, n := range []int{12, 26, 32} {
		s := warm(0, 0, CTX)
		ps := placeLive(n)
		var got []int
		var lum []float64
		for i, q := range ps {
			t := float64(i) / float64(maxi(1, len(ps)-1))
			s.plotStar(q, lerpRGB(cool, warmc, t), Floor)
		}
		for _, q := range ps {
			_, fg, _ := s.c.ResolveAt(q.x, q.y, term.Profile256)
			got = append(got, fg.Index256())
			lum = append(lum, luma(fg))
		}
		lo, hi := 999.0, -1.0
		for _, l := range lum {
			if l < lo {
				lo = l
			}
			if l > hi {
				hi = l
			}
		}
		fmt.Printf("  %2d stars -> %2d distinct cube entries rendered, luma spread %.0f..%.0f (%.0f)\n",
			n, distinct(got), lo, hi, hi-lo)
		if n == 26 {
			show("  D2 at 26 stars (the ASCII is identical to the control -- that IS the point,", asciiQuiet(s))
			fmt.Println("   the whole candidate lives in colour; run with -color to see it):")
		}
	}
	fmt.Println("  ⚠ And note the LUMA SPREAD in that table: a hue walk in this cube is a")
	fmt.Println("    brightness walk too -- 46 luma between oldest and newest at 26 stars. That")
	fmt.Println("    is D1 arriving through a side door, and it lands on the same floor.")
	fmt.Println("  ⇒ WEAK PASS ON THE RULE, REJECT ON THE MEASUREMENT. It binds recency, which")
	fmt.Println("    is a second variable on a channel that carries count -- allowed only if")
	fmt.Println("    nothing else needs the channel, and count does. And the cube cannot carry")
	fmt.Println("    it: a handful of entries over 32 stars is not a gradient, it is two or")
	fmt.Println("    three CLASSES, which is candidate C wearing a different coat.")
	fmt.Println("  ⇒ The honest reduction of D is A: if the only readable distinction is")
	fmt.Println("    \"newest vs the rest\", say it once, at the moment it happens, and let go.")
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
