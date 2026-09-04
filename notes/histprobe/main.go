// histprobe measures ONE fact the scrollback plan rests on: does Terminal.app
// put lines that scroll out of a row-1 DECSTBM band on the ALTERNATE screen
// into the tab's history, where the user can scroll to them?
//
// Session 13 concluded "only the main screen has history" and priced a whole
// scrollback implementation on it. That premise was never measured on the
// alternate screen -- and the owner's own report ("scrolling up through the
// text history, I reach a black point") was made on an alternate-screen
// session, which is only possible if that history exists.
//
// It is read back with `history of selected tab` over AppleScript, the same
// instrument the 2026-09-01 main-screen measurement used, so the two are
// directly comparable. `-main` is that control. Production's configuration
// otherwise: alternate screen, DECSTBM band pinned at the top, origin mode on.
package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
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

const lines = 40

func main() {
	w, h := termSize()
	if w == 0 {
		fmt.Println("run this in a Terminal.app window")
		return
	}
	alt := true
	hold := 8 * time.Second
	for _, a := range os.Args[1:] {
		if a == "-main" {
			alt = false
		}
	}
	band := h - h*9/20 // agent rows, the way host.Band does it

	if alt {
		fmt.Print("\x1b[?1049h")
	}
	// The rows below the band carry a marker, so a history read can tell a
	// scrolled-off band line from a scape row.
	for r := band + 1; r <= h; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2K~~~ SCAPE ROW %02d ~~~", r, r)
	}
	fmt.Printf("\x1b[1;%dr\x1b[?6h\x1b[H", band) // the band, origin mode, home
	// Forty lines through a band shorter than that: the excess scrolls out of
	// the top of the region, which is the only way a band line ever leaves.
	for i := 1; i <= lines; i++ {
		tag := fmt.Sprintf("HIST LINE %02d", i)
		fmt.Printf("%s %s\r\n", tag, strings.Repeat(".", max(0, w-len(tag)-8)))
	}
	fmt.Printf("%dx%d band 1..%d %s: %d lines written, %d scrolled off. Holding %s.",
		w, h, band, map[bool]string{true: "ALT", false: "MAIN"}[alt], lines, lines-band+1, hold)
	time.Sleep(hold)
	fmt.Print("\x1b[?6l\x1b[r")
	if alt {
		fmt.Print("\x1b[?1049l")
	}
	fmt.Printf("\nhistprobe done (%s)\n", map[bool]string{true: "ALT", false: "MAIN"}[alt])
	time.Sleep(hold)
}
