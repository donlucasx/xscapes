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
		c = c.Saturate(GlyphBoost)
	}
	return FromIndex256(c.Index256())
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
		dst = strconv.AppendInt(dst, int64(c.Index256()), 10)
	}
	return append(dst, 'm')
}

const Reset = "\x1b[0m"
