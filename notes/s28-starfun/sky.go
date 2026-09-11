package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// His scape, measured: 125 columns by 28 rows (125x62 window, host.Band gives
// the scape ~28). hy = int(H*0.42) = 11, so the sky is rows 0..10.
const (
	W    = 125
	H    = 28
	TOD  = 20.0 / 24 // night. The constellation is meant to be read at any hour,
	CTX  = 0.56      // but night is where a wrong tone shows first.
	Seed = 7
	// todoStarFloor, copied from shore.go. It is a FLOOR: the channel has to be
	// legible at noon, so nothing may be plotted under it.
	Floor = 0.85
)

type sky struct {
	sh  *scape.Shore
	c   *canvas.Canvas
	pal scape.Palette
	hy  int
}

// warm runs ONE shore forward. A fresh Shore per frame has no history --
// Update clamps a gap over a second to one nominal step -- and that trap has
// voided two instruments in this project already.
func warm(done, total int, ctx float64) *sky { return warmAt(done, total, ctx, TOD) }

func warmAt(done, total int, ctx, tod float64) *sky {
	sh := scape.NewShore(Seed, false)
	var c *canvas.Canvas
	t := 0.0
	for k := 0; k < 300; k++ {
		c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += 0.08
		sh.Update(c, t, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: ctx, TimeOfDay: tod,
			TodoDone: done, TodoTotal: total,
		})
	}
	return &sky{sh: sh, c: c, pal: scape.PaletteAt(tod), hy: skyRows()}
}

type pt struct{ x, y int }

// placeToday replicates todoStars' own arithmetic so a candidate can be drawn
// on a sky the product is NOT drawing stars into. It is verified against the
// product's rendered '*' cells in check() before anything is believed.
func placeHead(s *sky, total, done int) []pt {
	top, bot := 1, s.hy*3/5
	if bot <= top {
		bot = top + 1
	}
	var out []pt
	for i := 0; i < total && i < done; i++ {
		frac := math.Mod((float64(i)+0.5)*0.6180339887, 1)
		x := int(frac*float64(W-6)) + 3
		x += int(scape.HashF(i, 11, Seed+31)*3) - 1
		y := top + int(scape.HashF(i, 23, Seed+37)*float64(bot-top))
		if x < 0 || x >= W || y < 0 || y >= H {
			continue
		}
		if s.sh.DiscCovers(x, y) {
			continue
		}
		out = append(out, pt{x, y})
	}
	return out
}

// starCells reads the RENDERED frame back. '*' belongs to the constellation
// alone -- shore.go: "a channel that shares a glyph with the scenery is not a
// channel" -- so every '*' on screen is one finished unit of work.
func starCells(c *canvas.Canvas) []pt {
	var out []pt
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				out = append(out, pt{x, y})
			}
		}
	}
	return out
}

// asciiSky crops the rendered frame to the sky and one row of horizon. Every
// glyph comes off ResolveAt, not off what the source says it plotted.
func asciiSky(c *canvas.Canvas, hy int) []string {
	var rows []string
	for y := 0; y <= hy; y++ {
		line := make([]rune, W)
		for x := 0; x < W; x++ {
			r, _, _ := c.ResolveAt(x, y, term.Profile256)
			if r == 0 || r == ' ' {
				r = ' '
			}
			line[x] = r
		}
		rows = append(rows, trimRight(string(line)))
	}
	return rows
}

func trimRight(s string) string {
	i := len(s)
	for i > 0 && s[i-1] == ' ' {
		i--
	}
	return s[:i]
}

func show(label string, rows []string) {
	fmt.Printf("\n%s\n", label)
	fmt.Println("   +" + rule(W) + "+")
	for i, r := range rows {
		fmt.Printf("%2d |%-*s|\n", i, W, r)
	}
	fmt.Println("   +" + rule(W) + "+")
}

func rule(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}

func luma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// merges counts pairs that read as ONE star. A terminal cell is about twice as
// tall as it is wide, so the visual distance between (dx,dy) cells is
// hypot(dx, 2*dy): side by side merges, one above the other does not.
func merges(ps []pt) int {
	n := 0
	for i := range ps {
		for j := i + 1; j < len(ps); j++ {
			dx := float64(ps[i].x - ps[j].x)
			dy := float64(ps[i].y - ps[j].y)
			if math.Hypot(dx, 2*dy) < 2 {
				n++
			}
		}
	}
	return n
}

func skyRows() int {
	h := H
	return int(float64(h) * 0.42)
}

// asciiQuiet is asciiSky with the sky gradient's split cells blanked. Those
// half blocks carry BACKGROUND, not ink -- rows 2 and 9 here are the sky ramp
// and the sea's first row -- and they swamp a text render. Cells the disc
// reaches keep theirs, so the moon still shows.
func asciiQuiet(s *sky) []string {
	var rows []string
	for y := 0; y <= s.hy; y++ {
		line := make([]rune, W)
		for x := 0; x < W; x++ {
			r, _, _ := s.c.ResolveAt(x, y, term.Profile256)
			if r == 0 {
				r = ' '
			}
			if r == '▀' && !s.sh.DiscCovers(x, y) {
				r = ' '
			}
			line[x] = r
		}
		rows = append(rows, trimRight(string(line)))
	}
	return rows
}

// plotStar puts one constellation star on the near layer, the way todoStars
// does: '*', the palette's star colour, at the floor.
func (s *sky) plotStar(p pt, fg term.RGB, a float64) {
	s.c.Near().Plot(p.x, p.y, '*', fg, a)
}

func lerpRGB(a, b term.RGB, t float64) term.RGB { return term.Lerp(a, b, t) }

func rgbs(c term.RGB) string { return fmt.Sprintf("(%3d,%3d,%3d)", c.R, c.G, c.B) }

// placeLive is the placement the product ACTUALLY renders, whatever it is
// today. It recovers lit ORDER by advancing the count one at a time and taking
// the cell that appears -- B, C and D all need to know which star is which, and
// a frame read alone only gives a set.
//
// A slot the disc covers contributes no cell; it is skipped rather than
// guessed, which is the same thing the product does.
func placeLive(n int) []pt {
	prev := map[pt]bool{}
	var out []pt
	for i := 0; i <= n; i++ {
		s := warm(i, 32, CTX)
		cur := starCells(s.c)
		seen := map[pt]bool{}
		for _, p := range cur {
			seen[p] = true
			if i > 0 && !prev[p] {
				out = append(out, p)
			}
		}
		prev = seen
	}
	return out
}

// bandOf reports the rows the rendered constellation occupies.
func bandOf(ps []pt) (lo, hi int) {
	lo, hi = 99, -1
	for _, p := range ps {
		if p.y < lo {
			lo = p.y
		}
		if p.y > hi {
			hi = p.y
		}
	}
	return
}
