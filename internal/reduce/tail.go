package reduce

import (
	"fmt"
	"strings"
	"time"

	"github.com/donlucasx/asciiscapes/internal/event"
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
func format(l line) string {
	verb := string(l.op)
	if verb == "" || l.op == event.OpOther {
		verb = strings.ToLower(l.tool)
	}
	if verb == "" {
		verb = "run"
	}

	parts := []string{fmt.Sprintf("%-5s", verb)}
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
