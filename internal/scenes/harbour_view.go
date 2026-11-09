package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// paintView is the study's paintHarbour (rain_views.go) on the layout: the
// city across the water, the port, the traffic, the dock. Columns of the
// 80-column design are scaled (sx); the sky's rows hang from the horizon,
// the water's are spread over the water rows (wy), the dock's hang from its
// edge. The traffic runs on the session clock at its own pace: the world's.
func (h *Harbour) paintView(c *canvas.Canvas, p scape.Palette, l float64, t float64) {
	lay := h.lay
	seed := h.Seed
	sx := lay.sx
	off := lay.hz - 8 // the sky's design rows hang from the horizon
	sy := func(y int) int { return max(0, y+off) }
	haze, far2 := hazeAt(p, l), hazeFar(p, l)
	lf := litFrac(l)
	far, mid, near := c.Far(), c.Mid(), c.Near()
	ground := func(x, y int) rgb { return c.BGAt(x, y) }

	// The far horizon: hills, a row of towers, two stacks and their plume,
	// a mast with a beacon, then the near skyline.
	hills := term.Lerp(far2, p.SkyHorizon, 0.55)
	for x := 0; x < c.W; x++ {
		u := float64(x) / float64(c.W)
		hh := 4.6 - 1.4*math.Sin(u*2.6+0.9) - 0.7*math.Sin(u*7.3+1.7)
		fill(c, x, sy(int(math.Round(hh))), x, lay.hz, hills)
	}
	for _, tw := range [][3]int{{3, 4, 6}, {17, 3, 20}, {36, 4, 40}, {56, 3, 58}, {72, 4, 75}} {
		box(c, sx(tw[0]), sy(tw[1]), sx(tw[0])+tw[2]-tw[0], lay.hz, far2)
	}
	stack := sx(30)
	plot(mid, stack, sy(2), '|', far2, 1)
	plot(mid, stack, sy(3), '|', far2, 1)
	fill(c, stack, sy(4), stack, lay.hz, far2)
	fill(c, stack+3, sy(5), stack+3, lay.hz, far2)
	for k, x := 0, stack+1; x <= stack+8; k, x = k+1, x+2 {
		plot(far, x, sy(1+k/3), ',', lighten(far2, 0.2), 0.45)
	}
	for x := 2; x <= c.W-3; x++ {
		hh := 5 + int(scape.HashF(x/3, 91, seed)*3)
		if x >= sx(10) && x <= sx(10)+6 {
			hh = 7
		}
		fill(c, x, sy(hh), x, lay.hz, far2)
	}
	box(c, sx(22), sy(3), sx(22)+3, lay.hz, haze)
	box(c, sx(44), sy(2), sx(44)+5, lay.hz, haze)
	plot(mid, sx(44)+2, sy(1), '|', haze, 1)
	plot(mid, sx(44)+2, sy(0), '.', cube(175, 0, 0), 1)
	box(c, sx(63), sy(4), sx(63)+3, lay.hz, haze)
	for x := 3; x <= c.W-4; x += 2 {
		if scape.HashF(x, 92, seed) < 0.35*lf {
			plot(far, x, sy(7), '.', warmLight, 0.6)
		}
	}
	// A tanker far out on the horizon line.
	tk := sx(52)
	fill(c, tk, lay.hz, tk+7, lay.hz, darken(far2, 0.25))
	fill(c, tk+6, lay.hz-1, tk+6, lay.hz-1, darken(far2, 0.25))

	// The quay: gantry cranes and the containers under them; the
	// lighthouse at the right end.
	iron := darken(haze, 0.45)
	crane := func(x, top int) {
		top = sy(top)
		fill(c, x, top, x, lay.hz, iron)
		fill(c, x-1, top, x+6, top, iron)
		line(mid, x+6, top+1, x+1, top+3, iron, 0.9)
		plot(mid, x+4, top+1, '|', iron, 0.9)
		plot(mid, x+4, top+2, '|', iron, 0.9)
	}
	crane(sx(6), 4)
	crane(sx(14), 5)
	crane(sx(68), 4)
	boxes := []rgb{cube(95, 0, 0), cube(0, 95, 135), cube(135, 95, 0), cube(0, 95, 95)}
	for k, x := 0, sx(3); x <= sx(22); k, x = k+1, x+2 {
		if scape.HashF(x, 98, seed) < 0.75 {
			fill(c, x, lay.hz, x+1, lay.hz, term.Lerp(boxes[k%4], haze, 0.6))
		}
	}
	for k, x := 0, sx(66); x <= c.W-4; k, x = k+1, x+2 {
		if scape.HashF(x, 98, seed) < 0.75 {
			fill(c, x, lay.hz, x+1, lay.hz, term.Lerp(boxes[(k+1)%4], haze, 0.6))
		}
	}
	lhX := c.W - 5
	fill(c, lhX, sy(5), lhX, lay.hz, grey(20))
	fill(c, lhX, sy(4), lhX, sy(4), cube(175, 0, 0))

	// The water: the shore's own sea, with chop and the quay's lights laid
	// on it by night.
	top := p.SeaFar
	vramp(c, 0, lay.waterTop, c.W-1, lay.waterBot, p.SeaFar, p.SeaNear)
	for y := lay.waterTop; y <= lay.waterBot; y++ {
		for x := 0; x < c.W; x++ {
			if scape.HashF(x, y, seed+99) < 0.07 {
				plot(far, x, y, '~', lighten(top, 0.18), 0.5)
			}
		}
	}
	for x := 3; x <= c.W-4; x += 2 {
		if lf > 0.3 && scape.HashF(x, 92, seed) < 0.35*lf {
			for y := lay.waterTop; y <= lay.wy(11); y++ {
				if scape.HashF(x, y, seed+94) < 0.7 {
					plot(far, x, y, ':', warmLight, 0.5*lf)
				}
			}
		}
	}
	hull := greyBetween(1, 3, l)
	white := greyBetween(13, 19, l)
	wake := func(x0, x1, y int) {
		for x := x0; x <= x1; x++ {
			if scape.HashF(x, y, seed+95) < 0.7 {
				plot(far, x, y, '.', lighten(top, 0.3), 0.6)
			}
		}
	}
	drift := func(amp, w, ph float64) int { return int(math.Round(amp * math.Sin(t*w+ph))) }
	// Far: a container ship, long and low, and two sailboats.
	sh := sx(26)
	fill(c, sh, lay.wy(10), sh+17, lay.wy(10), hull)
	for k, x := 0, sh+2; x <= sh+11; k, x = k+1, x+2 {
		fill(c, x, lay.wy(10)-1, x+1, lay.wy(10)-1, term.Lerp(boxes[k%4], haze, 0.5))
	}
	fill(c, sh+14, lay.wy(10)-2, sh+16, lay.wy(10)-1, white)
	plot(mid, sh+15, lay.wy(10)-3, '|', hull, 1)
	wake(sh-7, sh-1, lay.wy(10))
	sail := func(x, y int) {
		fill(c, x, y, x+2, y, hull)
		plot(mid, x+1, y-2, '|', white, 1)
		c.SetBGQuad(x+1, y-1, white, ground(x+1, y-1), 0b1010)
		c.SetBGQuad(x+1, y-2, white, ground(x+1, y-2), 0b0010)
	}
	sail(sx(6)+drift(2, 0.08, 0), lay.wy(11))
	sail(sx(56)+drift(2, 0.1, 2), lay.wy(10))
	// Middle: the ferry crossing, its windows lit by night, a buoy.
	const ferryPeriod = 75.0
	fx := -14 + int(math.Mod(t, ferryPeriod)/ferryPeriod*float64(c.W+28))
	fy := lay.wy(12)
	fill(c, fx, fy, fx+13, fy, hull)
	fill(c, fx+1, fy-1, fx+12, fy-1, white)
	for x := fx + 2; x <= fx+11; x += 2 {
		col := darken(white, 0.35)
		if lf > 0.3 {
			col = warmLight
		}
		if x >= 0 && x < c.W {
			c.SetBG(x, fy-1, col)
		}
	}
	plot(mid, fx+7, fy-2, '|', hull, 1)
	plot(near, fx+13, fy-1, '.', cube(0, 135, 0), 0.8)
	wake(fx-7, fx-1, fy)
	plot(near, lay.mull+4, lay.wy(13), 'o', cube(175, 95, 95), 1)
	// Near: the tug working left, and the boat moored at our dock.
	const tugPeriod = 110.0
	// From past the right edge to past the left one, so the wrap happens
	// off screen (it jumped back mid-water, Kimi 2026-11-09).
	tx := c.W + 6 - int(math.Mod(t, tugPeriod)/tugPeriod*float64(c.W+22))
	ty := lay.waterBot
	fill(c, tx, ty, tx+9, ty, hull)
	fill(c, tx+2, ty-1, tx+5, ty-1, white)
	fill(c, tx+6, ty-1, tx+6, ty-1, hull)
	plot(near, tx+6, ty-2, '.', cube(175, 0, 0), 0.8)
	wake(tx+10, tx+15, ty)
	mb := sx(22)
	fill(c, mb, ty, mb+11, ty, hull)
	fill(c, mb+4, ty-1, mb+7, ty-1, white)
	mast := '|'
	if math.Sin(t*1.6) > 0.35 {
		mast = '/'
	}
	plot(mid, mb+2, ty-1, '|', hull, 1)
	plot(mid, mb+2, ty-2, mast, hull, 1)
	// Gulls by day, wheeling.
	if l > 0.4 {
		for k, g := range [][2]int{{30, 5}, {33, 4}, {58, 6}, {61, 5}} {
			plot(far, sx(g[0])+drift(3, 0.7, float64(k)), sy(g[1])+drift(1, 1.1, float64(k)*2), 'v', grey(18), 0.7)
		}
	}
	// The lighthouse flashes every four seconds by night.
	if lf > 0.3 && math.Mod(t, 4) < 0.5 {
		ly := sy(4)
		plot(near, lhX, ly, '*', warmLight, 1)
		for dx := -3; dx <= 0; dx++ {
			for dy := -1; dy <= 1; dy++ {
				x, y := lhX+dx, ly+dy
				if x >= 0 && y >= 0 {
					c.SetBG(x, y, term.Lerp(ground(x, y), warmLight, 0.25*float64(4+dx)/4))
				}
			}
		}
	}

	// The dock: its edge, the planks, bollards, crates, a lamp.
	deck := greyBetween(3, 7, l)
	d0 := lay.dockTop
	fill(c, 0, d0, c.W-1, d0, lighten(deck, 0.25))
	fill(c, 0, d0+1, c.W-1, lay.paneBot, deck)
	for y := d0 + 1; y <= lay.paneBot; y++ {
		for x := (y * 3) % 8; x < c.W; x += 8 {
			plot(far, x, y, '|', darken(deck, 0.3), 0.6)
		}
	}
	for _, bx := range []int{sx(8), sx(36)} {
		fill(c, bx, d0, bx, d0+1, darken(deck, 0.5))
	}
	crate := term.Lerp(cube(135, 95, 95), deck, 0.3)
	cx := sx(44)
	if d0+3 <= lay.paneBot {
		fill(c, cx, d0+2, cx+3, d0+3, crate)
		fill(c, cx+1, d0+1, cx+2, d0+1, crate)
		for y := d0 + 1; y <= d0+3; y++ {
			plot(mid, cx, y, '|', darken(crate, 0.4), 0.8)
			plot(mid, cx+3, y, '|', darken(crate, 0.4), 0.8)
		}
	}
	if d0+2 <= lay.paneBot {
		plot(near, sx(22), d0+2, 'o', darken(deck, 0.4), 1)
		plot(near, sx(22)+1, d0+2, 'o', darken(deck, 0.4), 1)
	}
	lamp := sx(54)
	for y := d0; y <= min(d0+2, lay.paneBot); y++ {
		plot(mid, lamp, y, '|', grey(9), 1)
	}
	if lf > 0.3 {
		plot(near, lamp, d0-1, '*', warmLight, 0.9)
		for dy := -1; dy <= 3; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x, y := lamp+dx, d0-1+dy
				k := 0.28 * lf * (1 - float64(dx*dx+dy*dy*2)/20)
				if k > 0 && x >= 0 && x < c.W && y >= 0 && y <= lay.paneBot {
					c.SetBG(x, y, term.Lerp(ground(x, y), warmLight, k))
				}
			}
		}
	}
}

// drawSillPlantAt is drawSillPlant (rain_study.go) moved down by dy rows,
// so it stands on the layout's sill.
func drawSillPlantAt(c *canvas.Canvas, x, dy int, l float64) { drawSillPlantShifted(c, x, dy, l) }
