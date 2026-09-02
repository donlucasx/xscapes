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
// runClaude is `xscapes claude`. It hosts the agent inside the scape, which is
// what that command means now; -beside gives the older tmux layout, which still
// works and is still the right answer if you want the agent in its own pane.
func runClaude(args []string) {
	for i, a := range args {
		if a == "--" {
			break
		}
		if a == "-beside" || a == "--beside" {
			rest := append(append([]string{}, args[:i]...), args[i+1:]...)
			runClaudeLauncher(rest)
			return
		}
	}
	runInside(args, "claude")
}

// runInside hosts a command inside the scape. With agent set, the trailing
// arguments belong to that agent; without it, they are the command to run.
func runInside(args []string, agent string) {
	fs := flag.NewFlagSet("inside", flag.ExitOnError)
	seed := fs.Int64("seed", 7, "scene seed")
	fps := fs.Float64("fps", 12, "scape frames per second")
	ascii := fs.Bool("ascii", false, "ASCII glyphs only, no Unicode")
	mirror := fs.Bool("mirror", true, "companion on the right")
	tod := fs.Float64("tod", 0, "pin the time of day, 0..1 (0 = the wall clock)")
	ctxUsed := fs.Float64("ctx", 0, "pin context used, 0..1 (0 = the session's own)")
	dry := fs.Bool("print", false, "print what would run and how the window splits, then exit")
	scapeH := fs.Int("scape", 0, "rows to give the scape (0 = two fifths of the window)")
	alt := fs.Bool("alt", true, "run on the alternate screen: resize-proof, but the agent's output does not go to your terminal's scrollback")
	fs.Usage = func() {
		if agent != "" {
			fmt.Fprintf(os.Stderr, `xscapes claude [flags] [%s arguments ...]

Runs %s in the top rows of this window with the scape below it.
Pass -beside for the older tmux layout, with the agent in its own pane.

`, agent, agent)
		} else {
			fmt.Fprint(os.Stderr, `xscapes inside [flags] [command ...]

Runs the command in the top rows of this window with the scape below it.
With no command, runs claude.

`)
		}
		fs.PrintDefaults()
	}
	fs.Parse(args)

	argv := fs.Args()
	if agent != "" {
		argv = append([]string{agent}, argv...)
	} else if len(argv) == 0 {
		argv = []string{"claude"}
	}

	cols, rows := termSize()
	agentRows, scapeRows := host.BandWith(rows, *scapeH)
	if *dry {
		fmt.Printf("window   %dx%d\n", cols, rows)
		fmt.Printf("agent    rows 1-%d   %s\n", agentRows, strings.Join(argv, " "))
		fmt.Printf("screen   %s\n", map[bool]string{true: "alternate (no scrollback; resize-proof)", false: "main (scrollback kept; resize displaces the agent)"}[*alt])
		if scapeRows > 0 {
			fmt.Printf("scape    rows %d-%d\n", agentRows+1, rows)
		} else {
			fmt.Printf("scape    none: %d rows are needed and the window has %d\n",
				host.MinAgentRows+host.MinScapeRows, rows)
		}
		return
	}
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
		Cmd:       exec.Command(argv[0], argv[1:]...),
		Size:      termSize,
		FPS:       *fps,
		ScapeRows: *scapeH,
		AltScreen: *alt,
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
