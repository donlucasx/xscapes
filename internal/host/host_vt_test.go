//go:build ghosttyvt

package host

import (
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/gvt"
)

var (
	vtCUP = regexp.MustCompile(`\x1b\[(\d+);1H`)
	vtCSI = regexp.MustCompile(`\x1b\[[0-9;:?=]*[A-Za-z]`)
)

// outerRows reads the outer terminal (itself a Ghostty engine) as text rows.
func outerRows(t *gvt.Term, rows int) []string {
	scr := make([]string, rows)
	s := string(t.Render(true, 0))
	locs := vtCUP.FindAllStringSubmatchIndex(s, -1)
	for i, l := range locs {
		y, _ := strconv.Atoi(s[l[2]:l[3]])
		end := len(s)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		if y-1 < rows {
			scr[y-1] = strings.TrimRight(vtCSI.ReplaceAllString(s[l[1]:end], ""), " ")
		}
	}
	return scr
}

// His Terminal.app run of 2026-11-08 20:21 (screenshots 63/64): items 18-20
// were printed while the window was narrow, and after a widen they stayed
// broken at the narrow width, one of them cut short for good. Through the
// engine, the agent's own text reflows both ways, whatever the outer
// terminal does with its alternate screen.
func TestTheEngineHostReflowsTheAgentsTextBothWays(t *testing.T) {
	const fact = "19. Varuna is a fast-rotating Kuiper Belt object whose light curve suggests it may also be stretched"
	child := `printf 'header\r\n` + fact + `\r\n> prompt'; sleep 3`

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	tty, err := openPTY()
	if err != nil {
		t.Skipf("no pty: %v", err)
	}
	defer tty.Close()

	const w1, w2, h = 100, 40, 30
	outer, _ := gvt.New(w1, h, 0)
	defer outer.Close()
	var mu sync.Mutex
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 65536)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				mu.Lock()
				outer.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	var sz struct {
		sync.Mutex
		w int
	}
	sz.w = w1
	hst := &Host{
		Cmd: exec.Command("sh", "-c", child), FPS: 30, Engine: "ghostty",
		In: tty.slave, Out: pw,
		Size: func() (int, int) { sz.Lock(); defer sz.Unlock(); return sz.w, h },
		Paint: func(cols, rows int) []string {
			out := make([]string, rows)
			for i := range out {
				out[i] = strings.Repeat("~", cols)
			}
			return out
		},
	}
	snaps := map[int][]string{}
	go func() {
		look := func(key int) {
			time.Sleep(400 * time.Millisecond)
			mu.Lock()
			snaps[key] = outerRows(outer, h)
			mu.Unlock()
		}
		look(w1)
		for _, nw := range []int{w2, w1} {
			mu.Lock()
			outer.Resize(nw, h) // the real terminal moves first, its own way
			mu.Unlock()
			sz.Lock()
			sz.w = nw
			sz.Unlock()
			look(nw*1000 + 1)
		}
	}()
	if err := hst.Run(); err != nil && err != io.EOF {
		t.Fatalf("host: %v", err)
	}
	pw.Close()
	<-done

	band, _ := Band(h)
	narrow := strings.Join(snaps[w2*1000+1][:band], "\n")
	if !strings.Contains(narrow, fact[:w2]) || !strings.Contains(narrow, fact[w2:2*w2]) {
		t.Errorf("narrow: the fact is not re-wrapped at %d:\n%s", w2, narrow)
	}
	wide := snaps[w1*1000+1][:band]
	if !strings.Contains(strings.Join(wide, "\n"), fact[:w1]) {
		t.Errorf("wide again: the fact did not join back:\n%s", strings.Join(wide, "\n"))
	}
	for i, r := range wide {
		if strings.Contains(r, "~") {
			t.Errorf("band row %d holds scape: %q", i+1, r)
		}
	}
}
