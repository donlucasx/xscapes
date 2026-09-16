package scenes

import (
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE OWL'S MOTION ROUND (s36, 2026-09-16). His notes on the review sheet of
// the five faces, verbatim:
//
//	Resting: "I like it, but every now and then it should open its eyes, and
//	         look to either side, and then close them again."
//	Working: "I like it, but it should look the other way every now and then.
//	         Also it should do a lil bob or something else- needs a bit more
//	         life. Let's try some alternatives. Maybe it hops, opens its wings
//	         for moment and settles back down?"
//	Needs you: "Lets add an eye blink every now and then. Or double blink as
//	         if trying to get your attention."
//	Done:    "needs some sort of action or movement too"
//	Worried: "can we try some alts?"
//	Owlets:  "lets expand on the animations"
//
// Every candidate here is a MOTION laid over a state's face: what the owl
// does now and then, on a clock, carrying no information (the argument that
// lets the crab's pincer shut and the cat's near eyes be random). The face
// stays the channel. One trade is real and is his to make: a resting owl
// that opens its eyes for a glance is, in that second, a screenshot of a
// working owl. Each candidate says what fraction of its period the face is
// not the state's own.
//
// Nothing here reaches the live vista until a pick is made: OwlMotionPick
// and OwletMotionPick are -1, which draws exactly what DrawOwlPose and
// DrawOwlets draw (TestTheMotionRoundStartsFromToday).

// Lids is what the eyelids do in a frame.
type Lids int

const (
	LidsState  Lids = iota // the state's own face
	LidsDown               // shut, a line across the top: resting, a blink
	LidsOpen               // open, the pupil where the state puts it
	LidsUp                 // shut upward, '^ ^': done
	LidsSlit               // amber slits: worried
	LidsHalf               // half open: a lid line over an open eye
	LidsSquint             // narrowed to a band a cell tall across the two rows
)

// OwlMod is one frame's departure from the state's still pose.
type OwlMod struct {
	Lids   Lids
	Look   int  // the pupil's column shift: +1 the other way
	OneEye bool // Lids and Look apply to the near eye only
	Brows  bool // worried brows over the eyes, raised toward the middle
	Hop    int  // whole rows up, eyes and all
	Wings  int  // see owlWings: 0 folded, 1 out, 2 up, 3 small out, 4 small up, 5 one small out, 6 one small up
	DX     int  // the body alone, in half-cells; the eyes stay
	DY     int  // the body alone, in pixels, positive up (2 = half a cell)
}

// OwlMotion is one candidate: what the owl does over one Period, given the
// seconds u since the period began.
type OwlMotion struct {
	Name, Note string
	Period     float64
	Round      int // the review round it was drawn for; the page shows the latest
	At         func(u float64) OwlMod
}

// flutter is the wings beating: small out and small up by turns, five
// beats a second, on both wings or on one.
func flutter(u float64, one bool) int {
	k := 3
	if one {
		k = 5
	}
	if int(u*10)%2 == 1 {
		k++
	}
	return k
}

// in says whether u is inside [a, b).
func in(u, a, b float64) bool { return u >= a && u < b }

// OwlMotions are the candidates per state, his notes above answered one
// by one; index 0 is always today's owl, the control.
var OwlMotions = map[companion.State][]OwlMotion{
	companion.Resting: {
		{Name: "today", Note: "lids down, still", Period: 12},
		{Name: "glance", Note: "opens, looks left, looks right, looks left, closes: 1.5 s in 12 (12% not resting)", Period: 12,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.4):
					return OwlMod{Lids: LidsOpen}
				case in(u, 0.4, 0.9):
					return OwlMod{Lids: LidsOpen, Look: 1}
				case in(u, 0.9, 1.3):
					return OwlMod{Lids: LidsOpen}
				case in(u, 1.3, 1.5):
					return OwlMod{Lids: LidsHalf}
				}
				return OwlMod{}
			}},
		{Name: "peek", Note: "one eye opens, looks the other way, closes: 1.2 s in 10", Period: 10,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.5):
					return OwlMod{Lids: LidsOpen, OneEye: true}
				case in(u, 0.5, 1.0):
					return OwlMod{Lids: LidsOpen, Look: 1, OneEye: true}
				case in(u, 1.0, 1.2):
					return OwlMod{Lids: LidsHalf, OneEye: true}
				}
				return OwlMod{}
			}},
		{Name: "slow wake", Note: "half-open, open with a look each way, half, shut: 2.4 s in 14", Period: 14,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.4):
					return OwlMod{Lids: LidsHalf}
				case in(u, 0.4, 1.0):
					return OwlMod{Lids: LidsOpen, Look: 1}
				case in(u, 1.0, 1.6):
					return OwlMod{Lids: LidsOpen}
				case in(u, 1.6, 2.0):
					return OwlMod{Lids: LidsHalf}
				case in(u, 2.0, 2.4):
					return OwlMod{Lids: LidsDown, DY: -2}
				}
				return OwlMod{}
			}},
	},
	companion.Working: {
		{Name: "today", Note: "open, one blink every 7 s", Period: 8},
		{Name: "look away", Note: "the pupils go the other way for 1.2 s every 9 s", Period: 9,
			At: func(u float64) OwlMod {
				if in(u, 0, 1.2) {
					return OwlMod{Look: 1}
				}
				return OwlMod{}
			}},
		{Name: "bob", Note: "the body lifts half a cell twice, the face stays: two bobs in 0.8 s every 8 s", Period: 8,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.2) || in(u, 0.4, 0.6) {
					return OwlMod{DY: 2}
				}
				return OwlMod{}
			}},
		{Name: "hop", Note: "small wings out, one row up, settles: 0.6 s every 10 s", Period: 10, Round: 1,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.15):
					return OwlMod{Wings: 3}
				case in(u, 0.15, 0.45):
					return OwlMod{Wings: 3, Hop: 1}
				case in(u, 0.45, 0.6):
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
		{Name: "stretch", Note: "small wings up for half a second, out, folded: 0.9 s every 12 s", Period: 12, Round: 1,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.5):
					return OwlMod{Wings: 4}
				case in(u, 0.5, 0.9):
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
		{Name: "flutter and look", Note: "the wings beat in place for half a second, then it looks the other way for a second: 1.5 s every 9 s", Period: 9, Round: 2,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.5):
					return OwlMod{Wings: flutter(u, false)}
				case in(u, 0.5, 1.5):
					return OwlMod{Look: 1}
				}
				return OwlMod{}
			}},
		{Name: "hop and flutter", Note: "up a row and the wings beat for most of a second before it lands and folds: 1.2 s every 10 s", Period: 10, Round: 2,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.15):
					return OwlMod{Wings: 3}
				case in(u, 0.15, 1.0):
					return OwlMod{Wings: flutter(u, false), Hop: 1}
				case in(u, 1.0, 1.2):
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
		{Name: "looks, blink, flutter", Note: "looks the other way, looks back, blinks, then the hop and flutter on both wings: 3.2 s every 10 s", Period: 10, Round: 3,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.7):
					return OwlMod{Look: 1}
				case in(u, 0.7, 1.3):
					return OwlMod{}
				case in(u, 1.3, 1.5):
					return OwlMod{Lids: LidsDown}
				case in(u, 2.0, 2.15):
					return OwlMod{Wings: 3}
				case in(u, 2.15, 3.0):
					return OwlMod{Wings: flutter(u, false), Hop: 1}
				case in(u, 3.0, 3.2):
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
		{Name: "long hop", Note: "the hop-and-flutter with a look the other way while it is up: 1.4 s every 10 s", Period: 10, Round: 2,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.15):
					return OwlMod{Wings: 3}
				case in(u, 0.15, 0.5):
					return OwlMod{Wings: flutter(u, false), Hop: 1}
				case in(u, 0.5, 1.1):
					return OwlMod{Wings: flutter(u, false), Hop: 1, Look: 1}
				case in(u, 1.1, 1.4):
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
	},
	companion.NeedsYou: {
		{Name: "today", Note: "wide, still", Period: 4},
		{Name: "blink", Note: "wide, one blink every 3 s", Period: 3,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.2) {
					return OwlMod{Lids: LidsDown}
				}
				return OwlMod{}
			}},
		{Name: "double blink", Note: "wide, two quick blinks every 4 s, as if to get your attention", Period: 4,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.15) || in(u, 0.3, 0.45) {
					return OwlMod{Lids: LidsDown}
				}
				return OwlMod{}
			}},
		{Name: "double blink, wave", Note: "the double blink every 4 s, and every other time the near wing waves with it", Period: 8, Round: 3,
			At: func(u float64) OwlMod {
				m := OwlMod{}
				if in(u, 0, 0.15) || in(u, 0.3, 0.45) || in(u, 4, 4.15) || in(u, 4.3, 4.45) {
					m.Lids = LidsDown
				}
				if in(u, 4, 5.2) {
					m.Wings = 5
					if int(u*5)%2 == 1 {
						m.Wings = 6
					}
				}
				return m
			}},
		{Name: "double blink, hop", Note: "the double blink, then a hop with the wings out", Period: 4,
			At: func(u float64) OwlMod {
				switch {
				case in(u, 0, 0.15) || in(u, 0.3, 0.45):
					return OwlMod{Lids: LidsDown}
				case in(u, 0.6, 0.75):
					return OwlMod{Wings: 1}
				case in(u, 0.75, 1.0):
					return OwlMod{Wings: 1, Hop: 1}
				case in(u, 1.0, 1.15):
					return OwlMod{Wings: 1}
				}
				return OwlMod{}
			}},
	},
	companion.Done: {
		{Name: "today", Note: "^ ^, still", Period: 5},
		{Name: "bounce", Note: "two hops, one row each, every 5 s", Period: 5,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.25) || in(u, 0.45, 0.7) {
					return OwlMod{Hop: 1}
				}
				return OwlMod{}
			}},
		{Name: "flap", Note: "wings out twice, every 5 s", Period: 5,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.25) || in(u, 0.45, 0.7) {
					return OwlMod{Wings: 1}
				}
				return OwlMod{}
			}},
		{Name: "wiggle", Note: "the body shimmies half a cell each way, the face stays: 0.8 s every 5 s", Period: 5,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.8) {
					if int(u*10)%2 == 0 {
						return OwlMod{DX: 1}
					}
					return OwlMod{DX: -1}
				}
				return OwlMod{}
			}},
		{Name: "flap and bounce", Note: "the small wings out with each of two hops", Period: 5,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.25) || in(u, 0.45, 0.7) {
					return OwlMod{Hop: 1, Wings: 3}
				}
				if in(u, 0.25, 0.35) || in(u, 0.7, 0.8) {
					return OwlMod{Wings: 3}
				}
				return OwlMod{}
			}},
	},
	companion.Worried: {
		{Name: "today", Note: "amber slits, still", Period: 6},
		{Name: "dart", Note: "the slit pupils flick the other way for 0.4 s every 3 s", Period: 3, Round: 1,
			At: func(u float64) OwlMod {
				if in(u, 0, 0.4) {
					return OwlMod{Look: 1}
				}
				return OwlMod{}
			}},
		{Name: "shiver", Note: "the body shakes half a cell each way for a second, every 6 s", Period: 6, Round: 1,
			At: func(u float64) OwlMod {
				if in(u, 0, 1.0) {
					if int(u*12)%2 == 0 {
						return OwlMod{DX: 1}
					}
					return OwlMod{DX: -1}
				}
				return OwlMod{}
			}},
		{Name: "ruffled", Note: "the wings held out the whole time: a wider owl, a shape a screenshot keeps; plus the dart", Period: 3, Round: 1,
			At: func(u float64) OwlMod {
				m := OwlMod{Wings: 1}
				if in(u, 0, 0.4) {
					m.Look = 1
				}
				return m
			}},
		{Name: "hunched", Note: "the body sits half a cell lower the whole time, the face stays; a slow blink every 6 s", Period: 6, Round: 1,
			At: func(u float64) OwlMod {
				m := OwlMod{DY: -2}
				if in(u, 0, 0.5) {
					m.Lids = LidsDown
				}
				return m
			}},
		{Name: "squint", Note: "the eyes narrow to a white band a cell tall with the pupil in it, held; the pupils dart every 3 s", Period: 3, Round: 2,
			At: func(u float64) OwlMod {
				m := OwlMod{Lids: LidsSquint}
				if in(u, 0, 0.4) {
					m.Look = 1
				}
				return m
			}},
		{Name: "squint, brows", Note: "the squint with worried brows raised toward the middle, held; the dart", Period: 3, Round: 2,
			At: func(u float64) OwlMod {
				m := OwlMod{Lids: LidsSquint, Brows: true}
				if in(u, 0, 0.4) {
					m.Look = 1
				}
				return m
			}},
		{Name: "squint, one wing waves", Note: "the squint, and the near wing waves three beats every 6 s", Period: 6, Round: 2,
			At: func(u float64) OwlMod {
				m := OwlMod{Lids: LidsSquint}
				if in(u, 0, 1.2) {
					m.Wings = 5
					if int(u*5)%2 == 1 {
						m.Wings = 6
					}
				}
				return m
			}},
		{Name: "squint, brows, wave", Note: "brows and the one-wing wave together", Period: 6, Round: 2,
			At: func(u float64) OwlMod {
				m := OwlMod{Lids: LidsSquint, Brows: true}
				if in(u, 0, 1.2) {
					m.Wings = 5
					if int(u*5)%2 == 1 {
						m.Wings = 6
					}
				}
				return m
			}},
		{Name: "squint, shiver", Note: "the squint with the body's half-cell shiver for a second every 6 s", Period: 6, Round: 2,
			At: func(u float64) OwlMod {
				m := OwlMod{Lids: LidsSquint}
				if in(u, 0, 1.0) {
					m.DX = 1
					if int(u*12)%2 == 1 {
						m.DX = -1
					}
				}
				return m
			}},
	},
}

// OwlMotionPick is the candidate the live vista draws per state; absent or
// -1 is today's owl. HIS PICKS, 2026-09-16: resting = peek (R2) · needs you =
// double blink (N2) · done = bounce (D1), then D4 an hour later.
// Second round, same day: working = hop and flutter (W6, both wings) · done
// = flap and bounce (D4) · worried = squint with brows (X6).
// Third round: working = looks, blink, flutter (index 7; the record calls
// it W8 and the pick was written as 8, which is "long hop": found by a
// parallel session's audit, and TestThePicksAreByName holds every pick by
// name now) · needs you = double blink with a one-wing wave every other
// time (N3).
// After his first look live, the afternoon of the same day: "small wings
// for both", so done's flap (D4) beats the working owl's small wings and
// no picked motion needs the 14-column art (TestThePickedOwlNeverGrows).
var OwlMotionPick = map[companion.State]int{
	companion.Resting:  2,
	companion.Working:  7, // "looks, blink, flutter"; 8 was "long hop", an off-by-one a parallel audit caught 2026-09-16
	companion.NeedsYou: 3,
	companion.Done:     4,
	companion.Worried:  6,
}

// OwlMotionPeriodOverride, when positive, replaces every motion's period:
// a review page shows an event every few seconds instead of every twelve.
// Zero live.
var OwlMotionPeriodOverride = 0.0

// PickedOwlMotion is the candidate picked for a state, or nil for today's.
func PickedOwlMotion(st companion.State) *OwlMotion {
	i, ok := OwlMotionPick[st]
	if !ok || i <= 0 || i >= len(OwlMotions[st]) {
		return nil
	}
	return &OwlMotions[st][i]
}

// PickedOwletMotion is the litter's picked candidate, or nil for today's.
func PickedOwletMotion() *OwletMotion {
	if OwletMotionPick <= 0 || OwletMotionPick >= len(OwletMotions) {
		return nil
	}
	return &OwletMotions[OwletMotionPick]
}

// stateLids is the state's own face.
func stateLids(st companion.State) Lids {
	switch st {
	case companion.Resting:
		return LidsDown
	case companion.Done:
		return LidsUp
	case companion.Worried:
		return LidsSlit
	}
	return LidsOpen
}

// The wings. Kinds 1 and 2 are overlays on a 14-wide art whose body sits
// in columns 1..12: the big wings of the first round, which no picked motion
// uses since his "small wings for both" (2026-09-16); they stay for the
// candidates he passed over. Kinds 3 to 6 are his
// "wings should be smaller": overlays on the body's own 12 columns, in the
// margins the barn owl's art leaves at columns 0-1 and 10-11, so the owl's
// box never grows. 5 and 6 are the near wing alone, for a wave.
var owlWings = [7][]string{
	nil,
	{ // 1 out, big
		"..............",
		"..............",
		"vv..........vv",
		"###........###",
		"^^^........^^^",
		"..............",
		"..............",
	},
	{ // 2 up, big
		"#............#",
		"#............#",
		"##..........##",
		".##........##.",
		"..............",
		"..............",
		"..............",
	},
	// ⚠ The first cut put the right wing at columns 8-9, INSIDE the body,
	// where the overlay rule (only over an empty cell) painted nothing: he
	// saw a one-winged flutter and asked for "BOTH wings". The margins are
	// columns 0-1 and 10-11; TestTheSmallWingsAreTwo holds them there.
	{ // 3 small out
		"............",
		"............",
		".v........v.",
		"##........##",
		".^........^.",
		"............",
		"............",
	},
	{ // 4 small up
		"............",
		".v........v.",
		"##........##",
		".^........^.",
		"............",
		"............",
		"............",
	},
	{ // 5 one wing, small out
		"............",
		"............",
		".v..........",
		"##..........",
		".^..........",
		"............",
		"............",
	},
	{ // 6 one wing, small up
		"............",
		".v..........",
		"##..........",
		".^..........",
		"............",
		"............",
		"............",
	},
}

// wingsWide says whether a wing kind needs the 14-column art.
func wingsWide(k int) bool { return k == 1 || k == 2 }

// expandArt turns cell art of any size into a bitmap, two pixels a column
// and four a row, the marks of OwlAlt.art.
func expandArt(art []string, w, h int) []string {
	px := make([][]byte, 4*h)
	for y := range px {
		px[y] = []byte(strings.Repeat(".", 2*w))
	}
	for cy, row := range art {
		for cx, m := range row {
			x0, y0 := 2*cx, 4*cy
			set := func(dx0, dx1, dy0, dy1 int) {
				for y := y0 + dy0; y <= y0+dy1; y++ {
					for x := x0 + dx0; x <= x0+dx1; x++ {
						px[y][x] = '#'
					}
				}
			}
			switch m {
			case '#':
				set(0, 1, 0, 3)
			case 'v':
				set(0, 1, 2, 3)
			case '^':
				set(0, 1, 0, 1)
			case '<':
				set(1, 1, 0, 3)
			case '>':
				set(0, 0, 0, 3)
			}
		}
	}
	out := make([]string, len(px))
	for y := range px {
		out[y] = string(px[y])
	}
	return out
}

// shiftPx moves a bitmap dx pixels right and dy pixels up, clipping.
func shiftPx(bm []string, dx, dy int) []string {
	h, w := len(bm), len(bm[0])
	out := make([]string, h)
	for y := 0; y < h; y++ {
		row := make([]byte, w)
		for x := 0; x < w; x++ {
			row[x] = '.'
			sx, sy := x-dx, y+dy
			if sx >= 0 && sx < w && sy >= 0 && sy < h {
				row[x] = bm[sy][sx]
			}
		}
		out[y] = string(row)
	}
	return out
}

// DrawOwlMoving draws the owl in a state with a motion laid over it. A nil
// motion, or one whose At is nil, draws today's owl exactly. period, when
// positive, replaces the motion's own so a page can show an event sooner
// than it comes live.
func DrawOwlMoving(c *canvas.Canvas, x, y int, t float64, st companion.State, m *OwlMotion, period float64) {
	var mod OwlMod
	if m != nil && m.At != nil {
		p := m.Period
		if period > 0 {
			p = period
		}
		mod = m.At(math.Mod(t, p))
	}
	// Today's working blink keeps its own clock unless the motion has the lids.
	if st == companion.Working && mod.Lids == LidsState && math.Mod(t, 7) < 0.25 {
		mod.Lids = LidsDown
	}
	drawOwlWith(c, x, y, st, mod)
}

// drawOwlWith is DrawOwlPose with a modification: the same body, the same
// eye cells, every colour a cube entry, drawn once rather than drawn open
// and painted over.
func drawOwlWith(c *canvas.Canvas, x, y int, st companion.State, mod OwlMod) {
	pick := OwlPick
	if pick < 0 || pick >= len(OwlAlts) {
		pick = 0
	}
	alt := &OwlAlts[pick]
	near := c.Near()
	y -= mod.Hop

	// The body. Padded to 14 columns only for the big wings, so the folded
	// owl and the small-winged owl are the same cells as DrawOwl's box.
	art, light, ox, w := alt.art, alt.light, x, 12
	if mod.Wings > 0 && mod.Wings < len(owlWings) {
		pad := ""
		if wingsWide(mod.Wings) {
			w, ox, pad = 14, x-1, "."
		}
		art = make([]string, 7)
		light = make([]string, 7)
		for r := range art {
			b := []byte(pad + alt.art[r] + pad)
			for k, m := range owlWings[mod.Wings][r] {
				if m != '.' && b[k] == '.' {
					b[k] = byte(m)
				}
			}
			art[r] = string(b)
			light[r] = pad + alt.light[r] + pad
		}
	}
	bm := expandArt(art, w, 7)
	if mod.DX != 0 || mod.DY != 0 {
		bm = shiftPx(bm, mod.DX, mod.DY)
	}
	q := mustBitmap(bm, 2*w, 28, alt.Name).ToQuadrant()
	companion.PlotRim(near, q, ox, y)
	for dy, row := range q {
		for dx, r := range []rune(row) {
			switch r {
			case ' ':
				continue
			case '█':
				g := OwlCoat
				if light[dy][dx] == 'o' {
					g = owlLight
				}
				near.PlotOn(ox+dx, y+dy, ' ', g, g, 1)
			default:
				near.PlotOn(ox+dx, y+dy, r, OwlCoat, c.BGAt(ox+dx, y+dy), 1)
			}
		}
	}

	// The eyes.
	wide := st == companion.NeedsYou
	for i, ex := range [2]int{alt.eyeL, alt.eyeR} {
		lids, look := stateLids(st), mod.Look
		if mod.Lids != LidsState && (!mod.OneEye || i == 0) {
			lids = mod.Lids
		}
		if mod.OneEye && i == 1 {
			look = 0
		}
		x0, ew, pupil := ex, alt.eyeW, alt.pupil
		if wide {
			ew, pupil = alt.eyeW+1, 1
			if i == 0 {
				x0 = ex - 1
			}
		}
		pc := min(max(pupil+look, 0), ew-1)
		for r := 0; r < alt.eyeH; r++ {
			for k := 0; k < ew; k++ {
				cx, cy := x+x0+k, y+alt.eyeRow+r
				switch lids {
				case LidsDown, LidsUp:
					g := ' '
					if r == 0 {
						g = '-'
						if lids == LidsUp {
							g = '^'
						}
					}
					near.PlotOn(cx, cy, g, owlDark, owlLight, 1)
				case LidsSlit:
					if r == 0 {
						near.PlotOn(cx, cy, ' ', owlLight, owlLight, 1)
					} else {
						col := owlBeak
						if k == pc {
							col = owlDark
						}
						near.PlotOn(cx, cy, ' ', col, col, 1)
					}
				case LidsSquint:
					// A band a cell tall across the cell boundary: the lower
					// half of the top row and the upper half of the bottom row
					// are eye, the rest is lid in the face's colour. Both rows
					// use U+2584 alone (term.LowerHalf's reason), the colours
					// swapped for the bottom row.
					eye := owlWhite
					if k == pc {
						eye = owlDark
					}
					if r == 0 {
						near.PlotOn(cx, cy, '\u2584', eye, owlLight, 1)
					} else {
						near.PlotOn(cx, cy, '\u2584', owlLight, eye, 1)
					}
				case LidsHalf:
					if r == 0 {
						near.PlotOn(cx, cy, '-', owlDark, owlWhite, 1)
					} else {
						col := owlWhite
						if k == pc {
							col = owlDark
						}
						near.PlotOn(cx, cy, ' ', col, col, 1)
					}
				default:
					col := owlWhite
					if k == pc {
						col = owlDark
					}
					near.PlotOn(cx, cy, ' ', col, col, 1)
					if k == pc && r == 0 {
						near.PlotOn(cx, cy, '˙', owlWhite, owlDark, 1)
					}
				}
			}
		}
	}
	if mod.Brows && alt.eyeRow > 0 {
		// Raised toward the middle: the worried arch, over each eye's inner column.
		near.PlotOn(x+alt.eyeL+alt.eyeW-1, y+alt.eyeRow-1, '/', owlDark, owlLight, 1)
		near.PlotOn(x+alt.eyeR, y+alt.eyeRow-1, '\\', owlDark, owlLight, 1)
	}
	near.PlotOn(x+5, y+alt.beakRow, '\\', owlDark, owlBeak, 1)
	near.PlotOn(x+6, y+alt.beakRow, '/', owlDark, owlBeak, 1)
}

// THE OWLETS' MOTION. "lets expand on the animations": today they are
// still. Each candidate gives owlet k of n something to do on its own
// clock, staggered, so the litter never moves as one block.
type OwletMod struct {
	Blink bool
	Hop   int
	DX    int
	Cheep bool
}

type OwletMotion struct {
	Name, Note string
	Period     float64
	// PerOwlet, when set, stretches the period to at least PerOwlet*n so a
	// bigger litter is not a busier one.
	PerOwlet float64
	// At gives owlet k of n its frame: u is seconds into the period p, cyc
	// the number of periods elapsed.
	At func(u, p float64, cyc, k, n int) OwletMod
}

var OwletMotions = []OwletMotion{
	{Name: "today", Note: "still", Period: 4},
	{Name: "blink", Note: "each on its own clock, 0.2 s every 4 s", Period: 4,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			return OwletMod{Blink: in(u, 0.7*float64(k), 0.7*float64(k)+0.2)}
		}},
	{Name: "hop", Note: "one at a time, a row up for a quarter second", Period: 4,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			if in(u, 0.9*float64(k), 0.9*float64(k)+0.25) {
				return OwletMod{Hop: 1}
			}
			return OwletMod{}
		}},
	{Name: "cheep", Note: "a beak opens on one owlet at a time, with a blink after", Period: 4,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			a := 1.1 * float64(k)
			return OwletMod{Cheep: in(u, a, a+0.35), Blink: in(u, a+0.5, a+0.65)}
		}},
	{Name: "wiggle", Note: "one at a time shimmies half a cell each way", Period: 4,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			a := 1.0 * float64(k)
			if in(u, a, a+0.6) {
				if int((u-a)*10)%2 == 0 {
					return OwletMod{DX: 1}
				}
				return OwletMod{DX: -1}
			}
			return OwletMod{}
		}},
	{Name: "all of it", Note: "blink, hop and cheep, staggered", Period: 4,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			a := 1.2 * float64(k)
			m := OwletMod{Blink: in(u, a+0.9, a+1.05), Cheep: in(u, a+0.4, a+0.7)}
			if in(u, a, a+0.25) {
				m.Hop = 1
			}
			return m
		}},
	// THE LITTER'S LIFE: his "I like L1, L2, L3, L4 - all of it basically.
	// What's the best criteria to implement these." The criterion, four rules:
	//
	//  1. ONE OWLET AT A TIME. The period is cut into n windows and owlet k
	//     owns window k, so no two bodies move at once and the litter never
	//     moves as a block -- a block moving reads as an event.
	//  2. EVERY OWLET GETS ONE ACT A PERIOD, the acts rotating hop, cheep,
	//     wiggle by owlet and by period (k + cyc), so the same owlet does
	//     not do the same thing twice running and a screenshot at any two
	//     moments differs.
	//  3. THE LITTER IS STILL AT LEAST HALF THE TIME at any count: an act
	//     takes at most 0.6 s of a window, and the period stretches to
	//     PerOwlet*n once the litter is bigger than five, so the fraction of
	//     time with any body moving is capped at n*0.6/max(6, 1.2n) <= 50%.
	//  4. BLINKS ARE FREE: 0.2 s, once a period each, in the owlet's own
	//     window, because a blink moves no body and cannot be read as one.
	//
	// Nothing here carries information: the count is the channel, and the
	// count is what a glance still gets.
	{Name: "the litter's life", Note: "blink, hop, cheep and wiggle, one owlet at a time in its own window, the act rotating each period; still at least half the time at any count", Period: 6, PerOwlet: 1.2,
		At: func(u, p float64, cyc, k, n int) OwletMod {
			slot := p / float64(max(n, 1))
			a := slot * float64(k)
			m := OwletMod{Blink: in(u, a+0.05, a+0.25)}
			switch (k + cyc) % 3 {
			case 0:
				if in(u, a+0.3, a+0.55) {
					m.Hop = 1
				}
			case 1:
				m.Cheep = in(u, a+0.3, a+0.65)
				if in(u, a+0.75, a+0.9) {
					m.Blink = true
				}
			case 2:
				if in(u, a+0.3, a+0.9) {
					m.DX = 1
					if int((u-a)*10)%2 == 1 {
						m.DX = -1
					}
				}
			}
			return m
		}},
}

// OwletMotionPick is the candidate the live vista draws; -1 is today's.
// HIS PICK, 2026-09-16: all four acts, under the criterion above (L6).
var OwletMotionPick = 6

// DrawOwletMoving is DrawOwlet with a modification; a zero OwletMod draws
// today's owlet exactly.
func DrawOwletMoving(c *canvas.Canvas, coat term.RGB, x, y int, faceLeft bool, mod OwletMod) {
	st := &OwletStyles[0]
	if OwletPick >= 0 && OwletPick < len(OwletStyles) {
		st = &OwletStyles[OwletPick]
	}
	y -= mod.Hop
	var rows []string
	if st.art == nil {
		rows = owlet
		if faceLeft {
			rows = mirrorRows(rows)
		}
	} else {
		rows = expandSmall(st.art)
	}
	if mod.DX != 0 {
		rows = shiftPx(rows, mod.DX, 0)
	}
	q := mustBitmap(rows, 12, 16, st.Name).ToQuadrant()
	near := c.Near()
	companion.PlotRim(near, q, x, y)
	for dy, row := range q {
		for dx, r := range []rune(row) {
			switch r {
			case ' ':
				continue
			case '█':
				g := coat
				if st.light != nil && st.light[dy][dx] == 'o' {
					g = owlLight
				}
				near.PlotOn(x+dx, y+dy, ' ', g, g, 1)
			default:
				near.PlotOn(x+dx, y+dy, r, coat, c.BGAt(x+dx, y+dy), 1)
			}
		}
	}
	if mod.Blink {
		near.PlotOn(x+1, y+1, '-', owlDark, owlLight, 1)
		near.PlotOn(x+4, y+1, '-', owlDark, owlLight, 1)
	} else {
		near.PlotOn(x+1, y+1, '•', owlDark, owlWhite, 1)
		near.PlotOn(x+4, y+1, '•', owlDark, owlWhite, 1)
	}
	if st.beak || mod.Cheep {
		g := 'v'
		if mod.Cheep {
			g = 'V'
		}
		near.PlotOn(x+2, y+2, g, owlDark, owlBeak, 1)
	}
}

// mirrorRows flips a bitmap's rows, for the study owlet's facing.
func mirrorRows(rows []string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		b := []byte(r)
		for a, z := 0, len(b)-1; a < z; a, z = a+1, z-1 {
			b[a], b[z] = b[z], b[a]
		}
		out[i] = string(b)
	}
	return out
}

// DrawOwletsMoving is DrawOwlets with a motion; nil draws today's litter.
func DrawOwletsMoving(c *canvas.Canvas, x, grassY, n, minX int, t float64, m *OwletMotion, period float64) int {
	drawn := 0
	for k := 1; k <= n; k++ {
		ox := x - OwletW*k
		if ox < minX || ox < 0 {
			break
		}
		var mod OwletMod
		if m != nil && m.At != nil {
			p := m.Period
			if period > 0 {
				p = period
			}
			if m.PerOwlet > 0 {
				p = math.Max(p, m.PerOwlet*float64(n))
			}
			mod = m.At(math.Mod(t, p), p, int(t/p), k-1, n)
		}
		DrawOwletMoving(c, OwlCoat, ox, grassY, k%2 == 1, mod)
		drawn++
	}
	return drawn
}

// WHERE THE LITTER SITS, AND HOW IT COMES AND GOES (s36). His question:
// "Could owlets fly around? Maybe gather around the fire?"
//
// Not around: an owlet in the air is motion, and the wind is the vista's
// motion channel; a litter wheeling over the meadow is a count nobody can
// take at a glance and a busy sky that reads as work. So the flight is the
// EVENT and not the state: an owlet flies in from the owl when its
// subagent starts, and flies back to the owl when the reducer lets it go.
// While it is here it sits, with the litter's life (L6).
//
// Where it sits is OwletPlace: 0 beside the owl, as today; 1 around the
// fire, on both sides of it, nearest first, facing it -- the picture he
// asked to see; then "explore both placements, simultaneously": 2 the
// nest first (two beside the owl) and the rest around the fire; 3 by
// turns, one beside the owl, one at the fire.
var OwletPlace = 3 // HIS PICK, 2026-09-16: by turns

// OwletPlaces names them for a page.
var OwletPlaces = []struct{ Name, Note string }{
	{"beside the owl", "as today: nearest first, leftward, stopping short of the fire"},
	{"around the fire", "both sides of it, nearest first, facing it"},
	{"the nest, then the fire", "the first two beside the owl, every owlet after that around the fire"},
	{"by turns", "one beside the owl, one at the fire, one beside the owl"},
}

// OwletFlyIn is the least an arrival takes; a farther spot takes longer
// (flyTime), so the fire is not reached at a sprint. The departure takes
// the reducer's own KittenExit, whose progress it hands us.
const OwletFlyIn = 1.0

func flyTime(dist int) float64 { return OwletFlyIn + float64(dist)/50 }

// OwletFlock remembers when each sitting owlet arrived, so the flight in
// can be drawn from a count that only ever says how many.
type OwletFlock struct {
	arrived []float64
}

// Sync fits the flock to the count at time t: newcomers arrive now, the
// departed are dropped from the far end. An owlet that was there before
// the flock existed is taken as long arrived.
func (f *OwletFlock) Sync(n int, t float64) {
	for len(f.arrived) < n {
		if f.arrived == nil {
			f.arrived = make([]float64, 0, n)
			for len(f.arrived) < n {
				f.arrived = append(f.arrived, t-flyTime(200))
			}
			break
		}
		f.arrived = append(f.arrived, t)
	}
	if len(f.arrived) > n {
		f.arrived = f.arrived[:n]
	}
}

// OwletSpots is where owlet k of a litter sits for a placement: beside
// the owl at x (nearest first, stopping at minX), or around the fire at
// fireX, alternating sides, nearest first. ok is false when there is no
// room, and the count then shows fewer rather than piling them.
func OwletSpot(place, k int, owlX, fireX, w, minX int) (x int, faceLeft, ok bool) {
	nest := func(k int) (int, bool, bool) {
		x := owlX - OwletW*(k+1)
		return x, k%2 == 0, x >= minX && x >= 0
	}
	fire := func(k int) (int, bool, bool) {
		side, ring := k%2, k/2
		if side == 0 {
			x := fireX - 9 - OwletW*ring
			return x, false, x >= 1
		}
		x := fireX + 6 + OwletW*ring
		return x, true, x+6 <= owlX-3
	}
	switch place {
	case 1:
		return fire(k)
	case 2:
		if k < 2 {
			return nest(k)
		}
		return fire(k - 2)
	case 3:
		if k%2 == 0 {
			return nest(k / 2)
		}
		return fire(k / 2)
	}
	return nest(k)
}

// flightAt is where an owlet is on its way between the owl and its spot,
// u 0 at the owl and 1 at the spot, in three legs: straight up behind the
// owl to the cruising row (the first fifth), across at that row (three
// fifths), straight down onto the spot (the last fifth). The cruising row
// is the owl's own top row, five above the grass, so a flight clears
// every sitting owlet by a row; the first cut flew a low arc and crossed
// the nest at the litter's own height (his note, 2026-09-16 11:55).
func flightAt(u float64, fromX, fromY, toX, toY, cruiseY int) (x, y int) {
	switch {
	case u < 0.2:
		return fromX, fromY + int(math.Round(u/0.2*float64(cruiseY-fromY)))
	case u < 0.8:
		return fromX + int(math.Round((u-0.2)/0.6*float64(toX-fromX))), cruiseY
	}
	return toX, cruiseY + int(math.Round((u-0.8)/0.2*float64(toY-cruiseY)))
}

// DrawOwletFlying draws an owlet in the air: the winged art, the wings
// beating four times a second.
func DrawOwletFlying(c *canvas.Canvas, coat term.RGB, x, y int, t float64) {
	st := &OwletStyles[4] // wings up
	rows := expandSmall(st.art)
	if int(t*8)%2 == 1 {
		rows = expandSmall([]string{".^..^.", "######", "######", ".^..^."}) // wings down: the chick
	}
	q := mustBitmap(rows, 12, 16, st.Name).ToQuadrant()
	near := c.Near()
	companion.PlotRim(near, q, x, y)
	for dy, row := range q {
		for dx, r := range []rune(row) {
			switch r {
			case ' ':
				continue
			case '█':
				g := coat
				if st.light[dy][dx] == 'o' {
					g = owlLight
				}
				near.PlotOn(x+dx, y+dy, ' ', g, g, 1)
			default:
				near.PlotOn(x+dx, y+dy, r, coat, c.BGAt(x+dx, y+dy), 1)
			}
		}
	}
	near.PlotOn(x+1, y+1, '•', owlDark, owlWhite, 1)
	near.PlotOn(x+4, y+1, '•', owlDark, owlWhite, 1)
}

// DrawFlock draws the litter for a placement: n sitting (each in flight
// from the owl for its first flyTime seconds), and one owlet flying back
// to the owl per exit in progress. Returns how many sit on screen.
//
// It is drawn BEFORE the owl, and a flight starts and ends INSIDE the
// owl's box, so an owlet comes out from behind its parent and goes back
// behind it; drawn after, the first build painted the newcomer over the
// owl's face (his screenshot, 2026-09-16 11:49).
func DrawFlock(c *canvas.Canvas, f *OwletFlock, n int, exits []float64, owlX, owlY, grassY, fireX, minX int, t float64, m *OwletMotion, period float64) int {
	f.Sync(n, t)
	fromX, fromY, cruiseY := owlX+3, owlY+3, grassY-5
	drawn := 0
	type flight struct{ x, y int }
	var flying []flight
	for k := 0; k < n; k++ {
		x, faceLeft, ok := OwletSpot(OwletPlace, k, owlX, fireX, c.W, minX)
		if !ok {
			break
		}
		if age, fly := t-f.arrived[k], flyTime(fromX-x); age < fly {
			fx, fy := flightAt(age/fly, fromX, fromY, x, grassY, cruiseY)
			flying = append(flying, flight{fx, fy})
			continue
		}
		var mod OwletMod
		if m != nil && m.At != nil {
			p := m.Period
			if period > 0 {
				p = period
			}
			if m.PerOwlet > 0 {
				p = math.Max(p, m.PerOwlet*float64(n))
			}
			mod = m.At(math.Mod(t, p), p, int(t/p), k, n)
		}
		DrawOwletMoving(c, OwlCoat, x, grassY, faceLeft, mod)
		drawn++
	}
	for i, p := range exits {
		x, _, ok := OwletSpot(OwletPlace, n+i, owlX, fireX, c.W, minX)
		if !ok || p >= 1 {
			continue
		}
		fx, fy := flightAt(1-p, fromX, fromY, x, grassY, cruiseY)
		flying = append(flying, flight{fx, fy})
	}
	// The flyers over the sitters, and the owl over all of them (drawVista).
	for _, fl := range flying {
		DrawOwletFlying(c, OwlCoat, fl.x, fl.y, t)
	}
	return drawn
}

// FlightBox is the cells an arriving owlet k of n occupies at u along its
// flight for a placement, for a test to sweep against the litter.
func FlightBox(place, k int, u float64, owlX, owlY, grassY, fireX, w, minX int) (x0, y0, x1, y1 int, ok bool) {
	x, _, ok := OwletSpot(place, k, owlX, fireX, w, minX)
	if !ok {
		return 0, 0, 0, 0, false
	}
	fx, fy := flightAt(u, owlX+3, owlY+3, x, grassY, grassY-5)
	return fx, fy, fx + 6, fy + 4, true
}
