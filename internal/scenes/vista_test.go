package scenes

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// The layout holds its order at every size, and the owl stays in the frame.
func TestTheVistaLayoutHoldsAtEverySize(t *testing.T) {
	for _, g := range [][2]int{{80, 24}, {125, 28}, {124, 30}, {143, 27}, {107, 51}, {60, 20}, {40, 12}, {30, 8}} {
		lay := vistaLayoutFor(g[0], g[1], 0)
		if !(lay.lakeTop < lay.meadowTop && lay.meadowTop < lay.bandTop && lay.bandTop < g[1]) {
			t.Errorf("%dx%d: rows out of order: lake %d meadow %d band %d", g[0], g[1], lay.lakeTop, lay.meadowTop, lay.bandTop)
		}
		if g[1]-lay.bandTop < 3 {
			t.Errorf("%dx%d: the band has %d rows, want at least 3", g[0], g[1], g[1]-lay.bandTop)
		}
		// Below sixteen rows there is no room for a seven-row owl over a
		// three-row band and a meadow; it overlaps the sky and the scene
		// still draws. The shore's floor is 40x12; the vista's is 16 rows.
		if g[1] >= VistaMinRows && (lay.owlY < 0 || lay.owlY+7 > lay.bandTop+1) {
			t.Errorf("%dx%d: the owl's box rows %d..%d do not sit above the band at %d", g[0], g[1], lay.owlY, lay.owlY+6, lay.bandTop)
		}
		if lay.owlX+12 > g[0] {
			t.Errorf("%dx%d: the owl's box runs off the right edge (x %d)", g[0], g[1], lay.owlX)
		}
		// And it paints without panicking, at every hour and level.
		v := NewVista(7, false)
		for _, tod := range []float64{0, 0.24, 0.5, 0.78} {
			for _, lv := range []float64{0, 0.5, 1} {
				c := canvas.New(g[0], g[1], canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				v.Update(c, 1.0, scape.Activity{Level: lv, Working: lv > 0, TimeOfDay: tod, ContextUsed: 0.7, TodoDone: 9, TodoTotal: 12})
			}
		}
	}
	if lay := vistaLayoutFor(80, 24, 0); lay != vista80() {
		t.Errorf("80x24 does not reproduce the study's layout: %+v", lay)
	}
}

// Each of the owl's five states is a different picture: the eye cells,
// read off the frame, differ pairwise. The positive control is that the
// same state drawn twice reads the same.
func TestTheOwlsFiveStatesAreFiveDifferentFaces(t *testing.T) {
	face := func(st companion.State) string {
		c := canvas.New(40, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		DrawOwlPose(c, 10, 2, 1.0, st) // t=1.0: mid-blink-cycle, eyes open when working
		var s []byte
		for y := 2; y < 9; y++ {
			for x := 10; x < 22; x++ {
				r, fg, bg := c.ResolveAt(x, y, term.Profile256)
				s = append(s, byte(r), byte(fg.Index256()), byte(bg.Index256()))
			}
		}
		return string(s)
	}
	states := []companion.State{companion.Resting, companion.Working, companion.NeedsYou, companion.Done, companion.Worried}
	names := []string{"resting", "working", "needs-you", "done", "worried"}
	faces := make([]string, len(states))
	for i, st := range states {
		faces[i] = face(st)
		if faces[i] != face(st) {
			t.Errorf("%s drawn twice differs from itself", names[i])
		}
	}
	for i := range states {
		for j := i + 1; j < len(states); j++ {
			if faces[i] == faces[j] {
				t.Errorf("%s and %s draw the same face", names[i], names[j])
			}
		}
	}
	// The blink: working shuts its eyes for a quarter second in seven.
	c := canvas.New(40, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	DrawOwlPose(c, 10, 2, 7.1, companion.Working)
	blink := false
	for x := 10; x < 22; x++ {
		if r, _, _ := c.ResolveAt(x, 4, term.Profile256); r == '-' {
			blink = true
		}
	}
	if !blink {
		t.Errorf("the working owl does not blink at t=7.1")
	}
}

// The wind is the work: more of the scrub lies flat and the smoke leans
// further as the level rises. Coverage and position, read off the frame.
func TestTheWindRisesWithTheLevel(t *testing.T) {
	bent := func(level float64) int {
		c := canvas.New(125, 28, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		v := NewVista(7, false)
		v.Update(c, 1.0, scape.Activity{Level: level, Working: level > 0, TimeOfDay: 0.5})
		_, _, _, bandTop := v.Layout()
		n := 0
		for y := 0; y < bandTop; y++ {
			for x := 0; x < 125; x++ {
				if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == ',' || r == '_' {
					n++
				}
			}
		}
		return n
	}
	calm, half, gale := bent(0), bent(0.5), bent(1)
	t.Logf("flat scrub cells: calm %d, half %d, gale %d", calm, half, gale)
	if !(calm < half && half < gale) {
		t.Errorf("the flattened scrub does not rise with the level: %d, %d, %d", calm, half, gale)
	}
}

// The context sinks the moon, as on the shore, and never into the range.
func TestTheVistaMoonSinksWithTheContext(t *testing.T) {
	lay := vistaLayoutFor(125, 28, 0)
	fresh, full := lay.moonRow(0), lay.moonRow(1)
	if !(fresh < full) {
		t.Errorf("the moon does not sink: row %d fresh, %d full", fresh, full)
	}
	if full+1 >= lay.farBase/2 {
		t.Errorf("a full context puts the moon (rows %d..%d) into the far range at row %d", full, full+1, lay.farBase/2)
	}
}
