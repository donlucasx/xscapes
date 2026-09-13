package companion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE EYES MUST BE ON THE HEAD, mid-stride included.
//
// This is the assertion that was missing. The cat had no front-view step, so
// Draw reached for CatWalk -- a 16-cell SIDE view -- drew it inside the 12-cell
// box, and then plotted the front-view eye glyphs at their usual cells, which
// landed on empty sky. It rendered 4,037 times across 31 recorded sessions and
// nobody ever reported it, because each occurrence lasts 0.28 s.
//
// ⚠ IT IS A ROW TEST, NOT AN 8-NEIGHBOURHOOD ONE. A 3x3 adjacency check passes
// on the broken sprite: the detached eye has no ink on its own row but does
// have ink diagonally below it. The eye has to be attached to the face it is
// supposed to be in, which means ink on ITS OWN ROW.
func TestTheEyesAreOnTheHeadInEveryStride(t *testing.T) {
	for _, arm := range []struct {
		name string
		make func() *Cat
		eyes [2]int
		row  int
	}{
		{"cat", NewCat, catEyeCells, 2},
		{"crab", NewCrab, crabEyeCells, crabEyeRow},
	} {
		for _, stepping := range []bool{false, true} {
			for _, mirror := range []bool{false, true} {
				c := arm.make()
				c.FaceLeft(mirror)
				c.SetStepping(stepping)
				cw, ch := c.Size()
				cv := canvas.New(cw+8, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				c.Draw(cv.Near(), 4, 1, 1.3, Working)

				row := 1 + arm.row
				var line []rune
				for x := 0; x < cv.W; x++ {
					r, _, _ := cv.ResolveAt(x, row, term.Profile256)
					if r == 0 {
						r = ' '
					}
					line = append(line, r)
				}
				for _, e := range arm.eyes {
					ex := 4 + e
					if mirror {
						ex = 4 + cw - 1 - e
					}
					// Ink on the eye's OWN row, on at least one side of it.
					left := ex > 0 && line[ex-1] != ' '
					right := ex+1 < len(line) && line[ex+1] != ' '
					if !left && !right {
						t.Errorf("%s stepping=%v mirror=%v: the eye at column %d has no body "+
							"beside it on its own row -- it is floating.\n   row: |%s|",
							arm.name, stepping, mirror, ex, strings.TrimRight(string(line), " "))
					}
				}
			}
		}
	}
}

// And the step must not move the head, or the eye cells and the balloon's
// pointer move with it.
func TestTheStrideMovesOnlyThePaws(t *testing.T) {
	render := func(stepping bool) []string {
		c := NewCat()
		c.FaceLeft(true)
		c.SetStepping(stepping)
		cw, ch := c.Size()
		cv := canvas.New(cw+8, ch+2, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Draw(cv.Near(), 4, 1, 1.3, Working)
		var out []string
		for y := 0; y < cv.H; y++ {
			var b strings.Builder
			for x := 0; x < cv.W; x++ {
				r, _, _ := cv.ResolveAt(x, y, term.Profile256)
				if r == 0 {
					r = ' '
				}
				b.WriteRune(r)
			}
			out = append(out, b.String())
		}
		return out
	}
	stand, step := render(false), render(true)
	for y := range stand {
		if y >= len(step) {
			break
		}
		// Cell rows 1..5 of the sprite are head, body and the tail's sweep.
		// Only the bottom row -- the paws -- may differ.
		if y < 1+6 && stand[y] != step[y] {
			t.Errorf("stride changed cell row %d, which is above the paws:\n   stand |%s|\n   step  |%s|",
				y-1, stand[y], step[y])
		}
	}
	if stand[1+6] == step[1+6] {
		t.Error("the paw row is identical standing and stepping -- there is no stride")
	}
	fmt.Printf("standing paws |%s|\nstepping paws |%s|\n",
		strings.TrimRight(stand[1+6], " "), strings.TrimRight(step[1+6], " "))
}
