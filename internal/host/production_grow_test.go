package host

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// A grow, through a real Host wired the way production wires it for each
// terminal, with a child that redraws its prompt the way Hermes's
// prompt_toolkit does: relatively from the cursor, on SIGWINCH, never
// touching the rows above.
//
// The test whose absence shipped f166952 inert in Ghostty (2026-11-08, Kimi
// audit axis 6): the grow tests fed resizeSequence the history by hand, and
// History -- which was then the only thing that kept a model -- is off by
// default outside Terminal.app (inside.go). So in Ghostty nothing came back
// and the prompt stayed stranded. Here each leg runs with that terminal's
// production defaults and the terminal model under that terminal's rules,
// and the bytes on the wire are checked as well as the screen.
func TestAGrowBringsThePromptDownWithEachTerminalsDefaults(t *testing.T) {
	const child = `i=1; while [ $i -le 40 ]; do printf 'HIST %02d\r\n' $i; i=$((i+1)); done
printf '+---- old box ----+\r\n| > hi            |\r\n+-----------------+\r\033[1A\033[7C'
trap 'printf "\r\033[1A\033[J+==== new box ====+\r\n| > hi again      |\r\n+=================+\r\033[1A\033[7C"' WINCH
n=0; while [ $n -lt 26 ]; do sleep 0.05; n=$((n+1)); done`

	for _, leg := range []struct {
		name    string
		rules   Rules
		history bool // inside.go: on in Terminal.app only
	}{
		{"Terminal.app", AppleTerminalRules, true},
		{"Ghostty", XTermRules, false},
	} {
		t.Run(leg.name, func(t *testing.T) {
			const w, h1, h2 = 100, 30, 40
			sc, raw := runProduction(t, leg.rules, leg.history, child, w, h1, h2)
			assertTranscriptRunsUnbroken(t, sc, h2)
			band, _ := Band(h2)
			grow := band - func() int { b, _ := Band(h1); return b }()

			var box []int
			for y := 0; y < band; y++ {
				if strings.Contains(sc.rowAt(y), "new box") {
					box = append(box, y+1)
				}
			}
			if len(box) != 1 || box[0] != band-2 {
				for y := 0; y < band; y++ {
					t.Logf("%2d|%s", y+1, strings.TrimRight(sc.rowAt(y), " "))
				}
				t.Errorf("new box at rows %v, want row %d: the prompt is not on the band's bottom (band 1..%d)", box, band-2, band)
			}
			if r := sc.rowAt(0); !strings.Contains(r, "HIST") {
				t.Errorf("band row 1 = %q, want a history row brought back", strings.TrimSpace(r))
			}
			// On the wire: Ghostty kept the text at the top, so the host
			// scrolls it down by the band's grow; Terminal.app pushed it down
			// by the window's grow, so the host takes back the difference.
			want := fmt.Sprintf("\x1b[%dT", grow)
			if leg.rules.GrowPushesDown {
				want = fmt.Sprintf("\x1b[%dS", (h2-h1)-grow)
			}
			if !strings.Contains(raw, want) {
				t.Errorf("no %q on the wire", want)
			}
		})
	}
}

// A shrink, then a grow back. The shrink pushes the band's top rows off the
// screen; they are the newest history when the grow comes, and nothing may go
// missing or come back out of order.
//
// His Ghostty run of 2026-11-08 16:11 (trace 20261108-161047): 34 -> 22 rows
// lost facts 29-36 for good (a resize, unlike a scroll, kept nothing) and left
// two blank rows at the band's top (the shrink's SD); the grows then brought
// back 11-28 above that gap. Replayed through the model, exactly his screen.
func TestAShrinkThenAGrowLosesNoRow(t *testing.T) {
	const child = `i=1; while [ $i -le 40 ]; do printf 'HIST %02d\r\n' $i; i=$((i+1)); done
printf '+---- old box ----+\r\n| > hi            |\r\n+-----------------+\r\033[1A\033[7C'
trap 'printf "\r\033[1A\033[J+==== new box ====+\r\n| > hi again      |\r\n+=================+\r\033[1A\033[7C"' WINCH
n=0; while [ $n -lt 40 ]; do sleep 0.05; n=$((n+1)); done`
	for _, leg := range []struct {
		name    string
		rules   Rules
		history bool
	}{
		{"Terminal.app", AppleTerminalRules, true},
		{"Ghostty", XTermRules, false},
	} {
		t.Run(leg.name, func(t *testing.T) {
			const w = 100
			sc, _ := runProductionSteps(t, leg.rules, leg.history, child, w, 40, []int{28, 40})
			assertTranscriptRunsUnbroken(t, sc, 40)
		})
	}
}

// assertTranscriptRunsUnbroken: every band row above the new box is a HIST
// row, numbered consecutively, ending at HIST 40 right above the box.
func assertTranscriptRunsUnbroken(t *testing.T, sc *screen, rows int) {
	t.Helper()
	band, _ := Band(rows)
	box := -1
	for y := 0; y < band; y++ {
		if strings.Contains(sc.rowAt(y), "new box") {
			box = y
		}
	}
	if box < 0 {
		t.Fatalf("no new box in the band")
	}
	want := 40
	for y := box - 1; y >= 0; y-- {
		if got := strings.TrimSpace(sc.rowAt(y)); got != fmt.Sprintf("HIST %02d", want) {
			for y := 0; y < band; y++ {
				t.Logf("%2d|%s", y+1, strings.TrimRight(sc.rowAt(y), " "))
			}
			t.Fatalf("band row %d = %q, want HIST %02d: a row is missing or out of order", y+1, got, want)
		}
		want--
	}
}

// runProduction runs a Host with the given rules and History, a terminal
// model that resizes under the same rules, and one grow from h1 to h2 rows.
// It returns the model as it stood just before the child exited, and the raw
// bytes written.
func runProduction(t *testing.T, rules Rules, history bool, child string, w, h1, h2 int) (*screen, string) {
	t.Helper()
	return runProductionSteps(t, rules, history, child, w, h1, []int{h2})
}

// runProductionSteps is runProduction through several heights in turn.
func runProductionSteps(t *testing.T, rules Rules, history bool, child string, w, h1 int, steps []int) (*screen, string) {
	t.Helper()
	var sizes [][2]int
	for _, h := range steps {
		sizes = append(sizes, [2]int{w, h})
	}
	return runProductionSizes(t, rules, history, child, w, h1, sizes)
}

// runProductionSizes is runProduction through several window sizes, width
// and height, in turn.
func runProductionSizes(t *testing.T, rules Rules, history bool, child string, w1, h1 int, sizes [][2]int) (*screen, string) {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	tty, err := openPTY()
	if err != nil {
		t.Skipf("no pty available: %v", err)
	}
	defer tty.Close()

	sc := newScreen(w1, h1)
	sc.restoreAbsolute = !rules.RestoreHomes
	var mu sync.Mutex
	var raw strings.Builder
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		answered := false
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				mu.Lock()
				raw.Write(buf[:n])
				sc.feed(string(buf[:n]))
				if history && !answered && strings.Contains(raw.String(), "\x1b[6n") {
					answered = true
					tty.master.Write([]byte("\x1b[1;1R"))
				}
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	var sz struct {
		sync.Mutex
		w, h int
	}
	sz.w, sz.h = w1, h1
	h := &Host{
		Cmd:       exec.Command("sh", "-c", child),
		FPS:       30,
		AltScreen: true,
		Rules:     rules,
		History:   history,
		In:        tty.slave,
		Out:       pw,
		Size: func() (int, int) {
			sz.Lock()
			defer sz.Unlock()
			return sz.w, sz.h
		},
		Paint: func(cols, rows int) []string {
			out := make([]string, rows)
			for i := range out {
				out[i] = scapeRow(cols, i, 'S')
			}
			return out
		},
	}
	var snap *screen
	go func() {
		time.Sleep(500 * time.Millisecond) // the child has printed and parked
		for _, size := range sizes {
			mu.Lock()
			sc.resizeAltFor(rules, size[0], size[1]) // the terminal moves first
			mu.Unlock()
			sz.Lock()
			sz.w, sz.h = size[0], size[1]
			sz.Unlock()
			time.Sleep(400 * time.Millisecond) // the host's tick, SIGWINCH, the redraw
		}
		mu.Lock()
		snap = sc.clone()
		mu.Unlock()
	}()
	if err := h.Run(); err != nil && err != io.EOF {
		t.Fatalf("host: %v", err)
	}
	pw.Close()
	<-done
	if snap == nil {
		t.Fatal("no snapshot before the host exited")
	}
	mu.Lock()
	defer mu.Unlock()
	return snap, raw.String()
}

// Narrowed, then widened again: the text the narrowing cut comes back.
//
// His Ghostty screenshot of 2026-11-08 16:35: "after narrowing the window and
// widening it again its broken" -- forty one-line facts cut at the narrow
// width, in a window twice as wide. Ghostty's alternate screen deletes the
// cells past a narrower width (PageList.zig resizeWithoutReflow) and nothing
// repaints a line the agent does not rewrite. Terminal.app keeps the cells
// itself (notes/width-audit.md), so only Ghostty needs the host to put them
// back. A line the agent rewrites while narrow must NOT get its old tail back.
func TestAWidenBringsBackWhatANarrowingCut(t *testing.T) {
	const child = `i=1; while [ $i -le 12 ]; do printf 'LONG %02d the quick brown fox jumps over the lazy dog END%02d\r\n' $i $i; i=$((i+1)); done
printf 'prompt> a long status line that runs well past thirty columns STALE'
trap 'printf "\rPROMPT>"' WINCH
n=0; while [ $n -lt 40 ]; do sleep 0.05; n=$((n+1)); done`
	// Widened by a drag, in steps, as his window was (56 -> 72 -> 93 -> ...
	// in trace 20261108-180840): the first step's repaint used to count as
	// the agent touching the row, and dropped the rest of every tail.
	sc, _ := runProductionSizes(t, XTermRules, false, child, 100, 40, [][2]int{{30, 40}, {45, 40}, {70, 40}, {100, 40}})
	band, _ := Band(40)
	var full, cut []string
	for y := 0; y < band; y++ {
		r := strings.TrimRight(sc.rowAt(y), " ")
		if strings.HasPrefix(r, "LONG") {
			if strings.HasSuffix(r, fmt.Sprintf("END%s", r[5:7])) {
				full = append(full, r[:7])
			} else {
				cut = append(cut, r)
			}
		}
	}
	if len(cut) > 0 || len(full) != 12 {
		for y := 0; y < band; y++ {
			t.Logf("%2d|%s", y+1, strings.TrimRight(sc.rowAt(y), " "))
		}
		t.Errorf("%d lines whole, %d still cut after the widen: %q", len(full), len(cut), cut)
	}
	// The prompt line was written over while narrow, without an erase: what
	// it had past the narrow width is no longer its content, and a widen
	// must not bring it back.
	for y := 0; y < band; y++ {
		if r := sc.rowAt(y); strings.Contains(r, "STALE") {
			t.Errorf("a rewritten line got its old tail back: %q", strings.TrimRight(r, " "))
		}
	}
}
