package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
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
		sc: gifScene{
			name: "hero-live", tod: 0.30, todEnd: 1.30, secs: 15, cols: 96,
			agentRows: 12, speed: loopSecs / 15.0, beats: windowLoop(),
			fps: 6, pal: pal, cube: !truecolorFX,
		},
	}
}

// stateClips are the legend, and they are deliberately ONE scene at five
// moments rather than five separate clips. An encoding is learned by
// comparison: the same beach, the same companion, the same width, and only
// the thing being named different. Nine unrelated clips could not do that,
// which is why the section they replace needed 392 words of prose.
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
		mk("needs", "it needs you", "A solid balloon and a chime. The companion comes closer and puts a claw up, and it holds the pose until you come back.", 30, 0.80, 4),
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

// ruleClip is the page's section divider, and it is a real strip of the
// product rather than a row of punctuation.
//
// His note of 2026-09-14: "I dont see any use of ascii that stands out
// (except the splash page, which I love)" -- the dotted rules and the prompt
// glyphs were decoration pretending to be terminal. The splash works because
// it is the actual sea. So every section rule is three rows of the actual
// sea, moving, full width.
func ruleClip(pal *canvas.HTMLPalette) fxClip {
	return fxClip{
		key: "rule", label: "rule",
		// Rows 9..12 are OPEN SEA. Taking the waterline instead put sand,
		// crablets and a bit of the writing band into a decorative strip,
		// which read as a slice of something rather than as water.
		sc: gifScene{name: "rule", at: 14, tod: 0.52, secs: 4, fps: 6, cols: 120,
			crop: [4]int{0, 9, 120, 12}, pal: pal, cube: !truecolorFX},
	}
}

// renderFX renders every embedded animation and returns the page's script
// payload and the stylesheet the frames share.
func renderFX(seed int64) (js, css string, err error) {
	pal := &canvas.HTMLPalette{}
	clips := append([]fxClip{heroClip(pal)}, stateClips(pal)...)
	clips = append(clips, swapClips(pal)...)
	clips = append(clips, ruleClip(pal))
	out := map[string]fxPayload{}
	for _, cl := range clips {
		frames, err := gifFrames(seed, cl.sc)
		if err != nil {
			return "", "", fmt.Errorf("%s: %w", cl.key, err)
		}
		out[cl.key] = fxPayload{
			Cols: cl.sc.colsOf(), Rows: cl.sc.rowsOf(), FPS: cl.sc.rate(),
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
