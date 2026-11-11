//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// termSize asks the kernel how big the window is, via the TIOCGWINSZ ioctl.
//
// It used to shell out to `stty size`, which is where this went wrong: stty
// reads the size from its STDIN, and a child started by exec.Command inherits
// /dev/null unless told otherwise, so it printed "stdin isn't a terminal" and
// termSize silently returned the 80x24 fallback -- every time, on every
// machine. The scene has therefore always been 80x24 no matter how big the
// window was, which is one bug wearing three costumes: a scene that does not
// fill the frame, a scene that garbles when the window is shrunk below 80
// columns (the over-long lines wrap), and a scene that never repaints on a
// resize (the "new" size always compared equal to the old one).
//
// The ioctl has no such problem, needs no subprocess, and costs microseconds
// rather than milliseconds. Ask the controlling terminal directly rather than
// any particular stream, so a redirected stdout does not blind us.
func termSize() (w, h int) {
	for _, fd := range ttyFDs() {
		if c, r, ok := winSize(fd); ok {
			return c, r
		}
	}
	return 80, 24
}

// ttyFDs lists the descriptors worth asking, most authoritative first. If all
// three standard streams are redirected, /dev/tty still reaches the terminal
// this process is attached to.
func ttyFDs() []uintptr {
	fds := []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()}
	if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
		defer f.Close()
		fds = append(fds, f.Fd())
	}
	return fds
}

// winsize mirrors struct winsize from <sys/ioctl.h>: rows, cols, then the
// pixel dimensions, which nothing reports reliably and we do not use.
type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func winSize(fd uintptr) (cols, rows int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.cols < 8 || ws.rows < 4 {
		return 0, 0, false
	}
	return int(ws.cols), int(ws.rows), true
}

// notifyResize wakes the loop on a window resize. The loop asks for the size
// every frame anyway (see runLive), so this only saves an idle terminal from
// sleeping through the change.
func notifyResize(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}
