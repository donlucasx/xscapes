package companion

import "testing"

// Two swimmers never share columns. Lanes are a row apart, and a swimmer in
// the lane above a neighbour that shares its columns paints over the
// neighbour's face: his 2026-09-05 note on the page, "two sub agents
// overlapping (the one behind's face breaks)".
func TestSwimmersNeverShareColumns(t *testing.T) {
	for _, n := range []int{2, 3, 5, 8, 12} {
		for _, w := range []int{80, 100, 120, 152} {
			for _, seed := range []int64{1, 7, 42, 1234} {
				cat := NewCat()
				cat.FaceLeft(true)
				cw, ch := cat.Size()
				px, py := w-cw-2, 30-2-ch
				var idx []int
				for i := 0; i < n; i++ {
					idx = append(idx, i)
				}
				spans := cat.swimmerSpans(idx, px, py, w, 10, 0, 1.3, seed)
				for a := 0; a < len(spans); a++ {
					for b := a + 1; b < len(spans); b++ {
						sa, sb := spans[a], spans[b]
						if sa.y == sb.y {
							continue // one lane: the slot logic keeps them apart already
						}
						if sa.x0 <= sb.x1 && sb.x0 <= sa.x1 {
							t.Errorf("n=%d w=%d seed=%d: swimmers at rows %d and %d share columns %d-%d and %d-%d",
								n, w, seed, sa.y, sb.y, sa.x0, sa.x1, sb.x0, sb.x1)
						}
					}
				}
			}
		}
	}
}
