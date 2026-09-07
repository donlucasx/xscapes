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
//
// 2026-09-07, the DL arm. His report of a stale strip down the far right at
// 143x62, measured to columns 144-145 of a 143-column terminal -- the window's
// own inset, holding retained cells. Erase cannot reach them (item 2 of
// width-audit.md) and neither can a repaint at the narrow width (item 3), and
// no column-addressed sequence can even name them. The one mechanism left that
// could plausibly drop them is one that REALLOCATES row storage rather than
// erasing cells: DL. Rows 10-20 are deleted inside a scroll region and
// repainted at the narrow width; rows 1-4, 8-9 and 21+ are the control and must
// still show their tails when the window widens again, or the read-back is not
// measuring anything.
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

const (
	dlTop = 10 // the DL region, 1-based and inclusive
	dlBot = 20
)

func paint(w, h int) string {
	var b strings.Builder
	for r := 1; r <= h; r++ {
		letter := rune('A' + (r-1)%26)
		body := strings.Repeat(string(letter), w-4)
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[48;5;%dmR%02d%s<\x1b[0m", r, 16+(r%6)*36, r, body)
	}
	return b.String()
}

func main() {
	if _, _, err := size(os.Stdout.Fd()); err != nil {
		fmt.Fprintln(os.Stderr, "widthprobe: not a terminal:", err)
		os.Exit(1)
	}
	os.Stdout.WriteString("\x1b[?1049h\x1b[H\x1b[2J")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	quit := func() { os.Stdout.WriteString("\x1b[?1049l") }

	// The driver sets the WIDE size first, so the picture is painted late
	// enough to be painted at it.
	first := time.After(3 * time.Second)
	// At 8s the driver has narrowed. Fire every arm at the new width.
	arms := time.After(8 * time.Second)

	for {
		select {
		case <-sig:
			quit()
			return
		case <-time.After(90 * time.Second):
			quit()
			return
		case <-first:
			w, h, _ := size(os.Stdout.Fd())
			os.Stdout.WriteString(paint(w, h))
			fmt.Fprintf(os.Stdout, "\x1b[%d;%dH", h/2, w/2)
			first = nil
		case <-arms:
			nw, nh, _ := size(os.Stdout.Fd())
			var e strings.Builder
			// The 2026-09-05 arms, kept: erase-line and a narrow repaint.
			e.WriteString("\x1b[5;1H\x1b[0m\x1b[2K")
			e.WriteString("\x1b[6;3H\x1b[0m\x1b[K")
			fmt.Fprintf(&e, "\x1b[7;1H\x1b[48;5;28mR07%s<\x1b[0m", strings.Repeat("g", nw-4))
			// The DL arm. DECSTBM homes the cursor, so the row is set after
			// the region, not before. Deleting the whole region replaces every
			// row in it with a NEW blank row -- if Terminal.app allocates that
			// row at the visible width, the retained tail cannot survive it.
			if dlBot <= nh {
				fmt.Fprintf(&e, "\x1b[%d;%dr", dlTop, dlBot)
				fmt.Fprintf(&e, "\x1b[%d;1H", dlTop)
				fmt.Fprintf(&e, "\x1b[%dM", dlBot-dlTop+1)
				e.WriteString("\x1b[r") // release the region before repainting
				for r := dlTop; r <= dlBot; r++ {
					fmt.Fprintf(&e, "\x1b[%d;1H\x1b[48;5;90mR%02d%s<\x1b[0m", r, r, strings.Repeat("d", nw-4))
				}
			}
			fmt.Fprintf(&e, "\x1b[%d;%dH", nh/2, nw/2)
			os.Stdout.WriteString(e.String())
			arms = nil
		}
	}
}
