package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/term"
)

// probeCols prints what each eye colour actually becomes through the glyph
// path, so the page can be checked for them by the colour that LANDS.
func probeCols() {
	for _, c := range []struct {
		n string
		c term.RGB
	}{{"parent eye", eyeShine}, {"crablet eye", crabletEyeCol}, {"alert", eyeAlert}, {"worried amber", eyeWorried}, {"salmon coat", coats[0].col}} {
		got := term.Profile256.Quantise(c.c, true)
		fmt.Printf("%-14s %3d,%3d,%3d -> %3d,%3d,%3d  #%02x%02x%02x  idx %d\n",
			c.n, c.c.R, c.c.G, c.c.B, got.R, got.G, got.B, got.R, got.G, got.B, got.Index256())
	}
}
