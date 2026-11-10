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
	totalRaw, totalGz := 0, 0
	for _, cl := range allFX(pal) {
		frames, cols, rows, _, err := framesOf(7, cl)
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
		t.Logf("%-8s %3d frames  %2dx%-3d  raw %7d  gzip %6d", cl.key, len(frames), cols, rows, raw, g)
	}
	t.Logf("palette: %d bytes of CSS, %d gzipped", len(pal.CSS()), gz(pal.CSS()))
	totalGz += gz(pal.CSS())
	t.Logf("TOTAL    raw %d  gzip %d  (today's GIFs: 5,100,000 raw)", totalRaw, totalGz)
	// 900 KB was Claude's line when the page went to text; 925 KB from
	// 2026-11-10, when the harbour clip became the live scape (a real session
	// fold costs ~19 KB more gzipped than the study it replaced).
	if totalGz > 925_000 {
		t.Errorf("the embedded frames gzip to %d bytes, which is too much page to wait for", totalGz)
	}
}
