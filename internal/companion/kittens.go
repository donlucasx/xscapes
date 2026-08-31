package companion

import (
	"math"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/term"
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
}

// kitTier is one rung of the size ladder.
//
// A kitten's tier comes from its INDEX, never from how many there are. If tier
// followed the count, spawning a sixth subagent would visibly shrink the five
// already sitting there -- a global relayout caused by a local event. Keyed on
// index instead, kitten #1 is large for its whole life and #9 is tiny for its
// whole life, and a new sibling changes nothing that already exists.
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
	{"kittenSmall", KittenSmall, eyeCols["kittenSmall"], 1, 0.86, 4},
	{"kittenTiny", KittenTiny, eyeCols["kittenTiny"], 2, 0.70, 2},
}

// tierFor: the first five are large, the next three small, the rest tiny.
func tierFor(i int) int {
	switch {
	case i < 5:
		return 0
	case i < 8:
		return 1
	}
	return 2
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
func (c *Cat) DrawKittens(near, mid *canvas.Layer, px, py, n, w int, t float64, seed int64) int {
	if n <= 0 {
		return 0
	}
	pw, ph := c.Size()
	x := px + pw + 1
	drawn := 0

	for i := 0; i < n; i++ {
		ti := tierFor(i)
		spec := kitTiers[ti]
		bm := c.kitBitmap(ti)
		kw, kh := bm.W/2, bm.H/4
		if x+kw > w {
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

		// Smaller tiers sit further up the beach and on the mid layer, so the
		// alpha model already in the renderer makes them recede.
		layer := near
		if ti > 0 {
			layer = mid
		}
		ky := py + ph - kh - spec.depth
		if HashF(i, 6, seed) > 0.6 {
			ky--
		}
		(&Sprite{Rows: f.ToQuadrant(), Body: furCol, Alpha: spec.alpha}).Draw(layer, x, ky)

		e1, e2 := x+spec.eyes[0], x+spec.eyes[1]
		if mirror {
			e1, e2 = x+kw-1-spec.eyes[0], x+kw-1-spec.eyes[1]
		}
		glyph := 'o'
		if math.Mod(t+phase, blinkT) < 0.17 {
			glyph = '-'
		}
		layer.Plot(e1, ky+1, glyph, kittenEye, spec.alpha)
		layer.Plot(e2, ky+1, glyph, kittenEye, spec.alpha)

		x += kw + 1
		drawn++
	}
	return drawn
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
