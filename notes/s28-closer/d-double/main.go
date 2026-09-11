// Command double is DIRECTION D of the s28 "come closer" study: the exact 2x.
//
// Take the shipped 24x28 source, blow every pixel up to 2x2, render it through
// the same ToQuadrant the product uses, and print it. No re-authoring and no
// judgement calls -- the reference point the other three directions get
// measured against.
//
//	go run ./notes/s28-closer/d-double
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
)

// The eye at 2x sits at cell rows 3-4, cell columns 8-9 and 14-15. Derived,
// not chosen: at 2x one cell column IS one original source column, so the
// shipped eye cells {4, 7} -- source columns 8-9 and 14-15 -- keep their
// columns. The ROW moves up one, because at 2x the stalk tip and the shell top
// stop sharing a cell and the eye can finally sit on the stalk instead of on
// the shell. Both placements are printed in section 5.
var (
	eyeCols2x = [2]int{8, 14}
	eyeRow2x  = 3
)

func rule(s string) {
	n := 74 - len(s)
	if n < 1 {
		n = 1
	}
	fmt.Printf("\n%s %s\n", s, strings.Repeat("-", n))
}

func show(label string, q []string) {
	fmt.Printf("\n  %s  (%d cells wide x %d tall)\n", label, len([]rune(q[0])), len(q))
	for _, r := range q {
		fmt.Printf("    |%s|\n", r)
	}
}

// dump prints the raw source rows of the 2x art, for pasting into a report or
// into crab.go. `go run ./notes/s28-closer/d-double -dump`.
func dump() {
	fmt.Println("== upper: crabAsk at 2x (48 x 24)")
	for _, r := range scale(crabAsk, 2, 2) {
		fmt.Println(r)
	}
	fmt.Println("== lower: crabLower at 2x (48 x 32)")
	for _, r := range scale(crabLower, 2, 2) {
		fmt.Println(r)
	}
	fmt.Println("== frame 1 upper: crabAsk at 1x (24 x 12)")
	for _, r := range crabAsk {
		fmt.Println(r)
	}
	fmt.Println("== eyeAlert (4 x 8)")
	for _, r := range eyeAlert {
		fmt.Println(r)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-dump" {
		dump()
		return
	}
	work1 := crabFrame(crabWork)
	ask1 := crabFrame(crabAsk)
	work2 := scale(work1, 2, 2)
	ask2 := scale(ask1, 2, 2)

	rule("1. THE DOUBLER")
	fmt.Println(`
  scale(rows, 2, 2): every source pixel becomes 2x2 source pixels, then the
  production ToQuadrant. Source 24x28 -> 48x56, which is 24 cells x 14.`)
	q1, q2 := render(work1), render(work2)
	fmt.Println("\n  SHIPPED 12x7                EXACT 2x, 24x14")
	for _, r := range sideBySide(6, q1, q2) {
		fmt.Printf("    %s\n", r)
	}

	rule("2. IS AN EXACT 2x OF THE SOURCE AN EXACT 2x OF THE PICTURE? NO -- IT IS BETTER")
	fmt.Println(`
  ToQuadrant halves the source VERTICALLY first: rows 2k and 2k+1 are OR'd into
  one subpixel row, and only then are 2x2 packed into a cell. So the shipped
  render already throws away half the vertical detail of the art. A doubled
  source does not have to make that merge -- both rows of every pair came from
  the same original row, so the OR is lossless. Count the pairs that today's
  render is actually merging:`)
	for _, c := range []struct {
		name string
		rows []string
	}{{"working", work1}, {"ask", ask1}, {"worried", crabFrame(crabWorried)}} {
		same, total := pairedRows(c.rows)
		fmt.Printf("    %-8s %2d of %2d row-pairs identical -> %d pairs are MERGED on screen today\n",
			c.name, same, total, total-same)
	}
	naive := cellDouble(q1)
	fmt.Println(`
  Left: the picture on screen today, blown up cell for cell -- "just scale the
  picture", which is what a person means by 2x. Right: the exact 2x of the
  SOURCE. Same 24x14 footprint, so the difference is the rendering and nothing
  else.`)
	fmt.Println()
	for _, r := range sideBySide(6, naive, q2) {
		fmt.Printf("    %s\n", r)
	}
	d, n := diff(naive, q2)
	fmt.Printf("\n    %d of %d cells differ (%.1f%%), and every one of them is the 2x showing a\n"+
		"    source row the shipped render had merged away. The shell's rim goes from a\n"+
		"    2-cell step to a 1-cell step because of it.\n", d, n, 100*float64(d)/float64(n))

	rule("3. THE STAIR-STEP, MEASURED")
	fmt.Println(`
  Doubling horizontally puts BOTH subpixels of a cell in the same original
  source column, so TL always equals TR and BL always equals BR. Only indices
  0, 3, 12 and 15 are reachable -- space, upper half, lower half, full block.
  The quarter blocks and the left/right halves become unusable.`)
	fmt.Printf("\n    shipped 12x7  : %s\n", histLine(q1))
	fmt.Printf("    exact 2x      : %s\n", histLine(q2))
	fmt.Printf("    naive cell 2x : %s\n", histLine(naive))
	fmt.Println(`
  ⚠ The tempting conclusion -- "2x halves the horizontal resolution" -- is
  WRONG, and worth writing down because it is the first thing I believed. As a
  fraction of the ANIMAL nothing changes: half a cell of a 12-cell body and one
  whole cell of a 24-cell body are both one twenty-fourth of its width. The
  horizontal detail is identical; the VERTICAL doubles, from one fourteenth of
  the body's height to one twenty-eighth.

  What does change is absolute size, and that is the honest complaint: the same
  30 edge cells now sit in a body four times the area, so each staircase step is
  physically twice as tall and twice as wide on his screen. The stair-stepping
  he would see is not new -- it is the staircase already in the art, magnified,
  minus the half-cell dither that was hiding it. Which is what walking toward
  something in a terminal has to look like.`)

	rule("4. SCALE2X REPAIRS MOST OF IT, FOR FREE")
	fmt.Println(`
  There is a better doubler and it costs one function. EPX/Scale2x fills a
  corner from a neighbour when the two orthogonal neighbours agree and the
  diagonal disagrees. On 1-bit art it rounds outer corners and fills inner
  ones, which is exactly the sub-cell horizontal detail the integer double
  cannot reach.`)
	eq := render(epx(work1))
	fmt.Println("\n  EXACT 2x                        SCALE2X")
	for _, r := range sideBySide(8, q2, eq) {
		fmt.Printf("    %s\n", r)
	}
	fmt.Printf("\n    exact 2x  : %s\n", histLine(q2))
	fmt.Printf("    scale2x   : %s\n", histLine(eq))
	de, ne := diff(q2, eq)
	fmt.Printf("    %d of %d cells differ (%.1f%%).\n", de, ne, 100*float64(de)/float64(ne))
	fmt.Printf("    symmetric under the production mirror: exact 2x %v, scale2x %v\n",
		symmetric(work2), symmetric(epx(work1)))
	fmt.Println(`
  Verdict on this one, plainly: scale2x is rounder and it is still exact where
  the art is straight, but it invents ink the author did not draw -- look at the
  claw tips. For a REFERENCE POINT that is the wrong trade, so the rest of this
  study stays on the exact double. If 2x is ever built, start from scale2x and
  hand-correct the claws.`)

	rule("5. THE EYE. THIS IS THE PART TO LOOK AT.")
	fmt.Println(`
  The eyes are not in the bitmap. They are single-cell GLYPHS plotted on top at
  crabEyeCells {4, 7}, crabEyeRow 2 (crab.go:259, 342-366). A glyph cannot be
  scaled: a cell is a cell. Double the body and the eye stays exactly the size
  it was.

  At 2x one cell column IS one original source column, so the eye keeps columns
  8-9 and 14-15. Three renders, same footprint, so the comparison is on the page:`)

	glyphEye := overlay(overlay(q2, 8, 4, []string{"o"}), 14, 4, []string{"o"})
	bmLow := overlay(overlay(q2, 8, 4, render(eyeOpen)), 14, 4, render(eyeOpen))
	bmHigh := overlay(overlay(q2, 8, 3, render(eyeOpen)), 14, 3, render(eyeOpen))
	fmt.Println("\n   A: 2x body, today's 'o'      B: 2x2 eye at row 4        C: 2x2 eye at row 3")
	for _, r := range sideBySide(6, glyphEye, bmLow, bmHigh) {
		fmt.Printf("    %s\n", r)
	}
	fmt.Println(`
  Same three with the eye pass marked E, because a transcript has no colour and
  an eye you cannot find in the evidence is not a measurement:`)
	fmt.Println()
	for _, r := range sideBySide(6,
		inkMap(q2, [][3]interface{}{{8, 4, []string{"#"}}, {14, 4, []string{"#"}}}),
		inkMap(q2, [][3]interface{}{{8, 4, eyeOpen}, {14, 4, eyeOpen}}),
		inkMap(q2, [][3]interface{}{{8, 3, eyeOpen}, {14, 3, eyeOpen}})) {
		fmt.Printf("    %s\n", r)
	}

	hLow, cLow := holePunch(work2, eyeOpen, 8, 4)
	hHigh, cHigh := holePunch(work2, eyeOpen, 8, 3)
	fmt.Printf(`
  Row 4 puts the eye half on the stalk tip and half on the SHELL, which is
  where the shipped 1x eye sits only because at 1x those two share one cell.
  Row 3 caps the stalk and leaves the shell alone, which is what a stalked eye
  is. Row 3 also costs less: the eye pass REPLACES whole cells, so body ink the
  eye does not cover shows the sea through the crab.
      eye at row 4 : %d body subpixels lost under %d eye cells
      eye at row 3 : %d body subpixels lost under %d eye cells
  Zero at row 3, so the bitmap eye needs no PlotOn and no eyeGround fill -- its
  lower cell row is solid and lands exactly on the two solid stalk cells. That
  is why every bitmap in eyes.go has a solid base; it is not a style.

`, hLow, cLow, hHigh, cHigh)

	fmt.Println("  For scale, the shipped crab at the size it actually runs at, eyes in place:")
	ship := overlay(overlay(q1, 4, 2, []string{"o"}), 7, 2, []string{"o"})
	show("SHIPPED 12x7", ship)

	fmt.Printf(`
  The arithmetic, so nobody re-derives it. Eye ink as a fraction of the face:
    shipped    1 cell of a 12x7 box  = 1 in 84   = 1.19%%
    2x, glyph  1 cell of a 24x14 box = 1 in 336  = 0.30%%   (a QUARTER of today)
    2x, bitmap 4 cells of a 24x14    = 4 in 336  = 1.19%%   (today's fraction, exactly)
  A 2 cell by 2 cell eye is not a preference. It is the only size that holds the
  ratio the face reads at now.

`)

	fmt.Println("  The five states as 4x8 bitmaps, rendered (2 cells x 2 each):")
	for _, e := range []struct {
		name string
		art  []string
	}{
		{"open    working / resting", eyeOpen},
		{"alert   NeedsYou, amber", eyeAlert},
		{"shut    blink, resting", eyeShut},
		{"done    the content ^ ^", eyeDone},
		{"worried bead gone, tip only", eyeWorried},
	} {
		q := render(e.art)
		h, _ := holePunch(work2, e.art, 8, 3)
		fmt.Printf("    %-30s %s   holes: %d\n    %-30s %s\n", e.name, q[0], h, "", q[1])
	}
	fmt.Println(`
  ⚠ ONE THING THAT DOES NOT SURVIVE THE BLOW-UP, and it is the ask state. At 1x
  the NeedsYou eye is 'O' where working is 'o' -- the SAME CELL, so the alert
  already reads by colour alone (amber) and by the raised claw, never by size.
  At 2x with a bitmap, alert can finally be BIGGER than open: 4 solid cells
  against a bead with its corners knocked off. That is a channel this feature
  gets for free and the shipped crab has never had.`)

	fmt.Println(`
  And the frame the whole feature exists to produce -- the ask pose at 2x,
  wearing the alert eye:`)
	askEyes := overlay(overlay(render(ask2), eyeCols2x[0], eyeRow2x, render(eyeAlert)),
		eyeCols2x[1], eyeRow2x, render(eyeAlert))
	show("2x ASK, alert eyes", askEyes)
	show("the same, eye pass marked E",
		inkMap(render(ask2), [][3]interface{}{{8, 3, eyeAlert}, {14, 3, eyeAlert}}))
	fmt.Printf("    the alert eye adds a whole cell ROW of amber above the stalks that the\n" +
		"    working eye does not have, so ask/working is now a shape difference and\n" +
		"    not only a colour one. At 1x it was 'O' against 'o' in one cell.\n")
	fmt.Println("  Mirrored, which is how it actually ships (companion on the RIGHT):")
	askM := overlay(overlay(render(mirrored(ask2)), 8, 3, render(eyeAlert)), 14, 3, render(eyeAlert))
	show("2x ASK, mirrored", askM)
	fmt.Printf("    eye columns {8,9,14,15} in a 24-wide box mirror to {15,14,9,8}: the same\n"+
		"    set, so the eye needs no mirror arithmetic. Body symmetric: %v.\n",
		symmetric(scale(crabLower, 2, 2)))

	rule("6. DOES IT FIT? AT ONE OF HIS SIX GEOMETRIES, AND IT IS THE TALLEST")
	fmt.Println(`
  The companion is drawn at top = H-2-chh, so it always keeps two rows under
  its feet and its head is at H-2-chh. It is dry when that row is at or below
  the waterline, i.e. when the beach is at least chh+2 rows deep. Beach depth
  measured on a real settled Shore at the level his asks fire at (p50 0.59,
  from notes/s28-askroom), tide on, which is the default since 31264bc.`)
	fmt.Printf("\n  %-9s %-27s %s\n", "", "  beach rows (h - sandTop)", "rows of sprite in the sea @p50")
	fmt.Printf("  %-9s %6s %6s %6s %6s   %11s %12s\n",
		"geom", "rest", "p10", "p50", "p90", "1x (7 tall)", "2x (14 tall)")
	geoms := []struct{ w, h int }{
		{40, 12}, {80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}, {143, 62},
	}
	for _, g := range geoms {
		b0 := beach(g.w, g.h, 0.0)
		b10, b50, b90 := beach(g.w, g.h, 0.54), beach(g.w, g.h, 0.59), beach(g.w, g.h, 0.65)
		fmt.Printf("  %-9s %6d %6d %6d %6d   %11d %12d\n",
			fmt.Sprintf("%dx%d", g.w, g.h), b0, b10, b50, b90, wet(b50, 7), wet(b50, 14))
	}
	fmt.Println(`
  Read the last two columns, and read them against each other rather than
  against zero. The SHIPPED crab already stands with one or two rows of claw in
  the water at the median ask; that is the tolerance he lives with and has never
  reported. At his five short windows a 2x crab is EIGHT or NINE rows under at
  the same moment, which is not a tolerance, it is a crab in the sea. At 143x62
  it is two -- the shipped crab's own number at 124x22.

  ⚠ 143x62 is the one to check, because it is the geometry the brief calls out
  as having 16 rows. It does -- AT REST. At the level his asks actually fire at
  it has 14, and a fully dry 14-row crab needs 16. That is the trap: measure the
  beach at the moment the feature fires, not while the agent is idle.

  But 14 is not a failure, it is his own tolerance. At 143x62/p50 a 2x crab
  stands with TWO rows of claw in the water -- which is exactly what the SHIPPED
  crab does at 124x22 today, and he has never reported it. So the honest answer
  to "when does a 2x crab stop being absurd" has two lines, not one:`)

	fmt.Printf("\n    %-10s %-32s %s\n", "width", "livable (2 rows wet, as today)", "fully dry")
	for _, w := range []int{80, 111, 124, 143, 153} {
		fmt.Printf("    %-10s %-32s %s\n", itoa(w)+" cols",
			firstFit(w, 0.59, 14)+" rows tall @p50", firstFit(w, 0.59, 16)+" rows tall @p50")
	}
	fmt.Println(`
  60 rows livable (65 at 80 columns), 70 fully dry. His tallest real window is
  143x62 and it clears the livable line by two rows; every other geometry -- 80x24,
  111x25, 124x22, 143x27, 153x51 -- is between 2 and 9 rows short. The median
  is 25 rows. A 2x crab is a full-height-terminal feature and he runs one
  window out of six that qualifies.`)

	rule("7. THE WIDTH COST, WHICH IS THE HALF NOBODY COUNTS")
	fmt.Println(`
  compose() puts the companion at catX = w - catW - (2 + w/32) and hands the
  sand everything to its left minus the pacing span. Doubling catW takes twelve
  columns straight out of the writing band. The brief's own envelope is 16
  cells wide; 2x is 24.`)
	fmt.Printf("\n  %-10s %12s %12s %9s %14s\n", "geom", "sand @ 12", "sand @ 24", "lost", "crab % of w")
	for _, w := range []int{40, 80, 111, 124, 143, 153} {
		a, b := sandCols(w, 12), sandCols(w, 24)
		fmt.Printf("  %-10s %12d %12d %9d %13.0f%%\n",
			fmt.Sprintf("%d cols", w), a, b, a-b, 100*24.0/float64(w))
	}
	fmt.Println(`
  At 40 columns the crab is 60% of the frame and the sand is gone. At 80, the
  design target, the band drops 56 -> 44 and the tail starts dropping pieces.`)

	rule("8. THE APPROACH: THE MEDIUM HAS EXACTLY TWO RUNGS")
	fmt.Println(`
  His ask is a WALK -- "walk up closer and get bigger". An exact double cannot
  walk, because there is nothing between 1x and 2x that is also exact. Half a
  cell is the smallest step this medium has for POSITION (two source rows, which
  is what breathing already uses); for SIZE the step is a whole cell, and an
  integer scale only lands on 12x7 and 24x14.

  Can the ladder be generated instead of drawn? Nearest-neighbour to an
  arbitrary size, sampled from the pixel centre so the map is symmetric:`)
	for _, s := range []struct{ w, h, cw, ch int }{{32, 36, 16, 9}, {36, 42, 18, 11}, {40, 48, 20, 12}} {
		r := resample(work1, s.w, s.h)
		fmt.Printf("\n  resample to %dx%d source = %d cells x %d, symmetric: %v\n",
			s.w, s.h, s.cw, s.ch, symmetric(r))
		for _, line := range render(r) {
			fmt.Printf("    |%s|\n", line)
		}
	}
	fmt.Println(`
  16x9 is the biggest that fits the brief's envelope and it survives: the shell,
  the four legs, the two claws and both eye stalks are all still there, and it
  is symmetric. 18x11 comes out ASYMMETRIC -- an odd row count off a 28-row
  source lands the sampler on different rows left and right -- and it merges the
  eye stalks into the shell rim, which takes the face away. 20x12 is symmetric
  again. So a generated ladder works, but only at sizes whose row count divides
  the source cleanly, and 16x9 is the one that both fits and survives.

  ⚠ But note what the repo already decided about this, before I did: Crablet
  (12x16), CrabletSmall (10x12) and CrabletTiny (8x12) are HAND-AUTHORED and
  none is a scale of another -- crab.go says so in its own comment, "a one-pixel
  notch still survives the halving, which is why a crablet keeps an open
  pincer". The litter's ladder was drawn by hand because scaling did not hold
  the features that make it a crab. The same will be true of the parent's.

  So Direction D's approach animation is two frames and it is a CUT:`)
	for i, f := range [][]string{crabAsk, scale(crabAsk, 2, 2)} {
		show(fmt.Sprintf("approach frame %d (upper half only)", i+1), render(f))
	}
	fmt.Println(`  Which does not answer his ask. It pops; it does not walk.`)

	rule("9. WHAT I DREW AND REJECTED")
	fmt.Println(`
  2x VERTICAL ONLY (24x56 -> 12 cells x 14). Costs the sand nothing and doubles
  the vertical fidelity for free:`)
	show("2x vertical only", render(scale(work1, 1, 2)))
	fmt.Println(`  Rejected: a crab is a WIDE animal -- that is the whole reason he picked it
  over the cat -- and stretching it up turns the shell into a tower. It reads
  as a different animal, not a nearer one.`)
	fmt.Println(`
  2x HORIZONTAL ONLY (48x28 -> 24 cells x 7). Fits every one of his windows on
  height, since it is still seven rows:`)
	show("2x horizontal only", render(scale(work1, 2, 1)))
	fmt.Println(`  Rejected: it does not read as closer, it reads as squashed -- and it pays the
  full twelve-column sand bill anyway.`)
	fmt.Println(`
  NAIVE CELL DOUBLE (blow the finished render up cell for cell). Printed in
  section 2. Rejected on the histogram: one glyph, the full block. It is the
  only one of these that genuinely stair-steps, and it is what "just scale the
  picture" gives you.`)
}

// beach is the number of rows from the waterline to the bottom of the frame:
// everything the companion has to stand on.
func beach(w, h int, level float64) int { return h - sandTopAt(w, h, level) }

// wet is how many rows of a chh-tall companion are above the waterline, given
// that it is drawn at top = H-2-chh.
func wet(beachRows, chh int) int {
	if n := chh + 2 - beachRows; n > 0 {
		if n > chh {
			return chh
		}
		return n
	}
	return 0
}

func firstFit(w int, level float64, need int) string {
	for h := 12; h <= 120; h++ {
		if beach(w, h, level) >= need {
			return itoa(h)
		}
	}
	return ">120"
}

// sandCols is the writing band's width, straight out of compose().
func sandCols(w, catW int) int {
	right := 2 + w/32
	span := w / 16
	if span > 6 {
		span = 6
	}
	if span < 1 {
		span = 1
	}
	catX := w - catW - right
	if catX < 0 {
		catX = 0
	}
	if n := (catX - 1 - span) - 2; n > 0 {
		return n
	}
	return 0
}

// sandTopAt settles a real Shore at a level and reads its mean waterline. The
// same helper as notes/s28-closer, so the two instruments cannot disagree.
func sandTopAt(w, h int, level float64) int {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	return sh.SandTop()
}
