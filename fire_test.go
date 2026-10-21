package main

// The fire clip (2026-09-21): the vista at night, the fire lit, the owl and
// its owlets working through two turns while the context fills, the moon's
// gauge emptying, then a compact so the loop closes. For a post while the
// hackathon's winners are pending ("watching our agents work while we wait
// for the winners"). Nothing here ships in the product; this file is the
// only seam, and like trailer_test.go it hands notes/trailer/render.mjs
// (clip mode) the same markup the site plays.
//
//	XSCAPES_FIRE=<dir> [XSCAPES_FIRE_SIZE=82x28] [XSCAPES_FIRE_SECS=21.5] go test -run TestFireExport .
//
// writes <dir>/fire.js: window.TRAILER.fire = {cols, rows, fps, frames, css},
// the scape alone (no agent rows), 16:9 when cols ~= 2.95 * rows (Menlo's
// advance is 0.602 em, a row is one em, so at 1080 px tall the frame holds
// rows * 2.95 columns across 1920).
//
// THE ENCODING STANDS. In the product the fire is the work (with the wind)
// and the moon is the context (the vista's gauge on the arc): this clip
// shows both as they are, and the caption says which is which.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// fireFrames is montageFrames without the agent's pane (agentPane cannot be
// zero rows: it slices the transcript by rows-4) and without the shore: the
// vista alone, every frame, the session wound on by the clip's speed and the
// wind and the owl keeping real time.
func fireFrames(seed int64, m montage, pal *canvas.HTMLPalette) ([]string, error) {
	sc := gifScene{name: m.key, secs: m.secs, speed: m.speed, fps: m.fps, cols: m.cols,
		rows: m.rows, beats: m.beats, pal: pal, cube: !truecolorFX}
	run := newSessionRun(sc)
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	c := canvas.New(m.cols, m.rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	ccw, _ := companion.New(companion.DefaultName).Size()
	lay := compose(c.W, ccw, true)
	v := scenes.NewVista(seed, false)
	v.OwlX = lay.CatX
	n := int(m.secs * float64(m.fps))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		st, t, now := run.at(i, n)
		st.Act.TimeOfDay = m.todAt(t * m.speed)
		v.Update(c, t, st.Act)
		st.Tail = st.FitTail(now, c.W-4)
		drawVistaWith(c, v, lay, st, t, nil)
		out = append(out, c.HTMLFragmentClassed(GIFPx, prof, pal))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no frames", m.key)
	}
	return out, nil
}

// fireMontage is a second and a half of nothing (his note, 2026-09-21: "should see
// the fire activity pickup from nothing to WORKING within the scene": still
// air, a straight column of smoke), then the hero's two turns (heroTurn: a
// read, a search, a fan-out of three owlets, an edit, a write, the test, the
// finish), the context from 30% to 98% so the readout appears at 40% and
// turns warm at 85%, then the compact at the end that empties it and clears
// the stars, which is what closes the loop. The hour runs from dusk into the
// night behind it (his "a timelapse behind as the day gets darker"), so the
// fire's light comes up on the meadow as the session goes.
func fireMontage(cols, rows int, secs float64) montage {
	m := montage{key: "fire", label: "by the fire", fps: 30, cols: cols, agentRows: 0, rows: rows}
	const (
		prompt1 = "add rate limiting to the auth endpoint"
		done1   = "Rate limiting is in. 100 req/min per IP."
		prompt2 = "wire it into the router and ship it"
		done2   = "Shipped. The router limits every route to 100 req/min."
	)
	// v2 (his note, 2026-09-21 ~14:00: "the video should ramp up faster- takes
	// a bit too long for the bonfire to start burning and zoom out"): the
	// prompt at 1.5 s instead of 4, the session at speed 10 instead of the
	// trailer's 8, so the flames stream within a second and a half of it.
	const speed = 10.0 // session seconds per clip second
	b := heroTurn(1.5*speed, "a", prompt1, done1, false, true, 0.30)
	b = append(b, heroTurn(11*speed, "b", prompt2, done2, false, false, 0.64)...) // continues where the first turn ended (0.30 + 0.34)
	b = append(b, loopBeat{at: 20.5 * speed, evs: []event.Event{ev(event.Compact), ctx(0.06), todo(0, 5)}, clear: true})
	m.beats = b
	m.secs, m.speed = secs, speed
	m.stops = []todStop{{0, 0.70}, {secs * speed, 1.05}}
	m.segs = []heroSeg{{0, ScapeVista, "owl", nil}}
	return m
}

func TestFireExport(t *testing.T) {
	dir := os.Getenv("XSCAPES_FIRE")
	if dir == "" {
		t.Skip("set XSCAPES_FIRE=<dir> to write the fire clip's frames")
	}
	cols, rows := 82, 28
	if s := os.Getenv("XSCAPES_FIRE_SIZE"); s != "" {
		if _, err := fmt.Sscanf(s, "%dx%d", &cols, &rows); err != nil {
			t.Fatal(err)
		}
	}
	secs := 21.5
	if s := os.Getenv("XSCAPES_FIRE_SECS"); s != "" {
		if _, err := fmt.Sscan(s, &secs); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pal := &canvas.HTMLPalette{}
	m := fireMontage(cols, rows, secs)
	frames, err := fireFrames(7, m, pal)
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
	p := filepath.Join(dir, "fire.js")
	if err := os.WriteFile(p, []byte("window.TRAILER=window.TRAILER||{};window.TRAILER.fire="+string(b)+";"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("fire: %dx%d, %d frames (%.1f s at %d fps, speed %.2f), %d bytes", cols, rows, len(frames), secs, m.fps, m.speed, len(b))
}
