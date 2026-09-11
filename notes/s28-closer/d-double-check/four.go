package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/companion"
)

// composite overlays the eye render onto the body render the way Sprite.Draw
// does it: a non-space rune REPLACES the cell outright (canvas.Plot), so a
// solid eye cell wipes whatever body was there.
func composite(body []string, eye []string, at [][2]int) []string {
	out := make([][]rune, len(body))
	for i, r := range body {
		out[i] = []rune(r)
	}
	for _, p := range at {
		cx, cy := p[0], p[1]
		for dy, er := range eye {
			for dx, r := range []rune(er) {
				if r == ' ' {
					continue
				}
				y, x := cy+dy, cx+dx
				if y >= 0 && y < len(out) && x >= 0 && x < len(out[y]) {
					out[y][x] = r
				}
			}
		}
	}
	res := make([]string, len(out))
	for i, r := range out {
		res[i] = string(r)
	}
	return res
}

func partFour() {
	fmt.Println("\n=== 12. THE EYE, COMPOSITED (what NeedsYou actually shows) ===")
	src := companion.ParseBitmap(append(append([]string{}, theirUpper...), theirLower...))
	eq := companion.ParseBitmap(theirEye).ToQuadrant()

	body := src.ToQuadrant()
	withEye := composite(body, eq, [][2]int{{8, 3}, {14, 3}})
	fmt.Println("theirs, eyeAlert at cols 8-9 / 14-15, rows 3-4 (eye cells are the SOLID blocks):")
	for i, r := range withEye {
		mark := "  "
		if i == 3 || i == 4 {
			mark = "->"
		}
		fmt.Printf("%s %2d |%s|\n", mark, i, r)
	}
	fmt.Println("\nfor scale, the shipped NeedsYou eye is the single glyph 'O' at cols {4,7} row 2:")
	sh := companion.ParseBitmap(append(append([]string{}, shippedAsk...), shippedLower...))
	sq := sh.ToQuadrant()
	shWithEye := composite(sq, []string{"O"}, [][2]int{{4, 2}, {7, 2}})
	for i, r := range shWithEye {
		mark := "  "
		if i == 2 {
			mark = "->"
		}
		fmt.Printf("%s %2d |%s|\n", mark, i, r)
	}

	// --- breathing ---
	fmt.Println("\n=== 13. BREATHING (lift = 2 source rows, the smallest step the medium has) ===")
	lift := func(b *companion.Bitmap, n int) *companion.Bitmap {
		f := b.Blank()
		for sy := 0; sy < b.H; sy++ {
			for sx := 0; sx < b.W; sx++ {
				if sy+n < b.H && b.On[(sy+n)*b.W+sx] {
					f.Set(sx, sy)
				}
			}
		}
		return f
	}
	a := src.ToQuadrant()
	bq := lift(src, 2).ToQuadrant()
	diff := 0
	for i := range a {
		if a[i] != bq[i] {
			diff++
		}
	}
	fmt.Printf("2x sprite: lift 0 vs lift 2 changes %d of %d cell rows\n", diff, len(a))
	sa := sh.ToQuadrant()
	sb := lift(sh, 2).ToQuadrant()
	sdiff := 0
	for i := range sa {
		if sa[i] != sb[i] {
			sdiff++
		}
	}
	fmt.Printf("shipped:   lift 0 vs lift 2 changes %d of %d cell rows\n", sdiff, len(sa))
	fmt.Println("lift=2 on theirs, rendered:")
	for i, r := range bq {
		fmt.Printf("   %2d |%s|\n", i, r)
	}

	// --- sand cost, from compose()'s real arithmetic ---
	fmt.Println("\n=== 14. SAND COST (compose(): SandTo = catX-1-span, catX = w-catW-(2+w/32)) ===")
	fmt.Printf("%-6s %8s %8s %8s %8s\n", "w", "span", "sand@12", "sand@24", "loss")
	for _, w := range []int{40, 80, 111, 124, 143, 153} {
		right := 2 + w/32
		sp := paceSpanLocal(w)
		band := func(catW int) int {
			catX := w - catW - right
			if catX < 0 {
				catX = 0
			}
			return (catX - 1 - sp) - 2
		}
		fmt.Printf("%-6d %8d %8d %8d %8d\n", w, sp, band(12), band(24), band(12)-band(24))
	}
}

// paceSpanLocal is pace.go:43 verbatim.
func paceSpanLocal(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}
