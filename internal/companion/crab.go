package companion

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// Hero, the crab companion. His pick, 2026-09-07, from four silhouettes over
// two rounds; the study is in notes/charstudy/.
//
// A crab is the first companion that is WIDE where the cat is tall. It fills
// the same 12 cell by 7 box -- the box every companion is drawn into -- but
// uses all twelve columns where the cat's ink uses nine, at within a tenth of
// the same weight of ink. It is not bigger; it is the same size arranged
// sideways, and the three extra columns were always reserved and never used,
// because Size() returns the BOX and not the ink.
//
// The body splits at source row 12: the LOWER half (shell bottom and legs)
// never changes and the UPPER half carries the claws, the stalks and the whole
// expression. That is the cat's arrangement too -- one body, a handful of cells
// doing the talking -- and it is what makes the states cheap.
//
// The claw is the channel the cat does not have. Where a cat flicks a tail, a
// crab raises a claw, and a raised claw is POSITION, so it survives a
// screenshot. The encoding rule asks for exactly that.

// CrabCoat is salmon, index 210 of the cube. Locked 2026-09-07. It is an exact
// cube entry, so it survives the glyph path's 2.6x saturate unchanged -- every
// coat the cat ships with is a near-neutral the boost leaves alone, and this is
// the first companion colour with enough chroma that it had to be checked.
var CrabCoat = term.RGB{R: 255, G: 135, B: 135}

var crabLower = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	"..##..############..##..",
	"..##..############..##..",
	".##....##########....##.",
	".##....##########....##.",
	"##......########......##",
	"##......########......##",
	"##.......######.......##",
	".#.......######.......#.",
	".#........####........#.",
	".#....................#.",
	".#....................#.",
}

// Mid-stride: the legs swap phase. Unused until pacing lands, and kept here so
// the walk is authored beside the stand it has to match.
var crabLowerStep = []string{
	"########################",
	"########################",
	".######################.",
	"..####################..",
	"...##################...",
	".##...############...##.",
	".##...############...##.",
	"##.....##########.....##",
	"##.....##########.....##",
	".#......########......#.",
	".#......########......#.",
	"##.......######.......##",
	"##.......######.......##",
	".#........####........#.",
	".#....................#.",
	"..#..................#..",
}

var crabWork = []string{
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

// One source row of drift. The whole idle motion is the claws breathing.
var crabWork2 = []string{
	"........................",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	"######............######",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

var crabRest = []string{
	"........................",
	"........................",
	"........................",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	".####..............####.",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

// One claw up. Authored right, which is screen LEFT once the sprite is
// mirrored -- so the raised claw points at the agent's transcript rather than
// at the frame edge.
var crabAsk = []string{
	"..................##..##",
	"..................##..##",
	"..................######",
	"..................######",
	"##..##............######",
	"##..##.............####.",
	"######..............##..",
	".####...............##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

var crabDone = []string{
	"##..##............##..##",
	"##..##............##..##",
	"##..##............##..##",
	"######............######",
	"######............######",
	".####..............####.",
	"..##................##..",
	"..##................##..",
	"..##....##....##....##..",
	"..##....##....##....##..",
	"..####################..",
	".######################.",
}

// Arms collapsed against the body, stalks short. The largest silhouette becomes
// the smallest, which is the point: distress reads as a shrink.
var crabWorried = []string{
	"........................",
	"........................",
	"........................",
	"........................",
	"........##....##........",
	"........##....##........",
	"..####..##....##..####..",
	"..####..##....##..####..",
	"..######..........######",
	"..######..........######",
	"..####################..",
	".######################.",
}

// The young: six cells by four, the kittens' size and the kittens' place. A
// sub-cell is one pixel wide by two tall, so at this size a one-pixel notch
// still survives the halving -- which is why a crablet keeps an open pincer.
var Crablet = []string{
	"#.#......#.#",
	"###......###",
	".#........#.",
	".#...##...#.",
	".#...##...#.",
	"..########..",
	".##########.",
	"############",
	".##########.",
	"..########..",
	"##..####..##",
	"##..####..##",
	"#....##....#",
	"#....##....#",
	"#..........#",
	"............",
}

// A crab does not swim. It walks in until the water closes over the shell and
// the eyestalks carry on above the surface -- what the animal actually does,
// and the only thing that reads at five cells by two. Same 10x8 box as
// KittenSwim, so a crablet takes exactly a kitten's room in a lane and the lane
// arithmetic needs no new numbers.
var CrabletSwim = []string{
	"..#....#..",
	"..#....#..",
	"..#....#..",
	"..######..",
	".########.",
	".########.",
	"..######..",
	"...####...",
}

// Kind is which animal this companion is.
type Kind int

const (
	KindCat Kind = iota
	KindCrab
)

// crabEyes are the parent's eye cells and the row they sit on, in the authored
// (unmirrored) frame. Written down rather than derived from sprite width: the
// same trap kittens.go records hitting, where a formula was right for two
// scales and wrong for the third.
var (
	crabEyeCells    = [2]int{4, 7}
	crabEyeRow      = 2
	crabletEyeCells = [2]int{2, 3}
	crabletSwimEyes = [2]int{1, 3}
)

// NewCrab builds Hero. It returns the same type the cat does, so every existing
// construction site keeps working: the sprite data is a field, not a hardcoded
// bitmap, and no interface is needed anywhere.
func NewCrab() *Cat {
	return &Cat{
		kind:    KindCrab,
		body:    ParseBitmap(append(append([]string{}, crabWork...), crabLower...)),
		worried: ParseBitmap(append(append([]string{}, crabWorried...), crabLower...)),
		coat:    CrabCoat,
	}
}

// crabUpper picks the claws for a state, and the idle drift within it.
func crabUpper(st State, t float64) []string {
	switch st {
	case Resting:
		if math.Mod(t, 7.2) < 3.6 {
			return crabRest
		}
		return crabWork2
	case NeedsYou:
		return crabAsk
	case Done:
		return crabDone
	case Worried:
		return crabWorried
	}
	// Working: the claws breathe on the cat's own working period.
	if math.Mod(t, 2.2) < 1.1 {
		return crabWork
	}
	return crabWork2
}

// drawCrab is the crab's whole Draw. Compose the upper half for the state onto
// the fixed lower half, breathe, then plot the two eyes on top.
func (c *Cat) drawCrab(l *canvas.Layer, x, y int, t float64, st State) {
	rows := append(append([]string{}, crabUpper(st, t)...), crabLower...)
	src := ParseBitmap(rows)

	// Breathing, exactly as the cat does it: one quadrant subpixel is two
	// source rows, so a two-row shift moves the body by half a character cell
	// -- the smallest vertical step this medium has.
	period := 3.6
	switch st {
	case Working:
		period = 2.2
	case NeedsYou:
		period = 1.6
	case Worried:
		period = 1.9
	}
	lift := 0
	if math.Sin(t*2*math.Pi/period) > 0.35 {
		lift = 2
	}
	f := src.Blank()
	for sy := 0; sy < f.H; sy++ {
		for sx := 0; sx < f.W; sx++ {
			if src.at(sx, sy+lift) {
				f.Set(sx, sy)
			}
		}
	}
	if c.mirror {
		f = f.Mirrored()
	}
	q := f.ToQuadrant()
	plotRim(l, q, x, y)
	(&Sprite{Rows: q, Body: c.coat}).Draw(l, x, y)

	glyph, col := 'o', eyeCol
	switch st {
	case Resting:
		glyph = '-'
	case NeedsYou:
		glyph, col = 'O', eyeAlert
	case Worried:
		col = eyeWorried
	case Done:
		glyph = '^'
	}
	if st != Resting && st != Worried && st != Done && math.Mod(t, 5.3) < 0.16 {
		glyph = '-' // blink
	}
	w, _ := c.Size()
	for _, e := range crabEyeCells {
		ex := x + e
		if c.mirror {
			ex = x + w - 1 - e
		}
		if ground, ok := c.eyeGround(); ok {
			l.PlotOn(ex, y+crabEyeRow, glyph, col, ground, 1)
			continue
		}
		l.Plot(ex, y+crabEyeRow, glyph, col, 1)
	}
}

// Names are what the CLI and the config file speak.
const (
	NameCat  = "cat"
	NameCrab = "crab"
)

// Names lists every companion, for the CLI's help and its validation.
func Names() []string { return []string{NameCat, NameCrab} }

// New builds a companion by name. An unknown name falls back to the default
// rather than failing: a typo in a config file should not stop the scape from
// running, and the CLI validates before it ever writes one.
func New(name string) *Cat {
	switch name {
	case NameCat:
		return NewCat()
	case NameCrab:
		return NewCrab()
	}
	return New(DefaultName)
}

// DefaultName is what a fresh install gets. His ruling 2026-09-07: the crab.
const DefaultName = NameCrab

// Name is which animal this is, for the CLI to report.
func (c *Cat) Name() string {
	if c.kind == KindCrab {
		return NameCrab
	}
	return NameCat
}
