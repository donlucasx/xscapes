# Commons: what the rules say, and what the submission mechanism actually requires

Surveyed 2026-09-02 by two independent agents (one via public web + `curl` against
Commons' static assets and unauthenticated API; one via the shipped client bundle
plus a Wayback diff). **Trust this file; do not re-derive it.** Re-run only if
something below turns out to be stale, or once someone reads the pages SIGNED IN
(see "The one thing nobody has verified", at the bottom).

This supersedes the 2026-08-30 conclusion in `CLAUDE.md` ("Building on Commons'
platform - resolved. No rule requires it."). That sentence is still true and is
still not the whole answer.

## Short answer

**No rule requires building on Commons. The submission API still only accepts a
Commons app.** Those are both true at once, and the second one binds.

## 1. The written rules, verbatim (all seven)

From `https://commonsmade.com/assets/ShellFrame-UQCgVAQP.js`, the bundle served at
`/hackathons`. Title `Rules`, subtitle `Make Waiting for AI Fun - hosted by Commons`.

- **What to build** - "Today, users send a prompt and stare at a loading screen. Make use of that time and build something for users while the agent is thinking."
- **No limits** - "There are no limits to what you can build: creative, fun, productive - anything."
- **The point** - "You are designing what happens while the agent thinks."
- **What we want** - "We care more about the creativity than polish."
- **Everything counts** - "Every token you spend from the moment you join counts toward the board - no toggles, no attribution."
- **One entry** - "One build per builder, changeable right up to the deadline. Your group is for help, not a shared entry."
- **The deadline** - "Entries close Sep 17. The board settles to submitted builds only; anything unsubmitted drops off."

Footer: "three weeks - 19 builders paid - judging is final".

That is the complete rules text. **No sentence requires the project to be built on
Commons, deployed on Commons, or runnable in a browser.** There is also no text
saying judges will not install software.

No FAQ, terms or changelog page exists publicly: `/rules` and `/faq` both redirect
to `/`.

## 2. The binding constraint: submission takes an appId and nothing else

Verbatim from `https://commonsmade.com/assets/index-DW7_59r0.js`:

```js
function ao(e,t){return B(`/hackathons/${encodeURIComponent(e)}/submissions`,{method:`PUT`,body:JSON.stringify({appId:t})})}
```

API base `https://api.commonsmade.com`. **The only payload is `{appId}`.** The
stored submission is `{appId, title, description, screenshotUrl, creatorUserId,
creatorName, xHandle, githubHandle, submittedAt}` - no field for an external URL,
a repo, a binary or a video.

There is no form. The entry UI is a checkbox list of the signed-in user's own
builds, populated from `Dt({page:1,limit:50,sort:'recent'}).apps` filtered to
`e.visibility === 'public'`. Its strings: `Your entry` / `No public builds yet -
publish one from its manage page first.` / `enter as many as you like` /
`add / remove ->` / `Entry submitted.`

A Commons "app" is a Commons-hosted web app: `POST /api/apps/{id}/publish`,
deployed to `<slug>.commons.app`. **There is no import path.** GitHub appears only
as a Privy identity (`github_oauth`), as `githubHandle` on a submission, and as
`githubRepositoryUrl` - an *outbound* repo link Commons attaches to an app it
built. No "import repo", "add URL" or "external project" flow exists in any
loaded bundle.

**Conclusion (inferred from client code, strongly supported, not stated in prose):
a Go binary in a GitHub repo cannot be entered as-is. Something must exist as a
published-public Commons app to appear on the entry list at all.**

Corroboration that entries really are Commons apps: the public, unauthenticated
`https://vibe.commonsmade.com/api/apps/public?page=1&limit=40&sort=recent` returns
hackathon-themed builds ("Thinklings", "JoyBox" - the latter's stored prompt opens
"Build JoyBox as a single-file web app (must work on mobile)").

## 3. Commons cannot host a terminal binary. Not a limitation to route around

- Code tile blurb, verbatim: `Build a full-stack web app`.
- Publish pipeline is hardcoded: `Building production bundle...` / `Deploying to the edge network...` / `Your app is live`.
- Infrastructure is Cloudflare: allowed hosts `new Set(['vibe.commonsmade.com','staging.build.cloudflare.dev','localhost','127.0.0.1'])`; crt.sh shows `*.vibe.commonsmade.com`, `api.`, `gateway.`, `privy.`. Workers/Pages means JS/TS/WASM on the edge. A long-running native process is not hostable. (INFERRED from the Cloudflare host plus the `.commons.app` publish target.)
- The builder's "terminal" is a **one-way log**: websocket events `terminal_output` (stdout/stderr), `file_changed`, `server_log`, surfaced as a read-only `terminalLogs` array. **There is no input channel.**
- The only language list in the bundle (`ts, tsx, js, jsx, json, css, html, md, py, sql, yaml`) is the editor's syntax highlighter, not runtime support. No occurrence of Docker, Rust, Go, golang, binary or CLI anywhere in the builder chunk.

## 4. Nothing changed after 2026-08-30

Today's bundle diffed against the Wayback snapshot of
`https://web.archive.org/web/20260828215510id_/https://commonsmade.com/assets/ShellFrame-B_Rop35I.js`
(captured 2026-08-28 21:55 UTC):

- the `rules` + `structure` + `criteria` block is **byte-identical**;
- the submission logic is functionally identical (only minified identifiers differ).

No announcement, founder statement or staff post about a platform requirement
exists anywhere public - X, HN, Product Hunt, Reddit, LinkedIn, YouTube all
searched. **Nothing found. Do not read a rule out of that silence.**

## 5. Judging criteria - CONFIRMED unchanged

`30% Waiting Experience` "Does it genuinely make waiting better?" ·
`25% Originality` "Is the idea new or surprising?" ·
`20% Fit` "Does it feel native to using AI agents?" ·
`15% Repeatability` "Would users enjoy it again?" ·
`10% Execution` "Does the prototype clearly prove the idea?"

Footer: "The biggest opportunity may be building the company that owns the waiting
layer for AI." The table in `CLAUDE.md` is exactly right.

**Judges: still `To be announced`.** The avatar row in the code
(`VitalikButerin`, `HsakaTrades`, `0xngmi`, `punk6529`, `CryptoHayes`) is
logged-out decoration, NOT an announcement. Do not treat those as the judges.

## 6. Corrections to what we believed

- **The leaderboard we dismissed as marketing is not the hackathon's board.** The
  widely-referenced Commons leaderboard is **"Genesis Season"**, a vouch/airdrop
  game scored off GitHub stars, followers and X reputation
  (`https://api.commonsmade.com/game/events/genesis`, public). It **ended
  2026-08-24, three days before the hackathon opened**, and pays an airdrop, not a
  hackathon prize.
- **The hackathon's own board is token spend**, per the "Everything counts" rule,
  and it is the only public ranking surface on the event page ("Tokens spent"). It
  is still absent from the five judging criteria. It is an incentive, not a score -
  but it is no longer accurate to call it a different program's leaderboard.
- **The rules contradict the UI on entry count**: rules say "One build per
  builder", the picker says `enter as many as you like` and the code supports an
  array with add/remove. **Do not rely on multi-entry.**
- The Vault is a staking product: "$20K, 649% APY, and 80% of all hackathon revenue
  distributed to stakers" (`https://x.com/commonsmade`), with `/token-locks/deposits`,
  `/vault`, `/stake` routes and USDC payouts to match.

## 7. Unresolved oddities

- `https://api.commonsmade.com/hackathons/public/release` - the only open endpoint -
  returns `{"slug":"group-chats-2026","opensAt":"2026-08-27T15:00:00Z","status":"open"}`.
  The date matches this hackathon exactly; the slug does not match its name. Unexplained.
- An unverified Google-index paraphrase of the gated rules page says apps "must be
  open source". Not confirmable against any public source. Costs us nothing either
  way - xscapes is already MIT and public.

## The one thing nobody has verified

**Rules, criteria, prize split and judges are all server-overridable.** The client
merges a server `terms` object over the built-in defaults (`terms.rules`,
`terms.criteria`, `terms.judges`, `terms.prizeTotal`, `terms.payouts`). Everything
quoted above is the **shipped default**, which is what renders signed out.

Every authenticated endpoint returns `{"detail":"Missing Privy token"}`:
`/hackathons`, `/hackathons/{slug}`, `/hackathons/{slug}/board`,
`/hackathons/{slug}/submissions`. The Chrome the agent could reach was signed out
(`ep.auth_state=signed_out`; localStorage had `privy:connections` but no
`privy:token`), and switching to the other connected Chrome profile was blocked.

**ACTION FOR LUCAS: open `commonsmade.com/hackathons` signed in, click `rules ->`
and `criteria ->`, and confirm the text matches section 1 and section 5 above.**
Until that is done, treat this file as the default copy, not necessarily the copy
a participant sees.

## 8. Measured 2026-09-04: what a published Commons app can actually do

Playwright, signed out, against live apps. Supersedes the "inferred" parts of §3
(the platform is more capable than the bundle suggested) and one URL in §2.

- **Apps live at `https://<slug>.vibe.commonsmade.com`.** The `<slug>.commons.app`
  form in §2 does not resolve (ERR_NAME_NOT_RESOLVED). Server header `cloudflare`.
- **Apps have a server side.** `/api/*` routes answer GET and POST with
  `access-control-allow-origin: *` (WAITLAYER `/api/health` reports `hasSql:true`;
  Thinklings `/api/me`, `/api/time`). A missing route returns JSON 404
  `{"error":"unknown endpoint"}`, i.e. a Worker, not a static host.
- **`capability-probe.vibe.commonsmade.com`** (by "OUTIS", 2026-08-29) is another
  builder's honest platform test. Read live: shared storage PASS, "from Durable
  Object storage"; scheduled jobs: a 1-minute cron appending timestamps, ticking
  18:19-18:39 UTC while we watched; **realtime PASS, reproduced by us**: text typed
  in one Playwright page appeared in a second one within ~2 s over
  `wss://capability-probe.vibe.commonsmade.com/api/realtime`; runtime AI **FAIL**:
  "the x402 pay-per-request API requires a paid endpoint plus the
  COMMONS_X402_API_KEY secret which is not provisioned". So: Durable Objects,
  WebSockets, cron and SQL yes; free model calls from a deployed app no.
- **554 public apps.** Keyword scan of every stored prompt: 4 say "claude code"
  (all pasted specs, none a bridge); "Agent Bridge" is the one-line question "how do
  i connect commonsmade into my hermes agent or codex/claude cli?"; 32 say
  "terminal" (all in-page fakes); 23 "websocket". Two aim at browser extensions
  over claude.ai (Landed HUD, Cache Squirrel). **Nobody has wired a local coding
  agent into a Commons page.**
- **A web twin of xscapes is feasible on this platform**: a local hook shim POSTs
  the event protocol to `/api/event`, a Durable Object fans it out on
  `/api/realtime`, the page renders. The Go renderer compiles under
  `GOOS=js GOARCH=wasm` (`canvas`, `term`, `scape`, `companion`, `reduce`, `event`,
  `envx`, `notify`); a probe binary linking canvas+reduce+scape+term is 3.4 MB,
  0.96 MB gzipped. Only `internal/host` (pty, termios) cannot, so **"the agent runs
  inside the scape" has no browser equivalent** - the web version is a tab beside
  the terminal, the layout ruled out for the TUI on 2026-09-01.
- **Still unverified (login-gated; the Chrome extension was not connected):** file
  upload or paste limits in the builder, credits and per-model rates, and whether
  the builder can be told to write the Durable Object + WebSocket worker (Capability
  Probe is evidence it can). An endpoint-enumeration probe for the source of
  `sourceVisible` apps was blocked by tool policy and not retried.
