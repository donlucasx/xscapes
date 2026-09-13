package main

import (
	"fmt"
	"math"
	"sort"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// starfall is the study behind his ruling of 2026-09-10 -- "ok let the shooting
// star be the star arriving". Nothing here ships. It renders candidate
// arrivals on REAL composed frames so he can rule on pictures instead of on a
// description, and prints beside every picture the numbers the picture cannot
// carry.
//
// Three rules this file exists to obey, all of them paid for:
//
//  1. It never replicates the product. The home cell an arriving star is
//     heading for is found by DIFFING two rendered frames, not by copying
//     starPlaces; the ink at every path cell comes from scape.StarInkAt, not
//     from a constant. The s28 study replicated the star's alpha as 0.85, a
//     floor shore.go had already deleted, and rendered 25 of 32 stars wrong.
//  2. It measures the frame that SHIPS. Every count is read back with
//     ResolveAt AFTER drawScene, because drawScene draws the readout and an
//     OPAQUE ask balloon into the same near layer the stars live in, and a
//     knock fires in the same instant as two thirds of real arrivals.
//  3. One warmed Shore per frame, never a fresh one per frame: a fresh Shore
//     has no wave clock and no tide, and that trap has voided two instruments
//     in this project already.
type starPt struct{ X, Y int }

// starMark is one cell of an arrival: the head, or a cell of its trail.
type starMark struct {
	X, Y int
	R    rune
	Step int // 0 is the head, 1.. walk back up the track
	// Lift is the arrival flare: 0 is the star's own settled ink, 1 is white.
	// It is walked with term.Ramp rather than by rounding each step, because
	// rounding a fade is exactly what banded the sky in session 15, and a
	// flare is the same problem one cell wide.
	Lift float64
}

// markReport is what the mark actually became once the terminal had it.
type markReport struct {
	X, Y      int
	Step      int
	Rune      rune
	Ink       term.RGB
	Alpha     float64
	Rung      int
	FullAlpha bool
	Exhausted bool
	Contrast  float64 // rendered luma over rendered ground, scape's own measure
	BGShift   float64 // |delta luma| of this cell's ground against the same frame without the mark
	Eaten     bool    // ResolveAt did not return the mark's own rune
}

// starScene is one frame's worth of everything the product needs, plus the
// marks the study is asking it to carry.
type starScene struct {
	W, H        int
	Seed        int64
	Tod, Ctx    float64
	Done, Total int
	Mirror      bool
	Bubble      string
	Ask         bool
	T           float64
	Mag         float64 // the arriving star's own magnitude
	Marks       []starMark
}

// starfallStage builds the scene exactly as the product does, up to but not
// including the marks and drawScene. The warm-up is mockup.go's idiom: ONE
// Shore, thirty steps, so the sea and the tide are where a running session
// would have them.
func starfallStage(sc starScene) (*canvas.Canvas, *scape.Shore, *companion.Cat, layout, reduce.State) {
	c := canvas.New(sc.W, sc.H, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(sc.Seed, false)
	cat := companion.New(companion.DefaultName)
	cat.FaceLeft(sc.Mirror)
	ccw, _ := cat.Size()
	lay := compose(sc.W, ccw, sc.Mirror)
	sh.MoonX = lay.MoonX

	st := reduce.State{
		Act: scape.Activity{
			Level: 0.5, Working: true,
			ContextUsed: sc.Ctx, TimeOfDay: sc.Tod,
			TodoDone: sc.Done, TodoTotal: sc.Total,
		},
		Pose: companion.Working,
	}
	if sc.Bubble != "" {
		st.Bubble, st.BubbleAsk = sc.Bubble, sc.Ask
		if sc.Ask {
			st.Pose = companion.NeedsYou
		} else {
			st.Pose = companion.Done
		}
	}
	for i := 0; i < 30; i++ {
		sh.Update(c, sc.T-1.5+float64(i)/20, st.Act)
	}
	sh.Update(c, sc.T, st.Act)
	return c, sh, cat, lay, st
}

// starfallRender composites one frame with its marks and reports what became
// of each of them. The marks are plotted BEFORE drawScene because that is
// where todoStars plots, and because the readout and the balloon must be
// allowed to win the cell exactly as they would in the product.
func starfallRender(sc starScene) (*canvas.Canvas, []markReport) {
	c, sh, cat, lay, st := starfallStage(sc)

	// The same frame WITHOUT the marks, so each mark's effect on the sky
	// behind it is measured rather than assumed. A split-cell gradient row
	// collapses when a glyph lands in it, and that is invisible to any
	// instrument that only looks at the glyph.
	var bare *canvas.Canvas
	if len(sc.Marks) > 0 {
		bare, _, _, _, _ = starfallStage(sc)
	}

	reps := make([]markReport, 0, len(sc.Marks))
	for _, m := range sc.Marks {
		if m.X < 0 || m.X >= c.W || m.Y < 0 || m.Y >= c.H {
			continue
		}
		ink := sh.StarInkAt(c, m.X, m.Y, sc.Mag)
		c.Near().Plot(m.X, m.Y, m.R, ink.Ink, ink.Alpha)
		if m.Lift > 0 {
			// The ramp has to run between the RENDERED endpoints, not the raw
			// ones: the glyph path saturates chroma before quantising, so a
			// ramp aimed at the palette tone lands on a grey. Read the settled
			// end off the frame we just drew instead of deriving it.
			_, settled, _ := c.ResolveAt(m.X, m.Y, term.Profile256)
			r := term.NewRamp(term.RGB{R: 255, G: 255, B: 255}, settled)
			c.Near().Plot(m.X, m.Y, m.R, r.Tone(1-m.Lift), 1)
			ink.Ink, ink.Alpha = r.Tone(1-m.Lift), 1
		}
		rep := markReport{
			X: m.X, Y: m.Y, Step: m.Step, Rune: m.R,
			Ink: ink.Ink, Alpha: ink.Alpha, Rung: ink.Rung,
			FullAlpha: ink.FullAlpha, Exhausted: ink.Exhausted, Contrast: ink.Contrast,
		}
		if bare != nil {
			_, _, g0 := bare.ResolveAt(m.X, m.Y, term.Profile256)
			rep.BGShift = math.Abs(starLuma(g0))
		}
		reps = append(reps, rep)
	}

	_, chh := cat.Size()
	drawScene(c, sh, cat, lay, st, sc.T, sc.Seed, c.H-2-chh)

	// Read every mark BACK off the composed frame. This is the only step that
	// can say a balloon ate one.
	for i := range reps {
		r, _, g := c.ResolveAt(reps[i].X, reps[i].Y, term.Profile256)
		reps[i].Eaten = r != reps[i].Rune
		if bare != nil {
			_, _, g0 := bare.ResolveAt(reps[i].X, reps[i].Y, term.Profile256)
			reps[i].BGShift = math.Abs(starLuma(g) - starLuma(g0))
		}
	}
	return c, reps
}

func starLuma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
}

// starCellsIn reads the constellation off a composed frame. '*' belongs to the
// constellation alone -- shore.go: "a channel that shares a glyph with the
// scenery is not a channel" -- so every '*' on screen is one unit of work.
func starCellsIn(c *canvas.Canvas) []starPt {
	var out []starPt
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			if r, _, _ := c.ResolveAt(x, y, term.Profile256); r == '*' {
				out = append(out, starPt{x, y})
			}
		}
	}
	return out
}

// starfallArrival finds the cell the next star is going to land on, by
// rendering the scene one star short and one star long and taking the
// difference. It is the only instrument here that cannot drift from what
// ships, because it asks the product rather than re-deriving it.
//
// It also returns the stars ALREADY LIT, which is the set a flight path has to
// keep away from.
func starfallArrival(sc starScene) (home starPt, lit []starPt, ok bool) {
	before := sc
	before.Done, before.Marks = sc.Done-1, nil
	bc, _ := starfallRender(before)
	lit = starCellsIn(bc)

	after := sc
	after.Marks = nil
	ac, _ := starfallRender(after)

	was := make(map[starPt]bool, len(lit))
	for _, p := range lit {
		was[p] = true
	}
	var fresh []starPt
	for _, p := range starCellsIn(ac) {
		if !was[p] {
			fresh = append(fresh, p)
		}
	}
	if len(fresh) != 1 {
		return starPt{}, lit, false
	}
	return fresh[0], lit, true
}

// forbiddenCache keys the measured no-fly footprint by the geometry it was
// measured at. It is measured, never mirrored: readoutGround is itself a
// mirror of drawReadout, and no guard inside package scape can see the ask
// balloon at all, because drawScene runs after Shore.Update and the balloon is
// an OPAQUE sprite into the same near layer.
var forbiddenCache = map[string]map[starPt]bool{}

func starfallForbidden(sc starScene) map[starPt]bool {
	key := fmt.Sprintf("%dx%d/%d/%.3f/%v", sc.W, sc.H, sc.Seed, sc.Tod, sc.Mirror)
	if f, ok := forbiddenCache[key]; ok {
		return f
	}
	out := map[starPt]bool{}
	rows := skyRowsOf(sc.H) + 2
	for step := 0; step <= 20; step++ {
		for _, b := range []struct {
			text string
			ask  bool
		}{{"", false}, {"allow Bash?", true}, {"tests passed", false}} {
			s := sc
			s.Ctx, s.Bubble, s.Ask, s.Marks = float64(step)/20, b.text, b.ask, nil
			c, sh, cat, lay, st := starfallStage(s)
			was := skySnapshot(c, rows)
			_, chh := cat.Size()
			drawScene(c, sh, cat, lay, st, s.T, s.Seed, c.H-2-chh)
			now := skySnapshot(c, rows)
			for k, v := range now {
				if was[k] != v {
					out[k] = true
				}
			}
		}
	}
	forbiddenCache[key] = out
	return out
}

func skyRowsOf(h int) int { return int(float64(h) * 0.42) }

func skySnapshot(c *canvas.Canvas, rows int) map[starPt]string {
	out := map[starPt]string{}
	if rows > c.H {
		rows = c.H
	}
	for y := 0; y < rows; y++ {
		for x := 0; x < c.W; x++ {
			r, fg, bg := c.ResolveAt(x, y, term.Profile256)
			out[starPt{x, y}] = fmt.Sprintf("%c|%d|%d", r, fg.Index256(), bg.Index256())
		}
	}
	return out
}

// trackCfg is every number the flight path is made of. Each one is a slider on
// the page unless it is marked measured there, because a number nobody has
// justified should be looked at rather than asserted.
type trackCfg struct {
	N        int     // columns travelled per row of drop
	Rmax     int     // longest track, in columns
	DyMax    int     // deepest drop, in rows
	TrackMin int     // shorter than this and the star arrives in place
	Sep      float64 // how close the head may pass a lit star, in screen units
	NoGuards bool    // the mutation control: walk anyway, and show what happens
}

// starfallTrack walks OUTWARD from the home cell and stops at the first cell
// the sky will not give it. The track is then flown in reverse, so the star
// falls INTO its own place.
//
// Walking outward rather than searching for a launch point is the whole design.
// Everything it trims is the FAR end, so the landing is never what gets cut,
// and it cannot construct a launch that is off the canvas -- the failure that
// broke between 8.7% and 33.6% of the drafted arrivals. That class of bug is
// not fixed here; it is unreachable.
//
// The side is chosen AWAY FROM THE MOON. starCells already keeps every home
// cell out of the disc's column corridor, so a walk that only ever moves
// further from moonX can never enter it: the disc guard becomes dead code
// rather than a guard that has to fire.
func starfallTrack(sc starScene, home starPt, lit []starPt, forbidden map[starPt]bool,
	sh *scape.Shore, cfg trackCfg) ([]starPt, string) {

	moonCol := int(float64(sc.W) * sh.MoonX)
	try := func(dir int) ([]starPt, string) {
		cells := []starPt{home}
		x, y, dy := home.X, home.Y, 0
		for k := 1; k <= cfg.Rmax; k++ {
			x += dir
			if cfg.N > 0 && k%cfg.N == 0 && dy < cfg.DyMax {
				y--
				dy++
			}
			p := starPt{x, y}
			if !cfg.NoGuards {
				switch {
				case x < 0 || x >= sc.W || y < 0:
					return cells, "canvas edge"
				case sh.DiscCovers(x, y):
					return cells, "the disc"
				case forbidden[p]:
					return cells, "readout or balloon"
				case onLitStar(p, lit):
					return cells, "a lit star"
				case nearLitStar(p, lit, cfg.Sep):
					return cells, "separation bar"
				}
			} else if x < 0 || x >= sc.W || y < 0 {
				// Even with the guards off the canvas is still the canvas.
				return cells, "canvas edge"
			}
			cells = append(cells, p)
		}
		return cells, "cap"
	}

	// Away from the moon first.
	dir := 1
	if home.X < moonCol {
		dir = -1
	}
	cells, why := try(dir)
	if len(cells) < cfg.TrackMin {
		if alt, why2 := try(-dir); len(alt) > len(cells) {
			cells, why = alt, why2+" (flipped)"
		}
	}
	if len(cells) < cfg.TrackMin {
		return []starPt{home}, "in place (" + why + ")"
	}
	// Flown in reverse: launch first, home last.
	out := make([]starPt, len(cells))
	for i, p := range cells {
		out[len(cells)-1-i] = p
	}
	return out, why
}

func onLitStar(p starPt, lit []starPt) bool {
	for _, q := range lit {
		if q == p {
			return true
		}
	}
	return false
}

// nearLitStar uses SCREEN units: a terminal cell is about twice as tall as it
// is wide, so two marks side by side merge and two one row apart do not.
// Squared, the way starPlaces compares, because math.Hypot measured 175-481 us
// a frame here -- up to 45% of a whole frame's budget.
func nearLitStar(p starPt, lit []starPt, sep float64) bool {
	if sep <= 0 {
		return false
	}
	bar := sep * sep
	for _, q := range lit {
		dx := float64(p.X - q.X)
		dy := 2 * float64(p.Y-q.Y)
		if dx*dx+dy*dy < bar {
			return true
		}
	}
	return false
}

// starfallEase is the curve. k of 1 is linear; above that it decelerates.
func starfallEase(p, k float64) float64 {
	if p <= 0 {
		return 0
	}
	if p >= 1 {
		return 1
	}
	if k == 1 {
		return p
	}
	return 1 - math.Pow(1-p, k)
}

// starfallHead is which cell of the track the head occupies at phase p.
func starfallHead(track []starPt, p, k float64) int {
	if len(track) < 2 {
		return 0
	}
	return int(math.Round(starfallEase(p, k) * float64(len(track)-1)))
}

// starfallOffsets is the per-frame sequence of cells-from-home, which is what
// decides whether the head skips a cell. The criterion is: never skip;
// repeats are allowed. A skipped cell is a teleport at 12 fps.
func starfallOffsets(track []starPt, dur, fps, k float64) ([]int, string) {
	n := int(math.Round(dur * fps))
	if n < 1 {
		n = 1
	}
	var out []int
	for i := 0; i <= n; i++ {
		p := float64(i) / fps / dur
		if p > 1 {
			p = 1
		}
		out = append(out, len(track)-1-starfallHead(track, p, k))
	}
	verdict, rep := "clean", false
	for i := 1; i < len(out); i++ {
		switch d := out[i-1] - out[i]; {
		case d > 1:
			verdict = "SKIPS A CELL"
		case d == 0:
			rep = true
		}
	}
	if verdict == "clean" && rep {
		verdict = "clean (repeats a cell)"
	}
	return out, verdict
}

// starfallMarks is the head, plus however much tail it has been given, at one
// phase of one track.
func starfallMarks(track []starPt, p, k float64, tail []rune) []starMark {
	if len(track) == 0 {
		return nil
	}
	i := starfallHead(track, p, k)
	out := []starMark{{X: track[i].X, Y: track[i].Y, R: '*', Step: 0}}
	for s, r := range tail {
		j := i - 1 - s
		if j < 0 {
			break
		}
		out = append(out, starMark{X: track[j].X, Y: track[j].Y, R: r, Step: s + 1})
	}
	return out
}

// starMinSep is the closest any two stars come on a composed frame, in screen
// units. The settled sky's closest pair anywhere measures exactly 2.00, so the
// bar has no headroom at all and a path cell that crowds a star is visible.
func starMinSep(ps []starPt) float64 {
	best := math.MaxFloat64
	for i := range ps {
		for j := i + 1; j < len(ps); j++ {
			dx := float64(ps[i].X - ps[j].X)
			dy := 2 * float64(ps[i].Y-ps[j].Y)
			if d := math.Hypot(dx, dy); d < best {
				best = d
			}
		}
	}
	if best == math.MaxFloat64 {
		// Fewer than two stars: there is no pair, which is not the same thing
		// as a pair at zero. Reporting 0 here read as a merge in the first
		// sweep and was the instrument, not the sky.
		return -1
	}
	return best
}

// starAngles reports the track's angle two ways, because they disagree and the
// disagreement is the point: the scene's own aspect constant is 2.0, and his
// Terminal.app cell measures 14.000 x 30.0 px, which is 2.143. Every angle
// quoted in cells is about 7% shallower than it will look on his glass.
func starAngles(n int) (scene, glass float64) {
	if n <= 0 {
		return 0, 0
	}
	scene = math.Atan2(2.0, float64(n)) * 180 / math.Pi
	glass = math.Atan2(30.0, 14.0*float64(n)) * 180 / math.Pi
	return
}

func starSortPts(ps []starPt) {
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].Y != ps[j].Y {
			return ps[i].Y < ps[j].Y
		}
		return ps[i].X < ps[j].X
	})
}
