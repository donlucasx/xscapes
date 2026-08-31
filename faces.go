package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// facePage is the companion study: how much face a nine-cell head can carry,
// and what the body looks like in something other than cream.
func facePage(seed int64) string {
	// A patch of the real scene behind each portrait, because a companion is
	// only ever seen against sand and never against a swatch.
	portrait := func(f companion.Face, coat term.RGB, st companion.State, prof term.Profile, t float64) string {
		c := canvas.New(16, 10, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		for i := 0; i < 12; i++ {
			sh.Update(c, 2+float64(i)/20, scape.Activity{Working: true, Level: 0.5})
		}
		cat := companion.NewCat()
		cat.SetFace(f)
		cat.SetCoat(coat)
		_, chh := cat.Size()
		cat.Draw(c.Near(), 2, c.H-1-chh, t, st)
		return c.HTMLFragmentAs(15, prof)
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:11px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:32px 0 4px}
	h2+p{margin-top:4px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; the companion, one mark at a time</h1>`)
	b.WriteString(`<p class="nt">Every feature is a character overlay drawn after the body, in a ` +
		`colour derived from the coat. The head is four rows and the eyes own one of them, so ` +
		`there is room for very little &mdash; which is the argument for adding almost nothing.</p>`)

	cell := func(label, note, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b>%s</div><div class="win">%s</div></div>`,
			label, note, body)
	}

	// Each feature alone, so it can be judged without anything else competing.
	b.WriteString(`<h2>One feature at a time &mdash; cream</h2>`)
	b.WriteString(`<p class="nt">The plain cat first, then each mark on its own.</p><div class="row">`)
	for _, s := range companion.Singles {
		cell(s.Key, " &middot; "+s.Note,
			portrait(s.Face, companion.Coats["cream"], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same, on charcoal &mdash; where pale marks actually show</h2>`)
	b.WriteString(`<div class="row">`)
	for _, s := range companion.Singles {
		cell(s.Key, "",
			portrait(s.Face, companion.Coats["charcoal"], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	notes := map[string]string{
		"plain":   "ships today",
		"nose":    "last session's pick",
		"hint":    "nose + chin",
		"soft":    "+ muzzle",
		"cat":     "nose + whiskers + muzzle",
		"mitten":  "muzzle, bib, toes, tail tip",
		"classic": "most of it",
		"full":    "everything",
	}

	b.WriteString(`<h2>Combinations, quietest first</h2><div class="row">`)
	for _, k := range companion.FaceOrder {
		cell(k, " &middot; "+notes[k],
			portrait(companion.Faces[k], companion.Coats["cream"], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Coats &mdash; the three you liked, and six more</h2>`)
	b.WriteString(`<p class="nt">Shown on "cat": nose, whiskers, muzzle.</p><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, "", portrait(companion.Faces["cat"], companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same coats, plain &mdash; so the colour is the only variable</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, "", portrait(companion.Faces["plain"], companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>As Terminal.app paints them &mdash; 256 colours</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, " &middot; 256", portrait(companion.Faces["cat"], companion.Coats[coat], companion.Working, term.Profile256, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state</h2>`)
	b.WriteString(`<p class="nt">Whiskers droop when something is broken and splay when the agent ` +
		`needs you; the eyes carry the state itself.</p>`)
	for _, coat := range []string{"cream", "slate", "charcoal"} {
		b.WriteString(`<div class="row">`)
		for _, st := range []struct {
			s companion.State
			n string
		}{{companion.Resting, "resting"}, {companion.Working, "working"},
			{companion.NeedsYou, "needs you"}, {companion.Worried, "broken"}} {
			cell(coat+" · "+st.n, "",
				portrait(companion.Faces["cat"], companion.Coats[coat], st.s, term.ProfileTrueColor, 3.1))
		}
		b.WriteString(`</div>`)
	}

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
