package main

import (
	"encoding/json"
	"fmt"
	"math"
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

// TestFirePage writes the campfire's every stage and animation as one page
// when XSCAPES_FIREPAGE names a file, his ask of 2026-09-16: "show me on a
// separate page all the stages/animations of the firepit". The page came
// back with the fire fixed (the 09-14 pick: the wind alone is the work) and
// his word on it was "the wind + fire (both)", so the fire climbs with the
// level now and the page shows that: the five levels side by side, the
// light with the dark (which never follows the work), every frame of the
// flicker, the fire in its frame, and the five stages as stills.
func TestFirePage(t *testing.T) {
	out := os.Getenv("XSCAPES_FIREPAGE")
	if out == "" {
		t.Skip("set XSCAPES_FIREPAGE=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 125, 28
	const fps = 12.0
	const t0 = 5.0
	const night = 22.0 / 24
	pal := &canvas.HTMLPalette{}
	tail := []reduce.Line{
		{Text: "read  internal/scenes/forest.go", Age: 0.6},
		{Text: "edit  firepage_test.go  +200", Age: 0.3},
		{Text: "shell go test ./...  ok", Age: 0.0},
	}
	state := func(level, tod float64, kittens int) reduce.State {
		return reduce.State{Act: scape.Activity{Working: true, Level: level, ContextUsed: 0.45, TimeOfDay: tod, TodoDone: 3}, Pose: companion.Working, Kittens: kittens, Tail: tail}
	}
	// The fire's place at 125x28: its column from the layout, its base on
	// the meadow three rows under the lake, read off one frame.
	probe := renderVista(t, W, H, state(0, night, 0), t0)
	fireX := probe.vista.FireX()
	_, _, _, bandTop := probe.vista.Layout()
	var b strings.Builder
	// player opens a section with its own bar; n frames, labels optional.
	player := func(id, title, note string, n int, labels []string) {
		lb := ""
		if labels != nil {
			j, _ := json.Marshal(labels)
			lb = ` data-labels='` + strings.ReplaceAll(string(j), "'", "&#39;") + `'`
		}
		fmt.Fprintf(&b, `<section class="player" id="%s" data-n="%d" data-fps="%v"%s><h2>%s</h2><p class="note">%s</p><div class="bar"><button type="button">Pause</button><input type="range" min="0" max="%d" value="0" aria-label="frame"><span class="fr"></span></div>`, id, n, fps, lb, title, note, n-1)
	}
	clip := func(id string, n int, render func(i int) string) {
		fmt.Fprintf(&b, `<div class="clip" id="%s">`, id)
		for i := 0; i < n; i++ {
			h := render(i)
			if i == 0 {
				h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
			}
			b.WriteString(h)
		}
		b.WriteString(`</div>`)
	}
	crop := func(f *frames, x0, y0, x1, y1, px int) string {
		return f.c.HTMLFragmentCropClassed(max(0, x0), max(0, y0), min(W, x1), min(H, y1), px, term.Profile256, pal)
	}

	// 1. The wind through the fire: five levels side by side, by night.
	player("wind", "the wind and the fire through the work", `From left to right the agent's activity level is 0, ¼, ½, ¾ and 1, by night. The fire: two rows at rest and six at full stretch, wider as it climbs, sparks thrown from the tip, the smoke thicker. The wind: the flames lean at the tip, the smoke streams sideways, the scrub lies flat and the air fills with what the wind picked up. The light on the meadow is the same in all five; it is the night's, not the work's. Four seconds a loop at the live rate; the debris pattern is the frame's width, so it does not close on the flames' four seconds and shows a seam at the loop.`, 48, nil)
	b.WriteString(`<div class="row">`)
	for _, lv := range []float64{0, 0.25, 0.5, 0.75, 1} {
		fmt.Fprintf(&b, `<div class="cell"><h3>level %v</h3>`, lv)
		clip(fmt.Sprintf("wind-%v", lv), 48, func(i int) string {
			f := renderVista(t, W, H, state(lv, night, 0), t0+float64(i)/fps)
			return crop(f, fireX-6, bandTop-15, fireX+18, bandTop, 15)
		})
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div></section>`)

	// 2. The light through the evening, at rest.
	const evSteps = 25 // 17:00 to 23:00 by quarter hours
	labels := make([]string, evSteps)
	for k := range labels {
		hour := 17 + 0.25*float64(k)
		sv := scape.PaletteAt(hour / 24).StarVis
		lit := math.Max(0, math.Min(1, (sv-0.5)/0.3))
		hh, mm := int(hour), int((hour-float64(int(hour)))*60)
		labels[k] = fmt.Sprintf("%02d:%02d · firelight %d%%", hh, mm, int(lit*100+0.5))
	}
	player("evening", "the light comes up with the dark", `At rest (an ember: the level is a tenth), from 17:00 to 23:00 by quarter hours. The fire lights the meadow around it only as the night comes: nothing by day, rising with the stars' visibility from half to four fifths, full by night. The label says how much of the light is painted at each step. By day the meadow keeps its ramp untouched, which is what closed your dusk note on the dark patch.`, evSteps, labels)
	b.WriteString(`<div class="ctx">`)
	clip("evening-c", evSteps, func(i int) string {
		hour := 17 + 0.25*float64(i)
		f := renderVista(t, W, H, state(0.1, hour/24, 0), t0)
		return crop(f, fireX-18, bandTop-12, fireX+19, bandTop, 18)
	})
	b.WriteString(`</div></section>`)

	// 3. Every frame of the pattern, as stills.
	b.WriteString(`<section><h2>every frame of the flicker</h2><p class="note">The pattern clock runs six frames a second and the flames close on twenty-four of them, four seconds; the smoke closes on twelve. In each frame a flame cell is <code>^</code>, <code>*</code> or <code>)</code> by a hash of its cell and the frame, a fifth of the cells are empty, and the colour is by row: orange at the base, amber above the middle, pale at the tip. The fuel under it is five cells that never change. Working at half, so the flames are four rows; night.</p><div class="strip">`)
	for i := 0; i < 24; i++ {
		f := renderVista(t, W, H, state(0.5, night, 0), float64(i)/6)
		fmt.Fprintf(&b, `<div class="still"><div class="lab">%d</div>%s</div>`, i, crop(f, fireX-6, bandTop-15, fireX+8, bandTop-1, 12))
	}
	b.WriteString(`</div></section>`)

	// 4. In the vista: the whole frame by night, working at half, two owlets.
	player("frame", "in the vista", `The whole live frame at 125x28 by night, working at half with two owlets, one of them at the fire by turns. The fire sits at three eighths of the width, on the meadow; the litter keeps out of it and its smoke.`, 48, nil)
	b.WriteString(`<div class="ctx">`)
	clip("frame-c", 48, func(i int) string {
		f := renderVista(t, W, H, state(0.5, night, 2), t0+float64(i)/fps)
		return f.c.HTMLFragmentClassed(9, term.Profile256, pal)
	})
	b.WriteString(`</div></section>`)

	// 5. The five stages, still.
	b.WriteString(`<section><h2>the fire's five stages</h2><p class="note">One frame at each level, still, night. The height is two rows at rest and six at full stretch (two plus four times the level, rounded), the width follows the height, sparks are thrown from the tip at full stretch and almost never at rest, and the smoke's density runs from a fifth to four fifths. The fuel and the three colours by row never change, and neither does the light.</p><div class="row">`)
	for _, lv := range []float64{0, 0.25, 0.5, 0.75, 1} {
		f := renderVista(t, W, H, state(lv, night, 0), t0)
		fmt.Fprintf(&b, `<div class="cell"><h3>level %v</h3>%s</div>`, lv, crop(f, fireX-6, bandTop-15, fireX+18, bandTop, 15))
	}
	b.WriteString(`</div></section>`)

	// 6. The logs, six ways, for his pick.
	wasLog := scenes.LogPick
	defer func() { scenes.LogPick = wasLog }()
	b.WriteString(`<section><h2>the logs, six ways</h2><p class="note">Your note: polish the logs under the firepit. Six ways to draw the fuel, each at rest by night, at full stretch by night, and at noon working at half. Style 0 is today. The two-row styles keep their second row only while it is still meadow, so a short window loses the rim, not the log. Say the number.</p><div class="row">`)
	for i, ls := range scenes.LogStyles {
		scenes.LogPick = i
		fmt.Fprintf(&b, `<div class="cell"><h3>L%d %s</h3><p class="what">%s</p><div class="trio">`, i, ls.Name, ls.Note)
		for _, sc := range []struct {
			lv, tod float64
		}{{0, night}, {1, night}, {0.5, 0.5}} {
			f := renderVista(t, W, H, state(sc.lv, sc.tod, 0), t0+0.25)
			b.WriteString(crop(f, fireX-7, bandTop-9, fireX+9, bandTop, 20))
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`</div></section>`)
	scenes.LogPick = wasLog

	page := `<meta charset="utf-8"><title>The Campfire</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:24px}
main{max-width:1240px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
.bar{display:flex;flex-wrap:wrap;gap:14px;align-items:center;margin:0 0 14px}
.bar button{font:inherit;font-size:13px;background:transparent;color:var(--ink);border:1px solid var(--mute);padding:4px 14px;cursor:pointer}
.bar button:focus-visible,.bar input:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.bar input[type=range]{flex:1;min-width:200px;max-width:520px;accent-color:var(--acc)}
.bar .fr{font-size:14px;font-variant-numeric:tabular-nums;min-width:12ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
h3{font-size:13px;font-weight:700;color:var(--mute);margin:0 0 6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 14px}
.row{display:flex;flex-wrap:wrap;gap:18px 22px;align-items:flex-start}
.cell{max-width:100%}
.ctx{overflow-x:auto}
.strip{display:flex;flex-wrap:wrap;gap:10px 12px;align-items:flex-start}
.still .lab{color:var(--mute);font-size:12px;margin-bottom:2px}
.what{color:var(--mute);font-size:13px;max-width:40ch;margin:0 0 8px;min-height:2.6em}
.trio{display:flex;gap:6px;align-items:flex-end}
.trio pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.still pre,.cell>pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The campfire, every stage</h1>
<p class="lede">What the fire does in the vista as installed. Since your word of 2026-09-16 the work is the wind and the fire both: the flames climb from an ember to a blaze, throw sparks and thicken the smoke as the level rises, while the wind leans everything. What stays the night's is the light on the meadow, which never follows the work. Each section plays on its own.</p>
` + b.String() + `
<p class="foot">Made from the tree: <code>XSCAPES_FIREPAGE=&lt;file&gt; go test -run TestFirePage .</code> · 125x28, seed 7, the 256-colour cube, 22:00 where it says night.</p>
</main>
<script>
(function(){
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
[].slice.call(document.querySelectorAll('.player')).forEach(function(pl){
var clips=[].slice.call(pl.querySelectorAll('.clip')),n=+pl.dataset.n,fps=+pl.dataset.fps,labels=pl.dataset.labels?JSON.parse(pl.dataset.labels):null;
var sl=pl.querySelector('input'),fr=pl.querySelector('.fr'),play=pl.querySelector('button'),k=0,timer=null;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;fr.textContent=labels?labels[i]:(i/fps).toFixed(2)+' s';}
function start(){if(timer)return;timer=setInterval(function(){show((k+1)%n);},labels?400:1000/fps);play.textContent='Pause';}
function stop(){clearInterval(timer);timer=null;play.textContent='Play';}
play.addEventListener('click',function(){timer?stop():start();});
sl.addEventListener('input',function(){stop();show(+sl.value);});
show(0);if(reduced){play.textContent='Play';}else{start();}
});
})();
</script>`
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s, %d bytes", out, len(page))
}
