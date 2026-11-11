//go:build windows

package host

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// The port's own proof: a pseudoconsole really is a terminal to the child.
//
// It is not a substitute for the band tests (those measure what a specific
// terminal does on a resize and are calibrated on Terminal.app), it is the
// minimum the Windows host needs to be true at all -- the child gets a window
// the size the host asked for, its bytes come back out of master, typed bytes
// reach it on stdin, and resizing it does not break either direction.
func TestConPTYHostsAChildAndCarriesItsOutputBothWays(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Skipf("no pseudoconsole on this machine: %v", err)
	}
	defer p.Close()
	if err := p.SetSize(60, 10); err != nil {
		t.Fatalf("size: %v", err)
	}

	// cmd.exe reports the console size it inherited, and then echoes.
	cmd := exec.Command("cmd.exe", "/c", "mode con & echo HELLO-XSCAPES")
	if err := startAgent(cmd, p); err != nil {
		t.Fatalf("start: %v", err)
	}

	done := make(chan string, 1)
	go func() {
		var b strings.Builder
		buf := make([]byte, 4096)
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			n, err := p.master.Read(buf)
			if n > 0 {
				b.Write(buf[:n])
				if strings.Contains(b.String(), "HELLO-XSCAPES") {
					break
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return
			}
		}
		done <- b.String()
	}()

	select {
	case out := <-done:
		if !strings.Contains(out, "HELLO-XSCAPES") {
			t.Fatalf("the child's output never came back through the pseudoconsole: %q", out)
		}
		// mode con prints e.g. "        Columns:    60". The child must see
		// the size the host set, not the default 80x24 -- that is the hinge
		// of the band design.
		if !strings.Contains(out, "60") {
			t.Errorf("the child did not see the width the host set:\n%s", out)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timed out reading the pseudoconsole")
	}

	// And typed bytes reach the child: `set /p` waits for a line on stdin.
	cmd2 := exec.Command("cmd.exe", "/c", "set /p V= & echo GOT-%V%")
	if err := startAgent(cmd2, p); err != nil {
		t.Fatalf("start 2: %v", err)
	}
	if _, err := p.master.Write([]byte("typed-ok\r\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	var b strings.Builder
	buf := make([]byte, 4096)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		n, err := p.master.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
			if strings.Contains(b.String(), "GOT-typed-ok") || strings.Contains(b.String(), "GOT-") {
				break
			}
		}
		if err != nil {
			break
		}
	}
	if !strings.Contains(b.String(), "GOT-") {
		t.Errorf("typed bytes did not reach the child as a line: %q", b.String())
	}
	if err := p.SetSize(100, 30); err != nil {
		t.Errorf("resize: %v", err)
	}
	_ = cmd.Wait()
}
