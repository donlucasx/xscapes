package main

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// A pose is one companion silhouette drawn the way the product draws one.
//
// The shipped crab goes through the PRODUCT: companion.NewCrab().Draw, the
// real thing, no copy. Everything larger has to go through a replica of
// companion's drawCrab, because that method is unexported and there is no
// other door into it. The replica is checked against the product cell for
// cell at the top of every run (see replicaCheck) -- an untested replica is
// exactly the kind of harness that hands back a clean bill of health for a
// picture the product would never paint.

// eyeAlert is copied verbatim from internal/companion/cat.go. It is the colour
// the ask eye is drawn in, and it is unexported there.
var eyeAlert = term.RGB{R: 232, G: 252, B: 226}

type pose struct {
	name string // "today's crabAsk", "the near ask"

	// product, when set, IS the product: draw delegates to companion's own
	// Draw and nothing here touches the pixels.
	product *companion.Cat

	// The replica path.
	upper, lower []string
	// eyeLeft is the LEFT cell of each eye in the AUTHORED (unmirrored) box.
	eyeLeft []int
	eyeRow  int
	eyeW    int // 1 for the shipped one-cell glyph, 2 for the bitmap eye
	// bm maps the shipped glyph vocabulary onto bitmaps, for eyeW == 2.
	bm map[rune][]string
	// ground draws the eye with PlotOn and the coat underneath, so the hole
	// in the ring is a socket rather than a window onto the sea.
	ground bool

	// state is the companion state this pose is drawn in. Set explicitly on
	// every constructor: companion.Resting is the zero value, so a forgotten
	// field would silently draw the wrong animal.
	state companion.State
	// forceGlyph pins the eye to one of 'O' 'o' '-' '^' instead of letting the
	// state and the blink pick it. For the eye vocabulary panel only.
	forceGlyph rune

	// xoff is where the sprite is drawn relative to lay.CatX. An even-width
	// sprite at catX-(W-12)/2 keeps the shipped head centre; see headReport.
	xoff int

	// crabEyes says this pose's eyes sit where the crab's do. False for the
	// cat, whose eye cells are its own -- so the head-column claim and the eye
	// probes leave it alone instead of measuring the wrong cells.
	crabEyes bool
}

// size is the cell footprint, computed the way companion.Size does it: the
// source bitmap's W/2 by H/4.
func (p *pose) size() (w, h int) {
	if p.product != nil {
		return p.product.Size()
	}
	b := companion.ParseBitmap(p.rows())
	return b.W / 2, b.H / 4
}

func (p *pose) rows() []string {
	return append(append([]string{}, p.upper...), p.lower...)
}

// draw paints this pose in its NeedsYou (ask) state at (x, y).
func (p *pose) draw(l *canvas.Layer, x, y int, t float64) {
	if p.product != nil {
		p.product.Draw(l, x, y, t, p.state)
		return
	}
	p.drawReplica(l, x, y, t)
}

// drawReplica is internal/companion's drawCrab, for the NeedsYou state, on a
// sprite of any size. Line for line with the original: compose upper onto
// lower, breathe by shifting two source rows, mirror, quadrantise, cut a rim,
// plot the body in the coat, then the eyes on top.
func (p *pose) drawReplica(l *canvas.Layer, x, y int, t float64) {
	src := companion.ParseBitmap(p.rows())

	// Breathing. One quadrant subpixel is two source rows, so a two-row shift
	// moves the body by half a character cell. NeedsYou's period is 1.6.
	const period = 1.6
	lift := 0
	if math.Sin(t*2*math.Pi/period) > 0.35 {
		lift = 2
	}
	f := src.Blank()
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if bitAt(src, sx, sy+lift) {
				f.Set(sx, sy)
			}
		}
	}
	f = f.Mirrored() // the shipped composition is mirrored; the crab always is
	q := f.ToQuadrant()
	plotRim(l, q, x, y)
	(&companion.Sprite{Rows: q, Body: companion.CrabCoat}).Draw(l, x, y)

	// The eyes. NeedsYou is 'O' in eyeAlert, and it still blinks -- drawCrab's
	// blink excludes Resting, Worried and Done, which leaves Working and
	// NeedsYou blinking.
	glyph := 'O'
	if math.Mod(t, 5.3) < 0.16 {
		glyph = '-'
	}
	if p.forceGlyph != 0 {
		glyph = p.forceGlyph
	}
	for _, sp := range p.eyeSpansAt(x) {
		p.plotEye(l, sp[0], p.eyeRowAt(y), glyph)
	}
}

// plotEye draws one eye, either as the shipped single character or as the
// 2x2 bitmap. PlotOn when the pose asks for a ground: the cell then owns its
// own background, which is what stops the hole in the ring showing the sea.
func (p *pose) plotEye(l *canvas.Layer, x, y int, glyph rune) {
	coat := companion.CrabCoat
	if p.eyeW == 1 {
		if p.ground {
			l.PlotOn(x, y, glyph, eyeAlert, coat, 1)
			return
		}
		l.Plot(x, y, glyph, eyeAlert, 1)
		return
	}
	rows := companion.ParseBitmap(p.bm[glyph]).ToQuadrant()
	for dy, r := range rows {
		for dx, ch := range []rune(r) {
			if ch == ' ' {
				continue
			}
			if p.ground {
				l.PlotOn(x+dx, y+dy, ch, eyeAlert, coat, 1)
				continue
			}
			l.Plot(x+dx, y+dy, ch, eyeAlert, 1)
		}
	}
}

// eyeSpansAt is where this pose's eyes land, in canvas columns, for a sprite
// drawn at x and mirrored. Returned as [left, right] cell pairs.
func (p *pose) eyeSpansAt(x int) [][2]int {
	if !p.crabEyes {
		return nil
	}
	w, _ := p.size()
	left, ew := p.eyeLeft, p.eyeW
	if p.product != nil {
		// The product's crab: its eye cells and row are unexported, so they
		// are stated here and checked against the rendered frame in section 4.
		left, ew = []int{4, 7}, 1
	}
	var out [][2]int
	for _, e := range left {
		ex := x + w - e - ew
		out = append(out, [2]int{ex, ex + ew - 1})
	}
	return out
}

// eyeSpanAbs is eyeSpansAt for a sprite placed by the scene, at catX+xoff.
func (p *pose) eyeSpanAbs(catX int) [][2]int { return p.eyeSpansAt(catX + p.xoff) }

// headCol is where this pose's balloon pointer is anchored, inside its own
// box. A product companion answers for itself (the cat's eyes are not the
// crab's); every replica pose deliberately keeps the SHIPPED crab's answer,
// because "the anchor does not have to change" is the claim being tested.
func (p *pose) headCol() int {
	if p.product != nil {
		return p.product.HeadCol()
	}
	return shippedHeadCol()
}

// eyeRowAt is the eye's cell row for a sprite drawn with its top at y.
func (p *pose) eyeRowAt(y int) int {
	if p.product != nil {
		return y + 2 // crabEyeRow
	}
	return y + p.eyeRow
}

// eyeCentre is the midpoint of the eye pair in absolute columns, as a float.
// This is the number the balloon's pointer is aimed at, and the whole reason
// the near pose is drawn at catX-2 rather than at catX.
func (p *pose) eyeCentre(catX int) float64 {
	sp := p.eyeSpanAbs(catX)
	if len(sp) == 0 {
		return math.NaN()
	}
	lo, hi := sp[0][0], sp[0][1]
	for _, s := range sp {
		if s[0] < lo {
			lo = s[0]
		}
		if s[1] > hi {
			hi = s[1]
		}
	}
	return (float64(lo) + float64(hi)) / 2
}

func bitAt(b *companion.Bitmap, x, y int) bool {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return false
	}
	return b.On[y*b.W+x]
}

// plotRim is internal/companion's plotRim: clear a one-cell ring around the
// silhouette so whatever is behind it is visibly cut rather than merging into
// it. Unexported there, copied here so the replica composites identically.
func plotRim(l *canvas.Layer, rows []string, x, y int) {
	h := len(rows)
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	filled := func(cx, cy int) bool {
		if cy < 0 || cy >= h {
			return false
		}
		r := []rune(rows[cy])
		return cx >= 0 && cx < len(r) && r[cx] != ' '
	}
	for cy := -1; cy <= h; cy++ {
		for cx := -1; cx <= w; cx++ {
			if filled(cx, cy) {
				continue
			}
			adj := false
			for dy := -1; dy <= 1 && !adj; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if filled(cx+dx, cy+dy) {
						adj = true
						break
					}
				}
			}
			if adj {
				l.Plot(x+cx, y+cy, ' ', term.RGB{}, 1)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// The three rungs of the approach.
// ---------------------------------------------------------------------------

// shippedPose is the real product, in its ask pose, mirrored as shipped.
func shippedPose() *pose {
	c := companion.NewCrab()
	c.FaceLeft(true)
	return &pose{name: "today's crabAsk", product: c, state: companion.NeedsYou, crabEyes: true, xoff: 0}
}

// workingPose is the shipped crab BEFORE the ask -- the frame the approach
// starts from.
func workingPose() *pose {
	c := companion.NewCrab()
	c.FaceLeft(true)
	return &pose{name: "working", product: c, state: companion.Working, crabEyes: true, xoff: 0}
}

// replicaShipped is the shipped crab's own rows put through the replica draw,
// so the replica can be checked against the product. Nothing draws this except
// replicaCheck.
func replicaShipped() *pose {
	return &pose{
		name:     "replica of crabAsk",
		upper:    crabAskRows(),
		lower:    crabLowerRows(),
		eyeLeft:  []int{4, 7}, // crabEyeCells
		eyeRow:   2,           // crabEyeRow
		eyeW:     1,
		state:    companion.NeedsYou,
		crabEyes: true,
		xoff:     0,
	}
}

// nearPose is the pick: 16 cells x 9 rows, drawn at catX-2 so its eye centre
// lands exactly where the shipped crab's does.
func nearPose() *pose {
	return &pose{
		name:    "the near ask",
		upper:   pickUpperRows(),
		lower:   pickLower,
		eyeLeft: []int{5, 9},
		eyeRow:  1,
		eyeW:    2,
		bm: map[rune][]string{
			'O': eyeAlertBM, 'o': eyeOpenBM, '-': eyeShutBM, '^': eyeDoneBM,
		},
		ground:   true, // PlotOn, coat as ground -- the hole is a socket
		state:    companion.NeedsYou,
		crabEyes: true,
		xoff:     -2, // -(16-12)/2
	}
}

// midPose is A's 14x9 from the same round, used ONLY as the in-between frame
// of the walk-up. It is not a proposal; it is a rung that was already drawn.
func midPose() *pose {
	return &pose{
		name:    "A 14x9 (in-between)",
		upper:   aUpper,
		lower:   aLower,
		eyeLeft: []int{4, 8},
		eyeRow:  2,
		eyeW:    2,
		bm: map[rune][]string{
			'O': aEye, 'o': eyeOpenBM, '-': eyeShutBM, '^': eyeDoneBM,
		},
		ground:   true,
		state:    companion.NeedsYou,
		crabEyes: true,
		xoff:     -1, // -(14-12)/2
	}
}

// catPose is the other shipped animal, for -animal cat. Its eye cells differ
// from the crab's, so eyeSpansAt is not meaningful for it and section 4 leaves
// it alone; it is here to compare SILHOUETTES.
func catPose() *pose {
	c := companion.NewCat()
	c.FaceLeft(true)
	return &pose{name: "the cat", product: c, state: companion.NeedsYou, xoff: 0}
}
