// wheelprobe records what Terminal.app SENDS to a program on the alternate
// screen when the user turns the mouse wheel, with and without mouse
// reporting enabled.
//
// The scrollback plan needs a trigger. If the wheel already reaches the
// program as arrow keys (Terminal.app's "Scroll alternate screen"), every wheel
// tick in `xscapes claude` is landing in Claude Code's prompt as Up/Down today,
// and a scroll mode could key off it without mouse reporting. If it reaches
// nothing, the wheel is scrolling the terminal's own view and the plan has to
// turn mouse reporting on. Bytes received are written to -log, escaped.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/donlucasx/xscapes/internal/host"
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
	logPath := flag.String("log", "/tmp/wheelprobe.log", "where received bytes go")
	mouse := flag.Bool("mouse", false, "enable SGR mouse reporting (1000;1006)")
	hold := flag.Duration("hold", 12*time.Second, "how long to listen")
	flag.Parse()

	w, h := termSize()
	if w == 0 {
		fmt.Println("run this in a Terminal.app window")
		return
	}
	var t syscall.Termios
	fd := os.Stdin.Fd()
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&t)))
	saved := t
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	t.Cc[syscall.VMIN], t.Cc[syscall.VTIME] = 0, 1
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&t)))
	defer syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&saved)))

	band, _ := host.Band(h)
	fmt.Print(host.Open(true, band, h))
	if *mouse {
		fmt.Print("\x1b[?1000h\x1b[?1006h")
		defer fmt.Print("\x1b[?1006l\x1b[?1000l")
	}
	for r := 1; r <= band; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2KALT ROW %02d  wheelprobe mouse=%v listening %s", r, r, *mouse, *hold)
	}
	lf, _ := os.Create(*logPath)
	defer lf.Close()
	buf := make([]byte, 256)
	deadline := time.Now().Add(*hold)
	total := 0
	for time.Now().Before(deadline) {
		n, _ := os.Stdin.Read(buf)
		if n > 0 {
			total += n
			fmt.Fprintf(lf, "%s\n", strconv.Quote(string(buf[:n])))
		}
	}
	fmt.Print(host.Close(true, band, h))
	fmt.Printf("wheelprobe done mouse=%v: %d bytes received, see %s\n", *mouse, total, *logPath)
	time.Sleep(2 * time.Second)
}
