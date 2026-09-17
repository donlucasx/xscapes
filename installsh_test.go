package main

// The site's install line is `curl -fsSL .../install.sh | sh`. A friend of
// the maker followed the previous line, `go install ...@latest`, and got
// `zsh: command not found: xscapes` (2026-09-17): go install drops the
// binary in ~/go/bin, which is on no Mac's PATH by default, and prints
// nothing about it. The script exists so that one command puts the binary
// where a shell will find it, on a machine with or without Go.
//
// These tests run the real script under /bin/sh against a local server
// that serves the real binary, with a HOME and a PATH of their own, so
// nothing here touches the tester's shell files.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildForInstallTest builds this package once, as the asset the release
// would carry for the machine running the test.
func buildForInstallTest(t *testing.T) (dir, asset string) {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go on PATH")
	}
	dir = t.TempDir()
	asset = "xscapes-" + runtime.GOOS + "-" + runtime.GOARCH
	out, err := exec.Command(goTool, "build", "-o", filepath.Join(dir, asset), ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return dir, asset
}

type installScriptRun struct {
	home, bin, rc  string
	stdout, stderr string
	err            error
}

// runInstallScript runs site/install.sh with HOME set to a fresh directory, the
// given SHELL, the given PATH and the given download base, and reports what
// it wrote and said.
func runInstallScript(t *testing.T, base, shell, path string, home string) installScriptRun {
	t.Helper()
	if home == "" {
		home = t.TempDir()
	}
	script, err := filepath.Abs(filepath.Join("site", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", script)
	cmd.Env = []string{
		"HOME=" + home,
		"SHELL=" + shell,
		"PATH=" + path,
		"XSCAPES_DOWNLOAD_BASE=" + base,
	}
	var so, se strings.Builder
	cmd.Stdout, cmd.Stderr = &so, &se
	err = cmd.Run()
	return installScriptRun{
		home:   home,
		bin:    filepath.Join(home, ".local", "bin", "xscapes"),
		rc:     filepath.Join(home, ".zshrc"),
		stdout: so.String(),
		stderr: se.String(),
		err:    err,
	}
}

// systemPath is a PATH with the tools the script needs and NOT the go
// toolchain, which lives under /opt/homebrew or /usr/local/go on a Mac.
const systemPath = "/usr/bin:/bin"

func TestTheInstallScriptPutsTheBinaryOnThePath(t *testing.T) {
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("no /bin/sh")
	}
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("no curl")
	}
	dir, asset := buildForInstallTest(t)
	srv := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer srv.Close()
	if exec.Command("/bin/sh", "-c", "PATH="+systemPath+" command -v go").Run() == nil {
		t.Fatalf("the restricted PATH %q can see go; the fallback tests would not be testing the fallback", systemPath)
	}

	t.Run("download, then the rc line once, then the full-path next step", func(t *testing.T) {
		r := runInstallScript(t, srv.URL, "/bin/zsh", systemPath, "")
		if r.err != nil {
			t.Fatalf("install.sh: %v\nstdout:\n%s\nstderr:\n%s", r.err, r.stdout, r.stderr)
		}
		st, err := os.Stat(r.bin)
		if err != nil {
			t.Fatalf("no binary at %s: %v\n%s", r.bin, err, r.stdout)
		}
		if st.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable: %v", r.bin, st.Mode())
		}
		if out, err := exec.Command(r.bin, "-h").CombinedOutput(); err != nil {
			t.Fatalf("the installed binary does not run: %v\n%s", err, out)
		}
		rc, err := os.ReadFile(r.rc)
		if err != nil {
			t.Fatalf("no rc written: %v\n%s", err, r.stdout)
		}
		line := `export PATH="$HOME/.local/bin:$PATH"`
		if n := strings.Count(string(rc), line); n != 1 {
			t.Fatalf("rc carries the PATH line %d times, want 1:\n%s", n, rc)
		}
		// The current window's PATH is not changed by a child process, so the
		// next step has to work AS PRINTED, without a new terminal.
		if !strings.Contains(r.stdout, "~/.local/bin/xscapes install claude --apply") {
			t.Fatalf("the next step is not printed with a path that works now:\n%s", r.stdout)
		}
		if !strings.Contains(r.stdout, asset) {
			t.Fatalf("the report does not say which asset it installed:\n%s", r.stdout)
		}

		// Run it again: the binary is replaced and the rc is not touched twice.
		again := runInstallScript(t, srv.URL, "/bin/zsh", systemPath, r.home)
		if again.err != nil {
			t.Fatalf("second install.sh: %v\n%s\n%s", again.err, again.stdout, again.stderr)
		}
		rc, _ = os.ReadFile(r.rc)
		if n := strings.Count(string(rc), line); n != 1 {
			t.Fatalf("after a second run the rc carries the PATH line %d times, want 1:\n%s", n, rc)
		}
	})

	t.Run("already on PATH: no rc edit, bare next step", func(t *testing.T) {
		home := t.TempDir()
		path := filepath.Join(home, ".local", "bin") + ":" + systemPath
		r := runInstallScript(t, srv.URL, "/bin/zsh", path, home)
		if r.err != nil {
			t.Fatalf("install.sh: %v\n%s\n%s", r.err, r.stdout, r.stderr)
		}
		if _, err := os.Stat(r.rc); err == nil {
			t.Fatalf("the rc was written although the directory was already on PATH")
		}
		if !strings.Contains(r.stdout, "\n  xscapes install claude --apply") {
			t.Fatalf("expected the bare next step:\n%s", r.stdout)
		}
		if strings.Contains(r.stdout, "~/.local/bin/xscapes install") {
			t.Fatalf("a full path was printed although the bare command works:\n%s", r.stdout)
		}
	})

	t.Run("bash on this OS gets its own rc file", func(t *testing.T) {
		r := runInstallScript(t, srv.URL, "/bin/bash", systemPath, "")
		if r.err != nil {
			t.Fatalf("install.sh: %v\n%s\n%s", r.err, r.stdout, r.stderr)
		}
		want := ".bashrc"
		if runtime.GOOS == "darwin" {
			want = ".bash_profile" // Terminal.app opens LOGIN shells; bash reads .bash_profile, not .bashrc
		}
		if _, err := os.Stat(filepath.Join(r.home, want)); err != nil {
			t.Fatalf("no %s written for bash: %v\n%s", want, err, r.stdout)
		}
	})

	t.Run("no asset: go install into the same directory", func(t *testing.T) {
		// A stub `go` that records how it was called and writes a runnable
		// binary where GOBIN points; the real command was run by hand.
		stubs := t.TempDir()
		log := filepath.Join(stubs, "go.log")
		stub := "#!/bin/sh\nprintf '%s\\n' \"GOBIN=$GOBIN\" \"$@\" > " + log + "\n" +
			"printf '#!/bin/sh\\nexit 0\\n' > \"$GOBIN/xscapes\" && chmod +x \"$GOBIN/xscapes\"\n"
		if err := os.WriteFile(filepath.Join(stubs, "go"), []byte(stub), 0o755); err != nil {
			t.Fatal(err)
		}
		empty := httptest.NewServer(http.NotFoundHandler())
		defer empty.Close()
		home := t.TempDir()
		r := runInstallScript(t, empty.URL, "/bin/zsh", stubs+":"+systemPath, home)
		if r.err != nil {
			t.Fatalf("install.sh: %v\n%s\n%s", r.err, r.stdout, r.stderr)
		}
		got, err := os.ReadFile(log)
		if err != nil {
			t.Fatalf("the stub go was never called: %v\n%s\n%s", err, r.stdout, r.stderr)
		}
		want := "GOBIN=" + filepath.Join(home, ".local", "bin") + "\ninstall\ngithub.com/donlucasx/xscapes@latest\n"
		if string(got) != want {
			t.Fatalf("go was called with\n%s\nwant\n%s", got, want)
		}
		if _, err := os.Stat(r.bin); err != nil {
			t.Fatalf("no binary after the go fallback: %v", err)
		}
	})

	t.Run("no asset and no go: says what to install and fails", func(t *testing.T) {
		empty := httptest.NewServer(http.NotFoundHandler())
		defer empty.Close()
		r := runInstallScript(t, empty.URL, "/bin/zsh", systemPath, "")
		if r.err == nil {
			t.Fatalf("install.sh succeeded with nothing to install:\n%s", r.stdout)
		}
		if !strings.Contains(r.stderr, "brew install go") {
			t.Fatalf("the failure does not say how to get Go:\n%s", r.stderr)
		}
		if _, err := os.Stat(r.bin); err == nil {
			t.Fatalf("a binary exists after a failed install")
		}
	})
}
