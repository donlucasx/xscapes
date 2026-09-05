# xscapes brand

The identity is a block cursor over the x, drawn in reverse video. Locked
2026-09-05. The full rulebook with measurements is `guidelines.html` in this
folder (open it in a browser); this file is the same rules in plain text.

## The mark

    printf '\e[7mx\e[0mscapes'

One terminal cell with the x in reverse video, then `scapes` in the same
face. The cell is the text colour of its context; the x is the ground showing
through. Never a third colour in the cell.

- Geometry: one cell is 0.6 em wide and 1.2 em tall (1:2). Geist Mono 700,
  no tracking. The x sits on the terminal's baseline, 955/1200 from the top.
- Clear space: one cell on every side. Minimum size: 12 px for the lockup,
  8 px for the cell alone.
- The symbol (the cell alone) may blink about twice a second, only when it
  is the only thing moving. The off state is a hollow cell. The lockup never
  blinks.
- Not this: the cell over any other letter; a space after the cell; a square
  cell; rounded corners; an outline; a shadow; a recoloured x; `scapes` in
  the sans; any capital letter, ever.

## Colour

Every colour is an xterm-256 index from 16 to 255, the product's own rule.
Named as hex, index, SGR parameter (38 text, 48 background).

Dark, the default:

| role    | hex     | index | SGR      |
|---------|---------|-------|----------|
| ground  | #121212 | 233   | 48;5;233 |
| surface | #1c1c1c | 234   | 48;5;234 |
| line    | #303030 | 236   | 38;5;236 |
| dim     | #8a8a8a | 245   | 38;5;245 |
| ink     | #eeeeee | 255   | 38;5;255 |
| shallow | #87afff | 111   | 38;5;111 | accent, one highlight per surface

Light, for print and light decks:

| role    | hex     | index | SGR      |
|---------|---------|-------|----------|
| ground  | #eeeeee | 255   | 48;5;255 |
| surface | #e4e4e4 | 254   | 48;5;254 |
| line    | #d0d0d0 | 252   | 38;5;252 |
| dim     | #626262 | 241   | 38;5;241 |
| ink     | #121212 | 233   | 38;5;233 |
| dawn    | #005f87 | 24    | 38;5;24  | accent; sea 67 for rules only, not words

Claude's oranges (terracotta, ginger) are out as accents. The shore's own
palette (23, 67, 111, 180, 137, 223) appears only where the shore is shown,
as the shore.

## Reverse, the alternate colour application

The same cell filled with gold instead of ink; the x stays the ground, the
word stays ink. Use it once per surface where the brand is the subject: the
deck's title and closing slides, the Commons page header, the avatar and
social image. Everywhere the mark labels something else, the cell is ink.
In the terminal it is always plain reverse video.

| use                | hex     | index | SGR      |
|--------------------|---------|-------|----------|
| cell on dark       | #d7af5f | 179   | 48;5;179 |
| cell on light      | #af8700 | 136   | 48;5;136 |
| gold text on light | #875f00 | 94    | 38;5;94  |

    printf '\e[48;5;179;38;5;233mx\e[0mscapes'

## Type

Geist Mono 700 for the lockup; Geist Mono 400 for code, labels, captions;
Geist 400/500/600 for copy. Scale: 11 label, 13 code, 15 body, 22 subhead,
34 title; deck title 72 at 1920 by 1080. Fallbacks: SF Mono or Menlo, and
the system sans. In the terminal the mark takes whatever font the user runs.

## Voice

Plain over poetic, concrete over conceptual, evidence over claim. Fixed
lines: `xscapes` (always lowercase) · `a thinking screen for terminal agents`
· `the waiting layer for terminal agents` · `the water is the work, the sky
is the world` · `xscapes claude`. No em dashes, no exclamation marks, no
"AI-powered", no "delightful".

## Files

SVGs are the font's real outlines on the 600 x 1200 cell, written by
`make-brand.py` (needs fontTools; downloads Geist Mono 700 if it is not next
to the script). Dark and light files are transparent and assume grounds 233
and 255; the `-bg` files carry their own ground.

    lockup-dark.svg  lockup-light.svg  lockup-dark-bg.svg  lockup-light-bg.svg
    lockup-gold-dark.svg  lockup-gold-light.svg
    mark-dark.svg  mark-light.svg  mark-gold-dark.svg  mark-gold-light.svg
    avatar.svg  avatar-gold.svg        (512 square; export PNG for GitHub)

README use, both themes:

    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="assets/brand/lockup-dark.svg">
      <img src="assets/brand/lockup-light.svg" alt="xscapes" height="48">
    </picture>

Decisions and dates: `_FEEDBACK.md`, session 17.
