package host

import (
	"fmt"
	"strings"
)

// The escape sequences that hold the agent in its band.
//
// The band is DECSTBM, a scroll region covering rows 1..n. Inside it the
// agent's newlines scroll only those rows, its ESC[1B stops at the bottom
// margin, and its ESC[2K erases only the line the cursor is on -- so the scape
// below cannot be touched by anything Claude Code was measured emitting.
//
// Origin mode (DECOM) is the other half. Claude homes the cursor with ESC[H
// and repaints from the top when the window resizes; that is the only absolute
// move it makes. Origin mode makes ESC[H mean "the top of the region", so the
// repaint lands in the band. Claude never sets or resets DECOM itself, so once
// the host turns it on it stays on.
const (
	saveCursor    = "\x1b7"
	restoreCursor = "\x1b8"
	originOn      = "\x1b[?6h"
	originOff     = "\x1b[?6l"
	regionReset   = "\x1b[r"
)

// EnterBand pins the agent to rows 1..rows. Zero rows means the window had no
// room for a scape, so the terminal is left alone.
func EnterBand(rows int) string {
	if rows <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[1;%dr%s", rows, originOn)
}

// LeaveBand hands the whole screen back.
func LeaveBand() string { return regionReset + originOff }

// BeginPaint opens a window in which the rows below the band can be addressed.
//
// Order matters twice over. The cursor is saved FIRST because DECSTBM moves the
// cursor to home as a side effect -- saving afterwards would record the top of
// the screen and drop the agent's cursor there on every frame. Origin mode is
// dropped because with it on, absolute positions are clamped to the region, so
// the scape's rows are unaddressable.
func BeginPaint() string { return saveCursor + originOff + regionReset }

// EndPaint closes it again: the region first, then the cursor, so the restored
// position is interpreted under the same geometry it was saved in.
func EndPaint(rows int) string {
	if rows <= 0 {
		return restoreCursor
	}
	return EnterBand(rows) + restoreCursor
}

// LeaveTo hands the terminal back and leaves the cursor on row, so the shell
// that gets control next draws its prompt below the agent's last screen
// instead of on top of it.
//
// The order is the point. DECSTBM moves the cursor to home, and a parameterless
// ESC[r is still DECSTBM: resetting the region after placing the cursor pulls
// it back to row 1, and the first thing a zsh prompt emits is ESC[J, which from
// row 1 erases everything the agent ever drew. Measured, not reasoned: that is
// exactly how the agent's screen went missing on exit.
func LeaveTo(row int) string {
	if row < 1 {
		row = 1
	}
	return LeaveBand() + "\x1b[0m" + fmt.Sprintf("\x1b[%d;1H", row)
}

// Rebind moves the band to a new geometry after the window changed size,
// undoing any downward push the terminal applied, clearing the rows that
// changed hands, and leaving the agent's cursor exactly where it found it.
//
// The cursor is the whole point. DECSTBM homes it, so re-stating the region
// dropped the cursor on row 1 while the agent still believed it was in its
// input box -- and everything typed after a resize echoed onto row 1, on top
// of the transcript. Same rule as LeaveTo, from the other side: save before
// touching the region, restore after the last time you touch it.
//
// scrollUp undoes the terminal's own doing. Measured by eye on 2026-09-03
// (notes/contentprobe): Terminal.app's ALTERNATE screen anchors content to the
// BOTTOM edge, so a grow of N rows pushes the whole screen DOWN by N and
// inserts N blank rows at the top. The agent never notices -- it emits nothing
// at all on a resize -- so its UI simply slides out of the band, and whatever
// crosses the bottom edge is painted over by the scape on the next frame. SU
// over the full screen puts it back. This moves ROWS, which is something the
// host can do without modelling the UI living in them.
//
// The cursor needs no adjustment: the terminal moved the content and left the
// cursor on its absolute row, and SU moves the content back without moving the
// cursor, so the two end up in the same relationship they started in.
func Rebind(scrollUp, clearFrom, clearTo, agentRows int) string {
	var b strings.Builder
	b.WriteString(saveCursor)
	b.WriteString(originOff + regionReset)
	if scrollUp > 0 {
		fmt.Fprintf(&b, "\x1b[%dS", scrollUp)
	}
	b.WriteString(clearRowsBare(clearFrom, clearTo))
	b.WriteString(EnterBand(agentRows))
	b.WriteString(restoreCursor)
	return b.String()
}

const (
	altOn  = "\x1b[?1049h"
	altOff = "\x1b[?1049l"
)

// Open takes the screen and pins the agent's band to the top of it.
//
// The alternate screen is the answer to the one thing the band could not
// defend against. Growing a window makes the terminal pull scrolled-off lines
// back in from history, which pushes the agent's UI down and out of its band;
// the agent never notices, because it emits nothing at all on a resize and
// places its input purely by relative moves from wherever the cursor happens to
// be. Measured both ways: plain Claude survives that resize, Claude in a band on
// the main screen does not. The alternate screen has no history, so there is
// nothing to pull back and nothing moves.
//
// What it costs is the same thing: lines that scroll out of the agent's band
// are gone rather than going to scrollback. That is the trade, and it is why
// this is a flag.
//
// The switch comes first, because switching buffers resets the scroll region.
func Open(alt bool, agentRows, rows int) string {
	if alt {
		// Nothing to clear: the alternate buffer starts blank.
		return altOn + EnterBand(agentRows)
	}
	return clearRows(1, rows) + EnterBand(agentRows)
}

// Close gives the screen back.
//
// On the alternate screen that is one sequence and the shell's own screen
// returns untouched, scrollback and all. On the main screen there is nothing to
// return to, so the scape's rows have to be cleaned up by hand and the cursor
// parked below the band for whatever prompt comes next.
func Close(alt bool, agentRows, rows int) string {
	if alt {
		return LeaveBand() + "\x1b[0m" + altOff
	}
	return clearRows(agentRows+1, rows) + LeaveTo(agentRows+1)
}
