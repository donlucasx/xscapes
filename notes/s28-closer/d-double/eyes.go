package main

import "strings"

// The eye at 2x.
//
// A glyph cannot be scaled: a cell is a cell. On the shipped 12-cell crab the
// one-cell 'o' is a twelfth of the face and reads as an eye. On a 24-cell crab
// the same 'o' is a twenty-fourth of the face -- same ink, four times the shell
// around it -- and reads as a speck of dirt. The replacement is a small
// QUADRANT bitmap drawn as its own Sprite pass in the eye colour, the way the
// body is drawn, so it scales with the body.
//
// Sizing, from the medium: a cell is 2 source px wide and 4 tall, so a 2 cell
// by 2 cell eye is 4 px wide by 8 px tall. Every bitmap below is 4x8.
//
// The vertical halving is what shapes them. ToQuadrant ORs source rows 2k and
// 2k+1, so a subpixel is only OFF if BOTH rows of its pair are off. That is why
// every row below is written twice: not padding, it is the only way to control
// one subpixel.
//
// ⚠ THE RULE THAT DECIDES THE SHAPES: the eye pass PLOTS, and Plot replaces the
// whole cell. Wherever an eye cell has an off-subpixel over body ink, the coat
// is gone from that quarter and the SEA shows through the crab. Either every
// such cell gets PlotOn with the coat as its ground (that is what eyeGround and
// EyeFillCoat already exist for), or the eye is shaped so its lower cell row is
// SOLID -- and the lower row is the one that lands on the stalk. Every eye here
// takes the second option, so the eye pass punches zero holes and needs no
// ground fill. holePunch() in main.go checks that rather than trusting it.

// eyeOpen is the resting/working bead: a rounded cap on a solid base.
var eyeOpen = []string{
	".##.",
	".##.",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
}

// eyeAlert is the NeedsYou eye, and it is the whole point of the feature: the
// largest mark the box can hold. 'O' was trying to say this at 1x and could
// not, because 'O' and 'o' are the same cell.
var eyeAlert = []string{
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
}

// eyeShut is the blink and the resting '-': a lid over the bead, base kept
// solid so the stalk does not open a hole.
var eyeShut = []string{
	"....",
	"....",
	"####",
	"####",
	"####",
	"####",
	"####",
	"####",
}

// eyeDone is the content '^ ^': a chevron on a solid base.
var eyeDone = []string{
	".##.",
	".##.",
	"#..#",
	"#..#",
	"####",
	"####",
	"####",
	"####",
}

// eyeWorried drops the bead entirely and leaves the stalk tip. The pose is
// "the widest thing on the beach becomes the smallest", and the eye shrinking
// with it is the same sentence.
var eyeWorried = []string{
	"....",
	"....",
	"....",
	"....",
	"####",
	"####",
	"####",
	"####",
}

// overlay stamps art over a render at cell (x, y), the way a second Sprite pass
// paints over the body: the art REPLACES the cells it covers, which is exactly
// what Plot does and exactly why holes are possible. Cells the art leaves as a
// space pass through, matching Sprite.Draw's transparent space.
func overlay(q []string, x, y int, art []string) []string {
	out := append([]string(nil), q...)
	for dy, row := range art {
		ty := y + dy
		if ty < 0 || ty >= len(out) {
			continue
		}
		r := []rune(out[ty])
		for dx, c := range []rune(row) {
			tx := x + dx
			if tx < 0 || tx >= len(r) || c == ' ' {
				continue
			}
			r[tx] = c
		}
		out[ty] = string(r)
	}
	return out
}

// inkMap prints a render with the eye pass marked, so a monochrome dump still
// says which cells are eye colour and which are coat. The terminal separates
// them by colour; a transcript cannot, and a claim about an eye you cannot see
// in the evidence is not a measurement.
func inkMap(body []string, eyes [][3]interface{}) []string {
	out := append([]string(nil), body...)
	for _, e := range eyes {
		x, y, art := e[0].(int), e[1].(int), e[2].([]string)
		q := render(art)
		for dy, row := range q {
			ty := y + dy
			if ty < 0 || ty >= len(out) {
				continue
			}
			r := []rune(out[ty])
			for dx, c := range []rune(row) {
				tx := x + dx
				if tx < 0 || tx >= len(r) {
					continue
				}
				if c == ' ' {
					continue
				}
				r[tx] = 'E'
			}
			out[ty] = string(r)
		}
	}
	return out
}

// holePunch counts the subpixels the eye pass costs the body: body ink that
// the eye's own cell does NOT cover, and which therefore shows the scene
// instead of the coat once the eye replaces that cell.
func holePunch(bodySrc []string, eyeSrc []string, cellX, cellY int) (holes, eyeCells int) {
	body := render(bodySrc)
	eye := render(eyeSrc)
	// Work in subpixels: rebuild the halved grids the quadrant packer uses.
	bh := halve(bodySrc)
	eh := halve(eyeSrc)
	for dy, row := range eye {
		for dx := range []rune(row) {
			cy, cx := cellY+dy, cellX+dx
			if cy < 0 || cy >= len(body) || cx < 0 || cx >= len([]rune(body[cy])) {
				continue
			}
			eyeCells++
			for sy := 0; sy < 2; sy++ {
				for sx := 0; sx < 2; sx++ {
					bOn := at(bh, cx*2+sx, cy*2+sy)
					eOn := at(eh, dx*2+sx, dy*2+sy)
					if bOn && !eOn {
						holes++
					}
				}
			}
		}
	}
	return holes, eyeCells
}

// halve reproduces ToQuadrant's first step: OR source rows 2k and 2k+1.
func halve(rows []string) []string {
	out := make([]string, 0, (len(rows)+1)/2)
	for y := 0; y*2 < len(rows); y++ {
		a := []rune(rows[y*2])
		var b []rune
		if y*2+1 < len(rows) {
			b = []rune(rows[y*2+1])
		}
		var sb strings.Builder
		for x := range a {
			on := a[x] == '#'
			if b != nil && b[x] == '#' {
				on = true
			}
			if on {
				sb.WriteRune('#')
			} else {
				sb.WriteRune('.')
			}
		}
		out = append(out, sb.String())
	}
	return out
}

func at(rows []string, x, y int) bool {
	if y < 0 || y >= len(rows) {
		return false
	}
	r := []rune(rows[y])
	if x < 0 || x >= len(r) {
		return false
	}
	return r[x] == '#'
}
