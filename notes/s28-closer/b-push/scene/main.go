// Command scene puts direction B's crab into a REAL shore at his real
// geometries and at the real level the ask fires at, and prints where the
// waterline crosses it.
//
// The room tables count rows of beach against the companion's BOX height, and
// the box is not the animal: the top two cell rows of the 16x11 pose are the
// raised claw and nothing else. Whether that matters is a question about the
// picture, so this draws the picture instead of arguing about the number.
//
//	go run ./notes/s28-closer/b-push/scene
package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
	"github.com/donlucasx/xscapes/notes/s28-closer/b-push/pose"
)

var geoms = []struct{ w, h int }{{80, 24}, {111, 25}, {124, 22}, {143, 27}, {153, 51}}

func main() {
	fmt.Println("INK, in cells -- how much more of the screen the animal takes")
	for _, c := range []struct {
		name string
		rows []string
	}{
		{"today 12x7 ", join(pose.ShipAsk, pose.ShipLower)},
		{"mid   14x9 ", join(pose.MidAsk, pose.MidLower)},
		{"big   16x11", join(pose.BigAsk, pose.BigLower)},
	} {
		q := companion.ParseBitmap(c.rows).ToQuadrant()
		ink, box := 0, 0
		for _, r := range q {
			for _, ch := range r {
				box++
				if ch != ' ' {
					ink++
				}
			}
		}
		fmt.Printf("  %s  box %3d cells, ink %3d (%.0f%%)   %.1f%% of an 80x24 frame\n",
			c.name, box, ink, 100*float64(ink)/float64(box), 100*float64(ink)/(80*24))
	}
	fmt.Println()

	fmt.Println("PER-ROW INK of the 16x11 ask pose -- the box is not the animal:")
	q := companion.ParseBitmap(join(pose.BigAsk, pose.BigLower)).ToQuadrant()
	for i, r := range q {
		n := 0
		for _, ch := range r {
			if ch != ' ' {
				n++
			}
		}
		what := "shell / legs"
		switch {
		case i < 2:
			what = "RAISED CLAW ONLY"
		case i < 4:
			what = "eyes, stalks, low claw"
		case i == 4:
			what = "stalks and arms"
		}
		fmt.Printf("  row %2d: %2d of 16 cells inked   %s\n", i, n, what)
	}
	fmt.Println()

	fmt.Println("THE ANCHOR. The companion is drawn at top = H-2-chh, so its FEET are")
	fmt.Println("pinned and every extra row grows UPWARD, into the sea. Coming closer is")
	fmt.Println("the opposite motion, so both anchors are drawn: `under 2` is today's rule,")
	fmt.Println("`under 0` stands it on the last row and spends the two spare rows going DOWN.")
	fmt.Println()
	for _, g := range geoms {
		draw(g.w, g.h, 0.59, 2)
		draw(g.w, g.h, 0.59, 0)
	}
	fmt.Println("AND THE 14x9 RUNG, feet on the last row, at the two tightest geometries and")
	fmt.Println("at the level a modest tide withdrawal would reach:")
	for _, g := range []struct{ w, h int }{{124, 22}, {143, 27}} {
		for _, lv := range []float64{0.59, 0.35} {
			drawAs(g.w, g.h, lv, 0, 9, 14, pose.MidAsk, pose.MidLower, pose.MidEye, pose.MidEyeCells, pose.MidEyeRow, 7)
		}
	}

	fmt.Println("SAND WRITING COLUMNS -- compose() derives SandTo from the companion's width")
	fmt.Printf("%-9s %10s %10s %8s\n", "geom", "catW 12", "catW 16", "lost")
	for _, g := range geoms {
		a := sandCols(g.w, 12)
		b := sandCols(g.w, 16)
		fmt.Printf("%-9s %10d %10d %8d\n", fmt.Sprintf("%dx%d", g.w, g.h), a, b, a-b)
	}
}

// draw composes the big crab onto a settled shore and prints the frame with
// every row labelled SEA or SAND from the shore's own SandTop.
func draw(w, h int, level float64, under int) {
	drawAs(w, h, level, under, 11, 16, pose.BigAsk, pose.BigLower, pose.BigEyeSolid, [2]int{5, 9}, 2, 8)
}

func drawAs(w, h int, level float64, under, chh, ccw int, up, low, eyeb []string, eyeCells [2]int, eyeRow, headCol int) {
	sh := scape.NewShore(7, false)
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	tm := 0.0
	for k := 0; k < 400; k++ {
		tm += 0.08
		sh.Update(c, tm, scape.Activity{Level: level, Working: true, ContextUsed: 0.3, TimeOfDay: 11.0 / 24})
	}
	sandTop := sh.SandTop()

	top := h - under - chh
	right := 2 + w/32
	x := w - ccw - right

	rows := companion.ParseBitmap(join(up, low)).Mirrored().ToQuadrant()
	(&companion.Sprite{Rows: rows, Body: companion.CrabCoat}).Draw(c.Near(), x, top)
	eye := companion.ParseBitmap(eyeb).ToQuadrant()
	for _, e := range eyeCells { // symmetric, so the mirror maps them to themselves
		(&companion.Sprite{Rows: eye, Body: term.RGB{R: 232, G: 252, B: 226}}).Draw(c.Near(), x+e, top+eyeRow)
	}
	bub := companion.MirrorTail(companion.Bubble("allow Bash?"))
	bx := x + headCol - companion.TailCol(bub)
	if bx+len([]rune(bub[0])) > w {
		bx = w - len([]rune(bub[0]))
	}
	(&companion.Sprite{Rows: bub, Body: term.RGB{R: 250, G: 220, B: 170}, Opaque: true}).Draw(c.Near(), bx, top-len(bub))

	wet := 0
	for y := top; y < top+chh; y++ {
		if y < sandTop {
			wet++
		}
	}
	fmt.Printf("=== %dx%d level %.2f, under=%d, %dx%d crab -- waterline row %d, beach %d rows, companion rows %d-%d, %d of %d rows over the SEA\n",
		w, h, level, under, ccw, chh, sandTop, h-sandTop, top, top+chh-1, wet, chh)
	lines := strings.Split(c.RenderPlain(), "\n")
	lo := top - len(bub) - 1
	if lo < 0 {
		lo = 0
	}
	for y := lo; y < h; y++ {
		tag := "sea "
		if y >= sandTop {
			tag = "SAND"
		}
		mark := "  "
		if y == top {
			mark = "T>"
		}
		if y == top+chh-1 {
			mark = "B>"
		}
		// only the right-hand end of the frame, where the companion is
		r := []rune(lines[y])
		from := x - 26
		if from < 0 {
			from = 0
		}
		fmt.Printf("%s %s %2d |%s|\n", tag, mark, y, strings.TrimRight(string(r[from:]), " "))
	}
	fmt.Println()
}

func sandCols(w, catW int) int {
	right := 2 + w/32
	catX := w - catW - right
	span := 0
	// paceSpan is unexported; its live value is read off the shipped layout by
	// the a-step instrument. Zero here, so this reports the UPPER bound on the
	// sand and the difference between the two widths, which is what is asked.
	return (catX - 1 - span) - 2 + 1
}

func join(a, b []string) []string { return append(append([]string{}, a...), b...) }
