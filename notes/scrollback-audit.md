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
| MAIN (control) | **all 40** | all 40 |

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
last rows of the band, so this fires on **every shrink of more than two rows**, and a drag is
many small shrinks. Kimi's DECSC/DECRC-slot hypothesis is not needed to explain Report 1.

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

Two `screencapture`s during the 400-row burst both show the alternate screen, never the main
buffer. Two samples do not prove there is no flicker; **his eyes are the instrument for that**
(`go run ./notes/mirrorprobe -n 400 -gap 5ms` in a Terminal.app window and watch).

⚠ **After `ESC[?1049l` Terminal.app keeps only a few of the mirrored rows** (4 of 30 in the run;
it restores the main screen it saved at `1049h` and discards most of what scrolled meanwhile). So
the mirrored history is complete DURING the session and mostly gone after it. Cheap mitigation:
the host keeps the rows it mirrored and prints them once more on Close, after leaving the
alternate screen, so post-session review survives too. Gate that on `TERM_PROGRAM=Apple_Terminal`:
an xterm-like terminal keeps the main scrollback intact across the alternate screen and a replay
there would duplicate every line.

### 4. `wheelprobe` — what the wheel sends on the alternate screen. **VOID, do not quote.**

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
