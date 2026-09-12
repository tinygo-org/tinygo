// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (linux || darwin) && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"errors"
	"internal/itoa"
	"runtime"
	"sync/atomic"
	"syscall"
	_ "unsafe" // for go:linkname
)

// Process creation on a hosted OS uses posix_spawn(3) and not fork(2) plus
// execve(2). These targets run the threads scheduler and collect with Boehm,
// so a fork from Go gives the child one thread that holds the locks of the
// other threads, and the stop-the-world signal of the collector can arrive
// between the fork and the exec. posix_spawn does the clone and the exec
// inside libc, where no Go code runs.
//
// posix_spawn takes two POSIX objects whose shape is different for each OS, so
// the types are in exec_posix_spawn_linux.go and exec_posix_spawn_darwin.go.

// The only signal values guaranteed to be present in the os package on all
// systems are os.Interrupt (send the process an interrupt) and os.Kill (force
// the process to exit). On Windows, sending os.Interrupt to a process with
// os.Process.Signal is not implemented; it will return an error instead of
// sending a signal.
var (
	Interrupt Signal = syscall.SIGINT
	Kill      Signal = syscall.SIGKILL
)

// Give the child an empty signal mask. A blocked mask survives an exec, and
// the spawning thread can carry the signal of the collector blocked.
const _POSIX_SPAWN_SETSIGMASK = 0x08

// POSIX_SPAWN_SETPGROUP puts the child in the process group of
// posix_spawnattr_setpgroup. The value is 2 in lib/musl/include/spawn.h and in
// the <sys/spawn.h> of Darwin.
const _POSIX_SPAWN_SETPGROUP = 0x02

// Keep compatible with golang and always succeed and return new proc with pid.
func findProcess(pid int) (*Process, error) {
	return &Process{Pid: pid}, nil
}

func (p *Process) release() error {
	// NOOP for unix.
	p.Pid = -1
	// no need for a finalizer anymore
	runtime.SetFinalizer(p, nil)
	return nil
}

// ProcessState stores information about a process, as reported by Wait.
type ProcessState struct {
	pid    int                // The process's id.
	status syscall.WaitStatus // System-dependent status info.
	rusage *syscall.Rusage
}

// Pid returns the process id of the exited process.
func (p *ProcessState) Pid() int {
	return p.pid
}

func (p *ProcessState) String() string {
	if p == nil {
		return "<nil>"
	}
	status := p.status
	res := ""
	switch {
	case status.Exited():
		res = "exit status " + itoa.Itoa(status.ExitStatus())
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

// SysUsage returns system-dependent resource usage information about
// the exited process. Convert it to the appropriate underlying
// type, such as *syscall.Rusage on Unix, to access its contents.
func (p *ProcessState) SysUsage() interface{} {
	return p.rusage
}

func (p *ProcessState) Exited() bool {
	return p.status.Exited()
}

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
func (p *ProcessState) ExitCode() int {
	// return -1 if the process hasn't started.
	if p == nil || !p.status.Exited() {
		return -1
	}
	return p.status.ExitStatus()
}

// Wait waits for the Process to exit, and then returns a ProcessState
// describing its status and an error, if any.
func (p *Process) Wait() (*ProcessState, error) {
	if p.Pid == -1 {
		return nil, syscall.EINVAL
	}
	var status syscall.WaitStatus
	var rusage syscall.Rusage
	var wpid int
	var err error
	for {
		wpid, err = syscall.Wait4(p.Pid, &status, 0, &rusage)
		// The collector stops the world with a signal, so a thread in wait4
		// gets EINTR as a matter of course.
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		return nil, NewSyscallError("wait", err)
	}
	atomic.StoreInt32(&p.done, 1)
	return &ProcessState{pid: wpid, status: status, rusage: &rusage}, nil
}

// Signal sends a signal to the Process. Sending Interrupt on Windows is not
// implemented.
func (p *Process) Signal(sig Signal) error {
	if p.Pid == -1 {
		return errors.New("os: process already released")
	}
	if p.Pid == 0 {
		return errors.New("os: process not initialized")
	}
	if atomic.LoadInt32(&p.done) != 0 {
		return ErrProcessDone
	}
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	if err := syscall.Kill(p.Pid, s); err != nil {
		// Another goroutine can reap the process between the check above and
		// the kill. exec.CommandContext expects ErrProcessDone here.
		if err == syscall.ESRCH {
			return ErrProcessDone
		}
		return err
	}
	return nil
}

// Kill causes the Process to exit immediately. Kill does not wait until the
// Process has actually exited. This only kills the Process itself, not any
// other processes it may have started.
func (p *Process) Kill() error {
	return p.Signal(Kill)
}

// startProcess creates the child with posix_spawn instead of a fork and exec
// pair.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	if attr == nil {
		attr = new(ProcAttr)
	}
	if attr.Sys != nil {
		// Refuse by name every field that posix_spawn cannot express. Only
		// Setpgid and Pgid are honoured.
		if err := checkSysProcAttr(attr.Sys); err != nil {
			return nil, err
		}
	}

	pid, err := forkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return &Process{Pid: pid}, nil
}

// forkExec spawns the program at argv0 and returns its pid. It does not fork.
// posix_spawn reports a failed exec as its return value, so no status pipe is
// necessary.
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	if len(argv) == 0 {
		return 0, errors.New("exec: no argv")
	}
	if attr == nil {
		attr = new(ProcAttr)
	}

	argv0p, err := syscall.BytePtrFromString(argv0)
	if err != nil {
		return 0, err
	}
	argvp, err := syscall.SlicePtrFromStrings(argv)
	if err != nil {
		return 0, err
	}
	env := attr.Env
	if env == nil {
		// A nil Env means the environment of the parent.
		env = Environ()
	}
	envp, err := syscall.SlicePtrFromStrings(env)
	if err != nil {
		return 0, err
	}

	var fa spawnFileActions
	if errno := posix_spawn_file_actions_init(&fa); errno != 0 {
		return 0, syscall.Errno(errno)
	}
	defer posix_spawn_file_actions_destroy(&fa)

	var sa spawnAttr
	if errno := posix_spawnattr_init(&sa); errno != 0 {
		return 0, syscall.Errno(errno)
	}
	defer posix_spawnattr_destroy(&sa)

	var mask sigset
	if errno := posix_spawnattr_setsigmask(&sa, &mask); errno != 0 {
		return 0, syscall.Errno(errno)
	}

	flags := int16(_POSIX_SPAWN_SETSIGMASK)

	// Setpgid is the one SysProcAttr field that posix_spawn can express. A
	// Pgid of 0 makes a new group whose id is the pid of the child.
	if attr.Sys != nil && attr.Sys.Setpgid {
		if errno := posix_spawnattr_setpgroup(&sa, int32(attr.Sys.Pgid)); errno != 0 {
			return 0, syscall.Errno(errno)
		}
		flags |= _POSIX_SPAWN_SETPGROUP
	}

	if errno := posix_spawnattr_setflags(&sa, flags); errno != 0 {
		return 0, syscall.Errno(errno)
	}

	if attr.Dir != "" {
		dirp, err := syscall.BytePtrFromString(attr.Dir)
		if err != nil {
			return 0, err
		}
		// Darwin stores the pointer and not a copy of the path, and the
		// collector cannot see it. Keep the Go bytes alive until the spawn.
		defer runtime.KeepAlive(dirp)
		if errno := posix_spawn_file_actions_addchdir_np(&fa, dirp); errno != 0 {
			return 0, syscall.Errno(errno)
		}
	}

	defer runtime.KeepAlive(attr.Files)
	fds := make([]int, len(attr.Files))
	nextfd := len(fds)
	if nextfd < 3 {
		nextfd = 3
	}
	for i, f := range attr.Files {
		fd := ^uintptr(0)
		if f != nil {
			fd = f.Fd()
		}
		fds[i] = -1
		if fd != ^uintptr(0) {
			if fd >= 1<<31-1 {
				return 0, syscall.EBADF
			}
			fds[i] = int(fd)
			if int(fd) >= nextfd {
				nextfd = int(fd) + 1
			}
		}
	}

	// Save sources before an earlier action replaces or closes them.
	// See Go src/syscall/exec_linux.go, forkAndExecInChild, Pass 1.
	firstTemp := nextfd
	for i, fd := range fds {
		if fd >= 0 && fd < i {
			if nextfd >= 1<<31-1 {
				return 0, syscall.EINVAL
			}
			if errno := posix_spawn_file_actions_adddup2(&fa, int32(fd), int32(nextfd)); errno != 0 {
				return 0, syscall.Errno(errno)
			}
			fds[i] = nextfd
			nextfd++
		}
	}

	for i, fd := range fds {
		if fd == -1 {
			if err := addSpawnClose(&fa, int32(i)); err != nil {
				return 0, err
			}
			continue
		}
		// A dup2 onto the same descriptor clears FD_CLOEXEC and is not a
		// no-op, which is what an inherited os.Stdin needs.
		if errno := posix_spawn_file_actions_adddup2(&fa, int32(fd), int32(i)); errno != 0 {
			return 0, syscall.Errno(errno)
		}
	}

	// Close the standard descriptors that ProcAttr.Files does not name, which
	// is what syscall.forkAndExecInChild does in the standard library.
	for i := len(attr.Files); i < 3; i++ {
		if err := addSpawnClose(&fa, int32(i)); err != nil {
			return 0, err
		}
	}
	for fd := firstTemp; fd < nextfd; fd++ {
		if errno := posix_spawn_file_actions_addclose(&fa, int32(fd)); errno != 0 {
			return 0, syscall.Errno(errno)
		}
	}

	var childPid int32
	// ForkLock keeps a descriptor made without O_CLOEXEC out of a child that
	// another goroutine spawns at the same time.
	syscall.ForkLock.Lock()
	errno := posix_spawn(&childPid, argv0p, &fa, &sa, &argvp[0], &envp[0])
	syscall.ForkLock.Unlock()
	runtime.KeepAlive(argv0p)
	runtime.KeepAlive(argvp)
	runtime.KeepAlive(envp)
	if errno != 0 {
		return 0, syscall.Errno(errno)
	}

	return int(childPid), nil
}

// Bindings for the posix_spawn family. They use //go:linkname and not
// //export, because //export promises that pointer arguments do not escape,
// and the objects here outlive the call.

//go:linkname posix_spawn posix_spawn
func posix_spawn(pid *int32, path *byte, fa *spawnFileActions, sa *spawnAttr, argv **byte, envp **byte) int32

//go:linkname posix_spawn_file_actions_init posix_spawn_file_actions_init
func posix_spawn_file_actions_init(fa *spawnFileActions) int32

//go:linkname posix_spawn_file_actions_destroy posix_spawn_file_actions_destroy
func posix_spawn_file_actions_destroy(fa *spawnFileActions) int32

//go:linkname posix_spawn_file_actions_adddup2 posix_spawn_file_actions_adddup2
func posix_spawn_file_actions_adddup2(fa *spawnFileActions, fildes, newfildes int32) int32

//go:linkname posix_spawn_file_actions_addclose posix_spawn_file_actions_addclose
func posix_spawn_file_actions_addclose(fa *spawnFileActions, fildes int32) int32

// Present in musl since 1.1.24 and in macOS since 10.15.
//
//go:linkname posix_spawn_file_actions_addchdir_np posix_spawn_file_actions_addchdir_np
func posix_spawn_file_actions_addchdir_np(fa *spawnFileActions, path *byte) int32

//go:linkname posix_spawnattr_init posix_spawnattr_init
func posix_spawnattr_init(sa *spawnAttr) int32

//go:linkname posix_spawnattr_destroy posix_spawnattr_destroy
func posix_spawnattr_destroy(sa *spawnAttr) int32

//go:linkname posix_spawnattr_setflags posix_spawnattr_setflags
func posix_spawnattr_setflags(sa *spawnAttr, flags int16) int32

//go:linkname posix_spawnattr_setsigmask posix_spawnattr_setsigmask
func posix_spawnattr_setsigmask(sa *spawnAttr, mask *sigset) int32

//go:linkname posix_spawnattr_setpgroup posix_spawnattr_setpgroup
func posix_spawnattr_setpgroup(sa *spawnAttr, pgroup int32) int32

// unsupportedSysFieldError names the SysProcAttr field that this
// implementation cannot honour. It unwraps to ErrNotImplementedSys.
type unsupportedSysFieldError struct {
	field string
}

func (e *unsupportedSysFieldError) Error() string {
	return "os: SysProcAttr." + e.field + ": " + ErrNotImplementedSys.Error()
}

func (e *unsupportedSysFieldError) Unwrap() error {
	return ErrNotImplementedSys
}

func errUnsupportedSysField(field string) error {
	return &unsupportedSysFieldError{field: field}
}

// checkSysProcAttrCommon rejects every field that both Linux and Darwin
// declare and that posix_spawn cannot express. Setpgid and Pgid are absent,
// because forkExec honours them.
func checkSysProcAttrCommon(sys *syscall.SysProcAttr) error {
	switch {
	case sys.Chroot != "":
		return errUnsupportedSysField("Chroot")
	case sys.Credential != nil:
		return errUnsupportedSysField("Credential")
	case sys.Ptrace:
		return errUnsupportedSysField("Ptrace")
	case sys.Setsid:
		return errUnsupportedSysField("Setsid")
	case sys.Setctty:
		return errUnsupportedSysField("Setctty")
	case sys.Noctty:
		return errUnsupportedSysField("Noctty")
	case sys.Ctty != 0:
		return errUnsupportedSysField("Ctty")
	case sys.Foreground:
		return errUnsupportedSysField("Foreground")
	}
	return nil
}
