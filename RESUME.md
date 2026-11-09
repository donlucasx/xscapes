# Resume — xscapes

## NOW (wrapped 2026-11-09 00:54 PST, session 48)

**Goal:** ship the engine host + the harbour as the next release, and enter the Hermes Agent Plugin
Hackathon (due **WED NOV 11, 5 PM EST**: plugin, ~75 s demo, catalog PR, reply in Teknium's thread
x.com/Teknium/status/2119270884800401440, typeform form.typeform.com/to/FwJLvOln).

**Done (s48):**
- `5565794` the agent runs inside Ghostty's engine (libghostty-vt, `internal/gvt`, `-tags ghosttyvt`;
  `-engine=ghostty` default in such a build, `-engine=` = the old band host): text reflows on resize.
  He confirmed live in Terminal.app and Ghostty. `3bea2ef` reverts the autowrap patch. PUSHED.
- `f3231a1` the harbour scape (`xscapes scape harbour`; cat on the sill, rain on the glass = the work);
  his live look: "seems to work ok". `4a9a839` two defects + 8 tests (closed engine segfault; first
  kitten appeared sitting). LOCAL, NOT PUSHED.
- Hermes plugin: `~/Documents/claude/hermes-xscapes` (local git `c67daa3`; 5 pytest + `hermes plugins
  validate` pass, scan safe; end-to-end into the real binary verified). NOT on GitHub.
- Installed `~/.local/bin/xscapes` = 4a9a839 + the engine (dirty-tree stamp).

**Next:**
1. Claude: read `~/Documents/kimi/xscapes-engine-harbour-2026-11-09/AUDIT-REPORT.md` (still running at the
   wrap), RE-VERIFY every finding, fix what holds.
2. His word: push f3231a1 + 4a9a839 (`gh auth switch --user donlucasx && git push origin main`).
3. His word: create github `donlucasx/hermes-xscapes`, push; put the SHA in `docs/catalog-entry.yaml`.
4. Claude, his call on scope: release builds with the engine (release.sh cross-compiles CGO-off today,
   so the install line ships the OLD band host): libghostty-vt per platform, a tag, the install line.
5. Demo ~75 s at a fixed window size; then HE opens the catalog PR, replies to Teknium, submits.
6. His: the plugin and the shell hooks both fire today on this machine (`xscapes uninstall hermes`
   if the plugin is installed).

**Open questions for him:** ship the engine to installers before Nov 11 or after? · default scape
(vista today) and whether the harbour joins the site · the park and the block views (drawn, not built).

**Carried open items (harvested from the archive):**
- Engine host gaps: no scrollback while running (transcript printed at exit) · bell, title, OSC 52,
  cursor shape not forwarded · kitty keyboard flags reported to the agent even under Terminal.app.
- Band host (fallback): narrow-drag drift in Ghostty unfixed; Terminal.app duplicate line unexplained.
- s42 cards: drain at exit · a sub-second session never binds · watcher events miss the event log.
- Windows/WSL untested · the pond idea (tokens consumed as a fill, `ideas.md`) · Reddit: comments by
  hand, the Built with Claude post after ~100 karma (`notes/reddit/`).
- Commons dispute (s47): DMs sent 10-26; thread drafts `notes/commons/` (local only); his to post.

**Resume:** "Read RESUME.md NOW, then the Kimi AUDIT-REPORT.md; verify before fixing."

Design and rules: `CLAUDE.md` (encoding rule, locked decisions); the harbour's spec `notes/rain/README.md`.
