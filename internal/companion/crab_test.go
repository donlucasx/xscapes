package companion

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// The crab has to fit the box every companion is drawn into. The layout
// reserves 12 by 7 from Size(), and a companion that returns anything else
// moves the sand's right bound and the balloon with it.
func TestTheCrabFillsTheSameBoxAsTheCat(t *testing.T) {
	cw, ch := NewCrab().Size()
	kw, kh := NewCat().Size()
	if cw != kw || ch != kh {
		t.Fatalf("crab is %dx%d cells, cat is %dx%d -- the layout reserves one box for both", cw, ch, kw, kh)
	}
}

// Every state has to actually change the picture. Five states that render the
// same pixels are one state, and the companion is the channel that carries
// "something is broken".
func TestEveryCrabStateLooksDifferent(t *testing.T) {
	seen := map[string]State{}
	for _, st := range []State{Resting, Working, NeedsYou, Done, Worried} {
		c := canvas.New(30, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		crab := NewCrab()
		crab.FaceLeft(true)
		crab.Draw(c.Near(), 8, 4, 0, st)
		key := ""
		for _, cell := range c.Near().Cells {
			if cell.Set {
				key += string(cell.R)
			} else {
				key += "."
			}
		}
		if prev, dup := seen[key]; dup {
			t.Errorf("%v and %v draw the same frame", prev, st)
		}
		seen[key] = st
	}
}

// The crab is drawn in its coat, and the coat reaches the screen. Salmon is an
// exact cube entry precisely so the glyph path's saturate leaves it alone; if
// that ever stops being true this catches it.
func TestTheCrabIsDrawnInItsCoat(t *testing.T) {
	c := canvas.New(30, 14, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	crab := NewCrab()
	crab.FaceLeft(true)
	crab.Draw(c.Near(), 8, 4, 0, Working)
	n := 0
	for _, cell := range c.Near().Cells {
		if cell.Set && cell.FG == CrabCoat {
			n++
		}
	}
	if n < 20 {
		t.Fatalf("only %d cells in the crab's coat; the body is not being painted", n)
	}
}

// Every crablet that gets drawn keeps both eyes. Same guarantee the kittens
// have, and for the same reason: bodies are drawn before any face, so a
// neighbour's sprite seam cannot take an eye cell.
func TestEveryCrabletKeepsBothEyes(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5, 8, 12, 16} {
		for _, w := range []int{60, 80, 100, 120, 152} {
			for _, seed := range []int64{1, 7, 42, 1234} {
				for _, tt := range []float64{0.0, 1.3, 3.0, 7.7} {
					c := canvas.New(w, 30, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
					crab := NewCrab()
					crab.FaceLeft(true)
					cw, ch := crab.Size()
					px, py := w-cw-2, 30-2-ch
					drawn := crab.DrawKittens(c.Near(), c.Mid(), px, py, n, w, 10, 20, tt, seed)
					if drawn == 0 {
						continue
					}
					eyes := 0
					for _, cell := range c.Near().Cells {
						if cell.Set && cell.FG == kittenEye && (cell.R == 'o' || cell.R == '-') {
							eyes++
						}
					}
					if eyes != drawn*2 {
						t.Fatalf("n=%d w=%d seed=%d t=%.1f: %d crablets drawn but %d eyes, want %d",
							n, w, seed, tt, drawn, eyes, drawn*2)
					}
				}
			}
		}
	}
}

// Two crablets in the water must never overlap. Overlapping sprites interleave
// into one broken shape, which is a defect he reported on the kittens and which
// a new animal does not get to rediscover.
//
// Asserted on the PLACEMENT rather than on the painted frame. An earlier
// version of this test measured run lengths in a row and failed on two swimmers
// in NEIGHBOURING lanes whose rims legally touch -- a proxy catching something
// that was never the defect. The cat splits swimmerSpans out for the same
// reason.
func TestCrabSwimmersNeverShareColumns(t *testing.T) {
	kw := len(CrabletSwim[0]) / 2
	for _, n := range []int{2, 4, 8, 16, 24} {
		for _, w := range []int{60, 80, 120, 152} {
			for _, seed := range []int64{1, 7, 42, 1234} {
				var idx []int
				for i := 0; i < n; i++ {
					idx = append(idx, i)
				}
				spans := crabSwimSpans(idx, w, 6, 20, seed)
				for a := range spans {
					for b := a + 1; b < len(spans); b++ {
						if spans[a].lane != spans[b].lane {
							continue
						}
						// Bodies plus a ripple cell either side.
						x0, x1 := spans[a].x-1, spans[a].x+kw
						y0, y1 := spans[b].x-1, spans[b].x+kw
						if x0 <= y1 && y0 <= x1 {
							t.Fatalf("n=%d w=%d seed=%d: swimmers %d and %d share lane %d at x=%d and x=%d",
								n, w, seed, spans[a].i, spans[b].i, spans[a].lane, spans[a].x, spans[b].x)
						}
					}
				}
			}
		}
	}
}

// A leaving crablet sinks as it goes. The whole read of the exit is that the
// shell goes under and the stalks are the last thing above the line, so ink has
// to fall away monotonically as progress runs out.
func TestACrabExitSinksAsItLeaves(t *testing.T) {
	prev := 1 << 30
	for _, p := range []float64{0.0, 0.2, 0.4, 0.6, 0.8, 0.95} {
		c := canvas.New(120, 30, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		crab := NewCrab()
		crab.drawCrabExits(c.Near(), []float64{p}, 100, 120, 6, 20, 2.0, 7)
		ink := 0
		for _, cell := range c.Near().Cells {
			if cell.Set && cell.FG == CrabCoat {
				ink++
			}
		}
		if ink > prev {
			t.Fatalf("at p=%.2f the exit has %d cells of shell, more than the %d before it -- it should only sink", p, ink, prev)
		}
		prev = ink
	}
	if prev == 0 {
		// A zero at the end is right; a zero throughout would mean nothing drew.
		t.Log("gone by 0.95, as intended")
	}
}

// New() is what the CLI and the config file go through, and an unknown name
// must never stop the scape from running.
func TestNewFallsBackRatherThanFailing(t *testing.T) {
	if got := New(NameCat).Name(); got != NameCat {
		t.Errorf("New(cat) is %q", got)
	}
	if got := New(NameCrab).Name(); got != NameCrab {
		t.Errorf("New(crab) is %q", got)
	}
	if got := New("octopus").Name(); got != DefaultName {
		t.Errorf("New(octopus) is %q, want the default %q", got, DefaultName)
	}
}
