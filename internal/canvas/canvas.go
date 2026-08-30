// Package canvas is the layer/alpha compositor. Terminals have no per-character
// opacity, so depth is faked: every layer carries an alpha, and a glyph's colour
// is blended from the background up through the stack. Low alpha = far away.
package canvas

import (
	"strings"

	"github.com/donlucasx/asciiscapes/internal/term"
)

// Layer alphas. Far is sparse, slow and washed out; near is dense, fast, solid.
const (
	AlphaFar  = 0.30
	AlphaMid  = 0.60
	AlphaNear = 1.00
)

type Cell struct {
	R   rune
	FG  term.RGB
	A   float64 // per-cell alpha, multiplied by the layer alpha
	Set bool
}

type Layer struct {
	W, H  int
	Alpha float64
	Cells []Cell
}

func NewLayer(w, h int, alpha float64) *Layer {
	return &Layer{W: w, H: h, Alpha: alpha, Cells: make([]Cell, w*h)}
}

func (l *Layer) Clear() {
	for i := range l.Cells {
		l.Cells[i].Set = false
	}
}

func (l *Layer) Plot(x, y int, r rune, fg term.RGB, a float64) {
	if x < 0 || y < 0 || x >= l.W || y >= l.H || a <= 0 {
		return
	}
	l.Cells[y*l.W+x] = Cell{R: r, FG: fg, A: a, Set: true}
}

type Canvas struct {
	W, H   int
	BG     []term.RGB
	Layers []*Layer

	buf []byte
}

// New builds a canvas with one layer per alpha given, ordered far to near.
func New(w, h int, alphas ...float64) *Canvas {
	c := &Canvas{W: w, H: h, BG: make([]term.RGB, w*h)}
	for _, a := range alphas {
		c.Layers = append(c.Layers, NewLayer(w, h, a))
	}
	return c
}

func (c *Canvas) Resize(w, h int) {
	if w == c.W && h == c.H {
		return
	}
	c.W, c.H = w, h
	c.BG = make([]term.RGB, w*h)
	for i, l := range c.Layers {
		c.Layers[i] = NewLayer(w, h, l.Alpha)
	}
}

func (c *Canvas) Clear() {
	for _, l := range c.Layers {
		l.Clear()
	}
}

// BGAt reads the background, so a scape can blend into it rather than over it.
func (c *Canvas) BGAt(x, y int) term.RGB {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return term.RGB{}
	}
	return c.BG[y*c.W+x]
}

func (c *Canvas) SetBG(x, y int, col term.RGB) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	c.BG[y*c.W+x] = col
}

// Far/Mid/Near are conveniences so scapes read as depth, not indices.
func (c *Canvas) Far() *Layer  { return c.Layers[0] }
func (c *Canvas) Mid() *Layer  { return c.Layers[1] }
func (c *Canvas) Near() *Layer { return c.Layers[2] }

// composite resolves one cell to a final glyph plus foreground colour.
func (c *Canvas) composite(i int) (rune, term.RGB) {
	bg := c.BG[i]
	fg := bg
	ch := ' '
	for _, l := range c.Layers {
		cell := l.Cells[i]
		if !cell.Set {
			continue
		}
		fg = fg.Blend(cell.FG, l.Alpha*cell.A)
		ch = cell.R
	}
	return ch, fg
}

// Render emits the frame. Colours are only re-stated when they change, which
// keeps the byte count (and the flicker) down.
func (c *Canvas) Render(p term.Profile) string {
	c.buf = c.buf[:0]
	var lastFG, lastBG term.RGB
	haveLast := false
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			i := y*c.W + x
			ch, fg := c.composite(i)
			bg := c.BG[i]
			if !haveLast || bg != lastBG {
				c.buf = p.AppendBG(c.buf, bg)
				lastBG = bg
			}
			if !haveLast || fg != lastFG {
				c.buf = p.AppendFG(c.buf, fg)
				lastFG = fg
			}
			haveLast = true
			c.buf = append(c.buf, string(ch)...)
		}
		c.buf = append(c.buf, term.Reset...)
		haveLast = false
		if y < c.H-1 {
			c.buf = append(c.buf, '\n')
		}
	}
	return string(c.buf)
}

// RenderPlain drops all colour. Used to check scene structure in a pipe, where
// escape codes would be noise rather than signal.
func (c *Canvas) RenderPlain() string {
	var b strings.Builder
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			ch, _ := c.composite(y*c.W + x)
			b.WriteRune(ch)
		}
		if y < c.H-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// RenderHTML writes the frame as an HTML page, one span per cell. This exists
// so a frame can be looked at rather than asserted about — glyph dumps hide
// every colour bug, and colour is most of what this project is.
//
// The charset meta is load-bearing: a file:// page without it decodes the
// block and braille glyphs as mojibake.
func (c *Canvas) RenderHTML(title string) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>`)
	b.WriteString(title)
	b.WriteString(`</title><style>`)
	b.WriteString(`body{margin:0;background:#111;display:flex;justify-content:center;padding:24px 0}`)
	b.WriteString(`pre{margin:0;font-family:Menlo,"SF Mono",monospace;font-size:18px;line-height:1.0;letter-spacing:0}`)
	b.WriteString(`</style></head><body><pre>`)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			i := y*c.W + x
			ch, fg := c.composite(i)
			bg := c.BG[i]
			b.WriteString(`<span style="color:`)
			writeHex(&b, fg)
			b.WriteString(`;background:`)
			writeHex(&b, bg)
			b.WriteString(`">`)
			switch ch {
			case '<':
				b.WriteString("&lt;")
			case '>':
				b.WriteString("&gt;")
			case '&':
				b.WriteString("&amp;")
			default:
				b.WriteRune(ch)
			}
			b.WriteString(`</span>`)
		}
		b.WriteByte('\n')
	}
	b.WriteString(`</pre></body></html>`)
	return b.String()
}

const hexDigits = "0123456789abcdef"

func writeHex(b *strings.Builder, c term.RGB) {
	b.WriteByte('#')
	for _, v := range [3]uint8{c.R, c.G, c.B} {
		b.WriteByte(hexDigits[v>>4])
		b.WriteByte(hexDigits[v&0x0f])
	}
}
