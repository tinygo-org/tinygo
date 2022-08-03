//go:build darwin
// +build darwin

package syscall

import (
	"unsafe"
)

// The odd $INODE64 suffix is an Apple compatibility feature, see
// __DARWIN_INODE64 in /Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/usr/include/sys/cdefs.h
// and https://assert.cc/posts/darwin_use_64_bit_inode_vs_ctypes/
// Without it, you get the old, smaller struct stat from mac os 10.2 or so.
// It not needed on arm64.

// struct DIR * buf fdopendir(int fd);
//
//export fdopendir$INODE64
func libc_fdopendir(fd int32) unsafe.Pointer

// int readdir_r(struct DIR * buf, struct dirent *entry, struct dirent **result);
//
//export readdir_r$INODE64
func libc_readdir_r(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) int32

// int stat(const char *path, struct stat * buf);
//
//export stat$INODE64
func libc_stat(pathname *byte, ptr unsafe.Pointer) int32

// int fstat(int fd, struct stat * buf);
//
//export fstat$INODE64
func libc_fstat(fd int32, ptr unsafe.Pointer) int32

// int lstat(const char *path, struct stat * buf);
//
//export lstat$INODE64
func libc_lstat(pathname *byte, ptr unsafe.Pointer) int32
