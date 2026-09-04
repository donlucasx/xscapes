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

	"github.com/donlucasx/xscapes/internal/envx"
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

	// trace, when XSCAPES_TRACE is set, records every byte sent to the
	// terminal; traceLog records size changes against byte offsets so a replay
	// can resize where the real terminal did.
	trace    *os.File
	traceLog *os.File
	traceN   int64
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
	if h.trace != nil {
		n, _ := h.trace.Write([]byte(s))
		h.traceN += int64(n)
	}
	h.mu.Unlock()
}

// traceSize records a window size against the byte offset it took effect at.
//
// Without it a replay has to feed the whole stream into one fixed-size screen,
// which misreconstructs everything written BEFORE the resize -- and the resize
// is the thing under investigation. The sidecar keeps the trace itself a pure
// byte stream, so it stays replayable by anything.
func (h *Host) traceSize(cols, rows, agentRows int) {
	if h.trace == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.traceLog == nil {
		f, err := os.OpenFile(h.trace.Name()+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return
		}
		h.traceLog = f
	}
	fmt.Fprintf(h.traceLog, "%d %d %d %d\n", h.traceN, cols, rows, agentRows)
}

// openTrace tees everything the host sends the terminal into a file, when
// XSCAPES_TRACE names one.
//
// It exists because a screenshot is not evidence of what was SENT, and the
// difference between "the terminal moved it" and "we erased it" is not visible
// in a photograph. Replaying a trace through internal/host's screen model
// reconstructs the exact screen, which is the only way to argue about a
// resize that happened on someone else's machine.
//
// The agent's own bytes are in here too -- they pass through write on their
// way out -- so a trace CAN contain whatever the agent had on screen. It is a
// debugging tool that must be pointed at a file deliberately, never a default.
func (h *Host) openTrace() {
	path := envx.Lookup("TRACE")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return
	}
	h.trace = f
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
	h.openTrace()
	if h.trace != nil {
		defer h.trace.Close()
		defer func() {
			if h.traceLog != nil {
				h.traceLog.Close()
			}
		}()
	}

	cols, rows := h.Size()
	agentRows, scapeRows := BandWith(rows, h.ScapeRows)
	h.traceSize(cols, rows, agentRows)

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

				h.traceSize(cols, rows, agentRows)
				h.write(resizeSequence(h.AltScreen, oldRows, rows, oldAgent, agentRows))
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
//
// ⚠ The SGR reset is not tidiness, it is the erase's argument. EL does not
// write spaces; it fills with the CURRENT background (BCE), and under reverse
// video that is the foreground. Every caller here clears on a timer -- the
// resize tick -- which fires with no relation to where the agent is in its
// output, and Claude Code paints backgrounds constantly: its input box, the
// context bar, a selection. So the colour in force is not the host's to assume,
// and a clear that inherited it turned the rows it touched into a solid band of
// whatever Claude happened to be drawing. Those rows then scroll out of the
// band into scrollback, which is the wall of black he hit scrolling up.
//
// Safe inside Rebind and clearRows because both bracket this with DECSC/DECRC,
// and DECRC restores the agent's SGR along with its cursor.
//
// That last clause was challenged as an unmeasured assumption about
// Terminal.app, and it cannot be measured directly -- Terminal.app does not
// answer DECRQSS (probed 2026-09-03; the probe's control asked with a known
// rendition in force and got nothing back, so this is "unsupported", not
// "default"). But the running system already settles it: BeginPaint/EndPaint
// bracket the scape's own writes the same way, and every scape line ends in
// ESC[0m, twelve times a second. If DECRC did not restore SGR, the agent's
// rendition would be reset 12x/s and Claude Code's UI would be colourless.
// It is not. The inference is from behaviour in hand rather than a spec.
func clearRowsBare(first, last int) string {
	if first > last {
		return ""
	}
	var b strings.Builder
	b.WriteString(originOff + regionReset + "\x1b[0m")
	for r := first; r <= last; r++ {
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", r)
	}
	return b.String()
}

// resizeSequence is what the host sends the terminal when the window changed
// size: it moves the band to its new geometry, undoes whatever the terminal did
// to the rows, clears the rows that changed hands, and leaves the agent's
// cursor where the agent believes it is. One function, so the tests feed the
// exact bytes production emits.
//
// Not the whole screen. Claude Code cannot redraw a transcript that has
// scrolled, so clearing the lot wiped the session's history off the screen
// until the next turn happened to redraw it. Measured both ways: whole-screen
// loses the transcript, this does not.
//
// The band is re-pinned BEFORE the pty is resized (see Run), because the
// child's repaint arrives with the SIGWINCH and has to land in the new region
// rather than the one it is leaving.
//
// ⚠ The clear has to allow for the TERMINAL having moved things. Measured by
// eye on 2026-09-03 with notes/contentprobe, in production's configuration,
// after two earlier readings got this backwards: **Terminal.app anchors
// content to the BOTTOM edge.** A shrink of N pulls every row UP by N and loses
// the top ones; a grow of N pushes every row DOWN by N and inserts blanks at
// the top. Both buffers, both directions. The CURSOR does not move either way
// -- the trap that cost this project most of a day: notes/anchorprobe parks the
// cursor and reads it back with DSR, so it reported "anchored top" for a screen
// whose content was anchored bottom.
func resizeSequence(alt bool, oldRows, rows, oldAgent, agentRows int) string {
	// So the clear range starts where the old scape's first row LANDS.
	drop := oldRows - rows
	if drop < 0 {
		drop = 0
	}
	// And on a GROW the push is undone before anything is painted, because the
	// rows it pushed down are the agent's and the scape is about to paint over
	// the ones that crossed the boundary. Alt screen only: that is where it was
	// measured, and the main screen's grow pulls real scrollback back in, which
	// is content a scroll would push away again.
	grow := 0
	if alt && rows > oldRows {
		grow = rows - oldRows
	}
	from := oldAgent + 1 - drop
	if from < 1 {
		from = 1
	}
	to := min(agentRows, oldRows-drop)
	// A SHRINK on the alternate screen is the one case where the cursor, not
	// just the rows, has to be put back -- see RebindShrinkAlt.
	if alt && rows < oldRows {
		return RebindShrinkAlt(oldRows-rows, oldAgent-agentRows, agentRows)
	}
	return Rebind(grow, from, to, agentRows)
}
