// Package spend sums the tokens an agent's transcripts record, for the
// counter in the top-right corner of every scape (his ask of 2026-09-16).
//
// Claude Code writes a JSON line per message to the session's transcript,
// and every assistant line carries the API's usage for that response. A
// streamed message is written once per content block with the SAME usage
// (up to six times in one measured session), so a response is counted once
// per (message id, request id). Subagents write their own transcripts under
// the session's directory. All four usage fields count: fresh input, cache
// writes, cache reads and output. A token processed is a token spent,
// whatever it was billed at; in the measured session cache reads were 97%
// of the total, which is what makes "/1M" the wrong denominator.
package spend

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Tally is the running sum over a main transcript and its subagents'.
type Tally struct {
	files map[string]*tail
	seen  map[string]struct{}
	total int64
	// ctxUsed and ctxWindow come only from a transcript that records them
	// (Kimi's wire, kimi.go); the Claude transcript leaves them zero and the
	// moon reads the status line instead.
	ctxUsed   int64
	ctxWindow int
}

type tail struct {
	off  int64
	rest []byte
}

// New is an empty tally.
func New() *Tally {
	return &Tally{files: map[string]*tail{}, seen: map[string]struct{}{}}
}

// Total is the tokens counted so far.
func (t *Tally) Total() int64 { return t.total }

// Poll reads what the main transcript and any subagent transcripts beside
// it have written since the last poll: bytes past each file's last offset
// only, so a poll every couple of seconds costs nothing on an idle session.
// A missing file is a session that has not written yet. A DIRECTORY is a
// Kimi session (kimi.go).
func (t *Tally) Poll(main string) {
	if main == "" {
		return
	}
	if st, err := os.Stat(main); err == nil && st.IsDir() {
		t.pollKimi(main)
		return
	}
	t.read(main, t.line)
	dir := strings.TrimSuffix(main, ".jsonl") + "/subagents"
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".jsonl") {
			t.read(p, t.line)
		}
		return nil
	})
}

// read feeds the bytes a file has gained since the last poll, line by line,
// to the format's line reader.
func (t *Tally) read(path string, line func([]byte)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	st := t.files[path]
	if st == nil {
		st = &tail{}
		t.files[path] = st
	}
	if _, err := f.Seek(st.off, io.SeekStart); err != nil {
		return
	}
	b, err := io.ReadAll(f)
	if err != nil && len(b) == 0 {
		return
	}
	st.off += int64(len(b))
	b = append(st.rest, b...)
	for {
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			break
		}
		line(b[:i])
		b = b[i+1:]
	}
	st.rest = append([]byte(nil), b...)
}

// line counts one transcript line if it is an assistant message with usage
// not seen before.
func (t *Tally) line(b []byte) {
	if !bytes.Contains(b, []byte(`"usage"`)) {
		return
	}
	var l struct {
		Type      string `json:"type"`
		RequestID string `json:"requestId"`
		Message   struct {
			ID    string `json:"id"`
			Usage *struct {
				Input  int64 `json:"input_tokens"`
				CacheW int64 `json:"cache_creation_input_tokens"`
				CacheR int64 `json:"cache_read_input_tokens"`
				Output int64 `json:"output_tokens"`
			} `json:"usage"`
		} `json:"message"`
	}
	if json.Unmarshal(b, &l) != nil || l.Type != "assistant" || l.Message.Usage == nil {
		return
	}
	key := l.Message.ID + "/" + l.RequestID
	if _, dup := t.seen[key]; dup {
		return
	}
	t.seen[key] = struct{}{}
	u := l.Message.Usage
	t.total += u.Input + u.CacheW + u.CacheR + u.Output
}
