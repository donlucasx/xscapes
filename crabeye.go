package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// The crab's near eye, drafted from his screenshot. Every frame here is the
// PRODUCT drawing the crab at its nearest rung with a real ask balloon over it
// -- the same picture he screenshotted -- with only the eye art swapped through
// the study door. Nothing is composed by hand.

// crabNearShot renders the crab fully walked up, asking, in a real scene.
func crabNearShot(alert []string, col *term.RGB, seed int64, t float64, px int) string {
	companion.StudyNearEyeAlert, companion.StudyNearEyeCol = alert, col
	defer func() { companion.StudyNearEyeAlert, companion.StudyNearEyeCol = nil, nil }()

	crab := companion.NewCrab()
	crab.FaceLeft(true)
	// Walk it all the way up. The ladder is two rungs of nearStep each, and
	// Approach is the only thing that moves it -- drawScene does not call it,
	// the live loop does (frames.go). A study that forgets this renders the
	// shipped 12x7 crab and reports on a picture he has never complained about.
	for i := 0; i < 24; i++ {
		crab.Approach(0.1, true)
	}
	st := catState(companion.NeedsYou, "Claude needs your permission", true)
	c := catStage(118, 26, seed, 0.86, 0.42, true, crab, st, t)

	_, chh := crab.Size()
	ccw, _ := crab.Size()
	lay := compose(118, ccw, true)
	dx, _, dh := crab.DrawnBox(118)
	x0 := lay.CatX + dx - 34
	if x0 < 0 {
		x0 = 0
	}
	y0 := c.H - 2 - chh - (dh - chh) - 5
	if y0 < 0 {
		y0 = 0
	}
	return c.HTMLFragmentCropAs(x0, y0, c.W, c.H, px, term.Profile256)
}

// crabEyeTile renders one eye ALONE at a big size, so the sixteen subpixels can
// actually be looked at.
func crabEyeTile(alert []string, col term.RGB, px int) string {
	bm := companion.ParseBitmap(alert)
	q := bm.ToQuadrant()
	c := canvas.New(len([]rune(q[0]))+2, len(q)+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	// Painted over the coat, because that is what the hole shows: the eye is
	// drawn with PlotOn and its empty subpixels are salmon, not sky.
	coat := companion.Coats["cream"]
	coat = term.RGB{R: 255, G: 135, B: 135}
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			c.Near().Plot(x, y, ' ', coat, 1)
			c.SetBG(x, y, coat)
		}
	}
	for y, row := range q {
		for x, r := range []rune(row) {
			if r == ' ' {
				continue
			}
			c.Near().PlotOn(1+x, 1+y, r, col, coat, 1)
		}
	}
	return c.HTMLFragmentAs(px, term.Profile256)
}

// crabEyeMeasure is the colour ladder, measured through the product's own
// quantiser rather than argued about.
//
// ⚠ TWO DRAFTS CONTRADICTED EACH OTHER HERE and both were half right. One held
// that every softer mint is blocked -- desaturating crosses the chroma-30 cliff
// and gets BOOSTED, dimming falls onto the grey ramp. The other held that
// 186,238,196 lands on a distinct entry. The quantiser says: the cliff is real
// and the grey ramp is real, but there ARE usable entries below the shipped
// one, and the cost of taking them is not saturation -- it is that the ask's
// colour moves TOWARD the working bead's.
func crabEyeMeasure() string {
	luma := func(c term.RGB) float64 {
		return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	}
	dist := func(a, b term.RGB) float64 {
		dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
		return math.Sqrt(dr*dr + dg*dg + db*db)
	}
	chroma := func(c term.RGB) float64 {
		mx := math.Max(float64(c.R), math.Max(float64(c.G), float64(c.B)))
		mn := math.Min(float64(c.R), math.Min(float64(c.G), float64(c.B)))
		return mx - mn
	}
	q := func(c term.RGB) term.RGB { return term.Profile256.Quantise(c, true) }
	ship := q(companion.EyeAlert)
	work := q(term.RGB{R: 168, G: 236, B: 176})

	var b strings.Builder
	b.WriteString(`<table><tr><th>source</th><th>chroma</th><th>renders as</th><th>cube</th>` +
		`<th>luma</th><th>distance from the working bead</th></tr>`)
	row := func(name string, c term.RGB, mark string) {
		r := q(c)
		fmt.Fprintf(&b, `<tr><td>%s%s</td><td>%.0f</td><td>(%d,%d,%d)</td><td>%d</td>`+
			`<td>%.0f</td><td>%.0f</td></tr>`,
			name, mark, chroma(c), r.R, r.G, r.B, r.Index256(), luma(r), dist(r, work))
	}
	row("the shipped EyeAlert 232,252,226", companion.EyeAlert, " &larr; today")
	row("the working bead 168,236,176", term.RGB{R: 168, G: 236, B: 176}, " &larr; what it must not look like")
	for _, c := range []term.RGB{
		{R: 220, G: 244, B: 220}, {R: 210, G: 238, B: 212},
		{R: 200, G: 230, B: 200}, {R: 186, G: 238, B: 196},
		{R: 184, G: 210, B: 184}, {R: 170, G: 200, B: 172},
		{R: 240, G: 240, B: 235},
	} {
		mark := ""
		switch {
		case c.R == 186:
			mark = " &larr; proposed by one draft"
		case c.R == 184:
			mark = " &larr; proposed by another"
		case c.R == 240:
			mark = " &larr; a near-neutral white"
		}
		row(fmt.Sprintf("%d,%d,%d", c.R, c.G, c.B), c, mark)
	}
	b.WriteString(`</table>`)
	fmt.Fprintf(&b, `<div class="st">The shipped tone's own chroma is %.0f, which is UNDER the `+
		`neutralChroma cliff of 30, so it is never boosted. Sources at 24&ndash;28 all collapse `+
		`onto the same cube entry &mdash; softening within that band changes nothing at all. `+
		`Cross 30 and the entry does move, but toward the working bead: the ask-vs-working colour `+
		`distance falls from %.0f to as little as %.0f, which halves the very channel the near `+
		`pose was built to add. And a near-neutral white leaves the colour cube for the GREY RAMP, `+
		`which is what GlyphBoost exists to escape.<br>`+
		`<b>Two drafts disagreed about this and both were half right.</b> There ARE softer entries; `+
		`the cost is not saturation, it is separation.</div>`,
		chroma(companion.EyeAlert), dist(ship, work), 40.0)
	return b.String()
}

func crabEyeSection(seed int64) string {
	var b strings.Builder
	b.WriteString(`<h2>Round 5 &mdash; the crab's near eye</h2>`)
	b.WriteString(`<p class="nt">From your screenshot: <i>"can we mockup an alternative approach ` +
		`for the crab close up where the eyes read softer? maybe thiner lines or an alt approach ` +
		`where the companion reads gentler yet ready"</i>.<br>` +
		`Every frame below is <b>the product drawing the crab</b>, walked all the way up its ` +
		`ladder with a real ask balloon over it &mdash; the same picture you screenshotted &mdash; ` +
		`with only the eye art swapped. <b>"Thinner lines" cannot mean a thinner wall:</b> the ` +
		`eye is 4&times;8 source pixels, which is <b>sixteen subpixels</b> after the halving, and ` +
		`the ring's wall is already one subpixel thick. It can only mean LESS wall.<br>` +
		`&#9888; What it must not lose: the ask has to stay a different SHAPE from working. At the ` +
		`shipped size the ask was <span class="mono">O</span> against <span class="mono">o</span>, ` +
		`the same cell in a different weight, and the shape difference is the whole thing the near ` +
		`pose bought.</p>`)

	// The control first.
	fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">Today &mdash; the ring</div>`+
		`<div class="nt">What is on your screen now: a full 4&times;4 ring with four square corners `+
		`and a 2&times;2 hole painted in the coat. That salmon bar through the middle is the hole.`+
		`</div></div><div class="grid">`+
		`<div class="cell"><div class="lv">the eye alone, over coat</div>%s</div>`+
		`<div class="cell"><div class="lv">the working bead, for comparison</div>%s</div>`+
		`<div class="cell"><div class="lv">as it renders on screen</div>%s</div>`+
		`</div></div>`,
		crabEyeTile(companion.NearEyeShipped(), companion.EyeAlert, 34),
		crabEyeTile(companion.NearEyeBead(), term.RGB{R: 168, G: 236, B: 176}, 34),
		crabNearShot(nil, nil, seed, 4, 12))

	for _, e := range crabEyeAlts {
		col := e.Col
		var risks strings.Builder
		for _, x := range e.Risks {
			fmt.Fprintf(&risks, `<div class="risk1">&#9888; %s</div>`, x)
		}
		ink := 0
		for _, r := range e.Alert {
			ink += strings.Count(r, "#")
		}
		fmt.Fprintf(&b, `<div class="card wide"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div>`+
			`<div class="st">%d of 32 source pixels inked (the shipped ring is %d) &middot; `+
			`ink %d,%d,%d</div>`+
			`<div class="st">%s</div>%s</div><div class="grid">`+
			`<div class="cell"><div class="lv">the eye alone, over coat</div>%s</div>`+
			`<div class="cell"><div class="lv">as it renders on screen</div>%s</div>`+
			`</div></div>`,
			e.Name, e.Idea, ink, 24, col.R, col.G, col.B, e.Still, risks.String(),
			crabEyeTile(e.Alert, col, 34),
			crabNearShot(e.Alert, &col, seed, 4, 12))
	}

	b.WriteString(`<h2>The colour, measured rather than argued</h2>`)
	b.WriteString(crabEyeMeasure())
	return b.String()
}

// ---------------------------------------------------------------------------
// THE ROTATION. His ruling of 2026-09-12: keep four of the five and rotate
// them, "random each time".
// ---------------------------------------------------------------------------

// crabEyeRotation is the four he kept, in the order they were drafted.
// Sea Glass is deliberately absent: it was the only one of the five that
// changed the COLOUR rather than the shape, and the quantiser measurement says
// colour is the one axis with almost nowhere to go.
func crabEyeRotation() []crabEyeAlt {
	keep := map[string]bool{"round-off": true, "the-high-pupil": true,
		"the-hood": true, "the-hooded-ring": true}
	var out []crabEyeAlt
	for _, e := range crabEyeAlts {
		if keep[e.Key] {
			out = append(out, e)
		}
	}
	return out
}

// crabEyePick is which eye a given ask gets.
//
// TWO PROPERTIES, and the second is the one that decides whether this reads as
// character or as a glitch:
//
//  1. It must be RANDOM ACROSS asks. That is what makes it safe under the
//     encoding rule -- a random eye carries no information, so it cannot be a
//     second meaning on a channel that already says "the agent needs you".
//     The same argument is what permits the crab's claw twitch.
//  2. It must be CONSTANT WITHIN one ask. An eye chosen per frame would change
//     the companion's face while he is reading the question, twenty times a
//     second. So it is a function of an ASK ORDINAL, not of the clock, and the
//     ordinal only moves on the rising edge of NeedsYou.
//
// The seed is mixed in so two scapes started at the same moment do not show the
// same face, and so a restart does not always begin on the same eye.
func crabEyePick(askN int, seed int64) int {
	n := len(crabEyeRotation())
	if n == 0 {
		return 0
	}
	h := uint64(askN)*0x9E3779B97F4A7C15 ^ uint64(seed)*0xBF58476D1CE4E5B9
	h ^= h >> 31
	h *= 0x94D049BB133111EB
	h ^= h >> 29
	return int(h % uint64(n))
}

// crabEyeRotationSection shows the rotation as a sequence of asks, because the
// thing being judged is not any one eye -- it is whether four faces in rotation
// read as one animal or as four.
func crabEyeRotationSection(seed int64) string {
	rot := crabEyeRotation()
	var b strings.Builder
	b.WriteString(`<h2>The rotation</h2>`)
	fmt.Fprintf(&b, `<p class="nt">Your four, rotating at random: <b>%s</b>. `+
		`Sea Glass is out &mdash; it was the only one of the five that changed the COLOUR rather `+
		`than the shape, and the cube measurement says colour has almost nowhere to go.<br>`+
		`<b>Random is the right call and it is worth saying why.</b> A random eye carries no `+
		`information, so it cannot put a second meaning on a channel that already says "the agent `+
		`needs you" &mdash; the same argument that lets the claw twitch exist. Tying it to a kind `+
		`of moment would have bound a second variable to one channel, which the encoding rule `+
		`forbids.<br>`+
		`&#9888; <b>One property this has to have, or it reads as a glitch:</b> the eye is picked `+
		`on the RISING EDGE of the ask and held for its whole duration. An eye chosen per frame `+
		`would change the companion's face twenty times a second while you are reading the `+
		`question. So it is a function of an ask ordinal, never of the clock.</p>`,
		strings.Join(func() []string {
			var n []string
			for _, e := range rot {
				n = append(n, e.Name)
			}
			return n
		}(), " &middot; "))

	b.WriteString(`<div class="grid">`)
	for ask := 1; ask <= 8; ask++ {
		e := rot[crabEyePick(ask, seed)]
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">ask #%d &mdash; %s</div>%s</div>`,
			ask, e.Name, crabEyeTile(e.Alert, e.Col, 26))
	}
	b.WriteString(`</div>`)

	// The distribution, counted rather than asserted.
	counts := map[string]int{}
	const trials = 4000
	for i := 0; i < trials; i++ {
		counts[rot[crabEyePick(i, seed)].Name]++
	}
	b.WriteString(`<div class="st">Over ` + fmt.Sprint(trials) + ` asks: `)
	var parts []string
	for _, e := range rot {
		parts = append(parts, fmt.Sprintf("%s %.1f%%", e.Name, 100*float64(counts[e.Name])/trials))
	}
	b.WriteString(strings.Join(parts, " &middot; "))
	b.WriteString(`. Counted, not assumed &mdash; a hash that clumps would show here.</div>`)

	// The thing an even distribution does NOT tell you.
	repeats := 0
	for i := 1; i < trials; i++ {
		if crabEyePick(i, seed) == crabEyePick(i-1, seed) {
			repeats++
		}
	}
	fmt.Fprintf(&b, `<div class="st">&#9888; <b>But %.1f%% of asks draw the same eye as the one `+
		`before them</b> &mdash; %d of %d, which is exactly what four-way random should do. In the `+
		`eight asks above, the hood comes up three times and the hooded ring once. Over a working `+
		`day that reads as the rotation being broken rather than as chance. If you want it, the `+
		`cheap fix is to draw from the three that are NOT the previous eye: still random, still `+
		`carries no information, never repeats back to back. Say the word and it is one line.</div>`,
		100*float64(repeats)/float64(trials-1), repeats, trials-1)

	// And the same four, in the scene, so four faces can be compared as faces.
	b.WriteString(`<div class="grid">`)
	for _, e := range rot {
		col := e.Col
		fmt.Fprintf(&b, `<div class="cell"><div class="lv">%s</div>%s</div>`,
			e.Name, crabNearShot(e.Alert, &col, seed, 4, 11))
	}
	b.WriteString(`</div>`)
	return b.String()
}
