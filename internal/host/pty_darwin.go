//go:build darwin

package host

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// openPTY opens a pseudo-terminal pair the BSD way: /dev/ptmx hands back a
// master, then three ioctls grant, unlock and name the slave.
func openPTY() (*pty, error) {
	m, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}
	if err := ioctl(m.Fd(), syscall.TIOCPTYGRANT, 0); err != nil {
		m.Close()
		return nil, fmt.Errorf("grant pty: %w", err)
	}
	if err := ioctl(m.Fd(), syscall.TIOCPTYUNLK, 0); err != nil {
		m.Close()
		return nil, fmt.Errorf("unlock pty: %w", err)
	}
	var name [128]byte
	if err := ioctl(m.Fd(), syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&name[0]))); err != nil {
		m.Close()
		return nil, fmt.Errorf("name pty: %w", err)
	}
	path := string(bytes.TrimRight(name[:], "\x00"))
	s, err := os.OpenFile(path, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		m.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return &pty{master: m, slave: s}, nil
}
