// redeyes renders the companion's WORRIED eyes in every candidate colour,
// through the product's own pipeline, at the three hours the eye study uses.
// His ruling 2026-09-06: "when 'something broke', eyes should go red, not
// orange (urgency)." The question this page answers is which red, and it is
// answered by looking at the rendered frame rather than at the source RGB --
// glyph colours are saturated 2.6x and quantised to the cube before they are
// painted, so the colour you choose is not the colour that lands.
//
// Nothing here edits internal/companion. The worried eye is two cells plotted
// on top of the body bitmap at a known offset, so this replots those two cells
// in each candidate -- byte-identical to what moving eyeWorried would draw.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

type cand struct {
	name string
	col  term.RGB
	note string
}

// The candidates, his peer session's shortlist. The amber is the baseline --
// what he is calling orange and asking to replace.
var cands = []cand{
	{"amber", term.RGB{R: 244, G: 176, B: 96}, "ships today"},
	{"orange-red", term.RGB{R: 255, G: 95, B: 0}, "nearest to today"},
	{"pure red", term.RGB{R: 255, G: 0, B: 0}, ""},
	{"darker red", term.RGB{R: 215, G: 0, B: 0}, ""},
	{"deep red", term.RGB{R: 175, G: 0, B: 0}, ""},
	{"salmon", term.RGB{R: 255, G: 95, B: 95}, ""},
	{"brick", term.RGB{R: 215, G: 95, B: 95}, ""},
	{"coral", term.RGB{R: 255, G: 135, B: 95}, ""},
}

var hours = []struct {
	name string
	tod  float64
}{{"11:30, midday", 0.479}, {"18:40, dusk", 0.778}, {"22:20, night", 0.931}}

const w, h = 120, 24

// compose mirrors live.go's layout for the mirrored composition, which is what
// ships. Copied rather than imported: live.go is package main.
func catX(catW int) int {
	right := 2 + w/32
	x := w - catW - right
	if x < 0 {
		x = 0
	}
	return x
}

// frame paints the live composition with the companion worried, then replots
// the two eye cells in col. Returns the canvas, the cat's origin and width.
func frame(tod float64, col term.RGB, seed int64) (*canvas.Canvas, int, int, int) {
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28 // mirrored layout
	cat := companion.NewCat()
	cat.FaceLeft(true)
	cat.SetEyeFill(companion.EyeFillNone) // holes -- his locked ruling
	ccw, chh := cat.Size()
	x, y := catX(ccw), h-2-chh

	// Something is broken, so the sea is not flat: an error arrives mid-work.
	act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.3, TimeOfDay: tod, TodoDone: 2, TodoTotal: 5}
	for i := 0; i < 30; i++ {
		sh.Update(c, 2-1.5+float64(i)/20, act)
	}
	sh.Update(c, 2, act)
	cat.Draw(c.Near(), x, y, 2, companion.Worried)

	// The eye cells, exactly as companion.eyes() places them when mirrored.
	a, b := ccw-1-2, ccw-1-6
	c.Near().Plot(x+a, y+2, 'o', col, 1)
	c.Near().Plot(x+b, y+2, 'o', col, 1)
	return c, x, y, ccw
}

// painted reads the colour back OFF the rendered frame -- the only number that
// means anything, since the glyph path saturates and quantises on the way out.
func painted(c *canvas.Canvas, x, y int) (term.RGB, int) {
	_, fg, _ := c.ResolveAt(x, y, term.Profile256)
	return fg, fg.Index256()
}

func main() {
	out := flag.String("html", "notes/charstudy/07-worried-red.html", "write the page here")
	seed := flag.Int64("seed", 7, "scene seed")
	flag.Parse()

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start;margin-bottom:6px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	.dup{color:#e0a0a0}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:28px 0 6px}
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#b8b8c4;margin:8px 0 18px}
	td,th{border:1px solid #2a2a32;padding:3px 9px;text-align:left}
	th{color:#d8d8e0;font-weight:500}
	.sw{display:inline-block;width:11px;height:11px;vertical-align:-1px;border:1px solid #2a2a32}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the worried eye, in red</h1>`)
	b.WriteString(`<p class="nt">His ruling: <b>&ldquo;when &lsquo;something broke&rsquo;, eyes should go red, not orange (urgency).&rdquo;</b> ` +
		`Only the worried colour moves; the glyph, the pose and the eyes-as-holes all stay. ` +
		`Every frame is the live composition at 120 columns with the companion worried, as the 256 cube paints it in ` +
		`both terminals &mdash; the crops are the head at three times the size. ` +
		`<b>Judge these as rendered, not as the names.</b> Glyph colours are pushed 2.6x away from grey and then ` +
		`snapped to the cube before they are painted, so the index in the table is what actually lands on the screen; ` +
		`two candidates that read as different colours here can be the same cell.</p>`)

	// What actually lands, measured off the frame rather than computed.
	idx := map[int][]string{}
	b.WriteString(`<h2>What actually lands on the screen</h2><table><tr><th>candidate</th><th>chosen</th><th>painted</th><th>index</th><th></th></tr>`)
	for _, cd := range cands {
		c, x, y, ccw := frame(0.931, cd.col, *seed)
		got, i := painted(c, x+ccw-1-2, y+2)
		idx[i] = append(idx[i], cd.name)
		fmt.Fprintf(&b, `<tr><td><b>%s</b></td>`+
			`<td><span class="sw" style="background:rgb(%d,%d,%d)"></span> %d,%d,%d</td>`+
			`<td><span class="sw" style="background:rgb(%d,%d,%d)"></span> %d,%d,%d</td>`+
			`<td>%d</td><td>%s</td></tr>`,
			cd.name, cd.col.R, cd.col.G, cd.col.B, cd.col.R, cd.col.G, cd.col.B,
			got.R, got.G, got.B, got.R, got.G, got.B, i, cd.note)
	}
	b.WriteString(`</table>`)
	var dups []string
	for i, ns := range idx {
		if len(ns) > 1 {
			dups = append(dups, fmt.Sprintf("index %d is shared by %s", i, strings.Join(ns, " and ")))
		}
	}
	if len(dups) > 0 {
		fmt.Fprintf(&b, `<p class="nt dup"><b>Collision:</b> %s &mdash; those are one colour on screen, not two.</p>`, strings.Join(dups, "; "))
	} else {
		b.WriteString(`<p class="nt">All eight survive the boost as distinct cube entries. Nothing collides.</p>`)
	}

	cell := func(label, sub, body string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%s</i></div><div class="win">%s</div></div>`, label, sub, body)
	}
	for _, hr := range hours {
		fmt.Fprintf(&b, `<h2>%s</h2><div class="row">`, hr.name)
		for _, cd := range cands {
			c, x, y, ccw := frame(hr.tod, cd.col, *seed)
			_, i := painted(c, x+ccw-1-2, y+2)
			cell(cd.name, fmt.Sprintf("%d", i), c.HTMLFragmentCropAs(x-3, y-1, x+ccw+3, y+7, 33, term.Profile256))
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`<h2>In the whole frame &mdash; 22:20, night</h2>`)
	b.WriteString(`<p class="nt">The crop is how you compare them; this is how he will actually meet it. ` +
		`Two cells out of a 120-column scene is the whole cue.</p><div class="row">`)
	for _, cd := range []cand{cands[0], cands[2], cands[3]} {
		c, _, _, _ := frame(0.931, cd.col, *seed)
		cell(cd.name, "", c.HTMLFragmentAs(9, term.Profile256))
	}
	b.WriteString(`</div>`)

	page := canvas.HTMLPage("xscapes — the worried eye, in red", b.String())
	if err := os.WriteFile(*out, []byte(page), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
