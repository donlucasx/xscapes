package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// dump prints a body as the quadrant glyphs it actually becomes, so the
// silhouette can be judged before a frame is spent on it. Build the instrument
// before trusting the picture.
func dump(name string, rows []string, w, h int) {
	if len(rows) != h {
		fmt.Printf("%s: BAD -- %d rows, want %d\n", name, len(rows), h)
		return
	}
	for i, r := range rows {
		if len(r) != w {
			fmt.Printf("%s: BAD -- row %d is %d wide, want %d: %q\n", name, i, len(r), w, r)
			return
		}
	}
	q := companion.ParseBitmap(rows).ToQuadrant()
	fmt.Printf("--- %s (%d x %d cells) ---\n", name, w/2, h/4)
	for _, line := range q {
		fmt.Printf("  |%s|\n", strings.ReplaceAll(line, " ", "."))
	}
	fmt.Println()
}

// extent measures how much of the 12x7 box a body actually fills: the ink
// bounding box in cells and how many cells carry ink. "Is it as big as the cat"
// is a question about ink, not about the box -- every companion shares the box.
func extent(name string, rows []string) (w, h, filled int) {
	q := companion.ParseBitmap(rows).ToQuadrant()
	minX, maxX, minY, maxY := 99, -1, 99, -1
	for y, line := range q {
		// []rune, not a byte range: a quadrant glyph is three bytes, so ranging
		// the string gives byte offsets and reports a 12-cell body as 34 wide.
		for x, r := range []rune(line) {
			if r == ' ' {
				continue
			}
			filled++
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	return maxX - minX + 1, maxY - minY + 1, filled
}
