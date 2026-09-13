package main

// The art HE RULED ON, 2026-09-11, with the edits his rulings called for.
// Derived from the drafting round in catalts_art.go; kept separate because that
// file is generated and this one is hand-edited.

// catRungFaithful is "The Faithful Scale-Up" with the NOSE REMOVED, his ruling:
// "try the faithcul scale-up, rung 1, but without the nose (the original rung 0
// doesnt have any nose/mouth)."
//
// ⚠ HE IS RIGHT AND THE REASON IS A RENDERING ACCIDENT, not a design decision.
// The shipped cat DOES carry a muzzle gap in its source -- CatBody row 11,
// columns 8..10 -- but ToQuadrant ORs source rows 2k and 2k+1 before packing,
// and rows 10 and 12 are solid, so row 11 is swallowed whole. The cat has never
// had a nose ON SCREEN. The scale-up put its own muzzle gap across rows 16..17,
// which is a half-row boundary, so the gap survived the OR and a nose appeared
// for the first time as the animal walked toward you. That is a new feature
// arriving with proximity, not more of the same detail at a bigger size.
//
// Removed here: the notch at rows 16..17 of the level pose, and the open mouth
// at rows 16..19 of the ask. The ask still reads, and it reads by the same
// mechanism he picked for the shipped size -- the skull band between the ears
// is deleted so the ears stand free (rows 2..3).
var catRungFaithfulUpper = []string{
	".....##..........##.............",
	".....##..........##.............",
	"....#####......#####............",
	"....################............",
	"...##################...........",
	"..####################..........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".###....########....###.........",
	".###....########....###.........",
	".###....########....###.........",
	".###....########....###.........",
	".######################.........",
	".######################.........",
	"...##################...........",
	"....################............",
	".....##############.............",
	"......############..............",
	".....##############.............",
	"....################............",
}

var catRungFaithfulAsk = []string{
	".....##..........##.............",
	".....##..........##.............",
	"....#####......#####............",
	"....######....######............",
	"...##################...........",
	"..####################..........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".###....########....###.........",
	".###....########....###.........",
	".###....########....###.........",
	".###....########....###.........",
	".######################.........",
	".######################.........",
	"...##################...........",
	"....################............",
	".....##############.............",
	"......############..............",
	".....##############.............",
	"....################............",
}

var catRungFaithfulLower = []string{
	"..####################..........",
	"..####################..........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	".######################.........",
	"..####...######...####..........",
	"..####...######...####..........",
	"..####...######...####..........",
}

// crabSettled is the crab's Done pose, his pick: "settled on the sand".
var crabSettledUpper = []string{
	"........................",
	"........................",
	"........................",
	"........................",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	"..####..##....##..####..",
	"..####..##....##..####..",
	"..##......####......##..",
	"..##......####......##..",
}

var crabSettledLower = []string{
	"..####################..",
	"..####################..",
	"########################",
	"########################",
	"########################",
	"########################",
	"########################",
	"########################",
	"##..################..##",
	"##..################..##",
	"########################",
	"########################",
	"######..########..######",
	"######..########..######",
	"..####..########..####..",
	"..####..########..####..",
}

// crabSettledClaw is the same pose with the NEAR claw shut.
//
// His condition on the pick, and it is a design instruction in its own right:
// "it should not be completely static. it could close a claw every so often, or
// something similar."
//
// ⚠ Every other held pose in this project is deliberately STILL -- pace.go's
// stillFor() freezes the companion for NeedsYou, Done and Worried, because a
// held pose that drifts reads as the animal being unsure. An idle twitch is
// allowed here for one reason and it should be stated: a claw that shuts every
// so often carries NO INFORMATION. It is not a rate encoding of anything,
// because nothing varies with it. It says only "this animal is alive", which is
// the one thing the finish pose was missing.
//
// One claw, not both: two pincers closing in unison is a machine. The pincer is
// literally "##..##" in the source -- a two-column gap -- so closing it is
// filling that gap, and nothing else about the pose moves.
var crabSettledClaw = []string{
	"........................",
	"........................",
	"........................",
	"........................",
	"######............##..##",
	"######............##..##",
	"######............######",
	"######............######",
	"..####..##....##..####..",
	"..####..##....##..####..",
	"..##......####......##..",
	"..##......####......##..",
}

// crabSettledBoth is the louder version: BOTH pincers shut.
//
// On the page beside the single claw, because one claw is a single CELL of
// change at the shipped 12x7 size and that may simply be too quiet to notice --
// which is the same complaint that started the constellation work. Two claws is
// twice the signal and reads more like a machine resetting than an animal
// settling. His call, from the two rendered.
var crabSettledBoth = []string{
	"........................",
	"........................",
	"........................",
	"........................",
	"######............######",
	"######............######",
	"######............######",
	"######............######",
	"..####..##....##..####..",
	"..####..##....##..####..",
	"..##......####......##..",
	"..##......####......##..",
}

// ---------------------------------------------------------------------------
// RUNG 2, FIXED. His ruling 2026-09-12: "lets fix the rung 2 then, just fix the
// eyes and remove the mouth."
// ---------------------------------------------------------------------------

// catBodyNoMouth is CatBody with the muzzle gap filled.
//
// ⭐ HE CAUGHT THE SAME TRAP FROM THE OTHER SIDE. The shipped cat carries a
// muzzle gap at row 11 that NEVER REACHES THE SCREEN: ToQuadrant ORs source
// rows 2k and 2k+1, row 10 is solid, and the gap is swallowed. Doubling puts
// that same gap on a clean cell boundary, so a mouth APPEARS at 24x14 that
// exists nowhere else in the product. Two days ago he removed a nose from the
// 16x9 for the same reason; this is the mirror image of it.
//
// ⚠ Filling row 11 is provably invisible at 12x7 -- the OR already hid it --
// which is why this is one source and not two. Asserted in the test suite.
var catBodyNoMouth = []string{
	"...##.........##........",
	"...###.......###........",
	"...####.....####........",
	"...#############........",
	"..###############.......",
	".#################......",
	".#################......",
	".#################......",
	".###...#####...###......",
	".###...#####...###......",
	".#################......",
	".#################......", // <- was ".#######...#######......", the muzzle
	".#################......",
	"..###############.......",
	"...#############........",
	"....###########.........",
	".....#########..........",
	"....###########.........",
	"...#############........",
	"..###############.......",
	"..###############.......",
	".#################......",
	".#################......",
	".#################......",
	".#################......",
	".#################......",
	"..###..#####..###.......",
	"..###..#####..###.......",
}

// catRung2EyeCells is where the doubled cat's eye sockets actually are, on the
// rendered frame: cells 4..6 and 12..14 of row 4, three cells wide each.
//
// ⚠ MY MOCKUP HAD THIS WRONG AND THAT IS MOST OF WHAT HE WAS LOOKING AT. It
// passed {8, 14} -- the CRAB's near-pose eye cells, copied out of nearPoseFor.
// Measured on the rendered frame: cell 8 is INK, and mirrored (which is how the
// scape draws it) {8,14} become 15 and 9, BOTH INK. So both eye glyphs were
// painted on the animal's forehead rather than in its sockets.
var (
	catRung2Sockets = [2][2]int{{4, 6}, {12, 14}}
	catRung2EyeRow  = 4
)

// catRung2DeepEye is the doubled body with the eye socket carried DOWN a whole
// cell row, so an eye can be two cells tall instead of one.
//
// This is the other half of "fix the eyes". At 12x7 the eye is 1 cell of 84; at
// 24x14 a single glyph is 1 of 336, which is why it reads as a small mark stuck
// on a large blank face rather than as an eye. The crab solved the same problem
// at the same rung by going to a 2x2 bitmap. Only the doubled art is touched,
// so nothing at 12x7 or 16x9 moves.
func catRung2DeepEye() []string {
	rows := double2x(catBodyNoMouth)
	// Source rows 16..19 are the existing socket (cell row 4). Carrying it to
	// rows 20..23 makes the socket cell rows 4 and 5.
	for y := 20; y <= 23; y++ {
		r := []rune(rows[y])
		for _, span := range [][2]int{{8, 13}, {24, 29}} {
			for x := span[0]; x <= span[1]; x++ {
				r[x] = '.'
			}
		}
		rows[y] = string(r)
	}
	return rows
}
