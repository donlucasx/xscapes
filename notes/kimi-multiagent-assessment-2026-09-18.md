# Multi-agent assessment of the dirty tree (sessions 40–41 work)

2026-09-18, run by the Kimi session at the human's word ("Run a multi agent
assessment on your end about the xscapes build for kimi while I watch you work
and test for the latest flicker fix"). Five parallel read-only assessors:
build & suite health · Kimi adapter correctness · flicker fix ·
instrumentation & replay · docs & consistency. Nothing was modified.

Scope: the uncommitted working tree (hook.go, install_agents.go, eventlog.go,
kimilog_replay_test.go, internal/scenes/forest.go wind + unlooped clock,
vista.go tall-window fixes, frames.go, live.go, main.go, inside.go, README +2).

## Verdict

The tree builds, vets, formats, and passes the full suite — uncached, under
`-race`, and under a scrubbed environment. Root: 113 tests, 0 failures, no
flakes across three runs. The Kimi adapter is sound against every measured
payload shape, and the flicker fix is correct at all widths. The
blank-top-rows defect is confirmed NOT to be the painter; two hosted-path
mechanisms fit the symptom (below). The most actionable findings are one
verified test panic, one `KIMI_CODE_HOME` inconsistency, and the
`XSCAPES_HOOKLOG=1` footgun written into our own session note.

---

## 1. The blank top rows — the live suspect

Painter-side is ruled out: `paintVistaL` always vramps rows 0..lakeTop-1
(`internal/scenes/forest.go:543-550`), Render always emits full rows
(`internal/canvas/canvas.go:530-553`), and `Vista.Update`'s `c.Clear()`
(`internal/scenes/vista.go:75`) clears only glyph layers in memory. No code
path renders blank top rows.

**Mechanism 1 (primary suspect): unfiltered agent CSI + damage-tracker
persistence on the vista's most static rows.**

- The host filter strips ONLY DECSTBM. ED (`ESC[J`/`ESC[2J`), DECOM
  (`ESC[?6l`), RIS, SU/SD, IL/DL all pass through
  (`internal/host/filter.go:59-62`). ED erases from the cursor to end of
  screen regardless of the scroll region on xterm-likes and Terminal.app.
- The band design was measured against Claude Code's emissions
  (notes/claude-terminal-emissions.md); Kimi's emission profile has never been
  audited the same way. "Kimi's bytes were clean" (_FEEDBACK.md:4572) was
  established on a different window of a different run — it is not conclusive
  for the 18:57 occurrence.
- Amplifier: `refreshEvery = 50` frames (`internal/host/damage.go:28`) ≈ 4.2 s
  at the hosted 12 fps. The vista's top rows are the most static in the frame
  (leaves only from row 4 down, stars seed-fixed, arc moon slow, token counter
  changes rarely), so any untracked erase of rows 0–3 persists up to ~4
  seconds — exactly long enough to be photographed.

**Mechanism 2 (known, still open): reallocBand.** On Terminal.app width
changes it deletes the entire scape band via DL
(`internal/host/band_control.go:343-356`, called at `host.go:451-453`) →
whole-scape blank until the same tick's paint, and it ends with `regionReset`
WITHOUT re-pinning DECSTBM — between that write and the next paint's
`EndPaint` the region is the full screen and any agent scroll scrolls
everything, scape included. Flagged in the s38 audit; confirmed still open.

**Instrument blind spot:** `TestFlickerProbe` drives `fr.frame(now)` directly
(flickerprobe_test.go:131) — no host, no damage tracker, no band, no agent
byte stream; `kimilog_replay_test.go` replays hook events, likewise. The
entire replay suite is structurally blind to the paint-path/erase class of
defect. The planned `XSCAPES_TRACE=1` run is the right instrument.

**Checked and cleared:** paint/agent/mirror writes all serialized through
`h.write` under `h.mu` (host.go:108-122), one write per frame (host.go:473-482)
— no interleaved escape splitting. Resize sequences re-pin before `p.SetSize`
and end with `dmg.reset()`.

**Next live run:** `XSCAPES_TRACE=1 XSCAPES_HOOKLOG=<path>
XSCAPES_EVENTLOG=<path> xscapes kimi`; if blank rows appear, grep the trace
for `ESC[…J`, `ESC[?6l`, `ESC c` in AGENT chunks, then replay through the
internal/host screen model and diff model scape rows against the screenshot —
that discriminates "never sent" from "erased after sending".

**Cheap hardening, independent of the trace:**
- Count/flag ED (and DECOM resets) in the agent stream — log first, so the
  next occurrence is attributed rather than silently altered.
- reallocBand: append `EnterBand(agentRows)` inside the same write to close
  the unpinned-region race.
- Consider `refreshEvery` 50 → ~12: an untracked erase of static sky rows
  heals in 1 s instead of 4.2 s; cost ≈ 30 KB/s against the 0.6 MB/s the
  tracker avoids — negligible, and it caps the screenshot-ability of this
  whole defect class.

---

## 2. Flicker fix audit (forest.go wind, WindPick, vista tall-window)

**(a) Unlooped clock: correct at all widths, no residual periodic jump.**
`shift = int(math.Round(tRaw*speed)) % c.W` (forest.go:897-899), field sampled
at `((x-shift)%c.W+c.W)%c.W` (forest.go:904) — monotonic shift, seamless wrap
at any width. Empirical: `TestTheWindsChurn` at 131 wide shows worst air frame
85 vs mean 71.6 at level 1 (a seam would be hundreds). The looped clock's
remaining users are all wrap-safe (flames %24, smoke %12, sparks %24 — frame
reaches 24 exactly as t wraps; scrub flutter is period-2 in a 4-s loop).
Caveat: the study path (`live == nil`) keeps the looped clock and still jumps
at widths ≠ 80 — intentional, sha256-guarded at 80x24.

**(b) W1/W2/W3 churn claims: confirmed in substance, two numbers ~5% off.**
Measured via `TestTheWindsChurn` (131x24, level 1.0, night, 6 s @ 12 fps):

| style | measured cells/frame | claimed (CLAUDE.md) |
|---|---|---|
| W0 | 209.2 | ~216 |
| W1 | 84.1 | 88 |
| W2 | 65.1 | 65 ✓ |
| W3 | 53.1 | 55 |

Wiring: `XSCAPES_WIND=n` (main.go:57-62); an out-of-range n silently falls
back to W0 (the switch at forest.go:887 has no default) — safe but silent.

**(c) Tall-window fixes:** present and tested (pines follow the layout
forest.go:712-721; treeline dip guarded `2*owlY < nearBase` forest.go:648-661;
vista_tall_test.go all passing). `TestTheStudyVistaIsUnchanged` holds.

Minor: the `-live -level` pin lives only in the `f.red == nil` branch
(frames.go:187-189); if a session binds later the pin silently drops. Fine
for the A/B as used.

---

## 3. Kimi adapter correctness

Sound overall: payload shapes (array prompts, object errors, profile-named
subagents, TodoList, AskUserQuestion ask+answer, Esc→Interrupt, Stop
attribution) correct against the fixture; no crash paths; the hook cannot
damage the agent's turn (250 ms watchdog + panic recover + exit-0-always,
hook.go:251-258).

**F1 (medium) — Installer and launcher ignore `KIMI_CODE_HOME`; spend honors
it.** install_agents.go:83 hardcodes `~/.kimi-code/config.toml`;
`internal/spend/kimi.go:36-45` resolves `$KIMI_CODE_HOME` first, and
_FEEDBACK.md:4187 records that the variable relocates everything. With it set:
install writes hooks into the file Kimi never reads, hooks never fire, yet
`hooksInstalled("kimi")` (install_agents.go:235-262) checks that same wrong
file and PASSES — so `xscapes kimi` starts half-blind, exactly the failure the
refusal was built to prevent. Fix: route the adapter's path through
`spend.KimiHome()` or a shared helper.

**F2 (medium-low) — A main-turn `Stop` dropped while a background subagent is
open is never re-emitted; the comment claims otherwise.** reduce.go:296-299
drops `Done` outright when `kimiOpen` is non-empty; the `SubEnd` case
(reduce.go:473-489) never synthesizes it, and Kimi's `task.completed`
notification is deliberately nil (hook.go:485-491). The comment at
reduce.go:269-270 ("the cue rings when the last of them is in") is not
implemented. Foreground agents are safe (Kimi fires the main Stop last —
fixture lines 20-28); for detached agents the done-knock never fires and the
turn settles via `TurnSilence` (30 min). Fix the code or the comment.

**F3 (low) — Foreground-Agent attribution heuristic can misattribute
interleaved main-thread tool calls.** reduce.go:300-308 stamps any agent-less
ToolStart/ToolEnd/Error while an `OpSub` flight is open as
`Agent: "kimi-subagent"`. A main-thread batched `Agent` + `Read` would
mis-attribute the Read (no pacing, no worry). Inference from fixture behavior,
unmeasured with a mixed batch.

**F4 (low) — `StopFailure` doesn't check `is_interrupt`.** hook.go:513-518
maps it unconditionally to `event.Error`; `PostToolUseFailure` checks
`p.IsInterrupt` first (hook.go:459-461). If any Kimi path emits StopFailure
with is_interrupt, an Esc becomes a worried owl — the exact mis-signal the
Interrupt ruling exists to prevent. One-line defensive fix.

**F5 (low) — Refusal edge cases:**
- Hand-written (working) hooks are refused by installer AND launcher
  (hooksInstalled counts only marked-block entries, install_agents.go:252-261)
  — a real false-positive of "hooks are not installed"; the error explains it.
- `foreignHook` regex scans comments — a config comment mentioning "xscapes
  hook" falsely refuses install (install_agents.go:331).
- `kimiInlineHooks` `^\s*hooks\s*=` matches under any TOML table, not just
  top-level (install_agents.go:324).
- `hooksInstalled` passes on n > 0: a hand-trimmed marker block with 1 of 15
  events counts as installed (fail-open for a partial install).
- `--config <path>` installs are invisible to the launcher, which checks only
  the default path.

**F6 (low) — The fixture test partially reimplements the code under test.**
kimi_test.go:49-54 hand-folds ToolCallID/AgentName and calls `translate(p)`
directly, bypassing `hookTranslate` — a regression in the production folding
would pass the test. Have the fixture loop call
`hookTranslate([]string{p.Event, "kimi"}, line)`.

Notes, not defects: duplicate SubagentStart mints two instance keys (one
stranded until SubStale 60 min — the foreign-hook refusal is the realistic
guard); `KimiSessionDir` rescans all of session_index.jsonl on EVERY hook
invocation (hook.go:347-349) — O(index) per tool call, bounded by the
watchdog, grows with session count; unmatched SubagentStop still increments
tasksDone (reduce.go:474-475); `stop_hook_active` is not special-cased
(unobserved in the wild).

---

## 4. Instrumentation & replay

The logs are well-built and sufficient for the lost-events question (every hop
observable: hook receipt → translate → emit → bus → reducer). For the
blank-top-rows defect they record nothing frame-level — by design; the byte
trace is the frame-level instrument, and RESUME.md:96-100's combined command
line is the right one.

**F1 (verified panic) — `TestReplayAHookLog` crashes on a log with no payload
lines.** kimilog_replay_test.go:79 does `all[len(all)-1]` with no empty guard.
Reproduced read-only: pointing it at `testdata/kimi/session.jsonl` (raw
payloads, no `payload` key) panics. A truncated or outcome-only log hits the
same path. The sibling probe has the guard (flickerprobe_test.go:70-72); this
file doesn't. One-line fix.

**F2 — The replay test asserts nothing.** It is print-only
(kimilog_replay_test.go:59, 94); the "103 emitted == 103 applied" result was
produced by hand-comparing the two logs, not by code. If the next live run
loses an owlet again, the comparison is again manual. Add a cross-log matcher
(hook-log outcome lines × event-log applied lines by kind+id+session).

**F3 — Not hermetic:**
- Scape preference can un-pin the vista mid-test and nil-deref: the test
  mutates `fr.scapeName`/`fr.vista` (kimilog_replay_test.go:73-74) but
  `refreshScape` (frames.go:297) re-reads `scapePref()` on the first frame;
  with "shore" saved or `XSCAPES_SCAPE` set, `fr.vista.Layout()` nil-derefs.
  Fix: `t.Setenv("XSCAPES_SCAPE", "vista")` instead of the mutation.
- Sounds: the test does not set `XSCAPES_SILENT`; replaying a session with
  real asks can exec afplay during `go test`. flickerprobe sets it
  (flickerprobe_test.go:34); mirror that.
- Spend: the replay reads real `~/.kimi-code` session dirs via transcript
  resolution (frames.go:216-225, hook.go:347-349). Read-only, but
  machine-dependent.

**F4 — `XSCAPES_HOOKLOG=1` writes a file literally named "1" into the cwd of
every hook firing — and our own note recommends that form.** runHook treats
the value purely as a path (hook.go:268-271); unlike TRACE
(host.go:170-177), HOOKLOG/EVENTLOG have no `1`/`true` handling. The hook's
cwd is the agent's cwd — the user's repo — so the broken form litters the
worked-on project. notes/kimi-live-test-2026-09-17.md:43 says
"XSCAPES_HOOKLOG=1"; README.md:218 and RESUME.md:96 have it right. Fix the
note or add the TRACE-style special-case.

**F5 — EVENTLOG misses two event sources, contradicting its doc.**
eventlog.go:12 claims "every event its reducer applies", but watch-synth
events applied at inside.go:183-187 bypass the logging drain (frames.go:195-205)
and the bind moment itself (inside.go:192-199) is logged nowhere. Also the
logged `ts` is frame-drain time, not the wire `e.TS` — fine for ordering,
lossy for timing correlation with the hook log.

**PII:** none by default (both logs env-gated, 0600 under 0700, errors
swallowed). When enabled, the hook log stores raw payloads VERBATIM — full
prompts, cwd, transcript paths, full shell commands; the `program()` secret
scrubbing applies only to emitted events, not the log. One line in README
would let a user know before pasting it into an issue.

Confirmed sound: outcome lines carry no payload (pinned by kimi_test.go:279-298);
replay and runHook share `hookTranslate`; env propagation to hosted hooks via
`os.Environ()` (inside.go:203); bus drop/bad counters logged on change.

---

## 5. Build & suite health

Green everywhere measured: `go build ./...`, `go vet ./...`, `gofmt -l .`
(including notes/), `go test ./...` (uncached 13 packages ok), `-race` on the
root/event/host packages, and `env -i` scrubbed runs. ~130 s uncached total;
internal/scape alone is 81-84 s (brute-force star property tests — inherent,
not a defect).

**F1 (medium) — A test that can never run in a default environment.**
`TestWithXscapesNearUnsetTheCompanionKeepsItsShippedBox` (near_test.go:545)
skips whenever `companion.Near != 0`, but unset XSCAPES_NEAR maps to rung 2 by
design since 2026-09-11 (nearDefault = 2, internal/companion/crab_near.go:91-96).
It passes only under explicit XSCAPES_NEAR=0 — the opposite of what its name
and doc claim. Silent coverage loss. Fix: set `companion.Near = 0` with a
deferred restore, or retarget to assert the rung-2 default like
crab_near_test.go:161-165.

**F2 (medium, portability) — installsh_test.go fails where Go lives in
/usr/bin.** installsh_test.go:94-96 Fatalfs if `PATH=/usr/bin:/bin command -v
go` succeeds — i.e. any stock distro-packaged-Go Linux CI box. Passes here
because Go is in homebrew. Downgrade to skip or strip go's directory from the
restricted PATH.

**F3 (low) — Tests execute whatever `kimi`/`hermes` binaries are on PATH**
(install_agents_test.go:89-101, 191-209). The kimi assertion keys on "OK" in
`kimi doctor` output — a future kimi version changing that string breaks the
suite on this machine specifically.

**F4 (low) — XSCAPES_HOOKLOG/EVENTLOG are not scrubbed in TestMain.**
main_test.go:17-25 and internal/host/main_test.go:23-31 scrub HOME and TRACE
(both after real incidents); the new knobs aren't scrubbed. Verified safe
TODAY (no test passes a live bus to follow()), one future test away from
writing into the real config dir. Add the two Unsetenv lines for parity.

**F5 (low) — Global mutation without defer.** bubble_aim_test.go:100-101 sets
`companion.Near` and restores at line 124 without defer; a Fatalf in between
leaks rung≠0 into later tests.

Info: tty-gated tests skip in CI (no controlling terminal); 17 root skips are
deliberate env-gated artifact writers; adapter_test.go:147's 50 ms quiet
drain is a theoretical flake source, not observed.

---

## 6. Docs & consistency

No public-facing doc claims anything the code doesn't do. Every CLAUDE.md
session-40/41 claim spot-checked against the tree held: launcher refusal
(inside.go:85-90), WindPick styles (forest.go:1208-1223), unlooped clock
(forest.go:521, 905-907), tall-window fixes, eventlog, replay test, README's
Kimi paragraph clause-by-clause. "INSTALLED dirty (inode 87217050)" verified —
`ls -i` = 87217050, `vcs.revision=d880e40`, `vcs.modified=true`.

**F1 (low-medium) — The launcher refusal is a user-visible behavior change
with no doc mention.** `xscapes claude|kimi|hermes` now exits 2 without
installed hooks; README and the site never say so. The refusal message offers
`xscapes inside <agent>` but not the closer `-watch=on` bypass the code honors
(inside.go:85). Add a README line and mention `-watch=on` in the message
(install_agents.go:278).

**F2 (low)** — notes/kimi-live-test-2026-09-17.md:43's `XSCAPES_HOOKLOG=1`
instruction is wrong (see §4 F4).

**F3 (low)** — README.md:219's "every event its reducer applied" overclaims;
watch-synth events are never logged (see §4 F5).

**F4 (info)** — `XSCAPES_WIND` and `-live -level` are shipped behavior no user
doc names (arguably deliberate — A/B instruments).

**F5 (info)** — `xscapes install kimi` dies if `~/.kimi-code/config.toml`
doesn't exist ("run `kimi` once, then install"); README's quickstart doesn't
mention the precondition. The error is actionable, impact nil.

**F6 (info)** — site/install.sh prints claude-only next steps while README and
the site present kimi/hermes as equals. Pinned by installsh_test.go:123,151 —
deliberate, but a Kimi user's first terminal shows the wrong agent's commands.

**F7 (trivial)** — RESUME.md:110 cites the wind leaves at forest.go:852-871;
after the dirty rewrite that block is ~869-928.

**F8 (release skew)** — README's +2 lines (HOOKLOG/EVENTLOG) document
unreleased code; today there is no public mismatch (GitHub serves the
committed README, install.sh serves v0.4.3), but once committed, main's README
will document knobs absent from the latest release until v0.4.4. Keep the
commit and the tag together, as RESUME already sequences.

---

## Ranked fix list (suggested order)

1. kimilog_replay_test.go:79 empty-guard panic — one line, verified crash.
2. `KIMI_CODE_HOME`: route install/launcher through `spend.KimiHome()`.
3. Fix notes/kimi-live-test-2026-09-17.md:43 (`XSCAPES_HOOKLOG=1` → a path),
   or add TRACE-style `1`/`true` handling to HOOKLOG/EVENTLOG.
4. Replay hermeticity: `t.Setenv("XSCAPES_SCAPE","vista")` +
   `t.Setenv("XSCAPES_SILENT","1")`.
5. `StopFailure` is_interrupt check (hook.go:513-518).
6. Dead test: near_test.go:545 guard stale since the rung-2 default flip.
7. TestMain: scrub XSCAPES_HOOKLOG/EVENTLOG.
8. Hosted-path hardening for blank top rows: log ED/DECOM in the agent stream;
   re-pin DECSTBM inside reallocBand's write; consider refreshEvery 50→12.
9. reduce.go: implement the deferred `Done` on last SubEnd, or fix the
   comment at reduce.go:269-270.
10. kimi_test.go fixture loop through `hookTranslate` instead of hand-folding.
11. Cross-log matcher to make 103==103 repeatable on the next live run.
12. README: the launcher refusal + `-watch=on` escape; raw-payload note on
    HOOKLOG; "run kimi once" precondition.

## Not verified (honest limits)

- What Kimi 0.39 actually emits on redraw (ED/DECOM/RIS) — the crux of
  mechanism 1; answerable only from a traced live run.
- Whether Kimi fires a trailing Stop after detached/background agents (§3 F2's
  real-world impact).
- Subagent event shapes beyond the one captured session (agent_id vs profile
  name, instance ids on tool events). XSCAPES_HOOKLOG on the next run answers
  all of these.
- Real-tty behavior: tty-gated tests skipped in the harness shell, not failed.
- Non-darwin platforms: everything ran on darwin/arm64.
- Mutation-checking of the tall-window tests: they pass; mutation status is
  process, not tree content.
