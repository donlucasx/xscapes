package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE ANIMATED MOCKUPS. His rulings of 2026-09-24, prompted through one by
// one: a session arc per clip · dusk into night over the loop · 10 fps · on
// the review page · drops running down the glass, a dry pane at idle · the
// block's life (a beacon, a window switching, a plane, chimney smoke) · the
// park's plaza with a fountain as the work and the trees moving with the
// wind, the wind the work · the harbour's full traffic and a lighthouse ·
// the kittens' doors (the bulkhead, the terrace's side, the moored boat) ·
// the sky greyed toward overcast with the shore's disc sinking · the lit
// windows the world's, todos stars · the cat's walk-up one rung · no
// default view yet ("these are all mockups for now").
//
// Nothing here is wired to the reducer: the arc is a script of the same
// session the site's clips show, a function of time.

// RainClipSecs is the clip's length; RainClipFPS its rate (10 on the review
// page, his ruling; 30 for the video export, which re-renders every frame
// rather than holding each page frame three times).
const RainClipSecs = 16.0

var RainClipFPS = 10.0

// rainArc is the session at one moment of the clip.
type rainArc struct {
	level, tod, ctx float64
	state           companion.State
	bubble          string
	ask             bool
	tail            []string
	kits            []kitPos
}

// kitPos is one kitten's cell this frame (the sprite's top-left), or none
// if it is out of sight below the sill.
type kitPos struct {
	x, y   int
	on     bool
	mirror bool // facing left
	lift   bool // the breath: the body up half a cell
	blink  bool
}

// The kittens' places on the sill, inside the room, left of the cat (his
// ruling of 2026-09-24: "the kittens could poke from behind the window as
// they appear, first poking their eyes and ears and eventually hopping on
// top of the window sill"). The same in every view; the plant on the
// harbour sits left of them.
var sillHomes = [][2]int{{28, 16}, {38, 16}, {48, 16}}

// sillRise is a kitten's top row through its arrival: below the sill,
// then the ears over its edge, then the eyes, a held look, and the hop up
// to sitting. u runs 0..1 over the walk.
func sillRise(u float64) (y int, on bool) {
	switch {
	case u < 0.12:
		return rainSill, false
	case u < 0.38:
		return rainSill - 1, true // the ears
	case u < 0.66:
		return rainSill - 2, true // the eyes, looking
	case u < 0.82:
		return rainSill - 3, true // the hop
	default:
		return rainSill - 4, true // sitting
	}
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// ramp is 0 before a, 1 after b, linear between.
func ramp(t, a, b float64) float64 { return clamp01((t - a) / (b - a)) }

// rainArcAt is the script: idle, a prompt, the work rising, subagents out
// through the door, an ask with the walk-up, the work again, the finish,
// the litter home, idle. The hour runs dusk into night; the context fills.
func rainArcAt(t float64, view int) rainArc {
	a := rainArc{}
	a.tod = 0.78 + 0.2*t/RainClipSecs
	a.ctx = 0.15 + 0.35*t/RainClipSecs
	switch {
	case t < 2:
		a.state, a.level = companion.Resting, 0
	case t < 8:
		a.state = companion.Working
		a.level = 0.8 * ramp(t, 2, 5)
	case t < 10:
		a.state, a.level = companion.NeedsYou, 0.4
		a.bubble, a.ask = "allow Bash?", true
	case t < 12.5:
		a.state, a.level = companion.Working, 0.9
	case t < 14:
		a.state = companion.Done
		a.level = 0.9 * (1 - ramp(t, 12.5, 13.5))
		a.bubble = "Rate limiting is in."
	default:
		a.state, a.level = companion.Resting, 0
	}
	lines := []struct {
		at float64
		s  string
	}{
		{3.0, "search  rate.Limiter  3 files"},
		{5.0, "edit  internal/auth/handler.go  +18 -2"},
		{7.5, "shell  go  exit 1"},
		{10.5, "test  go  ok 1.4s"},
	}
	for _, ln := range lines {
		if t >= ln.at {
			a.tail = append(a.tail, ln.s)
		}
	}
	if len(a.tail) > 3 {
		a.tail = a.tail[len(a.tail)-3:]
	}
	arrivals := []float64{4.0, 5.5, 7.0}
	exits := []float64{13.2, 12.6, 12.0}
	const walk = 1.6
	for i := range arrivals {
		home := sillHomes[i]
		k := kitPos{x: home[0], mirror: i%2 == 1}
		// The breath and the blink on their own periods, as the shore's
		// litter has them, so no two kittens move in step.
		breathT := 1.9 + scape.HashF(i, 1, 7)*1.6
		blinkT := 3.4 + scape.HashF(i, 2, 7)*3.8
		phase := scape.HashF(i, 4, 7) * 6.283
		k.lift = math.Sin(2*math.Pi*t/breathT+phase) > 0
		k.blink = math.Mod(t+phase, blinkT) < 0.17
		switch {
		case t < arrivals[i] || t > exits[i]+walk:
		case t < arrivals[i]+walk:
			k.y, k.on = sillRise(ramp(t, arrivals[i], arrivals[i]+walk))
			k.lift = false
		case t < exits[i]:
			k.y, k.on = home[1], true
			// The wander (his note: "the kittens should be able to move
			// around a bit after appearing"): every few seconds a new spot
			// two cells either way in the kitten's own lane, a short walk to
			// it facing the way it goes, then sitting.
			seg := 3.0 + scape.HashF(i, 5, 7)*1.5
			since := t - (arrivals[i] + walk)
			n := int(since / seg)
			off := func(n int) int {
				if n < 0 {
					return 0
				}
				return int(scape.HashF(i, 20+n, 7)*5) - 2
			}
			from, to := off(n-1), off(n)
			u := ramp(since-float64(n)*seg, 0, 0.8)
			k.x = home[0] + lerpi(from, to, u)
			if to != from && u < 1 {
				k.mirror = to < from
				k.lift = int(u*8)%2 == 1 // the stride
			} else if to != from {
				k.mirror = to < from
			}
		default:
			k.y, k.on = sillRise(1 - ramp(t, exits[i], exits[i]+walk))
			k.lift = false
		}
		a.kits = append(a.kits, k)
	}
	return a
}

func lerpi(a, b int, u float64) int { return a + int(math.Round(float64(b-a)*u)) }

// The clip's shared state the painters read (zero in the stills).
var (
	rainAnim bool
	rainT    float64
	rainWind float64
	rainKits []kitPos
	rainCtx  float64
	rainCat  *companion.Cat
	rainArcN rainArc
)

// PaintRainFrame paints one frame of a view's clip at time t.
func PaintRainFrame(c *canvas.Canvas, view int, t float64, seed int64) {
	if rainCat == nil {
		rainCat = companion.NewCat()
		rainCat.FaceLeft(true)
	}
	arc := rainArcAt(t, view)
	rainAnim, rainT, rainWind, rainKits, rainCtx, rainArcN = true, t, arc.level, arc.kits, arc.ctx, arc
	defer func() { rainAnim, rainKits = false, nil }()
	rainCat.Approach(1.0/RainClipFPS, arc.state == companion.NeedsYou)
	was := RainView
	RainView = view
	defer func() { RainView = was }()
	c.Clear()
	paintRainStudy(c, arc.tod, t, arc.level, seed, rainStageRoof, len(arc.kits))
}

// overcast greys the shore's sky a little toward cloud, keeping the hour
// (his ruling: "greyed toward overcast").
func overcast(p scape.Palette) scape.Palette {
	q := p
	q.SkyTop = term.Lerp(p.SkyTop, grey(9), 0.3)
	q.SkyHorizon = term.Lerp(p.SkyHorizon, grey(11), 0.22)
	q.StarVis *= 0.6
	return q
}

// drawDisc is the shore's moon for the mockup: a disc sampled two halves a
// cell, a second disc sliding across it for the phase (clear of it when the
// context is fresh, concentric when it is spent), the unlit face three
// steps over the sky. It sinks from row 2 to row 10 as the context fills,
// and the skyline, painted after, occludes it.
func drawDisc(c *canvas.Canvas, p scape.Palette, ctx float64) {
	cx, cy := 13.0, 2.6+8*ctx
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
			for h := 0; h < 2; h++ {
				py := float64(y) + 0.25 + 0.5*float64(h)
				if in(px, py, 0) {
					set[h] = true
					half[h] = body
					if in(px, py, off) {
						half[h] = dark
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

// paintDrops is the rain on the glass as drops running down it (his
// ruling over the study's streaks): a slot is a column with its own speed
// and phase; a drop is born at the top of the pane when the slot's clock
// wraps, and only if the work at that moment reaches the slot's threshold,
// so the pane is dry at idle, fills as the work rises, and DRIES after it
// stops as the last drops run out (a dry pane at idle, his ruling). Each
// drop is a head and a trail that fades behind it.
func paintDrops(c *canvas.Canvas, t float64, seed int64, view int) {
	mid := c.Mid()
	head, trail := cube(175, 215, 255), cube(135, 175, 215)
	const slots, tail = 130, 4
	rows := float64(rainPaneBot + 1 + tail)
	for k := 0; k < slots; k++ {
		x := 2 + int(scape.HashF(k, 201, seed)*76)
		speed := 3.5 + scape.HashF(k, 202, seed)*3.5
		period := rows / speed
		phase := scape.HashF(k, 203, seed) * period
		n := math.Floor((t + phase) / period)
		born := n*period - phase
		if born < -period {
			continue
		}
		lvl := 0.0
		if born >= 0 {
			lvl = rainArcAt(born, view).level
		}
		if scape.HashF(k, 204+int(n), seed) >= lvl*1.1 {
			continue
		}
		yh := (t - born) * speed
		yi := int(yh) - 1
		if yi >= 0 && yi <= rainPaneBot {
			plot(mid, x, yi, 'o', head, 0.7)
		}
		for j := 1; j <= tail; j++ {
			y := yi - j
			if y < 0 || y > rainPaneBot {
				continue
			}
			plot(mid, x, y, '|', trail, 0.55-0.11*float64(j))
		}
	}
}

// drawClipCat is the cat with the arc's state and its balloon: the walk-up
// one rung on the ask (companion.Near = 1, set by the tool), the ask
// balloon warm and solid, the done knock cool and dotted, both over a
// darkened ground so the letters read on any view.
func drawClipCat(c *canvas.Canvas, t float64) {
	arc := rainArcN
	near := c.Near()
	rainCat.Draw(near, rainCatX, rainCatY, t, arc.state)
	if arc.bubble == "" {
		return
	}
	rows := companion.DoneBubble(arc.bubble)
	col := cube(175, 215, 255)
	if arc.ask {
		rows, col = companion.Bubble(arc.bubble), cube(255, 215, 135)
	}
	rows = companion.MirrorTail(rows)
	head := rainCatX + rainCat.DrawnHeadCol(c.W)
	x := head - companion.TailCol(rows)
	if x+len([]rune(rows[0])) > c.W-2 {
		x = c.W - 2 - len([]rune(rows[0]))
	}
	if x < 2 {
		x = 2
	}
	y := rainCatY - len(rows)
	for dy, row := range rows {
		for dx := range []rune(row) {
			c.SetBG(x+dx, y+dy, term.Lerp(c.BGAt(x+dx, y+dy), rgb{}, 0.55))
		}
	}
	(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).Draw(near, x, y)
}

// writeTail is the arc's activity tail on the desk, newest brightest.
func writeTail(c *canvas.Canvas, y0 int, surface rgb, lines []string) {
	fill(c, 0, y0, c.W-1, c.H-1, surface)
	ink := inkFor(surface)
	n := len(lines)
	for i, s := range lines {
		a := 1.0
		if n > 1 {
			a = 0.55 + 0.45*float64(i)/float64(n-1)
		}
		text(c.Near(), 2, y0+i, s, ink, a)
	}
}

// swayAt is the wind's shift of a crown's row: nothing at rest, a cell or
// two streaming at flat out (the wind is the work, his ruling), the upper
// rows more than the lower.
func swayAt(cx, dy, ry int) int {
	if !rainAnim || rainWind <= 0 {
		return 0
	}
	amp := 1.8 * rainWind
	if dy > 0 {
		amp *= 0.4
	}
	return int(math.Round(amp * (0.6 + 0.4*math.Sin(rainT*1.9+float64(cx)*0.37))))
}

// drawKits draws the frame's kittens where the arc put them, if the arc is
// running; the stills keep their own placement.
func drawKits(c *canvas.Canvas) bool {
	// In the clip the kittens are drawn by drawKitsNow, after the frame:
	// they sit on the sill inside the room, in front of the mullion like
	// the cat (a kitten mid-wander was cut by it in the first video still).
	return rainAnim
}

// drawKitsNow draws the frame's kittens where the arc put them.
func drawKitsNow(c *canvas.Canvas) {
	if !rainAnim {
		return
	}
	near := c.Near()
	for i, k := range rainKits {
		if !k.on {
			continue
		}
		drawKittenAnim(near, k.x, k.y, k.mirror, k.lift, k.blink)
		// Whatever hangs below the sill's edge is behind it: cleared, so a
		// kitten rising shows its ears first and its body only once it has
		// hopped up. The sill and the desk are painted after this.
		_ = i
		for y := rainSill; y < k.y+4; y++ {
			for x := k.x - 1; x <= k.x+6; x++ {
				near.Erase(x, y)
			}
		}
	}
}
