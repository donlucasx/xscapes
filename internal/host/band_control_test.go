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
