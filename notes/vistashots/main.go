// vistashots renders the mountain vista and its owl as standalone frame
// pages, so a REAL look is one command away: every defect found in session
// 32 (the lake as a stripe, the treeline as a skyline, the owl's eye row lost
// to quarter-cells, the post that did not read, the rain falling upwards)
// was seen in a screenshot and not in a text dump.
//
//	go run ./notes/vistashots <dir>
//	for f in <dir>/*.html; do "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
//	  --headless=new --disable-gpu --use-mock-keychain --hide-scrollbars \
//	  --window-size=1400,1200 --screenshot="${f%.html}.png" "file://$f"; done
//
// Then Read the PNGs. Profile256 throughout: what a terminal draws, not the
// page's truecolor.
package main

import (
	"fmt"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

const style = `<meta charset="utf-8"><style>html,body{margin:0;padding:12px;background:#000;color:#ccc;font:13px Menlo,monospace}pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}.row{display:flex;gap:10px;align-items:flex-start;margin-bottom:10px}.lab{width:150px;padding-top:50px;font-weight:bold}h2{font:bold 14px Menlo,monospace;color:#eee;margin:18px 0 8px;border-top:1px solid #333;padding-top:10px}</style>`

func vista(tod, level float64) *canvas.Canvas {
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	c.Clear()
	scenes.Forest[0].Paint(c, tod, 1.0, level, 7)
	return c
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vistashots <dir>")
		os.Exit(1)
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	pick, place, olet := scenes.OwlPick, scenes.OwlPlace, scenes.OwletPick

	// 1. The vista as it ships, three hours, idle and working hard.
	b := style + `<h2>the vista as it ships: noon, dusk, night; idle then working hard</h2>`
	for _, lv := range []float64{0.1, 0.85} {
		b += `<div class="row"><div class="lab">level ` + fmt.Sprint(lv) + `</div>`
		for _, tod := range []float64{0.5, 0.78, 0.0245} {
			b += `<div>` + vista(tod, lv).HTMLFragmentAs(8, term.Profile256) + `</div>`
		}
		b += `</div>`
	}
	os.WriteFile(dir+"/vista.html", []byte(b), 0o644)

	// 2. Every owl on every perch, with the picked owlets, at dusk.
	b = style + `<h2>every owl on every perch, at dusk (rows: owls; columns: mound, upper branch, lower branch, stump, fence)</h2>`
	for i := range scenes.OwlAlts {
		scenes.OwlPick = i
		b += `<div class="row"><div class="lab">` + scenes.OwlAlts[i].Name + `</div>`
		for p := 0; p <= 4; p++ {
			scenes.OwlPlace = p
			b += `<div>` + vista(0.78, 0.3).HTMLFragmentCropAs(36, 0, 80, 22, 6, term.Profile256) + `</div>`
		}
		b += `</div>`
	}
	scenes.OwlPick, scenes.OwlPlace = pick, place
	os.WriteFile(dir+"/owls.html", []byte(b), 0o644)

	// 3. Every owlet look, on the grass beside the picked owl, at three counts.
	b = style + `<h2>every owlet look, on the picked perch, at one, two and three</h2>`
	for i := range scenes.OwletStyles {
		scenes.OwletPick = i
		b += `<div class="row"><div class="lab">` + scenes.OwletStyles[i].Name + `</div>`
		for n := 1; n <= 3; n++ {
			scenes.OwletCount = n
			b += `<div>` + vista(0.78, 0.3).HTMLFragmentCropAs(40, 8, 80, 22, 9, term.Profile256) + `</div>`
		}
		b += `</div>`
	}
	scenes.OwletPick, scenes.OwletCount = olet, 2
	os.WriteFile(dir+"/owlets.html", []byte(b), 0o644)
	fmt.Println(dir + "/vista.html", dir+"/owls.html", dir+"/owlets.html")
}
