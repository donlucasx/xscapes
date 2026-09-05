package scape

import (
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// The quad-edge disc is one clean container: at every phase the disc's
// cells (lit or dark) form ONE connected silhouette, the dark face forms one
// connected region, and the lit face carries ONE tone. His 2026-09-05 note
// on the study: "the 'empty shadow' it leaves as context depletes is
// broken/messy. Should be a clean container underneath."
func TestTheQuadDiscIsOneCleanContainer(t *testing.T) {
	const w, h = 130, 22
	for _, used := range []float64{0.05, 0.30, 0.60, 0.85} {
		for _, tod := range []float64{0.556, 0.931} {
			render := func(edge string) (*canvas.Canvas, *Shore) {
				c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
				sh := NewShore(7, false)
				sh.MoonX = 0.28
				sh.MoonEdge = edge
				sh.Update(c, 2, Activity{Working: true, Level: 0.65, ContextUsed: used, TimeOfDay: tod})
				return c, sh
			}
			c, sh := render("quad")
			control, _ := render("none") // the same frame without the disc
			mx, my := sh.MoonPos()
			rx, ry := sh.MoonExtent()
			// The disc's cells are the ones that differ from the control frame;
			// a sky tone that happens to round to the dark tone is not the disc.
			lit := term.Profile256.Quantise(sh.litTone, false)
			disc, darks := map[cellPt]bool{}, map[cellPt]bool{}
			litTones := map[term.RGB]bool{}
			for dy := -ry - 1; dy <= ry+1; dy++ {
				for dx := -rx - 1; dx <= rx+1; dx++ {
					x, y := mx+dx, my+dy
					if x < 0 || y < 0 || x >= w || y >= h {
						continue
					}
					ch, fg, bg := c.ResolveAt(x, y, term.Profile256)
					ch0, fg0, bg0 := control.ResolveAt(x, y, term.Profile256)
					if ch == ch0 && fg == fg0 && bg == bg0 {
						continue
					}
					has := func(col term.RGB) bool { return bg == col || (ch != ' ' && fg == col) }
					disc[cellPt{x, y}] = true
					if sh.quadDark[[2]int{x, y}] {
						darks[cellPt{x, y}] = true
					}
					if has(lit) {
						if bg == lit {
							litTones[bg] = true
						} else {
							litTones[fg] = true
						}
					}
				}
			}
			if n := regions(disc); n != 1 {
				t.Errorf("used %.0f%% tod %.3f: the disc is %d regions, want one silhouette", used*100, tod, n)
			}
			if n := regions(darks); used > 0.1 && n != 1 {
				var cells []cellPt
				for p := range darks {
					cells = append(cells, cellPt{p.x - mx, p.y - my})
				}
				t.Errorf("used %.0f%% tod %.3f: the dark face is %d regions, want one; dark cells (dx,dy): %v", used*100, tod, n, cells)
			}
			if len(litTones) > 1 {
				t.Errorf("used %.0f%% tod %.3f: the lit face carries %d tones, want one", used*100, tod, len(litTones))
			}
		}
	}
}

type cellPt struct{ x, y int }

// regions counts 8-connected groups of cells.
func regions(cells map[cellPt]bool) int {
	seen := map[cellPt]bool{}
	n := 0
	for p := range cells {
		if seen[p] {
			continue
		}
		n++
		stack := []cellPt{p}
		seen[p] = true
		for len(stack) > 0 {
			q := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					r := cellPt{q.x + dx, q.y + dy}
					if cells[r] && !seen[r] {
						seen[r] = true
						stack = append(stack, r)
					}
				}
			}
		}
	}
	return n
}

// Night is a property of the sky, and the midday zenith is a dark saturated
// blue: it must not count. The halo and the sun's missing shadow both key
// on this.
func TestNightIsNotTheMiddayZenith(t *testing.T) {
	for _, tc := range []struct {
		tod   float64
		night bool
	}{{0.556, false}, {0.479, false}, {0.931, true}, {0.02, true}} {
		sh := NewShore(7, false)
		c := canvas.New(40, 12, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh.Update(c, 2, Activity{TimeOfDay: tc.tod})
		if got := sh.night(); got != tc.night {
			t.Errorf("tod %.3f: night=%v, want %v (zenith %v)", tc.tod, got, tc.night, sh.pal.SkyTop)
		}
	}
}
