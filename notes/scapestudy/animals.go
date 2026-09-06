package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// Every companion below goes through the cat's own pipeline: a 24x28 bitmap
// drawn at half width and quarter height as quadrant glyphs, one coat colour,
// a one-cell rim, and eyes plotted as characters into gaps the bitmap leaves.
// The young are 12x16, the kittens' size. Authored facing right and mirrored
// on the right-hand side of the frame, like the cat.
type animal struct {
	name, coat, note string
	body, young      []string
	eyes, yeyes      [2]int // eye cells, left to right in the authored frame; -1 = none
	eyeRow, yeyeRow  int    // the cell row the eyes sit on
	eyeGlyph         rune
	nose             int // a nose cell, or -1
	noseRow          int
}

var eyeShine = term.RGB{R: 168, G: 236, B: 176} // the cat's

var animals = []animal{
	{name: "Owl", coat: "slate",
		note: "Top of the ranking. Night-native, sits still, blinks; the tufts and the round head survive the quadrant halving, and the big eyes are two full cells. Owlets as the litter.",
		body: owlBody, young: owlet, eyes: [2]int{2, 9}, eyeRow: 2, yeyes: [2]int{1, 3}, yeyeRow: 1, eyeGlyph: 'O', nose: 5, noseRow: 3},
	{name: "Rabbit", coat: "fog",
		note: "The ears are the whole silhouette and they carry the worried pose for free (flat ears). Kits as the litter. The pale coat is close to the cat's; a darker one would separate them.",
		body: rabbitBody, young: kit, eyes: [2]int{2, 9}, eyeRow: 3, yeyes: [2]int{1, 3}, yeyeRow: 1, eyeGlyph: 'o', nose: -1},
	{name: "Frog", coat: "sage",
		note: "Ranked for the rainy window: the eyes are domes on top rather than gaps in a face, so the silhouette reads at any size, and the litter swims (froglets here, tadpoles later).",
		body: frogBody, young: froglet, eyes: [2]int{2, 9}, eyeRow: 0, yeyes: [2]int{1, 4}, yeyeRow: 0, eyeGlyph: 'o', nose: -1},
	{name: "Hedgehog", coat: "charcoal",
		note: "A profile animal: one eye, a snout, a jagged back. Reads best small and still, which is what the resting state is. Hoglets as the litter.",
		body: hedgehogBody, young: hoglet, eyes: [2]int{8, -1}, eyeRow: 4, yeyes: [2]int{4, -1}, yeyeRow: 2, eyeGlyph: 'o', nose: 11, noseRow: 4},
	{name: "Otter", coat: "taupe",
		note: "Water-native for the shore: sits upright, the tail along the sand. Pups as the litter, and they would swim where the kittens do. Taupe, because the honest brown is the banned orange.",
		body: otterBody, young: pup, eyes: [2]int{4, 9}, eyeRow: 1, yeyes: [2]int{2, 4}, yeyeRow: 0, eyeGlyph: 'o', nose: -1},
}

func mustBitmap(rows []string, w, h int, name string) *companion.Bitmap {
	if len(rows) != h {
		panic(fmt.Sprintf("%s: %d rows, want %d", name, len(rows), h))
	}
	for i, r := range rows {
		if len(r) != w {
			panic(fmt.Sprintf("%s row %d: %d wide, want %d", name, i, len(r), w))
		}
	}
	return companion.ParseBitmap(rows)
}

// drawAnimal draws one body through the pipeline at cell (x, y).
func drawAnimal(near *canvas.Layer, rows []string, w, h int, name string, coat term.RGB, eyes [2]int, eyeRow int, glyph rune, nose, noseRow int, x, y int, mirror bool) {
	bm := mustBitmap(rows, w, h, name)
	if mirror {
		bm = bm.Mirrored()
	}
	q := bm.ToQuadrant()
	companion.PlotRim(near, q, x, y)
	(&companion.Sprite{Rows: q, Body: coat}).Draw(near, x, y)
	cw := w / 2
	at := func(cell int) int {
		if mirror {
			return x + cw - 1 - cell
		}
		return x + cell
	}
	for _, e := range eyes {
		if e >= 0 {
			near.Plot(at(e), y+eyeRow, glyph, eyeShine, 1)
		}
	}
	if nose >= 0 {
		near.Plot(at(nose), y+noseRow, 'v', term.RGB{R: 38, G: 38, B: 44}, 1)
	}
}

// companionFrame is the shore at an hour with one animal where the cat sits
// and two of its young where the kittens sit, laid out exactly as live.go
// lays the cat out at 80x24: four columns off the right edge, two rows off
// the bottom, the litter growing leftward.
func companionFrame(a *animal, tod, t float64, seed int64) *canvas.Canvas {
	return shoreFrame(tod, t, seed, func(c *canvas.Canvas, sh *scape.Shore) {
		near := c.Near()
		coat := companion.Coats[a.coat]
		px, py := c.W-12-4, c.H-2-7
		drawAnimal(near, a.body, 24, 28, a.name, coat, a.eyes, a.eyeRow, a.eyeGlyph, a.nose, a.noseRow, px, py, true)
		ky := py + 7 - 4
		for i := 0; i < 2; i++ {
			x := px - 1 - 6 - i*7
			drawAnimal(near, a.young, 12, 16, a.name+" young", coat, a.yeyes, a.yeyeRow, 'o', -1, 0, x, ky, i == 0)
		}
	})
}

// catFrame is the reference: the shipped cat and two real kittens.
func catFrame(tod, t float64, seed int64) *canvas.Canvas {
	return shoreFrame(tod, t, seed, func(c *canvas.Canvas, sh *scape.Shore) {
		cat := companion.NewCat()
		cat.FaceLeft(true)
		px, py := c.W-12-4, c.H-2-7
		cat.Draw(c.Near(), px, py, t, companion.Working)
		cat.DrawKittens(c.Near(), c.Mid(), px, py, 2, c.W-1, int(float64(c.H)*0.42)+1, sh.SandTop()-2, t, seed)
	})
}

func shoreFrame(tod, t float64, seed int64, draw func(c *canvas.Canvas, sh *scape.Shore)) *canvas.Canvas {
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.5, TimeOfDay: tod, ContextUsed: 0.3}
	for k := 0; k < 8; k++ {
		sh.Update(c, t+float64(k)/20, act)
	}
	draw(c, sh)
	return c
}

var owlBody = []string{
	"...##..............##...",
	"...###............###...",
	"..#####..........#####..",
	"..####################..",
	"..####################..",
	".######################.",
	".######################.",
	".######################.",
	".###..############..###.",
	".###..############..###.",
	".######################.",
	".######################.",
	".######################.",
	"..####################..",
	"....################....",
	".....##############.....",
	"....################....",
	"....################....",
	"....################....",
	"....################....",
	"....################....",
	"....################....",
	"....################....",
	".....##############.....",
	"......############......",
	".......##########.......",
	"......###......###......",
	"......###......###......",
}

var rabbitBody = []string{
	".....###........###.....",
	".....###........###.....",
	".....###........###.....",
	".....###........###.....",
	".....###........###.....",
	".....###........###.....",
	".....####......####.....",
	"....################....",
	"...##################...",
	"...##################...",
	"...##################...",
	"...##################...",
	"...##..############..##.",
	"...##..############..##.",
	"...##################...",
	"...##################...",
	"....################....",
	".....##############.....",
	"....################....",
	"...##################...",
	"..####################..",
	"..####################..",
	"..####################..",
	"..####################..",
	"..####################..",
	"...##################...",
	"..#####..######..#####..",
	"..#####..######..#####..",
}

var frogBody = []string{
	"..######........######..",
	"..######........######..",
	"..##..##........##..##..",
	"..##..##........##..##..",
	"..######........######..",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	"..####################..",
	"..####################..",
	"..####################..",
	"..####################..",
	"..####################..",
	"..####################..",
	"...##################...",
	"...##################...",
	"...##################...",
	"....################....",
	".....##############.....",
	"...###..########..###...",
	"..####..########..####..",
	".#####..########..#####.",
	".####....######....####.",
	".####....######....####.",
	".###......####......###.",
}

var hedgehogBody = []string{
	"........#...#...#.......",
	".......##..###..##......",
	"......####.###.####.....",
	".....##############.....",
	"....################....",
	"...##################...",
	"..####################..",
	"..####################..",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".######################.",
	".###############..######",
	".###############..######",
	".#######################",
	".#######################",
	".######################.",
	".#####################..",
	".####################...",
	".###################....",
	"..##################....",
	"..##################....",
	"...####..#####..####....",
	"...####..#####..####....",
}

var otterBody = []string{
	".........##.....##......",
	"........############....",
	".......##############...",
	".......##############...",
	".......#..########..#...",
	".......#..########..#...",
	".......##############...",
	".......##############...",
	"........############....",
	".........##########.....",
	"..........########......",
	"..........########......",
	".........##########.....",
	"........############....",
	".......##############...",
	".......##############...",
	"......################..",
	"......################..",
	"......################..",
	"......################..",
	"......################..",
	"......################..",
	"......################..",
	"......################..",
	"###...################..",
	"#####.################..",
	"#######..######..#####..",
	"..#####..######..#####..",
}

var owlet = []string{
	".##......##.",
	"..########..",
	".##########.",
	".##########.",
	".##..##..##.",
	".##..##..##.",
	".##########.",
	".##########.",
	"..########..",
	"..########..",
	"..########..",
	"..########..",
	"...######...",
	"...######...",
	"..##....##..",
	"..##....##..",
}

var kit = []string{
	"..##....##..",
	"..##....##..",
	"..##....##..",
	"..##....##..",
	".##########.",
	".##..##..##.",
	".##..##..##.",
	".##########.",
	"..########..",
	"...######...",
	"..########..",
	".##########.",
	".##########.",
	".##########.",
	".##########.",
	"..##....##..",
}

var froglet = []string{
	".###....###.",
	".#.#....#.#.",
	".###....###.",
	".##########.",
	".##########.",
	".##########.",
	".##########.",
	".##########.",
	"..########..",
	"..########..",
	"...######...",
	"....####....",
	"..##.####.##",
	".##..####..#",
	".##..####..#",
	"##....##...#",
}

var hoglet = []string{
	"....#..#....",
	"...##.##.#..",
	"..########..",
	".#########..",
	".##########.",
	".##########.",
	".##########.",
	".##########.",
	".########.##",
	".########.##",
	".###########",
	".##########.",
	".#########..",
	".########...",
	"..##.###.#..",
	"..##.###.#..",
}

var pup = []string{
	"....##..##..",
	"...########.",
	"...#..##..#.",
	"...#..##..#.",
	"...########.",
	"....######..",
	".....####...",
	"....######..",
	"...########.",
	"...########.",
	"...########.",
	"...########.",
	"...########.",
	"##.########.",
	"###.#######.",
	"####.##..##.",
}
