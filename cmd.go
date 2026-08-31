package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/donlucasx/asciiscapes/internal/event"
)

// dispatch handles the subcommands. They are checked before flag.Parse so that
// `asciiscapes emit tool_start -tool Read` can have its own flag set without
// colliding with the renderer's twenty demo flags.
//
// Returns true if it handled the call.
func dispatch(args []string) bool {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return false
	}
	switch args[0] {
	case "hook":
		runHook(args[1:])
	case "emit":
		runEmit(args[1:])
	case "statusline":
		runStatusline(args[1:])
	case "install":
		runInstall(args[1:])
	case "uninstall":
		runUninstall(args[1:])
	case "replay":
		runReplay(args[1:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "asciiscapes: unknown command %q\n\n", args[0])
		usage()
		os.Exit(2)
	}
	return true
}

func usage() {
	fmt.Fprint(os.Stderr, `asciiscapes — a thinking screen for terminal agents

  asciiscapes                 render one frame
  asciiscapes -live           run the scape in this terminal

  asciiscapes install claude  add the hooks to Claude Code (prints a plan; --apply to write)
  asciiscapes uninstall claude
  asciiscapes emit <kind>     send one event by hand (for testing)
  asciiscapes replay <file>   feed a recorded event log to a running scape
  asciiscapes hook [Event]    adapter; reads a Claude Code hook payload on stdin
  asciiscapes statusline      adapter for the context moon; chains to your statusline

Run with -h for the renderer's flags.
`)
}

// runEmit sends one event from the command line. This is the whole reason the
// protocol is a protocol: you can drive the scene without an agent at all,
// which is how it gets tested and how a third adapter gets written.
func runEmit(args []string) {
	fs := flag.NewFlagSet("emit", flag.ExitOnError)
	var (
		tool    = fs.String("tool", "", "tool name")
		op      = fs.String("op", "", "op: read|write|edit|search|shell|web|subagent|todo|mcp")
		target  = fs.String("target", "", "file or subject")
		detail  = fs.String("detail", "", "short result")
		text    = fs.String("text", "", "message text")
		id      = fs.String("id", "", "correlation id for tool_start/tool_end")
		agent   = fs.String("agent", "", "subagent id")
		ms      = fs.Int64("ms", 0, "duration in milliseconds")
		frac    = fs.Float64("frac", -1, "0..1 reading (context used)")
		session = fs.String("session", "", "session id (default: $CLAUDE_CODE_SESSION_ID, else current)")
		quiet   = fs.Bool("q", false, "print nothing")
	)
	// The kind comes first and Go's flag package stops parsing at the first
	// non-flag argument, so `emit tool_start -tool Read` -- the form printed
	// in the help and the README -- silently dropped every flag and emitted an
	// empty event. Split the kind off before parsing.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(os.Stderr, "asciiscapes emit: need a kind, e.g. `asciiscapes emit tool_start -tool Read`")
		os.Exit(2)
	}
	kind := args[0]
	fs.Parse(args[1:])

	sess := *session
	if sess == "" {
		sess = event.SessionFromEnv()
	}
	if sess == "" {
		sess = event.Current()
	}

	e := event.Event{
		Kind: event.Kind(kind), Session: sess, Src: "manual",
		Tool: *tool, Op: event.Op(*op), Target: *target, Detail: *detail,
		Text: *text, ID: *id, Agent: *agent, MS: *ms,
	}
	if *frac >= 0 {
		f := *frac
		e.Frac = &f
	}

	viaSock, err := event.Emit(e)
	if *quiet {
		return
	}
	switch {
	case err != nil:
		fmt.Fprintln(os.Stderr, "asciiscapes:", err)
		os.Exit(1)
	case viaSock:
		fmt.Printf("sent %s to session %s\n", e.Kind, event.Short(sess))
	default:
		// Say so plainly. "It went to the spool" is the difference between
		// "no scape is running" and "the scape is broken", and guessing
		// wrong wastes an afternoon.
		fmt.Printf("spooled %s for session %s (no scape listening)\n", e.Kind, event.Short(sess))
	}
}

// statuslineInput is the part of Claude Code's statusline payload we want.
//
// The moon has no other source: no hook payload carries context remaining, and
// summing the transcript's usage means dividing by a context window we would
// have to hardcode -- which is measurably wrong on this machine, where the
// model in use has a 1M window and the obvious constant is 200k.
// used_percentage is computed by Claude Code against the real window size.
type statuslineInput struct {
	Context struct {
		UsedPercentage *float64 `json:"used_percentage"`
		WindowSize     int      `json:"context_window_size"`
	} `json:"context_window"`
	Session string `json:"session_id"`
}

// runStatusline is a pass-through: it reads the statusline payload, emits the
// context reading, then runs whatever statusline command the user already had
// and forwards its output verbatim.
//
// Chaining rather than replacing is not politeness. Lucas has a statusline he
// wrote; a tool that silently takes it over to drive a decoration has made a
// bad trade on his behalf.
func runStatusline(args []string) {
	defer func() { recover() }()

	in, _ := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))

	var s statuslineInput
	if json.Unmarshal(in, &s) == nil && s.Context.UsedPercentage != nil {
		sess := s.Session
		if sess == "" {
			sess = event.SessionFromEnv()
		}
		f := *s.Context.UsedPercentage / 100
		_, _ = event.Emit(event.Event{
			Kind: event.Context, Session: sess, Src: "claude", Frac: &f,
		})
	}

	// Strip the `--` separator. Without this the chain execs "--" and the
	// user's statusline vanishes -- and the instruction install prints tells
	// them to write exactly that, so following our own documentation deleted
	// the thing we promised to preserve.
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = bytes.NewReader(in)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// stdout here IS the statusline, so a silent failure leaves a blank
		// bar and no clue why. Say something short in the space it owns.
		fmt.Printf("asciiscapes: statusline chain failed: %v", err)
	}
}

// runReplay feeds a recorded log back through the bus, which is how the
// constants in internal/reduce get tuned against a real session instead of
// against a guess about one.
func runReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	speed := fs.Float64("speed", 1, "playback speed multiplier")
	session := fs.String("session", "", "session to replay into")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "asciiscapes replay: need a file")
		os.Exit(2)
	}
	b, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "asciiscapes:", err)
		os.Exit(1)
	}
	sess := *session
	if sess == "" {
		sess = event.SessionFromEnv()
	}
	if sess == "" {
		sess = event.Current()
	}

	var prev int64
	n := 0
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		e, err := event.Decode(line)
		if err != nil {
			continue
		}
		if prev > 0 && e.TS > prev && *speed > 0 {
			d := time.Duration(float64(e.TS-prev)/(*speed)) * time.Millisecond
			if d > 5*time.Second {
				d = 5 * time.Second
			}
			time.Sleep(d)
		}
		prev = e.TS
		e.Session, e.TS = sess, 0
		_, _ = event.Emit(e)
		n++
	}
	fmt.Printf("replayed %d events\n", n)
}
