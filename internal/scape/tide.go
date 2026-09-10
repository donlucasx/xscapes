package scape

import (
	"math"

	"github.com/donlucasx/xscapes/internal/envx"
)

// Tide is HIS idea, 2026-09-09, built behind a switch to be looked at rather
// than shipped: "the tide could be moving in and out of the land, instead of
// moving towards the left of the screen. Like a real tide."
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
//	XSCAPES_TIDE=1 xscapes claude
var Tide = envx.Lookup("TIDE") == "1"

// TideRange is how many rows, at scale 1, the water withdraws when the agent
// goes quiet. Clamped afterwards by the same guards that keep the waterline
// off the horizon and out of the writing.
var TideRange = 5.0

// tideOffset is how far the waterline pulls back from its busy position.
func tideOffset(level, scale float64) int {
	if !Tide {
		return 0
	}
	return int(math.Round((1 - clamp01(level)) * TideRange * scale))
}

// tideEdge is the waterline under Tide: one phase for every column, so the
// whole sheet advances and retreats together, over a coast whose own shape
// drifts thirteen times slower than the shipped one and so reads as a place
// rather than as a pattern going past.
func tideEdge(w, sy int, tt float64, act Activity, scale float64) []float64 {
	wash := (0.7 + act.Level*0.9) * scale * math.Sin(tt*0.5)
	e := make([]float64, w)
	for x := 0; x < w; x++ {
		fx := float64(x)
		e[x] = float64(sy) + wash +
			0.60*scale*math.Sin(fx*0.11+tt*0.06) +
			0.35*scale*math.Sin(fx*0.047-tt*0.03)
	}
	return e
}
