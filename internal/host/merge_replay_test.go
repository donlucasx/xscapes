package host

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// Replay a window of a real trace and report the FIRST byte offset at which a
// row in the model holds both rule glyphs and words -- the merge, caught as it
// FORMS, instead of found afterwards in the mirror.
//
//	MERGE_TRACE=~/.config/xscapes/traces/<t>.bin go test ./internal/host -run TestFindTheMerge -v
//
// It caught one on 2026-09-09 at offset 7231212224: the agent's status line
// drawn on top of its own full-width rule, columns 1-2 of the rule surviving
// because the status starts at ESC[3G.
//
// ⚠ It starts MID-TRACE with a cold model, so the first rows it reports may be
// the cold start rather than the defect. Confirm a hit by starting the window
// earlier and checking the same row still merges; the agent repaints its whole
// page often enough that a few hundred KB of lead-in is usually enough.
func TestFindTheMerge(t *testing.T) {
	path := os.Getenv("MERGE_TRACE")
	if path == "" {
		t.Skip("set MERGE_TRACE")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var from, to int64 = 7230000000, 7237000000
	if _, err := f.Seek(from, 0); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, to-from)
	n, _ := f.Read(buf)
	buf = buf[:n]
	t.Logf("replaying %d bytes from offset %d", n, from)

	sc := newScreen(153, 61)
	sc.capture = true
	sc.feed("\x1b[?47h") // enter the alt screen the way the host does

	word := regexp.MustCompile(`[A-Za-z]{4}`)
	merged := func() (int, string) {
		for y := 0; y < sc.h; y++ {
			r := sc.rowAt(y)
			if strings.Contains(r, "─") && word.MatchString(r) {
				return y, r
			}
		}
		return -1, ""
	}
	const step = 64
	for i := 0; i < len(buf); i += step {
		j := i + step
		if j > len(buf) {
			j = len(buf)
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					ctx := strings.ReplaceAll(string(buf[max(0, i-200):j]), "\x1b", "<ESC>")
					t.Fatalf("PANIC %v at byte %d (offset %d)\n  h=%d top=%d bot=%d len(cells)=%d\n  bytes: %s",
						r, i, from+int64(i), sc.h, sc.top, sc.bot, len(sc.cells), ctx)
				}
			}()
			sc.feed(string(buf[i:j]))
		}()
		if y, row := merged(); y >= 0 {
			t.Logf("FIRST MERGE at byte %d (offset %d), row %d:", i, from+int64(i), y)
			t.Logf("   %q", row[:min(140, len(row))])
			lo := i - 400
			if lo < 0 {
				lo = 0
			}
			ctx := strings.ReplaceAll(string(buf[lo:j]), "\x1b", "<ESC>")
			t.Logf("   preceding bytes:\n%s", ctx[max(0, len(ctx)-700):])
			return
		}
	}
	t.Log("no merge formed in this window")
}
