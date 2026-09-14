package main

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// What every embedded animation costs, so the frame budget is chosen from
// numbers rather than guessed. GitHub Pages serves gzipped, so the gzipped
// column is the one a visitor waits for.
func TestWhatTheEmbeddedFramesCost(t *testing.T) {
	gz := func(s string) int {
		var b bytes.Buffer
		w, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
		w.Write([]byte(s))
		w.Close()
		return b.Len()
	}
	pal := &canvas.HTMLPalette{}
	if truecolorFX {
		t.Log("rendering TRUECOLOR")
	}
	clips := append([]fxClip{heroClip(pal)}, stateClips(pal)...)
	clips = append(clips, swapClips(pal)...)
	clips = append(clips, ruleClip(pal))
	totalRaw, totalGz := 0, 0
	for _, cl := range clips {
		frames, err := gifFrames(7, cl.sc)
		if err != nil {
			t.Fatalf("%s: %v", cl.key, err)
		}
		raw := 0
		for _, f := range frames {
			raw += len(f)
		}
		all := ""
		for _, f := range frames {
			all += f
		}
		g := gz(all)
		totalRaw += raw
		totalGz += g
		t.Logf("%-8s %3d frames  %2dx%-3d  raw %7d  gzip %6d", cl.key, len(frames), cl.sc.colsOf(), cl.sc.rowsOf(), raw, g)
	}
	t.Logf("palette: %d bytes of CSS, %d gzipped", len(pal.CSS()), gz(pal.CSS()))
	totalGz += gz(pal.CSS())
	t.Logf("TOTAL    raw %d  gzip %d  (today's GIFs: 5,100,000 raw)", totalRaw, totalGz)
	if totalGz > 900_000 {
		t.Errorf("the embedded frames gzip to %d bytes, which is too much page to wait for", totalGz)
	}
}
