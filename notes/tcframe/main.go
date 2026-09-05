// tcframe writes one frame of the demo scene in truecolor and in 256, side by
// side, to judge the page's clips: his 2026-09-05 note, "we should show
// truecolor grabs, not 256. Aim for optimal depictions".
package main

import (
	"fmt"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	var b string
	b += `<!doctype html><html><head><meta charset="utf-8"><style>body{background:#000;margin:0}pre{margin:0;font:12px/1 Menlo,monospace;display:inline-block}.r{display:flex;gap:8px;margin:8px}</style></head><body>`
	for _, tc := range []struct {
		name string
		tod  float64
		used float64
	}{{"noon", 0.52, 0.10}, {"dusk", 0.80, 0.46}, {"night", 0.96, 0.46}} {
		c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: true, Level: 0.6, ContextUsed: tc.used, TimeOfDay: tc.tod}
		for i := 0; i < 30; i++ {
			sh.Update(c, 2+float64(i)/20, act)
		}
		b += `<div class="r">` + c.HTMLFragment(12) + c.HTMLFragmentAs(12, term.Profile256) + `</div>`
		fmt.Fprintln(os.Stderr, tc.name)
	}
	b += `</body></html>`
	os.Stdout.WriteString(b)
}
