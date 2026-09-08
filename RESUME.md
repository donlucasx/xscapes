# Resume — xscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/xscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. Sessions 22 and 23 are LANDED as eight commits,
HEAD `62f2664`, tree clean, all green and installed. ⚠ FIRST: `git log
origin/main..HEAD` — those eight are **UNPUSHED** and need his go-ahead. internal/companion is OURS now, the parallel session
released it. notes/claude-hooks-verified.md is the Claude Code hook schema — trust it, do not
re-derive it. Skim origin-chat.md only if you need the why; ignore ideas.md — it
is parked. Tell me where we left off, then pick up from ▶ NEXT — item 0 is a
question for me, not work for you.

Two things about how to work on this, learned the hard way in session 11:
build the instrument before trusting the picture, and check what the RENDERED
frame does rather than what the source says it should. And one from session 14:
do NOT drive Terminal.app (osascript, System Events) without asking me first.
```

## Brand workstream — where we left off (2026-09-05, session 17, a PARALLEL session; HEAD `9ba068e`, pushed)

> **Engineering note, 13:50 (claude-6e):** the renderer is settled for the deck — Lucas picked the
> **hue rim** for the disc and ruled the outstanding-todo ring OUT; `site/index.html` is re-rendered
> from that at `b13598a`. Rebuild the deck from it with `assets/deck/make-deck.py`; stage only
> `assets/deck/`. The brand session had exited before this could be sent to it.

**Copy-paste this prompt into a fresh BRAND session** (the engineering session owns the code, `notes/`,
this file's other sections and the `CLAUDE.md` banner; brand touches `site/`, `assets/`, README and appends
under its own heading in `_FEEDBACK.md`):

```
cd ~/Documents/claude/xscapes/ and read CLAUDE.md, then the "Brand workstream" section of RESUME.md,
then assets/brand/README.md. The identity is LOCKED (Cursor, with Reverse as the alternate colour);
the guidelines are LOCKED v1.0 (assets/brand/guidelines.html). The deck (assets/deck/) is a PARKED
DRAFT with my notes in _FEEDBACK.md session 17; do not edit it until I say so. Tell me where we
left off, then wait for my notes.
```

**Locked, in order, all on 2026-09-05 (verbatim quotes in `_FEEDBACK.md`, session 17):**
- *"i love Cursor ... lets keep/polish Cursor"* → **the identity is CURSOR**: a block cursor over the x,
  reverse video, one cell 0.6 x 1.2 em, the x on the terminal's baseline (955/1200), Geist Mono 700,
  greys 233/234/236/245/255 + the noon sea 111 `#87afff`; on light, dawn 24 `#005f87` as the accent.
- *"lock in the CURSOR option with REVERSE as an ALT color application"* → the same cell in gold 179
  `#d7af5f`, once per surface where the brand is the subject (136 in the cell on light, 94 as text).
- *"looks great, lock it in"* → **guidelines v1.0**: https://claude.ai/code/artifact/86fde655-b172-4fb5-9ca5-36e41a841eb4
  and `assets/brand/` (12 SVGs from the font's outlines, `make-brand.py`, `guidelines.html`, `README.md`).
- Four rules that hold everywhere: every colour an xterm-256 index 16–255 · every mark typeable
  (`printf '\e[7mx\e[0mscapes'`) · Claude's oranges out · plain words, the name lowercase always.

**Boards, for the record:** round one "Cursor, Waterline, Escape" `3042ea75-…` · round two "Cursor, Escape,
Reverse" `91c5853d-…` · guidelines `86fde655-…` · deck `6a42a717-…` (all this login's).

**The deck is a PARKED DRAFT** (`assets/deck/index.html` + `xscapes-deck.pdf`, 11 slides, three real frames
lifted from `site/index.html`). His notes so far: *"sun looks wrong and there are sub agents "swimming" on the
sand (also wrong) ... just save it as a draft for now (no edits needed, just notes)"*. Both are RENDERER facts
about the frames: the site has since been re-rendered (`9ba068e`; the kittens on the sand are now SITTERS by
design) and the sun disc changes ONCE MORE after his moon pick in the engineering session. More notes coming.

**▶ NEXT (brand):** 0. his further deck notes · 1. after the sun pick lands and he says go:
`go run . -site site` (if not already re-rendered) → `python3 assets/deck/make-deck.py` → print the PDF
(command in the script's docstring) → his review · 2. the Commons page restyle at its rebuild
(`site/template.html`: ground 233, gold lockup once at the top, blue 111 links, Geist; its `#e8c27a` is not
a cube entry and goes) · 3. README header via `<picture>` with the two lockup SVGs (snippet in
`assets/brand/README.md`) · 4. the CLI help header mark (two lines of reverse video in `cmd.go` `usage()`,
the engineering session's file: hand it over) · 5. GitHub avatar: export `avatar-gold.svg` to PNG at 512
and upload (his hands).

⚠ Housekeeping learned today: the Playwright MCP writes screenshots into the REPO ROOT and a `-A` commit
from the other session swept three in (`bff64d1`, since removed). Render previews with headless Chrome
`--screenshot=` straight into the scratchpad instead.

## Where we left off (2026-09-07 ~14:30, session 21 WRAPPED. HEAD `4047b24`, pushed, clean, INSTALLED, v0.3.0 tagged)

**Session 21 in one line:** the entry stopped being the thing that could kill this — a release tag, a
working `-h`, a page that no longer shows the bug he photographed, and the companion's alarm cut from
44.6% of active time to 16.0%; then he handed the crab to the parallel session and wrapped this one.

⚠ **`internal/companion` IS THE PARALLEL SESSION'S.** He is implementing the crab ("Hero") there and
will start a fresh session here afterwards. **Do not touch `internal/companion` in this line of work
until he says it has landed.** Its brief, his three rounds of notes and the pick are in `_FEEDBACK.md`
§"Session 21 (parallel)". The open question that sizes it is HIS and unmade: does Hero **replace** the
cat or **join** it as a choice? There is no companion interface anywhere and `NewCat()` is concrete at
**25 production call sites + 13 in tests**.

**SHIPPED TODAY, all pushed and installed:**
- **v0.3.0 TAGGED.** `go install github.com/donlucasx/xscapes@latest` was serving **v0.2.1, 115 commits
  behind** — and that exact command is on the README and in the live page's install block. Anyone who
  followed the entry installed a build from before almost everything.
- **The worry bar, his ruling "main-thread errors only":** `reduce.go` raises `Worried` only when
  `e.Agent == ""`. Folded through his real 330h: **worried 44.6% → 16.0%, working 49.6% → 78.1%**,
  needs-you and done unchanged. 590 of 675 errors fired inside a subagent. The error still drives the
  sea and still lands on the sand; only the FACE is quiet.
- **`xscapes -h` serves help.** It printed `flag needs an argument: -h` because `-h` is the height.
  `dispatch()` has had a `case "-h", "--help"` all along that could never run — it returns early on any
  `-` argument, so only the bare word `help` reached it.
- **The live page and clips rebuilt and republished.** They had been serving the star-on-the-disc
  defect he personally reported, for a day and a half. Verified live, byte-identical to the repo.
- **The Commons brief rescoped** (it still said "runs Claude Code inside a living shoreline"), the
  demo-video promise removed on his ruling, clip counts corrected 5/1.3MB → 15/5.4MB, and the Desktop
  paste kit regenerated from today's page.
- Earlier in the same session: the sky/sea ramp collapsed to one tone a cell on Terminal.app
  (`term.NoSplitCells`), the star-on-the-disc fix, and the crash recovery.

**HIS RULINGS TODAY, so nobody re-opens them:** the worried eye **stays amber** (red is darker than the
ground it sits on — WCAG 5.03 → 1.24 at night); the disc keeps its **flat caps**; the **shoreline stays
split**; **no demo video**; and of the four worry-bar families he took **A or D**, then the main-thread
variant of D.

## Where we left off (2026-09-06 ~22:30, session 20 — his two reports off the first live run. HEAD after the star fix, clean, installed)

**Session 20 in one line:** he launched a session, and inside a minute had two reports: a broken
column at the far right and a sun that is not pixel perfect. The right edge is Terminal.app's, not
ours, and nothing shipped for it; the sun turned out to be three separate things, and his ruling
shipped the one that is a bug.

⚠ **A measurement of ours was wrong and is now corrected everywhere:** the cell pitch is **exactly
14.000 device px**, column 0 at **x=132**, rows 30.0 px from **y=154**. The 14.28 figure divided the
whole content view by 143 and absorbed Terminal.app's 20 px inset on each side — and by that
arithmetic the bad strip lands INSIDE the grid, the opposite conclusion. Use 14/30 for every future
pixel reading of a Terminal.app shot at this profile. Also **REFUTED from 09-05**: the window IS
snapped to the cell grid (143 x 14 = 2002 exactly).

**The right edge — NOT OURS, ship nothing, and he is right that it is new.** The strip is columns
**144 and 145, in Terminal.app's right inset**, outside the 143 columns it reports. Our renderer
paints all 143 (a real render emits exactly 143 visible cells per row). The strip holds **our own
earlier frames**, retained — one frozen ~2 rows higher, the other 4–5, from when the window was 145
then 144 wide. Fourteen earlier screenshots have a clean inset, including two from 09-06 itself, so it
did appear tonight — during a resize, not a deploy. No fix is available: the periodic repaint only
rewrites columns 1–143, erase stops at the visible width, and a longer line would wrap and damage the
LEFT edge. Widening a couple of columns and back clears it until the next narrowing.
⚠ **Do not offer that as a diagnostic** — widening exposes those columns so the band repaints them,
and the strip vanishes for an unrelated reason.

**The sun — his ruling was "Just the star", and that shipped.** `stars()` plotted into the far layer
while `moon()` paints only backgrounds, so a dust glyph survived and composited onto the disc — in
53% of frames across a 6000-frame sweep. `Shore.DiscCovers` is now the single predicate for BOTH star
fields, and `discGeom()` fixes the disc's centre and radius BEFORE `stars()` runs (it did not, so a
naive guard would have read the previous frame's disc). ⚠ The predicate must sample BOTH half-rows —
whole rows miss nothing at his 143x27 and leak a row taller, which is why the test sweeps 4 widths x 6
heights x 5 seeds. Three mutations go red. At his frame exactly one cell changes.

**Still open on the sun, by his ruling, not by omission:**
- **The caps are pure rim** — a flat rose bar 5 cells wide top and bottom, zero lit pixels, because the
  rim band at `shore.go:693` is 0.55 rows while the sampler steps 0.5, so the outermost half-row can
  never hold lit body at ANY size. This is why the disc reads as a rounded square. He passed over "the
  star and the caps" and "star now, caps as a study"; the trade is a closed ring against a lit tip and
  it needs re-measuring at scape 24–28.
- **The hairlines** — five on the disc; the clearest is a rosy 1 px rule floating 12 px BELOW it in
  flat sky. Ink re-measured off his own screenshot: 0.965 at offset 17, **0.426 at offset 29**,
  confirming 17.0 → 29.4 of 30 a second time. **The row pitch has not moved, so his line-spacing test
  is still the open question.**

⚠ **A parallel session is live in this tree** (`notes/charstudy/`, a companion study served on
localhost:8777). **Stage by path.**

## Where we left off (2026-09-06, session 19 — his three reports; the MACHINE CRASHED mid-session, recovered 19:00. HEAD `8d84eeb`, pushed, clean, installed)

**Session 19 in one line:** he brought three defects out of long live sessions, all three were measured
off his screenshots, one ruling SHIPPED (the sun wanes as a crescent), one he PARKED on his own
terminal test (the thin lines), and the third — the scrollback corruption — was still being worked when
his Mac rebooted at **18:51:32**.

**Nothing was lost.** The tree was clean and `8d84eeb` was already committed at 18:44, seven minutes
before the crash; it is now pushed. The session transcript survived intact and everything that lived
only in it has been written down — his two rulings verbatim, the pre-costed menu he'd return to, the
line-spacing hypothesis and the two byte-level scrollback measurements, all in `_FEEDBACK.md`
§Session 19. What the reboot did destroy: `/tmp/apple.bin` (34 MB, 09-05) and eight other traces, and
two of his six pasted screenshots that were only ever in macOS temp dirs. All the numbers derived from
them are on disk; the corpus itself is gone.

**SHIPPED (`8d84eeb`, installed):** his ruling ***"Sun wanes as a crescent."*** `shore.go` had
`lit = 1 - ContextUsed` with the terminator a second disc sliding across at `2*rr*lit`, so a slate mass
grew on the SUN as context filled and cleared at a compaction — his *"the Sun seems to break sometimes,
and fix itself"*. Measured in his 134x71 frame at 49%: the lit body ends at x677, the slate ran on to
x719, **2.6 cells past it**. Now no unlit face is painted by day; at night the moon keeps its shaded
face (which is what stopped a moon one column into its phase reading as bitten). `SunShadow = "slate"`
keeps the old look for the study pages. Tests: `sun_crescent_test.go` + the changed `moon_quad_test.go`,
**both mutation-checked — 9 mutations, 9 RED.**

**PARKED BY HIS RULING — the thin lines.** He answered ***"Leave it, I'll try line spacing first"***:
ship nothing, **he tests Terminal.app's line spacing and reports back**. The block glyph is ~25.4px in
a 30px cell so the gap is leading — **Terminal → Settings → Profiles → Text → Font "Change…" → Line
Spacing** may kill the hairlines at no cost. Not verified; do not drive his terminal. If it fails, the
three costed alternatives are in `_FEEDBACK.md` — **do not re-derive them.**
⚠ **The 09-05 "U+2584 is bottom-exact" finding is REFUTED** (ink runs 17 → 29.4 of 30, half a logical
pixel out). ▄ took the hairline 5px → 1px, not to zero. Corrected in `internal/term/color.go`,
`CLAUDE.md` and the session 16 block below. `notes/lineprobe` draws a frame at the real geometry so the
hairline is visible in a browser — a page that paints a split cell as two flat halves cannot show it.

**OPEN, and the next work — the scrollback (c).** It is **not** a strikethrough: the marks sit in their
own cells and do not cross the letterforms. It is a **cell-level merge** — a dash glyph in the body
text's own colour (229,229,229) standing in cells the terminal showed as spaces, and past the text's end
the old content survives (`u63racode` = "ultracode" with "63" over two cells). New from the trace, before
it was destroyed: **SGR 9 never appears** (84 distinct SGR bodies, no strike) and **the agent emits no
cell-shifting sequences at all** (no ICH/DCH/IL/DL/ECH in 34 MB). So the merge is **ours or the
terminal's**, not the agent's. The session died designing the instrument for it, its own words:
*"whether the bad rows are bad **bytes** or bad **drawing** — so let me make the next occurrence answer
that by itself."* It was reading `host.go` 560–625 (replay/exit) and 320–360, `Close`/`keepFinalScreen`/
`replay()`, `event/paths.go` 1–45 and `mirror_test.go` `TestExitReplayCarriesTheTranscript`.
⚠ **Do not answer this with `XSCAPES_TRACE` alone** — it is opt-in and launch-time only, and the defect
only shows *after a longer session*, which is exactly why the third occurrence went untraced. The
proposal on the table is an **always-on mirror journal** written under `event.Home()` (`~/.config/xscapes`,
NOT `/tmp`, which this reboot proved volatile), recording each mirrored row so one grep answers the fork:
is the bad row already bad in `h.mirrored` (bad bytes, ours) or not (bad drawing, the terminal)?

## Where we left off (2026-09-06, session 18 WRAPPED ~03:10, HEAD `e07d82f`, pushed, clean)

**Session 18 in one line:** the submission page went online and was then rewritten three times on his
notes, ending with a hero clip that plays a whole session and a legend with the scene dissected;
his two rulings from the features test shipped; nothing about the product's behaviour is open.

**The page is LIVE and is the submission**: https://donlucasx.github.io/xscapes/ (`site/publish.sh`
force-pushes `site/index.html` + `anim/*` as an orphan `gh-pages` branch; **rerun it after every page
rebuild**). ⚠ The stored GitHub token has no `workflow` scope, so an Actions workflow cannot be pushed;
that is why it is a branch and not a workflow. ⚠ `.gitignore` excludes `site/anim/*.png` with a
NAME-SPECIFIC exception (`!site/anim/scape-*.png`) — rename a still and it silently stops being committed.

**Shipped to the product** (installed, `fd6a43c`): **`reduce.KittenDwell` 60 s**, so a subagent that
finishes in seconds still leaves a visible kitten — he had twice reported seeing none — and an **exit
queue rule** (`exitSpans`), because the dwell makes a fan-out leave at one instant and the old painter
drew all of them on top of each other. `CLAUDE.md` now says the companion is **12 columns by 7 rows**;
the "3–4 lines" figure was stale.

**Shipped to the page**, in the order he asked for it: the brand guidelines say ink, not gold · the copy
RESCOPED (any terminal agent, not Claude Code; a *scape* is any landscape and the shore is the FIRST;
"four states" replaced by "It runs on your clock") · the logline he picked from four
(*"Cozy ASCII scenes that react to your agent while it works"*) with the negative opener gone · the legend
promoted, reordered most-useful-first, and retitled **"How to read it"** · **"Choose your xscape"** with the
rainy window at dusk and the frog, the aquarium at night and the owl · ASCII rules on every section header ·
everything truecolor · **the hero clip is one whole session that loops**, 160 frames over 16 s: the sea, the
kittens arriving and swimming off, stars lighting, the moon sinking with the context, the companion asking
and finishing, then a **compact and a fresh list, which is what closes the loop**, with the day turning
underneath and the agent's transcript scrolling above · **nine dissected clips** beside the legend rows.

**How the clips are built** (`gifs.go`, `loopclips.go`, `site/make-gifs.py`): a clip carries its own
`cols`/`agentRows`/`crop`, runs its session on its own clock (`speed`) while the waves keep real time, can
ramp the day (`todEnd`), and is captured **across several pages** because one page of every frame outgrows a
headless screenshot. ⚠ **The moon and the constellation are washed out at midday by design**, so their clips
run at night — the first cut of both came back as empty frames.

**Commons is ready to submit and needs only him.** The rules were re-read signed in on 09-05
(`research/commons-submission.md`, last section): unchanged, and the entry picker takes **only a published
Commons build**, so the entry is this page served verbatim. The agent's brief is `site/commons-brief.md`;
the paste kit is on his Desktop at `~/Desktop/xscapes-commons/` (`brief.md`, `index.html` with absolute clip
URLs, `message.md` = both), regenerated whenever the page is. **Use Default · Quick, not Expert.**

**▶ NEXT**
0. ~~HIS CALL: the scrollback repro or the entry?~~ **ANSWERED 2026-09-08: the entry first, trace in
   the background. DONE — the entry is on the crab and LIVE**, page + 15 clips + deck + README,
   verified with real fetches. `7f68db6`.
0-now. ⏸ **HIS, in this order:**
   (0) **Close the two orphan windows** (ttys002 07:16, ttys005 07:21 — empty sessions), then restart
   THIS one by ID, not `--continue`:
   `XSCAPES_TRACE=1 xscapes claude -- --resume 2c247007-955d-479f-9266-7d5d4f8d6db1`
   ⚠ `--continue` resumes the most recent session in this directory, which is one of the empty ones.
   That restart is what finally delivers the companion live-refresh build.
   (1) **Regenerate the Commons paste kit FIRST** — all three files at `~/Desktop/xscapes-commons/`
   were built 09-07 from the CAT page. Then publish + submit. **9 days.**
   (2) The two cheap things below.
   (a) **Test Andale Mono** in Terminal &rarr; Settings &rarr; Profiles &rarr; Text &rarr; Font. Its
   block glyph fills **100.0%** of the line box against Menlo's **87.6%** (read from the font files,
   2026-09-08). If Terminal.app adds no leading of its own, that removes EVERY remaining hairline at
   zero code cost and lets `NoSplitCells` go off, giving back the split-cell gradient he traded away.
   It also decides whether the "collapse every edge cell" option is needed at all.
   (b) **Commons publish + submit** — see 0A. **Closes 09-17.**
0-crab. ~~WAIT ON THE CRAB~~ **DONE.** Hero shipped, is the DEFAULT, and `internal/companion` is ours.
0A. **HIS: publish + submit the Commons entry.** Everything is staged: `~/Desktop/xscapes-commons/`
   (brief.md · index.html with 16 absolute clip URLs · message.md), regenerated 09-07 14:02 from the
   live page. Fresh chat in the Commons Space, **Default · Quick, not Expert** → paste → publish the
   app public from its manage page → hackathons → Open hackathon → Your entry → Submit → screenshot.
   ⚠ **The publish step is the blocker, not the submit click** — the entry picker only accepts a
   PUBLISHED Commons build. **Closes 2026-09-17.**
0B. **His three picks for me.** ~~the right-edge DL fix~~ **SHIPPED `65e242e`** · ~~README + deck
   rescope~~ **DONE `7f68db6`** (the deck's title slide and the README; both said "Claude Code runs
   inside a shoreline"). Still unstarted: the **records + dead code + lying
   instrument** sweep (`tune`'s WORRY EPISODES block is hard-coded to the old rule and prints
   identical output under any change; `test_fail`/`test_pass` have no producer; `-ascii` ships
   Unicode) · the **README + deck rescope** (both still say "Claude Code runs inside a shoreline").
0a. ~~ASK HIM: did lowering Terminal.app's line spacing kill the hairlines?~~ **ANSWERED 09-07: "no,
   it didnt."** The hairline cannot be removed inside half blocks, only not drawn. ⏸ One untested
   lever left: the FONT. Menlo's block glyphs fill **85.4%** of the line box, Andale Mono **97.8%** —
   if that holds live it removes the hairline AND gives the gradient back, better than the trade he
   took. `Terminal → Settings → Profiles → Text → Font`.
0b. **The scrollback** — the one open product defect. ⚠ **The s24 mechanism is REFUTED (see the
   session 25 banner): the agent blanks all 33 of its rows before repainting, and the 300 KB
   "occurrence" cut is ~2.6 seconds.** ⭐ **Next move is `retainWidth`**: `Rules.RetainsWidth` is
   measured and honoured by `reallocBand`, but `host.go:272` builds the screen model with
   `newScreen(cols, rows)` and never sets it, so the model that feeds the SCROLLBACK MIRROR runs
   without Terminal.app's width rule. Wire it, replay, see whether the mirrored rows change. NOT the
   claimed cause. ⇒ **Then aim the next capture at the MIRROR moment, not the resize.**
   The older, still-true half: (i) The "it needs a long-lived window" claim is REFUTED: the corrupted
   window's scape started 20:28:20 and his screenshot is 21:04:44 — **36 minutes**. (ii) The RESIZE is
   on the record for the first time: the same window is titled **107x51** at 21:04 and **119x51** at
   21:18, and the corruption came back after. So the trace is NOT blocked on "you cannot trace a
   window that has already lived" — **launch traced, resize, work normally, and it fires inside the
   hour**:
   ```
   XSCAPES_TRACE=1 xscapes claude -- --continue
   ```
   ⚠ the `--` is REQUIRED · ⚠ ~7 MB/min, so kill it once it fires · ⚠ NOT `/tmp` (nine traces died in
   the 09-06 reboot; `=1` now picks `~/.config/xscapes/traces/` itself).
   **When it fires: STOP.** Do not scroll, resize or exit — read the window back by tty with his OK
   (`1049l` discards most mirrored rows) and pair it with the trace bytes at that offset. That is the
   first time we would have both halves.
   **Offline track, needs nobody**: drive an Ink-style differential redraw through a resize in
   `internal/host/screen_test.go` — draw a full-width rule, resize, repaint only the changed segments,
   and see whether the rows merge in the model. Never been done, and it does not depend on catching it
   live. ⚠ Run it at `AltScreen: true` or it proves nothing (the s13 lesson).
1. **HIS LOOK at the live page**, https://donlucasx.github.io/xscapes/ — rebuilt and republished
   2026-09-07 14:00 from HEAD, so what is up there is finally current.
2. **Commons: submit** — see 0A above, which supersedes this with the current paste kit and the
   publish-first warning
   "Entry submitted." **Closes 09-17.**
2. **His 45–60 s Terminal.app recording** of a real turn, no narration. The one deliverable nothing else
   can substitute for.
3. **HIS PICKS from the study** "Five Scenes, Five Companions"
   https://claude.ai/code/artifact/c9d41f7c-1b09-4a79-a262-32b9696bd3b9 — which scene, if any, before the
   deadline (my read: none), and which companion to draw properly. Two are already on the page as roadmap.
4. ⏸ **The DECK (`assets/deck/deck.tpl.html`) and the README still carry the old narrow scope** — the deck
   repeats *"Claude Code runs inside a shoreline"* verbatim. Both need the same rescope, his go-ahead.
5. The scrollback corruption — **now item 0b above, and seen a SIXTH and SEVENTH time on 09-07 night.**
   Still untraced. ⚠ **The line that used to sit here — "the defect only appears after a long session" —
   is REFUTED: 36 minutes was enough.** What it wants is RESIZES, not hours. Capture outside `/tmp` now:
   `XSCAPES_TRACE=~/.config/xscapes/t.bin xscapes claude` (the `--` separator is only needed when
   passing the agent flags, e.g. `xscapes claude -- --continue`). Then read the window back BY TTY with
   his OK (`1049l` discards most mirrored rows, so read BEFORE any restart), and offline
   `XSCAPES_TRACE=… KEPT_OUT=/tmp/kept.txt go test ./internal/host -run TestReplayTraceKept -v`:
   interleaved in the model ⇒ the model diverges, clean ⇒ the write side.
6. Still open from session 16, all needing him: the 6-column patch above the band (a scripted
   width+height drag, Terminal automation, his OK) · the two 2 AM stale rows · the transcript-start
   garble diff · the worry trigger's bar (the alarm was measured on 37% of active time) · the working-
   session items still unseen live: bubbles + sound, the mirror during work and its exit replay.
7. Small and mine whenever he wants them: the README header via `<picture>` with the two lockup SVGs ·
   the CLI help header mark in `cmd.go` `usage()` · the GitHub avatar PNG at 512 (his upload).

## Where we left off (2026-09-05, session 16 WRAPPED 14:40, HEAD `2f28667`+wrap, INSTALLED, pushed, clean)

**Session 16 in one line:** the live tests passed; shipped and installed: kitten swim-off · the
readout from 40% · ▄ split cells on Terminal.app · the disc-tip fix · the hue-rim disc · no todo
ring · swimmers never share columns; the site and the deck rebuilt on the brand with truecolor
clips, a cover and a one-page pitch. **Open threads, all his:** (1) LOOK at `site/index.html` and
`assets/deck/index.html`; one judgment call: the cover's headline crosses the ASCII sea on both ·
(2) publish + SUBMIT on Commons (the clips are files beside the page, `site/COMMONS-PROMPT.md`;
closes 09-17) · (3) his 45–60 s Terminal.app recording · (4) a TRACED working session for the
scrollback corruption seen in his live buffer · (5) the 6-column patch above the band and the 2 AM
stale rows, both open · (6) DONE, session 18 (2026-09-05 15:50): `assets/brand/guidelines.html` and the brand README
say ink on the page and the deck, his b&w ruling.

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
REDUCED — split cells are drawn as ▄ with the colours the other way up on Terminal.app
(`term.LowerHalf`, `DetectSplit`, `XSCAPES_SPLIT`). ⚠ **This paragraph said "FIXED … the hairlines
are gone" and that was wrong.** Re-measured 2026-09-06 (session 19): U+2584's ink runs 17 → **29.4**
of a 30px row, not to 30, so the ▄ swap took the hairline from 5px to **1px and no further**. It is
still there on every two-colour cell — 5 across his sky, 4 inside the disc. See the session 19
block at the top of this file. Defect (2) is EXPLAINED and documented, not fixable:
Terminal.app's alt screen RETAINS rows at their widest on a width change, erase reaches the
visible width only, and his 123.6-column window drew a partial 124th column of retained cells
(`notes/width-audit.md`, `notes/widthprobe`, model switch `screen.retainWidth`,
`TRACE_RETAIN=1`). The six-column patch above the band is NOT reproduced under either rule —
open. A read-back of his live window found the s14 #2 scrollback CORRUPTION in the main buffer
(`✻ Tomtotal reported 1259,cfetched)1259,sunique 1259`, rows merged where one had spaces) —
the mirror wrote what the model held; that session was untraced.

**Afternoon, 13:00–14:05:** the context readout SHIPPED at 40% used (his ruling; it had never been in
the live scene) · the outstanding-todo ring REMOVED (his ruling) · the disc's edge LOCKED as the hue
rim (his pick from "The Moon, Four Ways"; quad edge, shadowless sun and night halo stay as study
switches) · the SITE rebuilt on the brand with five ANIMATED CLIPS from the real reducer
(`xscapes -gifs`, `site/make-gifs.py`, `site/anim/*.gif`) and the DECK re-cut on them
(`assets/deck/make-deck.py`, PDF re-printed). The brand session (s17) locked the identity and
wrapped; its files ride in `assets/brand/` and `assets/deck/`.

**14:05–14:30, his six notes on the first pass → second pass SHIPPED (`f4fc084`):** swimmers never
share columns (real defect, slotted per lane before) · the clips in TRUECOLOR at a 14px cell, boxes
cut to the art · the page opens on a COVER (the b&w mark centered over six glyph-only frames of the
real night sea, CSS-stepped) then ONE PAGE (synopsis, four reasons to care, the hero clip full width,
install) · the deck has the same cover and one page, no gold. Pipeline: `go run . -site site &&
python3 site/make-gifs.py && python3 assets/deck/make-deck.py`, then the Chrome PDF print in
`make-deck.py`'s docstring.

**▶ NEXT (as session 16 left it, now superseded by the session 18 block above):** his look at the
page and the deck · the Commons sequence · a TRACED working session in Terminal.app
(`XSCAPES_TRACE=/tmp/apple.bin xscapes claude`) for the scrollback corruption · the 6-column patch
above the band · the working-session items still unseen: bubbles + sound, kittens live, the mirror
during work and its exit replay, the transcript start.


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

**⚠ 0. COMMIT SESSION 22.** Everything below shipped and is installed at `~/.local/bin/xscapes`, and
NONE of it is committed. Full suite green, vet clean. Five separate stories, and they should land as
five commits, not one blob:
   - the crab companion (`internal/companion/crab.go`, `crab_kittens.go`, + `kind` field on `Cat`)
   - the companion switcher (`companion_pref.go`, `cmd.go`, `frames.go`, `.claude/commands/companion.md`)
   - the disc-hairline fix (`internal/canvas/canvas.go` resolver gate + `internal/scape/shore.go`)
   - the right-edge DL fix (`internal/host/band_control.go`, `host.go`, `RetainsWidth` rule)
   - the trace fixes (`internal/host/host.go` openTrace)
   `notes/charstudy/` is 15 MB of study pages and generators — decide whether it is committed or
   gitignored; it is the record of how the crab was designed.

**1. ⭐ THE SCROLLBACK DEFECT, and it is closer than it has ever been.** The fork is SETTLED (bad
bytes, not bad drawing — see `_FEEDBACK.md`). The richest sample yet is on disk:
`~/Documents/Screenshots/corrupt-hist-515.txt`, **210 corrupted rows**, two distinct failures (122
rows ending in a run of `─`; 157 interleaved mid-text, two rows sharing one row cell by cell).
⚠ **What is still missing is a trace of the SAME window.** Traces are now easy
(`XSCAPES_TRACE=1 xscapes claude -- --resume <id>`, prints its path, never fails silently), but the
corruption earns itself through LONG LIFE AND REPEATED RESIZES — the 22-hour window had 210 rows, the
fresh one had none. **Trace the long-lived window, and resize it.** A trace grows ~7 MB/min; do not
leave one running.

**2. The disc's OUTLINE still carries a 1px rule** at its top and bottom edge cells — they must split
to stay round. Proposed, not built: pick the half-block ORIENTATION per cell (`▀` where the lower
colour continues below, `▄` where the upper continues above) so every gap is painted in the colour it
sits against. Same shape, no rule. Build it behind a switch so he can compare.

**3. ⚠ THE ENTRY SHOWS THE WRONG COMPANION.** He ruled the crab is the DEFAULT, knowing the live page,
its five clips and the deck all show the cat. Commons closes **09-17**. Rebuild them or the entry's
own pictures do not match the binary a judge installs.

**4. Pacing** — his design, built and studied but NOT shipped (`notes/charstudy/13-hero-paces.html`):
one step per MAIN-THREAD tool event, so it is a count and a position, never a rate. Needs
`reduce.State.Steps` (~3 lines beside the existing `Events`) and costs the activity tail ~6 columns.

**5. Crab gaps before it is finished**: `DrawKittens` still has ONE crablet size (the cat's ladder with
its 6/10-up, 4/8-down hysteresis is not ported) · the balloon pointer aims where the cat's shoulder is
· `SetEyeFill`'s three options mean something different on stalked eyes.

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
