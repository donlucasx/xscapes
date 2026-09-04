# What Claude Code actually writes to its terminal

Measured 2026-09-01, to decide whether xscapes can host the agent in a band
(PTY + scroll region, no VT parser) or whether it needs a real terminal
emulator. **Trust this file; do not re-derive it.**

## Method

Claude Code v2.1.252 run in a tmux pane with `TERM=xterm-256color`, its raw
output tapped with `tmux pipe-pane -o` (which taps the program's own bytes
before tmux re-renders them, so this is what Claude would emit to a plain
xterm-ish terminal). Two captures: (1) startup, typing into the prompt box,
resize 40->30->50 rows; (2) startup plus a real turn -- a Bash tool call
producing 60 lines, then a text answer. Analyser:
`scratchpad/analyze.py` in the session that produced this.

## What it emits

| Sequence | Startup capture | Real turn |
|---|---|---|
| Alternate screen `?1049` / `?47` | 0 | 0 |
| CUP absolute `ESC[r;cH` | 2 (both `ESC[H`, on RESIZE only) | **0** |
| VPA absolute row `ESC[nd` | 0 | 0 |
| ED erase display `ESC[J` | 0 (the one in the capture is zsh's prompt) | 0 |
| Scroll up/down `ESC[S` / `ESC[T` | 0 | 0 |
| Reverse index `ESC M` | 0 | 0 |
| Insert/delete line `ESC[L` / `ESC[M` | 0 | 0 |
| `ESC[r` (DECSTBM **reset**) | 1, at startup | 1, at startup |
| Origin mode `ESC[?6h/l` | 0 | 0 |
| Relative moves `A B C D` | yes | yes (247) |
| Column-only `ESC[nG` (CHA) | yes | yes (143) |
| Erase line `ESC[2K` (EL) | yes | yes (26) |

**Claude Code redraws with relative vertical movement, column addressing and
erase-line.** It stays on the normal screen and it does not address absolute
rows during a turn. Its transcript scrolls by writing newlines at the bottom
line, which a scroll region confines by construction.

Two exceptions, and they are the whole design:

1. **`ESC[r` at startup**, wrapped in `ESC 7` / `ESC 8`. This RESETS the scroll
   region to the full screen, so a region set before launch is wiped. A host
   must swallow those three bytes in the pass-through stream (a byte match, not
   a parser) or re-assert its region after startup.
2. **`ESC[H` on resize -- ONLY ON A FRESH SESSION. See the correction below.**
   On SIGWINCH Claude homes the cursor and repaints from
   the top with a `ESC[2K` + `ESC[1B` clear-down loop. Absolute home is the only
   absolute move it makes. Origin mode (`ESC[?6h`) makes `ESC[H` land at the
   region's top margin instead of the screen's -- and Claude never touches
   DECOM itself, so once the host sets it, it stays set. Inside a region the
   clear-down loop is safe: `ESC[1B` stops at the bottom margin and `ESC[2K`
   only erases the cursor's own line.

It also queries the terminal at startup (`ESC[c`, `ESC[>5u` kitty keyboard,
`ESC[?2026$p` synchronized output). A host must forward the real terminal's
replies back into the PTY byte-transparently.

## The scrollback constraint (this is the expensive one)

Lines that scroll out of a DECSTBM region reach the scrollback **only if the
region is anchored at row 1**. Measured in Terminal.app via
`history of tab` after printing 40 lines:

- region rows **1-10**: LINE1..LINE31 all present in history. Scrollback intact.
- region rows **5-14**: **zero** lines survive. Scrollback gone.

tmux 3.6b keeps history in the 1-10 case too.

**Therefore the agent's band must start at row 1**, and nothing of the scape can
be painted above it. The scape gets the rows below. That still fits the shape of
the thing: agent band, then a sky strip carrying the moon, then the sea, then the
beach -- a horizon works just as well with the agent above it as with sky above it.

## What this rules in and out

- **Pass-through band is viable.** PTY + DECSTBM + DECOM + swallow `ESC[r`. No
  VT parser, so no class of bug that can corrupt Claude's own UI.
- **What it cannot do**: show the sea *through* Claude's blank space. 83.5% of a
  real Claude pane is blank, and in a band that stays black. Only a real
  emulator, which owns every cell, can composite into it -- and it would then
  own the scrollback too, removing the row-1 constraint above.

## One more thing the host has to know

**DECSTBM moves the cursor to home, and a parameterless `ESC[r` is still
DECSTBM.** This is not about Claude Code; it is about every terminal. It cost
an afternoon anyway: the host placed the cursor below the band on exit and then
reset the scroll region one last time, which pulled the cursor back to row 1 --
and the first thing a zsh prompt emits is `ESC[J`, so from row 1 it erased
everything the agent had drawn. The screen went blank on every exit and the
agent's last output was simply gone.

The rule that falls out of it, and that `internal/host` now encodes in tests:
**save the cursor before touching the region, and place the cursor after the
last time you touch it.**

## Correction, 2026-09-02: Claude does NOT repaint on resize

The claim above -- "on SIGWINCH Claude homes the cursor and repaints from the
top" -- was measured on a session that had only just started, and what it caught
was the STARTUP draw, not a response to the resize. Generalising it was wrong,
and a whole resize implementation was built on it.

**Measured again, on a session with a real transcript: Claude Code emits ZERO
bytes on a resize.** Not a repaint, not a clear, nothing. Confirmed in both the
hosted case and plain Claude in tmux: strip the host's own frames out of the
capture and exactly zero bytes remain between the resize and the next keystroke.

**It places its input line purely by relative moves from wherever the cursor is:**

```
ESC[2D ESC[4B \r ESC[2C ESC[4A  good  ESC[8G  morning
```

Never a row number. It trusts the cursor completely, which is what makes it
survive a resize in a normal terminal -- the terminal moves the content and the
cursor together, so relative moves stay true -- and what makes it fragile inside
a band.

**And the mechanism that breaks a band: growing a window makes the terminal pull
scrolled-off lines back in from history**, pushing everything down. Proof: the
startup banner, long scrolled away, reappears at the band's bottom edge after a
grow. The agent's UI slides out of its band into the scape's rows, the agent
never notices, and the scape paints over it.

Controls that isolated it: hosted with no resize works; plain Claude with the
same resize works; hosted plus resize fails. So it is the band, not the agent.

This is why the host runs on the alternate screen: no history, nothing to pull
back, nothing moves. The cost is that output scrolling out of the band is gone
rather than saved, which is the exact trade the row-1 anchoring was made to
avoid -- and it turned out to be a scrollback that could not survive being used.

## Measurement, 2026-09-03: the two screen buffers resize DIFFERENTLY

The correction above ends "this is why the host runs on the alternate screen:
no history, nothing to pull back, nothing moves." That is right, and the host
then went on to apply a MAIN-screen rule to the alternate screen anyway.

Measured with `notes/anchorprobe`, which parks the cursor mid-screen, emits
nothing at all while the window is resized from outside (AppleScript, so the
drag is repeatable), and reads the cursor back with DSR. A terminal moves the
cursor with the content it moves, so the cursor answers the question without
anyone having to eyeball a row number mid-drag.

Same window, same 11-row shrink, cursor parked on row 21 of 43:

| screen | cursor after | meaning |
|---|---|---|
| **main** | row 10 (= 21 − 11) | keeps the **BOTTOM**; every row slides up by the rows lost |
| **alternate** | row 21 (unmoved) | keeps the **TOP**; content is truncated and nothing slides |

Run it both ways yourself: `go run ./notes/anchorprobe` and
`go run ./notes/anchorprobe -main`, resizing the window within four seconds.
A first run answered "KEPT TOP" only because the parked row was clamped to the
new height; the shrink was made gentler so no clamping is involved and the two
predictions are 21 vs 10, and the main-screen control gives the opposite answer
from the identical drag. One reading with a control beats three without.

**Consequence.** `drop` in the resize branch of `host.go` is a MAIN-screen
correction. On the alternate screen it subtracts rows that never moved, which
walks the clear up into the agent's transcript -- and at a big enough shrink it
reaches row 1 and takes the input box with it.

⚠ **The instrument that missed this ran `AltScreen: false`.** Every resize test
in the package did. A harness that never enters the mode production runs in
cannot see that mode's defects, and it was on that harness's word that the
resize damage was written off as Claude Code's to fix.

## Measurement, 2026-09-03 (second pass): the anchoring holds WITH a scroll region

An external audit made the sharp objection that the anchoring above was measured in the wrong
configuration: `notes/anchorprobe` set no scroll region and no origin mode, while production runs
DECSTBM 1..N + DECOM. A resize with an active region is a different path in a terminal -- the
region has to be remapped across it -- so the first result was a fact about the region-free case
only. That objection was correct and the gap is now closed.

Re-measured with `go run ./notes/anchorprobe -region`, which pins `ESC[1;Nr` + `ESC[?6h` exactly as
`band_control.go` does and parks the cursor inside the region:

| config | change | parked | cursor after | verdict |
|---|---|---|---|---|
| ALT + DECSTBM + DECOM | 24 -> 54 (**+30**) | row 12 | **row 12** | ANCHORED TOP |
| ALT + DECSTBM + DECOM | 54 -> 24 (**-30**) | row 27 | **row 24** (clamped) | ANCHORED TOP |

Anchored-bottom would have predicted row 42 and row 1. **The scroll region does not change the
anchoring**: the alternate screen is anchored-top in both directions, region or no region. So the
terminal does not move the hosted agent's content on a resize, and anything that moved it is ours
or the agent's.

## Terminal.app does not support DECRQSS (2026-09-03)

`ESC P $ q m ESC \` gets no reply even with a known rendition in force (the control). So the current
SGR cannot be read back from this terminal, and "does DECRC restore SGR" cannot be measured
directly. It is settled by the running system instead: the scape paint brackets its writes with
DECSC/DECRC and every scape line ends in `ESC[0m`, twelve times a second. If DECRC did not restore
SGR, Claude Code's UI would be colourless. It is not.

## ⚠ CORRECTION, 2026-09-03 (third pass): the anchoring measurement measured the CURSOR, not the CONTENT

Both anchoring sections above are measured with `notes/anchorprobe`, which parks the cursor, resizes,
and reads the cursor back with DSR. Its stated premise is "a terminal moves the cursor with the
content it moves". **That premise is an inference, and on Terminal.app's alternate screen it is
false.** The probe therefore measured cursor behaviour and the conclusion was written up as content
behaviour.

What the evidence actually says, from a trace of the real failure (`XSCAPES_TRACE`):

- The agent drew its startup screen at byte 184918 and emitted **NOTHING** from the first resize
  (byte 422004) until Ctrl-C (byte 825233) -- 400KB of host frames spanning six resizes with zero
  agent output. So the agent did not redraw, and did not move its own text.
- The trace contains **no ED at all**, and the host's clear ranges never reach the rows holding the
  agent's text (verified per-resize against `Band()`). So the host did not erase it.
- Both other writers eliminated, the text still moved. Only the terminal is left.
- And the owner's screenshots show WHERE it went: after a stretch, the startup banner sits on the
  LAST rows of the band with the whole band blank above it, i.e. pushed DOWN by the grow, with
  everything past the band's bottom painted over by the scape.

**So on the alternate screen Terminal.app pushes CONTENT down on a grow while leaving the CURSOR at
its absolute row.** The two move independently, which is why the cursor probe reported "anchored
top" for a screen whose content was anchored bottom.

This is the third instrument in one day that returned a clean answer to the wrong question, and the
external audit called this one before the measurement did. Its objection was that the probe ran in
the wrong CONFIGURATION; the deeper fault was that it measured the wrong OBSERVABLE.

⚠ **Do not quote the anchoring tables above for content.** They are valid for the cursor only. A
content measurement needs the numbered rows read off the screen by eye, which no probe here does.

## MEASURED, 2026-09-03 (fourth pass): the alternate screen is ANCHORED BOTTOM on a grow

Read off the screen by eye with `notes/contentprobe`, in production's configuration (alternate
screen, DECSTBM band, origin mode), because no escape sequence reads cells back from Terminal.app
and every attempt to infer content from something else has been wrong.

Every row painted with its own number, then the window stretched:

| | top of window | ROW 01 sits at | cursor |
|---|---|---|---|
| before, 120x30 | `ROW 01` | screen row 1 | row 1 |
| after, 120x51 (**+21**) | **blank** | **screen row 22** | **row 1** |

`1 + 21 = 22`. **Content moved down by exactly the grow delta. The cursor did not move at all.**

That is the whole story of this bug and of three wrong answers:

1. Terminal.app pushes CONTENT down on an alternate-screen grow, inserting blank rows at the top.
2. It leaves the CURSOR at its absolute row.
3. `notes/anchorprobe` measures the cursor, so it reported "anchored top" for a screen whose content
   was anchored bottom -- and that reading was written into this file twice as a fact about content.

**Consequence for the host.** On a grow the agent's text slides down by the delta; whatever crosses
the band's bottom edge is painted over by the scape on the next frame, and only the fragment still
inside the band survives. That is exactly what the owner photographed: a startup banner sitting on
the last rows of the band with everything above it blank, and on a bigger grow, nothing at all.

⚠ **The shrink direction is NOT yet measured for content.** If it is anchored-bottom too, content
slides UP on a shrink, which is what the `drop` correction in host.go was originally written for --
and `drop` was made main-screen-only on 2026-09-03 on the strength of the CURSOR measurement. That
change may be wrong. Measure before touching it again.

## MEASURED, 2026-09-03: the shrink direction too. Anchored BOTTOM, both ways.

Same probe, same configuration, window made SHORTER: a higher row number at the top -- content slid
UP by the rows lost, and the top rows are gone.

**So Terminal.app's alternate screen anchors content to the BOTTOM edge in BOTH directions, and the
cursor moves with NEITHER.** That is the whole fact this cost a day to establish:

| direction | content | cursor |
|---|---|---|
| grow by N | pushed DOWN by N, blanks inserted at the top | stays on its absolute row |
| shrink by N | pulled UP by N, the top N rows destroyed | stays on its absolute row |

Consequences, both now implemented in `host.go`:

1. **Grow.** The push has to be UNDONE before anything is painted, or the agent's UI is left sitting
   below the band where the scape paints over it. `Rebind` takes a scroll-up count and emits SU over
   the full screen. Moving rows needs no model of the UI in them, which is why the host is allowed
   to do it. The cursor needs no adjustment: the terminal moved the content and not the cursor, and
   SU moves the content back and not the cursor, so they end where they began.
2. **Shrink.** `drop` belongs on this screen after all. It was made main-screen-only earlier the
   same day on the strength of the CURSOR reading; that change was wrong and is reverted. The rows
   the terminal destroyed are gone either way -- what `drop` does is clear the scape that slid up
   into the band without blanking the agent rows that survived.

## MEASURED, 2026-09-03: the MAIN screen, both axes. Height is easy; WIDTH is fatal.

Measured because scrollback matters to the owner and only the main screen has any. `notes/contentprobe
-main`, numbered rows, read by eye.

**Height, 120x30 -> 120x47 (+17): ANCHORED TOP.** `ROW 01` stayed at the top and the blank rows
appeared at the BOTTOM. The opposite of the alternate screen, and it needs no correction at all.

⚠ Measured on a screen with almost no scrollback above it. The 2026-09-02 note says a grow pulls
scrolled-off lines back in from history; with nothing in history there is nothing to pull. **This
result is for an empty-scrollback screen and may not hold in a long session.** Do not generalise it
without re-measuring against real history.

**Width, 120x47 -> 78x47: TOTAL REFLOW, and it is not correctable.** Every full-width row re-wrapped
into TWO rows -- `ROW 08 ----` on one line, `---- ROW 08` on the next -- and the view shifted so the
top showed row 07's tail. Claude Code's UI is made of full-width lines (the input box borders, the
status line), so any narrowing doubles them.

**The host cannot undo this.** Everything it can do moves whole ROWS; reflow changes how many rows a
logical LINE occupies. There is no row-move that inverts it, and the host does not model logical
lines. The alternate screen never reflows, which is the property that makes the band possible at all.

⇒ **"Put the band back on the main screen to get scrollback" is dead.** Height would be easier there,
but width is unfixable, and windows get narrowed. Scrollback and the band cannot both come from
Terminal.app; if xscapes wants scrollback it has to own it.

## MEASURED, 2026-09-03 (session 14): five facts for the scrollback question, all read back by machine

Read with `history of selected tab of window id N` over `osascript`, which returns the ALTERNATE
screen's visible cells as well as the main buffer's — so "nothing reads cells back from
Terminal.app" (above) is retired. Probes in `notes/histprobe`, `notes/shrinkprobe`,
`notes/mirrorprobe`; the full write-up is `notes/scrollback-audit.md`.

1. **The alternate screen has NO history.** Forty lines through a row-1 band of 17: on the
   alternate screen only the 16 still visible are in `history`, and none after exit. On the main
   screen (control) all forty during, and 1–25 after exit (the shell prompt's `ESC[J` took the
   rest; Kimi F6). The row-1 anchoring rule above is a MAIN-screen
   fact; on the alternate screen a line that leaves the band is gone.
2. **A shrink homes the agent's cursor through the host's own Rebind.** Cursor on row 16 of a
   17-row band, window 30→24 (band 17→14): after the terminal's shrink the cursor is still on 16
   (content moved up 6); after `Rebind` it is on **row 1**. **DECRC into a region that no longer
   contains the saved row lands on row 1.** A Claude-style relative redraw then paints a second
   input box at the top of the band — Report 1 (2026-09-03 ~11:57), reproduced. Fixed by
   restoring while the region is still the full screen, moving relatively, saving again, pinning
   the band, restoring: cursor 13 and the box redrawn exactly over the moved one. **Eleven
   geometries** (1/2/3/6/20-row shrinks, cursor top/mid/bottom, a three-tick drag, a shrink then a
   grow, a shrink the band absorbs): overlay wherever the content survived; production misplaces
   the box on EVERY shrink, by the tick size for small ones. Table in `notes/scrollback-audit.md`.
3. **DECSET 47 switches buffers without clearing, in both directions.** 400 round trips
   (`ESC[?47l` · write a row at the main buffer's last line · `\r\n` · `ESC[?47h`) at 5ms: the
   alternate screen intact, all 400 rows in the main buffer's history, in order.
4. **Terminal.app's view puts the main buffer ABOVE the alternate screen.** `history` while the
   alternate screen is up lists the main rows, then the alternate rows. Scrolling up in
   `xscapes claude` shows the shell's main buffer — including its blank rows, which is the "black
   point" he reported, not cleared rows in scrollback.
5. **`ESC[?1049l` discards most of what scrolled into the main buffer meanwhile**: 4 of 30 mirrored
   rows survived exit. Anything mirrored for post-session review has to be printed again after
   leaving the alternate screen.

6. **The wheel on the alternate screen reaches the PROGRAM only under mouse reporting.** Measured
   after he granted Accessibility to Terminal.app (the first attempt was VOID: the CGEvent posts
   never arrived, and the control proved it). With SGR mouse reporting on, four ticks up and three
   down arrive as `ESC[<64;61;15M` ×4 and `ESC[<65;61;15M` ×3 -- the control. With no mouse
   reporting, production's state, the same ticks deliver **0 bytes**: Terminal.app does not
   translate the wheel into arrow keys. ⚠ Whether the wheel scrolls the VIEW cannot be measured
   synthetically: CGEvent scroll ticks never moved Terminal.app's view, main screen included.
7. **Keys over the alternate screen**: plain Page Up is delivered to the program (`ESC[5~`);
   **Shift+Page Up is kept by Terminal.app and scrolls the view** -- over the alternate screen,
   into the main buffer's rows above it (screenshot: MIRROR LINE 11-40 with the band scrolled
   out of sight). While a batch of rows lands below, the scrolled view HOLDS its position; only
   the scrollbar thumb moves (two captures five seconds apart, identical rows).

8. **The MAIN buffer on a grow WITH scrollback pulls history back in.** 200 rows of scrollback,
   30 numbered rows on screen, window grown 30→40: OLD HISTORY 194–200 and a blank row came in at
   the top and `MAIN ROW 01` moved from row 1 to row 9 -- content pushed DOWN by 8 of the 10 rows
   grown. The 2026-09-03 "anchored TOP on grow" measurement above was for an EMPTY scrollback and
   said so; this is the general case. A shrink keeps the bottom (anchorprobe, main). The mirror's
   write row follows both (`host.go`, the resize branch).
