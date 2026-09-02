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
	// ScapeRows fixes the scape's height. Zero takes the automatic split.
	ScapeRows int
	// AltScreen runs on the alternate screen, which has no history for the
	// terminal to pull back in when the window grows. See Open.
	AltScreen bool
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
	agentRows, scapeRows := BandWith(rows, h.ScapeRows)

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
	// One exit path, however the host ends: the child exiting, a signal, or an
	// error. Once, because the signal handler and the deferred call race, and
	// emitting the sequence twice repaints the shell's prompt over itself.
	//
	// It reads agentRows and rows as they are at the time, so a window resized
	// during the session is cleared to its real size rather than its first one.
	var leaving sync.Once
	leave := func() {
		leaving.Do(func() {
			h.write(Close(h.AltScreen, agentRows, rows))
			restoreTermios(in.Fd(), saved)
		})
	}
	defer leave()

	// A signal that kills the host must not leave the terminal in raw mode
	// with a scroll region pinned to the top of the screen.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sig
		leave()
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

	// Take the screen and pin the band. On the main screen that means clearing
	// first, because Claude Code never clears the display itself -- measured,
	// it emits no ED at all -- so it would print over whatever was already
	// there.
	h.write(Open(h.AltScreen, agentRows, rows))
	// Only what moved gets sent. See damage.
	var dmg damage

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
			leave()
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return nil // the agent exiting is not the host failing
			}
			return err
		case <-tick.C:
			nc, nr := h.Size()
			if nc != cols || nr != rows {
				oldAgent, oldRows := agentRows, rows
				cols, rows = nc, nr
				agentRows, scapeRows = BandWith(rows, h.ScapeRows)

				// Clear exactly the rows that changed hands: the ones that
				// were scape and are now inside the agent's band. Nothing
				// else.
				//
				// Not the whole screen. Claude Code repaints from SIGWINCH by
				// homing and clearing downward, but it can only redraw what it
				// still holds -- a transcript that has scrolled is gone, and
				// clearing the lot wiped the session's history off the screen
				// until the next turn happened to redraw it. Measured both
				// ways: whole-screen loses the transcript, this does not.
				//
				// The band is re-pinned BEFORE the pty is resized, because the
				// child's repaint arrives with the SIGWINCH and has to land in
				// the new region rather than the one it is leaving.
				h.write(Rebind(oldAgent+1, min(oldRows, agentRows), agentRows))
				// The screen was just touched behind the tracker's back.
				dmg.reset()
				p.SetSize(cols, agentRows)
			}
			if scapeRows <= 0 || h.Paint == nil {
				continue
			}
			lines := h.Paint(cols, scapeRows)
			if len(lines) == 0 {
				continue
			}
			if len(lines) > scapeRows {
				lines = lines[:scapeRows]
			}
			dirty := dmg.changed(lines)
			if len(dirty) == 0 {
				continue
			}
			var b strings.Builder
			b.WriteString(BeginPaint())
			for _, i := range dirty {
				fmt.Fprintf(&b, "\x1b[%d;1H", agentRows+1+i)
				b.WriteString(lines[i])
			}
			b.WriteString(EndPaint(agentRows))
			// One write, so the terminal sees a whole frame between the
			// cursor save and its restore.
			h.write(b.String())
		}
	}
}

// clearRows blanks rows first..last inclusive, in screen coordinates, and puts
// the cursor back where it was.
func clearRows(first, last int) string {
	if first > last {
		return ""
	}
	return saveCursor + clearRowsBare(first, last) + restoreCursor
}

// clearRowsBare is clearRows without the save and restore, for callers that
// own the cursor themselves. It still drops origin mode and the region,
// because rows outside the band cannot be addressed otherwise, and it leaves
// the cursor on the last row it cleared.
func clearRowsBare(first, last int) string {
	if first > last {
		return ""
	}
	var b strings.Builder
	b.WriteString(originOff + regionReset)
	for r := first; r <= last; r++ {
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", r)
	}
	return b.String()
}
