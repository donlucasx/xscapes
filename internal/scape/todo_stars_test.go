package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

func skyMarks(c *canvas.Canvas, hy int) (lit, pending int) {
	for y := 0; y < hy; y++ {
		for x := 0; x < c.W; x++ {
			cell := c.Near().Cells[y*c.W+x]
			if !cell.Set {
				continue
			}
			switch cell.R {
			case '*':
				lit++
			case '∘':
				pending++
			}
		}
	}
	return
}

func todoFrame(t *testing.T, tod float64, done, total int) (*canvas.Canvas, int) {
	t.Helper()
	c := canvas.New(80, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := NewShore(7, false)
	sh.MoonX = 0.28
	sh.Update(c, 3.0, Activity{Working: true, Level: 0.5, TimeOfDay: tod,
		ContextUsed: 0.3, TodoDone: done, TodoTotal: total})
	return c, int(float64(c.H) * 0.42)
}

// The sky says "n of N", not just n. A finished todo lights a star and an
// unfinished one leaves a mark, so the size of the job is visible as well as
// the progress through it.
func TestTheSkyCountsTheChecklist(t *testing.T) {
	for _, tc := range []struct{ done, total int }{{0, 5}, {2, 5}, {5, 5}, {1, 3}} {
		c, hy := todoFrame(t, 0.5, tc.done, tc.total)
		lit, pending := skyMarks(c, hy)
		t.Logf("%d of %d: %d lit, %d pending", tc.done, tc.total, lit, pending)
		if lit != tc.done || lit+pending != tc.total {
			t.Errorf("%d of %d: sky shows %d lit and %d pending", tc.done, tc.total, lit, pending)
		}
	}
}

// No list, no constellation. An empty sky and a list with nothing done are
// different states and must not look the same.
func TestNoChecklistDrawsNothing(t *testing.T) {
	c, hy := todoFrame(t, 0.5, 0, 0)
	if lit, pending := skyMarks(c, hy); lit+pending != 0 {
		t.Errorf("with no todo list the sky drew %d lit and %d pending marks", lit, pending)
	}
}

// ⭐ THE ONE THAT MATTERS. Completed todos are a fact about the AGENT; the star
// field's visibility is a fact about the WORLD, and StarVis is 0 at noon.
// Hanging the checklist on the clock would switch it off for the whole working
// day -- the exact bug moonVisFloor was added to fix, arriving again through a
// different channel. It has to read at every hour.
func TestTheChecklistIsLegibleAtEveryHour(t *testing.T) {
	lumaOf := func(c term.RGB) float64 {
		return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
	}
	for _, tod := range []float64{0, 0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875} {
		c, hy := todoFrame(t, tod, 4, 6)
		lit, _ := skyMarks(c, hy)
		if lit != 4 {
			t.Fatalf("tod %.3f: %d stars lit, want 4", tod, lit)
		}
		// Contrast is measured on what is PAINTED, through the same 256 path
		// the terminal uses, against the sky in the same row.
		worst := 999.0
		for y := 0; y < hy; y++ {
			for x := 0; x < c.W; x++ {
				if cell := c.Near().Cells[y*c.W+x]; !cell.Set || cell.R != '*' {
					continue
				}
				_, fg, _ := c.ResolveAt(x, y, term.Profile256)
				_, _, sky := c.ResolveAt(1, y, term.Profile256)
				if d := lumaOf(fg) - lumaOf(sky); d < worst {
					worst = d
				}
			}
		}
		t.Logf("tod %.3f: the dimmest lit star reads %+.1f luma above its sky", tod, worst)
		if worst < 40 {
			t.Errorf("tod %.3f: a lit todo star is only %+.1f luma above its sky -- the checklist is invisible at that hour",
				tod, worst)
		}
	}
}

// A star that moved between frames would encode nothing. Position is the
// channel; it has to hold still.
func TestTheConstellationHoldsItsPlace(t *testing.T) {
	var first []int
	for i := 0; i < 8; i++ {
		c := canvas.New(80, 26, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := NewShore(7, false)
		sh.MoonX = 0.28
		sh.Update(c, 3.0+float64(i)*0.7, Activity{Working: true, Level: 0.8,
			TimeOfDay: 0.5, ContextUsed: 0.3, TodoDone: 3, TodoTotal: 6})
		hy := int(float64(c.H) * 0.42)
		var at []int
		for y := 0; y < hy; y++ {
			for x := 0; x < c.W; x++ {
				if cell := c.Near().Cells[y*c.W+x]; cell.Set && cell.R == '*' {
					at = append(at, y*c.W+x)
				}
			}
		}
		if i == 0 {
			first = at
			continue
		}
		if len(at) != len(first) {
			t.Fatalf("frame %d: %d stars, want %d", i, len(at), len(first))
		}
		for j := range at {
			if at[j] != first[j] {
				t.Fatalf("frame %d: a star moved from cell %d to %d", i, first[j], at[j])
			}
		}
	}
}
