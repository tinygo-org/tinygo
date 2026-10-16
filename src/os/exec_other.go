//go:build (!aix && !android && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris) || baremetal || tinygo.wasm || nintendoswitch

package os

import (
	"syscall"
	"time"
)

var (
	Interrupt Signal = syscall.SIGINT
	Kill      Signal = syscall.SIGKILL
)

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

func (p *Process) wait() (*ProcessState, error) {
	return nil, ErrNotImplemented
}

func (p *Process) signal(_ Signal) error {
	return ErrNotImplemented
}

func (p *ProcessState) userTime() time.Duration {
	return 0
}

func (p *ProcessState) systemTime() time.Duration {
	return 0
}

func hostname() (string, error) {
	return "", ErrNotImplemented
}
