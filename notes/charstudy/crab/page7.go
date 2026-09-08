package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// page7 renders the SHIPPED companion package, not the study's own bitmaps.
// Everything above this page was a mockup; this is the binary's own output.
func page7(seed int64, frames int, fps float64) string {
	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 6px;padding-top:14px;border-top:1px solid #23232a}
	.stage{position:relative}.fr{display:none}.fr:first-child{display:block}
	.full{margin:0 0 26px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; Hero, shipped</h1>`)
	b.WriteString(`<p class="nt">Everything above this page was a mockup drawn from the study's own bitmaps. ` +
		`This one calls <code>companion.New("crab")</code> and draws through the shipped package &mdash; it is what ` +
		`<code>xscapes claude</code> now puts on the beach. Salmon is locked at index 210 and the crab is the ` +
		`default; <code>xscapes companion cat</code> puts the cat back.</p>`)

	draw := func(w, h int, name string, st companion.State, n int, exits []float64, t float64) *canvas.Canvas {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: st == companion.Working, Level: 0.55,
			TimeOfDay: 0.778, ContextUsed: 0.3, TodoDone: 2, TodoTotal: 5}
		for k := 0; k < 10; k++ {
			sh.Update(c, t+float64(k)/40, act)
		}
		comp := companion.New(name)
		comp.FaceLeft(true)
		cw, ch := comp.Size()
		px, py := c.W-cw-(2+c.W/32), c.H-2-ch
		comp.Draw(c.Near(), px, py, t, st)
		seaTop, seaBot := int(float64(c.H)*0.42)+1, sh.SandTop()-2
		if n > 0 {
			comp.DrawKittens(c.Near(), c.Mid(), px, py, n, c.W-1, seaTop, seaBot, t, seed)
		}
		if len(exits) > 0 {
			comp.DrawKittenExits(c.Near(), exits, px, py, c.W-1, seaTop, seaBot, t, seed)
		}
		return c
	}

	b.WriteString(`<h2>1 &middot; The five states, from the package</h2><div class="row">`)
	for _, s := range []struct {
		st companion.State
		n  string
	}{{companion.Resting, "resting"}, {companion.Working, "working"}, {companion.NeedsYou, "needs you"},
		{companion.Done, "done"}, {companion.Worried, "worried"}} {
		var st strings.Builder
		for f := 0; f < frames; f++ {
			t := 3 + float64(f)/fps
			c := draw(64, 20, "crab", s.st, 0, nil, t)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(12, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div><div class="stage">%s</div></div>`, s.n, st.String())
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>2 &middot; Beside the cat, same call, same frame</h2><div class="row">`)
	for _, nm := range []string{"crab", "cat"} {
		c := draw(72, 22, nm, companion.Working, 4, nil, 3)
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>4 subagents</i></div><div class="win">%s</div></div>`,
			nm, c.HTMLFragmentAs(12, term.Profile256))
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b, `<h2>3 &middot; The full beach at %dx%d</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">Nine subagents split between sand and sea by the shipped rule, two leaving.</p>`)
	var fs strings.Builder
	for f := 0; f < 6; f++ {
		t := 3 + float64(f)/3
		e := float64(f) / 6
		c := draw(fullW, fullH, "crab", companion.Working, 9, []float64{e, e*0.6 + 0.2}, t)
		fmt.Fprintf(&fs, `<div class="fr">%s</div>`, c.HTMLFragmentAs(7, term.Profile256))
	}
	fmt.Fprintf(&b, `<div class="full"><div class="stage win">%s</div></div>`, fs.String())

	fmt.Fprintf(&b, `<script>
document.querySelectorAll('.stage').forEach(function(st){
  var fr = Array.prototype.slice.call(st.children), i = 0;
  setInterval(function(){ fr[i].style.display='none'; i=(i+1)%%fr.length; fr[i].style.display='block'; }, %d);
});
</script>`, int(1000/fps))
	return canvas.HTMLPage("xscapes — Hero, shipped", b.String())
}
