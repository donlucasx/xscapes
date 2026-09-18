package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scenes"
)

// TestReplayAHookLog folds a raw hook log (XSCAPES_HOOKLOG's file) through
// the hook command's own translation and the reducer, at the log's real
// timestamps, and prints what the scene would have been told at each event:
// the instrument for "what did the scape see during his run".
//
//	XSCAPES_KIMILOG=~/.config/xscapes/kimi-hooks.jsonl go test -run TestReplayAHookLog -v .
func TestReplayAHookLog(t *testing.T) {
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
	r := reduce.New("replay")
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
			r.Apply(e, now)
			st := r.State(now)
			fmt.Printf("%7.1fs %-12s %-10s agent=%-12s src=%-5s | kittens=%d exits=%d level=%.2f pose=%v working=%v bubble=%q\n",
				now.Sub(t0).Seconds(), e.Kind, e.Tool, e.Agent, e.Src, st.Kittens, len(st.KittenExits), st.Act.Level, st.Pose, st.Act.Working, st.Bubble)
		}
	}

	// Then what the VISTA would have drawn at his geometry (131x54, the
	// scape's share of the rows), second by second: the litter the reducer
	// counts against the owlets DrawFlock puts on the canvas.
	cols, rows := 131, 54
	if v := os.Getenv("XSCAPES_KIMILOG_SIZE"); v != "" {
		fmt.Sscanf(v, "%dx%d", &cols, &rows)
	}
	_, scapeRows := host.BandWith(rows, 0)
	fr := newFrames(cols, scapeRows, 7, false, true, 0, 0)
	fr.scapeName = ScapeVista
	fr.vista = scenes.NewVista(7, false)
	fr.start = t0
	r2 := reduce.New("replay")
	fr.follow(nil, r2)
	i := 0
	if len(all) == 0 {
		t.Skip("no payload lines in the log")
	}
	last := all[len(all)-1].at
	fmt.Printf("\n-- the vista, %dx%d: t, kittens, exits, owlets drawn sitting, flying\n", cols, scapeRows)
	prevLine := ""
	for now := t0; !now.After(last.Add(90 * time.Second)); now = now.Add(time.Second) {
		for i < len(all) && !all[i].at.After(now) {
			r2.Apply(all[i].e, now)
			i++
		}
		tt := now.Sub(t0).Seconds()
		fr.frame(now)
		st := fr.state(now, tt)
		owlX, owlY, grassY, _ := fr.vista.Layout()
		sitting := scenes.DrawFlock(fr.c, &fr.vista.Flock, st.Kittens, st.KittenExits, owlX, owlY, grassY, fr.vista.FireX(), 3*fr.c.W/8+6, tt, scenes.PickedOwletMotion(), scenes.OwlMotionPeriodOverride)
		line := fmt.Sprintf("kittens=%d exits=%d sitting=%d", st.Kittens, len(st.KittenExits), sitting)
		if line != prevLine {
			fmt.Printf("%7.0fs %s  (owlX=%d fireX=%d minX=%d)\n", tt, line, owlX, fr.vista.FireX(), 3*fr.c.W/8+6)
			prevLine = line
		}
	}
}
