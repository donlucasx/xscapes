package main

// The QT card (2026-09-19): the site's "Scapes and companions" section as a
// 16:9 loop for a quote post of the trailer. Nothing here ships in the
// product; this file is the only seam, and like trailer_test.go it only
// hands the stage (notes/qt/render-card.mjs) the same markup the site plays.
//
//	XSCAPES_QT=<dir> go test -run TestQTExport .
//
// writes <dir>/cast.js: the four cast portraits (siteframes.go castClips: the
// working pose, 18x9, on the transparent ground, cropped to each animal's
// own ink box) at XSCAPES_QT_FPS (30) for the loop length of each, from
// XSCAPES_QT_SECS ("crab:4.4,cat:4.4,owl:3,frog:4.4"; TestQTPeriods measures
// them), with the palette CSS the frames share. TestQTPeriods prints, for
// each animal, the shortest loop that closes on itself at that rate.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/scenes"
	"github.com/donlucasx/xscapes/internal/term"
)

var qtCast = []string{companion.NameCrab, companion.NameCat, "owl", "frog"}

func qtKey(who string) string {
	switch who {
	case companion.NameCrab:
		return "crab"
	case companion.NameCat:
		return "cat"
	}
	return who
}

// qtCanvases draws one cast portrait frame by frame, as portraitCanvases
// does for the site's cast (the working pose, no session, no balloon), with
// two differences for a 30 fps loop: the owl moves (DrawOwlMoving with the
// working pick, as the legend draws it; the cast's owl is a still face that
// blinks, which at 30 fps would be a still), and a blink is HELD for
// blinkHold frames so it can be seen at all (one frame at 6 fps is 167 ms;
// at 30 fps it is 33 ms).
// qtOwlOffset is where in its 10 s motion the owl's clock starts: its
// picked working motion acts for the first 3.2 s of every period and is
// quiet for the rest, so starting 4 s in puts the quiet stretch across the
// loop's seam (at 13.2 s the owl is 7.2 s into a period, quiet too).
const qtOwlOffset = 4.0

// qtCatOffset starts the crab's and the cat's clocks half a second in. At
// t = 0 both are mid-blink (math.Mod(t, 5.3) < 0.16) and the breath sits on
// its zero crossing (math.Sin(2*pi*t/2.2) > 0 is a coin toss in floating
// point at every multiple of 2.2), so a loop that starts there seams on the
// eyes and on the whole body's row; half a second in, both are settled.
const qtCatOffset = 0.5

func qtCanvases(who string, fps int, secs float64, blinkHold int) (cs []*canvas.Canvas, box [4]int, err error) {
	const cols, rows, px, py = 18, 9, 3, 1
	n := int(secs*float64(fps) + 0.5)
	if n <= 0 {
		return nil, box, fmt.Errorf("%s: no frames", who)
	}
	var cat *companion.Cat
	switch who {
	case companion.NameCrab, companion.NameCat:
		cat = companion.New(who)
		cat.FaceLeft(true)
	}
	cs = make([]*canvas.Canvas, 0, n)
	x0, y0, x1, y1 := cols, rows, 0, 0
	for i := 0; i < n; i++ {
		t := float64(i) / float64(fps)
		// The frog's blink, twice a loop, held.
		half := max(1, n/2)
		blink := (i-3+half)%half < blinkHold
		c := canvas.New(cols, rows, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
		c.Clear()
		for y := 0; y < c.H; y++ {
			for x := 0; x < c.W; x++ {
				c.SetBG(x, y, portraitGround)
			}
		}
		switch who {
		case "owl":
			scenes.DrawOwlMoving(c, px, py, t+qtOwlOffset, companion.Working, scenes.PickedOwlMotion(companion.Working), 0)
		case "frog":
			scenes.DrawCandidate(c, scenes.FindAnimal("Frog"), px, py, true, blink)
		default:
			cat.Approach(1/float64(fps), false)
			dx, w, _ := cat.DrawnBox(c.W)
			x := (c.W-w)/2 - dx
			cat.Draw(c.Near(), x, py, t+qtCatOffset, companion.Working)
		}
		cs = append(cs, c)
		ax0, ay0, ax1, ay1 := inkBox(c)
		x0, y0, x1, y1 = min(x0, ax0), min(y0, ay0), max(x1, ax1), max(y1, ay1)
	}
	if x1 <= x0 || y1 <= y0 {
		return nil, box, fmt.Errorf("%s: nothing drawn", who)
	}
	box = [4]int{max(x0-1, 0), max(y0-1, 0), min(x1+1, cols), min(y1+1, rows)}
	return cs, box, nil
}

func qtFrames(who string, fps int, secs float64, blinkHold int, pal *canvas.HTMLPalette) (frames []string, cols, rows int, err error) {
	cs, box, err := qtCanvases(who, fps, secs, blinkHold)
	if err != nil {
		return nil, 0, 0, err
	}
	x0, y0, x1, y1 := box[0], box[1], box[2], box[3]
	for _, c := range cs {
		frames = append(frames, c.HTMLFragmentCropClassed(x0, y0, x1, y1, GIFPx, term.ProfileTrueColor, pal))
	}
	return frames, x1 - x0, y1 - y0, nil
}

// TestQTPeriods prints the shortest loop each animal closes on at 30 fps,
// measured on the rendered frames (no blink, so the blink cannot hide a
// seam): the smallest p with frame[k] == frame[k+p] for every k.
func TestQTPeriods(t *testing.T) {
	if os.Getenv("XSCAPES_QT") == "" {
		t.Skip("set XSCAPES_QT=<dir> to measure the cast's loop periods")
	}
	const fps, secs = 30, 60.0
	for _, who := range qtCast {
		pal := &canvas.HTMLPalette{Transparent: &portraitGround}
		fr, cols, rows, err := qtFrames(who, fps, secs, 0, pal)
		if err != nil {
			t.Fatal(err)
		}
		n := len(fr)
		found := 0
		for p := 1; p <= n/2; p++ {
			ok := true
			for k := 0; k+p < n; k++ {
				if fr[k] != fr[k+p] {
					ok = false
					break
				}
			}
			if ok {
				found = p
				break
			}
		}
		distinct := map[string]bool{}
		for _, f := range fr {
			distinct[f] = true
		}
		if found == 0 {
			t.Logf("%-5s %dx%d: no loop closes inside %.1f s (%d distinct frames of %d)", qtKey(who), cols, rows, secs/2, len(distinct), n)
			continue
		}
		t.Logf("%-5s %dx%d: closes every %d frames = %.3f s (%d distinct frames)", qtKey(who), cols, rows, found, float64(found)/fps, len(distinct))
	}
	// The seam at candidate loop lengths: the frame that would follow the
	// last one against the first, as cells that differ (glyph or colour).
	for _, L := range []float64{4.4, 6.6, 8.8, 11, 13.2, 15.4, 17.6, 22} {
		var parts []string
		for _, who := range qtCast {
			pal := &canvas.HTMLPalette{Transparent: &portraitGround}
			fr, _, _, err := qtFrames(who, fps, L+1, 0, pal)
			if err != nil {
				t.Fatal(err)
			}
			n := int(L*fps + 0.5)
			parts = append(parts, fmt.Sprintf("%s %d", qtKey(who), qtCellDiff(fr[0], fr[n])))
		}
		t.Logf("seam at %5.1f s: %s", L, strings.Join(parts, " · "))
	}
}

// qtCellDiff counts the cells whose markup differs between two frames of the
// same crop: the fragments are one <span> run per colour run, so compare
// them cell by cell after flattening runs to their cells.
func qtCellDiff(a, b string) int {
	ca, cb := qtCells(a), qtCells(b)
	if len(ca) != len(cb) {
		return max(len(ca), len(cb))
	}
	d := 0
	for i := range ca {
		if ca[i] != cb[i] {
			d++
		}
	}
	return d
}

// qtCells flattens a fragment to one string per cell, "class|rune".
func qtCells(frag string) []string {
	var out []string
	class := ""
	for i := 0; i < len(frag); {
		if strings.HasPrefix(frag[i:], "<span") {
			j := strings.IndexByte(frag[i:], '>')
			tag := frag[i : i+j+1]
			if k := strings.Index(tag, "class=\""); k >= 0 {
				rest := tag[k+7:]
				class = rest[:strings.IndexByte(rest, '"')]
			} else {
				class = ""
			}
			i += j + 1
			continue
		}
		if strings.HasPrefix(frag[i:], "</span>") {
			i += 7
			class = ""
			continue
		}
		if frag[i] == '<' { // <pre>, </pre>, <br> and the like
			j := strings.IndexByte(frag[i:], '>')
			i += j + 1
			continue
		}
		if frag[i] == '\n' {
			i++
			continue
		}
		if frag[i] == '&' { // an entity is one cell
			j := strings.IndexByte(frag[i:], ';')
			out = append(out, class+"|"+frag[i:i+j+1])
			i += j + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(frag[i:])
		out = append(out, class+"|"+string(r))
		i += size
	}
	return out
}

// TestQTExport writes the card's frames for the stage.
func TestQTExport(t *testing.T) {
	dir := os.Getenv("XSCAPES_QT")
	if dir == "" {
		t.Skip("set XSCAPES_QT=<dir> to write the card's frames")
	}
	fps := 30
	if s := os.Getenv("XSCAPES_QT_FPS"); s != "" {
		if _, err := fmt.Sscan(s, &fps); err != nil {
			t.Fatal(err)
		}
	}
	secs := map[string]float64{"crab": 4.4, "cat": 4.4, "owl": 4.4, "frog": 4.4}
	if s := os.Getenv("XSCAPES_QT_SECS"); s != "" {
		for _, kv := range strings.Split(s, ",") {
			k, v, ok := strings.Cut(strings.TrimSpace(kv), ":")
			if !ok {
				t.Fatalf("XSCAPES_QT_SECS: %q is not key:secs", kv)
			}
			var f float64
			if _, err := fmt.Sscan(v, &f); err != nil {
				t.Fatal(err)
			}
			secs[k] = f
		}
	}
	blinkHold := max(1, fps/10) // ~100 ms
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pal := &canvas.HTMLPalette{Transparent: &portraitGround}
	out := map[string]any{"fps": fps}
	for _, who := range qtCast {
		key := qtKey(who)
		fr, cols, rows, err := qtFrames(who, fps, secs[key], blinkHold, pal)
		if err != nil {
			t.Fatal(err)
		}
		out[key] = map[string]any{"cols": cols, "rows": rows, "frames": fr}
		t.Logf("%-5s %dx%d, %d frames (%.2f s at %d fps)", key, cols, rows, len(fr), secs[key], fps)
	}
	// The palette's CSS only after the last frame: every frame can add to it.
	out["css"] = pal.CSS()
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "cast.js")
	if err := os.WriteFile(p, []byte("window.TRAILER=window.TRAILER||{};window.TRAILER.cast="+string(b)+";"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s (%d bytes)", p, len(b))
}
