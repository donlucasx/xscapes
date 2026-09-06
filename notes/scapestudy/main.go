// scapestudy renders every proposed scene and companion through the product's
// own pipeline, as a 256-colour terminal shows them, onto one page -- so the
// question "which of these, if any" can be answered by looking rather than
// argued from descriptions. His ask, 2026-09-05: at least five of each,
// real frames, before anything is chosen.
package main

import (
	"flag"
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	out := flag.String("html", "assets/frames/scapestudy.html", "write the page here")
	px := flag.Int("px", 9, "cell font size in the page")
	seed := flag.Int64("seed", 7, "scene seed")
	level := flag.Float64("level", 0.5, "the work, 0..1, carried by each scene's motion slot")
	t := flag.Float64("t", 3.0, "the frame's time, seconds")
	flag.Parse()

	hours := []struct {
		name string
		tod  float64
	}{{"night", 0.0245}, {"noon", 0.5}, {"dusk", 0.75}}

	var b strings.Builder
	b.WriteString(`<meta charset="utf-8"><title>Five Scenes, Five Companions</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&family=Geist+Mono:wght@700&display=swap">
<style>
:root{--g:#eeeeee;--s:#e4e4e4;--ln:#d0d0d0;--d:#626262;--i:#121212;--acc:#005f87}
@media (prefers-color-scheme:dark){:root:not([data-theme="light"]){--g:#121212;--s:#1c1c1c;--ln:#303030;--d:#8a8a8a;--i:#eeeeee;--acc:#87afff}}
:root[data-theme="dark"]{--g:#121212;--s:#1c1c1c;--ln:#303030;--d:#8a8a8a;--i:#eeeeee;--acc:#87afff}
body{background:var(--g);color:var(--i);font:15px/1.5 Geist,-apple-system,Helvetica,Arial,sans-serif;margin:0;padding:40px 28px 80px}
.wrap{max-width:1180px;margin:0 auto}
.lockup{font:700 22px/1.2 "Geist Mono",Menlo,monospace;letter-spacing:0}
.blk{display:inline-block;width:1ch;background:var(--i);color:var(--g)}
h1{font-size:30px;line-height:1.15;font-weight:600;margin:18px 0 8px;text-wrap:balance}
h2{font-size:12px;letter-spacing:.12em;text-transform:uppercase;color:var(--d);font-weight:600;margin:56px 0 18px;padding-top:18px;border-top:1px solid var(--ln)}
p{max-width:66ch;color:var(--d);margin:0 0 12px}
.blk-sec{margin:0 0 40px}
.head{display:flex;gap:24px;align-items:baseline;flex-wrap:wrap;margin:0 0 6px}
.nm{font-size:20px;font-weight:600;color:var(--i)}
.coat{font:700 12px/1 "Geist Mono",Menlo,monospace;color:var(--acc)}
.slots{display:grid;grid-template-columns:max-content 1fr;gap:2px 14px;font-size:13px;color:var(--d);margin:8px 0 14px;max-width:66ch}
.slots b{color:var(--i);font-weight:500}
.frames{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start}
.win{background:#000;border:1px solid var(--ln);border-radius:3px;padding:6px 8px;overflow-x:auto;max-width:100%}
.win pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1.0;letter-spacing:0}
.cap{font-size:11px;letter-spacing:.08em;text-transform:uppercase;color:var(--d);margin:0 0 6px}
</style>
<div class="wrap">
<div class="lockup"><span class="blk">x</span>scapes</div>
<h1>Five scenes and five companions, as a 256-colour terminal shows them</h1>
<p>Every frame below is rendered through the product's own canvas, palette rules and sprite pipeline at the design size of 80 by 24, quantised the way Terminal.app will show it. Nothing is chosen. The shore stays as it ships; these are the candidates from the session 19 research, each drawn once so the choice can be made by looking.</p>
<p>Each scene fills the six slots the brief requires. The motion slot carries the work at a middle level; the sky slot carries the real clock, shown at night, noon and dusk. The writing band is the same three lines in every scene, on whatever surface the scene offers.</p>
<h2>Five scenes</h2>`)
	for _, sc := range scenes {
		b.WriteString(`<div class="blk-sec"><div class="head"><span class="nm">` + html.EscapeString(sc.name) + `</span></div>`)
		b.WriteString(`<p>` + html.EscapeString(sc.note) + `</p><div class="slots">`)
		for i, k := range []string{"light", "sky", "motion", "surface", "accumulator", "companion"} {
			b.WriteString(`<b>` + k + `</b><span>` + html.EscapeString(sc.slots[i]) + `</span>`)
		}
		b.WriteString(`</div><div class="frames">`)
		for _, hr := range hours {
			c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			c.Clear()
			sc.paint(c, hr.tod, *t, *level, *seed)
			b.WriteString(`<div><div class="cap">` + hr.name + `</div><div class="win">` + c.HTMLFragmentAs(*px, term.Profile256) + `</div></div>`)
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`<h2>Five companions, beside the cat</h2>
<p>Each animal sits where the cat sits, at the cat's size, with two of its young where the kittens sit. Same pipeline: one coat colour from the locked set, quadrant glyphs, a one-cell rim, eyes plotted into gaps the bitmap leaves. The cat first, as the reference.</p>`)
	two := []struct {
		name string
		tod  float64
	}{{"night", 0.0245}, {"noon", 0.5}}
	b.WriteString(`<div class="blk-sec"><div class="head"><span class="nm">Cat</span><span class="coat">cream, as shipped</span></div><p>The companion that ships, with two kittens, for scale and for the style every candidate has to sit beside.</p><div class="frames">`)
	for _, hr := range two {
		c := catFrame(hr.tod, *t, *seed)
		b.WriteString(`<div><div class="cap">` + hr.name + `</div><div class="win">` + c.HTMLFragmentAs(*px, term.Profile256) + `</div></div>`)
	}
	b.WriteString(`</div></div>`)
	for i := range animals {
		a := &animals[i]
		b.WriteString(`<div class="blk-sec"><div class="head"><span class="nm">` + html.EscapeString(a.name) + `</span><span class="coat">` + html.EscapeString(a.coat) + `</span></div>`)
		b.WriteString(`<p>` + html.EscapeString(a.note) + `</p><div class="frames">`)
		for _, hr := range two {
			c := companionFrame(a, hr.tod, *t, *seed)
			b.WriteString(`<div><div class="cap">` + hr.name + `</div><div class="win">` + c.HTMLFragmentAs(*px, term.Profile256) + `</div></div>`)
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`<p style="margin-top:40px">Rendered by <code>go run ./notes/scapestudy</code>. Colours are cube entries by construction; a tone that is not one panics the study.</p></div>`)
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
