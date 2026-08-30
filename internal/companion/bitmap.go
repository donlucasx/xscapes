package companion

import "strings"

// Bitmap is a 1-bit sprite source. Authoring here rather than in characters is
// the whole point: a character grid gives 1:2 stretched pixels at the lowest
// resolution available, while braille and quadrants give SQUARE pixels at 8x
// and 2x. Draw once at the highest resolution, render three ways.
type Bitmap struct {
	W, H int
	On   []bool
}

// ParseBitmap reads '#' as ink, anything else as empty. Rows must be equal
// length; ragged art is a bug that renders as a subtle shear, so it panics.
func ParseBitmap(rows []string) *Bitmap {
	h := len(rows)
	w := len([]rune(rows[0]))
	b := &Bitmap{W: w, H: h, On: make([]bool, w*h)}
	for y, row := range rows {
		r := []rune(row)
		if len(r) != w {
			panic("companion: ragged bitmap row " + itoa(y))
		}
		for x, c := range r {
			b.On[y*w+x] = c == '#'
		}
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// Set turns a pixel on, ignoring out-of-bounds so callers can draw curves
// without clipping arithmetic at every step.
func (b *Bitmap) Set(x, y int) {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return
	}
	b.On[y*b.W+x] = true
}

// Mirrored flips horizontally. Compose a frame facing right and mirror the
// whole thing -- mirroring the body alone would leave the legs walking backwards.
func (b *Bitmap) Mirrored() *Bitmap {
	m := b.Blank()
	for y := 0; y < b.H; y++ {
		for x := 0; x < b.W; x++ {
			if b.at(x, y) {
				m.Set(b.W-1-x, y)
			}
		}
	}
	return m
}

// Blank returns an empty bitmap of the same size.
func (b *Bitmap) Blank() *Bitmap {
	return &Bitmap{W: b.W, H: b.H, On: make([]bool, b.W*b.H)}
}

func (b *Bitmap) at(x, y int) bool {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return false
	}
	return b.On[y*b.W+x]
}

// ToBraille packs 2x4 pixels per cell. Dot bits are not in reading order --
// the low six are the historic 6-dot layout and dots 7,8 were appended later.
func (b *Bitmap) ToBraille() []string {
	var bits = [2][4]byte{
		{0x01, 0x02, 0x04, 0x40}, // left column: dots 1,2,3,7
		{0x08, 0x10, 0x20, 0x80}, // right column: dots 4,5,6,8
	}
	var out []string
	for cy := 0; cy < (b.H+3)/4; cy++ {
		var sb strings.Builder
		for cx := 0; cx < (b.W+1)/2; cx++ {
			var m byte
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					if b.at(cx*2+dx, cy*4+dy) {
						m |= bits[dx][dy]
					}
				}
			}
			sb.WriteRune(rune(0x2800 + int(m)))
		}
		out = append(out, sb.String())
	}
	return out
}

// quadrants indexed by TL=1 TR=2 BL=4 BR=8.
var quadrants = [16]rune{
	' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛',
	'▗', '▚', '▐', '▜', '▄', '▙', '▟', '█',
}

// ToQuadrant packs 2x2 pixels per cell, halving the source vertically first so
// the result lands on the same character footprint as ToBraille.
func (b *Bitmap) ToQuadrant() []string {
	half := &Bitmap{W: b.W, H: (b.H + 1) / 2, On: make([]bool, b.W*((b.H+1)/2))}
	for y := 0; y < half.H; y++ {
		for x := 0; x < b.W; x++ {
			half.On[y*b.W+x] = b.at(x, y*2) || b.at(x, y*2+1)
		}
	}
	var out []string
	for cy := 0; cy < (half.H+1)/2; cy++ {
		var sb strings.Builder
		for cx := 0; cx < (half.W+1)/2; cx++ {
			var i int
			if half.at(cx*2, cy*2) {
				i |= 1
			}
			if half.at(cx*2+1, cy*2) {
				i |= 2
			}
			if half.at(cx*2, cy*2+1) {
				i |= 4
			}
			if half.at(cx*2+1, cy*2+1) {
				i |= 8
			}
			sb.WriteRune(quadrants[i])
		}
		out = append(out, sb.String())
	}
	return out
}

// charRamp is nine steps of ink, ASCII only.
var charRamp = []rune(" .:-=+*#@")

// ToChars collapses 2x4 pixels into one character chosen by how much of the
// block is filled. This is the medium I had been hand-drawing in; generating it
// from the same source makes the resolution difference honest.
func (b *Bitmap) ToChars() []string {
	var out []string
	for cy := 0; cy < (b.H+3)/4; cy++ {
		var sb strings.Builder
		for cx := 0; cx < (b.W+1)/2; cx++ {
			n := 0
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					if b.at(cx*2+dx, cy*4+dy) {
						n++
					}
				}
			}
			sb.WriteRune(charRamp[n*(len(charRamp)-1)/8])
		}
		out = append(out, sb.String())
	}
	return out
}
