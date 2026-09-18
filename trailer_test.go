package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// TestTrailerExport writes the frames the X trailer is cut from (his ask,
// 2026-09-17: a release trailer for X). Nothing here ships in the product:
// the cut itself is notes/trailer (a browser stage, a camera track, the
// audio and the ffmpeg assembly), and this test is only the seam that hands
// it the same markup the site plays. Skipped unless the directory is named,
// so the suite never writes anywhere.
//
//	XSCAPES_TRAILER=<dir> go test -run TestTrailerExport .
//
// XSCAPES_TRAILER_PART=cover|hero writes one part; both by default. It writes:
//
//   - cover.js: the splash's own sea (site.go coverLayersAt: the shore at
//     night, six frames 0.7 s apart) twice over -- the bare glyphs the site
//     shows, and the same glyphs each in its own ink on nothing, which is
//     what "in full colour" means for a field the page draws in one grey.
//   - hero.js: H1, the site's hero, at 30 fps instead of the page's 6 (the
//     page pays for every frame in markup; a video does not), with the clip
//     seconds of every ask, every finish and every scape cut, so the cues
//     and the camera land on the product's own moments.
//   - palette.css: the classes hero.js's frames use.
func TestTrailerExport(t *testing.T) {
	dir := os.Getenv("XSCAPES_TRAILER")
	if dir == "" {
		t.Skip("set XSCAPES_TRAILER=<dir> to write the trailer's frames")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	part := os.Getenv("XSCAPES_TRAILER_PART")
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if part == "" || part == "cover" {
		rows := 57 // fills 1080 rows at the cell the sea's 168 columns give 1920
		if s := os.Getenv("XSCAPES_TRAILER_COVER_ROWS"); s != "" {
			if _, err := fmt.Sscan(s, &rows); err != nil {
				t.Fatal(err)
			}
		}
		frames, horizon := trailerCover(7, 168, rows)
		b, err := json.Marshal(map[string]any{
			"cols": 168, "rows": rows, "step": 0.7, "horizon": horizon, "frames": frames,
		})
		if err != nil {
			t.Fatal(err)
		}
		write("cover.js", "window.TRAILER=window.TRAILER||{};window.TRAILER.cover="+string(b)+";")
	}
	if part == "" || part == "hero" {
		pal := &canvas.HTMLPalette{}
		m := trailerMontage()
		frames, err := trailerFrames(7, m, pal)
		if err != nil {
			t.Fatal(err)
		}
		var asks, dones, cuts []float64
		for _, b := range m.beats {
			for _, e := range b.evs {
				switch e.Kind {
				case event.NeedsInput:
					asks = append(asks, b.at/m.speed)
				case event.Done:
					dones = append(dones, b.at/m.speed)
				}
			}
		}
		for _, s := range m.segs[1:] {
			cuts = append(cuts, s.at/m.speed)
		}
		b, err := json.Marshal(map[string]any{
			"cols": m.cols, "rows": m.rows + m.agentRows, "agentRows": m.agentRows, "fps": m.fps,
			"secs": m.secs, "speed": m.speed, "asks": asks, "dones": dones, "cuts": cuts,
			"frames": frames,
		})
		if err != nil {
			t.Fatal(err)
		}
		write("hero.js", "window.TRAILER=window.TRAILER||{};window.TRAILER.hero="+string(b)+";")
		write("palette.css", pal.CSS())
		fmt.Fprintf(os.Stderr, "hero: %d frames, asks %v, dones %v, cuts %v (clip seconds)\n", len(frames), asks, dones, cuts)
	}
	if part == "vista" {
		// The vista over a whole day, for a GIF under the post (his ask,
		// 2026-09-18 ~12:10): the site's own V1 variant, the hero's window,
		// at a GIF's frame rate. Self-contained: frames and palette in one.
		pal := &canvas.HTMLPalette{}
		m := heroVariant("v1")
		m.fps = 12
		frames, err := montageFrames(7, m, pal)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(map[string]any{
			"cols": m.cols, "rows": m.rows + m.agentRows, "fps": m.fps, "frames": frames, "css": pal.CSS(),
		})
		if err != nil {
			t.Fatal(err)
		}
		write("vista.js", "window.TRAILER=window.TRAILER||{};window.TRAILER.vista="+string(b)+";")
	}
	if part == "" || part == "cat" {
		frames, work, cols, rows, css, err := trailerCat()
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(map[string]any{
			"cols": cols, "rows": rows, "fps": 30, "workFrames": work, "frames": frames, "css": css,
		})
		if err != nil {
			t.Fatal(err)
		}
		write("cat.js", "window.TRAILER=window.TRAILER||{};window.TRAILER.cat="+string(b)+";")
	}
}

// trailerMontage is H1's session re-timed for the cut (his notes on the
// first full cut, 2026-09-18 ~02:00): the first wide shot runs until the
// owlets come, the owlets hide before the finish, the night sets in on the
// vista before the cut to the shore, the shore opens at night so the moon is
// up for its line, the crab asks in the wide and walks up, the finish comes
// at dawn. Same beats as heroTurn, different offsets; the site's hero is
// untouched.
func trailerMontage() montage {
	m := montage{key: "trailer", label: "the trailer", fps: 30, cols: 124, agentRows: 14, rows: 30}
	const (
		prompt1 = "add rate limiting to the auth endpoint"
		done1   = "Rate limiting is in. 100 req/min per IP."
		prompt2 = "wire it into the router and ship it"
		done2   = "Shipped. The router limits every route to 100 req/min."
	)
	turn := func(at0 float64, tag, prompt, done string, ask, first bool, ctx0 float64) []loopBeat {
		a := func(d float64) float64 { return at0 + d }
		id := func(s string) string { return s + tag }
		var promptPrint []string
		if !first {
			promptPrint = []string{"", "> " + prompt}
		}
		b := []loopBeat{
			{at: a(0), evs: []event.Event{{Kind: event.Prompt, Text: prompt}, todo(0, 5), ctx(ctx0)}, print: promptPrint},
			{at: a(3), evs: []event.Event{
				tool(event.ToolStart, id("t1"), event.OpRead, "Read", "internal/auth/handler.go", ""),
				tool(event.ToolEnd, id("t1"), event.OpRead, "Read", "internal/auth/handler.go", "142 lines"),
			}, print: []string{"*Read\tinternal/auth/handler.go\t142 lines"}},
			{at: a(8), evs: []event.Event{
				tool(event.ToolStart, id("t2"), event.OpSearch, "Grep", "rate.Limiter", ""),
				tool(event.ToolEnd, id("t2"), event.OpSearch, "Grep", "rate.Limiter", "3 files"),
				todo(1, 5),
			}, print: []string{"*Grep\trate.Limiter\t3 files"}},
			{at: a(18), evs: []event.Event{
				tool(event.ToolStart, id("t3"), event.OpEdit, "Edit", "internal/auth/handler.go", ""),
				tool(event.ToolEnd, id("t3"), event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
				todo(2, 5), ctx(ctx0 + 0.10),
			}, print: []string{"*Edit\tinternal/auth/handler.go\t+18 -2"}},
			{at: a(26), evs: []event.Event{
				tool(event.ToolStart, id("t4"), event.OpWrite, "Write", "internal/auth/limiter.go", ""),
				tool(event.ToolEnd, id("t4"), event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines"),
				todo(3, 5), ctx(ctx0 + 0.16),
			}, print: []string{"*Write\tinternal/auth/limiter.go\t64 lines"}},
			// the fan-out lands AFTER the edit and the write, so the wide
			// shot has the work rising before the owlets come (a(40): his
			// "hold the wide shots a bit longer", 2026-09-18 ~03:00)
			{at: a(40), evs: []event.Event{
				sub(id("a1"), "Explore", event.SubStart),
				sub(id("a2"), "Explore", event.SubStart),
				sub(id("a3"), "general-purpose", event.SubStart),
			}, print: []string{"*Task\tExplore x2, general-purpose\t3 agents"}},
		}
		test := a(56)
		if ask {
			// the question holds 40 session seconds (5 of clip): the walk up
			// in the wide, then the closeup with the eyes cycling
			b = append(b,
				loopBeat{at: a(30), evs: []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}}, print: []string{"", "!allow Bash?"}},
				loopBeat{at: a(70), evs: []event.Event{{Kind: event.Prompt, Text: "yes"}}, print: []string{"> yes"}},
			)
			test = a(72)
		}
		b = append(b,
			loopBeat{at: test, evs: []event.Event{
				tool(event.ToolStart, id("t5"), event.OpShell, "Bash", "go test ./internal/auth", ""),
				tool(event.ToolEnd, id("t5"), event.OpShell, "Bash", "go test ./internal/auth", "ok 1.4s"),
				todo(4, 5), ctx(ctx0 + 0.28),
			}, print: []string{"*Bash\tgo test ./internal/auth\tok 1.4s"}},
			// the subagents end past their (shortened) dwell, so they leave
			// before the finish
			loopBeat{at: test + 2, evs: []event.Event{sub(id("a1"), "Explore", event.SubEnd), sub(id("a2"), "Explore", event.SubEnd)}},
			loopBeat{at: test + 6, evs: []event.Event{sub(id("a3"), "general-purpose", event.SubEnd)}},
			loopBeat{at: test + 16, evs: []event.Event{todo(5, 5), ctx(ctx0 + 0.34), {Kind: event.Done, Text: done}},
				print: []string{"", "!" + done}},
		)
		return b
	}
	// turn a on the vista from 08:38 after the prompt is TYPED (trailerType,
	// his "like the beginning of the Matrix"), the night set in by the cut
	// at 120; turn b on the shore at night, the finish at dawn.
	//
	// THE SAND STEPS. The beach's tone is cube-exact by design (s27), so as
	// the palette runs from midnight to dawn the sand jumps from the night
	// ochre (cube 94) to the dawn tan (137) in ONE frame, measured on the
	// exported frames between hours 1.1284 and 1.1295. A real night takes
	// hours to cross that; the trailer's takes twelve seconds, and in v3 the
	// jump landed inside the crab's closeup (his note). The shore's clock
	// reaches 1.129 exactly on the cut to the wide at 194, so the jump is
	// the cut's.
	b := turn(16, "a", prompt1, done1, false, true, 0.08)
	b = append(b, turn(125, "b", prompt2, done2, true, false, 0.44)...)
	b = append(b, loopBeat{at: 219, evs: []event.Event{ev(event.Compact), ctx(0.06), todo(0, 5)}, clear: true})
	m.beats = b
	m.secs, m.speed = 27.6, 8
	m.stops = []todStop{{0, 0.36}, {120, 0.93}, {194, 1.129}, {219, 1.25}}
	m.segs = []heroSeg{{0, ScapeVista, "owl", nil}, {120, ScapeShore, companion.NameCrab, nil}}
	return m
}

// trailerType is the first prompt typed into the pane character by character
// from black (his note, 2026-09-18 ~11:30: "fade to black and THEN the
// terminal should start typing, like the beginning of the Matrix"): a bare
// cursor blinking, then the text at typeCPS, then the line as the pane opens
// on it. The first beat lands after the typing is done.
const (
	trailerTypeFrom = 0.5  // clip seconds: the cursor blinks alone until then
	trailerTypeCPS  = 30.0 // characters a second
)

func trailerTyped(lines []string, prompt string, t float64) []string {
	done := trailerTypeFrom + float64(len([]rune(prompt)))/trailerTypeCPS
	if t >= done {
		return lines
	}
	out := append([]string{}, lines...)
	var line string
	switch {
	case t < trailerTypeFrom:
		cur := "█"
		if int(t/0.5)%2 == 1 {
			cur = " "
		}
		line = "> " + cur
	default:
		n := int((t - trailerTypeFrom) * trailerTypeCPS)
		r := []rune(prompt)
		if n > len(r) {
			n = len(r)
		}
		line = "> " + string(r[:n]) + "█"
	}
	if len(out) > 0 && strings.HasPrefix(out[0], "> ") {
		out[0] = line
	} else {
		out = append([]string{line}, out...)
	}
	return out
}

// trailerCat is the cat for the end card (his ask, 2026-09-18 ~03:00: "the
// cat companion appear at the very end underneath the logo/website and
// emote"): the cast clip's portrait (siteframes.go castClips), alone on the
// transparent ground, working for a moment and then its finished face, one
// crop for both so the picture does not jump at the change.
func trailerCat() (frames []string, workFrames, cols, rows int, css string, err error) {
	pal := &canvas.HTMLPalette{Transparent: &portraitGround}
	crop := [4]int{0, 0, 18, 9}
	mk := func(pose companion.State, secs float64) fxClip {
		return fxClip{key: "trailer-cat",
			port: &portrait{who: companion.NameCat, pose: pose, cols: 18, rows: 9, x: 3, y: 1, crop: &crop},
			sc:   gifScene{name: "trailer-cat", tod: 0.62, secs: secs, fps: 30, pal: pal, cube: !truecolorFX}}
	}
	work, _, _, err := portraitFrames(mk(companion.Working, 1.6))
	if err != nil {
		return nil, 0, 0, 0, "", err
	}
	done, cols, rows, err := portraitFrames(mk(companion.Done, 3.6))
	if err != nil {
		return nil, 0, 0, 0, "", err
	}
	return append(work, done...), len(work), cols, rows, pal.CSS(), nil
}

// trailerFrames is montageFrames with the trailer's two liberties: the
// owlets' dwell is halved (reduce.KittenDwell, 60 s in the product) so a
// fan-out can come and go inside one shot, and the near crab's ask eye
// cycles through its four rotating variants (his "animate the eyes with
// another of the variants for fun"), through the study override that
// nothing in the product sets. Both restored on return.
func trailerFrames(seed int64, m montage, pal *canvas.HTMLPalette) ([]string, error) {
	wasDwell := reduce.KittenDwell
	reduce.KittenDwell = 30 * time.Second
	defer func() { reduce.KittenDwell = wasDwell; companion.StudyNearEyeAlert = nil }()

	sc := gifScene{name: m.key, secs: m.secs, speed: m.speed, fps: m.fps, cols: m.cols,
		agentRows: m.agentRows, rows: m.rows, beats: m.beats, pal: pal, cube: !truecolorFX}
	run := newSessionRun(sc)
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	c := canvas.New(m.cols, m.rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cats := map[string]*companion.Cat{}
	catOf := func(who string) *companion.Cat {
		if who == "owl" {
			return nil
		}
		if cats[who] == nil {
			k := companion.New(who)
			k.FaceLeft(true)
			cats[who] = k
		}
		return cats[who]
	}
	ccw, chh := companion.New(companion.DefaultName).Size()
	lay := compose(c.W, ccw, true)
	sh := scape.NewShore(seed, false)
	sh.MoonX = lay.MoonX
	v := scenes.NewVista(seed, false)
	v.OwlX = lay.CatX
	var prompt string
	for _, b := range m.beats {
		for _, e := range b.evs {
			if e.Kind == event.Prompt && prompt == "" {
				prompt = e.Text
			}
		}
	}
	n := int(m.secs * float64(m.fps))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		st, t, now := run.at(i, n)
		session := t * m.speed
		st.Act.TimeOfDay = m.todAt(session)
		seg := m.segAt(session)
		if seg.pose != nil {
			st.Pose = *seg.pose
		}
		// the eyes: one variant every 0.55 s of clip time while the ask holds
		companion.StudyNearEyeAlert = companion.NearEyeRotationArt(int(t/0.55) % 4)
		cat := catOf(seg.who)
		if seg.scape == ScapeVista {
			v.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, c.W-4)
			drawVistaWith(c, v, lay, st, t, cat)
		} else {
			if cat == nil {
				cat = catOf(companion.DefaultName)
			}
			sh.Update(c, t, st.Act)
			st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
			cat.Approach(1/float64(m.fps), st.Pose == companion.NeedsYou)
			drawScene(c, sh, cat, lay, st, t, seed, c.H-2-chh)
		}
		frame := c.HTMLFragmentClassed(GIFPx, prof, pal)
		pane := agentPane(c.W, m.agentRows, trailerTyped(run.lines, prompt, t))
		out = append(out, pane.HTMLFragmentClassed(GIFPx, prof, pal)+frame)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no frames", m.key)
	}
	return out, nil
}

// trailerCover is coverLayersAt's sea (site.go), the same seed, hour and six
// steps, exported CELL BY CELL: every glyph the splash shows with its own
// ink, [x, y, glyph, "rrggbb"], so the stage can reveal and light each one
// on its own (his note, 2026-09-17: the shine "around each ascii character
// ... as they wipe in", the stars "one at a time"). horizon is the first
// row that is sea rather than sky: the sky's specks are the same in every
// frame and arrive as stars; everything from the horizon down arrives with
// the wipe.
func trailerCover(seed int64, w, h int) (frames [][][]any, horizon int) {
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	act := scape.Activity{Working: true, Level: 0.55, TimeOfDay: 0.93}
	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for i := 0; i < 30; i++ {
		sh.Update(c, float64(i)/20, act)
	}
	perRow := make([]int, h)
	for i := 0; i < 6; i++ {
		sh.Update(c, 2+float64(i)*0.7, act)
		var cells [][]any
		for y, line := range strings.Split(c.RenderPlain(), "\n") {
			n := 0
			for x, r := range []rune(line) {
				if r == ' ' {
					continue
				}
				n++
				_, fg, _ := c.ResolveAt(x, y, term.ProfileTrueColor)
				cells = append(cells, []any{x, y, string(r), fmt.Sprintf("%02x%02x%02x", fg.R, fg.G, fg.B)})
			}
			perRow[y] = max(perRow[y], n)
		}
		frames = append(frames, cells)
	}
	// The sky holds a handful of specks a row; the sea's first row holds
	// dozens in at least one frame.
	horizon = h
	for y, n := range perRow {
		if n >= 20 {
			horizon = y
			break
		}
	}
	return frames, horizon
}
