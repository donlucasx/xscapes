package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// cubePage is the day cycle as Terminal.app actually paints it.
//
// It exists because every other HTML study in this repo renders true RGB,
// including -day, and that is why the sea and the sky could spend most of a
// working day as flat grey without a single test or preview noticing. A
// preview that shows colours the target terminal cannot produce is not a
// preview of the target terminal.
//
// Each hour is rendered twice: what the palette asks for, and what 256 colours
// can hold. The swatch strip underneath names the four background colours and
// the index each one lands on, so a collapse is readable as a number and not
// only as a picture.
func cubePage(seed int64) string {
	times := []struct {
		t    float64
		name string
	}{
		{0.00, "midnight"},
		{0.25, "dawn"},
		{0.375, "mid-morning"},
		{0.50, "noon"},
		{0.625, "mid-afternoon"},
		{0.75, "dusk"},
	}

	cat := companion.NewCat()
	cat.FaceLeft(true)
	_, chh := cat.Size()

	frame := func(t float64, p term.Profile) string {
		c := canvas.New(64, 20, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		for i := 0; i < 10; i++ {
			sh.Update(c, 3.0+float64(i)/20, scape.Activity{
				Working: true, Level: 0.5, TimeOfDay: t, ContextUsed: 0.3})
		}
		cat.Draw(c.Near(), c.W-14, c.H-2-chh, 3.0, companion.Working)
		return c.HTMLFragmentAs(11, p)
	}

	// bands counts what a 256 terminal really paints down one column, so the
	// caption under each hour is measured rather than asserted.
	bands := func(t float64) (int, int) {
		c := canvas.New(64, 20, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		for i := 0; i < 10; i++ {
			sh.Update(c, 3.0+float64(i)/20, scape.Activity{
				Working: true, Level: 0.5, TimeOfDay: t, ContextUsed: 0.3})
		}
		seen := map[int]bool{}
		run, worst, prev := 0, 0, -1
		for y := 0; y < c.H; y++ {
			i := c.BGAt(2, y).Index256()
			seen[i] = true
			if i == prev {
				run++
			} else {
				run = 1
			}
			if run > worst {
				worst = run
			}
			prev = i
		}
		return len(seen), worst
	}

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}` +
		`.row{display:flex;gap:14px;align-items:flex-start;margin-bottom:18px}` +
		`.lv{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:5px;text-transform:uppercase}` +
		`.cap{font-size:11px;color:#6f6f7c;margin-top:6px}` +
		`.sw{display:flex;gap:4px;margin-top:8px;flex-wrap:wrap}` +
		`.ch{font-size:9px;color:#8a8a94;text-align:center;width:74px}` +
		`.bar{height:16px;border-radius:3px;border:1px solid #2a2a32}` +
		`.hd{font-size:11px;color:#9a9aa6;margin:22px 0 10px;letter-spacing:.06em}</style>`)
	b.WriteString(`<h1>xscapes &mdash; what Terminal.app actually paints</h1>`)
	b.WriteString(`<div class="cap">Left: the palette's own RGB, which is what every other study in ` +
		`this repo shows. Right: the same frame through the real 256 quantiser, which is what ` +
		`Lucas sees. Swatches read: asked for &rarr; shown, with the xterm index.</div>`)

	for _, tt := range times {
		n, worst := bands(tt.t)
		b.WriteString(`<div class="hd">` + strings.ToUpper(tt.name) +
			fmt.Sprintf(` &middot; %.3f &middot; %d distinct background colours down one column, longest flat run %d rows`,
				tt.t, n, worst) + `</div>`)
		b.WriteString(`<div class="row">`)
		fmt.Fprintf(&b, `<div><div class="lv">palette RGB</div><div class="win">%s</div></div>`,
			frame(tt.t, term.ProfileTrueColor))
		fmt.Fprintf(&b, `<div><div class="lv">256 &middot; Terminal.app</div><div class="win">%s</div></div>`,
			frame(tt.t, term.Profile256))
		b.WriteString(`<div><div class="lv">backgrounds</div><div class="sw">`)
		p := scape.PaletteAt(tt.t)
		for _, f := range []struct {
			n string
			c term.RGB
		}{{"SkyTop", p.SkyTop}, {"SkyHorizon", p.SkyHorizon}, {"SeaFar", p.SeaFar}, {"SeaNear", p.SeaNear}} {
			shown := term.FromIndex256(f.c.Index256())
			kind := "cube"
			if f.c.Index256() >= 232 {
				kind = "GREY"
			}
			fmt.Fprintf(&b, `<div class="ch"><div class="bar" style="background:rgb(%d,%d,%d)"></div>`+
				`<div class="bar" style="background:rgb(%d,%d,%d)"></div>%s<br>%d %s</div>`,
				f.c.R, f.c.G, f.c.B, shown.R, shown.G, shown.B, f.n, f.c.Index256(), kind)
		}
		b.WriteString(`</div></div></div>`)
	}

	// The cube's own inventory, so the choice of colours is checkable.
	b.WriteString(`<div class="hd">EVERY CUBE COLOUR THAT SURVIVES AS A BLUE (B highest, B above R)</div>`)
	type ent struct {
		i int
		c term.RGB
		l float64
	}
	var blues []ent
	for i := 16; i <= 231; i++ {
		c := term.FromIndex256(i)
		if c.Index256() != i || !(c.B > c.R && c.B >= c.G) {
			continue
		}
		blues = append(blues, ent{i, c, 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)})
	}
	sort.Slice(blues, func(a, b int) bool { return blues[a].l < blues[b].l })
	b.WriteString(`<div class="sw">`)
	for _, e := range blues {
		fmt.Fprintf(&b, `<div class="ch"><div class="bar" style="background:rgb(%d,%d,%d)"></div>%d<br>luma %.0f</div>`,
			e.c.R, e.c.G, e.c.B, e.i, e.l)
	}
	b.WriteString(`</div>`)
	return canvas.HTMLPage("xscapes - the 256 cube", b.String())
}
