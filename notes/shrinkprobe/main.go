// shrinkprobe measures what happens to the AGENT'S CURSOR when the window
// shrinks on the alternate screen, through the host's real Rebind, and where a
// Claude-style RELATIVE redraw then lands.
//
// The facts it builds on were measured on 2026-09-03 (notes/contentprobe,
// notes/anchorprobe): on Terminal.app's alternate screen a shrink of N pulls
// CONTENT up by N and leaves the CURSOR on its absolute row. The host's Rebind
// saves and restores that cursor, so after a shrink the agent draws N rows
// below the text it believes it is drawing next to. Claude Code places its
// input box purely by relative moves, which is exactly the shape of a split
// input box -- the report from 2026-09-03 ~11:57 that was never reproduced.
//
// Read back with `contents of tab` over AppleScript, which returns the visible
// cells of the alternate screen too (measured with notes/histprobe), so nothing
// here is read by eye. The cursor is read with DSR, which the probe can do
// because it owns stdin; the host cannot.
//
//	-fix none  no correction after Rebind (what production does today)
//	-fix cuu   cursor up by the full shrink
//	-fix sd    scroll the screen down by the scape's share, cursor up by the band's
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
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

// raw puts the terminal in raw mode so the DSR reply can be read without echo
// or line buffering.
func raw(fd uintptr) (syscall.Termios, error) {
	var t syscall.Termios
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&t))); e != 0 {
		return t, e
	}
	saved := t
	t.Lflag &^= syscall.ECHO | syscall.ICANON
	t.Cc[syscall.VMIN], t.Cc[syscall.VTIME] = 0, 5 // half a second
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&t)))
	return saved, nil
}

func restore(fd uintptr, t syscall.Termios) {
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(&t)))
}

// cursorRow asks with DSR 6 and parses ESC [ row ; col R.
func cursorRow(in *bufio.Reader) int {
	fmt.Print("\x1b[6n")
	var buf []byte
	for {
		b, err := in.ReadByte()
		if err != nil {
			return -1
		}
		buf = append(buf, b)
		if b == 'R' {
			break
		}
	}
	s := string(buf)
	i := strings.LastIndex(s, "\x1b[")
	if i < 0 {
		return -1
	}
	var r, c int
	fmt.Sscanf(s[i+2:], "%d;%dR", &r, &c)
	return r
}

func main() {
	fix := flag.String("fix", "none", "none | cuu | sd")
	flag.Parse()

	w, h := termSize()
	if w == 0 {
		fmt.Println("run this in a Terminal.app window")
		return
	}
	saved, err := raw(os.Stdin.Fd())
	if err != nil {
		fmt.Println("raw:", err)
		return
	}
	defer restore(os.Stdin.Fd(), saved)
	in := bufio.NewReader(os.Stdin)

	band, _ := host.Band(h)
	fmt.Print(host.Open(true, band, h))
	// Scape rows, marked.
	fmt.Print(host.BeginPaint())
	for r := band + 1; r <= h; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2K~~~ SCAPE ROW %02d ~~~", r, r)
	}
	fmt.Print(host.EndPaint(band))
	// Transcript rows above the box, then the box on the last three band rows.
	// Origin mode is on, so row 1 here is the band's row 1.
	for r := 1; r <= band-3; r++ {
		fmt.Printf("\x1b[%d;1H\x1b[2KTRANSCRIPT ROW %02d", r, r)
	}
	fmt.Printf("\x1b[%d;1H\x1b[2K+---- old box ----+", band-2)
	fmt.Printf("\x1b[%d;1H\x1b[2K| > hi            |", band-1)
	fmt.Printf("\x1b[%d;1H\x1b[2K+-----------------+", band)
	fmt.Printf("\x1b[%d;7H", band-1) // cursor after "> hi", the way Claude leaves it
	before := cursorRow(in)

	// Wait for the window to shrink, the way host.Run polls.
	oldRows, oldAgent := h, band
	var rows, agent int
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		_, rows = termSize()
		if rows != oldRows {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if rows == oldRows {
		fmt.Print(host.Close(true, band, h))
		fmt.Println("no resize seen")
		return
	}
	agent, _ = host.Band(rows)
	afterResize := cursorRow(in)
	// host.Run's arithmetic, verbatim.
	drop := oldRows - rows
	if drop < 0 {
		drop = 0
	}
	from := oldAgent + 1 - drop
	if from < 1 {
		from = 1
	}
	to := min(agent, oldRows-drop)
	shrink := oldRows - rows
	bandShrink := oldAgent - agent
	afterRebind := -1

	var b strings.Builder
	switch *fix {
	case "none":
		fmt.Print(host.Rebind(0, from, to, agent))
		afterRebind = cursorRow(in)
	case "cuu", "sd":
		// A replacement for Rebind on a shrink. The order is the whole fix:
		// DECRC into a band that no longer contains the saved row lands on
		// row 1 (measured, fix=none), so the restore happens while the region
		// is still the full screen, the relative move follows it, and only
		// then is the band re-pinned -- around a fresh save of a row the new
		// band does contain.
		b.WriteString("\x1b7\x1b[?6l\x1b[r\x1b[0m")
		if *fix == "sd" {
			// The scape's share of the shrink is what left blank rows under
			// the agent's text; scrolling the screen down by it puts the
			// text's bottom back on the band's bottom, so nothing in the
			// band needs clearing.
			fmt.Fprintf(&b, "\x1b[%dT", shrink-bandShrink)
			b.WriteString("\x1b8")
			fmt.Fprintf(&b, "\x1b[%dA", bandShrink)
		} else {
			for r := from; r <= to; r++ {
				fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", r)
			}
			b.WriteString("\x1b8")
			fmt.Fprintf(&b, "\x1b[%dA", shrink)
		}
		b.WriteString("\x1b7" + host.EnterBand(agent) + "\x1b8")
		fmt.Print(b.String())
	}
	afterFix := cursorRow(in)

	// Claude's redraw: from the input row, up one, and rewrite the box with
	// relative moves only. '=' so the new box is distinguishable from the old.
	fmt.Print("\r\x1b[1A\x1b[2K+==== new box ====+")
	fmt.Print("\r\n\x1b[2K| > hi again      |")
	fmt.Print("\r\n\x1b[2K+=================+")
	fmt.Print("\r\x1b[1A\x1b[13C")
	afterDraw := cursorRow(in)

	// A status row for the reader, outside the band, absolute.
	fmt.Print("\x1b7\x1b[?6l\x1b[r")
	fmt.Printf("\x1b[%d;1H\x1b[2KSHRINK %d->%d (-%d) band %d->%d fix=%s cursor: before=%d afterResize=%d afterRebind=%d afterFix=%d afterDraw=%d",
		rows, oldRows, rows, shrink, oldAgent, agent, *fix, before, afterResize, afterRebind, afterFix, afterDraw)
	fmt.Print(host.EnterBand(agent) + "\x1b8")

	time.Sleep(6 * time.Second)
	fmt.Print(host.Close(true, agent, rows))
	fmt.Printf("shrinkprobe done fix=%s\n", *fix)
	time.Sleep(3 * time.Second)
}
