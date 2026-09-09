package companion

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// The crablets: one per subagent, at the kittens' size and in the kittens'
// place. Everything here mirrors the cat's litter deliberately -- the same
// split between sand and water, the same lanes, the same never-share-columns
// guarantee, the same bodies-before-faces ordering -- because those rules were
// each written down after a defect he reported, and a new animal does not get
// to rediscover them.

// laneAlloc places sprites into lanes so that no two ever overlap.
//
// The first version bucketed columns as x/kw and called that "never share a
// column". It does not: two sprites one cell apart can land in different
// buckets and still sit on top of each other, and the crab's own tests caught
// exactly that -- a body drawn over a neighbour's eye, which is the defect the
// cat's swimmerSpans was written to stop. Overlap is an interval question, so
// this keeps the intervals.
type laneAlloc struct {
	kw, lanes, w int
	taken        map[int][][2]int
}

func newLaneAlloc(kw, lanes, w int) *laneAlloc {
	return &laneAlloc{kw: kw, lanes: lanes, w: w, taken: map[int][][2]int{}}
}

// fits reports whether [x, x+kw) is clear in a lane, keeping one column of gap
// so two sprites never touch either.
func (a *laneAlloc) fits(lane, x int) bool {
	if x < 0 || x+a.kw > a.w {
		return false
	}
	for _, s := range a.taken[lane] {
		if x < s[1]+1 && x+a.kw > s[0]-1 {
			return false
		}
	}
	return true
}

// place finds a spot near the preferred lane and column, or reports that the
// water is full. Preference first so a subagent keeps its place between frames,
// then outward, so a crowded scene degrades by moving rather than by stacking.
func (a *laneAlloc) place(prefLane, prefX int) (lane, x int, ok bool) {
	step := a.kw + 1
	for dl := 0; dl < a.lanes; dl++ {
		lane = (prefLane + dl) % a.lanes
		for k := 0; k*step < a.w; k++ {
			for _, dir := range []int{1, -1} {
				x = prefX + dir*k*step
				if x < 2 {
					x = 2 + ((2 - x) % max(1, a.w-a.kw-2))
				}
				if x+a.kw > a.w-2 {
					x = a.w - 2 - a.kw
				}
				if a.fits(lane, x) {
					a.taken[lane] = append(a.taken[lane], [2]int{x, x + a.kw})
					return lane, x, true
				}
				if k == 0 {
					break
				}
			}
		}
	}
	return 0, 0, false
}

// crabTiers is the litter's size ladder, and it is the CAT's ladder ported
// rather than a second one invented.
//
// It was the last "partial" on Hero's API table, and the cost was measured
// rather than assumed: folded through his own 341 hours of recordings, the
// litter runs to 38 subagents at its peak and a p90 of 14. One size fits 18
// sitters at 153 columns and only EIGHT at 80, the design target -- so at 80
// columns the crab dropped subagents 12.6% of the time the litter existed, up
// to 18 at once, while the cat shrank them and kept going.
//
// Same rungs, same thresholds, same hysteresis: TierFor is shared, so the two
// animals cannot drift apart.
var crabTiers = []struct {
	rows []string
	eyes [2]int
}{
	{Crablet, crabletEyeCells},
	{CrabletSmall, [2]int{1, 3}},
	{CrabletTiny, [2]int{1, 2}},
}

func (c *Cat) crabBitmap(t int) *Bitmap {
	if c.kitCache == nil {
		c.kitCache = map[int]*Bitmap{}
	}
	// Offset so the crab's rungs cannot collide with the cat's in the cache of
	// a companion that has been swapped mid-run.
	key := 100 + t
	if b, ok := c.kitCache[key]; ok {
		return b
	}
	b := ParseBitmap(crabTiers[t].rows)
	c.kitCache[key] = b
	return b
}

// drawCrabKittens is DrawKittens for Hero.
func (c *Cat) drawCrabKittens(l *canvas.Layer, px, py, n, w, seaTop, seaBot int, t float64, seed int64) int {
	if n <= 0 {
		return 0
	}
	var sitters, swimmers []int
	for i := 0; i < n; i++ {
		if swims(i, seed) {
			swimmers = append(swimmers, i)
		} else {
			sitters = append(sitters, i)
		}
	}
	drawn := c.drawCrabSwimmers(l, swimmers, w, seaTop, seaBot, t, seed)

	// One size for the whole litter, chosen by how many are on the SAND --
	// the same split the cat makes, and for the same reason: mixing sizes
	// reads as a jumble.
	ti := TierFor(len(sitters), c.kitTier)
	c.kitTier = ti
	body := c.crabBitmap(ti)
	eyeCells := crabTiers[ti].eyes
	cw, ch := body.W/2, body.H/4
	ky := py + 7 - ch
	var pend []pendingEyes
	for n, i := range sitters {
		// Leftward from the parent, into the open beach rather than the frame
		// edge -- the direction the cat's litter grows when mirrored.
		x := px - 1 - cw - n*(cw+1)
		if x < 0 {
			break
		}
		b := body
		if n%2 == 0 {
			b = body.Mirrored()
		}
		rows := b.ToQuadrant()
		plotRim(l, rows, x, ky)
		(&Sprite{Rows: rows, Body: c.coat}).Draw(l, x, ky)

		glyph := 'o'
		// Its own blink PERIOD, not merely its own phase, so a row of them
		// never falls into step.
		if math.Mod(t+HashF(i, 12, seed)*6.283, 4.2+HashF(i, 16, seed)*3) < 0.17 {
			glyph = '-'
		}
		pend = append(pend, pendingEyes{x + eyeCells[0], x + eyeCells[1], ky, glyph, 1, l})
		drawn++
	}
	// Bodies first, faces after: a neighbour's seam must not take an eye.
	for _, p := range pend {
		p.draw()
	}
	return drawn
}

// crabSpan is one placed swimmer: which subagent, which lane, and the columns
// it occupies including the two ripple cells.
type crabSpan struct {
	i, lane, x, y int
}

// crabSwimSpans decides where every swimmer goes, separately from drawing them.
//
// Split out for the same reason the cat's swimmerSpans is: "no two swimmers
// overlap" is a claim about placement, and a test that infers it from the
// painted frame measures a proxy -- run lengths in a row also catch two sprites
// in NEIGHBOURING lanes whose rims legally touch, which is not the defect.
func crabSwimSpans(idx []int, w, seaTop, seaBot int, seed int64) []crabSpan {
	kw, kh := len(CrabletSwim[0])/2, len(CrabletSwim)/4
	top := seaTop + 1
	if seaBot <= top {
		return nil
	}
	lanes := (seaBot - top) / kh
	if lanes < 1 {
		return nil
	}
	alloc := newLaneAlloc(kw+2, lanes, w)
	var out []crabSpan
	for _, i := range idx {
		lane, slot, ok := alloc.place(
			int(HashF(i, 13, seed)*float64(lanes)),
			1+int(HashF(i, 14, seed)*float64(w-kw-6)))
		if !ok {
			continue // the water is full; better absent than stacked
		}
		out = append(out, crabSpan{i: i, lane: lane, x: slot + 1, y: top + lane*kh})
	}
	return out
}

// drawCrabSwimmers puts the rest in the water, stalks up.
func (c *Cat) drawCrabSwimmers(l *canvas.Layer, idx []int, w, seaTop, seaBot int, t float64, seed int64) int {
	if len(idx) == 0 {
		return 0
	}
	kw, kh := len(CrabletSwim[0])/2, len(CrabletSwim)/4
	var pend []pendingEyes
	drawn := 0
	for _, sp := range crabSwimSpans(idx, w, seaTop, seaBot, seed) {
		i, x, y := sp.i, sp.x, sp.y

		phase := HashF(i, 12, seed) * 6.283
		sink := 0
		if math.Sin(t*0.9+float64(x)*0.16+phase) > 0.35 {
			sink = 2
		}
		rows := ParseBitmap(sinkRows(CrabletSwim, sink)).ToQuadrant()
		plotRim(l, rows, x, y)
		(&Sprite{Rows: rows, Body: c.coat, Alpha: 1}).Draw(l, x, y)
		l.Plot(x-1, y+kh-1, '~', c.coat, 0.5)
		l.Plot(x+kw, y+kh-1, '~', c.coat, 0.5)

		glyph := 'o'
		if math.Mod(t+phase, 4.2+HashF(i, 16, seed)*3) < 0.17 {
			glyph = '-'
		}
		// The stalks stay up as the shell sinks, so the eyes never move with
		// the bob. That is the whole idea of the pose.
		pend = append(pend, pendingEyes{x + crabletSwimEyes[0], x + crabletSwimEyes[1], y, glyph, 1, l})
		drawn++
	}
	for _, p := range pend {
		p.draw()
	}
	return drawn
}

// drawCrabExits is DrawKittenExits for Hero.
//
// The cat's exit swims the surface toward the far edge and fades. A crab cannot
// do that and should not: it walks sideways into the surf and GOES UNDER, so
// the leaving is carried by the shell sinking as well as by travel, and the
// stalks are the last thing to go. Position and depth both, so a still frame
// reads it.
func (c *Cat) drawCrabExits(l *canvas.Layer, exits []float64, px, w, seaTop, seaBot int, t float64, seed int64) int {
	if len(exits) == 0 {
		return 0
	}
	kw, kh := len(CrabletSwim[0])/2, len(CrabletSwim)/4
	top := seaTop + 1
	if seaBot <= top {
		return 0
	}
	lanes := (seaBot - top) / kh
	if lanes < 1 {
		return 0
	}
	var pend []pendingEyes
	drawn := 0
	usedLane := map[int]bool{}
	for i, p := range exits {
		if p >= 1 {
			continue
		}
		lane := int(HashF(i, 17, seed) * float64(lanes))
		for k := 0; k < lanes && usedLane[lane]; k++ {
			lane = (lane + 1) % lanes
		}
		usedLane[lane] = true
		y := top + lane*kh

		startX, endX := px-10, -kw
		x := startX + int(float64(endX-startX)*p)
		if x+kw < 0 {
			continue
		}
		sink := int(p * 8)
		if sink >= len(CrabletSwim) {
			continue
		}
		alpha := 1.0
		if p > 0.5 {
			alpha = 1 - (p-0.5)*1.6
		}
		if alpha < 0.15 {
			alpha = 0.15
		}
		rows := ParseBitmap(sinkRows(CrabletSwim, sink)).ToQuadrant()
		plotRim(l, rows, x, y)
		(&Sprite{Rows: rows, Body: c.coat, Alpha: alpha}).Draw(l, x, y)
		l.Plot(x-1, y+kh-1, '~', c.coat, 0.5*alpha)
		l.Plot(x+kw, y+kh-1, '~', c.coat, 0.5*alpha)
		if sink < 6 {
			pend = append(pend, pendingEyes{x + crabletSwimEyes[0], x + crabletSwimEyes[1], y, 'o', 1, l})
		}
		drawn++
	}
	for _, p := range pend {
		p.draw()
	}
	return drawn
}

// sinkRows shifts a sprite's SOURCE ROWS down, letting the bottom clip off.
// Applied to the rows rather than to the parsed bitmap so the crest going under
// costs nothing but a slice; the kittens sink the same way, and for the same
// reason -- moving the whole sprite would put it in the lane above.
func sinkRows(src []string, sink int) []string {
	if sink <= 0 || sink >= len(src) {
		return src
	}
	blank := ""
	for range src[0] {
		blank += "."
	}
	out := make([]string, 0, len(src))
	for k := 0; k < sink; k++ {
		out = append(out, blank)
	}
	return append(out, src[:len(src)-sink]...)
}
