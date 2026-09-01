package scape

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// Shore: night beach. Stars and moon far, open water mid, foam and sand near.
type Shore struct {
	Seed  int64
	ASCII bool

	// Where the moon landed this frame, so callers can anchor a label to it
	// without recomputing the position and drifting out of sync.
	moonX, moonY int

	pal     Palette // this frame's colours, from the time of day
	ctxUsed float64 // this frame's context reading, for the moon and its shine

	// BlueSky pushes the sky into the pure-blue cube column. REJECTED as a
	// shipping option, kept only so the colour study can render the panel that
	// shows why.
	//
	// The reasoning was sound and the result is not. Only 7 of the 216 cube
	// colours are both genuinely dark (luma < 40) and genuinely coloured
	// (chroma >= 20), and five of those are one column -- #00005f through
	// #0000ff, pure blue with no red or green at all. So a dark blue sky IS
	// available on a 256-colour terminal. Rendered, it is an electric royal
	// blue that swallows the moon, and the moon is load-bearing: it carries
	// context remaining. See assets/frames/color-study.png, panel C.
	//
	// The conclusion is that 256 should not try to be a colour night. It is
	// already a decent monochrome one, and the signals that MUST read -- the
	// worried amber, the alert yellow, the calm cyan -- all survive
	// quantisation onto the cube with their hues intact. A black-and-white
	// night where the only colour on screen is the colour that means something
	// is a better answer than a garish blue one.
	BlueSky bool

	// MoonX is where the moon sits, as a fraction of the width. It moves with
	// the composition: the moon and the companion are the two things a glance
	// goes looking for, and stacking them in one column leaves half the frame
	// carrying nothing. Zero means the default.
	MoonX float64

	// lastEdge is the waterline Update most recently computed. Kept for the
	// same reason as moonX/moonY: it is the scene's real geometry, and a
	// caller (or a test) that recomputed it would drift out of sync with
	// what was actually painted.
	lastEdge []float64

	// phase is wave time, integrated rather than derived.
	//
	// This has to be state, and the reason only shows up once activity stops
	// being a constant. The obvious form -- tt := t * speed(Level) -- scales
	// the whole elapsed time, so raising Level does not speed the waves up,
	// it teleports them: at t=600s a 0.05 change in Level moved 447 of 1920
	// cells, against 44 cells at t=5s. The error grows with session age,
	// which is exactly backwards for a scene meant to run all day. Carrying
	// phase forward and adding dt*speed each frame keeps the wave field
	// continuous through any change in activity.
	phase float64
	lastT float64
}

// SandTop is the first row that is dry sand in every column.
//
// The activity tail has to hang off this rather than off the bottom of the
// canvas: anchored to the canvas it drifts upward as lines accumulate and ends
// up written across the water, which is both wrong and the exact opposite of
// what "written in the sand" means. The waterline moves with activity, so this
// moves too -- a busy sea reaches further up the beach and leaves less room to
// write, which is the right behaviour and not a bug.
func (s *Shore) SandTop() int {
	if len(s.lastEdge) == 0 {
		return 0
	}
	// The MEAN waterline, not the deepest. Taking the deepest means one wave
	// crest in one column pushes the whole block of text down the beach, and
	// at 80x24 with a busy sea that leaves zero rows to write in -- measured.
	// The mean is the brown band you actually see, and a crest washing over
	// the end of the oldest line is not a defect: it is the tide taking it,
	// which is what the brief asks the sand to show.
	var sum float64
	for _, v := range s.lastEdge {
		sum += v
	}
	return int(math.Ceil(sum/float64(len(s.lastEdge)))) + 1
}

// pureBlue moves a colour onto the cube's pure-blue column, keeping roughly
// its brightness. Red and green go to zero, because any red or green at all
// pulls the result back into the grey basin where the whole problem started.
func pureBlue(c term.RGB) term.RGB {
	l := 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	// Blue contributes 0.11 of luma, so matching brightness needs b = l/0.11,
	// which saturates fast -- clamp, and accept that a bright sky cannot be
	// pure blue.
	b := l / 0.11
	if b > 255 {
		b = 255
	}
	return term.RGB{R: 0, G: 0, B: uint8(b)}
}

// SandColor is the beach's colour this frame.
//
// The activity tail fades toward it as the tide takes each line, and that
// target has to come from the palette rather than a constant: pinned to one
// hour's sand it is exact at that hour and wrong at every other. It was pinned
// to midnight, so by mid-evening the oldest line sat 2.7 luma from the beach
// it was written on -- there, but unreadable.
func (s *Shore) SandColor() term.RGB { return s.pal.SandNear }

// MoonPos is the moon's centre cell from the last Update.
func (s *Shore) MoonPos() (x, y int) { return s.moonX, s.moonY }

func NewShore(seed int64, asciiOnly bool) *Shore { return &Shore{Seed: seed, ASCII: asciiOnly} }

func (s *Shore) Name() string { return "shore" }

func (s *Shore) Update(c *canvas.Canvas, t float64, act Activity) {
	c.Clear()
	s.pal = PaletteAt(act.TimeOfDay)
	s.ctxUsed = act.ContextUsed
	if s.BlueSky {
		s.pal.SkyTop = pureBlue(s.pal.SkyTop)
		s.pal.SkyHorizon = pureBlue(s.pal.SkyHorizon)
	}
	if c.W < 8 || c.H < 6 {
		return
	}
	hy := int(float64(c.H) * 0.42) // horizon
	// Mean waterline. Eight tenths down is right in a tall window and
	// collapses the beach in a short one -- at fourteen rows it left a single
	// row of sand, so three of the four lines of writing had nowhere to go.
	// Keep it proportional, but never let fewer than five rows of beach
	// survive. At 24 rows and above this computes exactly the number it always
	// did; it only bites on a short pane.
	sy := c.H - 5
	if p := c.H / 5; p > 5 {
		sy = c.H - p
	}
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > c.H-2 {
		sy = c.H - 2
	}

	// Everything sized in rows has to scale, or the scene that is composed at
	// 80x24 turns into a giant moon and a wall of foam at 40x12.
	scale := math.Min(float64(c.W)/80.0, float64(c.H)/24.0)
	scale = math.Max(0.45, math.Min(1.35, scale))

	// Working raises the sea; resting lets it settle. Integrate, do not scale:
	// see the note on Shore.phase.
	dt := t - s.lastT
	if dt < 0 || dt > 1 {
		// A jump backwards, or a gap longer than a frame -- a resized
		// terminal, a suspended process, or the first frame. Advance by one
		// nominal step rather than lurching the sea by the whole gap.
		dt = 0.05
	}
	s.lastT = t
	s.phase += dt * (0.55 + act.Level*1.45)
	tt := s.phase
	edge := s.waterline(c.W, sy, tt, act, scale)
	s.lastEdge = edge

	s.paintBG(c, hy, edge)
	s.stars(c, hy, t)
	s.moon(c, hy, scale, 1-clamp01(act.ContextUsed), s.pal.MoonVis)
	s.sea(c, hy, edge, tt, act)
	s.sand(c, edge)
	s.foam(c, edge, scale)
}

// waterline is where the sea meets the sand, per column, as a float. Keeping it
// fractional is the whole trick: the shoreline is painted into the background
// colour at sub-cell precision, so a two-row swing reads as one smooth curve
// instead of the staircase you get from rounding to a row.
func (s *Shore) waterline(w, sy int, tt float64, act Activity, scale float64) []float64 {
	reach := (0.8 + act.Level*2.1) * scale
	e := make([]float64, w)
	for x := 0; x < w; x++ {
		fx := float64(x)
		// Long wavelengths only — anything that cycles more than a couple of
		// times across 80 columns stops reading as water.
		e[x] = float64(sy) +
			reach*math.Sin(fx*0.11+tt*0.8) +
			0.55*scale*math.Sin(fx*0.047-tt*0.45)
	}
	return e
}

func (s *Shore) paintBG(c *canvas.Canvas, hy int, edge []float64) {
	fh := math.Max(1, float64(hy))
	for x := 0; x < c.W; x++ {
		ex := edge[x]
		for y := 0; y < c.H; y++ {
			fy := float64(y)
			var col term.RGB
			switch {
			case y <= hy:
				col = term.Lerp(s.pal.SkyTop, s.pal.SkyHorizon, fy/fh)
			case fy < ex-0.5:
				col = term.Lerp(s.pal.SeaFar, s.pal.SeaNear, (fy-float64(hy))/math.Max(1, ex-float64(hy)))
			case fy < ex+0.5:
				// The waterline cell straddles sea and sand. Mix by how much of
				// the cell the water actually covers.
				col = term.Lerp(s.pal.WetSand, s.pal.SeaNear, ex-fy+0.5)
			default:
				f := (fy - ex) / math.Max(1, float64(c.H-1)-ex)
				col = term.Lerp(s.pal.WetSand, s.pal.SandNear, math.Min(1, f*1.3))
			}
			c.SetBG(x, y, col)
		}
	}
}

func (s *Shore) stars(c *canvas.Canvas, hy int, t float64) {
	far := c.Far()
	glyphs := []rune{'.', '.', '*', '+'}
	for y := 0; y < hy; y++ {
		// Thin out toward the horizon, where haze would eat them.
		density := 0.045 * (1 - float64(y)/math.Max(1, float64(hy))*0.75)
		for x := 0; x < c.W; x++ {
			if HashF(x, y, s.Seed) > density {
				continue
			}
			ph := HashF(x, y, s.Seed+7) * 2 * math.Pi
			twinkle := (0.55 + 0.45*math.Sin(t*0.9+ph)) * s.pal.StarVis
			if twinkle <= 0.02 {
				continue
			}
			g := glyphs[int(HashF(x, y, s.Seed+3)*float64(len(glyphs)))%len(glyphs)]
			far.Plot(x, y, g, s.pal.Star, twinkle)
		}
	}
}

// moon is painted into the BACKGROUND rather than drawn as a block glyph.
// A full-block glyph depends on the font filling its cell exactly; where it
// does not, a disc comes out as horizontal bands. Background colour has no
// such dependency, so this renders the same in every terminal.
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// moon is painted into the BACKGROUND rather than drawn as a block glyph.
// A full-block glyph depends on the font filling its cell exactly; where it
// does not, a disc comes out as horizontal bands. Background colour has no
// such dependency, so this renders the same in every terminal.
//
// lit is the illuminated fraction, and it carries how much context is left:
// a fresh session is a full moon, and the light goes out as the window fills.
// The unlit face is still painted, faintly, so the moon never disappears --
// a missing moon reads as a bug, a dark moon reads as a warning.
func (s *Shore) moon(c *canvas.Canvas, hy int, scale, lit, vis float64) {
	frac := s.MoonX
	if frac <= 0 {
		frac = 0.72
	}
	mx := int(float64(c.W) * frac)
	// Altitude carries context too, alongside phase. Shape alone is hard to
	// judge on a five-cell disc; height above the horizon is easy, because the
	// horizon is a reference line right there. Two cues for one variable is
	// what makes it readable at a glance rather than on inspection.
	my := int(float64(hy) * (0.22 + 0.62*(1-lit)))
	if my < 1 {
		my = 1
	}
	rr := 2.0 * scale
	rim := 0.6 * scale
	// The terminator is approximated by a second disc sliding across the face:
	// fully clear of it at full, concentric at new.
	shadow := 2 * rr * lit
	dark := term.RGB{R: 62, G: 62, B: 74}

	s.moonX, s.moonY = mx, my
	ry := int(rr+rim) + 1
	rx := int((rr+rim)*2) + 1
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			fx := float64(dx) / 2.0
			d := math.Hypot(fx, float64(dy))
			if d > rr+rim {
				continue
			}
			a := 0.92 * vis
			if d > rr-rim {
				a *= (rr + rim - d) / (2 * rim)
			}
			if a <= 0 {
				continue
			}
			col := s.pal.Moon
			if math.Hypot(fx-shadow, float64(dy)) <= rr {
				col, a = dark, a*0.45 // earthshine on the unlit face
			}
			x, y := mx+dx, my+dy
			c.SetBG(x, y, c.BGAt(x, y).Blend(col, a))
		}
	}
}

func (s *Shore) sea(c *canvas.Canvas, hy int, edge []float64, tt float64, act Activity) {
	mid, near := c.Mid(), c.Near()
	ramp := []rune{'·', '-', '~', '≈'}
	cap := '≋'
	if s.ASCII {
		ramp = []rune{'.', '-', '~', '~'}
		cap = '='
	}
	amp := 0.9 + act.Level*1.3

	for x := 0; x < c.W; x++ {
		ex := edge[x]
		rows := math.Max(1, ex-float64(hy))
		for y := hy + 1; float64(y) < ex-0.5; y++ {
			depth := (float64(y) - float64(hy)) / rows // 0 horizon, 1 shore
			// The per-row phase offset matters: without it every row crests at
			// the same x and the depth-ramped threshold carves vertical wedges
			// instead of swell lines.
			ph := float64(x)*0.30 + float64(y)*0.9 + tt*(0.5+depth*1.5)
			wv := (math.Sin(ph) + 0.5*math.Sin(ph*0.47+tt*0.5)) * amp
			// Coverage carries the load, not speed. Speed is invisible in a
			// glance -- and a glance is the only budget this has -- so a busier
			// sea has to mean MORE of the sea in motion, not the same marks
			// moving faster.
			thr := (1.30 - depth*0.50) - act.Level*0.45
			if wv < thr {
				continue
			}
			strength := (wv - thr) / 1.3
			idx := int(strength * float64(len(ramp)))
			if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			g := ramp[idx]
			col := term.Lerp(s.pal.SeaNear, s.pal.Foam,
				math.Min(1, 0.25+depth*0.35+strength*0.45))
			a := 0.5 + depth*0.5
			// Whitecaps: the tallest crests break. This is the cue that makes
			// flat-out unmistakable against merely busy.
			if strength > 0.72 && act.Level > 0.45 {
				g, col = cap, s.pal.Foam
				a = math.Min(1, a+0.35)
			}
			if depth > 0.55 {
				near.Plot(x, y, g, col, a)
			} else {
				mid.Plot(x, y, g, col, a*0.9)
			}
		}
	}
	s.swells(c, hy, edge, tt, act)
	s.glitter(c, hy, edge, tt)
}

// swells are the actual waves: discrete crests that travel from the horizon to
// the shore, steepening as they arrive. Activity drives HOW MANY are in flight
// and HOW TALL they get -- both countable at a glance, unlike wave speed.
//
// A field of noise that thickens reads as static. Lines that move as objects
// read as an ocean, and you can count them.
func (s *Shore) swells(c *canvas.Canvas, hy int, edge []float64, tt float64, act Activity) {
	mid, near := c.Mid(), c.Near()
	n := 2 + int(act.Level*5.4) // 2 at rest, 7 flat out

	body := []rune{'-', '~', '≈'}
	cap := '≋'
	if s.ASCII {
		body, cap = []rune{'-', '~', '~'}, '='
	}

	for i := 0; i < n; i++ {
		// Each swell walks from horizon to shore and wraps; offsetting by
		// i/n keeps them evenly spaced instead of bunching.
		prog := math.Mod(tt*0.13+float64(i)/float64(n), 1)
		ph := float64(i) * 2.4
		// Crests steepen as they come in, and a busier sea sends bigger ones.
		amp := (0.35 + act.Level*1.25) * (0.25 + prog)

		for x := 0; x < c.W; x++ {
			top := float64(hy) + 1
			bot := edge[x] - 0.6
			if bot <= top {
				continue
			}
			y := top + prog*(bot-top) + amp*math.Sin(float64(x)*0.15+ph)
			yy := int(y + 0.5)
			if float64(yy) < top || float64(yy) > bot {
				continue
			}
			g := body[int(prog*float64(len(body)))%len(body)]
			a := 0.35 + prog*0.6
			col := term.Lerp(s.pal.SeaNear, s.pal.Foam, 0.3+prog*0.55)
			// The nearest crests break.
			if prog > 0.78 && act.Level > 0.4 {
				g, col = cap, s.pal.Foam
				a = math.Min(1, a+0.25)
			}
			if prog > 0.5 {
				near.Plot(x, yy, g, col, a)
			} else {
				mid.Plot(x, yy, g, col, a)
			}
		}
	}
}

// glitter is the moon's reflection, widening as it comes toward shore.
func (s *Shore) glitter(c *canvas.Canvas, hy int, edge []float64, tt float64) {
	mid := c.Mid()
	// The moonlight column has to sit UNDER the moon, so take the position the
	// moon actually painted rather than recomputing it. This is why MoonPos
	// exists, and this function ignored it: when the composition mirrored and
	// the moon moved to 0.28, the shine stayed behind at a hardcoded 0.72 --
	// a reflection with no source, on the wrong side of the frame.
	mx := float64(s.moonX)

	// A sliver of moon does not lay a path across the water. Tie the shine to
	// how much of the moon is lit, which is also the context reading, so a
	// nearly-full context dims the water as well as the moon.
	lit := 1 - clamp01(s.ctxUsed)
	strength := (0.35 + 0.65*lit) * s.pal.MoonVis
	if strength <= 0.02 {
		return
	}
	g := '•'
	if s.ASCII {
		g = '*'
	}
	for y := hy + 1; y < c.H; y++ {
		n := 1 + (y-hy)/3
		for k := -n; k <= n; k++ {
			x := int(mx + float64(k)*1.4 + 1.6*math.Sin(tt*0.7+float64(y)*0.6+float64(k)))
			if x < 0 || x >= c.W || float64(y) >= edge[x]-0.5 {
				continue
			}
			a := 0.45 + 0.5*math.Abs(math.Sin(tt*1.3+float64(y)*0.9+float64(k)*1.7))
			a *= 1 - math.Abs(float64(k))/float64(n+1)*0.6
			mid.Plot(x, y, g, s.pal.Glitter, a*strength)
		}
	}
}

func (s *Shore) sand(c *canvas.Canvas, edge []float64) {
	near := c.Near()
	g := '·'
	if s.ASCII {
		g = '.'
	}
	for x := 0; x < c.W; x++ {
		for y := int(edge[x]) + 2; y < c.H; y++ {
			if HashF(x, y, s.Seed+11) > 0.035 {
				continue
			}
			near.Plot(x, y, g, s.pal.Grain, 0.35)
		}
	}
}

// foam is speckle along the waterline, not a filled band. The hash keys on the
// absolute cell, so the pattern does not flicker frame to frame — the churn
// comes from the waterline itself moving through it.
func (s *Shore) foam(c *canvas.Canvas, edge []float64, scale float64) {
	near := c.Near()
	glyphs := []rune{'·', '∘', '°'}
	if s.ASCII {
		glyphs = []rune{'.', 'o', '*'}
	}
	for x := 0; x < c.W; x++ {
		y := int(math.Round(edge[x]))
		for dy := -1; dy <= 1; dy++ {
			yy := y + dy
			h := HashF(x, yy, s.Seed+23)
			cut := 0.25 + 0.30*scale
			if dy != 0 {
				cut = (0.10 + 0.12*scale) // thins out either side of the crest
			}
			if h > cut {
				continue
			}
			g := glyphs[int(h*37)%len(glyphs)]
			near.Plot(x, yy, g, s.pal.Foam, 0.5+0.45*(1-h/cut))
		}
	}
}
