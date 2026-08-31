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
	"github.com/donlucasx/asciiscapes/internal/term"
)

// colorPage is the 256-versus-truecolor decision, rendered rather than argued.
//
// Every panel is the SAME frame from the same reducer; only the colour
// treatment differs, and the 256 panels are put through the real quantiser so
// the preview shows what a 256-colour terminal will actually paint rather than
// what the palette wishes it would.
func colorPage(seed int64) string {
	base := time.Now()
	red := reduce.New("colors")
	at := func(s float64) time.Time { return base.Add(time.Duration(s * float64(time.Second))) }
	red.Apply(event.Event{Kind: event.Prompt}, at(0))
	for i, e := range []struct{ op, tool, target, detail string }{
		{"read", "Read", "internal/auth/handler.go", "142 lines"},
		{"search", "Grep", "rate.Limiter", "3 files"},
		{"edit", "Edit", "internal/auth/handler.go", "+18 -2"},
		{"write", "Write", "internal/auth/limiter.go", "64 lines"},
	} {
		red.Apply(event.Event{Kind: event.ToolEnd, ID: fmt.Sprint(i), Op: event.Op(e.op),
			Tool: e.tool, Target: e.target, Detail: e.detail}, at(float64(i)*0.4+1))
	}
	for i := 0; i < 3; i++ {
		red.Apply(event.Event{Kind: event.SubStart, Agent: fmt.Sprint("a", i)}, at(3))
	}

	frame := func(tod float64, blueSky bool, p term.Profile, worried bool) string {
		st := red.State(at(4))
		st.Act.Level, st.Act.Working, st.Act.TimeOfDay = 0.7, true, tod
		st.Pose = companion.Working
		if worried {
			st.Pose = companion.Worried
			if len(st.Tail) > 0 {
				st.Tail[len(st.Tail)-1].Bad = true
			}
		}
		c := canvas.New(78, 20, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(seed, false)
		sh.BlueSky = blueSky
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(78, ccw, true)
		sh.MoonX = lay.MoonX
		for i := 0; i < 24; i++ {
			sh.Update(c, 2+float64(i)/20, st.Act)
		}
		st.Tail = st.FitTail(at(4), lay.SandTo-lay.SandFrom)
		drawScene(c, sh, cat, lay, st, 3.2, seed, c.H-2-chh)
		return c.HTMLFragmentAs(11, p)
	}

	var b strings.Builder
	b.WriteString(`<style>
	.win{border:1px solid #2a2a32;border-radius:5px;overflow:hidden}
	.grid{display:flex;gap:14px;flex-wrap:wrap;align-items:flex-start}
	.col{display:flex;flex-direction:column;gap:5px}
	.lbl{font:11px ui-monospace,monospace;color:#8a8a99}
	h2{font:600 13px ui-monospace,monospace;color:#d8d8e0;margin:30px 0 10px}
	</style>`)
	b.WriteString(`<h1>asciiscapes &mdash; what 256 colours can and cannot do</h1>`)
	b.WriteString(`<p class="nt">The same frame, same reducer, same seed. The 256 panels are put ` +
		`through the real quantiser, so they show what Terminal.app will actually paint.</p>`)

	b.WriteString(`<h2>Midnight &mdash; the everyday case</h2>`)
	fmt.Fprintf(&b, `<div class="grid">
	  <div class="col"><div class="lbl">A &middot; truecolor (Ghostty, iTerm2, WezTerm)</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">B &middot; 256 today &mdash; a monochrome night</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">C &middot; 256 with the sky in the pure-blue column</div><div class="win">%s</div></div>
	</div>`,
		frame(0, false, term.ProfileTrueColor, false),
		frame(0, false, term.Profile256, false),
		frame(0, true, term.Profile256, false))

	b.WriteString(`<h2>Evening &mdash; when Lucas actually ran it</h2>`)
	fmt.Fprintf(&b, `<div class="grid">
	  <div class="col"><div class="lbl">A &middot; truecolor</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">B &middot; 256 today</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">C &middot; 256, blue sky</div><div class="win">%s</div></div>
	</div>`,
		frame(0.888, false, term.ProfileTrueColor, false),
		frame(0.888, false, term.Profile256, false),
		frame(0.888, true, term.Profile256, false))

	b.WriteString(`<h2>The thing that must never be lost</h2>`)
	b.WriteString(`<p class="nt">"Something is broken" is carried entirely by colour. It survives ` +
		`quantisation intact: the worried amber lands on cube index 215, the alert yellow on 222 and ` +
		`the calm cyan on 116 &mdash; three distinct hues, none of them grey. So even the monochrome ` +
		`night keeps its signals, which is arguably the more legible outcome: the only colour on ` +
		`screen is the colour that means something.</p>`)
	fmt.Fprintf(&b, `<div class="grid">
	  <div class="col"><div class="lbl">worried &middot; truecolor</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">worried &middot; 256 &mdash; the amber holds</div><div class="win">%s</div></div>
	</div>`,
		frame(0, false, term.ProfileTrueColor, true),
		frame(0, false, term.Profile256, true))

	return canvas.HTMLPage("asciiscapes — colour study", b.String())
}
