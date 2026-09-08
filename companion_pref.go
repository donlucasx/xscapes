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

// companionPref is which companion to draw: the environment first, so a run can
// be overridden without changing anything on disk, then the file, then the
// default.
func companionPref() string {
	if v := strings.TrimSpace(envx.Lookup("XSCAPES_COMPANION")); v != "" {
		return v
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
		return v
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
		if src := strings.TrimSpace(envx.Lookup("XSCAPES_COMPANION")); src != "" {
			fmt.Printf("  (from XSCAPES_COMPANION; the saved setting is not in use)\n")
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
