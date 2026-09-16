package scenes

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE VISTA AS A SCAPE. His pick of 2026-09-15 ("A: the wind is the work,
// the fire the night light"), on the page as a clip since then, a scape a
// user can run since 2026-09-16 (s35).
//
// The six slots, and what carries each:
//
//	light        the sky's own palette by the wall clock; the fire lights the
//	             meadow at night, fixed, so it says nothing about the agent
//	sky          the real sky over three ranges; the moon sinks with the
//	             context, as on the shore
//	motion       THE WIND: how far the smoke and the flames lean, what
//	             fraction of the scrub lies flat, what is in the air --
//	             coverage and position, never rate
//	surface      rock, snow, the lake, the meadow
//	accumulator  finished todos as stars in the sky; the activity tail
//	             written on the band below the meadow
//	companion    the barn owl on its mound, five states; owlets for subagents
//
// The painter is the study's (forest.go), parametrised by a layout so the
// same drawing sits in any window; at 80x24 it is byte-identical to the
// clip on the page (TestTheStudyVistaIsUnchanged).
type Vista struct {
	Seed  int64
	Ascii bool
	// OwlX is the column the composition gives the companion, so the owl
	// stands where the crab stands and the litter and the balloon use the
	// shore's own arithmetic. Zero takes the study's place.
	OwlX int
	lay  vistaLayout
	lit  float64
}

func NewVista(seed int64, ascii bool) *Vista { return &Vista{Seed: seed, Ascii: ascii} }

// VistaMinRows is the height under which the owl no longer fits between the
// sky and the writing and overlaps the sky. The scene still draws; it is
// the vista's floor the way 40x12 is the shore's.
const VistaMinRows = 16

func (v *Vista) Name() string { return "vista" }

// Update paints the whole frame but the owl and the writing, which the live
// composer draws over it with the reducer's pose and tail.
func (v *Vista) Update(c *canvas.Canvas, t float64, act scape.Activity) {
	v.lay = vistaLayoutFor(c.W, c.H, v.OwlX)
	v.lit = lit(scape.PaletteAt(act.TimeOfDay))
	paintVistaL(c, v.lay, act.TimeOfDay, t, act.Level, v.Seed, false, &vistaLive{
		ContextUsed: act.ContextUsed, TodoDone: act.TodoDone,
		SkipOwl: true, SkipBand: true,
	})
}

// Layout is the frame's geometry after the last Update. The owlets stand
// on the grass with their feet on the band's edge, four rows above it:
// anchored to the band and not to the meadow's top, because a short window
// gives the meadow four rows and a meadow-anchored owlet put its feet on the
// writing (seen at 60x20).
func (v *Vista) Layout() (owlX, owlY, grassY, bandTop int) {
	return v.lay.owlX, v.lay.owlY, v.lay.bandTop - 4, v.lay.bandTop
}

// BandTop is the first row of the writing.
func (v *Vista) BandTop() int { return v.lay.bandTop }

// BandColor is the writing's ground, for the sand's ink to be sampled from.
func (v *Vista) BandColor() term.RGB { return v.lay.bandColor(v.lit) }

// MoonAt is the moon's cell, for the context readout to sit under.
func (v *Vista) MoonAt(ctxUsed float64) (x, y int) {
	return v.lay.moonX, v.lay.moonRow(ctxUsed)
}

// vistaLayoutFor scales the 80x24 layout to a window. The band keeps the
// tail's four lines once there is room; the meadow and the lake take their
// 80x24 share of the rest; the ranges stand on the lake's edge as they did;
// the owl sits on the mound above the band, and the fire keeps its place at
// three eighths of the width.
func vistaLayoutFor(w, h int, owlX int) vistaLayout {
	if w == 80 && h == 24 && (owlX == 0 || owlX == 64) {
		return vista80()
	}
	lay := vistaLayout{W: w, H: h}
	band := 3
	if h >= 28 {
		band = 4
	}
	if h >= 40 {
		band = 5
	}
	meadow := max(2, (5*h+12)/24)
	lake := max(1, (2*h+12)/24)
	lay.bandTop = h - band
	lay.meadowTop = lay.bandTop - meadow
	lay.lakeTop = lay.meadowTop - lake
	if lay.lakeTop < 6 {
		// Too short for a sky worth the name: give the sky what is left.
		lay.lakeTop = max(2, lay.meadowTop-1)
	}
	lay.nearBase = 2 * lay.lakeTop
	lay.midBase = 2 * (lay.lakeTop - 1)
	lay.farBase = 2 * (lay.lakeTop - 3)
	lay.owlY = lay.bandTop - 9
	if lay.owlY < 0 {
		lay.owlY = 0
	}
	if owlX == 0 {
		owlX = w - 16
	}
	lay.owlX = owlX
	lay.fireX = 3 * w / 8
	lay.moonX = max(1, 12*w/80)
	return lay
}
