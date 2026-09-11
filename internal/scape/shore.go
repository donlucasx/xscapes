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

	// tideRow is the waterline's row BEFORE the wash, which is what the open
	// sea's ramp is anchored to under Tide. Taking it from the washed edge
	// re-ramps the whole sea every frame; see paintBG.
	tideRow int
	// tideAt is how far the water has withdrawn right now, in rows and as a
	// FLOAT: the waterline glides rather than stepping whole cells.
	tideAt float64

	// Where the moon landed this frame, so callers can anchor a label to it
	// without recomputing the position and drifting out of sync.
	moonX, moonY int
	// The disc's reach in cells from its centre, columns and rows, so a
	// caller can keep clear of it (the context readout).
	moonRX, moonRY int
	// The disc's radius in rows, and whether it is painted at all this frame.
	// Fixed by discGeom BEFORE the star fields run, so they can keep off it.
	moonRR      float64
	moonPainted bool
	// litTone and darkTone are the quad disc's two colours this frame, one
	// each for the whole disc (sampled against the sky at its centre), so
	// the lit face never rounds to two yellows across the sky's rows.
	litTone, darkTone term.RGB
	// quadDark records, per cell the quad disc painted, whether its shape
	// colour is the dark face -- the painter's own decision, for the test
	// that holds the dark face to one piece (a sky tone can round to the
	// dark tone by coincidence, so the frame alone cannot say).
	quadDark map[[2]int]bool

	// The constellation's layout, and the geometry it was laid out for. It
	// depends on nothing that changes between frames -- see starPlaces -- so it
	// is computed once per window size and kept.
	starKey starKey
	starPts [][2]int

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

	// SandFade sinks the lower beach toward black, 0 off and 1 fully dark at
	// the last row. Squared, so the sand stays sand for most of its depth and
	// only lets go near the bottom. NewShore sets DefaultSandFade; the field
	// stays settable for the tuner and for tests.
	SandFade float64

	// WriteRows reserves the bottom rows of the scape for the agent's work.
	//
	// Those rows are the page the last few actions are written on, and they
	// were not a page at all: the swell reached into them, so the text sat on
	// waves and foam and read against a background that changed under every
	// character. His words -- "because the sand changes colors its a bit
	// distracting from the bottom 4 lines".
	//
	// So they are held out of the water entirely and painted one flat colour.
	// Not the whole beach flat, which would throw away the depth the wet sand
	// carries; just the part that has writing on it.
	WriteRows int

	// SkyRows and SandRows override the proportional layout, in rows. Zero
	// means the default fractions.
	//
	// They exist because a fixed fraction is wrong for a tall window: the sky
	// is the least informative region on screen -- measured on a real 66x35
	// pane it paints 1.7 glyphs a row against the sea's 34.9 -- and giving it
	// 42% of every extra row spends the new space on nothing. The beach is
	// where the agent's actual work is written, so it is what should grow.
	SkyRows, SandRows int

	// MoonX is where the moon sits, as a fraction of the width. It moves with
	// the composition: the moon and the companion are the two things a glance
	// goes looking for, and stacking them in one column leaves half the frame
	// carrying nothing. Zero means the default.
	MoonX float64
	// MoonRim is the disc's edge: "hue" (the outer ring one tone darker in
	// the disc's own hue -- what ships, his pick of 2026-09-05 from "The Moon,
	// Four Ways"), "" (solid, what shipped from s11 to then) or "blend" (the
	// s7 mockup's fade, half toward the sky, which rounds to grey on 256).
	// NewShore sets "hue".
	MoonRim string
	// MoonEdge is a STUDY switch for how the disc's edge is sampled: "" (two
	// half-rows per cell, what ships) or "quad" (four quarters per cell, the
	// companion's own block glyphs used as two backgrounds -- twice the
	// horizontal resolution of the edge).
	MoonEdge string
	// SunShadow is a STUDY switch for the unlit face by day: "" (not painted,
	// what ships -- the sun wanes as a crescent) or "slate" (the dim tone the
	// moon uses at night, which is what shipped until 2026-09-06).
	SunShadow string
	// MoonHalo is a STUDY switch: a soft lightening of the sky around the
	// disc at night, the one piece of the s7 mockup's softness the grey ramp
	// can carry.
	MoonHalo bool

	// FlatCaps collapses the disc's edge cells to whole cells where the edge
	// is locally HORIZONTAL, and it is the answer to the last of his hairline
	// reports. Every split cell on Terminal.app leaks about one device pixel
	// of its background at the cell's bottom edge -- U+2584's ink stops at
	// 29.4 of a 30px row -- and no glyph avoids it: the full block's ink runs
	// 4.6..29.4, so whatever is drawn there the last row is the background.
	// Measured in his 8.27.52 PM crop of 2026-09-07, at the disc's centre
	// column: 17px of sky, 12px of rim, then ONE pixel of (100,113,134), which
	// is 0.43*rim + 0.57*sky to within two counts on every channel.
	//
	// What makes that pixel a LINE rather than a speck is the RUN. Counted
	// over 6 widths x 7 heights x 48 half-hours (notes/rulecount): the whole
	// frame carries four rules and all four are on the disc -- two of them
	// FIVE cells wide, at the top and bottom caps, and two one cell wide at
	// the shoulders. The caps are where the silhouette is flat, and a split
	// cell whose neighbours split at the same height buys no roundness at all;
	// it only draws seventy pixels of rule. So the caps take a whole cell and
	// the shoulders keep their split, which is where the curve actually lives.
	//
	// Gated on term.NoSplitCells, so only the terminal with the defect pays
	// for it: Ghostty draws the blocks pixel-exact and keeps the full half-row
	// edge everywhere. Set false to study the shipped-until-now edge.
	FlatCaps bool

	// writeTop is the first row of the writing band, or c.H when there is none.
	writeTop int

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

// MoonExtent is how far the disc reaches from MoonPos, in columns and rows.
func (s *Shore) MoonExtent() (rx, ry int) { return s.moonRX, s.moonRY }

func NewShore(seed int64, asciiOnly bool) *Shore {
	return &Shore{Seed: seed, ASCII: asciiOnly, SandFade: DefaultSandFade, WriteRows: DefaultWriteRows, MoonRim: "hue", FlatCaps: true}
}

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
	if s.SkyRows > 0 {
		hy = s.SkyRows
	}
	// The writing band is carved off the bottom FIRST, and everything else is
	// laid out in what is left.
	//
	// Doing it the other way round is what flattened the shoreline: the mean
	// waterline was placed against the bottom of the canvas and then clamped
	// off the band, so every trough of the swell hit the same row and the sea
	// met the sand along a ruled line. Give the swell the room above the band
	// instead and the waterline gets its shape back.
	// The writing band scales with the scene.
	//
	// Four rows is right in a tall scape and swamps a short one: at seventeen
	// rows -- which is what a 43-row window gives the scape -- four rows of
	// writing plus its wash is 41% of the picture, and the sea is squeezed to
	// three rows. WriteRows is the ceiling, a sixth of the height is the rule,
	// and two lines is the least worth reserving a band for.
	wr := s.WriteRows
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

	// Mean waterline. The beach is a proportional share of the scene, and the
	// writing band is INSIDE that share, not added to it.
	//
	// Added to it is what went wrong: the band took four rows off the bottom
	// and the beach was then laid out in what was left, so the sand grew by the
	// whole band and the sea lost it -- "now we have a lot of beach and not
	// enough sea. Did u have to eat into the sea to extend the beach?" Yes, and
	// this is where. The floor is the band plus a few rows for the wash to run
	// up and back, which is the least that still reads as a shore.
	beach := c.H / 5
	if min := wr + 3; wr > 0 && beach < min {
		beach = min
	}
	if beach < 4 {
		beach = 4
	}
	if s.SandRows > 0 {
		beach = s.SandRows
	}
	// Everything sized in rows has to scale, or the scene that is composed at
	// 80x24 turns into a giant moon and a wall of foam at 40x12.
	scale := math.Min(float64(c.W)/80.0, float64(c.H)/24.0)
	scale = math.Max(0.45, math.Min(1.35, scale))

	sy := c.H - beach
	// HIS idea, shipped on by default 2026-09-10: the water withdraws up the
	// frame when the agent goes quiet and comes back as it works. XSCAPES_TIDE=0
	// restores the old fixed waterline. See tide.go.
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > writeTop-1 {
		sy = writeTop - 1
	}

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
	// The tide EASES toward where the activity puts it. Snapping was caught by
	// his own TestActivityChangeDoesNotTeleportTheSea: a level step jumped the
	// waterline 1.03 rows against 0.02 in a normal frame, 47x, where the guard
	// allows 8x. A tide that arrives in one frame is not a tide.
	s.tideAt += (tideTarget(act.Level, scale) - s.tideAt) * math.Min(1, dt/TideEase)
	s.tideRow = sy - int(math.Round(s.tideAt))
	s.phase += dt * (0.55 + act.Level*1.45)
	tt := s.phase
	edge := s.waterline(c.W, sy, tt, act, scale)
	// Scale the swell to the room it has instead of clipping it.
	//
	// Clipping is what drew the ruled shoreline: every trough that reached past
	// the writing band landed on the same row. Scaling keeps the shape -- the
	// crests and hollows stay where they are, just shallower -- so the water
	// still runs up the sand and back, and never touches the writing.
	if writeTop < c.H {
		// Measured from where the water actually IS, not from where it would
		// be with no tide. This guard exists to stop the SWELL reaching the
		// writing band; under Tide the whole edge is displaced from sy by up
		// to five rows, so measuring against sy counted the tide itself as
		// swell and squashed it flat -- the tide's range collapsed from 5.1
		// rows to 1.7 and every quiet level pinned to the same row.
		ref := float64(sy)
		if Tide {
			ref = float64(s.tideRow)
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
	s.lastEdge = edge
	s.writeTop = writeTop

	s.paintBG(c, hy, edge)
	s.discGeom(c, hy, scale, 1-clamp01(act.ContextUsed))
	s.stars(c, hy, t, foamCeiling(sy, scale), act.TodoDone, act.TodoTotal)
	s.moon(c, hy, scale, 1-clamp01(act.ContextUsed), moonVis(s.pal))
	s.todoStars(c, hy, foamCeiling(sy, scale), act.TodoDone, act.TodoTotal)
	s.sea(c, hy, edge, tt, act)
	s.sand(c, edge)
	s.foam(c, edge, scale)
}

// waterline is where the sea meets the sand, per column, as a float. Keeping it
// fractional is the whole trick: the shoreline is painted into the background
// colour at sub-cell precision, so a two-row swing reads as one smooth curve
// instead of the staircase you get from rounding to a row.
func (s *Shore) waterline(w, sy int, tt float64, act Activity, scale float64) []float64 {
	if Tide {
		return tideEdge(w, sy, hyFloor(sy), s.tideAt, tt, act, scale)
	}
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

// voidCol is what the beach falls away into: not pure black, which reads as a
// hole punched in the frame, but the near-black a dark terminal already sits
// on, so the scene ends by running out rather than by being cut off.
var voidCol = term.RGB{R: 10, G: 10, B: 12}

// DefaultSandFade is locked at full (2026-09-01).
//
// Measured on the newest line of writing, which sits lowest and matters most,
// the contrast against its background goes from 132 to 204 at midday and 148
// to 204 at night. The equality is the real argument: a black bottom row is
// black at every hour, so legibility stops depending on the clock, which a
// flat beach could never manage -- it was always a compromise between noon
// sand and midnight sand.
//
// Full rather than a softer 0.8 because of where the scene is going: with the
// agent rendered inside the scape, a beach that ends in true black merges into
// the agent's own black background instead of stopping at a visible seam.
//
// Middle values are the worst place to sit. The ink flips from dark to light
// as the beach crosses mid-luma, so a half fade at midday actually cost
// contrast (132 down to 129) before the ink was taught to read the real
// background. Either off, or 0.8 and up.
const DefaultSandFade = 1.0

// sandFadeStart is how far down the beach the fade begins, as a fraction. The
// top of the sand has to stay sand: it meets the water, and a beach that
// started darkening at the waterline would read as wet, which is a thing the
// palette already says with WetSand.
const sandFadeStart = 0.30

// DefaultWriteRows matches reduce.TailLen: the beach holds four lines, so four
// rows are kept dry for them.
const DefaultWriteRows = 4

// The writing band's three sand tones, and why they are these exact numbers.
//
// They are cube-exact: each is a colour the xterm-256 palette contains
// outright, so it survives quantisation unchanged and looks the same on a
// 256-colour terminal as it does in truecolor.
//
// That is not fussiness. The cube has six levels per channel and almost no
// browns, so a sand tone chosen for how it looks in truecolor gets snapped to
// whatever is nearest -- measured across the day, the band landed on olive at
// dawn, dusty rose at nine, dusty red at dusk and flat grey at night. Four
// different wrong colours, which is exactly why the sand did not read as sand.
// Choosing from what the palette actually holds is the only way to get a brown
// that stays brown.
//
// An earlier version made this band black. It measured beautifully and was
// wrong: "it should be sand, not black". Legibility comes from the ink, which
// flips light or dark against whatever the band is. The band's job is to be a
// beach.
var (
	sandDay   = term.RGB{R: 215, G: 175, B: 135} // 256 index 180
	sandLow   = term.RGB{R: 175, G: 135, B: 95}  // 256 index 137
	sandNight = term.RGB{R: 135, G: 95, B: 0}    // 256 index 94
)

// writeBandColor picks by how bright the beach is at this hour, so the page
// darkens with the day without ever leaving those three tones.
func writeBandColor(p Palette) term.RGB {
	switch l := bandLuma(p.SandNear); {
	case l >= 160:
		return sandDay
	case l >= 100:
		return sandLow
	default:
		return sandNight
	}
}

// wetBandColor is the wet strip at the water's edge: the beach's own staircase,
// one tone DOWN.
//
// It used to be the interpolated palette value, and that is what he
// photographed on 2026-09-09 -- flat grey blocks along the shore, five cells
// wide, under the litter. The cube's channel levels are 0, 95, 135, 175, 215,
// 255, so below 95 in green and blue there is only zero: the darkest entry
// keeping a warm ordering is sandNight at luma 96, and under that the cube has
// only the pure reds, while the grey ramp steps by ten the whole way down.
// Palette.WetSand runs luma 52 to 108 across the day, so it spends most of it
// under that floor and the nearest colour the terminal owns is a grey.
// Measured hour by hour in notes/wetsand: warm at 09:00, GREY at 10:00, warm
// 11:00 to 14:00, grey from 15:00 on -- the strip changed hue back and forth
// through a working morning.
//
// The dry beach never had this problem because writeBandColor never leaves
// three cube-exact tones. This is the same three tones, shifted one step, so
// the wet strip is exactly as stable as the sand it sits on and is always one
// rung darker.
//
// At night the beach is already on the last warm tone the cube has, so there is
// nothing warm below it and the strip falls off the end into a grey. That is
// deliberate, and it is consistent there: by then the whole scene is dark, and
// the locked rule is that the darkness lives in the backgrounds while the
// colour lives in the glyphs.
func wetBandColor(p Palette) term.RGB {
	switch l := bandLuma(p.SandNear); {
	case l >= 160:
		return sandLow
	case l >= 100:
		return sandNight
	default:
		return wetNight
	}
}

// wetNight is a grey-ramp entry, chosen for its distance below sandNight
// rather than for a hue it cannot have.
var wetNight = term.RGB{R: 68, G: 68, B: 68}

func bandLuma(c term.RGB) float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}

func (s *Shore) paintBG(c *canvas.Canvas, hy int, edge []float64) {
	fh := math.Max(1, float64(hy))
	// The DRY beach is measured from the mean waterline, not from each
	// column's own crest.
	//
	// Measured: with a per-column anchor a single beach row carried 87
	// distinct colours across 120 cells, because every column got its own
	// ramp -- and the anchor moves with the swell, so the whole beach
	// re-coloured every frame. On a 256-colour terminal those near-identical
	// browns quantise apart into olive, dusty red and grey, which is why it
	// read as blotches rather than sand and would not sit still.
	//
	// The water keeps its per-column edge: the tide has to move, and the wet
	// cell at the waterline still straddles sea and sand. It is only the page
	// the agent's work is written on that holds still.
	mean := 0.0
	for _, e := range edge {
		mean += e
	}
	if len(edge) > 0 {
		mean /= float64(len(edge))
	}
	// The sky and the open sea are each ONE ramp, painted as a path through
	// the palette on 256 -- see term.Ramp. Each cell is told the span of the
	// ramp it covers, so a band edge that falls inside a row lands on the
	// right half of it.
	sky := term.NewRamp(s.pal.SkyTop, s.pal.SkyHorizon)
	sea := term.NewRamp(s.pal.SeaFar, s.pal.SeaNear)
	depth := math.Max(1, mean-float64(hy))
	if Tide {
		// The open sea's ramp is anchored to the TIDE's row, not to the mean of
		// the washed edge, and this is not a detail: the wash moves the edge
		// every single frame, so a depth taken from it re-ramps the whole sea
		// twelve times a second. Caught by his own guarantee --
		// TestTheBackdropHoldsStillBetweenFrames allows 8% of open-sea cells to
		// change between frames and the first tide build churned 59.13%, which
		// is the exact defect the per-column anchor caused and the note above
		// this describes. sy moves only when the ACTIVITY moves it, so the
		// backdrop holds still and steps when the tide does.
		depth = math.Max(1, float64(s.tideRow)-float64(hy))
	}
	for x := 0; x < c.W; x++ {
		ex := edge[x]
		for y := 0; y < c.H; y++ {
			fy := float64(y)
			var col term.RGB
			switch {
			case y <= hy:
				c.SetBGRamp(x, y, sky, (fy-0.5)/fh, (fy+0.5)/fh)
				continue
			case fy < ex-0.5:
				// Depth, from the mean waterline for the same reason the beach
				// is: this ramp says how far out the water is, which is a
				// property of the row, not of whichever crest happens to be in
				// this column. Per-column it stretched and squashed frame to
				// frame and put dozens of near-identical blues in a single row,
				// which 256 colours then pulled apart into bands.
				//
				// The wave motion did not live here anyway. It lives in the
				// glyphs and in the waterline cell below, which keeps its own
				// per-column edge.
				c.SetBGRamp(x, y, sea, (fy-0.5-float64(hy))/depth, (fy+0.5-float64(hy))/depth)
				continue
			case fy < ex+0.5:
				// The waterline cell straddles sea and sand. Mix by how much of
				// the cell the water actually covers.
				mix := ex - fy + 0.5
				if Tide {
					// Snapped to five steps. The tide's coast is FLAT -- under
					// a row of amplitude -- so the whole waterline lands in one
					// row and every column gets its own slightly different mix:
					// 56 distinct tones on one row, against the 40 his
					// TestNoRowIsAConfettiOfNearIdenticalTones allows, which is
					// the blotchy-sand defect coming back by another door.
					// Five steps is enough to read as a wet edge and few enough
					// that the row is a band rather than a gradient.
					mix = math.Round(mix*4) / 4
				}
				col = term.Lerp(wetBandColor(s.pal), s.pal.SeaNear, mix)
			case s.writeTop > 0 && y >= s.writeTop:
				// One flat tone, all the way across and all the way down. This
				// is the page, not the picture.
				col = writeBandColor(s.pal)
			default:
				// ALL the sand is one tone, and it is the water's edge that reads
				// against it: "ok for all the sand to be the same color (which
				// should vary from day to night), but we should be able to see the
				// water receding".
				//
				// This supersedes the graded beach and, with it, SandFade. The fade
				// was locked at full to buy contrast for the newest line and to let
				// the scene dissolve rather than stop at an edge. The flat sand buys
				// the contrast instead -- measured, and the writing sits directly on
				// it -- and the fade's other job had turned into a defect: a graded
				// beach above a flat band ended its last row in black and drew a hard
				// black line across the shore.
				col = writeBandColor(s.pal)
			}
			c.SetBG(x, y, col)
		}
	}
}

// ambientGlyphs is the dust's vocabulary, and it is wider than it was on his
// note of 2026-09-11: "any ideas around different characters? ... I see more
// characters varying in sizes and shapes appearing."
//
// No '*'. It belongs to the checklist, and a channel that shares a glyph with
// the scenery is not a channel: at midnight the ambient field drew stars
// indistinguishable from finished todos and the count could not be read off the
// sky at all. Nothing here reads as an asterisk either, which is the same rule
// one step further out -- a four-pointed sparkle would have cost the count just
// as surely as the asterisk itself.
//
// The variety is in SHAPE and in where the ink sits inside the cell, which is
// what "different sizes" looks like at terminal type sizes: the period and the
// comma sit on the baseline, the grave, the apostrophe and the double quote
// hang at the cap line, the middle dot and the colon are centred, and '+' and
// '°' are the two that fill a cell.
//
// Repeats are weights, and the slice is read by index so a repeat is simply a
// second slot: plain round dots take six of the thirteen, which is what keeps
// the sky reading as dust with punctuation in it rather than as typography.
//
// ⚠ Every one of these was measured in Menlo Regular before it was picked, not
// assumed from its codepoint: all thirteen are present and all thirteen have
// the same advance as 'M' (1233/2048 em). Half of them are East Asian
// Ambiguous, which is the same class as U+2580 and U+2584 -- the blocks the
// whole scape is drawn with -- so they carry no width risk this project has not
// already taken everywhere.
var ambientGlyphs = []rune{
	'.', '.', '.',
	'\u00b7', '\u00b7',
	',', '`', '\'',
	':', '"',
	'+', '\u00b0',
	'.',
}

// ambientPeak is the field's density at the top of the sky at the darkest hour.
// It is a fraction of cells, so at his 125x28 -- eleven rows of sky -- the
// midnight field is about forty-five specks.
const ambientPeak = 0.052

// ambientClockGamma bends StarVis into a share of that density.
//
// A straight multiply is too steep: mid-afternoon is StarVis 0.175, which would
// leave eight specks in a sky that already reads bare. The exponent keeps the
// day and the night ordered while giving the daylight hours a field that is
// actually there -- 0.175 becomes a share of 0.46, and 1.0 stays 1.0.
const ambientClockGamma = 0.45

// ambientContrast is the luma a speck must clear against the ground it is
// actually painted on, measured on the RENDERED cell. Half the constellation's
// bar: dust that competes with a finished todo would be reporting something.
const ambientContrast = 24.0

// ambientDimmest is the faintest a speck may start. The search only ever runs
// UPWARD from a speck's own magnitude, so this is a floor and not a ceiling.
// ambientDayFloor is HIS RULING, 2026-09-11: "make sure we can still see some
// during daytime, even if fainter than at nighttime."
//
// ⚠ IT OVERRIDES A DESIGN DECISION AND THAT IS WORTH SAYING. StarVis reaches
// exactly 0 at noon, and CLAUDE.md records the intent -- "the moon and the
// constellation are washed out at midday by design" -- because the sky is the
// WORLD and a real midday sky has no stars in it. He has ruled the other way,
// for the thing that matters more here: a sky with nothing in it reads as a
// scene that has stopped working, and he said so twice before this ("this felt
// too bare").
//
// So the clock keeps its whole range and gains a floor under it: at midnight
// the field is at full density, at noon it is at this fraction of it. Measured
// at his 125x28, that is 26 specks at midnight and 3 at noon -- present, sparse,
// and every one of them legible, which is the half of his note that matters
// ("even if fainter"). Before the count-not-alpha change, noon drew zero and
// dusk drew 39 of which 20 could not be told from the sky behind them.
const ambientDayFloor = 0.25

const ambientDimmest = 0.45

// ambientTwinkle is how much the shimmer may LIFT a speck above the rung that
// was measured. It only ever adds.
const ambientTwinkle = 0.30

// ambientInkSteps is how finely the lift from the palette's star tone to white
// is walked, the same eight rungs the constellation uses.
const ambientInkSteps = 8

// stars is the ambient dust field -- the sky's texture, not the checklist.
//
// ⭐ HIS REPORT, 2026-09-11: "during the daytime I dont see any other characters
// tho like I saw before during nightime giving depth to the constellation." He
// had reported the same sky twice the day before, an hour apart: 20:22 read
// bare and 20:55 read rich.
//
// BOTH OF THOSE FRAMES PLOTTED THE SAME THIRTY-NINE SPECKS. What differed was
// how many survived the cube. The field used to encode the clock in ALPHA --
// far.Plot(x, y, g, Star, (0.55+0.45*sin) * StarVis) -- and at a low StarVis the
// quantised foreground lands on the SAME cube index as the sky behind it. The
// specks were drawn and could not be seen: 53% of them lost at 16:00, 51% at
// 20:22, 0% at 20:55.
//
// ⭐ SO THE CLOCK MOVED TO COUNT, which is the house rule and not a preference:
// encode in coverage, count or position, never in rate -- and an alpha that
// fades a thing to invisible is a rate wearing a different coat. StarVis now
// decides how many cells are ELIGIBLE, through the same hash threshold the row
// falloff already used, and every speck that qualifies is drawn at an opacity
// MEASURED to clear its own background. Because the threshold moves and the
// hash does not, a daylight field is a strict SUBSET of the midnight one: the
// specks thin out, none of them moves.
//
// ⚠ AND THE GATE IS GONE. `if twinkle <= 0.02 { continue }` did not dim a
// speck, it deleted it, so below StarVis 0.2 every speck crossed the gate once
// a cycle and switched off -- 0.47 blink-outs a frame at 08:00, which at
// inside.go's 12 fps is about six disappearances a second. That is his "stars
// appearing and dissapearing" of 2026-09-10. Eligibility is now a fact about
// the CELL and the hour; the shimmer can only lift a speck above the rung that
// was measured, never below it, so it cannot take one away.
//
// ⚠ NOON STAYS EMPTY. StarVis is 0 at noon and the sky is the WORLD: stars are
// not visible at midday and this may not make them so. The threshold is zero
// there and the compare is >=, so no hash value can pass it.
//
// ⚠ AND IT KEEPS OFF THE CONSTELLATION'S OWN CELLS, which the old field did
// not. compositeBG BLENDS the layers rather than letting the top glyph own the
// cell, and a lit star's alpha is its magnitude -- 0.70 at the dimmest -- so
// whatever the far layer left underneath tints its ink. Measured before and
// after this change over 399 star cells at four geometries and seven hours: the
// two fields share no cell at his 125x28, but they share one at 143x27 and two
// at 80x24, and raising the dust's opacity moved FIVE of those stars by one
// cube step. One step is nothing to look at and the channel is still exact --
// which is the point. The checklist is a count of finished work and its pixels
// are not the dust's to move.
func (s *Shore) stars(c *canvas.Canvas, hy int, t float64, foamTop, done, total int) {
	far := c.Far()
	taken := s.constellationCells(c.W, hy, foamTop, done, total)
	share := ambientDayFloor + (1-ambientDayFloor)*math.Pow(s.pal.StarVis, ambientClockGamma)
	for y := 0; y < hy; y++ {
		// Thin out toward the horizon, where haze would eat them.
		density := ambientPeak * (1 - float64(y)/math.Max(1, float64(hy))*0.75) * share
		for x := 0; x < c.W; x++ {
			if HashF(x, y, s.Seed) >= density {
				continue
			}
			// Never on the disc. moon() paints backgrounds, so a glyph left
			// under it survives and composites against the body -- his report
			// of 2026-09-06, one tan speck on the sun's face at 143x62. The
			// checklist field has kept off the disc since it was written; the
			// ambient field never did.
			if s.DiscCovers(x, y) {
				continue
			}
			if taken[y*c.W+x] {
				continue
			}
			ph := HashF(x, y, s.Seed+7) * 2 * math.Pi
			// The shimmer is a LIFT, never a fade: 1.0 at the bottom of the
			// cycle, so the rung the search clears is also the floor.
			lift := 1 + ambientTwinkle*(0.5+0.5*math.Sin(t*0.9+ph))
			mag := ambientDimmest + HashF(x, y, s.Seed+11)*(1-ambientDimmest)
			g := ambientGlyphs[int(HashF(x, y, s.Seed+3)*float64(len(ambientGlyphs)))%len(ambientGlyphs)]
			s.speck(c, far, x, y, g, mag*lift)
		}
	}
}

// constellationCells is where todoStars is about to put its lit stars, so the
// dust can stay out of them. It asks the same question with the same arguments,
// so it gets the same memoised layout (starPlaces is keyed on the geometry and
// the disc, both of which discGeom has already fixed for this frame) and the
// checklist recomputes nothing. Reading only: nothing here decides anything
// about the constellation, it just declines to paint underneath it.
func (s *Shore) constellationCells(w, hy, foamTop, done, total int) map[int]bool {
	if done <= 0 || total <= 0 || hy < 3 {
		return nil
	}
	if done > total {
		done = total
	}
	top, bot := starBand(hy, foamTop)
	if bot < top {
		return nil
	}
	pts := s.starPlaces(w, hy, top, bot, total)
	out := make(map[int]bool, done)
	for i := 0; i < done && i < len(pts); i++ {
		x, y := pts[i][0], pts[i][1]
		if x < 0 || x >= w || y < 0 || y >= hy {
			continue
		}
		out[y*w+x] = true
	}
	return out
}

// speck draws one mote at the faintest ink and opacity that can actually be
// SEEN in that cell, and draws nothing where none can.
//
// This is starInk's method applied to the dust, and it is here for the same
// reason: a fixed alpha is a guess about a background, and the background is a
// gradient that moves all day. The difference is the bar and what happens when
// it cannot be met -- the constellation must appear, so it takes the best tone
// there is; the dust is texture, so a cell that cannot carry a visible mote
// simply does not get one. That is what keeps the count honest: everything this
// function plots is on screen.
//
// The opacity rungs run from the speck's own magnitude up through the far
// layer's full alpha to opaque. The far layer is 0.30, so 1/far.Alpha is the
// cell alpha at which Blend clamps to the ink itself -- the only way a mote on
// a bright midday sky reaches the bar at all. At night the first rung clears on
// its own and nothing is spent.
func (s *Shore) speck(c *canvas.Canvas, far *canvas.Layer, x, y int, g rune, a0 float64) {
	i := y*far.W + x
	was := far.Cells[i]
	opaque := a0
	if far.Alpha > 0 {
		if o := 1 / far.Alpha; o > opaque {
			opaque = o
		}
	}
	white := term.RGB{R: 255, G: 255, B: 255}
	for _, a := range [3]float64{a0, math.Max(a0, 1), opaque} {
		for step := 0; step <= ambientInkSteps; step++ {
			ink := term.Lerp(s.pal.Star, white, float64(step)/ambientInkSteps)
			far.Plot(x, y, g, ink, a)
			// Read the cell back rather than predicting it. A glyph collapses
			// a split background to the ramp's mid tone, so the ground a speck
			// really sits on is not the ground the empty cell reported.
			_, fg, bg := c.ResolveAt(x, y, term.Profile256)
			if d := bandLuma(fg) - bandLuma(bg); d >= ambientContrast || -d >= ambientContrast {
				return
			}
		}
	}
	far.Cells[i] = was
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
// ⚠ THE CHECKLIST'S OLD LEGIBILITY FLOOR WAS A FIXED ALPHA, todoStarFloor =
// 0.85, and it is gone. The reasoning behind it is not, and it is why this
// channel does not use StarVis: a completed todo is a fact about the AGENT and
// the ambient field's visibility is a fact about the WORLD, StarVis is 0 at
// noon, so hanging the checklist on the clock would switch it off for the whole
// working day -- the exact bug moonVisFloor exists to fix, arriving a second
// time through a different door.
//
// What changed is that a fixed alpha was standing in for a measurement. At 0.85
// with the palette's star tone, measured at the 40-luma bar the tests enforce,
// the checklist stops reading at 0.36 of the sky at the worst hour of the day
// -- SHALLOWER than the band that shipped. Legibility is bought by measuring
// against the painted ground instead, which is what lets the band go deep
// enough to answer his note.
//
// todoStarContrast is the contrast a lit star is held to, in luma, against the
// ground it is actually painted on. It is 55 rather than the 40 the tests
// enforce because the quantiser gets a vote: the ink is chosen in true colour
// and the eye reads the cube entry nearest it, and a cube step is worth up to
// about ten luma. The margin is the quantiser's, not decoration.
const todoStarContrast = 55.0

// A star's MAGNITUDE is its alpha, and it runs from starDimmest up to fully
// opaque. It is fixed by the slot's index and the seed, so it is a fact about
// that star forever, and starInk raises it from there wherever the sky needs
// it -- so it only ever runs UPWARD from the floor and nothing is dimmer for
// carrying it.
//
// Every star used to be the same tone, which is a thing skies do not do and
// dashboards do: a row of identical asterisks reads as UI.
//
// starDimmest is 0.70 and it is measured, not picked. At 0.85, where the alpha
// used to sit, thirty-two stars carry THREE distinct tones at night; at 0.70
// they carry six, and the faintest still reads +133 luma above its own ground
// against a bar of 40. Lower buys more tones and starts to cost the count,
// which is the one thing this channel may not spend.
const starDimmest = 0.70

func starMagnitude(i int, seed int64) float64 {
	return starDimmest + HashF(i, 53, seed+43)*(1-starDimmest)
}

// constellationSlots is how many places the sky lays out at a minimum, and it
// mirrors reduce.StarsCap. It does NOT decide how many are drawn: a longer
// checklist simply gets more places, and because each place depends only on the
// places below it, the first thirty-two are the same either way.
const constellationSlots = 32

// starKey is the geometry a cached constellation layout belongs to. moonR is
// the disc's radius in sixty-fourths of a row, because the corridor the layout
// keeps clear is measured from it.
type starKey struct {
	w, hy, top, bot, n int
	moonX, moonR       int
	rx, ry             int
}

// starBand is the slice of sky the constellation is allowed to use: rows
// top..bot inclusive.
//
// HIS NOTE, 2026-09-10: "can we spread them more vertically? currently sitting
// in a narrow band pretty high up". It was rows 1..hy*3/5 -- at his 125x28 that
// is six rows of eleven, all of them the top half of the sky.
//
// The bottom is MEASURED rather than chosen, and the thing that measures it is
// legibility. A star's ink is lifted against the background it sits on (see
// starInk), and the deepest row that still clears the 40-luma bar at the worst
// hour of the day -- 17:45, when the horizon is pale and warm -- is 0.82 of hy
// at his geometry, 0.80 at the smallest scape the design targets and 0.86 at
// the tallest.
//
// Three quarters rather than the four fifths that sweep allows, because four
// fifths meets the bar EXACTLY: over 8192 rendered stars, 8 geometries x 32
// hours, the dimmest reads +40.0 at 4/5 and +47.0 at 3/4. It costs nothing
// where he actually works -- at 125x28 and 143x27, hy is 11 and both fractions
// give row 8 -- and buys seven luma everywhere else.
//
// Below that the sky is simply brighter than white ink can beat: at midday the
// horizon row is luma 208 and the brightest ink there is 255, which after the
// glyph alpha is worth 40 exactly. There is no clever tone that reaches lower,
// which is why this is a measurement and not a preference.
func starBand(hy, foamTop int) (top, bot int) {
	// And never in the water. foamTop is the highest row the foam can ever
	// reach at this geometry (foamCeiling), so hard is the last row a star may
	// have at all and easy leaves a clear row under the deepest one.
	//
	// This bites at the short end and it bit the band that shipped too: at
	// 40x12 the sky is five rows and the tide pulls the water up to row four,
	// which was inside the old band as well. The foam ate a star there -- it
	// plots into the same near layer, after the constellation, and it wins the
	// cell, so the star is simply deleted.
	hard, easy := foamTop-1, foamTop-2
	top, bot = 1, hy*3/4
	if bot > easy {
		bot = easy
	}
	// The top row is a margin, and a margin is a luxury a thin sky cannot
	// afford. At 40x12 the band would be two rows of about thirty usable
	// columns for a cap of thirty-two stars -- the last few forced into
	// neighbouring cells, and a pair in neighbouring cells reads as one, which
	// is the count error this channel exists to avoid. Giving the row back
	// takes the closest pair at that size from 1.0 screen units to 2.0.
	if bot-top < 2 {
		top = 0
	}
	// And then, only if the water leaves it, one more row -- the spare one
	// rather than a wet one.
	if bot-top < 2 && bot+1 <= hard {
		bot++
	}
	if bot > hard {
		bot = hard
	}
	// bot < top means the water is in the sky and there is no band at all. The
	// caller draws nothing; at those heights -- eight and nine rows, where the
	// tide reaches row one of a three-row sky -- there is no sky to draw in.
	return top, bot
}

// foamCeiling is the highest row foam() can ever speckle at this geometry: one
// row above the waterline's own ceiling, and rounded the way foam() rounds --
// int(e+0.5)-1, not floor(e)-1. The difference is one row and it put a star in
// the foam at 30x8 with XSCAPES_TIDE=0, which is the geometry and the switch
// that a floor() version passed.
func foamCeiling(sy int, scale float64) int {
	return int(math.Floor(seaCeiling(sy, scale)+0.5)) - 1
}

// seaCeiling is the highest row the waterline can ever reach at this geometry,
// as a float. It is a static property -- the tide's range, the wash and the
// coast's own shape are all scaled by `scale` and bounded by their own
// amplitudes -- so nothing that depends on it moves with the activity or the
// frame, which is what lets the constellation be laid out against it.
//
// The swell rescale in Update only ever SHRINKS the deviation from the base
// row, so a real frame's edge is never higher than this.
func seaCeiling(sy int, scale float64) float64 {
	if !Tide {
		// reach (0.8 + 2.1*level), worst at full activity, and the second sine.
		return float64(sy) - 3.45*scale
	}
	// The tide's two terms pull in OPPOSITE directions with the activity: the
	// withdrawal is (1-level)*TideRange and the wash is (0.7 + 0.9*level), so
	// taking the worst of each independently is a row too pessimistic -- at
	// 80x24 it costs the band a whole row for a level that cannot happen.
	// Sweep the level instead and take the highest the edge actually gets.
	best := math.MaxFloat64
	for step := 0; step <= 32; step++ {
		level := float64(step) / 32
		base := float64(sy) - (1-level)*TideRange*scale
		if f := hyFloor(sy); base < f {
			base = f
		}
		// The wash, and the coast's own two sines (0.60 + 0.35).
		if e := base - (0.7+0.9*level)*scale - 0.95*scale; e < best {
			best = e
		}
	}
	return best
}

// starCellAspect is how much taller a terminal cell is than it is wide, and it
// is why the layout below measures in "screen units" rather than cells.
//
// Two stars one ROW apart are about as far apart on glass as two stars two
// COLUMNS apart. Spreading by cell distance therefore packs rows twice as
// tightly as columns, which is exactly the thing that makes a pair read as one
// mark -- a miscount, in a channel whose entire job is a count.
const starCellAspect = 2.0

// starMinSep is how far apart two stars must be, in screen units, before the
// layout stops looking for somewhere better -- about three columns on a row, or
// a row and a half straight up. Below that a pair starts to read as one mark,
// and the count is the channel.
//
// starDarts is the budget. It is spent only on collisions, so a sparse sky uses
// one dart per star; the whole layout is computed once per window size anyway
// (see the cache) rather than per frame.
// starHardSep is the floor under which a dart is not accepted at all: below it
// a pair starts to merge, so the layout stops sampling and takes the roomiest
// cell in the band instead of the roomiest dart.
// starSpreadN is how many stars are placed as far apart as the band allows
// before the layout relaxes into dart-throwing. Four, because four is what the
// two guarantees leave: the first few must be spread (s27, his two specks in
// the corner) and the rest must not be (2026-09-10, his "instead of from left
// to right"). Measured, three of the four corners of the band are taken by star
// four, and by twelve stars the column-gap spread is back at 1.0.
const (
	starMinSep  = 3.0
	starHardSep = 2.0
	starSpreadN = 4
	starDarts   = 64
)

// starPlaces lays the constellation out as blue noise: random-LOOKING, with a
// separation nothing random gives you.
//
// The problem this solves is HIS, 2026-09-10: "can we have the stars appear
// randomly across the sky, instead of from left to right?"
//
// The history, so nobody reverts into it. The first layout was
// frac = (i+0.5)/total: strictly left to right, and at 153 columns his
// first-ever lit constellation was two specks in the far-left corner with
// sixteen slots still short of halfway. The fix for that was a golden-ratio
// sequence, which is LOW-DISCREPANCY -- deliberately as even as a set of points
// can be. It spread them, and it made a ruler: gap sd/mean 0.31 at nineteen
// stars where a real random sky measures about 1.00.
//
// A pure hash is not the answer either, and the numbers say so: same counts,
// hashed positions, sd/mean 1.00 -- and five pairs closer than two cells at
// nineteen stars. Two stars in adjacent cells read as ONE. That is a count
// error, and the count is the whole channel.
//
// So: dart-throwing, which is Poisson-disk sampling done the simple way. For
// slot i, draw positions from the hash and take the FIRST that is at least
// starMinSep from everything already placed; if the whole budget of darts is
// spent without one, take the roomiest dart thrown. The distribution is as
// random as a sky, because every accepted dart IS a uniform draw -- it is only
// the collisions that are thrown away.
//
// Mitchell's best-candidate was built and measured first and is REJECTED, with
// the render to show why. Best-candidate maximises the nearest-neighbour
// distance, and this band is 119 columns wide by 16 screen units tall, so the
// arrangement that maximises it is ROWS: at twelve stars it put ten of them on
// two lines at the top and bottom of the band, which is a tidier version of the
// ruler the golden ratio drew. Optimal spacing is not what a sky looks like.
//
// Both share the property that is locked in CLAUDE.md: slot i depends only on
// slots below it, so a star still lights where it always was and never moves as
// later ones arrive.
//
// Two dimensions, not one: the old layout spread x and hashed y independently,
// which is what pinned the sky into a band. Distance is measured in screen
// units (see starCellAspect), so a pair a row apart is not mistaken for a pair
// that is genuinely far apart.
//
// The result is cached: the layout depends only on the geometry, the band, the
// seed and the moon's column, none of which change between frames.
func (s *Shore) starPlaces(w, hy, top, bot, n int) [][2]int {
	if n < constellationSlots {
		n = constellationSlots
	}
	if n > 256 {
		n = 256
	}
	key := starKey{w: w, hy: hy, top: top, bot: bot, n: n, moonX: s.moonX, moonR: int(math.Round(s.moonRR * 64)), rx: s.moonRX, ry: s.moonRY}
	if s.starKey == key && s.starPts != nil {
		return s.starPts
	}

	cells := s.starCells(w, hy, top, bot)
	if len(cells) == 0 {
		s.starKey, s.starPts = key, nil
		return nil
	}

	pts := make([][2]int, 0, n)
	// The room there is, across: one column is one screen unit.
	width := float64(w)
	for i := 0; i < n; i++ {
		// How far this star has to be from everything already placed. The first
		// few have to be as far apart as the band allows, so they ask for more
		// than any sky can give and end up taking the roomiest dart thrown --
		// which is the farthest corner. After that the ask drops to starMinSep
		// and the first dart that clears it wins, which is a uniform draw.
		//
		// The opening is not a flourish, it is the s27 guarantee kept:
		// TestTheFirstStarsAreSpreadAcrossTheSky exists because his first-ever
		// lit constellation was two specks in the far-left corner, and a layout
		// that is random from the very first star will sometimes put the second
		// one beside it. Two stars ARE the picture at that moment. Thirty are a
		// scatter and can afford to be a scatter -- and measured, they have to
		// be: holding the spread rule past the first few drags the column-gap
		// spread from 1.26 back down to 0.31, which is the ruler again.
		want := starMinSep
		if i < starSpreadN {
			want = width + 1 // unreachable on purpose: take the roomiest dart
		}
		want *= want
		bestX, bestY, bestD := 0, 0, -1.0
		for j := 0; j < starDarts; j++ {
			cell := cells[int(HashF(i*1021+j, 101, s.Seed+31)*float64(len(cells)))%len(cells)]
			x, y := cell[0], cell[1]
			d := math.MaxFloat64
			for _, p := range pts {
				dx := float64(x - p[0])
				dy := float64(y-p[1]) * starCellAspect
				if v := dx*dx + dy*dy; v < d {
					d = v
				}
			}
			if d > bestD {
				bestX, bestY, bestD = x, y, d
			}
			if d >= want {
				break
			}
		}
		if bestD < starHardSep*starHardSep {
			// The darts all landed on top of something. That happens when the
			// sky is genuinely too small for the cap -- at 40x12 the band is
			// three rows of about thirty usable columns and thirty-two stars
			// will not fit three apart however they are arranged -- so stop
			// sampling and take the best cell there IS. Deterministic, and it
			// depends on the same thing every other branch depends on: the
			// stars already placed.
			bestX, bestY, bestD = s.roomiestCell(cells, pts)
		}
		pts = append(pts, [2]int{bestX, bestY})
	}
	s.starKey, s.starPts = key, pts
	return pts
}

// roomiestCell is the exhaustive answer to "where is there most room": the cell
// in the band whose nearest placed star is furthest away. It is the fallback
// for a sky too small to hold the whole cap at arm's length, and it runs only
// when the darts have all failed -- and then only once per window size, since
// the layout is cached.
func (s *Shore) roomiestCell(cells, pts [][2]int) (int, int, float64) {
	bestX, bestY, bestD := cells[0][0], cells[0][1], -1.0
	for _, cell := range cells {
		d := math.MaxFloat64
		for _, p := range pts {
			dx := float64(cell[0] - p[0])
			dy := float64(cell[1]-p[1]) * starCellAspect
			if v := dx*dx + dy*dy; v < d {
				d = v
			}
		}
		if d > bestD {
			bestX, bestY, bestD = cell[0], cell[1], d
		}
	}
	return bestX, bestY, bestD
}

// starCells is every cell of the band the constellation may use: the band less
// a margin, less the moon's own corridor, less the ground the context readout
// can land on.
//
// THE MOON'S CORRIDOR is the fix for a defect he reported on 2026-09-10 -- "I
// can see some of the constellation stars appearing and disappearing - once
// they appear, they should not disappear IMO". The disc sinks as the context
// fills (discGeom: 0.22*hy to 0.84*hy) and todoStars has always skipped any
// cell it covers, so a star the moon drifted over went OUT. Measured before
// this change, over a full context sweep: one star at 124x22 and one at
// 153x51. Opening the band to three quarters of the sky would have made that
// worse, since the disc's whole descent is now inside it.
//
// Keeping off the moon's COLUMNS instead of its current cells fixes it for
// good, and it costs nine columns of a hundred and twenty-five. The exclusion
// depends only on the geometry -- moonX and moonRR are fixed by the width, the
// height and MoonX, never by the context -- so the positions stay put as the
// disc moves through them, which is the whole point.
//
// ⚠ THE READOUT'S GROUND is the same defect by another door, and it is one the
// deeper band CREATED: measured, the layout that shipped never collided with it
// (0 of 6072 readout cells over the sweep) and the first deep one did, at 50 --
// four of twenty-four window-and-seed pairs lost a star for a stretch of the
// context range. drawReadout plots into the same near layer AFTER the scape, so
// a star under the number is simply deleted, and unlike the moon it does not
// move on: from 40% used it is there for the rest of the session.
//
// ⚠ The arithmetic below MIRRORS drawReadout in live.go, which is a coupling
// and will rot if that moves. TestTheReadoutNeverCoversAStar holds the two ends
// together: it reproduces drawReadout's own placement and goes red if this
// stops covering it.
func (s *Shore) starCells(w, hy, top, bot int) [][2]int {
	lo, hi := 3, w-3
	if hi <= lo || bot < top {
		return nil
	}
	// The widest column offset DiscCovers can return true for, at any row:
	// it tests hypot((x-moonX)/2, dy) < moonRR, so |dx| < 2*moonRR.
	colR := int(math.Ceil(2*s.moonRR)) - 1
	if colR < 0 {
		colR = 0
	}
	blocked := s.readoutGround(w, hy, top, bot)
	cells := make([][2]int, 0, (hi-lo)*(bot-top+1))
	for y := top; y <= bot; y++ {
		for x := lo; x < hi; x++ {
			if s.moonRR > 0 && x >= s.moonX-colR && x <= s.moonX+colR {
				continue
			}
			if blocked[[2]int{x, y}] {
				continue
			}
			cells = append(cells, [2]int{x, y})
		}
	}
	if len(cells) == 0 {
		// A scape too narrow to have a sky beside the moon. Give the stars the
		// width back rather than drawing nothing; the draw-time guard in
		// todoStars still keeps them off the disc itself.
		for y := top; y <= bot; y++ {
			for x := lo; x < hi; x++ {
				cells = append(cells, [2]int{x, y})
			}
		}
	}
	return cells
}

// readoutLabel is the longest the context readout can be: "100% left". It is
// never actually that long -- the readout shows what is LEFT and only appears
// from 40% used -- but the layout is cheaper to reason about at the maximum
// than at every string the number can make.
const readoutLabel = 9

// readoutGround is every cell the context readout can ever be painted on at
// this geometry, over the whole descent of the disc. See starCells for why it
// is excluded and for the warning about the coupling.
func (s *Shore) readoutGround(w, hy, top, bot int) map[[2]int]bool {
	out := map[[2]int]bool{}
	rx, ry := s.moonRX, s.moonRY
	if rx == 0 && ry == 0 {
		return out
	}
	mark := func(x0, y int) {
		if y < top || y > bot {
			return
		}
		for i := 0; i <= readoutLabel; i++ {
			out[[2]int{x0 + i, y}] = true
		}
	}
	for step := 0; step <= 128; step++ {
		// discGeom's own altitude, swept over every context the session can
		// reach. int() rather than rounding, because that is what it does.
		my := int(float64(hy) * (0.22 + 0.62*float64(step)/128))
		if my < 1 {
			my = 1
		}
		if y := my + ry + 1; y <= hy {
			// Under the disc, centred on it. The label is at most
			// readoutLabel wide and drawReadout centres it, so it reaches
			// half that either side.
			mark(s.moonX-readoutLabel/2, y)
			continue
		}
		// Beside it, on the disc's own row. drawReadout prefers the right and
		// falls back to the left; which one it takes depends on the label's
		// real length, so both are held.
		mark(s.moonX+rx+2, my)
		mark(s.moonX-rx-1-readoutLabel, my)
	}
	return out
}

// starInk is the tone and the alpha for one lit star, chosen against the
// ground it is ACTUALLY painted on rather than against the palette's idea of a
// night sky.
//
// This is the sand's rule, applied to the sky: "Ink is sampled from the PAINTED
// background per row, never the palette's nominal sand." The reason is the same
// and it is what buys the deeper band. Measured at the 40-luma bar the tests
// hold: the palette's star tone at alpha 0.85 stops reading at 0.36 of the sky
// at the worst hour of the day, which is SHALLOWER than the band that shipped;
// lifted toward white it reaches 0.64; lifted and fully opaque, 0.82.
//
// Two things it must not do, both of which the first version did and both of
// which the rendered frame caught:
//
//   - Reason about c.BGAt. That is the NOMINAL background, and on 256 a sky
//     cell takes its tone from the ramp's own path instead. At 40x12 the
//     nominal ground is luma 159 and the ground the eye sees is 188, so an ink
//     chosen for 55 points of contrast delivered 19.
//   - Reason about the ink it passes in. The glyph path saturates chroma by
//     GlyphBoost before quantising, which moves the luma as well -- the
//     palette's star tone comes out twelve luma DARKER than it went in.
//
// So it walks the lift from the palette tone to white, and reads the contrast
// the same way the terminal will: quantised, glyph path, against the ground
// ResolveAt reports. The first rung that clears the bar wins, so at night
// nothing changes at all.
//
// mag is the star's own magnitude, 0 to 1, fixed by its slot. It is added to
// the floor, so a faint star is exactly as legible as the tests require and a
// bright one has more.
func (s *Shore) starInk(c *canvas.Canvas, x, y int, mag float64) (term.RGB, float64) {
	nominal := c.BGAt(x, y)
	// The cell has no glyph yet, so on 256 this reports it as its two halves:
	// the LOWER one, which in a sky that brightens downward is the brighter of
	// the two and so the harder to beat. Being wrong in that direction costs a
	// shade of white and nothing else.
	_, _, ground := c.ResolveAt(x, y, term.Profile256)
	white := term.RGB{R: 255, G: 255, B: 255}
	bestInk, bestA, best := s.pal.Star, mag, -math.MaxFloat64
	// The star's own magnitude first, then fully opaque: a bright night star
	// keeps the tone its magnitude gives it, and only a sky that beats it makes
	// it give that up.
	for _, a := range [2]float64{mag, 1.0} {
		for step := 0; step <= starInkSteps; step++ {
			ink := term.Lerp(s.pal.Star, white, float64(step)/starInkSteps)
			seen := term.Profile256.Quantise(nominal.Blend(ink, a), true)
			if d := bandLuma(seen) - bandLuma(ground); d >= todoStarContrast {
				return ink, a
			} else if d > best {
				bestInk, bestA, best = ink, a, d
			}
		}
	}
	// Nothing reaches: the sky there is brighter than white ink can beat. Take
	// the best there is rather than the dimmest, and let the band's own floor
	// (starBand) keep the layout out of rows where this happens.
	return bestInk, bestA
}

// starInkSteps is how finely the lift from the palette's star tone to white is
// walked. Nine rungs, because the cube has six levels a channel and a finer
// walk returns the same colours.
const starInkSteps = 8

// todoStars lights one star per finished todo, so the sky carries n.
//
// Position, not rate: each slot's place is fixed by its index and the scene
// seed, so a star lights where it always was rather than appearing somewhere
// new. A star that moved would encode nothing and read as noise, and a glance
// is the whole budget. See starPlaces for how the places are chosen and why
// they are neither a ruler nor a pure hash.
//
// Deliberately NOT a row across the top. A progress bar in the sky is a piece
// of UI and the brief cuts anything that makes this feel like a dashboard; a
// constellation filling in does the same work and belongs to the picture.
func (s *Shore) todoStars(c *canvas.Canvas, hy, foamTop, done, total int) {
	if total <= 0 || hy < 3 {
		return
	}
	if done > total {
		done = total
	}
	top, bot := starBand(hy, foamTop)
	if bot < top {
		return // the water is up in the sky: no band to put a constellation in
	}
	pts := s.starPlaces(c.W, hy, top, bot, total)
	near := c.Near()
	for i := 0; i < done && i < len(pts); i++ {
		x, y := pts[i][0], pts[i][1]
		if x < 0 || x >= c.W || y < 0 || y >= c.H {
			continue
		}
		// Never on the moon: it carries context remaining, and a star inside
		// the disc would be read as part of it. starCells already keeps the
		// whole layout out of the disc's columns; this is the guard that was
		// here before it and it stays, because DiscCovers is moon()'s own
		// sampling and a layout that ever drifts back into the disc's reach
		// should be caught by the shape itself rather than by arithmetic about
		// it. (It was a hand-fitted ellipse once, measured against a disc of
		// one particular size, and it drifted.)
		if s.DiscCovers(x, y) {
			continue
		}
		// Only finished todos are drawn. An unfinished one used to leave a
		// ring, so the sky read "n of N"; his ruling of 2026-09-05 -- "discard
		// the ring altogether, it's not clear what it means" -- so the sky
		// says n, and the size of the list is not on screen.
		ink, a := s.starInk(c, x, y, starMagnitude(i, s.Seed))
		near.Plot(x, y, '*', ink, a)
	}
}

// discGeom fixes the disc's centre and radius for the frame. It has to run
// BEFORE the star fields, and that is the whole reason it exists apart from
// moon(): moon() paints only BACKGROUNDS, so a glyph already plotted under the
// disc survives and composites against the body instead of the sky. That is
// his report of 2026-09-06 from a live 143x62 window -- one tan speck on the
// sun's face, an ambient dust glyph in the centre column.
func (s *Shore) discGeom(c *canvas.Canvas, hy int, scale, lit float64) {
	frac := s.MoonX
	if frac <= 0 {
		frac = 0.72
	}
	my := int(float64(hy) * (0.22 + 0.62*(1-lit)))
	if my < 1 {
		my = 1
	}
	s.moonX, s.moonY = int(float64(c.W)*frac), my
	s.moonRR = 2.0 * scale
	s.moonPainted = s.MoonEdge != "none"
}

// DiscCovers says the disc reaches this cell, by the SAME half-row sampling
// moon() paints with: a cell counts when either half-row is inside. Testing
// whole rows instead misses the ring of cells the disc touches with only one
// of them -- and misses none at all at 143x27, which is the trap, because a
// guard measured against his screenshot alone would look complete and still
// leak a row taller.
//
// The test is the disc's FOOTPRINT, not its lit face: by day the terminator is
// not painted, so a star inside the crescent's dark side would technically be
// over sky. Excluding it anyway keeps the field still. Tying it to the phase
// would make specks blink in and out along the terminator as the context
// filled, which is a worse thing to see than a slightly emptier patch of sky.
func (s *Shore) DiscCovers(x, y int) bool {
	if !s.moonPainted || s.moonRR <= 0 {
		return false
	}
	fx := float64(x-s.moonX) / 2.0
	for _, off := range [2]float64{-0.25, 0.25} {
		if math.Hypot(fx, float64(y-s.moonY)+off) < s.moonRR {
			return true
		}
	}
	return false
}

func (s *Shore) moon(c *canvas.Canvas, hy int, scale, lit, vis float64) {
	mx := s.moonX
	// Altitude carries context too, alongside phase. Shape alone is hard to
	// judge on a five-cell disc; height above the horizon is easy, because the
	// horizon is a reference line right there. Two cues for one variable is
	// what makes it readable at a glance rather than on inspection.
	// (Centre and radius are fixed in discGeom, which runs before the stars.)
	my := s.moonY
	rr := s.moonRR
	rim := 0.6 * scale
	// The terminator is approximated by a second disc sliding across the face:
	// fully clear of it at full, concentric at new.
	shadow := 2 * rr * lit
	// The unlit face. Dim, but it has to READ as part of the disc: at
	// (62,62,74) and 0.45 it blended into a night sky within one grey step,
	// so a moon one column into its phase looked bitten on the right rather
	// than shaded -- his second report, "the moon not looking so good", at
	// 133x27 and 9% context. Three steps above the night sky now, still well
	// under the lit body, still darker than a daylight sky.
	dark := term.RGB{R: 80, G: 80, B: 96}
	// By day the body is the SUN, and a sun has no unlit face. Painting the
	// moon's terminator on it made a slate mass that grew as the context
	// filled and cleared at a compaction -- his report of 2026-09-06 from a
	// long session, "the Sun seems to break sometimes, and fix itself".
	// Measured in that frame at 49%: the lit body ended at x677 and the slate
	// ran on to x719, 2.6 cells past it. His ruling: the sun WANES AS A
	// CRESCENT. At night the moon keeps its shaded face, which is what stops
	// a moon one column into its phase reading as bitten rather than shaded.
	noShadow := !s.night() && s.SunShadow != "slate"

	ry := int(rr+rim) + 1
	rx := int((rr+rim)*2) + 1
	s.moonRX, s.moonRY = rx, ry
	if s.MoonEdge == "none" {
		return // a control frame for the tests: everything but the disc
	}
	if s.MoonEdge == "quad" {
		s.moonQuad(c, mx, my, rx, ry, rr, shadow, vis, dark, noShadow)
		if s.MoonHalo {
			s.moonHalo(c, hy, mx, my, rx, ry, rr, vis)
		}
		return
	}
	// Each cell is judged as TWO half-rows, a quarter above and a quarter
	// below its centre, and painted with U+2580 where only one of them is
	// inside the disc (or where the terminator crosses between them). His
	// 124x52 window gives the scape 23 rows, a radius of 1.92, and at whole
	// rows that disc has no tips at all: three rows, all seven wide, a
	// rectangle. At half rows the tips come back as three half-cells. The
	// disc still ends at rr; only the sampling got finer.
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			fx := float64(dx) / 2.0
			x, y := mx+dx, my+dy
			var half [2]term.RGB
			var in [2]bool
			for k, off := range [2]float64{-0.25, 0.25} {
				fy := float64(dy) + off
				// Strictly inside. At a radius of exactly 2.25 rows the top
				// half of the centre column sits ON the edge, and a tie that
				// counts as in makes that one cell full while its neighbours
				// are halves: a one-column pip at twelve and six o'clock.
				if math.Hypot(fx, fy) >= rr {
					continue
				}
				// The disc ends at rr, not at rr+rim. rim used to be the width of
				// a FADE from rr-rim out to rr+rim; making the disc solid without
				// moving the cutoff painted every one of those cells at full
				// strength and the moon came out nearly twice the radius it
				// should be -- a pink blob with a ragged edge. rim now only says
				// how much of the outside was ever going to be soft.
				//
				// And the disc IS solid: a fading rim has nowhere to land on 256.
				// Blending the sun into the sky passes through the alpha where
				// its red and green cross, and the colours either side of that
				// carry chroma in the twenties and thirties, which is where the
				// greyscale ramp wins -- measured on his frame, a warm disc with
				// a grey fringe, which reads as a broken sprite. So the edge is
				// clean, and the half-row sampling is what rounds it.
				a := 0.92 * vis
				if a <= 0 {
					continue
				}
				col := s.pal.Moon
				lit := true
				if math.Hypot(fx-shadow, fy) <= rr {
					if noShadow {
						continue
					}
					col, a, lit = dark, a*0.6, false // earthshine on the unlit face
				}
				rim := lit && s.MoonRim != "" && math.Hypot(fx, fy) > rr-0.55
				if rim && s.MoonRim == "blend" {
					a *= 0.5
				}
				in[k] = true
				half[k] = c.BGAt(x, y).Blend(col, a)
				if rim && s.MoonRim == "hue" {
					// One tone darker than the FINISHED body, in its hue: the
					// body as 256 shows it, scaled, and quantised with the
					// ordering kept. Darkening the moon colour before the
					// blend was tried first and the blend took the chroma
					// with it -- an olive and two greys by day.
					q := term.FromIndex256(half[k].Index256Keeping())
					half[k] = term.RGB{R: uint8(float64(q.R) * 0.72), G: uint8(float64(q.G) * 0.72), B: uint8(float64(q.B) * 0.72)}
				}
			}
			switch {
			case in[0] && in[1] && half[0] == half[1]:
				c.SetBG(x, y, half[0])
			case in[0] && in[1]:
				// BOTH halves inside the disc: paint ONE tone.
				//
				// A half block on Terminal.app leaves a one-pixel rule at the
				// cell's bottom edge -- its ink runs to 29.4 of a 30px row, and
				// the last device pixel row falls back to the cell's
				// background, which is the OTHER half's colour. Measured on his
				// own screen 2026-09-07: at the disc's centre column, twelve
				// pixels of (209,166,124), then a single pixel of (185,142,101),
				// then sixty more of (209,166,124). Inside the disc that rule
				// runs clean across the face, which is what he has reported
				// three times as "thin hairlines breaking up the art", most
				// visibly on the sun and the moon.
				//
				// The SILHOUETTE is what makes the disc round, and the
				// silhouette is drawn by the two edge cases below -- they still
				// split, so nothing about the roundness measured over 184
				// (width, rows) pairs changes. Collapsing only the INTERIOR
				// costs the internal shading half a cell of vertical resolution
				// and takes every rule off the face.
				c.SetBG(x, y, term.Lerp(half[0], half[1], 0.5))
			case in[0]:
				// The disc's LOWER edge: only the upper half is inside.
				if s.flatEdge(dx, dy, 0, rr, shadow, noShadow) {
					c.SetBG(x, y, half[0])
				} else {
					c.SetBGHalves(x, y, half[0], c.BGAt(x, y))
				}
			case in[1]:
				// The disc's UPPER edge: only the lower half is inside.
				if s.flatEdge(dx, dy, 1, rr, shadow, noShadow) {
					c.SetBG(x, y, half[1])
				} else {
					c.SetBGHalves(x, y, c.BGAt(x, y), half[1])
				}
			}
		}
	}
}

// discHalf says whether one half-row of one cell is painted as disc. It is
// moon()'s own test, lifted out so a cell can ask about its NEIGHBOURS without
// doing their colour work. off is -0.25 for the upper half, +0.25 for the
// lower, matching the sampling moon() paints by.
func (s *Shore) discHalf(dx, dy int, off, rr, shadow float64, noShadow bool) bool {
	fx, fy := float64(dx)/2.0, float64(dy)+off
	if math.Hypot(fx, fy) >= rr {
		return false
	}
	// By day the sun wanes as a crescent and the unlit half is not painted at
	// all, so a half-row inside the circle can still be outside the disc.
	if noShadow && math.Hypot(fx-shadow, fy) <= rr {
		return false
	}
	return true
}

// flatEdge says an edge cell's split buys no roundness: a neighbouring column
// has its edge in the SAME cell and on the same side, so the silhouette runs
// horizontally through both and the two of them draw one straight rule.
// in is the half that is inside -- 0 upper, 1 lower.
//
// Only the neighbours are asked, and only for the same half. A cell whose
// neighbours split at a different height is on the curve, and there the split
// is the whole reason the disc is round rather than a rectangle; those keep it.
func (s *Shore) flatEdge(dx, dy, in int, rr, shadow float64, noShadow bool) bool {
	if !s.FlatCaps || !term.NoSplitCells {
		return false
	}
	for _, n := range [2]int{dx - 1, dx + 1} {
		up := s.discHalf(n, dy, -0.25, rr, shadow, noShadow)
		down := s.discHalf(n, dy, 0.25, rr, shadow, noShadow)
		if up == down {
			continue // that column is not an edge cell at this row
		}
		if (in == 0 && up) || (in == 1 && down) {
			return true
		}
	}
	return false
}

// moonQuad samples the disc at four quarters per cell (a quarter of a
// cell's width, a quarter of its height, about its centre) and paints each
// cell with canvas.SetBGQuad. A cell the terminator crosses inside the disc
// carries the lit and the unlit colour as its two quarters' colours; a cell
// on the edge carries the majority face against the sky.
func (s *Shore) moonQuad(c *canvas.Canvas, mx, my, rx, ry int, rr, shadow, vis float64, dark term.RGB, noShadow bool) {
	a := 0.92 * vis
	if a <= 0 {
		return
	}
	// One tone each for the whole disc, against the sky at its centre. Blended
	// row by row, the crescent's lower cells rounded to a different yellow
	// from its upper ones on 256 -- "broken/messy", his study note.
	centre := c.BGAt(mx, my)
	litCol := centre.Blend(s.pal.Moon, a)
	darkCol := centre.Blend(dark, a*0.6)
	s.litTone, s.darkTone = litCol, darkCol

	// Pass one: sample every cell's four quarters and decide its two colours.
	type qcell struct {
		x, y              int
		litMask, darkMask uint8
		dark              bool // the shape colour is the dark face
	}
	var cells []qcell
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			var q qcell
			q.x, q.y = mx+dx, my+dy
			for k, off := range [4][2]float64{{-0.125, -0.25}, {0.125, -0.25}, {-0.125, 0.25}, {0.125, 0.25}} {
				fx, fy := float64(dx)/2.0+off[0], float64(dy)+off[1]
				if math.Hypot(fx, fy) >= rr {
					continue
				}
				bit := uint8(8) >> uint(k)
				if math.Hypot(fx-shadow, fy) <= rr {
					if !noShadow {
						q.darkMask |= bit
					}
					continue
				}
				q.litMask |= bit
			}
			if q.litMask|q.darkMask == 0 {
				continue
			}
			// An edge cell can carry two colours, and the edge against the
			// sky is the one that has to stay clean, so the terminator inside
			// it goes to whichever face holds more of the cell, ties to the
			// dark face: a thin sliver at 30% reads as one dark rim along the
			// edge rather than as loose blocks.
			q.dark = q.litMask == 0 || (q.litMask|q.darkMask != 0b1111 && bits(q.darkMask) >= bits(q.litMask))
			cells = append(cells, q)
		}
	}
	// Pass two: the dark face is ONE piece. At 60% the crescent's horn can
	// leave a single dark cell adrift at its tip where the lens is thinner
	// than a cell; any dark cell not connected to the largest dark group goes
	// to the lit face. His note on the study: "should be a clean container".
	xs, ys, darks := make([]int, len(cells)), make([]int, len(cells)), make([]bool, len(cells))
	for i, q := range cells {
		xs[i], ys[i], darks[i] = q.x, q.y, q.dark
	}
	darkPieces(xs, ys, darks)
	s.quadDark = make(map[[2]int]bool, len(cells))
	for i := range cells {
		cells[i].dark = darks[i]
		full := cells[i].litMask|cells[i].darkMask == 0b1111
		s.quadDark[[2]int{cells[i].x, cells[i].y}] = (full && cells[i].litMask == 0) || (!full && darks[i])
	}

	for _, q := range cells {
		sky := c.BGAt(q.x, q.y)
		switch {
		case q.litMask|q.darkMask == 0b1111 && q.darkMask == 0:
			c.SetBG(q.x, q.y, litCol)
		case q.litMask|q.darkMask == 0b1111 && q.litMask == 0:
			c.SetBG(q.x, q.y, darkCol)
		case q.litMask|q.darkMask == 0b1111:
			c.SetBGQuad(q.x, q.y, litCol, darkCol, q.litMask)
		default:
			shape := litCol
			if q.dark {
				shape = darkCol
			}
			c.SetBGQuad(q.x, q.y, shape, sky, q.litMask|q.darkMask)
		}
	}
}

// darkPieces groups the dark-shaped cells 8-connected and flips every group
// but the largest to the lit face, so the dark face is one piece.
func darkPieces(xs, ys []int, dark []bool) {
	n := len(xs)
	idx := map[[2]int]int{}
	for i := 0; i < n; i++ {
		if dark[i] {
			idx[[2]int{xs[i], ys[i]}] = i
		}
	}
	seen := make([]bool, n)
	var groups [][]int
	for i := 0; i < n; i++ {
		if !dark[i] || seen[i] {
			continue
		}
		var g []int
		stack := []int{i}
		seen[i] = true
		for len(stack) > 0 {
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			g = append(g, j)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if k, ok := idx[[2]int{xs[j] + dx, ys[j] + dy}]; ok && !seen[k] {
						seen[k] = true
						stack = append(stack, k)
					}
				}
			}
		}
		groups = append(groups, g)
	}
	best := -1
	for i, g := range groups {
		if best < 0 || len(g) > len(groups[best]) {
			best = i
		}
	}
	for i, g := range groups {
		if i == best {
			continue
		}
		for _, j := range g {
			dark[j] = false
		}
	}
}

// night is whether the sky is dark enough for the body to be the moon: no
// channel of the zenith reaches 80. Luma cannot say -- the midday zenith is
// a saturated blue (0,95,175) whose luma is 76 -- and neither can the body's
// own colour, a pinkish grey at 22:20 that reads as warm.
func (s *Shore) night() bool {
	t := s.pal.SkyTop
	return t.R < 80 && t.G < 80 && t.B < 80
}

func bits(m uint8) int {
	n := 0
	for ; m != 0; m >>= 1 {
		n += int(m & 1)
	}
	return n
}

// moonHalo lightens the sky in a soft ring around the disc, at night only:
// the grey ramp has the steps for it, a daylight sky in the cube does not.
// Cells the disc painted are left alone.
func (s *Shore) moonHalo(c *canvas.Canvas, hy, mx, my, rx, ry int, rr, vis float64) {
	if !s.night() {
		return
	}
	const reach = 2.5
	for dy := -ry - 3; dy <= ry+3; dy++ {
		for dx := -rx - 5; dx <= rx+5; dx++ {
			x, y := mx+dx, my+dy
			if y >= hy {
				continue // the sky only: on the water it rounded to pink
			}
			d := math.Hypot(float64(dx)/2.0, float64(dy))
			if d < rr+0.5 || d > rr+reach {
				continue
			}
			t := 1 - (d-rr-0.5)/(reach-0.5)
			lift := 0.28 * t * t * vis
			if lift < 0.03 {
				continue
			}
			c.SetBG(x, y, c.BGAt(x, y).Blend(s.pal.Moon, lift))
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
			if Tide {
				// Toward the shore, not across the frame -- his report on the
				// tide build, "it still moves to the sides".
				//
				// A plane wave travels PERPENDICULAR to its crests, so as long
				// as x sits inside the travelling phase the pattern must slide
				// sideways: with x at 0.30 and y at 0.9 the old locus moved
				// (-0.32, -0.95) per step, up the frame and to the left, which
				// is backwards twice over. Flipping the sign would only send it
				// down and to the RIGHT.
				//
				// So x comes out of the travelling term and becomes a spatial
				// DISTORTION of a wave that travels in y alone: the crests
				// wiggle, and they come in.
				ph = (float64(y)+tt*(0.55+depth*1.65))*0.9 + 1.5*math.Sin(float64(x)*0.16)
			}
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
	strength := (0.35 + 0.65*lit) * moonVis(s.pal)
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
	bottom := c.H
	if s.writeTop > 0 && s.writeTop < bottom {
		bottom = s.writeTop
	}
	for x := 0; x < c.W; x++ {
		for y := int(edge[x]) + 2; y < bottom; y++ {
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
			if s.writeTop > 0 && yy >= s.writeTop {
				continue
			}
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

// moonVisFloor keeps the moon legible at every hour.
//
// The moon carries context remaining, which is a fact about the agent, and its
// visibility was bound to the wall clock, which is a fact about the world. That
// breaks the rule the whole scene is built on -- the water is the work, the sky
// is the world, and nothing crosses. In practice it meant the context readout
// was invisible for the entire working day: measured against its own sky, the
// moon sat +198 luma at midnight, +14 at half past eleven and +10 at noon.
//
// The hour still sets the moon's colour and the reach of its shine on the water.
// It no longer decides whether you can see it at all.
const moonVisFloor = 0.55

func moonVis(p Palette) float64 {
	if p.MoonVis < moonVisFloor {
		return moonVisFloor
	}
	return p.MoonVis
}
