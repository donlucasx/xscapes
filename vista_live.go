package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE BALLOON'S GROUND. His first live look at the vista (2026-09-16, 10:10):
// "prompt text gets lost/broken". The runes were all there; the INK was the
// ground's colour. The done knock's ink is grey 254 in the cube and the
// ranges under the balloon are snow, haze and rock on the same grey ramp
// (254, 246, 247, 244, 241, 236), so on the shore, where the balloon sits on
// a blue sky, a near-white ink never met its own colour, and on the vista it
// met it in 9 of 135 balloon glyphs at 125x28 and 12 of 135 at 80x24 -- the
// exact letters missing from his screenshot -- with most of the rest within
// 30 luma of it. No single ink works: snow and rock alternate cell by cell
// on the same row. So the balloon brings its own ground: every cell it
// covers is darkened by balloonShade toward black, keeping its hue, before
// the letters go on -- a tinted balloon the mountain still shows through.
//
// The criterion, measured over 6 geometries x 6 hours x both balloons x two
// texts (12,744 glyph cells, TestTheBalloonIsLegibleOnTheRanges): every
// glyph's drawn ink is at least balloonLegibleGap luma from its drawn
// ground, AFTER quantisation. Swept 2026-09-16: shade 0 min 0.0 · 0.30 min
// 50.2 (539 under 80) · 0.40 min 70.2 (70 under 80) · 0.45 min 90.2 (249
// under 100) · 0.50 min 90.8 (4 under 100) · 0.55 min 110.2, NONE under 100
// · 0.60 min 123.2. The shade is the smallest value that clears the floor
// everywhere, and the floor is where the last stragglers stop. The shore's
// balloon is untouched: it has never lost a letter, and its look is his.
var (
	balloonShade      = 0.55
	balloonLegibleGap = 100.0
)

// drawVista is drawScene for the vista: the same reducer state, the same
// balloon, the same sand writer, over the vista's own painting. The owl
// stands where the composition puts the companion (lay.CatX), so the litter
// and the balloon follow the shore's arithmetic without a second copy of it.
//
// What the vista does NOT do, and says so rather than pretending: the owl
// does not walk up to the screen (the ask is its wide eyes and the balloon),
// finished subagents do not fly off (they leave when the reducer drops them,
// after its dwell), and the companion does not pace -- an owl on a mound is
// still by nature, and its steps would be a count with nowhere to go.
func drawVista(c *canvas.Canvas, v *scenes.Vista, lay layout, st reduce.State, t float64) {
	owlX, owlY, grassY, bandTop := v.Layout()
	drawVistaReadout(c, v, st.Act.ContextUsed)
	// The spend counter, unless the owl's box reaches the counter's row:
	// under the vista's 16-row floor the owl sits in the sky's top-right
	// corner and would be drawn over it.
	if owlY > scape.TokensRow {
		drawTokens(c, st.Act.Tokens)
	}
	// The litter first, then the owl over it: an owlet's flight starts and
	// ends behind its parent. The litter sits beside the owl, around the
	// fire, or both (scenes.OwletPlace), each owlet flying in from the owl
	// when it starts and back when the reducer lets it go. Beside the owl it
	// keeps out of the fire and its smoke: the fire sits at three eighths of
	// the width and its light reaches a few cells.
	if st.Kittens > 0 || len(st.KittenExits) > 0 {
		scenes.DrawFlock(c, &v.Flock, st.Kittens, st.KittenExits, owlX, owlY, grassY, v.FireX(), 3*c.W/8+6, t, scenes.PickedOwletMotion(), scenes.OwlMotionPeriodOverride)
	}
	// The owl, with whatever motion he has picked from the round
	// (owl_anim.go); no pick draws today's owl exactly.
	scenes.DrawOwlMoving(c, owlX, owlY, t, st.Pose, scenes.PickedOwlMotion(st.Pose), scenes.OwlMotionPeriodOverride)
	if st.Bubble != "" {
		rows, col := companion.DoneBubble(st.Bubble), bubbleCol
		if st.BubbleAsk {
			rows, col = companion.Bubble(st.Bubble), bubbleAskCol
		}
		if lay.Mirror {
			rows = companion.MirrorTail(rows)
		}
		x := bubbleX(rows, owlX+scenes.OwlHeadCol, c.W)
		y := owlY - len(rows)
		if y < 0 {
			y = 0
		}
		// The balloon crosses the near ridge, whose edge is painted as
		// quarter-cells, and the canvas's rule is that a plain glyph over a
		// quarter-cell background loses to the quarters (a star must not
		// break the moon's disc). A balloon is TEXT: every cell it covers is
		// made whole first, so the letters win. Seen in the first
		// screenshot: "allow Bash?" with its a, o and w eaten by the
		// treeline. And made DARKER (balloonShade, above): whole was not
		// enough, because whole snow is the ink's own grey.
		for dy, row := range rows {
			for dx := range []rune(row) {
				c.SetBG(x+dx, y+dy, term.Lerp(c.BGAt(x+dx, y+dy), term.RGB{}, balloonShade))
			}
		}
		(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).Draw(c.Near(), x, y)
	}
	// The writing runs the whole band: the owl is above it, not beside it.
	drawSand(c, st.Tail, v.BandColor(), bandTop, 2, c.W-2)
}

// drawVistaReadout is the context number under the vista's moon, the same
// thresholds and colours as the shore's.
func drawVistaReadout(c *canvas.Canvas, v *scenes.Vista, used float64) {
	if used < ReadoutFrom {
		return
	}
	mx, my := v.MoonAt(used)
	pct := fmt.Sprintf("%.0f%%", (1-used)*100)
	// The quiet ink is the balloon's, not the shore's dim grey: on the
	// darkened ground below, a mid grey read 40 luma over snow. The warm
	// ink is the shore's.
	txt, col := pct, bubbleCol
	if used >= ReadoutWarn {
		txt, col = pct+" left", moonLabelWarn
	}
	x, y := mx+1-len(txt)/2, my+2
	if x < 1 {
		x = 1
	}
	if x+len(txt) > c.W-1 {
		x = c.W - 1 - len(txt)
	}
	// The same ground as the balloon (balloonShade): the number sits on the
	// ranges as the body sinks, and its dim grey was the snow's grey (his
	// crop, 2026-09-16 12:07, "gets lost on the mountains").
	for i, r := range txt {
		g := term.Lerp(c.BGAt(x+i, y), term.RGB{}, balloonShade)
		c.SetBG(x+i, y, g)
		c.Near().PlotOn(x+i, y, r, col, g, 1)
	}
}
