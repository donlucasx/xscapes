package scape

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// StarInk is what the constellation's ink search actually did, rather than
// only what it returned.
//
// It exists for one reason. starInk stops at the FIRST rung that clears
// todoStarContrast, so a reported contrast of 56 is the search's exit
// condition and not a margin: it cannot tell a cell that cleared at rung 0
// with seven rungs of headroom from one that cleared at rung 8 and full alpha
// with none. That difference is the whole question the moment a mark is asked
// to sit somewhere OUTSIDE the band starBand certifies -- which is exactly
// what an arriving star's flight path does.
type StarInk struct {
	Ink   term.RGB
	Alpha float64
	// Rung is which step of the lift toward white won, 0..starInkSteps.
	Rung int
	// FullAlpha says the search had to give up the star's own magnitude and
	// go fully opaque to clear the bar.
	FullAlpha bool
	// Exhausted says nothing cleared the bar at all and this is best-effort
	// ink: the sky there is brighter than white can beat.
	Exhausted bool
	// Contrast is the rendered luma over the rendered ground, by the same
	// band-aware measure starInk judges with.
	Contrast float64
}

// StarInkAt is starInk with its working shown. The scene itself never needs
// this; a study that has to justify a glyph's legibility does.
func (s *Shore) StarInkAt(c *canvas.Canvas, x, y int, mag float64) StarInk {
	return s.starInkReport(c, x, y, mag)
}

// StarMagnitudeAt is the magnitude of the i-th constellation slot.
//
// Exported so a study can ask the product what a star's weight is instead of
// replicating the formula. notes/s28-starfun/sky.go replicated the star's
// alpha as a constant 0.85 -- a floor shore.go had already deleted -- and its
// render differs from the product's for 25 of 32 stars at midnight.
func StarMagnitudeAt(i int, seed int64) float64 { return starMagnitude(i, seed) }
