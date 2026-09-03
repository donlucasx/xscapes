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
	"github.com/donlucasx/xscapes/internal/term"
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

	frame := func(tod float64, blueSky bool, p term.Profile, worried bool, boost float64) string {
		saved := term.GlyphBoost
		term.GlyphBoost = boost
		defer func() { term.GlyphBoost = saved }()
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
	b.WriteString(`<h1>xscapes &mdash; what 256 colours can and cannot do</h1>`)
	b.WriteString(`<p class="nt">The same frame, same reducer, same seed. The 256 panels are put ` +
		`through the real quantiser, so they show what Terminal.app will actually paint.</p>`)

	b.WriteString(`<h2>The question: can 256 do a night in colour?</h2>`)
	b.WriteString(`<p class="nt">Terminal.app is not greyscale &mdash; its palette is 216 real ` +
		`colours plus 24 greys. The night came out grey because the PALETTE is dark, and the ` +
		`colour cube has almost no resolution down there: 4 real colours below luma 25, against ` +
		`46 between 110 and 150 and 108 above 150. But a background is meant to be dark, and the ` +
		`glyphs are the bright part of the frame. So put the darkness in the ground and the ` +
		`colour in the texture, which is how ASCII art has always worked. Boosting glyph chroma ` +
		`before quantising takes them from 45% landing on a real colour to 100%.</p>`)

	for _, tod := range []struct {
		v    float64
		name string
	}{{0, "Midnight"}, {0.888, "Evening &mdash; when Lucas actually ran it"}} {
		fmt.Fprintf(&b, `<h2>%s</h2><div class="grid">`, tod.name)
		fmt.Fprintf(&b, `<div class="col"><div class="lbl">truecolor &mdash; the intended look</div><div class="win">%s</div></div>`,
			frame(tod.v, false, term.ProfileTrueColor, false, 1.0))
		for _, k := range []float64{1.0, 2.2} {
			lbl := fmt.Sprintf("256 &middot; glyph chroma %.1f&times;", k)
			if k == 1.0 {
				lbl = "256 &middot; no boost (what you saw)"
			}
			fmt.Fprintf(&b, `<div class="col"><div class="lbl">%s</div><div class="win">%s</div></div>`,
				lbl, frame(tod.v, false, term.Profile256, false, k))
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`<h2>The signals still have to read</h2>`)
	b.WriteString(`<p class="nt">"Something is broken" is carried entirely by colour, and it must ` +
		`not drown in a scene that now has colour everywhere.</p>`)
	fmt.Fprintf(&b, `<div class="grid">
	  <div class="col"><div class="lbl">worried &middot; truecolor</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">worried &middot; 256 no boost</div><div class="win">%s</div></div>
	  <div class="col"><div class="lbl">worried &middot; 256 at 2.0&times;</div><div class="win">%s</div></div>
	</div>`,
		frame(0, false, term.ProfileTrueColor, true, 1.0),
		frame(0, false, term.Profile256, true, 1.0),
		frame(0, false, term.Profile256, true, 2.0))

	return canvas.HTMLPage("xscapes — colour study", b.String())
}
