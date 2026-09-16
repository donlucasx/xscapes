package notify

import "testing"

// The nag is the case this exists for. Claude Code re-sends the same
// notification every sixty seconds while a permission prompt goes unanswered;
// one sound per prompt is a nudge, one a minute is an alarm clock.
func TestARepeatedNagSoundsOnce(t *testing.T) {
	var k Knocker
	k.Knock("", false, false) // attach on an idle session
	if _, ok := k.Knock("allow Bash?", true, false); !ok {
		t.Fatal("the first ask must sound")
	}
	for i := 0; i < 5; i++ {
		if _, ok := k.Knock("allow Bash?", true, false); ok {
			t.Errorf("nag %d sounded again", i+1)
		}
	}
}

// A scape can attach to a session that is ALREADY waiting on the user. It must
// not shout about something that happened before it existed.
func TestAttachingToAWaitingSessionIsSilent(t *testing.T) {
	var k Knocker
	if _, ok := k.Knock("allow Bash?", true, false); ok {
		t.Error("attaching mid-prompt sounded a knock for history")
	}
	// But the next, genuinely new question does sound.
	if _, ok := k.Knock("allow Read?", true, false); !ok {
		t.Error("a new question after attaching must sound")
	}
}

// The two knocks are locked as distinct cues, and a distinction that exists
// only on screen is no distinction to someone looking at another pane.
func TestAskAndDoneSoundDifferently(t *testing.T) {
	var k Knocker
	k.Knock("", false, false)

	got, ok := k.Knock("allow Bash?", true, false)
	if !ok || got != Ask {
		t.Errorf("needs_input gave (%v, %v), want (ask, true)", got, ok)
	}
	// The same text arriving as a finish is a different event: the agent
	// stopped asking and started reporting.
	got, ok = k.Knock("allow Bash?", false, false)
	if !ok || got != Done {
		t.Errorf("the flip to a finish gave (%v, %v), want (done, true)", got, ok)
	}
}

func TestBubbleClearingIsSilentAndRearms(t *testing.T) {
	var k Knocker
	k.Knock("", false, false)
	if _, ok := k.Knock("all done", false, false); !ok {
		t.Fatal("expected the finish knock")
	}
	if _, ok := k.Knock("", false, false); ok {
		t.Error("a bubble expiring must not sound")
	}
	// The same message returning after a gap is a new knock.
	if _, ok := k.Knock("all done", false, false); !ok {
		t.Error("the same message after a clear is a new knock and must sound")
	}
}

// Two different questions back to back are two questions.
func TestADifferentQuestionSoundsAgain(t *testing.T) {
	var k Knocker
	k.Knock("", false, false)
	k.Knock("allow Bash?", true, false)
	if _, ok := k.Knock("allow WebFetch?", true, false); !ok {
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

// His ruling of 2026-09-15: a finish that leaves a failure standing is its own
// knock. It changes whether you come now rather than later, which is the only
// thing a sound is for. It is the done knock in the worried state, not a new
// moment, so the edge detection is the same and a held finish still sounds
// once.
func TestAFinishWhileWorriedIsItsOwnKnock(t *testing.T) {
	var k Knocker
	k.Knock("", false, false)

	// A question is a question, worried or not. The ask must stay the ask:
	// the pose already lets a broken build outrank a question, and the sound
	// must not repeat that mistake.
	got, ok := k.Knock("allow Bash?", true, true)
	if !ok || got != Ask {
		t.Errorf("a question while worried gave (%v, %v), want (ask, true)", got, ok)
	}

	got, ok = k.Knock("it is in, but the tests fail", false, true)
	if !ok || got != Worried {
		t.Errorf("a finish while worried gave (%v, %v), want (worried, true)", got, ok)
	}
	for i := 0; i < 5; i++ {
		if _, ok := k.Knock("it is in, but the tests fail", false, true); ok {
			t.Errorf("the held broken finish sounded again on frame %d", i+1)
		}
	}

	k.Knock("", false, false)
	got, ok = k.Knock("all green", false, false)
	if !ok || got != Done {
		t.Errorf("a clean finish gave (%v, %v), want (done, true)", got, ok)
	}
}

// Three kinds, three names, so `xscapes notify <kind>` and Describe can say
// which one they mean.
func TestEveryKindHasItsOwnName(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range []Kind{Ask, Done, Worried} {
		s := k.String()
		if s == "" || seen[s] {
			t.Errorf("kind %d has name %q, want a distinct non-empty one", int(k), s)
		}
		seen[s] = true
	}
}
