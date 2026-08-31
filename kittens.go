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

	rows := []struct {
		scale companion.KittenScale
		name  string
		want  int
		note  string
	}{
		{companion.KitLarge, "large &middot; 6&times;4 cells", 8, "two poses and a tail. Reads best, fits fewest."},
		{companion.KitSmall, "small &middot; 5&times;3 cells", 8, "ears, sockets and legs kept; pose and tail dropped."},
		{companion.KitTiny, "tiny &middot; 4&times;3 cells", 8, "the smallest thing that still has a face."},
		{companion.KitSmall, "small &middot; 5&times;3 cells", 14, "how many actually fit."},
		{companion.KitTiny, "tiny &middot; 4&times;3 cells", 14, "how many actually fit."},
		{companion.KitTiny, "tiny &middot; 4&times;3 cells", 22, "past the point where counting them means anything."},
	}

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; how small can a kitten get?</h1>`)

	for _, r := range rows {
		var row strings.Builder
		fit := 0
		for _, t := range []float64{3.0, 3.6} {
			c := canvas.New(72, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			scape.NewShore(seed, false).Update(c, t, scape.Activity{
				Working: true, Level: math.Min(1, 0.3+float64(r.want)*0.06), ContextUsed: 0.3,
			})
			top := c.H - 2 - chh
			cat.Draw(c.Near(), 3, top, t, companion.Working)
			fit = cat.DrawKittensAt(c.Near(), 3, top, r.want, c.W-1, t, seed, r.scale)
			fmt.Fprintf(&row, `<div class="win">%s</div>`, c.HTMLFragment(12))
		}
		short := ""
		if fit < r.want {
			short = fmt.Sprintf(" &mdash; <b>only %d of %d fit</b>", fit, r.want)
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="rg">%d wanted &middot; %d drawn</div><div class="nt">%s%s</div></div>`+
			`<div style="display:flex;gap:12px">%s</div></div>`,
			r.name, r.want, fit, r.note, short, row.String())
	}
	return canvas.HTMLPage("asciiscapes — kitten scales", b.String())
}
