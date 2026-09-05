package main

import (
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scape"
)

// The context readout lives in the scene from 40% used (his ruling,
// 2026-09-05: "the readout should show up when context hits 40% used"), a dim
// number of what is left under the moon, and turns into a warm "NN% left"
// from 85%. Before 40% the sky carries no digits at all. The number always
// stays in the sky, above the horizon, even with the moon at the waterline.
func TestTheReadoutAppearsAtFortyPercentUsed(t *testing.T) {
	const w, h = 120, 24
	render := func(used float64) (text string, row int, warm bool) {
		c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		sh := scape.NewShore(7, false)
		cat := companion.NewCat()
		cat.FaceLeft(true)
		ccw, chh := cat.Size()
		lay := compose(w, ccw, true)
		sh.MoonX = lay.MoonX
		st := demoState(2, used, 0.93)
		st.Tail = st.FitTail(time.Now(), lay.SandTo-lay.SandFrom)
		sh.Update(c, 2, st.Act)
		drawScene(c, sh, cat, lay, st, 2, 7, c.H-2-chh)
		hy := int(float64(c.H)*0.42) + 1
		for y := 0; y < hy; y++ {
			var b strings.Builder
			for x := 0; x < w; x++ {
				cell := c.Near().Cells[y*w+x]
				if cell.Set && cell.R != ' ' {
					b.WriteRune(cell.R)
					if cell.FG == moonLabelWarn {
						warm = true
					}
				} else {
					b.WriteRune(' ')
				}
			}
			if s := strings.TrimSpace(b.String()); strings.Contains(s, "%") {
				return s, y, warm
			}
		}
		return "", -1, false
	}
	if s, _, _ := render(0.39); s != "" {
		t.Errorf("at 39%% used the sky shows %q, want no readout", s)
	}
	if s, y, warm := render(0.40); !strings.Contains(s, "60%") || warm || y < 0 {
		t.Errorf("at 40%% used: got %q (warm=%v, row %d), want a dim 60%%", s, warm, y)
	}
	if s, _, warm := render(0.85); s != "15% left" || !warm {
		t.Errorf("at 85%% used: got %q (warm=%v), want a warm \"15%% left\"", s, warm)
	}
	if s, y, _ := render(1.0); s != "0% left" || y < 0 {
		t.Errorf("at 100%% used: got %q on row %d, want \"0%% left\" still above the horizon", s, y)
	}
}
