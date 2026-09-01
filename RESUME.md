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

## Where we left off (2026-09-01, session 9, last code change `d70ff65`, 92 commits, `main`, pushed)

**Live: https://github.com/donlucasx/xscapes** (public, MIT). Milestone 1 is
COMPLETE, the hooks are installed and firing, and **the agent now runs INSIDE
the scape**.

**⭐ `xscapes claude` NOW RUNS THE AGENT INSIDE THE SCAPE.** His ruling — "the entire Claude experience should
happen within the xscape, not next to it" — is built. One window: Claude Code in
a band pinned to the top rows, the scape painting the rows below it, no tmux.
He chose the pass-through band over a full terminal emulator after the probe
came back clean (below), then ruled that `claude` should mean this: *"1. it
should"*. The old tmux launcher is untouched and still reachable as
`xscapes claude -beside`; `xscapes inside <cmd>` hosts anything.

**It is not a terminal emulator, on purpose.** The agent runs on a pty sized to
its band, held there by DECSTBM, and its bytes reach the terminal untouched
except for one three-byte sequence. Nothing xscapes fails to understand about
Claude Code's output can corrupt it — the failure a real emulator invites.

The companion study is still PARKED at his request — see below.

### Shipped 2026-09-01, session 9 (research only, no code)

- **`research/prior-art.md` — has anyone already built this?** A survey, not a
  code read. Trust it; do not re-run it unless something looks stale. Answer:
  the space is crowded but in four separate categories, and xscapes sits
  cleanly in none of them.
  - **Idle screensavers** (cmatrix, pipes.sh, asciiquarium, termsaver,
    ascsaver, cacademo). Own the whole screen, agent-blind. `ascsaver` is the
    only one with any trigger at all, and it is a raw no-I/O check.
    asciiquarium is the ONE real *place* in the whole survey, and it is twenty
    years old.
  - **Per-command spinners** (terminal-animations/tan, the ora lineage). One
    line, two states.
  - **Fake-activity generators** (genact, hollywood). Useful only as evidence
    that people will watch a terminal do nothing if it looks good.
  - **⭐ Agent-aware indicators — the real neighbourhood, and it is NEW.**
    Barely existed two years ago. `pi-animations` (MIT, ~26 stars, 26
    animations for the pi agent, hooked to thinking/working/tool, inline
    single-line plus 3-to-5-row widgets above the editor) is closest in
    spirit, and its own docs call them pure loading indicators, not scenes.
    `claude-code-mascot-statusline` (MIT, 23 stars) is closest on reacting to
    real state: a 16-cell half-block sprite across 9 hook-driven states, but
    event-driven rather than animated. Plus `tweakcc` (patches Claude Code's
    bundle) and a whole `terminal-pet` GitHub topic (`buddy` 102 stars,
    `buddymon` 44, `tokengotchi`, `desk-waifu`, `codex-pets`).
  - **Demand signal, all open, no maintainer reply**: claude-code #66284
    (customizable ASCII working animation, `area:tui`), #29200 (thinking
    words), #35249 (statusline mascot), opencode #24937 (TUI pet). Note the
    shape of #66284: if Anthropic ships its `command` variant, a
    statusline-style hook for the working animation, that is a DISTRIBUTION
    CHANNEL for a scape strip, not a threat to `xscapes inside`.
- **The four gaps, and they are the pitch.** (1) Nobody runs the agent inside
  the scene — every agent-aware project is a strip beside or above it, and the
  screensavers only run when the agent is absent. The pty band has no analog.
  (2) Nobody encodes the amount of work; everyone keys off a 3-to-9-value
  state enum, so "the water is the work" has no competitor. (3) Everything is
  a pet, not a place. (4) Everything is host-locked TypeScript; one Go binary
  with an adapter protocol is unusual here.
- **⚠ The risk it surfaced**: the terminal-pet topic is crowded and growing,
  and a skimming reader files xscapes there on sight. The first screenshot has
  to say "instrument in a place", not "companion".
- **One lead not chased**: an "ascii-agents" Rust TUI with weather and ambient
  effects surfaced only through an aggregator page, no GitHub URL in any
  result. If it is real it is the nearest competitor found.

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

1. **Run a day on `xscapes inside` and re-tune.** It has been proven in a real
   terminal but not lived in. The reducer's `TauFall=12s`, `TurnFloor=0.30`,
   `FlightFloor=0.45` have still never been tuned against his actual rhythm, and
   `~/.config/asciiscapes/run/*.jsonl` is a real recording that `xscapes replay`
   folds. ⚠ `reduce.TailLen` is still a hard 4 — it should be a function of the
   beach's rows now that the beach's height varies with the window.
2. **The 45-60s demo video**, now of the thing running inside one window.
   Record in a truecolor terminal, NOT Terminal.app. Entries close 2026-09-17.
4. **Run a day on it and re-tune.** The hooks are INSTALLED and firing (see
   below); what is missing is a day of real rhythm to tune `TauFall=12s`,
   `TurnFloor=0.30`, `FlightFloor=0.45` against. `xscapes replay` exists for
   exactly this, and the spool files in `~/.config/asciiscapes/run/*.jsonl` are
   already a real recording.
3. **Decide the double sound.** His own afplay beep hooks still fire alongside
   the xscapes knocks on Notification / Stop / PermissionRequest. The brief said
   "replace the beep"; his hooks are his, so this is his call.

3. **Add the statusline chain by hand** or the moon stays dark. Install prints
   the exact line rather than taking over the statusLine Lucas wrote.
4. **Re-tune the reducer against a recording.** `xscapes replay` exists for
   exactly this. Are `TauFall=12s`, `TurnFloor=0.30`, `FlightFloor=0.45` right
   against his actual rhythm? Do real fan-outs reach the kitten ladder's numbers?
5. **Keypress focuses the agent pane** — the last half of Milestone 1 #6. The
   live loop reads no keys at all today; it needs raw mode plus
   `tmux select-pane -t <agent>`. The launcher knows the pane id and could
   pass it.

6. **Re-open the README's first 10 lines against `research/prior-art.md`.** Cheap,
   and it moves the two heaviest rubric weights (25% originality + 20% fit). The
   survey says a skimming reader files xscapes in the `terminal-pet` topic on
   sight, and that topic is crowded (`buddy` alone has 102 stars). The README
   already leads with the protocol, which is right; what it does not yet do is
   say what nobody else does. The four gaps are the copy.

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
- **Stars for completed todos** — the last unbuilt channel. The protocol carries
  `todo` with `n`/`of`; nothing renders it.
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
