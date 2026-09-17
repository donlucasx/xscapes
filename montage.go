package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE HERO, IN VARIANTS (2026-09-16 night). His ask, verbatim in
// _FEEDBACK.md §Session 39: "im not sure which xscape to feature as the main,
// but im leaning towards the mountain scape. Can we test a couple different
// gifs for that section, a couple versions depicting the mountain scape, and
// a couple hybrids where we showcase both mountain/coast. Maybe even another
// couple variants where we make a montage/trailer of a terminal xscape
// session that begins with the owl in the mountain scape. It does a quick
// job with a few owlets. Then finishes the task, and the owl turns into the
// cat, which emotes for a moment, and then turns into the crab. The crab
// emotes, then the background changes from forest to coast (same time of
// the day and sun position, afternoon), and the crab and crablets get to
// work as the sun sets into the night".
//
// A montage is the hero's own machinery -- one session folded through the
// reducer, the agent's transcript above its scape in the same window, the
// hour turning underneath -- with two things the hero could not do: the
// SCAPE and the COMPANION can change part-way, on a timeline in session
// seconds, and the hour follows a schedule of stops rather than one straight
// run. Everything else is shared: the beats, the pane, the palette, the
// player. heroVariants is the six he is choosing from.

// heroSeg is one stretch of a montage: which scape is on screen and who
// lives in it, from a session second on.
type heroSeg struct {
	at    float64 // session seconds
	scape string  // ScapeShore or ScapeVista
	who   string  // "owl", companion.NameCat or companion.NameCrab
	// pose, when set, overrides the reducer's for the stretch: a companion
	// that has just arrived shows its finished face whatever the session is
	// doing. The balloon stays the reducer's, so the finish knock the owl
	// raised is still up when the cat takes its place.
	pose *companion.State
}

// todStop is the hour at a session second; between stops it runs straight.
type todStop struct{ at, tod float64 }

type montage struct {
	key, label, note string
	// secs is the loop in clip seconds; speed is session seconds per clip
	// second, as on the hero (loopSecs / 15).
	secs, speed           float64
	fps                   int
	cols, agentRows, rows int
	beats                 []loopBeat
	stops                 []todStop
	segs                  []heroSeg
}

func (m montage) todAt(session float64) float64 {
	s := m.stops
	if len(s) == 0 {
		return 0.5
	}
	if session <= s[0].at {
		return frac(s[0].tod)
	}
	for i := 1; i < len(s); i++ {
		if session <= s[i].at {
			u := (session - s[i-1].at) / (s[i].at - s[i-1].at)
			return frac(s[i-1].tod + (s[i].tod-s[i-1].tod)*u)
		}
	}
	return frac(s[len(s)-1].tod)
}

func frac(x float64) float64 { return x - math.Floor(x) }

func (m montage) segAt(session float64) heroSeg {
	cur := m.segs[0]
	for _, s := range m.segs {
		if s.at <= session {
			cur = s
		}
	}
	return cur
}

// montageFrames renders a montage frame by frame, the way gifFrames renders
// the hero: the session wound on by the clip's speed, the waves, the wind
// and the companions keeping real time. The shore and the vista are each
// made once and kept, so the tide and the flock remember across a switch.
func montageFrames(seed int64, m montage, pal *canvas.HTMLPalette) ([]string, error) {
	if m.fps <= 0 {
		m.fps = 6
	}
	if len(m.segs) == 0 {
		return nil, fmt.Errorf("%s: no segments", m.key)
	}
	sc := gifScene{name: m.key, secs: m.secs, speed: m.speed, fps: m.fps, cols: m.cols,
		agentRows: m.agentRows, rows: m.rows, beats: m.beats, pal: pal, cube: !truecolorFX}
	run := newSessionRun(sc)
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	c := canvas.New(m.cols, m.rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cats := map[string]*companion.Cat{}
	catOf := func(who string) *companion.Cat {
		if who == "owl" {
			return nil
		}
		if cats[who] == nil {
			k := companion.New(who)
			k.FaceLeft(true)
			cats[who] = k
		}
		return cats[who]
	}
	ccw, chh := companion.New(companion.DefaultName).Size()
	lay := compose(c.W, ccw, true)
	sh := scape.NewShore(seed, false)
	sh.MoonX = lay.MoonX
	v := scenes.NewVista(seed, false)
	v.OwlX = lay.CatX
	n := int(m.secs * float64(m.fps))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		st, t, now := run.at(i, n)
		session := t * m.speed
		st.Act.TimeOfDay = m.todAt(session)
		seg := m.segAt(session)
		if seg.pose != nil {
			st.Pose = *seg.pose
		}
		cat := catOf(seg.who)
		if seg.scape == ScapeVista {
			v.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, c.W-4)
			drawVistaWith(c, v, lay, st, t, cat)
		} else {
			if cat == nil {
				cat = catOf(companion.DefaultName) // an owl on the shore is the crab
			}
			sh.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
			cat.Approach(1/float64(m.fps), st.Pose == companion.NeedsYou)
			drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		}
		frame := c.HTMLFragmentClassed(GIFPx, prof, pal)
		pane := agentPane(c.W, m.agentRows, run.lines)
		out = append(out, pane.HTMLFragmentClassed(GIFPx, prof, pal)+frame)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no frames", m.key)
	}
	return out, nil
}

// heroTurn is one turn of work from session second at0: the prompt, a read
// and a search, a fan-out of three subagents, an edit and a write, a question
// the agent waits on if ask is set, the test, and the finish. The fan-out
// lands early because a subagent stays its dwell (reduce.KittenDwell, 60 s)
// and the finish should come after the litter has gone. Ids are tagged so
// two turns in one session never share one. first says the prompt is the one
// the pane already opens on (newSessionRun), so it is not printed twice.
func heroTurn(at0 float64, tag, prompt, done string, ask, first bool, ctx0 float64) []loopBeat {
	a := func(d float64) float64 { return at0 + d }
	id := func(s string) string { return s + tag }
	var promptPrint []string
	if !first {
		promptPrint = []string{"", "> " + prompt}
	}
	b := []loopBeat{
		{at: a(0), evs: []event.Event{{Kind: event.Prompt, Text: prompt}, todo(0, 5), ctx(ctx0)}, print: promptPrint},
		{at: a(3), evs: []event.Event{
			tool(event.ToolStart, id("t1"), event.OpRead, "Read", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, id("t1"), event.OpRead, "Read", "internal/auth/handler.go", "142 lines"),
		}, print: []string{"*Read\tinternal/auth/handler.go\t142 lines"}},
		{at: a(8), evs: []event.Event{
			tool(event.ToolStart, id("t2"), event.OpSearch, "Grep", "rate.Limiter", ""),
			tool(event.ToolEnd, id("t2"), event.OpSearch, "Grep", "rate.Limiter", "3 files"),
			todo(1, 5),
		}, print: []string{"*Grep\trate.Limiter\t3 files"}},
		{at: a(12), evs: []event.Event{
			sub(id("a1"), "Explore", event.SubStart),
			sub(id("a2"), "Explore", event.SubStart),
			sub(id("a3"), "general-purpose", event.SubStart),
		}, print: []string{"*Task\tExplore x2, general-purpose\t3 agents"}},
		{at: a(18), evs: []event.Event{
			tool(event.ToolStart, id("t3"), event.OpEdit, "Edit", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, id("t3"), event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
			todo(2, 5), ctx(ctx0 + 0.10),
		}, print: []string{"*Edit\tinternal/auth/handler.go\t+18 -2"}},
		{at: a(26), evs: []event.Event{
			tool(event.ToolStart, id("t4"), event.OpWrite, "Write", "internal/auth/limiter.go", ""),
			tool(event.ToolEnd, id("t4"), event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines"),
			todo(3, 5), ctx(ctx0 + 0.16),
		}, print: []string{"*Write\tinternal/auth/limiter.go\t64 lines"}},
	}
	test := a(40)
	if ask {
		// The question holds about thirty seconds, which is what the
		// come-closer walk needs to arrive (loopclips.go).
		b = append(b,
			loopBeat{at: a(34), evs: []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}}, print: []string{"", "!allow Bash?"}},
			loopBeat{at: a(62), evs: []event.Event{{Kind: event.Prompt, Text: "yes"}}, print: []string{"> yes"}},
		)
		test = a(64)
	}
	b = append(b,
		loopBeat{at: test, evs: []event.Event{
			tool(event.ToolStart, id("t5"), event.OpShell, "Bash", "go test ./internal/auth", ""),
			tool(event.ToolEnd, id("t5"), event.OpShell, "Bash", "go test ./internal/auth", "ok 1.4s"),
			todo(4, 5), ctx(ctx0 + 0.28),
		}, print: []string{"*Bash\tgo test ./internal/auth\tok 1.4s"}},
		loopBeat{at: a(74), evs: []event.Event{sub(id("a1"), "Explore", event.SubEnd), sub(id("a2"), "Explore", event.SubEnd)}},
		loopBeat{at: a(78), evs: []event.Event{sub(id("a3"), "general-purpose", event.SubEnd)}},
		loopBeat{at: a(84), evs: []event.Event{todo(5, 5), ctx(ctx0 + 0.34), {Kind: event.Done, Text: done}},
			print: []string{"", "!" + done}},
	)
	return b
}

// heroVariant is one of the six by key; it panics on a name that is not
// there, because the page's hero is built from it.
func heroVariant(key string) montage {
	for _, m := range heroVariants() {
		if m.key == key {
			return m
		}
	}
	panic("no hero variant " + key)
}

// heroVariants is the six candidates for the page's hero, all at the hero's
// own window (124 columns, 14 rows of transcript over a 30-row scape).
// HIS PICK, 2026-09-17: H1.
func heroVariants() []montage {
	done := companion.Done
	base := func(key, label, note string) montage {
		return montage{key: key, label: label, note: note, fps: 6, cols: 124, agentRows: 14, rows: 30}
	}
	// Two turns of one session: the hero's prompt, then shipping it.
	const (
		prompt1 = "add rate limiting to the auth endpoint"
		done1   = "Rate limiting is in. 100 req/min per IP."
		prompt2 = "wire it into the router and ship it"
		done2   = "Shipped. The router limits every route to 100 req/min."
	)
	twoTurns := func(askFirst, askSecond bool) []loopBeat {
		b := heroTurn(2, "a", prompt1, done1, askFirst, true, 0.08)
		b = append(b, heroTurn(92, "b", prompt2, done2, askSecond, false, 0.44)...)
		return append(b, loopBeat{at: 180, evs: []event.Event{ev(event.Compact), ctx(0.06), todo(0, 5)}, clear: true})
	}

	// V1 -- the hero's own session on the vista, the day turning under it.
	v1 := base("v1", "the vista, a whole session", "The hero's session as it is today, with the mountain in place of the shore: the owl, three owlets for the fan-out, the wind and the fire for the work, the moon on its arc as the window fills, a day turning underneath.")
	v1.secs, v1.speed = 15, loopSecs/15.0
	v1.beats = windowLoop()
	v1.stops = []todStop{{0, 0.30}, {loopSecs, 1.30}}
	v1.segs = []heroSeg{{0, ScapeVista, "owl", nil}}

	// V2 -- the same session, evening into night: the peaks take the last
	// light, the fire lights the meadow, stars for the finished items.
	v2 := base("v2", "the vista, evening into night", "The same session held to the vista's best hours: late afternoon into the night, so the alpenglow, the lit fire and the stars all get their turn. No dawn.")
	v2.secs, v2.speed = 15, loopSecs/15.0
	v2.beats = windowLoop()
	// 0.66 is 15:50: the first frames are late afternoon, the alpenglow
	// lands a third of the way in, and the second half is night. (0.58 was
	// tried and was a bright afternoon for half the clip.)
	v2.stops = []todStop{{0, 0.66}, {loopSecs, 1.08}}
	v2.segs = []heroSeg{{0, ScapeVista, "owl", nil}}

	// H1 -- the vista first, then the shore: two turns of one session, the
	// second on the shore with the crab, who walks up to ask.
	h1 := base("h1", "vista, then shore", "Two turns of one session. The owl takes the first on the vista with its owlets; at the second prompt the scape is the shore and the crab takes it, walks up to ask, and the crablets come. One day across both.")
	h1.secs, h1.speed = 22.5, 8
	h1.beats = twoTurns(false, true)
	h1.stops = []todStop{{0, 0.35}, {180, 1.25}}
	h1.segs = []heroSeg{{0, ScapeVista, "owl", nil}, {90, ScapeShore, companion.NameCrab, nil}}

	// H2 -- the shore first, then the vista at night.
	h2 := base("h2", "shore, then vista", "The mirror of H1: the crab's turn on the shore in the afternoon, with the ask and the walk up; then the owl's turn on the vista as the night comes and the fire is lit.")
	h2.secs, h2.speed = 22.5, 8
	h2.beats = twoTurns(true, false)
	h2.stops = []todStop{{0, 0.50}, {180, 1.30}}
	h2.segs = []heroSeg{{0, ScapeShore, companion.NameCrab, nil}, {90, ScapeVista, "owl", nil}}

	// M1 -- the trailer he described. The owl's quick job with owlets; at
	// the finish the owl becomes the cat, then the crab, each showing its
	// finished face; then the scene is the shore at the same hour and the
	// crab's second turn runs into the night.
	trailer := func(key, label, note string, typed bool) montage {
		m := base(key, label, note)
		b := heroTurn(2, "a", prompt1, done1, false, true, 0.08)
		if typed {
			b = append(b,
				loopBeat{at: 90, print: []string{"", "$ xscapes companion cat"}},
				loopBeat{at: 106, print: []string{"$ xscapes companion crab"}},
				loopBeat{at: 122, print: []string{"$ xscapes scape shore"}},
			)
		}
		b = append(b, heroTurn(126, "b", prompt2, done2, true, false, 0.44)...)
		b = append(b, loopBeat{at: 218, evs: []event.Event{ev(event.Compact), ctx(0.06), todo(0, 5)}, clear: true})
		m.beats = b
		m.secs, m.speed = 26, 218/26.0
		m.stops = []todStop{{0, 0.55}, {124, 0.62}, {218, 0.98}}
		m.segs = []heroSeg{
			{0, ScapeVista, "owl", nil},
			{92, ScapeVista, companion.NameCat, &done},
			{108, ScapeVista, companion.NameCrab, &done},
			{124, ScapeShore, companion.NameCrab, nil},
		}
		return m
	}
	m1 := trailer("m1", "the trailer", "His storyboard: the owl does a quick job with three owlets on the vista and finishes; the owl becomes the cat, which shows its finished face; the cat becomes the crab, which shows its own; the scene becomes the shore at the same afternoon hour; the crab's second turn, with the ask, the walk up and the crablets, runs as the sun sets into the night. The changes are cuts.", false)
	m2 := trailer("m2", "the trailer, with the commands typed", "The same trailer, and each change is something the user typed: the transcript shows xscapes companion cat, xscapes companion crab and xscapes scape shore before each cut, which is exactly what the product does.", true)
	return []montage{v1, v2, h1, h2, m1, m2}
}
