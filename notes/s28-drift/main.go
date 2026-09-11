// Command s28-drift hunts the SIDEWAYS DRIFT RESIDUE.
//
// His report on the tide build was "it still moves to the sides". The tide fix
// halved the net x-peak notes/drift measures (26% -> 5%) and session 27 left a
// weak pull at about three cells uncharted. This finds what carries it.
//
// THE MEASUREMENT NOTES/DRIFT CANNOT MAKE, and it is the whole point of this
// instrument: a NET cross-correlation peak only sees motion that is the same
// way everywhere. A pattern whose crests slide RIGHT in one part of the frame
// and LEFT in another has a net peak of zero and still, to the eye, moves to
// the sides everywhere you look. So this measures LOCAL drift -- a windowed
// cross-correlation per column block -- alongside the net one.
//
// PIECES:
//
//  1. A REPLICA, PROVEN CELL-EXACT AGAINST THE REAL SHORE over the sea band's
//     own glyphs. Every candidate carrier is a term inside an unexported
//     function, so the only way to vary one at a time without touching product
//     source is to copy the bodies here and prove the copy paints the same
//     frame. `verify` does that and must print 0 disagreeing cells.
//
//  2. NET drift: a Fourier phase velocity along x, per row, per spatial mode,
//     combined as a weighted circular mean of the per-frame phase increment.
//     Cells per second, signed, with a coherence so a number built out of
//     noise can be told from a number built out of a pattern.
//
//  3. LOCAL drift: 24-column windows, stepped by 6, cross-correlated between
//     frames a fixed gap apart, accumulated over many pairs. Reports the mean
//     ABSOLUTE local shift (how much sideways motion there is) and the signed
//     shift as a function of column (which way, where).
//
// ⚠ TRAPS THIS INSTRUMENT IS BUILT AROUND (both have voided instruments here):
// a fresh Shore per frame has no history and Update clamps the gap, so two
// fresh Shores render IDENTICAL frames -- ONE shore, warmed up, advanced in
// frame-sized steps. And `go test` caches, so this is a main package.
//
//	go run ./notes/s28-drift                 (the tide, the default)
//	XSCAPES_TIDE=0 go run ./notes/s28-drift  (the old fixed waterline)
//
// ============================ WHAT IT FOUND, 2026-09-10 =====================
//
//  1. THE "WEAK PULL AT ABOUT THREE CELLS" IS REFUTED. It is a one-sample
//     artifact of notes/drift, which takes exactly ONE frame pair. Replaying
//     that procedure 200 times at 200 different phases: mean arg-max -0.75, and
//     the AVERAGED profile's maximum is at dx=0 (spread 2.8%, lift over zero
//     0.0%). Two other instruments agree there is no current: run tracking's
//     signed mean is +0.05 cells and the Fourier phase velocity's coherence is
//     0.017. THE TIDE DID NOT HALVE THE DRIFT, IT REMOVED IT.
//     Positive control, same instrument at XSCAPES_TIDE=0: mean arg-max -2.04,
//     averaged profile monotone (spread 8.1%, lift 2.3%), run-tracking signed
//     -1.06 cells per 0.24s = -4.4 cells/s LEFTWARD, and the per-column signed
//     profile is negative in 63 of 64 columns. The old drift is unmistakable.
//
//  2. WHAT HE IS SEEING IS REAL AND IS NOT A CURRENT. Marks on the water slide
//     sideways 1.55 cells every 0.24 s -- 6.4 cells/s -- with no net direction.
//     The scene's INWARD motion, measured with the same tracker turned ninety
//     degrees, is 2.15 rows/s. SIDEWAYS IS 3.00 TIMES INWARD. The dominant
//     motion on screen is sideways; the tide only stopped it having a side.
//     (TIDE=0 was 8.2 cells/s sideways at 2.48:1, so the tide bought 22% of the
//     magnitude and all of the direction.)
//
//  3. THE CARRIER IS GEOMETRY, NOT A TERM. A mark's edge moves sideways at the
//     wave's inward speed times the edge's own |dx/dy|; the crests lie nearly
//     flat, measured at 4.78 cells of x per row of y over 7490 edges, and
//     4.78 x 0.498 = 2.38 predicted against 1.53 measured (the tracker
//     under-reads a known shift by ~26%, so this is the same number). Both
//     painters do it: swells() ALONE is 4.08:1 with edges at 5.75 cells/row,
//     sea() ALONE is 2.55:1 at 3.99. NO SINGLE TERM CARRIES IT -- see the
//     refutations below.
//
// REFUTED, so nobody raises them again (all at 120x26, level 0.6, 400 frames):
//
//	(a') the sea's 1.5*sin(x*0.16) warp -- dropping it moves sideways motion
//	     1.547 -> 1.574 cells, i.e. slightly WORSE.
//	(b)  the depth-dependent threshold -- 1.547 -> 1.549. No effect at all.
//	(c)  swells' bot = edge[x] -- 1.547 -> 1.579. Slightly worse.
//	(d)  glitter's tt inside its x -- 1.547 -> 1.540 on the drift glyph set;
//	     read with a mask that can SEE its bullet, glitter alone slides 0.323
//	     cells and freezing its tt takes that to 0.017, but it is 5,870 of
//	     86,540 glyphs and the whole scene only moves 1.405 -> 1.347.
//	(a)  depth inside the sea's TIME term -- 1.547 -> 1.434, 7%. Real but small
//	     ON ITS OWN, because sea() has TWO x-couplings and either one alone is
//	     enough: sea ALONE 1.931, warp dropped 2.038, depth flattened 1.036,
//	     BOTH dropped 0.129. That redundancy is why single ablations found
//	     nothing in session 27.
//	SESSION AGE -- s.phase is an unbounded accumulator that multiplies an
//	     x-dependent term, so the sea's x-structure really does get finer with
//	     session age: mean x-period 30.8 -> 26.9 cells and sideways 1.43 ->
//	     1.79 cells between a fresh scape and one 2.3 hours old. But IT
//	     SATURATES by about eight minutes of wave time and does not move
//	     again out to 2.3 hours. It is not a runaway. (His real sessions:
//	     counted in ~/.config/xscapes/run, 49 sessions with a measurable span,
//	     p50 3.4 h and 23 of 49 over 4 h, so the saturating range is the one
//	     he lives in.)
//
// COSTED, NOT SHIPPED -- his ruling. Steepening the crests is the only knob
// that moves the RATIO; slowing the wave cuts the absolute sideways motion but
// makes the ratio worse, because it slows the inward motion by the same factor.
//
//	swell crest height x4               6.37 -> 4.71 cells/s, 3.07 -> 1.89 : 1
//	both x wavenumbers 0.16/0.15 -> .64 6.37 -> 5.45 cells/s, 3.07 -> 2.66 : 1
//	k .48 + sea warp amp 3.0 + crest x3 6.37 -> 4.47 cells/s, 3.07 -> 1.95 : 1
//	wave speed x0.5                     6.37 -> 3.78 cells/s, 3.07 -> 3.41 : 1
//
// ===========================================================================
package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

var (
	W = 120
	H = 26
)

const (
	seed   = 7
	stepDT = 0.08 // seconds per frame, as notes/drift uses
)

var act = scape.Activity{Level: 0.6, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}

// ---------------------------------------------------------------- glyph sets

// driftSet is exactly what notes/drift counts. Only sea() and swells() emit
// these three, so it is the sea's own moving pattern and nothing else.
func driftSet(r rune) bool { return r == '≈' || r == '~' || r == '≋' }

// withGlitter is the sea's moving marks PLUS the moon's glitter, which draws a
// bullet notes/drift never counted. glitter() is the one place in the file with
// explicit time inside an x position, so it has to be read with a mask that can
// actually see it.
func withGlitter(r rune) bool { return driftSet(r) || r == '•' }

// seaInk is every glyph the sea band's FOREGROUND can carry: the sea ramp, the
// whitecap, the glitter, the foam speckle, the sand grain. Used for the
// replica check, because a background-derived shade block is not a foreground
// glyph and the replica paints no backgrounds.
func seaInk(r rune) bool {
	switch r {
	case '·', '-', '~', '≈', '≋', '•', '∘', '°':
		return true
	}
	return false
}

// ------------------------------------------------------------------- replica

type knobs struct {
	sea, swells, glitter, sand, foam bool

	// (a) the phase SPEED in y is a function of x, because `depth` divides by
	// rows = edge[x]-hy and edge varies with x. Neutralised by dividing by the
	// frame's mean rows in the TIME term only.
	flatDepthTime bool
	// (b) the same through the threshold that decides which cells carry a mark.
	flatDepthThr bool
	// (a') the SPATIAL WARP: under Tide the sea's phase carries
	// 1.5*sin(x*0.16), a static frequency modulation of a wave travelling in y.
	// Static in x -- but a travelling phase read through a fixed FM has crests
	// whose x positions move. Neutralised by dropping the term.
	noSeaXWarp bool
	// (c) swells: bot = edge[x]-0.6, so the crest row's x-shape is the coast's
	// own shape scaled by prog. Neutralised with the mean edge.
	flatSwellBot bool
	// (c2) swells: amp grows with prog over a fixed-phase sine in x.
	frozenSwellAmp bool
	// (c3) swells: the fixed-phase sine in x itself.
	noSwellXWave bool
	// (d) glitter: x = mx + k*1.4 + 1.6*sin(tt*0.7 + ...) -- EXPLICIT time in
	// an x position.
	stillGlitterX bool
	// diagnostic: the tide's own vertical wash, uniform in x.
	noWash bool

	// FIX KNOBS. The sideways speed of a mark's edge is the vertical speed of
	// the wave times the edge's own |dx/dy| -- a shallow crest slides a long
	// way sideways for a small step down the frame. These steepen the crests
	// by raising the phase's gradient in x. Zero means "as shipped".
	seaWarpA, seaWarpK float64 // shipped 1.5 and 0.16
	swellK             float64 // shipped 0.15
	swellAmpMul        float64 // scales the swell's crest height (shipped 1)
	speedMul           float64 // scales the wave's speed in y (shipped 1)
}

func (k knobs) warpA() float64 {
	if k.seaWarpA != 0 {
		return k.seaWarpA
	}
	return 1.5
}
func (k knobs) warpK() float64 {
	if k.seaWarpK != 0 {
		return k.seaWarpK
	}
	return 0.16
}
func (k knobs) swellWaveK() float64 {
	if k.swellK != 0 {
		return k.swellK
	}
	return 0.15
}
func (k knobs) swellAmp() float64 {
	if k.swellAmpMul != 0 {
		return k.swellAmpMul
	}
	return 1
}
func (k knobs) speed() float64 {
	if k.speedMul != 0 {
		return k.speedMul
	}
	return 1
}

func allOn() knobs { return knobs{sea: true, swells: true, glitter: true, sand: true, foam: true} }

type rep struct {
	k                    knobs
	phase, lastT, tideAt float64
	tideRow              int
	pal                  scape.Palette
	ctxUsed              float64
	moonX                int
	writeTop             int
}

func hyRow(h int) int { return int(math.Floor(float64(h) * 0.42)) }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

const moonVisFloor = 0.55

func moonVis(p scape.Palette) float64 {
	if p.MoonVis < moonVisFloor {
		return moonVisFloor
	}
	return p.MoonVis
}

func tideTarget(level, scale float64) float64 {
	if !scape.Tide {
		return 0
	}
	return (1 - clamp01(level)) * scape.TideRange * scale
}

func hyFloor(sy int) float64 { return float64(sy) - scape.TideRange - 1 }

// update is Shore.Update with everything that does not paint a FOREGROUND
// glyph into the sea band removed (backgrounds, stars, moon, todo stars).
// Geometry is recomputed identically; `verify` proves it.
func (r *rep) update(c *canvas.Canvas, t float64) {
	c.Clear()
	r.pal = scape.PaletteAt(act.TimeOfDay)
	r.ctxUsed = act.ContextUsed

	hy := hyRow(c.H)
	wr := scape.DefaultWriteRows
	if m := c.H / 6; wr > m {
		wr = m
	}
	if wr < 2 {
		wr = 0
	}
	writeTop := c.H
	if wr > 0 && c.H > wr+4 {
		writeTop = c.H - wr
	}
	beach := c.H / 5
	if min := wr + 3; wr > 0 && beach < min {
		beach = min
	}
	if beach < 4 {
		beach = 4
	}
	scale := math.Min(float64(c.W)/80.0, float64(c.H)/24.0)
	scale = math.Max(0.45, math.Min(1.35, scale))

	sy := c.H - beach
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > writeTop-1 {
		sy = writeTop - 1
	}

	dt := t - r.lastT
	if dt < 0 || dt > 1 {
		dt = 0.05
	}
	r.lastT = t
	r.tideAt += (tideTarget(act.Level, scale) - r.tideAt) * math.Min(1, dt/scape.TideEase)
	r.tideRow = sy - int(math.Round(r.tideAt))
	r.phase += dt * (0.55 + act.Level*1.45) * r.k.speed()
	tt := r.phase
	edge := r.waterline(c.W, sy, tt, scale)

	if writeTop < c.H {
		ref := float64(sy)
		if scape.Tide {
			ref = float64(r.tideRow)
		}
		room := float64(writeTop-1) - ref
		if room < 0.5 {
			room = 0.5
		}
		dev := 0.0
		for _, e := range edge {
			if d := math.Abs(e - ref); d > dev {
				dev = d
			}
		}
		if dev > room {
			k := room / dev
			for i := range edge {
				edge[i] = ref + (edge[i]-ref)*k
			}
		}
	}
	r.writeTop = writeTop

	if r.k.sea {
		r.seaGlyphs(c, hy, edge, tt)
	}
	if r.k.swells {
		r.swellGlyphs(c, hy, edge, tt)
	}
	if r.k.glitter {
		r.glitterGlyphs(c, hy, edge, tt)
	}
	if r.k.sand {
		r.sandGlyphs(c, edge)
	}
	if r.k.foam {
		r.foamGlyphs(c, edge, scale)
	}
}

func (r *rep) waterline(w, sy int, tt, scale float64) []float64 {
	if scape.Tide {
		wash := (0.7 + act.Level*0.9) * scale * math.Sin(tt*0.5)
		if r.k.noWash {
			wash = 0
		}
		base := float64(sy) - r.tideAt
		if fl := hyFloor(sy); base < fl {
			base = fl
		}
		e := make([]float64, w)
		for x := 0; x < w; x++ {
			fx := float64(x)
			e[x] = base + wash + 0.60*scale*math.Sin(fx*0.11) + 0.35*scale*math.Sin(fx*0.047)
		}
		return e
	}
	reach := (0.8 + act.Level*2.1) * scale
	e := make([]float64, w)
	for x := 0; x < w; x++ {
		fx := float64(x)
		e[x] = float64(sy) + reach*math.Sin(fx*0.11+tt*0.8) + 0.55*scale*math.Sin(fx*0.047-tt*0.45)
	}
	return e
}

func meanOf(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func (r *rep) seaGlyphs(c *canvas.Canvas, hy int, edge []float64, tt float64) {
	mid, near := c.Mid(), c.Near()
	ramp := []rune{'·', '-', '~', '≈'}
	capR := '≋'
	amp := 0.9 + act.Level*1.3

	meanRows := 0.0
	for x := range edge {
		meanRows += math.Max(1, edge[x]-float64(hy))
	}
	meanRows /= float64(len(edge))

	for x := 0; x < c.W; x++ {
		ex := edge[x]
		rows := math.Max(1, ex-float64(hy))
		for y := hy + 1; float64(y) < ex-0.5; y++ {
			depth := (float64(y) - float64(hy)) / rows
			depthT, depthThr := depth, depth
			if r.k.flatDepthTime {
				depthT = (float64(y) - float64(hy)) / meanRows
			}
			if r.k.flatDepthThr {
				depthThr = (float64(y) - float64(hy)) / meanRows
			}
			ph := float64(x)*0.30 + float64(y)*0.9 + tt*(0.5+depth*1.5)
			if scape.Tide {
				warp := r.k.warpA() * math.Sin(float64(x)*r.k.warpK())
				if r.k.noSeaXWarp {
					warp = 0
				}
				ph = (float64(y)+tt*(0.55+depthT*1.65))*0.9 + warp
			}
			wv := (math.Sin(ph) + 0.5*math.Sin(ph*0.47+tt*0.5)) * amp
			thr := (1.30 - depthThr*0.50) - act.Level*0.45
			if wv < thr {
				continue
			}
			strength := (wv - thr) / 1.3
			idx := int(strength * float64(len(ramp)))
			if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			g := ramp[idx]
			col := term.Lerp(r.pal.SeaNear, r.pal.Foam,
				math.Min(1, 0.25+depth*0.35+strength*0.45))
			a := 0.5 + depth*0.5
			if strength > 0.72 && act.Level > 0.45 {
				g, col = capR, r.pal.Foam
				a = math.Min(1, a+0.35)
			}
			if depth > 0.55 {
				near.Plot(x, y, g, col, a)
			} else {
				mid.Plot(x, y, g, col, a*0.9)
			}
		}
	}
}

func (r *rep) swellGlyphs(c *canvas.Canvas, hy int, edge []float64, tt float64) {
	mid, near := c.Mid(), c.Near()
	n := 2 + int(act.Level*5.4)
	body := []rune{'-', '~', '≈'}
	capR := '≋'
	meanEdge := meanOf(edge)

	for i := 0; i < n; i++ {
		prog := math.Mod(tt*0.13+float64(i)/float64(n), 1)
		ph := float64(i) * 2.4
		amp := (0.35 + act.Level*1.25) * (0.25 + prog) * r.k.swellAmp()
		if r.k.frozenSwellAmp {
			amp = (0.35 + act.Level*1.25) * (0.25 + 0.5) * r.k.swellAmp()
		}
		for x := 0; x < c.W; x++ {
			top := float64(hy) + 1
			bot := edge[x] - 0.6
			if r.k.flatSwellBot {
				bot = meanEdge - 0.6
			}
			if bot <= top {
				continue
			}
			wave := math.Sin(float64(x)*r.k.swellWaveK() + ph)
			if r.k.noSwellXWave {
				wave = 0
			}
			y := top + prog*(bot-top) + amp*wave
			yy := int(y + 0.5)
			if float64(yy) < top || float64(yy) > bot {
				continue
			}
			g := body[int(prog*float64(len(body)))%len(body)]
			a := 0.35 + prog*0.6
			col := term.Lerp(r.pal.SeaNear, r.pal.Foam, 0.3+prog*0.55)
			if prog > 0.78 && act.Level > 0.4 {
				g, col = capR, r.pal.Foam
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

func (r *rep) glitterGlyphs(c *canvas.Canvas, hy int, edge []float64, tt float64) {
	mid := c.Mid()
	mx := float64(r.moonX)
	lit := 1 - clamp01(r.ctxUsed)
	strength := (0.35 + 0.65*lit) * moonVis(r.pal)
	if strength <= 0.02 {
		return
	}
	g := '•'
	for y := hy + 1; y < c.H; y++ {
		n := 1 + (y-hy)/3
		for k := -n; k <= n; k++ {
			wob := 1.6 * math.Sin(tt*0.7+float64(y)*0.6+float64(k))
			if r.k.stillGlitterX {
				wob = 1.6 * math.Sin(float64(y)*0.6+float64(k))
			}
			x := int(mx + float64(k)*1.4 + wob)
			if x < 0 || x >= c.W || float64(y) >= edge[x]-0.5 {
				continue
			}
			a := 0.45 + 0.5*math.Abs(math.Sin(tt*1.3+float64(y)*0.9+float64(k)*1.7))
			a *= 1 - math.Abs(float64(k))/float64(n+1)*0.6
			mid.Plot(x, y, g, r.pal.Glitter, a*strength)
		}
	}
}

func (r *rep) sandGlyphs(c *canvas.Canvas, edge []float64) {
	near := c.Near()
	g := '·'
	bottom := c.H
	if r.writeTop > 0 && r.writeTop < bottom {
		bottom = r.writeTop
	}
	for x := 0; x < c.W; x++ {
		for y := int(edge[x]) + 2; y < bottom; y++ {
			if scape.HashF(x, y, seed+11) > 0.035 {
				continue
			}
			near.Plot(x, y, g, r.pal.Grain, 0.35)
		}
	}
}

func (r *rep) foamGlyphs(c *canvas.Canvas, edge []float64, scale float64) {
	near := c.Near()
	glyphs := []rune{'·', '∘', '°'}
	for x := 0; x < c.W; x++ {
		y := int(math.Round(edge[x]))
		for dy := -1; dy <= 1; dy++ {
			yy := y + dy
			if r.writeTop > 0 && yy >= r.writeTop {
				continue
			}
			h := scape.HashF(x, yy, seed+23)
			cut := 0.25 + 0.30*scale
			if dy != 0 {
				cut = (0.10 + 0.12*scale)
			}
			if h > cut {
				continue
			}
			g := glyphs[int(h*37)%len(glyphs)]
			near.Plot(x, yy, g, r.pal.Foam, 0.5+0.45*(1-h/cut))
		}
	}
}

// ----------------------------------------------------------------- sampling

type mask [][]bool

func maskOf(c *canvas.Canvas, keep func(rune) bool) mask {
	m := make(mask, H)
	for y := range m {
		m[y] = make([]bool, W)
		for x := 0; x < W; x++ {
			r, _, _ := c.ResolveAt(x, y, term.Profile256)
			m[y][x] = keep(r)
		}
	}
	return m
}

// frames renders a run and returns the masks, from ONE warmed-up shore.
type source func(n int) []mask

// warmFrames is how much history the shore has before anything is sampled. It
// is a PARAMETER because s.phase is an unbounded accumulator: it never wraps,
// so the scape at minute 60 is not the scape at second 30, and an instrument
// with a fixed warm-up cannot see that.
var warmFrames = 400

func replicaSource(k knobs, moonX int, keep func(rune) bool) source {
	r := &rep{k: k, moonX: moonX}
	t := 0.0
	adv := func() *canvas.Canvas {
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		r.update(c, t)
		return c
	}
	for i := 0; i < warmFrames; i++ { // warm up: real history, not a fresh shore
		adv()
	}
	return func(n int) []mask {
		out := make([]mask, n)
		for i := range out {
			out[i] = maskOf(adv(), keep)
		}
		return out
	}
}

func realSource(keep func(rune) bool) source {
	sh := scape.NewShore(seed, false)
	t := 0.0
	adv := func() *canvas.Canvas {
		c := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		sh.Update(c, t, act)
		return c
	}
	for i := 0; i < warmFrames; i++ {
		adv()
	}
	return func(n int) []mask {
		out := make([]mask, n)
		for i := range out {
			out[i] = maskOf(adv(), keep)
		}
		return out
	}
}

// xPeriod reports the pattern's spatial scale ALONG X: the power-weighted mean
// period in cells, and the period of the strongest single mode. If the sea's
// crests are being chopped finer and finer as the session ages, this is where
// it shows.
func xPeriod(ms []mask, y0, y1 int) (meanPeriod, peakPeriod float64) {
	half := W / 2
	pow := make([]float64, half+1)
	for _, m := range ms {
		for y := y0; y < y1; y++ {
			for k := 1; k <= half; k++ {
				kk := 2 * math.Pi * float64(k) / float64(W)
				var re, im float64
				for x := 0; x < W; x++ {
					if m[y][x] {
						re += math.Cos(kk * float64(x))
						im -= math.Sin(kk * float64(x))
					}
				}
				pow[k] += re*re + im*im
			}
		}
	}
	var num, den, best float64
	for k := 1; k <= half; k++ {
		num += float64(W) / float64(k) * pow[k]
		den += pow[k]
		if pow[k] > best {
			best, peakPeriod = pow[k], float64(W)/float64(k)
		}
	}
	if den > 0 {
		meanPeriod = num / den
	}
	return
}

// ------------------------------------------------------------- NET drift

// netDrift reports the pattern's phase velocity along x in cells per second,
// signed (+ = to the right), together with a coherence in [0,1]: how much of
// the phase increment points the same way frame to frame. A number with a
// coherence near zero is noise wearing a decimal point.
//
// For a pattern f(x - v t) the Fourier coefficient at wavenumber k turns as
// exp(-i k v t), so arg(C_f * conj(C_{f+1})) = k v dt. Summing that product as
// a COMPLEX number weights each frame by its own amplitude and never needs the
// phase unwrapped.
func netDrift(ms []mask, y0, y1, x0, x1 int) (v, coh, cells float64) {
	width := x1 - x0
	var z complex128
	var norm float64
	for m := 2; m <= 6; m++ { // spatial modes across the scored width
		k := 2 * math.Pi * float64(m) / float64(width)
		for y := y0; y < y1; y++ {
			var prev complex128
			have := false
			for f := range ms {
				var cf complex128
				for x := x0; x < x1; x++ {
					if ms[f][y][x] {
						cf += cmplx.Exp(complex(0, -k*float64(x)))
					}
				}
				if have {
					p := prev * cmplx.Conj(cf)
					// arg(p) = k*v*dt for a right-moving pattern; scale each
					// mode's contribution so they can be summed as velocities.
					mag := cmplx.Abs(p)
					if mag > 0 {
						ang := cmplx.Phase(p) / (k * stepDT) // cells/sec
						z += complex(mag*math.Cos(ang), mag*math.Sin(ang))
						norm += mag
					}
				}
				prev, have = cf, true
			}
		}
	}
	for f := range ms {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if ms[f][y][x] {
					cells++
				}
			}
		}
	}
	if norm == 0 {
		return 0, 0, cells
	}
	// z's angle is the amplitude-weighted circular mean of the per-frame
	// velocity estimates, wrapped into +-pi cells/sec; its magnitude over norm
	// is the coherence.
	return cmplx.Phase(z), cmplx.Abs(z) / norm, cells
}

// ----------------------------------------------------------- LOCAL drift

const maxLocal = 6

type local struct {
	centre int
	prof   []float64 // matches per shift, index dx+maxLocal
	acells float64
}

func (l *local) centroid() float64 {
	minv := math.Inf(1)
	for _, v := range l.prof {
		if v < minv {
			minv = v
		}
	}
	num, den := 0.0, 0.0
	for i, v := range l.prof {
		w := v - minv
		num += float64(i-maxLocal) * w
		den += w
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// localDrift cross-correlates narrow column windows between frames `gap`
// apart, accumulated over every pair in the run. This is what a NET
// correlation cannot see: motion that is rightward in one part of the frame
// and leftward in another.
func localDrift(ms []mask, y0, y1, gap, winW, step int) []*local {
	var out []*local
	for cx := maxLocal; cx+winW+maxLocal <= W; cx += step {
		l := &local{centre: cx + winW/2, prof: make([]float64, 2*maxLocal+1)}
		for f := 0; f+gap < len(ms); f++ {
			a, b := ms[f], ms[f+gap]
			for dx := -maxLocal; dx <= maxLocal; dx++ {
				n := 0
				for y := y0; y < y1; y++ {
					for x := cx; x < cx+winW; x++ {
						if a[y][x] && b[y][x+dx] {
							n++
						}
					}
				}
				l.prof[dx+maxLocal] += float64(n)
			}
			for y := y0; y < y1; y++ {
				for x := cx; x < cx+winW; x++ {
					if a[y][x] {
						l.acells++
					}
				}
			}
		}
		out = append(out, l)
	}
	return out
}

func localSummary(ls []*local) (meanAbs, meanSigned, maxAbs float64) {
	if len(ls) == 0 {
		return
	}
	for _, l := range ls {
		c := l.centroid()
		meanAbs += math.Abs(c)
		meanSigned += c
		if math.Abs(c) > maxAbs {
			maxAbs = math.Abs(c)
		}
	}
	return meanAbs / float64(len(ls)), meanSigned / float64(len(ls)), maxAbs
}

// ------------------------------------------------------- RUN TRACKING

// A run is one contiguous group of set cells in a row -- one visible mark on
// the water. Following a mark from frame to frame is the closest thing the
// instrument has to what the eye does, and unlike a correlation it reports a
// displacement in cells with no pedestal to subtract.
type run struct{ x0, x1 int }

func (r run) centre() float64 { return (float64(r.x0) + float64(r.x1)) / 2 }
func (r run) length() int     { return r.x1 - r.x0 + 1 }

func runsIn(row []bool, x0, x1 int) []run {
	var out []run
	i := x0
	for i < x1 {
		if !row[i] {
			i++
			continue
		}
		j := i
		for j+1 < x1 && row[j+1] {
			j++
		}
		out = append(out, run{i, j})
		i = j + 1
	}
	return out
}

type flow struct {
	n           float64
	sumSigned   float64
	sumAbs      float64
	byX         [64]float64 // signed displacement summed into 64 column bins
	byXn        [64]float64
	runs        float64
	runLenTotal float64

	// A mark's LEFT edge and its RIGHT edge move separately. Both the same way
	// is a mark sliding sideways; opposite ways is a mark growing or shrinking
	// in place, which a centre-of-run reading cannot tell apart from a slide.
	// sumTrans is |(dL+dR)/2|, sumBreath is |(dR-dL)/2|, and ss/sl/sr feed a
	// correlation between dL and dR: positive means slide, negative breath.
	sumTrans, sumBreath   float64
	sl, sr, sll, srr, slr float64
}

func (f *flow) meanSigned() float64 {
	if f.n == 0 {
		return 0
	}
	return f.sumSigned / f.n
}
func (f *flow) meanAbs() float64 {
	if f.n == 0 {
		return 0
	}
	return f.sumAbs / f.n
}
func (f *flow) meanTrans() float64 {
	if f.n == 0 {
		return 0
	}
	return f.sumTrans / f.n
}
func (f *flow) meanBreath() float64 {
	if f.n == 0 {
		return 0
	}
	return f.sumBreath / f.n
}

// edgeCorr is the correlation between a mark's left-edge displacement and its
// right-edge displacement. +1 is a mark sliding rigidly sideways; -1 is a mark
// growing or shrinking about a fixed centre.
func (f *flow) edgeCorr() float64 {
	n := f.n
	if n < 2 {
		return 0
	}
	cov := f.slr/n - (f.sl/n)*(f.sr/n)
	vl := f.sll/n - (f.sl/n)*(f.sl/n)
	vr := f.srr/n - (f.sr/n)*(f.sr/n)
	if vl <= 0 || vr <= 0 {
		return 0
	}
	return cov / math.Sqrt(vl*vr)
}

func (f *flow) meanRunLen() float64 {
	if f.runs == 0 {
		return 0
	}
	return f.runLenTotal / f.runs
}

// track matches every run in frame f to the nearest run in frame f+gap whose
// length is within a factor of two, and records the signed displacement of
// their centres. Runs with no partner within maxShift are dropped and counted.
func track(ms []mask, y0, y1, gap, maxShift int) (*flow, float64) {
	fl := &flow{}
	var offered, matched float64
	for f := 0; f+gap < len(ms); f++ {
		a, b := ms[f], ms[f+gap]
		for y := y0; y < y1; y++ {
			ra := runsIn(a[y], 0, W)
			rb := runsIn(b[y], 0, W)
			for _, r := range ra {
				fl.runs++
				fl.runLenTotal += float64(r.length())
			}
			if len(ra) == 0 || len(rb) == 0 {
				continue
			}
			for _, r := range ra {
				offered++
				best, bd := -1, math.Inf(1)
				for j, q := range rb {
					d := math.Abs(q.centre() - r.centre())
					if d > float64(maxShift) {
						continue
					}
					if q.length() > 2*r.length() || r.length() > 2*q.length() {
						continue
					}
					if d < bd {
						bd, best = d, j
					}
				}
				if best < 0 {
					continue
				}
				matched++
				d := rb[best].centre() - r.centre()
				dL := float64(rb[best].x0 - r.x0)
				dR := float64(rb[best].x1 - r.x1)
				fl.sumTrans += math.Abs((dL + dR) / 2)
				fl.sumBreath += math.Abs((dR - dL) / 2)
				fl.sl += dL
				fl.sr += dR
				fl.sll += dL * dL
				fl.srr += dR * dR
				fl.slr += dL * dR
				fl.n++
				fl.sumSigned += d
				fl.sumAbs += math.Abs(d)
				bin := int(r.centre() * 64 / float64(W))
				if bin >= 0 && bin < 64 {
					fl.byX[bin] += d
					fl.byXn[bin]++
				}
			}
		}
	}
	rate := 0.0
	if offered > 0 {
		rate = matched / offered
	}
	return fl, rate
}

// trackCols is `track` turned ninety degrees: it follows each mark DOWN a
// column instead of along a row, so the sideways speed can be put beside the
// inward speed the scene is actually trying to show. Direction is not read off
// it -- notes/drift records that the y axis is confounded under Tide, because
// the wash moves the whole band -- but the MAGNITUDE is what is wanted here.
func trackCols(ms []mask, y0, y1, gap, maxShift int) (*flow, float64) {
	fl := &flow{}
	var offered, matched float64
	col := make([]bool, H)
	col2 := make([]bool, H)
	for f := 0; f+gap < len(ms); f++ {
		a, b := ms[f], ms[f+gap]
		for x := 0; x < W; x++ {
			for y := 0; y < H; y++ {
				col[y], col2[y] = false, false
			}
			for y := y0; y < y1; y++ {
				col[y] = a[y][x]
				col2[y] = b[y][x]
			}
			ra := runsIn(col, y0, y1)
			rb := runsIn(col2, y0, y1)
			for _, r := range ra {
				fl.runs++
				fl.runLenTotal += float64(r.length())
			}
			if len(ra) == 0 || len(rb) == 0 {
				continue
			}
			for _, r := range ra {
				offered++
				best, bd := -1, math.Inf(1)
				for j, q := range rb {
					d := math.Abs(q.centre() - r.centre())
					if d > float64(maxShift) {
						continue
					}
					if q.length() > 2*r.length() || r.length() > 2*q.length() {
						continue
					}
					if d < bd {
						bd, best = d, j
					}
				}
				if best < 0 {
					continue
				}
				matched++
				d := rb[best].centre() - r.centre()
				fl.n++
				fl.sumSigned += d
				fl.sumAbs += math.Abs(d)
			}
		}
	}
	rate := 0.0
	if offered > 0 {
		rate = matched / offered
	}
	return fl, rate
}

// edgeSlope measures the marks' own geometry: how far a mark's left edge moves
// in X for one step in Y, within a single frame. That number IS the multiplier
// between the wave's inward speed and its sideways speed -- a crest lying
// nearly flat across the frame has its edge race sideways for a small step
// down. Reported as cells of x per row of y.
func edgeSlope(ms []mask, y0, y1 int) (mean float64, n float64) {
	for _, m := range ms {
		for y := y0; y+1 < y1; y++ {
			a := runsIn(m[y], 0, W)
			b := runsIn(m[y+1], 0, W)
			if len(a) == 0 || len(b) == 0 {
				continue
			}
			for _, r := range a {
				best, bd := -1.0, math.Inf(1)
				for _, q := range b {
					d := math.Abs(float64(q.x0 - r.x0))
					if d < bd {
						bd, best = d, float64(q.x0-r.x0)
					}
				}
				if bd > 12 {
					continue
				}
				mean += math.Abs(best)
				n++
			}
		}
	}
	if n > 0 {
		mean /= n
	}
	return
}

// shiftMask moves a mask sideways by dx, for calibrating the metrics against a
// displacement that is known exactly.
func shiftMask(m mask, dx int) mask {
	o := make(mask, H)
	for y := range m {
		o[y] = make([]bool, W)
		for x := 0; x < W; x++ {
			sx := x - dx
			if sx >= 0 && sx < W {
				o[y][x] = m[y][sx]
			}
		}
	}
	return o
}

// bestPeriod fits cos(2*pi*x/P + phi) to a signed per-column profile and
// returns the period that fits best, in columns, with its explained variance.
// A period names the term that carries the motion: 2*pi/0.16 = 39.3 cells is
// the sea's x-warp (half-period 19.6 for a sign alternation), 2*pi/0.15 = 41.9
// the swell sine, 2*pi/0.11 = 57.1 the coast's long wave.
func bestPeriod(xs, ys []float64) (period, r2 float64) {
	if len(xs) < 6 {
		return 0, 0
	}
	mean := 0.0
	for _, v := range ys {
		mean += v
	}
	mean /= float64(len(ys))
	d := make([]float64, len(ys))
	var sst float64
	for i := range ys {
		d[i] = ys[i] - mean
		sst += d[i] * d[i]
	}
	if sst == 0 {
		return 0, 0
	}
	for p := 8.0; p <= 160.0; p += 0.25 {
		k := 2 * math.Pi / p
		var sc, ss, cc, s2 float64
		for i := range xs {
			c, s := math.Cos(k*xs[i]), math.Sin(k*xs[i])
			sc += d[i] * c
			ss += d[i] * s
			cc += c * c
			s2 += s * s
		}
		if cc == 0 || s2 == 0 {
			continue
		}
		if rr := (sc*sc/cc + ss*ss/s2) / sst; rr > r2 {
			r2, period = rr, p
		}
	}
	return
}

// ------------------------------------------------------------------ verify

// verify renders the real Shore and the replica in lockstep and counts cells
// in the sea band where their FOREGROUND glyphs disagree. The replica paints no
// backgrounds, and on 256 a background can resolve to a shade or half block, so
// only the sea's own foreground glyphs are comparable -- which is exactly the
// set every measurement below reads.
func verify(moonX int) (bad, ink, frames int) {
	sh := scape.NewShore(seed, false)
	r := &rep{k: allOn(), moonX: moonX}
	t := 0.0
	hy := hyRow(H)
	for i := 0; i < 460; i++ {
		ca := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cb := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		sh.Update(ca, t, act)
		r.update(cb, t)
		if i < 400 {
			continue
		}
		frames++
		for y := hy + 1; y < H; y++ {
			for x := 0; x < W; x++ {
				g1, _, _ := ca.ResolveAt(x, y, term.Profile256)
				g2, _, _ := cb.ResolveAt(x, y, term.Profile256)
				a, b := seaInk(g1), seaInk(g2)
				if a {
					ink++
				}
				if a != b || (a && b && g1 != g2) {
					bad++
				}
			}
		}
	}
	return
}

// ------------------------------------------------ notes/drift, replayed

// driftOnce is notes/drift's own procedure, verbatim in structure: one shore,
// 40 frames to the a-frame, 18 more to the b-frame, cross-correlate over
// dx -4..+4 and take the arg-max. `phase` starts it at a different point in
// the wave so the same measurement can be repeated at many phases.
func driftReplay(trials int) (peaks []int, meanProfile []float64) {
	meanProfile = make([]float64, 9)
	sh := scape.NewShore(seed, false)
	t := 0.0
	adv := func(n int) *canvas.Canvas {
		var c *canvas.Canvas
		for i := 0; i < n; i++ {
			c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			t += stepDT
			sh.Update(c, t, act)
		}
		return c
	}
	grab := func(c *canvas.Canvas) [][]bool {
		g := make([][]bool, H)
		for y := range g {
			g[y] = make([]bool, W)
			for x := 0; x < W; x++ {
				r, _, _ := c.ResolveAt(x, y, term.Profile256)
				g[y][x] = driftSet(r)
			}
		}
		return g
	}
	for tr := 0; tr < trials; tr++ {
		a := grab(adv(40))
		b := grab(adv(18))
		score := func(dx int) int {
			n := 0
			for y := 2; y < H-2; y++ {
				for x := 2; x < W-2; x++ {
					xx := x + dx
					if xx < 0 || xx >= W {
						continue
					}
					if a[y][x] && b[y][xx] {
						n++
					}
				}
			}
			return n
		}
		best, bs := 0, -1
		for dx := -4; dx <= 4; dx++ {
			s := score(dx)
			meanProfile[dx+4] += float64(s)
			if s > bs {
				bs, best = s, dx
			}
		}
		peaks = append(peaks, best)
	}
	for i := range meanProfile {
		meanProfile[i] /= float64(trials)
	}
	return
}

// ------------------------------------------------------------------- report

type scen struct {
	name string
	k    knobs
}

type reading struct {
	glyphs   float64
	net      float64
	coh      float64
	flowAbs  float64 // cells per gap
	flowSgn  float64
	matched  float64
	runLen   float64
	locAbs   float64
	locSgn   float64
	locWorst float64
	trans    float64
	breath   float64
	ecorr    float64
	period   float64
	r2       float64
	byX      []float64
	byXc     []float64
}

func measure(src source, nframes, y0, y1, trackGap, corrGap int) reading {
	ms := src(nframes)
	var rd reading
	rd.net, rd.coh, rd.glyphs = netDrift(ms, y0, y1, maxLocal, W-maxLocal)
	fl, rate := track(ms, y0, y1, trackGap, 6)
	rd.flowAbs, rd.flowSgn, rd.matched, rd.runLen = fl.meanAbs(), fl.meanSigned(), rate, fl.meanRunLen()
	rd.trans, rd.breath, rd.ecorr = fl.meanTrans(), fl.meanBreath(), fl.edgeCorr()
	ls := localDrift(ms, y0, y1, corrGap, 12, 3)
	rd.locAbs, rd.locSgn, rd.locWorst = localSummary(ls)
	for i := 0; i < 64; i++ {
		if fl.byXn[i] > 0 {
			rd.byX = append(rd.byX, fl.byX[i]/fl.byXn[i])
			rd.byXc = append(rd.byXc, (float64(i)+0.5)*float64(W)/64)
		}
	}
	rd.period, rd.r2 = bestPeriod(rd.byXc, rd.byX)
	return rd
}

func main() {
	out := os.Stdout
	fmt.Fprintf(out, "scape.Tide=%v   term.NoSplitCells=%v   term.Shading=%v   canvas %dx%d   level %.2f   step %.2fs\n",
		scape.Tide, term.NoSplitCells, term.Shading, W, H, act.Level, stepDT)

	probe := scape.NewShore(seed, false)
	pc := canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	probe.Update(pc, stepDT, act)
	moonX, _ := probe.MoonPos()
	hy := hyRow(H)
	fmt.Fprintf(out, "hy=%d  moonX=%d\n\n", hy, moonX)

	y0, y1 := hy+1, H-6
	const nframes, trackGap, corrGap = 400, 3, 12

	fmt.Fprintln(out, "== 1. THE REPLICA IS CELL-EXACT AGAINST THE REAL SHORE ==")
	bad, ink, frames := verify(moonX)
	fmt.Fprintf(out, "   %d frames, %d sea-band foreground glyphs in the real frames, %d disagreeing cells\n",
		frames, ink, bad)
	if bad != 0 {
		fmt.Fprintln(out, "   REPLICA IS NOT FAITHFUL -- every number below is VOID.")
	} else {
		fmt.Fprintln(out, "   OK -- the replica may be varied.")
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 2. NOTES/DRIFT REPLAYED 200 TIMES AT DIFFERENT PHASES ==")
	fmt.Fprintln(out, "   notes/drift takes exactly ONE frame pair. This runs its own")
	fmt.Fprintln(out, "   procedure 200 times and reports where the arg-max actually lands.")
	{
		peaks, mp := driftReplay(200)
		hist := map[int]int{}
		sum := 0.0
		for _, p := range peaks {
			hist[p]++
			sum += float64(p)
		}
		fmt.Fprint(out, "   peak dx histogram: ")
		for dx := -4; dx <= 4; dx++ {
			fmt.Fprintf(out, "%+d:%d ", dx, hist[dx])
		}
		fmt.Fprintf(out, "\n   mean peak dx %+.2f over %d trials\n", sum/float64(len(peaks)), len(peaks))
		fmt.Fprint(out, "   mean profile:      ")
		mx, mn := 0.0, math.Inf(1)
		for dx := -4; dx <= 4; dx++ {
			v := mp[dx+4]
			fmt.Fprintf(out, "%+d:%.1f ", dx, v)
			if v > mx {
				mx = v
			}
			if v < mn {
				mn = v
			}
		}
		fmt.Fprintf(out, "\n   averaged spread (max-min)/max %.1f%%, lift over dx=0 %.1f%%\n",
			(mx-mn)/mx*100, (mx-mp[4])/mp[4]*100)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 3. CALIBRATION: both metrics against a KNOWN shift ==")
	{
		src := realSource(driftSet)
		ms := src(60)
		for _, dx := range []int{-3, -1, 0, 1, 3} {
			pair := []mask{ms[0], shiftMask(ms[0], dx)}
			fl, rate := track(pair, y0, y1, 1, 6)
			ls := localDrift(pair, y0, y1, 1, 12, 3)
			_, ns, _ := localSummary(ls)
			fmt.Fprintf(out, "   true %+d  ->  run-tracking %+.2f (match rate %.2f)   windowed corr %+.2f\n",
				dx, fl.meanSigned(), rate, ns)
		}
	}
	fmt.Fprintln(out)

	fmt.Fprintf(out, "== 4. THE REAL SHORE AS SHIPPED (%d frames = %.1fs, rows %d..%d) ==\n",
		nframes, float64(nframes)*stepDT, y0, y1-1)
	rd := measure(realSource(driftSet), nframes, y0, y1, trackGap, corrGap)
	fmt.Fprintf(out, "   sample: %.0f sea glyphs over %d frames; %.0f%% of marks matched a partner; mean mark %.1f cells long\n",
		rd.glyphs, nframes, rd.matched*100, rd.runLen)
	fmt.Fprintf(out, "   NET Fourier drift  %+.3f cells/s (coherence %.3f)\n", rd.net, rd.coh)
	fmt.Fprintf(out, "   RUN TRACKING       mean|dx| %.2f cells / %.2fs = %.2f cells/s sideways;  signed %+.2f cells\n",
		rd.flowAbs, float64(trackGap)*stepDT, rd.flowAbs/(float64(trackGap)*stepDT), rd.flowSgn)
	fmt.Fprintf(out, "     of which SLIDE  %.2f cells and BREATH (mark grows/shrinks) %.2f;  left-edge/right-edge correlation %+.2f\n",
		rd.trans, rd.breath, rd.ecorr)
	fmt.Fprintf(out, "   WINDOWED CORR      mean|dx| %.2f / %.2fs;  signed %+.2f;  worst window %.2f\n",
		rd.locAbs, float64(corrGap)*stepDT, rd.locSgn, rd.locWorst)
	{
		// The comparison that says whether sideways is a residue or the main
		// event: the same tracker turned ninety degrees.
		ms2 := realSource(driftSet)(nframes)
		fy, ry := trackCols(ms2, y0, y1, trackGap, 6)
		fx, _ := track(ms2, y0, y1, trackGap, 6)
		fmt.Fprintf(out, "   INWARD (vertical)  mean|dy| %.2f rows / %.2fs = %.2f rows/s (match %.2f)\n",
			fy.meanAbs(), float64(trackGap)*stepDT, fy.meanAbs()/(float64(trackGap)*stepDT), ry)
		fmt.Fprintf(out, "   SIDEWAYS : INWARD  = %.2f : 1\n", fx.meanAbs()/math.Max(1e-9, fy.meanAbs()))
	}
	fmt.Fprintf(out, "   signed-by-column alternation period %.1f cells (explained variance %.2f)\n", rd.period, rd.r2)
	fmt.Fprint(out, "   signed displacement by column: ")
	for i := range rd.byX {
		fmt.Fprintf(out, "%.0f:%+.2f ", rd.byXc[i], rd.byX[i])
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 5. ONE CARRIER AT A TIME (replica; everything else untouched) ==")
	base := allOn()
	mut := func(f func(*knobs)) knobs { k := base; f(&k); return k }
	scens := []scen{
		{"baseline (nothing neutralised)", base},
		{"(a') sea: drop 1.5*sin(x*0.16)", mut(func(k *knobs) { k.noSeaXWarp = true })},
		{"(a) flat depth in the TIME term", mut(func(k *knobs) { k.flatDepthTime = true })},
		{"(b) flat depth in the THRESHOLD", mut(func(k *knobs) { k.flatDepthThr = true })},
		{"(c) swells: flat bot (mean edge)", mut(func(k *knobs) { k.flatSwellBot = true })},
		{"(c2) swells: frozen amplitude", mut(func(k *knobs) { k.frozenSwellAmp = true })},
		{"(c3) swells: drop sin(x*0.15)", mut(func(k *knobs) { k.noSwellXWave = true })},
		{"(d) glitter: no time in its x", mut(func(k *knobs) { k.stillGlitterX = true })},
		{"diagnostic: wash frozen", mut(func(k *knobs) { k.noWash = true })},
		{"sea() ALONE", knobs{sea: true}},
		{"sea() ALONE, x-warp dropped", knobs{sea: true, noSeaXWarp: true}},
		{"swells() ALONE", knobs{swells: true}},
		{"swells() ALONE, sin(x*0.15) dropped", knobs{swells: true, noSwellXWave: true}},
		{"swells() ALONE, frozen amplitude", knobs{swells: true, frozenSwellAmp: true}},
		{"swells() ALONE, flat bot", knobs{swells: true, flatSwellBot: true}},
		{"(a')+(c3) both x terms dropped", mut(func(k *knobs) { k.noSeaXWarp, k.noSwellXWave = true, true })},
		{"sea ALONE, x fully decoupled", knobs{sea: true, noSeaXWarp: true, flatDepthTime: true, flatDepthThr: true}},
		{"sea ALONE, warp kept, depth flat", knobs{sea: true, flatDepthTime: true, flatDepthThr: true}},
		{"sea ALONE, wash frozen", knobs{sea: true, noWash: true}},
		{"swells ALONE, wash frozen", knobs{swells: true, noWash: true}},
	}
	fmt.Fprintf(out, "   %-38s %-8s %-7s %-8s %-8s %-8s %-8s %-8s %s\n",
		"variant", "glyphs", "match%", "runlen", "run|dx|", "run sgn", "slide", "breath", "edgeR")
	for _, s := range scens {
		r := measure(replicaSource(s.k, moonX, driftSet), nframes, y0, y1, trackGap, corrGap)
		if r.glyphs == 0 {
			fmt.Fprintf(out, "   %-38s EMPTY SAMPLE -- result VOID\n", s.name)
			continue
		}
		fmt.Fprintf(out, "   %-38s %-8.0f %-7.2f %-8.1f %-8.3f %-+8.3f %-8.3f %-8.3f %+.2f\n",
			s.name, r.glyphs, r.matched, r.runLen, r.flowAbs, r.flowSgn, r.trans, r.breath, r.ecorr)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 6. NULL CONTROL: a frame repeated (zero motion by construction) ==")
	{
		src := replicaSource(allOn(), moonX, driftSet)
		ms := src(40)
		still := make([]mask, len(ms))
		for i := range still {
			still[i] = ms[0]
		}
		fl, rate := track(still, y0, y1, trackGap, 6)
		ls := localDrift(still, y0, y1, corrGap, 12, 3)
		a, sg, w := localSummary(ls)
		fmt.Fprintf(out, "   run tracking |dx| %.3f signed %+.3f (match %.2f);  windowed corr |dx| %.2f signed %+.2f worst %.2f\n",
			fl.meanAbs(), fl.meanSigned(), rate, a, sg, w)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 7. HIS REAL WINDOW SIZES (real Shore, as shipped) ==")
	fmt.Fprintf(out, "   %-10s %-8s %-8s %-8s %-8s %s\n", "size", "glyphs", "run|dx|", "cells/s", "corr|dx|", "period")
	for _, sz := range [][2]int{{120, 26}, {124, 51}, {119, 51}, {107, 51}, {80, 24}, {153, 43}} {
		W, H = sz[0], sz[1]
		hy := hyRow(H)
		yy0, yy1 := hy+1, H-H/5
		if yy1 <= yy0+1 {
			yy1 = yy0 + 2
		}
		r := measure(realSource(driftSet), 300, yy0, yy1, trackGap, corrGap)
		fmt.Fprintf(out, "   %-10s %-8.0f %-8.3f %-8.2f %-8.2f %.1f\n",
			fmt.Sprintf("%dx%d", sz[0], sz[1]), r.glyphs, r.flowAbs,
			r.flowAbs/(float64(trackGap)*stepDT), r.locAbs, r.period)
	}
	W, H = 120, 26
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 8. ACTIVITY SWEEP ==")
	fmt.Fprintf(out, "   %-8s %-8s %-8s %-8s %s\n", "level", "glyphs", "run|dx|", "cells/s", "corr|dx|")
	saved := act.Level
	for _, lv := range []float64{0.0, 0.2, 0.4, 0.6, 0.8, 1.0} {
		act.Level = lv
		r := measure(realSource(driftSet), 300, y0, y1, trackGap, corrGap)
		if r.glyphs == 0 {
			fmt.Fprintf(out, "   %-8.2f EMPTY SAMPLE -- nothing drawn at this level, result VOID\n", lv)
			continue
		}
		fmt.Fprintf(out, "   %-8.2f %-8.0f %-8.3f %-8.2f %.2f\n", lv, r.glyphs, r.flowAbs,
			r.flowAbs/(float64(trackGap)*stepDT), r.locAbs)
	}
	act.Level = saved
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 9. GLITTER, READ WITH A MASK THAT CAN SEE IT ==")
	fmt.Fprintln(out, "   glitter() is the only x position in the file with explicit time in it,")
	fmt.Fprintln(out, "   and notes/drift never counted its bullet. Measured on its own glyph:")
	for _, g := range []scen{
		{"glitter ALONE, as shipped", knobs{glitter: true}},
		{"glitter ALONE, no time in its x", knobs{glitter: true, stillGlitterX: true}},
		{"whole scene, glitter visible", allOn()},
		{"whole scene, glitter x frozen", mut(func(k *knobs) { k.stillGlitterX = true })},
	} {
		r := measure(replicaSource(g.k, moonX, withGlitter), 300, y0, y1, trackGap, corrGap)
		if r.glyphs == 0 {
			fmt.Fprintf(out, "   %-38s EMPTY SAMPLE -- result VOID\n", g.name)
			continue
		}
		fmt.Fprintf(out, "   %-38s glyphs %-8.0f run|dx| %-8.3f slide %-8.3f breath %.3f\n",
			g.name, r.glyphs, r.flowAbs, r.trans, r.breath)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 9b. SESSION AGE. s.phase is an unbounded accumulator and it multiplies")
	fmt.Fprintln(out, "    an x-dependent term, so the picture at minute 60 is not the picture at")
	fmt.Fprintln(out, "    second 30. Same measurement, different amounts of history.")
	fmt.Fprintf(out, "   %-12s %-10s %-8s %-8s %-8s %-9s %s\n",
		"warm-up", "wave age", "glyphs", "run|dx|", "cells/s", "x-period", "peak x-period")
	savedWarm := warmFrames
	for _, wf := range []int{0, 25, 100, 400, 1600, 6400, 25600, 102400} {
		warmFrames = wf
		src := replicaSource(allOn(), moonX, driftSet)
		ms := src(150)
		fl, _ := track(ms, y0, y1, trackGap, 6)
		mp, pp := xPeriod(ms, y0, y1)
		n := 0.0
		for _, m := range ms {
			for y := y0; y < y1; y++ {
				for x := 0; x < W; x++ {
					if m[y][x] {
						n++
					}
				}
			}
		}
		if n == 0 {
			fmt.Fprintf(out, "   %-12d EMPTY SAMPLE -- result VOID\n", wf)
			continue
		}
		fmt.Fprintf(out, "   %-12d %-10.0f %-8.0f %-8.3f %-8.2f %-9.1f %.1f\n",
			wf, float64(wf)*stepDT*(0.55+act.Level*1.45), n, fl.meanAbs(),
			fl.meanAbs()/(float64(trackGap)*stepDT), mp, pp)
	}
	warmFrames = savedWarm
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 9c. THE SAME ROW, EARLY AND LATE (replica, whole scene) ==")
	for _, wf := range []int{0, 400, 6400, 102400} {
		warmFrames = wf
		src := replicaSource(allOn(), moonX, driftSet)
		ms := src(4)
		row := (y0 + y1) / 2
		fmt.Fprintf(out, "   warm-up %d frames (wave age %.0f):\n", wf, float64(wf)*stepDT*(0.55+act.Level*1.45))
		for i, m := range ms {
			b := make([]byte, 0, 80)
			for x := 20; x < 96; x++ {
				if m[row][x] {
					b = append(b, '#')
				} else {
					b = append(b, '.')
				}
			}
			fmt.Fprintf(out, "     f%-2d  %s\n", i, string(b))
		}
	}
	warmFrames = savedWarm
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 9d. THE MECHANISM, STATED AS A PREDICTION AND CHECKED ==")
	fmt.Fprintln(out, "    A mark's edge moves sideways at the wave's INWARD speed times the edge's")
	fmt.Fprintln(out, "    own |dx/dy|. Measure both halves separately and see if their product is")
	fmt.Fprintln(out, "    the sideways speed that was measured directly.")
	{
		ms := realSource(driftSet)(300)
		sl, n := edgeSlope(ms, y0, y1)
		fx, _ := track(ms, y0, y1, trackGap, 6)
		fy, _ := trackCols(ms, y0, y1, trackGap, 6)
		fmt.Fprintf(out, "   edge |dx/dy| %.2f cells per row (%.0f edges measured)\n", sl, n)
		fmt.Fprintf(out, "   inward |dy| %.3f rows per %.2fs\n", fy.meanAbs(), float64(trackGap)*stepDT)
		fmt.Fprintf(out, "   PREDICTED sideways %.3f cells,  MEASURED %.3f cells  (ratio %.2f)\n",
			sl*fy.meanAbs(), fx.meanAbs(), fx.meanAbs()/math.Max(1e-9, sl*fy.meanAbs()))
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 9e. COSTING THE FIX (replica; nothing shipped) ==")
	fmt.Fprintln(out, "    Steepen the crests, or slow the wave. Each row is the WHOLE scene with")
	fmt.Fprintln(out, "    one number changed, measured the same way as the baseline.")
	fmt.Fprintf(out, "   %-40s %-8s %-9s %-9s %-9s %-9s %s\n",
		"variant", "glyphs", "side c/s", "inward r/s", "side:in", "edge dx/dy", "x-period")
	costRow := func(name string, k knobs) {
		src := replicaSource(k, moonX, driftSet)
		ms := src(300)
		n := 0.0
		for _, m := range ms {
			for y := y0; y < y1; y++ {
				for x := 0; x < W; x++ {
					if m[y][x] {
						n++
					}
				}
			}
		}
		if n == 0 {
			fmt.Fprintf(out, "   %-40s EMPTY SAMPLE -- result VOID\n", name)
			return
		}
		fx, _ := track(ms, y0, y1, trackGap, 6)
		fy, _ := trackCols(ms, y0, y1, trackGap, 6)
		mp, _ := xPeriod(ms, y0, y1)
		dt := float64(trackGap) * stepDT
		sl, _ := edgeSlope(ms, y0, y1)
		fmt.Fprintf(out, "   %-40s %-8.0f %-9.2f %-9.2f %-9.2f %-9.2f %.1f\n",
			name, n, fx.meanAbs()/dt, fy.meanAbs()/dt,
			fx.meanAbs()/math.Max(1e-9, fy.meanAbs()), sl, mp)
	}
	costRow("baseline, as shipped", allOn())
	for _, kk := range []float64{0.24, 0.32, 0.48, 0.64} {
		k := allOn()
		k.seaWarpK = kk
		k.swellK = kk
		costRow(fmt.Sprintf("both x wavenumbers %.2f (shipped .16/.15)", kk), k)
	}
	for _, aa := range []float64{2.5, 3.5, 5.6} {
		k := allOn()
		k.seaWarpA = aa
		costRow(fmt.Sprintf("sea warp amplitude %.1f (shipped 1.5)", aa), k)
	}
	for _, sp := range []float64{0.75, 0.5, 0.25} {
		k := allOn()
		k.speedMul = sp
		costRow(fmt.Sprintf("wave speed x%.2f", sp), k)
	}
	for _, am := range []float64{1.5, 2.5, 4.0} {
		k := allOn()
		k.swellAmpMul = am
		costRow(fmt.Sprintf("swell crest height x%.1f", am), k)
	}
	{
		k := allOn()
		k.seaWarpK, k.swellK, k.seaWarpA, k.swellAmpMul = 0.32, 0.32, 3.0, 2.0
		costRow("k .32 + sea amp 3.0 + swell crest x2", k)
	}
	{
		k := allOn()
		k.seaWarpK, k.swellK, k.seaWarpA, k.swellAmpMul = 0.48, 0.48, 3.0, 3.0
		costRow("k .48 + sea amp 3.0 + swell crest x3", k)
	}
	costRow("sea() ALONE, as shipped", knobs{sea: true})
	costRow("swells() ALONE, as shipped", knobs{swells: true})
	{
		k := knobs{sea: true}
		k.seaWarpK, k.seaWarpA = 0.48, 2.0
		costRow("sea() ALONE, warp 2.0*sin(x*0.48)", k)
	}
	{
		k := knobs{swells: true}
		k.swellK, k.swellAmpMul = 0.48, 3.0
		costRow("swells() ALONE, k .48, crest x3", k)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "== 10. WHAT IT LOOKS LIKE: one sea row over 8 consecutive frames ==")
	{
		src := realSource(driftSet)
		ms := src(8)
		row := (y0 + y1) / 2
		for i, m := range ms {
			b := make([]byte, 0, 80)
			for x := 20; x < 96; x++ {
				if m[row][x] {
					b = append(b, '#')
				} else {
					b = append(b, '.')
				}
			}
			fmt.Fprintf(out, "   f%-2d row %d  %s\n", i, row, string(b))
		}
	}
}
