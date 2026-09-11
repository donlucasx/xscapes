// Command critic renders all four s28 candidates through the SAME pipeline the
// product uses (companion.ParseBitmap -> ToQuadrant) and measures them against
// each other and against the shipped crab. Nothing here is taken on trust from
// a draft: every number is recomputed and every silhouette redrawn.
//
//	go run ./notes/s28-closer/critic
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

func join(a, b []string) []string { return append(append([]string{}, a...), b...) }

func render(rows []string) []string {
	return companion.ParseBitmap(rows).ToQuadrant()
}
func renderM(rows []string) []string {
	return companion.ParseBitmap(rows).Mirrored().ToQuadrant()
}

// stamp overlays a sprite's rows onto a base at (cx, cy), replacing whole cells
// exactly as canvas.Plot does.
func stamp(base []string, over []string, cx, cy int) []string {
	out := append([]string(nil), base...)
	for dy, r := range over {
		y := cy + dy
		if y < 0 || y >= len(out) {
			continue
		}
		row := []rune(out[y])
		for dx, ch := range []rune(r) {
			x := cx + dx
			if x >= 0 && x < len(row) {
				row[x] = ch
			}
		}
		out[y] = string(row)
	}
	return out
}

func inkCells(q []string) int {
	n := 0
	for _, r := range q {
		for _, c := range r {
			if c != ' ' {
				n++
			}
		}
	}
	return n
}

func glyphHist(q []string) string {
	m := map[rune]int{}
	for _, r := range q {
		for _, c := range r {
			if c != ' ' {
				m[c]++
			}
		}
	}
	type kv struct {
		r rune
		n int
	}
	var v []kv
	for r, n := range m {
		v = append(v, kv{r, n})
	}
	sort.Slice(v, func(i, j int) bool { return v[i].n > v[j].n })
	var sb strings.Builder
	for _, x := range v {
		fmt.Fprintf(&sb, "%c%d ", x.r, x.n)
	}
	return fmt.Sprintf("%d glyphs: %s", len(v), sb.String())
}

// edgeCells counts cells carrying a PARTIAL quadrant -- the only place this
// medium has curvature. A full block is not an edge; a corner or half is.
func edgeCells(q []string) int {
	n := 0
	for _, r := range q {
		for _, c := range r {
			if c != ' ' && c != '█' {
				n++
			}
		}
	}
	return n
}

func srcBBox(rows []string) (w, h int) {
	b := companion.ParseBitmap(rows)
	x0, y0, x1, y1 := b.W, b.H, -1, -1
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if b.On[y*b.W+x] {
				if x < x0 {
					x0 = x
				}
				if x > x1 {
					x1 = x
				}
				if y < y0 {
					y0 = y
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	return x1 - x0 + 1, y1 - y0 + 1
}

func sideBySide(cols [][]string, heads []string, gap int, bottomAlign bool) string {
	h := 0
	for _, c := range cols {
		if len(c) > h {
			h = len(c)
		}
	}
	widths := make([]int, len(cols))
	for i, c := range cols {
		for _, r := range c {
			if n := len([]rune(r)); n > widths[i] {
				widths[i] = n
			}
		}
		if n := len(heads[i]); n > widths[i] {
			widths[i] = n
		}
	}
	var sb strings.Builder
	for i, hd := range heads {
		sb.WriteString(hd + strings.Repeat(" ", widths[i]-len(hd)+gap))
	}
	sb.WriteString("\n")
	for y := 0; y < h; y++ {
		for i, c := range cols {
			var row string
			var idx int
			if bottomAlign {
				idx = y - (h - len(c))
			} else {
				idx = y
			}
			if idx >= 0 && idx < len(c) {
				row = c[idx]
			}
			sb.WriteString(row + strings.Repeat(" ", widths[i]-len([]rune(row))+gap))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func sandTopAt(w, h int, level float64) int {
	sh := scape.NewShore(7, false)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear), tm,
			scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	return sh.SandTop()
}

type cand struct {
	key          string
	upper, lower []string
	eye          []string
	eyeCells     []int
	eyeRow       int
	eyeIsGlyph   bool
	glyph        rune
	glyphCells   []int
	glyphRow     int
}

func main() {
	ship := join(crabAskRows(), crabLowerRows())
	shipWork := join(crabWorkRows(), crabLowerRows())

	cands := []cand{
		{key: "SHIPPED", upper: crabAskRows(), lower: crabLowerRows(),
			eyeIsGlyph: true, glyph: 'O', glyphCells: []int{4, 7}, glyphRow: 2},
		{key: "A step", upper: aUpper, lower: aLower, eye: aEye, eyeCells: []int{4, 8}, eyeRow: 2},
		{key: "B push", upper: bUpper, lower: bLower, eye: bEye, eyeCells: []int{5, 9}, eyeRow: 2},
		{key: "C lean", upper: cUpper, lower: cLower,
			eyeIsGlyph: true, glyph: 'O', glyphCells: []int{3, 8}, glyphRow: 1},
		{key: "D 2x", upper: dUpper, lower: dLower, eye: dEye, eyeCells: []int{8, 14}, eyeRow: 3},
	}

	fmt.Println("=== 1. GEOMETRY. Source pixels are SQUARE on screen (bitmap.go: a cell is")
	fmt.Println("    2px wide by 4px tall, and a terminal cell is ~1:2), so the source bbox")
	fmt.Println("    IS the on-screen shape. A true push-in keeps the shipped 24:28 ratio.")
	fmt.Println()
	sw, sh0 := srcBBox(shipWork)
	fmt.Printf("%-9s %9s %8s %8s %8s %8s  %s\n", "cand", "cells", "src px", "xscale", "yscale", "aspect", "verdict")
	for _, c := range cands {
		rows := join(c.upper, c.lower)
		q := render(rows)
		w, h := srcBBox(rows)
		asp := float64(w) / float64(h)
		shipAsp := float64(sw) / float64(sh0)
		v := "UNIFORM push-in"
		r := asp / shipAsp
		switch {
		case r > 1.06:
			v = fmt.Sprintf("SQUAT: %.0f%% wider than a push-in", (r-1)*100)
		case r < 0.94:
			v = fmt.Sprintf("STRETCHED TALL: %.0f%% taller than a push-in", (1/r-1)*100)
		}
		fmt.Printf("%-9s %9s %8s %8.2f %8.2f %8.3f  %s\n", c.key,
			fmt.Sprintf("%dx%d", len([]rune(q[0])), len(q)),
			fmt.Sprintf("%dx%d", w, h),
			float64(w)/float64(sw), float64(h)/float64(sh0), asp, v)
	}

	fmt.Println("\n=== 2. THE SILHOUETTES, mirrored as shipped, BOTTOM-ALIGNED (all standing on")
	fmt.Println("    the same sand). This is the only fair way to ask 'does it read as closer'.")
	fmt.Println()
	var cols [][]string
	var heads []string
	for _, c := range cands {
		q := renderM(join(c.upper, c.lower))
		q = applyEyes(c, q, true)
		cols = append(cols, q)
		heads = append(heads, c.key)
	}
	fmt.Println(sideBySide(cols, heads, 3, true))

	fmt.Println("=== 3. THE SAME FIVE, TOP-ALIGNED (head held still, growth going DOWN).")
	fmt.Println("    Candidate A is the only one that proposes this anchoring.")
	fmt.Println()
	fmt.Println(sideBySide(cols, heads, 3, false))

	fmt.Println("=== 4. EDGE RESOLUTION. Roundness in this medium lives entirely in the")
	fmt.Println("    PARTIAL quadrants. A full block carries no curve.")
	fmt.Println()
	fmt.Printf("%-9s %6s %6s %7s  %s\n", "cand", "ink", "edge", "edge%", "glyph histogram")
	for _, c := range cands {
		q := render(join(c.upper, c.lower))
		ink, e := inkCells(q), edgeCells(q)
		fmt.Printf("%-9s %6d %6d %6.0f%%  %s\n", c.key, ink, e, 100*float64(e)/float64(ink), glyphHist(q))
	}

	fmt.Println("\n=== 5. EACH CANDIDATE'S EYE, drawn three ways: body alone, body + eye pass,")
	fmt.Println("    and the eye pass shown as E so it is findable in a colourless transcript.")
	for _, c := range cands {
		fmt.Printf("\n--- %s\n", c.key)
		q := render(join(c.upper, c.lower))
		e1 := applyEyes(c, q, false)
		e2 := applyEyesMark(c, q)
		fmt.Println(sideBySide([][]string{q, e1, e2}, []string{"body", "+eye", "+eye as E"}, 3, false))
	}

	fmt.Println("=== 6. CANDIDATE A's PUPIL PASS. canvas.Plot REPLACES the cell (canvas.go:53)")
	fmt.Println("    and Sprite.Draw calls Plot, so a second pass does not composite.")
	fmt.Println()
	aq := render(join(aUpper, aLower))
	eyeQ := render(aEye)
	pupQ := render(aPupil)
	a1 := stamp(stamp(aq, eyeQ, 4, 2), eyeQ, 8, 2)
	a2 := stamp(stamp(a1, pupQ, 4, 2), pupQ, 8, 2)
	fmt.Println(sideBySide([][]string{aq, a1, a2}, []string{"body", "+eye", "+eye+pupil"}, 3, false))
	fmt.Printf("    eye bitmap alone: %v      pupil bitmap alone: %v\n", eyeQ, pupQ)

	fmt.Println("\n=== 7. CANDIDATE C's MIDDLE BEAT. Their frames[2] puts 1px stalks at source")
	fmt.Println("    cols 8 and 15 (cell cols 4 and 7); their text puts the eyes at cells 4 and 7.")
	fmt.Println()
	cq := renderM(join(cOpen, cLower))
	cq4 := stamp(stamp(cq, []string{"o"}, 4, 1), []string{"o"}, 7, 1)
	cqFinal := renderM(join(cUpper, cLower))
	cqF := stamp(stamp(cqFinal, []string{"O"}, 3, 1), []string{"O"}, 8, 1)
	shipQ := renderM(shipWork)
	shipQ = stamp(stamp(shipQ, []string{"o"}, 4, 2), []string{"o"}, 7, 2)
	fmt.Println(sideBySide([][]string{shipQ, cq4, cqF},
		[]string{"shipped work", "C beat 2", "C beat 3 (ask)"}, 4, false))

	fmt.Println("=== 8. THE ROOM, remeasured on a real settled Shore at the median ask level")
	fmt.Println("    (0.59). A companion drawn at top = H-2-chh needs chh+2 dry rows to be")
	fmt.Println("    fully out of the water. 'wet' counts rows of the sprite BOX over the sea.")
	fmt.Println()
	geoms := []struct{ w, h int }{{40, 12}, {80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}}
	fmt.Printf("%-9s %5s", "geom", "dry")
	for _, c := range cands {
		fmt.Printf(" %11s", c.key)
	}
	fmt.Println("      (rows of the box over the sea, shipped anchor)")
	for _, g := range geoms {
		st := sandTopAt(g.w, g.h, 0.59)
		dry := g.h - st
		fmt.Printf("%-9s %5d", fmt.Sprintf("%dx%d", g.w, g.h), dry)
		for _, c := range cands {
			q := render(join(c.upper, c.lower))
			chh := len(q)
			top := g.h - 2 - chh
			wet := st - top
			if wet < 0 {
				wet = 0
			}
			fmt.Printf(" %11s", fmt.Sprintf("%d of %d", wet, chh))
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("    Same table with the feet spent DOWN into both free rows (top = H-1-chh-?)")
	fmt.Println("    -- i.e. standing on the LAST drawn row, which is A's and B's fallback:")
	for _, g := range geoms {
		st := sandTopAt(g.w, g.h, 0.59)
		fmt.Printf("%-9s %5d", fmt.Sprintf("%dx%d", g.w, g.h), g.h-st)
		for _, c := range cands {
			q := render(join(c.upper, c.lower))
			chh := len(q)
			top := g.h - chh
			wet := st - top
			if wet < 0 {
				wet = 0
			}
			fmt.Printf(" %11s", fmt.Sprintf("%d of %d", wet, chh))
		}
		fmt.Println()
	}

	fmt.Println("\n=== 9. THE SAND BILL. compose(): catX = w - catW - (2 + w/32);")
	fmt.Println("    SandTo = catX - 1 - paceSpan(w); SandFrom = 2. Columns of writing band:")
	fmt.Println()
	fmt.Printf("%-9s", "geom")
	for _, c := range cands {
		q := render(join(c.upper, c.lower))
		fmt.Printf(" %10s", fmt.Sprintf("%s(%d)", c.key[:1], len([]rune(q[0]))))
	}
	fmt.Println()
	for _, g := range geoms {
		fmt.Printf("%-9s", fmt.Sprintf("%dx%d", g.w, g.h))
		base := 0
		for i, c := range cands {
			q := render(join(c.upper, c.lower))
			catW := len([]rune(q[0]))
			span := g.w / 16
			if span > 6 {
				span = 6
			}
			if span < 1 {
				span = 1
			}
			catX := g.w - catW - (2 + g.w/32)
			if catX < 0 {
				catX = 0
			}
			band := (catX - 1 - span) - 2 + 1
			if band < 0 {
				band = 0
			}
			if i == 0 {
				base = band
			}
			d := ""
			if i > 0 {
				d = fmt.Sprintf("%+d", band-base)
			}
			fmt.Printf(" %10s", fmt.Sprintf("%d%s", band, d))
		}
		fmt.Println()
	}

	fmt.Println("\n=== 10. THE LITTER. crab_kittens.go:135 is `ky := py + 7 - ch` -- the parent's")
	fmt.Println("    height is the LITERAL 7. Where the crablets' feet land against the parent's:")
	fmt.Println()
	for _, c := range cands {
		q := render(join(c.upper, c.lower))
		chh := len(q)
		for _, ch := range []int{4, 3} {
			ky := 0 + 7 - ch
			off := (ky + ch) - chh
			fmt.Printf("    %-9s parent %2d rows, crablet %d rows: crablet feet %+d rows vs parent's\n",
				c.key, chh, ch, off)
		}
	}

	fmt.Println("\n=== 11. HOW LONG IS AN ASK OPEN? (his real recordings)")
	askDurations()

	fmt.Println("\n=== 12. THE LEVEL DURING AN OPEN ASK -- nobody measured this. The tide is IN\n    at the instant the ask fires and then withdraws while the crab stands there.")
	fmt.Println()
	askLevels()

	fmt.Println()
	ladder()
	pickShow()
	eyeVocab()
	scenes()

	_ = ship
}

func applyEyes(c cand, q []string, mirror bool) []string {
	out := append([]string(nil), q...)
	w := len([]rune(q[0]))
	if c.eyeIsGlyph {
		for _, e := range c.glyphCells {
			x := e
			if mirror {
				x = w - 1 - e
			}
			out = stamp(out, []string{string(c.glyph)}, x, c.glyphRow)
		}
		return out
	}
	eq := render(c.eye)
	ew := len([]rune(eq[0]))
	for _, e := range c.eyeCells {
		x := e
		if mirror {
			x = w - ew - e
		}
		out = stamp(out, eq, x, c.eyeRow)
	}
	return out
}

func applyEyesMark(c cand, q []string) []string {
	out := append([]string(nil), q...)
	if c.eyeIsGlyph {
		for _, e := range c.glyphCells {
			out = stamp(out, []string{"E"}, e, c.glyphRow)
		}
		return out
	}
	eq := render(c.eye)
	var mk []string
	for _, r := range eq {
		mk = append(mk, strings.Map(func(x rune) rune {
			if x == ' ' {
				return ' '
			}
			return 'E'
		}, r))
	}
	for _, e := range c.eyeCells {
		out = stamp(out, mk, e, c.eyeRow)
	}
	return out
}
