package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scenes"
)

// TestFlickerProbe replays a real hook log through the whole frame pipeline at
// the live frame rate and counts, per scape row, how often its rendered bytes
// change and how often they OSCILLATE (A -> B -> A), so "the art flickers
// while it works" becomes a count instead of an impression.
//
// The sea, the glitter and the swells move by design, so "changed every
// frame" is not a defect on its own. An oscillation -- a row flipping back to
// exactly what it was two frames ago -- is different: that is a thing that
// cannot make up its mind, which is what flicker reads as on the glass.
//
//	XSCAPES_KIMILOG=~/.config/xscapes/kimi-hooks.jsonl go test -run TestFlickerProbe -v .
//
// XSCAPES_KIMILOG_SIZE=131x54 overrides the window geometry (scape rows are
// derived the way inside.go derives them).
func TestFlickerProbe(t *testing.T) {
	t.Setenv("XSCAPES_SILENT", "1")
	path := os.Getenv("XSCAPES_KIMILOG")
	if path == "" {
		t.Skip("set XSCAPES_KIMILOG to a hook log to replay it")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	type stamped struct {
		at time.Time
		e  event.Event
	}
	var all []stamped
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 8<<20)
	var t0 time.Time
	for sc.Scan() {
		var l struct {
			TS      int64           `json:"ts"`
			Argv    []string        `json:"argv"`
			Payload json.RawMessage `json:"payload"`
		}
		if json.Unmarshal(sc.Bytes(), &l) != nil || len(l.Payload) == 0 {
			continue // an outcome line, or noise
		}
		now := time.UnixMilli(l.TS)
		if t0.IsZero() {
			t0 = now
		}
		for _, e := range hookTranslate(l.Argv, l.Payload) {
			all = append(all, stamped{now, e})
		}
	}
	if len(all) == 0 {
		t.Skip("no events in the log")
	}

	winCols, winRows := 131, 54
	if v := os.Getenv("XSCAPES_KIMILOG_SIZE"); v != "" {
		fmt.Sscanf(v, "%dx%d", &winCols, &winRows)
	}
	cols := winCols
	_, scapeRows := host.BandWith(winRows, 0)
	fr := newFrames(cols, scapeRows, 7, false, true, 0, 0)
	fr.start = t0
	r := reduce.New("replay")
	fr.follow(nil, r)

	// Per-row counters across the whole replay. When a row oscillates, the
	// column range of the flip is read off the stripped runes, so the report
	// says WHERE, not just how often.
	type rowstat struct {
		row        int
		changes    int
		oscillates int
		oscCol     [2]int // rune range of the last oscillation's flip
		last, prev string // raw rendered row, for the A/B samples
	}
	stats := make([]rowstat, scapeRows)
	for y := range stats {
		stats[y].row = y
	}
	// A short trace of whole rows from the first oscillation onward, so the
	// pattern can be seen, not just counted.
	var trace [][]string
	var rawFlip []string // the three raw versions of the first flipping row
	// The raw pair of the MOST RECENT oscillation anywhere, whatever row it
	// was on: the first flip shows a glyph blink, the steady state shows
	// what actually dominates the count.
	var lastOscA, lastOscB string
	lastOscRow := -1
	// Timeline of the first flipping cell: stripped char per frame, so the
	// on/off pattern over time is visible (periodic? sporadic? one frame?).
	var cellTL []rune
	cellTLRow, cellTLCol := -1, -1
	dirtyPerFrame := []int{}
	i := 0
	last := all[len(all)-1].at
	const fps = 12
	frame := time.Second / fps
	dirtyNow := 0
	flush := func() {
		if len(dirtyPerFrame) == 0 || dirtyNow != dirtyPerFrame[len(dirtyPerFrame)-1] {
			dirtyPerFrame = append(dirtyPerFrame, dirtyNow)
		} else {
			dirtyPerFrame[len(dirtyPerFrame)-1]++
		}
	}
	var prevRows []string
	for now := t0; !now.After(last.Add(30 * time.Second)); now = now.Add(frame) {
		for i < len(all) && !all[i].at.After(now) {
			r.Apply(all[i].e, now)
			i++
		}
		rows := strings.Split(fr.frame(now), "\n")
		dirtyNow = 0
		for y := 0; y < scapeRows && y < len(rows); y++ {
			s := &stats[y]
			if prevRows == nil {
				s.last = rows[y]
				continue
			}
			if rows[y] == s.last {
				continue
			}
			dirtyNow++
			s.changes++
			if rows[y] == s.prev && s.prev != "" {
				s.oscillates++
				s.oscCol = diffRange(stripANSI(s.prev), stripANSI(rows[y]))
				lastOscA, lastOscB, lastOscRow = rows[y], s.last, y
				if len(trace) == 0 {
					// First oscillation anywhere: start tracing here, and
					// keep the raw bytes of the flipping row -- stripped
					// text says WHERE, raw bytes say WHAT changed. s.prev
					// and s.last are that row's raw strings from the two
					// previous frames.
					rawFlip = []string{s.prev, s.last, rows[y]}
					dr := diffRange(stripANSI(s.prev), stripANSI(rows[y]))
					cellTLRow, cellTLCol = y, dr[0]
					for yy := 0; yy < scapeRows && yy < len(rows); yy++ {
						trace = append(trace, []string{stripANSI(rows[yy])})
					}
				}
			}
			s.prev = s.last
			s.last = rows[y]
		}
		if len(trace) > 0 && len(trace[0]) < 20 {
			for yy := 0; yy < scapeRows && yy < len(rows); yy++ {
				trace[yy] = append(trace[yy], stripANSI(rows[yy]))
			}
		}
		if cellTLRow >= 0 && len(cellTL) < 150 {
			r := []rune(stripANSI(rows[cellTLRow]))
			if cellTLCol < len(r) {
				cellTL = append(cellTL, r[cellTLCol])
			} else {
				cellTL = append(cellTL, '¿')
			}
		}
		prevRows = rows
		flush()
	}

	flat := []rowstat{}
	for y := range stats {
		if stats[y].changes > 0 {
			flat = append(flat, stats[y])
		}
	}
	sort.SliceStable(flat, func(a, b int) bool {
		if flat[a].oscillates != flat[b].oscillates {
			return flat[a].oscillates > flat[b].oscillates
		}
		return flat[a].changes > flat[b].changes
	})
	frames := int(last.Add(30*time.Second).Sub(t0) / frame)
	fmt.Printf("window %dx%d, %d events, %d frames at %d fps\n", cols, scapeRows, len(all), frames, fps)
	if showDirty := os.Getenv("XSCAPES_FLICKER_DIRTY"); showDirty != "" {
		fmt.Printf("dirty rows per frame, run-length encoded: %v\n", dirtyPerFrame)
	}
	fmt.Println("\nrow  changes  osc  cols   current content (stripped, truncated)")
	shown := 0
	for _, s := range flat {
		if shown >= 25 {
			break
		}
		fmt.Printf("%3d  %7d %5d  %3d-%-3d %q\n", s.row, s.changes, s.oscillates, s.oscCol[0], s.oscCol[1], trunc(stripANSI(s.last), 60))
		shown++
	}
	if lastOscRow >= 0 {
		fmt.Printf("\nlast oscillation, row %d, raw t-0 vs t-1 (the state between):\n  t-0 %q\n  t-1 %q\n",
			lastOscRow, lastOscA, lastOscB)
	}
	if len(cellTL) > 0 {
		fmt.Printf("\ncell (%d,%d) over %d frames, one char per frame:\n%s\n",
			cellTLRow, cellTLCol, len(cellTL), string(cellTL))
	}
	if len(rawFlip) == 3 {
		dr := diffRange(stripANSI(rawFlip[0]), stripANSI(rawFlip[1]))
		fmt.Printf("\nfirst flip: row bytes %d, stripped cols %d-%d differ between t-2 and t-1; t-0 == t-2\n",
			len(rawFlip[0]), dr[0], dr[1])
		for i, r := range rawFlip {
			sr := []rune(stripANSI(r))
			tail := r
			if len(tail) > 24 {
				tail = tail[len(tail)-24:]
			}
			fmt.Printf("  t-%d runes=%d bytes=%d tail=%q\n", 2-i, len(sr), len(r), tail)
		}
	}
	if len(trace) > 0 {
		fmt.Println("\n20 frames from the first oscillation, whole rows stripped:")
		for yy := range trace {
			fmt.Printf("row %2d | %s\n", yy, trunc(strings.Join(trace[yy], " │ "), 400))
		}
	}
}

// TestFlickerAmbientOnly renders the scape with NO events at all -- static
// state, only the wall clock advancing -- and counts how many rows change
// per frame and how many single-frame blinks occur. It exists because of
// his report of 2026-09-17 ("a flicker in the art as you are working... now
// worse"): replaying the live session log through the frame pipeline showed
// thousands of A->B->A row oscillations, and this test isolates the moving
// parts with nothing happening at all.
//
// The moving parts it measures, per the vista's own design (forest.go):
// the wind-blown leaves ("what the wind carries", drifted right at twenty
// columns a second, density 0.004+0.05*level), the fire's flames, smoke
// and sparks (re-rolled on the six-per-second animation frame, spark
// density 0.004+0.04*level), and the scrub's flutter. Every density climbs
// with the activity level, which is why the scene flickers harder while
// the agent works. The report the test prints is the baseline to fix
// against; XSCAPES_FLICKER_GATE turns the baseline into a hard gate.
//
//	XSCAPES_FLICKER_GATE=1 go test -run TestFlickerAmbientOnly -v .
func TestFlickerAmbientOnly(t *testing.T) {
	t.Setenv("XSCAPES_SILENT", "1")
	cols, scapeRows := 131, 24
	fr := newFrames(cols, scapeRows, 7, false, true, 0, 20.0/24) // pinned evening
	// Pin the scape: the measurement is of the vista, the scene the flicker
	// report came from, and the saved preference is the user's own.
	fr.scapeName = ScapeVista
	fr.vista = scenes.NewVista(7, false)
	fr.follow(nil, reduce.New("ambient"))
	start := time.Date(2026, 9, 17, 20, 0, 0, 0, time.UTC)
	fr.start = start

	stats := make([][2]int, scapeRows) // changes, single-frame blinks per row
	lastChange := make([]int, scapeRows)
	for i := range lastChange {
		lastChange[i] = -2
	}
	var prevRows []string
	const fps = 12
	frame := time.Second / fps
	blinks := 0 // row changes whose previous state lasted exactly one frame
	for f := 0; f < 60*fps; f++ {
		now := start.Add(time.Duration(f) * frame)
		rows := strings.Split(fr.frame(now), "\n")
		for y := 0; y < scapeRows && y < len(rows); y++ {
			s := &stats[y]
			if prevRows == nil {
				continue
			}
			if rows[y] == prevRows[y] {
				continue
			}
			s[0]++
			if lastChange[y] == f-1 {
				// The previous state lasted exactly one frame: a blink.
				s[1]++
				blinks++
			}
			lastChange[y] = f
		}
		prevRows = rows
	}
	tot := [2]int{}
	for y := range stats {
		tot[0] += stats[y][0]
		tot[1] += stats[y][1]
		if stats[y][0] > 0 {
			fmt.Printf("row %2d: %5d changes, %4d single-frame blinks\n", y, stats[y][0], stats[y][1])
		}
	}
	fmt.Printf("total: %d changes, %d single-frame blinks in 60s at %d fps\n", tot[0], blinks, fps)
	if os.Getenv("XSCAPES_FLICKER_GATE") != "" && blinks > 30 {
		t.Errorf("scape blinks %d times with nothing happening", blinks)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < '@' || s[i] > '~') {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// diffRange returns the rune column range where a and b differ, inclusive.
func diffRange(a, b string) [2]int {
	ar, br := []rune(a), []rune(b)
	lo, hi := 0, 0
	for lo < len(ar) && lo < len(br) && ar[lo] == br[lo] {
		lo++
	}
	ra, rb := len(ar)-1, len(br)-1
	for ra >= lo && rb >= lo && ar[ra] == br[rb] {
		ra--
		rb--
	}
	hi = ra
	if rb > ra {
		hi = rb
	}
	if hi < lo {
		hi = lo
	}
	return [2]int{lo, hi}
}
