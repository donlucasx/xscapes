# Width audit — Terminal.app's alternate screen on a WIDTH change

*2026-09-05, session 16. Measured with `notes/widthprobe` in windows a script opened
(resolved by the tty of the `do script` tab, no keystrokes), read back with
`history of tab`. His OK: "go ahead".*

## Why

His Terminal.app at 123x55, after a drag that changed the WIDTH as well as the
height (the trace: 34 resizes, 120 → 130 → 123 columns), showed the band's last
column painted with cells from several rows lower, and a six-column patch of scape
cells in the three rows above the band at the left edge. The host's bytes replay
CLEAN in the screen model (`TestReplayTrace`, `TRACE_RETAIN=0`). The model's width
rule for the alternate screen -- cut to the new width, pad with blanks -- had never
been measured; session 13 measured height only.

## Measured

1. **Retain and clip.** A row keeps every cell it ever had, at the widest the window
   has been. Narrowing 120 → 74 clips the display; the read-back still holds all 120
   cells of every row, `<` marker in column 120 included. Widening 74 → 80 → 86 shows
   the retained cells again in the new columns. No reflow, no wrap, no blanking.
   (Driver 1: before / 74 / 80 / 86 -- four identical read-backs.)
2. **Erase-line reaches the visible width only.** At 74 columns, `ESC[2K` on a
   120-cell row left cells 75–120 in place (`EEE…E<`, 46 cells); `ESC[K` from column 3
   kept `R0` and the same tail. Widening to 100 showed the tail in columns 75–100.
3. **A repaint at the narrow width leaves the tail.** Row 7 rewritten at 74 columns
   read back as the new 74 cells followed by the old 46 (`R07ggg…gGGG…G<`).
4. **The window is not snapped to the cell grid.** His screenshot's content spans
   123.6 columns: Terminal.app draws the partial 124th column, and that column holds
   RETAINED cells from the 130-wide layout -- the "last column painted from rows
   below". The model with the rule on (`TRACE_RETAIN=1`) reproduces exactly those
   cells at columns 124–130 of the band rows.

## Consequences for the host

- A narrowing followed by a widening exposes old cells in the new columns of EVERY
  row until something repaints them. The band repaints in full on the tick after a
  resize (`dmg.reset`), so its exposure lasts one tick. The agent's rows are Claude's:
  Ink erases with `ESC[2K`, which reaches the visible width only, so a widened window
  shows stale tails in transcript rows Claude does not rewrite. Not ours.
- The partial column cannot be painted (it is beyond the reported width) and cannot
  be cleared (no erase reaches hidden cells). Cosmetic, Terminal.app only, and only
  after a narrow-from-wider to a non-grid width. Documented, not fixed.
- The six-column patch above the band is NOT reproduced under either width rule.
  Still open. It is not retained-cell exposure (it sits at columns 1–7).

## The model

`screen.retainWidth` (off by default; Ghostty's source cuts and pads, and every
height test was written against that) keeps rows at their widest and clips erases to
the visible width. `TestReplayTrace` and the trace probes take `TRACE_RETAIN=1`.

## Also measured the same day

Terminal.app's U+2580 ink starts 5px below the cell top and stops 13px above the
bottom (Menlo, 30px rows); U+2584's ink runs from 17px to the bottom edge exactly.
So a split cell drawn with U+2584 and the colours the other way up has no hairline.
`term.LowerHalf`, picked by `DetectSplit` from `TERM_PROGRAM`; `XSCAPES_SPLIT`
overrides. Measured from his saved screenshot of a printf test, stripe by stripe.
