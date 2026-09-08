package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// page5 is the Hero page: his pick, everything built for it in one place, and
// the exit queue -- the one live behaviour that had nothing behind it.
func page5(seed int64, frames int, fps float64) string {
	col := coats[0].col
	s := &shapes[1] // hero
	up := [][]string{heroWork, heroWork2}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	.note{font:12px/1.5 ui-sans-serif,system-ui;color:#8a8a99;max-width:36ch;margin:2px 0 0}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 6px;padding-top:14px;border-top:1px solid #23232a}
	h3{font:600 12px ui-monospace,monospace;color:#b8b8c4;margin:22px 0 6px}
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#b8b8c4;margin:8px 0 18px}
	td,th{border:1px solid #2a2a32;padding:3px 9px;text-align:left}
	th{color:#d8d8e0;font-weight:500}
	.stage{position:relative}.fr{display:none}.fr:first-child{display:block}
	.full{margin:0 0 26px}
	.ok{color:#a8d8a8}.no{color:#e0a0a0}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; Hero</h1>`)
	b.WriteString(`<p class="nt">His pick, 2026-09-07: <b>&ldquo;hero&rdquo;</b>. Claws raised high on visible arms, ` +
		`clear of the shell. Everything built for it in one place, plus the exit queue &mdash; the one behaviour the ` +
		`live product calls that the crab had no answer for.</p>`)
	b.WriteString(`<p class="nt">Coat is salmon, index 210, which is what his brief asked for and what survives the ` +
		`glyph boost unchanged. <b>Not yet locked</b> &mdash; the five alternatives are still on the round-one page.</p>`)

	// ---- the exit queue ------------------------------------------------------
	b.WriteString(`<h2>1 &middot; The exit queue &mdash; NEW</h2>`)
	b.WriteString(`<p class="nt">When a subagent finishes, its crablet leaves. The cat's version swims along the ` +
		`surface toward the far edge and fades. <b>A crab cannot do that and should not:</b> it walks sideways into ` +
		`the surf and goes under, so the leaving is carried by the shell sinking as well as by travel. The stalks ` +
		`are the last thing to go &mdash; the same pose the swim state is built on, read in the one direction that ` +
		`means gone.</p>`)
	b.WriteString(`<p class="nt">Progress runs 0 to 1 and the reducer owns it; the litter count drops the moment the ` +
		`end event arrives. Position and depth carry it, so a still frame reads the leaving. That is the encoding ` +
		`rule doing its job: never in rate.</p>`)
	b.WriteString(`<h3>One crablet leaving, frame by frame</h3><div class="row">`)
	for _, p := range []float64{0, 0.2, 0.4, 0.6, 0.8, 0.95} {
		c, px, py, seaTop, seaBot := scene(72, 22, 0.778, 3, seed, 0.5)
		drawPosed(c.Near(), s, heroWork, 'o', eyeShine, col, px, py)
		drawExits(c.Near(), col, []float64{p}, px, c.W, seaTop, seaBot, 3, seed)
		_ = py
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%.0f%%</b></div><div class="win">%s</div></div>`,
			p*100, c.HTMLFragmentAs(9, term.Profile256))
	}
	b.WriteString(`</div><p class="nt">At 0 it is beside the litter with the crest up. By 60% the shell is under and ` +
		`only the stalks are reading. By 95% it is a fading ripple. At 1 it is gone and the count has already ` +
		`dropped.</p>`)

	b.WriteString(`<h3>Running, with three leaving at once</h3><div class="row">`)
	var st strings.Builder
	for f := 0; f < frames*2; f++ {
		t := 3 + float64(f)/fps
		prog := float64(f) / float64(frames*2)
		c, px, py, seaTop, seaBot := scene(88, 24, 0.778, t, seed, 0.55)
		drawPosed(c.Near(), s, up[f%2], 'o', eyeShine, col, px, py)
		sand, water := split(5, seed)
		drawLitter(c.Near(), s, col, px, py, sand, t, seed)
		drawSwimmers(c.Near(), col, water, c.W, seaTop, seaBot, t, seed)
		drawExits(c.Near(), col, []float64{prog, prog * 0.7, prog * 0.45}, px, c.W, seaTop, seaBot, t, seed)
		fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(11, term.Profile256))
	}
	fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>five working, three leaving</b></div><div class="stage">%s</div></div>`, st.String())
	b.WriteString(`</div>`)

	// ---- states --------------------------------------------------------------
	b.WriteString(`<h2>2 &middot; Hero, all five states</h2>`)
	for _, p := range posesFor("hero") {
		var ps strings.Builder
		for f := 0; f < frames; f++ {
			t := 3 + float64(f)/fps
			c, px, py, _, _ := scene(72, 22, 0.778, t, seed, 0.6)
			eye := p.eye
			if p.blink && f == frames-2 {
				eye = '-'
			}
			drawPosed(c.Near(), s, p.uppers[f%len(p.uppers)], eye, p.eyeCol, col, px, py)
			if p.say != "" {
				rows, bc := companion.DoneBubble(p.say), bubbleCol
				if p.ask {
					rows, bc = companion.Bubble(p.say), bubbleAskCol
				}
				(&companion.Sprite{Rows: rows, Body: bc, Opaque: true}).Draw(c.Near(), px-2, py-len(rows))
			}
			fmt.Fprintf(&ps, `<div class="fr">%s</div>`, c.HTMLFragmentAs(13, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="row"><div class="col"><div class="lbl"><b>%s</b></div>`+
			`<div class="stage">%s</div></div><p class="note">%s</p></div>`, p.name, ps.String(), p.note)
	}

	// ---- readiness -----------------------------------------------------------
	b.WriteString(`<h2>3 &middot; What is left before it can ship</h2>`)
	b.WriteString(`<p class="nt">The live product calls exactly four things on the companion (live.go 365-378). ` +
		`Everything else in the cat's API is studies only.</p>`)
	b.WriteString(`<table><tr><th>live call</th><th>Hero</th><th></th></tr>` +
		`<tr><td>Draw(near, x, y, t, Pose)</td><td class="ok">done</td><td>all five states</td></tr>` +
		`<tr><td>Size()</td><td class="ok">done</td><td>12 x 7, same box as the cat</td></tr>` +
		`<tr><td>DrawKittenExits(...)</td><td class="ok">done</td><td>this page, section 1</td></tr>` +
		`<tr><td>DrawKittens(...)</td><td class="no">partial</td><td>needs the size ladder with hysteresis and the real lane spans</td></tr>` +
		`</table>`)
	b.WriteString(`<p class="nt">Then two structural things. <b>There is no companion interface</b> &mdash; ` +
		`<code>companion.NewCat()</code> is concrete at 22 call sites, so nothing in the product can construct ` +
		`anything but a cat. How big that is depends on whether Hero <i>replaces</i> the cat or <i>joins</i> it as a ` +
		`choice. And the balloon pointer: <code>MirrorTail</code> puts the <code>v</code> under the cat's right ` +
		`shoulder, and Hero has an arm there rather than a shoulder, so it needs its own anchor cell.</p>`)
	b.WriteString(`<p class="nt">Measured and NOT a problem: the state machine needs no change (the reducer already ` +
		`emits a companion-agnostic pose) &middot; narrow panes cost nothing (cat and crab both stay whole to 12 ` +
		`columns, both first lose ink at 11) &middot; the swim lane geometry is free (same 10x8 box as KittenSwim) ` +
		`&middot; and <code>DrawWalk</code> is not required at all &mdash; live.go never walks the companion.</p>`)

	// ---- full beach ----------------------------------------------------------
	fmt.Fprintf(&b, `<h2>4 &middot; The full beach at %dx%d</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">His window. Nine subagents split between the sand and the sea by the shipped rule, ` +
		`two more on their way out.</p>`)
	var fs strings.Builder
	for f := 0; f < 6; f++ {
		t := 3 + float64(f)/3
		c, px, py, seaTop, seaBot := scene(fullW, fullH, 0.778, t, seed, 0.6)
		drawPosed(c.Near(), s, up[f%2], 'o', eyeShine, col, px, py)
		sand, water := split(9, seed)
		drawLitter(c.Near(), s, col, px, py, sand, t, seed)
		drawSwimmers(c.Near(), col, water, c.W, seaTop, seaBot, t, seed)
		e := float64(f) / 6
		drawExits(c.Near(), col, []float64{e, e*0.6 + 0.2}, px, c.W, seaTop, seaBot, t, seed)
		fmt.Fprintf(&fs, `<div class="fr">%s</div>`, c.HTMLFragmentAs(7, term.Profile256))
	}
	fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>Hero</b> <i>dusk, nine out, two leaving</i></div>`+
		`<div class="stage win">%s</div></div>`, fs.String())

	fmt.Fprintf(&b, `<script>
document.querySelectorAll('.stage').forEach(function(st){
  var fr = Array.prototype.slice.call(st.children), i = 0;
  setInterval(function(){
    fr[i].style.display = 'none';
    i = (i + 1) %% fr.length;
    fr[i].style.display = 'block';
  }, %d);
});
</script>`, int(1000/fps))
	_ = canvas.AlphaNear
	return canvas.HTMLPage("xscapes — Hero", b.String())
}
