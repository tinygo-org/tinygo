//go:build baremetal || js || wasm_freestanding
// +build baremetal js wasm_freestanding

package syscall

import (
	"internal/itoa"
)

// Most code here has been copied from the Go sources:
//   https://github.com/golang/go/blob/go1.12/src/syscall/syscall_js.go
// It has the following copyright note:
//
//     Copyright 2018 The Go Authors. All rights reserved.
//     Use of this source code is governed by a BSD-style
//     license that can be found in the LICENSE file.

// A Signal is a number describing a process signal.
// It implements the os.Signal interface.
type Signal int

const (
	_ Signal = iota
	SIGCHLD
	SIGINT
	SIGKILL
	SIGTRAP
	SIGQUIT
	SIGTERM
	SIGILL
	SIGABRT
	SIGBUS
	SIGFPE
	SIGSEGV
	SIGPIPE
)

func (s Signal) Signal() {}

func (s Signal) String() string {
	if 0 <= s && int(s) < len(signals) {
		str := signals[s]
		if str != "" {
			return str
		}
	}
	return "signal " + itoa.Itoa(int(s))
}

var signals = [...]string{}

// File system

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	O_RDONLY = 0
	O_WRONLY = 1
	O_RDWR   = 2

	O_CREAT  = 0100
	O_CREATE = O_CREAT
	O_TRUNC  = 01000
	O_APPEND = 02000
	O_EXCL   = 0200
	O_SYNC   = 010000

	O_CLOEXEC = 0
)

// Dummy values to allow compiling tests
// Dummy source: https://opensource.apple.com/source/xnu/xnu-7195.81.3/bsd/sys/mman.h.auto.html
const (
	PROT_NONE  = 0x00 // no permissions
	PROT_READ  = 0x01 // pages can be read
	PROT_WRITE = 0x02 // pages can be written
	PROT_EXEC  = 0x04 // pages can be executed

	MAP_SHARED  = 0x0001 // share changes
	MAP_PRIVATE = 0x0002 // changes are private

	MAP_FILE      = 0x0000 // map from file (default)
	MAP_ANON      = 0x1000 // allocated from memory, swap space
	MAP_ANONYMOUS = MAP_ANON
)

func runtime_envs() []string

func Getenv(key string) (value string, found bool) {
	env := runtime_envs()
	for _, keyval := range env {
		// Split at '=' character.
		var k, v string
		for i := 0; i < len(keyval); i++ {
			if keyval[i] == '=' {
				k = keyval[:i]
				v = keyval[i+1:]
			}
		}
		if k == key {
			return v, true
		}
	}
	return "", false
}

func Setenv(key, val string) (err error) {
	// stub for now
	return ENOSYS
}

func Unsetenv(key string) (err error) {
	// stub for now
	return ENOSYS
}

func Clearenv() (err error) {
	// stub for now
	return ENOSYS
}

func Environ() []string {
	env := runtime_envs()
	envCopy := make([]string, len(env))
	copy(envCopy, env)
	return envCopy
}

func Open(path string, mode int, perm uint32) (fd int, err error) {
	return 0, ENOSYS
}

func Read(fd int, p []byte) (n int, err error) {
	return 0, ENOSYS
}

func Seek(fd int, offset int64, whence int) (off int64, err error) {
	return 0, ENOSYS
}

func Close(fd int) (err error) {
	return ENOSYS
}

// Processes

type WaitStatus uint32

func (w WaitStatus) Exited() bool       { return false }
func (w WaitStatus) ExitStatus() int    { return 0 }
func (w WaitStatus) Signaled() bool     { return false }
func (w WaitStatus) Signal() Signal     { return 0 }
func (w WaitStatus) CoreDump() bool     { return false }
func (w WaitStatus) Stopped() bool      { return false }
func (w WaitStatus) Continued() bool    { return false }
func (w WaitStatus) StopSignal() Signal { return 0 }
func (w WaitStatus) TrapCause() int     { return 0 }

// XXX made up
type Rusage struct {
	Utime Timeval
	Stime Timeval
}

// XXX made up
type ProcAttr struct {
	Dir   string
	Env   []string
	Files []uintptr
	Sys   *SysProcAttr
}

type SysProcAttr struct {
}

func Getgroups() ([]int, error)         { return []int{1}, nil }
func Gettimeofday(tv *Timeval) error    { return ENOSYS }
func Kill(pid int, signum Signal) error { return ENOSYS }
func Pipe2(p []int, flags int) (err error) {
	return ENOSYS // TODO
}
func Sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	return 0, ENOSYS
}
func StartProcess(argv0 string, argv []string, attr *ProcAttr) (pid int, handle uintptr, err error) {
	return 0, 0, ENOSYS
}
func Wait4(pid int, wstatus *WaitStatus, options int, rusage *Rusage) (wpid int, err error) {
	return 0, ENOSYS
}

func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return nil, ENOSYS
}

func Munmap(b []byte) (err error) {
	return ENOSYS
}

type Timeval struct {
	Sec  int64
	Usec int64
}

func Getpagesize() int {
	// There is no right value to return here, but 4096 is a
	// common assumption when pagesize is unknown
	return 4096
}
