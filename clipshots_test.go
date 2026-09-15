package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donlucasx/xscapes/internal/canvas"
)

// TestClipShots writes every embedded clip as a frame page, so a REAL look at
// what the page will play is one command away:
//
//	XSCAPES_CLIPSHOTS=<dir> go test -run TestClipShots .
//	for f in <dir>/*.html; do "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
//	  --headless=new --disable-gpu --use-mock-keychain --hide-scrollbars \
//	  --window-size=1600,1400 --screenshot="${f%.html}.png" "file://$f"; done
//
// Eight frames of each clip, evenly spaced, in the page's own palette. It is
// the same instrument notes/vistashots is for the vista: every visual defect
// in session 32 was seen in a screenshot and none in a text dump. Skipped
// unless the directory is named, so the suite never writes anywhere.
func TestClipShots(t *testing.T) {
	dir := os.Getenv("XSCAPES_CLIPSHOTS")
	if dir == "" {
		t.Skip("set XSCAPES_CLIPSHOTS=<dir> to write the frame pages")
	}
	only := map[string]bool{}
	for _, k := range strings.Fields(os.Getenv("XSCAPES_CLIPSHOTS_ONLY")) {
		only[k] = true
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pal := &canvas.HTMLPalette{}
	type page struct{ key, body string }
	var pages []page
	for _, cl := range allFX(pal) {
		if len(only) > 0 && !only[cl.key] {
			continue
		}
		frames, cols, rows, fps, err := framesOf(7, cl)
		if err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, `<h2>%s -- %dx%d, %d frames at %d fps: %s</h2>`, cl.key, cols, rows, len(frames), fps, cl.note)
		n := 8
		if len(frames) < n {
			n = len(frames)
		}
		for i := 0; i < n; i++ {
			k := i * len(frames) / n
			fmt.Fprintf(&b, `<div class="row"><div class="lab">frame %d</div><div class="fx">%s</div></div>`, k, frames[k])
		}
		pages = append(pages, page{cl.key, b.String()})
	}
	// The palette is complete only after the last frame.
	style := `<meta charset="utf-8"><style>html,body{margin:0;padding:12px;background:#000;color:#ccc;font:13px Menlo,monospace}` +
		`.fx pre{margin:0;font-family:Menlo,"SF Mono",monospace;line-height:1;letter-spacing:0;white-space:pre;display:block;font-size:11px}` +
		`.row{display:flex;gap:10px;align-items:flex-start;margin-bottom:8px}.lab{width:90px;padding-top:6px}` +
		`h2{font:bold 13px Menlo,monospace;color:#eee;margin:10px 0 8px}` + pal.CSS() + `</style>`
	for _, p := range pages {
		if err := os.WriteFile(filepath.Join(dir, p.key+".html"), []byte(style+p.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d pages in %s", len(pages), dir)
}
