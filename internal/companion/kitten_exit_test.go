package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// A finished subagent's kitten swims OFF: from the water beside the litter
// toward the far edge as its exit progresses, receding as it goes, and never
// out of the water. Position carries the leaving, not rate.
func TestAnExitingKittenSwimsTowardTheFarEdge(t *testing.T) {
	const w, h = 100, 30
	seaTop, seaBot := 10, 20
	span := func(p float64, mirror bool) (minX, maxX, minY, maxY int, cells int) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat := NewCat()
		cat.FaceLeft(mirror)
		cw, ch := cat.Size()
		px := 5
		if mirror {
			px = w - cw - 5
		}
		py := h - 2 - ch
		if n := cat.DrawKittenExits(c.Near(), []float64{p}, px, py, w-1, seaTop, seaBot, 1.0, 7); p < 1 && n != 1 {
			t.Fatalf("p=%.2f mirror=%v: drew %d exits, want 1", p, mirror, n)
		}
		minX, minY, maxX, maxY = w, h, -1, -1
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				cell := c.Near().Cells[y*w+x]
				if !cell.Set || cell.R == ' ' || cell.R == '~' {
					continue
				}
				cells++
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
		return
	}
	for _, mirror := range []bool{true, false} {
		x0, _, y0, y1, n0 := span(0.05, mirror)
		x9, _, y9, y19, n9 := span(0.9, mirror)
		if n0 == 0 || n9 == 0 {
			t.Fatalf("mirror=%v: nothing drawn (%d, %d cells)", mirror, n0, n9)
		}
		// Companion on the right (mirror): open water is to the LEFT, so the
		// kitten moves left. Companion on the left: it moves right.
		if mirror && x9 >= x0 {
			t.Errorf("mirror: exiting kitten at x=%d then x=%d, want it further LEFT", x0, x9)
		}
		if !mirror && x9 <= x0 {
			t.Errorf("exiting kitten at x=%d then x=%d, want it further RIGHT", x0, x9)
		}
		for _, y := range []int{y0, y1, y9, y19} {
			if y <= seaTop || y >= seaBot {
				t.Errorf("mirror=%v: kitten row %d is outside the water (%d..%d)", mirror, y, seaTop+1, seaBot-1)
			}
		}
		if _, _, _, _, n := span(1.0, mirror); n != 0 {
			t.Errorf("mirror=%v: at progress 1 the kitten should be gone, %d cells drawn", mirror, n)
		}
	}
}
