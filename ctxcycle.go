package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// ctxCyclePage is the context cycle as the live scene paints it: the moon (the
// sun by day) waning from full to new and sinking from high to the horizon as
// context is used, its shine on the water fading with it. The numeric readout
// decided in session 6 (appears at 65% used, brightens at 85%) is shown on a
// second row for each hour, because it is NOT in the live scene today.
func ctxCyclePage(seed int64) string {
	const w, h = 120, 24
	frame := func(tod, used float64, readout bool) (*canvas.Canvas, int, int, int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		sh.MoonX = lay.MoonX
		st := demoState(2, used, tod)
		for i := 0; i < 30; i++ {
			sh.Update(c, 2-1.5+float64(i)/20, st.Act)
		}
		sh.Update(c, 2, st.Act)
		drawScene(c, sh, cat, lay, st, 2, seed, c.H-2-chh)
		mx, my := sh.MoonPos()
		if readout {
			pct := fmt.Sprintf("%.0f%%", (1-used)*100)
			if used >= 0.85 {
				label(c, mx-len(pct)/2-2, my+3, pct+" left", moonLabelWarn)
			} else if used >= 0.65 {
				label(c, mx-len(pct)/2, my+3, pct, moonLabelDim)
			}
		}
		return c, mx, my, int(float64(c.H)*0.42) + 1
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:10px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:28px 0 6px}
	h3{font:500 11px ui-monospace,monospace;color:#8a8a99;letter-spacing:.08em;text-transform:uppercase;margin:12px 0 6px}
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#8a8a99;margin:8px 0 4px}
	td,th{padding:3px 10px 3px 0;text-align:left;vertical-align:top}
	th{color:#d8d8e0;font-weight:500}
	</style>`)
	b.WriteString(`<h1>The Context Cycle</h1>`)
	b.WriteString(`<p class="nt">Context remaining is the moon &mdash; the sun by day, same body. Two cues carry one number: ` +
		`<b>phase</b>, full at 0% used and new at 100%, and <b>altitude</b>, high in the sky with a fresh session and ` +
		`down at the horizon when it is spent (row = 22% + 62% &times; used, of the sky's height). The shine on the ` +
		`water fades with the lit fraction. Every frame below is the live composition at 120 columns, painted as the ` +
		`256 cube paints it in both terminals. The crops show the sky around the body at twice the size.</p>`)
	b.WriteString(`<p class="nt"><b>The text percentage: there is none in the live scene today.</b> Session 6 decided ` +
		`style C &mdash; a dim number under the moon from 65% used, a warm &ldquo;NN% left&rdquo; from 85% &mdash; and ` +
		`the brief marks it done, but the label was only ever drawn in the study page. The second row for each hour ` +
		`shows what that decided readout looks like on today's frames, so you can confirm it or drop it.</p>`)

	levels := []float64{0, 0.10, 0.25, 0.40, 0.55, 0.65, 0.75, 0.85, 0.95, 1.0}
	b.WriteString(`<table><tr><th>used</th><th>left</th><th>phase (lit)</th><th>readout, decided</th></tr>`)
	for _, u := range levels {
		ro := "silent"
		if u >= 0.85 {
			ro = "warm: \"NN% left\""
		} else if u >= 0.65 {
			ro = "dim number"
		}
		fmt.Fprintf(&b, `<tr><td>%.0f%%</td><td>%.0f%%</td><td>%.0f%%</td><td>%s</td></tr>`, u*100, (1-u)*100, (1-u)*100, ro)
	}
	b.WriteString(`</table>`)

	cell := func(label, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div><div class="win">%s</div></div>`, label, body)
	}
	for _, hr := range []struct {
		name string
		tod  float64
	}{{"22:20, night", 0.931}, {"11:30, midday", 0.479}} {
		fmt.Fprintf(&b, `<h2>%s</h2><h3>as it ships today</h3><div class="row">`, hr.name)
		for _, u := range levels {
			c, mx, _, hy := frame(hr.tod, u, false)
			cell(fmt.Sprintf("%.0f%% used", u*100), c.HTMLFragmentCropAs(mx-14, 0, mx+15, hy+3, 18, term.Profile256))
		}
		b.WriteString(`</div><h3>with the decided readout (not built)</h3><div class="row">`)
		for _, u := range levels {
			c, mx, _, hy := frame(hr.tod, u, true)
			cell(fmt.Sprintf("%.0f%% used", u*100), c.HTMLFragmentCropAs(mx-14, 0, mx+15, hy+3, 18, term.Profile256))
		}
		b.WriteString(`</div><h3>whole frames</h3><div class="row">`)
		for _, u := range []float64{0, 0.55, 0.85, 1.0} {
			c, _, _, _ := frame(hr.tod, u, false)
			cell(fmt.Sprintf("%.0f%% used", u*100), c.HTMLFragmentAs(8, term.Profile256))
		}
		b.WriteString(`</div>`)
	}
	return canvas.HTMLPage("xscapes — the context cycle", b.String())
}
