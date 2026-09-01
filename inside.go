package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/event"
	"github.com/donlucasx/xscapes/internal/host"
	"github.com/donlucasx/xscapes/internal/reduce"
)

// runInside runs the agent INSIDE the scape rather than beside it.
//
// His ruling: "the entire Claude experience should happen within the xscape,
// not next to it", and "the taller the window the more sand below where we can
// see what claude is working on". The tmux launcher (`xscapes claude`) put the
// two in adjacent panes; this puts the agent in the top rows of one window and
// the scape underneath, in the same window, with no seam and no tmux.
//
// The agent's band is anchored at row 1 and that is not a preference. Lines
// scrolled out of a scroll region reach the terminal's scrollback only when
// the region starts at the top of the screen -- measured in Terminal.app,
// rows 1-10 keeps every line and rows 5-14 keeps none. Anything painted above
// the agent would cost the user the ability to scroll back through its output.
func runInside(args []string) {
	fs := flag.NewFlagSet("inside", flag.ExitOnError)
	seed := fs.Int64("seed", 7, "scene seed")
	fps := fs.Float64("fps", 20, "scape frames per second")
	ascii := fs.Bool("ascii", false, "ASCII glyphs only, no Unicode")
	mirror := fs.Bool("mirror", true, "companion on the right")
	tod := fs.Float64("tod", 0, "pin the time of day, 0..1 (0 = the wall clock)")
	ctxUsed := fs.Float64("ctx", 0, "pin context used, 0..1 (0 = the session's own)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `xscapes inside [flags] [command ...]

Runs the command in the top rows of this window with the scape below it.
With no command, runs claude.

`)
		fs.PrintDefaults()
	}
	fs.Parse(args)

	argv := fs.Args()
	if len(argv) == 0 {
		argv = []string{"claude"}
	}

	cols, rows := termSize()
	_, scapeRows := host.Band(rows)
	if scapeRows == 0 {
		fmt.Fprintf(os.Stderr, "xscapes: window is %d rows; %d are needed before a scape fits. Running %s bare.\n",
			rows, host.MinAgentRows+host.MinScapeRows, argv[0])
	}

	// Remember the session pointer that is already there. The agent has not
	// started yet, so whatever it names is the last session to run -- binding
	// to it succeeds and shows a session that ended yesterday underneath a
	// live one. Take the first session that differs.
	stale := event.Current()

	fr := newFrames(cols, max(scapeRows, 1), *seed, *ascii, *mirror, *ctxUsed, *tod)
	defer func() {
		if fr.bus != nil {
			fr.bus.Close()
		}
	}()
	var nextBind time.Time

	h := &host.Host{
		Cmd:  exec.Command(argv[0], argv[1:]...),
		Size: termSize,
		FPS:  *fps,
		Paint: func(cols, rows int) []string {
			now := time.Now()
			if w, hh := fr.size(); w != cols || hh != rows {
				fr.resize(cols, rows)
			}
			// Bind to the hosted agent once its hooks announce it. Polled
			// here rather than before starting it: the agent cannot name
			// itself until it is up.
			if !fr.following() && now.After(nextBind) {
				nextBind = now.Add(time.Second)
				if cur := event.Current(); cur != "" && cur != stale {
					if b, err := event.Listen(cur); err == nil {
						fr.follow(b, reduce.New(cur))
					}
				}
			}
			return strings.Split(fr.frame(now), "\n")
		},
	}
	h.Cmd.Env = os.Environ()
	if err := h.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "xscapes:", err)
		os.Exit(1)
	}
}
