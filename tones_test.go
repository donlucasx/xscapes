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

// TestTonesPage writes the vista's daytime tones as one page when
// XSCAPES_TONES names a file: every candidate as the whole live frame at
// 125x28, one slider over the hour from 06:00 to 20:00. His note of
// 2026-09-16: "can we pick more pleasant tones for the daytime cycle".
func TestTonesPage(t *testing.T) {
	out := os.Getenv("XSCAPES_TONES")
	if out == "" {
		t.Skip("set XSCAPES_TONES=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const steps = 29 // 06:00 to 20:00 by half hours
	pal := &canvas.HTMLPalette{}
	was := scenes.VistaTonePick
	defer func() { scenes.VistaTonePick = was }()
	tail := []reduce.Line{
		{Text: "read  internal/scenes/vista_tones.go", Age: 0.6},
		{Text: "edit  internal/scenes/forest.go  +6 -3", Age: 0.3},
		{Text: "shell go test ./...  ok", Age: 0.0},
	}
	var b strings.Builder
	for i, tone := range scenes.VistaTones {
		if i != 0 && i != was {
			continue // locked: today for reference, and his pick
		}
		scenes.VistaTonePick = i
		fmt.Fprintf(&b, `<section><h2><span class="k">T%d</span> %s</h2><p class="note">%s</p><div class="ctx"><div class="clip" id="t%d">`, i, tone.Name, tone.Note, i)
		for k := 0; k < steps; k++ {
			hour := 6 + 0.5*float64(k)
			st := reduce.State{Act: scape.Activity{Working: true, Level: 0.35, ContextUsed: 0.3, TimeOfDay: hour / 24, TodoDone: 2}, Pose: companion.Working, Kittens: 2, Tail: tail}
			f := renderVista(t, W, H, st, 3.0)
			h := f.c.HTMLFragmentClassed(9, term.Profile256, pal)
			if k == 0 {
				h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
			}
			b.WriteString(h)
		}
		b.WriteString(`</div></div></section>`)
	}
	page := `<meta charset="utf-8"><title>Daytime Tones</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:24px}
main{max-width:1240px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
.bar{position:sticky;top:env(safe-area-inset-top,0px);background:var(--bg);border-top:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:10px 0;display:flex;flex-wrap:wrap;gap:14px;align-items:center;z-index:2}
.bar button{font:inherit;font-size:13px;background:transparent;color:var(--ink);border:1px solid var(--mute);padding:4px 14px;cursor:pointer}
.bar button:focus-visible,.bar input:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.bar input[type=range]{flex:1;min-width:200px;max-width:520px;accent-color:var(--acc)}
.bar .fr{font-size:14px;font-variant-numeric:tabular-nums;min-width:12ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.k{color:var(--acc);margin-right:6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 14px}
.ctx{overflow-x:auto}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>Daytime tones for the vista</h1>
<p class="lede">Locked: T4, deep and cool, with today above it for reference. The hour runs from 06:00 to 20:00 on a loop. Since your note on the transitions: the sky passes through a mauve stop at the hours when the horizon is warm, so the ramp from a blue zenith to a peach or orange horizon never crosses grey, and the firelight on the meadow is painted only as the dark comes, because painting it by day dropped those cells off the meadow's ramp and left a patch. What remains is the cube itself: as the hour moves, a row of the sky changes its entry when the ramp's path re-rounds, and the bands crawl. That is every terminal's 216 colours, not a defect to fix here.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(steps-1) + `" value="0" aria-label="hour"><span class="fr" id="fr"></span></div>
` + b.String() + `
<p class="foot">Made from the tree: <code>XSCAPES_TONES=&lt;file&gt; go test -run TestTonesPage .</code> · 125x28, seed 7, working at 35% with two owlets and 30% of the window used, so the gauge and the litter are in every frame.</p>
</main>
<script>
(function(){
var clips=[].slice.call(document.querySelectorAll('.clip')),n=` + fmt.Sprint(steps) + `;
var sl=document.getElementById('sl'),fr=document.getElementById('fr'),play=document.getElementById('play'),k=0,timer=null;
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;var h=6+i/2;var hh=Math.floor(h),mm=(h%1)?'30':'00';fr.textContent=(hh<10?'0':'')+hh+':'+mm;}
function start(){if(timer)return;timer=setInterval(function(){show((k+1)%n);},400);play.textContent='Pause';}
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
