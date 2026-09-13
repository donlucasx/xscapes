package companion

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
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
	// Done is the finish knock: the turn closed and the companion holds a
	// content pose for a bounded while. It is a separate state because the
	// brief locks done and needs_input as DISTINCT cues -- both used to
	// raise NeedsYou, so "come back when you like" and "you are blocking
	// the agent" looked identical.
	Done
)

func (s State) String() string {
	switch s {
	case Working:
		return "working"
	case NeedsYou:
		return "needs you"
	case Worried:
		return "something is broken"
	case Done:
		return "all done"
	}
	return "resting"
}

var (
	furCol = term.RGB{R: 236, G: 228, B: 210}
	eyeCol = term.RGB{R: 168, G: 236, B: 176} // moonlit shine, not a highlight
	// EyeAlert is the open eye of the NeedsYou pose, and it is EXPORTED so that
	// a test outside this package can find the companion's FACE on a rendered
	// frame. That guarantee -- the ask balloon's pointer comes out of the face,
	// which s27 shipped after the `v` sat nine cells into bare sand -- was
	// checked by reading back the literal rune 'O', and the near pose's eye is
	// a bitmap with no 'O' in it. A glyph is the one property of the eye that
	// the come-closer walk is allowed to change; the colour is not.
	EyeAlert = term.RGB{R: 232, G: 252, B: 226}
	// Amber, against the moonlit green of every other state. Colour does the
	// work that a five-cell face cannot.
	eyeWorried = term.RGB{R: 244, G: 176, B: 96}
)

// Cat is the companion. The body is a bitmap; the tail is a curve evaluated per
// frame, which is why wagging costs no extra authoring.
type Cat struct {
	// kind is which animal this is. The sprite data is a field rather than a
	// hardcoded bitmap, so a second companion needs no interface and no change
	// at any of the 25 construction sites outside this package.
	kind    Kind
	body    *Bitmap
	walk    *Bitmap
	worried *Bitmap
	// ask is the body held while the agent is blocked on the user. nil is the
	// shipped behaviour -- the ask is carried by the eye glyph alone -- and
	// giving the cat one is what a design round is deciding. See SetAskArt.
	ask *Bitmap
	// done is the body held after a turn closes. Like ask, nil falls through
	// to the shipped behaviour.
	done *Bitmap
	// step is the front-view mid-stride body.
	step     *Bitmap
	kitCache map[int]*Bitmap
	kitTier  int // last chosen litter size, for hysteresis
	swim     *Bitmap

	// mirror flips the companion and reverses the litter's layout, for a
	// composition anchored to the right edge instead of the left.
	mirror bool

	// face is the detail drawn over the body; coat is what the body is made of.
	face Face
	coat term.RGB
	// eyeFill is one of the EyeFill constants; see SetEyeFill.
	eyeFill string

	// noRim suppresses the cleared ring. Study-only; see SetNoRim.
	noRim bool

	// askN counts asks, and asking remembers whether the last frame was one,
	// so askN moves on the RISING EDGE only. It drives the rotating near eye;
	// see Approach.
	askN   int
	asking bool

	// stepping is set for the frames a pace step is in flight, so the legs
	// swap phase mid-stride. It is a per-frame flag rather than an argument
	// because Draw already carries five and the litter's draws would each
	// need it too.
	stepping bool

	// approach is how far along the come-closer walk the companion is: 0 at
	// home, 1 at the near pose. The one piece of state the companion carries
	// BETWEEN frames -- everything else here is recomputed every draw -- and
	// it has to be, because a walk is a thing that takes time. Driven by
	// Approach and read by nearRung; see crab_near.go.
	approach float64
}

// SetStepping says a pace step is in flight this frame. See pace.go.
func (c *Cat) SetStepping(v bool) { c.stepping = v }

// SetFace chooses how much detail the companion's face carries.
func (c *Cat) SetFace(f Face) { c.face = f }

// SetCoat chooses the body colour.
func (c *Cat) SetCoat(col term.RGB) { c.coat = col }

// Eye fills. The eyes are characters plotted in gaps of the body bitmap, so
// by default an eye cell shows the SCENE behind the head: dark water at
// night, which reads as a socket with a shine, and bright sea by day, which
// read as two blue holes (his 12:49 screengrab, 2026-09-04). A fill gives the
// cell fur before the glyph, so the eye is a mark on the face at every hour.
const (
	EyeFillNone   = ""       // the scene shows through (the original)
	EyeFillCoat   = "coat"   // the coat itself
	EyeFillSocket = "socket" // the coat darkened a step: a socket, not a hole
)

// SetEyeFill picks one of the fills above.
func (c *Cat) SetEyeFill(fill string) { c.eyeFill = fill }

// eyeGround is the fill's colour, and whether there is one.
func (c *Cat) eyeGround() (term.RGB, bool) {
	switch c.eyeFill {
	case EyeFillCoat:
		return c.coat, true
	case EyeFillSocket:
		return term.RGB{R: uint8(float64(c.coat.R) * 0.78), G: uint8(float64(c.coat.G) * 0.78),
			B: uint8(float64(c.coat.B) * 0.78)}, true
	}
	return term.RGB{}, false
}

func NewCat() *Cat {
	return &Cat{
		body:    ParseBitmap(CatBody),
		worried: ParseBitmap(CatWorried),
		// His rulings of 2026-09-11: the cat asks with its ears and finishes
		// with its chin up. Before this the ask and the finish were the WORKING
		// body -- measured, 97% of a still of either was also a still of the
		// cat simply working.
		ask:  ParseBitmap(CatAsk),
		done: ParseBitmap(CatDone),
		step: ParseBitmap(CatBodyStep),
		coat: furCol,
	}
}

// Size is the character footprint, not the pixel size.
func (c *Cat) Size() (w, h int) { return c.body.W / 2, c.body.H / 4 }

// catEyeCells are the cat's eye cells in the authored, unmirrored box. Shared
// with HeadCol so the balloon and the face cannot drift apart.
var catEyeCells = [2]int{2, 6}

// HeadCol is the head's column within the sprite's own box, and it is where a
// balloon's pointer belongs.
//
// Measured, not guessed: the cat's eyes sit at cells 2 and 6 of twelve, at 9
// and 5 once mirrored; the crab's at 4 and 7, and the crab does not flip. The
// midpoint is the cell between them either way.
func (c *Cat) HeadCol() int {
	if c.kind == KindCrab {
		return (crabEyeCells[0] + crabEyeCells[1] + 1) / 2
	}
	a, b := catEyeCells[0], catEyeCells[1]
	if c.mirror {
		w, _ := c.Size()
		a, b = w-1-a, w-1-b
	}
	return (a + b) / 2
}

// FaceLeft mirrors the whole companion.
//
// The sprite is not symmetric: the body sits in the left nine of its twelve
// columns and the right three are clearance for the tail, which sweeps up from
// the right hip. Placed against the right edge of the frame the tail is pinned
// to the wall, so a companion on the right has to be flipped -- then the tail
// sweeps INTO the scene, which is also the direction the eye is travelling
// after reading the sand.
func (c *Cat) FaceLeft(v bool) { c.mirror = v }

// Draw composes this frame and plots it. Everything that moves is computed
// here; nothing is stored between frames.
func (c *Cat) Draw(l *canvas.Layer, x, y int, t float64, st State) {
	if c.kind == KindCrab {
		c.drawCrab(l, x, y, t, st)
		return
	}
	// The come-closer walk. Same gate the crab uses: OFF unless XSCAPES_NEAR is
	// set AND the approach has actually climbed a rung, so an armed flag on its
	// own still renders the shipped frame. l.W is the frame's own width, so the
	// draw path and DrawnBox cannot disagree about which rung is affordable.
	if rung := c.nearRung(l.W); rung > 0 {
		c.drawCatNear(l, x, y, t, st, rung)
		return
	}
	src := c.body
	switch {
	case st == Worried:
		src = c.worried
	case st == NeedsYou && c.ask != nil:
		src = c.ask
	case st == Done && c.done != nil:
		src = c.done
	case c.stepping:
		// The FRONT view mid-stride, not the side-view walk sprite. See
		// CatBodyStep: reaching for CatWalk here drew a 16-cell side view
		// inside the 12-cell box and put the eyes 6 and 10 cells off the head.
		src = c.step
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
	case Done:
		// Content is the full tail held high and still, against NeedsYou's
		// quick everything -- position, not rate, so a screenshot keeps it.
		period, wag = 3.0, 0
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
	c.tailAt(f, wag, lift, tailLen, 1)
	if c.mirror {
		f = f.Mirrored()
	}

	// The cleared ring. His ruling, 2026-09-11. The cat's own kittens have
	// called plotRim since they were written (kittens.go) and the PARENT was
	// the odd one out -- which is the case the ring exists for, because a
	// litter sits right beside its parent in the same cream at the same alpha
	// on the same layer, and two of those read as one malformed shape.
	q := f.ToQuadrant()
	if !c.noRim {
		plotRim(l, q, x, y)
	}
	(&Sprite{Rows: q, Body: c.coat}).Draw(l, x, y)
	c.drawFace(l, x, y, c.face, st)
	c.eyes(l, x, y, t, st)
}

// tailAt sweeps a curve up from the right hip. Two pixels thick at 1x so it
// survives the halving that ToQuadrant does.
//
// EVERY LENGTH HERE SCALES, including the stroke. The tail is the one part of
// this animal that is not a bitmap, which is exactly why the come-closer work
// has to touch it: nearDouble doubles BITMAPS, so a cat drawn at a near rung by
// doubling CatBody arrives with no tail at all. Measured at 2x with the stroke
// left at 2px, the curve also thins to a hairline the quadrant halving drops in
// places, so the thickness is a function of s and not a constant.
func (c *Cat) tailAt(b *Bitmap, wag float64, lift int, length float64, scale float64) {
	if scale < 1 {
		scale = 1
	}
	// FRACTIONAL on purpose. The near ladder's middle rung is 32 source
	// columns against the shipped 24 -- a scale of 4/3, not 2 -- so an
	// integer-only tail lands the curve mid-body there instead of on the hip.
	s := scale
	// The stroke is stamped in whole pixels, so its thickness rounds.
	th := int(s + 0.5)
	if th < 1 {
		th = 1
	}
	n := int(13 * length * s)
	steps := 13 * s
	for i := 0; i <= n; i++ {
		f := float64(i) / steps
		x := int(17*s+0.5) + int(5*s*f+2.2*s*math.Sin(wag*1.6+f*2.6)+0.5)
		y := int(24*s+0.5) - int(13*s*f+0.5) + lift
		// The stamp is an L, not a square: (x,y), (x+1,y), (x,y+1) at 1x.
		// Scaling it as a full square instead changed the shipped cat, and
		// crab_near_test.go's golden hash caught it on the first run -- the
		// bottom-right corner is what makes the curve taper rather than read
		// as a chain of blocks.
		for dy := 0; dy < 2*th; dy++ {
			for dx := 0; dx < 2*th; dx++ {
				if dx < th || dy < th {
					b.Set(x+dx, y+dy)
				}
			}
		}
	}
}

// TailAt draws the tail onto an arbitrary bitmap at a given scale, so a study
// can render a near rung without the tail simply going missing.
func (c *Cat) TailAt(b *Bitmap, wag float64, lift int, length float64, scale float64) {
	c.tailAt(b, wag, lift, length, scale)
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
		glyph, col = 'O', EyeAlert
	case Worried:
		glyph, col = 'o', eyeWorried
	case Done:
		glyph = '^' // content -- closed happy, so it cannot read as alert
	}
	if st != Resting && st != Worried && st != Done && math.Mod(t, 5.3) < 0.16 {
		glyph = '-' // blink
	}
	// The eyes are plotted as characters on top of the quadrant body, so they
	// have to be mirrored by hand: the flip happens to the bitmap, not here.
	a, b := catEyeCells[0], catEyeCells[1]
	if c.mirror {
		w, _ := c.Size()
		a, b = w-1-a, w-1-b
	}
	if ground, ok := c.eyeGround(); ok {
		l.PlotOn(x+a, y+2, glyph, col, ground, 1)
		l.PlotOn(x+b, y+2, glyph, col, ground, 1)
		return
	}
	l.Plot(x+a, y+2, glyph, col, 1)
	l.Plot(x+b, y+2, glyph, col, 1)
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
