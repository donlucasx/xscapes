package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
	"github.com/donlucasx/asciiscapes/internal/reduce"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// termSize asks stty rather than pulling in a dependency for one ioctl.
// Falls back to the design target when it cannot tell.
func termSize() (w, h int) {
	out, err := exec.Command("stty", "size").Output()
	if err == nil {
		if f := strings.Fields(string(out)); len(f) == 2 {
			if r, err1 := strconv.Atoi(f[0]); err1 == nil {
				if c, err2 := strconv.Atoi(f[1]); err2 == nil && r > 4 && c > 8 {
					return c, r
				}
			}
		}
	}
	return 80, 24
}

// runLive paints the scape to this terminal until interrupted. This is the
// only view that proves anything: the HTML harness models glyph advances, a
// terminal paints a fixed cell grid, and they disagree about non-ASCII glyphs.
//
// With a session to follow it renders that session. Without one it runs the
// demo cycle, which is what every frame in assets/frames was made from and
// what you want when showing the thing to somebody with no agent running.
func runLive(seed int64, fps float64, wIn, hIn int, ctxUsed, tod float64, ascii bool, session string) {
	w, h := wIn, hIn
	if w <= 0 || h <= 0 {
		w, h = termSize()
		h-- // leave the prompt a row to come back to
	}

	// Find the session. Inside the agent's own environment this is exact;
	// from a pane started by hand it falls back to whichever session last
	// announced itself.
	if session == "" {
		session = event.SessionFromEnv()
	}
	if session == "" {
		session = event.Current()
	}

	var bus *event.Bus
	var red *reduce.Reducer
	if session != "" {
		b, err := event.Listen(session)
		switch {
		case err == event.ErrBusy:
			fmt.Fprintf(os.Stderr, "asciiscapes: a scape is already following session %s\n", event.Tag(session))
			os.Exit(1)
		case err != nil:
			// Not fatal. A scape that cannot bind is still a scape; it just
			// runs the demo instead of going dark, and says why.
			fmt.Fprintln(os.Stderr, "asciiscapes:", err)
		default:
			bus, red = b, reduce.New(session)
			defer bus.Close()
		}
	}

	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, ascii)
	cat := companion.NewCat()
	_, chh := cat.Size()

	restore := func() { fmt.Print("\x1b[?25h\x1b[0m\n") }
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		if bus != nil {
			bus.Close()
		}
		restore()
		os.Exit(0)
	}()

	fmt.Print("\x1b[?25l\x1b[2J") // hide cursor, clear once
	defer restore()

	profile := term.DetectProfile()
	tick := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer tick.Stop()

	start := time.Now()
	for range tick.C {
		now := time.Now()
		t := now.Sub(start).Seconds()

		var st reduce.State
		if red != nil {
			// Drain everything that arrived since the last frame. Non
			// blocking: a frame must never wait on the bus.
			for draining := true; draining; {
				select {
				case e := <-bus.C:
					red.Apply(e, now)
				default:
					draining = false
				}
			}
			st = red.State(now)
			// The sky is the world: time of day is the wall clock, never
			// anything the agent did.
			st.Act.TimeOfDay = timeOfDay(now)
			if tod != 0 {
				st.Act.TimeOfDay = tod
			}
			if ctxUsed != 0 {
				st.Act.ContextUsed = ctxUsed
			}
		} else {
			st = demoState(t, ctxUsed, tod)
		}

		sh.Update(c, t, st.Act)
		top := c.H - 2 - chh
		cat.Draw(c.Near(), 5, top, t, st.Pose)

		if st.Kittens > 0 {
			cat.DrawKittens(c.Near(), c.Mid(), 5, top, st.Kittens, c.W-1,
				int(float64(c.H)*0.42)+1, t, seed)
		}
		if st.Bubble != "" {
			rows := companion.Bubble(st.Bubble)
			(&companion.Sprite{Rows: rows, Body: bubbleCol}).Draw(c.Near(), 12, top-len(rows))
		}
		drawSand(c, st.Tail)

		// Home the cursor and overwrite rather than clearing: clearing every
		// frame is what makes a terminal animation flicker.
		fmt.Print("\x1b[H" + c.Render(profile))
	}
}

// demoState is the state cycle used when no session is attached: it swings
// through every behaviour so one sitting shows the whole vocabulary.
func demoState(t, ctxUsed, tod float64) reduce.State {
	st := reduce.State{
		Act:  scape.Activity{Working: true, Level: 0.65, ContextUsed: ctxUsed, TimeOfDay: tod},
		Pose: companion.Working,
	}
	switch int(t/8) % 3 {
	case 1:
		st.Act = scape.Activity{ContextUsed: ctxUsed, TimeOfDay: tod}
		st.Pose = companion.Resting
	case 2:
		st.Pose = companion.NeedsYou
		st.Bubble = "tests passed"
	}
	return st
}

// timeOfDay maps the wall clock to the palette's 0..1, midnight at 0.
func timeOfDay(now time.Time) float64 {
	h, m, s := now.Clock()
	return float64(h*3600+m*60+s) / 86400
}

// drawSand writes the activity tail into the beach.
//
// Sand tones rather than terminal tones, newest brightest, older lines fading
// as the tide takes them. It is always visible, but it is scenery: the sea
// says how much work is happening and the sand says what it is, which is why
// no per-tool weather vocabulary is needed anywhere else.
func drawSand(c *canvas.Canvas, lines []reduce.Line) {
	if len(lines) == 0 {
		return
	}
	sand := term.RGB{R: 76, G: 65, B: 54}
	ink := term.RGB{R: 240, G: 230, B: 210}
	bad := term.RGB{R: 244, G: 176, B: 96}

	// Sit the block just above the bottom edge, newest last so the eye lands
	// on the current line closest to the companion.
	top := c.H - len(lines) - 1
	if top < 2 {
		top = 2
	}
	for i, ln := range lines {
		row := top + i
		if row >= c.H {
			break
		}
		base := ink
		if ln.Bad {
			base = bad
		}
		// Fade on the line's own age rather than on its position, so a line
		// that has been sitting there for two minutes looks it even when
		// nothing newer has arrived to push it down.
		col := term.Lerp(base, sand, 0.15+0.7*ln.Age)
		for x, r := range []rune(ln.Text) {
			if x+2 >= c.W {
				break
			}
			c.Near().Plot(x+2, row, r, col, 1)
		}
	}
}
