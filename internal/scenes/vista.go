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
//	motion       THE WIND AND THE FIRE (both since 2026-09-16, his word;
//	             the wind alone was the 09-14 pick): how far the smoke and
//	             the flames lean, how tall the fire, its sparks, what
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
	// MoonStyle is how the context body is drawn, an index into MoonStyles;
	// 0 is the study's block. Set from the pick when he rules.
	MoonStyle int
	// Flock is the litter's memory of arrivals, for the flight in.
	Flock OwletFlock
	lay   vistaLayout
	lit   float64
}

// FireX is the fire's column after the last Update.
func (v *Vista) FireX() int { return v.lay.fireX }

// NewVista is the live vista: the context body on the arc (MoonStyle 5,
// his pick of 2026-09-16). The study's clip on the page keeps the block.
func NewVista(seed int64, ascii bool) *Vista { return &Vista{Seed: seed, Ascii: ascii, MoonStyle: 5} }

// VistaMinRows is the height under which the owl no longer fits between the
// sky and the writing and overlaps the sky. The scene still draws; it is
// the vista's floor the way 40x12 is the shore's.
const VistaMinRows = 16

func (v *Vista) Name() string { return "vista" }

// Update paints the whole frame but the owl and the writing, which the live
// composer draws over it with the reducer's pose and tail.
func (v *Vista) Update(c *canvas.Canvas, t float64, act scape.Activity) {
	// Every glyph layer is cleared first, as the shore's Update does. The
	// vista's did not (s35), and nothing showed it until an owlet FLEW:
	// the flight left a trail of owlets along its arc, one per frame, on
	// the placement page (his screenshot, 2026-09-16 11:46). Live it means
	// that whatever the wind carried through the air stayed where each
	// frame put it: TestTheVistaClearsBetweenFrames measures the sky's
	// glyph count over sixty frames and holds it to one frame's worth.
	c.Clear()
	v.lay = vistaLayoutFor(c.W, c.H, v.OwlX)
	v.lit = lit(scape.PaletteAt(act.TimeOfDay))
	paintVistaL(c, v.lay, act.TimeOfDay, t, act.Level, v.Seed, false, &vistaLive{
		ContextUsed: act.ContextUsed, TodoDone: act.TodoDone,
		SkipOwl: true, SkipBand: true, MoonStyle: v.MoonStyle, FireWork: true,
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

// MoonAt is the body's cell, for the context readout to sit under: the
// block's top-left, or for a disc the cell whose readout row is the row
// under the disc's bottom, centred on the disc.
func (v *Vista) MoonAt(ctxUsed float64) (x, y int) {
	if v.MoonStyle == 0 {
		return v.lay.moonX, v.lay.moonRow(ctxUsed)
	}
	cx, cy, r := v.lay.moonCenter(v.MoonStyle, ctxUsed)
	return int(cx) - 1, int(cy+r-0.5) - 1
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
