// his measures the churn guarantee ONLY at geometries Lucas has actually run,
// read out of his own trace sidecars (~/.config/xscapes/traces/*.log, whose
// columns are traceN cols rows agentRows -- see host.traceSize), and driven by
// the level path his own reducer produces from ~/.config/xscapes/run/*.jsonl.
//
// The s28-guarantees sweep used H in {12,16,24,27,40,51}. Two of those (40, 51)
// are above host.MaxScapeRows and cannot occur; H=12 occurs in no logged window.
//
//	go run ./notes/s28-guarantees-verify/his
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

func edgeOf(sh *scape.Shore) []float64 {
	v := reflect.ValueOf(sh).Elem().FieldByName("lastEdge")
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface().([]float64)
}
func tideRowOf(sh *scape.Shore) int {
	v := reflect.ValueOf(sh).Elem().FieldByName("tideRow")
	return int(reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Int())
}

func newCanvas(w, h int) *canvas.Canvas {
	return canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
}

func openSea(edge []float64, h int) (int, int) {
	lo := h
	for _, e := range edge {
		if int(e) < lo {
			lo = int(e)
		}
	}
	return int(float64(h)*0.42) + 1, lo - 2
}

// geoms reads his real window sizes out of the trace sidecars.
func geoms() [][2]int {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, ".config/xscapes/traces/*.log"))
	seen := map[[2]int]int{}
	n := 0
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(fh)
		for sc.Scan() {
			p := strings.Fields(sc.Text())
			if len(p) != 4 {
				continue
			}
			cols, _ := strconv.Atoi(p[1])
			rows, _ := strconv.Atoi(p[2])
			ag, _ := strconv.Atoi(p[3])
			if rows-ag <= 0 {
				continue
			}
			seen[[2]int{cols, rows - ag}]++
			n++
		}
		fh.Close()
	}
	fmt.Printf("   %d size records in %d trace sidecars, %d distinct cols x scapeRows\n", n, len(files), len(seen))
	var out [][2]int
	for g := range seen {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][1] != out[j][1] {
			return out[i][1] < out[j][1]
		}
		return out[i][0] < out[j][0]
	})
	return out
}

type frame struct{ raw, ren []term.RGB }

func grab(c *canvas.Canvas, from, to, w, h int) frame {
	var f frame
	for y := from; y <= to; y++ {
		if y < 0 || y >= h {
			continue
		}
		for x := 0; x < w; x++ {
			f.raw = append(f.raw, c.BGAt(x, y))
			_, _, bg := c.ResolveAt(x, y, term.Profile256)
			f.ren = append(f.ren, bg)
		}
	}
	return f
}

func main() {
	fmt.Println("== the churn guarantee at geometries he has ACTUALLY run ==")
	fmt.Println("-- his logged window sizes --")
	gs := geoms()
	for _, g := range gs {
		fmt.Printf("      %d cols x %d scape rows\n", g[0], g[1])
	}
	fmt.Println()

	// --- the test's own synthetic STEP at his real sizes ---
	fmt.Println("-- A. the synthetic 0 -> 0.7 STEP (the s28 finding's own metric) at his sizes --")
	fmt.Printf("%-10s %9s %9s %9s %9s\n", "geom", "raw%", "REN%", "cells", "pairs>50%")
	for _, g := range gs {
		w, h := g[0], g[1]
		sh := scape.NewShore(7, false)
		c := newCanvas(w, h)
		tm := 0.0
		for i := 0; i < 400; i++ {
			tm += 0.1
			sh.Update(c, tm, scape.Activity{Level: 0, TimeOfDay: 0.3868})
		}
		var fr []*canvas.Canvas
		for i := 0; i < 24; i++ {
			cc := newCanvas(w, h)
			tm += 0.1
			sh.Update(cc, tm, scape.Activity{Level: 0.7, TimeOfDay: 0.3868})
			fr = append(fr, cc)
		}
		from, to := openSea(edgeOf(sh), h)
		cells, cr, ce, big := 0, 0, 0, 0
		for i := 1; i < len(fr); i++ {
			a, b := grab(fr[i-1], from, to, w, h), grab(fr[i], from, to, w, h)
			pc := 0
			for k := range a.raw {
				cells++
				if a.raw[k] != b.raw[k] {
					cr++
					pc++
				}
				if a.ren[k] != b.ren[k] {
					ce++
				}
			}
			if len(a.raw) > 0 && 100*float64(pc)/float64(len(a.raw)) > 50 {
				big++
			}
		}
		if cells == 0 {
			fmt.Printf("%-10s %9s %9s %9d  EMPTY BAND (rows %d..%d)\n", fmt.Sprintf("%dx%d", w, h), "-", "-", 0, from, to)
			continue
		}
		fmt.Printf("%-10s %8.2f%% %8.2f%% %9d %9d\n", fmt.Sprintf("%dx%d", w, h),
			100*float64(cr)/float64(cells), 100*float64(ce)/float64(cells), cells, big)
	}
	fmt.Println()

	// --- his real level path at his most common size ---
	fmt.Println("-- B. HIS REAL LEVEL PATH at his commonest size, counted per frame --")
	dir, err := event.RunDir()
	if err != nil {
		fmt.Println("   no run dir:", err)
		return
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	type S struct {
		name string
		evs  []event.Event
	}
	var ss []S
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		var evs []event.Event
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			if len(sc.Bytes()) == 0 {
				continue
			}
			var e event.Event
			if json.Unmarshal(sc.Bytes(), &e) != nil {
				continue
			}
			evs = append(evs, e)
		}
		fh.Close()
		if len(evs) > 0 {
			ss = append(ss, S{filepath.Base(f), evs})
		}
	}
	sort.Slice(ss, func(i, j int) bool { return len(ss[i].evs) > len(ss[j].evs) })
	if len(ss) > 4 {
		ss = ss[:4]
	}
	fmt.Printf("%-22s %-9s %8s %10s %10s %10s %9s %9s\n",
		"session", "geom", "frames", "rowCross", "raw>8", "REN>8", "worstRaw", "worstREN")
	for _, s := range ss {
		for _, g := range [][2]int{{124, 26}, {133, 27}, {119, 22}} {
			w, h := g[0], g[1]
			// The band is FIXED at the settled level-0.7 waterline -- the state
			// the guarantee's own test measures -- and not at whatever the first
			// event happens to leave. Taking it from a quiet first event gives an
			// empty band and a column of numbers that mean nothing; this
			// instrument printed EMPTY rather than 0.00% when that happened.
			bsh := scape.NewShore(7, false)
			bc := newCanvas(w, h)
			bt := 0.0
			for i := 0; i < 400; i++ {
				bt += 0.1
				bsh.Update(bc, bt, scape.Activity{Level: 0.7, TimeOfDay: 0.3868})
			}
			from, to := openSea(edgeOf(bsh), h)
			if from > to {
				fmt.Printf("%-22s %-9s  band empty even at level 0.7 -- unmeasurable here\n", s.name, fmt.Sprintf("%dx%d", w, h))
				continue
			}
			sh := scape.NewShore(7, false)
			c := newCanvas(w, h)
			tm := 0.0
			r := reduce.New(s.evs[0].Session)
			at0 := time.UnixMilli(s.evs[0].TS)
			r.Apply(s.evs[0], at0)
			for i := 0; i < 400; i++ {
				tm += 0.1
				sh.Update(c, tm, r.State(at0).Act)
			}
			prev := grab(c, from, to, w, h)
			prevRow := tideRowOf(sh)
			now := at0
			frames, cross, oR, oE, cells := 0, 0, 0, 0, 0
			wR, wE := 0.0, 0.0
			for _, e := range s.evs[1:] {
				at := time.UnixMilli(e.TS)
				gap := math.Min(60, math.Max(0, at.Sub(now).Seconds()))
				for i := 0; i < int(gap/0.1) && frames < 12000; i++ {
					now = now.Add(100 * time.Millisecond)
					cc := newCanvas(w, h)
					tm += 0.1
					sh.Update(cc, tm, r.State(now).Act)
					// The band is fixed at the settled geometry so every pair is
					// comparable; rows that leave the open sea are counted, which
					// is exactly what the guarantee itself does.
					cur := grab(cc, from, to, w, h)
					if len(cur.raw) == 0 {
						break
					}
					cr, ce := 0, 0
					for k := range cur.raw {
						if cur.raw[k] != prev.raw[k] {
							cr++
						}
						if cur.ren[k] != prev.ren[k] {
							ce++
						}
					}
					n := len(cur.raw)
					cells += n
					pr := 100 * float64(cr) / float64(n)
					pe := 100 * float64(ce) / float64(n)
					if pr > 8 {
						oR++
					}
					if pe > 8 {
						oE++
					}
					if pr > wR {
						wR = pr
					}
					if pe > wE {
						wE = pe
					}
					if tideRowOf(sh) != prevRow {
						cross++
					}
					prevRow = tideRowOf(sh)
					prev = cur
					frames++
				}
				now = at
				r.Apply(e, at)
				if frames >= 12000 {
					break
				}
			}
			if frames == 0 || cells == 0 {
				fmt.Printf("%-22s %-9s  EMPTY -- no frames or no cells, no conclusion\n", s.name, fmt.Sprintf("%dx%d", w, h))
				continue
			}
			fmt.Printf("%-22s %-9s %8d %10d %10d %10d %8.2f%% %8.2f%%   (%d cells)\n",
				s.name, fmt.Sprintf("%dx%d", w, h), frames, cross, oR, oE, wR, wE, cells)
		}
	}
}
