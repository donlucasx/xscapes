package host

import "testing"

// The host must hand every keystroke to the agent untouched: no line editing,
// no echo, and no interpretation of Ctrl-C, which belongs to Claude Code and
// not to the scape wrapped around it. It must also put the terminal back --
// a host that exits raw leaves the user with a shell that does not echo.

func TestRawModeClearsEchoAndLineDiscipline(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Fatalf("openPTY: %v", err)
	}
	defer p.Close()
	fd := p.slave.Fd()

	before, err := readTermios(fd)
	if err != nil {
		t.Fatalf("readTermios: %v", err)
	}
	if before.Lflag&lflagEcho == 0 {
		t.Fatal("a fresh pty already has echo off; the test proves nothing")
	}

	saved, err := makeRaw(fd)
	if err != nil {
		t.Fatalf("makeRaw: %v", err)
	}
	got, err := readTermios(fd)
	if err != nil {
		t.Fatalf("readTermios: %v", err)
	}
	if got.Lflag&lflagEcho != 0 {
		t.Error("echo still on in raw mode")
	}
	if got.Lflag&lflagCanon != 0 {
		t.Error("canonical mode still on in raw mode")
	}
	if got.Lflag&lflagISig != 0 {
		t.Error("signal generation still on: Ctrl-C would be eaten by the host, not passed to the agent")
	}

	if err := restoreTermios(fd, saved); err != nil {
		t.Fatalf("restoreTermios: %v", err)
	}
	after, err := readTermios(fd)
	if err != nil {
		t.Fatalf("readTermios: %v", err)
	}
	if after.Lflag != before.Lflag || after.Iflag != before.Iflag || after.Oflag != before.Oflag {
		t.Errorf("terminal not restored: lflag %x->%x, iflag %x->%x, oflag %x->%x",
			before.Lflag, after.Lflag, before.Iflag, after.Iflag, before.Oflag, after.Oflag)
	}
}
