# Scrollback plan audit — 2026-09-03, session 14

**Charge** (his words, mid-turn): *"audit the scrollback plan before you build it- ensure its the
right path. Run a kimi audit if needed as well."* The plan under audit is RESUME.md ▶ NEXT #1 as
written at the session 13 wrap: xscapes owns its own scrollback — promote the `screen` model,
tee the agent's bytes through it, ring-buffer the rows that leave the band, a scroll mode that
renders the ring and exits cleanly, mouse-wheel reporting, tests. Estimated ~1 day keyboard-only,
1.5–2 days with the wheel, two risks unsized.

**Verdict: do not build the plan as written.** Its premise is right (measured today: the
alternate screen has no history), but the mechanism it chose is the most expensive one available,
and it would have shipped on top of a resize defect that reproduces on every shrink. A cheaper
path exists that puts the transcript into the scrollback the user already has — Terminal.app's
own — with no viewer, no key handling and no mouse reporting. Measured, not reasoned; details
below. The revised plan is at the end.

Everything here was read back over AppleScript (`history of selected tab`), not by eye. That
instrument returns the alternate screen's visible cells too, which retires the session 13 line
"nothing reads cells back from Terminal.app".

## The five probes

All in `notes/`, all run in fresh Terminal.app windows opened and resized by `osascript`, at
120x30 (band 1..17, scape 18..30) unless stated. Outputs in the session scratchpad.

### 1. `histprobe` — does the alternate screen have history? **NO. Premise holds.**

Forty numbered lines written through a row-1 DECSTBM band of 17 rows, 24 of them scrolling out of
the top. Then `history of tab`.

| screen | in history while up | in history after exit |
|---|---|---|
| ALTERNATE (production) | lines **25–40 only** (the 16 still visible) | none |
| MAIN (control) | **all 40** | 1–25 (the rest sat below the shell prompt, whose redraw opens with `ESC[J`; Kimi F6) |

So a line that leaves the band on the alternate screen is gone. Session 13 was right about that,
and it had never been measured on this screen.

### 2. `shrinkprobe` — what a shrink does to the agent's cursor. **REPORT 1 REPRODUCED.**

Production configuration through the host's real `Rebind`. A three-row input box on the band's
last rows, cursor in it on row 16 (the way Claude Code leaves it), then the window shrunk 30→24
from outside (band 17→14), then a Claude-style RELATIVE redraw of the box (`\r ESC[1A ESC[2K …`).
Cursor read with DSR at each step; content with `history`.

| | before | after the terminal's shrink | after the host's Rebind | after the redraw |
|---|---|---|---|---|
| cursor row, **production today** | 16 | **16** (the cursor does not move) | **1** | 2 |
| where the box ended up | rows 15–17 | rows 9–11 (content moved UP 6) | | **a second box at rows 1–3, the old one still at 9–11** |

That is the split input box he photographed on 2026-09-03 ~11:57 and nobody could reproduce.
The mechanism: the terminal moves the CONTENT up by the shrink and leaves the CURSOR where it was
(measured in s13, both facts); `Rebind` then saves that cursor, re-pins the band to its new,
smaller height, and restores — and **DECRC into a region that no longer contains the saved row
lands on row 1** (new measurement). Claude Code's next relative draw then paints its input box at
the top of the band over the transcript. In an established session the cursor is always in the
last rows of the band, so this fires on **every shrink**: a one- or two-row tick leaves the box
offset by that many rows (measured, `none-bot-1` below), a larger one throws it to row 1, and a
drag is many small ticks. Kimi's DECSC/DECRC-slot hypothesis is not needed to explain Report 1.

Two corrections were tried; the first attempt at both FAILED because it ran after `Rebind` had
already homed the cursor (it saved row 1 and faithfully restored it). The fix has to change
Rebind's own order on a shrink: **restore while the region is still the full screen, then move
relatively, then save again, then pin the band, then restore into a band that contains the row.**

| variant | sequence after the terminal's shrink | cursor after | box after redraw |
|---|---|---|---|
| `cuu` | `ESC7 ?6l ESC[r ESC[0m` · clear the rows that slid up · `ESC8` · `CUU shrink` · `ESC7` · band · `ESC8` | **10** | drawn exactly over the moved box (rows 9–11); 3 blank rows under it |
| `sd` **(recommended)** | `ESC7 ?6l ESC[r ESC[0m` · `SD (shrink − bandShrink)` · `ESC8` · `CUU bandShrink` · `ESC7` · band · `ESC8` | **13** | drawn exactly over the moved box, now at rows 12–14 = the band's last rows; blank rows above |

`sd` keeps the input box on the band's last row, which is where Claude Code believes it is, and
needs no clear inside the band at all: the rows above the transcript are freshly inserted blanks
(SGR reset first — SD fills with the current background, same BCE lesson as the erase). This is
independent of scrollback and must ship regardless. The screen model must change with it:
`resizeAlt` moves the cursor with the content in both directions today, which is the MAIN-screen
rule, so no test can see this — the same class of blind spot as session 13's.

### 3. `mirrorprobe` — can rows be written into the main buffer's history while the alternate screen stays up? **YES.**

Around each row: `ESC7 ?6l ESC[r` · `ESC[?47l` (to the main buffer, no clear) · `CUP(h,1)` · the
row · `\r\n` · `ESC[?47h` (back, no clear) · band · `ESC8`.

| run | alternate screen after | rows in history above the alt screen |
|---|---|---|
| 30 rows, 40ms apart | all 17 band rows intact, cursor where it was | **30 of 30, in order** |
| **400 rows, 5ms apart** | all 17 intact | **400 of 400, in order** |

`history` while the alternate screen is up lists the main buffer's rows and THEN the alternate
screen's rows — which is how Terminal.app's view is built: the alternate screen sits at the bottom
of the same scroll view, and scrolling up shows the main buffer above it. That is also what he
described in session 13 (*"scrolling up through the text history, I reach a black point"*): the
"black point" is the main buffer's 27 blank rows sitting between his shell prompt and the band,
not cleared rows in scrollback — alternate-screen rows never reach scrollback (probe 1). The SGR
reset before every erase stays (BCE paints VISIBLE rows), but session 13's mechanism for the wall
of black was wrong.

One unique `screencapture` mid-burst (two of the three are byte-identical; Kimi F7) shows the
alternate screen, never the main buffer. One sample proves nothing about flicker; **his eyes are
the instrument for that**
(`go run ./notes/mirrorprobe -n 400 -gap 5ms` in a Terminal.app window and watch).

⚠ **After `ESC[?1049l` Terminal.app keeps only a few of the mirrored rows** (4 of 30 in the run;
it restores the main screen it saved at `1049h` and discards most of what scrolled meanwhile). So
the mirrored history is complete DURING the session and mostly gone after it. Cheap mitigation:
the host keeps the rows it mirrored and prints them once more on Close, after leaving the
alternate screen, so post-session review survives too. Gate that on `TERM_PROGRAM=Apple_Terminal`:
an xterm-like terminal keeps the main scrollback intact across the alternate screen and a replay
there would duplicate every line.

### 4. `wheelprobe` — what the wheel sends on the alternate screen. **First run VOID; re-measured after he granted Accessibility: 0 bytes without mouse reporting, SGR mouse events with it.** Terminal.app scrolls its own view. (Original text kept below.)

A Swift tool posted synthetic wheel events at the window; the probe received nothing — and the
positive control (SGR mouse reporting ON) received nothing either, so the events never reached
the window (CGEvent posting needs Accessibility trust). No measurement was made. It no longer
matters for the recommendation: with mirroring, scrolling is Terminal.app's own view scroll,
which he has already used.

### 5. The instrument upgrade

`history of selected tab of window id N` over `osascript` returns the alternate screen's visible
cells. Every session 13 content measurement was read by eye because this was believed impossible.
`notes/histprobe`, `shrinkprobe` and `mirrorprobe` all use it; `screencapture -l <AppleScript
window id>` captures the window as pixels. `set number of rows of window` resizes from outside.

## Findings, ranked

1. **The plan would have shipped on a live resize defect.** Every shrink over two rows homes the
   agent's cursor and Claude Code redraws its input box at the top of the band. Reproduced
   deterministically; fix verified in two variants; `sd` recommended. This is Report 1, closed.
2. **The plan chose the most expensive mechanism.** Its viewer, its key handling, its
   mouse-wheel reporting and its biggest risk ("exiting scroll mode must repaint the band
   perfectly") all exist to replace the terminal's scrollback. Mode-47 mirroring feeds the
   terminal's scrollback instead: no viewer, no keys, no repaint on exit, and the user's existing
   gestures (wheel, selection, Cmd-F, Cmd-K) work on it. What the plan got right and mirroring
   still needs: the promoted `screen` model with full SGR, the tee, and the capture of rows that
   leave the band.
3. **The plan's half-drawn-frame risk is smaller than stated.** A row leaves the band only by
   scrolling out of the top — a committed transcript row, older than everything visible — or by a
   shrink destroying the top rows, which the model still holds and can mirror at that moment. The
   host's own SU on a grow pushes off the blank rows the terminal inserted; those are not content
   and are not mirrored. The remaining fidelity risk is the model itself: wide glyphs (emoji, box
   drawing at double width) advance two cells in the terminal and one in the model, which garbles a
   mirrored line's alignment, not the band. Acceptable for v1; a width table is the fix.
4. **What is still unmeasured**: visible flicker on the round trip (his call, one minute with the
   probe); behaviour on iTerm2 / kitty / Ghostty, where the alternate screen hides the scrollback
   until exit, so mirrored rows appear AFTER the session rather than during it — still correct,
   just later; tmux, which emulates 47 itself. Terminal.app is the target by his 2026-09-02 ruling,
   so these are notes, not blockers.
5. **The DECSC/DECRC slot** (Kimi, s13): the traces show the agent emitting one or two save/restore
   pairs per session, at startup. A collision needs a host frame to land between them, across a
   pty read boundary. Possible, rare, startup-only; it cannot produce Report 1 on an established
   session, which finding 1 does. Keep it on the list; do not build for it.

## The revised plan

0. **Shrink fix** in `Rebind` (`sd` variant, alternate screen only; the main screen keeps `drop`).
   Model: `resizeAlt` stops moving the cursor; DECRC outside the region homes. A test that draws a
   box at the band bottom, shrinks, rebinds and redraws relatively, asserting one box, not two.
   Red today. ~2 hours.
1. **Promote `screen`** out of `_test.go`, cells carry fg + bg + attributes, serialise a row back
   to ANSI. **Tee** the agent's filtered bytes through it. ~3 hours.
2. **Mirror.** `Open` asks DSR for the shell's cursor row before the child starts (the host still
   owns stdin then), so mirrored rows begin right under the command he typed with no blank gap.
   Rows leaving the band's top (scroll at the bottom margin; the top rows a shrink destroys) are
   appended to the main buffer in one batched write per tick using the sequence measured in probe
   3, SGR reset first. Grow pushes are not mirrored. ~3 hours.
3. **Close replay** on `Apple_Terminal` only. ~1 hour.
4. **Tests**: the model's scrolled-off rows equal the rows that were in the band; the trace replay
   (`XSCAPES_TRACE`) shows the mirrored rows in the main buffer of the model in order; the shrink
   test above. ~2 hours.

About a day, like before, with the risk profile inverted: nothing here repaints the agent's band,
so a model error garbles a history line rather than the live UI.

**Waiting on him**: watch `mirrorprobe` once for flicker, and a go on this plan in place of the
one in RESUME.md. Then the Kimi packet (staged in `~/Documents/kimi/xscapes-scrollback-audit/`)
gets a verdict on this document before a line is built.

## Kimi round 1 — 2026-09-03, `~/Documents/kimi/xscapes-scrollback-audit/KIMI-REPORT.md`

**`VERDICT: REVISE`** — *"the audit's direction (ship the shrink fix first, get scrollback by
mode-47 mirroring instead of owning a viewer) is the right path and every load-bearing measurement
checks out against the raw outputs"*, with four gaps to close before building. Dispositions:

| # | finding | disposition |
|---|---|---|
| F1 MAJOR | the `sd` fix was measured in ONE geometry | **CLOSED by measurement**, matrix below |
| F2 MAJOR | the model aliases 47 with 1049 (one saved-buffer slot), moves the cursor on `resizeAlt` against the measured fact, and restores DECRC without the region check; step 4's test cannot pass as written | **ACCEPTED**: the two-buffer model, the cursor rule and the DECRC clamp are budgeted into steps 0 and 1 |
| F3 MAJOR | the DSR-at-Open premise is wrong about today's code: the child is started BEFORE Open (`host.go`), so a host read of stdin could eat Claude's startup replies | **ACCEPTED**: the DSR moves BEFORE `Cmd.Start()`; it is a main-screen fact and belongs before `1049h` anyway |
| F4 MAJOR | the forwarder goroutine is never joined; the agent's last bytes can land after `1049l`, and a Close replay gives them something to corrupt | **ACCEPTED**: drain the pty to EOF with a deadline before the replay |
| F5 MINOR | mirrored rows reflow on a narrowing drag (ordinary scrollback behaviour, say so); UNMEASURED whether Terminal.app snaps the view to the bottom when a batch lands while he is scrolled up | caveat added; the snap is **his eyes** (scroll up during `mirrorprobe`), synthetic keys are blocked on this machine |
| F6 MINOR | histprobe's MAIN after-exit cell said "all 40"; the output shows 1–25 | **corrected** here and in the notes file |
| F7 MINOR | one unique screenshot, not two | **corrected**; the eyeball gate stays |
| F8 MINOR | `1049l` keeps the OLDEST few mirrored rows, so a Close replay duplicates them | accepted as cosmetic; a separator line before the replay |
| F9 NOTE | wheelprobe's void has no artifacts in the packet | noted; keep probe stderr next time |
| F10 NOTE | the plan displaces his session-14 pick (the DECSC/DECRC interleave test first) without saying so; mirroring adds a high-frequency DECSC writer | **said here**: the shrink fix replaces that first step because it reproduces the photographed defect and the slot cannot; if a split box is photographed AFTER the fix, the slot is next, and the hardening is to restore with an absolute CUP from the model instead of DECRC |
| F11 NOTE | ~1 day is really ~1.5 | accepted: **1.5 days**, against 1.5–2 for the ring buffer, with the biggest risk gone |

### The shrink matrix (F1) — eleven geometries, all through the real `Rebind` / the `sd` sequence

`notes/shrinkprobe` now takes `-cursor top|mid|bottom` and `-resizes N`; a grow always takes
production's own Rebind. Window 120 wide, resized from outside; box rows read back from `history`
anchored on the probe's status row. "overlay" means the relative redraw landed exactly on the
moved box and no row of the old one survived.

| case | window | band | cursor before → after resize → after fix | result |
|---|---|---|---|---|
| sd, bottom, −1 | 30→29 | 17→16 | 16 → 16 → 15 | overlay |
| sd, bottom, −2 | 30→28 | 17→16 | 16 → 16 → 15 | overlay |
| sd, bottom, −3 | 30→27 | 17→15 | 16 → 16 → 14 | overlay |
| sd, bottom, −6 | 30→24 | 17→14 | 16 → 16 → 13 | overlay |
| sd, bottom, −20 (past the band) | 30→10 | 17→10 | 16 → **10** (terminal clamped) → 3 | the old box was destroyed by the terminal; the new one is drawn at rows 2–4 over the blank rows SD inserted |
| sd, top, −6 | 30→24 | 17→14 | 2 → 2 → 1 | old box destroyed (top 6 rows); new box at rows 1–3 over inserted blanks |
| sd, mid, −6 | 30→24 | 17→14 | 9 → 9 → 6 | overlay |
| sd, bottom, −6 then +6 | 30→24→30 | 17→14→17 | 16 → 13; grow: 13 → 13 | overlay; the box sits 3 rows above the band's bottom after the grow, with the cleared rows under it |
| sd, bottom, −1 with `bandShrink = 0` | 32→31 | 18→18 | 17 → 17 → 17 | overlay (SD by 1 undoes the whole slide) |
| **none (production), bottom, −1** | 30→29 | 17→16 | 16 → 16 → 16 | **split by one row**: the old top border survives above the new box |
| sd, bottom, drag as −2 −2 −2 | 30→28→26→24 | 17→16→15→14 | 16 → 15 → 14 → 13 | overlay |

So: production misplaces the box on ANY shrink, by the tick size for small ticks and to row 1 for
large ones; the `sd` sequence keeps the box on its content in every geometry where the content
survived, and puts it on blank rows where the terminal destroyed the content, which is the best
any host can do without a model.

### What is still unmeasured, and who measures it

- Flicker on the 47 round trip: **him**, `go run ./notes/mirrorprobe -n 400 -gap 5ms`.
- ~~Whether the view snaps to the bottom when a batch lands while scrolled up~~ **MEASURED, both
  ways**: he scrolled up during the burst and reported *"stable when I scroll up"*; then, with
  Accessibility granted, twelve synthetic wheel ticks during a 300-row burst at 50ms and two
  `screencapture`s five seconds apart (a hundred rows landed between them) are byte-identical.
  ⚠ That pair was VOID: the synthetic wheel never scrolls Terminal.app's view (the main-screen
  control did not move either; wheel ticks reach only the mouse-reporting path). Re-measured with
  keystrokes once Accessibility was granted: **Shift+Page Up scrolls the view over the alternate
  screen into the mirrored rows** (plain Page Up is delivered to the program as `ESC[5~`), and
  two `screencapture`s five seconds apart during a 300-row burst at 50ms show MIRROR LINE 11–40
  at the same rows, only the scrollbar thumb moved. The view holds under a landing batch.
  Flicker: *"not sure if it flashed"* at 200 switches a second; production does at most twelve.
  **Gate passed 2026-09-03; step 2 is ungated.**

## Kimi round 2 — the SHIPPED code, `~/Documents/kimi/xscapes-scrollback-audit/round2/KIMI-REPORT-2.md`

**`VERDICT: REVISE`** — *"round 1's four gaps are genuinely closed in the shipped code and the
live mirror is proven in the terminal's own readback and in pixels, but the exit replay corrupts
the very transcript it exists to preserve ... and silently drops the session's final screenful."*
Both were in the live evidence I had quoted as success (`live-mirror.txt`: "LIVE LINE 80" is a
row written over a longer survivor). Dispositions, all closed the same evening:

| # | finding | disposition |
|---|---|---|
| F1 MAJOR | the replay wrote rows over survivors without erasing; tails showed through | **FIXED**: SGR reset + EL before every replayed row; header ON the restored cursor row; ED after the last row. `TestReplayErasesTheRowsItWritesOver` |
| F2 MAJOR | the band's final screen and any rows scrolled since the last tick were in neither the mirror nor the replay | **FIXED, with a ruling I made**: at exit the last kept rows are mirrored and the band's rows are appended to the replay, trailing blanks dropped. A plain session leaves its final screen on the terminal when it ends, UI chrome included; this matches it. `TestExitReplayCarriesTheTranscriptAndTheFinalScreen` (24 lines, one red, exit at once: all 24 in order, the red one still red) |
| F3 MINOR | keys typed during an unanswered cursor wait were held until the next key | **FIXED**: `forwardKeys` is channel-driven and flushes when the wait ends either way. `TestForwardKeysReleasesHeldKeysWhenTheWaitEnds` |
| F4 MINOR | a forwarder past its drain deadline could write into the shell's screen after the replay | **FIXED**: `Host.muted` after the replay; `write` drops everything from then on |
| F5 MINOR | a row kept before a narrowing kept its old width and would wrap | **FIXED**: kept rows are cut to the new width on resize. `TestKeptRowsAreCutToTheNewWidth` |
| F6 MINOR | the suite was not `-race` clean (the harness's `tag`) | **FIXED**: under the harness lock; `go test -race ./internal/host` passes |
| F7 NOTE | untested claims | DSR answered end to end by the harness (transcript begins on row 7, asserted); replay tested; exit path tested; SGR through the pipeline asserted. The hosted resize-with-history case is still model-level only |
| F8 NOTE | flag help promised a replay everywhere; README silent on duplicates and the final screen | help text fixed; README updated below; RESUME corrected (the quoted evidence carried F1) |
| F9 SPECULATION | `mainRow` assumed the main buffer does not move on a resize | **MEASURED and FIXED**: with 200 rows of scrollback above, a grow of 10 pulled 8 history rows back in and pushed the rows DOWN by 8; a shrink keeps the bottom. While the buffer is filling, the write row now moves by the delta (exact for long history; a gap, never an overwrite, for short); once full, rows go to the last row regardless |
| F10 NOTE | the DECSC slot is narrower than feared for the measured agent | unchanged: next in line if a split box is photographed after the fix |
| F11 NOTE | DCS/APC payloads would paint into the model (a hosted sixel program) | documented; Claude emits none; contained to history lines |

Live after the fixes (`live-mirror3.txt`): the header on the row under the command, the 40 rows in
order (24 mirrored, 16 from the final screen), no tails, nothing below but the prompt.
