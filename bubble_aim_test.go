package main

import (
	"sort"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// alertEyeCols is the columns of the companion's open-eye ink on the rendered
// frame, found by COLOUR and confirmed against the resolver.
//
// ⚠ IT USED TO READ BACK THE LITERAL RUNE 'O' ON ONE ROW, top+2, and that is
// exactly as much of the eye as the shipped 12x7 crab has: one cell each side.
// The come-closer pose (internal/companion/crab_near.go) draws the eye as a
// BITMAP two cells by two, so there is no 'O' anywhere on the frame and the row
// it sits on moves down as the animal grows. Keyed on the glyph, this whole
// guarantee would have gone quietly vacuous at the rung the feature exists for.
//
// companion.EyeAlert is exported for this. The glyph is the one property of the
// eye an approach is allowed to change; the colour is not, and nothing else in
// the scene is that near-white green.
//
// TWO STAGES, and the second is the point of the original test: the candidates
// come off the RAW near layer, because that is the authored colour rather than
// a quantised one, and each is then confirmed to still carry its rune on the
// RESOLVED frame. A glyph the layer holds can still lose to the resolver, and
// an eye that lost would leave the balloon pointing at a cell of sea.
// It returns the distinct columns, and the MEAN column of every eye cell --
// which is the face's midline, and the number the near pose is built to hold
// still. Measured on the frame, all five widths, both facings: the crab's mean
// is 112.50 / 95.50 / 69.50 / 50.50 / 30.50 and the cat's 114 / 97 / 71 / 52 /
// 32, IDENTICAL at all three rungs. The pointer is +0.50 from it on the crab
// (its face is an even number of cells) and +0.00 on the cat, at every one.
func alertEyeCols(c *canvas.Canvas, top, h int) (cols []int, mean float64) {
	l := c.Near()
	seen := map[int]bool{}
	sum, n := 0, 0
	for y := top; y < top+h && y < l.H; y++ {
		if y < 0 {
			continue
		}
		for x := 0; x < l.W; x++ {
			cell := l.Cells[y*l.W+x]
			// A space is the hole in the middle of the ring, not ink: the near
			// eye plots every cell of its tile, spaces included, so that the
			// hole comes out coat-coloured instead of sea.
			if !cell.Set || cell.R == ' ' || cell.FG != companion.EyeAlert {
				continue
			}
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r != cell.R {
				continue // lost to the resolver; it is not on his screen
			}
			seen[x] = true
			sum += x
			n++
		}
	}
	cols = make([]int, 0, len(seen))
	for x := range seen {
		cols = append(cols, x)
	}
	sort.Ints(cols)
	if n > 0 {
		mean = float64(sum) / float64(n)
	}
	return cols, mean
}

// The balloon's pointer comes out of the companion's face, at every size the
// companion can be.
//
// It did not. Measured on the rendered frame at 124 columns before this
// shipped: the `v` sat on column 101 while the companion's box began at 107 and
// its nearest ink at 110, so the balloon spoke from nine cells of bare sand.
// The balloon was anchored to its own corner and cleared the companion by two
// columns; the unmirrored layout it was mirrored from overlaps on purpose.
//
// Read off the RENDERED cells rather than the layout arithmetic: the pointer
// and the eyes are both glyphs plotted over a background of half cells, and a
// glyph the layer holds can still lose to the resolver.
//
// THE RUNG LOOP IS WHY THIS TEST EARNS ITS KEEP AGAIN. His idea of 2026-09-10
// is that the companion walks up closer before it prompts, and the one moment
// it is bigger is the one moment a balloon is up. The near pose grows DOWNWARD
// and outward from the same top row and is centred on the same head column, so
// the claim is that the balloon needs no change at all -- and a claim like that
// is worth exactly what it is tested at. Rung 0 is the shipped 12x7 and the
// assertions there are the ones this test always made, unweakened.
func TestTheBalloonPointsAtTheCompanion(t *testing.T) {
	for _, rung := range []int{0, 1, 2} {
		for _, name := range []string{"cat", "crab"} {
			for _, mirror := range []bool{true, false} {
				for _, w := range []int{124, 107, 80, 60, 40} {
					const h = 27
					old := companion.Near
					companion.Near = rung
					c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					sh := scape.NewShore(7, false)
					cat := companion.New(name)
					cat.FaceLeft(mirror)
					ccw, chh := cat.Size()
					lay := compose(w, ccw, mirror)
					sh.MoonX = lay.MoonX
					st := demoState(2, 0.3, 0.93)
					st.Bubble = "allow Bash?"
					st.BubbleAsk = true
					st.Pose = companion.NeedsYou
					st.Tail = st.FitTail(time.Now(), lay.SandTo-lay.SandFrom)
					// Walk it all the way in. One rung is 0.28 s and the ladder
					// is two rungs deep, so three seconds of frames is more than
					// enough; Approach clamps at the top and stays there.
					for i := 0; rung > 0 && i < 60; i++ {
						cat.Approach(0.05, true)
					}
					_, _, boxH := cat.DrawnBox(w)
					sh.Update(c, 2, st.Act)
					top := c.H - 2 - chh
					drawScene(c, sh, cat, lay, st, 2, 7, top)
					companion.Near = old

					// The pointer is on the balloon's last row, which sits
					// directly above the companion. The near pose does not move
					// it: the sprite is anchored by its TOP row at the same y,
					// so the balloon's own row is untouched at every rung.
					var tail []int
					for x := 0; x < w; x++ {
						if r, _, _ := c.ResolveAt(x, top-1, term.Profile256); r == 'v' {
							tail = append(tail, x)
						}
					}
					if len(tail) != 1 {
						t.Fatalf("rung %d %s mirror=%v w=%d: %d pointers on the balloon's bottom row, want 1",
							rung, name, mirror, w, len(tail))
					}

					eyes, mid := alertEyeCols(c, top, boxH)
					if len(eyes) < 2 {
						t.Fatalf("rung %d %s mirror=%v w=%d: found %d columns of open-eye ink, want at least 2 "+
							"(one column each side); if this is zero the eye stopped being EyeAlert and the "+
							"test has gone blind rather than the frame having changed",
							rung, name, mirror, w, len(eyes))
					}
					// At the shipped size the eyes are exactly two cells, one
					// each side, and nothing else in the scene draws one. That
					// is the assertion this test has always made; keep it exact
					// where it has always been exact.
					if rung == 0 && len(eyes) != 2 {
						t.Fatalf("%s mirror=%v w=%d: found %d open eyes at the shipped size, want 2",
							name, mirror, w, len(eyes))
					}
					lo, hi := eyes[0], eyes[len(eyes)-1]
					if tail[0] < lo || tail[0] > hi {
						t.Errorf("rung %d %s mirror=%v w=%d: the pointer is on column %d, outside the eyes at %d and %d "+
							"-- the balloon is speaking from bare sand",
							rung, name, mirror, w, tail[0], lo, hi)
					}
					// AND IT IS ON THE MIDLINE, not merely somewhere between the
					// outermost eye cells. "Between the eyes" was a tight test
					// while each eye was one cell; against a bitmap eye four
					// cells across it is a wide target, and a two-cell error --
					// exactly what losing the near pose's centring costs -- sails
					// through it. Half a cell is not a tolerance chosen to pass:
					// it is the whole spread the measurement shows, +0.50 on the
					// crab and +0.00 on the cat, at every rung and every width.
					if d := mid - float64(tail[0]); d > 0.5 || d < -0.5 {
						t.Errorf("rung %d %s mirror=%v w=%d: the pointer is on column %d but the face's midline "+
							"is %.2f, %.2f cells away -- the balloon is aimed off the head",
							rung, name, mirror, w, tail[0], mid, -d)
					}
				}
			}
		}
	}
}
