package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// A settings file shaped like a real one: a big unrelated block, an existing
// hook we must not touch, and awkward characters inside a command string.
const sample = `{
  "permissions": {
    "allow": [
      "Bash(awk -F',' '{gsub(/[^0-9.-]/,\"\",$4)}')",
      "Bash(grep -qc VERCEL_TOKEN=)"
    ]
  },
  "model": "opus",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "cmd=$(jq -r '.tool_input.command // \"\"'); if printf '%s' \"$cmd\" | grep -qE '\\$\\{?[#!]?(VERCEL_TOKEN)'; then exit 2; fi",
            "timeout": 10
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "afplay /System/Library/Sounds/Funk.aiff & disown",
            "async": true
          }
        ]
      }
    ]
  },
  "statusLine": { "type": "command", "command": "~/.claude/statusline.sh" }
}
`

func TestInstallIsPurelyAdditive(t *testing.T) {
	out, actions, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != len(hookEvents) {
		t.Fatalf("got %d actions for %d events", len(actions), len(hookEvents))
	}
	if err := verifyPreserved([]byte(sample), out); err != nil {
		t.Fatalf("install failed its own safety check: %v", err)
	}

	var a, b map[string]json.RawMessage
	if err := json.Unmarshal([]byte(sample), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatalf("install produced invalid JSON: %v", err)
	}
	// The blocks that matter must come through untouched, byte for byte --
	// not merely equal after a round trip, which would hide a re-escape.
	for _, k := range []string{"permissions", "model", "statusLine"} {
		if !bytes.Equal(a[k], b[k]) {
			t.Errorf("install rewrote %q:\n before: %s\n  after: %s", k, a[k], b[k])
		}
	}
}

func TestExistingHooksSurvive(t *testing.T) {
	out, _, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	// The secret-expansion guard is the reason this whole file is careful.
	// Compare the DECODED command, not the raw bytes: the guard is full of
	// backslashes and comparing escaped forms tests the test, not the code.
	var before, after struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(sample), &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &after); err != nil {
		t.Fatal(err)
	}
	want := before.Hooks["PreToolUse"][0].Hooks[0].Command
	found := false
	for _, g := range after.Hooks["PreToolUse"] {
		for _, h := range g.Hooks {
			if h.Command == want {
				found = true
			}
		}
	}
	if !found {
		t.Error("the PreToolUse secret guard did not survive install intact")
	}
	if !bytes.Contains(out, []byte("afplay /System/Library/Sounds/Funk.aiff")) {
		t.Error("the existing Stop hook did not survive install")
	}

	var top struct {
		Hooks map[string][]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(out, &top); err != nil {
		t.Fatal(err)
	}
	if got := len(top.Hooks["PreToolUse"]); got != 2 {
		t.Errorf("PreToolUse has %d entries, want the original plus ours", got)
	}
	if got := len(top.Hooks["Stop"]); got != 2 {
		t.Errorf("Stop has %d entries, want the original plus ours", got)
	}
}

func TestInstallUninstallIsAByteForByteNoOp(t *testing.T) {
	installed, _, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	back, n, err := removeHooks(installed)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(hookEvents) {
		t.Errorf("uninstall found %d entries, want %d", n, len(hookEvents))
	}
	if err := verifyPreserved(installed, back); err != nil {
		t.Fatalf("uninstall failed the safety check: %v", err)
	}
	if !bytes.Equal(back, []byte(sample)) {
		t.Errorf("round trip is not a no-op:\n%s", firstDiff([]byte(sample), back))
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	once, _, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	twice, actions, err := addHooks(once, "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(once, twice) {
		t.Error("a second install changed the file; it should be a no-op")
	}
	for _, a := range actions {
		if !strings.HasPrefix(a, "skip") {
			t.Errorf("second install did something other than skip: %q", a)
		}
	}
}

// The safety check is the only thing standing between a bug in the splice and
// a destroyed permission allowlist. Prove it actually fires.
func TestSafetyCheckCatchesDamage(t *testing.T) {
	cases := map[string]string{
		"a dropped top-level key":   strings.Replace(sample, `"model": "opus",`, "", 1),
		"a rewritten permissions":   strings.Replace(sample, `"Bash(grep -qc VERCEL_TOKEN=)"`, `"Bash(anything)"`, 1),
		"a removed existing hook":   strings.Replace(sample, "afplay /System/Library/Sounds/Funk.aiff & disown", "rm -rf /", 1),
		"a foreign hook slipped in": strings.Replace(sample, `"model": "opus"`, `"model": "sonnet"`, 1),
	}
	for name, damaged := range cases {
		if err := verifyPreserved([]byte(sample), []byte(damaged)); err == nil {
			t.Errorf("the safety check passed %s; it must refuse", name)
		}
	}
	// And it must not cry wolf on the real, correct edit.
	good, _, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyPreserved([]byte(sample), good); err != nil {
		t.Errorf("the safety check refused a correct install: %v", err)
	}
}

func TestWorksOnAFileWithNoHooksAtAll(t *testing.T) {
	for _, src := range []string{`{}`, `{"model":"opus"}`, "{\n  \"model\": \"opus\"\n}\n"} {
		out, _, err := addHooks([]byte(src), "/usr/local/bin/asciiscapes")
		if err != nil {
			t.Fatalf("%q: %v", src, err)
		}
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(out, &probe); err != nil {
			t.Fatalf("%q produced invalid JSON: %v\n%s", src, err, out)
		}
		if err := verifyPreserved([]byte(src), out); err != nil {
			t.Errorf("%q: %v", src, err)
		}
		back, _, err := removeHooks(out)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(back, []byte(src)) {
			t.Errorf("%q did not round trip:\n%s", src, back)
		}
	}
}

// The recorded path ends up inside a shell command inside a JSON string.
func TestRefusesAShellMetacharacterInTheBinaryPath(t *testing.T) {
	for _, bad := range []string{"/tmp/a`whoami`/asciiscapes", "/tmp/a$HOME/x", `/tmp/a"b/x`} {
		if _, err := binPath(bad); err == nil {
			t.Errorf("binPath(%q) was accepted", bad)
		}
	}
	if _, err := binPath("/usr/local/bin/asciiscapes"); err != nil {
		t.Errorf("a normal path was refused: %v", err)
	}
}

// The command must be readable by the person asked to approve it.
func TestCommandIsNotHTMLEscaped(t *testing.T) {
	out, _, err := addHooks([]byte(sample), "/usr/local/bin/asciiscapes")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte(`\u003e`)) || bytes.Contains(out, []byte(`\u0026`)) {
		t.Error("the hook command is HTML-escaped; >/dev/null 2>&1 should be legible")
	}
	if !bytes.Contains(out, []byte(`>/dev/null 2>&1 || true # asciiscapes:v1`)) {
		t.Error("the hook command does not have the expected shape")
	}
}

func firstDiff(a, b []byte) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			lo := i - 60
			if lo < 0 {
				lo = 0
			}
			return "at byte " + itoa(i) + ":\n want ..." + string(a[lo:min(len(a), i+60)]) +
				"\n  got ..." + string(b[lo:min(len(b), i+60)])
		}
	}
	return "lengths differ: " + itoa(len(a)) + " vs " + itoa(len(b))
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b []byte
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
