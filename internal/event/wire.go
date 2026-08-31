package event

import (
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxLine caps an encoded event. macOS caps a unix datagram at
// net.local.dgram.maxdgram, which is 2048 here and returns EMSGSIZE above it,
// so an event that does not fit the datagram would silently take the slow path
// forever. Capping below that means socket and file always carry identical
// bytes.
const MaxLine = 2000

// Encode renders one line, without the newline. It never fails and never
// exceeds MaxLine: oversized events are shortened field by field rather than
// dropped, because the fields that get long (a pasted prompt, a shell command)
// are exactly the ones whose *existence* matters more than their tail.
func Encode(e Event) []byte {
	if e.V == 0 {
		e.V = 1
	}
	e.Text = clean(e.Text)
	e.Target = clean(e.Target)
	e.Detail = clean(e.Detail)
	e.Tool = clean(e.Tool)

	b, err := json.Marshal(e)
	if err != nil {
		// Marshal of this struct cannot fail, but a silent empty line would
		// be worse than a visible unknown one.
		return []byte(`{"v":1,"kind":"error","text":"encode failed"}`)
	}
	if len(b) <= MaxLine {
		return b
	}

	// Shorten in order of how much a glance needs it: prose first, then the
	// result, then the subject. Recompute rather than estimate -- JSON
	// escaping means byte length is not string length.
	for _, cut := range []func(*Event){
		func(x *Event) { x.Text = trunc(x.Text, 160) },
		func(x *Event) { x.Detail = trunc(x.Detail, 80) },
		func(x *Event) { x.Text = "" },
		func(x *Event) { x.Target = trunc(x.Target, 120) },
		func(x *Event) { x.Detail = "" },
		func(x *Event) { x.Target = trunc(x.Target, 60) },
	} {
		cut(&e)
		if b, err = json.Marshal(e); err == nil && len(b) <= MaxLine {
			return b
		}
	}
	// Nothing variable left. The identity fields are short by construction but
	// nothing enforces it -- an adapter is free to send a kilobyte of session
	// id -- and a skeleton over MaxLine cannot be sent at all, so bound them.
	b, _ = json.Marshal(Event{
		V: e.V, TS: e.TS, Session: trunc(e.Session, 64), Kind: Kind(trunc(string(e.Kind), 32)),
		Op: Op(trunc(string(e.Op), 32)), ID: trunc(e.ID, 64), Src: trunc(e.Src, 32),
	})
	if len(b) > MaxLine {
		b, _ = json.Marshal(Event{V: 1, Kind: e.Kind})
	}
	return b
}

// Decode parses one line. A line we cannot parse is an error, not a panic and
// not a zero Event: the caller counts them, because a rising parse-error count
// is how you find out an adapter is writing a format you did not agree to.
func Decode(line []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(line, &e); err != nil {
		return Event{}, err
	}
	if e.V == 0 {
		e.V = 1
	}
	// Clean on the way IN as well as out. The emitter is not necessarily
	// ours -- the protocol is meant to be written by other people's adapters
	// and by hand -- and everything below gets painted into a terminal.
	e.Text = clean(e.Text)
	e.Target = clean(e.Target)
	e.Detail = clean(e.Detail)
	e.Tool = clean(e.Tool)
	e.AgentType = clean(e.AgentType)
	return e, nil
}

// clean strips control characters and collapses whitespace.
//
// This is not tidiness. Every one of these strings is derived from something
// the agent touched -- a filename, a shell command, an error message -- and
// gets printed into a terminal by a program that also emits its own escape
// sequences for cursor positioning. A file named with a CSI sequence would
// otherwise be able to move the cursor, repaint the screen, or hide itself.
func clean(s string) string {
	if s == "" {
		return s
	}
	// The guard has to cover everything the slow path would change, or the
	// fast path becomes a hole: newline, carriage return and tab are exactly
	// the characters that break a frame, and they were being waved through.
	if !strings.ContainsFunc(s, func(r rune) bool {
		return badRune(r) || r == '\n' || r == '\r' || r == '\t'
	}) && !strings.Contains(s, "  ") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	space := false
	for _, r := range s {
		switch {
		case badRune(r):
			// Drop outright rather than substitute: a placeholder in a path
			// reads as part of the path.
			continue
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			space = true
		default:
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
		}
	}
	return b.String()
}

func badRune(r rune) bool {
	if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
		return false
	}
	// C0, DEL and C1, plus the line/paragraph separators that some terminals
	// treat as newlines, plus anything that did not decode.
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) ||
		r == 0x2028 || r == 0x2029 || r == utf8.RuneError ||
		unicode.Is(unicode.Cf, r)
}

// trunc shortens to n runes with a single-character ellipsis, keeping the end
// of a path rather than the start: "…/internal/scape/shore.go" identifies the
// file, "/Users/lucasgarzoli/Doc…" identifies nothing.
func trunc(s string, n int) string {
	if s == "" || utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	if strings.Contains(s, "/") {
		return "…" + string(r[len(r)-n+1:])
	}
	return string(r[:n-1]) + "…"
}
