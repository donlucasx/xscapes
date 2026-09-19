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
	"github.com/donlucasx/xscapes/internal/watch"
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
	appleTerminal := os.Getenv("TERM_PROGRAM") == "Apple_Terminal"
	history := fs.Bool("history", appleTerminal, "mirror rows that leave the agent's band into the terminal's own scrollback; in Terminal.app, which drops them at exit, the transcript and the final screen are printed again after the session (default: on in Terminal.app)")
	watchMode := fs.String("watch", "auto", "drive the scape from the program's own output when it has no hooks: auto (until an agent's hooks announce a session), on (always, hooks ignored), off (hooks only; the demo cycles until they bind)")
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
	// The agent's hooks are what the scape is for. -watch=on is the explicit
	// way to run on the output watcher instead; without it, a missing install
	// is said out loud rather than shown as a scape that never fills.
	if agent != "" && *watchMode != "on" {
		switch st, where, have, want := hooksState(agent); st {
		case hooksNone:
			fmt.Fprint(os.Stderr, missingHooks(agent, where))
			os.Exit(2)
		case hooksPartial:
			fmt.Fprint(os.Stderr, partialHooks(agent, where, have, want))
			os.Exit(2)
		}
	}

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
		fmt.Printf("screen   %s\n", map[bool]string{true: "alternate (resize-proof)", false: "main (scrollback kept; resize displaces the agent)"}[*alt])
		fmt.Printf("history  %s\n", map[bool]string{true: "rows leaving the band are mirrored into the terminal's scrollback; replayed on exit", false: "off"}[*history && *alt])
		fmt.Printf("watch    %s\n", map[string]string{
			"auto": "the program's own output drives the scape until an agent's hooks announce a session",
			"on":   "the program's own output drives the scape; hooks are ignored",
			"off":  "hooks only",
		}[*watchMode])
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

	// The agent has not started yet, so the session pointer names the last
	// session to run -- binding to it would show a session that ended
	// yesterday underneath a live one. Bind to the first pointer WRITTEN
	// after this moment, whatever it says: a resumed session writes the
	// same id again, and waiting for a different one never bound it (F1,
	// session 38; event.CurrentSince).
	launched := time.Now()

	fr := newFrames(cols, max(scapeRows, 1), *seed, *ascii, *mirror, *ctxUsed, *tod)
	defer func() {
		if fr.bus != nil {
			fr.bus.Close()
		}
	}()
	var nextBind time.Time

	// The generic adapter. Until an agent's hooks announce a session, and
	// for a program that has none at all, the scape follows the program's
	// own traffic: output is work, Enter is a prompt, quiet after work is
	// done. See internal/watch for what it can and cannot say.
	switch *watchMode {
	case "auto", "on", "off":
	default:
		fmt.Fprintf(os.Stderr, "xscapes: -watch must be auto, on or off, not %q\n", *watchMode)
		os.Exit(2)
	}
	var synth *watch.Synth
	var synthRed *reduce.Reducer
	if *watchMode != "off" {
		synth = watch.New(watch.Options{})
		synthRed = reduce.New("watch")
		fr.follow(nil, synthRed)
	}
	// unbound says the scape is not yet following a real session's hooks:
	// either nothing at all, or the generic adapter, which hooks outrank.
	unbound := func() bool { return fr.red == nil || fr.red == synthRed }

	h := &host.Host{
		Cmd:       exec.Command(argv[0], argv[1:]...),
		Size:      termSize,
		FPS:       *fps,
		ScapeRows: *scapeH,
		AltScreen: *alt,
		History:   *history && *alt,
		Replay:    *history && *alt && appleTerminal,
		Rules:     host.RulesFor(os.Getenv("TERM_PROGRAM")),
		OnOutput: func(b []byte) {
			if synth != nil {
				synth.Output(b)
			}
		},
		OnInput: func(b []byte) {
			if synth != nil {
				synth.Input(b)
			}
		},
		Paint: func(cols, rows int) []string {
			now := time.Now()
			if w, hh := fr.size(); w != cols || hh != rows {
				fr.resize(cols, rows)
			}
			if synth != nil && fr.red == synthRed {
				for _, e := range synth.Step(now) {
					synthRed.Apply(e, now)
				}
			}
			// Bind to the hosted agent once its hooks announce it. Polled
			// here rather than before starting it: the agent cannot name
			// itself until it is up. Hooks outrank the generic adapter,
			// unless -watch=on says otherwise.
			if *watchMode != "on" && unbound() && now.After(nextBind) {
				nextBind = now.Add(time.Second)
				if cur := event.CurrentSince(launched); cur != "" {
					// Since the launch: the hooks that fired between the
					// pointer and this poll spooled, and they are news
					// (a one-shot's first prompt is in that second).
					if b, err := event.ListenSince(cur, launched); err == nil {
						fr.follow(b, reduce.New(cur))
					}
				}
			}
			return strings.Split(fr.frame(now), "\n")
		},
	}
	h.Cmd.Env = os.Environ()
	fr.erases = h.AgentErases
	if err := h.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "xscapes:", err)
		os.Exit(1)
	}
}
