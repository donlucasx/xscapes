package companion

import (
	"math"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// State is what the companion is reacting to. Deliberately not the agent's
// event vocabulary -- the companion knows three moods, nothing about tools.
type State int

const (
	Resting State = iota
	Working
	NeedsYou
)

func (s State) String() string {
	switch s {
	case Working:
		return "working"
	case NeedsYou:
		return "needs you"
	}
	return "resting"
}

var (
	furCol   = term.RGB{R: 236, G: 228, B: 210}
	eyeCol   = term.RGB{R: 168, G: 236, B: 176} // moonlit shine, not a highlight
	eyeAlert = term.RGB{R: 232, G: 252, B: 226}
)

// Cat is the companion. The body is a bitmap; the tail is a curve evaluated per
// frame, which is why wagging costs no extra authoring.
type Cat struct{ body *Bitmap }

func NewCat() *Cat { return &Cat{body: ParseBitmap(CatBody)} }

// Size is the character footprint, not the pixel size.
func (c *Cat) Size() (w, h int) { return c.body.W / 2, c.body.H / 4 }

// Draw composes this frame and plots it. Everything that moves is computed
// here; nothing is stored between frames.
func (c *Cat) Draw(l *canvas.Layer, x, y int, t float64, st State) {
	f := c.body.Blank()

	// Breathing. One quadrant subpixel is two source rows, so shifting by two
	// moves the whole body by exactly half a character cell -- the smallest
	// vertical step this medium has, and slow enough to read as breath.
	period, wag := 3.6, math.Sin(t*0.6)*0.3
	switch st {
	case Working:
		period, wag = 2.2, math.Sin(t*2.4)
	case NeedsYou:
		period, wag = 1.6, math.Sin(t*5.0)
	}
	lift := 0
	if math.Sin(2*math.Pi*t/period) > 0 {
		lift = 2
	}
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if c.body.at(sx, sy-lift) {
				f.Set(sx, sy)
			}
		}
	}
	c.tail(f, wag, lift)

	(&Sprite{Rows: f.ToQuadrant(), Body: furCol}).Draw(l, x, y)
	c.eyes(l, x, y, t, st)
}

// tail sweeps a curve up from the right hip. Two pixels thick so it survives
// the halving that ToQuadrant does.
func (c *Cat) tail(b *Bitmap, wag float64, lift int) {
	for i := 0; i <= 13; i++ {
		f := float64(i) / 13
		x := 17 + int(5*f+2.2*math.Sin(wag*1.6+f*2.6)+0.5)
		y := 24 - int(13*f+0.5) + lift
		b.Set(x, y)
		b.Set(x+1, y)
		b.Set(x, y+1)
	}
}

// eyes are plotted as characters ON TOP of the quadrant body, in the gaps the
// bitmap leaves for them. Two cells carry the whole expression, which is what
// makes the animation cheap: the other eighty never change.
func (c *Cat) eyes(l *canvas.Layer, x, y int, t float64, st State) {
	glyph, col := 'o', eyeCol
	switch st {
	case Resting:
		glyph = '-' // dozing between turns
	case NeedsYou:
		glyph, col = 'O', eyeAlert
	}
	if st != Resting && math.Mod(t, 5.3) < 0.16 {
		glyph = '-' // blink
	}
	l.Plot(x+2, y+2, glyph, col, 1)
	l.Plot(x+6, y+2, glyph, col, 1)
}
