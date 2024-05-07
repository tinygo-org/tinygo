package os

import (
	"errors"
	"syscall"
)

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

// ErrProcessDone indicates a Process has finished.
var ErrProcessDone = errors.New("os: process already finished")

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

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
func (p *ProcessState) ExitCode() int {
	return -1 // TODO
}

type Process struct {
	Pid int
}

func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error) {
	return nil, &PathError{Op: "fork/exec", Path: name, Err: ErrNotImplemented}
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

func Ignore(sig ...Signal) {
	// leave all the signals unaltered
	return
}

// Release releases any resources associated with the Process p,
// rendering it unusable in the future.
// Release only needs to be called if Wait is not.
func (p *Process) Release() error {
	return p.release()
}

// Keep compatibility with golang and always succeed and return new proc with pid on Linux.
func FindProcess(pid int) (*Process, error) {
	return findProcess(pid)
}
