package companion

import (
	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// Face is the detail added on top of the body bitmap.
//
// It has to be character overlays rather than bitmap detail, and that is a
// consequence of the medium rather than a preference. The whole head is four
// character rows: ears, forehead, eyes, chin. The eye row already carries the
// eyes, and the bitmap halves its rows in pairs before quadranting, so a
// one-row nose drawn into the bitmap is OR'd away by the row above it -- which
// is exactly what happened to the muzzle gap that has been in CatBody all
// along and has never once reached the screen.
//
// Overlays do not have that problem: they are plotted after the body, at cell
// precision, in their own colour.
type Face struct {
	Nose      bool
	Whiskers  bool
	InnerEars bool
	// Cheeks widens the muzzle with a tuft either side of the nose.
	Cheeks bool
}

// Faces are the variants worth looking at, cheapest detail first.
var Faces = map[string]Face{
	"plain":    {},
	"nose":     {Nose: true},
	"whiskers": {Nose: true, Whiskers: true},
	"full":     {Nose: true, Whiskers: true, InnerEars: true, Cheeks: true},
}

// Coats are the body colours. The terracotta is Claude's own, which is the
// reference Lucas pointed at.
var Coats = map[string]term.RGB{
	"cream":      {R: 236, G: 228, B: 210},
	"terracotta": {R: 217, G: 119, B: 87},
	"ginger":     {R: 224, G: 146, B: 84},
	"slate":      {R: 142, G: 156, B: 176},
	"charcoal":   {R: 92, G: 96, B: 108},
}

var (
	noseCol    = term.RGB{R: 226, G: 138, B: 148} // muted rose, reads as a nose at one cell
	whiskerCol = term.RGB{R: 214, G: 206, B: 190}
	innerEar   = term.RGB{R: 198, G: 130, B: 132}
)

// drawFace plots the overlays. x,y is the sprite's top-left cell; the body is
// nine cells wide with the eyes at cells 2 and 6, so the muzzle centres on 4.
func (c *Cat) drawFace(l *canvas.Layer, x, y int, f Face, st State) {
	w, _ := c.Size()
	at := func(cell int) int {
		if c.mirror {
			return x + w - 1 - cell
		}
		return x + cell
	}

	if f.InnerEars {
		// The ears sit at cells 1 and 6 of the top row. A single tinted cell
		// inside each one reads as the pink inside of an ear.
		l.Plot(at(1), y, '▖', innerEar, 0.85)
		l.Plot(at(6), y, '▗', innerEar, 0.85)
	}

	// The nose sits one row below the eyes, centred between them. A worried
	// cat's nose does not change colour -- the eyes carry state, and two things
	// carrying one signal is how the weather ended up carrying busy AND broken.
	if f.Nose {
		l.Plot(at(4), y+3, '▾', noseCol, 1)
	}

	if f.Cheeks {
		l.Plot(at(3), y+3, '▁', whiskerCol, 0.5)
		l.Plot(at(5), y+3, '▁', whiskerCol, 0.5)
	}

	// Whiskers use the empty cells beside the head, so they cost no face area.
	// They droop a little when the companion is worried, which is free
	// expression from geometry that is already there.
	if f.Whiskers {
		row := y + 3
		if st == Worried {
			row = y + 4
		}
		a := 1.0
		if st == Resting {
			a = 0.7 // relaxed, less splayed
		}
		l.Plot(at(-1), row, '~', whiskerCol, a)
		l.Plot(at(9), row, '~', whiskerCol, a)
		if st != Resting {
			l.Plot(at(-1), row-1, '~', whiskerCol, a*0.55)
			l.Plot(at(9), row-1, '~', whiskerCol, a*0.55)
		}
	}
}
