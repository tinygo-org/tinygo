//go:build (!aix && !android && !darwin && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris) || baremetal || tinygo.wasm || nintendoswitch

package os

import "syscall"

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

// ProcessState is a placeholder on targets that have no process model.
type ProcessState struct {
}

func (p *ProcessState) String() string {
	return "" // TODO
}
func (p *ProcessState) Success() bool {
	return false // TODO
}

// Sys returns system-dependent exit information about
// the process. Convert it to the appropriate underlying
// type, such as syscall.WaitStatus on Unix, to access its contents.
func (p *ProcessState) Sys() interface{} {
	return nil // TODO
}

func (p *ProcessState) Exited() bool {
	return false // TODO
}

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
func (p *ProcessState) ExitCode() int {
	return -1 // TODO
}

func (p *Process) Wait() (*ProcessState, error) {
	if p.Pid == -1 {
		return nil, syscall.EINVAL
	}
	return nil, ErrNotImplemented
}

func (p *Process) Kill() error {
	return ErrNotImplemented
}

func (p *Process) Signal(sig Signal) error {
	return ErrNotImplemented
}

func forkExec(_ string, _ []string, _ *ProcAttr) (pid int, err error) {
	return 0, ErrNotImplemented
}

func startProcess(_ string, _ []string, _ *ProcAttr) (proc *Process, err error) {
	return &Process{Pid: 0}, ErrNotImplemented
}
