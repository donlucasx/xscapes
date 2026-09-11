package main

import "strings"

// The drafts.
//
// Authored as 12-column LEFT halves and mirrored into the 24-column row, so
// symmetry is a property of the builder rather than of my dot-counting. It is
// not taste: the sprite is flipped for the shipped right-hand composition, and
// my first hand-typed stalk pair was cols 6 and 15 -- 23-6 is 17, not 15 -- so
// one stalk stood under its eye and the other under a neighbouring cell. It
// rendered as a stray quadrant and it would have broken on the flip.
//
// mirror12 is checked against the shipped crabWork in main, which is the real
// proof that this builder produces the art it claims to.
func mirror12(left []string) []string {
	out := make([]string, len(left))
	for i, r := range left {
		if len(r) != 12 {
			panic("left half must be 12 columns: " + r)
		}
		b := []rune(r)
		rev := make([]rune, 12)
		for j := 0; j < 12; j++ {
			rev[j] = b[11-j]
		}
		out[i] = r + string(rev)
	}
	return out
}

// The parts, in left-half columns. Column 2c and 2c+1 are cell c, so cell 0 is
// cols 0-1, cell 4 is cols 8-9. Everything here is on the 2-pixel grid except
// the stalks, which are deliberately ONE pixel: a 1px column in a 2px cell is
// the only way this medium can draw a line thinner than a cell, and it is what
// makes a stalk exist at all.

// workLeft is the shipped working pose, rebuilt from halves.
var workLeft = []string{
	"##..##......",
	"##..##......",
	"######......",
	"######......",
	"######......",
	".####.......",
	"..##........",
	"..##........",
	"..##....##..",
	"..##....##..",
	"..##########",
	".###########",
}

// haltLeft: the stalk narrows from two pixels to one and the eye climbs a cell.
// Today's stalk is cols 8-9, which fills its cell completely and merges with
// the shell's rim -- so at parent scale the crab has never had a visible stalk,
// only an eye-shaped hole in the rim. One pixel leaves the other half of the
// cell to the background and the stalk appears.
var haltLeft = []string{
	"##..##......",
	"##..##......",
	"######......",
	"######......",
	"######......",
	".####.......",
	"..##........",
	"..##........",
	"..##....#...",
	"..##....#...",
	"..##########",
	".###########",
}

// wideLeft: the stalk steps one pixel outward, which moves the EYE CELL from 4
// to 3. The pair goes from three cells apart to five. Interocular distance
// growing is a looming cue, it costs nothing, and it leaves the balloon alone:
// (4+7+1)/2 and (3+8+1)/2 are both cell 6, so HeadCol does not move.
var wideLeft = []string{
	"##..##......",
	"##..##......",
	"######......",
	"######......",
	"######......",
	".####.......",
	"..##........",
	"..##........",
	"..##...#....",
	"..##...#....",
	"..##########",
	".###########",
}

// openLeft: the pincer gapes. The notch runs two source rows deeper, which
// takes the claw's middle cell from a half block to empty -- the background
// shows through the claw for the first time. Two extra rows of notch is the
// smallest change that alters a whole cell, because a cell is four source rows.
var openLeft = []string{
	"##..##......",
	"##..##......",
	"##..##......",
	"##..##......",
	"######......",
	".####.......",
	"..##........",
	"..##........",
	"..##...#....",
	"..##...#....",
	"..##########",
	".###########",
}

// askLeft: at the glass. The pincer stays open and the arm steps forward one
// cell, which opens a column of background between the elbow and the shell.
// That gap is the only foreshortening this medium has -- one flat colour cannot
// draw an arm in FRONT of a body, but it can draw the hole around it.
var askLeft = []string{
	"##..##......",
	"##..##......",
	"##..##......",
	"##..##......",
	"######......",
	".####.......",
	"..##........",
	"...##.......",
	"....##.#....",
	"....##.#....",
	"..##########",
	".###########",
}

var (
	genWork = mirror12(workLeft)
	halt    = mirror12(haltLeft)
	wide    = mirror12(wideLeft)
	open    = mirror12(openLeft)
	leanAsk = mirror12(askLeft)
)

// --- rejected: the claws brought round in FRONT of the shell -----------------
//
// The only true occlusion cue the medium offers, drawn as a gap carved out of
// the shell so the near arm reads over the far body. Kept for the record.
var frontClawLower = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	"..##..############..##..",
	"..##..############..##..",
	".##....##########....##.",
	".##....##########....##.",
	"##......########......##",
	"##......########......##",
	"##.......######.......##",
	".#.......######.......#.",
	".#........####........#.",
	".#....................#.",
	".#....................#.",
}

var frontClaw = mirror12([]string{
	"............",
	"............",
	"##..##......",
	"##..##......",
	"######......",
	".####.......",
	"..###.......",
	"...###......",
	"....####....",
	"......####..",
	"..##........",
	".###########",
})

// --- rejected: everything at once, to see the ceiling ------------------------
//
// Every cell in the upper half that CAN be filled, filled. Not a pose -- a
// measurement of how much of the picture direction C is allowed to move.
var maxFill = mirror12([]string{
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
	"############",
})

func literal(rows []string) string {
	var b strings.Builder
	for _, r := range rows {
		b.WriteString("\t\"" + r + "\",\n")
	}
	return b.String()
}

// --- two rival asks, because the box makes them fight over the same columns --
//
// bigClawLeft: the claw is the part of a crab NEAREST a viewer it is walking
// toward, so the honest perspective move is to grow the CLAW, not the animal.
// Four cells wide and two tall against today's three and two, gaping a whole
// cell. The cost is that the claws now occupy cells 0-3, so the eyes cannot
// step out to cell 3 -- big claws and wide eyes want the same twelve columns,
// and only one of them can have them.
var bigClawLeft = []string{
	"##....##....",
	"##....##....",
	"##....##....",
	"##....##....",
	"########....",
	"########....",
	".######.....",
	"..####......",
	"..##....#...",
	"..##....#...",
	"..##########",
	".###########",
}

// bulkLeft: no gape at all, everything solid, claws thickened and raised. The
// most ink direction C can put on the screen while still reading as a crab.
var bulkLeft = []string{
	"######......",
	"######......",
	"######......",
	"######......",
	"########....",
	"########....",
	"..######....",
	"..######....",
	"..##....#...",
	"..##....#...",
	"..##########",
	".###########",
}

var (
	bigClaw = mirror12(bigClawLeft)
	bulk    = mirror12(bulkLeft)
)

// openNarrowLeft: the pincer gapes while the eyes are still at cells 4 and 7.
// It exists so each beat of the approach changes exactly ONE thing -- the eye
// climbs, then the claw opens, then the eyes step apart, then the glyph grows.
// A beat that changes two things at once cannot be read at 20 fps.
var openNarrowLeft = []string{
	"##..##......",
	"##..##......",
	"##..##......",
	"##..##......",
	"######......",
	".####.......",
	"..##........",
	"..##........",
	"..##....#...",
	"..##....#...",
	"..##########",
	".###########",
}

var openNarrow = mirror12(openNarrowLeft)
