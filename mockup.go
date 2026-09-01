package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// mockupPage is the composition study: the current left-anchored layout beside
// the mirrored one, then the mirrored one across the terminal shapes a person
// actually has.
//
// Both arms go through the same drawScene the live loop uses, and the state
// comes out of the real reducer, so this cannot flatter the proposal by
// drawing it differently from how it would ship.
func mockupPage(seed int64) string {
	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden}
	.pair{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99;margin:0 0 4px 2px}
	.col{display:flex;flex-direction:column}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:26px 0 8px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; which side does the companion sit on?</h1>`)
	b.WriteString(`<p class="nt">Left column is what ships today. Right column is the mirror: ` +
		`companion on the right, litter growing leftward, tail written from the left margin, ` +
		`moon moved to 0.28 so the two things a glance looks for are not stacked in one column. ` +
		`Same reducer, same draw path, same seed.</p>`)

	// One frame, both compositions, at a given litter size and canvas.
	frame := func(w, h int, st reduce.State, mirror bool, t float64) string {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		cat := companion.NewCat()
		cat.FaceLeft(mirror)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, mirror)
		sh.MoonX = lay.MoonX
		// Run a few frames so wave phase is settled rather than at t=0.
		for i := 0; i < 30; i++ {
			sh.Update(c, t-1.5+float64(i)/20, st.Act)
		}
		sh.Update(c, t, st.Act)
		st.Tail = st.FitTail(time.Now(), lay.SandTo-lay.SandFrom)
		drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		return c.HTMLFragment(11)
	}

	// A state with a given number of subagents and a believable sand tail.
	build := func(kittens int, pose companion.State, bubble string) reduce.State {
		red := reduce.New("mockup")
		base := time.Now()
		at := func(s float64) time.Time { return base.Add(time.Duration(s * float64(time.Second))) }
		red.Apply(event.Event{Kind: event.Prompt}, at(0))
		tail := []struct{ op, tool, target, detail string }{
			{"read", "Read", "internal/auth/handler.go", "142 lines"},
			{"search", "Grep", "rate.Limiter", "3 files"},
			{"edit", "Edit", "internal/auth/handler.go", "+18 -2"},
			{"write", "Write", "internal/auth/limiter.go", "64 lines"},
		}
		for i, e := range tail {
			red.Apply(event.Event{Kind: event.ToolEnd, ID: fmt.Sprint(i), Op: event.Op(e.op),
				Tool: e.tool, Target: e.target, Detail: e.detail}, at(float64(i)*0.4+1))
		}
		for i := 0; i < kittens; i++ {
			red.Apply(event.Event{Kind: event.SubStart, Agent: fmt.Sprint("a", i)}, at(3))
		}
		if pose == companion.Worried {
			red.Apply(event.Event{Kind: event.Error, Op: event.OpShell, Tool: "Bash", Target: "go", Detail: "exit 1"}, at(3.5))
		}
		if bubble != "" {
			red.Apply(event.Event{Kind: event.NeedsInput, Text: bubble}, at(3.6))
		}
		st := red.State(at(4))
		st.Act.Level = 0.72
		st.Act.Working = true
		return st
	}

	card := func(title, note string, body string) {
		fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">%s</div>`+
			`<div class="nt">%s</div></div>%s</div>`, title, note, body)
	}

	b.WriteString(`<h2>The decision: how many kittens fit before they reach the text</h2>`)
	for _, cse := range []struct {
		n    int
		note string
	}{
		{0, "no subagents — the everyday case"},
		{3, "three — the tail and the litter should not touch"},
		{5, "five — about where the beach runs out on the right"},
		{9, "nine — the litter has shrunk a tier and starts meeting the text"},
		{14, "fourteen — overlap, which is the agreed behaviour at this size"},
	} {
		st := build(cse.n, companion.Working, "")
		card(fmt.Sprintf("%d subagents", cse.n), cse.note,
			fmt.Sprintf(`<div class="pair">`+
				`<div class="col"><div class="lbl">today — companion left</div><div class="win">%s</div></div>`+
				`<div class="col"><div class="lbl">mirrored — companion right</div><div class="win">%s</div></div>`+
				`</div>`,
				frame(80, 24, st, false, 4), frame(80, 24, st, true, 4)))
	}

	b.WriteString(`<h2>The two states that move the most furniture</h2>`)
	for _, cse := range []struct {
		pose   companion.State
		bubble string
		note   string
	}{
		{companion.NeedsYou, "allow Bash?", "the bubble has to open away from the edge it is anchored to"},
		{companion.Worried, "", "something is broken — the companion carries it, not the weather"},
	} {
		st := build(3, cse.pose, cse.bubble)
		card(cse.pose.String(), cse.note,
			fmt.Sprintf(`<div class="pair">`+
				`<div class="col"><div class="lbl">today — companion left</div><div class="win">%s</div></div>`+
				`<div class="col"><div class="lbl">mirrored — companion right</div><div class="win">%s</div></div>`+
				`</div>`,
				frame(80, 24, st, false, 4), frame(80, 24, st, true, 4)))
	}

	b.WriteString(`<h2>Every shape a real terminal comes in</h2>`)
	b.WriteString(`<p class="nt">The scene scales, but the composition is fixed fractions of ` +
		`HEIGHT, so height is what decides whether there is a beach to write in at all. ` +
		`Widths stretch freely. Row count is the number that matters.</p>`)
	st := build(4, companion.Working, "")
	for _, d := range []struct {
		w, h int
		note string
	}{
		{80, 24, "the design target"},
		{120, 30, "a comfortable modern window"},
		{200, 50, "full screen on a big monitor — lots of beach and sky"},
		{100, 20, "a tmux side split — the intended home"},
		{160, 14, "a wide bottom split — the beach nearly disappears"},
		{60, 45, "a narrow tall split — plenty of beach, the tail runs out of width"},
		{40, 12, "the stated minimum"},
	} {
		card(fmt.Sprintf("%d&times;%d", d.w, d.h), d.note,
			fmt.Sprintf(`<div class="pair"><div class="win">%s</div></div>`, frame(d.w, d.h, st, true, 4)))
	}
	return canvas.HTMLPage("asciiscapes — composition study", b.String())
}
