package main

import (
	"math"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// The writing in the sand has to stay readable at every hour, and an older
// line must never out-shout a newer one.
//
// This replaces a test in internal/scape that MIRRORED drawSand's ink rule
// instead of running it: it recomputed the ink from the palette's nominal sand
// and asserted against that. The moment the beach could fade toward black,
// nominal and painted stopped being the same colour, and the mirror kept
// passing while the real thing lost contrast. A test that reimplements the
// code it is checking can only prove the copy agrees with itself.
//
// So this renders a real frame and reads the pixels back: the glyph colour the
// renderer chose, against the background it actually painted underneath.
func roundAll(v []float64) []int {
	out := make([]int, len(v))
	for i, x := range v {
		out[i] = int(x + 0.5)
	}
	return out
}

func TestSandWritingStaysReadable(t *testing.T) {
	const w, h = 92, 40
	work := []string{
		"read   internal/auth/handler.go  142 lines",
		"edit   internal/auth/handler.go  +18 -2",
		"shell  go test ./...  4.1s",
		"read   README.md  61 lines",
	}

	for _, tod := range []float64{0.0, 0.25, 0.5, 0.75, 0.888} {
		lines := make([]reduce.Line, 0, len(work))
		for i, s := range work {
			lines = append(lines, reduce.Line{Text: s, Age: 1 - float64(i+1)/float64(len(work))})
		}

		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		sh.MoonX = 0.28
		for i := 0; i < 12; i++ {
			sh.Update(c, 2+float64(i)/20, scape.Activity{
				Working: true, Level: 0.6, TimeOfDay: tod, ContextUsed: 0.3})
		}
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		drawScene(c, sh, cat, lay,
			reduce.State{Pose: companion.Working, Tail: lines}, 3.1, 7, c.H-2-chh)

		// Contrast per row, oldest first: the tail is painted newest-lowest.
		near := c.Near()
		var got []float64
		for y := 0; y < c.H; y++ {
			var sum float64
			var n int
			for x := lay.SandFrom; x < lay.SandTo && x < c.W; x++ {
				cell := near.Cells[y*c.W+x]
				if !cell.Set || cell.R == ' ' {
					continue
				}
				sum += math.Abs(luma(cell.FG) - luma(c.BG[y*c.W+x]))
				n++
			}
			if n >= 10 {
				got = append(got, sum/float64(n))
			}
		}
		if len(got) < len(work) {
			t.Fatalf("tod %.3f: found %d written rows, want %d", tod, len(got), len(work))
		}
		got = got[len(got)-len(work):]

		t.Logf("tod %.3f: contrast oldest->newest = %v", tod, roundAll(got))

		// got runs top to bottom, which is oldest to NEWEST: the tail is
		// painted newest-lowest, nearest the companion and last in the
		// reading order. So contrast must climb as it goes down.
		prev := 0.0
		for i, contrast := range got {
			if contrast < 18 {
				t.Errorf("tod %.3f line %d: only %.1f luma from the beach it is written on -- unreadable",
					tod, i, contrast)
			}
			if contrast < prev-1e-6 {
				t.Errorf("tod %.3f line %d: contrast %.1f is below the older line's %.1f -- age must dim, not brighten",
					tod, i, contrast, prev)
			}
			prev = contrast
		}
	}
}
