package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// timeSweep holds context, count and geometry STILL and advances only the
// clock, which is what he was actually looking at: he watched for minutes, his
// context barely moved, and no star was earned or lost in between.
//
// It counts two fields separately, because they are two channels and only one
// of them is allowed to blink:
//
//	'*'                 the constellation. One per unit of work finished. Must
//	                    be steady -- his words, "once they appear, they should
//	                    not disappear".
//	'.' '·' '+'    the ambient field, which twinkles BY DESIGN -- but the
//	                    twinkle is gated `if twinkle <= 0.02 { continue }`, so
//	                    below a certain StarVis it stops dimming and starts
//	                    blinking out altogether.
func timeSweep() {
	fmt.Println("\n=== clock only: context, count and size all held still ===")
	fmt.Printf("%-9s %-6s %8s %10s %10s %10s %10s\n",
		"geom", "hour", "StarVis", "const-min", "const-max", "amb-mean", "amb-blinks")
	for _, hr := range []float64{0, 3, 6, 8, 9, 10, 11, 12, 15, 18, 20, 21, 22} {
		tod := hr / 24
		g := geom{143, 27}
		sh := scape.NewShore(7, false)
		tm := 0.0
		step := func() *canvas.Canvas {
			c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: 0.35, TimeOfDay: tod,
				TodoDone: 16, TodoTotal: 32,
			})
			return c
		}
		for k := 0; k < 60; k++ {
			step()
		}
		var cmin, cmax = 1 << 30, 0
		var ambSum, ambBlink float64
		prevAmb := map[[2]int]bool{}
		const frames = 300 // 24 s at 12.5 fps
		for f := 0; f < frames; f++ {
			c := step()
			nConst := 0
			amb := map[[2]int]bool{}
			for y := 0; y < g.h; y++ {
				for x := 0; x < g.w; x++ {
					r, _, _ := c.ResolveAt(x, y, term.Profile256)
					switch r {
					case '*':
						nConst++
					case '.', '·', '+':
						amb[[2]int{x, y}] = true
					}
				}
			}
			if nConst < cmin {
				cmin = nConst
			}
			if nConst > cmax {
				cmax = nConst
			}
			ambSum += float64(len(amb))
			if f > 0 {
				for cell := range prevAmb {
					if !amb[cell] {
						ambBlink++ // was drawn last frame, gone this frame
					}
				}
			}
			prevAmb = amb
		}
		fmt.Printf("%-9s %-6.0f %8.3f %10d %10d %10.1f %10.0f\n",
			fmt.Sprintf("%dx%d", g.w, g.h), hr, starVisAt(tod), cmin, cmax,
			ambSum/float64(frames), ambBlink)
	}
}

// starVisAt reads the palette's own StarVis, so this cannot drift from what the
// renderer uses.
func starVisAt(tod float64) float64 { return scape.PaletteAt(tod).StarVis }

var _ = math.Abs
