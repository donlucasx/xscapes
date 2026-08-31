# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative)
and RESUME.md before responding. notes/claude-hooks-verified.md is the Claude Code
hook schema — trust it, do not re-derive it. Skim origin-chat.md only if you need
the why; ignore ideas.md — it is parked. Tell me where we left off, then pick up
from ▶ NEXT.
```

## Where we left off (2026-08-31, commit `7b9c189`, 33 commits, branch `main`)

**The scene is wired to a real session, reviewed, and mirrored.** Hook events go
in one end and the sea, the companion, the kittens and the sand come out the
other. `go build`, `go vet`, `go test` and `go test -race` all clean.

- **`internal/event`** — the protocol. One JSON object per line; per-session unix
  datagram socket, JSON-lines spool as fallback. The scape never learns what a
  hook is and the adapter never learns what a wave is.
- **`internal/reduce`** — the fold: events → `scape.Activity` + `companion.State`
  + kitten count + sand tail. Explicit clock, so the constants are testable.
- **`hook.go`** — the Claude Code adapter. 2.37 ms p50, `async: true`, never
  writes stdout, never exits non-zero.
- **`install.go`** — plan by default, `--apply` to write. Install → uninstall is
  byte-for-byte identical on the real 60 KB `settings.json`.
- **`emit` / `replay` / `statusline`** — drive the scene with no agent at all.

### The composition is mirrored (decided 2026-08-31, with Lucas)

Companion on the RIGHT, litter growing leftward, sand written from the left
margin, moon moved to 0.28. The cat sprite is flipped, not just moved — its tail
sweeps from the right hip, so unflipped it would be pinned against the frame.

Measured: the mirror is also *more* robust narrow. The cat stays whole down to
14 columns (was clipping at 16) and the tail is what clips first; the sand gets
the whole left side, so it survives to 30 columns instead of 34.

Two composition rules came out of rendering every real terminal shape:
- **The waterline reserves a beach.** It was a flat 80% of height, which at 14
  rows left ONE row of sand. Now never fewer than five rows. Identical at 24
  rows and above.
- **The tail degrades by dropping whole pieces**, not by chopping: detail first,
  then directories, so a narrow pane reads `edit handler.go` rather than
  `edit internal/auth/ha`.

### The review (2026-08-31): 28 confirmed, 2 refuted — all fixed

The blocker was mine and I had waved it through: the bubble was gated on the
companion's pose, so after any failed command the Worried pose swallowed a
permission prompt the agent was *actually blocked on*. Also two secret leaks
(`VAR=secret cmd`, raw WebFetch URLs), an installer that could delete a user
hook that merely mentioned our marker, and a printed statusline instruction that
deleted the user's statusline when followed. See `git show 7b9c189`.

## How to look at things

```
go build -o asciiscapes . && ./asciiscapes -live      # the real thing, Ctrl-C to quit
./asciiscapes -wired assets/frames/wired.html         # a simulated turn through the REAL reducer
./asciiscapes -mockup assets/frames/composition-study.html   # left vs mirrored, every terminal shape
./asciiscapes install claude                          # prints a plan, writes nothing
./asciiscapes emit tool_start -tool Read -target x.go # drive the scene by hand
```
GIFs open directly. HTML demos need `python3 -m http.server` in `assets/frames/`.
Demo flags: `-wired -mockup -anim -compare -layout -context -day -busy -kittens -sheet -strip -html`.
`-mirror=false` gives the old left-anchored layout.

⚠ `-plain` is blind to the moon and the shoreline — both live in the background colour.

## ▶ NEXT — install it and run a day on it

Everything so far is tested against synthetic streams and the schema read out of
the Claude Code binary. **Nothing has yet run against a live session.**

1. `./asciiscapes install claude --apply`, restart Claude Code, and watch it.
   The installer is proven byte-safe on a copy of the real file, but the first
   real install is still the first real install — read the plan before applying.
2. Add the statusline chain by hand (install prints the exact line) or the moon
   stays at zero.
3. **Re-tune against a recording.** `asciiscapes replay` exists for this. Are
   `TauFall=12s`, `TurnFloor=0.30`, `FlightFloor=0.45` right against Lucas's
   actual rhythm? Do real subagent counts reach the kitten ladder's numbers?
4. **`done` vs `needs_input` are still visually identical** — both raise the same
   bubble. The brief locks *"distinct cues"* for them. Last known gap against a
   locked requirement.

## Open threads for Lucas

- **Name** — still `asciiscapes`; `iixscapes` / `xscapes` on the list. Decide late.
- **Stars for completed todos** — the last unbuilt channel. The protocol carries
  `todo` with `n`/`of`; nothing renders it.
- **A wide pane leaves an empty middle.** At 200x50 the cat and the tail sit at
  opposite edges with a lot of nothing between them. Not wrong, just unused.
- **The statusline chain is not installed automatically.** `install` prints the
  line to add rather than taking over the statusLine Lucas wrote. Until it is
  added by hand, the moon stays at zero.
- Swimmers: no perspective scaling, and 2 of 18 drop out when a lane is
  oversubscribed.
- `CLAUDE.md` Milestone 1 list is stale — the installer is done; the tmux
  launcher is the only piece of it left.
- No git remote. Nothing pushed anywhere.
