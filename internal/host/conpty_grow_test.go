//go:build windows

package host

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// readUntil reads from m until the output contains want or deadline passes.
func readUntil(t *testing.T, m *conptyStream, want string, deadline time.Time) string {
	t.Helper()
	var out strings.Builder
	buf := make([]byte, 4096)
	for time.Now().Before(deadline) {
		n, err := m.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			if strings.Contains(out.String(), want) {
				return out.String()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return out.String()
		}
	}
	return out.String()
}

// --- GROW MEASUREMENT ---

func TestConPTYGrowAnchors(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Skipf("no pseudoconsole: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(100, 30); err != nil {
		t.Fatalf("initial size: %v", err)
	}

	cmd := exec.Command("cmd.exe", "/c", "echo HIST01 && echo BOXSTART && echo .-BOX.-. && echo BOXEND && echo DONE")
	if err := startAgent(cmd, p); err != nil {
		t.Fatalf("start: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	out := readUntil(t, p.master, "DONE", deadline)
	t.Logf("=== BEFORE RESIZE (%d bytes) ===\n%s", len(out), out)

	t.Logf("Resizing from 30 to 40 rows...")
	if err := p.SetSize(100, 40); err != nil {
		t.Fatalf("resize: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	after := readUntil(t, p.master, "", deadline)
	t.Logf("=== AFTER RESIZE (%d bytes) ===\n%s", len(after), after)
}

// --- SHRINK MEASUREMENT ---

func TestConPTYShrinkAnchors(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Skipf("no pseudoconsole: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(100, 40); err != nil {
		t.Fatalf("initial size: %v", err)
	}

	cmd := exec.Command("cmd.exe", "/c", "echo HIST01 && echo BOXSTART && echo .-BOX.-. && echo BOXEND && echo DONE")
	if err := startAgent(cmd, p); err != nil {
		t.Fatalf("start: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	out := readUntil(t, p.master, "DONE", deadline)
	t.Logf("=== BEFORE SHRINK (%d bytes) ===\n%s", len(out), out)

	t.Logf("Resizing from 40 to 30 rows...")
	if err := p.SetSize(100, 30); err != nil {
		t.Fatalf("resize: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	after := readUntil(t, p.master, "", deadline)
	t.Logf("=== AFTER SHRINK (%d bytes) ===\n%s", len(after), after)
}

// --- MIRROR MEASUREMENT ---

func TestConPTYMirrorRows(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Skipf("no pseudoconsole: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(80, 30); err != nil {
		t.Fatalf("initial size: %v", err)
	}

	cmd := exec.Command("cmd.exe", "/c", "echo A1 && echo A2 && echo A3 && echo A4 && echo A5 && echo A6 && echo A7 && echo A8 && echo A9 && echo A10 && echo A11 && echo A12 && echo A13 && echo A14 && echo A15 && echo A16 && echo A17 && echo A18 && echo A19 && echo A20 && echo A21 && echo A22 && echo A23 && echo A24 && echo DONE")
	if err := startAgent(cmd, p); err != nil {
		t.Fatalf("start: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	out := readUntil(t, p.master, "DONE", deadline)
	t.Logf("=== BEFORE MIRROR CHECK (%d bytes) ===\n%s", len(out), out)
}

// --- WIDTH CHANGE MEASUREMENT ---

func TestConPTYWidthChange(t *testing.T) {
	p, err := openPTY()
	if err != nil {
		t.Skipf("no pseudoconsole: %v", err)
	}
	defer p.Close()

	if err := p.SetSize(40, 10); err != nil {
		t.Fatalf("initial size: %v", err)
	}

	cmd := exec.Command("cmd.exe", "/c", "echo AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB && echo BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB && echo DONE")
	if err := startAgent(cmd, p); err != nil {
		t.Fatalf("start: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	out := readUntil(t, p.master, "DONE", deadline)
	t.Logf("=== BEFORE WIDTH CHANGE (%d bytes) ===\n%s", len(out), out)

	t.Logf("Widening from 40 to 80 columns...")
	if err := p.SetSize(80, 10); err != nil {
		t.Fatalf("resize: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	after := readUntil(t, p.master, "", deadline)
	t.Logf("=== AFTER WIDTH CHANGE (%d bytes) ===\n%s", len(after), after)
}
