// Package scape holds the scenes. A scape reads only an Activity — it never
// knows which agent is running or what a "tool call" is. That separation is
// what lets one renderer serve every adapter.
package scape

import "github.com/donlucasx/xscapes/internal/canvas"

// Activity is the agent's state, normalised. Level is 0 (idle) to 1 (flat out).
type Activity struct {
	Working bool
	Level   float64

	// ContextUsed is 0 for a fresh session and 1 when the window is full.
	// Deliberately phrased as "used" rather than "left" so the zero value --
	// what you get if nobody sets it -- means a fresh session, not an
	// exhausted one.
	ContextUsed float64

	// TimeOfDay runs 0 at midnight, .25 dawn, .5 noon, .75 dusk. Midnight is
	// the zero value on purpose: it is the look the scape was designed around,
	// so forgetting to set this yields the good palette rather than a broken one.
	TimeOfDay float64

	// TodoDone and TodoTotal are the agent's checklist. Zero total means there
	// is no list, which is not the same as a list with nothing done -- the sky
	// shows nothing in the first case and an unlit constellation in the second.
	TodoDone, TodoTotal int

	// Arriving says the newest finished todo's star is still in flight: it has
	// not reached its place yet, so the sky draws one fewer settled star and
	// the missing one is falling into it. See arrival.go.
	//
	// ⚠ IT IS A SEPARATE BOOL ON PURPOSE, and ArrivalPhase must never be read
	// without it. A bare phase would make the zero value -- what every caller
	// that has never heard of this feature passes -- mean "launching right
	// now", so every pinned frame in the repo would fire a fall at once.
	Arriving bool
	// ArrivalPhase is 0 at launch and 1 the instant the star lands, and it is
	// meaningless unless Arriving is set. The reducer drives it from WALL-CLOCK
	// age rather than integrating dt: Shore.Update clamps a frame gap over a
	// second, so an integrated fall would freeze mid-air on a stalled render
	// and a suspended laptop would wake with the star still falling.
	ArrivalPhase float64
}

type Scape interface {
	Name() string
	Update(c *canvas.Canvas, t float64, act Activity)
}

// Hash2 is a cheap deterministic hash. Scene detail placed through it is stable
// for a given seed, so the same repo always gets the same shoreline.
func Hash2(x, y int, seed int64) uint32 {
	h := uint32(int64(x)*374761393 + int64(y)*668265263 + seed*2654435761)
	h ^= h >> 13
	h *= 1274126177
	return h ^ (h >> 16)
}

func HashF(x, y int, seed int64) float64 {
	return float64(Hash2(x, y, seed)) / 4294967296.0
}
