# asciiscapes — project brief for Claude Code

Working name: **asciiscapes** (not final; see open questions). A cozy ASCII "thinking screen" for terminal AI agents. While Claude Code (or any agent) works, a small living scene runs beside it — landscape, reactive weather, and a companion animal — and nudges the user, visually and with a sound, when the agent finishes or needs input.

Owner: Lucas (cinematographer + solo dev, LA). Solo build, AI-assisted. Start from this file; do not re-derive decisions already made here.

## Hackathon context

*(Facts below verified 2026-08-30 against the live site and Lucas's logged-in screenshots. Do not re-derive.)*

- Commons "Make Waiting for AI Fun" hackathon (`commonsmade.com/hackathons`). **Aug 27 – Sep 17, 2026**, three weeks. One entry per builder, editable until the deadline. Submit an early version by ~Sep 1 and iterate.
- Prizes: **1st $20,000 · 2nd $8,000 · 3rd $4,000 · 4th–19th $500 each · Vault $20,000 + 80% of all revenue generated.** $60,000 total. Treat payout as a bonus; the product is worth building regardless.

### Judging criteria — five weighted. Build against this.

| weight | criterion | the question asked |
|---|---|---|
| **30%** | Waiting Experience | Does it genuinely make waiting better? |
| **25%** | Originality | Is the idea new or surprising? |
| **20%** | Fit | Does it feel native to using AI agents? |
| **15%** | Repeatability | Would users enjoy it again? |
| **10%** | Execution | Does the prototype clearly prove the idea? |

Judges **to be announced**. Their framing line, verbatim: *"The biggest opportunity may be building the company that owns the waiting layer for AI."*

**What the rubric implies, and it should drive every tradeoff:**
- **Execution is only 10%.** Polish is nearly worthless. A rough scene that lands the idea beats a smooth one that doesn't. Never trade scope for finish.
- **Originality + Fit = 45%, and both favour the terminal build.** "Native to using AI agents" is precisely what a scene living in tmux beside Claude Code *is*; a browser tab is a worse answer. On a vibe-coding platform, near every rival entry will be a web app — a Go binary in a terminal is the only one of its kind in the pile. This is the moat.
- **Lead the submission with the layer, not the landscape.** The event protocol + pluggable adapters *is* "the waiting layer for AI" they say they want to fund. Most entries will be one clever screen. Frame asciiscapes as the protocol with a reference implementation, and the cozy scene as what it looks like.
- 30% Waiting Experience is the notification doing its job: the nudge has to genuinely beat a terminal bell.

### Building on Commons' platform — resolved

**No rule requires it.** What exists is a **token leaderboard** incentivising platform use; it is marketing, and it appears nowhere in the rubric. **Build the TUI here.** Use Commons only for genuinely web-shaped side pieces — the landing/submission page and a browser scape gallery for judges who will not install a binary — which feeds the leaderboard for free.

- Commons free tier: **150 AI credits/day**, **600/month**, resets 8/31/2026. Wallet **$0 and stays $0** — paying to climb a leaderboard that isn't scored buys nothing. (Daily-vs-monthly interaction is contradictory; `BALANCES`/`USAGE`/`MODELS` panels hold the real per-model rates and are still unexpanded.)
- Model for the web pieces: **Qwen3 Coder Next on Quick · $**, fed specs written here so it transcribes rather than designs. **Never Expert · $$$** (it pays premium to plan work already planned). Skip Superspeed · $$ (that buys latency, not quality). DeepSeek V4 Pro only if something truly needs reasoning.
- Code mode is chat-based (`/chat/new`); no sandbox or terminal observed. GitHub `donlucasx` is connected to the account.

### Demo video
Highest-leverage deliverable: 45–60s real screen recording, no narration — prompt in Claude Code → scene reacts → companion knocks → Enter → back in the agent. Record in iTerm2 or Ghostty (truecolor), **not** Terminal.app.

## Locked decisions

**Lifecycle**
- Session-long, not task-long: launch once when the agent session starts, exit when it ends. Tasks modulate the scene (weather rises while working, settles to a resting state while waiting on the user). Day/night cycle spans the session.
- Persistence is tiny: companion identity/memory + a per-repo seed (keyed by git root) so the same repo always gets the same landscape shape. Nothing else accumulates. No decay.

**Placement**
- Lives alongside the agent via tmux: popup (`display-popup`) on think by default, split pane as an option. Zellij floating pane later.
- Lucas uses Terminal.app with no tmux. `asciiscapes claude` must bootstrap tmux itself (claude in main pane, scape beside). Only dependency: tmux via brew. No-tmux fallback: second Terminal window via `osascript`. Never seize the agent's TTY.

**Stack**
- Go + bubbletea/lipgloss, single static binary. No Node, no Python.
- Design target 80×24; must look fine at 40×12.
- Glyphs: Unicode blocks + braille for water/fire, ASCII fallback. Truecolor with automatic 256-color fallback (Terminal.app is 256 max).
- Renderer is a real layer/alpha model from day one: three depth layers per scene (far α≈0.3 slow parallax sparse glyphs, mid α≈0.6, near α=1.0 dense fastest). Alpha = fg color blended toward bg/lower layer; quantize to palette on 256-color. Companion always in near layer. Weather modulates far-layer alpha (fog = one number).
- Frame pacing and clean redraws matter more than scene count; judges will screen-record it.

**Agent integration**
- Two layers: a tiny event protocol + thin adapters. Engine listens on a Unix socket with a JSON-lines file fallback. Adapters translate each agent into the protocol.
- Events: `session_start`, `prompt`, `tool` (with kind: read|write|edit|search|shell|web|subagent|todo|mcp), `error`, `test_pass`, `test_fail`, `compact`, `needs_input`, `done`, `session_end`.
- Adapter 1: Claude Code hooks (SessionStart/End, UserPromptSubmit, PreToolUse/PostToolUse, Notification, Stop, SubagentStop, PreCompact). Lucas already has a Stop/Notification hook that beeps — reuse it as the first adapter and replace the beep.
- Adapter 2: generic "watch this process" fallback (busy = alive + output activity; done = prompt back). Test targets: Claude Code, Kimi, Hermes. Verify what hooks Kimi/Hermes expose before writing adapters.
- Coarse output parsing is acceptable for agents without hooks.

**Scapes (three for v1, each with exactly one toy)**
- Shore: waves (layered sine + foam), stars, moon, sand. Toy: skip a stone.
- Campsite: campfire (Doom-fire algorithm) foreground, night sky, tent. Toy: log on the fire.
- Rainy window: rain streaks, blurred city lights, lightning. Companion-less scape — weather delivers the notification. Toy: wipe fog off the glass.
- Which survive is TBD after seeing them rendered.

**Weather / activity mapping**
- Coding: read=rain, edit=wind, write=new growth, shell=thunder, error=lightning, tests pass=clearing/rainbow, subagents=birds flocking, compaction=fog.
- Research: web search=stars appearing, fetch=lantern on horizon, long tool-less reasoning=tide/dusk deepening, todo=path stones laid. Research sessions favor sky scapes; coding favors ground scapes.
- Universal: elapsed time drives day cycle; needs_input=companion knocks / thunder; done=dawn.
- Growth: files touched → trees/objects, placed deterministically by path hash.

**Companion**
- 3–4 lines tall, resident not subject. States: resting, working (small idle motion), needs-you (walks to the edge nearest the agent pane, shows `!`).
- Shortlist: cat and bird. Fox and wisp as alternates. Decide after rendering real frames.
- Companion is global (same across repos); has a name.

**Interactivity (minimal — it must never become a game)**
- v1 ships exactly three interactions: pet the companion, plant a tree, one physics toy per scape.
- One keypress, zero commitment, never competes with the agent for attention.

**Notification**
- Distinct cues for `done` vs `needs_input`. Companion delivers it where present; weather (lightning/ember pop/foam) where not.
- One on-theme sound per scape (bell, bird, thunder) via `afplay`/`paplay`. Silent during work; ambient audio is off by default and optional.
- Keypress on notification focuses the agent pane (tmux `select-pane`). Fallback when pane hidden: tmux `display-message` + OS notification.
- One-line status on tap ("in auth/ for 3 min").

**Shipping**
- GitHub release binaries + brew tap. MIT, public from day one.
- State in `~/.config/asciiscapes/`.

## Open questions
1. Name: asciiscapes vs Meanwhile / Lull vs companion-first (Moss, Wisp, Pip).
2. Companion: cat or bird (decide from rendered frames).
3. Which of the three scapes survive.
4. ~~Must the build happen on Commons' platform?~~ **Answered 2026-08-30: no.** Token leaderboard only, absent from the rubric.
5. ~~Does the Commons "Code" tab have a sandbox/terminal?~~ **Chat-only as far as tested.** GitHub account is linked; repo import untested.
6. Per-model credit rates on Commons, and how 150/day reconciles with 600/month.

## Milestone 1 (target: ~Sep 1, submittable)
1. Go module, bubbletea app, 80×24 canvas, layer/alpha renderer with truecolor→256 fallback.
2. Shore scape, three layers, working/resting states.
3. Event protocol (socket + file), `asciiscapes emit <event>` CLI for testing.
4. Claude Code hook adapter (generate hook JSON into `~/.claude/settings.json` via `asciiscapes install claude`).
5. `asciiscapes claude` launcher that bootstraps tmux (main pane + popup/split).
6. needs_input/done cues with one sound, keypress focuses agent pane.
7. Cat companion, three states.
8. README with one-line install; submit.

## Milestone 2
Campsite + rainy window, weather mapping, plant-a-tree, per-repo seed, generic process adapter, Kimi/Hermes tests, bird companion, demo video.

## Working style
- Direct, honest tradeoffs over optimism. Verbatim-ready commands.
- Lucas will push back when advice doesn't match what he sees; take that seriously.
- Keep it lean. If a feature makes it feel like a game, cut it.
