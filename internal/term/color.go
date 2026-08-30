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
		dst = append(dst, '5', ';')
		dst = strconv.AppendInt(dst, int64(c.Index256()), 10)
	}
	return append(dst, 'm')
}

const Reset = "\x1b[0m"
