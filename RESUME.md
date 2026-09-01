# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code
hook schema — trust it, do not re-derive it. Skim origin-chat.md only if you need
the why; ignore ideas.md — it is parked. Tell me where we left off, then pick up
from ▶ NEXT.
```

## Where we left off (2026-09-01, last code change `427a047`, 60 commits, branch `main`)

**Milestone 1 is one README away from submittable.** The companion study is
PARKED at his request — see "Parked" below.

### Shipped 2026-09-01

- **Two notification sounds** (`internal/notify`). 30% of the rubric is the
  waiting experience and the note says the nudge must beat a terminal bell; a
  scape in a side pane is not the pane being looked at, so sound is the only
  channel that reaches the user. Bright chime = the agent is BLOCKED on you;
  deep sonar note = it finished. Keyed off the BUBBLE, not the pose (a broken
  build outranks a question in the pose, so a pose-driven sound would go silent
  on the one event that needs answering), and edge-detected so Claude's
  60-second nag rings once. Silent when following nothing, `ASCIISCAPES_SILENT`
  to mute, `asciiscapes notify` to audition.
- **`asciiscapes claude`** — the launcher. Bootstraps tmux, or joins the window
  it is already in, or falls back to a second Terminal via osascript. Agent
  keeps its own pane and TTY (exec, not wrap). `-print` is a dry run.
- **`-live -await`** — the scape used to bind once at startup, so launching both
  halves together left it in demo mode forever. It now keeps looking, and
  deliberately ignores the session pointer present at startup (binding to a
  stale pointer SUCCEEDS and shows a dead session beside a live agent).
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

**Nothing is defaulted** -- `NewCat()` still returns what shipped before
session 6.

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
go build -o asciiscapes . && ./asciiscapes -live      # the real thing, Ctrl-C to quit
./asciiscapes -faces  assets/frames/companion-study.html   # THE OPEN QUESTION
./asciiscapes -colors assets/frames/color-study.html       # 256 vs truecolor
./asciiscapes -wired  assets/frames/wired.html             # a turn through the REAL reducer
./asciiscapes -info                                        # profile, size, chroma
./asciiscapes -mockup assets/frames/composition-study.html   # left vs mirrored, every terminal shape
./asciiscapes install claude                          # prints a plan, writes nothing
./asciiscapes emit tool_start -tool Read -target x.go # drive the scene by hand
```
GIFs open directly. HTML demos need `python3 -m http.server` in `assets/frames/`.
Demo flags: `-wired -mockup -anim -compare -layout -context -day -busy -kittens -sheet -strip -html`.
`-mirror=false` gives the old left-anchored layout.

⚠ `-plain` is blind to the moon and the shoreline — both live in the background colour.

## ▶ NEXT

1. **README with a one-line install — it BLOCKS submission** (Milestone 1 #8,
   the only unbuilt item). Entries close 2026-09-17.
2. **Install for real and run a day on it.** Everything is still tested against
   synthetic streams and the schema read out of the Claude Code binary; no hook
   has ever fired into it. `./asciiscapes install claude` prints a plan and
   writes nothing; `--apply` writes, after a backup, and uninstall restores
   byte-for-byte.
3. **Add the statusline chain by hand** or the moon stays dark. Install prints
   the exact line rather than taking over the statusLine Lucas wrote.
4. **Re-tune the reducer against a recording.** `asciiscapes replay` exists for
   exactly this. Are `TauFall=12s`, `TurnFloor=0.30`, `FlightFloor=0.45` right
   against his actual rhythm? Do real fan-outs reach the kitten ladder's numbers?
5. **Keypress focuses the agent pane** — the last half of Milestone 1 #6. The
   live loop reads no keys at all today; it needs raw mode plus
   `tmux select-pane -t <agent>`. The launcher knows the pane id and could
   pass it.

## How to look at things (additions)

```
./asciiscapes claude -print          # the launcher's plan, writes nothing
./asciiscapes claude                 # agent left, scape right, one command
./asciiscapes notify                 # hear both knocks
ASCIISCAPES_SILENT=1 ./asciiscapes … # mute
```

## Open threads for Lucas

- **The companion pick** — coat + whisker revision + toes. Full study is in
  his hands now; waiting on his steer.
- **Name** — still `asciiscapes`; `iixscapes` / `xscapes` on the list. Late call.
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
- **No git remote. Nothing pushed anywhere.** 54 commits live only on this disk.
