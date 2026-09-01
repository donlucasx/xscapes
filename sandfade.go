package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// sandFadePage is a tuner for one number: how far the lower beach sinks toward
// black. Every strength is pre-rendered and a slider swaps between them, which
// is the only honest way to judge it -- the question is where legibility stops
// improving and the beach stops reading as a beach, and that is a threshold you
// find by sweeping past it, not by arguing about it.
//
// The writing is the reason the beach exists, and the NEWEST line sits lowest,
// so the line that matters most is the one mid-tone sand fights hardest.
func sandFadePage(seed int64) string {
	work := []string{
		"read   internal/auth/handler.go  142 lines",
		"search rate.Limiter  3 files",
		"edit   internal/auth/handler.go  +18 -2",
		"write  internal/auth/limiter.go  64 lines",
		"shell  go test ./...  4.1s",
		"read   internal/auth/limiter_test.go  88 lines",
		"edit   internal/auth/limiter.go  +6 -1",
		"shell  go vet ./...  0.9s",
		"search TestLimiter  2 files",
		"edit   internal/auth/limiter_test.go  +24 -0",
		"shell  go test ./internal/auth  1.2s",
		"read   README.md  61 lines",
	}
	lines := make([]reduce.Line, 0, len(work))
	for i, s := range work {
		lines = append(lines, reduce.Line{Text: s, Age: 1 - float64(i+1)/float64(len(work))})
	}

	frame := func(fade, tod float64) string {
		const w, h = 92, 44
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX, sh.SkyRows, sh.SandRows, sh.SandFade = 0.28, 10, 20, fade
		act := scape.Activity{Working: true, Level: 0.7, TimeOfDay: tod, ContextUsed: 0.35}
		for i := 0; i < 14; i++ {
			sh.Update(c, 2+float64(i)/20, act)
		}
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: lines}, 3.1, seed, c.H-2-chh)
		return c.HTMLFragment(11)
	}

	steps := []float64{0, 0.15, 0.3, 0.45, 0.6, 0.7, 0.8, 0.9, 1.0}
	tods := []struct {
		v float64
		n string
	}{{0.366, "morning"}, {0.52, "midday"}, {0.02, "night"}}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}
	.fr{display:none}.fr.on{display:block}
	.rowline{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start}
	.col{display:flex;flex-direction:column;gap:5px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	#sl{width:520px}
	.bar{margin:14px 0 18px;font:12px ui-monospace,monospace;color:#d8d8e0}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; how far should the beach fall away?</h1>`)
	b.WriteString(`<p class="nt">The lower beach sinks toward the terminal's own near-black, squared ` +
		`so the sand stays sand for most of its depth and only lets go near the bottom. Drag the ` +
		`slider. Watch the newest line, at the very bottom: that is the one the mid-tone sand was ` +
		`fighting. Watch midday too, where the sand is brightest and the fade has the most work to ` +
		`do &mdash; and watch where the beach stops reading as a beach.</p>`)

	b.WriteString(`<div class="bar">fade <input id="sl" type="range" min="0" max="8" value="0" step="1">
	<span id="val">0.00</span></div>`)

	b.WriteString(`<div class="rowline">`)
	for _, td := range tods {
		b.WriteString(`<div class="col"><div class="lbl">` + td.n + `</div><div class="win">`)
		for i, f := range steps {
			cls := "fr"
			if i == 0 {
				cls = "fr on"
			}
			fmt.Fprintf(&b, `<div class="%s" data-i="%d">%s</div>`, cls, i, frame(f, td.v))
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`</div>`)

	fmt.Fprintf(&b, `<script>
var steps = %s;
var sl = document.getElementById('sl'), val = document.getElementById('val');
sl.addEventListener('input', function(){
  var i = +sl.value;
  val.textContent = steps[i].toFixed(2);
  document.querySelectorAll('.fr').forEach(function(el){
    el.classList.toggle('on', +el.dataset.i === i);
  });
});
</script>`, floatsJS(steps))

	return canvas.HTMLPage("asciiscapes — sand fade", b.String())
}

func floatsJS(v []float64) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%.2f", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// sandFadeContrast reports the legibility the fade actually buys, which is the
// whole argument for it.
//
// It renders the real scene and then READS THE PIXELS BACK -- the glyph colour
// the renderer actually chose against the background it actually painted, on
// the newest line. Recomputing the ink here from the same rule drawSand uses
// would only prove the rule agrees with itself, and that is exactly the check
// that missed the midday collapse: nominal sand said 175 while the row said 62.
func sandFadeContrast(seed int64, tod float64) []string {
	const w, h = 92, 44
	work := []string{"shell  go test ./internal/auth  1.2s", "read   README.md  61 lines"}
	lines := []reduce.Line{{Text: work[0], Age: 0.4}, {Text: work[1], Age: 0}}

	out := []string{}
	for _, fade := range []float64{0, 0.5, 0.8, 1.0} {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX, sh.SkyRows, sh.SandRows, sh.SandFade = 0.28, 10, 20, fade
		for i := 0; i < 14; i++ {
			sh.Update(c, 2+float64(i)/20, scape.Activity{
				Working: true, Level: 0.7, TimeOfDay: tod, ContextUsed: 0.35})
		}
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: lines}, 3.1, seed, c.H-2-chh)

		// The newest line is the lowest one. Find its painted cells.
		best, row := 0.0, -1
		near := c.Near()
		for y := c.H - 1; y >= c.H-4 && y >= 0; y-- {
			var sum float64
			var n int
			for x := lay.SandFrom; x < lay.SandTo && x < c.W; x++ {
				cell := near.Cells[y*c.W+x]
				if !cell.Set || cell.R == ' ' {
					continue
				}
				sum += math.Abs(luma(cell.FG) - luma(c.BG[y*c.W+x]))
				n++
			}
			if n >= 8 {
				row = y
				best = sum / float64(n)
				break
			}
		}
		if row < 0 {
			out = append(out, fmt.Sprintf("fade %.2f: no text found", fade))
			continue
		}
		out = append(out, fmt.Sprintf("fade %.2f: newest line on row %d, beach luma %5.1f, |ink-beach| = %5.1f",
			fade, row, luma(c.BG[row*c.W+lay.SandFrom+2]), best))
	}
	return out
}
