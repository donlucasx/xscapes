// Package companion holds the resident creature: the thing that idles while the
// agent works and comes to tell you when it is done.
package companion

import (
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
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

// Bubble builds the ask balloon: the agent is blocked on the user. All three
// rows are the same width so the box cannot shear; the tail sits under the
// left shoulder, pointing at the creature. Solid strokes on purpose --
// DoneBubble is the same footprint with every stroke softened, and the pair
// has to stay distinguishable in a monochrome capture.
func Bubble(text string) []string {
	inner, n := bubbleInner(text)
	bar := strings.Repeat("-", n)
	return []string{
		"." + bar + ".",
		"|" + inner + "|",
		"'--v" + strings.Repeat("-", n-3) + "'",
	}
}

// DoneBubble is the finish knock. Same footprint as Bubble, nothing solid
// about it: dotted bars, colon walls. Shape carries the done/needs_input
// distinction wherever colour cannot -- a screenshot, or -plain.
func DoneBubble(text string) []string {
	inner, n := bubbleInner(text)
	dots := strings.Repeat(".", n)
	return []string{
		"." + dots + ".",
		":" + inner + ":",
		"'..v" + strings.Repeat(".", n-3) + "'",
	}
}

// MirrorTail moves a balloon's pointer from the left shoulder to the right,
// for the mirrored composition where the creature sits to the RIGHT of its
// balloon -- left where it was, the pointer aims at whatever kitten happens
// to be underneath instead.
func MirrorTail(rows []string) []string {
	out := append([]string(nil), rows...)
	last := []rune(out[len(out)-1])
	if i, j := 3, len(last)-4; j > 0 && j < len(last) && i < len(last) {
		last[i], last[j] = last[j], last[i]
	}
	out[len(out)-1] = string(last)
	return out
}

// bubbleInner pads the text into the balloon's middle row. Sized in runes,
// because a wide rune would make the drawn box narrower than the border it is
// measured against.
// MaxBubbleText is the most of a message a balloon will carry. The done knock
// carries the agent's last message, which can be a paragraph; drawn whole it
// ran across the entire sea in a dotted box (photographed 2026-09-03). A
// balloon is a glance, so it gets the first few words and an ellipsis.
const MaxBubbleText = 44

func bubbleInner(text string) (string, int) {
	text = NarrowOnly(text)
	if r := []rune(text); len(r) > MaxBubbleText {
		text = strings.TrimRight(string(r[:MaxBubbleText-1]), " ") + "…"
	}
	inner := " " + text + " "
	n := len([]rune(inner))
	if n < 4 {
		inner += strings.Repeat(" ", 4-n)
		n = 4
	}
	return inner, n
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
