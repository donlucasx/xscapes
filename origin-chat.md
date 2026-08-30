# Origin — condensed record of the planning conversation (Aug 28–29, 2026)

Reference only. Not for the agent's working context.

1. **Starting idea.** ASCII "screensaver" for the terminal: cozy landscapes like YouTube ambient videos (starry night, waves) rendered in ASCII, full-screen or pane-sized.

2. **Hackathon research.** Commons (commonsmade.com, a Solana "VibeFi" token project) posted a $50–60K hackathon. Rules screenshot revealed the theme: "Make Waiting for AI Fun — build something for users while the agent is thinking," creativity over polish, one entry per builder, deadline Sep 17, "every token you spend counts toward the board." The idea was reframed from screensaver to *thinking screen*.

3. **Video-to-ASCII.** Confirmed mature tooling (chafa, ffmpeg+libcaca, mpv --vo=caca). Lucas's cinematography background makes real-footage loops a unique angle. Decided procedural for v1, footage as a later ambient mode, hybrid (footage base + procedural overlay) as the long-term play.

4. **Design space.** Explored agent compatibility (hooks → structured streams → process signals → terminal state), full color (truecolor ladder, Kitty/Sixel rejected), and three philosophies: ambient (ignores the task), reflective (weather mirrors activity), constructive (the task builds the scene). Chose constructive-on-reflective: growing landscape + reactive weather + companion animal.

5. **Interactivity.** Lucas wants minimal — never a game. Settled on three v1 interactions: pet, plant, one physics toy per scape. Notification must nudge visually and by sound, replacing his existing Claude Code beep hook; keypress focuses the agent.

6. **Decision pass.** Name: asciiscapes for now. Session-long episodes (simpler and better than task-long). Terminal.app, no tmux → tool bootstraps tmux. Go + bubbletea. 80×24. Three scapes: shore, campsite (campfire + night sky merged), rainy window. Companion shortlist cat and bird; fox and wisp alternates. Test agents: Claude Code, Kimi, Hermes. Per-scape sound; companion delivers it, weather where no companion.

7. **Depth.** No per-character opacity in terminals; fake it with color blending toward bg, luminance-as-distance, glyph density ramps, parallax, per-cell bg. Decision: build a three-layer alpha renderer from day one.

8. **Commons platform.** Screenshot of their Code mode: presets Quick/$, Superspeed/$$, Expert/$$$ (plans + verifies); models GPT-5.6 Luna, DeepSeek V4 Flash/Pro, Qwen3.7 Flash/Plus, Qwen3 Coder Next, MiniMax M2.7. Wallet $0. If required to build there: DeepSeek V4 Pro primary, Expert for architecture, Qwen3 Coder Next for fast loops. Flagged that a browser chat can't run/test a TUI; whether builds must happen there is still unresolved.

9. **Handoff.** Decisions moved to `CLAUDE.md`, parked ideas to `docs/ideas.md`, this summary to `docs/origin-chat.md`.
