package main

import (
	"fmt"
	"math"
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

// harbourLuma is the same luma the vista's legibility tests use.
func harbourLuma(c term.RGB) float64 { return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B) }

// The balloon's letters read on the city behind it, at every hour, size and
// both balloons, on the cube: the vista's floor (balloonLegibleGap).
func TestTheHarboursBalloonIsLegible(t *testing.T) {
	worst, at := 999.0, ""
	for _, g := range [][2]int{{80, 24}, {125, 28}, {100, 30}, {153, 41}, {60, 18}} {
		for _, hr := range []float64{0, 0.25, 0.42, 0.5, 0.78, 0.9} {
			for _, ask := range []bool{false, true} {
				f := harbourFrames(t, g[0], g[1])
				pose := companion.Done
				if ask {
					pose = companion.NeedsYou
					for i := 0; i < 40; i++ {
						f.harbourCat.Approach(0.1, true)
					}
				}
				st := harbourState(0.3, pose, 0.3, hr, 1)
				f.compose(st, 3)
				catX, catY, _ := f.harbour.Layout()
				if _, _, bh := f.harbourCat.DrawnBox(g[0]); bh > 7 {
					catY -= bh - 7
				}
				rows := companion.DoneBubble(st.Bubble)
				if ask {
					rows = companion.Bubble(st.Bubble)
				}
				rows = companion.MirrorTail(rows)
				x, y := bubbleX(rows, catX+f.harbourCat.DrawnHeadCol(g[0]), g[0]), max(0, catY-len(rows))
				for dy, row := range rows {
					for dx, want := range []rune(row) {
						if want == ' ' {
							continue
						}
						r, fg, bg := f.c.ResolveAt(x+dx, y+dy, term.Profile256)
						if r != want {
							t.Errorf("%dx%d %.2f ask=%v: balloon %q at (%d,%d) drawn as %q", g[0], g[1], hr, ask, want, x+dx, y+dy, r)
							continue
						}
						if gap := math.Abs(harbourLuma(fg) - harbourLuma(bg)); gap < worst {
							worst, at = gap, fmt.Sprintf("%dx%d %.2f ask=%v %q", g[0], g[1], hr, ask, want)
						}
					}
				}
			}
		}
	}
	t.Logf("smallest balloon ink-to-ground gap %.1f at %s", worst, at)
	if worst < balloonLegibleGap {
		t.Errorf("a balloon glyph sits %.1f luma from its ground (floor %.0f): %s", worst, balloonLegibleGap, at)
	}
}

// The writing on the desk reads at every hour: every glyph at least 40 luma
// from its ground on the cube, the vista band's floor.
func TestTheHarboursDeskWritingIsLegible(t *testing.T) {
	worst, at := 999.0, ""
	for _, g := range [][2]int{{80, 24}, {125, 28}, {153, 41}} {
		for _, hr := range []float64{0, 0.25, 0.5, 0.78, 0.9} {
			f := harbourFrames(t, g[0], g[1])
			st := harbourState(0.3, companion.Working, 0.3, hr, 0)
			f.compose(st, 3)
			_, _, band := f.harbour.Layout()
			for y := band; y < g[1]; y++ {
				for x := 0; x < g[0]; x++ {
					r, fg, bg := f.c.ResolveAt(x, y, term.Profile256)
					if r == ' ' || r == 0 {
						continue
					}
					if gap := math.Abs(harbourLuma(fg) - harbourLuma(bg)); gap < worst {
						worst, at = gap, fmt.Sprintf("%dx%d %.2f %q at (%d,%d)", g[0], g[1], hr, r, x, y)
					}
				}
			}
		}
	}
	t.Logf("smallest desk ink-to-ground gap %.1f at %s", worst, at)
	if worst < 40 {
		t.Errorf("desk writing at %.1f luma from its ground: %s", worst, at)
	}
}

// Under the floor (HarbourMinRows) and at tiny widths the scape still
// composes without a panic: a window can be dragged that small.
func TestTheHarbourSurvivesTinyWindows(t *testing.T) {
	for _, g := range [][2]int{{40, 12}, {30, 10}, {20, 8}, {12, 6}, {200, 15}} {
		f := harbourFrames(t, g[0], g[1])
		for _, pose := range []companion.State{companion.Resting, companion.NeedsYou, companion.Done} {
			st := harbourState(0.8, pose, 0.9, 0.95, 4)
			st.KittenExits = []float64{0.3, 0.7}
			f.compose(st, 9)
			f.c.Render(term.Profile256)
		}
	}
}

// Kittens come up over the sill: at their first frame nothing of them shows
// above the sill but (at most) the ears, and after the walk they sit.
func TestTheHarboursKittensComeUpOverTheSill(t *testing.T) {
	f := harbourFrames(t, 125, 28)
	count := func() int {
		catX, catY, band := f.harbour.Layout()
		n := 0
		for y := catY; y < band-1; y++ {
			for x := 2; x < catX; x++ {
				if r, fg, _ := f.c.ResolveAt(x, y, term.Profile256); r == 'o' && fg.G > fg.R+20 {
					n++ // a kitten's eye: green 'o'
				}
			}
		}
		return n
	}
	st := harbourState(0.0, companion.Working, 0.2, 0.5, 0)
	f.compose(st, 10)
	st.Kittens = 1
	f.compose(st, 10.01)
	first := count()
	f.compose(st, 13)
	sat := count()
	if first != 0 || sat != 2 {
		t.Fatalf("kitten eyes: %d on arrival (want 0), %d once sitting (want 2)", first, sat)
	}
}
