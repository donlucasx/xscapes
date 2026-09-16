// Package watch is the generic adapter: it turns the raw traffic of a hosted
// program into protocol events for an agent that has no hooks at all.
//
// The brief's second adapter, verbatim: "busy = alive + output activity; done
// = prompt back". Alive is given -- the host runs the program. Output is the
// bytes the program writes to its pty, and "the prompt is back" is the one
// thing a terminal program does when it wants you: it stops writing. So:
//
//   - a keystroke that ends in Enter is a PROMPT, the user handing the agent
//     something;
//   - every second of continuous output is one TOOL_START, which is what feeds
//     the sea and paces the companion at about the rate real tool traffic does
//     (one Claude Code tool call is two events every few seconds);
//   - a burst of output that lasted at least MinWork and then went quiet for
//     Quiet is a DONE, and the sand gets one line: the last thing the program
//     printed, with the escape sequences stripped;
//   - a burst shorter than MinWork is an echo -- the program repeating the
//     keystroke, a menu redrawing -- and says nothing.
//
// What it cannot know, and does not pretend to: a QUESTION. A program waiting
// for an answer and a program that has finished look the same from the
// outside, so the ask cue never rings from here; the done cue is the "prompt
// back" nudge the brief asked for. Nor errors, subagents, todos or context.
// Those are the hooks' to say, and the moment a hosted agent's hooks announce
// a session the launcher drops this in favour of them.
//
// Encoded in COUNT and POSITION, per the rule: the events are a count of
// seconds of output, never a rate read off the byte stream.
package watch

import (
	"strings"
	"sync"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// Options tune the synthesiser. The zero value takes the defaults below.
type Options struct {
	// Quiet is how long the output has to stay silent after a burst before
	// the burst counts as finished. Four seconds: a slow API call inside a
	// spinner-less program pauses for one or two; a finished program stays
	// quiet for good.
	Quiet time.Duration
	// MinWork is the shortest burst that counts as work. Anything shorter
	// is an echo of a keystroke or a redraw.
	MinWork time.Duration
	// Pulse is how much continuous output makes one tool event.
	Pulse time.Duration
}

const (
	DefaultQuiet   = 4 * time.Second
	DefaultMinWork = 1500 * time.Millisecond
	DefaultPulse   = time.Second
	// tailBytes is how much recent output is kept for the sand line.
	tailBytes = 4096
	// lineRunes caps the sand line, which fitLine then shortens further.
	lineRunes = 48
)

// Synth is safe to feed from the host's read goroutines and to step from the
// paint loop.
type Synth struct {
	opt Options

	mu       sync.Mutex
	outBytes int64  // since the last Step
	enter    bool   // Enter was typed since the last Step
	recent   []byte // the last tailBytes of output, for the sand

	started    bool
	working    bool
	burstStart time.Time
	lastOut    time.Time
	pulseAt    time.Time
}

// New makes a synthesiser with the options' zero values filled in.
func New(opt Options) *Synth {
	if opt.Quiet <= 0 {
		opt.Quiet = DefaultQuiet
	}
	if opt.MinWork <= 0 {
		opt.MinWork = DefaultMinWork
	}
	if opt.Pulse <= 0 {
		opt.Pulse = DefaultPulse
	}
	return &Synth{opt: opt}
}

// Output records n bytes the program wrote. b is the bytes themselves, kept
// only for the sand line; n may exceed len(b) if the caller trimmed.
func (s *Synth) Output(b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outBytes += int64(len(b))
	s.recent = append(s.recent, b...)
	if len(s.recent) > tailBytes {
		s.recent = append(s.recent[:0:0], s.recent[len(s.recent)-tailBytes:]...)
	}
}

// Input records bytes the user typed. Only Enter matters.
func (s *Synth) Input(b []byte) {
	if strings.IndexAny(string(b), "\r\n") < 0 {
		return
	}
	s.mu.Lock()
	s.enter = true
	s.mu.Unlock()
}

// Step folds what arrived since the last call into events. Non-blocking, and
// cheap enough to call every frame.
func (s *Synth) Step(now time.Time) []event.Event {
	s.mu.Lock()
	out, enter := s.outBytes, s.enter
	s.outBytes, s.enter = 0, false
	var recent []byte
	if out > 0 || s.working {
		recent = append([]byte(nil), s.recent...)
	}
	s.mu.Unlock()

	var ev []event.Event
	if !s.started {
		s.started = true
		ev = append(ev, event.Event{Kind: event.SessionStart, Text: "startup"})
	}
	if enter {
		ev = append(ev, event.Event{Kind: event.Prompt})
	}
	if out > 0 {
		if !s.working {
			s.working = true
			s.burstStart = now
			s.pulseAt = now
		}
		s.lastOut = now
		// One tool event per Pulse of continuous output: a COUNT of seconds
		// worked, never a byte rate.
		for !now.Before(s.pulseAt) {
			ev = append(ev, event.Event{Kind: event.ToolStart, Tool: "said"})
			s.pulseAt = s.pulseAt.Add(s.opt.Pulse)
		}
		return ev
	}
	if s.working && now.Sub(s.lastOut) >= s.opt.Quiet {
		s.working = false
		if worked := s.lastOut.Sub(s.burstStart); worked >= s.opt.MinWork {
			ev = append(ev,
				event.Event{Kind: event.ToolEnd, Tool: "said", Target: LastLine(recent), MS: worked.Milliseconds()},
				event.Event{Kind: event.Done, Text: LastLine(recent)},
			)
		}
	}
	return ev
}

// LastLine is the last non-blank line a program printed, with its escape
// sequences stripped, capped at lineRunes. A full-screen program redraws with
// cursor motion rather than newlines, so for those it is the last run of
// printable text, which is still the last thing it said.
func LastLine(b []byte) string {
	var sb strings.Builder
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch {
		case c == 0x1b:
			i = skipEscape(b, i)
		case c == '\r':
			// A carriage return without a newline overwrites the line: a
			// spinner or a progress bar. Treat it as a line break.
			sb.WriteByte('\n')
		case c == '\n' || c == '\t' || c >= 0x20:
			sb.WriteByte(c)
		}
	}
	lines := strings.Split(sb.String(), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		r := []rune(l)
		if len(r) > lineRunes {
			r = r[:lineRunes]
		}
		return string(r)
	}
	return ""
}

// skipEscape returns the index of the last byte of the escape sequence that
// starts at i, so the caller's loop steps past it.
func skipEscape(b []byte, i int) int {
	if i+1 >= len(b) {
		return len(b)
	}
	switch b[i+1] {
	case '[': // CSI: parameters and intermediates, then a final 0x40..0x7e
		j := i + 2
		for j < len(b) && !(b[j] >= 0x40 && b[j] <= 0x7e) {
			j++
		}
		return j
	case ']': // OSC: up to BEL or ESC \
		j := i + 2
		for j < len(b) {
			if b[j] == 0x07 {
				return j
			}
			if b[j] == 0x1b && j+1 < len(b) && b[j+1] == '\\' {
				return j + 1
			}
			j++
		}
		return j
	default: // ESC + one byte (ESC 7, ESC ( B, ...)
		if b[i+1] == '(' || b[i+1] == ')' {
			return i + 2
		}
		return i + 1
	}
}
