// Package envx reads xscapes settings from the environment, and keeps the
// pre-rename ASCIISCAPES_* names working.
//
// The project was called asciiscapes until 2026-09-03. Every setting it reads
// was renamed with it, and a rename is the cheapest way there is to build a
// knob that nothing reads: the value looks applied, the picture does not
// change, and a measurement taken from it is wrong with nothing saying so.
// Most of these variables exist precisely to be typed by hand while judging a
// frame, so muscle memory is the expected input, not a mistake to punish.
//
// The two halves are deliberately separate. Lookup makes an old name still
// WORK, silently, because it is called from render paths that must not write
// to a terminal mid-frame. WarnLegacy makes it LOUD, once, from a place where
// printing is safe.
package envx

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// prefix is the current name. legacyPrefix is what it used to be.
const (
	prefix       = "XSCAPES_"
	legacyPrefix = "ASCIISCAPES_"
)

// Lookup returns the value of XSCAPES_<name>, falling back to the pre-rename
// ASCIISCAPES_<name>. The new name wins when both are set. Silent by design --
// see WarnLegacy for the announcement.
func Lookup(name string) string {
	if v := os.Getenv(prefix + name); v != "" {
		return v
	}
	return os.Getenv(legacyPrefix + name)
}

// Legacy lists the ASCIISCAPES_* variables actually present in the
// environment, sorted, each rendered as "OLD (use NEW)".
func Legacy() []string {
	var out []string
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(k, legacyPrefix) || v == "" {
			// An empty value is inert -- Lookup returns "" for it either way --
			// so warning about it is a false alarm, and a warning that cries
			// wolf on a clean environment is one nobody reads.
			continue
		}
		name := strings.TrimPrefix(k, legacyPrefix)
		if os.Getenv(prefix+name) != "" {
			out = append(out, fmt.Sprintf("%s (ignored; %s%s is set)", k, prefix, name))
			continue
		}
		out = append(out, fmt.Sprintf("%s (use %s%s)", k, prefix, name))
	}
	sort.Strings(out)
	return out
}

// WarnLegacy names every old-style variable on w. Call it once, early, from
// somewhere that is not painting a frame.
func WarnLegacy(w io.Writer) {
	old := Legacy()
	if len(old) == 0 {
		return
	}
	fmt.Fprintf(w, "xscapes: %s is the old name; these still work but will not forever:\n", strings.TrimSuffix(legacyPrefix, "_"))
	for _, s := range old {
		fmt.Fprintf(w, "  %s\n", s)
	}
}
