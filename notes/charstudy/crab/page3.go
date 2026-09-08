package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The two balloon colours, copied from sheet.go: the ask is warm, the finish
// knock is cool. Same distinction the cat ships with.
var (
	bubbleCol    = term.RGB{R: 224, G: 228, B: 238}
	bubbleAskCol = term.RGB{R: 244, G: 198, B: 122}
	eyeAlert     = term.RGB{R: 232, G: 252, B: 226}
	eyeWorried   = term.RGB{R: 244, G: 176, B: 96} // amber -- his ruling, it stays
)

type shape struct {
	key, name string
	lower     []string
	crablet   []string
	eyes      [2]int
	eyeRow    int
}

var shapes = []shape{
	{"pincer", "Pincer", pincerLower, crabletPincer, [2]int{4, 7}, 1},
	{"hero", "Hero", heroLower, crabletHero, [2]int{4, 7}, 2},
}

type pose struct {
	name   string
	uppers [][]string
	eye    rune
	eyeCol term.RGB
	blink  bool
	say    string
	ask    bool
	note   string
}

func posesFor(k string) []pose {
	if k == "pincer" {
		return []pose{
			{"resting", [][]string{pincerRest, pincerRest, pincerWork2, pincerRest}, '-', eyeShine, false, "", false,
				"Claws down, pincers closed, eyes half shut. The whole body drifts one pixel and settles."},
			{"working", [][]string{pincerWork, pincerWork2}, 'o', eyeShine, true, "", false,
				"Claws up and breathing, eyes open with the occasional blink. The sea behind is doing the real talking."},
			{"needs you", [][]string{pincerAsk, pincerAsk, pincerWork2}, 'O', eyeAlert, false, "allow Bash?", true,
				"ONE claw raised, and raised on the side facing the transcript. This is the gesture the cat does not have -- and it is position, not rate, so it survives a screenshot."},
			{"done", [][]string{pincerDone, pincerWork}, '^', eyeShine, false, "tests passed", false,
				"Both claws up and held, pincers open wide, content eyes. Still rather than waving, so it cannot read as another ask."},
			{"worried", [][]string{pincerWorried, pincerWorried}, 'o', eyeWorried, false, "", false,
				"Claws pulled in and down, stalks short, hunched. Amber eyes -- his ruling that amber stays applies here too. This one persists until it clears."},
		}
	}
	return []pose{
		{"resting", [][]string{heroRest, heroRest, heroWork2, heroRest}, '-', eyeShine, false, "", false,
			"The arms fold down onto the shell. A hero crab at rest stops being a hero, which is most of why the state reads."},
		{"working", [][]string{heroWork, heroWork2}, 'o', eyeShine, true, "", false,
			"Claws high on their arms, breathing a pixel. The tallest silhouette of the two."},
		{"needs you", [][]string{heroAsk, heroAsk, heroWork2}, 'O', eyeAlert, false, "allow Bash?", true,
			"One arm stays up, the other drops all the way. The biggest shape change of any state on either crab."},
		{"done", [][]string{heroDone, heroWork}, '^', eyeShine, false, "tests passed", false,
			"Both arms up, pincers open. Held still."},
		{"worried", [][]string{heroWorried, heroWorried}, 'o', eyeWorried, false, "", false,
			"Arms collapse against the body and the stalks shorten. The largest silhouette becomes the smallest, which is the point."},
	}
}

func compose28(upper, lower []string) []string {
	out := make([]string, 0, 28)
	out = append(out, upper...)
	return append(out, lower...)
}

// drawPosed draws one crab in one pose, eyes included.
func drawPosed(near *canvas.Layer, sh *shape, upper []string, eye rune, eyeCol term.RGB, col term.RGB, x, y int) {
	bm := companion.ParseBitmap(compose28(upper, sh.lower)).Mirrored()
	q := bm.ToQuadrant()
	companion.PlotRim(near, q, x, y)
	(&companion.Sprite{Rows: q, Body: col}).Draw(near, x, y)
	for _, e := range sh.eyes {
		near.Plot(x+12-1-e, y+sh.eyeRow, eye, eyeCol, 1)
	}
}

func page3(seed int64, frames int, fps float64) string {
	col := coats[0].col // salmon 210
	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.row{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start;margin-bottom:8px}
	.col{display:flex;flex-direction:column;gap:4px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	.lbl b{color:#d8d8e0;font-weight:500}
	.lbl i{font-style:normal;color:#6a6a78}
	.note{font:12px/1.5 ui-sans-serif,system-ui;color:#8a8a99;max-width:36ch;margin:2px 0 0}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 6px;padding-top:14px;border-top:1px solid #23232a}
	h3{font:600 12px ui-monospace,monospace;color:#b8b8c4;margin:22px 0 6px}
	table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#b8b8c4;margin:8px 0 18px}
	td,th{border:1px solid #2a2a32;padding:3px 9px;text-align:left}
	th{color:#d8d8e0;font-weight:500}
	.stage{position:relative}.fr{display:none}.fr:first-child{display:block}
	.full{margin:0 0 26px}
	</style>`)
	b.WriteString(`<h1>xscapes &mdash; the crab, full study</h1>`)
	b.WriteString(`<p class="nt">Pincer and Hero, his two picks, taken the rest of the way: measured against the cat, ` +
		`given sub companions, and animated through all five states. Salmon (index 210) throughout. ` +
		`Nothing is installed and no product code has been touched.</p>`)

	// ---- 1. is it as big as the cat -----------------------------------------
	b.WriteString(`<h2>1 &middot; Is it as big as the cat?</h2>`)
	b.WriteString(`<p class="nt">Measured, not asserted. Every companion is drawn into the same 12 cell by 7 cell box ` +
		`&mdash; that part is fixed by the pipeline. The question is how much of the box each one actually fills.</p>`)
	b.WriteString(`<table><tr><th></th><th>ink, in cells</th><th>cells carrying ink</th><th>of the box</th></tr>`)
	for _, s := range []struct {
		n string
		r []string
	}{{"the cat", companion.CatBody}, {"Pincer", compose28(pincerWork, pincerLower)}, {"Hero", compose28(heroWork, heroLower)}} {
		w, h, f := extent(s.n, s.r)
		fmt.Fprintf(&b, `<tr><td><b>%s</b></td><td>%d x %d</td><td>%d of 84</td><td>%.0f%%</td></tr>`, s.n, w, h, f, float64(f)/84*100)
	}
	b.WriteString(`</table>`)
	b.WriteString(`<p class="nt"><b>Same height, three cells wider, about the same weight of ink.</b> The cat's ink is ` +
		`9 cells across and the crabs use all 12 &mdash; but the cat fills 60 of the 84 cells in its box and the crabs ` +
		`fill 64 and 66, so they are within a tenth of the same amount of drawn material. They are not bigger; ` +
		`they are the same size arranged sideways.</p>`)
	b.WriteString(`<p class="nt">And the three extra columns cost nothing: <code>compose()</code> already reserves the ` +
		`full 12 from <code>cat.Size()</code>, which returns the box and not the ink. The cat has simply never used ` +
		`its last three columns. A crab needs no layout change and takes no room from the sand.</p><div class="row">`)
	for _, s := range []struct {
		n     string
		up    []string
		lo    []string
		isCat bool
	}{{"the cat", nil, nil, true}, {"Pincer", pincerWork, pincerLower, false}, {"Hero", heroWork, heroLower, false}} {
		c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		act := scape.Activity{Working: true, Level: 0.5, TimeOfDay: 0.778, ContextUsed: 0.3}
		for k := 0; k < 12; k++ {
			sh.Update(c, 3+float64(k)/20, act)
		}
		px, py := c.W-12-(2+c.W/32), c.H-2-7
		if s.isCat {
			cat := companion.NewCat()
			cat.FaceLeft(true)
			cat.Draw(c.Near(), px, py, 3, companion.Working)
		} else {
			sp := shape{lower: s.lo, eyes: [2]int{4, 7}, eyeRow: 1}
			if s.n == "Hero" {
				sp.eyeRow = 2
			}
			drawPosed(c.Near(), &sp, s.up, 'o', eyeShine, col, px, py)
		}
		fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b></div><div class="win">%s</div></div>`,
			s.n, c.HTMLFragmentCropAs(px-1, py-1, px+12+1, py+7+1, 33, term.Profile256))
	}
	b.WriteString(`</div><p class="nt">Same crop, same rows, no margin either side &mdash; the box edges are the crop edges.</p>`)

	// ---- 2. sub companions ---------------------------------------------------
	b.WriteString(`<h2>2 &middot; The sub companions</h2>`)
	b.WriteString(`<p class="nt">One crablet per subagent, at the kittens' size and in the kittens' place: six cells by ` +
		`four, growing leftward along the sand. A sub-cell is one pixel wide by two tall, so at this size a ` +
		`one-pixel notch still survives the halving &mdash; which is why a crablet can keep an open pincer where ` +
		`the froglet could not keep its toes.</p>`)
	b.WriteString(`<p class="nt">A kitten stays at least 60 seconds after its subagent starts before it swims off, and a ` +
		`crablet inherits that: he twice reported seeing no subagents at all because they lived seven and sixteen ` +
		`seconds against a six second exit.</p><div class="row">`)
	for i := range shapes {
		s := &shapes[i]
		for _, n := range []int{1, 2, 3} {
			c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh := scape.NewShore(seed, false)
			sh.MoonX = 0.28
			act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: 0.778, ContextUsed: 0.3}
			for k := 0; k < 12; k++ {
				sh.Update(c, 3+float64(k)/20, act)
			}
			px, py := c.W-12-(2+c.W/32), c.H-2-7
			up := pincerWork
			if s.key == "hero" {
				up = heroWork
			}
			drawPosed(c.Near(), s, up, 'o', eyeShine, col, px, py)
			idx := make([]int, n)
			for k := range idx {
				idx[k] = k + 1
			}
			drawLitter(c.Near(), s, col, px, py, idx, 3, seed)
			fmt.Fprintf(&b, `<div class="col"><div class="lbl"><b>%s</b> <i>%d subagent(s)</i></div><div class="win">%s</div></div>`,
				s.name, n, c.HTMLFragmentCropAs(px-24, py-1, px+12+1, py+7+1, 20, term.Profile256))
		}
	}
	b.WriteString(`</div>`)

	// ---- 3. the states, animated --------------------------------------------
	b.WriteString(`<h2>3 &middot; All five states, animated</h2>`)
	fmt.Fprintf(&b, `<p class="nt">%d frames each at %.0f fps, the sea running underneath. Every state is told by `+
		`POSITION &mdash; where the claws are &mdash; and not by how fast anything moves, so each one still reads `+
		`in a screenshot. That is the encoding rule, and it is the reason the claw is worth having.</p>`, frames, fps)
	for i := range shapes {
		s := &shapes[i]
		fmt.Fprintf(&b, `<h3>%s</h3>`, s.name)
		for _, p := range posesFor(s.key) {
			var st strings.Builder
			for f := 0; f < frames; f++ {
				t := 3 + float64(f)/fps
				c := canvas.New(72, 22, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				shr := scape.NewShore(seed, false)
				shr.MoonX = 0.28
				act := scape.Activity{TimeOfDay: 0.778, ContextUsed: 0.3}
				if p.name != "resting" {
					act.Working, act.Level = true, 0.6
				}
				for k := 0; k < 10; k++ {
					shr.Update(c, t+float64(k)/40, act)
				}
				px, py := c.W-12-(2+c.W/32), c.H-2-7
				eye := p.eye
				if p.blink && f%frames == frames-2 {
					eye = '-'
				}
				drawPosed(c.Near(), s, p.uppers[f%len(p.uppers)], eye, p.eyeCol, col, px, py)
				if p.say != "" {
					rows, bc := companion.DoneBubble(p.say), bubbleCol
					if p.ask {
						rows, bc = companion.Bubble(p.say), bubbleAskCol
					}
					(&companion.Sprite{Rows: rows, Body: bc, Opaque: true}).Draw(c.Near(), px-2, py-len(rows))
				}
				fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(13, term.Profile256))
			}
			fmt.Fprintf(&b, `<div class="row"><div class="col"><div class="lbl"><b>%s</b></div>`+
				`<div class="stage">%s</div></div><p class="note">%s</p></div>`, p.name, st.String(), p.note)
		}
	}

	// ---- 4. the full beach, animated ----------------------------------------
	fmt.Fprintf(&b, `<h2>4 &middot; The full beach at %dx%d, working</h2>`, fullW, fullH)
	b.WriteString(`<p class="nt">His own window, the sea running, two subagents out. This is the size it is actually ` +
		`met at &mdash; the companion is a small thing in the corner and everything above flatters it.</p>`)
	for i := range shapes {
		s := &shapes[i]
		var st strings.Builder
		up := [][]string{pincerWork, pincerWork2}
		if s.key == "hero" {
			up = [][]string{heroWork, heroWork2}
		}
		for f := 0; f < 4; f++ {
			t := 3 + float64(f)/3
			c := canvas.New(fullW, fullH, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			shr := scape.NewShore(seed, false)
			shr.MoonX = 0.28
			act := scape.Activity{Working: true, Level: 0.6, TimeOfDay: 0.778, ContextUsed: 0.3, TodoDone: 2, TodoTotal: 5}
			for k := 0; k < 10; k++ {
				shr.Update(c, t+float64(k)/40, act)
			}
			px, py := c.W-12-(2+c.W/32), c.H-2-7
			drawPosed(c.Near(), s, up[f%len(up)], 'o', eyeShine, col, px, py)
			drawLitter(c.Near(), s, col, px, py, []int{1, 2}, t, seed)
			fmt.Fprintf(&st, `<div class="fr">%s</div>`, c.HTMLFragmentAs(7, term.Profile256))
		}
		fmt.Fprintf(&b, `<div class="full"><div class="lbl"><b>%s</b> <i>dusk, working, two subagents</i></div>`+
			`<div class="stage win">%s</div></div>`, s.name, st.String())
	}

	fmt.Fprintf(&b, `<script>
document.querySelectorAll('.stage').forEach(function(st){
  var fr = Array.prototype.slice.call(st.children), i = 0;
  setInterval(function(){
    fr[i].style.display = 'none';
    i = (i + 1) %% fr.length;
    fr[i].style.display = 'block';
  }, %d);
});
</script>`, int(1000/fps))
	return canvas.HTMLPage("xscapes — the crab, full study", b.String())
}
