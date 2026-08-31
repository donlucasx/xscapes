# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code
hook schema — trust it, do not re-derive it. Skim origin-chat.md only if you need
the why; ignore ideas.md — it is parked. Tell me where we left off, then pick up
from ▶ NEXT.
```

## Where we left off (2026-08-31, commit `419ceee`, 30 commits, branch `main`)

**The scene is wired to a real session.** Hook events go in one end and the sea,
the companion, the kittens and the sand come out the other. Before this, every
frame was driven by hardcoded numbers.

- **`internal/event`** — the protocol. One JSON object per line; per-session unix
  datagram socket with a JSON-lines spool as the fallback. The scape never learns
  what a hook is and the adapter never learns what a wave is.
- **`internal/reduce`** — the fold: events → `scape.Activity` + `companion.State`
  + kitten count + sand tail. Takes an explicit clock, so the constants are
  testable at a hundred times real speed.
- **`hook.go`** — the Claude Code adapter. 2.37 ms p50, `async: true`, never
  writes to stdout, never exits non-zero.
- **`install.go`** — merges into `~/.claude/settings.json`. Prints a plan by
  default; `--apply` to write. Install → uninstall is byte-for-byte identical on
  the real 60 KB file.
- **`asciiscapes emit | replay | statusline`** — drive the scene with no agent at
  all, which is how the constants get tuned and how a third adapter gets written.

### What the binary told us that we had been guessing

`notes/claude-hooks-verified.md` holds the hook payload schema, extracted from the
Zod definitions embedded in the Claude Code 2.1.251 binary. Three long-open
questions closed:

1. **Subagent identity arrives.** `SubagentStart`/`SubagentStop` carry `agent_id`
   and `agent_type`. The field is `agent_type`, not `subagent_type`. (RESUME item 3
   from last session — answered.)
2. **The 60-second nag needs no timing heuristic.** `notification_type` is a
   required field and the nag is `idle_prompt`. The measured log (2,220 at exactly
   60 s after `Stop`, 41.6% of all Notifications) is what said a filter was needed;
   the schema said how to write it. The allow list fails closed.
3. **Context comes from the statusline, not a hook.** No hook carries it, and
   summing transcript usage against a hardcoded 200k is *measurably wrong here*:
   164,450 tokens is 82% of 200k and 16% of the 1M window actually in use.

Also fixed a latent renderer bug that only appears once activity moves: the wave
clock scaled *absolute* elapsed time by level, so any activity change teleported
the sea, worse the longer the session ran (447 of 1920 cells at ten minutes in
vs 44 at five seconds). Phase is integrated now, with a regression test that was
rewritten after the first version passed against the broken code.

## How to look at things

```
go build -o asciiscapes . && ./asciiscapes -live      # the real thing, Ctrl-C to quit
./asciiscapes -wired assets/frames/wired.html         # a simulated turn through the REAL reducer
./asciiscapes install claude                          # prints a plan, writes nothing
./asciiscapes emit tool_start -tool Read -target x.go # drive the scene by hand
```
GIFs open directly. HTML demos need `python3 -m http.server` in `assets/frames/`.
Demo flags: `-wired -anim -compare -layout -context -day -busy -kittens -sheet -strip -html`.

⚠ `-plain` is blind to the moon and the shoreline — both live in the background colour.

## ▶ NEXT — three known defects, then real data

Seen in `assets/frames/wired.png`, all three still open:

1. **The sand is written on the water.** `drawSand` (live.go) anchors to the canvas
   bottom, so as lines accumulate the tail drifts up across the sea and foam.
   It must anchor to the waterline — `Shore.lastEdge` already knows where that is.
2. **The tail collides with the companion.** Both start near x=2 and share rows
   19–21. Inset the text past the cat's width.
3. **`done` and `needs_input` look identical.** Both raise the same bubble in the
   NeedsYou pose. The brief locks *"distinct cues for `done` vs `needs_input`"* —
   this is a gap against a locked requirement.

Then:
4. **Install it for real and run a day on it.** Everything above is tested against
   synthetic streams and the verified schema; nothing has yet run against a live
   session. `./asciiscapes install claude --apply`, then a new Claude Code session.
5. **Re-tune the reducer against a recording.** `asciiscapes replay` exists for
   exactly this. Are `TauFall=12s`, `TurnFloor=0.30`, `FlightFloor=0.45` right
   against Lucas's actual working rhythm? Do real subagent counts ever reach the
   numbers the kitten ladder was designed for?

## Open threads for Lucas

- **Name** — still `asciiscapes`; `iixscapes` / `xscapes` on the list. Decide late.
- **Stars for completed todos** — the last unbuilt channel. The protocol carries
  `todo` with `n`/`of`; nothing renders it.
- **The statusline chain is not installed automatically.** `install` prints the
  line to add rather than taking over the statusLine Lucas wrote. Until it is
  added by hand, the moon stays at zero.
- Swimmers: no perspective scaling, and 2 of 18 drop out when a lane is
  oversubscribed.
- `CLAUDE.md` Milestone 1 list is stale — the tmux launcher and the installer are
  no longer both blockers; the installer is done.
- No git remote. Nothing pushed anywhere.
