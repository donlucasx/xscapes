//go:build linux

package host

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// openPTY opens a pseudo-terminal pair the SysV way: unlock the slave, ask for
// its number, open /dev/pts/<n>.
func openPTY() (*pty, error) {
	m, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}
	var unlock int32
	if err := ioctl(m.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); err != nil {
		m.Close()
		return nil, fmt.Errorf("unlock pty: %w", err)
	}
	var n int32
	if err := ioctl(m.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n))); err != nil {
		m.Close()
		return nil, fmt.Errorf("name pty: %w", err)
	}
	path := fmt.Sprintf("/dev/pts/%d", n)
	s, err := os.OpenFile(path, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		m.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return &pty{master: m, slave: s}, nil
}
