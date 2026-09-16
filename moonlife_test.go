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
	"github.com/donlucasx/xscapes/internal/term"
)

// TestMoonLife writes the sun and moon's whole life over a context window
// when XSCAPES_MOONLIFE names a file: the vista's block and the shore's disc
// side by side at three hours, at every five percent of context used, one
// slider. His ask of 2026-09-16: "Show me on a separate page the entire life
// cycle of the sun and moon as context depletes".
func TestMoonLife(t *testing.T) {
	out := os.Getenv("XSCAPES_MOONLIFE")
	if out == "" {
		t.Skip("set XSCAPES_MOONLIFE=<file> to write the page")
	}
	const W, H = 125, 28
	const steps = 21
	pal := &canvas.HTMLPalette{}
	hours := []struct {
		name string
		tod  float64
	}{{"morning, 10:10", 0.424}, {"dusk, 18:45", 0.78}, {"night, 22:00", 0.917}}
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
	for _, hr := range hours {
		fmt.Fprintf(&b, `<section><h2>%s</h2><div class="row">`, hr.name)
		// The vista.
		t.Setenv("XSCAPES_SCAPE", "vista")
		fmt.Fprintf(&b, `<figure class="ctx">`)
		clip("vista-"+fmt.Sprint(hr.tod), func(used float64) string {
			st := reduce.State{Act: scape.Activity{Working: true, Level: 0.3, ContextUsed: used, TimeOfDay: hr.tod, TodoDone: 3}, Pose: companion.Working}
			f := renderVista(t, W, H, st, 3.0)
			_, _, _, bandTop := f.vista.Layout()
			return f.c.HTMLFragmentCropClassed(0, 0, 66, bandTop-5, 11, term.Profile256, pal)
		})
		b.WriteString(`<figcaption>the vista: the study's 3x2 block, row 1 fresh, sinking to just above the far range, never into it; the readout under it from 40% used</figcaption></figure>`)
		// The shore.
		t.Setenv("XSCAPES_SCAPE", "shore")
		fmt.Fprintf(&b, `<figure class="ctx">`)
		clip("shore-"+fmt.Sprint(hr.tod), func(used float64) string {
			f := newFrames(W, H, 7, false, true, 0, 0)
			if f.vista != nil {
				t.Fatalf("the shore was asked for")
			}
			st := reduce.State{Act: scape.Activity{Working: true, Level: 0.3, ContextUsed: used, TimeOfDay: hr.tod, TodoDone: 3}, Pose: companion.Working}
			f.sh.Update(f.c, 3.0, st.Act)
			top := f.c.H - 2 - f.chh
			drawScene(f.c, f.sh, f.cat, f.lay, st, 3.0, f.seed, top)
			return f.c.HTMLFragmentCropClassed(0, 0, 72, f.sh.SandTop()+1, 11, term.Profile256, pal)
		})
		b.WriteString(`<figcaption>the shore: a disc with a phase, full and high when fresh, new and on the horizon when spent; by day a sun that wanes as a crescent with no unlit face; the same readout</figcaption></figure>`)
		b.WriteString(`</div></section>`)
	}
	page := `<meta charset="utf-8"><title>Sun and Moon by Context</title>
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
h2{font-size:17px;font-weight:700;margin:0 0 14px}
.row{display:flex;flex-wrap:wrap;gap:28px;align-items:flex-start}
figure{margin:0;max-width:100%}
figcaption{color:var(--mute);font-size:13px;margin-top:8px;max-width:60ch}
.ctx{overflow-x:auto}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
table{border-collapse:collapse;font-size:14px;margin:6px 0 10px}
td,th{text-align:left;padding:4px 18px 4px 0;border-bottom:1px solid var(--rule);vertical-align:top}
th{color:var(--mute);font-weight:400}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The sun and the moon over one context window</h1>
<p class="lede">The context runs from fresh to spent on a loop, a session compressed to seven seconds; pause it and drag the slider to sit on one reading. The vista's body is the study's block; the shore's is the disc with a phase. Both sink. Only the shore's changes shape. Three hours, so you see the sun by day, the low light at dusk and the moon by night; the hour changes only the palette and which body it is, never the altitude.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(steps-1) + `" value="0" aria-label="context used"><span class="fr" id="fr"></span></div>
` + b.String() + `
<section><h2>the mapping, as the code has it</h2>
<table>
<tr><th></th><th>the vista</th><th>the shore</th></tr>
<tr><td>altitude</td><td>row 1 at 0% used, down to four rows above the far range's top at 100%, linear; never into the range</td><td>0.22 of the horizon's height at 0% used, 0.84 at 100%, linear</td></tr>
<tr><td>shape</td><td>a 3x2 block, always</td><td>a disc, radius two rows, half-row sampled, hue rim; the terminator is a second disc sliding across: full at 0%, new at 100%</td></tr>
<tr><td>by day</td><td>the block in the sun's tan</td><td>a sun that wanes as a crescent; no unlit face is painted (his ruling of 2026-09-06)</td></tr>
<tr><td>readout</td><td>from 40% used, dim: the percent LEFT; from 85% used, warm: &ldquo;NN% left&rdquo;</td><td>the same thresholds, the same colours, the label beside the disc when there is no room under it</td></tr>
</table></section>
<p class="foot">Made from the tree: <code>XSCAPES_MOONLIFE=&lt;file&gt; go test -run TestMoonLife .</code> · 125x28, seed 7, the 256-colour cube, the left half of each frame.</p>
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
