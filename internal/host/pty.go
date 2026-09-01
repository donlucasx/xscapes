package host

import (
	"os"
	"syscall"
	"unsafe"
)

// pty is a pseudo-terminal pair. The host writes the agent's input to the
// master and reads its output from it; the agent gets the slave as its
// controlling terminal and cannot tell it apart from a real one.
//
// Opened with raw ioctls rather than a dependency, so the project stays on the
// standard library. Everything needed is already in package syscall on both
// platforms it runs on.
type pty struct{ master, slave *os.File }

// SetSize tells the child how big its window is.
//
// This is the hinge of the whole design: the child is sized to the band rather
// than the window, so Claude Code lays itself out inside the band without
// being asked to, and its text can never collide with the scape below. The
// kernel sends the child a SIGWINCH, and Claude repaints from the top -- the
// one moment it uses absolute cursor positioning, which is why the host also
// turns on origin mode.
func (p *pty) SetSize(cols, rows int) error {
	ws := struct{ rows, cols, x, y uint16 }{uint16(rows), uint16(cols), 0, 0}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, p.master.Fd(),
		uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return errno
	}
	return nil
}

func (p *pty) Close() error {
	if p.slave != nil {
		p.slave.Close()
	}
	if p.master != nil {
		return p.master.Close()
	}
	return nil
}

// ptySize reads a terminal's size back out of the kernel, for the tests and
// for the resize loop.
func ptySize(fd uintptr) (cols, rows int, ok bool) {
	var ws struct{ rows, cols, x, y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, false
	}
	return int(ws.cols), int(ws.rows), true
}

func ioctl(fd, req, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg); errno != 0 {
		return errno
	}
	return nil
}
