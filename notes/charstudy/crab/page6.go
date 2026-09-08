package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// drawPaced draws Hero at home plus a pace offset, swapping the lower half for
// the mid-stride legs while a foot is down.
func drawPaced(near *canvas.Layer, upper []string, eye rune, eyeCol, col term.RGB, homeX, homeY int, dx float64, moving bool) {
	lower := heroLower
	if moving {
		lower = heroLowerStep
	}
	s := shape{lower: lower, eyes: [2]int{4, 7}, eyeRow: 2}
	drawPosed(near, &s, upper, eye, eyeCol, col, homeX+int(dx+0.5), homeY)
}

// sandText writes a few lines of activity tail where live.go writes it: the
// BOTTOM rows, from the left margin to the companion's home column. It is here
// so the pacing can be judged against the thing it must not walk into.
func sandText(c *canvas.Canvas, lines []string, xFrom, xTo int, ink term.RGB) {
	top := c.H - len(lines)
	for i, ln := range lines {
		row := top + i
		if row < 0 || row >= c.H {
			continue
		}
		x := xFrom
		fade := 0.30 + 0.55*(1-float64(i)/float64(len(lines)))
		col := term.Lerp(term.RGB{R: 120, G: 112, B: 104}, ink, fade)
		for _, r := range []rune(companion.NarrowOnly(ln)) {
			if x >= xTo {
				break
			}
			c.Near().Plot(x, row, r, col, 1)
			x++
		}
	}
}

var tailLines = []string{
	"Read  internal/scape/shore.go",
	"Edit  internal/companion/cat.go",
	"Bash  go test ./internal/...",
	"Read  live.go",
}

func page6(seed int64, frames int, fps float64) string {
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
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#b8b8c4;margin:8px 0 18px}
	td,th{border:1px solid #2a2a32;padding:3px 9px;text-align:left}
	th{color:#d8d8e0;font-weight:500}
	.stage{position:relative}.fr{display:none}.fr:first-child{display:block}
	.full{margin:0 0 26px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; Hero paces, and every step is a tool call</h1>`)
	b.WriteString(`<p class="nt">His ask: <b>&ldquo;can we make the main agent move around a bit, specially when its ` +
		`working on its own, or w very few sub agents?&rdquo;</b> then his correction, which is the design: ` +
		`<b>&ldquo;they could be pacing a bit. Should be tied up to an actual agent action so its not random.&rdquo;</b></p>`)
	b.WriteString(`<p class="nt">He is right twice over. A random drift is decoration, and decoration sitting next to a ` +
		`scene where everything else means something is worse than stillness &mdash; it teaches the eye to ignore ` +
		`the companion. Tying it to real events also settles the encoding rule, which a wander could not: the rule ` +
		`forbids encoding in RATE, and one step per event is a <b>count</b>, ending in a <b>position</b>. Both ` +
		`survive a screenshot. Nothing about the pace itself ever varies &mdash; a step is one cell and always takes ` +
		`the same 0.28s. What varies is how many steps there are, because that is how many things the agent did.</p>`)
	b.WriteString(`<p class="nt"><b>And his first ask falls out for free.</b> An agent working alone makes all its tool ` +
		`calls on the main thread, so the crab paces constantly. An agent orchestrating six subagents makes far ` +
		`fewer itself &mdash; the work has moved to the crablets &mdash; so it settles. More active when it is on ` +
		`its own, with no rule anywhere that says so.</p>`)

	// ---- the strip -----------------------------------------------------------
	b.WriteString(`<h2>1 &middot; Where it can go, and why it is only one direction</h2>`)
	b.WriteString(`<pre class="d">  left margin                    sand text ...............  [HOME][pace ->]  edge
                                 |                          |             |
  the activity tail is written on the BOTTOM rows from the   |             |
  margin to the companion's column, and it draws LAST -- so  |             |
  anything drifting left is painted over by it. ------------ +             |
                                                                           |
  to the right there is only the companion's own margin, 2 + w/32 ---------+
  columns. Taking all but one of it costs nothing and needs no layout change.</pre>`)
	fmt.Fprintf(&b, `<p class="nt">That gives <b>%d cells at his 126-wide window</b> and %d at 80. The path is a `+
		`triangle &mdash; out to the far end and back &mdash; so it paces rather than walks off, and over a long `+
		`session it averages out at home instead of drifting to one side.</p>`, paceSpan(126), paceSpan(80))

	// ---- the three cases -----------------------------------------------------
	b.WriteString(`<h2>2 &middot; Three sessions, same rules</h2>`)
	cases := []struct {
		name, note string
		perMin     float64
		kittens    int
	}{
		{"working alone", "Every tool call is the main thread's. It paces the whole strip, turns, comes back. This is the case he asked about.", 42, 0},
		{"two subagents", "Some of the work has moved off the main thread. It still paces, with longer stands between steps.", 18, 2},
		{"six subagents", "The main thread is mostly orchestrating. It barely moves, and the beach is full of crablets doing the work instead.", 6, 6},
		{"idle, waiting on him", "No events, no steps. A still crab means nothing is happening, which is worth being able to read at a glance.", 0, 0},
	}
	for _, cs := range cases {
		ev := toolStream(float64(frames*8)/fps, cs.perMin, seed)
		var st strings.Builder
		for f := 0; f < frames*8; f++ {
			t := float64(f) / fps
			c, px, py, seaTop, seaBot := scene(88, 24, 0.778, 3+t, seed, 0.55)
			span := paceSpan(c.W)
			n, since := stepsBy(t, ev)
			dx, moving := paceAt(n, since, span)
			sand, water := split(cs.kittens, seed)
			drawLitter(c.Near(), &shapes[1], col, px-span, py, sand, 3+t, seed)
			drawSwimmers(c.Near(), col, water, c.W, seaTop, seaBot, 3+t, seed)
			up := heroWork
			if moving {
				up = heroWork2
			}
			drawPaced(c.Near(), up, 'o', eyeShine, col, px, py, dx, moving)
			sandText(c, tailLines, 2, px-1-span, term.RGB{R: 244, G: 236, B: 220})
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(11, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="row"><div class="col"><div class="lbl"><b>%s</b> <i>%.0f main-thread tool calls/min</i></div>`+
			`<div class="stage">%s</div></div><p class="note">%s</p></div>`, cs.name, cs.perMin, st.String(), cs.note)
	}

	// ---- states hold still ---------------------------------------------------
	b.WriteString(`<h2>3 &middot; It stops when it has something to say</h2>`)
	b.WriteString(`<p class="nt">Three states hold their ground however many events arrive. A raised claw that is also ` +
		`pacing is noise; <b>done</b> is explicitly held still so it cannot be misread as another ask; and ` +
		`<b>worried</b> persists until it clears, so it should look planted. Only resting and working pace.</p><div class="row">`)
	for _, p := range posesFor("hero") {
		var ps strings.Builder
		ev := toolStream(float64(frames*2)/fps, 42, seed)
		for f := 0; f < frames*2; f++ {
			t := float64(f) / fps
			c, px, py, _, _ := scene(72, 22, 0.778, 3+t, seed, 0.6)
			dx, moving := 0.0, false
			if !stillFor(p.name) {
				n, since := stepsBy(t, ev)
				dx, moving = paceAt(n, since, paceSpan(c.W))
			}
			eye := p.eye
			if p.blink && f%frames == frames-2 {
				eye = '-'
			}
			drawPaced(c.Near(), p.uppers[f%len(p.uppers)], eye, p.eyeCol, col, px, py, dx, moving)
			if p.say != "" {
				rows, bc := companion.DoneBubble(p.say), bubbleCol
				if p.ask {
					rows, bc = companion.Bubble(p.say), bubbleAskCol
				}
				(&companion.Sprite{Rows: rows, Body: bc, Opaque: true}).Draw(c.Near(), px-2, py-len(rows))
			}
			fmt.Fprintf(&ps, `<div class="fr">%s</div>`, c.HTMLFragmentAs(12, term.Profile256))
		}
		still := "paces"
		if stillFor(p.name) {
			still = "holds still"
		}
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%s</i></div><div class="stage">%s</div></div>`,
			p.name, still, ps.String())
	}
	b.WriteString(`</div>`)

	// ---- full beach ----------------------------------------------------------
	fmt.Fprintf(&b, `<h2>4 &middot; The full beach at %dx%d, working alone</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">His window, no subagents, the main thread doing all of it. The tail is written under ` +
		`the pacing strip and never touched by it.</p>`)
	ev := toolStream(float64(8*6)/fps, 42, seed)
	var fs strings.Builder
	for f := 0; f < 24; f++ {
		t := float64(f) / fps
		c, px, py, _, _ := scene(fullW, fullH, 0.778, 3+t, seed, 0.6)
		span := paceSpan(c.W)
		n, since := stepsBy(t, ev)
		dx, moving := paceAt(n, since, span)
		up := heroWork
		if moving {
			up = heroWork2
		}
		drawPaced(c.Near(), up, 'o', eyeShine, col, px, py, dx, moving)
		sandText(c, tailLines, 2, px-1-span, term.RGB{R: 244, G: 236, B: 220})
		fmt.Fprintf(&fs, `<div class="fr">%s</div>`, c.HTMLFragmentAs(7, term.Profile256))
	}
	fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>Hero</b> <i>dusk, alone, pacing</i></div>`+
		`<div class="stage win">%s</div></div>`, fs.String())

	// ---- what it needs -------------------------------------------------------
	b.WriteString(`<h2>5 &middot; What the product needs for this</h2>`)
	b.WriteString(`<p class="nt">Almost nothing, and that is because of work the other session shipped today. ` +
		`<code>Event.Agent</code> is empty for the main thread and set for a subagent &mdash; the field the worry ` +
		`bar now uses to raise Worried on main-thread errors only. The same field separates main-thread tool calls ` +
		`from subagent ones here.</p>`)
	b.WriteString(`<table><tr><th>needed</th><th>size</th></tr>` +
		`<tr><td><code>reduce.State.Steps int</code> &mdash; count tool events with <code>Agent == ""</code></td><td>about three lines, beside the existing <code>Events</code> counter</td></tr>` +
		`<tr><td>pass the offset to the companion draw</td><td>one argument, or read it off State in drawScene</td></tr>` +
		`<tr><td>a mid-stride lower half</td><td>done, one bitmap</td></tr>` +
		`</table>`)
	b.WriteString(`<p class="nt"><b>The sand bounds do not move.</b> <code>SandTo</code> stays computed from the ` +
		`companion's HOME column, so the activity tail never reflows &mdash; which it would do every frame if the ` +
		`offset were folded into <code>CatX</code>, since the tail drops whole pieces as its width changes.</p>`)

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
	return canvas.HTMLPage("xscapes — Hero paces", b.String())
}
