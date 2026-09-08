package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func luma(c term.RGB) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}

// rampProbe walks the sky down a column and reports where its brightness turns
// back on itself. A single row that is darker than BOTH its neighbours reads as
// a rule across the frame -- which is what he has been calling a hairline, and
// it is the ramp's own banding rather than any rendering artifact.
func rampProbe(seed int64) {
	rampSweep(seed)
	fmt.Println()
}

// rampSweep counts, across the whole day, how often the sky or sea contains a
// row that turns the gradient back on itself.
func rampSweep(seed int64) {
	term.LowerHalf = term.DetectSplit("Apple_Terminal")
	term.NoSplitCells = term.DetectNoSplit("Apple_Terminal")
	w := 126
	bad, total := 0, 0
	worst, worstTod, worstH := 0, 0.0, 0
	for _, h := range []int{20, 24, 26, 28, 30, 32, 34, 36, 40, 44, 47, 50} {
	for i := 0; i < 48; i++ {
		tod := float64(i) / 48
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3}
		for k := 0; k < 12; k++ {
			sh.Update(c, 3+float64(k)/40, act)
		}
		var ls []float64
		for y := 0; y < sh.SandTop(); y++ {
			_, _, bg := c.ResolveAt(3, y, term.Profile256)
			ls = append(ls, luma(bg))
		}
		n := 0
		for y := 1; y < len(ls)-1; y++ {
			if (ls[y] < ls[y-1] && ls[y] < ls[y+1]) || (ls[y] > ls[y-1] && ls[y] > ls[y+1]) {
				n++
			}
		}
		total++
		if n > 0 {
			bad++
		}
		if n > worst {
			worst, worstTod, worstH = n, tod, h
		}
	}
	}
	fmt.Printf("SWEEP: 48 half-hours x 12 scape heights at 126 wide\n")
	fmt.Printf("  %d of %d frames contain a row that reverses the gradient (%.0f%%)\n", bad, total, float64(bad)/float64(total)*100)
	fmt.Printf("  worst: %d such rows at %d rows tall, tod=%.3f (%02d:%02d)\n", worst, worstH, worstTod, int(worstTod*24), int(worstTod*24*60)%60)
}

func rampDetail(seed int64) {
	term.LowerHalf = term.DetectSplit("Apple_Terminal")
	term.NoSplitCells = term.DetectNoSplit("Apple_Terminal")
	w, h := 126, 47
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: 0.778, ContextUsed: 0.3}
	for k := 0; k < 12; k++ {
		sh.Update(c, 3+float64(k)/40, act)
	}
	col := 3 // open sky, clear of the disc and the companion
	var ls []float64
	var cs []term.RGB
	for y := 0; y < sh.SandTop(); y++ {
		_, _, bg := c.ResolveAt(col, y, term.Profile256)
		cs = append(cs, bg)
		ls = append(ls, luma(bg))
	}
	fmt.Printf("sky and sea, column %d, rows 0..%d\n\n", col, len(ls)-1)
	dips := 0
	for y := range ls {
		mark := ""
		if y > 0 && y < len(ls)-1 && ls[y] < ls[y-1] && ls[y] < ls[y+1] {
			mark = "   <-- DARKER THAN BOTH NEIGHBOURS: reads as a rule"
			dips++
		}
		if y > 0 && y < len(ls)-1 && ls[y] > ls[y-1] && ls[y] > ls[y+1] {
			mark = "   <-- lighter than both: reads as a rule"
			dips++
		}
		fmt.Printf("row %2d  %3d,%3d,%3d  luma %6.1f%s\n", y, cs[y].R, cs[y].G, cs[y].B, ls[y], mark)
	}
	fmt.Printf("\n%d rows turn the gradient back on itself.\n", dips)
}
