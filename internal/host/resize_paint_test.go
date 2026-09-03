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

// scapeRow is a recognisable full-width row: the row index repeated, so a stale
// row from a previous frame or a previous SIZE is obvious in the model.
func scapeRow(cols, i int, tag rune) string {
	b := make([]rune, cols)
	for x := range b {
		b[x] = tag
	}
	label := fmt.Sprintf("%d", i)
	copy(b, []rune(label))
	return string(b)
}

// runHosted drives a real Host through a size change and returns the screen as
// the terminal would have it.
func runHosted(t *testing.T, w1, h1, w2, h2 int, scrolling bool) *screen {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	sc := newScreen(w1, h1)
	var raw strings.Builder
	nbytes := 0
	var mu sync.Mutex
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				mu.Lock()
				nbytes += n
				raw.Write(buf[:n])
				sc.feed(string(buf[:n]))
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
	tag := 'A'

	// In has to be a real tty: the host puts it in raw mode, and a pipe
	// cannot do that. A pty's slave is a tty, which is exactly what the host
	// will be handed in production.
	tty, err := openPTY()
	if err != nil {
		t.Skipf("no pty available: %v", err)
	}
	defer tty.Close()

	h := &Host{
		Cmd:       exec.Command("sh", "-c", "sleep 0.8"),
		FPS:       60,
		AltScreen: false,
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
				out[i] = scapeRow(cols, i, tag)
			}
			return out
		},
	}
	// Snapshot BEFORE the host exits. Its close path hands the terminal back
	// and blanks the band, so a screen inspected after Run() is legitimately
	// empty -- which the first version of this test read as "the host paints
	// nothing" and blamed the host for.
	var snap *screen
	go func() {
		time.Sleep(600 * time.Millisecond)
		mu.Lock()
		snap = sc.clone()
		mu.Unlock()
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		// The model has to learn about the resize too, the way a real terminal
		// would: it keeps what fits and exposes blank rows where it grew.
		mu.Lock()
		if scrolling {
			sc.resizeScrolling(w2, h2)
		} else {
			sc.resize(w2, h2)
		}
		mu.Unlock()
		tag = 'B'
		sz.Lock()
		sz.w, sz.h = w2, h2
		sz.Unlock()
	}()

	if err := h.Run(); err != nil && err != io.EOF {
		t.Fatalf("host: %v", err)
	}
	pw.Close()
	<-done
	if snap == nil {
		t.Fatal("no snapshot taken before the host exited")
	}
	sc = snap
	if os.Getenv("DBG") != "" {
		fmt.Printf("--- captured %d bytes; screen %dx%d ---\n", nbytes, sc.w, sc.h)
		fmt.Printf("first 500 bytes: %q\n", trunc500(raw.String()))
		for y := 0; y < sc.h; y++ {
			if r := sc.rowAt(y); r != "" {
				fmt.Printf("row %2d: %q\n", y, trunc(r))
			}
		}
	}
	return sc
}

// After a resize, every row of the scape's band must be the NEW frame. A row
// left over from before the resize is the defect he photographed: a fragment of
// the old scape sitting above the band, and a strip of it down the right edge.
func TestEveryBandRowIsRepaintedAfterAResize(t *testing.T) {
	for _, tc := range []struct {
		name           string
		w1, h1, w2, h2 int
		scrolling      bool
	}{
		{"wider", 100, 40, 120, 40, false},
		{"narrower", 120, 40, 100, 40, false},
		{"taller", 100, 40, 100, 50, false},
		{"shorter", 100, 50, 100, 40, false},
		{"wider and shorter", 100, 50, 128, 44, false},
		{"his window", 120, 46, 128, 51, false},
		// The ones that matter: a terminal that keeps the BOTTOM when it
		// shrinks, which is what Terminal.app does.
		{"shorter, terminal scrolls", 100, 50, 100, 40, true},
		{"his window, shrunk", 128, 51, 120, 44, true},
		{"one row shorter", 128, 51, 128, 50, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := runHosted(t, tc.w1, tc.h1, tc.w2, tc.h2, tc.scrolling)
			agent, scape := BandWith(tc.h2, 0)
			if scape <= 0 {
				t.Skip("no scape at this size")
			}
			for i := 0; i < scape; i++ {
				row := agent + i // 0-based screen row
				got := sc.rowAt(row)
				want := strings.TrimRight(scapeRow(tc.w2, i, 'B'), " ")
				if got != want {
					t.Errorf("screen row %d (scape row %d of %d):\n  got  %q\n  want %q",
						row, i, scape, trunc(got), trunc(want))
				}
			}
			// And nothing of the scape may be left ABOVE the band.
			for row := 0; row < agent; row++ {
				if r := sc.rowAt(row); strings.ContainsRune(r, 'A') || strings.ContainsRune(r, 'B') {
					t.Errorf("screen row %d is inside the agent's band and still holds scape: %q",
						row, trunc(r))
				}
			}
		})
	}
}

func trunc(s string) string {
	if len(s) > 60 {
		return s[:60] + "..."
	}
	return s
}

func trunc500(s string) string {
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
