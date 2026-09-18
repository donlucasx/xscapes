package host

import (
	"strings"
	"sync/atomic"
)

// Filter strips the scroll-region reset out of a hosted agent's output on its
// way to the real terminal, and keeps the agent's erase-display inside its
// band.
//
// The host splits the window: the agent gets a band anchored at row 1, held
// there by DECSTBM, and the scape paints the rows below it. Claude Code emits
// ESC[r once at startup -- measured, notes/claude-terminal-emissions.md -- and
// that one sequence resets the region to the whole screen, after which the
// agent's next scroll walks straight over the scape. So the host owns the
// scroll region, and DECSTBM in the child's stream is not passed on.
//
// This is deliberately not a VT parser. It recognises where an escape sequence
// starts and ends, which is enough to remove one of them; every other byte is
// forwarded untouched, so nothing the filter fails to understand can be
// corrupted by it.
type Filter struct {
	// Band is the agent's rows. An erase-display (ED) is not bounded by the
	// scroll region: ESC[2J clears the whole terminal, scape included, and
	// ESC[J clears from the cursor to the bottom of the screen. Kimi Code
	// CLI sends ESC[2J ESC[H ESC[3J on every step of a window drag (his
	// trace of 2026-09-18: 21 steps, 21 clears; none during work), Claude
	// Code never sends an ED, and that was the whole difference between the
	// scape blinking in and out under one and holding still under the
	// other. The filter rewrites an ED into the band's own rows, using the
	// band's bottom margin to stop it. Set by the host, followed on resize.
	Band atomic.Int32
	// Erases counts the agent's erase-displays the filter confined or
	// dropped, for the event log: the next blank row is then attributed to
	// a clear instead of inferred from a screenshot (his ruling, 2026-09-18).
	Erases atomic.Int64
	// pend holds an escape sequence that a read cut in half. A read from a
	// PTY ends wherever the kernel's buffer ended, which is regularly inside
	// a sequence: without this, ESC[ would be forwarded and the r that
	// followed in the next read would be swallowed out of the middle of a word.
	pend []byte
}

// maxPend caps how long a sequence may be held. Real CSI parameter runs are a
// handful of bytes. Holding an unterminated one forever would stall the
// agent's output on a garbled stream, and the agent's text matters more than
// the filter's tidiness.
const maxPend = 64

// Filter returns the bytes to forward to the terminal.
func (f *Filter) Filter(in []byte) []byte {
	buf := in
	if len(f.pend) > 0 {
		buf = append(f.pend, in...)
		f.pend = nil
	}
	out := make([]byte, 0, len(buf))

	for i := 0; i < len(buf); {
		if buf[i] != esc {
			out = append(out, buf[i])
			i++
			continue
		}
		end, final, complete := csiEnd(buf, i)
		if !complete {
			// Incomplete, and short enough to still be a real sequence: hold
			// it for the next read.
			if len(buf)-i <= maxPend {
				f.pend = append([]byte(nil), buf[i:]...)
				return out
			}
			// Too long to be anything. Let it through as ordinary bytes.
			out = append(out, buf[i])
			i++
			continue
		}
		switch final {
		case 'r': // DECSTBM, the one the host keeps for itself
		case 'J':
			out = append(out, f.confineErase(buf[i:end+1])...)
		default:
			out = append(out, buf[i:end+1]...)
		}
		i = end + 1
	}
	return out
}

// confineErase rewrites an agent's ED so it reaches only the band. The
// rewrite runs in the agent's own stream, under the band's region and origin
// mode, where CUD stops at the bottom margin; the cursor is saved and put
// back around it, as ED leaves the cursor where it was.
//
//	ESC[J, ESC[0J  from the cursor down: the rest of this line, then every
//	               band row below it
//	ESC[2J         every band row
//	ESC[1J         above the cursor: already the band's own rows, passed
//	ESC[3J         the scrollback: DROPPED. The terminal's scrollback is the
//	               host's, where the mirror writes the agent's rows as they
//	               leave the band (Host.History); an agent clearing it on
//	               every resize would erase that record.
func (f *Filter) confineErase(seq []byte) []byte {
	band := int(f.Band.Load())
	p := strings.TrimPrefix(string(seq[2:len(seq)-1]), "?")
	if band <= 0 && p != "3" {
		return seq
	}
	f.Erases.Add(1)
	var b strings.Builder
	switch p {
	case "", "0":
		b.WriteString("\x1b[K\x1b7")
		for k := 0; k < band; k++ {
			b.WriteString("\x1b[1B\x1b[2K")
		}
		b.WriteString("\x1b8")
	case "2":
		b.WriteString("\x1b7\x1b[H")
		for k := 0; k < band; k++ {
			b.WriteString("\x1b[2K\x1b[1B")
		}
		b.WriteString("\x1b8")
	case "3":
		return nil
	default:
		return seq
	}
	return []byte(b.String())
}

// Flush releases anything held back, for the end of the stream.
func (f *Filter) Flush() []byte {
	out := f.pend
	f.pend = nil
	return out
}

const esc = 0x1b

// csiEnd finds the last byte of the escape sequence starting at i. It reports
// the sequence as incomplete only when the buffer genuinely runs out mid-way.
//
// Anything that is not a CSI -- ESC 7, ESC 8, ESC M, an OSC string -- is
// treated as ESC plus one byte and forwarded. The rest of an OSC string then
// travels as ordinary bytes, which is safe: the filter only ever removes a
// complete CSI, so text inside a string cannot be mistaken for one.
func csiEnd(buf []byte, i int) (end int, final byte, complete bool) {
	if i+1 >= len(buf) {
		return 0, 0, false
	}
	if buf[i+1] != '[' {
		return i + 1, buf[i+1], true
	}
	k := i + 2
	for k < len(buf) && isCSIParam(buf[k]) {
		k++
	}
	if k >= len(buf) {
		return 0, 0, false
	}
	return k, buf[k], true
}

// isCSIParam covers the parameter bytes (0x30-0x3F: digits, ; ? : < = >) and
// the intermediates (0x20-0x2F) that may precede a CSI's final byte.
func isCSIParam(b byte) bool { return b >= 0x20 && b <= 0x3f }
