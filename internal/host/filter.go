package host

// Filter strips the scroll-region reset out of a hosted agent's output on its
// way to the real terminal.
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
		if final != 'r' { // 'r' is DECSTBM, the one the host keeps for itself
			out = append(out, buf[i:end+1]...)
		}
		i = end + 1
	}
	return out
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
