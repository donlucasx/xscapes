// Command s28-litter-verify is an ADVERSARIAL re-check of notes/s28-litter's
// finding "the litter does not stand in the water at any height his windows
// produce". It is written to REFUTE that, not to agree with it.
//
// WHERE IT DELIBERATELY DIFFERS FROM THE INSTRUMENT IT IS CHECKING:
//
//  1. BOX-FREE FEET. s28-litter replays the sitter-placement arithmetic and
//     calls the bottom row of its own computed box "the feet". If that replay
//     were off by a row the feet would be measured at the wrong row and the
//     clean result would be an artefact. This uses an invariant instead:
//     the cat puts sitters at ky = py+ph-kh and the crab at ky = py+7-ch, so
//     for BOTH animals and EVERY tier the sitters' bottom row is exactly
//     py+6 -- the parent's own feet row. And neither swimmer path can reach
//     it (cat swimmers stop at py+ph-4, crab swimmers at seaBot-1 which is
//     above py+6 for every beach >= 4). So ANY litter ink on row py+6 is a
//     sitter's foot, with no box arithmetic anywhere.
//
//  2. SKY IS NOT SEA. s28-litter's tally does `case clSea, clSky: sea++`, so
//     its frSEA/cSEA columns count ink above the horizon as ink over water.
//     At h=8 that is 47.8% of the ink it counted. Here they are separate.
//
//  3. THE CAUSE IS ISOLATED, NOT INFERRED. s28-litter claims the cause is the
//     `if writeTop < c.H` gate in Shore.Update, "isolated by varying scape
//     HEIGHT" -- but height moves wr, beach, sy, scale, hy and the sprite's
//     row all at once. Shore.WriteRows and Shore.SandRows are exported, so
//     the two can be crossed at ONE height: gate on/off x beach 4/6.
//
//  4. THINGS IT NEVER VARIED: the WIDTH (the tide's coast has no time term in
//     x, so the width alone decides whether the litter sits in a hollow), the
//     TIME OF DAY (which is what the classifier's dry tone is taken from),
//     the litter SIZE at the wet heights, and the seeds.
//
// Traps avoided: ONE Shore per point, warmed up (a fresh Shore per frame has
// no history and Update clamps a gap over a second to one nominal step); a
// main package under notes/ run with `go run`, because `go test` caches.
//
//	go run ./notes/s28-litter-verify
//	XSCAPES_TIDE=0 go run ./notes/s28-litter-verify
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// --- transcribed from the product, because they live in package main -------
// compose/paceSpan -> live.go; sand tones + writeBandColor -> shore.go.

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

// catX and litterX are live.go's compose() + drawScene(), mirrored layout.
func catX(w, catW int) int {
	right := 2 + w/32
	x := w - catW - right
	if x < 0 {
		x = 0
	}
	return x
}

var (
	sandDay   = term.RGB{R: 215, G: 175, B: 135}
	sandLow   = term.RGB{R: 175, G: 135, B: 95}
	sandNight = term.RGB{R: 135, G: 95, B: 0}
)

func writeBandColor(p scape.Palette) term.RGB {
	l := 0.299*float64(p.SandNear.R) + 0.587*float64(p.SandNear.G) + 0.114*float64(p.SandNear.B)
	switch {
	case l >= 160:
		return sandDay
	case l >= 100:
		return sandLow
	default:
		return sandNight
	}
}

// --- the rig ---------------------------------------------------------------

type rig struct {
	w, h    int
	name    string
	cat     *companion.Cat
	sh      *scape.Shore
	c       *canvas.Canvas
	probe   *canvas.Canvas
	top     int
	litterX int
	feetRow int
	dry     term.RGB
	seed    int64
}

func newRig(name string, w, h int, seed int64, tod float64) *rig {
	cat := companion.New(name)
	cat.FaceLeft(true)
	ccw, chh := cat.Size()
	if chh != 7 {
		panic(fmt.Sprintf("%s sprite height is %d, not 7 -- the feet-row invariant does not hold", name, chh))
	}
	cx := catX(w, ccw)
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	top := h - 2 - chh
	return &rig{
		w: w, h: h, name: name, cat: cat, sh: sh,
		c:       canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear),
		probe:   canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear),
		top:     top,
		litterX: cx - paceSpan(w),
		feetRow: top + 6,
		dry:     term.Profile256.Quantise(writeBandColor(scape.PaletteAt(tod)), false),
		seed:    seed,
	}
}

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

// dryTop: per column, the top of the contiguous dry run that reaches the
// bottom of the frame. Read off the RESOLVED BACKGROUND, which is where the
// channel lives -- a glyph-only reader is blind to all of it.
func dryTop(c *canvas.Canvas, dry term.RGB) []int {
	out := make([]int, c.W)
	for x := 0; x < c.W; x++ {
		y := c.H - 1
		for y >= 0 {
			if _, _, bg := c.ResolveAt(x, y, term.Profile256); bg != dry {
				break
			}
			y--
		}
		out[x] = y + 1
	}
	return out
}

// leaks counts cells ABOVE the wash row whose background is nevertheless the
// dry tone. A non-zero count means the dry run was cut short by something --
// a glyph with its own ground, a colour collision -- and every class above it is
// then suspect. This is the integrity check s28-litter does not have: its GUARD
// only shows the LITTER does not move the background. Counted BELOW THE HORIZON
// only: the midday sun is a warm tan disc that quantises onto the sand index,
// which is real, is in the sky, and has nothing to do with the waterline.
func leaks(c *canvas.Canvas, dt []int, dry term.RGB, hy int) int {
	n := 0
	for x := 0; x < c.W; x++ {
		for y := hy + 1; y < dt[x]-1; y++ {
			if _, _, bg := c.ResolveAt(x, y, term.Profile256); bg == dry {
				n++
			}
		}
	}
	return n
}

type res struct {
	frames                int
	feetInk               int // total sitter-foot cells seen: the non-empty proof
	feetSeaFr, feetWashFr int
	feetSeaCells          int
	// wetMax is the MAXIMUM over frames and foot columns of (washRow - feetRow).
	// >= 1  open sea stands on the feet, that many rows deep
	//    0  the wash cell IS the feet row
	//  < 0  dry, with |v| rows of daylight between the wash and the feet
	wetMax  int
	sumWet  float64
	haveWet bool
	// the observable s28-litter redefined: is water DRAWN OVER the sitter?
	anySeaFr, anyWashFr int
	seaCells, skyCells  int
	sitInk              int
	rowsWet             int // max sprite rows with sea over them, in any frame
	leak                int
}

// sample runs one point.
//
// SITTERS ONLY, and the way to get it is not obvious. Handing DrawKittens
// seaTop=0, seaBot=0 kills the CRAB's swimmers (crabSwimSpans returns nil on
// seaBot <= top) and NOT the cat's: drawSwimmers only honours seaBot `if
// seaBot > 0`, so at zero it falls back to bot = py+ph-4 and swims anyway.
// That cost this instrument one reading -- the cat's sky and sea counts were
// its swimmers. seaTop above the frame and seaBot=1 empties both.
// Sitter placement reads neither value in either animal.
func (r *rig) sample(level float64, n int, tod float64, warm, frames int, checkLeak bool) res {
	act := scape.Activity{Working: true, Level: level, ContextUsed: 0.3, TimeOfDay: tod}
	t := 0.0
	for k := 0; k < warm; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
	}
	out := res{wetMax: -1 << 30}
	hy := int(float64(r.h) * 0.42)
	for k := 0; k < frames; k++ {
		t += 0.08
		r.sh.Update(r.c, t, act)
		dt := dryTop(r.c, r.dry)
		if checkLeak {
			out.leak += leaks(r.c, dt, r.dry, hy)
		}
		r.probe.Clear()
		// sitters only -- see the note above.
		r.cat.DrawKittens(r.probe.Near(), r.probe.Mid(), r.litterX, r.top, n, r.w-1, r.h+10, 1, t, r.seed)

		fSea, fWash := 0, 0
		wet := -1 << 30
		aSea, aWash := 0, 0
		rowWet := map[int]bool{}
		for x := 0; x < r.w; x++ {
			for y := 0; y < r.h; y++ {
				if !hasInk(r.probe, x, y) {
					continue
				}
				out.sitInk++
				sky := y <= hy
				sea := !sky && y < dt[x]-1
				wash := !sky && y == dt[x]-1
				if sky {
					out.skyCells++
				}
				if sea {
					out.seaCells++
					aSea++
					rowWet[y] = true
				}
				if wash {
					aWash++
				}
				if y == r.feetRow {
					out.feetInk++
					if sea {
						fSea++
						out.feetSeaCells++
					}
					if wash {
						fWash++
					}
					if c := (dt[x] - 1) - r.feetRow; c > wet {
						wet = c
					}
				}
			}
		}
		if fSea > 0 {
			out.feetSeaFr++
		}
		if fWash > 0 {
			out.feetWashFr++
		}
		if aSea > 0 {
			out.anySeaFr++
		}
		if aWash > 0 {
			out.anyWashFr++
		}
		if len(rowWet) > out.rowsWet {
			out.rowsWet = len(rowWet)
		}
		if wet != -1<<30 {
			if wet > out.wetMax {
				out.wetMax = wet
			}
			out.sumWet += float64(wet)
			out.haveWet = true
		}
		out.frames++
	}
	return out
}

func agg(a, b res) res {
	if a.frames == 0 {
		a.wetMax = -1 << 30
	}
	a.frames += b.frames
	a.feetInk += b.feetInk
	a.feetSeaFr += b.feetSeaFr
	a.feetWashFr += b.feetWashFr
	a.feetSeaCells += b.feetSeaCells
	a.anySeaFr += b.anySeaFr
	a.anyWashFr += b.anyWashFr
	a.seaCells += b.seaCells
	a.skyCells += b.skyCells
	a.sitInk += b.sitInk
	a.leak += b.leak
	a.sumWet += b.sumWet
	if b.rowsWet > a.rowsWet {
		a.rowsWet = b.rowsWet
	}
	if b.haveWet && (!a.haveWet || b.wetMax > a.wetMax) {
		a.wetMax, a.haveWet = b.wetMax, true
	}
	return a
}

func pct(a, b int) float64 {
	if b == 0 {
		return -1
	}
	return 100 * float64(a) / float64(b)
}

func main() {
	frames := flag.Int("frames", 200, "frames per point")
	warm := flag.Int("warm", 250, "warm-up Updates")
	flag.Parse()
	term.LowerHalf = term.DetectSplit(os.Getenv("TERM_PROGRAM"))
	term.NoSplitCells = term.DetectNoSplit(os.Getenv("TERM_PROGRAM"))
	fmt.Printf("Tide=%v TideRange=%.1f TideEase=%.1f  TERM_PROGRAM=%q NoSplitCells=%v Shading=%v Ramps=%v\n",
		scape.Tide, scape.TideRange, scape.TideEase, os.Getenv("TERM_PROGRAM"),
		term.NoSplitCells, term.Shading, term.Ramps)
	fmt.Printf("frames=%d warm=%d  seeds 2,3,5 (DIFFERENT from s28-litter's 1,7,42)\n\n", *frames, *warm)
	const tod = 13.0 / 24
	seeds := []int64{2, 3, 5}

	// ---------------------------------------------------------------- 0
	fmt.Println("== 0. IS THE CLASSIFIER SOUND, AND IS THE SAMPLE NON-EMPTY ==")
	fmt.Println("   leak = cells above the wash whose background is the DRY tone (should be 0)")
	fmt.Println("   feetInk = sitter foot cells actually found (a 0 here voids every % below)")
	fmt.Printf("%-6s %-9s %6s %10s %10s %10s\n", "animal", "geom", "level", "leak", "feetInk", "sitInk")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range [][2]int{{111, 22}, {80, 24}, {40, 12}, {111, 10}} {
			r := newRig(nm, g[0], g[1], 7, tod)
			o := r.sample(1.0, 4, tod, *warm, 60, true)
			fmt.Printf("%-6s %-9s %6.2f %10d %10d %10d\n", nm,
				fmt.Sprintf("%dx%d", g[0], g[1]), 1.0, o.leak, o.feetInk, o.sitInk)
		}
	}

	// ---------------------------------------------------------------- 1
	fmt.Println("\n== 1. BOX-FREE FEET, height sweep at 111 cols, level 1.0, n=4, NEW seeds ==")
	fmt.Println("   feet row = top+6, the invariant, found from the RENDERED ink, not from a box.")
	fmt.Println("   wetMax = max over frames/columns of (washRow - feetRow): >=1 open sea stands")
	fmt.Println("   on the feet that deep, 0 = the wash IS the feet row, <0 = dry with |v| rows spare.")
	fmt.Printf("%-6s %4s %8s %10s %11s %10s %9s\n", "animal", "h", "feetInk", "feetSEA%", "feetWASH%", "wetMax", "meanWet")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for h := 8; h <= 28; h++ {
			var a res
			for _, sd := range seeds {
				a = agg(a, newRig(nm, 111, h, sd, tod).sample(1.0, 4, tod, *warm, *frames, false))
			}
			fmt.Printf("%-6s %4d %8d %10.1f %11.1f %10d %9.2f\n", nm, h, a.feetInk,
				pct(a.feetSeaFr, a.frames), pct(a.feetWashFr, a.frames),
				a.wetMax, a.sumWet/float64(a.frames))
		}
	}

	// ---------------------------------------------------------------- 2
	fmt.Println("\n== 2. THE VARIABLE THEY NEVER SWEPT: WIDTH ==")
	fmt.Println("   Under the tide the coast has NO time term in x, so the width alone fixes")
	fmt.Println("   whether the litter's own columns sit in a hollow or on a rise.")
	fmt.Printf("%-6s %4s %5s %8s %10s %11s %9s\n", "animal", "h", "w", "feetInk", "feetSEA%", "feetWASH%", "wetMax")
	worst := map[int][3]int{} // h -> {worstW, feetWashPctx10, wetMax}
	for _, h := range []int{12, 14, 17, 22, 24} {
		for w := 40; w <= 200; w += 5 {
			var a res
			for _, sd := range seeds[:2] {
				a = agg(a, newRig(companion.NameCrab, w, h, sd, tod).sample(1.0, 4, tod, *warm, 150, false))
			}
			pw := pct(a.feetWashFr, a.frames)
			if cur, ok := worst[h]; !ok || int(pw*10) > cur[1] {
				worst[h] = [3]int{w, int(pw * 10), a.wetMax}
			}
			if pw > 0 || pct(a.feetSeaFr, a.frames) > 0 {
				fmt.Printf("%-6s %4d %5d %8d %10.1f %11.1f %9d\n", "crab", h, w, a.feetInk,
					pct(a.feetSeaFr, a.frames), pw, a.wetMax)
			}
		}
	}
	fmt.Println("   worst width found per height (h: w, feetWASH%, wetMax):")
	for _, h := range []int{12, 14, 17, 22, 24} {
		v := worst[h]
		fmt.Printf("     h=%d -> w=%d feetWASH=%.1f%% wetMax=%d\n", h, v[0], float64(v[1])/10, v[2])
	}

	// ---------------------------------------------------------------- 3
	fmt.Println("\n== 3. THE CAUSE, ISOLATED. gate x beach, both at h=22 w=111 level 1.0 ==")
	fmt.Println("   s28-litter blames `if writeTop < c.H` (the swell-rescale gate) and says it")
	fmt.Println("   isolated that by varying HEIGHT -- which moves wr, beach, sy, scale and the")
	fwr := func(wr, sand int) res {
		var a res
		for _, sd := range seeds {
			r := newRig(companion.NameCrab, 111, 22, sd, tod)
			r.sh.WriteRows = wr
			r.sh.SandRows = sand
			a = agg(a, r.sample(1.0, 4, tod, *warm, *frames, false))
		}
		return a
	}
	fmt.Println("   sprite's row together. WriteRows and SandRows are exported: cross them.")
	fmt.Printf("%-10s %-10s %8s %10s %11s %9s\n", "WriteRows", "SandRows", "feetInk", "feetSEA%", "feetWASH%", "wetMax")
	for _, cs := range [][2]int{{4, 0}, {0, 0}, {4, 4}, {0, 4}, {4, 5}, {0, 5}, {2, 5}, {3, 6}} {
		a := fwr(cs[0], cs[1])
		lbl := fmt.Sprintf("%d", cs[1])
		if cs[1] == 0 {
			lbl = "auto(6)"
		}
		gate := "ON"
		if cs[0] == 0 {
			gate = "OFF"
		}
		fmt.Printf("%-10s %-10s %8d %10.1f %11.1f %9d   gate %s\n",
			fmt.Sprintf("%d", cs[0]), lbl, a.feetInk,
			pct(a.feetSeaFr, a.frames), pct(a.feetWashFr, a.frames), a.wetMax, gate)
	}

	// ---------------------------------------------------------------- 4
	fmt.Println("\n== 4. TIME OF DAY -- the classifier's dry tone is taken from the palette ==")
	fmt.Printf("%-6s %6s %8s %8s %10s %11s\n", "geom", "hour", "leak", "feetInk", "feetSEA%", "feetWASH%")
	for _, g := range [][2]int{{111, 22}, {111, 10}} {
		for _, hr := range []float64{0, 3, 6, 9, 12, 15, 18, 21} {
			td := hr / 24
			var a res
			for _, sd := range seeds[:2] {
				a = agg(a, newRig(companion.NameCrab, g[0], g[1], sd, td).sample(1.0, 4, td, *warm, 120, true))
			}
			fmt.Printf("%-6s %6.0f %8d %8d %10.1f %11.1f\n",
				fmt.Sprintf("%dx%d", g[0], g[1]), hr, a.leak, a.feetInk,
				pct(a.feetSeaFr, a.frames), pct(a.feetWashFr, a.frames))
		}
	}

	// ---------------------------------------------------------------- 5
	fmt.Println("\n== 5. LITTER SIZE at the wet heights (his p90 is 14, his max 38) ==")
	fmt.Printf("%-6s %4s %5s %5s %8s %10s %11s %9s\n", "animal", "h", "w", "n", "feetInk", "feetSEA%", "feetWASH%", "wetMax")
	for _, h := range []int{14, 17, 22} {
		for _, w := range []int{80, 111, 153} {
			for _, n := range []int{1, 4, 8, 14, 24, 38} {
				var a res
				for _, sd := range seeds[:2] {
					a = agg(a, newRig(companion.NameCrab, w, h, sd, tod).sample(1.0, n, tod, *warm, 150, false))
				}
				if pct(a.feetSeaFr, a.frames) > 0 || pct(a.feetWashFr, a.frames) > 0 || n == 38 {
					fmt.Printf("%-6s %4d %5d %5d %8d %10.1f %11.1f %9d\n", "crab", h, w, n, a.feetInk,
						pct(a.feetSeaFr, a.frames), pct(a.feetWashFr, a.frames), a.wetMax)
				}
			}
		}
	}

	// ---------------------------------------------------------------- 6
	fmt.Println("\n== 6. THE OBSERVABLE THEY REDEFINED: is water DRAWN OVER the sitter? ==")
	fmt.Println("   The RESUME claim was 'the water reaches the sitters and crablets stand in")
	fmt.Println("   it'. s28-litter answers a narrower question (the FEET). This answers the")
	fmt.Println("   wider one, with SKY separated out -- s28-litter counts sky as sea.")
	fmt.Printf("%-6s %-9s %6s %9s %10s %9s %9s %8s\n",
		"animal", "geom", "level", "anySEA%", "anyWASH%", "seaCells", "skyCells", "maxRows")
	for _, nm := range []string{companion.NameCat, companion.NameCrab} {
		for _, g := range [][2]int{{40, 12}, {80, 24}, {111, 22}, {153, 28}, {111, 10}} {
			for _, lv := range []float64{0.5, 0.75, 1.0} {
				var a res
				for _, sd := range seeds {
					a = agg(a, newRig(nm, g[0], g[1], sd, tod).sample(lv, 4, tod, *warm, *frames, false))
				}
				fmt.Printf("%-6s %-9s %6.2f %9.1f %10.1f %9.2f %9.2f %8d\n", nm,
					fmt.Sprintf("%dx%d", g[0], g[1]), lv,
					pct(a.anySeaFr, a.frames), pct(a.anyWashFr, a.frames),
					float64(a.seaCells)/float64(a.frames),
					float64(a.skyCells)/float64(a.frames), a.rowsWet)
			}
		}
	}

	// ---------------------------------------------------------------- 6b
	fmt.Println("\n== 6b. THEIR TWO RECOMMENDED FIXES, MEASURED BEFORE BUILDING THEM ==")
	fmt.Println("   fix 2 'floor the beach at 5' predicts feetWASH at h=9,10,11 -> 0.0.")
	fmt.Println("   SandRows=5 is exactly that change, at those heights:")
	fmt.Printf("%-6s %10s %8s %10s %11s %8s\n", "h", "SandRows", "feetInk", "feetSEA%", "feetWASH%", "wetMax")
	for _, h := range []int{8, 9, 10, 11} {
		for _, sr := range []int{0, 5} {
			var a res
			for _, sd := range seeds {
				r := newRig(companion.NameCrab, 111, h, sd, tod)
				r.sh.SandRows = sr
				a = agg(a, r.sample(1.0, 4, tod, *warm, *frames, false))
			}
			lbl := "auto(4)"
			if sr > 0 {
				lbl = fmt.Sprintf("%d", sr)
			}
			fmt.Printf("%-6d %10s %8d %10.1f %11.1f %8d\n", h, lbl, a.feetInk,
				pct(a.feetSeaFr, a.frames), pct(a.feetWashFr, a.frames), a.wetMax)
		}
	}
	fmt.Println("   fix 1 'MinScapeRows 8 -> 12' predicts feetSEA AND feetWASH 0.0 at every")
	fmt.Println("   producible height. The worst width per height 12..23, litter of 14:")
	fmt.Printf("%-6s %6s %8s %10s %11s %8s\n", "h", "worstW", "feetInk", "feetSEA%", "feetWASH%", "wetMax")
	for h := 12; h <= 23; h++ {
		bw, bp, bwm, bi, bs := 0, -1.0, 0, 0, 0.0
		for w := 40; w <= 200; w += 5 {
			var a res
			for _, sd := range seeds[:2] {
				a = agg(a, newRig(companion.NameCrab, w, h, sd, tod).sample(1.0, 14, tod, *warm, 120, false))
			}
			if p := pct(a.feetWashFr, a.frames); p > bp {
				bw, bp, bwm, bi, bs = w, p, a.wetMax, a.feetInk, pct(a.feetSeaFr, a.frames)
			}
		}
		fmt.Printf("%-6d %6d %8d %10.1f %11.1f %8d\n", h, bw, bi, bs, bp, bwm)
	}

	// ---------------------------------------------------------------- 7
	fmt.Println("\n== 7. host.Band, re-asked: which window heights give a beach of 4 or 5 ==")
	fmt.Printf("%-8s %-8s %-8s %-8s %-8s\n", "window", "scape", "wr", "beach", "clearance(beach-3)")
	for wr := 18; wr <= 70; wr++ {
		_, sc := host.Band(wr)
		if sc == 0 {
			continue
		}
		b, w := beachRows(sc), writeRows(sc)
		if b <= 5 {
			fmt.Printf("%-8d %-8d %-8d %-8d %-8d\n", wr, sc, w, b, b-3)
		}
	}
}

// writeRows / beachRows are Shore.Update's own arithmetic, transcribed.
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
