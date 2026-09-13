package companion

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The cat's come-closer ladder. His ruling of 2026-09-12, confirming what the
// whole round was for: "the goal is for the cat to come closer when it needs
// human input, right? like the crab".
//
// The machinery is the crab's -- Approach, nearRung, nearFits, nearDX and
// DrawnBox are shared and were already animal-agnostic once two kind gates came
// out. Three things are the cat's own:
//
//  1. THE TAIL. It is a procedural curve, not part of any bitmap, so doubling
//     the body arrives with no tail at all. tailAt takes a fractional scale for
//     exactly this: the middle rung is 32/24 source columns, which is 4/3, not
//     2. Getting that wrong draws the curve through the middle of the animal.
//  2. THE EYE IS A GLYPH IN A HOLE, not a bitmap with a painted ground. His
//     ruling of 2026-09-05: "regarding eyes, keep holes (as is)." The crab's
//     near eye paints its own coat behind itself because at 2x2 an unpainted
//     hole showed the SEA through the eye; the cat's eye is one cell, so the
//     scene behind it is a sliver and reads as the hole it is meant to be.
//  3. THE EYE CELLS ARE DERIVED PER RUNG rather than shared. Rung 1's art puts
//     them at cells 2 and 8; rung 2's sockets are three cells wide and their
//     centres are 5 and 13. ⚠ The mockup for this round borrowed the CRAB's
//     {8, 14} and painted both eyes on the animal's forehead -- mirrored they
//     land on ink -- which is most of what he was looking at when he said the
//     eyes needed fixing.

// catNearRung2 is the top rung's body for a state.
//
// ⚠ HAND-DRAWN, not doubled, since 2026-09-12. "Ruffed sit" is authored at
// 48x56 for this rung; the poses that have no drawing of their own at this size
// fall back to the shipped bitmap doubled with the muzzle filled, which is what
// every pose used to do. Worried and Done are drawn at 12x7 only, and they
// reach rung 2 solely in transit -- the ladder climbs on NeedsYou and retreats
// through whatever pose follows the answer.
func catNearRung2(st State) ([]string, []string) {
	switch st {
	case NeedsYou:
		return CatNear2Ask, CatNear2Lower
	case Working:
		return CatNear2Upper, CatNear2Lower
	case Worried:
		return nearDouble(nearFillHidden(CatWorried)), nil
	case Done:
		return nearDouble(nearFillHidden(CatDone)), nil
	}
	return CatNear2Upper, CatNear2Lower
}

func (c *Cat) catPoseFor(rung int, st State, t float64) nearPose {
	if rung >= 2 {
		up, lo := catNearRung2(st)
		eyes, row := catNear2EyeCells, catNear2EyeRow
		if lo == nil {
			// A doubled fallback pose: its sockets are the doubled body's,
			// cells 4..6 and 12..14 on row 4, and the glyph sits at each
			// centre. His ruling of 2026-09-12, option "a" of three.
			eyes, row = [2]int{5, 13}, 4
		}
		return nearPose{
			upper:    up,
			lower:    lo,
			eyeCells: eyes,
			eyeRow:   row,
			eyeWide:  1,
			glyphEye: true,
			// One original source row is half a character cell at 1x, and the
			// shipped breath is exactly that. At 2x the same half cell is two
			// source rows doubled, so four.
			lift:      4,
			tailScale: 2,
		}
	}
	upper := CatNear1Upper
	if st == NeedsYou {
		upper = CatNear1Ask
	}
	return nearPose{
		upper:    upper,
		lower:    CatNear1Lower,
		eyeCells: [2]int{2, 8},
		eyeRow:   3,
		eyeWide:  1,
		glyphEye: true,
		lift:     2,
		// 32 source columns against the shipped 24.
		tailScale: 32.0 / 24.0,
	}
}

// drawCatNear is the cat at a near rung: the same compose-breathe-mirror-plot
// the shipped path does, plus the tail, then a one-cell eye glyph per socket.
func (c *Cat) drawCatNear(l *canvas.Layer, x, y int, t float64, st State, rung int) {
	p := c.nearPoseFor(rung, st, t)
	src := ParseBitmap(append(append([]string{}, p.upper...), p.lower...))

	// The breath and the wag come off the shipped cat's own clocks, not copies:
	// a near pose that breathes on its own numbers drifts out of step with the
	// animal the viewer just watched walk up.
	period, wag, tailLen := 3.6, math.Sin(t*0.6)*0.3, 1.0
	switch st {
	case Working:
		period, wag = 2.2, math.Sin(t*2.4)
	case NeedsYou:
		period, wag = 1.6, math.Sin(t*5.0)
	case Worried:
		period, wag, tailLen = 1.9, 0, 0.35
	case Done:
		period, wag = 3.0, 0
	}
	lift := 0
	if math.Sin(2*math.Pi*t/period) > 0 {
		lift = p.lift
	}
	f := src.Blank()
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if src.at(sx, sy-lift) {
				f.Set(sx, sy)
			}
		}
	}
	c.tailAt(f, wag, lift, tailLen, p.tailScale)
	if c.mirror {
		f = f.Mirrored()
	}
	q := f.ToQuadrant()

	w := len([]rune(q[0]))
	bx := x + c.nearDX(w)
	plotRim(l, q, bx, y)
	(&Sprite{Rows: q, Body: c.coat}).Draw(l, bx, y)

	glyph, col := catEyeGlyph(st, t)
	for _, e := range p.eyeCells {
		ex := bx + e
		if c.mirror {
			ex = bx + w - p.eyeWide - e
		}
		// Plot, not PlotOn: no ground, so the sky shows around the glyph and
		// the eye stays the hole he ruled it should be.
		l.Plot(ex, y+p.eyeRow, glyph, col, 1)
	}
}

// catEyeGlyph is the shipped eye vocabulary, shared by every rung so the face
// does not change character as the animal approaches.
func catEyeGlyph(st State, t float64) (rune, term.RGB) {
	glyph, col := 'o', eyeCol
	switch st {
	case Resting:
		glyph = '-'
	case NeedsYou:
		glyph, col = 'O', EyeAlert
	case Worried:
		col = eyeWorried
	case Done:
		glyph = '^'
	}
	if st != Resting && st != Worried && st != Done && math.Mod(t, 5.3) < 0.16 {
		glyph = '-' // blink
	}
	return glyph, col
}
