package event

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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

// What a listening engine takes in over its socket is written to its
// receive log; the spool never sees it.
func TestTheReceiveLogHoldsWhatTheSocketDelivered(t *testing.T) {
	if runtime.GOOS == "windows" {
		// The socket is the fast path and the spool is the fallback; Go's net
		// has no AF_UNIX on Windows, so this build has only the fallback and
		// "did it go over the socket" is a question with one answer.
		t.Skip("no unix datagram sockets on windows; every event takes the spool")
	}

	tmpHome(t)
	const s = "recv-log"
	b, err := Listen(s)
	if err != nil {
		t.Fatal(err)
	}
	if via, err := Emit(Event{Session: s, Kind: SubStart, Agent: "kit-1", TS: time.Now().UnixMilli()}); err != nil || !via {
		t.Fatalf("emit: via socket %v, err %v", via, err)
	}
	select {
	case <-b.C:
	case <-time.After(2 * time.Second):
		t.Fatal("nothing delivered")
	}
	b.Close()
	rp, _ := RecvPath(s)
	got, _ := os.ReadFile(rp)
	if !strings.Contains(string(got), `"kit-1"`) {
		t.Fatalf("receive log: %q", got)
	}
	sp, _ := SpoolPath(s)
	if sp, _ := os.ReadFile(sp); strings.Contains(string(sp), "kit-1") {
		t.Fatal("a socket delivery also landed in the spool")
	}
}
