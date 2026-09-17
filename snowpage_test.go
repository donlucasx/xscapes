package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestSnowPage writes the snow candidates as one page when XSCAPES_SNOWPAGE
// names a file: his note of 2026-09-16 20:31 on the vista ("the snow on the
// mountain peaks seems pretty bright at night ... the snow detail could be
// better"). Every style at his band (133 columns, 24 rows) at five hours,
// plus a close crop of the massif, plus the measured lumas.
func TestSnowPage(t *testing.T) {
	out := os.Getenv("XSCAPES_SNOWPAGE")
	if out == "" {
		t.Skip("set XSCAPES_SNOWPAGE=<file> to write the page")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	const W, H = 133, 24
	pal := &canvas.HTMLPalette{}
	tail := []reduce.Line{
		{Text: "read  internal/scenes/forest.go", Age: 0.6},
		{Text: "edit  snowpage_test.go  +120", Age: 0.3},
		{Text: "shell go test ./...  ok", Age: 0.0},
	}
	state := func(tod float64) reduce.State {
		return reduce.State{Act: scape.Activity{Working: true, Level: 0.4, ContextUsed: 0.35, TimeOfDay: tod, TodoDone: 3, Tokens: 17_000_000}, Pose: companion.Working, Tail: tail}
	}
	hours := []struct {
		tod  float64
		name string
	}{{0, "00:00 midnight"}, {20.5 / 24, "20:30, his screenshot"}, {0.78, "18:43 dusk"}, {0.24, "05:46 dawn"}, {0.5, "12:00 noon"}}
	was := scenes.SnowPick
	defer func() { scenes.SnowPick = was }()
	var b strings.Builder
	// Lumas after quantisation to the cube, which is what the terminal draws.
	lum := func(c term.RGB) float64 {
		q := term.FromIndex256(c.Index256Keeping())
		return 0.299*float64(q.R) + 0.587*float64(q.G) + 0.114*float64(q.B)
	}
	abs := func(v float64) float64 { return math.Abs(v) }
	hname := func(tod float64) string { m := int(tod*24*60 + 0.5); return fmt.Sprintf("%02d:%02d", m/60, m%60) }
	for i, st := range scenes.SnowStyles {
		scenes.SnowPick = i
		fmt.Fprintf(&b, `<section><h2>S%d %s</h2><p class="note">%s</p>`, i, st.Name, st.Note)
		b.WriteString(`<table class="lum"><tr><th>hour</th><th>snow</th><th>lit face</th><th>far range</th><th>moon</th><th>snow − face</th><th>snow − far</th><th>moon − snow</th></tr>`)
		for _, h := range hours {
			rt := scenes.VistaRangeTones(i, h.tod)
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%.0f</td><td>%.0f</td><td>%.0f</td><td>%.0f</td><td>%.0f</td><td>%.0f</td><td>%.0f</td></tr>`, h.name, lum(rt.Snow), lum(rt.MidLit), lum(rt.Far), lum(rt.Moon), lum(rt.Snow)-lum(rt.MidLit), lum(rt.Snow)-lum(rt.Far), lum(rt.Moon)-lum(rt.Snow))
		}
		b.WriteString(`</table>`)
		// The day every half hour: the smallest of each gap and its hour.
		type worst struct {
			v  float64
			at float64
		}
		minFace, maxFace, minFar, minFarMid := worst{1e9, 0}, worst{-1e9, 0}, worst{1e9, 0}, worst{1e9, 0}
		for hh := 0.0; hh < 24; hh += 0.5 {
			rt := scenes.VistaRangeTones(i, hh/24)
			g := lum(rt.Snow) - lum(rt.MidLit)
			if g < minFace.v {
				minFace = worst{g, hh / 24}
			}
			if g > maxFace.v {
				maxFace = worst{g, hh / 24}
			}
			if d := abs(lum(rt.Snow) - lum(rt.Far)); d < minFar.v {
				minFar = worst{d, hh / 24}
			}
			if d := abs(lum(rt.Far) - lum(rt.Mid)); d < minFarMid.v {
				minFarMid = worst{d, hh / 24}
			}
		}
		fmt.Fprintf(&b, `<p class="sweep">Over the day, every half hour, after quantisation: snow over the lit face from <b>%.0f</b> (%s) to <b>%.0f</b> (%s) · snow from the far range never closer than <b>%.0f</b> (%s) · far from mid range never closer than <b>%.0f</b> (%s).</p>`,
			minFace.v, hname(minFace.at), maxFace.v, hname(maxFace.at), minFar.v, hname(minFar.at), minFarMid.v, hname(minFarMid.at))
		for _, h := range hours {
			f := renderVista(t, W, H, state(h.tod), 3.0)
			_, _, _, bandTop := f.vista.Layout()
			fmt.Fprintf(&b, `<div class="hour"><div class="lab">%s</div><div class="ctx">%s</div><div class="ctx">%s</div></div>`, h.name,
				f.c.HTMLFragmentClassed(7, term.Profile256, pal),
				f.c.HTMLFragmentCropClassed(28, 0, 118, max(1, bandTop-8), 12, term.Profile256, pal))
		}
		b.WriteString(`</section>`)
	}
	scenes.SnowPick = was
	page := `<meta charset="utf-8"><title>The Snow</title>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap">
<style>
:root{--bg:#131311;--ink:#d9d6cf;--mute:#8d8a82;--rule:#2b2a27;--acc:#87afff}
html{color-scheme:dark}
body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.55 "JetBrains Mono",Menlo,"SF Mono",monospace;padding-block:0 72px;padding-inline:24px}
main{max-width:1240px;margin:0 auto}
h1{font-size:22px;font-weight:700;margin:36px 0 8px;text-wrap:balance}
.lede{color:var(--mute);max-width:66ch;margin:0 0 20px}
section{border-top:1px solid var(--rule);padding:26px 0 6px}
h2{font-size:17px;font-weight:700;margin:0 0 6px}
.note{color:var(--mute);max-width:66ch;margin:0 0 14px}
.lum{border-collapse:collapse;font-size:13px;margin:0 0 16px;font-variant-numeric:tabular-nums}
.lum th,.lum td{text-align:right;padding:3px 12px 3px 0;border-bottom:1px solid var(--rule)}
.lum th:first-child,.lum td:first-child{text-align:left}
.lum th{color:var(--mute);font-weight:400}
.sweep{color:var(--mute);font-size:13px;max-width:80ch;margin:0 0 14px}
.sweep b{color:var(--ink)}
.hour{margin:0 0 18px}
.lab{color:var(--mute);font-size:12px;margin-bottom:4px}
.ctx{overflow-x:auto;margin-bottom:6px}
pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}
.foot{color:var(--mute);font-size:13px;max-width:66ch;border-top:1px solid var(--rule);padding-top:16px;margin-top:20px}
code{font-size:13px;color:var(--acc)}
` + pal.CSS() + `
</style>
<main>
<h1>The snow, second round</h1>
<p class="lede">Your notes: bright at night, the detail could be better; then, on the first six, the moonlit ones did not look right, the layers of the range merged at some hours, and the contrast between snow and rock should be more subtle. This round the snow is the lit rock face itself lifted a little toward white, so it takes the rock's own light at every hour and can only be as far from the far range as the face is; and its edge follows the terrain: deeper in gullies and saddles, shorter on steep crests, deeper on the lee faces. S1 to S3 differ only in how much cover; S4 is S2 with a fainter lift. Each at your band, 133 by 24, at five hours with a close crop; the tables are lumas after quantisation, and the line under each is the whole day every half hour. Style 0 is today, and stays on the page's clip whatever you pick. Say the number.</p>
` + b.String() + `
<p class="foot">Made from the tree: <code>XSCAPES_SNOWPAGE=&lt;file&gt; go test -run TestSnowPage .</code> · 133x24, seed 7, the 256-colour cube, working at 0.4, three stars, the counter at 17M.</p>
</main>`
	if err := os.WriteFile(out, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s (%d bytes)", out, len(page))
}
