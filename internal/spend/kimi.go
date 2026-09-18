package spend

// Kimi Code CLI keeps a session's transcript as a directory rather than a
// file: sessions/wd_<slug>_<hash>/session_<id>/agents/<agent>/wire.jsonl,
// one wire per agent ("main", then "agent-0", "agent-1" for subagents), and
// ~/.kimi-code/session_index.jsonl maps a session id to that directory.
// Its hooks carry no transcript path, so the hook command resolves the
// directory from the session id and the composer polls it here.
//
// What is read, all measured on 0.39.1 (2026-09-17):
//
//	{"type":"usage.record","agentId":"main","usage":{"inputOther":2067,
//	 "output":69,"inputCacheRead":19200,"inputCacheCreation":0}, ...}
//	{"type":"token_counting.measured","agentId":"main","tokens":21336, ...}
//	{"type":"llm.request","agentId":"main","maxTokens":1048576, ...}
//
// One usage.record per LLM step, all four fields counted as in the Claude
// transcript; the last measured token count of the MAIN agent is the context
// in use and llm.request's maxTokens is the window, which gives the moon a
// reading without touching Kimi's status line. Kimi's own documentation
// calls the wire "local debug materials", so every line here is read on
// sufferance: an unknown shape counts nothing and the display shows nothing,
// never a wrong number and never a crash.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// KimiHome is Kimi's data directory: $KIMI_CODE_HOME, else ~/.kimi-code.
func KimiHome() string {
	if h := strings.TrimSpace(os.Getenv("KIMI_CODE_HOME")); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kimi-code")
}

// KimiSessionDir finds the session's directory through session_index.jsonl,
// last entry wins. Empty when the index has no such session yet: the index
// is written by Kimi a moment after SessionStart fires, so the first hook
// of a session may come up empty and a later one finds it.
func KimiSessionDir(session string) string {
	if session == "" {
		return ""
	}
	f, err := os.Open(filepath.Join(KimiHome(), "session_index.jsonl"))
	if err != nil {
		return ""
	}
	defer f.Close()
	var dir string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var l struct {
			ID  string `json:"sessionId"`
			Dir string `json:"sessionDir"`
		}
		if json.Unmarshal(sc.Bytes(), &l) == nil && l.ID == session && l.Dir != "" {
			dir = l.Dir
		}
	}
	if dir == "" {
		return ""
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return ""
	}
	return dir
}

// pollKimi reads every agent's wire under the session directory.
func (t *Tally) pollKimi(dir string) {
	_ = filepath.WalkDir(filepath.Join(dir, "agents"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Base(p) == "wire.jsonl" {
			t.read(p, t.kimiLine)
		}
		return nil
	})
}

// kimiLine counts one wire record.
func (t *Tally) kimiLine(b []byte) {
	if !bytes.Contains(b, []byte(`"type":"`)) {
		return
	}
	var l struct {
		Type    string `json:"type"`
		AgentID string `json:"agentId"`
		Usage   *struct {
			InputOther    int64 `json:"inputOther"`
			Output        int64 `json:"output"`
			CacheRead     int64 `json:"inputCacheRead"`
			CacheCreation int64 `json:"inputCacheCreation"`
		} `json:"usage"`
		Tokens    int64 `json:"tokens"`
		MaxTokens int   `json:"maxTokens"`
	}
	if json.Unmarshal(b, &l) != nil {
		return
	}
	switch l.Type {
	case "usage.record":
		if l.Usage != nil {
			t.total += l.Usage.InputOther + l.Usage.Output + l.Usage.CacheRead + l.Usage.CacheCreation
		}
	case "token_counting.measured":
		if l.AgentID == "main" && l.Tokens > 0 {
			t.ctxUsed = l.Tokens
		}
	case "llm.request":
		if l.AgentID == "main" && l.MaxTokens > 0 {
			t.ctxWindow = l.MaxTokens
		}
	}
}

// Context is the context in use and the window, when the transcript says.
func (t *Tally) Context() (used int64, window int, ok bool) {
	if t.ctxWindow <= 0 {
		return 0, 0, false
	}
	return t.ctxUsed, t.ctxWindow, true
}
