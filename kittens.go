package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

func kittenPage(seed int64) string {
	cat := companion.NewCat()
	_, chh := cat.Size()

	counts := []struct {
		n    int
		note string
	}{
		{2, "roughly one in three takes to the water"},
		{5, "beach tier is set by how many are ON the beach, not the total"},
		{8, "swimmers ride the swell and waddle sideways in their own lanes"},
		{12, "the sea absorbs what the sand cannot hold"},
		{18, "a proper fan-out"},
	}

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}</style>`)
	b.WriteString(`<h1>xscapes &mdash; the kitten ladder</h1>`)

	for _, cc := range counts {
		var row strings.Builder
		fit := 0
		for _, t := range []float64{3.0, 4.1} {
			c := canvas.New(76, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			scape.NewShore(seed, false).Update(c, t, scape.Activity{
				Working: true, Level: math.Min(1, 0.3+float64(cc.n)*0.05), ContextUsed: 0.3,
			})
			top := c.H - 2 - chh
			kc := companion.NewCat()
			kc.Draw(c.Near(), 3, top, t, companion.Working)
			fit = kc.DrawKittens(c.Near(), c.Mid(), 3, top, cc.n, c.W-1, int(float64(c.H)*0.42)+1, t, seed)
			fmt.Fprintf(&row, `<div class="win">%s</div>`, c.HTMLFragment(12))
		}
		short := ""
		if fit < cc.n {
			short = fmt.Sprintf(" &mdash; <b>%d of %d fit</b>", fit, cc.n)
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%d subagents</div>`+
			`<div class="rg">%d drawn</div><div class="nt">%s%s</div></div>`+
			`<div style="display:flex;gap:12px">%s</div></div>`,
			cc.n, fit, cc.note, short, row.String())
	}
	return canvas.HTMLPage("xscapes — kitten ladder", b.String())
}
