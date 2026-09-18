# The X trailer (2026-09-17, session 41 thread 2)

A release trailer for X. Nothing here ships in the product; `trailer_test.go`
(gated, `XSCAPES_TRAILER=<dir>`) is the only seam, and it only hands this
directory the same markup the site plays.

## His rulings (verbatim in `_FEEDBACK.md` §Session 41, thread 2)

- **Structure:** one take with a camera over the H1 montage (the site's hero:
  the owl's turn on the vista, the crab's on the shore, one day across both).
  Jump cuts were offered and not taken for now ("leaning to H1").
- **Format:** 16:9 at 1920x1080 first; a square track over the same frames
  after approval. His frame: "the frame can move up and down the terminal
  window, do zoom/crop ins on specific details".
- **Audio:** the product's own cues at their real moments plus a quiet bed,
  no VO. **Captions:** three or four, lowercase, five words or fewer.
- **The brand rules:** lowercase, "always stick to the brand manual unless I
  specify otherwise". The cell blinks alone (allowed); it keeps blinking
  through the slogan because he specified it; it settles solid after, because
  the lockup never blinks. The off state is the hollow cell.
- **The opener (his storyboard, 2026-09-17 ~20:30):** the site's splash sea
  wipes in left to right "as if painted by a giant brush stroke, or shone by
  the sun into existence", in full colour, "shinning on the right edge as it
  appears (lasts not much longer than 1sec)", then the colour drains to the
  site's grey and the cell blinks alone twice above it, in the splash's
  placement; the x appears, "scapes" types beside the held cell, the slogan
  glows in.

## The film (fourth cut, 2026-09-18 ~11:50, `xscapes-trailer-16x9-v4.mp4`, 42.3 s)

The montage is the TRAILER'S OWN (`trailerMontage` in `trailer_test.go`),
not the site's H1: the same two turns, re-timed on his notes. Clip seconds
below are its clock (session seconds / 8); the opener runs 7.6 s before it
and ends in black (v4: the sea's bands 0.35 s apart, the logo settled at 5.0,
a fade to black at 6.6).

| clip s | camera | what happens | sound | caption |
|---|---|---|---|---|
| 0–2.3 | CLOSEUP on the terminal (2.8x, free of the window's edge), from black | a bare cursor blinks, the first prompt TYPES at 30 cps from 0.5 s, the first beat lands at 2.0 | | |
| 2.3–2.7 | fast PULL to WIDE | | | |
| 2.7–7.45 | WIDE | Read, Grep, Edit, Write, the wind rising; the owlets fly in at 6.75 | wind bed | an ascii scene that reacts as your agents work (shifted right of the fire) |
| 7.45–10.55 | CUT to a CLOSEUP on the owlets (3x, row 35) | they sit by the owl, the test runs | | agents and subagents |
| 10.55–15.0 | CUT to WIDE | the owlets fly off at 10.5, the finish at 11.0, dusk then night, the fire lights | the drop | |
| 15.0 | the montage's own CUT to the shore at night, WIDE | the second prompt | bed to the sea | |
| 16.5–16.85 | fast ZOOM to the moon (2.6x) | the readout under it | | the moon is your context |
| 19.15–22.15 | CUT to WIDE | the ask at 19.4: the crab walks up in the wide, the crablets come | the bird | it pings you when it needs you |
| 22.15–24.25 | CUT to a CLOSEUP on the crab's ask (2.4x) | the balloon; its eyes cycle the four variants every 0.55 s | | |
| 24.25–27.6 | CUT to WIDE; THE SAND'S STEP IS ON THIS CUT | "> yes" at 24.4, the crablets walk into the water, dawn, "Shipped." at 26.6 | the drop | |
| end | the montage to black 0.6 · the logo in 0.6 with the cell blinking · `xscapes.vibe.commonsmade.com` typed at 30 cps · the cat in 0.3, working 1.6, its finished face · hold 2.0 · fade 0.6 | | keys at the URL, the drop at the cat's face, bed out | |

THE SAND STEPS (measured 2026-09-18). The beach's tone is cube-exact by
design (s27), so as the palette runs from midnight to dawn the dry sand
jumps from the night ochre (cube 94, #875f00) to the dawn tan (137, #af875f)
in ONE frame, between hours 1.1284 and 1.1295 (traced on the exported
frames: row 41's dominant ground). A real night takes hours to cross that;
the trailer's takes twelve seconds. The shore's clock is set to reach 1.129
on the cut to the wide (stops 120:0.93, 194:1.129, 219:1.25) so the jump is
the cut's; the trace confirms it on frame 728, the cut frame. Not a product
change.

Three liberties in the render, all restored on return and none in the
product: `reduce.KittenDwell` 60 → 30 s so the fan-out can come and go inside
one shot; `companion.StudyNearEyeAlert` (the study override nothing in the
product sets) cycling `NearEyeRotationArt(0..3)`; and the pane's first line
typed (`trailerTyped`). The cat is `trailerCat`: the cast clip's portrait on
the transparent ground, one crop for both poses, exported as `cat.js` with
its own palette (scoped under `#cat` on the stage).

## The pipeline

The workbench is the session scratchpad (`trailer/` there holds
`node_modules`, `frames/`, `out-*/`); this directory is the record and is
synced from it at milestones. The browser is the compositor: every video
frame is the page's markup under CSS the script sets for that instant,
screenshotted at 1920x1080 by Playwright driving the installed Google Chrome
(`channel: 'chrome'`, no browser download). Glyphs are vector, so any zoom
is crisp. No CSS animations anywhere: every state is a function of t.

```
# 1. frames (from the repo root; skips unless the dir is named)
XSCAPES_TRAILER=$SP/trailer/frames XSCAPES_TRAILER_PART=cover go test -run TestTrailerExport .
XSCAPES_TRAILER=$SP/trailer/frames XSCAPES_TRAILER_PART=hero  go test -run TestTrailerExport .
# 2. the opener's frames (stills first, then all)
cd $SP/trailer && npm i playwright && node render.mjs opener frames out-opener --stills
node render.mjs opener frames out-opener
# 3. sound, then the cut
python3 audio.py opener opener.json out-opener/opener.wav
sh assemble.sh out-opener out-opener/opener.wav out-opener/opener-v1.mp4
```

- `opener.json` is the opener's whole timeline; every number he tunes is
  there (`wipe.glowAlpha` is the lit edge's strength, `sea.grey` the drained
  sea's ink, `cell.px` and `line.px` the sizes).
- The cover is `site.go coverLayersAt`'s sea rendered twice from one canvas:
  the plain glyphs the splash shows, and the same glyphs each in its own ink
  on no ground (`trailerCover`, `inkedGlyphs`). 168x57 rows fills 1080 at
  the cell 168 columns give 1920 (`XSCAPES_TRAILER_COVER_ROWS` to change).
- Fonts: Geist Mono 400/700 and JetBrains Mono 700 as Google Fonts serves
  them (`fonts/`, OFL), rewritten to local paths so the render is offline.
- X's spec, measured 2026-09-17: MP4 H.264 + AAC, 1920x1080, 30 fps, 8-12
  Mbps, under 2:20 (free), under 60 s loops; ~80% of views muted; the preview
  tile is an early frame; the first 3 seconds decide.

## Traps paid for

- The module is `github.com/donlucasx/xscapes`, and `heropage_test.go`
  declares a helper named `strconv`, so a test file in `package main` cannot
  import the package of that name.
- Playwright's `file://` URL needs an absolute path (`path.resolve`).

## The film's pipeline

```
node render.mjs film frames out-film --stills 9.6,16.2,26.8   # key frames first
node render.mjs film frames out-film                          # ~1100 frames, ~45 s
python3 audio.py film out-film/film.json <repo>/internal/notify/sounds out-film/film.wav
sh assemble.sh out-film out-film/film.wav out-film/xscapes-trailer-16x9-v1.mp4
```

`render.mjs film` is one page: the H1 window under a camera (`shots.json`,
keyframes in clip seconds, eased, `cut:true` for the montage's own cut),
captions (bottom, or `pos:"top"` when the subject fills the bottom), the
opener as an opaque layer that dissolves away at its `out.start`, the end
card, black. It writes `film.json`, the resolved timeline in video seconds,
which `audio.py film` scores: the keys, a wind bed under the vista and the
sea's wash under the shore (crossfaded at the cut, RMS about -33 dB), and the
SHIPPED `ask.wav` and `done.wav` at the product's own moments. The window is
fitted to the frame's height at zoom 1 (cell 14.78 x 24.55 px, 45 px of
black at the sides); pushed in, the camera is clamped inside the window.

## Not built yet

The square track (a second camera over the same frames).

## The post (2026-09-18 ~12:25 PDT)

https://x.com/donlucas/status/2101029250996617468 — the v4 trailer as the
main post, the vista timelapse GIF (`gif-vista-timelapse.gif`, V1 rendered by
`render.mjs clip`) on the reply with the entry link. The entry's link card
never rendered on X although its tags and image check out; media in the
reply stood in for it. Deliverables on his Desktop: `xscapes-trailer/`.
