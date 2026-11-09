package host

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// The rows a grow brings back (pullBack) are already in the terminal's
// scrollback. When the agent scrolls them off the band again they must not be
// mirrored a second time: duplicated scrollback is the defect he reported
// through six sessions in September (see resizeScrolling).
func TestRowsAGrowBroughtBackAreNotMirroredTwice(t *testing.T) {
	setup := func() (*Host, *screen) {
		s := newScreen(40, 12)
		s.alt, s.capture = true, true
		null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { null.Close() })
		h := &Host{model: s, out: null, mainRow: 1, History: true}
		var b strings.Builder
		for i := 1; i <= 12; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1Hrow %02d", i, i)
		}
		s.feed(b.String())
		s.takeScrolled()
		// The agent scrolls three rows off: 01, 02, 03 are mirrored.
		s.feed("\x1b[12;1H\n\n\n")
		h.mirror(12, 12)
		return h, s
	}
	mirrored := func(h *Host) []string {
		var out []string
		for _, l := range h.mirrored {
			out = append(out, stripANSI(l))
		}
		return out
	}

	t.Run("brought back, then scrolled off again", func(t *testing.T) {
		h, s := setup()
		back := h.pullBack(2, 40)
		if len(back) != 2 || stripANSI(back[0]) != "row 02" || stripANSI(back[1]) != "row 03" {
			t.Fatalf("pullBack(2) = %q, want rows 02 and 03", back)
		}
		// Paint them back at the top the way RebindGrowAlt does, then let the
		// agent scroll four rows: 02 and 03 leave again, then 04 and 05.
		s.feed("\x1b[2T\x1b[1;1Hrow 02\x1b[2;1Hrow 03\x1b[12;1H\n\n\n\n")
		h.mirror(12, 12)
		got := strings.Join(mirrored(h), ",")
		if got != "row 01,row 02,row 03,row 04,row 05" {
			t.Errorf("mirrored %s, want each row once", got)
		}
		if len(h.back) != 0 {
			t.Errorf("%d rows still marked as brought back", len(h.back))
		}
	})

	t.Run("drawn over, the next row is mirrored and the rows can come back again", func(t *testing.T) {
		h, s := setup()
		h.pullBack(2, 40)
		// The agent draws over the top row and scrolls one new row off: that
		// row is new, and 02, which it covered, can come back again. 03 is
		// still on screen under it, so it is not (Kimi N2).
		s.feed("\x1b[1;1H\x1b[2Kfresh\x1b[12;1H\n")
		h.mirror(12, 12)
		got := strings.Join(mirrored(h), ",")
		if got != "row 01,row 02,row 03,fresh" {
			t.Errorf("mirrored %s, want the new row kept", got)
		}
		if again := h.pullBack(3, 40); len(again) != 3 || stripANSI(again[0]) != "row 01" || stripANSI(again[1]) != "row 02" || stripANSI(again[2]) != "fresh" {
			var plain []string
			for _, l := range again {
				plain = append(plain, stripANSI(l))
			}
			t.Errorf("a second grow brings back %q, want rows 01, 02 and fresh", plain)
		}
	})
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i++; i < len(s) && (s[i] < '@' || s[i] > '~' || s[i] == '['); i++ {
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return strings.TrimRight(b.String(), " ")
}

// Kimi's two reproductions from the 2026-11-08 audit, kept as regressions.
// Both were RED on f166952 (its audit run: "row 02" and "row 03" mirrored twice).
func TestBroughtBackRowsAreMirroredOnceThroughANarrowingAndAPartialOverwrite(t *testing.T) {
	newHost := func(t *testing.T, w int, line func(i int) string) (*Host, *screen) {
		s := newScreen(w, 12)
		s.alt, s.capture = true, true
		null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { null.Close() })
		h := &Host{model: s, out: null, mainRow: 1, History: true}
		var b strings.Builder
		for i := 1; i <= 12; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H%s", i, line(i))
		}
		s.feed(b.String())
		s.takeScrolled()
		s.feed("\x1b[12;1H\n\n\n") // 01, 02, 03 leave and are mirrored
		h.mirror(12, 12)
		return h, s
	}
	count := func(h *Host, what string) int {
		n := 0
		for _, l := range h.mirrored {
			if strings.Contains(stripANSI(l), what) {
				n++
			}
		}
		return n
	}

	t.Run("the window narrows after the grow (N1)", func(t *testing.T) {
		long := func(i int) string { return fmt.Sprintf("row %02d -- a tail long enough to be cut", i) }
		h, s := newHost(t, 40, long)
		h.pullBack(2, 40)
		s.feed("\x1b[2T\x1b[1;1H" + long(2) + "\x1b[2;1H" + long(3))
		s.resizeAlt(20, 12)
		s.feed("\x1b[12;1H\n\n")
		h.mirror(12, 12)
		if n := count(h, "row 02"); n != 1 {
			t.Errorf("row 02 mirrored %d times, want once", n)
		}
	})

	t.Run("the agent draws over the top one only (N2)", func(t *testing.T) {
		short := func(i int) string { return fmt.Sprintf("row %02d", i) }
		h, s := newHost(t, 40, short)
		h.pullBack(2, 40)
		s.feed("\x1b[2T\x1b[1;1Hrow 02\x1b[2;1Hrow 03")
		s.feed("\x1b[1;1H\x1b[2Kfresh\x1b[12;1H\n\n")
		h.mirror(12, 12)
		if n := count(h, "row 03"); n != 1 {
			t.Errorf("row 03 mirrored %d times, want once", n)
		}
		if n := count(h, "fresh"); n != 1 {
			t.Errorf("the new row mirrored %d times, want once", n)
		}
	})
}

// A row a shrink pushed off was never mirrored (a resize is not a scroll).
// Refilled at the band's top and then scrolled off by the agent, it goes to
// the scrollback then, once; a row a SCROLL took, refilled and scrolled off
// again, does not go twice.
func TestARowAShrinkTookIsMirroredOnceWhenItFinallyLeaves(t *testing.T) {
	s := newScreen(40, 12)
	s.alt, s.capture = true, true
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { null.Close() })
	h := &Host{model: s, out: null, mainRow: 1, History: true}
	var b strings.Builder
	for i := 1; i <= 12; i++ {
		fmt.Fprintf(&b, "\x1b[%d;1Hrow %02d", i, i)
	}
	s.feed(b.String())
	s.takeScrolled()
	s.feed("\x1b[12;1H\n") // row 01 scrolls off: mirrored
	h.mirror(12, 12)

	s.resizeGhostty(40, 9) // rows 02-04 pushed off by the shrink
	h.keepResized()
	back := h.pullBack(2, 40) // the shrink's refill: rows 03 and 04
	if len(back) != 2 || stripANSI(back[0]) != "row 03" {
		t.Fatalf("refill = %q, want rows 03 and 04", back)
	}
	s.feed("\x1b[2T\x1b[1;1Hrow 03\x1b[2;1Hrow 04\x1b[9;1H\n\n\n")
	h.mirror(9, 9)
	var got []string
	for _, l := range h.mirrored {
		got = append(got, stripANSI(l))
	}
	if strings.Join(got, ",") != "row 01,row 03,row 04,row 05" {
		t.Errorf("mirrored %q, want row 01 then 03, 04 (first time out) and 05", got)
	}
}
