//go:build !baremetal && !wasi && !wasm
// +build !baremetal,!wasi,!wasm

// This file assumes there is a libc available that runs on a real operating
// system.

package syscall

func Getuid() int  { return int(libc_getuid()) }
func Geteuid() int { return int(libc_geteuid()) }
func Getgid() int  { return int(libc_getgid()) }
func Getegid() int { return int(libc_getegid()) }
func Getpid() int  { return int(libc_getpid()) }
func Getppid() int { return int(libc_getppid()) }

// uid_t getuid(void)
//
//export getuid
func libc_getuid() int32

// gid_t getgid(void)
//
//export getgid
func libc_getgid() int32

// uid_t geteuid(void)
//
//export geteuid
func libc_geteuid() int32

// gid_t getegid(void)
//
//export getegid
func libc_getegid() int32

// gid_t getpid(void)
//
//export getpid
func libc_getpid() int32

// gid_t getppid(void)
//
//export getppid
func libc_getppid() int32
