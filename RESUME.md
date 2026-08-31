# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. Skim origin-chat.md only if you need the why;
ignore ideas.md — it is parked. Tell me where we left off, then pick up from
▶ NEXT.
```

## Where we left off (2026-08-31, commit `44be1c4`, 28 commits, branch `main`)

**The visual vocabulary is complete. Nothing is wired to a real session yet.**
Every frame you have seen is driven by hardcoded numbers.

Built and working:
- **Renderer** — three-layer alpha compositor, `TERM_PROGRAM`-gated 256/truecolor,
  O(1) xterm-256 quantisation, ANSI + plain + HTML output.
- **Shore scape** — day-cycle palettes (midnight/dawn/noon/dusk), travelling
  swells whose count and height carry activity, whitecaps above half activity,
  sub-cell waterline, foam, sand grain, moon carrying context in phase AND altitude.
- **Companion** — sitting cat in quadrants; states resting / working / needs-you /
  worried; breathing, blinking, tail; speech bubble; a side-view walk that exists
  but is NOT the plan.
- **Subagents as kittens** — uniform size chosen by count with hysteresis
  (6 to shrink, 4 to grow back), all on the near layer at full alpha, roughly one
  in three swimming in lanes pitched by sprite height, occlusion rims so overlaps
  read as depth.
- **Live terminal mode** — `./asciiscapes -live`.

## How to look at things

```
go build -o asciiscapes . && ./asciiscapes -live      # the real thing, Ctrl-C to quit
./asciiscapes -live -tod=0.75 -ctx=0.85               # dusk, low context
open assets/frames/cat-everything.gif                 # whole scene in motion
```
GIFs open directly. HTML demos need `python3 -m http.server` in `assets/frames/`.
Demo flags: `-anim -compare -layout -context -day -busy -kittens -sheet -strip -html`.

⚠ `-plain` is blind to the moon and the shoreline — both live in the background colour.

## ▶ NEXT — event plumbing (this is the gap)

1. **Event protocol** — Unix socket + JSON-lines file fallback, `asciiscapes emit <event>`.
2. **Claude Code hook adapter** — with the **60-second-nag suppression**: 43% of
   `Notification` events are an idle nag exactly 60s after `Stop`, measured over
   12,809 real events. Naive wiring makes the cat knock falsely every turn.
   ⚠ The installer must **merge** into `~/.claude/settings.json`, not overwrite —
   that file holds Lucas's secret-expansion PreToolUse guard.
3. **Verify `subagent_type` actually arrives** in the hook payload. Believed yes,
   never confirmed; the old debug log recorded event names only.
4. Then re-tune against real data: is the sea too twitchy on rapid tool calls?
   Does activity want smoothing over seconds or minutes? Do subagent counts ever
   reach the numbers we designed for?

## Open threads for Lucas

- **Name** — still `asciiscapes`; `iixscapes` / `xscapes` on the list. Decide late.
- **Stars for completed todos** — the last unbuilt channel.
- Swimmers: no perspective scaling (a kitten at the horizon is the same size as one
  near shore), and 2 of 18 drop out when a lane is oversubscribed.
- `CLAUDE.md` Milestone 1 list is stale — it still names the tmux launcher and the
  installer as blockers, both deprioritised.
- No git remote. Nothing pushed anywhere.
