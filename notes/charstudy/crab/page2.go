package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// round2 is the answer to "they look like a lobster": four new silhouettes
// built around an open pincer, the two he liked kept alongside for the
// before-and-after, and the full beach at his own window size.
var round2 = []crab{
	{"pincer", "Pincer", "The straight answer. Claws up at the top corners with the pincer open, a wide flat shell under them. The gap between the tips is a whole empty cell, which is the smallest opening that survives the halving.",
		pincerBody, pincerYoung, [2]int{4, 7}, 1},
	{"hero", "Hero", "The claws raised high on visible arms, clear of the shell. The most unmistakably crab of the set, and the one with the most to say later -- a raised claw is a gesture the cat does not have.",
		heroBody, pincerYoung, [2]int{4, 7}, 2},
	{"wide", "Wide", "The flattest shell of the set, claws forward and low at the front corners. Reads as something that scuttles sideways rather than stands up.",
		wideBody, crabYoung, [2]int{4, 7}, 0},
	{"stalker2", "Stalker II", "Your stalker, unchanged above the shell -- same tall eyestalks -- with the claws moved onto arms and given an open pincer. Same eyes, different arms.",
		stalker2Body, crabYoung, [2]int{4, 7}, 0},
}

var round1keep = []crab{
	{"broad", "Broad", "Round one. The claws sit in line with the shell at its sides, which is a lobster's profile exactly.", broadBody, crabYoung, [2]int{4, 7}, 1},
	{"stalker", "Stalker", "Round one. Same problem, and the long body under the stalks makes it worse.", stalkerBody, crabYoung, [2]int{3, 7}, 0},
}

// His own window, measured from the session 20 screenshots.
const fullW, fullH = 126, 74

func page2(seed int64) string {
	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	.note{font:12px/1.5 ui-sans-serif,system-ui;color:#8a8a99;max-width:34ch;margin:2px 0 0}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 6px;padding-top:14px;border-top:1px solid #23232a}
	pre.d{font:11px/1.35 ui-monospace,monospace;color:#8a8a99;background:#16161c;border:1px solid #23232a;padding:10px 12px;overflow-x:auto}
	.full{margin:0 0 26px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the crab, round two</h1>`)
	b.WriteString(`<p class="nt">His note on round one: <b>&ldquo;i like broad and stalker, but they are looking more like a ` +
		`lobster than a crab &hellip; I think we need to see the claws so it looks more like a crab.&rdquo;</b></p>`)
	b.WriteString(`<p class="nt">He is right, and the cause is placement rather than size. In round one the claws sat ` +
		`<b>in line with the shell at its sides</b> &mdash; which is a lobster's profile exactly: a long body with ` +
		`appendages down it. Two things make a crab read as a crab at this size:</p>`)
	b.WriteString(`<pre class="d">1. an OPEN PINCER -- a gap between two tips. At 12x7 the gap has to be a
   whole empty cell, because a sub-cell is 1px wide by 2px tall and a
   notch thinner than that closes when the bitmap is halved. Two of
   these four lost their pincer to exactly that before it was caught.

2. claws held UP and FORWARD of the shell, not beside it. This is the
   whole lobster/crab difference and it costs no extra cells -- the
   claws move into rows the cat leaves empty anyway.</pre>`)
	b.WriteString(`<p class="nt">Same 24x28 grid, same 12 cells by 7, same aspect and footprint as every other ` +
		`companion &mdash; nothing here is bigger than the cat. All salmon (index 210), which is the coat that ` +
		`survived the glyph boost unchanged.</p>`)

	cell := func(label, sub, body, note string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%s</i></div><div class="win">%s</div>`, label, sub, body)
		if note != "" {
			fmt.Fprintf(&b, `<p class="note">%s</p>`, note)
		}
		b.WriteString(`</div>`)
	}
	crop := func(cr *crab, tod float64) string {
		c, px, py := crabFrame(cr, coats[0].col, 80, 24, tod, 3, seed, false)
		return c.HTMLFragmentCropAs(px-3, py-1, px+12+3, py+7+1, 33, term.Profile256)
	}

	b.WriteString(`<h2>1 &middot; The four new shapes, at dusk</h2><div class="row">`)
	for i := range round2 {
		cr := &round2[i]
		cell(cr.name, "", crop(cr, 0.778), cr.note)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>2 &middot; Against round one &mdash; what changed</h2>`)
	b.WriteString(`<p class="nt">The two you liked, beside their replacements. Same coat, same hour, same crop.</p><div class="row">`)
	for i := range round1keep {
		cr := &round1keep[i]
		cell(cr.name, "round 1", crop(cr, 0.778), cr.note)
	}
	cell(round2[0].name, "round 2", crop(&round2[0], 0.778), "Broad's shell, with the claws lifted off the sides and opened.")
	cell(round2[3].name, "round 2", crop(&round2[3], 0.778), "Stalker's eyes kept exactly, with the claws lifted off the sides and opened.")
	b.WriteString(`</div>`)

	b.WriteString(`<h2>3 &middot; Through the day</h2>`)
	for _, hr := range hours {
		fmt.Fprintf(&b, `<div class="lbl" style="margin:14px 0 4px"><b>%s</b></div><div class="row">`, hr.name)
		for i := range round2 {
			cell(round2[i].name, "", crop(&round2[i], hr.tod), "")
		}
		b.WriteString(`</div>`)
	}

	// ---- the full beach ------------------------------------------------------
	fmt.Fprintf(&b, `<h2>4 &middot; The full beach, at your window &mdash; %dx%d</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">His ask, and the honest answer to it: at the real size the companion is a small ` +
		`thing in the bottom corner of a wide scene, and every crop above flatters it. The sea is working, ` +
		`the checklist is two of five, the litter is out. This is the size he actually meets it at.</p>`)
	for i := range round2 {
		cr := &round2[i]
		c, _, _ := crabFrame(cr, coats[0].col, fullW, fullH, 0.778, 3, seed, true)
		fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>%s</b> <i>dusk</i></div><div class="win">%s</div></div>`,
			cr.name, c.HTMLFragmentAs(7, term.Profile256))
	}
	b.WriteString(`<p class="nt">And the dark case &mdash; the lower beach falls away to black by design, so night is ` +
		`where a coat has to earn itself.</p>`)
	for _, i := range []int{0, 3} {
		cr := &round2[i]
		c, _, _ := crabFrame(cr, coats[0].col, fullW, fullH, 0.931, 3, seed, true)
		fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>%s</b> <i>night</i></div><div class="win">%s</div></div>`,
			cr.name, c.HTMLFragmentAs(7, term.Profile256))
	}

	return canvas.HTMLPage("xscapes — the crab, round two", b.String())
}
