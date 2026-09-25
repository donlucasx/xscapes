# The rainy window: the study, archived for the build (sessions 45-46, 2026-09-23 to 25)

Everything decided and drawn before the third scape is implemented. Nothing here is
wired to the product: the painters are study code in `internal/scenes/rain_*.go`, the
clips are a scripted arc, the pages are review artifacts. The build implements
`scape.Scape` off the reducer as the vista did (`internal/scenes/vista.go`,
`vista_live.go`), taking these rulings as the spec's inputs.

## The rulings, in order (verbatim in `_FEEDBACK.md` §Session 45)

- **The views.** Three go to the scape: THE BLOCK (midtown behind brick walk-ups, a lower
  roof across the alley), THE PARK (a plaza with a fountain, trees, a terrace) and THE
  HARBOUR (a working port across the water, a dock). Rooftops and the canyon are out
  (kept in the code as `RainView` 1 and 2). No default view yet: *"these are all
  mockups for now"*.
- **The work.** The rain on the glass: DROPS running down it, born only while the work
  reaches a column's threshold, so the pane is DRY AT IDLE and dries after the finish.
  In the park the FOUNTAIN's jet (height and spread) and the trees' SWAY are the work
  too, the wind being the work as on the vista. The harbour's traffic is the world's,
  at a constant pace.
- **The sky.** The shore's palette greyed a little toward overcast, keeping the hour;
  stars dimmer; the moon THE SHORE'S DISC with its phase, sinking behind the skyline as
  the context fills (a mockup disc in `rain_anim.go`; the build reuses the shore's).
- **The lit windows** are the world's (a dim share of the upper storeys by night, one
  switching now and then). Finished todos are STARS, as on the vista. The spend as city
  light stays carded.
- **The companion** is the cat, on the sill, right; its walk-up on needs-you is ONE RUNG
  (`companion.Near = 1`, 12x7 to 16x9). The balloon as on the shore.
- **The kittens** arrive FROM INSIDE in every view: ears over the sill's edge, the eyes,
  a held look, the hop up to sitting; they leave the same way (`sillRise`). On the sill
  they WANDER in their own lanes (a new spot two cells either way every few seconds),
  breathe and blink on their own periods. Three places at x 28/38/48 at 80 columns.
- **The park's trees and fountain are SPRITES** (his pick over a silhouette treeline and
  glyph foliage): authored bitmaps through the quadrant pipeline (`rain_park.go`:
  round, conifer, poplar, oak, willow, bush), the fountain a stone sprite with glyph
  water. Dark at night (every green through `parkNight`). The plaza plain stone, small,
  in the LEFT pane so the fountain is clear of the mullion; the bandstand in the right.
- **The harbour's water** is the shore's own `SeaFar`/`SeaNear`. The potted plant on the
  sill lives on the harbour only (`RainPot`). The frame is painted wood, pale, with a
  darker rabbet, in every view.
- **The clip**: a session arc of 16 s (idle · prompt 2 s · the level to 0.8 by 5 s ·
  kittens at 4/5.5/7 s · the ask 8-10 s · the work 0.9 · the finish 12.5 s · the
  litter home 12-14.6 s · idle), dusk 0.78 into night 0.98, the context 0.15 to 0.5.
  10 fps on the page, 30 fps for the videos.
- **Standing assumptions, kept**: no lightning; the writing on a desk under the sill;
  the composition mirrored (cat right); the layout scales to the window as the vista's.

## The files

| file | what |
|---|---|
| `internal/scenes/rain_study.go` | the window (glass, frame, sill, desk, plant), the block's skyline and walk-ups, the lower roof, the fire escape, the kitten sprites |
| `internal/scenes/rain_views.go` | the four other views (rooftops, canyon, park, harbour) and the park's blob crowns (`RainParkStyle` 0) |
| `internal/scenes/rain_park.go` | the park three ways; `RainParkStyle = 2`, sprites, his pick |
| `internal/scenes/rain_anim.go` | the arc, the drops, the disc, the sill arrival and wander, the cat and balloon in the clip, `PaintRainFrame` |
| `notes/rainshots/main.go` | the tool: frame pages for a look, the review page, the clips page (`clips.html` + `clip-*.json`), the video export (`XSCAPES_RAINEXPORT=<dir>`) |
| `notes/rainshots/out/` | ignored; the published clips page and its frame files |
| `notes/qt/render-fire.mjs`, `assemble.sh` | the video stage (the fire clip's), with `block.json` / `park.json` / `harbour.json` cameras beside it in the scratchpad |

## How to see it again

```
go run ./notes/rainshots <dir>          # today.html, sky-*.html, views.html, beats.html, review.html, clips.html
XSCAPES_RAINEXPORT=<dir>/frames go run ./notes/rainshots <dir>   # + block.js park.js harbour.js at 30 fps
cd <stage with playwright> && node render-fire.mjs frames out-park --name park && sh assemble.sh out-park none park.mp4
```

Pages: the stills https://claude.ai/artifact/EXLsGMYpEhe6nbpbQjW9nJ (V9 = the park three ways;
earlier versions hold every round) · the clips https://claude.ai/artifact/MFWAbAVzQ2MxWJvNPYbPtQ
(V6). Videos: `~/Desktop/xscapes-qt/xscapes-rain-{block,park,harbour}-16x9.mp4`.

## What the build has to do that the mockup does not

- Implement `scape.Scape` (`Update(c, t, act)`) for the window with a view field, off the
  reducer's `Activity`: level → the drops' births and the fountain/sway; ContextUsed → the
  disc; TodoDone → stars; TimeOfDay → the hour (overcast); the composer draws the cat,
  kittens, balloon and tail as `drawVistaWith` does for the vista.
- The kittens' sill arrival and wander on the reducer's litter (count, dwell, exits), the
  lanes derived from the width; the shore's `DrawKittens` is the model.
- The drops need the frame's real dt (a stalled frame must not teleport them); the birth
  rule reads the level at the birth instant, so the reducer's level must be sampled per
  frame and kept for the drops' lifetime.
- Layout for any window (the pane, the sill, the desk rows, the cat's column) the way
  `vistaLayoutFor` scales the vista; the 80x24 constants in `rain_study.go` are the
  design.
- The world's motion in the harbour (the ferry, the tug, sails, gulls, the lighthouse)
  and the block (the beacon, a window, the plane, the smoke) on wall-clock time.
- A minimum height, as the vista's 16 rows.
- `xscapes scape window` (name to rule) and a view preference; the site's clip.
