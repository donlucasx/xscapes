// Package term handles colour: the RGB model the renderer composites in, and
// the translation down to whatever the host terminal can actually show.
package term

import (
	"os"
	"strconv"
	"strings"
)

type RGB struct{ R, G, B uint8 }

// Blend returns c with over painted on top at alpha a. This is the whole
// depth model: a far layer is drawn at low alpha so its colour falls back
// toward the background behind it, which reads as distance.
func (c RGB) Blend(over RGB, a float64) RGB {
	switch {
	case a <= 0:
		return c
	case a >= 1:
		return over
	}
	return RGB{
		R: uint8(float64(c.R)*(1-a) + float64(over.R)*a + 0.5),
		G: uint8(float64(c.G)*(1-a) + float64(over.G)*a + 0.5),
		B: uint8(float64(c.B)*(1-a) + float64(over.B)*a + 0.5),
	}
}

// Lerp ramps between two colours, for gradients.
func Lerp(a, b RGB, t float64) RGB {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return a.Blend(b, t)
}

type Profile int

const (
	Profile256 Profile = iota
	ProfileTrueColor
)

func (p Profile) String() string {
	if p == ProfileTrueColor {
		return "truecolor"
	}
	return "256"
}

// DetectProfile decides how many colours we may emit.
//
// Deliberately does NOT trust COLORTERM on its own. Measured 2026-08-29: inside
// a Claude Code session in Terminal.app the environment carries
// COLORTERM=truecolor while Terminal.app renders 256 at most, and that is
// exactly the environment asciiscapes launches into. Gate on the program.
func DetectProfile() Profile {
	switch strings.ToLower(os.Getenv("ASCIISCAPES_COLOR")) {
	case "truecolor", "24bit", "full":
		return ProfileTrueColor
	case "256", "ansi256":
		return Profile256
	}
	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" {
		return Profile256
	}
	switch strings.ToLower(os.Getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return ProfileTrueColor
	}
	return Profile256
}

// xterm-256 cube levels. Indices 0-15 are theme-dependent and unreliable, so
// we quantise into the 6x6x6 cube (16-231) and the greyscale ramp (232-255).
var cubeLevels = [6]uint8{0, 95, 135, 175, 215, 255}

func nearestCube(v uint8) (idx int, val uint8) {
	best, bestD := 0, 1<<30
	for i, l := range cubeLevels {
		d := int(v) - int(l)
		if d < 0 {
			d = -d
		}
		if d < bestD {
			bestD, best = d, i
		}
	}
	return best, cubeLevels[best]
}

func dist(a, b RGB) int {
	dr, dg, db := int(a.R)-int(b.R), int(a.G)-int(b.G), int(a.B)-int(b.B)
	// Rough luma weighting; good enough and far cheaper than a palette search.
	return 30*dr*dr + 59*dg*dg + 11*db*db
}

// Index256 maps an RGB to the closest xterm-256 index, in constant time.
func (c RGB) Index256() int {
	ri, rv := nearestCube(c.R)
	gi, gv := nearestCube(c.G)
	bi, bv := nearestCube(c.B)
	cubeIdx := 16 + 36*ri + 6*gi + bi
	cubeErr := dist(c, RGB{rv, gv, bv})

	// Greyscale ramp: 232..255 = 8, 18, ... 238.
	g := (int(c.R)*30 + int(c.G)*59 + int(c.B)*11) / 100
	gi2 := (g - 8 + 5) / 10
	if gi2 < 0 {
		gi2 = 0
	} else if gi2 > 23 {
		gi2 = 23
	}
	gv2 := uint8(8 + gi2*10)
	greyErr := dist(c, RGB{gv2, gv2, gv2})

	if greyErr < cubeErr {
		return 232 + gi2
	}
	return cubeIdx
}

// Index256Keeping is Index256 with one more requirement: where the colour has a
// clear ordering between two channels, the answer has to keep it.
//
// Independent per-channel rounding does not. rgb(48,112,170) is a mid blue with
// green sixty-four clear of red; the nearest cube colour rounds red UP to 95
// and green DOWN to 95, and hands back rgb(95,95,175), which is a violet. The
// sky was doing this for six rows a day around five in the afternoon, and no
// amount of choosing better endpoints fixes it -- the rounding happens between
// them.
//
// The cost of fixing it is small and worth naming: rgb(0,95,175) is 3% further
// from the source by distance than the violet is. Hue is worth more than 3% of
// distance in a sky, where a wrong hue is the only error anyone can see.
//
// Backgrounds only. Glyphs take the chroma boost first and are a different
// problem: there the point is to escape the grey ramp, not to preserve a
// relationship the boost has already changed.
func (c RGB) Index256Keeping() int {
	plain := c.Index256()
	// A GREY answer is left alone, and this is the whole safety of the thing.
	//
	// Grey has no channel ordering, so it fails ordersKept by definition -- and
	// without this line the search then goes looking for a cube colour that
	// does have one, in a region of the cube where the only candidates are the
	// pure #0000xx column. That is precisely the electric royal blue night
	// Shore.BlueSky records as rendered and rejected, and it is what this
	// function shipped as before the check was added: a midnight sky in
	// saturated navy with a teal sea, nothing like the frame it was quantising.
	//
	// Washing out and turning the wrong colour are different failures. This
	// function is only allowed to fix the second one.
	if plain >= 232 {
		return plain
	}

	ri, _ := nearestCube(c.R)
	gi, _ := nearestCube(c.G)
	bi, _ := nearestCube(c.B)
	best, bestScore := plain, 1<<62
	if c.ordersKept(FromIndex256(plain)) {
		bestScore = c.hueScore(FromIndex256(plain))
	}
	for _, dr := range [2]int{0, -1} {
		for _, dg := range [2]int{0, -1} {
			for _, db := range [2]int{0, -1} {
				r, g, b := clampIdx(ri+dr), clampIdx(gi+dg), clampIdx(bi+db)
				for _, rr := range [2]int{r, clampIdx(r + 1)} {
					for _, gg := range [2]int{g, clampIdx(g + 1)} {
						for _, bb := range [2]int{b, clampIdx(b + 1)} {
							cand := RGB{cubeLevels[rr], cubeLevels[gg], cubeLevels[bb]}
							if !c.ordersKept(cand) {
								continue
							}
							if sc := c.hueScore(cand); sc < bestScore {
								bestScore, best = sc, 16+36*rr+6*gg+bb
							}
						}
					}
				}
			}
		}
	}
	return best
}

// hueWeight is how much a wrong HUE costs against a wrong distance.
//
// Nearest-by-distance is not good enough on a ramp, and the failure is a stripe
// rather than a shade. Down a noon sky the greenness of the painted colour ran
// 95, 95, 135, 40, 40, 80, 40... -- it climbed into a cyan for one row and came
// straight back to blue, because red starts at 0 where the cube's first step is
// 95 wide while green starts at 95 where the steps are 40. Green advances a
// level before red does, and that one row is a cyan band across the whole sky.
//
// No band count can see it: the cyan band is exactly as distinct as the good
// ones. Weighing the hue instead makes the run monotonic, and it does it
// without giving up the deep zenith -- the alternative was lightening the sky
// until red starts at 95 too, which removed the wobble and more than half the
// bands with it, 9 down to 4.
//
// 8 is the smallest weight that fixes the measured case; the cost of picking
// rgb(0,95,175) over rgb(0,135,175) there is 26% more distance for 20 times
// less hue error.
var hueWeight = 8

// hueScore is distance with the two hue relationships weighted in.
func (c RGB) hueScore(q RGB) int {
	dgr := (int(c.G) - int(c.R)) - (int(q.G) - int(q.R))
	dbg := (int(c.B) - int(c.G)) - (int(q.B) - int(q.G))
	return dist(c, q) + hueWeight*(dgr*dgr+dbg*dbg)
}

func clampIdx(i int) int {
	if i < 0 {
		return 0
	}
	if i > 5 {
		return 5
	}
	return i
}

// ordersKept says whether q preserves every channel ordering c states clearly.
// Twenty is "clearly": below that the source itself is close to neutral in that
// pair and has no ordering worth defending.
func (c RGB) ordersKept(q RGB) bool {
	pair := func(a, b, qa, qb uint8) bool {
		switch d := int(a) - int(b); {
		case d >= 20:
			return qa > qb
		case d <= -20:
			return qa < qb
		}
		return true
	}
	return pair(c.R, c.G, q.R, q.G) &&
		pair(c.G, c.B, q.G, q.B) &&
		pair(c.R, c.B, q.R, q.B)
}

// Saturate pushes a colour away from grey without changing how bright it is.
//
// This is what makes a 256-colour terminal show a blue night. The cube has no
// resolution in the dark, so BACKGROUNDS there are doomed to grey -- but a
// background is meant to be dark, and dark IS night. The glyphs are the bright
// part of the frame, and brightness is exactly where the cube gets generous:
// 4 real colours below luma 25, but 46 between 110 and 150 and 108 above 150.
// So the darkness stays in the ground and the colour goes in the texture,
// which is how ASCII art has always worked.
func (c RGB) Saturate(k float64) RGB {
	// Leave near-neutrals alone. The companion's fur is a barely-warm white,
	// and multiplying that turns the cat orange -- which measured perfectly and
	// looked absurd.
	mx, mn := c.R, c.R
	for _, v := range []uint8{c.G, c.B} {
		if v > mx {
			mx = v
		}
		if v < mn {
			mn = v
		}
	}
	ch := float64(mx) - float64(mn)
	if ch < neutralChroma {
		return c
	}

	// Lift TOWARD a target chroma rather than multiplying by a factor.
	//
	// Multiplying is unbounded, so it does nothing useful to a colour that is
	// already saturated and a great deal of damage: Claude's terracotta went
	// neon on a 256 terminal, because 2.6x of an already-vivid orange clips
	// every channel it can. The point of the boost is only to clear the grey
	// ramp, and around 100 points of chroma is enough for that -- past it there
	// is no gain, just distortion. So colours that need help get a lot and
	// colours that do not get none.
	want := ch + (chromaTarget-ch)*(k-1)/2
	if want <= ch {
		return c
	}
	scale := want / ch

	l := 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	f := func(v uint8) uint8 {
		x := l + (float64(v)-l)*scale
		if x < 0 {
			x = 0
		} else if x > 255 {
			x = 255
		}
		return uint8(x + 0.5)
	}
	return RGB{f(c.R), f(c.G), f(c.B)}
}

// chromaTarget is how much colour is enough to clear the greyscale ramp.
// Measured: the ramp stops winning on distance somewhere around 90.
const chromaTarget = 100

// neutralChroma is how close to grey a colour has to be before the boost
// leaves it alone.
//
// The threshold is not arbitrary: measured across the palette, everything
// MEANT to read as white clusters tightly just under 30 -- cat fur 26, the
// moon 26, foam 24, the newest line of sand 24 -- and everything meant to
// read as a colour is far above it: sea glyphs 80, the worried amber 148.
// There is a clean gap, so one number separates them.
//
// Getting this wrong is visible rather than subtle. At 18 the cat was inside
// the boost and turned peach, which measured perfectly and looked absurd.
const neutralChroma = 30

// GlyphBoost is how hard glyph colours are pushed away from grey before they
// are quantised for a 256-colour terminal. 1.0 disables it.
//
// Measured on a full frame: at 1.0 only 45% of glyphs land on a real colour and
// the rest fall into the grey ramp; at 2.0 it is 100%. Backgrounds are
// deliberately NOT boosted -- they are the dark part of a night and belong in
// the greys.
var GlyphBoost = 2.6

func init() {
	// Tunable without a rebuild, because the right value is a judgement made
	// by looking at a real terminal, not by reading a number.
	if v := os.Getenv("ASCIISCAPES_CHROMA"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 1 && f <= 6 {
			GlyphBoost = f
		}
	}
	switch os.Getenv("ASCIISCAPES_SHADE") {
	case "0", "off", "no":
		Shading = false
	}
	switch os.Getenv("ASCIISCAPES_SHADE_BLOCKS") {
	case "1", "on", "yes":
		ShadeBlocks = true
	}
	if v := os.Getenv("ASCIISCAPES_HUE_WEIGHT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 64 {
			hueWeight = n
		}
	}
}

// Shading splits a cell in half so a vertical ramp gets twice the resolution.
//
// The cube has six levels per channel with 40 between them, so a sky ramped
// over seventeen rows lands on about seven colours and the rest is step, step,
// step. Choosing better colours does not help; the levels are the levels.
//
// A terminal cell is not one colour, though. U+2580 paints the top half in the
// foreground and leaves the bottom to the background, so a row can carry two
// steps of the ramp instead of one. It costs no accuracy and adds no texture --
// the same colours, placed twice as finely -- and it is what actually made the
// gradients look smoother. See ShadeBlocks for the idea that did not.
//
// Off for ASCII-only output, and ASCIISCAPES_SHADE=0 turns it off everywhere.
var Shading = true

// ShadeBlocks is the tone-blending half of the smoothing, and it is OFF.
//
// It works, in the sense that it does what it claims: it puts three more tones
// between every pair of cube colours and the measured tone count in the sky
// went from 11 to 14. It was still the wrong answer. Rendered beside the plain
// version, a shaded sky reads as a band of stipple rather than as a smoother
// ramp -- the dot pattern of U+2591 and U+2592 is a texture the eye finds
// before it finds the tone, and a sky is the one region with nothing else in it
// to hide behind. More tones, worse picture.
//
// Kept rather than deleted for one reason: that judgement was made from a
// browser render at 11px, and shade blocks are exactly the kind of glyph a real
// terminal at a real font size may resolve differently. ASCIISCAPES_SHADE_BLOCKS=1
// turns it on so the question can be settled where it matters.
var ShadeBlocks = false

// shadeSteps are the coverages available in one cell, and the glyph for each.
var shadeSteps = [...]struct {
	f float64
	r rune
}{
	{0.00, ' '},
	{0.25, '\u2591'},
	{0.50, '\u2592'},
	{0.75, '\u2593'},
	{1.00, ' '}, // the upper colour, painted as a plain background
}

// bracket returns the cube levels a channel value falls between and how far
// along it sits.
func bracket(v uint8) (lo, hi uint8, f float64) {
	for i := 0; i+1 < len(cubeLevels); i++ {
		l, h := cubeLevels[i], cubeLevels[i+1]
		if v >= l && v <= h {
			return l, h, (float64(v) - float64(l)) / (float64(h) - float64(l))
		}
	}
	return cubeLevels[5], cubeLevels[5], 0
}

// SplitPair was the idea of stacking the two cube colours a tone sits between
// so the cell averages to something the palette does not contain. It is kept as
// a record of a measured dead end, unused.
//
// Two things killed it. Measured, it tracked the gradient WORSE than placing
// the band edge does -- 242 against 239 mean squared error on a 17-row sky --
// because each half is now a full step away from its own true colour even
// though the pair averages correctly. And the averaging is the assumption that
// fails: it only holds if the eye cannot resolve half a cell, and half a cell
// in a terminal is three or four pixels of line height. It resolves. What you
// get is stripes.
func SplitPair(c RGB) (lo, hi RGB, ok bool) {
	lr, hr, _ := bracket(c.R)
	lg, hg, _ := bracket(c.G)
	lb, hb, _ := bracket(c.B)
	if int(hr)-int(lr) > 40 || int(hg)-int(lg) > 40 || int(hb)-int(lb) > 40 {
		return c, c, false
	}
	lo, hi = RGB{lr, lg, lb}, RGB{hr, hg, hb}
	if dist(c, lo.Blend(hi, 0.5)) >= dist(c, FromIndex256(c.Index256())) {
		return c, c, false
	}
	return lo, hi, true
}

// ShadeCell approximates c with two cube colours and a shade glyph.
//
// It reports ok=false when that would be WORSE than simply quantising, which
// happens whenever the channels sit at different points in their gaps: one
// coverage cannot serve three fractions, and the greyscale ramp is often nearer
// than either corner. The caller falls back to a plain cell there.
func ShadeCell(c RGB) (bg, fg RGB, r rune, ok bool) {
	if !ShadeBlocks {
		return c, c, ' ', false
	}
	lr, hr, _ := bracket(c.R)
	lg, hg, _ := bracket(c.G)
	lb, hb, _ := bracket(c.B)
	lo := RGB{lr, lg, lb}
	hi := RGB{hr, hg, hb}

	// The coverage is chosen by measured error rather than by averaging the
	// three channel fractions. Averaging is what you reach for first and it is
	// wrong whenever the channels sit at different points in their gaps -- the
	// mean lands between all three and matches none of them.

	// Only blend across a SMALL gap. The cube's first step is 0 to 95 and the
	// rest are 40, and a half-shaded cell whose two colours are 95 apart does
	// not read as a tone between them -- it reads as crosshatch, which is
	// exactly the texture this is supposed to avoid. Rendered and rejected:
	// the upper sky at noon came out visibly woven.
	if int(hr)-int(lr) > 40 || int(hg)-int(lg) > 40 || int(hb)-int(lb) > 40 {
		return c, c, ' ', false
	}

	best, bestErr := 0, 1<<30
	for i, s := range shadeSteps {
		if d := dist(c, lo.Blend(hi, s.f)); d < bestErr {
			bestErr, best = d, i
		}
	}
	// A plain cell is the thing to beat. The bar was once "beat it twice over",
	// which sounded prudent and left the smoothing switched off in 94% of
	// cells; with the gap limit above doing the work of keeping the pattern
	// quiet, plain improvement is the right bar.
	if bestErr >= dist(c, FromIndex256(c.Index256())) {
		return c, c, ' ', false
	}
	switch st := shadeSteps[best]; {
	case st.f == 0:
		return lo, lo, ' ', true
	case st.f == 1:
		return hi, hi, ' ', true
	default:
		return lo, hi, st.r, true
	}
}

// AppendFGRaw writes a foreground WITHOUT the chroma boost.
//
// The boost exists for glyphs, which are the bright texture of the frame. A
// shade block is not texture, it is half a background, and lifting its chroma
// would pull one half of a cell away from the other and put an edge where the
// whole point was to remove one.
func (p Profile) AppendFGRaw(dst []byte, c RGB) []byte {
	dst = append(dst, 0x1b, '[', '3', '8', ';')
	if p == ProfileTrueColor {
		dst = append(dst, '2', ';')
		dst = strconv.AppendInt(dst, int64(c.R), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(c.G), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(c.B), 10)
	} else {
		dst = append(dst, '5', ';')
		// A split cell's foreground is half a background, so it takes the
		// background's hue-preserving quantiser rather than the glyph one.
		dst = strconv.AppendInt(dst, int64(c.Index256Keeping()), 10)
	}
	return append(dst, 'm')
}

// FromIndex256 is the inverse of Index256: the colour a terminal actually
// paints for an index. Needed to SHOW what a 256-colour terminal will do with
// a palette, rather than describing it -- the HTML harness renders true RGB,
// so without this a preview flatters the design by showing colours the target
// terminal cannot produce.
func FromIndex256(i int) RGB {
	switch {
	case i >= 232 && i <= 255:
		v := uint8(8 + (i-232)*10)
		return RGB{v, v, v}
	case i >= 16 && i <= 231:
		n := i - 16
		return RGB{cubeLevels[n/36], cubeLevels[(n/6)%6], cubeLevels[n%6]}
	}
	return RGB{}
}

// Quantise returns the colour as the given profile would actually show it.
// fg says whether this is a glyph rather than a background, because only
// glyphs get the chroma boost.
func (p Profile) Quantise(c RGB, fg bool) RGB {
	if p == ProfileTrueColor {
		return c
	}
	if fg {
		// Glyphs keep the plain quantiser. The chroma boost has already moved
		// the colour on purpose; preserving a channel ordering the boost just
		// changed would be defending the wrong thing.
		return FromIndex256(c.Saturate(GlyphBoost).Index256())
	}
	return FromIndex256(c.Index256Keeping())
}

// AppendFG / AppendBG write an SGR sequence without allocating a string.
func (p Profile) AppendFG(dst []byte, c RGB) []byte { return p.appendSGR(dst, c, true) }
func (p Profile) AppendBG(dst []byte, c RGB) []byte { return p.appendSGR(dst, c, false) }

func (p Profile) appendSGR(dst []byte, c RGB, fg bool) []byte {
	dst = append(dst, 0x1b, '[')
	if fg {
		dst = append(dst, '3', '8', ';')
	} else {
		dst = append(dst, '4', '8', ';')
	}
	if p == ProfileTrueColor {
		dst = append(dst, '2', ';')
		dst = strconv.AppendInt(dst, int64(c.R), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(c.G), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(c.B), 10)
	} else {
		if fg {
			c = c.Saturate(GlyphBoost)
		}
		dst = append(dst, '5', ';')
		if fg {
			dst = strconv.AppendInt(dst, int64(c.Index256()), 10)
		} else {
			dst = strconv.AppendInt(dst, int64(c.Index256Keeping()), 10)
		}
	}
	return append(dst, 'm')
}

const Reset = "\x1b[0m"
