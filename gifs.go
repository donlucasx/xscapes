package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
)

// gifScene is one animated clip of the demo turn: which beat it starts on,
// at what hour, for how long. Frames come out of the same reducer fold the
// site's stills came from, so nothing in them is posed.
type gifScene struct {
	name string
	at   float64 // the beat, seconds from the prompt
	tod  float64
	secs float64
	note string
}

var gifScenes = []gifScene{
	{"hero", 14, 0.52, 6, "the fan-out at noon: busy sea, five kittens, the sand carrying the tool calls"},
	{"worried", 22, 0.62, 4, "a command exited 1: the companion carries it"},
	{"ask", 30, 0.80, 4, "dusk: the agent needs permission, solid balloon, alert pose"},
	{"done", 44, 0.96, 5, "night: done, dotted balloon, the constellation, the moon with its readout"},
	{"resting", 72, 0.27, 8, "dawn, half a minute later: flat sea, the writing receding, the litter swims off as its dwell ends"},
}

// GIFFPS is the clip rate. Ten a second is the scape's own cadence rounded
// down; the sea and the companion read at it, and the pages stay under the
// height a headless capture will take.
const GIFFPS = 10

// GIFPx is the cell's font size in the clips. 14px Menlo is 8.4 by 14
// pixels a cell, 672 by 336 for a frame: crisp at the page's width, and a
// clip still under a megabyte.
const GIFPx = 14

// gifPages writes one HTML page per scene into dir, each holding every frame
// of the clip stacked under a magenta separator, plus manifest.json. A
// capture script (site/make-gifs.py) screenshots each page once, slices it
// on the separators and encodes the GIF. Frames are 80 columns by 24 rows
// at GIFPx Menlo, in truecolor.
func gifPages(seed int64, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	type entry struct {
		Name   string  `json:"name"`
		Frames int     `json:"frames"`
		FPS    int     `json:"fps"`
		Note   string  `json:"note"`
		Secs   float64 `json:"secs"`
		Px     int     `json:"px"`
	}
	var manifest []entry
	for _, sc := range gifScenes {
		frames, err := gifFrames(seed, sc)
		if err != nil {
			return err
		}
		var b strings.Builder
		fmt.Fprintf(&b, `<!doctype html><html><head><meta charset="utf-8"><style>`+
			`body{margin:0;background:#000}`+
			`.f{display:inline-block;width:max-content;height:%dpx;overflow:hidden;vertical-align:top;border-right:3px solid #ff00ff;font-family:Menlo,"SF Mono","DejaVu Sans Mono",monospace;font-size:%dpx;line-height:1}`+
			`.f pre{margin:0;font-family:inherit;font-size:%dpx;line-height:1;letter-spacing:0;display:block}`+
			`.sep{width:100%%;height:4px;background:#ff00ff}`+
			`</style></head><body>`, 24*GIFPx, GIFPx, GIFPx)
		for _, f := range frames {
			b.WriteString(`<div class="sep"></div><div><div class="f">`)
			b.WriteString(f)
			b.WriteString(`</div></div>`)
		}
		b.WriteString(`<div class="sep"></div></body></html>`)
		if err := os.WriteFile(filepath.Join(dir, sc.name+".html"), []byte(b.String()), 0o644); err != nil {
			return err
		}
		manifest = append(manifest, entry{sc.name, len(frames), GIFFPS, sc.note, sc.secs, GIFPx})
	}
	js, _ := json.MarshalIndent(manifest, "", "  ")
	return os.WriteFile(filepath.Join(dir, "manifest.json"), js, 0o644)
}

// gifFrames folds the demo turn up to the scene's beat and then renders the
// clip frame by frame, applying any beat that falls inside the clip as its
// time passes.
func gifFrames(seed int64, sc gifScene) ([]string, error) {
	base := time.Now()
	red := reduce.New("site")
	sh := scape.NewShore(seed, false)
	cat := companion.NewCat()
	cat.FaceLeft(true)
	ccw, chh := cat.Size()
	beats := demoTurn()
	applied := 0
	apply := func(upto float64) {
		for applied < len(beats) && beats[applied].at <= upto {
			now := base.Add(time.Duration(beats[applied].at * float64(time.Second)))
			for _, e := range beats[applied].evs {
				red.Apply(e, now)
			}
			applied++
		}
	}
	apply(sc.at)
	// Settle the sea's phase before the first frame, as the stills do.
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	lay := compose(c.W, ccw, true)
	sh.MoonX = lay.MoonX
	n := int(sc.secs * GIFFPS)
	var out []string
	for i := 0; i < n; i++ {
		t := sc.at + float64(i)/GIFFPS
		apply(t)
		now := base.Add(time.Duration(t * float64(time.Second)))
		st := red.State(now)
		st.Act.TimeOfDay = sc.tod
		sh.Update(c, t, st.Act)
		st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
		drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		// Truecolor, his direction of 2026-09-05: the page shows the scene at
		// its best, not as the cube rounds it. The product still runs on
		// the cube; the clips are what it is reaching for.
		out = append(out, c.HTMLFragment(GIFPx))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scene %s rendered no frames", sc.name)
	}
	return out, nil
}
