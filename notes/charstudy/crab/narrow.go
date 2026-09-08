package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/companion"
)

// narrowProbe answers a question the crab raises and the cat never did: the
// cat's ink is 9 cells inside a 12 cell box, so it has three columns of slack
// at the frame edge. The crab has none. This walks the width down and reports
// the first width at which a cell of ink falls outside the frame.
func narrowProbe() {
	type cand struct {
		name string
		rows []string
	}
	for _, c := range []cand{
		{"the cat", companion.CatBody},
		{"Pincer", compose28(pincerWork, pincerLower)},
		{"Hero", compose28(heroWork, heroLower)},
	} {
		q := companion.ParseBitmap(c.rows).Mirrored().ToQuadrant()
		// Which cell columns carry ink, in the mirrored (shipped) orientation.
		inked := map[int]bool{}
		for _, line := range q {
			for x, r := range []rune(line) {
				if r != ' ' {
					inked[x] = true
				}
			}
		}
		lost := -1
		for w := 40; w >= 8; w-- {
			// compose(), copied: the companion's margin grows with the width.
			right := 2 + w/32
			x := w - 12 - right
			if x < 0 {
				x = 0
			}
			out := 0
			for col := range inked {
				if x+col < 0 || x+col >= w {
					out++
				}
			}
			if out > 0 {
				lost = w
				break
			}
		}
		fmt.Printf("%-8s ink in cell columns %v\n         whole down to width %d, first loses ink at %d\n",
			c.name, sortedKeys(inked), lost+1, lost)
	}
}

func sortedKeys(m map[int]bool) []int {
	out := []int{}
	for k := range m {
		out = append(out, k)
	}
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
