package os

import (
	"errors"
	"syscall"
)

// Errors StartProcess returns for a ProcAttr that it cannot honour. On a
// hosted OS only ErrNotImplementedSys is reachable. The other two stay because
// they are part of the exported API of this package.
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

type Process struct {
	Pid int

	// done reports whether Wait reaped this process. A signal to a reaped pid
	// is unsafe, because the number can belong to an unrelated process.
	done int32
}

// StartProcess starts a new process with the program, arguments and attributes specified by name, argv and attr.
// Arguments to the process (os.Args) are passed via argv.
func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error) {
	return startProcess(name, argv, attr)
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
