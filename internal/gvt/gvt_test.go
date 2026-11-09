//go:build ghosttyvt

package gvt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	cup = regexp.MustCompile(`\x1b\[(\d+);1H`)
	csi = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)
)

// rows applies a Render to a screen of plain text rows.
func apply(scr []string, b []byte) {
	s := string(b)
	locs := cup.FindAllStringSubmatchIndex(s, -1)
	for i, l := range locs {
		y, _ := strconv.Atoi(s[l[2]:l[3]])
		end := len(s)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		if y-1 < len(scr) {
			scr[y-1] = csi.ReplaceAllString(s[l[1]:end], "")
		}
	}
}

func TestANarrowingReflowsAndAWideningJoinsBack(t *testing.T) {
	v, err := New(40, 6, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	line := "Varuna is a fast-rotating Kuiper Belt object"
	v.Write([]byte("> prompt\r\n" + line + "\r\n$ "))
	scr := make([]string, 6)
	apply(scr, v.Render(true, 0))
	if !strings.Contains(scr[1], "Varuna is a fast-rotating Kuiper Belt o") {
		t.Fatalf("40 cols row 2 = %q", scr[1])
	}

	v.Resize(20, 6)
	scr = make([]string, 6)
	apply(scr, v.Render(true, 0))
	joined := strings.Join(scr, "|")
	for _, want := range []string{"Varuna is a fast-rot", "ating Kuiper Belt ob", "ject"} {
		if !strings.Contains(joined, want) {
			t.Errorf("20 cols: no row %q in %q", want, joined)
		}
	}

	v.Resize(60, 6)
	scr = make([]string, 6)
	apply(scr, v.Render(true, 0))
	joined = strings.Join(scr, "|")
	if !strings.Contains(joined, line) {
		t.Errorf("60 cols: the line did not join back: %q", joined)
	}
	c := v.Cursor()
	if got := strings.TrimRight(scr[c.Y], " "); !strings.HasPrefix(got, "$") {
		t.Errorf("cursor row %d = %q, want the prompt", c.Y, got)
	}
}

func TestACursorReportIsAnsweredByTheEngine(t *testing.T) {
	v, _ := New(30, 5, 0)
	defer v.Close()
	r := v.Write([]byte("\x1b[3;7H\x1b[6n"))
	if string(r) != "\x1b[3;7R" {
		t.Fatalf("reply %q", r)
	}
}

func TestOnlyChangedRowsAreRedrawn(t *testing.T) {
	v, _ := New(30, 5, 0)
	defer v.Close()
	v.Write([]byte("a\r\nb\r\nc"))
	v.Render(true, 0)
	v.Write([]byte("\x1b[2;1Hz"))
	got := cup.FindAllStringSubmatch(string(v.Render(false, 0)), -1)
	// Row 3 may come too: the cursor left it. Row 1 must not.
	rows := map[string]bool{}
	for _, g := range got {
		rows[g[1]] = true
	}
	if !rows["2"] || rows["1"] {
		t.Fatalf("redrawn rows %v, want row 2 and not row 1", got)
	}
}

func TestModesAreVisible(t *testing.T) {
	v, _ := New(30, 5, 0)
	defer v.Close()
	v.Write([]byte("\x1b[?2004h\x1b[?1h"))
	if !v.Mode(2004, false) || !v.Mode(1, false) || v.Mode(1000, false) {
		t.Fatal("modes not reported")
	}
}

func TestTranscriptJoinsWraps(t *testing.T) {
	v, _ := New(10, 3, 1<<20)
	defer v.Close()
	v.Write([]byte("abcdefghijklmnop\r\nline2\r\nline3\r\nline4"))
	tr := string(v.Transcript(false))
	if !strings.Contains(tr, "abcdefghijklmnop") || !strings.Contains(tr, "line4") {
		t.Fatalf("transcript %q", tr)
	}
}

// A closed terminal answers nothing and does not crash: the host's reader
// can write once more after the host has closed it.
func TestAClosedTerminalIsInert(t *testing.T) {
	v, _ := New(10, 3, 0)
	v.Close()
	v.Write([]byte("late bytes"))
	v.Resize(20, 5)
	_ = v.Render(true, 0)
	_ = v.Cursor()
	_ = v.Mode(1, false)
	_ = v.KittyFlags()
	_ = v.Transcript(true)
	v.Close()
}

// Two terminals render at once without bleeding into each other, and a row
// wider than 1024 columns is drawn whole (Kimi's probes, 2026-11-09: a
// shared static row buffer, capped at 1024).
func TestTerminalsDoNotShareARowAndWideRowsDrawWhole(t *testing.T) {
	a, _ := New(1400, 3, 0)
	b, _ := New(1400, 3, 0)
	defer a.Close()
	defer b.Close()
	a.Write([]byte(strings.Repeat("A", 1400)))
	b.Write([]byte(strings.Repeat("B", 1400)))
	done := make(chan []byte)
	for _, v := range []*Term{a, b} {
		go func(v *Term) {
			var last []byte
			for i := 0; i < 50; i++ {
				last = v.Render(true, 0)
			}
			done <- last
		}(v)
	}
	for i := 0; i < 2; i++ {
		out := csi.ReplaceAllString(string(<-done), "")
		if strings.Contains(out, "A") && strings.Contains(out, "B") {
			t.Fatal("one terminal's frame holds the other's cells")
		}
		if n := strings.Count(out, "A") + strings.Count(out, "B"); n != 1400 {
			t.Fatalf("row drawn with %d cells, want 1400", n)
		}
	}
}

// A grapheme cluster longer than the cell buffer still draws its base
// character, not a hole.
func TestALongClusterDrawsItsBase(t *testing.T) {
	v, _ := New(20, 2, 0)
	defer v.Close()
	v.Write([]byte("x" + strings.Repeat("́", 40) + "y"))
	out := csi.ReplaceAllString(string(v.Render(true, 0)), "")
	if !strings.HasPrefix(strings.TrimLeft(out, " "), "x") || !strings.Contains(out, "y") {
		t.Fatalf("rendered %q", out)
	}
}

// The viewport scrolls back into the history, stays put while output
// arrives, and returns to the bottom; Render draws what it shows.
func TestTheViewportScrollsBackAndReturns(t *testing.T) {
	tm, err := New(20, 5, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer tm.Close()
	for i := 1; i <= 30; i++ {
		tm.Write([]byte(fmt.Sprintf("line %02d\r\n", i)))
	}
	if b := tm.Back(); b != 0 {
		t.Fatalf("at the start the viewport is %d rows back", b)
	}
	tm.Scroll(-10)
	if b := tm.Back(); b != 10 {
		t.Fatalf("after scrolling up 10 the viewport is %d rows back", b)
	}
	if got := string(tm.Render(true, 0)); !strings.Contains(got, "line 17") || strings.Contains(got, "line 30") {
		t.Fatalf("scrolled back, the render shows:\n%q", got)
	}
	tm.Write([]byte("line 31\r\n"))
	if got := string(tm.Render(true, 0)); !strings.Contains(got, "line 17") {
		t.Fatalf("new output moved a scrolled-back viewport:\n%q", got)
	}
	tm.Scroll(-1000)
	if got := string(tm.Render(true, 0)); !strings.Contains(got, "line 01") {
		t.Fatalf("scrolled to the top, the render lacks line 01:\n%q", got)
	}
	tm.ScrollToBottom()
	if b := tm.Back(); b != 0 {
		t.Fatalf("after the return the viewport is %d rows back", b)
	}
	if got := string(tm.Render(true, 0)); !strings.Contains(got, "line 31") {
		t.Fatalf("back at the bottom, the render lacks line 31:\n%q", got)
	}
}
