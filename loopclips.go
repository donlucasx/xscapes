package main

import (
	"github.com/donlucasx/xscapes/internal/event"
)

// A loop clip is a short session written to end where it began, so the GIF
// cuts back to its first frame without a visible jump. The session runs on
// its own accelerated clock (gifScene.speed) while the waves and the
// companion keep real time, which is what lets a whole arc fit in seconds.
type loopBeat struct {
	at    float64 // session seconds
	evs   []event.Event
	print []string // lines to append to the agent's pane, if it has one
	clear bool     // wipe the pane back to the opening prompt first
}

func ev(k event.Kind) event.Event { return event.Event{Kind: k} }

func tool(k event.Kind, id string, op event.Op, name, target, detail string) event.Event {
	return event.Event{Kind: k, ID: id, Op: op, Tool: name, Target: target, Detail: detail}
}

func sub(id, kind string, k event.Kind) event.Event {
	return event.Event{Kind: k, Agent: id, AgentType: kind, Op: event.OpSub}
}

func todo(n, of int) event.Event {
	return event.Event{Kind: event.Todo, Op: event.OpTodo, N: n, Of: of}
}

func ctx(used float64) event.Event {
	return event.Event{Kind: event.Context, Frac: &used}
}

// windowLoop is the hero: one whole session in one loop. The sea rises and
// falls with the work, subagents arrive and swim off, stars light as the plan
// gets done, the moon fills and sinks with the context window, the companion
// asks for permission and later reports that it is finished -- and then the
// context is compacted, the list starts over, and the clip is back where it
// opened. The day turns underneath all of it.
//
// The opening prompt is already on the pane at frame zero, so it is printed
// by the clear at the end rather than by a beat at the start.
func windowLoop() []loopBeat {
	return []loopBeat{
		{at: 2, evs: []event.Event{{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"}, todo(0, 5), ctx(0.06)}},
		{at: 5, evs: []event.Event{
			tool(event.ToolStart, "t1", event.OpRead, "Read", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, "t1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines"),
		}, print: []string{"*Read\tinternal/auth/handler.go\t142 lines"}},
		{at: 11, evs: []event.Event{
			tool(event.ToolStart, "t2", event.OpSearch, "Grep", "rate.Limiter", ""),
			tool(event.ToolEnd, "t2", event.OpSearch, "Grep", "rate.Limiter", "3 files"),
			todo(1, 5),
		}, print: []string{"*Grep\trate.Limiter\t3 files"}},
		{at: 18, evs: []event.Event{
			sub("a1", "Explore", event.SubStart),
			sub("a2", "Explore", event.SubStart),
			sub("a3", "general-purpose", event.SubStart),
		}, print: []string{"*Task\tExplore x2, general-purpose\t3 agents"}},
		{at: 26, evs: []event.Event{
			tool(event.ToolStart, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, "t3", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
			todo(2, 5), ctx(0.24),
		}, print: []string{"*Edit\tinternal/auth/handler.go\t+18 -2"}},
		{at: 34, evs: []event.Event{
			tool(event.ToolStart, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", ""),
			tool(event.ToolEnd, "t4", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines"),
			todo(3, 5),
		}, print: []string{"*Write\tinternal/auth/limiter.go\t64 lines"}},
		{at: 42, evs: []event.Event{sub("a1", "Explore", event.SubEnd), sub("a2", "Explore", event.SubEnd)}},
		{at: 48, evs: []event.Event{
			tool(event.ToolStart, "t5", event.OpShell, "Bash", "go test ./internal/auth", ""),
			tool(event.Error, "t5", event.OpShell, "Bash", "go test ./internal/auth", "exit 1"),
			ctx(0.44),
		}, print: []string{"*Bash\tgo test ./internal/auth\texit 1"}},
		{at: 56, evs: []event.Event{sub("a4", "code-reviewer", event.SubStart), sub("a5", "general-purpose", event.SubStart)}},
		{at: 66, evs: []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}},
			print: []string{"", "!allow Bash?"}},
		{at: 78, evs: []event.Event{{Kind: event.Prompt, Text: "yes"}, ctx(0.62)},
			print: []string{"> yes"}},
		{at: 86, evs: []event.Event{
			tool(event.ToolStart, "t6", event.OpShell, "Bash", "go test ./internal/auth", ""),
			tool(event.ToolEnd, "t6", event.OpShell, "Bash", "go test ./internal/auth", "ok 1.4s"),
			todo(4, 5),
		}, print: []string{"*Bash\tgo test ./internal/auth\tok 1.4s"}},
		{at: 94, evs: []event.Event{
			sub("a3", "general-purpose", event.SubEnd),
			sub("a4", "code-reviewer", event.SubEnd),
			sub("a5", "general-purpose", event.SubEnd),
			ctx(0.82),
		}},
		{at: 104, evs: []event.Event{todo(5, 5), {Kind: event.Done, Text: "Rate limiting is in. 100 req/min per IP."}},
			print: []string{"", "!Rate limiting is in. 100 req/min per IP."}},
		// The window fills, the transcript is compacted, and a fresh list
		// starts: the scene lands back where it opened and the loop closes.
		{at: 122, evs: []event.Event{ev(event.Compact), ctx(0.06), todo(0, 5)}, clear: true},
	}
}

// loopSecs is how long the hero's session runs before it repeats.
const loopSecs = 132

// glanceClips are the small dissected animations beside the legend: the same
// renderer and the same events, cropped to the one part of the scene that
// carries each fact. Each is written to arrive at its point inside four
// seconds.
func glanceClips() []gifScene {
	work := func(id string, at float64) loopBeat {
		return loopBeat{at: at, evs: []event.Event{
			tool(event.ToolStart, id, event.OpEdit, "Edit", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, id, event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
		}}
	}
	base := func(name, note string, crop [4]int, secs float64, beats []loopBeat) gifScene {
		return gifScene{name: name, note: note, tod: 0.55, secs: secs, speed: 6, crop: crop, beats: beats}
	}
	// The moon and the constellation are held at a visibility floor and are
	// washed out at midday by design, so their own clips run at night.
	night := func(name, note string, crop [4]int, secs float64, beats []loopBeat) gifScene {
		sc := base(name, note, crop, secs, beats)
		sc.tod = 0.94
		return sc
	}
	return []gifScene{
		// The companion and the balloon, right of frame.
		base("g-ask", "the balloon rises and the companion turns to face you", [4]int{44, 10, 80, 21}, 4, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			work("t1", 4), work("t2", 8),
			{at: 12, evs: []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}}},
		}),
		base("g-done", "the finish knock, and the companion holds the pose", [4]int{44, 10, 80, 21}, 4, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			work("t1", 4), work("t2", 8),
			{at: 12, evs: []event.Event{{Kind: event.Done, Text: "Rate limiting is in."}}},
		}),
		base("g-worried", "a command exits 1 and the companion carries it", [4]int{44, 10, 80, 21}, 4, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			work("t1", 4),
			{at: 10, evs: []event.Event{tool(event.Error, "t2", event.OpShell, "Bash", "go", "exit 1")}},
		}),
		// The open water, left of frame.
		base("g-sea", "the sea gets up as the work does, and settles after", [4]int{2, 8, 38, 19}, 5, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			work("t1", 4), work("t2", 6), work("t3", 8), work("t4", 10), work("t5", 12),
			{at: 16, evs: []event.Event{{Kind: event.Done}}},
		}),
		// The waterline and the writing under it.
		base("g-sand", "each tool call is written at the waterline, and taken", [4]int{0, 13, 36, 24}, 5, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			{at: 4, evs: []event.Event{tool(event.ToolEnd, "t1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines")}},
			{at: 9, evs: []event.Event{tool(event.ToolEnd, "t2", event.OpSearch, "Grep", "rate.Limiter", "3 files")}},
			{at: 14, evs: []event.Event{tool(event.ToolEnd, "t3", event.OpEdit, "Edit", "internal/auth/limiter.go", "+18 -2")}},
			{at: 19, evs: []event.Event{tool(event.ToolEnd, "t4", event.OpShell, "Bash", "go test", "ok")}},
		}),
		// The disc, high left, with its readout.
		night("g-moon", "the disc wanes and sinks as the window fills", [4]int{4, 0, 40, 11}, 6, moonRamp()),
		// The beach beside the companion, where the litter is.
		base("g-kittens", "one kitten per subagent, and each swims off when it is done", [4]int{40, 11, 76, 22}, 5, []loopBeat{
			{at: 2, evs: []event.Event{{Kind: event.Prompt}}},
			{at: 4, evs: []event.Event{sub("a1", "Explore", event.SubStart), sub("a2", "Explore", event.SubStart)}},
			{at: 8, evs: []event.Event{sub("a3", "general-purpose", event.SubStart), sub("a4", "code-reviewer", event.SubStart)}},
			{at: 16, evs: []event.Event{sub("a1", "Explore", event.SubEnd), sub("a2", "Explore", event.SubEnd)}},
			{at: 22, evs: []event.Event{sub("a3", "general-purpose", event.SubEnd), sub("a4", "code-reviewer", event.SubEnd)}},
		}),
		// The upper sky, where the constellation is.
		// A whole day over the water, which is the one channel that is not the
		// agent at all.
		{name: "g-sky", note: "the sky on your own clock, a day in five seconds",
			tod: 0.30, todEnd: 1.30, secs: 5, speed: 6, crop: [4]int{0, 0, 36, 11},
			beats: []loopBeat{{at: 1, evs: []event.Event{{Kind: event.Prompt}}}}},
		night("g-stars", "a star for every finished item, and a new list clears them", [4]int{34, 0, 70, 11}, 5, []loopBeat{
			{at: 1, evs: []event.Event{{Kind: event.Prompt}, todo(0, 6)}},
			{at: 4, evs: []event.Event{todo(1, 6)}},
			{at: 8, evs: []event.Event{todo(2, 6)}},
			{at: 12, evs: []event.Event{todo(3, 6)}},
			{at: 16, evs: []event.Event{todo(4, 6)}},
			{at: 20, evs: []event.Event{todo(5, 6)}},
			{at: 24, evs: []event.Event{todo(6, 6)}},
		}),
	}
}

// moonRamp fills the context window in small steps so the disc visibly wanes
// and sinks rather than jumping, then compacts and comes back whole.
func moonRamp() []loopBeat {
	out := []loopBeat{{at: 1, evs: []event.Event{{Kind: event.Prompt}, ctx(0.04)}}}
	for i := 1; i <= 14; i++ {
		out = append(out, loopBeat{at: 1 + float64(i)*2, evs: []event.Event{ctx(0.04 + 0.062*float64(i))}})
	}
	out = append(out, loopBeat{at: 32, evs: []event.Event{ev(event.Compact), ctx(0.04)}})
	return out
}
