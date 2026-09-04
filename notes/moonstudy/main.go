// moonstudy renders the moon's edge three ways, at three hours, cropped to
// the cells around the disc, as a 256-colour terminal shows them -- so the
// question "why not the mockup's soft moon" can be answered by looking.
package main

import (
	"flag"
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	out := flag.String("html", "assets/frames/moonstudy.html", "write the page here")
	w := flag.Int("w", 133, "scape width")
	h := flag.Int("h", 27, "scape height")
	ctx := flag.Float64("ctx", 0.14, "context used")
	flag.Parse()

	hours := []struct {
		name string
		tod  float64
	}{{"night, 00:35", 0.0245}, {"morning, 10:48", 0.45}, {"dusk, 18:00", 0.75}}
	modes := []struct{ key, name, note string }{
		{"", "solid, what ships", "one tone, a clean edge; the shape comes from half-row sampling"},
		{"hue", "a rim in the moon's own hue", "the outer ring one tone darker, hue kept by the background quantiser"},
		{"blend", "the mockup's fade", "the outer ring blended half toward the sky, as in s7"},
	}
	var b strings.Builder
	b.WriteString(`<title>The Moon's Edge</title><style>
:root{--bg:#0b0b10;--fg:#c9c9d4;--dim:#8a8a96;--line:#1c1c24;--hi:#e8c27a}
body{margin:0;background:var(--bg);color:var(--fg);font:14px/1.5 -apple-system,Helvetica,Arial,sans-serif;padding:28px}
h1{font-size:20px;margin:0 0 6px;font-weight:600}p{max-width:78ch;color:var(--dim);margin:0 0 18px}
table{border-collapse:separate;border-spacing:0 18px}th{font-size:10px;letter-spacing:.1em;text-transform:uppercase;color:#55555f;font-weight:600;text-align:left;padding:0 14px 0 0}
td{padding:0 14px 0 0;vertical-align:top}td.h{font-size:13px;color:var(--fg);white-space:nowrap;padding-top:8px}
.win{display:inline-block;background:#000;border:1px solid #22222a;border-radius:4px;padding:4px 6px}
pre{margin:0;font:16px/1.0 Menlo,monospace;letter-spacing:0}
.m{font-size:12px;color:var(--dim);max-width:34ch;margin-top:6px}.m b{color:var(--hi);font-weight:600}
</style>
<h1>The moon's edge, three ways, as Terminal.app shows it</h1>
<p>Each cell is one terminal cell at the scape's 133x27, cropped to the sky around the disc, quantised to the 256 palette the way the renderer does it. The mockup's soft moon was truecolor; a 256 terminal has no tone between the disc and the sky except the ones its palette holds, and the fade's blend cells land wherever the nearest palette entry is, which by day is a grey. Below each frame: the tones the edge actually uses.</p>
<table><tr><th></th>`)
	for _, m := range modes {
		b.WriteString(`<th>` + html.EscapeString(m.name) + `</th>`)
	}
	b.WriteString(`</tr>`)
	for _, hr := range hours {
		b.WriteString(`<tr><td class="h">` + html.EscapeString(hr.name) + `</td>`)
		for _, m := range modes {
			c := canvas.New(*w, *h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh := scape.NewShore(7, false)
			sh.MoonX = 0.28
			sh.MoonRim = m.key
			for k := 0; k < 8; k++ {
				sh.Update(c, 3.0+float64(k)/20, scape.Activity{Working: true, Level: 0.5, TimeOfDay: hr.tod, ContextUsed: *ctx})
			}
			mx, my := sh.MoonPos()
			tones := map[term.RGB]bool{}
			var pre strings.Builder
			for y := my - 4; y <= my+4; y++ {
				for x := mx - 12; x <= mx+12; x++ {
					if x < 0 || y < 0 || x >= c.W || y >= c.H {
						pre.WriteString(" ")
						continue
					}
					ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
					if x >= mx-6 && x <= mx+6 && y >= my-3 && y <= my+3 {
						tones[bg] = true
						if ch == '▀' {
							tones[fg] = true
						}
					}
					fmt.Fprintf(&pre, `<span style="color:#%02x%02x%02x;background:#%02x%02x%02x">%s</span>`,
						fg.R, fg.G, fg.B, bg.R, bg.G, bg.B, html.EscapeString(string(ch)))
				}
				pre.WriteString("\n")
			}
			var sw strings.Builder
			for t := range tones {
				fmt.Fprintf(&sw, `<span style="display:inline-block;width:12px;height:12px;border:1px solid #333;border-radius:2px;vertical-align:middle;margin-right:3px;background:#%02x%02x%02x"></span>`, t.R, t.G, t.B)
			}
			b.WriteString(`<td><div class="win"><pre>` + pre.String() + `</pre></div><div class="m">` + sw.String() + `<br>` + html.EscapeString(m.note) + `</div></td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</table>`)
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
