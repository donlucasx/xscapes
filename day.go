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
	render := func(tod float64) (tc, smooth, raw, blocks string, distinct, worst int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = lay.MoonX
		for i := 0; i < 14; i++ {
			sh.Update(c, 3.0+float64(i)/20, scape.Activity{
				Working: true, Level: 0.55, TimeOfDay: tod, ContextUsed: 0.3})
		}
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: tail}, 3.7, seed, c.H-2-chh)

		// Counted at HALF-ROW resolution and over the SKY only. A cell on 256
		// can carry two colours, so anything reading c.BG alone is blind to
		// half of what is drawn; and the flat writing band would otherwise
		// dominate the longest-run figure with a flatness that is deliberate.
		// The horizon is the biggest downward step in luma, not the first step
		// over a threshold. A fixed threshold works by day and finds nothing at
		// midnight, where sky and sea are two neighbouring greys -- and "found
		// nothing" came out as a caption reading "0 distinct tones".
		sky, drop := 0, 0.0
		for y := 1; y < c.H*3/4; y++ {
			if d := luma(c.BGAt(2, y-1)) - luma(c.BGAt(2, y)); d > drop {
				drop, sky = d, y
			}
		}
		seen := map[term.RGB]bool{}
		run, prev := 0, term.RGB{R: 1, G: 2, B: 3}
		for y := 0; y < sky; y++ {
			ch, fg, bg := c.ResolveAt(2, y, term.Profile256)
			up, dn := bg, bg
			switch ch {
			case '\u2580':
				up = fg
			case '\u2591':
				up, dn = bg.Blend(fg, 0.25), bg.Blend(fg, 0.25)
			case '\u2592':
				up, dn = bg.Blend(fg, 0.5), bg.Blend(fg, 0.5)
			case '\u2593':
				up, dn = bg.Blend(fg, 0.75), bg.Blend(fg, 0.75)
			}
			for _, v := range []term.RGB{up, dn} {
				seen[v] = true
				if v == prev {
					run++
				} else {
					run = 1
				}
				if run > worst {
					worst = run
				}
				prev = v
			}
		}
		tc = c.HTMLFragmentAs(11, term.ProfileTrueColor)
		smooth = c.HTMLFragmentAs(11, term.Profile256)
		sh1, sh2 := term.Shading, term.ShadeBlocks
		term.ShadeBlocks = true
		blocks = c.HTMLFragmentAs(11, term.Profile256)
		term.ShadeBlocks = sh2
		term.Shading = false
		raw = c.HTMLFragmentAs(11, term.Profile256)
		term.Shading, term.ShadeBlocks = sh1, sh2
		return tc, smooth, raw, blocks, len(seen), worst
	}

	var b strings.Builder
	b.WriteString(`<style>
.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}
.pair{display:flex;gap:16px;align-items:flex-start;flex-wrap:wrap}
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
		`<b style="color:#c9c9d4">SMOOTHED</b> is the default now: on 256 a cell can carry ` +
		`two colours, so where a band edge falls inside a cell it is split with U+2580 and ` +
		`the ramp gets twice the vertical resolution. No extra texture, same colours, placed ` +
		`more finely. <code>ASCIISCAPES_SHADE=0</code> is the panel beside it.<br>` +
		`<b style="color:#c9c9d4">SHADE BLOCKS</b> is the idea that lost. It buys three more ` +
		`tones between every pair of cube colours &mdash; 11 to 14 in the sky &mdash; and it ` +
		`still looks worse, because the dot pattern reads as stipple before it reads as tone. ` +
		`Judge it in a real terminal before believing me: <code>ASCIISCAPES_SHADE_BLOCKS=1</code>.<br>` +
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
		tc, smooth, raw, blocks, distinct, worst := render(tod)
		on := ""
		if i == 12 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="moment%s" data-i="%d">`, on, i)
		fmt.Fprintf(&b, `<div class="lv"><b>%02d:00</b> &middot; tod %.3f &middot; the SKY carries %d `+
			`distinct tones at half-row resolution, longest flat run %d half-rows</div>`, i, tod, distinct, worst)
		b.WriteString(`<div class="pair">`)
		fmt.Fprintf(&b, `<div><div class="lv">truecolor &mdash; what the palette asks for</div>`+
			`<div class="win">%s</div></div>`, tc)
		fmt.Fprintf(&b, `<div><div class="lv">256 &mdash; smoothed</div>`+
			`<div class="win">%s</div></div>`, smooth)
		fmt.Fprintf(&b, `<div><div class="lv">256 &mdash; no smoothing</div>`+
			`<div class="win">%s</div></div>`, raw)
		fmt.Fprintf(&b, `<div><div class="lv">256 + shade blocks &mdash; rejected</div>`+
			`<div class="win">%s</div></div>`, blocks)

		b.WriteString(`</div>`)
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
		b.WriteString(`</div></div></div>`)
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
