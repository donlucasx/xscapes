package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/companion"
)

// show renders rows through the real ParseBitmap/ToQuadrant and prints them
// with a frame, so a mis-sized row or a lost pixel is visible rather than
// inferred.
func show(name string, rows []string) []string {
	b := companion.ParseBitmap(rows)
	q := b.ToQuadrant()
	w := 0
	for _, r := range q {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	fmt.Printf("%s  --  %dx%d px  =>  %d cells wide x %d tall\n", name, b.W, b.H, w, len(q))
	fmt.Printf("   +%s+\n", rep("-", w))
	for i, r := range q {
		fmt.Printf("%2d |%-*s|\n", i, w, r)
	}
	fmt.Printf("   +%s+\n\n", rep("-", w))
	return q
}

func rep(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func mirrored(rows []string) []string {
	return companion.ParseBitmap(rows).Mirrored().ToQuadrant()
}
