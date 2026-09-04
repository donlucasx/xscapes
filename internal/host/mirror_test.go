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
	// The harness answered the host's cursor query with row 7, so the
	// transcript begins there: right under where the shell's command would
	// be, not at the bottom of the buffer.
	if got := sc.otherRowAt(6); got != "AGENTLINE01" {
		t.Errorf("main buffer row 7 is %q, want AGENTLINE01 (the DSR answer was not used)", got)
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

// The replay must not leave the tails of the rows it writes over. Terminal.app
// keeps the rows still on the main screen at 1049l and restores the shell's
// cursor onto them, so the replay lands on its own survivors; a shorter line
// over a longer one showed "LIVE LINE 80" in the first live run.
func TestReplayErasesTheRowsItWritesOver(t *testing.T) {
	sc := newScreen(40, 8)
	sc.feed("prompt$ xscapes claude\r\n")
	// What Terminal.app kept: survivors longer than what will be replayed.
	sc.feed("LIVE LINE 10 and a long tail\r\nLIVE LINE 11 and a long tail\r\nLIVE LINE 12\r\n")
	sc.feed("\x1b[2;1H") // the cursor 1049l restores: under the command, on the survivors
	h := &Host{mirrored: []string{"LIVE LINE 8", "", "LIVE LINE 9"}}
	sc.feed(h.replay())
	got := []string{sc.rowAt(1), sc.rowAt(2), sc.rowAt(3), sc.rowAt(4), sc.rowAt(5)}
	want := []string{"xscapes: the agent's transcript, 3 rows", "LIVE LINE 8", "", "LIVE LINE 9", ""}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("after the replay rows 2..6 are %q, want %q (header on the restored row, nothing left below)", got, want)
	}
}

// The end of the session is in the replay too: the rows that scrolled after
// the last tick and the band as it stood at exit -- a plain session leaves
// that screen behind when it ends, and so must this one.
func TestExitReplayCarriesTheTranscriptAndTheFinalScreen(t *testing.T) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	tty, err := openPTY()
	if err != nil {
		t.Skipf("no pty available: %v", err)
	}
	defer tty.Close()
	sc := newScreen(80, 30)
	var mu sync.Mutex
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				mu.Lock()
				sc.feed(string(buf[:n]))
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	// Twenty-four lines, one of them red, then exit at once: no sleep, so the
	// tail of the output and the band's final content both have to be picked
	// up by the exit path, not by a tick.
	child := `i=1; while [ $i -le 24 ]; do if [ $i -eq 5 ]; then printf '\033[31mAGENTLINE%02d\033[0m\r\n' $i; else printf 'AGENTLINE%02d\r\n' $i; fi; i=$((i+1)); done`
	h := &Host{
		Cmd: exec.Command("sh", "-c", child), FPS: 60, AltScreen: true, History: true, Replay: true,
		In: tty.slave, Out: pw,
		Size:  func() (int, int) { return 80, 30 },
		Paint: func(cols, rows int) []string { return nil },
	}
	if err := h.Run(); err != nil && err != io.EOF {
		t.Fatalf("host: %v", err)
	}
	pw.Close()
	<-done
	mu.Lock()
	defer mu.Unlock()
	if sc.alt {
		t.Fatal("the screen was not handed back")
	}
	var seen []string
	headerAt := -1
	for y := 0; y < sc.h; y++ {
		r := sc.rowAt(y)
		if strings.HasPrefix(r, "xscapes: the agent's transcript") {
			headerAt = y
		}
		if headerAt >= 0 && y > headerAt && strings.HasPrefix(r, "AGENTLINE") {
			seen = append(seen, r)
		}
	}
	if headerAt < 0 {
		t.Fatalf("no replay header on the main screen; rows: %q", allRows(sc))
	}
	var want []string
	for i := 1; i <= 24; i++ {
		want = append(want, fmt.Sprintf("AGENTLINE%02d", i))
	}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Errorf("the replay carries %v\nwant all 24 lines in order (rows that left the band AND the final screen)", seen)
	}
	// The rendition came through the pipeline, not just the glyphs.
	for y := headerAt + 1; y < sc.h; y++ {
		if strings.HasPrefix(sc.rowAt(y), "AGENTLINE05") {
			if fg := sc.cells[y][0].fg; fg != 1 {
				t.Errorf("AGENTLINE05 was red (fg 1) when the agent wrote it; the replay has fg %d", fg)
			}
		}
	}
}

func allRows(sc *screen) []string {
	var out []string
	for y := 0; y < sc.h; y++ {
		out = append(out, sc.rowAt(y))
	}
	return out
}

// Keys typed while the host waits for its cursor report are released the
// moment the wait ends, even when no report ever comes.
func TestForwardKeysReleasesHeldKeysWhenTheWaitEnds(t *testing.T) {
	inR, inW := io.Pipe()
	var got syncBuf
	h := &Host{}
	h.wantCPR.Store(true)
	waitDone := make(chan struct{})
	cpr := make(chan int, 1)
	go h.forwardKeys(inR, &got, cpr, waitDone)
	inW.Write([]byte("abc"))
	time.Sleep(50 * time.Millisecond)
	if got.String() != "" {
		t.Fatalf("keys forwarded during the wait: %q", got.String())
	}
	h.wantCPR.Store(false)
	close(waitDone)
	time.Sleep(50 * time.Millisecond)
	if got.String() != "abc" {
		t.Errorf("held keys not released when the wait ended: %q", got.String())
	}
	inW.Write([]byte("\x1b[9;1R")) // a late report is just bytes now
	time.Sleep(50 * time.Millisecond)
	if got.String() != "abc\x1b[9;1R" {
		t.Errorf("a late report must go through untouched: %q", got.String())
	}
	inW.Close()
}

type syncBuf struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}
