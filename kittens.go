package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
)

func kittenPage(seed int64) string {
	cat := companion.NewCat()
	_, chh := cat.Size()

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}` +
		`.lv{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:5px}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; kittens are subagents</h1>`)

	for _, n := range []int{1, 3, 5, 9, 15} {
		var row strings.Builder
		for _, t := range []float64{3.0, 3.6} {
			c := canvas.New(64, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			scape.NewShore(seed, false).Update(c, t, scape.Activity{
				Working: true, Level: math.Min(1, 0.3+float64(n)*0.08), ContextUsed: 0.3,
			})
			top := c.H - 2 - chh
			cat.Draw(c.Near(), 4, top, t, companion.Working)
			cat.DrawKittens(c.Near(), c.Mid(), 4, top, n, t, seed)
			fmt.Fprintf(&row, `<div class="win">%s</div>`, c.HTMLFragment(12))
		}
		label := fmt.Sprintf("%d subagents", n)
		if n == 1 {
			label = "1 subagent"
		}
		note := "front row: full size, faces, own breathing and tail"
		if n > 5 {
			note = "past five they move up the beach: smaller, faceless, on the mid " +
				"layer so alpha makes them recede. Depth, not clutter."
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div></div><div style="display:flex;gap:12px">%s</div></div>`,
			label, note, row.String())
	}
	return canvas.HTMLPage("asciiscapes — kittens", b.String())
}
