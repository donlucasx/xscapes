package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// When the moon moves, it leaves nothing behind.
//
// His daylight screengrab, 2026-09-04 10:49, 133x61: the sun ringed by dark
// notches. Sampled from the pixels, every notch was rgb(26,26,26) over
// rgb(180,180,180) -- the NIGHT sky's tone over the NIGHT moon's tone, in a
// blue sky at eleven in the morning. They were the moon's half-row tips from
// the session's first hours: the moon climbs as context is used, the sky is
// repainted every frame, and the cell's half-row record was never cleared, so
// a cell that had once been a tip showed that tip forever.
//
// One canvas, two frames far apart, the same way a session runs.
func TestTheMoonLeavesNoStaleTipsWhenItMoves(t *testing.T) {
	c := canvas.New(133, 27, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	sh.MoonX = 0.28
	// Night, context nearly empty: the moon low and grey.
	for k := 0; k < 4; k++ {
		sh.Update(c, 3.0+float64(k)/20, Activity{Working: true, Level: 0.5, TimeOfDay: 0.0245, ContextUsed: 0.06})
	}
	mx0, my0 := sh.MoonPos()
	// Late morning, context a third used: the moon higher, the sky blue.
	for k := 0; k < 4; k++ {
		sh.Update(c, 4.0+float64(k)/20, Activity{Working: true, Level: 0.5, TimeOfDay: 0.45, ContextUsed: 0.30})
	}
	mx1, my1 := sh.MoonPos()
	if my0 == my1 {
		t.Fatalf("the moon did not move (row %d both times); the test needs it to", my0)
	}
	chroma := func(col term.RGB) int {
		mx, mn := int(col.R), int(col.R)
		for _, v := range []int{int(col.G), int(col.B)} {
			if v > mx {
				mx = v
			}
			if v < mn {
				mn = v
			}
		}
		return mx - mn
	}
	// Every split cell in the sky must carry the sky's hue in its sky half.
	// A grey half in a blue sky is a leftover.
	hy := int(float64(c.H) * 0.42)
	stale := 0
	for y := 0; y < hy; y++ {
		for x := 0; x < c.W; x++ {
			ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
			if ch != '▀' {
				continue
			}
			if chroma(fg) == 0 && chroma(bg) == 0 {
				stale++
				if stale <= 6 {
					t.Errorf("cell (%d,%d): a split cell of grey #%02x%02x%02x over grey #%02x%02x%02x in a blue sky -- the old moon (%d,%d) left a tip behind; it is at (%d,%d) now",
						x, y, fg.R, fg.G, fg.B, bg.R, bg.G, bg.B, mx0, my0, mx1, my1)
				}
			}
		}
	}
	if stale > 6 {
		t.Errorf("... and %d more", stale-6)
	}
}
