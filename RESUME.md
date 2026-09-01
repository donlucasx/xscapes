# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code
hook schema — trust it, do not re-derive it. Skim origin-chat.md only if you need
the why; ignore ideas.md — it is parked. Tell me where we left off, then pick up
from ▶ NEXT.
```

## Where we left off (2026-08-31 session 7, last code change `bf4a329`, 54 commits, branch `main`)

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

### Paused, awaiting Lucas (unchanged except as noted)

**Nothing is defaulted** -- `NewCat()` still returns what shipped before
session 6.

- **Coats in the running**: cream, slate, sage, mauve, charcoal. He said
  cream/sage/slate "stand out best". Terracotta and ginger are OUT, too close
  to Claude's own mark. (Told this session: slate is the safe pick while
  Terminal.app is the daily driver; charcoal is a truecolor bet.)
- **Settled**: the nose, the toe tips, and inner ears = **inner shadow**.
- **Whiskers: LINES, his guideline plain** (`b93d679`), after he rejected the
  braille detour ("you had it going fine w the lines... use it"). ONE design:
  top '─' on the nose row x2 cells, bottom '‾' tucked on the row below x1,
  flush at the fur, passing behind the tail. **Root cause of the two earlier
  line rejections found**: the study portraits drew the cat one row lower
  than every live surface, parking the nose row ON the waterline -- waves
  visually extended the whiskers ("too long") and the bottom one floated in
  the wave field ("too low"). Portraits now use c.H-2-chh like live. Study
  resent; AWAITING his verdict on the corrected framing. The guide file:
  `~/Downloads/Screenshot-2026-08-31-at-2.52.07 PMsd.gif` (nbsp in name).
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

1. **Take Lucas's whisker steer off the FULL study** (he has now been sent the
   untruncated one) and build the revision round he asked for -- then coat and
   toes. Make the pick the default and put it in the live scene so he sees it
   moving rather than posed.
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
