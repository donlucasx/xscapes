// sgrprobe answers: does Terminal.app's DECRC (ESC 8) restore the SGR that
// DECSC (ESC 7) saved?
//
// xscapes' resize-clear fix depends on it. The clear writes ESC[0m before
// erasing so it cannot inherit the hosted agent's colour, and that is called
// "safe" because DECSC/DECRC bracket it and DECRC is said to put the agent's
// colour back. That is DEC/xterm behaviour; whether Terminal.app implements it
// was never measured, and it is the only terminal that matters here. If it does
// not, the reset LEAKS: after every resize the agent's rendition is silently
// default until it next emits SGR.
//
// It ASKS the terminal (DECRQSS: ESC P $ q m ESC \) rather than painting and
// having a person look. All escape I/O goes to /dev/tty, never stdout, so the
// report can be redirected without sending the queries into a file. It writes
// every step to the report as it goes: a probe that dies silently is
// indistinguishable from one that measured "nothing", which is how a wrong
// answer gets believed.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	tty *os.File
	rep strings.Builder
	out = "/tmp/sgr-answer.txt"
)

func save() { os.WriteFile(out, []byte(rep.String()), 0o644) }

func logf(f string, a ...any) {
	fmt.Fprintf(&rep, f+"\n", a...)
	save() // after every step, so a crash still leaves the trail
}

func stty(args ...string) (string, error) {
	c := exec.Command("stty", args...)
	c.Stdin = tty
	o, err := c.Output()
	return strings.TrimSpace(string(o)), err
}

func ask(q string) string {
	tty.WriteString(q)
	buf := make([]byte, 1)
	var sb strings.Builder
	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		n, err := tty.Read(buf)
		if n > 0 {
			sb.WriteByte(buf[0])
			if buf[0] == '\\' || buf[0] == 0x07 {
				break
			}
			continue
		}
		if err != nil {
			break // EOF from a VTIME expiry: nothing more is coming
		}
		if sb.Len() > 0 {
			break // a reply arrived and then stopped
		}
	}
	return sb.String()
}

func show(s string) string {
	return strings.NewReplacer("\x1b", "<ESC>", "\x07", "<BEL>").Replace(s)
}

func main() {
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	var err error
	tty, err = os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		out = "/tmp/sgr-answer.txt"
		rep.WriteString("FAIL: cannot open /dev/tty: " + err.Error() + "\n")
		save()
		return
	}
	defer tty.Close()
	logf("opened /dev/tty")

	saved, err := stty("-g")
	if err != nil {
		logf("FAIL: stty -g: %v", err)
		return
	}
	// min 0 time 3 makes read() return after 0.3s with nothing, which is how a
	// tty does timeouts. Go's SetReadDeadline does not work on a character
	// device -- the first version of this probe hung there forever and had to
	// be killed, leaving a truncated report that looked like a measurement.
	if _, err := stty("raw", "-echo", "min", "0", "time", "3"); err != nil {
		logf("FAIL: stty raw: %v", err)
		return
	}
	defer stty(saved)
	logf("raw mode on")

	// Control first. Without it, "no reply" and "reply says default" are
	// indistinguishable and the measurement is worthless.
	tty.WriteString("\x1b[44m")
	control := ask("\x1bP$qm\x1b\\")
	logf("control (blue in force, no save/restore) : %q -> %s", control, show(control))

	if strings.TrimSpace(control) == "" {
		tty.WriteString("\x1b[0m")
		logf("")
		logf("DECRQSS UNSUPPORTED: no answer with a known rendition in force.")
		logf("This probe CANNOT settle the question. Nothing was measured; infer nothing.")
		return
	}

	tty.WriteString("\x1b[44m")
	tty.WriteString("\x1b7")   // DECSC with blue in force
	tty.WriteString("\x1b[0m") // reset, exactly as the xscapes clear does
	tty.WriteString("\x1b8")   // DECRC -- should bring blue back
	after := ask("\x1bP$qm\x1b\\")
	tty.WriteString("\x1b[0m")
	logf("after   (blue, DECSC, reset, DECRC)      : %q -> %s", after, show(after))
	logf("")

	switch {
	case strings.Contains(after, "44"):
		logf("VERDICT: DECRC RESTORES SGR. The xscapes resize-clear fix is safe as claimed.")
	case strings.TrimSpace(after) == "":
		logf("VERDICT: INCONCLUSIVE -- the control answered but this did not. Infer nothing.")
	default:
		logf("VERDICT: DECRC DOES NOT RESTORE SGR here. The \"safe because DECRC restores")
		logf("         it\" comment in host.go is WRONG on this terminal: the reset leaks.")
	}
}
