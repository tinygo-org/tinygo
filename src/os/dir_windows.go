// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/windows"
	"io"
	"io/fs"
	"syscall"
	"unsafe"
)

type dirInfo struct {
	buf   [4096]byte
	bufp  int
	class uint32
	path  string
	eof   bool
}

func (f *File) readdir(n int, mode readdirMode) ([]string, []DirEntry, []FileInfo, error) {
	if f.handle == nil {
		return nil, nil, nil, &PathError{Op: "readdir", Path: f.name, Err: ErrClosed}
	}
	h := syscallFd(f.handle.(unixFileHandle))
	if f.dirinfo == nil {
		path, err := windows.FinalPath(h, windows.FILE_NAME_OPENED)
		if err != nil {
			return nil, nil, nil, &PathError{Op: "GetFinalPathNameByHandle", Path: f.name, Err: err}
		}
		f.dirinfo = &dirInfo{class: windows.FileFullDirectoryRestartInfo, path: path}
	}
	d := f.dirinfo
	var names []string
	var dirents []DirEntry
	var infos []FileInfo
	wantAll := n <= 0
	for (wantAll || n > 0) && !d.eof {
		if d.bufp == 0 {
			err := windows.GetFileInformationByHandleEx(h, d.class, &d.buf[0], uint32(len(d.buf)))
			if err != nil {
				// MS-FSA 2.1.5.6.3 permits STATUS_NO_SUCH_FILE for an empty root.
				// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fsa/fa8194e0-53ec-413b-8315-e8fa85396fd8
				if err == syscall.ERROR_NO_MORE_FILES || err == syscall.ERROR_FILE_NOT_FOUND && d.class == windows.FileFullDirectoryRestartInfo {
					d.eof = true
					break
				}
				if info, statErr := f.Stat(); statErr == nil && !info.IsDir() {
					err = syscall.ENOTDIR
				}
				return names, dirents, infos, &PathError{Op: "readdir", Path: f.name, Err: err}
			}
			d.class = windows.FileFullDirectoryInfo
		}
		entry := (*windows.FILE_FULL_DIR_INFO)(unsafe.Pointer(&d.buf[d.bufp]))
		d.bufp += int(entry.NextEntryOffset)
		if entry.NextEntryOffset == 0 {
			d.bufp = 0
		}
		name := syscall.UTF16ToString(unsafe.Slice(&entry.FileName[0], entry.FileNameLength/2))
		if name == "." || name == ".." {
			continue
		}
		if mode == readdirName {
			names = append(names, name)
		} else {
			info := &fileStat{
				name:             name,
				FileAttributes:   entry.FileAttributes,
				CreationTime:     entry.CreationTime,
				LastAccessTime:   entry.LastAccessTime,
				LastWriteTime:    entry.LastWriteTime,
				FileSizeHigh:     uint32(entry.EndOfFile >> 32),
				FileSizeLow:      uint32(entry.EndOfFile),
				Reserved0:        entry.EaSize,
				path:             d.path,
				appendNameToPath: true,
			}
			if mode == readdirDirEntry {
				dirents = append(dirents, fs.FileInfoToDirEntry(info))
			} else {
				infos = append(infos, info)
			}
		}
		n--
	}
	if !wantAll && len(names)+len(dirents)+len(infos) == 0 {
		return nil, nil, nil, io.EOF
	}
	return names, dirents, infos, nil
}
