package os

import "unsafe"

type dirInfo struct {
	nbuf int
	bufp int
	buf  [blockSize]byte
}

const (
	blockSize = 8192 - 2*unsafe.Sizeof(int(0))
)

func (d *dirInfo) close() {
}

type file struct {
	handle  FileHandle
	name    string
	dirinfo *dirInfo
}

func (f *file) close() (err error) {
	if f.dirinfo != nil {
		f.dirinfo.close()
		f.dirinfo = nil
	}
	return f.handle.Close()
}
