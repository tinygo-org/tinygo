//go:build !aix && !android && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris
// +build !aix,!android,!freebsd,!linux,!netbsd,!openbsd,!plan9,!solaris

package os

func findProcess(pid int) (*Process, error) {
	return &Process{Pid: pid}, nil
}

func (p *Process) release() error {
	p.Pid = -1
	return nil
}

func forkExec(_ string, _ []string, _ *ProcAttr) (pid int, err error) {
	return 0, ErrNotImplemented
}

func startProcess(_ string, _ []string, _ *ProcAttr) (proc *Process, err error) {
	return &Process{Pid: 0}, ErrNotImplemented
}
