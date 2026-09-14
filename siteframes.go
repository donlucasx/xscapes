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
	mk := func(key, label, note string, at, tod, secs float64) fxClip {
		return fxClip{key: key, label: label, note: note,
			sc: gifScene{name: "st-" + key, at: at, tod: tod, secs: secs, fps: 6, pal: pal, cube: !truecolorFX}}
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

// renderFX renders every embedded animation and returns the page's script
// payload and the stylesheet the frames share.
func renderFX(seed int64) (js, css string, err error) {
	pal := &canvas.HTMLPalette{}
	clips := append([]fxClip{heroClip(pal)}, stateClips(pal)...)
	clips = append(clips, swapClips(pal)...)
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
