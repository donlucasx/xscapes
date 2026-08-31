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
	// Worried persists while something is broken, unlike the other states,
	// which track what the agent is doing right now. It is the companion's
	// job rather than the weather's: a storm cannot say whether the sea is
	// rough because the agent is busy or because the tests are red.
	Worried
)

func (s State) String() string {
	switch s {
	case Working:
		return "working"
	case NeedsYou:
		return "needs you"
	case Worried:
		return "something is broken"
	}
	return "resting"
}

var (
	furCol   = term.RGB{R: 236, G: 228, B: 210}
	eyeCol   = term.RGB{R: 168, G: 236, B: 176} // moonlit shine, not a highlight
	eyeAlert = term.RGB{R: 232, G: 252, B: 226}
	// Amber, against the moonlit green of every other state. Colour does the
	// work that a five-cell face cannot.
	eyeWorried = term.RGB{R: 244, G: 176, B: 96}
)

// Cat is the companion. The body is a bitmap; the tail is a curve evaluated per
// frame, which is why wagging costs no extra authoring.
type Cat struct {
	body     *Bitmap
	walk     *Bitmap
	worried  *Bitmap
	kitCache map[int]*Bitmap
	kitTier  int // last chosen litter size, for hysteresis
	swim     *Bitmap
}

func NewCat() *Cat {
	return &Cat{body: ParseBitmap(CatBody), worried: ParseBitmap(CatWorried)}
}

// Size is the character footprint, not the pixel size.
func (c *Cat) Size() (w, h int) { return c.body.W / 2, c.body.H / 4 }

// Draw composes this frame and plots it. Everything that moves is computed
// here; nothing is stored between frames.
func (c *Cat) Draw(l *canvas.Layer, x, y int, t float64, st State) {
	src := c.body
	if st == Worried {
		src = c.worried
	}
	f := src.Blank()

	// Breathing. One quadrant subpixel is two source rows, so shifting by two
	// moves the whole body by exactly half a character cell -- the smallest
	// vertical step this medium has, and slow enough to read as breath.
	period, wag, tailLen := 3.6, math.Sin(t*0.6)*0.3, 1.0
	switch st {
	case Working:
		period, wag = 2.2, math.Sin(t*2.4)
	case NeedsYou:
		period, wag = 1.6, math.Sin(t*5.0)
	case Worried:
		// Shallow, quick breathing and a tail tucked flat: distress reads as
		// stillness where the other states read as motion.
		period, wag, tailLen = 1.9, 0, 0.35
	}
	lift := 0
	if math.Sin(2*math.Pi*t/period) > 0 {
		lift = 2
	}
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if src.at(sx, sy-lift) {
				f.Set(sx, sy)
			}
		}
	}
	c.tail(f, wag, lift, tailLen)

	(&Sprite{Rows: f.ToQuadrant(), Body: furCol}).Draw(l, x, y)
	c.eyes(l, x, y, t, st)
}

// tail sweeps a curve up from the right hip. Two pixels thick so it survives
// the halving that ToQuadrant does.
func (c *Cat) tail(b *Bitmap, wag float64, lift int, length float64) {
	n := int(13 * length)
	for i := 0; i <= n; i++ {
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
	case Worried:
		glyph, col = 'o', eyeWorried
	}
	if st != Resting && st != Worried && math.Mod(t, 5.3) < 0.16 {
		glyph = '-' // blink
	}
	l.Plot(x+2, y+2, glyph, col, 1)
	l.Plot(x+6, y+2, glyph, col, 1)
}

// walkBody is loaded lazily so NewCat stays cheap for scapes that never walk.
func (c *Cat) walkBitmap() *Bitmap {
	if c.walk == nil {
		c.walk = ParseBitmap(CatWalk)
	}
	return c.walk
}

// WalkSize is the side view's character footprint.
func (c *Cat) WalkSize() (w, h int) {
	b := c.walkBitmap()
	return b.W / 2, b.H / 4
}

// DrawWalk plots the side view mid-stride. dir is +1 walking right, -1 left.
//
// phase advances with DISTANCE rather than with time, so the legs stay locked
// to the ground however fast the cat crosses. Driving a gait off a clock is
// what makes animated characters look like they are skating.
func (c *Cat) DrawWalk(l *canvas.Layer, x, y int, phase float64, dir int) {
	body := c.walkBitmap()
	f := body.Blank()

	bob := 0
	if math.Sin(phase*2) > 0.5 {
		bob = 1 // the whole body rises slightly at mid-stride
	}
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if body.at(sx, sy-bob) {
				f.Set(sx, sy)
			}
		}
	}
	c.legs(f, phase, bob)
	c.walkTail(f, math.Sin(phase*0.7), bob)

	if dir < 0 {
		f = f.Mirrored()
	}
	(&Sprite{Rows: f.ToQuadrant(), Body: furCol}).Draw(l, x, y)

	// The eye gap sits at source cols 25-26, cell column 12; mirrored it lands
	// at the far side of the sprite instead.
	ex := x + 12
	if dir < 0 {
		w, _ := c.WalkSize()
		ex = x + w - 1 - 12
	}
	l.Plot(ex, y+1, 'o', eyeCol, 1)
}

// legs: four bars swinging fore and aft, lifting off at the top of the swing.
// Diagonal pairs share a phase, which is what a real quadruped walk does.
func (c *Cat) legs(b *Bitmap, phase float64, bob int) {
	bases := [4]int{9, 13, 21, 25}
	offs := [4]float64{0, math.Pi, math.Pi, 0}
	for i, bx := range bases {
		ph := phase + offs[i]
		swing := int(2.5*math.Sin(ph) + 0.5)
		lift := 0
		if math.Sin(ph) > 0.3 {
			lift = 2
		}
		for yy := 19 + bob; yy <= 27-lift; yy++ {
			b.Set(bx+swing, yy)
			b.Set(bx+swing+1, yy)
		}
	}
}

func (c *Cat) walkTail(b *Bitmap, wag float64, bob int) {
	for i := 0; i <= 12; i++ {
		f := float64(i) / 12
		x := 6 - int(4*f+0.5)
		y := 14 - int(9*f+2.0*math.Sin(wag+f*2.0)+0.5) + bob
		b.Set(x, y)
		b.Set(x, y+1)
	}
}
