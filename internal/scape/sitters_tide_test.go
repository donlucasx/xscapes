package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
)

// WHERE THE SITTING CRABLETS ARE, AGAINST WHERE THE WATER GETS TO.
//
// His card, and his ruling on it was "measure it first" rather than a fix: the
// companion and the crablets beside it are anchored to the BOTTOM OF THE FRAME,
// while the swimmers are bounded by the waterline, which moves with the tide.
// Two different anchors is a fact about the code, verified by reading it. How
// often the water actually reaches the sitters is not, and that is what this
// measures.
//
// The two rows come from the call sites, quoted so a reader can check them:
//
//	frames.go:203   top := f.c.H - 2 - f.chh        // chh is 7, the box height
//	crab_kittens.go ky := py + 7 - ch               // py is that top
//
// so the litter's top row is H-2-ch and its bottom row is H-3, where ch is the
// crablet's own height at the tier in use. The bitmaps are read from the
// companion package rather than transcribed.
//
// A cell is WATER on exactly the renderer's own condition -- shore.go's
// `case fy < ex-0.5` -- against the shore's real per-column edge, so this is
// the painted waterline and not a model of it.
//
// ⇒ HIS RULING 2026-09-13, on the numbers below: LEAVE IT. Three of the four
// grounds are things this test measures and the fourth is arithmetic:
//
//   - The litter sits INSIDE the companion's own row span (22-25 against 19-25
//     at 125x28), so it can never be wetter than the animal beside it. The
//     companion's rows are wet from level 0.50 and 100% of frames from 0.90.
//     There is no fixing one without moving the other off the beach.
//   - There is nowhere to pin them to. They already hold the driest rows in the
//     frame, with two rows left below; pinning to the waterline means moving
//     them DOWN, and there is no down.
//   - The TIDE IMPROVED THIS. See the !Tide branch below for the numbers.
//   - An opaque sprite on the near layer over the shallows reads as an animal
//     at the water's edge, which is the picture the scape is for.
func TestSittersAndTheTide(t *testing.T) {
	tiers := map[string]int{
		"Crablet":      companion.ParseBitmap(companion.Crablet).H / 4,
		"CrabletSmall": companion.ParseBitmap(companion.CrabletSmall).H / 4,
		"CrabletTiny":  companion.ParseBitmap(companion.CrabletTiny).H / 4,
	}
	t.Logf("crablet heights in cells: Crablet=%d CrabletSmall=%d CrabletTiny=%d",
		tiers["Crablet"], tiers["CrabletSmall"], tiers["CrabletTiny"])
	ch := tiers["Crablet"] // the biggest rung: the worst case, and what a small litter gets

	geoms := []struct{ w, h int }{
		{125, 28}, // his usual
		{143, 27},
		{107, 51},
		{120, 26}, // what every other test in this package measures
		{80, 24},  // the design target
		{40, 12},  // the small end
	}
	levels := []float64{0.00, 0.25, 0.50, 0.75, 0.80, 0.85, 0.90, 0.95, 1.00}

	for _, g := range geoms {
		for _, lv := range levels {
			// ONE shore, warmed up: a fresh Shore per frame has no history, so
			// the tide's easing reads as if the session just started.
			sh := NewShore(7, false)
			c := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			tt := 0.0
			for i := 0; i < 200; i++ { // 10s at 20fps, against TideEase 3.0
				tt += 0.05
				sh.Update(c, tt, Activity{Level: lv, TimeOfDay: 0.55})
			}

			sitTop := g.h - 2 - ch
			sitBot := g.h - 3
			// The companion's own span, from the same call site: chh is the
			// box height, 7 for both animals. The litter sits INSIDE this, so
			// whenever a sitter row is wet a companion row is wet too.
			const chh = 7
			catTop := g.h - 2 - chh
			frames, wetFrames, wetRightFrames, wetCat := 0, 0, 0, 0
			worstReach, worstCols := 0.0, 0
			for i := 0; i < 120; i++ { // 6s of samples
				tt += 0.05
				sh.Update(c, tt, Activity{Level: lv, TimeOfDay: 0.55})
				edge := sh.LastEdgeExport()
				if len(edge) == 0 {
					t.Fatalf("%dx%d: no waterline, so this measures nothing", g.w, g.h)
				}
				frames++
				cols, colsRight, reach, catCols := 0, 0, 0.0, 0
				for x, e := range edge {
					if e > reach {
						reach = e
					}
					for y := catTop; y <= sitBot; y++ {
						if float64(y) < e-0.5 {
							catCols++
							break
						}
					}
					// Does the water cover any row the litter stands on?
					for y := sitTop; y <= sitBot; y++ {
						if float64(y) < e-0.5 {
							cols++
							// The composition is mirrored by default: the
							// companion sits on the RIGHT and the litter grows
							// leftward from it, so the right-hand third is
							// where a sitter actually is.
							if x >= len(edge)*2/3 {
								colsRight++
							}
							break
						}
					}
				}
				if cols > 0 {
					wetFrames++
				}
				if catCols > 0 {
					wetCat++
				}
				if colsRight > 0 {
					wetRightFrames++
				}
				if cols > worstCols {
					worstCols = cols
				}
				if reach > worstReach {
					worstReach = reach
				}
			}
			pct := 100 * wetFrames / frames
			t.Logf("%3dx%-3d level %.2f | litter rows %2d-%2d, companion rows %2d-%2d | water to row %5.2f | "+
				"wet frames: litter %3d%% (its own columns %3d%%), COMPANION %3d%% | worst %d of %d columns",
				g.w, g.h, lv, sitTop, sitBot, catTop, sitBot, worstReach,
				pct, 100*wetRightFrames/frames, 100*wetCat/frames, worstCols, g.w)

			// THE GUARANTEE THIS MEASUREMENT EARNS, and it is the useful half:
			// at rest and through ordinary work the litter is on DRY SAND at
			// every size the product designs for. The water only arrives at the
			// top of the activity range.
			//
			// 40x12 is excluded by name and with its number rather than by
			// silence: it is wet in 8% of frames at level 0.00 and 100% at
			// 1.00, which is the same small end he ruled "leave it as is" for
			// the shooting star. Excluding it quietly would have hidden a
			// measurement; this way a reader sees it.
			if g.w == 40 && g.h == 12 {
				continue
			}
			// AND THE GUARANTEE IS THE TIDE'S, which is the finding that
			// settled the card. With XSCAPES_TIDE=0 the sea sits at a fixed
			// high waterline and the litter has ALWAYS stood in it: at 125x28
			// its own columns are wet 36% of frames at level 0.00 and 65% at
			// 1.00, against 0% and 30% with the tide on. So the feature the
			// card blamed is the one that halved it, and asserting this without
			// the tide would be asserting something that was never true.
			if !Tide {
				continue
			}
			if lv <= 0.75 && pct != 0 {
				t.Errorf("%dx%d at level %.2f: the water is on a sitter row in %d%% of frames. "+
					"The litter is meant to be on dry sand below full activity", g.w, g.h, lv, pct)
			}
		}
	}
}
