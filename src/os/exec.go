package os

import "syscall"

type Signal interface {
	String() string
	Signal() // to distinguish from other Stringers
}

// Getpid returns the process id of the caller, or -1 if unavailable.
func Getpid() int {
	return syscall.Getpid()
}

// Getppid returns the process id of the caller's parent, or -1 if unavailable.
func Getppid() int {
	return syscall.Getppid()
}

type ProcAttr struct {
	Dir   string
	Env   []string
	Files []*File
	Sys   *syscall.SysProcAttr
}

type ProcessState struct {
}

func (p *ProcessState) String() string {
	return "" // TODO
}
func (p *ProcessState) Success() bool {
	return false // TODO
}

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
func (p *ProcessState) ExitCode() int {
	return -1 // TODO
}

type Process struct {
	Pid int
}

func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error) {
	return nil, &PathError{"fork/exec", name, ErrNotImplemented}
}

func (p *Process) Wait() (*ProcessState, error) {
	return nil, ErrNotImplemented
}

func (p *Process) Kill() error {
	return ErrNotImplemented
}

func (p *Process) Signal(sig Signal) error {
	return ErrNotImplemented
}
