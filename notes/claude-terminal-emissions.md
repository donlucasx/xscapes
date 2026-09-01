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
2. **`ESC[H` on resize.** On SIGWINCH Claude homes the cursor and repaints from
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
