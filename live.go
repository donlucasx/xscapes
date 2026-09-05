package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/canvas"
	"github.com/donlucasx/xscapes/internal/companion"
	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/reduce"
	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
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

	f := newFrames(w, h, seed, ascii, mirror, ctxUsed, tod)

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
			fmt.Fprintf(os.Stderr, "xscapes: a scape is already following session %s\n", event.Short(session))
			os.Exit(1)
		case err != nil:
			// Not fatal. A scape that cannot bind is still a scape; it just
			// runs the demo instead of going dark, and says why.
			fmt.Fprintln(os.Stderr, "xscapes:", err)
			return false
		}
		f.follow(b, reduce.New(session))
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
	defer func() {
		if f.bus != nil {
			f.bus.Close()
		}
	}()

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
		if f.bus != nil {
			f.bus.Close()
		}
		restore()
		os.Exit(0)
	}()

	fmt.Print("\x1b[?1049h\x1b[?25l\x1b[2J") // alt screen, hide cursor, clear once
	defer restore()

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

	var nextBind time.Time
	for range tick.C {
		now := time.Now()

		// Drain any pending signals so they do not queue up during a drag.
		for drained := true; drained; {
			select {
			case <-winch:
			default:
				drained = false
			}
		}
		if wIn <= 0 || hIn <= 0 {
			if nw, nh := termSize(); nw != f.c.W || nh-1 != f.c.H {
				// Only clear when the window got SMALLER. Growing needs no
				// clear -- the next frame paints strictly more cells than the
				// last one -- and clearing anyway is a blank flash on every
				// step of a drag.
				if f.resize(nw, nh-1) {
					fmt.Print("\x1b[2J")
				}
			}
		}

		// -await: keep looking for the session. The launcher starts the scape
		// and the agent at the same moment, and the agent cannot announce
		// itself until it is up, so a scape that gave up at startup would run
		// the demo beside a live session for the rest of the day.
		if !f.following() && await && now.After(nextBind) {
			nextBind = now.Add(time.Second)
			if cur := event.Current(); cur != "" && cur != stale {
				bind(cur)
			}
		}

		// Home the cursor and overwrite rather than clearing: clearing every
		// frame is what makes a terminal animation flicker.
		fmt.Print("\x1b[H" + f.frame(now))
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
	// The companion's own margin from the frame edge grows with the width:
	// 2 at 40 columns, 4 at 80, 5 at 124. His report at 124: "the companion
	// feels too pushed to the side of the screen. a bit". Two columns at
	// every width was right for a narrow pane and cramped in a wide one.
	right := margin + w/32
	if !mirror {
		return layout{
			CatX: 5, BubbleX: 12,
			SandFrom: 5 + catW + 2, SandTo: w - margin,
			MoonX: 0.72, Mirror: false,
		}
	}
	catX := w - catW - right
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
	phase := int(t/8) % 4
	switch phase {
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
	// A five-item checklist filling up as the cycle runs, so the newest channel
	// is in the demo rather than only in a live session. It matters more here
	// than the others: TodoWrite has been called ZERO times in the whole
	// recorded history, so for now this is the only place the stars light.
	st.Act.TodoTotal = 5
	st.Act.TodoDone = phase + 1
	if phase == 3 {
		st.Act.TodoDone = 5
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

// The context readout. The moon carries context as phase and altitude; the
// number under it says the same thing in digits once it starts to matter.
// Silent before ReadoutFrom (his ruling 2026-09-05: "the readout should show
// up when context hits 40% used" -- the session-6 study had said 65%), a dim
// figure of what is LEFT from there, and a warm "NN% left" from ReadoutWarn.
// The reveal itself is the first warning, and it teaches what the moon means
// at the moment it starts to matter.
const (
	ReadoutFrom = 0.40
	ReadoutWarn = 0.85
)

// drawReadout puts the number under the moon, kept inside the sky: with the
// moon at the waterline the label would otherwise land in the sea.
func drawReadout(c *canvas.Canvas, sh *scape.Shore, used float64) {
	if used < ReadoutFrom {
		return
	}
	mx, my := sh.MoonPos()
	pct := fmt.Sprintf("%.0f%%", (1-used)*100)
	txt, col := pct, moonLabelDim
	if used >= ReadoutWarn {
		txt, col = pct+" left", moonLabelWarn
	}
	hy := int(float64(c.H)*0.42) + 1
	y := my + 3
	if y > hy-1 {
		y = hy - 1
	}
	x := mx - len(txt)/2
	if x < 1 {
		x = 1
	}
	if x+len(txt) > c.W-1 {
		x = c.W - 1 - len(txt)
	}
	label(c, x, y, txt, col)
}

// drawScene paints one composed frame: the companion, its litter, the bubble
// and the sand. The live loop and the mockup both go through here, so a change
// to the composition cannot land in one and miss the other.
func drawScene(c *canvas.Canvas, sh *scape.Shore, cat *companion.Cat, lay layout,
	st reduce.State, t float64, seed int64, top int) {
	drawReadout(c, sh, st.Act.ContextUsed)
	cat.Draw(c.Near(), lay.CatX, top, t, st.Pose)
	if st.Kittens > 0 {
		// Swimmers stay above the shore's mean waterline with a row to spare
		// for the swell's crests; the waterline moves with activity, so this
		// is read off the shore every frame rather than derived from height.
		cat.DrawKittens(c.Near(), c.Mid(), lay.CatX, top, st.Kittens, c.W-1,
			int(float64(c.H)*0.42)+1, sh.SandTop()-2, t, seed)
	}
	if len(st.KittenExits) > 0 {
		// Finished subagents swim off along the top lane; same water bounds.
		cat.DrawKittenExits(c.Near(), st.KittenExits, lay.CatX, top, c.W-1,
			int(float64(c.H)*0.42)+1, sh.SandTop()-2, t, seed)
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
	//
	// Sampled from the PAINTED background of each row rather than from the
	// palette's nominal sand, because those two stopped being the same thing
	// the moment the beach could fall away toward black. Measured at midday
	// with a full fade, the nominal sand still read 175 and chose dark ink,
	// while the row it landed on had already sunk to 62: dark on dark, and the
	// newest line -- the lowest one, the one that matters most -- was the worst
	// hit. It is also more correct without the fade, since the top of the beach
	// is wet and darker than the bottom.
	bad := term.RGB{R: 244, G: 176, B: 96}
	beachAt := func(row int) term.RGB {
		if row < 0 || row >= c.H || len(c.BG) < c.W*c.H {
			return sand
		}
		var r, g, b, n int
		for x := xFrom; x < xTo && x < c.W; x += 4 {
			p := c.BG[row*c.W+x]
			r, g, b, n = r+int(p.R), g+int(p.G), b+int(p.B), n+1
		}
		if n == 0 {
			return sand
		}
		return term.RGB{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n)}
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
		beach := beachAt(row)
		base := term.RGB{R: 244, G: 236, B: 220}
		if luma(beach) > 140 {
			base = term.RGB{R: 34, G: 26, B: 20}
		}
		if ln.Bad {
			base = bad
		}
		// Fade on the line's own age rather than on its position, so a line
		// that has sat there two minutes looks it even when nothing newer has
		// arrived to push it down. It stops at 0.72 rather than 1: the tide
		// takes a line by TTL, not by fading it into unreadability first.
		col := term.Lerp(base, beach, 0.10+0.62*ln.Age)
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
