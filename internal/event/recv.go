package event

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// ErrBusy means another engine already serves this session.
var ErrBusy = errors.New("another asciiscapes is already listening for this session")

// readBuffer sizes the kernel's receive queue.
//
// This is load-bearing and it is invisible when wrong: with the default queue
// (a few kilobytes) a burst of parallel tool calls fills it, and the *emitter*
// pays -- each hook blocks to its deadline and then writes a file. Nothing
// breaks, the scene still animates, and every tool call in the session has
// quietly become two milliseconds slower. A megabyte is roughly five hundred
// events in flight, which no real burst approaches.
const readBuffer = 1 << 20

// spoolPoll is how often the fallback file is checked. The socket is the fast
// path; the file only carries events emitted while no engine was up, so a
// quarter second of latency on it costs nothing a viewer can see.
const spoolPoll = 250 * time.Millisecond

// Bus receives events for one session.
type Bus struct {
	C chan Event

	session string
	sock    string
	conn    *net.UnixConn
	spool   *os.File
	off     int64
	part    []byte
	stop    chan struct{}
	// spoolDone lets Close wait for the spool reader instead of reaching into
	// its state while it is still running.
	spoolDone chan struct{}

	// dropped counts events discarded because the consumer fell behind. A
	// drop is preferable to a stall, but it must be visible: the number is
	// how you tell "the agent went quiet" from "we stopped listening".
	dropped atomic.Int64
	// bad counts lines that would not parse -- an adapter writing a format
	// we did not agree to.
	bad atomic.Int64
}

// Listen binds the session's socket and starts following its spool file.
func Listen(session string) (*Bus, error) {
	if _, err := EnsureRunDir(); err != nil {
		return nil, err
	}
	p, err := SockPath(session)
	if err != nil {
		return nil, err
	}

	conn, err := bind(p)
	if err != nil {
		return nil, err
	}
	// Best effort: bind() takes its mode from the umask, so tighten it. The
	// 0700 run directory is the real guarantee.
	_ = os.Chmod(p, 0o600)
	_ = conn.SetReadBuffer(readBuffer)

	b := &Bus{
		C:       make(chan Event, 256),
		session: session,
		sock:    p,
		conn:      conn,
		stop:      make(chan struct{}),
		spoolDone: make(chan struct{}),
	}

	if sp, err := SpoolPath(session); err == nil {
		if f, err := os.Open(sp); err == nil {
			// Start at the end. Everything already in the file was written
			// while no engine was listening, which makes it history, not
			// news -- replaying it would rewrite the last hour of weather
			// into the next two seconds. `asciiscapes replay` exists for
			// when you actually want the history.
			if off, err := f.Seek(0, io.SeekEnd); err == nil {
				b.spool, b.off = f, off
			} else {
				f.Close()
			}
		}
	}

	go b.readSock()
	go b.readSpool()
	return b, nil
}

// bind takes the socket, clearing a stale one left by a killed engine.
//
// A leftover socket inode is indistinguishable from a live one by looking at
// it, so we ask: dial it and try to write. A live engine accepts the datagram;
// a dead one's inode answers ECONNREFUSED. Only then do we remove the file.
// Probing before removing is what stops two panes starting together from each
// deleting the other's socket.
func bind(p string) (*net.UnixConn, error) {
	addr, err := net.ResolveUnixAddr("unixgram", p)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUnixgram("unixgram", addr)
	if err == nil {
		return conn, nil
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
	return net.ListenUnixgram("unixgram", addr)
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

func (b *Bus) readSock() {
	buf := make([]byte, MaxLine+64)
	for {
		select {
		case <-b.stop:
			return
		default:
		}
		// A deadline is what lets Close() actually stop this goroutine:
		// ReadFrom on a closed conn returns an error, but without a deadline
		// we would sit in the syscall until a datagram arrived.
		_ = b.conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := b.conn.ReadFromUnix(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}
		if n == 0 {
			continue
		}
		b.deliver(buf[:n])
	}
}

func (b *Bus) readSpool() {
	defer close(b.spoolDone)
	t := time.NewTicker(spoolPoll)
	defer t.Stop()
	for {
		select {
		case <-b.stop:
			return
		case <-t.C:
			b.drainSpool()
		}
	}
}

func (b *Bus) drainSpool() {
	if b.spool == nil {
		// The file may not have existed when we started. Pick it up if a
		// fallback write has since created it.
		sp, err := SpoolPath(b.session)
		if err != nil {
			return
		}
		f, err := os.Open(sp)
		if err != nil {
			return
		}
		b.spool, b.off = f, 0
	}

	st, err := b.spool.Stat()
	if err != nil {
		return
	}
	switch {
	case st.Size() < b.off:
		// Truncated or replaced underneath us. Re-reading from zero would
		// replay the whole file; starting over at the new end is the only
		// choice that cannot double-deliver.
		b.off = st.Size()
		b.part = nil
		return
	case st.Size() == b.off:
		return
	}

	buf := make([]byte, st.Size()-b.off)
	n, err := b.spool.ReadAt(buf, b.off)
	if n == 0 {
		return
	}
	if err != nil && err != io.EOF {
		return
	}
	b.off += int64(n)
	data := append(b.part, buf[:n]...)

	for {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			// Hold the remainder. A writer can be mid-append; parsing half
			// an event would count as a bad line and lose a real one.
			b.part = append(b.part[:0], data...)
			return
		}
		b.deliver(data[:i])
		data = data[i+1:]
	}
}

func (b *Bus) deliver(line []byte) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return
	}
	e, err := Decode(line)
	if err != nil {
		b.bad.Add(1)
		return
	}
	select {
	case b.C <- e:
	default:
		b.dropped.Add(1)
	}
}

// Stats reports what the bus had to throw away.
func (b *Bus) Stats() (dropped, bad int64) {
	return b.dropped.Load(), b.bad.Load()
}

// Close stops the bus and removes the socket, so the next engine binds without
// having to probe a corpse.
func (b *Bus) Close() error {
	select {
	case <-b.stop:
		return nil // already closed
	default:
		close(b.stop)
	}
	// Wait for the spool reader to stop before touching its file: it opens
	// b.spool lazily, so closing it from here is a race on the field itself.
	<-b.spoolDone
	err := b.conn.Close()
	if b.spool != nil {
		b.spool.Close()
	}
	os.Remove(b.sock)
	return err
}
