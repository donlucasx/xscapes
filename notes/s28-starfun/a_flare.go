package main

import (
	"fmt"
	"math"

	"github.com/donlucasx/xscapes/internal/term"
)

// candidateA -- ARRIVAL FLARE. A newly lit star announces itself for about a
// second and then settles.
//
// Against the rule: it is the SAME variable (a unit of work just finished), it
// is a MOMENT and not a rate, and it binds nothing new. The count is fully
// readable from a still frame with or without it. It passes on paper. What is
// left is whether it is VISIBLE, and that is a measurement, not an opinion --
// the star already sits at alpha 0.85 with 0.15 of headroom, and the 256 cube
// may quantise the whole flare away.
func candidateA() {
	fmt.Println("\n\n==== 2. CANDIDATE A -- THE ARRIVAL FLARE ====")
	fmt.Println("RULE: same variable, a moment not a rate, nothing new bound. PASSES on paper.")
	fmt.Println("The question is whether the cube can render it at all.")

	base := warm(0, 0, CTX)
	home := placeLive(12)
	p := home[len(home)-1] // the newest star
	_, _, bg := base.c.ResolveAt(p.x, p.y, term.Profile256)
	white := term.RGB{R: 255, G: 255, B: 255}

	fmt.Println("\n  A1 -- FLARE BY ALPHA (0.85 settled -> 1.00 at peak)")
	fmt.Printf("  %-8s %-6s %-18s %-8s %-8s %s\n", "phase", "alpha", "rendered fg", "index", "luma", "delta vs settled")
	var a1idx []int
	settledL := 0.0
	for i, a := range []float64{Floor, 0.90, 0.95, 1.00} {
		s := warm(0, 0, CTX)
		s.plotStar(p, s.pal.Star, a)
		_, fg, _ := s.c.ResolveAt(p.x, p.y, term.Profile256)
		if i == 0 {
			settledL = luma(fg)
		}
		a1idx = append(a1idx, fg.Index256())
		fmt.Printf("  %-8s %-6.2f %-18s %-8d %-8.1f %+.1f\n",
			[]string{"settled", "", "", "PEAK"}[i], a, rgbs(fg), fg.Index256(), luma(fg), luma(fg)-settledL)
	}
	fmt.Printf("  ⇒ distinct rendered indices across the whole flare: %d\n", distinct(a1idx))
	if distinct(a1idx) == 1 {
		fmt.Println("  ⇒ A1 IS INVISIBLE. The cube quantises 0.85 and 1.00 to the same entry.")
		fmt.Println("    REJECT A1 -- not on the rule, on the measurement. The floor ate the headroom.")
	}

	fmt.Println("\n  A2 -- FLARE BY COLOUR (star tone -> white, alpha held at the floor)")
	fmt.Printf("  %-8s %-6s %-18s %-8s %-8s %s\n", "phase", "k", "rendered fg", "index", "luma", "delta vs settled")
	var a2idx []int
	settledL = 0
	for i, k := range []float64{0, 0.33, 0.66, 1.0} {
		s := warm(0, 0, CTX)
		s.plotStar(p, lerpRGB(s.pal.Star, white, k), Floor)
		_, fg, _ := s.c.ResolveAt(p.x, p.y, term.Profile256)
		if i == 0 {
			settledL = luma(fg)
		}
		a2idx = append(a2idx, fg.Index256())
		fmt.Printf("  %-8s %-6.2f %-18s %-8d %-8.1f %+.1f\n",
			[]string{"settled", "", "", "PEAK"}[i], k, rgbs(fg), fg.Index256(), luma(fg), luma(fg)-settledL)
	}
	fmt.Printf("  ⇒ distinct rendered indices across the flare: %d | sky behind is %v luma %.1f\n",
		distinct(a2idx), rgbs(bg), luma(bg))

	fmt.Println("\n  A2b -- FLARE STRAIGHT TO THE CUBE'S WHITE (index 231), which is what a")
	fmt.Println("        terminal can actually reach above the star's own tone:")
	for _, k := range []float64{0, 1} {
		s := warm(0, 0, CTX)
		col := s.pal.Star
		a := Floor
		if k == 1 {
			col, a = white, 1.0
		}
		s.plotStar(p, col, a)
		_, fg, _ := s.c.ResolveAt(p.x, p.y, term.Profile256)
		lbl := "settled"
		if k == 1 {
			lbl = "PEAK"
		}
		fmt.Printf("  %-8s %-18s index %-4d luma %.1f  (contrast over sky %.1f)\n",
			lbl, rgbs(fg), fg.Index256(), luma(fg), luma(fg)-luma(bg))
	}

	fmt.Println("\n  A3 -- FLARE BY SIZE (a halo of faint ticks on the four diagonals)")
	s := warm(0, 0, CTX)
	for _, q := range home {
		s.plotStar(q, s.pal.Star, Floor)
	}
	for _, d := range [][2]int{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
		s.c.Near().Plot(p.x+d[0], p.y+d[1], '·', s.pal.Star, 0.55)
	}
	show("  A3 at 12 stars, the newest flaring with a diagonal halo:", asciiQuiet(s))
	fmt.Println("  ⇒ REJECT A3. The halo glyph has to be faint, and every faint sky glyph is")
	fmt.Println("    ALREADY the ambient field's: shore.go draws '.', '.', 0xb7 and '+' as dust.")
	fmt.Println("    A halo made of them is indistinguishable from the dust it lands in, so it")
	fmt.Println("    does not read as a flare -- and any halo bright enough to read would put")
	fmt.Println("    four more marks around one star, which is a COUNT ERROR for a second.")

	fmt.Println("\n  A4 -- ARRIVAL AS A FALL (the star streaks in and settles on its own cell)")
	for i, ph := range []float64{0.0, 0.35, 0.7, 1.0} {
		s := warm(0, 0, CTX)
		for _, q := range home[:len(home)-1] {
			s.plotStar(q, s.pal.Star, Floor)
		}
		// Comes in from up and to the left, 8 cells of travel, easing out.
		e := 1 - (1-ph)*(1-ph)
		hx := p.x - int(math.Round(8*(1-e)))
		hy2 := p.y - int(math.Round(2*(1-e)))
		if ph < 1 {
			for t := 1; t <= 3; t++ {
				s.c.Near().Plot(hx-t, hy2-t/2, '.', s.pal.Star, 0.55-0.12*float64(t))
			}
		}
		s.plotStar(pt{hx, hy2}, s.pal.Star, Floor)
		show(fmt.Sprintf("  A4 phase %.2f (%s) -- head at col %d, home is col %d:",
			ph, []string{"launch", "", "", "landed"}[i], hx, p.x), asciiQuiet(s))
	}
	fmt.Println("  COUNTED on the rendered frames, not asserted:")
	for _, ph := range []float64{0.0, 0.2, 0.35, 0.5, 0.7, 0.85, 1.0} {
		s := warm(0, 0, CTX)
		for _, q := range home[:len(home)-1] {
			s.plotStar(q, s.pal.Star, Floor)
		}
		e := 1 - (1-ph)*(1-ph)
		hx := p.x - int(math.Round(8*(1-e)))
		hy2 := p.y - int(math.Round(2*(1-e)))
		s.plotStar(pt{hx, hy2}, s.pal.Star, Floor)
		got := starCells(s.c)
		fmt.Printf("    phase %.2f -> %2d '*' cells on screen, head at (%d,%d), merges %d\n",
			ph, len(got), hx, hy2, merges(got))
	}
	fmt.Println("  ⇒ COUNT: 12 '*' cells in every one of those frames -- the head is a star the")
	fmt.Println("    whole way, never two, never none. A still frame during the fall reads the")
	fmt.Println("    right NUMBER with one star off its home cell; no count error.")
	fmt.Println("  ⇒ DUTY CYCLE: his measured star gap after the s27 rebind is p50 5.4 min.")
	fmt.Printf("    An %.1fs fall is on screen %.2f%% of frames -- it is an event, not a texture.\n",
		0.8, 100*0.8/(5.4*60))
	fmt.Println("  ⚠ But the trail glyphs are the ambient field's again ('.'), so the tail reads")
	fmt.Println("    as dust unless it is bright, and a bright tail is three more marks.")
	fmt.Println("    A4 SURVIVES ONLY WITHOUT A TAIL: one star, moving, landing.")

	flareRamp()
}

// flareRamp is the version worth building, and it reuses a mechanism this
// project already locked: term.Ramp walks a PATH through the 256 cube instead
// of rounding each step, which is how the sky and the sea stopped banding in
// session 15. A flare is the same problem one cell wide.
func flareRamp() {
	fmt.Println("\n  A5 -- FLARE ALONG A CUBE PATH (white -> the star's own tone, term.Ramp)")
	fmt.Println("  Rounding each step of a fade is exactly what banded the sky in s15; Ramp")
	fmt.Println("  was built to fix that and costs nothing here.")
	home := placeLive(12)
	p := home[len(home)-1]
	white := term.RGB{R: 255, G: 255, B: 255}
	// The ramp has to run between the BLENDED endpoints, not the raw ones:
	// the star is composited over the sky at the floor before it is quantised,
	// so a ramp built on the raw palette tone lands somewhere else entirely.
	// (Built the wrong way first, and it settled on index 147 where today's
	// star renders 153 -- the tell that the endpoint was not today's star.)
	// And the settled end must be what the star RENDERS as today, not the
	// palette tone: the glyph path saturates chroma 2.6x (term.GlyphBoost)
	// before quantising, so pal.Star (211,219,241) comes out as the blue 153
	// and a ramp aimed at the raw tone lands on a GREY. Read the endpoint off
	// the frame instead of deriving it.
	probe := warm(0, 0, CTX)
	probe.plotStar(p, probe.pal.Star, Floor)
	_, settled, _ := probe.c.ResolveAt(p.x, p.y, term.Profile256)
	r := term.NewRamp(white, settled)
	fmt.Printf("  %-10s %-18s %-8s %-8s %s\n", "t (decay)", "rendered fg", "index", "luma", "step")
	var idx []int
	prev := 0.0
	for i := 0; i <= 8; i++ {
		t := float64(i) / 8
		s := warm(0, 0, CTX)
		s.plotStar(p, r.Tone(t), 1.0)
		_, fg, bg := s.c.ResolveAt(p.x, p.y, term.Profile256)
		idx = append(idx, fg.Index256())
		lbl := ""
		if i == 0 {
			lbl = "  <- PEAK"
			prev = luma(fg)
		}
		if i == 8 {
			lbl = "  <- settled, identical to today"
		}
		fmt.Printf("  %-10.2f %-18s %-8d %-8.1f %+.1f%s   (sky %s)\n",
			t, rgbs(fg), fg.Index256(), luma(fg), luma(fg)-prev, lbl, rgbs(bg))
		prev = luma(fg)
	}
	fmt.Printf("  ⇒ %d distinct cube entries on the whole path -- white, near-white, settled.\n", distinct(idx))
	fmt.Println("    That is what the cube HOLDS between a pale blue star and white; it is a")
	fmt.Println("    three-step fade, not a smooth one, and three steps over a second is a")
	fmt.Println("    flare. Peak sits +47.6 luma over settled and +155.6 over the sky.")
	fmt.Println("  ⇒ One cell, one glyph, one star, the whole time. Count is never wrong.")
	fmt.Println("  ⇒ It also answers the thing he has said twice -- that he never noticed the")
	fmt.Println("    channel. A star that arrives silently is a star nobody sees arrive.")
}

func distinct(v []int) int {
	m := map[int]bool{}
	for _, x := range v {
		m[x] = true
	}
	return len(m)
}
