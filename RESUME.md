# Resume — xscapes

## NOW (checkpoint 2026-11-11 ~16:55 PST, session 50: CATALOG MERGED · CI LIVE · WINDOWS PR IN REVIEW)

**Goal:** Hermes Agent Plugin Hackathon, waiting on results (posted by @NousResearch). Meanwhile: the Windows port
from a contributor (PR #1) and the Commons callout post.

**Done:**
- Catalog PR #152904 approved + MERGED by Teknium 11-11 23:53 UTC (merge b797853; pin 3a12d11, v0.2.1). Fork branch
  deleted. His QT of the demo post with the install line is live.
- CI on xscapes main: `.github/workflows/test.yml` (linux/mac required, windows non-blocking) d278e98 e8ea38e; test
  fixes f19b66a (per-arch vista hash, trace test isolates XDG_CONFIG_HOME). main protected; fork PRs need approval.
- PR #1 (Celestin/CaptN420, windows-port, 3da834f) audited 3 ways (Claude, independent agent, Kimi): no malice,
  accept with changes. 10-item fix list sent by him (local: notes/contrib/pr1-reply-2026-11-11.txt).
- Commons evidence archive notes/commons/evidence/2026-11-11/ (local, excluded). Post v8 drafted; DM to Slayed sent.

**Next:**
1. Celestin pushes fixes -> Lucas clicks "Approve and run" on PR #1 -> Claude reviews ONLY commits after 3da834f
   + the windows CI job. Merge after Hermes results + fixes + green windows run; then make windows a required check.
2. Commons post (Lucas, Thu 11-12 ~10-11 AM PT) if Slayed hasn't answered: Claude re-sweeps X first, updates dates
   (seven->eight weeks if later), adds "dm'd slayed again on nov 11". File: notes/commons/post-v8-2026-11-11.txt.
   Lucas: screenshot the judges' DM threads + top of the Slayed thread for the archive (X chat won't load in Playwright).
3. Lucas: send the "fork it, MIT, calm not a game" reply to Celestin (drafted in chat 11-11) if he wants.

**Blockers:** none. **Known:** 74f0534 holds the old DM draft in public history; rewrite blocked by protection, advised against.
**Open questions:** none.
**Later:** release the Ghostty engine (libghostty-vt per platform). Older carried items: RESUME-archive-2026-11-09.md.

**Resume:** "Read RESUME.md NOW; `gh pr view 1 --json commits` (new commits after 3da834f?) and check X for Slayed/@commonsmade."

Design and rules: `CLAUDE.md` (encoding rule, locked decisions); the harbour's spec `notes/rain/README.md`.
