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
type Face struct {
	Nose      bool
	Whiskers  bool
	InnerEars bool
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
}

// Faces are the presets worth comparing. Each adds to the one before it.
var Faces = map[string]Face{
	"plain":   {},
	"nose":    {Nose: true},
	"classic": {Nose: true, Whiskers: true, InnerEars: true},
	"tabby":   {Nose: true, Whiskers: true, InnerEars: true, Brow: true, EarTufts: true},
	"tuxedo":  {Nose: true, Whiskers: true, InnerEars: true, Muzzle: true, Bib: true, Toes: true},
	"full": {Nose: true, Whiskers: true, InnerEars: true, EarTufts: true,
		Brow: true, Muzzle: true, Bib: true, Toes: true},
}

// FaceOrder is the reading order for the study: a build-up, not a grid.
var FaceOrder = []string{"plain", "nose", "classic", "tabby", "tuxedo", "full"}

// Coats are the body colours.
var Coats = map[string]term.RGB{
	"cream":    {R: 236, G: 228, B: 210},
	"slate":    {R: 142, G: 156, B: 176},
	"charcoal": {R: 92, G: 96, B: 108},
	"ginger":   {R: 224, G: 146, B: 84},
}

var CoatOrder = []string{"cream", "slate", "charcoal", "ginger"}

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
	if f.InnerEars {
		l.Plot(at(1), y, '▖', innerEar, 0.85)
		l.Plot(at(6), y, '▗', innerEar, 0.85)
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

	// Whiskers use the empty cells beside the head, so they cost no face area.
	//
	// Two horizontal strokes a side, not one. A single mark reads as
	// punctuation -- a bracket sitting next to the cat -- while a pair reads as
	// a fan coming off the muzzle. That was judged from rendered text rather
	// than reasoned about; the bracket version looked fine in the source.
	if f.Whiskers {
		top, bot := y+2, y+3
		if st == Worried {
			top, bot = y+3, y+4 // drooping
		}
		near, far := 1.0, 0.6
		if st == Resting {
			near, far = 0.75, 0.4
		}
		for _, cell := range []int{-1, 9} {
			l.Plot(at(cell), bot, '─', light, near)
			l.Plot(at(cell), top, '─', light, far)
		}
		// Splayed when the agent wants something: one more stroke, further out.
		if st == NeedsYou {
			l.Plot(at(-2), bot, '─', light, 0.55)
			l.Plot(at(10), bot, '─', light, 0.55)
		}
	}

	// Two rows, narrowing, so the bib reads as fur rather than a rectangle.
	if f.Bib {
		l.Plot(at(4), y+4, '▄', pale, 0.9)
		l.Plot(at(3), y+5, '▖', pale, 0.75)
		l.Plot(at(4), y+5, '█', pale, 0.9)
		l.Plot(at(5), y+5, '▗', pale, 0.75)
	}

	if f.Toes {
		l.Plot(at(2), y+6, '▖', pale, 0.8)
		l.Plot(at(6), y+6, '▗', pale, 0.8)
	}
}
