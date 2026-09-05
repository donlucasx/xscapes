package host

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestReplayTrace reconstructs a real session's screen from a trace file.
//
// Not a test of anything: an instrument, skipped unless pointed at a trace by
// XSCAPES_TRACE. A screenshot shows what a screen LOOKED like; it cannot say
// whether the terminal moved the content or the host erased it. Replaying the
// exact bytes the host sent, through the same model the resize tests use,
// answers that.
//
//	XSCAPES_TRACE=/tmp/t.bin xscapes claude     # in a real window, then resize
//	XSCAPES_TRACE=/tmp/t.bin go test ./internal/host -run TestReplayTrace -v
//
// TRACE_SIZE=WxH gives the window size to model; TRACE_ALT=0 replays on the
// main screen. A size change is applied as the ALTERNATE screen does it
// (anchored bottom, cursor unmoved -- notes/contentprobe and shrinkprobe), or
// as "keep what fits" for the main screen.
func TestReplayTrace(t *testing.T) {
	path := os.Getenv("XSCAPES_TRACE")
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace file to replay it")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The sidecar says which size was in force from which byte offset, so the
	// replay resizes exactly where the real terminal did. Feeding the whole
	// stream into one fixed-size screen would misreconstruct everything
	// written before the resize, which is the part under investigation.
	marks := readTraceMarks(t, path)
	w, h := 120, 59
	if s := os.Getenv("TRACE_SIZE"); s != "" {
		if _, err := fmt.Sscanf(s, "%dx%d", &w, &h); err != nil {
			t.Fatalf("TRACE_SIZE=%q: want WxH", s)
		}
	}
	if len(marks) > 0 {
		w, h = marks[0].cols, marks[0].rows
	}
	sc := newScreen(w, h)
	// TRACE_RETAIN=1 replays with Terminal.app's width rule (retain and clip,
	// measured 2026-09-05) instead of cut-and-pad.
	sc.retainWidth = os.Getenv("TRACE_RETAIN") == "1"
	fed := 0
	for i, m := range marks {
		if i > 0 {
			sc.feed(string(b[fed:min(m.off, len(b))]))
			fed = min(m.off, len(b))
			if os.Getenv("TRACE_ALT") == "0" {
				sc.resize(m.cols, m.rows)
			} else {
				sc.resizeAlt(m.cols, m.rows)
			}
			t.Logf("-- resize at byte %d -> %dx%d (band 1..%d)", m.off, m.cols, m.rows, m.agent)
		}
	}
	// Stop before the host hands the terminal back: Close emits ESC[?1049l,
	// which swaps away the alternate buffer -- the one everything under
	// investigation was drawn on. Replaying past it shows a blank screen and
	// looks exactly like "everything was erased".
	tail := string(b[fed:])
	if i := strings.LastIndex(tail, "\x1b[?1049l"); i >= 0 {
		tail = tail[:i]
		t.Logf("-- stopping before the host's exit (ESC[?1049l)")
	}
	sc.feed(tail)

	t.Logf("replayed %d bytes at %dx%d; scroll region rows %d..%d, cursor r%d c%d, alt=%v",
		len(b), w, h, sc.top+1, sc.bot+1, sc.y+1, sc.x+1, sc.alt)
	// TRACE_DUMP=dir writes the alternate screen and the main buffer as text,
	// one row per line, so a real session's readback (`history of tab`) can be
	// diffed against what the model believes -- the fidelity test for the
	// mirror, on the agent's real bytes.
	if dir := os.Getenv("TRACE_DUMP"); dir != "" {
		var alt, main strings.Builder
		for y := 0; y < sc.h; y++ {
			alt.WriteString(sc.rowAt(y) + "\n")
			main.WriteString(sc.otherRowAt(y) + "\n")
		}
		os.WriteFile(dir+"/model-alt.txt", []byte(alt.String()), 0o644)
		os.WriteFile(dir+"/model-main.txt", []byte(main.String()), 0o644)
	}
	for y := 0; y < sc.h; y++ {
		row := sc.rowAt(y)
		bg, uniform := sc.bgRunAt(y)
		mark := "  "
		switch {
		case uniform && bg != defaultBG:
			mark = "##" // erased under a colour: a solid band on the real screen
		case row == "":
			mark = ".." // genuinely empty
		}
		if len(row) > 100 {
			row = row[:100] + "…"
		}
		t.Logf("%s row %2d | %s", mark, y+1, row)
	}
}

// traceMark is one size change, at the byte offset it took effect.
type traceMark struct{ off, cols, rows, agent int }

func readTraceMarks(t *testing.T, path string) []traceMark {
	t.Helper()
	var marks []traceMark
	lb, err := os.ReadFile(path + ".log")
	if err != nil {
		return nil
	}
	for _, ln := range strings.Split(strings.TrimSpace(string(lb)), "\n") {
		var m traceMark
		if _, err := fmt.Sscanf(ln, "%d %d %d %d", &m.off, &m.cols, &m.rows, &m.agent); err == nil {
			marks = append(marks, m)
		}
	}
	return marks
}

// TestTraceRightEdge asks one question of a real trace: was every scape row
// painted all the way to the last column the host believed in?
//
// He photographed a strip of stale scape down the far right. That is what a row
// painted narrower than the terminal leaves behind -- the columns past the end
// are never written, so they keep whatever the previous frame, or the previous
// SIZE, put there. This separates the two causes: rows short of the host's own
// `cols` mean the RENDERER is wrong; rows full to `cols` mean the host's idea
// of `cols` was behind the window.
func TestTraceRightEdge(t *testing.T) {
	path := os.Getenv("XSCAPES_TRACE")
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace file")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	if len(marks) == 0 {
		t.Skip("no size sidecar next to the trace")
	}
	sc := newScreen(marks[0].cols, marks[0].rows)
	fed := 0
	for i, m := range marks {
		if i == 0 {
			continue
		}
		sc.feed(string(b[fed:min(m.off, len(b))]))
		fed = min(m.off, len(b))
		sc.resize(m.cols, m.rows)
	}
	// Stop before the host hands the terminal back: Close emits ESC[?1049l,
	// which swaps away the very buffer under investigation.
	tail := string(b[fed:])
	if i := strings.LastIndex(tail, "\x1b[?1049l"); i >= 0 {
		tail = tail[:i]
	}
	sc.feed(tail)

	last := marks[len(marks)-1]
	t.Logf("final geometry the HOST believed: %dx%d, band 1..%d", last.cols, last.rows, last.agent)
	short := 0
	for y := last.agent; y < sc.h && y < len(sc.cells); y++ {
		row := sc.cells[y]
		end := -1
		for x := len(row) - 1; x >= 0; x-- {
			if row[x].r != ' ' || row[x].bg != defaultBG {
				end = x
				break
			}
		}
		if end != sc.w-1 {
			short++
			t.Errorf("scape row %d painted only through column %d of %d", y+1, end+1, sc.w)
		}
	}
	if short == 0 {
		t.Logf("every scape row is painted through column %d of %d -- the renderer is not the cause;"+
			" any strip beyond column %d means the window was wider than the host knew",
			sc.w, sc.w, sc.w)
	}
}
