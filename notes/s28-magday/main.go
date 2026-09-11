// Command magday answers two of his 2026-09-11 notes at once:
//
//	"Magnitude. Do they seem right in the first screenshot? i dont see a big
//	 difference in brightness between stars."
//	"during the daytime I dont see any other characters tho like I saw before
//	 during nightime giving depth to the constellation"
//
// His first screenshot is 12:34 in a bright blue sky. The suspicion to test is
// that starInk, which lifts the star's tone toward white until it clears the
// contrast bar, RUNS OUT OF ROOM by day: every magnitude gets lifted to the
// same white and the variation he was promised collapses to one tone.
//
//	go run ./notes/s28-magday
package main

import (
	"fmt"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	w, h := 114, 28
	hy := int(float64(h) * 0.42)
	fmt.Printf("his window 114x64, scape band ~%dx%d\n\n", w, h)
	fmt.Printf("%-8s %8s %10s %10s %12s   %s\n",
		"hour", "StarVis", "const'n", "distinct", "ambient", "how it reads")
	for _, hr := range []float64{0, 6, 9, 12, 12.6, 15, 18, 20, 22} {
		con, tones, amb := probe(w, h, hy, hr/24)
		read := "varied"
		if len(tones) <= 1 {
			read = "ONE TONE -- no magnitude visible"
		} else if len(tones) <= 2 {
			read = "two tones"
		}
		extra := ""
		if amb == 0 {
			extra = "  + NO ambient dust at all"
		}
		fmt.Printf("%-8.1f %8.3f %10d %10d %12d   %s%s\n",
			hr, scape.PaletteAt(hr/24).StarVis, con, len(tones), amb, read, extra)
	}
	lumaSpread()
	fmt.Println("\ntones are the DISTINCT quantised foregrounds the constellation renders in.")
	fmt.Println("Six at night was the design target; one means every star was lifted to the")
	fmt.Println("same white to clear the contrast bar against a bright sky.")
}

func probe(w, h, hy int, tod float64) (con int, tones []int, amb int) {
	sh := scape.NewShore(7, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 200; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.26, TimeOfDay: tod,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	seen := map[int]bool{}
	for y := 0; y < hy; y++ {
		for x := 0; x < w; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			switch r {
			case '*':
				con++
				seen[fg.Index256()] = true
			case '.', '·', '+':
				if fg != bg {
					amb++
				}
			}
		}
	}
	for t := range seen {
		tones = append(tones, t)
	}
	sort.Ints(tones)
	return con, tones, amb
}

// lumaSpread is the question his eye is actually asking. Counting DISTINCT
// palette indices says the magnitudes reached the screen; it says nothing about
// whether they are far enough apart to see. Two cube entries can be neighbours.
func lumaSpread() {
	w, h := 114, 28
	hy := int(float64(h) * 0.42)
	fmt.Printf("\n%-8s %9s %9s %9s %9s   %s\n", "hour", "dimmest", "brightest", "spread", "vs sky", "verdict")
	for _, hr := range []float64{0, 9, 12.6, 18, 22} {
		sh := scape.NewShore(7, false)
		var c *canvas.Canvas
		tm := 0.0
		for k := 0; k < 200; k++ {
			c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: 0.26, TimeOfDay: hr / 24,
				TodoDone: 19, TodoTotal: 32,
			})
		}
		lo, hi, sky := 1e9, -1e9, 0.0
		n := 0
		for y := 0; y < hy; y++ {
			for x := 0; x < w; x++ {
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				if r != '*' {
					continue
				}
				l := luma(fg)
				if l < lo {
					lo = l
				}
				if l > hi {
					hi = l
				}
				sky += luma(bg)
				n++
			}
		}
		if n == 0 {
			continue
		}
		sky /= float64(n)
		spread := hi - lo
		// A cube step is worth about 10 luma, and the eye needs a few of them
		// side by side to call two marks different brightnesses.
		verdict := "visible"
		switch {
		case spread < 12:
			verdict = "TOO CLOSE TO SEE -- about one cube step"
		case spread < 30:
			verdict = "subtle"
		}
		fmt.Printf("%-8.1f %9.1f %9.1f %9.1f %9.1f   %s\n", hr, lo, hi, spread, hi-sky, verdict)
	}
}

func luma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}
