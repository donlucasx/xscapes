package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

type meterStyle int

const (
	meterNone meterStyle = iota
	meterAlways
	meterThreshold
	meterInSand
)

var (
	moonLabelDim  = term.RGB{R: 150, G: 150, B: 166}
	moonLabelWarn = term.RGB{R: 244, G: 226, B: 176}
	sandLabel     = term.RGB{R: 168, G: 152, B: 128}
)

// contextScene renders one shore at a given context level with one treatment of
// the numeric readout, so the treatments can be compared rather than described.
func contextScene(seed int64, used float64, style meterStyle) *canvas.Canvas {
	c := canvas.New(56, 20, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.Update(c, 3.0, scape.Activity{Working: true, Level: 0.55, ContextUsed: used})

	cat := companion.NewCat()
	_, chh := cat.Size()
	cat.Draw(c.Near(), 4, c.H-2-chh-3, 3.0, companion.Working)
	writeInSand(c, activity[2:], 17)

	pct := fmt.Sprintf("%.0f%%", (1-used)*100)
	mx, my := sh.MoonPos()

	switch style {
	case meterAlways:
		label(c, mx-len(pct)/2, my+3, pct, moonLabelDim)
	case meterThreshold:
		// Silent while there is nothing to think about; appears quietly at 65%,
		// brightens at 85%. The reveal also teaches what the moon means.
		if used >= 0.85 {
			label(c, mx-len(pct)/2, my+3, pct+" left", moonLabelWarn)
		} else if used >= 0.65 {
			label(c, mx-len(pct)/2, my+3, pct, moonLabelDim)
		}
	case meterInSand:
		txt := "ctx " + pct
		label(c, c.W-len(txt)-2, c.H-2, txt, sandLabel)
	}
	return c
}

func label(c *canvas.Canvas, x, y int, s string, col term.RGB) {
	(&companion.Sprite{Rows: []string{s}, Body: col, Alpha: 1, Opaque: true}).Draw(c.Near(), x, y)
}

func contextPage(seed int64) string {
	levels := []float64{0.30, 0.70, 0.92}
	variants := []struct {
		style meterStyle
		name  string
		note  string
	}{
		{meterNone, "A &middot; moon only", "Purest. The phase and altitude carry it alone. Beautiful, and it has to be learned once before it means anything."},
		{meterAlways, "B &middot; always beside the moon", "Anchored to the thing it describes, so the number teaches the moon every time you glance. But it is noise for the 60% of a session when you do not care."},
		{meterThreshold, "C &middot; appears at 65%, brightens at 85%", "Silent while there is nothing to think about. The reveal is itself the warning, and it teaches the metaphor at the moment you need it."},
		{meterInSand, "D &middot; in the sand with the activity", "Grouped with the other facts rather than floating in the sky. Consistent, but it reads as a status line and pulls the eye down."},
	}

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}` +
		`.row{display:flex;gap:14px}.row>div{text-align:center}` +
		`.lv{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:5px}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; context: how much number is too much?</h1>`)

	for _, v := range variants {
		var row strings.Builder
		for _, lv := range levels {
			c := contextScene(seed, lv, v.style)
			fmt.Fprintf(&row, `<div><div class="lv">%.0f%% USED</div><div class="win">%s</div></div>`,
				lv*100, c.HTMLFragment(12))
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div></div><div class="row">%s</div></div>`,
			v.name, v.note, row.String())
	}
	return canvas.HTMLPage("asciiscapes — context meter", b.String())
}
