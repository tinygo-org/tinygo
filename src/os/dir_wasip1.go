// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was derived from src/os/dir_darwin.go since the logic for WASI is
// fairly similar: we use fdopendir, fdclosedir, and readdir from wasi-libc in
// a similar way that the darwin code uses functions from libc.

//go:build wasip1

package os

import (
	"io"
	"runtime"
	"syscall"
	"unsafe"
)

// opaque DIR* returned by fdopendir
//
// We add an unused field so it is not the empty struct, which is usually
// a special case in Go.
type dirInfo struct{ _ int32 }

func (d *dirInfo) close() {
	syscall.Fdclosedir(uintptr(unsafe.Pointer(d)))
}

func (f *File) readdir(n int, mode readdirMode) (names []string, dirents []DirEntry, infos []FileInfo, err error) {
	if f.dirinfo == nil {
		dir, errno := syscall.Fdopendir(syscallFd(f.handle.(unixFileHandle)))
		if errno != nil {
			return nil, nil, nil, &PathError{Op: "fdopendir", Path: f.name, Err: errno}
		}
		f.dirinfo = (*dirInfo)(unsafe.Pointer(dir))
	}
	d := uintptr(unsafe.Pointer(f.dirinfo))

	// see src/os/dir_unix.go
	if n == 0 {
		n = -1
	}

	for n != 0 {
		dirent, errno := syscall.Readdir(d)
		if errno != nil {
			if errno == syscall.EINTR {
				continue
			}
			return names, dirents, infos, &PathError{Op: "readdir", Path: f.name, Err: errno}
		}
		if dirent == nil { // EOF
			break
		}
		name := dirent.Name()
		// Check for useless names before allocating a string.
		if string(name) == "." || string(name) == ".." {
			continue
		}
		if n > 0 {
			n--
		}
		if mode == readdirName {
			names = append(names, string(name))
		} else if mode == readdirDirEntry {
			de, err := newUnixDirent(f.name, string(name), dtToType(dirent.Type))
			if IsNotExist(err) {
				// File disappeared between readdir and stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, dirents, nil, err
			}
			dirents = append(dirents, de)
		} else {
			info, err := lstat(f.name + "/" + string(name))
			if IsNotExist(err) {
				// File disappeared between readdir + stat.
				// Treat as if it didn't exist.
				continue
			}
			if err != nil {
				return nil, nil, infos, err
			}
			infos = append(infos, info)
		}
		runtime.KeepAlive(f)
	}

	if n > 0 && len(names)+len(dirents)+len(infos) == 0 {
		return nil, nil, nil, io.EOF
	}
	return names, dirents, infos, nil
}

func dtToType(typ uint8) FileMode {
	switch typ {
	case syscall.DT_BLK:
		return ModeDevice
	case syscall.DT_CHR:
		return ModeDevice | ModeCharDevice
	case syscall.DT_DIR:
		return ModeDir
	case syscall.DT_FIFO:
		return ModeNamedPipe
	case syscall.DT_LNK:
		return ModeSymlink
	case syscall.DT_REG:
		return 0
	}
	return ^FileMode(0)
}
