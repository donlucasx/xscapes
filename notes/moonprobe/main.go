// moonprobe prints the cells around the moon as a 256-colour terminal shows
// them, with the palette paths on and off, so a report about the moon can be
// read cell by cell instead of argued about.
package main

import (
	"flag"
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	w := flag.Int("w", 124, "scape width")
	h := flag.Int("h", 27, "scape height")
	tod := flag.Float64("tod", 0.0245, "time of day")
	ctx := flag.Float64("ctx", 0.06, "context used")
	flag.Parse()
	for _, ramps := range []bool{false, true} {
		term.Ramps = ramps
		c := canvas.New(*w, *h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: *tod, ContextUsed: *ctx}
		for k := 0; k < 14; k++ {
			sh.Update(c, 3.0+float64(k)/20, act)
		}
		mx, my := sh.MoonPos()
		fmt.Printf("\n== ramps=%v  moon centre (%d,%d)  each cell: glyph fg/bg as 256 shows them; '.' = space ==\n", ramps, mx, my)
		for y := my - 4; y <= my+4; y++ {
			if y < 0 || y >= c.H {
				continue
			}
			fmt.Printf("row %2d: ", y)
			for x := mx - 7; x <= mx+7; x++ {
				if x < 0 || x >= c.W {
					continue
				}
				ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
				g := string(ch)
				if ch == ' ' {
					g = "."
				}
				if fg == bg || ch == ' ' {
					fmt.Printf(" %s %02x%02x%02x      ", g, bg.R, bg.G, bg.B)
				} else {
					fmt.Printf(" %s %02x%02x%02x/%02x%02x%02x", g, fg.R, fg.G, fg.B, bg.R, bg.G, bg.B)
				}
			}
			fmt.Println()
		}
		fmt.Println("   true bg under the same cells (what truecolor paints):")
		for y := my - 4; y <= my+4; y++ {
			if y < 0 || y >= c.H {
				continue
			}
			fmt.Printf("row %2d: ", y)
			for x := mx - 7; x <= mx+7; x++ {
				if x < 0 || x >= c.W {
					continue
				}
				bg := c.BGAt(x, y)
				fmt.Printf(" %02x%02x%02x", bg.R, bg.G, bg.B)
			}
			fmt.Println()
		}
	}
}
