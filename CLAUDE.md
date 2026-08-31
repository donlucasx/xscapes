# asciiscapes — project brief for Claude Code

> **Where we left off — 2026-08-31 session 7, last code change `bf4a329`, 54 commits.**
> **done and needs_input are now DISTINCT cues** — the oldest gap against a
> locked requirement is closed. `done` is its own companion state (content `^ ^`
> eyes, full tail held high and still) with a soft dotted balloon in the cool
> bubble colour; the ask keeps the solid box, now in a warm attention colour.
> Balloons are opaque (the sea used to write glyphs into the words) and the
> mirrored pointer now aims at the cat, not a kitten. Also: the companion-study
> PNG had ALWAYS been a cut-off capture — the five-coats, 256 and every-state
> rows were never in the file Lucas reviewed; recaptured full height.
> **The companion pick is still open**: he wants another whisker revision round
> (no specifics yet — steer expected after he sees the full study), toes still
> tbd, coat undecided. Nothing is defaulted.
> `go build`, `go vet`, `go test` and `-race` clean; no remote.
> **`notes/claude-hooks-verified.md` is the hook payload schema, read out of the
> Claude Code binary itself — trust it, do not re-derive it.**
> ▶ NEXT is his whisker/coat/toes steer, then installing it and running a real
> day. See `RESUME.md`.

Working name: **asciiscapes** (not final; see open questions). A cozy ASCII "thinking screen" for terminal AI agents. While Claude Code (or any agent) works, a small living scene runs beside it — a shoreline whose sea rises with the work, and a companion animal — and nudges the user, visually and with a sound, when the agent finishes or needs input.

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

### The encoding rule (decided 2026-08-30)

**The water is the work. The sky is the world.** Sea state always means the
agent; sky, light and time always mean reality. Nothing crosses. Every variable
gets its own perceptual channel, and a channel may be bound to the real world
only if nothing else needs it.

| variable | channel | state |
|---|---|---|
| agent busy, how hard | **swells**: how many are travelling, how tall, whitecaps above half | done |
| something is broken | **the companion**, not the weather: ears back, hunched, tail flat, amber eyes; persists until it clears | done |
| context remaining | **moon** phase *and* altitude; numeric readout silent until 65%, brightens at 85% | done |
| time of day | **sky colour**, real wall clock | done |
| weather | **deferred, not rejected** &mdash; no rain, clouds, fog or sync in v1; the thinking is parked in `ideas.md` | deferred 2026-08-30 |
| needs you | **bubble**, rare: needs_input, error, done. Nothing else | done — distinct cues shipped 2026-08-31: ask = warm SOLID box + alert pose; done = cool DOTTED knock + content `^ ^` pose, bounded by DoneHold |
| companion identity | **coat + face**: cream/slate/sage/mauve/charcoal, nose, toes, inner-shadow ears, whiskers | ⏸ options built, awaiting Lucas's pick; nothing defaulted |
| what it is doing now | **text written in the sand**, newest brightest, older fading as the tide takes them | done — anchored to the waterline, degrades by dropping whole pieces when narrow |
| todos completed | **star count** | not built |
| subagents | **kittens** | done — `agent_id`/`agent_type`, counted live |

Rejected and why: session-elapsed as its own variable (the real clock covers it,
and a session-relative sky lies about the world); weather carrying activity (it
was carrying two masters, which is the collision this rule exists to prevent);
tide-as-time-since-input, horizon-glow, driftwood-count (all fine, none
load-bearing &mdash; eight variables is already at the edge of what reads without
a legend).

**Encode in coverage, count or position &mdash; never in rate.** Activity was first
mapped to wave *speed* and idle was indistinguishable from flat-out, because a
glance is the entire budget and a screenshot has no motion at all.

### Six slots every scape must fill
light, sky, motion (the WATER, not the air), surface, accumulator, companion. Any scape providing all six
works with the whole encoding system for free. The rainy window currently fails
on companion &mdash; fix by putting the cat on the inside sill.

### Layout &mdash; supersedes the popup default below
The scape is a full pane and carries the activity tail written into the sand.
Not a popup: a popup covers the session, and the user should always be able to
see what the agent is doing.

**The composition is MIRRORED (locked 2026-08-31).** Companion on the RIGHT,
litter growing leftward, sand written from the left margin, moon at 0.28. The
cat sprite is flipped rather than moved, because its tail sweeps from the right
hip and would otherwise be pinned to the frame edge. Measured, this is also more
robust in a narrow pane than the old left-anchored layout, not less: the cat
stays whole to 14 columns instead of clipping at 16, and the sand survives to 30
columns instead of 34. `-mirror=false` still renders the old layout.

**The waterline reserves a beach**: never fewer than five rows of sand, which is
identical to the old flat 80% at 24 rows and above and only bites on a short
pane, where 80% used to leave a single row.


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
- **256 is not greyscale** — it is 216 colours plus 24 greys. A dark palette
  collapses to grey because the colour cube has almost no resolution below luma
  25 (4 entries, all pure blue) against 108 above 150. So the darkness lives in
  the BACKGROUNDS and the colour lives in the GLYPHS, which are the bright part
  of the frame. `term.GlyphBoost` (2.6, `ASCIISCAPES_CHROMA` to override) lifts
  glyph chroma toward a target of 100 before quantising; near-neutrals below
  chroma 30 are left alone so the companion does not turn orange.
- **Never emit an ANSI index below 16.** Those sixteen are the only colours a
  terminal profile can repaint; 16-255 are fixed by the xterm standard, which is
  what makes the scene look the same on every profile. Proven exhaustively.
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

**Activity mapping** *(supersedes the per-tool weather taxonomy, discarded 2026-08-30)*
- **Weather is deferred, not rejected.** No rain, clouds, fog, storms or
  real-weather sync in v1; the ideas are parked in `ideas.md` for after the
  vocabulary settles. The sky is time of day, stars and the moon; nothing else.
  This removes a network dependency, a location permission, and the risk of
  atmospheric motion being misread as agent activity.
- **Tool events do not each get their own visual.** There is no read=rain,
  shell=thunder taxonomy to memorise. Every tool event feeds ONE aggregate
  activity level, which drives the swells: how many are travelling and how tall.
- **Identity lives in the sand, not in the scene.** The activity tail names the
  tool and the file. The sea says how much; the sand says what. That split is
  why no per-tool vocabulary is needed.
- `error` / `test_fail` put the companion into its worried pose, which persists
  until it clears. `needs_input` and `done` raise the bubble.
- Growth: files touched leave driftwood on the sand, placed by path hash.
  Completed todos light stars. Subagents appear as kittens.

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
