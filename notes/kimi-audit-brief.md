# Audit brief: two open defects in xscapes

You are auditing another agent's work. **Your job is to falsify, not to agree.**
The last time you audited this project you caught a confident wrong answer that
four of our own instruments had missed, and that is exactly what is wanted again.
If a claim below is supported, say so briefly and move on. Spend your effort on
the ones that are NOT.

Read the code and the evidence. Do not take the prose below as fact -- every
numbered claim is something we believe and could be wrong about.

**READ ONLY. Do not edit, create, move or delete any file.**

Repo: `~/Documents/claude/xscapes` (Go, no deps). Target terminal: macOS
Terminal.app, Menlo, 256-colour, cells 14 x 30 device px.

---

## ISSUE 1 -- "hairlines": 1px rules through the sun and moon

The scape paints sub-cell shapes by splitting a cell into two background
colours, drawn with a half-block glyph. On Terminal.app the user sees thin
horizontal rules through the disc. Three fixes have shipped; he still reports it.

Read: `internal/term/color.go` (LowerHalf, NoSplitCells, Split),
`internal/canvas/canvas.go` (resolve, SetBGHalves),
`internal/scape/shore.go` (moon, discHalf, flatEdge, FlatCaps),
`notes/rulecount/main.go` (the instrument),
`internal/scape/disc_rule_test.go`, `internal/scape/moon_shape_test.go`.

**Claims to attack:**

1. Menlo's block ink occupies y=4.6..29.4 of a 30px row, so a split cell always
   leaves ~0.6px of its BACKGROUND at the cell's bottom edge, and that sliver is
   the rule. Calibration: in a user screenshot the disc's centre column reads
   17px sky (50,95,131), 12px rim (165,137,136), ONE px (100,113,134); and
   0.43*rim + 0.57*sky = (96,112,133).
2. Therefore a cell has exactly TWO achievable partitions -- U+2584 (error 0.6px
   at the bottom) and U+2580 (error 4.6px at the top) -- and NO glyph avoids the
   leak. We used this to reject a proposed "choose the half-block orientation
   per cell" fix. **Is the two-partition claim actually exhaustive?** Consider
   U+2581-2587, U+2594, quadrants, reverse video, combining marks, underline.
3. What the user sees is the RUN, not the pixel: at his geometry the frame
   carries 4 rules, two of them 5 cells wide at the disc's flat top and bottom
   caps. Collapsing only edge cells whose left/right neighbour has its edge in
   the same cell on the same side takes the longest run 5 -> 1 with the
   silhouette's width profile unchanged.
4. Going to ZERO rules requires collapsing every edge cell by area coverage, and
   that makes the disc a rectangle at 3 of 13 tested heights.
5. **Weakest claim, attack it hardest:** we read font metrics with fontTools and
   got U+2588 filling 87.6% of Menlo's hhea line box (ascent-descent+lineGap)
   against 100.0% for Andale Mono, and concluded a font swap could remove every
   hairline. **Is hhea the line box Terminal.app actually uses for row height?**
   macOS text layout may use OS/2 sTypo* or usWin* metrics instead. If we are
   reading the wrong table, this recommendation is worthless and the user is
   about to waste time on it.

---

## ISSUE 2 -- corrupted terminal scrollback

The agent runs INSIDE the scape on the alternate screen; the host owns the
screen, hosts the agent in the top rows, paints the scape below, and mirrors
rows leaving the agent region into the terminal's scrollback via DECSET 47.
Users see scrollback rows where two different texts occupy one row, e.g.
`wi=1.aTestMain now clears it.` and `xigrowingaate6.6eMB/min`. Five sessions
failed to explain it.

Read: `internal/host/host.go` (resizeSequence, write, openTrace),
`internal/host/band_control.go` (Rebind, RebindShrinkAlt, reallocBand),
`internal/host/screen.go` (the host's own screen model),
and the evidence at `~/Documents/Screenshots/xscapes-evidence/`:
`README.md`, `resize-occurrence-20260908.sizes`, and
`resize-occurrence-20260908.bin` -- 300 KB of the host's real output starting at
a resize, captured while the defect occurred. It is raw terminal bytes; read it
with a hex/escape dump, not as text.

**Claims to attack:**

1. The mechanism is the agent's post-resize repaint: it emits a run of
   `ESC[2K ESC[1B`, then `ESC[H`, then repaints rows as column-addressed
   segments (`ESC[3G One ESC[7G defect ESC[14G you`) with NO per-row erase, so
   any cell not covered by a segment keeps what was there. The merged text is
   never emitted by anyone; it is assembled in the terminal's cells.
2. `wi=1.aTestMain` is explained exactly by `ESC[3G =1 ESC[39m . ESC[7G TestMain`
   -- columns 1-2 never written, column 6 never written.
3. A note this project carried since session 11, "Claude Code emits nothing on a
   resize", is REFUTED: it emits within ~0.1 s of the resize.
4. The trigger is a RESIZE, not window age: the size sidecar records exactly two
   sizes in a 6h17m session (119x51 at offset 0, 102x59 at offset 1868283253)
   and the occurrence follows the second.
5. **Open, and the question we most want answered:** our own resize sequence
   emits `ESC[8S` (scroll the WHOLE screen up 8 lines on a row-grow), clears only
   the 4 rows changing hands, then `ESC[34;59r ESC[34;1H ESC[26M`. Does that move
   the agent's rows out from under a repaint that assumes they did not move? If
   so, what is the correct host behaviour? If not, what is?
6. Is this fixable from a host process at all, or is any host that reflows a
   region under a differentially-redrawing TUI going to produce this?

**A trap, recorded so you do not repeat it:** the host tees everything it writes
to the terminal into the trace, INCLUDING this session's own terminal output. We
grepped a trace for a corrupted string, found it, and concluded "we wrote it" --
but we had PRINTED that string to the terminal moments earlier and found our own
diagnostic. Any search of a trace must be restricted to bytes written before the
search term was ever displayed.

---

## What to return

1. Any claim above that the evidence does not support, and why.
2. The single most likely-wrong claim, with your reasoning.
3. Anything simpler we have missed on either issue.
4. For issue 2 claim 5: your read of whether the host or the agent is at fault,
   and the concrete fix if there is one.

Be blunt. Cite file:line or byte offsets. If you cannot verify something from
the evidence, say "unverifiable from what is here" rather than guessing.
