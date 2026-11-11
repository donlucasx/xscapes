//go:build !windows

package event

import (
	"net"
	"os"
	"time"
)

// THE SOCKET: a unix datagram socket per session, with the JSON-lines spool as
// the fallback. Windows has no AF_UNIX in Go's net package, so that build runs
// on the spool alone (sock_windows.go) -- the same contract, a quarter of a
// second slower.

type unixSock struct{ *net.UnixConn }

func (u unixSock) read(buf []byte) (int, error) {
	n, _, err := u.ReadFromUnix(buf)
	return n, err
}

func (u unixSock) setReadDeadline(t time.Time) error { return u.SetReadDeadline(t) }

// release drops whatever per-session bookkeeping a bound socket carried.
// The unix bus holds nothing but the conn -- Close() is enough -- but the
// Windows bus keeps an active-session map and a busy-lock file, so recv.go
// calls this on every Close and both builds define it.
func release(string) {}

// bind takes the socket, clearing a stale one left by a killed engine.
//
// A leftover socket inode is indistinguishable from a live one by looking at
// it, so we ask: dial it and try to write. A live engine accepts the datagram;
// a dead one's inode answers ECONNREFUSED. Only then do we remove the file.
// Probing before removing is what stops two panes starting together from each
// deleting the other's socket.
func bind(p string) (sockConn, error) {
	addr, err := net.ResolveUnixAddr("unixgram", p)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUnixgram("unixgram", addr)
	if err == nil {
		_ = conn.SetReadBuffer(readBuffer)
		return unixSock{conn}, nil
	}
	if _, statErr := os.Stat(p); statErr != nil {
		return nil, err // not an in-use problem; report the original
	}

	// Reclaiming a stale socket means unlinking a file, and unlinking a file
	// somebody else is listening on is the worst outcome available here. Two
	// scapes starting together could each probe, each see the other not yet
	// bound, and each remove the other's socket. Serialise the whole
	// probe-remove-bind with an exclusive create, so only one process is ever
	// in that window.
	lock := p + ".lock"
	lf, lerr := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if lerr != nil {
		// Someone else is mid-reclaim. If the lock is stale, take it over.
		if st, sErr := os.Stat(lock); sErr == nil && time.Since(st.ModTime()) > 10*time.Second {
			os.Remove(lock)
			lf, lerr = os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		}
		if lerr != nil {
			return nil, ErrBusy
		}
	}
	defer func() { lf.Close(); os.Remove(lock) }()

	// Re-probe inside the lock: the holder may have finished binding since we
	// first looked.
	if probeAlive(p) {
		return nil, ErrBusy
	}
	if rmErr := os.Remove(p); rmErr != nil {
		return nil, err
	}
	c, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		return nil, err
	}
	_ = c.SetReadBuffer(readBuffer)
	return unixSock{c}, nil
}

// probeAlive reports whether something is actually listening at p.
func probeAlive(p string) bool {
	addr, err := net.ResolveUnixAddr("unixgram", p)
	if err != nil {
		return false
	}
	c, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return false
	}
	defer c.Close()
	_ = c.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
	// A zero-length datagram is a valid send and decodes to nothing, so a
	// live engine counts it as a bad line at worst.
	_, err = c.Write(nil)
	return err == nil
}

// sendSock tries the live engine. Every failure mode here is a normal state of
// the world, not an anomaly: no engine has ever run (ENOENT), one ran and was
// killed (ECONNREFUSED on a stale inode), or one is alive but behind
// (deadline). All three mean the same thing to us -- use the file.
func sendSock(session string, line []byte) error {
	p, err := SockPath(session)
	if err != nil {
		return err
	}
	addr, err := net.ResolveUnixAddr("unixgram", p)
	if err != nil {
		return err
	}
	c, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
		return err
	}
	_, err = c.Write(line)
	return err
}
