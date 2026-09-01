package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

func dayPage(seed int64) string {
	times := []struct {
		t    float64
		name string
	}{
		{0.00, "midnight"},
		{0.18, "first light"},
		{0.25, "dawn"},
		{0.50, "noon"},
		{0.75, "dusk"},
		{0.88, "late evening"},
	}

	cat := companion.NewCat()
	_, chh := cat.Size()

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}` +
		`.grid{display:flex;flex-wrap:wrap;gap:16px}` +
		`.lv{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:5px}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; the day cycle</h1>`)
	b.WriteString(`<div class="grid">`)
	for _, tt := range times {
		c := canvas.New(56, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		scape.NewShore(seed, false).Update(c, 3.0, scape.Activity{
			Working: true, Level: 0.5, TimeOfDay: tt.t, ContextUsed: 0.3,
		})
		cat.Draw(c.Near(), 4, c.H-2-chh, 3.0, companion.Working)
		fmt.Fprintf(&b, `<div><div class="lv">%s &middot; %.2f</div><div class="win">%s</div></div>`,
			strings.ToUpper(tt.name), tt.t, c.HTMLFragment(12))
	}
	b.WriteString(`</div>`)
	return canvas.HTMLPage("asciiscapes — day cycle", b.String())
}
