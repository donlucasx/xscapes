package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/donlucasx/asciiscapes/internal/canvas"
	"github.com/donlucasx/asciiscapes/internal/companion"
	"github.com/donlucasx/asciiscapes/internal/event"
	"github.com/donlucasx/asciiscapes/internal/reduce"
	"github.com/donlucasx/asciiscapes/internal/scape"
	"github.com/donlucasx/asciiscapes/internal/term"
)

// termSize asks the kernel how big the window is, via the TIOCGWINSZ ioctl.
//
// It used to shell out to `stty size`, which is where this went wrong: stty
// reads the size from its STDIN, and a child started by exec.Command inherits
// /dev/null unless told otherwise, so it printed "stdin isn't a terminal" and
// termSize silently returned the 80x24 fallback -- every time, on every
// machine. The scene has therefore always been 80x24 no matter how big the
// window was, which is one bug wearing three costumes: a scene that does not
// fill the frame, a scene that garbles when the window is shrunk below 80
// columns (the over-long lines wrap), and a scene that never repaints on a
// resize (the "new" size always compared equal to the old one).
//
// The ioctl has no such problem, needs no subprocess, and costs microseconds
// rather than milliseconds. Ask the controlling terminal directly rather than
// any particular stream, so a redirected stdout does not blind us.
func termSize() (w, h int) {
	for _, fd := range ttyFDs() {
		if c, r, ok := winSize(fd); ok {
			return c, r
		}
	}
	return 80, 24
}

// ttyFDs lists the descriptors worth asking, most authoritative first. If all
// three standard streams are redirected, /dev/tty still reaches the terminal
// this process is attached to.
func ttyFDs() []uintptr {
	fds := []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()}
	if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
		defer f.Close()
		fds = append(fds, f.Fd())
	}
	return fds
}

// winsize mirrors struct winsize from <sys/ioctl.h>: rows, cols, then the
// pixel dimensions, which nothing reports reliably and we do not use.
type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func winSize(fd uintptr) (cols, rows int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.cols < 8 || ws.rows < 4 {
		return 0, 0, false
	}
	return int(ws.cols), int(ws.rows), true
}

// runLive paints the scape to this terminal until interrupted. This is the
// only view that proves anything: the HTML harness models glyph advances, a
// terminal paints a fixed cell grid, and they disagree about non-ASCII glyphs.
//
// With a session to follow it renders that session. Without one it runs the
// demo cycle, which is what every frame in assets/frames was made from and
// what you want when showing the thing to somebody with no agent running.
func runLive(seed int64, fps float64, wIn, hIn int, ctxUsed, tod float64, ascii bool, session string, mirror bool) {
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
			fmt.Fprintf(os.Stderr, "asciiscapes: a scape is already following session %s\n", event.Short(session))
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
	cat.FaceLeft(mirror)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, mirror)
	sh.MoonX = lay.MoonX

	// React to a resized window. Without this the scene keeps the size it had
	// when it started: drag the pane wider and the picture just sits there in
	// the old rectangle. Polling is not an option -- termSize shells out to
	// stty, which is far too expensive to do every frame -- so take the signal.
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)

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
	// A zero fps builds a ticker with a zero duration (panic), and anything
	// under one frame a second fights the shore's own large-gap clamp.
	if fps < 1 {
		fps = 1
	}
	if fps > 120 {
		fps = 120
	}
	tick := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer tick.Stop()

	start := time.Now()
	for range tick.C {
		now := time.Now()
		t := now.Sub(start).Seconds()

		select {
		case <-winch:
			if wIn <= 0 || hIn <= 0 {
				nw, nh := termSize()
				nh--
				if nw != c.W || nh != c.H {
					c.Resize(nw, nh)
					lay = compose(nw, ccw, mirror)
					sh.MoonX = lay.MoonX
					fmt.Print("\x1b[2J") // one clear, or the old frame's edges linger
				}
			}
		default:
		}

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
			// Re-render the sand to the columns this layout actually leaves
			// it, so a narrow pane loses whole pieces of a line rather than
			// getting a path chopped mid-word.
			st.Tail = st.FitTail(now, lay.SandTo-lay.SandFrom)
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
		drawScene(c, sh, cat, lay, st, t, seed, top)

		// Home the cursor and overwrite rather than clearing: clearing every
		// frame is what makes a terminal animation flicker.
		fmt.Print("\x1b[H" + c.Render(profile))
	}
}

// layout is where the composition lives: which side the companion sits on and
// what that implies for everything anchored to it. One function, so the mockup
// and the live loop cannot disagree about it.
type layout struct {
	CatX     int // left cell of the companion sprite
	BubbleX  int
	SandFrom int
	SandTo   int
	MoonX    float64
	Mirror   bool
}

// composeRight is the mirrored composition: companion on the right, litter
// growing leftward, tail written from the left margin, moon moved across.
//
// The moon moves because it is the other thing a glance goes looking for. Left
// where it was, at 0.72 of the width, it and the companion would stack into one
// column and leave half the frame carrying nothing.
func compose(w int, catW int, mirror bool) layout {
	const margin = 2
	if !mirror {
		return layout{
			CatX: 5, BubbleX: 12,
			SandFrom: 5 + catW + 2, SandTo: w - margin,
			MoonX: 0.72, Mirror: false,
		}
	}
	catX := w - catW - margin
	if catX < 0 {
		// Narrower than the sprite. Pin it to the left edge rather than let it
		// slide off: clipping the far side costs the tail, clipping the near
		// side would start eating the face, and the face is the whole point.
		catX = 0
	}
	bx := catX - 2
	if bx < margin {
		bx = margin
	}
	return layout{
		CatX: catX, BubbleX: bx,
		SandFrom: margin, SandTo: catX - 1,
		MoonX: 0.28, Mirror: true,
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

// drawScene paints one composed frame: the companion, its litter, the bubble
// and the sand. The live loop and the mockup both go through here, so a change
// to the composition cannot land in one and miss the other.
func drawScene(c *canvas.Canvas, sh *scape.Shore, cat *companion.Cat, lay layout,
	st reduce.State, t float64, seed int64, top int) {
	cat.Draw(c.Near(), lay.CatX, top, t, st.Pose)
	if st.Kittens > 0 {
		cat.DrawKittens(c.Near(), c.Mid(), lay.CatX, top, st.Kittens, c.W-1,
			int(float64(c.H)*0.42)+1, t, seed)
	}
	if st.Bubble != "" {
		rows := companion.Bubble(st.Bubble)
		x := lay.BubbleX
		if lay.Mirror {
			if w := bubbleWidth(rows); x-w >= 0 {
				x -= w
			} else {
				x = 0
			}
		}
		(&companion.Sprite{Rows: rows, Body: bubbleCol}).Draw(c.Near(), x, top-len(rows))
	}
	drawSand(c, st.Tail, sh.SandTop(), lay.SandFrom, lay.SandTo)
}

func bubbleWidth(rows []string) int {
	w := 0
	for _, r := range rows {
		if n := len([]rune(r)); n > w {
			w = n
		}
	}
	return w
}

// drawSand writes the activity tail into the beach.
//
// Sand tones rather than terminal tones, newest brightest, older lines fading
// as the tide takes them. Always visible, but scenery: the sea says how much
// work is happening and the sand says what it is, which is why no per-tool
// weather vocabulary is needed anywhere else.
//
// The block hangs off the WATERLINE, not off the bottom of the canvas. Anchored
// to the canvas it drifts upward as lines accumulate and ends up written across
// open water, which is the opposite of what "written in the sand" means.
func drawSand(c *canvas.Canvas, lines []reduce.Line, sandTop, xFrom, xTo int) {
	if len(lines) == 0 || xTo-xFrom < 12 {
		return
	}
	sand := term.RGB{R: 76, G: 65, B: 54}
	ink := term.RGB{R: 240, G: 230, B: 210}
	bad := term.RGB{R: 244, G: 176, B: 96}

	// However many rows the beach actually has. A short pane -- a wide bottom
	// split is the obvious case -- has almost no beach, and writing six lines
	// into two rows of sand would just put four of them back in the sea.
	room := c.H - sandTop
	if room < 1 {
		return
	}
	if len(lines) > room {
		lines = lines[len(lines)-room:]
	}

	// Newest at the bottom: furthest from the water, nearest the companion,
	// and last in the reading order.
	top := c.H - len(lines)
	for i, ln := range lines {
		row := top + i
		if row < 0 || row >= c.H {
			continue
		}
		base := ink
		if ln.Bad {
			base = bad
		}
		// Fade on the line's own age rather than on its position, so a line
		// that has sat there two minutes looks it even when nothing newer has
		// arrived to push it down.
		col := term.Lerp(base, sand, 0.15+0.7*ln.Age)
		x := xFrom
		for _, r := range []rune(companion.NarrowOnly(ln.Text)) {
			if x >= xTo {
				break
			}
			c.Near().Plot(x, row, r, col, 1)
			x++
		}
	}
}
