// Command s28-preview shows HIS near-ask idea in HIS terminal, at HIS window
// size, through the real canvas and the real 256-colour quantiser.
//
//	go run ./notes/s28-preview
//
// The precedent is `xscapes shades`, and this follows its shape deliberately:
// a rendering decision is settled by printing the alternatives into the
// terminal the thing actually runs in, one under the other, and looking. A
// browser preview is a different instrument -- "an RGB HTML preview is NOT a
// 256 preview" is a lesson this project already paid for.
//
// Four sections, in this order:
//
//	1  the two poses side by side, bottom-aligned, on the beach's own colour
//	2  both composited into a REAL shore frame at this window's size
//	3  the approach, animated on the pace's own clock (stepDur = 0.28 s)
//	4  the eye, both ways, with the ring's hole colour measured off the frame
//
// Flags, all optional:
//
//	-w, -h      force a canvas size (default: this window, one row kept back
//	            for the prompt, exactly as the live loop does)
//	-level      activity level (default 0.59, the measured median at an ask)
//	-tod        hour of day, 0..24 (default: this machine's clock)
//	-animal     "cat" also shows the cat, for comparison
//	-still      skip the animation, so it is safe in a pipe or a screenshot
//	            (also forced automatically when stdout is not a terminal)
//	-mid        include A's 14x9 as the in-between frame (default true)
//	-litter     how many crablets sit beside the companion (default 2)
//	-seed       scene seed (default 7)
//
// Nothing here writes to the product. Nothing here ships.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// stepDur is pace.go's own constant, not a new one: one cell of the
// companion's walk always takes this long, and the approach is measured in
// those steps rather than in a duration invented for the occasion.
const stepDur = 0.28

var prof term.Profile

func main() {
	// The same two lines the product runs before anything paints. Terminal.app
	// draws U+2580 with a gap above it and leaves a 1px rule at the cell's
	// bottom edge either way up, so both switches change the picture. A preview
	// that skips them is previewing another terminal.
	term.LowerHalf = term.DetectSplit(os.Getenv("TERM_PROGRAM"))
	term.NoSplitCells = term.DetectNoSplit(os.Getenv("TERM_PROGRAM"))

	var (
		wIn    = flag.Int("w", 0, "canvas width (default: this window)")
		hIn    = flag.Int("h", 0, "canvas height (default: this window, less one row)")
		level  = flag.Float64("level", 0.59, "activity level at the ask")
		todIn  = flag.Float64("tod", -1, "hour of day 0..24 (default: this machine's clock)")
		animal = flag.String("animal", "", `"cat" also shows the cat for comparison`)
		still  = flag.Bool("still", false, "skip the animation")
		useMid = flag.Bool("mid", true, "include A's 14x9 as the in-between frame")
		litter = flag.Int("litter", 2, "crablets beside the companion")
		seed   = flag.Int64("seed", 7, "scene seed")
	)
	flag.Parse()

	prof = term.DetectProfile()

	tw, th, isTTY := windowSize()
	w, h := *wIn, *hIn
	if w <= 0 {
		w = tw
	}
	if h <= 0 {
		h = th - 1 // the live loop keeps a row back for the prompt
	}
	if !isTTY {
		*still = true
	}
	if w < 40 || h < 14 {
		fmt.Fprintf(os.Stderr,
			"note: %dx%d is smaller than this preview needs. Try -w 143 -h 27.\n", w, h)
	}

	tod := *todIn
	if tod < 0 {
		now := time.Now()
		tod = float64(now.Hour()) + float64(now.Minute())/60
	}
	todFrac := math.Mod(tod, 24) / 24

	act := scape.Activity{
		Working: true, Level: *level, ContextUsed: 0.30, TimeOfDay: todFrac,
	}

	header(w, h, tod, act, isTTY)

	sh, t := warmShore(w, h, *seed, act, 0.28)

	poses := []*pose{shippedPose(), nearPose()}
	if strings.Contains(strings.ToLower(*animal), "cat") {
		poses = append(poses, catPose())
	}

	// The ground for section 1 is MEASURED off the real frame section 2 draws,
	// so build that frame first and only then print, in order.
	ground, gN, gTot := groundUnderCompanion(sh, w, h, t, act, *seed)
	sectionPoses(poses, ground, gN, gTot, t)
	sectionScene(sh, w, h, t, act, poses, *litter, *seed)
	sectionApproach(sh, w, h, t, act, *litter, *seed, *still, *useMid, isTTY)
	sectionEye(sh, w, h, t, act, *seed)

	fmt.Printf("\n\x1b[2mRun it again with -tod 11 for midday, -tod 21 for the evening, "+
		"-level 0.9 for a busy sea,\n-animal cat to put the cat beside it, "+
		"-still for a screenshot. This window: %dx%d.\x1b[0m\n", w, h)
}

// ---------------------------------------------------------------------------

func header(w, h int, tod float64, act scape.Activity, isTTY bool) {
	fmt.Printf("\n\x1b[1mxscapes -- the near ask, in this terminal\x1b[0m\n")
	fmt.Printf("\x1b[2m%dx%d  ·  %s  ·  TERM_PROGRAM=%s  ·  split=%s  no-split-cells=%v  ·  "+
		"%02d:%02d  ·  level %.2f  ·  tty=%v\x1b[0m\n",
		w, h, prof, os.Getenv("TERM_PROGRAM"), splitName(), term.NoSplitCells,
		int(tod), int(math.Mod(tod, 1)*60), act.Level, isTTY)
	if prof != term.Profile256 {
		fmt.Fprintf(os.Stderr,
			"\nnote: this terminal reports %s. The scape is built for the 256 cube;\n"+
				"      run with XSCAPES_COLOR=256 to see what Terminal.app gets.\n", prof)
	}
	same, total, samples, ink := replicaCheck()
	verdict := "\x1b[31mDIFFERS -- do not trust the large poses\x1b[0m"
	if same == total {
		verdict = "identical"
	}
	if ink == 0 {
		verdict = "\x1b[31mVOID -- nothing was drawn\x1b[0m"
	}
	fmt.Printf("\x1b[2mreplica check: this package's drawCrab replica against companion's own, on\n"+
		"the SHIPPED rows. %d of %d resolved cells match over %d breathing samples: %s\n"+
		"(%d of those cells carry sprite ink, so the panels were not empty; and moving the\n"+
		"replica's eye row by one makes the same check fail in %d cells, so it can fail)\x1b[0m\n",
		same, total, samples, verdict, ink, negativeControl())
}

func splitName() string {
	if term.LowerHalf {
		return "lower"
	}
	return "upper"
}

// replicaCheck renders the shipped crab twice -- once through
// companion.NewCrab().Draw, once through this package's replica of drawCrab --
// and compares every resolved cell.
//
// Without this the large poses would be drawn by code nothing has ever
// checked, which is the harness that returns a clean bill of health for a
// picture the product would never paint.
func replicaCheck() (same, total, samples, ink int) {
	bg := term.RGB{R: 40, G: 36, B: 32}
	a := shippedPose()
	b := replicaShipped()
	for k := 0; k < 240; k++ {
		t := float64(k) * 0.05 // 12 s: past the 1.6 s breath and the 5.3 s blink
		ca := flatPanel(20, 11, bg)
		cb := flatPanel(20, 11, bg)
		a.draw(ca.Near(), 3, 2, t)
		b.draw(cb.Near(), 3, 2, t)
		for y := 0; y < ca.H; y++ {
			for x := 0; x < ca.W; x++ {
				r1, f1, g1 := ca.ResolveAt(x, y, prof)
				r2, f2, g2 := cb.ResolveAt(x, y, prof)
				total++
				if r1 == r2 && f1 == f2 && g1 == g2 {
					same++
				}
				// Count the ink. Two BLANK canvases match perfectly, and a
				// pass from a panel with nothing in it looks exactly like a
				// real pass -- so the ink count is printed beside the verdict.
				if r1 != ' ' {
					ink++
				}
			}
		}
		samples++
	}
	return same, total, samples, ink
}

// negativeControl runs the same comparison against a replica that is wrong on
// purpose. A check that cannot fail proves nothing; this says how many cells it
// catches when the eye is moved one row.
func negativeControl() int {
	bg := term.RGB{R: 40, G: 36, B: 32}
	a := shippedPose()
	b := replicaShipped()
	b.eyeRow++
	diff := 0
	for k := 0; k < 240; k++ {
		t := float64(k) * 0.05
		ca := flatPanel(20, 11, bg)
		cb := flatPanel(20, 11, bg)
		a.draw(ca.Near(), 3, 2, t)
		b.draw(cb.Near(), 3, 2, t)
		for y := 0; y < ca.H; y++ {
			for x := 0; x < ca.W; x++ {
				r1, f1, g1 := ca.ResolveAt(x, y, prof)
				r2, f2, g2 := cb.ResolveAt(x, y, prof)
				if r1 != r2 || f1 != f2 || g1 != g2 {
					diff++
				}
			}
		}
	}
	return diff
}

// ---------------------------------------------------------------------------
// 1. THE TWO POSES SIDE BY SIDE
// ---------------------------------------------------------------------------

func sectionPoses(poses []*pose, bg term.RGB, gN, gTot int, t float64) {
	rule("1. THE TWO POSES, side by side, bottom-aligned on a common feet row.")
	fmt.Printf("   Mirrored as shipped, in real colour, on flat ground -- rgb(%d,%d,%d), index %d.\n"+
		"   Not a colour I picked: it is the commonest background the terminal paints under\n"+
		"   the shipped companion in section 2's frame, %d of its %d cells, read back with\n"+
		"   ResolveAt. Size is the only difference here.\n\n",
		bg.R, bg.G, bg.B, bg.Index256Keeping(), gN, gTot)

	lbl, art := poseStrip(poses, bg, t, true)
	fmt.Println(lbl)
	fmt.Println(art)
	fmt.Printf("\n   \x1b[2mBottom-aligned isolates the size. In the SCENE they share a TOP row and the\n" +
		"   near pose grows DOWNWARD into the two rows already under the shipped crab's\n" +
		"   feet -- section 2 shows that, and it is the anchoring being proposed.\x1b[0m\n")
}

// poseStrip lays poses out on one flat canvas, so the whole strip is a single
// Render and the terminal never has to interleave two coloured blocks. Each
// panel is as wide as the wider of its sprite and its label, so the labels
// cannot run into one another at any size.
func poseStrip(poses []*pose, bg term.RGB, t float64, bottomAlign bool) (labels, art string) {
	const gap, margin = 4, 2
	maxH := 0
	total := margin
	xs := make([]int, len(poses))
	labs := make([]string, len(poses))
	for i, p := range poses {
		pw, ph := p.size()
		if ph > maxH {
			maxH = ph
		}
		labs[i] = fmt.Sprintf("%dx%d  %s", pw, ph, p.name)
		cell := pw
		if n := len([]rune(labs[i])); n > cell {
			cell = n
		}
		xs[i] = total
		total += cell + gap
	}
	total = total - gap + margin
	H := maxH + 2
	c := flatPanel(total, H, bg)
	for i, p := range poses {
		_, ph := p.size()
		y := 1
		if bottomAlign {
			y = 1 + maxH - ph
		}
		p.draw(c.Near(), xs[i], y, t)
	}
	row := []rune(strings.Repeat(" ", total))
	for i := range poses {
		for j, r := range []rune(labs[i]) {
			if xs[i]+j < len(row) {
				row[xs[i]+j] = r
			}
		}
	}
	return "  \x1b[2m" + strings.TrimRight(string(row), " ") + "\x1b[0m", indent(c.Render(prof), 2)
}

// ---------------------------------------------------------------------------
// 2. BOTH COMPOSITED INTO A REAL SHORE FRAME
// ---------------------------------------------------------------------------

// groundUnderCompanion is the background the shipped companion actually stands
// on most in a real frame. Section 1 paints its panels with it rather than with
// a colour chosen to flatter the coat.
func groundUnderCompanion(sh *scape.Shore, w, h int, t float64, act scape.Activity, seed int64) (term.RGB, int, int) {
	p := shippedPose()
	c, lay, top := buildFrame(sh, frameSpec{
		w: w, h: h, t: t, act: act, p: p,
		bubble: "allow Bash?", litter: 0, seed: seed, tail: sampleTail(),
	})
	pw, ph := p.size()
	return dominantBG(c, lay.CatX, top, lay.CatX+pw-1, top+ph-1)
}

func sectionScene(sh *scape.Shore, w, h int, t float64, act scape.Activity,
	poses []*pose, litter int, seed int64) {

	rule("2. THE SAME TWO IN A REAL SHORE FRAME, at this window's size.")
	fmt.Printf("   Level %.2f (his measured median at an ask), the balloon up, %d crablets at the\n"+
		"   real litter baseline, the real sand tail. One shore, warmed 400 steps, updated at\n"+
		"   the SAME t for every frame -- so the sea is identical and the companion is the only\n"+
		"   thing that changed.\n\n", act.Level, litter)

	lay := compose(w)
	top := h - 2 - shippedH

	// The head column, asserted and printed. This is the whole reason the near
	// pose is drawn at catX-2.
	headReport(lay.CatX, poses)

	for _, p := range poses {
		c, _, _ := buildFrame(sh, frameSpec{
			w: w, h: h, t: t, act: act, p: p,
			bubble: "allow Bash?", litter: litter, seed: seed, tail: sampleTail(),
		})
		pw, ph := p.size()
		x0 := lay.CatX + p.xoff
		wet, feet := waterRowsUnder(sh.SandTop(), top, ph)
		// The pointer sits on the balloon's LAST row, which is the row
		// immediately above the companion's top.
		v := pointerCol(c, top-1)
		fmt.Printf("\n   \x1b[1m%s -- %dx%d at catX%+d\x1b[0m\n", p.name, pw, ph, p.xoff)
		fmt.Printf("   \x1b[2mbox columns %d..%d  ·  top row %d, feet row %d (canvas is 0..%d)\n"+
			"   waterline (the shore's own SandTop) row %d  ·  rows standing in the sea: %d of %d\n"+
			"   the pace strip is %d columns wide and reserved from the sand and the litter;\n"+
			"   this sprite takes %d of them, leaving %d  ·  the sand tail stops at column %d\n"+
			"   balloon pointer 'v' read back from the RENDERED frame: column %d\x1b[0m\n",
			x0, x0+pw-1, top, feet, h-1, sh.SandTop(), wet, ph,
			lay.PaceSpan, -p.xoff, lay.PaceSpan+p.xoff, lay.SandTo, v)
		fmt.Println(c.Render(prof))
	}
	fmt.Printf("\n   \x1b[2mThe waterline is drawn as it falls. A raised claw standing in the sea is\n" +
		"   shown standing in the sea -- the measured beach is 7-8 rows at every width he\n" +
		"   runs, at both activity levels, so the tide buys the near pose no extra room.\x1b[0m\n")
}

// headReport asserts, and prints, that the near pose leaves the head where it
// was. The balloon is anchored to lay.CatX + HeadCol() and nothing else; if
// the eye centre moves, the pointer stops aiming at the face.
func headReport(catX int, poses []*pose) {
	anchor := catX + shippedHeadCol()
	fmt.Printf("   \x1b[1mHEAD COLUMN\x1b[0m  balloon anchor = catX + HeadCol() = %d + %d = \x1b[1m%d\x1b[0m\n",
		catX, shippedHeadCol(), anchor)
	want, first := 0.0, true
	for _, p := range poses {
		if !p.crabEyes {
			continue // the cat's eye cells are its own; not part of this claim
		}
		ctr := p.eyeCentre(catX)
		if first {
			want, first = ctr, false
		}
		mark := "\x1b[31mMOVED\x1b[0m"
		if ctr == want {
			mark = "unchanged"
		}
		pw, ph := p.size()
		fmt.Printf("     %-17s %2dx%d at catX%+d: eyes %s, centre %.1f   %s\n",
			p.name, pw, ph, p.xoff, spanStr(p.eyeSpanAbs(catX)), ctr, mark)
	}
	// The counterfactual, so -2 is justified rather than asserted.
	bad := nearPose()
	bad.xoff = 0
	fmt.Printf("     %-17s 16x9 at catX+0: eyes %s, centre %.1f   \x1b[31m%+.0f columns off -- "+
		"this is what catX-2 buys\x1b[0m\n",
		"near, MISPLACED", spanStr(bad.eyeSpanAbs(catX)), bad.eyeCentre(catX),
		bad.eyeCentre(catX)-want)
}

func spanStr(sp [][2]int) string {
	s := append([][2]int(nil), sp...)
	sort.Slice(s, func(i, j int) bool { return s[i][0] < s[j][0] })
	var parts []string
	for _, s := range s {
		if s[0] == s[1] {
			parts = append(parts, fmt.Sprint(s[0]))
			continue
		}
		parts = append(parts, fmt.Sprintf("%d-%d", s[0], s[1]))
	}
	return strings.Join(parts, " and ")
}

// pointerCol finds the balloon's 'v' in the RENDERED frame, rather than
// recomputing where it should be from the same arithmetic that placed it.
func pointerCol(c *canvas.Canvas, row int) int {
	if row < 0 || row >= c.H {
		return -1
	}
	for x := 0; x < c.W; x++ {
		if r, _, _ := c.ResolveAt(x, row, prof); r == 'v' {
			return x
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// 3. THE APPROACH
// ---------------------------------------------------------------------------

type beat struct {
	p      *pose
	dur    float64
	bubble string
	note   string
}

func approachBeats(useMid bool) []beat {
	bs := []beat{
		{workingPose(), 4 * stepDur, "", "working -- 12x7, no balloon"},
		{shippedPose(), stepDur, "allow Bash?", "t=0: the ask. Balloon up, 12x7"},
	}
	if useMid {
		bs = append(bs, beat{midPose(), stepDur, "allow Bash?", "+1 step: 14x9"})
	}
	bs = append(bs, beat{nearPose(), 6 * stepDur, "allow Bash?", "+1 step: 16x9, held"})
	return bs
}

func sectionApproach(sh *scape.Shore, w, h int, t float64, act scape.Activity,
	litter int, seed int64, still, useMid bool, isTTY bool) {

	beats := approachBeats(useMid)
	loop := 0.0
	for _, b := range beats {
		loop += b.dur
	}
	rule("3. THE APPROACH, on the pace's own clock.")
	fmt.Printf("   One step is stepDur = %.2f s, the constant pace.go already uses for a cell of\n"+
		"   the walk. The ask does NOT wait for the walk: the balloon goes up at t=0 and the\n"+
		"   crab comes up underneath it -- an ask stays open a median 130 s, so nothing here\n"+
		"   sits in front of the notification.\n", stepDur)
	for i, b := range beats {
		fmt.Printf("     %d. %-34s %.2f s\n", i+1, b.note, b.dur)
	}
	fmt.Printf("   one pass: %.2f s\n\n", loop)

	if still || !isTTY {
		fmt.Printf("   \x1b[2m-still: the rungs as stills, TOP-aligned, which is how they are anchored --\n" +
			"   same top row, the body growing down into the rows under its feet.\x1b[0m\n\n")
		var ps []*pose
		at := 0.0
		for _, b := range beats[1:] {
			q := *b.p
			q.name = fmt.Sprintf("at t+%.2f s", at)
			ps = append(ps, &q)
			at += b.dur
		}
		bg := term.RGB{R: 44, G: 38, B: 33}
		lbl, art := poseStrip(ps, bg, t, false)
		fmt.Println(lbl)
		fmt.Println(art)
		return
	}

	fmt.Printf("   \x1b[2mplaying three passes, then the near pose is left on screen. ctrl-C is safe.\x1b[0m\n\n")
	playApproach(sh, w, h, t, act, litter, seed, beats, 3)
}

func playApproach(sh *scape.Shore, w, h int, t0 float64, act scape.Activity,
	litter int, seed int64, beats []beat, passes int) {

	restore := func() { fmt.Print("\x1b[?25h") }
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		restore()
		fmt.Print("\n")
		os.Exit(0)
	}()
	defer restore()
	defer signal.Stop(sig)
	fmt.Print("\x1b[?25l")

	var total float64
	for _, b := range beats {
		total += b.dur
	}
	pick := func(e float64) *beat {
		e = math.Mod(e, total)
		for i := range beats {
			if e < beats[i].dur {
				return &beats[i]
			}
			e -= beats[i].dur
		}
		return &beats[len(beats)-1]
	}

	const fps = 20.0
	start := time.Now()
	run := total * float64(passes)
	first := true
	t := t0
	last := start
	for {
		now := time.Now()
		e := now.Sub(start).Seconds()
		if e > run {
			break
		}
		t += now.Sub(last).Seconds()
		last = now
		b := pick(e)
		c, _, _ := buildFrame(sh, frameSpec{
			w: w, h: h, t: t, act: act, p: b.p,
			bubble: b.bubble, litter: litter, seed: seed, tail: sampleTail(),
		})
		if !first {
			fmt.Printf("\x1b[%dA\r", h)
		}
		first = false
		fmt.Print(c.Render(prof) + "\n")
		time.Sleep(time.Duration(float64(time.Second) / fps))
	}
	// Leave the near pose up, which is the thing being ruled on.
	c, _, _ := buildFrame(sh, frameSpec{
		w: w, h: h, t: t, act: act, p: beats[len(beats)-1].p,
		bubble: beats[len(beats)-1].bubble, litter: litter, seed: seed, tail: sampleTail(),
	})
	fmt.Printf("\x1b[%dA\r", h)
	fmt.Print(c.Render(prof) + "\n")
}

// ---------------------------------------------------------------------------
// 4. THE EYE
// ---------------------------------------------------------------------------

func sectionEye(sh *scape.Shore, w, h int, t float64, act scape.Activity, seed int64) {
	rule("4. THE EYE, both ways, at both sizes.")

	// The sea's own colour, measured off the real frame rather than taken from
	// the palette: open water between the horizon and two rows above the
	// waterline, which is what sits behind a companion's head when the tide is
	// in and what produced his "two blue holes" report.
	c, _, _ := buildFrame(sh, frameSpec{
		w: w, h: h, t: t, act: act, p: shippedPose(), litter: 0, seed: seed,
	})
	hz := int(float64(h)*0.42) + 1
	seaBot := maxInt(hz, sh.SandTop()-2)
	sea, sN, sTot := dominantBG(c, 2, hz, w-3, seaBot)
	fmt.Printf("   Ground is rgb(%d,%d,%d), index %d: the commonest background the terminal paints\n"+
		"   across the OPEN SEA of the frame above, rows %d..%d (%d of %d cells). That is the\n"+
		"   colour that shows through a hole in a face.\n\n",
		sea.R, sea.G, sea.B, sea.Index256Keeping(), hz, seaBot, sN, sTot)

	// The shipped vocabulary at one cell, and the bitmap vocabulary at 2x2.
	glyphs := []struct {
		g    rune
		name string
	}{{'O', "alert 'O'"}, {'o', "open 'o'"}, {'-', "shut '-'"}, {'^', "done '^'"}}

	var row1, row2 []*pose
	for _, g := range glyphs {
		a := replicaShipped()
		a.forceGlyph = g.g
		a.name = g.name
		row1 = append(row1, a)
		b := nearPose()
		b.forceGlyph = g.g
		b.name = g.name
		b.xoff = 0
		row2 = append(row2, b)
	}
	fmt.Println("   \x1b[1mSHIPPED: one character in one cell, no ground (EyeFillNone -- his locked ruling)\x1b[0m")
	lbl, art := poseStrip(row1, sea, t, true)
	fmt.Println(lbl)
	fmt.Println(art)
	fmt.Println("\n   \x1b[1mNEAR: a 2x2 bitmap, PlotOn with the coat as ground. Alert is a RING -- the\n" +
		"   hole is what says wide open.\x1b[0m")
	lbl, art = poseStrip(row2, sea, t, true)
	fmt.Println(lbl)
	fmt.Println(art)

	// The proof, measured off the rendered cells.
	fmt.Println("\n   \x1b[1mWHAT THE HOLE ACTUALLY CAME OUT\x1b[0m -- read back with canvas.ResolveAt from")
	fmt.Println("   a frame painted on that sea colour, not from what the source says it does.")
	fmt.Printf("      sea as the terminal paints it   rgb(%3d,%3d,%3d)  index %d\n",
		quant(sea).R, quant(sea).G, quant(sea).B, sea.Index256Keeping())
	coat := companion.CrabCoat
	fmt.Printf("      coat as the terminal paints it  rgb(%3d,%3d,%3d)  index %d\n",
		quant(coat).R, quant(coat).G, quant(coat).B, coat.Index256Keeping())
	fmt.Println()

	type probe struct {
		name   string
		p      *pose
		ground bool
	}
	for _, pr := range []probe{
		{"near 16x9, ring with PlotOn (coat ground)", nearPose(), true},
		{"near 16x9, ring with Plot   (no ground)  ", nearPose(), false},
		{"shipped 12x7, 'O' with Plot (no ground)  ", replicaShipped(), false},
	} {
		pr.p.ground = pr.ground
		pr.p.xoff = 0
		pr.p.forceGlyph = 'O'
		pc := flatPanel(24, 13, sea)
		pr.p.draw(pc.Near(), 3, 2, 0.0)
		sp := pr.p.eyeSpansAt(3)
		x, y := sp[0][0], pr.p.eyeRowAt(2)
		r, _, bg := pc.ResolveAt(x, y, prof)
		what := "\x1b[31mthe SEA -- a window, not a socket\x1b[0m"
		switch {
		case bg == quant(coat):
			what = "the COAT"
		case bg == quant(sea):
			what = "\x1b[31mthe SEA -- a window, not a socket\x1b[0m"
		default:
			what = "neither"
		}
		fmt.Printf("      %s  cell (%d,%d) glyph %q  cell background rgb(%3d,%3d,%3d) = %s\n",
			pr.name, x, y, string(r), bg.R, bg.G, bg.B, what)
	}
	fmt.Println("\n   \x1b[2mThe one-cell glyph has no interior, so its 'hole' is the whole cell and his\n" +
		"   ruling stands. A RING does have an interior, and without PlotOn that interior is\n" +
		"   a hole in the middle of the eye -- which is the 'two blue holes' defect again.\x1b[0m")
}

func quant(c term.RGB) term.RGB { return prof.Quantise(c, false) }

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------

func rule(title string) {
	fmt.Printf("\n\x1b[1m%s\x1b[0m\n%s\n", title, strings.Repeat("-", min(len(title), 78)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

// windowSize is live.go's termSize, plus whether anything answered. Asking the
// controlling terminal directly rather than any particular stream, so a
// redirected stdout does not blind us -- and reporting the failure, so the
// animation can stand down in a pipe instead of writing cursor moves into a file.
func windowSize() (w, h int, ok bool) {
	fds := []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()}
	if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
		defer f.Close()
		fds = append(fds, f.Fd())
	}
	tty := isTerminal(os.Stdout.Fd())
	for _, fd := range fds {
		if c, r, got := winSize(fd); got {
			return c, r, tty
		}
	}
	return 143, 27, false
}

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func winSize(fd uintptr) (cols, rows int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.cols < 8 || ws.rows < 4 {
		return 0, 0, false
	}
	return int(ws.cols), int(ws.rows), true
}

func isTerminal(fd uintptr) bool {
	_, _, ok := winSize(fd)
	return ok
}
