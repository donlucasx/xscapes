package notify

import "testing"

// The nag is the case this exists for. Claude Code re-sends the same
// notification every sixty seconds while a permission prompt goes unanswered;
// one sound per prompt is a nudge, one a minute is an alarm clock.
func TestARepeatedNagSoundsOnce(t *testing.T) {
	var k Knocker
	k.Knock("", false) // attach on an idle session
	if _, ok := k.Knock("allow Bash?", true); !ok {
		t.Fatal("the first ask must sound")
	}
	for i := 0; i < 5; i++ {
		if _, ok := k.Knock("allow Bash?", true); ok {
			t.Errorf("nag %d sounded again", i+1)
		}
	}
}

// A scape can attach to a session that is ALREADY waiting on the user. It must
// not shout about something that happened before it existed.
func TestAttachingToAWaitingSessionIsSilent(t *testing.T) {
	var k Knocker
	if _, ok := k.Knock("allow Bash?", true); ok {
		t.Error("attaching mid-prompt sounded a knock for history")
	}
	// But the next, genuinely new question does sound.
	if _, ok := k.Knock("allow Read?", true); !ok {
		t.Error("a new question after attaching must sound")
	}
}

// The two knocks are locked as distinct cues, and a distinction that exists
// only on screen is no distinction to someone looking at another pane.
func TestAskAndDoneSoundDifferently(t *testing.T) {
	var k Knocker
	k.Knock("", false)

	got, ok := k.Knock("allow Bash?", true)
	if !ok || got != Ask {
		t.Errorf("needs_input gave (%v, %v), want (ask, true)", got, ok)
	}
	// The same text arriving as a finish is a different event: the agent
	// stopped asking and started reporting.
	got, ok = k.Knock("allow Bash?", false)
	if !ok || got != Done {
		t.Errorf("the flip to a finish gave (%v, %v), want (done, true)", got, ok)
	}
}

func TestBubbleClearingIsSilentAndRearms(t *testing.T) {
	var k Knocker
	k.Knock("", false)
	if _, ok := k.Knock("all done", false); !ok {
		t.Fatal("expected the finish knock")
	}
	if _, ok := k.Knock("", false); ok {
		t.Error("a bubble expiring must not sound")
	}
	// The same message returning after a gap is a new knock.
	if _, ok := k.Knock("all done", false); !ok {
		t.Error("the same message after a clear is a new knock and must sound")
	}
}

// Two different questions back to back are two questions.
func TestADifferentQuestionSoundsAgain(t *testing.T) {
	var k Knocker
	k.Knock("", false)
	k.Knock("allow Bash?", true)
	if _, ok := k.Knock("allow WebFetch?", true); !ok {
		t.Error("a second, different question must sound")
	}
}

// Muting has to actually mute, and the flag for it is an env var so it can be
// set for one run without editing anything.
func TestSilentEnvMutes(t *testing.T) {
	t.Setenv(SilentEnv, "1")
	p := New()
	if !p.Silent() {
		t.Errorf("player is %q with %s set, want silent", p.Describe(), SilentEnv)
	}
	p.Play(Ask) // must not panic or write anything
}

// A player with no sound files and no binary still has to be safe to call.
func TestZeroPlayerIsSafe(t *testing.T) {
	var p *Player
	p.Play(Ask)
	if got := p.Describe(); got != "silent" {
		t.Errorf("nil player describes as %q", got)
	}
	p2 := &Player{}
	p2.Play(Done)
	if !p2.Silent() {
		t.Error("the zero player should report silent")
	}
}
