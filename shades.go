package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// runShades prints one frame three ways, stacked, in THIS terminal.
//
// The smoothing question cannot be settled in a browser. Shade blocks are dot
// patterns drawn by the terminal's own font at the terminal's own size, and
// whether they read as a tone or as dirt depends entirely on that -- a preview
// at 11px in Chrome is a different instrument. So this prints the same frame
// with no smoothing, with cells split, and with the shade blocks on, one under
// the other, and the answer is whichever one looks like a sky.
func runShades(args []string) {
	fs := flag.NewFlagSet("shades", flag.ExitOnError)
	tod := fs.Float64("tod", 0.5, "time of day: 0 midnight, .25 dawn, .5 noon, .75 dusk")
	w := fs.Int("w", 64, "width")
	h := fs.Int("h", 14, "height of each panel")
	seed := fs.Int64("seed", 7, "scene seed")
	fs.Parse(args)

	if term.DetectProfile() != term.Profile256 {
		fmt.Fprintf(os.Stderr,
			"note: this terminal reports %s, where all three panels are identical.\n"+
				"      run it with ASCIISCAPES_COLOR=256 to see what Terminal.app gets.\n\n",
			term.DetectProfile())
	}

	tail := []reduce.Line{
		{Text: "read   internal/auth/handler.go  142 lines", Age: 0.75},
		{Text: "edit   internal/auth/handler.go  +18 -2", Age: 0.5},
		{Text: "shell  go test ./...  4.1s", Age: 0.25},
		{Text: "grep   rate.Limiter  3 files", Age: 0.0},
	}

	panels := []struct {
		name       string
		split, dot bool
		why        string
	}{
		{"no smoothing", false, false, "one colour per cell, the palette straight onto the cube"},
		{"split cells", true, false, "U+2580 so a band edge can land mid-cell -- the default"},
		{"split + shade blocks", true, true, "U+2591/2/3 blending two cube colours -- more tones, and I think it reads as stipple"},
	}

	for i, p := range panels {
		term.Shading, term.ShadeBlocks = p.split, p.dot
		c := canvas.New(*w, *h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(*seed, false)
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(*w, ccw, true)
		sh.MoonX = lay.MoonX
		// Same wave phase in all three, or the panels differ in the water as
		// well as the colour and cannot be compared for colour.
		for k := 0; k < 12; k++ {
			sh.Update(c, 3.0+float64(k)/20, scape.Activity{
				Working: true, Level: 0.55, TimeOfDay: *tod, ContextUsed: 0.3})
		}
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: tail}, 3.7, *seed, c.H-2-chh)

		fmt.Printf("\n\x1b[1m%d. %s\x1b[0m  \x1b[2m%s\x1b[0m\n", i+1, p.name, p.why)
		fmt.Println(c.Render(term.DetectProfile()))
	}
	term.Shading, term.ShadeBlocks = true, false

	fmt.Printf("\n\x1b[2mSame frame, same wave phase, %02d:00. Only the smoothing differs.\n"+
		"If 3 looks better than 2 in THIS terminal, say so -- ASCIISCAPES_SHADE_BLOCKS=1 makes it the default.\n"+
		"xscapes shades -tod 0.75 for dusk, -tod 0 for midnight, -h 20 for a taller panel.\x1b[0m\n", int(*tod*24+0.5)%24)
}
