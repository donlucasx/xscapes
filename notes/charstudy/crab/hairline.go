package main

import (
	"fmt"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// hairlineProbe counts the cells in a real frame that are still painted with a
// half block, at HIS geometry and with production's own switches. Every one of
// those is a cell with two colours in it, and on Terminal.app a two-colour cell
// leaves a one-pixel rule that cannot be removed inside half blocks.
func hairlineProbe(seed int64) {
	term.LowerHalf = term.DetectSplit("Apple_Terminal")
	term.NoSplitCells = term.DetectNoSplit("Apple_Terminal")
	fmt.Printf("production switches: LowerHalf=%v NoSplitCells=%v\n\n", term.LowerHalf, term.NoSplitCells)

	w, h := 126, 47 // his window, the scape band under a 27-row agent
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: 0.778, ContextUsed: 0.3, TodoDone: 2, TodoTotal: 5}
	for k := 0; k < 12; k++ {
		sh.Update(c, 3+float64(k)/40, act)
	}
	comp := companion.New("crab")
	comp.FaceLeft(true)
	cw, ch := comp.Size()
	px, py := w-cw-(2+w/32), h-2-ch
	comp.Draw(c.Near(), px, py, 3, companion.Working)
	comp.DrawKittens(c.Near(), c.Mid(), px, py, 4, w-1, int(float64(h)*0.42)+1, sh.SandTop()-2, 3, seed)

	runes := map[rune]int{}
	perRow := map[int]int{}
	total := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			runes[r]++
			// A cell is two-coloured if it carries ink AND its fg differs from
			// its bg -- that is what leaves a rule, whatever the glyph is.
			if r != ' ' && r != 0 && fg != bg {
				perRow[y]++
				total++
			}
		}
	}
	type rc struct {
		r rune
		n int
	}
	var rs []rc
	for r, n := range runes {
		rs = append(rs, rc{r, n})
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].n > rs[j].n })
	fmt.Println("most common runes in the frame:")
	for i, e := range rs {
		if i >= 8 {
			break
		}
		name := string(e.r)
		if e.r == ' ' {
			name = "<space>"
		} else if e.r == 0 {
			name = "<none>"
		}
		fmt.Printf("  %-8s %5d\n", name, e.n)
	}
	fmt.Println()
	fmt.Printf("split cells in the frame: %d of %d (%.1f%%)\n", total, w*h, float64(total)/float64(w*h)*100)
	fmt.Printf("sand line at row %d, waterline region below row %d\n\n", sh.SandTop(), int(float64(h)*0.42)+1)
	rows := []int{}
	for y := range perRow {
		rows = append(rows, y)
	}
	sort.Ints(rows)
	for _, y := range rows {
		where := "sky"
		switch {
		case y >= sh.SandTop():
			where = "BEACH"
		case y >= int(float64(h)*0.42)+1:
			where = "sea"
		}
		bar := ""
		for i := 0; i < perRow[y] && i < 60; i++ {
			bar += "#"
		}
		fmt.Printf("row %2d  %-5s %3d cells %s\n", y, where, perRow[y], bar)
	}
}
