package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

// busyPage sweeps the activity level so the range can be judged. The question
// is not whether each level looks nice on its own -- it is whether idle and
// flat-out are distinguishable at a glance, which only a sweep can answer.
func busyPage(seed int64) string {
	levels := []struct {
		lv   float64
		name string
		note string
	}{
		{0.00, "idle", "waiting on you &mdash; the sea should settle, not stop"},
		{0.25, "reading", "a tool or two; barely stirred"},
		{0.50, "editing", "steady work"},
		{0.75, "building", "several tools in flight"},
		{1.00, "flat out", "subagents, tests, the lot"},
	}

	cat := companion.NewCat()
	_, chh := cat.Size()

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}` +
		`.lv{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:5px}</style>`)
	b.WriteString(`<h1>xscapes &mdash; how hard is it working?</h1>`)

	for _, l := range levels {
		st := companion.Working
		if l.lv == 0 {
			st = companion.Resting
		}
		// Two moments of each level, because a single frame cannot show speed.
		var row strings.Builder
		for _, t := range []float64{3.0, 3.5} {
			c := canvas.New(64, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			scape.NewShore(seed, false).Update(c, t, scape.Activity{
				Working: l.lv > 0, Level: l.lv, ContextUsed: 0.3,
			})
			cat.Draw(c.Near(), 4, c.H-2-chh, t, st)
			fmt.Fprintf(&row, `<div class="win">%s</div>`, c.HTMLFragment(12))
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="rg">Level %.2f</div><div class="nt">%s</div></div>`+
			`<div style="display:flex;gap:12px">%s</div></div>`,
			l.name, l.lv, l.note, row.String())
	}
	return canvas.HTMLPage("xscapes — activity level", b.String())
}
