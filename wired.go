package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// wiredPage renders one simulated session through the real reducer.
//
// Every frame in assets/frames before this one was posed by hand. These are
// not: the same events a Claude Code hook would emit go through the same fold
// the live loop uses, and what comes out is drawn. If the sand collides with
// the companion or the sea reads wrong at some point in a turn, it shows up
// here rather than in a demo recording.
func wiredPage(seed int64) string {
	beats := demoTurn()

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
	b.WriteString(`<h1>xscapes &mdash; driven by real events</h1>`)
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
	return canvas.HTMLPage("xscapes — wired", b.String())
}

func poseWord(p companion.State) string { return p.String() }
