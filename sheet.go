package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

var bubbleCol = term.RGB{R: 224, G: 228, B: 238}

// contactSheet renders every candidate twice: enlarged so the drawing can be
// judged, and composited into a real 80x24 shore so the proportion is honest.
// A companion shown only enlarged always looks better than it is.
func contactSheet(seed int64) string {
	var b strings.Builder
	b.WriteString(`<h1>asciiscapes &mdash; references translated onto the canvas</h1>`)

	for i := range companion.Candidates {
		s := &companion.Candidates[i]
		sw, sh := s.Size()

		big := canvas.New(sw+2, sh+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < big.H; y++ {
			for x := 0; x < big.W; x++ {
				big.SetBG(x, y, term.RGB{R: 27, G: 27, B: 33})
			}
		}
		s.Draw(big.Near(), 1, 1)

		// Real target size, so the height cost is visible rather than argued.
		live := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		scape.NewShore(seed, false).Update(live, 2.0, scape.Activity{})
		top := live.H - 2 - sh
		s.Draw(live.Near(), 8, top)
		if s.Say != "" {
			rows := companion.Bubble(s.Say)
			bub := companion.Sprite{Rows: rows, Body: bubbleCol}
			bub.Draw(live.Near(), 8+sw-2, top-len(rows))
		}

		fmt.Fprintf(&b,
			`<div class="card"><div class="meta"><div class="nm">%s</div>`+
				`<div class="rg">%d rows &middot; %s</div><div class="nt">%s</div></div>`+
				`<div class="big"><div class="lbl">enlarged</div>%s</div>`+
				`<div><div class="lbl">on an 80&times;24 shore, 1:1</div>%s</div></div>`,
			s.Name, sh, s.Source, s.Note,
			big.HTMLFragment(26), live.HTMLFragment(10))
	}
	return canvas.HTMLPage("asciiscapes — translated references", b.String())
}
