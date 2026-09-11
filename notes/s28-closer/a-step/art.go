package main

import (
	"fmt"
	"strings"
)

// span-built rows. Typing 28-character strings by hand is how a sprite ends up
// one pixel out of symmetry without anyone noticing; a row built from spans and
// then PRINTED as a literal is the same art with the error class removed.
type span [2]int // inclusive [a,b]

func row(w int, sp ...span) string {
	b := make([]byte, w)
	for i := range b {
		b[i] = '.'
	}
	for _, s := range sp {
		for x := s[0]; x <= s[1] && x < w; x++ {
			if x >= 0 {
				b[x] = '#'
			}
		}
	}
	return string(b)
}

// mir mirrors a span within width w, so every symmetric feature is authored
// once. The crab is a front-on animal and every pose but the raised claw is a
// palindrome; the art should make that structural rather than hoped for.
func mir(w int, s span) span { return span{w - 1 - s[1], w - 1 - s[0]} }

func both(w int, sp ...span) []span {
	out := append([]span{}, sp...)
	for _, s := range sp {
		out = append(out, mir(w, s))
	}
	return out
}

func literal(name string, rows []string) {
	fmt.Printf("var %s = []string{\n", name)
	for _, r := range rows {
		fmt.Printf("\t%q,\n", r)
	}
	fmt.Printf("}\n\n")
}

// ---------------------------------------------------------------------------
// NEAR: 28 x 36 px = 14 cells wide x 9 tall.
//
// Split 16 upper + 20 lower, both multiples of 4, so the join lands on a cell
// boundary and the eye row does not shift when a pose changes.
//
// Column plan (w=28, centre between 13 and 14):
//	claw block  0-7    and 20-27   (cells 0-3, 10-13)
//	claw arm    2-5    and 22-25   (cells 1-2, 11-12)
//	eyestalk    8-11   and 16-19   (cells 4-5, 8-9)
//	gap between stalks 12-15       (cells 6-7)
// ---------------------------------------------------------------------------

const nearW = 28

// armPx is set from the command line so the two candidate arm widths can be
// rendered side by side instead of argued about.
var armPx = 4

var (
	nearClawL = span{0, 7}
	nearArmL  = span{2, 5} // overridden by -arm; see armVariant
	// The stalk is authored in two pieces so the SILHOUETTE is continuous under
	// the eye: a 2 px neck where the eye's bulb is, flaring to 4 px where it
	// meets the shell. Without it the body has a hole at the eye, plotRim rings
	// nothing there, and the eye pass paints straight onto open sky.
	nearNeck  = []span{{9, 10}, {17, 18}}
	nearStalk = []span{{8, 11}, {16, 19}}
	// The shell's top two rows live in the UPPER half, which is where the
	// shipped crab keeps them: it is what puts the stalks and the shell edge in
	// ONE cell row and gives the face its texture. Move them down into the
	// lower half and that row goes bare, which is a change of character rather
	// than a change of size.
	nearShellTop = []span{{2, 25}, {1, 26}}
)

// nearAsk: the right claw is raised a full cell row above the left, exactly the
// lift the shipped pose uses. Authored right, which is screen LEFT once
// mirrored, so the raised claw points at the agent's transcript.
func nearAsk() []string {
	w := nearW
	cl, ar := nearClawL, nearArmL
	cr, arR := mir(w, cl), mir(w, ar)
	// prongs of a claw: two tips with a notch between them.
	prongs := func(c span) []span {
		return []span{{c[0], c[0] + 2}, {c[1] - 2, c[1]}}
	}
	pr, pl := prongs(cr), prongs(cl)
	taper := func(c span) span { return span{c[0] + 1, c[1] - 1} }
	nk, nkR := nearNeck[0], nearNeck[1]
	var rows []string
	add := func(sp ...span) { rows = append(rows, row(w, sp...)) }

	// The claw's own clock, both claws the same, the raised one four source
	// rows -- one cell row -- ahead. That is the shipped lift exactly.
	//   3 rows prongs | 2 knuckle | 1 taper | the rest arm
	// The timing matters more than it looks: it is what keeps the lowered
	// claw's bulk OUT of the eye's cell row, so the face has air around it.
	add(pr[0], pr[1])                        // 0
	add(pr[0], pr[1])                        // 1
	add(pr[0], pr[1])                        // 2
	add(cr)                                  // 3
	add(cr, pl[0], pl[1])                    // 4
	add(taper(cr), pl[0], pl[1])             // 5
	add(arR, pl[0], pl[1])                   // 6
	add(arR, cl)                             // 7
	add(arR, cl)                             // 8
	add(arR, taper(cl), nk, nkR)             // 9
	add(arR, ar, nk, nkR)                    // 10
	add(arR, ar, nk, nkR)                    // 11
	add(arR, ar, nearStalk[0], nearStalk[1]) // 12
	add(arR, ar, nearStalk[0], nearStalk[1]) // 13
	add(nearShellTop[0])                     // 14
	add(nearShellTop[1])                     // 15
	return rows
}

// nearWork is the same body with both claws up and no lift, so the near size is
// honest in every pose and not only at the ask. Not the deliverable, but the
// walk-back passes through it.
func nearWork() []string {
	w := nearW
	cl, ar := nearClawL, nearArmL
	cr, arR := mir(w, cl), mir(w, ar)
	prongs := func(c span) []span { return []span{{c[0], c[0] + 2}, {c[1] - 2, c[1]}} }
	pl, pr := prongs(cl), prongs(cr)
	taper := func(c span) span { return span{c[0] + 1, c[1] - 1} }
	nk, nkR := nearNeck[0], nearNeck[1]
	var rows []string
	add := func(sp ...span) { rows = append(rows, row(w, sp...)) }
	// Both claws on the same clock, neither lifted: 4 prongs, 2 knuckle,
	// 1 taper, 5 arm. One row longer than the ask pose's claw because nothing
	// is held above the frame.
	for i := 0; i < 4; i++ {
		add(pl[0], pl[1], pr[0], pr[1])
	}
	add(cl, cr)
	add(cl, cr)
	add(taper(cl), taper(cr))
	add(ar, arR)
	for i := 0; i < 4; i++ {
		add(ar, arR, nk, nkR)
	}
	add(ar, arR, nearStalk[0], nearStalk[1])
	add(ar, arR, nearStalk[0], nearStalk[1])
	add(nearShellTop[0])
	add(nearShellTop[1])
	return rows
}

// lowerHalf builds the shell and the legs for a given width. The MID and the
// NEAR share one schedule on purpose: the approach has to read as one animal
// getting nearer, and a leg that lands on a different rung at each size reads
// as a different animal instead.
//
// apron is the widest ink run per row as the shell tapers; leg is the outer
// pair, authored left and mirrored.
func lowerHalf(w int, phase int) []string {
	c := func(n int) span { return span{(w - n) / 2, (w-n)/2 + n - 1} }
	var rows []string
	for _, n := range []int{w, w, w - 2, w - 4, w - 6, w - 8} {
		rows = append(rows, row(w, c(n)))
	}
	apron := []int{14, 14, 12, 12, 10, 10, 8, 8, 6, 6, 4, 4, 0, 0}
	// Stand and mid-stride: the legs swap phase, the trick crabLowerStep plays.
	// The approach is a WALK, so a size step has to land on a stride.
	legs := [][]span{
		{{2, 3}, {2, 3}, {1, 2}, {1, 2}, {0, 1}, {0, 1}, {0, 1}, {0, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
		{{1, 2}, {1, 2}, {0, 1}, {0, 1}, {1, 2}, {1, 2}, {2, 3}, {2, 3}, {2, 2}, {2, 2}, {2, 2}, {2, 2}, {2, 2}, {2, 2}},
	}
	for i, n := range apron {
		body := c(n)
		if n == 0 {
			body = span{-1, -2}
		}
		l := legs[phase][i]
		rows = append(rows, row(w, body, l, mir(w, l)))
	}
	return rows
}

func nearLower() []string     { return lowerHalf(nearW, 0) }
func nearLowerStep() []string { return lowerHalf(nearW, 1) }
func midLower() []string      { return lowerHalf(midW, 0) }
func midLowerStep() []string  { return lowerHalf(midW, 1) }

// ---------------------------------------------------------------------------
// MID: 26 x 32 px = 13 cells wide x 8 tall. Upper 12, lower 20.
//
// The body grows FIRST and the head second. That is not a compromise with the
// medium, it is the only order the medium allows -- a size step is a whole cell
// row -- and it happens to read right: the animal gets more substantial, then
// rears up to ask.
//
// 13 cells is odd, so the mirror-symmetric one-cell eye pair is {4,8}, which is
// where the glyph eyes go. The eye does not change medium at this step.
// ---------------------------------------------------------------------------

const midW = 26

var (
	midClawL = span{0, 5}
	midArmL  = span{2, 5}
)

// midAsk is the SHIPPED upper half, two pixels wider. Nothing about its
// structure changes: same claw clock, same one-cell stalks, same two shell-top
// rows at the bottom. That is deliberate -- the first step of the approach
// should cost the eye as little as possible, and a step that only widens by a
// cell and lengthens the body by a row is about as little as this medium has.
func midAsk() []string {
	w := midW
	cl, ar := midClawL, midArmL
	cr, arR := mir(w, cl), mir(w, ar)
	prongs := func(c span) []span { return []span{{c[0], c[0] + 1}, {c[1] - 1, c[1]}} }
	pl, pr := prongs(cl), prongs(cr)
	taper := func(c span) span { return span{c[0] + 1, c[1] - 1} }
	st, stR := span{8, 9}, mir(w, span{8, 9})
	var rows []string
	add := func(sp ...span) { rows = append(rows, row(w, sp...)) }
	add(pr[0], pr[1])            // 0
	add(pr[0], pr[1])            // 1
	add(cr)                      // 2
	add(cr)                      // 3
	add(cr, pl[0], pl[1])        // 4
	add(taper(cr), pl[0], pl[1]) // 5
	add(arR, cl)                 // 6
	add(arR, taper(cl))          // 7
	add(arR, ar, st, stR)        // 8
	add(arR, ar, st, stR)        // 9
	add(span{2, 23})             // 10  shell top
	add(span{1, 24})             // 11  shell top
	return rows
}

// ---------------------------------------------------------------------------
// The eye, as a quadrant bitmap.
//
// A glyph cannot be scaled -- a cell is a cell -- so on a 14x9 body the shipped
// one-cell 'O' is a pinprick. The eye becomes its own Sprite pass in eyeAlert,
// drawn over the body exactly as the glyph is today.
//
// The rounding has to be authored in HALF-ROW space, not source-row space:
// ToQuadrant ORs source rows 2k and 2k+1 together before it packs, so a corner
// notched one source row deep is eaten before it is ever seen. Every feature
// here is therefore two source rows tall.
// ---------------------------------------------------------------------------

// nearEye is 4 px wide x 8 px tall = 2 cells x 2 cells.
var nearEye = []string{
	".##.",
	".##.",
	"####",
	"####",
	"####",
	"####",
	".##.",
	".##.",
}

// nearPupil sits inside it, one cell, offset toward the middle of the face. Two
// passes, two colours -- the same trick the sand uses to keep ink off the
// palette's nominal colour.
var nearPupil = []string{
	"....",
	"....",
	".##.",
	".##.",
	"....",
	"....",
	"....",
	"....",
}

// nearEyeFlat is the smaller alternative: 2 cells wide, 1 cell tall. Half the
// pop at the arrival, and half the face.
var nearEyeFlat = []string{
	".##.",
	".##.",
	"####",
	"####",
}

func art() {
	rule("DIRECTION A -- the art, rendered")

	up, lo := nearAsk(), nearLower()
	checkWidths("nearAsk", up)
	checkWidths("nearLower", lo)
	near := join(up, lo)
	if b, f := symmetry(lo); b != 0 {
		panic(fmt.Sprintf("nearLower is asymmetric at row %d", f))
	}
	q := show("NEAR ask pose", near)
	showMirrored("NEAR ask pose", near)
	fmt.Printf("cell rows: %d, cells wide: %d\n\n", len(q), len([]rune(q[0])))

	mu, ml := midAsk(), midLower()
	checkWidths("midAsk", mu)
	checkWidths("midLower", ml)
	mid := join(mu, ml)
	if b, f := symmetry(ml); b != 0 {
		panic(fmt.Sprintf("midLower is asymmetric at row %d", f))
	}
	show("MID ask pose", mid)
	showMirrored("MID ask pose", mid)

	show("NEAR working pose (the walk-back passes through it)", join(nearWork(), nearLower()))
	show("NEAR mid-stride lower half", join(nearAsk(), nearLowerStep()))

	rule("THE THREE SIZES, top-aligned -- which is how the approach shows them")
	sideBySide(
		[]string{"far 12x7 (shipped)", "mid 13x8", "near 14x9"},
		[][]string{join(farAsk, farLower), mid, near})
	fmt.Println("  the top row of ink never moves. The feet advance:")
	fmt.Println("  far feet H-3, mid feet H-2, near feet H-1 -- into the two rows")
	fmt.Println("  that already sit empty under the sprite.")

	rule("THE EYE as its own quadrant bitmap")
	show("nearEye  (its own Sprite pass, in eyeAlert)", nearEye)
	show("nearPupil (second pass, one shade down)", nearPupil)
	show("nearEyeFlat (the 2x1 alternative, for comparison)", nearEyeFlat)
	fmt.Println("composited onto the near body at cells {4,5} and {8,9}, cell rows 2-3:")
	composite(near, []int{4, 8}, 2, true)
	fmt.Println("and MIRRORED, which is the shipped composition. The eye cells {4,5} and")
	fmt.Println("{8,9} map onto each other under the flip (13-4=9, 13-5=8), so the face")
	fmt.Println("needs no second set of coordinates and nearEye needs no mirrored twin:")
	mirNear := make([]string, len(near))
	for i, r := range near {
		b := []rune(r)
		for j, k := 0, len(b)-1; j < k; j, k = j+1, k-1 {
			b[j], b[k] = b[k], b[j]
		}
		mirNear[i] = string(b)
	}
	composite(mirNear, []int{4, 8}, 2, true)

	rule("LITERAL ROWS, ready to paste into internal/companion/crab.go")
	literal("crabNearAsk", up)
	literal("crabNearWork", nearWork())
	literal("crabNearLower", lo)
	literal("crabNearLowerStep", nearLowerStep())
	literal("crabMidAsk", mu)
	literal("crabMidLower", ml)
	literal("crabNearEye", nearEye)
	literal("crabNearPupil", nearPupil)
}

// composite overlays the eye passes at SOURCE PIXEL resolution, which is the
// only honest place to do it: a cell holds one colour per quadrant, so marking
// whole cells makes a 2px pupil look like a 2-cell one. The three passes are
// drawn as three planes and printed a pixel to a character.
//
// Cell (cx, cy) covers source columns 2cx..2cx+1 and source rows 4cy..4cy+3.
func composite(bodyRows []string, eyeCells []int, eyeRow int, pupil bool) {
	w := len([]rune(bodyRows[0]))
	grid := make([][]rune, len(bodyRows))
	for y, r := range bodyRows {
		grid[y] = []rune(r)
	}
	plot := func(art []string, cx, cy int, mark rune) {
		for dy, r := range art {
			for dx, ch := range []rune(r) {
				if ch != '#' {
					continue
				}
				y, x := cy*4+dy, cx*2+dx
				if y >= 0 && y < len(grid) && x >= 0 && x < w {
					grid[y][x] = mark
				}
			}
		}
	}
	for _, c := range eyeCells {
		plot(nearEye, c, eyeRow, 'O')
		if pupil {
			plot(nearPupil, c, eyeRow, 'x')
		}
	}
	fmt.Println("  source pixels, one character each -- # coat, O eye, x pupil:")
	for _, r := range grid {
		fmt.Printf("    %s\n", strings.ReplaceAll(string(r), ".", " "))
	}
	// And the silhouette the terminal actually draws, with the eye's ink folded
	// in so the OUTLINE including its overhang can be checked.
	merged := make([]string, len(grid))
	for y, r := range grid {
		merged[y] = strings.Map(func(c rune) rune {
			if c == '.' {
				return '.'
			}
			return '#'
		}, string(r))
	}
	fmt.Println()
	show("  body + eye, quadrant silhouette", merged)
}

// sideBySide prints candidates top-aligned, because the approach pins the TOP
// row and grows downward -- so top-aligned is how they are actually seen.
func sideBySide(titles []string, arts [][]string) {
	var cols [][]string
	wide := 0
	for _, a := range arts {
		q := quad(a)
		cols = append(cols, q)
		if n := len([]rune(q[0])); n > wide {
			wide = n
		}
	}
	tall := 0
	for _, c := range cols {
		if len(c) > tall {
			tall = len(c)
		}
	}
	var head strings.Builder
	for i, t := range titles {
		head.WriteString(fmt.Sprintf("%-*s", wide+4, t))
		_ = i
	}
	fmt.Println("  " + head.String())
	for y := 0; y < tall; y++ {
		var sb strings.Builder
		for _, c := range cols {
			cell := ""
			if y < len(c) {
				cell = c[y]
			}
			sb.WriteString(fmt.Sprintf("%-*s", wide+4, "|"+cell+"|"))
		}
		fmt.Println("  " + sb.String())
	}
	fmt.Println()
}
