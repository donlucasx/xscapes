package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The sea and the beach are backdrop: the water says how much work, the sand
// carries the last few actions. Neither is meant to shimmer.
//
// Both ramps used to be measured from each column's own instantaneous
// waterline, so every column got its own gradient and the anchor moved with the
// swell. Measured at 120x26: a single row carried up to 87 distinct colours, and
// on a 256-colour terminal those near-identical browns quantise apart into
// olive, dusty red and grey -- which is why the beach read as blotches rather
// than sand, and would not sit still.
//
// The waterline band itself still moves. That is the tide, and it is the one
// place motion belongs.

func scapeFrames(n int) []*canvas.Canvas {
	sh := NewShore(7, false)
	var out []*canvas.Canvas
	for i := 0; i < n; i++ {
		c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, float64(i)*0.1, Activity{Level: 0.7, TimeOfDay: 0.3868})
		out = append(out, c)
	}
	return out
}

// rowTones counts the distinct background colours in each row.
func rowTones(c *canvas.Canvas) []int {
	out := make([]int, c.H)
	for y := 0; y < c.H; y++ {
		seen := map[term.RGB]bool{}
		for x := 0; x < c.W; x++ {
			seen[c.BGAt(x, y)] = true
		}
		out[y] = len(seen)
	}
	return out
}

func TestNoRowIsAConfettiOfNearIdenticalTones(t *testing.T) {
	c := scapeFrames(1)[0]
	tones := rowTones(c)
	worst, at := 0, 0
	for y, n := range tones {
		if n > worst {
			worst, at = n, y
		}
	}
	t.Logf("worst row is %d with %d distinct background tones", at, worst)
	// 87 before this was fixed. The waterline band legitimately varies per
	// column, so this is not one; it is a guard against the ramps going back
	// to being per-column.
	if worst > 40 {
		t.Errorf("row %d carries %d distinct tones, want at most 40", at, worst)
	}
}

// openSea returns the rows that are pure depth: below the horizon, above the
// band the swell moves through. Derived from the shore's own geometry rather
// than hardcoded, because the layout moves when the writing band changes size
// and a test pinned to row numbers starts measuring the wrong thing.
func openSea(sh *Shore, c *canvas.Canvas) (from, to int) {
	lo := c.H
	for _, e := range sh.lastEdge {
		if int(e) < lo {
			lo = int(e)
		}
	}
	return int(float64(c.H)*0.42) + 1, lo - 2
}

func TestTheSeaBandsAreFlatAcrossARow(t *testing.T) {
	sh := NewShore(7, false)
	c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh.Update(c, 0.1, Activity{Level: 0.7, TimeOfDay: 0.3868})
	tones := rowTones(c)
	from, to := openSea(sh, c)
	busy := 0
	for y := from; y <= to; y++ {
		if tones[y] > 2 {
			busy++
			t.Logf("sea row %d has %d tones", y, tones[y])
		}
	}
	if busy > 0 {
		t.Errorf("%d open-sea rows are not flat; depth is a property of the row, not of the column", busy)
	}
}

func TestTheBackdropHoldsStillBetweenFrames(t *testing.T) {
	sh := NewShore(7, false)
	var fr []*canvas.Canvas
	for i := 0; i < 24; i++ {
		c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, float64(i)*0.1, Activity{Level: 0.7, TimeOfDay: 0.3868})
		fr = append(fr, c)
	}
	from, to := openSea(sh, fr[0])
	changed, total := 0, 0
	for i := 1; i < len(fr); i++ {
		for y := from; y <= to; y++ { // open sea, above the tide
			for x := 0; x < fr[i].W; x++ {
				total++
				if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
					changed++
				}
			}
		}
	}
	pct := 100 * float64(changed) / float64(total)
	t.Logf("open-sea cells changing between frames: %.2f%%", pct)
	// 58.60% before the ramps were anchored, 3.73% after -- sixteen times less
	// motion in the backdrop. What is left is the mean waterline rising and
	// falling, which is the sea's whole job: it says how hard the agent is
	// working. That must keep moving.
	if pct > 8 {
		t.Errorf("the open sea is restless again: %.2f%% of its cells change every frame, want under 8%%", pct)
	}
}

// The writing band is the page the agent's work is written on: the bottom rows,
// held out of the water, one flat tone. Before it existed the swell reached into
// them and the text sat on waves -- "because the sand changes colors its a bit
// distracting from the bottom 4 lines".

func TestWritingBandIsOneFlatToneAndHoldsStill(t *testing.T) {
	fr := scapeFrames(24)
	c := fr[0]
	top := c.H - DefaultWriteRows

	tones := map[term.RGB]bool{}
	for y := top; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			tones[c.BGAt(x, y)] = true
		}
	}
	if len(tones) != 1 {
		t.Errorf("writing band uses %d tones, want exactly 1", len(tones))
	}

	for i := 1; i < len(fr); i++ {
		for y := top; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				if fr[i].BGAt(x, y) != fr[i-1].BGAt(x, y) {
					t.Fatalf("writing band changed at row %d col %d between frames", y, x)
				}
			}
		}
	}
}

// The swell must never reach the writing band, at any activity level -- a busy
// sea is exactly when the most is being written.
func TestTheWaterNeverReachesTheWritingBand(t *testing.T) {
	for _, level := range []float64{0, 0.3, 0.6, 0.9, 1.0} {
		sh := NewShore(7, false)
		for i := 0; i < 60; i++ {
			c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			sh.Update(c, float64(i)*0.1, Activity{Working: true, Level: level, TimeOfDay: 0.3868})
			top := c.H - DefaultWriteRows
			for _, e := range sh.lastEdge {
				if e > float64(top)-1 {
					t.Fatalf("level %.1f: waterline reached %.1f, past the writing band at row %d", level, e, top)
				}
			}
		}
	}
}

// The sea has to meet the sand along a ragged line, not a ruled one.
//
// Carving the writing band off the bottom BEFORE laying out the beach is what
// buys this: the mean waterline sits high enough that the whole swell fits
// above the band. Clamping it off the band afterwards put every trough on the
// same row -- "The line between ocean and sand beach is fully straight, while
// before it was more organic and you could see the waves crashing on the sand."
func TestTheShorelineIsRaggedWhenTheAgentIsWorking(t *testing.T) {
	sh := NewShore(7, false)
	c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for i := 0; i < 40; i++ {
		sh.Update(c, float64(i)*0.1, Activity{Working: true, Level: 0.8, TimeOfDay: 0.3868})
	}
	lo, hi := 999.0, -999.0
	for _, e := range sh.lastEdge {
		if e < lo {
			lo = e
		}
		if e > hi {
			hi = e
		}
	}
	t.Logf("waterline spans rows %.2f to %.2f (%.2f rows of relief)", lo, hi, hi-lo)
	if hi-lo < 1.5 {
		t.Errorf("shoreline spans only %.2f rows: it is a ruled line, not a shore", hi-lo)
	}
}

// The whole beach is one tone, so the only thing that reads against it is the
// water's edge. It also guards a seam that appeared when a graded beach sat
// above a flat band: the graded part ended its last row in black and drew a
// hard black line across the shore -- "the beach has a black line and is not
// working".
func TestAllTheSandIsOneToneWithNoBlackSeam(t *testing.T) {
	sh := NewShore(7, false)
	c := canvas.New(120, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for i := 0; i < 20; i++ {
		sh.Update(c, float64(i)*0.1, Activity{Working: true, Level: 0.7, TimeOfDay: 0.3868})
	}
	hi := 0.0
	for _, e := range sh.lastEdge {
		if e > hi {
			hi = e
		}
	}
	tones := map[term.RGB]bool{}
	for y := int(hi) + 2; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			bg := c.BGAt(x, y)
			tones[bg] = true
			if bandLuma(bg) < 30 {
				t.Fatalf("row %d col %d is near-black rgb(%d,%d,%d): a seam across the shore",
					y, x, bg.R, bg.G, bg.B)
			}
		}
	}
	if len(tones) != 1 {
		t.Errorf("the sand below the water uses %d tones, want exactly 1", len(tones))
	}
}
