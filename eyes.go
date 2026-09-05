package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// eyesPage is the study for his pick on the companion's eyes: the eye cells
// are gaps in the body bitmap, so they show the scene behind the head -- two
// blue holes by day. Three fills, three hours, the live composition at his
// Terminal.app geometry, as the 256 cube paints it.
func eyesPage(seed int64) string {
	const w, h = 120, 24
	frame := func(tod float64, fill string) (*canvas.Canvas, int, int, int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		cat := companion.NewCat()
		cat.FaceLeft(true)
		cat.SetEyeFill(fill)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		sh.MoonX = lay.MoonX
		st := demoState(2, 0.3, tod)
		for i := 0; i < 30; i++ {
			sh.Update(c, 2-1.5+float64(i)/20, st.Act)
		}
		sh.Update(c, 2, st.Act)
		drawScene(c, sh, cat, lay, st, 2, seed, c.H-2-chh)
		return c, lay.CatX, c.H - 2 - chh, ccw
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start;margin-bottom:6px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:28px 0 6px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the companion's eyes</h1>`)
	b.WriteString(`<p class="nt">The eyes are characters plotted in two gaps the body bitmap leaves for them, ` +
		`so each eye cell shows whatever is behind the head. At night that is dark water and it reads as a ` +
		`socket with a shine. By day it is the sea, and the head has two blue holes in it. ` +
		`<b>holes</b> is what ships. <b>coat</b> paints the cell in the coat before the glyph. ` +
		`<b>socket</b> paints it in the coat darkened a step. ` +
		`Every frame below is the live composition at 120 columns, as the 256 cube paints it in both terminals; ` +
		`the crops are the head at three times the size.</p>`)

	cell := func(label, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div><div class="win">%s</div></div>`, label, body)
	}
	fills := []struct{ name, fill string }{
		{"holes (ships today)", companion.EyeFillNone},
		{"coat", companion.EyeFillCoat},
		{"socket", companion.EyeFillSocket},
	}
	for _, hr := range []struct {
		name string
		tod  float64
	}{{"11:30, midday", 0.479}, {"18:40, dusk", 0.778}, {"22:20, night", 0.931}} {
		fmt.Fprintf(&b, `<h2>%s</h2><div class="row">`, hr.name)
		for _, f := range fills {
			c, catX, top, ccw := frame(hr.tod, f.fill)
			cell(f.name, c.HTMLFragmentCropAs(catX-3, top-1, catX+ccw+3, top+7, 33, term.Profile256))
		}
		b.WriteString(`</div><div class="row">`)
		for _, f := range fills[:2] {
			c, _, _, _ := frame(hr.tod, f.fill)
			cell("whole frame, "+f.name, c.HTMLFragmentAs(9, term.Profile256))
		}
		b.WriteString(`</div>`)
	}
	return canvas.HTMLPage("xscapes — the companion's eyes", b.String())
}
