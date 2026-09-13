package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// decisionsPage is every art decision still open, in one place, with the
// options rendered side by side. His ask of 2026-09-12: "why dont you make a
// new page with all the outstanding decisions on art so we can see and pick."
//
// It is a CONSOLIDATION, not another round. Everything here has already been
// drawn; what is missing is a pick.

func decisionsPage(seed int64) string {
	var b strings.Builder
	b.WriteString(decisionsCSS)
	b.WriteString(`<h1>xscapes &mdash; what is still open</h1>`)
	sha := "unknown"
	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	fmt.Fprintf(&b, `<div class="hdr">HEAD <b>%s</b> &middot; %s &middot; `+
		`<span class="mono">go run . -decisions &lt;path&gt;</span><br>`+
		`Four decisions. Everything else on the companion is ruled and built. `+
		`Every frame is the product's own draw.</div>`,
		sha, time.Now().Format("2006-01-02 15:04"))

	b.WriteString(decisionTopRung())
	b.WriteString(decisionEyeRepeat())
	b.WriteString(decisionMidStride())
	b.WriteString(decisionStarTouch(seed))

	b.WriteString(`<script>
var sl=document.getElementById('sl'),play=document.getElementById('play'),timer=null;
function show(i){
  document.querySelectorAll('.stage').forEach(function(st){
    var n=st.children.length; if(n<2) return;
    var per=2*n-2, k=i-Math.floor(i/per)*per, j=(k<n)?k:per-k;
    for(var q=0;q<n;q++) st.children[q].classList.toggle('on', q===j);
  });
}
sl.addEventListener('input',function(){show(+sl.value);});
play.addEventListener('click',function(){
  if(timer){clearInterval(timer);timer=null;play.innerHTML='&#9654; play';return;}
  play.innerHTML='&#10073;&#10073; pause';
  timer=setInterval(function(){var v=+sl.value+1; if(v>45){v=0;} sl.value=v; show(v);},83);});
show(0);
</script>`)
	return canvas.HTMLPage("xscapes - open decisions", b.String())
}

// ---------------------------------------------------------------------------

// decisionTopRung is the one he has already half-answered: "2a (the eyes) or
// round 4, the faithtful scaled up works."
func decisionTopRung() string {
	var b strings.Builder
	b.WriteString(`<h2>1 &mdash; the cat's top rung, 24&times;14</h2>`)
	b.WriteString(`<div class="ctl"><button class="btn" id="play">&#9654; play</button>` +
		`<input type="range" id="sl" min="0" max="45" value="0">` +
		`<div class="now">the slider breathes every sprite on the page</div></div>`)
	b.WriteString(`<p class="nt">The size the cat stops at when it walks up to ask. ` +
		`<b>A is what is built and running right now</b> &mdash; the shipped body doubled, muzzle ` +
		`filled, one eye glyph centred in each socket. B, C and D are hand-drawn at 48&times;56 ` +
		`source. You said A or B both work; this is them at the same size, side by side.</p>`)

	b.WriteString(`<div class="grid">`)
	// A: the product's own rung 2.
	b.WriteString(`<div class="cell"><div class="lv"><b>A</b> &mdash; the 2&times; body, fixed ` +
		`<span class="good">(built)</span></div>` + productRung2(companion.NeedsYou, 22) + `</div>`)
	labels := []string{"B", "C", "D"}
	for i, r := range catTopRungs {
		if i >= len(labels) {
			break
		}
		ask := append(append([]string{}, r.UpperAsk...), r.Lower...)
		fmt.Fprintf(&b, `<div class="cell"><div class="lv"><b>%s</b> &mdash; %s</div>%s</div>`,
			labels[i], r.Name, catRungStage(ask, r.EyeCells, r.EyeRow, 2, companion.NeedsYou, 22))
	}
	b.WriteString(`</div>`)
	b.WriteString(`<div class="st"><b>The difference that matters:</b> A is a 2&times; blow-up, so ` +
		`every edge is a 2&times;2 block and the ears are square. B, C and D have tapered ears, a ` +
		`jaw line and split paws, because someone drew them at this size. A costs nothing more; ` +
		`the others cost a bitmap each and are already written.</div>`)
	return b.String()
}

// productRung2 renders what the product actually draws at the top rung, by
// walking a real cat up its real ladder. Nothing here is composed by hand.
func productRung2(st companion.State, px int) string {
	var b strings.Builder
	b.WriteString(`<div class="stage">`)
	for i := 0; i < 24; i++ {
		cat := companion.NewCat()
		cat.SetCoat(companion.Coats["cream"])
		cat.FaceLeft(true)
		for k := 0; k < 40; k++ {
			cat.Approach(0.1, true)
		}
		const fw = 120
		_, bw, bh := cat.DrawnBox(fw)
		c := canvas.New(bw+34, bh+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat.Draw(c.Near(), bw+16, 1, float64(i)/12, st)
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s">%s</div>`, on,
			c.HTMLFragmentCropAs(6, 0, c.W, c.H, px, term.Profile256))
	}
	b.WriteString(`</div>`)
	return b.String()
}

// ---------------------------------------------------------------------------

func decisionEyeRepeat() string {
	var b strings.Builder
	b.WriteString(`<h2>2 &mdash; the rotating eye, when it repeats</h2>`)
	b.WriteString(`<p class="nt">Four eyes, picked at random per ask. Random is right &mdash; a ` +
		`random eye carries no information, so it cannot put a second meaning on a channel that ` +
		`already says "the agent needs you".<br>` +
		`<b>But four-way random repeats.</b> Measured over 4,000 asks: <b>24.3% draw the same eye ` +
		`as the one before.</b> That is exactly what random does, and over a working day it may ` +
		`read as the rotation being broken.</p>`)

	rows := []struct {
		name, desc string
		pick       func(int) int
	}{
		{"A &mdash; pure random <span class=\"good\">(built)</span>",
			"What is running now. Every ask is independent.", nil},
		{"B &mdash; never twice running",
			"Draw from the three that are NOT the previous eye. Still random, still carries no " +
				"information, never repeats back to back.", nil},
		{"C &mdash; a shuffled bag",
			"Shuffle all four, deal them out, reshuffle. Each face appears exactly once every " +
				"four asks, in a random order. The most even, and the least random.", nil},
	}
	b.WriteString(`<table><tr><th>rule</th><th>first 12 asks</th><th>repeats</th></tr>`)
	for i, r := range rows {
		var seq []int
		prev := -1
		bag := []int{}
		for a := 0; a < 4000; a++ {
			var p int
			switch i {
			case 0:
				p = hashPick(a, 4)
			case 1:
				p = hashPick(a, 3)
				if prev >= 0 && p >= prev {
					p++ // skip the previous, keeping the draw uniform over the other three
				}
			case 2:
				if len(bag) == 0 {
					bag = []int{0, 1, 2, 3}
					// A deterministic shuffle off the ask index.
					for k := len(bag) - 1; k > 0; k-- {
						j := hashPick(a*7+k, k+1)
						bag[k], bag[j] = bag[j], bag[k]
					}
				}
				p, bag = bag[0], bag[1:]
			}
			if a < 12 {
				seq = append(seq, p)
			}
			prev = p
			if a >= 12 && i != 2 {
				break // only the bag rule needs the long run to stay in phase
			}
		}
		// Count repeats over the whole run, from a clean start.
		prev, rep := -1, 0
		bag = nil
		for a := 0; a < 4000; a++ {
			var p int
			switch i {
			case 0:
				p = hashPick(a, 4)
			case 1:
				p = hashPick(a, 3)
				if prev >= 0 && p >= prev {
					p++
				}
			case 2:
				if len(bag) == 0 {
					bag = []int{0, 1, 2, 3}
					for k := len(bag) - 1; k > 0; k-- {
						j := hashPick(a*7+k, k+1)
						bag[k], bag[j] = bag[j], bag[k]
					}
				}
				p, bag = bag[0], bag[1:]
			}
			if a > 0 && p == prev {
				rep++
			}
			prev = p
		}
		var tiles strings.Builder
		for _, p := range seq {
			tiles.WriteString(crabEyeTile(nearEyeRotationArt(p), companion.EyeAlert, 15))
		}
		cls := ""
		if rep > 0 {
			cls = ` class="bad"`
		}
		fmt.Fprintf(&b, `<tr><td><b>%s</b><div class="st">%s</div></td>`+
			`<td><div class="row">%s</div></td><td%s>%.1f%%</td></tr>`,
			r.name, r.desc, tiles.String(), cls, 100*float64(rep)/3999)
	}
	b.WriteString(`</table>`)
	return b.String()
}

func hashPick(n, mod int) int {
	h := uint64(n) * 0x9E3779B97F4A7C15
	h ^= h >> 31
	h *= 0x94D049BB133111EB
	h ^= h >> 29
	return int(h % uint64(mod))
}

// ---------------------------------------------------------------------------

func decisionMidStride() string {
	var b strings.Builder
	b.WriteString(`<h2>3 &mdash; the cat mid-stride, which is a live defect</h2>`)
	b.WriteString(`<p class="nt">You asked where you can see this. <b>Here</b> &mdash; and in your ` +
		`own terminal, for a quarter of a second every time a tool event lands while the cat is ` +
		`the companion. Folded through the real reducer over your 126 spools: <b>4,037 times ` +
		`across 31 sessions</b>, about 65 times per hour of working time.<br>` +
		`The cat has no mid-stride body, so <span class="mono">Draw</span> reaches for ` +
		`<span class="mono">CatWalk</span> &mdash; a 16-cell SIDE view &mdash; and draws it inside ` +
		`the 12-cell box, then plots the front-view eye glyphs at their usual cells. The eyes land ` +
		`on nothing.</p>`)

	shot := func(stepping bool, px int) string {
		cat := companion.NewCat()
		cat.SetCoat(companion.Coats["cream"])
		cat.FaceLeft(true)
		cat.SetStepping(stepping)
		cw, ch := cat.Size()
		c := canvas.New(cw+6, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat.Draw(c.Near(), 3, 1, 1.3, companion.Working)
		return c.HTMLFragmentAs(px, term.Profile256)
	}
	b.WriteString(`<div class="grid">` +
		`<div class="cell"><div class="lv">standing &mdash; correct</div>` + shot(false, 24) + `</div>` +
		`<div class="cell"><div class="lv"><b class="bad">mid-stride &mdash; the defect</b></div>` +
		shot(true, 24) + `</div>` +
		`<div class="cell"><div class="lv">the crab mid-stride, for contrast</div>` +
		crabStrideShot(24) + `</div>` +
		`</div>`)
	b.WriteString(`<div class="st"><b>The decision is which fix.</b> ` +
		`<b>A:</b> author a front-view step pose in the 12-cell box &mdash; one bitmap, the way ` +
		`the crab has <span class="mono">crabLowerStep</span> under <span class="mono">crabLower</span>. ` +
		`The cat keeps a stride and it is drawn rather than borrowed. ` +
		`<b>B:</b> stop swapping the body at all &mdash; the cat simply does not change shape ` +
		`mid-step, like it does not today at a near rung. Costs no art and loses the stride. ` +
		`<br>I would author the bitmap (A): the crab has one, and a companion that never changes ` +
		`shape while walking is the thing you reported about the first cat &mdash; ` +
		`<i>"I dont see this one moving"</i>.</div>`)
	return b.String()
}

func crabStrideShot(px int) string {
	crab := companion.NewCrab()
	crab.FaceLeft(true)
	crab.SetStepping(true)
	cw, ch := crab.Size()
	c := canvas.New(cw+6, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	crab.Draw(c.Near(), 3, 1, 1.3, companion.Working)
	return c.HTMLFragmentAs(px, term.Profile256)
}

// ---------------------------------------------------------------------------

func decisionStarTouch(seed int64) string {
	var b strings.Builder
	b.WriteString(`<h2>4 &mdash; two stars touching at 40&times;12 &mdash; you ruled "dont let em touch"</h2>`)
	b.WriteString(`<p class="nt">The separation bar is <b>2.0 screen units</b>, and two stars one ` +
		`row apart in the same column measure <b>exactly 2.00</b> &mdash; so the bar has zero ` +
		`headroom and lets them sit adjacent. 8 pairs at 19 stars; 97 at 32.<br>` +
		`<b>The decision is how.</b> The band at 40&times;12 is three rows of about 36 usable ` +
		`columns, so there is not much room to give.</p>`)
	b.WriteString(`<div class="st"><b>A:</b> raise the bar above 2.00 &mdash; stars stop stacking ` +
		`vertically, and at 40&times;12 fewer of them fit, so the count goes wrong in a different ` +
		`way. <b>B:</b> keep the bar and forbid the specific case &mdash; same column, adjacent ` +
		`row. Cheapest and targeted. <b>C:</b> cap the constellation at small sizes: draw fewer ` +
		`stars rather than crowded ones, and say so.<br>` +
		`I would take B: it is the case you named, it costs one comparison, and it does not ` +
		`change how many stars a small window can hold.</div>`)
	return b.String()
}

func nearEyeRotationArt(i int) []string { return companion.NearEyeRotationArt(i) }

const decisionsCSS = `<style>
h2{font:600 14px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 8px;letter-spacing:.06em;text-transform:uppercase}
.hdr{font:11px ui-monospace,monospace;color:#8a8a99;background:#141419;border:1px solid #222229;
  border-radius:6px;padding:12px 14px;margin:0 0 20px;line-height:1.6}
.ctl{position:sticky;top:0;z-index:9;display:flex;gap:14px;align-items:center;margin:0 0 16px;
  padding:12px 14px;background:#16161c;border:1px solid #272730;border-radius:8px}
.ctl input[type=range]{width:220px;accent-color:#6f8fd0}
.btn{background:#22222b;color:#d8d8e0;border:1px solid #33333f;border-radius:5px;padding:5px 11px;
  font:12px ui-monospace,monospace;cursor:pointer}
.now{font:12px ui-monospace,monospace;color:#9a9aa8}
.stage .fr{display:none}.stage .fr.on{display:block}
.grid{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start}
.row{display:flex;gap:4px;flex-wrap:wrap;align-items:center}
.cell{border:1px solid #222229;border-radius:6px;padding:8px;background:#101015}
.lv{font:10px ui-monospace,monospace;color:#7d7d8a;margin-bottom:5px;letter-spacing:.04em}
.st{font:11px ui-monospace,monospace;color:#8a8a99;margin-top:8px;line-height:1.6;
  background:#141419;border:1px solid #222229;border-left:3px solid #6f8fd0;
  border-radius:6px;padding:12px 14px}
.bad{color:#ff8080}.good{color:#7fd18f}
.mono{font-family:ui-monospace,monospace}
table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#9a9aa8;margin:8px 0}
th,td{border:1px solid #26262e;padding:8px 10px;text-align:left;vertical-align:top}
th{color:#c8c8d4;font-weight:600}
</style>`

var _ = math.Abs
