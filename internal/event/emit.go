package event

import (
	"net"
	"os"
	"time"
)

// writeDeadline bounds the socket attempt.
//
// A unix datagram send is not the "never blocks" call it is often assumed to
// be: when the receiver's buffer is full, Go's netpoller parks the write until
// the socket is writable. That is fine for a normal program and fatal here,
// because this code runs inside a hook on every tool call, and a scape that
// stopped draining would become a brake on the agent. The deadline turns the
// worst case from "as long as the reader is wedged" into two milliseconds,
// after which we take the file path instead.
const writeDeadline = 2 * time.Millisecond

// Emit sends one event, best effort, and tells you which path it took.
//
// The contract that matters to the caller is the one it does NOT express: this
// never blocks meaningfully, never panics, and its error is advisory. A hook
// must not fail the agent's turn because a decoration is not running.
func Emit(e Event) (viaSocket bool, err error) {
	if e.TS == 0 {
		e.TS = time.Now().UnixMilli()
	}
	line := Encode(e)

	if sendSock(e.Session, line) == nil {
		return true, nil
	}
	return false, appendSpool(e.Session, line)
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

// appendSpool writes the line to the session's JSON-lines file.
//
// One write(2) of one line under O_APPEND is what keeps concurrent hooks from
// interleaving halves of two events. POSIX guarantees the offset moves to EOF
// atomically but not that the write itself is indivisible; in practice a
// sub-page append to a local file is, on both darwin and linux, and MaxLine
// keeps every line far under a page. The reader is built to survive the case
// anyway -- it holds a partial trailing line rather than parsing it.
func appendSpool(session string, line []byte) error {
	if _, err := EnsureRunDir(); err != nil {
		return err
	}
	p, err := SpoolPath(session)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}
