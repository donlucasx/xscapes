package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/term"
)

// harbourFrames is a frames on the harbour, as `xscapes scape harbour`
// leaves a running scape.
func harbourFrames(t testing.TB, w, h int) *frames {
	t.Helper()
	if tt, ok := t.(*testing.T); ok {
		tt.Setenv("XSCAPES_SCAPE", "harbour")
	}
	f := newFrames(w, h, 7, false, true, 0, 0)
	if f.harbour == nil {
		t.Fatal("XSCAPES_SCAPE=harbour did not select the harbour")
	}
	return f
}

func harbourState(level float64, pose companion.State, ctx, tod float64, kits int) reduce.State {
	st := pinnedState(level, ctx, tod)
	st.Pose = pose
	st.Kittens = kits
	st.Tail = []reduce.Line{{Text: "edit  internal/auth/handler.go  +18 -2"}, {Text: "test  go  ok 1.4s"}}
	switch pose {
	case companion.NeedsYou:
		st.Bubble, st.BubbleAsk = "allow Bash?", true
	case companion.Done:
		st.Bubble = "tests passed"
	}
	return st
}

// The glass is dry at idle and wet while working: the rain is the work.
func TestTheHarboursRainIsTheWork(t *testing.T) {
	count := func(level float64) int {
		f := harbourFrames(t, 100, 30)
		n := 0
		// Run a minute of frames at the level, so drops born at it fill the
		// pane, then count the drops' heads and trails on the glass.
		var st reduce.State
		for i := 0; i <= 600; i++ {
			st = harbourState(level, companion.Working, 0.2, 0.9, 0)
			f.compose(st, float64(i)/10)
		}
		_, _, band := f.harbour.Layout()
		for y := 0; y < band; y++ {
			for x := 0; x < f.c.W; x++ {
				cell := f.c.Mid().Cells[y*f.c.W+x]
				if r := cell.R; cell.Set && (r == 'o' || r == '|') {
					n++
				}
			}
		}
		return n
	}
	idle, busy := count(0), count(1)
	if idle > 20 || busy < 5*max(idle, 10) {
		t.Fatalf("drop cells on the glass: idle %d, flat out %d; want a dry pane at idle and a wet one at work", idle, busy)
	}
}

// Every window from the floor up composes without a panic, puts the cat on
// the sill and the writing under it.
func TestTheHarbourFitsEveryWindow(t *testing.T) {
	for _, g := range [][2]int{{60, 16}, {80, 24}, {100, 30}, {125, 28}, {153, 41}, {200, 60}} {
		f := harbourFrames(t, g[0], g[1])
		for _, pose := range []companion.State{companion.Resting, companion.Working, companion.NeedsYou, companion.Done, companion.Worried} {
			f.compose(harbourState(0.6, pose, 0.7, 0.85, 3), 12)
		}
		catX, catY, band := f.harbour.Layout()
		if catY+7 != band-1 {
			t.Errorf("%dx%d: cat rows %d..%d, sill %d", g[0], g[1], catY, catY+6, band-1)
		}
		if catX < 0 || catX+12 > g[0] {
			t.Errorf("%dx%d: cat at column %d", g[0], g[1], catX)
		}
	}
}

// XSCAPES_HARBOURPAGE=<dir> writes a page of frames to look at.
func TestHarbourPage(t *testing.T) {
	dir := os.Getenv("XSCAPES_HARBOURPAGE")
	if dir == "" {
		t.Skip("set XSCAPES_HARBOURPAGE=<dir>")
	}
	type shot struct {
		w, h      int
		tod, lvl  float64
		pose      companion.State
		kits      int
		ctx, secs float64
		label     string
	}
	shots := []shot{
		{80, 24, 0.55, 0, companion.Resting, 0, 0.1, 30, "80x24 afternoon, idle"},
		{125, 28, 0.80, 0.8, companion.Working, 2, 0.3, 40, "125x28 dusk, working, two kittens"},
		{125, 28, 0.95, 0.5, companion.NeedsYou, 2, 0.5, 50, "125x28 night, the ask"},
		{153, 41, 0.97, 0.9, companion.Done, 3, 0.8, 60, "153x41 night, done, three kittens"},
		{100, 30, 0.30, 0.3, companion.Worried, 1, 0.6, 70, "100x30 morning, worried"},
		{60, 18, 0.9, 0.6, companion.Working, 1, 0.45, 80, "60x18 night"},
	}
	var b strings.Builder
	pal := &canvas.HTMLPalette{}
	for _, s := range shots {
		f := harbourFrames(t, s.w, s.h)
		for i := 0; i <= 300; i++ {
			st := harbourState(s.lvl, s.pose, s.ctx, s.tod, s.kits)
			f.harbourCat.Approach(0.1, s.pose == companion.NeedsYou)
			f.compose(st, s.secs-30+float64(i)/10)
		}
		fmt.Fprintf(&b, "<h3>%s</h3>\n%s\n", s.label, f.c.HTMLFragmentClassed(14, term.Profile256, pal))
	}
	page := canvas.HTMLPage("harbour", "<style>"+pal.CSS()+"</style>"+b.String())
	if err := os.WriteFile(filepath.Join(dir, "harbour.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The walked-up cat stands on the sill: nothing of it reaches the desk.
func TestTheHarboursCatWalksUpOnTheSill(t *testing.T) {
	f := harbourFrames(t, 125, 28)
	for i := 0; i < 40; i++ {
		f.harbourCat.Approach(0.1, true)
	}
	if _, _, bh := f.harbourCat.DrawnBox(125); bh <= 7 {
		t.Skipf("the walk-up is off here (box height %d)", bh)
	}
	st := harbourState(0.5, companion.NeedsYou, 0.3, 0.9, 0)
	st.Tail = nil
	f.compose(st, 5)
	_, _, band := f.harbour.Layout()
	for y := band; y < f.c.H; y++ {
		for x := 0; x < f.c.W; x++ {
			if cell := f.c.Near().Cells[y*f.c.W+x]; cell.Set {
				t.Fatalf("the cat reaches the desk at %d,%d", x, y)
			}
		}
	}
}
