//go:build wasip2

// mini libc wrapping wasi preview2 calls in a libc api

package syscall

import (
	"unsafe"

	"internal/cm"

	"internal/wasi/cli/v0.2.0/environment"
	"internal/wasi/cli/v0.2.0/stderr"
	"internal/wasi/cli/v0.2.0/stdin"
	"internal/wasi/cli/v0.2.0/stdout"
	wallclock "internal/wasi/clocks/v0.2.0/wall-clock"
	"internal/wasi/filesystem/v0.2.0/preopens"
	"internal/wasi/filesystem/v0.2.0/types"
	ioerror "internal/wasi/io/v0.2.0/error"
	"internal/wasi/io/v0.2.0/streams"
	"internal/wasi/random/v0.2.0/random"
)

func goString(cstr *byte) string {
	return unsafe.String(cstr, strlen(cstr))
}

//export strlen
func strlen(cstr *byte) uintptr {
	if cstr == nil {
		return 0
	}
	ptr := unsafe.Pointer(cstr)
	var i uintptr
	for p := (*byte)(ptr); *p != 0; p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1)) {
		i++
	}
	return i
}

// ssize_t write(int fd, const void *buf, size_t count)
//
//export write
func write(fd int32, buf *byte, count uint) int {
	if stream, ok := wasiStreams[fd]; ok {
		return writeStream(stream, buf, count, 0)
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if stream.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}

	n := pwrite(fd, buf, count, int64(stream.offset))
	if n == -1 {
		return -1
	}
	stream.offset += int64(n)
	return int(n)
}

// ssize_t read(int fd, void *buf, size_t count);
//
//export read
func read(fd int32, buf *byte, count uint) int {
	if stream, ok := wasiStreams[fd]; ok {
		return readStream(stream, buf, count, 0)
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if stream.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}

	n := pread(fd, buf, count, int64(stream.offset))
	if n == -1 {
		// error during pread
		return -1
	}
	stream.offset += int64(n)
	return int(n)
}

// At the moment, each time we have a file read or write we create a new stream.  Future implementations
// could change the current in or out file stream lazily.  We could do this by tracking input and output
// offsets individually, and if they don't match the current main offset, reopen the file stream at that location.

type wasiFile struct {
	d      types.Descriptor
	oflag  int32 // original open flags: O_RDONLY, O_WRONLY, O_RDWR
	offset int64 // current fd offset; updated with each read/write
	refs   int
}

// Need to figure out which system calls we're using:
//   stdin/stdout/stderr want streams, so we use stream read/write
//   but for regular files we can use the descriptor and explicitly write a buffer to the offset?
//   The mismatch comes from trying to combine these.

var wasiFiles map[int32]*wasiFile = make(map[int32]*wasiFile)

func findFreeFD() int32 {
	var newfd int32
	for wasiStreams[newfd] != nil || wasiFiles[newfd] != nil {
		newfd++
	}
	return newfd
}

var wasiErrno ioerror.Error

type wasiStream struct {
	in   *streams.InputStream
	out  *streams.OutputStream
	refs int
}

// This holds entries for stdin/stdout/stderr.

var wasiStreams map[int32]*wasiStream

func init() {
	sin := stdin.GetStdin()
	sout := stdout.GetStdout()
	serr := stderr.GetStderr()
	wasiStreams = map[int32]*wasiStream{
		0: &wasiStream{
			in:   &sin,
			refs: 1,
		},
		1: &wasiStream{
			out:  &sout,
			refs: 1,
		},
		2: &wasiStream{
			out:  &serr,
			refs: 1,
		},
	}
}

func readStream(stream *wasiStream, buf *byte, count uint, offset int64) int {
	if stream.in == nil {
		// not a stream we can read from
		libcErrno = EBADF
		return -1
	}

	if offset != 0 {
		libcErrno = EINVAL
		return -1
	}

	libcErrno = 0
	result := stream.in.BlockingRead(uint64(count))
	if err := result.Err(); err != nil {
		if err.Closed() {
			libcErrno = 0
			return 0
		} else if err := err.LastOperationFailed(); err != nil {
			wasiErrno = *err
			libcErrno = EWASIERROR
		}
		return -1
	}

	dst := unsafe.Slice(buf, count)
	list := result.OK()
	copy(dst, list.Slice())
	return int(list.Len())
}

func writeStream(stream *wasiStream, buf *byte, count uint, offset int64) int {
	if stream.out == nil {
		// not a stream we can write to
		libcErrno = EBADF
		return -1
	}

	if offset != 0 {
		libcErrno = EINVAL
		return -1
	}

	src := unsafe.Slice(buf, count)
	var remaining = count

	// The blocking-write-and-flush call allows a maximum of 4096 bytes at a time.
	// We loop here by instead of doing subscribe/check-write/poll-one/write by hand.
	for remaining > 0 {
		len := uint(4096)
		if len > remaining {
			len = remaining
		}
		result := stream.out.BlockingWriteAndFlush(cm.ToList(src[:len]))
		if err := result.Err(); err != nil {
			if err.Closed() {
				libcErrno = 0
				return 0
			} else if err := err.LastOperationFailed(); err != nil {
				wasiErrno = *err
				libcErrno = EWASIERROR
			}
			return -1
		}
		remaining -= len
	}

	return int(count)
}

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)

// ssize_t pread(int fd, void *buf, size_t count, off_t offset);
//
//export pread
func pread(fd int32, buf *byte, count uint, offset int64) int {
	// TODO(dgryski): Need to be consistent about all these checks; EBADF/EINVAL/... ?

	if stream, ok := wasiStreams[fd]; ok {
		return readStream(stream, buf, count, offset)

	}

	streams, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = EBADF
		return -1
	}
	if streams.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}
	if streams.oflag&O_RDONLY == 0 {
		libcErrno = EBADF
		return -1
	}

	result := streams.d.Read(types.FileSize(count), types.FileSize(offset))
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	list := result.OK().F0
	copy(unsafe.Slice(buf, count), list.Slice())

	// TODO(dgryski): EOF bool is ignored?
	return int(list.Len())
}

// ssize_t pwrite(int fd, void *buf, size_t count, off_t offset);
//
//export pwrite
func pwrite(fd int32, buf *byte, count uint, offset int64) int {
	// TODO(dgryski): Need to be consistent about all these checks; EBADF/EINVAL/... ?
	if stream, ok := wasiStreams[fd]; ok {
		return writeStream(stream, buf, count, 0)
	}

	streams, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = EBADF
		return -1
	}
	if streams.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}
	if streams.oflag&O_WRONLY == 0 {
		libcErrno = EBADF
		return -1
	}

	result := streams.d.Write(cm.NewList(buf, count), types.FileSize(offset))
	if err := result.Err(); err != nil {
		// TODO(dgryski):
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return int(*result.OK())
}

// ssize_t lseek(int fd, off_t offset, int whence);
//
//export lseek
func lseek(fd int32, offset int64, whence int) int64 {
	if _, ok := wasiStreams[fd]; ok {
		// can't lseek a stream
		libcErrno = EBADF
		return -1
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if stream.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}

	switch whence {
	case 0: // SEEK_SET
		stream.offset = offset
	case 1: // SEEK_CUR
		stream.offset += offset
	case 2: // SEEK_END
		result := stream.d.Stat()
		if err := result.Err(); err != nil {
			libcErrno = errorCodeToErrno(*err)
			return -1
		}
		stream.offset = int64(result.OK().Size) + offset
	}

	return int64(stream.offset)
}

// int close(int fd)
//
//export close
func close(fd int32) int32 {
	if streams, ok := wasiStreams[fd]; ok {
		if streams.out != nil {
			// ignore any error
			streams.out.BlockingFlush()
		}

		if streams.refs--; streams.refs == 0 {
			if streams.out != nil {
				streams.out.ResourceDrop()
				streams.out = nil
			}
			if streams.in != nil {
				streams.in.ResourceDrop()
				streams.in = nil
			}
		}

		delete(wasiStreams, fd)
		return 0
	}

	streams, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if streams.refs--; streams.refs == 0 && streams.d != cm.ResourceNone {
		streams.d.ResourceDrop()
		streams.d = 0
	}
	delete(wasiFiles, fd)

	return 0
}

// int dup(int fd)
//
//export dup
func dup(fd int32) int32 {
	// is fd a stream?
	if stream, ok := wasiStreams[fd]; ok {
		newfd := findFreeFD()
		stream.refs++
		wasiStreams[newfd] = stream
		return newfd
	}

	// is fd a file?
	if file, ok := wasiFiles[fd]; ok {
		// scan for first free file descriptor
		newfd := findFreeFD()
		file.refs++
		wasiFiles[newfd] = file
		return newfd
	}

	// unknown file descriptor
	libcErrno = EBADF
	return -1
}

// void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset);
//
//export mmap
func mmap(addr unsafe.Pointer, length uintptr, prot, flags, fd int32, offset uintptr) unsafe.Pointer {
	libcErrno = ENOSYS
	return unsafe.Pointer(^uintptr(0))
}

// int munmap(void *addr, size_t length);
//
//export munmap
func munmap(addr unsafe.Pointer, length uintptr) int32 {
	libcErrno = ENOSYS
	return -1
}

// int mprotect(void *addr, size_t len, int prot);
//
//export mprotect
func mprotect(addr unsafe.Pointer, len uintptr, prot int32) int32 {
	libcErrno = ENOSYS
	return -1
}

// int chmod(const char *pathname, mode_t mode);
//
//export chmod
func chmod(pathname *byte, mode uint32) int32 {
	return 0
}

// int mkdir(const char *pathname, mode_t mode);
//
//export mkdir
func mkdir(pathname *byte, mode uint32) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.CreateDirectoryAt(relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int rmdir(const char *pathname);
//
//export rmdir
func rmdir(pathname *byte) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.RemoveDirectoryAt(relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int rename(const char *from, *to);
//
//export rename
func rename(from, to *byte) int32 {
	fromPath := goString(from)
	fromDir, fromRelPath := findPreopenForPath(fromPath)
	if fromDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	toPath := goString(to)
	toDir, toRelPath := findPreopenForPath(toPath)
	if toDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := fromDir.d.RenameAt(fromRelPath, toDir.d, toRelPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int symlink(const char *from, *to);
//
//export symlink
func symlink(from, to *byte) int32 {
	fromPath := goString(from)
	fromDir, fromRelPath := findPreopenForPath(fromPath)
	if fromDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	toPath := goString(to)
	toDir, toRelPath := findPreopenForPath(toPath)
	if toDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	if fromDir.d != toDir.d {
		libcErrno = EACCES
		return -1
	}

	// TODO(dgryski): check fromDir == toDir?

	result := fromDir.d.SymlinkAt(fromRelPath, toRelPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int link(const char *from, *to);
//
//export link
func link(from, to *byte) int32 {
	fromPath := goString(from)
	fromDir, fromRelPath := findPreopenForPath(fromPath)
	if fromDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	toPath := goString(to)
	toDir, toRelPath := findPreopenForPath(toPath)
	if toDir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	if fromDir.d != toDir.d {
		libcErrno = EACCES
		return -1
	}

	// TODO(dgryski): check fromDir == toDir?

	result := fromDir.d.LinkAt(0, fromRelPath, toDir.d, toRelPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int fsync(int fd);
//
//export fsync
func fsync(fd int32) int32 {
	if _, ok := wasiStreams[fd]; ok {
		// can't sync a stream
		libcErrno = EBADF
		return -1
	}

	streams, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if streams.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}
	if streams.oflag&O_WRONLY == 0 {
		libcErrno = EBADF
		return -1
	}

	result := streams.d.SyncData()
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// ssize_t readlink(const char *path, void *buf, size_t count);
//
//export readlink
func readlink(pathname *byte, buf *byte, count uint) int {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.ReadLinkAt(relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	s := *result.OK()
	size := uintptr(count)
	if size > uintptr(len(s)) {
		size = uintptr(len(s))
	}

	memcpy(unsafe.Pointer(buf), unsafe.Pointer(unsafe.StringData(s)), size)
	return int(size)
}

// int unlink(const char *pathname);
//
//export unlink
func unlink(pathname *byte) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.UnlinkFileAt(relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	return 0
}

// int getpagesize(void);
//
//export getpagesize
func getpagesize() int {
	return 65536
}

// int stat(const char *path, struct stat * buf);
//
//export stat
func stat(pathname *byte, dst *Stat_t) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.StatAt(0, relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	setStatFromWASIStat(dst, result.OK())

	return 0
}

// int fstat(int fd, struct stat * buf);
//
//export fstat
func fstat(fd int32, dst *Stat_t) int32 {
	if _, ok := wasiStreams[fd]; ok {
		// TODO(dgryski): fill in stat buffer for stdin etc
		return -1
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return -1
	}
	if stream.d == cm.ResourceNone {
		libcErrno = EBADF
		return -1
	}
	result := stream.d.Stat()
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	setStatFromWASIStat(dst, result.OK())

	return 0
}

func setStatFromWASIStat(sstat *Stat_t, wstat *types.DescriptorStat) {
	// This will cause problems for people who want to compare inodes
	sstat.Dev = 0
	sstat.Ino = 0
	sstat.Rdev = 0

	sstat.Nlink = uint64(wstat.LinkCount)

	sstat.Mode = p2fileTypeToStatType(wstat.Type)

	// No uid/gid
	sstat.Uid = 0
	sstat.Gid = 0
	sstat.Size = int64(wstat.Size)

	// made up numbers
	sstat.Blksize = 512
	sstat.Blocks = (sstat.Size + 511) / int64(sstat.Blksize)

	setOptTime := func(t *Timespec, o *wallclock.DateTime) {
		t.Sec = 0
		t.Nsec = 0
		if o != nil {
			t.Sec = int32(o.Seconds)
			t.Nsec = int64(o.Nanoseconds)
		}
	}

	setOptTime(&sstat.Atim, wstat.DataAccessTimestamp.Some())
	setOptTime(&sstat.Mtim, wstat.DataModificationTimestamp.Some())
	setOptTime(&sstat.Ctim, wstat.StatusChangeTimestamp.Some())
}

// int lstat(const char *path, struct stat * buf);
//
//export lstat
func lstat(pathname *byte, dst *Stat_t) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.StatAt(0, relPath)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	setStatFromWASIStat(dst, result.OK())

	return 0
}

func init() {
	populateEnvironment()
	populatePreopens()
}

type wasiDir struct {
	d    types.Descriptor // wasip2 descriptor
	root string           // root path for this descriptor
	rel  string           // relative path under root
}

var libcCWD wasiDir

var wasiPreopens map[string]types.Descriptor

func populatePreopens() {
	var cwd string

	// find CWD
	result := environment.InitialCWD()
	if s := result.Some(); s != nil {
		cwd = *s
	} else if s, _ := Getenv("PWD"); s != "" {
		cwd = s
	}

	dirs := preopens.GetDirectories().Slice()
	preopens := make(map[string]types.Descriptor, len(dirs))
	for _, tup := range dirs {
		desc, path := tup.F0, tup.F1
		if path == cwd {
			libcCWD.d = desc
			libcCWD.root = path
			libcCWD.rel = ""
		}
		preopens[path] = desc
	}
	wasiPreopens = preopens
}

// -- BEGIN fs_wasip1.go --
// The following section has been taken from upstream Go with the following copyright:
// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:nosplit
func appendCleanPath(buf []byte, path string, lookupParent bool) ([]byte, bool) {
	i := 0
	for i < len(path) {
		for i < len(path) && path[i] == '/' {
			i++
		}

		j := i
		for j < len(path) && path[j] != '/' {
			j++
		}

		s := path[i:j]
		i = j

		switch s {
		case "":
			continue
		case ".":
			continue
		case "..":
			if !lookupParent {
				k := len(buf)
				for k > 0 && buf[k-1] != '/' {
					k--
				}
				for k > 1 && buf[k-1] == '/' {
					k--
				}
				buf = buf[:k]
				if k == 0 {
					lookupParent = true
				} else {
					s = ""
					continue
				}
			}
		default:
			lookupParent = false
		}

		if len(buf) > 0 && buf[len(buf)-1] != '/' {
			buf = append(buf, '/')
		}
		buf = append(buf, s...)
	}
	return buf, lookupParent
}

// joinPath concatenates dir and file paths, producing a cleaned path where
// "." and ".." have been removed, unless dir is relative and the references
// to parent directories in file represented a location relative to a parent
// of dir.
//
// This function is used for path resolution of all wasi functions expecting
// a path argument; the returned string is heap allocated, which we may want
// to optimize in the future. Instead of returning a string, the function
// could append the result to an output buffer that the functions in this
// file can manage to have allocated on the stack (e.g. initializing to a
// fixed capacity). Since it will significantly increase code complexity,
// we prefer to optimize for readability and maintainability at this time.
func joinPath(dir, file string) string {
	buf := make([]byte, 0, len(dir)+len(file)+1)
	if isAbs(dir) {
		buf = append(buf, '/')
	}

	buf, lookupParent := appendCleanPath(buf, dir, true)
	buf, _ = appendCleanPath(buf, file, lookupParent)
	// The appendCleanPath function cleans the path so it does not inject
	// references to the current directory. If both the dir and file args
	// were ".", this results in the output buffer being empty so we handle
	// this condition here.
	if len(buf) == 0 {
		buf = append(buf, '.')
	}
	// If the file ended with a '/' we make sure that the output also ends
	// with a '/'. This is needed to ensure that programs have a mechanism
	// to represent dereferencing symbolic links pointing to directories.
	if buf[len(buf)-1] != '/' && isDir(file) {
		buf = append(buf, '/')
	}
	return unsafe.String(&buf[0], len(buf))
}

func isAbs(path string) bool {
	return hasPrefix(path, "/")
}

func isDir(path string) bool {
	return hasSuffix(path, "/")
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func hasSuffix(s, x string) bool {
	return len(s) >= len(x) && s[len(s)-len(x):] == x
}

// findPreopenForPath finds which preopen it relates to and return that descriptor/root and the path relative to that directory descriptor/root
func findPreopenForPath(path string) (wasiDir, string) {
	dir := "/"
	var wasidir wasiDir

	if !isAbs(path) {
		dir = libcCWD.root
		wasidir = libcCWD
		if libcCWD.rel != "" && libcCWD.rel != "." && libcCWD.rel != "./" {
			path = libcCWD.rel + "/" + path
		}
	}
	path = joinPath(dir, path)

	var best string
	for k, v := range wasiPreopens {
		if len(k) > len(best) && hasPrefix(path, k) {
			wasidir = wasiDir{d: v, root: k}
			best = wasidir.root
		}
	}

	if hasPrefix(path, wasidir.root) {
		path = path[len(wasidir.root):]
	}
	for isAbs(path) {
		path = path[1:]
	}
	if len(path) == 0 {
		path = "."
	}

	return wasidir, path
}

// -- END fs_wasip1.go --

// int open(const char *pathname, int flags, mode_t mode);
//
//export open
func open(pathname *byte, flags int32, mode uint32) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	var dflags types.DescriptorFlags
	if (flags & O_RDONLY) == O_RDONLY {
		dflags |= types.DescriptorFlagsRead
	}
	if (flags & O_WRONLY) == O_WRONLY {
		dflags |= types.DescriptorFlagsWrite
	}

	var oflags types.OpenFlags
	if flags&O_CREAT == O_CREAT {
		oflags |= types.OpenFlagsCreate
	}
	if flags&O_DIRECTORY == O_DIRECTORY {
		oflags |= types.OpenFlagsDirectory
	}
	if flags&O_EXCL == O_EXCL {
		oflags |= types.OpenFlagsExclusive
	}
	if flags&O_TRUNC == O_TRUNC {
		oflags |= types.OpenFlagsTruncate
	}

	// By default, follow symlinks for open() unless O_NOFOLLOW was passed
	var pflags types.PathFlags = types.PathFlagsSymlinkFollow
	if flags&O_NOFOLLOW == O_NOFOLLOW {
		// O_NOFOLLOW was passed, so turn off SymlinkFollow
		pflags &^= types.PathFlagsSymlinkFollow
	}

	result := dir.d.OpenAt(pflags, relPath, oflags, dflags)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	stream := wasiFile{
		d:     *result.OK(),
		oflag: flags,
		refs:  1,
	}

	if flags&(O_WRONLY|O_APPEND) == (O_WRONLY | O_APPEND) {
		result := stream.d.Stat()
		if err := result.Err(); err != nil {
			libcErrno = errorCodeToErrno(*err)
			return -1
		}
		stream.offset = int64(result.OK().Size)
	}

	libcfd := findFreeFD()

	wasiFiles[libcfd] = &stream

	return int32(libcfd)
}

func errorCodeToErrno(err types.ErrorCode) Errno {
	switch err {
	case types.ErrorCodeAccess:
		return EACCES
	case types.ErrorCodeWouldBlock:
		return EAGAIN
	case types.ErrorCodeAlready:
		return EALREADY
	case types.ErrorCodeBadDescriptor:
		return EBADF
	case types.ErrorCodeBusy:
		return EBUSY
	case types.ErrorCodeDeadlock:
		return EDEADLK
	case types.ErrorCodeQuota:
		return EDQUOT
	case types.ErrorCodeExist:
		return EEXIST
	case types.ErrorCodeFileTooLarge:
		return EFBIG
	case types.ErrorCodeIllegalByteSequence:
		return EILSEQ
	case types.ErrorCodeInProgress:
		return EINPROGRESS
	case types.ErrorCodeInterrupted:
		return EINTR
	case types.ErrorCodeInvalid:
		return EINVAL
	case types.ErrorCodeIO:
		return EIO
	case types.ErrorCodeIsDirectory:
		return EISDIR
	case types.ErrorCodeLoop:
		return ELOOP
	case types.ErrorCodeTooManyLinks:
		return EMLINK
	case types.ErrorCodeMessageSize:
		return EMSGSIZE
	case types.ErrorCodeNameTooLong:
		return ENAMETOOLONG
	case types.ErrorCodeNoDevice:
		return ENODEV
	case types.ErrorCodeNoEntry:
		return ENOENT
	case types.ErrorCodeNoLock:
		return ENOLCK
	case types.ErrorCodeInsufficientMemory:
		return ENOMEM
	case types.ErrorCodeInsufficientSpace:
		return ENOSPC
	case types.ErrorCodeNotDirectory:
		return ENOTDIR
	case types.ErrorCodeNotEmpty:
		return ENOTEMPTY
	case types.ErrorCodeNotRecoverable:
		return ENOTRECOVERABLE
	case types.ErrorCodeUnsupported:
		return ENOSYS
	case types.ErrorCodeNoTTY:
		return ENOTTY
	case types.ErrorCodeNoSuchDevice:
		return ENXIO
	case types.ErrorCodeOverflow:
		return EOVERFLOW
	case types.ErrorCodeNotPermitted:
		return EPERM
	case types.ErrorCodePipe:
		return EPIPE
	case types.ErrorCodeReadOnly:
		return EROFS
	case types.ErrorCodeInvalidSeek:
		return ESPIPE
	case types.ErrorCodeTextFileBusy:
		return ETXTBSY
	case types.ErrorCodeCrossDevice:
		return EXDEV
	}
	return Errno(err)
}

type libc_DIR struct {
	d types.DirectoryEntryStream
}

// DIR *fdopendir(int);
//
//export fdopendir
func fdopendir(fd int32) unsafe.Pointer {
	if _, ok := wasiStreams[fd]; ok {
		libcErrno = EBADF
		return nil
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = EBADF
		return nil
	}
	if stream.d == cm.ResourceNone {
		libcErrno = EBADF
		return nil
	}

	result := stream.d.ReadDirectory()
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return nil
	}

	return unsafe.Pointer(&libc_DIR{d: *result.OK()})
}

// int fdclosedir(DIR *);
//
//export fdclosedir
func fdclosedir(dirp unsafe.Pointer) int32 {
	if dirp == nil {
		return 0

	}
	dir := (*libc_DIR)(dirp)
	if dir.d == cm.ResourceNone {
		return 0
	}

	dir.d.ResourceDrop()
	dir.d = cm.ResourceNone

	return 0
}

// struct dirent *readdir(DIR *);
//
//export readdir
func readdir(dirp unsafe.Pointer) *Dirent {
	if dirp == nil {
		return nil

	}
	dir := (*libc_DIR)(dirp)
	if dir.d == cm.ResourceNone {
		return nil
	}

	result := dir.d.ReadDirectoryEntry()
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return nil
	}

	entry := result.OK().Some()
	if entry == nil {
		libcErrno = 0
		return nil
	}

	// The dirent C struct uses a flexible array member to indicate that the
	// directory name is laid out in memory right after the struct data:
	//
	// struct dirent {
	//   ino_t d_ino;
	//   unsigned char d_type;
	//   char d_name[];
	// };
	buf := make([]byte, unsafe.Sizeof(Dirent{})+uintptr(len(entry.Name)))
	dirent := (*Dirent)((unsafe.Pointer)(&buf[0]))

	// No inodes in wasi
	dirent.Ino = 0
	dirent.Type = p2fileTypeToDirentType(entry.Type)
	copy(buf[unsafe.Offsetof(dirent.Type)+1:], entry.Name)

	return dirent
}

func p2fileTypeToDirentType(t types.DescriptorType) uint8 {
	switch t {
	case types.DescriptorTypeUnknown:
		return DT_UNKNOWN
	case types.DescriptorTypeBlockDevice:
		return DT_BLK
	case types.DescriptorTypeCharacterDevice:
		return DT_CHR
	case types.DescriptorTypeDirectory:
		return DT_DIR
	case types.DescriptorTypeFIFO:
		return DT_FIFO
	case types.DescriptorTypeSymbolicLink:
		return DT_LNK
	case types.DescriptorTypeRegularFile:
		return DT_REG
	case types.DescriptorTypeSocket:
		return DT_FIFO
	}

	return DT_UNKNOWN
}

func p2fileTypeToStatType(t types.DescriptorType) uint32 {
	switch t {
	case types.DescriptorTypeUnknown:
		return 0
	case types.DescriptorTypeBlockDevice:
		return S_IFBLK
	case types.DescriptorTypeCharacterDevice:
		return S_IFCHR
	case types.DescriptorTypeDirectory:
		return S_IFDIR
	case types.DescriptorTypeFIFO:
		return S_IFIFO
	case types.DescriptorTypeSymbolicLink:
		return S_IFLNK
	case types.DescriptorTypeRegularFile:
		return S_IFREG
	case types.DescriptorTypeSocket:
		return S_IFSOCK
	}

	return 0
}

var libc_envs map[string]string

func populateEnvironment() {
	libc_envs = make(map[string]string)
	for _, kv := range environment.GetEnvironment().Slice() {
		libc_envs[kv[0]] = kv[1]
	}
}

// char * getenv(const char *name);
//
//export getenv
func getenv(key *byte) *byte {
	k := goString(key)

	v, ok := libc_envs[k]
	if !ok {
		return nil
	}

	// The new allocation is zero-filled; allocating an extra byte and then
	// copying the data over will leave the last byte untouched,
	// null-terminating the string.
	vbytes := make([]byte, len(v)+1)
	copy(vbytes, v)
	return unsafe.SliceData(vbytes)
}

// int setenv(const char *name, const char *value, int overwrite);
//
//export setenv
func setenv(key, value *byte, overwrite int) int {
	k := goString(key)
	if _, ok := libc_envs[k]; ok && overwrite == 0 {
		return 0
	}

	v := goString(value)
	libc_envs[k] = v

	return 0
}

// int unsetenv(const char *name);
//
//export unsetenv
func unsetenv(key *byte) int {
	k := goString(key)
	delete(libc_envs, k)
	return 0
}

// void arc4random_buf (void *, size_t);
//
//export arc4random_buf
func arc4random_buf(p unsafe.Pointer, l uint) {
	result := random.GetRandomBytes(uint64(l))
	s := result.Slice()
	memcpy(unsafe.Pointer(p), unsafe.Pointer(unsafe.SliceData(s)), uintptr(l))
}

// int chdir(char *name)
//
//export chdir
func chdir(name *byte) int {
	path := goString(name) + "/"

	if !isAbs(path) {
		path = joinPath(libcCWD.root+"/"+libcCWD.rel+"/", path)
	}

	if path == "." {
		return 0
	}

	dir, rel := findPreopenForPath(path)
	if dir.d == cm.ResourceNone {
		libcErrno = EACCES
		return -1
	}

	result := dir.d.OpenAt(types.PathFlagsSymlinkFollow, rel, types.OpenFlagsDirectory, types.DescriptorFlagsRead)
	if err := result.Err(); err != nil {
		libcErrno = errorCodeToErrno(*err)
		return -1
	}

	libcCWD = dir
	// keep the same cwd base but update "rel" to point to new base path
	libcCWD.rel = rel

	return 0
}

// char *getcwd(char *buf, size_t size)
//
//export getcwd
func getcwd(buf *byte, size uint) *byte {

	cwd := libcCWD.root
	if libcCWD.rel != "" && libcCWD.rel != "." && libcCWD.rel != "./" {
		cwd += libcCWD.rel
	}

	if buf == nil {
		b := make([]byte, len(cwd)+1)
		buf = unsafe.SliceData(b)
	} else if size == 0 {
		libcErrno = EINVAL
		return nil
	}

	if size < uint(len(cwd)+1) {
		libcErrno = ERANGE
		return nil
	}

	s := unsafe.Slice(buf, size)
	s[size-1] = 0 // Enforce NULL termination
	copy(s, cwd)
	return buf
}

// int truncate(const char *path, off_t length);
//
//export truncate
func truncate(path *byte, length int64) int32 {
	libcErrno = ENOSYS
	return -1
}
