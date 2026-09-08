package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

var hours = []struct {
	name string
	tod  float64
}{{"11:30, midday", 0.479}, {"18:40, dusk", 0.778}, {"22:20, night", 0.931}}

// painted reads a colour back OFF the rendered frame. The only number that
// means anything: glyph colours are saturated and quantised on the way out, so
// the coat you choose is not necessarily the coat that lands.
func painted(c *canvas.Canvas, x, y int) (term.RGB, int) {
	_, fg, _ := c.ResolveAt(x, y, term.Profile256)
	return fg, fg.Index256()
}

// bodyCell finds a cell inside the shell, for reading the coat back.
func bodyCell(px, py int) (int, int) { return px + 6, py + 3 }

func main() {
	out := flag.String("html", "notes/charstudy/08-crab.html", "write the page here")
	round := flag.Int("round", 1, "which study round to write")
	seed := flag.Int64("seed", 7, "scene seed")
	dumpOnly := flag.String("dump", "", "dump silhouettes as quadrant text and exit")
	flag.Parse()
	if *dumpOnly == "disc" {
		discProbe(*seed)
		return
	}
	if *dumpOnly == "ramp" {
		rampProbe(*seed)
		return
	}
	if *dumpOnly == "hairline" {
		hairlineProbe(*seed)
		return
	}
	if *dumpOnly == "pace" {
		paceProbe(*seed, 8, 5)
		return
	}
	if *dumpOnly == "narrow" {
		narrowProbe()
		return
	}
	if *dumpOnly == "cols" {
		probeCols()
		return
	}
	if *dumpOnly == "size" {
		for _, s := range []struct {
			n string
			r []string
		}{{"the cat", companion.CatBody}, {"pincer", pincerBody}, {"hero", heroBody},
			{"broad", broadBody}, {"stalker2", stalker2Body}, {"owl (s19)", nil}} {
			if s.r == nil {
				continue
			}
			w, h, f := extent(s.n, s.r)
			fmt.Printf("%-10s ink %2d x %d cells, %2d of 84 cells carry ink (%.0f%%)\n", s.n, w, h, f, float64(f)/84*100)
		}
		return
	}
	if *dumpOnly != "" {
		dump("broad", broadBody, 24, 28)
		dump("lowrider", lowriderBody, 24, 28)
		dump("stalker", stalkerBody, 24, 28)
		dump("fiddler", fiddlerBody, 24, 28)
		dump("pincer", pincerBody, 24, 28)
		dump("hero", heroBody, 24, 28)
		dump("wide", wideBody, 24, 28)
		dump("stalker2", stalker2Body, 24, 28)
		dump("pincer young", pincerYoung, 12, 16)
		dump("crablet pincer", crabletPincer, 12, 16)
		dump("crablet hero", crabletHero, 12, 16)
		for _, s := range []struct {
			n string
			u []string
		}{{"pincer rest", pincerRest}, {"pincer ask", pincerAsk}, {"pincer done", pincerDone}, {"pincer worried", pincerWorried}, {"hero ask", heroAsk}, {"hero worried", heroWorried}} {
			dump(s.n, append(append([]string{}, s.u...), pincerLower...), 24, 28)
		}
		dump("crab young", crabYoung, 12, 16)
		dump("fiddler young", fiddlerYoung, 12, 16)
		return
	}

	if *round == 7 {
		if err := os.WriteFile(*out, []byte(page7(*seed, 8, 5)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}
	if *round == 6 {
		if err := os.WriteFile(*out, []byte(page6(*seed, 8, 5)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}
	if *round == 5 {
		if err := os.WriteFile(*out, []byte(page5(*seed, 8, 5)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}
	if *round == 4 {
		if err := os.WriteFile(*out, []byte(page4(*seed, 8, 5)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}
	if *round == 3 {
		if err := os.WriteFile(*out, []byte(page3(*seed, 8, 5)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}
	if *round == 2 {
		if err := os.WriteFile(*out, []byte(page2(*seed)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(*out)
		return
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	.note{font:12px/1.5 ui-sans-serif,system-ui;color:#8a8a99;max-width:34ch;margin:2px 0 0}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 6px;padding-top:14px;border-top:1px solid #23232a}
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#b8b8c4;margin:8px 0 18px}
	td,th{border:1px solid #2a2a32;padding:3px 9px;text-align:left}
	th{color:#d8d8e0;font-weight:500}
	.sw{display:inline-block;width:11px;height:11px;vertical-align:-1px;border:1px solid #2a2a32}
	.warn{color:#e0a0a0}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the crab</h1>`)
	b.WriteString(`<p class="nt">His brief: a crab that lives on the beach, coloured to stand out, ` +
		`&ldquo;maybe a salmon pink&rdquo;, in the same style as the existing companions, a few alternatives to choose from, ` +
		`the main companion first. Four silhouettes and six coats below, every one drawn through the companion ` +
		`pipeline in the shipped cat's box &mdash; a 24x28 bitmap halved and quartered into quadrant glyphs, ` +
		`12 cells wide by 7 tall &mdash; and rendered over the real shore as the 256 cube paints it. ` +
		`Nothing is chosen and nothing is installed.</p>`)
	b.WriteString(`<p class="nt">A crab is the first candidate that is <b>wide where the cat is tall</b>. ` +
		`That is the interesting part: the box has always been filled by an animal that sits up, and the shell ` +
		`fills it the other way round. It also fills all six slots the encoding rule needs, and the claw is a ` +
		`gesture the cat does not have.</p>`)

	// ---- what the coats actually become -------------------------------------
	b.WriteString(`<h2>1 &middot; What the coats actually become</h2>`)
	b.WriteString(`<p class="nt">The body is drawn as glyphs, so a coat goes through the glyph path: pushed 2.6x ` +
		`away from grey, then snapped to the 256 cube. <b>The colour named is not necessarily the colour painted.</b> ` +
		`Every coat below was read back off a shell cell in the rendered night frame, not computed. ` +
		`This matters more for a crab than for anything before it: every shipped coat is a near-neutral that the ` +
		`boost leaves alone, and a salmon pink is the first one with enough chroma to move.</p>`)
	idx := map[int][]string{}
	b.WriteString(`<table><tr><th>coat</th><th>chosen</th><th>painted</th><th>index</th><th></th></tr>`)
	for _, ct := range coats {
		c, px, py := crabFrame(&crabs[0], ct.col, 80, 24, 0.931, 3, *seed, false)
		bx, by := bodyCell(px, py)
		got, i := painted(c, bx, by)
		idx[i] = append(idx[i], ct.name)
		moved := ""
		if got != ct.col {
			moved = "moved"
		}
		fmt.Fprintf(&b, `<tr><td><b>%s</b></td>`+
			`<td><span class="sw" style="background:rgb(%d,%d,%d)"></span> %d,%d,%d</td>`+
			`<td><span class="sw" style="background:rgb(%d,%d,%d)"></span> %d,%d,%d</td>`+
			`<td>%d</td><td>%s %s</td></tr>`,
			ct.name, ct.col.R, ct.col.G, ct.col.B, ct.col.R, ct.col.G, ct.col.B,
			got.R, got.G, got.B, got.R, got.G, got.B, i, ct.note, moved)
	}
	b.WriteString(`</table>`)
	var dups []string
	for i, ns := range idx {
		if len(ns) > 1 {
			dups = append(dups, fmt.Sprintf("index %d is shared by %s", i, strings.Join(ns, " and ")))
		}
	}
	if len(dups) > 0 {
		fmt.Fprintf(&b, `<p class="nt warn"><b>Collision:</b> %s &mdash; those are one colour on screen, not two. `+
			`Pick between them by name and you are picking nothing.</p>`, strings.Join(dups, "; "))
	} else {
		b.WriteString(`<p class="nt">All six survive as distinct cube entries. Nothing collides.</p>`)
	}

	cell := func(label, sub, body, note string) {
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%s</i></div><div class="win">%s</div>`, label, sub, body)
		if note != "" {
			fmt.Fprintf(&b, `<p class="note">%s</p>`, note)
		}
		b.WriteString(`</div>`)
	}

	// ---- the silhouettes ----------------------------------------------------
	b.WriteString(`<h2>2 &middot; Four silhouettes, in salmon, at dusk</h2>`)
	b.WriteString(`<p class="nt">The shape decision. Each is the crab at three times size where the companion sits, ` +
		`with the shipped cat last for scale &mdash; same box, same margin, same row.</p><div class="row">`)
	for i := range crabs {
		cr := &crabs[i]
		c, px, py := crabFrame(cr, coats[0].col, 80, 24, 0.778, 3, *seed, false)
		cell(cr.name, "", c.HTMLFragmentCropAs(px-3, py-1, px+12+3, py+7+1, 33, term.Profile256), cr.note)
	}
	c, px, py := catFrame(80, 24, 0.778, 3, *seed)
	cell("the cat", "ships today", c.HTMLFragmentCropAs(px-3, py-1, px+12+3, py+7+1, 33, term.Profile256),
		"The reference. Tall, narrow, and it leaves the bottom corners of the box empty -- which is exactly the room a crab uses.")
	b.WriteString(`</div>`)

	// ---- the coats ----------------------------------------------------------
	b.WriteString(`<h2>3 &middot; Six coats on one silhouette, through the day</h2>`)
	b.WriteString(`<p class="nt">The colour decision, on Broad so the shape is held still. The lower beach falls away ` +
		`to black by design, so the companion always sits on dark sand &mdash; which is what a salmon has to beat.</p>`)
	for _, hr := range hours {
		fmt.Fprintf(&b, `<h2 style="border:0;margin:18px 0 6px">%s</h2><div class="row">`, hr.name)
		for _, ct := range coats {
			c, px, py := crabFrame(&crabs[0], ct.col, 80, 24, hr.tod, 3, *seed, false)
			bx, by := bodyCell(px, py)
			_, i := painted(c, bx, by)
			cell(ct.name, fmt.Sprintf("%d", i), c.HTMLFragmentCropAs(px-3, py-1, px+12+3, py+7+1, 33, term.Profile256), "")
		}
		b.WriteString(`</div>`)
	}

	// ---- in the whole frame -------------------------------------------------
	b.WriteString(`<h2>4 &middot; In the whole scene, with a litter</h2>`)
	b.WriteString(`<p class="nt">How he will actually meet it: the companion is a small thing in the corner of a ` +
		`wide scene, and the crop above flatters everything. Two young at the kittens' size and position, ` +
		`salmon, at night.</p>`)
	for i := range crabs {
		cr := &crabs[i]
		c, _, _ := crabFrame(cr, coats[0].col, 80, 24, 0.931, 3, *seed, true)
		b.WriteString(`<div class="row">`)
		cell(cr.name, "with litter", c.HTMLFragmentAs(11, term.Profile256), "")
		b.WriteString(`</div>`)
	}

	page := canvas.HTMLPage("xscapes — the crab", b.String())
	if err := os.WriteFile(*out, []byte(page), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(*out)
}
