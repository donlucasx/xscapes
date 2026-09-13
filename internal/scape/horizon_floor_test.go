package scape

import "testing"

// THE WATER MAY NOT WITHDRAW INTO THE SKY.
//
// hyFloor was a function of sy alone -- the row the beach starts at -- and
// TideRange is five rows, so at a short window the arithmetic permitted a
// waterline ABOVE THE HORIZON. Measured at 40x12 before the fix: four rows
// above it. Nothing in the layout could notice, because the only guard was
// against the writing band below.
func TestTheTideCannotWithdrawAboveTheHorizon(t *testing.T) {
	caught := 0
	for _, g := range [][2]int{{40, 12}, {30, 8}, {60, 14}, {80, 24}, {125, 28}, {200, 50}} {
		w, h := g[0], g[1]
		hy := int(float64(h) * 0.42)
		for beach := 2; beach < h-1; beach++ {
			sy := h - beach
			if sy <= hy+1 {
				sy = hy + 2
			}
			f := hyFloor(sy, hy)
			if f < float64(hy)+1 {
				t.Errorf("%dx%d beach=%d: the water may reach row %.1f, which is at or above "+
					"the horizon at %d", w, h, beach, f, hy)
			}
			// The negative control: the OLD rule really did allow it, or this
			// test is guarding something that was never broken.
			if old := float64(sy) - TideRange - 1; old < float64(hy)+1 {
				caught++
			}
		}
	}
	if caught == 0 {
		t.Fatal("the old rule never once put water above the horizon in this sweep, so this " +
			"test does not measure the defect it was written for")
	}
	t.Logf("the old rule put water at or above the horizon in %d of the swept geometries", caught)
}
