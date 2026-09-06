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
	"github.com/donlucasx/xscapes/internal/term"
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

	// cols is the frame's width in cells; 0 means the design size, 80.
	cols int
	// agentRows, when set, puts that many rows of the agent's own output
	// above the scape, so the clip is the whole terminal window rather than
	// the scape alone. The scape keeps its 24 rows underneath.
	agentRows int
}

const gifRows = 24

// rowsOf is the frame's total height in cells.
func (sc gifScene) rowsOf() int { return gifRows + sc.agentRows }

// colsOf is the frame's width in cells.
func (sc gifScene) colsOf() int {
	if sc.cols > 0 {
		return sc.cols
	}
	return 80
}

var gifScenes = []gifScene{
	{name: "window", at: 14, tod: 0.52, secs: 6, cols: 108, agentRows: 16,
		note: "the whole window at noon: the agent's own output above, the scape below it"},
	{name: "hero", at: 14, tod: 0.52, secs: 6, note: "the fan-out at noon: busy sea, five kittens, the sand carrying the tool calls"},
	{name: "worried", at: 22, tod: 0.62, secs: 4, note: "a command exited 1: the companion carries it"},
	{name: "ask", at: 30, tod: 0.80, secs: 4, note: "dusk: the agent needs permission, solid balloon, alert pose"},
	{name: "done", at: 44, tod: 0.96, secs: 5, note: "night: done, dotted balloon, the constellation, the moon with its readout"},
	{name: "resting", at: 72, tod: 0.27, secs: 8, note: "dawn, half a minute later: flat sea, the writing receding, the litter swims off as its dwell ends"},
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
		Cols   int     `json:"cols"`
		Rows   int     `json:"rows"`
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
			`</style></head><body>`, sc.rowsOf()*GIFPx, GIFPx, GIFPx)
		for _, f := range frames {
			b.WriteString(`<div class="sep"></div><div><div class="f">`)
			b.WriteString(f)
			b.WriteString(`</div></div>`)
		}
		b.WriteString(`<div class="sep"></div></body></html>`)
		if err := os.WriteFile(filepath.Join(dir, sc.name+".html"), []byte(b.String()), 0o644); err != nil {
			return err
		}
		manifest = append(manifest, entry{sc.name, len(frames), GIFFPS, sc.note, sc.secs, GIFPx, sc.colsOf(), sc.rowsOf()})
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
	c := canvas.New(sc.colsOf(), gifRows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
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
		frame := c.HTMLFragment(GIFPx)
		if sc.agentRows > 0 {
			// The agent's own output sits above its scape, in the same window,
			// which is the thing the page is actually claiming.
			frame = agentPane(sc.colsOf(), sc.agentRows).HTMLFragment(GIFPx) + frame
		}
		out = append(out, frame)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scene %s rendered no frames", sc.name)
	}
	return out, nil
}

// agentPane draws the agent's half of the window: the prompt that started the
// turn and the tool calls the demo turn has made by the time the clip opens,
// printed the way an agent prints them, with the cursor waiting at the bottom.
//
// It is a canvas rather than markup so it goes through the same renderer and
// lands on the same grid as the scape below it, cell for cell.
func agentPane(cols, rows int) *canvas.Canvas {
	c := canvas.New(cols, rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	c.Clear()
	var (
		bg     = term.RGB{R: 26, G: 26, B: 26}
		bright = term.RGB{R: 238, G: 238, B: 238}
		tool   = term.RGB{R: 135, G: 175, B: 255}
		body   = term.RGB{R: 200, G: 200, B: 200}
		dim    = term.RGB{R: 122, G: 122, B: 122}
		rule   = term.RGB{R: 58, G: 58, B: 58}
	)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			c.SetBG(x, y, bg)
		}
	}
	near := c.Near()
	put := func(x, y int, s string, col term.RGB) {
		for i, r := range []rune(s) {
			if x+i < cols && y < rows {
				near.Plot(x+i, y, r, col, 1)
			}
		}
	}
	// The same events the scape below is folding, printed as output.
	type line struct{ kind, target, detail string }
	calls := []line{
		{"Read", "internal/auth/handler.go", "142 lines"},
		{"Grep", "rate.Limiter", "3 files"},
		{"Edit", "internal/auth/handler.go", "+18 -2"},
		{"Write", "internal/auth/limiter.go", "64 lines"},
		{"Bash", "go test ./internal/auth", "running"},
		{"Task", "code-reviewer, general-purpose x2, Explore x2", "5 agents"},
	}
	y := 1
	put(2, y, ">", dim)
	put(4, y, "add rate limiting to the auth endpoint", bright)
	y += 2
	for _, l := range calls {
		put(2, y, "*", dim)
		put(4, y, l.kind, tool)
		put(12, y, l.target, body)
		if col := cols - 2 - len([]rune(l.detail)); col > 12+len([]rune(l.target))+2 {
			put(col, y, l.detail, dim)
		}
		y++
	}
	y++
	put(2, y, "thinking", dim)
	for i := 0; i < 3; i++ {
		put(11+i, y, ".", dim)
	}
	// The input line, with the cursor the mark is made of.
	y = rows - 2
	for x := 1; x < cols-1; x++ {
		near.Plot(x, y-1, '_', rule, 0.5)
	}
	put(2, y, ">", dim)
	c.SetBG(4, y, bright)
	return c
}
