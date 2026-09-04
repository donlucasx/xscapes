package main

import "testing"

// The companion keeps a margin from the right edge that grows with the
// width, and never loses its face to the left edge when the pane is narrower
// than the sprite. His report at 124 columns: "too pushed to the side of the
// screen. a bit".
func TestTheCompanionKeepsItsDistanceFromTheEdge(t *testing.T) {
	const catW = 13
	for _, tc := range []struct{ w, wantRight int }{{124, 5}, {80, 4}, {40, 3}, {30, 2}} {
		lay := compose(tc.w, catW, true)
		right := tc.w - (lay.CatX + catW)
		if right != tc.wantRight {
			t.Errorf("w=%d: the companion sits %d columns from the right edge, want %d", tc.w, right, tc.wantRight)
		}
		if lay.SandTo != lay.CatX-1 || lay.BubbleX > lay.CatX {
			t.Errorf("w=%d: the sand (to %d) and the bubble (at %d) are not anchored to the companion at %d", tc.w, lay.SandTo, lay.BubbleX, lay.CatX)
		}
	}
	if lay := compose(10, catW, true); lay.CatX != 0 {
		t.Errorf("narrower than the sprite: the companion should pin to the left edge, got CatX %d", lay.CatX)
	}
}
