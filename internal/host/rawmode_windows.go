//go:build windows

package host

import (
	"fmt"
	"syscall"
)

// Raw mode on Windows is two console modes, not a termios struct.
//
// The intent is exactly the same as the unix side: every byte the user types
// reaches the agent as typed. Disabling ECHO_INPUT and LINE_INPUT stops the
// console echoing and buffering it; disabling PROCESSED_INPUT stops Ctrl-C
// becoming a console control event, so the agent gets 0x03 as a byte and can
// interrupt its own turn (which is what Ctrl-C means inside Claude Code or
// Hermes); and ENABLE_VIRTUAL_TERMINAL_INPUT makes arrows and function keys
// arrive as the VT sequences the agent expects from every other terminal.
//
// The OUTPUT mode is switched too, because none of the scape's escapes do
// anything without ENABLE_VIRTUAL_TERMINAL_PROCESSING. It is restored with the
// rest, so a program started afterwards finds the console as it found it.

const (
	enableProcessedInput    = 0x0001
	enableLineInput         = 0x0002
	enableEchoInput         = 0x0004
	enableWindowInput       = 0x0008
	enableMouseInput        = 0x0010
	enableVirtualTermInput  = 0x0200
	enableProcessedOutput   = 0x0001
	enableVirtualTermOutput = 0x0004
	enableWrapAtEOLOutput   = 0x0002
)

// rawState is the console modes as they were, for restoreTermios.
type rawState struct {
	in, out         syscall.Handle
	inMode, outMode uint32
	haveOut         bool
}

func getConsoleMode(h syscall.Handle) (uint32, error) {
	var mode uint32
	if err := syscall.GetConsoleMode(h, &mode); err != nil {
		return 0, err
	}
	return mode, nil
}

func setConsoleMode(h syscall.Handle, mode uint32) error {
	if r, _, e := procSetConsoleModeWin.Call(uintptr(h), uintptr(mode)); r == 0 {
		return fmt.Errorf("SetConsoleMode: %v", e)
	}
	return nil
}

func makeRaw(fd uintptr) (*rawState, error) {
	in := syscall.Handle(fd)
	inMode, err := getConsoleMode(in)
	if err != nil {
		// Not a console at all: a pipe, a file, a CI runner. There is
		// nothing to switch and nothing to restore, and the host is still
		// worth running -- the agent's bytes pass through either way.
		return &rawState{}, nil
	}
	st := &rawState{in: in, inMode: inMode}
	raw := inMode
	raw &^= enableEchoInput | enableLineInput | enableProcessedInput | enableMouseInput
	raw |= enableVirtualTermInput | enableWindowInput
	if err := setConsoleMode(in, raw); err != nil {
		return nil, err
	}

	out := syscall.Handle(syscall.Stdout)
	if outMode, err := getConsoleMode(out); err == nil {
		st.out, st.outMode, st.haveOut = out, outMode, true
		m := outMode | enableProcessedOutput | enableVirtualTermOutput
		_ = setConsoleMode(out, m)
	}
	return st, nil
}

func restoreTermios(fd uintptr, st *rawState) error {
	if st == nil {
		return nil
	}
	err := setConsoleMode(st.in, st.inMode)
	if st.haveOut {
		if e := setConsoleMode(st.out, st.outMode); err == nil {
			err = e
		}
	}
	return err
}

// EnableVT turns on VT escape processing for this process's console, and is a
// no-op when output is a pipe or the mode is already on.
//
// Called before anything is painted: without it a Windows console prints the
// scape's escapes as text. Safe to call when there is no console at all.
func EnableVT() {
	h := syscall.Handle(syscall.Stdout)
	mode, err := getConsoleMode(h)
	if err != nil {
		return
	}
	if mode&enableVirtualTermOutput != 0 {
		return
	}
	_ = setConsoleMode(h, mode|enableProcessedOutput|enableVirtualTermOutput)
	// UTF-16 to UTF-8: the scape is full of block glyphs and braille.
	procSetConsoleCP.Call(65001)
}
