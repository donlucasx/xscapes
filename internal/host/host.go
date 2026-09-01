package host

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Host runs an agent inside the top rows of this terminal and paints the scape
// in the rows below it.
//
// It is not a terminal emulator. The agent's bytes go to the real terminal
// untouched but for one three-byte sequence (see Filter), which is what makes
// this safe: nothing the host fails to understand about Claude Code's output
// can be corrupted by it. What holds the two apart is the scroll region --
// the agent is told its window is band-sized, and DECSTBM stops it scrolling
// past the bottom of that band.
//
// Everything it relies on was measured rather than assumed; see
// notes/claude-terminal-emissions.md.
type Host struct {
	// Cmd is the agent. Its stdio is replaced with the pty.
	Cmd *exec.Cmd
	// Size reports the real terminal's size. Injected rather than measured
	// here so there is one TIOCGWINSZ in the project, not two.
	Size func() (cols, rows int)
	// Paint returns the scape's rows: exactly rows strings, each at most cols
	// wide. Nil, or a nil return, paints nothing.
	Paint func(cols, rows int) []string
	// FPS is how often the scape repaints. The agent's output is never held
	// up by it; it goes out the moment it arrives.
	FPS float64
	// In and Out default to os.Stdin and os.Stdout. Overridable for tests.
	In  *os.File
	Out *os.File

	mu  sync.Mutex // serialises writes to Out
	out *os.File
}

// write is the only path to the terminal. Serialised, because a scape frame
// landing in the middle of a chunk of the agent's output would split an escape
// sequence in half and paint garbage into the agent's own window.
func (h *Host) write(s string) {
	if s == "" {
		return
	}
	h.mu.Lock()
	io.WriteString(h.out, s)
	h.mu.Unlock()
}

func (h *Host) Run() error {
	in, out := h.In, h.Out
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	h.out = out

	cols, rows := h.Size()
	agentRows, scapeRows := Band(rows)

	p, err := openPTY()
	if err != nil {
		return err
	}
	defer p.Close()
	if err := p.SetSize(cols, agentRows); err != nil {
		return fmt.Errorf("size the pty: %w", err)
	}

	// Raw mode, so every keystroke reaches the agent as typed -- including
	// Ctrl-C, which Claude Code uses to interrupt a turn and which the host
	// must not turn into a signal of its own.
	saved, err := makeRaw(in.Fd())
	if err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	restore := func() {
		h.write(LeaveBand() + "\x1b[0m")
		restoreTermios(in.Fd(), saved)
	}
	defer restore()

	// A signal that kills the host must not leave the terminal in raw mode
	// with a scroll region pinned to the top of the screen.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sig
		restore()
		os.Exit(1)
	}()

	h.Cmd.Stdin, h.Cmd.Stdout, h.Cmd.Stderr = p.slave, p.slave, p.slave
	if h.Cmd.SysProcAttr == nil {
		h.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	h.Cmd.SysProcAttr.Setsid = true
	h.Cmd.SysProcAttr.Setctty = true
	if err := h.Cmd.Start(); err != nil {
		return err
	}
	// The child owns the slave now. Holding it open here would hide the EOF
	// when the child exits, and the output loop would never finish.
	p.slave.Close()
	p.slave = nil

	// Clear the scape's rows once, then pin the agent. Clearing first because
	// whatever the shell left there is not scenery.
	h.write(clearRows(agentRows+1, rows) + EnterBand(agentRows))

	// The agent's output, filtered and forwarded. Never buffered by the frame
	// loop: a keystroke echoes at the speed the agent produces it.
	go func() {
		var f Filter
		buf := make([]byte, 8192)
		for {
			n, err := p.master.Read(buf)
			if n > 0 {
				h.write(string(f.Filter(buf[:n])))
			}
			if err != nil {
				h.write(string(f.Flush()))
				return
			}
		}
	}()

	// Keystrokes, and the terminal's replies to the agent's startup queries
	// (device attributes, the kitty keyboard protocol, synchronized output),
	// which travel the same path and must arrive byte for byte.
	go io.Copy(p.master, in)

	done := make(chan error, 1)
	go func() { done <- h.Cmd.Wait() }()

	fps := h.FPS
	if fps < 1 {
		fps = 1
	}
	if fps > 120 {
		fps = 120
	}
	tick := time.NewTicker(time.Duration(float64(time.Second) / fps))
	defer tick.Stop()

	for {
		select {
		case err := <-done:
			// Give the terminal its screen back, below the band, so the shell
			// prompt does not land on top of the beach.
			h.write(LeaveBand() + clearRows(agentRows+1, rows) + fmt.Sprintf("\x1b[%d;1H", agentRows+1))
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return nil // the agent exiting is not the host failing
			}
			return err
		case <-tick.C:
			nc, nr := h.Size()
			if nc != cols || nr != rows {
				cols, rows = nc, nr
				agentRows, scapeRows = Band(rows)
				p.SetSize(cols, agentRows)
				// The agent repaints itself from SIGWINCH. The band has to be
				// re-stated because the region is in screen coordinates, and
				// the rows below it cleared because the old scape is still
				// sitting wherever the last size put it.
				h.write(clearRows(agentRows+1, rows) + EnterBand(agentRows))
			}
			if scapeRows <= 0 || h.Paint == nil {
				continue
			}
			lines := h.Paint(cols, scapeRows)
			if len(lines) == 0 {
				continue
			}
			var b strings.Builder
			b.WriteString(BeginPaint())
			for i, ln := range lines {
				if i >= scapeRows {
					break
				}
				fmt.Fprintf(&b, "\x1b[%d;1H", agentRows+1+i)
				b.WriteString(ln)
			}
			b.WriteString(EndPaint(agentRows))
			// One write, so the terminal sees a whole frame between the
			// cursor save and its restore.
			h.write(b.String())
		}
	}
}

// clearRows blanks rows first..last inclusive, in screen coordinates. Used
// outside the band, so it is only ever called between BeginPaint's region
// reset and EndPaint -- or before the band exists at all.
func clearRows(first, last int) string {
	if first > last {
		return ""
	}
	var b strings.Builder
	b.WriteString(saveCursor + originOff + regionReset)
	for r := first; r <= last; r++ {
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", r)
	}
	b.WriteString(restoreCursor)
	return b.String()
}
