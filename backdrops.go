package main

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/scape"
)

// SECTION BACKDROPS: a different, very quiet, moving field of glyphs behind
// each section of the page.
//
// His idea, 2026-09-14: "each 'section' could have different animated ascii
// backgrounds (or elements), all very subtle. They dont HAVE to be b&w, as
// long as colors are not taking away from what we are communicating." It is
// also his answer to the dividers, which failed three times. Ruled 2026-09-15:
// one subtle pass now.
//
// Each backdrop is six frames of plain text, stacked as <pre> layers and
// cycled by the cover's own step animation (0.7 s a frame, 4.2 s a loop), at
// 7% opacity in one colour behind the section's content. Every field is
// written so that its motion CLOSES on the sixth frame -- a wave whose phase
// is k/6 of a turn, rain whose pattern repeats every 24 rows at 4 rows a
// frame -- because a loop that jumps is the one thing a subtle layer must not
// do: nobody sees a backdrop until it stutters.
//
// They replace the one static sea that used to sit behind the whole page.

// bdCols and bdRows size a backdrop: wider than any main column at 14px, and
// tall enough that the tallest section is mostly covered; a mask fades the
// top and bottom so a shorter section just sees the middle.
const (
	bdCols   = 170
	bdRows   = 100
	bdFrames = 6
)

// A backdrop is one section's field: which section, its colour class, and
// the painter that says what glyph a cell holds on frame k.
type backdrop struct {
	key   string
	class string
	glyph func(x, y, k int, seed int64) rune
}

// h is a stable per-cell number in [0, 1).
func h(x, y int, seed int64) float64 { return scape.HashF(x, y, seed) }

// wave is a travelling wave that makes exactly one turn over the six
// frames, so frame 6 is frame 0.
func wave(x, y, k int, lambda, tilt float64) float64 {
	return math.Sin(2*math.Pi*(float64(x)/lambda+float64(k)/bdFrames) + float64(y)*tilt)
}

var backdrops = []backdrop{
	// The problem: water with almost nothing on it, the way the sea sits while
	// the agent thinks and nothing has happened yet.
	{"problem", "sea", func(x, y, k int, seed int64) rune {
		if h(x, y, seed) > 0.045 {
			return ' '
		}
		switch v := wave(x, y, k, 46, 0.35); {
		case v > 0.55:
			return '~'
		case v > -0.1:
			return '-'
		}
		return '.'
	}},
	// While you wait: the same sea, working -- denser, shorter swells, crests.
	{"wait", "sea", func(x, y, k int, seed int64) rune {
		if h(x, y, seed+1) > 0.11 {
			return ' '
		}
		switch v := wave(x, y, k, 28, 0.5); {
		case v > 0.82:
			return '^'
		case v > 0.35:
			return '~'
		case v > -0.3:
			return '-'
		}
		return '.'
	}},
	// Read it at a glance: the constellation, each star on its own twinkle.
	{"glance", "star", func(x, y, k int, seed int64) rune {
		if h(x, y, seed+2) > 0.012 {
			return ' '
		}
		ph := int(h(x, y, seed+3) * bdFrames)
		return []rune("·.+*+.")[(k+ph)%bdFrames]
	}},
	// How you get it: rain, four rows a frame through a 24-row pattern, so
	// six frames are one period and the streaks fall without a seam.
	{"get", "rain", func(x, y, k int, seed int64) rune {
		yy := ((y-4*k)%24 + 24) % 24
		v := h(x, yy, seed+4)
		switch {
		case v < 0.018:
			return '|'
		case v < 0.03:
			return '\''
		case v < 0.04:
			return '.'
		}
		return ' '
	}},
	// A layer, not a screen: the meadow in the wind. A gust runs through the
	// grass and every blade leans with it.
	{"layer", "grass", func(x, y, k int, seed int64) rune {
		if h(x, y, seed+5) > 0.16 {
			return ' '
		}
		switch v := wave(x, y, k, 60, 0.15); {
		case v > 0.45:
			return ','
		case v < -0.45:
			return '`'
		}
		return '\''
	}},
	// Still good in a month: the aquarium's bubbles, two rows a frame through
	// a 12-row pattern, and a fish or two crossing.
	{"month", "water", func(x, y, k int, seed int64) rune {
		// The fish: five cells a frame over a 30-cell period, so six frames
		// bring each one back to where it started.
		fx := ((x-5*k)%30 + 30) % 30
		if y%17 == 6 && h(0, y, seed+7) < 0.5 && fx < 3 {
			return []rune("><>")[fx]
		}
		yy := ((y+2*k)%12 + 12) % 12
		v := h(x, yy, seed+6)
		switch {
		case v < 0.006:
			return 'O'
		case v < 0.016:
			return 'o'
		case v < 0.03:
			return '.'
		}
		return ' '
	}},
}

// renderBackdrops returns each section's stacked frames by key.
func renderBackdrops(seed int64) map[string]string {
	out := map[string]string{}
	for _, bd := range backdrops {
		var b strings.Builder
		for k := 0; k < bdFrames; k++ {
			var f strings.Builder
			for y := 0; y < bdRows; y++ {
				row := make([]rune, bdCols)
				for x := range row {
					row[x] = bd.glyph(x, y, k, seed)
				}
				f.WriteString(strings.TrimRight(string(row), " "))
				f.WriteByte('\n')
			}
			fmt.Fprintf(&b, `<pre style="animation-delay:-%.1fs">%s</pre>`, float64(k)*0.7, html.EscapeString(f.String()))
		}
		out[bd.key] = b.String()
	}
	return out
}
