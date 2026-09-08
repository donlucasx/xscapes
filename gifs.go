package main

import (
	"encoding/json"
	"fmt"
	"math"
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

	// beats, when set, replaces the shared demo turn with a session written
	// for this clip and written to end where it began.
	beats []loopBeat
	// speed is session seconds per second of clip. The waves and the
	// companion keep real time either way; only the session is wound on, so
	// a whole arc fits in a few seconds without the water looking sped up.
	speed float64
	// todEnd, when set, runs the time of day from tod to todEnd across the
	// clip instead of holding it still.
	todEnd float64
	// crop, when set, is the cell rectangle of the scape the frame keeps:
	// x0, y0, x1, y1. It is how one channel is shown on its own.
	crop [4]int
}

const gifRows = 24

// cropped reports whether this clip keeps only part of the scape.
func (sc gifScene) cropped() bool { return sc.crop != [4]int{} }

// rowsOf is the frame's height in cells, after any crop.
func (sc gifScene) rowsOf() int {
	if sc.cropped() {
		return sc.crop[3] - sc.crop[1]
	}
	return gifRows + sc.agentRows
}

// colsOf is the frame's width in cells, after any crop.
func (sc gifScene) colsOf() int {
	if sc.cropped() {
		return sc.crop[2] - sc.crop[0]
	}
	if sc.cols > 0 {
		return sc.cols
	}
	return 80
}

// sceneCols is the width the scape is rendered at, before any crop.
func (sc gifScene) sceneCols() int {
	if sc.cols > 0 {
		return sc.cols
	}
	return 80
}

// gifPageCap is the tallest page a headless capture will take without
// trouble. A long clip is split across several, each screenshotted and
// sliced on its own; the frames are concatenated in order.
const gifPageCap = 15000

func (sc gifScene) framesPerPage() int {
	n := gifPageCap / (sc.rowsOf()*GIFPx + 4)
	if n < 1 {
		n = 1
	}
	return n
}

var gifScenes = []gifScene{
	{name: "window", tod: 0.30, todEnd: 1.30, secs: 16, cols: 108, agentRows: 16,
		speed: loopSecs / 16.0, beats: windowLoop(),
		note: "one whole session in one loop, the day turning under it"},
	{name: "hero", at: 14, tod: 0.52, secs: 6, note: "the fan-out at noon: busy sea, five crablets, the sand carrying the tool calls"},
	{name: "worried", at: 22, tod: 0.62, secs: 4, note: "a command exited 1: the companion carries it"},
	{name: "ask", at: 30, tod: 0.80, secs: 4, note: "dusk: the agent needs permission, solid balloon, alert pose"},
	{name: "done", at: 44, tod: 0.96, secs: 5, note: "night: done, dotted balloon, the constellation, the moon with its readout"},
	{name: "resting", at: 72, tod: 0.27, secs: 8, note: "dawn, half a minute later: flat sea, the writing receding, the litter swims off as its dwell ends"},
}

// allGIFScenes is every clip the site wants: the five of the demo turn plus
// the dissected ones beside the legend.
func allGIFScenes() []gifScene { return append(append([]gifScene{}, gifScenes...), glanceClips()...) }

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
		Pages  int     `json:"pages"`
	}
	var manifest []entry
	for _, sc := range allGIFScenes() {
		frames, err := gifFrames(seed, sc)
		if err != nil {
			return err
		}
		per := sc.framesPerPage()
		pages := 0
		for from := 0; from < len(frames); from += per {
			to := min(from+per, len(frames))
			var b strings.Builder
			fmt.Fprintf(&b, `<!doctype html><html><head><meta charset="utf-8"><style>`+
				`body{margin:0;background:#000}`+
				`.f{display:inline-block;width:max-content;height:%dpx;overflow:hidden;vertical-align:top;border-right:3px solid #ff00ff;font-family:Menlo,"SF Mono","DejaVu Sans Mono",monospace;font-size:%dpx;line-height:1}`+
				`.f pre{margin:0;font-family:inherit;font-size:%dpx;line-height:1;letter-spacing:0;display:block}`+
				`.sep{width:100%%;height:4px;background:#ff00ff}`+
				`</style></head><body>`, sc.rowsOf()*GIFPx, GIFPx, GIFPx)
			for _, f := range frames[from:to] {
				b.WriteString(`<div class="sep"></div><div><div class="f">`)
				b.WriteString(f)
				b.WriteString(`</div></div>`)
			}
			b.WriteString(`<div class="sep"></div></body></html>`)
			name := fmt.Sprintf("%s.%d.html", sc.name, pages)
			if err := os.WriteFile(filepath.Join(dir, name), []byte(b.String()), 0o644); err != nil {
				return err
			}
			pages++
		}
		manifest = append(manifest, entry{sc.name, len(frames), GIFFPS, sc.note, sc.secs, GIFPx, sc.colsOf(), sc.rowsOf(), pages})
	}
	js, _ := json.MarshalIndent(manifest, "", "  ")
	return os.WriteFile(filepath.Join(dir, "manifest.json"), js, 0o644)
}

// gifFrames renders a clip frame by frame. A clip either walks the shared
// demo turn from one of its beats, or runs a session of its own written to
// end where it began; either way the waves and the companion keep real time
// and only the session is wound on, by sc.speed.
func gifFrames(seed int64, sc gifScene) ([]string, error) {
	base := time.Now()
	red := reduce.New("site")
	sh := scape.NewShore(seed, false)
	// The SHIPPED default, not this machine's saved preference: the page has to
	// render the same on any checkout, and it has to show what a fresh install
	// actually gets. companionPref() would make the entry depend on whatever is
	// in ~/.config/xscapes/companion, which is the one thing a submission must
	// not do.
	cat := companion.New(companion.DefaultName)
	cat.FaceLeft(true)
	ccw, chh := cat.Size()
	speed := sc.speed
	if speed <= 0 {
		speed = 1
	}

	// The pane opens on the prompt that started the turn; beats add to it.
	lines := []string{"> add rate limiting to the auth endpoint"}
	applied := 0
	var apply func(upto float64)
	if sc.beats != nil {
		apply = func(upto float64) {
			for applied < len(sc.beats) && sc.beats[applied].at <= upto {
				b := sc.beats[applied]
				now := base.Add(time.Duration(b.at * float64(time.Second)))
				for _, e := range b.evs {
					red.Apply(e, now)
				}
				if b.clear {
					lines = lines[:1]
				}
				lines = append(lines, b.print...)
				applied++
			}
		}
	} else {
		beats := demoTurn()
		apply = func(upto float64) {
			for applied < len(beats) && beats[applied].at <= upto {
				now := base.Add(time.Duration(beats[applied].at * float64(time.Second)))
				for _, e := range beats[applied].evs {
					red.Apply(e, now)
				}
				applied++
			}
		}
		apply(sc.at)
	}

	c := canvas.New(sc.sceneCols(), gifRows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	lay := compose(c.W, ccw, true)
	sh.MoonX = lay.MoonX
	n := int(sc.secs * GIFFPS)
	var out []string
	for i := 0; i < n; i++ {
		// t is the picture's own clock, in real seconds, so the swell and the
		// breathing read at their designed rate whatever the session does.
		t := float64(i) / GIFFPS
		session := t * speed
		if sc.beats == nil {
			t = sc.at + t
			session = t
		}
		apply(session)
		now := base.Add(time.Duration(session * float64(time.Second)))
		st := red.State(now)
		st.Act.TimeOfDay = sc.tod
		if sc.todEnd != 0 {
			tod := sc.tod + (sc.todEnd-sc.tod)*float64(i)/float64(n)
			st.Act.TimeOfDay = tod - math.Floor(tod)
		}
		sh.Update(c, t, st.Act)
		st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
		drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		// Truecolor, his direction of 2026-09-05: the page shows the scene at
		// its best, not as the cube rounds it. The product still runs on the
		// cube; the clips are what it is reaching for.
		var frame string
		if sc.cropped() {
			frame = c.HTMLFragmentCropAs(sc.crop[0], sc.crop[1], sc.crop[2], sc.crop[3], GIFPx, term.ProfileTrueColor)
		} else {
			frame = c.HTMLFragment(GIFPx)
		}
		if sc.agentRows > 0 {
			// The agent's own output sits above its scape, in the same
			// window, which is the thing the page is actually claiming.
			frame = agentPane(sc.sceneCols(), sc.agentRows, lines).HTMLFragment(GIFPx) + frame
		}
		out = append(out, frame)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scene %s rendered no frames", sc.name)
	}
	return out, nil
}

// agentPane draws the agent's half of the window: what it has printed so far,
// scrolled so the newest is at the bottom, with the cursor waiting under it.
//
// It is a canvas rather than markup so it goes through the same renderer and
// lands on the same grid as the scape below it, cell for cell.
func agentPane(cols, rows int, lines []string) *canvas.Canvas {
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
			if x+i < cols-1 && y >= 0 && y < rows {
				near.Plot(x+i, y, r, col, 1)
			}
		}
	}
	// The transcript scrolls: the last however many lines fit above the
	// input row, oldest first.
	room := rows - 4
	from := 0
	if len(lines) > room {
		from = len(lines) - room
	}
	y := 1
	for _, l := range lines[from:] {
		switch {
		case strings.HasPrefix(l, "> "):
			put(2, y, ">", dim)
			put(4, y, l[2:], bright)
		case strings.HasPrefix(l, "*"):
			// A tool call: name, target, and the result against the right
			// edge, the way an agent lays them out.
			f := strings.SplitN(l[1:], "\t", 3)
			put(2, y, "*", dim)
			put(4, y, f[0], tool)
			if len(f) > 1 {
				put(13, y, f[1], body)
			}
			if len(f) > 2 {
				if col := cols - 2 - len([]rune(f[2])); col > 14+len([]rune(f[1])) {
					put(col, y, f[2], dim)
				}
			}
		case strings.HasPrefix(l, "!"):
			put(4, y, l[1:], bright)
		default:
			put(4, y, strings.TrimSpace(l), dim)
		}
		y++
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
