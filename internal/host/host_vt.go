//go:build ghosttyvt

package host

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/donlucasx/xscapes/internal/gvt"
)

// EngineAvailable reports whether this binary carries the Ghostty engine.
const EngineAvailable = true

// runVT hosts the agent the way tmux does: its output goes into a terminal
// emulator of the host's own (Ghostty's engine, internal/gvt), and the host
// draws the emulator's rows onto the real terminal. The agent never writes to
// the real terminal, so nothing the real terminal does on a resize (crop the
// alternate screen, move content, reset the scroll region) can reach it. On a
// resize the engine REFLOWS the agent's text and the host repaints the band.
//
// Replaces, for this path, everything the band host needed to survive the
// real terminal: the DECSTBM band, the Filter, the screen model, the
// per-terminal Rules, pull-back and the mirror.
func (h *Host) runVT() error {
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

	vt, err := gvt.New(cols, agentRows, 8<<20)
	if err != nil {
		return err
	}
	defer vt.Close()

	p, err := openPTY()
	if err != nil {
		return err
	}
	defer p.Close()
	if err := p.SetSize(cols, agentRows); err != nil {
		return fmt.Errorf("size the pty: %w", err)
	}
	saved, err := makeRaw(in.Fd())
	if err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}

	var modes mirroredModes
	var leaving sync.Once
	leave := func() {
		leaving.Do(func() {
			h.raw("\x1b[?2026l" + modes.reset() + "\x1b[0m\x1b[?7h\x1b[?25h\x1b[0 q\x1b[?1049l")
			restoreTermios(in.Fd(), saved)
		})
	}
	defer leave()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sig
		leave()
		os.Exit(1)
	}()

	waitDone := make(chan struct{})
	close(waitDone)
	go h.forwardKeys(in, p.master, nil, waitDone)

	h.Cmd.Stdin, h.Cmd.Stdout, h.Cmd.Stderr = p.slave, p.slave, p.slave
	if h.Cmd.SysProcAttr == nil {
		h.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	h.Cmd.SysProcAttr.Setsid = true
	h.Cmd.SysProcAttr.Setctty = true
	if err := h.Cmd.Start(); err != nil {
		return err
	}
	p.slave.Close()
	p.slave = nil

	// The whole screen is the host's: alternate screen, cleared, autowrap off
	// for good (every row is placed by CUP and never runs past the margin),
	// no scroll region.
	h.raw("\x1b[?1049h\x1b[r\x1b[0m\x1b[2J\x1b[?7l")

	// The agent's output goes into the engine. kick wakes the frame loop so a
	// keystroke's echo is drawn at once, not on the next scape tick.
	kick := make(chan struct{}, 1)
	fwdDone := make(chan struct{})
	go func() {
		defer close(fwdDone)
		buf := make([]byte, 32768)
		for {
			n, err := p.master.Read(buf)
			if n > 0 {
				if h.OnOutput != nil {
					h.OnOutput(buf[:n])
				}
				if reply := vt.Write(buf[:n]); len(reply) > 0 {
					p.master.Write(reply)
				}
				select {
				case kick <- struct{}{}:
				default:
				}
			}
			if err != nil {
				return
			}
		}
	}()

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

	var dmg damage
	full := true

	// frame draws the agent's changed rows, the scape when asked, and the
	// cursor, as one synchronized update.
	frame := func(scape bool) {
		var b strings.Builder
		b.WriteString("\x1b[?2026h\x1b[?25l")
		if full {
			b.WriteString("\x1b[r\x1b[0m\x1b[2J")
		}
		b.Write(vt.Render(full, 0))
		if full {
			// Rows of the band the engine has no row for (never, today) and
			// the scape's tracker both start over.
			dmg.reset()
		}
		full = false
		if scape && scapeRows > 0 && h.Paint != nil {
			if lines := h.Paint(cols, scapeRows); len(lines) > 0 {
				if len(lines) > scapeRows {
					lines = lines[:scapeRows]
				}
				for _, i := range dmg.changed(lines) {
					fmt.Fprintf(&b, "\x1b[%d;1H\x1b[0m", agentRows+1+i)
					b.WriteString(lines[i])
				}
				b.WriteString("\x1b[0m")
			}
		}
		b.WriteString(modes.sync(vt))
		c := vt.Cursor()
		cur := fmt.Sprintf("\x1b[%d;%dH", c.Y+1, c.X+1)
		if c.Visible && c.InView && c.Y < agentRows {
			cur += "\x1b[?25h"
		}
		b.WriteString(cur)

		b.WriteString("\x1b[?2026l")
		h.raw(b.String())
	}

	var lastKick time.Time
	for {
		select {
		case err := <-done:
			select {
			case <-fwdDone:
			case <-time.After(500 * time.Millisecond):
			}
			frame(false)
			leave()
			// The transcript, on the main screen, where the real terminal
			// keeps it and reflows it: what a plain session leaves behind.
			if h.Replay {
				// Each line erases the row it lands on first: Terminal.app
				// keeps main-screen rows across 1049l, and a shorter line
				// over a longer survivor would leave its tail (the band
				// host's replay does the same, host.go replay).
				t := strings.TrimRight(strings.ReplaceAll(string(vt.Transcript(true)), "\r\n", "\n"), " \n")
				if t != "" {
					var b strings.Builder
					for _, ln := range strings.Split(t, "\n") {
						b.WriteString("\r\x1b[0m\x1b[2K")
						b.WriteString(ln)
						b.WriteString("\x1b[0m\r\n")
					}
					b.WriteString("\x1b[0m\x1b[J")
					h.raw(b.String())
				}
			}
			h.muted.Store(true)
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return nil
			}
			return err
		case <-kick:
			// Coalesce a burst: at most one agent frame per 8 ms.
			if d := time.Since(lastKick); d < 8*time.Millisecond {
				time.Sleep(8*time.Millisecond - d)
				select {
				case <-kick:
				default:
				}
			}
			lastKick = time.Now()
			frame(false)
		case <-tick.C:
			nc, nr := h.Size()
			if nc != cols || nr != rows {
				cols, rows = nc, nr
				agentRows, scapeRows = BandWith(rows, h.ScapeRows)
				h.traceSize(cols, rows, agentRows)
				vt.Resize(cols, agentRows)
				p.SetSize(cols, agentRows)
				// Everything moves: the next frame clears the glass and
				// draws every row again, inside its synchronized update.
				full = true
			}
			frame(true)
		}
	}
}

// raw writes to the real terminal as is.
func (h *Host) raw(s string) { h.writeAgent(s) }

// mirroredModes copies to the real terminal the agent's modes that change
// what the real terminal SENDS (keys, mouse, paste, focus), since the agent
// reads the real terminal's input. Modes that change what is DRAWN stay in
// the engine.
type mirroredModes struct {
	// mu: reset runs from the signal goroutine while the frame loop syncs
	// (Kimi, 2026-11-09: a concurrent map read and write is a fatal error,
	// on exactly the path that restores the terminal).
	mu     sync.Mutex
	on     map[int]bool
	kitty  int
	keypad bool
}

var inputModes = []int{1, 1000, 1002, 1003, 1004, 1005, 1006, 1007, 1015, 1016, 2004}

func (m *mirroredModes) sync(vt *gvt.Term) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.on == nil {
		m.on = map[int]bool{}
	}
	var b strings.Builder
	for _, mode := range inputModes {
		v := vt.Mode(mode, false)
		if v != m.on[mode] {
			m.on[mode] = v
			if v {
				fmt.Fprintf(&b, "\x1b[?%dh", mode)
			} else {
				fmt.Fprintf(&b, "\x1b[?%dl", mode)
			}
		}
	}
	if kp := vt.Mode(66, false); kp != m.keypad {
		m.keypad = kp
		if kp {
			b.WriteString("\x1b=")
		} else {
			b.WriteString("\x1b>")
		}
	}
	if k := vt.KittyFlags(); k != m.kitty {
		m.kitty = k
		fmt.Fprintf(&b, "\x1b[=%d;1u", k)
	}
	return b.String()
}

func (m *mirroredModes) reset() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var b strings.Builder
	for mode, v := range m.on {
		if v {
			fmt.Fprintf(&b, "\x1b[?%dl", mode)
		}
	}
	if m.keypad {
		b.WriteString("\x1b>")
	}
	if m.kitty != 0 {
		b.WriteString("\x1b[=0;1u")
	}
	return b.String()
}
