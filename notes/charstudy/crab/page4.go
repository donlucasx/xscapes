package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// scene paints the shore and hands back the canvas plus the companion's origin
// and the two rows the sea runs between.
func scene(w, h int, tod, t float64, seed int64, level float64) (*canvas.Canvas, int, int, int, int) {
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: level, TimeOfDay: tod, ContextUsed: 0.3, TodoDone: 2, TodoTotal: 5}
	for k := 0; k < 10; k++ {
		sh.Update(c, t+float64(k)/40, act)
	}
	px, py := c.W-12-(2+c.W/32), c.H-2-7
	return c, px, py, int(float64(c.H)*0.42) + 1, sh.SandTop() - 2
}

func page4(seed int64, frames int, fps float64) string {
	col := coats[0].col
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
	pre.d{font:11px/1.35 ui-monospace,monospace;color:#8a8a99;background:#16161c;border:1px solid #23232a;padding:10px 12px;overflow-x:auto}
	.stage{position:relative}.fr{display:none}.fr:first-child{display:block}
	.full{margin:0 0 26px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the crablets: eyes, and the water</h1>`)
	b.WriteString(`<p class="nt">Two gaps he caught: the sub companions had no eyes, and there was no answer at all for ` +
		`what a crab does when its subagent goes in the sea. Both below, for Pincer and Hero, salmon throughout.</p>`)

	// ---- 1. eyes -------------------------------------------------------------
	b.WriteString(`<h2>1 &middot; The crablets have eyes now</h2>`)
	b.WriteString(`<p class="nt">Two sockets, plotted as characters over the shell in a dimmer green than the parent's ` +
		`&mdash; the kittens use a dimmer eye for the same reason, so a litter never out-shouts the companion it ` +
		`belongs to. Each crablet blinks on <b>its own period</b>, not merely its own phase, so a row of them never ` +
		`falls into step. That is what makes a litter read as alive rather than as a stamped repeat.</p>`)
	b.WriteString(`<p class="nt">Bodies are drawn first and every face after, across the whole litter. That ordering is ` +
		`not fussiness: a neighbour's sprite seam will take an eye cell if the faces go down as each body does, ` +
		`and losing an eye that way is a defect he has already reported once on the kittens.</p><div class="row">`)
	for i := range shapes {
		s := &shapes[i]
		for _, n := range []int{2, 4} {
			c, px, py, _, _ := scene(80, 24, 0.778, 3, seed, 0.5)
			up := pincerWork
			if s.key == "hero" {
				up = heroWork
			}
			drawPosed(c.Near(), s, up, 'o', eyeShine, col, px, py)
			idx := make([]int, n)
			for k := range idx {
				idx[k] = k + 1
			}
			drawLitter(c.Near(), s, col, px, py, idx, 3, seed)
			fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%d on the sand</i></div><div class="win">%s</div></div>`,
				s.name, n, c.HTMLFragmentCropAs(px-30, py-1, px+12+1, py+7+1, 19, term.Profile256))
		}
	}
	b.WriteString(`</div>`)

	// ---- 2. the water --------------------------------------------------------
	b.WriteString(`<h2>2 &middot; In the water</h2>`)
	b.WriteString(`<p class="nt"><b>A crab does not swim.</b> It walks in until the water closes over the shell and the ` +
		`eyestalks carry on above the surface. That is what a real one does, and it is also the thing that survives ` +
		`at five cells by two &mdash; the kitten's swim sprite is a head and shoulders, and the crab's is a shell ` +
		`crest with two periscopes.</p>`)
	b.WriteString(`<pre class="d">KittenSwim  (10 x 8 px)        crabletSwim (10 x 8 px)
  .##....##.                     ..#....#..     the stalks stay up
  .##....##.                     ..#....#..
  .########.                     ..#....#..
  .########.                     ..######..     the crest breaks the surface
  .##.##.##.                     .########.
  .##.##.##.                     .########.
  .########.                     ..######..
  ..######..                     ...####...

Same box, so a crablet takes exactly the room in a lane that a kitten
takes and the shipped lane logic needs no new numbers.</pre>`)
	b.WriteString(`<p class="nt">The bob happens <b>inside</b> the sprite box, by shifting the source rows down and ` +
		`letting the crest clip off the bottom &mdash; the kittens do it the same way, because moving the whole ` +
		`sprite up would put it in the lane above. The effect is the one the pose was chosen for: <b>the shell sinks ` +
		`and comes back, and the eyes never move.</b> Two <code>~</code> at half alpha either side carry the ` +
		`waterline.</p>`)

	swimRow := func(s *shape, n int, tod float64, label string) {
		var st strings.Builder
		for f := 0; f < frames; f++ {
			t := 3 + float64(f)/fps
			c, px, py, seaTop, seaBot := scene(80, 24, tod, t, seed, 0.55)
			up := pincerWork
			if s.key == "hero" {
				up = heroWork
			}
			drawPosed(c.Near(), s, up, 'o', eyeShine, col, px, py)
			idx := make([]int, n)
			for k := range idx {
				idx[k] = k + 1
			}
			drawSwimmers(c.Near(), col, idx, c.W, seaTop, seaBot, t, seed)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(12, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%s</i></div><div class="stage">%s</div></div>`,
			s.name, label, st.String())
	}
	b.WriteString(`<h3>The bob, at dusk</h3><div class="row">`)
	swimRow(&shapes[0], 3, 0.778, "3 in the water")
	swimRow(&shapes[1], 3, 0.778, "3 in the water")
	b.WriteString(`</div><h3>And at night, where the coat has to earn itself</h3><div class="row">`)
	swimRow(&shapes[0], 4, 0.931, "4 in the water")
	b.WriteString(`</div>`)

	// ---- 3. mixed ------------------------------------------------------------
	b.WriteString(`<h2>3 &middot; Both at once, by the shipped rule</h2>`)
	b.WriteString(`<p class="nt">Which subagents go in is not a per-frame decision &mdash; it is keyed on the subagent's ` +
		`index and the seed, so one cannot climb in and out of the sea between frames. Roughly one in three swims, ` +
		`which is the shipped ratio. Swimmers use vertical room the beach does not have, so the two together hold ` +
		`far more than either alone.</p>`)
	for i := range shapes {
		s := &shapes[i]
		var st strings.Builder
		for f := 0; f < frames; f++ {
			t := 3 + float64(f)/fps
			c, px, py, seaTop, seaBot := scene(96, 26, 0.778, t, seed, 0.6)
			up := pincerWork
			if s.key == "hero" {
				up = heroWork
			}
			drawPosed(c.Near(), s, up, 'o', eyeShine, col, px, py)
			sand, water := split(7, seed)
			drawLitter(c.Near(), s, col, px, py, sand, t, seed)
			drawSwimmers(c.Near(), col, water, c.W, seaTop, seaBot, t, seed)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(11, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>%s</b> <i>7 subagents, split by the shipped rule</i></div>`+
			`<div class="stage win">%s</div></div>`, s.name, st.String())
	}

	// ---- 4. full beach -------------------------------------------------------
	fmt.Fprintf(&b, `<h2>4 &middot; The full beach at %dx%d</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">His window, working, a litter of nine split between the sand and the sea.</p>`)
	for i := range shapes {
		s := &shapes[i]
		var st strings.Builder
		for f := 0; f < 4; f++ {
			t := 3 + float64(f)/3
			c, px, py, seaTop, seaBot := scene(fullW, fullH, 0.778, t, seed, 0.6)
			up := [][]string{pincerWork, pincerWork2}
			if s.key == "hero" {
				up = [][]string{heroWork, heroWork2}
			}
			drawPosed(c.Near(), s, up[f%2], 'o', eyeShine, col, px, py)
			sand, water := split(9, seed)
			drawLitter(c.Near(), s, col, px, py, sand, t, seed)
			drawSwimmers(c.Near(), col, water, c.W, seaTop, seaBot, t, seed)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(7, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>%s</b> <i>nine subagents, dusk</i></div>`+
			`<div class="stage win">%s</div></div>`, s.name, st.String())
	}

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
	_ = companion.HashF
	return canvas.HTMLPage("xscapes — crablets: eyes and water", b.String())
}
