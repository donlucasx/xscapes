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
		h := &Host{model: s, out: null, mainRow: 1}
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
		// The agent repaints the band (Claude Code after a resize) and scrolls
		// one new row off: that row is new, and 02, 03 can come back again.
		s.feed("\x1b[1;1H\x1b[2Kfresh\x1b[12;1H\n")
		h.mirror(12, 12)
		got := strings.Join(mirrored(h), ",")
		if got != "row 01,row 02,row 03,fresh" {
			t.Errorf("mirrored %s, want the new row kept", got)
		}
		if again := h.pullBack(3, 40); len(again) != 3 || stripANSI(again[0]) != "row 02" || stripANSI(again[2]) != "fresh" {
			var plain []string
			for _, l := range again {
				plain = append(plain, stripANSI(l))
			}
			t.Errorf("a second grow brings back %q, want rows 02, 03 and fresh", plain)
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
