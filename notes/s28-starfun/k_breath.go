package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/term"
)

// candidateK -- THE NEWEST STAR BREATHES. Mine, not his: all the stars steady
// except the most recent, which pulses gently forever, so the sky always says
// "the last thing you finished is that one".
//
// It is the most attractive idea in this study and it is the clearest failure
// of the rule, which is worth showing precisely because it is attractive.
func candidateK() {
	fmt.Println("\n\n==== 7. CANDIDATE K -- THE NEWEST STAR BREATHES (mine, and it fails) ====")
	p := placeLive(12)[11]
	var idx []int
	for _, a := range []float64{0.85, 0.92, 1.00} {
		s := warm(0, 0, CTX)
		s.plotStar(p, s.pal.Star, a)
		_, fg, _ := s.c.ResolveAt(p.x, p.y, term.Profile256)
		idx = append(idx, fg.Index256())
	}
	fmt.Printf("  The breath has %d tones to breathe between (the floor and 1.00).\n", distinct(idx))
	fmt.Println("  ⇒ REJECT K, on the rule and not on the tones:")
	fmt.Println("    \"Encode in coverage, count or position -- NEVER in rate.\" Which star is")
	fmt.Println("    newest would be carried ONLY by motion, and a screenshot has no motion. He")
	fmt.Println("    screenshots this thing constantly; every one would lose the information.")
	fmt.Println("  ⇒ THE LINE THIS DRAWS, and it is the one that decides most of this study:")
	fmt.Println("    a transient that ANNOUNCES a change to something the still frame already")
	fmt.Println("    carries is fine -- the count is right in every frame with or without it.")
	fmt.Println("    A cue that is the ONLY carrier of something is a rate encoding and dies.")
	fmt.Println("    A (arrival) is the first kind. K is the second. D is the second wearing")
	fmt.Println("    a static coat, which is why it only half-escapes.")
	fmt.Println("  ⚠ And it re-opens a live report. On 2026-09-10 he said, inside the installed")
	fmt.Println("    build: \"once they appear, they should not disappear IMO.\" A breathing star")
	fmt.Println("    is a star that looks like it is going out.")
}
