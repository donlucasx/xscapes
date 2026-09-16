package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scenes"
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
	scenes.DrawOwlPose(c, owlX, owlY, t, st.Pose)
	if st.Kittens > 0 {
		// The litter keeps out of the fire and its smoke: the fire sits at
		// three eighths of the width and its light reaches a few cells.
		scenes.DrawOwlets(c, owlX, grassY, st.Kittens, 3*c.W/8+6)
	}
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
		// made whole first, keeping the cell's mean colour, so the letters
		// win. Seen in the first screenshot: "allow Bash?" with its a, o and
		// w eaten by the treeline.
		for dy, row := range rows {
			for dx := range []rune(row) {
				c.SetBG(x+dx, y+dy, c.BGAt(x+dx, y+dy))
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
	txt, col := pct, moonLabelDim
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
	for i, r := range txt {
		c.Near().PlotOn(x+i, y, r, col, c.BGAt(x+i, y), 1)
	}
}
