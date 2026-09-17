package main

import (
	"os"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
)

// TestOGImagePage writes the page a share card is screenshotted from: the
// live mountain vista at dusk, 80x25 cells at 25px, which is 1200x625 in
// Menlo, the owl working with two owlets and three stars, the last tool
// calls on the band. His ask of 2026-09-17: a thumbnail when the entry is
// shared, "possibly of the mountainscape".
//
//	XSCAPES_OGPAGE=<file>.html go test -run TestOGImagePage .
//	"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new \
//	  --window-size=1200,630 --screenshot=site/og.png "file://<file>.html"
func TestOGImagePage(t *testing.T) {
	out := os.Getenv("XSCAPES_OGPAGE")
	if out == "" {
		t.Skip("set XSCAPES_OGPAGE=<file> to write the share card page")
	}
	c := canvas.New(80, 25, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	lay := compose(c.W, 12, true)
	v := scenes.NewVista(7, false)
	v.OwlX = lay.CatX
	act := scape.Activity{Working: true, Level: 0.6, ContextUsed: 0.35, TimeOfDay: 0.78, TodoDone: 3}
	st := reduce.State{Act: act, Pose: companion.Working, Kittens: 2, Tail: []reduce.Line{
		{Text: "read  internal/auth/handler.go  142 lines", Age: 0.6},
		{Text: "edit  internal/auth/handler.go  +18 -2", Age: 0.3},
		{Text: "write  internal/auth/limiter.go  64 lines", Age: 0},
	}}
	// Twice: the first draw syncs the owlets' arrival, the second, eight
	// seconds on, has them landed beside the owl.
	v.Update(c, 0, act)
	drawVistaWith(c, v, lay, st, 0, nil)
	v.Update(c, 8, act)
	drawVistaWith(c, v, lay, st, 8, nil)
	page := `<!doctype html><meta charset="utf-8"><style>html,body{margin:0;background:#000}` +
		`pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0;white-space:pre}</style>` +
		c.HTMLFragment(25)
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
}
