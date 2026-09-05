package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// moonPage is the study for his question of 2026-09-05: "is this the best
// moon/sun we can draw within the existing limitations?" Five treatments,
// two hours, three context levels, at his Ghostty geometry, as the 256 cube
// paints them. The switches are study-only; what ships is the first column.
func moonPage(seed int64) string {
	const w, h = 130, 22
	type variant struct {
		name string
		set  func(*scape.Shore)
	}
	variants := []variant{
		{"ships today", func(*scape.Shore) {}},
		{"quad edge", func(s *scape.Shore) { s.MoonEdge = "quad" }},
		{"quad + sun without shadow", func(s *scape.Shore) { s.MoonEdge = "quad"; s.SunShadow = "sky" }},
		{"quad + no shadow + night halo", func(s *scape.Shore) { s.MoonEdge = "quad"; s.SunShadow = "sky"; s.MoonHalo = true }},
		{"hue rim (halves)", func(s *scape.Shore) { s.MoonRim = "hue" }},
	}
	frame := func(tod, used float64, v variant) (*canvas.Canvas, int, int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		v.set(sh)
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
		return c, mx, my
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
	</style>`)
	b.WriteString(`<h1>The Moon, Four Ways</h1>`)
	b.WriteString(`<p class="nt">The session 7 mockup was a truecolor preview. On the 256 cube, which every terminal now gets, ` +
		`its soft edge rounds to grey fringes and its navy sky has no entry, so the disc went solid and the night went grey. ` +
		`What the cube can still carry: <b>quad edge</b> samples the disc at four quarters per cell instead of two half-rows, ` +
		`doubling the horizontal resolution of the edge (exact in Ghostty; in Terminal.app a cell whose two top quarters ` +
		`differ takes a five-pixel notch). <b>sun without shadow</b> paints no unlit face by day, so the sun wanes as a crescent ` +
		`rather than showing a slate bite. <b>night halo</b> lightens the grey sky in a soft ring around the moon, the one ` +
		`piece of the mockup's softness the grey ramp has steps for. <b>hue rim</b> is the session 15 option: the outer ring ` +
		`one tone darker in the disc's own hue. Frames at 130 columns, 22 scape rows; crops of the sky at twice the size.</p>`)

	cell := func(label, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div><div class="win">%s</div></div>`, label, body)
	}
	for _, hr := range []struct {
		name string
		tod  float64
	}{{"13:20, day", 0.556}, {"22:20, night", 0.931}} {
		for _, used := range []float64{0.05, 0.30, 0.60} {
			fmt.Fprintf(&b, `<h2>%s &middot; %.0f%% used</h2><div class="row">`, hr.name, used*100)
			for _, v := range variants {
				c, mx, my := frame(hr.tod, used, v)
				cell(v.name, c.HTMLFragmentCropAs(mx-12, max(my-6, 0), mx+13, my+7, 16, term.Profile256))
			}
			b.WriteString(`</div>`)
		}
	}
	b.WriteString(`<h2>whole frames, 30% used</h2>`)
	for _, hr := range []struct {
		name string
		tod  float64
	}{{"13:20, day", 0.556}, {"22:20, night", 0.931}} {
		fmt.Fprintf(&b, `<h3>%s</h3><div class="row">`, hr.name)
		for _, v := range []variant{variants[0], variants[3]} {
			c, _, _ := frame(hr.tod, 0.30, v)
			cell(v.name, c.HTMLFragmentAs(8, term.Profile256))
		}
		b.WriteString(`</div>`)
	}
	return canvas.HTMLPage("xscapes — the moon, four ways", b.String())
}
