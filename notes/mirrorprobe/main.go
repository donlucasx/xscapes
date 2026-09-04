// mirrorprobe measures whether rows can be written into the MAIN buffer's
// scrollback while the ALTERNATE screen stays on display, by switching buffers
// without a clear (DECSET 47) around each write.
//
// If it works, lines that leave the agent's band can be mirrored into the
// terminal's own history -- the scrollback the user already knows how to use,
// with the wheel, selection and search that come with it -- and no viewer has
// to be built. The alternate screen has no history of its own (measured with
// notes/histprobe), so this is the only way native scrollback and the
// alternate screen can coexist.
//
// Read back over AppleScript: `contents of tab` for the alternate screen being
// intact, `history of tab` for the mirrored lines. -mode 1049 is the control
// that is expected to destroy the alternate screen on every round trip.
package main

import (
	"flag"
	"fmt"
	"os"
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
	mode := flag.String("mode", "47", "47 | 1047 | 1049: the switch used for the round trip")
	n := flag.Int("n", 30, "lines to mirror")
	gap := flag.Duration("gap", 40*time.Millisecond, "pause between mirrored lines")
	flag.Parse()

	w, h := termSize()
	if w == 0 {
		fmt.Println("run this in a Terminal.app window")
		return
	}
	band, _ := host.Band(h)
	fmt.Print(host.Open(true, band, h))
	fmt.Print(host.BeginPaint())
	for r := band + 1; r <= h; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2K~~~ SCAPE ROW %02d ~~~", r, r)
	}
	fmt.Print(host.EndPaint(band))
	for r := 1; r <= band; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2KALT ROW %02d (must survive)", r, r)
	}
	fmt.Printf("\x1b[%d;28H", band) // park the cursor somewhere recognisable

	toMain := fmt.Sprintf("\x1b[?%sl", *mode)
	toAlt := fmt.Sprintf("\x1b[?%sh", *mode)
	for i := 1; i <= *n; i++ {
		// Save the agent's cursor and region, go to the main buffer, append one
		// line at its bottom so it scrolls into history, come back, re-pin.
		fmt.Print("\x1b7\x1b[?6l\x1b[r")
		fmt.Print(toMain)
		fmt.Printf("\x1b[%d;1H\x1b[0mMIRROR LINE %02d\r\n", h, i)
		fmt.Print(toAlt)
		fmt.Print(host.EnterBand(band) + "\x1b8")
		time.Sleep(*gap)
	}
	fmt.Print("\x1b7\x1b[?6l\x1b[r")
	fmt.Printf("\x1b[%d;1H\x1b[2Kmirrorprobe mode=%s: %d lines mirrored, %dx%d band 1..%d. Holding.", h, *mode, *n, w, h, band)
	fmt.Print(host.EnterBand(band) + "\x1b8")
	time.Sleep(6 * time.Second)
	fmt.Print(host.Close(true, band, h))
	fmt.Printf("mirrorprobe done mode=%s\n", *mode)
	time.Sleep(3 * time.Second)
}
