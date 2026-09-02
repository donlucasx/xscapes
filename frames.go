package main

import (
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/notify"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// frames paints the scape one frame at a time.
//
// It exists because there are now two places that need a frame -- the pane
// that owns its whole window, and the band below a hosted agent -- and a
// scene assembled twice is a scene that drifts. Everything about what the
// scape looks like lives here; the callers only decide where to put it.
type frames struct {
	c        *canvas.Canvas
	sh       *scape.Shore
	cat      *companion.Cat
	lay      layout
	ccw, chh int
	mirror   bool
	profile  term.Profile
	seed     int64

	// Set once a session is found. Until then the demo cycle runs.
	bus     *event.Bus
	red     *reduce.Reducer
	player  *notify.Player
	knocker notify.Knocker

	ctxUsed, tod float64
	start        time.Time
}

func newFrames(w, h int, seed int64, ascii, mirror bool, ctxUsed, tod float64) *frames {
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, ascii)
	cat := companion.NewCat()
	cat.FaceLeft(mirror)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, mirror)
	sh.MoonX = lay.MoonX
	return &frames{
		c: c, sh: sh, cat: cat, lay: lay, ccw: ccw, chh: chh,
		mirror: mirror, profile: term.DetectProfile(), seed: seed,
		player: notify.New(), ctxUsed: ctxUsed, tod: tod, start: time.Now(),
	}
}

// resize refits the scene, and says whether the window got smaller -- the only
// case that needs the terminal cleared, since growing paints strictly more
// cells than the frame before it.
func (f *frames) resize(w, h int) (shrank bool) {
	shrank = w < f.c.W || h < f.c.H
	f.c.Resize(w, h)
	f.lay = compose(w, f.ccw, f.mirror)
	f.sh.MoonX = f.lay.MoonX
	return shrank
}

func (f *frames) size() (w, h int) { return f.c.W, f.c.H }

// follow attaches the scene to a session's event bus.
func (f *frames) follow(bus *event.Bus, red *reduce.Reducer) { f.bus, f.red = bus, red }

func (f *frames) following() bool { return f.red != nil }

// state drains everything that has arrived since the last frame and returns
// what the scene should show now. Non-blocking: a frame must never wait on
// the bus.
func (f *frames) state(now time.Time, t float64) reduce.State {
	if f.red == nil {
		// The sky is the world: time of day is the wall clock whether or not
		// there is a session to follow. Without this the demo ran at tod 0 --
		// midnight, which is deliberately monochrome -- so every launch showed
		// a black-and-white scene for the second or two before the agent
		// announced itself, and then jumped to colour.
		tod := f.tod
		if tod == 0 {
			tod = timeOfDay(now)
		}
		return demoState(t, f.ctxUsed, tod)
	}
	for draining := true; draining; {
		select {
		case e := <-f.bus.C:
			f.red.Apply(e, now)
		default:
			draining = false
		}
	}
	st := f.red.State(now)
	// Re-render the sand to the columns this layout actually leaves it, so a
	// narrow pane loses whole pieces of a line rather than getting a path
	// chopped mid-word.
	st.Tail = st.FitTail(now, f.lay.SandTo-f.lay.SandFrom)
	// The sky is the world: time of day is the wall clock, never anything the
	// agent did.
	st.Act.TimeOfDay = timeOfDay(now)
	if f.tod != 0 {
		st.Act.TimeOfDay = f.tod
	}
	if f.ctxUsed != 0 {
		st.Act.ContextUsed = f.ctxUsed
	}
	return st
}

// frame paints one frame and returns it as ANSI, rows joined by newlines.
//
// It also rings the audible half of the nudge, keyed off the bubble rather
// than the pose and edge-detected, so a sixty-second permission nag knocks
// once rather than once a minute. Only when following a real session: the
// demo cycles every state on a timer, and a scape nobody attached to has
// nothing to announce.
func (f *frames) frame(now time.Time) string {
	t := now.Sub(f.start).Seconds()
	st := f.state(now, t)
	if f.red != nil {
		if k, ring := f.knocker.Knock(st.Bubble, st.BubbleAsk); ring {
			f.player.Play(k)
		}
	}
	f.sh.Update(f.c, t, st.Act)
	top := f.c.H - 2 - f.chh
	drawScene(f.c, f.sh, f.cat, f.lay, st, t, f.seed, top)
	return f.c.Render(f.profile)
}
