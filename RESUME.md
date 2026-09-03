# Resume — xscapes (directory still `asciiscapes/`)

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. The project is named xscapes now; the directory
and the ASCIISCAPES_* env vars still say asciiscapes on purpose.
notes/claude-hooks-verified.md is the Claude Code hook schema — trust it, do not
re-derive it. Skim origin-chat.md only if you need the why; ignore ideas.md — it
is parked. Tell me where we left off, then pick up from ▶ NEXT — starting with
the decision at the top of it, which is mine to make, not yours.
```

## Where we left off (2026-09-02, session 11, `main`, NOT yet pushed)

**⭐ SECOND PASS: the 256 sky was not just banded, it was the WRONG HUE.** He
looked at the day page and said the gradients did not read as smooth. Chasing
that found something better than smoothing: **independent per-channel rounding
was turning blues into violets.** rgb(48,112,170) is a mid blue with green
sixty-four clear of red; the quantiser rounds red UP to 95 and green DOWN to 95
and hands back rgb(95,95,175), a lavender. Six rows of sky a day, worst around
five in the afternoon. `Index256Keeping` now keeps any channel ordering the
source states clearly, for BACKGROUNDS only -- glyphs keep the old path, since
the chroma boost has already moved them on purpose. It costs about 3% of
distance and buys the hue.

**Three smoothing ideas, one kept.** Splitting a cell with U+2580 so a band edge
can fall mid-cell: kept, 5% closer to the true ramp, no texture. Shade blocks
(U+2591/2/3) between two cube colours: measured a real gain -- 11 tones to 14 --
and rejected on sight, the dot pattern reads as stipple before it reads as tone.
Stacking the two bracketing colours as a 1x2 dither: rejected, worse on
measurement AND the averaging assumption fails, since half a cell is three or
four pixels and the eye resolves it. `ASCIISCAPES_SHADE_BLOCKS=1` renders the
rejected one in a real terminal if he wants to overrule me.

**Endpoints re-picked by search, then re-picked again.** A band appears wherever
the ramp crosses a quantisation boundary in ANY channel, so the smoothest
gradient is the one whose channels travel furthest -- `#005fd7` to `#afd7ff`
looks like the better noon sky and renders worse than `#005faf` to `#afd7ff`.
⚠ The first search ranked on band count alone and picked a horizon that put a
lavender stripe across the sky: **ten bands, one of them wrong.**
`TestTheSkyNeverGoesViolet` exists because of that, and it reads the RENDERED
frame -- the first version recomputed the ramp and was blind to the fix.

**Sea endpoints are held by a second constraint the search cannot see**: a longer
ramp repaints more cells when the mean waterline moves, and the open sea holding
still is his. Of the pairs giving ten bands, `#005f5f` to `#87afff` is the only
one under the 8% churn `TestTheBackdropHoldsStillBetweenFrames` allows.

**Dusk is no longer violet** -- not on taste, on measurement. A deep-blue-to-violet
zenith leg passes through rgb(48,48,175), which has plenty of chroma and no home
in the cube, and lands on grey 58 at half past four.

**⭐ THE SKY AND THE SEA ARE CHOSEN FROM THE 256 CUBE NOW** -- ▶ NEXT #1, done.
The note this session started from was half right and the half it got wrong was
the important half. The complaint on record was that the sea's depth gradient
collapsed onto one teal. It does, but that is a detail: **measured across 48
half-hours of the day, `SeaFar` landed on the GREYSCALE ramp at 40 of them and
`SkyTop` at 36.** For most of a working day the two biggest regions on screen
had no colour at all on the terminal he uses. Across the 25 daylight half-hours
the count is now **zero** for all four background colours.

It went unseen because **every HTML study in this repo renders true RGB**,
`-day` included, so the previews have always shown a blue sea that his terminal
never painted. **`./xscapes -day <file>.html` is now the honest one**: every hour
rendered TWICE from the same frame, truecolor beside 256, with a slider, a play
button and a show-every-hour grid. `assets/frames/cube-study.png` is the
before/after against the old palette, and it cannot be regenerated -- it needed
two binaries.

Two regression tests now hold it, both proven RED against the old palette:
`TestSkyAndSeaKeepTheirColourThroughTheWorkingDay` (13 daylight samples grey on
the old palette, 0 on this one) and `TestTheSeaShowsItsDepthOn256` (11 of 18 sea
rows were one colour at noon; the cap is half).

**⭐ THE RESIZE BUG IS FOUND, AND IT WAS THE TERMINAL MOVING THINGS.**
He photographed a fragment of sky above the band and a strip down the right
after resizing. **Terminal.app keeps the BOTTOM of the screen when a window
shrinks**, so every row slides up by the difference -- and the scape is painted
at the bottom, so its rows slide straight into the agent's band. Neither side
cleans them up: the host cleared only the rows that changed hands, assuming
nothing moved, and Claude Code emits nothing at all on a resize. The clear now
starts where the old scape's first row LANDS rather than where it was.

**`internal/host/screen_test.go` is a small terminal**, and it is what found
this. Reading the escape sequences tells you what was SENT; only replaying them
tells you what is on screen. It models CUP/EL/ED/DECSTBM/DECOM/save-restore,
autowrap and region scrolling, plus `resizeScrolling` for what Terminal.app does
on a shrink. `TestEveryBandRowIsRepaintedAfterAResize` drives a REAL Host through
nine resizes and is proven red on seven rows without the fix.
⚠ Two instrument bugs to know about if you extend it: it must buffer a partial
escape sequence across reads, and the snapshot must be taken BEFORE `Run()`
returns, because the close path legitimately blanks the band.

**⚠ THREE DEFECTS HE FOUND BY RUNNING IT, ALL NOW FIXED.**

0. **The moon became a blob** -- my own regression from the fix below. Making
   the disc solid, I left the cutoff at `rr+rim`, where `rim` had been the width
   of a FADE from `rr-rim` outward. Every cell that used to be falling away got
   painted at full strength and the disc came out nearly twice its radius. It
   ends at `rr` now.
1. **The sun had a grey fringe.** Its rim faded into the sky, and the alpha
   where its red and green cross -- 0.667 at noon -- makes a colour of chroma 20
   to 40, which is where the greyscale ramp wins. Measured on his frame:
   `rgb(193,188,151)` painted as grey 188. **The disc is solid now**; softness is
   not available on this palette, so a clean edge is the honest version.
2. **A cyan stripe across the sky.** The cube's first step is 95 wide and the
   rest are 40, so red, starting at 0, always crosses its levels later than
   green does -- and for three rows the sky went cyan and came back. Fixed by
   weighing HUE as well as distance in the background quantiser
   (`hueWeight`, 8, `ASCIISCAPES_HUE_WEIGHT` to try another). The alternative was
   starting red at 95, which removes the wobble and more than half the bands
   with it: 9 down to 4.
3. **Banding, which is NOT fixed and may not be fixable.** ↓

**⚠ THE OPEN ONE: BANDING IS MUCH WORSE IN A REAL WINDOW THAN IN ANY STUDY.**
He ran it at **152x57** and the sky came out in four or five hard blocks with a
grey stripe through the middle. Reproduced and measured: the number of cube
colours a ramp crosses barely changes with height -- 8 at 68x22, 9 at 152x57 --
while the sky grows from 10 rows to 24. **So the bands get two and a half times
fatter and nothing about the palette fixes it.** Every study in this repo, and
the first version of `xscapes shades`, asked the question at the size where the
problem is mildest.

Two things follow. `shades` now defaults to the real window and takes
`-only N` to show one variant at full height. And **the shade-block rejection
should be re-judged**: he called it "too busy" at 64x14, which is the size where
banding is least and stipple is most obvious, and that is the opposite of the
size that matters. A curved sky ramp was tried and is NOT the answer -- it trades
grey rows for longer flat runs, 5 to 12 at his size.

**⚠ TWO MORE FOR HIS EYE, both one constant away from being changed.**

1. ~~Dusk reads magenta.~~ **GONE**, and it had to go for a reason better than
   taste -- see above. Both twilights are deep blue overhead now with the hour
   in the horizon.
2. **Daylight is brighter and the sea is more turquoise.** That one is not a style choice, it is the cube's
   shape -- below luma 60 it holds no blue except the electric `#0000xx` column
   that `Shore.BlueSky` already documents as rejected, so a sea that keeps its
   blue has to sit above luma 70. If it reads too tropical the dial is
   `seaDeep`/`seaShallow`, and the constraint is that the pair stays at least TWO
   cube steps apart in some channel or the ramp shows two bands and nothing
   between.

Every colour named above is in `internal/scape/palette.go`, in one `var` block
at the top, each with its xterm index in a trailing comment.

## Earlier (2026-09-02, session 10, last code change `0e7f265`, `main`, pushed)

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
  60-second nag rings once. Silent when following nothing, `ASCIISCAPES_SILENT`
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
  ⚠ NOT renamed, deliberately: `ASCIISCAPES_*` env vars and
  `~/.config/asciiscapes/` (live state; a rename orphans an installed hook),
  the working directory, and these docs. That migration is his call.
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
  backup in `~/.config/asciiscapes/backups/`. Real events arrived within seconds
  on THREE live sessions at once, no restart needed. Proven in tmux: his own
  commands written into the sand, whitecaps on the sea, cat in the working pose.
  This retires "no hook has ever fired into it", which had been true all project.
- **The test suite was writing the user's live session pointer.** adapter_test
  feeds real payloads through `translate()`, which records the session as
  current — into `~/.config/asciiscapes/run/current` with no override. `TestMain`
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
ASCIISCAPES_SILENT=1 ./xscapes … # mute
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
- **Double sound.** His own afplay beeps still fire on Notification / Stop /
  PermissionRequest alongside the xscapes knocks, so he hears both. The brief
  always said "replace the beep", but they are HIS hooks and still work with no
  scape running. Never remove them silently.

- **The companion pick** — coat + whisker revision + toes. Full study is in
  his hands now; waiting on his steer.
- ~~Name~~ **DECIDED 2026-09-01: `xscapes`**, with the repo. The dir and these docs still say asciiscapes; renaming those (plus `ASCIISCAPES_*` and `~/.config/asciiscapes/`) is a migration he has not asked for.
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
