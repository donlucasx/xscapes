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

// RebindShrinkAlt moves the band after the window got SHORTER on the alternate
// screen, where the terminal has already pulled the content up by the rows
// lost and left the agent's cursor on its absolute row.
//
// Rebind cannot be used there, and the reason is the whole fix. It saves the
// cursor, pins the smaller band, and restores -- and a restore under origin
// mode into a region that no longer contains the saved row lands on row 1
// (measured 2026-09-03, notes/shrinkprobe). Claude Code then draws its input
// box, by relative moves, at the TOP of the band over the transcript; on a
// one- or two-row tick the restore succeeds and the box is simply drawn that
// many rows below its own text. Either way a split input box, on every shrink.
//
// So: the restore happens while the region is still the full screen, then the
// cursor moves relatively, then a fresh save, and only then the band -- around
// a restore into a band that does contain the row. The content is moved too:
// the scape's share of the shrink (shrink minus the band's own shrink) is what
// left blank rows under the agent's text, and scrolling the whole screen down
// by that much puts the text's bottom back on the band's bottom, which is
// where Claude Code believes its input box is. Nothing in the band needs
// clearing after that: the rows above the text are freshly inserted blanks
// (after an SGR reset, because SD fills with the current background like any
// erase), and the old scape rows land exactly on the new scape area, which
// repaints in full.
//
// Verified in eleven geometries, notes/scrollback-audit.md: single-row ticks,
// a three-tick drag, the cursor at the top and the middle of the band, a
// shrink past the band, one the band absorbs, and a shrink followed by a grow.
func RebindShrinkAlt(shrink, bandShrink, agentRows int) string {
	var b strings.Builder
	b.WriteString(saveCursor)
	b.WriteString(originOff + regionReset + "\x1b[0m")
	if k := shrink - bandShrink; k > 0 {
		fmt.Fprintf(&b, "\x1b[%dT", k)
	}
	b.WriteString(restoreCursor)
	if bandShrink > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", bandShrink)
	}
	b.WriteString(saveCursor)
	b.WriteString(EnterBand(agentRows))
	b.WriteString(restoreCursor)
	return b.String()
}

const (
	altOn  = "\x1b[?1049h"
	altOff = "\x1b[?1049l"
	// The no-clear switch. Measured 2026-09-03 (notes/mirrorprobe): DECSET 47
	// swaps the two buffers and clears nothing, in both directions -- 400
	// round trips at 5ms with the alternate screen intact. 1049 would save,
	// clear and discard; it is the wrong tool for a round trip.
	toMain = "\x1b[?47l"
	toAlt  = "\x1b[?47h"
)

// MirrorBatch writes rows that have left the agent's band into the MAIN
// buffer, where the terminal keeps its scrollback, while the alternate screen
// stays on display.
//
// The alternate screen has no history of its own (notes/histprobe), and
// Terminal.app's view shows the main buffer ABOVE the alternate screen: scroll
// up in `xscapes claude` and the shell's last screen is there. So the agent's
// transcript can be put where the user already looks for it, with the wheel,
// selection and search that come with it, and nothing has to be built to show
// it. Rows are appended at mainRow, which starts where the shell left its
// cursor and walks down to the last row; once past it (rows+1 means full)
// every row scrolls the main buffer by one -- into history -- and lands on
// the last row.
//
// Bracketed like a paint frame: cursor and rendition saved, origin mode off
// and the region reset so the main buffer's last row can be addressed, an SGR
// reset because the erase and the scroll fill with the current background,
// and the band re-pinned around the restore. One write, so the terminal never
// sees the main buffer between two of these.
func MirrorBatch(lines []string, mainRow *int, rows, agentRows int) string {
	if len(lines) == 0 {
		return ""
	}
	if *mainRow < 1 {
		*mainRow = 1
	}
	if *mainRow > rows+1 {
		*mainRow = rows + 1
	}
	var b strings.Builder
	b.WriteString(saveCursor + originOff + regionReset + "\x1b[0m" + toMain)
	for _, l := range lines {
		// Once the buffer is full (mainRow past the last row) each row scrolls
		// FIRST and is then written on the last row, so the newest mirrored
		// row sits directly above the band with no blank row between them.
		if *mainRow > rows {
			fmt.Fprintf(&b, "\x1b[%d;1H\r\n", rows)
		}
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K%s", min(*mainRow, rows), l)
		if *mainRow <= rows {
			*mainRow++
		}
	}
	b.WriteString("\x1b[0m" + toAlt + EnterBand(agentRows) + restoreCursor)
	return b.String()
}

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

// Rules is what a terminal does to the ALTERNATE screen when the window
// changes height -- the one thing the host cannot find out by asking: the
// cursor answers DSR from wherever it is, and the content moved, or did not,
// without telling anyone. Two terminals have been read closely.
//
// Terminal.app, measured by eye 2026-09-03 (notes/contentprobe, shrinkprobe):
// content anchored to the BOTTOM both ways -- a grow pushes it down and
// inserts blanks at the top, a shrink pulls it up -- and the cursor stays on
// its absolute row through both.
//
// Ghostty 1.3.1, read from its source 2026-09-04 (src/terminal/PageList.zig
// resizeWithoutReflow, Terminal.zig restoreCursor): a grow with the cursor
// above the last row keeps content at the TOP and appends blank rows; a
// shrink scrolls rows off the top and the cursor FOLLOWS its row; DECRC
// restores the saved row verbatim. Rows of painted spaces count as text, so
// the scape is never trimmed as "trailing blank rows". His first Ghostty
// session ran under Terminal.app's rules and the transcript moved by the
// tick on every resize (_FEEDBACK.md, s15, 2026-09-04).
//
// The zero value is Ghostty's set, which is also what xterm's descendants are
// expected to do; only Terminal.app has been measured to differ. A terminal
// that behaves like neither will show up the way Ghostty did, and gets its
// own entry then.
type Rules struct {
	// GrowPushesDown: a grow of N rows pushes the content down N and inserts
	// blank rows at the top, so the host scrolls it back up before painting.
	GrowPushesDown bool
	// ShrinkKeepsCursor: a shrink pulls the content up but leaves the cursor
	// on its absolute row, so the host has to move the cursor itself.
	ShrinkKeepsCursor bool
}

var (
	AppleTerminalRules = Rules{GrowPushesDown: true, ShrinkKeepsCursor: true}
	XTermRules         = Rules{}
)

// RulesFor picks the rules by TERM_PROGRAM.
func RulesFor(termProgram string) Rules {
	if termProgram == "Apple_Terminal" {
		return AppleTerminalRules
	}
	return XTermRules
}

// RebindShrinkAltFollow is RebindShrinkAlt for a terminal whose cursor went
// up with its row through the shrink (Ghostty). The content still has to come
// down by the scape's share of the shrink, so the text's bottom lands on the
// new band's bottom -- and this time the cursor comes down with it, because
// it went up with the content and SD leaves it where it is. The save around
// DECSTBM is only there because DECSTBM homes the cursor; the first save
// carries origin mode through the region reset, as in RebindShrinkAlt.
func RebindShrinkAltFollow(shrink, bandShrink, agentRows int) string {
	var b strings.Builder
	b.WriteString(saveCursor)
	b.WriteString(originOff + regionReset + "\x1b[0m")
	k := shrink - bandShrink
	if k > 0 {
		fmt.Fprintf(&b, "\x1b[%dT", k)
	}
	b.WriteString(restoreCursor)
	if k > 0 {
		fmt.Fprintf(&b, "\x1b[%dB", k)
	}
	b.WriteString(saveCursor)
	b.WriteString(EnterBand(agentRows))
	b.WriteString(restoreCursor)
	return b.String()
}
