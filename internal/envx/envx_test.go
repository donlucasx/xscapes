package envx

import (
	"bytes"
	"strings"
	"testing"
)

func TestLookupPrefersNewName(t *testing.T) {
	t.Setenv("XSCAPES_COLOR", "new")
	t.Setenv("ASCIISCAPES_COLOR", "old")
	if got := Lookup("COLOR"); got != "new" {
		t.Errorf("Lookup = %q, want the new name to win", got)
	}
}

func TestLookupFallsBackToLegacy(t *testing.T) {
	t.Setenv("XSCAPES_COLOR", "")
	t.Setenv("ASCIISCAPES_COLOR", "256")
	if got := Lookup("COLOR"); got != "256" {
		t.Errorf("Lookup = %q, want the pre-rename name to still work", got)
	}
}

// The positive control for the one above: with neither set, Lookup must be
// empty. Without this the fallback test passes for a reader that returns "256"
// unconditionally.
func TestLookupEmptyWhenNeitherSet(t *testing.T) {
	t.Setenv("XSCAPES_COLOR", "")
	t.Setenv("ASCIISCAPES_COLOR", "")
	if got := Lookup("COLOR"); got != "" {
		t.Errorf("Lookup = %q, want empty", got)
	}
}

func TestWarnLegacyNamesTheVariable(t *testing.T) {
	t.Setenv("ASCIISCAPES_SHADE_BLOCKS", "1")
	var b bytes.Buffer
	WarnLegacy(&b)
	out := b.String()
	if !strings.Contains(out, "ASCIISCAPES_SHADE_BLOCKS") {
		t.Errorf("warning does not name the old variable:\n%s", out)
	}
	if !strings.Contains(out, "XSCAPES_SHADE_BLOCKS") {
		t.Errorf("warning does not name the replacement:\n%s", out)
	}
}

// A warning that fires when nothing is deprecated is noise on every run, which
// is how a warning gets ignored.
func TestWarnLegacySilentWhenClean(t *testing.T) {
	t.Setenv("ASCIISCAPES_SHADE_BLOCKS", "")
	var b bytes.Buffer
	WarnLegacy(&b)
	if b.Len() != 0 {
		t.Errorf("warned with no legacy variables set:\n%s", b.String())
	}
}

func TestWarnLegacySaysWhichOneLost(t *testing.T) {
	t.Setenv("XSCAPES_CHROMA", "2.6")
	t.Setenv("ASCIISCAPES_CHROMA", "1.0")
	var b bytes.Buffer
	WarnLegacy(&b)
	if !strings.Contains(b.String(), "ignored") {
		t.Errorf("shadowed variable not reported as ignored:\n%s", b.String())
	}
}
