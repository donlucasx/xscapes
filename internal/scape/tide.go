package scape

import (
	"math"

	"github.com/donlucasx/xscapes/internal/envx"
)

// Tide is HIS idea, 2026-09-09, and SHIPPED ON at his word the next day: "turn
// XSCAPES_TIDE=1 ON by default." He asked for it as "the tide could be moving
// in and out of the land, instead of moving towards the left of the screen.
// Like a real tide", ran it for an evening and kept it -- "new tide seems to
// work well ... I like it".
//
// He is right about both halves, and both are measured.
//
// WHAT MOVES SIDEWAYS: the waterline is a sine in x whose phase advances with
// time -- reach*sin(fx*0.11 + tt*0.8) -- so the shape of the coast slides
// across the frame. Water does not do that. A wave travels toward the shore;
// the shoreline itself runs up the sand and back.
//
// WHAT DOES NOT MOVE AT ALL: the sea never comes up the beach. The waterline's
// base row is `sy := c.H - beach`, which is a function of the canvas height and
// the writing band and NOTHING else. Rendered and averaged over eighty phases,
// the mean waterline sits at row 18.75 at level 0.00 and row 18.75 at level
// 1.00, moving 0.09 of a row across the entire activity range. All that grows
// with the work is the wiggle: three rows of swing at rest, four flat out.
//
// So this does two things. The sheet runs in and out TOGETHER, one phase for
// every column, with the coast's own shape holding nearly still underneath it.
// And how far up the beach the water sits becomes the activity channel, which
// is the fix the sea needs anyway: a position, countable at a glance, and it
// currently carries nothing.
//
// ⚠ THE DIRECTION IS FORCED, and it is worth knowing before looking. There is
// almost no room to come further IN -- the writing band is right below the
// waterline and `sy` is already clamped against it -- and a lot of room to go
// OUT, eight rows at his geometry. So the busy position is where the water
// already sits, and REST pulls it back up the frame toward the horizon. Idle
// reads as low tide with a wide beach, and that is the honest way round.
//
// XSCAPES_TIDE=0 puts the old waterline back -- a fixed row with a sine sliding
// across it -- which is the picture every frame before 2026-09-10 was reviewed
// against, and what the A/B in notes/searange and notes/drift compares to.
var Tide = true

func init() {
	switch envx.Lookup("TIDE") {
	case "0", "off", "no":
		Tide = false
	}
}

// TideRange is how many rows, at scale 1, the water withdraws when the agent
// goes quiet. Clamped afterwards by the same guards that keep the waterline
// off the horizon and out of the writing.
var TideRange = 5.0

// TideEase is the time constant the water takes to reach where the activity
// puts it. Seconds.
var TideEase = 3.0

// tideTarget is how far the waterline should have withdrawn at this level.
func tideTarget(level, scale float64) float64 {
	if !Tide {
		return 0
	}
	return (1 - clamp01(level)) * TideRange * scale
}

// hyFloor is the highest row the water may withdraw to, so a long quiet stretch
// cannot pull the shoreline up into the sky.
func hyFloor(sy int) float64 { return float64(sy) - TideRange - 1 }

// tideEdge is the waterline under Tide: one phase for every column, so the
// whole sheet advances and retreats together, over a coast whose own shape
// drifts thirteen times slower than the shipped one and so reads as a place
// rather than as a pattern going past.
func tideEdge(w, sy int, floor, at, tt float64, act Activity, scale float64) []float64 {
	wash := (0.7 + act.Level*0.9) * scale * math.Sin(tt*0.5)
	base := float64(sy) - at
	if base < floor {
		base = floor
	}
	e := make([]float64, w)
	for x := 0; x < w; x++ {
		fx := float64(x)
		// No time in the x terms at all. Any of it, however slow, is the coast
		// itself sliding across the frame -- which is the thing he asked to be
		// rid of, and 0.06 was still several radians a minute. The shoreline is
		// a PLACE; all the motion belongs to the wash.
		e[x] = base + wash +
			0.60*scale*math.Sin(fx*0.11) +
			0.35*scale*math.Sin(fx*0.047)
	}
	return e
}
