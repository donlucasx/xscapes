package main

import "strings"

// pick: the size the room actually allows that is CLOSEST to a uniform push-in
// of the shipped crab -- 16 cells x 9 rows, 32x36 source px, 1.33x wide and
// 1.29x tall, 4% off a true magnification and the largest such step the beach
// can pay for.
//
// Authored, not scaled: every generated rung came out asymmetric (see ladder).
// The structure is the shipped crab's own with every feature multiplied by 4/3
// and rounded to what the medium can hold: 3px pincer fingers instead of 2, a
// 4px arm instead of 2, and 2px stalks straddling a cell boundary so that a
// 4px (2 cell) eye caps them exactly.
//
// Built from column RANGES rather than typed by hand. My first version was
// typed and put one stalk at cols 11-12 and the other at 20-21; it rendered
// one stalk as a full cell and the other as two half cells, and the shell rim
// grew a bump on one side only. Symmetry in this medium is a property of even
// column parity, not of the eye.

const pw = 32 // source width

type span struct{ a, b int }

func row32(sp ...span) string {
	r := []rune(strings.Repeat(".", pw))
	for _, s := range sp {
		for i := s.a; i <= s.b; i++ {
			r[i] = '#'
		}
	}
	return string(r)
}

var (
	rFing  = []span{{24, 26}, {29, 31}} // raised pincer, two prongs
	rPalm  = []span{{24, 31}}
	rWrist = []span{{25, 30}}
	rArm   = []span{{26, 29}}
	lFing  = []span{{0, 2}, {5, 7}} // low pincer, the mirror of it
	lPalm  = []span{{0, 7}}
	lWrist = []span{{1, 6}}
	lArm   = []span{{2, 5}}
	stalks = []span{{11, 12}, {19, 20}} // straddling, so a 2-cell eye caps them
)

func cat2(a ...[]span) []span {
	var out []span
	for _, s := range a {
		out = append(out, s...)
	}
	return out
}

// pickUpperRows is the ask pose: one claw up, authored RIGHT so the flip puts
// it on screen-left pointing at the transcript, exactly as crabAsk does.
func pickUpperRows() []string {
	return []string{
		row32(rFing...),                       // r0
		row32(rFing...),                       // r1
		row32(rFing...),                       // r2
		row32(rPalm...),                       // r3
		row32(cat2(rPalm, stalks)...),         // r4
		row32(cat2(rPalm, stalks, lFing)...),  // r5
		row32(cat2(rWrist, stalks, lFing)...), // r6
		row32(cat2(rArm, stalks, lFing)...),   // r7
		row32(cat2(rArm, stalks, lPalm)...),   // r8
		row32(cat2(rArm, stalks, lPalm)...),   // r9
		row32(cat2(rArm, stalks, lWrist)...),  // r10
		row32(cat2(rArm, stalks, lArm)...),    // r11
		row32(cat2(rArm, stalks, lArm)...),    // r12
		row32(cat2(rArm, stalks, lArm)...),    // r13
		row32(span{3, 28}),                    // r14 shell top
		row32(span{1, 30}),                    // r15
	}
}

// pickWorkRows is the same animal with both claws down and level -- the
// Working pose at the new size, drawn only to prove the ask reads as a
// DIFFERENT pose and not merely as a bigger crab.
func pickWorkRows() []string {
	rf := []span{{24, 26}, {29, 31}}
	lf := []span{{0, 2}, {5, 7}}
	return []string{
		row32(cat2(rf, lf)...),
		row32(cat2(rf, lf)...),
		row32(cat2(rf, lf)...),
		row32(cat2(lPalm, rPalm)...),
		row32(cat2([]span{{1, 6}, {25, 30}}, stalks)...),
		row32(cat2([]span{{1, 6}, {25, 30}}, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(cat2(lArm, rArm, stalks)...),
		row32(span{3, 28}),
		row32(span{1, 30}),
	}
}

var pickLower = []string{
	"################################", // r0
	"################################", // r1
	".##############################.", // r2
	"..############################..", // r3
	"...##########################...", // r4
	"...##########################...", // r5
	"..###...################...###..", // r6  legs
	"..###...################...###..", // r7
	".###.....##############.....###.", // r8
	".###.....##############.....###.", // r9
	"###.......############.......###", // r10
	"###.......############.......###", // r11
	"###........##########........###", // r12
	"###........##########........###", // r13
	".##.........########.........##.", // r14
	".##.........########.........##.", // r15
	".##..........######..........##.", // r16
	".##..........######..........##.", // r17
	".##..........................##.", // r18
	".##..........................##.", // r19
}

// The eye: 4px wide by 8px tall = 2 cells by 2. Round, no pupil -- a pupil
// needs a second colour inside the same cell, and canvas.Plot REPLACES the
// cell, so a second pass deletes the eye rather than compositing into it.
var pickEye = []string{
	".##.",
	".##.",
	"####",
	"####",
	"####",
	"####",
	".##.",
	".##.",
}

// The eye vocabulary at 2 cells x 2. The shipped face distinguishes states with
// a GLYPH swap in one cell -- 'o' working, 'O' alert, '-' blinking, '^' done.
// A bitmap has to reproduce that, and it can: a 4x8 source halves to a 4x4
// subpixel grid, which is exactly enough for a ring, a bead and a lid.
var (
	eyeAlertBM = []string{ // wide open: a RING, the 'O'
		"####",
		"####",
		"#..#",
		"#..#",
		"#..#",
		"#..#",
		"####",
		"####",
	}
	eyeOpenBM = []string{ // working: a bead, the 'o'
		"....",
		"....",
		".##.",
		".##.",
		".##.",
		".##.",
		"....",
		"....",
	}
	eyeShutBM = []string{ // blink / resting: a lid, the '-'
		"....",
		"....",
		"....",
		"....",
		"####",
		"####",
		"....",
		"....",
	}
	eyeDoneBM = []string{ // content: the '^'
		"....",
		"....",
		".##.",
		".##.",
		"#..#",
		"#..#",
		"....",
		"....",
	}
)
