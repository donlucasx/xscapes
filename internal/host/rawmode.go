package host

import (
	"syscall"
	"unsafe"
)

// Raw mode on the host's own terminal.
//
// The host is a pass-through: every byte the user types goes to the agent
// exactly as typed. That means no echo (the agent draws its own input line),
// no line buffering (Claude reacts per keystroke), and no signal generation --
// Ctrl-C belongs to Claude Code, which uses it to interrupt a turn, and a host
// that turned it into a SIGINT would kill the scape instead.
//
// The terminal's replies to the agent's startup queries (device attributes,
// the kitty keyboard protocol, synchronized output) arrive on the same path,
// so it has to stay byte-transparent in both directions.
const (
	lflagEcho  = syscall.ECHO
	lflagCanon = syscall.ICANON
	lflagISig  = syscall.ISIG
)

func readTermios(fd uintptr) (*syscall.Termios, error) {
	var t syscall.Termios
	if err := ioctl(fd, ioctlReadTermios, uintptr(unsafe.Pointer(&t))); err != nil {
		return nil, err
	}
	return &t, nil
}

func writeTermios(fd uintptr, t *syscall.Termios) error {
	return ioctl(fd, ioctlWriteTermios, uintptr(unsafe.Pointer(t)))
}

// makeRaw switches the terminal to raw mode and returns what it was, for
// restoreTermios. The flag set is the classic cfmakeraw one.
func makeRaw(fd uintptr) (*syscall.Termios, error) {
	old, err := readTermios(fd)
	if err != nil {
		return nil, err
	}
	raw := *old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := writeTermios(fd, &raw); err != nil {
		return nil, err
	}
	return old, nil
}

func restoreTermios(fd uintptr, t *syscall.Termios) error {
	if t == nil {
		return nil
	}
	return writeTermios(fd, t)
}
