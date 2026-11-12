package os

import (
	"errors"
	"internal/itoa"
	"syscall"
	"time"
)

var (
	ErrNotImplementedDir   = errors.New("directory setting not implemented")
	ErrNotImplementedSys   = errors.New("sys setting not implemented")
	ErrNotImplementedFiles = errors.New("files setting not implemented")
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
	pid    int
	status syscall.WaitStatus
	rusage *syscall.Rusage
}

func (p *ProcessState) Pid() int {
	return p.pid
}

func (p *ProcessState) String() string {
	if p == nil {
		return "<nil>"
	}
	status := p.Sys().(syscall.WaitStatus)
	res := ""
	switch {
	case status.Exited():
		code := status.ExitStatus()
		res = "exit status " + itoa.Itoa(code)
	case status.Signaled():
		res = "signal: " + status.Signal().String()
	case status.Stopped():
		res = "stop signal: " + status.StopSignal().String()
		if status.StopSignal() == syscall.SIGTRAP && status.TrapCause() != 0 {
			res += " (trap " + itoa.Itoa(status.TrapCause()) + ")"
		}
	case status.Continued():
		res = "continued"
	}
	if status.CoreDump() {
		res += " (core dumped)"
	}
	return res
}

func (p *ProcessState) Success() bool {
	return p.status.ExitStatus() == 0
}

// Sys returns system-dependent exit information about
// the process. Convert it to the appropriate underlying
// type, such as syscall.WaitStatus on Unix, to access its contents.
func (p *ProcessState) Sys() interface{} {
	return p.status
}

func (p *ProcessState) SysUsage() interface{} {
	return p.rusage
}

func (p *ProcessState) UserTime() time.Duration {
	return p.userTime()
}

func (p *ProcessState) SystemTime() time.Duration {
	return p.systemTime()
}

func (p *ProcessState) Exited() bool {
	return p.status.Exited()
}

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
func (p *ProcessState) ExitCode() int {
	// Return -1 if the process hasn't exited or was terminated by a signal.
	if !p.Exited() {
		return -1
	}
	return p.status.ExitStatus()
}

type Process struct {
	Pid int
}

// StartProcess starts a new process with the program, arguments and attributes specified by name, argv and attr.
// Arguments to the process (os.Args) are passed via argv.
func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error) {
	return startProcess(name, argv, attr)
}

func (p *Process) Wait() (*ProcessState, error) {
	if p.Pid == -1 {
		return nil, syscall.EINVAL
	}
	return p.wait()
}

func (p *Process) Kill() error {
	return p.Signal(Kill)
}

func (p *Process) Signal(sig Signal) error {
	return p.signal(sig)
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

// FindProcess looks for a running process by its pid.
// Keep compatibility with golang and always succeed and return new proc with pid on Linux.
func FindProcess(pid int) (*Process, error) {
	return findProcess(pid)
}
