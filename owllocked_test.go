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

// TestOwlLocked writes the owl's locked behaviour as one page when
// XSCAPES_OWLLOCKED names a file: his picks for the five states, the
// litter's schedule and its placement, one clip each, and a short list of
// what is still open. His word of 2026-09-16: "lock in all the previous
// decisions we made w regards to the owl behavior and only keep pending
// decisions on the page".
func TestOwlLocked(t *testing.T) {
	out := os.Getenv("XSCAPES_OWLLOCKED")
	if out == "" {
		t.Skip("set XSCAPES_OWLLOCKED=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const hr = 0.424
	const fps = 12.0
	const loop = 8.0
	const t0 = 5.0
	n := int(loop * fps)
	scenes.OwlMotionPeriodOverride = 8
	defer func() { scenes.OwlMotionPeriodOverride = 0 }()
	ground := term.RGB{R: 255, G: 0, B: 255}
	pal := &canvas.HTMLPalette{Transparent: &ground}
	tail := []reduce.Line{
		{Text: "read  internal/scenes/owl.go", Age: 0.7},
		{Text: "edit  internal/scenes/owl_anim.go  +400", Age: 0.4},
		{Text: "shell go test ./...  exit 1", Age: 0.2, Bad: true},
		{Text: "shell go build ./...  1.2s", Age: 0.0},
	}
	act := func(level float64, working bool) scape.Activity {
		return scape.Activity{Working: working, Level: level, ContextUsed: 0.45, TimeOfDay: hr, TodoDone: 3}
	}
	var b strings.Builder
	clip := func(id string, render func(tt float64) string) {
		fmt.Fprintf(&b, `<div class="clip" id="%s">`, id)
		for i := 0; i < n; i++ {
			h := render(t0 + float64(i)/fps)
			if i == 0 {
				h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
			}
			b.WriteString(h)
		}
		b.WriteString(`</div>`)
	}
	alone := func(tt float64, st companion.State) string {
		c := canvas.New(18, 10, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, ground)
			}
		}
		scenes.DrawOwlMoving(c, 3, 2, tt, st, scenes.PickedOwlMotion(st), 8)
		return c.HTMLFragmentClassed(22, term.Profile256, pal)
	}
	ctx := func(f *frames, left, up int) string {
		owlX, owlY, _, bandTop := f.vista.Layout()
		return f.c.HTMLFragmentCropClassed(owlX-left, max(0, owlY-up), W, min(H, bandTop+2), 13, term.Profile256, pal)
	}
	type sec struct {
		name, ruling string
		st           reduce.State
	}
	secs := []sec{
		{"resting", "peek: one eye opens, looks the other way, closes, 1.2 s in 10", reduce.State{Act: act(0, false), Pose: companion.Resting}},
		{"working", "looks the other way, looks back, blinks, then the hop and flutter on both wings, 3.2 s in 10; the blink every 7 s stays", reduce.State{Act: act(0.65, true), Pose: companion.Working, Tail: tail}},
		{"needs you", "double blink every 4 s, and every other time the near wing waves with it", reduce.State{Act: act(0.1, false), Pose: companion.NeedsYou, Bubble: "allow Bash?", BubbleAsk: true, Tail: tail}},
		{"done", "flap and bounce: the small wings out with each of two hops, every 5 s (your afternoon ruling: small wings for both)", reduce.State{Act: act(0.05, false), Pose: companion.Done, Bubble: "done", Tail: tail}},
		{"worried", "squint with brows: the eyes a white band a cell tall with the pupil in it, worried brows raised toward the middle, held; the pupils dart every 3 s", reduce.State{Act: act(1.0, true), Pose: companion.Worried, Tail: tail}},
	}
	b.WriteString(`<section><h2>the five faces, locked</h2><div class="cands">`)
	for _, sc := range secs {
		id := strings.ReplaceAll(sc.name, " ", "-")
		fmt.Fprintf(&b, `<div class="cand"><h3>%s</h3><p class="what">%s</p><div class="pair">`, sc.name, sc.ruling)
		clip(id+"-a", func(tt float64) string { return alone(tt, sc.st.Pose) })
		b.WriteString(`<div class="ctx">`)
		clip(id+"-c", func(tt float64) string { return ctx(renderVista(t, W, H, sc.st, tt), 16, 4) })
		b.WriteString(`</div></div></div>`)
	}
	b.WriteString(`</div></section>`)

	b.WriteString(`<section><h2>the litter, locked</h2><p class="note">One owlet per subagent; the count is the channel. <b>The litter's life:</b> blink, hop, cheep and wiggle, one owlet at a time in its own window of the period, the act rotating each period, the litter still at least half the time at any count, blinks free. <b>By turns:</b> one beside the owl, one at the fire, one beside the owl. <b>The flights:</b> an owlet comes out from behind the owl when its subagent starts, climbs straight up to the owl's top row, crosses above the litter, and drops straight down onto its spot; it leaves the same way in reverse. In the clip one owlet is already here, four more arrive at t = 5.3, 6.3, 7.3 and 8.3 s, by turns, and the oldest flies off from 10.0 s.</p><div class="cands">`)
	f := newFrames(W, H, 7, false, true, 0, 0)
	b.WriteString(`<div class="cand wide"><div class="ctx">`)
	clip("litter", func(tt float64) string {
		st := reduce.State{Act: act(0.65, true), Pose: companion.Working, Tail: tail, Kittens: 1}
		switch {
		case tt >= 10.0:
			st.Kittens, st.KittenExits = 4, []float64{(tt - 10.0) / 1.5}
		case tt >= 8.3:
			st.Kittens = 5
		case tt >= 7.3:
			st.Kittens = 4
		case tt >= 6.3:
			st.Kittens = 3
		case tt >= 5.3:
			st.Kittens = 2
		}
		f.vista.OwlX = f.lay.CatX
		f.vista.Update(f.c, tt, st.Act)
		drawVista(f.c, f.vista, f.lay, st, tt)
		owlX, owlY, _, bandTop := f.vista.Layout()
		return f.c.HTMLFragmentCropClassed(f.vista.FireX()-24, max(0, owlY-4), min(W, owlX+14), min(H, bandTop+2), 12, term.Profile256, pal)
	})
	b.WriteString(`</div></div></div></section>`)

	page := `<meta charset="utf-8"><title>The Owl, Locked</title>
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
.bar input[type=range]{flex:1;min-width:160px;max-width:440px;accent-color:var(--acc)}
.bar .fr{color:var(--mute);font-size:13px;font-variant-numeric:tabular-nums;min-width:22ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.note{max-width:70ch;margin:0 0 18px}
.note b{color:var(--warm);font-weight:700}
.cands{display:flex;flex-wrap:wrap;gap:26px 30px;align-items:flex-start}
.cand{max-width:100%}
.cand.wide{width:100%}
h3{font-size:15px;font-weight:700;margin:0 0 2px}
.what{color:var(--mute);font-size:13px;max-width:44ch;margin:0 0 10px;min-height:2.6em}
.pair{display:flex;gap:14px;align-items:flex-end}
.ctx{overflow-x:auto;max-width:100%}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
ul{max-width:66ch;padding-left:1.2em}
li{margin:6px 0}
a{color:var(--acc)}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The owl, locked</h1>
<p class="lede">Your rulings of 2026-09-16 on the owl and its litter, as the installed vista draws them. Done's flap is on the small wings now, your ruling of the afternoon; nothing about the owl is open. One clip each. Every period is compressed to eight seconds here, one loop, so each event shows once and the needs-you wave shows on its second blink.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(n-1) + `" value="0" aria-label="frame"><span class="fr" id="fr"></span></div>
` + b.String() + `
<section><h2>still open</h2><ul>
<li><b>Working and needs-you, rebuilt to your notes:</b> the looks-blink-flutter sequence and the wave on every other double blink. Your look.</li>
<li><b>The litter by turns, now actually by turns.</b> The pick had not reached the code when you looked; it has.</li>
<li><b>The sun and moon: the arc, locked,</b> bright when fresh and dimming as the window spends. <a href="https://claude.ai/artifact/EgRNj7JE1gsLrVYb82AeZF">Your look at the ramp</a>.</li>
</ul></section>
<p class="foot">Made from the tree: <code>XSCAPES_OWLLOCKED=&lt;file&gt; go test -run TestOwlLocked .</code> · 125x28, seed 7, 10:10, the 256-colour cube. The candidates he passed over stay in <code>owl_anim.go</code>; the round page is superseded by this one.</p>
</main>
<script>
(function(){
var clips=[].slice.call(document.querySelectorAll('.clip')),n=` + fmt.Sprint(n) + `,k=0,fps=` + fmt.Sprint(fps) + `;
var sl=document.getElementById('sl'),fr=document.getElementById('fr'),play=document.getElementById('play');
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
var on=!reduced,timer=null;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;fr.textContent='frame '+(i+1)+'/'+n+' · t = '+(5+i/fps).toFixed(2)+' s';}
function tick(){show((k+1)%n);}
function start(){if(timer)return;timer=setInterval(tick,Math.round(1000/fps));play.textContent='Pause';on=true;}
function stop(){clearInterval(timer);timer=null;play.textContent='Play';on=false;}
play.addEventListener('click',function(){on?stop():start();});
sl.addEventListener('input',function(){stop();show(+sl.value);});
show(0);if(on)start();else play.textContent='Play';
})();
</script>`
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s, %d bytes", out, len(page))
}
