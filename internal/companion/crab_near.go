package companion

import (
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/envx"
	"github.com/donlucasx/xscapes/internal/term"
)

// HIS IDEA, 2026-09-10: when the companion needs input it should "come CLOSER
// to the user before it prompts it ... walk up closer to the screen and get
// bigger", and then, on the four drafted directions: "lets test the bigger crab
// -- it should walk up closer to prompt, without overlapping sub agents. OK if
// a part of its body is cut on the bottom of the screen."
//
// THAT LAST SENTENCE IS THE WHOLE DESIGN. Four directions were drafted and
// blind-checked under the assumption that the sprite must fit ENTIRELY on the
// beach, and that assumption killed the two big ones -- the beach is 8 rows at
// 80 columns, 7 at 124, and the tide buys none of them back (measured: activity
// at the ask is p50 0.59 and the beach row count is the same at both ends of
// the range). With the bottom allowed to clip, the ceiling stops being the
// sprite's HEIGHT and becomes the sprite's TOP ROW: the top row must stay out
// of the sea, and everything below it may run off the frame.
//
// So the near crab is anchored by its TOP ROW at exactly the y the shipped crab
// is drawn at, and grows DOWNWARD and outward. Three things fall out of that,
// and they are why it is built this way rather than by moving the animal down
// the beach:
//
//   - The top row can never rise into the sea, because it does not move at all.
//     That is the one guarantee this feature has to hold and it is now
//     structural rather than arithmetic.
//   - The eyes stay within a row or two of where they have always been, so the
//     balloon -- which hangs off `top`, not off the sprite -- needs no change.
//   - It is also what an approach LOOKS like from a camera near the ground: the
//     head holds its line and the feet come toward you, off the bottom of the
//     frame.
//
// canvas.Plot and PlotOn both bounds-check and return silently (canvas.go:49
// and :59), so drawing past the bottom is safe and this file contains no
// clipping logic of its own. Verified, not assumed.
//
// ⚠ OFF BY DEFAULT. With XSCAPES_NEAR unset nothing in this file runs: the
// ladder has one rung, Approach pins the progress at 0, and drawCrab takes the
// shipped path it always took. TestNothingMovesWithTheNearPoseUnset holds that
// against a hash of the pre-change render.
//
// Two rungs on purpose, because the size is HIS call and not mine -- he picked
// Hero from four silhouettes over two rounds and the coat from five, and the
// difference between these two is the kind of thing that only reads live:
//
//	XSCAPES_NEAR=1   the authored 16x9 pose (notes/s28-closer/critic/pick.go)
//	XSCAPES_NEAR=2   the exact 2x, 24 cells x 14 rows, which clips at the bottom
//
// Near is which rung of the ladder is armed. Package var read once from the
// environment, the shape scape.Tide uses for exactly this -- a feature he tries
// behind a switch before ruling on it.
var Near = 0

func init() { Near = nearFromEnv(envx.Lookup("NEAR")) }

// nearFromEnv reads the flag. Anything that is not a rung is OFF, silently and
// on purpose: this is typed by hand mid-session while judging a frame, and a
// typo should leave him with the picture he already knows rather than an error
// in the middle of his work. Pure so the mapping can be tested; init cannot.
func nearFromEnv(v string) int {
	switch v {
	case "0", "off", "no":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	}
	// Unset is the FULL approach, his ruling of 2026-09-11. He tried it behind
	// the switch -- "looks great" -- and then asked of his own live session
	// "The main companion does not come closer to the screen when prompting the
	// user- did you push that feature live?", which it was not. It is now.
	// Rung 2 rather than rung 1 because the one he approved is the CROPPED one:
	// "I was referring to the cropped version, which I liked- the horizontal
	// crop." XSCAPES_NEAR=0 puts the shipped crab back.
	return nearDefault
}

// nearDefault is the rung an unset XSCAPES_NEAR gets. The same shape the tide
// took: behind a switch, lived in, then turned on at his word.
const nearDefault = 2

// nearStep is how long ONE rung of the walk is held, in seconds.
//
// It is pace.go's stepDur, and it is deliberately not a new number: the
// companion already crosses one cell in 0.28 s every time a main-thread tool
// event lands, so an approach built on the same beat reads as the same animal
// taking the same strides toward you. Duplicated rather than imported because
// pace.go is in package main; the same reason bitmap.go carries its own copy of
// scape's hash.
const nearStep = 0.28

// shippedCrabCells is the box every companion is drawn into and the width the
// near poses are centred against. Size() returns this forever -- it is the
// LAYOUT contract, and compose() derives the companion's margin, the litter's
// room, SandTo and the moon's column from it.
const shippedCrabCells = 12

// Approach eases the companion toward (near = true) or away from (near = false)
// the near pose. dt is seconds since the last frame.
//
// Self-contained on purpose: no reducer change, no new state anywhere else, and
// the retreat falls out of the same call. The retreat matters as much as the
// walk -- the instant he answers is the one frame he is certainly looking at --
// and it is also what keeps the pose honest when Worried outranks NeedsYou
// (reduce.go:606) and a main-thread error during an ask flips the face: the
// caller simply stops asking for near, and the animal walks back rather than
// vanishing a row at a time.
//
// The progress is linear, and the RUNGS are what make it read as steps rather
// than as a zoom: one rung per nearStep, so XSCAPES_NEAR=2 goes 12x7 -> 16x9 ->
// 24x14 in two strides of 0.28 s. A continuous scale between those sizes is not
// available in this medium anyway (every generated in-between rung came out
// asymmetric -- notes/s28-closer/critic/ladder.go) and would have read as a
// camera push rather than as an animal walking.
func (c *Cat) Approach(dt float64, near bool) {
	steps := c.nearSteps()
	if steps == 0 {
		// Off, or the cat, which has no near art. Pin it: a progress left
		// part-way by a flag flipped mid-session would size a box that no
		// longer has a pose to fill it.
		c.approach = 0
		return
	}
	if dt < 0 {
		dt = 0
	}
	if dt > nearStep {
		// A stalled frame must not teleport the walk. The live loop delivers
		// dt at the frame rate, but a resize, a compaction or a laptop lid
		// delivers whatever it likes, and the shore hit the same class of bug
		// scaling elapsed time instead of advancing phase (shore.go:155).
		dt = nearStep
	}
	d := dt / (float64(steps) * nearStep)
	if !near {
		d = -d
	}
	c.approach += d
	if c.approach > 1 {
		c.approach = 1
	}
	if c.approach < 0 {
		c.approach = 0
	}
}

// DrawnBox is the cells the sprite will ACTUALLY occupy this frame, as an
// offset from the x that will be passed to Draw, plus the drawn size. dx is
// <= 0, because the near poses are centred on the shipped head column and
// therefore hang off the box to the LEFT, which is into the scene rather than
// into the frame edge. With XSCAPES_NEAR unset this is always (0, 12, 7).
//
// It exists so the root package can keep the litter out from under a companion
// that is now wider than its box, without knowing anything about the ladder.
func (c *Cat) DrawnBox(frameW int) (dx, w, h int) {
	w, h = c.Size()
	r := c.nearRung(frameW)
	if r == 0 {
		return 0, w, h
	}
	b := nearBoxes[r]
	return c.nearDX(b.w), b.w, b.h
}

// nearDX places a near pose so that growing does not eat the right margin.
//
// HIS RULING, 2026-09-10, on the first live look at the near crab: "Only think
// is Id keep the companion away from the far right edge of the screen, needs
// some negative space when it comes closer."
//
// It is the same note he made in session 15 at 124 columns -- "the companion
// feels too pushed to the side of the screen. a bit" -- which is why compose()
// grows the margin with the width at all (right = 2 + w/32). The near pose
// reintroduced it, and measured on the painted frame the gap between the
// companion's rightmost ink and the frame edge ran BACKWARDS as it approached:
//
//	            rung 0    rung 1    rung 2
//	w=115         9         7         3
//	w=80          7         5         1
//	w=40          3         1         0      <- touching the edge
//
// The first version CENTRED each rung on the shipped head column, which is
// cheap -- an even-width sprite at catX-(W-12)/2 keeps the head column exactly,
// so nothing downstream moves -- but half of every new column is spent to the
// RIGHT, into the margin. An animal that comes closer was given less air, which
// is backwards: the bigger it is, the more it needs.
//
// So the sprite hangs entirely to the LEFT, into the scene, and then a little
// further. The first term keeps the right edge exactly where the shipped
// crab's is, so no rung is ever tighter than today; the second is his "some
// negative space", a quarter of the growth, so the air OPENS as it arrives:
// at his window the gap goes 9 / 10 / 12 instead of 9 / 7 / 3.
//
// The cost is that the head column does move, so the balloon must follow it --
// see DrawnHeadCol, which live.go anchors the pointer to. That is the whole
// price and it is one call site.
// ⚠ MIRRORED ONLY. The companion is on the RIGHT in the shipped composition
// and the scene is to its left, so hanging left is hanging INTO the scene. The
// UNMIRRORED layout (-mirror=false) puts it at CatX 5 on the LEFT with the sand
// to its right, and there the same shift walks it off the frame -- measured,
// the pointer landed on column 3 with the eyes at 4 and 5 at every width. The
// rule that covers both is "grow into the scene, away from your own edge", so
// unmirrored grows RIGHT from CatX and needs no offset at all.
func (c *Cat) nearDX(w int) int {
	if !c.mirror {
		return 0
	}
	grow := w - shippedCrabCells
	return -grow - grow/4
}

// nearFits caps the rung at what the frame can actually pay for.
//
// The near pose takes its columns from the SAND and the LITTER, which share the
// beach with it, and at the design floor there is simply not room: measured at
// 40 columns, rung 2 is 24 cells of a 40-cell frame and the activity tail got
// ZERO cells -- the writing channel went dark altogether. That is the same
// trade s28 already rejected once, when reserving the litter's whole span from
// SandTo left 8 of 56 columns of sand at 80 columns with fourteen subagents.
//
// Half the frame is the bar. It is not tuned: a companion wider than half the
// picture has stopped being a companion IN a scene and become the scene. At 40
// columns that drops rung 2 to rung 1 (16 of 40) and the tail comes back; at 48
// and up rung 2 stands, which covers every window he works in.
func nearFits(rung, frameW int) int {
	for rung > 0 && nearBoxes[rung].w*2 > frameW {
		rung--
	}
	return rung
}

// DrawnHeadCol is where the balloon's pointer belongs, as an offset from the x
// that will be passed to Draw. At rung 0 it is HeadCol() and nothing has
// changed; at a near rung it is that rung's own eye midpoint, shifted by the
// same nearDX the sprite is drawn at.
//
// It exists because nearDX stopped preserving the head column (see above). The
// eye positions are read from the pose's own eyeCells rather than assumed to be
// the box centre, so the pointer cannot drift from the face if the art is ever
// redrawn -- which is the same reason HeadCol reads catEyeCells.
func (c *Cat) DrawnHeadCol(frameW int) int {
	r := c.nearRung(frameW)
	if r == 0 || c.kind != KindCrab {
		return c.HeadCol()
	}
	p := c.nearPoseFor(r, NeedsYou, 0)
	// eyeCells are the LEFT cell of each 2-cell eye, so the pair spans
	// eyeCells[0] .. eyeCells[1]+1 and the midpoint is between them.
	return c.nearDX(nearBoxes[r].w) + (p.eyeCells[0]+p.eyeCells[1]+nearEyeWide)/2
}

// nearSteps is how many rungs the walk climbs. Zero disables everything in this
// file.
func (c *Cat) nearSteps() int {
	if c.kind != KindCrab {
		// The near pose is Hero's. The cat has no art at this size and is not
		// getting any redrawn for a switch he has not ruled on yet, so the cat
		// simply never approaches -- and the root package can still call
		// Approach and DrawnBox unconditionally.
		return 0
	}
	if Near < 1 {
		return 0
	}
	if Near > len(nearBoxes)-1 {
		return len(nearBoxes) - 1
	}
	return Near
}

// nearRung is which pose is drawn this frame: 0 the shipped crab, 1 the
// authored 16x9, 2 the exact 2x.
//
// Floor, not round, so each rung is held for a whole nearStep rather than for
// half of one at each end, and so rung 0 is what is drawn for any progress
// below the first boundary -- which is what makes the walk start from the
// shipped picture instead of jumping into it.
func (c *Cat) nearRung(frameW int) int {
	steps := c.nearSteps()
	if steps == 0 {
		return 0
	}
	r := int(c.approach*float64(steps) + 1e-9)
	if r > steps {
		r = steps
	}
	if r < 0 {
		r = 0
	}
	// Cap it at what the frame can pay for. Zero means nobody has said how wide
	// the frame is, and an unknown width must not silently shrink the animal --
	// every study page and every test that predates this draws without one.
	if frameW > 0 {
		r = nearFits(r, frameW)
	}
	return r
}

// ---------------------------------------------------------------------------
// RUNG 1: the authored 16x9.
//
// Moved in verbatim from notes/s28-closer/critic/pick.go, where it was drawn
// and blind-checked. Not redrawn here and not to be redrawn: it is the size the
// room actually allows that is CLOSEST to a uniform push-in of the shipped crab
// -- 32x36 source px, 1.33x wide and 1.29x tall, 4% off a true magnification.
//
// Authored rather than scaled because every generated rung came out asymmetric.
// It is built from column RANGES rather than typed by hand for the same reason:
// the first typed version put one stalk at cols 11-12 and the other at 20-21,
// which rendered one stalk as a full cell and the other as two half cells and
// grew a bump on one side of the shell. Symmetry in this medium is a property
// of even column parity, not of the eye.
// ---------------------------------------------------------------------------

const nearW1 = 32 // source width of the rung-1 art

type nearSpan struct{ a, b int }

func nearRow(sp ...nearSpan) string {
	r := []rune(strings.Repeat(".", nearW1))
	for _, s := range sp {
		for i := s.a; i <= s.b; i++ {
			r[i] = '#'
		}
	}
	return string(r)
}

func nearSpans(a ...[]nearSpan) []nearSpan {
	var out []nearSpan
	for _, s := range a {
		out = append(out, s...)
	}
	return out
}

// Every feature of the shipped crab multiplied by 4/3 and rounded to what the
// medium can hold: 3px pincer fingers instead of 2, a 4px arm instead of 2, and
// 2px stalks STRADDLING a cell boundary so that a 4px (2 cell) eye caps them
// exactly.
var (
	nearRFing  = []nearSpan{{24, 26}, {29, 31}} // raised pincer, two prongs
	nearRPalm  = []nearSpan{{24, 31}}
	nearRWrist = []nearSpan{{25, 30}}
	nearRArm   = []nearSpan{{26, 29}}
	nearLFing  = []nearSpan{{0, 2}, {5, 7}} // the low pincer, its mirror
	nearLPalm  = []nearSpan{{0, 7}}
	nearLWrist = []nearSpan{{1, 6}}
	nearLArm   = []nearSpan{{2, 5}}
	nearStalks = []nearSpan{{11, 12}, {19, 20}}
)

// nearAsk1 is the ask pose: one claw up, authored RIGHT so the production flip
// puts it on screen-LEFT pointing at the transcript, exactly as crabAsk does.
var nearAsk1 = []string{
	nearRow(nearRFing...),
	nearRow(nearRFing...),
	nearRow(nearRFing...),
	nearRow(nearRPalm...),
	nearRow(nearSpans(nearRPalm, nearStalks)...),
	nearRow(nearSpans(nearRPalm, nearStalks, nearLFing)...),
	nearRow(nearSpans(nearRWrist, nearStalks, nearLFing)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLFing)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLPalm)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLPalm)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLWrist)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLArm)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLArm)...),
	nearRow(nearSpans(nearRArm, nearStalks, nearLArm)...),
	nearRow(nearSpan{3, 28}), // shell top
	nearRow(nearSpan{1, 30}),
}

// nearLevel1 is the same animal with both claws down and level. It is what a
// rung-1 crab wears for every state that is not the ask, and it exists for the
// RETREAT: he answers, the pose leaves NeedsYou, and the animal has to walk
// back down the ladder wearing something that is not a raised claw.
//
// ⚠ The cost, stated rather than hidden: rung 1 has two authored poses, so a
// retreat into Worried shows level claws for the 0.28 s of that one rung before
// the shipped worried silhouette takes over. Worried's own channel -- the
// widest thing on the beach becoming the smallest -- is intact at rung 0 and at
// rung 2, and buying it at rung 1 means drawing a third 16x9 pose by hand for
// one step of one transition. Rung 2 needs no such note because it is generated
// from whatever the shipped state draws.
var nearLevel1 = []string{
	nearRow(nearSpans(nearRFing, nearLFing)...),
	nearRow(nearSpans(nearRFing, nearLFing)...),
	nearRow(nearSpans(nearRFing, nearLFing)...),
	nearRow(nearSpans(nearLPalm, nearRPalm)...),
	nearRow(nearSpans([]nearSpan{{1, 6}, {25, 30}}, nearStalks)...),
	nearRow(nearSpans([]nearSpan{{1, 6}, {25, 30}}, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpans(nearLArm, nearRArm, nearStalks)...),
	nearRow(nearSpan{3, 28}),
	nearRow(nearSpan{1, 30}),
}

var nearLower1 = []string{
	"################################",
	"################################",
	".##############################.",
	"..############################..",
	"...##########################...",
	"...##########################...",
	"..###...################...###..", // legs
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
	".##..........................##.",
	".##..........................##.",
}

// ---------------------------------------------------------------------------
// RUNG 2: the exact 2x.
//
// Generated, never typed: every source pixel becomes 2x2 source pixels and the
// production ToQuadrant does the rest. 24x28 source -> 48x56 -> 24 cells x 14
// rows. Generating it is the point -- it cannot drift from the shipped art, and
// every state, every claw frame and the mid-stride legs come along for free.
//
// ⚠ ITS KNOWN COST, MEASURED, and it is exactly the kind of trade he rules on:
// the glyph vocabulary collapses from 11 distinct quadrant glyphs to 3 (full,
// upper half, lower half). Doubling horizontally puts both subpixels of a cell
// in the same original source column, so TL always equals TR and BL always
// equals BR, and no partial quadrant can occur. Every diagonal becomes a
// one-cell staircase.
//
// ⚠ The tempting conclusion -- "2x halves the horizontal resolution" -- is
// WRONG and was the first thing believed. As a fraction of the ANIMAL nothing
// changes: half a cell of a 12-cell body and one whole cell of a 24-cell body
// are both a twenty-fourth of its width, and the VERTICAL detail doubles,
// because ToQuadrant's OR of source rows 2k and 2k+1 is lossless when both rows
// came from the same original. What changes is absolute size: the same edge
// cells sit in a body four times the area, so each step of the staircase is
// twice as tall and twice as wide on his screen.
// ---------------------------------------------------------------------------

// nearDouble is nearest-neighbour INTEGER scaling of a SOURCE bitmap: every
// source pixel becomes 2x2 pixels. Exact by construction -- no interpolation,
// no threshold, no new ink anywhere the source had none, and symmetric because
// an integer factor maps every column the same way.
func nearDouble(rows []string) []string {
	out := make([]string, 0, len(rows)*2)
	for _, r := range rows {
		var sb strings.Builder
		for _, c := range r {
			sb.WriteRune(c)
			sb.WriteRune(c)
		}
		line := sb.String()
		out = append(out, line, line)
	}
	return out
}

// crab2x is every shipped crab bitmap and its double, keyed by the address of
// the source's first row -- the backing array, which is identity and cannot
// collide the way the row TEXT can (crabWork2 and crabRest open with the same
// blank row). A pose that is not in the table falls through to doubling on the
// spot rather than panicking: a new shipped state should cost a frame of work,
// not a crash in the middle of his session.
var crab2x = map[*string][]string{}

func init() {
	for _, rows := range [][]string{
		crabWork, crabWork2, crabRest, crabAsk, crabDone, crabWorried,
		crabLower, crabLowerStep,
	} {
		crab2x[&rows[0]] = nearDouble(rows)
	}
}

func double2x(rows []string) []string {
	if d, ok := crab2x[&rows[0]]; ok {
		return d
	}
	return nearDouble(rows)
}

// ---------------------------------------------------------------------------
// THE EYE AT SIZE.
//
// The eyes are single-cell GLYPHS today -- 'O' alert, 'o', '-', '^' -- and a
// character cannot be scaled: a cell is a cell. On the shipped 12-cell crab one
// cell is a twelfth of the face and reads as an eye; on a 24-cell crab the same
// glyph is 1 cell of 336 against 1 of 84, a QUARTER of the fraction it reads at
// now, and it becomes a speck of dirt on a shell. So the near eye is its own
// small BITMAP pass: 4 source px by 8, which is 2 cells by 2, which is 4 of 336
// -- today's fraction, exactly. It is not a preference, it is the only size
// that holds the ratio.
//
// ⚠ THE ALERT EYE IS A RING AND THE RING'S HOLE IS PAINTED. The hole is what
// says wide open -- a solid block is a blob -- but canvas.Plot REPLACES a cell,
// so an unpainted hole shows the SEA through the middle of the eye. That is his
// own "two blue holes" report from 2026-09-04, four times the area. Every cell
// of the eye pass therefore goes through PlotOn with a ground, INCLUDING the
// cells whose glyph is a space: a space on a coat ground is a solid coat cell,
// and that is what fills the ring.
//
// The ground is eyeGround() when a fill is configured, and the coat otherwise,
// rather than eyeGround()'s "no fill" answer -- at this size there is no honest
// version of the hole that is not the animal.
// ---------------------------------------------------------------------------

// nearEyeAlert is NeedsYou: the ring. This is the state the whole feature
// exists for, and at 2 cells by 2 the ask can finally be a different SHAPE from
// working rather than the same cell in a different colour -- at 1x it was 'O'
// against 'o', which is a channel the shipped crab has never really had.
// nearEyeWide is how many CELLS one eye spans. The eye art is four source
// columns and a cell is two of them, so the pair spans eyeCells[0] to
// eyeCells[1]+1 and DrawnHeadCol reads the midpoint off that.
const nearEyeWide = 2

var nearEyeAlert = []string{
	"####",
	"####",
	"#..#",
	"#..#",
	"#..#",
	"#..#",
	"####",
	"####",
}

// nearEyeOpen is the working / worried bead. Written twice per subpixel row
// because ToQuadrant ORs source rows 2k and 2k+1: a row on its own is eaten
// before it is ever drawn.
var nearEyeOpen = []string{
	"....",
	"....",
	".##.",
	".##.",
	".##.",
	".##.",
	"....",
	"....",
}

// nearEyeShut is the blink and the resting '-': a lid.
var nearEyeShut = []string{
	"....",
	"....",
	"....",
	"....",
	"####",
	"####",
	"....",
	"....",
}

// nearEyeDone is the content '^'.
var nearEyeDone = []string{
	"....",
	"....",
	".##.",
	".##.",
	"#..#",
	"#..#",
	"....",
	"....",
}

// nearEye picks the eye bitmap and its colour for a state, on exactly the
// shipped mapping and the shipped blink schedule, so the near pose says the
// same things the small one does.
func nearEye(st State, t float64) ([]string, term.RGB) {
	art, col := nearEyeOpen, eyeCol
	switch st {
	case Resting:
		art = nearEyeShut
	case NeedsYou:
		art, col = nearEyeAlert, EyeAlert
	case Worried:
		col = eyeWorried
	case Done:
		art = nearEyeDone
	}
	if st != Resting && st != Worried && st != Done && math.Mod(t, 5.3) < 0.16 {
		art = nearEyeShut
	}
	return art, col
}

// ---------------------------------------------------------------------------
// THE POSES, AND WHERE THE EYES SIT ON THEM.
// ---------------------------------------------------------------------------

// nearPose is one rung's art for one frame.
type nearPose struct {
	upper, lower []string
	// eyeCells are the LEFT cell of each 2-cell eye, in the authored
	// (unmirrored) box; eyeRow is the TOP cell row of it.
	eyeCells [2]int
	eyeRow   int
	// lift is the breath, in source rows. Even, because one quadrant subpixel
	// is two source rows and an odd lift would land between subpixels.
	lift int
}

// nearBoxes is each rung's drawn footprint in cells. Rung 0 is the shipped box.
// The other two are measured off the art at startup rather than written down,
// so a change to the art cannot leave the layout describing the old size.
var nearBoxes = [3]struct{ w, h int }{
	{shippedCrabCells, 7},
	nearBoxOf(nearAsk1, nearLower1),
	// nearDouble rather than double2x: package var initialisers run BEFORE
	// init(), so the cache is still empty here. Same answer, no dependence on
	// the order.
	nearBoxOf(nearDouble(crabAsk), nearDouble(crabLower)),
}

func nearBoxOf(upper, lower []string) struct{ w, h int } {
	q := ParseBitmap(append(append([]string{}, upper...), lower...)).ToQuadrant()
	return struct{ w, h int }{len([]rune(q[0])), len(q)}
}

// nearPoseFor assembles the rung's art for this state and frame.
//
// THE EYE ROWS ARE CHOSEN, NOT INHERITED, and the reason is one measurement.
// At 1x the stalk tip and the shell top SHARE a cell -- the stalk is half a
// character tall -- so the shipped eye glyph sits half on the stalk and half on
// the shell because there is nowhere else for it to go. At both near sizes they
// stop sharing, and a 2-cell eye put where the 1x one sits is TALLER than the
// stalk: it eats the whole thing and the eyes read as two tiles stuck on the
// shell. Each rung's eye is therefore raised until a full cell of salmon stalk
// stands between it and the shell, which is what a stalked eye is:
//
//	rung 1  stalk at cell rows 1-3 (3 is where the shell arrives under it),
//	        eye at rows 0-1, leaving row 2 of clear stalk
//	rung 2  stalk at cell row 4 alone, shell from 5, eye at rows 2-3
//
// The eye pass paints its own ground, so a row above the body costs nothing and
// simply carries the stalk up: the tile IS the head.
func (c *Cat) nearPoseFor(rung int, st State, t float64) nearPose {
	if rung >= 2 {
		lower := crabLower
		if c.stepping && st != Worried {
			lower = crabLowerStep
		}
		return nearPose{
			upper: double2x(crabUpper(st, t)),
			lower: double2x(lower),
			// At 2x one cell column IS one original source column, so the
			// shipped eye cells {4, 7} -- source columns 8-9 and 14-15 -- keep
			// their columns exactly. The stalk is cell row 4 and the shell
			// starts at 5.
			eyeCells: [2]int{8, 14},
			eyeRow:   2,
			// TWO source rows, not four, and this one was measured rather than
			// reasoned. A doubled bitmap shifted by two rows moves by exactly
			// ONE ORIGINAL source row, which is half a character cell -- the
			// shipped breath exactly. Four rows would keep the same FRACTION of
			// a body four times the area, and that is a WHOLE cell: the eye
			// pass does not breathe with the body (the shipped one does not
			// either, because at 1x a half-cell lift cannot leave its cell), so
			// a whole-cell lift slides the shell up into the eye's own row and
			// the stalk disappears on every inhale. Rendered both ways.
			lift: 2,
		}
	}
	upper := nearLevel1
	if st == NeedsYou {
		upper = nearAsk1
	}
	// The stalks are authored at source cols 11-12 and 19-20, straddling a cell
	// boundary on purpose, so a 4px (2 cell) eye caps them exactly: cells 5-6
	// and 9-10. Symmetry in this medium is a property of even column parity.
	return nearPose{upper: upper, lower: nearLower1, eyeCells: [2]int{5, 9}, eyeRow: 0, lift: 2}
}

// drawCrabNear is the whole near draw: one branch off drawCrab, the same
// compose-breathe-mirror-plot it does, then the eye pass.
func (c *Cat) drawCrabNear(l *canvas.Layer, x, y int, t float64, st State, rung int) {
	p := c.nearPoseFor(rung, st, t)
	src := ParseBitmap(append(append([]string{}, p.upper...), p.lower...))

	f := src.Blank()
	period := crabBreathPeriod(st)
	lift := 0
	if math.Sin(t*2*math.Pi/period) > 0.35 {
		lift = p.lift
	}
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if src.at(sx, sy+lift) {
				f.Set(sx, sy)
			}
		}
	}
	if c.mirror {
		f = f.Mirrored()
	}
	q := f.ToQuadrant()

	w := len([]rune(q[0]))
	bx := x + c.nearDX(w)
	plotRim(l, q, bx, y)
	(&Sprite{Rows: q, Body: c.coat}).Draw(l, bx, y)

	art, col := nearEye(st, t)
	ground, ok := c.eyeGround()
	if !ok {
		ground = c.coat
	}
	eye := ParseBitmap(art).ToQuadrant()
	for _, e := range p.eyeCells {
		ex := bx + e
		if c.mirror {
			// The eye is two cells wide and is plotted in CELL space after the
			// bitmap flip, so its LEFT cell mirrors to w-2-e, not to w-1-e.
			ex = bx + w - 2 - e
		}
		for dy, row := range eye {
			for dx, r := range []rune(row) {
				// Every cell, spaces included: a space on the coat ground is a
				// solid coat cell, and that is what keeps the sea out of the
				// ring's hole.
				l.PlotOn(ex+dx, y+p.eyeRow+dy, r, col, ground, 1)
			}
		}
	}
}
