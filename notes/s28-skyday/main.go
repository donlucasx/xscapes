// Command skyday answers his 2026-09-10 note -- "looking at this current
// session as it gets darker, I see more characters varying in sizes and shapes
// appearing. Thats what I was talking about ... this felt too bare, how can we
// dial it for daytime?"
//
// He is comparing two shots of the SAME sky an hour apart: 20:22 reads bare,
// 20:55 reads rich. What changed is StarVis, the AMBIENT field's day/night
// visibility, which is a WORLD channel -- it is the real clock and nothing else.
//
//	go run ./notes/s28-skyday
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	w, h := 125, 28
	hy := int(float64(h) * 0.42)
	fmt.Printf("his scape band %dx%d, horizon row %d, %d rows of sky\n\n", w, h, hy, hy)
	fmt.Printf("%-7s %8s %9s %9s %9s %9s   %s\n",
		"hour", "StarVis", "ambient", "constell", "distinct", "total", "how it reads")
	for _, hr := range []float64{0, 4, 6, 7, 8, 9, 10, 12, 14, 16, 17, 18, 19, 20, 20.4, 20.9, 21, 22, 23} {
		amb, con, glyphs := count(w, h, hy, hr/24, 19)
		tot := amb + con
		read := "rich"
		switch {
		case tot < 8:
			read = "BARE  <- his complaint"
		case tot < 20:
			read = "thin"
		}
		fmt.Printf("%-7.1f %8.3f %9d %9d %9d %9d   %s\n",
			hr, scape.PaletteAt(hr/24).StarVis, amb, con, len(glyphs), tot, read)
	}
	fmt.Println("\nThe constellation is FLAT across the day by design -- todoStarFloor 0.85 keeps")
	fmt.Println("it legible at every hour, because a finished unit of work is a fact about the")
	fmt.Println("AGENT and must not be switched off by the clock. Everything that varies is the")
	fmt.Println("ambient field, and its whole range is spent between about 17:00 and 06:00.")
	drawnVsSeen()
	fmt.Println("\nHis 20:22 shot and his 20:55 shot are the two rows marked 20.4 and 20.9.")
}

func count(w, h, hy int, tod float64, done int) (amb, con int, glyphs map[rune]int) {
	sh := scape.NewShore(7, false)
	var c *canvas.Canvas
	tm := 0.0
	for k := 0; k < 120; k++ {
		c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		tm += 0.08
		sh.Update(c, tm, scape.Activity{
			Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: tod,
			TodoDone: done, TodoTotal: 32,
		})
	}
	glyphs = map[rune]int{}
	for y := 0; y < hy; y++ {
		for x := 0; x < w; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			if fg == bg {
				continue // invisible against its own ground
			}
			switch r {
			case '*':
				con++
				glyphs[r]++
			case '.', '·', '+':
				amb++
				glyphs[r]++
			}
		}
	}
	return amb, con, glyphs
}

// drawnVsSeen separates "plotted" from "legible". The ambient field encodes the
// clock in ALPHA -- far.Plot(x, y, g, Star, twinkle) -- so as StarVis falls the
// specks are still drawn, they just quantise into the sky behind them. That is
// the difference between his two screenshots, and it is not a count.
func drawnVsSeen() {
	w, h := 125, 28
	hy := int(float64(h) * 0.42)
	fmt.Printf("\n%-7s %8s %9s %9s %8s\n", "hour", "StarVis", "plotted", "legible", "lost")
	for _, hr := range []float64{12, 16, 18, 20.4, 20.9, 22, 0} {
		sh := scape.NewShore(7, false)
		var c *canvas.Canvas
		tm := 0.0
		for k := 0; k < 120; k++ {
			c = canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tm += 0.08
			sh.Update(c, tm, scape.Activity{
				Level: 0.5, Working: true, ContextUsed: 0.56, TimeOfDay: hr / 24,
				TodoDone: 19, TodoTotal: 32,
			})
		}
		plotted, seen := 0, 0
		for y := 0; y < hy; y++ {
			for x := 0; x < w; x++ {
				cell := c.Far().Cells[y*w+x]
				if !cell.Set {
					continue
				}
				switch cell.R {
				case '.', '·', '+':
					plotted++
					if r, fg, bg := c.ResolveAt(x, y, term.Profile256); fg != bg && r == cell.R {
						seen++
					}
				}
			}
		}
		lost := 0
		if plotted > 0 {
			lost = 100 * (plotted - seen) / plotted
		}
		fmt.Printf("%-7.1f %8.3f %9d %9d %7d%%\n", hr, scape.PaletteAt(hr/24).StarVis, plotted, seen, lost)
	}
}
