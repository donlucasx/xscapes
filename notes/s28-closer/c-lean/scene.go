package main

import (
	"fmt"
	"strings"

	"github.com/donlucasx/xscapes/internal/companion"
)

// A mock beach, so the silhouette is judged in the place it actually stands
// rather than on a blank page. The sea texture and the sand are stand-ins for
// colour, not for the shore's real glyphs -- what is being read here is whether
// the crab separates from its ground and where it sits in the frame.

type scene struct {
	w, h  int
	cells [][]rune
}

func newScene(w, h, waterline int) *scene {
	s := &scene{w: w, h: h}
	for y := 0; y < h; y++ {
		row := make([]rune, w)
		for x := 0; x < w; x++ {
			switch {
			case y < waterline-1:
				row[x] = '~'
			case y == waterline-1:
				row[x] = '='
			default:
				row[x] = '.'
			}
		}
		s.cells = append(s.cells, row)
	}
	return s
}

func (s *scene) put(x, y int, r rune) {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		return
	}
	s.cells[y][x] = r
}

// blit draws a sprite with the background rim plotRim gives it: every empty
// cell touching a painted one is cleared, which is what stops the sea writing
// glyphs into the crab's outline.
func (s *scene) blit(rows []string, x, y int) {
	filled := func(cx, cy int) bool {
		if cy < 0 || cy >= len(rows) {
			return false
		}
		r := []rune(rows[cy])
		return cx >= 0 && cx < len(r) && r[cx] != ' '
	}
	for cy := -1; cy <= len(rows); cy++ {
		for cx := -1; cx <= 12; cx++ {
			if filled(cx, cy) {
				continue
			}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if filled(cx+dx, cy+dy) {
						s.put(x+cx, y+cy, ' ')
					}
				}
			}
		}
	}
	for cy, row := range rows {
		for cx, r := range []rune(row) {
			if r != ' ' {
				s.put(x+cx, y+cy, r)
			}
		}
	}
}

func (s *scene) lines() []string {
	var out []string
	for _, r := range s.cells {
		out = append(out, string(r))
	}
	return out
}

// beach draws one frame of the whole composition: the crab where the layout
// puts it, the balloon above it, and the activity tail in the sand.
func beach(p pose, w, h, waterline, catX, drop int, say string) []string {
	s := newScene(w, h, waterline)
	top := h - 2 - 7 + drop
	rows := render(p, true, 0)
	// The tail is written from the left margin out to the companion's own
	// column, never into it.
	tail := "Read pace.go"
	for i, c := range tail {
		if 2+i < catX-2 {
			s.put(2+i, h-2, c)
		}
	}
	// Order matters and it is drawScene's order: the companion first, the
	// balloon on top of it. Drawn the other way round, the companion's own
	// background rim eats the balloon's bottom row -- which is exactly what my
	// first version of this mock did.
	s.blit(rows, catX, top)
	if say != "" {
		b := companion.MirrorTail(companion.Bubble(say))
		head := catX + 6 // HeadCol: (4+7+1)/2 and (3+8+1)/2 are both cell 6
		bx := head - companion.TailCol(b)
		if bw := len([]rune(b[0])); bx+bw > w {
			bx = w - bw
		}
		if bx < 0 {
			bx = 0
		}
		for dy, r := range b {
			for dx, c := range []rune(r) {
				s.put(bx+dx, top-len(b)+dy, c)
			}
		}
	}
	return s.lines()
}

// sideBySide prints two scenes with a gutter so the before and after are read
// as one picture.
func sideBySide(title string, a, b []string) {
	fmt.Println("\n" + title)
	for i := range a {
		fmt.Printf("|%s|   |%s|\n", a[i], b[i])
	}
	fmt.Println(strings.Repeat("-", 2*len(a[0])+8))
}

// filmstrip is the whole approach in the place it happens: each beat drawn at
// the column and row it would stand on, so the question "does this read as
// coming closer" is asked of the composition and not of the sprite.
func filmstrip(ps []pose, w, h, waterline, home int, dx, drop []int, say string) {
	var cols [][]string
	for i, p := range ps {
		s := ""
		if i == len(ps)-1 {
			s = say
		}
		cols = append(cols, beach(p, w, h, waterline, home+dx[i], drop[i], s))
	}
	head := ""
	for i, p := range ps {
		head += fmt.Sprintf("%-*s", w+3, fmt.Sprintf("%d %s  (x%+d y%+d)", i, p.name, dx[i], drop[i]))
	}
	fmt.Println("\n" + head)
	for r := 0; r < h; r++ {
		line := ""
		for _, c := range cols {
			line += fmt.Sprintf("%-*s", w+3, c[r])
		}
		fmt.Println(line)
	}
}
