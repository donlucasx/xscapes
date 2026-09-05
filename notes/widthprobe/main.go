// widthprobe paints a known picture on the ALTERNATE screen and waits, so a
// driver can change the window's WIDTH from outside and read the cells back.
//
// Every row is full width: "R01" at the left, a run of the row's own letter,
// and "<" in the LAST column, so a read-back shows at once whether a row
// wrapped (its tail lands at column 1 of the next row), shifted, was cut, or
// had cells pulled up from the row below when the window widened.
//
// Measured for on 2026-09-05: his Terminal.app at 123x55 after a width+height
// drag showed a six-column patch of scape cells above the band and the band's
// last column painted from rows below, and the host's bytes replay clean in
// the model -- whose width rule for the alternate screen had never been
// measured.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// size is the terminal's columns and rows, straight from the kernel.
func size(fd uintptr) (w, h int, err error) {
	var ws struct{ rows, cols, x, y uint16 }
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws))); errno != 0 {
		return 0, 0, errno
	}
	return int(ws.cols), int(ws.rows), nil
}

func main() {
	w, h, err := size(os.Stdout.Fd())
	if err != nil {
		fmt.Fprintln(os.Stderr, "widthprobe: not a terminal:", err)
		os.Exit(1)
	}
	var b strings.Builder
	b.WriteString("\x1b[?1049h\x1b[H\x1b[2J")
	for r := 1; r <= h; r++ {
		letter := rune('A' + (r-1)%26)
		body := strings.Repeat(string(letter), w-4)
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[48;5;%dmR%02d%s<\x1b[0m", r, 16+(r%6)*36, r, body)
	}
	// Park the cursor mid-screen so a rule that keys on it is visible too.
	fmt.Fprintf(&b, "\x1b[%d;%dH", h/2, w/2)
	os.Stdout.WriteString(b.String())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	// At 8s (the driver has narrowed the window by then): erase row 5 with
	// ESC[2K and row 6 from column 3 with ESC[K, and rewrite row 7 at the NEW
	// width, so a later widening shows whether an erase or an overwrite
	// reaches the cells the narrowing hid.
	erase := time.After(8 * time.Second)
	for {
		select {
		case <-sig:
			os.Stdout.WriteString("\x1b[?1049l")
			return
		case <-time.After(120 * time.Second):
			os.Stdout.WriteString("\x1b[?1049l")
			return
		case <-erase:
			nw, _, _ := size(os.Stdout.Fd())
			var e strings.Builder
			e.WriteString("\x1b[5;1H\x1b[0m\x1b[2K")
			e.WriteString("\x1b[6;3H\x1b[0m\x1b[K")
			fmt.Fprintf(&e, "\x1b[7;1H\x1b[48;5;28mR07%s<\x1b[0m", strings.Repeat("g", nw-4))
			fmt.Fprintf(&e, "\x1b[%d;%dH", h/2, w/2)
			os.Stdout.WriteString(e.String())
			erase = nil
		}
	}
}
