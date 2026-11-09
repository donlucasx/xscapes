//go:build ghosttyvt

package host

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/gvt"
)

// engineRun runs the engine host on a child with a fixed window, an outer
// terminal (a Ghostty engine standing in for the real one) fed every byte
// the host writes, and keys typed after a delay. It returns the outer
// screen's rows just before the child exits, the raw bytes, and the outer
// screen after the exit.
func engineRun(t *testing.T, child string, w, h int, keys string) (during []string, raw string, after string) {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	tty, err := openPTY()
	if err != nil {
		t.Skipf("no pty: %v", err)
	}
	defer tty.Close()
	outer, _ := gvt.New(w, h, 1<<20)
	defer outer.Close()
	var mu sync.Mutex
	var rawB strings.Builder
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 65536)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				mu.Lock()
				rawB.Write(buf[:n])
				outer.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	hst := &Host{
		Cmd: exec.Command("sh", "-c", child), FPS: 30, Engine: "ghostty", Replay: true,
		In: tty.slave, Out: pw,
		Size: func() (int, int) { return w, h },
		Paint: func(cols, rows int) []string {
			out := make([]string, rows)
			for i := range out {
				out[i] = strings.Repeat("~", cols)
			}
			return out
		},
	}
	snapped := make(chan struct{})
	go func() {
		defer close(snapped)
		time.Sleep(600 * time.Millisecond)
		if keys != "" {
			tty.master.Write([]byte(keys))
		}
		time.Sleep(600 * time.Millisecond)
		mu.Lock()
		during = outerRows(outer, h)
		mu.Unlock()
	}()
	if err := hst.Run(); err != nil && err != io.EOF {
		t.Fatalf("host: %v", err)
	}
	pw.Close()
	<-done
	<-snapped
	return during, rawB.String(), string(outer.Transcript(false))
}

// What he types reaches the agent, and the agent's echo is drawn in its band.
func TestTheEngineHostTypesIntoTheAgent(t *testing.T) {
	during, _, _ := engineRun(t, `stty -echo; read line; printf 'got:%s\r\n' "$line"; sleep 1`, 80, 24, "hello there\r")
	if !strings.Contains(strings.Join(during, "\n"), "got:hello there") {
		t.Fatalf("the typed line never reached the agent:\n%s", strings.Join(during, "\n"))
	}
}

// The agent's own question to the terminal (where is my cursor?) is
// answered by the engine, in the agent's own coordinates.
func TestTheEngineAnswersTheAgentsCursorQuery(t *testing.T) {
	child := `stty raw -echo; printf '\033[5;9H\033[6n'; r=$(dd bs=1 count=6 2>/dev/null | od -An -c | tr -d ' \n'); stty sane; printf '\r\nreply:%s\r\n' "$r"; sleep 2`
	during, _, _ := engineRun(t, child, 80, 24, "")
	if !strings.Contains(strings.Join(during, "\n"), `reply:033[5;9R`) {
		t.Fatalf("no cursor report 5;9 reached the agent:\n%s", strings.Join(during, "\n"))
	}
}

// A mode that changes what the real terminal SENDS (bracketed paste) is
// switched on there when the agent asks, and off again when the host exits.
func TestTheEngineHostMirrorsInputModes(t *testing.T) {
	_, raw, _ := engineRun(t, `printf '\033[?2004h> '; sleep 1`, 80, 24, "")
	on := strings.Index(raw, "\x1b[?2004h")
	off := strings.LastIndex(raw, "\x1b[?2004l")
	if on < 0 || off < on {
		t.Fatalf("bracketed paste: on at %d, off at %d", on, off)
	}
}

// At exit the agent's whole transcript goes to the main screen, joined
// where the engine soft-wrapped it, after the alternate screen is left.
func TestTheEngineHostLeavesTheTranscript(t *testing.T) {
	long := strings.Repeat("abcdefghij", 12) // 120 columns, wraps at 80
	child := `i=1; while [ $i -le 30 ]; do printf 'line %02d\r\n' $i; i=$((i+1)); done; printf '` + long + `\r\n'; sleep 0.3`
	_, raw, _ := engineRun(t, child, 80, 24, "")
	leave := strings.LastIndex(raw, "\x1b[?1049l")
	if leave < 0 {
		t.Fatal("the host never left the alternate screen")
	}
	tail := raw[leave:]
	for _, want := range []string{"line 01", "line 30", long} {
		if !strings.Contains(tail, want) {
			t.Errorf("transcript after the exit lacks %q", want[:min(20, len(want))])
		}
	}
	// Every line erases the row it lands on (Terminal.app keeps survivors).
	if !strings.Contains(tail, "\x1b[2Kline 01") && !strings.Contains(tail, "\x1b[2K\x1b[0mline 01") {
		i := strings.Index(tail, "line 01")
		t.Errorf("the transcript's line lands without an erase: %q", tail[max(0, i-16):i+7])
	}
}

// Wide characters take two columns and do not shift what follows.
func TestTheEngineHostDrawsWideCharacters(t *testing.T) {
	during, _, _ := engineRun(t, `printf '日本語|after\r\n🐱|cat\r\n'; sleep 1.5`, 80, 24, "")
	joined := strings.Join(during, "\n")
	if !strings.Contains(joined, "日本語|after") || !strings.Contains(joined, "🐱|cat") {
		t.Fatalf("wide characters:\n%s", joined)
	}
}

// A full-screen agent (its own alternate screen inside the engine) is drawn
// in the band, and its screen goes away when it leaves it.
func TestTheEngineHostHostsAFullScreenAgent(t *testing.T) {
	child := `printf 'before\r\n'; printf '\033[?1049h\033[H\033[2JFULLSCREEN'; sleep 0.8; printf '\033[?1049l'; printf 'back\r\n'; sleep 1`
	during, _, _ := engineRun(t, child, 80, 24, "")
	j := strings.Join(during, "\n")
	if strings.Contains(j, "FULLSCREEN") || !strings.Contains(j, "back") {
		t.Fatalf("after the agent left its full screen:\n%s", j)
	}
}

// The signal path resets the mirrored modes while the frame loop syncs
// them: run under -race, this is the map race Kimi found (2026-11-09).
func TestTheModeMirrorSurvivesASignalMidFrame(t *testing.T) {
	v, _ := gvt.New(20, 5, 0)
	defer v.Close()
	v.Write([]byte("\x1b[?2004h\x1b[?1000h"))
	var m mirroredModes
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			m.sync(v)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			_ = m.reset()
		}
	}()
	wg.Wait()
}
