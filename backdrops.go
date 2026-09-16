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
// low opacity in one colour behind the section's content. Every field is
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

// THE CRITERION, his question of 2026-09-15 ("Placement - whats the criteria?
// what is under 'features'? grass?"): each backdrop is THE THING ITS SECTION
// IS ABOUT, drawn so that it reads as that thing at a glance -- the first
// pass scattered single marks, and a working sea of scattered ^ and ~ read as
// grass. A section whose content already floats on the page (the companion
// states) gets none, and that alternation is what separates the sections.
var backdrops = []backdrop{
	// The problem: the wait itself. The spinner Claude Code turns while you
	// look at nothing, scattered across the section and turning.
	{"problem", "spin", func(x, y, k int, seed int64) rune {
		if h(x, y, seed) > 0.006 {
			return ' '
		}
		ph := int(h(x, y, seed+10) * bdFrames)
		return []rune("✻✳✶✽✢·")[(k+ph)%bdFrames]
	}},
	// Features: the sea, the product's own picture. Crests are RUNS of glyphs
	// along a row, not scattered marks -- that is what reads as water -- on
	// some rows and not others, travelling.
	{"wait", "sea", func(x, y, k int, seed int64) rune {
		if h(0, y, seed+1) > 0.42 {
			return ' '
		}
		// Each row's own phase, so crests do not line up into columns.
		v := math.Sin(2*math.Pi*(float64(x)/22+float64(k)/bdFrames) + h(1, y, seed+11)*2*math.Pi)
		switch {
		case v > 0.78:
			return '≈'
		case v > 0.45:
			return '~'
		case v > 0.25:
			return '-'
		}
		return ' '
	}},
	// Install: the terminal. Prompts, and a cursor that blinks on each.
	{"get", "term", func(x, y, k int, seed int64) rune {
		if y%4 != 1 {
			return ' '
		}
		if h(x, y, seed+4) < 0.012 {
			return '$'
		}
		if h(x-2, y, seed+4) < 0.012 && k%2 == 0 {
			return '▌'
		}
		return ' '
	}},
	// Scapes and companions: rain, which is the rainy window's whole slot,
	// four rows a frame through a 24-row pattern, so six frames are one
	// period and the streaks fall without a seam.
	{"layer", "rain", func(x, y, k int, seed int64) rune {
		yy := ((y-4*k)%24 + 24) % 24
		v := h(x, yy, seed+5)
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
	// Different every session: a night sky, no two loops of it the same to
	// look at -- each star on its own twinkle, and now and then one falls:
	// three cells down and along over three frames, then gone.
	{"month", "star", func(x, y, k int, seed int64) rune {
		// The falls: one per 25-row band, from a hashed start, frames 0-2.
		band := y / 25
		sx := 10 + int(h(band, 0, seed+8)*(bdCols-30))
		sy := band*25 + 3 + int(h(band, 1, seed+8)*12)
		if k < 3 && x == sx+2*k && y == sy+k {
			return '*'
		}
		if k < 3 && k > 0 && x == sx+2*k-2 && y == sy+k-1 {
			return '·'
		}
		if h(x, y, seed+2) > 0.012 {
			return ' '
		}
		ph := int(h(x, y, seed+3) * bdFrames)
		return []rune("·.+*+.")[(k+ph)%bdFrames]
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
