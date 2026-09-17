package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donlucasx/xscapes/internal/envx"
	"github.com/donlucasx/xscapes/internal/event"
)

// THE SCAPE IS A CHOICE NOW (s35, 2026-09-16), the way the companion has been
// since s22: one line in ~/.config/xscapes/scape, XSCAPES_SCAPE to override a
// run, and `xscapes scape <name>` to set it -- which a running scape picks up
// on its next poll, so there is nothing to restart.
//
// THE VISTA IS THE DEFAULT since 2026-09-17, his line for the page: "the
// mountains vista ships default, but you can swap landscapes or companions
// between sessions." (The shore was the default from s35 to then.) The
// rainy window and the aquarium are drawn and on the page and are next.

const (
	ScapeShore = "shore"
	ScapeVista = "vista"
)

// ScapeNames is every scape that ships, the default first.
func ScapeNames() []string { return []string{ScapeVista, ScapeShore} }

func scapePath() (string, error) {
	h, err := event.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "scape"), nil
}

func scapeEnv() string { return strings.TrimSpace(envx.Lookup("SCAPE")) }

// scapePref is which scape to draw: the environment, then the file, then the
// shore. Normalised and validated at the point of reading, as companionPref
// is, so a stray value never makes the live loop rebuild the scene twice a
// second.
func scapePref() string {
	if v := scapeEnv(); v != "" {
		return scapeName(v)
	}
	p, err := scapePath()
	if err != nil {
		return ScapeVista
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ScapeVista
	}
	return scapeName(string(b))
}

func scapeName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, n := range ScapeNames() {
		if n == v {
			return v
		}
	}
	return ScapeVista
}

func setScapePref(name string) error {
	p, err := scapePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(name+"\n"), 0o644)
}

// cmdScape is `xscapes scape [name]`: no argument reports, an argument sets.
func cmdScape(args []string) int {
	if len(args) == 0 {
		fmt.Printf("scape: %s\n", scapePref())
		if src := scapeEnv(); src != "" {
			fmt.Printf("  (from XSCAPES_SCAPE = %q; the saved setting is not in use)\n", src)
			if scapeName(src) != strings.ToLower(strings.TrimSpace(src)) {
				fmt.Printf("  ⚠ %q is not a scape, so the shore is drawn instead\n", src)
			}
		}
		fmt.Printf("available: %s\n", strings.Join(ScapeNames(), ", "))
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	if name == "-h" || name == "--help" || name == "help" {
		fmt.Printf("usage: xscapes scape [%s]\n", strings.Join(ScapeNames(), "|"))
		fmt.Printf("  no argument prints the current scape. A running scape switches on its next frame.\n")
		return 0
	}
	if scapeName(name) != name {
		fmt.Fprintf(os.Stderr, "xscapes: no scape called %q. available: %s\n", name, strings.Join(ScapeNames(), ", "))
		return 2
	}
	if err := setScapePref(name); err != nil {
		fmt.Fprintf(os.Stderr, "xscapes: %v\n", err)
		return 1
	}
	p, _ := scapePath()
	fmt.Printf("scape: %s (saved to %s)\n", name, p)
	if src := scapeEnv(); src != "" {
		fmt.Printf("  ⚠ XSCAPES_SCAPE = %q is set in this shell and overrides it\n", src)
	}
	return 0
}
