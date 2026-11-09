//go:build ghosttyvt

// Package gvt is a terminal emulator in a box: libghostty-vt (Ghostty's own
// VT engine) behind a small Go API.
//
// The host feeds the agent's bytes in, and draws the agent's rows out of it
// onto the real terminal, the way tmux does. Because the agent never writes
// to the real terminal directly, the real terminal's quirks on a resize (an
// alternate screen that crops instead of reflowing; Terminal.app and Ghostty
// each moving content their own way) stop mattering: on a resize the engine
// REFLOWS the agent's text, Ghostty's main-screen reflow, and the host
// repaints every row.
//
// Build with -tags ghosttyvt. The static library lives in
// third_party/libghostty-vt (built from ghostty tag vt-1.0.0 with
// `zig build -Demit-lib-vt -Doptimize=ReleaseFast`).
package gvt

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/libghostty-vt/include -DGHOSTTY_STATIC
#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/../../third_party/libghostty-vt/lib/darwin-arm64/libghostty-vt.a
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ghostty/vt.h>

typedef struct {
	GhosttyTerminal t;
	GhosttyRenderState rs;
	GhosttyRenderStateRowIterator it;
	GhosttyRenderStateRowCells cells;
	// Replies the engine wants sent back to the agent (cursor reports,
	// device attributes, ...), drained by Go after every write.
	uint8_t *reply; size_t rlen, rcap;
	// The frame being built.
	char *out; size_t olen, ocap;
	// One row's cells while it is drawn, sized to the widest row seen; per
	// terminal, so two terminals never share it (Kimi, 2026-11-09).
	struct cellv *row; int rowcap;
} gvt;

static void put(gvt *g, const char *s, size_t n) {
	if (g->olen + n > g->ocap) {
		size_t c = g->ocap ? g->ocap * 2 : 65536;
		while (c < g->olen + n) c *= 2;
		g->out = realloc(g->out, c);
		g->ocap = c;
	}
	memcpy(g->out + g->olen, s, n);
	g->olen += n;
}
static void puts_(gvt *g, const char *s) { put(g, s, strlen(s)); }

static void on_write_pty(GhosttyTerminal t, void *ud, const uint8_t *b, size_t n) {
	gvt *g = ud;
	if (g->rlen + n > g->rcap) {
		size_t c = g->rcap ? g->rcap * 2 : 256;
		while (c < g->rlen + n) c *= 2;
		g->reply = realloc(g->reply, c);
		g->rcap = c;
	}
	memcpy(g->reply + g->rlen, b, n);
	g->rlen += n;
}

static gvt *gvt_new(uint16_t cols, uint16_t rows, size_t scrollback) {
	gvt *g = calloc(1, sizeof *g);
	if (ghostty_terminal_new(NULL, &g->t, cols, rows) != GHOSTTY_SUCCESS) { free(g); return NULL; }
	ghostty_terminal_set(g->t, GHOSTTY_TERMINAL_OPT_USERDATA, g);
	ghostty_terminal_set(g->t, GHOSTTY_TERMINAL_OPT_WRITE_PTY, (const void *)on_write_pty);
	ghostty_terminal_set(g->t, GHOSTTY_TERMINAL_OPT_SCROLLBACK_MAX_BYTES, &scrollback);
	ghostty_render_state_new(NULL, &g->rs);
	ghostty_render_state_row_iterator_new(NULL, &g->it);
	ghostty_render_state_row_cells_new(NULL, &g->cells);
	return g;
}

static void gvt_free(gvt *g) {
	if (!g) return;
	ghostty_render_state_row_cells_free(g->cells);
	ghostty_render_state_row_iterator_free(g->it);
	ghostty_render_state_free(g->rs);
	ghostty_terminal_free(g->t);
	free(g->reply); free(g->out); free(g->row); free(g);
}

static void gvt_write(gvt *g, const uint8_t *b, size_t n) { ghostty_terminal_vt_write(g->t, b, n); }
static int gvt_resize(gvt *g, uint16_t cols, uint16_t rows) {
	return ghostty_terminal_resize(g->t, cols, rows, 8, 16) == GHOSTTY_SUCCESS;
}

static bool gvt_mode(gvt *g, uint16_t mode, bool ansi) {
	GhosttyTerminalModeConfig m = { .mode = ghostty_mode_new(mode, ansi), .value = false };
	ghostty_terminal_get(g->t, GHOSTTY_TERMINAL_DATA_MODE, &m);
	return m.value;
}
static uint8_t gvt_kitty(gvt *g) {
	uint8_t f = 0;
	ghostty_terminal_get(g->t, GHOSTTY_TERMINAL_DATA_KITTY_KEYBOARD_FLAGS, &f);
	return f;
}

static void color(gvt *g, char *b, int base, GhosttyStyleColor c) {
	if (c.tag == GHOSTTY_STYLE_COLOR_PALETTE) {
		snprintf(b, 32, ";%d;5;%u", base, c.value.palette);
	} else if (c.tag == GHOSTTY_STYLE_COLOR_RGB) {
		snprintf(b, 32, ";%d;2;%u;%u;%u", base, c.value.rgb.r, c.value.rgb.g, c.value.rgb.b);
	} else { b[0] = 0; return; }
	puts_(g, b);
}

// sgr writes a full SGR for st, starting from a reset.
static void sgr(gvt *g, GhosttyStyle st) {
	char b[48];
	puts_(g, "\x1b[0");
	uint16_t f = st.flags;
	if (f & GHOSTTY_STYLE_FLAG_BOLD) puts_(g, ";1");
	if (f & GHOSTTY_STYLE_FLAG_FAINT) puts_(g, ";2");
	if (f & GHOSTTY_STYLE_FLAG_ITALIC) puts_(g, ";3");
	int ul = GHOSTTY_STYLE_UNDERLINE(st);
	if (ul == 1) puts_(g, ";4");
	else if (ul > 1) { snprintf(b, sizeof b, ";4:%d", ul); puts_(g, b); }
	if (f & GHOSTTY_STYLE_FLAG_BLINK) puts_(g, ";5");
	if (f & GHOSTTY_STYLE_FLAG_INVERSE) puts_(g, ";7");
	if (f & GHOSTTY_STYLE_FLAG_INVISIBLE) puts_(g, ";8");
	if (f & GHOSTTY_STYLE_FLAG_STRIKETHROUGH) puts_(g, ";9");
	if (f & GHOSTTY_STYLE_FLAG_OVERLINE) puts_(g, ";53");
	color(g, b, 38, st.fg_color);
	color(g, b, 48, st.bg_color);
	color(g, b, 58, st.underline_color);
	puts_(g, "m");
}

static int same(GhosttyStyle a, GhosttyStyle b) {
	return memcmp(&a.fg_color, &b.fg_color, sizeof a.fg_color) == 0 &&
	       memcmp(&a.bg_color, &b.bg_color, sizeof a.bg_color) == 0 &&
	       memcmp(&a.underline_color, &b.underline_color, sizeof a.underline_color) == 0 &&
	       a.flags == b.flags;
}

// One row's cells, captured before anything is written, so trailing blanks
// can be cut and replaced by one erase.
typedef struct cellv { GhosttyStyle st; uint8_t txt[64]; uint8_t n; uint8_t wide; } cellv;

// gvt_render draws rows that changed since the last render (every row when
// full), each placed at screen row top+y, and returns the bytes in g->out.
// Rows past maxRows are not drawn.
static void gvt_render(gvt *g, int full, int top, int maxRows, int cols) {
	g->olen = 0;
	if (ghostty_render_state_update(g->rs, g->t) != GHOSTTY_SUCCESS) return;
	if (ghostty_render_state_get(g->rs, GHOSTTY_RENDER_STATE_DATA_ROW_ITERATOR, &g->it) != GHOSTTY_SUCCESS) return;
	if (cols > g->rowcap) {
		cellv *r = realloc(g->row, (size_t)cols * sizeof(cellv));
		if (!r) return;
		g->row = r;
		g->rowcap = cols;
	}
	cellv *row = g->row;
	uint16_t y = 0;
	for (;;) {
		bool more;
		if (full) { more = ghostty_render_state_row_iterator_next(g->it); }
		else { more = ghostty_render_state_row_iterator_next_dirty(g->it, &y); }
		if (!more) break;
		int vy = y;
		if (full) { vy = y; y++; }
		if (vy >= maxRows) continue;
		if (ghostty_render_state_row_get(g->it, GHOSTTY_RENDER_STATE_ROW_DATA_CELLS, &g->cells) != GHOSTTY_SUCCESS) continue;
		int n = 0, last = -1;
		while (n < cols && ghostty_render_state_row_cells_next(g->cells)) {
			cellv *c = &row[n];
			memset(c, 0, sizeof *c);
			GhosttyCell raw = 0;
			ghostty_render_state_row_cells_get(g->cells, GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_RAW, &raw);
			GhosttyCellWide w = GHOSTTY_CELL_WIDE_NARROW;
			ghostty_cell_get(raw, GHOSTTY_CELL_DATA_WIDE, &w);
			c->wide = (uint8_t)w;
			ghostty_render_state_row_cells_get(g->cells, GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_STYLE, &c->st);
			GhosttyBuffer buf = { .ptr = c->txt, .cap = sizeof c->txt, .len = 0 };
			GhosttyResult gr = ghostty_render_state_row_cells_get(g->cells, GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_UTF8, &buf);
			if (gr == GHOSTTY_SUCCESS) {
				c->n = (uint8_t)buf.len;
			} else if (gr == GHOSTTY_OUT_OF_SPACE) {
				// A cluster longer than the cell buffer: draw its base
				// codepoint rather than a hole.
				uint32_t cp[1]; GhosttyCodepointBuffer cb = { .ptr = cp, .cap = 1, .len = 0 };
				ghostty_render_state_row_cells_get(g->cells, GHOSTTY_RENDER_STATE_ROW_CELLS_DATA_GRAPHEMES_BUF, &cb);
				uint32_t u = cb.len ? cp[0] : 0xFFFD;
				if (u < 0x80) { c->txt[0] = u; c->n = 1; }
				else if (u < 0x800) { c->txt[0] = 0xC0 | (u >> 6); c->txt[1] = 0x80 | (u & 0x3F); c->n = 2; }
				else if (u < 0x10000) { c->txt[0] = 0xE0 | (u >> 12); c->txt[1] = 0x80 | ((u >> 6) & 0x3F); c->txt[2] = 0x80 | (u & 0x3F); c->n = 3; }
				else { c->txt[0] = 0xF0 | (u >> 18); c->txt[1] = 0x80 | ((u >> 12) & 0x3F); c->txt[2] = 0x80 | ((u >> 6) & 0x3F); c->txt[3] = 0x80 | (u & 0x3F); c->n = 4; }
			}
			bool styled = c->st.bg_color.tag != GHOSTTY_STYLE_COLOR_NONE ||
				(c->st.flags & (GHOSTTY_STYLE_FLAG_INVERSE)) || GHOSTTY_STYLE_UNDERLINE(c->st) ||
				(c->st.flags & (GHOSTTY_STYLE_FLAG_STRIKETHROUGH | GHOSTTY_STYLE_FLAG_OVERLINE));
			if (c->n > 0 || styled || w == GHOSTTY_CELL_WIDE_SPACER_TAIL) last = n;
			n++;
		}
		char b[32];
		snprintf(b, sizeof b, "\x1b[%d;1H", top + vy + 1);
		puts_(g, b);
		GhosttyStyle cur; memset(&cur, 0, sizeof cur);
		puts_(g, "\x1b[0m");
		for (int x = 0; x <= last; x++) {
			cellv *c = &row[x];
			if (c->wide == GHOSTTY_CELL_WIDE_SPACER_TAIL) continue;
			if (!same(c->st, cur)) { sgr(g, c->st); cur = c->st; }
			if (c->wide == GHOSTTY_CELL_WIDE_SPACER_HEAD || c->n == 0) put(g, " ", 1);
			else put(g, (const char *)c->txt, c->n);
		}
		puts_(g, "\x1b[0m");
		if (last + 1 < cols) puts_(g, "\x1b[K");
	}
	ghostty_render_state_clean(g->rs);
}

// gvt_cursor reports the agent's cursor: x, y, visible (bit 0), in the
// viewport (bit 1), style in bits 8+.
static void gvt_cursor(gvt *g, int *x, int *y, int *flags) {
	GhosttyRenderStateCursor c;
	memset(&c, 0, sizeof c);
	c.size = sizeof c;
	ghostty_render_state_update(g->rs, g->t);
	ghostty_render_state_get(g->rs, GHOSTTY_RENDER_STATE_DATA_CURSOR, &c);
	*x = c.viewport_x; *y = c.viewport_y;
	*flags = (c.visible ? 1 : 0) | (c.viewport_has_value ? 2 : 0) | ((int)c.visual_style << 8) | (c.blinking ? 4 : 0);
}

// gvt_transcript formats the active screen (scrollback included) as VT with
// soft wraps joined, so the real terminal can reflow it.
static void gvt_transcript(gvt *g, int vt) {
	g->olen = 0;
	GhosttyFormatter fm;
	if (ghostty_formatter_terminal_new(NULL, &fm, g->t) != GHOSTTY_SUCCESS) return;
	GhosttyFormatterFormat emit = vt ? GHOSTTY_FORMATTER_FORMAT_VT : GHOSTTY_FORMATTER_FORMAT_PLAIN;
	bool unwrap = true;
	GhosttyFormatterExtras ex = GHOSTTY_FORMATTER_EXTRAS_NONE;
	ghostty_formatter_set(fm, GHOSTTY_FORMATTER_OPT_EMIT, &emit);
	ghostty_formatter_set(fm, GHOSTTY_FORMATTER_OPT_UNWRAP, &unwrap);
	ghostty_formatter_set(fm, GHOSTTY_FORMATTER_OPT_EXTRAS, &ex);
	uint8_t *o = NULL; size_t n = 0;
	if (ghostty_formatter_format_alloc(fm, NULL, &o, &n) == GHOSTTY_SUCCESS && o) {
		put(g, (const char *)o, n);
		ghostty_free(NULL, o, n);
	}
	ghostty_formatter_free(fm);
}

static size_t gvt_take_reply(gvt *g, uint8_t *dst, size_t cap) {
	size_t n = g->rlen < cap ? g->rlen : cap;
	memcpy(dst, g->reply, n);
	memmove(g->reply, g->reply + n, g->rlen - n);
	g->rlen -= n;
	return n;
}
static size_t gvt_reply_len(gvt *g) { return g->rlen; }
static char *gvt_out(gvt *g) { return g->out; }
static size_t gvt_out_len(gvt *g) { return g->olen; }
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

// Term is one emulated terminal. Safe for concurrent use.
type Term struct {
	mu   sync.Mutex
	g    *C.gvt
	cols int
	rows int
}

// New makes a cols x rows terminal with scrollbackBytes of history.
func New(cols, rows, scrollbackBytes int) (*Term, error) {
	g := C.gvt_new(C.uint16_t(cols), C.uint16_t(rows), C.size_t(scrollbackBytes))
	if g == nil {
		return nil, errors.New("gvt: cannot create terminal")
	}
	return &Term{g: g, cols: cols, rows: rows}, nil
}

// Close frees the terminal. Every method is a no-op after it: the host's
// reader can outlive the host by a write when a grandchild holds the pty
// open past the exit (the engine host waits half a second for it), and a
// call into a freed engine is a segfault, not an error.
func (t *Term) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g != nil {
		C.gvt_free(t.g)
		t.g = nil
	}
}

// Write feeds the agent's output. It returns the bytes the terminal answers
// with (cursor reports and the like), which belong on the agent's input.
func (t *Term) Write(p []byte) (reply []byte) {
	if len(p) == 0 {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return nil
	}
	C.gvt_write(t.g, (*C.uint8_t)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	if n := int(C.gvt_reply_len(t.g)); n > 0 {
		reply = make([]byte, n)
		C.gvt_take_reply(t.g, (*C.uint8_t)(unsafe.Pointer(&reply[0])), C.size_t(n))
	}
	return reply
}

// Resize reflows the terminal to cols x rows.
func (t *Term) Resize(cols, rows int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return
	}
	if cols == t.cols && rows == t.rows {
		return
	}
	C.gvt_resize(t.g, C.uint16_t(cols), C.uint16_t(rows))
	t.cols, t.rows = cols, rows
}

// Size reports the terminal's size.
func (t *Term) Size() (cols, rows int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cols, t.rows
}

// Render returns the bytes that draw the rows changed since the last Render
// (every row when full) onto a real terminal, row y of this terminal on
// screen row top+y+1. Each drawn row ends with SGR reset; nothing else about
// the real terminal's state is touched (no cursor save, no modes).
func (t *Term) Render(full bool, top int) []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return nil
	}
	f := C.int(0)
	if full {
		f = 1
	}
	C.gvt_render(t.g, f, C.int(top), C.int(t.rows), C.int(t.cols))
	return C.GoBytes(unsafe.Pointer(C.gvt_out(t.g)), C.int(C.gvt_out_len(t.g)))
}

// Cursor is the agent's cursor.
type Cursor struct {
	X, Y     int // 0-based, in this terminal
	Visible  bool
	InView   bool
	Blinking bool
	Style    int // GhosttyCursorStyle
}

// Cursor reports the agent's cursor.
func (t *Term) Cursor() Cursor {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return Cursor{}
	}
	var x, y, fl C.int
	C.gvt_cursor(t.g, &x, &y, &fl)
	return Cursor{X: int(x), Y: int(y), Visible: fl&1 != 0, InView: fl&2 != 0, Blinking: fl&4 != 0, Style: int(fl >> 8)}
}

// Mode reports a DEC private mode (ansi false) or an ANSI mode.
func (t *Term) Mode(mode int, ansi bool) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return false
	}
	return bool(C.gvt_mode(t.g, C.uint16_t(mode), C.bool(ansi)))
}

// KittyFlags reports the kitty keyboard protocol flags in effect.
func (t *Term) KittyFlags() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return 0
	}
	return int(C.gvt_kitty(t.g))
}

// Transcript is the active screen, scrollback included, with soft wraps
// joined so the real terminal can reflow it. vt keeps colours.
func (t *Term) Transcript(vt bool) []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.g == nil {
		return nil
	}
	v := C.int(0)
	if vt {
		v = 1
	}
	C.gvt_transcript(t.g, v)
	return C.GoBytes(unsafe.Pointer(C.gvt_out(t.g)), C.int(C.gvt_out_len(t.g)))
}
