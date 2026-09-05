# xscapes — project brief for Claude Code

*(Renamed end to end on 2026-09-03: directory, env vars, state path and hook marker. Two names are kept on purpose and are not leftovers -- `internal/envx` still reads `ASCIISCAPES_*` and warns, and `install.go` still RECOGNISES the `# asciiscapes:v1` marker so the hooks it wrote before the rename can be found and removed.)*

> **Session 16 (2026-09-05).** Live tests PASSED in both terminals; shipped: kitten swim-off ·
> the context READOUT in the scene from 40% used (his ruling) · ▄ split cells on Terminal.app (no
> hairlines) · the disc-tip sky half through the ramp · the outstanding-todo ring REMOVED (his
> ruling). Measured: Terminal.app's alt screen RETAINS rows at their widest on a width change
> (`notes/width-audit.md`). Studies for his pick: "The Moon, Four Ways" (quad edge, sun without
> shadow, night halo; `xscapes -moon`). **Brand LOCKED 2026-09-05** by the parallel brand session
> (Cursor + Reverse alt; guidelines v1.0 in `assets/brand/`); deck = parked draft; see `RESUME.md`
> "Brand workstream". Details in `RESUME.md`.
>
> **Session 15 (2026-09-04).** His picks: gradients **cube-path ONLY**; Terminal
> automation **this session only** (tty rule); the page stays unpublished until the
> pending items are in. **Built, measured, and INSTALLED** (*"Install it"*): the sky and the sea are
> painted as one PATH through the 256 palette (`term.Ramp`, `canvas.SetBGRamp`) instead
> of rounding each row — hard edges sky 84→65, sea 120→99, half-hours with a step ≥30
> 29→6, the largest step unchanged at 33 (the cube's green step; unavoidable). Page:
> artifact "Sky and Sea Repainted". Pushed; `~/.local/bin/xscapes` rebuilt. His first live
> look (124x52): the moon was a RECTANGLE — pre-existing, the disc's radius is 1.92 rows there
> and lost its tips — **FIXED** with half-row sampling (`canvas.SetBGHalves`); the companion's
> right margin now grows with the width (5 columns at 124, was 2) — **FIXED**. NEXT: live in it.
> **Afternoon, his first run in Ghostty**: the sun differed (truecolor painted the raw tan blend)
> and a resize broke the layout (Ghostty's alternate screen anchors content to the TOP on a grow
> and moves the cursor with its row on a shrink; the host's tick encoded Terminal.app only). Both
> SHIPPED and installed on *"standardize the experience"*: **the cube on every terminal**
> (`DetectProfile`), and **`host.Rules` per TERM_PROGRAM** (`RulesFor`; Terminal.app byte-identical,
> everything else xterm-like). Committed `b9d65e7`, pushed. His last note: the companion's EYES are
> holes to the sea by day (by design, eyes sit in gaps of the body bitmap) — his pick. Details in `RESUME.md`.
>
> **Session 14, the evening.** He lived in it and came back with six reports
> (`_FEEDBACK.md`): three FIXED (swimmers above the waterline, balloon text
> capped, the model's ESC ( B); the scroll-back glitch NOT reproduced in the
> buffer; the transcript-start garble PARTLY diagnosed (Ink re-rendering while
> his 747 permission warnings scroll the band; one model divergence fixed, the
> proper diff still to do); and **the gradient review MEASURED, nothing changed,
> his decision pending** (`notes/gradientaudit`, artifact "Sky and Sea by the
> Hour"). ⚠ An automation incident typed `/exit` into HIS window: no Terminal
> driving without his OK, and never `front window` (hub memory).
>
> **Where we left off — 2026-09-03 (session 14).** The submission page exists
> (`site/`, one static file, `xscapes -site site`) and is NOT yet published or
> submitted — his hands. The scrollback plan below was AUDITED at his ask and
> REPLACED: the alternate screen has no history (measured), but DECSET 47
> switches buffers without clearing, so the terminal's OWN scrollback can be fed
> while the band stays up; and every shrink was misplacing Claude's input box
> (the terminal leaves the cursor, the host restored it into a band that no
> longer held its row) — Report 1, reproduced and FIXED. His ruling: *"Go:
> shrink fix, then mirroring."* Plan and measurements in
> `notes/scrollback-audit.md`. **Both SHIPPED the same evening**: `RebindShrinkAlt`
> and `Host.History` (rows leaving the band mirrored into the main buffer
> through DECSET 47, replayed on exit; `-history`, default on in Terminal.app).
> The screen model is production now (`internal/host/screen.go`). NEXT: live in
> it for a day.
>
> **Where we left off — 2026-09-03 (session 13), HEAD `9452bc6`, pushed, tree clean.**
> **Session 13: the resize damage was OURS, twice, and he had to report it four
> times.** Both fixed and mutation-proven — the clear was painting rows in
> Claude's own background (an erase fills with the CURRENT background), and the
> host never undid the terminal's downward push on a grow.
>
> **The measured fact that settles all of it:** Terminal.app's ALTERNATE screen
> anchors CONTENT to the BOTTOM edge both ways, and the CURSOR moves with
> NEITHER. **The rule worth more than the fixes:** an instrument can answer a
> question you did not ask — four did today, each producing a confident wrong
> answer to him. An external audit (agent Kimi, at his suggestion) caught the
> worst one before any measurement did.
>
> ⚠ ~~**SCROLLBACK IS A REQUIREMENT WITH NO CHEAP ANSWER.**~~ **Superseded
> 2026-09-03 (s14): there IS a cheap answer, measured — mirror into the main
> buffer through DECSET 47 (see above).** The rest of this paragraph was true
> as far as it went: the main screen reflows on width, tmux stacking leaves a
> seam, and the alternate screen has no history of its own.
>
> **Session 12 was the rename.**
> Finished end to end: directory
> `~/Documents/claude/xscapes/`, `XSCAPES_*` env vars, `~/.config/xscapes/`, and
> the `# xscapes:v1` marker on the twelve installed hooks. `ASCIISCAPES_*` still
> works and warns (`internal/envx`); `install.go` still RECOGNISES the old marker
> so pre-rename hooks can be found and removed. ⚠ Two running scapes went deaf in
> the move and need `xscapes claude` again. Nothing else about the project changed
> — session 11's findings below still stand, as do the three things waiting on him.
>
> **Session 11:**
> **Live: https://github.com/donlucasx/xscapes** (public, MIT). Milestone 1 is
> COMPLETE, the hooks are installed and firing, and the agent runs INSIDE the
> scape on the alternate screen (`xscapes claude`; `-beside` is the old tmux
> layout, `inside <cmd>` hosts anything, `-alt=false` reverses the alt screen).
>
> **⭐ SESSION 11 FOUND THAT NOTHING MARKED DONE ACTUALLY WAS.** Seven defects in
> shipped features, none visible in any study, all fixed: the sea and sky had NO
> COLOUR on Terminal.app for most of a working day · the 256 sky was the wrong
> HUE, not just banded · an electric night that shipped AND was pushed · a
> grey-fringed sun · a blob moon · a resize leaving sky in the agent's transcript
> · kittens losing an eye to their neighbour's seam. Two channels were measured
> for the first time and both were broken: the sea's dynamic range was collapsed
> into a fifth of itself (the floors CLAMPED instead of lifting), and the
> companion's alarm is on **37% of active time**.
>
> **The durable part is the instruments**, and the rule they teach: build the
> instrument before trusting the picture, and measure the RENDERED frame rather
> than what the source says it should do.
> `xscapes tune` folds real recordings through the reducer offline (`-sweep` for
> settings) · `internal/host/screen_test.go` is a small terminal that replays the
> host's real bytes through eleven resize modes · `xscapes shades` asks the
> smoothing question in HIS terminal at HIS window size · `-day` shows five
> panels an hour including the glyph cells 256 changed.
>
> **⭐ TARGET IS TERMINAL.APP, his ruling 2026-09-02.** Cube-exact colour is the
> general rule; `term.Index256Keeping` preserves hue for backgrounds and leaves
> GREYS ALONE, which is the guard that stops it forcing an electric night.
>
> ⚠ **Three things wait on him**: submit the entry (no record it is done, plan
> said ~Sep 1) · the worry trigger (a locked channel, so raising the bar is his
> call) · the banding decision at full window size.
> ⚠ ~~**One limitation that is not a bug**: a resize scrambles the AGENT's own
> text. The scape is provably right.~~ **WRONG, and corrected 2026-09-03 (s13).
> Both halves of that damage were the HOST'S**, found after he reported it a
> third time. (1) The resize clear emitted `ESC[2K` with no SGR of its own, and
> an erase fills with the CURRENT background -- so it painted rows in whatever
> colour Claude was mid-draw, and those rows scrolled into scrollback as a wall
> of black. (2) `drop` is a MAIN-screen correction and was applied on the
> ALTERNATE screen too, where nothing slides on a shrink, so the clear walked up
> into the transcript and at a big enough shrink ate the input box.
>
> **Both were invisible to the instrument that exonerated the host**: `screen`
> discarded SGR ("colour is not what these tests are about") so a row filled
> with colour read as blank, and every resize test ran `AltScreen: false` while
> production runs on the alternate screen. The lesson is not about resizing --
> it is that a harness which no-ops a platform behaviour, or never enters the
> mode production runs in, will keep returning a clean bill of health.
>
> It remains TRUE that Claude Code emits nothing on a resize, so anything the
> TERMINAL moves stays moved until a keystroke. That part was never the bug.

**Name: `xscapes`** (decided 2026-09-01 with the repo; asciiscapes and iixscapes are out; carried through everything 2026-09-03). A cozy ASCII "thinking screen" for terminal AI agents. While Claude Code (or any agent) works, a small living scene runs WITH it — a shoreline whose sea rises with the work, and a companion animal — and nudges the user, visually and with a sound, when the agent finishes or needs input.

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
- **Lead the submission with the layer, not the landscape.** The event protocol + pluggable adapters *is* "the waiting layer for AI" they say they want to fund. Most entries will be one clever screen. Frame xscapes as the protocol with a reference implementation, and the cozy scene as what it looks like.
- 30% Waiting Experience is the notification doing its job: the nudge has to genuinely beat a terminal bell.

### Building on Commons' platform — resolved

**No rule requires it.** What exists is a **token leaderboard** incentivising platform use; it is marketing, and it appears nowhere in the rubric. **Build the TUI here.** Use Commons only for genuinely web-shaped side pieces — the landing/submission page and a browser scape gallery for judges who will not install a binary — which feeds the leaderboard for free.

- Commons free tier: **150 AI credits/day**, **600/month**, resets 8/31/2026. Wallet **$0 and stays $0** — paying to climb a leaderboard that isn't scored buys nothing. (Daily-vs-monthly interaction is contradictory; `BALANCES`/`USAGE`/`MODELS` panels hold the real per-model rates and are still unexpanded.)
- Model for the web pieces: **Qwen3 Coder Next on Quick · $**, fed specs written here so it transcribes rather than designs. **Never Expert · $$$** (it pays premium to plan work already planned). Skip Superspeed · $$ (that buys latency, not quality). DeepSeek V4 Pro only if something truly needs reasoning.
- Code mode is chat-based (`/chat/new`); no sandbox or terminal observed. GitHub `donlucasx` is connected to the account.

### Demo video
Highest-leverage deliverable: 45–60s real screen recording, no narration — prompt in Claude Code → scene reacts → companion knocks → Enter → back in the agent. **Record in Terminal.app** — that follows from his 2026-09-02 ruling, and it reverses what this line said before. The scape is now built for the 256-colour cube, so a truecolor terminal would show a picture no user gets.

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
| context remaining | **moon** phase *and* altitude; numeric readout of what is LEFT under the moon from **40% used** (his ruling 2026-09-05), warm &ldquo;NN% left&rdquo; from 85% | done &mdash; the readout was decided in s6 and marked done, but was never in the live scene until 2026-09-05 (`drawReadout`, `ReadoutFrom`) |
| time of day | **sky colour**, real wall clock | done |
| weather | **deferred, not rejected** &mdash; no rain, clouds, fog or sync in v1; the thinking is parked in `ideas.md` | deferred 2026-08-30 |
| needs you | **bubble**, rare: needs_input, error, done. Nothing else | done — distinct cues shipped 2026-08-31: ask = warm SOLID box + alert pose; done = cool DOTTED knock + content `^ ^` pose, bounded by DoneHold |
| companion identity | **coat + face**: cream/slate/sage/mauve/charcoal, nose, toes, inner-shadow ears, whiskers | ⏸ options built, awaiting Lucas's pick; nothing defaulted |
| what it is doing now | **text written in the sand**, newest brightest, older fading as the tide takes them | done — anchored to the waterline, degrades by dropping whole pieces when narrow. **The lower beach falls away to black (`DefaultSandFade` = 1.0, locked 2026-09-01)**: contrast on the newest line 132→204 at midday, 148→204 at night, and equal at every hour, so legibility stops depending on the clock. Ink is sampled from the PAINTED background per row, never the palette's nominal sand. |
| todos completed | **star count** | done 2026-09-02 &mdash; a constellation in the upper sky, `*` for each finished todo. ~~`&#8728;` for each outstanding one, so it reads *n of N*~~ &mdash; **the ring is GONE (his ruling 2026-09-05: "discard the ring altogether, it's not clear what it means")**; the sky says *n*. Position is fixed by index and seed so a star lights where it always was. Held at a visibility floor like the moon: a completed todo is a fact about the AGENT and `StarVis` is 0 at noon. ⚠ **TodoWrite has been called ZERO times in the whole recorded history** &mdash; 13,682 tool events &mdash; so today it only lights from `xscapes emit todo` or the demo cycle. |
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

### Layout &mdash; ⚠ SUPERSEDED 2026-09-01, agent goes INSIDE the scape

**Everything in this section describes the side-by-side design, which he has
ruled out**: *"The entire Claude experience should happen within the xscape,
not next to it."* It still ships and still works, so it is kept as the record
of what is running today, but it is no longer the target. The replacement is
mocked in `-overlay` and not built. Read `RESUME.md` ▶ NEXT before touching it.

### Layout (the shipped, superseded design) &mdash; supersedes the popup default below
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
- Lucas uses Terminal.app with no tmux. `xscapes claude` must bootstrap tmux itself (claude in main pane, scape beside). Only dependency: tmux via brew. No-tmux fallback: second Terminal window via `osascript`. Never seize the agent's TTY.

**Stack**
- Go + bubbletea/lipgloss, single static binary. No Node, no Python.
- Design target 80×24; must look fine at 40×12.
- Glyphs: Unicode blocks + braille for water/fire, ASCII fallback. **The 256-colour cube on EVERY terminal** (decided 2026-09-04 when Ghostty's truecolor picture differed: the scene is designed for the cube, and indices 16-255 are the same everywhere). `XSCAPES_COLOR=truecolor` opts into the untuned raw palette.
- **256 is not greyscale** — it is 216 colours plus 24 greys. A dark palette
  collapses to grey because the colour cube has almost no resolution below luma
  25 (4 entries, all pure blue) against 108 above 150. So the darkness lives in
  the BACKGROUNDS and the colour lives in the GLYPHS, which are the bright part
  of the frame. `term.GlyphBoost` (2.6, `XSCAPES_CHROMA` to override) lifts
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
- State in `~/.config/xscapes/`.

## Open questions
1. ~~Name~~ **ANSWERED 2026-09-01: `xscapes`.**
2. Companion: cat or bird (decide from rendered frames).
3. Which of the three scapes survive.
4. ~~Must the build happen on Commons' platform?~~ **Answered 2026-08-30: no.** Token leaderboard only, absent from the rubric.
5. ~~Does the Commons "Code" tab have a sandbox/terminal?~~ **Chat-only as far as tested.** GitHub account is linked; repo import untested.
6. Per-model credit rates on Commons, and how 150/day reconciles with 600/month.

## Milestone 1 (target: ~Sep 1, submittable) — status 2026-09-01
1. ✅ Go module, 80×24 canvas, layer/alpha renderer with truecolor→256 fallback. (No bubbletea; stdlib only.)
2. ✅ Shore scape, three layers, working/resting states.
3. ✅ Event protocol (socket + file), `xscapes emit <event>` CLI for testing.
4. ✅ Claude Code hook adapter + `xscapes install claude`.
5. ✅ `xscapes claude` launcher — bootstraps tmux, joins an existing session,
   osascript fallback, `-print` dry run. Verified in a real tmux.
6. ◑ needs_input/done cues ✅ (distinct: warm solid ask box vs cool dotted knock, and
   two sounds via `internal/notify`, edge-detected so a 60s nag rings once).
   **Keypress-focuses-agent-pane is NOT built** — the live loop reads no keys.
7. ✅ Cat companion — five states (resting, working, needs-you, done, worried).
8. ✅ README + MIT LICENSE written 2026-09-01; every command in it verified by running it.
   Published 2026-09-01 at github.com/donlucasx/xscapes; clone-and-build verified from the public repo.

## Milestone 2
Campsite + rainy window, weather mapping, plant-a-tree, per-repo seed, generic process adapter, Kimi/Hermes tests, bird companion, demo video.

## Working style
- Direct, honest tradeoffs over optimism. Verbatim-ready commands.
- Lucas will push back when advice doesn't match what he sees; take that seriously.
- Keep it lean. If a feature makes it feel like a game, cut it.
