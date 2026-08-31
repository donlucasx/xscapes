// Package companion holds the resident creature: the thing that idles while the
// agent works and comes to tell you when it is done.
package companion

import (
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// Register is how a sprite is drawn. Kept explicit because it decides the glyph
// vocabulary and how the sprite composites, not just how it looks.
type Register string

const (
	Line    Register = "line-art"
	Block   Register = "silhouette"
	Braille Register = "braille"
)

// Sprite is a small rune grid. A space is transparent — it lets the scene show
// through rather than punching a hole in it.
type Sprite struct {
	Name     string
	Register Register
	Note     string // what the animation would hang off
	Rows     []string
	Body     term.RGB
	Accent   term.RGB // eyes and highlights
	AccentOf string   // runes drawn in Accent instead of Body
	Alpha    float64
	Say      string // if set, a speech bubble is drawn above
	Source   string // which reference this was translated from

	// Opaque makes spaces paint rather than pass through. Creature sprites want
	// transparent spaces so the scene shows between the ears; TEXT does not --
	// a transparent space lets scenery bleed into the middle of a word.
	Opaque bool
}

// Bubble builds a speech balloon. All three rows are the same width so the box
// cannot shear; the tail sits under the left shoulder, pointing at the creature.
func Bubble(text string) []string {
	// The box is sized in runes, so a wide rune would make the drawn box
	// narrower than the border it is measured against.
	inner := " " + NarrowOnly(text) + " "
	n := len([]rune(inner))
	if n < 4 {
		inner += strings.Repeat(" ", 4-n)
		n = 4
	}
	bar := strings.Repeat("-", n)
	return []string{
		"." + bar + ".",
		"|" + inner + "|",
		"'--v" + strings.Repeat("-", n-3) + "'",
	}
}

func (s *Sprite) Size() (w, h int) {
	for _, r := range s.Rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	return w, len(s.Rows)
}

// Draw plots the sprite with its top-left at (x, y).
func (s *Sprite) Draw(l *canvas.Layer, x, y int) {
	a := s.Alpha
	if a == 0 {
		a = 1
	}
	for dy, row := range s.Rows {
		for dx, r := range []rune(row) {
			if r == ' ' && !s.Opaque {
				continue
			}
			col := s.Body
			if s.AccentOf != "" && strings.ContainsRune(s.AccentOf, r) {
				col = s.Accent
			}
			l.Plot(x+dx, y+dy, r, col, a)
		}
	}
}
