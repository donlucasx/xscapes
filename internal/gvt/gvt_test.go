//go:build ghosttyvt

package gvt

import (
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
