//go:build !windows

package host

import (
	"os/exec"
	"syscall"
)

// startAgent hands the slave to the child as its controlling terminal: the
// session leader bit is what makes the slave its ctty, so keystrokes on it
// become signals and a Ctrl-C the user types reaches the agent, not the host.
func startAgent(cmd *exec.Cmd, p *pty) error {
	cmd.Stdin, cmd.Stdout, cmd.Stderr = p.slave, p.slave, p.slave
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
	cmd.SysProcAttr.Setctty = true
	if err := cmd.Start(); err != nil {
		return err
	}
	// The child owns the slave now. Holding it open here would hide the EOF
	// when the child exits, and the output loop would never finish.
	p.slave.Close()
	p.slave = nil
	return nil
}
