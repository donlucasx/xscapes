// Package scape holds the scenes. A scape reads only an Activity — it never
// knows which agent is running or what a "tool call" is. That separation is
// what lets one renderer serve every adapter.
package scape

import "github.com/donlucasx/asciiscapes/internal/canvas"

// Activity is the agent's state, normalised. Level is 0 (idle) to 1 (flat out).
type Activity struct {
	Working bool
	Level   float64
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
