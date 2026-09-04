package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// turnBeat is one moment of the demo turn: seconds from the prompt, a note,
// and the events that land then.
type turnBeat struct {
	at   float64
	note string
	evs  []event.Event
}

// demoTurn is the one simulated Claude Code turn every rendered study shares:
// a prompt, some reading, edits, a fan-out of five subagents, a failing
// command, a question, the finish, and half a minute of quiet. The payload
// shapes are the ones verified out of the Claude Code binary.
//
// It lives in one place so the wired study and the submission page cannot
// drift apart: a beat changed here changes in both.
func demoTurn() []turnBeat {
	sub := func(id, kind string, k event.Kind) event.Event {
		return event.Event{Kind: k, Agent: id, AgentType: kind, Op: event.OpSub}
	}
	tool := func(k event.Kind, id string, op event.Op, name, target, detail string, ms int64) event.Event {
		return event.Event{Kind: k, ID: id, Op: op, Tool: name, Target: target, Detail: detail, MS: ms}
	}
	todo := func(n, of int) event.Event {
		return event.Event{Kind: event.Todo, Op: event.OpTodo, N: n, Of: of}
	}
	return []turnBeat{
		{0, "the prompt lands — thinking, no tool yet", []event.Event{
			{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"},
			todo(0, 5),
		}},
		{6, "reading around", []event.Event{
			tool(event.ToolStart, "t1", event.OpRead, "Read", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines", 412),
			tool(event.ToolStart, "t2", event.OpSearch, "Grep", "rate.Limiter", "", 0),
			tool(event.ToolEnd, "t2", event.OpSearch, "Grep", "rate.Limiter", "3 files", 88),
			todo(1, 5),
		}},
		{11, "working hard — edits landing", []event.Event{
			tool(event.ToolStart, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2", 120),
			tool(event.ToolStart, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "", 0),
			tool(event.ToolEnd, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines", 95),
			tool(event.ToolStart, "t5", event.OpShell, "Bash", "go", "", 0),
			todo(2, 5),
		}},
		{14, "a fan-out — five subagents", []event.Event{
			sub("a1", "code-reviewer", event.SubStart),
			sub("a2", "general-purpose", event.SubStart),
			sub("a3", "general-purpose", event.SubStart),
			sub("a4", "Explore", event.SubStart),
			sub("a5", "Explore", event.SubStart),
		}},
		{22, "a test fails — the companion carries it, not the weather", []event.Event{
			sub("a4", "Explore", event.SubEnd),
			sub("a5", "Explore", event.SubEnd),
			tool(event.Error, "t5", event.OpShell, "Bash", "go", "exit 1", 4100),
			todo(3, 5),
		}},
		{30, "the user is asked for something", []event.Event{
			{Kind: event.NeedsInput, Text: "allow Bash?"},
		}},
		{44, "done — the sea settles, the sand still holds the record", []event.Event{
			sub("a1", "code-reviewer", event.SubEnd),
			sub("a2", "general-purpose", event.SubEnd),
			sub("a3", "general-purpose", event.SubEnd),
			{Kind: event.Prompt, Text: "yes"},
			todo(5, 5),
			{Kind: event.Done, Text: "Rate limiting is in. 100 req/min per IP."},
		}},
		{72, "half a minute later — flat, the writing receding", nil},
	}
}

// siteFrame names one frame of the submission page: which beat of the demo
// turn, at what time of day, and the marker in the template it replaces.
type siteFrame struct {
	marker string
	at     float64 // the beat, by its seconds from the prompt
	dt     float64 // seconds after the beat, so a wave is mid-travel
	tod    float64 // 0 midnight, .25 dawn, .5 noon, .75 dusk
}

var siteFrames = []siteFrame{
	{"hero", 14, 0.35, 0.52},   // the fan-out at noon: busy sea, kittens, sand
	{"resting", 72, 0.2, 0.27}, // flat at dawn, the writing receding
	{"worried", 22, 0.5, 0.62}, // the failed command, afternoon
	{"ask", 30, 0.3, 0.80},     // the question, dusk
	{"done", 44, 0.3, 0.96},    // the finish, night, the constellation full
}

// sitePage fills site/template.html with frames from the real reducer,
// rendered as a 256-colour terminal would show them.
//
// 256 on purpose. The page is for people deciding whether to install this,
// and the target terminal is Terminal.app; a truecolor preview would show a
// picture no user gets. Nothing in the frames is posed: the same events a
// Claude Code hook emits go through the same fold the live loop uses.
func sitePage(seed int64, dir string) (string, error) {
	tmpl, err := os.ReadFile(filepath.Join(dir, "template.html"))
	if err != nil {
		return "", err
	}
	page := string(tmpl)

	base := time.Now()
	red := reduce.New("site")
	sh := scape.NewShore(seed, false)
	cat := companion.NewCat()
	cat.FaceLeft(true)
	ccw, chh := cat.Size()

	// Fold the whole turn once, in order, and pick the frames off as their
	// beats pass. The shore is stateful, so the frames must come out of one
	// timeline, the way the wired study does it.
	wanted := map[float64][]siteFrame{}
	for _, f := range siteFrames {
		wanted[f.at] = append(wanted[f.at], f)
	}
	var missing []string
	var pal canvas.HTMLPalette
	for _, bt := range demoTurn() {
		now := base.Add(time.Duration(bt.at * float64(time.Second)))
		for _, e := range bt.evs {
			red.Apply(e, now)
		}
		for _, f := range wanted[bt.at] {
			st := red.State(now)
			st.Act.TimeOfDay = f.tod
			c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			t := bt.at + f.dt
			lay := compose(c.W, ccw, true)
			sh.MoonX = lay.MoonX
			sh.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
			drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
			marker := "{{" + f.marker + "}}"
			if !strings.Contains(page, marker) {
				missing = append(missing, marker)
				continue
			}
			page = strings.Replace(page, marker, c.HTMLFragmentClassed(12, term.Profile256, &pal), 1)
		}
	}
	if !strings.Contains(page, "{{palette}}") {
		missing = append(missing, "{{palette}}")
	}
	page = strings.Replace(page, "{{palette}}", pal.CSS(), 1)
	if len(missing) > 0 {
		return "", fmt.Errorf("template has no %s", strings.Join(missing, ", "))
	}
	if i := strings.Index(page, "{{"); i >= 0 {
		end := i + 40
		if end > len(page) {
			end = len(page)
		}
		return "", fmt.Errorf("template marker not filled: %q", page[i:end])
	}
	return page, nil
}
