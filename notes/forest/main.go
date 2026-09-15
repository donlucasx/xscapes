// forest renders the three candidates for the third scape -- a mountain in
// the woods, with the owl -- as a 256-colour terminal shows them, each one
// looping at dusk beside a night and a noon still, onto one page, so the
// question "what carries the work when there is no sea" is answered by
// looking. His ask, 2026-09-15.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"math"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	out := flag.String("html", "forest.html", "write the page here")
	seed := flag.Int64("seed", 7, "scene seed")
	fps := flag.Int("fps", 6, "frames a second in the loop")
	flag.Parse()

	pal := &canvas.HTMLPalette{}
	frame := func(sc *scenes.Scene, tod, t, level float64) string {
		c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Clear()
		sc.Paint(c, tod, t, level, *seed)
		return c.HTMLFragmentClassed(11, term.Profile256, pal)
	}
	n := int(math.Round(scenes.LoopSecs * float64(*fps)))
	type loop struct {
		FPS    int      `json:"fps"`
		Frames []string `json:"frames"`
	}
	loops := map[string]loop{}
	var body strings.Builder
	letters := []string{"A", "B", "C"}
	for i := range scenes.Forest {
		sc := &scenes.Forest[i]
		key := fmt.Sprintf("opt%d", i)
		var fr []string
		for k := 0; k < n; k++ {
			t := float64(k) / float64(*fps)
			level := 0.5 - 0.4*math.Cos(2*math.Pi*float64(k)/float64(n))
			fr = append(fr, frame(sc, 0.75, t, level))
		}
		loops[key] = loop{FPS: *fps, Frames: fr}
		body.WriteString(`<section class="opt"><div class="head"><span class="letter">` + letters[i] + `</span><h2>` + html.EscapeString(sc.Name) + `</h2></div>`)
		body.WriteString(`<p class="note">` + html.EscapeString(sc.Note) + `</p>`)
		body.WriteString(`<div class="row"><div class="cell wide"><div class="cap">dusk, looping: the work rises from 0.1 to 0.9 and settles</div><div class="win"><div class="fx" data-loop="` + key + `"></div></div></div></div>`)
		body.WriteString(`<div class="row">`)
		for _, hr := range []struct {
			name  string
			tod   float64
			level float64
		}{{"night, working hard (0.8)", 0.0245, 0.8}, {"noon, idle (0.05)", 0.5, 0.05}} {
			body.WriteString(`<div class="cell"><div class="cap">` + hr.name + `</div><div class="win"><div class="fx">` + frame(sc, hr.tod, 1.5, hr.level) + `</div></div></div>`)
		}
		body.WriteString(`</div><div class="slots">`)
		for j, k := range []string{"light", "sky", "motion", "surface", "accumulator", "companion"} {
			body.WriteString(`<b>` + k + `</b><span>` + html.EscapeString(sc.Slots[j]) + `</span>`)
		}
		body.WriteString(`</div></section>`)
	}
	js, _ := json.Marshal(loops)

	page := `<title>Three Ways Up the Mountain</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Geist:wght@400;500;600&display=swap">
<style>
:root{--g:#f3f2ee;--s:#e9e7e1;--ln:#d3d0c8;--d:#5f5c55;--i:#141412;--acc:#2f5f8f;--warm:#8a6a1e;--win:#000}
@media (prefers-color-scheme:dark){:root:not([data-theme="light"]){--g:#121212;--s:#1b1b1b;--ln:#2e2e2e;--d:#8f8f8f;--i:#eeeeee;--acc:#87afff;--warm:#d7af5f}}
:root[data-theme="dark"]{--g:#121212;--s:#1b1b1b;--ln:#2e2e2e;--d:#8f8f8f;--i:#eeeeee;--acc:#87afff;--warm:#d7af5f}
*{box-sizing:border-box}
body{margin:0;background:var(--g);color:var(--i);font:15px/1.55 Geist,system-ui,-apple-system,Helvetica,Arial,sans-serif;padding-inline:16px;padding-block:36px 72px}
.wrap{max-width:1120px;margin:0 auto}
.lockup{font:700 20px/1.2 "JetBrains Mono",Menlo,monospace}
.blk{display:inline-block;width:1ch;background:var(--i);color:var(--g);text-align:center}
h1{font-size:clamp(24px,3.4vw,34px);line-height:1.15;font-weight:600;margin:14px 0 8px;text-wrap:balance;max-width:26em}
.lede{max-width:66ch;color:var(--d);margin:0 0 10px}
.lede b{color:var(--i);font-weight:600}
.ask{max-width:66ch;border-left:3px solid var(--warm);padding-left:14px;color:var(--i);margin:22px 0 40px}
.opt{border-top:1px solid var(--ln);padding-top:26px;margin:0 0 44px}
.head{display:flex;align-items:baseline;gap:14px;margin:0 0 8px}
.letter{font:700 13px/1 "JetBrains Mono",Menlo,monospace;color:var(--g);background:var(--acc);padding:6px 8px;border-radius:3px}
h2{font-size:22px;font-weight:600;margin:0}
.note{max-width:70ch;color:var(--d);margin:0 0 18px}
.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin:0 0 14px}
.cell{flex:1 1 380px;min-width:0}
.cell.wide{flex-basis:100%}
.cap{font:600 11px/1.4 Geist,sans-serif;letter-spacing:.08em;text-transform:uppercase;color:var(--d);margin:0 0 6px}
.win{background:var(--win);border:1px solid var(--ln);border-radius:4px;padding:8px 10px;overflow-x:auto}
.fx pre{margin:0;font-family:Menlo,"SF Mono","DejaVu Sans Mono",ui-monospace,monospace;line-height:1;letter-spacing:0;white-space:pre}
.slots{display:grid;grid-template-columns:max-content 1fr;gap:3px 16px;font-size:13.5px;color:var(--d);max-width:70ch;margin:16px 0 0}
.slots b{color:var(--i);font-weight:600}
h3{font-size:12px;letter-spacing:.1em;text-transform:uppercase;color:var(--d);font-weight:600;margin:40px 0 12px;padding-top:16px;border-top:1px solid var(--ln)}
.rec{max-width:70ch}
.rec p{margin:0 0 12px}
.rec b{font-weight:600}
table{border-collapse:collapse;font-size:13.5px;max-width:70ch}
td,th{text-align:left;padding:6px 14px 6px 0;vertical-align:top;border-bottom:1px solid var(--ln)}
th{font-weight:600;color:var(--i)}
td{color:var(--d)}
td:first-child{color:var(--i);font-weight:600;white-space:nowrap}
.foot{margin-top:40px;font:13px/1.5 "JetBrains Mono",Menlo,monospace;color:var(--d)}
` + pal.CSS() + `
</style>
<div class="wrap">
<div class="lockup"><span class="blk">x</span>scapes</div>
<h1>Three ways up the mountain</h1>
<p class="lede">The third scape is a mountain in the woods with the owl, and the design question is the one the shore never had to answer, because the sea answered it: <b>with no water, what carries the work?</b> Each option below gives one answer. Everything else is held the same: the same ridge, the same real sky and moon, the same owl on the right where the companion always sits, the same writing on the floor. All three loop through one cycle of work at dusk and are shown still at night working hard and at noon idle, as a 256-colour terminal draws them.</p>
<p class="ask">Pick one letter, or say what to change. Whichever wins goes through the pipeline built this morning for the rainy window and is on the page the same day.</p>
` + body.String() + `
<h3>Recommendation</h3>
<div class="rec">
<p><b>A, the wind.</b> It is the only one of the three whose motion is neither water nor fire. The aquarium already carries the work as water and the hearth as fire, so B and C would each be a second instance of a scape the study already has; A is the first of its kind, and a third scape has to be a third kind or it is not worth its section. It also encodes the way the rule asks: how many crowns lean is a count, how far is a coverage, and both survive a screenshot.</p>
<p><b>The risk with A</b> is the quiet end. At level 0.1 one crown moves a cell and there are a couple of dozen needles in the air, which is subtler than one swell on a flat sea. If that reads as "nothing is happening", the fix is a floor on the needles rather than a change of channel.</p>
<p><b>B reads strongest and costs the most.</b> The flame's height is the most legible work channel of the three, but the fire is also the light, so a busy agent brightens the whole clearing, which puts two meanings on one thing. The hearth lives with the same collision indoors; outdoors it is more visible.</p>
<p><b>C is the safe one.</b> It works because the shore works. That is also the argument against it.</p>
</div>
<table><tr><th>option</th><th>the work is</th><th>encoded as</th><th>collides with</th></tr>
<tr><td>A wind</td><td>how many crowns sway, how far, needles in the air</td><td>count + coverage</td><td>nothing</td></tr>
<tr><td>B fire</td><td>flame height, sparks</td><td>position + count</td><td>the light slot</td></tr>
<tr><td>C creek</td><td>foam coverage</td><td>coverage</td><td>the aquarium and the shore, which are already water</td></tr></table>
<p class="foot">go run ./notes/forest · internal/scenes/forest.go · seed ` + fmt.Sprint(*seed) + ` · Profile256 (as the product draws it; the page's clips are truecolor)</p>
</div>
<script>
(function(){
  var L=` + string(js) + `;
  var reduce=window.matchMedia&&window.matchMedia("(prefers-reduced-motion:reduce)").matches;
  var els=document.querySelectorAll("[data-loop]");
  for(var i=0;i<els.length;i++){(function(el){
    var lp=L[el.getAttribute("data-loop")]; if(!lp) return;
    var k=0; el.innerHTML=lp.frames[0];
    if(reduce) return;
    setInterval(function(){k=(k+1)%lp.frames.length; el.innerHTML=lp.frames[k];},1000/lp.fps);
  })(els[i]);}
})();
</script>
`
	if err := os.WriteFile(*out, []byte(page), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
