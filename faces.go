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
	.row{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start;margin-bottom:6px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:30px 0 10px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; how much cat fits in nine cells</h1>`)
	b.WriteString(`<p class="nt">The head is four character rows: ears, forehead, eyes, chin. ` +
		`The eyes already own their row, and the bitmap ORs its rows in pairs before quadranting, ` +
		`so detail drawn INTO the bitmap on one row is erased by the row above it. Everything ` +
		`below is a character overlay, plotted after the body at cell precision, in a colour ` +
		`derived from the coat so a new coat needs no new colours.</p>`)

	notes := map[string]string{
		"plain":   "what ships today",
		"nose":    "+ a rose nose below the eyes",
		"classic": "+ whiskers and inner ears",
		"tabby":   "+ the tabby M and ear tufts",
		"tuxedo":  "+ pale muzzle, chest bib, toe tips",
		"full":    "everything at once",
	}

	for _, coat := range companion.CoatOrder {
		fmt.Fprintf(&b, `<h2>%s</h2><div class="row">`, coat)
		for _, k := range companion.FaceOrder {
			fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; %s</div><div class="win">%s</div></div>`,
				k, notes[k],
				portrait(companion.Faces[k], companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`<h2>Terminal.app, 256 colours &mdash; the full face</h2>`)
	b.WriteString(`<p class="nt">Markings derived from the coat move with it through quantisation, ` +
		`so a coat that survives keeps its face.</p><div class="row">`)
	for _, coat := range companion.CoatOrder {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; 256</div><div class="win">%s</div></div>`,
			coat, portrait(companion.Faces["full"], companion.Coats[coat], companion.Working, term.Profile256, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state, in the three coats Lucas is leaning toward</h2>`)
	b.WriteString(`<p class="nt">The whiskers droop and the eyes go amber when something is ` +
		`broken; they splay and the eyes widen when the agent needs you.</p>`)
	for _, coat := range []string{"cream", "slate", "charcoal"} {
		b.WriteString(`<div class="row">`)
		for _, st := range []struct {
			s companion.State
			n string
		}{{companion.Resting, "resting"}, {companion.Working, "working"},
			{companion.NeedsYou, "needs you"}, {companion.Worried, "something is broken"}} {
			fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; %s</div><div class="win">%s</div></div>`,
				coat, st.n,
				portrait(companion.Faces["full"], companion.Coats[coat], st.s, term.ProfileTrueColor, 3.1))
		}
		b.WriteString(`</div>`)
	}

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
