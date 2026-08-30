# Resume — asciiscapes

**Copy-paste this prompt into a fresh session:**

```
cd ~/Documents/claude/asciiscapes/ and read CLAUDE.md (the brief, authoritative) and RESUME.md
before responding. Skim origin-chat.md only if you need the why; ignore ideas.md for Milestone 1.
Tell me where we left off, then pick up from the immediate next move.
```

## Current state (2026-08-29, session 1 — brief absorbed, no code yet)
- **Scope locked.** `CLAUDE.md` is the state of record: cozy ASCII thinking screen for terminal
  agents, Go + bubbletea, 80x24, three-layer alpha renderer, tmux popup, event protocol + adapters,
  three scapes, three interactions. Never a game.
- `ideas.md` = parked, must not influence Milestone 1. `origin-chat.md` = how we got here.
- **No code, no Go module, no git repo.** Nothing built.

## Environment (verified 2026-08-29)
- **Go: NOT INSTALLED.** `brew install go` is the first move. Hard blocker on Milestone 1 item 1.
- tmux **3.6b** at `/opt/homebrew/bin/tmux` — `display-popup` needs >=3.2, so fine.
- Terminal.app v466, 256 colors max. ffmpeg + afplay present. chafa absent (parked idea only).

## Two findings already measured (don't re-derive)
1. **`Notification` is not `needs_input`.** Across 12,809 real hook events in `~/.claude/hook-debug.log`
   (3.5 months), 2334 of 5436 Notifications (**43%**) fire at exactly 60s after a `Stop` — Claude Code's
   idle nag, not a request for input. Naive mapping makes the companion knock falsely on every turn.
   Suppress Notifications within ~61s of a Stop with no `UserPromptSubmit` between.
2. **`COLORTERM=truecolor` is a false positive here.** It's set inside a Claude Code session in
   Terminal.app, which renders 256 only, and it does not come from any shell rc. Gate colour on
   `TERM_PROGRAM`, not `COLORTERM`, or the default launch environment picks the wrong path.

⚠ `asciiscapes install claude` must **merge** into `~/.claude/settings.json`, not overwrite it — that
file holds the `PreToolUse`/Bash hook that blocks secret expansion. Overwriting removes a live guard.

## Hackathon rubric (verified 2026-08-30) — build against this
30% Waiting Experience · 25% Originality · 20% Fit ("native to using AI agents") ·
15% Repeatability · **10% Execution**. Judges TBA. Their framing line:
"The biggest opportunity may be building the company that owns the waiting layer for AI."

⇒ Execution barely counts. Originality + Fit (45%) both favour the **terminal** build, not a web app.
**There is no rule requiring builds on Commons' platform** — only a token leaderboard, which is not
in the rubric. Plan stays: build the TUI. Use Commons for web-shaped side pieces only.

Commons free tier: 150 credits/day, 600/month, resets 8/31. Wallet $0 and **stays $0**.
Model for the web pieces: **Qwen3 Coder Next on Quick·$**. Never Expert·$$$ (it pays to plan
work we already planned here).

## Built so far (2026-08-30)
Go 1.27 module. `go build ./...` and `go vet ./...` clean.
- `internal/term` — RGB + alpha blend, colour-profile detection, O(1) xterm-256 quantisation.
- `internal/canvas` — the three-layer alpha compositor (far .30 / mid .60 / near 1.0) plus three
  renderers: ANSI, plain glyphs, and **HTML** (one span per cell, for looking at frames).
- `internal/scape` — `Scape` interface, `Activity{Working, Level}`, deterministic hash.
- `internal/scape/shore.go` — night shore: sky gradient, twinkling stars, moon, swell lines,
  moon-glitter path on the water, sub-cell waterline, foam speckle, sand grain. Working vs resting
  changes wave speed, amplitude and how far the water runs up the sand. Moon size, wave reach
  and foam density scale with the canvas, so 40x12 composes rather than just fitting.

Reference frames: `assets/frames/shore-{resting,working}.{html,png}`.

**Verified, not assumed:** colour gate returns 256 in Terminal.app despite `COLORTERM=truecolor`,
truecolor when overridden, truecolor when `TERM_PROGRAM` is absent · `-ascii` emits **zero**
non-ASCII bytes · renders at 40x12.

### How to look at a frame
```
go run . -html=assets/frames/x.html -working -frames=40
cd assets/frames && python3 -m http.server 8731     # Playwright blocks file://
```
⚠ `-plain` is blind to the moon and the shoreline — both are painted into the background colour now.

## Companion (2026-08-30)
Quadrant rendering chosen: best legibility at 1:1 and exact ASCII glyph metrics.
`internal/companion` = Sprite, narrow-safety allow-list + tests, Bitmap with
braille/quadrant/char converters, `Bubble()`, and `Cat` with three states.
The cat is a 24x28 bitmap; the tail is a curve evaluated per frame, so wagging
costs no extra art. Breathing shifts the body by 2 source pixels = one quadrant
subpixel = half a character cell, the smallest vertical step the medium has.
Eyes are two characters plotted over the quadrant body -- `-` dozing, `o` open,
`O` alert, plus a blink. Two cells carry the whole expression.

Look at it: `go run . -anim=assets/frames/companion-anim.html` then serve the
dir (Playwright and the page's JS both need http, not file://).

## Immediate next move
1. Better cat art -- the current one reads, but it is still chunky.
2. Event protocol: Unix socket + JSON-lines file fallback, `asciiscapes emit <event>`.
3. Claude Code hook adapter **with the 60s-nag suppression** — see the measured finding above.
4. bubbletea TUI wrapper + tmux launcher. Renderer is deliberately stdlib-only so far; bubbletea
   only wraps it.

## Git
Repo initialised 2026-08-30, branch `main`, one commit (`5cdf62c`). No remote yet, nothing pushed.
`.gitignore` covers the built binary, generated `assets/frames/*.html`, and `.DS_Store`; the
reference **PNGs are tracked deliberately** as the visual record.

## Open questions for Lucas
- Expand Commons' `BALANCES`/`USAGE`/`MODELS` — per-model credit rates still unknown.
- Move `CLAUDE.md`/`ideas.md`/`origin-chat.md` into `docs/` as origin-chat §9 intended?
