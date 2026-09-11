// Command stargrab answers his 2026-09-10 question -- "heres a screengrab of
// how the stars are showing up on the constellation, is this as expected?" --
// by rendering the constellation at HIS geometry and counting what lands where.
//
// His window is 125x62. `xscapes claude` gives the scape a band, not the whole
// window, so the scape is 125 wide by whatever host.Band allows (capped at
// MaxScapeRows). The screenshot's scape is about 28 rows.
//
//	go run ./notes/s28-stargrab
package main

import (
	"fmt"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	const w, h = 125, 28
	fmt.Printf("his window 125x62, scape band ~%dx%d, StarsCap %d\n\n", w, h, 32)

	for _, done := range []int{12, 16, 19, 24, 32} {
		sh := scape.NewShore(7, false)
		var c *canvas.Canvas
		tm := 0.0
		for k := 0; k < 120; k++ {
			c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: 0.58,
				TimeOfDay: 19.8 / 24, // 7:50 PM, his clock
				TodoDone:  done, TodoTotal: 32,
			})
		}
		var cols, rows []int
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
					cols = append(cols, x)
					rows = append(rows, y)
				}
			}
		}
		sort.Ints(cols)
		sort.Ints(rows)
		if len(cols) == 0 {
			fmt.Printf("done=%-3d NO STARS DRAWN\n", done)
			continue
		}
		// How evenly are they spread? Largest gap between neighbours, against
		// the average gap. A pile in one corner shows up here as a huge ratio.
		gap, mean := 0, float64(cols[len(cols)-1]-cols[0])/float64(len(cols))
		for i := 1; i < len(cols); i++ {
			if d := cols[i] - cols[i-1]; d > gap {
				gap = d
			}
		}
		half := 0
		for _, x := range cols {
			if x > w/2 {
				half++
			}
		}
		fmt.Printf("done=%-3d drawn=%-3d cols %d..%d  rows %d..%d  widest gap %d (mean %.1f)  %d of %d past halfway\n",
			done, len(cols), cols[0], cols[len(cols)-1], rows[0], rows[len(rows)-1], gap, mean, half, len(cols))
	}

	fmt.Println("\nthe sky at 19 lit, which is about what his screengrab shows:")
	sh := scape.NewShore(7, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 120; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.58, TimeOfDay: 19.8 / 24,
			TodoDone: 19, TodoTotal: 32,
		})
	}
	for y := 0; y < 12; y++ {
		line := ""
		for x := 0; x < w; x++ {
			r, _, _ := c.ResolveAt(x, y, term.Profile256)
			if r == '*' {
				line += "*"
			} else {
				line += " "
			}
		}
		fmt.Printf("%2d |%s|\n", y, line)
	}
}
