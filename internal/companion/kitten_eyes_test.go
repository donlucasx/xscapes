package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// Every kitten that gets drawn has to keep both eyes.
//
// He photographed swimmers in the water with one eye. A face with one eye does
// not read as a face at all, and the litter is the subagent count -- a channel
// nobody can read is a channel that is not there.
func TestEveryKittenKeepsBothEyes(t *testing.T) {
	// Sweep seeds and times as well as counts: the eye-eating needs two
	// kittens to land next to each other, and one seed can easily miss it.
	for _, n := range []int{1, 2, 3, 5, 8, 12, 16, 20, 24} {
		for _, w := range []int{60, 80, 100, 120, 152} {
			for _, seed := range []int64{1, 7, 42, 1234} {
				for _, tt := range []float64{0.0, 1.3, 3.0, 7.7} {
					c := canvas.New(w, 30, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					cat := NewCat()
					cat.FaceLeft(true)
					cw, ch := cat.Size()
					px, py := w-cw-2, 30-2-ch
					drawn := cat.DrawKittens(c.Near(), c.Mid(), px, py, n, w, 10, tt, seed)
					if drawn == 0 {
						continue
					}
					// Count eye glyphs per row-run. Eyes are 'o' or '-' in the eye
					// colour; counting them by colour avoids catching the sea.
					eyes := 0
					for y := 0; y < c.H; y++ {
						for x := 0; x < c.W; x++ {
							cell := c.Near().Cells[y*c.W+x]
							if cell.Set && cell.FG == kittenEye && (cell.R == 'o' || cell.R == '-') {
								eyes++
							}
						}
					}
					if eyes != drawn*2 {
						t.Fatalf("n=%d w=%d seed=%d t=%.1f: %d kittens drawn but %d eyes on screen (want %d)",
							n, w, seed, tt, drawn, eyes, drawn*2)
					}
				}
			}
		}
	}
}
