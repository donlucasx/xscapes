# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code
hook schema — trust it, do not re-derive it. Skim origin-chat.md only if you need
the why; ignore ideas.md — it is parked. Tell me where we left off, then pick up
from ▶ NEXT.
```

## Where we left off (2026-08-31, commit `7b67126`, 49 commits, branch `main`)

**It runs in a real terminal now, and Lucas has run it.** Most of this session
was his findings from doing that, plus a companion design study that is PAUSED
awaiting his pick.

### Shipped this session

- **The scene was never the size of the window.** `termSize` shelled out to
  `stty size`, which reads from ITS OWN stdin, and `exec.Command` gives a child
  /dev/null -- so it failed every time and returned the 80x24 fallback, always,
  on every machine. One bug wearing three costumes: doesn't fill the frame,
  garbles when shrunk, glitches when stretched. Now the `TIOCGWINSZ` ioctl,
  verified in a real pty at five sizes and across a live resize.
- **The composition is mirrored** -- companion right, litter growing leftward,
  sand from the left margin, moon at 0.28. The sprite is *flipped*, not moved:
  its tail sweeps from the right hip. Measured, it is also MORE robust when
  narrow, not less -- the cat stays whole to 14 columns against 16 before.
- **The waterline reserves a beach** (never fewer than 5 rows; identical to the
  old flat 80% at 24 rows and up) and **the sand tail degrades by dropping whole
  pieces** rather than chopping, so a narrow pane reads `edit handler.go` and
  not `edit internal/auth/ha`.
- **Colour on 256.** Terminal.app is NOT greyscale -- 216 real colours plus 24
  greys. The night went grey because the *palette* is dark and the cube has
  almost no resolution down there. The fix is to keep the darkness in the
  BACKGROUNDS and push chroma into the GLYPHS, which are the bright part of the
  frame: glyphs meant to carry colour went 76% -> 100% off the grey ramp.
  **`GlyphBoost` locked at 2.6**, tunable live with `ASCIISCAPES_CHROMA`.
- **Moonlight follows the moon.** `glitter()` hardcoded 0.72 while the moon had
  moved to 0.28 -- a reflection with no source. It dims as the moon wanes too.
- **The sand writes in ink the beach can be read against**, derived per frame
  from the palette rather than a constant pinned to midnight's sand.
- `-live` runs on the alternate screen, and repaints on resize by polling the
  now-cheap ioctl every frame instead of draining one SIGWINCH per frame.
- **`assets/frames/wired-turn.gif`** -- 15 seconds of one turn, real reducer,
  loops. A screenshot cannot show motion and this is the first artefact where
  the whole vocabulary moves at once.

### Paused, awaiting Lucas

The **companion study**. He is looking at `assets/frames/companion-study.png`
and will come back with a pick. **Nothing is defaulted** -- `NewCat()` still
returns exactly what shipped before this session.

- **Coats in the running**: cream, slate, sage, mauve, charcoal. He said
  cream/sage/slate "stand out best". Terracotta and ginger are OUT, too close
  to Claude's own mark.
- **Settled**: the nose, the toe tips, and inner ears = **inner shadow** (the
  coat's own dark tone, so the cat stays monochrome and a two-cell detail is not
  the only hue on the body).
- **Open**: which whisker variant -- lower long, upper long, even, short, sweep.
  They now FIND the fur and grow outward from it, which took four attempts: the
  head is not centred in its sprite, so any fixed cell offset connects on one
  side and gaps on the other. Three of those rounds were spent moving a number
  when the placement model was wrong.

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

1. **Take Lucas's companion pick** and make it the default -- coat, whisker
   variant, whether the toes stay. Then put it in the live scene so he sees it
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
5. **`done` and `needs_input` still look identical** -- both raise the same
   bubble. The brief locks *"distinct cues"* for them. Oldest open gap against a
   locked requirement.

## Open threads for Lucas

- **The companion pick** — the one thing blocking. `companion-study.png`.
- **Name** — still `asciiscapes`; `iixscapes` / `xscapes` on the list. Late call.
- **Charcoal is a bet on the terminal.** It looks best in truecolor and worst in
  256, where it goes grey. Slate is the safe pick if Terminal.app stays the
  daily driver; charcoal wins if Ghostty or iTerm2 gets installed. **No
  truecolor terminal is installed on this machine today.**
- **Stars for completed todos** — the last unbuilt channel. The protocol carries
  `todo` with `n`/`of`; nothing renders it.
- **A wide pane leaves an empty middle.** At 200x50 the cat and the tail sit at
  opposite edges with a lot of nothing between. Not wrong, just unused.
- Swimmers: no perspective scaling, and 2 of 18 drop out when a lane is
  oversubscribed.
- `CLAUDE.md` Milestone 1 list is stale — the installer is done; the tmux
  launcher is the only piece of it left.
- **No git remote. Nothing pushed anywhere.** 49 commits live only on this disk.
