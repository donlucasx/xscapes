package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// The crab has to survive every shape a real session can put it in without
// panicking or drawing outside the canvas. It uses all twelve columns of its
// box where the cat's ink uses nine, so the narrow and short cases are new
// ground even though the box is not.
func TestTheCrabSurvivesEveryShape(t *testing.T) {
	for _, w := range []int{20, 40, 60, 80, 100, 126, 200} {
		for _, h := range []int{8, 12, 24, 30, 74} {
			for _, n := range []int{0, 1, 6, 24, 60} {
				for _, st := range []State{Resting, Working, NeedsYou, Done, Worried} {
					c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					crab := NewCrab()
					crab.FaceLeft(true)
					cw, ch := crab.Size()
					px, py := w-cw-2, h-2-ch
					crab.Draw(c.Near(), px, py, 2.5, st)
					crab.DrawKittens(c.Near(), c.Mid(), px, py, n, w-1, h/3, h*2/3, 2.5, 7)
					crab.DrawKittenExits(c.Near(), []float64{0, 0.3, 0.7, 0.99}, px, py, w-1, h/3, h*2/3, 2.5, 7)
				}
			}
		}
	}
}

// Degenerate geometry must not panic either: a pane can be one column wide
// mid-resize, and the water can be inverted before the shore has settled.
func TestTheCrabSurvivesDegenerateGeometry(t *testing.T) {
	for _, g := range [][4]int{{1, 1, 0, 0}, {4, 4, 5, 2}, {12, 7, 0, 0}, {30, 10, 9, 3}, {80, 24, 20, 20}} {
		c := canvas.New(max(g[0], 1), max(g[1], 1), canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		crab := NewCrab()
		crab.FaceLeft(true)
		crab.Draw(c.Near(), 0, 0, 1, Working)
		crab.DrawKittens(c.Near(), c.Mid(), 0, 0, 8, g[0], g[2], g[3], 1, 3)
		crab.DrawKittenExits(c.Near(), []float64{0.5}, 0, 0, g[0], g[2], g[3], 1, 3)
	}
}
