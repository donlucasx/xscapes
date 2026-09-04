package host

import (
	"fmt"
	"strings"
	"testing"
)

// The whole point, end to end: rows that scroll out of the agent's band end
// up in the MAIN buffer, in order, while the band keeps what is still in it.
// The snapshot's model holds both buffers, so the main one is read here the
// way Terminal.app would show it above the alternate screen.
func TestRowsLeavingTheBandAreMirroredIntoTheMainBuffer(t *testing.T) {
	child := `sleep 0.35; i=1; while [ $i -le 24 ]; do printf 'AGENTLINE%02d\r\n' $i; i=$((i+1)); done; sleep 1.2`
	sc := runHostedOpts(t, 80, 30, 80, 30, "alt", child, true, true)

	band, _ := Band(30)
	if !sc.alt {
		t.Fatal("the snapshot is not on the alternate screen; the band under test is not on display")
	}
	// 24 lines through a 17-row band: the newline after line 17 scrolls, so
	// lines 1..8 left the top (the last CRLF scrolls once more).
	left := 24 - band + 1
	var main []string
	for y := 0; y < sc.h; y++ {
		if r := sc.otherRowAt(y); strings.Contains(r, "AGENTLINE") {
			main = append(main, r)
		}
	}
	var want []string
	for i := 1; i <= left; i++ {
		want = append(want, fmt.Sprintf("AGENTLINE%02d", i))
	}
	if strings.Join(main, ",") != strings.Join(want, ",") {
		t.Errorf("main buffer holds %v\nwant %v", main, want)
	}
	// And the band still shows the rest, untouched by the mirroring.
	if top := sc.rowAt(0); top != fmt.Sprintf("AGENTLINE%02d", left+1) {
		t.Errorf("band row 1 is %q, want AGENTLINE%02d", top, left+1)
	}
	for y := 0; y < sc.h; y++ {
		if strings.Contains(sc.rowAt(y), "AGENTLINE") && y >= band {
			t.Errorf("agent text on row %d, below the band", y+1)
		}
	}
}

// With the mirror off nothing may be written to the main buffer at all: that
// is the state every other test in this package runs in, and the control for
// the one above.
func TestNothingReachesTheMainBufferWithHistoryOff(t *testing.T) {
	sc := runHostedAlt(t, 80, 30, 80, 30, "alt", agentLines, true)
	for y := 0; y < sc.h; y++ {
		if r := sc.otherRowAt(y); strings.Contains(r, "AGENTLINE") {
			t.Fatalf("main buffer row %d holds agent text with the mirror off: %q", y+1, r)
		}
	}
}

// MirrorBatch walks mainRow down the main buffer and scrolls once it reaches
// the last row, so the transcript starts under the shell's command and then
// flows into history one row at a time.
func TestMirrorBatchWalksThenScrolls(t *testing.T) {
	sc := newScreen(20, 5)
	sc.feed("prompt$ cmd\r\n") // the shell left its cursor on row 2
	sc.feed(Open(true, 3, 5))
	mainRow := 2
	sc.feed(MirrorBatch([]string{"one", "two", "three", "four", "five"}, &mainRow, 5, 3))
	if !sc.alt {
		t.Fatal("the batch must leave the alternate screen on display")
	}
	got := []string{sc.otherRowAt(0), sc.otherRowAt(1), sc.otherRowAt(2), sc.otherRowAt(3), sc.otherRowAt(4)}
	// Rows 2..5 took one..four; "five" scrolled the prompt into history and
	// took the last row, so the newest row sits at the bottom with no blank
	// row under it. Every row from here on does the same.
	want := []string{"one", "two", "three", "four", "five"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("main buffer %v, want %v", got, want)
	}
	if mainRow != 6 {
		t.Errorf("mainRow %d after filling the buffer, want 6 (past the last row: full, scroll first)", mainRow)
	}
	if sc.top != 0 || sc.bot != 2 {
		t.Errorf("band not re-pinned: region %d..%d", sc.top+1, sc.bot+1)
	}
	if MirrorBatch(nil, &mainRow, 5, 3) != "" {
		t.Error("an empty batch must write nothing")
	}
}

// The one reply the host asks for is taken out of the key stream; everything
// around it, and any later report, goes through.
func TestTakeCPRFindsTheReportAndKeepsTheRest(t *testing.T) {
	row, rest, ok := takeCPR([]byte("ab\x1b[12;1Rcd"))
	if !ok || row != 12 || string(rest) != "abcd" {
		t.Errorf("got row=%d rest=%q ok=%v", row, rest, ok)
	}
	if _, _, ok := takeCPR([]byte("\x1b[12;")); ok {
		t.Error("a cut report must not match yet")
	}
	if _, _, ok := takeCPR([]byte("\x1b[A\x1b[B")); ok {
		t.Error("arrow keys are not a report")
	}
	if row, rest, ok := takeCPR([]byte("\x1b[A\x1b[3;7R")); !ok || row != 3 || string(rest) != "\x1b[A" {
		t.Errorf("report after a key: row=%d rest=%q ok=%v", row, rest, ok)
	}
}
