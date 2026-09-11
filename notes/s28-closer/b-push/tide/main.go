// Command tide answers the half of direction B that decides whether the other
// half is worth drawing: an 11-row companion needs 11 rows of beach, and at the
// median ask he has 7 or 8.
//
// The ask fires at level 0.59 because needs_input does not clear turnOpn, so
// TurnFloor/FlightFloor hold the water in (reduce.go:554). Nothing here changes
// reduce.go. It feeds the SHORE a lower level directly and reports what that
// buys, which is the same thing a "withdraw the tide at the ask" rule would do
// downstream.
//
//	go run ./notes/s28-closer/b-push/tide
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

// His geometries, from the brief. 153x51 is the big one; the first four are
// where he actually works.
var geoms = []struct{ w, h int }{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}}

func main() {
	fmt.Printf("tide: %v   TideRange %.1f rows   TideEase %.1fs\n\n", scape.Tide, scape.TideRange, scape.TideEase)

	fmt.Println("BEACH ROWS BY LEVEL  (rows from the waterline to the bottom of the frame,")
	fmt.Println("inclusive -- a companion N rows tall standing on the bottom row needs N)")
	fmt.Println()
	fmt.Printf("%-6s", "level")
	for _, g := range geoms {
		fmt.Printf("%9s", fmt.Sprintf("%dx%d", g.w, g.h))
	}
	fmt.Printf("   %s\n", "min of the four he works in")
	for lv := 0.0; lv <= 0.701; lv += 0.05 {
		fmt.Printf("%-6.2f", lv)
		minWork := 999
		for i, g := range geoms {
			b := beach(g.w, g.h, lv)
			fmt.Printf("%9d", b)
			if i < 4 && b < minWork {
				minWork = b
			}
		}
		fmt.Printf("   %d\n", minWork)
	}

	fmt.Println()
	fmt.Println("the same table on the TYPICAL frame (mean over 10s), which is the number")
	fmt.Println("s28-askroom prints and the one the brief quotes:")
	fmt.Printf("%-6s", "level")
	for _, g := range geoms {
		fmt.Printf("%9s", fmt.Sprintf("%dx%d", g.w, g.h))
	}
	fmt.Println()
	for lv := 0.0; lv <= 0.701; lv += 0.05 {
		fmt.Printf("%-6.2f", lv)
		for _, g := range geoms {
			fmt.Printf("%9d", beachTyp(g.w, g.h, lv))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("named levels, worst frame / typical frame:")
	for _, n := range []struct {
		name string
		lv   float64
	}{
		{"p10 at the ask", 0.54},
		{"p50 at the ask (today)", 0.59},
		{"p90 at the ask", 0.65},
		{"TurnFloor only", 0.30},
		{"fully withdrawn", 0.00},
	} {
		fmt.Printf("  %-24s", n.name)
		for _, g := range geoms {
			w, t := beach2(g.w, g.h, n.lv)
			fmt.Printf("%9s", fmt.Sprintf("%d/%d", w, t))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("lowest level at which N rows exist in ALL FOUR of the geometries he works in")
	fmt.Println("(80x24, 111x25, 124x22, 143x27) -- 'never' means not even at level 0.00:")
	for n := 5; n <= 13; n++ {
		fmt.Printf("  %2d rows: %s\n", n, lowestFor(n, 4))
	}
	fmt.Println()
	fmt.Println("same, counting 153x51 as well:")
	for n := 5; n <= 13; n++ {
		fmt.Printf("  %2d rows: %s\n", n, lowestFor(n, 5))
	}

	fmt.Println()
	fmt.Println("PER GEOMETRY, the most beach the tide can ever buy (level 0.00, worst frame):")
	for _, g := range geoms {
		fmt.Printf("  %-8s %2d rows\n", fmt.Sprintf("%dx%d", g.w, g.h), beach(g.w, g.h, 0.0))
	}

	fmt.Println()
	fmt.Println("WHAT THE 16x11 CRAB ACTUALLY NEEDS, by anchor and by how much of it must be dry.")
	fmt.Println("The box is not the animal: rows 0-1 are the raised claw, 2-3 the eyes, 5-10 the shell.")
	fmt.Printf("%-46s %-8s %s\n", "requirement", "rows", "lowest level that buys it, all four")
	for _, r := range []struct {
		name string
		rows int
	}{
		{"under=0 (feet on the last row), shell dry", 6},
		{"under=0, eyes dry too", 9},
		{"under=0, whole box dry", 11},
		{"under=2 (today's rule), shell dry", 8},
		{"under=2, eyes dry too", 11},
		{"under=2, whole box dry", 13},
		{"-- and the 14x9 rung, for comparison --", 0},
		{"14x9 under=0, shell dry (shell at cell row 4)", 5},
		{"14x9 under=0, eyes dry too (eye at cell row 2)", 7},
		{"14x9 under=0, whole box dry", 9},
		{"14x9 under=2, whole box dry", 11},
	} {
		if r.rows == 0 {
			fmt.Println(r.name)
			continue
		}
		fmt.Printf("%-46s %-8d %s\n", r.name, r.rows, lowestFor(r.rows, 4))
	}
}

// lowestFor walks the level DOWN and reports the highest level (i.e. the least
// the tide has to withdraw) at which every geometry has n rows. Reported as
// "the level at the ask would have to be at most X".
func lowestFor(n, k int) string {
	for lv := 0.70; lv >= -0.001; lv -= 0.01 {
		ok := true
		for i := 0; i < k; i++ {
			if beach(geoms[i].w, geoms[i].h, lv) < n {
				ok = false
				break
			}
		}
		if ok {
			return fmt.Sprintf("level <= %.2f", lv)
		}
	}
	return "never -- not even with the water fully out"
}

// beach settles the shore at a fixed level and reports the WORST frame of the
// following ten seconds.
//
// SandTop is the MEAN waterline of one frame (shore.go:175) and it moves with
// the swell, so a single snapshot is a sample of an oscillation -- the first
// version of this table read one frame and reported 11 rows at 143x27 for level
// 0.10 and 10 for level 0.15, which is the swell phase and not the tide. The
// companion has to keep its head dry on EVERY frame, so the minimum is the
// number that decides it.
func beach(w, h int, level float64) int { w0, _ := beach2(w, h, level); return w0 }

func beachTyp(w, h int, level float64) int { _, t := beach2(w, h, level); return t }

func beach2(w, h int, level float64) (worst, typical int) {
	sh := scape.NewShore(7, false)
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	tm := 0.0
	for k := 0; k < 400; k++ { // 32s, an order of magnitude past TideEase
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	worst = 1 << 30
	sum := 0
	for k := 0; k < 125; k++ { // 10s more, keeping the worst frame and the mean
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
		b := (h - 1) - sh.SandTop() + 1
		if b < worst {
			worst = b
		}
		sum += b
	}
	return worst, sum / 125
}
