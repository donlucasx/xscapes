package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// vistaState is one composed situation for the vista to draw.
type vistaState struct {
	name string
	st   reduce.State
}

func vistaStates(tod float64) []vistaState {
	tail := []reduce.Line{
		{Text: "read  internal/scape/shore.go", Age: 0.7},
		{Text: "edit  internal/scenes/owl.go  +40 -2", Age: 0.4},
		{Text: "shell go test ./...  exit 1", Age: 0.2, Bad: true},
		{Text: "shell go build ./...  1.2s", Age: 0.0},
	}
	act := func(level float64, working bool) scape.Activity {
		return scape.Activity{Working: working, Level: level, ContextUsed: 0.55, TimeOfDay: tod, TodoDone: 7, TodoTotal: 12}
	}
	return []vistaState{
		{"resting", reduce.State{Act: act(0, false), Pose: companion.Resting}},
		{"working, 3 subagents", reduce.State{Act: act(0.65, true), Pose: companion.Working, Kittens: 3, Tail: tail}},
		{"needs you", reduce.State{Act: act(0.1, false), Pose: companion.NeedsYou, Bubble: "allow Bash?", BubbleAsk: true, Tail: tail}},
		{"done", reduce.State{Act: act(0.05, false), Pose: companion.Done, Bubble: "done", Tail: tail}},
		{"worried, flat out", reduce.State{Act: act(1.0, true), Pose: companion.Worried, Kittens: 1, Tail: tail}},
	}
}

// renderVista draws one vista frame through the live path.
func renderVista(t *testing.T, w, h int, st reduce.State, tm float64) *frames {
	t.Helper()
	f := newFrames(w, h, 7, false, true, 0, 0)
	if f.vista == nil {
		t.Fatalf("XSCAPES_SCAPE=vista did not select the vista")
	}
	f.vista.OwlX = f.lay.CatX
	f.vista.Update(f.c, tm, st.Act)
	drawVista(f.c, f.vista, f.lay, st, tm)
	return f
}

// TestVistaShots writes the vista's live composition as frame pages when
// XSCAPES_VISTASHOTS names a directory, one page per geometry, a row per
// state at three hours -- to be screenshotted with headless Chrome and LOOKED
// AT, which is how every vista defect so far was found (s32).
func TestVistaShots(t *testing.T) {
	dir := os.Getenv("XSCAPES_VISTASHOTS")
	if dir == "" {
		t.Skip("set XSCAPES_VISTASHOTS=<dir> to write the pages")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	const style = `<meta charset="utf-8"><style>html,body{margin:0;padding:12px;background:#000;color:#ccc;font:13px Menlo,monospace}pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0}.row{display:flex;gap:10px;align-items:flex-start;margin-bottom:10px}.lab{width:150px;padding-top:50px;font-weight:bold}h2{font:bold 14px Menlo,monospace;color:#eee;margin:18px 0 8px;border-top:1px solid #333;padding-top:10px}</style>`
	for _, g := range [][2]int{{80, 24}, {125, 28}, {124, 30}, {60, 20}, {40, 12}} {
		var b strings.Builder
		b.WriteString(style)
		for _, hr := range []float64{0, 0.78, 0.5} {
			fmt.Fprintf(&b, "<h2>%dx%d at %.2f</h2>", g[0], g[1], hr)
			for _, vs := range vistaStates(hr) {
				f := renderVista(t, g[0], g[1], vs.st, 3.0)
				fmt.Fprintf(&b, `<div class="row"><div class="lab">%s</div>%s</div>`, vs.name, f.c.HTMLFragmentAs(11, term.Profile256))
			}
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("vista-%dx%d.html", g[0], g[1])), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestTheVistaComposesEveryStateAtEveryWidth is the composition's own
// guarantee: at each geometry and state the owl's ink is inside the frame,
// the writing lands on the band and nowhere above it, the litter is counted
// and never in the fire, and a balloon is drawn when there is one.
func TestTheVistaComposesEveryStateAtEveryWidth(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "vista")
	for _, g := range [][2]int{{80, 24}, {125, 28}, {124, 30}, {143, 27}, {60, 20}, {40, 12}} {
		for _, hr := range []float64{0, 0.5, 0.78} {
			for _, vs := range vistaStates(hr) {
				f := renderVista(t, g[0], g[1], vs.st, 3.0)
				owlX, owlY, _, bandTop := f.vista.Layout()
				// The owl is on screen: against the same frame with nothing
				// composed over it, its box differs in most of its cells.
				bare := newFrames(g[0], g[1], 7, false, true, 0, 0)
				bare.vista.OwlX = bare.lay.CatX
				bare.vista.Update(bare.c, 3.0, vs.st.Act)
				owlCells := 0
				for y := owlY; y < owlY+7 && y < g[1]; y++ {
					for x := owlX; x < owlX+12 && x < g[0]; x++ {
						r1, f1, b1 := f.c.ResolveAt(x, y, term.Profile256)
						r2, f2, b2 := bare.c.ResolveAt(x, y, term.Profile256)
						if r1 != r2 || f1 != f2 || b1 != b2 {
							owlCells++
						}
					}
				}
				if owlCells < 50 {
					t.Errorf("%dx%d %.2f %s: only %d cells of the owl's box changed by drawing it at (%d,%d)", g[0], g[1], hr, vs.name, owlCells, owlX, owlY)
				}
				// The writing: its letters are on rows >= bandTop only.
				for y := 0; y < g[1]; y++ {
					for x := 0; x < g[0]; x++ {
						r, _, _ := f.c.ResolveAt(x, y, term.Profile256)
						if y < bandTop && (r == 'o' || r == 'w') && x > 30 && x < owlX-8 && y > bandTop-3 {
							// letters of the sample above the band would be a leak; the
							// scrub uses '"' ',' '_' and the fire '^' '*' ')'.
							t.Errorf("%dx%d %s: a letter %q above the band at (%d,%d)", g[0], g[1], vs.name, r, x, y)
						}
					}
				}
				// The litter stands above the band, never on the writing.
				if vs.st.Kittens > 0 {
					_, _, grassY, _ := f.vista.Layout()
					if grassY+4 > bandTop {
						t.Errorf("%dx%d %s: owlets on rows %d..%d overlap the band at %d", g[0], g[1], vs.name, grassY, grassY+3, bandTop)
					}
				}
				if len(vs.st.Tail) > 0 && g[0] >= 60 {
					found := false
					for y := bandTop; y < g[1]; y++ {
						for x := 0; x < g[0]; x++ {
							if r, _, _ := f.c.ResolveAt(x, y, term.Profile256); r == 'g' {
								found = true
							}
						}
					}
					if !found {
						t.Errorf("%dx%d %.2f %s: the tail was not written on the band", g[0], g[1], hr, vs.name)
					}
				}
				if vs.st.Bubble != "" && g[0] >= 60 {
					// Every letter of the balloon's text is on screen above the
					// owl, in order on one row: the treeline's quarter-cells ate
					// three of "allow Bash?" in the first build.
					found := false
					for y := 0; y < owlY; y++ {
						var row []rune
						for x := 0; x < g[0]; x++ {
							r, _, _ := f.c.ResolveAt(x, y, term.Profile256)
							row = append(row, r)
						}
						if strings.Contains(string(row), vs.st.Bubble) {
							found = true
						}
					}
					if !found {
						t.Errorf("%dx%d %.2f %s: the balloon text %q is not intact above the owl", g[0], g[1], hr, vs.name, vs.st.Bubble)
					}
				}
			}
		}
	}
}

// TestTheScapeSwitchReachesARunningScape: the preference is polled, so a
// change lands without a restart, in both directions.
func TestTheScapeSwitchReachesARunningScape(t *testing.T) {
	t.Setenv("XSCAPES_SCAPE", "shore")
	f := newFrames(80, 24, 7, false, true, 0, 0)
	if f.vista != nil {
		t.Fatalf("shore was asked for; the vista was built")
	}
	t.Setenv("XSCAPES_SCAPE", "vista")
	now := time.Now().Add(2 * companionPoll)
	f.refreshScape(now)
	if f.vista == nil {
		t.Fatalf("the switch to the vista did not reach the running frames")
	}
	t.Setenv("XSCAPES_SCAPE", "shore")
	f.refreshScape(now.Add(2 * companionPoll))
	if f.vista != nil {
		t.Fatalf("the switch back to the shore did not reach the running frames")
	}
}
