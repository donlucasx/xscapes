package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
)

// XSCAPES_EVENTLOG=<file>: the scape appends every event its reducer applies,
// one JSON line each, and a line whenever the bus reports a drop. With the
// hook command's own log (XSCAPES_HOOKLOG) on the other end, a lost event
// can be placed: before the socket or after it. Written for his 2026-09-17
// swarm, where six sub-agent starts were in the hook log and no owlet was
// on the scape, and nothing recorded which side had lost them. Errors are
// swallowed: a diagnostic must never cost a frame.

func appendEventLog(path string, e event.Event, now time.Time) {
	writeEventLogLine(path, map[string]interface{}{
		"ts": now.UnixMilli(), "kind": e.Kind, "tool": e.Tool, "op": e.Op,
		"agent": e.Agent, "src": e.Src, "session": e.Session, "id": e.ID,
	})
}

func appendEventLogStats(path string, dropped, bad int64, now time.Time) {
	writeEventLogLine(path, map[string]interface{}{
		"ts": now.UnixMilli(), "dropped": dropped, "bad": bad,
	})
}

func writeEventLogLine(path string, m map[string]interface{}) {
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	line, err := json.Marshal(m)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
}
