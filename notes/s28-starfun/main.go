// Command starfun is a STUDY, not a change. It renders candidate answers to
// his open question of 2026-09-10 -- "any other ways we can make the stars more
// fun" -- the way Hero was picked from four silhouettes and the coat from five.
//
// Nothing here ships. Every candidate is judged against the encoding rule
// before it is judged on looks:
//
//  1. does it keep the stars COUNTABLE at a glance?
//  2. does it encode in coverage / count / position, NEVER in rate?
//  3. does it bind a SECOND variable to a channel that already has one?
//
// A candidate that fails any of the three is rejected here with the measurement
// that killed it, which is a real result and cheaper than shipping it.
//
// The distinction that decides most of them, stated once: a transient that
// ANNOUNCES a change to a statically-encoded variable is not a rate encoding.
// The count is fully readable from a still frame either way; the transient only
// says "look". A cue that carries information a still frame does NOT hold --
// which star is newest, say -- IS a rate encoding, and the rule kills it.
//
//	go run ./notes/s28-starfun          # the study
//	go run ./notes/s28-starfun -color   # the same frames in real 256 colour
package main

import (
	"flag"
	"fmt"
)

var colorOut = flag.Bool("color", false, "emit the frames as real 256-colour ANSI")

func main() {
	flag.Parse()
	fmt.Println("STARFUN -- candidates for the constellation, rendered at 125x28, night, 56% context.")
	fmt.Println("Every glyph below is read back off the RENDERED frame with canvas.ResolveAt.")

	check()
	control()
	candidateA()
	candidateB()
	candidateC()
	treeMagnitude()
	candidateD()
	candidateE()
	candidateK()
	paperRejects()
	verdict()
	if *colorOut {
		colorRun()
	}
}

// check says WHICH sky this run measured. Every candidate below is drawn on
// positions read back off the product's own rendered frame (placeLive), so the
// study follows whatever todoStars does today -- but the reader still has to
// know whether that is HEAD's constellation or someone's working copy.
//
// ⚠ THIS IS NOT DECORATION. The first version of this instrument replicated
// todoStars' arithmetic instead of reading the frame. A second agent changed
// the placement in this tree at 20:36:30 while the study was running, and the
// control silently became a different sky. The replica is kept ONLY as the
// fingerprint that catches that.
func check() {
	fmt.Println("\n\n==== 0. WHICH SKY IS THIS? ====")
	fmt.Println("Candidates are drawn on positions read off the RENDERED frame, so they follow")
	fmt.Println("whatever todoStars does today. placeHead() is HEAD c8efcc7's golden-ratio")
	fmt.Println("arithmetic, kept as a fingerprint: MATCH means this run measured the shipped")
	fmt.Println("constellation, MISMATCH means the working tree has moved on.")
	ok := true
	for _, done := range []int{12, 19, 26, 32} {
		live := warm(done, 32, CTX)
		got := starCells(live.c)
		blank := warm(0, 0, CTX)
		want := placeHead(blank, 32, done)
		same := len(got) == len(want)
		if same {
			seen := map[pt]bool{}
			for _, p := range got {
				seen[p] = true
			}
			for _, p := range want {
				if !seen[p] {
					same = false
				}
			}
		}
		if !same {
			ok = false
		}
		lo, hi := bandOf(got)
		fmt.Printf("  %2d earned -> %2d '*' rendered, rows %d..%d, merging pairs %d   vs HEAD: %s\n",
			done, len(got), lo, hi, merges(got),
			map[bool]string{true: "MATCH", false: "MISMATCH"}[same])
	}
	if !ok {
		fmt.Println("  ⇒ THE WORKING TREE'S CONSTELLATION IS NOT HEAD'S. Everything below was")
		fmt.Println("    measured on the tree's sky, which is the right thing to measure and the")
		fmt.Println("    wrong thing to quote as \"today\".")
	} else {
		fmt.Println("  ⇒ This run measured HEAD's shipped constellation.")
	}
}
