package scape

import "math"

// The arriving star -- "the fall", his ruling of 2026-09-12: *"lets go with
// 'the fall'"*, arm A3 of the study on assets/frames/s29-starfall.html.
//
// A finished todo used to light its star by simply existing on the next frame.
// The fall gives the arrival a motion: the star ITSELF travels six columns and
// three rows into its own place over half a second, linear, with no tail. The
// '*' that moves is the star -- not a streak flying into a star that is already
// lit, which was the other arm on the page and the one that broke nothing.
//
// ⚠ THIS IS THE ONE PLACE IN THE SCENE WHERE SOMETHING MOVES TO SAY SOMETHING,
// and it does not break the encoding rule. The rule forbids encoding a variable
// in a RATE, because a glance is the whole budget and a screenshot has no
// motion at all. The count is still in the star COUNT and every star still
// lights where it always was; the fall carries no information of its own. It is
// the same argument that lets the claw twitch and the rotating eye exist -- a
// channel that says nothing cannot collide with one that says something. A
// screenshot taken mid-fall shows one star short, which is the cost, and the
// star is on screen the whole time rather than missing.
const (
	// arrivalCols is the longest track, in columns. Six, his pick.
	arrivalCols = 6
	// arrivalRows is how far ABOVE its place the star launches, in rows. Three.
	arrivalRows = 3
	// arrivalEvery is how many columns the star travels per row of drop, so the
	// fall runs at about one row per two columns and reads as a fall rather
	// than as a slide.
	arrivalEvery = 2
	// arrivalMin is the shortest track worth flying, in cells. Below it the
	// star simply arrives in place: a two-cell fall at 20 fps is one frame of
	// motion, which reads as a flicker and not as an arrival. Measured over
	// 1,805 swept arrivals, 6.5% land in place at this bar.
	arrivalMin = 4
	// arrivalSep is how close the head may pass a lit star, in SCREEN units --
	// the same measure starPlaces keeps the constellation apart with, because a
	// terminal cell is about twice as tall as it is wide and two marks side by
	// side merge where two a row apart do not.
	arrivalSep = 2.0
)

// arrivalHead is which cell of the track the star occupies at this phase.
//
// LINEAR, his pick over the drafted ease-out. The criterion the study applied
// is that the head must never SKIP a cell -- a skipped cell at 20 fps is a
// teleport -- and an ease-out spends its first frames crossing several columns
// at once. Repeats are fine: a head that dwells a frame reads as a fall, a head
// that jumps reads as two stars.
func arrivalHead(n int, phase float64) int {
	if n < 2 {
		return 0
	}
	return int(math.Round(clamp01(phase) * float64(n-1)))
}

// arrivalTrack is the path the star falls down, launch first and its own place
// last, or nil if the sky will not give it one.
//
// IT WALKS OUTWARD FROM HOME AND FLIES THE RESULT IN REVERSE, and that is the
// whole design rather than an implementation detail. Searching for a launch
// point and walking in would let a guard trim the LANDING -- the one cell that
// has to be exact, because it is where the star lives for the rest of the
// session. Walking outward means everything a guard trims is the far end, and
// a launch off the canvas is unreachable rather than fixed. The drafted version
// searched inward and put between 8.7% and 33.6% of its launches off screen,
// depending on the geometry.
//
// THE SIDE IS CHOSEN AWAY FROM THE MOON, and the argument that this makes the
// disc and readout guards redundant is WRONG. It was written here first and
// measured afterwards, which is the wrong order; the numbers are below.
//
// Over the 30,132 swept arrivals of arrival_test.go, what stops the walk:
//
//	cap      13860  46.0%   the full six columns, nothing in the way
//	canvas   11070  36.7%   the top or the side of the sky
//	sep       3528  11.7%   passing too close to a lit star
//	onstar    1476   4.9%   a lit star's own cell
//	readout    108   0.36%  ground the context number can be painted on
//	disc        90   0.30%  the moon itself
//
// Both of the supposedly dead guards fire, and for different reasons. The DISC
// guard fires 90 times and every single one is on the FLIP -- the fallback that
// walks back TOWARD the moon when the outward side has no room for a track --
// so the structural argument holds for the chosen direction and the flip is
// exactly the hole in it. The READOUT guard fires 108 times on the PRIMARY
// direction with no flip involved, because readoutGround covers the label's
// alternate placement BESIDE the disc as well as the one under it, and that
// block sits nine columns out from the moon on the same side a star may be
// walking. Mutation-tested: switching either guard off turns the sweep red at
// 125x28, his own window.
//
// ⚠ So the lesson is the repo's own, paid for again: an argument that a guard
// cannot fire is worth nothing until something counts.
func (s *Shore) arrivalTrack(w, hy, top, bot, idx int, pts [][2]int) [][2]int {
	if idx < 0 || idx >= len(pts) {
		return nil
	}
	key := arrKey{star: s.starKey, idx: idx, set: true}
	if s.arrKey == key {
		// A REFUSED path is cached too, and that is the reason for the `set`
		// field rather than testing the slice for nil. At 40x12 most arrivals
		// are refused, and a nil result treated as "not computed yet" rebuilt
		// readoutGround on every frame of every one of them.
		return s.arrTrack
	}
	home := pts[idx]
	// The stars already on screen. The head may not land on one and may not
	// pass close enough to merge with one: either would read as the count
	// going wrong, which is the one thing this channel may not do.
	lit := pts[:idx]
	blocked := s.readoutGround(w, hy, top, bot)

	walk := func(dir int) []([2]int) {
		cells := [][2]int{home}
		x, y, dy := home[0], home[1], 0
		for k := 1; k <= arrivalCols; k++ {
			x += dir
			if k%arrivalEvery == 0 && dy < arrivalRows {
				y--
				dy++
			}
			switch {
			case x < 0 || x >= w || y < 0:
				return cells // the canvas edge
			case s.DiscCovers(x, y):
				return cells
			case blocked[[2]int{x, y}]:
				return cells // where the context readout can land
			case onLitStar(x, y, lit) || nearLitStar(x, y, lit):
				return cells
			}
			cells = append(cells, [2]int{x, y})
		}
		return cells
	}

	dir := 1
	if home[0] < s.moonX {
		dir = -1
	}
	cells := walk(dir)
	if len(cells) < arrivalMin {
		// Only if the outward side is too cramped to fly at all: a fall toward
		// the moon is worse than a fall on the near side, and better than none.
		// Measured, 2,106 of 30,132 swept arrivals flip -- 7.0% -- and it is the
		// flip that makes the disc guard above load-bearing rather than dead.
		if alt := walk(-dir); len(alt) > len(cells) {
			cells = alt
		}
	}
	var track [][2]int
	if len(cells) >= arrivalMin {
		track = make([][2]int, len(cells))
		for i, p := range cells {
			track[len(cells)-1-i] = p
		}
	}
	s.arrKey, s.arrTrack = key, track
	return track
}

func onLitStar(x, y int, lit [][2]int) bool {
	for _, q := range lit {
		if q[0] == x && q[1] == y {
			return true
		}
	}
	return false
}

// nearLitStar compares squared distances in screen units, the way starPlaces
// does and for the same reason: math.Hypot measured 175-481 us a frame in this
// loop, up to 45% of a whole frame's budget.
func nearLitStar(x, y int, lit [][2]int) bool {
	const bar = arrivalSep * arrivalSep
	for _, q := range lit {
		dx := float64(x - q[0])
		dy := 2 * float64(y-q[1])
		if dx*dx+dy*dy < bar {
			return true
		}
	}
	return false
}

// arrKey is the geometry and the slot a cached track was built for. The track
// is held for the whole fall rather than recomputed per frame, and not to save
// the work: readoutGround moves with the disc, which sinks as the context
// fills, so a path rebuilt every frame could shift under the star mid-flight
// and the fall would jump sideways.
type arrKey struct {
	star starKey
	idx  int
	// set distinguishes "computed, and the sky refused it" from "never asked",
	// so a refused path costs one walk and not one a frame.
	set bool
}
