package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
	"github.com/donlucasx/asciiscapes/internal/reduce"
	"github.com/donlucasx/asciiscapes/internal/scape"
)

// wiredPage renders one simulated session through the real reducer.
//
// Every frame in assets/frames before this one was posed by hand. These are
// not: the same events a Claude Code hook would emit go through the same fold
// the live loop uses, and what comes out is drawn. If the sand collides with
// the companion or the sea reads wrong at some point in a turn, it shows up
// here rather than in a demo recording.
func wiredPage(seed int64) string {
	// A believable turn, in the payload shapes verified out of the Claude Code
	// binary. Times are seconds from the prompt.
	type beat struct {
		at   float64
		note string
		evs  []event.Event
	}
	sub := func(id, kind string, k event.Kind) event.Event {
		return event.Event{Kind: k, Agent: id, AgentType: kind, Op: event.OpSub}
	}
	tool := func(k event.Kind, id string, op event.Op, name, target, detail string, ms int64) event.Event {
		return event.Event{Kind: k, ID: id, Op: op, Tool: name, Target: target, Detail: detail, MS: ms}
	}

	beats := []beat{
		{0, "the prompt lands — thinking, no tool yet", []event.Event{
			{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"},
		}},
		{6, "reading around", []event.Event{
			tool(event.ToolStart, "t1", event.OpRead, "Read", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines", 412),
			tool(event.ToolStart, "t2", event.OpSearch, "Grep", "rate.Limiter", "", 0),
			tool(event.ToolEnd, "t2", event.OpSearch, "Grep", "rate.Limiter", "3 files", 88),
		}},
		{11, "working hard — edits landing", []event.Event{
			tool(event.ToolStart, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2", 120),
			tool(event.ToolStart, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "", 0),
			tool(event.ToolEnd, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines", 95),
			tool(event.ToolStart, "t5", event.OpShell, "Bash", "go", "", 0),
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
		}},
		{30, "the user is asked for something", []event.Event{
			{Kind: event.NeedsInput, Text: "allow Bash?"},
		}},
		{44, "done — the sea settles, the sand still holds the record", []event.Event{
			sub("a1", "code-reviewer", event.SubEnd),
			sub("a2", "general-purpose", event.SubEnd),
			sub("a3", "general-purpose", event.SubEnd),
			{Kind: event.Prompt, Text: "yes"},
			{Kind: event.Done, Text: "Rate limiting is in. 100 req/min per IP."},
		}},
		{72, "half a minute later — flat, the writing receding", nil},
	}

	base := time.Now()
	red := reduce.New("demo")
	// One Shore across every beat, so wave phase integrates the way it does
	// live. A fresh Shore per frame would hide exactly the bug this project
	// just fixed.
	sh := scape.NewShore(seed, false)
	cat := companion.NewCat()
	cat.FaceLeft(true)
	ccw, chh := cat.Size()

	var b strings.Builder
	b.WriteString(`<style>.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}
	.rowline{display:flex;gap:12px;flex-wrap:wrap}</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; driven by real events</h1>`)
	b.WriteString(`<p class="nt">One simulated Claude Code turn, folded by the same reducer the live ` +
		`loop uses. Nothing here is posed.</p>`)

	for _, bt := range beats {
		now := base.Add(time.Duration(bt.at * float64(time.Second)))
		for _, e := range bt.evs {
			red.Apply(e, now)
		}
		st := red.State(now)

		var row strings.Builder
		for _, dt := range []float64{0, 0.35} {
			c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			t := bt.at + dt
			top := c.H - 2 - chh
			lay := compose(c.W, ccw, true)
			sh.MoonX = lay.MoonX
			sh.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
			drawScene(c, sh, cat, lay, st, t, seed, top)
			fmt.Fprintf(&row, `<div class="win">%s</div>`, c.HTMLFragment(11))
		}

		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">t+%.0fs &middot; %s</div>`+
			`<div class="rg">level %.2f &middot; %s &middot; %d kittens &middot; %d sand lines</div>`+
			`<div class="nt">%s</div></div><div class="rowline">%s</div></div>`,
			bt.at, st.Pose, st.Act.Level, poseWord(st.Pose), st.Kittens, len(st.Tail),
			bt.note, row.String())
	}
	return canvas.HTMLPage("asciiscapes — wired", b.String())
}

func poseWord(p companion.State) string { return p.String() }
