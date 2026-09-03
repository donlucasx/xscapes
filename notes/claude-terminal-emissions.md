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
