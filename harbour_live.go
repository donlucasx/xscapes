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

// drawHarbour is drawScene for the harbour: the reducer's state over the
// harbour's painting. The cat sits on the sill where the composition puts
// the companion (lay.CatX), so the balloon follows the shore's arithmetic;
// the kittens are drawn by the scape itself, coming up over the sill.
func drawHarbour(c *canvas.Canvas, h *scenes.Harbour, cat *companion.Cat, lay layout, st reduce.State, t float64) {
	catX, catY, bandTop := h.Layout()
	drawHarbourReadout(c, h, st.Act.ContextUsed)
	if catY > scape.TokensRow {
		drawTokens(c, st.Act.Tokens)
	}
	// The walk-up: a nearer cat is a bigger one standing on the same sill,
	// so it grows UP from the sill, not down into the writing.
	if _, _, bh := cat.DrawnBox(c.W); bh > 7 {
		catY = max(0, catY-(bh-7))
	}
	cat.Draw(c.Near(), catX, catY, t, st.Pose)
	head := catX + cat.DrawnHeadCol(c.W)
	if st.Bubble != "" {
		rows, col := companion.DoneBubble(st.Bubble), bubbleCol
		if st.BubbleAsk {
			rows, col = companion.Bubble(st.Bubble), bubbleAskCol
		}
		if lay.Mirror {
			rows = companion.MirrorTail(rows)
		}
		x := bubbleX(rows, head, c.W)
		y := catY - len(rows)
		if y < 0 {
			y = 0
		}
		// The balloon's own ground, as on the vista: the city behind it is
		// greys and lights, any of which can be the ink's own colour.
		for dy, row := range rows {
			for dx := range []rune(row) {
				c.SetBG(x+dx, y+dy, term.Lerp(c.BGAt(x+dx, y+dy), term.RGB{}, balloonShade))
			}
		}
		(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).Draw(c.Near(), x, y)
	}
	drawSandInk(c, st.Tail, h.DeskColor(), bandTop, 2, c.W-2, harbourInkPale, harbourInkDark, true)
}

// The desk's inks: a warm paper white and a near-black, picked per row
// against the painted desk as on the beach.
var (
	harbourInkPale = term.RGB{R: 236, G: 228, B: 210}
	harbourInkDark = term.RGB{R: 22, G: 18, B: 14}
)

// drawHarbourReadout is the context number under the harbour's moon.
func drawHarbourReadout(c *canvas.Canvas, h *scenes.Harbour, used float64) {
	if used < ReadoutFrom {
		return
	}
	mx, my := h.MoonAt(used)
	pct := fmt.Sprintf("%.0f%%", (1-used)*100)
	txt, col := pct, bubbleCol
	if used >= ReadoutWarn {
		txt, col = pct+" left", moonLabelWarn
	}
	x, y := mx+1-len(txt)/2, my+2
	if x < 2 {
		x = 2
	}
	if x+len(txt) > c.W-2 {
		x = c.W - 2 - len(txt)
	}
	for i, r := range txt {
		g := term.Lerp(c.BGAt(x+i, y), term.RGB{}, balloonShade)
		c.SetBG(x+i, y, g)
		c.Near().PlotOn(x+i, y, r, col, g, 1)
	}
}
