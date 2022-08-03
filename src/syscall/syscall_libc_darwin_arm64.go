//go:build darwin
// +build darwin

package syscall

import (
	"unsafe"
)

// struct DIR * buf fdopendir(int fd);
//
//export fdopendir
func libc_fdopendir(fd int32) unsafe.Pointer

// int readdir_r(struct DIR * buf, struct dirent *entry, struct dirent **result);
//
//export readdir_r
func libc_readdir_r(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) int32

// int stat(const char *path, struct stat * buf);
//
//export stat
func libc_stat(pathname *byte, ptr unsafe.Pointer) int32

// int fstat(int fd, struct stat * buf);
//
//export fstat
func libc_fstat(fd int32, ptr unsafe.Pointer) int32

// int lstat(const char *path, struct stat * buf);
//
//export lstat
func libc_lstat(pathname *byte, ptr unsafe.Pointer) int32
