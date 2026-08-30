# Parked ideas

Not in v1. Kept so they're not lost. Nothing here should influence Milestone 1.

## Weather, clouds, storms (deferred 2026-08-30, NOT rejected)
Cut from v1 to keep the encoding clean, not because the ideas are bad. The rule
that pushed them out: **the water is the work, the sky is the world.** Swells
carry activity and the companion carries breakage, so weather had no job left --
and atmospheric motion risked being misread as agent activity.

Worth bringing back once the vocabulary is settled and taught:
- **Clouds** as a pure mood layer, drifting on the far layer at low alpha. Safe
  because they are slow and horizontal, unlike anything the sea does.
- **Storms** as an event punctuation rather than a state: a single lightning
  flash on a hard failure, gone in two frames. The persistent state stays with
  the companion; the flash is just the exclamation mark.
- **Fog on compaction.** The one per-event weather mapping worth keeping -- fog
  reads as "losing detail", which is exactly what compaction does.
- **Rain** as an ambient mode the user chooses, never event-driven.
- **Real weather synced from the user's location**, which is safe only while
  weather carries no information. Needs a network call and a location
  permission, so it must be opt-in with clear-sky as the default.
- **Real weather picking the SCAPE** rather than the weather within it: raining
  outside gives you the rainy-window scape, clear night the shore. The world
  sets the stage, the agent drives the drama, and they never touch.

## Scenes and rendering
- Pre-rendered footage loops: shoot or AI-generate 20–30s cozy loops (real surf, fireplace, rain on window), convert with `chafa` to text frames, pack into the binary, composite procedural overlays on top. Pipeline:
  ```
  ffmpeg -i waves.mp4 -vf "fps=12,scale=160:-1" frames/%04d.png
  for f in frames/*.png; do chafa --size=80x24 --format=symbols --colors=full "$f" > "${f%.png}.txt"; done
  ```
  Use steady-state footage (surf, fire) and crossfade the last second into the first to hide the loop seam. This is a visual signature nobody else in the hackathon will have. Tradeoff: fixed aspect, not reactive.
- Kitty graphics protocol / Sixel for real pixels in Kitty, WezTerm, Ghostty, iTerm. Rejected for v1: stops being terminal-native and half the terminals lose it.
- More scapes: snowfall over a cabin, koi pond, forest clearing with wind, aurora (needs truecolor), night city skyline, lighthouse.
- Time-of-day palette from the local clock, independent of session length.

## World
- Persistent per-repo world that grows over months; untouched files wither, deleted files become dead trees to prune, stale branches become leaves to rake. Rejected for v1 (sad-looking repos, more state); episodes instead.
- Camera drift: arrow keys pan across the world to "see" where the tests live.
- Session as journey: each tool call is a step on a path; arrival = done.
- Shared skies: opt-in, other users' agents thinking right now appear as distant stars or lanterns. Nice Commons tie-in.
- Long tasks accrue "daylight" the user can spend on a sunrise.

## Companion
- Pet memory that deepens: sits closer over time, remembers being fed.
- Looks toward the mouse cursor in the pane; startles on error.
- Carries a lantern to the door; letting it in (Enter) is the notification.
- Companion-first product framing (the Clippy move): product name = the pet's name.

## Interaction
- Catch a firefly, whistle for wind (hold key, trees bend), throw a stone into the pond.
- Micro-status on tap.
- Ambient audio companion process (off by default).

## Integration
- Adapters for Codex, Aider, OpenCode, Cursor CLI via their structured output streams (`--output-format stream-json` style).
- Web overlay version for IDE-integrated agents (Cursor, Windsurf) — a different product, not a feature.
- Library/spinner drop-in agent authors can call instead of a spinner.
- Parse Bash output for test pass/fail, error streaks, retry loops.

## Hackathon-only angles
- Onchain-reactive scene (stars = txs, wave height = volume) if a crypto tie-in helps with Commons judges.
- Paid scene packs / token-gated unlocks for the "Vault" revenue prize. Only if it doesn't compromise the product.

## Names considered
Moment-first: Meanwhile, Lull, Interlude, Idle, Between. Place-first: Hearth, Porch, Nook, Cabin. Companion-first: Moss, Wisp, Pip, Ember, Tuft.
