package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/donlucasx/xscapes/internal/spend"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// THE SECOND AND THIRD ADAPTERS: Kimi Code CLI and Hermes Agent.
//
// Both were surveyed on this machine (s35, 2026-09-16) rather than from
// documentation: Kimi's hook schema is read off the TypeScript embedded in its
// compiled binary, Hermes's off the installed checkout under ~/.hermes. Both
// run user-configured shell hooks with a JSON payload on STDIN, which is the
// shape `xscapes hook` already reads, so an adapter here is an installer and
// a few lines in translate() -- not a parser.
//
// The installers are TEXT edits, marker-delimited, never a parse-and-dump:
// config.toml and config.yaml are hand-written files with comments in them,
// and a round-trip through a TOML or YAML library would rewrite every line
// the user typed. Ours is the block between two marker comments; nothing
// outside it is touched, and uninstall removes exactly that block.
//
//   - Kimi: ~/.kimi-code/config.toml, top-level `hooks` array of tables
//     (`[[hooks]]` with event / command / timeout), appended at the end of
//     the file, where array-of-tables syntax is unambiguous.
//   - Hermes: ~/.hermes/config.yaml, top-level `hooks:` mapping of event name
//     to a list of {command, timeout}. Items are appended under the event's
//     key, or the key is added; `hooks: {}` becomes a block.
//
// ⚠ Hermes asks for CONSENT the first time each (event, command) pair fires,
// on the TTY, and records it in ~/.hermes/shell-hooks-allowlist.json. The
// installer does NOT flip `hooks_auto_accept` -- that is the user's security
// switch -- it says so instead. `hermes hooks doctor` banks the consent in
// one sitting.

// kimiHookEvents are the Kimi Code CLI events this program registers. The
// names are Claude Code's, and translate() handles them as such; StopFailure
// is Kimi's own and maps to a worry.
var kimiHookEvents = []string{
	"SessionStart", "SessionEnd", "UserPromptSubmit",
	"PreToolUse", "PostToolUse", "PostToolUseFailure",
	"PermissionRequest", "PermissionResult", "Notification", "Stop", "StopFailure", "Interrupt",
	"SubagentStart", "SubagentStop", "PreCompact",
}

// hermesHookEvents are the Hermes shell-hook events this program registers.
var hermesHookEvents = []string{
	"on_session_start", "on_session_reset", "on_session_end", "on_session_finalize",
	"pre_llm_call", "post_llm_call",
	"pre_tool_call", "post_tool_call", "pre_approval_request",
	"subagent_stop",
}

const (
	blockBegin = marker + " begin -- written by `xscapes install`; `xscapes uninstall` removes this block"
	blockEnd   = marker + " end"
	// hookTimeout is what the agents are told to wait. `xscapes hook` exits
	// inside 250ms whatever happens; five seconds is the agent's ceiling.
	hookTimeout = 5
)

// agentAdapter is one installable agent: where its config lives, and how to
// add and remove our block in it.
type agentAdapter struct {
	name    string
	path    func() (string, error)
	add     func(src []byte, bin string) ([]byte, []string, error)
	remove  func(src []byte) ([]byte, int, error)
	after   string // printed once the file is written
	missing string // printed when the config file does not exist
}

var adapters = map[string]agentAdapter{
	"kimi": {
		name: "Kimi Code CLI",
		// Kimi's home is $KIMI_CODE_HOME, else ~/.kimi-code: the same
		// answer the spend counter reads (spend.KimiHome). It was hardcoded
		// here, so with the variable set the hooks went into a file Kimi
		// never reads and the launcher's check passed on that same file
		// (Kimi's assessment, 2026-09-18, F1 of the adapter section).
		path:   func() (string, error) { return filepath.Join(spend.KimiHome(), "config.toml"), nil },
		add:    addKimiHooks,
		remove: removeKimiHooks,
		after: "Verify with `kimi doctor config`. Then run Kimi inside the scape:\n\n" +
			"  xscapes kimi\n",
		missing: "Kimi Code CLI writes config.toml under $KIMI_CODE_HOME (else ~/.kimi-code) on first run; run `kimi` once, then install.",
	},
	"hermes": {
		name:   "Hermes Agent",
		path:   func() (string, error) { return homePath(".hermes", "config.yaml") },
		add:    addHermesHooks,
		remove: removeHermesHooks,
		after: "Hermes asks for consent the first time each hook fires, on the TTY, and\n" +
			"remembers it in ~/.hermes/shell-hooks-allowlist.json. To answer once for all\n" +
			"of them, run `hermes hooks doctor` now; `hermes hooks list` shows the result.\n" +
			"(hooks_auto_accept in config.yaml is your switch; this program does not touch it.)\n\n" +
			"Then run Hermes inside the scape:\n\n  xscapes hermes\n",
		missing: "Hermes writes ~/.hermes/config.yaml on `hermes setup`; run that first, then install.",
	},
}

func homePath(parts ...string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, parts...)...), nil
}

// runInstallAgent is `xscapes install kimi|hermes`: a plan by default, the
// file written with --apply, the same shape as `install claude`.
func runInstallAgent(target string, args []string) {
	ad := adapters[target]
	fs := flag.NewFlagSet("install "+target, flag.ExitOnError)
	apply := fs.Bool("apply", false, "actually write the file (default: print the plan)")
	cfg := fs.String("config", "", "config file (default: "+ad.name+"'s own)")
	binp := fs.String("bin", "", "path to record for the hook command (default: this binary)")
	fs.Parse(args)

	path := *cfg
	if path == "" {
		p, err := ad.path()
		if err != nil {
			die(err)
		}
		path = p
	}
	bin, err := binPath(*binp)
	if err != nil {
		die(err)
	}
	orig, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		die(errors.New(ad.missing))
	} else if err != nil {
		die(err)
	}
	out, actions, err := ad.add(orig, bin)
	if err != nil {
		die(err)
	}
	fmt.Printf("agent     %s\nconfig    %s\nbinary    %s\n\n", ad.name, path, bin)
	for _, a := range actions {
		fmt.Println("  " + a)
	}
	fmt.Println()
	if !changed(orig, out) {
		fmt.Println("Nothing to do — already installed.")
		return
	}
	fmt.Printf("%d bytes → %d bytes\n", len(orig), len(out))
	if !*apply {
		fmt.Println("\nThis is a plan. Nothing was written.")
		fmt.Println("Run again with --apply to write it.")
		return
	}
	bak, err := backupConfig(target, path, orig)
	if err != nil {
		die(err)
	}
	fmt.Println("backup    " + bak)
	if err := writeConfig(path, out); err != nil {
		die(err)
	}
	fmt.Println("Written.")
	fmt.Println()
	fmt.Print(ad.after)
}

// runUninstallAgent is `xscapes uninstall kimi|hermes`.
func runUninstallAgent(target string, args []string) {
	ad := adapters[target]
	fs := flag.NewFlagSet("uninstall "+target, flag.ExitOnError)
	apply := fs.Bool("apply", false, "actually write the file (default: print the plan)")
	cfg := fs.String("config", "", "config file (default: "+ad.name+"'s own)")
	fs.Parse(args)
	path := *cfg
	if path == "" {
		p, err := ad.path()
		if err != nil {
			die(err)
		}
		path = p
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		die(err)
	}
	out, n, err := ad.remove(orig)
	if err != nil {
		die(err)
	}
	fmt.Printf("agent     %s\nconfig    %s\n%d xscapes hook entries found\n", ad.name, path, n)
	if n == 0 {
		return
	}
	if !*apply {
		fmt.Println("\nThis is a plan. Nothing was written. Run again with --apply.")
		return
	}
	if err := writeConfig(path, out); err != nil {
		die(err)
	}
	fmt.Println("Removed.")
}

// backupConfig keeps the agent's config exactly as it was before this
// program touched it, beside the Claude settings backups. Uninstall removes
// what was written; a copy of the original is what "restore" means, and it
// is what the page promises ("after a backup"). Until 2026-09-17 only the
// Claude installer kept one.
func backupConfig(target, path string, orig []byte) (string, error) {
	dir, err := homePath(".config", "xscapes", "backups")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	bak := filepath.Join(dir, target+"-config."+stamp+filepath.Ext(path))
	if err := os.WriteFile(bak, orig, 0o600); err != nil {
		return "", err
	}
	return bak, nil
}

// hooksInstalled says whether this program's hooks are in the agent's config,
// and where it looked. `xscapes claude|kimi|hermes` asks before starting:
// without hooks the scape runs on the output watcher, half-blind, and nothing
// on screen says so (his first Kimi run, 2026-09-17: no owlets, no spend, no
// tool names, and the config had no hooks at all).
func hooksInstalled(agent string) (bool, string) {
	var path string
	var err error
	if agent == "claude" {
		path, err = settingsPath("")
	} else if ad, ok := adapters[agent]; ok {
		path, err = ad.path()
	} else {
		return true, ""
	}
	if err != nil {
		return false, path
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return false, path
	}
	var n int
	switch agent {
	case "claude":
		_, n, err = removeHooks(src)
	case "kimi":
		_, n, err = removeKimiHooks(src)
	case "hermes":
		_, n, err = removeHermesHooks(src)
	}
	return err == nil && n > 0, path
}

// missingHooks is what the launcher prints instead of starting.
func missingHooks(agent, where string) string {
	name := agent
	if ad, ok := adapters[agent]; ok {
		name = ad.name
	} else if agent == "claude" {
		name = "Claude Code"
	}
	return fmt.Sprintf(`xscapes: %s's hooks are not installed (nothing of xscapes in %s).
Without them the scape only watches %s's output: no tool names, no asks, no sub-agents.

  xscapes install %s --apply    # writes the hooks (without --apply it prints the plan)
  xscapes %s                    # then this again

To run without hooks anyway: xscapes %s -watch=on   (or: xscapes inside %s)
`, name, where, name, agent, agent, agent, agent)
}

// writeConfig replaces path with out through a temp file in the same
// directory, following a symlink to its target and keeping the file's mode --
// the same care install claude takes with settings.json.
func writeConfig(path string, out []byte) error {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	mode := os.FileMode(0o600)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".xscapes-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Chmod(name, mode); err != nil {
		os.Remove(name)
		return err
	}
	return os.Rename(name, path)
}

// hookCommand is the line both agents run. No shell redirects and no trailing
// comment: Hermes splits it with shlex and runs it without a shell, and
// `xscapes hook` prints nothing and exits 0 whatever happens, so neither is
// needed. The agent name rides along so Src says who spoke.
func hookCommand(bin, ev, agent string) string {
	return fmt.Sprintf("%s hook %s %s", bin, ev, agent)
}

// ---- Kimi: TOML, an array of tables appended at the end -------------------

var kimiInlineHooks = regexp.MustCompile(`(?m)^\s*hooks\s*=`)

// foreignHook matches a hook entry that calls this program from OUTSIDE the
// marked block: one written by hand, or by an agent asked to "wire up
// xscapes" (the first outside tester's Kimi config had twelve of them,
// 2026-09-17). Installing over them would fire every event twice, and a
// doubled SubagentStart is two kittens for one subagent.
var foreignHook = regexp.MustCompile(`xscapes["']?\s+hook\b`)

// refuseForeignHooks is the error for a file that already carries such
// entries. The way out is theirs: the entries were not written by this
// program, so it does not remove them.
func refuseForeignHooks(rest []byte, file string) error {
	n := len(foreignHook.FindAll(rest, -1))
	if n == 0 {
		return nil
	}
	return fmt.Errorf("%s already carries %d hook entr%s calling `xscapes hook` outside the xscapes block (written by hand, or by an agent?). Remove them, or restore the backup they came with, then run install again; installed over them, every event would fire twice", file, n, map[bool]string{true: "y", false: "ies"}[n == 1])
}

// tomlString quotes s as a TOML basic string.
func tomlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func kimiBlock(bin string) string {
	var b strings.Builder
	b.WriteString(blockBegin + "\n")
	for _, ev := range kimiHookEvents {
		fmt.Fprintf(&b, "[[hooks]]\nevent = %q\ncommand = %s\ntimeout = %d\n",
			ev, tomlString(hookCommand(bin, ev, "kimi")), hookTimeout)
	}
	b.WriteString(blockEnd + "\n")
	return b.String()
}

// addKimiHooks appends our block, replacing an earlier one. A file that
// already declares `hooks = [...]` inline cannot also carry `[[hooks]]`
// tables -- TOML refuses to define one key twice -- so that is an error with
// a way out rather than a file Kimi will not load.
func addKimiHooks(src []byte, bin string) ([]byte, []string, error) {
	out, had, err := cutBlock(src)
	if err != nil {
		return nil, nil, err
	}
	if kimiInlineHooks.Match(out) {
		return nil, nil, errors.New("config.toml already sets `hooks = [...]` inline; add the xscapes entries to that array by hand, or move it to [[hooks]] tables and run install again")
	}
	if err := refuseForeignHooks(out, "config.toml"); err != nil {
		return nil, nil, err
	}
	if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}
	if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n\n")) {
		out = append(out, '\n')
	}
	out = append(out, kimiBlock(bin)...)
	var actions []string
	verb := "add"
	if had > 0 {
		verb = "rewrite"
	}
	for _, ev := range kimiHookEvents {
		actions = append(actions, fmt.Sprintf("%-8s [[hooks]] event = %q", verb, ev))
	}
	return out, actions, nil
}

func removeKimiHooks(src []byte) ([]byte, int, error) {
	out, had, err := cutBlock(src)
	if err != nil {
		return nil, 0, err
	}
	out = bytes.TrimRight(out, "\n")
	if len(out) > 0 {
		out = append(out, '\n')
	}
	return out, had, nil
}

// cutBlock removes the marker-delimited block and reports how many `[[hooks]]`
// or `- command:` entries it held. Legacy markers are recognised too, so a
// block written before the rename still comes out.
func cutBlock(src []byte) (out []byte, entries int, err error) {
	lines := strings.Split(string(src), "\n")
	var keep []string
	in := false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		switch {
		case !in && isBlockBegin(t):
			in = true
			continue
		case in && isBlockEnd(t):
			in = false
			continue
		case in:
			if strings.HasPrefix(t, "[[hooks]]") {
				entries++
			}
			continue
		}
		keep = append(keep, l)
	}
	if in {
		return nil, 0, errors.New("found the xscapes begin marker but no end marker; the block was edited by hand, remove it manually")
	}
	return []byte(strings.Join(keep, "\n")), entries, nil
}

func isBlockBegin(t string) bool {
	for _, m := range append([]string{marker}, legacyMarkers...) {
		if strings.HasPrefix(t, m+" begin") {
			return true
		}
	}
	return false
}

func isBlockEnd(t string) bool {
	for _, m := range append([]string{marker}, legacyMarkers...) {
		if t == m+" end" {
			return true
		}
	}
	return false
}

// ---- Hermes: YAML, items under a top-level `hooks:` mapping ----------------

// yamlString quotes s as a double-quoted YAML scalar.
func yamlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

var (
	hermesHooksKey = regexp.MustCompile(`^hooks:\s*(.*?)\s*(#.*)?$`)
	topLevelKey    = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*:`)
	eventKey       = regexp.MustCompile(`^  ([A-Za-z_][A-Za-z0-9_]*):\s*(.*?)\s*(#.*)?$`)
)

// hermesItem is one list entry under an event: two lines, marked on the first.
func hermesItem(bin, ev string) []string {
	return []string{
		fmt.Sprintf("    - command: %s  %s", yamlString(hookCommand(bin, ev, "hermes")), marker),
		fmt.Sprintf("      timeout: %d", hookTimeout),
	}
}

func isOurItem(l string) bool {
	t := strings.TrimSpace(l)
	if !strings.HasPrefix(t, "- command:") {
		return false
	}
	for _, m := range append([]string{marker}, legacyMarkers...) {
		if strings.HasSuffix(t, m) {
			return true
		}
	}
	return false
}

// hooksSpan finds the top-level `hooks:` line and the extent of its block:
// [start, end) line indices, end exclusive. inline is the scalar on the key
// line (`{}`, or something we refuse to edit).
func hooksSpan(lines []string) (start, end int, inline string, found bool) {
	for i, l := range lines {
		m := hermesHooksKey.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		start, inline, found = i, m[1], true
		end = len(lines)
		for j := i + 1; j < len(lines); j++ {
			t := lines[j]
			if t == "" || strings.HasPrefix(t, " ") || strings.HasPrefix(t, "#") && false {
				continue
			}
			if !strings.HasPrefix(t, " ") {
				end = j
				break
			}
		}
		// Trailing blank lines belong to the next key, not to the block.
		for end > start+1 && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		return
	}
	return 0, 0, "", false
}

// addHermesHooks puts one item under each event, creating the key when it is
// missing and the whole mapping when the file has `hooks: {}` or no hooks
// key at all. Existing items -- the user's own hooks -- are untouched and ours
// go after them.
func addHermesHooks(src []byte, bin string) ([]byte, []string, error) {
	out, had, err := removeHermesItems(src)
	if err != nil {
		return nil, nil, err
	}
	if err := refuseForeignHooks(out, "config.yaml"); err != nil {
		return nil, nil, err
	}
	lines := strings.Split(string(out), "\n")
	start, end, inline, found := hooksSpan(lines)
	switch {
	case !found:
		// No hooks key: add one at the end.
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		lines = append(lines, "hooks:")
		start, end = len(lines)-1, len(lines)
	case inline == "{}":
		lines[start] = "hooks:"
	case inline != "":
		return nil, nil, fmt.Errorf("config.yaml has `hooks: %s` on one line; write it as a block mapping and run install again", inline)
	}
	body := append([]string(nil), lines[start+1:end]...)
	// Append under an existing event key, or add the key.
	for _, ev := range hermesHookEvents {
		item := hermesItem(bin, ev)
		placed := false
		for i := 0; i < len(body); i++ {
			m := eventKey.FindStringSubmatch(body[i])
			if m == nil || m[1] != ev {
				continue
			}
			if m[2] != "" && m[2] != "[]" {
				return nil, nil, fmt.Errorf("config.yaml has `%s: %s` on one line under hooks; write it as a list and run install again", ev, m[2])
			}
			if m[2] == "[]" {
				body[i] = "  " + ev + ":"
			}
			// The event's items run until the next `  key:` at two spaces.
			j := i + 1
			for j < len(body) && !eventKey.MatchString(body[j]) {
				j++
			}
			// Drop trailing blank lines inside the span so ours sits with
			// the others.
			k := j
			for k > i+1 && strings.TrimSpace(body[k-1]) == "" {
				k--
			}
			body = append(body[:k], append(item, body[k:]...)...)
			placed = true
			break
		}
		if !placed {
			body = append(body, "  "+ev+":")
			body = append(body, item...)
		}
	}
	lines = append(lines[:start+1], append(body, lines[end:]...)...)
	res := strings.Join(lines, "\n")
	if !strings.HasSuffix(res, "\n") {
		res += "\n"
	}
	var actions []string
	verb := "add"
	if had > 0 {
		verb = "rewrite"
	}
	for _, ev := range hermesHookEvents {
		actions = append(actions, fmt.Sprintf("%-8s hooks.%s: - command: xscapes hook %s hermes", verb, ev, ev))
	}
	return []byte(res), actions, nil
}

// removeHermesItems drops our items and the event keys they leave empty, and
// collapses an emptied mapping back to `hooks: {}`.
func removeHermesItems(src []byte) ([]byte, int, error) {
	lines := strings.Split(string(src), "\n")
	start, end, inline, found := hooksSpan(lines)
	if !found || inline != "" {
		return src, 0, nil
	}
	body := lines[start+1 : end]
	var kept []string
	n := 0
	for i := 0; i < len(body); i++ {
		if !isOurItem(body[i]) {
			kept = append(kept, body[i])
			continue
		}
		n++
		// The item's continuation lines are indented deeper than its dash.
		indent := len(body[i]) - len(strings.TrimLeft(body[i], " "))
		for i+1 < len(body) {
			nx := body[i+1]
			if strings.TrimSpace(nx) == "" {
				break
			}
			if len(nx)-len(strings.TrimLeft(nx, " ")) > indent {
				i++
				continue
			}
			break
		}
	}
	if n == 0 {
		return src, 0, nil
	}
	// Event keys with nothing under them go too.
	var body2 []string
	for i := 0; i < len(kept); i++ {
		if m := eventKey.FindStringSubmatch(kept[i]); m != nil && m[2] == "" {
			empty := true
			for j := i + 1; j < len(kept); j++ {
				if strings.TrimSpace(kept[j]) == "" {
					continue
				}
				if eventKey.MatchString(kept[j]) {
					break
				}
				empty = false
				break
			}
			if empty {
				continue
			}
		}
		body2 = append(body2, kept[i])
	}
	nonBlank := 0
	for _, l := range body2 {
		if strings.TrimSpace(l) != "" {
			nonBlank++
		}
	}
	var out []string
	out = append(out, lines[:start]...)
	if nonBlank == 0 {
		out = append(out, "hooks: {}")
	} else {
		out = append(out, "hooks:")
		out = append(out, body2...)
	}
	out = append(out, lines[end:]...)
	return []byte(strings.Join(out, "\n")), n, nil
}

func removeHermesHooks(src []byte) ([]byte, int, error) {
	return removeHermesItems(src)
}
