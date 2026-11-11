//go:build windows

package main

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

// THE WINDOW SIZE ON WINDOWS.
//
// There is no ioctl and no SIGWINCH: the console API reports the size, and a
// resize arrives as an INPUT_RECORD the program has to read stdin to see --
// which the hosted agent owns, so the host must not be reading it. So
// notifyResize polls instead, at a tenth of the scape's frame rate, and the
// loop asks for the size itself every frame regardless (that is what makes a
// window drag coalesce on every platform).

var procGetConsoleScreenBufferInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleScreenBufferInfo")

// smallRect is SMALL_RECT as the API lays it out.
type smallRect struct {
	Left, Top, Right, Bottom int16
}

// consoleScreenBufferInfo is CONSOLE_SCREEN_BUFFER_INFO.
type consoleScreenBufferInfo struct {
	Size              struct{ X, Y int16 }
	CursorPosition    struct{ X, Y int16 }
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize struct{ X, Y int16 }
}

// winSize reads the visible window's size, not the whole buffer: the buffer
// grows when the user scrolls, and a scape painted to the buffer's width
// would be wrong the moment the console has a scrollbar.
func winSize(fd uintptr) (cols, rows int, ok bool) {
	var info consoleScreenBufferInfo
	r, _, _ := procGetConsoleScreenBufferInfo.Call(fd, uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 0, 0, false
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	h := int(info.Window.Bottom-info.Window.Top) + 1
	if w < 8 || h < 4 {
		return 0, 0, false
	}
	return w, h, true
}

func termSize() (w, h int) {
	for _, fd := range ttyFDs() {
		if c, r, ok := winSize(fd); ok {
			return c, r
		}
	}
	return 80, 24
}

func ttyFDs() []uintptr {
	return []uintptr{uintptr(syscall.Stdout), uintptr(syscall.Stderr), uintptr(syscall.Stdin)}
}

// notifyResize polls the console and pokes the channel when the size changed.
// The value sent is never inspected by the loop -- it drains the channel to
// coalesce a drag -- so the signal type is only a carrier.
func notifyResize(ch chan os.Signal) {
	go func() {
		lastW, lastH := termSize()
		for {
			time.Sleep(250 * time.Millisecond)
			w, h := termSize()
			if w == lastW && h == lastH {
				continue
			}
			lastW, lastH = w, h
			select {
			case ch <- syscall.Signal(28):
			default:
			}
		}
	}()
}
