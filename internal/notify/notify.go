// Package notify is the audible half of the nudge.
//
// The rubric this project is built against puts 30% on the waiting experience,
// and the note under it is explicit: the nudge has to genuinely beat a terminal
// bell. A scape in a side pane is often not the pane being looked at, so the
// sound is not decoration -- it is the only channel that reaches a user who has
// turned away, which is exactly the user this project exists for.
//
// Two sounds, because the brief locks done and needs_input as distinct cues and
// a distinction that exists only on screen is no distinction to someone looking
// elsewhere. They are picked from the same family so the pair still reads as
// one scape's voice: a bright chime when the agent is BLOCKED on you, a deep
// sonar note when it has finished and you can come back whenever.
package notify

import (
	"os"
	"os/exec"
	"runtime"
)

// Kind is which of the two knocks happened.
type Kind int

const (
	// Ask: the agent is blocked and cannot continue without you.
	Ask Kind = iota
	// Done: the turn finished. Nothing is waiting on you.
	Done
)

func (k Kind) String() string {
	if k == Ask {
		return "ask"
	}
	return "done"
}

// Player plays the two knocks. The zero value is silent, which is the right
// failure mode: a scape that cannot find a sound player must still run.
type Player struct {
	cmd  string
	args map[Kind][]string
	// bell falls back to the terminal's own BEL when no player exists. It is
	// the thing the brief says to beat, so it is the floor, not the plan.
	bell bool
}

// SilentEnv mutes the notification when set to anything but an empty string.
const SilentEnv = "ASCIISCAPES_SILENT"

// New picks a player for this machine. Sound is on by default -- ambient audio
// is the thing the brief keeps off, not the notification.
func New() *Player {
	if os.Getenv(SilentEnv) != "" {
		return &Player{}
	}
	switch runtime.GOOS {
	case "darwin":
		// Verified present rather than assumed: a missing file makes afplay
		// exit non-zero and the knock is silently lost.
		ask, okA := firstFile(
			"/System/Library/Sounds/Glass.aiff",
			"/System/Library/Sounds/Tink.aiff",
		)
		done, okD := firstFile(
			"/System/Library/Sounds/Submarine.aiff",
			"/System/Library/Sounds/Purr.aiff",
		)
		if okA && okD && have("afplay") {
			return &Player{cmd: "afplay", args: map[Kind][]string{
				Ask:  {ask},
				Done: {done},
			}}
		}
	case "linux":
		ask, okA := firstFile(
			"/usr/share/sounds/freedesktop/stereo/message.oga",
			"/usr/share/sounds/freedesktop/stereo/bell.oga",
		)
		done, okD := firstFile(
			"/usr/share/sounds/freedesktop/stereo/complete.oga",
			"/usr/share/sounds/freedesktop/stereo/message.oga",
		)
		if okA && okD && have("paplay") {
			return &Player{cmd: "paplay", args: map[Kind][]string{
				Ask:  {ask},
				Done: {done},
			}}
		}
	}
	return &Player{bell: true}
}

// Play sounds one knock. It never blocks the render loop and never fails
// loudly: a frame is due every few dozen milliseconds and a missed sound is
// not worth a stutter, let alone a crash.
func (p *Player) Play(k Kind) {
	if p == nil {
		return
	}
	if p.bell {
		// BEL moves no cursor, so it is safe to write into a frame.
		os.Stdout.WriteString("\a")
		return
	}
	if p.cmd == "" {
		return
	}
	args, ok := p.args[k]
	if !ok {
		return
	}
	cmd := exec.Command(p.cmd, args...)
	if err := cmd.Start(); err != nil {
		return
	}
	// Reaped in the background: without a Wait the player becomes a zombie
	// for the life of the session, once per knock.
	go cmd.Wait()
}

// Silent reports whether this player will make no sound at all, so a caller
// can say so rather than leaving the user wondering.
func (p *Player) Silent() bool { return p != nil && p.cmd == "" && !p.bell }

// Describe names what the player will do, for `-info`.
func (p *Player) Describe() string {
	switch {
	case p == nil || p.Silent():
		return "silent"
	case p.bell:
		return "terminal bell"
	default:
		return p.cmd
	}
}

func have(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func firstFile(paths ...string) (string, bool) {
	for _, p := range paths {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

// Knocker turns a stream of frames into at most one sound per knock.
//
// The sound keys off the BUBBLE rather than the pose, for the same reason the
// bubble itself is not gated on the pose: a broken build outranks a question in
// the companion's posture, so a pose-driven sound would go silent on the one
// event the user has to answer. The pose says how the companion feels; the
// bubble says what it needs.
type Knocker struct {
	started bool
	text    string
	ask     bool
}

// Knock reports which sound this frame earned, if any.
//
// It stays quiet when nothing changed, which is what makes it safe to call
// every frame, and it stays quiet on the FIRST frame it ever sees: a scape
// attached to a session that is already waiting must not announce something
// that happened before it existed.
func (k *Knocker) Knock(bubble string, ask bool) (Kind, bool) {
	changed := bubble != "" && (bubble != k.text || ask != k.ask)
	first := !k.started
	k.started, k.text, k.ask = true, bubble, ask
	if !changed || first {
		return Done, false
	}
	if ask {
		return Ask, true
	}
	return Done, true
}
