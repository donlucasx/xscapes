package event

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donlucasx/xscapes/internal/envx"
)

// Home is where everything lives. The brief locks this to
// ~/.config/xscapes/, and the real argument for honouring it exactly is
// not convention: the emitter and the engine are separate processes that
// share no context but the environment, so the path has to be derivable from
// $HOME alone by both. Two derivation rules that can disagree is a silent,
// undebuggable failure -- the hook writes somewhere the scape is not reading
// and the scene simply never moves.
func Home() (string, error) {
	if v := envx.Lookup("HOME"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "xscapes"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "xscapes"), nil
}

// RunDir holds the per-session sockets and spool files.
func RunDir() (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "run"), nil
}

// EnsureRunDir creates it 0700. The directory is the access control: a unix
// socket created by bind() takes its mode from the umask (0755 under the
// common umask 022), and there is an unavoidable window between bind and
// chmod, so the socket's own mode is a second line and not the first.
func EnsureRunDir() (string, error) {
	d, err := RunDir()
	if err != nil {
		return "", err
	}
	return d, os.MkdirAll(d, 0o700)
}

// maxSunPath is the usable sun_path on darwin. The struct field is 104 bytes,
// one of which is the terminator; linux allows 108. Using the smaller keeps a
// path that works on the dev machine working everywhere.
const maxSunPath = 103

// ErrPathTooLong means the socket path will not fit in sockaddr_un.
var ErrPathTooLong = errors.New("socket path too long for sockaddr_un; set XSCAPES_HOME to something shorter")

// Tag reduces a session id to the short, filesystem-safe token that names its
// socket and spool.
//
// A HASH of the whole id, not a prefix of it. A prefix is only unique if the
// ids are, and nothing in the protocol promises that: two adapters numbering
// their sessions "session-1" and "session-2" would have collided on the same
// eight characters, silently sharing one socket, and each scape would have
// shown the other's work. Hashing also makes the sanitising free -- a session
// of "../../.ssh/id_rsa" cannot become a path when the output is hex.
func Tag(session string) string {
	if session == "" {
		return "anon"
	}
	sum := sha256.Sum256([]byte(session))
	return hex.EncodeToString(sum[:6])
}

// Short is a human-facing abbreviation of a session id, for messages. It is
// never a path: use Tag for that.
func Short(session string) string {
	if session == "" {
		return "(none)"
	}
	if len(session) > 8 {
		return session[:8]
	}
	return session
}

// SockPath is where the engine for this session listens.
func SockPath(session string) (string, error) {
	d, err := RunDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(d, Tag(session)+".sock")
	if len(p) > maxSunPath {
		return p, ErrPathTooLong
	}
	return p, nil
}

// SpoolPath is the JSON-lines fallback for this session, used whenever the
// socket is not there. It doubles as the replay log.
func SpoolPath(session string) (string, error) {
	d, err := RunDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, Tag(session)+".jsonl"), nil
}

// currentPath names the file holding the most recently started session.
func currentPath() (string, error) {
	d, err := RunDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "current"), nil
}

// SetCurrent records a session as the newest one. A scape pane started by
// hand, in a terminal that never saw the agent's environment, has no other way
// to find out which session to show.
//
// It makes the run directory first. It used to assume the directory was
// there, and on a fresh machine it is NOT: the first SessionStart hook is the
// first thing that ever writes under run/, so the pointer failed with ENOENT,
// the error was discarded, and the scape above never bound -- the spool
// created the directory a moment later, so the SECOND session worked, which
// is why no machine that had run xscapes once ever showed it (found
// 2026-09-17 by running the probes under a state directory that had never
// existed).
func SetCurrent(session string) error {
	if _, err := EnsureRunDir(); err != nil {
		return err
	}
	p, err := currentPath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(session), 0o600)
}

// Current reads it back. Empty is a normal answer, not an error.
func Current() string {
	p, err := currentPath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// CurrentSince is the session pointer only if it was WRITTEN after t, which
// is how a launcher tells the session it just started from the one that ran
// before it. The launchers used to wait for an id that DIFFERED from the
// pointer at launch; a resumed session (--resume, --continue) writes the
// same id again, so it was never bound: a restart by id showed the watcher
// instead of the hooks, and the next SessionStart from another window could
// bind the scape to that window's session instead (F1, session 38).
func CurrentSince(t time.Time) string {
	p, err := currentPath()
	if err != nil {
		return ""
	}
	st, err := os.Stat(p)
	if err != nil || !st.ModTime().After(t) {
		return ""
	}
	return Current()
}

// SessionFromEnv is how a scape launched inside the agent's own environment
// binds to the right session with no configuration at all. Verified on
// 2026-08-30: Claude Code exports CLAUDE_CODE_SESSION_ID to every child, and
// its value equals the transcript's basename.
func SessionFromEnv() string {
	if v := strings.TrimSpace(envx.Lookup("SESSION")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("CLAUDE_CODE_SESSION_ID")); v != "" {
		return v
	}
	return ""
}
