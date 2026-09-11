// Command todcheck is a positive control for section E of s28-drift-verify.
//
// Section E reports IDENTICAL numbers for time of day 01h, 06h, 13h and 21h,
// for context used 0.00 and 0.95, and for term.NoSplitCells true and false. A
// variation that changes nothing looks exactly like a variation that was never
// applied, so this proves each knob DOES change the rendered frame -- just not
// the three glyphs the drift measurement counts.
//
//	go run ./notes/s28-drift-verify/todcheck
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

const (
	W, H   = 120, 26
	seed   = 7
	stepDT = 0.08
)

type cell struct {
	r      rune
	fg, bg term.RGB
}

func frame(act scape.Activity, warm int) []cell {
	sh := scape.NewShore(seed, false)
	t := 0.0
	var c *canvas.Canvas
	for i := 0; i <= warm; i++ {
		c = canvas.New(W, H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		t += stepDT
		sh.Update(c, t, act)
	}
	out := make([]cell, 0, W*H)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			out = append(out, cell{r, fg, bg})
		}
	}
	return out
}

func drift(r rune) bool { return r == '≈' || r == '~' || r == '≋' }

func diff(a, b []cell) (all, glyphOnly, driftOnly int) {
	for i := range a {
		if a[i] != b[i] {
			all++
			if a[i].r != b[i].r {
				glyphOnly++
			}
		}
		if drift(a[i].r) != drift(b[i].r) {
			driftOnly++
		}
	}
	return
}

func main() {
	base := scape.Activity{Level: 0.6, Working: true, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	ref := frame(base, 400)
	fmt.Printf("Terminal cells that differ from the 13h / ctx 0.30 / NoSplitCells=false frame,\n")
	fmt.Printf("out of %d cells in a %dx%d frame (warm-up 400):\n\n", W*H, W, H)
	fmt.Printf("  %-40s %-14s %-14s %s\n", "variant", "cells differ", "glyph differs", "drift-glyph differs")

	show := func(name string, c []cell) {
		a, g, d := diff(ref, c)
		fmt.Printf("  %-40s %-14d %-14d %d\n", name, a, g, d)
	}
	for _, tod := range []float64{1.0 / 24, 6.0 / 24, 21.0 / 24} {
		a := base
		a.TimeOfDay = tod
		show(fmt.Sprintf("time of day %04.1fh", tod*24), frame(a, 400))
	}
	for _, cu := range []float64{0.0, 0.95} {
		a := base
		a.ContextUsed = cu
		show(fmt.Sprintf("context used %.2f", cu), frame(a, 400))
	}
	was := term.NoSplitCells
	term.NoSplitCells = true
	show("term.NoSplitCells=true", frame(base, 400))
	term.NoSplitCells = was
	fmt.Printf("  (restored term.NoSplitCells=%v)\n", term.NoSplitCells)
	fmt.Println()
	fmt.Println("  A zero in the first column would mean the knob never reached the renderer")
	fmt.Println("  and section E measured nothing. A large first column with a zero in the last")
	fmt.Println("  is the real result: these knobs repaint the frame but do not move the sea's")
	fmt.Println("  three drift glyphs, so the drift measurement is genuinely insensitive to them.")
}
