package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/term"
)

// ⚠ XSCAPES_HOME, not HOME. event.Home() checks XSCAPES_HOME first, so a test
// that redirects HOME alone reads and WRITES the real ~/.config/xscapes and
// tramples the user's own saved companion.
func prefHome(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Setenv("XSCAPES_HOME", d)
	t.Setenv("XSCAPES_COMPANION", "")
	t.Setenv("XSCAPES_XSCAPES_COMPANION", "")
	t.Setenv("ASCIISCAPES_COMPANION", "")
	return d
}

// TestCompanionEnvOverrideIsLive is the assertion that was missing, and its
// absence is why XSCAPES_COMPANION could be advertised in README.md while
// doing nothing at all: envx.Lookup PREFIXES its argument, so the code asked
// for XSCAPES_XSCAPES_COMPANION. No test in the tree had ever selected a
// companion by environment.
func TestCompanionEnvOverrideIsLive(t *testing.T) {
	for _, tc := range []struct {
		name, env, val, want string
	}{
		{"the documented name", "XSCAPES_COMPANION", "cat", "cat"},
		{"the documented name, crab", "XSCAPES_COMPANION", "crab", "crab"},
		{"the old doubled spelling still works", "XSCAPES_XSCAPES_COMPANION", "cat", "cat"},
		{"the pre-rename prefix still works", "ASCIISCAPES_COMPANION", "cat", "cat"},
		{"case is normalised", "XSCAPES_COMPANION", "CAT", "cat"},
		{"whitespace is trimmed", "XSCAPES_COMPANION", "  cat \n", "cat"},
		{"a name that is not a companion falls back", "XSCAPES_COMPANION", "octopus", companion.DefaultName},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prefHome(t)
			t.Setenv(tc.env, tc.val)
			if got := companionPref(); got != tc.want {
				t.Fatalf("%s=%q: companionPref() = %q, want %q", tc.env, tc.val, got, tc.want)
			}
			// The name is not the point; the ANIMAL is. Assert what actually
			// gets drawn, because a preference nothing renders is a knob
			// nobody reads.
			if got := companion.New(companionPref()).Name(); got != tc.want {
				t.Fatalf("%s=%q: the companion drawn is %q, want %q", tc.env, tc.val, got, tc.want)
			}
		})
	}
}

// The correct spelling wins over the old one when both are set, so a stale
// export cannot quietly outrank the documented variable.
func TestCompanionEnvPrecedence(t *testing.T) {
	prefHome(t)
	t.Setenv("XSCAPES_COMPANION", "crab")
	t.Setenv("XSCAPES_XSCAPES_COMPANION", "cat")
	if got := companionPref(); got != "crab" {
		t.Fatalf("both spellings set: got %q, want the correct name to win (crab)", got)
	}
}

// The environment outranks the saved file, and the saved file outranks the
// default. Both directions asserted, because "the env wins" is the thing that
// will surprise someone who has it exported in a shell profile.
func TestCompanionFileAndEnvOrder(t *testing.T) {
	d := prefHome(t)
	if err := os.WriteFile(filepath.Join(d, "companion"), []byte("cat\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := companionPref(); got != "cat" {
		t.Fatalf("saved file alone: got %q, want cat", got)
	}
	t.Setenv("XSCAPES_COMPANION", "crab")
	if got := companionPref(); got != "crab" {
		t.Fatalf("env over file: got %q, want crab", got)
	}
}

// A file holding a name that is not a companion must not leave the live loop
// rebuilding the animal on every poll forever: companionPref normalises at the
// point of READING, so the refresh compares a name against itself.
func TestCompanionFileGarbageSettles(t *testing.T) {
	d := prefHome(t)
	for _, junk := range []string{"Crab\n", "octopus\n", "  CAT  \n", "\n"} {
		if err := os.WriteFile(filepath.Join(d, "companion"), []byte(junk), 0o644); err != nil {
			t.Fatal(err)
		}
		got := companionPref()
		if companion.New(got).Name() != got {
			t.Errorf("file %q -> pref %q, which does not name itself: the live loop would "+
				"rebuild the companion on every poll", junk, got)
		}
	}
}

// And the swap has to reach the PICTURE. A rendered frame of a cat must not be
// a rendered frame of a crab -- the existing companion_live_test.go asserts
// MoonX, which is a constant per layout and cannot fail in either direction.
func TestTheTwoCompanionsRenderDifferently(t *testing.T) {
	frame := func(name string) string {
		c := canvas.New(40, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		cat := companion.New(name)
		cat.FaceLeft(true)
		cat.Draw(c.Near(), 4, 3, 1.7, companion.Working)
		return c.HTMLFragmentAs(8, term.Profile256)
	}
	if frame("cat") == frame("crab") {
		t.Fatal("the cat and the crab render the same frame: the swap reaches nothing")
	}
}
