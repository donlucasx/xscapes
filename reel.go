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

// reelPage stacks the frames of one simulated turn vertically, with no chrome
// and a line height fixed in pixels, so a single screenshot slices into exact
// frames for a GIF. Capturing frames one at a time would be dozens of round
// trips; this is one.
//
// The turn runs through the REAL reducer at a real frame rate, so what the GIF
// shows is what the live loop would paint -- not a hand-posed sequence.
func reelPage(seed int64, from, count int, fps float64) string {
	const (
		w, h   = 76, 22
		fontPx = 13
	)

	// One turn, compressed enough to fit a loop: prompt, work, a fan-out, a
	// failure, a question, the finish. Times are scene seconds.
	type beat struct {
		at  float64
		evs []event.Event
	}
	tool := func(k event.Kind, id string, op event.Op, name, target, detail string, ms int64) event.Event {
		return event.Event{Kind: k, ID: id, Op: op, Tool: name, Target: target, Detail: detail, MS: ms}
	}
	sub := func(id string, k event.Kind) event.Event {
		return event.Event{Kind: k, Agent: id, AgentType: "general-purpose", Op: event.OpSub}
	}
	beats := []beat{
		{0.4, []event.Event{{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"}}},
		{1.6, []event.Event{
			tool(event.ToolStart, "t1", event.OpRead, "Read", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines", 412),
		}},
		{2.6, []event.Event{
			tool(event.ToolStart, "t2", event.OpSearch, "Grep", "rate.Limiter", "", 0),
			tool(event.ToolEnd, "t2", event.OpSearch, "Grep", "rate.Limiter", "3 files", 88),
		}},
		{3.6, []event.Event{
			tool(event.ToolStart, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "", 0),
			tool(event.ToolEnd, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2", 120),
		}},
		{4.6, []event.Event{
			tool(event.ToolStart, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "", 0),
			tool(event.ToolEnd, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines", 95),
			sub("a1", event.SubStart), sub("a2", event.SubStart), sub("a3", event.SubStart),
			sub("a4", event.SubStart), sub("a5", event.SubStart),
		}},
		{7.0, []event.Event{
			tool(event.ToolStart, "t5", event.OpShell, "Bash", "go", "", 0),
		}},
		{8.6, []event.Event{
			sub("a4", event.SubEnd), sub("a5", event.SubEnd),
			tool(event.Error, "t5", event.OpShell, "Bash", "go", "exit 1", 1600),
		}},
		{10.4, []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}}},
		{13.0, []event.Event{
			sub("a1", event.SubEnd), sub("a2", event.SubEnd), sub("a3", event.SubEnd),
			{Kind: event.Prompt, Text: "yes"},
			{Kind: event.Done, Text: "Rate limiting is in. 100 req/min per IP."},
		}},
	}

	base := time.Now()
	red := reduce.New("reel")
	sh := scape.NewShore(seed, false)
	cat := companion.NewCat()
	cat.FaceLeft(true)
	ccw, chh := cat.Size()

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html><head><meta charset="utf-8"><style>`+
		`html,body{margin:0;padding:0;background:#0b0b10}`+
		`pre{margin:0;padding:0;display:block;font-family:Menlo,monospace;`+
		`font-size:%dpx;line-height:%dpx;letter-spacing:0}`+
		`</style></head><body>`, fontPx, fontPx)

	next := 0
	for i := 0; i < from+count; i++ {
		t := float64(i) / fps
		now := base.Add(time.Duration(t * float64(time.Second)))
		for next < len(beats) && beats[next].at <= t {
			for _, e := range beats[next].evs {
				red.Apply(e, now)
			}
			next++
		}
		st := red.State(now)
		st.Act.TimeOfDay = 0 // hold midnight; the sky is not what this shows

		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		lay := compose(w, ccw, true)
		sh.MoonX = lay.MoonX
		sh.Update(c, t, st.Act)
		st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
		drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)

		// The shore is stateful, so every frame has to be computed in order;
		// only the requested window is written out.
		if i >= from {
			b.WriteString(c.HTMLFragment(fontPx))
		}
	}
	b.WriteString(`</body></html>`)
	return b.String()
}
