package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestSunStudy writes the vista's context body seven ways when
// XSCAPES_SUNSTUDY names a file: every style in scenes.MoonStyles by day and
// by night over one context window, one slider. His ask of 2026-09-16:
// "explore further visual approaches for the vista sun/moon and how it can
// visually represent the context being used up over time".
func TestSunStudy(t *testing.T) {
	out := os.Getenv("XSCAPES_SUNSTUDY")
	if out == "" {
		t.Skip("set XSCAPES_SUNSTUDY=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const steps = 21
	pal := &canvas.HTMLPalette{}
	var b strings.Builder
	clip := func(id string, render func(used float64) string) {
		fmt.Fprintf(&b, `<div class="clip" id="%s">`, id)
		for i := 0; i < steps; i++ {
			h := render(float64(i) / float64(steps-1))
			if i == 0 {
				h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
			}
			b.WriteString(h)
		}
		b.WriteString(`</div>`)
	}
	// Round two: he does not love the block or the disc and leans toward
	// the arc, so the page shows the arc's bodies beside the two he passed.
	for _, si := range []int{5} {
		ms := scenes.MoonStyles[si]
		fmt.Fprintf(&b, `<section><h2><span class="k">S%d</span> %s</h2><p class="note">%s</p><div class="row">`, si, ms.Name, ms.Note)
		for _, hr := range []struct {
			name string
			tod  float64
		}{{"10:10", 0.424}, {"22:00", 0.917}} {
			fmt.Fprintf(&b, `<figure class="ctx">`)
			clip(fmt.Sprintf("s%d-%s", si, strings.ReplaceAll(hr.name, ":", "")), func(used float64) string {
				st := reduce.State{Act: scape.Activity{Working: true, Level: 0.3, ContextUsed: used, TimeOfDay: hr.tod, TodoDone: 3}, Pose: companion.Working}
				f := newFrames(W, H, 7, false, true, 0, 0)
				f.vista.OwlX = f.lay.CatX
				f.vista.MoonStyle = si
				f.vista.Update(f.c, 3.0, st.Act)
				drawVista(f.c, f.vista, f.lay, st, 3.0)
				_, _, _, bandTop := f.vista.Layout()
				x1 := 70
				if strings.HasPrefix(scenes.MoonStyles[si].Name, "the arc") {
					x1 = W
				}
				return f.c.HTMLFragmentCropClassed(0, 0, x1, bandTop-5, 11, term.Profile256, pal)
			})
			fmt.Fprintf(&b, `<figcaption>%s</figcaption></figure>`, hr.name)
		}
		b.WriteString(`</div></section>`)
	}
	page := `<meta charset="utf-8"><title>Spending the Window</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff;--warm:#f4c67a}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:24px}
main{max-width:1240px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
.bar{position:sticky;top:env(safe-area-inset-top,0px);background:var(--bg);border-top:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:10px 0;display:flex;flex-wrap:wrap;gap:14px;align-items:center;z-index:2}
.bar button{font:inherit;font-size:13px;background:transparent;color:var(--ink);border:1px solid var(--mute);padding:4px 14px;cursor:pointer}
.bar button:focus-visible,.bar input:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.bar input[type=range]{flex:1;min-width:200px;max-width:520px;accent-color:var(--acc)}
.bar .fr{font-size:14px;font-variant-numeric:tabular-nums;min-width:30ch}
.bar .fr b{color:var(--warm)}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.k{color:var(--acc);margin-right:6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 14px}
.row{display:flex;flex-wrap:wrap;gap:28px;align-items:flex-start}
figure{margin:0;max-width:100%}
figcaption{color:var(--mute);font-size:13px;margin-top:8px}
.ctx{overflow-x:auto}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>Spending the window: the arc, locked</h1>
<p class="lede">Locked: the arc, S5, with your ramp. The body is bright when the window is fresh and dims as it spends: yellow to a muted darker yellow by day, light grey to darker grey by night. The path is unchanged: low on the left when fresh, overhead at 40%, setting behind the range on the right when spent. The readout follows the body, the stars keep their cells as it passes, and the lake's light follows it by night. Your look is the last word.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(steps-1) + `" value="0" aria-label="context used"><span class="fr" id="fr"></span></div>
` + b.String() + `
<section><h2>what the arc spends</h2><p class="note">The arc encodes context in position along a path, which is the rule's own currency, and it reads at a glance against two references, the top of the sky and the range it sets behind; the ramp is a second cue on the same variable, as phase is beside altitude on the shore. What it costs: the body crosses the todo stars, which keep their cells as it passes, and at the end it sets on the owl's side of the frame.</p></section>
<p class="foot">Made from the tree: <code>XSCAPES_SUNSTUDY=&lt;file&gt; go test -run TestSunStudy .</code> · 125x28, seed 7, the 256-colour cube; the disc styles sample four quarters a cell, the massif's own method.</p>
</main>
<script>
(function(){
var clips=[].slice.call(document.querySelectorAll('.clip')),n=` + fmt.Sprint(steps) + `;
var sl=document.getElementById('sl'),fr=document.getElementById('fr'),play=document.getElementById('play'),k=0,timer=null;
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;var used=Math.round(100*i/(n-1));var s='context used '+used+'% · left '+(100-used)+'%';if(used>=85){s+=' · <b>warm readout</b>';}else if(used>=40){s+=' · readout on';}fr.innerHTML=s;}
function start(){if(timer)return;timer=setInterval(function(){show((k+1)%n);},350);play.textContent='Pause';}
function stop(){clearInterval(timer);timer=null;play.textContent='Play';}
play.addEventListener('click',function(){timer?stop():start();});
sl.addEventListener('input',function(){stop();show(+sl.value);});
show(0);if(reduced){play.textContent='Play';}else{start();}
})();
</script>`
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s, %d bytes", out, len(page))
}
