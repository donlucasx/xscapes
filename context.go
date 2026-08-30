package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// contextPage shows the moon carrying context remaining. Rendered as one wide
// scape per level rather than cropped moons, because the question is whether it
// reads in situ -- a moon shown on its own always reads.
func contextPage(seed int64) string {
	cat := companion.NewCat()
	_, chh := cat.Size()

	steps := []struct {
		used  float64
		label string
		note  string
	}{
		{0.00, "fresh session", "full moon &mdash; the whole window ahead of you"},
		{0.35, "35% used", "just off full; nothing to think about yet"},
		{0.60, "60% used", "visibly waning &mdash; the first glanceable warning"},
		{0.85, "85% used", "thin crescent &mdash; compaction is close"},
		{0.97, "97% used", "almost dark, but never absent: a missing moon reads as a bug"},
	}

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;` +
		`overflow:hidden;display:inline-block}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; the moon is the context gauge</h1>`)

	for _, st := range steps {
		c := canvas.New(72, 20, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		act := scape.Activity{Working: true, Level: 0.55, ContextUsed: st.used}
		scape.NewShore(seed, false).Update(c, 3.0, act)
		cat.Draw(c.Near(), 5, c.H-2-chh-3, 3.0, companion.Working)
		writeInSand(c, activity[1:], 17)

		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="rg">ContextUsed %.2f</div><div class="nt">%s</div></div>`+
			`<div><div class="win">%s</div></div></div>`,
			st.label, st.used, st.note, c.HTMLFragment(13))
	}

	// A zoom on just the moons, since a 5-cell disc is small in the wide shot.
	b.WriteString(`<div class="card"><div class="meta"><div class="nm">the moons alone</div>` +
		`<div class="rg">same discs, enlarged</div>` +
		`<div class="nt">Fresh to nearly spent, left to right.</div></div><div style="display:flex;gap:18px">`)
	for _, st := range steps {
		m := canvas.New(11, 7, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < m.H; y++ {
			for x := 0; x < m.W; x++ {
				m.SetBG(x, y, term.RGB{R: 12, G: 14, B: 30})
			}
		}
		sh := scape.NewShore(seed, false)
		sh.MoonOnly(m, 6, 1.0, 1-st.used)
		fmt.Fprintf(&b, `<div><div class="lbl">%.0f%% used</div>%s</div>`, st.used*100, m.HTMLFragment(24))
	}
	b.WriteString(`</div></div>`)

	return canvas.HTMLPage("asciiscapes — context moon", b.String())
}
