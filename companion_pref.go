package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/envx"
	"github.com/donlucasx/xscapes/internal/event"
)

// The companion is GLOBAL, not per-repo -- the brief locks that: the landscape
// is keyed to the repo, the animal is yours everywhere. So it lives as one line
// in ~/.config/xscapes/companion, beside everything else that has to be
// derivable from $HOME alone by two processes that share no other context.

func companionPath() (string, error) {
	h, err := event.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "companion"), nil
}

// companionEnv is the environment override, and it recognises TWO spellings on
// purpose.
//
// ⚠ THE DOCUMENTED NAME NEVER WORKED. envx.Lookup PREFIXES what it is given,
// so `envx.Lookup("XSCAPES_COMPANION")` read XSCAPES_XSCAPES_COMPANION -- the
// variable README.md has advertised since the rename did nothing at all, and
// the only spelling that ever had an effect was the doubled one nobody would
// type. Fixed 2026-09-11 at his ruling, "fix it, recognise both names".
//
// The doubled spelling is kept for one release rather than deleted because
// something on a machine may have been set to whatever actually worked, and a
// key that is renamed out from under installed state orphans it silently --
// the same lesson a hook marker and a config directory both taught here. The
// correct name wins when both are set.
func companionEnv() string {
	if v := strings.TrimSpace(envx.Lookup("COMPANION")); v != "" {
		return v
	}
	return strings.TrimSpace(envx.Lookup("XSCAPES_COMPANION"))
}

// companionPref is which companion to draw: the environment first, so a run can
// be overridden without changing anything on disk, then the file, then the
// default.
//
// The name is NORMALISED and VALIDATED here, at the point of READING, not only
// where it is written. A file holding "Crab" or "octopus" used to leave the
// live loop rebuilding the companion twice a second forever, because the
// refresh compared the raw string against the animal's own name and never
// agreed with itself.
func companionPref() string {
	if v := companionEnv(); v != "" {
		return companionName(v)
	}
	p, err := companionPath()
	if err != nil {
		return companion.DefaultName
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return companion.DefaultName
	}
	if v := strings.TrimSpace(string(b)); v != "" {
		return companionName(v)
	}
	return companion.DefaultName
}

// companionName lower-cases and validates, falling back to the default rather
// than handing the renderer a name it will have to fall back from on every
// frame.
func companionName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, n := range companion.Names() {
		if n == v {
			return v
		}
	}
	return companion.DefaultName
}

func setCompanionPref(name string) error {
	p, err := companionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(name+"\n"), 0o644)
}

// cmdCompanion is `xscapes companion [name]`: no argument reports, an argument
// sets. Validated before it writes, so the file can never hold a name the
// renderer would have to fall back from.
func cmdCompanion(args []string) int {
	if len(args) == 0 {
		cur := companionPref()
		fmt.Printf("companion: %s\n", cur)
		if src := companionEnv(); src != "" {
			name := "XSCAPES_COMPANION"
			if strings.TrimSpace(envx.Lookup("COMPANION")) == "" {
				name = "XSCAPES_XSCAPES_COMPANION (the old spelling; use XSCAPES_COMPANION)"
			}
			fmt.Printf("  (from %s = %q; the saved setting is not in use)\n", name, src)
			if companionName(src) != strings.ToLower(strings.TrimSpace(src)) {
				fmt.Printf("  ⚠ %q is not a companion, so the default is drawn instead\n", src)
			}
		}
		fmt.Printf("available: %s\n", strings.Join(companion.Names(), ", "))
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	if name == "-h" || name == "--help" || name == "help" {
		fmt.Printf("usage: xscapes companion [%s]\n", strings.Join(companion.Names(), "|"))
		fmt.Printf("  no argument prints the current companion.\n")
		return 0
	}
	ok := false
	for _, n := range companion.Names() {
		if n == name {
			ok = true
		}
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "xscapes: no companion called %q. available: %s\n",
			name, strings.Join(companion.Names(), ", "))
		return 2
	}
	if err := setCompanionPref(name); err != nil {
		fmt.Fprintf(os.Stderr, "xscapes: %v\n", err)
		return 1
	}
	p, _ := companionPath()
	fmt.Printf("companion: %s\n", name)
	fmt.Printf("saved to %s -- a running scape picks it up on its next frame.\n", p)
	return 0
}
