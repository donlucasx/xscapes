# Resume — xscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/xscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code hook schema — trust it, do not
re-derive it. Skim origin-chat.md only if you need the why; ignore ideas.md — it
is parked. Tell me where we left off, then pick up from ▶ NEXT — item 0 is a
question for me, not work for you.

Two things about how to work on this, learned the hard way in session 11:
build the instrument before trusting the picture, and check what the RENDERED
frame does rather than what the source says it should. And one from session 14:
do NOT drive Terminal.app (osascript, System Events) without asking me first.
```

## Where we left off (2026-09-05, session 16, INSTALLED, pushed, clean — SHAs in `git log`)

**The live tests ran, in both terminals.** From his screenshots (`_FEEDBACK.md` s16): resize
PASSES in Ghostty and Terminal.app, both directions · the Ghostty sun is the cube's peach (the
profile fix, live) · the moon is a disc at night · the companion margin holds · the stale-tip fix
held through a night (a clean sun at 11:32). **Three defects found in the pixels, none fixed:**
(1) two NIGHT scape rows (greys + the pink `d7afaf` moon tone, painted 01:55–02:50 by the sweep)
left at window rows 28–29 of 61 in Terminal.app — the model with Terminal.app's rules is clean
for every grow/drag (`apple_grow_probe_test.go`, kept), cause unknown, needs his 2 AM history;
(2) at 123x55 after a WIDTH+height drag: a 6-column patch of scape cells above the band at the
left and the band's last column painted from rows below — the trace (`/tmp/apple.bin`, 34
resizes, width 120→130→123) replays CLEAN in the model, so the terminal does something on a
width change the model does not know (its alt-screen width rule was never measured); (3) the
moon's tips show Terminal.app's U+2580 hairline (*"a tad sloppy"*); the ▄ swap fixes it only if
U+2584 is bottom-exact — his printf screenshot was a clipboard image, unresolved.

**Shipped and installed (one restart shows it):** a finished subagent's kitten SWIMS OFF along
the top lane over 6 s (`reduce.KittenExit`, `State.KittenExits`, `Cat.DrawKittenExits`; the
count still drops at the end event). Built for his pick and then NOT taken: eye fills
(`canvas.Layer.PlotOn`, `Cat.SetEyeFill` none/coat/socket, `xscapes -eyes`, artifact "The
Companion's Eyes") — **LOCKED: the eyes stay holes.** Login flip 7 (donlucasx): s15's two
artifacts are read-only here.

**Later the same afternoon (his *"go ahead"* + a saved printf screenshot):** defect (3) is
FIXED — U+2584 is bottom-exact in Terminal.app, so split cells are drawn as ▄ with the colours
the other way up there (`term.LowerHalf`, `DetectSplit`, `XSCAPES_SPLIT`); the hairlines on the
disc and across the sky are gone. Defect (2) is EXPLAINED and documented, not fixable:
Terminal.app's alt screen RETAINS rows at their widest on a width change, erase reaches the
visible width only, and his 123.6-column window drew a partial 124th column of retained cells
(`notes/width-audit.md`, `notes/widthprobe`, model switch `screen.retainWidth`,
`TRACE_RETAIN=1`). The six-column patch above the band is NOT reproduced under either rule —
open. A read-back of his live window found the s14 #2 scrollback CORRUPTION in the main buffer
(`✻ Tomtotal reported 1259,cfetched)1259,sunique 1259`, rows merged where one had spaces) —
the mirror wrote what the model held; that session was untraced.

**▶ NEXT:** 0. a TRACED working session in Terminal.app (`XSCAPES_TRACE=/tmp/apple.bin xscapes
claude`) that he works in until the mirrored rows corrupt, then a read-back of the same window
(his OK stands for read-only) → diff the model's kept rows against the buffer, fix the divergence ·
the patch above the band (a hosted probe under scripted width+height drags, pixels via
`screencapture -l` of the script's own window) · the working-session items still unseen:
bubbles + sound, kittens live (now with the swim-off), the mirror during work and its exit
replay, the transcript start · then the Commons sequence: rebuild both binaries after any fix → `xscapes -site site` →
his 45–60 s Terminal.app recording → `site/COMMONS-PROMPT.md` → he prompts the rebuild (page +
gallery only, no web twin) → publish → SUBMIT (closes 09-17).

## Where we left off (2026-09-04 afternoon, session 15 continued, INSTALLED, committed `b9d65e7`, pushed)

**His first run in Ghostty, two reports, both measured, both SHIPPED on his *"proceed w ur
recommendation to standardize the experience"*.** (1) The sun differed: Ghostty took the truecolor
profile and painted the palette's raw blend, a flat tan; Terminal.app's cube quantiser turns that
tan into peach over cream. **Now every terminal gets the cube** (`term.DetectProfile` returns 256
unless `XSCAPES_COLOR=truecolor`; COLORTERM no longer consulted; `internal/term/profile_test.go`).
(2) The layout broke on a resize: Ghostty 1.3.1's alternate screen (read from its source) keeps
content at the TOP on a grow and moves the cursor WITH its row on a shrink, DECRC verbatim, where
Terminal.app is bottom-anchored with the cursor unmoved. The host's tick encoded Terminal.app only.
**Now `host.Rules`** (`GrowPushesDown`, `ShrinkKeepsCursor`), picked by `host.RulesFor(TERM_PROGRAM)`:
Terminal.app keeps its measured sequences byte for byte, everything else gets the xterm-like set
(no SU on a grow; `RebindShrinkAltFollow` = SD k + CUD k on a shrink). Red-first test
`internal/host/ghostty_resize_test.go` (Ghostty's rules in the model: `resizeGhostty`,
`screen.restoreAbsolute`), eight geometries. Also measured from his screenshot's pixels:
Terminal.app's U+2580 glyph starts 5px below the cell top (12 of 30 px), so every half-block
edge shows a hairline of the lower colour — the "thin lines through the sun"; Ghostty is
pixel-exact. Instrument `notes/sunprobe`. `~/.local/bin/xscapes` rebuilt (new inode, `-info`
verified: profile=256 under ghostty/Apple_Terminal/iTerm.app). Commons research: `research/
commons-submission.md` §8 — a web twin is feasible (Durable Objects, WebSockets, cron proven live;
renderer compiles to wasm, 3.4 MB); not started, his call.

**▶ NEXT (afternoon):** 0. CONSOLIDATE first — three sessions ran on 09-04 (the morning one that
shipped the ramps and the moon, `1eb6e92`; this one, `3601ac0`…`9414e96`; a hub session whose wrap
rewrote the hub `MEMORY.md` xscapes line): one pass over `RESUME.md`, `_FEEDBACK.md` s15 and the
xscapes memory, then the live tests · he restarts `xscapes claude` in Ghostty and resizes — the fix is proven in
the model, not yet live · the two fully BLANK sky rows in his screenshot are NOT reproduced offline
(`XSCAPES_TRACE=/tmp/ghostty.bin xscapes claude` in Ghostty, then resize, would replay them) · the
sun still reads rough on 256 in both terminals: the disc is quantised cell by cell against the
sky gradient (peach over cream) — "quantise the disc once" is the small follow-up, his pick ·
the companion's eyes are HOLES to the scene by day (measured from his 12:49 screengrab: the eye
cells are the sea's colour; by design the eyes sit in gaps the body bitmap leaves) — fill the two
eye cells with fur before the glyph, his pick · still his from the morning: solid moon or hue rim ·
kittens vanish in <7s · the mirror corruption needs a mirror-era trace · publish the page, submit
the entry (closes 09-17) · the Commons web twin, if wanted (~2-3 days).

## Where we left off (2026-09-04, session 15, INSTALLED and pushed)

**His four answers at the start** (verbatim in `_FEEDBACK.md`): the submission page is NOT
published (*"we need to work on some of the pending items and update it later"*) · gradients:
**cube-path, and ONLY that** (re-timing the day, a stable moon colour and the grey night were
offered beside it and not picked) · Terminal automation: *"yes, only for this session"*, tty rule ·
the scroll glitch (report #2): did not check. An account switch at 23:49 (session limit, 6th flip).

**The cube-path gradients are BUILT, MEASURED, and INSTALLED** — he saw the page and said
*"Install it"*; committed, pushed, `~/.local/bin/xscapes` rebuilt the same night.

- `internal/term/ramp.go`: `term.Ramp`, a gradient as ONE PATH through the 240 fixed palette
  entries. The cheapest walk from the quantised start to the quantised end: the sum of SQUARED
  CIE76 steps (two of 15 beat one of 30), plus an off-line charge along the way (a tone's distance
  from the true ramp, in Lab, against its nearest sample), plus a charge for flattening a channel
  lead the truth states strongly, scaled by the tone's own chroma (a muted slate is a tint, a
  saturated violet is a stripe). Cube entries stay inside the box the two ends span; greys by
  brightness. Knobs: `RampHop` 26 (the grey-ramp entries beside a slate blue sit 24.2–24.7 away;
  at 24 the walk was refused the one detour that beats the 32-point green step every dawn took),
  `RampLambda` 1. `XSCAPES_RAMP=0` puts every row back through the row-by-row quantiser.
- `canvas.SetBGRamp` binds a cell to the span of a ramp it covers; `resolve` on 256 takes the
  path's tones (both halves of a split cell), the quantiser everywhere else. `BGAt` and truecolor
  report the true colour, unchanged. A glyph over a ramp gets the path's tone as its background.
- `shore.paintBG`: the sky and the open sea are painted through ramps. Nothing else moved.
- Tests, all reading the RENDERED frame at his geometry: `internal/term/ramp_test.go` (ends are
  the quantiser's, every tone a real entry in the box and never repeated, greys walk the grey
  ramp one step at a time, the 05:00 sky has fewer hard edges than rounding, equal shares);
  `internal/scape/gradient_path_test.go` (all 48 half-hours: no tone comes back, at most two
  grey/colour crossings, at most two steps of 25 or over, largest 33 — the measured floor);
  `TestTheSkyHueDoesNotWander` moved onto the rendered frame. `go test ./...` green.
- **Measured** (`notes/gradientaudit`, which now has a `-dump` flag, a `hard` column, and a
  "256 before" frame per hour): hard edges (ΔE ≥ 20) **sky 84 → 65, sea 120 → 99**; half-hours
  with a step ≥ 30: **29 → 6**; the largest step is UNCHANGED at 33 (the first green step out of
  the daylight zenith with blue pinned by a pale horizon, 08:00 and 14:30–15:30; no path through
  the cube gets under it). Two half-hours got ONE edge worse: 17:00 sea (the path takes `#5f87d7`
  for one row where rounding did not) and 21:00 sea (grey → teal → blue over two steps where
  rounding jumped grey → blue in one). Tones per region fell a little (sky 8.3 → 7.4) because a
  path never revisits a tone. Rejected on the audit before this objective: minimax (smaller steps
  by leaving the ramp: a dawn through hot pink), a drift bound (cuts the graph where the ramp
  passes a neutral nothing sits near), a hybrid end rule (4 half-hours changed, 2 better, 1 worse).
- **Page**: `assets/frames/gradients.html` (ignored; `go run ./notes/gradientaudit -html
  assets/frames/gradients.html`) and the artifact **Sky and Sea Repainted**, https://claude.ai/code/artifact/3bf5eb42-4096-40d7-b607-ac88b388d2d6
  (this account's; "Sky and Sea by the Hour" `b465bc20-…` belongs to the prior login, read-only).
- ⚠ **The audit's first before/after came out IDENTICAL** and was nearly quoted: rendering the
  "before" frame set `term.Ramps = true` afterwards, so `XSCAPES_RAMP=0` lasted one hour of 48.
  Fixed (save and restore); the before run now matches the pre-change baseline table on all 48
  half-hours, which is the check that makes the numbers above quotable.

**His first live look, 00:36, at 124x52** (`_FEEDBACK.md`): *"looking good, on a first
impression, the moon looks worse than before"* and *"more separation between the far right edge
and the main companion ... a bit"*. Measured first (`notes/moonprobe`, cells around the moon with
the paths on and off): **the moon was identical in both** — at the 23 rows a 52-row window gives
the scape, the disc's radius is 1.92 rows, just under two, so its tip rows fell outside it and the
three that remained were all seven wide: a rectangle, pre-existing, invisible at his earlier 62-row
windows where the radius is two. **FIXED**: the disc is sampled at half rows and painted with
U+2580 where its edge falls inside a row (`canvas.SetBGHalves`, a new background primitive; the
moon painter in `shore.go`); `TestTheMoonIsRoundAtEveryHeight` holds it at 18–30 rows, red first at
22 and 23. **FIXED**: the companion's right margin grows with the width, 2 + w/32 (5 at 124, was
2; `compose` in `live.go`, `TestTheCompanionKeepsItsDistanceFromTheEdge`). Both installed the same
night (commit, push, both binaries). His session has to be restarted to show them.
⚠ **The install itself bit him**: `cp` over the already-executed `~/.local/bin/xscapes` left macOS's
cached signature stale and the kernel SIGKILLed it (*"zsh: killed"*, exit 137) twice. Fixed by
replacing through a NEW inode (`rm -f` then `cp`) and verified by RUNNING `-info`. **Install rule
from now on: `go build -o xscapes . && rm -f ~/.local/bin/xscapes && cp xscapes ~/.local/bin/xscapes
&& ~/.local/bin/xscapes -info`** — a grep for a marker proves the bytes, not that they run.

**Morning of 09-04, the moon, four more defects and a study** (`_FEEDBACK.md`, the peer session
"xscapes-e6" relayed his live 133x61 run; then his own screengrab and *"I liked how the moon/sun
looked in the original mockup — why changed?"*). FIXED and installed (f3dfe52 and after): the tip
row's centre cell at exactly the radius counted as inside (a pip at 12 and 6; tie now out) · the
unlit limb blended into the night sky so a moon one column into its phase looked bitten (earthshine
lifted) · **`SetBGRamp` never cleared a cell's half-row record, so the moon's tips from the night
stayed in the sky until morning as grey notches around the sun** — found by sampling his
screenshot's pixels (26,26,26 over 180,180,180 = night sky over night moon, in a blue sky); two
red-first tests · the shading split blended a flat moon cell toward a neighbour's half-cell MEAN
(grey over olive) and a far star over a tip cell replaced the halves (dark dot on the crown) — both
fixed in the canvas. The mockup question is answered by `notes/moonstudy` → artifact **The Moon's
Edge**: solid (ships) · a same-hue rim one tone darker (`Shore.MoonRim="hue"`, study-only) · the
mockup's fade, which is the grey fringe s11 removed. **⏸ HIS PICK: solid or the hue rim.**

**Not done**: a day lived in it with the paths on. The moon at 20:33 is still sun-coloured:
that was the option he did not pick. `~/.local/bin/xscapes` is a COPY of the repo binary, not
a symlink; rebuild both after any change.

## Where we left off (2026-09-03, session 14 WRAP, pushed, tree clean)

**He lived in it for an evening and came back with six reports** (verbatim in `_FEEDBACK.md`).
State of each:

1. *Chat history looks good on a first impression.* ✓
2. *Scape pixel lines broke at the top-left after scrolling back down.* **NOT REPRODUCED in the
   buffer**: a scroll-up/scroll-down cycle over a scrolling band at 124x62, read back, shows a
   clean band and a clean seam. Likely a Terminal.app display artefact that the 50-frame full
   repaint (~4s at 12fps) clears. ⏸ Ask him whether it healed on its own.
3. *A subagent swimming on the sand at the tide line.* **FIXED**: swimmer lanes end two rows
   above the shore's mean waterline (`sh.SandTop()-2`), not a fixed distance above the cat.
   Unreported but in the same grab: the DONE balloon carrying a whole paragraph across the sea.
   **FIXED**: balloon text capped at 44 runes with an ellipsis (`companion.MaxBubbleText`).
4. *Does the moon look correct?* At 20:33 it is a tan/salmon block because the palette blends
   dusk (18:00) straight into midnight (24:00) over six hours: at 20:33 the sky is still 58%
   dusk and the body is 58% sun. Real LA sunset in early September is ~19:15. Part of #5.
5. *Review all sky/water gradients; 256 translations abrupt; smoother, cleaner; REVIEW TOGETHER
   BEFORE ANY CHANGES.* **Measured, nothing changed.** `notes/gradientaudit` renders every half
   hour at 124x27 and reads the 256 frame back: distinct tones, band heights, largest CIE76 step.
   Page: `assets/frames/gradients.html` (ignored, regenerate with `go run ./notes/gradientaudit
   -html assets/frames/gradients.html`) and the artifact "Sky and Sea by the Hour"
   (https://claude.ai/code/artifact/b465bc20-2c01-4062-b7c3-937c92a1c909). Findings: 22:00–01:30
   greys only (ΔE 5, smooth, colourless); 02:00–03:00 a pink dawn band on grey (ΔE 18); 03:30–21:00
   a saturated blue slab over grey in the sky and a teal slab over grey in the sea (ΔE 25–33, the
   edge he sees); the body is sun-coloured 03:00–21:00 and flips lavender/grey at night. Four
   options put to him: cube-path gradients (recommended) · re-time the day to real sunrise/sunset ·
   a stable cube-entry moon · keep the grey night. **⏸ HE WRAPPED BEFORE ANSWERING. His call.**
6. *The beginning of the transcript broke (repeated banners, two garbled rows).* **PARTLY
   DIAGNOSED**: the repeats are Claude Code re-rendering its header while its own permission
   warnings (his 747 rules) scroll a 35-row band; a plain terminal would keep them too. The
   garbled rows are the MODEL diverging from the terminal. Diffing the model against Terminal.app's
   readback of a real startup found one divergence, ESC ( B drawn as a "B" -- **FIXED**. The
   frozen-readback comparison that would find the rest was INVALID (wrong window, see the
   incident) and is still to be done properly.

⚠ **INCIDENT, mine.** The second traced startup addressed Terminal.app's `front window`, which
was HIS live window: the script read its history, brought it to the front and typed `/exit` +
Return into his Claude session, which put up "Exit and stop tasks" over his running agents. He
was told to press Escape. Lesson saved to hub memory: resolve the window from the `do script`
tab's tty, assert before any keystroke, never "history contains" for cleanup. **No Terminal
automation without his explicit OK from here.**



**⭐ SCROLLBACK SHIPS: the mirror is built, tested and seen working in Terminal.app.**
After his gate answers (*"not sure if it flashed ... stable when I scroll up"*, then
Accessibility granted so Shift+Page Up could be sent: the view holds while rows land,
in pixels), the revised plan was built end to end:

- `internal/host/screen.go` — the model, promoted from the tests: cells carry fg, bg and
  attributes; two real buffers (47 swaps without clearing, 1049 saves/clears/discards);
  rows that leave the alternate band by scrolling, or that a shrink destroys, are kept
  (`capture`); rows the host itself scrolled in and never wrote are not; `rowANSI`
  writes a row back as bytes (round-trip tested).
- `Host.History` — every byte the host sends is fed through the model; once a tick the
  kept rows go to the main buffer in ONE write (`MirrorBatch`: `ESC7 ?6l ESC[r ESC[0m
  ?47l` · CUP · row · ... · `?47h` · band · `ESC8`), starting on the row the shell left
  its cursor (DSR asked BEFORE the child starts; the key forwarder keeps that one reply)
  and, once the buffer is full, scrolling first so the newest row touches the band.
- `Host.Replay` — after the pty is drained (500ms deadline) and the alternate screen given
  back, the mirrored rows are printed again under a dim separator, because Terminal.app
  keeps only the oldest few across `1049l`. Gated on `TERM_PROGRAM=Apple_Terminal`.
- `xscapes claude -history` (default on in Terminal.app, off elsewhere; implies `-alt`).
- Seen live: 40 lines through a 17-row band → 24 rows above the band during the session,
  starting right under the command line. ⚠ The first replay readback I quoted as success
  carried a defect Kimi round 2 read and I had not: rows written over longer survivors
  without erasing ("LIVE LINE 80"), and the final screen missing. **Both fixed the same
  evening** (see the audit note, "Kimi round 2"); the replay now holds the transcript AND
  the band's final screen, 40 of 40, nothing left over. Early in a session the main
  buffer's unused bottom rows show as blank rows between the transcript and the band until
  the transcript has filled them; inherent.
- Tests: model (capture, host rows skipped, shrink capture, 47 vs 1049, ANSI round trip),
  `MirrorBatch` walk-then-scroll, `takeCPR`, and the hosted end-to-end (24 lines, the
  eight that left are in the snapshot model's main buffer in order; a control with the
  mirror off finds nothing there). `go test ./...` green.

**Kimi round 2 ran on the shipped code**: REVISE → two MAJORs in the exit replay (overwrite
without erase; the final screen omitted) plus four minors, ALL FIXED with tests; its one
speculation (the main buffer moves on a grow with real scrollback) was MEASURED true and the
write row now follows. `go test -race ./internal/host` is clean. **Not yet**: a real
`xscapes claude` session lived in with the mirror on (his day). Known limits: wide glyphs misalign a
mirrored row from that glyph on (the model advances one cell); the replay repeats the
few rows Terminal.app kept; other terminals get the rows after exit only.

## Earlier in session 14

**Two deliverables and a locked decision.** Resumed under Fable; both ▶ NEXT questions
were put to him and answered, then he asked for the scrollback plan to be AUDITED before
it was built, with Kimi if needed.

1. **The submission page exists**: `site/index.html`, one static file, five frames from
   the real reducer at 256 colours, copy in `site/template.html`, regenerated with
   `xscapes -site site`. Publishing steps and the Commons prompt in `site/COMMONS-PROMPT.md`.
   ⚠ **Still not published, still not submitted** — his hands (login-gated); deadline 09-17.
2. **The scrollback plan was audited and replaced** (`notes/scrollback-audit.md`; Kimi round
   1: REVISE, direction right, four gaps, all closed or budgeted). Three measurements, read
   back by machine (`history of tab` returns the alternate screen's cells — retire "nothing
   reads cells back from Terminal.app"):
   - the alternate screen has NO history (premise holds);
   - **every shrink misplaced Claude's input box** — the terminal leaves the cursor, the host
     restored it into a band that no longer held its row, Terminal.app answers row 1. Report 1,
     reproduced on demand, **FIXED and shipped** (`RebindShrinkAlt`, red-first in the
     corrected model, eleven geometries in the real terminal, mutation-proven);
   - **DECSET 47 switches buffers without clearing**, so rows can be written into the
     terminal's OWN scrollback while the band stays up: 400 rows at 5ms, alt intact.
   ⭐ **HIS RULING: *"Go: shrink fix, then mirroring."*** Own-the-viewer is dead. The plan is
   at the end of the audit note (1.5 days): promote the model with full SGR and two buffers,
   tee the agent's bytes, mirror rows that leave the band into the main buffer (DSR for the
   shell's row BEFORE the child starts; SGR reset; batched per tick), drain the pty before a
   Close replay on `Apple_Terminal`. **Step 2 is gated on him watching `mirrorprobe` once**
   for flicker and for the view snapping while scrolled up; synthetic wheel/keys are refused
   on this machine, so his eyes are the instrument.
   ⚠ One flake seen once: `TestEveryBandRowIsRepaintedAfterAResize` failed on the run right
   after a mutation revert (old scape rows in the band); three clean reruns since. Timing
   harness (300ms resize / 600ms snapshot), pre-existing.

## Where we left off (2026-09-03, session 13, HEAD `9452bc6`, pushed, tree clean)

**The resize damage was OURS, in two separate ways, and it took him reporting it
four times to establish that.** Both fixed. The durable output is not the fixes,
it is one measured fact and one rule.

**THE FACT** (`notes/contentprobe`, read off the screen by eye, both directions):
Terminal.app's ALTERNATE screen anchors CONTENT to the BOTTOM edge — a grow of N
pushes everything DOWN by N, a shrink pulls it UP and destroys the top N rows —
and the CURSOR moves with NEITHER. Grow of +21: `ROW 01` landed on screen row 22
with the cursor still on row 1.

**THE RULE** ⭐ **an instrument can answer a question you did not ask.** FOUR did,
today, and each one produced a confident wrong answer to him:
- `screen` discarded SGR, so a row erased under a colour read as blank.
- every resize test ran `AltScreen: false` while production runs alt.
- `screen` implemented NO relative cursor motion — CUU/CUD/CUF/CUB/CHA, which is
  all Claude Code uses — so every reconstruction of the agent's band was fiction.
- `notes/anchorprobe` measured the CURSOR and I wrote it up as CONTENT. Twice.

**Shipped, all mutation-proven in both directions:**
1. `clearRowsBare` resets SGR before erasing. An erase fills with the CURRENT
   background, and the clear runs off a timer with no relation to where the agent
   is in its output — so it was painting rows in whatever colour Claude was
   mid-draw. That was his wall of black.
2. **Grow:** `Rebind` takes a scroll-up count and undoes the terminal's push over
   the full screen before anything is painted. Without it the agent's first row
   lands at screen row 31 after a 30→59 grow. Moving ROWS needs no model of the
   UI in them, which is why the host may do it.
3. **Shrink:** `drop` restored for BOTH screens. I made it alt-exempt earlier the
   same day on the strength of the cursor reading; that was wrong and is reverted.
4. The model learned SGR, the alternate screen, and every cursor motion Claude
   emits. `XSCAPES_TRACE` + `TestReplayTrace` reconstruct a real session's screen
   from the bytes the host sent; `TestTraceRightEdge` separates "renderer painted
   short" from "the host's width belief was stale".

**An external audit earned its keep** (*"why dont u use agent kimi"*). Kimi called
the anchorprobe's configuration gap before any measurement did, and its F1
mechanism was right while my refutation of it was wrong. Its other live finding is
UNFIXED — see ▶ NEXT.

⚠ **SCROLLBACK IS NOW A REQUIREMENT AND IT HAS NO CHEAP ANSWER.** Measured: the
MAIN screen (the only one with history) is anchored TOP on height — easier than
alt — but **reflows totally on WIDTH**, every full-width row becoming two, and the
host cannot undo that because it moves rows while reflow changes how many rows a
line takes. And tmux's stacked-pane border leaves a seam no styling hides, against
a sky that changes colour hourly. He ruled: *"should feel like an embedded
experience"*. Seamless + history therefore forces xscapes to own the scrollback.
⏸ **Estimated and NOT started — his call.** See ▶ NEXT #1.

## Superseded — session 13's first half (HEAD `a0e9a83`)

**He reported the resize damage a third time, and this time it was ours --
both halves of it.** Session 11 had written it off as Claude Code's screen
being moved by the terminal. That verdict came from an instrument that could
not see either defect.

1. **The clear was painting.** `clearRowsBare` emitted `ESC[2K` with no SGR of
   its own. An erase does not write spaces, it fills with the CURRENT
   background -- and it runs off the resize TICK, a timer with no relation to
   where the agent is in its output. Claude paints backgrounds constantly, so
   the cleared rows came out solid in whatever it was mid-draw, then scrolled
   into scrollback. That is the wall of black he hit scrolling up. Red before
   the fix at 5 rows on a shrink, 9 on a grow.
2. **`drop` is a MAIN-screen correction, applied to both screens.** Measured
   with `notes/anchorprobe` (cursor parked mid-screen, window resized from
   outside, read back with DSR): on the same 11-row shrink the main screen
   comes back on row 10 and the ALTERNATE screen on row 21. Main keeps the
   bottom and slides everything up; alternate has no history, keeps the top,
   truncates. `xscapes claude` runs on the alternate screen, so the correction
   subtracted rows that never moved and walked the clear up into the
   transcript -- at a big enough shrink, into the input box. Red before the
   fix: 3 of the top 13 rows survived.

**Why eleven resize tests passed through all of it:** `screen` stored runes and
discarded SGR, so a coloured row read as blank; and every resize test ran
`AltScreen: false`. The model now carries a background per cell and can take
the alternate screen. Both fixes are mutation-proven in both directions -- the
main-screen control goes red if `drop` is merely deleted.

⚠ **The general lesson, worth more than the fixes:** a harness that no-ops a
platform behaviour, or never enters the mode production runs in, keeps
returning a clean bill of health. Two sessions trusted it.

## Where we left off (2026-09-03, session 12, HEAD `c9a019e`, pushed)

**The rename is finished.** `~/Documents/claude/xscapes/`, `XSCAPES_*`,
`~/.config/xscapes/`, and the `# xscapes:v1` marker on the twelve installed
hooks. This closes the migration sessions 8-11 deferred on purpose.

The risk was never the strings, it was the **marker**: it is uninstall's only
handle on its own work, so changing the constant alone would have left twelve
hooks nothing could see -- uninstall reporting zero, install adding a second
copy beside each. `install.go` now writes `# xscapes:v1` and RECOGNISES
`# asciiscapes:v1`; emptying `legacyMarkers` turns the tests red and the failure
shows the orphan exactly. The applied diff on his settings.json was 24 lines,
all marker; the Funk.aiff hooks, the VERCEL_TOKEN guard, the statusLine and 747
permission rules came through byte-identical.

`ASCIISCAPES_*` still WORKS and says so on stderr (`internal/envx`) -- his call,
over a hard cut, because a renamed knob nothing reads is the failure where the
value looks applied and the measurement is silently wrong. Live state moved
with it, so `xscapes tune` still folds the corpus: 12 sessions, 19,904 events,
155h52m, verified after the move.

Still saying asciiscapes on purpose: `legacyMarkers`, envx's `legacyPrefix`, the
verbatim quotes in `_FEEDBACK.md`, `origin-chat.md`, and the superseded bullets
below. They are the record, not leftovers.

⚠ **Two `xscapes claude` scapes were running through the migration and are now
deaf** -- their sockets are under the old path. Restart with `xscapes claude`.

⚠ Nothing else changed. **The three things still waiting on him are unchanged**:
submit the entry, the worry trigger, the banding decision. See ▶ NEXT.

## Session 11 (2026-09-02 → 09-03, HEAD `e886b7e`, pushed)

**Live: https://github.com/donlucasx/xscapes** (public, MIT). Sixteen commits.
Everything below is pushed and the tree is clean.

**The session in one line: nothing that was marked done actually was, and the
way that got found was building instruments rather than looking at pictures.**

Seven defects in shipped features, every one invisible in every study:
the sea and sky had **no colour at all** on Terminal.app for most of a working
day · the 256 sky was the wrong HUE, not just banded · **an electric night that
I shipped AND pushed** before catching it · the sun had a grey fringe · the moon
became a blob (my own regression) · a resize left sky sitting in the agent's
transcript · kittens lost an eye to their neighbour's seam. Plus two channels
measured for the first time and both found broken: the sea's dynamic range was
collapsed into a fifth of itself, and **the companion's alarm is on 37% of the
time.**

### The instruments, which are the durable part

- **`xscapes tune`** — folds every real spool through the reducer OFFLINE,
  samples once a second, prints distributions; `-sweep` re-folds for each
  candidate setting. `replay` was never this: it plays into a live scape at
  wall-clock speed, so checking one value meant watching an afternoon.
- **`internal/host/screen_test.go`** — a small terminal (CUP/EL/ED/DECSTBM/
  DECOM/save-restore/autowrap/region scroll, plus two resize behaviours). Drives
  a REAL Host with a real pty and replays every byte. Eleven resize modes.
- **`xscapes shades`** — one frame three ways in HIS terminal, at his window
  size, for the smoothing question a browser cannot answer.
- **`-day`** — five panels an hour: truecolor · 256 smoothed · 256 raw · the
  rejected shade blocks · the glyph cells the 256 pass changed.

### What shipped

- **Sky and sea chosen from what the cube holds.** Daylight colour loss 40/48
  half-hours → 0/25. Six day keyframes, not four.
- **`Index256Keeping`** — hue-preserving background quantisation, weighted
  (`hueWeight` 8) so a ramp cannot wander into cyan and back. Greys are left
  ALONE, which is the guard that stops it forcing an electric night.
- **Cell splitting (U+2580)** for gradients; shade blocks built, measured
  (11→14 tones) and REJECTED as stipple, kept behind
  `XSCAPES_SHADE_BLOCKS=1`.
- **Stars for completed todos** — the last unbuilt channel in the locked table.
  The hook now emits a real `todo` event; it had classified TodoWrite as an op
  and stopped, so `n`/`of` were never filled.
- **The floors LIFT the sea instead of clamping it.** Bins went 45/32/8/5/4/2/4
  to 18/24/21/17/9/5/5.
- Resize clear allows for the terminal moving things; kitten faces draw after
  every body; sun solid, moon at `rr`; `emit -n/-of`.

### ⚠ Three things waiting on him

1. **Submit the entry.** Nothing in any doc records it and the plan said ~Sep 1.
   Editable until the 17th, so a rough submission costs nothing.
2. **The worry trigger.** The cat is alarmed 37% of active time — 31 episodes,
   median 15m27s, longest 2h02m, **65% raised by a SINGLE error**. `worried` is
   set by any error and cleared only by his next prompt. The clear rule is sound
   ("hooks can tell us a command failed, never that the code is fixed"); the
   TRIGGER is too loose. **The brief locks this channel, so raising the bar is
   his call.** Recommended: require two errors in a window, or one the agent does
   not recover from.
3. **The banding call** — `xscapes shades -only 2` against `-only 3` at full
   window size. I have judged it wrong twice by judging it small.

### ⚠ The limitation that is not a bug

**A resize scrambles the AGENT's own text, and the host cannot fix it.** The
scape's half is provably right in eleven resize modes. What is left is Claude
Code's screen being moved by the terminal — and it emits nothing at all on a
resize, so it stays where the terminal left it until a keystroke heals it.
Repainting it means modelling it, which is the emulator decided against on
09-01. His call whether that reopens; with the deadline where it is, it should
not.

## Earlier (2026-09-02, session 10, `0e7f265`)

**Live: https://github.com/donlucasx/xscapes** (public, MIT, tagged `v0.2.1`).
Milestone 1 COMPLETE. The agent runs inside the scape, on the **alternate
screen**, and he has used it all day.

**⭐ NEW RULING, and it reverses a standing recommendation: TARGET TERMINAL.APP.**
*"at this point I want to optimize the experience for terminal.app which should
be the most used?"* I had been telling him to install a truecolor terminal;
he is not going to. **That makes cube-exact colour the general rule for the
whole scape, not a special case.** Only the sand and the sun are cube-exact
today. Everything else is still chosen for truecolor and mangled by the 256
cube -- measured, the sea's depth gradient collapses THREE different blues onto
one saturated teal, `rgb(0,95,135)`, errors 38-47. See ▶ NEXT #1.

**Install note that cost an hour**: `~/go/bin` is NOT on his PATH. His only
binary is `~/.local/bin/xscapes`. Build straight to it:
`go build -o ~/.local/bin/xscapes .`

### Shipped 2026-09-02, session 10

- **`xscapes claude` now means the agent INSIDE the scape.** `-beside` keeps the
  tmux layout, `inside <cmd>` hosts anything, `-print` works on all three.
- **⭐ THE ALTERNATE SCREEN.** The resize bug was never an off-by-one. Claude Code
  emits ZERO bytes on a resize and places its input purely by relative moves;
  growing a window makes the terminal pull scrollback back in, which pushes the
  agent's UI out of its band. Controls that pinned it: no-resize works, plain
  Claude plus the same resize works, hosted plus resize fails. The alternate
  screen has no history to pull back. Cost: output scrolling out of the band is
  gone rather than saved. `-alt=false` takes both back.
  ⚠ An earlier note in this repo said Claude repaints on SIGWINCH. That was
  measured on a FRESH session and was its startup draw. Corrected in
  `notes/claude-terminal-emissions.md`; it cost a morning.
- **DECSTBM homes the cursor -- three times.** Paint, exit, and resize each had to
  learn it: save the cursor before touching the scroll region, place it after the
  last time you touch it. All three are in tests now.
- **The beach, in four corrections from him.** One flat sand tone that varies day
  to night; a ragged waterline (2.9 rows of relief resting, 7.5 at full) scaled
  to fit rather than clipped; the writing band INSIDE the beach's share; no black
  seam. Sea and beach ramps anchored to the MEAN waterline: one row carried 87
  distinct tones and 58.6% of open-sea cells changed every frame, now 24 and 3.7%.
- **⭐ The moon is the sun by day.** His call. Context is carried by phase AND
  altitude, so a second body would have needed a second encoding for one
  variable; one disc whose colour follows the hour is the only version that
  holds. It also fixed a rule violation: MoonVis bound an AGENT channel to the
  clock, so the context readout was invisible all working day (+10 luma at noon,
  now never below +54).
- **The statusline is chained**, so context reaches the sky for the first time.
  Verified end to end; his own statusline renders unchanged.
- Scape takes 9/20 of the window; `-scape N` overrides. Full repaint every 50th
  frame so stale cells heal.

### Shipped 2026-09-01, session 8

- **⭐ `xscapes inside [command]` — the agent runs inside the scape.** One
  window, no tmux, no seam. `internal/host`:
  - **A pty on stdlib syscalls** (darwin and linux), so the no-dependency rule
    holds. The child is sized to the band, not the window, so Claude lays
    itself out inside it with no further help and its text can never collide
    with the scape.
  - **`Filter`** strips DECSTBM out of the child's stream. Claude emits `ESC[r`
    once at startup; that one sequence resets the region to the whole screen
    and its next scroll would walk over the scape. Handles the sequence
    arriving split across reads.
  - **`Band`** splits the window: a third to the scape, clamped to 8..20 rows,
    the rest to the agent. 27/13 at 40 rows, 35/17 at 52, 14/8 at 22 — all
    three verified on screen.
  - **Raw mode**, so every keystroke reaches the agent as typed. Ctrl-C belongs
    to Claude Code, which uses it to interrupt a turn.
- **The band MUST be anchored at row 1, and that is measured, not chosen.**
  Lines scrolled out of a scroll region reach the scrollback only when the
  region starts at the top: Terminal.app keeps every line for a region on rows
  1-10 and **none** for one on rows 5-14. So nothing can be painted above the
  agent, and the scape reads downward from it instead — sky strip, sea, beach.
  Verified live: 99 of 100 scrolled lines stayed reachable.
- **⚠ DECSTBM homes the cursor, and a parameterless `ESC[r` is still DECSTBM.**
  This cost an afternoon. The exit path placed the cursor below the band and
  then reset the region once more from a deferred call, which pulled the cursor
  back to row 1 — and a zsh prompt redraw opens with `ESC[J`, so from row 1 it
  erased everything the agent had drawn. Every exit blanked the screen. The rule
  now in tests: **save the cursor before touching the region, place it after the
  last time you touch it.**
- **`frames.go`** extracts the frame producer the live pane and the band now
  share, so the two compositions cannot drift.
- **Probe first, build second.** `notes/claude-terminal-emissions.md` is what
  Claude Code actually writes to a terminal, measured off `tmux pipe-pane` over
  two sessions including a real turn. Trust it; do not re-derive it.

### Shipped 2026-09-01, session 7

- **Two notification sounds** (`internal/notify`). 30% of the rubric is the
  waiting experience and the note says the nudge must beat a terminal bell; a
  scape in a side pane is not the pane being looked at, so sound is the only
  channel that reaches the user. Bright chime = the agent is BLOCKED on you;
  deep sonar note = it finished. Keyed off the BUBBLE, not the pose (a broken
  build outranks a question in the pose, so a pose-driven sound would go silent
  on the one event that needs answering), and edge-detected so Claude's
  60-second nag rings once. Silent when following nothing, `XSCAPES_SILENT`
  to mute, `xscapes notify` to audition.
- **`xscapes claude`** — the launcher. Bootstraps tmux, or joins the window
  it is already in, or falls back to a second Terminal via osascript. Agent
  keeps its own pane and TTY (exec, not wrap). `-print` is a dry run.
- **`-live -await`** — the scape used to bind once at startup, so launching both
  halves together left it in demo mode forever. It now keeps looking, and
  deliberately ignores the session pointer present at startup (binding to a
  stale pointer SUCCEEDS and shows a dead session beside a live agent).
- **Published, and named.** `xscapes` on the `donlucasx` account, public, MIT.
  The Go module path had to follow the repo URL (Go resolves modules by URL, so
  a mismatch breaks install for everyone), renamed across 37 files, which also
  makes the binary `xscapes`. Clone-and-build verified from the public repo.
  ⚠ NOT renamed at the time, deliberately: `ASCIISCAPES_*` env vars,
  `~/.config/asciiscapes/` (live state; a rename orphans an installed hook),
  the working directory, and these docs. **Superseded 2026-09-03 — session 12
  did the migration; see the top of this file.**
- **README + MIT LICENSE.** Leads with the protocol, then the encoding table.
  Every command was run before being written down; three first-draft claims
  were wrong and were fixed (event names are `sub_start`/`sub_end`; the process
  adapter is NOT built; `go install` cannot work with no remote).
- **⭐ THE BEACH FALLS AWAY TO BLACK** (`DefaultSandFade = 1.0`, locked). Newest
  line contrast 132→204 midday, 148→204 night, and equal at every hour. It
  exposed two things it did not cause: `drawSand` picked ink from the palette's
  NOMINAL sand while painting onto a darkened background, so a half fade at
  midday LOST contrast (now samples the painted background per row); and
  `internal/scape/sandfade_test.go` mirrored that rule instead of running it, so
  it passed throughout — deleted, replaced by `sand_ink_test.go` which renders a
  frame and reads the pixels back. Verified on 256: monotonic, all indices ≥16,
  and it ADDS depth (without it the lower half of a tall beach was one flat colour).
- **⭐ NEW DIRECTION: the agent goes INSIDE the scape** — *"the entire Claude
  experience should happen within the xscape, not next to it"*, and *"the taller
  the window the more sand below"*. Mocked with `-overlay`; 83.5% of a real
  Claude pane is blank. `Shore.SkyRows/SandRows` added so a taller window spends
  its extra rows on beach (4 lines of history at 24 rows, 9 at 43, 16 at 60).
  **NOT BUILT**: it needs xscapes to host Claude in a PTY and composite, i.e. a
  terminal emulator. See ▶ NEXT.
- **⭐ INSTALLED FOR REAL, and the whole chain works.** `xscapes install claude
  --apply` on 2026-09-01, with his say-so. His four existing hooks survived
  byte-for-byte (the VERCEL_TOKEN secret guard included), statusLine untouched,
  backup in `~/.config/xscapes/backups/`. Real events arrived within seconds
  on THREE live sessions at once, no restart needed. Proven in tmux: his own
  commands written into the sand, whitecaps on the sea, cat in the working pose.
  This retires "no hook has ever fired into it", which had been true all project.
- **The test suite was writing the user's live session pointer.** adapter_test
  feeds real payloads through `translate()`, which records the session as
  current — into `~/.config/xscapes/run/current` with no override. `TestMain`
  now points the main package at a temp dir.

## Earlier in session 7 (2026-08-31)

Session 7 opened by asking for the companion pick. He was not ready: *"pull up
the latest companion study"*, whiskers *"need to revise this one"*, toes
*"tbd, let me see with and without still"*, and he chose **build on** -- so the
session closed the oldest gap against a locked requirement while the pick
stays open.

### Shipped this session

- **done and needs_input are DISTINCT cues** (the brief locks this; both used
  to raise the same NeedsYou pose and identical balloon).
  - `companion.Done` is a fifth state: content `^ ^` eyes, full tail held
    high and STILL, slow breath. Position and shape, not rate -- it survives a
    screenshot. The reducer's DoneHold window resolves here now.
  - Two balloon shapes: the **ask** is the solid box, now drawn in a warm
    attention colour (`bubbleAskCol`, same family as the worried eyes); the
    **finish knock** is `DoneBubble` -- dotted bars, colon walls, cool
    `bubbleCol`. `reduce.State.BubbleAsk` says which; an open ask outranks a
    stale knock on both channels. Tests cover all of it.
- **Balloons are opaque now.** Transparent spaces let the sea write glyphs
  into the middle of the words ("Rate limiting:is=in.") -- the exact defect
  `sprite.go`'s own comment warns text about. Every balloon draw site fixed.
- **The mirrored balloon pointer finds the cat.** The `v` sat under the LEFT
  shoulder, aiming at whatever kitten was underneath, since the mirror landed.
  `companion.MirrorTail` moves it right.
- **The companion study PNG was a cut-off capture -- always had been.**
  2200px viewport against a ~3380px page: the five-coats row, the Terminal.app
  256 comparison and the every-state row were NEVER in the file Lucas was
  reviewing. Recaptured full height and trimmed; he has been sent the full
  version with a correction note. The study also gains a "done" state cell.

### PARKED — the companion study (2026-09-01)

*"let's save the progress on the character study and defer the decision for
later. Let's keep working and keep the characters as is for now."*

**Do not re-ask, do not default anything.** Everything below is built and
waiting for him. `NewCat()` still returns what shipped before session 6.

- **Coats in the running**: cream, slate, sage, mauve, charcoal. He said
  cream/sage/slate "stand out best". Terracotta and ginger are OUT, too close
  to Claude's own mark. (Told this session: slate is the safe pick while
  Terminal.app is the daily driver; charcoal is a truecolor bet.)
- **Settled**: the nose, the toe tips, and inner ears = **inner shadow**.
- **Whiskers: FOUR VARIANTS on the table, awaiting his pick** (`b192179`).
  He caught the bottom pair "floating in the air" -- cause: both pairs
  anchored to the NOSE ROW's span while the block under the muzzle is two
  cells narrower. Variants: **tucked** (bottom anchored to its own row) ·
  **double** / **double long** ('═', two parallel strokes in ONE cell, so
  both whiskers sit inside the nose block) · **current** (unchanged, floats).
  Test `TestAttachedVariantsReachFurOnTheirOwnRow` walks inward on one row
  and demands fur -- proven red on current, green on the other three.
  Earlier rounds: lines are locked (braille rejected); the top pair is a
  dash + half-dash tip; the study portraits were fixed to c.H-2-chh (they
  had sat the cat a row low, putting the waterline at nose height).
  Guide file: `~/Downloads/Screenshot-2026-08-31-at-2.52.07 PMsd.gif` (nbsp).
- **Ear shadows and toes: he LOVES them** -- but whether all 3 details ship
  together is a separate decision he wants AFTER the whisker lock.

## How to look at things

```
go build -o xscapes . && ./xscapes claude     # THE REAL THING: agent inside the scape
./xscapes inside sh -c "seq 1 100; sleep 30"  # host anything; proves the band holds
./xscapes claude -beside                      # the OLD side-by-side tmux layout
./xscapes -live                               # the scape alone, Ctrl-C to quit
./xscapes -faces  assets/frames/companion-study.html   # THE OPEN QUESTION
./xscapes -colors assets/frames/color-study.html       # 256 vs truecolor
./xscapes -day    assets/frames/day-cycle.html         # ALL 24 HOURS, TRUECOLOR vs 256, with a slider
./xscapes tune                                        # FOLD THE REAL SPOOLS THROUGH THE REDUCER
./xscapes tune -sweep                                 # ... and re-fold them for each candidate setting
./xscapes shades -only 2                              # judge the 256 smoothing at the real window size
./xscapes -todo 3/5 -tod 0.5 -working                 # the checklist constellation
./xscapes emit todo -n 3 -of 5                        # drive it in a live scape
go test ./internal/host/ -run Resize -v               # replay the host through 11 resize modes
./xscapes -wired  assets/frames/wired.html             # a turn through the REAL reducer
./xscapes -info                                        # profile, size, chroma
./xscapes -mockup assets/frames/composition-study.html   # left vs mirrored, every terminal shape
./xscapes install claude                          # prints a plan, writes nothing
./xscapes emit tool_start -tool Read -target x.go # drive the scene by hand
```
GIFs open directly. HTML demos need `python3 -m http.server` in `assets/frames/`.
Demo flags: `-wired -mockup -anim -compare -layout -context -day -busy -kittens -sheet -strip -html`.
`-mirror=false` gives the old left-anchored layout.

⚠ `-plain` is blind to the moon and the shoreline — both live in the background colour.

## ▶ NEXT

**Order for a fresh session, 14 days to the deadline (closes 2026-09-17):**

0. **⭐⭐ PUBLISH THE PAGE AND SUBMIT.** `site/COMMONS-PROMPT.md` has the steps; the page is
   `site/index.html`. His hands only. **2026-09-04: not yet, by his choice** — *"we need to work
   on some of the pending items and update it later."* Update the page with what ships, then
   publish. Deadline 09-17.

1. ~~Check the incident's aftermath.~~ Automation: **allowed for session 15 ONLY** (*"yes, only
   for this session"*), tty rule; off again after unless he says otherwise. Whether his session
   survived the stray `/exit` was not asked; ask if it matters.

2. ~~HIS LOOK, then the install: the cube-path gradients.~~ **INSTALLED 2026-09-04** on his
   *"Install it"*. What is left is his eyes on it in a live session; if an hour looks wrong the
   knobs are `RampHop`/`RampLambda` (env `XSCAPES_RAMP_HOP`/`_LAMBDA`), `XSCAPES_RAMP=0` is the
   old rounding for a side-by-side, and `go run ./notes/gradientaudit -dump <hours>` shows the
   tones. Rebuild BOTH binaries after a change (`~/.local/bin/xscapes` is a copy).

3. **⭐ THE MIRRORED TRANSCRIPT IS CORRUPT in a long session** (peer session, 09-04, 133x61:
   rows interleaved character by character, blocks duplicated, the input box four times) — this
   subsumes his report #6. Offline, the s13 traces replay CLEAN (`TestReplayTraceKept`), so it
   needs a trace of a session that shows it: `XSCAPES_TRACE=/tmp/scroll.bin xscapes claude`,
   live in it until it corrupts, then `KEPT_OUT=/tmp/kept.txt go test ./internal/host -run
   TestReplayTraceKept -v` and read the kept rows. Interleaved there ⇒ the MODEL diverges (find
   the sequence, like ESC ( B); clean there ⇒ the WRITE side (`MirrorBatch` row accounting in the
   real terminal) — then the frozen comparison against Terminal.app's readback, tty rule only,
   with his OK.

4. **His report #2** (scape lines broken after scrolling): *"Did not check"* (2026-09-04). He
   watches for it next time he lives in it; if it persists, reproduce with a real Claude session.

5. **Live in it another day** with the fixes installed (swimmers, balloon cap, ESC ( B).

3. **The right-edge strip** — a 1–2 column strip of stale scape down the far
   right, photographed twice. `TestTraceRightEdge` proves the renderer paints to
   the full width the host knows, so it is a stale WIDTH BELIEF, not a paint bug.
   Unexplained: his title bar read 122 while the host's last known was 120, and
   TIOCGWINSZ agrees with Terminal.app exactly on both screens. Most likely the
   host lagging a drag; not confirmed to settle.
4. **The worry trigger** (see "waiting on him" above). Biggest signal defect in
   the project and the cheapest real fix left.
5. **Dial `TauFall`.** `xscapes tune -sweep` answers it in a second now. 12s
   gives median 0.54 / p90 0.80 / 3.5% saturated; 20s gives 0.64 / 0.90 / 6.0%.
   40s pins the sea and should not ship.
6. **The 45-60s demo video**, Terminal.app. ⚠ Before recording: **only 4
   `needs_input` events exist in the entire 18,919-event record.** The cue the
   30%-weighted Waiting Experience is built on essentially never fires on its
   own, so the video either drives it deliberately or leads with `done` (119)
   and the worried pose (91 errors).
4. **Waves in the sea** -- his idea, and there is no encoding conflict: the
   swells already ARE the activity channel, so making them look like waves is a
   rendering upgrade, not a new variable. Two constraints: keep the SPEED
   constant (coverage/count, never rate) and watch occlusion of the swimming
   kittens (16 at once at the peak, water empty 57% of the time). Milestone 2,
   after the video. **He asked whether to lock in first; the answer was yes and
   it still is.**

### Older items, still true

1. ~~**Make the whole scape cube-exact.**~~ **SKY AND SEA DONE 2026-09-02
   (session 11)** -- see "Where we left off". What is left of it, and it is
   smaller than it sounds:
   - **The GLYPHS have never been checked this way.** Only backgrounds were
     measured. Glyph colours take the `GlyphBoost` chroma lift before
     quantisation, so they are a different problem with a different answer, and
     `Foam`, `Grain`, `Star` and `WetSand` are all still free-chosen RGB.
   - ~~`-day` renders true RGB~~ **FIXED**: it renders both profiles from one
     frame now. **`-wired`, `-mockup`, `-faces`, `-strip`, `-anim` and `-reel`
     still flatter every palette they show.** `HTMLFragmentAs(px,
     term.Profile256)` is the fix and it is one argument per call site.
   - The night stays monochrome and that is still the right answer; the cube has
     nothing dark and coloured except the pure-blue column, which was rendered
     and rejected. Dawn and dusk DID move off grey: between luma 40 and 80 there
     are violets and blues (`#5f00af`, `#005faf`) the earlier survey missed
     because it only looked below luma 40.
2. **Re-tune the reducer against a real recording.** ⭐ **THE INSTRUMENT NOW
   EXISTS AND IT FOUND SOMETHING: `xscapes tune`.** It folds every spool through
   the real reducer offline, samples the level once a second and prints the
   distribution; `-sweep` re-folds the lot for each candidate setting. `replay`
   was never the tool for this -- it plays a spool into a live scape at
   wall-clock speed, so checking one value meant watching an afternoon.

   **What it found, from 11 sessions and 18,919 events: the floors were spending
   the sea's whole dynamic range.** 77% of all working time sat in two bins
   between 0.30 and 0.50, because `TurnFloor` is 0.30 and `FlightFloor` 0.45 and
   the level was CLAMPED to them. A quiet turn and one that had just done ten
   things read 0.300 and 0.300 -- identical. Fixed: the floors LIFT the range
   now (`floor + (1-floor)*heat`), which keeps every promise they were added for
   and gives the range back. Bins go 18/24/21/17/9/5/5 instead of
   45/32/8/5/4/2/4; saturation 3.1% to 3.5%.

   **⭐ AND IT FOUND A SECOND ONE, IN THE COMPANION: THE CAT IS WORRIED 37% OF
   ACTIVE TIME.** Its five states divide a session 3.5 resting / 58 working /
   1.2 done / 0.1 needs-you / **37 worried**, so the alarm is on for more than a
   third of the record and two states carry everything. Worse, the episodes:
   **31 of them, median 15m27s, 90th 58m, longest 2h02m -- and 65% were raised
   by a SINGLE error.** `worried` is set by any Error/TestFail and cleared ONLY
   by the next `Prompt`. The reasoning for that clear rule is sound and written
   down ("hooks can tell us a command failed, never that the code is fixed"); it
   is the TRIGGER that is too loose. In auto mode one grep with no match sets
   the alarm for a quarter of an hour.
   ⚠ **HIS CALL, because the brief locks "something is broken -> the companion,
   persists until it clears".** The recommendation is to keep the clear rule and
   require corroboration to raise it: two errors inside a window, or one error
   the agent does not recover from within ~30s of successful tool events.

   **Still unverified, and now cheap to check**: `TauFall`, `Impulse`,
   `TurnFloor`, `FlightFloor` are vars now so `-sweep` can move them. The
   remaining data: **46 spool files, 16,465 events** as of 2026-09-02. Mix:
   6,889 `tool_start` / 6,793 `tool_end`, 1,894 `context`, 377 `sub_end`,
   133 `sub_start`, 129 `prompt`, 119 `done`, 91 `error`, and **only 4
   `needs_input` in the whole record** -- worth knowing before building a demo
   around the cue that is 30% of the rubric. `xscapes replay <file>` folds one. `TauFall=12s`, `TurnFloor=0.30`, `FlightFloor=0.45` are all unverified.
   ⚠ **Correction to a note that was here and was wrong**: it said `TailLen` is
   a hard 4 while the write band scales with height. The band only ever scales
   DOWN -- `WriteRows` starts at 4 and is clamped by `H/6` -- so the two agree
   at 4 and there is no mismatch. What IS true is the other way round: **a tall
   window gets more beach but still only four lines of history**, which is not
   what the s8 note ("4 lines at 24 rows, 9 at 43, 16 at 60") describes. That
   was about `SandRows`, not about how much gets written. Growing the tail with
   the window is unbuilt, and it is the cheapest way to make a big window feel
   like it is using its space.
3. **The 45-60s demo video.** Entries close 2026-09-17. Record on Terminal.app
   now, not a truecolor terminal -- that follows from the ruling.
4. ~~**Watch the kitten accounting.**~~ **CHECKED 2026-09-02 and it is FINE.**
   Across all 46 spools, 16,465 events: 133 `sub_start` against 377 `sub_end`,
   which looks alarming and is not. 127 of the 133 starts match an end by agent
   id; the 6 that do not are sessions that ended mid-subagent. The 250 unmatched
   ends delete a key that is not in the map, which is a no-op. `r.subs` is a map
   keyed by agent id, not a counter, so it cannot drift or go negative.

## How to look at things (additions)

```
./xscapes claude -print          # the launcher's plan, writes nothing
./xscapes claude                 # agent left, scape right, one command
./xscapes notify                 # hear both knocks
XSCAPES_SILENT=1 ./xscapes … # mute
```

## Open threads for Lucas

- **Is "ascii-agents" real?** A Rust TUI for Claude Code with weather and ambient
  effects, multi-floor layout for concurrent agents, surfaced 2026-09-01 through an
  aggregator page only, with no GitHub URL in any search result. If it exists it is
  the nearest competitor found in the whole survey. Five minutes to settle.

- ~~**Build the embedded terminal?**~~ **DECIDED 2026-09-01: the pass-through
  band, and it is built.** Not the full emulator. What that costs: the sea does
  not show through the agent's own blank space — its band is opaque. Upgrading
  to a real emulator later would buy that back (and would own the scrollback,
  lifting the row-1 constraint), but it was days of work against 16 days left
  and a parser bug corrupts Claude's UI rather than just the picture.
- **⏸ SCROLLBACK: build it, or live without it?** He said *"the scrollback is
  important"* and then *"should feel like an embedded experience"*, and those two
  together are only satisfiable by xscapes owning its own history. Options priced
  in session 13 and all but one ruled out by measurement:
  ~~main screen~~ (reflows on width, uncorrectable) · ~~tmux stacked panes~~
  (border seam, no styling hides it against an hourly-changing sky) ·
  `-beside` (works today, zero build, but side-by-side is what he rejected) ·
  a history dump to a pager (half a day, not live scrolling) ·
  **own the scrollback** (~1 day keyboard-only, the only one that satisfies both).
  See ▶ NEXT #1 for scope, estimate and risks.

- **Housekeeping from session 13** — throwaway probe binaries left in his home
  directory (`~/anchorprobe`, `~/contentprobe`, `~/sgrprobe`, `~/sizecmp`) and
  session traces in `/tmp` (`t.bin`, `fail.bin`, `repro*.bin`, plus `.log`
  sidecars). The traces contain session screen content. Offered to delete; he did
  not answer. Delete on request.

- **Double sound.** His own afplay beeps still fire on Notification / Stop /
  PermissionRequest alongside the xscapes knocks, so he hears both. The brief
  always said "replace the beep", but they are HIS hooks and still work with no
  scape running. Never remove them silently.

- **The companion pick** — coat + whisker revision + toes. Full study is in
  his hands now; waiting on his steer.
- ~~Name~~ **DECIDED 2026-09-01: `xscapes`**, with the repo. ~~The dir and docs still say asciiscapes~~ **MIGRATED 2026-09-03**: directory, docs, env vars, `~/.config/xscapes/`, and the hook marker. Closed.
- **Charcoal is a bet on the terminal.** It looks best in truecolor and worst in
  256, where it goes grey. Slate is the safe pick if Terminal.app stays the
  daily driver; charcoal wins if Ghostty or iTerm2 gets installed. **No
  truecolor terminal is installed on this machine today.**
- ~~**Stars for completed todos**~~ **BUILT 2026-09-02.** It was two halves that
  had never been joined: the hook classified TodoWrite as an *op* and stopped
  there, so `n`/`of` were never filled and the reducer's `Todo` case could only
  be reached by hand from `xscapes emit`. Now the hook emits a real `todo` event
  and the upper sky carries a constellation: `*` per finished item, `∘` per
  outstanding one, position fixed by index and seed, held at a visibility floor
  like the moon so the clock cannot switch an agent channel off.
  ⚠ **The payload shape is INFERRED, not measured** — `notes/claude-hooks-verified.md`
  says nothing about TodoWrite's `tool_input` because **TodoWrite has been called
  zero times across 13,682 recorded tool events and every transcript since the
  hook was installed.** `todoCounts` fails quiet on an unrecognised shape.
  Drive it by hand: `xscapes -todo 3/5`, or `xscapes emit todo -n 3 -of 5`.
- **`wired-turn.gif` is STALE** — recorded before the distinct cues; its finish
  beat shows the old identical balloon. Regenerate when the companion settles,
  not before (one GIF round, not two).
- **A wide pane leaves an empty middle.** At 200x50 the cat and the tail sit at
  opposite edges with a lot of nothing between. Not wrong, just unused.
- Swimmers: no perspective scaling, and 2 of 18 drop out when a lane is
  oversubscribed.
- `CLAUDE.md` Milestone 1 list is stale — the installer is done; the tmux
  launcher is the only piece of it left.
- ~~No git remote~~ **PUBLISHED 2026-09-01**: github.com/donlucasx/xscapes, public, MIT.
