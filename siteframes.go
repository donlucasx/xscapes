package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE PAGE'S ANIMATIONS ARE TEXT, NOT PICTURES.
//
// The clips shipped as GIFs until now, and that was a downgrade of the
// engine's own output: gifPages already writes real <pre> frames with per-cell
// colour, and site/make-gifs.py screenshots them through headless Chrome to
// turn that text back into a raster. 5.1 MB of the published page is
// photographs of markup we already had.
//
// So the page embeds the frames themselves. They are crisp at any zoom, they
// inherit the page's own font, they need no Chrome and no gifsicle, and the
// colours live once in a stylesheet instead of once per run -- which is what
// makes it small enough to ship at all.
//
// The cover already proved the mechanism: six stacked <pre> frames cycled by a
// CSS step animation. This is the same idea with a frame count no stylesheet
// wants to hold, so a very small player swaps one frame at a time.

// truecolorFX decides whether the page's embedded frames are truecolor or the
// 256 cube.
//
// HIS RULING 2026-09-05, and it is why this is a named constant rather than a
// quiet choice: "the page shows the scene at its best, not as the cube rounds
// it." The product runs on the cube; the page is what it is reaching for.
//
// It is not free, because an embedded frame pays for its colours in a
// stylesheet and truecolor gradients give nearly every run its own pair.
// Measured both ways, all eight clips: cube 37KB of CSS, truecolor 920KB --
// see TestWhatTheEmbeddedFramesCost for what that is after gzip.
const truecolorFX = true

// fxClip is one embedded animation: how to render it, and what the page calls it.
type fxClip struct {
	key   string
	label string
	note  string
	sc    gifScene

	// scene, when set, is a study painter (internal/scenes) rendered as a
	// seamless loop instead of a shore session folded through the reducer:
	// frames frames at fps, the work rising and settling once across them.
	scene  *scenes.Scene
	tod    float64
	frames int
	fps    int
	// with is the study companion drawn in the scene, or nil for the cat.
	with *scenes.Animal

	// port, when set, is the companion ALONE on a flat ground: no sea, no
	// sand, no litter. See portraitFrames.
	port *portrait
}

// A PORTRAIT IS THE COMPANION BY ITSELF.
//
// His ask of 2026-09-15: "the companion by itself (instead of within the
// scene), so it reads cleaner and more focused on what we are showing. Every
// 'state' should be toggable like right now, but should only showcase the
// companion, and its states. Same for 'a layer, not a screen', where it should
// feature all companions we have drafted thus far (crab, cat, owl, frog)."
//
// One animal on one flat colour, in a state, breathing, with its balloon when
// the state raises one. The legend's five are still folded from the session
// (sc.beats / sc.at), so the pose is the reducer's and not a posed one; the
// cast holds a pose, because its subject is who they are and not what they
// are doing.
type portrait struct {
	// who is companion.NameCrab or NameCat, drawn by the companion package
	// itself, or "owl" / "frog", the two candidates drawn in internal/scenes.
	who string
	// pose is held for every frame when the clip carries no session.
	pose companion.State
	// ground is the flat colour under the animal; zero means portraitGround.
	ground     term.RGB
	cols, rows int
	// x, y is the companion's box origin. The near poses grow DOWN from y,
	// exactly as they do on the beach, so rows below the box are left for the
	// ask to grow into. x is the box's column at its shipped size; a near
	// rung is re-centred on the frame (see portraitFrames), because a
	// portrait has no scene for it to walk into.
	x, y int
}

// portraitGround is the colour under every portrait, and the page writes it
// as TRANSPARENT: the animal sits on the page's own ground. "No background
// whatsoever", his note of 2026-09-15 on the first build, which had put the
// five states on a flat dusk sand. A colour no scene paints, so nothing else
// on the page can fall through it.
var portraitGround = term.RGB{R: 1, G: 0, B: 1}

// sceneClips are the other scapes, animated. Until 2026-09-15 they were
// checked-in stills, because the painters lived in a main package the site
// could not import.
func sceneClips(pal *canvas.HTMLPalette) []fxClip {
	return []fxClip{
		{key: "rain", label: "the rainy window", note: "Rain on the glass is the work.",
			scene: scenes.Find("Rainy window"), tod: 0.75, frames: scenes.RainPeriod, fps: 6, sc: gifScene{pal: pal}},
		// His pick of 2026-09-15: the wind is the work, the fire is the light.
		// Dusk, because that is when the peaks take the light.
		{key: "vista", label: "the mountain vista", note: "The wind is the work.",
			scene: &scenes.Forest[0], tod: 0.78, frames: int(scenes.LoopSecs * 6), fps: 6, sc: gifScene{pal: pal}},
		// The last still on the page, animated 2026-09-15 at his word. Night,
		// so the lamp is lit; the frog, which is the companion drawn for it.
		{key: "aquarium", label: "the aquarium", note: "More fish come out as the work picks up.",
			scene: scenes.Find("Aquarium"), tod: 0.02, frames: int(scenes.AquariumLoop * 6), fps: 6,
			with: scenes.FindAnimal("Frog"), sc: gifScene{pal: pal}},
	}
}

// allFX is every embedded animation, in one place so the page and the cost
// test cannot disagree about what ships.
func allFX(pal *canvas.HTMLPalette) []fxClip {
	clips := append([]fxClip{heroClip(pal)}, stateClips(pal)...)
	clips = append(clips, castClips(pal)...)
	return append(clips, sceneClips(pal)...)
}

// sceneFrames renders a study painter as a loop that closes on itself: the
// painter's own time runs frame by frame, and the work follows one cosine
// from 0.1 up to 0.9 and back, so the motion slot is seen at every level.
func sceneFrames(cl fxClip) ([]string, error) {
	if cl.scene == nil {
		return nil, fmt.Errorf("%s: no scene", cl.key)
	}
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	scenes.SetCompanion(cl.with)
	defer scenes.SetCompanion(nil)
	out := make([]string, 0, cl.frames)
	for k := 0; k < cl.frames; k++ {
		t := float64(k) / float64(cl.fps)
		level := 0.5 - 0.4*math.Cos(2*math.Pi*float64(k)/float64(cl.frames))
		c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Clear()
		cl.scene.Paint(c, cl.tod, t, level, 7)
		out = append(out, c.HTMLFragmentClassed(GIFPx, prof, cl.sc.pal))
	}
	return out, nil
}

// portraitFrames renders a portrait: the companion alone on a transparent
// ground, CROPPED to the animal's own bounding box.
//
// His note of 2026-09-15 on the first floating build: "the first one is 'it
// needs you' and it stands out because it comes closer. Should the bounding
// box/crop be the same as the session? so the character reads same as it
// will? and also so the rest of the 'states' have less negative space around
// them." So the frame is the ink's own box, one cell of air around it, the
// union over every frame of the clip so the picture does not jump as it
// breathes -- and every state plays at the same cell size, so the ask is
// twice the resting crab the way it is on the beach.
func portraitFrames(cl fxClip) (frames []string, cols, rows int, err error) {
	p := cl.port
	prof := term.ProfileTrueColor
	if !truecolorFX {
		prof = term.Profile256
	}
	n := int(cl.sc.secs * float64(cl.sc.rate()))
	if n <= 0 {
		return nil, 0, 0, fmt.Errorf("%s: no frames", cl.key)
	}
	// A session, when the clip carries one, decides the pose and the balloon.
	var run *sessionRun
	if cl.sc.beats != nil || cl.sc.at > 0 {
		run = newSessionRun(cl.sc)
	}
	var cat *companion.Cat
	switch p.who {
	case companion.NameCrab, companion.NameCat:
		cat = companion.New(p.who)
		cat.FaceLeft(true)
	}
	ground := p.ground
	if ground == (term.RGB{}) {
		ground = portraitGround
	}
	// Every frame is kept as a canvas until the crop is known.
	cs := make([]*canvas.Canvas, 0, n)
	x0, y0, x1, y1 := p.cols, p.rows, 0, 0
	for i := 0; i < n; i++ {
		pose, bubble, ask := p.pose, "", false
		t := float64(i) / float64(cl.sc.rate())
		if run != nil {
			st, tt, _ := run.at(i, n)
			pose, bubble, ask, t = st.Pose, st.Bubble, st.BubbleAsk, tt
		}
		c := canvas.New(p.cols, p.rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Clear()
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, ground)
			}
		}
		// One blink a loop, at a fixed frame, so the loop closes on itself.
		blink := i == 3
		switch p.who {
		case "owl":
			scenes.DrawOwlPicked(c, p.x, p.y, blink)
		case "frog":
			scenes.DrawCandidate(c, scenes.FindAnimal("Frog"), p.x, p.y, true, blink)
		default:
			// The walk up before the ask, the same call the beach makes.
			cat.Approach(1/float64(cl.sc.rate()), pose == companion.NeedsYou)
			// CENTRED AT EVERY RUNG. On the beach a near pose hangs left of
			// the box so the right margin holds (nearDX); here there is no
			// margin and no scene, so the drawn box is centred on the frame
			// and the approach reads as the animal coming closer.
			dx, w, _ := cat.DrawnBox(c.W)
			x := (c.W-w)/2 - dx
			cat.Draw(c.Near(), x, p.y, t, pose)
			if bubble != "" {
				rows, col := companion.DoneBubble(bubble), bubbleCol
				if ask {
					rows, col = companion.Bubble(bubble), bubbleAskCol
				}
				rows, bx := portraitBubble(rows, x+cat.DrawnHeadCol(c.W), c.W)
				(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).Draw(c.Near(), bx, p.y-len(rows))
			}
		}
		cs = append(cs, c)
		ax0, ay0, ax1, ay1 := inkBox(c)
		x0, y0, x1, y1 = min(x0, ax0), min(y0, ay0), max(x1, ax1), max(y1, ay1)
	}
	if x1 <= x0 || y1 <= y0 {
		return nil, 0, 0, fmt.Errorf("%s: nothing drawn", cl.key)
	}
	x0, y0, x1, y1 = max(x0-1, 0), max(y0-1, 0), min(x1+1, p.cols), min(y1+1, p.rows)
	frames = make([]string, 0, n)
	for _, c := range cs {
		frames = append(frames, c.HTMLFragmentCropClassed(x0, y0, x1, y1, GIFPx, prof, cl.sc.pal))
	}
	return frames, x1 - x0, y1 - y0, nil
}

// inkBox is the smallest cell rectangle holding every glyph on the canvas,
// as x0, y0 inclusive and x1, y1 exclusive. A cleared rim around a sprite is
// spaces, so it falls outside and the margin added by the caller is honest.
func inkBox(c *canvas.Canvas) (x0, y0, x1, y1 int) {
	x0, y0, x1, y1 = c.W, c.H, 0, 0
	for y, row := range strings.Split(c.RenderPlain(), "\n") {
		for x, r := range []rune(row) {
			if r == ' ' {
				continue
			}
			x0, y0 = min(x0, x), min(y0, y)
			x1, y1 = max(x1, x+1), max(y1, y+1)
		}
	}
	return
}

// portraitBubble lays a balloon over a centred companion: the box is centred
// on the head as far as the frame allows, and the POINTER is moved to sit
// over the head wherever the box ends up.
//
// On the beach the pointer is fixed at a shoulder and bubbleX slides the box
// so that shoulder is over the head, clamping at the frame edge -- which is
// right for a scene, where the companion stands at one side. A portrait has
// the animal in the middle of a 48-cell frame, and the done knock is 44 cells
// wide: clamped, its shoulder pointer landed 16 cells off the head.
func portraitBubble(rows []string, head, w int) ([]string, int) {
	bw := bubbleWidth(rows)
	x := head - bw/2
	if x+bw > w {
		x = w - bw
	}
	if x < 0 {
		x = 0
	}
	out := append([]string(nil), rows...)
	last := []rune(out[len(out)-1])
	// Erase the shoulder pointer with the row's own stroke, then place it.
	filler := last[len(last)-2]
	for i, r := range last {
		if r == 'v' {
			last[i] = filler
		}
	}
	at := head - x
	if at < 1 {
		at = 1
	}
	if at > len(last)-2 {
		at = len(last) - 2
	}
	last[at] = 'v'
	out[len(out)-1] = string(last)
	return out, x
}

// framesOf renders one clip whichever way it is declared.
func framesOf(seed int64, cl fxClip) (frames []string, cols, rows, fps int, err error) {
	if cl.scene != nil {
		frames, err = sceneFrames(cl)
		return frames, 80, 24, cl.fps, err
	}
	if cl.port != nil {
		frames, cols, rows, err = portraitFrames(cl)
		return frames, cols, rows, cl.sc.rate(), err
	}
	frames, err = gifFrames(seed, cl.sc)
	return frames, cl.sc.colsOf(), cl.sc.rowsOf(), cl.sc.rate(), err
}

// fxPayload is what the page's player receives.
type fxPayload struct {
	Cols   int      `json:"cols"`
	Rows   int      `json:"rows"`
	FPS    int      `json:"fps"`
	Label  string   `json:"label"`
	Note   string   `json:"note"`
	Frames []string `json:"frames"`
}

// heroClip is the whole session in one loop, the day turning under it, with
// the agent's own transcript above its scape in the same window.
//
// Frame rate and width are the two dials that decide whether this ships. Every
// frame is markup, so 10fps at 108 columns is four times the page a judge will
// wait for; the sea and the companion still read at 6.
func heroClip(pal *canvas.HTMLPalette) fxClip {
	return fxClip{
		key:   "hero",
		label: "a whole session",
		note:  "one turn, start to finish, with the day turning under it",
		// A REAL WINDOW, not a thumbnail of one. His own is 125x62; this
		// renders at 124x44 -- 14 rows of transcript over a 30-row scape --
		// because the page had been showing a cramped version of a product
		// whose whole argument is that it has room to breathe. The detail
		// clips below stay tight on purpose; this one is the wide shot.
		sc: gifScene{
			name: "hero-live", tod: 0.30, todEnd: 1.30, secs: 15, cols: 124,
			agentRows: 14, rows: 30, speed: loopSecs / 15.0, beats: windowLoop(),
			fps: 6, pal: pal, cube: !truecolorFX,
		},
	}
}

// stateClips are the legend, and they are deliberately ONE scene at five
// moments rather than five separate clips. An encoding is learned by
// comparison: the same beach, the same companion, the same width, and only
// the thing being named different. Nine unrelated clips could not do that,
// which is why the section they replace needed 392 words of prose.
// askBeats is a session that arrives at a question WITH NOTHING BROKEN, and it
// exists because of a defect I introduced and he had to report twice.
//
// The four other states read their moment out of the shared demo turn, which
// is right: one session, five moments, and only the named thing different.
// But that turn has a command exit 1 at beat 22, and Worried OUTRANKS
// NeedsYou in the reducer -- it is supposed to, because a failure you have not
// seen must not be cancelled by the next question. So the ask clip was drawing
// a WORRIED crab with a question balloon over it, and since the pose was never
// NeedsYou the companion never walked up either. The rung was never the
// problem; the pose was.
//
// The clip the old page shipped had its own beats for exactly this reason.
// This restores that, at full scene size: work lands, subagents arrive, and
// then it asks, with no error outstanding.
func askBeats() []loopBeat {
	return []loopBeat{
		{at: 0, evs: []event.Event{{Kind: event.Prompt, Text: "add rate limiting to the auth endpoint"}, todo(0, 5), ctx(0.46)}},
		{at: 1, evs: []event.Event{
			tool(event.ToolStart, "t1", event.OpSearch, "Grep", "rate.Limiter", ""),
			tool(event.ToolEnd, "t1", event.OpSearch, "Grep", "rate.Limiter", "3 files"),
			tool(event.ToolStart, "t2", event.OpEdit, "Edit", "internal/auth/handler.go", ""),
			tool(event.ToolEnd, "t2", event.OpEdit, "Edit", "internal/auth/handler.go", "+18 -2"),
			tool(event.ToolStart, "t3", event.OpWrite, "Write", "internal/auth/limiter.go", ""),
			tool(event.ToolEnd, "t3", event.OpWrite, "Write", "internal/auth/limiter.go", "64 lines"),
			todo(2, 5),
			sub("a1", "Explore", event.SubStart),
			sub("a2", "Explore", event.SubStart),
			sub("a3", "general-purpose", event.SubStart),
		}},
		{at: 2, evs: []event.Event{{Kind: event.NeedsInput, Text: "allow Bash?"}}},
	}
}

func stateClips(pal *canvas.HTMLPalette) []fxClip {
	// THE COMPANION ALONE, his ask of 2026-09-15. The five used to be the
	// whole beach at 62 columns; the sea, the sand and the moon are named in
	// "While you wait" now, and this section is the companion's own five
	// states. Same frame for all five, so only the animal changes.
	//
	// 48 columns, because the come-closer ladder refuses a rung wider than
	// half the frame and the ask's top rung is 24 cells. 17 rows: three for
	// the balloon above the box's row, and fourteen below it for the biggest
	// rung, which grows DOWN from that row. On the beach the sand's edge
	// crops those legs and he liked the crop; on a transparent ground a
	// cropped leg reads as a cut one, so the whole pose fits.
	port := func() *portrait {
		return &portrait{who: companion.NameCrab, cols: 48, rows: 17, x: 18, y: 3}
	}
	mk := func(key, label, note string, at, tod, secs float64) fxClip {
		return fxClip{key: key, label: label, note: note, port: port(),
			sc: gifScene{name: "st-" + key, at: at, tod: tod, secs: secs, fps: 6, pal: pal, cube: !truecolorFX}}
	}
	return []fxClip{
		{key: "needs", label: "it needs you",
			note: "A solid balloon and a chime, and the companion walks up the beach to ask -- three strides, twice its resting size, so the question is impossible to miss from across the room. It holds the pose until you come back.",
			port: port(),
			sc: gifScene{name: "st-needs", tod: 0.80, secs: 5, fps: 6,
				beats: askBeats(), speed: 2, pal: pal, cube: !truecolorFX}},
		mk("working", "it is working", "Eyes open, breathing quicker, a blink now and then. How hard is the sea's to say; the companion only says that it is at it.", 14, 0.52, 4),
		mk("broke", "something broke", "The companion carries it, never the weather: claws down, stalks short, amber eyes, until the trouble clears.", 22, 0.62, 4),
		// 7 s, because the settled pincer shuts once every 7 (crabClawPeriod)
		// and a shorter loop would never show the one thing that moves.
		mk("done", "it finished", "A dotted balloon and a low note, both claws up, settled on the sand with one pincer shutting every few seconds.", 44, 0.96, 7),
		mk("quiet", "it is waiting", "Eyes closed, breathing slow. Nothing is running and nothing is asked; the next prompt wakes it.", 72, 0.27, 4),
	}
}

// castClips are every companion drawn so far, each by itself, working: the
// two that ship and the two drawn for the scapes that are next. "Feature all
// companions we have drafted thus far (crab, cat, owl, frog)", his ask of
// 2026-09-15. One frame size, one ground, one pose, so the four compare.
//
// 4.4 s is two of the crab's working breaths, so the loop closes; the study
// animals blink once a loop.
func castClips(pal *canvas.HTMLPalette) []fxClip {
	mk := func(key, who, label, note string) fxClip {
		return fxClip{key: "cast-" + key, label: label, note: note,
			port: &portrait{who: who, pose: companion.Working, cols: 18, rows: 9, x: 3, y: 1},
			sc:   gifScene{name: "cast-" + key, tod: 0.62, secs: 4.4, fps: 6, pal: pal, cube: !truecolorFX}}
	}
	return []fxClip{
		mk("crab", companion.NameCrab, "Hero the crab", "Ships as the default."),
		mk("cat", companion.NameCat, "the cat", "Came first; one command away."),
		mk("owl", "owl", "the owl", "Drawn for the mountain vista. Next."),
		mk("frog", "frog", "the frog", "Drawn for the aquarium. Next."),
	}
}

// renderFX renders every embedded animation and returns the page's script
// payload and the stylesheet the frames share.
func renderFX(seed int64) (js, css string, err error) {
	pal := &canvas.HTMLPalette{Transparent: &portraitGround}
	out := map[string]fxPayload{}
	for _, cl := range allFX(pal) {
		frames, cols, rows, fps, err := framesOf(seed, cl)
		if err != nil {
			return "", "", fmt.Errorf("%s: %w", cl.key, err)
		}
		out[cl.key] = fxPayload{
			Cols: cols, Rows: rows, FPS: fps,
			Label: cl.label, Note: cl.note, Frames: frames,
		}
	}
	// The palette must be asked for its CSS only after the last frame, since
	// every frame can add pairs to it.
	b, err := json.Marshal(out)
	if err != nil {
		return "", "", err
	}
	var sb strings.Builder
	sb.WriteString("window.FX=")
	sb.Write(b)
	sb.WriteString(";")
	return sb.String(), pal.CSS(), nil
}
