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

// The wind's churn, counted. His report of 2026-09-17, in a five-agent Kimi
// run on the vista: "theres a flicker in the art as you are working (it was
// there before, now is worse)". Kimi's report placed it on the leaves the
// wind carries (forest.go: the whole air from row 4, density 0.4% idle to
// 5.4% at full, 20 columns a second) and the fire's re-roll. This probe
// renders the live vista at his window and counts, frame to frame, how many
// cells change in each band of the picture -- the air above the lake, the
// lake and the meadow, the fire's columns -- at three levels, by night and
// by day; and it counts the leaves in the air. The numbers, not the
// impression, are what the styles are judged on.
//
// It measures two things the eye reads as flicker. Churn: at the live 12
// fps, 20 columns a second is 1.67 cells a frame, so every leaf steps by one
// cell and then two, and 140 of them do it at once. The seam: the painter
// loops its clock every four seconds, and the leaf pattern closes on that
// loop only where 20 x 4 = 80 is the frame's width; at any other width every
// leaf jumps 80 columns at once when the clock wraps, which the max column
// below shows as one frame that changes hundreds of cells.
//
//	go test -run TestTheWindsChurn -v .
type windChurn struct {
	air, near, fire float64 // mean cells changed per frame, by band
	maxAir          int     // the worst single frame in the air
	leaves          float64 // mean leaf glyphs in the air per frame
	// tracks is the share of leaves in a frame that have a leaf in the
	// same cell or one to three cells to their left in the frame before: a
	// leaf that stepped, or held its cell for a frame.
	// A field that jumps as a whole keeps none of its tracks for that one
	// frame while changing no more cells than usual, so this is the number
	// that sees the seam; minTracks is its worst frame.
	tracks, minTracks float64
	// rowsDirty is the mean number of ROWS with any change per frame over
	// the whole frame, owl and all: the host rewrites a row that changed
	// anywhere in it, so this is what reaches the terminal each frame.
	rowsDirty float64
}

// windProbe folds one configuration and returns its churn.
func windProbe(t *testing.T, style int, level, tod float64, W, H int, secs float64) windChurn {
	t.Helper()
	const fps = 12.0
	was := scenes.WindPick
	scenes.WindPick = style
	defer func() { scenes.WindPick = was }()
	st := reduce.State{Act: scape.Activity{Working: true, Level: level, ContextUsed: 0.45, TimeOfDay: tod, TodoDone: 3, Tokens: 17_000_000}, Pose: companion.Working}
	type cell struct {
		r      rune
		fg, bg term.RGB
	}
	var prev []cell
	var out windChurn
	n := 0
	for i := 0; float64(i) < secs*fps; i++ {
		// From t0 = 1 so the first frames are not the clock's own start,
		// and long enough to cross the four-second wrap at least once.
		f := renderVista(t, W, H, st, 1.0+float64(i)/fps)
		lakeTop, meadowTop, bandTop := f.vista.Rows()
		fireX := f.vista.FireX()
		owlX, owlY, _, _ := f.vista.Layout()
		cur := make([]cell, W*H)
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				r, fg, bg := f.c.ResolveAt(x, y, term.Profile256)
				cur[y*W+x] = cell{r, fg, bg}
			}
		}
		if prev != nil {
			var air, near, fire int
			for y := 0; y < H; y++ {
				for x := 0; x < W; x++ {
					if cur[y*W+x] != prev[y*W+x] {
						out.rowsDirty++
						break
					}
				}
			}
			for y := 0; y < bandTop; y++ {
				for x := 0; x < W; x++ {
					if cur[y*W+x] == prev[y*W+x] {
						continue
					}
					// The owl's own motion is its channel, not the wind's;
					// its wings reach a little past the box.
					if x >= owlX-3 && x < owlX+15 && y >= owlY-1 && y < owlY+8 {
						continue
					}
					switch {
					case x >= fireX-7 && x <= fireX+19:
						fire++ // the flames, the smoke leaning right, the sparks
					case y < lakeTop:
						air++
					default:
						near++
					}
				}
			}
			out.air += float64(air)
			out.near += float64(near)
			out.fire += float64(fire)
			if air > out.maxAir {
				out.maxAir = air
			}
			n++
		}
		// Leaves are counted above the meadow, where nothing else is drawn
		// with their glyphs (the scrub bends into ',' and '_').
		isLeaf := func(c cell) bool { return c.r == '\'' || c.r == ',' || c.r == '-' }
		leaves, kept := 0, 0
		for y := 0; y < meadowTop; y++ {
			for x := 0; x < W; x++ {
				if x >= fireX-7 && x <= fireX+19 || !isLeaf(cur[y*W+x]) {
					continue
				}
				leaves++
				if prev == nil {
					continue
				}
				for k := 0; k <= 3 && x-k >= 0; k++ {
					if isLeaf(prev[y*W+x-k]) {
						kept++
						break
					}
				}
			}
		}
		out.leaves += float64(leaves)
		if prev != nil && leaves > 0 {
			fr := float64(kept) / float64(leaves)
			out.tracks += fr
			if i == 1 || fr < out.minTracks {
				out.minTracks = fr
			}
		}
		prev = cur
	}
	out.air /= float64(n)
	out.near /= float64(n)
	out.fire /= float64(n)
	out.leaves /= float64(n + 1)
	out.tracks /= float64(n)
	out.rowsDirty /= float64(n)
	return out
}

// windTable renders every style at every level and hour as one table, for
// the log and for the page.
func windTable(t *testing.T, W, H int) (rows []string, html string) {
	t.Helper()
	var hb strings.Builder
	hb.WriteString(`<table class="lum"><tr><th>style</th><th>level</th><th>hour</th><th>leaves above the meadow</th><th>air: cells changed / frame</th><th>worst air frame</th><th>tracks kept</th><th>worst frame</th><th>lake + meadow</th><th>fire</th><th>rows rewritten / frame</th></tr>`)
	for si, ws := range scenes.WindStyles {
		for _, lv := range []float64{0, 0.5, 1} {
			for _, h := range []struct {
				tod  float64
				name string
			}{{22.0 / 24, "night"}, {0.5, "noon"}} {
				if lv != 1 && h.name == "noon" {
					continue
				}
				c := windProbe(t, si, lv, h.tod, W, H, 6)
				line := fmt.Sprintf("W%d %-24s level %.1f %-5s  leaves %5.1f  air %6.1f/frame (worst %4d)  tracks %3.0f%% (worst %3.0f%%)  near %5.1f  fire %5.1f  rows %4.1f/%d", si, ws.Name, lv, h.name, c.leaves, c.air, c.maxAir, c.tracks*100, c.minTracks*100, c.near, c.fire, c.rowsDirty, H)
				rows = append(rows, line)
				fmt.Fprintf(&hb, `<tr><td>W%d %s</td><td>%.1f</td><td>%s</td><td>%.0f</td><td>%.0f</td><td>%d</td><td>%.0f%%</td><td>%.0f%%</td><td>%.0f</td><td>%.0f</td><td>%.1f of %d</td></tr>`, si, ws.Name, lv, h.name, c.leaves, c.air, c.maxAir, c.tracks*100, c.minTracks*100, c.near, c.fire, c.rowsDirty, H)
			}
		}
	}
	hb.WriteString(`</table>`)
	return rows, hb.String()
}

func TestTheWindsChurn(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "vista")
	rows, _ := windTable(t, 131, 24)
	for _, r := range rows {
		t.Log(r)
	}
	if os.Getenv("XSCAPES_WINDPROBE") != "" {
		fmt.Println(strings.Join(rows, "\n"))
	}
}

// TestWindPage writes the wind candidates as one page when XSCAPES_WINDPAGE
// names a file: every style at his window, 131 by 24, playing at the live
// rate at full stretch by night, at half by night and at full stretch by
// noon, with the churn table. His pick.
func TestWindPage(t *testing.T) {
	out := os.Getenv("XSCAPES_WINDPAGE")
	if out == "" {
		t.Skip("set XSCAPES_WINDPAGE=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 131, 24
	const fps = 12.0
	const n = 36 // three seconds: the page has to stay under the artifact's 16 MB
	pal := &canvas.HTMLPalette{}
	rows, table := windTable(t, W, H)
	for _, r := range rows {
		t.Log(r)
	}
	tail := []reduce.Line{
		{Text: "read  internal/scenes/forest.go", Age: 0.6},
		{Text: "edit  windprobe_test.go  +200", Age: 0.3},
		{Text: "shell go test ./...  ok", Age: 0.0},
	}
	state := func(level, tod float64) reduce.State {
		return reduce.State{Act: scape.Activity{Working: true, Level: level, ContextUsed: 0.45, TimeOfDay: tod, TodoDone: 3, Tokens: 17_000_000}, Pose: companion.Working, Kittens: 2, Tail: tail}
	}
	var b strings.Builder
	was := scenes.WindPick
	defer func() { scenes.WindPick = was }()
	for si, ws := range scenes.WindStyles {
		scenes.WindPick = si
		fmt.Fprintf(&b, `<section class="player" id="w%d" data-n="%d" data-fps="%v"><h2>W%d %s</h2><p class="note">%s</p><div class="bar"><button type="button">Pause</button><input type="range" min="0" max="%d" value="0" aria-label="frame"><span class="fr"></span></div>`, si, n, fps, si, ws.Name, ws.Note, n-1)
		for _, sc := range []struct {
			lv, tod float64
			name    string
		}{{1, 22.0 / 24, "full stretch, night"}, {0.5, 22.0 / 24, "half, night"}, {1, 0.5, "full stretch, noon"}} {
			fmt.Fprintf(&b, `<h3>%s</h3><div class="ctx"><div class="clip">`, sc.name)
			for i := 0; i < n; i++ {
				f := renderVista(t, W, H, state(sc.lv, sc.tod), 1.0+float64(i)/fps)
				_, _, bandTop := f.vista.Rows()
				h := f.c.HTMLFragmentCropClassed(0, 0, W, bandTop, 7, term.Profile256, pal)
				if i == 0 {
					h = strings.Replace(h, "<pre ", `<pre class="on" `, 1)
				}
				b.WriteString(h)
			}
			b.WriteString(`</div></div>`)
		}
		b.WriteString(`</section>`)
	}
	scenes.WindPick = was
	page := `<meta charset="utf-8"><title>The Wind</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:16px}
main{max-width:1240px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
.lede b,.note b{color:var(--ink)}
.bar{display:flex;flex-wrap:wrap;gap:14px;align-items:center;margin:0 0 14px}
.bar button{font:inherit;font-size:13px;background:transparent;color:var(--ink);border:1px solid var(--mute);padding:4px 14px;cursor:pointer}
.bar button:focus-visible,.bar input:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
.bar input[type=range]{flex:1;min-width:200px;max-width:520px;accent-color:var(--acc)}
.bar .fr{font-size:14px;font-variant-numeric:tabular-nums;min-width:12ch}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
h3{font-size:13px;font-weight:700;color:var(--mute);margin:12px 0 6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 14px}
.ctx{overflow-x:auto}
.clip pre{display:none;margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.clip pre.on{display:block}
.lum{border-collapse:collapse;font-size:13px;margin:0 0 16px;font-variant-numeric:tabular-nums;display:block;overflow-x:auto}
.lum th,.lum td{text-align:right;padding:3px 12px 3px 0;border-bottom:1px solid var(--rule);white-space:nowrap}
.lum th:first-child,.lum td:first-child{text-align:left}
.lum th{color:var(--mute);font-weight:400}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The wind, four ways</h1>
<p class="lede">Your report, on the vista under five Kimi agents: a flicker in the art while it works, there before and worse now. Counted at your window, 131 by 24, at the live twelve frames a second: at full stretch the picture changes about <b>216 cells a frame</b>, and <b>seven in ten of them are the leaves</b> the wind carries: single glyphs over the whole air from row 4, 0.4% of the cells idle and 5.4% at full stretch, blown right at twenty columns a second, which at twelve frames is one cell and then two. The fire is not it: its own columns settle at about thirty changes a frame once leaves stop flying through them. A second thing, fixed in every style here including today's: the painter loops its clock every four seconds and the leaf pattern closed on that loop only at 80 columns wide, so on your window every leaf jumped 80 columns at once every four seconds. Each style below plays the whole frame for three seconds at full stretch by night, at half by night, and at full stretch by noon (the clip wraps at three seconds and shows a seam there that the live scape never has); the table under them is the count. Coverage and count still say how hard; only the amount and the speed change. Say the number.</p>
` + b.String() + `
<section><h2>the count</h2><p class="note">Cells changed per frame in each band of the picture, averaged over six seconds at twelve frames a second, two owlets, the owl's own box left out. "Tracks kept" is the share of leaves that held their cell or stepped one to three cells from where a leaf was the frame before; a field that jumps as a whole keeps none for that frame. Leaves are counted above the meadow, where nothing else uses their glyphs.</p>` + table + `</section>
<p class="foot">Made from the tree: <code>XSCAPES_WINDPAGE=&lt;file&gt; go test -run TestWindPage .</code> · 131x24, seed 7, the 256-colour cube, 22:00 where it says night, the context at 45%, three stars, the counter at 17M.</p>
</main>
<script>
(function(){
var reduced=window.matchMedia&&window.matchMedia('(prefers-reduced-motion: reduce)').matches;
[].slice.call(document.querySelectorAll('.player')).forEach(function(pl){
var clips=[].slice.call(pl.querySelectorAll('.clip')),n=+pl.dataset.n,fps=+pl.dataset.fps;
var sl=pl.querySelector('input'),fr=pl.querySelector('.fr'),play=pl.querySelector('button'),k=0,timer=null;
function show(i){k=i;clips.forEach(function(c){var ps=c.children;for(var j=0;j<ps.length;j++){ps[j].className=j===i?'on':'';}});sl.value=i;fr.textContent=(i/fps).toFixed(2)+' s';}
function start(){if(timer)return;timer=setInterval(function(){show((k+1)%n);},1000/fps);play.textContent='Pause';}
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
