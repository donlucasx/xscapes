package main

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// TestFoldTheEventLog folds a scape's event log back through the reducer and
// prints the activity level over the session: when the wave was, how high it
// went, how long it held. His question of 2026-09-18: did the five-agent
// wave in the traced Kimi session reach the level his earlier flicker was
// seen at? The log's ts is the frame that applied the event, which is what
// the scape drew from.
//
//	XSCAPES_EVENTLOG_REPLAY=~/.config/xscapes/kimi-events.jsonl go test -run TestFoldTheEventLog -v .
func TestFoldTheEventLog(t *testing.T) {
	path := os.Getenv("XSCAPES_EVENTLOG_REPLAY")
	if path == "" {
		t.Skip("set XSCAPES_EVENTLOG_REPLAY to an event log to fold it")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var evs []event.Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 8<<20)
	for sc.Scan() {
		var e event.Event
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Kind == "" || e.TS == 0 {
			continue // a stats line, or noise
		}
		evs = append(evs, e)
	}
	if len(evs) == 0 {
		t.Fatal("no events in the log")
	}
	r := reduce.New("fold")
	t0, t1 := time.UnixMilli(evs[0].TS), time.UnixMilli(evs[len(evs)-1].TS)
	i, above, maxL := 0, 0, 0.0
	var maxAt time.Time
	kinds := map[event.Kind]int{}
	for _, e := range evs {
		kinds[e.Kind]++
	}
	for now := t0; !now.After(t1.Add(30 * time.Second)); now = now.Add(time.Second) {
		for i < len(evs) && !time.UnixMilli(evs[i].TS).After(now) {
			r.Apply(evs[i], time.UnixMilli(evs[i].TS))
			i++
		}
		st := r.State(now)
		if st.Act.Level > maxL {
			maxL, maxAt = st.Act.Level, now
		}
		if st.Act.Level >= 0.8 {
			above++
		}
		if int(now.Sub(t0).Seconds())%60 == 0 {
			t.Logf("%s  level %.2f  owlets %d  pose %v", now.Local().Format("15:04:05"), st.Act.Level, st.Kittens, st.Pose)
		}
	}
	t.Logf("%d events (%v), %s to %s", len(evs), kinds, t0.Local().Format("15:04:05"), t1.Local().Format("15:04:05"))
	t.Logf("max level %.2f at %s; %d seconds at or above 0.8", maxL, maxAt.Local().Format("15:04:05"), above)
}
