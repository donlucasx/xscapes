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
	b.WriteString(`<h1>asciiscapes &mdash; whiskers on the face, ears inside the ears</h1>`)
	b.WriteString(`<p class="nt">Both were placement bugs. Measured off a rendered frame: the ears ` +
		`occupy cells 1-2 and 6-7, and the chin row runs 0-8 with cells 0 and 8 only HALF filled. ` +
		`The old whiskers sat at cell -1 and 9 &mdash; adjacent to the head's bounding box but two ` +
		`cells from solid fur, which is why they floated. The old ear mark sat on the left half of ` +
		`BOTH ears, so it was not even symmetrical.</p>`)

	cell := func(label, note, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b>%s</div><div class="win">%s</div></div>`,
			label, note, body)
	}

	// Whiskers, with and without the toes, on the coats that stand out most.
	for _, coat := range []string{"slate", "cream", "charcoal"} {
		fmt.Fprintf(&b, `<h2>Whiskers &mdash; %s</h2>`, coat)
		for _, toes := range []bool{true, false} {
			label := "with toes"
			if !toes {
				label = "without toes"
			}
			fmt.Fprintf(&b, `<h3>%s</h3><div class="row">`, label)
			for _, w := range companion.WhiskerStyles {
				f := companion.Face{Nose: true, Toes: toes, Whiskers: w.Style}
				cell(w.Name, " &middot; "+w.Note,
					portrait(f, companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
			}
			b.WriteString(`</div>`)
		}
	}

	for _, coat := range []string{"slate", "cream", "charcoal"} {
		fmt.Fprintf(&b, `<h2>Inner ears &mdash; %s</h2>`, coat)
		for _, toes := range []bool{true, false} {
			label := "with toes"
			if !toes {
				label = "without toes"
			}
			fmt.Fprintf(&b, `<h3>%s</h3><div class="row">`, label)
			for _, e := range companion.EarStyles {
				f := companion.Face{Nose: true, Toes: toes, Ears: e.Style}
				cell(e.Name, " &middot; "+e.Note,
					portrait(f, companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
			}
			b.WriteString(`</div>`)
		}
	}

	// The two together.
	b.WriteString(`<h2>Together &mdash; flush whiskers, inner ear</h2><div class="row">`)
	both := companion.Face{Nose: true, Toes: true,
		Whiskers: companion.WhiskerFlush, Ears: companion.EarInner}
	for _, coat := range companion.CoatOrder {
		cell(coat, "", portrait(both, companion.Coats[coat], companion.Working, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same, as Terminal.app paints it</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, " &middot; 256", portrait(both, companion.Coats[coat], companion.Working, term.Profile256, 3.1))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state &mdash; slate</h2><div class="row">`)
	for _, st := range []struct {
		s companion.State
		n string
	}{{companion.Resting, "resting"}, {companion.Working, "working"},
		{companion.NeedsYou, "needs you"}, {companion.Worried, "broken"}} {
		cell(st.n, "", portrait(both, companion.Coats["slate"], st.s, term.ProfileTrueColor, 3.1))
	}
	b.WriteString(`</div>`)

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
