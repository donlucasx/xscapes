// haloprobe: does Shore.MoonHalo change any cell? Prints the background and
// the resolved 256 tone on the moon's row, with the switch off and on.
package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	for _, halo := range []bool{false, true} {
		c := canvas.New(130, 22, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		sh.MoonEdge = "quad"
		sh.MoonHalo = halo
		act := scape.Activity{Working: true, Level: 0.65, ContextUsed: 0.05, TimeOfDay: 0.931}
		sh.Update(c, 2, act)
		mx, my := sh.MoonPos()
		rx, ry := sh.MoonExtent()
		fmt.Printf("halo=%v moon at (%d,%d) extent %dx%d\n", halo, mx, my, rx, ry)
		for _, dy := range []int{-2, 0, 2, 4} {
			if my+dy < 0 {
				continue
			}
			fmt.Printf("  row %+d:", dy)
			for dx := 0; dx <= rx+6; dx++ {
				bg := c.BGAt(mx+dx, my+dy)
				_, _, q := c.ResolveAt(mx+dx, my+dy, term.Profile256)
				fmt.Printf(" %02x%02x%02x/%02x", bg.R, bg.G, bg.B, q.R)
			}
			fmt.Println()
		}
	}
}
