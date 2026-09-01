package host

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPTYCarriesAChildsOutput(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Fatalf("openPTY: %v", err)
	}
	defer p.Close()

	cmd := exec.Command("sh", "-c", "echo marker-9f3a")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = p.slave, p.slave, p.slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	p.slave.Close() // the child owns it now; keeping it open hides EOF
	defer cmd.Wait()

	var got bytes.Buffer
	done := make(chan struct{})
	go func() {
		io.Copy(&got, p.master)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out reading the pty")
	}
	if !strings.Contains(got.String(), "marker-9f3a") {
		t.Errorf("pty output %q does not contain the child's line", got.String())
	}
}

// The whole design rests on the child believing the window is band-sized: size
// the pty to (rows - scape) and Claude lays itself out inside the band with no
// further help. If the size does not reach the child, nothing else works.
func TestPTYReportsTheSizeTheHostSets(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Fatalf("openPTY: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(101, 21); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	cols, rows, ok := ptySize(p.slave.Fd())
	if !ok {
		t.Fatal("could not read the size back from the slave side")
	}
	if cols != 101 || rows != 21 {
		t.Errorf("child sees %dx%d, want 101x21", cols, rows)
	}
}

func TestPTYResizeReachesTheChild(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Fatalf("openPTY: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(80, 24); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	if err := p.SetSize(120, 33); err != nil {
		t.Fatalf("resize: %v", err)
	}
	cols, rows, _ := ptySize(p.slave.Fd())
	if cols != 120 || rows != 33 {
		t.Errorf("after resize the child sees %dx%d, want 120x33", cols, rows)
	}
}
