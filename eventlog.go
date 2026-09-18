package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// logPath resolves XSCAPES_HOOKLOG and XSCAPES_EVENTLOG: a path as given, or
// for 1, true or yes a file of the given name under xscapes's home, the way
// XSCAPES_TRACE=1 picks its own path. The value was taken as a path whatever
// it said, and =1 wrote a file named "1" into the agent's working directory,
// which for the hook is the user's project (Kimi's assessment, 2026-09-18).
func logPath(v, name string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return ""
	case "1", "true", "yes":
		home, err := event.Home()
		if err != nil {
			return ""
		}
		return filepath.Join(home, name)
	}
	return v
}

func appendEventLog(path string, e event.Event, now time.Time) {
	writeEventLogLine(path, map[string]interface{}{
		"ts": now.UnixMilli(), "kind": e.Kind, "tool": e.Tool, "op": e.Op,
		"agent": e.Agent, "src": e.Src, "session": e.Session, "id": e.ID,
	})
}

// appendEventLogStats writes the counts whenever one changes: events the bus
// dropped or could not parse, and erase-displays the agent sent that the host
// confined to its band (or dropped), so a blank row is attributed to a clear
// rather than inferred from a screenshot.
func appendEventLogStats(path string, dropped, bad, erased int64, now time.Time) {
	writeEventLogLine(path, map[string]interface{}{
		"ts": now.UnixMilli(), "dropped": dropped, "bad": bad, "erased": erased,
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
