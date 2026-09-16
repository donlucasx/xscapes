package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestOwlRound writes the motion round as one page when XSCAPES_OWLROUND
// names a file: every candidate for every state, the owl alone and in the
// live composition, over one four-second loop at 12 frames a second, every
// motion's period compressed to four seconds so each event shows once a
// loop; the owlets' candidates on a litter of three. His notes of
// 2026-09-16 are the section heads, verbatim, in owl_anim.go.
func TestOwlRound(t *testing.T) {
	out := os.Getenv("XSCAPES_OWLROUND")
	if out == "" {
		t.Skip("set XSCAPES_OWLROUND=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const hr = 0.424
	const fps = 12.0
	const loop = 4.0
	const t0 = 5.0 // a loop that holds today's working blink at t = 7
	n := int(loop * fps)
	scenes.OwlMotionPeriodOverride = 4
	defer func() { scenes.OwlMotionPeriodOverride = 0 }()
	picks, lpick := map[companion.State]int{}, scenes.OwletMotionPick
	for k, v := range scenes.OwlMotionPick {
		picks[k] = v
	}
	defer func() { scenes.OwlMotionPick = picks; scenes.OwletMotionPick = lpick }()
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
	type sec struct {
		name, note string
		st         reduce.State
	}
	letter := map[companion.State]string{companion.Resting: "R", companion.Working: "W", companion.NeedsYou: "N", companion.Done: "D", companion.Worried: "X"}
	secs := []sec{
		{"resting", "I like it, but every now and then it should open its eyes, and look to either side, and then close them again.", reduce.State{Act: act(0, false), Pose: companion.Resting}},
		{"working", "can we try a longer hop, so it hops and flutters for a moment before landing again? Wings should be smaller. instead of look and bob, try a flutter and look", reduce.State{Act: act(0.65, true), Pose: companion.Working, Tail: tail}},
		{"needs you", "Lets add an eye blink every now and then. Or double blink as if trying to get your attention.", reduce.State{Act: act(0.1, false), Pose: companion.NeedsYou, Bubble: "allow Bash?", BubbleAsk: true, Tail: tail}},
		{"done", "needs some sort of action or movement too", reduce.State{Act: act(0.05, false), Pose: companion.Done, Bubble: "done", Tail: tail}},
		{"worried", "I want to see other alternatives for the eyes. These look tired. Maybe its a squint action, maybe another alt of the squint with a wave (only one wing wave)", reduce.State{Act: act(1.0, true), Pose: companion.Worried, Tail: tail}},
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
	alone := func(tt float64, st companion.State, m *scenes.OwlMotion) string {
		c := canvas.New(18, 10, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, ground)
			}
		}
		scenes.DrawOwlMoving(c, 3, 2, tt, st, m, 4)
		return c.HTMLFragmentClassed(22, term.Profile256, pal)
	}
	ctx := func(f *frames, left, up int) string {
		owlX, owlY, _, bandTop := f.vista.Layout()
		return f.c.HTMLFragmentCropClassed(owlX-left, max(0, owlY-up), W, min(H, bandTop+2), 13, term.Profile256, pal)
	}
	for _, sc := range secs {
		ms := scenes.OwlMotions[sc.st.Pose]
		pick, maxRound := scenes.OwlMotionPick[sc.st.Pose], 0
		for i := range ms {
			maxRound = max(maxRound, ms[i].Round)
		}
		fmt.Fprintf(&b, `<section><h2>%s</h2><p class="note">&ldquo;%s&rdquo;</p><div class="cands">`, sc.name, sc.note)
		for i := range ms {
			m := &ms[i]
			if maxRound > 0 && i != 0 && i != pick && m.Round < maxRound {
				continue // an earlier round's candidate he passed over
			}
			id := fmt.Sprintf("%s-%d", strings.ReplaceAll(sc.name, " ", "-"), i)
			tag := ""
			if i == pick && i > 0 {
				tag = ` <span class="pick">his pick</span>`
			}
			fmt.Fprintf(&b, `<div class="cand"><h3><span class="k">%s%d</span> %s%s</h3><p class="what">%s</p><div class="pair">`, letter[sc.st.Pose], i, m.Name, tag, m.Note)
			if i == 0 {
				clip(id+"-a", func(tt float64) string { return alone(tt, sc.st.Pose, nil) })
			} else {
				clip(id+"-a", func(tt float64) string { return alone(tt, sc.st.Pose, m) })
			}
			scenes.OwlMotionPick[sc.st.Pose] = i
			b.WriteString(`<div class="ctx">`)
			clip(id+"-c", func(tt float64) string { return ctx(renderVista(t, W, H, sc.st, tt), 16, 4) })
			b.WriteString(`</div></div></div>`)
		}
		if v, ok := picks[sc.st.Pose]; ok {
			scenes.OwlMotionPick[sc.st.Pose] = v
		} else {
			delete(scenes.OwlMotionPick, sc.st.Pose)
		}
		b.WriteString(`</div></section>`)
	}
	// The owlets, on a litter of three, working.
	b.WriteString(`<section><h2>owlets</h2><p class="note">&ldquo;I like L1, L2, L3, L4- all of it basically. Whats the best criteria to implement these.&rdquo;</p><div class="cands">`)
	lst := reduce.State{Act: act(0.65, true), Pose: companion.Working, Kittens: 3, Tail: tail}
	for i := range scenes.OwletMotions {
		m := &scenes.OwletMotions[i]
		scenes.OwletMotionPick = i
		tag := ""
		if i == lpick && i > 0 {
			tag = ` <span class="pick">his pick</span>`
		}
		fmt.Fprintf(&b, `<div class="cand wide"><h3><span class="k">L%d</span> %s%s</h3><p class="what">%s</p><div class="ctx">`, i, m.Name, tag, m.Note)
		clip(fmt.Sprintf("owlets-%d", i), func(tt float64) string { return ctx(renderVista(t, W, H, lst, tt), 30, 2) })
		b.WriteString(`</div></div>`)
	}
	scenes.OwletMotionPick = lpick
	b.WriteString(`</div></section>`)

	// Placement: the litter beside the owl and around the fire, each with
	// a scripted count so the flights show: one owlet long present, a second
	// flies in at t = 5.5, a third at 6.5, and the oldest flies off from 7.5.
	b.WriteString(`<section><h2>placement, and the flights</h2><p class="note">&ldquo;Could owlets fly around? Maybe gather around the fire?&rdquo; then &ldquo;explore both placements, simultaneously&rdquo;. Not around: the flight is the event, not the state. An owlet comes out from behind the owl when its subagent starts, flies to its spot, sits there with L6, and flies back behind the owl when the reducer lets it go. A farther spot takes a longer flight. Four places to sit. In each clip one owlet is already here, four more arrive at t = 5.3, 5.9, 6.5 and 7.1 s, and the oldest flies off from 8.0 s.</p><div class="cands">`)
	for place, pl := range scenes.OwletPlaces {
		was := scenes.OwletPlace
		scenes.OwletPlace = place
		f := newFrames(W, H, 7, false, true, 0, 0)
		fmt.Fprintf(&b, `<div class="cand wide"><h3><span class="k">P%d</span> %s</h3><p class="what">%s</p><div class="ctx">`, place, pl.Name, pl.Note)
		clip(fmt.Sprintf("place-%d", place), func(tt float64) string {
			st := reduce.State{Act: act(0.65, true), Pose: companion.Working, Tail: tail, Kittens: 1}
			switch {
			case tt >= 8.0:
				st.Kittens, st.KittenExits = 4, []float64{(tt - 8.0) / 1.5}
			case tt >= 7.1:
				st.Kittens = 5
			case tt >= 6.5:
				st.Kittens = 4
			case tt >= 5.9:
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
		b.WriteString(`</div></div>`)
		scenes.OwletPlace = was
	}
	b.WriteString(`</div></section>`)

	keys := make([]string, 0)
	for _, sc := range secs {
		keys = append(keys, sc.name)
	}
	sort.Strings(keys)
	page := `<meta charset="utf-8"><title>The Owl Moves</title>
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
.bar input[type=range]{flex:1;min-width:160px;max-width:440px;accent-color:var(--acc)}
.bar .fr{color:var(--mute);font-size:13px;font-variant-numeric:tabular-nums;min-width:22ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 18px}
.cands{display:flex;flex-wrap:wrap;gap:26px 30px;align-items:flex-start}
.cand{max-width:100%}
.cand.wide{width:100%}
h3{font-size:15px;font-weight:700;margin:0 0 2px}
.k{color:var(--acc);margin-right:6px}
.pick{color:#f4c67a;font-weight:400;font-size:13px;margin-left:8px}
.what{color:var(--mute);font-size:13px;max-width:44ch;margin:0 0 10px;min-height:2.6em}
.pair{display:flex;gap:14px;align-items:flex-end}
.ctx{overflow-x:auto;max-width:100%}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The owl moves: candidates for every state</h1>
<p class="lede">Round three. Every state carries your pick, marked. The working wings are two now (the right one had been drawn inside the body, where nothing paints). The litter shows all four acts under one schedule, L6, and the new section at the end shows where it can sit and how an owlet comes and goes. Each candidate is a motion laid over the state's face, on a clock, carrying no information. Candidate 0 in every row is today's owl. Every clip runs the same four-second loop at twelve frames a second; every period is compressed to four seconds here so each event shows once a loop, and each note gives the period proposed live. Name them by their letter and number.</p>
<div class="bar"><button type="button" id="play">Pause</button><input type="range" id="sl" min="0" max="` + fmt.Sprint(n-1) + `" value="0" aria-label="frame"><span class="fr" id="fr"></span></div>
` + b.String() + `
<section><h2>the one trade</h2><p class="note">A resting owl that opens its eyes for a glance is, in that second, a screenshot of a working owl. R1 spends 12% of its period that way, R2 12%, R3 17%. Everything else moves the body or the lids and leaves the face's reading alone. The needs-you double blink and the worried alternatives keep their state's eyes throughout.</p></section>
<p class="foot">Made from the tree: <code>XSCAPES_OWLROUND=&lt;file&gt; go test -run TestOwlRound .</code> · 125x28, seed 7, 10:10, the 256-colour cube. Nothing here is live until a pick is set; the round's painter with no pick draws today's owl cell for cell (a test says so).</p>
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
	t.Logf("wrote %s, %d bytes, %d frames a clip", out, len(page), n)
}
