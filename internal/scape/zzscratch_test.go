package scape

import (
	"fmt"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/term"
)

// Throwaway probe: not a real test. Renders a frame at 80x24 and inspects the
// moon-disc / ramp-sky boundary on 256, printing indices so we can see
// whether the moon's SetBG cells produce a different quantisation than the
// surrounding SetBGRamp sky cells.
func TestZZMoonSeam(t *testing.T) {
	s := NewShore(1, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 0.3, ContextUsed: 0.1, TimeOfDay: 0.02, TodoDone: 0, TodoTotal: 0}
	s.Update(c, 1.0, act)
	mx, my := s.MoonPos()
	fmt.Printf("moon at %d,%d\n", mx, my)

	// Print a grid of 256-index for background around the moon.
	for y := my - 5; y <= my+5; y++ {
		if y < 0 || y >= 24 {
			continue
		}
		row := ""
		for x := mx - 9; x <= mx+9; x++ {
			if x < 0 || x >= 80 {
				continue
			}
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			row += fmt.Sprintf("%4d", bg.Index256Keeping())
		}
		fmt.Println(row)
	}

	// Also dump a plain sky row far from the moon, for comparison of step size.
	fmt.Println("far-from-moon sky row (y=1):")
	row := ""
	for x := 0; x < 80; x += 4 {
		_, _, bg := c.ResolveAt(x, 1, term.Profile256)
		row += fmt.Sprintf("%4d", bg.Index256Keeping())
	}
	fmt.Println(row)
}

// TestZZSeaSkyDegenerate probes very short panes (40x12, and shorter) for
// crashes / degenerate ramp geometry.
func TestZZSeaSkyDegenerate(t *testing.T) {
	sizes := [][2]int{{40, 12}, {80, 6}, {8, 6}, {40, 7}, {80, 24}}
	for _, sz := range sizes {
		s := NewShore(2, false)
		c := canvas.New(sz[0], sz[1], canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		act := Activity{Working: true, Level: 0.9, ContextUsed: 0.5, TimeOfDay: 0.6}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic at %dx%d: %v", sz[0], sz[1], r)
				}
			}()
			for frame := 0; frame < 3; frame++ {
				s.Update(c, float64(frame), act)
			}
		}()
		fmt.Printf("%dx%d OK\n", sz[0], sz[1])
	}
}

// TestZZResizeMidFrame probes a resize between Update calls.
func TestZZResizeMidFrame(t *testing.T) {
	s := NewShore(3, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 0.5, TimeOfDay: 0.3}
	s.Update(c, 1.0, act)
	c.Resize(40, 12)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic on resize: %v", r)
		}
	}()
	s.Update(c, 1.1, act)
	fmt.Println("resize mid-frame OK")
}

// TestZZTruecolorVsPath checks whether a truecolor render differs from the
// 256 ramp render (it should be much smoother / not snapped to the same path
// nodes) and whether disabling term.Ramps changes the 256 output.
func TestZZTruecolorVsPath(t *testing.T) {
	s := NewShore(4, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 0.2, TimeOfDay: 0.05}
	s.Update(c, 1.0, act)

	fmt.Println("truecolor sky column x=40:")
	for y := 0; y < 10; y++ {
		_, _, bg := c.ResolveAt(40, y, term.ProfileTrueColor)
		fmt.Printf("y=%d rgb=%v\n", y, bg)
	}

	fmt.Println("256 sky column x=40 (ramps on):")
	for y := 0; y < 10; y++ {
		_, _, bg := c.ResolveAt(40, y, term.Profile256)
		fmt.Printf("y=%d idx=%d rgb=%v\n", y, bg.Index256Keeping(), bg)
	}

	was := term.Ramps
	term.Ramps = false
	defer func() { term.Ramps = was }()
	fmt.Println("256 sky column x=40 (ramps FORCED off, same BG data):")
	for y := 0; y < 10; y++ {
		_, _, bg := c.ResolveAt(40, y, term.Profile256)
		fmt.Printf("y=%d idx=%d rgb=%v\n", y, bg.Index256Keeping(), bg)
	}
}

// TestZZIndexFloor scans a full day for any background or foreground index
// below 16, across both truecolor->256 quantisation and the ramp path.
func TestZZIndexFloor(t *testing.T) {
	s := NewShore(5, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	bad := 0
	for h := 0; h < 48; h++ {
		act := Activity{Working: true, Level: 0.7, ContextUsed: 0.3, TimeOfDay: float64(h) / 48.0, TodoDone: 2, TodoTotal: 5}
		s.Update(c, float64(h), act)
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				_, fg, bg := c.ResolveAt(x, y, term.Profile256)
				if fg.Index256Keeping() < 16 {
					bad++
					t.Errorf("fg idx<16 at h=%d (%d,%d): %d", h, x, y, fg.Index256Keeping())
				}
				if bg.Index256Keeping() < 16 {
					bad++
					t.Errorf("bg idx<16 at h=%d (%d,%d): %d", h, x, y, bg.Index256Keeping())
				}
			}
		}
	}
	fmt.Printf("bad cells: %d\n", bad)
}

// TestZZMoonSeamCompareRampsOff runs the same moon frame with term.Ramps
// forced off, to see whether the earthshine "off" cell next to the disc is a
// pre-existing artifact of the per-cell quantiser or something the ramp
// switch specifically introduces.
func TestZZMoonSeamCompareRampsOff(t *testing.T) {
	was := term.Ramps
	term.Ramps = false
	defer func() { term.Ramps = was }()

	s := NewShore(1, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 0.3, ContextUsed: 0.1, TimeOfDay: 0.02}
	s.Update(c, 1.0, act)
	mx, my := s.MoonPos()
	fmt.Printf("[ramps off] moon at %d,%d\n", mx, my)
	for y := my - 2; y <= my+2; y++ {
		if y < 0 || y >= 24 {
			continue
		}
		row := ""
		for x := mx - 9; x <= mx+9; x++ {
			if x < 0 || x >= 80 {
				continue
			}
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			row += fmt.Sprintf("%4d", bg.Index256Keeping())
		}
		fmt.Println(row)
	}
}

// TestZZShortPaneSeaGap probes whether a very short, very active pane can
// produce a column where the wave edge sits above hy+1, so the sea case in
// paintBG's switch never fires for that column and the row that should be
// water/waterline instead falls to the sand-colored default branch, right
// next to the sky.
func TestZZShortPaneSeaGap(t *testing.T) {
	s := NewShore(9, false)
	c := canvas.New(80, 6, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 1.0, TimeOfDay: 0.5}
	sandCol := term.RGB{}
	found := false
	for frame := 0; frame < 200; frame++ {
		s.Update(c, float64(frame)*0.3, act)
		sandCol = writeBandColorExport(s)
		for x := 0; x < 80; x++ {
			for y := 0; y < 6; y++ {
				bg := c.BGAt(x, y)
				if bg == sandCol {
					// A cell painted exactly the flat sand tone. Is it above
					// the sand region (i.e. touching sky, not part of the
					// beach)?
					_ = y
				}
			}
		}
	}
	_ = found
	fmt.Println("scan complete; see per-frame dump below")

	// Now do ONE detailed dump at a frame/activity chosen to stress it, and
	// print the full BG-index column per x for a few x, plus the edge array.
	s2 := NewShore(9, false)
	c2 := canvas.New(80, 6, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act2 := Activity{Working: true, Level: 1.0, TimeOfDay: 0.5}
	for frame := 0; frame < 40; frame++ {
		s2.Update(c2, float64(frame)*0.31, act2)
	}
	fmt.Println("hy, edge sample, sy-ish geometry after 40 frames:")
	for x := 0; x < 80; x += 10 {
		fmt.Printf("x=%d edge=%v\n", x, s2.LastEdgeExport()[x])
	}
	for y := 0; y < 6; y++ {
		row := ""
		for x := 0; x < 80; x += 5 {
			_, _, bg := c2.ResolveAt(x, y, term.Profile256)
			row += fmt.Sprintf("%4d", bg.Index256Keeping())
		}
		fmt.Printf("y=%d: %s\n", y, row)
	}
}

func writeBandColorExport(s *Shore) term.RGB { return writeBandColor(s.pal) }

func (s *Shore) LastEdgeExport() []float64 { return s.lastEdge }

// TestZZWaterlineSeam checks continuity between the last SEA ramp row and the
// waterline-mix row directly below it (which is painted the old, non-ramp
// way), at several hours and activity levels, to see whether switching only
// the open-sea rows to the path painter created a step right at the sand's
// edge that wasn't there when the whole region used one quantiser.
func TestZZWaterlineSeam(t *testing.T) {
	s := NewShore(11, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	maxJump := 0
	var worst string
	for h := 0; h < 48; h++ {
		for _, lvl := range []float64{0.0, 0.3, 0.7, 1.0} {
			act := Activity{Working: true, Level: lvl, TimeOfDay: float64(h) / 48.0}
			s.Update(c, float64(h)*3, act)
			edge := s.LastEdgeExport()
			for x := 0; x < 80; x += 3 {
				ex := edge[x]
				lastSeaRow := int(ex - 0.5) // last row satisfying fy < ex-0.5 roughly
				waterRow := lastSeaRow + 1
				if lastSeaRow < 1 || waterRow >= 24 {
					continue
				}
				_, _, bgSea := c.ResolveAt(x, lastSeaRow, term.Profile256)
				_, _, bgWater := c.ResolveAt(x, waterRow, term.Profile256)
				d := int(bgSea.Index256Keeping()) - int(bgWater.Index256Keeping())
				if d < 0 {
					d = -d
				}
				if d > maxJump {
					maxJump = d
					worst = fmt.Sprintf("h=%d lvl=%.1f x=%d seaRow=%d(idx=%d) waterRow=%d(idx=%d)",
						h, lvl, x, lastSeaRow, bgSea.Index256Keeping(), waterRow, bgWater.Index256Keeping())
				}
			}
		}
	}
	fmt.Printf("max sea/waterline index jump: %d\nworst: %s\n", maxJump, worst)
}

// TestZZWaterlineSeamDetail dumps full context for the worst jump found above.
func TestZZWaterlineSeamDetail(t *testing.T) {
	s := NewShore(11, false)
	c := canvas.New(80, 24, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	act := Activity{Working: true, Level: 0.3, TimeOfDay: 14.0 / 48.0}
	s.Update(c, 14.0*3, act)
	edge := s.LastEdgeExport()
	x := 45
	fmt.Printf("edge[%d]=%v pal.SeaNear=%v pal.WetSand=%v pal.SandNear=%v\n",
		x, edge[x], s.SandColor(), s.pal.WetSand, s.pal.SandNear)
	for y := 10; y <= 20; y++ {
		bgTrue := c.BGAt(x, y)
		_, _, bgQ := c.ResolveAt(x, y, term.Profile256)
		fmt.Printf("y=%d trueBG=%v -> idx=%d rgb=%v\n", y, bgTrue, bgQ.Index256Keeping(), bgQ)
	}
}
