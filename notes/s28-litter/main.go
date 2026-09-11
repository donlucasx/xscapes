// Command s28-litter asks one question of the RENDERED frame: does the litter
// stand in the water?
//
// The claim in RESUME.md is that it does -- "the litter is anchored to the
// COMPANION, not the waterline, so at full activity the water reaches the
// sitters and crablets stand in it" -- and that claim was stated, never
// measured. This measures it.
//
// HOW IT READS THE FRAME. The channel lives in the BACKGROUND: paintBG gives
// every cell of the beach ONE flat tone (writeBandColor, the same tone for the
// dry sand and for the writing band), gives the open sea a ramp between SeaFar
// and SeaNear, and gives the single straddling waterline cell a Lerp between
// wetBandColor and SeaNear. So a glyph-only reader is blind to all of it. Every
// cell here is classified by canvas.ResolveAt's BACKGROUND against the dry
// tone, which makes three classes exact and separable:
//
//	DRY    bg == writeBandColor            -- sand, or the writing band
//	WASH   the one cell above the dry run  -- the wet edge, half sea half sand
//	SEA    anything above that             -- open water
//
// WASH and SEA are DIFFERENT DEFECTS. A wash crest licking the lowest row of a
// sitter is a beach animal at the water's edge, which is what a beach animal
// does. A sitter two or more rows above the dry run is standing in open water,
// which is the thing worth fixing.
//
// TWO TRAPS THIS AVOIDS. One Shore, warmed up, then sampled: a fresh Shore per
// frame has no history and Update clamps a gap over a second to one nominal
// step, so two frames off two new Shores can come out identical. And it is a
// main package under notes/, run with `go run`, because `go test` caches and
// would hand back the previous run's output for a different XSCAPES_TIDE.
//
//	go run ./notes/s28-litter                 # the tide, the shipped default
//	XSCAPES_TIDE=0 go run ./notes/s28-litter  # the old fixed waterline
package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// ---------------------------------------------------------------------------
// COPIED VERBATIM from the product, because they live in package main and
// cannot be imported. A measurement of a different composition is worthless,
// so these are transcriptions, not re-derivations.
//   - layout, compose  -> live.go:222-267
//   - paceSpan         -> pace.go
//   - writeBandColor, bandLuma, sandDay/sandLow/sandNight -> internal/scape/shore.go
//   - swims            -> internal/companion/kittens.go
// ---------------------------------------------------------------------------

type layout struct {
	CatX     int
	PaceSpan int
	SandFrom int
	SandTo   int
	MoonX    float64
	Mirror   bool
}

func compose(w int, catW int, mirror bool) layout {
	const margin = 2
	right := margin + w/32
	if !mirror {
		return layout{
			CatX:     5,
			SandFrom: 5 + catW + 2, SandTo: w - margin,
			MoonX: 0.72, Mirror: false,
		}
	}
	span := paceSpan(w)
	catX := w - catW - right
	if catX < 0 {
		catX = 0
	}
	return layout{
		CatX: catX, PaceSpan: span,
		SandFrom: margin, SandTo: catX - 1 - span,
		MoonX: 0.28, Mirror: true,
	}
}

func paceSpan(w int) int {
	s := w / 16
	if s > 6 {
		s = 6
	}
	if s < 1 {
		s = 1
	}
	return s
}

var (
	sandDay   = term.RGB{R: 215, G: 175, B: 135} // 256 index 180
	sandLow   = term.RGB{R: 175, G: 135, B: 95}  // 256 index 137
	sandNight = term.RGB{R: 135, G: 95, B: 0}    // 256 index 94
)

func bandLuma(c term.RGB) float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}

func writeBandColor(p scape.Palette) term.RGB {
	switch l := bandLuma(p.SandNear); {
	case l >= 160:
		return sandDay
	case l >= 100:
		return sandLow
	default:
		return sandNight
	}
}

func swims(i int, seed int64) bool { return i > 0 && companion.HashF(i, 11, seed) > 0.64 }

// ---------------------------------------------------------------------------
// The rig.
// ---------------------------------------------------------------------------

// cellClass is what the background under one cell of ink actually is.
type cellClass int

const (
	clSky cellClass = iota
	clSea
	clWash
	clDry
)

// rig is one composed scene, built exactly the way frames.frame builds it.
type rig struct {
	w, h     int
	name     string
	cat      *companion.Cat
	sh       *scape.Shore
	c        *canvas.Canvas
	lay      layout
	top      int // frames.go: top := f.c.H - 2 - f.chh
	litterX  int // live.go drawScene: litterX := lay.CatX - lay.PaceSpan
	ccw, chh int
	seed     int64
}

func newRig(name string, w, h int, seed int64) *rig {
	cat := companion.New(name)
	cat.FaceLeft(true) // the mirrored composition is the shipped one
	ccw, chh := cat.Size()
	lay := compose(w, ccw, true)
	sh := scape.NewShore(seed, false)
	sh.MoonX = lay.MoonX
	return &rig{
		w: w, h: h, name: name, cat: cat, sh: sh,
		c:   canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear),
		lay: lay, top: h - 2 - chh, litterX: lay.CatX - lay.PaceSpan,
		ccw: ccw, chh: chh, seed: seed,
	}
}

// sitterBoxes replays the sitter placement arithmetic for a litter of n.
//
// Cat (kittens.go DrawKittens, mirrored):   x -= kw; ky = py + ph - kh; x -= 1
// Crab (crab_kittens.go drawCrabKittens):   x = px - 1 - cw - k*(cw+1); ky = py + 7 - ch
// Both come out at x_k = px - 1 - (k+1)*kw - k, and both put the litter's FEET
// on the parent's last row.
func (r *rig) sitterBoxes(n int) (boxes [][4]int, ky, kh int, nsit int) {
	var sitters []int
	for i := 0; i < n; i++ {
		if !swims(i, r.seed) {
			sitters = append(sitters, i)
		}
	}
	nsit = len(sitters)
	if nsit == 0 {
		return nil, 0, 0, 0
	}
	// The tier the cat has settled into. It starts at 0 and is fed the same
	// count every frame, so one extra application of TierFor reaches the same
	// fixpoint the running Cat is in -- the hysteresis only bites when the
	// count moves.
	ti := companion.TierFor(nsit, companion.TierFor(nsit, 0))
	var rows []string
	if r.name == companion.NameCrab {
		rows = [][]string{companion.Crablet, companion.CrabletSmall, companion.CrabletTiny}[ti]
	} else {
		rows = [][]string{companion.KittenSit, companion.KittenSmall, companion.KittenTiny}[ti]
	}
	bm := companion.ParseBitmap(rows)
	kw, kh := bm.W/2, bm.H/4
	_, ph := r.cat.Size()
	ky = r.top + ph - kh
	x := r.litterX - 1
	for range sitters {
		x -= kw
		if x < 0 {
			break
		}
		boxes = append(boxes, [4]int{x, ky, kw, kh})
		x--
	}
	return boxes, ky, kh, nsit
}

// dryTop is, per column, the first row of the contiguous dry-sand run that
// reaches the bottom of the frame. Read off the RESOLVED background.
func dryTop(c *canvas.Canvas, dry term.RGB) []int {
	out := make([]int, c.W)
	for x := 0; x < c.W; x++ {
		y := c.H - 1
		for y >= 0 {
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			if bg != dry {
				break
			}
			y--
		}
		out[x] = y + 1
	}
	return out
}

func classify(x, y int, dt []int, hy int) cellClass {
	switch {
	case y >= dt[x]:
		return clDry
	case y == dt[x]-1:
		return clWash
	case y <= hy:
		return clSky
	default:
		return clSea
	}
}

// hasInk says whether a draw painted a visible glyph in this cell of a blank
// probe canvas. plotRim's spaces do not count: they clear, they do not draw.
func hasInk(p *canvas.Canvas, x, y int) bool {
	if x < 0 || y < 0 || x >= p.W || y >= p.H {
		return false
	}
	for _, l := range p.Layers {
		if cl := l.Cells[y*l.W+x]; cl.Set && cl.R != ' ' {
			return true
		}
	}
	return false
}

// inkCells lists every painted cell of a probe canvas.
func inkCells(p *canvas.Canvas) [][2]int {
	var out [][2]int
	for y := 0; y < p.H; y++ {
		for x := 0; x < p.W; x++ {
			if hasInk(p, x, y) {
				out = append(out, [2]int{x, y})
			}
		}
	}
	return out
}

type tally struct {
	frames                       int
	framesSea, framesWash        int
	cellsSea, cellsWash, cellsDy int
	maxSea, maxWash              int
	inkTotal                     int
	sitRows                      [2]int // lowest, highest sitter row seen
	meanWL                       float64
	nsit                         int
	kh                           int
	bgDisturbed                  int // columns whose classification the litter changed
	// rowSea/rowWash/rowInk are per ROW OF THE SPRITE, 0 = the top of the
	// sitter box, kh-1 = the feet. This is the distinction that decides
	// whether there is a defect at all: a creature drawn in side elevation
	// has the sea BEHIND its head by construction, and the parent has it
	// too. Water under the FEET is the thing that is wrong.
	rowSea, rowWash, rowInk [8]int
	feetSeaFrames           int
	feetWashFrames          int
	cellsSky                int // ink ABOVE THE HORIZON: a different defect again
}

// run samples one (companion, geometry, level, litter size) point.
func (r *rig) run(level float64, n int, tod float64, warm, frames int) tally {
	act := scape.Activity{Working: true, Level: level, ContextUsed: 0.3, TimeOfDay: tod}
	t := 0.0
	// ONE shore, warmed up. See the package comment.
	for k := 0; k < warm; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
	}
	pal := scape.PaletteAt(tod)
	dry := term.Profile256.Quantise(writeBandColor(pal), false)
	tl := tally{sitRows: [2]int{1 << 30, -1}}
	probe := canvas.New(r.w, r.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	for k := 0; k < frames; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
		hy := int(float64(r.h) * 0.42)
		seaTop := int(float64(r.h)*0.42) + 1
		seaBot := r.sh.SandTop() - 2

		// The classification is taken BEFORE the litter is drawn and checked
		// again after, so "the litter cannot change the background" is a
		// measurement rather than an assumption.
		dt := dryTop(r.c, dry)
		// The real draw, on the real canvas, exactly as drawScene does it.
		r.cat.DrawKittens(r.c.Near(), r.c.Mid(), r.litterX, r.top, n, r.w-1, seaTop, seaBot, t, r.seed)
		after := dryTop(r.c, dry)
		for i := range dt {
			if dt[i] != after[i] {
				tl.bgDisturbed++
			}
		}
		// The same draw on a blank probe, to find WHICH cells it painted.
		probe.Clear()
		r.cat.DrawKittens(probe.Near(), probe.Mid(), r.litterX, r.top, n, r.w-1, seaTop, seaBot, t, r.seed)

		boxes, ky, kh, nsit := r.sitterBoxes(n)
		tl.nsit, tl.kh = nsit, kh
		sum := 0.0
		for _, v := range dt {
			sum += float64(v)
		}
		tl.meanWL += sum / float64(len(dt))

		sea, wash := 0, 0
		feetSea, feetWash := 0, 0
		for _, b := range boxes {
			for y := b[1]; y < b[1]+b[3]; y++ {
				for x := b[0]; x < b[0]+b[2]; x++ {
					if x < 0 || x >= r.w || y < 0 || y >= r.h || !hasInk(probe, x, y) {
						continue
					}
					tl.inkTotal++
					if y < tl.sitRows[0] {
						tl.sitRows[0] = y
					}
					if y > tl.sitRows[1] {
						tl.sitRows[1] = y
					}
					sr := y - b[1] // which row of the sprite
					if sr >= 0 && sr < len(tl.rowInk) {
						tl.rowInk[sr]++
					}
					feet := y == b[1]+b[3]-1
					switch cl := classify(x, y, dt, hy); cl {
					case clSea, clSky:
						sea++
						if cl == clSky {
							tl.cellsSky++
						}
						if sr >= 0 && sr < len(tl.rowSea) {
							tl.rowSea[sr]++
						}
						if feet {
							feetSea++
						}
					case clWash:
						wash++
						if sr >= 0 && sr < len(tl.rowWash) {
							tl.rowWash[sr]++
						}
						if feet {
							feetWash++
						}
					default:
						tl.cellsDy++
					}
				}
			}
		}
		if feetSea > 0 {
			tl.feetSeaFrames++
		}
		if feetWash > 0 {
			tl.feetWashFrames++
		}
		_ = ky
		tl.cellsSea += sea
		tl.cellsWash += wash
		if sea > 0 {
			tl.framesSea++
		}
		if wash > 0 {
			tl.framesWash++
		}
		if sea > tl.maxSea {
			tl.maxSea = sea
		}
		if wash > tl.maxWash {
			tl.maxWash = wash
		}
		tl.frames++
	}
	tl.meanWL /= float64(tl.frames)
	return tl
}

type geom struct {
	w, h  int
	label string
}

// The heights are the ones the product can actually be: host.Band gives the
// scape 9/20 of the window, floor 8, cap 28, so a 51-row window is a 22-row
// scape and a 63-row window is the 28-row cap. 24 is the design target and the
// full-pane (-alt=false) size.
var geoms = []geom{
	{40, 12, "40x12 narrow pane / short band"},
	{80, 24, "80x24 design target"},
	{111, 22, "111x22 his 51-row window, hosted"},
	{153, 28, "153x28 wide window, the 28-row cap"},
}

var levels = []float64{0.0, 0.25, 0.5, 0.75, 1.0}

func main() {
	warm := flag.Int("warm", 250, "warm-up Update calls before sampling")
	frames := flag.Int("frames", 40, "frames sampled per point")
	litter := flag.Int("n", 4, "litter size for the main grid")
	flag.Parse()

	// Exactly what main.go does at startup, so this measures the shipped
	// Terminal.app configuration rather than a test default.
	term.LowerHalf = term.DetectSplit(os.Getenv("TERM_PROGRAM"))
	term.NoSplitCells = term.DetectNoSplit(os.Getenv("TERM_PROGRAM"))

	fmt.Printf("XSCAPES_TIDE -> scape.Tide=%v  TideRange=%.1f  TideEase=%.1f\n",
		scape.Tide, scape.TideRange, scape.TideEase)
	fmt.Printf("TERM_PROGRAM=%q  NoSplitCells=%v  Shading=%v  Ramps=%v\n",
		os.Getenv("TERM_PROGRAM"), term.NoSplitCells, term.Shading, term.Ramps)
	fmt.Printf("sitter sprite heights (cells): kitten %d/%d/%d  crablet %d/%d/%d\n",
		companion.ParseBitmap(companion.KittenSit).H/4,
		companion.ParseBitmap(companion.KittenSmall).H/4,
		companion.ParseBitmap(companion.KittenTiny).H/4,
		companion.ParseBitmap(companion.Crablet).H/4,
		companion.ParseBitmap(companion.CrabletSmall).H/4,
		companion.ParseBitmap(companion.CrabletTiny).H/4)
	fmt.Printf("warm=%d frames=%d litter=%d seeds=1,7,42 tod=13:00\n\n", *warm, *frames, *litter)

	// Which WINDOW sizes produce which scape heights. Asked of host.Band
	// itself rather than worked out from its comment, because the whole point
	// of the height sweep is which of its rows a user can actually get.
	fmt.Println("host.Band: window rows -> scape rows (this is the h in every table below)")
	line := "  "
	for _, wr := range []int{18, 20, 22, 24, 25, 26, 27, 30, 33, 36, 40, 45, 51, 52, 60, 63, 80} {
		_, sc := host.Band(wr)
		line += fmt.Sprintf("%d:%d  ", wr, sc)
	}
	fmt.Println(line)
	fmt.Println()

	seeds := []int64{1, 7, 42}
	const tod = 13.0 / 24

	fmt.Println("== 1. GEOMETRY: where the litter sits, and where the water is ==")
	fmt.Printf("%-6s %-10s %5s %5s %8s %8s %10s %10s\n",
		"animal", "geom", "top", "feet", "sitTop", "kh", "meanWL@0.0", "meanWL@1.0")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range geoms {
			r0 := newRig(nm, g.w, g.h, 7)
			t0 := r0.run(0.0, *litter, tod, *warm, 80)
			r1 := newRig(nm, g.w, g.h, 7)
			t1 := r1.run(1.0, *litter, tod, *warm, 80)
			b, ky, kh, _ := r1.sitterBoxes(*litter)
			_ = b
			fmt.Printf("%-6s %-10s %5d %5d %8d %8d %10.2f %10.2f\n",
				nm, fmt.Sprintf("%dx%d", g.w, g.h), r1.top, r1.top+6, ky, kh,
				t0.meanWL, t1.meanWL)
		}
	}

	fmt.Printf("\n== 2. WET SITTERS, litter of %d, %d frames x %d seeds per cell ==\n", *litter, *frames, len(seeds))
	fmt.Println("   frSEA = % of frames with >=1 sitter cell over OPEN WATER")
	fmt.Println("   frWASH= % of frames with >=1 sitter cell on the WASH cell only")
	fmt.Println("   cSEA  = mean sitter cells over open water per frame (max in brackets)")
	fmt.Printf("%-6s %-8s %6s %7s %7s %9s %9s %8s %7s\n",
		"animal", "geom", "level", "frSEA%", "frWASH%", "cSEA", "cWASH", "ink/fr", "sitters")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range geoms {
			for _, lv := range levels {
				var agg tally
				for _, sd := range seeds {
					r := newRig(nm, g.w, g.h, sd)
					t := r.run(lv, *litter, tod, *warm, *frames)
					agg.frames += t.frames
					agg.framesSea += t.framesSea
					agg.framesWash += t.framesWash
					agg.cellsSea += t.cellsSea
					agg.cellsWash += t.cellsWash
					agg.inkTotal += t.inkTotal
					agg.nsit = t.nsit
					if t.maxSea > agg.maxSea {
						agg.maxSea = t.maxSea
					}
					if t.maxWash > agg.maxWash {
						agg.maxWash = t.maxWash
					}
				}
				fmt.Printf("%-6s %-8s %6.2f %7.1f %7.1f %5.2f[%2d] %5.2f[%2d] %8.1f %7d\n",
					nm, fmt.Sprintf("%dx%d", g.w, g.h), lv,
					100*float64(agg.framesSea)/float64(agg.frames),
					100*float64(agg.framesWash)/float64(agg.frames),
					float64(agg.cellsSea)/float64(agg.frames), agg.maxSea,
					float64(agg.cellsWash)/float64(agg.frames), agg.maxWash,
					float64(agg.inkTotal)/float64(agg.frames), agg.nsit)
			}
		}
	}

	{
		// The guard behind the whole method: the litter plots glyphs and never
		// PlotOn, so it cannot change a background. Measured, not assumed --
		// run() reclassifies every column before and after the draw.
		g := newRig(companion.NameCrab, 111, 22, 7).run(1.0, *litter, tod, *warm, 80)
		fmt.Printf("\nGUARD: columns whose class the litter draw changed, 80 frames x 111 cols: %d\n",
			g.bgDisturbed)
	}

	fmt.Println("\n== 1b. ONE REAL FRAME, so the classifier can be checked by eye ==")
	fmt.Println("   '~' open sea   '=' the wash cell   '.' dry sand/writing band   '^' sky")
	fmt.Println("   ink carries its own class: sitter S=sea s=wash _=dry; parent P p -;")
	fmt.Println("   other litter W w D. UPPERCASE = over OPEN SEA. 'D' = swimmer on dry sand.")
	dumpFrame(companion.NameCrab, 111, 22, 1.0, tod, *warm, 7)
	dumpFrame(companion.NameCrab, 111, 10, 1.0, tod, *warm, 7)

	fmt.Println("\n== 2a. WHICH ROW OF THE SITTER IS WET, and are the FEET in it ==")
	fmt.Println("   row 0 is the top of the sitter box, the last row is its FEET.")
	fmt.Println("   %SEA is the share of that row's ink cells over open water.")
	fmt.Println("   A creature in side elevation has sea BEHIND its head by construction,")
	fmt.Println("   so only the feet rows say anything. feetSEA/feetWASH are % OF FRAMES.")
	fmt.Printf("%-6s %-8s %6s  %-40s %8s %9s\n", "animal", "geom", "level", "%SEA by row (top->feet) | %WASH by row", "feetSEA%", "feetWASH%")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range geoms {
			for _, lv := range []float64{0.5, 1.0} {
				var agg tally
				for _, sd := range seeds {
					r := newRig(nm, g.w, g.h, sd)
					t := r.run(lv, *litter, tod, *warm, *frames)
					agg.frames += t.frames
					agg.feetSeaFrames += t.feetSeaFrames
					agg.feetWashFrames += t.feetWashFrames
					agg.kh = t.kh
					for i := range agg.rowSea {
						agg.rowSea[i] += t.rowSea[i]
						agg.rowWash[i] += t.rowWash[i]
						agg.rowInk[i] += t.rowInk[i]
					}
				}
				prof, wprof := "", ""
				for i := 0; i < agg.kh; i++ {
					if agg.rowInk[i] == 0 {
						prof, wprof = prof+"   -", wprof+"   -"
						continue
					}
					prof += fmt.Sprintf(" %3.0f", 100*float64(agg.rowSea[i])/float64(agg.rowInk[i]))
					wprof += fmt.Sprintf(" %3.0f", 100*float64(agg.rowWash[i])/float64(agg.rowInk[i]))
				}
				prof += "  |" + wprof
				fmt.Printf("%-6s %-8s %6.2f  %-40s %8.1f %9.1f\n", nm,
					fmt.Sprintf("%dx%d", g.w, g.h), lv, prof,
					100*float64(agg.feetSeaFrames)/float64(agg.frames),
					100*float64(agg.feetWashFrames)/float64(agg.frames))
			}
		}
	}

	fmt.Println("\n== 2b. THE HEIGHT SWEEP at 111 columns, level 1.0 and 0.0 ==")
	fmt.Println("   Every scape height host.Band can produce: floor 8, cap 28.")
	fmt.Println("   feetSEA@1 / feetWASH@1 are the only columns that mean 'in the water'.")
	fmt.Printf("%-6s %4s %5s %6s %6s %6s %8s %8s %9s %10s %8s\n",
		"animal", "h", "wr", "top", "feet", "beach", "frSEA@1", "cSEA@1", "feetSEA@1", "feetWASH@1", "skyINK%")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for h := 8; h <= 28; h++ {
			var a1 tally
			for _, sd := range seeds {
				r1 := newRig(nm, 111, h, sd)
				t1 := r1.run(1.0, *litter, tod, *warm, *frames)
				a1.frames += t1.frames
				a1.framesSea += t1.framesSea
				a1.cellsSea += t1.cellsSea
				a1.feetSeaFrames += t1.feetSeaFrames
				a1.feetWashFrames += t1.feetWashFrames
				a1.cellsSky += t1.cellsSky
				a1.inkTotal += t1.inkTotal
			}
			r := newRig(nm, 111, h, 7)
			fmt.Printf("%-6s %4d %5d %6d %6d %6d %8.1f %8.2f %9.1f %10.1f %8.1f\n", nm, h,
				writeRows(h), r.top, r.top+6, beachRows(h),
				100*float64(a1.framesSea)/float64(a1.frames),
				float64(a1.cellsSea)/float64(a1.frames),
				100*float64(a1.feetSeaFrames)/float64(a1.frames),
				100*float64(a1.feetWashFrames)/float64(a1.frames),
				100*float64(a1.cellsSky)/float64(max(1, a1.inkTotal)))
		}
	}

	fmt.Println("\n== 3. THE LADDER: does a bigger litter get wetter or drier? (111x22, level 1.0) ==")
	fmt.Printf("%-6s %5s %8s %6s %7s %7s %9s\n", "animal", "n", "sitters", "kh", "sitTop", "frSEA%", "cSEA")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, n := range []int{1, 2, 4, 6, 8, 12, 20, 38} {
			var agg tally
			var ky int
			for _, sd := range seeds {
				r := newRig(nm, 111, 22, sd)
				t := r.run(1.0, n, tod, *warm, *frames)
				_, ky, _, _ = r.sitterBoxes(n)
				agg.frames += t.frames
				agg.framesSea += t.framesSea
				agg.cellsSea += t.cellsSea
				agg.nsit += t.nsit
				agg.kh = t.kh
			}
			fmt.Printf("%-6s %5d %8.1f %6d %7d %7.1f %9.2f\n", nm, n,
				float64(agg.nsit)/float64(len(seeds)), agg.kh, ky,
				100*float64(agg.framesSea)/float64(agg.frames),
				float64(agg.cellsSea)/float64(agg.frames))
		}
	}

	fmt.Println("\n== 4. THE PARENT, THE SWIMMERS AND THE EXIT QUEUE (level 1.0, litter 4) ==")
	fmt.Println("   parentSEA = companion's own ink cells over open water, per frame")
	fmt.Println("   pFeetSEA% = % of frames with the PARENT'S OWN FEET ROW over open water")
	fmt.Println("   swimDRY   = swimmer ink cells sitting on DRY SAND (a kitten on the beach)")
	fmt.Println("   exitDRY   = exit-queue ink cells on DRY SAND")
	fmt.Printf("%-6s %-8s %10s %10s %10s %10s %10s\n", "animal", "geom", "parentSEA", "pFeetSEA%", "swimDRY", "swimINK", "exitDRY")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range geoms {
			var pSea, pFeet, sDry, sInk, eDry, eInk, fr int
			pp := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			lp := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			ep := canvas.New(g.w, g.h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
			for _, sd := range seeds {
				r := newRig(nm, g.w, g.h, sd)
				act := scape.Activity{Working: true, Level: 1.0, ContextUsed: 0.3, TimeOfDay: tod}
				t := 0.0
				for k := 0; k < *warm; k++ {
					t += 0.08
					r.sh.Update(r.c, t, act)
				}
				dry := term.Profile256.Quantise(writeBandColor(scape.PaletteAt(tod)), false)
				for k := 0; k < *frames; k++ {
					t += 0.08
					r.sh.Update(r.c, t, act)
					hy := int(float64(g.h) * 0.42)
					seaTop := hy + 1
					seaBot := r.sh.SandTop() - 2
					dt := dryTop(r.c, dry)
					fr++

					// The parent, drawn where drawScene puts it (pace dx = 0
					// at rest; Steps is 0 in this synthetic state).
					pp.Clear()
					r.cat.Draw(pp.Near(), r.lay.CatX, r.top, t, companion.Working)
					pFeetWet := false
					for _, cell := range inkCells(pp) {
						if c := classify(cell[0], cell[1], dt, hy); c == clSea {
							pSea++
							if cell[1] == r.top+6 { // the companion's own feet row
								pFeetWet = true
							}
						}
					}
					if pFeetWet {
						pFeet++
					}

					// The whole litter, then subtract the sitter boxes: what
					// is left is the swimmers.
					lp.Clear()
					r.cat.DrawKittens(lp.Near(), lp.Mid(), r.litterX, r.top, 4, g.w-1, seaTop, seaBot, t, r.seed)
					boxes, _, _, _ := r.sitterBoxes(4)
					inBox := func(x, y int) bool {
						for _, b := range boxes {
							if x >= b[0] && x < b[0]+b[2] && y >= b[1] && y < b[1]+b[3] {
								return true
							}
						}
						return false
					}
					for _, cell := range inkCells(lp) {
						if inBox(cell[0], cell[1]) {
							continue
						}
						sInk++
						if classify(cell[0], cell[1], dt, hy) == clDry {
							sDry++
						}
					}

					// The exit queue, four subagents mid-swim-off.
					ep.Clear()
					r.cat.DrawKittenExits(ep.Near(), []float64{0.1, 0.35, 0.6, 0.85},
						r.litterX, r.top, g.w-1, seaTop, seaBot, t, r.seed)
					for _, cell := range inkCells(ep) {
						eInk++
						if classify(cell[0], cell[1], dt, hy) == clDry {
							eDry++
						}
					}
				}
			}
			fmt.Printf("%-6s %-8s %10.2f %10.1f %10.2f %10.1f %10.2f\n", nm,
				fmt.Sprintf("%dx%d", g.w, g.h),
				float64(pSea)/float64(fr), 100*float64(pFeet)/float64(fr),
				float64(sDry)/float64(fr),
				float64(sInk)/float64(fr), float64(eDry)/float64(fr))
			_ = eInk
		}
	}

	fmt.Println("\n== 5. HIS REAL SESSIONS: how often is there a litter at all, and how big ==")
	realLog()
}

// realLog folds the live event spools through the real reducer, the way
// tune.go's fold does, and counts. Do not reason about his sessions: count them.
func realLog() {
	dir, err := event.RunDir()
	if err != nil {
		fmt.Println("  no run dir:", err)
		return
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	fmt.Printf("  %d spool files in %s\n", len(files), dir)
	var kits []int
	var lvlWithLitter []float64
	sessions, evs, litterSecs, totalSecs := 0, 0, 0, 0
	maxKit := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var list []event.Event
		for _, line := range bytes.Split(b, []byte("\n")) {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			e, err := event.Decode(line)
			if err != nil || e.TS == 0 {
				continue
			}
			list = append(list, e)
		}
		if len(list) < 20 {
			continue
		}
		sort.Slice(list, func(i, j int) bool { return list[i].TS < list[j].TS })
		sessions++
		evs += len(list)
		r := reduce.New("s28")
		at := func(ms int64) time.Time { return time.UnixMilli(ms) }
		start, end := at(list[0].TS), at(list[len(list)-1].TS)
		i := 0
		const maxGap = 3 * time.Minute
		for t := start; !t.After(end); t = t.Add(time.Second) {
			for i < len(list) && !at(list[i].TS).After(t) {
				r.Apply(list[i], at(list[i].TS))
				i++
			}
			st := r.State(t)
			totalSecs++
			if st.Kittens > 0 {
				litterSecs++
				kits = append(kits, st.Kittens)
				lvlWithLitter = append(lvlWithLitter, st.Act.Level)
				if st.Kittens > maxKit {
					maxKit = st.Kittens
				}
			}
			if i < len(list) {
				if gap := at(list[i].TS).Sub(t); gap > maxGap {
					t = at(list[i].TS).Add(-time.Second)
				}
			}
		}
	}
	fmt.Printf("  %d sessions, %d events, %d sampled seconds, %d of them with a litter (%.1f%%)\n",
		sessions, evs, totalSecs, litterSecs, 100*float64(litterSecs)/float64(max(1, totalSecs)))
	if len(kits) == 0 {
		fmt.Println("  NO LITTER SECONDS AT ALL -- the sample is empty, believe nothing below")
		return
	}
	sort.Ints(kits)
	fmt.Printf("  litter size while it exists: p50 %d  p90 %d  p99 %d  max %d\n",
		kits[len(kits)*50/100], kits[len(kits)*90/100], kits[len(kits)*99/100], maxKit)
	sort.Float64s(lvlWithLitter)
	q := func(p int) float64 { return lvlWithLitter[len(lvlWithLitter)*p/100] }
	fmt.Printf("  activity level while a litter exists: p10 %.2f  p50 %.2f  p90 %.2f  p99 %.2f\n",
		q(10), q(50), q(90), q(99))
	over := 0
	for _, v := range lvlWithLitter {
		if v >= 0.75 {
			over++
		}
	}
	fmt.Printf("  litter-seconds at level >= 0.75: %.1f%%   >= 0.50: %.1f%%\n",
		100*float64(over)/float64(len(lvlWithLitter)), 100*float64(countGE(lvlWithLitter, 0.5))/float64(len(lvlWithLitter)))
	_ = math.Abs
}

func countGE(v []float64, x float64) int {
	n := 0
	for _, f := range v {
		if f >= x {
			n++
		}
	}
	return n
}

// dumpFrame prints one composed frame as a classification map. An instrument
// nobody can check is an instrument nobody should believe, and the two mistakes
// this project has paid for -- a clean reading off an empty window, and a proxy
// reported as the thing itself -- both survive a table and die on a picture.
func dumpFrame(nm string, w, h int, level, tod float64, warm int, seed int64) {
	r := newRig(nm, w, h, seed)
	act := scape.Activity{Working: true, Level: level, ContextUsed: 0.3, TimeOfDay: tod}
	t := 0.0
	for k := 0; k < warm; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
	}
	// Walk on to the crest of the wash: the worst frame, not a random one.
	best, bestT := -1, t
	for k := 0; k < 400; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
		st := r.sh.SandTop()
		if st > best {
			best, bestT = st, t
		}
	}
	t = bestT
	r.sh.Update(r.c, t, act)
	dry := term.Profile256.Quantise(writeBandColor(scape.PaletteAt(tod)), false)
	hy := int(float64(h) * 0.42)
	seaTop, seaBot := hy+1, r.sh.SandTop()-2
	dt := dryTop(r.c, dry)

	lp := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	r.cat.DrawKittens(lp.Near(), lp.Mid(), r.litterX, r.top, 4, w-1, seaTop, seaBot, t, r.seed)
	pp := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	r.cat.Draw(pp.Near(), r.lay.CatX, r.top, t, companion.Working)
	boxes, _, _, _ := r.sitterBoxes(4)
	inBox := func(x, y int) bool {
		for _, b := range boxes {
			if x >= b[0] && x < b[0]+b[2] && y >= b[1] && y < b[1]+b[3] {
				return true
			}
		}
		return false
	}
	x0 := 0
	if boxes != nil && boxes[len(boxes)-1][0]-4 > 0 {
		x0 = boxes[len(boxes)-1][0] - 4
	}
	fmt.Printf("   %s %dx%d level %.2f seed %d, worst wash frame; top=%d feet=%d cols %d..%d\n",
		nm, w, h, level, seed, r.top, r.top+6, x0, w-1)
	for y := hy; y < h; y++ {
		line := fmt.Sprintf("   r%-3d ", y)
		for x := x0; x < w; x++ {
			g := '.'
			switch classify(x, y, dt, hy) {
			case clSea:
				g = '~'
			case clWash:
				g = '='
			case clSky:
				g = '^'
			}
			// The ink must not HIDE the class underneath it -- that is exactly
			// how a wet cell would go unseen. Case carries the class:
			// UPPER = over open sea, lower = over the wash, '_'/'-' = dry.
			cls := classify(x, y, dt, hy)
			pick := func(sea, wash, dry rune) rune {
				switch cls {
				case clSea, clSky:
					return sea
				case clWash:
					return wash
				}
				return dry
			}
			if hasInk(lp, x, y) {
				if inBox(x, y) {
					g = pick('S', 's', '_')
				} else {
					g = pick('W', 'w', 'D')
				}
			} else if hasInk(pp, x, y) {
				g = pick('P', 'p', '-')
			}
			line += string(g)
		}
		fmt.Println(line)
	}
}

// writeRows is Shore.Update's writing-band arithmetic, transcribed. It matters
// because the band is what CLAMPS the swell: `if writeTop < c.H` gates the
// rescale, so a scape with no band (wr == 0, h < 12) has nothing stopping the
// water reaching the bottom row.
func writeRows(h int) int {
	wr := 4
	if m := h / 6; wr > m {
		wr = m
	}
	if wr < 2 {
		wr = 0
	}
	return wr
}

// beachRows is Shore.Update's own beach arithmetic, transcribed, so the sweep
// can print WHY a height is wet without opening the source again.
func beachRows(h int) int {
	wr := writeRows(h)
	beach := h / 5
	if min := wr + 3; wr > 0 && beach < min {
		beach = min
	}
	if beach < 4 {
		beach = 4
	}
	return beach
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
