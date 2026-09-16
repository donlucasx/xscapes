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

// TestOwlSheet writes the owl's five faces and the owlets as one review
// page when XSCAPES_OWLSHEET names a file: each state as the owl alone,
// large, and as the live composition at 125x28 around the owl, both over one
// four-second loop of the vista at 6 frames a second, scrubbed by one slider.
// His ask of 2026-09-16: "can I see all the OWL states in a separate page so
// I can review its actions and animations? same for the sub agents".
func TestOwlSheet(t *testing.T) {
	out := os.Getenv("XSCAPES_OWLSHEET")
	if out == "" {
		t.Skip("set XSCAPES_OWLSHEET=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const hr = 0.424 // 10:10, the hour of his first look
	const fps = 6.0  // the vista's own step rate
	n := int(scenes.LoopSecs * fps)
	t0 := 4.0 // a whole loop that holds one blink (t mod 7 < 0.25 at t = 7)
	// The ground the page writes as transparent: a cube entry, because these
	// frames are rendered as the 256-colour terminal shows them and a colour
	// off the cube would quantise to black before the palette saw it (it did:
	// the first render put the owl on a black square). Magenta is a colour no
	// scene paints.
	ground := term.RGB{R: 255, G: 0, B: 255}
	pal := &canvas.HTMLPalette{Transparent: &ground}
	tail := []reduce.Line{
		{Text: "read  internal/scenes/owl.go", Age: 0.7},
		{Text: "edit  internal/scenes/vista.go  +12 -3", Age: 0.4},
		{Text: "shell go test ./...  exit 1", Age: 0.2, Bad: true},
		{Text: "shell go build ./...  1.2s", Age: 0.0},
	}
	act := func(level float64, working bool) scape.Activity {
		return scape.Activity{Working: working, Level: level, ContextUsed: 0.45, TimeOfDay: hr, TodoDone: 3}
	}
	type face struct {
		name, rule string
		st         reduce.State
	}
	faces := []face{
		{"resting", "Both lids down: a dark line across the top of each eye. Between turns, when the agent is waiting on you. Still.",
			reduce.State{Act: act(0, false), Pose: companion.Resting}},
		{"working", "Open, the pupil out. One blink every seven seconds, on the crab's pincer clock: idle motion, no meaning. The one face with motion of its own; here the blink lands at t = 7.0 s, frames 19 and 20.",
			reduce.State{Act: act(0.65, true), Pose: companion.Working, Tail: tail}},
		{"needs you", "Wide: each eye a cell wider, the pupil centred, the highlight kept. The ask, with the warm solid balloon and the bird. The owl does not walk up the way the crab does; that round is not built.",
			reduce.State{Act: act(0.1, false), Pose: companion.NeedsYou, Bubble: "allow Bash?", BubbleAsk: true, Tail: tail}},
		{"done", "Content: the eyes shut upward, ^ ^. The finish, with the cool dotted knock and the drop.",
			reduce.State{Act: act(0.05, false), Pose: companion.Done, Bubble: "done", Tail: tail}},
		{"worried", "Amber slits: the lower row only, in the beak's colour, the pupil dark. A main-thread failure you have not seen. It holds until the failure clears, and a finish in this state rings the falling bird.",
			reduce.State{Act: act(1.0, true), Pose: companion.Worried, Tail: tail}},
	}
	var b strings.Builder
	clip := func(id string, render func(tt float64) string) {
		fmt.Fprintf(&b, `<div class="clip" id="%s">`, id)
		for i := 0; i < n; i++ {
			tt := t0 + float64(i)/fps
			h := render(tt)
			if i == 0 {
				h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
			}
			b.WriteString(h)
		}
		b.WriteString(`</div>`)
	}
	owlAlone := func(tt float64, pose companion.State) string {
		c := canvas.New(16, 9, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, ground)
			}
		}
		scenes.DrawOwlPose(c, 2, 1, tt, pose)
		return c.HTMLFragmentClassed(26, term.Profile256, pal)
	}
	crop := func(f *frames, left int) string {
		owlX, owlY, _, bandTop := f.vista.Layout()
		return f.c.HTMLFragmentCropClassed(owlX-left, max(0, owlY-5), W, min(H, bandTop+2), 12, term.Profile256, pal)
	}

	// The faces.
	for _, fc := range faces {
		id := strings.ReplaceAll(fc.name, " ", "-")
		fmt.Fprintf(&b, `<section><h2>%s</h2><p class="rule">%s</p><div class="row">`, fc.name, fc.rule)
		fmt.Fprintf(&b, `<figure>`)
		clip(id+"-owl", func(tt float64) string { return owlAlone(tt, fc.st.Pose) })
		fmt.Fprintf(&b, `<figcaption>the owl alone, 26 px cells</figcaption></figure>`)
		fmt.Fprintf(&b, `<figure class="ctx">`)
		clip(id+"-ctx", func(tt float64) string { return crop(renderVista(t, W, H, fc.st, tt), 50) })
		fmt.Fprintf(&b, `<figcaption>in the vista, 125x28, 10:10, seed 7, the frame's right half</figcaption></figure>`)
		b.WriteString(`</div></section>`)
	}

	// The owlets.
	b.WriteString(`<section><h2>owlets</h2><p class="rule">One owlet per live subagent; the count is the whole channel. They stand on the grass to the left of the owl, nearest first, seven cells apart, and stop short of the fire: a frame too narrow shows fewer rather than piling them. An owlet appears the moment the reducer counts its subagent and leaves when the reducer drops it, at least sixty seconds after it started. No motion of its own, no arrival, no flight off: it is there, then it is not.</p><div class="row">`)
	b.WriteString(`<figure>`)
	clip("owlet-alone", func(tt float64) string {
		c := canvas.New(8, 5, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, ground)
			}
		}
		scenes.DrawOwlet(c, scenes.OwlCoat, 1, 0, false)
		return c.HTMLFragmentClassed(26, term.Profile256, pal)
	})
	b.WriteString(`<figcaption>one owlet, the egg, 26 px cells</figcaption></figure></div>`)
	for _, k := range []int{1, 2, 3, 5} {
		st := reduce.State{Act: act(0.65, true), Pose: companion.Working, Kittens: k, Tail: tail}
		fmt.Fprintf(&b, `<div class="row one"><figure class="ctx">`)
		clip(fmt.Sprintf("owlets-%d", k), func(tt float64) string { return crop(renderVista(t, W, H, st, tt), 56) })
		word := "subagents"
		if k == 1 {
			word = "subagent"
		}
		fmt.Fprintf(&b, `<figcaption>%d %s, working</figcaption></figure></div>`, k, word)
	}
	b.WriteString(`</section>`)

	page := `<meta charset="utf-8"><title>The Owl's Five Faces</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:24px}
main{max-width:1180px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
.bar{position:sticky;top:env(safe-area-inset-top,0px);background:var(--bg);border-top:1px solid var(--rule);border-bottom:1px solid var(--rule);padding:10px 0;display:flex;flex-wrap:wrap;gap:14px;align-items:center;z-index:2}
.bar button{font:inherit;font-size:13px;background:transparent;color:var(--ink);border:1px solid var(--mute);padding:4px 14px;cursor:pointer}
.bar button:focus-visible,.bar input:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.bar input[type=range]{flex:1;min-width:160px;max-width:440px;accent-color:var(--acc)}
.bar .fr{color:var(--mute);font-size:13px;font-variant-numeric:tabular-nums;min-width:19ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
section:first-of-type{border-top:0}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.rule{max-width:66ch;margin:0 0 18px}
.row{display:flex;flex-wrap:wrap;gap:28px;align-items:flex-start;margin-bottom:18px}
figure{margin:0;max-width:100%}
figcaption{color:var(--mute);font-size:13px;margin-top:8px}
.ctx{overflow-x:auto}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
ul{max-width:66ch;padding-left:1.2em}
li{margin:4px 0}
code{font-size:13px;color:var(--acc)}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
` + pal.CSS() + `
</style>
<main>
<h1>The owl's five faces, and the owlets</h1>
<p class="lede">Every state the barn owl has, as the vista draws it today: alone, large, and in the live composition at your window's width. One slider runs every clip through the same four-second loop of the vista, six frames a second. The face is the channel; the owl itself is still by nature, so the only motion of its own is the working blink.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(n-1) + `" value="0" aria-label="frame"><span class="fr" id="fr"></span></div>
` + b.String() + `
<section><h2>not built, said plainly</h2><ul>
<li>The owl walking up to the screen for the ask, the way the crab does. The ask here is the wide eyes and the balloon.</li>
<li>Owlets arriving and flying off. They appear and vanish on the reducer's count.</li>
<li>A disc moon with a phase. The vista's moon and sun are the study's 3x2 block; it sinks with the context and carries the readout from 40% used, as on the shore, but it has no phase and no round edge.</li>
<li>Pacing. The owl does not step with the tool count; it sits.</li>
</ul></section>
<p class="foot">Made from the installed painter: <code>XSCAPES_OWLSHEET=&lt;file&gt; go test -run TestOwlSheet .</code> · 125x28, seed 7, 10:10, the 256-colour cube, the balloon on its new ground.</p>
</main>
<script>
(function(){
var clips=[].slice.call(document.querySelectorAll('.clip')),n=` + fmt.Sprint(n) + `,k=0,fps=` + fmt.Sprint(fps) + `,t0=` + fmt.Sprint(t0) + `;
var sl=document.getElementById('sl'),fr=document.getElementById('fr'),play=document.getElementById('play');
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
var on=!reduced,timer=null;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;fr.textContent='frame '+(i+1)+'/'+n+' · t = '+(t0+i/fps).toFixed(2)+' s';}
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
	t.Logf("wrote %s, %d bytes, %d frames a clip", out, len(page), n)
}
