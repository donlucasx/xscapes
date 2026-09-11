// Command a-step draws DIRECTION A of his 2026-09-10 idea -- "can we animate
// the main Agent/companion to come CLOSER to the user before it prompts it?
// So the companion (crab, in this case) would walk up closer to the screen and
// get bigger."
//
// A is the minimal step: 12x7 cells today -> 13x8 -> 14x9. Every pose is
// rendered through companion.ParseBitmap and ToQuadrant, which is the only way
// to know what the art actually looks like: the medium halves the source
// vertically before it packs quadrants, so a row drawn is not a row seen.
//
//	go run ./notes/s28-closer/a-step
package main

import (
	"flag"
	"fmt"
)

func main() {
	stage := flag.String("stage", "all", "baseline | nn | art | frames | all")
	flag.IntVar(&armPx, "arm", 4, "near claw-arm width in source pixels (2 or 4)")
	flag.Parse()
	nearArmL = span{(8 - armPx) / 2, (8-armPx)/2 + armPx - 1}

	switch *stage {
	case "baseline":
		baseline()
	case "nn":
		nnExperiment()
	case "art":
		art()
	case "frames":
		frames()
	default:
		baseline()
		nnExperiment()
		art()
		frames()
	}
}

func rule(s string) {
	fmt.Printf("\n%s\n%s\n\n", s, "================================================================")
}
