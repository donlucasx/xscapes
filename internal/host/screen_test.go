package host

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
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
// cell is a glyph and the background it was painted on. The background is the
// whole point: an erase does not write spaces, it fills with the CURRENT
// background (BCE), so a row cleared while some other colour was in force comes
// out as a solid band of that colour. A model that stores only runes shows that
// row as blank and calls the host innocent.
type cell struct {
	r  rune
	bg int // -1 is the terminal's default
}

const defaultBG = -1

type screen struct {
	w, h       int
	curBG      int
	sBG        int
	cells      [][]cell
	x, y       int // 0-based, screen coordinates
	top, bot   int // scroll region, 0-based inclusive
	origin     bool
	wrapNext   bool
	sx, sy     int
	sTop, sBot int
	sOrigin    bool

	// alt and mainCells model the alternate screen. It is not decoration: the
	// two buffers behave DIFFERENTLY on a resize, measured in Terminal.app on
	// 2026-09-03 with the cursor parked mid-screen and read back with DSR --
	// the main screen keeps the BOTTOM and slides content up by the rows lost,
	// the alternate screen keeps the TOP and truncates. The host shipped one
	// rule for both.
	alt          bool
	mainCells    [][]cell
	mainX, mainY int

	// pending holds a trailing partial escape sequence. Reads off a pipe split
	// wherever they like, and a parser that discards an incomplete sequence
	// silently drops everything after it -- which is how the first version of
	// this model showed a blank screen and blamed the host.
	pending string
}

func newScreen(w, h int) *screen {
	s := &screen{w: w, h: h, curBG: defaultBG, sBG: defaultBG}
	s.cells = make([][]cell, h)
	for i := range s.cells {
		s.cells[i] = blankRow(w)
	}
	s.top, s.bot = 0, h-1
	return s
}

// blankRow is an erase to the DEFAULT background: what a fresh screen holds.
func blankRow(w int) []cell { return bgRow(w, defaultBG) }

// bgRow is an erase to a given background, which is what EL and ED actually do.
func bgRow(w, bg int) []cell {
	r := make([]cell, w)
	for i := range r {
		r[i] = cell{' ', bg}
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
		s.cells = append([][]cell{}, kept...)
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
		rows := make([][]cell, h)
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
//
// It resizes the SAVED main buffer too. A terminal resizes both buffers, and a
// model that does not restores a stale-sized screen on 1049l -- which panicked
// the first trace analysis with "index out of range [31] with length 30".
func (s *screen) resize(w, h int) {
	if s.mainCells != nil {
		rows := make([][]cell, h)
		for i := range rows {
			rows[i] = blankRow(w)
			if i < len(s.mainCells) {
				copy(rows[i], s.mainCells[i])
			}
		}
		s.mainCells = rows
	}
	old := s.cells
	s.cells = make([][]cell, h)
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

func (s *screen) rowAt(y int) string {
	r := make([]rune, len(s.cells[y]))
	for i, c := range s.cells[y] {
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
		s.cells[s.y][s.x] = cell{r, s.curBG}
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
			s.sx, s.sy, s.sTop, s.sBot, s.sOrigin, s.sBG = s.x, s.y, s.top, s.bot, s.origin, s.curBG
			i++
		case c == '\x1b' && i+1 < len(r) && r[i+1] == '8':
			s.x, s.y, s.origin, s.curBG = s.sx, s.sy, s.sOrigin, s.sBG
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
	case priv && (arg(0, 0) == 1049 || arg(0, 0) == 47) && final == 'h':
		// Take the alternate screen: a blank buffer, the main one kept aside.
		s.mainCells, s.mainX, s.mainY = s.cells, s.x, s.y
		s.alt = true
		s.cells = make([][]cell, s.h)
		for i := range s.cells {
			s.cells[i] = blankRow(s.w)
		}
		s.x, s.y = 0, 0
	case priv && (arg(0, 0) == 1049 || arg(0, 0) == 47) && final == 'l':
		if s.mainCells != nil {
			s.cells, s.x, s.y = s.mainCells, s.mainX, s.mainY
			s.mainCells = nil
		}
		s.alt = false
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
	case final == 'm':
		s.sgr(nums, body)
	case final == 'K':
		switch arg(0, 0) {
		case 1:
			for x := 0; x <= s.x && x < s.w; x++ {
				s.cells[s.y][x] = cell{' ', s.curBG}
			}
		case 2:
			s.cells[s.y] = bgRow(s.w, s.curBG)
		default:
			for x := s.x; x < s.w; x++ {
				s.cells[s.y][x] = cell{' ', s.curBG}
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
				s.cells[s.y][x] = cell{' ', s.curBG}
			}
			for y := s.y + 1; y < s.h; y++ {
				s.cells[y] = bgRow(s.w, s.curBG)
			}
		}
	}
}

// sgr tracks the background colour, and nothing else.
//
// Background only, deliberately: an erase fills with the background, so that is
// the one attribute that can turn a cleared row into a visible band. Tracking
// the foreground too would double the model's size and answer no question these
// tests ask.
func (s *screen) sgr(nums []int, body string) {
	if body == "" {
		s.curBG = defaultBG
		return
	}
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			s.curBG = defaultBG
		case n >= 40 && n <= 47:
			s.curBG = n - 40
		case n >= 100 && n <= 107:
			s.curBG = n - 100 + 8
		case n == 49:
			s.curBG = defaultBG
		case n == 48 && i+1 < len(nums) && nums[i+1] == 5:
			if i+2 < len(nums) {
				s.curBG = nums[i+2]
			}
			i += 2
		case n == 48 && i+1 < len(nums) && nums[i+1] == 2:
			// Truecolor: collapsed to a single sentinel. These tests ask
			// "default or not", never "which shade".
			s.curBG = 1 << 20
			i += 4
		}
	}
}

// clone is a deep copy, for snapshotting the screen mid-run.
func (s *screen) clone() *screen {
	c := *s
	c.cells = make([][]cell, len(s.cells))
	for i := range s.cells {
		c.cells[i] = append([]cell(nil), s.cells[i]...)
	}
	return &c
}

// TestReplayTrace reconstructs a real session's screen from a trace file.
//
// Not a test of anything: an instrument, skipped unless pointed at a trace by
// XSCAPES_TRACE. A screenshot shows what a screen LOOKED like; it cannot say
// whether the terminal moved the content or the host erased it. Replaying the
// exact bytes the host sent, through the same model the resize tests use,
// answers that.
//
//	XSCAPES_TRACE=/tmp/t.bin xscapes claude     # in a real window, then resize
//	XSCAPES_TRACE=/tmp/t.bin go test ./internal/host -run TestReplayTrace -v
//
// TRACE_SIZE=WxH gives the window size to model; TRACE_ALT=0 replays on the
// main screen. Both resize directions on the alternate screen are ANCHORED TOP
// (measured 2026-09-03, notes/anchorprobe), so a size change is applied here as
// "keep what fits", which is what the model's resize does.
func TestReplayTrace(t *testing.T) {
	path := os.Getenv("XSCAPES_TRACE")
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace file to replay it")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The sidecar says which size was in force from which byte offset, so the
	// replay resizes exactly where the real terminal did. Feeding the whole
	// stream into one fixed-size screen would misreconstruct everything
	// written before the resize, which is the part under investigation.
	marks := readTraceMarks(t, path)
	w, h := 120, 59
	if s := os.Getenv("TRACE_SIZE"); s != "" {
		if _, err := fmt.Sscanf(s, "%dx%d", &w, &h); err != nil {
			t.Fatalf("TRACE_SIZE=%q: want WxH", s)
		}
	}
	if len(marks) > 0 {
		w, h = marks[0].cols, marks[0].rows
	}
	sc := newScreen(w, h)
	fed := 0
	for i, m := range marks {
		if i > 0 {
			sc.feed(string(b[fed:min(m.off, len(b))]))
			fed = min(m.off, len(b))
			// Measured 2026-09-03: the alternate screen is anchored-top in
			// BOTH directions, which is what resize does.
			sc.resize(m.cols, m.rows)
			t.Logf("-- resize at byte %d -> %dx%d (band 1..%d)", m.off, m.cols, m.rows, m.agent)
		}
	}
	sc.feed(string(b[fed:]))

	t.Logf("replayed %d bytes at %dx%d; scroll region rows %d..%d, cursor r%d c%d, alt=%v",
		len(b), w, h, sc.top+1, sc.bot+1, sc.y+1, sc.x+1, sc.alt)
	for y := 0; y < sc.h; y++ {
		row := sc.rowAt(y)
		bg, uniform := sc.bgRunAt(y)
		mark := "  "
		switch {
		case uniform && bg != defaultBG:
			mark = "##" // erased under a colour: a solid band on the real screen
		case row == "":
			mark = ".." // genuinely empty
		}
		if len(row) > 100 {
			row = row[:100] + "…"
		}
		t.Logf("%s row %2d | %s", mark, y+1, row)
	}
}

// traceMark is one size change, at the byte offset it took effect.
type traceMark struct{ off, cols, rows, agent int }

func readTraceMarks(t *testing.T, path string) []traceMark {
	t.Helper()
	var marks []traceMark
	lb, err := os.ReadFile(path + ".log")
	if err != nil {
		return nil
	}
	for _, ln := range strings.Split(strings.TrimSpace(string(lb)), "\n") {
		var m traceMark
		if _, err := fmt.Sscanf(ln, "%d %d %d %d", &m.off, &m.cols, &m.rows, &m.agent); err == nil {
			marks = append(marks, m)
		}
	}
	return marks
}

// TestTraceRightEdge asks one question of a real trace: was every scape row
// painted all the way to the last column the host believed in?
//
// He photographed a strip of stale scape down the far right. That is what a row
// painted narrower than the terminal leaves behind -- the columns past the end
// are never written, so they keep whatever the previous frame, or the previous
// SIZE, put there. This separates the two causes: rows short of the host's own
// `cols` mean the RENDERER is wrong; rows full to `cols` mean the host's idea
// of `cols` was behind the window.
func TestTraceRightEdge(t *testing.T) {
	path := os.Getenv("XSCAPES_TRACE")
	if path == "" {
		t.Skip("set XSCAPES_TRACE to a trace file")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marks := readTraceMarks(t, path)
	if len(marks) == 0 {
		t.Skip("no size sidecar next to the trace")
	}
	sc := newScreen(marks[0].cols, marks[0].rows)
	fed := 0
	for i, m := range marks {
		if i == 0 {
			continue
		}
		sc.feed(string(b[fed:min(m.off, len(b))]))
		fed = min(m.off, len(b))
		sc.resize(m.cols, m.rows)
	}
	// Stop before the host hands the terminal back: Close emits ESC[?1049l,
	// which swaps away the very buffer under investigation.
	tail := string(b[fed:])
	if i := strings.LastIndex(tail, "\x1b[?1049l"); i >= 0 {
		tail = tail[:i]
	}
	sc.feed(tail)

	last := marks[len(marks)-1]
	t.Logf("final geometry the HOST believed: %dx%d, band 1..%d", last.cols, last.rows, last.agent)
	short := 0
	for y := last.agent; y < sc.h && y < len(sc.cells); y++ {
		row := sc.cells[y]
		end := -1
		for x := len(row) - 1; x >= 0; x-- {
			if row[x].r != ' ' || row[x].bg != defaultBG {
				end = x
				break
			}
		}
		if end != sc.w-1 {
			short++
			t.Errorf("scape row %d painted only through column %d of %d", y+1, end+1, sc.w)
		}
	}
	if short == 0 {
		t.Logf("every scape row is painted through column %d of %d -- the renderer is not the cause;"+
			" any strip beyond column %d means the window was wider than the host knew",
			sc.w, sc.w, sc.w)
	}
}
