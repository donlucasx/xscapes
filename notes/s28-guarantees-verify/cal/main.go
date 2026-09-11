// cal is the counterfactual the s28-guarantees finding never ran: it measures
// the churn at three geometries so shore.go:546 can be edited underneath it.
// Run it, patch the depth anchor, run it again, revert.
package main

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func edgeOf(sh *scape.Shore) []float64 {
	v := reflect.ValueOf(sh).Elem().FieldByName("lastEdge")
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface().([]float64)
}

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

func openSea(edge []float64, h int) (int, int) {
	lo := h
	for _, e := range edge {
		if int(e) < lo {
			lo = int(e)
		}
	}
	return int(float64(h)*0.42) + 1, lo - 2
}

// run is churnRun from the owning test: 24 frames at 0.1s, the band taken from
// the final frame, counted on BGAt (the test) and on the 256 cube (the screen).
func run(sh *scape.Shore, w, h int, act scape.Activity, tm *float64) (raw, ren float64, cells, big int) {
	var fr []*canvas.Canvas
	for i := 0; i < 24; i++ {
		cc := newCanvas(w, h)
		*tm += 0.1
		sh.Update(cc, *tm, act)
		fr = append(fr, cc)
	}
	from, to := openSea(edgeOf(sh), h)
	cr, ce := 0, 0
	for i := 1; i < len(fr); i++ {
		pc, n := 0, 0
		for y := from; y <= to; y++ {
			if y < 0 || y >= h {
				continue
			}
			for x := 0; x < w; x++ {
				cells++
				n++
				if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
					cr++
					pc++
				}
				_, _, a := fr[i-1].ResolveAt(x, y, term.Profile256)
				_, _, b := fr[i].ResolveAt(x, y, term.Profile256)
				if a != b {
					ce++
				}
			}
		}
		if n > 0 && 100*float64(pc)/float64(n) > 50 {
			big++
		}
	}
	if cells == 0 {
		return 0, 0, 0, 0
	}
	return 100 * float64(cr) / float64(cells), 100 * float64(ce) / float64(cells), cells, big
}

func main() {
	fmt.Printf("%-9s %-6s %9s %9s %9s %9s\n", "geom", "mode", "raw%", "REN%", "cells", "pairs>50%")
	for _, g := range [][2]int{{120, 26}, {80, 24}, {128, 27}} {
		w, h := g[0], g[1]
		for _, mode := range []string{"WARM", "STEP"} {
			sh := scape.NewShore(7, false)
			c := newCanvas(w, h)
			tm := 0.0
			lv := 0.7
			warm := scape.Activity{Level: lv, TimeOfDay: 0.3868}
			if mode == "STEP" {
				warm = scape.Activity{Level: 0, TimeOfDay: 0.3868}
			}
			for i := 0; i < 400; i++ {
				tm += 0.1
				sh.Update(c, tm, warm)
			}
			raw, ren, cells, big := run(sh, w, h, scape.Activity{Level: lv, TimeOfDay: 0.3868}, &tm)
			fmt.Printf("%-9s %-6s %8.2f%% %8.2f%% %9d %9d\n",
				fmt.Sprintf("%dx%d", w, h), mode, raw, ren, cells, big)
		}
	}
	_ = math.Pi
}
