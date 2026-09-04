package companion

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

var kittenEye = term.RGB{R: 150, G: 214, B: 158} // dimmer than the parent's

// eyeCols is where each sprite's sockets land in CELL columns. Derived once and
// written down rather than computed from sprite width, which happened to be
// right for two scales and wrong for the third.
var eyeCols = map[string][2]int{
	"kittenSit":   {1, 3},
	"kittenCurl":  {1, 3},
	"kittenSmall": {1, 3},
	"kittenTiny":  {1, 2},
	"kittenSwim":  {1, 3},
}

// swims decides, once and for all, whether a given subagent is a swimmer.
// Keyed on index so a subagent does not climb in and out of the sea between
// frames. Roughly one in three, which is enough to make the beach look less
// like a police line-up without emptying it.
func swims(i int, seed int64) bool { return i > 0 && HashF(i, 11, seed) > 0.64 }

// kitTier is one rung of the size ladder.
//
// Every kitten on screen is the SAME tier: mixing sizes read as a jumble, and
// the smaller ones sat up the beach where they looked like they were floating
// in the sea. One size, chosen by how many there are.
type kitTier struct {
	name    string
	rows    []string
	eyes    [2]int
	depth   int     // 0 nearest; larger sits further up the beach
	alpha   float64 // near layer at 1.0, further tiers recede
	tailLen int     // how many segments; 0 for none
}

var kitTiers = []kitTier{
	{"kittenSit", KittenSit, eyeCols["kittenSit"], 0, 1.00, 6},
	{"kittenSmall", KittenSmall, eyeCols["kittenSmall"], 0, 1.00, 4},
	{"kittenTiny", KittenTiny, eyeCols["kittenTiny"], 0, 1.00, 2},
}

// TierFor picks one size for the whole litter from how many there are.
//
// The thresholds have HYSTERESIS: it takes six to shrink but only four to grow
// back. Without the gap, a count oscillating around a boundary would flip every
// kitten between sizes on alternate frames.
func TierFor(n, prev int) int {
	up := []int{6, 10}  // grow the litter -> shrink the kittens at these counts
	down := []int{4, 8} // shrink the litter -> grow them back only here
	tier := prev
	for tier < 2 && n >= up[tier] {
		tier++
	}
	for tier > 0 && n <= down[tier-1] {
		tier--
	}
	return tier
}

func (c *Cat) kitBitmap(t int) *Bitmap {
	if c.kitCache == nil {
		c.kitCache = map[int]*Bitmap{}
	}
	if b, ok := c.kitCache[t]; ok {
		return b
	}
	b := ParseBitmap(kitTiers[t].rows)
	c.kitCache[t] = b
	return b
}

// KittenSize is the largest tier's footprint, for callers laying out space.
func (c *Cat) KittenSize() (w, h int) {
	b := c.kitBitmap(0)
	return b.W / 2, b.H / 4
}

// DrawKittens draws one kitten per running subagent and returns how many fit.
//
// Every kitten runs its own breath, blink and tail on periods drawn from a hash
// of its index -- periods, not just phases, so they never fall into step and
// some coincidentally align, which is what a litter actually looks like.
// DrawKittens draws one kitten per running subagent and returns how many fit.
//
// Some sit on the sand and some swim. Swimmers use vertical space the beach
// does not have, so the two together hold far more than either alone, and the
// sea gives the litter somewhere to spread that still reads as one scene.
func (c *Cat) DrawKittens(near, mid *canvas.Layer, px, py, n, w, seaTop, seaBot int, t float64, seed int64) int {
	_ = mid // every kitten is on the near layer; kept for call-site stability
	if n <= 0 {
		return 0
	}
	pw, ph := c.Size()
	drawn := 0

	// Split first: the beach tier is chosen by how many are actually ON it.
	var sitters, swimmers []int
	for i := 0; i < n; i++ {
		if swims(i, seed) {
			swimmers = append(swimmers, i)
		} else {
			sitters = append(sitters, i)
		}
	}
	ti := TierFor(len(sitters), c.kitTier)
	c.kitTier = ti
	spec := kitTiers[ti]

	drawn += c.drawSwimmers(near, swimmers, px, py, w, seaTop, seaBot, t, seed)

	// Sitters walk AWAY from the parent: rightward when the companion is on
	// the left, leftward when it is on the right. Either way the litter grows
	// into the open beach rather than into the frame edge.
	x := px + pw + 1
	if c.mirror {
		x = px - 1
	}
	var pend []pendingEyes
	for _, i := range sitters {
		bm := c.kitBitmap(ti)
		kw, kh := bm.W/2, bm.H/4
		if c.mirror {
			x -= kw
			if x < 0 {
				break
			}
		} else if x+kw > w {
			break
		}

		// Per-kitten timing. Periods vary as well as phases.
		breathT := 1.9 + HashF(i, 1, seed)*1.6
		blinkT := 3.4 + HashF(i, 2, seed)*3.8
		tailT := 1.3 + HashF(i, 3, seed)*1.9
		phase := HashF(i, 4, seed) * 6.283
		mirror := HashF(i, 5, seed) > 0.5

		f := bm.Blank()
		lift := 0
		if math.Sin(2*math.Pi*t/breathT+phase) > 0 {
			lift = 2
		}
		for sy := 0; sy < f.H; sy++ {
			for sx := 0; sx < f.W; sx++ {
				if bm.at(sx, sy-lift) {
					f.Set(sx, sy)
				}
			}
		}
		c.kitTail(f, spec.tailLen, bm.W, math.Sin(t/tailT+phase), lift)
		if mirror {
			f = f.Mirrored()
		}

		// Every kitten stands on the sand with the parent, on the near layer at
		// full alpha. Pushing the small ones up the beach on the mid layer made
		// them look like they were floating in the water, and washed them out
		// until they read as grey blocks rather than cats.
		layer := near
		ky := py + ph - kh
		rows := f.ToQuadrant()
		plotRim(layer, rows, x, ky)
		(&Sprite{Rows: rows, Body: furCol, Alpha: spec.alpha}).Draw(layer, x, ky)

		e1, e2 := x+spec.eyes[0], x+spec.eyes[1]
		if mirror {
			e1, e2 = x+kw-1-spec.eyes[0], x+kw-1-spec.eyes[1]
		}
		glyph := 'o'
		if math.Mod(t+phase, blinkT) < 0.17 {
			glyph = '-'
		}
		// The eyes go in a SECOND pass, after every body is down.
		//
		// plotRim clears a ring around each sprite so overlapping kittens read
		// as separate rather than as one malformed shape -- and drawn inline,
		// the ring of kitten k+1 was landing on the EYES of kitten k, which had
		// already been plotted. Measured: twelve kittens, twenty eyes. A face
		// with one eye does not read as a face, and the litter is the subagent
		// count, so a litter nobody can read is a channel that is not there.
		//
		// A body that genuinely covers another's face still hides it, which is
		// occlusion and correct. Only the one-cell seam stops eating them.
		pend = append(pend, pendingEyes{e1, e2, ky + 1, glyph, spec.alpha, layer})

		if c.mirror {
			x -= 1 // the width was already taken off before drawing
		} else {
			x += kw + 1
		}
		drawn++
	}
	for _, p := range pend {
		p.draw()
	}
	return drawn
}

// pendingEyes is one kitten's face, held back until every body in its group is
// on the layer. See the comment where they are queued.
type pendingEyes struct {
	x1, x2, y int
	glyph     rune
	alpha     float64
	layer     *canvas.Layer
}

func (p pendingEyes) draw() {
	p.layer.Plot(p.x1, p.y, p.glyph, kittenEye, p.alpha)
	p.layer.Plot(p.x2, p.y, p.glyph, kittenEye, p.alpha)
}

// kitTail scales the parent's tail down to whatever the tier can carry. At four
// cells wide it is two pixels, which is enough to flick.
func (c *Cat) kitTail(b *Bitmap, segs, srcW int, wag float64, lift int) {
	if segs <= 0 {
		return
	}
	rootX := srcW - 3
	rootY := b.H - 4
	for i := 0; i <= segs; i++ {
		f := float64(i) / float64(segs)
		x := rootX + int(2.0*f+1.1*math.Sin(wag*1.5+f*2.2)+0.5)
		y := rootY - int(float64(segs)*0.8*f+0.5) + lift
		b.Set(x, y)
	}
}

// drawSwimmers puts kittens in the sea, bobbing on the swell and waddling
// slowly sideways.
//
// Positions are SLOTTED, not random. Each swimmer gets a lane, and within a
// lane the width is divided evenly so every swimmer owns a slot; the waddle is
// then bounded by whatever slack is left in that slot. Random home columns
// looked fine at three swimmers and piled them on top of each other at twelve.
func (c *Cat) drawSwimmers(l *canvas.Layer, idx []int, px, py, w, seaTop, seaBot int, t float64, seed int64) int {
	if len(idx) == 0 {
		return 0
	}
	if c.swim == nil {
		c.swim = ParseBitmap(KittenSwim)
	}
	kw, kh := c.swim.W/2, c.swim.H/4
	// Open water runs from just under the horizon down to a clear row above the
	// sand, so a swimmer in the lowest lane cannot touch a sitter's ears -- and
	// not past the shore's own waterline either. The lanes used to end a fixed
	// distance above the companion's row, which on a tall scape is BELOW the
	// water: he photographed a kitten swimming on the wet sand where the tide
	// meets the shore. seaBot is the last row that is still water everywhere;
	// the caller takes it from the shore, which knows where its sand starts.
	_, ph := c.Size()
	top := seaTop + 1
	bot := py + ph - 4
	if seaBot > 0 && seaBot < bot {
		bot = seaBot
	}
	// Lanes must be at least a swimmer tall. At one row apart, any two
	// swimmers in neighbouring lanes shared a row, and if their columns
	// crossed they interleaved into one broken shape.
	lanes := (bot - top) / kh
	if lanes < 1 {
		return 0
	}

	// Bucket by lane, keeping index order so the layout is deterministic.
	byLane := make([][]int, lanes)
	for _, i := range idx {
		ln := int(HashF(i, 13, seed) * float64(lanes))
		if ln >= lanes {
			ln = lanes - 1
		}
		byLane[ln] = append(byLane[ln], i)
	}

	// Start clear of the parent. Swimmers are drawn on the same layer and after
	// it, so anything overlapping its footprint paints straight over its face.
	pw, _ := c.Size()
	x0 := px + pw + 1
	span := w - x0 - 1
	if c.mirror {
		// Open water is everything to the LEFT of the companion.
		x0 = 1
		span = px - 2
	}
	if span < kw+2 {
		return 0
	}
	drawn := 0
	var pend []pendingEyes

	for ln := 0; ln < lanes; ln++ {
		members := byLane[ln]
		if len(members) == 0 {
			continue
		}
		slot := span / len(members)
		if slot < kw+1 {
			// Lane is oversubscribed; take what fits and drop the rest.
			members = members[:max(1, span/(kw+1))]
			slot = span / len(members)
		}
		for k, i := range members {
			phase := HashF(i, 12, seed) * 6.283
			slack := (slot - kw) / 2
			if slack > 3 {
				slack = 3
			}
			if slack < 0 {
				slack = 0
			}
			x := x0 + k*slot + (slot-kw)/2 + int(float64(slack)*math.Sin(t*0.32+phase))
			y := top + ln*kh
			if x < 1 || x+kw >= w || y < top || y+kh > bot {
				continue
			}

			// The bob happens INSIDE the sprite box, not by moving it: two
			// source rows is half a character cell, and the chin clipping off
			// the bottom reads as the cat settling into the water. Moving the
			// whole sprite up a row would put it in the lane above.
			sink := 0
			if math.Sin(t*0.9+float64(x)*0.16+phase) > 0.35 {
				sink = 2
			}
			f := c.swim.Blank()
			for sy := 0; sy < f.H; sy++ {
				for sx := 0; sx < f.W; sx++ {
					if c.swim.at(sx, sy-sink) {
						f.Set(sx, sy)
					}
				}
			}
			if HashF(i, 15, seed) > 0.5 {
				f = f.Mirrored()
			}
			rows := f.ToQuadrant()
			plotRim(l, rows, x, y)
			(&Sprite{Rows: rows, Body: furCol, Alpha: 1}).Draw(l, x, y)

			e := eyeCols["kittenSwim"]
			glyph := 'o'
			if math.Mod(t+phase, 4.2+HashF(i, 16, seed)*3) < 0.17 {
				glyph = '-'
			}
			l.Plot(x-1, y+kh-1, '~', furCol, 0.5)
			l.Plot(x+kw, y+kh-1, '~', furCol, 0.5)
			pend = append(pend, pendingEyes{x + e[0], x + e[1], y + 1, glyph, 1, l})
			drawn++
		}
	}
	// Bodies first, faces after: a neighbour's seam must not take an eye.
	for _, p := range pend {
		p.draw()
	}
	return drawn
}

// plotRim clears a one-cell ring around a sprite's silhouette before it is
// drawn, so whatever is behind it is visibly cut rather than merging into it.
//
// Two cream sprites at the same alpha overlapping on the same layer read as one
// malformed shape, not as depth. Painter order alone does not fix that -- the
// front sprite has to leave a seam. The ring also parts the water slightly
// around each swimmer, which is the right thing for something floating in it.
func plotRim(l *canvas.Layer, rows []string, x, y int) {
	h := len(rows)
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	filled := func(cx, cy int) bool {
		if cy < 0 || cy >= h {
			return false
		}
		r := []rune(rows[cy])
		return cx >= 0 && cx < len(r) && r[cx] != ' '
	}
	for cy := -1; cy <= h; cy++ {
		for cx := -1; cx <= w; cx++ {
			if filled(cx, cy) {
				continue
			}
			adj := false
			for dy := -1; dy <= 1 && !adj; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if filled(cx+dx, cy+dy) {
						adj = true
						break
					}
				}
			}
			if adj {
				l.Plot(x+cx, y+cy, ' ', term.RGB{}, 1)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
