package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
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
	"● The handler has no throttling. I'll add a token bucket keyed by client IP,",
	"  with the limit configurable and a sane default.",
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
	"● Rate limiting is in. 100 req/min per IP by default, override with",
	"  AUTH_RATE_LIMIT. Tests pass.",
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

func layoutPage(seed int64) string {
	cat := companion.NewCat()
	_, chh := cat.Size()
	act := scape.Activity{Working: true, Level: 0.6}

	// A: scape as a bottom strip, agent output above it.
	strip := canvas.New(80, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(strip, 3.0, act)
	cat.Draw(strip.Near(), 5, strip.H-2-chh, 3.0, companion.Working)

	// B: scape as a narrow side pane.
	side := canvas.New(30, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	scape.NewShore(seed, false).Update(side, 3.0, act)
	cat.Draw(side.Near(), 3, side.H-2-chh, 3.0, companion.Working)

	var b strings.Builder
	b.WriteString(`<style>` +
		`.tty{margin:0;font-family:Menlo,monospace;font-size:13px;line-height:14px;` +
		`background:#0b0b0d;padding:8px 10px;color:#9aa0a6}` +
		`.tty .you{color:#e6e6ea}.tty .act{color:#8fc7a8}.tty .dim{color:#6e747a}` +
		`.win{border:1px solid #2a2a32;border-radius:6px;overflow:hidden;display:inline-block}` +
		`.split{display:flex}.split>*{flex:none}` +
		`.cap{font-size:12px;color:#6e747a;margin:10px 0 18px;max-width:640px;line-height:1.5}` +
		`</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; where does the scape actually live?</h1>`)

	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">A &middot; bottom strip</div>`+
		`<div class="rg">80&times;14 under the session</div>`+
		`<div class="nt">Scrollback stays readable. Landscape format suits a horizon. `+
		`The scape becomes the floor of the terminal rather than a window beside it.</div></div>`+
		`<div><div class="win">%s%s</div></div></div>`,
		termBlock(fakeSession, 20), strip.HTMLFragment(13))

	fmt.Fprintf(&b, `<div class="card"><div class="meta"><div class="nm">B &middot; side pane</div>`+
		`<div class="rg">30&times;24 beside the session</div>`+
		`<div class="nt">What the brief currently implies. Costs 30 columns of width `+
		`permanently, and a tall narrow frame fights a horizon.</div></div>`+
		`<div><div class="win split">%s%s</div></div></div>`,
		termBlock(fakeSession, 24), side.HTMLFragment(13))

	return canvas.HTMLPage("asciiscapes — layout", b.String())
}
