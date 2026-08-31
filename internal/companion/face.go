package companion

import (
	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// Face is the feline detail drawn on top of the body bitmap.
//
// It has to be character overlays rather than bitmap detail, and that is the
// medium's doing rather than a preference. The whole head is four character
// rows -- ears, forehead, eyes, chin -- the eyes already own their row, and the
// bitmap ORs its rows in PAIRS before quadranting, so a one-row detail is
// erased by the row above it. That is not hypothetical: CatBody has carried a
// muzzle gap since it was drawn and it has never once reached the screen.
//
// Overlays are plotted after the body, at cell precision, in their own colour.
// WhiskerStyle is how the whiskers are drawn. They are different approaches,
// not different amounts: a dotted whisker and a stroked one are trying
// different things at this size.
type WhiskerStyle int

const (
	NoWhiskers WhiskerStyle = iota
	// WhiskerStrokes: two horizontal lines a side, outside the head.
	WhiskerStrokes
	// WhiskerDots: only the tips, so the eye infers the line between them.
	WhiskerDots
	// WhiskerFan: diagonals, splaying up and down from the muzzle.
	WhiskerFan
	// WhiskerTicks: short marks pressed against the cheek, not reaching out.
	WhiskerTicks
	// WhiskerCarved: drawn INTO the fur as darker cells rather than beside it,
	// so the silhouette stays clean.
	WhiskerCarved
	// WhiskerLow: strokes below the muzzle line, where a cat's actually sit.
	WhiskerLow
)

// EarStyle is how the inside of the ear is treated.
type EarStyle int

const (
	NoEars EarStyle = iota
	// EarRose: a tinted cell, the pink inside of an ear.
	EarRose
	// EarDark: the same shape in a darker coat tone -- shadow, not skin.
	EarDark
	// EarTuft: a wisp of fur inside the ear instead of a colour.
	EarTuft
	// EarDot: the smallest possible mark.
	EarDot
	// EarRim: the ear edge lightened rather than the inside filled.
	EarRim
)

type Face struct {
	Nose     bool
	Whiskers WhiskerStyle
	Ears     EarStyle
	// EarTufts are the lynx-like points a cat carries at the ear tip.
	EarTufts bool
	// Brow is the tabby M, the marking almost every tabby wears on its
	// forehead -- the most cat available in three cells.
	Brow bool
	// Muzzle lightens the snout around the nose, which is what gives a cat its
	// short-faced look from across a room.
	Muzzle bool
	// Bib is the chest patch. It reads best on a dark coat, where it is the
	// difference between a cat and a silhouette.
	Bib bool
	// Toes are the pale tips of the front paws.
	Toes bool
	// Chin is a single pale cell under the muzzle -- the smallest mark on the
	// sheet, and the one that costs the least.
	Chin bool
	// TailTip pales the last of the tail, which is a marking a lot of cats
	// carry and which nothing else on the body competes with.
	TailTip bool
	// Brows are two faint dots above the eyes, where a cat's whisker-brows sit.
	Brows bool
}

// Base is what Lucas has settled on: a nose and pale toe tips.
var Base = Face{Nose: true, Toes: true}

// WhiskerStyles and EarStyles are the alternatives to compare, in the order
// the sheet shows them.
var WhiskerStyles = []struct {
	Style WhiskerStyle
	Name  string
	Note  string
}{
	{NoWhiskers, "none", "the base cat"},
	{WhiskerStrokes, "strokes", "two lines a side, outside the head"},
	{WhiskerDots, "tips", "only the ends; the eye fills in the line"},
	{WhiskerFan, "fan", "diagonals splaying from the muzzle"},
	{WhiskerTicks, "ticks", "short marks against the cheek"},
	{WhiskerCarved, "carved", "darker cells INSIDE the fur"},
	{WhiskerLow, "low", "strokes below the muzzle line"},
}

var EarStyles = []struct {
	Style EarStyle
	Name  string
	Note  string
}{
	{NoEars, "none", "the base cat"},
	{EarRose, "rose", "a tinted cell -- skin"},
	{EarDark, "shadow", "the same shape in a darker coat tone"},
	{EarTuft, "tuft", "a wisp of fur instead of a colour"},
	{EarDot, "dot", "the smallest possible mark"},
	{EarRim, "rim", "the ear EDGE lightened, not the inside filled"},
}

// Coats are the body colours.
var Coats = map[string]term.RGB{
	"cream":    {R: 236, G: 228, B: 210}, // what ships today
	"snow":     {R: 248, G: 246, B: 242},
	"fog":      {R: 206, G: 202, B: 198},
	"slate":    {R: 142, G: 156, B: 176},
	"ink":      {R: 96, G: 110, B: 142},
	"taupe":    {R: 172, G: 154, B: 138},
	"sage":     {R: 150, G: 170, B: 150},
	"mauve":    {R: 172, G: 152, B: 178},
	"charcoal": {R: 92, G: 96, B: 108},
}

var CoatOrder = []string{"cream", "fog", "slate", "sage", "mauve", "charcoal"}

var (
	noseCol  = term.RGB{R: 226, G: 138, B: 148} // muted rose
	innerEar = term.RGB{R: 198, G: 130, B: 132}
)

// shade lightens or darkens the coat, so a marking belongs to the cat rather
// than being painted on it -- and so a new coat needs no new colours.
func shade(c term.RGB, f float64) term.RGB {
	if f >= 0 {
		return term.Lerp(c, term.RGB{R: 255, G: 250, B: 242}, f)
	}
	return term.Lerp(c, term.RGB{R: 26, G: 26, B: 32}, -f)
}

// drawFace plots the overlays. x,y is the sprite's top-left cell. The body is
// nine cells wide with the eyes at cells 2 and 6, so the muzzle centres on 4.
func (c *Cat) drawFace(l *canvas.Layer, x, y int, f Face, st State) {
	w, _ := c.Size()
	// Feature positions are given in the sprite's own frame, left to right, and
	// flipped here when the companion faces the other way -- so each feature is
	// authored once and stays on the correct side of the face.
	at := func(cell int) int {
		if c.mirror {
			return x + w - 1 - cell
		}
		return x + cell
	}

	light := shade(c.coat, 0.45)
	dark := shade(c.coat, -0.32)
	pale := shade(c.coat, 0.72)

	if f.EarTufts {
		l.Plot(at(1), y-1, '\'', light, 0.7)
		l.Plot(at(6), y-1, '\'', light, 0.7)
	}

	switch f.Ears {
	case EarRose:
		l.Plot(at(1), y, '▖', innerEar, 0.85)
		l.Plot(at(6), y, '▗', innerEar, 0.85)
	case EarDark:
		l.Plot(at(1), y, '▖', dark, 0.9)
		l.Plot(at(6), y, '▗', dark, 0.9)
	case EarTuft:
		l.Plot(at(1), y, '\'', pale, 0.8)
		l.Plot(at(6), y, '\'', pale, 0.8)
	case EarDot:
		l.Plot(at(1), y, '·', innerEar, 0.9)
		l.Plot(at(6), y, '·', innerEar, 0.9)
	case EarRim:
		l.Plot(at(1), y, '▘', light, 0.75)
		l.Plot(at(6), y, '▝', light, 0.75)
	}

	// The tabby M, between the ears on the forehead row.
	if f.Brow {
		l.Plot(at(3), y+1, '\\', dark, 0.8)
		l.Plot(at(4), y+1, 'ʌ', dark, 0.8)
		l.Plot(at(5), y+1, '/', dark, 0.8)
	}

	// Drawn before the nose, so the nose sits on top of it.
	if f.Muzzle {
		l.Plot(at(3), y+3, '▗', pale, 0.9)
		l.Plot(at(5), y+3, '▖', pale, 0.9)
	}

	// The nose does not change colour with state. The eyes carry state, and two
	// things carrying one signal is the collision the encoding rule exists to
	// prevent.
	if f.Nose {
		l.Plot(at(4), y+3, '▾', noseCol, 1)
	}

	// Whiskers. Every style below is a different idea rather than a different
	// amount, because at nine cells wide the choice is not how many strokes to
	// draw but whether a whisker is a line, a pair of tips, or a shadow in the
	// fur -- and those read completely differently.
	if f.Whiskers != NoWhiskers {
		top, bot := y+2, y+3
		if st == Worried {
			top, bot = y+3, y+4 // drooping
		}
		near, far := 1.0, 0.6
		if st == Resting {
			near, far = 0.75, 0.4
		}
		switch f.Whiskers {
		case WhiskerStrokes:
			for _, cell := range []int{-1, 9} {
				l.Plot(at(cell), bot, '─', light, near)
				l.Plot(at(cell), top, '─', light, far)
			}
		case WhiskerDots:
			for _, cell := range []int{-1, 9} {
				l.Plot(at(cell), bot, '·', light, near)
				l.Plot(at(cell), top, '·', light, far*0.9)
			}
		case WhiskerFan:
			l.Plot(at(-1), top, '╱', light, far)
			l.Plot(at(-1), bot, '╲', light, near)
			l.Plot(at(9), top, '╲', light, far)
			l.Plot(at(9), bot, '╱', light, near)
		case WhiskerTicks:
			l.Plot(at(-1), bot, '╴', light, near)
			l.Plot(at(9), bot, '╶', light, near)
		case WhiskerCarved:
			// Inside the silhouette, so the outline stays clean.
			l.Plot(at(1), bot, '╌', dark, 0.85)
			l.Plot(at(7), bot, '╌', dark, 0.85)
		case WhiskerLow:
			for _, cell := range []int{-1, 9} {
				l.Plot(at(cell), bot+1, '─', light, near)
				l.Plot(at(cell), bot, '─', light, far)
			}
		}
	}

	// The bib is deliberately smaller than it was. A four-cell block read as a
	// bib worn by the cat rather than as its own chest; two cells narrowing to
	// one reads as fur, and on a pale coat it should barely register at all.
	if f.Bib {
		l.Plot(at(4), y+4, '▄', pale, 0.6)
		l.Plot(at(4), y+5, '▐', pale, 0.7)
	}

	if f.Toes {
		l.Plot(at(2), y+6, '▖', pale, 0.55)
		l.Plot(at(6), y+6, '▗', pale, 0.55)
	}

	if f.Chin {
		l.Plot(at(4), y+4, '▘', pale, 0.55)
	}

	// The tail sweeps up on the far side of the body, two cells past the hip.
	if f.TailTip {
		l.Plot(at(10), y+3, '▘', pale, 0.7)
	}

	if f.Brows {
		l.Plot(at(2), y+1, '˙', light, 0.5)
		l.Plot(at(6), y+1, '˙', light, 0.5)
	}
}
