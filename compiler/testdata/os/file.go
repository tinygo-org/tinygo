package os

import "io"

type File struct{ *file }

func (f *File) Read(b []byte) (int, error) {
	return 0, io.EOF
}

func (f *File) Write(b []byte) (int, error) {
	return len(b), nil
}

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

func Pipe() (r, w *File, err error) {
	r = new(File)
	w = new(File)
	return
}
