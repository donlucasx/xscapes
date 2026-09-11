package main

import "strings"

// The art. NOTHING here is authored in this directory -- every row is copied
// verbatim from where it already lives, and the file is arranged so a reader
// can check that by diffing:
//
//	pickUpperRows / pickLower / the eye bitmaps  <- notes/s28-closer/critic/pick.go
//	aUpper / aLower / aEye                       <- notes/s28-closer/critic/art.go
//	crabAskRows / crabLowerRows                  <- notes/s28-closer/critic/shipped.go,
//	                                                which copied internal/companion/crab.go
//
// The picked pose was blind-checked at those exact pixels. Redrawing it here
// to "clean it up" would throw that result away and quietly preview something
// he was never offered.

// ---------------------------------------------------------------------------
// THE PICK, verbatim from notes/s28-closer/critic/pick.go.
// 16 cells x 9 rows, 32x36 source px.
// ---------------------------------------------------------------------------

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

// The eye vocabulary at 2 cells x 2, verbatim from pick.go. A 4x8 source
// halves to a 4x4 subpixel grid, which is exactly enough for a ring, a bead
// and a lid -- the same four states the shipped one-cell glyph carries.
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

// ---------------------------------------------------------------------------
// A, the minimal step -- 14 cells x 9 rows, 28x36 px. Verbatim from
// notes/s28-closer/critic/art.go. Used HERE only as the in-between frame of
// the approach: it is a rung of the same ladder that was already drawn and
// already looked at, so the walk-up needs no new art.
// ---------------------------------------------------------------------------

var aUpper = []string{
	"....................###..###",
	"....................###..###",
	"....................###..###",
	"....................########",
	"###..###............########",
	"###..###.............######.",
	"###..###..............####..",
	"########..............####..",
	"########..............####..",
	".######..##......##...####..",
	"..####...##......##...####..",
	"..####...##......##...####..",
	"..####..####....####..####..",
	"..####..####....####..####..",
	"..########################..",
	".##########################.",
}

var aLower = []string{
	"############################",
	"############################",
	".##########################.",
	"..########################..",
	"...######################...",
	"....####################....",
	"..##...##############...##..",
	"..##...##############...##..",
	".##.....############.....##.",
	".##.....############.....##.",
	"##.......##########.......##",
	"##.......##########.......##",
	"##........########........##",
	"##........########........##",
	".#.........######.........#.",
	".#.........######.........#.",
	".#..........####..........#.",
	".#..........####..........#.",
	".#........................#.",
	".#........................#.",
}

var aEye = []string{".##.", ".##.", "####", "####", "####", "####", ".##.", ".##."}

// ---------------------------------------------------------------------------
// The SHIPPED crab, verbatim from notes/s28-closer/critic/shipped.go, which
// copied internal/companion/crab.go. Only used to prove that the replica draw
// in pose.go renders identically to companion's own unexported drawCrab; the
// shipped crab everywhere else in this preview is the real companion.NewCrab().
// ---------------------------------------------------------------------------

func crabLowerRows() []string {
	return []string{
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
}

func crabAskRows() []string {
	return []string{
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
}
