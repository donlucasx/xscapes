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
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scape"
)

// The time constants. Every one of these is a claim about how a person glances
// at a screen, so each is written down with its reason rather than tuned until
// it looked nice.
//
// Vars rather than consts, and that is not a style choice. They had gone
// unverified for the whole project partly because there was no way to ask what
// a different value would have LOOKED like without editing the source and
// rebuilding -- so nobody ever did. `xscapes tune -sweep` folds every real
// recording again for each candidate, which needs to set them.
var (
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

	// KittenExit is how long a finished subagent's kitten takes to swim off.
	// The litter count drops the moment the end arrives (the count is live);
	// the exit is a position, not a rate, so a glance still reads it.
	KittenExit = 6 * time.Second

	// KittenDwell is the least time a kitten stays in the litter after its
	// subagent starts. A subagent that finishes in seconds was invisible: its
	// kitten had come and swum off before anyone looked (his report, twice,
	// 2026-09-04 and 2026-09-05, agents of 6.9 s and 15.7 s). An end that
	// arrives inside the dwell is remembered, and the swim-off waits for it.
	// His ruling, 2026-09-05.
	KittenDwell = 60 * time.Second

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
	// BubbleAsk says which of the two knocks the bubble is: true means the
	// agent is blocked on the user (needs_input), false means the finish
	// knock. The renderer draws them as different balloons, because the
	// brief locks done and needs_input as distinct cues.
	BubbleAsk bool

	// Kittens is how many subagents are running right now, plus any that
	// finished inside their KittenDwell and are still sitting it out.
	Kittens int
	// KittenExits is one entry per subagent whose swim-off is under way: how
	// far along it is, 0 when it leaves the litter, 1 when gone.
	// Oldest first.
	KittenExits []float64

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

	// Steps is how many MAIN-THREAD tool events this session has seen, and
	// StepAge how long ago the last one was. One step of the companion's pace
	// each. An agent working alone makes all its calls on the main thread and
	// paces constantly; one orchestrating six subagents makes few itself and
	// settles down, which is his earlier ask falling out rather than being
	// special-cased.
	Steps   int
	StepAge float64
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

	// steps counts main-thread tool events; stepAt is when the last one landed.
	// The companion takes one step of its pace per step.
	steps  int
	stepAt time.Time

	flight map[string]inflight
	subs   map[string]time.Time
	gone   map[string]time.Time // subagents that ended: when each kitten leaves (may be ahead of now, the dwell), until it has swum off

	worried    bool
	needsInput bool
	doneAt     time.Time
	bubble     string

	ctx    float64
	ctxSet bool

	todoDone, todoOf int
	// turnsDone is how many turns have closed this session, and tasksDone how
	// many subagents have finished. Together they light the constellation when
	// the agent keeps no checklist. See StarsCap and TasksPerStar.
	turnsDone, tasksDone int

	tail    tail
	session string
	last    time.Time
	count   int
}

func New(session string) *Reducer {
	return &Reducer{
		flight:  map[string]inflight{},
		subs:    map[string]time.Time{},
		gone:    map[string]time.Time{},
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
		// One step of the companion's pace, and ONLY for the main thread.
		// Subagent work belongs to the litter; if it moved the parent too, the
		// same event would be spending two channels. His correction is what
		// this encodes: "should be tied up to an actual agent action so its not
		// random" -- a step is a COUNT and where it ends up is a POSITION, both
		// of which survive a screenshot, where a drift is decoration.
		if e.Agent == "" {
			r.steps++
			r.stepAt = now
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
		// MAIN-THREAD ERRORS ONLY -- his ruling of 2026-09-07. The companion was
		// Worried 44.6% of active time, measured over 330 hours of his own
		// recordings, which is nearly as much as it was Working; a face that
		// says "something is broken" for half the session says nothing.
		//
		// 590 of those 675 errors -- 87% -- fired inside a SUBAGENT: a fan-out
		// grepping for something that was not there, a probe exiting 1. That is
		// the subagent's business, and it almost always carried on: 99.4% of all
		// errors are followed by a successful tool call, median 1.7s later, and
		// only 3 of 675 had nothing happen after them at all.
		//
		// Event.Agent is set only when the hook fires inside a subagent
		// (notes/claude-hooks-verified.md), so the main thread is exactly
		// Agent == "". Folded back through the real spools: worried 44.6% ->
		// 16.4%, working 49.6% -> 77.7%, episodes 90 -> 50, needs-you and done
		// unchanged.
		//
		// The error is still recorded everywhere else -- it drives the sea and
		// it goes on the sand. Only the FACE is quiet, because the face is the
		// channel that was crying wolf.
		if e.Agent == "" {
			r.worried = true
		}
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
		r.turnsDone++
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
			delete(r.gone, e.Agent) // back in the litter, mid-swim or not
		}
		r.heat += Impulse

	case event.SubEnd:
		if e.Agent != "" {
			r.tasksDone++
			if started, ok := r.subs[e.Agent]; ok {
				// The kitten leaves at the end, or at the end of its dwell,
				// whichever is later. Until then it stays in the litter.
				leave := now
				if d := started.Add(KittenDwell); d.After(now) {
					leave = d
				}
				r.gone[e.Agent] = leave
				if !leave.After(now) {
					delete(r.subs, e.Agent)
				}
			}
		}
		r.heat += Impulse

	case event.Compact:
		r.heat += Impulse

	case event.Context:
		if e.Frac != nil {
			r.ctx, r.ctxSet = clamp01(*e.Frac), true
		}

	case event.Todo:
		// A todo list that shrinks to nothing is a list that was CLEARED, not
		// one that was finished, and the difference matters: the stars would
		// blink out at the moment the work completed. So a zero total leaves
		// the last real reading standing, and only a new list replaces it.
		if e.Of > 0 {
			r.todoDone, r.todoOf = e.N, e.Of
			if r.todoDone > r.todoOf {
				r.todoDone = r.todoOf
			}
		}
		r.heat += Impulse
	}
}

// StarsCap is how many places the constellation holds when it is counting
// closed turns rather than a checklist.
//
// Measured over his own recorded sessions: turns closed run p50 12, p75 16,
// p90 24 and max 32, so a sky of 32 fills through a long session and never
// pegs. That range is the whole reason this channel could be rebound at all --
// the two other accumulators that fire often enough were both unusable, files
// changed at p50 2 and p90 107 and finished subagents at p50 10 and p90 172.
// A channel that is empty half the time and saturated the rest is not a
// channel; that is the same defect the sea has above level 0.6.
const StarsCap = 32

// TasksPerStar is how many finished subagents make one star.
//
// HIS observation, and the measurements back it: "I feel like the agents
// complete a lot more tasks than they would actual turns (specially on long
// sprints)." During a fan-out a subagent finishes every 31 seconds at the
// median, against fourteen minutes between closed turns.
//
// Tasks INSTEAD of turns does not work, and that is why this is a ratio rather
// than a swap: subagent counts are bimodal across his sessions -- twenty of
// thirty have between zero and twenty-six, ten have between sixty-four and
// seven hundred and thirty-four -- so counting them alone leaves nineteen of
// thirty sessions with almost no stars at all while a workflow session pegs the
// sky in minutes.
//
// Eight is the grain that survives both. Measured over the real event stream,
// gaps between stars in minutes:
//
//	turns only        p50 14.5   p90 130.8
//	+ tasks/32        p50 12.6   p90 109.8
//	+ tasks/16        p50  9.9   p90  92.8
//	+ tasks/8         p50  5.4   p90  58.8   <- this
//
// and per session it moves the median from 12 stars to 16, pegging the sky in
// three sessions of thirty rather than one. A finer grain buys less than it
// costs: tasks/4 pegs nine of thirty.
const TasksPerStar = 8

// stars is what the constellation counts.
//
// The channel was spec'd as todos completed and has NEVER lit: across his whole
// recorded history -- 60,000+ tool calls -- TodoWrite has been called zero
// times, which hook.go had already noticed. So it counts CLOSED TURNS instead,
// which is the same idea (units of work finished) carried by an event that
// actually fires, about a dozen times a session.
//
// A checklist still wins where one exists. An agent that keeps todos is telling
// us something more specific than "a turn ended", and other agents are not
// Claude Code; this only falls back when nothing has ever sent one.
//
// ⚠ THE RULE TENSION, NAMED RATHER THAN HIDDEN: `done` also raises the finish
// knock, and the encoding rule forbids one event driving two channels. The
// argument for it here is that the two are different KINDS of variable -- the
// knock is momentary and says "it just finished, come back", the stars are
// cumulative and say "this session has got through N" -- and they never
// compete, because they light at the same instant in different parts of the
// frame. It is one line to bind this elsewhere if he rules the other way.
func stars(todoDone, todoOf, turns, tasks int) int {
	if todoOf > 0 {
		return todoDone
	}
	n := turns + tasks/TasksPerStar
	if n > StarsCap {
		return StarsCap
	}
	return n
}

// starTotal is how many PLACES the sky lays out, which fixes where each star
// sits. It is deliberately not the count: positions are keyed by index, so a
// total that moved would slide every existing star sideways as the next one
// arrived, and the brief asks that a star light where it always was.
func starTotal(todoOf int) int {
	if todoOf > 0 {
		return todoOf
	}
	return StarsCap
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
	for id, t := range r.gone {
		if t.After(now) {
			continue // still dwelling, still in the litter
		}
		delete(r.subs, id) // its dwell is over: the swim-off has begun
		if now.Sub(t) >= KittenExit {
			delete(r.gone, id)
		}
	}
}

// kittenExits is the swim-off progress of every recently finished subagent,
// oldest first (ties by id, so two ending in one tick draw the same way every
// frame).
func (r *Reducer) kittenExits(now time.Time) []float64 {
	if len(r.gone) == 0 {
		return nil
	}
	type g struct {
		id string
		at time.Time
	}
	gs := make([]g, 0, len(r.gone))
	for id, t := range r.gone {
		if t.After(now) {
			continue // dwelling: counted in Kittens, not leaving yet
		}
		gs = append(gs, g{id, t})
	}
	if len(gs) == 0 {
		return nil
	}
	sort.Slice(gs, func(i, j int) bool {
		if !gs[i].at.Equal(gs[j].at) {
			return gs[i].at.Before(gs[j].at)
		}
		return gs[i].id < gs[j].id
	})
	out := make([]float64, len(gs))
	for i, x := range gs {
		out[i] = clamp01(float64(now.Sub(x.at)) / float64(KittenExit))
	}
	return out
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

	// The floors LIFT the range rather than clamping it, and that is worth
	// more than it sounds.
	//
	// Clamping was the obvious way to write this and it spent the sea's whole
	// dynamic range below the point where the sea actually lives. Folded from
	// 11 real sessions, 18,919 events: 77% of all working time sat in two bins
	// between 0.30 and 0.50, because the floors ARE 0.30 and 0.45 and heat only
	// modulated what was above them. The swells carry how hard the agent is
	// working and they were carrying it across a fifth of their range.
	//
	// Lifting instead -- floor + (1-floor)*heat -- keeps every promise the
	// floors were added for (a running tool never reads idle, thinking time
	// never reads idle) and gives the whole range back above it. Same data:
	// the bins go 18/24/21/17/9/5/5 instead of 45/32/8/5/4/2/4, and saturation
	// moves 3.1% to 3.5%.
	floor := 0.0
	if r.turnOpn && floor < TurnFloor {
		floor = TurnFloor
	}
	if len(r.flight) > 0 && floor < FlightFloor {
		floor = FlightFloor
	}
	lvl = floor + (1-floor)*lvl
	working := r.turnOpn || len(r.flight) > 0

	st := State{
		Steps:   r.steps,
		StepAge: sinceOr(now, r.stepAt),
		Act: scape.Activity{
			Working:     working,
			Level:       clamp01(lvl),
			ContextUsed: r.ctx,
			TodoDone:    stars(r.todoDone, r.todoOf, r.turnsDone, r.tasksDone),
			TodoTotal:   starTotal(r.todoOf),
		},
		Pose:        r.pose(now),
		Kittens:     len(r.subs),
		KittenExits: r.kittenExits(now),
		Tail:        r.tail.lines(now),
		tail:        &r.tail,
		Session:     r.session,
		LastEvent:   r.last,
		Events:      r.count,
	}
	// The bubble is NOT gated on the pose. Gating it meant that after any
	// failed command the companion went Worried and swallowed everything --
	// including a permission prompt that the agent is actually blocked on, so
	// the session could sit waiting for input with no signal at all. The pose
	// says how the companion feels; the bubble says what it needs. They are
	// different channels and only one of them is allowed to be silent.
	// An open ask outranks a stale knock, same as it does for the pose.
	switch {
	case r.needsInput:
		st.Bubble, st.BubbleAsk = r.bubble, true
	case !r.doneAt.IsZero() && now.Sub(r.doneAt) < DoneHold:
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
		return companion.Done
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
	r.gone = map[string]time.Time{}
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

// sinceOr is seconds since t, or a large number when t is the zero time, so a
// session that has not stepped yet reads as "long ago" rather than "just now".
func sinceOr(now, t time.Time) float64 {
	if t.IsZero() {
		return 1e9
	}
	return now.Sub(t).Seconds()
}
