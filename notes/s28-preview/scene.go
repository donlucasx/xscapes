package main

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The composition, replicated from live.go. compose(), bubbleX() and
// drawSand() are package main at the repo root and cannot be imported, so they
// are copied here with their anchors intact. Everything they depend on --
// paceSpan's strip, the 12-cell box, the h-2-chh anchor -- is copied with them,
// because getting any one of those wrong previews a composition he is not
// being offered.

// catBoxW is the box EVERY companion is drawn into, and what compose() sizes
// the layout from. It stays 12 whatever the sprite does: a 16-cell sprite is
// drawn at catX-2 and overhangs the box by two on each side.
const catBoxW = 12

// shippedH is the shipped sprite's cell height, and the anchor's only input.
// drawScene draws the companion at c.H-2-chh with chh from Size(); keeping
// this at 7 for both poses is what makes the near pose grow DOWNWARD into the
// two rows already under the shipped crab's feet, rather than upward into the
// sea.
const shippedH = 7

// bubbleAskCol is sheet.go's ask balloon colour, verbatim.
var bubbleAskCol = term.RGB{R: 244, G: 198, B: 122}

type layout struct {
	CatX     int
	PaceSpan int
	SandFrom int
	SandTo   int
	MoonX    float64
}

// paceSpan, verbatim from pace.go.
func paceSpan(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

// compose is live.go's compose() for the mirrored layout, which is the shipped
// one. The unmirrored branch is not reachable here.
func compose(w int) layout {
	const margin = 2
	right := margin + w/32
	span := paceSpan(w)
	catX := w - catBoxW - right
	if catX < 0 {
		catX = 0
	}
	return layout{
		CatX: catX, PaceSpan: span,
		SandFrom: margin, SandTo: catX - 1 - span,
		MoonX: 0.28,
	}
}

// bubbleX, verbatim from live.go: put the balloon's POINTER over the
// companion's head and let the rest of the box fall where it falls.
func bubbleX(rows []string, head, w int) int {
	x := head - companion.TailCol(rows)
	if bw := bubbleWidth(rows); x+bw > w {
		x = w - bw
	}
	if x < 0 {
		x = 0
	}
	return x
}

func bubbleWidth(rows []string) int {
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	return w
}

func luma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// warmShore returns a Shore that has HISTORY.
//
// A fresh scape.Shore has none, and Update clamps any gap over a second to one
// nominal step -- so a single Update on a new Shore renders a sea that has had
// exactly 0.05 s to exist, with the tide still at zero. Two instruments in this
// project have been voided by that. 400 steps of 0.08 s is 32 seconds of sea,
// well past the tide's own easing constant.
func warmShore(w, h int, seed int64, act scape.Activity, moonX float64) (*scape.Shore, float64) {
	sh := scape.NewShore(seed, false)
	sh.MoonX = moonX
	t := 0.0
	scratch := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for k := 0; k < 400; k++ {
		t += 0.08
		sh.Update(scratch, t, act)
	}
	return sh, t
}

type frameSpec struct {
	w, h   int
	t      float64
	act    scape.Activity
	p      *pose
	bubble string
	litter int
	seed   int64
	tail   []reduce.Line
}

// buildFrame paints one composed frame, in drawScene's own order: companion,
// litter, balloon, sand. The readout is skipped -- it is silent below 40%
// context and this preview runs at 30%, so drawing it would be inventing a
// channel rather than previewing one.
func buildFrame(sh *scape.Shore, s frameSpec) (*canvas.Canvas, layout, int) {
	c := canvas.New(s.w, s.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	lay := compose(s.w)
	sh.MoonX = lay.MoonX
	sh.Update(c, s.t, s.act)

	top := s.h - 2 - shippedH

	// The companion. NeedsYou is stillFor(), so pace() returns 0 and the
	// sprite sits at its home column -- no pacing to model.
	s.p.draw(c.Near(), lay.CatX+s.p.xoff, top, s.t)

	// The litter, through the product's own code and at the real baseline.
	if s.litter > 0 {
		lit := companion.NewCrab()
		lit.FaceLeft(true)
		lit.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, s.litter, c.W-1,
			int(float64(c.H)*0.42)+1, sh.SandTop()-2, s.t, s.seed)
	}

	// The balloon. Anchored to lay.CatX + HeadCol() of the SHIPPED crab, which
	// is the anchor bubbleX has today -- the point of drawing the near pose at
	// catX-2 is that this number does not have to change.
	if s.bubble != "" {
		rows := companion.MirrorTail(companion.Bubble(s.bubble))
		x := bubbleX(rows, lay.CatX+s.p.headCol(), c.W)
		(&companion.Sprite{Rows: rows, Body: bubbleAskCol, Opaque: true}).
			Draw(c.Near(), x, top-len(rows))
	}

	drawSand(c, s.tail, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
	return c, lay, top
}

// shippedHeadCol is companion.Cat.HeadCol() for the crab: (4+7+1)/2.
func shippedHeadCol() int {
	c := companion.NewCrab()
	c.FaceLeft(true)
	return c.HeadCol()
}

// drawSand, from live.go. Ink sampled from the PAINTED background per row,
// never the palette's nominal sand.
func drawSand(c *canvas.Canvas, lines []reduce.Line, sand term.RGB, sandTop, xFrom, xTo int) {
	if len(lines) == 0 || xTo-xFrom < 12 {
		return
	}
	bad := term.RGB{R: 244, G: 176, B: 96}
	beachAt := func(row int) term.RGB {
		if row < 0 || row >= c.H || len(c.BG) < c.W*c.H {
			return sand
		}
		var r, g, b, n int
		for x := xFrom; x < xTo && x < c.W; x += 4 {
			p := c.BG[row*c.W+x]
			r, g, b, n = r+int(p.R), g+int(p.G), b+int(p.B), n+1
		}
		if n == 0 {
			return sand
		}
		return term.RGB{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n)}
	}
	room := c.H - sandTop
	if room < 1 {
		return
	}
	if len(lines) > room {
		lines = lines[len(lines)-room:]
	}
	top := c.H - len(lines)
	for i, ln := range lines {
		row := top + i
		if row < 0 || row >= c.H {
			continue
		}
		beach := beachAt(row)
		base := term.RGB{R: 244, G: 236, B: 220}
		if luma(beach) > 140 {
			base = term.RGB{R: 34, G: 26, B: 20}
		}
		if ln.Bad {
			base = bad
		}
		col := term.Lerp(base, beach, 0.10+0.62*ln.Age)
		x := xFrom
		for _, r := range []rune(companion.NarrowOnly(ln.Text)) {
			if x >= xTo {
				break
			}
			c.Near().Plot(x, row, r, col, 1)
			x++
		}
	}
}

// waterRowsUnder reports, for the frame just built, which of the companion's
// own rows are below the mean waterline -- the honest answer to "is the raised
// claw standing in the sea?". Counted from the shore's own SandTop, which is
// the same number drawScene hangs the sand off.
func waterRowsUnder(sandTop, top, spriteH int) (wet int, feet int) {
	feet = top + spriteH - 1
	for y := top; y <= feet; y++ {
		if y < sandTop {
			wet++
		}
	}
	return wet, feet
}

// dominantBG is the single background colour the terminal actually paints most
// often over a rectangle of a RENDERED frame.
//
// Not the mean of the stored backgrounds: those are true colours the cube never
// shows, and averaging a sky's ramp produces a tone that appears nowhere on
// screen. This resolves every cell exactly as Render does and returns the most
// common result, with the count, so a caller can say how big the sample was.
func dominantBG(c *canvas.Canvas, x0, y0, x1, y1 int) (col term.RGB, n, total int) {
	seen := map[term.RGB]int{}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if x < 0 || y < 0 || x >= c.W || y >= c.H {
				continue
			}
			_, _, bg := c.ResolveAt(x, y, prof)
			seen[bg]++
			total++
		}
	}
	for k, v := range seen {
		if v > n || (v == n && lessRGB(k, col)) {
			col, n = k, v
		}
	}
	return col, n, total
}

// lessRGB is only a tie-break, so the dominant colour is deterministic rather
// than whatever the map happened to hand back first.
func lessRGB(a, b term.RGB) bool {
	if a.R != b.R {
		return a.R < b.R
	}
	if a.G != b.G {
		return a.G < b.G
	}
	return a.B < b.B
}

// flatPanel is a canvas of one background colour, for looking at a silhouette
// with nothing else moving. The colour is measured off a real frame, never
// invented -- see section 1.
func flatPanel(w, h int, bg term.RGB) *canvas.Canvas {
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c.SetBG(x, y, bg)
		}
	}
	return c
}

// sampleTail is a plausible activity tail, the length drawSand expects.
func sampleTail() []reduce.Line {
	return []reduce.Line{
		{Text: "read   internal/auth/handler.go   142 lines", Age: 0.75},
		{Text: "edit   internal/auth/handler.go   +18 -2", Age: 0.5},
		{Text: "grep   rate.Limiter   3 files", Age: 0.25},
		{Text: "bash   go test ./internal/auth/   ok 0.41s", Age: 0.0},
	}
}
