// lineprobe is the instrument for his two reports of 2026-09-06: the thin
// lines on 256, and the sun that "breaks and fixes itself" as the context
// depletes.
//
// It does one thing a plain HTML preview cannot: it draws each cell with
// TERMINAL.APP'S OWN GEOMETRY, measured from his 134x71 frame that afternoon
// (Screenshot 3.56.23 PM, 2x Retina). Scape rows sit on a 30 device-pixel
// grid, and inside a split cell the upper colour runs to y+17, the lower to
// y+29, and y+29 is a blend of the two -- U+2584's ink stops about half a
// logical pixel short of the cell's bottom edge and the background shows
// through. That sliver IS the thin line, and a page that paints a split cell
// as two flat halves cannot show the defect at all.
//
//	go run ./notes/lineprobe -out /tmp/lines.html
//	go run ./notes/lineprobe                        # column report, no page
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// Terminal.app's measured cell geometry, in device pixels of a 30px row.
const (
	cellPx = 30.0
	inkTop = 17.0
	inkBot = 29.4
	upTop  = 5.0 // U+2580's ink starts here, measured 2026-09-05
)

func rgb(c term.RGB) string { return fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B) }

// cellCSS paints one cell the way Terminal.app paints it.
func cellCSS(ch rune, fg, bg term.RGB) string {
	pc := func(v float64) float64 { return v / cellPx * 100 }
	switch ch {
	case '▄': // lower half: bg, ink, then bg again for the last pixel row
		return fmt.Sprintf("background:linear-gradient(to bottom,%s 0 %.2f%%,%s %.2f%% %.2f%%,%s %.2f%% 100%%)",
			rgb(bg), pc(inkTop), rgb(fg), pc(inkTop), pc(inkBot), rgb(bg), pc(inkBot))
	case '▀': // upper half: five pixels of bg before the ink starts
		return fmt.Sprintf("background:linear-gradient(to bottom,%s 0 %.2f%%,%s %.2f%% %.2f%%,%s %.2f%% 100%%)",
			rgb(bg), pc(upTop), rgb(fg), pc(upTop), pc(inkTop), rgb(bg), pc(inkTop))
	}
	return "background:" + rgb(bg)
}

func split(ch rune) bool { return ch == '▄' || ch == '▀' }

// grid renders a canvas crop as Terminal.app would draw it.
func grid(c *canvas.Canvas, x0, y0, x1, y1 int, zoom float64) string {
	var b strings.Builder
	x0, y0 = max(x0, 0), max(y0, 0)
	x1, y1 = min(x1, c.W), min(y1, c.H)
	b.WriteString(`<div class="grid">`)
	for y := y0; y < y1; y++ {
		fmt.Fprintf(&b, `<div class="r" style="height:%.0fpx">`, cellPx*zoom/2)
		for x := x0; x < x1; x++ {
			ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
			glyph := ""
			if ch != ' ' && !split(ch) {
				glyph = fmt.Sprintf(`<span style="color:%s">%s</span>`, rgb(fg), string(ch))
			}
			fmt.Fprintf(&b, `<i style="width:%.0fpx;%s">%s</i>`, 16*zoom/2, cellCSS(ch, fg, bg), glyph)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func frame(cols, rows int, seed int64, tod, used float64, tune func(*scape.Shore)) (*canvas.Canvas, int, int) {
	c := canvas.New(cols, rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	if tune != nil {
		tune(sh)
	}
	act := scape.Activity{TimeOfDay: tod, ContextUsed: used}
	for i := 0; i < 30; i++ {
		sh.Update(c, 0.5+float64(i)/20, act)
	}
	sh.Update(c, 2, act)
	mx, my := sh.MoonPos()
	return c, mx, my
}

type variant struct {
	name string
	tune func(*scape.Shore)
}

func main() {
	cols := flag.Int("cols", 134, "columns")
	rows := flag.Int("rows", 28, "scape rows")
	seed := flag.Int64("seed", 7, "scene seed")
	tod := flag.Float64("tod", 0.80, "time of day")
	out := flag.String("out", "", "write the HTML page here")
	flag.Parse()

	term.LowerHalf = true // Terminal.app

	if *out == "" {
		c, _, _ := frame(*cols, *rows, *seed, *tod, 0.49, nil)
		splits := 0
		for y := 0; y < *rows; y++ {
			ch, fg, bg := c.ResolveAt(10, y, term.Profile256)
			kind := "flat "
			if split(ch) {
				kind, splits = "SPLIT", splits+1
			}
			fmt.Printf("row %2d %s up=%v down=%v\n", y, kind, bg, fg)
		}
		fmt.Printf("split rows in column 10: %d of %d\n", splits, *rows)
		return
	}

	var b strings.Builder
	b.WriteString(`<style>
	body{background:#16161c;color:#d8d8e0;font:13px/1.5 ui-monospace,monospace;margin:24px;max-width:1700px}
	h1{font-size:19px;margin:0 0 4px} h2{font-size:14px;margin:30px 0 6px;color:#d8d8e0}
	p{color:#9a9aa8;max-width:74ch}
	.grid{display:inline-block;border:1px solid #2a2a32;overflow:hidden;vertical-align:top}
	.r{display:flex} .r i{display:block;height:100%;font-style:normal;font-size:9px;text-align:center;line-height:1}
	.row{display:flex;gap:16px;flex-wrap:wrap;align-items:flex-start}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font-size:11px;color:#8a8a99} .lbl b{color:#d8d8e0;font-weight:500}
	</style>`)
	b.WriteString(`<h1>The thin lines, and the sun's shadow</h1>`)
	b.WriteString(`<p>Every cell below is painted with Terminal.app's own measured geometry, not as two flat halves. ` +
		`In a split cell the upper colour runs to 17/30 of the row, the lower to 29.4/30, and the last device pixel ` +
		`row is the upper colour again, because U+2584's ink stops just short of the cell's bottom edge. That sliver ` +
		`is the thin line. It shows wherever the row below carries the split's lower colour, which is every gradient ` +
		`step and every edge inside the disc. Ghostty draws the blocks exact and has none of this.</p>`)

	b.WriteString(`<h2>1 &middot; the sky gradient</h2>`)
	b.WriteString(`<p>Split cells give the sky twice the vertical resolution. Two or three rows a frame are split, ` +
		`and each of them draws one line the whole way across.</p><div class="row">`)
	shipped, _, _ := frame(*cols, *rows, *seed, *tod, 0.49, nil)
	fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>as it ships</b> &mdash; splits on</div>%s</div>`,
		grid(shipped, 0, 0, 26, 13, 3))
	term.Shading = false
	flat, _, _ := frame(*cols, *rows, *seed, *tod, 0.49, nil)
	term.Shading = true
	fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>splits off</b> &mdash; one tone a row</div>%s</div>`,
		grid(flat, 0, 0, 26, 13, 3))
	b.WriteString(`</div>`)

	b.WriteString(`<h2>2 &middot; the disc</h2>`)
	b.WriteString(`<p>The hue rim is sampled at half rows, so a cell whose upper half is rim and lower half is body ` +
		`is a split cell and draws a line across the sun.</p><div class="row">`)
	for _, v := range []variant{
		{"as it ships (hue rim)", nil},
		{"no rim", func(s *scape.Shore) { s.MoonRim = "" }},
		{"quad edge", func(s *scape.Shore) { s.MoonEdge = "quad" }},
	} {
		c, mx, my := frame(*cols, *rows, *seed, *tod, 0.49, v.tune)
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div>%s</div>`,
			v.name, grid(c, mx-11, my-4, mx+11, my+6, 3))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>3 &middot; the sun, as the context fills</h2>`)
	b.WriteString(`<p>lit = 1 &minus; context used. Until 2026-09-06 the unlit face was painted by day as well as by ` +
		`night, so the sun carried a lunar terminator: a slate mass that grew as the context filled and cleared at a ` +
		`compaction &mdash; his "the Sun seems to break sometimes, and fix itself". It now wanes as a crescent; the ` +
		`second row is the old look, kept as the study switch SunShadow = "slate".</p>`)
	for _, v := range []variant{
		{"as it ships (crescent)", nil},
		{"the old slate face", func(s *scape.Shore) { s.SunShadow = "slate" }},
	} {
		fmt.Fprintf(&b, `<div class="lbl" style="margin-top:12px"><b>%s</b></div><div class="row">`, v.name)
		for _, used := range []float64{0.05, 0.30, 0.49, 0.70, 0.90} {
			c, mx, my := frame(*cols, *rows, *seed, *tod, used, v.tune)
			fmt.Fprintf(&b, `<div class="col"><div class="lbl">%.0f%% used</div>%s</div>`,
				used*100, grid(c, mx-9, my-4, mx+9, my+6, 3))
		}
		b.WriteString(`</div>`)
	}

	if err := os.WriteFile(*out, []byte(canvas.HTMLPage("xscapes - thin lines and the sun", b.String())), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
