package scape

import (
	"math"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
)

var (
	skyTop     = term.RGB{R: 6, G: 8, B: 22}
	skyHorizon = term.RGB{R: 40, G: 46, B: 78}
	starCol    = term.RGB{R: 206, G: 216, B: 240}
	moonCol    = term.RGB{R: 242, G: 238, B: 214}
	seaFar     = term.RGB{R: 16, G: 24, B: 46}
	seaNear    = term.RGB{R: 30, G: 52, B: 84}
	foamCol    = term.RGB{R: 214, G: 232, B: 238}
	wetSand    = term.RGB{R: 58, G: 51, B: 45}
	sandNear   = term.RGB{R: 76, G: 65, B: 54}
	grainCol   = term.RGB{R: 116, G: 101, B: 84}
	glitterCol = term.RGB{R: 236, G: 232, B: 200}
)

// Shore: night beach. Stars and moon far, open water mid, foam and sand near.
type Shore struct {
	Seed  int64
	ASCII bool

	// Where the moon landed this frame, so callers can anchor a label to it
	// without recomputing the position and drifting out of sync.
	moonX, moonY int
}

// MoonPos is the moon's centre cell from the last Update.
func (s *Shore) MoonPos() (x, y int) { return s.moonX, s.moonY }

func NewShore(seed int64, asciiOnly bool) *Shore { return &Shore{Seed: seed, ASCII: asciiOnly} }

func (s *Shore) Name() string { return "shore" }

func (s *Shore) Update(c *canvas.Canvas, t float64, act Activity) {
	c.Clear()
	if c.W < 8 || c.H < 6 {
		return
	}
	hy := int(float64(c.H) * 0.42) // horizon
	sy := int(float64(c.H) * 0.80) // mean waterline
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

	// Working raises the sea; resting lets it settle.
	tt := t * (0.55 + act.Level*1.45)
	edge := s.waterline(c.W, sy, tt, act, scale)

	s.paintBG(c, hy, edge)
	s.stars(c, hy, t)
	s.moon(c, hy, scale, 1-clamp01(act.ContextUsed))
	s.sea(c, hy, edge, tt, act)
	s.sand(c, edge)
	s.foam(c, edge, scale)
}

// waterline is where the sea meets the sand, per column, as a float. Keeping it
// fractional is the whole trick: the shoreline is painted into the background
// colour at sub-cell precision, so a two-row swing reads as one smooth curve
// instead of the staircase you get from rounding to a row.
func (s *Shore) waterline(w, sy int, tt float64, act Activity, scale float64) []float64 {
	reach := (0.9 + act.Level*1.4) * scale
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
				col = term.Lerp(skyTop, skyHorizon, fy/fh)
			case fy < ex-0.5:
				col = term.Lerp(seaFar, seaNear, (fy-float64(hy))/math.Max(1, ex-float64(hy)))
			case fy < ex+0.5:
				// The waterline cell straddles sea and sand. Mix by how much of
				// the cell the water actually covers.
				col = term.Lerp(wetSand, seaNear, ex-fy+0.5)
			default:
				f := (fy - ex) / math.Max(1, float64(c.H-1)-ex)
				col = term.Lerp(wetSand, sandNear, math.Min(1, f*1.3))
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
			twinkle := 0.55 + 0.45*math.Sin(t*0.9+ph)
			g := glyphs[int(HashF(x, y, s.Seed+3)*float64(len(glyphs)))%len(glyphs)]
			far.Plot(x, y, g, starCol, twinkle)
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
func (s *Shore) moon(c *canvas.Canvas, hy int, scale, lit float64) {
	mx := int(float64(c.W) * 0.72)
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
			a := 0.92
			if d > rr-rim {
				a *= (rr + rim - d) / (2 * rim)
			}
			if a <= 0 {
				continue
			}
			col := moonCol
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
	if s.ASCII {
		ramp = []rune{'.', '-', '~', '~'}
	}
	amp := 0.9 + act.Level*0.9

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
			thr := 1.15 - depth*0.55
			if wv < thr {
				continue
			}
			idx := int((wv - thr) * 2.2)
			if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			col := term.Lerp(seaNear, foamCol, 0.25+depth*0.45)
			if depth > 0.55 {
				near.Plot(x, y, ramp[idx], col, 0.55+depth*0.45)
			} else {
				mid.Plot(x, y, ramp[idx], col, 0.5+depth*0.5)
			}
		}
	}
	s.glitter(c, hy, edge, tt)
}

// glitter is the moon's reflection, widening as it comes toward shore.
func (s *Shore) glitter(c *canvas.Canvas, hy int, edge []float64, tt float64) {
	mid := c.Mid()
	mx := float64(c.W) * 0.72
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
			mid.Plot(x, y, g, glitterCol, a)
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
			near.Plot(x, y, g, grainCol, 0.35)
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
			near.Plot(x, yy, g, foamCol, 0.5+0.45*(1-h/cut))
		}
	}
}

// MoonOnly exposes the moon for the context demo, so the phases can be shown
// enlarged without reproducing the drawing logic somewhere it could drift.
func (s *Shore) MoonOnly(c *canvas.Canvas, hy int, scale, lit float64) {
	s.moon(c, hy, scale, lit)
}
