package event

import (
	"fmt"
	"testing"
	"time"
)

// TestABindReplaysWhatArrivedSinceTheLaunch: the hosted launcher binds to the
// session its agent announces, within its one-second poll, and the hooks
// that fire in that window spool. Listen starts at the spool's end on
// purpose (an hour of an unlistened session is history), so those lines were
// dropped: a Hermes one-shot's session start and its first prompt, six
// milliseconds apart, both spooled and the scape applied only the session
// end (his hooks log, 2026-09-18). ListenSince keeps the rule for everything
// before the launch and takes what came after it as news.
func TestABindReplaysWhatArrivedSinceTheLaunch(t *testing.T) {
	tmpHome(t)
	const s = "hermes-oneshot"
	launched := time.Now()
	old := Event{Session: s, Kind: ToolStart, Tool: "old", TS: launched.Add(-time.Hour).UnixMilli()}
	start := Event{Session: s, Kind: SessionStart, TS: launched.Add(2 * time.Millisecond).UnixMilli()}
	prompt := Event{Session: s, Kind: Prompt, Text: "reply ok", TS: launched.Add(8 * time.Millisecond).UnixMilli()}
	for _, e := range []Event{old, start, prompt} {
		if via, err := Emit(e); err != nil || via {
			t.Fatalf("emit %s: via socket %v, err %v (no listener yet: it must spool)", e.Kind, via, err)
		}
	}
	b, err := ListenSince(s, launched)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	var got []Kind
	deadline := time.After(2 * time.Second)
loop:
	for {
		select {
		case e := <-b.C:
			got = append(got, e.Kind)
			if len(got) == 2 {
				// A third would be the hour-old line replayed: give it a
				// moment to show up before calling the count final.
				select {
				case e := <-b.C:
					got = append(got, e.Kind)
				case <-time.After(400 * time.Millisecond):
				}
				break loop
			}
		case <-deadline:
			break loop
		}
	}
	want := []Kind{SessionStart, Prompt}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("delivered %v, want %v: the two lines since the launch, not the hour before", got, want)
	}
}
