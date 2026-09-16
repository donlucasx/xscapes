package spend

import (
	"os"
	"path/filepath"
	"testing"
)

const line1 = `{"type":"assistant","requestId":"req_1","message":{"id":"msg_1","usage":{"input_tokens":2,"cache_creation_input_tokens":100,"cache_read_input_tokens":1000,"output_tokens":30}}}` + "\n"
const line2 = `{"type":"assistant","requestId":"req_2","message":{"id":"msg_2","usage":{"input_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":2000,"output_tokens":40}}}` + "\n"
const user = `{"type":"user","message":{"role":"user","content":"hi"}}` + "\n"
const subLine = `{"type":"assistant","isSidechain":true,"agentId":"a1","requestId":"req_3","message":{"id":"msg_3","usage":{"input_tokens":2,"cache_creation_input_tokens":100,"cache_read_input_tokens":1000,"output_tokens":30}}}` + "\n"

// TestTheTallyCountsEachResponseOnce: a streamed message written three
// times with the same usage counts once; a user line counts nothing; a line
// cut mid-write is finished on the next poll; a subagent's transcript that
// appears later is added (its own message ids: a subagent's response is a
// response of its own); and a poll with nothing new changes nothing.
func TestTheTallyCountsEachResponseOnce(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "sess.jsonl")
	if err := os.WriteFile(main, []byte(line1+line1+user+line1), 0o644); err != nil {
		t.Fatal(err)
	}
	tl := New()
	tl.Poll(main)
	if got := tl.Total(); got != 1132 {
		t.Fatalf("three copies of one response: total %d, want 1132", got)
	}
	// Half a line, then the rest.
	f, _ := os.OpenFile(main, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(line2[:40])
	f.Close()
	tl.Poll(main)
	if got := tl.Total(); got != 1132 {
		t.Fatalf("a half-written line counted: %d", got)
	}
	f, _ = os.OpenFile(main, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(line2[40:])
	f.Close()
	tl.Poll(main)
	if got := tl.Total(); got != 1132+2045 {
		t.Fatalf("the finished line: %d, want %d", got, 1132+2045)
	}
	// A subagent under the session's directory.
	sub := filepath.Join(dir, "sess", "subagents", "workflows", "wf_1")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "agent-a1.jsonl"), []byte(subLine), 0o644)
	tl.Poll(main)
	if got := tl.Total(); got != 1132+2045+1132 {
		t.Fatalf("with the subagent: %d, want %d", got, 1132+2045+1132)
	}
	tl.Poll(main)
	if got := tl.Total(); got != 1132+2045+1132 {
		t.Fatalf("nothing new, yet the total moved: %d", got)
	}
	// No transcript yet is no figure, not an error.
	fresh := New()
	fresh.Poll(filepath.Join(dir, "nothing.jsonl"))
	if fresh.Total() != 0 {
		t.Fatal("a missing transcript counted something")
	}
}

// TestTheTallyOnHisOwnTranscript reads a real session's transcript when
// XSCAPES_SPEND_TRANSCRIPT names one and prints the sum, so the figure can
// be checked against an independent count.
func TestTheTallyOnHisOwnTranscript(t *testing.T) {
	p := os.Getenv("XSCAPES_SPEND_TRANSCRIPT")
	if p == "" {
		t.Skip("set XSCAPES_SPEND_TRANSCRIPT=<file>")
	}
	tl := New()
	tl.Poll(p)
	t.Logf("%s: %d tokens over %d responses", p, tl.Total(), len(tl.seen))
}
