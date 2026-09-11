// Command s28smallsverify is an ADVERSARIAL re-check of notes/s28-smalls.
//
// It re-asks the sand-across-the-litter question with the ONE input the other
// instrument never varied: the SEED. notes/s28-smalls hardcodes seed 1 in
// three places (warmShore(1,...), DrawKittens(...,1), HashF(i,11,1)); the
// shipped default is 7 (main.go:72, inside.go:50). The seed decides which
// subagents SIT and which SWIM -- swims(i,seed) = i>0 && HashF(i,11,seed)>0.64
// -- so it decides how many crablets are on the beach, which rung of the size
// ladder they take, and how far left the litter reaches.
//
//	go run ./notes/s28-smalls-verify              # everything, ~40s
//	go run ./notes/s28-smalls-verify -part seeds  # the seed sweep only
//	go run ./notes/s28-smalls-verify -part short  # short panes only
//	go run ./notes/s28-smalls-verify -part order  # rendered draw-order check
//
// Everything is read back through canvas.ResolveAt at Profile256; the litter's
// footprint is a DIFF of two rendered frames (shore+parent vs shore+parent+
// litter), never a bounding box.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

// warmShore: ONE shore, 400 steps, so anything that integrates has a history.
func warmShore(seed int64, w, h int, act scape.Activity) (*scape.Shore, float64) {
	sh := scape.NewShore(seed, false)
	sh.MoonX = 0.28
	t := 0.0
	for k := 0; k < 400; k++ {
		t += 0.08
		sh.Update(newCanvas(w, h), t, act)
	}
	return sh, t
}

type cell struct {
	r      rune
	fg, bg term.RGB
}

func readAll(c *canvas.Canvas) []cell {
	out := make([]cell, 0, c.W*c.H)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			out = append(out, cell{r, fg, bg})
		}
	}
	return out
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

type layout struct{ CatX, PaceSpan, SandFrom, SandTo int }

// compose: COPIED from live.go:240, mirrored branch (the shipped one).
func compose(w, catW int) layout {
	const margin = 2
	right := margin + w/32
	span := paceSpan(w)
	catX := w - catW - right
	if catX < 0 {
		catX = 0
	}
	return layout{CatX: catX, PaceSpan: span, SandFrom: margin, SandTo: catX - 1 - span}
}

// composeFlat: COPIED from live.go:247, the UNMIRRORED branch. Nobody measured
// this one; -mirror=false still ships.
func composeFlat(w, catW int) layout {
	const margin = 2
	return layout{CatX: 5, PaceSpan: 0, SandFrom: 5 + catW + 2, SandTo: w - margin}
}

var act = scape.Activity{Working: true, Level: 0.55, ContextUsed: 0.3, TimeOfDay: 13.0 / 24}

type shoreKey struct {
	w, h int
	seed int64
}

var (
	shoreCache = map[shoreKey]*scape.Shore{}
	shoreTs    = map[shoreKey]float64{}
)

func shoreFor(w, h int, seed int64) (*scape.Shore, float64) {
	k := shoreKey{w, h, seed}
	if s, ok := shoreCache[k]; ok {
		return s, shoreTs[k]
	}
	sh, t := warmShore(seed, w, h, act)
	shoreCache[k], shoreTs[k] = sh, t
	return sh, t
}

var sandTopCache = map[shoreKey]int{}

func sandTopFor(w, h int, seed int64) int {
	k := shoreKey{w, h, seed}
	if v, ok := sandTopCache[k]; ok {
		return v
	}
	sh, t := shoreFor(w, h, seed)
	c := newCanvas(w, h)
	sh.Update(c, t+0.08, act)
	sandTopCache[k] = sh.SandTop()
	return sandTopCache[k]
}

// litterInk renders TWO frames -- shore+parent, and shore+parent+litter -- and
// returns the cells that differ and are not blank. That is the litter's INK as
// the terminal would draw it, gaps between the legs excluded.
type inkKey struct {
	w, h, n int
	seed    int64
	animal  string
	flat    bool
}

var (
	inkCache = map[inkKey]map[[2]int]bool{}
	eyeCache = map[inkKey]map[[2]int]bool{}
)

func litterInk(w, h, n int, seed int64, animal string, flat bool) (ink, eyes map[[2]int]bool) {
	k := inkKey{w, h, n, seed, animal, flat}
	if v, ok := inkCache[k]; ok {
		return v, eyeCache[k]
	}
	sh, t := shoreFor(w, h, seed)
	mk := func(withLitter bool) []cell {
		cat := companion.New(animal)
		cat.FaceLeft(!flat)
		cw, ch := cat.Size()
		lay := compose(w, cw)
		if flat {
			lay = composeFlat(w, cw)
		}
		c := newCanvas(w, h)
		sh.Update(c, t+0.08, act)
		top := h - 2 - ch
		cat.Draw(c.Near(), lay.CatX, top, 12.34, companion.Working)
		if withLitter {
			cat.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, n, w-1,
				int(float64(h)*0.42)+1, sh.SandTop()-2, 12.34, seed)
		}
		return readAll(c)
	}
	base, got := mk(false), mk(true)
	ink, eyes = map[[2]int]bool{}, map[[2]int]bool{}
	for i := range base {
		if got[i] == base[i] || got[i].r == ' ' {
			continue
		}
		ink[[2]int{i % w, i / w}] = true
		if got[i].r == 'o' || got[i].r == '-' {
			eyes[[2]int{i % w, i / w}] = true
		}
	}
	inkCache[k], eyeCache[k] = ink, eyes
	return ink, eyes
}

// collides: the sand's glyph cells against the litter's ink mask, with the
// sand's rows placed exactly as drawSand places them (bottom-anchored, clamped
// to the beach's room).
func collides(w, h int, seed int64, animal string, flat bool, kittens int, lines []reduce.Line) (inkHit, eyeHit int) {
	cw := 12
	lay := compose(w, cw)
	if flat {
		lay = composeFlat(w, cw)
	}
	ink, eyes := litterInk(w, h, kittens, seed, animal, flat)
	if len(ink) == 0 {
		return 0, 0
	}
	room := h - sandTopFor(w, h, seed)
	if room < 1 {
		return 0, 0
	}
	ls := lines
	if len(ls) > room {
		ls = ls[len(ls)-room:]
	}
	top := h - len(ls)
	for i, ln := range ls {
		row := top + i
		n := len([]rune(companion.NarrowOnly(ln.Text)))
		for x := lay.SandFrom; x < lay.SandFrom+n && x < lay.SandTo; x++ {
			if ink[[2]int{x, row}] {
				inkHit++
				if eyes[[2]int{x, row}] {
					eyeHit++
				}
			}
		}
	}
	return inkHit, eyeHit
}

func readEvents(path string) []event.Event {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var evs []event.Event
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		e, err := event.Decode(line)
		if err != nil || e.TS == 0 {
			continue
		}
		evs = append(evs, e)
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].TS < evs[j].TS })
	return evs
}

// combo is one (width, height, seed, animal, layout) question.
type combo struct {
	w, h   int
	seed   int64
	animal string
	flat   bool
}

type acc struct {
	samples, withLitter, both, collide, cells, eyes int
	sitterSum, sitterN                              int
	eyeFrames                                       int
}

// fold walks every recording once and asks every combo at every sampled second.
func fold(combos []combo) (map[combo]*acc, int, int, int, [64]int) {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "xscapes", "run")
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	sort.Strings(files)
	res := map[combo]*acc{}
	for _, c := range combos {
		res[c] = &acc{}
	}
	sessions, events, samples := 0, 0, 0
	var kittenHist [64]int
	for _, f := range files {
		evs := readEvents(f)
		if len(evs) < 20 {
			continue
		}
		sessions++
		events += len(evs)
		r := reduce.New("s28v")
		at := func(ms int64) time.Time { return time.UnixMilli(ms) }
		start, end := at(evs[0].TS), at(evs[len(evs)-1].TS)
		i := 0
		for t := start; !t.After(end); t = t.Add(time.Second) {
			for i < len(evs) && !at(evs[i].TS).After(t) {
				r.Apply(evs[i], at(evs[i].TS))
				i++
			}
			st := r.State(t)
			samples++
			if st.Kittens >= 0 && st.Kittens < len(kittenHist) {
				kittenHist[st.Kittens]++
			}
			for _, cb := range combos {
				a := res[cb]
				a.samples++
				if st.Kittens > 0 {
					a.withLitter++
				}
				lay := compose(cb.w, 12)
				if cb.flat {
					lay = composeFlat(cb.w, 12)
				}
				lines := st.FitTail(t, lay.SandTo-lay.SandFrom)
				if st.Kittens == 0 || len(lines) == 0 {
					continue
				}
				a.both++
				sit := 0
				for k := 0; k < st.Kittens; k++ {
					if !(k > 0 && companion.HashF(k, 11, cb.seed) > 0.64) {
						sit++
					}
				}
				a.sitterSum += sit
				a.sitterN++
				hit, eye := collides(cb.w, cb.h, cb.seed, cb.animal, cb.flat, st.Kittens, lines)
				if hit > 0 {
					a.collide++
					a.cells += hit
					a.eyes += eye
					if eye > 0 {
						a.eyeFrames++
					}
				}
			}
			if i < len(evs) {
				if gap := at(evs[i].TS).Sub(t); gap > 3*time.Minute {
					t = at(evs[i].TS).Add(-time.Second)
				}
			}
		}
	}
	return res, sessions, events, samples, kittenHist
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return 100 * float64(a) / float64(b)
}

// ------------------------------------------------------------------ parts

func partSeeds() {
	fmt.Println("\n=========== THE SEED, WHICH THEY NEVER VARIED ===========")
	fmt.Println("notes/s28-smalls hardcodes seed 1. The shipped default is 7")
	fmt.Println("(main.go:72 `seed = flag.Int64(\"seed\", 7, ...)`, inside.go:50 the same).")
	fmt.Println("swims(i,seed) = i>0 && HashF(i,11,seed)>0.64 decides who sits.")
	fmt.Println("\nSITTERS out of n, per seed (i=0 always sits):")
	fmt.Printf("  %6s", "seed")
	ns := []int{1, 2, 4, 6, 8, 10, 14, 20, 30, 38}
	for _, n := range ns {
		fmt.Printf(" n=%-3d", n)
	}
	fmt.Println()
	for _, sd := range []int64{1, 7, 42, 999, 12345} {
		fmt.Printf("  %6d", sd)
		for _, n := range ns {
			sit := 0
			for k := 0; k < n; k++ {
				if !(k > 0 && companion.HashF(k, 11, sd) > 0.64) {
					sit++
				}
			}
			fmt.Printf(" %5d", sit)
		}
		fmt.Println()
	}

	widths := []int{60, 80, 100, 107, 124, 153}
	seeds := []int64{1, 7, 42, 999, 12345}
	var combos []combo
	for _, sd := range seeds {
		for _, w := range widths {
			combos = append(combos, combo{w, 51, sd, companion.NameCrab, false})
		}
	}
	res, sessions, events, samples, kh := fold(combos)
	nz := 0
	for k, v := range kh {
		if k > 0 {
			nz += v
		}
	}
	fmt.Printf("\nSAMPLE: %d recordings folded, %d events, %d one-second samples\n", sessions, events, samples)
	fmt.Printf("  samples with a litter on screen: %d (%.2f%%) -- NOT an empty sample\n", nz, pct(nz, samples))

	fmt.Println("\nCOLLISION RATE (%% of samples that have BOTH a litter and a tail), h=51, crab:")
	fmt.Printf("  %6s", "seed")
	for _, w := range widths {
		fmt.Printf("   w=%-5d", w)
	}
	fmt.Println("   mean sitters")
	for _, sd := range seeds {
		fmt.Printf("  %6d", sd)
		var ms float64
		for _, w := range widths {
			a := res[combo{w, 51, sd, companion.NameCrab, false}]
			fmt.Printf("  %6.2f%%  ", pct(a.collide, a.both))
			ms = float64(a.sitterSum) / float64(max(1, a.sitterN))
		}
		fmt.Printf("   %.2f\n", ms)
	}
	fmt.Println("\nEYE CELLS LOST, same sweep (0 everywhere would confirm their claim):")
	for _, sd := range seeds {
		fmt.Printf("  seed %-6d", sd)
		for _, w := range widths {
			a := res[combo{w, 51, sd, companion.NameCrab, false}]
			fmt.Printf(" w%d=%d", w, a.eyes)
		}
		fmt.Println()
	}
	fmt.Println("\nDENOMINATORS (identical across seeds by construction; printed so a zero row is visible):")
	for _, sd := range seeds {
		a := res[combo{107, 51, sd, companion.NameCrab, false}]
		fmt.Printf("  seed %-6d w=107  both=%d  collide=%d  cells=%d\n", sd, a.both, a.collide, a.cells)
	}
}

func partOthers() {
	fmt.Println("\n=========== TWO MORE THEY NEVER VARIED: SHORT PANES, AND -mirror=false ===========")
	// Their heights were 24/40/51. The brief promises 40x12.
	var combos []combo
	heights := []int{12, 16, 20, 24, 51}
	for _, h := range heights {
		combos = append(combos, combo{107, h, 7, companion.NameCrab, false})
		combos = append(combos, combo{80, h, 7, companion.NameCrab, false})
	}
	// the unmirrored layout, which still ships behind -mirror=false
	combos = append(combos, combo{107, 51, 7, companion.NameCrab, true})
	combos = append(combos, combo{124, 51, 7, companion.NameCrab, true})
	// the cat, over the whole log rather than 3 spot checks
	for _, w := range []int{80, 107, 124} {
		combos = append(combos, combo{w, 51, 7, companion.NameCat, false})
	}
	res, _, _, samples, _ := fold(combos)

	fmt.Println("\nSHORT PANES (seed 7, crab, mirrored). room = beach rows available to the sand:")
	fmt.Printf("  %5s %5s %8s %8s %10s %9s %8s\n", "w", "h", "sandTop", "room", "sandrows", "both", "collide%")
	for _, h := range heights {
		for _, w := range []int{80, 107} {
			a := res[combo{w, h, 7, companion.NameCrab, false}]
			st := sandTopFor(w, h, 7)
			room := h - st
			nl := 4
			if nl > room {
				nl = room
			}
			fmt.Printf("  %5d %5d %8d %8d %10s %8d %7.2f%%\n", w, h, st, room,
				fmt.Sprintf("%d..%d", h-nl, h-1), a.both, pct(a.collide, a.both))
		}
	}
	fmt.Println("\nUNMIRRORED LAYOUT (-mirror=false, live.go:247; sand runs left->right, litter grows LEFT from x=5):")
	for _, w := range []int{107, 124} {
		a := res[combo{w, 51, 7, companion.NameCrab, true}]
		lay := composeFlat(w, 12)
		ink, _ := litterInk(w, 51, 8, 7, companion.NameCrab, true)
		fmt.Printf("  w=%d  sand cols %d..%d  litter ink cells visible on screen at n=8: %d  collide %.2f%% of %d\n",
			w, lay.SandFrom, lay.SandTo-1, len(ink), pct(a.collide, a.both), a.both)
	}
	fmt.Println("\nTHE CAT, over the whole log rather than three spot checks (seed 7, h=51):")
	for _, w := range []int{80, 107, 124} {
		a := res[combo{w, 51, 7, companion.NameCat, false}]
		ac := res[combo{w, 51, 7, companion.NameCrab, false}]
		_ = ac
		fmt.Printf("  w=%-4d cat: collide %.2f%% of %d, cells %d, eyes %d\n",
			w, pct(a.collide, a.both), a.both, a.cells, a.eyes)
	}
	_ = samples
}

// partOrder re-asks "who wins the cell" at seed 7 by rendering the whole scene
// twice -- once with drawSand, once without -- and comparing the resolved runes.
func partOrder() {
	fmt.Println("\n=========== WHO WINS THE CELL, AT SEED 7 ===========")
	w, h := 124, 51
	seed := int64(7)
	sh, t := shoreFor(w, h, seed)
	// a synthetic tail long enough to reach, made of REAL-shaped text
	mkLines := func(n int, txt string) []reduce.Line {
		out := make([]reduce.Line, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, reduce.Line{Text: txt, Age: float64(n-1-i) / float64(max(1, n-1))})
		}
		return out
	}
	fmt.Printf("  %8s %10s %9s %9s %10s  %s\n", "kittens", "overlap", "sandwon", "litwon", "eyeslost", "example")
	tot, totSand := 0, 0
	for _, n := range []int{1, 2, 4, 8, 14, 20, 38} {
		cat := companion.New(companion.NameCrab)
		cat.FaceLeft(true)
		cw, ch := cat.Size()
		lay := compose(w, cw)
		top := h - 2 - ch
		budget := lay.SandTo - lay.SandFrom
		txt := ""
		for len(txt) < budget {
			txt += "edit internal/host/screen.go "
		}
		txt = txt[:budget]
		lines := mkLines(4, txt)

		render := func(withSand bool) []cell {
			c := newCanvas(w, h)
			sh.Update(c, t+0.08, act)
			cat.Draw(c.Near(), lay.CatX, top, 12.34, companion.Working)
			cat.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, n, w-1,
				int(float64(h)*0.42)+1, sh.SandTop()-2, 12.34, seed)
			if withSand {
				drawSandCopy(c, lines, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
			}
			return readAll(c)
		}
		noSand, withSand := render(false), render(true)
		// the sand's own footprint, measured on a frame with NO litter
		c2 := newCanvas(w, h)
		sh.Update(c2, t+0.08, act)
		bare := readAll(c2)
		c3 := newCanvas(w, h)
		sh.Update(c3, t+0.08, act)
		drawSandCopy(c3, lines, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
		sandOnly := readAll(c3)

		ink, eyes := litterInk(w, h, n, seed, companion.NameCrab, false)
		over, sandWon, litWon, eyeLost := 0, 0, 0, 0
		ex := ""
		for i := range withSand {
			x, y := i%w, i/w
			if !ink[[2]int{x, y}] {
				continue
			}
			if sandOnly[i] == bare[i] {
				continue // the sand put nothing here
			}
			over++
			if withSand[i].r == sandOnly[i].r {
				sandWon++
				if ex == "" {
					ex = fmt.Sprintf("(%d,%d) %q -> %q", x, y, string(noSand[i].r), string(withSand[i].r))
				}
			} else if withSand[i].r == noSand[i].r {
				litWon++
			}
			if eyes[[2]int{x, y}] && withSand[i].r != noSand[i].r {
				eyeLost++
			}
		}
		tot += over
		totSand += sandWon
		fmt.Printf("  %8d %10d %9d %9d %10d  %s\n", n, over, sandWon, litWon, eyeLost, ex)
	}
	fmt.Printf("  TOTAL overlapping cells %d, sand won %d (%.1f%%)\n", tot, totSand, pct(totSand, tot))
	if tot == 0 {
		fmt.Println("  !! ZERO overlap -- this pass would be an EMPTY SAMPLE and must not be quoted")
	}
}

// drawSandCopy is COPIED from live.go:450.
func drawSandCopy(c *canvas.Canvas, lines []reduce.Line, sand term.RGB, sandTop, xFrom, xTo int) {
	if len(lines) == 0 || xTo-xFrom < 12 {
		return
	}
	bad := term.RGB{R: 244, G: 176, B: 96}
	beachAt := func(row int) term.RGB {
		if row < 0 || row >= c.H || len(c.BG) < c.W*c.H {
			return sand
		}
		var r, g, b, n int
		for x := xFrom; x < xTo && x < c.W; x += 4 {
			p := c.BG[row*c.W+x]
			r, g, b, n = r+int(p.R), g+int(p.G), b+int(p.B), n+1
		}
		if n == 0 {
			return sand
		}
		return term.RGB{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n)}
	}
	room := c.H - sandTop
	if room < 1 {
		return
	}
	if len(lines) > room {
		lines = lines[len(lines)-room:]
	}
	top := c.H - len(lines)
	for i, ln := range lines {
		row := top + i
		if row < 0 || row >= c.H {
			continue
		}
		beach := beachAt(row)
		base := term.RGB{R: 244, G: 236, B: 220}
		if 0.2126*float64(beach.R)+0.7152*float64(beach.G)+0.0722*float64(beach.B) > 140 {
			base = term.RGB{R: 34, G: 26, B: 20}
		}
		if ln.Bad {
			base = bad
		}
		col := term.Lerp(base, beach, 0.10+0.62*ln.Age)
		x := xFrom
		for _, r := range []rune(companion.NarrowOnly(ln.Text)) {
			if x >= xTo {
				break
			}
			c.Near().Plot(x, row, r, col, 1)
			x++
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	part := flag.String("part", "all", "seeds | short | order | cat | all")
	flag.Parse()
	fmt.Printf("scape.Tide = %v  (XSCAPES_TIDE=%q)\n", scape.Tide, os.Getenv("XSCAPES_TIDE"))
	if *part == "order" || *part == "all" {
		partOrder()
	}
	if *part == "short" || *part == "all" {
		partOthers()
	}
	if *part == "cat" || *part == "all" {
		partCat()
	}
	if *part == "seeds" || *part == "all" {
		partSeeds()
	}
}

// ---------------------------------------------------------------- the cat

// partCat is the question notes/s28-smalls never asked. Its litterInk() is
// hardcoded to companion.NameCrab, so every "eyes lost = 0" in it is a
// statement about the CRAB only. The cat still ships (`xscapes companion cat`).
func partCat() {
	fmt.Println("\n=========== THE CAT'S FACE, WHICH THEY NEVER ASKED ABOUT ===========")
	fmt.Println("notes/s28-smalls:1105  func litterInk(...) { return litterInkAnimal(companion.NameCrab, ...) }")
	fmt.Println("-> collides() only ever built a CRAB mask, so its eye column is crab-only.")
	fmt.Println("kittens.go:193 pendingEyes{e1, e2, ky+1, ...}  vs  crab_kittens.go:155 pendingEyes{..., ky, ...}")
	fmt.Println("The cat's kitten eyes sit ONE ROW LOWER inside the sprite than the crablet's.")

	w, h := 124, 51
	fmt.Printf("\nGEOMETRY at h=%d (seed 7): sand rows are %d..%d for a 4-line tail\n", h, h-4, h-1)
	fmt.Printf("  %6s %8s %14s %14s %10s\n", "animal", "kittens", "litter rows", "EYE rows", "in sand?")
	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		for _, n := range []int{1, 2, 8, 14, 20, 38} {
			ink, eyes := litterInk(w, h, n, 7, animal, false)
			r0, r1, e0, e1 := 999, -1, 999, -1
			for c := range ink {
				if c[1] < r0 {
					r0 = c[1]
				}
				if c[1] > r1 {
					r1 = c[1]
				}
			}
			for c := range eyes {
				if c[1] < e0 {
					e0 = c[1]
				}
				if c[1] > e1 {
					e1 = c[1]
				}
			}
			inSand := 0
			for c := range eyes {
				if c[1] >= h-4 {
					inSand++
				}
			}
			fmt.Printf("  %6s %8d %14s %14s %10d  (%d eye cells total)\n", animal, n,
				fmt.Sprintf("%d..%d", r0, r1), fmt.Sprintf("%d..%d", e0, e1), inSand, len(eyes))
		}
	}

	fmt.Println("\nRENDERED, full-budget 4-line tail, sand drawn LAST as drawScene draws it:")
	fmt.Printf("  %6s %6s %8s %9s %9s %9s %10s  %s\n", "animal", "seed", "kittens", "overlap", "sandwon", "litwon", "EYESLOST", "example")
	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		for _, seed := range []int64{1, 7} {
			for _, n := range []int{8, 14, 20, 38} {
				over, sw, lw, el, ex := renderOne(w, h, seed, animal, n)
				fmt.Printf("  %6s %6d %8d %9d %9d %9d %10d  %s\n", animal, seed, n, over, sw, lw, el, ex)
			}
		}
	}

	fmt.Println("\nOVER THE REAL LOG: frames with at least one CAT kitten eye eaten by the sand")
	var combos []combo
	for _, seed := range []int64{1, 7} {
		for _, w := range []int{80, 107, 124} {
			combos = append(combos, combo{w, 51, seed, companion.NameCat, false})
			combos = append(combos, combo{w, 51, seed, companion.NameCrab, false})
		}
	}
	res, _, _, samples, _ := fold(combos)
	fmt.Printf("  (%d one-second samples; 114360 of them have both a litter and a tail)\n", samples)
	fmt.Printf("  %6s %6s %6s %12s %12s %12s\n", "animal", "seed", "w", "collide%", "eyecells", "eyeframes")
	for _, animal := range []string{companion.NameCat, companion.NameCrab} {
		for _, seed := range []int64{1, 7} {
			for _, w := range []int{80, 107, 124} {
				a := res[combo{w, 51, seed, animal, false}]
				fmt.Printf("  %6s %6d %6d %11.2f%% %12d %12d\n", animal, seed, w,
					pct(a.collide, a.both), a.eyes, a.eyeFrames)
			}
		}
	}
}

// renderOne renders the scene with and without the sand and reports who won.
func renderOne(w, h int, seed int64, animal string, n int) (over, sandWon, litWon, eyeLost int, ex string) {
	sh, t := shoreFor(w, h, seed)
	cat := companion.New(animal)
	cat.FaceLeft(true)
	cw, ch := cat.Size()
	lay := compose(w, cw)
	top := h - 2 - ch
	budget := lay.SandTo - lay.SandFrom
	txt := ""
	for len(txt) < budget {
		txt += "edit internal/host/screen.go "
	}
	txt = txt[:budget]
	lines := make([]reduce.Line, 4)
	for i := range lines {
		lines[i] = reduce.Line{Text: txt, Age: float64(3-i) / 3}
	}
	paint := func(withLitter, withSand bool) []cell {
		c := newCanvas(w, h)
		sh.Update(c, t+0.08, act)
		cat.Draw(c.Near(), lay.CatX, top, 12.34, companion.Working)
		if withLitter {
			cat.DrawKittens(c.Near(), c.Mid(), lay.CatX-lay.PaceSpan, top, n, w-1,
				int(float64(h)*0.42)+1, sh.SandTop()-2, 12.34, seed)
		}
		if withSand {
			drawSandCopy(c, lines, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
		}
		return readAll(c)
	}
	noSand, withSand := paint(true, false), paint(true, true)
	bare, sandOnly := paint(false, false), paint(false, true)
	ink, eyes := litterInk(w, h, n, seed, animal, false)
	for i := range withSand {
		x, y := i%w, i/w
		if !ink[[2]int{x, y}] || sandOnly[i] == bare[i] {
			continue
		}
		over++
		if withSand[i].r == sandOnly[i].r {
			sandWon++
			if ex == "" {
				ex = fmt.Sprintf("(%d,%d) %q -> %q", x, y, string(noSand[i].r), string(withSand[i].r))
			}
		} else if withSand[i].r == noSand[i].r {
			litWon++
		}
		if eyes[[2]int{x, y}] && withSand[i].r != noSand[i].r {
			eyeLost++
			ex = fmt.Sprintf("EYE at (%d,%d) %q -> %q", x, y, string(noSand[i].r), string(withSand[i].r))
		}
	}
	return over, sandWon, litWon, eyeLost, ex
}
