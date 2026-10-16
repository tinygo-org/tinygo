// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"errors"
	"internal/poll"
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

type pollFD struct {
	fd *poll.FD
}

func (p pollFD) Close() error {
	return p.fd.Close()
}

func (p pollFD) Exist() bool {
	return p.fd != nil
}

func (f *File) ensurePollFD() (*poll.FD, error) {
	if f.handle == nil {
		return nil, ErrClosed
	}
	if f.pfd.fd != nil {
		return f.pfd.fd, nil
	}
	h, ok := f.handle.(fd)
	if !ok {
		return nil, ErrNotImplemented
	}
	pfd := &poll.FD{
		Sysfd:         int(h.Fd()),
		IsStream:      true,
		ZeroReadIsEOF: true,
	}
	_ = pfd.Init("file", true)
	f.pfd.fd = pfd
	return pfd, nil
}

func (f *File) read(b []byte) (int, error) {
	if f.pfd.fd != nil {
		return f.pfd.fd.Read(b)
	}
	n, err := f.handle.Read(b)
	if err == syscall.EAGAIN {
		if pfd, perr := f.ensurePollFD(); perr == nil {
			return pfd.Read(b)
		}
	}
	return n, err
}

func (f *File) write(b []byte) (int, error) {
	if f.pfd.fd != nil {
		return f.pfd.fd.Write(b)
	}
	n, err := f.handle.Write(b)
	if err == syscall.EAGAIN {
		if pfd, perr := f.ensurePollFD(); perr == nil {
			return pfd.Write(b)
		}
	}
	return n, err
}

type rawConn struct {
	file *File
}

func (c *rawConn) Control(fn func(uintptr)) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawControl(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Read(fn func(uintptr) bool) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawRead(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Write(fn func(uintptr) bool) error {
	pfd, err := c.file.ensurePollFD()
	if err != nil {
		return err
	}
	err = pfd.RawWrite(fn)
	runtime.KeepAlive(c.file)
	return err
}

func (f *File) syscallConn() (syscall.RawConn, error) {
	if _, err := f.ensurePollFD(); err != nil {
		return nil, err
	}
	return &rawConn{file: f}, nil
}

func (f *File) setDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetDeadline(t)
}

func (f *File) setReadDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetReadDeadline(t)
}

func (f *File) setWriteDeadline(t time.Time) error {
	pfd, err := f.ensurePollFD()
	if err != nil {
		return err
	}
	return pfd.SetWriteDeadline(t)
}

const _UTIME_OMIT = (1 << 30) - 2

func Chtimes(name string, atime time.Time, mtime time.Time) error {
	var utimes [2]syscall.Timespec
	set := func(i int, t time.Time) {
		if t.IsZero() {
			utimes[i] = syscall.Timespec{Sec: _UTIME_OMIT, Nsec: _UTIME_OMIT}
		} else {
			utimes[i] = syscall.NsecToTimespec(t.UnixNano())
		}
	}
	set(0, atime)
	set(1, mtime)
	if e := syscall.UtimesNano(fixLongPath(name), utimes[0:]); e != nil {
		return &PathError{Op: "chtimes", Path: name, Err: e}
	}
	return nil
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
