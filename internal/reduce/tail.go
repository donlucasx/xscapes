package reduce

import (
	"fmt"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// TailLen is how many lines the sand holds. Four fits the beach at 80x24 with
// the companion beside it, and more than four stops being scenery and starts
// being a log window -- which the brief rules out: the sea says how much, the
// sand says what, and the agent's own pane is right there for the rest.
const TailLen = 4

// TailTTL is how long a line survives.
//
// Deliberately far longer than TauFall. The sea is *now* and settles in
// seconds; the sand is the record of what was done and recedes over minutes.
// Right after a turn ends there is a window where the water is flat and the
// sand still reads busy, and that is the intended reading rather than a
// mismatch: the tide has gone out, the writing is still there.
const TailTTL = 3 * time.Minute

// Line is one entry in the sand, already formatted.
type Line struct {
	Text string
	// Age is 0 for the newest and 1 for one about to be taken by the tide.
	// The renderer fades on this, so the caller never does the arithmetic.
	Age float64
	// Bad marks a failure, so the sand can carry the same worry the
	// companion does without the companion being the only place to see it.
	Bad bool

	// src is kept so the line can be re-rendered to a narrower budget without
	// the reducer having to know the terminal's width.
	src line
}

type line struct {
	at     time.Time
	op     event.Op
	tool   string
	target string
	detail string
	ms     int64
	bad    bool
}

type tail struct {
	items []line
}

func (t *tail) push(l line) {
	t.items = append(t.items, l)
	if len(t.items) > TailLen*2 {
		t.items = append(t.items[:0:0], t.items[len(t.items)-TailLen*2:]...)
	}
}

// Fit renders the tail to a column budget.
//
// Chopping a line where it runs out of room produces `edit internal/auth/ha`,
// which names no file at all. Drop whole pieces in order of how much they
// carry instead: the result column first, then the directories, so a narrow
// pane still reads `edit handler.go` -- short, but it still says what happened
// and to what.
func (t *tail) fit(now time.Time, cols int) []Line {
	out := t.lines(now)
	for i := range out {
		out[i].Text = fitLine(out[i].src, cols)
	}
	return out
}

func fitLine(l line, cols int) string {
	if cols <= 0 {
		return ""
	}
	for _, form := range []func(line) string{
		func(x line) string { return format(x) },
		func(x line) string { x.detail, x.ms = "", 0; return format(x) },
		func(x line) string { x.target = fileName(x.target); return format(x) },
		func(x line) string { x.detail, x.ms = "", 0; x.target = fileName(x.target); return format(x) },
		func(x line) string { return strings.TrimSpace(verbOf(x)) + " " + fileName(x.target) },
	} {
		if s := form(l); len([]rune(s)) <= cols {
			return s
		}
	}
	// Even the shortest form does not fit: keep the verb, which at least says
	// the agent is doing something rather than nothing.
	return trimRunes(strings.TrimSpace(verbOf(l)), cols)
}

// fileName keeps the filename and drops the directories.
func fileName(s string) string {
	if i := strings.LastIndexByte(s, '/'); i >= 0 && i+1 < len(s) {
		return s[i+1:]
	}
	return s
}

func trimRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (t *tail) lines(now time.Time) []Line {
	if len(t.items) == 0 {
		return nil
	}
	out := make([]Line, 0, TailLen)
	for _, it := range t.items {
		age := now.Sub(it.at)
		if age > TailTTL {
			continue
		}
		out = append(out, Line{
			Text: format(it),
			Age:  float64(age) / float64(TailTTL),
			Bad:  it.bad,
			src:  it,
		})
	}
	if len(out) > TailLen {
		out = out[len(out)-TailLen:]
	}
	return out
}

// format writes one line of sand: what it did, to what, and how it went.
//
// The verb is the op rather than the tool name, because "read" is legible to
// someone who has never used this agent and "Read" is a Claude Code noun. The
// target keeps its tail rather than its head -- the end of a path identifies
// the file, the start identifies the home directory.
func verbOf(l line) string {
	verb := string(l.op)
	if verb == "" || l.op == event.OpOther {
		verb = strings.ToLower(l.tool)
	}
	if verb == "" {
		verb = "run"
	}
	return fmt.Sprintf("%-5s", verb)
}

func format(l line) string {
	parts := []string{verbOf(l)}
	if l.target != "" {
		parts = append(parts, l.target)
	}

	detail := l.detail
	if detail == "" && l.ms >= 1000 {
		detail = fmt.Sprintf("%.1fs", float64(l.ms)/1000)
	}
	if detail != "" {
		parts = append(parts, detail)
	}
	return strings.Join(parts, "  ")
}
