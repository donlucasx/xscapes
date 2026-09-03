// anchorprobe measures ONE fact xscapes currently guesses at: when a
// Terminal.app window shrinks while the ALTERNATE screen is up, does the
// terminal keep the TOP of the screen or the BOTTOM?
//
// It measures itself rather than asking a person to eyeball a row number
// mid-drag, and it measures the CURSOR rather than the cells, because a
// terminal moves the cursor with the content it moves. Park the cursor at a
// known row, let the window shrink, then ask where the cursor is with DSR:
//
//	kept BOTTOM (content scrolls up by the rows lost) -> parked - lost
//	kept TOP    (content truncated, cursor clamped)   -> min(parked, newHeight)
//
// Like Claude Code, it emits NOTHING between the paint and the measurement, so
// whatever happens is the terminal's doing alone.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type winsize struct{ rows, cols, xpixel, ypixel uint16 }

// termSize uses TIOCGWINSZ, not `stty size`: stty reads the size from ITS
// stdin, and a child from exec.Command inherits /dev/null unless told
// otherwise, so it silently reports the fallback every time.
func termSize() (int, int) {
	fds := []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()}
	if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
		defer f.Close()
		fds = append(fds, f.Fd())
	}
	for _, fd := range fds {
		var ws winsize
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
			uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
		if errno == 0 && ws.cols >= 8 && ws.rows >= 4 {
			return int(ws.cols), int(ws.rows)
		}
	}
	return 0, 0
}

// stty needs the terminal on ITS stdin, which is the whole lesson above.
func stty(args ...string) (string, error) {
	c := exec.Command("stty", args...)
	c.Stdin = os.Stdin
	out, err := c.Output()
	return strings.TrimSpace(string(out)), err
}

// cursorRow asks the terminal where the cursor is (DSR 6) and parses the
// CPR reply, ESC [ row ; col R.
func cursorRow() (int, error) {
	fmt.Print("\x1b[6n")
	buf := make([]byte, 1)
	var sb strings.Builder
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		sb.WriteByte(buf[0])
		if buf[0] == 'R' {
			s := sb.String()
			i := strings.LastIndex(s, "\x1b[")
			if i < 0 {
				return 0, fmt.Errorf("no CSI in reply %q", s)
			}
			body := s[i+2 : len(s)-1]
			parts := strings.Split(body, ";")
			return strconv.Atoi(parts[0])
		}
	}
	return 0, fmt.Errorf("no cursor report within 2s")
}

func main() {
	out := "/tmp/anchorprobe.verdict"
	useAlt := true
	region := false
	for _, a := range os.Args[1:] {
		if a == "-region" {
			// ⚠ THE CONFIGURATION PRODUCTION ACTUALLY RUNS IN: a DECSTBM
			// scroll region pinned to the top of the screen plus origin mode.
			// A resize with an active region is a different path in a terminal
			// -- the region has to be remapped across it -- so a measurement
			// taken without one says nothing about this case.
			region = true
			continue
		}
		if a == "-main" {
			// The control: the MAIN screen has scrollback, so if the two
			// differ this is where the difference shows.
			useAlt = false
			continue
		}
		out = a
	}
	report := func(s string) {
		os.WriteFile(out, []byte(s), 0o644)
		fmt.Print(s)
	}

	w, h := termSize()
	if w == 0 {
		report("FAIL: no terminal on any stream\n")
		os.Exit(1)
	}

	saved, err := stty("-g")
	if err != nil {
		report(fmt.Sprintf("FAIL: stty -g: %v\n", err))
		os.Exit(1)
	}
	if _, err := stty("raw", "-echo"); err != nil {
		report(fmt.Sprintf("FAIL: stty raw: %v\n", err))
		os.Exit(1)
	}
	restore := func() { stty(saved) }

	if useAlt {
		fmt.Print("\x1b[?1049h")
	}
	for r := 1; r <= h; r++ {
		tag := fmt.Sprintf(" ROW %02d ", r)
		mid := strings.Repeat("-", maxi(0, w-2*len(tag)))
		fmt.Printf("\x1b[%d;1H\x1b[2K%s%s%s", r, tag, mid, tag)
	}

	// Park mid-screen. The last row is useless (both hypotheses predict the
	// new bottom) and so is row 1 (both predict 1); the middle separates them.
	band := h * 2 / 3
	if region {
		// Pin the band exactly the way internal/host does.
		fmt.Printf("\x1b[1;%dr\x1b[?6h", band)
	}
	parked := h / 2
	if region && parked > band {
		parked = band / 2
	}
	fmt.Printf("\x1b[%d;1H\x1b[7m  CURSOR PARKED ON ROW %02d -- measuring, do not touch  \x1b[0m", parked, parked)
	fmt.Printf("\x1b[%d;1H", parked)

	// Silence. The window is resized from outside during this.
	time.Sleep(4 * time.Second)

	w2, h2 := termSize()
	row, cerr := cursorRow()

	if region {
		fmt.Print("\x1b[?6l\x1b[r")
	}
	if useAlt {
		fmt.Print("\x1b[?1049l")
	}
	restore()

	if cerr != nil {
		report(fmt.Sprintf("FAIL: %v\n", cerr))
		os.Exit(1)
	}

	// delta is signed: negative shrank, positive grew. Both directions have to
	// be measured. They are NOT mirror images -- a terminal can truncate the
	// bottom on a shrink and still push content DOWN on a grow.
	delta := h2 - h
	clampRow := func(v int) int { return maxi(1, mini(v, h2)) }
	predBottom := clampRow(parked + delta) // content anchored to the BOTTOM: moves with the edge
	predTop := clampRow(parked)            // content anchored to the TOP: stays put

	var verdict string
	switch {
	case delta == 0:
		verdict = "INCONCLUSIVE: the window did not change height"
	case predBottom == predTop:
		verdict = "AMBIGUOUS: both hypotheses predict the same row; use a different shrink"
	case row == predTop:
		verdict = "ANCHORED TOP -- content stayed put; blank rows appear at the BOTTOM"
	case row == predBottom:
		verdict = "ANCHORED BOTTOM -- content moved with the bottom edge by " + strconv.Itoa(delta) + " row(s)"
	default:
		verdict = "NEITHER -- the terminal did something else"
	}

	report(fmt.Sprintf(
		"screen    %s\n"+
			"region    %s\n"+
			"size      %dx%d -> %dx%d   (delta: %+d rows)\n"+
			"parked    row %d of %d\n"+
			"cursor    row %d after the resize\n"+
			"predicts  ANCHORED TOP -> %d   |   ANCHORED BOTTOM -> %d\n"+
			"VERDICT   %s\n",
		map[bool]string{true: "ALTERNATE", false: "MAIN"}[useAlt],
		map[bool]string{true: fmt.Sprintf("DECSTBM 1..%d + DECOM  (what production runs)", band), false: "none"}[region],
		w, h, w2, h2, delta, parked, h, row, predTop, predBottom, verdict))
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}
