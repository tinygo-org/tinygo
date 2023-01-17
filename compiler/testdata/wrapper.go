package main

import (
	"io"
	"sync"
)

type PathError struct {
	Op   string
	Path string
	Err  error
}

type dirInfo struct {
	nbuf int
	bufp int
	buf  [blockSize]byte
}

func (d *dirInfo) close() {
}

type FileHandle interface {
	Close() error
}

type file struct {
	handle  FileHandle
	name    string
	dirInfo dirInfo
}

func (f *file) close() (err error) {
	if f.dirinfo != nil {
		f.dirinfo.close()
		f.dirinfo = nil
	}
	return f.handle.Close()
}

type File struct{ *file }

func (f *File) Close() (err error) {
	if f.handle == nil {
		err = ErrClosed
	} else {
		err = f.file.close()
		if err == nil {
			f.handle = nil
		}
	}
	if err != nil {
		err = &PathError{Op: "close", Path: f.name, Err: err}
	}
	return
}

type closeOnce struct {
	*File

	once sync.Once
	err  error
}

func (c *closeOnce) Close() error {
	c.once.Do(c.close)
	return c.err
}

func (c *closeOnce) close() {
	c.err = c.File.Close()
}

var f File
var c io.Closer = &closeOnce{&f}

func main() {}
