package spend

import (
	"os"
	"path/filepath"
	"testing"
)

// Records as Kimi Code CLI 0.39.1 writes them (2026-09-17), one wire per agent.
const kimiMain = `{"type":"metadata","protocol_version":"1.5"}
{"type":"llm.request","agentId":"main","kind":"loop","maxTokens":1048576,"turnStep":"0.1"}
{"type":"usage.record","agentId":"main","model":"kimi-code/k3","usage":{"inputOther":2067,"output":69,"inputCacheRead":19200,"inputCacheCreation":0},"usageScope":"turn"}
{"type":"token_counting.measured","agentId":"main","length":4,"tokens":21336}
{"type":"context.append_loop_event","agentId":"main","event":{"type":"step.end","usage":{"inputOther":2067,"output":69,"inputCacheRead":19200,"inputCacheCreation":0}}}
{"type":"usage.record","agentId":"main","model":"kimi-code/k3","usage":{"inputOther":118,"output":114,"inputCacheRead":21248,"inputCacheCreation":0},"usageScope":"turn"}
{"type":"token_counting.measured","agentId":"main","length":6,"tokens":21480}
`
const kimiSub = `{"type":"llm.request","agentId":"agent-0","maxTokens":262144}
{"type":"usage.record","agentId":"agent-0","usage":{"inputOther":1590,"output":56,"inputCacheRead":7680,"inputCacheCreation":0},"usageScope":"turn"}
{"type":"token_counting.measured","agentId":"agent-0","length":2,"tokens":9000}
`

func TestKimiWiresAreSummedAndTheContextRead(t *testing.T) {
	dir := t.TempDir()
	for agent, body := range map[string]string{"main": kimiMain, "agent-0": kimiSub} {
		d := filepath.Join(dir, "agents", agent)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "wire.jsonl"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tl := New()
	tl.Poll(dir)
	// usage.record only: the step.end copy of the same usage is not counted.
	want := int64(2067+69+19200) + int64(118+114+21248) + int64(1590+56+7680)
	if got := tl.Total(); got != want {
		t.Fatalf("total %d, want %d", got, want)
	}
	used, window, ok := tl.Context()
	if !ok || used != 21480 || window != 1048576 {
		// The MAIN agent's last measurement and window; the subagent's are its own.
		t.Fatalf("context = %d/%d ok=%v, want 21480/1048576", used, window, ok)
	}
	// A poll with nothing new changes nothing; a new step adds.
	tl.Poll(dir)
	if tl.Total() != want {
		t.Fatalf("a second poll changed the total")
	}
	f, _ := os.OpenFile(filepath.Join(dir, "agents", "main", "wire.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"type":"usage.record","agentId":"main","usage":{"inputOther":1,"output":2,"inputCacheRead":3,"inputCacheCreation":4}}` + "\n")
	f.Close()
	tl.Poll(dir)
	if tl.Total() != want+10 {
		t.Fatalf("after a new record: %d, want %d", tl.Total(), want+10)
	}
}

func TestKimiSessionDirComesFromTheIndexLastEntryWins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", home)
	old := filepath.Join(home, "sessions", "wd_x", "session_a")
	cur := filepath.Join(home, "sessions", "wd_y", "session_a")
	for _, d := range []string{old, cur} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	idx := `{"sessionId":"session_a","sessionDir":"` + old + `","workDir":"/x"}` + "\n" +
		`{"sessionId":"session_b","sessionDir":"/nowhere","workDir":"/y"}` + "\n" +
		`{"sessionId":"session_a","sessionDir":"` + cur + `","workDir":"/y"}` + "\n"
	if err := os.WriteFile(filepath.Join(home, "session_index.jsonl"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := KimiSessionDir("session_a"); got != cur {
		t.Fatalf("KimiSessionDir = %q, want %q", got, cur)
	}
	if got := KimiSessionDir("session_b"); got != "" {
		t.Fatalf("a directory that does not exist was returned: %q", got)
	}
	if got := KimiSessionDir("session_none"); got != "" {
		t.Fatalf("an unknown session was resolved: %q", got)
	}
}
