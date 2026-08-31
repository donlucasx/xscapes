package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// marker identifies our hook entries on a re-run. It lives inside the command
// string as a shell comment rather than as a sibling JSON field, because a
// comment survives anything that re-marshals the file, and because a person
// reading their own settings.json should be able to see who put it there.
const marker = "# asciiscapes:v1"

// hookEvents are the Claude Code events we register, each mapped to the
// protocol. The list is short on purpose: every entry is one more process
// exec per turn, and an event we do not use is pure cost.
//
// Notification is registered WITHOUT a matcher. Claude Code can filter on
// notification_type for us, which would save an exec on each idle nag, but a
// matcher that silently stops matching fails OPEN -- we would get the nag back
// and the cat would knock once a minute. Filtering in Go fails closed. The
// nag costs one 2ms async exec a minute; being wrong costs the whole idea.
var hookEvents = []string{
	"SessionStart",
	"SessionEnd",
	"UserPromptSubmit",
	"PreToolUse",
	"PostToolUse",
	"PostToolUseFailure",
	"PermissionRequest",
	"Notification",
	"Stop",
	"SubagentStart",
	"SubagentStop",
	"PreCompact",
}

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	var (
		apply  = fs.Bool("apply", false, "actually write the file (default: print the plan)")
		target = fs.String("settings", "", "settings file (default: ~/.claude/settings.json)")
		binp   = fs.String("bin", "", "path to record for the hook command (default: this binary)")
	)
	// The target word comes before the flags, and Go's flag package stops
	// parsing at the first non-flag argument -- so `install claude --settings X`
	// would silently ignore --settings and plan against the real file. Strip
	// the word first. (Found exactly that way: the plan named ~/.claude even
	// though --settings pointed at a copy.)
	args = takeTarget(args)
	fs.Parse(args)

	path, err := settingsPath(*target)
	if err != nil {
		die(err)
	}
	bin, err := binPath(*binp)
	if err != nil {
		die(err)
	}

	orig, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		orig = []byte("{}\n")
	} else if err != nil {
		die(err)
	}

	out, actions, err := addHooks(orig, bin)
	if err != nil {
		die(err)
	}

	fmt.Printf("settings  %s\n", path)
	fmt.Printf("binary    %s\n", bin)
	fmt.Println()
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
		// Plan by default. This file holds Lucas's permission allowlist and
		// a PreToolUse hook that blocks secret expansion; a tool that edits
		// it silently on first run has not earned that.
		fmt.Println("\nThis is a plan. Nothing was written.")
		fmt.Println("Run again with --apply to write it.")
		return
	}
	if err := writeSettings(path, orig, out); err != nil {
		die(err)
	}
	fmt.Println("Written.")
	fmt.Println("\nThe context moon needs one more line, which install does NOT add because")
	fmt.Println("it would take over a statusLine you wrote yourself. To wire it, make your")
	fmt.Println("statusLine command:")
	fmt.Printf("\n  %q statusline -- <your existing statusline command>\n", bin)
}

func runUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	var (
		apply  = fs.Bool("apply", false, "actually write the file (default: print the plan)")
		target = fs.String("settings", "", "settings file (default: ~/.claude/settings.json)")
	)
	args = takeTarget(args)
	fs.Parse(args)

	path, err := settingsPath(*target)
	if err != nil {
		die(err)
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		die(err)
	}
	out, n, err := removeHooks(orig)
	if err != nil {
		die(err)
	}
	fmt.Printf("settings  %s\n%d asciiscapes entries found\n", path, n)
	if n == 0 {
		return
	}
	if !*apply {
		fmt.Println("\nThis is a plan. Nothing was written. Run again with --apply.")
		return
	}
	if err := writeSettings(path, orig, out); err != nil {
		die(err)
	}
	fmt.Println("Removed.")
}

// takeTarget consumes a leading `claude` so the flags after it still parse.
func takeTarget(args []string) []string {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		if args[0] != "claude" {
			fmt.Fprintf(os.Stderr, "asciiscapes: only `claude` is supported, got %q\n", args[0])
			os.Exit(2)
		}
		return args[1:]
	}
	return args
}

func settingsPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".claude", "settings.json"), nil
}

// binPath is the absolute path recorded in the hook command.
//
// Symlinks are deliberately NOT resolved. Lucas's own claude binary is reached
// through ~/.local/bin/claude -> versions/2.1.251; resolving that kind of link
// records a path that a later upgrade invalidates, while the link itself keeps
// working.
func binPath(override string) (string, error) {
	p := override
	if p == "" {
		var err error
		if p, err = os.Executable(); err != nil {
			return "", err
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	// The path is about to be embedded in a shell command inside a JSON
	// string. Quoting handles spaces; these characters would survive quoting
	// and execute.
	if strings.ContainsAny(abs, "\"$`\\\n") {
		return "", fmt.Errorf("refusing to install: binary path contains a shell metacharacter: %q", abs)
	}
	return abs, nil
}

// command builds the hook line.
//
// The redirect comes first and `|| true` second, so that a moved or deleted
// binary produces exactly nothing: no output in the agent's session, no
// non-zero exit, no blocked turn. A decoration that breaks the tool it
// decorates is worse than no decoration.
func command(bin, ev string) string {
	return fmt.Sprintf("%q hook %s >/dev/null 2>&1 || true %s", bin, ev, marker)
}

// entry is the matcher-group object we insert into hooks.<Event>.
// No matcher field: an omitted matcher matches every event of that type, which
// is what we want everywhere, and it avoids guessing at matcher semantics that
// differ per event (tool_name here, notification_type there, agent_type for
// the subagent events).
func entry(bin, ev, indent string) string {
	i1 := indent
	i2 := indent + "  "
	i3 := indent + "    "
	// Not json.Marshal: it HTML-escapes, turning `>/dev/null 2>&1` into
	// `\u003e/dev/null 2\u003e\u00261`. Functionally identical, unreadable in
	// a file the user is being asked to review before it is written.
	var cb bytes.Buffer
	enc := json.NewEncoder(&cb)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(command(bin, ev))
	cmd := bytes.TrimRight(cb.Bytes(), "\n")
	return fmt.Sprintf("{\n%s\"hooks\": [\n%s{\n%s\"type\": \"command\",\n%s\"command\": %s,\n%s\"async\": true\n%s}\n%s]\n%s}",
		i2, i3, i3+"  ", i3+"  ", cmd, i3+"  ", i3, i2, i1)
}

// addHooks splices our entries into the file without reformatting it.
//
// The whole design is one rule: only ever INSERT bytes, never rewrite a span
// we did not author. A round-trip through map[string]any would reorder every
// key (Go marshals maps sorted) and re-escape every string, producing a
// thousand-line diff on a file whose contents are load-bearing. Insertion
// keeps the diff to the lines we added, which is also the only diff a person
// can actually review.
func addHooks(src []byte, bin string) ([]byte, []string, error) {
	var actions []string
	out := append([]byte(nil), src...)

	for _, ev := range hookEvents {
		has, err := hasOurEntry(out, ev)
		if err != nil {
			return nil, nil, err
		}
		if has {
			actions = append(actions, "skip    "+ev+"  (already installed)")
			continue
		}
		next, where, err := insertEntry(out, ev, bin)
		if err != nil {
			return nil, nil, err
		}
		out = next
		actions = append(actions, "add     "+ev+"  ("+where+")")
	}
	return out, actions, nil
}

// insertEntry adds one matcher-group, creating hooks or hooks.<Event> if
// needed. Returns a description of which of the three cases applied, because
// "created the hooks object" and "appended to an existing array" are very
// different amounts of risk and the plan should say which one it is.
func insertEntry(src []byte, ev, bin string) ([]byte, string, error) {
	hooksSpan, err := valueSpan(src, nil, "hooks")
	if err != nil {
		return nil, "", err
	}

	if hooksSpan == nil {
		// No hooks key at all: add one at the end of the top-level object.
		closing, err := lastBrace(src)
		if err != nil {
			return nil, "", err
		}
		body := fmt.Sprintf("\"hooks\": {\n    %q: [\n      %s\n    ]\n  }", ev, entry(bin, ev, "      "))
		if hasMembers(src, closing) {
			return splice(src, trimTrailingNewlineIndent(src, closing), ",\n  "+body), "created hooks", nil
		}
		return splice(src, closing, "\n  "+body+"\n"), "created hooks", nil
	}

	evSpan, err := valueSpan(src, hooksSpan, ev)
	if err != nil {
		return nil, "", err
	}

	if evSpan == nil {
		// hooks exists but not this event: add the key inside it.
		closing := hooksSpan.end - 1 // the '}' of the hooks object
		body := fmt.Sprintf("%q: [\n      %s\n    ]", ev, entry(bin, ev, "      "))
		if hasMembers(src[hooksSpan.start:], closing-hooksSpan.start) {
			return splice(src, trimTrailingNewlineIndent(src, closing), ",\n    "+body), "added event", nil
		}
		return splice(src, closing, "\n    "+body+"\n  "), "added event", nil
	}

	// The array exists: append before its ']'.
	closing := evSpan.end - 1
	body := entry(bin, ev, "      ")
	if hasMembers(src[evSpan.start:], closing-evSpan.start) {
		return splice(src, trimTrailingNewlineIndent(src, closing), ",\n      "+body), "appended", nil
	}
	return splice(src, closing, "\n      "+body+"\n    "), "appended", nil
}

type span struct{ start, end int }

// valueSpan finds the byte range of the value of key, either at the top level
// (parent nil) or inside parent's object. Offsets are into src.
//
// This uses the decoder's own offsets rather than scanning for braces, because
// a brace inside a string literal -- and this file is full of shell commands
// with braces in them -- would fool any scanner that does not tokenise.
func valueSpan(src []byte, parent *span, key string) (*span, error) {
	buf := src
	base := 0
	if parent != nil {
		buf = src[parent.start:parent.end]
		base = parent.start
	}
	dec := json.NewDecoder(bytes.NewReader(buf))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("settings is not valid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		name, _ := kt.(string)
		afterKey := int(dec.InputOffset())

		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := int(dec.InputOffset())
		if name != key {
			continue
		}
		// Walk forward from the key to the first byte of the value, past the
		// colon and any whitespace.
		s := afterKey
		for s < end && (buf[s] == ':' || buf[s] == ' ' || buf[s] == '\t' || buf[s] == '\n' || buf[s] == '\r') {
			s++
		}
		return &span{start: base + s, end: base + end}, nil
	}
	return nil, nil
}

// lastBrace finds the closing brace of the top-level object.
func lastBrace(src []byte) (int, error) {
	dec := json.NewDecoder(bytes.NewReader(src))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return 0, fmt.Errorf("settings is not valid JSON: %w", err)
	}
	end := int(dec.InputOffset())
	for i := end - 1; i >= 0; i-- {
		if src[i] == '}' {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no closing brace")
}

// hasMembers reports whether the container ending at closing already holds
// something, which decides whether we need a leading comma.
func hasMembers(src []byte, closing int) bool {
	for i := closing - 1; i >= 0; i-- {
		switch src[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return false
		default:
			return true
		}
	}
	return false
}

// trimTrailingNewlineIndent moves an insertion point back over the whitespace
// before a closing bracket, so the comma lands after the last member rather
// than after a blank line.
func trimTrailingNewlineIndent(src []byte, closing int) int {
	i := closing
	for i > 0 {
		c := src[i-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i--
			continue
		}
		break
	}
	return i
}

func splice(src []byte, at int, ins string) []byte {
	out := make([]byte, 0, len(src)+len(ins))
	out = append(out, src[:at]...)
	out = append(out, ins...)
	return append(out, src[at:]...)
}

// hasOurEntry reports whether an asciiscapes entry is already registered for
// this event, so a second install is a no-op instead of a duplicate.
func hasOurEntry(src []byte, ev string) (bool, error) {
	hooksSpan, err := valueSpan(src, nil, "hooks")
	if err != nil || hooksSpan == nil {
		return false, err
	}
	evSpan, err := valueSpan(src, hooksSpan, ev)
	if err != nil || evSpan == nil {
		return false, err
	}
	return bytes.Contains(src[evSpan.start:evSpan.end], []byte(marker)), nil
}

// removeHooks strips our entries by splicing them out.
//
// The first version re-marshalled the parsed tree, which is much shorter code
// and was rejected by verifyPreserved for a good reason: MarshalIndent rewrites
// every key it touches, so uninstalling would have reformatted 747 permission
// rules and re-escaped the regexes inside the secret-expansion guard. Install
// and uninstall have to be equally careful or the safe one is pointless.
func removeHooks(src []byte) ([]byte, int, error) {
	hooksKey, hooksVal, err := memberSpan(src, nil, "hooks")
	if err != nil || hooksVal == nil {
		return src, 0, err
	}

	// Collect every deletion first, then apply them back to front so earlier
	// offsets stay valid.
	type del struct{ start, end int }
	var dels []del
	removed := 0

	var events []string
	var m map[string]json.RawMessage
	if err := json.Unmarshal(src[hooksVal.start:hooksVal.end], &m); err != nil {
		return nil, 0, err
	}
	for k := range m {
		events = append(events, k)
	}
	sort.Strings(events)

	emptied := 0
	for _, ev := range events {
		evKey, evVal, err := memberSpan(src, hooksVal, ev)
		if err != nil || evVal == nil {
			continue
		}
		elems, err := elementSpans(src, *evVal)
		if err != nil {
			continue
		}
		mine := 0
		for _, e := range elems {
			if bytes.Contains(src[e.start:e.end], []byte(marker)) {
				mine++
			}
		}
		if mine == 0 {
			continue
		}
		removed += mine
		if mine == len(elems) {
			// Nothing of the user's left in this event: take the whole key,
			// so uninstall does not leave `"SessionStart": []` behind.
			emptied++
			s, e := withComma(src, evKey.start, evVal.end)
			dels = append(dels, del{s, e})
			continue
		}
		for _, e := range elems {
			if bytes.Contains(src[e.start:e.end], []byte(marker)) {
				s, e2 := withComma(src, e.start, e.end)
				dels = append(dels, del{s, e2})
			}
		}
	}

	if removed == 0 {
		return src, 0, nil
	}

	// If every event we touched emptied the hooks object entirely, take that
	// too rather than leaving `"hooks": {}`.
	if emptied == len(events) {
		s, e := withComma(src, hooksKey.start, hooksVal.end)
		dels = []del{{s, e}}
	}

	sort.Slice(dels, func(i, j int) bool { return dels[i].start > dels[j].start })
	out := append([]byte(nil), src...)
	for _, d := range dels {
		if d.start < 0 || d.end > len(out) || d.start >= d.end {
			continue
		}
		out = append(out[:d.start], out[d.end:]...)
	}
	return out, removed, nil
}

// withComma widens a member or element span to swallow one separating comma
// and the whitespace that came with it, preferring the one before so the
// remaining members keep their indentation.
func withComma(src []byte, start, end int) (int, int) {
	// Back over whitespace to whatever precedes this member.
	s := start
	for s > 0 {
		c := src[s-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s--
			continue
		}
		break
	}
	if s > 0 && src[s-1] == ',' {
		return s - 1, end
	}

	// No comma before, so this was the first member. Forward over whitespace
	// to whatever follows.
	e := end
	for e < len(src) {
		c := src[e]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			e++
			continue
		}
		break
	}
	if e < len(src) && src[e] == ',' {
		return s, e + 1
	}

	// No comma on either side: the sole member. Take the whitespace on both
	// sides too, or `{"hooks": {...}}` uninstalls to `{\n}` instead of `{}`
	// and the round trip stops being a no-op.
	return s, e
}

// memberSpan returns the byte span of a whole `"key": value` member and of its
// value alone. Both use the decoder's offsets rather than a brace scan, which
// is the only way to be correct on a file full of shell commands containing
// braces and quotes.
func memberSpan(src []byte, parent *span, key string) (member *span, value *span, err error) {
	buf, base := src, 0
	if parent != nil {
		buf, base = src[parent.start:parent.end], parent.start
	}
	dec := json.NewDecoder(bytes.NewReader(buf))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("settings is not valid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, nil, fmt.Errorf("expected a JSON object")
	}
	for dec.More() {
		before := int(dec.InputOffset())
		kt, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}
		name, _ := kt.(string)
		afterKey := int(dec.InputOffset())

		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, nil, err
		}
		end := int(dec.InputOffset())
		if name != key {
			continue
		}
		// Walk forward from the key to the value's first byte.
		vs := afterKey
		for vs < end && (buf[vs] == ':' || buf[vs] == ' ' || buf[vs] == '\t' || buf[vs] == '\n' || buf[vs] == '\r') {
			vs++
		}
		// And back from `before` to the opening quote of the key.
		ks := before
		for ks < afterKey && buf[ks] != '"' {
			ks++
		}
		return &span{base + ks, base + end}, &span{base + vs, base + end}, nil
	}
	return nil, nil, nil
}

// elementSpans returns the byte span of every element of an array.
func elementSpans(src []byte, arr span) ([]span, error) {
	buf := src[arr.start:arr.end]
	dec := json.NewDecoder(bytes.NewReader(buf))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, fmt.Errorf("expected a JSON array")
	}
	var out []span
	for dec.More() {
		before := int(dec.InputOffset())
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := int(dec.InputOffset())
		s := before
		for s < end && (buf[s] == ',' || buf[s] == ' ' || buf[s] == '\t' || buf[s] == '\n' || buf[s] == '\r') {
			s++
		}
		out = append(out, span{arr.start + s, arr.start + end})
	}
	return out, nil
}

// writeSettings backs up, verifies, and replaces atomically.
func writeSettings(path string, orig, out []byte) error {
	// Refuse to write anything that is not valid JSON. This is the last gate
	// before touching a file that controls what the agent is allowed to run.
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(out, &probe); err != nil {
		return fmt.Errorf("refusing to write: result is not valid JSON: %w", err)
	}
	if err := verifyPreserved(orig, out); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "asciiscapes", "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	bak := filepath.Join(dir, "claude-settings."+stamp+".json")
	if err := os.WriteFile(bak, orig, 0o600); err != nil {
		return err
	}
	fmt.Println("backup    " + bak)

	// Temp file in the same directory so the rename is atomic: a crash mid
	// write must leave the old settings intact, not a half-written file that
	// Claude Code cannot parse.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".asciiscapes-settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// verifyPreserved proves the edit did not touch anything the user had.
//
// The first version of this check tried to prove additivity by striking our
// entries back out and comparing bytes. It refused to write -- correctly, but
// for an uninteresting reason: removing an entry cannot remove the "Stop": []
// key that adding it created, so the inverse was never exact. Worse, that
// inverse needed a brace scanner over a file full of shell commands, which is
// a lot of machinery to trust in the one place where being wrong destroys the
// user's permission allowlist and disarms their secret-expansion guard.
//
// So assert the property directly instead. Everything outside "hooks" must be
// byte-identical; every hook entry that was there before must still be there,
// unchanged and in order; and everything new must be ours.
func verifyPreserved(orig, out []byte) error {
	var a, b map[string]json.RawMessage
	if err := json.Unmarshal(orig, &a); err != nil {
		return fmt.Errorf("refusing to write: the original is not valid JSON: %w", err)
	}
	if err := json.Unmarshal(out, &b); err != nil {
		return fmt.Errorf("refusing to write: the result is not valid JSON: %w", err)
	}

	// Every other top-level key survives byte-for-byte. This is what catches
	// a re-marshal quietly reordering 747 permission rules or re-escaping the
	// regexes inside the PreToolUse guard.
	for k, av := range a {
		if k == "hooks" {
			continue
		}
		bv, ok := b[k]
		if !ok {
			return fmt.Errorf("refusing to write: the edit dropped the top-level key %q", k)
		}
		if !bytes.Equal(av, bv) {
			return fmt.Errorf("refusing to write: the edit rewrote the top-level key %q", k)
		}
	}

	ah, bh := map[string][]json.RawMessage{}, map[string][]json.RawMessage{}
	if raw, ok := a["hooks"]; ok {
		if err := json.Unmarshal(raw, &ah); err != nil {
			return fmt.Errorf("refusing to write: could not read the existing hooks: %w", err)
		}
	}
	if raw, ok := b["hooks"]; ok {
		if err := json.Unmarshal(raw, &bh); err != nil {
			return fmt.Errorf("refusing to write: could not read the resulting hooks: %w", err)
		}
	}

	// One rule covers install and uninstall both: an entry that is not ours
	// must survive, unchanged and in order. Install adds only marked entries,
	// so everything survives; uninstall removes only marked entries, so
	// everything the user wrote survives. Stating it once means the two
	// directions cannot drift apart, which is how the careful one ends up
	// paired with a careless one.
	for ev, was := range ah {
		now := bh[ev]
		i := 0
		for _, w := range was {
			if bytes.Contains(w, []byte(marker)) {
				continue // ours; may legitimately disappear
			}
			found := false
			for ; i < len(now); i++ {
				if bytes.Equal(canon(w), canon(now[i])) {
					found, i = true, i+1
					break
				}
			}
			if !found {
				return fmt.Errorf("refusing to write: an existing %s hook was changed or removed", ev)
			}
		}
	}

	// And nothing arrived that is not ours.
	for ev, now := range bh {
		for _, e := range now {
			if bytes.Contains(e, []byte(marker)) {
				continue
			}
			if !containsEqual(ah[ev], e) {
				return fmt.Errorf("refusing to write: the edit added a %s hook that is not an asciiscapes entry", ev)
			}
		}
	}
	return nil
}

// canon compacts a raw value so two spellings of the same JSON compare equal.
func canon(r json.RawMessage) []byte {
	var buf bytes.Buffer
	if err := json.Compact(&buf, r); err != nil {
		return r
	}
	return buf.Bytes()
}

func containsEqual(list []json.RawMessage, want json.RawMessage) bool {
	for _, x := range list {
		if bytes.Equal(canon(x), canon(want)) {
			return true
		}
	}
	return false
}

func changed(a, b []byte) bool { return !bytes.Equal(a, b) }

func die(err error) {
	fmt.Fprintln(os.Stderr, "asciiscapes:", err)
	os.Exit(1)
}
