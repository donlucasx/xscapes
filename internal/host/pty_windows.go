//go:build windows

package host

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// THE WINDOWS HOST: a pseudoconsole, not a pty.
//
// Windows has no /dev/ptmx. What it has, since Windows 10 1809, is ConPTY:
// CreatePseudoConsole hands back an HPCON that a child is attached to through
// a process attribute, and the parent talks to the child through two pipes.
// The child sees a real console (it can query its size, it gets VT output
// sequences, it receives keystrokes) exactly the way it does under a pty on
// the other two platforms -- which is the whole reason the band host works at
// all: the agent lays itself out in a window the size of the band, and its
// text can never reach the scape's rows.
//
// Everything here is called through NewLazyDLL rather than golang.org/x/sys,
// so the project stays on the standard library as the brief requires.
//
// Three details are load-bearing and were established by measurement, not from
// the docs (notes/windows-conpty.md):
//
//   - PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE takes the HPCON ITSELF as its value,
//     not its address. Passing &hpc creates the console and launches the child
//     with no error at all -- and the child comes up attached to the PARENT's
//     console instead: its output lands on the real screen, and the pipes stay
//     silent for the rest of the session.
//   - STARTF_USESTDHANDLES with three NULL handles is what stops the child
//     inheriting the parent's std handles (which is the same failure, worn
//     differently).
//   - The pipes are read and written with ReadFile/WriteFile on the raw
//     handle, not through os.File: a CreatePipe handle is synchronous, and
//     handing it to Go's poller leaves a Read that never returns.

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreatePseudoConsole  = modkernel32.NewProc("CreatePseudoConsole")
	procResizePseudoConsole  = modkernel32.NewProc("ResizePseudoConsole")
	procClosePseudoConsole   = modkernel32.NewProc("ClosePseudoConsole")
	procCreateProcessW       = modkernel32.NewProc("CreateProcessW")
	procInitAttributeList    = modkernel32.NewProc("InitializeProcThreadAttributeList")
	procUpdateAttribute      = modkernel32.NewProc("UpdateProcThreadAttribute")
	procDeleteAttributeList  = modkernel32.NewProc("DeleteProcThreadAttributeList")
	procSetConsoleModeWin    = modkernel32.NewProc("SetConsoleMode")
	procSetConsoleCP         = modkernel32.NewProc("SetConsoleOutputCP")
	procSetConsoleScreenBuf  = modkernel32.NewProc("SetConsoleScreenBufferSize")
	procGetConsoleScreenBuf  = modkernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorInfo = modkernel32.NewProc("SetConsoleCursorInfo")
)

// PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE is the attribute a child is created with
// to hand it the pseudoconsole. It is a windows.h constant, not in package
// syscall.
const procThreadAttributePseudoConsole = 0x00020016

// _DEPRECATED_EXTENDED_STARTUPINFO_PRESENT: the process must be created with
// a STARTUPINFOEX or the attribute list is ignored.
const (
	createUnicodeEnvironment   = 0x00000400
	extendedStartupInfoPresent = 0x00080000
	startfUseStdHandles        = 0x00000100
)

// pty is the pair the host holds: master is the wire to the child, and slave
// is nil -- a pseudoconsole has no slave path, the child is attached by
// attribute instead (see startAgent).
type pty struct {
	master *conptyStream
	hpc    syscall.Handle
	slave  *os.File
	// the console's own ends of the two pipes, kept open for the console's
	// lifetime: closing them here would stop its output for good.
	ptyIn, ptyOut syscall.Handle
	closeOnce     sync.Once
}

// conptyStream is one ReadWriteCloser over the two pipes: Write goes to the
// child's stdin, Read comes from its stdout.
type conptyStream struct {
	in  syscall.Handle // write end of the child's input
	out syscall.Handle // read end of the child's output
	// closed is set by Close before the handles go: the host's output
	// forwarder runs in its own goroutine and can be inside Read when the
	// session ends, and a closed handle number can be handed to something
	// else the moment it is released.
	closed atomic.Bool
}

var errClosed = fmt.Errorf("pseudoconsole closed")

func (c *conptyStream) Read(p []byte) (int, error) {
	if c.closed.Load() {
		return 0, errClosed
	}
	var n uint32
	if err := syscall.ReadFile(c.out, p, &n, nil); err != nil {
		return int(n), err
	}
	if n == 0 {
		return 0, errClosed
	}
	return int(n), nil
}

func (c *conptyStream) Write(p []byte) (int, error) {
	if c.closed.Load() {
		return 0, errClosed
	}
	var n uint32
	if err := syscall.WriteFile(c.in, p, &n, nil); err != nil {
		return int(n), err
	}
	return int(n), nil
}

func (c *conptyStream) Close() error {
	c.closed.Store(true)
	err := syscall.CloseHandle(c.out)
	if e := syscall.CloseHandle(c.in); err == nil {
		err = e
	}
	return err
}

// packSize is COORD in the one machine word CreatePseudoConsole takes it in:
// width in the low word, height in the high one.
func packSize(cols, rows int) uintptr {
	return uintptr(uint32(uint16(cols)) | uint32(uint16(rows))<<16)
}

func openPTY() (*pty, error) {
	var ptyIn, cmdIn, cmdOut, ptyOut syscall.Handle
	if err := syscall.CreatePipe(&ptyIn, &cmdIn, nil, 0); err != nil {
		return nil, fmt.Errorf("pipe in: %w", err)
	}
	if err := syscall.CreatePipe(&cmdOut, &ptyOut, nil, 0); err != nil {
		syscall.CloseHandle(ptyIn)
		syscall.CloseHandle(cmdIn)
		return nil, fmt.Errorf("pipe out: %w", err)
	}
	// The console starts at 80x24 and is resized to the band a moment later
	// by SetSize, the same way the pty path does it.
	var hpc syscall.Handle
	hr, _, _ := procCreatePseudoConsole.Call(packSize(80, 24),
		uintptr(ptyIn), uintptr(ptyOut), 0, uintptr(unsafe.Pointer(&hpc)))
	if hr != 0 {
		for _, h := range []syscall.Handle{ptyIn, cmdIn, cmdOut, ptyOut} {
			syscall.CloseHandle(h)
		}
		return nil, fmt.Errorf("CreatePseudoConsole: HRESULT 0x%08x (needs Windows 10 1809 or later)", uint32(hr))
	}
	return &pty{
		master: &conptyStream{in: cmdIn, out: cmdOut},
		hpc:    hpc,
		ptyIn:  ptyIn,
		ptyOut: ptyOut,
	}, nil
}

// SetSize resizes the pseudoconsole, which sends the child a window-size
// change the way TIOCSWINSZ does on unix.
func (p *pty) SetSize(cols, rows int) error {
	if cols < 1 || rows < 1 {
		return nil
	}
	hr, _, _ := procResizePseudoConsole.Call(uintptr(p.hpc), packSize(cols, rows))
	if hr != 0 {
		return fmt.Errorf("ResizePseudoConsole: HRESULT 0x%08x", uint32(hr))
	}
	return nil
}

func (p *pty) Close() error {
	var err error
	p.closeOnce.Do(func() {
		if p.hpc != 0 {
			procClosePseudoConsole.Call(uintptr(p.hpc))
			p.hpc = 0
		}
		// NOT p.master = nil: the output forwarder holds this pointer and
		// reads through it, so the stream itself must outlive the pair.
		if p.master != nil {
			err = p.master.Close()
		}
		syscall.CloseHandle(p.ptyIn)
		syscall.CloseHandle(p.ptyOut)
	})
	return err
}

// startAgent creates the agent attached to the pseudoconsole.
//
// exec.Cmd cannot be used for this: syscall.StartProcess builds its own
// attribute list (for the parent process and the handle list) and there is no
// exported way to add a third attribute, so CreateProcessW is called here and
// the resulting pid is handed back to exec.Cmd as its Process. Wait, and the
// whole of Run's exit path, then work unchanged.
func startAgent(cmd *exec.Cmd, p *pty) error {
	line := composeCmdLine(cmd.Args)
	linep, err := syscall.UTF16PtrFromString(line)
	if err != nil {
		return err
	}
	appp, err := syscall.UTF16PtrFromString(cmd.Path)
	if err != nil {
		return err
	}
	var dirp *uint16
	if cmd.Dir != "" {
		if dirp, err = syscall.UTF16PtrFromString(cmd.Dir); err != nil {
			return err
		}
	}
	var envp *uint16
	if cmd.Env != nil {
		block, err := environmentBlock(cmd.Env)
		if err != nil {
			return err
		}
		if len(block) > 0 {
			envp = &block[0]
		}
	}

	// The attribute list: one entry, the pseudoconsole handle.
	var size uintptr
	procInitAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&size)))
	if size == 0 {
		return fmt.Errorf("InitializeProcThreadAttributeList: no size")
	}
	buf := make([]byte, size)
	if hr, _, _ := procInitAttributeList.Call(uintptr(unsafe.Pointer(&buf[0])), 1, 0, uintptr(unsafe.Pointer(&size))); hr == 0 {
		return fmt.Errorf("InitializeProcThreadAttributeList failed")
	}
	attr := unsafe.Pointer(&buf[0])
	if hr, _, _ := procUpdateAttribute.Call(uintptr(attr), 0, procThreadAttributePseudoConsole,
		uintptr(p.hpc), unsafe.Sizeof(p.hpc), 0, 0); hr == 0 {
		procDeleteAttributeList.Call(uintptr(attr))
		return fmt.Errorf("UpdateProcThreadAttribute(PSEUDOCONSOLE) failed")
	}
	defer procDeleteAttributeList.Call(uintptr(attr))

	// STARTF_USESTDHANDLES with three NULL handles: the child's std handles
	// come from the pseudoconsole, not by inheriting ours.
	var si struct {
		Cb            uint32
		_             *uint16
		Desktop       *uint16
		Title         *uint16
		X             uint32
		Y             uint32
		XSize         uint32
		YSize         uint32
		XCountChars   uint32
		YCountChars   uint32
		FillAttribute uint32
		Flags         uint32
		ShowWindow    uint16
		_             uint16
		_             *byte
		StdInput      syscall.Handle
		StdOutput     syscall.Handle
		StdErr        syscall.Handle
		AttributeList unsafe.Pointer
	}
	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = startfUseStdHandles
	si.AttributeList = attr

	var pi struct {
		Process   syscall.Handle
		Thread    syscall.Handle
		ProcessId uint32
		ThreadId  uint32
	}
	flags := uintptr(createUnicodeEnvironment | extendedStartupInfoPresent)
	if envp == nil {
		flags &^= createUnicodeEnvironment
	}
	hr, _, e := procCreateProcessW.Call(
		uintptr(unsafe.Pointer(appp)), uintptr(unsafe.Pointer(linep)),
		0, 0, 0, flags,
		uintptr(unsafe.Pointer(envp)), uintptr(unsafe.Pointer(dirp)),
		uintptr(unsafe.Pointer(&si)), uintptr(unsafe.Pointer(&pi)))
	if hr == 0 {
		return fmt.Errorf("CreateProcess %s: %v", cmd.Path, e)
	}
	syscall.CloseHandle(pi.Thread)
	proc, err := os.FindProcess(int(pi.ProcessId))
	syscall.CloseHandle(pi.Process)
	if err != nil {
		return fmt.Errorf("attach to pid %d: %w", pi.ProcessId, err)
	}
	cmd.Process = proc
	return nil
}

// environmentBlock is CREATE_UNICODE_ENVIRONMENT's sorted, NUL-separated,
// double-NUL-terminated array. Sorted case-insensitively, which is what
// Windows expects.
//
// Built entry by entry: syscall.UTF16FromString REJECTS a string containing a
// NUL byte, and the block is nothing but NUL-separated strings, so handing it
// the whole block returns EINVAL -- which CreateProcess then reports as
// "invalid argument" against the process, a message that names nothing.
func environmentBlock(env []string) ([]uint16, error) {
	vals := append([]string(nil), env...)
	sort.Slice(vals, func(i, j int) bool {
		return strings.ToUpper(vals[i]) < strings.ToUpper(vals[j])
	})
	out := make([]uint16, 0, len(vals)*32+1)
	for _, kv := range vals {
		u, err := syscall.UTF16FromString(kv)
		if err != nil {
			return nil, err
		}
		out = append(out, u...) // u already ends in its NUL
	}
	out = append(out, 0)
	return out, nil
}

// composeCmdLine quotes argv the way CommandLineToArgvW reads it back.
func composeCmdLine(args []string) string {
	var b strings.Builder
	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(quoteArg(a))
	}
	return b.String()
}

func quoteArg(a string) string {
	if a != "" && !strings.ContainsAny(a, " \t\n\v\"") {
		return a
	}
	var b strings.Builder
	b.WriteByte('"')
	slashes := 0
	for _, r := range a {
		switch r {
		case '\\':
			slashes++
		case '"':
			b.WriteString(strings.Repeat("\\", slashes*2+1))
			b.WriteRune(r)
			slashes = 0
			continue
		default:
			b.WriteString(strings.Repeat("\\", slashes))
			slashes = 0
		}
		b.WriteRune(r)
	}
	b.WriteString(strings.Repeat("\\", slashes*2))
	b.WriteByte('"')
	return b.String()
}
