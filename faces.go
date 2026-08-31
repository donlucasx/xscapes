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
		`below is therefore a character overlay, plotted after the body at cell precision.</p>`)

	faces := []struct {
		key  string
		note string
	}{
		{"plain", "what ships today"},
		{"nose", "one rose cell below the eyes"},
		{"whiskers", "+ whiskers, in the empty cells beside the head"},
		{"full", "+ inner ears and cheek tufts"},
	}

	b.WriteString(`<h2>Face detail &mdash; cream, truecolor</h2><div class="row">`)
	for _, f := range faces {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; %s</div><div class="win">%s</div></div>`,
			f.key, f.note, portrait(companion.Faces[f.key], companion.Coats["cream"], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	coats := []string{"cream", "terracotta", "ginger", "slate", "charcoal"}

	b.WriteString(`<h2>Coats &mdash; full face, truecolor</h2>`)
	b.WriteString(`<p class="nt">Terracotta is Claude's own colour, which is the reference. ` +
		`The eyes stay bright in every coat on purpose: they carry the companion's state, and ` +
		`state is the one thing that must read before anything else does.</p><div class="row">`)
	for _, k := range coats {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s</div><div class="win">%s</div></div>`,
			k, portrait(companion.Faces["full"], companion.Coats[k], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same coats as Terminal.app will paint them</h2>`)
	b.WriteString(`<p class="nt">256 colours, glyph chroma 2.6&times;. A solid coat has far more ` +
		`chroma than cream does, so it survives quantisation where cream lands in the greys &mdash; ` +
		`which is an argument for a coloured coat that has nothing to do with taste.</p><div class="row">`)
	for _, k := range coats {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; 256</div><div class="win">%s</div></div>`,
			k, portrait(companion.Faces["full"], companion.Coats[k], companion.Working, term.Profile256, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state, in the two leading coats</h2>`)
	b.WriteString(`<p class="nt">The whiskers droop when something is broken &mdash; expression ` +
		`from geometry that is already on screen.</p>`)
	for _, coat := range []string{"cream", "terracotta"} {
		fmt.Fprintf(&b, `<div class="row">`)
		for _, st := range []struct {
			s companion.State
			n string
		}{{companion.Resting, "resting"}, {companion.Working, "working"},
			{companion.NeedsYou, "needs you"}, {companion.Worried, "something is broken"}} {
			fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s &middot; %s</div><div class="win">%s</div></div>`,
				coat, st.n, portrait(companion.Faces["full"], companion.Coats[coat], st.s, term.ProfileTrueColor, 3.1))
		}
		b.WriteString(`</div>`)
	}

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
