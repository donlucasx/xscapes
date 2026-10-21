# The QT card (2026-09-19, session 43)

Art for a quote post of the trailer (https://x.com/donlucas/status/2101029250996617468),
the day after it went up: the site's "Scapes and companions" section as a 16:9
loop. His brief: "16:9, header with ascii separator on the top, underneath 4
companions, animated, on a loop, maybe a gif?" Nothing here ships in the
product; `qt_test.go` (gated, `XSCAPES_QT=<dir>`) is the only seam, and it
only hands this directory the same markup the site plays.

## What it is

The section's own header rule (`┌─ SCAPES AND COMPANIONS ────…┐`, drawn the
way the site's `drawBox` draws a `.sec`: JetBrains Mono, the title bold and
uppercase, the rule filling the width in whole characters), the four cast
portraits under it (the crab, the cat, the owl, the frog: `castClips`'
working pose, 18x9 on the transparent ground, each cropped to its own ink
box, every one at the same cell size), and the site's captions under each.
Dark theme only (the site's `--g` #121212). 1920x1080, 30 fps, 13.2 s, a
seamless loop.

## The loop's length, measured

The animals run on clocks that never close together exactly: the crab and
the cat breathe every 2.2 s but the cat's wag is `sin(t*2.4)` (2.618 s) and
both blink on `mod(t, 5.3)`; the owl's picked working motion ("looks, blink,
flutter") acts for 3.2 s of every 10 and blinks on `mod(t, 7)`; the frog is a
still that blinks. `TestQTPeriods` renders each animal with no blink and finds
no exact period inside 30 s for three of them. So the seam is measured
instead: the frame that would follow the last against the first, as cells
that differ, at candidate lengths:

| loop | crab | cat | owl | frog |
|---|---|---|---|---|
| 4.4 s | 0 | 16 | 0 | 0 |
| 6.6 s | 0 | 17 | 8 | 0 |
| 8.8 s | 0 | 18 | 66 | 0 |
| 11.0 s | 0 | 16 | 0 | 0 |
| **13.2 s** | **0** | **0** | **0** | **0** |
| 15.4 s | 2 | 19 | 0 | 0 |

13.2 s is six breaths and 5.04 wags, close enough that the cat's tail lands on
the same cells. Two offsets make it hold: the owl's clock starts 4 s in
(`qtOwlOffset`), so its quiet stretch spans the seam and its action plays at
6–9.2 s; the crab's and the cat's start 0.5 s in (`qtCatOffset`), because at
t = 0 both are mid-blink AND the breath sits on its zero crossing, where
`sin(2πt/2.2) > 0` is a floating-point coin toss at every multiple of 2.2 s
(the first measurement seamed 2 cells on the crab at every length: the eyes).
The frog blinks twice a loop, held ~100 ms (one frame at the site's 6 fps is
167 ms; at 30 fps it is 33 ms and invisible).

## The pipeline

```
# 1. frames (from the repo root)
XSCAPES_QT=$SP/qt/frames XSCAPES_QT_SECS=crab:13.2,cat:13.2,owl:13.2,frog:13.2 go test -run TestQTExport .
# 2. the stage, screenshotted by the installed Chrome through Playwright
cd $SP/qt && node render-card.mjs frames out-card --stills 0,6.6    # check first
node render-card.mjs frames out-card                                # 396 frames
# 3. the cut
sh assemble.sh out-card none xscapes-companions-16x9.mp4 30
ffmpeg -framerate 30 -i out-card/f%05d.png -vf "fps=15,scale=1280:-1:flags=lanczos,split[a][b];[a]palettegen=max_colors=128:stats_mode=full[p];[b][p]paletteuse=dither=none:diff_mode=rectangle" -loop 0 card.gif
gifsicle -O3 card.gif -o xscapes-companions-1280.gif
```

`card.json` holds every number he tunes: the title, the paddings, the rule's
size, the cell size, the captions. The workbench is the session scratchpad
(`qt/` there has `node_modules` linked from the trailer's, `fonts/`, `frames/`,
`out-card/`); this directory is the record. Fonts: the trailer's Geist Mono
and JetBrains Mono 700 plus JetBrains Mono 400/500 fetched from Google Fonts
for the rule and the captions (the site's furniture weights).

## The post (2026-09-19 11:20 PDT)

https://x.com/donlucas/status/2101375821940392422 : the MP4 quoting the trailer, posted and verified
by the parallel session. His rulings there: no Unicode "italics" from a text
formatter (the name must stay searchable and copyable), and no extra link in
the body, since the quoted trailer's second post carries it.

## Deliverables

`~/Desktop/xscapes-qt/`: the MP4 (post this one), the GIF (a fallback), the
first frame as a still, `post.md` with the suggested text.

## The fire clip (2026-09-21, session 44)

His ask: "a cool video/animation of the fireplace as the context window ... about
watching our agents work while we wait for the winners to be announced", then
"we should be able to see the fire in context (so should zoom out or zoom in at
some point), also should see the fire activity pickup from nothing to WORKING
within the scene- maybe even a timelapse behind as the day gets darker".

THE ENCODING STANDS: the fire is the work (with the wind) and the moon is the
context; the clip shows both as they are and the caption says which is which.

`fire_test.go` (gated, `XSCAPES_FIRE=<dir>`, `XSCAPES_FIRE_SIZE=100x34`) writes
`fire.js`: the vista alone (no agent pane: `agentPane` cannot be zero rows, so
`fireFrames` is `montageFrames` without it), 100x34 (Menlo's advance is 0.602 em,
so rows x 2.95 columns fill 16:9 at 1080), 30 fps, 28 s at speed 8. The session (v2,
his "should ramp up faster"): a second and a half of nothing, then the hero's two
turns at speed 10 (context 0.30 → 0.64 → 0.98) and the compact at 20.5 s; 21.5 s
in all (v1 was the prompt at 4 s, speed 8, 28 s). The hour: 0.70 (dusk) → 1.05
(night). `render-fire.mjs` + `fire.json`: the trailer's camera alone; a 3x
closeup on the fire (col 38, row 25) to 3.8 s, the pull to the wide by 6.3 s,
wide to the end (an end push-in cut the owl and its finish knock; dropped).

```
XSCAPES_FIRE=$SP/qt/fire100 XSCAPES_FIRE_SIZE=100x34 go test -run TestFireExport .
cd $SP/qt && node render-fire.mjs fire100 out-fire --stills 1,4.5,6.5,9,13,20,27
node render-fire.mjs fire100 out-fire                      # 645 frames, ~26 s
sh assemble.sh out-fire none xscapes-by-the-fire-16x9.mp4 30
```

Deliverables in `~/Desktop/xscapes-qt/`: the MP4 (v2 4.5 MB, v1 5 MB), a 1280 GIF, the
opening and a night still, captions in `post.md`.

POSTED 2026-09-21 ~14:30 PDT, quoting the Commons "almost down to the top 25"
post: https://x.com/donlucas/status/2102142846644142500


## The shore clip (2026-09-21, session 44)

For his reply on X: "The tides shows you how hard your agents are working on the
shore scape". `shore_test.go` (gated `XSCAPES_SHORE=<dir>`): the site's own shore
session (sceneClips "shore") with two seconds of flat water before the prompt and
three after the finish, 80x27, tod 0.62 held, 14 s at 30 fps, speed 4; its own
frame loop (`shoreFrames`) because `agentPane` cannot be zero rows.
`render-fire.mjs --name shore` with the static camera in `shore.json`. v2 (his
"starts with low tide, ramps up quickly, focus on the tide"): the tide state
starts at zero withdrawal = HIGH tide and eases over TideEase 3 s, so the clip
warms the shore up with fifteen quiet seconds before frame 0 and sets TideEase
to 1 s for the render (both restored); speed 6, the prompt at 1.5 s, 11 s; no
zoom (a 1.2 zoom on the waterline was tried and dropped: the writing in the
sand and the crab cannot both fit above 1.03).

```
XSCAPES_SHORE=$SP/qt/shore80 go test -run TestShoreExport .
cd $SP/qt && node render-fire.mjs shore80 out-shore --name shore --stills 1,4,7,10,13.5
node render-fire.mjs shore80 out-shore --name shore          # 420 frames
sh assemble.sh out-shore none xscapes-shore-16x9.mp4 30
```

## Whole-pixel cells (2026-09-21, after his "look at all the lines")

The stage sizes the font from the frame's height, so Menlo's advance is a
fraction of a pixel (24.0875 px at 40 px); the cell edges drift across the row
and Chrome blends the edge pixel wherever two backgrounds meet: 1-px seams, 3-7%
off, at cell boundaries (measured on the source frames). `render-fire.mjs`
now searches the font size whose row is exactly cols x round(advance) wide
and holds the line height at the row's height; the row measures within 1/64 px
and the seam residual is one unit. GIFs: no dither, the 960 one by an exact 2x
area downscale. The cast card (`render-card.mjs`) and the fire clip predate the
fix.
