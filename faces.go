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
	.row{display:flex;gap:11px;flex-wrap:wrap;align-items:flex-start;margin-bottom:10px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:30px 0 4px}
	h3{font:500 11px ui-monospace,monospace;color:#8a8a99;letter-spacing:.08em;
	   text-transform:uppercase;margin:14px 0 6px}
	h2+p{margin-top:4px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; whiskers on the muzzle</h1>`)
	b.WriteString(`<p class="nt">Ears are settled: inner shadow, the coat's own dark tone inside ` +
		`each ear, so the cat stays monochrome. What is left is where the four whiskers sit. ` +
		`All of these keep them on the muzzle rows &mdash; the chin row the nose sits on, and the ` +
		`one below it &mdash; and vary how far each reaches. The old eye-row version is included ` +
		`last so the difference is visible rather than asserted.</p>`)

	cell := func(label, note, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b>%s</div><div class="win">%s</div></div>`,
			label, note, body)
	}
	face := func(w companion.WhiskerStyle, toes bool) companion.Face {
		return companion.Face{Nose: true, Toes: toes, Whiskers: w, Ears: companion.EarInnerDark}
	}

	for _, coat := range []string{"slate", "cream", "charcoal"} {
		fmt.Fprintf(&b, `<h2>%s</h2>`, coat)
		for _, toes := range []bool{true, false} {
			label := "with toes"
			if !toes {
				label = "without toes"
			}
			fmt.Fprintf(&b, `<h3>%s</h3><div class="row">`, label)
			for _, w := range companion.WhiskerStyles {
				cell(w.Name, " &middot; "+w.Note,
					portrait(face(w.Style, toes), companion.Coats[coat], companion.Working,
						term.ProfileTrueColor, 3.1))
			}
			b.WriteString(`</div>`)
		}
	}

	b.WriteString(`<h2>"muzzle" across all five coats</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, "", portrait(face(companion.WhiskerMuzzle, true), companion.Coats[coat],
			companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same, as Terminal.app paints it</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, " &middot; 256", portrait(face(companion.WhiskerMuzzle, true), companion.Coats[coat],
			companion.Working, term.Profile256, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state &mdash; slate, "muzzle"</h2>`)
	b.WriteString(`<p class="nt">The whiskers drop a row when something is broken.</p><div class="row">`)
	for _, st := range []struct {
		s companion.State
		n string
	}{{companion.Resting, "resting"}, {companion.Working, "working"},
		{companion.NeedsYou, "needs you"}, {companion.Worried, "broken"}} {
		cell(st.n, "", portrait(face(companion.WhiskerMuzzle, true), companion.Coats["slate"],
			st.s, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
