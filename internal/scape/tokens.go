package scape

import "fmt"

// THE SPEND. His ask of 2026-09-16: "can we add the token spent counter?
// Maybe it can be a simple number on the top right of the xscape?" One
// readable device every scape carries (his re-brief of s35), the way the
// moon carries the context: a number in the top-right corner, drawn by the
// composer on its own darkened ground. Encoded as a count, never a rate. No
// denominator: the spend is a running total across compactions and passes
// the model's window many times over, so "/1M" would read as a budget it is
// not; the window's fill is the moon's channel. The figure is summed off the
// agent's transcript (internal/spend): Claude Code's status payload carries
// only the WINDOW's tokens from the last response and a dollar figure, not
// the session's tokens (its docs, read 2026-09-16).

// TokensRow is the row the counter sits on: one in from the top edge.
const TokensRow = 1

// TokensMargin is the cells kept free between the counter and the right edge.
const TokensMargin = 1

// tokensWidest is the longest label TokensLabel can make: "999k tokens" and
// "9.9M tokens" are eleven cells; the bare number is four at most
// (TestFormatTokens holds it there).
const tokensWidest, tokensNumberWidest = 11, 4

// TokensMinWidth is the width the counter needs at all: the shore's own
// floor. Under it a sky of three rows loses cells it cannot spare (the
// star-touch sweep at 30x8 went from 54 touching pairs to 56 with even the
// bare number's four cells reserved), and nothing is drawn or reserved.
const TokensMinWidth = 40

// TokensUnitFrom is the width from which the counter says "tokens" after
// the number; narrower, the number stands alone. Sixty: at forty columns
// the full label is a quarter of the frame.
const TokensUnitFrom = 60

// FormatTokens says a spend in at most four characters: 950, 9.8k, 845k,
// 1.2M, 12M.
func FormatTokens(n int64) string {
	short := func(v float64, unit string) string {
		s := fmt.Sprintf("%.1f%s", v, unit)
		if len(s) > 4 { // 9.97 rounds to "10.0k": say "10k"
			s = fmt.Sprintf("%.0f%s", v, unit)
		}
		return s
	}
	// The boundaries sit where %.0f rounds up, so 999,999 is "1.0M" and
	// never a five-character "1000k".
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 10_000:
		return short(float64(n)/1000, "k")
	case n < 999_500:
		return fmt.Sprintf("%.0fk", float64(n)/1000)
	case n < 10_000_000:
		return short(float64(n)/1e6, "M")
	case n < 999_500_000:
		return fmt.Sprintf("%.0fM", float64(n)/1e6)
	default:
		return short(float64(n)/1e9, "B")
	}
}

// TokensLabel is the counter's text at a width and where it starts: the
// number with its unit when there is room, the number alone on a narrow
// frame, nothing at all under TokensMinWidth or with no reading yet.
func TokensLabel(w int, n int64) (text string, x, y int, ok bool) {
	if n <= 0 || w < TokensMinWidth {
		return "", 0, 0, false
	}
	text = FormatTokens(n)
	if w >= TokensUnitFrom {
		text += " tokens"
	}
	return text, w - TokensMargin - len(text), TokensRow, true // the last cell is TokensGround's x1
}

// TokensGround is every cell the counter can ever be painted on at a width,
// as [x0, x1] on TokensRow. Kept free of placed stars whether or not a spend
// is known yet, so a star lit before the first reading is never under the
// number later: the count is the checklist's channel and must not drop. A
// FALLING star may cross it (8.4% of flights on the arrival sweep would, and
// blocking them cost his flight floors): the composer's label yields the
// cell to the head for the frame or two it is there.
func TokensGround(w int) (x0, x1, y int) {
	if w < TokensMinWidth {
		return 1, 0, TokensRow // empty: nothing is drawn, nothing is kept free
	}
	x1 = w - 1 - TokensMargin
	widest := tokensWidest
	if w < TokensUnitFrom {
		widest = tokensNumberWidest // the bare number: a tiny sky keeps its cells
	}
	x0 = x1 - widest + 1
	if x0 < 0 {
		x0 = 0
	}
	return x0, x1, TokensRow
}

// inTokensGround says whether a cell is one the counter can be painted on.
func inTokensGround(w int, cell [2]int) bool {
	x0, x1, y := TokensGround(w)
	return cell[1] == y && cell[0] >= x0 && cell[0] <= x1
}

// cellsOffTheCounter is cells less the counter's, for the roomiest-cell
// fallback; the full list when nothing is left, so a tiny sky still places.
func cellsOffTheCounter(w int, cells [][2]int) [][2]int {
	out := make([][2]int, 0, len(cells))
	for _, c := range cells {
		if !inTokensGround(w, c) {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return cells
	}
	return out
}
