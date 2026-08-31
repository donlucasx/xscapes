// Package reduce folds a stream of protocol events into the continuous state
// the renderer consumes.
//
// Everything time-dependent takes an explicit now, so the whole package is
// testable against a synthetic stream with no clock and no sleeping. That is
// not a stylistic preference: the constants below are the difference between
// a scene that feels alive and one that twitches, and a constant you cannot
// test at a hundred times real speed is a constant you will never tune.
package reduce

import (
	"math"
	"time"

	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
	"github.com/donlucasx/asciiscapes/internal/scape"
)

// The time constants. Every one of these is a claim about how a person glances
// at a screen, so each is written down with its reason rather than tuned until
// it looked nice.
const (
	// TauFall is the sea's decay. After the last tool event the swells fall
	// to 37% in this long and read as flat at roughly three times it. Long
	// enough that the two-second pause between tool calls does not flatten
	// the water; short enough that a finished session settles while you are
	// still looking at it.
	TauFall = 12 * time.Second

	// Impulse is what one tool event adds to heat. Level saturates, so this
	// is really a statement about how many events in a TauFall window count
	// as flat out: about six.
	Impulse = 0.30

	// TurnFloor is the sea while a turn is open but nothing is running --
	// the agent is thinking. Measured on this machine, the gap from prompt
	// to first tool call runs to seventeen seconds; without a floor the
	// scene says "idle" through the whole of it, which is the exact lie the
	// project exists to prevent.
	TurnFloor = 0.30

	// FlightFloor is the sea while at least one tool is actually running. A
	// single ninety-second shell command is the hardest case in the whole
	// design: one event, then silence, then one more. Holding the water up
	// for its duration is the "encode in coverage, not rate" rule applied to
	// time rather than space.
	FlightFloor = 0.45

	// DoneHold is how long the companion holds the finished pose. Bounded on
	// purpose: an alert that waits for acknowledgement has to be dismissed,
	// and the brief is explicit that this must never compete with the agent
	// for attention.
	DoneHold = 20 * time.Second

	// TurnSilence force-closes a turn nothing ever ended. Stop is the only
	// event that closes one, so a crashed agent or a killed pane would
	// otherwise leave the cat working forever.
	TurnSilence = 5 * time.Minute

	// SubStale drops a subagent whose end event never arrived, so one lost
	// event does not strand a kitten on the beach for the rest of the day.
	SubStale = 30 * time.Minute

	// FlightStale drops a tool whose end never arrived. Without it, one lost
	// tool_end pins the sea and the companion at "working" forever -- and the
	// case is not exotic: kill the agent while a Bash is running and no
	// PostToolUse, no Stop and no SessionEnd is ever sent. Generous, because a
	// genuinely long command must not be swept while it is still running.
	FlightStale = 20 * time.Minute
)

// State is everything the renderer needs, and nothing else. It is a plain
// value so a caller can snapshot it, log it, or diff two of them in a test.
type State struct {
	Act    scape.Activity
	Pose   companion.State
	Bubble string

	// Kittens is how many subagents are running right now.
	Kittens int

	// Tail is the sand: newest last.
	Tail []Line

	// Session is who we are following, for the footer.
	Session string

	// tail lets a renderer re-fit the sand to its own width.
	tail *tail
	// LastEvent is when we last heard anything at all.
	LastEvent time.Time
	// Events counts everything applied, which is the cheapest possible
	// answer to "is this thing actually wired up?".
	Events int
}

type inflight struct {
	started time.Time
	op      event.Op
	tool    string
	target  string
}

// Reducer folds events. Not safe for concurrent use: it is owned by the render
// goroutine, which is what lets the whole thing be mutex-free.
type Reducer struct {
	heat    float64
	heatAt  time.Time
	turnOpn bool
	turnAt  time.Time

	flight map[string]inflight
	subs   map[string]time.Time

	worried    bool
	needsInput bool
	doneAt     time.Time
	bubble     string

	ctx    float64
	ctxSet bool

	tail    tail
	session string
	last    time.Time
	count   int
}

func New(session string) *Reducer {
	return &Reducer{
		flight:  map[string]inflight{},
		subs:    map[string]time.Time{},
		session: session,
	}
}

// Apply folds one event in.
func (r *Reducer) Apply(e event.Event, now time.Time) {
	r.decay(now)
	r.last = now
	r.count++

	// Any sign of work refreshes the turn clock. It used to be set once, at
	// the prompt, so TurnSilence measured how LONG the turn was rather than
	// how quiet it had gone -- and every turn past five minutes put the cat to
	// sleep while the agent was still working.
	if e.Busy() || e.Kind == event.SubStart || e.Kind == event.SubEnd ||
		e.Kind == event.Error || e.Kind == event.TestPass || e.Kind == event.TestFail {
		r.turnAt = now
	}

	switch e.Kind {
	case event.SessionStart:
		// Claude Code re-announces the session after an auto-compaction, in
		// the middle of a live turn. Resetting there wipes the sea, the
		// litter and the sand while the agent is still working. Only a
		// genuinely new conversation resets; the source rides in Text.
		switch e.Text {
		case "compact", "fork":
			r.heat += Impulse
		default: // startup, clear, resume, or an unknown future source
			r.reset()
		}

	case event.SessionEnd:
		r.reset()

	case event.Prompt:
		// A new prompt is the user acknowledging whatever came before, which
		// is the only honest moment to clear a worry: hooks can tell us a
		// command failed, never that the code is fixed. Encoding it as "you
		// have not looked at this yet" is a claim we can actually support.
		r.turnOpn, r.turnAt = true, now
		r.worried = false
		r.needsInput = false
		r.doneAt = time.Time{}
		r.bubble = ""
		r.heat += Impulse

	case event.ToolStart:
		if e.ID != "" {
			r.flight[e.ID] = inflight{started: now, op: e.Op, tool: e.Tool, target: e.Target}
		}
		// A scape attached partway through a session never saw the prompt, so
		// tool traffic has to be able to open a turn by itself -- otherwise it
		// spends the whole turn insisting the agent is idle.
		r.turnOpn, r.turnAt = true, now
		// A tool starting means a permission prompt, if there was one, was
		// answered. Nothing else clears it that reliably.
		r.needsInput = false
		r.heat += Impulse

	case event.ToolEnd:
		if e.ID != "" {
			delete(r.flight, e.ID)
		}
		r.heat += Impulse
		r.tail.push(line{
			at: now, op: e.Op, tool: e.Tool, target: e.Target,
			detail: e.Detail, ms: e.MS, bad: false,
		})

	case event.Error, event.TestFail:
		if e.ID != "" {
			delete(r.flight, e.ID)
		}
		r.worried = true
		r.heat += Impulse
		r.tail.push(line{
			at: now, op: e.Op, tool: e.Tool, target: e.Target,
			detail: e.Detail, ms: e.MS, bad: true,
		})

	case event.TestPass:
		r.heat += Impulse
		r.tail.push(line{at: now, op: e.Op, tool: e.Tool, target: e.Target, detail: e.Detail, ms: e.MS})

	case event.NeedsInput:
		r.needsInput = true
		r.bubble = e.Text

	case event.Done:
		r.turnOpn = false
		r.needsInput = false
		r.doneAt = now
		r.flight = map[string]inflight{}
		// Clear first: a knock with no words is honest, a knock still showing
		// the last permission request is a lie about what is being asked.
		r.bubble = e.Text

	case event.SubStart:
		if e.Agent != "" {
			r.subs[e.Agent] = now
		}
		r.heat += Impulse

	case event.SubEnd:
		if e.Agent != "" {
			delete(r.subs, e.Agent)
		}
		r.heat += Impulse

	case event.Compact:
		r.heat += Impulse

	case event.Context:
		if e.Frac != nil {
			r.ctx, r.ctxSet = clamp01(*e.Frac), true
		}

	case event.Todo:
		r.heat += Impulse
	}
}

// Tick advances time with no event, which is what the render loop calls on
// every frame.
func (r *Reducer) Tick(now time.Time) { r.decay(now) }

func (r *Reducer) decay(now time.Time) {
	if r.heatAt.IsZero() {
		r.heatAt = now
		return
	}
	dt := now.Sub(r.heatAt)
	if dt <= 0 {
		return
	}
	r.heatAt = now
	r.heat *= math.Exp(-dt.Seconds() / TauFall.Seconds())
	if r.heat < 1e-4 {
		r.heat = 0
	}

	// A turn nothing closed, and a subagent nothing stopped, are both single
	// points of failure for a scene that never settles. Time them out.
	for id, f := range r.flight {
		if now.Sub(f.started) > FlightStale {
			delete(r.flight, id)
		}
	}
	// TurnSilence is not conditioned on the flight map being empty. It used to
	// be, which meant the one case the timeout was written for -- an agent
	// killed mid-tool -- was the exact case that could not fire it.
	if r.turnOpn && now.Sub(r.turnAt) > TurnSilence {
		r.turnOpn = false
	}
	for id, t := range r.subs {
		if now.Sub(t) > SubStale {
			delete(r.subs, id)
		}
	}
}

// State reads out the current scene. TimeOfDay comes from the caller's wall
// clock rather than from here, because the sky is the world and the reducer
// only knows about the work.
func (r *Reducer) State(now time.Time) State {
	r.decay(now)

	// Saturating, so the mapping stays meaningful whether a session fires six
	// events a minute or six hundred. Linear would put every real session in
	// the bottom tenth of the range and then clip the one time it mattered.
	lvl := 1 - math.Exp(-r.heat)
	if len(r.flight) > 0 && lvl < FlightFloor {
		lvl = FlightFloor
	}
	if r.turnOpn && lvl < TurnFloor {
		lvl = TurnFloor
	}
	working := r.turnOpn || len(r.flight) > 0

	st := State{
		Act: scape.Activity{
			Working:     working,
			Level:       clamp01(lvl),
			ContextUsed: r.ctx,
		},
		Pose:      r.pose(now),
		Kittens:   len(r.subs),
		Tail:      r.tail.lines(now),
		tail:      &r.tail,
		Session:   r.session,
		LastEvent: r.last,
		Events:    r.count,
	}
	// The bubble is NOT gated on the pose. Gating it meant that after any
	// failed command the companion went Worried and swallowed everything --
	// including a permission prompt that the agent is actually blocked on, so
	// the session could sit waiting for input with no signal at all. The pose
	// says how the companion feels; the bubble says what it needs. They are
	// different channels and only one of them is allowed to be silent.
	if r.needsInput || (!r.doneAt.IsZero() && now.Sub(r.doneAt) < DoneHold) {
		st.Bubble = r.bubble
	}
	return st
}

// pose resolves the companion. The order is the whole point: a broken thing
// outranks a question, a question outranks work, and work outranks rest. Two
// of these are true at once constantly -- the agent is almost always "working"
// while it is also waiting for permission -- so without a stated precedence
// the pose would depend on which event happened to arrive last.
func (r *Reducer) pose(now time.Time) companion.State {
	switch {
	case r.worried:
		return companion.Worried
	case r.needsInput:
		return companion.NeedsYou
	case !r.doneAt.IsZero() && now.Sub(r.doneAt) < DoneHold:
		return companion.NeedsYou
	case r.turnOpn || len(r.flight) > 0:
		return companion.Working
	default:
		return companion.Resting
	}
}

func (r *Reducer) reset() {
	r.heat = 0
	r.turnOpn = false
	r.flight = map[string]inflight{}
	r.subs = map[string]time.Time{}
	r.worried, r.needsInput = false, false
	r.doneAt = time.Time{}
	r.bubble = ""
	r.tail = tail{}
}

// FitTail re-renders the sand to a column budget, dropping whole pieces of a
// line rather than chopping it where it runs out of room.
func (s State) FitTail(now time.Time, cols int) []Line {
	if s.tail == nil {
		return s.Tail
	}
	return s.tail.fit(now, cols)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
