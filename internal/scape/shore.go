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

func NewShore(seed int64, asciiOnly bool) *Shore {
	return &Shore{Seed: seed, ASCII: asciiOnly, SandFade: DefaultSandFade, WriteRows: DefaultWriteRows}
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
	sy := c.H - beach
	if sy <= hy+1 {
		sy = hy + 2
	}
	if sy > writeTop-1 {
		sy = writeTop - 1
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
	// Scale the swell to the room it has instead of clipping it.
	//
	// Clipping is what drew the ruled shoreline: every trough that reached past
	// the writing band landed on the same row. Scaling keeps the shape -- the
	// crests and hollows stay where they are, just shallower -- so the water
	// still runs up the sand and back, and never touches the writing.
	if writeTop < c.H {
		room := float64(writeTop-1) - float64(sy)
		if room < 0.5 {
			room = 0.5
		}
		dev := 0.0
		for _, e := range edge {
			if d := math.Abs(e - float64(sy)); d > dev {
				dev = d
			}
		}
		if dev > room {
			k := room / dev
			for i := range edge {
				edge[i] = float64(sy) + (edge[i]-float64(sy))*k
			}
		}
	}
	s.lastEdge = edge
	s.writeTop = writeTop

	s.paintBG(c, hy, edge)
	s.stars(c, hy, t)
	s.moon(c, hy, scale, 1-clamp01(act.ContextUsed), moonVis(s.pal))
	s.todoStars(c, hy, act.TodoDone, act.TodoTotal)
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
				col = term.Lerp(s.pal.WetSand, s.pal.SeaNear, ex-fy+0.5)
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

func (s *Shore) stars(c *canvas.Canvas, hy int, t float64) {
	far := c.Far()
	// No '*' here. It belongs to the checklist, and a channel that shares a
	// glyph with the scenery is not a channel: at midnight the ambient field
	// drew stars indistinguishable from finished todos and the count could not
	// be read off the sky at all.
	glyphs := []rune{'.', '.', '\u00b7', '+'}
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
// todoStarFloor keeps the checklist legible at every hour.
//
// The same rule as the moon and needed for the same reason: a completed todo is
// a fact about the AGENT, and the ambient star field's visibility is a fact
// about the WORLD. StarVis is 0 at noon. Hanging the checklist on it would
// switch the readout off for the whole working day -- which is exactly the bug
// moonVisFloor exists to fix, arriving a second time through a different door.
const todoStarFloor = 0.85

// todoStars lights one star per finished todo and leaves a faint mark where one
// is still outstanding, so the sky carries "n of N" rather than just n.
//
// Position, not rate: each slot's place is fixed by its index and the scene
// seed, so a star lights where it always was rather than appearing somewhere
// new. A star that moved would encode nothing and read as noise, and a glance
// is the whole budget.
//
// Deliberately NOT a row across the top. A progress bar in the sky is a piece
// of UI and the brief cuts anything that makes this feel like a dashboard; a
// constellation filling in does the same work and belongs to the picture.
func (s *Shore) todoStars(c *canvas.Canvas, hy, done, total int) {
	if total <= 0 || hy < 3 {
		return
	}
	if done > total {
		done = total
	}
	// The upper sky is the darkest part of it at every hour, and so the only
	// part where a white mark reads by day as well as by night.
	top, bot := 1, hy*3/5
	if bot <= top {
		bot = top + 1
	}
	near := c.Near()
	for i := 0; i < total; i++ {
		// Spread across the width by index so the constellation grows outward
		// rather than piling up, then jittered so it is not a ruled line.
		frac := (float64(i) + 0.5) / float64(total)
		x := int(frac*float64(c.W-6)) + 3
		x += int(HashF(i, 11, s.Seed+31)*3) - 1
		y := top + int(HashF(i, 23, s.Seed+37)*float64(bot-top))
		if x < 0 || x >= c.W || y < 0 || y >= c.H {
			continue
		}
		// Never on the moon: it carries context remaining, and a star inside
		// the disc would be read as part of it. Cells are twice as tall as they
		// are wide, so the exclusion is an ellipse, not a circle.
		if dx, dy := x-s.moonX, y-s.moonY; dx*dx+4*dy*dy <= 16 {
			continue
		}
		if i < done {
			near.Plot(x, y, '*', s.pal.Star, todoStarFloor)
		} else {
			// A ring, not a dot: the ambient field is made of dots, and an
			// outstanding todo has to be legible as a DIFFERENT thing rather
			// than as a dim one.
			near.Plot(x, y, '\u2218', s.pal.Star, 0.30)
		}
	}
}

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
				if math.Hypot(fx, fy) > rr {
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
				if math.Hypot(fx-shadow, fy) <= rr {
					col, a = dark, a*0.45 // earthshine on the unlit face
				}
				in[k] = true
				half[k] = c.BGAt(x, y).Blend(col, a)
			}
			switch {
			case in[0] && in[1] && half[0] == half[1]:
				c.SetBG(x, y, half[0])
			case in[0] && in[1]:
				c.SetBGHalves(x, y, half[0], half[1])
			case in[0]:
				c.SetBGHalves(x, y, half[0], c.BGAt(x, y))
			case in[1]:
				c.SetBGHalves(x, y, c.BGAt(x, y), half[1])
			}
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
