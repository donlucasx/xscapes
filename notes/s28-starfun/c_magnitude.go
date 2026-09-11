package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/term"
)

// candidateC -- MAGNITUDE. Every fifth star brighter, so the eye counts in
// groups the way it reads a tally in fives.
//
// The risk named up front: it invites "what do the big ones MEAN?", which is
// the exact failure the outstanding-todo ring was killed for on 2026-09-05 --
// his words, "discard the ring altogether, it's not clear what it means".
// But the thing that actually decides it is arithmetic.
func candidateC() {
	fmt.Println("\n\n==== 4. CANDIDATE C -- MAGNITUDE, A TALLY IN FIVES ====")
	white := term.RGB{R: 255, G: 255, B: 255}

	for _, n := range []int{12, 26} {
		s := warm(0, 0, CTX)
		ps := placeLive(n)
		for i, p := range ps {
			if (i+1)%5 == 0 {
				s.plotStar(p, white, 1.0)
			} else {
				s.plotStar(p, s.pal.Star, Floor)
			}
		}
		show(fmt.Sprintf("  C at %d stars -- every 5th at index 231 (luma 255) against 153 (207):", n), asciiQuiet(s))
	}

	fmt.Println("  A TALLY ONLY WORKS IF THE GROUPS ARE GROUPS. Each bright star should own")
	fmt.Println("  four faint ones; measured as a nearest-bright partition of the real")
	fmt.Println("  positions (cell distance hypot(dx,2dy), because a cell is twice as tall):")
	fmt.Printf("  %-8s %-14s %-28s %s\n", "stars", "bright stars", "faint stars owned by each", "should be")
	for _, n := range []int{12, 19, 26, 32} {
		ps := placeLive(n)
		var big []pt
		var small []pt
		for i, p := range ps {
			if (i+1)%5 == 0 {
				big = append(big, p)
			} else {
				small = append(small, p)
			}
		}
		own := make([]int, len(big))
		for _, sp := range small {
			bi, bd := -1, math.MaxFloat64
			for i, b := range big {
				d := math.Hypot(float64(sp.x-b.x), 2*float64(sp.y-b.y))
				if d < bd {
					bd, bi = d, i
				}
			}
			if bi >= 0 {
				own[bi]++
			}
		}
		fmt.Printf("  %-8d %-14d %-28s %s\n", n, len(big), fmt.Sprint(own), "4 each")
	}
	fmt.Println("\n  ⇒ REJECT C, and be precise about WHY, because the placement changes half of")
	fmt.Println("    the answer. On HEAD's golden ratio the partition is [9 1 1 1 9] at 26 stars")
	fmt.Println("    -- hopeless. On the tree's blue noise it is [3 4 4 4 6], which is close to")
	fmt.Println("    four. THE COUNTS BEING RIGHT IS NOT THE POINT: a tally in fives reads")
	fmt.Println("    because the five marks TOUCH. Spread over 119 columns and eight rows you")
	fmt.Println("    cannot SEE which four belong to which bright one, so you count them one at")
	fmt.Println("    a time anyway -- now with two weights to reconcile instead of one.")
	fmt.Println("  ⇒ The only arrangement where a tally reads is a ROW, and shore.go rules that")
	fmt.Println("    out in its own words: \"A progress bar in the sky is a piece of UI and the")
	fmt.Println("    brief cuts anything that makes this feel like a dashboard.\"")
	fmt.Println("  ⇒ And the ring's lesson stands: a second weight in the sky is a second")
	fmt.Println("    MEANING to a reader who was not told the rule.")
}

// treeMagnitude measures the magnitude that is ALREADY IN THIS WORKING TREE.
// A second agent added todoStarMag while this study was running: magnitude
// fixed by the slot's index and the seed, running UPWARD from the floor.
//
// That is a different claim from candidate C and it dodges C's rejection --
// it is texture, not a tally, so it never asks you to count in groups. What it
// inherits is the other risk: a second weight in the sky is a second meaning to
// a reader who was not told the rule, which is what killed the ring on
// 2026-09-05. That part is HIS call. What can be measured is whether the faint
// ones stay countable, and at what hour.
func treeMagnitude() {
	fmt.Println("\n  4b. THE MAGNITUDE ALREADY IN THE WORKING TREE (not mine, and not candidate C)")
	fmt.Println("  Per-slot brightness, upward from the floor. Measured on the rendered frame at")
	fmt.Println("  four hours, 26 stars -- the bar the suite holds a star to is 40 luma of")
	fmt.Println("  contrast against the background it is painted on.")
	fmt.Printf("  %-10s %-14s %-16s %-16s %s\n", "hour", "distinct fg", "faintest star", "brightest star", "verdict")
	for _, h := range []struct {
		name string
		t    float64
	}{{"02:00", 2.0 / 24}, {"08:00", 8.0 / 24}, {"12:00", 12.0 / 24}, {"17:45", 17.75 / 24}} {
		sh := warmAt(26, 32, CTX, h.t)
		ps := starCells(sh.c)
		idx := map[int]bool{}
		lo, hi := 9999.0, -1.0
		for _, p := range ps {
			_, fg, bg := sh.c.ResolveAt(p.x, p.y, term.Profile256)
			idx[fg.Index256()] = true
			d := luma(fg) - luma(bg)
			if d < lo {
				lo = d
			}
			if d > hi {
				hi = d
			}
		}
		v := "ok"
		if lo < 40 {
			v = "BELOW THE 40 BAR"
		}
		fmt.Printf("  %-10s %-14d %-16.1f %-16.1f %s  (%d stars)\n", h.name, len(idx), lo, hi, v, len(ps))
	}
	fmt.Println("\n  ⚠ 08:00 IS BELOW THE BAR. Swept over the whole day, 26 and 32 stars,")
	fmt.Println("    counting stars whose rendered contrast against their own background is")
	fmt.Println("    under 40 luma -- this is the tree's constellation, not a candidate:")
	worst := 0.0
	worstH := ""
	for half := 0; half < 48; half++ {
		tod := float64(half) / 48
		var bad int
		var lo = 9999.0
		for _, n := range []int{26, 32} {
			sh := warmAt(n, 32, CTX, tod)
			for _, p := range starCells(sh.c) {
				_, fg, bg := sh.c.ResolveAt(p.x, p.y, term.Profile256)
				d := luma(fg) - luma(bg)
				if d < 40 {
					bad++
				}
				if d < lo {
					lo = d
				}
			}
		}
		if bad > 0 {
			fmt.Printf("    %02d:%02d  %2d stars under 40, faintest %.1f\n",
				half/2, (half%2)*30, bad, lo)
		}
		if lo < worst || worstH == "" {
			worst, worstH = lo, fmt.Sprintf("%02d:%02d", half/2, (half%2)*30)
		}
	}
	fmt.Printf("    worst of the day: %.1f at %s\n", worst, worstH)
	fmt.Println("  ⚠ NOT A DISCOVERY, AND SAY SO. Their own TestEveryStarReadsAtEveryHour is")
	fmt.Println("    already RED in this tree on the same bar (dimmest +9.4 at 40x12), so the")
	fmt.Println("    work is mid-flight and this is an independent confirmation, not a catch.")
	fmt.Println("    What it does add is the HOUR: the failures cluster at 07:30-08:00 and")
	fmt.Println("    13:30-14:00, which is the working day, not the night the sky was tuned in.")

	fmt.Println("  ⇒ Magnitude-as-texture is NOT candidate C and my rejection of C does not")
	fmt.Println("    reach it: it never asks the eye to group. The open question is the ring's,")
	fmt.Println("    and it is his: does a second weight read as a second MEANING?")
}
