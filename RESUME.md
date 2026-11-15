# Resume — xscapes

## NOW (checkpoint 2026-11-14 late, session 50 end: LISTING v0.2.2 IN REVIEW · COMMONS POST MONDAY)

**Goal:** Hermes Agent Plugin Hackathon, waiting on results (@NousResearch). Side tracks: Windows port (PR #1), Commons callout.

**Done:**
- Catalog: #152904 merged 11-11 (v0.2.1); #154720 merged 11-13 (Nous-palette banner, 26d14fc). Found: the marketplace
  NEVER renders `screenshots:`, and strips raw HTML from the README -> GIFs only show as Markdown images.
- hermes-xscapes v0.2.2 `4c71d9c` (README Markdown images pinned to xscapes 63fffa5, typing hero loop, trimmed copy,
  disclosure now names cwd + provider). Catalog PR **#155988** open: sha, version, bulleted description (his wording,
  slop-passed), screenshots removed. Kimi copy audit: ~/Documents/kimi/xscapes-listing-copy-2026-11-14 (fixes verified).
- Short install line `hermes plugins install xscapes --enable` everywhere (b620c1d); site published, verified live.
- Hero loop recipe: notes/hermes-demo/pipeline/edit-hero-typing.json (render.mjs needs npm i playwright @xterm/xterm
  @xterm/addon-webgl + an endsrc/ folder; set up in a scratch copy).
- CI on main (linux/mac required), main protected. PR #1 (Celestin, 3da834f) audited, fix list sent; no new commits.

**Next:**
1. Lucas, Mon 11-16: Commons post (notes/commons/post-v10-2026-11-12.txt; NOT posted 11-12; "seven weeks" still true).
   Claude re-checks X that morning. Then DMs: notes/commons/dm-sendlist-2026-11-12.txt (21, fill [link]).
2. #155988 merges -> Lucas: `hermes plugins update xscapes`; Claude deletes fork branch plugin-catalog-xscapes-0.2.2
   and checks the live page shows the hero + 4 GIFs.
3. Celestin pushes -> Lucas clicks "Approve and run" -> Claude reviews only commits after 3da834f + windows CI.
   Merge after Hermes results; then make windows a required check.

**Deferred by him:** the "fork it, MIT, calm not a game" reply to Celestin.
**Ideas:** per-subagent work in the scene needs Hermes hooks for child tool calls (ideas.md, 11-14).
**Known:** 74f0534 old DM draft in public history (advised leave). Next xscapes release carries the short install hint.
**Later:** release the Ghostty engine. Older carried items: RESUME-archive-2026-11-09.md.

**Resume:** "Read RESUME.md NOW; `gh pr view 155988 -R NousResearch/hermes-agent`; `gh pr view 1 --json commits`."

Design and rules: `CLAUDE.md` (encoding rule, locked decisions); the harbour's spec `notes/rain/README.md`.
