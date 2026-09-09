package main

import (
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The balloon's pointer comes out of the companion's face.
//
// It did not. Measured on the rendered frame at 124 columns before this
// shipped: the `v` sat on column 101 while the companion's box began at 107 and
// its nearest ink at 110, so the balloon spoke from nine cells of bare sand.
// The balloon was anchored to its own corner and cleared the companion by two
// columns; the unmirrored layout it was mirrored from overlaps on purpose.
//
// Read off the RENDERED cells rather than the layout arithmetic: the pointer
// and the eyes are both glyphs plotted over a background of half cells, and a
// glyph the layer holds can still lose to the resolver.
func TestTheBalloonPointsAtTheCompanion(t *testing.T) {
	for _, name := range []string{"cat", "crab"} {
		for _, mirror := range []bool{true, false} {
			for _, w := range []int{124, 107, 80, 60, 40} {
				const h = 27
				c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				sh := scape.NewShore(7, false)
				cat := companion.New(name)
				cat.FaceLeft(mirror)
				ccw, chh := cat.Size()
				lay := compose(w, ccw, mirror)
				sh.MoonX = lay.MoonX
				st := demoState(2, 0.3, 0.93)
				st.Bubble = "allow Bash?"
				st.BubbleAsk = true
				st.Pose = companion.NeedsYou
				st.Tail = st.FitTail(time.Now(), lay.SandTo-lay.SandFrom)
				sh.Update(c, 2, st.Act)
				top := c.H - 2 - chh
				drawScene(c, sh, cat, lay, st, 2, 7, top)

				read := func(y int, want rune) []int {
					var cols []int
					for x := 0; x < w; x++ {
						if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == want {
							cols = append(cols, x)
						}
					}
					return cols
				}
				// The pointer is on the balloon's last row, which sits directly
				// above the companion.
				tail := read(top-1, 'v')
				if len(tail) != 1 {
					t.Fatalf("%s mirror=%v w=%d: %d pointers on the balloon's bottom row, want 1", name, mirror, w, len(tail))
				}
				// NeedsYou opens both eyes as 'O', and nothing else in the
				// scene draws one.
				eyes := read(top+2, 'O')
				if len(eyes) != 2 {
					t.Fatalf("%s mirror=%v w=%d: found %d open eyes, want 2", name, mirror, w, len(eyes))
				}
				if tail[0] < eyes[0] || tail[0] > eyes[1] {
					t.Errorf("%s mirror=%v w=%d: the pointer is on column %d, outside the eyes at %d and %d",
						name, mirror, w, tail[0], eyes[0], eyes[1])
				}
			}
		}
	}
}
