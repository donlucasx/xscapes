package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE HARBOUR AS A SCAPE. The third view of the rainy window study (s45,
// notes/rain/README.md), his pick for the release with the engine host
// (2026-11-09: "The Harbour with the cat companion").
//
// The six slots, and what carries each:
//
//	light        the sky's own palette by the wall clock, greyed toward
//	             overcast; the quay's lamps and the lighthouse by night
//	sky          the real sky over the city across the water; the moon (the
//	             shore's disc, with its phase) sinks behind the skyline as
//	             the context fills
//	motion       THE RAIN ON THE GLASS: drops born only while the work
//	             reaches a column's threshold, so the pane is dry at idle
//	             and dries after the finish. The harbour's traffic (the
//	             ferry, the tug, the sails, the gulls, the lighthouse) is the
//	             world's and keeps its own pace whatever the agent does
//	surface      the water (the shore's SeaFar/SeaNear), the dock, the sill
//	accumulator  finished todos as stars; the activity tail on the desk
//	companion    the cat on the sill, right; the kittens come up from inside
//	             the room for subagents and wander on the sill
//
// The study painted at 80x24 with fixed rows; this is the same drawing on a
// layout (harbourLayoutFor), so it sits in any window.
type Harbour struct {
	Seed  int64
	Ascii bool
	// CatX is the column the composition gives the companion (the shore's
	// lay.CatX), so the balloon uses the shore's arithmetic.
	CatX int
	// Kittens and Exits are the reducer's litter (State.Kittens,
	// State.KittenExits), set by the composer before Update.
	Kittens int
	Exits   []float64

	lay        harbourLayout
	desk       term.RGB
	arrived    []float64 // when each sitting kitten came up, oldest first
	litterSeen bool
	levels     []levelAt // the work over the last minute, for the drops' births
}

type levelAt struct{ t, level float64 }

// HarbourMinRows is the height under which the pane has no room for a sky,
// the water and the dock between the frame's top and the sill.
const HarbourMinRows = 16

// NewHarbour is the live harbour.
func NewHarbour(seed int64, ascii bool) *Harbour { return &Harbour{Seed: seed, Ascii: ascii} }

func (h *Harbour) Name() string { return "harbour" }

// harbourLayout is the frame's geometry for one window.
type harbourLayout struct {
	W, H               int
	hz                 int // the horizon: the last sky row
	waterTop, waterBot int
	dockTop, paneBot   int
	sill, bandTop      int
	rail, mull         int // the meeting rail's row, the mullion's left column
	catX, catY         int
	plantX             int
	moonX              int
}

// sx scales a column of the 80-column design to this width.
func (l harbourLayout) sx(x int) int { return int(math.Round(float64(x) * float64(l.W) / 80)) }

// wy maps a water row of the design (9..14) into this window's water.
func (l harbourLayout) wy(y int) int {
	n := l.waterBot - l.waterTop
	return l.waterTop + int(math.Round(float64(y-9)*float64(n)/5))
}

// harbourLayoutFor scales the 80x24 design: the writing keeps its rows on
// the desk (three, four from 28 rows, five from 40, as the vista's band),
// the sill is the row above it, and the pane above that is shared between
// the sky, the water and the dock in the design's proportions (9:6:5).
func harbourLayoutFor(w, h, catX int) harbourLayout {
	l := harbourLayout{W: w, H: h}
	band := 3
	if h >= 28 {
		band = 4
	}
	if h >= 40 {
		band = 5
	}
	l.bandTop = h - band
	l.sill = l.bandTop - 1
	pane := l.sill // rows 0..sill-1
	dock := clampInt(int(math.Round(float64(pane)*5/20)), 3, 6)
	water := max(3, int(math.Round(float64(pane-dock)*6/15)))
	sky := max(3, pane-dock-water)
	l.hz = sky - 1
	l.waterTop, l.waterBot = sky, sky+water-1
	l.dockTop, l.paneBot = l.waterBot+1, pane-1
	l.rail = int(math.Round(float64(pane) * 6 / 20))
	l.mull = w/2 - 1
	if catX == 0 {
		catX = w - 16
	}
	l.catX, l.catY = catX, l.sill-7
	l.plantX = l.sx(12)
	l.moonX = l.sx(13)
	return l
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Layout is the frame's geometry after the last Update: the cat's cell and
// the first row of the writing.
func (h *Harbour) Layout() (catX, catY, bandTop int) {
	return h.lay.catX, h.lay.catY, h.lay.bandTop
}

// DeskColor is the writing's ground.
func (h *Harbour) DeskColor() term.RGB { return h.desk }

// discAt is the moon's centre for the context used: it sinks from near the
// top of the sky to below the horizon, where the skyline hides it.
func (l harbourLayout) discAt(ctx float64) (cx, cy float64) {
	return float64(l.moonX), 2.6 + float64(l.hz)*ctx
}

// MoonAt is the cell the context readout centres under: the row under the
// disc's bottom.
func (h *Harbour) MoonAt(ctx float64) (x, y int) {
	cx, cy := h.lay.discAt(ctx)
	return int(cx) - 1, int(cy+1.35-0.5) - 1
}

// Update paints the frame but the cat, its balloon and the writing, which
// the live composer draws over it.
func (h *Harbour) Update(c *canvas.Canvas, t float64, act scape.Activity) {
	c.Clear()
	h.lay = harbourLayoutFor(c.W, c.H, h.CatX)
	l := h.lay
	p := overcast(scape.PaletteAt(act.TimeOfDay))
	lt := lit(scape.PaletteAt(act.TimeOfDay))
	tn := rainTonesAt(lt)
	h.desk = tn.desk
	h.remember(t, act.Level)

	// 1. The sky, its stars, the moon.
	far := c.Far()
	vramp(c, 0, 0, c.W-1, l.hz, p.SkyTop, p.SkyHorizon)
	for y := 0; y <= l.hz-3; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, h.Seed) < 0.05*p.StarVis {
				plot(far, x, y, '.', p.Star, 0.6*p.StarVis)
			}
		}
	}
	cx, cy := l.discAt(act.ContextUsed)
	drawDiscAt(c, p, act.ContextUsed, cx, cy)

	// 2. The view: the city, the water, the traffic, the dock.
	h.paintView(c, p, lt, t)

	// 3. Finished todos, as stars over everything in the sky.
	h.todoStars(c, p, act.TodoDone)

	// 4. The glass: the rain is the work.
	h.paintDrops(c, t)

	// 5. The frame: posts, the mullion, the meeting rail. Wood: nothing
	// outside shows on it.
	frameCell := func(x, y int, col rgb) {
		if x < 0 || x >= c.W {
			return
		}
		setSolid(c, x, y, col)
		c.Far().Erase(x, y)
		c.Mid().Erase(x, y)
		c.Near().Erase(x, y)
	}
	rabbet := clampGrey(greyK(tn.frame) - 3)
	for y := 0; y <= l.paneBot; y++ {
		frameCell(0, y, tn.frame)
		frameCell(1, y, rabbet)
		frameCell(l.mull, y, rabbet)
		frameCell(l.mull+1, y, tn.frame)
		frameCell(c.W-2, y, rabbet)
		frameCell(c.W-1, y, tn.frame)
	}
	for x := 0; x < c.W; x++ {
		frameCell(x, l.rail, tn.frame)
	}

	// 6. Inside: the kittens coming up over the sill, the sill, the plant,
	// the desk.
	h.drawLitter(c, t)
	fill(c, 0, l.sill, c.W-1, l.sill, tn.sill)
	if c.W >= 60 {
		drawSillPlantAt(c, l.plantX, l.sill-20, lt)
	}
	fill(c, 0, l.bandTop, c.W-1, c.H-1, tn.desk)
}

// remember keeps the work level over the last minute, so a drop's birth
// reads the level at the moment it was born, however late it is drawn.
func (h *Harbour) remember(t, level float64) {
	if n := len(h.levels); n > 0 && t < h.levels[n-1].t {
		h.levels = h.levels[:0] // the clock went back: a new session
	}
	h.levels = append(h.levels, levelAt{t, level})
	cut := 0
	for cut < len(h.levels) && h.levels[cut].t < t-60 {
		cut++
	}
	h.levels = h.levels[cut:]
}

// levelAtTime is the work at time t, or zero before what is remembered.
func (h *Harbour) levelAtTime(t float64) float64 {
	lv := 0.0
	for _, s := range h.levels {
		if s.t > t {
			break
		}
		lv = s.level
	}
	return lv
}

// paintDrops is the study's rain on the glass (paintDrops), on the layout:
// a slot is a column with its own speed and phase; a drop is born at the
// top of the pane when the slot's clock wraps, and only if the work at that
// moment reaches the slot's threshold.
func (h *Harbour) paintDrops(c *canvas.Canvas, t float64) {
	mid := c.Mid()
	head, trail := cube(175, 215, 255), cube(135, 175, 215)
	const tail = 4
	slots := 130 * c.W / 80
	rows := float64(h.lay.paneBot + 1 + tail)
	for k := 0; k < slots; k++ {
		x := 2 + int(scape.HashF(k, 201, h.Seed)*float64(c.W-4))
		speed := 3.5 + scape.HashF(k, 202, h.Seed)*3.5
		period := rows / speed
		phase := scape.HashF(k, 203, h.Seed) * period
		n := math.Floor((t + phase) / period)
		born := n*period - phase
		if scape.HashF(k, 204+int(n), h.Seed) >= h.levelAtTime(born)*1.1 {
			continue
		}
		yi := int((t-born)*speed) - 1
		if yi >= 0 && yi <= h.lay.paneBot {
			plot(mid, x, yi, 'o', head, 0.7)
		}
		for j := 1; j <= tail; j++ {
			y := yi - j
			if y < 0 || y > h.lay.paneBot {
				continue
			}
			plot(mid, x, y, '|', trail, 0.55-0.11*float64(j))
		}
	}
}

// todoStars lights one star per finished todo in the sky above the city,
// each at a place fixed by its index, kept off the moon's column and the
// spend counter.
func (h *Harbour) todoStars(c *canvas.Canvas, p scape.Palette, n int) {
	rows := h.lay.hz - 4
	if rows < 2 || n <= 0 {
		return
	}
	near := c.Near()
	const phi = 0.6180339887
	for i := 0; i < n && i < 64; i++ {
		fx := math.Mod(float64(i)*phi+scape.HashF(i, 7, h.Seed), 1)
		x := 3 + int(fx*float64(c.W-6))
		y := 1 + int(scape.HashF(i, 9, h.Seed)*float64(rows-1))
		if x >= h.lay.moonX-3 && x <= h.lay.moonX+3 {
			x = (x+7)%(c.W-6) + 3
		}
		if tx0, tx1, ty := scape.TokensGround(c.W); y == ty && x >= tx0 && x <= tx1 {
			y = ty + 1
		}
		g := c.BGAt(x, y)
		near.PlotOn(x, y, '*', vistaStarInk(c, x, y, p.Star), g, 1)
	}
}

// drawLitter is the reducer's kittens on the sill: each comes up from
// inside the room (the ears over the sill's edge, the eyes, a held look,
// the hop up) when it arrives and goes back down when it leaves, and
// wanders in its own lane while it sits. They sit left of the cat, nearest
// first, as many as there is room for before the plant.
func (h *Harbour) drawLitter(c *canvas.Canvas, t float64) {
	const walk = 1.6
	// The first frame takes the litter as it finds it, already sitting (a
	// scape opened mid-session); from then on every new kitten comes up over
	// the sill. "First frame" is its own flag: a litter of zero is not a
	// litter never seen, or the session's first kitten would appear sitting.
	if !h.litterSeen {
		h.litterSeen = true
		for len(h.arrived) < h.Kittens {
			h.arrived = append(h.arrived, t-walk)
		}
	}
	for len(h.arrived) < h.Kittens {
		h.arrived = append(h.arrived, t)
	}
	if len(h.arrived) > h.Kittens {
		h.arrived = h.arrived[:h.Kittens]
	}
	minX := 2
	if c.W >= 60 {
		minX = h.lay.plantX + 11
	}
	home := func(k int) int { return h.lay.catX - 16 - 10*k }
	near := c.Near()
	draw := func(k int, y int, on bool, x int, mirror, lift, blink bool) {
		if !on || x < minX || x+6 > h.lay.catX {
			return
		}
		drawKittenAnim(near, x, y, mirror, lift, blink)
		for yy := h.lay.sill; yy < y+4; yy++ {
			for xx := x - 1; xx <= x+6; xx++ {
				near.Erase(xx, yy)
			}
		}
	}
	sit := h.lay.sill - 4
	rise := func(u float64) (int, bool) {
		y, on := sillRise(u)
		return y - rainSill + h.lay.sill, on
	}
	for k, at := range h.arrived {
		x0 := home(k)
		mirror := k%2 == 1
		breathT := 1.9 + scape.HashF(k, 1, 7)*1.6
		blinkT := 3.4 + scape.HashF(k, 2, 7)*3.8
		ph := scape.HashF(k, 4, 7) * 6.283
		lift := math.Sin(2*math.Pi*t/breathT+ph) > 0
		blink := math.Mod(t+ph, blinkT) < 0.17
		if t < at+walk {
			y, on := rise(ramp(t, at, at+walk))
			draw(k, y, on, x0, mirror, false, blink)
			continue
		}
		seg := 3.0 + scape.HashF(k, 5, 7)*1.5
		since := t - (at + walk)
		n := int(since / seg)
		off := func(n int) int {
			if n < 0 {
				return 0
			}
			return int(scape.HashF(k, 20+n, 7)*5) - 2
		}
		from, to := off(n-1), off(n)
		u := ramp(since-float64(n)*seg, 0, 0.8)
		x := x0 + lerpi(from, to, u)
		if to != from {
			mirror = to < from
			if u < 1 {
				lift = int(u*8)%2 == 1
			}
		}
		draw(k, sit, true, x, mirror, lift, blink)
	}
	// The ones leaving go back down behind the sill, from the places past
	// the sitters.
	for i, e := range h.Exits {
		y, on := rise(1 - clamp01(e))
		draw(0, y, on, home(len(h.arrived)+i), i%2 == 0, false, false)
	}
}

// drawDiscAt is the study's drawDisc at a centre of the layout's choosing.
func drawDiscAt(c *canvas.Canvas, p scape.Palette, ctx, cx, cy float64) {
	rx, ry := 2.2, 1.35
	off := 2 * rx * (1 - ctx)
	body := term.Lerp(c.BGAt(int(cx), int(cy)), p.Moon, p.MoonVis*0.9)
	dark := term.RGB{R: 80, G: 80, B: 96}
	in := func(px, py, ox float64) bool {
		dx, dy := (px-cx-ox)/rx, (py-cy)/ry
		return dx*dx+dy*dy <= 1
	}
	for y := int(cy - ry - 1); y <= int(cy+ry+1); y++ {
		for x := int(cx - rx - 1); x <= int(cx+rx+1); x++ {
			if x < 0 || y < 0 || x >= c.W || y >= c.H {
				continue
			}
			px := float64(x) + 0.5
			var half [2]rgb
			var set [2]bool
			for hh := 0; hh < 2; hh++ {
				py := float64(y) + 0.25 + 0.5*float64(hh)
				if in(px, py, 0) {
					set[hh] = true
					half[hh] = body
					if in(px, py, off) {
						half[hh] = dark
					}
				}
			}
			switch {
			case set[0] && set[1]:
				if half[0] == half[1] {
					c.SetBG(x, y, half[0])
				} else {
					c.SetBGHalves(x, y, half[0], half[1])
				}
			case set[0]:
				c.SetBGHalves(x, y, half[0], c.BGAt(x, y))
			case set[1]:
				c.SetBGHalves(x, y, c.BGAt(x, y), half[1])
			}
		}
	}
}
