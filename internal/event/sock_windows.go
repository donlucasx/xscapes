//go:build windows

package event

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// THE SPOOL IS THE ONLY PATH ON WINDOWS.
//
// Go's net package has no AF_UNIX support on Windows (the OS has AF_UNIX since
// 1803, the runtime does not expose it), so sendSock fails by construction and
// every event goes to the session's JSON-lines spool, which the bus already
// follows -- that is the same two-path design the other platforms have, with
// the fast path simply absent. The cost a viewer can see is spoolPoll: 250 ms
// between an event being written and the scene noticing it, against a 12 fps
// scape.

// errNoSock is what sendSock reports: the caller treats every error the same
// way and falls to the spool, so the only thing this has to be is non-nil.
var errNoSock = errors.New("unix sockets are not available on windows; using the spool")

type noSock struct{}

// read blocks until the bus is closed. Returning an error immediately would
// end readSock at once and the goroutine would spin down silently, which is
// fine too -- but this way the shape matches the unix one: one read loop, one
// Close, nothing that looks like a failure.
func (noSock) read(buf []byte) (int, error)    { time.Sleep(time.Second); return 0, errNoSock }
func (noSock) setReadDeadline(time.Time) error { return nil }
func (noSock) Close() error                    { return nil }

func sendSock(session string, line []byte) error { return errNoSock }

// sessionTracker tracks which sessions are currently bound so bind() can
// return ErrBusy when a second listener claims the same one. Two mechanisms
// cooperate: an in-memory map catches same-process duplicates, and a lock
// file in the run directory catches cross-process ones (a crashed process
// leaves a stale lock that expires after staleLockTimeout).
var (
	mu     sync.Mutex
	active = make(map[string]struct{}) // tag -> present
)

const staleLockTimeout = 10 * time.Second

// lockInfo extracts the session tag from a socket path and returns the path of
// the per-session busy-lock file. p has the shape <run_dir>/<tag>.sock.
func lockInfo(p string) (tag string, lockPath string, err error) {
	base := filepath.Base(p)
	if len(base) < 5 || base[len(base)-5:] != ".sock" {
		return "", "", errors.New("unexpected socket path shape")
	}
	tag = base[:len(base)-5]
	d, err := EnsureRunDir()
	if err != nil {
		return "", "", err
	}
	return tag, filepath.Join(d, tag+".busy"), nil
}

// bind claims a session. It returns ErrBusy when another engine (this process
// or a sibling) already holds the session, mirroring the unix socket-bind
// semantics that the spool alone cannot enforce.
func bind(p string) (sockConn, error) {
	tag, lockPath, err := lockInfo(p)
	if err != nil {
		return nil, err
	}

	mu.Lock()
	if _, held := active[tag]; held {
		mu.Unlock()
		return nil, ErrBusy
	}
	active[tag] = struct{}{}
	mu.Unlock()

	// Cross-process: create an exclusive lock file. If it already exists and is
	// fresh, another process holds this session. If it's stale (crashed
	// process), overwrite it -- the in-memory map above only covers one process.
	st, statErr := os.Stat(lockPath)
	if statErr == nil {
		if time.Since(st.ModTime()) < staleLockTimeout {
			// Someone else has it; unwind the in-memory claim and report busy.
			mu.Lock()
			delete(active, tag)
			mu.Unlock()
			return nil, ErrBusy
		}
		// Stale lock: fall through to overwrite it.
	}

	if err := os.WriteFile(lockPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0o600); err != nil {
		mu.Lock()
		delete(active, tag)
		mu.Unlock()
		return nil, err
	}

	return noSock{}, nil
}

// release unclaims a session and removes its lock file. Called from Bus.Close.
func release(p string) {
	tag, _, err := lockInfo(p)
	if err != nil {
		return
	}
	mu.Lock()
	delete(active, tag)
	mu.Unlock()
	os.Remove(filepath.Join(filepath.Dir(p), tag+".busy"))
}
