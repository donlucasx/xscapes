package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const kimiFixture = `# Kimi Code CLI configuration
default_model = "kimi-code/kimi-for-coding"

[providers."managed:kimi-code"]
type = "kimi"
base_url = "https://example.invalid/v1"

[models."kimi-code/kimi-for-coding"]
provider = "managed:kimi-code"
model = "kimi-for-coding"
max_context_size = 262144

[thinking]
enabled = true
`

const hermesFixture = `# Hermes Agent configuration
model:
  default: "test-model"
hooks: {}
hooks_auto_accept: false
personalities: {}
security:
  approval_mode: ask
`

// The Kimi block appends at the end, comes out again byte-for-byte, and the
// file Kimi will load is one its own doctor accepts.
func TestKimiInstallIsAppendedValidatedAndReversible(t *testing.T) {
	bin := "/opt/xscapes/bin/xscapes"
	out, actions, err := addKimiHooks([]byte(kimiFixture), bin)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != len(kimiHookEvents) {
		t.Errorf("%d actions for %d events", len(actions), len(kimiHookEvents))
	}
	if !bytes.HasPrefix(out, []byte(kimiFixture)) {
		t.Errorf("the user's own file is not preserved verbatim at the top")
	}
	if got := bytes.Count(out, []byte("[[hooks]]")); got != len(kimiHookEvents) {
		t.Errorf("%d [[hooks]] tables, want %d", got, len(kimiHookEvents))
	}
	if !bytes.Contains(out, []byte(`command = "/opt/xscapes/bin/xscapes hook PreToolUse kimi"`)) {
		t.Errorf("the PreToolUse command is missing or misquoted:\n%s", out)
	}
	// Idempotent: a second install rewrites the block to the same bytes.
	again, _, err := addKimiHooks(out, bin)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, out) {
		t.Errorf("a second install changed the file:\n%s", again)
	}
	// A new binary path replaces the block rather than adding a second one.
	moved, _, _ := addKimiHooks(out, "/usr/local/bin/xscapes")
	if bytes.Contains(moved, []byte("/opt/xscapes")) || bytes.Count(moved, []byte(blockBegin)) != 1 {
		t.Errorf("reinstall with a new path left the old block behind")
	}
	// Reversible.
	back, n, err := removeKimiHooks(out)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(kimiHookEvents) {
		t.Errorf("uninstall counted %d entries, want %d", n, len(kimiHookEvents))
	}
	if string(back) != kimiFixture {
		t.Errorf("uninstall did not restore the original:\n%s", back)
	}
	// Refuses a file that already has the inline form.
	if _, _, err := addKimiHooks([]byte(kimiFixture+"hooks = []\n"), bin); err == nil {
		t.Errorf("an inline `hooks = [...]` was not refused")
	}

	// Kimi's own validator, when it is here. This is the one check that is
	// not self-consistency: the file is judged by the program that loads it.
	kimi, err := exec.LookPath("kimi")
	if err != nil {
		t.Skip("kimi not installed; the TOML is unvalidated by its consumer")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := exec.Command(kimi, "doctor", "config", path).CombinedOutput()
	if err != nil || !bytes.Contains(res, []byte("OK")) {
		t.Errorf("kimi doctor rejects the installed config (%v):\n%s", err, res)
	}
}

// The Hermes items go under each event, an existing user hook keeps its
// place with ours after it, `hooks: {}` opens into a block and closes back
// to `{}` on uninstall, and Hermes itself lists what was written.
func TestHermesInstallMergesAndIsReversible(t *testing.T) {
	bin := "/opt/xscapes/bin/xscapes"
	out, actions, err := addHermesHooks([]byte(hermesFixture), bin)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != len(hermesHookEvents) {
		t.Errorf("%d actions for %d events", len(actions), len(hermesHookEvents))
	}
	s := string(out)
	if strings.Contains(s, "hooks: {}") {
		t.Errorf("hooks: {} was not opened into a block")
	}
	if !strings.Contains(s, "hooks_auto_accept: false") {
		t.Errorf("hooks_auto_accept was touched; it is the user's switch")
	}
	for _, ev := range hermesHookEvents {
		if !strings.Contains(s, "  "+ev+":\n    - command: \"/opt/xscapes/bin/xscapes hook "+ev+" hermes\"  "+marker+"\n      timeout: 5\n") {
			t.Errorf("event %s is not written as expected:\n%s", ev, s)
		}
	}
	if !strings.HasSuffix(s, "security:\n  approval_mode: ask\n") {
		t.Errorf("the keys after hooks were disturbed:\n%s", s)
	}
	again, _, err := addHermesHooks(out, bin)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, out) {
		t.Errorf("a second install changed the file:\n%s", again)
	}
	back, n, err := removeHermesHooks(out)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(hermesHookEvents) {
		t.Errorf("uninstall counted %d items, want %d", n, len(hermesHookEvents))
	}
	if string(back) != hermesFixture {
		t.Errorf("uninstall did not restore the original:\n%s", back)
	}

	// A user's own hook under one of our events stays, and ours goes after it.
	withUser := strings.Replace(hermesFixture, "hooks: {}\n",
		"hooks:\n  pre_tool_call:\n    - matcher: \"terminal\"\n      command: \"/home/me/audit.sh\"\n      timeout: 10\n", 1)
	merged, _, err := addHermesHooks([]byte(withUser), bin)
	if err != nil {
		t.Fatal(err)
	}
	m := string(merged)
	if strings.Count(m, "  pre_tool_call:\n") != 1 {
		t.Errorf("pre_tool_call was duplicated:\n%s", m)
	}
	ui := strings.Index(m, "/home/me/audit.sh")
	oi := strings.Index(m, "xscapes hook pre_tool_call hermes")
	if ui < 0 || oi < 0 || oi < ui {
		t.Errorf("the user's hook should come first and ours after it:\n%s", m)
	}
	back2, _, err := removeHermesHooks(merged)
	if err != nil {
		t.Fatal(err)
	}
	if string(back2) != withUser {
		t.Errorf("uninstall did not leave the user's hook exactly as it was:\n%s", back2)
	}

	// No hooks key at all: one is added at the end.
	noKey := strings.Replace(hermesFixture, "hooks: {}\n", "", 1)
	added, _, err := addHermesHooks([]byte(noKey), bin)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(added), "\nhooks:\n  on_session_start:\n") {
		t.Errorf("a missing hooks key was not added:\n%s", added)
	}
	// A one-line mapping we cannot edit is refused, not mangled.
	if _, _, err := addHermesHooks([]byte(strings.Replace(hermesFixture, "hooks: {}", "hooks: {pre_tool_call: []}", 1)), bin); err == nil {
		t.Errorf("an inline hooks mapping was not refused")
	}

	// Hermes's own reading of the file, when it is here: HERMES_HOME points
	// it at a temp home holding only this config, and `hermes hooks list`
	// must list every event we wrote. A run that reports none would mean
	// the override is ignored -- and the check would say so, not pass.
	hermes, err := exec.LookPath("hermes")
	if err != nil {
		t.Skip("hermes not installed; the YAML is unvalidated by its consumer")
	}
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.yaml"), out, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(hermes, "hooks", "list")
	cmd.Env = append(os.Environ(), "HERMES_HOME="+home)
	res, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hermes hooks list failed (%v):\n%s", err, res)
	}
	for _, ev := range hermesHookEvents {
		if !bytes.Contains(res, []byte(ev)) {
			t.Errorf("hermes hooks list does not show %s:\n%s", ev, res)
		}
	}
}

// The first outside tester's Kimi config (2026-09-17) carried twelve
// `[[hooks]]` entries an agent had written by hand, each calling
// `xscapes hook <Event>` with no marker around them. Installing over them
// would fire every event twice. The installer refuses and says why; it does
// not remove what it did not write.
func TestInstallRefusesHooksWrittenByHand(t *testing.T) {
	hand := "default_model = \"kimi-code/k3\"\n\n" +
		"[[hooks]]\nevent = \"SessionStart\"\ncommand = \"/Users/h/go/bin/xscapes hook SessionStart\"\ntimeout = 5\n\n" +
		"[[hooks]]\nevent = \"Stop\"\ncommand = \"/Users/h/go/bin/xscapes hook Stop\"\ntimeout = 5\n"
	if _, _, err := addKimiHooks([]byte(hand), "/usr/local/bin/xscapes"); err == nil || !strings.Contains(err.Error(), "2 hook entries") {
		t.Fatalf("kimi: hand-written hooks were installed over: err=%v", err)
	}
	// Our own block, however many times, is fine: it is rewritten.
	own, _, err := addKimiHooks([]byte("default_model = \"x\"\n"), "/usr/local/bin/xscapes")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := addKimiHooks(own, "/usr/local/bin/xscapes"); err != nil {
		t.Fatalf("our own block was refused as foreign: %v", err)
	}
	handY := "hooks:\n  SessionStart:\n    - command: xscapes hook SessionStart hermes\n      timeout: 5\n"
	if _, _, err := addHermesHooks([]byte(handY), "/usr/local/bin/xscapes"); err == nil || !strings.Contains(err.Error(), "1 hook entry") {
		t.Fatalf("hermes: hand-written hooks were installed over: err=%v", err)
	}
}

// `xscapes install kimi|hermes --apply` keeps the config as it was, beside the
// Claude settings backups. Uninstall removes what was written; a copy of the
// original is what "restore" means, and the page has promised one.
func TestAgentInstallBacksUpTheOriginal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".kimi-code")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	orig := []byte("default_model = \"kimi-code/k3\"\n")
	if err := os.WriteFile(path, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	runInstallAgent("kimi", []string{"--apply", "--config", path, "--bin", "/usr/local/bin/xscapes"})
	baks, _ := filepath.Glob(filepath.Join(home, ".config", "xscapes", "backups", "kimi-config.*.toml"))
	if len(baks) != 1 {
		t.Fatalf("backups: %v, want one kimi-config.<stamp>.toml", baks)
	}
	if b, _ := os.ReadFile(baks[0]); !bytes.Equal(b, orig) {
		t.Fatalf("the backup is not the original:\n%s", b)
	}
	if now, _ := os.ReadFile(path); !bytes.Contains(now, []byte("[[hooks]]")) {
		t.Fatalf("the config was not written")
	}
}

// `xscapes kimi` (claude, hermes) refuses to start when that agent's hooks are
// not installed, because the scape would run half-blind on the output watcher
// and nothing on screen would say so. His first Kimi run, 2026-09-17: no
// owlets, no spend, no tool names, and the config had no hooks at all.
func TestTheLauncherKnowsWhenHooksAreMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, agent := range []string{"claude", "kimi", "hermes"} {
		if ok, _ := hooksInstalled(agent); ok {
			t.Fatalf("%s: installed with no config file at all", agent)
		}
	}
	// Kimi: a config without the block, then with it.
	kp := filepath.Join(home, ".kimi-code", "config.toml")
	os.MkdirAll(filepath.Dir(kp), 0o755)
	os.WriteFile(kp, []byte("default_model = \"x\"\n"), 0o644)
	if ok, where := hooksInstalled("kimi"); ok || where != kp {
		t.Fatalf("kimi without the block: ok=%v where=%q", ok, where)
	}
	out, _, err := addKimiHooks([]byte("default_model = \"x\"\n"), "/usr/local/bin/xscapes")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(kp, out, 0o644)
	if ok, _ := hooksInstalled("kimi"); !ok {
		t.Fatalf("kimi with the block: not seen")
	}
	// Hermes.
	hp := filepath.Join(home, ".hermes", "config.yaml")
	os.MkdirAll(filepath.Dir(hp), 0o755)
	os.WriteFile(hp, []byte("model: x\n"), 0o644)
	if ok, _ := hooksInstalled("hermes"); ok {
		t.Fatalf("hermes without items: seen")
	}
	hout, _, err := addHermesHooks([]byte("model: x\n"), "/usr/local/bin/xscapes")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(hp, hout, 0o644)
	if ok, _ := hooksInstalled("hermes"); !ok {
		t.Fatalf("hermes with items: not seen")
	}
	// Claude Code, through its own marker.
	cp := filepath.Join(home, ".claude", "settings.json")
	os.MkdirAll(filepath.Dir(cp), 0o755)
	os.WriteFile(cp, []byte("{}\n"), 0o644)
	if ok, _ := hooksInstalled("claude"); ok {
		t.Fatalf("claude with empty settings: seen")
	}
	cout, _, err := addHooks([]byte("{}\n"), "/usr/local/bin/xscapes")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(cp, cout, 0o644)
	if ok, _ := hooksInstalled("claude"); !ok {
		t.Fatalf("claude with the marked entries: not seen")
	}
	// The message names the agent, the install line and the way around.
	msg := missingHooks("kimi", kp)
	for _, want := range []string{"Kimi Code CLI", "xscapes install kimi --apply", "xscapes inside kimi", kp} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message lacks %q:\n%s", want, msg)
		}
	}
}

// TestKimiInstallHonoursKimiCodeHome: the installer and the launcher's check
// read Kimi's config where Kimi reads it. $KIMI_CODE_HOME relocates Kimi's
// home (the spend counter honoured it; the installer hardcoded ~/.kimi-code,
// so with the variable set the hooks went into a file Kimi never read and
// the launcher's check passed on that same file).
func TestKimiInstallHonoursKimiCodeHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", dir)
	path, err := adapters["kimi"].path()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "config.toml"); path != want {
		t.Fatalf("kimi config path %q, want %q", path, want)
	}
	if ok, where := hooksInstalled("kimi"); ok || where != path {
		t.Fatalf("with no config: installed %v at %q", ok, where)
	}
	if err := os.WriteFile(path, []byte(kimiFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	if ok, _ := hooksInstalled("kimi"); ok {
		t.Fatal("a config with no hooks counts as installed")
	}
	out, _, err := addKimiHooks([]byte(kimiFixture), "/opt/xscapes/bin/xscapes")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
	if ok, _ := hooksInstalled("kimi"); !ok {
		t.Fatal("the hooks written under KIMI_CODE_HOME are not seen by the launcher's check")
	}
}
