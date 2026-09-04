// sunprobe prints the cells around the sun as a truecolor terminal and a
// 256-colour terminal each show them, so a report that the sun differs between
// Ghostty and Terminal.app can be read cell by cell.
package main

import (
	"flag"
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	w := flag.Int("w", 120, "scape width")
	h := flag.Int("h", 26, "scape height")
	tod := flag.Float64("tod", 0.5037, "time of day")
	ctx := flag.Float64("ctx", 0.15, "context used")
	flag.Parse()
	c := canvas.New(*w, *h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(7, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: *tod, ContextUsed: *ctx}
	for k := 0; k < 14; k++ {
		sh.Update(c, 3.0+float64(k)/20, act)
	}
	mx, my := sh.MoonPos()
	fmt.Printf("size %dx%d tod %.4f ctx %.2f  sun centre (%d,%d)\n", *w, *h, *tod, *ctx, mx, my)
	for _, p := range []term.Profile{term.ProfileTrueColor, term.Profile256} {
		fmt.Printf("\n== %s: glyph fg/bg ('.' = space, same colour shown once) ==\n", p)
		for y := my - 5; y <= my+5; y++ {
			if y < 0 || y >= c.H {
				continue
			}
			fmt.Printf("row %2d: ", y)
			for x := mx - 6; x <= mx+6; x++ {
				if x < 0 || x >= c.W {
					continue
				}
				ch, fg, bg := c.ResolveAt(x, y, p)
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
	}
	fmt.Println("\n   nominal bg under the same cells (the blend before any quantiser):")
	for y := my - 5; y <= my+5; y++ {
		if y < 0 || y >= c.H {
			continue
		}
		fmt.Printf("row %2d: ", y)
		for x := mx - 6; x <= mx+6; x++ {
			if x < 0 || x >= c.W {
				continue
			}
			bg := c.BGAt(x, y)
			fmt.Printf(" %02x%02x%02x", bg.R, bg.G, bg.B)
		}
		fmt.Println()
	}
}
