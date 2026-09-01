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
	"github.com/donlucasx/asciiscapes/internal/notify"
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
func runLive(seed int64, fps float64, wIn, hIn int, ctxUsed, tod float64, ascii bool, session string, mirror bool, await bool) {
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
	// bind attaches to a session, or reports that there is nothing to attach
	// to yet. Split out because -await calls it again every second: the
	// launcher starts the scape and the agent together, and whichever comes
	// up first must not decide the scape's fate for the rest of the session.
	bind := func(session string) bool {
		if session == "" {
			return false
		}
		b, err := event.Listen(session)
		switch {
		case err == event.ErrBusy:
			fmt.Fprintf(os.Stderr, "asciiscapes: a scape is already following session %s\n", event.Short(session))
			os.Exit(1)
		case err != nil:
			// Not fatal. A scape that cannot bind is still a scape; it just
			// runs the demo instead of going dark, and says why.
			fmt.Fprintln(os.Stderr, "asciiscapes:", err)
			return false
		}
		bus, red = b, reduce.New(session)
		return true
	}
	// -await means "the agent has not started yet; wait for it", so the
	// pointer left behind by whatever ran last is exactly the wrong thing to
	// follow. Binding to it succeeds -- Listen happily opens a socket for a
	// session that is over -- and the scape then runs beside a live agent
	// showing a session that ended yesterday. Remember it instead, and take
	// the first session that differs.
	stale := ""
	if await {
		stale = event.Current()
	} else {
		bind(session)
	}
	// Closed here rather than where it is opened: -await can bind on any
	// frame, and a defer inside that closure would close the bus the instant
	// it was created.
	defer func() {
		if bus != nil {
			bus.Close()
		}
	}()

	c := canvas.New(w, h, canvas.AlphaFar, canvas.AlphaMid, canvas.AlphaNear)
	sh := scape.NewShore(seed, ascii)
	cat := companion.NewCat()
	cat.FaceLeft(mirror)
	ccw, chh := cat.Size()
	lay := compose(w, ccw, mirror)
	sh.MoonX = lay.MoonX

	// React to a resized window by ASKING, every frame.
	//
	// The signal is still registered, because it wakes a terminal that would
	// otherwise sit idle, but the size check no longer depends on it. Dragging
	// a window edge fires dozens of SIGWINCHes and draining one per frame left
	// the canvas chasing the window a step behind, which is what tore during
	// the drag. An ioctl costs microseconds, so simply asking every frame
	// coalesces the whole drag for free.
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)

	// The alternate screen buffer gives the scene a page of its own: nothing
	// of the shell shows through, and quitting puts the terminal back exactly
	// as it was rather than leaving a beach in the scrollback.
	restore := func() { fmt.Print("\x1b[0m\x1b[?25h\x1b[?1049l") }
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

	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[2J") // alt screen, hide cursor, clear once
	defer restore()

	profile := term.DetectProfile()
	sound := notify.New()
	var knocker notify.Knocker
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
	var nextBind time.Time
	for range tick.C {
		now := time.Now()
		t := now.Sub(start).Seconds()

		// Drain any pending signals so they do not queue up during a drag.
		for drained := true; drained; {
			select {
			case <-winch:
			default:
				drained = false
			}
		}
		if wIn <= 0 || hIn <= 0 {
			if nw, nh := termSize(); nw != c.W || nh-1 != c.H {
				shrank := nw < c.W || nh-1 < c.H
				c.Resize(nw, nh-1)
				lay = compose(nw, ccw, mirror)
				sh.MoonX = lay.MoonX
				// Only clear when the window got SMALLER. Growing needs no
				// clear -- the next frame paints strictly more cells than the
				// last one -- and clearing anyway is a blank flash on every
				// step of a drag.
				if shrank {
					fmt.Print("\x1b[2J")
				}
			}
		}

		// -await: keep looking for the session. The launcher starts the scape
		// and the agent at the same moment, and the agent cannot announce
		// itself until it is up, so a scape that gave up at startup would run
		// the demo beside a live session for the rest of the day.
		if red == nil && await && now.After(nextBind) {
			nextBind = now.Add(time.Second)
			if cur := event.Current(); cur != "" && cur != stale {
				bind(cur)
			}
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

		// The audible half of the nudge. Keyed off the bubble rather than the
		// pose, and edge-detected, so a sixty-second permission nag knocks
		// once rather than once a minute.
		//
		// Only when following a real session: the demo cycles every state on a
		// timer, and a scape nobody attached to has nothing to announce.
		// `asciiscapes notify` is how the sounds get auditioned.
		if red != nil {
			if k, ring := knocker.Knock(st.Bubble, st.BubbleAsk); ring {
				sound.Play(k)
			}
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
	switch int(t/8) % 4 {
	case 1:
		st.Act = scape.Activity{ContextUsed: ctxUsed, TimeOfDay: tod}
		st.Pose = companion.Resting
	case 2:
		st.Pose = companion.NeedsYou
		st.Bubble, st.BubbleAsk = "allow Bash?", true
	case 3:
		st.Act = scape.Activity{Level: 0.3, ContextUsed: ctxUsed, TimeOfDay: tod}
		st.Pose = companion.Done
		st.Bubble = "tests passed"
	}
	return st
}

// timeOfDay maps the wall clock to the palette's 0..1, midnight at 0.
func timeOfDay(now time.Time) float64 {
	h, m, s := now.Clock()
	return float64(h*3600+m*60+s) / 86400
}

// luma is the perceived brightness of a colour, for deciding whether to write
// light on dark or dark on light.
func luma(c term.RGB) float64 {
	return 0.30*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)
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
		rows, col := companion.DoneBubble(st.Bubble), bubbleCol
		if st.BubbleAsk {
			rows, col = companion.Bubble(st.Bubble), bubbleAskCol
		}
		x := lay.BubbleX
		if lay.Mirror {
			rows = companion.MirrorTail(rows)
			if w := bubbleWidth(rows); x-w >= 0 {
				x -= w
			} else {
				x = 0
			}
		}
		// Opaque: a balloon is TEXT, and a transparent space lets the sea
		// write glyphs into the middle of the words.
		(&companion.Sprite{Rows: rows, Body: col, Opaque: true}).Draw(c.Near(), x, top-len(rows))
	}
	drawSand(c, st.Tail, sh.SandColor(), sh.SandTop(), lay.SandFrom, lay.SandTo)
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
func drawSand(c *canvas.Canvas, lines []reduce.Line, sand term.RGB, sandTop, xFrom, xTo int) {
	if len(lines) == 0 || xTo-xFrom < 12 {
		return
	}
	// The ink is derived from the beach, not fixed.
	//
	// Writing in sand is only legible by contrast with the sand, and the beach
	// changes colour all day. A pale ink is right at midnight and invisible at
	// noon, when the sand is brighter than the ink is. So pick the direction
	// from the beach -- light on a dark beach, dark on a bright one -- and let
	// age close the gap without ever closing it completely.
	bad := term.RGB{R: 244, G: 176, B: 96}
	toward := term.RGB{R: 244, G: 236, B: 220}
	if luma(sand) > 140 {
		toward = term.RGB{R: 34, G: 26, B: 20}
	}

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
		base := toward
		if ln.Bad {
			base = bad
		}
		// Fade on the line's own age rather than on its position, so a line
		// that has sat there two minutes looks it even when nothing newer has
		// arrived to push it down. It stops at 0.72 rather than 1: the tide
		// takes a line by TTL, not by fading it into unreadability first.
		col := term.Lerp(base, sand, 0.10+0.62*ln.Age)
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
