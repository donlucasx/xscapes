package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// dayPage is the day cycle, every hour, in both colour profiles at once.
//
// It used to render true RGB only, which made it the study most likely to
// mislead: the sea and the sky spent most of a working day as flat grey on
// Terminal.app and this page showed them blue throughout. Both profiles are
// rendered from the SAME frame now, so the gap between what the palette asks
// for and what the target terminal can hold is the thing the page is about.
//
// The wave phase is identical at every hour on purpose. The scene is otherwise
// in motion, and two frames that differ in both colour and water cannot be
// compared for colour.
func dayPage(seed int64) string {
	const hours = 24
	const w, h = 68, 22

	work := []string{
		"read   internal/auth/handler.go  142 lines",
		"edit   internal/auth/handler.go  +18 -2",
		"shell  go test ./...  4.1s",
		"grep   rate.Limiter  3 files",
	}
	tail := make([]reduce.Line, 0, len(work))
	for i, s := range work {
		tail = append(tail, reduce.Line{Text: s, Age: 1 - float64(i+1)/float64(len(work))})
	}

	cat := companion.NewCat()
	cat.FaceLeft(true)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, true)

	// render paints one hour and hands back both profiles plus the measurement,
	// from a single frame so the two panels cannot drift apart.
	render := func(tod float64) (tc, c256 string, distinct, worst int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = lay.MoonX
		for i := 0; i < 14; i++ {
			sh.Update(c, 3.0+float64(i)/20, scape.Activity{
				Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3})
		}
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: tail}, 3.7, seed, c.H-2-chh)

		seen := map[int]bool{}
		run, prev := 0, -1
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
		// HTMLFragmentAs restores the canvas, so the same frame renders twice.
		return c.HTMLFragmentAs(11, term.ProfileTrueColor),
			c.HTMLFragmentAs(11, term.Profile256), len(seen), worst
	}

	var b strings.Builder
	b.WriteString(`<style>
.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}
.pair{display:flex;gap:16px;align-items:flex-start}
.lv{font-size:10px;color:#55555f;letter-spacing:.11em;margin-bottom:5px;text-transform:uppercase}
.lv b{color:#c9c9d4;font-weight:600}
.moment{margin-bottom:20px}
#reel.single .moment{display:none}
#reel.single .moment.on{display:block}
#reel.all .moment{display:block}
.bar2{display:flex;gap:5px;margin-top:9px;flex-wrap:wrap}
.ch{font-size:9px;color:#8a8a94;text-align:center;width:78px}
.sw{height:15px;border-radius:3px;border:1px solid #2a2a32}
.grey{color:#e0a0a0}
.ctl{display:flex;gap:14px;align-items:center;margin:0 0 20px;padding:14px 16px;
     background:#141419;border:1px solid #222229;border-radius:8px;position:sticky;top:0;z-index:5}
.ctl input[type=range]{flex:1;accent-color:#6f8fd0}
.btn{background:#22222b;border:1px solid #33333e;color:#c9c9d4;border-radius:5px;
     padding:6px 12px;font:inherit;font-size:12px;cursor:pointer}
.btn:hover{background:#2c2c36}
.now{font-size:19px;color:#f0f0f4;font-weight:600;width:74px;font-variant-numeric:tabular-nums}
.cap{font-size:12px;color:#7d7d8a;line-height:1.65;max-width:1000px;margin-bottom:20px}
.hd{font-size:11px;color:#9a9aa6;margin:34px 0 10px;letter-spacing:.06em}
</style>`)
	b.WriteString(`<h1>xscapes &mdash; the day cycle, truecolor vs 256</h1>`)
	b.WriteString(`<div class="cap">The same frame, rendered twice. <b style="color:#c9c9d4">TRUECOLOR</b> ` +
		`is what the palette asks for. <b style="color:#c9c9d4">256</b> is what Terminal.app can hold, ` +
		`through the real quantiser &mdash; that is the one you get.<br>` +
		`Wave phase is fixed across all 24 hours so the only thing changing is colour. ` +
		`Drag the slider, press play, or show every hour at once.</div>`)

	b.WriteString(`<div class="ctl">` +
		`<div class="now" id="lbl">00:00</div>` +
		`<button class="btn" id="play">&#9654; play</button>` +
		`<input type="range" id="sl" min="0" max="` + fmt.Sprint(hours-1) + `" value="12">` +
		`<button class="btn" id="mode">show every hour</button>` +
		`</div>`)

	b.WriteString(`<div id="reel" class="single">`)
	for i := 0; i < hours; i++ {
		tod := float64(i) / float64(hours)
		tc, c256, distinct, worst := render(tod)
		on := ""
		if i == 12 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="moment%s" data-i="%d">`, on, i)
		fmt.Fprintf(&b, `<div class="lv"><b>%02d:00</b> &middot; tod %.3f &middot; %d distinct background `+
			`colours down one column on 256, longest flat run %d rows</div>`, i, tod, distinct, worst)
		b.WriteString(`<div class="pair">`)
		fmt.Fprintf(&b, `<div><div class="lv">truecolor &mdash; what the palette asks for</div>`+
			`<div class="win">%s</div></div>`, tc)
		fmt.Fprintf(&b, `<div><div class="lv">256 &mdash; Terminal.app</div>`+
			`<div class="win">%s</div></div>`, c256)

		p := scape.PaletteAt(tod)
		b.WriteString(`<div><div class="lv">backgrounds: asked &rarr; shown</div><div class="bar2">`)
		for _, f := range []struct {
			n string
			c term.RGB
		}{{"SkyTop", p.SkyTop}, {"SkyHorizon", p.SkyHorizon}, {"SeaFar", p.SeaFar},
			{"SeaNear", p.SeaNear}, {"Moon", p.Moon}, {"Foam", p.Foam}} {
			shown := term.FromIndex256(f.c.Index256())
			kind, cls := "cube", ""
			if f.c.Index256() >= 232 {
				kind, cls = "GREY RAMP", ` class="grey"`
			}
			fmt.Fprintf(&b, `<div class="ch"><div class="sw" style="background:rgb(%d,%d,%d)"></div>`+
				`<div class="sw" style="background:rgb(%d,%d,%d)"></div>%s<br><span%s>%d %s</span></div>`,
				f.c.R, f.c.G, f.c.B, shown.R, shown.G, shown.B, f.n, cls, f.c.Index256(), kind)
		}
		b.WriteString(`</div></div></div></div>`)
	}
	b.WriteString(`</div>`)

	// The palette this scape may choose from, so a colour call is checkable.
	b.WriteString(`<div class="hd">EVERY CUBE COLOUR THAT SURVIVES AS A BLUE ` +
		`(blue highest, above red &mdash; these are what the sky and sea may be built from)</div>`)
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
	b.WriteString(`<div class="bar2">`)
	for _, e := range blues {
		fmt.Fprintf(&b, `<div class="ch"><div class="sw" style="background:rgb(%d,%d,%d)"></div>%d<br>luma %.0f</div>`,
			e.c.R, e.c.G, e.c.B, e.i, e.l)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<script>
var reel=document.getElementById('reel'),sl=document.getElementById('sl'),
    lbl=document.getElementById('lbl'),play=document.getElementById('play'),
    mode=document.getElementById('mode'),ms=reel.querySelectorAll('.moment'),timer=null;
function show(i){for(var k=0;k<ms.length;k++)ms[k].classList.toggle('on',k==i);
  lbl.textContent=(i<10?'0':'')+i+':00';}
sl.addEventListener('input',function(){show(+sl.value);});
play.addEventListener('click',function(){
  if(timer){clearInterval(timer);timer=null;play.innerHTML='&#9654; play';return;}
  play.innerHTML='&#10073;&#10073; pause';
  timer=setInterval(function(){sl.value=(+sl.value+1)%ms.length;show(+sl.value);},550);});
mode.addEventListener('click',function(){
  var all=reel.classList.toggle('all');reel.classList.toggle('single',!all);
  mode.textContent=all?'show one hour':'show every hour';});
show(12);
</script>`)
	return canvas.HTMLPage("xscapes - day cycle, truecolor vs 256", b.String())
}
