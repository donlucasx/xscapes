package main

import (
	"fmt"

	"github.com/donlucasx/xscapes/internal/term"
)

// colorRun emits the frames the ASCII cannot show. A and D live entirely in
// colour: their ASCII renders are byte-identical to the control, which is the
// whole point of them and also why they must be looked at, not read about.
func colorRun() {
	fmt.Println("\n\n==== 9. THE SAME FRAMES IN REAL 256 COLOUR ====")
	white := term.RGB{R: 255, G: 255, B: 255}

	home := placeLive(12)
	newest := home[len(home)-1]
	probe := warm(0, 0, CTX)
	probe.plotStar(newest, probe.pal.Star, Floor)
	_, settled, _ := probe.c.ResolveAt(newest.x, newest.y, term.Profile256)
	ramp := term.NewRamp(white, settled)

	fmt.Printf("\n  CONTROL -- 12 stars, today. The newest is at column %d, row %d.\n", newest.x, newest.y)
	emit(warmWith(12, 32))

	for i, t := range []float64{0.0, 0.4, 1.0} {
		s := warm(0, 0, CTX)
		for _, q := range home[:len(home)-1] {
			s.plotStar(q, s.pal.Star, Floor)
		}
		s.plotStar(newest, ramp.Tone(t), 1.0)
		fmt.Printf("\n  A5 ARRIVAL FLARE, decay t=%.1f (%s) -- one cell, one glyph, three tones.\n",
			t, []string{"peak", "mid", "settled"}[i])
		emit(s)
	}

	s := warm(0, 0, CTX)
	ps := placeLive(26)
	for i, p := range ps {
		if (i+1)%5 == 0 {
			s.plotStar(p, white, 1.0)
		} else {
			s.plotStar(p, s.pal.Star, Floor)
		}
	}
	fmt.Println("\n  C MAGNITUDE at 26 -- every 5th bright. Look for the fives. They are not there.")
	emit(s)

	cool := term.RGB{R: 185, G: 205, B: 255}
	warmc := term.RGB{R: 255, G: 225, B: 180}
	s = warm(0, 0, CTX)
	ps = placeLive(26)
	for i, p := range ps {
		s.plotStar(p, lerpRGB(cool, warmc, float64(i)/float64(maxi(1, len(ps)-1))), Floor)
	}
	fmt.Println("\n  D2 AGE BY HUE at 26 -- oldest cool, newest warm. Six cube entries for 26 stars.")
	emit(s)
}

func warmWith(done, total int) *sky { return warm(done, total, CTX) }

// emit writes the sky rows as real 256-colour ANSI, cell by cell off the
// resolved frame -- the same bytes the scape would send.
func emit(s *sky) {
	var buf []byte
	for y := 0; y <= s.hy; y++ {
		buf = append(buf, "     "...)
		for x := 0; x < W; x++ {
			r, fg, bg := s.c.ResolveAt(x, y, term.Profile256)
			if r == 0 {
				r = ' '
			}
			buf = term.Profile256.AppendBG(buf, bg)
			buf = term.Profile256.AppendFG(buf, fg)
			buf = append(buf, string(r)...)
		}
		buf = append(buf, "\x1b[0m\n"...)
	}
	fmt.Print(string(buf))
}
