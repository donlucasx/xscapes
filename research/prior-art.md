# Prior art: ASCII waiting screens for the terminal

Surveyed 2026-09-01, to answer one question: has anyone already built what
xscapes is building? **Trust this file; do not re-derive it.** Re-run the
survey only if something below turns out to be stale.

Short answer: the space is crowded, but in four separate categories, and
xscapes does not sit cleanly in any of them. The four gaps at the bottom are
the reason the project still has room.

## Method

Web survey, not a code read. Star counts, licences and integration details
come from each project's own README as of this date, except where marked
unverified. Nothing here was measured by running it.

## 1. Terminal screensavers (idle-triggered or run by hand)

The oldest category. Looks closest at a glance, is agent-blind in every case.

| Project | Language | Notes |
|---|---|---|
| [cmatrix](https://cmatrix.org/) | C | Matrix rain, the canonical one |
| [unimatrix](https://linuxcommandlibrary.com/man/unimatrix) | Python | Matrix rain with Unicode |
| [pipes.sh](https://github.com/badjuice/pipes.sh) / [pipesX.sh](https://github.com/pipeseroni/pipesX.sh) | bash | The pipeseroni family |
| [asciiquarium](https://github.com/cmatsuoka/asciiquarium) | perl | A single script. **The only real *place* in the whole survey** |
| [termsaver](https://github.com/brunobraga/termsaver) | Python | Pluggable framework: matrix, asciimation, clock, RSS, source scroll |
| [ASCII Saver / ascsaver](https://gitlab.com/mezantrop/ascsaver) | C | **Has a trigger**: watches for absence of terminal I/O and starts |
| [ascii-splash](https://github.com/reowens/ascii-splash) | -- | 17 patterns, 102 presets, 5 themes, mouse |
| cacademo (libcaca) | C | Metaballs, plasma, moire, transitions |

All of these own the whole screen and know nothing about what you are waiting
for. Wallpaper, not instrumentation. `ascsaver` is the only one with any
notion of "the terminal went quiet", and even that is a raw I/O check with no
idea what produced the silence.

## 2. Spinners and per-command wait animations

- [terminal-animations (tan)](https://jorexdeveloper.github.io/terminal-animations/) -- bash, wraps a command, animates while it runs, multi-shell
- [Python `animation`](https://pypi.org/project/animation/) -- decorator based; bar, spinner, dots, ellipses
- [delayviewer](https://pypi.org/project/delayviewer) -- spinner plus stopwatch
- the cli-spinners / ora lineage that everything else copies

One line, bound to a process lifetime, two states (running, done).

## 3. Fake-activity generators

- [genact](https://github.com/svenstaro/genact) -- Rust, MIT. Fake compiles, fake mining, fake botnets. [Show HN, 2018](https://news.ycombinator.com/item?id=16166973)
- [hollywood](https://opensource.com/article/18/2/command-line-tools-productivity) -- Apache 2.0, splits the terminal and launches busy-looking apps

Relevant only as evidence: people will watch a terminal do nothing, at length,
if it looks good enough. That is the whole bet xscapes makes about dead time.

## 4. Agent-aware indicators (the real neighbourhood, all recent)

This category barely existed two years ago. It is where the competition is.

**[pi-animations](https://github.com/arpagon/pi-animations)** -- MIT, ~26
stars, TypeScript. The closest in spirit. 26 animations for the pi agent,
hooked to three phases (thinking, working, tool execution), thinking wins
ties. Single-line variants render inline via `setWorkingMessage()`; multi-line
ones (3 to 5 rows) render as a widget above the editor via `setWidget()`.
Full-width effects (fire, plasma) scale to the terminal. Requires pi v0.60+,
true colour, and a Nerd Font for the icon ones. Its own docs call them pure
loading indicators, not scene compositions: each is a standalone render
function of a frame counter plus a state enum.

**[claude-code-mascot-statusline](https://github.com/TeXmeijin/claude-code-mascot-statusline)**
-- MIT, 23 stars, TypeScript. Closest on the *reacts to real state* axis. A
half-block pixel sprite, 16 cells wide, driven by the Claude Code hook system
across 9 states (idle, thinking, tool running, tool success, tool failure,
permission prompt, subagent, done, auth success), plus a summary line with
branch, model, tool count, context percent, cost. Explicitly event-driven
rather than continuously animated, and the README admits the displayed state
can lag the session.

**[tweakcc](https://github.com/Piebald-AI/tweakcc)** -- patches Claude Code's
own bundle to swap spinners, thinking verbs, themes, input styling.

**[Claude Code's own spinner](https://medium.com/@kyletmartinez/reverse-engineering-claudes-ascii-spinner-animation-eec2804626e0)**
-- reverse-engineered writeup.

**[GitHub Copilot CLI's animated ASCII banner](https://github.blog/engineering/from-pixels-to-characters-the-engineering-behind-github-copilot-clis-animated-ascii-banner/)**
-- a real engineering post on pixels-to-characters. Worth reading for the
glyph work regardless of the competitive angle.

**The [terminal-pet topic](https://github.com/topics/terminal-pet)**, filling
up fast: `buddy` (102 stars), `buddymon` (44), `petsonality` (16 MBTI ASCII
animals), `tokengotchi` (feeds on LLM tokens), `desk-waifu` (mirrors Claude
Code work state via hooks), `familiar`, `codex-pets`, `commitachi`,
`purrminal`, `zenith-terminal-buddy`.

### Demand signal

All open, no maintainer response visible as of this date:

- [claude-code#66284](https://github.com/anthropics/claude-code/issues/66284) -- customizable ASCII working animation. Labelled `area:tui` + `enhancement`. Asks for `workingAnimation` in settings.json, either a frames array or a `command` script mirroring `statusLine`
- [claude-code#29200](https://github.com/anthropics/claude-code/issues/29200) -- custom thinking words
- [claude-code#35249](https://github.com/anthropics/claude-code/issues/35249) -- animated mascot in the status line
- [opencode#24937](https://github.com/anomalyco/opencode/issues/24937) -- terminal pet plus CAVA audio visualiser in the TUI sidebar

Note the shape of #66284: if Anthropic ships the `command` variant, a
statusline-style hook for the working animation, that is a distribution
channel for a scape strip. It is not a threat to `xscapes inside`.

### Unverified

An "ascii-agents" Rust TUI (weather and ambient effects, multi-floor layout
for concurrent agents, hook-safe) surfaced only via an aggregator page
([toolhunter.cc](https://www.toolhunter.cc/tools/ascii-agents)). No GitHub URL
appeared in any result. If it is real it is the nearest competitor found, so
this is the one thing worth chasing down.

## The four gaps

1. **Nobody runs the agent inside the scene.** Every agent-aware project is a
   strip beside or above the agent (statusline, widget, inline spinner). The
   screensavers own the screen but only when the agent is absent. The pty band
   plus scroll region in `xscapes inside` has no analog in this survey.
2. **Nobody encodes the work.** These are decorative loops keyed at most to a
   3-to-9-value state enum. Nothing maps *how much* happened into *what the
   scene shows*. "The water is the work, the sky is the world" has no
   competitor because nobody else treats the animation as an instrument.
3. **Everything is a pet, not a place.** The agent-aware category is almost
   entirely mascots and tamagotchis. The one real place is asciiquarium, which
   is twenty years old and knows nothing about your session.
4. **Everything is host-locked TypeScript.** tweakcc patches Claude Code's
   bundle; pi-animations needs pi v0.60+. A pluggable adapter protocol over an
   event stream, shipped as one Go binary, is genuinely unusual here.

## What this means for the README

Gap 3 cuts both ways. The terminal-pet topic is crowded and growing, and a
casual reader will file xscapes there on sight. The first screenshot has to
make it obvious this is an instrument in a place, not a companion. The
`assets/frames/wired.png` framing (one turn, driven by real events) is the
right instinct; the risk is that a reader who skims sees a cat and stops.
