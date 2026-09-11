package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

// The approach, and what it costs.
//
// Two claims are checked here rather than asserted, because both are the kind
// that sound obviously true and decide whether the whole direction works:
//
//  1. Growing DOWNWARD puts no part of the crab in the sea that is not there
//     today. The companion is drawn at top = H-2-chh, so two rows already sit
//     empty under its feet; if the approach anchors the FEET instead of the
//     top, the top row of ink never moves and the extra rows come out of that
//     empty margin.
//
//  2. The two columns cost the sand are small enough to ignore.

// askLevels are the activity levels his real log shows at a needs_input, from
// notes/s28-askroom. p50 is the one that matters; p90 is the stress case.
var askLevels = []struct {
	name  string
	level float64
}{{"p10", 0.54}, {"p50", 0.59}, {"p90", 0.65}, {"busy=1", 1.00}}

var geoms = []struct{ w, h int }{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}}

func frames() {
	rule("CLAIM 1: growing downward puts no new ink in the sea")
	fmt.Println("Two anchoring rules, measured against the same waterline.")
	fmt.Println()
	fmt.Println("  TOP-anchored is what live.go does today: top = H-2-chh, chh from")
	fmt.Println("      cat.Size(). A taller sprite therefore rises and its HEAD goes")
	fmt.Println("      into the sea, which is the wrong direction for coming closer.")
	fmt.Println("  FEET-anchored is the proposal: top = H-2-7 whatever the sprite's own")
	fmt.Println("      height is, so the extra rows come out of the two empty rows")
	fmt.Println("      already under its feet.")
	fmt.Println()
	fmt.Println("The number is how many rows of the sprite's BOX sit above the waterline.")
	fmt.Println("Lower is better; the shipped crab's number is the one to not exceed.")
	fmt.Println()
	fmt.Printf("%-9s %-7s %7s |%18s |%18s\n", "", "", "", "TOP-anchored", "FEET-anchored")
	fmt.Printf("%-9s %-7s %7s |%6s%6s%6s |%6s%6s%6s\n",
		"geom", "level", "sandTop", "12x7", "13x8", "14x9", "12x7", "13x8", "14x9")
	worstTop, worstFeet, shipped := -99, -99, -99
	for _, g := range geoms {
		for _, a := range askLevels {
			st := sandTop(g.w, g.h, a.level)
			var topAnc, feetAnc [3]int
			for i, h := range []int{7, 8, 9} {
				topAnc[i] = st - (g.h - 2 - h)  // live.go's rule today
				feetAnc[i] = st - (g.h - 2 - 7) // the proposal: feet advance instead
			}
			fmt.Printf("%-9s %-7s %7d |%6d%6d%6d |%6d%6d%6d\n",
				fmt.Sprintf("%dx%d", g.w, g.h), a.name, st,
				topAnc[0], topAnc[1], topAnc[2], feetAnc[0], feetAnc[1], feetAnc[2])
			if a.name != "busy=1" {
				if topAnc[2] > worstTop {
					worstTop = topAnc[2]
				}
				if feetAnc[2] > worstFeet {
					worstFeet = feetAnc[2]
				}
				if topAnc[0] > shipped {
					shipped = topAnc[0]
				}
			}
		}
	}
	fmt.Println()
	fmt.Printf("Worst case at p10..p90, 14x9: TOP-anchored %d rows in the sea, FEET-anchored %d.\n",
		worstTop, worstFeet)
	fmt.Printf("The shipped 12x7 crab's worst case is %d, so FEET-anchored is EXACTLY the\n", shipped)
	fmt.Println("status quo and TOP-anchored is two rows worse. The feet-anchored columns")
	fmt.Println("are identical across the three sizes because the top row of ink does not")
	fmt.Println("move: that is the finding, not an artifact of the table.")
	fmt.Println()
	fmt.Println("40x12 already stands in the water with the shipped crab -- 4 rows at the")
	fmt.Println("median ask. It does not get worse, but it was never good, and 40x12 is the")
	fmt.Println("'must look fine' floor rather than the design target.")

	rule("CLAIM 2: what 14 cells costs the sand -- nothing, if it grows LEFT")
	fmt.Println("compose() puts the companion at catX = w - catW - (2 + w/32) and reserves")
	fmt.Println("a strip of paceSpan(w) columns to its LEFT from both the sand and the")
	fmt.Println("litter, because the crab paces inward through it. So there are already")
	fmt.Println("free columns exactly where a wider crab wants to grow.")
	fmt.Println()
	fmt.Println("Keep the LAYOUT's catW at 12 and draw the near sprite two columns left of")
	fmt.Println("catX. Then its right edge does not move, and the sand's last column has to")
	fmt.Println("clear catX-3 rather than catX-1.")
	fmt.Println()
	fmt.Printf("%-9s %8s %8s %12s %14s %8s\n",
		"geom", "paceSpan", "catX", "last sand col", "near left edge", "clear?")
	allClear := true
	for _, g := range geoms {
		span := paceSpanOf(g.w)
		catX := g.w - 12 - (2 + g.w/32)
		lastSand := catX - 2 - span // sandfade draws x < SandTo
		nearLeft := catX - 2
		ok := lastSand < nearLeft
		if !ok {
			allClear = false
		}
		fmt.Printf("%-9s %8d %8d %12d %14d %8v\n",
			fmt.Sprintf("%dx%d", g.w, g.h), span, catX, lastSand, nearLeft, ok)
	}
	fmt.Printf("\nclear at every geometry: %v -- paceSpan is never below 1, so the last\n", allClear)
	fmt.Println("sand column is always at least one column left of where the near crab")
	fmt.Println("starts. The writing band keeps ALL its columns and never reflows.")
	fmt.Println()
	fmt.Println("The condition that buys this: the pace must be frozen at home during the")
	fmt.Println("approach. It should be anyway -- the crab is walking at the user, not")
	fmt.Println("pacing sideways -- but if dx is left running, the two things add up and")
	fmt.Println("the crab walks over its own litter.")
	fmt.Println()
	fmt.Println("For contrast, what it WOULD cost if catW were simply raised to 14 and the")
	fmt.Println("layout recomputed (the sand reflows every time the crab grows):")
	fmt.Printf("%-9s %10s %10s %8s\n", "geom", "sand @12", "sand @14", "lost")
	for _, g := range geoms {
		a, b := sandCols(g.w, 12), sandCols(g.w, 14)
		pct := 0.0
		if a > 0 {
			pct = 100 * float64(a-b) / float64(a)
		}
		fmt.Printf("%-9s %10d %10d %7.1f%%\n", fmt.Sprintf("%dx%d", g.w, g.h), a, b, pct)
	}

	rule("THE APPROACH")
	fmt.Println("stepDur is 0.28 s and it is a constant on purpose (pace.go: 'a step is")
	fmt.Println("always one cell and always takes stepDur'). The approach is three of")
	fmt.Println("those, so it runs on the clock the crab already walks on:")
	fmt.Println()
	fmt.Println("  t=0.00  needs_input arrives. Ring the cue NOW, not at the end -- the")
	fmt.Println("          sound is what makes him look up, and the walk is what he")
	fmt.Println("          should see when he does.")
	fmt.Println("  t=0.00  frame 1: far 12x7, ask pose, claw up, feet H-3.")
	fmt.Println("  t=0.28  frame 2: mid 13x8, mid-stride legs, feet H-2.")
	fmt.Println("  t=0.56  frame 3: near 14x9, stand legs, feet H-1, THE EYE OPENS.")
	fmt.Println("  t=0.84  the balloon goes up at top-3, where it already is.")
	fmt.Println()
	fmt.Printf("At the default 12 fps (inside.go) that is %d frames of scape: ", 10)
	fmt.Println("3, 3 and 4.")
	fmt.Println("Three is the fewest that reads as walking rather than as a zoom, and 0.84 s")
	fmt.Println("is the most delay worth adding to a notification that carries 30% of the")
	fmt.Println("rubric on its own.")
	fmt.Println()

	fmt.Println("the three upper halves, in order:")
	for i, f := range [][]string{farAsk, midAsk(), nearAsk()} {
		fmt.Printf("\n  frame %d:\n", i+1)
		for _, r := range quad(join(f, [][]string{farLower, midLower(), nearLower()}[i])) {
			fmt.Printf("    |%s|\n", r)
		}
	}
}

func sandTop(w, h int, level float64) int {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	return sh.SandTop()
}

// sandCols reproduces compose()'s mirrored arithmetic for a given companion
// width. Copied rather than called because compose is in package main of the
// root command and this is a separate program; the numbers are checked against
// it by eye, and the formula is three lines long.
func paceSpanOf(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

func sandCols(w, catW int) int {
	right := 2 + w/32
	span := paceSpanOf(w)
	catX := w - catW - right
	if catX < 0 {
		catX = 0
	}
	n := (catX - 1 - span) - 2
	if n < 0 {
		return 0
	}
	return n
}
