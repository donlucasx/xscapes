package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scape"
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
	ctx := func(used float64) event.Event {
		return event.Event{Kind: event.Context, Frac: &used}
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
			// The session is past 40% of its context here, so the readout
			// under the moon is on from this beat (his ruling, 2026-09-05).
			ctx(0.46),
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

// sitePage writes the submission page: the template as it is, with the
// animated clips rendered beside it into <dir>/anim by gifPages. The clips
// are the demo turn folded through the real reducer, as a 256-colour
// terminal shows it; site/make-gifs.py then captures and encodes them.
// The page used to carry five stills lifted from the same fold; his
// direction of 2026-09-05: animated clips, no still screens.
func sitePage(seed int64, dir string) (string, error) {
	tmpl, err := os.ReadFile(filepath.Join(dir, "template.html"))
	if err != nil {
		return "", err
	}
	page := string(tmpl)
	cover := coverLayers(seed)
	page = strings.Replace(page, "{{cover}}", cover, 1)
	if err := os.WriteFile(filepath.Join(dir, "anim", "cover.html"), []byte(cover), 0o644); err != nil {
		return "", err
	}
	if i := strings.Index(page, "{{"); i >= 0 {
		end := min(i+40, len(page))
		return "", fmt.Errorf("template marker not filled: %q", page[i:end])
	}
	if err := gifPages(seed, filepath.Join(dir, "anim")); err != nil {
		return "", err
	}
	return page, nil
}

// coverLayers is the cover's background: six frames of the real shore at
// night, glyphs only, stacked as <pre> layers a CSS step animation cycles.
// The mark sits over its own sea. Written to <dir>/anim/cover.html too, so
// the deck's title slide can carry the same.
func coverLayers(seed int64) string {
	const w, h = 168, 44
	var b strings.Builder
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: 0.93}
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for i := 0; i < 30; i++ {
		sh.Update(c, float64(i)/20, act)
	}
	for i := 0; i < 6; i++ {
		sh.Update(c, 2+float64(i)*0.7, act)
		fmt.Fprintf(&b, `<pre style="animation-delay:-%.1fs">%s</pre>`, float64(i)*0.7, html.EscapeString(c.RenderPlain()))
	}
	return b.String()
}
