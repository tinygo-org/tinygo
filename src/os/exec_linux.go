// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"errors"
	"runtime"
	"syscall"
	"time"
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

func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	if attr == nil {
		attr = &ProcAttr{}
	}
	if attr.Sys == nil && attr.Dir != "" {
		if _, err := Stat(attr.Dir); err != nil {
			pe := err.(*PathError)
			pe.Op = "chdir"
			return nil, pe
		}
	}

	sysattr := &syscall.ProcAttr{
		Dir: attr.Dir,
		Env: attr.Env,
		Sys: attr.Sys,
	}
	if sysattr.Env == nil {
		sysattr.Env = Environ()
	}
	sysattr.Files = make([]uintptr, 0, len(attr.Files))
	for _, f := range attr.Files {
		if f == nil {
			sysattr.Files = append(sysattr.Files, ^uintptr(0))
		} else {
			sysattr.Files = append(sysattr.Files, f.Fd())
		}
	}

	pid, _, e := syscall.StartProcess(name, argv, sysattr)
	runtime.KeepAlive(attr)
	if e != nil {
		return nil, &PathError{Op: "fork/exec", Path: name, Err: e}
	}

	return findProcess(pid)
}

func (p *Process) wait() (*ProcessState, error) {
	if p.Pid == -1 {
		return nil, syscall.EINVAL
	}
	var (
		status syscall.WaitStatus
		rusage syscall.Rusage
		pid1   int
		err    error
	)
	for {
		pid1, err = syscall.Wait4(p.Pid, &status, 0, &rusage)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		return nil, NewSyscallError("wait", err)
	}
	p.release()
	return &ProcessState{
		pid:    pid1,
		status: status,
		rusage: &rusage,
	}, nil
}

func (p *Process) signal(sig Signal) error {
	if p.Pid == -1 {
		return errors.New("os: process already released")
	}
	if p.Pid == 0 {
		return errors.New("os: process not initialized")
	}
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	if err := syscall.Kill(p.Pid, s); err != nil {
		if err == syscall.ESRCH {
			return ErrProcessDone
		}
		return err
	}
	return nil
}

func (p *ProcessState) userTime() time.Duration {
	if p.rusage == nil {
		return 0
	}
	return time.Duration(p.rusage.Utime.Nano()) * time.Nanosecond
}

func (p *ProcessState) systemTime() time.Duration {
	if p.rusage == nil {
		return 0
	}
	return time.Duration(p.rusage.Stime.Nano()) * time.Nanosecond
}

func hostname() (name string, err error) {
	var un syscall.Utsname
	err = syscall.Uname(&un)

	var buf [512]byte
	for i, b := range un.Nodename[:] {
		buf[i] = uint8(b)
		if b == 0 {
			name = string(buf[:i])
			break
		}
	}
	if err == nil && len(name) > 0 && len(name) < 64 {
		return name, nil
	}

	f, err := Open("/proc/sys/kernel/hostname")
	if err != nil {
		return "", err
	}
	defer f.Close()

	n, err := f.Read(buf[:])
	if err != nil {
		return "", err
	}

	if n > 0 && buf[n-1] == '\n' {
		n--
	}
	return string(buf[:n]), nil
}
