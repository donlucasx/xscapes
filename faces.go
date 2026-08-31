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
	b.WriteString(`<h1>asciiscapes &mdash; whiskers and ears, other ways round</h1>`)
	b.WriteString(`<p class="nt">Every cat below has the nose and the toe tips, which are settled. ` +
		`What changes is the whiskers and the inside of the ear &mdash; and these are different ` +
		`ideas rather than different amounts, because at nine cells wide the question is not how ` +
		`many strokes to draw but whether a whisker is a line, a pair of tips, or a shadow in the fur.</p>`)

	cell := func(label, note, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b>%s</div><div class="win">%s</div></div>`,
			label, note, body)
	}
	shot := func(f companion.Face, coat string, st companion.State, p term.Profile) string {
		return portrait(f, companion.Coats[coat], st, p, 3.1)
	}

	// Whiskers alone, on a dark coat where a pale mark reads, then on cream.
	for _, coat := range []string{"charcoal", "cream"} {
		fmt.Fprintf(&b, `<h2>Whiskers &mdash; %s</h2><div class="row">`, coat)
		for _, w := range companion.WhiskerStyles {
			f := companion.Base
			f.Whiskers = w.Style
			cell(w.Name, " &middot; "+w.Note, shot(f, coat, companion.Working, term.ProfileTrueColor))
		}
		b.WriteString(`</div>`)
	}

	for _, coat := range []string{"charcoal", "cream"} {
		fmt.Fprintf(&b, `<h2>Inner ears &mdash; %s</h2><div class="row">`, coat)
		for _, e := range companion.EarStyles {
			f := companion.Base
			f.Ears = e.Style
			cell(e.Name, " &middot; "+e.Note, shot(f, coat, companion.Working, term.ProfileTrueColor))
		}
		b.WriteString(`</div>`)
	}

	// The two together, so the pairing can be judged rather than each half.
	b.WriteString(`<h2>Paired &mdash; slate</h2>`)
	b.WriteString(`<p class="nt">Each whisker style with the ear treatment that suits it.</p><div class="row">`)
	for _, pr := range []struct {
		w companion.WhiskerStyle
		e companion.EarStyle
		n string
	}{
		{companion.NoWhiskers, companion.EarDot, "none + dot"},
		{companion.WhiskerDots, companion.EarDot, "tips + dot"},
		{companion.WhiskerTicks, companion.EarRim, "ticks + rim"},
		{companion.WhiskerCarved, companion.EarDark, "carved + shadow"},
		{companion.WhiskerFan, companion.EarTuft, "fan + tuft"},
		{companion.WhiskerStrokes, companion.EarRose, "strokes + rose"},
		{companion.WhiskerLow, companion.EarRose, "low + rose"},
	} {
		f := companion.Base
		f.Whiskers, f.Ears = pr.w, pr.e
		cell(pr.n, "", shot(f, "slate", companion.Working, term.ProfileTrueColor))
	}
	b.WriteString(`</div>`)

	// One promising pairing across every coat that survived.
	b.WriteString(`<h2>"tips + dot" across your six coats</h2>`)
	b.WriteString(`<p class="nt">The quietest pairing: whisker ends only, and a single cell in ` +
		`the ear.</p><div class="row">`)
	quiet := companion.Base
	quiet.Whiskers, quiet.Ears = companion.WhiskerDots, companion.EarDot
	for _, coat := range companion.CoatOrder {
		cell(coat, "", shot(quiet, coat, companion.Working, term.ProfileTrueColor))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>The same, as Terminal.app paints it</h2><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, " &middot; 256", shot(quiet, coat, companion.Working, term.Profile256))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Base only &mdash; nose and toes, no whiskers or ears</h2>`)
	b.WriteString(`<p class="nt">For comparison: the cat you already like, in all six.</p><div class="row">`)
	for _, coat := range companion.CoatOrder {
		cell(coat, "", shot(companion.Base, coat, companion.Working, term.ProfileTrueColor))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Every state &mdash; slate, tips + dot</h2><div class="row">`)
	for _, st := range []struct {
		s companion.State
		n string
	}{{companion.Resting, "resting"}, {companion.Working, "working"},
		{companion.NeedsYou, "needs you"}, {companion.Worried, "broken"}} {
		cell(st.n, "", shot(quiet, "slate", st.s, term.ProfileTrueColor))
	}
	b.WriteString(`</div>`)

	return canvas.HTMLPage("asciiscapes — companion study", b.String())
}
