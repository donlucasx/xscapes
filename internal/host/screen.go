package host

import (
	"strconv"
	"strings"
)

// screen is the smallest terminal that can answer "what would this look like",
// and, since session 14, the host's own memory of the agent's band.
//
// It began as a test instrument: the composition bugs in this package are not
// visible in the bytes -- a stale row, a band that moved, a line that wrapped
// where it should not have. Reading the escape sequences tells you what was
// SENT; only replaying them tells you what is on the screen. It is production
// now because the host mirrors rows that leave the band into the terminal's
// own scrollback, and the only way to know which rows those are, and what was
// in them, is to keep the screen.
//
// It models what the host emits and what Claude Code was measured emitting
// (notes/claude-terminal-emissions.md): CUP, EL, ED, DECSTBM, DECOM,
// DECSC/DECRC, the relative motions, autowrap, scrolling within the region,
// SGR in full, and both screen buffers -- with DECSET 47 switching between them
// without a clear and 1049 saving, clearing and restoring, because those two
// behave differently and the mirror depends on the difference.
//
// What it does not model, knowingly: double-width glyphs advance one cell here
// and two on the terminal, so a mirrored row containing them is misaligned
// from that glyph on. That costs a history line its alignment, never the band.

// cell is a glyph and the rendition it was painted with. The background is
// load-bearing: an erase does not write spaces, it fills with the CURRENT
// background (BCE), so a row cleared while some other colour was in force
// comes out as a solid band of that colour. A model that stores only runes
// shows that row as blank and calls the host innocent.
type cell struct {
	r      rune
	fg, bg int   // colour encoding below
	attr   uint8 // attribute bits below
	host   bool  // inserted by the host (a scroll), not written by the agent
}

// Colour encoding for cell.fg and cell.bg: defaultBG for the terminal's
// default; 0..255 for an indexed colour; 1<<24 | r<<16 | g<<8 | b for
// truecolor, which keeps every truecolor value >= 1<<20 so older checks for
// "any colour at all" still hold.
const (
	defaultBG = -1
	rgbBit    = 1 << 24
)

const (
	attrBold uint8 = 1 << iota
	attrDim
	attrItalic
	attrUnderline
	attrBlink
	attrReverse
	attrHidden
	attrStrike
)

type screen struct {
	w, h int

	// The rendition in force, and the copy DECSC keeps.
	curFG, curBG int
	curAttr      uint8
	sFG, sBG     int
	sAttr        uint8

	cells      [][]cell
	x, y       int // 0-based, screen coordinates
	top, bot   int // scroll region, 0-based inclusive
	origin     bool
	wrapNext   bool
	sx, sy     int
	sTop, sBot int
	sOrigin    bool

	// The two buffers. cells is the one on display; other is the one kept
	// aside. alt says which is which. Measured in Terminal.app on 2026-09-03:
	// DECSET 47 swaps them and clears nothing, in both directions (400 round
	// trips, the alternate screen intact); 1049 saves the cursor, switches
	// and starts the alternate screen blank, and 1049l discards it.
	alt          bool
	other        [][]cell
	mainX, mainY int

	// capture, when set, keeps every row that leaves the top of the region
	// on the ALTERNATE screen by scrolling, and every row a shrink destroys,
	// so the host can mirror them. Rows the host itself inserted and never
	// wrote into are not kept: they were never the agent's.
	capture  bool
	scrolled [][]cell

	// pending holds a trailing partial escape sequence. Reads off a pipe
	// split wherever they like, and a parser that discards an incomplete
	// sequence silently drops everything after it -- which is how the first
	// version of this model showed a blank screen and blamed the host.
	pending string
}

func newScreen(w, h int) *screen {
	s := &screen{w: w, h: h, curFG: defaultBG, curBG: defaultBG, sFG: defaultBG, sBG: defaultBG}
	s.cells = blankRows(w, h)
	s.top, s.bot = 0, h-1
	return s
}

func blankRows(w, h int) [][]cell {
	rows := make([][]cell, h)
	for i := range rows {
		rows[i] = blankRow(w)
	}
	return rows
}

// blankRow is an erase to the DEFAULT background: what a fresh screen holds.
func blankRow(w int) []cell { return bgRow(w, defaultBG) }

// bgRow is an erase to a given background, which is what EL and ED actually do.
func bgRow(w, bg int) []cell {
	r := make([]cell, w)
	for i := range r {
		r[i] = cell{r: ' ', fg: defaultBG, bg: bg}
	}
	return r
}

// hostRow is a row the host scrolled into existence: blank, and marked so the
// mirror can tell it from a blank line the agent wrote.
func hostRow(w, bg int) []cell {
	r := bgRow(w, bg)
	for i := range r {
		r[i].host = true
	}
	return r
}

// keep records a row that is leaving the alternate screen for good.
func (s *screen) keep(row []cell) {
	if !s.capture || !s.alt {
		return
	}
	agent := false
	for _, c := range row {
		if !c.host {
			agent = true
			break
		}
	}
	if !agent {
		return
	}
	s.scrolled = append(s.scrolled, append([]cell(nil), row...))
}

// takeScrolled hands over the rows kept since the last call, oldest first.
func (s *screen) takeScrolled() [][]cell {
	out := s.scrolled
	s.scrolled = nil
	return out
}

// resizeScrolling is what Terminal.app does when a MAIN-screen window
// SHRINKS: it keeps the BOTTOM of the screen and lets the top go, so every
// remaining row moves up by the difference, and the cursor with it.
func (s *screen) resizeScrolling(w, h int) {
	if delta := s.h - h; delta > 0 {
		for _, row := range s.cells[:delta] {
			s.keep(row)
		}
		kept := s.cells[delta:]
		s.cells = append([][]cell{}, kept...)
		s.h = h
		for i := range s.cells {
			row := blankRow(w)
			copy(row, s.cells[i])
			s.cells[i] = row
		}
		s.w = w
		s.resizeOther(w, h)
		s.top, s.bot = 0, h-1
		s.y = clamp(s.y-delta, 0, h-1)
		s.x = clamp(s.x, 0, w-1)
		return
	}
	s.resize(w, h)
}

// resizeAnchoredBottom keeps the BOTTOM and pushes everything DOWN when the
// window grows, so blank rows appear at the top instead of at the bottom. The
// rows it inserts are the terminal's, not the agent's.
func (s *screen) resizeAnchoredBottom(w, h int) {
	if grow := h - s.h; grow > 0 {
		rows := make([][]cell, h)
		for i := range rows {
			rows[i] = hostRow(w, defaultBG)
		}
		for i := range s.cells {
			copy(rows[i+grow], s.cells[i])
		}
		s.cells, s.w, s.h = rows, w, h
		s.resizeOther(w, h)
		s.top, s.bot = 0, h-1
		s.y = clamp(s.y+grow, 0, h-1)
		return
	}
	s.resize(w, h)
}

// resizeAlt is what Terminal.app's ALTERNATE screen does, measured by eye with
// notes/contentprobe on 2026-09-03: content is anchored to the BOTTOM edge in
// both directions. A grow pushes it down by the delta and inserts blank rows
// at the TOP; a shrink pulls it up and loses the top rows.
//
// The cursor stays on its absolute row (notes/anchorprobe, and again with
// notes/shrinkprobe: 16 before, 16 after a six-row shrink), and only a shrink
// past it clamps it to the new last row (16 -> 10 on a shrink to ten rows).
// The two helpers above move it with the content, which is the MAIN screen's
// rule; on this screen that rule hid the split input box from every test in
// this package.
func (s *screen) resizeAlt(w, h int) {
	y := s.y
	if h > s.h {
		s.resizeAnchoredBottom(w, h)
	} else {
		s.resizeScrolling(w, h)
	}
	s.y = clamp(y, 0, h-1)
}

// resize keeps what fits, which is what a terminal does when it GROWS.
func (s *screen) resize(w, h int) {
	s.resizeOther(w, h)
	old := s.cells
	s.cells = blankRows(w, h)
	for i := range s.cells {
		if i < len(old) {
			copy(s.cells[i], old[i])
		}
	}
	s.w, s.h = w, h
	s.top, s.bot = 0, h-1
	if s.y >= h {
		s.y = h - 1
	}
	if s.x >= w {
		s.x = w - 1
	}
}

// resizeOther resizes the buffer kept aside. A terminal resizes both buffers,
// and a model that does not restores a stale-sized screen on 1049l -- which
// panicked the first trace analysis with "index out of range [31] with length 30".
func (s *screen) resizeOther(w, h int) {
	if s.other == nil {
		return
	}
	rows := blankRows(w, h)
	for i := range rows {
		if i < len(s.other) {
			copy(rows[i], s.other[i])
		}
	}
	s.other = rows
}

func (s *screen) rowAt(y int) string {
	r := make([]rune, len(s.cells[y]))
	for i, c := range s.cells[y] {
		r[i] = c.r
	}
	return strings.TrimRight(string(r), " ")
}

// otherRowAt is rowAt for the buffer kept aside.
func (s *screen) otherRowAt(y int) string {
	if s.other == nil || y >= len(s.other) {
		return ""
	}
	r := make([]rune, len(s.other[y]))
	for i, c := range s.other[y] {
		r[i] = c.r
	}
	return strings.TrimRight(string(r), " ")
}

// bgRunAt reports the background filling row y, and whether the WHOLE row
// carries it. A row erased under a stray colour is uniform in that colour,
// which is exactly the signature being hunted.
func (s *screen) bgRunAt(y int) (bg int, uniform bool) {
	if len(s.cells[y]) == 0 {
		return defaultBG, true
	}
	bg = s.cells[y][0].bg
	for _, c := range s.cells[y] {
		if c.bg != bg {
			return bg, false
		}
	}
	return bg, true
}

// paintedRows lists the rows in [from,to] (1-based, inclusive) left holding a
// non-default background across their whole width.
func (s *screen) paintedRows(from, to int) []int {
	var out []int
	for y := from - 1; y <= to-1 && y < s.h; y++ {
		if y < 0 {
			continue
		}
		if bg, uniform := s.bgRunAt(y); uniform && bg != defaultBG {
			out = append(out, y+1)
		}
	}
	return out
}

// scrollUp is a scroll at the bottom margin: a newline or a wrap on the last
// row of the region. The row leaving the top is the agent's, and on the
// alternate screen it is gone for good -- so it is the row the mirror keeps.
func (s *screen) scrollUp() {
	s.keep(s.cells[s.top])
	for y := s.top; y < s.bot; y++ {
		s.cells[y] = s.cells[y+1]
	}
	s.cells[s.bot] = bgRow(s.w, s.curBG)
}

func (s *screen) put(r rune) {
	if s.wrapNext {
		s.x = 0
		if s.y == s.bot {
			s.scrollUp()
		} else if s.y < s.h-1 {
			s.y++
		}
		s.wrapNext = false
	}
	if s.y >= 0 && s.y < s.h && s.x >= 0 && s.x < s.w {
		s.cells[s.y][s.x] = cell{r: r, fg: s.curFG, bg: s.curBG, attr: s.curAttr}
	}
	if s.x == s.w-1 {
		s.wrapNext = true
	} else {
		s.x++
	}
}

// move places the cursor, honouring origin mode.
func (s *screen) move(row, col int) {
	if s.origin {
		row += s.top
	}
	s.y, s.x = clamp(row, 0, s.h-1), clamp(col, 0, s.w-1)
	s.wrapNext = false
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (s *screen) feed(in string) {
	r := []rune(s.pending + in)
	s.pending = ""

	for i := 0; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '\x1b' && i+1 >= len(r):
			s.pending = string(r[i:])
			return
		case c == '\x1b' && r[i+1] == '7':
			s.sx, s.sy, s.sTop, s.sBot, s.sOrigin = s.x, s.y, s.top, s.bot, s.origin
			s.sFG, s.sBG, s.sAttr = s.curFG, s.curBG, s.curAttr
			i++
		case c == '\x1b' && r[i+1] == '8':
			s.x, s.y, s.origin = s.sx, s.sy, s.sOrigin
			s.curFG, s.curBG, s.curAttr = s.sFG, s.sBG, s.sAttr
			s.wrapNext = false
			// Measured 2026-09-03 (notes/shrinkprobe): a restore under origin
			// mode into a region that no longer contains the saved row lands
			// on row 1. That is the whole mechanism of the split input box,
			// and a model that restored the row verbatim could not see it.
			if s.origin && (s.y < s.top || s.y > s.bot) {
				s.y, s.x = s.top, 0
			}
			i++
		case c == '\x1b' && r[i+1] == '[':
			j := i + 2
			for j < len(r) && !((r[j] >= 'a' && r[j] <= 'z') || (r[j] >= 'A' && r[j] <= 'Z')) {
				j++
			}
			if j >= len(r) {
				s.pending = string(r[i:])
				return
			}
			s.csi(string(r[i+2:j]), r[j])
			i = j
		case c == '\x1b' && r[i+1] == ']':
			// OSC: a string ending in BEL or ST. Titles and hyperlinks;
			// nothing on screen. Swallowed whole, or held if it is cut.
			j := i + 2
			for j < len(r) && r[j] != '\a' && !(r[j] == '\x1b' && j+1 < len(r) && r[j+1] == '\\') {
				j++
			}
			if j >= len(r) {
				s.pending = string(r[i:])
				return
			}
			if r[j] == '\x1b' {
				j++
			}
			i = j
		case c == '\x1b':
			// ESC plus one byte that is not modelled (charset, keypad, ...).
			i++
		case c == '\n':
			if s.y == s.bot {
				s.scrollUp()
			} else if s.y < s.h-1 {
				s.y++
			}
			s.wrapNext = false
		case c == '\r':
			s.x, s.wrapNext = 0, false
		case c == '\b':
			if s.x > 0 {
				s.x--
			}
			s.wrapNext = false
		case c == '\t':
			s.x = clamp((s.x/8+1)*8, 0, s.w-1)
			s.wrapNext = false
		case c < 0x20 || c == 0x7f:
			// Other controls draw nothing.
		default:
			s.put(c)
		}
	}
}

func (s *screen) csi(params string, final rune) {
	priv := strings.HasPrefix(params, "?")
	body := strings.TrimPrefix(params, "?")
	var nums []int
	for _, p := range strings.Split(body, ";") {
		n, _ := strconv.Atoi(p)
		nums = append(nums, n)
	}
	arg := func(i, def int) int {
		if i < len(nums) && nums[i] != 0 {
			return nums[i]
		}
		if i < len(nums) && body != "" && nums[i] == 0 && strings.Contains(body, "0") {
			return 0
		}
		return def
	}
	switch {
	case priv && final == 'h' && arg(0, 0) == 6:
		s.origin = true
		s.move(0, 0)
	case priv && final == 'l' && arg(0, 0) == 6:
		s.origin = false
		s.move(0, 0)
	case priv && arg(0, 0) == 1049 && final == 'h':
		// Save the cursor, keep the main buffer aside, start the alternate
		// screen blank.
		s.mainX, s.mainY = s.x, s.y
		s.other = s.cells
		s.cells = blankRows(s.w, s.h)
		s.alt = true
		s.x, s.y = 0, 0
	case priv && arg(0, 0) == 1049 && final == 'l':
		// Back to the main buffer as it was; the alternate screen is gone.
		if s.other != nil {
			s.cells, s.other = s.other, nil
		}
		s.x, s.y = s.mainX, s.mainY
		s.alt = false
	case priv && arg(0, 0) == 47 && final == 'h':
		if !s.alt {
			if s.other == nil {
				s.other = blankRows(s.w, s.h)
			}
			s.cells, s.other = s.other, s.cells
			s.alt = true
		}
	case priv && arg(0, 0) == 47 && final == 'l':
		if s.alt {
			s.cells, s.other = s.other, s.cells
			s.alt = false
		}
	case priv:
		// Everything else private is not modelled and not needed here.
	case final == 'H' || final == 'f':
		s.move(arg(0, 1)-1, arg(1, 1)-1)
	case final == 'r':
		if body == "" {
			s.top, s.bot = 0, s.h-1
		} else {
			s.top, s.bot = clamp(arg(0, 1)-1, 0, s.h-1), clamp(arg(1, s.h)-1, 0, s.h-1)
		}
		// DECSTBM homes the cursor. This is the single fact that has cost
		// this project the most afternoons; the model has to have it or the
		// tests would bless the bug.
		s.move(0, 0)
	// The cursor motions Claude Code actually uses. The notes record it placing
	// its input "purely by RELATIVE moves from wherever the cursor is" and give
	// the exact shape: ESC[2D ESC[4B \r ESC[2C ESC[4A ... ESC[8G.
	case final == 'A':
		s.y = clamp(s.y-arg(0, 1), s.regionTop(), s.h-1)
		s.wrapNext = false
	case final == 'B':
		s.y = clamp(s.y+arg(0, 1), 0, s.regionBot())
		s.wrapNext = false
	case final == 'C':
		s.x = clamp(s.x+arg(0, 1), 0, s.w-1)
		s.wrapNext = false
	case final == 'D':
		s.x = clamp(s.x-arg(0, 1), 0, s.w-1)
		s.wrapNext = false
	case final == 'E':
		s.y, s.x = clamp(s.y+arg(0, 1), 0, s.regionBot()), 0
		s.wrapNext = false
	case final == 'F':
		s.y, s.x = clamp(s.y-arg(0, 1), s.regionTop(), s.h-1), 0
		s.wrapNext = false
	case final == 'G':
		s.x = clamp(arg(0, 1)-1, 0, s.w-1)
		s.wrapNext = false
	case final == 'd':
		s.move(arg(0, 1)-1, s.x)
	case final == 'X': // ECH: erase n cells, BCE like the others
		n := arg(0, 1)
		for i := 0; i < n && s.x+i < s.w; i++ {
			s.cells[s.y][s.x+i] = s.blank()
		}
	case final == '@': // ICH: open n cells, shifting the rest right
		n := arg(0, 1)
		row := s.cells[s.y]
		for x := s.w - 1; x >= s.x+n; x-- {
			row[x] = row[x-n]
		}
		for i := 0; i < n && s.x+i < s.w; i++ {
			row[s.x+i] = s.blank()
		}
	case final == 'P': // DCH: delete n cells, pulling the rest left
		n := arg(0, 1)
		row := s.cells[s.y]
		for x := s.x; x < s.w; x++ {
			if x+n < s.w {
				row[x] = row[x+n]
			} else {
				row[x] = s.blank()
			}
		}
	case final == 'L': // IL: insert n blank lines at the cursor, within the region
		s.scrollRegionDown(s.y, arg(0, 1))
	case final == 'M': // DL: delete n lines at the cursor, within the region
		s.scrollRegionUp(s.y, arg(0, 1))
	case final == 'S':
		s.scrollRegionUp(s.top, arg(0, 1))
	case final == 'T':
		s.scrollRegionDown(s.top, arg(0, 1))
	case final == 'm':
		s.sgr(nums, body)
	case final == 'K':
		switch arg(0, 0) {
		case 1:
			for x := 0; x <= s.x && x < s.w; x++ {
				s.cells[s.y][x] = s.blank()
			}
		case 2:
			s.cells[s.y] = bgRow(s.w, s.curBG)
		default:
			for x := s.x; x < s.w; x++ {
				s.cells[s.y][x] = s.blank()
			}
		}
	case final == 'J':
		switch arg(0, 0) {
		case 2:
			for y := 0; y < s.h; y++ {
				s.cells[y] = bgRow(s.w, s.curBG)
			}
		default:
			for x := s.x; x < s.w; x++ {
				s.cells[s.y][x] = s.blank()
			}
			for y := s.y + 1; y < s.h; y++ {
				s.cells[y] = bgRow(s.w, s.curBG)
			}
		}
	}
}

// blank is one erased cell under the current background.
func (s *screen) blank() cell { return cell{r: ' ', fg: defaultBG, bg: s.curBG} }

// sgr tracks the rendition in full, so a row can be written back out the way
// it was drawn. The background was always the one that mattered for the
// erase tests; the rest is for the mirror.
func (s *screen) sgr(nums []int, body string) {
	if body == "" {
		s.curFG, s.curBG, s.curAttr = defaultBG, defaultBG, 0
		return
	}
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			s.curFG, s.curBG, s.curAttr = defaultBG, defaultBG, 0
		case n == 1:
			s.curAttr |= attrBold
		case n == 2:
			s.curAttr |= attrDim
		case n == 3:
			s.curAttr |= attrItalic
		case n == 4:
			s.curAttr |= attrUnderline
		case n == 5:
			s.curAttr |= attrBlink
		case n == 7:
			s.curAttr |= attrReverse
		case n == 8:
			s.curAttr |= attrHidden
		case n == 9:
			s.curAttr |= attrStrike
		case n == 22:
			s.curAttr &^= attrBold | attrDim
		case n == 23:
			s.curAttr &^= attrItalic
		case n == 24:
			s.curAttr &^= attrUnderline
		case n == 25:
			s.curAttr &^= attrBlink
		case n == 27:
			s.curAttr &^= attrReverse
		case n == 28:
			s.curAttr &^= attrHidden
		case n == 29:
			s.curAttr &^= attrStrike
		case n >= 30 && n <= 37:
			s.curFG = n - 30
		case n >= 90 && n <= 97:
			s.curFG = n - 90 + 8
		case n == 39:
			s.curFG = defaultBG
		case n >= 40 && n <= 47:
			s.curBG = n - 40
		case n >= 100 && n <= 107:
			s.curBG = n - 100 + 8
		case n == 49:
			s.curBG = defaultBG
		case (n == 38 || n == 48) && i+1 < len(nums) && nums[i+1] == 5:
			v := defaultBG
			if i+2 < len(nums) {
				v = clamp(nums[i+2], 0, 255)
			}
			if n == 38 {
				s.curFG = v
			} else {
				s.curBG = v
			}
			i += 2
		case (n == 38 || n == 48) && i+1 < len(nums) && nums[i+1] == 2:
			v := defaultBG
			if i+4 < len(nums) {
				v = rgbBit | clamp(nums[i+2], 0, 255)<<16 | clamp(nums[i+3], 0, 255)<<8 | clamp(nums[i+4], 0, 255)
			}
			if n == 38 {
				s.curFG = v
			} else {
				s.curBG = v
			}
			i += 4
		}
	}
}

// regionTop and regionBot are the limits relative motion respects. Outside the
// region the whole screen is the limit, which is what a real terminal does.
func (s *screen) regionTop() int {
	if s.y >= s.top && s.y <= s.bot {
		return s.top
	}
	return 0
}

func (s *screen) regionBot() int {
	if s.y >= s.top && s.y <= s.bot {
		return s.bot
	}
	return s.h - 1
}

// scrollRegionUp deletes n lines at row `at`, pulling the rest of the region up
// and opening blanks at the bottom margin. This is SU and DL -- the host's
// undo of a grow push, never the agent's newline -- so the rows it removes are
// not kept and the rows it opens are marked as the host's.
func (s *screen) scrollRegionUp(at, n int) {
	if at < s.top || at > s.bot {
		return
	}
	for i := 0; i < n; i++ {
		for y := at; y < s.bot; y++ {
			s.cells[y] = s.cells[y+1]
		}
		s.cells[s.bot] = hostRow(s.w, s.curBG)
	}
}

// scrollRegionDown inserts n blank lines at row `at`, pushing the rest of the
// region down and losing what falls past the bottom margin.
func (s *screen) scrollRegionDown(at, n int) {
	if at < s.top || at > s.bot {
		return
	}
	for i := 0; i < n; i++ {
		for y := s.bot; y > at; y-- {
			s.cells[y] = s.cells[y-1]
		}
		s.cells[at] = hostRow(s.w, s.curBG)
	}
}

// clone is a deep copy, for snapshotting the screen mid-run.
func (s *screen) clone() *screen {
	c := *s
	c.cells = make([][]cell, len(s.cells))
	for i := range s.cells {
		c.cells[i] = append([]cell(nil), s.cells[i]...)
	}
	if s.other != nil {
		c.other = make([][]cell, len(s.other))
		for i := range s.other {
			c.other[i] = append([]cell(nil), s.other[i]...)
		}
	}
	return &c
}

// rowANSI writes a row back out as the bytes that would draw it: runs of one
// rendition under one SGR, trailing default cells dropped, a reset at the end
// so nothing leaks into whatever follows. This is what the mirror sends to the
// main buffer, so what it produces is what the user scrolls back to.
func rowANSI(row []cell) string {
	end := len(row)
	for end > 0 {
		c := row[end-1]
		if c.r != ' ' || c.bg != defaultBG {
			break
		}
		end--
	}
	if end == 0 {
		return ""
	}
	var b strings.Builder
	var fg, bg = defaultBG, defaultBG
	var attr uint8
	first := true
	for _, c := range row[:end] {
		if first || c.fg != fg || c.bg != bg || c.attr != attr {
			fg, bg, attr = c.fg, c.bg, c.attr
			writeSGR(&b, fg, bg, attr)
			first = false
		}
		b.WriteRune(c.r)
	}
	b.WriteString("\x1b[0m")
	return b.String()
}

// writeSGR emits one complete rendition, from a reset, so it does not depend
// on what was in force before it.
func writeSGR(b *strings.Builder, fg, bg int, attr uint8) {
	b.WriteString("\x1b[0")
	for _, ac := range attrCodes {
		if attr&ac.bit != 0 {
			b.WriteByte(';')
			b.WriteString(ac.code)
		}
	}
	writeColour(b, fg, 38)
	writeColour(b, bg, 48)
	b.WriteByte('m')
}

// attrCodes in a fixed order, so the same rendition always serialises to the
// same bytes and a round trip can be compared byte for byte.
var attrCodes = []struct {
	bit  uint8
	code string
}{
	{attrBold, "1"}, {attrDim, "2"}, {attrItalic, "3"}, {attrUnderline, "4"},
	{attrBlink, "5"}, {attrReverse, "7"}, {attrHidden, "8"}, {attrStrike, "9"},
}

func writeColour(b *strings.Builder, v, base int) {
	switch {
	case v == defaultBG:
	case v >= rgbBit:
		b.WriteString(";" + strconv.Itoa(base) + ";2;" +
			strconv.Itoa(v>>16&0xff) + ";" + strconv.Itoa(v>>8&0xff) + ";" + strconv.Itoa(v&0xff))
	default:
		b.WriteString(";" + strconv.Itoa(base) + ";5;" + strconv.Itoa(v))
	}
}
