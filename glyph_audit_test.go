package main

import (
	"fmt"
	"sort"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The 256 audit has only ever been run on backgrounds. This runs it on every
// GLYPH the scene draws, at every hour, by rendering the frame twice -- once as
// truecolor, which gives the colour the scene asked for, and once as 256, which
// gives the colour Terminal.app paints -- and comparing them cell by cell.
//
// Two failures are looked for, and they are different things:
//
//   - COLOUR LOST: the source had real chroma and the terminal painted a grey.
//   - HUE FLIPPED: the source stated a channel ordering clearly and the painted
//     colour reverses it, which is how a blue becomes a violet.
func TestProbeGlyphColoursOn256(t *testing.T) {
	const w, h = 90, 30
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
	chromaOf := func(c term.RGB) int {
		mx, mn := int(c.R), int(c.R)
		for _, v := range []int{int(c.G), int(c.B)} {
			if v > mx {
				mx = v
			}
			if v < mn {
				mn = v
			}
		}
		return mx - mn
	}
	flipped := func(src, q term.RGB) bool {
		pair := func(a, b, qa, qb uint8) bool {
			switch d := int(a) - int(b); {
			case d >= 25:
				return qa > qb
			case d <= -25:
				return qa < qb
			}
			return true
		}
		return !(pair(src.R, src.G, q.R, q.G) && pair(src.G, src.B, q.G, q.B) && pair(src.R, src.B, q.R, q.B))
	}

	type stat struct{ lost, flip, seen int }
	byHour := map[float64]*stat{}
	var hours []float64
	totalLost, totalFlip, totalGlyphs := 0, 0, 0

	for i := 0; i < 24; i++ {
		tod := float64(i) / 24
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		sh.MoonX = lay.MoonX
		for k := 0; k < 12; k++ {
			sh.Update(c, 3.0+float64(k)/20, scape.Activity{
				Working: true, Level: 0.6, TimeOfDay: tod, ContextUsed: 0.3})
		}
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: tail}, 3.7, 7, c.H-2-chh)

		st := &stat{}
		byHour[tod] = st
		hours = append(hours, tod)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				ch, src, _ := c.ResolveAt(x, y, term.ProfileTrueColor)
				if ch == ' ' {
					continue // no glyph: that is the background audit's business
				}
				_, painted, _ := c.ResolveAt(x, y, term.Profile256)
				st.seen++
				totalGlyphs++
				// Three outcomes, and the first version of this probe ran the
				// first two together: a glyph that goes grey has no channel
				// ordering left to reverse, so it was being counted as a hue
				// flip as well and the flip column read four times too high.
				// Washed out and wrong-coloured are different defects with
				// different fixes.
				switch {
				case chromaOf(src) >= 40 && chromaOf(painted) < 20:
					st.lost++
					totalLost++
				case chromaOf(painted) >= 20 && flipped(src, painted):
					st.flip++
					totalFlip++
				}
			}
		}
	}
	sort.Float64s(hours)
	fmt.Printf("%-7s %8s %16s %18s\n", "hour", "glyphs", "washed to grey", "wrong hue, coloured")
	for _, tod := range hours {
		s := byHour[tod]
		fmt.Printf("%02d:00   %8d %8d %4.1f%% %8d %4.1f%%\n", int(tod*24), s.seen,
			s.lost, 100*float64(s.lost)/float64(s.seen),
			s.flip, 100*float64(s.flip)/float64(s.seen))
	}
	fmt.Printf("\nTOTAL   %8d %8d %4.1f%% %8d %4.1f%%\n", totalGlyphs,
		totalLost, 100*float64(totalLost)/float64(totalGlyphs),
		totalFlip, 100*float64(totalFlip)/float64(totalGlyphs))
}
