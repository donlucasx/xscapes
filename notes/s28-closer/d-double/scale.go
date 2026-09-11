package main

import (
	"sort"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// scale is nearest-neighbour INTEGER scaling of a SOURCE bitmap: every source
// pixel becomes sx by sy pixels. Exact by construction -- no interpolation, no
// threshold, no new ink anywhere the source had none, and symmetric because an
// integer factor maps every column the same way.
func scale(rows []string, sx, sy int) []string {
	out := make([]string, 0, len(rows)*sy)
	for _, r := range rows {
		var sb strings.Builder
		for _, c := range r {
			for k := 0; k < sx; k++ {
				sb.WriteRune(c)
			}
		}
		line := sb.String()
		for k := 0; k < sy; k++ {
			out = append(out, line)
		}
	}
	return out
}

// resample is nearest-neighbour to an ARBITRARY size, sampling from the centre
// of each destination pixel so the map is symmetric about the middle. Written
// to test whether the ladder between 1x and 2x can be generated instead of
// drawn; symmetry is checked in main, not assumed here.
func resample(rows []string, w, h int) []string {
	sw, sh := len([]rune(rows[0])), len(rows)
	out := make([]string, 0, h)
	for y := 0; y < h; y++ {
		sy := (y*2 + 1) * sh / (2 * h)
		if sy >= sh {
			sy = sh - 1
		}
		src := []rune(rows[sy])
		var sb strings.Builder
		for x := 0; x < w; x++ {
			sx := (x*2 + 1) * sw / (2 * w)
			if sx >= sw {
				sx = sw - 1
			}
			sb.WriteRune(src[sx])
		}
		out = append(out, sb.String())
	}
	return out
}

// epx is Scale2x/EPX: the same 2x footprint as scale(rows,2,2), but a pixel
// whose two orthogonal neighbours agree with each other and disagree across
// the diagonal gets that neighbour's value in the corresponding corner. On
// 1-bit art it rounds outer corners and fills inner ones, which is exactly the
// horizontal sub-cell detail an integer double throws away.
func epx(rows []string) []string {
	w, h := len([]rune(rows[0])), len(rows)
	g := make([][]rune, h)
	for y, r := range rows {
		g[y] = []rune(r)
	}
	at := func(x, y int) rune {
		if x < 0 || y < 0 || x >= w || y >= h {
			return '.'
		}
		return g[y][x]
	}
	out := make([]string, 0, h*2)
	for y := 0; y < h; y++ {
		var top, bot strings.Builder
		for x := 0; x < w; x++ {
			p := at(x, y)
			a, b, c, d := at(x, y-1), at(x+1, y), at(x-1, y), at(x, y+1)
			e0, e1, e2, e3 := p, p, p, p
			if c == a && c != d && a != b {
				e0 = a
			}
			if a == b && a != c && b != d {
				e1 = b
			}
			if d == c && d != b && c != a {
				e2 = c
			}
			if b == d && b != a && d != c {
				e3 = d
			}
			top.WriteRune(e0)
			top.WriteRune(e1)
			bot.WriteRune(e2)
			bot.WriteRune(e3)
		}
		out = append(out, top.String(), bot.String())
	}
	return out
}

// render is the production path: parse the source rows, pack to quadrants.
func render(rows []string) []string {
	return companion.ParseBitmap(rows).ToQuadrant()
}

// mirrored runs the source through the production mirror, so a claim about a
// sprite surviving the flip is measured on the flip and not on the intention.
func mirrored(rows []string) []string {
	return bitmapRows(companion.ParseBitmap(rows).Mirrored())
}

func bitmapRows(b *companion.Bitmap) []string {
	out := make([]string, 0, b.H)
	for y := 0; y < b.H; y++ {
		var sb strings.Builder
		for x := 0; x < b.W; x++ {
			if b.On[y*b.W+x] {
				sb.WriteRune('#')
			} else {
				sb.WriteRune('.')
			}
		}
		out = append(out, sb.String())
	}
	return out
}

func symmetric(rows []string) bool {
	for _, r := range rows {
		c := []rune(r)
		for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
			if c[i] != c[j] {
				return false
			}
		}
	}
	return true
}

// cellDouble is the OTHER way to make a sprite twice as big, and the one a
// person means when they say "just scale the picture up": take the FINISHED
// quadrant render and blow every cell up into 2x2 cells. Here to be compared
// against, not to be used.
func cellDouble(q []string) []string {
	var out []string
	for _, row := range q {
		var top, bot strings.Builder
		for _, r := range row {
			i := quadIndex(r)
			top.WriteRune(blockIf(i&1 != 0))
			top.WriteRune(blockIf(i&2 != 0))
			bot.WriteRune(blockIf(i&4 != 0))
			bot.WriteRune(blockIf(i&8 != 0))
		}
		out = append(out, top.String(), bot.String())
	}
	return out
}

func blockIf(on bool) rune {
	if on {
		return '█'
	}
	return ' '
}

var quadRunes = []rune{
	' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛',
	'▗', '▚', '▐', '▜', '▄', '▙', '▟', '█',
}

func quadIndex(r rune) int {
	for i, q := range quadRunes {
		if q == r {
			return i
		}
	}
	return 0
}

// glyphHist counts which of the sixteen quadrant glyphs a render reaches for.
// The vocabulary a render can use is the whole question in section 3.
func glyphHist(q []string) (counts map[rune]int, distinct int) {
	counts = map[rune]int{}
	for _, row := range q {
		for _, r := range row {
			if r == ' ' {
				continue
			}
			counts[r]++
		}
	}
	return counts, len(counts)
}

func histLine(q []string) string {
	c, n := glyphHist(q)
	keys := make([]rune, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return c[keys[i]] > c[keys[j]] })
	var sb strings.Builder
	full, part := 0, 0
	for _, k := range keys {
		sb.WriteString(string(k))
		sb.WriteString(itoa(c[k]))
		sb.WriteString(" ")
		if k == '█' {
			full += c[k]
		} else {
			part += c[k]
		}
	}
	return itoa(n) + " glyphs, " + itoa(full+part) + " inked cells of which " +
		itoa(part) + " carry an edge: " + sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	if neg {
		return "-" + string(d)
	}
	return string(d)
}

// pairedRows counts how many of a source bitmap's rows are IDENTICAL to their
// partner in the vertical-halving pair (rows 2k and 2k+1).
//
// This is the number that decides whether an exact 2x is also an exact
// magnification of what is on screen today. ToQuadrant ORs each pair into one
// subpixel row, so where a pair differs, today's render is showing a merge the
// 2x does not have to make -- the 2x differs there by REVEALING a row, not by
// losing one.
func pairedRows(rows []string) (same, total int) {
	for k := 0; k+1 < len(rows); k += 2 {
		total++
		if rows[k] == rows[k+1] {
			same++
		}
	}
	return same, total
}

// diff counts cells that differ between two renders of the same footprint.
func diff(a, b []string) (differing, cells int) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for y := 0; y < n; y++ {
		ra, rb := []rune(a[y]), []rune(b[y])
		w := len(ra)
		if len(rb) < w {
			w = len(rb)
		}
		for x := 0; x < w; x++ {
			cells++
			if ra[x] != rb[x] {
				differing++
			}
		}
	}
	return differing, cells
}

// sideBySide prints renders in columns with a gutter, so a comparison is on the
// page and not in a claim about it.
func sideBySide(gap int, panes ...[]string) []string {
	widths := make([]int, len(panes))
	n := 0
	for i, p := range panes {
		for _, r := range p {
			if w := len([]rune(r)); w > widths[i] {
				widths[i] = w
			}
		}
		if len(p) > n {
			n = len(p)
		}
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var sb strings.Builder
		for j, p := range panes {
			r := ""
			if i < len(p) {
				r = p[i]
			}
			sb.WriteString(r)
			if j < len(panes)-1 {
				sb.WriteString(strings.Repeat(" ", widths[j]-len([]rune(r))+gap))
			}
		}
		out = append(out, strings.TrimRight(sb.String(), " "))
	}
	return out
}
