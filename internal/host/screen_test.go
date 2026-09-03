package host

import (
	"strconv"
	"strings"
)

// screen is the smallest terminal that can answer "what would this look like".
//
// It exists because the composition bugs in this package are not visible in the
// bytes: a stale row, a band that moved, a line that wrapped where it should
// not have. Reading the escape sequences tells you what was SENT. Only replaying
// them tells you what is on the screen, and that is the thing being claimed.
//
// It models exactly what the host emits and Claude Code was measured emitting:
// CUP, EL, ED, DECSTBM, DECOM, DECSC/DECRC, autowrap and scrolling within the
// region. SGR is parsed and discarded -- colour is not what these tests are
// about, and keeping it would make every expectation unreadable.
type screen struct {
	w, h       int
	cells      [][]rune
	x, y       int // 0-based, screen coordinates
	top, bot   int // scroll region, 0-based inclusive
	origin     bool
	wrapNext   bool
	sx, sy     int
	sTop, sBot int
	sOrigin    bool

	// pending holds a trailing partial escape sequence. Reads off a pipe split
	// wherever they like, and a parser that discards an incomplete sequence
	// silently drops everything after it -- which is how the first version of
	// this model showed a blank screen and blamed the host.
	pending string
}

func newScreen(w, h int) *screen {
	s := &screen{w: w, h: h}
	s.cells = make([][]rune, h)
	for i := range s.cells {
		s.cells[i] = blankRow(w)
	}
	s.top, s.bot = 0, h-1
	return s
}

func blankRow(w int) []rune {
	r := make([]rune, w)
	for i := range r {
		r[i] = ' '
	}
	return r
}

// resizeScrolling is what Terminal.app does when a window SHRINKS: it keeps the
// BOTTOM of the screen and lets the top go, so every remaining row moves up by
// the difference.
//
// That is the case the host was not defending against, and it is what he
// photographed. The scape is painted at the bottom, so when the window shrinks
// the scape's own rows slide UP into the agent's band -- and neither side
// cleans them up. The host cleared only the rows that changed hands assuming
// nothing moved, and Claude Code emits nothing at all on a resize, so a strip
// of old sky sits above the band until something else happens to overwrite it.
func (s *screen) resizeScrolling(w, h int) {
	if delta := s.h - h; delta > 0 {
		kept := s.cells[delta:]
		s.cells = append([][]rune{}, kept...)
		s.h = h
		for i := range s.cells {
			row := blankRow(w)
			copy(row, s.cells[i])
			s.cells[i] = row
		}
		s.w = w
		s.top, s.bot = 0, h-1
		s.y = clamp(s.y-delta, 0, h-1)
		s.x = clamp(s.x, 0, w-1)
		return
	}
	s.resize(w, h)
}

// resizeAnchoredBottom is the other thing a terminal can do on a resize: keep
// the BOTTOM and push everything DOWN when the window grows, so blank rows
// appear at the top instead of at the bottom.
//
// It is here to mark the boundary of what the host can fix. The scape repaints
// itself either way, so its rows are always right. The AGENT's screen is a
// different matter: Claude Code emits nothing at all on a resize, so wherever
// the terminal puts its transcript is where it stays until the next keystroke.
// No amount of clearing helps, because the host cannot redraw a UI it does not
// model. That is the cost of not being a terminal emulator, and it was decided
// with eyes open.
func (s *screen) resizeAnchoredBottom(w, h int) {
	if grow := h - s.h; grow > 0 {
		rows := make([][]rune, h)
		for i := range rows {
			rows[i] = blankRow(w)
		}
		for i := range s.cells {
			copy(rows[i+grow], s.cells[i])
		}
		s.cells, s.w, s.h = rows, w, h
		s.top, s.bot = 0, h-1
		s.y = clamp(s.y+grow, 0, h-1)
		return
	}
	s.resize(w, h)
}

// resize keeps what fits, which is what a terminal does when it GROWS.
func (s *screen) resize(w, h int) {
	old := s.cells
	s.cells = make([][]rune, h)
	for i := range s.cells {
		s.cells[i] = blankRow(w)
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

func (s *screen) rowAt(y int) string { return strings.TrimRight(string(s.cells[y]), " ") }

func (s *screen) scrollUp() {
	for y := s.top; y < s.bot; y++ {
		copy(s.cells[y], s.cells[y+1])
	}
	s.cells[s.bot] = blankRow(s.w)
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
		s.cells[s.y][s.x] = r
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
		case c == '\x1b' && i+1 < len(r) && r[i+1] == '7':
			s.sx, s.sy, s.sTop, s.sBot, s.sOrigin = s.x, s.y, s.top, s.bot, s.origin
			i++
		case c == '\x1b' && i+1 < len(r) && r[i+1] == '8':
			s.x, s.y, s.origin = s.sx, s.sy, s.sOrigin
			s.wrapNext = false
			i++
		case c == '\x1b' && i+1 < len(r) && r[i+1] == '[':
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
		case c == '\n':
			if s.y == s.bot {
				s.scrollUp()
			} else if s.y < s.h-1 {
				s.y++
			}
			s.wrapNext = false
		case c == '\r':
			s.x, s.wrapNext = 0, false
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
	case priv:
		// 1049 and friends: not modelled, and not needed -- these tests run
		// the host with AltScreen off.
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
	case final == 'K':
		switch arg(0, 0) {
		case 1:
			for x := 0; x <= s.x && x < s.w; x++ {
				s.cells[s.y][x] = ' '
			}
		case 2:
			s.cells[s.y] = blankRow(s.w)
		default:
			for x := s.x; x < s.w; x++ {
				s.cells[s.y][x] = ' '
			}
		}
	case final == 'J':
		switch arg(0, 0) {
		case 2:
			for y := 0; y < s.h; y++ {
				s.cells[y] = blankRow(s.w)
			}
		default:
			for x := s.x; x < s.w; x++ {
				s.cells[s.y][x] = ' '
			}
			for y := s.y + 1; y < s.h; y++ {
				s.cells[y] = blankRow(s.w)
			}
		}
	}
}

// clone is a deep copy, for snapshotting the screen mid-run.
func (s *screen) clone() *screen {
	c := *s
	c.cells = make([][]rune, len(s.cells))
	for i := range s.cells {
		c.cells[i] = append([]rune(nil), s.cells[i]...)
	}
	return &c
}
