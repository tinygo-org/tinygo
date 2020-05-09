// +build baremetal

package syscall

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
)

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

func Getenv(key string) (value string, found bool) {
	return "", false // stub
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

func Getegid() int                      { return 1 }
func Geteuid() int                      { return 1 }
func Getgid() int                       { return 1 }
func Getgroups() ([]int, error)         { return []int{1}, nil }
func Getppid() int                      { return 2 }
func Getpid() int                       { return 3 }
func Gettimeofday(tv *Timeval) error    { return ENOSYS }
func Getuid() int                       { return 1 }
func Kill(pid int, signum Signal) error { return ENOSYS }
func Sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	return 0, ENOSYS
}
func StartProcess(argv0 string, argv []string, attr *ProcAttr) (pid int, handle uintptr, err error) {
	return 0, 0, ENOSYS
}
func Wait4(pid int, wstatus *WaitStatus, options int, rusage *Rusage) (wpid int, err error) {
	return 0, ENOSYS
}

type Timeval struct {
	Sec  int64
	Usec int64
}
