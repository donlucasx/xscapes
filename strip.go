package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// stripPage stacks every frame vertically with no chrome and a line height
// fixed in pixels, so one screenshot can be sliced into exact frames and
// assembled into a GIF. Capturing frames one at a time would be dozens of
// round trips; this is one.
func stripPage(seed int64, frames int, fps float64, mode string) string {
	const (
		w, h   = 72, 24
		fontPx = 14
	)
	cat := companion.NewCat()

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html><head><meta charset="utf-8"><style>`+
		`html,body{margin:0;padding:0;background:#000}`+
		`pre{margin:0;padding:0;display:block;font-family:Menlo,monospace;`+
		`font-size:%dpx;line-height:%dpx;letter-spacing:0}`+
		`</style></head><body>`, fontPx, fontPx)

	for i := 0; i < frames; i++ {
		t := float64(i) / fps
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)

		switch mode {
		case "walk":
			ww, wh := cat.WalkSize()
			x0, x1 := 4, w-ww-4
			half := frames / 2
			px, dir := x0+(x1-x0)*i/half, 1
			if i >= half {
				px, dir = x1-(x1-x0)*(i-half)/half, -1
			}
			scape.NewShore(seed, false).Update(c, t, scape.Activity{Working: true, Level: 0.5})
			cat.DrawWalk(c.Near(), px, c.H-2-wh, float64(px)*0.8*float64(dir), dir)
		default:
			st := companion.Working
			act := scape.Activity{Working: true, Level: 0.65}
			switch mode {
			case "resting":
				st, act = companion.Resting, scape.Activity{}
			case "needsyou":
				st = companion.NeedsYou
			case "kittens":
				st, act = companion.Working, scape.Activity{Working: true, Level: 0.7}
			case "worried":
				st, act = companion.Worried, scape.Activity{Working: true, Level: 0.35}
			case "tour":
				// One loop through every state, so a single GIF carries the
				// whole vocabulary.
				switch (i * 4) / frames {
				case 0:
					st, act = companion.Working, scape.Activity{Working: true, Level: 0.9}
				case 1:
					st, act = companion.Worried, scape.Activity{Working: true, Level: 0.3}
				case 2:
					st, act = companion.Resting, scape.Activity{}
				default:
					st = companion.NeedsYou
				}
			}
			scape.NewShore(seed, false).Update(c, t, act)
			cw, chh := cat.Size()
			top := c.H - 2 - chh
			cat.Draw(c.Near(), 6, top, t, st)
			if mode == "kittens" {
				cat.DrawKittens(c.Near(), c.Mid(), 6, top, 8, t, seed)
			}
			if st == companion.Worried {
				rows := companion.Bubble("tests failing")
				(&companion.Sprite{Rows: rows, Body: term.RGB{R: 240, G: 200, B: 150}}).
					Draw(c.Near(), 6+cw-2, top-len(rows))
			}
			if st == companion.NeedsYou {
				rows := companion.Bubble("tests passed")
				(&companion.Sprite{Rows: rows, Body: term.RGB{R: 226, G: 230, B: 240}}).
					Draw(c.Near(), 6+cw-2, top-len(rows))
			}
		}
		b.WriteString(c.HTMLFragment(fontPx))
	}
	b.WriteString(`</body></html>`)
	return b.String()
}
