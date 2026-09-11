package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// show renders a bitmap the way the scape does -- ToQuadrant -- and prints it
// with a ruler, so a silhouette can be read off the page rather than imagined.
func show(title string, rows []string) []string {
	src := companion.ParseBitmap(rows)
	q := src.ToQuadrant()
	w := len([]rune(q[0]))
	fmt.Printf("%s  --  %dx%d px  =  %d cells wide x %d tall\n", title, src.W, src.H, w, len(q))
	ruler(w)
	for i, r := range q {
		fmt.Printf("  %d |%s|\n", i, r)
	}
	ruler(w)
	fmt.Println()
	return q
}

// showMirrored is the shipped composition: the sprite is flipped for the
// right-hand layout, so every candidate has to be read that way too.
func showMirrored(title string, rows []string) {
	src := companion.ParseBitmap(rows).Mirrored()
	q := src.ToQuadrant()
	w := len([]rune(q[0]))
	fmt.Printf("%s (MIRRORED, as shipped)  --  %d cells wide x %d tall\n", title, w, len(q))
	ruler(w)
	for i, r := range q {
		fmt.Printf("  %d |%s|\n", i, r)
	}
	ruler(w)
	fmt.Println()
}

func ruler(w int) {
	var a, b strings.Builder
	a.WriteString("    ")
	b.WriteString("    ")
	for i := 0; i < w; i++ {
		a.WriteByte(byte('0' + (i/10)%10))
		b.WriteByte(byte('0' + i%10))
	}
	if w >= 10 {
		fmt.Println(a.String())
	}
	fmt.Println(b.String())
}

// join stacks an upper half onto a lower half, the repo's own arrangement.
func join(upper, lower []string) []string {
	return append(append([]string{}, upper...), lower...)
}

// checkWidths panics the same way ParseBitmap would, but names the row, so a
// ragged draft is caught while it is being typed rather than at render time.
func checkWidths(name string, rows []string) {
	w := len([]rune(rows[0]))
	for i, r := range rows {
		if len([]rune(r)) != w {
			panic(fmt.Sprintf("%s row %d is %d wide, want %d", name, i, len([]rune(r)), w))
		}
	}
}

// nearest is nearest-neighbour scaling, here to be measured and rejected.
func nearest(rows []string, dw, dh int) []string {
	sh, sw := len(rows), len([]rune(rows[0]))
	out := make([]string, dh)
	for y := 0; y < dh; y++ {
		src := []rune(rows[y*sh/dh])
		var sb strings.Builder
		for x := 0; x < dw; x++ {
			sb.WriteRune(src[x*sw/dw])
		}
		out[y] = sb.String()
	}
	return out
}

// symmetry counts rows whose ink is not left-right symmetric. The crab is a
// front-on animal: every row of the authored art is a palindrome, and that is
// the property nearest-neighbour destroys.
func symmetry(rows []string) (bad int, first int) {
	first = -1
	for y, r := range rows {
		c := []rune(r)
		ok := true
		for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
			if (c[i] == '#') != (c[j] == '#') {
				ok = false
				break
			}
		}
		if !ok {
			bad++
			if first < 0 {
				first = y
			}
		}
	}
	return
}

// rowRuns reports the widest ink run per row, which is how a shell's taper is
// read as a number instead of by eye.
func rowRuns(rows []string) []int {
	out := make([]int, len(rows))
	for y, r := range rows {
		best, n := 0, 0
		for _, c := range r {
			if c == '#' {
				n++
				if n > best {
					best = n
				}
			} else {
				n = 0
			}
		}
		out[y] = best
	}
	return out
}

// quad is ToQuadrant, reached through the package so the draft and the product
// cannot drift.
func quad(rows []string) []string { return companion.ParseBitmap(rows).ToQuadrant() }
