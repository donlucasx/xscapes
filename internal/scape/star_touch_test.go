package scape

import (
	"math"
	"testing"
)

// HIS RULING, 2026-09-12: "dont let em touch." ⚠ NOT YET HONOURED, and this
// test is the characterisation that says so rather than a fix pretending to be
// one. It FAILS THE BUILD only if the numbers get worse.
//
// What was tried and rejected, measured both ways:
//
//   - The obvious read was that the bar has no headroom: two stars one row
//     apart in the same column are EXACTLY starHardSep apart and the
//     acceptance test is `<`, so the boundary case passes. Making it strict
//     made it WORSE -- 10 touching pairs became 12 at 40x12 seed 7 -- because
//     rejecting a dart pushes the placement into roomiestCell, which takes the
//     best cell there IS and can be worse than the dart it replaced.
//   - The next read was that the sky is too small. It is not: at 40x12 the
//     band is 3 rows of 40 columns, room for dozens at the bar, and EIGHT
//     stars still produce three touching pairs.
//
// So the cause is in the placement itself and is not yet found. The likely
// direction is the candidate cell set -- constellationCells keeps stars off
// the disc's columns, and at 40x12 the disc may be eating most of a three-row
// band -- but that is a hypothesis and it has not been measured.
//
// Two stars that render in adjacent cells read as ONE mark, and this channel's
// whole job is to be counted. The separation floor is 2.0 SCREEN units -- a
// terminal cell is about twice as tall as it is wide, so the distance between
// (dx,dy) cells is hypot(dx, 2*dy) -- and two stars one row apart in the same
// column measure EXACTLY 2.00. The acceptance test was `< floor`, so the
// boundary case passed: the bar had no headroom at all.
func TestNoTwoStarsTouch(t *testing.T) {
	worst := 0
	for _, g := range [][2]int{{40, 12}, {30, 8}, {60, 14}, {80, 24}, {125, 28}, {143, 27}, {200, 50}} {
		w, h := g[0], g[1]
		hy := int(float64(h) * 0.42)
		for _, seed := range []int64{1, 7, 42} {
			for _, total := range []int{8, 19, 26, 32} {
				sh := NewShore(seed, false)
				scale := math.Min(float64(w)/80.0, float64(h)/24.0)
				scale = math.Max(0.45, math.Min(1.35, scale))
				sy := h - 6
				if sy <= hy+1 {
					sy = hy + 2
				}
				top, bot := starBand(hy, foamCeiling(sy, hy, scale))
				if bot < top {
					continue
				}
				pts := sh.starPlaces(w, hy, top, bot, total)
				touching := 0
				for i := range pts {
					for j := i + 1; j < len(pts); j++ {
						dx := float64(pts[i][0] - pts[j][0])
						dy := float64(pts[i][1]-pts[j][1]) * starCellAspect
						if math.Hypot(dx, dy) < starHardSep+0.001 {
							touching++
						}
					}
				}
				if touching > worst {
					worst = touching
				}
				if touching > 0 {
					t.Logf("%dx%d seed %d, %d stars: %d pairs touching", w, h, seed, total, touching)
				}
			}
		}
	}
	// The characterisation. 12 is what the shipped placement produces today,
	// at 40x12 seed 7. If a change makes it worse, that is a regression and
	// this fails; if a change fixes it, this fails too and the number comes
	// down with a note saying what did it.
	// 54 is the sweep's true worst, and it is not the 12 an earlier run of this
	// file reported: that number came from reading the first few lines of the
	// output, which happened to be 40x12. Measure the whole population before
	// quoting one of it.
	const known = 54
	if worst > known {
		t.Errorf("touching pairs got WORSE: %d against the known %d", worst, known)
	}
	if worst < known {
		t.Errorf("touching pairs improved to %d from %d -- good, now lower `known` and say "+
			"in the comment what fixed it", worst, known)
	}
	t.Logf("worst case across the sweep: %d touching pairs (unchanged, still open)", worst)
}
