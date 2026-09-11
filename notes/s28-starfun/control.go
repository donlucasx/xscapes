package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/term"
)

// control is what he is looking at today. Everything else is measured against
// this frame, not against a description of it.
func control() {
	fmt.Println("\n\n==== 1. THE CONTROL -- TODAY'S SKY ====")
	for _, done := range []int{12, 19, 26} {
		s := warm(done, 32, CTX)
		ps := starCells(s.c)
		rows, cols := map[int]bool{}, map[int]bool{}
		lo, hi := 99, -1
		for _, p := range ps {
			rows[p.y] = true
			cols[p.x] = true
			if p.y < lo {
				lo = p.y
			}
			if p.y > hi {
				hi = p.y
			}
		}
		eaten := done - len(ps)
		show(fmt.Sprintf("CONTROL -- %d earned, %d on screen (%d eaten by the disc), rows %d..%d of 0..%d sky",
			done, len(ps), eaten, lo, hi, s.hy-1), asciiSky(s.c, s.hy))
		if done == 26 {
			show("CONTROL at 26, same frame with the sky ramp's half blocks blanked so the",
				asciiQuiet(s))
			fmt.Println("   constellation is readable -- this is the view every candidate below uses.")
		}
		fmt.Printf("   rows occupied %d of %d sky rows | distinct columns %d | pairs that merge %d\n",
			len(rows), s.hy, len(cols), merges(ps))
	}

	s := warm(12, 32, CTX)
	ps := starCells(s.c)
	if len(ps) > 0 {
		_, fg, bg := s.c.ResolveAt(ps[0].x, ps[0].y, term.Profile256)
		fmt.Printf("\n   THE STAR AS RENDERED: fg %v (index %d, luma %.1f) on sky %v (index %d, luma %.1f)\n",
			fg, fg.Index256(), luma(fg), bg, bg.Index256(), luma(bg))
		fmt.Printf("   contrast %.1f luma. Palette Star is %v at alpha %.2f -- that alpha is the FLOOR,\n",
			math.Abs(luma(fg)-luma(bg)), s.pal.Star, Floor)
		fmt.Printf("   so there are only %.2f of alpha left above it to play with.\n", 1-Floor)
	}
}
