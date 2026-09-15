package scenes

import (
	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
)

// The study companions drawn on their own, for the page's cast. The crab and
// the cat draw themselves (companion.Cat.Draw); these two have no motion of
// their own, so a blink is the one thing a portrait can give them.

// DrawOwlPicked draws the owl as picked -- OwlPick in OwlCoat -- with its
// top-left cell at x, y. blink closes its eyes: the eye cells take the face's
// pale colour with a lid drawn on the lower row.
func DrawOwlPicked(c *canvas.Canvas, x, y int, blink bool) {
	alt := &OwlAlts[OwlPick]
	DrawOwl(c, alt, OwlCoat, x, y)
	if !blink {
		return
	}
	near := c.Near()
	for _, ex := range []int{alt.eyeL, alt.eyeR} {
		for r := 0; r < alt.eyeH; r++ {
			g := ' '
			if r == alt.eyeH-1 {
				g = '-'
			}
			for k := 0; k < alt.eyeW; k++ {
				near.PlotOn(x+ex+k, y+alt.eyeRow+r, g, owlDark, owlLight, 1)
			}
		}
	}
}

// DrawCandidate draws a study companion through the cat's pipeline with its
// top-left cell at x, y. blink closes its eyes.
func DrawCandidate(c *canvas.Canvas, a *Animal, x, y int, faceLeft, blink bool) {
	glyph := a.eyeGlyph
	if blink {
		glyph = '-'
	}
	drawAnimal(c.Near(), a.body, 24, 28, a.Name, companion.Coats[a.Coat],
		a.eyes, a.eyeRow, glyph, a.nose, a.noseRow, x, y, faceLeft)
}
