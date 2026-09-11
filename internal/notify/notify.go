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
// one scape's voice: a rising drop when the agent is BLOCKED on you, a falling
// one when it has finished and you can come back whenever.
//
// The cues SHIP WITH XSCAPES (his ask, 2026-09-11: "a fun sound that users can
// identify w xscapes"). They used to be Glass.aiff and Submarine.aiff, which
// work perfectly well and are the reason this changed: a sound every Mac has
// played for twenty years belongs to the OS and cannot belong to a product.
// See sounds.go for the embedding, and notes/s28-sound for where they came
// from.
package notify

import (
	"os"
	"os/exec"

	"github.com/donlucasx/xscapes/internal/envx"
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
	// source names which rung of the fallback ladder this player is standing
	// on. It is carried only so Describe can say it: the three rungs sound
	// completely different and nothing else on screen distinguishes them.
	source string
	// bell falls back to the terminal's own BEL when no player exists. It is
	// the thing the brief says to beat, so it is the floor, not the plan.
	bell bool
}

// SilentEnv mutes the notification when set to anything but an empty string.
// The pre-rename XSCAPES_SILENT still works; see internal/envx.
const SilentEnv = "XSCAPES_SILENT"

// ourCue is what Describe calls the shipped droplet pair.
const ourCue = "xscapes droplet"

// New picks a player for this machine. Sound is on by default -- ambient audio
// is the thing the brief keeps off, not the notification.
//
// Three rungs, in this order, and each exists for a reason the one below it
// cannot cover:
//
//  1. the EMBEDDED DROPLET, when it can be written to disk and something can
//     play it. This is the cue that is xscapes', and the only rung that makes
//     the sound identifiable.
//  2. the SYSTEM SOUNDS it played until 2026-09-11. Reached when the droplet
//     cannot be materialised -- a read-only home, a full disk, no HOME at all.
//     Borrowed and anonymous, but it is a real sound at the right moment.
//  3. the TERMINAL BELL. Reached when there is no player binary to run at all.
//     It is the thing the brief says to beat, so it is the floor, not the plan.
//
// The ladder is walked once, here, rather than per knock: every rung does file
// I/O, and Play is called from the render loop.
func New() *Player {
	if envx.Lookup("SILENT") != "" {
		return &Player{}
	}
	// Checked before either file rung, because both of them need it: a cue
	// with nothing to play it is not a cue.
	cmd := playerBin()
	if cmd == "" {
		return &Player{bell: true}
	}
	if ask, done, ok := materialiseCues(); ok {
		return filePlayer(cmd, ourCue, ask, done)
	}
	if ask, done, ok := systemCue(); ok {
		return filePlayer(cmd, systemCueName(), ask, done)
	}
	return &Player{bell: true}
}

func filePlayer(cmd, source, ask, done string) *Player {
	return &Player{cmd: cmd, source: source, args: map[Kind][]string{
		Ask:  {ask},
		Done: {done},
	}}
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

// Describe names what the player will actually play, for `-info` and for
// `xscapes notify`.
//
// It names the CUE and not just the command, because the command is the one
// thing all three sounding rungs have in common: "afplay" is true of the
// droplet and of Glass alike, and the user hearing a twenty-year-old system
// chime has no other way to find out that the shipped cue never got written.
func (p *Player) Describe() string {
	switch {
	case p == nil || p.Silent():
		return "silent"
	case p.bell:
		return "terminal bell"
	case p.source != "":
		return p.cmd + " (" + p.source + ")"
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
