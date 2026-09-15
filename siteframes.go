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
}

// sceneClips are the other scapes, animated. Until 2026-09-15 they were
// checked-in stills, because the painters lived in a main package the site
// could not import.
func sceneClips(pal *canvas.HTMLPalette) []fxClip {
	return []fxClip{
		{key: "rain", label: "the rainy window", note: "Rain on the glass is the work.",
			scene: scenes.Find("Rainy window"), tod: 0.75, frames: scenes.RainPeriod, fps: 6, sc: gifScene{pal: pal}},
	}
}

// allFX is every embedded animation, in one place so the page and the cost
// test cannot disagree about what ships.
func allFX(pal *canvas.HTMLPalette) []fxClip {
	clips := append([]fxClip{heroClip(pal)}, stateClips(pal)...)
	clips = append(clips, swapClips(pal)...)
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
	scenes.SetCompanion(nil)
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

// framesOf renders one clip whichever way it is declared.
func framesOf(seed int64, cl fxClip) (frames []string, cols, rows, fps int, err error) {
	if cl.scene != nil {
		frames, err = sceneFrames(cl)
		return frames, 80, 24, cl.fps, err
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
	// CROPPED, and all five identically. His note of 2026-09-14 was that the
	// ask needs to be a close-up -- at 80 columns the companion is a thumbnail
	// and the balloon it raises is the whole point of that state. But the
	// crop has to be the SAME for all five, because the section's argument is
	// that only the thing being named changes; a different framing per state
	// would break exactly the comparison it exists to make.
	//
	// NARROWER, NOT CROPPED. His note asked for the ask to be a close-up, and
	// cropping an 80-column scene was the wrong way to get one: it cut the
	// moon in half at x=20, and at x=14 it sliced the writing in the sand
	// mid-word, which reads as a broken render rather than a detail shot.
	//
	// Rendering the scene AT 62 columns instead lets the layout do it properly
	// -- the companion, the moon, the litter and the sand text are all placed
	// for that width, so everything is a third bigger and nothing is cut. The
	// five stay identical to each other, which is the whole point of the
	// section: only the thing being named changes.
	const cols = 62
	mk := func(key, label, note string, at, tod, secs float64) fxClip {
		return fxClip{key: key, label: label, note: note,
			sc: gifScene{name: "st-" + key, at: at, tod: tod, secs: secs, fps: 6, cols: cols, pal: pal, cube: !truecolorFX}}
	}
	return []fxClip{
		{key: "needs", label: "it needs you",
			note: "A solid balloon and a chime, and the companion walks up the beach to ask -- three strides, twice its resting size, so the question is impossible to miss from across the room. It holds the pose until you come back.",
			sc: gifScene{name: "st-needs", tod: 0.80, secs: 5, fps: 6, cols: cols,
				beats: askBeats(), speed: 2, pal: pal, cube: !truecolorFX}},
		mk("working", "it is working", "The sea. How many swells are travelling and how tall, whitecaps once it is flat out. Flat water means it is waiting on you.", 14, 0.52, 4),
		mk("broke", "something broke", "The companion carries it, never the weather: claws down, stalks short, amber eyes, until the trouble clears.", 22, 0.62, 4),
		mk("done", "it finished", "A dotted balloon and a low note, both claws up, the constellation lit and the moon carrying what is left of the context.", 44, 0.96, 4),
		mk("quiet", "it is waiting", "Flat sea, the writing in the sand receding as the tide takes it, the companion still.", 72, 0.27, 4),
	}
}

// swapClips are the same beat with each animal, because "you can change the
// companion" is a claim a page can only make properly by showing both. They
// are cropped to the animal's own corner of the beach, so the pair costs a
// fraction of a full scene.
func swapClips(pal *canvas.HTMLPalette) []fxClip {
	mk := func(key, label, animal, note string) fxClip {
		return fxClip{key: key, label: label, note: note,
			sc: gifScene{name: "sw-" + key, at: 30, tod: 0.62, secs: 3, fps: 6,
				crop: [4]int{46, 8, 80, 24}, animal: animal, pal: pal, cube: !truecolorFX}}
	}
	return []fxClip{
		mk("crab", "crab", companion.NameCrab, "Hero the crab ships as the default."),
		mk("cat", "cat", companion.NameCat, "The cat came first, and is one command away."),
	}
}

// renderFX renders every embedded animation and returns the page's script
// payload and the stylesheet the frames share.
func renderFX(seed int64) (js, css string, err error) {
	pal := &canvas.HTMLPalette{}
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
