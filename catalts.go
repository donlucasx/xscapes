package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// catalts is the design round for the cat: alternatives rendered, he rules.
// The same shape Hero was picked in, and the same shape the coat was.
//
// Every candidate goes through the PRODUCT'S OWN Draw -- the same breathing,
// the same procedural tail, the same mirror, the same eye pass -- via the study
// doors in internal/companion/study_art.go. Nothing here re-implements the
// companion, because a study that re-implements the thing it is judging ends up
// judging the study.

// catAlt is one candidate body.
type catAlt struct {
	Key string
	// Animal is "cat" or "crab". The crab's Done candidates are rendered by
	// the study rather than by Cat.Draw, because the crab's draw path takes
	// its upper block from crabUpper(st, t) and there is no Done slot in it
	// yet -- adding one is the feature, not the mockup.
	Animal string
	Name   string
	Idea   string
	Rows   []string // 24x28, the seated body
	Note   string
	Risks  []string
}

// catRung is one candidate for the middle rung of the come-closer ladder.
type catRung struct {
	Key      string
	Name     string
	Idea     string
	Upper    []string // 32 wide, the level head and chest
	UpperAsk []string // the same block, asking
	Lower    []string // 32 wide, body and legs; upper+lower = 36 rows
	EyeCells [2]int
	EyeRow   int
	Note     string
	Risks    []string
}

// catAltCheck is the mechanical validation. Art that fails is shown failing
// rather than quietly dropped: a design round that hides its rejects teaches
// nobody why the constraint exists.
type catAltCheck struct {
	Rows, Cols   int
	Ragged       bool
	Ink          int
	Cells        string // the rendered cell box, "12x7"
	EyeHoles     bool   // the eye cells are gaps, per his 2026-09-05 ruling
	TailCorridor int    // ink found in the tail's corridor, which must be 0
	Symmetric    int    // cells that differ between the sprite and its mirror
	DeltaVsBody  int    // CELLS that differ from the shipped Working pose
}

const (
	catSrcW, catSrcH = 24, 28
	catCells         = 12
)

func checkCatAlt(rows []string) catAltCheck { return checkAltAgainst(rows, companion.CatBody) }

func checkAltAgainst(rows, baseArt []string) catAltCheck {
	var k catAltCheck
	k.Rows = len(rows)
	if len(rows) == 0 {
		return k
	}
	k.Cols = len([]rune(rows[0]))
	for _, r := range rows {
		if len([]rune(r)) != k.Cols {
			k.Ragged = true
		}
		for _, c := range r {
			if c == '#' {
				k.Ink++
			}
		}
	}
	if k.Ragged || k.Rows != catSrcH || k.Cols != catSrcW {
		return k
	}
	bm := companion.ParseBitmap(rows)
	q := bm.ToQuadrant()
	k.Cells = fmt.Sprintf("%dx%d", len([]rune(q[0])), len(q))

	// His ruling of 2026-09-05: "regarding eyes, keep holes (as is)." The eye
	// glyph is plotted on top at cells 2 and 6, cell row 2.
	//
	// ⚠ THE FIRST VERSION OF THIS CHECK TESTED SOURCE ROWS 8..11 AND FAILED
	// THE SHIPPED CAT. One cell is four source rows, but CatBody opens the eye
	// on rows 8..9 only and fills 10..11 with cheek -- the glyph replaces the
	// whole cell either way, so the hole is the TOP half being clear, which is
	// what puts sky behind the eye. A checker that rejects the art it was
	// derived from is measuring its author's assumption; catEyeHolesControl is
	// the positive control that now says so out loud.
	k.EyeHoles = true
	for _, cx := range []int{4, 12} {
		for dx := 0; dx < 2; dx++ {
			for y := 8; y < 10; y++ {
				if []rune(rows[y])[cx+dx] == '#' {
					k.EyeHoles = false
				}
			}
		}
	}
	// The tail is drawn AFTER the body, sweeping up from the right hip. Ink in
	// its corridor is ink the tail will be drawn on top of.
	for y := 11; y < catSrcH; y++ {
		for x := 18; x < catSrcW; x++ {
			if []rune(rows[y])[x] == '#' {
				k.TailCorridor++
			}
		}
	}
	// Symmetry is a property of even column parity in this medium, so it is
	// measured on the RENDERED cells rather than on the source pixels.
	m := bm.Mirrored().ToQuadrant()
	for y := range q {
		a, b := []rune(q[y]), []rune(m[y])
		for x := range a {
			if a[x] != b[len(b)-1-x] {
				k.Symmetric++
			}
		}
	}
	// The number this whole round exists for: how many CELLS this pose differs
	// from the one the animal holds while working. Today that number is 0.
	base := companion.ParseBitmap(baseArt).ToQuadrant()
	for y := range q {
		if y >= len(base) {
			break
		}
		a, b := []rune(q[y]), []rune(base[y])
		for x := range a {
			if x < len(b) && a[x] != b[x] {
				k.DeltaVsBody++
			}
		}
	}
	return k
}

// catStage builds a real scene with a given companion, so a candidate is
// judged where it will live rather than on a slide.
func catStage(w, h int, seed int64, tod, ctx float64, mirror bool,
	cat *companion.Cat, st reduce.State, t float64) *canvas.Canvas {

	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	cat.FaceLeft(mirror)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, mirror)
	sh.MoonX = lay.MoonX
	st.Act.TimeOfDay, st.Act.ContextUsed = tod, ctx
	for i := 0; i < 30; i++ {
		sh.Update(c, t-1.5+float64(i)/20, st.Act)
	}
	sh.Update(c, t, st.Act)
	drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
	return c
}

// catState is a believable state for one pose.
func catState(pose companion.State, bubble string, ask bool) reduce.State {
	st := reduce.State{Pose: pose, Act: scape.Activity{Working: true, Level: 0.55}}
	switch pose {
	case companion.Resting:
		st.Act = scape.Activity{}
	case companion.NeedsYou:
		st.Bubble, st.BubbleAsk = bubble, true
	case companion.Done:
		st.Bubble, st.BubbleAsk = bubble, false
	}
	_ = ask
	return st
}

// catSprite renders one candidate body ALONE, big, on nothing -- the silhouette
// test. A companion has to read as its animal before it has to read in a scene.
func catSprite(rows []string, pose companion.State, t float64, px int, mirror bool) string {
	cat := companion.NewCat()
	cat.SetCoat(companion.Coats["cream"])
	if pose == companion.NeedsYou && rows != nil {
		cat.SetAskArt(rows)
	} else if rows != nil {
		cat.SetBodyArt(rows)
	}
	cat.FaceLeft(mirror)
	cw, ch := cat.Size()
	c := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cat.Draw(c.Near(), 2, 1, t, pose)
	return c.HTMLFragmentAs(px, term.Profile256)
}

// catInScene renders the candidate where it lives, cropped to the companion's
// own corner so the animal is what is being looked at.
func catInScene(rows []string, pose companion.State, t float64, px int,
	w, h int, seed int64, tod float64, bubble string) string {

	cat := companion.NewCat()
	cat.SetCoat(companion.Coats["cream"])
	if rows != nil {
		cat.SetAskArt(rows)
	}
	st := catState(pose, bubble, pose == companion.NeedsYou)
	c := catStage(w, h, seed, tod, 0.45, true, cat, st, t)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, true)
	x0 := lay.CatX - 26
	if x0 < 0 {
		x0 = 0
	}
	y0 := c.H - 2 - chh - 6
	if y0 < 0 {
		y0 = 0
	}
	return c.HTMLFragmentCropAs(x0, y0, c.W, c.H, px, term.Profile256)
}

// catAnim is the candidate breathing and wagging: a still cannot show either,
// and both are the cat's own advantage over the crab -- measured at 92 distinct
// body pictures against the crab's 8.
func catAnim(rows []string, pose companion.State, frames int, px int) string {
	var b strings.Builder
	b.WriteString(`<div class="stage cstage">`)
	for i := 0; i < frames; i++ {
		on := ""
		if i == 0 {
			on = " on"
		}
		fmt.Fprintf(&b, `<div class="fr%s" data-i="%d">%s</div>`,
			on, i, catSprite(rows, pose, float64(i)/12, px, true))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func catCheckRow(k catAltCheck) string {
	bad := func(ok bool, s string) string {
		if ok {
			return s
		}
		return `<b class="bad">` + s + `</b>`
	}
	return fmt.Sprintf(`<div class="st">%s &middot; %s &middot; %d px of ink &middot; `+
		`%s &middot; %s &middot; <b>%d cells differ from working</b></div>`,
		bad(!k.Ragged && k.Rows == catSrcH && k.Cols == catSrcW,
			fmt.Sprintf("%dx%d source", k.Cols, k.Rows)),
		k.Cells, k.Ink,
		bad(k.EyeHoles, map[bool]string{true: "eyes are holes", false: "EYES ARE PAINTED"}[k.EyeHoles]),
		bad(k.TailCorridor == 0, fmt.Sprintf("%d px in the tail corridor", k.TailCorridor)),
		k.DeltaVsBody)
}

func catAltsHeader() string {
	sha := "unknown"
	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	return fmt.Sprintf(`<div class="hdr">HEAD <b>%s</b> &middot; generated %s &middot; `+
		`<span class="mono">go run . -catalts &lt;path&gt;</span><br>`+
		`Every candidate is drawn by the product's own <span class="mono">Cat.Draw</span> &mdash; `+
		`the same breath, the same procedural tail, the same mirror, the same eye pass &mdash; `+
		`through the study doors in <span class="mono">internal/companion/study_art.go</span>. `+
		`Nothing here is re-implemented, and nothing here ships.<br>`+
		`Rendered at <b>Profile256</b>, the cube the product paints with.`+
		`</div>`, sha, time.Now().Format("2006-01-02 15:04"))
}

func catDelta(a, b []string) int {
	qa := companion.ParseBitmap(a).ToQuadrant()
	qb := companion.ParseBitmap(b).ToQuadrant()
	n := 0
	for y := range qa {
		if y >= len(qb) {
			break
		}
		ra, rb := []rune(qa[y]), []rune(qb[y])
		for x := range ra {
			if x < len(rb) && ra[x] != rb[x] {
				n++
			}
		}
	}
	return n
}

// double2x doubles a bitmap source in both axes, which is how the near ladder's
// top rung is made: 24x28 doubles to 48x56, exactly 24x14 cells.
func double2x(rows []string) []string {
	out := make([]string, 0, len(rows)*2)
	for _, r := range rows {
		var b strings.Builder
		for _, c := range r {
			b.WriteRune(c)
			b.WriteRune(c)
		}
		out = append(out, b.String(), b.String())
	}
	return out
}

// catRungSprite composes one rung by hand, because the product has NO cat near
// path yet -- crab_near.go returns zero rungs for the cat. This is the one
// place on the page that is a study composition rather than the product's own
// draw, and it says so on the page.
func catRungSprite(rows []string, eyeCells [2]int, eyeRow int, scale float64,
	pose companion.State, t float64, px int, mirror bool) string {

	bm := companion.ParseBitmap(rows)
	cat := companion.NewCat()
	wag := math.Sin(t * 2.4)
	if pose == companion.NeedsYou {
		wag = math.Sin(t * 5.0)
	}
	// scale 0 means "no tail": the crab's whole body is bitmap and it has none.
	if scale > 0 {
		cat.TailAt(bm, wag, 0, 1.0, scale)
	}
	if mirror {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	cw, ch := len([]rune(q[0])), len(q)
	c := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	(&companion.Sprite{Rows: q, Body: companion.Coats["cream"]}).Draw(c.Near(), 2, 1)

	glyph, col := 'o', term.RGB{R: 168, G: 236, B: 176}
	if pose == companion.NeedsYou {
		glyph, col = 'O', companion.EyeAlert
	}
	for _, cx := range eyeCells {
		x := cx
		if mirror {
			x = cw - 1 - cx
		}
		c.Near().Plot(2+x, 1+eyeRow, glyph, col, 1)
	}
	return c.HTMLFragmentAs(px, term.Profile256)
}

// catAltsReport is the validation table, printed to stderr when the page is
// generated. Art that fails a constraint is REPORTED, not silently dropped: a
// design round that hides its rejects teaches nobody why the constraint exists.
func catAltsReport() []string {
	var out []string
	out = append(out, "THE CAT, DRAFTED -- candidate validation")
	// The positive control, first and always: the SHIPPED cat must pass every
	// constraint the candidates are held to. A checker nobody has run against
	// known-good art is an opinion with a table around it.
	if ctl := checkCatAlt(companion.CatBody); !ctl.EyeHoles || ctl.TailCorridor > 0 ||
		ctl.Ragged || ctl.Rows != catSrcH || ctl.Cols != catSrcW {
		out = append(out, fmt.Sprintf("!! THE CHECKER REJECTS THE SHIPPED CAT "+
			"(%dx%d, eyeholes %v, %d px in the tail corridor) -- the instrument is wrong, "+
			"not the art. Every row below is void.",
			ctl.Cols, ctl.Rows, ctl.EyeHoles, ctl.TailCorridor))
	} else {
		out = append(out, "positive control: the shipped CatBody passes every constraint below.")
	}
	out = append(out, fmt.Sprintf("%-34s %-10s %-7s %6s %8s %7s %8s",
		"candidate", "source", "cells", "ink", "eyeholes", "tailpx", "vs work"))
	check := func(label string, rows []string, crab bool) {
		base := companion.CatBody
		if crab {
			base = companion.CrabWorkingArt()
		}
		k := checkAltAgainst(rows, base)
		flag := ""
		if k.Ragged || k.Rows != catSrcH || k.Cols != catSrcW {
			flag = "   <-- LOOK"
		} else if !crab && (!k.EyeHoles || k.TailCorridor > 0) {
			// The eye-hole and tail-corridor rules are the CAT'S. The crab has
			// no procedural tail and puts its eyes on stalks; scoring its art
			// against the cat's constraints is the same error as the checker
			// that failed the shipped cat.
			flag = "   <-- LOOK"
		}
		if k.DeltaVsBody == 0 {
			flag = "   <-- IDENTICAL TO WORKING, it carries nothing"
		}
		holes, tail := fmt.Sprint(k.EyeHoles), fmt.Sprint(k.TailCorridor)
		if crab {
			holes, tail = "n/a", "n/a"
		}
		out = append(out, fmt.Sprintf("%-34s %-10s %-7s %6d %8s %7s %8d%s",
			label, fmt.Sprintf("%dx%d", k.Cols, k.Rows), k.Cells, k.Ink,
			holes, tail, k.DeltaVsBody, flag))
	}
	for _, a := range catAskAlts {
		check("ask/"+a.Key, a.Rows, false)
	}
	for _, a := range catDoneAlts {
		check("done/"+a.Key, a.Rows, a.Animal == "crab")
	}
	for _, r := range catRungAlts {
		full := append(append([]string{}, r.Upper...), r.Lower...)
		w := 0
		if len(full) > 0 {
			w = len([]rune(full[0]))
		}
		q := ""
		if len(full) > 0 {
			qq := companion.ParseBitmap(full).ToQuadrant()
			q = fmt.Sprintf("%dx%d", len([]rune(qq[0])), len(qq))
		}
		flag := ""
		if len(full) != 36 || w != 32 {
			flag = "   <-- LOOK: want 32x36"
		}
		out = append(out, fmt.Sprintf("%-34s %-10s %-7s %6s %8s %7s %8s%s",
			"rung/"+r.Key, fmt.Sprintf("%dx%d", w, len(full)), q, "-", "-", "-", "-", flag))
	}
	out = append(out, "")
	out = append(out, "\"vs work\" is how many CELLS the pose differs from THAT ANIMAL working:")
	out = append(out, "the cat against CatBody, the crab against its own shipped working art.")
	out = append(out, "Today that number is 0 for both the ask and the finish, which is the defect.")
	return out
}

// catFrameSet is every DISTINCT picture a pose can produce, hashed.
//
// This is the instrument that matters, and it replaced two worse ones. Comparing
// the BITMAPS proves only that a thing equals itself. Comparing two frames at
// the same clock t measures the two poses' breath and wag being out of PHASE --
// Working breathes at 2.2 s and wags at sin(2.4t), NeedsYou at 1.6 s and
// sin(5.0t) -- and reports a difference that exists only because the clock is
// shared, which is a rate difference wearing a still frame's clothes.
//
// The question a screenshot actually asks is: is this picture ambiguous? So the
// measurement is a SET OVERLAP. If a picture the cat draws while asking is also
// a picture it draws while working, then a person looking at a screenshot of it
// cannot tell which it was, however different the two animations look in motion.
func catFrameSet(st companion.State, samples int) map[string]bool {
	out := map[string]bool{}
	cat := companion.NewCat()
	cat.SetCoat(companion.Coats["cream"])
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	for i := 0; i < samples; i++ {
		t := float64(i) / 20 // 20 fps, the renderer's own default
		c := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat.Draw(c.Near(), 2, 1, t, st)
		var b strings.Builder
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				r, fg, _ := c.ResolveAt(x, y, term.Profile256)
				fmt.Fprintf(&b, "%c%d;", r, fg.Index256())
			}
		}
		out[b.String()] = true
	}
	return out
}

// catFrameSetBody is the same, with the two eye cells BLANKED -- so it measures
// the body alone, which is what a silhouette at a glance reads. The eye is two
// cells of colour out of eighty-four and it is the thing the ask currently
// rests its entire meaning on.
func catFrameSetBody(rows []string, st companion.State, samples int) map[string]bool {
	out := map[string]bool{}
	cat := companion.NewCat()
	cat.SetCoat(companion.Coats["cream"])
	cat.FaceLeft(true)
	if rows != nil {
		cat.SetAskArt(rows)
	}
	cw, ch := cat.Size()
	for i := 0; i < samples; i++ {
		t := float64(i) / 20
		c := canvas.New(cw+4, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat.Draw(c.Near(), 2, 1, t, st)
		var b strings.Builder
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				if catIsEyeCell(x-2, y-1) {
					b.WriteString("EYE;")
					continue
				}
				r, fg, _ := c.ResolveAt(x, y, term.Profile256)
				fmt.Fprintf(&b, "%c%d;", r, fg.Index256())
			}
		}
		out[b.String()] = true
	}
	return out
}

// catAmbiguity is the share of one pose's body pictures that the OTHER pose
// also draws. 100% means a screenshot can never tell them apart.
func catAmbiguity(rows []string, a, b companion.State, samples int) (shared, total int) {
	sa := catFrameSetBody(rows, a, samples)
	sb := catFrameSetBody(rows, b, samples)
	for k := range sa {
		if sb[k] {
			shared++
		}
	}
	return shared, len(sa)
}

// catIsEyeCell says whether a cell inside the sprite box is one of the two the
// eye glyph is plotted into, in the MIRRORED layout the scape ships.
func catIsEyeCell(x, y int) bool {
	if y != 2 {
		return false
	}
	w := catCells
	for _, e := range []int{2, 6} {
		if x == w-1-e {
			return true
		}
	}
	return false
}

// catRung2Shot renders the fixed rung 2 with one of the candidate eye
// treatments, in colour, through a real Sprite draw plus the eye pass.
//
// treatment: "glyph" one centred glyph · "outline" the socket edged in eye
// colour with a dark pupil · "tall" a two-cell pupil in a deepened socket.
func catRung2Shot(treatment string, pose companion.State, t float64, px int, mirror bool) string {
	rows := double2x(companion.CatBodyNear2)
	if treatment == "tall" {
		rows = catRung2DeepEye()
	}
	bm := companion.ParseBitmap(rows)
	cat := companion.NewCat()
	wag := math.Sin(t * 2.4)
	if pose == companion.NeedsYou {
		wag = math.Sin(t * 5.0)
	}
	cat.TailAt(bm, wag, 0, 1.0, 2)
	if mirror {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	w := len([]rune(q[0]))
	c := canvas.New(w+4, len(q)+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	(&companion.Sprite{Rows: q, Body: companion.Coats["cream"]}).Draw(c.Near(), 2, 1)

	glyph, col := 'o', term.RGB{R: 168, G: 236, B: 176}
	switch pose {
	case companion.NeedsYou:
		glyph, col = 'O', companion.EyeAlert
	case companion.Resting:
		glyph = '-'
	case companion.Done:
		glyph = '^'
	}
	put := func(cx, cy int, r rune) {
		x := cx
		if mirror {
			x = w - 1 - cx
		}
		c.Near().Plot(2+x, 1+cy, r, col, 1)
	}
	for _, s := range catRung2Sockets {
		mid := (s[0] + s[1]) / 2
		switch treatment {
		case "glyph":
			put(mid, catRung2EyeRow, glyph)
		case "outline":
			// The socket's outer edges in eye colour, its middle left as sky:
			// an eye with a dark pupil rather than a bright dot.
			put(s[0], catRung2EyeRow, '▐')
			put(s[1], catRung2EyeRow, '▌')
		case "tall":
			put(mid, catRung2EyeRow, glyph)
			put(mid, catRung2EyeRow+1, glyph)
		}
	}
	return c.HTMLFragmentAs(px, term.Profile256)
}
