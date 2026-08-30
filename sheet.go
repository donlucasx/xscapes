package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// contactSheet renders every candidate twice: enlarged so the drawing can be
// judged, and composited into the shore at 1:1 so it can be judged in the only
// place that matters. Showing one without the other misleads.
func contactSheet(seed int64) string {
	var b strings.Builder
	b.WriteString(`<h1>asciiscapes &mdash; companion candidates</h1>`)

	for i := range companion.Candidates {
		s := &companion.Candidates[i]
		sw, sh := s.Size()

		// Enlarged, on a neutral ground.
		big := canvas.New(sw+2, sh+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		for y := 0; y < big.H; y++ {
			for x := 0; x < big.W; x++ {
				big.SetBG(x, y, term.RGB{R: 27, G: 27, B: 33})
			}
		}
		s.Draw(big.Near(), 1, 1)

		// 1:1 on the shore, standing on the sand just back from the waterline.
		live := canvas.New(46, 18, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sc := scape.NewShore(seed, false)
		sc.Update(live, 2.0, scape.Activity{})
		s.Draw(live.Near(), 6, live.H-sh-1)

		fmt.Fprintf(&b,
			`<div class="card"><div class="meta"><div class="nm">%s</div>`+
				`<div class="rg">%s</div><div class="nt">%s</div></div>`+
				`<div class="big"><div class="lbl">enlarged</div>%s</div>`+
				`<div><div class="lbl">1:1 on the shore</div>%s</div></div>`,
			s.Name, s.Register, s.Note,
			big.HTMLFragment(30), live.HTMLFragment(13))
	}
	return canvas.HTMLPage("asciiscapes — companion candidates", b.String())
}
