// contentprobe measures whether Terminal.app moves CONTENT when the window
// grows, in the configuration xscapes actually runs in.
//
// The earlier anchorprobe measured the CURSOR and its result was written up as
// a fact about content. That was wrong: the cursor can stay at its absolute row
// while the content moves out from under it, which is exactly what the failure
// traces imply. There is no escape sequence that reads cells back from
// Terminal.app, so this one is read by eye -- one number.
//
// Production's configuration exactly: alternate screen, DECSTBM band pinned at
// the top, origin mode on.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

type winsize struct{ rows, cols, xpixel, ypixel uint16 }

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

func main() {
	w, h := termSize()
	if w == 0 {
		fmt.Println("run this in a Terminal.app window")
		return
	}
	band := h * 9 / 20
	band = h - band // agent rows, the way host.Band does it

	fmt.Print("\x1b[?1049h")                     // alternate screen
	fmt.Printf("\x1b[1;%dr\x1b[?6h", band)       // the band, and origin mode
	defer fmt.Print("\x1b[?6l\x1b[r\x1b[?1049l") // hand it all back

	// Fill every row with its own number, both edges, so the top row is
	// readable no matter how far the window is dragged.
	for r := 1; r <= h; r++ {
		tag := fmt.Sprintf(" ROW %02d ", r)
		mid := strings.Repeat("-", max(0, w-2*len(tag)))
		fmt.Printf("\x1b[?6l\x1b[%d;1H\x1b[2K%s%s%s\x1b[?6h", r, tag, mid, tag)
	}

	msg := fmt.Sprintf("  %dx%d, band 1..%d.  STRETCH THE WINDOW TALLER, then look at the TOP row.  ", w, h, band)
	fmt.Printf("\x1b[?6l\x1b[%d;1H\x1b[7m%s\x1b[0m\x1b[?6h", h/2, msg)

	bufio.NewReader(os.Stdin).ReadString('\n')

	_, h2 := termSize()
	fmt.Print("\x1b[?6l\x1b[r\x1b[?1049l")
	fmt.Printf("\n  height %d -> %d  (grew by %d)\n\n", h, h2, h2-h)
	fmt.Println("  Which row number was at the very TOP of the window after you stretched it?")
	fmt.Println()
	fmt.Println("    ROW 01           -> content did NOT move. I am wrong again.")
	fmt.Printf("    blank, or ROW %02d  -> content was PUSHED DOWN by the grow. That is the bug.\n", 1)
	fmt.Println()
	fmt.Println("  (If the top was blank, say so -- that is the same answer as 'pushed down'.)")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
