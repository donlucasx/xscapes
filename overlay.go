package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// overlayPage mocks the agent running INSIDE the scape instead of beside it,
// and the layout that follows from it: a taller window should spend its extra
// rows on beach, because the beach is where the agent's work is written.
//
// It is a mockup and not the thing. It composites a captured snapshot of a real
// Claude Code pane over a real scape frame by the rule the built version would
// use -- the agent's glyphs win, and everywhere the agent left a blank the
// scape shows through. A snapshot cannot be typed into, which is exactly the
// point: it answers the look question for an afternoon rather than for the
// price of writing a terminal emulator.
//
// The measurement that motivated it: 83.5% of a real 145x43 Claude Code pane is
// blank, and in a 66x35 scape the sky paints 1.7 glyphs a row against the sea's
// 34.9. Five sixths of the agent's window is free, and the region currently
// given 42% of every new row is the emptiest one on screen.
func overlayPage(seed int64, paneFile string) string {
	raw, err := os.ReadFile(paneFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asciiscapes:", err)
		os.Exit(1)
	}
	pane := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	paneW := 0
	for _, ln := range pane {
		if n := len([]rune(ln)); n > paneW {
			paneW = n
		}
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden;display:inline-block}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:26px 0 6px}
	.cap{font:11px ui-monospace,monospace;color:#8a8a99;margin:0 0 8px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; the agent inside the scape</h1>`)
	b.WriteString(`<p class="nt">A real Claude Code pane composited over a real scape frame: the ` +
		`agent's glyphs win, and every cell it left blank shows the sea. No split, no second pane. ` +
		`Below that, the same thing at three window heights with the layout Lucas asked for &mdash; ` +
		`the sky held to a constant band and every extra row spent on BEACH, so a taller window ` +
		`shows more of what the agent has been doing.</p>`)

	// A believable run of work, so the sand has something to say. Built
	// directly rather than through the reducer because the reducer caps the
	// tail at four lines, which is the constant this mockup exists to question.
	work := []string{
		"read   internal/auth/handler.go  142 lines",
		"search rate.Limiter  3 files",
		"edit   internal/auth/handler.go  +18 -2",
		"write  internal/auth/limiter.go  64 lines",
		"shell  go test ./...  4.1s",
		"read   internal/auth/limiter_test.go  88 lines",
		"edit   internal/auth/limiter.go  +6 -1",
		"shell  go vet ./...  0.9s",
		"search TestLimiter  2 files",
		"edit   internal/auth/limiter_test.go  +24 -0",
		"shell  go test ./internal/auth  1.2s",
		"read   README.md  61 lines",
	}
	tail := func(n int) []reduce.Line {
		if n > len(work) {
			n = len(work)
		}
		out := make([]reduce.Line, 0, n)
		start := len(work) - n
		for i, s := range work[start:] {
			out = append(out, reduce.Line{Text: s, Age: 1 - float64(i+1)/float64(n)})
		}
		return out
	}

	frame := func(w, h, skyRows, sandRows int, tod, lvl float64, lines []reduce.Line, withPane bool) string {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.MoonX = 0.28
		sh.SkyRows, sh.SandRows = skyRows, sandRows
		act := scape.Activity{Working: lvl > 0.3, Level: lvl, TimeOfDay: tod, ContextUsed: 0.35}
		for i := 0; i < 14; i++ {
			sh.Update(c, 2+float64(i)/20, act)
		}
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		pose := companion.Working
		if lvl < 0.3 {
			pose = companion.Resting
		}
		lay := compose(w, ccw, true)
		st := reduce.State{Pose: pose, Tail: lines}
		drawScene(c, sh, cat, lay, st, 3.1, seed, c.H-2-chh)
		if withPane {
			paintPane(c, pane)
		}
		return c.HTMLFragment(11)
	}

	b.WriteString(`<h2>His window, 145&times;43, agent composited in</h2>`)
	b.WriteString(`<p class="cap">Current layout. The sky takes 18 of 43 rows and paints almost nothing.</p>`)
	fmt.Fprintf(&b, `<div class="win">%s</div>`, frame(paneW, len(pane), 0, 0, 0.366, 0.85, tail(4), true))

	b.WriteString(`<h2>Same window, proposed layout</h2>`)
	b.WriteString(`<p class="cap">Sky held to 10 rows; the beach takes what is left and carries ` +
		`nine lines of work instead of four.</p>`)
	fmt.Fprintf(&b, `<div class="win">%s</div>`, frame(paneW, len(pane), 10, 16, 0.366, 0.85, tail(9), true))

	b.WriteString(`<h2>The scene alone, as the window grows</h2>`)
	b.WriteString(`<p class="cap">Sky constant at 10 rows, sea constant, every extra row becomes ` +
		`beach: 4 lines of history at 24 rows, 9 at 43, 16 at 60.</p>`)
	for _, s := range []struct {
		h, sand, lines int
		note           string
	}{{24, 7, 4, "24 rows"}, {43, 16, 9, "43 rows"}, {60, 26, 16, "60 rows"}} {
		fmt.Fprintf(&b, `<div class="win">%s</div> `,
			frame(84, s.h, 10, s.sand, 0.366, 0.85, tail(s.lines), false))
	}

	return canvas.HTMLPage("asciiscapes — agent inside the scape", b.String())
}

// paintPane draws the captured agent screen into the near layer. A space is
// left alone so the scape shows through, which is the whole idea; everything
// else is painted opaque, because text over moving water is unreadable unless
// the text owns its cell outright.
func paintPane(c *canvas.Canvas, pane []string) {
	fg := term.RGB{R: 236, G: 238, B: 244}
	for y, ln := range pane {
		if y >= c.H {
			break
		}
		for x, r := range []rune(ln) {
			if x >= c.W || r == ' ' {
				continue
			}
			c.Near().Plot(x, y, r, fg, 1)
		}
	}
}
