// Command s28smalls closes the two small unruled items on RESUME item 4, by
// measurement rather than by reading the source.
//
//	go run ./notes/s28-smalls              # both parts
//	go run ./notes/s28-smalls -part eyes   # SetEyeFill on stalked eyes
//	go run ./notes/s28-smalls -part sand   # the sand line written across the litter
//
// SMALL 1 -- SetEyeFill. Three fills, five states, two animals, 200 breathing
// phases each. Every reading is off the RESOLVED cell (rune, fg AND background)
// through canvas.ResolveAt at Profile256, because both fills live entirely in
// the cell's BACKGROUND and a glyph-only reader sees nothing at all.
//
// The "what did the sprite author in that cell" question is answered by
// rebuilding the crab's body in this file from copied bitmaps and drawing it
// with NO eye glyph on top -- and the copy is validated against the product's
// own render at every non-eye cell before any number from it is quoted.
//
// SMALL 2 -- the sand across the litter. Two passes.
//
//	pass 1  every second of every real recording in ~/.config/xscapes/run,
//	        folded through the real reducer, with the real fitted tail, against
//	        the real composition arithmetic. Cheap, so it can run over all of it.
//	pass 2  full RENDERED frames for a sample of the same cases: shore, litter,
//	        sand, read back through ResolveAt. Three renders per case (shore
//	        only / +litter / +sand) so the litter's footprint and the sand's
//	        footprint are both measured from the picture, not asserted.
//	        Pass 2 is what says which one WINS the cell; pass 1 is validated
//	        against it.
//
// The functions marked COPIED are verbatim from the product and must not drift;
// the instrument prints its own check that the copies still agree with what the
// product renders.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func main() {
	part := flag.String("part", "both", "eyes | sand | both")
	flag.Parse()
	if *part == "eyes" || *part == "both" {
		partEyes()
	}
	if *part == "sand" || *part == "both" {
		partSand()
	}
}

// ---------------------------------------------------------------- plumbing

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

// warmShore is the pattern notes/searange records: ONE shore, advanced a few
// hundred steps, so anything that integrates has a history. A fresh shore per
// frame reads as if the session had just started.
func warmShore(seed int64, w, h int, act scape.Activity, moonX float64) (*scape.Shore, float64) {
	sh := scape.NewShore(seed, false)
	sh.MoonX = moonX
	t := 0.0
	for k := 0; k < 400; k++ {
		t += 0.08
		sh.Update(newCanvas(w, h), t, act)
	}
	return sh, t
}

type cell struct {
	r      rune
	fg, bg term.RGB
}

func (c cell) String() string {
	return fmt.Sprintf("%q fg(%3d,%3d,%3d) bg(%3d,%3d,%3d)", string(c.r),
		c.fg.R, c.fg.G, c.fg.B, c.bg.R, c.bg.G, c.bg.B)
}

func readBox(c *canvas.Canvas, x0, y0, w, h int) []cell {
	out := make([]cell, 0, w*h)
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			out = append(out, cell{r, fg, bg})
		}
	}
	return out
}

// ------------------------------------------------- COPIED from live.go / pace.go

// paceSpan is COPIED from pace.go.
func paceSpan(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

type layout struct {
	CatX, PaceSpan, SandFrom, SandTo int
	MoonX                            float64
}

// compose is COPIED from live.go, mirrored branch only (mirror is the shipped
// composition).
func compose(w int, catW int) layout {
	const margin = 2
	right := margin + w/32
	span := paceSpan(w)
	catX := w - catW - right
	if catX < 0 {
		catX = 0
	}
	return layout{CatX: catX, PaceSpan: span, SandFrom: margin, SandTo: catX - 1 - span, MoonX: 0.28}
}

// luma is COPIED from live.go.
func luma(c term.RGB) float64 {
	return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
}

// drawSand is COPIED from live.go. Only the plot positions and the ink matter
// here, and both are kept as they are so the copy cannot flatter the product.
func drawSand(c *canvas.Canvas, lines []reduce.Line, sand term.RGB, sandTop, xFrom, xTo int) {
	if len(lines) == 0 || xTo-xFrom < 12 {
		return
	}
	bad := term.RGB{R: 244, G: 176, B: 96}
	beachAt := func(row int) term.RGB {
		if row < 0 || row >= c.H || len(c.BG) < c.W*c.H {
			return sand
		}
		var r, g, b, n int
		for x := xFrom; x < xTo && x < c.W; x += 4 {
			p := c.BG[row*c.W+x]
			r, g, b, n = r+int(p.R), g+int(p.G), b+int(p.B), n+1
		}
		if n == 0 {
			return sand
		}
		return term.RGB{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n)}
	}
	room := c.H - sandTop
	if room < 1 {
		return
	}
	if len(lines) > room {
		lines = lines[len(lines)-room:]
	}
	top := c.H - len(lines)
	for i, ln := range lines {
		row := top + i
		if row < 0 || row >= c.H {
			continue
		}
		beach := beachAt(row)
		base := term.RGB{R: 244, G: 236, B: 220}
		if luma(beach) > 140 {
			base = term.RGB{R: 34, G: 26, B: 20}
		}
		if ln.Bad {
			base = bad
		}
		col := term.Lerp(base, beach, 0.10+0.62*ln.Age)
		x := xFrom
		for _, r := range []rune(companion.NarrowOnly(ln.Text)) {
			if x >= xTo {
				break
			}
			c.Near().Plot(x, row, r, col, 1)
			x++
		}
	}
}

// ------------------------------------------------------- SMALL 1: the eyes

// The crab's sprite data, COPIED from internal/companion/crab.go so the eye
// cells can be rendered with no eye glyph on top. Validated below against the
// product's own render before anything is claimed from it.
var (
	xCrabLower = []string{
		"########################",
		"########################",
		".######################.",
		"..####################..",
		"...##################...",
		"..##..############..##..",
		"..##..############..##..",
		".##....##########....##.",
		".##....##########....##.",
		"##......########......##",
		"##......########......##",
		"##.......######.......##",
		".#.......######.......#.",
		".#........####........#.",
		".#....................#.",
		".#....................#.",
	}
	xCrabWork = []string{
		"##..##............##..##",
		"##..##............##..##",
		"######............######",
		"######............######",
		"######............######",
		".####..............####.",
		"..##................##..",
		"..##................##..",
		"..##....##....##....##..",
		"..##....##....##....##..",
		"..####################..",
		".######################.",
	}
	xCrabWork2 = []string{
		"........................",
		"##..##............##..##",
		"##..##............##..##",
		"######............######",
		"######............######",
		"######............######",
		"..##................##..",
		"..##................##..",
		"..##....##....##....##..",
		"..##....##....##....##..",
		"..####################..",
		".######################.",
	}
	xCrabRest = []string{
		"........................",
		"........................",
		"........................",
		"##..##............##..##",
		"##..##............##..##",
		"######............######",
		"######............######",
		".####..............####.",
		"..##....##....##....##..",
		"..##....##....##....##..",
		"..####################..",
		".######################.",
	}
	xCrabAsk = []string{
		"..................##..##",
		"..................##..##",
		"..................######",
		"..................######",
		"##..##............######",
		"##..##.............####.",
		"######..............##..",
		".####...............##..",
		"..##....##....##....##..",
		"..##....##....##....##..",
		"..####################..",
		".######################.",
	}
	xCrabDone = []string{
		"##..##............##..##",
		"##..##............##..##",
		"##..##............##..##",
		"######............######",
		"######............######",
		".####..............####.",
		"..##................##..",
		"..##................##..",
		"..##....##....##....##..",
		"..##....##....##....##..",
		"..####################..",
		".######################.",
	}
	xCrabWorried = []string{
		"........................",
		"........................",
		"........................",
		"........................",
		"........##....##........",
		"........##....##........",
		"..####..##....##..####..",
		"..####..##....##..####..",
		"..######..........######",
		"..######..........######",
		"..####################..",
		".######################.",
	}
)

// xCrabUpper is COPIED from crab.go's crabUpper.
func xCrabUpper(st companion.State, t float64) []string {
	switch st {
	case companion.Resting:
		if math.Mod(t, 7.2) < 3.6 {
			return xCrabRest
		}
		return xCrabWork2
	case companion.NeedsYou:
		return xCrabAsk
	case companion.Done:
		return xCrabDone
	case companion.Worried:
		return xCrabWorried
	}
	if math.Mod(t, 2.2) < 1.1 {
		return xCrabWork
	}
	return xCrabWork2
}

// xCrabLift is COPIED from drawCrab's breathing.
func xCrabLift(st companion.State, t float64) int {
	period := 3.6
	switch st {
	case companion.Working:
		period = 2.2
	case companion.NeedsYou:
		period = 1.6
	case companion.Worried:
		period = 1.9
	}
	if math.Sin(t*2*math.Pi/period) > 0.35 {
		return 2
	}
	return 0
}

// xPlotRim is COPIED from kittens.go's plotRim.
func xPlotRim(l *canvas.Layer, rows []string, x, y int) {
	h := len(rows)
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	filled := func(cx, cy int) bool {
		if cy < 0 || cy >= h {
			return false
		}
		r := []rune(rows[cy])
		return cx >= 0 && cx < len(r) && r[cx] != ' '
	}
	for cy := -1; cy <= h; cy++ {
		for cx := -1; cx <= w; cx++ {
			if filled(cx, cy) {
				continue
			}
			adj := false
			for dy := -1; dy <= 1 && !adj; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if filled(cx+dx, cy+dy) {
						adj = true
						break
					}
				}
			}
			if adj {
				l.Plot(x+cx, y+cy, ' ', term.RGB{}, 1)
			}
		}
	}
}

// drawCrabBodyOnly is drawCrab with the two eye plots left out, so the cell the
// eye lands in can be read as the SPRITE authored it.
func drawCrabBodyOnly(l *canvas.Layer, x, y int, t float64, st companion.State, mirror bool) {
	rows := append(append([]string{}, xCrabUpper(st, t)...), xCrabLower...)
	lift := xCrabLift(st, t)
	shifted := make([]string, len(rows))
	for i := range rows {
		if i+lift < len(rows) {
			shifted[i] = rows[i+lift]
		} else {
			shifted[i] = strings.Repeat(".", len(rows[0]))
		}
	}
	b := companion.ParseBitmap(shifted)
	if mirror {
		b = b.Mirrored()
	}
	q := b.ToQuadrant()
	xPlotRim(l, q, x, y)
	(&companion.Sprite{Rows: q, Body: companion.CrabCoat}).Draw(l, x, y)
}

// inkAt counts how many of a cell's EIGHT source pixels the bitmap inks. A cell
// is 2 px wide by 4 px tall: ToQuadrant ORs pairs of rows and then takes 2x2 of
// the result, which is why Size() is W/2 by H/4. Validated below against the
// crab's eyeless render -- 8/8 must come back as a full block, 4/8 as a half.
func inkAt(rows []string, cx, cy int) int {
	n := 0
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 2; dx++ {
			py, px := cy*4+dy, cx*2+dx
			if py < 0 || py >= len(rows) {
				continue
			}
			r := []rune(rows[py])
			if px >= 0 && px < len(r) && r[px] != '.' && r[px] != ' ' {
				n++
			}
		}
	}
	return n
}

func partEyes() {
	const w, h = 124, 51
	fmt.Println("================ SMALL 1 -- SetEyeFill on stalked eyes ================")
	fmt.Printf("geometry: %dx%d, mirrored composition (the shipped one), Profile256\n", w, h)

	fills := []struct{ name, f string }{
		{"none(SHIPS)", companion.EyeFillNone},
		{"coat", companion.EyeFillCoat},
		{"socket", companion.EyeFillSocket},
	}
	states := []struct {
		name string
		s    companion.State
	}{
		{"resting", companion.Resting}, {"working", companion.Working},
		{"needsyou", companion.NeedsYou}, {"done", companion.Done},
		{"worried", companion.Worried},
	}
	const phases = 200

	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		cat := companion.New(animal)
		cat.FaceLeft(true)
		cw, ch := cat.Size()
		lay := compose(w, cw)
		act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
		sh, t := warmShore(9, w, h, act, lay.MoonX)
		c := newCanvas(w, h)
		t += 0.08
		sh.Update(c, t, act)
		top := h - 2 - ch
		// A window two cells wider and one taller than the box, so anything
		// that moves outside the sprite is caught too.
		x0, y0, bw, bh := lay.CatX-2, top-1, cw+4, ch+2

		fmt.Printf("\n-- %s: box %dx%d at (%d,%d), window (%d,%d) %dx%d, %d breathing phases\n",
			animal, cw, ch, lay.CatX, top, x0, y0, bw, bh, phases)
		fmt.Printf("   %-9s %-12s %6s %6s %6s   %s\n", "state", "fill", "min", "med", "max", "cells that ever differ (col,row in box)")

		for _, stt := range states {
			base := make([][]cell, phases)
			for p := 0; p < phases; p++ {
				tp := 0.05 + 0.11*float64(p)
				cat.SetEyeFill(companion.EyeFillNone)
				c.Near().Clear()
				cat.Draw(c.Near(), lay.CatX, top, tp, stt.s)
				base[p] = readBox(c, x0, y0, bw, bh)
			}
			for _, f := range fills[1:] {
				counts := make([]int, phases)
				ever := map[int]int{}
				for p := 0; p < phases; p++ {
					tp := 0.05 + 0.11*float64(p)
					cat.SetEyeFill(f.f)
					c.Near().Clear()
					cat.Draw(c.Near(), lay.CatX, top, tp, stt.s)
					got := readBox(c, x0, y0, bw, bh)
					for i := range got {
						if got[i] != base[p][i] {
							counts[p]++
							ever[i]++
						}
					}
				}
				sort.Ints(counts)
				var keys []int
				for k := range ever {
					keys = append(keys, k)
				}
				sort.Ints(keys)
				var parts []string
				for _, k := range keys {
					parts = append(parts, fmt.Sprintf("(%d,%d)x%d", k%bw-2, k/bw-1, ever[k]))
				}
				fmt.Printf("   %-9s %-12s %6d %6d %6d   %s\n", stt.name, f.name,
					counts[0], counts[phases/2], counts[phases-1], strings.Join(parts, " "))
			}
		}

		// The eye cells themselves, one representative phase per state, read as
		// (rune, fg, bg). The fill is entirely a BACKGROUND change.
		fmt.Printf("\n   the eye cells, resolved (phase t=1.05; box column, row):\n")
		for _, stt := range states {
			for _, f := range fills {
				cat.SetEyeFill(f.f)
				c.Near().Clear()
				cat.Draw(c.Near(), lay.CatX, top, 1.05, stt.s)
				var got []string
				for _, e := range eyeColsFor(animal, cw) {
					r, fg, bg := c.ResolveAt(lay.CatX+e, top+eyeRowFor(animal), term.Profile256)
					got = append(got, fmt.Sprintf("col%2d %s", e, cell{r, fg, bg}))
				}
				fmt.Printf("   %-9s %-12s %s\n", stt.name, f.name, strings.Join(got, " | "))
			}
		}

		// What the sprite authored in those cells, and what the neighbours got.
		if animal == companion.NameCrab {
			crabAuthored(c, sh, lay, top, cw, ch, act, t)
		} else {
			catAuthored(c, lay, top, cw)
		}
	}
	eyeContrast()
	fillReach()
}

// fillReach asks how far the switch reaches: over a WHOLE frame with a litter
// on the beach, how many cells does changing the fill move? If the litter's own
// faces answered to it the count would be far more than two.
func fillReach() {
	w, h := 124, 51
	fmt.Printf("\n-- how far the switch REACHES (whole %dx%d frame, 14 subagents):\n", w, h)
	act := scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		sh, t := shoreFor(w, h)
		frame := func(fill string) []cell {
			cat := companion.New(animal)
			cat.FaceLeft(true)
			cat.SetEyeFill(fill)
			cw, ch := cat.Size()
			lay := compose(w, cw)
			c := newCanvas(w, h)
			sh.Update(c, t+0.08, act)
			top := h - 2 - ch
			cat.Draw(c.Near(), lay.CatX, top, 12.34, companion.Working)
			cat.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, 14, w-1,
				int(float64(h)*0.42)+1, sh.SandTop()-2, 12.34, 1)
			return readBox(c, 0, 0, w, h)
		}
		none := frame(companion.EyeFillNone)
		for _, f := range []struct{ name, f string }{
			{"coat", companion.EyeFillCoat}, {"socket", companion.EyeFillSocket},
		} {
			got := frame(f.f)
			n := 0
			for i := range got {
				if got[i] != none[i] {
					n++
				}
			}
			_, litEyes := litterInkAnimal(animal, w, h, 14)
			fmt.Printf("   %-5s %-7s cells changed: %d   (the litter's own eye cells in the same frame: %d, none of them respond)\n",
				animal, f.name, n, len(litEyes))
		}
	}
}

// eyeColsFor is where the eyes land in the MIRRORED box -- the shipped layout.
func eyeColsFor(animal string, cw int) []int {
	if animal == companion.NameCrab {
		return []int{cw - 1 - 4, cw - 1 - 7} // crabEyeCells {4,7}
	}
	return []int{cw - 1 - 2, cw - 1 - 6} // catEyeCells {2,6}
}

func eyeRowFor(animal string) int {
	if animal == companion.NameCrab {
		return 2 // crabEyeRow
	}
	return 2 // the cat plots its eyes at y+2 too
}

// crabAuthored rebuilds the crab's body with no eye glyph, validates the
// rebuild against the product's render at every non-eye cell, and then reports
// what the sprite had put in the eye cells.
func crabAuthored(c *canvas.Canvas, sh *scape.Shore, lay layout, top, cw, ch int, act scape.Activity, t float64) {
	fmt.Printf("\n   WHAT THE SPRITE AUTHORED IN THE EYE CELL (crab), copy validated per state:\n")
	states := []struct {
		name string
		s    companion.State
	}{
		{"resting", companion.Resting}, {"working", companion.Working},
		{"needsyou", companion.NeedsYou}, {"done", companion.Done},
		{"worried", companion.Worried},
	}
	cat := companion.New(companion.NameCrab)
	cat.FaceLeft(true)
	cat.SetEyeFill(companion.EyeFillNone)
	eyes := eyeColsFor(companion.NameCrab, cw)
	for _, stt := range states {
		const tp = 1.05
		c.Near().Clear()
		cat.Draw(c.Near(), lay.CatX, top, tp, stt.s)
		prod := readBox(c, lay.CatX-2, top-1, cw+4, ch+2)
		c.Near().Clear()
		drawCrabBodyOnly(c.Near(), lay.CatX, top, tp, stt.s, true)
		mine := readBox(c, lay.CatX-2, top-1, cw+4, ch+2)
		bw := cw + 4
		mismatch := 0
		for i := range prod {
			col, row := i%bw-2, i/bw-1
			isEye := row == eyeRowFor(companion.NameCrab) && (col == eyes[0] || col == eyes[1])
			if !isEye && prod[i] != mine[i] {
				mismatch++
			}
		}
		// The authored content of the eye cells, from the eyeless rebuild.
		var got []string
		rows := append(append([]string{}, xCrabUpper(stt.s, tp)...), xCrabLower...)
		for _, e := range eyes {
			r, fg, bg := c.ResolveAt(lay.CatX+e, top+2, term.Profile256)
			// ink coverage in the AUTHORED (unmirrored) frame
			src := cw - 1 - e
			got = append(got, fmt.Sprintf("col%2d %s ink above/at/below=%d/%d/%d of 8",
				e, cell{r, fg, bg}, inkAt(rows, src, 1), inkAt(rows, src, 2), inkAt(rows, src, 3)))
		}
		fmt.Printf("   %-9s non-eye cells differing from the product: %d  |  %s\n",
			stt.name, mismatch, strings.Join(got, " | "))
	}
}

func catAuthored(c *canvas.Canvas, lay layout, top, cw int) {
	fmt.Printf("\n   WHAT THE SPRITE AUTHORED IN THE EYE CELL (cat), from companion.CatBody:\n")
	for _, e := range eyeColsFor(companion.NameCat, cw) {
		src := cw - 1 - e
		fmt.Printf("   box col %2d (authored col %d): ink above/at/below the eye row = %d/%d/%d of 8\n",
			e, src, inkAt(companion.CatBody, src, 1), inkAt(companion.CatBody, src, 2),
			inkAt(companion.CatBody, src, 3))
	}
	fmt.Printf("   the whole eye row, authored ink per cell: ")
	for src := 0; src < cw; src++ {
		fmt.Printf("%d:%d ", src, inkAt(companion.CatBody, src, 2))
	}
	fmt.Println()
}

// eyeContrast prints how legible the eye glyph is on each fill's ground, and
// what is actually behind the head at three hours -- the fill's whole job is to
// stop the answer depending on the hour.
func eyeContrast() {
	fmt.Printf("\n-- what is BEHIND the head, and what each fill puts under the glyph\n")
	w, h := 124, 51
	for _, hr := range []struct {
		name string
		tod  float64
	}{{"11:30", 0.479}, {"18:40", 0.778}, {"22:20", 0.931}} {
		for _, animal := range []string{companion.NameCat, companion.NameCrab} {
			cat := companion.New(animal)
			cat.FaceLeft(true)
			cw, ch := cat.Size()
			lay := compose(w, cw)
			act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.3, TimeOfDay: hr.tod}
			sh, t := warmShore(9, w, h, act, lay.MoonX)
			c := newCanvas(w, h)
			sh.Update(c, t+0.08, act)
			top := h - 2 - ch
			e := eyeColsFor(animal, cw)[0]
			var out []string
			for _, f := range []struct{ name, f string }{
				{"none", companion.EyeFillNone}, {"coat", companion.EyeFillCoat},
				{"socket", companion.EyeFillSocket},
			} {
				cat.SetEyeFill(f.f)
				c.Near().Clear()
				cat.Draw(c.Near(), lay.CatX, top, 1.05, companion.Working)
				_, fg, bg := c.ResolveAt(lay.CatX+e, top+2, term.Profile256)
				out = append(out, fmt.Sprintf("%s bg(%3d,%3d,%3d) dLuma=%5.1f",
					f.name, bg.R, bg.G, bg.B, math.Abs(luma(fg)-luma(bg))))
			}
			fmt.Printf("   %-6s %-5s %s\n", hr.name, animal, strings.Join(out, " | "))
		}
	}
}

// ------------------------------------------------- SMALL 2: sand vs litter

type sample struct {
	kittens int
	lines   []reduce.Line
	when    time.Time
}

func partSand() {
	fmt.Println("\n================ SMALL 2 -- a sand line across the litter ================")

	tideCheck()

	// --- the real event log. Rule 3: count it, do not reason about it.
	dir := filepath.Join(os.Getenv("HOME"), ".config", "xscapes", "run")
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	sort.Strings(files)

	widths := []int{60, 80, 100, 107, 124, 153}
	heights := []int{24, 40, 51}

	// Pass 1 folds every recording once and asks the geometry question for
	// every (width, height) at every sampled second.
	type key struct{ w, h int }
	type acc struct {
		samples, withLitter, withTail, live, collide, cellsHit int
		spanHit, eyeHit, inkCells                              int
		worst                                                  int
		byLine, nlines                                         [8]int
		sumLen, nLen, maxLen                                   int
	}
	res := map[key]*acc{}
	for _, w := range widths {
		for _, h := range heights {
			res[key{w, h}] = &acc{}
		}
	}
	sessions, events, samples := 0, 0, 0
	var kittenHist [64]int
	var lineLen []int
	// keep a few real colliding cases for pass 2
	type probe struct {
		w, h    int
		kittens int
		lines   []reduce.Line
	}
	var probes []probe
	var nonProbes []probe

	for _, f := range files {
		evs := readEvents(f)
		if len(evs) < 20 {
			continue
		}
		sessions++
		events += len(evs)
		r := reduce.New("s28")
		at := func(ms int64) time.Time { return time.UnixMilli(ms) }
		start, end := at(evs[0].TS), at(evs[len(evs)-1].TS)
		i := 0
		tiers := map[key]int{}
		for t := start; !t.After(end); t = t.Add(time.Second) {
			for i < len(evs) && !at(evs[i].TS).After(t) {
				r.Apply(evs[i], at(evs[i].TS))
				i++
			}
			st := r.State(t)
			samples++
			if st.Kittens >= 0 && st.Kittens < len(kittenHist) {
				kittenHist[st.Kittens]++
			}
			for _, k := range keysOf(widths, heights) {
				a := res[key{k.w, k.h}]
				a.samples++
				if st.Kittens > 0 {
					a.withLitter++
				}
				lay := compose(k.w, 12)
				lines := st.FitTail(t, lay.SandTo-lay.SandFrom)
				if len(lines) > 0 {
					a.withTail++
					for _, ln := range lines {
						n := len([]rune(companion.NarrowOnly(ln.Text)))
						a.sumLen += n
						a.nLen++
						if n > a.maxLen {
							a.maxLen = n
						}
					}
				}
				if st.Kittens == 0 || len(lines) == 0 {
					continue
				}
				a.live++
				prev := tiers[key{k.w, k.h}]
				span, tier := collidesSpan(k.w, k.h, st.Kittens, lines, prev)
				tiers[key{k.w, k.h}] = tier
				if span > 0 {
					a.spanHit++
				}
				hit, eye, byLine := collides(k.w, k.h, st.Kittens, lines)
				if hit > 0 {
					a.collide++
					a.cellsHit += hit
					a.eyeHit += eye
					for li, v := range byLine {
						if v > 0 && li < len(a.byLine) {
							a.byLine[li]++
						}
					}
					a.nlines[min(len(lines), 7)]++
					if hit > a.worst {
						a.worst = hit
					}
					if len(probes) < 40 && k.w == 124 && k.h == 51 {
						probes = append(probes, probe{k.w, k.h, st.Kittens, lines})
					}
				} else if len(nonProbes) < 10 && k.w == 124 && k.h == 51 {
					nonProbes = append(nonProbes, probe{k.w, k.h, st.Kittens, lines})
				}
			}
			if len(st.Tail) > 0 {
				for _, l := range st.Tail {
					lineLen = append(lineLen, len([]rune(l.Text)))
				}
			}
			if i < len(evs) {
				if gap := at(evs[i].TS).Sub(t); gap > 3*time.Minute {
					t = at(evs[i].TS).Add(-time.Second)
				}
			}
		}
	}

	fmt.Printf("\nSAMPLE: %d recordings in %s, %d files on disk, %d events, %d one-second samples\n",
		sessions, dir, len(files), events, samples)
	sort.Ints(lineLen)
	if len(lineLen) > 0 {
		fmt.Printf("  real tail lines: %d of them, length p10=%d p50=%d p90=%d max=%d\n",
			len(lineLen), lineLen[len(lineLen)/10], lineLen[len(lineLen)/2],
			lineLen[len(lineLen)*9/10], lineLen[len(lineLen)-1])
	}
	nz := 0
	for k, v := range kittenHist {
		if k > 0 {
			nz += v
		}
	}
	fmt.Printf("  samples with a litter on screen: %d (%.2f%% of all samples) -- the sample is NOT empty\n",
		nz, 100*float64(nz)/float64(max(1, samples)))

	fmt.Printf("\nPASS 1 -- every sampled second, crab companion (his setting), litter INK measured\n")
	fmt.Printf("          off rendered frames; 'box' is the looser bounding-box question.\n")
	fmt.Printf("  %5s %5s %7s %9s %9s %9s %8s %8s %8s %7s %6s\n",
		"w", "h", "budget", "samples", "litter>0", "both", "box", "INK", "%of both", "cells", "eyes")
	for _, h := range heights {
		for _, w := range widths {
			a := res[key{w, h}]
			lay := compose(w, 12)
			fmt.Printf("  %5d %5d %7d %9d %9d %9d %8d %8d %7.2f%% %7d %6d\n", w, h, lay.SandTo-lay.SandFrom,
				a.samples, a.withLitter, a.live, a.spanHit, a.collide,
				100*float64(a.collide)/float64(max(1, a.live)), a.cellsHit, a.eyeHit)
		}
	}
	fmt.Printf("\n  ('both' = litter on screen AND a tail in the sand; %% is of those.)\n")
	fmt.Printf("  worst single frame, ink cells destroyed: ")
	for _, w := range widths {
		fmt.Printf("w%d=%d ", w, res[key{w, 51}].worst)
	}
	fmt.Println()
	fmt.Printf("  mean/max FITTED sand line length, which is what has to reach: ")
	for _, w := range widths {
		a := res[key{w, 51}]
		fmt.Printf("w%d=%.1f/%d ", w, float64(a.sumLen)/float64(max(1, a.nLen)), a.maxLen)
	}
	fmt.Println()
	fmt.Printf("  mean ink cells destroyed per colliding frame: ")
	for _, w := range widths {
		a := res[key{w, 51}]
		fmt.Printf("w%d=%.1f ", w, float64(a.cellsHit)/float64(max(1, a.collide)))
	}
	fmt.Println()
	fmt.Printf("  WHICH line collides (0 = oldest/faintest, last = newest), and how many\n")
	fmt.Printf("  lines the tail had, over colliding frames:\n")
	for _, w := range []int{60, 80, 107, 124} {
		a := res[key{w, 51}]
		fmt.Printf("   w=%-4d line idx %v | tail length %v\n", w, a.byLine[:5], a.nlines[:5])
	}
	fmt.Printf("  %% of ALL samples, which is what \"1.8%% of samples\" would have to mean:\n")
	for _, h := range heights {
		var b strings.Builder
		for _, w := range widths {
			a := res[key{w, h}]
			fmt.Fprintf(&b, " w%d=%.2f%%", w, 100*float64(a.collide)/float64(max(1, a.samples)))
		}
		fmt.Printf("   h=%d:%s\n", h, b.String())
	}

	// --- PASS 2: the rendered frame.
	fmt.Printf("\nPASS 2 -- RENDERED frames, %d colliding cases + %d clean controls at 124x51:\n",
		len(probes), len(nonProbes))
	fmt.Printf("  %4s %8s %9s %10s %10s %9s %9s\n",
		"case", "kittens", "litterpx", "sandcells", "overlap", "sandwins", "litterwins")
	totOver, totSand := 0, 0
	for n, p := range append(append([]probe{}, probes...), nonProbes...) {
		over, sandWon, litWon, litCells, sandCells, detail := renderProbe(p.w, p.h, p.kittens, p.lines)
		totOver += over
		totSand += sandWon
		tag := "hit"
		if n >= len(probes) {
			tag = "ctl"
		}
		fmt.Printf("  %4s %8d %9d %10d %10d %9d %9d  %s\n",
			tag, p.kittens, litCells, sandCells, over, sandWon, litWon, detail)
		if n > 14 && n < len(probes) {
			continue
		}
	}
	fmt.Printf("\n  total overlapping cells %d, sand won %d of them (%.0f%%)\n",
		totOver, totSand, 100*float64(totSand)/float64(max(1, totOver)))

	// --- why "eyes lost 0" is not an empty sample.
	fmt.Printf("\nWHY NO EYE IS EVER LOST (the mask is not empty -- it is counted):\n")
	fmt.Printf("  %5s %8s %8s %8s %10s %12s %12s\n",
		"w", "kittens", "inkcells", "eyecells", "eyerow", "litter rows", "sand rows")
	for _, n := range []int{2, 5, 8, 14, 30} {
		ink, eyes := litterInk(124, 51, n, 0)
		r0, r1 := 99, -1
		er0, er1 := 99, -1
		for c := range ink {
			if c[1] < r0 {
				r0 = c[1]
			}
			if c[1] > r1 {
				r1 = c[1]
			}
		}
		for c := range eyes {
			if c[1] < er0 {
				er0 = c[1]
			}
			if c[1] > er1 {
				er1 = c[1]
			}
		}
		// the sand's rows for a full four-line tail
		st := sandTopFor(124, 51)
		room := 51 - st
		nl := 4
		if nl > room {
			nl = room
		}
		fmt.Printf("  %5d %8d %8d %8d %6d..%-3d %6d..%-5d %6d..%-5d\n",
			124, n, len(ink), len(eyes), er0, er1, r0, r1, 51-nl, 50)
	}

	// --- what the obvious fix would cost.
	fmt.Printf("\nCOST OF RESERVING THE LITTER'S SPAN THE WAY PaceSpan RESERVES THE PACE STRIP:\n")
	fmt.Printf("  the litter's leftmost INK column on the sand's rows, per litter size.\n")
	fmt.Printf("  %8s %9s %13s %13s %12s\n", "kittens", "share", "leftmost@124", "budget left", "leftmost@80")
	for _, n := range []int{1, 2, 4, 6, 8, 10, 14, 20, 30, 38} {
		l124 := leftmostInk(124, 51, n)
		l80 := leftmostInk(80, 51, n)
		share := 0.0
		if n < len(kittenHist) {
			share = 100 * float64(kittenHist[n]) / float64(max(1, samples))
		}
		fmt.Printf("  %8d %8.2f%% %13d %13d %12d\n", n, share, l124, l124-1-2, l80)
	}
	cum := 0
	for n := 1; n < len(kittenHist); n++ {
		cum += kittenHist[n]
	}
	fmt.Printf("  (share is of all %d samples; %d had a litter at all. -1 = the litter\n", samples, cum)
	fmt.Printf("   puts no ink on the sand's rows at that size.)\n")

	// --- does the CAT's litter sit anywhere different? His companion is the
	// crab, but the cat still ships, and a fix has to cover both.
	fmt.Printf("\nTHE CAT'S LITTER, same question (124x51):\n")
	fmt.Printf("  %6s %8s %10s %12s %12s\n", "animal", "kittens", "inkcells", "rows", "leftmost")
	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		for _, n := range []int{2, 8, 14} {
			ink := litterInkFor(animal, 124, 51, n)
			r0, r1, lm := 99, -1, 999
			for c := range ink {
				if c[1] > 40 {
					if c[1] < r0 {
						r0 = c[1]
					}
					if c[1] > r1 {
						r1 = c[1]
					}
					if c[0] < lm {
						lm = c[0]
					}
				}
			}
			fmt.Printf("  %6s %8d %10d %6d..%-5d %12d\n", animal, n, len(ink), r0, r1, lm)
		}
	}

	// --- a worked picture of the worst case, so the defect is legible.
	fmt.Printf("\nWORST CASE PICTURE (124x51, the litter's rows, before and after the sand):\n")
	worstPicture()
}

// leftmostInk is the litter's leftmost inked column on the rows the sand can
// reach -- the column a reservation would have to stop at.
func leftmostInk(w, h, n int) int {
	ink, _ := litterInk(w, h, n, 0)
	st := sandTopFor(w, h)
	room := h - st
	nl := 4
	if nl > room {
		nl = room
	}
	first := h - nl
	best := -1
	for c := range ink {
		if c[1] < first {
			continue
		}
		if best < 0 || c[0] < best {
			best = c[0]
		}
	}
	return best
}

// tideCheck asks whether the answer depends on the tide. The sand's rows are
// min(len(tail), H-SandTop) counted up from the bottom, and SandTop MOVES with
// the tide, which is now the default -- so if a high tide left room for fewer
// than three lines the collision could not happen at all.
func tideCheck() {
	fmt.Printf("\nDOES THE TIDE CHANGE THE ANSWER? (Tide=%v) rows the sand gets, by level:\n", scape.Tide)
	fmt.Printf("  %5s %8s %8s %8s %8s %8s\n", "h", "lvl0.0", "lvl0.25", "lvl0.5", "lvl0.75", "lvl1.0")
	for _, h := range []int{24, 40, 51} {
		fmt.Printf("  %5d", h)
		for _, lv := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
			act := scape.Activity{Working: true, Level: lv, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
			sh, t := warmShore(1, 124, h, act, 0.28)
			c := newCanvas(124, h)
			sh.Update(c, t+0.08, act)
			fmt.Printf(" %8d", h-sh.SandTop())
		}
		fmt.Println()
	}
	fmt.Printf("  (a tail is at most %d lines; three or more are needed to reach the litter)\n", reduce.TailLen)
}

func keysOf(ws, hs []int) []struct{ w, h int } {
	var out []struct{ w, h int }
	for _, w := range ws {
		for _, h := range hs {
			out = append(out, struct{ w, h int }{w, h})
		}
	}
	return out
}

func readEvents(path string) []event.Event {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var evs []event.Event
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		e, err := event.Decode(line)
		if err != nil || e.TS == 0 {
			continue
		}
		evs = append(evs, e)
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].TS < evs[j].TS })
	return evs
}

// litterInk is the litter's INK, measured off a rendered frame: the cells
// DrawKittens changes that are not the blanking rim it paints around each
// sprite. Keyed by (w,h,n,tier) and cached, because a run over a quarter of a
// million samples cannot render one frame each.
//
// This is what makes pass 1 a measurement of the picture rather than of a
// bounding box: a sprite's span includes the gaps between its legs, and a sand
// glyph landing in one of those is not a defect.
type inkKey struct{ w, h, n, tier int }

var (
	inkCache   = map[inkKey]map[[2]int]bool{}
	eyeCache   = map[inkKey]map[[2]int]bool{}
	shoreCache = map[[2]int]*scape.Shore{}
	shoreT     = map[[2]int]float64{}
)

func shoreFor(w, h int) (*scape.Shore, float64) {
	if s, ok := shoreCache[[2]int{w, h}]; ok {
		return s, shoreT[[2]int{w, h}]
	}
	act := scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	sh, t := warmShore(1, w, h, act, 0.28)
	shoreCache[[2]int{w, h}] = sh
	shoreT[[2]int{w, h}] = t
	return sh, t
}

func litterInkFor(animal string, w, h, n int) map[[2]int]bool {
	ink, _ := litterInkAnimal(animal, w, h, n)
	return ink
}

func litterInk(w, h, n, tier int) (ink, eyes map[[2]int]bool) {
	return litterInkAnimal(companion.NameCrab, w, h, n)
}

func litterInkAnimal(animal string, w, h, n int) (ink, eyes map[[2]int]bool) {
	tier := 0
	if animal == companion.NameCat {
		tier = 1
	}
	k := inkKey{w, h, n, tier}
	if v, ok := inkCache[k]; ok {
		return v, eyeCache[k]
	}
	act := scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	sh, t := shoreFor(w, h)
	cat := companion.New(animal)
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	lay := compose(w, cw)
	top := h - 2 - ch
	const tp = 12.34
	// base: shore + parent. The litter is the only difference.
	a := newCanvas(w, h)
	sh.Update(a, t+0.08, act)
	cat.Draw(a.Near(), lay.CatX, top, tp, companion.Working)
	base := readBox(a, 0, 0, w, h)
	b := newCanvas(w, h)
	sh.Update(b, t+0.08, act)
	// A FRESH companion, so the size ladder is TierFor(sitters, 0). The live
	// scape carries its last rung for hysteresis; ignoring that can only move
	// a sample one rung at the 4/6 and 8/10 boundaries, and it is stated
	// rather than hidden.
	cat2 := companion.New(animal)
	cat2.FaceLeft(true)
	cat2.Draw(b.Near(), lay.CatX, top, tp, companion.Working)
	cat2.DrawKittens(b.Near(), b.Mid(), lay.CatX-lay.PaceSpan, top, n, w-1,
		int(float64(h)*0.42)+1, sh.SandTop()-2, tp, 1)
	got := readBox(b, 0, 0, w, h)
	ink, eyes = map[[2]int]bool{}, map[[2]int]bool{}
	for i := range base {
		if got[i] == base[i] || got[i].r == ' ' {
			continue
		}
		ink[[2]int{i % w, i / w}] = true
		if got[i].r == 'o' || got[i].r == '-' {
			eyes[[2]int{i % w, i / w}] = true
		}
	}
	inkCache[k], eyeCache[k] = ink, eyes
	return ink, eyes
}

// tierSeed is a sitter count that lands the ladder on the wanted rung.
func tierSeed(tier int) int {
	switch tier {
	case 1:
		return 12
	case 2:
		return 30
	}
	return 1
}

// crabletSpans is the sitters' cell columns, from crab_kittens.go's arithmetic.
// Returns the spans [x, x+cw) and the rows they occupy. Validated in pass 2
// against a real DrawKittens render.
func crabletSpans(w, h, n, prevTier int) (spans [][2]int, row0, rowN, tier int) {
	lay := compose(w, 12)
	px := lay.CatX - lay.PaceSpan
	py := h - 2 - 7
	var sitters int
	for i := 0; i < n; i++ {
		if !(i > 0 && companion.HashF(i, 11, 1) > 0.64) { // swims(), seed 1
			sitters++
		}
	}
	tier = companion.TierFor(sitters, prevTier)
	rows := [][]string{companion.Crablet, companion.CrabletSmall, companion.CrabletTiny}[tier]
	cw, chh := len(rows[0])/2, len(rows)/4
	ky := py + 7 - chh
	for k := 0; k < sitters; k++ {
		x := px - 1 - cw - k*(cw+1)
		if x < 0 {
			break
		}
		spans = append(spans, [2]int{x, x + cw})
	}
	return spans, ky, ky + chh, tier
}

// collides asks whether any sand glyph lands on a cell of the litter's INK,
// using the mask measured off a rendered frame.
func collides(w, h, kittens int, lines []reduce.Line) (inkHit, eyeHit int, byLine [8]int) {
	lay := compose(w, 12)
	ink, eyes := litterInk(w, h, kittens, 0)
	if len(ink) == 0 {
		return 0, 0, byLine
	}
	room := h - sandTopFor(w, h)
	if room < 1 {
		return 0, 0, byLine
	}
	ls := lines
	if len(ls) > room {
		ls = ls[len(ls)-room:]
	}
	top := h - len(ls)
	for i, ln := range ls {
		row := top + i
		n := len([]rune(companion.NarrowOnly(ln.Text)))
		for x := lay.SandFrom; x < lay.SandFrom+n && x < lay.SandTo; x++ {
			if ink[[2]int{x, row}] {
				inkHit++
				if i < len(byLine) {
					byLine[i]++
				}
				if eyes[[2]int{x, row}] {
					eyeHit++
				}
			}
		}
	}
	return inkHit, eyeHit, byLine
}

// collidesSpan is the looser question: does the sand reach a crablet's BOUNDING
// BOX at all, the gaps between its legs and its blanking rim included.
func collidesSpan(w, h, kittens int, lines []reduce.Line, prevTier int) (cells, tier int) {
	lay := compose(w, 12)
	spans, row0, rowN, tier := crabletSpans(w, h, kittens, prevTier)
	if len(spans) == 0 {
		return 0, tier
	}
	// the sand's rows, exactly as drawSand places them
	room := h - sandTopFor(w, h)
	if room < 1 {
		return 0, tier
	}
	ls := lines
	if len(ls) > room {
		ls = ls[len(ls)-room:]
	}
	top := h - len(ls)
	for i, ln := range ls {
		row := top + i
		if row < row0 || row >= rowN {
			continue
		}
		n := len([]rune(companion.NarrowOnly(ln.Text)))
		x0, x1 := lay.SandFrom, lay.SandFrom+n
		if x1 > lay.SandTo {
			x1 = lay.SandTo
		}
		for _, s := range spans {
			lo, hi := max(x0, s[0]), min(x1, s[1])
			if hi > lo {
				cells += hi - lo
			}
		}
	}
	return cells, tier
}

// sandTopFor renders one frame and asks the shore where the sand starts. The
// waterline moves with the tide, so this is read off the shore, not derived.
var sandTopCache = map[[2]int]int{}

func sandTopFor(w, h int) int {
	if v, ok := sandTopCache[[2]int{w, h}]; ok {
		return v
	}
	act := scape.Activity{Working: true, Level: 0.5, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	sh, t := warmShore(1, w, h, act, 0.28)
	c := newCanvas(w, h)
	sh.Update(c, t+0.08, act)
	v := sh.SandTop()
	sandTopCache[[2]int{w, h}] = v
	return v
}

// renderProbe paints the real thing three times -- shore, +litter, +sand -- and
// reads the difference back out of the RESOLVED cells.
func renderProbe(w, h, kittens int, lines []reduce.Line) (overlap, sandWon, litterWon, litCells, sandCells int, detail string) {
	act := scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	cat := companion.New(companion.NameCrab)
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	lay := compose(w, cw)
	sh, t := warmShore(1, w, h, act, lay.MoonX)
	top := h - 2 - ch
	const tp = 12.34

	// (a) shore only
	a := newCanvas(w, h)
	sh.Update(a, t+0.08, act)
	base := readBox(a, 0, 0, w, h)

	// (b) shore + companion + litter
	b := newCanvas(w, h)
	sh.Update(b, t+0.08, act)
	litterX := lay.CatX - lay.PaceSpan
	cat.Draw(b.Near(), lay.CatX, top, tp, companion.Working)
	cat.DrawKittens(b.Near(), b.Mid(), litterX, top, kittens, w-1,
		int(float64(h)*0.42)+1, sh.SandTop()-2, tp, 1)
	withLit := readBox(b, 0, 0, w, h)

	// (c) the same frame with the sand written over it
	cat2 := companion.New(companion.NameCrab)
	cat2.FaceLeft(true)
	cc := newCanvas(w, h)
	sh.Update(cc, t+0.08, act)
	cat2.Draw(cc.Near(), lay.CatX, top, tp, companion.Working)
	cat2.DrawKittens(cc.Near(), cc.Mid(), litterX, top, kittens, w-1,
		int(float64(h)*0.42)+1, sh.SandTop()-2, tp, 1)
	drawSand(cc, lines, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
	withSand := readBox(cc, 0, 0, w, h)

	// The litter's INK: cells DrawKittens changed that are not the blanking
	// rim (a space) it paints around every sprite. A sand glyph landing on a
	// rim cell is not a defect; one landing on ink is.
	litter := map[int]bool{}
	eyes := map[int]bool{}
	for i := range base {
		if withLit[i] == base[i] || withLit[i].r == ' ' {
			continue
		}
		litter[i] = true
		litCells++
		if withLit[i].r == 'o' || withLit[i].r == '-' {
			eyes[i] = true
		}
	}
	// Who WINS the cell: compare the final rune against the rune the sand
	// asked for at that column, rather than merely noting that the cell
	// changed -- "it changed" cannot tell the two answers apart.
	eyesLost := 0
	room := h - sh.SandTop()
	ls := lines
	if room >= 1 && len(ls) > room {
		ls = ls[len(ls)-room:]
	}
	sandTopRow := h - len(ls)
	for li, ln := range ls {
		row := sandTopRow + li
		if row < 0 || row >= h {
			continue
		}
		rs := []rune(companion.NarrowOnly(ln.Text))
		for k, r := range rs {
			x := lay.SandFrom + k
			if x >= lay.SandTo || x >= w {
				break
			}
			i := row*w + x
			sandCells++
			if !litter[i] {
				continue
			}
			overlap++
			switch withSand[i].r {
			case r:
				sandWon++
				if eyes[i] {
					eyesLost++
				}
			case withLit[i].r:
				litterWon++
			}
		}
	}
	if overlap > 0 {
		for i := range withLit {
			if litter[i] && withSand[i] != withLit[i] {
				detail = fmt.Sprintf("first ink at (%d,%d): %q -> %q, eyes lost %d",
					i%w, i/w, string(withLit[i].r), string(withSand[i].r), eyesLost)
				break
			}
		}
	}
	return overlap, sandWon, litterWon, litCells, sandCells, detail
}

// worstPicture prints the litter's rows as text, with and without the sand, at
// a tail long enough to reach -- so what the defect LOOKS like is on the record.
func worstPicture() {
	w, h := 124, 51
	act := scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}
	cat := companion.New(companion.NameCrab)
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	lay := compose(w, cw)
	sh, t := warmShore(1, w, h, act, lay.MoonX)
	top := h - 2 - ch
	const tp = 12.34
	budget := lay.SandTo - lay.SandFrom
	lines := []reduce.Line{
		{Text: strings.Repeat("x", budget), Age: 0.8},
		{Text: "edit internal/companion/crab_kittens.go   +18 -2", Age: 0.5},
		{Text: strings.Repeat("y", budget), Age: 0.3},
		{Text: "bash go test ./internal/...   ok 0.412s", Age: 0.0},
	}
	render := func(withSand bool) []cell {
		c := newCanvas(w, h)
		sh.Update(c, t+0.08, act)
		cat.Draw(c.Near(), lay.CatX, top, tp, companion.Working)
		cat.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, 8, w-1,
			int(float64(h)*0.42)+1, sh.SandTop()-2, tp, 1)
		if withSand {
			drawSand(c, lines, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
		}
		return readBox(c, 0, 0, w, h)
	}
	before, after := render(false), render(true)
	fmt.Printf("  sand columns %d..%d, litter starts at %d, %d rows shown from row %d\n",
		lay.SandFrom, lay.SandTo-1, lay.CatX-lay.PaceSpan, 6, h-6)
	for row := h - 6; row < h; row++ {
		var b1, b2 strings.Builder
		for x := 40; x < w; x++ {
			b1.WriteRune(before[row*w+x].r)
			b2.WriteRune(after[row*w+x].r)
		}
		fmt.Printf("  r%02d before |%s|\n", row, b1.String())
		fmt.Printf("  r%02d after  |%s|\n", row, b2.String())
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
