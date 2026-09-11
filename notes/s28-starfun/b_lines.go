package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/term"
)

// candidateB -- CONSTELLATION LINES. Faint strokes joining the stars in the
// order they lit.
//
// Against the rule: order is a SECOND variable on a channel that already
// carries count. The brief allows that only if nothing else needs the channel,
// and something does. But the rule is not what kills it -- the geometry is, and
// the geometry is a consequence of the fix that shipped on 2026-09-10.
func candidateB() {
	fmt.Println("\n\n==== 3. CANDIDATE B -- CONSTELLATION LINES ====")
	base := warm(0, 0, CTX)

	fmt.Println("\n  THE MEASUREMENT THAT DECIDES IT, before any picture: a constellation line")
	fmt.Println("  is only short if consecutive slots are NEAR each other, and no placement this")
	fmt.Println("  channel can use does that. The golden ratio at HEAD deliberately puts each new")
	fmt.Println("  index in the largest remaining gap (measured: mean segment 57 cells, longest")
	fmt.Println("  76). The tree's blue noise scatters them (mean 36, longest 83). Both give long")
	fmt.Println("  strokes, because both are trying to keep the stars APART -- which is the thing")
	fmt.Println("  that makes them countable.")
	fmt.Printf("  %-8s %-10s %-10s %-10s %-10s %s\n", "stars", "segments", "mean len", "max len", "line cells", "% of sky cells")
	for _, n := range []int{12, 19, 26, 32} {
		ps := placeLive(n)
		segs, cells := lineCells(ps)
		var tot, mx float64
		for _, s := range segs {
			if s > mx {
				mx = s
			}
			tot += s
		}
		skyCells := W * base.hy
		fmt.Printf("  %-8d %-10d %-10.1f %-10.0f %-10d %.1f%%\n",
			n, len(segs), tot/math.Max(1, float64(len(segs))), mx, len(cells),
			100*float64(len(cells))/float64(skyCells))
	}

	fmt.Println("\n  AND WHAT THE INK LANDS ON, at 26 and 32:")
	fmt.Printf("  %-8s %-12s %-14s %-16s %s\n", "stars", "crossings", "cells on disc", "passes within 1", "rows spanned")
	for _, n := range []int{26, 32} {
		ps := placeLive(n)
		_, cells := lineCells(ps)
		seen, cross, onDisc, near := map[pt]int{}, 0, 0, 0
		star := map[pt]bool{}
		for _, p := range ps {
			star[p] = true
		}
		lo, hi := 99, -1
		for _, c := range cells {
			seen[c]++
			if seen[c] == 2 {
				cross++
			}
			if base.sh.DiscCovers(c.x, c.y) {
				onDisc++
			}
			if c.y < lo {
				lo = c.y
			}
			if c.y > hi {
				hi = c.y
			}
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				if star[pt{c.x + d[0], c.y + d[1]}] {
					near++
					break
				}
			}
		}
		fmt.Printf("  %-8d %-12d %-14d %-16d %d..%d\n", n, cross, onDisc, near, lo, hi)
	}

	for _, n := range []int{12, 26} {
		s := warm(0, 0, CTX)
		ps := placeLive(n)
		_, cells := lineCells(ps)
		for _, c := range cells {
			if c.x < 0 || c.x >= W || c.y < 0 || c.y >= H {
				continue
			}
			s.c.Near().Plot(c.x, c.y, lineGlyph(c, cells), s.pal.Star, 0.45)
		}
		for _, p := range ps {
			s.plotStar(p, s.pal.Star, Floor)
		}
		show(fmt.Sprintf("  B at %d stars, joined in the order they lit:", n), asciiQuiet(s))
	}

	fmt.Println("  ⇒ REJECT B, on three counts and each is measured above:")
	fmt.Println("    1. The strokes cross the DISC. The moon carries context remaining, and a")
	fmt.Println("       line over its face reads as part of it -- the same defect his own")
	fmt.Println("       2026-09-06 report caught with one tan speck on the sun.")
	fmt.Println("    2. The ink outweighs the channel. At 26 stars the lines put 5-10x more")
	fmt.Println("       marks in the sky than the stars themselves, and the count is the channel.")
	fmt.Println("    3. A line passing a star turns it into a KINK, and a kink is not countable.")
	fmt.Println("  ⚠ AND NO PLACEMENT SAVES IT. Measured both ways: at HEAD's golden ratio 26")
	fmt.Println("    stars put 1391 line cells in the sky (101% of its cells, 352 crossings, 34")
	fmt.Println("    on the disc); on the tree's blue noise, 843 cells (61%, 248 crossings, 28 on")
	fmt.Println("    the disc). Better, still hopeless. Short lines need consecutive stars to be")
	fmt.Println("    NEIGHBOURS, which means placing them in lit order -- which is exactly the")
	fmt.Println("    left-to-right sky he could not see at all before 2026-09-10.")
}

// lineCells walks a Bresenham segment between each consecutive pair.
func lineCells(ps []pt) (lens []float64, cells []pt) {
	for i := 1; i < len(ps); i++ {
		a, b := ps[i-1], ps[i]
		lens = append(lens, math.Hypot(float64(b.x-a.x), 2*float64(b.y-a.y)))
		dx, dy := abs(b.x-a.x), -abs(b.y-a.y)
		sx, sy := sign(b.x-a.x), sign(b.y-a.y)
		err := dx + dy
		x, y := a.x, a.y
		for {
			if !(x == a.x && y == a.y) && !(x == b.x && y == b.y) {
				cells = append(cells, pt{x, y})
			}
			if x == b.x && y == b.y {
				break
			}
			e2 := 2 * err
			if e2 >= dy {
				err += dy
				x += sx
			}
			if e2 <= dx {
				err += dx
				y += sy
			}
		}
	}
	return
}

func lineGlyph(c pt, all []pt) rune {
	up, down := false, false
	for _, o := range all {
		if o.x == c.x-1 || o.x == c.x+1 {
			if o.y < c.y {
				up = true
			}
			if o.y > c.y {
				down = true
			}
		}
	}
	switch {
	case up && !down:
		return '/'
	case down && !up:
		return '\\'
	}
	return '-'
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

var _ = term.Profile256
