package host

import (
	"strings"
	"testing"
)

// The mirror lives or dies on one question: which rows left the band, and
// what was in them. These pin the model's answer.

func TestRowsScrollingOffTheAltBandAreKept(t *testing.T) {
	sc := newScreen(20, 6)
	sc.capture = true
	sc.feed(Open(true, 3, 6)) // alt screen, band 1..3, origin mode
	for i := 1; i <= 5; i++ {
		sc.feed("line " + string(rune('0'+i)) + "\r\n")
	}
	got := sc.takeScrolled()
	lines := rowTexts(got)
	// Five lines through a three-row band: the newline on the last row
	// scrolls, so lines 1, 2 and 3 left the top in that order.
	want := []string{"line 1", "line 2", "line 3"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("kept %q, want %q", lines, want)
	}
	if again := sc.takeScrolled(); len(again) != 0 {
		t.Errorf("takeScrolled handed the same rows over twice: %d", len(again))
	}
}

func TestNothingIsKeptOnTheMainScreen(t *testing.T) {
	sc := newScreen(20, 4)
	sc.capture = true
	sc.feed(EnterBand(2))
	for i := 0; i < 6; i++ {
		sc.feed("main\r\n")
	}
	if got := sc.takeScrolled(); len(got) != 0 {
		t.Errorf("%d rows kept from the main screen; the terminal keeps those itself", len(got))
	}
}

// The host's own scrolls open blank rows -- the SU that undoes a grow, the SD
// that puts the text back after a shrink -- and those rows, if nothing is ever
// written into them, must not be mirrored as blank lines the agent never
// produced. A blank line the agent DID write is content and is kept.
func TestHostInsertedBlankRowsAreNotKeptButAgentBlankLinesAre(t *testing.T) {
	sc := newScreen(20, 8)
	sc.capture = true
	sc.feed(Open(true, 6, 8))
	sc.feed("a\r\n\r\nb\r\n") // a, blank, b on rows 1..3, cursor on 4
	// The host scrolls the whole screen down two, the way RebindShrinkAlt does.
	sc.feed("\x1b7\x1b[?6l\x1b[r\x1b[0m\x1b[2T\x1b8" + EnterBand(6))
	// Rows now: host, host, a, blank, b, blank. Scroll all six through.
	sc.feed("\x1b[6;1H")
	for i := 0; i < 6; i++ {
		sc.feed("\r\n")
	}
	lines := rowTexts(sc.takeScrolled())
	want := []string{"a", "", "b", ""}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("kept %q, want %q (the two host rows skipped, the agent's blanks kept)", lines, want)
	}
}

func TestAShrinkKeepsTheRowsTheTerminalDestroys(t *testing.T) {
	sc := newScreen(20, 8)
	sc.capture = true
	sc.feed(Open(true, 5, 8))
	for i := 1; i <= 4; i++ {
		sc.feed("\x1b[" + string(rune('0'+i)) + ";1Hrow" + string(rune('0'+i)))
	}
	sc.resizeAlt(20, 6) // two rows lost off the top
	lines := rowTexts(sc.takeScrolled())
	want := []string{"row1", "row2"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("kept %q, want %q", lines, want)
	}
	if sc.rowAt(0) != "row3" {
		t.Errorf("row 1 after the shrink is %q, want row3", sc.rowAt(0))
	}
}

// DECSET 47 and 1049 are not the same switch, and the mirror depends on the
// difference: 47 swaps buffers and clears nothing (measured, 400 round trips),
// 1049 saves, clears and on the way back discards the alternate screen.
func TestMode47SwapsBuffersWithoutClearingAnd1049DoesNot(t *testing.T) {
	sc := newScreen(20, 4)
	sc.feed("shell\r\n")          // main buffer, row 1
	sc.feed("\x1b[?1049h")        // alternate, blank, cursor saved
	sc.feed("\x1b[1;1Halt row 1") // draw on the alternate screen
	sc.feed("\x1b[?47l")          // to the main buffer, no clear
	if !strings.HasPrefix(sc.rowAt(0), "shell") {
		t.Fatalf("after 47l the main buffer should be on display; row 1 is %q", sc.rowAt(0))
	}
	sc.feed("\x1b[4;1Hmirrored\r\n") // write at the main buffer's last row and scroll it
	sc.feed("\x1b[?47h")             // back to the alternate screen
	if sc.rowAt(0) != "alt row 1" {
		t.Errorf("47h cleared the alternate screen: row 1 is %q, want %q", sc.rowAt(0), "alt row 1")
	}
	if sc.otherRowAt(2) != "mirrored" {
		t.Errorf("the mirrored row is not in the main buffer where it was written: rows %q %q %q %q",
			sc.otherRowAt(0), sc.otherRowAt(1), sc.otherRowAt(2), sc.otherRowAt(3))
	}
	sc.feed("\x1b[?1049l")
	if sc.rowAt(2) != "mirrored" || sc.alt {
		t.Errorf("1049l did not bring the main buffer back with its rows: row 3 %q alt=%v", sc.rowAt(2), sc.alt)
	}
	if sc.other != nil {
		t.Error("1049l should discard the alternate screen")
	}
}

// A row written back out draws the same cells it was read from.
func TestRowANSIRoundTrips(t *testing.T) {
	src := "\x1b[1;38;5;214mbold orange\x1b[0m plain \x1b[48;2;10;20;30;4m under blue\x1b[0m \x1b[7mrev\x1b[0m tail  "
	a := newScreen(60, 2)
	a.feed(src)
	out := rowANSI(a.cells[0])
	b := newScreen(60, 2)
	b.feed(out)
	for x := 0; x < 60; x++ {
		if a.cells[0][x] != b.cells[0][x] {
			t.Fatalf("cell %d differs after the round trip: %+v vs %+v\nansi: %q", x, a.cells[0][x], b.cells[0][x], out)
		}
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Errorf("a mirrored row must end in a reset: %q", out)
	}
	if strings.HasSuffix(strings.TrimSuffix(out, "\x1b[0m"), "  ") {
		t.Errorf("trailing default cells should be dropped: %q", out)
	}
	if rowANSI(blankRow(10)) != "" {
		t.Error("an empty row serialises to nothing")
	}
}

// rowTexts is the glyphs of each kept row, trailing blanks dropped.
func rowTexts(rows [][]cell) []string {
	var out []string
	for _, row := range rows {
		r := make([]rune, len(row))
		for i, c := range row {
			r[i] = c.r
		}
		out = append(out, strings.TrimRight(string(r), " "))
	}
	return out
}

// A row kept before the window narrows is cut to the new width, or it would
// wrap in the real terminal and the next row's erase would take the tail.
func TestKeptRowsAreCutToTheNewWidth(t *testing.T) {
	sc := newScreen(30, 6)
	sc.capture = true
	sc.feed(Open(true, 3, 6))
	sc.feed(strings.Repeat("x", 30) + "\r\n\r\n\r\n") // a full-width row scrolls off
	sc.resizeAlt(20, 6)
	kept := sc.takeScrolled()
	if len(kept) == 0 {
		t.Fatal("nothing kept")
	}
	if len(kept[0]) != 20 {
		t.Errorf("kept row is %d cells wide after a narrowing to 20", len(kept[0]))
	}
}

// ESC ( B designates a charset and draws nothing. Claude Code emits it at
// startup; the first model drew the 'B'.
func TestCharsetDesignationDrawsNothing(t *testing.T) {
	sc := newScreen(20, 3)
	sc.feed("ab\x1b(Bcd\x1b)0\x1b#8ef")
	if got := sc.rowAt(0); got != "abcdef" {
		t.Errorf("row 1 is %q, want abcdef (the designation's final byte was drawn)", got)
	}
	// Cut across a read boundary, it must still be one sequence.
	sc = newScreen(20, 3)
	sc.feed("x\x1b(")
	sc.feed("By")
	if got := sc.rowAt(0); got != "xy" {
		t.Errorf("split designation: row 1 is %q, want xy", got)
	}
}
