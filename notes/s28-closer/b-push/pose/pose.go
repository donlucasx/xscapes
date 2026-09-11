// Package pose holds direction B's push-in art, so the renderer and the scene
// instrument draw the SAME bitmaps. Two copies of a sprite is two sprites.
package pose

// DIRECTION B, the real push-in: 16 cells wide by 11 tall, which is 32 x 44
// source pixels. 1.33x the shipped crab's width and 1.57x its height.
//
// The vertical split follows the shipped crab's: an UPPER half that carries the
// claws, the stalks and the whole expression, and a LOWER half (shell and legs)
// that never changes. Upper is 20 rows (5 cell rows), lower 24 (6), so the shell
// crown lands at cell row 5 of 11 -- 45%, against the shipped crab's 43%. The
// animal has to be recognisably the same one, only nearer, and the proportion
// of the box that is shell is the strongest cue for that.
//
// THREE MEASURED RULES came out of drawing this, and the first draft broke all
// three:
//
//  1. A claw must be CELL-ALIGNED and EVEN. The first mid pose gave its claws
//     seven source columns; seven is three and a half cells, so the outer edge
//     of every claw came out as a half block and the pincer read as gravel.
//     Legs are the exception -- a leg is a diagonal, and the half-cell steps
//     are what let it splay.
//  2. A claw must START on a multiple of four source rows. The first draft
//     dropped the low claw six rows, which is one and a half cell rows, and the
//     fingers rendered as "▄▖▗▄" -- four disconnected corner marks.
//  3. The ASK pose does not RAISE a claw, it DROPS one. Read off the shipped
//     bitmaps: crabWork and crabAsk have an identical right claw, and the ask's
//     left claw is four source rows lower. The asymmetry is the whole signal,
//     and it survives being copied at any size.

// BigAsk is the ask pose. The right claw stays where working left it and the
// left drops two cell rows -- proportionally the same gesture as the shipped
// one cell row out of seven, and twice the drop in absolute cells so it still
// reads at this size. Authored RIGHT, so the held claw points at the transcript
// once the sprite is mirrored.
//
//	cols  0- 7  left claw      cols  2- 5 its arm
//	cols 11-12  left stalk     cols 19-20 right stalk
//	cols 24-31  right claw     cols 26-29 its arm
var BigAsk = []string{
	"........................###..###", //  0 right fingers, open
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................###..###",
	"........................########", //  6 right palm
	"........................########",
	"###..###...##......##...########", //  8 left fingers, stalks begin
	"###..###...##......##...########",
	"###..###...##......##.....####..", // 10 right wrist and arm
	"###..###...##......##.....####..",
	"###..###...##......##.....####..",
	"###..###...##......##.....####..",
	"########...##......##.....####..", // 14 left palm
	"########...##......##.....####..",
	"########...##......##.....####..",
	"########...##......##.....####..",
	"..####.....##......##.....####..", // 18 left wrist and arm
	"..####.....##......##.....####..",
}

// BigWork is both claws at working height -- the pose the ask drops out of, and
// the pose the approach arrives in.
var BigWork = []string{
	"###..###................###..###",
	"###..###................###..###",
	"###..###................###..###",
	"###..###................###..###",
	"###..###................###..###",
	"###..###................###..###",
	"########................########",
	"########................########",
	"########...##......##...########",
	"########...##......##...########",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
	"..####.....##......##.....####..",
}

// BigLower is the shell and the legs, and it never changes with state.
//
// Row 2 is one column narrower on each side ON PURPOSE: it and row 3 are OR'd
// into the same half row, and if both are full width the shell's crown renders
// as sixteen cells of unbroken block. The shipped crab rounds its crown the
// same way, and that rounding is most of what stops a big shell reading as a
// wall.
var BigLower = []string{
	"################################",
	"################################",
	".##############################.",
	".##############################.",
	"..############################..",
	"..############################..",
	"...##########################...",
	"...##########################...",
	"..###...################...###..",
	"..###...################...###..",
	".###.....##############.....###.",
	".###.....##############.....###.",
	"###.......############.......###",
	"###.......############.......###",
	"###........##########........###",
	"###........##########........###",
	".##.........########.........##.",
	".##.........########.........##.",
	".##..........######..........##.",
	".##..........######..........##.",
	".#............####............#.",
	".#............####............#.",
	".#............................#.",
	".#............................#.",
}

// The eye, drawn as its own Sprite pass in the eye colour, because a glyph
// cannot be scaled and a one-cell eye on a sixteen-cell body is a pinprick.
//
// 4 px wide by 8 tall = 2 cells by 2, remembering that ToQuadrant halves
// vertically FIRST: rows 2k and 2k+1 are OR'd, so a feature has to be two rows
// tall to survive at all and four to own a cell row.
var BigEyeSolid = []string{
	".##.",
	".##.",
	"####",
	"####",
	"####",
	"####",
	".##.",
	".##.",
}

// The same eye hollowed, which is what the shipped 'O' actually is. The hole
// shows the SCENE, exactly as the shipped eyes do by default -- and that is the
// thing EyeFillSocket exists to fix, so this one is drawn to be rejected rather
// than to be shipped.
var BigEyeRing = []string{
	".##.",
	".##.",
	"#..#",
	"#..#",
	"#..#",
	"#..#",
	".##.",
	".##.",
}

// BigEyeCells / BigEyeRow: where the eye pass lands, in the authored box.
// Written down rather than derived, for the reason crab.go gives.
var (
	BigEyeCells = [2]int{5, 9} // left cell of each 2-cell-wide eye
	BigEyeRow   = 2            // top cell row of the 2-cell-tall eye
)

// The middle rung of the push-in, 14 cells by 9 = 28 x 36 px.
//
// It exists because a size change in this medium is not a scale transform. A
// bitmap has no in-between: 12x7 -> 16x11 is either a hand-drawn rung between
// them or a pop. This is that rung, and its cost -- one more upper, one more
// lower, one more eye, and a second set of eye cells -- is the honest price of
// the animation.
//
// Shell crown at cell row 4 of 9 = 44%, between the shipped 43% and the big
// pose's 45%, so the three sizes read as one animal.
//
// Claws are SIX source columns here, not seven: an odd width straddles a cell
// and renders the outer edge as a half block.

var MidAsk = []string{
	"......................##..##", //  0 right fingers
	"......................##..##",
	"......................######", //  2 right palm
	"......................######",
	"##..##................######", //  4 left fingers
	"##..##................######",
	"######................######", //  6 left palm
	"######................######",
	"######...##......##.....##..", //  8 right arm, stalks
	"######...##......##.....##..",
	"######...##......##.....##..",
	"######...##......##.....##..",
	"..##.....##......##.....##..", // 12 left arm
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
}

var MidWork = []string{
	"##..##................##..##",
	"##..##................##..##",
	"######................######",
	"######................######",
	"######................######",
	"######................######",
	"######................######",
	"######................######",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
	"..##.....##......##.....##..",
}

var MidLower = []string{
	"############################",
	"############################",
	".##########################.",
	".##########################.",
	"..########################..",
	"..########################..",
	"..##...##############...##..",
	"..##...##############...##..",
	".##.....############.....##.",
	".##.....############.....##.",
	"##.......##########.......##",
	"##.......##########.......##",
	".#........########........#.",
	".#........########........#.",
	".#.........######.........#.",
	".#.........######.........#.",
	".#..........####..........#.",
	".#..........####..........#.",
	".#........................#.",
	".#........................#.",
}

// One cell row tall, two wide: 4 px by 4. The rung between a single glyph and
// the big pose's 2x2 bitmap.
var MidEye = []string{
	".##.",
	".##.",
	"####",
	"####",
}

var (
	MidEyeCells = [2]int{4, 8}
	MidEyeRow   = 2
)

// The shipped crab, copied verbatim from internal/companion/crab.go so this
// draft can be compared against it side by side without importing unexported
// data. If these ever drift, the comparison is worthless -- they are checked
// against the real ones by shipCheck().
var ShipLower = []string{
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

var ShipAsk = []string{
	"..................##..##",
	"..................##..##",
	"..................######",
	"..................######",
	"##..##............######",
	"##..##.............####.",
	"######..............##..",
	".####...............##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

var ShipWork = []string{
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}
