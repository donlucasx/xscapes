// gradientaudit measures the 256-colour sky and sea gradients hour by hour, at
// a given geometry, and writes what it found as a table and as a page.
//
// He asked for smoother gradients on Terminal.app and for a review before
// anything changes. This is the review's evidence: for each hour, how many
// distinct 256-colour tones the sky and the sea actually carry, how tall each
// band of one tone is, and the largest perceptual jump between neighbouring
// rows -- measured on the RENDERED frame, quantised the way the terminal
// shows it, at the size of his window, not on the palette's nominal colours.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	w := flag.Int("w", 124, "scape width")
	h := flag.Int("h", 27, "scape height (the rows the scape gets at his 62-row window)")
	out := flag.String("html", "", "write the hour-by-hour page here")
	steps := flag.Int("steps", 48, "samples across the day (48 = every half hour)")
	flag.Parse()

	type hourStat struct {
		tod                float64
		skyRows, seaRows   int
		skyBands, seaBands []band
		skyMaxDE, seaMaxDE float64
		skyMaxAt, seaMaxAt int
		moon               term.RGB
		moonIdx            int
		frame256, frameTC  string
	}
	var stats []hourStat
	var pal canvas.HTMLPalette

	for i := 0; i < *steps; i++ {
		tod := float64(i) / float64(*steps)
		c := canvas.New(*w, *h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3}
		for k := 0; k < 14; k++ {
			sh.Update(c, 3.0+float64(k)/20, act)
		}
		// Column 2: clear of the companion and the litter, open sea and sky.
		// Half-row resolution: a split cell (U+2580) carries two tones.
		var tones []term.RGB
		for y := 0; y < c.H; y++ {
			ch, fg, bg := c.ResolveAt(2, y, term.Profile256)
			up, dn := bg, bg
			if ch == '▀' {
				up = fg
			}
			tones = append(tones, up, dn)
		}
		// The horizon is the biggest downward luma step in the upper part;
		// the sand begins where the shore says it does.
		hy := 0
		drop := 0.0
		for y := 1; y < c.H*3/4; y++ {
			if d := luma(c.BGAt(2, y-1)) - luma(c.BGAt(2, y)); d > drop {
				drop, hy = d, y
			}
		}
		sandTop := sh.SandTop()
		if sandTop <= hy || sandTop > c.H {
			sandTop = c.H
		}
		st := hourStat{tod: tod, skyRows: hy, seaRows: sandTop - hy}
		st.skyBands, st.skyMaxDE, st.skyMaxAt = bands(tones[:2*hy])
		st.seaBands, st.seaMaxDE, st.seaMaxAt = bands(tones[2*hy : 2*sandTop])
		st.seaMaxAt += hy * 2
		mx, my := sh.MoonPos()
		if mx >= 0 && my >= 0 && my < c.H && mx < c.W {
			_, fg, _ := c.ResolveAt(mx, my, term.Profile256)
			st.moon = fg
			st.moonIdx = fg.Index256()
		}
		st.frame256 = c.HTMLFragmentClassed(6, term.Profile256, &pal)
		st.frameTC = c.HTMLFragmentClassed(6, term.ProfileTrueColor, &pal)
		stats = append(stats, st)
	}

	// The table.
	fmt.Printf("%-6s %-6s %-9s %-8s %-6s %-9s %-8s %-10s\n", "time", "sky", "sky tones", "maxΔE@", "sea", "sea tones", "maxΔE@", "body 256")
	for _, st := range stats {
		fmt.Printf("%-6s %-6s %-9s %-8s %-6s %-9s %-8s %-10s\n",
			clock(st.tod),
			fmt.Sprintf("%d", st.skyRows), fmt.Sprintf("%d %s", len(st.skyBands), runs(st.skyBands)),
			fmt.Sprintf("%.0f@%d", st.skyMaxDE, st.skyMaxAt/2+1),
			fmt.Sprintf("%d", st.seaRows), fmt.Sprintf("%d %s", len(st.seaBands), runs(st.seaBands)),
			fmt.Sprintf("%.0f@%d", st.seaMaxDE, st.seaMaxAt/2+1),
			fmt.Sprintf("#%02x%02x%02x/%d", st.moon.R, st.moon.G, st.moon.B, st.moonIdx))
	}

	if *out == "" {
		return
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>xscapes gradients by the hour</title><style>
body{margin:0;background:#0b0b10;color:#c9c9d4;font:14px/1.5 -apple-system,Helvetica,Arial,sans-serif;padding:28px}
h1{font-size:20px;margin:0 0 6px}p{max-width:80ch;color:#8a8a96;margin:0 0 18px}
.hour{display:grid;grid-template-columns:70px 1fr 1fr 260px;gap:14px;align-items:start;margin-bottom:14px;padding-bottom:12px;border-bottom:1px solid #1c1c24}
.t{font-size:18px;font-weight:600;color:#f0f0f4}.lbl{font-size:10px;letter-spacing:.1em;text-transform:uppercase;color:#55555f;margin-bottom:4px}
pre{margin:0;font-family:Menlo,monospace;line-height:1.0;letter-spacing:0;display:inline-block}
.win{background:#000;border:1px solid #22222a;border-radius:4px;padding:4px 6px;overflow:hidden}
.m{font-size:12px;color:#8a8a96}.m b{color:#e8c27a;font-weight:600}
.sw{display:inline-block;width:14px;height:14px;border-radius:2px;vertical-align:middle;margin-right:2px;border:1px solid #333}
` + "</style></head><body>")
	b.WriteString(`<h1>xscapes: the sky and the sea, hour by hour, as Terminal.app shows them</h1>
<p>Left: the 256-colour frame, which is what you get. Middle: the truecolor frame the palette asks for. Right: what the 256 sky and sea actually carry in column 2, at half-row resolution: how many distinct tones, how tall each band of one tone is (top to bottom), and the largest perceptual jump between neighbouring half-rows (CIE76 &Delta;E; under 5 is barely visible, over 20 is a hard edge). Rendered at ` + fmt.Sprintf("%dx%d", *w, *h) + `, the scape your 124x62 window gets.</p>`)
	for _, st := range stats {
		b.WriteString(`<div class="hour"><div class="t">` + clock(st.tod) + `</div>`)
		b.WriteString(`<div><div class="lbl">256, what you get</div><div class="win">` + st.frame256 + `</div></div>`)
		b.WriteString(`<div><div class="lbl">truecolor, what the palette asks for</div><div class="win">` + st.frameTC + `</div></div>`)
		b.WriteString(`<div class="m"><div class="lbl">sky, ` + fmt.Sprintf("%d rows", st.skyRows) + `</div>`)
		b.WriteString(swatches(st.skyBands) + fmt.Sprintf(`<br><b>%d</b> tones, bands %s, max &Delta;E <b>%.0f</b> at row %d<br><br>`, len(st.skyBands), runs(st.skyBands), st.skyMaxDE, st.skyMaxAt/2+1))
		b.WriteString(`<div class="lbl">sea, ` + fmt.Sprintf("%d rows", st.seaRows) + `</div>`)
		b.WriteString(swatches(st.seaBands) + fmt.Sprintf(`<br><b>%d</b> tones, bands %s, max &Delta;E <b>%.0f</b> at row %d<br><br>`, len(st.seaBands), runs(st.seaBands), st.seaMaxDE, st.seaMaxAt/2+1))
		b.WriteString(fmt.Sprintf(`<div class="lbl">sun / moon</div><span class="sw" style="background:#%02x%02x%02x"></span> 256 index %d`, st.moon.R, st.moon.G, st.moon.B, st.moonIdx))
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`<style>` + pal.CSS() + `</style></body></html>`)
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}

// bands folds a column of half-row tones into runs of one tone, and finds
// the largest perceptual step between neighbours.
// band is a run of half-rows carrying one tone.
type band struct {
	c    term.RGB
	rows int
}

func bands(tones []term.RGB) (out []band, maxDE float64, at int) {
	for i, c := range tones {
		if len(out) > 0 && out[len(out)-1].c == c {
			out[len(out)-1].rows++
			continue
		}
		if i > 0 {
			if d := deltaE(tones[i-1], c); d > maxDE {
				maxDE, at = d, i
			}
		}
		out = append(out, band{c, 1})
	}
	return out, maxDE, at
}

func runs(bs []band) string {
	var parts []string
	for _, b := range bs {
		parts = append(parts, fmt.Sprintf("%.1f", float64(b.rows)/2))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func swatches(bs []band) string {
	var b strings.Builder
	for _, x := range bs {
		fmt.Fprintf(&b, `<span class="sw" style="background:#%02x%02x%02x;width:%dpx"></span>`, x.c.R, x.c.G, x.c.B, 6+x.rows*4)
	}
	return b.String()
}

func clock(tod float64) string {
	m := int(math.Round(tod * 24 * 60))
	return fmt.Sprintf("%02d:%02d", (m/60)%24, m%60)
}

func luma(c term.RGB) float64 { return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B) }

// deltaE is CIE76 between two sRGB colours: good enough to rank steps.
func deltaE(a, b term.RGB) float64 {
	l1, a1, b1 := lab(a)
	l2, a2, b2 := lab(b)
	return math.Sqrt((l1-l2)*(l1-l2) + (a1-a2)*(a1-a2) + (b1-b2)*(b1-b2))
}

func lab(c term.RGB) (float64, float64, float64) {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	x := (0.4124*r + 0.3576*g + 0.1805*bl) / 0.95047
	y := (0.2126*r + 0.7152*g + 0.0722*bl) / 1.0
	z := (0.0193*r + 0.1192*g + 0.9505*bl) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	fx, fy, fz := f(x), f(y), f(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}
