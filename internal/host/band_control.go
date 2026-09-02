package host

import "fmt"

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
// clearing the rows that changed hands, and leaves the agent's cursor exactly
// where it found it.
//
// The cursor is the whole point. DECSTBM homes it, so re-stating the region
// dropped the cursor on row 1 while the agent still believed it was in its
// input box -- and everything typed after a resize echoed onto row 1, on top
// of the transcript. Same rule as LeaveTo, from the other side: save before
// touching the region, restore after the last time you touch it.
func Rebind(clearFrom, clearTo, agentRows int) string {
	return saveCursor + clearRowsBare(clearFrom, clearTo) + EnterBand(agentRows) + restoreCursor
}
