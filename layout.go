package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// A plausible slice of agent output, so the mockups show the scape competing
// with real text rather than with lorem ipsum.
var fakeSession = []string{
	"&gt; add rate limiting to the auth endpoint",
	"",
	"● I'll add rate limiting. Let me read the current handler first.",
	"",
	"● Read(internal/auth/handler.go)",
	"  ⎿  Read 142 lines",
	"",
	"● Edit(internal/auth/handler.go)",
	"  ⎿  Added 18 lines, removed 2",
	"",
	"● Edit(internal/auth/limiter.go)",
	"  ⎿  Created 64 lines",
	"",
	"● Bash(go test ./internal/auth/...)",
	"  ⎿  ok    github.com/acme/api/internal/auth    0.412s",
	"",
	"● Rate limiting is in. 100 req/min per IP by default.",
}

// activity is the tail of what the agent is doing, oldest first.
var activity = []string{
	"read  internal/auth/handler.go    142 lines",
	"edit  internal/auth/handler.go    +18 -2",
	"write internal/auth/limiter.go    64 lines",
	"bash  go test ./internal/auth/    ok 0.412s",
}

func termBlock(lines []string, rows int) string {
	var b strings.Builder
	b.WriteString(`<pre class="tty">`)
	shown := lines
	if len(shown) > rows {
		shown = shown[len(shown)-rows:]
	}
	for _, l := range shown {
		cls := "dim"
		switch {
		case strings.HasPrefix(l, "&gt;"):
			cls = "you"
		case strings.HasPrefix(l, "●"):
			cls = "act"
		}
		fmt.Fprintf(&b, `<span class="%s">%s</span>`+"\n", cls, l)
	}
	for i := len(shown); i < rows; i++ {
		b.WriteString("\n")
	}
	b.WriteString(`</pre>`)
	return b.String()
}

// writeInSand renders the activity tail into the beach itself, in sand tones
// rather than terminal tones. Newest is brightest; older lines fade toward the
// sand as the tide takes them. Always visible, but scenery rather than a HUD.
func writeInSand(c *canvas.Canvas, lines []string, sandTop int) {
	sand := term.RGB{R: 76, G: 65, B: 54}
	ink := term.RGB{R: 240, G: 230, B: 210}
	n := len(lines)
	den := n - 1
	if den < 1 {
		den = 1
	}
	for i, ln := range lines {
		row := sandTop + i
		if row >= c.H {
			break
		}
		age := float64(n-1-i) / float64(den)
		col := term.Lerp(ink, sand, age*0.70)
		(&companion.Sprite{Rows: []string{ln}, Body: col, Alpha: 1, Opaque: true}).Draw(c.Near(), 3, row)
	}
}

func layoutPage(seed int64) string {
	cat := companion.NewCat()
	_, chh := cat.Size()
	act := scape.Activity{Working: true, Level: 0.6}

	// A: scape on top, the agent output as a strip beneath it.
	top := canvas.New(80, 16, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(top, 3.0, act)
	cat.Draw(top.Near(), 5, top.H-2-chh, 3.0, companion.Working)

	// B: one full scape, activity tail written into the sand.
	full := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(full, 3.0, act)
	cat.Draw(full.Near(), 5, full.H-2-chh-4, 3.0, companion.Working)
	writeInSand(full, activity, 20)

	// C: same, plus the newest event riding in on driftwood.
	drift := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(drift, 3.0, act)
	cat.Draw(drift.Near(), 5, drift.H-2-chh-4, 3.0, companion.Working)
	writeInSand(drift, activity[:3], 21)
	(&companion.Sprite{
		Rows:  []string{"__/\\__", "limiter.go"},
		Body:  term.RGB{R: 232, G: 220, B: 196},
		Alpha: 1,
	}).Draw(drift.Near(), 50, 17)

	var b strings.Builder
	b.WriteString(`<style>` +
		`.tty{margin:0;font-family:Menlo,monospace;font-size:13px;line-height:14px;` +
		`background:#0b0b0d;padding:6px 10px;color:#9aa0a6}` +
		`.tty .you{color:#e6e6ea}.tty .act{color:#8fc7a8}.tty .dim{color:#6e747a}` +
		`.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden;display:inline-block}` +
		`</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; where does the agent's work appear?</h1>`)

	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">A &middot; strip below</div>`+
		`<div class="rg">scape 80&times;16 + activity 80&times;7</div>`+
		`<div class="nt">Literal and unambiguous. Costs nothing in legibility and gains `+
		`nothing in originality: a screensaver with a status pane.</div></div>`+
		`<div><div class="win">%s%s</div></div></div>`,
		top.HTMLFragment(13), termBlock(fakeSession, 7))

	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">B &middot; written in the sand</div>`+
		`<div class="rg">one 80&times;24 scape</div>`+
		`<div class="nt">The activity tail IS the beach. Newest brightest; older lines fade into `+
		`the sand as the tide takes them. Always visible, but scenery, not a HUD.</div></div>`+
		`<div><div class="win">%s</div></div></div>`, full.HTMLFragment(13))

	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">C &middot; sand + driftwood</div>`+
		`<div class="rg">B, plus the newest event washing in</div>`+
		`<div class="nt">The current file arrives on driftwood; the tide leaves it in the sand. `+
		`The log writes itself by physics.</div></div>`+
		`<div><div class="win">%s</div></div></div>`, drift.HTMLFragment(13))

	return canvas.HTMLPage("asciiscapes — layout", b.String())
}
