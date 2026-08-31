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
// WhiskerStyle is how the whiskers are drawn.
//
// Cell geometry, measured off a rendered frame rather than off the bitmap
// source, because the source is two transforms away from what lands:
//
//	cell: 0123456789
//	ears  .++...++.     ears occupy cells 1-2 and 6-7
//	chin  +########.#   fur runs 0-8; 0 and 8 are HALF-filled edges, 10 is tail
//
// That half-filled edge is why whiskers drawn at cell -1 and 9 looked detached:
// they were adjacent to the head's bounding box but two cells from solid fur.
type WhiskerStyle int

const (
	NoWhiskers WhiskerStyle = iota
	// WhiskerFlush sits ON the fur edge, cells 0 and 8, so it grows out of the
	// face instead of hovering beside it.
	WhiskerFlush
	// WhiskerCheek is further in still, on solid fur at cells 1 and 7.
	WhiskerCheek
	// WhiskerLong runs from the fur edge out past it -- flush plus a second
	// cell beyond.
	WhiskerLong
	// WhiskerFlushOne is a single stroke a side, on the fur edge.
	WhiskerFlushOne
	// WhiskerGap is the old placement, kept so the difference is visible.
	WhiskerGap
)

// EarStyle is how the inside of the ear is treated. Every style below sits
// WITHIN cells 1-2 and 6-7, which is where the ears actually are.
type EarStyle int

const (
	NoEars EarStyle = iota
	// EarInner marks the inward half of each ear: cells 2 and 6.
	EarInner
	// EarOuter marks the outward half: cells 1 and 7.
	EarOuter
	// EarFill tints both cells of each ear.
	EarFill
	// EarInnerDark is EarInner in the coat's own shadow rather than rose, so
	// the cat stays monochrome.
	EarInnerDark
	// EarInnerTuft puts a wisp of fur in the inward half instead of a colour.
	EarInnerTuft
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
	{WhiskerFlush, "flush", "on the fur edge, cells 0 and 8"},
	{WhiskerCheek, "cheek", "further in, on solid fur"},
	{WhiskerLong, "long", "from the edge, reaching out one more"},
	{WhiskerFlushOne, "flush single", "one stroke a side, on the edge"},
	{WhiskerGap, "gap (before)", "the old placement, for comparison"},
}

var EarStyles = []struct {
	Style EarStyle
	Name  string
	Note  string
}{
	{NoEars, "none", "the base cat"},
	{EarInner, "inner", "inward half of each ear"},
	{EarOuter, "outer", "outward half"},
	{EarFill, "fill", "both cells, fading outward"},
	{EarInnerDark, "inner shadow", "coat's own shadow, not rose"},
	{EarInnerTuft, "inner tuft", "a wisp of fur instead of a colour"},
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

var CoatOrder = []string{"cream", "slate", "sage", "mauve", "charcoal"}

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
	case EarInner:
		l.Plot(at(2), y, '▖', innerEar, 0.9)
		l.Plot(at(6), y, '▗', innerEar, 0.9)
	case EarOuter:
		l.Plot(at(1), y, '▖', innerEar, 0.9)
		l.Plot(at(7), y, '▗', innerEar, 0.9)
	case EarFill:
		l.Plot(at(1), y, '▖', innerEar, 0.7)
		l.Plot(at(2), y, '▖', innerEar, 0.9)
		l.Plot(at(6), y, '▗', innerEar, 0.9)
		l.Plot(at(7), y, '▗', innerEar, 0.7)
	case EarInnerDark:
		l.Plot(at(2), y, '▖', dark, 0.9)
		l.Plot(at(6), y, '▗', dark, 0.9)
	case EarInnerTuft:
		l.Plot(at(2), y, '\'', pale, 0.85)
		l.Plot(at(6), y, '\'', pale, 0.85)
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

	// Whiskers. The question these variants answer is where a whisker STARTS:
	// the earlier version began a cell past the head's bounding box, which put
	// it two cells from solid fur because the edge cell is only half filled --
	// so it read as a mark near the cat rather than a mark on it.
	if f.Whiskers != NoWhiskers {
		top, bot := y+2, y+3
		if st == Worried {
			top, bot = y+3, y+4 // drooping
		}
		near, far := 1.0, 0.6
		if st == Resting {
			near, far = 0.75, 0.4
		}
		pair := func(a, b int, rows ...int) {
			for i, r := range rows {
				alpha := near
				if i > 0 {
					alpha = far
				}
				l.Plot(at(a), r, '─', light, alpha)
				l.Plot(at(b), r, '─', light, alpha)
			}
		}
		switch f.Whiskers {
		case WhiskerFlush:
			pair(0, 8, bot, top)
		case WhiskerCheek:
			pair(1, 7, bot, top)
		case WhiskerLong:
			pair(0, 8, bot, top)
			l.Plot(at(-1), bot, '─', light, far)
			l.Plot(at(9), bot, '─', light, far)
		case WhiskerFlushOne:
			pair(0, 8, bot)
		case WhiskerGap:
			pair(-1, 9, bot, top)
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
