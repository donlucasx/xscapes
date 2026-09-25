// rainshots renders the rainy window as standalone frame pages, so a REAL
// look is one command away (the vistashots pattern):
//
//	go run ./notes/rainshots <dir>
//	for f in <dir>/*.html; do "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
//	  --headless=new --disable-gpu --use-mock-keychain --hide-scrollbars \
//	  --window-size=1500,1400 --screenshot="${f%.html}.png" "file://$f"; done
//
// Profile256 throughout: what a terminal draws, not the page's truecolor.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

const style = `<meta charset="utf-8"><style>html,body{margin:0;padding:12px;background:#000;color:#ccc;font:13px Menlo,monospace}pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}.row{display:flex;gap:10px;align-items:flex-start;margin-bottom:10px}.lab{width:120px;padding-top:50px;font-weight:bold}h2{font:bold 14px Menlo,monospace;color:#eee;margin:18px 0 8px;border-top:1px solid #333;padding-top:10px}</style>`

func paint(sc *scenes.Scene, tod, level float64, t float64) *canvas.Canvas {
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	c.Clear()
	sc.Paint(c, tod, t, level, 7)
	return c
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: rainshots <dir>")
		os.Exit(1)
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	sc := scenes.Find("Rainy window")
	b := style + `<h2>the rainy window as drawn today (the study painter): noon, dusk, night; idle then working hard</h2>`
	for _, lv := range []float64{0.1, 0.85} {
		b += `<div class="row"><div class="lab">level ` + fmt.Sprint(lv) + `</div>`
		for _, tod := range []float64{0.5, 0.78, 0.0245} {
			b += `<div>` + paint(sc, tod, lv, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
		}
		b += `</div>`
	}
	os.WriteFile(dir+"/today.html", []byte(b), 0o644)
	fmt.Println("wrote", dir+"/today.html")

	// 2. Each skyline: the roof stage at three hours, then the sill at
	// night and the fire escape at dusk and night (his notes' frames).
	names := []string{"midtown", "row", "bridge", "anycity"}
	for sk := range names {
		scenes.RainSkyline = sk
		b := style + `<h2>skyline ` + names[sk] + `: B the roof at noon, dusk, night (level 0.6, three kittens)</h2><div class="row">`
		for _, tod := range []float64{0.5, 0.78, 0.0245} {
			b += `<div>` + paint(&scenes.RainStudy[1], tod, 0.6, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
		}
		b += `</div><h2>A the sill at night · C the fire escape at dusk · C at night</h2><div class="row">`
		b += `<div>` + paint(&scenes.RainStudy[0], 0.0245, 0.6, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
		b += `<div>` + paint(&scenes.RainStudy[2], 0.78, 0.6, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
		b += `<div>` + paint(&scenes.RainStudy[2], 0.0245, 0.6, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
		b += `</div>`
		f := fmt.Sprintf("%s/sky-%s.html", dir, names[sk])
		os.WriteFile(f, []byte(b), 0o644)
		fmt.Println("wrote", f)
	}
	scenes.RainSkyline = 0

	// 3. The four views, each at three hours, the roof stage's kittens.
	{
		vnames := []string{"", "rooftops", "the canyon", "the park", "the harbour"}
		b := style
		for v := 1; v <= 4; v++ {
			scenes.RainView = v
			b += `<h2>view ` + vnames[v] + `: noon, dusk, night (level 0.6, three kittens)</h2><div class="row">`
			for _, tod := range []float64{0.5, 0.78, 0.0245} {
				b += `<div>` + paint(&scenes.RainStudy[1], tod, 0.6, 1.0).HTMLFragmentAs(11, term.Profile256) + `</div>`
			}
			b += `</div>`
		}
		scenes.RainView = 0
		os.WriteFile(dir+"/views.html", []byte(b), 0o644)
		fmt.Println("wrote", dir+"/views.html")
	}
	review(dir)
	clips(dir)
	if ed := os.Getenv("XSCAPES_RAINEXPORT"); ed != "" {
		export(ed)
	}
}

// export writes each clip for the video stage (notes/qt/render-fire.mjs
// --name <key>): <dir>/<key>.js holding window.TRAILER.<key> = {cols, rows,
// fps, secs, frames, css}, every frame re-rendered at 30 fps.
func export(dir string) {
	companion.Near = 1
	os.MkdirAll(dir, 0o755)
	scenes.RainClipFPS = 30
	n := int(scenes.RainClipSecs * scenes.RainClipFPS)
	for _, cv := range clipViews {
		pal := &canvas.HTMLPalette{}
		frames := make([]string, n)
		for i := 0; i < n; i++ {
			frames[i] = frameAt(cv.view, float64(i)/scenes.RainClipFPS).HTMLFragmentClassed(14, term.Profile256, pal)
		}
		b, _ := json.Marshal(map[string]any{
			"cols": 80, "rows": 24, "fps": int(scenes.RainClipFPS), "secs": scenes.RainClipSecs,
			"frames": frames, "css": pal.CSS(),
		})
		p := dir + "/" + cv.key + ".js"
		os.WriteFile(p, []byte("window.TRAILER=window.TRAILER||{};window.TRAILER."+cv.key+"="+string(b)+";"), 0o644)
		fmt.Printf("export %s: %d frames at %d fps, %d bytes\n", cv.key, n, int(scenes.RainClipFPS), len(b))
	}
	scenes.RainClipFPS = 10
}

// clipViews are the three views that go to animation (his ruling), with
// the view index each takes in the painters.
var clipViews = []struct {
	name, key string
	view      int
	note      string
}{
	{"The block", "block", 0, "Out of the roof's bulkhead door and back in. The beacon blinks, a window switches now and then, a plane crosses with its lights, smoke drifts off a chimney."},
	{"The park", "park", 3, "In from the terrace's side. The fountain's jet and the trees' sway are the work with the rain; the plaza's bandstand and lamps light as the night comes."},
	{"The harbour", "harbour", 4, "Off the moored boat onto the dock. The ferry crosses, the tug works, the sailboats drift, the mast rocks, gulls wheel while there is light, the lighthouse flashes once the night is in."},
}

// frameAt renders one clip frame at time t.
func frameAt(view int, t float64) *canvas.Canvas {
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scenes.PaintRainFrame(c, view, t, 7)
	return c
}

// clips writes the animated review page: each view's clip as frames the
// page plays at RainClipFPS, one shared palette a clip, and a beats page of
// stills at the arc's moments for a look before publishing.
func clips(dir string) {
	companion.Near = 1 // the cat's walk-up one rung, his ruling
	scenes.RainClipFPS = 10
	n := int(scenes.RainClipSecs * scenes.RainClipFPS)
	var page strings.Builder
	page.WriteString(`<title>The Rainy Window, Animated</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#0c0e12;--panel:#14171d;--ink:#d8dcd6;--dim:#8a919c;--rule:#262b34;--rain:#87afd7;--lamp:#ffd787;color-scheme:dark}
body{background:var(--bg);color:var(--ink);font:14px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding:0 16px;padding-block:28px 64px}
main{max-width:1200px;margin:0 auto}
.col{max-width:66ch}
h1{font-size:22px;font-weight:700;letter-spacing:-.01em;margin:0 0 6px;text-wrap:balance}
h2{font-size:16px;font-weight:700;margin:44px 0 4px}
.k{font-size:11px;letter-spacing:.12em;text-transform:uppercase;color:var(--rain);margin:0 0 4px}
p{margin:6px 0}
.dim{color:var(--dim)}
ul{margin:8px 0 0;padding-left:1.2em}li{margin:3px 0}
.clipwrap{overflow-x:auto;padding:14px 0 6px}
.clip{display:inline-block;background:#000;padding:2px;border:1px solid var(--rule);cursor:pointer}
pre{margin:0;font-family:"JetBrains Mono",Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.beats{display:flex;gap:10px;font-size:12px;color:var(--dim);flex-wrap:wrap;margin:6px 0 0}
.beats span{padding:2px 8px;border:1px solid var(--rule);border-radius:3px}
.beats span.on{color:var(--ink);border-color:var(--rain)}
.rule{border:0;border-top:1px solid var(--rule);margin:40px 0 0}
</style>
<main>
<div class="col">
<p class="k">xscapes · the third scape, before it is built · the animated mockups · 2026-09-24</p>
<h1>The Rainy Window, Animated</h1>
<p>Three views, each as a session arc of sixteen seconds at ten frames a second, dusk into night: idle, a prompt, the work rising, three subagents out through the door, an ask with the cat's one-rung walk-up, the work again, the finish, the litter home, idle. The rain is drops running down the glass, born only while the agent works, so the pane is dry at idle and dries after the finish. The sky is greyed toward overcast; the shore's disc sinks behind the skyline as the context fills. Every frame on the 256-colour cube. Click a clip to pause it.</p>
<p class="dim">Nothing is wired: the arc is a script, the same session in all three. The beats under each clip light as the clip passes them.</p>
</div>
`)
	beats := []struct {
		at   float64
		name string
	}{{0, "idle"}, {2, "prompt"}, {4, "subagents"}, {8, "ask"}, {10, "answered"}, {12.5, "finish"}, {14, "idle, night"}}
	var script strings.Builder
	script.WriteString("<script>\nconst FPS=" + fmt.Sprint(int(scenes.RainClipFPS)) + ";const CLIPS={};\n")
	var css strings.Builder
	// ONE palette for all three clips: a palette a clip gave each clip its
	// own .p0, .p1, ... and the last CSS written (the harbour's) recoloured
	// the block and the park ("the park and the block are broken", his look
	// at V1).
	pal := &canvas.HTMLPalette{}
	for _, cv := range clipViews {
		frames := make([]string, n)
		for i := 0; i < n; i++ {
			t := float64(i) / scenes.RainClipFPS
			frames[i] = frameAt(cv.view, t).HTMLFragmentClassed(10, term.Profile256, pal)
		}
		// The frames go in a file beside the page (17 MB inline was over
		// the page's 16 MB limit): the page fetches it and plays.
		enc, _ := json.Marshal(frames)
		os.WriteFile(dir+"/clip-"+cv.key+".json", enc, 0o644)
		page.WriteString(`<hr class="rule"><p class="k">clip</p><h2 style="margin-top:0">` + cv.name + `</h2><p class="col dim">` + cv.note + `</p>`)
		page.WriteString(`<div class="clipwrap"><div class="clip" id="clip-` + cv.key + `">` + frames[0] + `</div></div><div class="beats" id="beats-` + cv.key + `">`)
		for _, b := range beats {
			page.WriteString(fmt.Sprintf(`<span data-at="%g">%s</span>`, b.at, b.name))
		}
		page.WriteString(`</div>`)
	}
	css.WriteString(pal.CSS())
	script.WriteString(`function play(key,frames){const el=document.getElementById('clip-'+key);const bs=document.querySelectorAll('#beats-'+key+' span');let i=0,run=true;el.addEventListener('click',()=>{run=!run});setInterval(()=>{if(!run)return;i=(i+1)%frames.length;el.innerHTML=frames[i];const t=i/FPS;let cur=-1;bs.forEach((b,j)=>{if(t>=parseFloat(b.dataset.at))cur=j});bs.forEach((b,j)=>b.classList.toggle('on',j===cur));},1000/FPS);}
for(const key of ['block','park','harbour']){fetch('clip-'+key+'.json').then(r=>r.json()).then(f=>play(key,f)).catch(e=>{const el=document.getElementById('clip-'+key);el.insertAdjacentHTML('afterend','<p class="dim">the clip file did not load: '+e+'</p>');});}
</script>`)
	page.WriteString(`<div class="col"><p class="dim" style="margin-top:36px;font-size:12px">Rendered by the product's own canvas, Profile256, seed 7. Study painters in <code>internal/scenes/rain_anim.go</code>, <code>rain_views.go</code>, <code>rain_study.go</code>; nothing wired to the product.</p></div></main>` + "\n<style>" + css.String() + "</style>\n" + script.String())
	os.WriteFile(dir+"/clips.html", []byte(page.String()), 0o644)
	fmt.Println("wrote", dir+"/clips.html")

	// The beats page: stills at the arc's moments, for a look. Each beat is
	// reached by rendering every frame up to it, so the cat's approach has
	// the progress the clip gives it.
	upTo := func(view int, t float64) *canvas.Canvas {
		var c *canvas.Canvas
		for i := 0; float64(i)/scenes.RainClipFPS <= t+1e-9; i++ {
			c = frameAt(view, float64(i)/scenes.RainClipFPS)
		}
		return c
	}
	b := style
	for _, cv := range clipViews {
		b += `<h2>` + cv.name + `: idle 1 s · rising 4.6 s · working 6.5 s · ask 9 s · working 11.5 s · finish 13 s · idle 15.5 s</h2><div class="row">`
		for _, t := range []float64{1, 4.6, 6.5, 9, 11.5, 13, 15.5} {
			b += `<div>` + upTo(cv.view, t).HTMLFragmentAs(9, term.Profile256) + `</div>`
		}
		b += `</div>`
	}
	os.WriteFile(dir+"/beats.html", []byte(b), 0o644)
	fmt.Println("wrote", dir+"/beats.html")
	// Two close looks: the park at 1 s and the block at 9 s.
	cl := style + `<div class="row">` + upTo(3, 1).HTMLFragmentAs(16, term.Profile256) + upTo(3, 11.5).HTMLFragmentAs(16, term.Profile256) + `</div>`
	os.WriteFile(dir+"/close-beats.html", []byte(cl), 0o644)
}

func init() {
	// A close look at one frame: XSCAPES_RAINCLOSE=<b|c|a> writes close.html
	// at 16px for the stage named, dusk, level 0.6, three kittens.
	if v := os.Getenv("XSCAPES_RAINCLOSE"); v != "" && len(os.Args) > 1 {
		i := int(v[0] - 'a')
		if i >= 0 && i < len(scenes.RainStudy) {
			sc := &scenes.RainStudy[i]
			b := style + `<div class="row">` + paint(sc, 0.78, 0.6, 1.0).HTMLFragmentAs(18, term.Profile256) + paint(sc, 0.0245, 0.6, 1.0).HTMLFragmentAs(18, term.Profile256) + `</div>`
			os.MkdirAll(os.Args[1], 0o755)
			os.WriteFile(os.Args[1]+"/close-"+v+".html", []byte(b), 0o644)
		}
	}
}

// review writes the review page: the park three ways, for his pick.
func review(dir string) {
	fr := func(tod, level float64) string {
		was := scenes.RainKittens
		scenes.RainKittens = 0
		defer func() { scenes.RainKittens = was }()
		wv := scenes.RainView
		scenes.RainView = 3
		defer func() { scenes.RainView = wv }()
		return `<div class="fr">` + paint(&scenes.RainStudy[1], tod, level, 1.0).HTMLFragmentAs(10, term.Profile256) + `</div>`
	}
	rows := func(style int) string {
		was := scenes.RainParkStyle
		scenes.RainParkStyle = style
		defer func() { scenes.RainParkStyle = was }()
		return `<div class="row">` + fr(0.5, 0.6) + fr(0.78, 0.6) + fr(0.0245, 0.6) + `</div>` +
			`<div class="cap"><span>noon, working</span><span>dusk, working</span><span>night, working</span></div>` +
			`<div class="row two">` + fr(0.78, 0.05) + fr(0.78, 0.95) + `</div>` +
			`<div class="cap two"><span>dusk, idle</span><span>dusk, flat out</span></div>`
	}
	var b strings.Builder
	b.WriteString(`<title>The Park, Three Ways</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#0c0e12;--panel:#14171d;--ink:#d8dcd6;--dim:#8a919c;--rule:#262b34;--rain:#87afd7;--lamp:#ffd787;color-scheme:dark}
body{background:var(--bg);color:var(--ink);font:14px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding:0 16px;padding-block:28px 64px}
main{max-width:1500px;margin:0 auto}
.col{max-width:66ch}
h1{font-size:22px;font-weight:700;letter-spacing:-.01em;margin:0 0 6px;text-wrap:balance}
h2{font-size:16px;font-weight:700;margin:44px 0 4px}
.k{font-size:11px;letter-spacing:.12em;text-transform:uppercase;color:var(--rain);margin:0 0 4px}
p{margin:6px 0}
.dim{color:var(--dim)}
blockquote{margin:10px 0;padding:8px 14px;border-left:2px solid var(--rule);color:var(--dim);font-style:italic}
.row{display:flex;gap:14px;overflow-x:auto;padding:14px 0 6px;align-items:flex-start}
.fr{flex:0 0 auto;background:#000;padding:2px;border:1px solid var(--rule)}
pre{margin:0;font-family:"JetBrains Mono",Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.cap{display:flex;gap:14px;font-size:12px;color:var(--dim);margin-bottom:6px}
.cap span{flex:0 0 auto;width:494px}
.pick{border-left:3px solid var(--lamp);padding:10px 14px;background:var(--panel);margin-top:18px}
.pick b{color:var(--lamp)}
.rule{border:0;border-top:1px solid var(--rule);margin:40px 0 0}
@media (max-width:700px){.cap span{width:auto}}
</style>
<main>
<div class="col">
<p class="k">xscapes · the third scape, before it is built · the park, three ways · 2026-09-25</p>
<h1>The Park, Three Ways</h1>
<p>Three ways of making the trees and the fountain, each borrowed from something the product already does well, beside the blob crowns of the last rounds. Stills: noon, dusk and night working, then dusk idle and flat out for the fountain. No kittens, so the park is what you look at. Every frame on the 256-colour cube.</p>
<blockquote>"the fountain / park needs work- its a lot of detail and currently does not read great. Try a couple different approaches for the park scene and how you are generating the trees and the fountain within"</blockquote>
</div>

<hr class="rule"><p class="k">As it is</p><h2 style="margin-top:0">Blob crowns</h2>
<p class="col dim">The last rounds: crowns as ellipses with a lit side, kinds of tree, the paved plaza, the bandstand.</p>
` + rows(0) + `

<hr class="rule"><p class="k">1</p><h2 style="margin-top:0">Silhouette, the vista's way</h2>
<p class="col dim">A treeline of bumps in one tone, their caps a shade lighter, swaying with the wind; two black crowns framing the view; a plain pool on the lawn with a single white plume; two lamps. No plaza. Few tones, big shapes.</p>
` + rows(1) + `

<hr class="rule"><p class="k">2</p><h2 style="margin-top:0">Sprites, the animals' way</h2>
<p class="col dim">Authored tree bitmaps through the same quadrant pipeline as the cat: round trees, conifers and poplars with their trunks, drawn far to near with the animals' own rim between them; a stone fountain in tiers with its water in glyphs.</p>
` + rows(2) + `

<hr class="rule"><p class="k">3</p><h2 style="margin-top:0">Glyphs, the ASCII way</h2>
<p class="col dim">Foliage as masses of &amp; % @ in two greens on the lawn, conifers of ^, trunks of |, a pool as a ring of ~, the jet in glyphs. The most terminal of the three.</p>
` + rows(3) + `

<div class="col pick">
<p><b>My recommendation: 2, the sprites.</b> Looking at the three rendered rather than at the idea of them: the sprite park is the one where you can name the trees, and a park is its trees; the fountain in stone tiers reads as a fountain before its water moves. It would want two or three more tree bitmaps so the same crown does not repeat, which is a morning's drawing. 1 reads as a lawn with a hedge behind it, calm but not a park; 3 is a taste, and a strong one.</p>
</div>
<div class="col"><p class="dim" style="margin-top:36px;font-size:12px">Rendered by the product's own canvas, Profile256, seed 7. Study painters in <code>internal/scenes/rain_park.go</code>; nothing wired to the product.</p></div>
</main>
`)
	os.WriteFile(dir+"/review.html", []byte(b.String()), 0o644)
	fmt.Println("wrote", dir+"/review.html")
}
