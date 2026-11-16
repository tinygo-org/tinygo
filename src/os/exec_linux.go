// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// The only signal values guaranteed to be present in the os package on all
// systems are os.Interrupt (send the process an interrupt) and os.Kill (force
// the process to exit). On Windows, sending os.Interrupt to a process with
// os.Process.Signal is not implemented; it will return an error instead of
// sending a signal.
var (
	Interrupt Signal = syscall.SIGINT
	Kill      Signal = syscall.SIGKILL
)

// processState stops Signal from reaching a reused pid after Wait reaps the child.
type processState struct {
	done  atomic.Bool
	sigMu sync.RWMutex
}

// Keep compatible with golang and always succeed and return new proc with pid on Linux.
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

func (p *Process) wait() (*ProcessState, error) {
	// wait4 with a pid of 0 or less waits for any child in a process group, see wait(2).
	if p.Pid <= 0 {
		return nil, errors.New("os: process not initialized")
	}
	if err := p.blockUntilWaitable(); err != nil {
		return nil, err
	}
	p.done.Store(true)
	p.sigMu.Lock()
	p.sigMu.Unlock()

	var status syscall.WaitStatus
	pid, err := syscall.Wait4(p.Pid, &status, 0, nil)
	for err == syscall.EINTR {
		pid, err = syscall.Wait4(p.Pid, &status, 0, nil)
	}
	if err != nil {
		return nil, NewSyscallError("wait", err)
	}
	return &ProcessState{pid: pid, status: status}, nil
}

// blockUntilWaitable waits for the child to exit but leaves it for wait4 to reap.
func (p *Process) blockUntilWaitable() error {
	const _P_PID = 1
	var siginfo [16]uint64
	for {
		_, _, errno := syscall.Syscall6(syscall.SYS_WAITID, _P_PID, uintptr(p.Pid), uintptr(unsafe.Pointer(&siginfo[0])), syscall.WEXITED|syscall.WNOWAIT, 0, 0)
		switch errno {
		case 0:
			return nil
		case syscall.EINTR:
			continue
		default:
			return NewSyscallError("waitid", errno)
		}
	}
}

func (p *Process) signal(sig Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	if p.Pid == -1 {
		return errors.New("os: process already released")
	}
	// kill with a pid of 0 or less signals a process group, see kill(2).
	if p.Pid <= 0 {
		return errors.New("os: process not initialized")
	}

	p.sigMu.RLock()
	defer p.sigMu.RUnlock()
	if p.done.Load() {
		return ErrProcessDone
	}
	if err := syscall.Kill(p.Pid, s); err != nil {
		if err == syscall.ESRCH {
			return ErrProcessDone
		}
		return err
	}
	return nil
}

// forkExec starts argv0 with posix_spawn.
func forkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	if argv == nil {
		return 0, errors.New("exec: no argv")
	}

	if len(argv) == 0 {
		return 0, errors.New("exec: no argv")
	}

	return spawn(argv0, argv, attr)
}

func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	if attr == nil {
		return nil, errors.New("exec: no ProcAttr")
	}

	if attr.Sys != nil {
		return startProcessWithSys(name, argv, attr)
	}

	// Check Dir first to report a clear error, as upstream os/exec_posix.go does.
	if attr.Dir != "" {
		if _, err := Stat(attr.Dir); err != nil {
			pe := err.(*PathError)
			pe.Op = "chdir"
			return nil, pe
		}
	}

	pid, err := forkExec(name, argv, attr)
	if err != nil {
		return nil, err
	}

	return findProcess(pid)
}

func startProcessWithSys(name string, argv []string, attr *ProcAttr) (*Process, error) {
	sysattr := &syscall.ProcAttr{
		Dir:   attr.Dir,
		Env:   attr.Env,
		Files: make([]uintptr, 0, len(attr.Files)),
		Sys:   attr.Sys,
	}
	if sysattr.Env == nil {
		sysattr.Env = Environ()
	}
	for _, f := range attr.Files {
		if f == nil {
			sysattr.Files = append(sysattr.Files, ^uintptr(0))
		} else {
			sysattr.Files = append(sysattr.Files, f.Fd())
		}
	}

	pid, _, err := syscall.StartProcess(name, argv, sysattr)
	runtime.KeepAlive(attr)
	if err != nil {
		return nil, &PathError{Op: "fork/exec", Path: name, Err: err}
	}

	return findProcess(pid)
}
