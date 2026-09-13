package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// starfallPage is the mockup his ruling gets judged on. Every glyph on it came
// out of a real composited canvas; there is no hand-drawn ASCII anywhere, and
// nothing here changes what the product does.

// armCfg is one candidate arrival, as it will be rendered.
type armCfg struct {
	Key   string
	Name  string
	Note  string
	Kind  string // none | flare | fall | fly
	Cfg   trackCfg
	Dur   float64
	Curve float64
	Tail  []rune
	Flare bool // land with a flare as well as a fall
}

// starStat is what one rendered frame actually says, read back off it.
type starStat struct {
	Stars int
	// Want is the count this frame SHOULD carry: the settled sky one star
	// short, plus the one that is arriving. Comparing against Done instead
	// charges the fall for a shortfall that predates it -- 40x12 draws 21 of
	// 32 stars whatever anybody does.
	Want    int
	Done    int
	Head    starPt
	HasHead bool
	MinSep  float64
	Eaten   int
	Rung    int
	FullA   int
	Exhaust int
	Worst   float64
	BGShift float64
}

const starfallFPS = 12.0

// starfallArm renders one arm end to end: settled sky, the arrival, settled sky
// again. pre and post are frames of stillness either side, because an event
// that is only ever shown mid-event cannot be judged as an event.
func starfallArm(base starScene, a armCfg, home starPt, lit []starPt,
	track []starPt, pre, post, px int) ([]string, []starStat) {

	var html []string
	var stats []starStat

	// The baseline, measured once: the settled sky one star short.
	settled := base
	settled.Done, settled.Marks = base.Done-1, nil
	sc0, _ := starfallRender(settled)
	n0 := len(starCellsIn(sc0))

	emit := func(sc starScene, want int) {
		c, reps := starfallRender(sc)
		ps := starCellsIn(c)
		st := starStat{Stars: len(ps), Want: want, Done: sc.Done,
			MinSep: starMinSep(ps), Worst: math.MaxFloat64}
		for _, r := range reps {
			if r.Step == 0 {
				st.Head, st.HasHead = starPt{r.X, r.Y}, true
			}
			if r.Eaten {
				st.Eaten++
			}
			if r.Rung > st.Rung {
				st.Rung = r.Rung
			}
			if r.FullAlpha {
				st.FullA++
			}
			if r.Exhausted {
				st.Exhaust++
			}
			if r.Contrast < st.Worst {
				st.Worst = r.Contrast
			}
			if r.BGShift > st.BGShift {
				st.BGShift = r.BGShift
			}
		}
		if st.Worst == math.MaxFloat64 {
			st.Worst = 0
		}
		rows := skyRowsOf(sc.H) + 3
		if rows > sc.H {
			rows = sc.H
		}
		html = append(html, c.HTMLFragmentCropAs(0, 0, sc.W, rows, px, term.Profile256))
		stats = append(stats, st)
	}

	// Before: the sky one star short, standing still.
	for i := 0; i < pre; i++ {
		s := base
		s.Done, s.Marks = base.Done-1, nil
		s.T = base.T + float64(i)/starfallFPS
		emit(s, n0)
	}

	n := int(math.Round(a.Dur * starfallFPS))
	if n < 1 {
		n = 1
	}
	for i := 0; i <= n; i++ {
		p := float64(i) / starfallFPS / a.Dur
		if p > 1 {
			p = 1
		}
		s := base
		s.T = base.T + float64(pre+i)/starfallFPS
		switch a.Kind {
		case "none":
			// What ships today: the star is simply there.
			s.Done, s.Marks = base.Done, nil
		case "flare":
			s.Done = base.Done - 1
			s.Marks = []starMark{{X: home.X, Y: home.Y, R: '*', Lift: 1 - starfallEase(p, 1)}}
		case "fall":
			s.Done = base.Done - 1
			s.Marks = starfallMarks(track, p, a.Curve, a.Tail)
			if a.Flare && p >= 1 {
				s.Marks[0].Lift = 1
			}
		case "fly":
			// The star is home and still from the first frame; a separate
			// mark flies into it. Nothing in the sky is displaced.
			s.Done = base.Done
			if p < 1 {
				i := starfallHead(track, p, a.Curve)
				s.Marks = []starMark{{X: track[i].X, Y: track[i].Y, R: '─'}}
			}
		}
		emit(s, n0+1)
	}

	for i := 0; i < post; i++ {
		s := base
		s.Done, s.Marks = base.Done, nil
		s.T = base.T + float64(pre+n+1+i)/starfallFPS
		emit(s, n0+1)
	}
	return html, stats
}

func starStatLine(st starStat) string {
	cls := ""
	if st.Stars != st.Want {
		cls = " bad"
	}
	head := "&mdash;"
	if st.HasHead {
		head = fmt.Sprintf("(%d,%d)", st.Head.X, st.Head.Y)
	}
	sep := fmt.Sprintf("%.2f", st.MinSep)
	if st.MinSep < 0 {
		sep = "n/a"
	} else if st.MinSep < 2 {
		sep = `<b class="bad">` + sep + `</b>`
	}
	extra := ""
	if st.Eaten > 0 {
		extra += fmt.Sprintf(` <b class="bad">%d EATEN</b>`, st.Eaten)
	}
	if st.Exhaust > 0 {
		extra += fmt.Sprintf(` <b class="bad">%d exhausted</b>`, st.Exhaust)
	}
	return fmt.Sprintf(`<div class="st%s">%d of %d &middot; head %s &middot; sep %s%s</div>`,
		cls, st.Stars, st.Want, head, sep, extra)
}

// starfallStrip lays an arm out as a stage: one frame visible at a time,
// driven by the page's own slider, with the numbers under it.
func starfallStrip(id string, html []string, stats []starStat) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="stage" id="%s">`, id)
	for i, h := range html {
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s%s</div>`, on, i, h, starStatLine(stats[i]))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func starfallHead2(sc starScene) (starPt, []starPt, *scape.Shore, map[starPt]bool, bool) {
	home, lit, ok := starfallArrival(sc)
	_, sh, _, _, _ := starfallStage(sc)
	return home, lit, sh, starfallForbidden(sc), ok
}

func starfallPage(seed int64) string {
	// Mirror TRUE because that is what ships: -mirror defaults true, the
	// companion sits on the right and the moon moves to 0.28. The first
	// version of this page left the field at its zero value and rendered the
	// OLD left-anchored layout under a header that said "companion right".
	base := starScene{W: 125, H: 28, Seed: seed, Tod: 0.75, Ctx: 0.56,
		Done: 12, Total: 32, Mirror: true, T: 4}
	base.Mag = scape.StarMagnitudeAt(base.Done-1, seed)

	home, lit, sh, forb, ok := starfallHead2(base)
	if !ok {
		return canvas.HTMLPage("xscapes - the star arriving",
			"<h1>no single fresh star at this scene; nothing to draw</h1>")
	}

	var b strings.Builder
	b.WriteString(starfallCSS)
	b.WriteString(`<h1>xscapes &mdash; the star arriving</h1>`)
	b.WriteString(starfallHeader(seed, base))

	b.WriteString(`<h2>The sweep, before any picture</h2>`)
	b.WriteString(`<p class="nt">Every geometry against four hours, 32 slots each, three phases per ` +
		`arrival, counted off the composed frame. <b>A measurement at one seed, one geometry or one ` +
		`hour is not a measurement</b> &mdash; that is what made the last two studies confidently ` +
		`wrong. The wider population (every geometry &times; every hour, then one axis at a time for ` +
		`seed, context, mirror and balloon) prints to stderr when this page is generated.</p>`)
	b.WriteString(starfallSweepHTML(seed))

	// ---- A. THE RULING ROW -------------------------------------------------
	arms := []armCfg{
		{Key: "a1", Name: "Nothing &mdash; what ships today", Kind: "none",
			Note: "The star appears on its own cell.", Dur: 0.50},
		{Key: "a2", Name: "Flare only", Kind: "flare", Dur: 0.45,
			Note: "No travel. The cell steps up a cube path from white and decays to the star's own tone."},
		{Key: "a3", Name: "The fall (proposed)", Kind: "fall", Dur: 0.50, Curve: 1,
			Cfg:  trackCfg{N: 2, Rmax: 6, DyMax: 3, TrackMin: 4, Sep: 2},
			Note: "6 columns, 3 rows, away from the moon, linear, no tail. The '*' that moves IS the star."},
		{Key: "a4", Name: "A4 as drafted", Kind: "fall", Dur: 0.80, Curve: 2,
			Cfg:  trackCfg{N: 4, Rmax: 8, DyMax: 2, TrackMin: 4, Sep: 2},
			Note: "8 columns, 2 rows, ease-out p=2.0. What the session 28 study proposed."},
		{Key: "a5", Name: "A mark flies in", Kind: "fly", Dur: 0.50, Curve: 1,
			Cfg:  trackCfg{N: 2, Rmax: 6, DyMax: 3, TrackMin: 4, Sep: 2},
			Note: "The star lights at home on frame 0 and never moves; a streak flies into it."},
	}

	b.WriteString(`<h2>A &mdash; the ruling row</h2>`)
	b.WriteString(`<p class="nt">Star #11 at (110,7) arriving, ` +
		`<b>Done 11 &rarr; 12</b>. One slider moves all five arms to the same frame. ` +
		`A1 is today. A3 is the proposal. A4 is the draft his ruling selected. ` +
		`A5 is the only version that moves no star at all &mdash; it breaks no existing ` +
		`guarantee, and it is not literally what he asked for. ` +
		`<b>Q1 is A1 against the rest. Q2 is A3 against A5.</b></p>`)

	const pre, post = 6, 8
	nFrames := 0
	b.WriteString(`<div class="ctl"><button class="btn" id="play">&#9654; play</button>` +
		`<input type="range" id="sl" min="0" max="0" value="0" disabled>` +
		`<div class="now" id="lbl">frame 0</div></div>`)

	for _, a := range arms {
		track, why := starfallTrack(base, home, lit, forb, sh, a.Cfg)
		html, stats := starfallArm(base, a, home, lit, track, pre, post, 16)
		if len(html) > nFrames {
			nFrames = len(html)
		}
		note := a.Note
		if a.Kind == "fall" || a.Kind == "fly" {
			sceneA, glassA := starAngles(a.Cfg.N)
			offs, verdict := starfallOffsets(track, a.Dur, starfallFPS, a.Curve)
			note += fmt.Sprintf(`<br><span class="mono">track %d cells, stopped by %s &middot; `+
				`%.0f&deg; scene / %.0f&deg; on his glass &middot; %.2fs &middot; `+
				`offsets %v &middot; %s</span>`,
				len(track), why, sceneA, glassA, a.Dur, offs, verdict)
		}
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div></div>%s</div>`,
			a.Name, note, starfallStrip("arm-"+a.Key, html, stats))
	}

	// ---- B. THE COUNT FILMSTRIP -------------------------------------------
	b.WriteString(`<h2>B &mdash; the count, frame by frame</h2>`)
	b.WriteString(`<p class="nt"><i>"once they appear, they should not dissapear IMO"</i> ` +
		`&mdash; 2026-09-10, his spelling. Counted off the rendered frame with ResolveAt, ` +
		`after drawScene, not asserted. The session 28 study's "no count error" line was ` +
		`one star, one seed, one geometry.</p>`)
	for _, sel := range []struct {
		arm  armCfg
		done int
		note string
	}{
		{arms[2], 12, "A3, star #11 at (110,7) &mdash; the proposal"},
		{arms[3], 12, "A4, star #11 at (110,7) &mdash; as drafted"},
		{arms[3], 4, "A4, star #3 at (47,1) &mdash; one row under the top of the canvas"},
	} {
		s := base
		s.Done = sel.done
		s.Mag = scape.StarMagnitudeAt(sel.done-1, seed)
		h2, l2, sh2, f2, ok2 := starfallHead2(s)
		if !ok2 {
			continue
		}
		tr, _ := starfallTrack(s, h2, l2, f2, sh2, sel.arm.Cfg)
		html, stats := starfallArm(s, sel.arm, h2, l2, tr, 1, 1, 13)
		var row strings.Builder
		row.WriteString(`<div class="film">`)
		for i := range html {
			fmt.Fprintf(&row, `<div class="cel">%s%s</div>`, html[i], starStatLine(stats[i]))
		}
		row.WriteString(`</div>`)
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div></div>%s</div>`,
			sel.note, row.String())
	}

	// ---- C. THE ANGLE ------------------------------------------------------
	b.WriteString(`<h2>C &mdash; the angle</h2>`)
	b.WriteString(`<p class="nt">A4's 8-across-2-down is <b>28&deg; on his glass</b> &mdash; a ` +
		`sideways slide. Sideways motion is the only motion he has ever remarked on ` +
		`unprompted: <i>"it still moves to the sides"</i>. 2:1 is 47&deg;, and falls. ` +
		`The scene's own aspect constant is 2.0 while his Terminal.app cell measures ` +
		`14.000 &times; 30.0 px = 2.143, so every angle quoted in cells is about 7% ` +
		`shallower than it will look.</p><div class="grid">`)
	for _, n := range []int{1, 2, 3, 4, 6, 8} {
		cfg := trackCfg{N: n, Rmax: 8, DyMax: 4, TrackMin: 2, Sep: 2}
		tr, _ := starfallTrack(base, home, lit, forb, sh, cfg)
		s := base
		s.Done = base.Done - 1
		s.Marks = starfallMarks(tr, 0.45, 1, nil)
		c, _ := starfallRender(s)
		sceneA, glassA := starAngles(n)
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">%d across per row &middot; `+
			`<b>%.0f&deg; scene</b> &middot; <b>%.0f&deg; glass</b> &middot; track %d</div>%s</div>`,
			n, sceneA, glassA, len(tr),
			c.HTMLFragmentCropAs(max(0, home.X-16), 0, min(base.W, home.X+6), skyRowsOf(base.H)+2, 18, term.Profile256))
	}
	b.WriteString(`</div>`)

	// ---- D. STEPPING -------------------------------------------------------
	b.WriteString(`<h2>D &mdash; the curve, ruled on the numbers</h2>`)
	b.WriteString(`<p class="nt">The criterion is: <b>never skip a cell; repeats are allowed.</b> ` +
		`A skipped cell is a teleport. 12&nbsp;fps is what he actually runs inside Claude Code ` +
		`(inside.go); 20&nbsp;fps is the standalone default.</p><table><tr><th>curve</th>` +
		`<th>offsets at 12 fps</th><th>verdict</th><th>offsets at 20 fps</th><th>verdict</th></tr>`)
	for _, cv := range []struct {
		k    float64
		name string
	}{{1, "linear"}, {1.5, "ease-out p=1.5"}, {2, "ease-out p=2.0 (A4)"}, {3, "ease-out p=3.0"}} {
		for _, a := range []armCfg{arms[2], arms[3]} {
			tr, _ := starfallTrack(base, home, lit, forb, sh, a.Cfg)
			o12, v12 := starfallOffsets(tr, a.Dur, 12, cv.k)
			o20, v20 := starfallOffsets(tr, a.Dur, 20, cv.k)
			fmt.Fprintf(&b, `<tr><td>%s &middot; %s</td><td class="mono">%v</td><td>%s</td>`+
				`<td class="mono">%v</td><td>%s</td></tr>`,
				cv.name, a.Key, o12, starVerdict(v12), o20, starVerdict(v20))
		}
	}
	b.WriteString(`</table>`)

	// ---- E. THE TAIL -------------------------------------------------------
	b.WriteString(`<h2>E &mdash; the tail, off by default</h2>`)
	b.WriteString(`<p class="nt">The sky's entire ink vocabulary is ten runes: the constellation's ` +
		`<span class="mono">*</span> and the ambient dust's nine &mdash; ` +
		`<span class="mono">. &middot; &deg; , + ` + "`" + ` &quot; ' :</span>. ` +
		`Three of those nine are exactly what a streak wants, which is why the session 28 study ` +
		`ranked A4 second <i>on the condition that it ships without a trail</i>. ` +
		`Anything else has to be a rune the scene never draws.<br>` +
		`<b>The unsettled part:</b> a run of <span class="mono">&#9586;</span> joins into one line ` +
		`only on a path that drops one row per column. At 2:1 it does not join &mdash; it reads as ` +
		`steep ticks crossing a shallow track. <span class="mono">&#9472;</span> joins along the ` +
		`horizontal runs and breaks at each row drop. Neither joins a 2:1 path completely. ` +
		`No measurement can settle that; only his window can.</p><div class="grid">`)
	for _, tl := range []struct {
		runes []rune
		name  string
	}{
		{nil, "none"},
		{[]rune{'─'}, "&#9472; U+2500 &times;1"},
		{[]rune{'─', '─'}, "&#9472; &times;2"},
		{[]rune{'╲', '╲'}, "&#9586; U+2572 &times;2"},
		{[]rune{'┄', '┄'}, "&#9476; U+2504 &times;2"},
	} {
		tr, _ := starfallTrack(base, home, lit, forb, sh, arms[2].Cfg)
		s := base
		s.Done = base.Done - 1
		s.Marks = starfallMarks(tr, 0.55, 1, tl.runes)
		c, reps := starfallRender(s)
		eaten := 0
		for _, r := range reps {
			if r.Eaten {
				eaten++
			}
		}
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">%s &middot; %d marks &middot; %d eaten</div>%s</div>`,
			tl.name, len(reps), eaten,
			c.HTMLFragmentCropAs(max(0, home.X-14), 0, min(base.W, home.X+5), skyRowsOf(base.H)+2, 18, term.Profile256))
	}
	b.WriteString(`</div>`)

	// ---- F. THE SEAM TEST --------------------------------------------------
	b.WriteString(`<h2>F &mdash; the seam, which nobody has looked at</h2>`)
	b.WriteString(`<p class="nt"><span class="mono">&#9472;</span> overhangs its advance by ` +
		`0.227&nbsp;px a side; <span class="mono">&#9586;</span> by 1.24&nbsp;px. Terminal.app clips ` +
		`at the cell boundary, so <span class="mono">&#9472;</span> is the likelier of the two to ` +
		`show a seam. <b>Paste the line below into your own window and look at the joins.</b> ` +
		`This project does not predict a render.</p>`)
	for _, px := range []int{11, 14, 18} {
		fmt.Fprintf(&b, `<div class="seam" style="font-size:%dpx">`+
			`&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&nbsp;&nbsp;`+
			`&#9586;&#9586;&#9586;&#9586;&#9586;&#9586;&#9586;&#9586;&nbsp;&nbsp;`+
			`&#9585;&#9585;&#9585;&#9585;&#9585;&#9585;&#9585;&#9585;</div>`, px)
	}
	b.WriteString(`<pre class="paste">────────  ` +
		`╲╲╲╲╲╲╲╲  ` +
		`╱╱╱╱╱╱╱╱</pre>`)

	// ---- G. THE GUARDS, SHOWN FAILING -------------------------------------
	b.WriteString(`<h2>G &mdash; the guards, shown failing</h2>`)
	b.WriteString(`<p class="nt"><b>A guard shown only succeeding is a guarantee claimed, not held.</b> ` +
		`Left is the guard on. Right is the same star with the guards off, at the phase where the ` +
		`unguarded walk is measurably worst &mdash; found by sweeping, not chosen for looks.<br>` +
		`The forbidden ground is <b>measured, not mirrored</b>: the readout's own guard is ` +
		`unexported and inside package scape, and nothing in there can see the ask balloon at all, ` +
		`because drawScene runs after Shore.Update and the balloon is an OPAQUE sprite into the ` +
		`same near layer. So it is found by rendering the composed frame across every context step ` +
		`and both balloons and recording every sky cell something else claimed. ` +
		`<b>That footprint is 1.5% of the sky at 125&times;28, 2.6% at 80&times;24 &mdash; and ` +
		`47.0% at 40&times;12.</b></p>`)
	b.WriteString(starfallGuards(base, seed))

	// ---- H. WHERE THE SKY REFUSES -----------------------------------------
	b.WriteString(`<h2>H &mdash; where the sky refuses</h2>`)
	b.WriteString(`<p class="nt">Some stars have no room. He is being asked whether that ` +
		`inconsistency reads as restraint or as a bug.</p><div class="grid">`)
	for _, done := range []int{1, 4, 16, 11} {
		s := base
		s.Done = done
		s.Mag = scape.StarMagnitudeAt(done-1, seed)
		h2, l2, sh2, f2, ok2 := starfallHead2(s)
		if !ok2 {
			continue
		}
		tr, why := starfallTrack(s, h2, l2, f2, sh2, arms[2].Cfg)
		s.Done = done - 1
		s.Marks = starfallMarks(tr, 0.5, 1, nil)
		c, _ := starfallRender(s)
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">star #%d at (%d,%d) &middot; `+
			`track %d &middot; stopped by %s</div>%s</div>`,
			done-1, h2.X, h2.Y, len(tr), why,
			c.HTMLFragmentCropAs(0, 0, base.W, skyRowsOf(base.H)+2, 12, term.Profile256))
	}
	b.WriteString(`</div>`)

	// ---- I. THE HOURS ------------------------------------------------------
	b.WriteString(`<h2>I &mdash; the hours, including the one nobody sampled</h2>`)
	b.WriteString(`<p class="nt"><b>14:00 is the binding hour</b> and no prior sweep contained it. ` +
		`His ruling: <i>"make sure we can still see some during daytime, even if fainter than at ` +
		`nighttime."</i> The rung and alpha printed under each frame are what the ink search ` +
		`actually SPENT &mdash; a contrast of 56 is the search's exit condition, not a margin.</p>` +
		`<div class="grid">`)
	for _, hr := range []float64{0, 6, 12, 14, 18, 21} {
		s := base
		s.Tod = hr / 24
		h2, l2, sh2, f2, ok2 := starfallHead2(s)
		if !ok2 {
			continue
		}
		tr, _ := starfallTrack(s, h2, l2, f2, sh2, arms[2].Cfg)
		s.Done = base.Done - 1
		s.Marks = starfallMarks(tr, 0.5, 1, nil)
		c, reps := starfallRender(s)
		rung, full, exh, worst := 0, 0, 0, math.MaxFloat64
		for _, r := range reps {
			if r.Rung > rung {
				rung = r.Rung
			}
			if r.FullAlpha {
				full++
			}
			if r.Exhausted {
				exh++
			}
			if r.Contrast < worst {
				worst = r.Contrast
			}
		}
		warn := ""
		if exh > 0 {
			warn = ` <b class="bad">EXHAUSTED</b>`
		}
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">%02.0f:00 &middot; worst rung %d &middot; `+
			`%d at full alpha &middot; worst contrast %.1f%s</div>%s</div>`,
			hr, rung, full, worst, warn,
			c.HTMLFragmentCropAs(max(0, h2.X-16), 0, min(base.W, h2.X+6), skyRowsOf(base.H)+2, 16, term.Profile256))
	}
	b.WriteString(`</div>`)

	// ---- J. GEOMETRIES -----------------------------------------------------
	b.WriteString(`<h2>J &mdash; the shapes a real window comes in</h2>`)
	b.WriteString(`<p class="nt">At 40&times;12 the band is three rows and most arrivals cannot ` +
		`move at all. "No fall below this size" is a decision to state, not to discover.</p>`)
	for _, g := range []struct{ w, h int }{{143, 27}, {125, 28}, {80, 24}, {40, 12}} {
		s := base
		s.W, s.H = g.w, g.h
		h2, l2, sh2, f2, ok2 := starfallHead2(s)
		if !ok2 {
			fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%d&times;%d</div>`+
				`<div class="nt">no single fresh star at this size</div></div></div>`, g.w, g.h)
			continue
		}
		tr, why := starfallTrack(s, h2, l2, f2, sh2, arms[2].Cfg)
		s.Done = base.Done - 1
		s.Marks = starfallMarks(tr, 0.5, 1, nil)
		c, _ := starfallRender(s)
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%d&times;%d</div>`+
			`<div class="nt">home (%d,%d) &middot; track %d &middot; stopped by %s &middot; `+
			`%d forbidden sky cells</div></div><div class="win">%s</div></div>`,
			g.w, g.h, h2.X, h2.Y, len(tr), why, len(f2), c.HTMLFragmentAs(11, term.Profile256))
	}

	// ---- K. CADENCE --------------------------------------------------------
	b.WriteString(`<h2>K &mdash; how often he would actually see it</h2>`)
	b.WriteString(starfallCadenceHTML(0.50))

	// ---- L. THE STILL-FRAME AUDIT -----------------------------------------
	b.WriteString(`<h2>L &mdash; the still-frame audit</h2>`)
	b.WriteString(`<p class="nt"><b>He reviews by screenshot.</b> A3 and A4 mid-flight show the ` +
		`right count with one star off its home cell &mdash; the cost, and no cue at all. ` +
		`A5 shows the cue and no cost. <b>This is the frame that decides Q2.</b></p>`)
	lbase := base
	lbase.Done, lbase.Marks = base.Done-1, nil
	lc0, _ := starfallRender(lbase)
	lWant := len(starCellsIn(lc0)) + 1
	for _, a := range []armCfg{arms[2], arms[3], arms[4]} {
		tr, _ := starfallTrack(base, home, lit, forb, sh, a.Cfg)
		var row strings.Builder
		row.WriteString(`<div class="film">`)
		for _, ph := range []struct {
			p    float64
			name string
		}{{0, "launch"}, {0.5, "mid-flight"}, {1, "landed"}} {
			s := base
			if a.Kind == "fly" {
				s.Done = base.Done
				if ph.p < 1 {
					i := starfallHead(tr, ph.p, a.Curve)
					s.Marks = []starMark{{X: tr[i].X, Y: tr[i].Y, R: '─'}}
				}
			} else {
				s.Done = base.Done - 1
				s.Marks = starfallMarks(tr, ph.p, a.Curve, a.Tail)
			}
			c, _ := starfallRender(s)
			ps := starCellsIn(c)
			cls := ""
			if len(ps) != lWant {
				cls = " bad"
			}
			fmt.Fprintf(&row, `<div class="cel"><div class="lv">%s</div>%s`+
				`<div class="st%s">%d of %d stars</div></div>`,
				ph.name, c.HTMLFragmentCropAs(0, 0, base.W, skyRowsOf(base.H)+2, 11, term.Profile256),
				cls, len(ps), lWant)
		}
		row.WriteString(`</div>`)
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div></div>%s</div>`, a.Name, a.Note, row.String())
	}

	// ---- M. THE CONTROL PAIR ----------------------------------------------
	b.WriteString(`<h2>M &mdash; the control pair</h2>`)
	b.WriteString(`<p class="nt">The last frame of the fall, beside the frame the product renders ` +
		`today at the same count. <b>They must be indistinguishable</b> &mdash; the feature leaves ` +
		`no residue once the star lands.</p>`)
	{
		tr, _ := starfallTrack(base, home, lit, forb, sh, arms[2].Cfg)
		land := base
		land.Done = base.Done - 1
		land.Marks = starfallMarks(tr, 1, 1, nil)
		lc, _ := starfallRender(land)
		today := base
		today.Marks = nil
		tc, _ := starfallRender(today)
		same := skySnapshotEq(lc, tc, skyRowsOf(base.H)+2)
		verdict := `<b class="bad">THEY DIFFER</b>`
		if same {
			verdict = `<b class="good">byte-identical across the whole sky</b>`
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">landed vs today</div>`+
			`<div class="nt">%s</div></div><div class="film">`+
			`<div class="cel"><div class="lv">the fall, landed</div>%s</div>`+
			`<div class="cel"><div class="lv">the product today</div>%s</div></div></div>`,
			verdict,
			lc.HTMLFragmentCropAs(0, 0, base.W, skyRowsOf(base.H)+2, 12, term.Profile256),
			tc.HTMLFragmentCropAs(0, 0, base.W, skyRowsOf(base.H)+2, 12, term.Profile256))
	}

	b.WriteString(starfallRisks)

	fmt.Fprintf(&b, `<script>
var N=%d, sl=document.getElementById('sl'), lbl=document.getElementById('lbl'),
    play=document.getElementById('play'), timer=null;
sl.max=N-1; sl.disabled=false;
function show(i){
  document.querySelectorAll('.stage').forEach(function(st){
    var n=st.children.length, k=Math.min(i,n-1);
    for(var j=0;j<n;j++) st.children[j].classList.toggle('on', j===k);
  });
  lbl.textContent='frame '+i;
}
sl.addEventListener('input',function(){show(+sl.value);});
play.addEventListener('click',function(){
  if(timer){clearInterval(timer);timer=null;play.innerHTML='&#9654; play';return;}
  play.innerHTML='&#10073;&#10073; pause';
  timer=setInterval(function(){sl.value=(+sl.value+1)%%N;show(+sl.value);},83);});
show(0);
</script>`, nFrames)

	return canvas.HTMLPage("xscapes - the star arriving", b.String())
}

func starVerdict(v string) string {
	if strings.HasPrefix(v, "SKIPS") {
		return `<b class="bad">` + v + `</b>`
	}
	return v
}

func skySnapshotEq(a, b *canvas.Canvas, rows int) bool {
	sa, sb := skySnapshot(a, rows), skySnapshot(b, rows)
	if len(sa) != len(sb) {
		return false
	}
	for k, v := range sa {
		if sb[k] != v {
			return false
		}
	}
	return true
}

// starfallGuards renders each guard ON beside the same star with the guards
// OFF, and says what the guards-off frame actually did wrong -- read off the
// frame, at the phase where it is worst.
//
// The cases are not hand-picked for looks. They were found by sweeping every
// star at three geometries, three contexts and both balloon states for the
// arrivals where the guarded and unguarded walks DIFFER, and keeping the ones
// where the unguarded walk measurably breaks something. A guard shown only
// succeeding is a guarantee claimed, not held.
func starfallGuards(base starScene, seed int64) string {
	var b strings.Builder
	cases := []struct {
		name, note string
		done       int
		ctx        float64
		w, h       int
		bubble     string
		ask        bool
	}{
		{"two stars, one cell", "the head walks into a cell a star already holds, and the sky " +
			"reads one short &mdash; the count error this channel exists to avoid",
			18, 0.05, 80, 24, "", false},
		{"the separation bar", "the head passes 1.00 screen units from a lit star. Nothing is " +
			"lost; two marks simply read as one, which is the same error wearing a different coat",
			18, 0.05, 125, 28, "", false},
		{"the readout, and then a star", "the guarded walk refuses this one entirely and the star " +
			"arrives in place; unguarded it crosses the readout's ground and lands on a neighbour",
			23, 0.56, 125, 28, "", false},
		{"the disc", "his ruling of 2026-09-06 on the sun's face: <i>\"Just the star.\"</i> " +
			"The walk goes AWAY from the moon, so this only arises on a flipped track",
			23, 0.05, 80, 24, "", false},
		{"the balloon eats the star, at 40&times;12", "a knock fires in the same instant as two " +
			"thirds of real arrivals, so at two thirds of them an opaque balloon is on screen in " +
			"the same frame &mdash; and at this size it covers the band. <b>47.0% of this sky is " +
			"ground something else draws on</b>, against 2.6% at 80&times;24 and 1.5% at 125&times;28",
			21, 0.95, 40, 12, "allow Bash?", true},
	}
	cfgOn := trackCfg{N: 2, Rmax: 6, DyMax: 3, TrackMin: 4, Sep: 2}
	for _, cs := range cases {
		s := base
		s.Done, s.Ctx, s.W, s.H = cs.done, cs.ctx, cs.w, cs.h
		s.Bubble, s.Ask = cs.bubble, cs.ask
		s.Mag = scape.StarMagnitudeAt(cs.done-1, seed)
		h2, l2, ok := starfallArrival(s)
		if !ok {
			continue
		}
		_, sh2, _, _, _ := starfallStage(s)
		f2 := starfallForbidden(s)

		gb := s
		gb.Done, gb.Marks = cs.done-1, nil
		gc0, _ := starfallRender(gb)
		gWant := len(starCellsIn(gc0)) + 1

		cfgOff := cfgOn
		cfgOff.NoGuards = true
		trOn, whyOn := starfallTrack(s, h2, l2, f2, sh2, cfgOn)
		trOff, _ := starfallTrack(s, h2, l2, f2, sh2, cfgOff)

		// The phase where the unguarded walk is worst, found rather than guessed.
		worstP, worstNote := 0.5, ""
		for _, p := range []float64{0, .2, .35, .5, .7, .85, 1} {
			fr := s
			fr.Done = cs.done - 1
			fr.Marks = starfallMarks(trOff, p, 1, nil)
			fc, reps := starfallRender(fr)
			ps := starCellsIn(fc)
			eat := 0
			for _, r := range reps {
				if r.Eaten {
					eat++
				}
			}
			sep := starMinSep(ps)
			switch {
			case len(ps) != gWant:
				worstP, worstNote = p, fmt.Sprintf("%d of %d stars &mdash; a star is missing", len(ps), gWant)
			case eat > 0 && worstNote == "":
				worstP, worstNote = p, fmt.Sprintf("%d marks eaten by something drawn later", eat)
			case sep >= 0 && sep < 2 && worstNote == "":
				worstP, worstNote = p, fmt.Sprintf("two marks %.2f screen units apart &mdash; they read as one", sep)
			}
		}
		if worstNote == "" {
			worstNote = "nothing measurably broke at this star"
		}

		var cells strings.Builder
		cells.WriteString(`<div class="film">`)
		for _, arm := range []struct {
			track []starPt
			label string
			off   bool
		}{{trOn, "guard ON", false}, {trOff, "guard OFF", true}} {
			fr := s
			fr.Done = cs.done - 1
			fr.Marks = starfallMarks(arm.track, worstP, 1, nil)
			c, reps := starfallRender(fr)
			ps := starCellsIn(c)
			eat := 0
			for _, r := range reps {
				if r.Eaten {
					eat++
				}
			}
			sep := starMinSep(ps)
			sepS := "n/a"
			if sep >= 0 {
				sepS = fmt.Sprintf("%.2f", sep)
			}
			note := fmt.Sprintf("track %d &middot; %d of %d stars &middot; sep %s &middot; %d eaten",
				len(arm.track), len(ps), gWant, sepS, eat)
			if !arm.off {
				note += " &middot; stopped by " + whyOn
			}
			if len(ps) != gWant || eat > 0 || (sep >= 0 && sep < 2) {
				note = `<b class="bad">` + note + `</b>`
			}
			fmt.Fprintf(&cells, `<div class="cel"><div class="lv">%s</div>%s<div class="st">%s</div></div>`,
				arm.label,
				c.HTMLFragmentCropAs(0, 0, s.W, skyRowsOf(s.H)+2, 13, term.Profile256), note)
		}
		cells.WriteString(`</div>`)
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s<br><span class="mono">%d&times;%d &middot; ctx %.2f &middot; `+
			`star #%d at (%d,%d) &middot; phase %.2f</span><br>`+
			`<b>with the guard off:</b> %s</div></div>%s</div>`,
			cs.name, cs.note, cs.w, cs.h, cs.ctx, cs.done-1, h2.X, h2.Y, worstP,
			worstNote, cells.String())
	}
	return b.String()
}

func starfallHeader(seed int64, sc starScene) string {
	sha := "unknown"
	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	dirty := ""
	if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil && len(out) > 0 {
		dirty = " (working tree modified)"
	}
	return fmt.Sprintf(`<div class="hdr">`+
		`HEAD <b>%s</b>%s &middot; generated %s &middot; `+
		`<span class="mono">go run . -starfall &lt;path&gt; -seed %d</span><br>`+
		`Scene: %d&times;%d, seed %d, %02.0f:00, context %.2f, Total %d, companion right. `+
		`Rendered through <b>Profile256</b> &mdash; the cube the product paints with, not truecolor. `+
		`Every frame is a real composited canvas read back with ResolveAt AFTER drawScene.`+
		`</div>`, sha, dirty, time.Now().Format("2006-01-02 15:04"), seed,
		sc.W, sc.H, seed, sc.Tod*24, sc.Ctx, sc.Total)
}

const starfallCSS = `<style>
h2{font:600 14px ui-monospace,monospace;color:#d8d8e0;margin:34px 0 8px;letter-spacing:.06em;text-transform:uppercase}
.hdr{font:11px ui-monospace,monospace;color:#8a8a99;background:#141419;border:1px solid #222229;
  border-radius:6px;padding:12px 14px;margin:0 0 20px;line-height:1.6}
.card.wide .meta{width:100%;margin-bottom:10px}
.card.wide{display:block}
.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden;display:inline-block}
.ctl{position:sticky;top:0;z-index:9;display:flex;gap:14px;align-items:center;margin:0 0 16px;
  padding:12px 14px;background:#16161c;border:1px solid #272730;border-radius:8px}
.ctl input[type=range]{flex:1;accent-color:#6f8fd0}
.btn{background:#22222b;color:#d8d8e0;border:1px solid #33333f;border-radius:5px;padding:5px 11px;
  font:12px ui-monospace,monospace;cursor:pointer}
.now{font:12px ui-monospace,monospace;color:#9a9aa8;min-width:72px}
.stage .fr{display:none}.stage .fr.on{display:block}
.film{display:flex;gap:8px;flex-wrap:wrap;align-items:flex-start}
.cel{border:1px solid #222229;border-radius:5px;padding:5px;background:#101015}
.grid{display:flex;gap:12px;flex-wrap:wrap;align-items:flex-start}
.cell{border:1px solid #222229;border-radius:6px;padding:8px;background:#101015}
.lv{font:10px ui-monospace,monospace;color:#7d7d8a;margin-bottom:5px;letter-spacing:.04em}
.st{font:10px ui-monospace,monospace;color:#6f6f7d;margin-top:5px}
.st.bad{color:#ff8080}
.bad{color:#ff8080}.good{color:#7fd18f}
.mono{font-family:ui-monospace,monospace}
table{border-collapse:collapse;font:11px ui-monospace,monospace;color:#9a9aa8;margin:8px 0 4px}
th,td{border:1px solid #26262e;padding:5px 9px;text-align:left}
th{color:#c8c8d4;font-weight:600}
.seam{font-family:Menlo,monospace;color:#e8e8f0;background:#101015;border:1px solid #222229;
  border-radius:5px;padding:8px 10px;margin:6px 0;letter-spacing:0}
.paste{font:13px Menlo,monospace;color:#d8d8e0;background:#0a0a0d;border:1px solid #26262e;
  border-radius:5px;padding:10px;user-select:all}
.risk{font:12px ui-monospace,monospace;color:#8a8a99;line-height:1.65;background:#141419;
  border:1px solid #222229;border-left:3px solid #6f8fd0;border-radius:6px;padding:14px 16px;margin:10px 0 30px}
.risk b{color:#d8d8e0}
</style>`

const starfallRisks = `<h2>The open risks, stated plainly</h2><div class="risk">
<b>1. Nobody has seen this move in a terminal.</b> Every number on this page is cell-and-index
measurement inside the compositor. Whether 6 cells over 6 frames at 12&nbsp;fps reads as a fall or
as a glitch is his eye's call, and this project's rule is that only a terminal proves a terminal.<br><br>
<b>2. A still frame of the proposed arm shows the cost and not the cue.</b> Mid-fall it is n
asterisks with one in the wrong place &mdash; visually indistinguishable from the layout bug his
2026-09-10 report was about. He reviews by screenshot. That is the strongest argument for A5,
and it is why section L exists.<br><br>
<b>3. The locked position row is suspended for the duration of every arrival.</b> CLAUDE.md:
<i>"Position is fixed by index and seed so a star lights where it always was."</i> The count rule
holds; the position rule does not, for up to half a second. The session 28 study noticed this and
did not discharge it. Only he can.<br><br>
<b>4. He will miss most of them, and the fall is not the fix for "not shinning".</b> The only
perception threshold this project owns is the kitten dwell: he missed subagents that lived 6.9&nbsp;s
and 15.7&nbsp;s, and saw them at 2.5&nbsp;min. A sub-second event is an order of magnitude under
that. The star persists and carries the information, so the fall is a garnish on a channel he has
twice said he never notices &mdash; not the answer to it.<br><br>
<b>5. The existing suite stays green and is therefore blind.</b> star_steady_test.go builds its lit
set from <span class="mono">r == '*'</span>; with an arrival gated on a field the tests never set,
it never fires. A new test is not optional.<br><br>
<b>6. If the '*' itself moves, that test's contract must be reopened</b> &mdash; and it is the test
that locked the fix his 2026-09-10 report bought. A5 is the only arm that rewrites nothing.<br><br>
<b>7. The balloon hole cannot be closed from inside package scape.</b> drawScene runs after
Shore.Update, so the opaque balloon wins the cell. At 40&times;12 with an ask up it can cover the
entire star band. The honest options at small sizes are: switch the fall off, or do not make the
moving mark a '*'.<br><br>
<b>8. Menlo Regular, Terminal.app, Profile256, this machine.</b> Ghostty, iTerm, WezTerm and
SF&nbsp;Mono were not probed, and the seam in section F is unmeasured on glass for both candidate
tail glyphs.<br><br>
<b>9. Brightness is already spoken for.</b> Per-star magnitude is an open question he is actively
tuning &mdash; <i>"i dont see a big difference in brightness between stars"</i>. A landing flare
spends the same alpha headroom. Two rulings on one channel should not be made in ignorance of each
other, which is why the flare is an arm here and not the proposal.
</div>`
