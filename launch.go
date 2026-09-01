package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// The launcher. `asciiscapes claude` is the one command that has to work on a
// machine where nothing is set up, because it is the whole product from the
// user's side: agent on the left, scape on the right, one line to start.
//
// Lucas runs Terminal.app with no tmux session of his own, so the launcher
// cannot assume it is already inside one -- it bootstraps tmux when there is
// none, joins the current window when there is, and says something useful
// rather than failing when tmux is missing entirely.
//
// It never takes the agent's terminal. The agent gets the pane the user typed
// into and keeps stdin, stdout and its own TTY; the scape is what moves.

// scapeWidth is the scape's share of the window. The design target is 80
// columns and it degrades to 40, so on a typical 200-column terminal 38% is
// about 76 -- enough for the full composition while leaving the agent the
// wider half, which is the one being read.
const scapeWidth = "38%"

func runClaudeLauncher(args []string) {
	fs := flag.NewFlagSet("claude", flag.ExitOnError)
	print := fs.Bool("print", false, "print the commands and exit, changing nothing")
	width := fs.String("width", scapeWidth, "the scape pane's share of the window")
	agent := fs.String("agent", "claude", "the agent command to run beside the scape")
	// Everything after -- belongs to the agent, not to us.
	fs.Parse(args)
	agentArgs := fs.Args()

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "asciiscapes: cannot find my own binary:", err)
		os.Exit(1)
	}

	if _, err := exec.LookPath(*agent); err != nil {
		fmt.Fprintf(os.Stderr, "asciiscapes: %s is not on your PATH\n", *agent)
		os.Exit(1)
	}

	tmux, err := exec.LookPath("tmux")
	if err != nil {
		noTmux(self, *agent, agentArgs, *print)
		return
	}

	// The scape waits for the session: both halves start at once and the
	// agent cannot announce itself until it is up.
	scape := []string{self, "-live", "-await"}

	if os.Getenv("TMUX") != "" {
		// Already inside tmux. Split this window, put the scape in the new
		// pane, and hand this pane to the agent.
		split := []string{tmux, "split-window", "-h", "-d", "-l", *width}
		split = append(split, scape...)
		run := append([]string{*agent}, agentArgs...)
		if *print {
			fmt.Println(shellJoin(split))
			fmt.Println("exec " + shellJoin(run))
			return
		}
		if out, err := exec.Command(split[0], split[1:]...).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "asciiscapes: tmux split failed: %v\n%s", err, out)
			os.Exit(1)
		}
		// Replace this process with the agent so it owns the pane outright --
		// no wrapper sitting between the user's keystrokes and the agent.
		execAgent(*agent, agentArgs)
		return
	}

	// No tmux session yet: build one. The agent goes in the first pane
	// because that is the pane the user will be typing into.
	name := "asciiscapes"
	newSess := []string{tmux, "new-session", "-d", "-s", name}
	newSess = append(newSess, *agent)
	newSess = append(newSess, agentArgs...)

	split := []string{tmux, "split-window", "-h", "-d", "-l", *width, "-t", name + ":0"}
	split = append(split, scape...)

	focus := []string{tmux, "select-pane", "-t", name + ":0.0"}
	attach := []string{tmux, "attach-session", "-t", name}

	if *print {
		for _, c := range [][]string{newSess, split, focus, attach} {
			fmt.Println(shellJoin(c))
		}
		return
	}

	// A session of this name may be left over from a previous run. Joining it
	// would drop the user into a dead agent, so take a fresh name instead of
	// killing something that might still be alive.
	if exec.Command(tmux, "has-session", "-t", name).Run() == nil {
		for i := 2; ; i++ {
			try := fmt.Sprintf("%s-%d", name, i)
			if exec.Command(tmux, "has-session", "-t", try).Run() != nil {
				name = try
				break
			}
		}
		newSess[4] = name
		split[len(split)-4] = name + ":0"
		focus[3] = name + ":0.0"
		attach[3] = name
	}

	for _, c := range [][]string{newSess, split, focus} {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "asciiscapes: %s failed: %v\n%s", c[1], err, out)
			os.Exit(1)
		}
	}
	// attach must REPLACE this process: tmux wants the terminal, and running
	// it as a child would leave two programs reading one keyboard.
	execAgent(attach[0], attach[1:])
}

// noTmux is the fallback. The brief allows exactly one dependency, tmux via
// brew, so this does not try to reinvent a pane manager: on macOS it opens a
// second Terminal window for the scape and hands this one to the agent, and
// anywhere else it says what to install.
func noTmux(self, agent string, agentArgs []string, print bool) {
	if _, err := exec.LookPath("osascript"); err != nil {
		fmt.Fprintln(os.Stderr, `asciiscapes: tmux is not installed.

  brew install tmux

Or run the two halves in two terminals yourself:

  asciiscapes -live -await`)
		os.Exit(1)
	}

	script := fmt.Sprintf(`tell application "Terminal"
	do script "%s -live -await"
	activate
end tell`, strings.ReplaceAll(self, `"`, `\"`))

	if print {
		fmt.Printf("osascript -e %q\n", script)
		fmt.Println("exec " + shellJoin(append([]string{agent}, agentArgs...)))
		return
	}
	if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "asciiscapes: could not open a second Terminal window: %v\n%s", err, out)
		fmt.Fprintln(os.Stderr, "\nInstall tmux for the side-by-side layout:\n\n  brew install tmux")
		os.Exit(1)
	}
	execAgent(agent, agentArgs)
}

// execAgent replaces this process with the agent. Exec rather than spawn: a
// wrapper process in the middle would have to shuttle signals, resizes and the
// TTY between the terminal and the agent, and the brief is explicit that the
// agent's terminal is never to be seized.
func execAgent(name string, args []string) {
	path, err := exec.LookPath(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "asciiscapes: %s is not on your PATH\n", name)
		os.Exit(1)
	}
	argv := append([]string{name}, args...)
	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "asciiscapes: could not start %s: %v\n", name, err)
		os.Exit(1)
	}
}

// shellJoin renders a command the way a person would have to type it, for
// -print. Quoting only what needs it keeps the output copy-pasteable.
func shellJoin(argv []string) string {
	out := make([]string, len(argv))
	for i, a := range argv {
		if a == "" || strings.ContainsAny(a, " \t\"'$`\\*?[]{}()|&;<>#~") {
			out[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
			continue
		}
		out[i] = a
	}
	return strings.Join(out, " ")
}
