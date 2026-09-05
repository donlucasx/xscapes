// Package canvas is the layer/alpha compositor. Terminals have no per-character
// opacity, so depth is faked: every layer carries an alpha, and a glyph's colour
// is blended from the background up through the stack. Low alpha = far away.
package canvas

import (
	"strconv"
	"strings"

	"github.com/donlucasx/xscapes/internal/term"
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
	// BG is the cell's own background, when HasBG: a glyph that brings its
	// ground with it instead of showing the scene through. The companion's
	// eyes are the case -- characters plotted in gaps of the body bitmap,
	// which showed the sea behind the head by day.
	BG    term.RGB
	HasBG bool
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

// PlotOn is Plot with the cell's own background. The background is blended
// toward the scene by the same alpha as the glyph, so a far layer's ground
// recedes with it; at the near layer it is exact.
func (l *Layer) PlotOn(x, y int, r rune, fg, bg term.RGB, a float64) {
	if x < 0 || y < 0 || x >= l.W || y >= l.H || a <= 0 {
		return
	}
	l.Cells[y*l.W+x] = Cell{R: r, FG: fg, A: a, Set: true, BG: bg, HasBG: true}
}

type Canvas struct {
	W, H   int
	BG     []term.RGB
	Layers []*Layer

	// ramp binds a background cell to the gradient it is part of, so that on
	// 256 the cell takes its tone from the ramp's path rather than from
	// rounding its own colour. Nil where the background is a flat tone.
	ramp []rampRef

	// half is a background cell painted as two colours, top and bottom, for
	// a shape whose edge falls inside a row: the moon's tips. Unset where
	// the background is one colour.
	half []halfRef

	buf []byte
}

// halfRef is a cell's two backgrounds, for a shape edge that falls mid-row.
type halfRef struct {
	up, down term.RGB
	set      bool
}

// rampRef is a cell's place in a ramp: the ramp, and the span of it the
// cell covers, top edge to bottom edge, so the two halves of the cell can
// each take the tone the path puts there.
type rampRef struct {
	r      *term.Ramp
	t0, t1 float64
}

// New builds a canvas with one layer per alpha given, ordered far to near.
func New(w, h int, alphas ...float64) *Canvas {
	c := &Canvas{W: w, H: h, BG: make([]term.RGB, w*h), ramp: make([]rampRef, w*h), half: make([]halfRef, w*h)}
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
	c.ramp = make([]rampRef, w*h)
	c.half = make([]halfRef, w*h)
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
	c.ramp[y*c.W+x] = rampRef{}
	c.half[y*c.W+x] = halfRef{}
}

// SetBGHalves paints a cell's top half one colour and its bottom half another,
// with U+2580. This is how a shape gets an edge that falls inside a row: the
// moon's tips, which at a radius just under two rows used to fall outside the
// disc entirely and leave a rectangle. BGAt reports the mean of the two, so
// what blends INTO the cell (a star, the glitter) sees one colour.
func (c *Canvas) SetBGHalves(x, y int, up, down term.RGB) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	i := y*c.W + x
	c.BG[i] = up.Blend(down, 0.5)
	c.ramp[i] = rampRef{}
	c.half[i] = halfRef{up: up, down: down, set: true}
}

// SetBGRamp paints a cell as part of a ramp. t0 and t1 are where the cell's
// top and bottom edges fall along it, 0 at the ramp's start and 1 at its end.
// The background reads as the true colour at the cell's centre, which is what
// a truecolor terminal paints and what BGAt reports; on 256 the cell resolves
// through the ramp's path instead of through the quantiser.
func (c *Canvas) SetBGRamp(x, y int, r *term.Ramp, t0, t1 float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H || r == nil {
		return
	}
	i := y*c.W + x
	c.BG[i] = r.True((t0 + t1) / 2)
	c.ramp[i] = rampRef{r: r, t0: t0, t1: t1}
	// Every paint clears the other kinds of paint. This line was missing for
	// ten hours and the moon's tips from the night stayed in the sky until
	// eleven in the morning, as grey notches around the sun.
	c.half[i] = halfRef{}
}

// Far/Mid/Near are conveniences so scapes read as depth, not indices.
func (c *Canvas) Far() *Layer  { return c.Layers[0] }
func (c *Canvas) Mid() *Layer  { return c.Layers[1] }
func (c *Canvas) Near() *Layer { return c.Layers[2] }

// composite resolves one cell to a final glyph plus foreground colour, and
// says whether any layer actually put something there.
func (c *Canvas) composite(i int) (rune, term.RGB, bool) {
	ch, fg, _, _, set := c.compositeBG(i)
	return ch, fg, set
}

// compositeBG is composite with the cell's own ground, when the topmost glyph
// brought one (PlotOn): own is that ground blended toward the scene by the
// glyph's alpha, and hasOwn says it applies. A glyph plotted over an own-ground
// cell without a ground of its own keeps the ground under it.
func (c *Canvas) compositeBG(i int) (ch rune, fg, own term.RGB, hasOwn, set bool) {
	bg := c.BG[i]
	fg = bg
	ch = ' '
	own = bg
	for _, l := range c.Layers {
		cell := l.Cells[i]
		if !cell.Set {
			continue
		}
		if cell.HasBG {
			own = own.Blend(cell.BG, l.Alpha*cell.A)
			hasOwn = true
		}
		fg = fg.Blend(cell.FG, l.Alpha*cell.A)
		ch = cell.R
		set = true
	}
	return ch, fg, own, hasOwn, set
}

// resolved is one cell ready to draw. glyph says the foreground came from a
// layer rather than from shading, which decides whether it gets the chroma
// boost -- a shade block is half a background and must not be lifted away from
// the other half.
type resolved struct {
	ch     rune
	fg, bg term.RGB
	glyph  bool
}

// halfStep is how far apart two rows may be, in luma, and still be treated as
// part of one ramp rather than as an edge.
//
// It matters at the horizon: there the row above is bright sky and the row
// below is dark water, and splitting that cell in half would paint a band of
// sky-into-sea mix along the waterline. That is a halo, not smoothing. Inside a
// ramp consecutive rows are a few luma apart, so one number separates the two
// cases.
const halfStep = 26

func lumaOf(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// halves returns the colours a quarter of a row above and below this cell's
// centre, or ok=false if this row is an edge rather than part of a ramp.
func (c *Canvas) halves(x, y int) (up, down term.RGB, ok bool) {
	i := y*c.W + x
	up, down = c.BG[i], c.BG[i]
	// A neighbour painted as two halves records the MEAN of them as its
	// background, which is a bookkeeping value and not a tone anyone sees;
	// blending toward it split a flat rim cell of the moon into a grey over
	// an olive. A shape's edge is not part of any ramp.
	if (y > 0 && c.half[i-c.W].set) || (y < c.H-1 && c.half[i+c.W].set) {
		return up, down, false
	}
	if y > 0 {
		a := c.BG[i-c.W]
		if d := lumaOf(a) - lumaOf(c.BG[i]); d > halfStep || d < -halfStep {
			return up, down, false
		}
		up = term.Lerp(a, c.BG[i], 0.75)
	}
	if y < c.H-1 {
		b := c.BG[i+c.W]
		if d := lumaOf(b) - lumaOf(c.BG[i]); d > halfStep || d < -halfStep {
			return up, down, false
		}
		down = term.Lerp(c.BG[i], b, 0.25)
	}
	return up, down, true
}

// resolve is the single place a cell becomes colours. Both renderers go
// through it so a preview and a terminal cannot disagree about a frame.
//
// Two things happen here and only on 256, and only where no layer drew
// anything -- a glyph owns its cell and a background showing through one cannot
// also be two colours.
//
//  1. A cell is split in half. U+2580 paints the top half in the foreground and
//     leaves the bottom to the background, so one row can carry TWO steps of a
//     vertical ramp. This costs nothing in tone accuracy and adds no texture:
//     it is the same colours, placed twice as finely.
//  2. Where the two halves land on the same colour anyway, the cell falls back
//     to blending BETWEEN two cube colours with a shade block.
//
// The order is deliberate. Splitting is free and invisible; shading is a real
// pattern on the screen, so it is the second choice rather than the first.
func (c *Canvas) resolve(x, y int, p term.Profile) resolved {
	i := y*c.W + x
	ch, fg, own, hasOwn, set := c.compositeBG(i)
	bg := c.BG[i]
	if hasOwn {
		// A glyph that brought its own ground owns the whole cell: no split,
		// no ramp tone, no shading. The eye sits on fur whatever is behind
		// the head.
		return resolved{ch: ch, fg: fg, bg: own, glyph: true}
	}
	if hf := c.half[i]; hf.set {
		// A shape's edge inside the row. Both halves are backgrounds and take
		// the background quantiser. The halves win over a glyph: a far star
		// on the moon's tip was rendering as one flat cell of the mean, a
		// dark dot on the crown, and a star behind the moon's edge is not a
		// thing anyone misses.
		up, down := p.Quantise(hf.up, false), p.Quantise(hf.down, false)
		if up != down {
			ch, fg, bg := term.Split(up, down)
			return resolved{ch: ch, fg: fg, bg: bg}
		}
		return resolved{ch: ' ', fg: up, bg: up, glyph: true}
	}
	if p == term.Profile256 && term.Ramps && c.ramp[i].r != nil {
		// A cell of a ramp takes the path's tones, not the quantiser's. Both
		// halves, where the path puts an edge inside the cell; and the same
		// tone under a glyph as the open cells beside it, or every star
		// would sit in a one-cell halo of a slightly different blue.
		rr := c.ramp[i]
		mid := rr.r.Tone((rr.t0 + rr.t1) / 2)
		if !set {
			if term.Shading {
				up := rr.r.Tone(rr.t0 + 0.25*(rr.t1-rr.t0))
				down := rr.r.Tone(rr.t0 + 0.75*(rr.t1-rr.t0))
				if up != down {
					ch, fg, bg := term.Split(up, down)
					return resolved{ch: ch, fg: fg, bg: bg}
				}
			}
			return resolved{ch: ' ', fg: mid, bg: mid, glyph: true}
		}
		return resolved{ch: ch, fg: fg, bg: mid, glyph: true}
	}
	if p == term.Profile256 && !set && term.Shading {
		if up, down, ok := c.halves(x, y); ok {
			ui, di := up.Index256Keeping(), down.Index256Keeping()
			if ui != di {
				// A band edge falls inside this cell: put it there.
				ch, fg, bg := term.Split(term.FromIndex256(ui), term.FromIndex256(di))
				return resolved{ch: ch, fg: fg, bg: bg}
			}
		}
		if sbg, sfg, r, ok := term.ShadeCell(bg); ok {
			return resolved{ch: r, fg: sfg, bg: sbg}
		}
	}
	return resolved{ch: ch, fg: fg, bg: bg, glyph: true}
}

// ResolveAt reports what will actually be drawn in one cell for a profile:
// the glyph, its colour, and the background. Studies and tests need it because
// on 256 a cell can carry two colours, and anything that reads c.BG alone is
// blind to half of what the terminal shows.
func (c *Canvas) ResolveAt(x, y int, p term.Profile) (rune, term.RGB, term.RGB) {
	r := c.resolve(x, y, p)
	fg := r.fg
	if r.glyph {
		fg = p.Quantise(fg, true)
	}
	return r.ch, fg, p.Quantise(r.bg, false)
}

// Render emits the frame. Colours are only re-stated when they change, which
// keeps the byte count (and the flicker) down.
func (c *Canvas) Render(p term.Profile) string {
	c.buf = c.buf[:0]
	var lastFG, lastBG term.RGB
	haveLast := false
	lastGlyph := false
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			r := c.resolve(x, y, p)
			if !haveLast || r.bg != lastBG {
				c.buf = p.AppendBG(c.buf, r.bg)
				lastBG = r.bg
			}
			if !haveLast || r.fg != lastFG || r.glyph != lastGlyph {
				if r.glyph {
					c.buf = p.AppendFG(c.buf, r.fg)
				} else {
					c.buf = p.AppendFGRaw(c.buf, r.fg)
				}
				lastFG, lastGlyph = r.fg, r.glyph
			}
			haveLast = true
			c.buf = append(c.buf, string(r.ch)...)
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
			// Deliberately the un-shaded glyph: -plain is for reading the
			// scene's structure, and shade blocks are tone, not structure.
			ch, _, _ := c.composite(y*c.W + x)
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
	return HTMLPage(title, c.HTMLFragment(18))
}

// HTMLFragment is the frame as a bare <pre>. Font size is a parameter so the
// same glyph data can be shown large enough to judge the drawing AND at the
// size it will really appear — without asking anyone to mentally rescale.
// HTMLFragmentAs renders as a given profile would actually show it, so a
// preview of a 256-colour terminal is honest rather than flattering.
func (c *Canvas) HTMLFragmentAs(fontPx int, p term.Profile) string {
	return c.htmlFragment(fontPx, p, true)
}

func (c *Canvas) HTMLFragment(fontPx int) string {
	return c.htmlFragment(fontPx, term.ProfileTrueColor, false)
}

// HTMLPalette shares colour classes across every fragment rendered with it,
// so a page of several frames carries each (foreground, background) pair once
// in a <style> block instead of once per run. A five-frame page went from
// 112KB to under half that; it exists because the submission page has to be
// pasted into a chat box.
type HTMLPalette struct {
	idx   map[[2]term.RGB]int
	pairs [][2]term.RGB
}

func (h *HTMLPalette) class(fg, bg term.RGB) int {
	if h.idx == nil {
		h.idx = map[[2]term.RGB]int{}
	}
	k := [2]term.RGB{fg, bg}
	if i, ok := h.idx[k]; ok {
		return i
	}
	i := len(h.pairs)
	h.idx[k] = i
	h.pairs = append(h.pairs, k)
	return i
}

// CSS is the rules for every class handed out so far. Put it in the page's
// <style> after the last fragment is rendered.
func (h *HTMLPalette) CSS() string {
	var b strings.Builder
	for i, k := range h.pairs {
		b.WriteString(".p")
		b.WriteString(strconv.Itoa(i))
		b.WriteString("{color:")
		writeHex(&b, k[0])
		b.WriteString(";background:")
		writeHex(&b, k[1])
		b.WriteString("}")
	}
	return b.String()
}

// HTMLFragmentClassed is HTMLFragmentAs with the colours taken out of the
// markup and into pal.
func (c *Canvas) HTMLFragmentClassed(fontPx int, p term.Profile, pal *HTMLPalette) string {
	return c.htmlFragmentWith(fontPx, p, true, pal)
}

func (c *Canvas) htmlFragment(fontPx int, p term.Profile, quantise bool) string {
	return c.htmlFragmentWith(fontPx, p, quantise, nil)
}

// HTMLFragmentCropAs is HTMLFragmentAs for the cells x0..x1-1, y0..y1-1 only,
// so a study can show one part of a real frame at a size the whole would not
// fit. Clamped to the canvas.
func (c *Canvas) HTMLFragmentCropAs(x0, y0, x1, y1, fontPx int, p term.Profile) string {
	x0, y0 = max(x0, 0), max(y0, 0)
	x1, y1 = min(x1, c.W), min(y1, c.H)
	if x1 <= x0 || y1 <= y0 {
		return ""
	}
	return c.htmlFragmentRect(fontPx, p, true, nil, x0, y0, x1, y1)
}

func (c *Canvas) htmlFragmentWith(fontPx int, p term.Profile, quantise bool, pal *HTMLPalette) string {
	return c.htmlFragmentRect(fontPx, p, quantise, pal, 0, 0, c.W, c.H)
}

func (c *Canvas) htmlFragmentRect(fontPx int, p term.Profile, quantise bool, pal *HTMLPalette, x0, y0, x1, y1 int) string {
	var b strings.Builder
	b.WriteString(`<pre style="font-size:`)
	b.WriteString(strconv.Itoa(fontPx))
	b.WriteString(`px">`)
	// Runs of identical colour share one span. Sky rows are a single colour
	// across their whole width, so this is the difference between a readable
	// animation and a 15MB page.
	var run strings.Builder
	var rFG, rBG term.RGB
	runGlyph := false // whether the current run holds anything but spaces
	flush := func() {
		if run.Len() == 0 {
			return
		}
		if pal != nil {
			b.WriteString(`<span class="p`)
			b.WriteString(strconv.Itoa(pal.class(rFG, rBG)))
			b.WriteString(`">`)
		} else {
			b.WriteString(`<span style="color:`)
			writeHex(&b, rFG)
			b.WriteString(`;background:`)
			writeHex(&b, rBG)
			b.WriteString(`">`)
		}
		b.WriteString(run.String())
		b.WriteString(`</span>`)
		run.Reset()
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			r := c.resolve(x, y, p)
			fg, bg := r.fg, r.bg
			if quantise {
				// A shade block's colours are already exact cube entries, and
				// they must not take the glyph boost -- see AppendFGRaw.
				fg = p.Quantise(fg, r.glyph)
				bg = p.Quantise(bg, false)
			}
			// A space shows only its background, so it joins any run with
			// that background, and a run made only of spaces takes the
			// foreground of the first glyph that follows it. Fewer runs, and
			// the same picture: a fifth of a page was space-only spans.
			space := r.ch == ' '
			switch {
			case run.Len() == 0:
				rFG, rBG, runGlyph = fg, bg, !space
			case bg != rBG || (!space && runGlyph && fg != rFG):
				flush()
				rFG, rBG, runGlyph = fg, bg, !space
			case !space && !runGlyph:
				rFG, runGlyph = fg, true
			}
			switch r.ch {
			case '<':
				run.WriteString("&lt;")
			case '>':
				run.WriteString("&gt;")
			case '&':
				run.WriteString("&amp;")
			default:
				run.WriteRune(r.ch)
			}
		}
		flush()
		b.WriteByte('\n')
	}
	b.WriteString(`</pre>`)
	return b.String()
}

// HTMLPage wraps fragments in a document. The charset meta is load-bearing: a
// page without it decodes block and braille glyphs as mojibake.
func HTMLPage(title, body string) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>`)
	b.WriteString(title)
	b.WriteString(`</title><style>`)
	b.WriteString(`body{margin:0;background:#0d0d0f;color:#8a8a94;padding:28px 32px;` +
		`font-family:Menlo,"SF Mono",monospace}`)
	b.WriteString(`pre{margin:0;font-family:Menlo,"Apple Symbols",monospace;line-height:1.0;letter-spacing:0}`)
	b.WriteString(`h1{font-size:15px;font-weight:600;color:#d8d8e0;margin:0 0 22px;letter-spacing:.08em;text-transform:uppercase}`)
	b.WriteString(`.card{display:flex;gap:28px;align-items:center;padding:16px 18px;margin-bottom:14px;` +
		`background:#141419;border:1px solid #222229;border-radius:8px}`)
	b.WriteString(`.meta{width:190px;flex:none}`)
	b.WriteString(`.nm{font-size:17px;color:#f0f0f4;font-weight:600}`)
	b.WriteString(`.rg{font-size:12px;color:#7d7d8a;margin-top:3px}`)
	b.WriteString(`.nt{font-size:12px;color:#5f5f6b;margin-top:8px;line-height:1.45}`)
	b.WriteString(`.big{flex:none;padding:10px 14px;background:#1b1b21;border-radius:6px}`)
	b.WriteString(`.lbl{font-size:10px;color:#55555f;letter-spacing:.1em;margin-bottom:6px;text-transform:uppercase}`)
	b.WriteString(`</style></head><body>`)
	b.WriteString(body)
	b.WriteString(`</body></html>`)
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
