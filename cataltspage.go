package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// catAltsPage is the design round. Alternatives rendered, he rules -- the same
// shape Hero was picked in and the coat was chosen in.

func catAltsPage(seed int64) string {
	var b strings.Builder
	b.WriteString(catAltsCSS)
	b.WriteString(`<h1>xscapes &mdash; the cat, drafted</h1>`)
	b.WriteString(catAltsHeader())

	b.WriteString(`<div class="ctl"><button class="btn" id="play">&#9654; play</button>` +
		`<input type="range" id="sl" min="0" max="45" value="0">` +
		`<div class="now" id="lbl">frame 0</div>` +
		`<div class="now">every sprite is breathing; the slider moves them together. ` +
		`The strip plays forward then back &mdash; the breath and the tail have ` +
		`incommensurable periods, so no finite loop closes seamlessly.</div></div>`)

	// ---- HIS PICKS ---------------------------------------------------------
	b.WriteString(`<h2>His picks, 2026-09-11</h2>`)
	b.WriteString(`<p class="nt">Ruled from the round below: <b>Pricked ears</b> for the ask, ` +
		`<b>Chin up</b> for the cat's finish, <b>Settled on the sand</b> for the crab's, and ` +
		`<b>the Faithful Scale-Up</b> for the middle rung &mdash; <i>"but without the nose ` +
		`(the original rung 0 doesnt have any nose/mouth)"</i>.</p>`)
	b.WriteString(catPicksHTML(seed))

	// ---- THE GAP ----------------------------------------------------------
	b.WriteString(`<h2>The gap, before any alternative</h2>`)
	b.WriteString(`<p class="nt">These two frames are the cat <b>working</b> and the cat ` +
		`<b>asking you for something</b>, as it ships today. The body is the same bitmap. ` +
		`At a matched breath phase <b>zero cells differ</b> &mdash; the whole ask is one eye ` +
		`glyph changing from <span class="mono">o</span> to <span class="mono">O</span>, and a ` +
		`tail held still at a position the working tail passes through on every cycle. ` +
		`A screenshot cannot tell them apart. The crab got a shape difference in session 28; ` +
		`the cat never did.</p>`)
	b.WriteString(`<div class="grid">`)
	for _, p := range []struct {
		st   companion.State
		name string
	}{{companion.Working, "working"}, {companion.NeedsYou, "asking &mdash; today"},
		{companion.Done, "done &mdash; today"}, {companion.Worried, "worried (has its own art)"}} {
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">%s</div>%s</div>`,
			p.name, catAnim(nil, p.st, 24, 20))
	}
	b.WriteString(`</div>`)
	b.WriteString(`<div class="st">Measured off the RENDERED frames through ` +
		`<span class="mono">Cat.Draw</span>. Not by comparing bitmaps, which only proves a thing ` +
		`equals itself; and not by diffing two frames at the same clock, which measures the two ` +
		`poses' breath and wag being out of PHASE and calls a rate difference a still-frame one. ` +
		`The question a screenshot asks is whether the picture is AMBIGUOUS:</div><table>` +
		`<tr><th>pair</th><th>share of its pictures that are also working pictures</th></tr>`)
	for _, pr := range []struct {
		a, b companion.State
		name string
	}{
		{companion.NeedsYou, companion.Working, "a still of the cat ASKING, against every picture it draws while working"},
		{companion.Done, companion.Working, "a still of the cat DONE, against every picture it draws while working"},
		{companion.Resting, companion.Working, "a still of the cat RESTING, against working"},
		{companion.Worried, companion.Working, "a still of the cat WORRIED, against working"},
	} {
		shared, total := catAmbiguity(nil, pr.a, pr.b, 1200)
		pct := 100 * float64(shared) / float64(max(1, total))
		cls := ""
		if pct > 50 {
			cls = ` class="bad"`
		}
		fmt.Fprintf(&b, `<tr><td>%s</td><td%s>%d of %d (%.0f%%)</td></tr>`,
			pr.name, cls, shared, total, pct)
	}
	b.WriteString(`</table><div class="st">1,200 frames per pose at 20&nbsp;fps, ` +
		`bodies only &mdash; the two eye cells are blanked, because two cells of colour out of ` +
		`eighty-four is not what a glance reads. <b>A high number means a screenshot of that ` +
		`pose is a picture the cat also draws while simply working</b>, so the still is ` +
		`ambiguous however different the two look in motion.</div>`)

	// ---- ROUND 1: THE ASK -------------------------------------------------
	b.WriteString(`<h2>Round 1 &mdash; the ask pose</h2>`)
	b.WriteString(`<p class="nt">Each candidate is drawn by the product's own <span class="mono">` +
		`Cat.Draw</span>, breathing, with its real tail, mirrored the way it sits on screen. ` +
		`Left is the candidate asking; beside it is the cat working, unchanged, so the ` +
		`difference is what you are judging. Then the same pose where it actually lives.</p>`)
	if len(catAskAlts) == 0 {
		b.WriteString(`<p class="nt"><b class="bad">No candidates compiled in.</b></p>`)
	}
	for _, a := range catAskAlts {
		k := checkCatAlt(a.Rows)
		var risks strings.Builder
		for _, r := range a.Risks {
			fmt.Fprintf(&risks, `<div class="risk1">&#9888; %s</div>`, r)
		}
		shared, total := catAmbiguity(a.Rows, companion.NeedsYou, companion.Working, 1200)
		pct := 100 * float64(shared) / float64(max(1, total))
		amb := fmt.Sprintf(`<div class="st"><b class="good">%d of %d (%.0f%%)</b> of this pose's `+
			`pictures are also pictures the cat draws while working &mdash; today that number `+
			`is 97%%</div>`, shared, total, pct)
		if pct > 0 {
			amb = fmt.Sprintf(`<div class="st"><b class="bad">%d of %d (%.0f%%)</b> of this `+
				`pose's pictures are ALSO pictures the cat draws while working &mdash; a `+
				`screenshot is still ambiguous that often</div>`, shared, total, pct)
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta">`+
			`<div class="nm">%s</div><div class="nt">%s<br><i>%s</i></div>%s%s%s</div>`+
			`<div class="grid">`+
			`<div class="cell"><div class="lv">asking</div>%s</div>`+
			`<div class="cell"><div class="lv">working (unchanged)</div>%s</div>`+
			`<div class="cell"><div class="lv">asking, in the scene, with the balloon</div>%s</div>`+
			`</div></div>`,
			a.Name, a.Idea, a.Note, amb, catCheckRow(k), risks.String(),
			catAnim(a.Rows, companion.NeedsYou, 24, 22),
			catAnim(nil, companion.Working, 24, 22),
			catInScene(a.Rows, companion.NeedsYou, 4, 13, 96, 24, seed, 0.78, "allow Bash?"))
	}

	// ---- ROUND 2: DONE ----------------------------------------------------
	b.WriteString(`<h2>Round 2 &mdash; the finish pose, both animals</h2>`)
	b.WriteString(`<p class="nt">Your ruling of today: both animals get one. The brief locks ` +
		`<b>done</b> and <b>needs_input</b> as distinct cues &mdash; "come back when you like" ` +
		`against "you are blocking me" &mdash; and today, at a matched breath phase, ` +
		`<b>done and working render the same body</b> for both.</p>`)
	if len(catDoneAlts) == 0 {
		b.WriteString(`<p class="nt"><b class="bad">No candidates compiled in.</b></p>`)
	}
	for _, a := range catDoneAlts {
		crab := a.Animal == "crab"
		row := catCheckRow(checkCatAlt(a.Rows))
		if crab {
			row = catCheckRow(checkAltAgainst(a.Rows, companion.CrabWorkingArt()))
		}
		cand, today := "", ""
		if crab {
			// The crab's eye cells are {4, 7} in its 12-cell sprite, and its
			// whole body is bitmap -- no procedural tail to add.
			row = fmt.Sprintf(`<div class="st">crab art &middot; %d rows &middot; `+
				`composed by the study, not by drawCrab &mdash; there is no Done slot in `+
				`crabUpper yet</div>`, len(a.Rows))
			cand = catRungStage(a.Rows, [2]int{4, 7}, 2, 0, companion.Done, 22)
			today = catCrabStage(companion.Working, 22)
		} else {
			cand = catAnim(a.Rows, companion.Done, 24, 22)
			today = catAnim(nil, companion.Working, 24, 22)
		}
		var risks strings.Builder
		for _, x := range a.Risks {
			fmt.Fprintf(&risks, `<div class="risk1">&#9888; %s</div>`, x)
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta">`+
			`<div class="nm">%s &middot; <span class="lv">%s</span></div>`+
			`<div class="nt">%s<br><i>%s</i></div>%s%s</div>`+
			`<div class="grid">`+
			`<div class="cell"><div class="lv">done &mdash; candidate</div>%s</div>`+
			`<div class="cell"><div class="lv">working &mdash; today</div>%s</div>`+
			`</div></div>`,
			a.Name, a.Animal, a.Idea, a.Note, row, risks.String(), cand, today)
	}

	// ---- ROUND 3: THE LADDER ----------------------------------------------
	b.WriteString(`<h2>Round 3 &mdash; coming closer</h2>`)
	b.WriteString(`<p class="nt">The crab walks up in rungs &mdash; <b>12&times;7 &rarr; 16&times;9 ` +
		`&rarr; 24&times;14</b>, one rung per 0.28&nbsp;s, the same beat as its stride. The cat ` +
		`cannot: <span class="mono">crab_near.go</span> returns zero rungs for it, because the ` +
		`near art is Hero's.<br>` +
		`<b>The top rung is free for the cat</b> &mdash; doubling its 24&times;28 body lands on ` +
		`exactly 24&times;14 cells. <b>The middle rung is the only drawing job</b>, and the ` +
		`question under this section is whether it is worth doing at all: three rungs read as ` +
		`an animal walking, two read as one big stride.<br>` +
		`&#9888; These rung frames are composed by the study, not by the product &mdash; there is ` +
		`no cat near path yet. The tail IS the product's, scaled, which is new today: ` +
		`<span class="mono">nearDouble</span> doubles bitmaps, so before this the doubled cat ` +
		`arrived with no tail at all.</p>`)
	b.WriteString(catLadderHTML())

	b.WriteString(catRung2FixSection())
	b.WriteString(catTopRungSection())
	b.WriteString(crabEyeSection(seed))
	b.WriteString(crabEyeRotationSection(seed))

	// ---- ROUND 4: THE RIM -------------------------------------------------
	b.WriteString(`<h2>Round 6 &mdash; the cleared ring, already ruled</h2>`)
	b.WriteString(`<p class="nt">Your ruling today: the cat's parent body gets the ring its own ` +
		`kittens have always had. This is what it buys &mdash; the cat drawn over its litter, ` +
		`with the ring and without.</p>`)
	b.WriteString(catRimHTML(seed))

	fmt.Fprint(&b, `<script>
var sl=document.getElementById('sl'),lbl=document.getElementById('lbl'),
    play=document.getElementById('play'),timer=null;
function show(i){
  // PING-PONG, and it is not a shortcut. The companion's breath and its tail
  // have INCOMMENSURABLE periods -- at the ask, breath 1.6 s against a wag of
  // sin(5t), whose ratio is exactly pi/4 -- so no finite strip of frames ever
  // closes seamlessly. Measured on the shipped draw: a straight 24-frame loop
  // put its single biggest jump of the whole animation, 34 cells, exactly at
  // the wrap. Walking the strip forward then backward is continuous at both
  // ends by construction, and both motions are sinusoidal, so a reversed
  // traversal is a thing the animal actually does.
  document.querySelectorAll('.stage').forEach(function(st){
    var n=st.children.length;
    if(n<2){ return; }
    var per=2*n-2, k=i-Math.floor(i/per)*per, j=(k<n)?k:per-k;
    for(var q=0;q<n;q++) st.children[q].classList.toggle('on', q===j);
  });
  lbl.textContent='frame '+i;
}
sl.addEventListener('input',function(){show(+sl.value);});
play.addEventListener('click',function(){
  if(timer){clearInterval(timer);timer=null;play.innerHTML='&#9654; play';return;}
  play.innerHTML='&#10073;&#10073; pause';
  timer=setInterval(function(){var v=+sl.value+1; if(v>45){v=0;} sl.value=v; show(v);},83);});
show(0);
</script>`)
	return canvas.HTMLPage("xscapes - the cat, drafted", b.String())
}

// catLadderHTML draws the walk-up for each candidate middle rung, and the
// two-rung alternative that needs no new art at all.
func catLadderHTML() string {
	var b strings.Builder
	for _, r := range catRungAlts {
		full := append(append([]string{}, r.Upper...), r.Lower...)
		ask := append(append([]string{}, r.UpperAsk...), r.Lower...)
		bad := ""
		if len(full) != 36 {
			bad = fmt.Sprintf(` <b class="bad">%d rows, want 36</b>`, len(full))
		}
		if len(r.Upper) > 0 && len([]rune(r.Upper[0])) != 32 {
			bad += fmt.Sprintf(` <b class="bad">%d columns, want 32</b>`, len([]rune(r.Upper[0])))
		}
		var risks strings.Builder
		for _, x := range r.Risks {
			fmt.Fprintf(&risks, `<div class="risk1">&#9888; %s</div>`, x)
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s<br><i>%s</i></div>`+
			`<div class="st">eye at cells %v, row %d%s</div>%s</div><div class="grid">`+
			`<div class="cell"><div class="lv">rung 0 &mdash; 12&times;7, home</div>%s</div>`+
			`<div class="cell"><div class="lv">rung 1 &mdash; 16&times;9, this candidate</div>%s</div>`+
			`<div class="cell"><div class="lv">rung 2 &mdash; 24&times;14, doubled, free</div>%s</div>`+
			`<div class="cell"><div class="lv">rung 1 asking</div>%s</div>`+
			`</div></div>`,
			r.Name, r.Idea, r.Note, r.EyeCells, r.EyeRow, bad, risks.String(),
			catAnim(nil, companion.Working, 24, 16),
			catRungStage(full, r.EyeCells, r.EyeRow, 32.0/24.0, companion.Working, 16),
			catRungStage(double2x(companion.CatBody), [2]int{8, 14}, 4, 2, companion.Working, 16),
			catRungStage(ask, r.EyeCells, r.EyeRow, 32.0/24.0, companion.NeedsYou, 16))
	}
	// The alternative that costs nothing: skip the middle rung entirely.
	b.WriteString(`<div class="card wide"><div class="meta">` +
		`<div class="nm">No middle rung &mdash; one stride</div>` +
		`<div class="nt">12&times;7 straight to 24&times;14. No new art at all: the top rung is ` +
		`the shipped body doubled. The cost is that it reads as a jump rather than as an animal ` +
		`walking, which is the distinction the crab's ladder was built on.<br>` +
		`<i>This is the cheap answer, and it is on the page so the expensive one has to earn it.</i>` +
		`</div></div><div class="grid">` +
		`<div class="cell"><div class="lv">rung 0 &mdash; 12&times;7</div>` + catAnim(nil, companion.Working, 24, 16) + `</div>` +
		`<div class="cell"><div class="lv">rung 1 &mdash; 24&times;14, one stride</div>` +
		catRungStage(double2x(companion.CatBody), [2]int{8, 14}, 4, 2, companion.Working, 16) + `</div>` +
		`</div></div>`)
	return b.String()
}

// catCrabStage draws the real crab through its own path, as the control the
// crab's Done candidates are judged against.
func catCrabStage(pose companion.State, px int) string {
	var b strings.Builder
	b.WriteString(`<div class="stage cstage">`)
	for i := 0; i < 24; i++ {
		on := ""
		if i == 0 {
			on = " on"
		}
		crab := companion.NewCrab()
		crab.FaceLeft(true)
		cw, ch := crab.Size()
		c := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		crab.Draw(c.Near(), 2, 1, float64(i)/12, pose)
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s</div>`, on, i,
			c.HTMLFragmentAs(px, term.Profile256))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func catRungStage(rows []string, eyeCells [2]int, eyeRow int, scale float64, pose companion.State, px int) string {
	var b strings.Builder
	b.WriteString(`<div class="stage cstage">`)
	for i := 0; i < 24; i++ {
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s</div>`, on, i,
			catRungSprite(rows, eyeCells, eyeRow, scale, pose, float64(i)/12, px, true))
	}
	b.WriteString(`</div>`)
	return b.String()
}

// catRimHTML shows what the cleared ring buys: the parent drawn beside its own
// litter, with and without.
func catRimHTML(seed int64) string {
	frame := func(rim bool) string {
		cat := companion.NewCat()
		cat.SetCoat(companion.Coats["cream"])
		cat.SetNoRim(!rim)
		st := catState(companion.Working, "", false)
		st.Kittens = 6
		c := catStage(96, 24, seed, 0.80, 0.45, true, cat, st, 4)
		ccw, chh := cat.Size()
		lay := compose(96, ccw, true)
		x0 := lay.CatX - 30
		if x0 < 0 {
			x0 = 0
		}
		y0 := c.H - 3 - chh
		if y0 < 0 {
			y0 = 0
		}
		return c.HTMLFragmentCropAs(x0, y0, c.W, c.H, 15, term.Profile256)
	}
	return `<div class="grid">` +
		`<div class="cell"><div class="lv">no ring &mdash; what shipped</div>` + frame(false) + `</div>` +
		`<div class="cell"><div class="lv">the ring &mdash; your ruling</div>` + frame(true) + `</div>` +
		`</div>`
}

const catAltsCSS = `<style>
h2{font:600 14px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 8px;letter-spacing:.06em;text-transform:uppercase}
.hdr{font:11px ui-monospace,monospace;color:#8a8a99;background:#141419;border:1px solid #222229;
  border-radius:6px;padding:12px 14px;margin:0 0 20px;line-height:1.6}
.card.wide{display:block}
.card.wide .meta{width:100%;margin-bottom:10px}
.ctl{position:sticky;top:0;z-index:9;display:flex;gap:14px;align-items:center;margin:0 0 16px;
  padding:12px 14px;background:#16161c;border:1px solid #272730;border-radius:8px}
.ctl input[type=range]{width:220px;accent-color:#6f8fd0}
.btn{background:#22222b;color:#d8d8e0;border:1px solid #33333f;border-radius:5px;padding:5px 11px;
  font:12px ui-monospace,monospace;cursor:pointer}
.now{font:12px ui-monospace,monospace;color:#9a9aa8}
.stage .fr{display:none}.stage .fr.on{display:block}
.grid{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start}
.cell{border:1px solid #222229;border-radius:6px;padding:8px;background:#101015}
.lv{font:10px ui-monospace,monospace;color:#7d7d8a;margin-bottom:5px;letter-spacing:.04em}
.st{font:10px ui-monospace,monospace;color:#6f6f7d;margin-top:6px}
.risk1{font:10px ui-monospace,monospace;color:#8a8a99;margin-top:4px}
.bad{color:#ff8080}.good{color:#7fd18f}
table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#9a9aa8;margin:8px 0 4px}
th,td{border:1px solid #26262e;padding:5px 9px;text-align:left}
th{color:#c8c8d4;font-weight:600}
.mono{font-family:ui-monospace,monospace}
</style>`

// catPicksHTML is what he chose, rendered, plus the two edits his rulings
// called for: the nose off the middle rung and an idle twitch in the crab's
// finish pose.
func catPicksHTML(seed int64) string {
	var b strings.Builder

	// The ask he picked, where it lives.
	pick := catAskAlts[0]
	for _, a := range catAskAlts {
		if strings.HasPrefix(a.Key, "pricked") {
			pick = a
		}
	}
	fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">The ask &mdash; %s</div>`+
		`<div class="nt">%s</div></div><div class="grid">`+
		`<div class="cell"><div class="lv">asking</div>%s</div>`+
		`<div class="cell"><div class="lv">working, for comparison</div>%s</div>`+
		`<div class="cell"><div class="lv">in the scene, with the balloon</div>%s</div>`+
		`</div></div>`, pick.Name, pick.Idea,
		catAnim(pick.Rows, companion.NeedsYou, 24, 22),
		catAnim(nil, companion.Working, 24, 22),
		catInScene(pick.Rows, companion.NeedsYou, 4, 13, 96, 24, seed, 0.78, "allow Bash?"))

	// The middle rung, nose removed.
	lvl := append(append([]string{}, catRungFaithfulUpper...), catRungFaithfulLower...)
	ask := append(append([]string{}, catRungFaithfulAsk...), catRungFaithfulLower...)
	b.WriteString(`<div class="card wide"><div class="meta">` +
		`<div class="nm">The middle rung &mdash; the Faithful Scale-Up, nose removed</div>` +
		`<div class="nt"><b>You were right, and the reason is a rendering accident rather than a ` +
		`design decision.</b> The shipped cat DOES carry a muzzle gap in its source &mdash; ` +
		`<span class="mono">CatBody</span> row 11, columns 8&ndash;10 &mdash; but ToQuadrant ORs ` +
		`source rows 2k and 2k+1 before packing, and rows 10 and 12 are solid, so row 11 is ` +
		`swallowed whole. <b>The cat has never had a nose on screen.</b> The scale-up put its own ` +
		`muzzle gap across rows 16&ndash;17, which is a half-row boundary, so it survived the OR ` +
		`and a nose appeared for the first time as the animal walked toward you &mdash; a new ` +
		`feature arriving with proximity, not more of the same detail at a bigger size.<br>` +
		`Removed: the notch on the level pose and the open mouth on the ask. The ask still reads, ` +
		`and it reads by the mechanism you picked at the shipped size &mdash; the skull band ` +
		`between the ears is deleted so the ears stand free.</div></div><div class="grid">` +
		`<div class="cell"><div class="lv">rung 0 &mdash; 12&times;7</div>` +
		catAnim(nil, companion.Working, 24, 16) + `</div>` +
		`<div class="cell"><div class="lv">rung 1 &mdash; 16&times;9, no nose</div>` +
		catRungStage(lvl, [2]int{2, 8}, 3, 32.0/24.0, companion.Working, 16) + `</div>` +
		`<div class="cell"><div class="lv">rung 1 asking &mdash; ears free</div>` +
		catRungStage(ask, [2]int{2, 8}, 3, 32.0/24.0, companion.NeedsYou, 16) + `</div>` +
		`<div class="cell"><div class="lv"><b class="bad">rung 2 &mdash; REJECTED</b> ` +
		`&middot; the naive 2&times; blow-up</div>` +
		catRungStage(double2x(companion.CatBody), [2]int{8, 14}, 4, 2, companion.NeedsYou, 16) + `</div>` +
		`</div><div class="st">&#9888; <i>"The middle rung 'rung 2' is no good."</i> It was the only ` +
		`rung nobody drew &mdash; <span class="mono">nearDouble</span> turns every source pixel ` +
		`into a 2&times;2 block. It renders at the right size and looks like nothing anyone ` +
		`authored. Redrawn in Round 4.</div></div>`)

	// The cat's finish.
	var chin catAlt
	for _, a := range catDoneAlts {
		if strings.Contains(a.Key, "chin") {
			chin = a
		}
	}
	if chin.Rows != nil {
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta">`+
			`<div class="nm">The cat's finish &mdash; %s</div><div class="nt">%s</div></div>`+
			`<div class="grid"><div class="cell"><div class="lv">done</div>%s</div>`+
			`<div class="cell"><div class="lv">working</div>%s</div></div></div>`,
			chin.Name, chin.Idea,
			catAnim(chin.Rows, companion.Done, 24, 22),
			catAnim(nil, companion.Working, 24, 22))
	}

	// The crab's finish, and the twitch he asked for.
	b.WriteString(`<div class="card wide"><div class="meta">` +
		`<div class="nm">The crab's finish &mdash; Settled on the sand, and the claw</div>` +
		`<div class="nt"><i>"it should not be completely static. it could close a claw every so ` +
		`often, or something similar."</i><br>` +
		`&#9888; Every other held pose in this project is deliberately STILL &mdash; ` +
		`<span class="mono">stillFor()</span> freezes the companion for NeedsYou, Done and ` +
		`Worried, because a held pose that drifts reads as the animal being unsure. This twitch ` +
		`is allowed for one reason worth stating: <b>a claw that shuts every so often carries no ` +
		`information.</b> Nothing varies with it, so it is not a rate encoding of anything. It ` +
		`says only that the animal is alive, which is what the finish pose was missing.<br>` +
		`The pincer is literally <span class="mono">##..##</span> in the source, a two-column ` +
		`gap, so closing it is filling that gap and nothing else moves. ` +
		`<b>One claw is a single CELL of change at 12&times;7</b> &mdash; possibly too quiet, ` +
		`which is the complaint that started the constellation work &mdash; so both are here.` +
		`</div></div><div class="grid">` +
		`<div class="cell"><div class="lv">settled &mdash; open</div>` +
		catRungStage(append(append([]string{}, crabSettledUpper...), crabSettledLower...),
			[2]int{4, 7}, 2, 0, companion.Done, 26) + `</div>` +
		`<div class="cell"><div class="lv">near claw shut &mdash; 1 cell</div>` +
		catRungStage(append(append([]string{}, crabSettledClaw...), crabSettledLower...),
			[2]int{4, 7}, 2, 0, companion.Done, 26) + `</div>` +
		`<div class="cell"><div class="lv">both claws shut &mdash; 2 cells</div>` +
		catRungStage(append(append([]string{}, crabSettledBoth...), crabSettledLower...),
			[2]int{4, 7}, 2, 0, companion.Done, 26) + `</div>` +
		`<div class="cell"><div class="lv">the twitch, at 1 close every 7 s</div>` +
		crabClawIdle(7.0, 0.45, 26) + `</div>` +
		`</div></div>`)
	return b.String()
}

// crabClawIdle animates the finish pose with the claw shutting on a slow cycle,
// so the cadence can be looked at rather than argued about. period is seconds
// between closes, hold is how long the claw stays shut.
func crabClawIdle(period, hold float64, px int) string {
	open := append(append([]string{}, crabSettledUpper...), crabSettledLower...)
	shut := append(append([]string{}, crabSettledClaw...), crabSettledLower...)
	var b strings.Builder
	b.WriteString(`<div class="stage cstage">`)
	// 24 frames at 12 fps is 2 s, so the slider alone cannot show a 7 s cycle.
	// Sample the whole period across those 24 slots instead and say so.
	for i := 0; i < 24; i++ {
		t := period * float64(i) / 24
		rows := open
		if t < hold {
			rows = shut
		}
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s</div>`, on, i,
			catRungSprite(rows, [2]int{4, 7}, 2, 0, companion.Done, t, px, true))
	}
	b.WriteString(`</div>`)
	return b.String()
}

// catTopRungSection is the redraw of the size the animal actually stops at.
func catTopRungSection() string {
	var b strings.Builder
	b.WriteString(`<h2>Round 4 &mdash; the top rung, redrawn</h2>`)
	b.WriteString(`<p class="nt"><b>24&times;14 is the size the cat stops at when it walks up to ` +
		`ask you something</b>, so it is the one you look at while deciding. It was the only rung ` +
		`nobody had drawn &mdash; the shipped body doubled, every pixel a 2&times;2 block.<br>` +
		`All three drafts below are hand-authored at 48&times;56 source, carry the tail at its own ` +
		`scale, and use the same ask mechanism you picked at the shipped size: the band between ` +
		`the ears is deleted so they stand free. <b>None of them has a nose or a mouth.</b></p>`)

	// The control: what he rejected.
	fmt.Fprintf(&b, `<div class="card wide"><div class="meta">`+
		`<div class="nm"><b class="bad">Rejected</b> &mdash; the naive 2&times; blow-up</div>`+
		`<div class="nt">For comparison, at the same size. Every feature is a 2&times;2 block, so `+
		`the ears are square, the jaw has no taper and the paws are three rectangles.</div></div>`+
		`<div class="grid"><div class="cell">%s</div></div></div>`,
		catRungStage(double2x(companion.CatBody), [2]int{8, 14}, 4, 2, companion.Working, 20))

	for _, r := range catTopRungs {
		full := append(append([]string{}, r.Upper...), r.Lower...)
		ask := append(append([]string{}, r.UpperAsk...), r.Lower...)
		var risks strings.Builder
		for _, x := range r.Risks {
			fmt.Fprintf(&risks, `<div class="risk1">&#9888; %s</div>`, x)
		}
		// The nose check, run rather than trusted: a centred gap on the jaw
		// rows is what the shipped cat does NOT have.
		nose := catHasNose(full)
		noseS := `<b class="good">no nose, no mouth</b>`
		if nose {
			noseS = `<b class="bad">A CENTRED GAP APPEARS ON THE JAW &mdash; look</b>`
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div><div class="st">eye at cells %v, row %d &middot; %s</div>%s</div>`+
			`<div class="grid">`+
			`<div class="cell"><div class="lv">24&times;14 &mdash; level</div>%s</div>`+
			`<div class="cell"><div class="lv">24&times;14 &mdash; asking</div>%s</div>`+
			`<div class="cell"><div class="lv">the ladder: 12&times;7</div>%s</div>`+
			`<div class="cell"><div class="lv">16&times;9</div>%s</div>`+
			`</div></div>`,
			r.Name, r.Idea, r.EyeCells, r.EyeRow, noseS, risks.String(),
			catRungStage(full, r.EyeCells, r.EyeRow, 2, companion.Working, 20),
			catRungStage(ask, r.EyeCells, r.EyeRow, 2, companion.NeedsYou, 20),
			catAnim(nil, companion.Working, 24, 20),
			catRungStage(append(append([]string{}, catRungFaithfulUpper...), catRungFaithfulLower...),
				[2]int{2, 8}, 3, 32.0/24.0, companion.Working, 20))
	}
	return b.String()
}

// catHasNose looks for the thing his ruling forbids: a gap on the CENTRE LINE
// of the jaw that survives the row OR.
//
// Checked rather than trusted, because the shipped cat carries a muzzle gap in
// its source that never reaches the screen -- rows 10 and 12 are solid and the
// OR swallows row 11 -- so "there is a gap in the art" and "there is a nose on
// screen" are different questions, and only the rendered frame answers the
// second one.
func catHasNose(rows []string) bool {
	q := companion.ParseBitmap(rows).ToQuadrant()
	if len(q) == 0 {
		return false
	}
	w := len([]rune(q[0]))
	// The jaw is the band under the eyes and above the waist: cell rows 5..7
	// of 14. A hole on the centre two columns there is a muzzle.
	for y := 5; y <= 7 && y < len(q); y++ {
		r := []rune(q[y])
		for _, x := range []int{w/2 - 1, w / 2} {
			if x >= 0 && x < len(r) && r[x] == ' ' {
				return true
			}
		}
	}
	return false
}

// catRung2FixSection is what he actually asked for: the doubled rung 2 kept,
// with its two defects fixed.
func catRung2FixSection() string {
	var b strings.Builder
	b.WriteString(`<h2>Rung 2, fixed</h2>`)
	b.WriteString(`<p class="nt"><i>"lets fix the rung 2 then, just fix the eyes and remove the ` +
		`mouth"</i> &mdash; so the 2&times; body stays and the two defects go.</p>`)

	b.WriteString(`<div class="card wide"><div class="meta">` +
		`<div class="nm">1. The mouth</div>` +
		`<div class="nt"><b>You caught the same trap from the other side.</b> Two days ago you had ` +
		`me take a nose OFF the 16&times;9 because <i>"the original rung 0 doesnt have any ` +
		`nose/mouth"</i>. It doesn't &mdash; but it does carry a muzzle gap in its art, at ` +
		`<span class="mono">CatBody</span> row 11, which never reaches the screen: the renderer ORs ` +
		`source rows in pairs, row 10 is solid, and the gap is swallowed. <b>Doubling puts that ` +
		`same gap on a clean cell boundary, so the mouth appears at 24&times;14</b> and nowhere ` +
		`else in the product.<br>` +
		`Filling row 11 is <b>provably invisible at 12&times;7</b> &mdash; the OR already hid it &mdash; ` +
		`so it is one source, not two, and the suite asserts it.</div></div><div class="grid">` +
		`<div class="cell"><div class="lv">with the mouth &mdash; what you screenshotted</div>` +
		catRungStage(double2x(companion.CatBody), [2]int{8, 14}, 4, 2, companion.NeedsYou, 20) + `</div>` +
		`<div class="cell"><div class="lv">mouth removed</div>` +
		catRungStage(double2x(catBodyNoMouth), [2]int{8, 14}, 4, 2, companion.NeedsYou, 20) + `</div>` +
		`</div></div>`)

	b.WriteString(`<div class="card wide"><div class="meta">` +
		`<div class="nm">2. The eyes &mdash; and most of what you were looking at was my bug</div>` +
		`<div class="nt">The mockup passed <span class="mono">{8, 14}</span> as the eye cells: ` +
		`<b>the CRAB's</b>, copied out of its near pose. Measured on the rendered frame, the cat's ` +
		`sockets are cells <b>4&ndash;6</b> and <b>12&ndash;14</b> &mdash; and mirrored, which is how ` +
		`the scape draws it, {8,14} become 15 and 9, <b>both of them ink</b>. So both eye glyphs ` +
		`were painted on the animal's forehead rather than in its sockets.<br>` +
		`That is the position. There is a second half: at 12&times;7 the eye is 1 cell of 84, and at ` +
		`24&times;14 a single glyph is 1 of 336 &mdash; which is why it reads as a small mark stuck ` +
		`on a large blank face. Three treatments, all with the eyes in the right place:</div></div>` +
		`<div class="grid">` +
		`<div class="cell"><div class="lv">a &mdash; one glyph, centred in the socket</div>` +
		catRung2Stage("glyph", 20) + `</div>` +
		`<div class="cell"><div class="lv">b &mdash; the socket edged, dark pupil</div>` +
		catRung2Stage("outline", 20) + `</div>` +
		`<div class="cell"><div class="lv">c &mdash; a two-cell eye in a deepened socket</div>` +
		catRung2Stage("tall", 20) + `</div>` +
		`</div><div class="st">&#9888; <b>a</b> is the faithful one: it is exactly what 12&times;7 ` +
		`does, just in the right cell. <b>b</b> keeps your "eyes are holes" ruling most literally ` +
		`&mdash; the pupil is sky, the socket is edged. <b>c</b> is the only one that makes the eye ` +
		`proportional to the face, and it is the only one that changes the ART (the socket is ` +
		`carried down one cell row), so it touches nothing at 12&times;7 or 16&times;9.</div></div>`)
	return b.String()
}

func catRung2Stage(treatment string, px int) string {
	var b strings.Builder
	b.WriteString(`<div class="stage cstage">`)
	for i := 0; i < 24; i++ {
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s</div>`, on, i,
			catRung2Shot(treatment, companion.NeedsYou, float64(i)/12, px, true))
	}
	b.WriteString(`</div>`)
	return b.String()
}
