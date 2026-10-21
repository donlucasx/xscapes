package main

// The shore clip (2026-09-21): the sea is the work. For his reply to "the
// flames really drive the vibe": "The tides shows you how hard your agents
// are working on the shore scape". Nothing here ships in the product; like
// fire_test.go it hands notes/qt/render-fire.mjs (--name shore) the same
// markup the site plays.
//
//	XSCAPES_SHORE=<dir> [XSCAPES_SHORE_SIZE=80x27] [XSCAPES_SHORE_SECS=11] go test -run TestShoreExport .
//
// writes <dir>/shore.js: window.TRAILER.shore = {cols, rows, fps, secs, frames, css},
// the scape alone (no agent rows).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// shoreMontage is the site's own shore clip (sceneClips "shore": a prompt,
// two crablets, the work rising through a read, an edit, a write and the
// test, the finish) with a second and a half of low tide first, so the sea is
// seen at rest before it rises, and three seconds after the finish to settle.
// Late afternoon, held (the site's 0.62): the sea reads best by day and the
// hour is not the point here.
func shoreMontage(cols, rows int, secs float64) montage {
	m := montage{key: "shore", label: "the shore", fps: 30, cols: cols, agentRows: 0, rows: rows}
	// v2 (his note, 2026-09-21 ~16:20: "fix the beginning of the animation so
	// it starts with low tide, and ramps up quickly so we can see activity and
	// focus on the tide"): speed 6 (the site's clip runs at 4), the prompt at
	// 1.5 s, so the sea is full within two seconds of it.
	const speed = 6.0                                                // session seconds per clip second
	const lead = 1.5                                                 // clip seconds of low tide before the prompt
	a := func(d float64) float64 { return (lead + d/speed) * speed } // the site's beat seconds, shifted
	m.beats = []loopBeat{
		{at: a(1), evs: []event.Event{{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"}, todo(0, 3), ctx(0.30)}},
		{at: a(3), evs: []event.Event{
			tool(event.ToolStart, "s1", event.OpRead, "Read", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, "s1", event.OpRead, "Read", "internal/auth/handler.go", "142 lines"),
		}},
		{at: a(6), evs: []event.Event{sub("sa1", "Explore", event.SubStart), sub("sa2", "Explore", event.SubStart)}},
		{at: a(10), evs: []event.Event{
			tool(event.ToolStart, "s2", event.OpEdit, "Edit", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, "s2", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
			todo(1, 3),
		}},
		{at: a(16), evs: []event.Event{
			tool(event.ToolStart, "s3", event.OpWrite, "Write", "internal/auth/limiter.go", ""),
			tool(event.ToolEnd, "s3", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines"),
			todo(2, 3), ctx(0.38),
		}},
		{at: a(26), evs: []event.Event{
			tool(event.ToolStart, "s4", event.OpShell, "Bash", "go test ./internal/auth", ""),
			tool(event.ToolEnd, "s4", event.OpShell, "Bash", "go test ./internal/auth", "ok 1.4s"),
		}},
		{at: a(38), evs: []event.Event{todo(3, 3), {Kind: event.Done, Text: "Rate limiting is in."}}},
	}
	m.secs, m.speed = secs, speed
	m.stops = []todStop{{0, 0.62}}
	m.segs = []heroSeg{{0, ScapeShore, companion.NameCrab, nil}}
	return m
}

// shoreFrames is montageFrames' shore branch without the agent pane
// (agentPane cannot be zero rows) and without the vista.
func shoreFrames(seed int64, m montage, pal *canvas.HTMLPalette) ([]string, error) {
	sc := gifScene{name: m.key, secs: m.secs, speed: m.speed, fps: m.fps, cols: m.cols,
		rows: m.rows, beats: m.beats, pal: pal, cube: !truecolorFX}
	run := newSessionRun(sc)
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	c := canvas.New(m.cols, m.rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cat := companion.New(companion.NameCrab)
	cat.FaceLeft(true)
	ccw, chh := companion.New(companion.DefaultName).Size()
	lay := compose(c.W, ccw, true)
	sh := scape.NewShore(seed, false)
	sh.MoonX = lay.MoonX
	// TWO RENDER LIBERTIES, restored on return, neither in the product.
	// (1) The shore's tide state starts at zero withdrawal, which is HIGH
	// tide, and eases toward low tide over TideEase (3 s), so a clip that
	// opens on a quiet session opens high and recedes (his "so it starts
	// with low tide"). Fifteen quiet seconds of Update before the first frame
	// settle it at low tide: each one-second step closes two thirds of the
	// gap, (1/3)^15 is nothing. (2) TideEase 3 s -> 1 s, so the tide arrives
	// within a couple of seconds of the work instead of nine.
	wasEase := scape.TideEase
	scape.TideEase = 1.0
	defer func() { scape.TideEase = wasEase }()
	{
		st0, _, _ := run.at(0, 1)
		quiet := st0.Act
		quiet.Level = 0
		quiet.TimeOfDay = m.todAt(0)
		for k := 15; k >= 1; k-- {
			sh.Update(c, -float64(k), quiet)
		}
	}
	n := int(m.secs * float64(m.fps))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		st, t, now := run.at(i, n)
		st.Act.TimeOfDay = m.todAt(t * m.speed)
		sh.Update(c, t, st.Act)
		st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
		cat.Approach(1/float64(m.fps), st.Pose == companion.NeedsYou)
		drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		out = append(out, c.HTMLFragmentClassed(GIFPx, prof, pal))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no frames", m.key)
	}
	return out, nil
}

func TestShoreExport(t *testing.T) {
	dir := os.Getenv("XSCAPES_SHORE")
	if dir == "" {
		t.Skip("set XSCAPES_SHORE=<dir> to write the shore clip's frames")
	}
	cols, rows := 80, 27
	if s := os.Getenv("XSCAPES_SHORE_SIZE"); s != "" {
		if _, err := fmt.Sscanf(s, "%dx%d", &cols, &rows); err != nil {
			t.Fatal(err)
		}
	}
	secs := 11.0
	if s := os.Getenv("XSCAPES_SHORE_SECS"); s != "" {
		if _, err := fmt.Sscan(s, &secs); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pal := &canvas.HTMLPalette{}
	m := shoreMontage(cols, rows, secs)
	frames, err := shoreFrames(7, m, pal)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(map[string]any{
		"cols": m.cols, "rows": m.rows, "fps": m.fps, "secs": m.secs, "speed": m.speed,
		"frames": frames, "css": pal.CSS(),
	})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "shore.js")
	if err := os.WriteFile(p, []byte("window.TRAILER=window.TRAILER||{};window.TRAILER.shore="+string(b)+";"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("shore: %dx%d, %d frames (%.1f s at %d fps, speed %.1f), %d bytes", cols, rows, len(frames), secs, m.fps, m.speed, len(b))
}
