package companion

import (
	"testing"

	"github.com/donlucasx/asciiscapes/internal/canvas"
)

// Every whisker must touch the fur.
//
// They used to be placed at fixed cell offsets, which looks symmetrical in the
// source and is not on screen: the head is not centred in its sprite, so at the
// chin row the left edge lands on a half-filled cell while the right edge is
// solid. A fixed offset therefore connects on one side and leaves half a cell
// of daylight on the other.
func TestWhiskersTouchTheFur(t *testing.T) {
	styles := []struct {
		s WhiskerStyle
		n string
	}{
		{WhiskerLowerLong, "lower long"}, {WhiskerUpperLong, "upper long"},
		{WhiskerEven, "even"}, {WhiskerShort, "short"}, {WhiskerSweep, "sweep"},
	}
	for _, mirror := range []bool{false, true} {
		for _, st := range []State{Resting, Working, NeedsYou, Worried} {
			for _, w := range styles {
				c := canvas.New(20, 10, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				cat := NewCat()
				cat.FaceLeft(mirror)
				cat.SetFace(Face{Nose: true, Toes: true, Whiskers: w.s, Ears: EarInnerDark})
				cat.Draw(c.Near(), 3, 1, 3.1, st)

				l := c.Near()
				found := 0
				for y := 0; y < l.H; y++ {
					for x := 0; x < l.W; x++ {
						cell := l.Cells[y*l.W+x]
						if !cell.Set || cell.R != '─' {
							continue
						}
						found++
						// Walk toward the body; a solid fur cell must be
						// reachable without crossing empty space.
						if !touchesFur(l, x, y, -1) && !touchesFur(l, x, y, +1) {
							t.Errorf("%s mirror=%v state=%v: whisker at (%d,%d) is not connected to the fur",
								w.n, mirror, st, x, y)
						}
					}
				}
				if found == 0 {
					t.Errorf("%s mirror=%v state=%v: no whiskers drawn at all", w.n, mirror, st)
				}
			}
		}
	}
}

// touchesFur walks in direction d from x until it hits solid fur (connected) or
// an empty cell (a gap, so not connected).
func touchesFur(l *canvas.Layer, x, y, d int) bool {
	for cx := x + d; cx >= 0 && cx < l.W; cx += d {
		cell := l.Cells[y*l.W+cx]
		switch {
		case cell.Set && cell.R == '█':
			return true
		case cell.Set && cell.R == '─':
			continue // another whisker cell in the same run
		default:
			return false
		}
	}
	return false
}

// The eye row must stay clear: a stroke beside the eyes reads as a brow.
func TestWhiskersStayOffTheEyeRow(t *testing.T) {
	c := canvas.New(20, 10, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	cat := NewCat()
	cat.SetFace(Face{Nose: true, Whiskers: WhiskerLowerLong})
	const top = 1
	cat.Draw(c.Near(), 3, top, 3.1, Working)

	l := c.Near()
	eyeRow := top + 2
	for x := 0; x < l.W; x++ {
		if cell := l.Cells[eyeRow*l.W+x]; cell.Set && cell.R == '─' {
			t.Errorf("a whisker landed on the eye row at x=%d", x)
		}
	}
}
