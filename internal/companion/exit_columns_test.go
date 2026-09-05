package companion

import "testing"

// Kittens that leave together must not be drawn on top of each other. The
// dwell (reduce.KittenDwell) makes a fan-out leave at one instant, so every
// exit had the same progress and the same column: five kittens swimming off
// read as one. Each exit now trails the one before it by a kitten's width.
func TestExitsNeverShareColumns(t *testing.T) {
	for _, mirror := range []bool{true, false} {
		for _, n := range []int{2, 3, 5, 8} {
			for _, w := range []int{80, 100, 120, 152} {
				for _, p := range []float64{0.05, 0.5, 0.9} {
					cat := NewCat()
					cat.FaceLeft(mirror)
					cw, ch := cat.Size()
					px := w - cw - 2
					if !mirror {
						px = 2
					}
					py := 30 - 2 - ch
					exits := make([]float64, n)
					for i := range exits {
						exits[i] = p
					}
					spans := cat.exitSpans(exits, px, py, w, 10, 0)
					if p == 0.5 && len(spans) < 2 {
						t.Errorf("mirror=%v n=%d w=%d p=%.2f: only %d exits placed, want at least 2 side by side", mirror, n, w, p, len(spans))
					}
					for a := 0; a < len(spans); a++ {
						for b := a + 1; b < len(spans); b++ {
							sa, sb := spans[a], spans[b]
							if sa.x0 <= sb.x1 && sb.x0 <= sa.x1 {
								t.Errorf("mirror=%v n=%d w=%d p=%.2f: exits %d and %d share columns %d-%d and %d-%d",
									mirror, n, w, p, sa.i, sb.i, sa.x0, sa.x1, sb.x0, sb.x1)
							}
						}
					}
				}
			}
		}
	}
}

// Exits that are already spread out keep their own places: the queue rule
// only holds a kitten back from the one ahead of it, it never bunches them.
func TestSpreadExitsKeepTheirPlaces(t *testing.T) {
	cat := NewCat()
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	w := 120
	px, py := w-cw-2, 30-2-ch
	exits := []float64{0.9, 0.5, 0.1} // oldest first, the oldest furthest along
	spans := cat.exitSpans(exits, px, py, w, 10, 0)
	if len(spans) != 3 {
		t.Fatalf("placed %d of 3 spread exits", len(spans))
	}
	// Mirrored: leaving means travelling LEFT, so progress 0.9 is the smallest x.
	if !(spans[0].x0 < spans[1].x0 && spans[1].x0 < spans[2].x0) {
		t.Errorf("order broken: x0 = %d, %d, %d for progress 0.9, 0.5, 0.1", spans[0].x0, spans[1].x0, spans[2].x0)
	}
}
