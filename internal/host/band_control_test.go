package host

import (
	"strings"
	"testing"
)

func idx(t *testing.T, s, sub, what string) int {
	t.Helper()
	i := strings.Index(s, sub)
	if i < 0 {
		t.Fatalf("%s missing from %q", what, s)
	}
	return i
}

func TestEnterBandPinsTheAgentToTheTopRows(t *testing.T) {
	got := EnterBand(20)
	idx(t, got, "\x1b[1;20r", "scroll region rows 1-20")
	idx(t, got, "\x1b[?6h", "origin mode on")
}

// Origin mode is what makes the band survive a resize. Claude homes the cursor
// with ESC[H and repaints from the top when the window changes -- its only
// absolute move. Without origin mode that lands on the screen's row 1 whatever
// the region says; with it, it lands on the band's first row.
func TestEnterBandTurnsOnOriginModeSoResizeHomesIntoTheBand(t *testing.T) {
	if !strings.Contains(EnterBand(12), "\x1b[?6h") {
		t.Error("EnterBand does not set origin mode")
	}
}

// DECSTBM moves the cursor to home as a side effect. Saving after resetting the
// region would therefore save the wrong position and drop the agent's cursor at
// the top of its band on every frame the scape paints.
func TestPaintSavesTheCursorBeforeTouchingTheScrollRegion(t *testing.T) {
	got := BeginPaint()
	save := idx(t, got, "\x1b7", "cursor save")
	reset := idx(t, got, "\x1b[r", "scroll region reset")
	if save > reset {
		t.Errorf("cursor saved at %d, after the region was reset at %d: %q", save, reset, got)
	}
}

// The scape lives below the band, which origin mode makes unaddressable: with
// it on, absolute positions are relative to the region. Painting has to drop
// both, and put both back.
func TestPaintDropsOriginModeSoRowsBelowTheBandCanBeAddressed(t *testing.T) {
	if !strings.Contains(BeginPaint(), "\x1b[?6l") {
		t.Error("BeginPaint leaves origin mode on, so the scape rows cannot be addressed")
	}
}

func TestEndPaintRestoresTheRegionBeforeTheCursor(t *testing.T) {
	got := EndPaint(20)
	region := idx(t, got, "\x1b[1;20r", "scroll region restored")
	restore := idx(t, got, "\x1b8", "cursor restore")
	if region > restore {
		t.Errorf("region restored at %d, after the cursor at %d: %q", region, restore, got)
	}
	if !strings.Contains(got, "\x1b[?6h") {
		t.Error("EndPaint does not put origin mode back")
	}
}

func TestLeaveBandGivesTheWholeScreenBack(t *testing.T) {
	got := LeaveBand()
	idx(t, got, "\x1b[r", "scroll region reset")
	idx(t, got, "\x1b[?6l", "origin mode off")
}

// A band of zero rows means the window was too short for a scape. Pinning the
// agent to nothing would be worse than not pinning it at all.
func TestNoBandWhenThereAreNoRowsToSpare(t *testing.T) {
	if got := EnterBand(0); got != "" {
		t.Errorf("EnterBand(0) = %q, want nothing", got)
	}
}

// DECSTBM homes the cursor -- ESC[r included. Measured: on exit the host put
// the cursor below the band and then reset the region one last time, which
// pulled it back to row 1, and the shell's prompt redraw (which begins with
// ESC[J) then erased the agent's whole screen from there.
//
// So the rule is the mirror of BeginPaint's: nothing may touch the scroll
// region AFTER the cursor has been placed.
func TestLeaveToPlacesTheCursorAfterTheLastRegionChange(t *testing.T) {
	got := LeaveTo(19)
	last := strings.LastIndex(got, "\x1b[r")
	if last < 0 {
		t.Fatalf("no scroll-region reset in %q", got)
	}
	cur := idx(t, got, "\x1b[19;1H", "cursor placed on row 19")
	if cur < last {
		t.Errorf("cursor placed at %d, before the last region reset at %d: %q", cur, last, got)
	}
}

func TestLeaveToStillGivesTheWholeScreenBack(t *testing.T) {
	got := LeaveTo(5)
	idx(t, got, "\x1b[r", "scroll region reset")
	idx(t, got, "\x1b[?6l", "origin mode off")
}

// Re-pinning the band on a resize must not move the agent's cursor.
//
// DECSTBM homes the cursor, so re-stating the region left it at row 1 while the
// agent still believed it was in its input box. Everything typed after a resize
// then echoed onto row 1, on top of the transcript -- which is exactly what he
// saw: "PeHello claude lets get to work" written over "Permission allow rule".
func TestRebindLeavesTheCursorWhereItFoundIt(t *testing.T) {
	got := Rebind(0, 21, 30, 48)
	save := idx(t, got, "\x1b7", "cursor save")
	restore := strings.LastIndex(got, "\x1b8")
	if restore < 0 {
		t.Fatalf("no cursor restore in %q", got)
	}
	lastRegion := strings.LastIndex(got, "r\x1b[?6h") // the end of EnterBand
	if save != 0 {
		t.Errorf("cursor saved at %d, want first: %q", save, got)
	}
	if restore < lastRegion {
		t.Errorf("cursor restored at %d, before the last region change at %d: %q", restore, lastRegion, got)
	}
}

func TestRebindClearsTheRowsItIsToldTo(t *testing.T) {
	got := Rebind(0, 21, 23, 48)
	for _, row := range []string{"\x1b[21;1H", "\x1b[22;1H", "\x1b[23;1H"} {
		idx(t, got, row+"\x1b[2K", "clear of row "+row)
	}
	if strings.Contains(got, "\x1b[24;1H\x1b[2K") {
		t.Error("cleared a row outside the range it was given")
	}
}

// The alternate screen exists to remove the reflow entirely.
//
// Measured: growing a window makes the terminal pull scrolled-off lines back in
// from history, which pushes the agent's UI down and out of its band -- and the
// agent never notices, because it emits nothing at all on a resize and places
// its input purely by relative moves. The alternate screen has no history, so
// there is nothing to pull back and nothing moves.

func TestAltScreenIsTakenBeforeTheBandIsPinned(t *testing.T) {
	got := Open(true, 44, 72)
	screen := idx(t, got, "\x1b[?1049h", "switch to the alternate screen")
	region := idx(t, got, "\x1b[1;44r", "band pinned")
	if screen > region {
		t.Errorf("screen switched at %d, after the band was pinned at %d: switching resets the scroll region, so it has to come first", screen, region)
	}
}

func TestAltScreenNeedsNoClearingOfTheMainScreen(t *testing.T) {
	got := Open(true, 44, 72)
	if strings.Contains(got, "\x1b[2K") {
		t.Error("cleared rows one at a time on a screen that starts blank")
	}
}

func TestMainScreenStillClearsBeforePinning(t *testing.T) {
	got := Open(false, 20, 30)
	idx(t, got, "\x1b[1;1H\x1b[2K", "row 1 cleared")
	idx(t, got, "\x1b[1;20r", "band pinned")
}

func TestClosingTheAltScreenGivesTheShellItsScreenBack(t *testing.T) {
	got := Close(true, 44, 72)
	region := idx(t, got, "\x1b[r", "scroll region released")
	back := idx(t, got, "\x1b[?1049l", "main screen restored")
	if back < region {
		t.Errorf("main screen restored at %d, before the region was released at %d: the region would follow us back", back, region)
	}
}

// On the main screen there is no screen to hand back, so the rows have to be
// cleaned up and the cursor parked below the band instead.
func TestClosingTheMainScreenClearsTheScapeAndParksTheCursor(t *testing.T) {
	got := Close(false, 20, 30)
	idx(t, got, "\x1b[21;1H\x1b[2K", "scape row cleared")
	idx(t, got, "\x1b[21;1H", "cursor parked below the band")
	if strings.Contains(got, "\x1b[?1049l") {
		t.Error("tried to leave an alternate screen it never entered")
	}
}

// A grow has to undo the terminal's downward push BEFORE the rows are cleared,
// or the clear is computed against a screen whose content is still displaced.
func TestRebindScrollsBackUpBeforeClearing(t *testing.T) {
	got := Rebind(12, 21, 30, 48)
	su := idx(t, got, "\x1b[12S", "scroll up")
	firstClear := strings.Index(got, "\x1b[2K")
	if firstClear < 0 {
		t.Fatal("no erase in the rebind at all")
	}
	if su > firstClear {
		t.Errorf("scroll-up at %d comes AFTER the first erase at %d: %q", su, firstClear, got)
	}
	region := strings.Index(got, regionReset)
	if region > su {
		t.Errorf("the region is reset at %d, after the scroll at %d -- the scroll would be confined to the band", region, su)
	}
	// And the cursor bracket still holds.
	if !strings.HasPrefix(got, saveCursor) || !strings.HasSuffix(got, restoreCursor) {
		t.Errorf("the cursor save/restore no longer brackets the sequence: %q", got)
	}
}

// Zero means no push to undo, and must emit no scroll at all: a stray SU on
// every shrink would throw away a row of the agent's screen each time.
func TestRebindWithoutAGrowEmitsNoScroll(t *testing.T) {
	got := Rebind(0, 21, 30, 48)
	if strings.Contains(got, "S") && strings.Contains(got, "\x1b[0S") {
		t.Errorf("emitted a zero scroll: %q", got)
	}
	for _, n := range []string{"\x1b[1S", "\x1b[2S", "\x1b[12S"} {
		if strings.Contains(got, n) {
			t.Errorf("emitted %q with no grow: %q", n, got)
		}
	}
}
